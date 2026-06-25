import os

content = """import React, { useState, useEffect } from 'react';
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
        <div className="hig-table-container">
            <table>
                <thead>
                    <tr>
                        <th style={{width: '60px', textAlign: 'center'}}>排名</th>
                        <th>節點名稱</th>
                        <th style={{textAlign: 'center'}}>綜合評分</th>
                        <th style={{textAlign: 'center'}}>連線延遲</th>
                        <th style={{textAlign: 'center'}}>網路抖動</th>
                        <th>分發狀態</th>
                    </tr>
                </thead>
                <tbody>
                    {stats.map((s: any, idx: number) => {
                        const isDead = s.is_dead || false;
                        const scoreStr = isDead ? "FAIL" : s.Score;
                        const isExpanded = expandedNode === s.Name;
                        
                        return (
                            <React.Fragment key={s.Name}>
                                <tr 
                                    onClick={() => handleRowClick(s.Name)}
                                    style={{
                                        cursor: 'pointer',
                                        backgroundColor: isExpanded ? 'var(--hig-bg-tertiary)' : 'transparent'
                                    }}
                                >
                                    <td style={{textAlign: 'center', fontWeight: '600', color: idx < 3 ? 'var(--hig-system-blue)' : 'inherit'}}>
                                        {idx + 1}
                                    </td>
                                    <td style={{fontWeight: '500', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{s.Name}</td>
                                    <td style={{textAlign: 'center'}}>
                                        <div className={`hig-badge ${isDead ? 'red' : 'blue'}`}>
                                            {scoreStr}
                                        </div>
                                    </td>
                                    <td style={{textAlign: 'center', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{!isDead ? `${s.AvgDelay} ms` : '-'}</td>
                                    <td style={{textAlign: 'center', color: isDead ? 'var(--hig-text-secondary)' : 'inherit'}}>{!isDead ? `${s.Jitter} ms` : '-'}</td>
                                    <td>
                                        {s.highest_in_groups && s.highest_in_groups.length > 0 ? (
                                            <div style={{display:'flex', gap:'4px', flexWrap:'wrap'}}>
                                                {s.highest_in_groups.map((g: string) => <span key={g} className="hig-badge green">{g}</span>)}
                                            </div>
                                        ) : (
                                            <span style={{color: 'var(--hig-text-secondary)', fontSize: '13px'}}>閒置中</span>
                                        )}
                                    </td>
                                </tr>
                                {isExpanded && (
                                    <tr>
                                        <td colSpan={6} style={{padding: '24px', backgroundColor: 'var(--hig-bg-secondary)', borderBottom: '1px solid var(--hig-separator)'}}>
                                            <div style={{height: '300px', width: '100%'}}>
                                                {isLoading ? (
                                                    <div style={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--hig-text-secondary)'}}>
                                                        <span className="material-symbols-outlined spin" style={{marginRight: '8px'}}>refresh</span>
                                                        載入歷史數據中...
                                                    </div>
                                                ) : chartData.length === 0 ? (
                                                    <div style={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--hig-text-secondary)'}}>
                                                        尚無足夠的歷史數據可供繪製
                                                    </div>
                                                ) : (
                                                    <ResponsiveContainer width="100%" height="100%">
                                                        <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                                                            <CartesianGrid strokeDasharray="3 3" stroke="var(--hig-separator)" vertical={false} />
                                                            <XAxis dataKey="time" stroke="var(--hig-text-secondary)" fontSize={12} tickMargin={10} axisLine={false} tickLine={false} />
                                                            
                                                            {/* Y-Axis for Ping */}
                                                            <YAxis yAxisId="left" stroke="var(--hig-system-blue)" fontSize={12} tickMargin={10} axisLine={false} tickLine={false} />
                                                            
                                                            {/* Y-Axis for Browser */}
                                                            <YAxis yAxisId="right" orientation="right" stroke="var(--hig-system-green)" fontSize={12} tickMargin={10} axisLine={false} tickLine={false} />
                                                            
                                                            <Tooltip 
                                                                contentStyle={{backgroundColor: 'var(--hig-bg-primary)', borderRadius: '8px', border: '1px solid var(--hig-separator)', boxShadow: '0 4px 12px rgba(0,0,0,0.1)', color: 'var(--hig-text-primary)'}}
                                                                itemStyle={{fontWeight: 500}}
                                                            />
                                                            <Legend wrapperStyle={{paddingTop: '20px'}} />
                                                            <Line yAxisId="left" type="monotone" dataKey="Ping" stroke="var(--hig-system-blue)" strokeWidth={3} dot={false} activeDot={{ r: 6 }} name="Ping 延遲 (ms)" />
                                                            <Line yAxisId="right" type="monotone" dataKey="Browser" stroke="var(--hig-system-green)" strokeWidth={3} dot={false} activeDot={{ r: 6 }} name="網頁加載 (ms)" />
                                                        </LineChart>
                                                    </ResponsiveContainer>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                )}
                            </React.Fragment>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}
"""

with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f:
    f.write(content)
