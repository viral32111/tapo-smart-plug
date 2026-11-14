package tests

import (
	"bytes"
	"testing"

	"github.com/viral32111/tapo-smart-plug/internal/helpers"
)

func TestPKCS7(test *testing.T) {
	for _, size := range []int{0, 1, 15, 16, 17, 31, 32} {
		data := bytes.Repeat([]byte{0xAB}, size)
		padded := helpers.PKCS7Pad(data, 16)

		if len(padded)%16 != 0 {
			test.Errorf("PKCS7 padded length %d is not a multiple of 16", len(padded))
		}

		unpadded, err := helpers.PKCS7Unpad(padded)
		if err != nil {
			test.Fatalf("Error in PKCS7 unpad: %v", err)
		}

		if !bytes.Equal(data, unpadded) {
			test.Errorf("PKCS7 pad/unpad mismatch for size %d", size)
		}
	}
}
