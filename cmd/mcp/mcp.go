package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/selimacerbas/flow/internal/common"
	"github.com/selimacerbas/flow/tools"
)

type McpCmdOptions struct {
	HTTPAddress string
}

var defaults = &McpCmdOptions{
	HTTPAddress: "",
}

func init() {
	d := defaults
	f := McpCmd.Flags()

	f.StringVar(&d.HTTPAddress, "http-address", d.HTTPAddress, "Streamable HTTP address")

}

var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run flow as a mcp server",
	Run: func(cmd *cobra.Command, args []string) {
		d := defaults

		srcDir, err := cmd.Flags().GetString(common.FlagSrcDir)
		if err != nil {
			slog.Error("failed to get src-dir flag", "error", err)
			return
		}
		funcSub, err := cmd.Flags().GetString(common.FlagFunctionsSubDir)
		if err != nil {
			slog.Error("failed to get functions-subdir flag", "error", err)
			return
		}
		svcSub, err := cmd.Flags().GetString(common.FlagServicesSubDir)
		if err != nil {
			slog.Error("failed to get services-subdir flag", "error", err)
			return
		}

		opts := &mcp.ServerOptions{
			Instructions: "Use this server!",
		}
		server := mcp.NewServer(&mcp.Implementation{Name: "flow"}, opts)
		server.AddPrompt(&mcp.Prompt{Name: "greet"}, prompt)
		server.AddResource(&mcp.Resource{
			Name:     "info",
			MIMEType: "text/plain",
			URI:      "embedded/info",
		}, embeddedResource)

		mcp.AddTool(server, &mcp.Tool{Name: "changed", Description: "find changed folders"}, tools.ChangedTool(srcDir, funcSub, svcSub))

		if d.HTTPAddress != "" {
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
				return server
			}, nil)
			slog.Info("MCP handler listening", "address", d.HTTPAddress)
			http.ListenAndServe(d.HTTPAddress, handler)
		} else {
			t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
			if err := server.Run(context.Background(), t); err != nil {
				slog.Error("Server failed", "error", err)
			}
		}
	},
}

func prompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Hi promt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Say hi to " + req.Params.Arguments["name"]},
			},
		},
	}, nil
}

var embeddedResources = map[string]string{
	"info": "Flow CLI MCP Server",
}

func embeddedResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	u, err := url.Parse(req.Params.URI)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "embedded" {
		return nil, fmt.Errorf("wrong schema: %q", u.Scheme)
	}

	key := u.Opaque
	text, ok := embeddedResources[key]
	if !ok {
		return nil, fmt.Errorf("no embedded resource names %q", key)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "text/plain",
				Text:     text,
			},
		},
	}, nil
}
