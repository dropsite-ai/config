package config

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// Load reads a YAML config from the given path into a value of type T.
// If the file doesn't exist, it writes the provided defaultConfig to that path
// and returns it. It also processes string fields ending in "Secret", "Path",
// "User", or "URL".
func Load[T any](yamlPath string, defaultConfig T) (T, error) {
	expandedPath, err := ExpandPath(yamlPath)
	if err != nil {
		return defaultConfig, err
	}

	if _, err = os.Stat(expandedPath); os.IsNotExist(err) {
		// Process defaultConfig to apply custom logic.
		if err = Process(&defaultConfig); err != nil {
			return defaultConfig, err
		}
		if err = Save(expandedPath, defaultConfig); err != nil {
			return defaultConfig, err
		}
		log.Printf("Generated new config at %s", expandedPath)
		return defaultConfig, nil
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return defaultConfig, err
	}

	var cfg T
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig, err
	}

	// Process the loaded config for custom logic.
	if err := Process(&cfg); err != nil {
		return defaultConfig, err
	}

	return cfg, nil
}

// Save writes the provided config (cfg) to the YAML file at yamlPath.
func Save(yamlPath string, cfg interface{}) error {
	expandedPath, err := ExpandPath(yamlPath)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(expandedPath, data, 0600)
}

// Process inspects the struct (pointed to by cfg) for exported string fields
// with names ending in "Secret", "Path", "User", or "URL" and applies custom logic.
// - "Secret": generates a new secret if empty.
// - "Path": expands the path (e.g., "~" becomes the home directory).
// - "User": validates the username.
// - "URL": validates that the string is a well-formed URL.
func Process(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return errors.New("processConfig expects a pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		// Process only exported string fields.
		if field.Kind() == reflect.String && field.CanSet() {
			val := field.String()

			// If field name ends with "Secret", generate a secret if empty.
			if strings.HasSuffix(fieldType.Name, "Secret") {
				if val == "" {
					newSecret, err := GenerateJWTSecret()
					if err != nil {
						return err
					}
					field.SetString(newSecret)
					val = newSecret
				}
			}

			// If field name ends with "Path", expand the path.
			if strings.HasSuffix(fieldType.Name, "Path") {
				if val != "" {
					expanded, err := ExpandPath(val)
					if err != nil {
						return err
					}
					field.SetString(expanded)
					val = expanded
				}
			}

			// If field name ends with "User", validate the username.
			if strings.HasSuffix(fieldType.Name, "User") {
				if val != "" {
					if err := ValidateUsername(val); err != nil {
						return err
					}
				}
			}

			// If field name ends with "URL", validate that it is a well-formed URL.
			if strings.HasSuffix(fieldType.Name, "URL") {
				if val != "" {
					if err := ValidateURL(val); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
