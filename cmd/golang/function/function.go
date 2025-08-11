package function

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/cmd/golang/function/run"
)

type FunctionOptions struct{}

// Activate this in case we pass value to it.
// var functionCmdDefaults = &FunctionOptions{}

var FunctionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage Go cloud-functions (mod, vendor, build, clean)",
}

func init() {
	FunctionCmd.AddCommand(run.RunCmd)
}
