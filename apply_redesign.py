import os
import re

# 1. Update index.html
html_path = "frontend/index.html"
with open(html_path, "r", encoding="utf-8") as f:
    html = f.read()

if "fonts.googleapis.com" not in html:
    html = html.replace('</head>', '  <link rel="preconnect" href="https://fonts.googleapis.com">\n  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>\n  <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700&family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">\n</head>')
    with open(html_path, "w", encoding="utf-8") as f:
        f.write(html)

# 2. Write index.css
css_content = """
:root {
    /* MD3 Color Tokens - Dark Mode (Default) */
    --md-sys-color-primary: #a8c7fa;
    --md-sys-color-on-primary: #062e6f;
    --md-sys-color-primary-container: #0842a0;
    --md-sys-color-on-primary-container: #d3e3fd;

    --md-sys-color-secondary: #c2c7cf;
    --md-sys-color-on-secondary: #2b313a;
    --md-sys-color-secondary-container: #414751;
    --md-sys-color-on-secondary-container: #dee3eb;

    --md-sys-color-tertiary: #c4c3ea;
    --md-sys-color-on-tertiary: #2d2d4d;
    --md-sys-color-tertiary-container: #434465;
    --md-sys-color-on-tertiary-container: #e1dfff;

    --md-sys-color-error: #ffb4ab;
    --md-sys-color-error-container: #93000a;
    --md-sys-color-on-error: #690005;
    --md-sys-color-on-error-container: #ffdad6;

    --md-sys-color-success: #6dd58c;
    --md-sys-color-success-container: #00521c;
    --md-sys-color-on-success-container: #8bf2a6;

    --md-sys-color-warning: #eaddff;
    --md-sys-color-warning-container: #5b467a;
    --md-sys-color-on-warning-container: #f5eeff;

    --md-sys-color-background: #0f1014;
    --md-sys-color-on-background: #e2e2e9;

    --md-sys-color-surface: #0f1014;
    --md-sys-color-on-surface: #e2e2e9;
    --md-sys-color-surface-variant: #44474e;
    --md-sys-color-on-surface-variant: #c4c6d0;

    --md-sys-color-surface-container-lowest: #0c0d11;
    --md-sys-color-surface-container-low: #18191e;
    --md-sys-color-surface-container: #1c1d22;
    --md-sys-color-surface-container-high: #27282d;
    --md-sys-color-surface-container-highest: #323338;

    --md-sys-color-outline: #8e9099;
    --md-sys-color-outline-variant: #44474e;

    /* Shadow & Glass */
    --md-sys-elevation-1: 0px 1px 3px 1px rgba(0,0,0,0.4), 0px 1px 2px 0px rgba(0,0,0,0.2);
    --md-sys-elevation-2: 0px 2px 6px 2px rgba(0,0,0,0.4), 0px 1px 2px 0px rgba(0,0,0,0.2);
    --md-sys-elevation-3: 0px 4px 8px 3px rgba(0,0,0,0.4), 0px 1px 3px 0px rgba(0,0,0,0.2);
    --glass-bg: rgba(28, 29, 34, 0.6);
    --glass-border: rgba(255, 255, 255, 0.08);
}

:root.light-theme {
    /* MD3 Color Tokens - Light Mode */
    --md-sys-color-primary: #0842a0;
    --md-sys-color-on-primary: #ffffff;
    --md-sys-color-primary-container: #d3e3fd;
    --md-sys-color-on-primary-container: #001c3b;

    --md-sys-color-secondary: #535f70;
    --md-sys-color-on-secondary: #ffffff;
    --md-sys-color-secondary-container: #d7e3f8;
    --md-sys-color-on-secondary-container: #101c2b;

    --md-sys-color-tertiary: #5a5c7e;
    --md-sys-color-on-tertiary: #ffffff;
    --md-sys-color-tertiary-container: #e1dfff;
    --md-sys-color-on-tertiary-container: #171937;

    --md-sys-color-error: #ba1a1a;
    --md-sys-color-error-container: #ffdad6;
    --md-sys-color-on-error: #ffffff;
    --md-sys-color-on-error-container: #410002;

    --md-sys-color-success: #146c2e;
    --md-sys-color-success-container: #a0f9b9;
    --md-sys-color-on-success-container: #002106;

    --md-sys-color-warning: #6d5e8f;
    --md-sys-color-warning-container: #f5eeff;
    --md-sys-color-on-warning-container: #271a47;

    --md-sys-color-background: #fdfbff;
    --md-sys-color-on-background: #1a1c1e;

    --md-sys-color-surface: #fdfbff;
    --md-sys-color-on-surface: #1a1c1e;
    --md-sys-color-surface-variant: #dfe2eb;
    --md-sys-color-on-surface-variant: #44474e;

    --md-sys-color-surface-container-lowest: #ffffff;
    --md-sys-color-surface-container-low: #f4f3f7;
    --md-sys-color-surface-container: #eeedf1;
    --md-sys-color-surface-container-high: #e8e7eb;
    --md-sys-color-surface-container-highest: #e2e2e5;

    --md-sys-color-outline: #74777f;
    --md-sys-color-outline-variant: #c4c6d0;

    --md-sys-elevation-1: 0px 1px 3px 1px rgba(0,0,0,0.1), 0px 1px 2px 0px rgba(0,0,0,0.06);
    --md-sys-elevation-2: 0px 2px 6px 2px rgba(0,0,0,0.1), 0px 1px 2px 0px rgba(0,0,0,0.06);
    --md-sys-elevation-3: 0px 4px 8px 3px rgba(0,0,0,0.1), 0px 1px 3px 0px rgba(0,0,0,0.06);
    --glass-bg: rgba(238, 237, 241, 0.7);
    --glass-border: rgba(0, 0, 0, 0.05);
}

* { box-sizing: border-box; }

body {
    background-color: var(--md-sys-color-background);
    color: var(--md-sys-color-on-background);
    font-family: 'Roboto', sans-serif;
    margin: 0;
    padding: 0;
    -webkit-font-smoothing: antialiased;
}

/* Typography */
h1, h2, h3, h4, h5, h6, .brand, .hero-title {
    font-family: 'Outfit', sans-serif;
    margin: 0;
}
.md3-headline-large { font-family: 'Outfit', sans-serif; font-size: 32px; font-weight: 600; line-height: 40px; }
.md3-title-large { font-family: 'Outfit', sans-serif; font-size: 22px; font-weight: 500; line-height: 28px; }
.md3-title-medium { font-family: 'Outfit', sans-serif; font-size: 16px; font-weight: 500; line-height: 24px; letter-spacing: 0.15px; }
.md3-body-large { font-size: 16px; font-weight: 400; line-height: 24px; letter-spacing: 0.5px; }
.md3-body-medium { font-size: 14px; font-weight: 400; line-height: 20px; letter-spacing: 0.25px; }
.md3-label-large { font-size: 14px; font-weight: 500; line-height: 20px; letter-spacing: 0.1px; }

/* Icons */
.material-symbols-outlined {
    vertical-align: middle;
    font-variation-settings: 'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24;
}

/* ====== APP LAYOUT ====== */
.app-layout {
    display: flex;
    min-height: 100vh;
    width: 100%;
}

/* ====== NAVIGATION RAIL ====== */
.nav-rail {
    width: 80px;
    background-color: var(--md-sys-color-surface-container);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 0;
    gap: 32px;
    border-right: 1px solid var(--md-sys-color-outline-variant);
    position: sticky;
    top: 0;
    height: 100vh;
    z-index: 50;
}

.nav-rail-top {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
}

.brand-icon {
    width: 48px;
    height: 48px;
    background-color: var(--md-sys-color-primary-container);
    color: var(--md-sys-color-on-primary-container);
    border-radius: 16px;
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 24px;
    box-shadow: var(--md-sys-elevation-1);
}

.nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--md-sys-color-on-surface-variant);
    width: 100%;
    padding: 12px 0;
    position: relative;
    transition: color 0.2s;
}

.nav-item .icon-container {
    width: 56px;
    height: 32px;
    border-radius: 16px;
    display: flex;
    justify-content: center;
    align-items: center;
    transition: background-color 0.2s, color 0.2s;
    position: relative;
    overflow: hidden;
}

.nav-item .icon-container::after {
    content: '';
    position: absolute;
    inset: 0;
    background-color: var(--md-sys-color-on-surface);
    opacity: 0;
    transition: opacity 0.2s;
}
.nav-item:hover .icon-container::after { opacity: 0.08; }
.nav-item:active .icon-container::after { opacity: 0.12; }

.nav-item.active {
    color: var(--md-sys-color-on-surface);
}
.nav-item.active .icon-container {
    background-color: var(--md-sys-color-secondary-container);
    color: var(--md-sys-color-on-secondary-container);
}
.nav-label { font-size: 12px; font-weight: 500; font-family: 'Roboto', sans-serif;}

.nav-rail-spacer { flex: 1; }

/* ====== MAIN CONTENT ====== */
.main-content {
    flex: 1;
    padding: 32px 48px;
    max-width: 1600px;
    margin: 0 auto;
    width: 100%;
}
@media(max-width: 768px) {
    .app-layout { flex-direction: column; }
    .nav-rail { width: 100%; height: 80px; flex-direction: row; justify-content: space-around; border-right: none; border-bottom: 1px solid var(--md-sys-color-outline-variant); padding: 0 16px; }
    .nav-rail-top { display: none; }
    .nav-rail-spacer { display: none; }
    .main-content { padding: 16px; }
    .nav-item { padding: 8px; }
}

/* ====== HERO SECTION (Dashboard) ====== */
.hero-card {
    background: linear-gradient(135deg, var(--md-sys-color-primary-container) 0%, var(--md-sys-color-surface-container) 100%);
    border-radius: 28px;
    padding: 32px 40px;
    margin-bottom: 40px;
    position: relative;
    overflow: hidden;
    box-shadow: var(--md-sys-elevation-2);
    border: 1px solid var(--glass-border);
}
.hero-card::before {
    content: '';
    position: absolute;
    top: -50px; left: -50px; width: 200px; height: 200px;
    background: var(--md-sys-color-tertiary-container);
    filter: blur(80px); opacity: 0.4; border-radius: 50%;
}
.hero-card-content {
    position: relative;
    z-index: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 24px;
}
.hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.5px;
}
.hero-badge.running { background-color: var(--md-sys-color-success-container); color: var(--md-sys-color-on-success-container); }
.hero-badge.paused { background-color: var(--md-sys-color-warning-container); color: var(--md-sys-color-on-warning-container); }

/* ====== CARDS ====== */
.card {
    background-color: var(--md-sys-color-surface-container-low);
    border-radius: 24px;
    box-shadow: var(--md-sys-elevation-1);
    padding: 32px;
    margin-bottom: 32px;
    border: 1px solid var(--glass-border);
}

.group-card {
    background-color: var(--md-sys-color-surface-container-highest);
    border-radius: 24px;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    transition: transform 0.2s, box-shadow 0.2s;
}
.group-card:hover {
    transform: translateY(-2px);
    box-shadow: var(--md-sys-elevation-1);
}

.grid-groups {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 24px;
    margin-bottom: 40px;
}

/* ====== BUTTONS ====== */
.btn {
    background-color: var(--md-sys-color-primary);
    color: var(--md-sys-color-on-primary);
    border: none;
    padding: 0 24px;
    height: 48px;
    border-radius: 24px; /* Pill */
    font-family: 'Inter', 'Roboto', sans-serif;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    position: relative;
    overflow: hidden;
    transition: background-color 0.2s, box-shadow 0.2s, color 0.2s, transform 0.1s;
    white-space: nowrap;
}
.btn::after {
    content: ''; position: absolute; inset: 0; background-color: var(--md-sys-color-on-primary); opacity: 0; transition: opacity 0.2s;
}
.btn:hover { box-shadow: var(--md-sys-elevation-1); }
.btn:hover::after { opacity: 0.08; }
.btn:active { transform: scale(0.97); }
.btn:active::after { opacity: 0.12; }

.btn.secondary { background-color: var(--md-sys-color-secondary-container); color: var(--md-sys-color-on-secondary-container); }
.btn.secondary::after { background-color: var(--md-sys-color-on-secondary-container); }

.btn.sm { height: 40px; padding: 0 20px; font-size: 14px; }
.btn.tonal { background-color: var(--md-sys-color-tertiary-container); color: var(--md-sys-color-on-tertiary-container); }

.icon-btn {
    width: 40px; height: 40px; padding: 0; border-radius: 50%;
    background: transparent; color: var(--md-sys-color-on-surface-variant);
    border: none; cursor: pointer; display: flex; align-items: center; justify-content: center;
    position: relative; overflow: hidden;
}
.icon-btn::after {
    content:''; position:absolute; inset:0; background-color: var(--md-sys-color-on-surface-variant); opacity:0; transition: opacity 0.2s;
}
.icon-btn:hover::after { opacity: 0.08; }
.icon-btn.active { color: var(--md-sys-color-primary); }

/* ====== FILTER CHIPS (Replaces Checkboxes) ====== */
.chip-group { display: flex; flex-wrap: wrap; gap: 12px; }
.chip {
    appearance: none;
    background-color: var(--md-sys-color-surface-container);
    color: var(--md-sys-color-on-surface-variant);
    border: 1px solid var(--md-sys-color-outline-variant);
    border-radius: 8px;
    padding: 8px 16px;
    font-family: inherit;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    transition: all 0.2s ease;
    user-select: none;
}
.chip:hover {
    background-color: var(--md-sys-color-surface-container-high);
}
.chip.selected {
    background-color: var(--md-sys-color-secondary-container);
    color: var(--md-sys-color-on-secondary-container);
    border-color: transparent;
}
.chip.selected .check-icon {
    display: inline-block;
    font-family: 'Material Symbols Outlined';
    font-size: 18px;
    content: 'check';
}

/* ====== INPUT GROUP ====== */
.input-group {
    display: flex;
    background-color: var(--md-sys-color-surface-container);
    border-radius: 16px;
    border: 1px solid var(--md-sys-color-outline-variant);
    overflow: hidden;
    height: 48px;
}
.input-group select {
    flex: 1;
    background: transparent;
    color: var(--md-sys-color-on-surface);
    border: none;
    padding: 0 16px;
    font-family: inherit;
    font-size: 15px;
    outline: none;
    cursor: pointer;
    appearance: none;
}
.input-group .input-btn {
    background-color: var(--md-sys-color-secondary-container);
    color: var(--md-sys-color-on-secondary-container);
    border: none;
    border-left: 1px solid var(--md-sys-color-outline-variant);
    padding: 0 20px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
}
.input-group .input-btn:hover {
    background-color: color-mix(in srgb, var(--md-sys-color-secondary-container) 90%, var(--md-sys-color-on-secondary-container) 10%);
}

/* ====== TABLES ====== */
.table-container {
    background-color: var(--md-sys-color-surface-container-lowest);
    border-radius: 20px;
    overflow: hidden;
    border: 1px solid var(--md-sys-color-outline-variant);
}
table { width: 100%; border-collapse: collapse; text-align: left; }
th {
    background-color: var(--md-sys-color-surface-container-low);
    color: var(--md-sys-color-on-surface-variant);
    font-family: 'Outfit', sans-serif;
    font-size: 14px;
    font-weight: 500;
    padding: 16px 24px;
    border-bottom: 1px solid var(--md-sys-color-outline-variant);
    white-space: nowrap;
}
td {
    padding: 16px 24px;
    border-bottom: 1px solid var(--md-sys-color-surface-variant);
    font-size: 14px;
    color: var(--md-sys-color-on-surface);
}
tr:last-child td { border-bottom: none; }
.node-row { transition: background-color 0.2s; cursor: pointer; }
.node-row:hover { background-color: var(--md-sys-color-surface-container); }

/* Badges & Scores */
.badge { display: inline-flex; alignItems: center; padding: 4px 10px; border-radius: 6px; font-size: 12px; font-weight: 600; gap: 4px; }
.badge.primary { background-color: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); }
.score-box { font-size: 15px; font-weight: 600; font-family: 'Outfit', sans-serif; color: var(--md-sys-color-primary); padding: 4px 12px; border-radius: 8px; background-color: var(--md-sys-color-surface-container-highest); display: inline-block; }

/* Console */
.console {
    background-color: #0b0c0f;
    color: #e2e2e9;
    font-family: 'Roboto Mono', monospace;
    padding: 24px;
    border-radius: 16px;
    height: 600px;
    overflow-y: auto;
    font-size: 13px;
    line-height: 1.6;
    border: 1px solid var(--md-sys-color-outline-variant);
}
.log-line { display: flex; margin-bottom: 6px; animation: slideIn 0.3s ease-out; }
@keyframes slideIn { from{transform: translateY(10px); opacity:0;} to{transform: translateY(0); opacity:1;} }
.log-time { color: var(--md-sys-color-outline); margin-right: 12px; white-space: nowrap;}
.log-badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; margin-right: 12px; font-weight: bold; }
.log-info .log-badge { background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); }
.log-success .log-badge { background: var(--md-sys-color-success-container); color: var(--md-sys-color-on-success-container); }
.log-warning .log-badge { background: var(--md-sys-color-warning-container); color: var(--md-sys-color-on-warning-container); }
.log-error .log-badge { background: var(--md-sys-color-error-container); color: var(--md-sys-color-on-error-container); }
.log-success .log-msg { color: var(--md-sys-color-success); }
.log-warning .log-msg { color: var(--md-sys-color-warning); }
.log-error .log-msg { color: var(--md-sys-color-error); }

.tab-content { display: none; }
.tab-content.active { display: block; animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* Animations */
.spin { animation: spin 2s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }
"""

with open("frontend/src/index.css", "w", encoding="utf-8") as f:
    f.write(css_content)

# 3. Rewrite App.tsx
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
        <div className="app-layout">
            {/* MD3 Navigation Rail */}
            <nav className="nav-rail">
                <div className="nav-rail-top">
                    <div className="brand-icon">
                        <span className="material-symbols-outlined" style={{fontSize: '32px'}}>rocket_launch</span>
                    </div>
                </div>
                
                <button className={`nav-item ${activeTab === 'home' ? 'active' : ''}`} onClick={() => setActiveTab('home')}>
                    <div className="icon-container">
                        <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'home' ? "'FILL' 1" : "'FILL' 0"}}>home</span>
                    </div>
                    <span className="nav-label">首頁</span>
                </button>
                
                <button className={`nav-item ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
                    <div className="icon-container">
                        <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'logs' ? "'FILL' 1" : "'FILL' 0"}}>terminal</span>
                    </div>
                    <span className="nav-label">日誌</span>
                </button>

                <div className="nav-rail-spacer"></div>

                <button className="nav-item" onClick={toggleTheme}>
                    <div className="icon-container">
                        <span className="material-symbols-outlined">{isLightTheme ? 'dark_mode' : 'light_mode'}</span>
                    </div>
                    <span className="nav-label">主題</span>
                </button>
            </nav>

            <main className="main-content">
                {/* Hero Section - Always visible */}
                <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} />

                {/* Tabs Content */}
                <div className={`tab-content ${activeTab === 'home' ? 'active' : ''}`}>
                    <div className="md3-headline-large" style={{marginBottom: '24px'}}>節點群組管理</div>
                    <div className="grid-groups">
                        {groups.map(g => (
                            <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                        ))}
                    </div>
                    
                    <div className="md3-headline-large" style={{marginBottom: '24px', marginTop: '48px'}}>節點即時排行榜</div>
                    <NodeRanking stats={stats} />
                </div>

                <div className={`tab-content ${activeTab === 'logs' ? 'active' : ''}`}>
                    <div className="md3-headline-large" style={{marginBottom: '24px'}}>系統日誌</div>
                    <div className="card" style={{padding: '0', border: 'none'}}>
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
with open("frontend/src/App.tsx", "w", encoding="utf-8") as f:
    f.write(app_tsx)

# 4. Rewrite Dashboard.tsx
dash_tsx = """import type { Status } from '../hooks/useApi';

export default function Dashboard({ status, triggerTest, togglePause }: { status: Status, triggerTest: () => void, togglePause: () => void }) {
    return (
        <div className="hero-card">
            <div className="hero-card-content">
                <div>
                    <div className="md3-headline-large" style={{marginBottom:'12px'}}>
                        Clash Node Rover 核心
                    </div>
                    <div style={{display: 'flex', gap: '16px', alignItems: 'center'}}>
                        <span className={`hero-badge ${status.is_paused ? 'paused' : 'running'}`}>
                            <span className="material-symbols-outlined" style={{fontSize: '18px'}}>{status.is_paused ? 'pause_circle' : 'check_circle'}</span>
                            {status.is_paused ? '引擎已暫停' : '引擎運作中'}
                        </span>
                        <span className="md3-body-medium" style={{color: 'var(--md-sys-color-on-surface-variant)'}}>
                            自動測速週期：5 分鐘
                        </span>
                    </div>
                </div>
                
                <div style={{display:'flex', gap:'16px'}}>
                    <button className="btn" onClick={triggerTest} disabled={status.is_running}>
                        <span className={`material-symbols-outlined ${status.is_running ? 'spin' : ''}`}>
                            {status.is_running ? 'refresh' : 'speed'}
                        </span>
                        {status.is_running ? '測速中...' : '立即測速'}
                    </button>
                    <button className={`btn ${status.is_paused ? 'primary' : 'tonal'}`} onClick={togglePause}>
                        <span className="material-symbols-outlined">{status.is_paused ? 'play_arrow' : 'pause'}</span>
                        {status.is_paused ? '恢復運作' : '暫停核心'}
                    </button>
                </div>
            </div>
        </div>
    );
}
"""
with open("frontend/src/components/Dashboard.tsx", "w", encoding="utf-8") as f:
    f.write(dash_tsx)


# 5. Rewrite GroupCard.tsx (Filter Chips)
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
            if (r === val) isR = !currentSelected; // Toggle
            if (isR) regexes.push(REGION_PRESETS[r]);
        });
        saveFilter(group.name, {
            keyword_regex: regexes.join('|'),
            check_chatgpt: isChatGPT,
            check_gemini: isGemini,
            check_antigravity: isAntigravity
        });
    };

    const handleServiceChange = (service: string, currentSelected: boolean) => {
        saveFilter(group.name, {
            keyword_regex: rx,
            check_chatgpt: service === 'chatgpt' ? !currentSelected : isChatGPT,
            check_gemini: service === 'gemini' ? !currentSelected : isGemini,
            check_antigravity: service === 'antigravity' ? !currentSelected : isAntigravity
        });
    };

    return (
        <div className="group-card">
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start'}}>
                <div>
                    <div className="md3-title-large" style={{marginBottom: '4px'}}>{group.name}</div>
                    <div className="md3-body-medium" style={{color: 'var(--md-sys-color-on-surface-variant)'}}>
                        運行中 &bull; 共 {group.all_count} 個節點
                    </div>
                </div>
                <button className={`icon-btn ${group.locked ? 'active' : ''}`} onClick={() => toggleGroupLock(group.name, !group.locked)} title={group.locked ? '點擊解鎖 (恢復自動切換)' : '點擊鎖定 (停止自動切換)'}>
                    <span className="material-symbols-outlined">{group.locked ? 'lock' : 'lock_open_right'}</span>
                </button>
            </div>

            <div style={{backgroundColor: 'var(--md-sys-color-surface-container)', padding: '16px', borderRadius: '16px', border: '1px solid var(--md-sys-color-outline-variant)'}}>
                <div className="md3-label-large" style={{color: 'var(--md-sys-color-primary)', marginBottom: '8px'}}>目前使用節點</div>
                <div className="md3-title-medium">{group.now || '未選擇'}</div>
                {group.provider && <div className="badge primary" style={{marginTop: '12px'}}><span className="material-symbols-outlined" style={{fontSize:'14px'}}>corporate_fare</span> {group.provider}</div>}
            </div>

            <div className="input-group">
                <select id={`select-${group.name}`} defaultValue={group.now}>
                    {group.all_nodes.map((n: string) => <option key={n} value={n} >{n}</option>)}
                </select>
                <button className="input-btn" onClick={() => {
                    const sel = document.getElementById(`select-${group.name}`) as HTMLSelectElement;
                    manualSwitch(group.name, sel.value);
                }}>手動切換</button>
            </div>
            
            <div style={{marginTop: '8px'}}>
                <div className="md3-label-large" style={{marginBottom: '12px', color: 'var(--md-sys-color-on-surface-variant)'}}>地區限制</div>
                <div className="chip-group">
                    {['US', 'HK', 'TW', 'JP', 'SG', 'UK'].map(r => {
                        const isSelected = rx.includes(r + '|');
                        return (
                            <div key={r} className={`chip ${isSelected ? 'selected' : ''}`} onClick={() => handleRegionChange(r, isSelected)}>
                                {isSelected && <span className="material-symbols-outlined check-icon">check</span>}
                                {r === 'US' ? '🇺🇸 美國' : r === 'HK' ? '🇭🇰 香港' : r === 'TW' ? '🇹🇼 台灣' : r === 'JP' ? '🇯🇵 日本' : r === 'SG' ? '🇸🇬 新加坡' : '🇬🇧 英國'}
                            </div>
                        );
                    })}
                </div>

                <div className="md3-label-large" style={{marginTop: '20px', marginBottom: '12px', color: 'var(--md-sys-color-on-surface-variant)'}}>服務連線驗證</div>
                <div className="chip-group">
                    <div className={`chip ${isChatGPT ? 'selected' : ''}`} onClick={() => handleServiceChange('chatgpt', isChatGPT)}>
                        {isChatGPT && <span className="material-symbols-outlined check-icon">check</span>} 🤖 ChatGPT
                    </div>
                    <div className={`chip ${isGemini ? 'selected' : ''}`} onClick={() => handleServiceChange('gemini', isGemini)}>
                        {isGemini && <span className="material-symbols-outlined check-icon">check</span>} ✨ Gemini
                    </div>
                    <div className={`chip ${isAntigravity ? 'selected' : ''}`} onClick={() => handleServiceChange('antigravity', isAntigravity)}>
                        {isAntigravity && <span className="material-symbols-outlined check-icon">check</span>} 🚀 Antigravity
                    </div>
                </div>
            </div>
        </div>
    );
}
"""
with open("frontend/src/components/GroupCard.tsx", "w", encoding="utf-8") as f:
    f.write(gc_tsx)


# 6. Rewrite NodeRanking.tsx
nr_tsx = """export default function NodeRanking({ stats }: any) {
    if (!stats || stats.length === 0) {
        return (
            <div className="card" style={{textAlign:'center', padding:'60px 20px', color:'var(--md-sys-color-on-surface-variant)'}}>
                <span className="material-symbols-outlined" style={{fontSize:'48px', marginBottom:'16px', opacity: 0.5}}>hourglass_empty</span>
                <div className="md3-title-medium">目前尚無節點數據，請等待系統測速完成...</div>
            </div>
        );
    }

    return (
        <div className="table-container">
            <table>
                <thead>
                    <tr>
                        <th style={{width: '80px', textAlign: 'center'}}>排名</th>
                        <th>節點名稱</th>
                        <th style={{textAlign: 'center'}}>綜合分數</th>
                        <th style={{textAlign: 'center'}}>平均延遲</th>
                        <th style={{textAlign: 'center'}}>抖動 (Jitter)</th>
                        <th>所屬群組狀態</th>
                    </tr>
                </thead>
                <tbody>
                    {stats.map((s: any, idx: number) => {
                        const scoreStr = s.score.toFixed(2);
                        return (
                            <tr key={s.node_name} className="node-row">
                                <td style={{textAlign: 'center', fontWeight: '600', color: idx < 3 ? 'var(--md-sys-color-primary)' : 'inherit'}}>
                                    {idx === 0 ? '🏆 1' : idx === 1 ? '🥈 2' : idx === 2 ? '🥉 3' : idx + 1}
                                </td>
                                <td style={{fontWeight: '500'}}>{s.node_name}</td>
                                <td style={{textAlign: 'center'}}><div className="score-box">{scoreStr}</div></td>
                                <td style={{textAlign: 'center'}}>{s.avg_delay > 0 ? `${s.avg_delay} ms` : '-'}</td>
                                <td style={{textAlign: 'center'}}>{s.jitter > 0 ? `${s.jitter} ms` : '-'}</td>
                                <td>
                                    {s.in_groups && s.in_groups.length > 0 ? (
                                        <div style={{display:'flex', gap:'8px', flexWrap:'wrap'}}>
                                            {s.in_groups.map((g: string) => <span key={g} className="badge primary">{g}</span>)}
                                        </div>
                                    ) : (
                                        <span style={{color: 'var(--md-sys-color-outline)'}}>未被選用</span>
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
with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f:
    f.write(nr_tsx)

print("Redesign complete.")
