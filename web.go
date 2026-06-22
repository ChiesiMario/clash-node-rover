package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

type WebServer struct {
	port int
	db   *DB
	cfg  *Config
}

func NewWebServer(cfg *Config, db *DB) *WebServer {
	return &WebServer{
		port: cfg.WebPort,
		db:   db,
		cfg:  cfg,
	}
}

func (s *WebServer) Start() {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/stats", s.handleStats)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 Web 儀表板已啟動，請訪問: http://127.0.0.1%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Web 伺服器啟動失敗: %v", err)
	}
}

func (s *WebServer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	scores, err := s.db.GetScores(s.cfg.HistoryDays)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list := make([]NodeScore, 0)
	for _, sc := range scores {
		list = append(list, sc)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	json.NewEncoder(w).Encode(list)
}

func (s *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Node Rover Dashboard</title>
    <style>
        :root {
            --bg-color: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --accent: #3b82f6;
            --accent-hover: #60a5fa;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
        }
        body {
            margin: 0;
            padding: 0;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%);
            color: var(--text-main);
            min-height: 100vh;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
            padding: 2rem;
        }
        h1 {
            text-align: center;
            font-weight: 300;
            letter-spacing: 2px;
            margin-bottom: 2rem;
            color: var(--accent-hover);
        }
        .glass-panel {
            background: var(--card-bg);
            backdrop-filter: blur(12px);
            -webkit-backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 1.5rem;
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
            margin-bottom: 2rem;
            animation: fadeIn 0.5s ease-out;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 1rem;
            text-align: left;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }
        th {
            color: var(--text-muted);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 1px;
        }
        tr:hover td {
            background: rgba(255, 255, 255, 0.03);
            transition: background 0.3s;
        }
        .score-badge {
            background: rgba(59, 130, 246, 0.2);
            color: var(--accent-hover);
            padding: 4px 10px;
            border-radius: 20px;
            font-size: 0.9rem;
            font-weight: bold;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 NODE ROVER</h1>
        <div class="glass-panel">
            <h2>質量分數排行榜 (近期記錄)</h2>
            <table id="statsTable">
                <thead>
                    <tr>
                        <th style="padding: 15px;">排名</th>
                        <th style="text-align: left; padding: 15px;">節點名稱</th>
                        <th style="padding: 15px;">質量分數</th>
                        <th style="padding: 15px;">成功率</th>
                        <th style="padding: 15px;">平均延遲 (MS)</th>
                        <th style="padding: 15px;">平均下載速度</th>
                        <th style="padding: 15px;">累計消耗流量</th>
                    </tr>
                </thead>
                <tbody>
                    <tr><td colspan="7" style="text-align:center;color:var(--text-muted);">載入中...</td></tr>
                </tbody>
            </table>
        </div>
    </div>
    <script>
        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                const tbody = document.querySelector('#statsTable tbody');
                tbody.innerHTML = '';
                
                if (!data || data.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:var(--text-muted);">目前還沒有足夠的測速資料，請稍候...</td></tr>';
                    return;
                }

                data.forEach((node, index) => {
                    const tr = document.createElement('tr');
                    const successColor = node.SuccessRate > 0.9 ? 'var(--success)' : (node.SuccessRate > 0.5 ? 'var(--warning)' : 'var(--danger)');
                    const speedStr = node.AvgBandwidth > 0 ? (node.AvgBandwidth > 1000 ? (node.AvgBandwidth/1024).toFixed(2) + ' MB/s' : node.AvgBandwidth.toFixed(1) + ' KB/s') : '-';
                    
                    let consumedStr = '-';
                    if (node.TotalConsumedBytes > 0) {
                        const mb = node.TotalConsumedBytes / (1024 * 1024);
                        if (mb >= 1024) {
                            consumedStr = (mb / 1024).toFixed(2) + ' GB';
                        } else {
                            consumedStr = mb.toFixed(2) + ' MB';
                        }
                    }
                    
                    tr.innerHTML = '<td>#' + (index + 1) + '</td>' +
                        '<td style="font-weight: 500;">' + node.Name + '</td>' +
                        '<td><span class="score-badge">' + node.Score + '</span></td>' +
                        '<td style="color: ' + successColor + ';">' + (node.SuccessRate * 100).toFixed(1) + '%</td>' +
                        '<td>' + node.AvgDelay.toFixed(1) + '</td>' +
                        '<td>' + speedStr + '</td>' +
                        '<td>' + consumedStr + '</td>';
                    tbody.appendChild(tr);
                });
            } catch (err) {
                console.error('Error fetching stats:', err);
            }
        }
        
        fetchStats();
        setInterval(fetchStats, 10000);
    </script>
</body>
</html>`
	fmt.Fprint(w, html)
}
