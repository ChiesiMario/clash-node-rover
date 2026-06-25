
const REGION_PRESETS: Record<string, string> = {
    'US': 'US|United States|us|美國|美国',
    'HK': 'HK|Hong Kong|香港',
    'TW': 'TW|Taiwan|台灣|台湾|臺',
    'JP': 'JP|Japan|日本',
    'SG': 'SG|Singapore|新加坡|狮城',
    'UK': 'UK|United Kingdom|英國|英国'
};

export default function GroupCard({ group, manualSwitch, toggleGroupLock, saveFilter }: any) {
    const rx = group.filter?.keyword_regex || "";
    const isChatGPT = group.filter?.check_chatgpt || false;
    const isGemini = group.filter?.check_gemini || false;
    const isAntigravity = group.filter?.check_antigravity || false;

    const handleRegionChange = (val: string, checked: boolean) => {
        let regexes: string[] = [];
        ['US', 'HK', 'TW', 'JP', 'SG', 'UK'].forEach(r => {
            let isR = rx.includes(REGION_PRESETS[r]);
            if (r === val) isR = checked;
            if (isR) regexes.push(REGION_PRESETS[r]);
        });
        saveFilter(group.name, {
            keyword_regex: regexes.join('|'),
            check_chatgpt: isChatGPT,
            check_gemini: isGemini,
            check_antigravity: isAntigravity
        });
    };

    const handleServiceChange = (service: string, checked: boolean) => {
        saveFilter(group.name, {
            keyword_regex: rx,
            check_chatgpt: service === 'chatgpt' ? checked : isChatGPT,
            check_gemini: service === 'gemini' ? checked : isGemini,
            check_antigravity: service === 'antigravity' ? checked : isAntigravity
        });
    };

    return (
        <div className="group-card">
            <div className="group-header md3-title-medium" style={{display:'flex', justifyContent:'space-between', alignItems:'center'}}>
                <span>{group.name}</span>
                <button className={`btn icon-btn ${group.locked ? '' : 'secondary'}`} onClick={() => toggleGroupLock(group.name, !group.locked)} title={group.locked ? '點擊解鎖 (恢復自動切換)' : '點擊鎖定 (停止自動切換)'} style={{width:'32px', height:'32px'}}>
                    <span className="material-symbols-outlined" style={{fontSize:'16px'}}>{group.locked ? 'lock' : 'lock_open'}</span>
                </button>
            </div>
            <div className="group-now md3-title-large" style={{color:'var(--md-sys-color-primary)'}}>{group.now || '未選擇'}</div>
            {group.provider && <div className="badge primary" style={{alignSelf: 'flex-start'}}><span className="material-symbols-outlined" style={{fontSize:'14px'}}>corporate_fare</span> {group.provider}</div>}
            <div className="md3-body-medium" style={{color: 'var(--md-sys-color-on-surface-variant)', marginTop:'8px', marginBottom:'8px'}}>運行中 &bull; 共 {group.all_count} 個節點</div>
            <div style={{display: 'flex', gap: '8px'}}>
                <select className="md3-select" id={`select-${group.name}`} defaultValue={group.now} style={{flex:1, height: '40px'}}>
                    {group.all_nodes.map((n: string) => <option key={n} value={n} >{n}</option>)}
                </select>
                <button onClick={() => {
                    const sel = document.getElementById(`select-${group.name}`) as HTMLSelectElement;
                    manualSwitch(group.name, sel.value);
                }} className="btn" style={{padding: '0 20px', whiteSpace: 'nowrap', height: '40px'}}>切換</button>
            </div>
            
            <div style={{marginTop: '16px', paddingTop: '12px', borderTop: '1px solid var(--md-sys-color-outline-variant)'}}>
                <div style={{fontSize: '13px', fontWeight: 500, marginBottom: '8px', color: 'var(--md-sys-color-primary)'}}>節點地區篩選</div>
                <div style={{display: 'flex', gap: '16px', rowGap: '12px', flexWrap: 'wrap', fontSize: '13px'}}>
                    {['US', 'HK', 'TW', 'JP', 'SG', 'UK'].map(r => (
                        <label key={r} className="md3-checkbox-label">
                            <input type="checkbox" checked={rx.includes(r + '|')} onChange={(e) => handleRegionChange(r, e.target.checked)} /> 
                            {r === 'US' ? '🇺🇸 美國' : r === 'HK' ? '🇭🇰 香港' : r === 'TW' ? '🇹🇼 台灣' : r === 'JP' ? '🇯🇵 日本' : r === 'SG' ? '🇸🇬 新加坡' : '🇬🇧 英國'}
                        </label>
                    ))}
                </div>
                <div style={{fontSize: '13px', fontWeight: 500, marginTop: '12px', marginBottom: '8px', color: 'var(--md-sys-color-primary)'}}>必備服務驗證</div>
                <div style={{display: 'flex', gap: '16px', rowGap: '12px', fontSize: '13px'}}>
                    <label className="md3-checkbox-label"><input type="checkbox" checked={isChatGPT} onChange={(e) => handleServiceChange('chatgpt', e.target.checked)} /> 🤖 ChatGPT</label>
                    <label className="md3-checkbox-label"><input type="checkbox" checked={isGemini} onChange={(e) => handleServiceChange('gemini', e.target.checked)} /> ✨ Gemini</label>
                    <label className="md3-checkbox-label"><input type="checkbox" checked={isAntigravity} onChange={(e) => handleServiceChange('antigravity', e.target.checked)} /> 🚀 Antigravity</label>
                </div>
            </div>
        </div>
    );
}
