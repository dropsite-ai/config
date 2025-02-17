package config

import (
	"fmt"
	"net/url"
	"regexp"
)

// usernameRegex enforces Linux username restrictions:
// - 1 to 32 characters,
// - starts with a lowercase letter or underscore,
// - contains only lowercase letters, digits, underscores, or dashes.
var usernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

// validateUsername checks if the provided username matches Linux username restrictions.
func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username: %q", username)
	}
	return nil
}

// validateURL verifies that the string is a valid URL (with a scheme and host).
func ValidateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL: %q", u)
	}
	return nil
}
