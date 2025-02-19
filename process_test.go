package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProcess_SimpleStruct verifies that a configuration defined as a struct with a Variables field
// is processed correctly.
func TestProcess_SimpleStruct(t *testing.T) {
	type Variables struct {
		Endpoints map[string]string `yaml:"endpoints"`
		Secrets   map[string]string `yaml:"secrets"`
		Users     map[string]string `yaml:"users"`
		Paths     map[string]string `yaml:"paths"`
	}
	type myConfig struct {
		Variables Variables `yaml:"variables"`
		Ignored   int       `yaml:"ignored"`
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	cfg := myConfig{
		Variables: Variables{
			Endpoints: map[string]string{
				"callback": "http://example.com/callback",
			},
			Secrets: map[string]string{
				"api": "",
			},
			Users: map[string]string{
				"owner": "myuser",
			},
			Paths: map[string]string{
				"database": "~/some/path",
			},
		},
		Ignored: 42,
	}

	if err := Process(&cfg); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}

	// The secret should be generated.
	if cfg.Variables.Secrets["api"] == "" {
		t.Errorf("Expected Secrets['api'] to be generated, got empty string")
	}

	// The path should be expanded.
	expectedPath := filepath.Join(home, "some/path")
	if cfg.Variables.Paths["database"] != expectedPath {
		t.Errorf("Expected Paths['database'] to be %q, got %q", expectedPath, cfg.Variables.Paths["database"])
	}

	// Endpoints and users should remain unchanged.
	if cfg.Variables.Endpoints["callback"] != "http://example.com/callback" {
		t.Errorf("Expected Endpoints['callback'] to remain unchanged, got %q", cfg.Variables.Endpoints["callback"])
	}
	if cfg.Variables.Users["owner"] != "myuser" {
		t.Errorf("Expected Users['owner'] to remain unchanged, got %q", cfg.Variables.Users["owner"])
	}
}

// TestProcess_NoVariables confirms that if there is no "variables" section, Process is a no-op.
func TestProcess_NoVariables(t *testing.T) {
	cfg := struct {
		Name string `yaml:"name"`
	}{
		Name: "test",
	}

	if err := Process(&cfg); err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if cfg.Name != "test" {
		t.Errorf("Expected Name to remain 'test', got %q", cfg.Name)
	}
}

// TestProcess_NestedMap uses a map configuration with a "variables" key and tests that:
// - an invalid username in variables.users returns an error,
// - an empty secret in variables.secrets is generated,
// - a path in variables.paths is expanded.
func TestProcess_NestedMap(t *testing.T) {
	cfg := map[string]interface{}{
		"variables": map[string]interface{}{
			"endpoints": map[string]interface{}{
				"service": "http://service.example.com",
			},
			"secrets": map[string]interface{}{
				"token": "",
			},
			"users": map[string]interface{}{
				"admin": "INVALID", // invalid username – should trigger an error
			},
			"paths": map[string]interface{}{
				"config": "~/config",
			},
		},
	}

	err := Process(cfg)
	if err == nil {
		t.Errorf("Expected error for invalid username, got nil")
	}

	// Fix the invalid username and try again.
	vars := cfg["variables"].(map[string]interface{})
	users := vars["users"].(map[string]interface{})
	users["admin"] = "adminuser"

	err = Process(cfg)
	if err != nil {
		t.Fatalf("Process returned error after fixing username: %v", err)
	}

	// Verify endpoints remain unchanged.
	endpoints := vars["endpoints"].(map[string]interface{})
	if endpoints["service"] != "http://service.example.com" {
		t.Errorf("Expected endpoint 'service' to remain unchanged, got %v", endpoints["service"])
	}

	// Verify that the secret is generated.
	secrets := vars["secrets"].(map[string]interface{})
	if secrets["token"] == "" {
		t.Errorf("Expected secret 'token' to be generated, got empty string")
	}

	// Verify that the path is expanded.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	expectedConfigPath := filepath.Join(home, "config")
	paths := vars["paths"].(map[string]interface{})
	if paths["config"] != expectedConfigPath {
		t.Errorf("Expected path 'config' to be %q, got %v", expectedConfigPath, paths["config"])
	}
}

// TestProcess_InvalidURL confirms that an invalid URL in variables.endpoints returns an error.
func TestProcess_InvalidURL(t *testing.T) {
	cfg := map[string]interface{}{
		"variables": map[string]interface{}{
			"endpoints": map[string]interface{}{
				"bad": "://invalid-url",
			},
		},
	}
	err := Process(cfg)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

// TestProcess_NilPointer ensures that passing a nil pointer to Process does not cause a crash.
func TestProcess_NilPointer(t *testing.T) {
	var ptr *string
	if err := Process(ptr); err != nil {
		t.Errorf("Process(nil pointer) returned unexpected error: %v", err)
	}
}

// TestProcess_Slice demonstrates processing when a slice of maps (each with a variables section) is passed.
// Since the new Process function looks for a top-level "variables" key, we must process each slice element.
func TestProcess_Slice(t *testing.T) {
	cfg := []interface{}{
		map[string]interface{}{
			"variables": map[string]interface{}{
				"secrets": map[string]interface{}{
					"key": "",
				},
			},
		},
		map[string]interface{}{
			"name": "test",
		},
	}
	// Process each element in the slice.
	for _, item := range cfg {
		if err := Process(item); err != nil {
			t.Fatalf("Process returned error for slice element: %v", err)
		}
	}
	// Check that in the first element the secret was generated.
	first := cfg[0].(map[string]interface{})
	vars, ok := first["variables"].(map[string]interface{})
	if ok {
		secrets, ok := vars["secrets"].(map[string]interface{})
		if ok && secrets["key"] == "" {
			t.Errorf("Expected secret 'key' to be generated in slice element, got empty")
		}
	}
}
