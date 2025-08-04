package function

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/cmd/golang/function/run"
)

type FunctionCmd struct{}

// Activate this in case we pass value to it.
// var functionCmdDefaults = &FunctionCmd{}

var GoFunctionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage Go cloud-functions (mod, vendor, build, clean)",
}

func init() {
	GoFunctionCmd.AddCommand(run.GoRunCmd)
}
