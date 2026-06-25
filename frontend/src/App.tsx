import { useEffect, useState } from 'react';
import { useApi } from './hooks/useApi';
import { useWebSocket } from './hooks/useWebSocket';
import Dashboard from './components/Dashboard';
import GroupCard from './components/GroupCard';
import NodeRanking from './components/NodeRanking';
import SetupModal from './components/SetupModal';
import SettingsPage from './components/SettingsPage';
import logo from './assets/logo.png';

function App() {
    const { stats, status, groups, fetchStats, fetchStatus, fetchGroups, triggerTest, togglePause, manualSwitch, toggleGroupLock, saveFilter } = useApi();
    const [activeTab, setActiveTab] = useState('home');
    const [isLightTheme, setIsLightTheme] = useState(false);
    const [setupState, setSetupState] = useState<{ isConfigured: boolean; apiUrl: string } | null>(null);
    const [showSetupModal, setShowSetupModal] = useState(false);

    const fetchSetupStatus = async () => {
        try {
            const res = await fetch('/api/setup');
            if (res.ok) {
                const data = await res.json();
                setSetupState({ isConfigured: data.is_configured, apiUrl: data.api_url });
                if (data.is_configured) {
                    fetchGroups();
                    fetchStats();
                    fetchStatus();
                }
            }
        } catch (err) {
            console.error('Failed to fetch setup status', err);
        }
    };

    const { logs } = useWebSocket(() => {
        fetchGroups();
        fetchStats();
        fetchStatus();
        fetchSetupStatus();
    });

    useEffect(() => {
        fetchSetupStatus();
        
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
        <>
        {setupState && (!setupState.isConfigured || showSetupModal) && (
            <SetupModal 
                defaultUrl={setupState.apiUrl} 
                canCancel={setupState.isConfigured}
                onCancel={() => setShowSetupModal(false)}
                onSuccess={() => {
                    setShowSetupModal(false);
                    fetchSetupStatus();
                }} 
            />
        )}
        <div className="app-layout" style={{ filter: (setupState && !setupState.isConfigured) ? 'blur(4px)' : 'none', pointerEvents: (setupState && !setupState.isConfigured) ? 'none' : 'auto' }}>
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

                <button className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`} onClick={() => setActiveTab('settings')}>
                    <span className="material-symbols-outlined" style={{fontVariationSettings: activeTab === 'settings' ? "'FILL' 1" : "'FILL' 0"}}>settings</span>
                    <span className="hig-body">設定</span>
                </button>

                <div className="sidebar-spacer"></div>

                <button className="nav-item" onClick={toggleTheme}>
                    <span className="material-symbols-outlined">{isLightTheme ? 'dark_mode' : 'light_mode'}</span>
                    <span className="hig-body">外觀切換</span>
                </button>
            </aside>

            <main className="main-content">
                <Dashboard status={status} triggerTest={triggerTest} togglePause={togglePause} apiUrl={setupState?.apiUrl} />

                <div className={`tab-content ${activeTab === 'home' ? 'active' : ''}`}>
                    <div className="hig-title-2" style={{marginBottom: '24px'}}>節點群組管理</div>
                    {groups.length === 0 ? (
                        <div className="hig-card" style={{textAlign:'center', padding:'60px 20px', color:'var(--hig-text-secondary)'}}>
                            <span className="material-symbols-outlined" style={{fontSize:'48px', marginBottom:'16px', opacity: 0.5}}>folder_open</span>
                            <div className="hig-headline">尚無監控的節點群組</div>
                            <div className="hig-body" style={{marginTop:'8px', opacity: 0.8}}>請先前往「系統設定」加入需要監控的目標群組</div>
                        </div>
                    ) : (
                        <div className="grid-groups">
                            {groups.map(g => (
                                <GroupCard key={g.name} group={g} manualSwitch={manualSwitch} toggleGroupLock={toggleGroupLock} saveFilter={saveFilter} />
                            ))}
                        </div>
                    )}
                    
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

                <div className={`tab-content ${activeTab === 'settings' ? 'active' : ''}`}>
                    <SettingsPage 
                        apiConnected={status.api_connected} 
                        onSaveSuccess={() => {
                            fetchStatus();
                            fetchSetupStatus();
                        }} 
                    />
                </div>
            </main>
        </div>
        </>
    );
}

export default App;
