package goexec

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

func SetENVGOPrivate(privateSSHHosts, privateHTTPSHosts []string) error {
	privateHosts := append(privateSSHHosts, privateHTTPSHosts...)

	if len(privateHosts) == 0 {
		fmt.Println("No private hosts provided; using existing GOPRIVATE.")
		return nil
	}

	comaSeparatedHosts := strings.Join(privateHosts, ",")

	fmt.Println("Setting GOPRIVATE to:", comaSeparatedHosts)
	err := os.Setenv("GOPRIVATE", comaSeparatedHosts)
	if err != nil {
		return fmt.Errorf("failed to set GOPRIVATE: %w", err)
	}

	return nil
}

func SetENVGoOS(flagGoOS string) string {
	if flagGoOS != "" {
		return flagGoOS
	}
	if fromConfig := viper.GetString("go.os"); fromConfig != "" {
		return fromConfig
	}
	if fromEnv := os.Getenv("GOOS"); fromEnv != "" {
		return fromEnv
	}
	return runtime.GOOS // fallback to current system default
}

func SetENVGoArch(flagGoArch string) string {
	if flagGoArch != "" {
		return flagGoArch
	}
	if fromConfig := viper.GetString("go.arch"); fromConfig != "" {
		return fromConfig
	}
	if fromEnv := os.Getenv("GOARCH"); fromEnv != "" {
		return fromEnv
	}
	return runtime.GOARCH // fallback to current arch
}
