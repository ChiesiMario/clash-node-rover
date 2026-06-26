import { useState, useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

export default function DevConsole() {
    const [isOpen, setIsOpen] = useState(false);
    const { logs } = useWebSocket(() => {});

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.ctrlKey && e.key === '`') {
                setIsOpen(prev => !prev);
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, []);

    if (!isOpen) return null;

    return (
        <div style={{
            position: 'fixed',
            bottom: 0,
            left: 0,
            right: 0,
            height: '40vh',
            backgroundColor: 'rgba(15, 23, 42, 0.95)',
            backdropFilter: 'blur(10px)',
            color: '#0f0',
            fontFamily: 'monospace',
            zIndex: 9999,
            padding: '16px',
            overflowY: 'auto',
            borderTop: '1px solid #334155',
            boxShadow: '0 -4px 20px rgba(0,0,0,0.5)'
        }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', borderBottom: '1px solid #334155', paddingBottom: '8px' }}>
                <span style={{color: '#94a3b8', fontSize: '12px'}}>Developer Console (Ctrl + `)</span>
                <button onClick={() => setIsOpen(false)} style={{background: 'none', border: 'none', color: '#ef4444', cursor: 'pointer', fontSize: '16px'}}>✖</button>
            </div>
            <div className="console-content" style={{fontSize: '12px', lineHeight: '1.4'}}>
                {logs.map((log, i) => (
                    <div key={i} style={{marginBottom: '4px', display: 'flex', gap: '8px'}}>
                        <span style={{color: '#64748b', minWidth: '80px'}}>[{log.time}]</span>
                        <span style={{
                            color: log.level === 'error' ? '#ef4444' : 
                                   log.level === 'warning' ? '#f59e0b' : 
                                   log.level === 'success' ? '#22c55e' : '#38bdf8',
                            minWidth: '40px'
                        }}>
                            {log.level === 'success' ? 'OK' : log.level === 'warning' ? 'WARN' : log.level === 'error' ? 'FAIL' : 'INFO'}
                        </span>
                        <span style={{color: '#e2e8f0', wordBreak: 'break-all'}}>{log.message.replace(/^[💡✅⚠️❌] /, '')}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}
