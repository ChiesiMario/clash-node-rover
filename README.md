[English](README_en.md) | [繁體中文](README.md) | [简体中文](README_zh-CN.md)

# Clash Node Rover

<p align="center">
  <img src="Image/1.png" alt="Clash Node Rover Dashboard" width="100%" />
</p>

**Clash Node Rover** 是一個基於 Rust 與 Tauri 構建的高效能背景引擎與圖形化管理工具，專門用來監控並自動將您的 Clash / Clash Meta 代理切換到當下最快、最穩定的節點。

---

## ✨ 核心亮點

- 🚀 **智慧動態測速引擎**：在背景使用極低的資源持續並發測試您的代理節點。採用加權平均與抖動 (Jitter) 計算出綜合分數，一旦發現顯著優於當前的節點，即會自動無縫切換。
- 📊 **歷史效能追蹤與圖表**：告別單一時間點的測速！獨家提供高達 7 天的歷史連線延遲圖表，節點在不同時段穩不穩定一目了然。
- 🛡️ **在地網路防護機制 (Local Network Check)**：在執行任何測速與切換前，引擎會自動檢測本機端網路（透過 DNS 探測），避免因本機斷網而誤判所有優質節點失效。
- 🌐 **HTTP 深度連網驗證**：支援在切換前透過指定的 HTTP Proxy 進行真實網頁存取測試，確保節點不僅「Ping 得通」，而且「能順暢連網」。
- 🎨 **現代化多語系介面**：使用 React + TailwindCSS 打造精美的玻璃擬物化 (Glassmorphism) UI，支援繁體中文、簡體中文與英文。

---

## 📸 介面預覽

### 儀表板 (Dashboard)
即時監看系統狀態、API 連線狀態、活躍群組與目前的測速進度。

<img src="Image/1.png" alt="Dashboard" width="100%" />

### 節點排行與歷史圖表 (Node Ranking & History)
將所有節點依照分數排序，支援地區過濾，點擊節點即可展開查看 24H / 3天 / 7天的歷史延遲折線圖。

<img src="Image/2.png" alt="Node Ranking" width="100%" />

### 進階設定 (Settings)
高度客製化的容忍度 (Tolerance)、退避機制 (Backoff Rounds) 與並發數量設定，滿足各種網路環境的需求。

<img src="Image/3.png" alt="Settings" width="100%" />

---

## 📥 安裝與使用

### 1. 下載安裝
請前往 [Releases](../../releases) 頁面下載最新版本的 `.exe` 或 `.msi` 安裝包進行安裝。

### 2. 初始設定 (Setup Wizard)
1. **連接 API**：啟動後，軟體將引導您連接至 Clash API (通常為 `http://127.0.0.1:9090`)，如果您有設定 Secret 請一併輸入。
2. **選擇群組**：勾選您希望 Rover 監控並自動切換的**代理群組 (Selectors)**。
3. **完成**：開始享受全自動的網路優化體驗！

---

## ⚙️ 進階核心機制說明

### 切換容忍度 (Switch Tolerance)
為了避免頻繁切換節點導致網路連線中斷（例如看影片或遊戲時的卡頓），Rover 引入了容忍度機制。當新節點的綜合分數並沒有超過「當前節點分數 + 容忍度」時，系統將維持現有節點。

### 退避演算法 (Backoff Rounds)
對於連續測速超時 (Timeout) 或真實連網測試失敗的節點，系統會根據您設定的次數，將該節點標記為「退避狀態」。處於退避狀態的節點在接下來的幾個測速輪次中將被直接跳過，大幅節省系統發送無效請求的負擔。

---

## 🛠️ 開發與自行編譯

本專案使用 `Tauri v2`、`Rust` 與 `React` 進行開發。如果您希望自行編譯或貢獻程式碼：

```bash
# 1. 安裝前端依賴
npm install

# 2. 啟動開發者模式
npm run dev

# 3. 編譯正式發行版 (Windows 下將產生 .exe 與 .msi)
npm run tauri build
```

## 📄 授權條款
MIT License
