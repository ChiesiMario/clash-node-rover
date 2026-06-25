export default function NodeRanking({ stats }: any) {
    if (!stats || stats.length === 0) {
        return (
            <div className="card" style={{textAlign:'center', padding:'60px 20px', color:'var(--md-sys-color-on-surface-variant)'}}>
                <span className="material-symbols-outlined" style={{fontSize:'48px', marginBottom:'16px', opacity: 0.5}}>hourglass_empty</span>
                <div className="md3-title-medium">目前尚無節點數據，請等待系統測速完成...</div>
            </div>
        );
    }

    return (
        <div className="table-container">
            <table>
                <thead>
                    <tr>
                        <th style={{width: '80px', textAlign: 'center'}}>排名</th>
                        <th>節點名稱</th>
                        <th style={{textAlign: 'center'}}>綜合分數</th>
                        <th style={{textAlign: 'center'}}>平均延遲</th>
                        <th style={{textAlign: 'center'}}>抖動 (Jitter)</th>
                        <th>所屬群組狀態</th>
                    </tr>
                </thead>
                <tbody>
                    {stats.map((s: any, idx: number) => {
                        const isDead = s.is_dead || false;
                        const scoreStr = isDead ? "失敗" : s.Score;
                        return (
                            <tr key={s.Name} className="node-row">
                                <td style={{textAlign: 'center', fontWeight: '600', color: idx < 3 ? 'var(--md-sys-color-primary)' : 'inherit'}}>
                                    {idx === 0 ? '🏆 1' : idx === 1 ? '🥈 2' : idx === 2 ? '🥉 3' : idx + 1}
                                </td>
                                <td style={{fontWeight: '500', color: isDead ? 'var(--md-sys-color-outline)' : 'inherit'}}>{s.Name}</td>
                                <td style={{textAlign: 'center'}}>
                                    <div className="score-box" style={isDead ? {background:'var(--md-sys-color-error-container)', color:'var(--md-sys-color-on-error-container)'} : {}}>
                                        {scoreStr}
                                    </div>
                                </td>
                                <td style={{textAlign: 'center', color: isDead ? 'var(--md-sys-color-outline)' : 'inherit'}}>{!isDead ? `${s.AvgDelay} ms` : '-'}</td>
                                <td style={{textAlign: 'center', color: isDead ? 'var(--md-sys-color-outline)' : 'inherit'}}>{!isDead ? `${s.Jitter} ms` : '-'}</td>
                                <td>
                                    {s.highest_in_groups && s.highest_in_groups.length > 0 ? (
                                        <div style={{display:'flex', gap:'8px', flexWrap:'wrap'}}>
                                            {s.highest_in_groups.map((g: string) => <span key={g} className="badge primary">{g}</span>)}
                                        </div>
                                    ) : (
                                        <span style={{color: 'var(--md-sys-color-outline)'}}>未被選用</span>
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}
