// Command server runs the Jira MCP server over HTTP.
package main

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jira-mcp-server/internal/config"
	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

// Compile-time check that jira.Client satisfies the interface the tool
// handlers depend on. This lives here, not in package jira, so that jira
// doesn't need to import tools (which would create an import cycle, since
// tools already imports jira for its domain types).
var _ tools.JiraClient = (*jira.Client)(nil)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	client := jira.NewClient(cfg.JiraBaseURL, cfg.JiraPAT, cfg.JiraProjectKey)

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.ResolveFields(startupCtx, cfg.EpicLinkFieldOverride, cfg.StoryPointsFieldOverride); err != nil {
		slog.Error("failed to resolve Jira custom fields", "error", err)
		os.Exit(1)
	}

	handlers := tools.NewHandlers(client)

	server := mcp.NewServer(&mcp.Implementation{Name: "jira-mcp-server", Version: "v0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_epics",
		Description: "List every Epic in the configured Jira project",
	}, handlers.ListEpics)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a Task under a given Epic",
	}, handlers.CreateTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_subtask",
		Description: "Create a Sub-task under a given Task",
	}, handlers.CreateSubtask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_subtask_points",
		Description: "Set the point estimate on a Sub-task",
	}, handlers.SetSubtaskPoints)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_subtasks",
		Description: "List the Sub-tasks under a given Task, with their point estimates",
	}, handlers.ListSubtasks)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	authMiddleware := auth.RequireBearerToken(staticTokenVerifier(cfg.MCPAuthToken), nil)

	addr := ":" + cfg.Port
	slog.Info("jira-mcp-server listening", "addr", addr)
	if err := http.ListenAndServe(addr, authMiddleware(mcpHandler)); err != nil {
		slog.Error("http server exited", "error", err)
		os.Exit(1)
	}
}

// staticTokenVerifier returns a TokenVerifier that accepts exactly one
// fixed token, compared in constant time.
func staticTokenVerifier(expected string) auth.TokenVerifier {
	return func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			return nil, auth.ErrInvalidToken
		}
		return &auth.TokenInfo{Scopes: []string{"read", "write"}}, nil
	}
}
