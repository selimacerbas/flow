package hook

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/pkg/commit"
)

var HookCmd = &cobra.Command{
	Use:    "hook <path-to-commit-msg-file>",
	Short:  "Internal: validate commit message file (used by git commit-msg hook)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("error: cannot read commit message file %s: %v", path, err)
		}
		msg := string(data)
		if err := commit.ValidateMessage(msg); err != nil {
			fmt.Println("Commit message is invalid.")
			fmt.Printf("Expected format: %s\n\n", common.Example)
			fmt.Println("Please amend your commit message to match the required format.")
			log.Fatal(err) // exit 1 → blocks the commit
		}
		// success: be quiet; exit 0
	},
}

func init() {

}
