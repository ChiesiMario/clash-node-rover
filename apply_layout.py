import re

# 1. Update index.css
with open('frontend/src/index.css', 'r', encoding='utf-8') as f:
    css = f.read()

desktop_layout_css = """
.desktop-layout {
    display: grid;
    grid-template-columns: 65fr 35fr;
    gap: 24px;
    align-items: start;
}

@media (max-width: 1024px) {
    .desktop-layout {
        grid-template-columns: 1fr;
    }
}
"""

if '.desktop-layout' not in css:
    css += desktop_layout_css

with open('frontend/src/index.css', 'w', encoding='utf-8') as f:
    f.write(css)


# 2. Update App.tsx
app_tsx_content = """import { useEffect, useState } from 'react';
import { useApi } from './hooks/useApi';
import { useWebSocket } from './hooks/useWebSocket';
import Dashboard from './components/Dashboard';
import GroupCard from './components/GroupCard';
import NodeRanking from './components/NodeRanking';

function App() {
    const { stats, status, groups, fetchStats, fetchStatus, fetchGroups, triggerTest, togglePause, manualSwitch, toggleGroupLock, saveFilter } = useApi();
    const { logs } = useWebSocket(() => {
        fetchGroups();
        fetchStats();
        fetchStatus();
    });

    const [activeTab, setActiveTab] = useState('groups');
    const [isLightTheme, setIsLightTheme] = useState(false);

    useEffect(() => {
        fetchGroups();
        fetchStats();
        fetchStatus();
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme === 'light') {
            setIsLightTheme(true);
            document.documentElement.classList.add('light-theme');
        }
    }, []);

    const toggleTheme = () => {
        const newTheme = !isLightTheme;
        setIsLightTheme(newTheme);
        document.documentElement.classList.toggle('light-theme', newTheme);
        localStorage.setItem('theme', newTheme ? 'light' : 'dark');
    };

    return (
        <div className="container">
            <div className="top-app-bar" style={{marginBottom: '24px', borderRadius: '16px'}}>
                <div className="app-title">
                    <span className="material-symbols-outlined" style={{color: "var(--md-sys-color-primary)", fontSize: "28px"}}>rocket_launch</span>
                    Clash Node Rover
                </div>
                <button className="icon-btn" onClick={toggleTheme} title="切換深色/淺色主題">
                    <span className="material-symbols-outlined" id="themeIcon">
                        {isLightTheme ? 'dark_mode' : 'light_mode'}
                    </span>
                </button>
            </div>

            <div className="desktop-layout">
                {/* Left Column: Dashboard + Ranking */}
                <div className="layout-left">
                    <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} />
                    <NodeRanking stats={stats} />
                </div>

                {/* Right Column: Tabs + Groups/Logs */}
                <div className="layout-right">
                    <div className="segmented-button" style={{display: 'flex', width: '100%', marginBottom: '16px'}}>
                        <button id="btn-groups" style={{flex: 1, justifyContent: 'center'}} className={`seg-btn ${activeTab === 'groups' ? 'active' : ''}`} onClick={() => setActiveTab('groups')}>
                            <span className="material-symbols-outlined" style={{fontSize:'18px'}}>grid_view</span> 群組監控
                        </button>
                        <button id="btn-logs" style={{flex: 1, justifyContent: 'center'}} className={`seg-btn ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
                            <span className="material-symbols-outlined" style={{fontSize:'18px'}}>terminal</span> 系統日誌
                        </button>
                    </div>

                    <div id="tab-groups" className={`tab-content ${activeTab === 'groups' ? 'active' : ''}`} style={{display: activeTab === 'groups' ? 'block' : 'none'}}>
                        <div className="grid" id="groupsGrid" style={{display: 'flex', flexDirection: 'column', gap: '16px'}}>
                            {groups.map(g => (
                                <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                            ))}
                        </div>
                    </div>

                    <div id="tab-logs" className={`tab-content ${activeTab === 'logs' ? 'active' : ''}`} style={{display: activeTab === 'logs' ? 'block' : 'none'}}>
                        <div className="card">
                            <div style={{fontWeight:500, marginBottom:'16px'}}>即時系統日誌</div>
                            <div className="console-wrapper" style={{maxHeight: '600px'}}>
                                <div id="terminalBody" className="console" style={{height: '100%', overflowY: 'auto'}}>
                                    {logs.map((log, i) => (
                                        <div key={i} className={`log-line log-${log.level === 'success' ? 'success' : log.level === 'warning' ? 'warning' : log.level === 'error' ? 'error' : 'info'}`}>
                                            <div className="log-time">[{log.time}]</div>
                                            <div className="log-badge">{log.level === 'success' ? 'OK' : log.level === 'warning' ? 'WARN' : log.level === 'error' ? 'FAIL' : 'INFO'}</div>
                                            <div className="log-msg">{log.message.replace(/^[💡✅⚠️❌] /, '')}</div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default App;
"""

with open('frontend/src/App.tsx', 'w', encoding='utf-8') as f:
    f.write(app_tsx_content)


# Read task.md and update
with open(r'C:\Users\Noah\.gemini\antigravity-ide\brain\5c682920-2755-40ef-a795-f50e8cacbd6d\task.md', 'r', encoding='utf-8') as f:
    task = f.read()

task = task.replace('- `[ ]` 1. Update CSS', '- `[x]` 1. Update CSS')
task = task.replace('- `[ ]` Add `.desktop-layout` grid styles', '- `[x]` Add `.desktop-layout` grid styles')
task = task.replace('- `[ ]` 2. Update `App.tsx`', '- `[x]` 2. Update `App.tsx`')
task = task.replace('- `[ ]` Restructure left column (Dashboard + NodeRanking)', '- `[x]` Restructure left column (Dashboard + NodeRanking)')
task = task.replace('- `[ ]` Restructure right column (Tabs + GroupCard/Logs)', '- `[x]` Restructure right column (Tabs + GroupCard/Logs)')

with open(r'C:\Users\Noah\.gemini\antigravity-ide\brain\5c682920-2755-40ef-a795-f50e8cacbd6d\task.md', 'w', encoding='utf-8') as f:
    f.write(task)
