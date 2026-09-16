// Package tools implements the MCP tool handlers for the Epic -> Task ->
// Sub-task -> estimate workflow, against a JiraClient.
package tools

import (
	"context"

	"github.com/jira-mcp-server/internal/jira"
)

// JiraClient is everything the tool handlers need from Jira. It is defined
// here, at the point of consumption, so tests can supply a mock without
// depending on the concrete HTTP-based implementation in package jira.
type JiraClient interface {
	ListEpics(ctx context.Context) ([]jira.Epic, error)
	CreateTask(ctx context.Context, epicKey, summary, description string) (jira.IssueRef, error)
	CreateSubtask(ctx context.Context, parentKey, summary string) (jira.IssueRef, error)
	SetSubtaskPoints(ctx context.Context, subtaskKey string, points float64) error
	ListSubtasks(ctx context.Context, parentKey string) ([]jira.Subtask, error)
}

// Handlers holds the dependencies shared by every tool handler.
type Handlers struct {
	client JiraClient
}

// NewHandlers returns a Handlers backed by the given JiraClient.
func NewHandlers(client JiraClient) *Handlers {
	return &Handlers{client: client}
}
