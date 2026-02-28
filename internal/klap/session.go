package klap

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	HttpClient *http.Client

	BaseUrl       string
	SessionCookie *http.Cookie

	Key            []byte
	InitVector     []byte
	SignatureKey   []byte
	SequenceNumber int32

	MutexLock sync.Mutex // Protects sequenceNumber
}

// Performs the full KLAP handshake and returns a ready session
func NewSession(ipAddress, emailAddress, password string) (*Session, error) {
	// Setup a HTTP client for future requests
	baseUrl := fmt.Sprintf("http://%s/app", ipAddress)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	localSeed := make([]byte, 16)
	if _, err := rand.Read(localSeed); err != nil {
		return nil, fmt.Errorf("Error generating local seed: %w", err)
	}

	// Step 1: Hash credentials for authentication
	authHash := hashAuthentication(emailAddress, password)

	// Step 2: Perform 1st handshake - exchange seeds, establish session cookie & verify server knowledge of credentials
	remoteSeed, sessionCookie, err := klapHandshake1(httpClient, baseUrl, localSeed, authHash)
	if err != nil {
		return nil, fmt.Errorf("Error during 1st handshake: %w", err)
	}

	// Step 3: Perform 2nd handshake - prove knowledge of credentials
	if err := klapHandshake2(httpClient, baseUrl, sessionCookie, localSeed, remoteSeed, authHash); err != nil {
		return nil, fmt.Errorf("Error during 2nd handshake: %w", err)
	}

	// Step 4: Derive starting sequence number & session keys from seeds & auth hash
	sequenceNumber := deriveInitialSequenceNumber(localSeed, remoteSeed, authHash)
	key, initVector, signatureKey := deriveSessionKeys(localSeed, remoteSeed, authHash)

	return &Session{
		HttpClient: httpClient,

		BaseUrl:       baseUrl,
		SessionCookie: sessionCookie,

		Key:            key,
		InitVector:     initVector,
		SignatureKey:   signatureKey,
		SequenceNumber: sequenceNumber,
	}, nil
}

// Encrypts a JSON payload, sends it, and returns the decrypted response body
func (session *Session) Request(payload []byte) ([]byte, error) {
	// Safely increment sequence number for this request
	session.MutexLock.Lock()
	session.SequenceNumber++
	sequenceNumber := session.SequenceNumber
	session.MutexLock.Unlock()

	encryptedRequestPayload, err := encrypt(session.Key, session.InitVector, session.SignatureKey, sequenceNumber, payload)
	if err != nil {
		return nil, fmt.Errorf("Error encrypting data: %w", err)
	}

	// Perform HTTP request with sequence number, encrypted payload & session cookie
	url := fmt.Sprintf("%s/request?seq=%d", session.BaseUrl, sequenceNumber)
	request, err := http.NewRequest("POST", url, bytes.NewReader(encryptedRequestPayload))
	request.Header.Set("Content-Type", "application/octet-stream")
	request.AddCookie(session.SessionCookie)
	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP request: %w", err)
	}

	response, err := session.HttpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error sending HTTP request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading HTTP response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode, string(body))
	}

	decryptedResponsePayload, err := decrypt(session.Key, session.InitVector, sequenceNumber, body)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting response: %w", err)
	}

	return decryptedResponsePayload, nil
}
