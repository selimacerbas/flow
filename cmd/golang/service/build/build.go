package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/selimacerbas/flow-cli/internal/common"
	"github.com/selimacerbas/flow-cli/internal/goexec"
	"github.com/selimacerbas/flow-cli/internal/utils"

	"github.com/selimacerbas/flow-cli/pkg/golang/service"
)

type BuildCmd struct {
	ImageTag         string
	ImageVersion     string
	ImageBuildMethod string
	Targets          []string
	CustomCommand    string
	CloudProvider    string
	CloudRegion      string
	CloudProjectId   string
}

var buildCmdDefaults = &BuildCmd{
	ImageTag:         "",
	ImageVersion:     "",
	ImageBuildMethod: "",
	CloudProvider:    "",
	CloudRegion:      "",
	CloudProjectId:   "",
	Targets:          []string{},
	CustomCommand:    "",
}
var (
	goImgCleanCache   bool
	goImgEnableMod    bool
	goImgEnableVendor bool
	goImgLocal        bool
	goImgCloud        bool
	goImgDocker       bool
	goImgFromPrivate  bool
	goImgUseSSH       bool
	goImgUseHTTPS     bool
	goImgTargets      []string

	goImgVersion string
	goImgTag     string
)

var GoServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage Go container images (clean, mod/vendor, local/cloud/docker builds)",
	Run: func(cmd *cobra.Command, args []string) {
		d := buildCmdDefaults

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

		servicesDir, err := service.ResolveServicesDir(projectRoot, srcDir, servicesSubdir)
		if err != nil {
			log.Fatalf("failed to resolve directories %v", err)
		}

		targetDirs, err := service.ResolveServiceTargetDirs(servicesDir, d.Targets)
		if err != nil {
			log.Fatalf("failed to resolve service targets %v", err)
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

		imageBuildMethod := common.ResolveImageBuildMethod(d.ImageBuildMethod)
		imageTag := common.ResolveImageTag(d.ImageTag)

		switch imageBuildMethod {
		case "local":
			fmt.Println("Building local Docker images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				tag := fmt.Sprintf("local/%s:%s", name, imageTag)

				cmd := exec.Command("docker", "build", "-t", tag, dir)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Local build failed for %s: %v\n", name, err)
					os.Exit(1)
				}
			}

		case "docker":
			fmt.Println("Building & pushing Docker images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				// GCP Artifact Registry Image Tag/URL looks like this, [LOCATION]-docker.pkg.dev/[PROJECT-ID]/[REPOSITORY]/[IMAGE]:[TAG]
				tag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s", region, project, repo, name, imageTag)

				buildCmd := exec.Command("docker", "build", "-t", tag, dir)
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr

				if err := buildCmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Build failed for %s: %v\n", name, err)
					os.Exit(1)
				}

				pushCmd := exec.Command("docker", "push", tag)
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr

				if err := pushCmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Push failed for %s: %v\n", name, err)
					os.Exit(1)
				}
			}

		case "cloud-build":
			fmt.Println("→ Submitting Cloud Build jobs...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				substs := fmt.Sprintf("_SERVICE=%s,_REGION=%s,_PROJECT=%s,_REPOSITORY=%s,_TAG=%s",
					name, region, project, repo, imageTag,
				)

				cmd := exec.Command("gcloud", "builds", "submit", dir,
					"--config="+filepath.Join(dir, "cloudbuild.yaml"),
					"--substitutions="+substs,
				)

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "❌ Cloud Build failed for %s: %v\n", name, err)
					os.Exit(1)
				}
			}

		default:
			fmt.Fprintln(os.Stderr, "❌ No valid build method selected. Use --build-method flag.")
			os.Exit(1)
		}
		// image builds
		switch {
		case goImgLocal:
			fmt.Println("Building local Docker images (tag=", goImgTag, ")…")
			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
				name := filepath.Base(dir)
				cmd := utils.ExecCommand("docker", "build", "-t", "local/"+name+":"+goImgTag, dir)
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "local build failed for %s: %v\n", name, err)
					os.Exit(1)
				}
			}

		case goImgCloud:
			region := viper.GetString("cloud.region")
			proj := viper.GetString("cloud.project")
			repo := viper.GetString("cloud.repository")
			if region == "" || proj == "" || repo == "" {
				fmt.Fprintln(os.Stderr, "cloud.region, cloud.project and cloud.repository must be set")
				os.Exit(1)
			}
			fmt.Println("Triggering Cloud Build…")
			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
				name := filepath.Base(dir)
				substs := fmt.Sprintf("_SERVICE=%s,_REGION=%s,_PROJECT=%s,_REPOSITORY=%s,_TAG=%s",
					name, region, proj, repo, goImgTag,
				)
				cmd := utils.ExecCommand("gcloud", "builds", "submit", dir,
					"--config="+filepath.Join(dir, "cloudbuild.yaml"),
					"--substitutions="+substs,
					"--async",
				)
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "cloud build submit failed for %s: %v\n", name, err)
					os.Exit(1)
				}
			}

		case goImgDocker:
			region := viper.GetString("cloud.region")
			proj := viper.GetString("cloud.project")
			repo := viper.GetString("cloud.repository")
			if region == "" || proj == "" || repo == "" {
				fmt.Fprintln(os.Stderr, "cloud.region, cloud.project and cloud.repository must be set")
				os.Exit(1)
			}
			fmt.Println("Building & pushing Docker images…")
			for _, dir := range utils.DirList(servicesDir, goImgTargets) {
				name := filepath.Base(dir)
				imageTag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s",
					region, proj, repo, name, goImgTag,
				)
				c := utils.ExecCommand("docker", "build", "-t", imageTag, dir)
				if err := c.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "docker build failed: %v\n", err)
					os.Exit(1)
				}
				c = utils.ExecCommand("docker", "push", imageTag)
				if err := c.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "docker push failed: %v\n", err)
					os.Exit(1)
				}
			}
		default:
			fmt.Println("No build strategy selected; use --local-build, --cloud-build, or --docker-build.")
		}
	},
}

func init() {
	d := serviceCmdDefaults
	f := GoServiceCmd.Flags()

	f.StringVar(&d.ImageBuildMethod, "image-build-method", d.ImageBuildMethod, "Image build method (one of: local, docker, cloud-build)")
	// original flags
	f.BoolVar(&goImgCleanCache, "clean-cache", false, "Clean vendor & build dirs")
	f.BoolVarP(&goImgEnableMod, "mod", "m", false, "Run go mod tidy")
	f.BoolVarP(&goImgEnableVendor, "vendor", "v", false, "Run go mod vendor")
	f.BoolVar(&goImgLocal, "local-build", false, "Local Docker builds")
	f.BoolVar(&goImgCloud, "cloud-build", false, "Trigger Cloud Build")
	f.BoolVar(&goImgDocker, "docker-build", false, "Docker build & push")
	f.StringSliceVarP(&goImgTargets, "target", "t", nil, "Comma-separated list of services (default=all)")

	// language-specific & image
	f.StringVar(&goImgVersion, "go-version", "", "Go version (overrides config.go.default_version)")
	f.StringVar(&goImgTag, "tag", "", "Image tag (overrides config.image.tag)")

	// auth flags
	f.BoolVar(&goImgFromPrivate, "from-private", false, "Enable private module support")
	f.BoolVar(&goImgUseSSH, "use-ssh", false, "Use SSH for private git")
	f.BoolVar(&goImgUseHTTPS, "use-https", false, "Use HTTPS+GITHUB_TOKEN for private git")

	// bind into Viper
	_ = viper.BindPFlag("go.default_version", f.Lookup("go-version"))
	_ = viper.BindPFlag("image.tag", f.Lookup("tag"))
}
