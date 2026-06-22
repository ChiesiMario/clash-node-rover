package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func StartWebServer(db *DB, rover *Rover, port int) {
	http.HandleFunc("/", handleIndex)

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		scores, err := db.GetScores(7)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		list := make([]NodeScore, 0)
		for _, sc := range scores {
			list = append(list, sc)
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

		// 嘗試非阻塞寫入 channel
		select {
		case rover.ManualTrigger <- struct{}{}:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
    <title>Clash Node Rover</title>
    <style>
        :root {
            --bg-dark: #0f172a;
            --panel-bg: rgba(30, 41, 59, 0.7);
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --primary: #3b82f6;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
        }
        body {
            background-color: var(--bg-dark);
            color: var(--text-main);
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .glass-panel {
            background: var(--panel-bg);
            backdrop-filter: blur(10px);
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        }
        table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        th { text-align: left; padding: 12px; border-bottom: 2px solid rgba(255,255,255,0.1); }
        td { padding: 12px; border-bottom: 1px solid rgba(255,255,255,0.05); }
        tr:hover { background: rgba(255,255,255,0.05); cursor: pointer; }
        .score-badge {
            background: rgba(59, 130, 246, 0.2);
            color: var(--primary);
            padding: 4px 12px;
            border-radius: 999px;
            font-weight: bold;
        }
        @keyframes spin { 100% { transform: rotate(360deg); } }
        .chart-container { width: 100%; height: 300px; padding: 15px; box-sizing: border-box; }
        .expanded-row { background: rgba(0,0,0,0.2) !important; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <h1>🤖 NODE ROVER</h1>
        <div class="glass-panel">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
                <div>
                    <h2>質量分數排行榜</h2>
                    <div style="color: var(--text-muted);">點擊節點可查看 24 小時歷史延遲圖表。</div>
                </div>
                <button id="triggerBtn" onclick="triggerTest()" style="background: var(--primary); color: white; border: none; padding: 10px 20px; border-radius: 8px; cursor: pointer; font-weight: bold; display: flex; align-items: center; gap: 8px;">
                    <span id="triggerIcon">🚀</span> <span id="triggerText">立即全局測速</span>
                </button>
            </div>
            <table id="statsTable">
                <thead>
                    <tr>
                        <th>排名</th>
                        <th>節點名稱</th>
                        <th>質量分數</th>
                        <th>成功率</th>
                        <th>平均延遲 (MS)</th>
                        <th>平均下載速度</th>
                        <th>累計流量</th>
                    </tr>
                </thead>
                <tbody id="tbody">
                    <tr><td colspan="7" style="text-align:center;color:var(--text-muted);">載入中...</td></tr>
                </tbody>
            </table>
        </div>
    </div>
    <script>
        let chartInstances = {};

        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                const tbody = document.getElementById('tbody');
                
                if (!data || data.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:var(--text-muted);">目前還沒有足夠的測速資料...</td></tr>';
                    return;
                }

                if (tbody.children.length === 1 && tbody.children[0].textContent.includes('載入中')) {
                    tbody.innerHTML = '';
                }

                data.forEach((node, index) => {
                    let tr = document.getElementById('row-' + index);
                    if (!tr) {
                        tr = document.createElement('tr');
                        tr.id = 'row-' + index;
                        tr.onclick = () => toggleChart(node.Name, index);
                        tbody.appendChild(tr);
                    }
                    
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
            chartRow.innerHTML = '<td colspan="7"><div class="chart-container"><canvas id="canvas-' + index + '"></canvas></div></td>';
            tr.parentNode.insertBefore(chartRow, tr.nextSibling);

            try {
                const res = await fetch('/api/history?node=' + encodeURIComponent(nodeName));
                const history = await res.json();
                
                if (!history || history.length === 0) {
                    chartRow.innerHTML = '<td colspan="7" style="text-align:center; padding: 20px;">無歷史資料</td>';
                    return;
                }

                const ctx = document.getElementById('canvas-' + index).getContext('2d');
                
                const labels = history.map(h => {
                    const d = new Date(h.Timestamp * 1000);
                    return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0');
                });
                const data = history.map(h => h.Delay);

                chartInstances[index] = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: labels,
                        datasets: [{
                            label: '延遲 (ms)',
                            data: data,
                            borderColor: '#3b82f6',
                            backgroundColor: 'rgba(59, 130, 246, 0.1)',
                            borderWidth: 2,
                            fill: true,
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } },
                        scales: {
                            y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#94a3b8' } },
                            x: { grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#94a3b8', maxTicksLimit: 12 } }
                        }
                    }
                });
            } catch (err) {
                chartRow.innerHTML = '<td colspan="7" style="text-align:center;">載入失敗</td>';
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
                    btn.style.opacity = '0.7';
                    btn.style.cursor = 'not-allowed';
                    icon.innerHTML = '🔄';
                    icon.style.animation = 'spin 2s linear infinite';
                    text.innerText = '測速執行中...';
                    btn.disabled = true;
                } else {
                    btn.style.opacity = '1';
                    btn.style.cursor = 'pointer';
                    icon.innerHTML = '🚀';
                    icon.style.animation = 'none';
                    text.innerText = '立即全局測速';
                    btn.disabled = false;
                }
            } catch (err) {}
        }

        async function triggerTest() {
            const btn = document.getElementById('triggerBtn');
            if (btn.disabled) return;
            try {
                await fetch('/api/trigger', { method: 'POST' });
                checkStatus();
            } catch (err) {}
        }

        fetchStats();
        checkStatus();
        setInterval(fetchStats, 10000);
        setInterval(checkStatus, 2000);
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
