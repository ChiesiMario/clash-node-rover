import re

# 1. Fix App.tsx Logo
with open('frontend/src/App.tsx', 'r', encoding='utf-8') as f:
    app = f.read()

app = app.replace('<img id="app-logo-img" src="/logo.png" alt="Logo" width="28" height="28" />', 
                  '<span className="material-symbols-outlined" style={{color: "var(--md-sys-color-primary)", fontSize: "28px"}}>rocket_launch</span>')
with open('frontend/src/App.tsx', 'w', encoding='utf-8') as f:
    f.write(app)


# 2. Fix GroupCard.tsx Filtering Logic and <select> selected attribute
with open('frontend/src/components/GroupCard.tsx', 'r', encoding='utf-8') as f:
    gc = f.read()

gc = gc.replace("let isR = rx.includes(r + '|');", "let isR = rx.includes(REGION_PRESETS[r]);")
gc = gc.replace("selected={n === group.now}", "")
gc = gc.replace('<select id={`select-${group.name}`}', '<select id={`select-${group.name}`} defaultValue={group.now}')

with open('frontend/src/components/GroupCard.tsx', 'w', encoding='utf-8') as f:
    f.write(gc)


# 3. Add Chart implementation back to NodeRanking.tsx
chart_code = """import { useState, useEffect, useRef } from 'react';
import type { NodeStat } from '../hooks/useApi';
import Chart from 'chart.js/auto';

function ChartRow({ nodeName, safeId, avgDelay, jitter, score }: { nodeName: string, safeId: string, avgDelay: number, jitter: number, score: number }) {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const chartRef = useRef<Chart | null>(null);

    useEffect(() => {
        let isMounted = true;
        fetch('/api/history?node=' + encodeURIComponent(nodeName))
            .then(res => res.json())
            .then(dataRaw => {
                if (!isMounted || !dataRaw || !dataRaw.ping || dataRaw.ping.length === 0) return;
                
                const labels = dataRaw.ping.map((h: any) => {
                    const d = new Date(h.Timestamp * 1000);
                    return d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0');
                });
                const pingData = dataRaw.ping.map((h: any) => h.Delay);
                const browserData = dataRaw.ping.map((p: any) => {
                    const b = dataRaw.browser ? dataRaw.browser.find((b: any) => Math.abs(b.Timestamp - p.Timestamp) < 300) : null;
                    return b ? b.LoadTimeMs : null;
                });

                if (canvasRef.current) {
                    Chart.defaults.color = document.documentElement.classList.contains('light-theme') ? '#44474e' : '#c4c6d0';
                    Chart.defaults.font.family = "'Roboto', sans-serif";

                    chartRef.current = new Chart(canvasRef.current, {
                        type: 'line',
                        data: {
                            labels: labels,
                            datasets: [
                                { label: 'Ping (ms)', data: pingData, borderColor: '#a8c7fa', backgroundColor: 'rgba(168,199,250,0.1)', fill: true, tension: 0.4, yAxisID: 'y' },
                                { label: 'Browser (ms)', data: browserData, borderColor: '#6dd58c', fill: false, tension: 0.1, yAxisID: 'y2' }
                            ]
                        },
                        options: { responsive: true, maintainAspectRatio: false, scales: { y: { position: 'left' }, y2: { position: 'right', grid: { drawOnChartArea: false } } } }
                    });
                }
            });

        return () => {
            isMounted = false;
            if (chartRef.current) {
                chartRef.current.destroy();
            }
        };
    }, [nodeName]);

    return (
        <tr className="chart-row expanded-row">
            <td colSpan={6}>
                <div style={{padding: '16px'}}>
                    <div style={{display:'flex', gap:'16px', padding:'16px', marginBottom:'16px', background:'var(--md-sys-color-surface-container)', borderRadius:'16px'}}>
                        <div style={{flex:1}}><div style={{fontSize:'12px', color:'var(--md-sys-color-on-surface-variant)'}}>平均延遲</div><div style={{fontSize:'18px'}}>{avgDelay} ms</div></div>
                        <div style={{flex:1}}><div style={{fontSize:'12px', color:'var(--md-sys-color-on-surface-variant)'}}>抖動 (Jitter)</div><div style={{fontSize:'18px'}}>{jitter} ms</div></div>
                        <div style={{flex:1}}><div style={{fontSize:'12px', color:'var(--md-sys-color-on-surface-variant)'}}>綜合分數</div><div style={{fontSize:'18px'}}>{score}</div></div>
                    </div>
                    <div className="chart-container" style={{height: '300px', width: '100%'}}>
                        <canvas ref={canvasRef}></canvas>
                    </div>
                </div>
            </td>
        </tr>
    );
}

export default function NodeRanking({ stats }: { stats: NodeStat[] }) {
    const [expandedNode, setExpandedNode] = useState<string | null>(null);

    return (
        <div className="card" style={{padding: '0'}}>
            <div className="card-header" style={{padding: '24px 24px 0 24px', display:'flex', alignItems:'center', gap:'12px', marginBottom:'16px'}}>
                <span className="material-symbols-outlined" style={{fontSize:'28px', color:'var(--md-sys-color-primary)'}}>bar_chart</span>
                <span style={{fontSize:'20px', fontWeight:500}}>節點排行榜 (即時)</span>
            </div>
            <div style={{overflowX:'auto', padding:'0 24px 24px 24px'}}>
                <table style={{width:'100%', borderCollapse:'collapse', minWidth:'800px', fontSize:'14px'}}>
                    <thead>
                        <tr style={{textAlign:'left', color:'var(--md-sys-color-on-surface-variant)', borderBottom:'1px solid var(--md-sys-color-outline-variant)'}}>
                            <th style={{padding:'16px 8px'}}>排名</th>
                            <th style={{padding:'16px 8px'}}>節點名稱</th>
                            <th style={{padding:'16px 8px'}}>綜合分數 (越低越好)</th>
                            <th style={{padding:'16px 8px'}}>平均延遲</th>
                            <th style={{padding:'16px 8px'}}>抖動 (Jitter)</th>
                            <th style={{padding:'16px 8px'}}>群組狀態</th>
                        </tr>
                    </thead>
                    <tbody>
                        {stats.map((node, index) => {
                            let scoreHtml = node.is_dead ? 
                                <span className="score-box" style={{background:'var(--md-sys-color-error-container)', color:'var(--md-sys-color-on-error-container)'}}>失敗</span> : 
                                <span className="score-box">{node.Score}</span>;

                            return (
                                <React.Fragment key={node.Name}>
                                    <tr className={`node-row ${expandedNode === node.Name ? 'expanded-row' : ''}`} style={{borderBottom:'1px solid var(--md-sys-color-outline-variant)', cursor: 'pointer'}} onClick={() => setExpandedNode(expandedNode === node.Name ? null : node.Name)}>
                                        <td style={{padding:'16px 8px'}}>#{index + 1}</td>
                                        <td style={{padding:'16px 8px', fontWeight:500, color: node.is_dead ? 'var(--md-sys-color-outline)' : 'inherit'}}>
                                            {node.Name}
                                            {node.provider && <><br/><div className="badge primary" style={{marginTop:'4px', fontSize:'10px'}}><span className="material-symbols-outlined" style={{fontSize:'12px'}}>corporate_fare</span> {node.provider}</div></>}
                                        </td>
                                        <td style={{padding:'16px 8px'}}>{scoreHtml}</td>
                                        <td style={{padding:'16px 8px', color: node.is_dead ? 'var(--md-sys-color-outline)' : 'inherit'}}>{node.is_dead ? 'N/A' : `${node.AvgDelay} ms`}</td>
                                        <td style={{padding:'16px 8px', color: node.is_dead ? 'var(--md-sys-color-outline)' : 'inherit'}}>{node.is_dead ? 'N/A' : `${node.Jitter} ms`}</td>
                                        <td style={{padding:'16px 8px'}}>
                                            {node.highest_in_groups?.map(g => <div key={g} className="badge success" style={{marginTop:'4px', marginLeft:'4px', fontSize:'10px'}}><span className="material-symbols-outlined" style={{fontSize:'12px'}}>workspace_premium</span> {g}</div>)}
                                            {node.backoff_remaining > 0 && <div className="badge error" style={{marginTop:'4px', marginLeft:'4px', fontSize:'10px'}}><span className="material-symbols-outlined" style={{fontSize:'12px'}}>timer_off</span> Ping 退避 ({node.backoff_remaining} 輪)</div>}
                                            {node.browser_backoff_remaining && Object.entries(node.browser_backoff_remaining).map(([url, rem]) => rem > 0 && (
                                                <div key={url} className="badge warning" style={{marginTop:'4px', marginLeft:'4px', fontSize:'10px'}}><span className="material-symbols-outlined" style={{fontSize:'12px'}}>web_asset_off</span> {url.includes('chatgpt') ? 'ChatGPT' : url.includes('gemini') ? 'Gemini' : url.includes('generative') ? 'Antigravity' : '網頁'} 退避 ({rem} 輪)</div>
                                            ))}
                                        </td>
                                    </tr>
                                    {expandedNode === node.Name && !node.is_dead && (
                                        <ChartRow nodeName={node.Name} safeId={btoa(encodeURIComponent(node.Name)).replace(/=/g, '')} avgDelay={node.AvgDelay} jitter={node.Jitter} score={node.Score} />
                                    )}
                                </React.Fragment>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
"""
with open('frontend/src/components/NodeRanking.tsx', 'w', encoding='utf-8') as f:
    f.write(chart_code)
