package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a leading "~" in file paths.
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
