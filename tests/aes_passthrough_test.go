package tests

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/viral32111/tapo-smart-plug/internal/aes_passthrough"
)

func TestMain(m *testing.M) {
	currentDirectory, _ := os.Getwd()

	for {
		dotEnvFilePath := filepath.Join(currentDirectory, ".env")
		if _, err := os.Stat(dotEnvFilePath); err == nil {
			godotenv.Load(dotEnvFilePath)
			break
		}

		parentDirectory := filepath.Dir(currentDirectory)
		if parentDirectory == currentDirectory {
			break
		}

		currentDirectory = parentDirectory
	}

	os.Exit(m.Run())
}

func TestHashCredential(test *testing.T) {
	scenarios := []struct {
		input       string
		expectation string
	}{
		{"user@example.com", "NjNhNzEwNTY5MjYxYTI0YjM3NjYyNzViNzAwMGNlOGQ3YjMyZTJmNw=="},
		{"P4ssw0rd", "MjY0YmMwNzY4MzYyYTY4OTg0ZmFlYTkyM2VmYWEyMWY2N2Y0ZDEwYQ=="},
	}

	for _, scenario := range scenarios {
		result := aes_passthrough.HashCredential(scenario.input)
		if result != scenario.expectation {
			test.Errorf("Credential hash mismatch for %q\nActual: %q\nExpected: %q", scenario.input, result, scenario.expectation)
		}
	}
}

func TestEncryptDecrypt(test *testing.T) {
	key := bytes.Repeat([]byte{0x01}, 16)
	initVector := bytes.Repeat([]byte{0x02}, 16)

	payload := []byte(`{"method":"get_device_info","params":{}}`)

	encrypted, err := aes_passthrough.Encrypt(payload, key, initVector)
	if err != nil {
		test.Fatalf("Error while encrypting: %v", err)
	}

	decrypted, err := aes_passthrough.Decrypt(encrypted, key, initVector)
	if err != nil {
		test.Fatalf("Error while decrypting: %v", err)
	}

	if !bytes.Equal(payload, decrypted) {
		test.Errorf("Encrypt/decrypt mismatch\nActual: %q\nExpected: %q", decrypted, payload)
	}
}

func TestHandshake(test *testing.T) {
	TAPO_HUB_IP_ADDRESS := os.Getenv("TAPO_HUB_IP_ADDRESS")
	if len(TAPO_HUB_IP_ADDRESS) == 0 {
		test.Skip("TAPO_HUB_IP_ADDRESS environment variable not set, skipping handshake test")
	}

	baseUrl := fmt.Sprintf("https://%s:443/app", TAPO_HUB_IP_ADDRESS)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	session, err := aes_passthrough.Handshake(httpClient, baseUrl)
	if err != nil {
		test.Errorf("Error during handshake: %v", err)
	}

	if len(session.Key) != 16 {
		test.Errorf("Unexpected session key length: %d", len(session.Key))
	}

	if len(session.IV) != 16 {
		test.Errorf("Unexpected session IV length: %d", len(session.IV))
	}

	if session.SessionCookie == nil {
		test.Errorf("Session cookie is nil")
	}
}

func TestLogin(test *testing.T) {
	TAPO_HUB_IP_ADDRESS := os.Getenv("TAPO_HUB_IP_ADDRESS")
	if len(TAPO_HUB_IP_ADDRESS) == 0 {
		test.Skip("TAPO_HUB_IP_ADDRESS environment variable not set, skipping handshake test")
	}

	TAPO_ACCOUNT_EMAIL := os.Getenv("TAPO_ACCOUNT_EMAIL")
	if len(TAPO_ACCOUNT_EMAIL) == 0 {
		test.Skip("TAPO_ACCOUNT_EMAIL environment variable not set, skipping handshake test")
	}

	TAPO_ACCOUNT_PASSWORD := os.Getenv("TAPO_ACCOUNT_PASSWORD")
	if len(TAPO_ACCOUNT_PASSWORD) == 0 {
		test.Skip("TAPO_ACCOUNT_PASSWORD environment variable not set, skipping handshake test")
	}

	session, err := aes_passthrough.New(TAPO_HUB_IP_ADDRESS, TAPO_ACCOUNT_EMAIL, TAPO_ACCOUNT_PASSWORD)
	if err != nil {
		test.Errorf("Error during login: %v", err)
	}

	if session.Token == "" {
		test.Errorf("Session token is empty after login")
	}
}

func TestInvalidLogin(test *testing.T) {
	TAPO_HUB_IP_ADDRESS := os.Getenv("TAPO_HUB_IP_ADDRESS")
	if len(TAPO_HUB_IP_ADDRESS) == 0 {
		test.Skip("TAPO_HUB_IP_ADDRESS environment variable not set, skipping handshake test")
	}

	_, err := aes_passthrough.New(TAPO_HUB_IP_ADDRESS, "user@example.com", "P4ssw0rd")
	test.Logf("Received expected error during login with invalid credentials: %v", err)
	if err == nil {
		test.Errorf("Expected error during login with invalid credentials, but got none")
	}
}
