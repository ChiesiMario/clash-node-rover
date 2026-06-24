import type { NodeStat } from '../hooks/useApi';

export default function NodeRanking({ stats }: { stats: NodeStat[] }) {
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
                                <tr key={node.Name} className="node-row" style={{borderBottom:'1px solid var(--md-sys-color-outline-variant)'}}>
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
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
