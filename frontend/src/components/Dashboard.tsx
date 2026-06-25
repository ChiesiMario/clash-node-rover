import type { Status } from '../hooks/useApi';

export default function Dashboard({ status, triggerTest, togglePause }: { status: Status, triggerTest: () => void, togglePause: () => void }) {
    return (
        <div className="hig-card" style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '24px'}}>
            <div>
                <div className="hig-title-1" style={{marginBottom:'8px'}}>控制中心</div>
                <div style={{display: 'flex', gap: '12px', alignItems: 'center'}}>
                    <span className={`hig-badge ${status.is_paused ? 'orange' : 'green'}`}>
                        {status.is_paused ? '已暫停' : '監控中'}
                    </span>
                    <span className="hig-footnote" style={{color: 'var(--hig-text-secondary)'}}>
                        每 5 分鐘自動巡測
                    </span>
                </div>
            </div>
            
            <div style={{display:'flex', gap:'12px'}}>
                <button className="btn" onClick={triggerTest} disabled={status.is_running}>
                    <span className={`material-symbols-outlined ${status.is_running ? 'spin' : ''}`}>
                        {status.is_running ? 'refresh' : 'speed'}
                    </span>
                    {status.is_running ? '測試中...' : '立即測試'}
                </button>
                <button className={`btn secondary`} onClick={togglePause}>
                    <span className={`material-symbols-outlined ${status.is_paused ? 'fill' : ''}`}>{status.is_paused ? 'play_arrow' : 'pause'}</span>
                    {status.is_paused ? '恢復' : '暫停'}
                </button>
            </div>
        </div>
    );
}
