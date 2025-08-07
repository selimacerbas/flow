package function

import (
	"fmt"
	"os"
	"path/filepath"

)

func FormAbsolutePathToFunctionsDir(projectRoot, srcDir, functionsSubdir string) string {
	return filepath.Join(projectRoot, srcDir, functionsSubdir)

}

// returns absolute path to target functions
func FormAbsolutePathToFunctionTargetDirs(functionsDir string, targets []string) ([]string, error) {
	var formed []string

	if len(targets) > 0 {
		// Specific targets passed via --target
		for _, t := range targets {
			full := filepath.Join(functionsDir, t)
			if _, err := os.Stat(full); os.IsNotExist(err) {
				return nil, fmt.Errorf("target %s does not exist at path %s", t, full)
			}
			formed = append(formed, full)
		}
	} else {
		// No targets passed: resolve all directories in functionsDir
		entries, err := os.ReadDir(functionsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read functions dir: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				formed = append(formed, filepath.Join(functionsDir, entry.Name()))
			}
		}
	}

	return formed, nil
}
