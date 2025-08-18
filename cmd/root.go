package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/selimacerbas/flow/cmd/commit"
	"github.com/selimacerbas/flow/cmd/get"
	"github.com/selimacerbas/flow/cmd/golang"

	"github.com/selimacerbas/flow/internal/config"
)

// This will be overwritten at link time by -ldflags.
// Keep it unexported (lowercase) – that's fine.
var version = "dev"

type RootCmd struct {
	Config          string
	SrcDir          string
	FunctionsSubdir string
	ServicesSubdir  string
}

var defaults = &RootCmd{
	Config:          "",
	SrcDir:          "",
	FunctionsSubdir: "",
	ServicesSubdir:  "",
}

var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "flow is a CLI for multi-language dependency management and automation",
	Long:  `flow is a tool to manage dependencies, build, inject, and image tasks across local repositories and runners, supporting multiple languages (e.g. Go, Python).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		d := defaults

		config.LoadConfig(d.Config)
	},
}

func Execute() {
	rootCmd.AddCommand(
		golang.GoCmd,
		get.GetCmd,
		commit.CommitCmd,
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	d := defaults
	pf := rootCmd.PersistentFlags()

    pf.StringVar(&d.Config, "config", d.Config, "Path to config file. Defaults to 'flow.yaml' in the repo root.")

    pf.StringVar(&d.SrcDir, "src-dir", d.SrcDir, "Root source directory. Reads from config key 'dirs.src'.")
    pf.StringVar(&d.FunctionsSubdir, "functions-subdir", d.FunctionsSubdir, "Subdirectory under src for cloud functions. Reads from 'dirs.functions_subdir'.")
    pf.StringVar(&d.ServicesSubdir, "services-subdir", d.ServicesSubdir, "Subdirectory under src for services. Reads from 'dirs.services_subdir'.")

	config.SetDefaults()

	_ = viper.BindPFlag("dirs.src", pf.Lookup("src-dir"))
	_ = viper.BindPFlag("dirs.functions_subdir", pf.Lookup("functions-subdir"))
	_ = viper.BindPFlag("dirs.services_subdir", pf.Lookup("services-subdir"))

}
