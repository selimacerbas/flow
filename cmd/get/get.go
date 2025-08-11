package get

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/cmd/get/changed"
	"github.com/selimacerbas/flow/cmd/get/commitsha"
	"github.com/selimacerbas/flow/cmd/get/mergebase"
)

type GetSubCmds struct {
	Changed   string
	CommitSHA string
	MergeBase string
}

var args = &GetSubCmds{
	Changed:   "changed",
	CommitSHA: "commit-sha",
	MergeBase: "merge-base",
}

var GetCmd = &cobra.Command{
	Use:   "get [operation]",
	Short: "Provided useful commands",
	ValidArgs: []string{
		args.Changed,
		args.CommitSHA,
		args.MergeBase,
	},
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
}

func init() {
	GetCmd.AddCommand(changed.ChangedCmd)
	GetCmd.AddCommand(commitsha.CommitSHACmd)
	GetCmd.AddCommand(mergebase.MergeBaseCmd)

}
