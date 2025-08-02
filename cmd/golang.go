package cmd

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/cmd/golang"
)

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Commands for Go projects (functions, images)",
	Long:  "Manage Go cloud-functions and container images with one tool",
}

func init() {
	goCmd.AddCommand(golang.GoFunctionCmd)
	// goCmd.AddCommand(golang.GoImageCmd)
}
