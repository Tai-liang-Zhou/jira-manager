# /jira-place

A Claude Code Skill that places the work you're about to do into a Jira Server/Data Center project's **Epic → Task → Sub-task** hierarchy. It reuses suitable issues, creates the missing levels, and sets Story Points. It shows you a Placement Plan first and writes nothing until you confirm.

```
/jira-place 我今天要處理 v2.3 的 QA 任務，2 點
```

It needs no Personal Access Token or API token: you log in to Jira in a real browser, and the Skill drives [`playwright-cli`](https://github.com/microsoft/playwright-cli) to call Jira's REST API with that browser session. There is no code of our own to build or run.

- [SPEC.md](SPEC.md) / [SPEC.zh-TW.md](SPEC.zh-TW.md): full specification (English / 繁體中文)
- [CONTEXT.md](CONTEXT.md): glossary (Work Item, Placement, Placement Plan, Own Issue, …)
- [docs/adr/](docs/adr/): why browser-session auth (0001) and why `playwright-cli` with no script (0003)

## Setup

1. Install `playwright-cli`:
   ```sh
   npm install -g @playwright/cli@latest
   ```
2. Install the Skill by symlinking it into your user skills, so `/jira-place` works from any directory:
   ```sh
   ln -s "$PWD/skills/jira-place" ~/.claude/skills/jira-place
   ```
3. Run `/jira-place` in Claude Code. On first run it asks for your Jira base URL and project key, and saves them to `~/.config/jira-placement/config.json`. It then opens a browser for you to log in (SSO/MFA included) and detects the "Epic Link", "Epic Name" and "Story Points" custom field IDs.

**Finding your project key:** it's the prefix on every issue in the project (e.g. `PROJ` in `PROJ-123`), also visible in the project URL (`.../projects/PROJ/summary`).

## Files outside the repo

| Path | What it is |
|---|---|
| `~/.config/jira-placement/config.json` | Base URL, project key and cached custom field IDs. Edit the field IDs here if auto-detection picks the wrong field. |
| `~/.config/jira-placement/profile/` | The persistent browser profile holding your Jira login. **Treat it as a credential**: never commit or share it. Delete it to log out. |

## Verifying

`skills/jira-place/examples.md` lists sample Work Items with their expected Placement Plans. For the first run against a real project, use a Work Item that only reuses existing issues, so you can check the session, field IDs and tree before anything is created.
