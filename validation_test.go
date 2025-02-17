package config

import "testing"

func TestValidateUsername_Valid(t *testing.T) {
	validUsernames := []string{
		"a",
		"user",
		"abc_123",
		"a123456789012345678901234567890", // 31 characters after the first
	}
	for _, username := range validUsernames {
		if err := ValidateUsername(username); err != nil {
			t.Errorf("Expected username %q to be valid, got error: %v", username, err)
		}
	}
}

func TestValidateUsername_Invalid(t *testing.T) {
	invalidUsernames := []string{
		"User",  // uppercase letter not allowed
		"1user", // must start with letter or underscore
		"a very long username that exceeds the limit",
		"invalid!",
	}
	for _, username := range invalidUsernames {
		if err := ValidateUsername(username); err == nil {
			t.Errorf("Expected username %q to be invalid", username)
		}
	}
}

func TestValidateURL_Valid(t *testing.T) {
	validURLs := []string{
		"http://example.com",
		"https://example.com/path",
	}
	for _, u := range validURLs {
		if err := ValidateURL(u); err != nil {
			t.Errorf("Expected URL %q to be valid, got error: %v", u, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	invalidURLs := []string{
		"http://",
		"not a url",
		"://missing.scheme.com",
	}
	for _, u := range invalidURLs {
		if err := ValidateURL(u); err == nil {
			t.Errorf("Expected URL %q to be invalid", u)
		}
	}
}
