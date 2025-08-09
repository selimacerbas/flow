package run

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/internal/golang"
	"github.com/selimacerbas/flow-cli/internal/utils"

	"github.com/selimacerbas/flow-cli/pkg/golang/service"
)

type RunOptions struct {
	Targets       []string
	CustomCommand string
	GoOS          string
	GoArch        string
	GoPrivate     string
	AuthMethod    string
	GitUsername   string
	GitToken      string
}

var runCmdDefaults = &RunOptions{
	Targets:       []string{},
	CustomCommand: "",
	GoOS:          "",
	GoArch:        "",
	GoPrivate:     "",
	AuthMethod:    "",
	GitUsername:   "",
	GitToken:      "",
}

var RunCmd = &cobra.Command{
	Use:   "run [operation]",
	Short: "Manage Go cloud-functions (clean, mod, vendor, build, custom)",
	ValidArgs: []string{
		golang.CmdClean,
		golang.CmdMod,
		golang.CmdVendor,
		golang.CmdBuild,
	},
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		d := runCmdDefaults
		operation := args[0]

		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}
		subDir, err := cmd.Flags().GetString(common.FlagServicesSubDir)
		if err != nil {
			log.Fatalf("failed to get functions-subdir flag: %v", err)
		}

		projectRoot, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root %v", err)
		}

		srcDir = common.ResolveSrcDir(srcDir)
		subDir = common.ResolveFunctionsDir(subDir)
		absPath := service.FormAbsolutePathToServicesDir(projectRoot, srcDir, subDir)
		targetAbsPaths, err := service.FormAbsolutePathToServiceTargetDirs(absPath, d.Targets)
		if err != nil {
			log.Fatalf("failed to form absolute path to function targets %v", err)
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
		case golang.CmdClean:
			if err := golang.RunGoClean(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go clean %v", err)
			}

		case golang.CmdMod:
			if err := golang.RunGoMod(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go mod %v", err)
			}

		case golang.CmdVendor:
			if err := golang.RunGoVendor(targetAbsPaths); err != nil {
				log.Fatalf("failed to run go vendor %v", err)
			}

		case golang.CmdBuild:
			goOS := golang.ResolveENVGoOS(d.GoOS)
			goArch := golang.ResolveENVGoArch(d.GoArch)
			if err := golang.RunGoBuild(targetAbsPaths, goOS, goArch); err != nil {
				log.Fatalf("failed to run go build %v", err)
			}

		default:
			// should be unreachable thanks to ValidArgs/Args
			log.Fatalf("invalid operation %q (expected one of: clean, mod, vendor, build, custom)", operation)
		}
	},
}

func init() {
	d := runCmdDefaults
	f := RunCmd.Flags()

	// keep the flags users can pass to any operation
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
