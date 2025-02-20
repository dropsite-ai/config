package config

import (
	"crypto/rand"
	"encoding/hex"
)

// generateJWTSecret returns a 32-byte cryptographically random key in hex.
func generateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
