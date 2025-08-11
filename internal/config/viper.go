package config

import (
	"fmt"
	"os"

	"github.com/selimacerbas/flow/internal/utils"
	"github.com/spf13/viper"
)

// SetDefaults establishes all fallback defaults.
func SetDefaults() {
	// paths
	// viper.SetDefault("dirs.src", "src")
	// viper.SetDefault("dirs.functions_subdir", "cloud-functions")
	// viper.SetDefault("dirs.containers_subdir", "cloud-runs")

	// Go defaults
	// viper.SetDefault("go.os", "linux")
	// viper.SetDefault("go.arch", "amd64")

	// image defaults
	viper.SetDefault("image.tag", "latest")

	// cloud build defaults
	viper.SetDefault("cloud.provider", "")
	viper.SetDefault("cloud.region", "")
	viper.SetDefault("cloud.project", "")
	viper.SetDefault("cloud.repository", "")
}

// LoadConfig reads ./flow.* or ~/.flow.* + ENV (with prefix FLOW_).
func LoadConfig(flagConfigFilePath string) {
	viper.SetEnvPrefix("FLOW")
	viper.AutomaticEnv()

	if flagConfigFilePath != "" {
		viper.SetConfigFile(flagConfigFilePath)
	} else {
		// Only check project root (not current dir or home)
		if repoRoot, err := utils.DetectProjectRoot(); err == nil {
			viper.AddConfigPath(repoRoot) // Tries to find .git folder.
			viper.AddConfigPath(".")      // Falbacks to current dir.
			viper.SetConfigName("flow")
			viper.SetConfigType("yaml")
		} else {
			fmt.Fprintln(os.Stderr, " Could not detect project root, no config will be loaded.")
			return
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "→ Using config file:", viper.ConfigFileUsed())
	}
}
