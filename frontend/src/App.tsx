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
