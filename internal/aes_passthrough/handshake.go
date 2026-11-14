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
func HashCredential(input string) string {
	hash := sha1.Sum([]byte(input))
	hex := fmt.Sprintf("%x", hash)
	return base64.StdEncoding.EncodeToString([]byte(hex))
}

func Encrypt(plainText []byte, key []byte, initVector []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := helpers.PKCS7Pad(plainText, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, initVector).CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(cipherTextB64 string, key []byte, initVector []byte) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return nil, fmt.Errorf("Error decoding base64 cipher-text: %w", err)
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("Cipher-text length %d is not a multiple of AES block size %d", len(cipherText), aes.BlockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plainText := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, initVector).CryptBlocks(plainText, cipherText)

	return helpers.PKCS7Unpad(plainText)
}

func EncodePublicKeyPEM(derBytes []byte) string {
	derB64 := base64.StdEncoding.EncodeToString(derBytes)

	var strBuilder strings.Builder
	strBuilder.WriteString("-----BEGIN PUBLIC KEY-----\n")

	for len(derB64) > 64 {
		strBuilder.WriteString(derB64[:64])
		strBuilder.WriteByte('\n')
		derB64 = derB64[64:]
	}

	if len(derB64) > 0 {
		strBuilder.WriteString(derB64)
		strBuilder.WriteByte('\n')
	}

	strBuilder.WriteString("-----END PUBLIC KEY-----\n")

	return strBuilder.String()
}

func Handshake(client *http.Client, baseUrl string) (*Session, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("Error generating RSA-1024 key pair: %w", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("Error DER encoding public key: %w", err)
	}

	publicKeyPEM := EncodePublicKeyPEM(publicKeyDER)

	payload, _ := json.Marshal(map[string]any{
		"method": "handshake",
		"params": map[string]any{
			"key":             publicKeyPEM,
			"requestTimeMils": 0,
		},
	})

	response, err := client.Post(baseUrl, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("Error sending HTTP POST request to '%s' for handshake: %w", baseUrl, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading HTTP response body for handshake: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status code %d for handshake: %s", response.StatusCode, body)
	}

	var data struct {
		ErrorCode int `json:"error_code"`
		Result    struct {
			Key string `json:"key"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("Error parsing handshake response: %w", err)
	}
	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code %d upon handshake", data.ErrorCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range response.Cookies() {
		if c.Name == "TP_SESSIONID" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		return nil, fmt.Errorf("Unable to extract 'TP_SESSIONID' session cookie in handshake response")
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(data.Result.Key, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("Error decoding base64 key: %w", err)
	}

	keyBlob, err := rsa.DecryptPKCS1v15(nil, privateKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("Error RSA decrypting key: %w", err)
	}
	if len(keyBlob) != 32 {
		return nil, fmt.Errorf("Expected 32 bytes for AES key & IV, but got %d", len(keyBlob))
	}

	return &Session{
		HttpClient: client,

		BaseUrl:       baseUrl,
		SessionCookie: sessionCookie,

		Key: keyBlob[:16],
		IV:  keyBlob[16:],

		Token: "",
	}, nil
}
