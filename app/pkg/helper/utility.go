package helper

import (
	"crypto/rand"
	"encoding/hex"
)

// Generates a 8-byte hex string
func TraceID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func TinyUuid() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	bytes := make([]byte, length)

	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	alphabetLen := byte(len(alphabet))

	for i := 0; i < length; i++ {
		bytes[i] = alphabet[randomBytes[i]%alphabetLen]
	}

	return string(bytes)
}
