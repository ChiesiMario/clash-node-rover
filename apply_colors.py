import os

# 1. Append themes and popup styles to index.css
with open("frontend/src/index.css", "a", encoding="utf-8") as f:
    f.write("""

/* ====== MD3 COLOR THEMES ====== */

/* Theme: Green */
:root.theme-green {
    --md-sys-color-primary: #82da9d; --md-sys-color-on-primary: #003919;
    --md-sys-color-primary-container: #005226; --md-sys-color-on-primary-container: #9df7b8;
    --md-sys-color-secondary: #b3ccb9; --md-sys-color-on-secondary: #1f3527;
    --md-sys-color-secondary-container: #354b3c; --md-sys-color-on-secondary-container: #cfe8d4;
}
:root.light-theme.theme-green {
    --md-sys-color-primary: #006d34; --md-sys-color-on-primary: #ffffff;
    --md-sys-color-primary-container: #9df7b8; --md-sys-color-on-primary-container: #00210b;
    --md-sys-color-secondary: #4c6353; --md-sys-color-on-secondary: #ffffff;
    --md-sys-color-secondary-container: #cfe8d4; --md-sys-color-on-secondary-container: #092013;
}

/* Theme: Purple */
:root.theme-purple {
    --md-sys-color-primary: #d0bcff; --md-sys-color-on-primary: #381e72;
    --md-sys-color-primary-container: #4f378b; --md-sys-color-on-primary-container: #eaddff;
    --md-sys-color-secondary: #cbc2db; --md-sys-color-on-secondary: #332d41;
    --md-sys-color-secondary-container: #4a4458; --md-sys-color-on-secondary-container: #e8def8;
}
:root.light-theme.theme-purple {
    --md-sys-color-primary: #6750a4; --md-sys-color-on-primary: #ffffff;
    --md-sys-color-primary-container: #eaddff; --md-sys-color-on-primary-container: #21005d;
    --md-sys-color-secondary: #625b71; --md-sys-color-on-secondary: #ffffff;
    --md-sys-color-secondary-container: #e8def8; --md-sys-color-on-secondary-container: #1d192b;
}

/* Theme: Rose */
:root.theme-rose {
    --md-sys-color-primary: #ffb4ab; --md-sys-color-on-primary: #690005;
    --md-sys-color-primary-container: #93000a; --md-sys-color-on-primary-container: #ffdad6;
    --md-sys-color-secondary: #e7bdb8; --md-sys-color-on-secondary: #442926;
    --md-sys-color-secondary-container: #5d3f3b; --md-sys-color-on-secondary-container: #ffdad6;
}
:root.light-theme.theme-rose {
    --md-sys-color-primary: #ba1a1a; --md-sys-color-on-primary: #ffffff;
    --md-sys-color-primary-container: #ffdad6; --md-sys-color-on-primary-container: #410002;
    --md-sys-color-secondary: #775652; --md-sys-color-on-secondary: #ffffff;
    --md-sys-color-secondary-container: #ffdad6; --md-sys-color-on-secondary-container: #2c1512;
}

/* ====== PALETTE POPUP ====== */
.palette-popup-container {
    position: relative;
    width: 100%;
}
.palette-popup {
    position: absolute;
    bottom: 0;
    left: 80px; /* Right of Nav Rail */
    background-color: var(--md-sys-color-surface-container-high);
    border-radius: 16px;
    padding: 16px;
    box-shadow: var(--md-sys-elevation-3);
    display: flex;
    flex-direction: column;
    gap: 12px;
    z-index: 100;
    width: 120px;
    animation: slideInLeft 0.2s cubic-bezier(0.2, 0, 0, 1);
}
@media(max-width: 768px) {
    .palette-popup {
        bottom: 80px; /* Above bottom nav */
        left: auto;
        right: 16px;
        animation: slideInUp 0.2s cubic-bezier(0.2, 0, 0, 1);
    }
}
@keyframes slideInLeft {
    from { opacity: 0; transform: translateX(-10px); }
    to { opacity: 1; transform: translateX(0); }
}
@keyframes slideInUp {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
}

.color-option {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 12px;
    border-radius: 8px;
    cursor: pointer;
    transition: background-color 0.2s;
    font-size: 14px;
    font-weight: 500;
    color: var(--md-sys-color-on-surface);
}
.color-option:hover { background-color: var(--md-sys-color-surface-container-highest); }
.color-option.active { background-color: var(--md-sys-color-secondary-container); color: var(--md-sys-color-on-secondary-container); }
.color-dot {
    width: 16px;
    height: 16px;
    border-radius: 50%;
}
.dot-blue { background-color: #a8c7fa; }
.dot-green { background-color: #82da9d; }
.dot-purple { background-color: #d0bcff; }
.dot-rose { background-color: #ffb4ab; }
:root.light-theme .dot-blue { background-color: #0842a0; }
:root.light-theme .dot-green { background-color: #006d34; }
:root.light-theme .dot-purple { background-color: #6750a4; }
:root.light-theme .dot-rose { background-color: #ba1a1a; }
""")

# 2. Update App.tsx
app_tsx = """import { useEffect, useState, useRef } from 'react';
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
    
    // Theme Color Picker
    const [themeColor, setThemeColor] = useState('blue');
    const [showPalette, setShowPalette] = useState(false);
    const popupRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        fetchGroups();
        fetchStats();
        fetchStatus();
        
        // Load initial theme mode
        const savedThemeMode = localStorage.getItem('themeMode');
        if (savedThemeMode === 'light') {
            setIsLightTheme(true);
            document.documentElement.classList.add('light-theme');
        }
        
        // Load initial theme color
        const savedThemeColor = localStorage.getItem('themeColor') || 'blue';
        setThemeColor(savedThemeColor);
        applyThemeColor(savedThemeColor);

        // Click outside to close popup
        const handleClickOutside = (event: MouseEvent) => {
            if (popupRef.current && !popupRef.current.contains(event.target as Node)) {
                setShowPalette(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const toggleTheme = () => {
        const newThemeMode = !isLightTheme;
        setIsLightTheme(newThemeMode);
        document.documentElement.classList.toggle('light-theme', newThemeMode);
        localStorage.setItem('themeMode', newThemeMode ? 'light' : 'dark');
    };

    const applyThemeColor = (color: string) => {
        document.documentElement.classList.remove('theme-blue', 'theme-green', 'theme-purple', 'theme-rose');
        document.documentElement.classList.add(`theme-${color}`);
    };

    const handleColorSelect = (color: string) => {
        setThemeColor(color);
        applyThemeColor(color);
        localStorage.setItem('themeColor', color);
        setShowPalette(false);
    };

    const COLORS = [
        { id: 'blue', name: '預設藍', cls: 'dot-blue' },
        { id: 'green', name: '薄荷綠', cls: 'dot-green' },
        { id: 'purple', name: '丁香紫', cls: 'dot-purple' },
        { id: 'rose', name: '玫瑰紅', cls: 'dot-rose' }
    ];

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

                {/* Palette Picker */}
                <div className="palette-popup-container" ref={popupRef}>
                    <button className={`nav-item ${showPalette ? 'active' : ''}`} onClick={() => setShowPalette(!showPalette)}>
                        <div className="icon-container">
                            <span className="material-symbols-outlined" style={{fontVariationSettings: showPalette ? "'FILL' 1" : "'FILL' 0"}}>palette</span>
                        </div>
                        <span className="nav-label">色彩</span>
                    </button>
                    {showPalette && (
                        <div className="palette-popup">
                            <div className="md3-label-large" style={{color: 'var(--md-sys-color-on-surface-variant)', padding: '0 8px', marginBottom: '4px'}}>主題色彩</div>
                            {COLORS.map(c => (
                                <div key={c.id} className={`color-option ${themeColor === c.id ? 'active' : ''}`} onClick={() => handleColorSelect(c.id)}>
                                    <div className={`color-dot ${c.cls}`}></div>
                                    <span>{c.name}</span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Dark/Light Mode */}
                <button className="nav-item" onClick={toggleTheme}>
                    <div className="icon-container">
                        <span className="material-symbols-outlined">{isLightTheme ? 'dark_mode' : 'light_mode'}</span>
                    </div>
                    <span className="nav-label">外觀</span>
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

print("Color picker applied.")
