package tools_test

import (
	"context"

	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

// mockJiraClient is a hand-written fake satisfying tools.JiraClient. Each
// method delegates to a function field so individual test cases only need
// to set the one method they exercise; calling an unset method panics,
// which surfaces a test bug immediately rather than silently returning a
// zero value.
type mockJiraClient struct {
	listEpicsFunc        func(ctx context.Context) ([]jira.Epic, error)
	createTaskFunc       func(ctx context.Context, epicKey, summary, description string) (jira.IssueRef, error)
	createSubtaskFunc    func(ctx context.Context, parentKey, summary string) (jira.IssueRef, error)
	setSubtaskPointsFunc func(ctx context.Context, subtaskKey string, points float64) error
	listSubtasksFunc     func(ctx context.Context, parentKey string) ([]jira.Subtask, error)
}

var _ tools.JiraClient = (*mockJiraClient)(nil)

func (m *mockJiraClient) ListEpics(ctx context.Context) ([]jira.Epic, error) {
	return m.listEpicsFunc(ctx)
}

func (m *mockJiraClient) CreateTask(ctx context.Context, epicKey, summary, description string) (jira.IssueRef, error) {
	return m.createTaskFunc(ctx, epicKey, summary, description)
}

func (m *mockJiraClient) CreateSubtask(ctx context.Context, parentKey, summary string) (jira.IssueRef, error) {
	return m.createSubtaskFunc(ctx, parentKey, summary)
}

func (m *mockJiraClient) SetSubtaskPoints(ctx context.Context, subtaskKey string, points float64) error {
	return m.setSubtaskPointsFunc(ctx, subtaskKey, points)
}

func (m *mockJiraClient) ListSubtasks(ctx context.Context, parentKey string) ([]jira.Subtask, error) {
	return m.listSubtasksFunc(ctx, parentKey)
}
