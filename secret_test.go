package config

import (
	"regexp"
	"testing"
)

func TestGenerateJWTSecret(t *testing.T) {
	secret1, err := GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret returned error: %v", err)
	}
	if len(secret1) != 64 {
		t.Errorf("Expected secret length 64, got %d", len(secret1))
	}
	secret2, err := GenerateJWTSecret()
	if err != nil {
		t.Fatalf("GenerateJWTSecret returned error: %v", err)
	}
	if secret1 == secret2 {
		t.Errorf("Expected different secrets, got same")
	}
	hexRegex := regexp.MustCompile("^[a-f0-9]+$")
	if !hexRegex.MatchString(secret1) {
		t.Errorf("Secret %q is not valid hex", secret1)
	}
}
