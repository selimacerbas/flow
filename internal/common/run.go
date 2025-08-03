package common

import (
	"fmt"
	"os"
	"os/exec"
)

func RunCustomCommand(targetDirs []string, command string, dryRun, verbose bool) error {
	for _, dir := range targetDirs {
		fmt.Printf("Target directory: %s\n", dir)
		fmt.Printf("Command: %s\n", command)

		if dryRun {
			fmt.Println("(dry-run) Skipping execution.")
			continue
		}

		if verbose {
			fmt.Println("Executing...")
		}

		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = dir

		if verbose {
			// show stdout/stderr live
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			// silence command output unless it fails
			cmd.Stdout = nil
			cmd.Stderr = nil
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed in %s: %w", dir, err)
		}
	}
	return nil
}
