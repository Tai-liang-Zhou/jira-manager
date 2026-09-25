# Jira Work Placement — Spec

English | [繁體中文](SPEC.zh-TW.md)

See [CONTEXT.md](CONTEXT.md) for the glossary (Work Item, Placement, Placement Plan, Own Issue, Related Task, Referenced Issue, Open) and [docs/adr/](docs/adr/) for the decisions behind this design.

## Problem Statement

The team structures work in Jira Server/Data Center as **Epic → Task → Sub-task**, with Story Points recorded on Sub-tasks. Every day, before starting on a piece of work, the user has to browse Jira to find where that work belongs, decide whether a suitable Task or Sub-task already exists, and create whatever is missing, which takes a lot of manual clicking.

An earlier Go MCP server automated the Jira calls, but it authenticated with a Personal Access Token. Company policy now forbids PATs and API tokens, so that server can no longer be used. It also never handled the hard part: deciding **where** a piece of work belongs.

## Solution

A Claude Code Skill (`/jira-manager`) that drives `playwright-cli` directly, with no script of our own (ADR 0003).

- The Skill opens a named `playwright-cli` session (`jira`) on a persistent browser profile. The user logs in once in a headed window (handling SSO/MFA themselves), and the profile keeps the login for later runs (ADR 0001).
- All Jira reads and writes are REST API v2 calls made with `fetch` inside `playwright-cli eval`. They run same-origin in the logged-in page and carry its session cookies. The JavaScript for each call is a fixed reference snippet in the Skill directory.
- The user invokes `/jira-manager` with one or more Work Items, e.g. `/jira-manager 我今天要處理 v2.3 的 QA 任務，0.5 點`.
- Claude reads the Open issue tree, performs Placement for each Work Item, and presents one combined **Placement Plan**. The plan says which Epic, Task and Sub-task to reuse or create and gives Story Points, suggested by Claude if the user gave none.
- After the user confirms, Claude executes the plan's steps in order and reports the result of each.

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
15. As a user, I want `/jira-manager` available from any directory but only started when I invoke it explicitly, so that casual remarks never start a workflow that writes to Jira.

## Implementation Decisions

**Tooling**
- `playwright-cli` (`@playwright/cli`, installed globally by the user) is the only runtime dependency. The Skill checks for it and stops with install instructions if it's missing.
- The Go MCP server is removed; it remains in git history. There is no script or package in this repo.

**Skill layout**
- `skills/jira-manager/SKILL.md`: the entry point. It picks a capability from the arguments (currently only Placement) and does the shared setup and session handling. It sets `disable-model-invocation: true`, so only `/jira-manager` starts it. Future capabilities (e.g. Story Point analysis) are added as rows in its capability table plus their own reference file.
- `skills/jira-manager/references/placement.md`: the Placement workflow, rules, Placement Plan and apply procedure.
- `skills/jira-manager/references/jira-rest.md`: fixed JavaScript snippets for the session check, field resolution, tree, issue position and writes. Claude only fills in placeholders, and runs each snippet through a quoted heredoc so the shell never alters it.
- `skills/jira-manager/references/placement-examples.md`: a fixture tree with Work Items and their expected Placement Plans.
- It is installed by symlinking `skills/jira-manager` to `~/.claude/skills/jira-manager`.

**Jira target & authentication**
- Jira Server/Data Center, REST API v2, single project.
- Session: `playwright-cli -s=jira open <baseUrl> --browser=chrome --profile=~/.config/jira-manager/profile`, falling back to bundled Chromium if Chrome is missing. If the session check fails, the Skill reopens with `--headed` and waits for the user to log in. Claude never types credentials.
- An expired session shows up as a 401, or as HTML (an SSO/login page) where JSON was expected.
- Writes send `X-Atlassian-Token: no-check`, which is required for cookie-authenticated writes.

**Configuration**
- `~/.config/jira-manager/config.json` holds `baseUrl`, `projectKey` and `fields` (the `epicLink`, `epicName` and `storyPoints` custom field IDs). On first run, the Skill asks for the base URL and project key.
- Field IDs are resolved from `GET /rest/api/2/field` by display name ("Epic Link", "Epic Name", "Story Points") and cached in the config. The user can edit them there if detection fails.
- Issue type names are the standard English `"Epic"`, `"Task"`, `"Sub-task"`.

**Reading**
- Tree: one `eval` runs three paginated `POST /rest/api/2/search` queries (Epics, Tasks, Sub-tasks, each scoped to the project and to `statusCategory != Done` unless Done issues are requested). It returns `me` plus the nested tree. Each Epic and Task has key, summary, a description truncated to 300 characters, fixVersions, labels, assignee and status. Tasks and Sub-tasks have an `own` flag, and Sub-tasks have Story Points (null if unset). Tasks with no Epic are returned separately.
- Issue position: for a Referenced Issue, its type, whether it is in the hierarchy, its status and ownership, plus its parent Task and Epic.

**Placement Plan**
- A table of ordered steps. Each step is one of `createEpic` (summary, epicName), `createTask` (epic, summary), `createSubtask` (parent, summary), `setPoints` (issue, points) or `linkRelates` (from, to).
- Create steps carry a lowercase `ref`. Later steps target a real key or a ref defined by an earlier step, which is how several Work Items share one new Task.
- New issues are assigned to the current user. New Epics get an Epic Name.
- Before the first write, Claude checks the whole plan (known ops, required fields, non-negative points, unique refs defined before use).
- Steps run in order. 5xx/429 are retried up to 2 more times (1s, 2s backoff). A 401 pauses for re-login and then resumes from that step. Any other error stops execution: later steps are skipped and nothing is rolled back.
- The final report lists each step as done (with key), failed (with Jira's message) or skipped, plus a browse link for every Sub-task.

**Placement rules** (applied in this order; see `SKILL.md` for the exact wording)
1. A Referenced Issue decides Placement by its position. If it is outside the hierarchy (e.g. Bug) or Done, it is flagged and the user is asked.
2. Otherwise the Work Item is matched semantically against Open Epics and Tasks.
3. A matching Task is reused only if it is an Own Issue. Otherwise it is a Related Task: a new Own Task is planned under its Epic and linked with "relates to". Unassigned counts as not own.
4. A matching Sub-task is reused only if it is Open and an Own Issue. Points that are already set are never changed.
5. One clear match → propose it. Several → at most 3 candidates with reasons plus "create new". None → create the missing levels, with ⚠ on a new Epic.
6. Missing Story Points → remind the user and suggest a value: 1 point = 1 day; the median of the chosen Task's estimated Sub-tasks, or an estimate in days when it has none, rounded to one decimal place.

## Testing Decisions

- There is no code, so there are no unit tests. The workflow's correctness rests on the fixed snippets and the rules in `SKILL.md`.
- Every snippet in `references/jira-rest.md` must parse as a JavaScript function (checked by a syntax pass when the file is edited). The heredoc + `eval` path has been checked end to end against a public JSON API, including CJK text, quotes and `$`.
- Placement judgement and plan execution are verified manually against `skills/jira-manager/references/placement-examples.md`, including the partial-failure case.
- The first real run against the company Jira should use a Work Item that only reuses existing issues (no writes), to confirm the session, field IDs and tree before any issue is created.

## Out of Scope

- PAT, API-token or OAuth authentication.
- Driving the Jira UI by clicking (only if REST via session is ever blocked; see ADR 0001).
- Maintaining a script or package of our own (see ADR 0003).
- Jira Cloud and multi-project support.
- Auto-triggering the Skill from casual conversation.
- Rolling back partially applied plans, or deleting issues.
- Reassigning or editing other people's issues, and changing Story Points that are already set.
- Status transitions, sprints, boards, comments, attachments.
- Localized or configurable issue type names.
- Restricting Story Points to a fixed scale.

## Further Notes

- The profile directory `~/.config/jira-manager/profile` is equivalent to a login credential and must never be committed, copied or shared.
- Session lifetime is controlled by the company's Jira/SSO configuration. When it expires, the Skill reopens the browser headed for the user to log in again.
- The plan-safety rules (validate first, in order, stop on failure) are instructions, not code. If Claude is ever seen deviating from them, that is the signal to revisit ADR 0003 and bring back a script for `apply`.
