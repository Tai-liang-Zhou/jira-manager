# Placement

Turn one or more **Work Items** into Sub-tasks in the right place in Jira. Setup and the session must already be done (see `SKILL.md`). Snippets referred to in **bold** are in [jira-rest.md](jira-rest.md). Manual verification cases are in [placement-examples.md](placement-examples.md).

**Never write to Jira before the user has confirmed the Placement Plan.**

## 1. Read the tree

Run the **tree** snippet from [jira-rest.md](jira-rest.md). It returns the current user (`me`) and the Open Epic → Task → Sub-task tree. Each Task and Sub-task has an `own` flag. Tasks with no Epic are listed separately in `tasksWithoutEpic`.

For each Work Item that contains an issue key (e.g. `PROJ-456`), also run the **issue position** snippet for that key.

## 2. Placement

Split `$ARGUMENTS` into Work Items. Apply these rules to each, in order:

1. **Referenced Issue.** If the Work Item contains an issue key, place it by that issue's position:
   - An Own, Open Sub-task → reuse it.
   - A Task → treat it as the matching Task in rule 3.
   - An Epic → treat it as the matching Epic.
   - A type outside the hierarchy (e.g. Bug) or a Done issue → flag it and ask the user.
2. **Semantic match.** Otherwise compare the Work Item with the Open Epics and Tasks: summary, description, fixVersions, labels.
3. **Task ownership.** Reuse a matching Task only if `own` is true. Otherwise it's a **Related Task**: plan a new Own Task under the same Epic, and a `linkRelates` from the new Task to it. Unassigned Tasks count as not own. Never reassign or edit other people's issues.
4. **Sub-task.** Reuse a matching Sub-task only if it is Open and `own`. Otherwise plan a new Sub-task under the chosen Task. Never change points that are already set.
5. **Confidence.**
   - One clear match → propose it.
   - Several plausible matches → show at most 3 candidates with a one-line reason each, plus "新建", and let the user choose.
   - No match → propose creating the missing levels. A new Epic gets Epic Name = summary (suggest a shorter one if it's long) and is marked **⚠ 將新建 Epic**.
6. **Story Points.** If the Work Item gives points, use them. If not, remind the user and suggest a value, marked as suggested, using the rule below.

   - 1 point = 1 day of work. Typical values are fractions of a day: 0.3, 0.5, 0.7, 1.0.
   - If the chosen Task already has Sub-tasks with points, suggest the **median** of those points, rounded to one decimal place.
   - If it has none (e.g. a new Task), estimate how much of a day the Work Item takes and express it in the same unit, rounded to one decimal place.
   - Always tell the user which one you used: "同 Task 下 N 個 Sub-task 的中位數" or "依工作內容估計".

7. **Batch.** When several Work Items need the same new Task, create it once and share it through its `ref`.

## 3. Placement Plan

Show one combined plan, one block per Work Item: its Epic, Task and Sub-task (each reused `KEY summary` or **新建** `summary`), the Story Points (and whether they're suggested), and any link. Below that, list the exact steps that will run:

| # | op | ref | fields |
|---|---|---|---|
| 1 | createTask | t1 | epic=PROJ-1, summary="v2.3 QA 驗證（我）" |
| 2 | linkRelates | | from=t1, to=PROJ-11 |
| 3 | createSubtask | s1 | parent=t1, summary="登入流程回歸測試" |
| 4 | setPoints | | issue=s1, points=2 |

The ops are `createEpic` (summary, epicName), `createTask` (epic, summary), `createSubtask` (parent, summary), `setPoints` (issue, points) and `linkRelates` (from, to). Every target is a real key or a `ref` created by an **earlier** step. Refs are lowercase and never look like issue keys.

Wait for the user to confirm or change the plan. If they change it, show the revised plan again.

## 4. Apply

Before the first write, check the whole plan: every op is known, required fields are present, points are non-negative numbers, refs are unique and defined before use. If anything is wrong, fix the plan and ask for confirmation again.

Then run the steps **in order** using the **write** snippets. Every created issue is assigned to `me.name`. After each step:

- **2xx** → record the key (for creates, map the step's `ref` to it) and continue.
- **5xx or 429** → retry the same step up to 2 more times, waiting 1s, then 2s.
- **401, or `loggedIn: false`** → the session expired. Stop, have the user log in (SKILL.md, Setup step 4), then continue from this step.
- **Any other error** → stop. Don't run later steps and don't delete anything.

Finish with a report table listing every step as done (with key), failed (with Jira's message) or skipped, followed by links `<baseUrl>/browse/<KEY>` for every Sub-task. After a failure, tell the user that re-running `/jira-manager` will reuse the issues already created, because they are Own Issues now.
