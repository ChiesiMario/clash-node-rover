import { useState } from 'react';

export interface NodeStat {
    Name: string;
    AvgDelay: number;
    Jitter: number;
    Score: number;
    provider: string;
    highest_in_groups: string[];
    backoff_remaining: number;
    browser_backoff_remaining: Record<string, number>;
    is_dead: boolean;
}

export interface Status {
    is_running: boolean;
    is_paused: boolean;
}

export function useApi() {
    const [stats, setStats] = useState<NodeStat[]>([]);
    const [status, setStatus] = useState<Status>({ is_running: false, is_paused: false });
    const [groups, setGroups] = useState<any[]>([]);

    const fetchStats = async () => {
        try {
            const res = await fetch('/api/stats');
            const data = await res.json();
            setStats(data || []);
        } catch (e) {
            console.error('Failed to fetch stats', e);
        }
    };

    const fetchStatus = async () => {
        try {
            const res = await fetch('/api/status');
            const data = await res.json();
            setStatus(data);
        } catch (e) {
            console.error('Failed to fetch status', e);
        }
    };

    const fetchGroups = async () => {
        try {
            const res = await fetch('/api/groups');
            const data = await res.json();
            setGroups(data || []);
        } catch (e) {
            console.error('Failed to fetch groups', e);
        }
    };

    const triggerTest = async () => {
        await fetch('/api/trigger', { method: 'POST' });
        fetchStatus();
    };

    const togglePause = async () => {
        await fetch('/api/pause', { method: 'POST' });
        fetchStatus();
    };

    const manualSwitch = async (group: string, node: string) => {
        await fetch('/api/switch', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({group, node})
        });
        fetchGroups();
    };

    const toggleGroupLock = async (group: string, locked: boolean) => {
        await fetch('/api/groups/lock', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({group, locked})
        });
        fetchGroups();
    };

    const saveFilter = async (group: string, filterData: any) => {
        await fetch('/api/groups/filter?group=' + encodeURIComponent(group), {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(filterData)
        });
    };

    const fetchNodeHistory = async (nodeName: string) => {
        try {
            const res = await fetch('/api/history?node=' + encodeURIComponent(nodeName));
            return await res.json();
        } catch (e) {
            console.error('Failed to fetch node history', e);
            return { ping: [], browser: [] };
        }
    };

    return { stats, status, groups, fetchStats, fetchStatus, fetchGroups, triggerTest, togglePause, manualSwitch, toggleGroupLock, saveFilter, fetchNodeHistory };
}
