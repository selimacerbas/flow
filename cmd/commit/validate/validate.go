package validate

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/pkg/commit"

	"github.com/selimacerbas/flow/internal/common"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate \"<message>\"",
	Short: "Validate a commit message string against the required format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		msg := args[0]
		if err := commit.ValidateMessage(msg); err != nil {
			fmt.Println("Commit message is invalid.")
			fmt.Printf("Expected format: %s\n", common.Example)
			log.Fatal(err) // exit 1
		}
		fmt.Println("Commit message is valid.")
	},
}
