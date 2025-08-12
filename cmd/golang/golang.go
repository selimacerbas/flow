package golang

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/cmd/golang/build"
	"github.com/selimacerbas/flow/cmd/golang/run"
)

var GoCmd = &cobra.Command{
	Use:   "go",
	Short: "Commands for Go projects (functions, services)",
	Long:  "Manage Go cloud-functions and container images with one tool",
}

func init() {
	GoCmd.AddCommand(run.RunCmd)
	GoCmd.AddCommand(build.BuildCmd)
}
