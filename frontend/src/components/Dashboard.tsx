import type { Status } from '../hooks/useApi';

export default function Dashboard({ status, triggerTest, togglePause }: { status: Status, triggerTest: () => void, togglePause: () => void }) {
    return (
        <div className="card" style={{marginBottom: '24px'}}>
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', flexWrap:'wrap', gap:'16px'}}>
                <div>
                    <div className="md3-title-large" style={{marginBottom:'8px'}}>
                        Brain 核心引擎 <span className="md3-label-medium" style={{padding:'4px 8px', borderRadius:'12px', backgroundColor: status.is_paused ? 'var(--md-sys-color-warning-container)' : 'var(--md-sys-color-success-container)', color: status.is_paused ? 'var(--md-sys-color-on-warning-container)' : 'var(--md-sys-color-on-success-container)', verticalAlign:'middle', marginLeft:'8px'}}>
                            {status.is_paused ? '已暫停' : '運作中'}
                        </span>
                    </div>
                    <div className="md3-body-medium" style={{color:'var(--md-sys-color-on-surface-variant)'}}>自動測速週期：5 分鐘</div>
                </div>
                <div style={{display:'flex', gap:'12px'}}>
                    <button id="triggerBtn" className="btn" onClick={triggerTest} disabled={status.is_running}>
                        <span className={`material-symbols-outlined ${status.is_running ? 'spin' : ''}`} id="triggerIcon">
                            {status.is_running ? 'refresh' : 'speed'}
                        </span>
                        <span id="triggerText">{status.is_running ? 'Testing...' : 'Run Test'}</span>
                    </button>
                    <button id="pauseBtn" className={`btn ${status.is_paused ? '' : 'secondary'}`} onClick={togglePause}>
                        <span className="material-symbols-outlined">{status.is_paused ? 'play_arrow' : 'pause'}</span>
                        {status.is_paused ? 'Resume Brain' : 'Pause Brain'}
                    </button>
                </div>
            </div>
        </div>
    );
}
