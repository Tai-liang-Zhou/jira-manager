# Points Report examples

Manual verification cases. The snippet's date logic was checked against this fixture with a fake Jira (`TZ=Asia/Taipei`). The expected report below is what the formatting rules in [points-report.md](points-report.md) should produce from it.

## Fixture (range 「9 月」 = 2026-09-01 … 2026-10-01, Asia/Taipei)

| Key | Assignee | State | Date | Points | Note |
|---|---|---|---|---|---|
| D1 | Alice | Done | resolved Tue 9/1 09:00 +08 | 0.5 | week 8/31–9/6 (partial, Monday in August) |
| D2 | Alice | Done | resolved Sun 9/6 16:30 **UTC** | 0.7 | = Mon 9/7 00:30 local → week 9/7–9/13 |
| D3 | Bob | Done | resolved 9/10 | — | unestimated |
| D4 | Bob | Done | no resolution; status changed 9/15 | 1 | fallback date |
| D5 | Bob | Done | resolved 8/31 23:00 +08 | 9 | before the range → excluded |
| D6 | — | Done | resolved 9/20 | 0.3 | unassigned |
| O1 | Alice | In Progress | | 0.5 | |
| O2 | Alice | In Progress | | — | unestimated → listed |
| O3 | Bob | Not Started | | — | unestimated, **not** listed |
| O4 | — | Not Started | | 1.4 | unassigned |

## Expected report

Range: 2026-09-01 – 2026-09-30

**Done by week**

| | 8/31–9/6* | 9/7–9/13 | 9/14–9/20 | 9/21–9/27 | 9/28–10/4* |
|---|---|---|---|---|---|
| Alice | 0.5 | 0.7 | — | — | — |
| Bob | — | — (+1 未估) | 1 | — | — |
| （未指派） | — | — | 0.3 | — | — |
| **Team 合計** | 0.5 | 0.7 (+1 未估) | 1 | — | — |

**Done by month**

| | 2026-09 |
|---|---|
| Alice | 1.2 |
| Bob | 1 (+1 未估) |
| （未指派） | 0.3 |
| **Team 合計** | 2.2 (+1 未估) |

**Current snapshot**

| | In Progress | Not Started |
|---|---|---|
| Alice | 0.5 (+1 未估) | — |
| Bob | — | — (+1 未估) |
| （未指派） | — | 1.4 |
| **Team 合計** | 0.5 (+1 未估) | — (+1 未估) |

Footnotes: `*` partial weeks; 「1 張 Done 沒有 Resolution，以狀態變更日代替完成日」; 未估點清單: Alice — O2 (In Progress); Bob — D3 (Done).

Checks:
- D5 (8/31 23:00) is excluded even though the JQL pre-filter fetched it.
- Unassigned 0.3 and 1.4 are **not** in Team 合計.
- O3 is counted as unestimated in the snapshot but is not in the 未估點清單 (Not Started).
- 8/31–9/6 is grouped under August (its Monday), but D1 still counts in September's month total. That's why the month-vs-weeks note applies here.
