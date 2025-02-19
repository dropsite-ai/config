package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProcess_SimpleStruct(t *testing.T) {
	type myConfig struct {
		TestSecret string
		TestPath   string
		TestUser   string
		TestURL    string
		Ignored    int
	}

	cfg := myConfig{
		TestSecret: "",
		TestPath:   "~/some/path",
		TestUser:   "myuser",
		TestURL:    "http://example.com",
		Ignored:    42,
	}

	if err := Process(&cfg); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}

	// Secret should be auto-generated if empty
	if cfg.TestSecret == "" {
		t.Errorf("Expected TestSecret to be generated, got empty string")
	}

	// Path should be expanded
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, "some/path")
	if cfg.TestPath != expected {
		t.Errorf("Expected TestPath = %q, got %q", expected, cfg.TestPath)
	}

	// User & URL shouldn't have changed if they are valid
	if cfg.TestUser != "myuser" {
		t.Errorf("Expected TestUser to remain %q, got %q", "myuser", cfg.TestUser)
	}
	if cfg.TestURL != "http://example.com" {
		t.Errorf("Expected TestURL to remain http://example.com, got %q", cfg.TestURL)
	}
}

func TestProcess_NestedStruct(t *testing.T) {
	type subConfig struct {
		DbSecret string
		DbURL    string
	}

	type topConfig struct {
		Name         string
		Nested       subConfig
		AnotherValue int
	}

	cfg := topConfig{
		Name: "top-level",
		Nested: subConfig{
			DbSecret: "",
			DbURL:    "https://mydb.local",
		},
		AnotherValue: 100,
	}

	if err := Process(&cfg); err != nil {
		t.Fatalf("Process returned error for nested struct: %v", err)
	}

	// DbSecret should be auto-generated
	if cfg.Nested.DbSecret == "" {
		t.Error("Expected DbSecret to be generated in nested struct, got empty")
	}

	// URL is valid, should remain unchanged
	if cfg.Nested.DbURL != "https://mydb.local" {
		t.Errorf("Expected DbURL to remain unchanged, got %q", cfg.Nested.DbURL)
	}
}

func TestProcess_NestedMap(t *testing.T) {
	// A generic map with nested maps and some suffix-based keys
	cfg := map[string]interface{}{
		"userURL":  "http://example.org",
		"somePath": "~/myapp",
		"nested": map[string]interface{}{
			"innerSecret": "",
			"invalidUser": "INVALID", // This should fail validation
		},
	}

	err := Process(cfg) // pass map directly
	if err == nil {
		t.Errorf("Expected error for invalid username, got nil")
	}

	// Fix the invalid user, try again
	cfg["nested"].(map[string]interface{})["invalidUser"] = "validuser"

	err = Process(cfg)
	if err != nil {
		t.Fatalf("Process returned error after fixing user: %v", err)
	}

	// userURL is valid, so it should remain the same
	if got := cfg["userURL"]; got != "http://example.org" {
		t.Errorf("Expected userURL = http://example.org, got %v", got)
	}

	// somePath should have been expanded
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, "myapp")
	if got := cfg["somePath"]; got != expected {
		t.Errorf("Expected somePath = %q, got %v", expected, got)
	}

	// innerSecret should have been generated
	if cfg["nested"].(map[string]interface{})["innerSecret"] == "" {
		t.Errorf("Expected innerSecret to be generated, got empty string")
	}
}

func TestProcess_SliceOfMaps(t *testing.T) {
	cfg := []interface{}{
		map[string]interface{}{
			"backendURL": "http://backend.local",
			"someSecret": "",
		},
		map[string]interface{}{
			"dataPath": "~/data",
			"someUser": "alice",
		},
	}

	err := Process(&cfg)
	if err != nil {
		t.Fatalf("Process returned error for slice of maps: %v", err)
	}

	// In first map: someSecret should be generated
	first := cfg[0].(map[string]interface{})
	if first["someSecret"] == "" {
		t.Error("Expected 'someSecret' to be generated, got empty")
	}
	// backendURL is valid, so no change
	if first["backendURL"] != "http://backend.local" {
		t.Errorf("Expected backendURL = http://backend.local, got %v", first["backendURL"])
	}

	// In second map: dataPath should be expanded, someUser is valid user
	second := cfg[1].(map[string]interface{})
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, "data")
	if second["dataPath"] != expected {
		t.Errorf("Expected dataPath = %q, got %v", expected, second["dataPath"])
	}
	if second["someUser"] != "alice" {
		t.Errorf("Expected someUser to remain 'alice', got %v", second["someUser"])
	}
}

func TestProcess_InvalidURL(t *testing.T) {
	cfg := map[string]interface{}{
		"badURL": "://invalid-url",
	}
	err := Process(cfg)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestProcess_NilPointer(t *testing.T) {
	// A nil pointer should just no-op and not crash
	var ptr *string
	if err := Process(ptr); err != nil {
		t.Errorf("Process(nil pointer) returned unexpected error: %v", err)
	}
}
