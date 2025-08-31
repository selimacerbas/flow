package changed

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/pkg/get"
)

type ChangedCmdArgOptions struct {
	Ref1 string
	Ref2 string
}

type ChangedCmdFlagOptions struct {
	Scope  string
	Output string
}

var flagOpts = &ChangedCmdFlagOptions{
	Scope:  "",
	Output: "",
}

type ChangedCmdOutput struct {
	Functions []string `json:"functions,omitempty"`
	Services  []string `json:"services,omitempty"`
}

func init() {
	o := flagOpts
	f := ChangedCmd.Flags()

	f.StringVar(&o.Scope, "scope", o.Scope, "Kind of target (function|service|all).")
	f.StringVarP(&o.Output, "output", "o", o.Output, "Output format (text|json).")

}

var ChangedCmd = &cobra.Command{
	Use:   "changed {branch|tag|ref|sha} {branch|tag|ref|sha}",
	Short: "List top-level changed folders under functions/services.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		o := flagOpts

		argOpts := &ChangedCmdArgOptions{
			Ref1: args[0],
			Ref2: args[1],
		}

		// read persistent flags defined higher up in your CLI
		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}
		funcSub, err := cmd.Flags().GetString(common.FlagFunctionsSubDir)
		if err != nil {
			log.Fatalf("failed to get functions-subdir flag: %v", err)
		}
		svcSub, err := cmd.Flags().GetString(common.FlagServicesSubDir)
		if err != nil {
			log.Fatalf("failed to get services-subdir flag: %v", err)
		}

		scope := common.ResolveScope(o.Scope)

		result, err := get.GetChanged(argOpts.Ref1, argOpts.Ref2, scope, srcDir, funcSub, svcSub)
		if err != nil {
			log.Fatalf("failed to get changed: %v", err)
		}

		output := &ChangedCmdOutput{
			Functions: result.Functions,
			Services:  result.Services,
		}

		switch o.Output {
		case "json":
			if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
				log.Fatalf("failed to write json: %v", err)
			}

		case "text":
			for _, name := range output.Functions {
				fmt.Println(name)
			}
			for _, name := range output.Services {
				fmt.Println(name)
			}
		default:
			log.Fatalf("invalid --output: %q (expected: text|json)", o.Output)
		}
	},
}
