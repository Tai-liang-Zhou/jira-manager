# Jira MCP Server

An [MCP](https://modelcontextprotocol.io) server, written in Go, that exposes a Jira Server/Data Center project's **Epic → Task → Sub-task** workflow as tools an AI assistant (e.g. Claude) can call directly — list epics, create tasks and sub-tasks, and set story-point estimates, without leaving the conversation.

It runs as a standalone HTTP service (not stdio), so it can be hosted centrally and reused across sessions/clients.

## Features

- **`list_epics`** — list every Epic in the configured Jira project
- **`create_task`** — create a Task under a given Epic
- **`create_subtask`** — create a Sub-task under a given Task
- **`set_subtask_points`** — set the story-point estimate on a Sub-task
- **`list_subtasks`** — list the Sub-tasks under a Task, with their point estimates
- Auto-detects the "Epic Link" and "Story Points" custom field IDs at startup (with manual override if detection fails)
- Bearer-token authentication on the HTTP endpoint, independent of the Jira Personal Access Token
- Automatic retry with backoff for transient Jira API failures (5xx, 429); non-transient errors (4xx) fail immediately
- Outbound requests to Jira honor the standard `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` environment variables

See [SPEC.md](SPEC.md) for the full problem statement and user stories behind this design.

## Requirements

- Go 1.26+
- A Jira Server/Data Center instance and a [Personal Access Token](https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html) for it

## Configuration

Copy `.env.example` to `.env` and fill in the values, or export the equivalent environment variables:

| Variable | Required | Description |
|---|---|---|
| `JIRA_BASE_URL` | yes | Base URL of your Jira Server/Data Center instance |
| `JIRA_PAT` | yes | Jira Personal Access Token |
| `JIRA_PROJECT_KEY` | yes | The single project this server is scoped to |
| `MCP_AUTH_TOKEN` | yes | Bearer token clients must present to call this server |
| `PORT` | no | HTTP port to listen on (default `8080`) |
| `JIRA_EPIC_LINK_FIELD` | no | Override for the auto-detected "Epic Link" custom field ID |
| `JIRA_STORY_POINTS_FIELD` | no | Override for the auto-detected "Story Points" custom field ID |
| `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` | no | Standard Go proxy environment variables, for routing Jira requests through an outbound proxy |

## Build & run

```sh
make build   # build the binary into bin/jira-mcp-server
make run     # build and run
make test    # run all tests with race detection
make fmt     # gofmt
make vet     # go vet
```

Or directly:

```sh
go run ./cmd/server
```

The binary reads configuration directly from the process environment — it does **not** load `.env` automatically. Export the variables first, e.g. by sourcing your `.env` file:

```sh
set -a; source .env; set +a
./bin/jira-mcp-server
```

or pass them inline for a one-off run:

```sh
JIRA_BASE_URL=https://jira.example.com \
JIRA_PAT=your_pat \
JIRA_PROJECT_KEY=PROJ \
MCP_AUTH_TOKEN=your_token \
./bin/jira-mcp-server
```

`make run` builds and runs the binary the same way, so the environment must already be exported before invoking it too.

On startup, the server resolves the Jira custom field IDs, registers its MCP tools, and listens on `:$PORT`. Every request must include an `Authorization: Bearer <MCP_AUTH_TOKEN>` header.

## Client configuration

The server speaks MCP over streamable HTTP at the root path (e.g. `http://localhost:8080`), so any MCP client that supports a remote/HTTP transport can connect.

### opencode

Add it to `opencode.json` (project root) or `~/.config/opencode/opencode.json` (global):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "jira": {
      "type": "remote",
      "url": "http://localhost:8080",
      "enabled": true,
      "headers": {
        "Authorization": "Bearer {env:MCP_AUTH_TOKEN}"
      }
    }
  }
}
```

`{env:MCP_AUTH_TOKEN}` reads the token from the environment at runtime instead of hardcoding it. Point `url` at the server's actual host if it's not running locally.

## Project layout

```
cmd/server/         entrypoint: wires config, Jira client, and MCP tools into an HTTP server
internal/config/    environment-variable configuration loading
internal/jira/      Jira REST API v2 client (auth, retries, field resolution, issue operations)
internal/tools/     MCP tool handlers, one file per tool
```
