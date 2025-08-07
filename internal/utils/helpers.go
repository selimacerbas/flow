package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DetectProjectRoot walks up from current dir to find the repo root (with go.mod).
func DetectProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}

func ResolveStringValue(flagVal, configKey string, envVars ...string) string {
	if flagVal != "" {
		return flagVal
	}
	if fromConfig := viper.GetString(configKey); fromConfig != "" {
		return fromConfig
	}
	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			return val
		}
	}
	return ""
}

