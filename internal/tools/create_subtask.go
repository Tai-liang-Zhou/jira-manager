package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateSubtaskInput describes the Sub-task to create.
type CreateSubtaskInput struct {
	ParentKey string `json:"parentKey" jsonschema:"the key of the parent Task, e.g. PROJ-2"`
	Summary   string `json:"summary" jsonschema:"the Sub-task's summary"`
}

// CreateSubtaskOutput identifies the newly created Sub-task.
type CreateSubtaskOutput struct {
	Key string `json:"key" jsonschema:"the new Sub-task's issue key"`
}

// CreateSubtask creates a Sub-task under the given parent Task.
func (h *Handlers) CreateSubtask(ctx context.Context, _ *mcp.CallToolRequest, in CreateSubtaskInput) (*mcp.CallToolResult, CreateSubtaskOutput, error) {
	if in.ParentKey == "" {
		return nil, CreateSubtaskOutput{}, fmt.Errorf("tools: create_subtask: parentKey is required")
	}
	if in.Summary == "" {
		return nil, CreateSubtaskOutput{}, fmt.Errorf("tools: create_subtask: summary is required")
	}

	ref, err := h.client.CreateSubtask(ctx, in.ParentKey, in.Summary)
	if err != nil {
		return nil, CreateSubtaskOutput{}, err
	}
	return nil, CreateSubtaskOutput{Key: ref.Key}, nil
}
