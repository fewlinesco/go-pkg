package secure

import (
	"crypto/rand"
	"fmt"
)

func APIKey() string {
	return Code(16)
}

func Code(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	return fmt.Sprintf("%x", b)
}
