package golang

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
//
// 	"github.com/selimacerbas/flow-cli/internal/utils"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// )
//
// var (
// 	goImgCleanCache   bool
// 	goImgEnableMod    bool
// 	goImgEnableVendor bool
// 	goImgLocal        bool
// 	goImgCloud        bool
// 	goImgDocker       bool
// 	goImgFromPrivate  bool
// 	goImgUseSSH       bool
// 	goImgUseHTTPS     bool
// 	goImgTargets      []string
//
// 	goImgVersion string
// 	goImgTag     string
// )
//
// var GoImageCmd = &cobra.Command{
// 	Use:   "image",
// 	Short: "Manage Go container images (clean, mod/vendor, local/cloud/docker builds)",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		// resolve dirs
// 		srcDir := viper.GetString("dirs.src")
// 		containersSub := viper.GetString("dirs.containers_subdir")
// 		containersDir := filepath.Join(srcDir, containersSub)
//
// 		// version & private setup
// 		if err := utils.SetENVGoVersion(goImgVersion); err != nil {
// 			fmt.Println("Error switching Go version:", err)
// 			os.Exit(1)
// 		}
// 		if err := utils.SetupPrivateDependencies(
// 			goImgFromPrivate, goImgUseSSH, goImgUseHTTPS,
// 			viper.GetString("image.private_mod"),
// 			os.Getenv("GITHUB_TOKEN"),
// 		); err != nil {
// 			fmt.Println("Error setting up private deps:", err)
// 			os.Exit(1)
// 		}
//
// 		// clean cache (vendor & build)
// 		if goImgCleanCache {
// 			if err := utils.CleanCache(servicesDir, true, goImgTargets); err != nil {
// 				fmt.Println("Clean cache failed:", err)
// 				os.Exit(1)
// 			}
// 		}
//
// 		// mod/vendor
// 		if goImgEnableMod || goImgEnableVendor {
// 			action := "go mod tidy"
// 			if goImgEnableVendor {
// 				action = "go mod vendor"
// 			}
// 			if err := utils.RunModVendor(servicesDir, action, true, goImgTargets); err != nil {
// 				fmt.Println("Mod/vendor failed:", err)
// 				os.Exit(1)
// 			}
// 		}
//
// 		// image builds
// 		switch {
// 		case goImgLocal:
// 			fmt.Println("Building local Docker images (tag=", goImgTag, ")…")
// 			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
// 				name := filepath.Base(dir)
// 				cmd := utils.ExecCommand("docker", "build", "-t", "local/"+name+":"+goImgTag, dir)
// 				if err := cmd.Run(); err != nil {
// 					fmt.Fprintf(os.Stderr, "local build failed for %s: %v\n", name, err)
// 					os.Exit(1)
// 				}
// 			}
//
// 		case goImgCloud:
// 			region := viper.GetString("cloud.region")
// 			proj := viper.GetString("cloud.project")
// 			repo := viper.GetString("cloud.repository")
// 			if region == "" || proj == "" || repo == "" {
// 				fmt.Fprintln(os.Stderr, "cloud.region, cloud.project and cloud.repository must be set")
// 				os.Exit(1)
// 			}
// 			fmt.Println("Triggering Cloud Build…")
// 			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
// 				name := filepath.Base(dir)
// 				substs := fmt.Sprintf("_SERVICE=%s,_REGION=%s,_PROJECT=%s,_REPOSITORY=%s,_TAG=%s",
// 					name, region, proj, repo, goImgTag,
// 				)
// 				cmd := utils.ExecCommand("gcloud", "builds", "submit", dir,
// 					"--config="+filepath.Join(dir, "cloudbuild.yaml"),
// 					"--substitutions="+substs,
// 					"--async",
// 				)
// 				if err := cmd.Run(); err != nil {
// 					fmt.Fprintf(os.Stderr, "cloud build submit failed for %s: %v\n", name, err)
// 					os.Exit(1)
// 				}
// 			}
//
// 		case goImgDocker:
// 			region := viper.GetString("cloud.region")
// 			proj := viper.GetString("cloud.project")
// 			repo := viper.GetString("cloud.repository")
// 			if region == "" || proj == "" || repo == "" {
// 				fmt.Fprintln(os.Stderr, "cloud.region, cloud.project and cloud.repository must be set")
// 				os.Exit(1)
// 			}
// 			fmt.Println("Building & pushing Docker images…")
// 			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
// 				name := filepath.Base(dir)
// 				imageTag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s",
// 					region, proj, repo, name, goImgTag,
// 				)
// 				c := utils.ExecCommand("docker", "build", "-t", imageTag, dir)
// 				if err := c.Run(); err != nil {
// 					fmt.Fprintf(os.Stderr, "docker build failed: %v\n", err)
// 					os.Exit(1)
// 				}
// 				c = utils.ExecCommand("docker", "push", imageTag)
// 				if err := c.Run(); err != nil {
// 					fmt.Fprintf(os.Stderr, "docker push failed: %v\n", err)
// 					os.Exit(1)
// 				}
// 			}
// 		default:
// 			fmt.Println("No build strategy selected; use --local-build, --cloud-build, or --docker-build.")
// 		}
// 	},
// }
//
// func init() {
// 	f := GoImageCmd.Flags()
//
// 	// original flags
// 	f.BoolVar(&goImgCleanCache, "clean-cache", false, "Clean vendor & build dirs")
// 	f.BoolVarP(&goImgEnableMod, "mod", "m", false, "Run go mod tidy")
// 	f.BoolVarP(&goImgEnableVendor, "vendor", "v", false, "Run go mod vendor")
// 	f.BoolVar(&goImgLocal, "local-build", false, "Local Docker builds")
// 	f.BoolVar(&goImgCloud, "cloud-build", false, "Trigger Cloud Build")
// 	f.BoolVar(&goImgDocker, "docker-build", false, "Docker build & push")
// 	f.StringSliceVarP(&goImgTargets, "target", "t", nil, "Comma-separated list of services (default=all)")
//
// 	// language-specific & image
// 	f.StringVar(&goImgVersion, "go-version", "", "Go version (overrides config.go.default_version)")
// 	f.StringVar(&goImgTag, "tag", "", "Image tag (overrides config.image.tag)")
//
// 	// auth flags
// 	f.BoolVar(&goImgFromPrivate, "from-private", false, "Enable private module support")
// 	f.BoolVar(&goImgUseSSH, "use-ssh", false, "Use SSH for private git")
// 	f.BoolVar(&goImgUseHTTPS, "use-https", false, "Use HTTPS+GITHUB_TOKEN for private git")
//
// 	// bind into Viper
// 	_ = viper.BindPFlag("go.default_version", f.Lookup("go-version"))
// 	_ = viper.BindPFlag("image.tag", f.Lookup("tag"))
// }
