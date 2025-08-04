package golang

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/cmd/golang/function"
)

var GoCmd = &cobra.Command{
	Use:   "go",
	Short: "Commands for Go projects (functions, services)",
	Long:  "Manage Go cloud-functions and container images with one tool",
}

func init() {
	GoCmd.AddCommand(function.GoFunctionCmd)
	// GoCmd.AddCommand(golang.GoImageCmd)
}
