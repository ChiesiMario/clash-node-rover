import type { Status } from '../hooks/useApi';

export default function Dashboard({ status, triggerTest, togglePause, apiUrl }: { status: Status, triggerTest: () => void, togglePause: () => void, apiUrl?: string }) {
    let dotColor = 'var(--hig-system-green)';
    let dotShadow = 'rgba(48, 209, 88, 0.4)';
    let badgeText = '監控中';
    const isApiReady = status.is_configured !== false && status.api_connected !== false;
    
    if (status.is_configured === false) {
        dotColor = 'var(--hig-system-red)';
        dotShadow = 'rgba(255, 69, 58, 0.4)';
        badgeText = '未設置 API';
    } else if (status.api_connected === false) {
        dotColor = 'var(--hig-system-red)';
        dotShadow = 'rgba(255, 69, 58, 0.4)';
        badgeText = 'API 未連接';
    } else if (status.is_paused) {
        dotColor = 'var(--hig-system-orange)';
        dotShadow = 'rgba(255, 159, 10, 0.4)';
        badgeText = '已暫停';
    }

    return (
        <div className="hig-card" style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '24px'}}>
            <div>
                <div style={{display: 'flex', alignItems: 'center', gap: '16px'}}>
                    <div className="status-dot" style={{backgroundColor: dotColor, '--dot-shadow': dotShadow} as any}></div>
                    <span style={{fontSize: '32px', fontWeight: 700, color: 'var(--hig-text-primary)'}}>
                        {badgeText}
                    </span>
                </div>
                <div className="hig-footnote" style={{color: 'var(--hig-text-secondary)', marginTop: '8px', display: 'flex', alignItems: 'center', gap: '6px'}}>
                    <span className="material-symbols-outlined" style={{fontSize: '16px'}}>link</span>
                    {apiUrl || '尚未設定'}
                </div>
            </div>
            
            <div style={{display:'flex', gap:'12px'}}>
                <button className="btn" onClick={triggerTest} disabled={status.is_running || !isApiReady}>
                    <span className={`material-symbols-outlined ${status.is_running ? 'spin' : ''}`}>
                        {status.is_running ? 'refresh' : 'speed'}
                    </span>
                    {status.is_running ? '測試中...' : '立即測試'}
                </button>
                <button className={`btn secondary`} onClick={togglePause} disabled={!isApiReady}>
                    <span className={`material-symbols-outlined ${status.is_paused ? 'fill' : ''}`}>{status.is_paused ? 'play_arrow' : 'pause'}</span>
                    {status.is_paused ? '恢復' : '暫停'}
                </button>
            </div>
        </div>
    );
}
