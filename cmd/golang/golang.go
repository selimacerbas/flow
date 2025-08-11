package golang

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/cmd/golang/function"
	"github.com/selimacerbas/flow/cmd/golang/service"
)

var GoCmd = &cobra.Command{
	Use:   "go",
	Short: "Commands for Go projects (functions, services)",
	Long:  "Manage Go cloud-functions and container images with one tool",
}

func init() {
	GoCmd.AddCommand(function.FunctionCmd)
	GoCmd.AddCommand(service.ServiceCmd)
}
