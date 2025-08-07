package service

import (
	"fmt"
	"os"
	"path/filepath"
)

func FormAbsolutePathToServicesDir(projectRoot, srcDir, servicesSubdir string) string {
	return filepath.Join(projectRoot, srcDir, servicesSubdir)

}

// returns absolute path to target services
func FormAbsolutePathToServiceTargetDirs(servicesDir string, targets []string) ([]string, error) {
	var resolved []string

	if len(targets) > 0 {
		// Specific targets passed via --target
		for _, t := range targets {
			full := filepath.Join(servicesDir, t)
			if _, err := os.Stat(full); os.IsNotExist(err) {
				return nil, fmt.Errorf("target %s does not exist at path %s", t, full)
			}
			resolved = append(resolved, full)
		}
	} else {
		// No targets passed: resolve all directories in servicesDir
		entries, err := os.ReadDir(servicesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read services dir: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				resolved = append(resolved, filepath.Join(servicesDir, entry.Name()))
			}
		}
	}

	return resolved, nil
}
