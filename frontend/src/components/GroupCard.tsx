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
            if (r === val) isR = !currentSelected;
            if (isR) regexes.push(REGION_PRESETS[r]);
        });
        saveFilter(group.name, { keyword_regex: regexes.join('|'), check_chatgpt: isChatGPT, check_gemini: isGemini, check_antigravity: isAntigravity });
    };

    const handleServiceChange = (service: string, currentSelected: boolean) => {
        saveFilter(group.name, { keyword_regex: rx, check_chatgpt: service === 'chatgpt' ? !currentSelected : isChatGPT, check_gemini: service === 'gemini' ? !currentSelected : isGemini, check_antigravity: service === 'antigravity' ? !currentSelected : isAntigravity });
    };

    return (
        <div className="hig-card" style={{display: 'flex', flexDirection: 'column', gap: '20px'}}>
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'flex-start'}}>
                <div>
                    <div className="hig-title-3" style={{marginBottom: '2px'}}>{group.name}</div>
                    <div className="hig-footnote" style={{color: 'var(--hig-text-secondary)'}}>
                        包含 {group.all_count} 個節點
                    </div>
                </div>
                <button className="icon-btn" onClick={() => toggleGroupLock(group.name, !group.locked)} title={group.locked ? '解除鎖定' : '鎖定群組'}>
                    <span className={`material-symbols-outlined ${group.locked ? 'fill' : ''}`} style={{color: group.locked ? 'var(--hig-system-blue)' : 'var(--hig-text-secondary)'}}>{group.locked ? 'lock' : 'lock_open_right'}</span>
                </button>
            </div>

            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>手動切換節點</div>
                <div style={{display: 'flex', gap: '8px'}}>
                    <div className="hig-picker" style={{flex: 1}}>
                        <select id={`select-${group.name}`} defaultValue={group.now}>
                            {group.all_nodes && group.all_nodes.map((n: string) => <option key={n} value={n} >{n}</option>)}
                        </select>
                        <span className="material-symbols-outlined">expand_more</span>
                    </div>
                    <button className="btn secondary" style={{height: '36px'}} onClick={() => {
                        const sel = document.getElementById(`select-${group.name}`) as HTMLSelectElement;
                        manualSwitch(group.name, sel.value);
                    }}>套用</button>
                </div>
            </div>
            
            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>地區過濾 (REGIONS)</div>
                <div className="chip-group">
                    {['US', 'HK', 'TW', 'JP', 'SG', 'UK'].map(r => {
                        const isSelected = rx.includes(r + '|');
                        return (
                            <div key={r} className={`hig-chip ${isSelected ? 'selected' : ''}`} onClick={() => handleRegionChange(r, isSelected)}>
                                {r}
                            </div>
                        );
                    })}
                </div>
            </div>

            <div>
                <div className="hig-caption-1" style={{color: 'var(--hig-text-secondary)', marginBottom: '8px'}}>服務過濾 (SERVICES)</div>
                <div className="chip-group">
                    <div className={`hig-chip ${isChatGPT ? 'selected' : ''}`} onClick={() => handleServiceChange('chatgpt', isChatGPT)}>
                        ChatGPT
                    </div>
                    <div className={`hig-chip ${isGemini ? 'selected' : ''}`} onClick={() => handleServiceChange('gemini', isGemini)}>
                        Gemini
                    </div>
                    <div className={`hig-chip ${isAntigravity ? 'selected' : ''}`} onClick={() => handleServiceChange('antigravity', isAntigravity)}>
                        Antigravity
                    </div>
                </div>
            </div>
        </div>
    );
}
