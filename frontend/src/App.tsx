import { useEffect, useState } from 'react';
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
        <div className="container">
            <div className="top-app-bar" style={{marginBottom: '24px', borderRadius: '16px'}}>
                <div className="app-title md3-title-large">
                    <span className="material-symbols-outlined" style={{color: "var(--md-sys-color-primary)", fontSize: "28px"}}>rocket_launch</span>
                    Clash Node Rover
                </div>
                <button className="icon-btn" onClick={toggleTheme} title="切換深色/淺色主題">
                    <span className="material-symbols-outlined" id="themeIcon">
                        {isLightTheme ? 'dark_mode' : 'light_mode'}
                    </span>
                </button>
            </div>

                        <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} />

            <div className="segmented-button" style={{display: 'inline-flex', marginBottom: '24px'}}>
                <button id="btn-home" className={`seg-btn ${activeTab === 'home' ? 'active' : ''}`} onClick={() => setActiveTab('home')}>
                    <span className="material-symbols-outlined" style={{fontSize:'18px'}}>home</span> 首頁
                </button>
                <button id="btn-logs" className={`seg-btn ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
                    <span className="material-symbols-outlined" style={{fontSize:'18px'}}>terminal</span> 系統日誌
                </button>
            </div>

            <div id="tab-home" className={`tab-content ${activeTab === 'home' ? 'active' : ''}`} style={{display: activeTab === 'home' ? 'block' : 'none'}}>
                <div style={{marginBottom: '24px'}}>
                    <div className="grid" id="groupsGrid" style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '16px'}}>
                        {groups.map(g => (
                            <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                        ))}
                    </div>
                </div>
                <NodeRanking stats={stats} />
            </div>

            <div id="tab-logs" className={`tab-content ${activeTab === 'logs' ? 'active' : ''}`} style={{display: activeTab === 'logs' ? 'block' : 'none'}}>
                <div className="card">
                    <div className="md3-title-medium" style={{marginBottom:'16px'}}>即時系統日誌</div>
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
    );
}

export default App;
