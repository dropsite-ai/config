package config

import (
	"testing"
)

func TestGenerateJWTSecret(t *testing.T) {
	s1, err := generateJWTSecret()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(s1) != 64 {
		t.Errorf("Expected 64 hex characters, got %d", len(s1))
	}

	s2, _ := generateJWTSecret()
	if s1 == s2 {
		t.Errorf("Expected two calls to generate different secrets, got same: %q", s1)
	}
}
