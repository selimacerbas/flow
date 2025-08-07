package run

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/internal/goexec"
	"github.com/selimacerbas/flow-cli/internal/utils"

	"github.com/selimacerbas/flow-cli/pkg/golang/service"
)

type RunCmd struct {
	CleanCache    bool
	EnableMod     bool
	EnableVendor  bool
	EnableBuild   bool
	Targets       []string
	CustomCommand string
	GoOS          string
	GoArch        string
	GoPrivate     string
	AuthMethod    string
	GitUsername   string
	GitToken      string
}

var runCmdDefaults = &RunCmd{
	CleanCache:    false,
	EnableMod:     false,
	EnableVendor:  false,
	EnableBuild:   false,
	Targets:       []string{},
	CustomCommand: "",
	GoOS:          "",
	GoArch:        "",
	GoPrivate:     "",
	AuthMethod:    "",
	GitUsername:   "",
	GitToken:      "",
}

var GoRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage Go container images (clean, mod/vendor, local/cloud/docker builds)",
	Run: func(cmd *cobra.Command, args []string) {
		d := runCmdDefaults

		srcDir, err := cmd.Flags().GetString("src-dir")
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}

		servicesSubdir, err := cmd.Flags().GetString("services-subdir")
		if err != nil {
			log.Fatalf("failed to get functions-subdir flag: %v", err)
		}

		projectRoot, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root %v", err)
		}

		srcDir = common.ResolveSrcDir(srcDir)
		servicesSubdir = common.ResolveFunctionsDir(servicesSubdir)
		servicesDirAbsPath := service.FormAbsolutePathToServicesDir(projectRoot, srcDir, servicesSubdir)
		targetDirs, err := service.FormAbsolutePathToServiceTargetDirs(servicesDirAbsPath, d.Targets)
		if err != nil {
			log.Fatalf("failed to form absolute path to function targets %v", err)
		}

		// 3) clean cache
		if d.CleanCache {
			if err := goexec.RunGoClean(targetDirs); err != nil {
				log.Fatalf("failed to run go clean %v", err)
			}
		}

		privateHosts := goexec.ResolveGoPrivate(d.GoPrivate)

		// 1. Set GOPRIVATE
		if err := goexec.SetEnvGOPrivate(privateHosts); err != nil {
			log.Fatalf("failed to set GOPRIVATE: %v", err)
		}

		authMethod := common.ResolveAuthMethod(d.AuthMethod)

		switch authMethod {
		case "ssh":
			err = common.SetGitAuthSSH(privateHosts)
			if err != nil {
				log.Fatalf("failed to set git auth SSH %v", err)
			}

		case "https":
			username := common.ResolveGitUsername(d.GitUsername)
			token := common.ResolveGitToken(d.GitToken)

			if username == "" || token == "" {
				log.Fatalf("--git-username and --git-token is required when configuring private HTTPS hosts. consider passsing via flag, config or ENV")
			}

			err = common.SetGitAuthHTTPS(privateHosts, username, token)
			if err != nil {
				log.Fatalf("failed to git auth HTTPS %v", err)
			}

		case "":
			fmt.Printf("no --auth-method has passed or configured. Meaning there is no private hosts to be authenticated.")

		default:
			log.Fatalf("invalid auth-method: %q (expected 'ssh' or 'https')", authMethod)
		}

		if d.EnableMod {
			err = goexec.RunGoMod(targetDirs)
			if err != nil {
				log.Fatalf("failed to run go mod %v", err)

			}
		}

		if d.EnableVendor {
			err = goexec.RunGoVendor(targetDirs)
			if err != nil {
				log.Fatalf("failed to run go vendor %v", err)

			}
		}

		if d.EnableBuild {
			goOS := goexec.ResolveENVGoOS(d.GoOS)
			goArch := goexec.ResolveENVGoArch(d.GoArch)
			err = goexec.RunGoBuild(targetDirs, goOS, goArch)
			if err != nil {
				log.Fatalf("failed to run go build %v", err)

			}
		}

		if d.CustomCommand != "" {
			if !strings.HasPrefix(d.CustomCommand, "go ") {
				err := fmt.Errorf("only `go` commands are allowed with --command or -c")
				if err != nil {
					os.Exit(1)
				}
			}
			if err := common.RunCustomCommand(targetDirs, d.CustomCommand); err != nil {
				log.Fatalf("Custom command failed: %v", err)
			}
		}

	},
}

func init() {
	d := runCmdDefaults
	f := GoRunCmd.Flags()

	// bind flags to struct fields using defaults
	f.BoolVar(&d.CleanCache, "clean-cache", d.CleanCache, "Clean vendor & build dirs")
	f.BoolVarP(&d.EnableMod, "mod", "m", d.EnableMod, "Run go mod tidy")
	f.BoolVarP(&d.EnableVendor, "vendor", "v", d.EnableVendor, "Run go mod vendor")
	f.BoolVarP(&d.EnableBuild, "build", "b", d.EnableBuild, "Build function binaries")
	f.StringSliceVarP(&d.Targets, "target", "t", d.Targets, "List of function names")
	f.StringVarP(&d.CustomCommand, "command", "c", "", "Custom Go-related shell command(s) to run in each target (e.g. 'go clean . && go mod tidy && go build')")

	f.StringVar(&d.GoOS, "os", d.GoOS, "Target OS (overrides config.go.os)")
	f.StringVar(&d.GoArch, "arch", d.GoArch, "Target ARCH (overrides config.go.arch)")
	f.StringVar(&d.GoPrivate, "private", d.GoPrivate, "Coma separated private module hosts (e.g. github.com,gitlab.com)")

	f.StringVar(&d.AuthMethod, "auth-method", "", "Git authentication method to use (ssh or https)")
	f.StringVar(&d.GitUsername, "git-username", d.GitUsername, "Git username for HTTPS")
	f.StringVar(&d.GitToken, "git-token", d.GitToken, "Git token or app password")

	// bind to viper
	_ = viper.BindPFlag("go.os", f.Lookup("os"))
	_ = viper.BindPFlag("go.arch", f.Lookup("arch"))
	_ = viper.BindPFlag("go.private", f.Lookup("private"))

	_ = viper.BindPFlag("git.auth_method", f.Lookup("auth-method"))
	_ = viper.BindPFlag("git.git-username", f.Lookup("git-username"))
	_ = viper.BindPFlag("git.git-token", f.Lookup("git-token"))
}
