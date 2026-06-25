import type { Status } from '../hooks/useApi';

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
