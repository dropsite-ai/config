package config

import (
	"fmt"
	"os"

	"github.com/dropsite-ai/yamledit"
	"gopkg.in/yaml.v3"
)

// Variables holds the processed maps.
type Variables struct {
	Endpoints map[string]string
	Secrets   map[string]string
	Users     map[string]string
	Paths     map[string]string
}

// CallbackDefinition represents one callback definition.
type CallbackDefinition struct {
	Name      string         `yaml:"name"`
	Events    []string       `yaml:"events"`
	Timing    string         `yaml:"timing"` // expected to be "pre" or "post"
	Target    CallbackTarget `yaml:"target"`
	Endpoints []string       `yaml:"endpoints"`
}

// CallbackTarget represents a callback's target.
type CallbackTarget struct {
	Type string `yaml:"type"` // expected to be "file" or "directory"
	Path string `yaml:"path"`
}

// ProcessCallbacks accepts a YAML node and a prefix indicating where an array of CallbackDefinition structs
// is located. It reads and validates the definitions and returns them.
// If the section is missing, an empty slice is returned.
func ProcessCallbacks(doc *yaml.Node, prefix string, vars *Variables) ([]CallbackDefinition, error) {
	var callbacks []CallbackDefinition

	// Attempt to read the callbacks slice at the given prefix.
	if err := yamledit.ReadNode(doc, prefix, &callbacks); err != nil {
		// If the section doesn't exist, return an empty slice without error.
		return []CallbackDefinition{}, nil
	}

	// Validate each callback.
	for _, cb := range callbacks {
		if cb.Timing != "pre" && cb.Timing != "post" {
			return nil, fmt.Errorf("invalid timing for callback %q: %q", cb.Name, cb.Timing)
		}
		if cb.Target.Type != "file" && cb.Target.Type != "directory" {
			return nil, fmt.Errorf("invalid target type for callback %q: %q", cb.Name, cb.Target.Type)
		}
		// Validate that each endpoint key exists in the provided Variables map.
		for _, epKey := range cb.Endpoints {
			if _, exists := vars.Endpoints[epKey]; !exists {
				return nil, fmt.Errorf("callback %q refers to unknown endpoint key %q", cb.Name, epKey)
			}
		}
	}

	return callbacks, nil
}

// ProcessVariables accepts a YAML node and a prefix (e.g. "variables" or "custom") indicating
// where the maps are located. It processes each section and returns a new Variables struct without
// modifying the original YAML node.
func ProcessVariables(doc *yaml.Node, prefix string) (*Variables, error) {
	var vars Variables

	// Process endpoints: validate each URL.
	endpointsPath := prefix + ".endpoints"
	if err := yamledit.ReadNode(doc, endpointsPath, &vars.Endpoints); err == nil {
		for key, endpoint := range vars.Endpoints {
			if err := validateURL(endpoint); err != nil {
				return nil, fmt.Errorf("invalid endpoint for %q: %v", key, err)
			}
		}
	}

	// Process secrets: generate a secret if the value is empty, and update the YAML node.
	secretsPath := prefix + ".secrets"
	var secretsMap map[string]string
	var secretsNode yaml.Node
	// Read both the mapping into a Go map and also keep the YAML node.
	if err := yamledit.ReadNode(doc, secretsPath, &secretsMap); err == nil {
		// Retrieve the YAML node corresponding to the secrets map.
		if err := yamledit.ReadNode(doc, secretsPath, &secretsNode); err != nil {
			return nil, err
		}
		// Process the mapping and update the YAML node.
		// YAML mapping nodes have key/value pairs as sequential elements.
		for i := 0; i < len(secretsNode.Content); i += 2 {
			keyNode := secretsNode.Content[i]
			valueNode := secretsNode.Content[i+1]
			// Check if the secret is empty.
			if valueNode.Value == "" {
				newSecret, err := generateJWTSecret()
				if err != nil {
					return nil, fmt.Errorf("generating secret for %q: %w", keyNode.Value, err)
				}
				// Update the YAML node value.
				valueNode.Value = newSecret
				// Also update the Go map.
				secretsMap[keyNode.Value] = newSecret
			}
		}
		// Assign the modified map to your variables struct.
		vars.Secrets = secretsMap
	}

	// Process users: validate each username.
	usersPath := prefix + ".users"
	if err := yamledit.ReadNode(doc, usersPath, &vars.Users); err == nil {
		for key, username := range vars.Users {
			if err := validateUsername(username); err != nil {
				return nil, fmt.Errorf("invalid username for %q: %v", key, err)
			}
		}
	}

	// Process paths: expand "~" to the user's home directory.
	pathsPath := prefix + ".paths"
	if err := yamledit.ReadNode(doc, pathsPath, &vars.Paths); err == nil {
		for key, p := range vars.Paths {
			expanded, err := ExpandPath(p)
			if err != nil {
				return nil, fmt.Errorf("expanding path for %q: %w", key, err)
			}
			vars.Paths[key] = expanded
		}
	}

	return &vars, nil
}

// Load opens the YAML file at the given path, or if the file is not found,
// uses the provided defaultYAML string. It then parses the content into a document node,
// processes variables and callbacks, and returns the document, Variables, and callbacks.
func Load(path string, defaultYAML []byte) (*yaml.Node, *Variables, []CallbackDefinition, error) {
	yamlBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && len(defaultYAML) != 0 {
			yamlBytes = defaultYAML
		} else {
			return nil, nil, nil, fmt.Errorf("reading YAML file: %w", err)
		}
	}

	doc, err := yamledit.Parse(yamlBytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Process variables under the "variables" key.
	vars, err := ProcessVariables(doc, "variables")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("processing variables: %w", err)
	}

	// Process callbacks under the "callbacks" key.
	callbacks, err := ProcessCallbacks(doc, "callbacks", vars)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("processing callbacks: %w", err)
	}

	return doc, vars, callbacks, nil
}

// Save encodes the provided YAML document and writes it to the specified path.
func Save(path string, doc *yaml.Node) error {
	yamlBytes, err := yamledit.Encode(doc)
	if err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	if err := os.WriteFile(path, yamlBytes, 0644); err != nil {
		return fmt.Errorf("writing YAML file: %w", err)
	}
	return nil
}
