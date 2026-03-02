package aes_passthrough

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Session struct {
	HttpClient *http.Client

	BaseUrl       string
	SessionCookie *http.Cookie

	Key []byte
	IV  []byte

	Token string
}

func New(ipAddress string, emailAddress string, password string) (*Session, error) {
	baseUrl := fmt.Sprintf("http://%s/app", ipAddress)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	session, err := Handshake(httpClient, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("Error during handshake: %w", err)
	}

	result, err := session.Request("login_device", map[string]any{
		"username": hashCredential(emailAddress),
		"password": hashCredential(password),
	})
	if err != nil {
		return nil, fmt.Errorf("Error during login: %w", err)
	}

	var loginResult struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(result, &loginResult); err != nil {
		return nil, fmt.Errorf("Error decoding login response: %w", err)
	}

	if loginResult.Token == "" {
		return nil, fmt.Errorf("No token in login response?!")
	}

	session.Token = loginResult.Token

	return session, nil
}

func (session *Session) Request(method string, params any) (json.RawMessage, error) {
	payload, err := json.Marshal(map[string]any{
		"method":          method,
		"params":          params,
		"requestTimeMils": time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, fmt.Errorf("Error encoding request for method '%s' as JSON: %w", method, err)
	}

	encrypted, err := encrypt(payload, session.Key, session.IV)
	if err != nil {
		return nil, fmt.Errorf("Error encrypting request for method '%s' payload: %w", method, err)
	}

	passthroughPayload, err := json.Marshal(map[string]any{
		"method": "securePassthrough",
		"params": map[string]any{
			"request": encrypted,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Error encoding passthrough request for method '%s' as JSON: %w", method, err)
	}

	url := session.BaseUrl
	if session.Token != "" {
		url += "?token=" + session.Token
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(passthroughPayload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(session.SessionCookie)

	response, err := session.HttpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error sending HTTP request for method '%s': %w", method, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading HTTP response for method '%s': %w", method, err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status code %d for method '%s': %s", response.StatusCode, method, string(body))
	}

	var data struct {
		ErrorCode int `json:"error_code"`
		Result    struct {
			Response string `json:"response"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for method '%s': %w", method, err)
	}
	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device returned error code %d for method '%s'", data.ErrorCode, method)
	}

	decrypted, err := decrypt(data.Result.Response, session.Key, session.IV)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting response for method '%s': %w", method, err)
	}

	var passthroughData struct {
		ErrorCode int             `json:"error_code"`
		Result    json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(decrypted, &passthroughData); err != nil {
		return nil, fmt.Errorf("Error parsing decrypted response for method '%s': %w", method, err)
	}
	if passthroughData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device returned error code %d inside decrypted response for method '%s'", passthroughData.ErrorCode, method)
	}

	return passthroughData.Result, nil
}
