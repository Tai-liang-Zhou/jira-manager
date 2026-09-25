# Points Report

Summarise the Team's Story Points on Sub-tasks, split by Progress State. Setup and the session must already be done (see `SKILL.md`). The data comes from the **points** snippet in [jira-rest.md](jira-rest.md). Manual verification cases are in [points-report-examples.md](points-report-examples.md).

This capability only reads. It never writes to Jira.

## 1. Range

Read the range from the user's words and turn it into local dates `[FROM, TO)`:

- 「9 月」 → `2026-09-01` to `2026-10-01`
- 「最近 4 週」 → the Monday 3 weeks before this week's Monday, to next Monday
- Nothing said → from the first day of **last month** to the first day of **next month** (last month and this month)

Say which range you used at the top of the report.

## 2. Fetch

Run the **points** snippet with `__FROM__` / `__TO__` filled in. It returns:

- `weeks`: every week (Monday–Sunday) that overlaps the range, with `label` (e.g. `9/7–9/13`), the `month` it belongs to (the month of its Monday), and `partial` if the range cuts it.
- `people`: one entry per assignee. `name: null` means unassigned. Each entry has Done `weeks` / `months` cells, plus the `inProgress` and `notStarted` snapshot. Every cell is `{ points, unestimated }`.
- `fallbackCount`: how many Done Sub-tasks had no resolution date and were dated by their status-category change.
- `unestimatedKeys`: In Progress and in-range Done Sub-tasks with no points.

## 3. Report

Render markdown tables in the terminal. Don't invent numbers; everything comes from the snippet output.

**Cell format:** `points` (written `—` when it is 0), followed by `(+N 未估)` when `unestimated > 0`. For example: `1.2`, `0.7 (+1 未估)`, `— (+1 未估)`, `—`. Round to at most 2 decimals.

**Rows:** one per Team member, sorted by display name, then `（未指派）` if it has any numbers, then a separator and **Team 合計**. Team 合計 sums the members only, never the unassigned row. That includes the unestimated counts.

**Tables, in this order:**

1. **Done by week**: one column per week in `weeks`, with the header `label`, suffixed `*` when `partial`. Group the columns under their `month`.
2. **Done by month**: one column per calendar month in the range. A month total counts Sub-tasks by their actual resolution date, so it can differ from the sum of the weeks listed under that month (weeks are grouped by their Monday). Say so in one line under the table when it happens.
3. **Current snapshot**: columns In Progress and Not Started. These have no Period; they are the state right now.

**Footnotes:**
- `*` weeks: 「部分週：只含範圍內的日期」.
- If `fallbackCount > 0`: 「N 張 Done 沒有 Resolution，以狀態變更日代替完成日」.
- **未估點清單**: `unestimatedKeys` grouped by assignee, as `KEY summary (state)` with a link `<baseUrl>/browse/<KEY>`.

## 4. CSV (only when asked)

If the user asks for CSV (「存成 CSV」、「匯出」), also write `~/Downloads/jira-points-<FROM>_<TO>.csv` with one row per person per Period or snapshot. Columns: `person,kind,period,points,unestimated`, where `kind` is `done-week`, `done-month`, `in-progress` or `not-started`, and `period` is the week's Monday, the `YYYY-MM` month, or empty for snapshots. Include the unassigned rows with `person` = `（未指派）`, but not a Team total row. Tell the user the path.

Never publish the report anywhere else: it is internal workload data.
