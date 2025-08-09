package golang

import (
	"runtime"

	"github.com/selimacerbas/flow-cli/internal/utils"
)

// expects coma separated string
func ResolveGoPrivate(flagGoPrivate string) string {
	return utils.ResolveStringValue(flagGoPrivate, "go.private", "FLOW_GOPRIVATE", "GOPRIVATE")
}

func ResolveENVGoOS(flagGoOS string) string {
	val := utils.ResolveStringValue(flagGoOS, "go.os", "FLOW_GOOS", "GOOS")
	if val != "" {
		return val
	}
	return runtime.GOOS
}

func ResolveENVGoArch(flagGoArch string) string {
	val := utils.ResolveStringValue(flagGoArch, "go.arch", "FLOW_GOARCH", "GOARCH")
	if val != "" {
		return val
	}
	return runtime.GOARCH
}
