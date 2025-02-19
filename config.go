package config

import (
	"log"
	"os"

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
