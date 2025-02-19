package config

import (
	"testing"
)

type Credentials struct {
	APIKey string
	Secret string
}

type NestedConfig struct {
	Credentials Credentials
	Count       int
}

type Config struct {
	Variables NestedConfig
	Name      string
}

func TestCopyProperty_ValidNested(t *testing.T) {
	src := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "my-api-key",
				Secret: "my-secret",
			},
			Count: 42,
		},
		Name: "source",
	}
	dst := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "",
				Secret: "",
			},
			Count: 0,
		},
		Name: "dest",
	}

	// Copy a nested string field.
	err := CopyProperty(&src, "Variables.Credentials.APIKey", &dst, "Variables.Credentials.APIKey")
	if err != nil {
		t.Fatalf("Unexpected error copying nested field: %v", err)
	}
	if dst.Variables.Credentials.APIKey != "my-api-key" {
		t.Errorf("Expected APIKey to be 'my-api-key', got %q", dst.Variables.Credentials.APIKey)
	}

	// Copy a nested int field.
	err = CopyProperty(&src, "Variables.Count", &dst, "Variables.Count")
	if err != nil {
		t.Fatalf("Unexpected error copying nested field: %v", err)
	}
	if dst.Variables.Count != 42 {
		t.Errorf("Expected Count to be 42, got %d", dst.Variables.Count)
	}
}

func TestCopyProperty_InvalidSourceField(t *testing.T) {
	src := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "my-api-key",
				Secret: "my-secret",
			},
			Count: 42,
		},
		Name: "source",
	}
	dst := Config{}

	// Attempt to copy from a non-existent source field.
	err := CopyProperty(&src, "Variables.Credentials.NonExistent", &dst, "Variables.Credentials.APIKey")
	if err == nil {
		t.Error("Expected error for invalid source field, got nil")
	}
}

func TestCopyProperty_InvalidDestinationField(t *testing.T) {
	src := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "my-api-key",
				Secret: "my-secret",
			},
			Count: 42,
		},
		Name: "source",
	}
	dst := Config{}

	// Attempt to copy to a non-existent destination field.
	err := CopyProperty(&src, "Variables.Credentials.APIKey", &dst, "Variables.Credentials.NonExistent")
	if err == nil {
		t.Error("Expected error for invalid destination field, got nil")
	}
}

func TestCopyProperty_TypeMismatch(t *testing.T) {
	src := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "my-api-key",
				Secret: "my-secret",
			},
			Count: 42,
		},
		Name: "source",
	}
	dst := Config{
		Variables: NestedConfig{
			Credentials: Credentials{
				APIKey: "",
				Secret: "",
			},
			Count: 0,
		},
		Name: "dest",
	}

	// Attempt to copy an int field into a string field.
	err := CopyProperty(&src, "Variables.Count", &dst, "Variables.Credentials.APIKey")
	if err == nil {
		t.Error("Expected error due to type mismatch, got nil")
	}
}

func TestCopyProperty_NonStructInput(t *testing.T) {
	// Passing a non-struct pointer should return an error.
	type NonStruct int
	var ns NonStruct = 5
	dst := Config{}
	err := CopyProperty(&ns, "NonExistent", &dst, "Variables.Credentials.APIKey")
	if err == nil {
		t.Error("Expected error for non-struct source, got nil")
	}
}
