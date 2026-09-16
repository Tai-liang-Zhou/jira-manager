package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListSubtasksInput identifies the parent Task.
type ListSubtasksInput struct {
	ParentKey string `json:"parentKey" jsonschema:"the key of the parent Task, e.g. PROJ-2"`
}

// SubtaskOut is one Sub-task under the queried parent Task.
type SubtaskOut struct {
	Key     string   `json:"key" jsonschema:"the Sub-task's issue key"`
	Summary string   `json:"summary" jsonschema:"the Sub-task's summary"`
	Points  *float64 `json:"points,omitempty" jsonschema:"the Sub-task's point estimate, absent if not yet estimated"`
}

// ListSubtasksOutput lists every Sub-task under the queried parent Task.
type ListSubtasksOutput struct {
	Subtasks []SubtaskOut `json:"subtasks"`
}

// ListSubtasks lists the Sub-tasks under a given parent Task, along with
// their current point estimates.
func (h *Handlers) ListSubtasks(ctx context.Context, _ *mcp.CallToolRequest, in ListSubtasksInput) (*mcp.CallToolResult, ListSubtasksOutput, error) {
	if in.ParentKey == "" {
		return nil, ListSubtasksOutput{}, fmt.Errorf("tools: list_subtasks: parentKey is required")
	}

	subtasks, err := h.client.ListSubtasks(ctx, in.ParentKey)
	if err != nil {
		return nil, ListSubtasksOutput{}, err
	}

	out := ListSubtasksOutput{Subtasks: make([]SubtaskOut, 0, len(subtasks))}
	for _, s := range subtasks {
		out.Subtasks = append(out.Subtasks, SubtaskOut{Key: s.Key, Summary: s.Summary, Points: s.Points})
	}
	return nil, out, nil
}
