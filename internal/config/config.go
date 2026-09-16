// Package config loads server configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds all environment-driven configuration for the server.
type Config struct {
	JiraBaseURL    string
	JiraPAT        string
	JiraProjectKey string

	MCPAuthToken string
	Port         string

	// EpicLinkFieldOverride and StoryPointsFieldOverride, when non-empty,
	// take precedence over auto-detected custom field IDs.
	EpicLinkFieldOverride    string
	StoryPointsFieldOverride string
}

// Load reads configuration from environment variables, returning an error
// if any required variable is missing.
func Load() (Config, error) {
	cfg := Config{
		JiraBaseURL:              os.Getenv("JIRA_BASE_URL"),
		JiraPAT:                  os.Getenv("JIRA_PAT"),
		JiraProjectKey:           os.Getenv("JIRA_PROJECT_KEY"),
		MCPAuthToken:             os.Getenv("MCP_AUTH_TOKEN"),
		Port:                     os.Getenv("PORT"),
		EpicLinkFieldOverride:    os.Getenv("JIRA_EPIC_LINK_FIELD"),
		StoryPointsFieldOverride: os.Getenv("JIRA_STORY_POINTS_FIELD"),
	}

	required := map[string]string{
		"JIRA_BASE_URL":    cfg.JiraBaseURL,
		"JIRA_PAT":         cfg.JiraPAT,
		"JIRA_PROJECT_KEY": cfg.JiraProjectKey,
		"MCP_AUTH_TOKEN":   cfg.MCPAuthToken,
	}
	for name, value := range required {
		if value == "" {
			return Config{}, fmt.Errorf("config: missing required environment variable %s", name)
		}
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
