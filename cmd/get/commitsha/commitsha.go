package commitsha

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/utils"
	"github.com/selimacerbas/flow-cli/pkg/get"
)

type CommitSHACmdOptions struct {
	Short bool
}

var defaults = &CommitSHACmdOptions{
	Short: false,
}

var CommitSHACmd = &cobra.Command{
	Use:   "commit-sha {branch|tag|rev expr}",
	Short: "Print the SHA of a ref",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults
		ref := args[0]

		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		sha, err := get.GetCommitSHA(root, ref)
		if err != nil {
			log.Fatalf("git rev-parse %s failed: %v", ref, err)
		}

		if d.Short {
			sha = get.Shorten(sha, 7)
		}
		fmt.Println(sha)
	},
}

func init() {
	d := defaults
	f := CommitSHACmd.Flags()

	f.BoolVar(&d.Short, "short", d.Short, "Print 7-char SHA")
}
