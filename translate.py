import os

# 1. App.tsx
with open("frontend/src/App.tsx", "r", encoding="utf-8") as f:
    app = f.read()

app = app.replace('<span className="hig-body">Dashboard</span>', '<span className="hig-body">系統總覽</span>')
app = app.replace('<span className="hig-body">Logs</span>', '<span className="hig-body">系統日誌</span>')
app = app.replace('<span className="hig-body">Appearance</span>', '<span className="hig-body">外觀設定</span>')
app = app.replace('<div className="hig-headline">Node Rover</div>', '<div className="hig-headline">節點巡測</div>')
app = app.replace('<div className="hig-title-2" style={{marginBottom: \'24px\'}}>Groups</div>', '<div className="hig-title-2" style={{marginBottom: \'24px\'}}>節點群組管理</div>')
app = app.replace('<div className="hig-title-2" style={{marginBottom: \'24px\', marginTop: \'48px\'}}>Node Rankings</div>', '<div className="hig-title-2" style={{marginBottom: \'24px\', marginTop: \'48px\'}}>節點即時排行榜</div>')
app = app.replace('<div className="hig-title-2" style={{marginBottom: \'24px\'}}>System Logs</div>', '<div className="hig-title-2" style={{marginBottom: \'24px\'}}>系統即時日誌</div>')

with open("frontend/src/App.tsx", "w", encoding="utf-8") as f:
    f.write(app)

# 2. Dashboard.tsx
with open("frontend/src/components/Dashboard.tsx", "r", encoding="utf-8") as f:
    dash = f.read()

dash = dash.replace('<div className="hig-title-1" style={{marginBottom:\'8px\'}}>Overview</div>', '<div className="hig-title-1" style={{marginBottom:\'8px\'}}>控制中心</div>')
dash = dash.replace("{status.is_paused ? 'Paused' : 'Active'}", "{status.is_paused ? '已暫停' : '監控中'}")
dash = dash.replace("5-minute automated test cycle", "每 5 分鐘自動巡測")
dash = dash.replace("{status.is_running ? 'Testing...' : 'Test Now'}", "{status.is_running ? '測試中...' : '立即測試'}")
dash = dash.replace("{status.is_paused ? 'Resume' : 'Pause'}", "{status.is_paused ? '恢復' : '暫停'}")

with open("frontend/src/components/Dashboard.tsx", "w", encoding="utf-8") as f:
    f.write(dash)

# 3. GroupCard.tsx
with open("frontend/src/components/GroupCard.tsx", "r", encoding="utf-8") as f:
    gc = f.read()

gc = gc.replace('{group.all_count} nodes running', '包含 {group.all_count} 個節點')
gc = gc.replace('title={group.locked ? \'Unlock\' : \'Lock\'}', 'title={group.locked ? \'解除鎖定\' : \'鎖定群組\'}')
gc = gc.replace('MANUAL SELECT', '手動切換節點')
gc = gc.replace('}>Apply</button>', '}>套用</button>')
gc = gc.replace('REGIONS', '地區過濾 (REGIONS)')
gc = gc.replace('SERVICES', '服務過濾 (SERVICES)')

with open("frontend/src/components/GroupCard.tsx", "w", encoding="utf-8") as f:
    f.write(gc)

# 4. NodeRanking.tsx
with open("frontend/src/components/NodeRanking.tsx", "r", encoding="utf-8") as f:
    nr = f.read()

nr = nr.replace('No node data available yet.', '尚無節點連線數據')
nr = nr.replace('<th>Rank</th>', '<th style={{width: \'60px\', textAlign: \'center\'}}>排名</th>')
nr = nr.replace('<th>Node Name</th>', '<th>節點名稱</th>')
nr = nr.replace('<th>Score</th>', '<th style={{textAlign: \'center\'}}>綜合評分</th>')
nr = nr.replace('<th>Delay</th>', '<th style={{textAlign: \'center\'}}>連線延遲</th>')
nr = nr.replace('<th>Jitter</th>', '<th style={{textAlign: \'center\'}}>網路抖動</th>')
nr = nr.replace('<th>Status</th>', '<th>分發狀態</th>')
# Also fix the <th style=...> properly if it had styles:
nr = nr.replace('<th style={{width: \'60px\', textAlign: \'center\'}}>Rank</th>', '<th style={{width: \'60px\', textAlign: \'center\'}}>排名</th>')
nr = nr.replace('<th style={{textAlign: \'center\'}}>Score</th>', '<th style={{textAlign: \'center\'}}>綜合評分</th>')
nr = nr.replace('<th style={{textAlign: \'center\'}}>Delay</th>', '<th style={{textAlign: \'center\'}}>連線延遲</th>')
nr = nr.replace('<th style={{textAlign: \'center\'}}>Jitter</th>', '<th style={{textAlign: \'center\'}}>網路抖動</th>')

nr = nr.replace('>Unused<', '>閒置中<')

with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f:
    f.write(nr)

print("Translation completed.")
