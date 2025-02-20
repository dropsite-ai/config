package config

import "testing"

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"root", false},
		{"_system", false},
		{"user123", false},
		{"UPPER", true},
		{"", true},
		{"123abc", true},
	}

	for _, c := range cases {
		err := validateUsername(c.name)
		if (err != nil) != c.wantErr {
			t.Errorf("validateUsername(%q) => error=%v, wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestValidateURL(t *testing.T) {
	cases := []struct {
		url     string
		wantErr bool
	}{
		{"http://example.com", false},
		{"https://example.com/path", false},
		{"://no-scheme", true},
		{"not-a-url", true},
		{"http://", true},
	}
	for _, c := range cases {
		err := validateURL(c.url)
		if (err != nil) != c.wantErr {
			t.Errorf("validateURL(%q) => error=%v, wantErr=%v", c.url, err, c.wantErr)
		}
	}
}
