package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type KLAPSession struct {
	client *http.Client

	baseUrl       string
	sessionCookie *http.Cookie

	key            []byte
	initVector     []byte
	signatureKey   []byte
	sequenceNumber int32

	terminalUUID string

	mutexLock sync.Mutex // Protects sequenceNumber
}

/************************************************************/

// Derives the hash used in the KLAP handshake
func hashAuthentication(email, password string) []byte {
	emailHash := sha1.Sum([]byte(email))
	passwordHash := sha1.Sum([]byte(password))

	credentials := append(emailHash[:], passwordHash[:]...)
	hash := sha256.Sum256(credentials)

	return hash[:]
}

// Derives the AES key, IV base, and signature key from the two seeds and the credential hash
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

// Returns the per-request IV by XOR-ing the last 4 bytes of the base IV with the current sequence number
func computeIV(baseIV []byte, sequenceNumber int32) []byte {
	sequenceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceBytes, uint32(sequenceNumber)) // Cast for the byte packing
	return append(baseIV, sequenceBytes...)                           // 12 + 4 = 16 bytes, plain concatenation
}

// Pad with PKCS7 to AES block size (16 bytes)
func pad(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

// Remove PKCS7 padding
func unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 || length%aes.BlockSize != 0 {
		return nil, fmt.Errorf("Invalid padded length: %d", length)
	}

	padding := int(data[length-1])
	if padding == 0 || padding > aes.BlockSize {
		return nil, fmt.Errorf("Invalid padding byte: %d", padding)
	}

	return data[:length-padding], nil
}

// Encrypts the plain-text & prepends the header
func encrypt(key, baseIV, signatureKey []byte, sequenceNumber int32, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Error calling aes.NewCipher(): %w", err)
	}

	initVector := computeIV(baseIV, sequenceNumber)
	padded := pad(plainText)
	cipherText := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, initVector).CryptBlocks(cipherText, padded)

	sequenceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceBytes, uint32(sequenceNumber))

	// sigKey(28) + seq(4) + ciphertext
	signatureInput := append(signatureKey, append(sequenceBytes, cipherText...)...)
	signatureHash := sha256.Sum256(signatureInput)

	// sig(32) + ciphertext
	return append(signatureHash[:], cipherText...), nil
}

// Strips the header & decrypts the cipher-text
func decrypt(key, baseIV []byte, sequenceNumber int32, payload []byte) ([]byte, error) {
	if len(payload) < 32 {
		return nil, fmt.Errorf("Unexpected payload length (too short), got %d byte(s)", len(payload))
	}

	// Skip header
	cipherText := payload[32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Error calling aes.NewCipher(): %w", err)
	}

	initVector := computeIV(baseIV, sequenceNumber)
	plainText := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, initVector).CryptBlocks(plainText, cipherText)

	return unpad(plainText)
}

func doKLAPHandshake1(client *http.Client, baseUrl string, localSeed, authHash []byte) (remoteSeed []byte, sessionCookie *http.Cookie, err error) {
	response, err := client.Post(baseUrl+"/handshake1", "application/octet-stream", bytes.NewReader(localSeed))
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, sessionCookie, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode, string(body))
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

func doKLAPHandshake2(client *http.Client, baseUrl string, sessionCookie *http.Cookie, localSeed, remoteSeed, authHash []byte) (sequenceNumber int32, err error) {
	// Prove we know the credentials — note order is remote+local (swapped from h1)
	proof := sha256.Sum256(append(append(remoteSeed, localSeed...), authHash...))

	request, err := http.NewRequest("POST", baseUrl+"/handshake2", bytes.NewReader(proof[:]))
	request.Header.Set("Content-Type", "application/octet-stream")
	request.AddCookie(sessionCookie)
	if err != nil {
		return 0, err
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode, string(body))
	}

	sequenceNumber = deriveInitialSequenceNumber(localSeed, remoteSeed, authHash)
	return sequenceNumber, nil
}

/************************************************************/

// Performs the full KLAP handshake and returns a ready session
func NewKLAPSession(ipAddress, emailAddress, password string) (*KLAPSession, error) {
	authHash := hashAuthentication(emailAddress, password)
	baseUrl := fmt.Sprintf("http://%s/app", ipAddress)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	localSeed := make([]byte, 16)
	if _, err := rand.Read(localSeed); err != nil {
		return nil, fmt.Errorf("Error generating local seed: %w", err)
	}

	// 1st handshake: send local seed, receive remote seed & server proof
	remoteSeed, sessionCookie, err := doKLAPHandshake1(client, baseUrl, localSeed, authHash)
	if err != nil {
		return nil, fmt.Errorf("Error during 1st handshake: %w", err)
	}

	// 2nd handshake: prove knowledge of credentials, receive initial sequence number
	sequenceNumber, err := doKLAPHandshake2(client, baseUrl, sessionCookie, localSeed, remoteSeed, authHash)
	if err != nil {
		return nil, fmt.Errorf("Error during 2nd handshake: %w", err)
	}

	terminalUUID := make([]byte, 16)
	rand.Read(terminalUUID)

	key, initVector, signatureKey := deriveSessionKeys(localSeed, remoteSeed, authHash)

	return &KLAPSession{
		client: client,

		baseUrl:       baseUrl,
		sessionCookie: sessionCookie,

		key:            key,
		initVector:     initVector,
		signatureKey:   signatureKey,
		sequenceNumber: sequenceNumber,

		terminalUUID: base64.StdEncoding.EncodeToString(terminalUUID),
	}, nil
}

// Encrypts a JSON payload, sends it, and returns the decrypted response body
func (session *KLAPSession) Request(payload []byte) ([]byte, error) {
	session.mutexLock.Lock()
	session.sequenceNumber++
	sequenceNumber := session.sequenceNumber
	session.mutexLock.Unlock()

	encrypted, err := encrypt(session.key, session.initVector, session.signatureKey, sequenceNumber, payload)
	if err != nil {
		return nil, fmt.Errorf("Error encrypting data: %w", err)
	}

	requestUrl := fmt.Sprintf("%s/request?seq=%d", session.baseUrl, sequenceNumber)

	request, err := http.NewRequest("POST", requestUrl, bytes.NewReader(encrypted))
	request.Header.Set("Content-Type", "application/octet-stream")
	request.AddCookie(session.sessionCookie) // ← this was missing
	if err != nil {
		return nil, err
	}

	response, err := session.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status code %d: %s", response.StatusCode, string(body))
	}

	return decrypt(session.key, session.initVector, sequenceNumber, body)
}
