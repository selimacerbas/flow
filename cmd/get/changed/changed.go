package changed

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/selimacerbas/flow/pkg/get"

	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/internal/utils"
)

type ChangedCmdOptions struct {
	Scope  string
	Output string
}

var defaults = &ChangedCmdOptions{
	Scope:  "",
	Output: "",
}

var ChangedCmd = &cobra.Command{
	Use:   "changed {branch|tag|ref|sha} {branch|tag|ref|sha}",
	Short: "List top-level changed folders under functions/services.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults
		ref1 := args[0]
		ref2 := args[1]

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
			log.Fatalf("failed to get services-subdir flag: %v", err)
		}

		// normalize dirs
		srcDir = common.ResolveSrcDir(srcDir)
		funcSub = common.ResolveFunctionsDir(funcSub)
		svcSub = common.ResolveServicesDir(svcSub)

		funcRelPath := filepath.Join(srcDir, funcSub)
		svcRelPath := filepath.Join(srcDir, svcSub)

		ref1SHA, err := get.GetCommitSHA(root, ref1)
		if err != nil {
			log.Fatalf("failed to get commit SHA for %s: %v", ref1, err)
		}
		if ref1SHA == common.ZeroCommit {
			ref1SHA = common.EmptyTree
		}

		ref2SHA, err := get.GetCommitSHA(root, ref2)
		if err != nil {
			log.Fatalf("failed to get commit SHA for %s: %v", ref2, err)
		}
		if ref2SHA == common.ZeroCommit {
			ref2SHA = common.EmptyTree
		}

		var funcs, svcs []string
		switch d.Scope {
		case "function":
			funcs, err = get.GetChangedDirs(root, funcRelPath, ref1SHA, ref2SHA)
			if err != nil {
				log.Fatalf("failed to detect chaged function dirs: %v", err)
			}

		case "service":
			svcs, err = get.GetChangedDirs(root, svcRelPath, ref1SHA, ref2SHA)
			if err != nil {
				log.Fatalf(" failed detect changed service dirs: %v", err)
			}
		case "":
			// meaning both dir changes

			funcs, err = get.GetChangedDirs(root, funcRelPath, ref1SHA, ref2SHA)
			if err != nil {
				log.Fatalf("failed to detect chaged function dirs: %v", err)
			}
			svcs, err = get.GetChangedDirs(root, svcRelPath, ref1SHA, ref2SHA)
			if err != nil {
				log.Fatalf(" failed detect changed service dirs: %v", err)
			}
		}

		// inside Run:
		switch d.Output {
		case "json":
			var payload any
			switch d.Scope {
			case "function":
				payload = funcs // -> ["f1","f2"]
			case "service":
				payload = svcs // -> ["s1","s2"]
			default:
				payload = map[string]any{ // -> {"functions":[...],"services":[...]}
					"functions": funcs,
					"services":  svcs,
				}
			}

			// print JSON
			if err := json.NewEncoder(os.Stdout).Encode(payload); err != nil {
				log.Fatalf("failed to write json: %v", err)
			}

		case "text", "":
			for _, name := range funcs {
				fmt.Println(name)
			}
			for _, name := range svcs {
				fmt.Println(name)
			}
		default:
			log.Fatalf("invalid --output: %q (expected: text|json)", d.Output)
		}
	},
}

func init() {
	d := defaults
	f := ChangedCmd.Flags()

	f.StringVar(&d.Scope, "scope", d.Scope, "Scope to scan: function|service")
}
