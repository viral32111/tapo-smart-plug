package aes_passthrough

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/viral32111/tapo-smart-plug/internal/helpers"
)

// base64(hex(sha1(input))) as required by login_device.
func hashCredential(input string) string {
	hash := sha1.Sum([]byte(input))
	hex := fmt.Sprintf("%x", hash)
	return base64.StdEncoding.EncodeToString([]byte(hex))
}

func encrypt(plaintext []byte, key []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := helpers.PKCS7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(b64ciphertext string, key []byte, iv []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(b64ciphertext)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not a multiple of block size", len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	return helpers.PKCS7Unpad(plaintext)
}

func Handshake(client *http.Client, baseUrl string) (*Session, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key pair: %w", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %w", err)
	}

	publicKeyPEM := "-----BEGIN PUBLIC KEY-----\n" +
		base64.StdEncoding.EncodeToString(publicKeyDER) +
		"\n-----END PUBLIC KEY-----\n"

	payload, _ := json.Marshal(map[string]any{
		"method": "handshake",
		"params": map[string]any{
			"key":             publicKeyPEM,
			"requestTimeMils": 0,
		},
	})

	response, err := client.Post(baseUrl+"/app", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("handshake HTTP request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading handshake response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", response.StatusCode, body)
	}

	var data struct {
		ErrorCode int `json:"error_code"`
		Result    struct {
			Key string `json:"key"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing handshake response: %w", err)
	}
	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("handshake error_code %d (wrong firmware/credentials?)", data.ErrorCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range response.Cookies() {
		if c.Name == "TP_SESSIONID" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		return nil, fmt.Errorf("no TP_SESSIONID cookie in handshake response")
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(data.Result.Key, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("base64-decoding device key: %w", err)
	}

	keyBlob, err := rsa.DecryptPKCS1v15(nil, privateKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("RSA-decrypting device key: %w", err)
	}
	if len(keyBlob) != 32 {
		return nil, fmt.Errorf("expected 32 bytes for AES key+IV, got %d", len(keyBlob))
	}

	return &Session{
		HttpClient: client,

		BaseUrl:       baseUrl,
		SessionCookie: sessionCookie,

		Key: keyBlob[:16],
		IV:  keyBlob[16:],
	}, nil
}
