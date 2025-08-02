package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/selimacerbas/flow-cli/internal/utils"
)

// SetDefaults establishes all fallback defaults.
func SetDefaults() {
	// paths
	viper.SetDefault("dirs.src", "src")
	viper.SetDefault("dirs.functions_subdir", "cloud-functions")
	viper.SetDefault("dirs.containers_subdir", "cloud-runs")

	// Go defaults
	viper.SetDefault("go.os", "linux")
	viper.SetDefault("go.arch", "amd64")

	// image defaults
	viper.SetDefault("image.tag", "latest")

	// cloud build defaults
	viper.SetDefault("cloud.provider", "")
	viper.SetDefault("cloud.region", "")
	viper.SetDefault("cloud.project", "")
	viper.SetDefault("cloud.repository", "")
}

// LoadConfig reads ./flow.* or ~/.flow.* + ENV (with prefix FLOW_).

func LoadConfig(cfgFile string) {
	viper.SetEnvPrefix("FLOW")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Add dynamic repo root
		if repoRoot, err := utils.DetectProjectRoot(); err == nil {
			viper.AddConfigPath(repoRoot)
		}

		// fallback to current dir and home dir
		viper.AddConfigPath(".")
		viper.AddConfigPath(os.Getenv("HOME"))

		viper.SetConfigName("flow") // for flow.yaml
		viper.SetConfigType("yaml") // explicit
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "→ Using config file:", viper.ConfigFileUsed())
	}
}
