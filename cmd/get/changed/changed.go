package changed

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/internal/utils"
	"github.com/selimacerbas/flow-cli/pkg/get"
)

type Options struct {
	Ref    string
	Before string
	After  string
	Scope  string // functions|services|both
	JSON   bool
}

var defaults = &Options{
	Ref:    "HEAD",
	Before: "",
	After:  "",
	Scope:  "functions",
	JSON:   false,
}

var Cmd = &cobra.Command{
	Use:   "changed",
	Short: "List top-level changed folders under functions/services",
	Run: func(cmd *cobra.Command, _ []string) {
		d := defaults

		// repo root
		root, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root: %v", err)
		}

		// read persistent flags defined higher up in your CLI
		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}
		funcSub, err := cmd.Flags().GetString(common.FlagFunctionsSubDir)
		if err != nil {
			log.Fatalf("failed to get functions-subdir flag: %v", err)
		}
		svcSub, err := cmd.Flags().GetString(common.FlagServicesSubDir)
		if err != nil {
			// okay if not set yet
			svcSub = ""
		}

		// normalize dirs
		srcDir = common.ResolveSrcDir(srcDir)
		funcSub = common.ResolveFunctionsDir(funcSub)
		svcSub = common.ResolveServicesDir(svcSub) // reuse if you don't have a separate resolver

		before, after := resolveBeforeAfter(root, d)

		var funcs, svcs []string
		var errF, errS error

		scope := strings.TrimSpace(d.Scope)
		if scope == "" {
			scope = "functions"
		}

		if scope == "functions" || scope == "both" {
			funcs, errF = get.ChangedTopLevelDirs(root, get.JoinRel(srcDir, funcSub), before, after)
			if errF != nil {
				log.Fatalf("detect changed function dirs: %v", errF)
			}
		}
		if (scope == "services" || scope == "both") && strings.TrimSpace(svcSub) != "" {
			svcs, errS = get.ChangedTopLevelDirs(root, get.JoinRel(srcDir, svcSub), before, after)
			if errS != nil {
				log.Fatalf("detect changed service dirs: %v", errS)
			}
		}

		sort.Strings(funcs)
		sort.Strings(svcs)
		for _, name := range funcs {
			fmt.Println(name)
		}
		for _, name := range svcs {
			fmt.Println(name)
		}
	},
}

func init() {
	d := defaults
	f := Cmd.Flags()

	f.StringVar(&d.Ref, "ref", d.Ref, "Ref to base comparisons on (default HEAD)")
	f.StringVar(&d.Before, "before", d.Before, "Override BEFORE commit/ref")
	f.StringVar(&d.After, "after", d.After, "Override AFTER commit/ref")
	f.StringVar(&d.Scope, "scope", d.Scope, "Scope to scan: functions|services|both")
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
