package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

type testConfig struct {
	TestSecret string
	TestPath   string
	TestUser   string
	TestURL    string
	Ignored    int
}

func TestSaveLoad_NewFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "config.yaml")

	// Ensure file does not exist.
	if _, err := os.Stat(filePath); err == nil {
		t.Fatalf("Temp file %s already exists", filePath)
	}
	defaultCfg := testConfig{
		TestSecret: "",
		TestPath:   "~/new/path",
		TestUser:   "user",
		TestURL:    "http://example.com",
		Ignored:    100,
	}
	_, err := Load(filePath, defaultCfg)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	// The file should have been created.
	if _, err = os.Stat(filePath); err != nil {
		t.Errorf("Expected file %s to be created", filePath)
	}
	// Load again to test the branch where the file exists.
	cfg2, err := Load(filePath, defaultCfg)
	if err != nil {
		t.Fatalf("Second Load returned error: %v", err)
	}
	if cfg2.TestUser != "user" || cfg2.TestURL != "http://example.com" || cfg2.Ignored != 100 {
		t.Errorf("Loaded config does not match expected values")
	}
	if cfg2.TestSecret == "" {
		t.Errorf("Expected TestSecret to be set in loaded config")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "save.yaml")
	cfg := testConfig{
		TestSecret: "secret",
		TestPath:   "/some/path",
		TestUser:   "user",
		TestURL:    "http://example.com",
		Ignored:    200,
	}
	if err := Save(filePath, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	var loaded testConfig
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal saved file: %v", err)
	}
	if !reflect.DeepEqual(cfg, loaded) {
		t.Errorf("Saved and loaded config differ:\nexpected: %+v\ngot: %+v", cfg, loaded)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "invalid.yaml")
	invalidContent := []byte("::: not valid yaml :::")
	if err := os.WriteFile(filePath, invalidContent, 0600); err != nil {
		t.Fatalf("Failed to write invalid YAML file: %v", err)
	}
	defaultCfg := testConfig{}
	_, err := Load(filePath, defaultCfg)
	if err == nil {
		t.Error("Expected error when loading invalid YAML, got nil")
	}
}
