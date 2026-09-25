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

**Story Points**:
A free-form numeric effort estimate, recorded only on Sub-tasks.
_Avoid_: estimate, points on Tasks/Epics
