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
	"github.com/selimacerbas/flow-cli/internal/utils"

	"github.com/selimacerbas/flow-cli/pkg/golang/service"
)

type BuildCmdOptions struct {
	ImageTag         string
	ImageRepository  string
	ImageBuildMethod string
	Targets          []string
	CustomCommand    string
	CloudProvider    string
	GCPRegion        string
	GCPProjectId     string
	AWSRegion        string
	AWSAccountId     string
	AZURERegistry    string
}

var defaults = &BuildCmdOptions{
	ImageTag:         "",
	ImageRepository:  "",
	ImageBuildMethod: "",
	Targets:          []string{},
	CustomCommand:    "",
	CloudProvider:    "",
	GCPRegion:        "",
	GCPProjectId:     "",
	AWSRegion:        "",
	AWSAccountId:     "",
	AZURERegistry:    "",
}

func init() {
	d := defaults
	f := BuildCmd.Flags()

	// image settings
	f.StringVar(&d.ImageTag, "image-tag", d.ImageTag, "Image tag")
	f.StringVar(&d.ImageRepository, "image-repository", d.ImageRepository, "Image repository")
	f.StringVar(&d.ImageBuildMethod, "image-build-method", d.ImageBuildMethod, "Image build method (local|docker|cloud-build)")
	// targets & custom command
	f.StringSliceVarP(&d.Targets, "targets", "t", d.Targets, "List of service names")
	f.StringVarP(&d.CustomCommand, "command", "c", "", "Custom Go-related shell command(s) to run in each target (e.g. 'go clean . && go mod tidy && go build')")

	// cloud provider settings
	f.StringVar(&d.CloudProvider, "cloud-provider", d.CloudProvider, "Cloud provider (gcp|aws|azure)")
	f.StringVar(&d.GCPRegion, "gcp-region", d.GCPRegion, "GCP region")
	f.StringVar(&d.GCPProjectId, "gcp-project-id", d.GCPProjectId, "GCP project ID")
	f.StringVar(&d.AWSRegion, "aws-region", d.AWSRegion, "AWS region")
	f.StringVar(&d.AWSAccountId, "aws-account-id", d.AWSAccountId, "AWS account ID")
	f.StringVar(&d.AZURERegistry, "azure-registry", d.AZURERegistry, "Azure registry")

	// bind to viper
	_ = viper.BindPFlag("image.tag", f.Lookup("image-tag"))
	_ = viper.BindPFlag("image.repository", f.Lookup("image-repository"))
	_ = viper.BindPFlag("image.build_method", f.Lookup("image-build-method"))

	_ = viper.BindPFlag("cloud.provider", f.Lookup("cloud-provider"))
	_ = viper.BindPFlag("cloud.gcp.region", f.Lookup("gcp-region"))
	_ = viper.BindPFlag("cloud.gcp.project_id", f.Lookup("gcp-project-id"))
	_ = viper.BindPFlag("cloud.aws.region", f.Lookup("aws-region"))
	_ = viper.BindPFlag("cloud.aws.account_id", f.Lookup("aws-account-id"))
	_ = viper.BindPFlag("cloud.azure.registry", f.Lookup("azure-registry"))

	// required flags
	// _ = GoBuildCmd.MarkFlagRequired("image-build-method")
}

var BuildCmd = &cobra.Command{
	Use:   "build ...",
	Short: "Manage Go container images (clean, mod/vendor, local/cloud/docker builds)",
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults

		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			log.Fatalf("failed to get src-dir flag: %v", err)
		}

		servicesSubdir, err := cmd.Flags().GetString(common.FlagServicesSubDir)
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

		imageTag := common.ResolveImageTag(d.ImageTag)
		imageRepository := common.ResolveImageRepository(d.ImageRepository)
		imageBuildMethod := common.ResolveImageBuildMethod(d.ImageBuildMethod)
		cloudProvider := common.ResolveCloudProvider(d.CloudProvider)
		gcpRegion := common.ResolveGCPRegion(d.GCPRegion)
		gcpProjectId := common.ResolveGCPProjectId(d.GCPProjectId)
		awsRegion := common.ResolveAWSRegion(d.AWSRegion)
		awsAccountId := common.ResolveAWSAccountId(d.AWSAccountId)
		azureRegistry := common.ResolveAzureRegistry(d.AZURERegistry)

		switch {
		case imageBuildMethod == "local":
			fmt.Println("Building local Docker images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				tag := fmt.Sprintf("local/%s:%s", name, imageTag)

				cmd := exec.Command("docker", "build", "-t", tag, dir)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					log.Fatalf("docker build failed for %s: %v", name, err)
				}
			}

		case cloudProvider == "gcp" && imageBuildMethod == "docker":
			if gcpRegion == "" || gcpProjectId == "" {
				log.Fatal("--gcp-region and --gcp-project are required for GCP docker builds")
			}
			fmt.Println("Building & pushing GCP Docker images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				tag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:%s",
					gcpRegion, gcpProjectId, imageRepository, name, imageTag,
				)

				// build
				buildCmd := exec.Command("docker", "build", "-t", tag, dir)
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr
				if err := buildCmd.Run(); err != nil {
					log.Fatalf("docker build failed for %s: %v", name, err)
				}

				// push
				pushCmd := exec.Command("docker", "push", tag)
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr
				if err := pushCmd.Run(); err != nil {
					log.Fatalf("docker push failed for %s: %v", name, err)
				}
			}

		case cloudProvider == "aws" && imageBuildMethod == "docker":
			if awsAccountId == "" || awsRegion == "" {
				log.Fatal("--aws-account and --aws-region is required for AWS docker builds")
			}

			fmt.Println("Building & pushing AWS ECR images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				tag := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
					awsAccountId, awsRegion, imageRepository, imageTag,
				)

				buildCmd := exec.Command("docker", "build", "-t", tag, dir)
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr
				if err := buildCmd.Run(); err != nil {
					log.Fatalf("docker build failed for %s: %v", name, err)
				}

				pushCmd := exec.Command("docker", "push", tag)
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr
				if err := pushCmd.Run(); err != nil {
					log.Fatalf("docker push failed for %s: %v", name, err)
				}
			}

		case cloudProvider == "azure" && imageBuildMethod == "docker":
			if azureRegistry == "" {
				log.Fatal("--azure-registry is required for Azure docker builds")
			}
			fmt.Println("Building & pushing Azure ACR images...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				tag := fmt.Sprintf("%s.azurecr.io/%s:%s", azureRegistry, imageRepository, imageTag)

				buildCmd := exec.Command("docker", "build", "-t", tag, dir)
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr
				if err := buildCmd.Run(); err != nil {
					log.Fatalf("docker build failed for %s: %v", name, err)
				}

				pushCmd := exec.Command("docker", "push", tag)
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr
				if err := pushCmd.Run(); err != nil {
					log.Fatalf("docker push failed for %s: %v", name, err)
				}
			}

		case cloudProvider == "gcp" && imageBuildMethod == "cloud-build":
			if gcpRegion == "" || gcpProjectId == "" {
				log.Fatal("--gcp-region and --gcp-project are required for GCP cloud-build")
			}
			fmt.Println("→ Submitting GCP Cloud Build jobs...")
			for _, dir := range targetDirs {
				name := filepath.Base(dir)
				substs := fmt.Sprintf("_SERVICE=%s,_REGION=%s,_PROJECT=%s,_REPOSITORY=%s,_TAG=%s",
					name, gcpRegion, gcpProjectId, imageRepository, imageTag,
				)
				cmd := exec.Command(
					"gcloud", "builds", "submit", dir,
					"--config="+filepath.Join(dir, "cloudbuild.yaml"),
					"--substitutions="+substs,
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					log.Fatalf("gcloud build submit failed for %s: %v", name, err)
				}
			}

		case imageBuildMethod == "":
			// noting to built

		default:
			log.Fatalf("unsupported combination: provider=%q method=%q", cloudProvider, imageBuildMethod)
		}
	},
}
