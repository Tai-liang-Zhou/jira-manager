package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// searchIssue is the shape of one entry in a /rest/api/2/search response.
// Fields is kept as a raw map because the set of fields requested (and the
// custom field IDs within it) varies by call.
type searchIssue struct {
	Key    string         `json:"key"`
	Fields map[string]any `json:"fields"`
}

type searchResponse struct {
	Issues []searchIssue `json:"issues"`
}

func (c *Client) search(ctx context.Context, jql, fields string) ([]searchIssue, error) {
	v := url.Values{}
	v.Set("jql", jql)
	v.Set("fields", fields)
	v.Set("maxResults", "100")

	resp, err := c.request(ctx, http.MethodGet, "/rest/api/2/search?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("jira: decode search response: %w", err)
	}
	return out.Issues, nil
}

// ListEpics returns every Epic in the configured project.
func (c *Client) ListEpics(ctx context.Context) ([]Epic, error) {
	jql := fmt.Sprintf("project = %q AND issuetype = Epic", c.projectKey)
	issues, err := c.search(ctx, jql, "summary,status")
	if err != nil {
		return nil, fmt.Errorf("jira: list epics: %w", err)
	}

	epics := make([]Epic, 0, len(issues))
	for _, issue := range issues {
		epics = append(epics, Epic{
			Key:     issue.Key,
			Summary: stringField(issue.Fields, "summary"),
			Status:  statusName(issue.Fields),
		})
	}
	return epics, nil
}

// ListSubtasks returns every Sub-task under the given parent Task key,
// including each one's current point estimate (nil if unestimated).
func (c *Client) ListSubtasks(ctx context.Context, parentKey string) ([]Subtask, error) {
	if c.storyPointsField == "" {
		return nil, fmt.Errorf("jira: story points field not resolved; call ResolveFields first")
	}

	jql := fmt.Sprintf("parent = %q", parentKey)
	issues, err := c.search(ctx, jql, "summary,"+c.storyPointsField)
	if err != nil {
		return nil, fmt.Errorf("jira: list subtasks of %s: %w", parentKey, err)
	}

	subtasks := make([]Subtask, 0, len(issues))
	for _, issue := range issues {
		subtasks = append(subtasks, Subtask{
			Key:     issue.Key,
			Summary: stringField(issue.Fields, "summary"),
			Points:  floatField(issue.Fields, c.storyPointsField),
		})
	}
	return subtasks, nil
}

type createIssueResponse struct {
	Key string `json:"key"`
}

// CreateTask creates a Task in the configured project, linked to the given
// Epic. description may be empty.
func (c *Client) CreateTask(ctx context.Context, epicKey, summary, description string) (IssueRef, error) {
	if c.epicLinkField == "" {
		return IssueRef{}, fmt.Errorf("jira: epic link field not resolved; call ResolveFields first")
	}

	fields := map[string]any{
		"project":       map[string]string{"key": c.projectKey},
		"issuetype":     map[string]string{"name": "Task"},
		"summary":       summary,
		c.epicLinkField: epicKey,
	}
	if description != "" {
		fields["description"] = description
	}

	ref, err := c.createIssue(ctx, fields)
	if err != nil {
		return IssueRef{}, fmt.Errorf("jira: create task under epic %s: %w", epicKey, err)
	}
	return ref, nil
}

// CreateSubtask creates a Sub-task under the given parent Task.
func (c *Client) CreateSubtask(ctx context.Context, parentKey, summary string) (IssueRef, error) {
	fields := map[string]any{
		"project":   map[string]string{"key": c.projectKey},
		"issuetype": map[string]string{"name": "Sub-task"},
		"parent":    map[string]string{"key": parentKey},
		"summary":   summary,
	}

	ref, err := c.createIssue(ctx, fields)
	if err != nil {
		return IssueRef{}, fmt.Errorf("jira: create subtask under %s: %w", parentKey, err)
	}
	return ref, nil
}

func (c *Client) createIssue(ctx context.Context, fields map[string]any) (IssueRef, error) {
	resp, err := c.request(ctx, http.MethodPost, "/rest/api/2/issue", map[string]any{"fields": fields})
	if err != nil {
		return IssueRef{}, err
	}
	defer resp.Body.Close()

	var out createIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return IssueRef{}, fmt.Errorf("decode create issue response: %w", err)
	}
	return IssueRef{Key: out.Key}, nil
}

// SetSubtaskPoints sets the Story Points custom field on the given
// Sub-task.
func (c *Client) SetSubtaskPoints(ctx context.Context, subtaskKey string, points float64) error {
	if c.storyPointsField == "" {
		return fmt.Errorf("jira: story points field not resolved; call ResolveFields first")
	}

	body := map[string]any{
		"fields": map[string]any{
			c.storyPointsField: points,
		},
	}
	resp, err := c.request(ctx, http.MethodPut, "/rest/api/2/issue/"+subtaskKey, body)
	if err != nil {
		return fmt.Errorf("jira: set points on %s: %w", subtaskKey, err)
	}
	defer resp.Body.Close()
	return nil
}

func stringField(fields map[string]any, key string) string {
	s, _ := fields[key].(string)
	return s
}

func statusName(fields map[string]any) string {
	status, ok := fields["status"].(map[string]any)
	if !ok {
		return ""
	}
	name, _ := status["name"].(string)
	return name
}

func floatField(fields map[string]any, key string) *float64 {
	v, ok := fields[key]
	if !ok || v == nil {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	return &f
}
