package tools_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jira-mcp-server/internal/jira"
	"github.com/jira-mcp-server/internal/tools"
)

func TestHandlers_ListEpics(t *testing.T) {
	tests := []struct {
		name        string
		epics       []jira.Epic
		clientErr   error
		expected    []tools.EpicOut
		expectError bool
	}{
		{
			name: "returns every epic from the client",
			epics: []jira.Epic{
				{Key: "PROJ-1", Summary: "First epic", Status: "In Progress"},
				{Key: "PROJ-2", Summary: "Second epic", Status: "To Do"},
			},
			expected: []tools.EpicOut{
				{Key: "PROJ-1", Summary: "First epic", Status: "In Progress"},
				{Key: "PROJ-2", Summary: "Second epic", Status: "To Do"},
			},
		},
		{
			name:     "empty project returns an empty list, not an error",
			epics:    []jira.Epic{},
			expected: []tools.EpicOut{},
		},
		{
			name:        "client error propagates",
			clientErr:   errors.New("jira: unreachable"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockJiraClient{
				listEpicsFunc: func(ctx context.Context) ([]jira.Epic, error) {
					return tt.epics, tt.clientErr
				},
			}
			h := tools.NewHandlers(client)

			_, out, err := h.ListEpics(context.Background(), nil, tools.ListEpicsInput{})

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(out.Epics, tt.expected) {
				t.Errorf("got epics %+v, want %+v", out.Epics, tt.expected)
			}
		})
	}
}
