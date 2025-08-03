package golang

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

	"github.com/selimacerbas/flow-cli/pkg/golang/function"
)

type FunctionCmd struct {
	SrcDir            string
	FunctionsSubdir   string
	GoCleanCache      bool
	GoEnableMod       bool
	GoEnableVendor    bool
	GoEnableBuild     bool
	GoTargets         []string
	GoCustomCommand   string
	GoOS              string
	GoArch            string
	GoPrivate         []string
	PrivateSSHHosts   []string
	PrivateHTTPSHosts []string
	AuthMethod        string
	GitUsername       string
	GitToken          string
	DryRun            bool
	Verbose           bool
}

var functionCmdDefaults = &FunctionCmd{
	SrcDir:            "",
	FunctionsSubdir:   "",
	GoCleanCache:      false,
	GoEnableMod:       false,
	GoEnableVendor:    false,
	GoEnableBuild:     false,
	GoTargets:         []string{},
	GoCustomCommand:   "",
	GoOS:              "",
	GoArch:            "",
	GoPrivate:         []string{},
	PrivateSSHHosts:   []string{},
	PrivateHTTPSHosts: []string{},
	AuthMethod:        "",
	GitUsername:       "",
	GitToken:          "",
	DryRun:            false,
	Verbose:           false,
}

var GoFunctionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage Go cloud-functions (mod, vendor, build, clean)",
	Run: func(cmd *cobra.Command, args []string) {
		d := functionCmdDefaults

		projectRoot, err := utils.DetectProjectRoot()
		if err != nil {
			log.Fatalf("failed to detect project root %v", err)
		}

		functionsDir, err := function.ResolveFunctionsDir(projectRoot, d.SrcDir, d.FunctionsSubdir)
		if err != nil {
			log.Fatalf("failed to resolve directories %v", err)
		}

		targetDirs, err := function.ResolveFunctionTargetDirs(functionsDir, d.GoTargets)
		if err != nil {
			log.Fatalf("failed to resolve function targets %v", err)
		}

		// 3) clean cache
		if d.GoCleanCache {
			if err := goexec.RunGoClean(targetDirs); err != nil {
				log.Fatalf("failed to run go clean %v", err)
			}
		}

		privateHosts := goexec.ResolveGoPrivate(d.GoPrivate)

		// 1. Set GOPRIVATE
		if err := goexec.SetEnvGOPrivate(privateHosts); err != nil {
			log.Fatalf("failed to set GOPRIVATE: %v", err)
		}

		authMethod, err := common.ResolveAuthMethod(d.AuthMethod)
		if err != nil {
			log.Fatalf("failed to resolve auth method %v", err)
		}

		switch authMethod {
		case "ssh":
			err = common.SetGitAuthSSH(privateHosts)
			if err != nil {
				log.Fatalf("failed to set git auth SSH %v", err)
			}

		case "https":
			username, token, err := common.ResolveGitAuthHTTPSCredentials(d.GitUsername, d.GitToken)
			if err != nil {
				log.Fatalf("failed to resolve git auth HTTPS credentials %v", err)
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

		if d.GoEnableMod {
			err = goexec.RunGoMod(targetDirs)
			if err != nil {
				log.Fatalf("failed to run go mod %v", err)

			}
		}

		if d.GoEnableVendor {
			err = goexec.RunGoVendor(targetDirs)
			if err != nil {
				log.Fatalf("failed to run go vendor %v", err)

			}
		}

		if d.GoEnableBuild {
			goOS := goexec.ResolveENVGoOS(d.GoOS)
			goArch := goexec.ResolveENVGoArch(d.GoArch)
			err = goexec.RunGoBuild(targetDirs, goOS, goArch)
			if err != nil {
				log.Fatalf("failed to run go build %v", err)

			}
		}

		if d.GoCustomCommand != "" {
			if !strings.HasPrefix(d.GoCustomCommand, "go ") {
				err := fmt.Errorf("only `go` commands are allowed with --command or -c")
				if err != nil {
					os.Exit(1)
				}
			}
			if err := common.RunCustomCommand(targetDirs, d.GoCustomCommand, d.DryRun, d.Verbose); err != nil {
				log.Fatalf("Custom command failed: %v", err)
			}
		}

	},
}

func init() {
	d := functionCmdDefaults
	f := GoFunctionCmd.Flags()

	//utils
	f.BoolVar(&d.DryRun, "dry-run", false, "Preview actions without executing commands (no side effects)")
	f.BoolVar(&d.Verbose, "verbose", false, "Enable verbose output for command execution")

	// function dirs
	f.StringVar(&d.SrcDir, "src-dir", d.SrcDir, "Root source directory (default from config: dirs.src)")
	f.StringVar(&d.FunctionsSubdir, "functions-subdir", d.FunctionsSubdir, "Subdirectory for cloud functions (default from config: dirs.functions_subdir)")
	// f.StringVar(&ContainersSubdir, "containers-subdir", d.ContainersSubdir, "Subdirectory for containers (default from config: dirs.containers_subdir)")

	// bind flags to struct fields using defaults
	f.BoolVar(&d.GoCleanCache, "clean-cache", d.GoCleanCache, "Clean vendor & build dirs")
	f.BoolVarP(&d.GoEnableMod, "mod", "m", d.GoEnableMod, "Run go mod tidy")
	f.BoolVarP(&d.GoEnableVendor, "vendor", "v", d.GoEnableVendor, "Run go mod vendor")
	f.BoolVarP(&d.GoEnableBuild, "build", "b", d.GoEnableBuild, "Build function binaries")
	f.StringSliceVarP(&d.GoTargets, "target", "t", d.GoTargets, "List of function names")
	f.StringVarP(&d.GoCustomCommand, "command", "c", "", "Custom Go-related shell command(s) to run in each target (e.g. 'go clean . && go mod tidy && go build')")

	f.StringVar(&d.GoOS, "os", d.GoOS, "Target OS (overrides config.go.os)")
	f.StringVar(&d.GoArch, "arch", d.GoArch, "Target ARCH (overrides config.go.arch)")
	f.StringSliceVar(&d.GoPrivate, "private", d.GoPrivate, "List of private module hosts (e.g. github.com, gitlab.com)")

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

	_ = viper.BindPFlag("dirs.src", f.Lookup("src-dir"))
	_ = viper.BindPFlag("dirs.functions_subdir", f.Lookup("functions-subdir"))
	// _ = viper.BindPFlag("dirs.containers_subdir", f.Lookup("containers-subdir"))
}
