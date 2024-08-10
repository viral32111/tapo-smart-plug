package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type RequestPayload struct {
	Action     string       `json:"method"`
	Parameters LoginPayload `json:"params"`
}

type ResponsePayload struct {
	ErrorCode int            `json:"error_code"`
	Result    *LoginResponse `json:"result"`
}

type LoginPayload struct {
	AppType      string `json:"appType"` // Kasa_Android, Tapo_Ios
	EmailAddress string `json:"cloudUserName"`
	Password     string `json:"cloudPassword"`
	TerminalUUID string `json:"terminalUUID"`
}

type LoginResponse struct {
	Identifier          string `json:"accountId"`
	RegisteredAt        string `json:"regTime"`
	CountryCode         string `json:"countryCode"`
	RiskDetected        int    `json:"riskDetected"`
	DisplayName         string `json:"nickname"`
	EmailAddress        string `json:"email"`
	AuthenticationToken string `json:"token"`
}

func main() {
	emailAddress := ""
	password := ""
	fmt.Printf("Email Address: '%s'\n", emailAddress)
	fmt.Printf("Password: '%s'\n\n", password)

	schema := "http"
	ipAddress := "192.168.0.207"
	tcpPort := 80
	httpPath := "/app"

	url := fmt.Sprintf("%s://%s:%d%s", schema, ipAddress, tcpPort, httpPath)
	fmt.Printf("Plug URL: '%s'\n", url)

	cloudUrl := "https://eu-wap.tplinkcloud.com"
	fmt.Printf("Cloud URL: '%s'\n\n", url)

	authenticationTokenFilePath := "token.txt"
	if _, err := os.Stat(authenticationTokenFilePath); os.IsNotExist(err) {
		fmt.Printf("Authentication token file does not exist: '%s'\n", authenticationTokenFilePath)

		terminalUUIDFilePath := "uuid.txt"
		if _, err := os.Stat(terminalUUIDFilePath); os.IsNotExist(err) {
			freshTerminalUUID := uuid.New()
			fmt.Printf("Generated new terminal UUID: '%s'\n", freshTerminalUUID)

			err := os.WriteFile(terminalUUIDFilePath, []byte(freshTerminalUUID.String()), 0644)
			if err != nil {
				fmt.Printf("Error writing terminal UUID to file: '%s'\n", err)
				os.Exit(1)
			}
		}

		rawTerminalUUID, err := os.ReadFile(terminalUUIDFilePath)
		if err != nil {
			fmt.Printf("Error reading terminal UUID from file: '%s'\n", err)
			os.Exit(1)
		}

		terminalUUID := string(rawTerminalUUID)
		fmt.Printf("Loaded existing terminal UUID: '%s'\n", terminalUUID)

		loginPayload := RequestPayload{
			Action: "login",
			Parameters: LoginPayload{
				AppType:      "Tapo_Ios",
				EmailAddress: emailAddress,
				Password:     password,
				TerminalUUID: terminalUUID,
			},
		}

		encodedLoginPayload, err := json.Marshal(loginPayload)
		if err != nil {
			fmt.Printf("Error marshalling request payload: '%s'\n", err)
			os.Exit(1)
		}
		fmt.Printf("Encoded payload: '%s'\n", string(encodedLoginPayload))

		loginRequest, err := http.NewRequest("POST", cloudUrl, bytes.NewBuffer(encodedLoginPayload))
		if err != nil {
			fmt.Printf("Error creating HTTP request: '%s'\n", err)
			os.Exit(1)
		}
		loginRequest.Header.Set("Accept", "application/json")
		loginRequest.Header.Set("Content-Type", "application/json")
		loginRequest.Header.Set("User-Agent", "Tapo Smart Plug/0.0.0 (https://viral32111.com; contact@viral32111.com)")
		loginRequest.Header.Set("From", "contact@viral32111.com")
		loginRequest.Body = io.NopCloser(bytes.NewReader(encodedLoginPayload))

		loginResponse, err := http.DefaultClient.Do(loginRequest)
		if err != nil {
			fmt.Printf("Error sending HTTP request to server: '%s'\n", err)
			os.Exit(1)
		}
		defer loginResponse.Body.Close()

		/*
			{
				"error_code": 0,
				"result": {
					"accountId": "9156386568355",
					"regTime": "2024-08-08 15:26:56",
					"countryCode": "GB",
					"riskDetected": 0,
					"nickname": "",
					"email": "",
					"token": "54571945-54571945-BTgznOUvb8BBSADOqiPfPodJ"
				}
			}
		*/

		body, err := io.ReadAll(loginResponse.Body)
		if err != nil {
			fmt.Printf("Error reading HTTP response body: '%s'\n", err)
			os.Exit(1)
		}
		fmt.Printf("Response: '%s'\n", string(body))

		var loginResponsePayload ResponsePayload
		err = json.Unmarshal(body, &loginResponsePayload)
		if err != nil {
			fmt.Printf("Error parsing error response: '%s'\n", err)
			os.Exit(1)
		}

		if loginResponsePayload.ErrorCode != 0 {
			fmt.Printf("Error code: %d\n", loginResponsePayload.ErrorCode)
			os.Exit(1)
		}

		fmt.Printf("Account ID: '%s'\n", loginResponsePayload.Result.Identifier)
		fmt.Printf("Registration Time: '%s'\n", loginResponsePayload.Result.RegisteredAt)
		fmt.Printf("Country Code: '%s'\n", loginResponsePayload.Result.CountryCode)
		fmt.Printf("Risk Detected: '%d'\n", loginResponsePayload.Result.RiskDetected)
		fmt.Printf("Display Name: '%s'\n", loginResponsePayload.Result.DisplayName)
		fmt.Printf("Email Address: '%s'\n", loginResponsePayload.Result.EmailAddress)
		fmt.Printf("Authentication Token: '%s'\n", loginResponsePayload.Result.AuthenticationToken) // 54571945-BTgznOUvb8BBSADOqiPfPodJ, 54571945-BTyUYorI14pjT1mJYPAmA9T

		err = os.WriteFile(authenticationTokenFilePath, []byte(loginResponsePayload.Result.AuthenticationToken), 0600)
		if err != nil {
			fmt.Printf("Error writing authentication token to file: '%s'\n", err)
			os.Exit(1)
		}
	} else {
		rawAuthenticationToken, err := os.ReadFile(authenticationTokenFilePath)
		if err != nil {
			fmt.Printf("Error reading authentication token from file: '%s'\n", err)
			os.Exit(1)
		}

		authenticationToken := string(rawAuthenticationToken)
		fmt.Printf("Loaded existing authentication token: '%s'\n", authenticationToken)
	}

	os.Exit(0)
}
