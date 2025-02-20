package config

import (
	"fmt"
	"net/url"
	"regexp"
)

var usernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

// validateUsername ensures Linux-style: 1–32 chars, start [a-z_], then [a-z0-9_-].
func validateUsername(name string) error {
	if !usernameRegex.MatchString(name) {
		return fmt.Errorf("username %q is invalid (must match [a-z_][a-z0-9_-]{0,31})", name)
	}
	return nil
}

// validateURL checks the URL has a non-empty scheme and host.
func validateURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL: %q", u)
	}
	return nil
}
