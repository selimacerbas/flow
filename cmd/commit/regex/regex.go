
package regex

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/internal/common"
)

var RegexCmd = &cobra.Command{
	Use:   "regex",
	Short: "Show the commit message regex, an example, and Conventional Commits types",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Println("Commit regex:")
		fmt.Printf("  %s\n\n", common.Pattern)

		fmt.Println("Example:")
		fmt.Printf("  %s\n\n", common.Example)

		// Cheat sheet for Conventional Commits (only the types allowed by your regex)
		fmt.Println("Conventional Commits (type → when to use):")

		tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		for _, r := range conventionalTypes {
			fmt.Fprintf(tw, "  %s\t%s\n", r.Type, r.Desc)
		}
		_ = tw.Flush()

		fmt.Println("\nNotes:")
		fmt.Println("  • Format: <type>(<scope>): <description> (#<issue>)")
		fmt.Println("  • Add '!' after type or scope for breaking changes, e.g., refactor(api)!: ...")
		fmt.Println("  • Keep scope lowercase/kebab-case, e.g., ui, api, core, http-client")
	},
}

type typeRow struct {
	Type string
	Desc string
}

var conventionalTypes = []typeRow{
	{"feat",    "Add a new feature"},
	{"fix",     "Bug fix"},
	{"chore",   "Maintenance that doesn’t touch app behavior (e.g., tooling)"},
	{"docs",    "Documentation-only changes"},
	{"style",   "Code style/formatting; no logic changes"},
	{"refactor","Code changes that neither fix a bug nor add a feature"},
	{"test",    "Add or update tests"},
	{"build",   "Build system or external dependencies (bundler, deps, etc.)"},
	{"ci",      "CI configuration and scripts"},
	{"perf",    "Performance improvements"},
}

