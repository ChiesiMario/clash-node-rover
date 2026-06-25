import type { Status } from '../hooks/useApi';

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
