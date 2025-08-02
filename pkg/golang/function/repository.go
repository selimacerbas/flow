package function

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func ResolveFunctionsDir(projectRoot, flagSrcDir, flagFunctionsSubdir string) (string, error) {
	// Priority: flag → config → default (already handled via Viper)
	src := flagSrcDir
	if src == "" {
		src = viper.GetString("dirs.src")
	}

	functionsSubdir := flagFunctionsSubdir
	if functionsSubdir == "" {
		functionsSubdir = viper.GetString("dirs.functions_subdir")
	}

	functionsDir := filepath.Join(projectRoot, src, functionsSubdir)

	return functionsDir, nil
}

// returns absolute path to target functions
func ResolveFunctionTargetDirs(functionsDir string, targets []string) ([]string, error) {
	var resolved []string

	if len(targets) > 0 {
		// Specific targets passed via --target
		for _, t := range targets {
			full := filepath.Join(functionsDir, t)
			if _, err := os.Stat(full); os.IsNotExist(err) {
				return nil, fmt.Errorf("target %s does not exist at path %s", t, full)
			}
			resolved = append(resolved, full)
		}
	} else {
		// No targets passed: resolve all directories in functionsDir
		entries, err := os.ReadDir(functionsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read functions dir: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				resolved = append(resolved, filepath.Join(functionsDir, entry.Name()))
			}
		}
	}

	return resolved, nil
}
