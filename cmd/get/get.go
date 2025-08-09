package get

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/pkg/get"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/internal/utils"
)

type GetOptions struct {
	Ref      string
	Before   string
	After    string
	ThreeDot bool
	Short    bool
	JSON     bool
	Scope    string
}

var getCmdDefaults = &GetOptions{
	Ref:      "HEAD",
	Before:   "",
	After:    "",
	ThreeDot: false,
	Short:    false,
	JSON:     false,
	Scope:    "",
}

type GetSubCmds struct {
	Range        string
	Changed      string
	MergeBase    string
	BeforeCommit string
	AfterCommit  string
}

var getSubCmdArgs = &GetSubCmds{
	Range:        "range",
	Changed:      "changed",
	MergeBase:    "merge-base",
	BeforeCommit: "before-commit",
	AfterCommit:  "after-commit",
}

var GetCmd = &cobra.Command{
	Use:   "get [operation]",
	Short: "Print a single Git SHA (parent or HEAD)",
	ValidArgs: []string{
		getSubCmdArgs.BeforeCommit,
		getSubCmdArgs.AfterCommit,
	},
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		operation := args[0]

		// detect repo root
		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
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

		// normalize dirs
		srcDir = common.ResolveSrcDir(srcDir)
		funcSub = common.ResolveFunctionsDir(funcSub)
		svcSub = common.ResolveFunctionsDir(svcSub)

		var sha string
		switch operation {
		case getCmdArgs.BeforeCommit:
			sha, err = get.GetBeforeCommit(root)
		case getCmdArgs.AfterCommit:
			sha, err = get.GetAfterCommit(root)
		default:
			// this should never happen thanks to ValidArgs/Args, but just in case:
			log.Fatalf("invalid argument %q; must be before-commit or after-commit", operation)
		}
		if err != nil {
			log.Fatalf("git rev-parse failed: %v", err)
		}

		fmt.Println(sha)
	},
}

func init() {

}
