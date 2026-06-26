import { useState, useEffect, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetLogHistory } from '../../wailsjs/go/main/App';

export interface LogEntry {
    time: string;
    level: string;
    message: string;
}

export function useWebSocket(onRefresh: () => void) {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const onRefreshRef = useRef(onRefresh);

    useEffect(() => {
        onRefreshRef.current = onRefresh;
    }, [onRefresh]);

    useEffect(() => {
        // Load initial history
        GetLogHistory().then((history: any) => {
            if (history) setLogs(history);
        });

        // Listen for new logs
        EventsOn('log', (entry: any) => {
            setLogs(prev => [...prev.slice(-199), entry]);
        });

        // Listen for refresh
        EventsOn('refresh', () => {
            if (onRefreshRef.current) {
                onRefreshRef.current();
            }
        });

        return () => {
            EventsOff('log');
            EventsOff('refresh');
        };
    }, []);

    return { logs };
}
