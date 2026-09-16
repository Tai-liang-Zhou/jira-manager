package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SetSubtaskPointsInput identifies the Sub-task and the new point value.
type SetSubtaskPointsInput struct {
	SubtaskKey string  `json:"subtaskKey" jsonschema:"the Sub-task's issue key, e.g. PROJ-3"`
	Points     float64 `json:"points" jsonschema:"the point estimate to set on the Sub-task"`
}

// SetSubtaskPointsOutput confirms the value that was set.
type SetSubtaskPointsOutput struct {
	SubtaskKey string  `json:"subtaskKey"`
	Points     float64 `json:"points"`
}

// validatePoints checks whether a point value is acceptable before it is
// sent to Jira. Point values are free-form per the spec (no fixed scale
// like Fibonacci) — but that still leaves the question of what counts as a
// *valid* number.
//
// TODO(you): implement this. Decide: should 0 be allowed (e.g. for a spike
// that turned out to need no work)? Should negative numbers be rejected
// outright, since Jira will silently accept them as a valid float on the
// Story Points field? Return a non-nil error for whatever you decide is
// invalid — it will surface to the caller as a clear tool error, not a
// Jira API failure.
func validatePoints(points float64) error {
	// TODO(you): fill in the validation rule.
	return nil
}

// SetSubtaskPoints sets the Story Points value on the given Sub-task.
func (h *Handlers) SetSubtaskPoints(ctx context.Context, _ *mcp.CallToolRequest, in SetSubtaskPointsInput) (*mcp.CallToolResult, SetSubtaskPointsOutput, error) {
	if in.SubtaskKey == "" {
		return nil, SetSubtaskPointsOutput{}, fmt.Errorf("tools: set_subtask_points: subtaskKey is required")
	}
	if err := validatePoints(in.Points); err != nil {
		return nil, SetSubtaskPointsOutput{}, fmt.Errorf("tools: set_subtask_points: %w", err)
	}

	if err := h.client.SetSubtaskPoints(ctx, in.SubtaskKey, in.Points); err != nil {
		return nil, SetSubtaskPointsOutput{}, err
	}
	return nil, SetSubtaskPointsOutput{SubtaskKey: in.SubtaskKey, Points: in.Points}, nil
}
