package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

type McpCmdOptions struct {
	HTTP string
}

var defaults = &McpCmdOptions{
	HTTP: "",
}

func init() {
	d := defaults
	f := McpCmd.Flags()

	f.StringVar(&d.HTTP, "http", d.HTTP, "Streamable HTTP address")

}

var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run flow as a mcp server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		d := defaults

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

		mcp.AddTool(server, &mcp.Tool{Name: "changed", Description: "find changed files"}, changedTool)

		if d.HTTP != "" {
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
				return server
			}, nil)
			slog.Info("MCP handler listening at %s", d.HTTP)
			http.ListenAndServe(d.HTTP, handler)
		} else {
			t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
			if err := server.Run(context.Background(), t); err != nil {
				slog.Error("Server failed %v", err)
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

type args struct {
	Name string `json:"name" jsonschema:"the name to say hi to"`
}

// contentTool is a tool that returns unstructured content.
//
// Since its output type is 'any', no output schema is created.
func changedTool(ctx context.Context, req *mcp.CallToolRequest, args args) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hi " + args.Name},
		},
	}, nil, nil
}
