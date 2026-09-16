package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateTaskInput describes the Task to create.
type CreateTaskInput struct {
	EpicKey     string `json:"epicKey" jsonschema:"the key of the Epic this Task belongs to, e.g. PROJ-1"`
	Summary     string `json:"summary" jsonschema:"the Task's summary"`
	Description string `json:"description,omitempty" jsonschema:"an optional longer description of the Task"`
}

// CreateTaskOutput identifies the newly created Task.
type CreateTaskOutput struct {
	Key string `json:"key" jsonschema:"the new Task's issue key"`
}

// CreateTask creates a Task under the given Epic.
func (h *Handlers) CreateTask(ctx context.Context, _ *mcp.CallToolRequest, in CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
	if in.EpicKey == "" {
		return nil, CreateTaskOutput{}, fmt.Errorf("tools: create_task: epicKey is required")
	}
	if in.Summary == "" {
		return nil, CreateTaskOutput{}, fmt.Errorf("tools: create_task: summary is required")
	}

	ref, err := h.client.CreateTask(ctx, in.EpicKey, in.Summary, in.Description)
	if err != nil {
		return nil, CreateTaskOutput{}, err
	}
	return nil, CreateTaskOutput{Key: ref.Key}, nil
}
