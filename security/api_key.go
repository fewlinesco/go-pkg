package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

var ErrCantGenerateAPIKey = errors.New("can't generate an API-Key")

func GenerateAPIKeyWithCustomSize(size int) (string, error) {
	b, err := generateRandomBytes(size)
	return base64.URLEncoding.EncodeToString(b), err
}

func GenerateAPIKey() (string, error) {
	return GenerateAPIKeyWithCustomSize(32)
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCantGenerateAPIKey, err)
	}

	return b, nil
}
