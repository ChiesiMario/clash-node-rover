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
			Name     string `json:"name"`
			Now      string `json:"now"`
			Provider string `json:"provider"`
			All      int    `json:"all_count"`
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

		type StatNode struct {
			NodeScore
			Provider string `json:"provider"`
		}
		list := make([]StatNode, 0)
		for _, sc := range scores {
			list = append(list, StatNode{
				NodeScore: sc,
				Provider:  GetNodeProvider(sc.Name),
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
		history, err := db.GetNodeHistory(nodeName, 24)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"is_running": rover.IsRunning})
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

        /* Terminal Logs */
        .terminal-container {
            margin-top: 2rem;
            background: var(--bg-base);
            border: 1px solid var(--border-light);
            border-radius: 12px;
            overflow: hidden;
            box-shadow: inset 0 2px 10px rgba(0,0,0,0.5);
        }

        .terminal-header {
            background: rgba(255, 255, 255, 0.03);
            padding: 10px 16px;
            font-family: 'Outfit', sans-serif;
            font-weight: 600;
            color: var(--text-muted);
            border-bottom: 1px solid var(--border-light);
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .terminal-body {
            height: calc(100vh - 250px);
            min-height: 400px;
            overflow-y: auto;
            padding: 16px;
            font-family: 'Consolas', 'Courier New', monospace;
            font-size: 0.85rem;
            line-height: 1.5;
            color: #d1d5db;
        }

        .terminal-body::-webkit-scrollbar {
            width: 8px;
        }
        .terminal-body::-webkit-scrollbar-thumb {
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
        }

        .log-line {
            margin-bottom: 4px;
            word-wrap: break-word;
        }

        .log-time { color: #6b7280; margin-right: 8px; }
        .log-level-info { color: #3b82f6; }
        .log-level-success { color: #10b981; }
        .log-level-warning { color: #f59e0b; }
        .log-level-error { color: #ef4444; font-weight: bold; }
        .log-level-header { color: #c084fc; font-weight: bold; }
        .log-level-muted { color: #6b7280; }
        .log-level-tree { color: #9ca3af; }
        .log-level-group { color: #60a5fa; font-weight: bold; margin-top: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">NODE ROVER</div>
            <button id="triggerBtn" class="btn" onclick="triggerTest()">
                <span id="triggerIcon">🚀</span> <span id="triggerText">立即測速</span>
            </button>
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
                                <th>Rank</th>
                                <th>Node Name</th>
                                <th>Score</th>
                                <th>Success</th>
                                <th>Avg Ping</th>
                                <th>Jitter (σ)</th>
                                <th>Samples</th>
                                <th>Avg Speed</th>
                                <th>Data Used</th>
                                <th>Web Success</th>
                                <th>Web Load</th>
                            </tr>
                        </thead>
                        <tbody id="tbody">
                            <tr><td colspan="11" style="text-align:center;color:var(--text-muted);padding:40px;">Initializing Data...</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>

        <div id="tab-logs" class="tab-content">
            <div class="terminal-container" style="margin-top: 0;">
                <div class="terminal-header">
                    <span style="color: var(--success); font-size: 1.2rem;">&bull;</span> 系統即時日誌 (Live Terminal)
                </div>
                <div id="terminalBody" class="terminal-body">
                    <!-- Logs injected here -->
                </div>
            </div>
        </div>
    </div>

    <script>
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
                    html += '<div class="glass-panel group-card">' +
                            '<div class="group-header">' +
                                '<h3 class="group-title">' + g.name + '</h3>' +
                            '</div>' +
                            '<div class="node-now">' + (g.now || '未選擇') + '</div>' +
                            (g.provider ? '<div style="margin-bottom: 12px; display: inline-block; padding: 4px 10px; border-radius: 6px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1); font-size: 0.8rem; color: #d1d5db;">🏢 ' + g.provider + '</div>' : '') +
                            '<div class="node-meta">' +
                                '<div class="status-dot"></div>' +
                                '<span>當前運行中 &bull; 總計 ' + g.all_count + ' 個節點</span>' +
                            '</div>' +
                        '</div>';
                });
                grid.innerHTML = html;
            } catch(err) {}
        }

        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                const tbody = document.getElementById('tbody');
                
                if (!data || data.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="11" style="text-align:center;color:var(--text-muted);padding:40px;">目前沒有節點數據</td></tr>';
                    return;
                }

                if (tbody.children.length === 1 && tbody.children[0].textContent.includes('Initializing')) {
                    tbody.innerHTML = '';
                }

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
                    const speedStr = node.AvgBandwidth > 0 ? (node.AvgBandwidth > 1000 ? (node.AvgBandwidth/1024).toFixed(2) + ' MB/s' : node.AvgBandwidth.toFixed(1) + ' KB/s') : '<span style="color:var(--text-muted)">-</span>';
                    
                    let consumedStr = '<span style="color:var(--text-muted)">-</span>';
                    if (node.TotalConsumedBytes > 0) {
                        const mb = node.TotalConsumedBytes / (1024 * 1024);
                        if (mb >= 1024) {
                            consumedStr = (mb / 1024).toFixed(2) + ' GB';
                        } else {
                            consumedStr = mb.toFixed(1) + ' MB';
                        }
                    }

                    const rankClass = index < 3 ? 'rank-' + (index+1) : '';
                    let providerTag = node.provider ? '<br><span style="font-size:0.75rem; color:var(--text-muted); background:rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.05); padding:2px 6px; border-radius:4px; display:inline-block; margin-top:4px;">🏢 ' + node.provider + '</span>' : '';

                    let jitterColor = '#94a3b8'; // text-muted
                    // V3: Jitter 現在是標準差，數值比 MAX-MIN 小很多，調整門檻
                    if (node.Jitter > 150) {
                        jitterColor = 'var(--danger)';
                    } else if (node.Jitter > 50) {
                        jitterColor = 'var(--warning)';
                    }

                    const webSuccessColor = node.BrowserSuccessRate >= 0.9 ? 'var(--success)' : (node.BrowserSuccessRate >= 0.5 ? 'var(--warning)' : 'var(--danger)');
                    let webSuccessStr = '<span style="color:var(--text-muted)">-</span>';
                    let webLoadStr = '<span style="color:var(--text-muted)">-</span>';
                    if (node.AvgBrowserLoadTime > 0 || node.BrowserSuccessRate > 0) {
                        webSuccessStr = '<span style="color: ' + webSuccessColor + ';">' + (node.BrowserSuccessRate * 100).toFixed(0) + '%</span>';
                        if (node.AvgBrowserLoadTime > 0) {
                            webLoadStr = (node.AvgBrowserLoadTime / 1000).toFixed(2) + ' s';
                        }
                    }

                    // V3: 顯示樣本數
                    const sampleStr = node.SampleCount || 0;

                    tr.innerHTML = '<td class="rank ' + rankClass + '">#' + (index + 1) + '</td>' +
                        '<td style="font-weight: 600; color: #fff;">' + node.Name + providerTag + '</td>' +
                        '<td><span class="score-badge">' + node.Score + '</span></td>' +
                        '<td class="success-rate" style="color: ' + successColor + ';">' + (node.SuccessRate * 100).toFixed(1) + '%</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif;">' + node.AvgDelay.toFixed(0) + ' ms</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-weight: 600; color: ' + jitterColor + ';">' + node.Jitter + ' ms</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; color: var(--text-muted);">' + sampleStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-weight: 500;">' + speedStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif;">' + consumedStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif; font-weight: 600;">' + webSuccessStr + '</td>' +
                        '<td style="font-family: \'Outfit\', sans-serif;">' + webLoadStr + '</td>';
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
            chartRow.innerHTML = '<td colspan="11"><div class="chart-container"><canvas id="canvas-' + index + '"></canvas></div></td>';
            tr.parentNode.insertBefore(chartRow, tr.nextSibling);

            try {
                const res = await fetch('/api/history?node=' + encodeURIComponent(nodeName));
                const history = await res.json();
                
                if (!history || history.length === 0) {
                    chartRow.innerHTML = '<td colspan="11" style="text-align:center; padding: 40px; color: var(--text-muted);">無歷史資料</td>';
                    return;
                }

                const ctx = document.getElementById('canvas-' + index).getContext('2d');
                
                const labels = history.map(h => {
                    const d = new Date(h.Timestamp * 1000);
                    return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0');
                });
                const data = history.map(h => h.Delay);

                Chart.defaults.color = '#9ca3af';
                Chart.defaults.font.family = "'Inter', sans-serif";

                chartInstances[index] = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: labels,
                        datasets: [{
                            label: 'Ping (ms)',
                            data: data,
                            borderColor: '#8b5cf6',
                            backgroundColor: 'rgba(139, 92, 246, 0.1)',
                            borderWidth: 3,
                            pointBackgroundColor: '#c084fc',
                            pointBorderColor: '#030712',
                            pointBorderWidth: 2,
                            pointRadius: 4,
                            pointHoverRadius: 6,
                            fill: true,
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { display: false },
                            tooltip: {
                                backgroundColor: 'rgba(17, 24, 39, 0.9)',
                                titleFont: { size: 14, family: "'Outfit', sans-serif" },
                                bodyFont: { size: 14, family: "'Inter', sans-serif" },
                                padding: 12,
                                cornerRadius: 8,
                                displayColors: false
                            }
                        },
                        scales: {
                            y: { 
                                beginAtZero: true, 
                                grid: { color: 'rgba(255,255,255,0.05)', drawBorder: false },
                                title: { display: true, text: 'Delay (ms)', color: '#6b7280' }
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
            div.className = 'log-line';
            
            let colorClass = 'log-level-info';
            if (entry.level) {
                colorClass = 'log-level-' + entry.level;
            }

            div.innerHTML = '<span class="log-time">[' + entry.time + ']</span> <span class="' + colorClass + '">' + escapeHtml(entry.message) + '</span>';
            term.appendChild(div);

            // Auto scroll to bottom
            term.scrollTop = term.scrollHeight;

            // Keep only last 200 elements in DOM
            while (term.children.length > 200) {
                term.removeChild(term.firstChild);
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
