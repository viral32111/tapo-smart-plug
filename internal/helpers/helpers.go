package helpers

import (
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
