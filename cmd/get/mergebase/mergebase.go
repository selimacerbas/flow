package mergebase

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/internal/utils"
	"github.com/selimacerbas/flow/pkg/get"
)

type MergeBaseCmdOptions struct {
	Short bool
}

var defaults = &MergeBaseCmdOptions{
	Short: false,
}

type MergeBaseFlags struct {
	Short string
}

var MergeBaseCmd = &cobra.Command{
	Use:   "merge-base {branch|tag|ref|SHA} {branch|tag|ref|SHA} ",
	Short: "Print the merge-base SHA between two refs",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults
		ref1 := args[0]
		ref2 := args[1]

		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		base, err := get.GetMergeBase(root, ref1, ref2)
		if err != nil {
			log.Fatalf("git merge-base %s %s failed: %v", ref1, ref2, err)
		}

		if d.Short {
			base = get.Shorten(base, 7)
		}
		fmt.Println(base)
	},
}

func init() {
	d := defaults
	f := MergeBaseCmd.Flags()

	f.BoolVar(&d.Short, "short", d.Short, "Print 7-char SHA")
}
