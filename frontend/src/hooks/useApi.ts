import { useState } from 'react';
import { GetStats, GetStatus, GetGroups, ManualTrigger, TogglePause, SwitchNode, SetGroupLocked } from '../../wailsjs/go/main/App';
import { GetHistory } from '../../wailsjs/go/main/App';

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
    is_configured?: boolean;
    api_connected?: boolean;
    is_ready?: boolean;
}

export function useApi() {
    const [stats, setStats] = useState<NodeStat[]>([]);
    const [status, setStatus] = useState<Status>({ is_running: false, is_paused: false });
    const [groups, setGroups] = useState<any[]>([]);

    const fetchStats = async () => {
        try {
            const data = await GetStats();
            // @ts-ignore
            setStats(data || []);
        } catch (e) {
            console.error('Failed to fetch stats', e);
        }
    };

    const fetchStatus = async () => {
        try {
            const data = await GetStatus();
            // @ts-ignore
            setStatus(data);
        } catch (e) {
            console.error('Failed to fetch status', e);
        }
    };

    const fetchGroups = async () => {
        try {
            const data = await GetGroups();
            setGroups(data || []);
        } catch (e) {
            console.error('Failed to fetch groups', e);
        }
    };

    const triggerTest = async () => {
        await ManualTrigger();
        fetchStatus();
    };

    const togglePause = async () => {
        await TogglePause();
        fetchStatus();
    };

    const manualSwitch = async (group: string, node: string) => {
        await SwitchNode(group, node);
        fetchGroups();
    };

    const toggleGroupLock = async (group: string, locked: boolean) => {
        await SetGroupLocked(group, locked);
        fetchGroups();
    };

    const saveFilter = async (_group: string, _filterData: any) => {
        // ...
    };

    const fetchNodeHistory = async (nodeName: string) => {
        try {
            const res = await GetHistory(nodeName);
            return res;
        } catch (e) {
            console.error('Failed to fetch node history', e);
            return { ping: [], browser: [] };
        }
    };

    return { stats, status, groups, fetchStats, fetchStatus, fetchGroups, triggerTest, togglePause, manualSwitch, toggleGroupLock, saveFilter, fetchNodeHistory };
}
