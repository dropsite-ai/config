package config

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/dropsite-ai/yamledit"
	"gopkg.in/yaml.v3"
)

func TestProcessVariables(t *testing.T) {
	t.Run("Valid config with default prefix", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot determine user home directory")
		}

		yamlStr := `
variables:
  endpoints:
    service1: "http://example.com"
  secrets:
    secret1: ""
    secret2: "existingsecret"
  users:
    user1: "root"
  paths:
    path1: "~"
    path2: "~/folder"
`
		var doc yaml.Node
		if err = yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}

		vars, err := ProcessVariables(&doc, "variables")
		if err != nil {
			t.Fatalf("ProcessVariables returned error: %v", err)
		}

		// Endpoints: unchanged and valid.
		if ep, ok := vars.Endpoints["service1"]; !ok || ep != "http://example.com" {
			t.Errorf("expected endpoint 'http://example.com', got %v", vars.Endpoints["service1"])
		}

		// Secrets: empty secret should be replaced.
		secret1, ok := vars.Secrets["secret1"]
		if !ok || secret1 == "" {
			t.Errorf("secret1 was not updated")
		}
		if len(secret1) != 64 {
			t.Errorf("expected secret1 to be 64 hex characters, got %d", len(secret1))
		}
		if vars.Secrets["secret2"] != "existingsecret" {
			t.Errorf("expected secret2 to remain 'existingsecret', got %v", vars.Secrets["secret2"])
		}

		// Users: valid username.
		if user, ok := vars.Users["user1"]; !ok || user != "root" {
			t.Errorf("expected user1 to be 'root', got %v", vars.Users["user1"])
		}

		// Paths: "~" should be expanded.
		if p, ok := vars.Paths["path1"]; !ok || p != home {
			t.Errorf("expected path1 to be expanded to %q, got %v", home, vars.Paths["path1"])
		}
		expectedPath2 := filepath.Join(home, "folder")
		if p, ok := vars.Paths["path2"]; !ok || p != expectedPath2 {
			t.Errorf("expected path2 to be expanded to %q, got %v", expectedPath2, vars.Paths["path2"])
		}
	})

	t.Run("Invalid endpoint with default prefix", func(t *testing.T) {
		yamlStr := `
variables:
  endpoints:
    service1: "invalid-url"
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		_, err := ProcessVariables(&doc, "variables")
		if err == nil {
			t.Fatalf("expected error due to invalid endpoint, got nil")
		}
		if !regexp.MustCompile(`invalid endpoint`).MatchString(err.Error()) {
			t.Errorf("expected error message to mention 'invalid endpoint', got %v", err)
		}
	})

	t.Run("Invalid username with default prefix", func(t *testing.T) {
		yamlStr := `
variables:
  users:
    user1: "InvalidUser"
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		_, err := ProcessVariables(&doc, "variables")
		if err == nil {
			t.Fatalf("expected error due to invalid username, got nil")
		}
		if !regexp.MustCompile(`invalid username`).MatchString(err.Error()) {
			t.Errorf("expected error message to mention 'invalid username', got %v", err)
		}
	})

	t.Run("Missing sections with default prefix", func(t *testing.T) {
		// YAML without the "variables" key. ProcessVariables should simply skip missing sections.
		yamlStr := `
other:
  key: value
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		vars, err := ProcessVariables(&doc, "variables")
		if err != nil {
			t.Fatalf("ProcessVariables returned error on missing sections: %v", err)
		}
		// Returned maps should be nil (or remain unset) when the section is missing.
		if vars.Endpoints != nil {
			t.Errorf("expected endpoints to be nil, got %v", vars.Endpoints)
		}
		if vars.Secrets != nil {
			t.Errorf("expected secrets to be nil, got %v", vars.Secrets)
		}
		if vars.Users != nil {
			t.Errorf("expected users to be nil, got %v", vars.Users)
		}
		if vars.Paths != nil {
			t.Errorf("expected paths to be nil, got %v", vars.Paths)
		}
	})
}

func TestProcessCallbacks(t *testing.T) {
	t.Run("Valid callbacks with default prefix", func(t *testing.T) {
		yamlStr := `
callbacks:
  - name: "callback1"
    events: ["event1", "event2"]
    timing: "pre"
    target:
      type: "file"
      path: "some/path"
    endpoints: ["service1"]
  - name: "callback2"
    events: ["event3"]
    timing: "post"
    target:
      type: "directory"
      path: "another/path"
    endpoints: ["service2"]
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		// Provide a Variables struct with matching endpoint keys.
		vars := &Variables{
			Endpoints: map[string]string{
				"service1": "http://example.com",
				"service2": "http://example.org",
			},
		}
		callbacks, err := ProcessCallbacks(&doc, "callbacks", vars)
		if err != nil {
			t.Fatalf("ProcessCallbacks returned error: %v", err)
		}
		if len(callbacks) != 2 {
			t.Fatalf("expected 2 callbacks, got %d", len(callbacks))
		}
	})

	t.Run("Invalid timing in callback", func(t *testing.T) {
		yamlStr := `
callbacks:
  - name: "callback1"
    events: ["event1"]
    timing: "middle"
    target:
      type: "file"
      path: "some/path"
    endpoints: []
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		vars := &Variables{
			Endpoints: map[string]string{},
		}
		_, err := ProcessCallbacks(&doc, "callbacks", vars)
		if err == nil {
			t.Fatal("expected error due to invalid timing, got nil")
		}
		if !regexp.MustCompile(`invalid timing`).MatchString(err.Error()) {
			t.Errorf("expected error message to mention 'invalid timing', got %v", err)
		}
	})

	t.Run("Invalid target type in callback", func(t *testing.T) {
		yamlStr := `
callbacks:
  - name: "callback1"
    events: ["event1"]
    timing: "pre"
    target:
      type: "invalid"
      path: "some/path"
    endpoints: []
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		vars := &Variables{
			Endpoints: map[string]string{},
		}
		_, err := ProcessCallbacks(&doc, "callbacks", vars)
		if err == nil {
			t.Fatal("expected error due to invalid target type, got nil")
		}
		if !regexp.MustCompile(`invalid target type`).MatchString(err.Error()) {
			t.Errorf("expected error message to mention 'invalid target type', got %v", err)
		}
	})

	t.Run("Unknown endpoint key in callback", func(t *testing.T) {
		yamlStr := `
callbacks:
  - name: "callback1"
    events: ["event1"]
    timing: "pre"
    target:
      type: "file"
      path: "some/path"
    endpoints: ["nonexistent"]
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		// Provide a Variables struct that does not contain the endpoint "nonexistent".
		vars := &Variables{
			Endpoints: map[string]string{
				"service1": "http://example.com",
			},
		}
		_, err := ProcessCallbacks(&doc, "callbacks", vars)
		if err == nil {
			t.Fatal("expected error due to unknown endpoint key, got nil")
		}
		if !regexp.MustCompile(`unknown endpoint key`).MatchString(err.Error()) {
			t.Errorf("expected error message to mention 'unknown endpoint key', got %v", err)
		}
	})

	t.Run("Missing callbacks section", func(t *testing.T) {
		yamlStr := `
other:
  key: value
`
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
			t.Fatalf("failed to unmarshal YAML: %v", err)
		}
		vars := &Variables{
			Endpoints: map[string]string{},
		}
		callbacks, err := ProcessCallbacks(&doc, "callbacks", vars)
		if err != nil {
			t.Fatalf("ProcessCallbacks returned error on missing callbacks: %v", err)
		}
		if len(callbacks) != 0 {
			t.Errorf("expected no callbacks, got %d", len(callbacks))
		}
	})
}

func TestSecretsNodeUpdate(t *testing.T) {
	yamlStr := `
variables:
  endpoints:
    service1: "http://example.com"
  secrets:
    secret1: ""
    secret2: "existingsecret"
  users:
    user1: "root"
  paths:
    path1: "~"
`
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// ProcessVariables now updates the YAML node for secrets when a secret is empty.
	vars, err := ProcessVariables(&doc, "variables")
	if err != nil {
		t.Fatalf("ProcessVariables returned error: %v", err)
	}

	// Check that the generated secret in the returned Variables is valid.
	secret1, ok := vars.Secrets["secret1"]
	if !ok || secret1 == "" {
		t.Errorf("secret1 was not updated in the Variables map")
	}
	if len(secret1) != 64 {
		t.Errorf("expected secret1 to be 64 hex characters, got %d", len(secret1))
	}

	// Now re-read the secrets mapping from the updated YAML document.
	var updatedSecrets map[string]string
	if err := yamledit.ReadNode(&doc, "variables.secrets", &updatedSecrets); err != nil {
		t.Fatalf("failed to re-read secrets from YAML doc: %v", err)
	}

	// Verify that the YAML node now contains the generated secret.
	if updatedSecrets["secret1"] != secret1 {
		t.Errorf("YAML node not updated: expected %q, got %q", secret1, updatedSecrets["secret1"])
	}

	// Ensure that existing secrets are not overwritten.
	if updatedSecrets["secret2"] != "existingsecret" {
		t.Errorf("expected secret2 to remain 'existingsecret', got %q", updatedSecrets["secret2"])
	}
}

func TestLoad(t *testing.T) {
	// YAML content with variables and callbacks.
	yamlContent := `
variables:
  endpoints:
    service1: "http://example.com"
  secrets:
    secret1: ""
    secret2: "existingsecret"
  users:
    user1: "root"
  paths:
    path1: "~"
callbacks:
  - name: "callback1"
    events: ["event1", "event2"]
    timing: "pre"
    target:
      type: "file"
      path: "some/path"
    endpoints: ["service1"]
`
	// Create a temporary file with the YAML content.
	tmpFile, err := os.CreateTemp("", "config_load_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write YAML content: %v", err)
	}
	tmpFile.Close()

	// Call Load to parse the file.
	_, vars, callbacks, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Check that variables were processed correctly.
	secret1, ok := vars.Secrets["secret1"]
	if !ok || secret1 == "" {
		t.Errorf("expected secret1 to be generated, got empty")
	}
	if len(secret1) != 64 {
		t.Errorf("expected secret1 to be 64 hex characters, got length %d", len(secret1))
	}
	if len(callbacks) != 1 {
		t.Errorf("expected 1 callback, got %d", len(callbacks))
	}
}

func TestSave(t *testing.T) {
	// YAML content with variables (no callbacks needed for Save).
	yamlContent := `
variables:
  endpoints:
    service1: "http://example.com"
  secrets:
    secret1: ""
    secret2: "existingsecret"
  users:
    user1: "root"
  paths:
    path1: "~"
`
	// Create a temporary file with the YAML content.
	tmpFile, err := os.CreateTemp("", "config_save_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write YAML content: %v", err)
	}
	tmpFile.Close()

	// Load the document (which will update secret1 if empty).
	doc, vars, _, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	secret1, ok := vars.Secrets["secret1"]
	if !ok || secret1 == "" {
		t.Fatalf("expected secret1 to be generated")
	}

	// Save the updated document to a new temporary file.
	saveFile, err := os.CreateTemp("", "config_save_output_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file for Save: %v", err)
	}
	saveFileName := saveFile.Name()
	saveFile.Close()
	defer os.Remove(saveFileName)

	if err = Save(saveFileName, doc); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// Read the saved file and re-parse it.
	savedBytes, err := os.ReadFile(saveFileName)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	savedDoc, err := yamledit.Parse(savedBytes)
	if err != nil {
		t.Fatalf("failed to parse saved YAML: %v", err)
	}

	// Verify that the updated secret is persisted in the saved document.
	var savedSecrets map[string]string
	if err := yamledit.ReadNode(savedDoc, "variables.secrets", &savedSecrets); err != nil {
		t.Fatalf("failed to read secrets from saved YAML: %v", err)
	}
	if savedSecrets["secret1"] != secret1 {
		t.Errorf("saved secret1 does not match; expected %q, got %q", secret1, savedSecrets["secret1"])
	}
	if savedSecrets["secret2"] != "existingsecret" {
		t.Errorf("expected secret2 to remain unchanged, got %q", savedSecrets["secret2"])
	}
}
