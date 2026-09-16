package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListEpicsInput takes no parameters; the project is fixed by server
// configuration.
type ListEpicsInput struct{}

// EpicOut is one Epic in the configured project.
type EpicOut struct {
	Key     string `json:"key" jsonschema:"the Epic's issue key, e.g. PROJ-1"`
	Summary string `json:"summary" jsonschema:"the Epic's summary"`
	Status  string `json:"status" jsonschema:"the Epic's current workflow status"`
}

// ListEpicsOutput lists every Epic in the configured project.
type ListEpicsOutput struct {
	Epics []EpicOut `json:"epics" jsonschema:"every Epic in the configured project"`
}

// ListEpics lists the Epics in the configured Jira project.
func (h *Handlers) ListEpics(ctx context.Context, _ *mcp.CallToolRequest, _ ListEpicsInput) (*mcp.CallToolResult, ListEpicsOutput, error) {
	epics, err := h.client.ListEpics(ctx)
	if err != nil {
		return nil, ListEpicsOutput{}, err
	}

	out := ListEpicsOutput{Epics: make([]EpicOut, 0, len(epics))}
	for _, e := range epics {
		out.Epics = append(out.Epics, EpicOut{Key: e.Key, Summary: e.Summary, Status: e.Status})
	}
	return nil, out, nil
}
