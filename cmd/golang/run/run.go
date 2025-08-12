package run

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/internal/golang"
	"github.com/selimacerbas/flow/internal/utils"
)

type RunCmdOptions struct {
	Scope         string
	Targets       []string
	CustomCommand string
	GoOS          string
	GoArch        string
	GoPrivate     string
	AuthMethod    string
	GitUsername   string
	GitToken      string
}

var defaults = &RunCmdOptions{
	Scope:         "",
	Targets:       []string{},
	CustomCommand: "",
	GoOS:          "",
	GoArch:        "",
	GoPrivate:     "",
	AuthMethod:    "",
	GitUsername:   "",
	GitToken:      "",
}

type RunSubCmds struct {
	Clean  string
	Mod    string
	Vendor string
	Build  string
	Custom string
}

var subs = &RunSubCmds{
	Clean:  "clean",
	Mod:    "mod",
	Vendor: "vendor",
	Build:  "build",
	Custom: "custom",
}

func init() {
	d := defaults
	f := RunCmd.Flags()

	// I would want to keep the flags users can pass to any operation
	f.StringVar(&d.Scope, "scope", d.Scope, "Scope to scan (function|service)")
	f.StringSliceVarP(&d.Targets, "targets", "t", d.Targets, "List of function names")
	f.StringVarP(&d.CustomCommand, "command", "c", d.CustomCommand, "Custom Go-related shell command(s) to run in each target (e.g. 'go clean . && go mod tidy && go build')")

	f.StringVar(&d.GoOS, "os", d.GoOS, "Target OS (overrides config.go.os)")
	f.StringVar(&d.GoArch, "arch", d.GoArch, "Target ARCH (overrides config.go.arch)")
	f.StringVar(&d.GoPrivate, "private", d.GoPrivate, "Coma separated private module hosts (e.g. github.com,gitlab.com)")

	f.StringVar(&d.AuthMethod, "auth-method", d.AuthMethod, "Git authentication method to use (ssh or https)")
	f.StringVar(&d.GitUsername, "git-username", d.GitUsername, "Git username for HTTPS")
	f.StringVar(&d.GitToken, "git-token", d.GitToken, "Git token or app password")

	// bind to viper (same as before)
	_ = viper.BindPFlag("go.os", f.Lookup("os"))
	_ = viper.BindPFlag("go.arch", f.Lookup("arch"))
	_ = viper.BindPFlag("go.private", f.Lookup("private"))
	_ = viper.BindPFlag("git.auth_method", f.Lookup("auth-method"))
	_ = viper.BindPFlag("git.username", f.Lookup("git-username"))
	_ = viper.BindPFlag("git.token", f.Lookup("git-token"))
}

var RunCmd = &cobra.Command{
	Use:   "run [operation]",
	Short: "Manage Go functions (clean, mod, vendor, build, custom)",
	ValidArgs: []string{
		subs.Clean,
		subs.Mod,
		subs.Vendor,
		subs.Build,
		subs.Custom,
	},
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults
		operation := args[0]

		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}
		subFuncDir, err := cmd.Flags().GetString(common.FlagFunctionsSubDir)
		if err != nil {
			log.Fatalf("failed to get functions-subdir flag: %v", err)
		}
		subSvcDir, err := cmd.Flags().GetString(common.FlagServicesSubDir)
		if err != nil {
			log.Fatalf("failed to get service-subdir flag: %v", err)
		}

		projectRoot, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root %v", err)
		}

		srcDir = common.ResolveSrcDir(srcDir)
		scope := common.ResolveScope(d.Scope)
		var subDir string

		switch scope {
		case "function":
			subDir = common.ResolveFunctionsDir(subFuncDir)

		case "service":
			subDir = common.ResolveServicesDir(subSvcDir)
		}

		absPath := utils.FormAbsolutePathToDir(projectRoot, srcDir, subDir)
		targetAbsPaths, err := utils.FormAbsolutePathToTargetDirs(absPath, d.Targets)
		if err != nil {
			log.Fatalf("failed to form absolute path to %s targets %v", scope, err)
		}

		// Configure GOPRIVATE + auth (safe even if no private hosts)
		privateHosts := golang.ResolveGoPrivate(d.GoPrivate)
		if err := golang.SetEnvGOPrivate(privateHosts); err != nil {
			log.Fatalf("failed to set GOPRIVATE: %v", err)
		}

		authMethod := common.ResolveAuthMethod(d.AuthMethod)
		switch authMethod {
		case "ssh":
			if err := common.SetGitAuthSSH(privateHosts); err != nil {
				log.Fatalf("failed to set git auth SSH %v", err)
			}
		case "https":
			username := common.ResolveGitUsername(d.GitUsername)
			token := common.ResolveGitToken(d.GitToken)
			if username == "" || token == "" {
				log.Fatalf("--git-username and --git-token is required when configuring private HTTPS hosts. consider passing via flag, config or ENV")
			}
			if err := common.SetGitAuthHTTPS(privateHosts, username, token); err != nil {
				log.Fatalf("failed to git auth HTTPS %v", err)
			}
		case "":
			fmt.Printf("no --auth-method has passed or configured. Meaning there is no private hosts to be authenticated.")
		default:
			log.Fatalf("invalid auth-method: %q (expected 'ssh' or 'https')", authMethod)
		}

		if d.CustomCommand != "" {
			if !strings.HasPrefix(d.CustomCommand, "go ") {
				err := fmt.Errorf("only `go` commands are allowed with --command or -c")
				if err != nil {
					os.Exit(1)
				}
			}
			if err := common.RunCustomCommand(targetAbsPaths, d.CustomCommand); err != nil {
				log.Fatalf("Custom command failed: %v", err)
			}
		}

		switch operation {
		case subs.Clean:
			if err := golang.RunGoClean(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go clean %v", err)
			}

		case subs.Mod:
			if err := golang.RunGoMod(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go mod %v", err)
			}

		case subs.Vendor:
			if err := golang.RunGoVendor(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go vendor %v", err)
			}

		case subs.Build:
			goOS := golang.ResolveENVGoOS(d.GoOS)
			goArch := golang.ResolveENVGoArch(d.GoArch)
			if err := golang.RunGoBuild(targetAbsPaths, goOS, goArch); err != nil {
				log.Fatalf("failed to run go build %v", err)
			}
		case subs.Custom:
			if d.CustomCommand != "" {
				if !strings.HasPrefix(d.CustomCommand, "go ") {
					err := fmt.Errorf("only `go` commands are allowed with --command or -c")
					if err != nil {
						os.Exit(1)
					}
				}
				if err := common.RunCustomCommand(targetAbsPaths, d.CustomCommand); err != nil {
					log.Fatalf("Custom command failed: %v", err)
				}
			} else {
				log.Fatalf("along with custom operations --command or -c flag has to be set %v", err)
			}

		default:
			log.Fatalf("invalid operation %q (expected one of: clean, mod, vendor, build, custom)", operation)
		}
	},
}
