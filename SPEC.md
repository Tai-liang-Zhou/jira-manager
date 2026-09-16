# Jira MCP Server — Spec

## Problem Statement

Managing work in Jira Server/Data Center for a team that structures work as **Epic → Task → Sub-task** requires a lot of manual clicking through the Jira UI: browsing to find the right Epic, creating a Task under it, breaking that Task into Sub-tasks, and then estimating each Sub-task in story points. There's no way to drive this workflow from an AI assistant (e.g. Claude) — every step has to happen inside Jira itself, which breaks flow when planning is happening in conversation with an assistant.

## Solution

Build an MCP (Model Context Protocol) server, written in Go using the official `github.com/modelcontextprotocol/go-sdk`, that exposes the Epic → Task → Sub-task → estimate workflow as MCP tools. The server talks to a single Jira Server/Data Center instance via its REST API v2 using a Personal Access Token, and is exposed over HTTP (not stdio) so it can run as a standalone service, protected by its own bearer-token authentication layer independent of the Jira PAT.

The server auto-detects the Jira custom fields it depends on (Epic Link, Story Points) at startup rather than requiring them to be hardcoded, since these custom field IDs vary per Jira instance.

## User Stories

1. As a user planning work, I want to list the Epics in my Jira project, so that I know what Epic to attach new Tasks to without leaving my conversation with the assistant.
2. As a user planning work, I want to create a Task under a specific Epic, so that new work is automatically organized under the right initiative.
3. As a user breaking down a Task, I want to create a Sub-task under a specific Task, so that I can decompose work into smaller, estimable units.
4. As a user estimating work, I want to set or update the story point estimate on a Sub-task, so that the team has a shared understanding of effort.
5. As a user reviewing a Task, I want to list all Sub-tasks under it along with their current point estimates, so that I can see estimation progress and totals at a glance.
6. As a user, I want estimates to only be recorded on Sub-tasks (not on Tasks or Epics), so that the estimation granularity matches how my team actually works.
7. As a user, I want to enter any numeric value as a point estimate, so that I'm not forced into a specific estimation scale (e.g. Fibonacci) my team doesn't use.
8. As an operator deploying this server, I want it to run as an HTTP service rather than requiring a local stdio-connected process, so that it can be hosted centrally and reused across sessions/clients.
9. As an operator, I want the HTTP endpoint protected by a bearer token distinct from the Jira PAT, so that anyone who can reach the server on the network can't create/modify Jira issues without authorization.
10. As an operator, I want to authenticate to Jira using a Personal Access Token, so that I don't have to store or manage a Jira account password.
11. As an operator, I want the server to auto-detect the "Epic Link" and "Story Points" custom field IDs on startup, so that I don't have to manually look up and hardcode instance-specific custom field IDs.
12. As an operator, I want to be able to override the auto-detected custom field IDs via environment variables, so that I have an escape hatch if auto-detection fails (e.g. because a field was renamed).
13. As an operator, I want the server scoped to a single, fixed Jira project (set via configuration), so that tool calls stay simple and don't require specifying a project on every call.
14. As a user, I want clear, immediate error messages when I reference an Epic or Task key that doesn't exist, so that I can correct my input right away instead of waiting on a retry loop.
15. As an operator, I want transient Jira API failures (5xx, timeouts, rate limiting) to be retried automatically with backoff, so that a momentary blip in Jira doesn't surface as a hard failure to the user.
16. As an operator, I want non-transient errors (4xx, auth failures, not-found) to fail immediately without retry, so that the system doesn't waste time retrying requests that will never succeed.
17. As a developer maintaining this server, I want the Jira API interactions abstracted behind an interface, so that tool handler logic can be unit tested without hitting a real or fake HTTP server.
18. As a user, I want to create Tasks and Sub-tasks using the standard English Jira issue type names ("Task", "Sub-task"), so that no extra configuration is needed for issue type resolution.

## Implementation Decisions

**Language/runtime & SDK**
- Go, using the official `github.com/modelcontextprotocol/go-sdk` for MCP server/tool registration and the HTTP (streamable HTTP) transport.

**Jira target**
- Jira Server/Data Center only (not Jira Cloud) — REST API v2.
- Authentication to Jira: Personal Access Token, sent as a Bearer token on all Jira API requests.

**Scope/configuration**
- Single, fixed Jira project per server instance. Project key supplied via configuration (environment variable), not as a per-call tool parameter.
- Configuration is environment-variable driven:
  - `JIRA_BASE_URL` — base URL of the Jira instance
  - `JIRA_PAT` — Personal Access Token for Jira
  - `JIRA_PROJECT_KEY` — the single project this server operates on
  - `MCP_AUTH_TOKEN` — bearer token required by clients calling this MCP server's HTTP endpoint
  - `PORT` — HTTP listen port
  - `JIRA_EPIC_LINK_FIELD` (optional) — overrides auto-detected Epic Link custom field ID
  - `JIRA_STORY_POINTS_FIELD` (optional) — overrides auto-detected Story Points custom field ID

**Transport & server-side auth**
- MCP server runs as an HTTP service (streamable HTTP transport), not stdio.
- A middleware layer requires a valid `MCP_AUTH_TOKEN` bearer token on incoming requests before any tool is dispatched. This is independent of, and in addition to, the Jira PAT used for outbound calls to Jira.

**Custom field resolution**
- On startup, the server calls `GET /rest/api/2/field` and matches fields by display name ("Epic Link", "Story Points") to resolve their custom field IDs (e.g. `customfield_10008`).
- If `JIRA_EPIC_LINK_FIELD` / `JIRA_STORY_POINTS_FIELD` are set, they take precedence over auto-detection.
- If a required field cannot be resolved (not found by name, and no override set), the server fails to start with a clear error rather than starting in a broken state.

**Issue types**
- Issue type names are hardcoded as the Jira standard English names: `"Epic"`, `"Task"`, `"Sub-task"`. No auto-detection or configuration needed for issue types (confirmed the target project uses standard naming).

**MCP Tools** (five tools; this is the full tool surface for this spec)
1. `list_epics` — no input beyond the fixed project scope; returns Epic key, summary, and status for each Epic in the configured project.
2. `create_task` — input: Epic key, summary, (optional) description; creates a Task issue in the configured project with its Epic Link field set to the given Epic; returns the new Task's key.
3. `create_subtask` — input: parent Task key, summary; creates a Sub-task issue linked to the given parent Task; returns the new Sub-task's key.
4. `set_subtask_points` — input: Sub-task key, points (numeric, free-form — no restriction to a fixed estimation scale); sets the Story Points custom field on that Sub-task.
5. `list_subtasks` — input: parent Task key; returns each Sub-task's key, summary, and current point value (points field absent/null if not yet estimated).

**Estimation rule**
- Only Sub-tasks carry a point estimate. Tasks and Epics have no point-related fields exposed by any tool.
- Point values are unrestricted numbers — no validation against a fixed set (e.g. Fibonacci).

**Jira client abstraction**
- All outbound Jira REST interactions are defined behind a `JiraClient` interface (e.g. `ListEpics`, `CreateTask`, `CreateSubtask`, `SetSubtaskPoints`, `ListSubtasks`, plus a field-resolution method used at startup). MCP tool handlers depend only on this interface, never on the concrete HTTP implementation directly.

**Retry policy**
- Outbound Jira API calls are wrapped with a retry policy: on 5xx responses, network timeouts, or HTTP 429, retry up to 2 additional times with exponential backoff.
- On 4xx responses (400, 401, 404, etc.), fail immediately with no retry, and surface the error clearly to the caller (e.g. "Epic ABC-123 not found").

## Testing Decisions

- **Seam:** the `JiraClient` interface is the sole seam for testing tool handler logic. Tool handlers under test receive a mock implementation of `JiraClient` rather than a real or fake-HTTP-backed client. This keeps the number of test seams in the codebase to one, per project convention of minimizing seams.
- **Test style:** table-driven unit tests (idiomatic Go / prior art: the Go standard library and most Go projects' testing conventions), one table per tool handler. Each table entry specifies: input parameters to the tool, the mock `JiraClient`'s configured behavior/responses for that case, and the expected tool output or expected error.
- **Modules tested:**
  - `create_task` handler — including the case of an unknown Epic key producing an immediate, non-retried error.
  - `create_subtask` handler — including unknown parent Task key.
  - `set_subtask_points` handler — including non-numeric/invalid point input.
  - `list_epics` handler — including an empty-project case (no Epics).
  - `list_subtasks` handler — including a parent Task with zero Sub-tasks, and Sub-tasks with unset point values.
  - Retry policy logic itself — tested in isolation against a fake `JiraClient` (or fake round-tripper) that simulates transient (5xx/timeout/429) vs. non-transient (4xx) failures, asserting retry counts and backoff behavior.
- **What is a good test here:** tests assert on the tool handler's externally observable behavior (its return value / error, and what it called on `JiraClient` with what arguments) — not on internal implementation details of the handler. The concrete HTTP-based `JiraClient` implementation (actual request construction against the real Jira REST API) is a separate concern from this handler-level testing and is not covered by these table-driven tests.

## Out of Scope

- Jira Cloud support (this spec targets Server/Data Center only).
- Multi-project support (project is fixed per server instance via configuration).
- Sprint and board management (adding issues to sprints, board queries, etc.).
- Issue status/workflow transitions, comments, attachments, and free-form JQL queries.
- Restricting point estimates to a fixed scale (e.g. Fibonacci, T-shirt sizes).
- Auto-detection or localization of issue type names — standard English names are hardcoded.
- OAuth or any Jira auth mechanism other than PAT.
- TLS termination for the HTTP transport (assumed to be handled by a reverse proxy or other infrastructure outside this server).
- Automated/integration tests that exercise the concrete Jira REST client against a real or fake Jira HTTP server (only the `JiraClient`-interface seam is covered by tests in this spec).
- Pagination controls or configuration for `list_epics` / `list_subtasks` beyond a sensible default page size.

## Further Notes

- The Jira Personal Access Token is created by the end user in Jira under Profile → Personal Access Tokens (`/plugins/servlet/pat`) and supplied to the server via `JIRA_PAT`; this is an operational step, not something the server automates.
- Because custom field IDs are auto-detected by display name, a future rename of the "Epic Link" or "Story Points" fields in Jira would break auto-detection; the environment-variable override exists specifically as a manual escape hatch for that case.
- If multi-project support or Jira Cloud support becomes necessary later, the `JiraClient` interface boundary established here should make it possible to add a second implementation without changing tool handler logic or their tests.
