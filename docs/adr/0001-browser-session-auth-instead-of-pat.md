# Authenticate to Jira with a browser session, not a PAT

Company policy forbids Personal Access Tokens and API tokens, so we cannot call Jira REST with a Bearer credential. Instead, the user logs in to Jira manually in a Playwright-driven browser (handling SSO/MFA themselves), and all reads/writes go through Jira's REST API v2 using that browser session's cookies. We verified `/rest/api/2/myself` returns JSON in a logged-in browser, so REST is reachable this way.

## Considered Options

- **PAT / API token** — forbidden by company policy.
- **Drive the Jira UI by clicking (create dialogs, form fields)** — rejected: brittle against Jira upgrades, field layout, and locale; only a fallback if REST via session were ever blocked.
- **Browser login + REST via session cookie** — chosen.

## Consequences

- There is no long-lived credential; every run needs a live, logged-in browser session, and expired sessions require the user to log in again.
- Write requests may need Jira's XSRF handling (e.g. `X-Atlassian-Token: no-check`) since they're cookie-authenticated.
