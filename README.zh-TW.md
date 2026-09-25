# /jira-manager

[English](README.md) | 繁體中文

一個 Claude Code Skill，用來處理 Jira Server/Data Center 專案中 **Epic → Task → Sub-task** 三層結構的工作。目前有兩個功能：

- **Placement**：把你接下來要做的工作放到正確的位置。有適合的單子就沿用，缺少的層級就新建，並設定 Story Points。寫入之前一定會先給你看 Placement Plan，你確認之後才會寫入 Jira。
- **Points Report**（唯讀）：列出團隊每個人的 Story Points。Done 的點數會依完成的週和月分組，In Progress 和 Not Started 另外列出，沒有估點的 Sub-task 也會特別標示。

不需要 Personal Access Token 或 API token：你在真的瀏覽器裡登入 Jira，Skill 會透過 [`playwright-cli`](https://github.com/microsoft/playwright-cli)，用這個瀏覽器 session 呼叫 Jira 的 REST API。整個專案沒有任何需要建置或執行的程式碼。

- [SPEC.md](SPEC.md) / [SPEC.zh-TW.md](SPEC.zh-TW.md)：完整規格（English / 繁體中文）
- [CONTEXT.md](CONTEXT.md)：詞彙表（Work Item、Placement、Placement Plan、Own Issue、Points Report 等）
- [docs/adr/](docs/adr/)：為什麼用瀏覽器 session 認證（0001），以及為什麼用 `playwright-cli` 而不寫腳本（0003）

## 需求

| 依賴 | 版本 | 用途 |
|---|---|---|
| [Claude Code](https://claude.com/claude-code) | 近期版本即可 | 執行這個 Skill |
| [Node.js](https://nodejs.org) + npm | ≥ 18 | 安裝與執行 `playwright-cli` |
| [`@playwright/cli`](https://www.npmjs.com/package/@playwright/cli)（`playwright-cli`） | 最新版 | 操作瀏覽器並呼叫 Jira 的 REST API |
| Google Chrome | 可選 | 公司 SSO 如果有裝置信任或用戶端憑證的要求，用 Chrome 比較容易通過。沒有安裝時，會改用 Playwright 內建的 Chromium。 |
| Jira Server / Data Center | REST API v2 | 必須有 **Epic Link**、**Epic Name**、**Story Points** 三個 custom field，以及 **Epic**、**Task**、**Sub-task** 三種 issue type |

另外，你還需要一個能透過瀏覽器登入、並且有權限在該專案建立 Issue 的 Jira 帳號。不需要 API token。

## 安裝

1. **安裝 `playwright-cli`**，並確認可以執行：
   ```sh
   npm install -g @playwright/cli@latest
   playwright-cli --version
   ```
2. **為它安裝瀏覽器**（已經裝了 Google Chrome 就可以跳過這步）：
   ```sh
   playwright-cli install-browser chromium
   ```
3. **把這個 repo clone 到任意位置**，然後**把 Skill 連結到你的使用者 skills 目錄**，這樣在任何目錄都能使用 `/jira-manager`：
   ```sh
   git clone <this repo> ~/jira-manager
   mkdir -p ~/.claude/skills
   ln -s ~/jira-manager/skills/jira-manager ~/.claude/skills/jira-manager
   ```
   因為是 symlink，之後在 repo 裡 `git pull` 就會同時更新 Skill。
4. **重新啟動 Claude Code**（或開一個新的 session），輸入 `/jira` 時應該會看到 `/jira-manager`。

### 第一次執行

第一次呼叫 `/jira-manager` 時，會帶你完成設定：

1. 詢問你的 **Jira base URL**（例如 `https://jira.example.com`）和**專案 key**，並存到 `~/.config/jira-manager/config.json`。
   - 專案 key 是每張單號的前綴，例如 `PROJ-123` 裡的 `PROJ`。也可以從專案網址看到（`…/projects/PROJ/summary`）。
2. 打開一個**瀏覽器視窗**讓你登入，包含 SSO/MFA。登入完成後告訴 Claude 即可。Claude 絕不會代你輸入帳密。
3. 自動偵測 **Epic Link**、**Epic Name**、**Story Points** 三個 custom field 的 ID，並快取在設定檔中。如果某個欄位找不到，會請你提供 ID。

建議第一次執行時先跑點數報表（`/jira-manager 點數報表`）。它只會讀取資料，可以在任何寫入發生之前，先確認登入、欄位 ID 和資料都正確。

## 使用方法

`/jira-manager` 只有在你輸入指令時才會執行，不會自己啟動。

### Placement：把今天的工作放進 Jira

```
/jira-manager 我今天要處理 v2.3 的 QA 任務，0.5 點
/jira-manager 我今天要處理 PROJ-456，0.3 點
/jira-manager 我今天要幫忙 v2.3 效能測試 0.5 點、再做 v2.3 壓力測試 0.5 點
```

1. 描述一或多個 **Work Item**，可以附上 Story Points 和單號。
2. Claude 讀取 Open 的 Epic → Task → Sub-task 樹，判斷每個 Work Item 應該放在哪裡：
   - 有適合的、你自己的 Open Task 或 Sub-task 時，直接沿用。
   - 如果吻合的 Task 是別人的（或沒有指派人），就在同一個 Epic 底下建立**你自己的 Task**，並用「relates to」連到原本那張。絕不修改別人的單子。
   - 有好幾個位置都說得通時，列出最多 3 個候選讓你選。完全沒有吻合的時，提議新建單子；如果要新建 Epic，會標上 **⚠**。
   - 如果你沒給點數，會提醒你並**建議**一個值：1 點 = 1 天，取同一個 Task 底下其他 Sub-task 點數的中位數；沒有可參考的 Sub-task 時，就依工作內容估計。
3. 你會看到一份 **Placement Plan**，列出所有將要執行的步驟。你可以確認，也可以修改。
4. Claude 依序執行，並把每一步標成 done、failed 或 skipped，附上連結。失敗時會停下，不會刪除任何東西。重新執行時，會沿用已經建好的單子。

### Points Report：團隊 Story Points

```
/jira-manager 點數報表              # 上個月 + 本月
/jira-manager 9 月點數報表
/jira-manager 最近 4 週點數，存成 CSV
```

- **列**：每人一列，另外有一列「未指派」，最後是不含未指派的 **Team 合計**。
- **Done** 依完成日按週（週一到週日）和按月分組，時間以你電腦的時區為準。
- **In Progress / Not Started** 是當下的快照。
- 沒有估點的 Sub-task 會顯示為 `(+N 未估)`，其中 In Progress 和 Done 的會列出單號，方便你補點數。
- 要求 CSV 時，會另外存一份 `~/Downloads/jira-points-<from>_<to>.csv`。報表絕不會發布到其他地方。

### 登入與 session

- 瀏覽器的登入狀態保存在一個持久化的 profile 裡，通常只要登入一次，之後的執行都會沿用。
- Jira 或 SSO 的 session 過期時，Skill 會重新打開瀏覽器讓你登入，然後接著繼續。
- 名為 `jira` 的 `playwright-cli` session 在兩次執行之間會保持開啟。要關閉的話，執行 `playwright-cli -s=jira close`。

## repo 以外的檔案

| 路徑 | 內容 |
|---|---|
| `~/.config/jira-manager/config.json` | base URL、專案 key，以及快取的 custom field ID。如果自動偵測抓錯欄位，可以直接在這裡修改。 |
| `~/.config/jira-manager/profile/` | 保存 Jira 登入狀態的持久化瀏覽器 profile。**請把它當作帳密看待**：絕不 commit 或分享。刪除它就等於登出。 |

## 疑難排解

| 症狀 | 解法 |
|---|---|
| `playwright-cli: command not found` | 重做安裝的第 1 步，並確認 npm 的全域 bin 目錄在 `PATH` 裡。 |
| 瀏覽器無法啟動 | 安裝 Chrome，或執行 `playwright-cli install-browser chromium`。 |
| 每次都要重新登入 | 確認 `~/.config/jira-manager/profile/` 存在，而且可以寫入。 |
| 寫入時出現 `XSRF check failed` | 你的 Jira 拒絕了這個寫入請求。請把完整的錯誤訊息回報出來。 |
| custom field 錯誤或缺少 | 在 `~/.config/jira-manager/config.json` 的 `fields` 裡修正 ID。 |
| 殘留的瀏覽器程序 | `playwright-cli kill-all` |

## 解除安裝

```sh
playwright-cli -s=jira close
rm ~/.claude/skills/jira-manager          # 只移除 symlink
rm -rf ~/.config/jira-manager             # 設定檔與已保存的登入狀態
npm uninstall -g @playwright/cli
```

## 驗證

`skills/jira-manager/references/placement-examples.md` 和 `points-report-examples.md` 收錄了範例輸入，以及對應的預期 Placement Plan 和報表。修改 Skill 之後，可以用它們手動檢查。
