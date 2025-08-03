package common

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func ResolveGitAuthHTTPSCredentials(flagUsername, flagToken string) (string, string, error) {
	username := flagUsername
	token := flagToken

	if username == "" {
		username = viper.GetString("git.username")
	}
	if token == "" {
		token = viper.GetString("git.token")
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if username == "" || token == "" {
		return "", "", fmt.Errorf("--git-username and --git-token is required when configuring private HTTPS hosts. consider passsing via flag, config or ENV")
	}

	return username, token, nil
}

func ResolvePrivateHTTPSHosts(flagHTTPS []string) []string {
	var hostsHTTPS []string

	if len(flagHTTPS) > 0 {
		hostsHTTPS = flagHTTPS
	} else {
		hostsHTTPS = viper.GetStringSlice("git.private_https_hosts")
	}

	return hostsHTTPS
}

func ResolvePrivateSSHHosts(flagSSH []string) []string {
	var hostsSSH []string

	// Use flags if passed; fallback to config for missing ones
	if len(flagSSH) > 0 {
		hostsSSH = flagSSH
	} else {
		hostsSSH = viper.GetStringSlice("git.private_ssh_hosts")
	}

	return hostsSSH
}
func ResolveAuthMethod(flagMethod string) (string, error) {
	var method string

	if flagMethod != "" {
		method = flagMethod
	} else {
		method = viper.GetString("git.auth_method")
	}

	if method != "ssh" && method != "https" {
		return "", fmt.Errorf("invalid auth-method: '%s' (expected 'ssh' or 'https')", method)
	}

	return method, nil
}
