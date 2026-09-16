package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	epicLinkFieldName    = "Epic Link"
	storyPointsFieldName = "Story Points"

	maxRetries     = 2
	initialBackoff = 200 * time.Millisecond
)

// Client is a Jira Server/Data Center REST API v2 client, scoped to a
// single project.
type Client struct {
	baseURL    string
	pat        string
	projectKey string
	httpClient *http.Client

	epicLinkField    string
	storyPointsField string
}

// NewClient creates a Jira client for the given base URL, Personal Access
// Token, and project key. Call ResolveFields before using any method that
// depends on the Epic Link or Story Points custom fields.
//
// Requests honor the standard HTTP_PROXY, HTTPS_PROXY, and NO_PROXY
// environment variables, so Jira traffic can be routed through an
// outbound proxy when required.
func NewClient(baseURL, pat, projectKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		pat:        pat,
		projectKey: projectKey,
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
		},
	}
}

type jiraField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ResolveFields discovers the custom field IDs for "Epic Link" and "Story
// Points" by calling GET /rest/api/2/field and matching on display name.
// epicLinkOverride and storyPointsOverride, when non-empty, are used
// instead of the auto-detected value.
func (c *Client) ResolveFields(ctx context.Context, epicLinkOverride, storyPointsOverride string) error {
	if epicLinkOverride != "" && storyPointsOverride != "" {
		c.epicLinkField = epicLinkOverride
		c.storyPointsField = storyPointsOverride
		return nil
	}

	resp, err := c.request(ctx, http.MethodGet, "/rest/api/2/field", nil)
	if err != nil {
		return fmt.Errorf("jira: resolve fields: %w", err)
	}
	defer resp.Body.Close()

	var fields []jiraField
	if err := json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return fmt.Errorf("jira: decode field list: %w", err)
	}

	byName := make(map[string]string, len(fields))
	for _, f := range fields {
		byName[f.Name] = f.ID
	}

	c.epicLinkField = epicLinkOverride
	if c.epicLinkField == "" {
		id, ok := byName[epicLinkFieldName]
		if !ok {
			return fmt.Errorf("jira: could not find %q custom field; set JIRA_EPIC_LINK_FIELD to override", epicLinkFieldName)
		}
		c.epicLinkField = id
	}

	c.storyPointsField = storyPointsOverride
	if c.storyPointsField == "" {
		id, ok := byName[storyPointsFieldName]
		if !ok {
			return fmt.Errorf("jira: could not find %q custom field; set JIRA_STORY_POINTS_FIELD to override", storyPointsFieldName)
		}
		c.storyPointsField = id
	}

	return nil
}

// request sends an HTTP request to the Jira API, retrying transient
// failures (5xx, 429, network errors) up to maxRetries times with
// exponential backoff. Non-transient failures (4xx) are returned
// immediately without retry. The caller must close the returned response
// body on success.
func (c *Client) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("jira: marshal request body: %w", err)
		}
		bodyBytes = b
	}

	backoff := initialBackoff
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			backoff *= 2
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("jira: build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.pat)
		req.Header.Set("Accept", "application/json")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("jira: %s %s: %w", method, path, err)
			continue
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("jira: %s %s: status %d: %s", method, path, resp.StatusCode, string(data))
			continue
		}

		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("jira: %s %s: status %d: %s", method, path, resp.StatusCode, string(data))
		}

		return resp, nil
	}

	return nil, lastErr
}
