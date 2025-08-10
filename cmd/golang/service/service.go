package service

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/cmd/golang/service/build"
	"github.com/selimacerbas/flow-cli/cmd/golang/service/run"
)

type ServiceCmdOptions struct{}

// Activate this in case we pass value to it.
// var serviceCmdDefaults = &ServiceOptions{}

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage Go container services.",
}

func init() {
	ServiceCmd.AddCommand(run.RunCmd)
	ServiceCmd.AddCommand(build.BuildCmd)
}
