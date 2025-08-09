package golang

import (
	"fmt"
	"os"
	"os/exec"
)

func RunGoClean(targetsDir []string) error {
	for _, dir := range targetsDir {
		fmt.Println("Running go clean . in", dir)
		cmd := exec.Command("go", "clean", ".")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go clean . failed %s: %w", dir, err)
		}
	}
	return nil
}
func RunGoMod(targetDirs []string) error {
	for _, dir := range targetDirs {
		fmt.Println("Running go mod tidy in", dir)
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go mod tidy failed %s: %w", dir, err)
		}
	}
	return nil
}

func RunGoVendor(targetsDir []string) error {
	for _, dir := range targetsDir {
		fmt.Printf("Running `go mod vendor` in %s\n", dir)
		cmd := exec.Command("go", "mod", "vendor")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run go mod vendor in %s: %w", dir, err)
		}
	}
	return nil
}

func RunGoBuild(targetDirs []string, goos, goarch string) error {
	for _, dir := range targetDirs {
		fmt.Printf("→ Building function in %s [GOOS=%s, GOARCH=%s]\n", dir, goos, goarch)

		cmd := exec.Command("go", "build", ".")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build failed in %s: %w", dir, err)
		}
	}
	return nil
}
