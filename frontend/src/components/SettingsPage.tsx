import { useState, useEffect } from 'react';
import { GetSelectors, GetConfigInfo, SaveConfig, TestConnection } from '../../wailsjs/go/main/App';

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
        GetSelectors()
            .then(data => setSelectors(data || []))
            .catch(err => console.error("Failed to load selectors", err));
    };

    useEffect(() => {
        GetConfigInfo()
            .then(data => setConfig(data))
            .catch(err => console.error("Failed to load config", err));
    }, []);

    useEffect(() => {
        if (apiConnected) {
            fetchSelectors();
        }
    }, [apiConnected]);

    const handleChange = (field: string, value: any) => {
        setConfig((prev: any) => ({ ...prev, [field]: value }));
        setSaveStatus('idle');
    };


    const handleSave = async () => {
        setIsSaving(true);
        setSaveStatus('idle');
        try {
            await SaveConfig(config);
            setSaveStatus('success');
            setTimeout(() => setSaveStatus('idle'), 3000);
            // 延遲 0.5 秒以等待後端重新初始化 APIClient
            setTimeout(() => {
                fetchSelectors();
                if (onSaveSuccess) onSaveSuccess();
            }, 500);
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
            await TestConnection(config.api_url, config.api_secret);
            setTestResult('success');
        } catch (e: any) {
            console.error(e);
            setTestResult('error');
            setTestErrorMsg(e.message || String(e));
        } finally {
            setIsTesting(false);
        }
    };

    if (!config) return <div className="hig-body">載入中...</div>;

    return (
        <div className="hig-card" style={{ maxWidth: '800px', margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <div className="hig-title-1">系統設定</div>
            
            <div className="apple-segmented-control">
                <button className={activeTab === 'basic' ? 'active' : ''} onClick={() => handleTabClick('basic')}>基礎設定</button>
                <button className={activeTab === 'speed' ? 'active' : ''} onClick={() => handleTabClick('speed')} disabled={!apiConnected}>測速設定</button>
                <button className={activeTab === 'browser' ? 'active' : ''} onClick={() => handleTabClick('browser')} disabled={!apiConnected}>無頭瀏覽器</button>
                <button className={activeTab === 'advanced' ? 'active' : ''} onClick={() => handleTabClick('advanced')} disabled={!apiConnected}>進階與維護</button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column' }}>
                {activeTab === 'basic' && (
                    <div className="apple-list-group">
                        <InputField label="Clash API 網址" value={config.api_url} onChange={(v: any) => handleChange('api_url', v)} placeholder="http://127.0.0.1:9090" />
                        <InputField label="API 密鑰" value={config.api_secret} onChange={(v: any) => handleChange('api_secret', v)} type="password" />
                        <div className="apple-list-item">
                            <label className="apple-list-label">連線狀態</label>
                            <div className="apple-list-control" style={{ gap: '12px', alignItems: 'center' }}>
                                {testResult === 'success' && <span style={{ color: 'var(--hig-system-green)', fontSize: '14px' }}>連線成功</span>}
                                {testResult === 'error' && <span style={{ color: 'var(--hig-system-red)', fontSize: '14px' }}>連線失敗: {testErrorMsg}</span>}
                                <button className="btn secondary" onClick={handleTestConnection} disabled={isTesting} style={{ padding: '4px 12px', borderRadius: '16px', fontSize: '13px' }}>
                                    {isTesting ? '測試中...' : '測試連通性'}
                                </button>
                            </div>
                        </div>
                        <InputField label="Web UI 埠號" value={config.web_port} onChange={(v: any) => handleChange('web_port', Number(v))} type="number" />
                    </div>
                )}

                {activeTab === 'speed' && (
                    <>
                        <DynamicList 
                            label="監控目標群組" 
                            items={config.target_groups} 
                            onChange={(items: any) => handleChange('target_groups', items)} 
                            options={selectors.map(s => ({value: s, label: s}))}
                        />

                        <div className="hig-footnote" style={{ marginBottom: '8px', marginLeft: '16px', color: 'var(--hig-text-secondary)', fontSize: '13px', textTransform: 'uppercase' }}>測速細節</div>
                        <div className="apple-list-group">
                            <InputField 
                                label="專屬測速群組" 
                                value={config.dedicated_test_group} 
                                onChange={(v: any) => handleChange('dedicated_test_group', v)} 
                                type="select"
                                options={[{value: '', label: '(無)'}, ...selectors.map(s => ({value: s, label: s}))]}
                            />
                            <InputField label="Clash Proxy 網址" value={config.clash_proxy_url} onChange={(v: any) => handleChange('clash_proxy_url', v)} />
                        </div>
                        
                        <DynamicList 
                            label="測速目標網址" 
                            items={config.test_urls} 
                            onChange={(items: any) => handleChange('test_urls', items)} 
                        />

                        <div className="hig-footnote" style={{ marginBottom: '8px', marginLeft: '16px', color: 'var(--hig-text-secondary)', fontSize: '13px', textTransform: 'uppercase' }}>進階測速參數</div>
                        <div className="apple-list-group">
                            <InputField label="測速週期 (秒)" value={config.check_interval} onChange={(v: any) => handleChange('check_interval', Number(v))} type="number" />
                            <InputField label="連線超時時間 (秒)" value={config.test_timeout} onChange={(v: any) => handleChange('test_timeout', Number(v))} type="number" />
                            <InputField label="測速容忍誤差 (毫秒)" value={config.tolerance_ms} onChange={(v: any) => handleChange('tolerance_ms', Number(v))} type="number" />
                            <InputField label="最大併發測速數" value={config.max_concurrent} onChange={(v: any) => handleChange('max_concurrent', Number(v))} type="number" />
                            <InputField label="最大退避懲罰次數" value={config.max_backoff_cycles} onChange={(v: any) => handleChange('max_backoff_cycles', Number(v))} type="number" />
                        </div>
                    </>
                )}

                {activeTab === 'browser' && (
                    <>
                        <div className="apple-list-group">
                            <InputField 
                                label="啟用無頭瀏覽器測試" 
                                value={config.enable_browser_test} 
                                onChange={(v: any) => handleChange('enable_browser_test', v)} 
                                type="checkbox" 
                            />
                        </div>
                        <DynamicList 
                            label="瀏覽器測速網址" 
                            items={config.browser_test_urls} 
                            onChange={(items: any) => handleChange('browser_test_urls', items)} 
                        />
                    </>
                )}

                {activeTab === 'advanced' && (
                    <div className="apple-list-group">
                        <InputField label="歷史紀錄清理天數" value={config.cleanup_days} onChange={(v: any) => handleChange('cleanup_days', Number(v))} type="number" />
                    </div>
                )}
            </div>

            <div style={{ marginTop: '32px' }}>
                <button className="btn" onClick={handleSave} disabled={isSaving} style={{ width: '100%', padding: '16px 0', fontSize: '17px', borderRadius: '12px' }}>
                    <span className="material-symbols-outlined">{isSaving ? 'progress_activity' : 'check_circle'}</span>
                    {isSaving ? '儲存中...' : '儲存並套用'}
                </button>
                {saveStatus === 'success' && <div style={{ color: 'var(--hig-system-green)', textAlign: 'center', marginTop: '12px', fontSize: '14px' }}>設定已成功儲存並生效</div>}
                {saveStatus === 'error' && <div style={{ color: 'var(--hig-system-red)', textAlign: 'center', marginTop: '12px', fontSize: '14px' }}>儲存失敗，請檢查輸入內容</div>}
            </div>
        </div>
    );
}

function InputField({ label, value, onChange, type = 'text', placeholder = '', options }: any) {
    return (
        <div className="apple-list-item">
            <label className="apple-list-label">{label}</label>
            <div className="apple-list-control">
                {type === 'select' ? (
                    <select 
                        className="apple-select"
                        value={value || ''}
                        onChange={(e) => onChange(e.target.value)}
                    >
                        {options?.map((opt: any, idx: number) => (
                            <option key={idx} value={opt.value}>{opt.label}</option>
                        ))}
                    </select>
                ) : type === 'checkbox' ? (
                    <label className="apple-switch">
                        <input type="checkbox" checked={!!value} onChange={(e) => onChange(e.target.checked)} />
                        <span className="slider"></span>
                    </label>
                ) : (
                    <input 
                        type={type} 
                        className="apple-input"
                        value={value || ''} 
                        onChange={(e) => onChange(e.target.value)} 
                        placeholder={placeholder}
                    />
                )}
            </div>
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

    return (
        <>
            <div className="hig-footnote" style={{ marginBottom: '8px', marginLeft: '16px', color: 'var(--hig-text-secondary)', fontSize: '13px', textTransform: 'uppercase' }}>{label}</div>
            <div className="apple-list-group">
                {items.map((item: string, index: number) => (
                    <div key={index} className="apple-list-item">
                        <div className="apple-list-control" style={{ marginRight: '16px', justifyContent: 'flex-start' }}>
                            {options && options.length > 0 ? (
                                <select className="apple-select" style={{textAlign: 'left', width: '100%'}} value={item} onChange={(e) => handleUpdate(index, e.target.value)}>
                                    {options.map((opt: any, idx: number) => (
                                        <option key={idx} value={opt.value}>{opt.label}</option>
                                    ))}
                                </select>
                            ) : (
                                <input type="text" className="apple-input" style={{textAlign: 'left', width: '100%'}} value={item} onChange={(e) => handleUpdate(index, e.target.value)} placeholder={placeholder} />
                            )}
                        </div>
                        <button className="btn secondary" onClick={() => handleRemove(index)} style={{padding: '4px', color: 'var(--hig-system-red)', background: 'transparent', border: 'none'}}>
                            <span className="material-symbols-outlined" style={{ fontSize: '20px' }}>remove_circle</span>
                        </button>
                    </div>
                ))}
                <div className="apple-list-item" style={{cursor: 'pointer'}} onClick={handleAdd}>
                    <div className="apple-list-label" style={{color: 'var(--hig-system-blue)', display: 'flex', alignItems: 'center', gap: '8px'}}>
                        <span className="material-symbols-outlined" style={{ fontSize: '20px' }}>add_circle</span>
                        新增項目
                    </div>
                </div>
            </div>
        </>
    );
}
