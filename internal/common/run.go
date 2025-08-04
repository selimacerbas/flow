package common

import (
	"fmt"
	"os"
	"os/exec"
)

func RunCustomCommand(targetDirs []string, command string) error {
	for _, dir := range targetDirs {
		fmt.Printf("Target directory: %s\n", dir)
		fmt.Printf("Command: %s\n", command)

		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed in %s: %w", dir, err)
		}
	}
	return nil
}
