package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	expanded, err := ExpandPath("~")
	if err != nil {
		t.Fatalf("ExpandPath returned error: %v", err)
	}
	if expanded != home {
		t.Errorf("Expected %q, got %q", home, expanded)
	}

	expanded, err = ExpandPath("~/folder")
	if err != nil {
		t.Fatalf("ExpandPath returned error: %v", err)
	}
	expected := filepath.Join(home, "folder")
	if expanded != expected {
		t.Errorf("Expected %q, got %q", expected, expanded)
	}
}

func TestExpandPath_WithoutTilde(t *testing.T) {
	path := "/some/path"
	expanded, err := ExpandPath(path)
	if err != nil {
		t.Fatalf("ExpandPath returned error: %v", err)
	}
	if expanded != path {
		t.Errorf("Expected %q, got %q", path, expanded)
	}
}
