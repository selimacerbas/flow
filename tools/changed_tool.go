package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/selimacerbas/flow/pkg/get"
)

type ChangedToolArgs struct {
	Ref1  string `json:"ref1" jsonschema:"The first git reference (branch, tag, ref, or sha)"`
	Ref2  string `json:"ref2" jsonschema:"The second git reference (branch, tag, ref, or sha)"`
	Scope string `json:"scope" jsonschema:"Kind to scan (function|service|all)."`
}

type ChangedToolOutput struct {
	Functions []string `json:"functions,omitempty"`
	Services  []string `json:"services,omitempty"`
}

// factory and closure function
func ChangedTool(srcDir, funcSub, svcSub string) func(context.Context, *mcp.CallToolRequest, ChangedToolArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args ChangedToolArgs) (*mcp.CallToolResult, any, error) {
		result, err := get.GetChanged(args.Ref1, args.Ref2, args.Scope, srcDir, funcSub, svcSub)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get changed: %w", err)
		}

		output := ChangedToolOutput{
			Functions: result.Functions,
			Services:  result.Services,
		}

		return nil, output, nil
	}
}
