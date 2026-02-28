package klap

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

// Derives the hash used in the KLAP handshake
func hashAuthentication(email, password string) []byte {
	emailHash := sha1.Sum([]byte(email))
	passwordHash := sha1.Sum([]byte(password))

	credentials := append(emailHash[:], passwordHash[:]...)
	hash := sha256.Sum256(credentials)

	return hash[:]
}

// Derives the AES key, IV base & signature key from the two seeds and the credential hash
func deriveSessionKeys(localSeed, remoteSeed, authHash []byte) (key, initVector, signature []byte) {
	derive := func(prefix string) []byte {
		hash := sha256.Sum256(append([]byte(prefix), append(localSeed, append(remoteSeed, authHash...)...)...))
		return hash[:]
	}

	return derive("lsk")[:16], derive("iv")[:12], derive("ldk")[:28]
}

// Derives the initial sequence number from the seeds and auth hash, using the same method as IV derivation
func deriveInitialSequenceNumber(localSeed, remoteSeed, authHash []byte) int32 {
	hash := sha256.Sum256(append([]byte("iv"), append(localSeed, append(remoteSeed, authHash...)...)...))
	remainder := hash[28:] // Last 4 bytes
	return int32(remainder[0])<<24 | int32(remainder[1])<<16 | int32(remainder[2])<<8 | int32(remainder[3])
}

// 1st handshake
func klapHandshake1(client *http.Client, baseUrl string, localSeed, authHash []byte) (remoteSeed []byte, sessionCookie *http.Cookie, err error) {
	response, err := client.Post(baseUrl+"/handshake1", "application/octet-stream", bytes.NewReader(localSeed))
	if err != nil {
		return nil, nil, fmt.Errorf("Error sending HTTP request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode)
	}

	// Extract TP_SESSIONID from the response cookies
	for _, cookie := range response.Cookies() {
		if cookie.Name == "TP_SESSIONID" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		return nil, nil, fmt.Errorf("No 'TP_SESSIONID' cookie in response")
	}

	// remoteSeed(16) + serverHash(32) = 48 bytes
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, sessionCookie, fmt.Errorf("Error reading HTTP response: %w", err)
	}
	if len(body) != 48 {
		return nil, sessionCookie, fmt.Errorf("Expected 48 bytes in response, got %d byte(s)", len(body))
	}

	remoteSeed = body[:16]
	serverHash := body[16:]

	// Verify the server knows the same credentials we do
	expectation := sha256.Sum256(append(append(localSeed, remoteSeed...), authHash...))
	if !bytes.Equal(expectation[:], serverHash) {
		return nil, sessionCookie, fmt.Errorf("Server hash mismatch (likely incorrect credentials or wrong device)")
	}

	return remoteSeed, sessionCookie, nil
}

// 2nd handshake
func klapHandshake2(client *http.Client, baseUrl string, sessionCookie *http.Cookie, localSeed, remoteSeed, authHash []byte) error {
	// Prove we know the credentials — note order is remote+local (swapped from h1)
	proof := sha256.Sum256(append(append(remoteSeed, localSeed...), authHash...))

	request, err := http.NewRequest("POST", baseUrl+"/handshake2", bytes.NewReader(proof[:]))
	request.Header.Set("Content-Type", "application/octet-stream")
	request.AddCookie(sessionCookie)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode, string(body))
	}

	return nil
}
