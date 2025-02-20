package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping path expansion test, no home dir")
	}

	p, err := ExpandPath("~")
	if err != nil {
		t.Fatalf("ExpandPath error: %v", err)
	}
	if p != home {
		t.Errorf("Expected %q, got %q", home, p)
	}

	p2, err := ExpandPath("~/folder")
	if err != nil {
		t.Fatalf("ExpandPath error: %v", err)
	}
	want2 := filepath.Join(home, "folder")
	if p2 != want2 {
		t.Errorf("Expected %q, got %q", want2, p2)
	}

	// Non-tilde path remains unchanged
	p3, err := ExpandPath("/usr/local")
	if err != nil {
		t.Fatalf("ExpandPath error: %v", err)
	}
	if p3 != "/usr/local" {
		t.Errorf("Expected '/usr/local', got %q", p3)
	}
}
