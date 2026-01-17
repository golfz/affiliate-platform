package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	alphanumericChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	shortCodeMinLen   = 8
	shortCodeMaxLen   = 12
)

// generateShortCode generates a random alphanumeric short code (8-12 chars)
func generateShortCode() (string, error) {
	// Random length between 8-12
	length, err := rand.Int(rand.Reader, big.NewInt(shortCodeMaxLen-shortCodeMinLen+1))
	if err != nil {
		return "", fmt.Errorf("failed to generate random length: %w", err)
	}
	codeLen := int(length.Int64()) + shortCodeMinLen

	// Generate random alphanumeric string
	bytes := make([]byte, codeLen)
	charsLen := big.NewInt(int64(len(alphanumericChars)))
	for i := range bytes {
		idx, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random char: %w", err)
		}
		bytes[i] = alphanumericChars[idx.Int64()]
	}

	return string(bytes), nil
}
