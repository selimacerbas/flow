package common

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// expects comaseperated string
func SetGitAuthSSH(hosts string) error {

	hostsList := strings.Split(hosts, ",")
	for _, host := range hostsList {
		fmt.Printf("Configuring SSH for %s\n", host)
		cmd := exec.Command("git", "config", "--global", fmt.Sprintf("url.git@%s:.insteadOf", host), fmt.Sprintf("https://%s/", host))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run exec command for SSH %s: %w", host, err)
		}
	}
	return nil
}

// expects comaseperated string
func SetGitAuthHTTPS(hosts string, username, token string) error {

	hostsList := strings.Split(hosts, ",")
	for _, host := range hostsList {
		fmt.Printf("Configuring HTTPS token auth for %s\n", host)
		authURL := fmt.Sprintf("https://%s:%s@%s/", username, token, host)
		cmd := exec.Command("git", "config", "--global", fmt.Sprintf("url.%s.insteadOf", authURL), fmt.Sprintf("https://%s/", host))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run exec command for HTTPS %s: %w", host, err)
		}
	}

	return nil
}
