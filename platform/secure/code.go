package secure

import (
	"crypto/rand"
	"fmt"
)

// APIKey generate a random string of 16 characters
func APIKey() string {
	return Code(16)
}

// Code generate a random string of any length
func Code(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	return fmt.Sprintf("%x", b)
}
