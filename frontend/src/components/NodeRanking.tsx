import React, { useState } from 'react';
import { useApi } from '../hooks/useApi';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts';

export default function NodeRanking({ stats }: any) {
    const { fetchNodeHistory } = useApi();
    const [expandedNode, setExpandedNode] = useState<string | null>(null);
    const [chartData, setChartData] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(false);


    const handleRowClick = async (nodeName: string) => {
        if (expandedNode === nodeName) {
            setExpandedNode(null);
            setChartData([]);
            return;
        }

        setExpandedNode(nodeName);
        setIsLoading(true);
        const history = await fetchNodeHistory(nodeName);
        
        // Merge ping and browser history by timestamp (nearest)
        // For simplicity, we just format the timestamp to HH:mm
        const merged: Record<string, any> = {};
        
        if (history.ping) {
            history.ping.forEach((p: any) => {
                const d = new Date(p.Timestamp * 1000);
                const time = `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
                merged[time] = { ...merged[time], time, ping: p.Delay };
            });
        }
        
        if (history.browser) {
            history.browser.forEach((b: any) => {
                const d = new Date(b.Timestamp * 1000);
                const time = `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
                merged[time] = { ...merged[time], time, browser: b.LoadTimeMs };
            });
        }

        // Sort by time (this is a simplified string sort, works for same day, for across 24h it might wrap, but good enough for UI)
        // Better: store raw timestamp, sort, then format
        const rawData: any[] = [];
        if (history.ping) {
            history.ping.forEach((p: any) => {
                rawData.push({ ts: p.Timestamp * 1000, ping: p.Delay });
            });
        }
        if (history.browser) {
            history.browser.forEach((b: any) => {
                const existing = rawData.find(r => Math.abs(r.ts - b.Timestamp * 1000) < 60000);
                if (existing) {
                    existing.browser = b.LoadTimeMs;
                } else {
                    rawData.push({ ts: b.Timestamp * 1000, browser: b.LoadTimeMs });
                }
            });
        }
        
        rawData.sort((a, b) => a.ts - b.ts);
        
        const finalData = rawData.map(d => {
            const date = new Date(d.ts);
            return {
                time: `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`,
                Ping: d.ping,
                Browser: d.browser
            };
        });

        setChartData(finalData);
        setIsLoading(false);
    };

    if (!stats || stats.length === 0) {
        return (
            <div className="hig-card" style={{textAlign:'center', padding:'60px 20px', color:'var(--hig-text-secondary)'}}>
                <span className="material-symbols-outlined" style={{fontSize:'48px', marginBottom:'16px', opacity: 0.5}}>hourglass_empty</span>
                <div className="hig-headline">尚無節點連線數據</div>
            </div>
        );
    }

    return (
        <div className="apple-list-group">
            {stats.map((s: any, idx: number) => {
                const isDead = s.is_dead || false;
                const scoreStr = isDead ? (s.backoff_remaining > 0 ? `退避 (${s.backoff_remaining})` : "失敗") : s.Score;
                const isExpanded = expandedNode === s.Name;
                
                return (
                    <React.Fragment key={s.Name}>
                        <div 
                            className="apple-list-item"
                            onClick={() => handleRowClick(s.Name)}
                            style={{
                                cursor: 'pointer',
                                backgroundColor: isExpanded ? 'var(--hig-fill-secondary)' : 'transparent',
                                display: 'flex',
                                alignItems: 'center',
                                padding: '16px'
                            }}
                        >
                            <div style={{ flex: 1, display: 'flex', alignItems: 'center', gap: '16px' }}>
                                <div style={{ 
                                    width: '28px', height: '28px', 
                                    borderRadius: '14px', 
                                    backgroundColor: idx < 3 ? 'var(--hig-system-blue)' : 'var(--hig-fill-primary)',
                                    color: idx < 3 ? '#fff' : 'var(--text-secondary)',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    fontSize: '13px', fontWeight: 'bold', flexShrink: 0
                                }}>
                                    {idx + 1}
                                </div>
                                <div>
                                    <div style={{ fontSize: '17px', fontWeight: '500', color: isDead ? 'var(--hig-text-secondary)' : 'var(--hig-text-primary)' }}>
                                        {s.Name}
                                    </div>
                                    <div style={{ display: 'flex', gap: '8px', marginTop: '4px', flexWrap: 'wrap' }}>
                                        {s.highest_in_groups && s.highest_in_groups.length > 0 ? (
                                            s.highest_in_groups.map((g: string) => <span key={g} className="hig-badge green" style={{ padding: '2px 6px', fontSize: '10px' }}>{g}</span>)
                                        ) : (
                                            <span style={{ fontSize: '13px', color: 'var(--hig-text-secondary)' }}>閒置中</span>
                                        )}
                                    </div>
                                </div>
                            </div>

                            <div style={{ display: 'flex', gap: '24px', alignItems: 'center', textAlign: 'right' }}>
                                <div style={{ display: 'flex', flexDirection: 'column' }}>
                                    <span style={{ fontSize: '15px', color: isDead ? 'var(--hig-text-secondary)' : 'var(--hig-text-primary)' }}>
                                        {scoreStr}
                                    </span>
                                    <span style={{ fontSize: '12px', color: 'var(--hig-text-secondary)' }}>評分</span>
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', width: '50px' }}>
                                    <span style={{ fontSize: '15px', color: isDead ? 'var(--hig-text-secondary)' : 'var(--hig-text-primary)' }}>
                                        {!isDead ? `${s.Jitter}ms` : '-'}
                                    </span>
                                    <span style={{ fontSize: '12px', color: 'var(--hig-text-secondary)' }}>抖動</span>
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', width: '50px' }}>
                                    <span style={{ fontSize: '15px', color: isDead ? 'var(--hig-text-secondary)' : 'var(--hig-text-primary)' }}>
                                        {!isDead ? `${s.AvgDelay}ms` : '-'}
                                    </span>
                                    <span style={{ fontSize: '12px', color: 'var(--hig-text-secondary)' }}>延遲</span>
                                </div>
                                <span className="material-symbols-outlined" style={{ color: 'var(--hig-text-secondary)', transform: isExpanded ? 'rotate(90deg)' : 'none', transition: 'transform 0.2s' }}>
                                    chevron_right
                                </span>
                            </div>
                        </div>
                        {isExpanded && (
                            <div style={{ padding: '24px', backgroundColor: 'var(--hig-bg-secondary)', borderBottom: '1px solid rgba(0,0,0,0.05)' }}>
                                <div style={{height: '300px', width: '100%'}}>
                                    {isLoading ? (
                                        <div style={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--hig-text-secondary)'}}>
                                            <span className="material-symbols-outlined spin" style={{marginRight: '8px'}}>progress_activity</span>
                                            讀取歷史數據中...
                                        </div>
                                    ) : chartData.length > 0 ? (
                                        <ResponsiveContainer width="100%" height="100%">
                                            <LineChart data={chartData} margin={{top: 5, right: 20, bottom: 5, left: 0}}>
                                                <CartesianGrid strokeDasharray="3 3" stroke="var(--hig-separator)" vertical={false} />
                                                <XAxis dataKey="time" stroke="var(--hig-text-secondary)" fontSize={12} tickMargin={10} />
                                                <YAxis stroke="var(--hig-text-secondary)" fontSize={12} width={40} />
                                                <Tooltip 
                                                    contentStyle={{backgroundColor: 'var(--hig-glass-bg)', borderRadius: '12px', border: '1px solid var(--hig-separator)', backdropFilter: 'blur(20px)'}}
                                                    itemStyle={{color: 'var(--hig-text-primary)'}}
                                                />
                                                <Legend />
                                                <Line type="monotone" dataKey="Ping" stroke="var(--hig-system-blue)" strokeWidth={2} dot={{r: 3}} activeDot={{r: 6}} />
                                                <Line type="monotone" dataKey="Browser" stroke="var(--hig-system-green)" strokeWidth={2} dot={{r: 3}} />
                                            </LineChart>
                                        </ResponsiveContainer>
                                    ) : (
                                        <div style={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--hig-text-secondary)'}}>
                                            無歷史數據
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}
                    </React.Fragment>
                );
            })}
        </div>
    );
}
