import { useState, useEffect, useRef } from 'react';

export interface LogEntry {
    time: string;
    level: string;
    message: string;
}

export function useWebSocket(onRefresh: () => void) {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const wsRef = useRef<WebSocket | null>(null);

    const onRefreshRef = useRef(onRefresh);

    useEffect(() => {
        onRefreshRef.current = onRefresh;
    }, [onRefresh]);

    useEffect(() => {
        const connect = () => {
            const wsUrl = (window.location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/api/ws';
            const ws = new WebSocket(wsUrl);

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.type === 'refresh') {
                        if (onRefreshRef.current) {
                            onRefreshRef.current();
                        }
                    } else if (msg.type === 'log') {
                        setLogs(prev => [...prev.slice(-199), msg.entry]);
                    } else if (msg.type === 'log_history') {
                        setLogs(msg.history || []);
                    }
                } catch (e) {
                    console.error('WS parse error', e);
                }
            };

            ws.onclose = () => {
                setTimeout(connect, 3000);
            };

            wsRef.current = ws;
        };

        connect();

        return () => {
            if (wsRef.current) {
                wsRef.current.onclose = null;
                wsRef.current.close();
            }
        };
    }, []);

    return { logs };
}
