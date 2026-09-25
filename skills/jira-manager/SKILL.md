---
name: jira-manager
description: Manage work in a Jira Server/Data Center project's Epic → Task → Sub-task hierarchy — currently Placement (put today's work under the right Epic/Task as a Sub-task, creating missing levels and setting Story Points). Drives Jira through playwright-cli with a browser login session.
argument-hint: "<Work Item>[, <Work Item>…]  e.g. 我今天要處理 v2.3 的 QA 任務，0.5 點"
disable-model-invocation: true
---

# /jira-manager

The single entry point for working with Jira. Vocabulary (Work Item, Placement, Placement Plan, Own Issue, Related Task, Referenced Issue, Open, Story Points) is defined in the repo's `CONTEXT.md`. Talk to the user in their language.

All Jira access goes through `playwright-cli` session `jira`. The JavaScript for every call is in [references/jira-rest.md](references/jira-rest.md). Use those snippets verbatim and only fill in the `__PLACEHOLDERS__`.

## Capabilities

Pick the capability from `$ARGUMENTS`, do **Setup** first, then follow that capability's file.

| The user… | Capability | Follow |
|---|---|---|
| describes work they're about to do (e.g. 「我今天要處理…」, optionally with points or issue keys) | **Placement** | [references/placement.md](references/placement.md) |

If `$ARGUMENTS` is empty, ask what they want to do. If it asks for something not in the table (e.g. analysing Story Point totals), say that capability doesn't exist yet. Don't improvise Jira calls for it.

## Setup and session

1. Check `command -v playwright-cli`. If it's missing, tell the user to run `npm install -g @playwright/cli@latest` and stop.
2. Read `~/.config/jira-manager/config.json`. If it doesn't exist, ask the user for the Jira base URL and project key, then write:
   ```json
   { "baseUrl": "https://jira.example.com", "projectKey": "PROJ", "fields": {} }
   ```
3. If `playwright-cli list` already shows session `jira`, skip to the session check. Otherwise open it (the persistent profile keeps a previous login):
   ```bash
   playwright-cli -s=jira open "<baseUrl>" --browser=chrome --profile="$HOME/.config/jira-manager/profile"
   ```
   If `--browser=chrome` fails (Chrome not installed), retry without it.
4. Run the **session check** snippet. If it reports `loggedIn: false`:
   - `playwright-cli -s=jira close`, then reopen with the same command plus `--headed`.
   - Ask the user to log in in that window (SSO/MFA included) and to tell you when they're done. Then run the session check again. Never type credentials yourself.
5. If `config.fields` lacks `epicLink`, `epicName` or `storyPoints`, run the **resolve fields** snippet and save the IDs into `config.json`. If a field isn't found, ask the user for its custom field ID.

Leave the session open; the next run reuses it. The profile directory is a login credential, so never copy or print its contents.
