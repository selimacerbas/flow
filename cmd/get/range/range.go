package grange

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/utils"
	"github.com/selimacerbas/flow-cli/pkg/get"
)

type Options struct {
	Ref      string
	Before   string
	After    string
	ThreeDot bool
	JSON     bool
}

var defaults = &Options{
	Ref:      "HEAD",
	Before:   "",
	After:    "",
	ThreeDot: false,
	JSON:     false,
}

var RangeCmd = &cobra.Command{
	Use:   "range",
	Short: "Print BEFORE..AFTER (or BEFORE...AFTER with --three-dot)",
	Run: func(cmd *cobra.Command, _ []string) {
		d := defaults

		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		before, after := resolveBeforeAfter(root, d)

		rng := get.RangeString(before, after, d.ThreeDot)

		fmt.Println(rng)
	},
}

func init() {
	d := defaults
	f := RangeCmd.Flags()

	f.StringVar(&d.Ref, "ref", d.Ref, "Ref to base comparisons on (default HEAD)")
	f.StringVar(&d.Before, "before", d.Before, "Override BEFORE commit/ref")
	f.StringVar(&d.After, "after", d.After, "Override AFTER commit/ref")
	f.BoolVar(&d.ThreeDot, "three-dot", d.ThreeDot, "Use symmetric diff (BEFORE...AFTER)")
	f.BoolVar(&d.JSON, "json", d.JSON, "Output JSON")
}

func resolveBeforeAfter(root string, d *Options) (string, string) {
	// explicit overrides win
	if strings.TrimSpace(d.Before) != "" && strings.TrimSpace(d.After) != "" {
		return d.Before, d.After
	}

	var (
		before string
		after  string
		err    error
	)

	if strings.TrimSpace(d.Before) != "" {
		before = d.Before
	} else {
		before, err = get.GetParentOrZero(root, d.Ref)
		if err != nil {
			log.Fatalf("resolve before failed: %v", err)
		}
	}
	if strings.TrimSpace(d.After) != "" {
		after = d.After
	} else {
		after, err = get.GetCommitSHA(root, d.Ref)
		if err != nil {
			log.Fatalf("resolve after failed: %v", err)
		}
	}
	return before, after
}

// 1) “three-dot” in your command
// You already wire it via --three-dot to decide between BEFORE..AFTER and BEFORE...AFTER (presumably in get.RangeString).
// A..B (two dots)
// For diff: compares the two tips (git diff A B).
// For log: shows commits in B but not A.
// A...B (three dots)
// For diff: compares merge-base(A,B) → B (i.e., “what B introduced since it forked from A”).
// For log: shows the symmetric difference (in A or B but not both).
// Use --three-dot when you want “changes introduced by AFTER relative to the common base with BEFORE.”
// Example:
//
// git diff main..feature → “current snapshots differ how?”
// git diff main...feature → “what did feature add since it branched from main?”
