package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsClients      = make(map[*websocket.Conn]bool)
	wsClientsMutex sync.Mutex
)

func BroadcastRefresh() {
	wsClientsMutex.Lock()
	defer wsClientsMutex.Unlock()
	for client := range wsClients {
		err := client.WriteJSON(map[string]string{"type": "refresh"})
		if err != nil {
			client.Close()
			delete(wsClients, client)
		}
	}
}

func BroadcastSingleLog(entry WebLogEntry) {
	wsClientsMutex.Lock()
	defer wsClientsMutex.Unlock()
	msg := map[string]interface{}{
		"type":  "log",
		"entry": entry,
	}
	for client := range wsClients {
		if err := client.WriteJSON(msg); err != nil {
			client.Close()
			delete(wsClients, client)
		}
	}
}

func StartWebServer(db *DB, rover *Rover, port int) {
	http.HandleFunc("/", handleIndex)

	http.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		type GroupStatus struct {
			Name     string   `json:"name"`
			Now      string   `json:"now"`
			Provider string   `json:"provider"`
			All      int      `json:"all_count"`
			AllNodes []string `json:"all_nodes"`
		}
		var statuses []GroupStatus
		for _, gName := range rover.GetConfig().TargetGroups {
			g, err := rover.api.GetProxyGroup(gName)
			if err == nil {
				statuses = append(statuses, GroupStatus{
					Name:     gName,
					Now:      g.Now,
					Provider: GetNodeProvider(g.Now),
					All:      len(g.All),
					AllNodes: g.All,
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		scores, err := db.GetScores(rover.GetConfig().HistoryDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		highestInGroups := make(map[string][]string)
		for _, groupName := range rover.GetConfig().TargetGroups {
			g, err := rover.GetAPI().GetProxyGroup(groupName)
			if err == nil {
				highestScore := -999999
				highestNode := ""
				for _, name := range g.All {
					if sc, ok := scores[name]; ok && sc.Score > highestScore {
						highestScore = sc.Score
						highestNode = name
					}
				}
				if highestNode != "" {
					highestInGroups[highestNode] = append(highestInGroups[highestNode], groupName)
				}
			}
		}

		type StatNode struct {
			NodeScore
			Provider          string   `json:"provider"`
			HighestInGroups   []string `json:"highest_in_groups"`
			LastInterviewTime int64    `json:"last_interview_time"`
			CooldownMinutes   int      `json:"cooldown_minutes"`
		}
		list := make([]StatNode, 0)
		for _, sc := range scores {
			t := rover.GetLastInterviewTime(sc.Name)
			var lastInt int64
			if !t.IsZero() {
				lastInt = t.Unix()
			}
			list = append(list, StatNode{
				NodeScore:         sc,
				Provider:          GetNodeProvider(sc.Name),
				HighestInGroups:   highestInGroups[sc.Name],
				LastInterviewTime: lastInt,
				CooldownMinutes:   rover.GetConfig().ExplorationCooldown,
			})
		}

		// Sort by score descending
		for i := 0; i < len(list); i++ {
			for j := i + 1; j < len(list); j++ {
				if list[i].Score < list[j].Score {
					list[i], list[j] = list[j], list[i]
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.URL.Query().Get("node")
		pingHistory, err := db.GetNodeHistory(nodeName, 24)
		browserHistory, _ := db.GetBrowserHistory(nodeName, 24)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ping":    pingHistory,
			"browser": browserHistory,
		})
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"is_running": rover.IsRunning,
			"is_paused":  rover.GetIsPaused(),
		})
	})

	http.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		isPaused := rover.TogglePause()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"is_paused": isPaused})
	})

	http.HandleFunc("/api/switch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Group string `json:"group"`
			Node  string `json:"node"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		err := rover.api.SelectProxy(req.Group, req.Node)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logInfo("⚡ 收到 Web UI 手動切換指令：將群組 [%s] 切換至 %s", req.Group, req.Node)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/trigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		select {
		case rover.ManualTrigger <- struct{}{}:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wsClientsMutex.Lock()
		wsClients[conn] = true
		wsClientsMutex.Unlock()

		// Send log history to new client
		logHistoryMutex.Lock()
		historyCopy := make([]WebLogEntry, len(logHistory))
		copy(historyCopy, logHistory)
		logHistoryMutex.Unlock()

		conn.WriteJSON(map[string]interface{}{
			"type":    "log_history",
			"history": historyCopy,
		})

		// 讓連線保持開啟直到斷線
		go func() {
			defer func() {
				wsClientsMutex.Lock()
				delete(wsClients, conn)
				wsClientsMutex.Unlock()
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🌐 Web 儀表板已啟動，請訪問: http://127.0.0.1%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Web 伺服器啟動失敗: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Clash Node Rover</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Outfit:wght@500;700;800&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --bg-base: #030712;
            --bg-panel: rgba(17, 24, 39, 0.7);
            --bg-panel-hover: rgba(31, 41, 55, 0.8);
            --border-light: rgba(255, 255, 255, 0.08);
            --text-main: #f9fafb;
            --text-muted: #9ca3af;
            --primary: #3b82f6;
            --primary-glow: rgba(59, 130, 246, 0.5);
            --accent: #8b5cf6;
            --success: #10b981;
            --success-bg: rgba(16, 185, 129, 0.15);
            --warning: #f59e0b;
            --danger: #ef4444;
        }
        
        body {
            background-color: var(--bg-base);
            background-image: 
                radial-gradient(circle at 15% 50%, rgba(59, 130, 246, 0.08) 0%, transparent 50%),
                radial-gradient(circle at 85% 30%, rgba(139, 92, 246, 0.08) 0%, transparent 50%);
            color: var(--text-main);
            font-family: 'Inter', sans-serif;
            margin: 0;
            padding: 30px 20px;
            min-height: 100vh;
        }

        .container {
            max-width: 1280px;
            margin: 0 auto;
        }

        h1, h2, h3 {
            font-family: 'Outfit', sans-serif;
            margin: 0;
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
        }

        .logo {
            font-size: 2.5rem;
            font-weight: 800;
            background: linear-gradient(135deg, #60a5fa, #c084fc);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -1px;
        }

        .btn {
            background: linear-gradient(135deg, var(--primary), var(--accent));
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 12px;
            font-family: 'Outfit', sans-serif;
            font-size: 1rem;
            font-weight: 700;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 10px;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px var(--primary-glow);
        }

        .btn:hover:not(:disabled) {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(139, 92, 246, 0.6);
        }

        .btn:disabled {
            background: #374151;
            box-shadow: none;
            cursor: not-allowed;
            color: #9ca3af;
        }

        /* Group Cards */
        .groups-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2.5rem;
        }

        .glass-panel {
            background: var(--bg-panel);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid var(--border-light);
            border-radius: 16px;
            padding: 24px;
            transition: transform 0.3s ease, border-color 0.3s ease;
        }

        .group-card:hover {
            transform: translateY(-4px);
            border-color: rgba(255, 255, 255, 0.2);
            background: var(--bg-panel-hover);
        }

        .group-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 1rem;
        }

        .group-title {
            font-size: 1.25rem;
            font-weight: 700;
            color: #e5e7eb;
        }

        .node-now {
            font-size: 1.5rem;
            font-weight: 800;
            color: #fff;
            margin-bottom: 0.5rem;
            word-break: break-all;
        }

        .node-meta {
            color: var(--text-muted);
            font-size: 0.875rem;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background-color: var(--success);
            box-shadow: 0 0 10px var(--success);
            display: inline-block;
        }

        /* Leaderboard */
        .leaderboard {
            margin-top: 2rem;
        }

        .leaderboard-header {
            margin-bottom: 1.5rem;
        }

        .leaderboard-title {
            font-size: 1.8rem;
            font-weight: 700;
            color: #fff;
            margin-bottom: 0.5rem;
        }

        table {
            width: 100%;
            border-collapse: separate;
            border-spacing: 0;
            text-align: left;
        }

        th {
            padding: 16px;
            font-family: 'Outfit', sans-serif;
            font-weight: 600;
            color: var(--text-muted);
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            border-bottom: 1px solid var(--border-light);
        }

        td {
            padding: 16px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.03);
            font-size: 0.95rem;
            transition: background 0.2s;
            white-space: nowrap;
        }

        tr.node-row:hover td {
            background: rgba(255, 255, 255, 0.03);
            cursor: pointer;
        }

        tr.expanded-row td {
            background: rgba(0, 0, 0, 0.3) !important;
        }

        .rank {
            font-family: 'Outfit', sans-serif;
            font-weight: 800;
            font-size: 1.1rem;
            color: var(--text-muted);
        }

        .rank-1 { color: #fbbf24; text-shadow: 0 0 10px rgba(251, 191, 36, 0.3); }
        .rank-2 { color: #94a3b8; }
        .rank-3 { color: #b45309; }

        .score-badge {
            background: rgba(139, 92, 246, 0.15);
            color: #c084fc;
            padding: 6px 12px;
            border-radius: 8px;
            font-family: 'Outfit', sans-serif;
            font-weight: 700;
            font-size: 1rem;
            border: 1px solid rgba(139, 92, 246, 0.3);
        }

        .success-rate {
            font-weight: 600;
        }

        .chart-container {
            width: 100%;
            height: 350px;
            padding: 20px;
            box-sizing: border-box;
            background: rgba(0, 0, 0, 0.2);
            border-radius: 12px;
            margin: 10px 0;
        }

        @keyframes spin { 100% { transform: rotate(360deg); } }
        
        .spin-icon {
            display: inline-block;
            animation: spin 2s linear infinite;
        }

        /* Tabs */
        .tabs {
            display: flex;
            gap: 12px;
            margin-bottom: 24px;
            border-bottom: 1px solid var(--border-light);
            padding-bottom: 12px;
        }

        .tab-btn {
            background: transparent;
            color: var(--text-muted);
            border: none;
            padding: 8px 16px;
            font-size: 1.1rem;
            font-family: 'Outfit', sans-serif;
            font-weight: 600;
            cursor: pointer;
            border-radius: 8px;
            transition: all 0.2s;
        }

        .tab-btn:hover {
            background: rgba(255, 255, 255, 0.05);
            color: var(--text-main);
        }

        .tab-btn.active {
            background: rgba(59, 130, 246, 0.15);
            color: var(--primary);
            border: 1px solid rgba(59, 130, 246, 0.3);
        }

        .tab-content {
            display: none;
        }

        .tab-content.active {
            display: block;
            animation: fadeIn 0.3s ease;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(5px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Terminal Logs Redesign */
        .terminal-container {
            margin-top: 2rem;
            background: rgba(3, 7, 18, 0.85);
            backdrop-filter: blur(24px);
            -webkit-backdrop-filter: blur(24px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6), inset 0 1px 0 rgba(255, 255, 255, 0.1);
        }

        .terminal-header {
            background: rgba(255, 255, 255, 0.03);
            padding: 14px 20px;
            font-family: 'Outfit', sans-serif;
            font-weight: 600;
            color: #e5e7eb;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
            display: flex;
            align-items: center;
        }

        .mac-dots {
            display: flex;
            gap: 8px;
            margin-right: 16px;
        }
        .mac-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }
        .mac-dot.close { background: #ff5f56; box-shadow: 0 0 10px rgba(255, 95, 86, 0.4); }
        .mac-dot.min { background: #ffbd2e; box-shadow: 0 0 10px rgba(255, 189, 46, 0.4); }
        .mac-dot.max { background: #27c93f; box-shadow: 0 0 10px rgba(39, 201, 63, 0.4); }

        .terminal-title {
            font-size: 0.95rem;
            letter-spacing: 1px;
            color: #9ca3af;
            text-transform: uppercase;
        }

        .terminal-body {
            height: calc(100vh - 280px);
            min-height: 400px;
            overflow-y: auto;
            padding: 20px 24px;
            font-family: 'JetBrains Mono', 'Consolas', monospace;
            font-size: 0.9rem;
            line-height: 1.6;
            color: #d1d5db;
        }

        .terminal-body::-webkit-scrollbar { width: 8px; }
        .terminal-body::-webkit-scrollbar-track { background: rgba(0,0,0,0.2); }
        .terminal-body::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.15); border-radius: 4px; }
        .terminal-body::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.25); }

        .log-line {
            display: flex;
            align-items: flex-start;
            padding: 8px 12px;
            margin-bottom: 8px;
            border-radius: 8px;
            background: rgba(255, 255, 255, 0.02);
            border-left: 4px solid transparent;
            transition: all 0.2s ease;
            animation: slideIn 0.3s ease-out forwards;
        }

        @keyframes slideIn {
            from { opacity: 0; transform: translateX(-10px); }
            to { opacity: 1; transform: translateX(0); }
        }

        .log-line:hover {
            background: rgba(255, 255, 255, 0.04);
            transform: translateX(2px);
        }

        .log-time {
            color: #6b7280;
            font-size: 0.85rem;
            padding-right: 16px;
            white-space: nowrap;
            user-select: none;
            padding-top: 2px;
        }

        .log-content {
            flex: 1;
            word-break: break-word;
        }

        .log-badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 6px;
            font-size: 0.75rem;
            font-weight: 800;
            margin-right: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            vertical-align: middle;
        }

        /* Level Specific Styles */
        .log-level-info { border-left-color: #3b82f6; }
        .log-level-info .log-msg { color: #bfdbfe; }
        .badge-info { background: rgba(59, 130, 246, 0.2); color: #60a5fa; border: 1px solid rgba(59,130,246,0.3); }

        .log-level-success { border-left-color: #10b981; background: rgba(16, 185, 129, 0.05); }
        .log-level-success .log-msg { color: #a7f3d0; font-weight: 500; }
        .badge-success { background: rgba(16, 185, 129, 0.2); color: #34d399; border: 1px solid rgba(16,185,129,0.3); }

        .log-level-warning { border-left-color: #f59e0b; background: rgba(245, 158, 11, 0.05); }
        .log-level-warning .log-msg { color: #fde68a; font-weight: 500; }
        .badge-warning { background: rgba(245, 158, 11, 0.2); color: #fbbf24; border: 1px solid rgba(245,158,11,0.3); }

        .log-level-error { border-left-color: #ef4444; background: rgba(239, 68, 68, 0.08); }
        .log-level-error .log-msg { color: #fca5a5; font-weight: 600; }
        .badge-error { background: rgba(239, 68, 68, 0.2); color: #f87171; border: 1px solid rgba(239,68,68,0.3); }

        .log-level-header {
            border-left: none;
            background: linear-gradient(90deg, rgba(139, 92, 246, 0.15), transparent);
            margin: 20px 0 12px 0;
            padding: 12px 16px;
            border-radius: 8px;
            justify-content: center;
        }
        .log-level-header .log-time { display: none; }
        .log-level-header .log-content {
            text-align: center;
            color: #c084fc;
            font-weight: 800;
            letter-spacing: 2px;
            text-shadow: 0 0 15px rgba(192, 132, 252, 0.4);
        }

        .log-level-group {
            border-left-color: #8b5cf6;
            background: rgba(139, 92, 246, 0.05);
            margin-top: 16px;
        }
        .log-level-group .log-msg { color: #ddd6fe; font-weight: 700; }
        .badge-group { background: rgba(139, 92, 246, 0.2); color: #a78bfa; border: 1px solid rgba(139,92,246,0.3); }

        .log-level-tree {
            border-left-color: transparent;
            padding-left: 28px;
            background: transparent;
        }
        .log-level-tree:hover { background: rgba(255, 255, 255, 0.02); }
        .log-level-tree .log-msg { color: #9ca3af; }
        
        .log-level-muted {
            border-left-color: #4b5563;
        }
        .log-level-muted .log-msg { color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">NODE ROVER</div>
            <div style="display: flex; gap: 12px;">
                <button id="pauseBtn" class="btn" style="background: linear-gradient(135deg, var(--warning), #d97706);" onclick="togglePause()">
                    <span id="pauseIcon">⏸️</span> <span id="pauseText">暫停大腦</span>
                </button>
                <button id="triggerBtn" class="btn" onclick="triggerTest()">
                    <span id="triggerIcon">🚀</span> <span id="triggerText">立即測速</span>
                </button>
            </div>
        </div>

        <div class="tabs">
            <button class="tab-btn active" onclick="switchTab('dashboard')" id="btn-dashboard">📊 儀表板</button>
            <button class="tab-btn" onclick="switchTab('logs')" id="btn-logs">📝 系統日誌</button>
        </div>

        <div id="tab-dashboard" class="tab-content active">
            <div id="groupsGrid" class="groups-grid">
                <!-- Group Cards Injected Here -->
            </div>

            <div class="glass-panel leaderboard">
                <div class="leaderboard-header">
                    <h2 class="leaderboard-title">節點質量排行榜</h2>
                    <div style="color: var(--text-muted);">全球節點綜合評分。點擊節點可展開 24 小時延遲趨勢圖。</div>
                </div>
                
                <div style="overflow-x: auto;">
                    <table id="statsTable">
                        <thead>
                            <tr>
                                <th>排名</th>
                                <th>節點名稱</th>
                                <th>綜合評分</th>
                                <th>Ping 成功率</th>
                                <th>網頁成功率</th>
                                <th>網頁載入</th>
                                <th>上次測試時間</th>
                                <th>上次面試時間</th>
                            </tr>
                        </thead>
                        <tbody id="tbody">
                            <tr><td colspan="6" style="text-align:center;color:var(--text-muted);padding:40px;">Initializing Data...</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <div id="tab-logs" class="tab-content">
            <div class="terminal-container" style="margin-top: 0;">
                <div class="terminal-header">
                    <div class="mac-dots">
                        <div class="mac-dot close"></div>
                        <div class="mac-dot min"></div>
                        <div class="mac-dot max"></div>
                    </div>
                    <div class="terminal-title">System Logs Console</div>
                </div>
                <div id="terminalBody" class="terminal-body">
                    <!-- Logs injected here -->
                </div>
            </div>
        </div>
    </div>

    <script>
        function timeAgo(timestamp) {
            if (!timestamp) return '';
            const seconds = Math.floor(Date.now() / 1000) - timestamp;
            if (seconds < 60) return seconds + ' 秒前';
            const minutes = Math.floor(seconds / 60);
            if (minutes < 60) return minutes + ' 分鐘前';
            const hours = Math.floor(minutes / 60);
            if (hours < 24) return hours + ' 小時前';
            return Math.floor(hours / 24) + ' 天前';
        }

        function switchTab(tabId) {
            document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
            document.querySelectorAll('.tab-btn').forEach(el => el.classList.remove('active'));
            
            document.getElementById('tab-' + tabId).classList.add('active');
            document.getElementById('btn-' + tabId).classList.add('active');

            if (tabId === 'logs') {
                const term = document.getElementById('terminalBody');
                term.scrollTop = term.scrollHeight;
            }
        }

        let chartInstances = {};

        async function fetchGroups() {
            try {
                const res = await fetch('/api/groups');
                const data = await res.json();
                const grid = document.getElementById('groupsGrid');
                
                if (!data || data.length === 0) return;

                let html = '';
                data.forEach(g => {
                    let options = g.all_nodes.map(n => '<option value="' + n + '" ' + (n === g.now ? 'selected' : '') + '>' + n + '</option>').join('');
                    html += '<div class="glass-panel group-card">' +
                            '<div class="group-header">' +
                                '<h3 class="group-title">' + g.name + '</h3>' +
                            '</div>' +
                            '<div class="node-now">' + (g.now || '未選擇') + '</div>' +
                            (g.provider ? '<div style="margin-bottom: 12px; display: inline-block; padding: 4px 10px; border-radius: 6px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1); font-size: 0.8rem; color: #d1d5db;">🏢 ' + g.provider + '</div>' : '') +
                            '<div class="node-meta" style="margin-bottom: 16px;">' +
                                '<div class="status-dot"></div>' +
                                '<span>當前運行中 &bull; 總計 ' + g.all_count + ' 個節點</span>' +
                            '</div>' +
                            '<div style="display: flex; gap: 8px;">' +
                                '<select id="select-' + g.name + '" style="flex:1; background: rgba(0,0,0,0.3); color: white; border: 1px solid var(--border-light); border-radius: 8px; padding: 6px 10px;">' + options + '</select>' +
                                '<button onclick="manualSwitch(\'' + g.name + '\')" style="background: var(--primary); border:none; color: white; border-radius: 8px; padding: 6px 12px; cursor: pointer; font-weight: bold;">強制切換</button>' +
                            '</div>' +
                        '</div>';
                });
                grid.innerHTML = html;
            } catch(err) {}
        }

        async function manualSwitch(groupName) {
            const select = document.getElementById('select-' + groupName);
            if (!select) return;
            const targetNode = select.value;
            try {
                await fetch('/api/switch', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({group: groupName, node: targetNode})
                });
                fetchGroups(); // Refresh UI
            } catch (err) {
                alert("切換失敗: " + err);
            }
        }

        async function togglePause() {
            try {
                const res = await fetch('/api/pause', {method: 'POST'});
                const data = await res.json();
                updatePauseUI(data.is_paused);
            } catch (err) {}
        }

        function updatePauseUI(isPaused) {
            const btn = document.getElementById('pauseBtn');
            const icon = document.getElementById('pauseIcon');
            const text = document.getElementById('pauseText');
            if (isPaused) {
                btn.style.background = 'linear-gradient(135deg, var(--success), #059669)';
                icon.innerText = '▶️';
                text.innerText = '恢復大腦';
            } else {
                btn.style.background = 'linear-gradient(135deg, var(--warning), #d97706)';
                icon.innerText = '⏸️';
                text.innerText = '暫停大腦';
            }
        }

        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                const tbody = document.getElementById('tbody');
                
                if (!data || data.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;color:var(--text-muted);padding:40px;">目前沒有節點數據</td></tr>';
                    return;
                }

                if (tbody.children.length === 1 && tbody.children[0].textContent.includes('Initializing')) {
                    tbody.innerHTML = '';
                }

                window.nodeList = data;
                data.forEach((node, index) => {
                    let tr = document.getElementById('row-' + index);
                    if (!tr) {
                        tr = document.createElement('tr');
                        tr.id = 'row-' + index;
                        tr.className = 'node-row';
                        tr.onclick = () => toggleChart(node.Name, index);
                        tbody.appendChild(tr);
                    }
                    
                    const successColor = node.SuccessRate > 0.9 ? 'var(--success)' : (node.SuccessRate > 0.5 ? 'var(--warning)' : 'var(--danger)');
                    
                    const rankClass = index < 3 ? 'rank-' + (index+1) : '';
                    let providerTag = node.provider ? '<br><span style="font-size:0.75rem; color:var(--text-muted); background:rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.05); padding:2px 6px; border-radius:4px; display:inline-block; margin-top:4px;">🏢 ' + node.provider + '</span>' : '';

                    const webSuccessColor = node.BrowserSuccessRate >= 0.9 ? 'var(--success)' : (node.BrowserSuccessRate >= 0.5 ? 'var(--warning)' : 'var(--danger)');
                    let webSuccessStr = '<span style="color:var(--text-muted)">-</span>';
                    let webLoadStr = '<span style="color:var(--text-muted)">-</span>';
                    if (node.BrowserTested) {
                        webSuccessStr = '<span style="color: ' + webSuccessColor + ';">' + (node.BrowserSuccessRate * 100).toFixed(0) + '%</span>';
                        if (node.AvgBrowserLoadTime > 0) {
                            webLoadStr = (node.AvgBrowserLoadTime / 1000).toFixed(2) + ' s';
                        } else {
                            webLoadStr = '<span style="color:var(--danger)">Timeout</span>';
                        }
                    }

                    let lastTestTime = node.LastBandwidthTime || 0;
                    if (node.LastBrowserTime > lastTestTime) {
                        lastTestTime = node.LastBrowserTime;
                    }
                    let lastTestStr = lastTestTime ? timeAgo(lastTestTime) + ' (' + new Date(lastTestTime * 1000).toLocaleTimeString('zh-TW', {hour: '2-digit', minute:'2-digit', hour12: false}) + ')' : '<span style="color:var(--text-muted)">未測速</span>';

                    let interviewStr = '<span style="color:var(--text-muted)">從未面試</span>';
                    if (node.last_interview_time > 0) {
                        const diffMin = Math.floor(Date.now() / 60000) - Math.floor(node.last_interview_time / 60);
                        const remainMin = node.cooldown_minutes - diffMin;
                        if (remainMin <= 0) {
                            interviewStr = timeAgo(node.last_interview_time) + ' <span style="color:var(--success);font-size:0.8rem;">(冷卻完畢)</span>';
                        } else {
                            interviewStr = timeAgo(node.last_interview_time) + ' <span style="color:var(--warning);font-size:0.8rem;">(冷卻還剩 ' + remainMin + ' 分鐘)</span>';
                        }
                    }

                    let groupBadges = '';
                    if (node.highest_in_groups && node.highest_in_groups.length > 0) {
                        groupBadges = node.highest_in_groups.map(g => '<span class="badge-success" style="margin-left:8px; padding: 2px 6px; font-size: 0.75rem; border-radius: 4px; display: inline-flex; align-items: center;"><svg style="width:12px;height:12px;margin-right:2px;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/></svg>「' + escapeHtml(g) + '」冠軍</span>').join('');
                    }

                    tr.innerHTML = '<td class="rank ' + rankClass + '">#' + (index + 1) + '</td>' +
                        '<td style="font-weight: 600; color: #fff;">' + escapeHtml(node.Name) + providerTag + groupBadges + '</td>' +
                        '<td><span class="score-badge">' + node.Score + '</span></td>' +
                        '<td class="success-rate" style="color: ' + successColor + ';">' + (node.SuccessRate * 100).toFixed(1) + '%</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-weight: 600;">' + webSuccessStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif;">' + webLoadStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-size: 0.85rem;">' + lastTestStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-size: 0.85rem;">' + interviewStr + '</td>';
                });
            } catch (err) {}
        }

        async function toggleChart(nodeName, index) {
            const tr = document.getElementById('row-' + index);
            let chartRow = document.getElementById('chart-row-' + index);
            
            if (chartRow) {
                chartRow.remove();
                tr.classList.remove('expanded-row');
                if (chartInstances[index]) {
                    chartInstances[index].destroy();
                    delete chartInstances[index];
                }
                return;
            }

            document.querySelectorAll('.chart-row').forEach(el => el.remove());
            document.querySelectorAll('tr').forEach(el => el.classList.remove('expanded-row'));
            Object.values(chartInstances).forEach(c => c.destroy());
            chartInstances = {};

            tr.classList.add('expanded-row');
            
            chartRow = document.createElement('tr');
            chartRow.id = 'chart-row-' + index;
            chartRow.className = 'chart-row expanded-row';
            
            const node = window.nodeList[index];
            const speedStr = node.AvgBandwidth > 0 ? (node.AvgBandwidth > 1000 ? (node.AvgBandwidth/1024).toFixed(2) + ' MB/s' : node.AvgBandwidth.toFixed(1) + ' KB/s') : '<span style="color:var(--text-muted)">-</span>';
            let consumedStr = '<span style="color:var(--text-muted)">-</span>';
            if (node.TotalConsumedBytes > 0) {
                const mb = node.TotalConsumedBytes / (1024 * 1024);
                if (mb >= 1024) consumedStr = (mb / 1024).toFixed(2) + ' GB';
                else consumedStr = mb.toFixed(1) + ' MB';
            }
            
            let bwTimeStr = node.LastBandwidthTime ? '<span style="font-size: 0.8rem; color: var(--text-muted); margin-left: 8px;">(' + timeAgo(node.LastBandwidthTime) + ')</span>' : '';
            let pingTimeStr = node.LastPingTime ? '<span style="font-size: 0.8rem; color: var(--text-muted); margin-left: 8px;">(' + timeAgo(node.LastPingTime) + ')</span>' : '';

            let jitterColor = node.Jitter > 150 ? 'var(--danger)' : (node.Jitter > 50 ? 'var(--warning)' : '#9ca3af');

            let detailsHtml = '<div style="display: flex; flex-wrap: wrap; gap: 20px; padding: 20px; background: rgba(0,0,0,0.15); border-radius: 12px; margin-bottom: 15px; border: 1px solid rgba(255,255,255,0.05);">' +
                '<div style="flex: 1; min-width: 140px;"><div style="color:var(--text-muted); font-size:0.85rem; margin-bottom:4px; font-weight:600; text-transform:uppercase; letter-spacing:1px;">平均延遲</div><div style="font-size:1.1rem; font-family:\'Outfit\',sans-serif; color:#fff;">' + node.AvgDelay.toFixed(0) + ' ms</div></div>' +
                '<div style="flex: 1; min-width: 140px;"><div style="color:var(--text-muted); font-size:0.85rem; margin-bottom:4px; font-weight:600; text-transform:uppercase; letter-spacing:1px;">網路抖動 (σ)</div><div style="font-size:1.1rem; font-family:\'Outfit\',sans-serif; color:' + jitterColor + '; font-weight:600;">' + node.Jitter + ' ms</div></div>' +
                '<div style="flex: 1; min-width: 140px;"><div style="color:var(--text-muted); font-size:0.85rem; margin-bottom:4px; font-weight:600; text-transform:uppercase; letter-spacing:1px;">樣本數</div><div style="font-size:1.1rem; font-family:\'Outfit\',sans-serif; color:#fff;">' + (node.SampleCount || 0) + '</div></div>' +
                '<div style="flex: 1; min-width: 140px;"><div style="color:var(--text-muted); font-size:0.85rem; margin-bottom:4px; font-weight:600; text-transform:uppercase; letter-spacing:1px;">平均測速</div><div style="font-size:1.1rem; font-family:\'Outfit\',sans-serif; color:#fff;">' + speedStr + '</div></div>' +
                '<div style="flex: 1; min-width: 140px;"><div style="color:var(--text-muted); font-size:0.85rem; margin-bottom:4px; font-weight:600; text-transform:uppercase; letter-spacing:1px;">已用流量</div><div style="font-size:1.1rem; font-family:\'Outfit\',sans-serif; color:#fff;">' + consumedStr + '</div></div>' +
            '</div>';

            chartRow.innerHTML = '<td colspan="8"><div style="padding: 10px;">' + detailsHtml + '<div class="chart-container"><canvas id="canvas-' + index + '"></canvas></div></div></td>';
            tr.parentNode.insertBefore(chartRow, tr.nextSibling);

            try {
                const res = await fetch('/api/history?node=' + encodeURIComponent(nodeName));
                const dataRaw = await res.json();
                
                if (!dataRaw || !dataRaw.ping || dataRaw.ping.length === 0) {
                    chartRow.innerHTML = '<td colspan="6"><div style="padding: 10px;">' + detailsHtml + '<div style="text-align:center; padding: 40px; color: var(--text-muted);">無歷史趨勢圖表資料</div></div></td>';
                    return;
                }

                const ctx = document.getElementById('canvas-' + index).getContext('2d');
                
                const labels = dataRaw.ping.map(h => {
                    const d = new Date(h.Timestamp * 1000);
                    return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0');
                });
                const pingData = dataRaw.ping.map(h => h.Delay);

                const browserData = dataRaw.ping.map(p => {
                    const b = dataRaw.browser ? dataRaw.browser.find(b => Math.abs(b.Timestamp - p.Timestamp) < 300) : null;
                    return b ? b.LoadTimeMs : null;
                });

                Chart.defaults.color = '#9ca3af';
                Chart.defaults.font.family = "'Inter', sans-serif";

                chartInstances[index] = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: labels,
                        datasets: [
                            {
                                label: 'Ping (ms)',
                                data: pingData,
                                borderColor: '#8b5cf6',
                                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                                borderWidth: 3,
                                pointBackgroundColor: '#c084fc',
                                pointBorderColor: '#030712',
                                pointBorderWidth: 2,
                                pointRadius: 4,
                                pointHoverRadius: 6,
                                fill: true,
                                tension: 0.4,
                                yAxisID: 'y'
                            },
                            {
                                label: 'Browser Load (ms)',
                                data: browserData,
                                borderColor: '#10b981',
                                backgroundColor: 'transparent',
                                borderWidth: 3,
                                pointBackgroundColor: '#10b981',
                                pointBorderColor: '#030712',
                                pointBorderWidth: 2,
                                pointRadius: 6,
                                pointHoverRadius: 8,
                                fill: false,
                                spanGaps: true,
                                tension: 0.1,
                                yAxisID: 'y2'
                            }
                        ]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        interaction: { mode: 'index', intersect: false },
                        plugins: {
                            legend: { 
                                display: true, 
                                position: 'top',
                                labels: { color: '#e5e7eb', font: { family: "'Outfit', sans-serif" } }
                            },
                            tooltip: {
                                backgroundColor: 'rgba(17, 24, 39, 0.9)',
                                titleFont: { size: 14, family: "'Outfit', sans-serif" },
                                bodyFont: { size: 14, family: "'Inter', sans-serif" },
                                padding: 12,
                                cornerRadius: 8,
                                displayColors: true
                            }
                        },
                        scales: {
                            y: { 
                                beginAtZero: true, 
                                position: 'left',
                                grid: { color: 'rgba(255,255,255,0.05)', drawBorder: false },
                                title: { display: true, text: 'Ping Delay (ms)', color: '#8b5cf6' }
                            },
                            y2: { 
                                beginAtZero: true, 
                                position: 'right',
                                grid: { drawOnChartArea: false },
                                title: { display: true, text: 'Browser Load (ms)', color: '#10b981' }
                            },
                            x: { 
                                grid: { display: false }, 
                                ticks: { maxTicksLimit: 12 } 
                            }
                        }
                    }
                });
            } catch (err) {
                chartRow.innerHTML = '<td colspan="10" style="text-align:center; color: var(--danger);">載入圖表失敗</td>';
            }
        }

        async function checkStatus() {
            try {
                const res = await fetch('/api/status');
                const data = await res.json();
                const btn = document.getElementById('triggerBtn');
                const icon = document.getElementById('triggerIcon');
                const text = document.getElementById('triggerText');
                
                if (data.is_running) {
                    btn.disabled = true;
                    icon.innerHTML = '🔄';
                    icon.classList.add('spin-icon');
                    text.innerText = '測速執行中...';
                } else {
                    btn.disabled = false;
                    icon.innerHTML = '🚀';
                    icon.classList.remove('spin-icon');
                    text.innerText = '立即測速';
                }
            } catch (err) {}
        }

        async function triggerTest() {
            const btn = document.getElementById('triggerBtn');
            if (btn.disabled) return;
            try {
                btn.disabled = true;
                const icon = document.getElementById('triggerIcon');
                const text = document.getElementById('triggerText');
                icon.innerHTML = '🔄';
                icon.classList.add('spin-icon');
                text.innerText = '正在發起...';
                await fetch('/api/trigger', { method: 'POST' });
            } catch (err) {
                btn.disabled = false;
            }
        }

        // Initialize first load
        fetchGroups();
        fetchStats();
        checkStatus();

        // WebSocket for real-time updates
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/api/ws';
            const ws = new WebSocket(wsUrl);

            ws.onmessage = function(event) {
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.type === 'refresh') {
                        // Flash UI to indicate real-time update
                        document.body.style.transition = 'background-color 0.3s ease';
                        document.body.style.backgroundColor = '#1a1f3c';
                        setTimeout(() => document.body.style.backgroundColor = '#0f172a', 300);

                        fetchGroups();
                        fetchStats();
                        checkStatus();
                    } else if (msg.type === 'log') {
                        appendLog(msg.entry);
                    } else if (msg.type === 'log_history') {
                        const term = document.getElementById('terminalBody');
                        term.innerHTML = '';
                        if (msg.history) {
                            msg.history.forEach(entry => appendLog(entry));
                        }
                    }
                } catch(e) {}
            };

            ws.onclose = function() {
                setTimeout(connectWebSocket, 3000); // Reconnect on close
            };
        }

        function appendLog(entry) {
            const term = document.getElementById('terminalBody');
            const div = document.createElement('div');
            
            let levelClass = 'log-level-info';
            let badgeClass = 'badge-info';
            let badgeText = '';

            if (['info', 'success', 'warning', 'error', 'header', 'muted', 'tree', 'group'].includes(entry.level)) {
                levelClass = 'log-level-' + entry.level;
            }

            if (entry.level === 'info') { badgeClass = 'badge-info'; badgeText = 'INFO'; }
            else if (entry.level === 'success') { badgeClass = 'badge-success'; badgeText = 'OK'; }
            else if (entry.level === 'warning') { badgeClass = 'badge-warning'; badgeText = 'WARN'; }
            else if (entry.level === 'error') { badgeClass = 'badge-error'; badgeText = 'FAIL'; }
            else if (entry.level === 'group') { badgeClass = 'badge-group'; badgeText = 'GRP'; }

            div.className = 'log-line ' + levelClass;
            
            let badgeHtml = '';
            if (badgeText) {
                badgeHtml = '<span class="log-badge ' + badgeClass + '">' + badgeText + '</span>';
            }

            let msgText = entry.message;
            // Clean up emojis from message if they were prepended by logger
            if (entry.level === 'info' && msgText.startsWith('💡 ')) msgText = msgText.substring('💡 '.length);
            if (entry.level === 'success' && msgText.startsWith('✅ ')) msgText = msgText.substring('✅ '.length);
            if (entry.level === 'warning' && msgText.startsWith('⚠️ ')) msgText = msgText.substring('⚠️ '.length);
            if (entry.level === 'error' && msgText.startsWith('❌ ')) msgText = msgText.substring('❌ '.length);
            if (entry.level === 'error' && msgText.startsWith('🚑 ')) msgText = msgText.substring('🚑 '.length);

            let msgHtml = '<div class="log-content">' + badgeHtml + '<span class="log-msg">' + escapeHtml(msgText) + '</span></div>';
            
            div.innerHTML = '<div class="log-time">[' + entry.time + ']</div>' + msgHtml;
            
            const isScrolledToBottom = term.scrollHeight - term.clientHeight <= term.scrollTop + 50;
            
            term.appendChild(div);

            // Keep only last 200 elements in DOM
            while (term.children.length > 200) {
                const first = term.firstChild;
                const h = first.offsetHeight;
                term.removeChild(first);
                if (!isScrolledToBottom) {
                    term.scrollTop -= h; // Adjust scroll position to prevent jumping
                }
            }

            // Auto scroll to bottom only if user is already at the bottom
            if (isScrolledToBottom) {
                term.scrollTop = term.scrollHeight;
            }
        }

        function escapeHtml(unsafe) {
            return unsafe
                 .replace(/&/g, "&amp;")
                 .replace(/</g, "&lt;")
                 .replace(/>/g, "&gt;")
                 .replace(/"/g, "&quot;")
                 .replace(/'/g, "&#039;");
        }

        connectWebSocket();
        // Fallback polling for status just in case
        setInterval(checkStatus, 5000);
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
