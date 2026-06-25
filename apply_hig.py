import os

css = """
:root {
    /* Apple HIG - Dark Mode (Default) */
    --hig-bg-primary: #000000;
    --hig-bg-secondary: #1c1c1e;
    --hig-bg-tertiary: #2c2c2e;
    
    --hig-text-primary: #ffffff;
    --hig-text-secondary: rgba(235, 235, 245, 0.6);
    --hig-text-tertiary: rgba(235, 235, 245, 0.3);
    
    --hig-system-blue: #0a84ff;
    --hig-system-blue-active: #0066cc;
    --hig-system-green: #30d158;
    --hig-system-red: #ff453a;
    --hig-system-orange: #ff9f0a;
    
    --hig-separator: rgba(84, 84, 88, 0.65);
    
    --hig-fill-primary: rgba(118, 118, 128, 0.24);
    --hig-fill-secondary: rgba(120, 120, 128, 0.16);
    
    --hig-shadow-sm: 0 2px 8px rgba(0,0,0,0.4);
    --hig-shadow-lg: 0 10px 30px rgba(0,0,0,0.5);
    
    --hig-glass-bg: rgba(30, 30, 30, 0.65);
}

:root.light-theme {
    /* Apple HIG - Light Mode */
    --hig-bg-primary: #f2f2f7;
    --hig-bg-secondary: #ffffff;
    --hig-bg-tertiary: #ffffff;
    
    --hig-text-primary: #000000;
    --hig-text-secondary: rgba(60, 60, 67, 0.6);
    --hig-text-tertiary: rgba(60, 60, 67, 0.3);
    
    --hig-system-blue: #007aff;
    --hig-system-blue-active: #005bb5;
    --hig-system-green: #34c759;
    --hig-system-red: #ff3b30;
    --hig-system-orange: #ff9500;
    
    --hig-separator: rgba(60, 60, 67, 0.36);
    
    --hig-fill-primary: rgba(116, 116, 128, 0.08);
    --hig-fill-secondary: rgba(120, 120, 128, 0.04);
    
    --hig-shadow-sm: 0 2px 8px rgba(0,0,0,0.04);
    --hig-shadow-lg: 0 10px 30px rgba(0,0,0,0.08);
    
    --hig-glass-bg: rgba(255, 255, 255, 0.65);
}

* { box-sizing: border-box; -webkit-tap-highlight-color: transparent; }

body {
    background-color: var(--hig-bg-primary);
    color: var(--hig-text-primary);
    font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    margin: 0;
    padding: 0;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
}

/* Typography HIG */
.hig-title-1 { font-size: 34px; font-weight: 700; line-height: 41px; letter-spacing: 0.37px; }
.hig-title-2 { font-size: 28px; font-weight: 700; line-height: 34px; letter-spacing: 0.36px; }
.hig-title-3 { font-size: 20px; font-weight: 600; line-height: 25px; letter-spacing: 0.38px; }
.hig-headline { font-size: 17px; font-weight: 600; line-height: 22px; letter-spacing: -0.41px; }
.hig-body { font-size: 17px; font-weight: 400; line-height: 22px; letter-spacing: -0.41px; }
.hig-subhead { font-size: 15px; font-weight: 400; line-height: 20px; letter-spacing: -0.24px; }
.hig-footnote { font-size: 13px; font-weight: 400; line-height: 18px; letter-spacing: -0.08px; }
.hig-caption-1 { font-size: 12px; font-weight: 500; line-height: 16px; letter-spacing: 0px; }

/* Icons - Mimic SF Symbols Thin */
.material-symbols-outlined {
    vertical-align: middle;
    font-variation-settings: 'FILL' 0, 'wght' 300, 'GRAD' 0, 'opsz' 24;
}
.material-symbols-outlined.fill {
    font-variation-settings: 'FILL' 1;
}

/* ====== APP LAYOUT ====== */
.app-layout {
    display: flex;
    min-height: 100vh;
    width: 100%;
}

/* ====== SIDEBAR (macOS Style) ====== */
.sidebar {
    width: 240px;
    background-color: var(--hig-glass-bg);
    backdrop-filter: blur(40px);
    -webkit-backdrop-filter: blur(40px);
    border-right: 1px solid var(--hig-separator);
    display: flex;
    flex-direction: column;
    padding: 24px 12px;
    position: sticky;
    top: 0;
    height: 100vh;
    z-index: 50;
}
.sidebar-header {
    padding: 0 12px 24px 12px;
    display: flex;
    align-items: center;
    gap: 12px;
}
.brand-icon {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, var(--hig-system-blue), #5bc0de);
    color: white;
    border-radius: 8px;
    display: flex;
    justify-content: center;
    align-items: center;
    box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}
.nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--hig-text-primary);
    width: 100%;
    padding: 8px 12px;
    border-radius: 8px;
    margin-bottom: 4px;
    transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
}
.nav-item .material-symbols-outlined { color: var(--hig-system-blue); font-size: 20px;}
.nav-item:hover { background-color: var(--hig-fill-secondary); }
.nav-item:active { background-color: var(--hig-fill-primary); transform: scale(0.98); }
.nav-item.active { background-color: var(--hig-fill-primary); font-weight: 600; }
.sidebar-spacer { flex: 1; }

@media(max-width: 768px) {
    .app-layout { flex-direction: column; }
    .sidebar { width: 100%; height: auto; flex-direction: row; justify-content: space-around; border-right: none; border-bottom: 1px solid var(--hig-separator); padding: 8px; z-index: 100; position: fixed; bottom: 0; top: auto; background-color: var(--hig-glass-bg); backdrop-filter: blur(20px);}
    .sidebar-header, .sidebar-spacer { display: none; }
    .nav-item { flex-direction: column; padding: 4px; gap: 4px; margin-bottom: 0; font-size: 10px;}
    .nav-item .material-symbols-outlined { font-size: 24px; }
    .main-content { padding: 16px; padding-bottom: 80px; }
}

/* ====== MAIN CONTENT ====== */
.main-content {
    flex: 1;
    padding: 32px 48px;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
}

/* ====== CARDS (HIG Panels) ====== */
.hig-card {
    background-color: var(--hig-bg-secondary);
    border-radius: 16px;
    box-shadow: var(--hig-shadow-sm);
    padding: 24px;
    margin-bottom: 24px;
    border: 1px solid var(--hig-separator);
}

.grid-groups {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
}

/* ====== BUTTONS (HIG iOS/macOS) ====== */
.btn {
    background-color: var(--hig-system-blue);
    color: #ffffff;
    border: none;
    padding: 0 16px;
    height: 36px;
    border-radius: 8px;
    font-family: inherit;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
}
.btn:hover { background-color: var(--hig-system-blue-active); }
.btn:active { transform: scale(0.96); opacity: 0.8; }

.btn.secondary {
    background-color: var(--hig-fill-primary);
    color: var(--hig-text-primary);
}
.btn.secondary:hover { background-color: var(--hig-fill-secondary); }

.icon-btn {
    width: 32px; height: 32px; padding: 0; border-radius: 8px;
    background: transparent; color: var(--hig-system-blue);
    border: none; cursor: pointer; display: flex; align-items: center; justify-content: center;
    transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
}
.icon-btn:hover { background-color: var(--hig-fill-secondary); }
.icon-btn:active { background-color: var(--hig-fill-primary); transform: scale(0.92); }

/* ====== PICKER (Select) HIG ====== */
.hig-picker {
    display: flex;
    align-items: center;
    background-color: var(--hig-fill-primary);
    border-radius: 8px;
    height: 36px;
    padding: 0 12px;
    position: relative;
    cursor: pointer;
    transition: background-color 0.2s;
}
.hig-picker:hover { background-color: var(--hig-fill-secondary); }
.hig-picker select {
    flex: 1;
    background: transparent;
    color: var(--hig-text-primary);
    border: none;
    font-family: inherit;
    font-size: 15px;
    outline: none;
    cursor: pointer;
    appearance: none;
    width: 100%;
}
.hig-picker .material-symbols-outlined { color: var(--hig-text-secondary); pointer-events: none; position: absolute; right: 12px;}

/* ====== CHIPS / TOKENS ====== */
.chip-group { display: flex; flex-wrap: wrap; gap: 8px; }
.hig-chip {
    background-color: transparent;
    color: var(--hig-text-primary);
    border: 1px solid var(--hig-separator);
    border-radius: 16px;
    padding: 0 12px;
    height: 32px;
    font-family: inherit;
    font-size: 14px;
    font-weight: 400;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
    user-select: none;
}
.hig-chip:active { transform: scale(0.96); }
.hig-chip.selected {
    background-color: var(--hig-system-blue);
    color: #fff;
    border-color: var(--hig-system-blue);
    font-weight: 600;
}

/* ====== TABLES ====== */
.hig-table-container {
    background-color: var(--hig-bg-secondary);
    border-radius: 12px;
    border: 1px solid var(--hig-separator);
    overflow: hidden;
}
table { width: 100%; border-collapse: collapse; text-align: left; }
th {
    background-color: var(--hig-bg-tertiary);
    color: var(--hig-text-secondary);
    font-family: inherit;
    font-size: 13px;
    font-weight: 600;
    padding: 12px 16px;
    border-bottom: 1px solid var(--hig-separator);
}
td {
    padding: 16px;
    border-bottom: 1px solid var(--hig-separator);
    font-size: 15px;
    color: var(--hig-text-primary);
    font-family: inherit;
}
tr:last-child td { border-bottom: none; }
tr { transition: background-color 0.2s; }
tr:hover { background-color: var(--hig-fill-secondary); }

/* Badges */
.hig-badge { display: inline-flex; align-items: center; padding: 4px 8px; border-radius: 6px; font-size: 12px; font-weight: 600; gap: 4px; }
.hig-badge.blue { background-color: rgba(10, 132, 255, 0.15); color: var(--hig-system-blue); }
.hig-badge.green { background-color: rgba(48, 209, 88, 0.15); color: var(--hig-system-green); }
.hig-badge.red { background-color: rgba(255, 69, 58, 0.15); color: var(--hig-system-red); }
.hig-badge.orange { background-color: rgba(255, 159, 10, 0.15); color: var(--hig-system-orange); }

/* Console */
.console {
    background-color: #000;
    color: #fff;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    padding: 16px;
    border-radius: 12px;
    height: 600px;
    overflow-y: auto;
    font-size: 12px;
    line-height: 1.6;
    border: 1px solid var(--hig-separator);
}
.log-line { display: flex; margin-bottom: 4px; animation: slideIn 0.3s ease-out; }
@keyframes slideIn { from{transform: translateY(10px); opacity:0;} to{transform: translateY(0); opacity:1;} }
.log-time { color: var(--hig-text-secondary); margin-right: 12px; white-space: nowrap;}
.log-badge { padding: 2px 6px; border-radius: 4px; font-size: 10px; margin-right: 12px; font-weight: 700; }
.log-info .log-badge { background: rgba(10, 132, 255, 0.2); color: #0a84ff; }
.log-success .log-badge { background: rgba(48, 209, 88, 0.2); color: #30d158; }
.log-warning .log-badge { background: rgba(255, 159, 10, 0.2); color: #ff9f0a; }
.log-error .log-badge { background: rgba(255, 69, 58, 0.2); color: #ff453a; }
.log-success .log-msg { color: #30d158; }
.log-warning .log-msg { color: #ff9f0a; }
.log-error .log-msg { color: #ff453a; }

.tab-content { display: none; }
.tab-content.active { display: block; animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: scale(0.98); } to { opacity: 1; transform: scale(1); } }
.spin { animation: spin 2s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }
"""
with open("frontend/src/index.css", "w", encoding="utf-8") as f: f.write(css)

app_tsx = """import { useEffect, useState } from 'react';
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

    const [activeTab, setActiveTab] = useState('home');
    const [isLightTheme, setIsLightTheme] = useState(false);

    useEffect(() => {
        fetchGroups();
        fetchStats();
        fetchStatus();
        
        const savedThemeMode = localStorage.getItem('themeMode');
        if (savedThemeMode === 'light') {
            setIsLightTheme(true);
            document.documentElement.classList.add('light-theme');
        }
    }, []);

    const toggleTheme = () => {
        const newThemeMode = !isLightTheme;
        setIsLightTheme(newThemeMode);
        document.documentElement.classList.toggle('light-theme', newThemeMode);
        localStorage.setItem('themeMode', newThemeMode ? 'light' : 'dark');
    };

    return (
        <div className="app-layout">
            <aside className="sidebar">
                <div className="sidebar-header">
                    <div className="brand-icon">
                        <span className="material-symbols-outlined" style={{fontSize: '20px'}}>rocket_launch</span>
                    </div>
                    <div className="hig-headline">Node Rover</div>
                </div>
                
                <button className={`nav-item ${activeTab === 'home' ? 'active' : ''}`} onClick={() => setActiveTab('home')}>
                    <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'home' ? "'FILL' 1" : "'FILL' 0"}}>home</span>
                    <span className="hig-body">Dashboard</span>
                </button>
                
                <button className={`nav-item ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
                    <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'logs' ? "'FILL' 1" : "'FILL' 0"}}>terminal</span>
                    <span className="hig-body">Logs</span>
                </button>

                <div className="sidebar-spacer"></div>

                <button className="nav-item" onClick={toggleTheme}>
                    <span className="material-symbols-outlined">{isLightTheme ? 'dark_mode' : 'light_mode'}</span>
                    <span className="hig-body">Appearance</span>
                </button>
            </aside>

            <main className="main-content">
                <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} />

                <div className={`tab-content ${activeTab === 'home' ? 'active' : ''}`}>
                    <div className="hig-title-2" style={{marginBottom: '24px'}}>Groups</div>
                    <div className="grid-groups">
                        {groups.map(g => (
                            <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                        ))}
                    </div>
                    
                    <div className="hig-title-2" style={{marginBottom: '24px', marginTop: '48px'}}>Node Rankings</div>
                    <NodeRanking stats={stats} />
                </div>

                <div className={`tab-content ${activeTab === 'logs' ? 'active' : ''}`}>
                    <div className="hig-title-2" style={{marginBottom: '24px'}}>System Logs</div>
                    <div className="hig-card" style={{padding: '0', border: 'none'}}>
                        <div className="console">
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
            </main>
        </div>
    );
}

export default App;
"""
with open("frontend/src/App.tsx", "w", encoding="utf-8") as f: f.write(app_tsx)

dash_tsx = """import type { Status } from '../hooks/useApi';

export default function Dashboard({ status, triggerTest, togglePause }: { status: Status, triggerTest: () => void, togglePause: () => void }) {
    return (
        <div className="hig-card" style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '24px'}}>
            <div>
                <div className="hig-title-1" style={{marginBottom:'8px'}}>Overview</div>
                <div style={{display: 'flex', gap: '12px', alignItems: 'center'}}>
                    <span className={`hig-badge ${status.is_paused ? 'orange' : 'green'}`}>
                        {status.is_paused ? 'Paused' : 'Active'}
                    </span>
                    <span className="hig-footnote" style={{color: 'var(--hig-text-secondary)'}}>
                        5-minute automated test cycle
                    </span>
                </div>
            </div>
            
            <div style={{display:'flex', gap:'12px'}}>
                <button className="btn" onClick={triggerTest} disabled={status.is_running}>
                    <span className={`material-symbols-outlined ${status.is_running ? 'spin' : ''}`}>
                        {status.is_running ? 'refresh' : 'speed'}
                    </span>
                    {status.is_running ? 'Testing...' : 'Test Now'}
                </button>
                <button className={`btn secondary`} onClick={togglePause}>
                    <span className={`material-symbols-outlined ${status.is_paused ? 'fill' : ''}`}>{status.is_paused ? 'play_arrow' : 'pause'}</span>
                    {status.is_paused ? 'Resume' : 'Pause'}
                </button>
            </div>
        </div>
    );
}
"""
with open("frontend/src/components/Dashboard.tsx", "w", encoding="utf-8") as f: f.write(dash_tsx)

gc_tsx = """const REGION_PRESETS: Record<string, string> = {
    'US': 'US|United States|us|美國|美国',
    'HK': 'HK|Hong Kong|香港',
    'TW': 'TW|Taiwan|台灣|台湾|臺',
    'JP': 'JP|Japan|日本',
    'SG': 'SG|Singapore|新加坡|狮城',
    'UK': 'UK|United Kingdom|英國|英国'
};

export default function GroupCard({ group, manualSwitch, toggleGroupLock, saveFilter }: any) {
    const rx = group.filter?.keyword_regex || "";
    const isChatGPT = group.filter?.check_chatgpt || false;
    const isGemini = group.filter?.check_gemini || false;
    const isAntigravity = group.filter?.check_antigravity || false;

    const handleRegionChange = (val: string, currentSelected: boolean) => {
        let regexes: string[] = [];
        ['US', 'HK', 'TW', 'JP', 'SG', 'UK'].forEach(r => {
            let isR = rx.includes(REGION_PRESETS[r]);
            if (r === val) isR = !currentSelected;
            if (isR) regexes.push(REGION_PRESETS[r]);
        });
        saveFilter(group.name, { keyword_regex: regexes.join('|'), check_chatgpt: isChatGPT, check_gemini: isGemini, check_antigravity: isAntigravity });
    };

    const handleServiceChange = (service: string, currentSelected: boolean) => {
        saveFilter(group.name, { keyword_regex: rx, check_chatgpt: service === 'chatgpt' ? !currentSelected : isChatGPT, check_gemini: service === 'gemini' ? !currentSelected : isGemini, check_antigravity: service === 'antigravity' ? !currentSelected : isAntigravity });
    };

    return (
        <div className="hig-card" style={{display: 'flex', flexDirection: 'column', gap: '20px'}}>
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start'}}>
                <div>
                    <div className="hig-title-3" style={{marginBottom: '2px'}}>{group.name}</div>
                    <div className="hig-footnote" style={{color: 'var(--hig-text-secondary)'}}>
                        {group.all_count} nodes running
                    </div>
                </div>
                <button className="icon-btn" onClick={() => toggleGroupLock(group.name, !group.locked)} title={group.locked ? 'Unlock' : 'Lock'}>
                    <span className={`material-symbols-outlined ${group.locked ? 'fill' : ''}`} style={{color: group.locked ? 'var(--hig-system-blue)' : 'var(--hig-text-secondary)'}}>{group.locked ? 'lock' : 'lock_open_right'}</span>
                </button>
            </div>

            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>MANUAL SELECT</div>
                <div style={{display: 'flex', gap: '8px'}}>
                    <div className="hig-picker" style={{flex: 1}}>
                        <select id={`select-${group.name}`} defaultValue={group.now}>
                            {group.all_nodes && group.all_nodes.map((n: string) => <option key={n} value={n} >{n}</option>)}
                        </select>
                        <span className="material-symbols-outlined">expand_more</span>
                    </div>
                    <button className="btn secondary" style={{height: '36px'}} onClick={() => {
                        const sel = document.getElementById(`select-${group.name}`) as HTMLSelectElement;
                        manualSwitch(group.name, sel.value);
                    }}>Apply</button>
                </div>
            </div>
            
            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>REGIONS</div>
                <div className="chip-group">
                    {['US', 'HK', 'TW', 'JP', 'SG', 'UK'].map(r => {
                        const isSelected = rx.includes(r + '|');
                        return (
                            <div key={r} className={`hig-chip ${isSelected ? 'selected' : ''}`} onClick={() => handleRegionChange(r, isSelected)}>
                                {r}
                            </div>
                        );
                    })}
                </div>
            </div>

            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>SERVICES</div>
                <div className="chip-group">
                    <div className={`hig-chip ${isChatGPT ? 'selected' : ''}`} onClick={() => handleServiceChange('chatgpt', isChatGPT)}>
                        ChatGPT
                    </div>
                    <div className={`hig-chip ${isGemini ? 'selected' : ''}`} onClick={() => handleServiceChange('gemini', isGemini)}>
                        Gemini
                    </div>
                    <div className={`hig-chip ${isAntigravity ? 'selected' : ''}`} onClick={() => handleServiceChange('antigravity', isAntigravity)}>
                        Antigravity
                    </div>
                </div>
            </div>
        </div>
    );
}
"""
with open("frontend/src/components/GroupCard.tsx", "w", encoding="utf-8") as f: f.write(gc_tsx)

nr_tsx = """import { useState, Fragment } from 'react';
import type { NodeStat } from '../hooks/useApi';

export default function NodeRanking({ stats }: any) {
    if (!stats || stats.length === 0) {
        return (
            <div className="hig-card" style={{textAlign:'center', padding:'60px 20px', color:'var(--hig-text-secondary)'}}>
                <span className="material-symbols-outlined" style={{fontSize:'48px', marginBottom:'16px', opacity: 0.5}}>hourglass_empty</span>
                <div className="hig-headline">No node data available yet.</div>
            </div>
        );
    }

    return (
        <div className="hig-table-container">
            <table>
                <thead>
                    <tr>
                        <th style={{width: '60px', textAlign: 'center'}}>Rank</th>
                        <th>Node Name</th>
                        <th style={{textAlign: 'center'}}>Score</th>
                        <th style={{textAlign: 'center'}}>Delay</th>
                        <th style={{textAlign: 'center'}}>Jitter</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {stats.map((s: any, idx: number) => {
                        const isDead = s.is_dead || false;
                        const scoreStr = isDead ? "FAIL" : s.Score;
                        return (
                            <tr key={s.Name}>
                                <td style={{textAlign: 'center', fontWeight: '600', color: idx < 3 ? 'var(--hig-system-blue)' : 'inherit'}}>
                                    {idx + 1}
                                </td>
                                <td style={{fontWeight: '500', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{s.Name}</td>
                                <td style={{textAlign: 'center'}}>
                                    <div className={`hig-badge ${isDead ? 'red' : 'blue'}`}>
                                        {scoreStr}
                                    </div>
                                </td>
                                <td style={{textAlign: 'center', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{!isDead ? `${s.AvgDelay} ms` : '-'}</td>
                                <td style={{textAlign: 'center', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{!isDead ? `${s.Jitter} ms` : '-'}</td>
                                <td>
                                    {s.highest_in_groups && s.highest_in_groups.length > 0 ? (
                                        <div style={{display:'flex', gap:'4px', flexWrap:'wrap'}}>
                                            {s.highest_in_groups.map((g: string) => <span key={g} className="hig-badge green">{g}</span>)}
                                        </div>
                                    ) : (
                                        <span style={{color: 'var(--hig-text-secondary)', fontSize: '13px'}}>Unused</span>
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}
"""
with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f: f.write(nr_tsx)

print("HIG styling applied.")
