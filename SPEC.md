# Jira Work Placement — Spec

English | [繁體中文](SPEC.zh-TW.md)

See [CONTEXT.md](CONTEXT.md) for the glossary (Work Item, Placement, Placement Plan, Own Issue, Related Task, Referenced Issue, Open) and [docs/adr/](docs/adr/) for the decisions behind this design.

## Problem Statement

The team structures work in Jira Server/Data Center as **Epic → Task → Sub-task**, with Story Points recorded on Sub-tasks. Every day, before starting on a piece of work, the user has to browse Jira to find where that work belongs, decide whether a suitable Task or Sub-task already exists, and create whatever is missing, which takes a lot of manual clicking.

An earlier Go MCP server automated the Jira calls, but it authenticated with a Personal Access Token. Company policy now forbids PATs and API tokens, so that server can no longer be used. It also never handled the hard part: deciding **where** a piece of work belongs.

## Solution

A Claude Code Skill (`/jira-place`) plus a TypeScript command-line script built on the official `playwright` library.

- The user logs in to Jira once in a Playwright-driven browser (handling SSO/MFA themselves). The script saves the session as `storageState` and makes all REST API v2 calls with those cookies, without opening a browser (ADR 0001).
- The user invokes `/jira-place` with one or more Work Items, e.g. `/jira-place 我今天要處理 v2.3 的 QA 任務，2 點`.
- The Skill (Claude) reads the Open issue tree from the script, performs Placement for each Work Item, and presents a single **Placement Plan**. The plan says which Epic, Task and Sub-task to reuse or create and gives Story Points, suggested by Claude if the user gave none.
- After the user confirms the plan, Claude writes it to a JSON file and the script's `apply` command executes it (ADR 0002).

## User Stories

1. As a user, I want to state the work I'm about to do in natural language and have it placed under the right Epic and Task, so that I don't have to browse Jira myself.
2. As a user, I want every Work Item to end up as exactly one Sub-task, so that the work can be estimated.
3. As a user, I want to be reminded when I don't give Story Points, along with a suggested value based on sibling Sub-tasks, so that no Sub-task is created unestimated by accident.
4. As a user, I want to see and confirm a Placement Plan before anything is written to Jira, so that nothing is created that I didn't approve.
5. As a user, I want only Open Issues (status category not Done) considered, so that work is never attached to finished Epics or Tasks.
6. As a user, I want only my own Tasks and Sub-tasks reused. If the matching Task belongs to someone else or is unassigned, a new Task assigned to me is created under the same Epic and linked to it with "relates to", so that I never write into other people's tickets.
7. As a user, I want Epics reused regardless of who owns them, so that shared initiatives aren't duplicated.
8. As a user, when my Work Item mentions an issue key, I want that issue's position in the hierarchy to decide Placement, so that I get an exact result instead of a semantic guess.
9. As a user, when several Epics or Tasks plausibly match, I want to pick from at most three candidates, each with a reason, plus a "create new" option. When one clearly matches, I want it proposed directly.
10. As a user, when no Epic fits, I want a new Epic proposed with an Epic Name defaulting to its summary and a prominent warning in the plan, so that I look twice before creating an Epic.
11. As a user, I want to place several Work Items in one invocation with one combined plan, where a newly created Task can be shared by several Work Items, so that duplicates aren't created.
12. As a user, if applying a plan fails partway through, I want to be told exactly which issues were created (with keys) and which weren't, and nothing rolled back, so that re-running the workflow reuses what already exists.
13. As a user, I want to list the Open Epic → Task → Sub-task tree, and optionally include Done issues, so that I can review the project structure.
14. As a user, I want to log in through a real browser only when my saved session has expired, so that I'm not asked to log in on every run.
15. As a user, I want `/jira-place` available from any directory but only started when I invoke it explicitly, so that casual remarks never start a workflow that writes to Jira.

## Implementation Decisions

**Runtime**
- TypeScript on Node (≥ 20.6), run with `npx tsx`. The only runtime dependency for Jira access is the official `playwright` package.
- The Go MCP server (`cmd/`, `internal/`, `go.mod`, `Makefile`) is removed; it remains in git history.

**Jira target & authentication**
- Jira Server/Data Center, REST API v2, single project.
- `login` first checks the saved session with `GET /rest/api/2/myself`. If that isn't 200, it launches a headed Chromium at `JIRA_BASE_URL`, waits until `/rest/api/2/myself` returns 200 in that browser context, and saves `storageState`.
- The session file is stored at `~/.config/jira-placement/auth.json` with mode `600`, outside the repo.
- All other commands use `request.newContext({ baseURL, storageState })` and never open a browser. Write requests send `X-Atlassian-Token: no-check`.
- If a command finds the session expired, it exits with a distinct error telling the Skill to run `login`.

**Configuration**
- `.env` in the repo, loaded by the script itself:
  - `JIRA_BASE_URL` (required)
  - `JIRA_PROJECT_KEY` (required)
  - `JIRA_EPIC_LINK_FIELD`, `JIRA_EPIC_NAME_FIELD`, `JIRA_STORY_POINTS_FIELD` (optional overrides)
- Custom field IDs for "Epic Link", "Epic Name" and "Story Points" are resolved from `GET /rest/api/2/field` by display name, unless overridden. If a field can't be resolved, the command fails with a clear error.
- Issue type names are the standard English `"Epic"`, `"Task"`, `"Sub-task"`.

**Script commands** (all output JSON on stdout)
- `login`: ensures a valid session and outputs the current user.
- `tree [--include-done]`: outputs the current user plus the Epic → Task → Sub-task tree. Each Epic and Task has key, summary, truncated description, fixVersions, labels, assignee and status. Each Sub-task has key, summary, assignee, status and Story Points (null if unset). By default only Open issues are included.
- `issue <KEY>`: outputs one issue's type, summary, status and assignee, plus its parent Task and Epic if it has them. The Skill uses it for Referenced Issues.
- `apply <plan.json>`: executes a Placement Plan (see below) and outputs the result of each step.

**Placement Plan format** (the input to `apply`)
- An ordered list of steps. Each step is one of: `createEpic`, `createTask`, `createSubtask`, `setPoints`, `linkRelates`.
- A step that creates an issue has a local `ref` (e.g. `"new-task-1"`). Later steps can point at an issue by real key or by `ref`, which is how several Work Items share one new Task.
- New Tasks and Sub-tasks are assigned to the current user. New Epics are assigned to the current user, and their Epic Name is set.
- Steps run in order and stop at the first failure. The output lists each step as `done` (with the created key), `failed` (with Jira's error) or `skipped`. Nothing is rolled back.

**Skill (`/jira-place`)**
- Source lives in the repo at `skills/jira-place/SKILL.md` and is symlinked to `~/.claude/skills/jira-place`. It sets `disable-model-invocation: true`.
- It calls the script by absolute path.
- Placement rules, in the order the Skill applies them:
  1. If the Work Item contains an issue key, use that Referenced Issue's position. If it is outside the Epic/Task/Sub-task hierarchy (e.g. a Bug), flag it and ask the user.
  2. Otherwise match the Work Item semantically against Open Epics and Tasks.
  3. A matching Task is reused only if it is an Own Issue. Otherwise it becomes a Related Task: create an Own Task under its Epic and link it with "relates to".
  4. A matching Sub-task is reused only if it is Open and an Own Issue. It is never modified if it already has Story Points.
  5. If there is one clear match, propose it. If several are plausible, offer at most 3 with reasons plus "create new". If nothing matches, propose creating the missing levels, with a ⚠ marker on a new Epic.
- The combined Placement Plan shows each Work Item's Epic, Task and Sub-task (reuse or create), the Story Points (marked as suggested when the user gave none), and every link to be created.
- Nothing is written until the user confirms.

**Retry policy**
- Retry on 5xx, network errors and 429, up to 2 additional attempts with exponential backoff. 4xx responses fail immediately with Jira's error message.

## Testing Decisions

- **Seam:** a single `JiraApi` interface (`myself`, `fields`, `search`, `getIssue`, `createIssue`, `linkIssues`, `updateIssue`) with an HTTP implementation and a fake in-memory implementation for tests.
- **Style:** table-driven Vitest tests that assert on command output and on the calls made to `JiraApi`, not on internal details.
- **Modules tested:**
  - `apply`: step ordering; `ref` resolution; several Work Items sharing one new Task; stop at failure on step N with a correct done/failed/skipped report; assignee and Epic Name set on creation.
  - `tree`: Done issues excluded by default and included with `--include-done`; correct hierarchy assembly; Sub-tasks with unset points reported as null.
  - Field resolution: detection by name, override precedence, and a clear error when a field is missing.
  - Retry policy: transient vs. non-transient failures and retry counts.
- **Not tested automatically:** the Skill's Placement judgement. `skills/jira-place/examples.md` holds sample Work Items with their expected Placement Plans for manual verification. The HTTP `JiraApi` implementation is verified manually against the real Jira.

## Out of Scope

- PAT, API-token or OAuth authentication.
- Driving the Jira UI by clicking (only if REST via session is ever blocked; see ADR 0001).
- Jira Cloud and multi-project support.
- Auto-triggering the Skill from casual conversation.
- Rolling back partially applied plans, or deleting issues.
- Reassigning or editing other people's issues, and changing Story Points that are already set.
- Status transitions, sprints, boards, comments, attachments.
- Localized or configurable issue type names.
- Restricting Story Points to a fixed scale.

## Further Notes

- The session file is equivalent to a login credential and must never be committed or shared.
- Session lifetime is controlled by the company's Jira/SSO configuration. When it expires, the Skill runs `login` and the user logs in again.
- Renaming the "Epic Link", "Epic Name" or "Story Points" fields in Jira breaks auto-detection; the environment overrides exist for that case.
