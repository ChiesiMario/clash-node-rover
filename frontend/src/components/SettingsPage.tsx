import { useState, useEffect } from 'react';

export default function SettingsPage({ apiConnected = false, onSaveSuccess }: { apiConnected?: boolean, onSaveSuccess?: () => void }) {
    const [config, setConfig] = useState<any>(null);
    const [selectors, setSelectors] = useState<string[]>([]);
    const [activeTab, setActiveTab] = useState('basic');

    useEffect(() => {
        if (!apiConnected && activeTab !== 'basic') {
            setActiveTab('basic');
        }
    }, [apiConnected]);

    const handleTabClick = (tab: string) => {
        if (!apiConnected && tab !== 'basic') {
            setActiveTab('basic');
            return;
        }
        setActiveTab(tab);
    };

    const [isSaving, setIsSaving] = useState(false);
    const [saveStatus, setSaveStatus] = useState<'idle' | 'success' | 'error'>('idle');
    const [isTesting, setIsTesting] = useState(false);
    const [testResult, setTestResult] = useState<'idle' | 'success' | 'error'>('idle');
    const [testErrorMsg, setTestErrorMsg] = useState('');

    const fetchSelectors = () => {
        fetch('/api/selectors')
            .then(res => res.json())
            .then(data => setSelectors(data || []))
            .catch(err => console.error("Failed to load selectors", err));
    };

    useEffect(() => {
        fetch('/api/config')
            .then(res => res.json())
            .then(data => setConfig(data))
            .catch(err => console.error("Failed to load config", err));
            
        fetchSelectors();
    }, []);

    const handleChange = (field: string, value: any) => {
        setConfig((prev: any) => ({ ...prev, [field]: value }));
        setSaveStatus('idle');
    };


    const handleSave = async () => {
        setIsSaving(true);
        setSaveStatus('idle');
        try {
            const res = await fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            if (res.ok) {
                setSaveStatus('success');
                setTimeout(() => setSaveStatus('idle'), 3000);
                // 延遲 0.5 秒以等待後端重新初始化 APIClient（因為後端已改為立刻熱更新）
                setTimeout(() => {
                    fetchSelectors();
                    if (onSaveSuccess) onSaveSuccess();
                }, 500);
            } else {
                setSaveStatus('error');
            }
        } catch (e) {
            console.error(e);
            setSaveStatus('error');
        } finally {
            setIsSaving(false);
        }
    };

    const handleTestConnection = async () => {
        setIsTesting(true);
        setTestResult('idle');
        setTestErrorMsg('');
        try {
            const res = await fetch('/api/test-connection', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ api_url: config.api_url, api_secret: config.api_secret })
            });
            if (res.ok) {
                setTestResult('success');
            } else {
                const text = await res.text();
                setTestResult('error');
                setTestErrorMsg(text);
            }
        } catch (e: any) {
            console.error(e);
            setTestResult('error');
            setTestErrorMsg(e.message);
        } finally {
            setIsTesting(false);
        }
    };

    if (!config) return <div className="hig-body">載入中...</div>;

    return (
        <div className="hig-card" style={{ maxWidth: '800px', margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <div className="hig-title-1">系統設定</div>
            
            <div style={{ display: 'flex', gap: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '12px' }}>
                <button className={`btn ${activeTab === 'basic' ? 'primary' : 'secondary'}`} onClick={() => handleTabClick('basic')}>基礎設定</button>
                <button className={`btn ${activeTab === 'speed' ? 'primary' : 'secondary'}`} onClick={() => handleTabClick('speed')} style={{ opacity: apiConnected ? 1 : 0.5, cursor: apiConnected ? 'pointer' : 'not-allowed' }}>測速設定</button>
                <button className={`btn ${activeTab === 'browser' ? 'primary' : 'secondary'}`} onClick={() => handleTabClick('browser')} style={{ opacity: apiConnected ? 1 : 0.5, cursor: apiConnected ? 'pointer' : 'not-allowed' }}>無頭瀏覽器</button>
                <button className={`btn ${activeTab === 'advanced' ? 'primary' : 'secondary'}`} onClick={() => handleTabClick('advanced')} style={{ opacity: apiConnected ? 1 : 0.5, cursor: apiConnected ? 'pointer' : 'not-allowed' }}>進階與維護</button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                {activeTab === 'basic' && (
                    <>
                        <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                            <div style={{ flex: 1 }}>
                                <InputField label="Clash API 網址" value={config.api_url} onChange={(v: any) => handleChange('api_url', v)} placeholder="http://127.0.0.1:9090" />
                            </div>
                            <div style={{ flex: 1 }}>
                                <InputField label="API 密鑰" value={config.api_secret} onChange={(v: any) => handleChange('api_secret', v)} type="password" />
                            </div>
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '-8px' }}>
                            <button className="btn secondary" onClick={handleTestConnection} disabled={isTesting}>
                                <span className={`material-symbols-outlined ${isTesting ? 'spin' : ''}`}>
                                    {isTesting ? 'refresh' : 'network_check'}
                                </span>
                                測試連通性
                            </button>
                            {testResult === 'success' && <span className="hig-footnote" style={{ color: 'var(--hig-green)' }}>連線成功！</span>}
                            {testResult === 'error' && <span className="hig-footnote" style={{ color: 'var(--hig-red)' }}>連線失敗: {testErrorMsg}</span>}
                        </div>
                        
                        <InputField label="Web UI 埠號" value={config.web_port} onChange={(v: any) => handleChange('web_port', Number(v))} type="number" />
                    </>
                )}

                {activeTab === 'speed' && (
                    <>
                        <DynamicList 
                            label="監控目標群組 (Target Groups)" 
                            items={config.target_groups} 
                            onChange={(items: any) => handleChange('target_groups', items)} 
                            options={selectors.map(s => ({value: s, label: s}))}
                        />

                        <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                            <div style={{ flex: 1 }}>
                                <InputField 
                                    label="專屬測速群組 (選填)" 
                                    value={config.dedicated_test_group} 
                                    onChange={(v: any) => handleChange('dedicated_test_group', v)} 
                                    type="select"
                                    options={[{value: '', label: '(無)'}, ...selectors.map(s => ({value: s, label: s}))]}
                                />
                            </div>
                            <div style={{ flex: 1 }}>
                                <InputField label="Clash Proxy 網址 (供本地測試用)" value={config.clash_proxy_url} onChange={(v: any) => handleChange('clash_proxy_url', v)} />
                            </div>
                        </div>
                        
                        <DynamicList 
                            label="測速目標網址 (Test URLs)" 
                            items={config.test_urls} 
                            onChange={(items: any) => handleChange('test_urls', items)} 
                        />

                        <InputField label="測速週期 (秒)" value={config.check_interval} onChange={(v: any) => handleChange('check_interval', Number(v))} type="number" />
                        <InputField label="連線超時時間 (秒)" value={config.test_timeout} onChange={(v: any) => handleChange('test_timeout', Number(v))} type="number" />
                        <InputField label="測速容忍誤差 (毫秒)" value={config.tolerance_ms} onChange={(v: any) => handleChange('tolerance_ms', Number(v))} type="number" />
                        <InputField label="最大併發測速數" value={config.max_concurrent} onChange={(v: any) => handleChange('max_concurrent', Number(v))} type="number" />
                        <InputField label="最大退避懲罰次數" value={config.max_backoff_cycles} onChange={(v: any) => handleChange('max_backoff_cycles', Number(v))} type="number" />
                    </>
                )}

                {activeTab === 'browser' && (
                    <>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <input type="checkbox" id="enable_browser" checked={config.enable_browser_test} onChange={(e: any) => handleChange('enable_browser_test', e.target.checked)} />
                            <label htmlFor="enable_browser" className="hig-body">啟用無頭瀏覽器測試 (需安裝 Chrome)</label>
                        </div>
                        <DynamicList 
                            label="瀏覽器測速網址" 
                            items={config.browser_test_urls} 
                            onChange={(items: any) => handleChange('browser_test_urls', items)} 
                        />
                    </>
                )}

                {activeTab === 'advanced' && (
                    <>
                        <InputField label="歷史紀錄清理天數" value={config.cleanup_days} onChange={(v: any) => handleChange('cleanup_days', Number(v))} type="number" />
                    </>
                )}
            </div>

            <div style={{ marginTop: '24px', display: 'flex', alignItems: 'center', gap: '16px' }}>
                <button className="btn" onClick={handleSave} disabled={isSaving}>
                    <span className="material-symbols-outlined">{isSaving ? 'sync' : 'save'}</span>
                    {isSaving ? '儲存中...' : '儲存並套用'}
                </button>
                {saveStatus === 'success' && <span style={{ color: 'var(--hig-green)' }}>設定已成功儲存並自動重新載入。</span>}
                {saveStatus === 'error' && <span style={{ color: 'var(--hig-red)' }}>儲存失敗，請檢查輸入內容。</span>}
            </div>
        </div>
    );
}

function InputField({ label, value, onChange, type = 'text', placeholder = '', options }: any) {
    const inputStyle = { 
        padding: '10px 12px', 
        borderRadius: '8px', 
        border: '1px solid var(--border)', 
        background: 'var(--bg-tertiary)', 
        color: 'var(--text-primary)',
        width: '100%',
        boxSizing: 'border-box' as 'border-box'
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
            <label className="hig-footnote" style={{ fontWeight: 500 }}>{label}</label>
            {type === 'select' ? (
                <select 
                    className="setup-form"
                    style={inputStyle}
                    value={value || ''}
                    onChange={(e) => onChange(e.target.value)}
                >
                    {options?.map((opt: any, idx: number) => (
                        <option key={idx} value={opt.value}>{opt.label}</option>
                    ))}
                </select>
            ) : (
                <input 
                    type={type} 
                    className="setup-form"
                    style={inputStyle} 
                    value={value || ''} 
                    onChange={(e) => onChange(e.target.value)} 
                    placeholder={placeholder}
                />
            )}
        </div>
    );
}

function DynamicList({ label, items = [], onChange, placeholder = '', options }: any) {
    const handleUpdate = (index: number, val: string) => {
        const newItems = [...items];
        newItems[index] = val;
        onChange(newItems);
    };

    const handleRemove = (index: number) => {
        const newItems = items.filter((_: any, i: number) => i !== index);
        onChange(newItems);
    };

    const handleAdd = () => {
        if (options && options.length > 0) {
            onChange([...items, options[0].value]);
        } else {
            onChange([...items, '']);
        }
    };

    const inputStyle = { 
        flex: 1,
        padding: '10px 12px', 
        borderRadius: '8px', 
        border: '1px solid var(--border)', 
        background: 'var(--bg-tertiary)', 
        color: 'var(--text-primary)' 
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '8px' }}>
            <label className="hig-footnote" style={{ fontWeight: 500 }}>{label}</label>
            {items.map((item: string, index: number) => (
                <div key={index} style={{ display: 'flex', gap: '8px' }}>
                    {options ? (
                        <select 
                            style={inputStyle}
                            value={item}
                            onChange={(e) => handleUpdate(index, e.target.value)}
                        >
                            {options.map((opt: any, idx: number) => (
                                <option key={idx} value={opt.value}>{opt.label}</option>
                            ))}
                        </select>
                    ) : (
                        <input 
                            type="text" 
                            style={inputStyle} 
                            value={item} 
                            onChange={(e) => handleUpdate(index, e.target.value)} 
                            placeholder={placeholder}
                        />
                    )}
                    <button className="btn secondary" onClick={() => handleRemove(index)} title="刪除">
                        <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>delete</span>
                    </button>
                </div>
            ))}
            <button className="btn secondary" onClick={handleAdd} style={{ alignSelf: 'flex-start', marginTop: '4px' }}>
                <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>add</span>
                新增項目
            </button>
        </div>
    );
}
