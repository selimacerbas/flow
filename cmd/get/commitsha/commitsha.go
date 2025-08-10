package commitsha

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

var CommitSHACmd = &cobra.Command{
	Use:   "commit-sha",
	Short: "Print the SHA of a ref (HEAD by default)",
	Run: func(cmd *cobra.Command, _ []string) {
		d := defaults

		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		sha, err := get.GetCommitSHA(root, d.Ref)
		if err != nil {
			log.Fatalf("git rev-parse %s failed: %v", d.Ref, err)
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

	f.StringVar(&d.Ref, "ref", d.Ref, "Ref to resolve (e.g. HEAD, main, origin/main, a tag, or a SHA)")
	f.BoolVar(&d.Short, "short", d.Short, "Print 7-char SHA")
}
