package helpers

import (
	"bytes"
	"crypto/aes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Creates a UUID from the hash of any value
func DeriveUUID(value string) (string, error) {
	hash := sha256.Sum256([]byte(value))
	bytes := hash[:16]

	bytes[6] = (bytes[6] & 0x0F) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3F) | 0x80 // Variant 10xx

	uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(bytes[0:4]),
		binary.BigEndian.Uint16(bytes[4:6]),
		binary.BigEndian.Uint16(bytes[6:8]),
		binary.BigEndian.Uint16(bytes[8:10]),
		bytes[10:16],
	)

	return uuid, nil
}

// Pads data to a multiple of blockSize using PKCS#7
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// Remove PKCS7 padding
func PKCS7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 || length%aes.BlockSize != 0 {
		return nil, fmt.Errorf("Invalid padded length: %d", length)
	}

	padding := int(data[length-1])
	if padding == 0 || padding > aes.BlockSize {
		return nil, fmt.Errorf("Invalid PKCS7 padding byte: %d", padding)
	}

	return data[:length-padding], nil
}
