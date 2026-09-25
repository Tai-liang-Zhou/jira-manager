# Jira Work Placement

Places the work a user is about to do into the right spot of a Jira project's Epic → Task → Sub-task hierarchy, creating the missing levels when nothing suitable exists.

## Language

**Work Item**:
A natural-language statement from the user of work they are about to do (e.g. "我今天要處理某個 QA 任務"), optionally with Story Points. It is the input to Placement, not a Jira issue, and always resolves to exactly one Sub-task.
_Avoid_: task (collides with the Jira Task level), job, request

**Placement**:
The decision of which Epic and Task a Work Item belongs under, including whether an existing issue is suitable or a new one must be created.
_Avoid_: classification, mapping, 歸類 (in code/docs)

**Referenced Issue**:
An Issue whose key appears in a Work Item (e.g. "處理 PROJ-456"). When present, its position in the hierarchy decides Placement instead of semantic matching.
_Avoid_: linked issue (that is a Jira link type)

**Placement Plan**:
The proposed outcome of a Placement shown to the user before anything is written to Jira: which Epic, Task and Sub-task are reused or created, and the Story Points (suggested when the Work Item has none).
_Avoid_: preview, dry run

**Issue**:
Any Jira ticket in the project, regardless of level.
_Avoid_: ticket, 單子 (in code/docs)

**Epic**:
The top level of the hierarchy: a large initiative that groups Tasks.

**Task**:
The middle level: a unit of work under an Epic that is broken down into Sub-tasks.
_Avoid_: story

**Sub-task**:
The bottom level: a piece of a Task. The only level that carries Story Points.

**Own Issue**:
A Task or Sub-task assigned to the logged-in user. Only Own Issues can be reused in a Placement; Epics can be reused regardless of assignee.
_Avoid_: my ticket

**Related Task**:
A Task that matches a Work Item but is not an Own Issue (assigned to someone else, or unassigned). It points Placement to the right Epic; a new Own Task is created there and linked to it with "relates to".
_Avoid_: other's task, reference task

**Open**:
An Issue whose status category is not Done. Only Open Issues are candidates for Placement.
_Avoid_: active, unresolved

**Progress State**:
Which of three buckets an Issue is in, taken from Jira's status category: **Not Started** (To Do), **In Progress**, or **Done**. Point totals are always reported per Progress State, never mixed.
_Avoid_: status (the workflow status name varies per project; the category does not)

**Points Report**:
A summary of Story Points on Sub-tasks, split by Progress State. Done points are grouped by Period; In Progress and Not Started points are a snapshot of the moment the report is made and have no Period. It has one row per Team member plus a Team total; points on unassigned Sub-tasks are shown on their own row and excluded from the Team total.
_Avoid_: velocity, burndown

**Unestimated Sub-task**:
A Sub-task with no Story Points. It adds nothing to point totals, but a Points Report shows how many there are next to each total, and lists the In Progress and Done ones by key so they can be estimated.
_Avoid_: zero-point task (0 is a real estimate)

**Team**:
Everyone who is the assignee of at least one Sub-task in the project. There is no separately maintained member list.
_Avoid_: group, squad

**Period**:
A calendar week (Monday–Sunday) or calendar month, in the user's local time zone. A Done Sub-task belongs to the Period containing its resolution date. A week that spans two months belongs to the month of its Monday.
_Avoid_: sprint (a Jira concept this project does not use for reporting)

**Story Points**:
A numeric effort estimate in days (1 point = 1 day of work, typically fractions such as 0.3, 0.5, 0.7, 1.0), recorded only on Sub-tasks.
_Avoid_: estimate, points on Tasks/Epics
