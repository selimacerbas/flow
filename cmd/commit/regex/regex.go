package regex

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/internal/common"
)

var RegexCmd = &cobra.Command{
	Use:   "regex",
	Short: "Show the commit message regex and an example",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Println("Commit regex:")
		fmt.Printf("  %s\n\n", common.Pattern)
		fmt.Println("Example:")
		fmt.Printf("  %s\n", common.Example)
	},
}
