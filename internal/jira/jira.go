// Package jira provides domain types shared between the Jira REST client
// and the MCP tool handlers that consume it.
package jira

// Epic is a Jira Epic within the configured project.
type Epic struct {
	Key     string
	Summary string
	Status  string
}

// IssueRef identifies a newly created issue.
type IssueRef struct {
	Key string
}

// Subtask is a Jira Sub-task, optionally carrying a point estimate.
type Subtask struct {
	Key     string
	Summary string
	// Points is nil when the sub-task has not been estimated yet.
	Points *float64
}
