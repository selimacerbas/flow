package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/config"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "flow",
		Short: "flow is a CLI for multi-language dependency management and automation",
		Long:  `flow is a tool to manage dependencies, build, inject, and image tasks across local repositories and runners, supporting multiple languages (e.g. Go, Python).`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadConfig(configFile)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.flow.yaml)")
	config.SetDefaults()
}

func Execute() {
	rootCmd.AddCommand(goCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
