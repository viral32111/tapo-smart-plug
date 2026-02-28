package klap

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

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
