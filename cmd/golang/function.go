package golang

import (
	"log"

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
	GoOS              string
	GoArch            string
	GoEnvPrivate      string
	PrivateSSHHosts   []string
	PrivateHTTPSHosts []string
	GitUsername       string
	GitToken          string
}

var functionCmdDefaults = &FunctionCmd{
	SrcDir:            "",
	FunctionsSubdir:   "",
	GoCleanCache:      false,
	GoEnableMod:       false,
	GoEnableVendor:    false,
	GoEnableBuild:     false,
	GoTargets:         []string{},
	GoOS:              "",
	GoArch:            "",
	PrivateSSHHosts:   []string{},
	PrivateHTTPSHosts: []string{},
	GitUsername:       "",
	GitToken:          "",
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

		privateSSHHosts := common.ResolvePrivateSSHHosts(d.PrivateSSHHosts)
		privateHTTPSHosts := common.ResolvePrivateHTTPSHosts(d.PrivateHTTPSHosts)

		// 1. Set GOPRIVATE
		err = goexec.SetENVGOPrivate(privateSSHHosts, privateHTTPSHosts)
		if err != nil {
			log.Fatalf("failed to set GOPRIVATE %v", err)
		}

		// Only when we have SSH hosts
		if len(privateSSHHosts) > 0 {
			err = common.SetGitAuthSSH(privateSSHHosts)
			if err != nil {
				log.Fatalf("failed to set git auth SSH %v", err)
			}
		}

		// Only when we have HTTPS hosts
		if len(privateHTTPSHosts) > 0 {
			username, token, err := common.ResolveGitAuthHTTPSCredentials(d.GitUsername, d.GitToken)
			if err != nil {
				log.Fatalf("failed to resolve git auth HTTPS credentials %v", err)
			}

			err = common.SetGitAuthHTTPS(privateHTTPSHosts, username, token)
			if err != nil {
				log.Fatalf("failed to git auth HTTPS %v", err)
			}
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
			goOS := goexec.SetENVGoOS(d.GoOS)
			goArch := goexec.SetENVGoArch(d.GoArch)
			err = goexec.RunGoBuild(targetDirs, goOS, goArch)
			if err != nil {
				log.Fatalf("failed to run go build %v", err)

			}
		}

	},
}

func init() {
	d := functionCmdDefaults
	f := GoFunctionCmd.Flags()

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

	f.StringVar(&d.GoOS, "os", d.GoOS, "Target OS (overrides config.go.os)")
	f.StringVar(&d.GoArch, "arch", d.GoArch, "Target ARCH (overrides config.go.arch)")

	f.StringSliceVar(&d.PrivateSSHHosts, "private-ssh-hosts", d.PrivateSSHHosts, "Private Git hosts using SSH")
	f.StringSliceVar(&d.PrivateHTTPSHosts, "private-https-hosts", d.PrivateHTTPSHosts, "Private Git hosts using HTTPS. Requires --git-username and --git-token to be set.")

	f.StringVar(&d.GitUsername, "git-username", d.GitUsername, "Git username for HTTPS")
	f.StringVar(&d.GitToken, "git-token", d.GitToken, "Git token or app password")

	// bind to viper
	_ = viper.BindPFlag("go.os", f.Lookup("os"))
	_ = viper.BindPFlag("go.arch", f.Lookup("arch"))

	_ = viper.BindPFlag("git.private_ssh_hosts", f.Lookup("private-ssh-hosts"))
	_ = viper.BindPFlag("git.private_https_hosts", f.Lookup("private-https-hosts"))

	_ = viper.BindPFlag("dirs.src", f.Lookup("src-dir"))
	_ = viper.BindPFlag("dirs.functions_subdir", f.Lookup("functions-subdir"))
	// _ = viper.BindPFlag("dirs.containers_subdir", f.Lookup("containers-subdir"))
}
