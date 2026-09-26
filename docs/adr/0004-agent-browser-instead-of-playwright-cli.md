# Drive Jira with `agent-browser` instead of `playwright-cli`

Amends ADR 0003. The Skill still drives a browser CLI directly, with no script of our own. Only the CLI changes. ADR 0003 picked `playwright-cli` because it costs fewer tokens than loading Playwright MCP's tool schemas. `agent-browser` is also a CLI, so that reason doesn't separate the two. `agent-browser` was already installed and used on the maintainer's machine, while `playwright-cli` would be a second browser CLI installed only for this Skill.

The Skill only needs three things from the tool: a persistent profile that keeps the login, a headed window for the user to log in, and `eval` of JavaScript in the page. `agent-browser` covers all three (`--profile <dir>`, `--headed`, `eval --stdin`). We checked each one before switching: the profile survives a close and reopen, `eval --stdin` passes CJK text, quotes and `$` through unchanged, a thrown error exits 1, and `--executable-path` launches the system Google Chrome.

## Considered Options

- **Keep `playwright-cli`** — rejected: works, but adds a dependency that duplicates a tool already in use.
- **`agent-browser --auto-connect` to reuse the user's everyday Chrome** — rejected: it would skip the separate login, but it gives Claude every site the user is logged in to. That breaks ADR 0001's isolated, Jira-only profile.
- **`agent-browser` with a dedicated `--profile` directory** — chosen.

## Consequences

- `agent-browser eval` evaluates an expression and doesn't call a function the way `playwright-cli eval` does. Every snippet in `references/jira-rest.md` is therefore an immediately invoked `(async () => { … })()`. A bare `async () => { … }` prints `{}` with no error, which the session check would silently misread. So the trailing `()` is load-bearing.
- `agent-browser` launches its bundled Chrome for Testing by default. The Skill passes `--executable-path` to the system Google Chrome, because company SSO may depend on device trust or client certificates, and it falls back to the bundled browser only when Chrome is missing.
- The profile path (`~/.config/jira-manager/profile`) is unchanged. A profile created by `playwright-cli` should keep its login, but this hasn't been verified against the company SSO. If it doesn't, the user logs in once more.
