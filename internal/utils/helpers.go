package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

// DetectProjectRoot walks up from current dir to find the repo root (with .git)
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

func FormAbsolutePathToDir(projectRoot, srcDir, subdir string) string {
	return filepath.Join(projectRoot, srcDir, subdir)

}

// returns absolute path to targets
func FormAbsolutePathToTargetDirs(absPath string, targets []string) ([]string, error) {
	var resolved []string

	if len(targets) > 0 {
		// Specific targets passed via --target
		for _, t := range targets {
			full := filepath.Join(absPath, t)
			if _, err := os.Stat(full); os.IsNotExist(err) {
				return nil, fmt.Errorf("target %s does not exist at path %s", t, full)
			}
			resolved = append(resolved, full)
		}
	} else {
		// No targets passed: resolve all directories in servicesDir
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read services dir: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				resolved = append(resolved, filepath.Join(absPath, entry.Name()))
			}
		}
	}

	return resolved, nil
}
func HasBin(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
