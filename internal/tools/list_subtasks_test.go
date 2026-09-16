package tools_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

func TestHandlers_ListSubtasks(t *testing.T) {
	points := 5.0

	tests := []struct {
		name        string
		input       tools.ListSubtasksInput
		subtasks    []jira.Subtask
		clientErr   error
		expected    []tools.SubtaskOut
		expectError bool
		callClient  bool
	}{
		{
			name:  "returns subtasks with and without a point estimate",
			input: tools.ListSubtasksInput{ParentKey: "PROJ-2"},
			subtasks: []jira.Subtask{
				{Key: "PROJ-20", Summary: "Estimated", Points: &points},
				{Key: "PROJ-21", Summary: "Not yet estimated", Points: nil},
			},
			expected: []tools.SubtaskOut{
				{Key: "PROJ-20", Summary: "Estimated", Points: &points},
				{Key: "PROJ-21", Summary: "Not yet estimated", Points: nil},
			},
			callClient: true,
		},
		{
			name:       "parent task with no subtasks returns an empty list",
			input:      tools.ListSubtasksInput{ParentKey: "PROJ-2"},
			subtasks:   []jira.Subtask{},
			expected:   []tools.SubtaskOut{},
			callClient: true,
		},
		{
			name:        "missing parent key is rejected before calling the client",
			input:       tools.ListSubtasksInput{},
			expectError: true,
		},
		{
			name:        "client error propagates",
			input:       tools.ListSubtasksInput{ParentKey: "PROJ-404"},
			clientErr:   errors.New("jira: parent PROJ-404 not found"),
			expectError: true,
			callClient:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			client := &mockJiraClient{
				listSubtasksFunc: func(ctx context.Context, parentKey string) ([]jira.Subtask, error) {
					called = true
					if parentKey != tt.input.ParentKey {
						t.Errorf("client called with parentKey %q, want %q", parentKey, tt.input.ParentKey)
					}
					return tt.subtasks, tt.clientErr
				},
			}
			h := tools.NewHandlers(client)

			_, out, err := h.ListSubtasks(context.Background(), nil, tt.input)

			if called != tt.callClient {
				t.Errorf("client called = %v, want %v", called, tt.callClient)
			}
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(out.Subtasks, tt.expected) {
				t.Errorf("got subtasks %+v, want %+v", out.Subtasks, tt.expected)
			}
		})
	}
}
