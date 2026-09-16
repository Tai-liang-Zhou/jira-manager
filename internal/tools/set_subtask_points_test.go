package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jira-mcp-server/internal/tools"
)

// TODO(you): once validatePoints (in set_subtask_points.go) has a real
// rule, add table cases here for the values it should reject (e.g.
// negative numbers), asserting callClient: false for each.
func TestHandlers_SetSubtaskPoints(t *testing.T) {
	tests := []struct {
		name        string
		input       tools.SetSubtaskPointsInput
		clientErr   error
		expectError bool
		callClient  bool
	}{
		{
			name:       "sets points on the given subtask",
			input:      tools.SetSubtaskPointsInput{SubtaskKey: "PROJ-3", Points: 5},
			callClient: true,
		},
		{
			name:        "missing subtask key is rejected before calling the client",
			input:       tools.SetSubtaskPointsInput{Points: 5},
			expectError: true,
		},
		{
			name:        "client error propagates",
			input:       tools.SetSubtaskPointsInput{SubtaskKey: "PROJ-404", Points: 5},
			clientErr:   errors.New("jira: subtask PROJ-404 not found"),
			expectError: true,
			callClient:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			client := &mockJiraClient{
				setSubtaskPointsFunc: func(ctx context.Context, subtaskKey string, points float64) error {
					called = true
					if subtaskKey != tt.input.SubtaskKey || points != tt.input.Points {
						t.Errorf("client called with (%q, %v), want (%q, %v)",
							subtaskKey, points, tt.input.SubtaskKey, tt.input.Points)
					}
					return tt.clientErr
				},
			}
			h := tools.NewHandlers(client)

			_, out, err := h.SetSubtaskPoints(context.Background(), nil, tt.input)

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
			if out.SubtaskKey != tt.input.SubtaskKey || out.Points != tt.input.Points {
				t.Errorf("got %+v, want subtaskKey=%q points=%v", out, tt.input.SubtaskKey, tt.input.Points)
			}
		})
	}
}
