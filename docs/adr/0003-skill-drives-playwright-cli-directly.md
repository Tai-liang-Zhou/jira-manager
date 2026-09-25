# The Skill drives `playwright-cli` directly, with no script of our own

Supersedes ADR 0002. We still replace the Go MCP server with a Claude Code Skill, but we dropped the TypeScript script. The user wanted the workflow built from an agent browser tool rather than code we maintain. The Skill therefore has Claude drive `playwright-cli` itself: a named session with a persistent profile holds the login, and Jira's REST API v2 is called with `fetch` inside `playwright-cli eval`, so it runs same-origin with the session cookies.

## Considered Options

- **Skill + TypeScript script with the `playwright` library (ADR 0002)** — built and tested, then dropped: it gave unit-tested plan validation and ref resolution, but at the cost of a codebase the user did not want to own.
- **Hybrid: `playwright-cli` for login, TS script for the rest** — rejected: it keeps the maintenance cost and adds a second tool.
- **Playwright MCP / agent-browser** — equivalent capability; `playwright-cli` was chosen because it costs fewer tokens than loading MCP tool schemas.

## Consequences

- The safety rules `apply` enforced in code are now instructions in the Skill (`references/placement.md`): validate the whole Placement Plan before any write, run steps in order, substitute created keys for refs, stop at the first failure and report done/failed/skipped. They can only be checked by manual verification against `references/placement-examples.md`.
- The JavaScript for recurring calls (resolving fields, building the tree) lives as reference snippets inside the Skill directory, so Claude does not improvise it on each run.
- The persistent browser profile under `~/.config/jira-manager/profile` is the login credential.
