

export default function NodeRanking({ stats }: any) {
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
                        return (
                            <tr key={s.Name}>
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
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}
