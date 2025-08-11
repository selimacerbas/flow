package commit

import (
	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/cmd/commit/hook"
	"github.com/selimacerbas/flow-cli/cmd/commit/regex"
	"github.com/selimacerbas/flow-cli/cmd/commit/set"
	"github.com/selimacerbas/flow-cli/cmd/commit/validate"
)

var CommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Conventional commit enforcement (regex, set hook, validate, hook)",
	Long:  "Enforce: <type>(<scope>): <description> (#<issue-number>)",
}

func init() {
	CommitCmd.AddCommand(regex.RegexCmd)
	CommitCmd.AddCommand(set.SetCmd)
	CommitCmd.AddCommand(validate.ValidateCmd)
	CommitCmd.AddCommand(hook.HookCmd) // hidden, used by the git hook
}
