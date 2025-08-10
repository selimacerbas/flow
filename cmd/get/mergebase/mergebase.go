package mergebase

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/utils"
	"github.com/selimacerbas/flow-cli/pkg/get"
)

type Options struct {
	Ref   string
	Short bool
}

var defaults = &Options{
	Ref:   "HEAD",
	Short: false,
}

var MergeBaseCmd = &cobra.Command{
	Use:   "merge-base <branch>",
	Short: "Print the merge-base SHA between --ref (default HEAD) and <branch>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults
		branch := args[0]

		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		base, err := get.MergeBase(root, d.Ref, branch)
		if err != nil {
			log.Fatalf("git merge-base %s %s failed: %v", d.Ref, branch, err)
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

	f.StringVar(&d.Ref, "ref", d.Ref, "Base ref (default HEAD)")
	f.BoolVar(&d.Short, "short", d.Short, "Print 7-char SHA")
}
