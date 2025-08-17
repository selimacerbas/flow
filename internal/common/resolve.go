package common

import (
	"github.com/selimacerbas/flow/internal/utils"
)

func ResolveSrcDir(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "dirs.src", "FLOW_SRC_DIR")
}

func ResolveFunctionsDir(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "dirs.functions_subdir", "FLOW_FUNCTIONS_SUBDIR")
}

func ResolveServicesDir(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "dirs.services_subdir", "FLOW_SERVICES_SUBDIR")
}
func ResolveScope(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "scope", "FLOW_SCOPE")
}

func ResolveGitOwner(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.owner", "FLOW_GIT_OWNER")
}

func ResolveGitRepo(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.repo", "FLOW_GIT_REPO")
}

func ResolveGitToken(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.token", "FLOW_GIT_TOKEN")
}

func ResolveGitWorkflow(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.workflow", "FLOW_GIT_WORKFLOW")
}

func ResolveGitBranch(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.branch", "FLOW_GIT_BRANCH")
}

func ResolveAuthMethod(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "git.auth_method", "FLOW_AUTH_METHOD")
}

func ResolveImageTag(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "image.tag", "FLOW_IMAGE_TAG")
}

func ResolveImageRepository(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "image.repository", "FLOW_IMAGE_REPOSITORY")
}

func ResolveImageBuildMethod(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "image.build_method", "FLOW_IMAGE_BUILD_METHOD")
}

func ResolveCloudProvider(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.provider", "FLOW_CLOUD_PROVIDER")
}

func ResolveGCPRegion(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.gcp.region", "FLOW_GCP_REGION")
}

func ResolveGCPProjectId(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.gcp.project_id", "FLOW_GCP_PROJECT_ID")
}

func ResolveAWSAccountId(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.aws.account_id", "FLOW_AWS_ACCOUNT_ID")
}

func ResolveAWSRegion(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.aws.region", "FLOW_AWS_REGION")
}

func ResolveAzureRegistry(flagVal string) string {
	return utils.ResolveStringValue(flagVal, "cloud.azure.registry", "FLOW_AZURE_REGISTRY")
}
