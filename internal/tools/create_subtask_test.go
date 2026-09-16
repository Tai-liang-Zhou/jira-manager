package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

func TestHandlers_CreateSubtask(t *testing.T) {
	tests := []struct {
		name        string
		input       tools.CreateSubtaskInput
		clientRef   jira.IssueRef
		clientErr   error
		expectKey   string
		expectError bool
		callClient  bool
	}{
		{
			name:       "creates a subtask under the given parent",
			input:      tools.CreateSubtaskInput{ParentKey: "PROJ-2", Summary: "A slice of the work"},
			clientRef:  jira.IssueRef{Key: "PROJ-20"},
			expectKey:  "PROJ-20",
			callClient: true,
		},
		{
			name:        "missing parent key is rejected before calling the client",
			input:       tools.CreateSubtaskInput{Summary: "A slice of the work"},
			expectError: true,
		},
		{
			name:        "missing summary is rejected before calling the client",
			input:       tools.CreateSubtaskInput{ParentKey: "PROJ-2"},
			expectError: true,
		},
		{
			name:        "unknown parent key error from the client propagates",
			input:       tools.CreateSubtaskInput{ParentKey: "PROJ-404", Summary: "A slice of the work"},
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
				createSubtaskFunc: func(ctx context.Context, parentKey, summary string) (jira.IssueRef, error) {
					called = true
					if parentKey != tt.input.ParentKey || summary != tt.input.Summary {
						t.Errorf("client called with (%q, %q), want (%q, %q)",
							parentKey, summary, tt.input.ParentKey, tt.input.Summary)
					}
					return tt.clientRef, tt.clientErr
				},
			}
			h := tools.NewHandlers(client)

			_, out, err := h.CreateSubtask(context.Background(), nil, tt.input)

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
			if out.Key != tt.expectKey {
				t.Errorf("got key %q, want %q", out.Key, tt.expectKey)
			}
		})
	}
}
