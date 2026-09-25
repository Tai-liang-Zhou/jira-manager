# Jira Work Placement — 規格

[English](SPEC.md) | 繁體中文

> 本文件為 [SPEC.md](SPEC.md) 的中文版，兩者內容應保持一致。領域詞彙（Work Item、Placement、Placement Plan、Own Issue、Related Task、Referenced Issue、Open）沿用 [CONTEXT.md](CONTEXT.md) 的英文原詞，不另行翻譯；背後的決策記錄在 [docs/adr/](docs/adr/)。

## 問題描述

團隊在 Jira Server/Data Center 上用 **Epic → Task → Sub-task** 三層結構管理工作，Story Points 記錄在 Sub-task 上。使用者每天開始一項工作之前，都得先在 Jira 裡翻找這項工作該放在哪裡、判斷是否已經有適合的 Task 或 Sub-task，再手動把缺的單子建出來，需要大量手動點擊。

先前的 Go MCP server 雖然把 Jira 呼叫自動化了，但它是用 Personal Access Token 做認證。公司政策現在禁止使用 PAT 和 API token，所以那個 server 已經不能用了。而且它也從來沒處理最困難的部分：判斷一件工作**該放在哪裡**。

## 解決方案

一個 Claude Code Skill（`/jira-place`），加上一支使用官方 `playwright` 函式庫的 TypeScript 命令列腳本。

- 使用者在 Playwright 開啟的瀏覽器中登入 Jira 一次（SSO/MFA 由使用者自己完成）。腳本把 session 存成 `storageState`，之後用這些 cookie 呼叫 REST API v2，不需要開瀏覽器（ADR 0001）。
- 使用者用一或多個 Work Item 呼叫 `/jira-place`，例如 `/jira-place 我今天要處理 v2.3 的 QA 任務，2 點`。
- Skill（Claude）從腳本讀取 Open 的 Issue 樹，對每個 Work Item 做 Placement，並產出一份合併的 **Placement Plan**。計畫中會說明要沿用或新建哪個 Epic、Task、Sub-task，以及 Story Points（使用者沒給時由 Claude 提供建議值）。
- 使用者確認計畫之後，Claude 把計畫寫成 JSON 檔，由腳本的 `apply` 指令執行（ADR 0002）。

## 使用者故事

1. 身為使用者，我想用自然語言描述我接下來要做的工作，並讓它被放到正確的 Epic 和 Task 底下，這樣我就不必自己在 Jira 裡翻找。
2. 身為使用者，我希望每個 Work Item 最後都對應到剛好一個 Sub-task，這樣這項工作就能被估點。
3. 身為使用者，我沒有給 Story Points 時，希望被提醒，並看到一個參考同 Task 下其他 Sub-task 得出的建議值，以免不小心建出沒估點的 Sub-task。
4. 身為使用者，我希望在任何內容寫入 Jira 之前先看到並確認 Placement Plan，這樣就不會建出我沒同意的東西。
5. 身為使用者，我希望只有 Open 的 Issue（status category 不是 Done）會被列為候選，這樣工作就不會被掛到已經結案的 Epic 或 Task 底下。
6. 身為使用者，我希望只沿用我自己的 Task 和 Sub-task。如果吻合的 Task 屬於別人或沒有指派人，就在同一個 Epic 底下新建一個指派給我的 Task，並用「relates to」連到原本的 Task，這樣我永遠不會寫進別人的單子。
7. 身為使用者，我希望 Epic 不論是誰開的都能沿用，這樣共用的大型計畫不會被重複建立。
8. 身為使用者，當我的 Work Item 提到某個單號時，我希望直接用那張單在階層中的位置決定 Placement，這樣得到的是精確結果，而不是語意上的猜測。
9. 身為使用者，當有好幾個 Epic 或 Task 都說得通時，我希望能從最多 3 個候選中挑選，每個附上理由，再加上一個「新建」選項。如果只有一個明顯吻合，就直接提議它。
10. 身為使用者，當沒有適合的 Epic 時，我希望系統提議新建 Epic，Epic Name 預設等於 summary，並在計畫中醒目標示警告，讓我在建立 Epic 前多看一眼。
11. 身為使用者，我希望一次呼叫就能放置多個 Work Item，並只產出一份合併的計畫，其中一個新建的 Task 可以被多個 Work Item 共用，這樣就不會開出重複的單子。
12. 身為使用者，如果計畫執行到一半失敗，我希望清楚知道哪些 Issue 已經建立（附單號）、哪些還沒有，而且不做任何回滾，這樣重新執行 workflow 時就會沿用已經建好的單子。
13. 身為使用者，我希望能列出 Open 的 Epic → Task → Sub-task 樹，也可以選擇把 Done 的 Issue 一起列出，方便我檢視專案結構。
14. 身為使用者，我希望只有在保存的 session 過期時，才需要透過真的瀏覽器重新登入，而不是每次執行都要登入。
15. 身為使用者，我希望 `/jira-place` 在任何目錄都能使用，但只有在我明確呼叫時才會啟動，這樣隨口一句話不會意外啟動一個會寫入 Jira 的 workflow。

## 實作決策

**執行環境**
- TypeScript，Node 20.6 以上，用 `npx tsx` 執行。存取 Jira 唯一的執行期依賴是官方的 `playwright` 套件。
- 刪除 Go MCP server（`cmd/`、`internal/`、`go.mod`、`Makefile`），保留在 git 歷史中。

**Jira 目標與認證**
- Jira Server/Data Center、REST API v2、單一專案。
- `login` 先用 `GET /rest/api/2/myself` 檢查已保存的 session。如果沒有回 200，就開一個有頭的 Chromium 前往 `JIRA_BASE_URL`，等到該瀏覽器環境中的 `/rest/api/2/myself` 回 200，再儲存 `storageState`。
- session 檔存放在 repo 外的 `~/.config/jira-placement/auth.json`，權限為 `600`。
- 其他指令都用 `request.newContext({ baseURL, storageState })`，不開瀏覽器。寫入請求會帶上 `X-Atlassian-Token: no-check`。
- 如果指令發現 session 已過期，會以一個專用的錯誤結束，告訴 Skill 去執行 `login`。

**設定**
- 設定放在 repo 內的 `.env`，由腳本自行載入：
  - `JIRA_BASE_URL`（必填）
  - `JIRA_PROJECT_KEY`（必填）
  - `JIRA_EPIC_LINK_FIELD`、`JIRA_EPIC_NAME_FIELD`、`JIRA_STORY_POINTS_FIELD`（選填，覆寫用）
- 「Epic Link」、「Epic Name」、「Story Points」這三個 custom field 的 ID，除非被覆寫，否則透過 `GET /rest/api/2/field` 依顯示名稱解析。如果某個欄位解析不到，指令會失敗並給出清楚的錯誤訊息。
- Issue type 名稱固定使用標準英文：`"Epic"`、`"Task"`、`"Sub-task"`。

**腳本指令**（全部以 JSON 輸出到 stdout）
- `login`：確保 session 有效，並輸出目前登入的使用者。
- `tree [--include-done]`：輸出目前使用者和 Epic → Task → Sub-task 樹。每個 Epic 和 Task 包含 key、summary、截斷後的 description、fixVersions、labels、assignee、status。每個 Sub-task 包含 key、summary、assignee、status 和 Story Points（未設定時為 null）。預設只包含 Open 的 Issue。
- `issue <KEY>`：輸出單一 Issue 的類型、summary、status、assignee，以及它的 parent Task 和 Epic（如果有的話）。Skill 處理 Referenced Issue 時會用到。
- `apply <plan.json>`：執行一份 Placement Plan（格式見下），並輸出每一步的結果。

**Placement Plan 格式**（`apply` 的輸入）
- 由有順序的步驟組成。每一步是以下其中之一：`createEpic`、`createTask`、`createSubtask`、`setPoints`、`linkRelates`。
- 建立 Issue 的步驟會帶一個本地的 `ref`（例如 `"new-task-1"`）。後面的步驟可以用真實單號或 `ref` 指向某個 Issue，多個 Work Item 就是靠這個方式共用同一個新建的 Task。
- 新建的 Task 和 Sub-task 指派給目前使用者。新建的 Epic 也指派給目前使用者，並設定 Epic Name。
- 步驟依序執行，遇到第一個失敗就停止。輸出會把每一步標成 `done`（附建立的單號）、`failed`（附 Jira 的錯誤訊息）或 `skipped`。不做任何回滾。

**Skill（`/jira-place`）**
- 原始碼放在 repo 的 `skills/jira-place/SKILL.md`，並用 symlink 連到 `~/.claude/skills/jira-place`。設定 `disable-model-invocation: true`。
- 用絕對路徑呼叫腳本。
- Skill 依照以下順序套用 Placement 規則：
  1. 如果 Work Item 含有單號，就用這個 Referenced Issue 的位置。如果它不屬於 Epic/Task/Sub-task 階層（例如 Bug），就標示出來並詢問使用者。
  2. 否則，用語意將 Work Item 與 Open 的 Epic 和 Task 比對。
  3. 吻合的 Task 只有在它是 Own Issue 時才沿用。否則它就是 Related Task：在它的 Epic 底下建立一個 Own Task，並用「relates to」連結。
  4. 吻合的 Sub-task 只有在它是 Open 且為 Own Issue 時才沿用。如果它已經有 Story Points，一律不修改。
  5. 只有一個明顯吻合時，直接提議它。有好幾個說得通時，提供最多 3 個附理由的選項，再加上「新建」。完全沒有吻合時，提議建立缺少的層級；如果要新建 Epic，就加上 ⚠ 標示。
- 合併的 Placement Plan 會列出每個 Work Item 對應的 Epic、Task、Sub-task（沿用或新建）、Story Points（使用者沒給時標示為建議值），以及所有要建立的連結。
- 使用者確認之前，不會寫入任何東西。

**重試策略**
- 遇到 5xx、網路錯誤和 429 時重試，最多再試 2 次，採用指數退避。遇到 4xx 時立即失敗，並回傳 Jira 的錯誤訊息。

## 測試決策

- **測試縫（seam）**：單一的 `JiraApi` 介面（`myself`、`fields`、`search`、`getIssue`、`createIssue`、`linkIssues`、`updateIssue`），有一個 HTTP 實作，以及一個測試用的記憶體內 fake 實作。
- **風格**：用 Vitest 寫表格驅動測試，驗證指令的輸出以及對 `JiraApi` 的呼叫，不驗證內部實作細節。
- **受測模組：**
  - `apply`：步驟順序；`ref` 解析；多個 Work Item 共用同一個新建的 Task；第 N 步失敗時停止，並正確回報 done/failed/skipped；建立時設定 assignee 和 Epic Name。
  - `tree`：預設排除 Done 的 Issue，加上 `--include-done` 時包含；階層組裝正確；未設定點數的 Sub-task 回報為 null。
  - 欄位解析：依名稱偵測、覆寫值優先、欄位不存在時給出清楚的錯誤。
  - 重試策略：暫時性與非暫時性失敗的區分，以及重試次數。
- **不做自動化測試的部分：** Skill 的 Placement 判斷。`skills/jira-place/examples.md` 收錄範例 Work Item 和對應的預期 Placement Plan，用來手動驗證。HTTP 版的 `JiraApi` 實作則對著真實的 Jira 手動驗證。

## 不在範圍內

- PAT、API token 或 OAuth 認證。
- 透過點擊操作 Jira UI（除非透過 session 呼叫 REST 的方式有一天被封鎖；見 ADR 0001）。
- Jira Cloud 和多專案支援。
- 從一般對話中自動觸發 Skill。
- 回滾只執行了一部分的計畫，或刪除 Issue。
- 改派或編輯別人的 Issue，以及修改已經設定的 Story Points。
- 狀態轉換、sprint、看板、留言、附件。
- 在地化或可設定的 Issue type 名稱。
- 將 Story Points 限制在固定的量表內。

## 補充說明

- session 檔等同於登入憑證，絕對不能 commit 或分享出去。
- session 的有效期限由公司的 Jira/SSO 設定決定。過期時，Skill 會執行 `login`，使用者再登入一次即可。
- 如果 Jira 中的「Epic Link」、「Epic Name」或「Story Points」欄位被改名，自動偵測就會失效；環境變數的覆寫值就是為了這種情況而設計的。
