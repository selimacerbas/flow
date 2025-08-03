package goexec

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

func ResolveGoPrivate(flagHosts []string) []string {
	if len(flagHosts) > 0 {
		return flagHosts
	}

	fromConfig := viper.GetStringSlice("go.private_hosts")
	if len(fromConfig) > 0 {
		return fromConfig
	}

	if existing := os.Getenv("GOPRIVATE"); existing != "" {
		fmt.Println("→ No private hosts from flags or config. Using existing GOPRIVATE:", existing)
	} else {
		fmt.Println("→ No private hosts provided and GOPRIVATE is not set.")
	}

	return nil
}

func ResolveENVGoOS(flagGoOS string) string {
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

func ResolveENVGoArch(flagGoArch string) string {
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
