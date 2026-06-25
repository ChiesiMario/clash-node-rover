import { useEffect, useState } from 'react';
import { useApi } from './hooks/useApi';
import { useWebSocket } from './hooks/useWebSocket';
import Dashboard from './components/Dashboard';
import GroupCard from './components/GroupCard';
import NodeRanking from './components/NodeRanking';
import logo from './assets/logo.png';

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
                    <div className="brand-icon" style={{background: 'transparent', boxShadow: 'none'}}>
                        <img src={logo} alt="Rover Logo" style={{width: '32px', height: '32px', borderRadius: '8px'}} />
                    </div>
                    <div className="hig-headline">Clash Node Rover</div>
                </div>
                
                <button className={`nav-item ${activeTab === 'home' ? 'active' : ''}`} onClick={() => setActiveTab('home')}>
                    <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'home' ? "'FILL' 1" : "'FILL' 0"}}>home</span>
                    <span className="hig-body">總覽</span>
                </button>
                
                <button className={`nav-item ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}>
                    <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'logs' ? "'FILL' 1" : "'FILL' 0"}}>terminal</span>
                    <span className="hig-body">系統日誌</span>
                </button>

                <div className="sidebar-spacer"></div>

                <button className="nav-item" onClick={toggleTheme}>
                    <span className="material-symbols-outlined">{isLightTheme ? 'dark_mode' : 'light_mode'}</span>
                    <span className="hig-body">外觀切換</span>
                </button>
            </aside>

            <main className="main-content">
                <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} />

                <div className={`tab-content ${activeTab === 'home' ? 'active' : ''}`}>
                    <div className="hig-title-2" style={{marginBottom: '24px'}}>節點群組管理</div>
                    <div className="grid-groups">
                        {groups.map(g => (
                            <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                        ))}
                    </div>
                    
                    <div className="hig-title-2" style={{marginBottom: '24px', marginTop: '48px'}}>節點即時排行榜</div>
                    <NodeRanking stats={stats} />
                </div>

                <div className={`tab-content ${activeTab === 'logs' ? 'active' : ''}`}>
                    <div className="hig-title-2" style={{marginBottom: '24px'}}>系統即時日誌</div>
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
