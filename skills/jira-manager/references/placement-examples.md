# Placement examples

Use these for manual verification: run the tree snippet against a project in this shape (or reason over it), give `/jira-manager` each Work Item, and compare the Placement Plan with the expected one. Nothing here is executed automatically.

## Fixture tree (me = `me`)

```
PROJ-1  Epic  v2.3 Release                    (Open)
  PROJ-10 Task  v2.3 QA 驗證（me）              own
    PROJ-20 Sub-task 登入流程回歸測試           own, 0.5 pts
    PROJ-21 Sub-task 結帳流程回歸測試           own, no points
    PROJ-22 Sub-task 會員流程回歸測試           own, 0.7 pts
    PROJ-23 Sub-task 搜尋流程回歸測試           own, 0.3 pts
  PROJ-11 Task  v2.3 效能測試                   assignee: alice
  PROJ-12 Task  v2.3 文件整理                   unassigned
PROJ-2  Epic  v2.2 Release                    (Done — not a candidate)
PROJ-3  Epic  會員系統                          (Open)
PROJ-30 Bug   Safari 登入頁跑版                 (outside the hierarchy)
```

## Cases

| # | Work Item | Expected Placement Plan |
|---|---|---|
| 1 | 我今天要做結帳流程回歸測試，0.5 點 | Reuse PROJ-21 (own, no points) → `setPoints PROJ-21 = 0.5`. No creates. |
| 2 | 我今天要做登入流程回歸測試 | Reuse PROJ-20. It already has 0.5 pts: no reminder, no writes. The plan says nothing needs to change. |
| 3 | 我今天要做 v2.3 購物車回歸測試 | Reuse Task PROJ-10 → new Sub-task. No points given: remind, suggest **0.5** (median of PROJ-20/22/23: 0.3, 0.5, 0.7), labelled 「同 Task 下 3 個 Sub-task 的中位數」. |
| 4 | 我今天要幫忙 v2.3 效能測試，0.7 點 | PROJ-11 matches but belongs to alice → new Own Task under PROJ-1 + `linkRelates` to PROJ-11 → new Sub-task → `setPoints 0.7`. |
| 5 | 我今天要整理 v2.3 文件，0.3 點 | PROJ-12 is unassigned, so it's treated as a Related Task. Same shape as case 4. PROJ-12 is not reassigned. |
| 6 | 我今天要處理 PROJ-30，0.5 點 | Referenced Issue is a Bug (outside the hierarchy) → flag it and ask the user where to place it. No semantic guessing. |
| 7 | 我今天要處理 PROJ-21，0.5 點 | Referenced Issue: an Own, Open Sub-task → reuse it, `setPoints 0.5`. |
| 8 | 我今天要研究 CI 改用 GitHub Actions，1 點 | No Epic matches → new Epic (**⚠ 將新建 Epic**, Epic Name suggested) → new Task → new Sub-task → `setPoints 1`. |
| 9 | 我今天要處理 QA 任務 | Ambiguous (v2.3 QA 驗證, v2.3 效能測試 …) → up to 3 candidates with reasons + "新建". |
| 10 | 我今天要幫忙 v2.3 效能測試 0.5 點、再做 v2.3 壓力測試 0.5 點 | Both Work Items share **one** new Own Task under PROJ-1 (linked to PROJ-11): a single `createTask`, two `createSubtask` with the same `ref` as parent. |
| 11 | 我今天要研究 CI 改用 GitHub Actions | Same shape as case 8, but no points given and the new Task has no Sub-tasks → suggest a value by estimating the work in days (e.g. 1.0), labelled 「依工作內容估計」. |

## Failure case

The plan from case 4, with `createSubtask` returning 400 (e.g. a required field is missing on the Sub-task screen):

- The report shows `createTask` as done (with the new key), `linkRelates` done, `createSubtask` failed (with Jira's message) and `setPoints` skipped.
- Nothing is deleted.
- Re-running the same Work Item must reuse the new Task, because it is now an Own Issue.
