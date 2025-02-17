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

func TestProcess_Valid(t *testing.T) {
	cfg := testConfig{
		TestSecret: "",
		TestPath:   "~/some/path",
		TestUser:   "user",
		TestURL:    "http://example.com",
		Ignored:    42,
	}
	err := Process(&cfg)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if cfg.TestSecret == "" {
		t.Errorf("Expected TestSecret to be set, got empty")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, "some/path")
	if cfg.TestPath != expected {
		t.Errorf("Expected TestPath to be %q, got %q", "/some/path", cfg.TestPath)
	}

	if cfg.TestUser != "user" {
		t.Errorf("TestUser changed unexpectedly")
	}
	if cfg.TestURL != "http://example.com" {
		t.Errorf("TestURL changed unexpectedly")
	}
}

func TestProcess_InvalidUser(t *testing.T) {
	cfg := testConfig{
		TestUser: "InvalidUser", // Uppercase letter not allowed by regex.
		TestURL:  "http://example.com",
	}
	err := Process(&cfg)
	if err == nil {
		t.Error("Expected error for invalid username, got nil")
	}
}

func TestProcess_InvalidURL(t *testing.T) {
	cfg := testConfig{
		TestUser: "user",
		TestURL:  "invalid-url",
	}
	err := Process(&cfg)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestProcess_NonStructPointer(t *testing.T) {
	i := 10
	// When passing a pointer to a non-struct, Process returns nil.
	if err := Process(&i); err != nil {
		t.Errorf("Expected nil error when processing pointer to non-struct, got: %v", err)
	}
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
