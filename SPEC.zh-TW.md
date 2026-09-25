# Jira Work Placement — 規格

[English](SPEC.md) | 繁體中文

> 本文件為 [SPEC.md](SPEC.md) 的中文版，兩者內容應保持一致。領域詞彙（Work Item、Placement、Placement Plan、Own Issue、Related Task、Referenced Issue、Open）沿用 [CONTEXT.md](CONTEXT.md) 的英文原詞，不另行翻譯；背後的決策記錄在 [docs/adr/](docs/adr/)。

## 問題描述

團隊在 Jira Server/Data Center 上用 **Epic → Task → Sub-task** 三層結構管理工作，Story Points 記錄在 Sub-task 上。使用者每天開始一項工作之前，都得先在 Jira 裡翻找這項工作該放在哪裡、判斷是否已經有適合的 Task 或 Sub-task，再手動把缺的單子建出來，需要大量手動點擊。

先前的 Go MCP server 雖然把 Jira 呼叫自動化了，但它是用 Personal Access Token 做認證。公司政策現在禁止使用 PAT 和 API token，所以那個 server 已經不能用了。而且它也從來沒處理最困難的部分：判斷一件工作**該放在哪裡**。

## 解決方案

一個直接操作 `playwright-cli` 的 Claude Code Skill（`/jira-place`），不需要我們自己維護任何腳本（ADR 0003）。

- Skill 會在一個持久化的瀏覽器 profile 上，開啟名為 `jira` 的 `playwright-cli` session。使用者在有頭的視窗中登入一次（SSO/MFA 由使用者自己完成），之後的執行都會沿用這個 profile 裡的登入狀態（ADR 0001）。
- 所有對 Jira 的讀寫，都是在 `playwright-cli eval` 裡用 `fetch` 呼叫 REST API v2。這些請求在已登入的頁面中以同源方式執行，會自動帶上 session cookie。每一種呼叫的 JavaScript 都是 Skill 目錄裡固定的參考片段。
- 使用者用一或多個 Work Item 呼叫 `/jira-place`，例如 `/jira-place 我今天要處理 v2.3 的 QA 任務，2 點`。
- Claude 讀取 Open 的 Issue 樹，對每個 Work Item 做 Placement，並產出一份合併的 **Placement Plan**。計畫中會說明要沿用或新建哪個 Epic、Task、Sub-task，以及 Story Points（使用者沒給時由 Claude 提供建議值）。
- 使用者確認之後，Claude 依序執行計畫中的步驟，並回報每一步的結果。

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

**工具**
- 唯一的執行期依賴是 `playwright-cli`（`@playwright/cli`，由使用者自行全域安裝）。Skill 會先檢查它是否存在；如果沒有安裝，就提供安裝指令並停止。
- 刪除 Go MCP server，保留在 git 歷史中。這個 repo 裡沒有任何腳本或套件。

**Skill 結構**
- `skills/jira-place/SKILL.md`：workflow 流程與 Placement 規則。設定 `disable-model-invocation: true`，所以只有 `/jira-place` 能啟動它。
- `skills/jira-place/references/jira-rest.md`：固定的 JavaScript 片段，包括 session 檢查、欄位解析、Issue 樹、Issue 位置查詢和寫入。Claude 只負責填入佔位符，並透過加了引號的 heredoc 執行，確保 shell 不會改動片段內容。
- `skills/jira-place/examples.md`：一棵範例 Issue 樹，搭配多個 Work Item 及對應的預期 Placement Plan。
- 安裝方式是把 `skills/jira-place` symlink 到 `~/.claude/skills/jira-place`。

**Jira 目標與認證**
- Jira Server/Data Center、REST API v2、單一專案。
- Session 的開啟方式是 `playwright-cli -s=jira open <baseUrl> --browser=chrome --profile=~/.config/jira-placement/profile`；如果沒有安裝 Chrome，就改用內建的 Chromium。session 檢查失敗時，Skill 會加上 `--headed` 重新開啟，等使用者登入。Claude 絕不代為輸入帳密。
- session 過期的判斷方式：收到 401，或是在預期 JSON 的地方收到 HTML（SSO 或登入頁）。
- 寫入請求會帶上 `X-Atlassian-Token: no-check`，這是用 cookie 認證寫入時的必要設定。

**設定**
- `~/.config/jira-placement/config.json` 存放 `baseUrl`、`projectKey` 和 `fields`（`epicLink`、`epicName`、`storyPoints` 三個 custom field 的 ID）。第一次執行時，Skill 會詢問 base URL 和專案 key。
- 欄位 ID 透過 `GET /rest/api/2/field` 依顯示名稱（「Epic Link」、「Epic Name」、「Story Points」）解析，並快取在設定檔中。如果自動偵測失敗，使用者可以直接在設定檔裡修改。
- Issue type 名稱固定使用標準英文：`"Epic"`、`"Task"`、`"Sub-task"`。

**讀取**
- Issue 樹：用一次 `eval` 執行三個有分頁的 `POST /rest/api/2/search` 查詢，分別查 Epic、Task、Sub-task。每個查詢都限定在專案內，而且除非要求包含 Done，否則加上 `statusCategory != Done`。回傳結果包含 `me` 和巢狀的樹。每個 Epic 和 Task 有 key、summary、截斷到 300 字元的 description、fixVersions、labels、assignee、status。Task 和 Sub-task 有 `own` 標記，Sub-task 另外有 Story Points（未設定時為 null）。沒有 Epic 的 Task 會另外列出。
- Issue 位置：查詢 Referenced Issue 的類型、是否屬於階層、狀態與歸屬，以及它的 parent Task 和 Epic。

**Placement Plan**
- 一張依序排列的步驟表。每一步是以下其中之一：`createEpic`（summary、epicName）、`createTask`（epic、summary）、`createSubtask`（parent、summary）、`setPoints`（issue、points）、`linkRelates`（from、to）。
- 建立型的步驟會帶一個小寫的 `ref`。後面的步驟可以指向真實單號，或前面步驟定義過的 ref，多個 Work Item 就是靠這個方式共用同一個新建的 Task。
- 新建的 Issue 都指派給目前使用者。新建的 Epic 會設定 Epic Name。
- 第一次寫入之前，Claude 會先檢查整份計畫：op 是否合法、必填欄位是否齊全、點數是否為非負數、ref 是否唯一且在使用前已定義。
- 步驟依序執行。遇到 5xx 或 429 時最多再重試 2 次（間隔 1 秒、2 秒）。遇到 401 時暫停，讓使用者重新登入後，再從同一步繼續。其他錯誤一律停止執行：後面的步驟跳過，已完成的也不回滾。
- 最後的報告會把每一步標成 done（附單號）、failed（附 Jira 的錯誤訊息）或 skipped，並附上每個 Sub-task 的瀏覽連結。

**Placement 規則**（依以下順序套用；確切文字見 `SKILL.md`）
1. Referenced Issue 依它在樹中的位置決定 Placement。如果它不屬於階層（例如 Bug）或已經 Done，就標示出來並詢問使用者。
2. 否則，用語意將 Work Item 與 Open 的 Epic 和 Task 比對。
3. 吻合的 Task 只有在它是 Own Issue 時才沿用。否則它就是 Related Task：在它的 Epic 底下規劃一個新的 Own Task，並用「relates to」連結。沒有指派人的 Task 也算不是自己的。
4. 吻合的 Sub-task 只有在它是 Open 且為 Own Issue 時才沿用。已經設定的點數絕不修改。
5. 只有一個明顯吻合時直接提議；有好幾個時，列出最多 3 個附理由的候選，再加上「新建」；完全沒有吻合時，建立缺少的層級，新建 Epic 時加上 ⚠。
6. 缺少 Story Points 時，提醒使用者並提供建議值：1 點 = 1 天；取所選 Task 底下已估點 Sub-task 的中位數，沒有可參考的 Sub-task 時依工作內容以天數估計，四捨五入到小數點後一位。

## 測試決策

- 沒有程式碼，所以沒有單元測試。workflow 的正確性建立在固定的片段和 `SKILL.md` 中的規則上。
- `references/jira-rest.md` 裡的每個片段都必須能被解析為合法的 JavaScript 函式（修改檔案時做一次語法檢查）。heredoc 加上 `eval` 的執行方式，已經對一個公開的 JSON API 做過端到端驗證，包含中日韓文字、引號和 `$`。
- Placement 判斷和計畫執行，都依照 `skills/jira-place/examples.md` 手動驗證，包括執行到一半失敗的情況。
- 第一次對公司的 Jira 實際執行時，應該用一個只會沿用既有 Issue、不會寫入任何東西的 Work Item，先確認 session、欄位 ID 和 Issue 樹都正確，再建立任何 Issue。

## 不在範圍內

- PAT、API token 或 OAuth 認證。
- 透過點擊操作 Jira UI（除非透過 session 呼叫 REST 的方式有一天被封鎖；見 ADR 0001）。
- 維護我們自己的腳本或套件（見 ADR 0003）。
- Jira Cloud 和多專案支援。
- 從一般對話中自動觸發 Skill。
- 回滾只執行了一部分的計畫，或刪除 Issue。
- 改派或編輯別人的 Issue，以及修改已經設定的 Story Points。
- 狀態轉換、sprint、看板、留言、附件。
- 在地化或可設定的 Issue type 名稱。
- 將 Story Points 限制在固定的量表內。

## 補充說明

- profile 目錄 `~/.config/jira-placement/profile` 等同於登入憑證，絕對不能 commit、複製或分享出去。
- session 的有效期限由公司的 Jira/SSO 設定決定。過期時，Skill 會以有頭模式重新開啟瀏覽器，讓使用者再登入一次。
- 計畫的安全規則（先驗證、依序執行、失敗就停）是寫給 Claude 的指示，不是程式碼。如果發現 Claude 沒有遵守這些規則，就是該重新檢討 ADR 0003、把 `apply` 改回腳本的訊號。
