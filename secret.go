package config

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateJWTSecret creates a secure random secret encoded as hex.
func GenerateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
