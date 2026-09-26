# /jira-manager

English | [繁體中文](README.zh-TW.md)

A Claude Code Skill for working with a Jira Server/Data Center project's **Epic → Task → Sub-task** hierarchy. It has two capabilities:

- **Placement** puts the work you're about to do in the right place. It reuses suitable issues, creates the missing levels, and sets Story Points. It shows you a Placement Plan first and writes nothing until you confirm.
- **Points Report** (read-only) shows the Team's Story Points per person. Done points are grouped by week and month of completion, In Progress and Not Started are shown separately, and unestimated Sub-tasks are flagged.

It needs no Personal Access Token or API token: you log in to Jira in a real browser, and the Skill drives [`agent-browser`](https://github.com/vercel-labs/agent-browser) to call Jira's REST API with that browser session. There is no code of our own to build or run.

- [SPEC.md](SPEC.md) / [SPEC.zh-TW.md](SPEC.zh-TW.md): full specification (English / 繁體中文)
- [CONTEXT.md](CONTEXT.md): glossary (Work Item, Placement, Placement Plan, Own Issue, Points Report, …)
- [docs/adr/](docs/adr/): why browser-session auth (0001), why a browser CLI with no script (0003), and why `agent-browser` (0004)

## Requirements

| Dependency | Version | Why |
|---|---|---|
| [Claude Code](https://claude.com/claude-code) | any recent | Runs the Skill |
| [Node.js](https://nodejs.org) + npm | ≥ 18 | Installs `agent-browser` |
| [`agent-browser`](https://github.com/vercel-labs/agent-browser) | latest | Drives the browser and calls Jira's REST API |
| Google Chrome | optional | Preferred browser for SSO (device trust, client certificates). Without it, agent-browser's bundled Chrome for Testing is used. |
| Jira Server / Data Center | REST API v2 | Must have **Epic Link**, **Epic Name** and **Story Points** custom fields, and the issue types **Epic**, **Task** and **Sub-task** |

You also need a Jira account that can log in through the browser and create issues in the project. No API token is needed.

## Installation

1. **Install `agent-browser`** and its bundled browser, then check that it works (skip this if you already use agent-browser):
   ```sh
   npm install -g agent-browser
   agent-browser install
   agent-browser --version
   ```
2. **Clone this repo** anywhere and **link the Skill** into your user skills, so `/jira-manager` works from any directory:
   ```sh
   git clone <this repo> ~/jira-manager
   mkdir -p ~/.claude/skills
   ln -s ~/jira-manager/skills/jira-manager ~/.claude/skills/jira-manager
   ```
   Because it's a symlink, a `git pull` in the repo updates the Skill.
3. **Restart Claude Code** (or start a new session) and check that `/jira-manager` appears when you type `/jira`.

### First run

The first `/jira-manager` call walks you through setup:

1. It asks for your **Jira base URL** (e.g. `https://jira.example.com`) and **project key**, and saves them to `~/.config/jira-manager/config.json`.
   - The project key is the prefix on every issue (e.g. `PROJ` in `PROJ-123`). You can also see it in the project URL (`…/projects/PROJ/summary`).
2. It opens a **browser window** for you to log in, including SSO/MFA. Tell Claude when you're done. Claude never types your credentials.
3. It detects the **Epic Link**, **Epic Name** and **Story Points** custom field IDs and caches them in the config. If one isn't found, it asks you for the ID.

We recommend making the first run a Points Report (`/jira-manager 點數報表`), because it only reads. That confirms the login, field IDs and data before anything is written.

## Usage

`/jira-manager` only runs when you type it; it never starts by itself.

### Placement: put today's work in Jira

```
/jira-manager 我今天要處理 v2.3 的 QA 任務，0.5 點
/jira-manager 我今天要處理 PROJ-456，0.3 點
/jira-manager 我今天要幫忙 v2.3 效能測試 0.5 點、再做 v2.3 壓力測試 0.5 點
```

1. Describe one or more **Work Items**, optionally with Story Points and issue keys.
2. Claude reads the Open Epic → Task → Sub-task tree and decides where each Work Item belongs:
   - It reuses your own Open Task or Sub-task when one fits.
   - If the matching Task is someone else's (or unassigned), it creates **your own Task** under the same Epic and links it to theirs with "relates to". It never edits other people's issues.
   - If several places fit, it shows up to 3 candidates to choose from. If nothing fits, it proposes new issues, and a new Epic is marked **⚠**.
   - If you didn't give points, it reminds you and **suggests** a value: 1 point = 1 day, based on the median of the Task's other Sub-tasks, or an estimate if there are none.
3. You get a **Placement Plan** listing every step that will run. Confirm or change it.
4. Claude runs the steps in order and reports each as done / failed / skipped, with links. On failure it stops and deletes nothing. Re-running reuses what was already created.

### Points Report: Team Story Points

```
/jira-manager 點數報表              # last month + this month
/jira-manager 9 月點數報表
/jira-manager 最近 4 週點數，存成 CSV
```

- **Rows:** one per person, a 「未指派」 row, and a **Team 合計** that excludes unassigned.
- **Done** by week (Mon–Sun) and by month, using the completion date in your local time zone.
- **In Progress / Not Started** as a snapshot of right now.
- Unestimated Sub-tasks are shown as `(+N 未估)`, and the In Progress / Done ones are listed so you can fill them in.
- Ask for CSV to also get `~/Downloads/jira-points-<from>_<to>.csv`. The report is never published anywhere else.

### Login and session

- The browser login is kept in a persistent profile, so you normally log in once and later runs reuse it.
- When the Jira/SSO session expires, the Skill opens the browser again for you to log in, then carries on.
- The `agent-browser` session named `jira` stays open between runs. Run `agent-browser --session jira close` to close it.
- The Skill uses its own profile directory and never connects to your everyday Chrome, so Claude only sees your Jira login.

## Files outside the repo

| Path | What it is |
|---|---|
| `~/.config/jira-manager/config.json` | Base URL, project key and cached custom field IDs. Edit the field IDs here if auto-detection picks the wrong field. |
| `~/.config/jira-manager/profile/` | The persistent browser profile holding your Jira login. **Treat it as a credential**: never commit or share it. Delete it to log out. |

## Troubleshooting

| Symptom | Fix |
|---|---|
| `agent-browser: command not found` | Redo installation step 1. Check that npm's global bin directory is on your `PATH`. |
| Browser fails to launch | Run `agent-browser doctor`. It checks the bundled Chrome and cleans up stale daemons. |
| Asked to log in on every run | Check that `~/.config/jira-manager/profile/` exists and is writable. |
| A write fails with `XSRF check failed` | Your Jira rejected the write request. Report it along with the exact error. |
| Wrong or missing custom field | Fix the ID in `~/.config/jira-manager/config.json` under `fields`. |
| Stale browser processes | `agent-browser close --all` |

## Uninstall

```sh
agent-browser --session jira close
rm ~/.claude/skills/jira-manager          # removes the symlink only
rm -rf ~/.config/jira-manager             # config and saved login
npm uninstall -g agent-browser             # skip if you use it elsewhere
```

## Verifying

`skills/jira-manager/references/placement-examples.md` and `points-report-examples.md` hold sample inputs with their expected Placement Plans and reports, for manual checks after changing the Skill.
