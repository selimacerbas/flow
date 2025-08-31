package get

import (
	"fmt"
	"path/filepath"

	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/internal/utils"
)

type ChangedOutput struct {
	Functions []string
	Services  []string
}

// GetChanged finds the changed functions and services between two git references.
func GetChanged(ref1, ref2, scope, srcDir, funcSub, svcSub string) (*ChangedOutput, error) {
	// repo root
	root, err := utils.DetectProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to detect project root: %w", err)
	}

	// normalize dirs
	srcDir = common.ResolveSrcDir(srcDir)
	funcSub = common.ResolveFunctionsDir(funcSub)
	svcSub = common.ResolveServicesDir(svcSub)

	funcRelPath := filepath.Join(srcDir, funcSub)
	svcRelPath := filepath.Join(srcDir, svcSub)

	ref1SHA, err := GetCommitSHA(root, ref1)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit SHA for %s: %w", ref1, err)
	}
	if ref1SHA == common.ZeroCommit {
		ref1SHA = common.EmptyTree
	}

	ref2SHA, err := GetCommitSHA(root, ref2)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit SHA for %s: %w", ref2, err)
	}
	if ref2SHA == common.ZeroCommit {
		ref2SHA = common.EmptyTree
	}

	var funcs, svcs []string
	switch scope {
	case "function":
		funcs, err = GetChangedDirs(root, funcRelPath, ref1SHA, ref2SHA)
		if err != nil {
			return nil, fmt.Errorf("failed to detect chaged function dirs: %w", err)
		}

	case "service":
		svcs, err = GetChangedDirs(root, svcRelPath, ref1SHA, ref2SHA)
		if err != nil {
			return nil, fmt.Errorf("failed detect changed service dirs: %w", err)
		}
	case "all":
		// meaning both dir changes
		funcs, err = GetChangedDirs(root, funcRelPath, ref1SHA, ref2SHA)
		if err != nil {
			return nil, fmt.Errorf("failed to detect chaged function dirs: %w", err)
		}
		svcs, err = GetChangedDirs(root, svcRelPath, ref1SHA, ref2SHA)
		if err != nil {
			return nil, fmt.Errorf("failed detect changed service dirs: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid scope: %q (expected: function|service|all)", scope)
	}

	return &ChangedOutput{
		Functions: funcs,
		Services:  svcs,
	}, nil
}
