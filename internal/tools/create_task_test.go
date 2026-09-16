package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

func TestHandlers_CreateTask(t *testing.T) {
	tests := []struct {
		name        string
		input       tools.CreateTaskInput
		clientRef   jira.IssueRef
		clientErr   error
		expectKey   string
		expectError bool
		// callClient is false for validation failures the handler must
		// reject before ever reaching the client.
		callClient bool
	}{
		{
			name:       "creates a task under the given epic",
			input:      tools.CreateTaskInput{EpicKey: "PROJ-1", Summary: "Do the thing", Description: "details"},
			clientRef:  jira.IssueRef{Key: "PROJ-10"},
			expectKey:  "PROJ-10",
			callClient: true,
		},
		{
			name:        "missing epic key is rejected before calling the client",
			input:       tools.CreateTaskInput{Summary: "Do the thing"},
			expectError: true,
		},
		{
			name:        "missing summary is rejected before calling the client",
			input:       tools.CreateTaskInput{EpicKey: "PROJ-1"},
			expectError: true,
		},
		{
			name:        "unknown epic key error from the client propagates",
			input:       tools.CreateTaskInput{EpicKey: "PROJ-404", Summary: "Do the thing"},
			clientErr:   errors.New("jira: epic PROJ-404 not found"),
			expectError: true,
			callClient:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			client := &mockJiraClient{
				createTaskFunc: func(ctx context.Context, epicKey, summary, description string) (jira.IssueRef, error) {
					called = true
					if epicKey != tt.input.EpicKey || summary != tt.input.Summary || description != tt.input.Description {
						t.Errorf("client called with (%q, %q, %q), want (%q, %q, %q)",
							epicKey, summary, description, tt.input.EpicKey, tt.input.Summary, tt.input.Description)
					}
					return tt.clientRef, tt.clientErr
				},
			}
			h := tools.NewHandlers(client)

			_, out, err := h.CreateTask(context.Background(), nil, tt.input)

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
