package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/config"
)

type RootCmd struct {
	Config string
}

var rootCmdDefaults = &RootCmd{
	Config: "",
}

var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "flow is a CLI for multi-language dependency management and automation",
	Long:  `flow is a tool to manage dependencies, build, inject, and image tasks across local repositories and runners, supporting multiple languages (e.g. Go, Python).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		d := rootCmdDefaults

		config.LoadConfig(d.Config)
	},
}

func Execute() {
	rootCmd.AddCommand(goCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	d := rootCmdDefaults
	pf := rootCmd.PersistentFlags()

	pf.StringVar(&d.Config, "config", d.Config, "Path to config file (default is flow.yaml or at project root)")
	config.SetDefaults()

}
