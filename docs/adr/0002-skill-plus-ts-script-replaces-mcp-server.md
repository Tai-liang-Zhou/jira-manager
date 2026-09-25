---
status: superseded by ADR-0003
---

# Replace the Go MCP server with a Claude Code Skill plus a TypeScript Playwright script

The original Go MCP server authenticated with a PAT, which is no longer allowed (see ADR 0001). Rather than port it, we split the workflow in two: a TypeScript script using the official `playwright` library performs the fixed Jira operations (login, list tree, create issues, set points) and prints JSON, and a Claude Code Skill handles judgement — deciding whether a suitable Epic/Task/Sub-task already exists — and calls the script. The Go code is removed; it remains in git history.

## Considered Options

- **Keep Go, swap auth to cookies (`playwright-go`)** — rejected: `playwright-go` is community-maintained and lags the official releases; most of the Go code was MCP/HTTP/bearer-auth shell we no longer need.
- **Skill + Playwright MCP / `playwright-cli` with no script** — rejected as the core: Claude would compose raw `fetch` calls every run, which is not reproducible or testable. `playwright-cli` may still be added later if an operation must drive the UI.
- **Standalone agent (Agent SDK)** — rejected: heaviest option and effectively rebuilds the server.

## Consequences

- The browser is only used to log in. The script saves `storageState` and makes all REST calls through `request.newContext({ storageState })`, with no browser open.
- Jira domain rules from the Go client must be carried over: resolving the "Epic Link" / "Story Points" custom field IDs by display name, linking a Sub-task by `parent`, and retrying only on 5xx/429.
