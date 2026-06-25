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

    const handleRegionChange = (val: string, currentSelected: boolean) => {
        let regexes: string[] = [];
        ['US', 'HK', 'TW', 'JP', 'SG', 'UK'].forEach(r => {
            let isR = rx.includes(REGION_PRESETS[r]);
            if (r === val) isR = !currentSelected; // Toggle
            if (isR) regexes.push(REGION_PRESETS[r]);
        });
        saveFilter(group.name, { keyword_regex: regexes.join('|'), check_chatgpt: isChatGPT, check_gemini: isGemini, check_antigravity: isAntigravity });
    };

    const handleServiceChange = (service: string, currentSelected: boolean) => {
        saveFilter(group.name, { keyword_regex: rx, check_chatgpt: service === 'chatgpt' ? !currentSelected : isChatGPT, check_gemini: service === 'gemini' ? !currentSelected : isGemini, check_antigravity: service === 'antigravity' ? !currentSelected : isAntigravity });
    };

    return (
        <div className="group-card">
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start'}}>
                <div>
                    <div className="md3-title-medium" style={{marginBottom: '2px'}}>{group.name}</div>
                    <div className="md3-body-medium" style={{color: 'var(--md-sys-color-on-surface-variant)'}}>
                        運行中 &bull; {group.all_count} 個節點
                    </div>
                </div>
                <button className={`icon-btn ${group.locked ? 'active' : ''}`} onClick={() => toggleGroupLock(group.name, !group.locked)} title={group.locked ? '解鎖 (自動切換)' : '鎖定 (停止自動切換)'}>
                    <span className={`material-symbols-outlined ${group.locked ? 'fill' : ''}`}>{group.locked ? 'lock' : 'lock_open_right'}</span>
                </button>
            </div>

            <div className="filled-field">
                <div className="field-label">手動選擇節點 (目前: {group.now || '未選擇'})</div>
                <select id={`select-${group.name}`} defaultValue={group.now}>
                    {group.all_nodes && group.all_nodes.map((n: string) => <option key={n} value={n} >{n}</option>)}
                </select>
                <span className="material-symbols-outlined dropdown-icon">arrow_drop_down</span>
                <div className="filled-field-action">
                    <button className="btn tonal" style={{height: '32px', padding: '0 16px', borderRadius: '16px'}} onClick={() => {
                        const sel = document.getElementById(`select-${group.name}`) as HTMLSelectElement;
                        manualSwitch(group.name, sel.value);
                    }}>套用</button>
                </div>
            </div>
            
            <div style={{marginTop: '4px'}}>
                <div className="md3-label-large" style={{marginBottom: '8px', color: 'var(--md-sys-color-on-surface-variant)'}}>地區限制</div>
                <div className="chip-group">
                    {['US', 'HK', 'TW', 'JP', 'SG', 'UK'].map(r => {
                        const isSelected = rx.includes(r + '|');
                        return (
                            <div key={r} className={`chip ${isSelected ? 'selected with-icon' : ''}`} onClick={() => handleRegionChange(r, isSelected)}>
                                {isSelected && <span className="material-symbols-outlined check-icon fill">check</span>}
                                {r === 'US' ? '🇺🇸 美國' : r === 'HK' ? '🇭🇰 香港' : r === 'TW' ? '🇹🇼 台灣' : r === 'JP' ? '🇯🇵 日本' : r === 'SG' ? '🇸🇬 新加坡' : '🇬🇧 英國'}
                            </div>
                        );
                    })}
                </div>

                <div className="md3-label-large" style={{marginTop: '16px', marginBottom: '8px', color: 'var(--md-sys-color-on-surface-variant)'}}>服務驗證</div>
                <div className="chip-group">
                    <div className={`chip ${isChatGPT ? 'selected with-icon' : ''}`} onClick={() => handleServiceChange('chatgpt', isChatGPT)}>
                        {isChatGPT && <span className="material-symbols-outlined check-icon fill">check</span>} 🤖 ChatGPT
                    </div>
                    <div className={`chip ${isGemini ? 'selected with-icon' : ''}`} onClick={() => handleServiceChange('gemini', isGemini)}>
                        {isGemini && <span className="material-symbols-outlined check-icon fill">check</span>} ✨ Gemini
                    </div>
                    <div className={`chip ${isAntigravity ? 'selected with-icon' : ''}`} onClick={() => handleServiceChange('antigravity', isAntigravity)}>
                        {isAntigravity && <span className="material-symbols-outlined check-icon fill">check</span>} 🚀 Antigravity
                    </div>
                </div>
            </div>
        </div>
    );
}
