import React, { useState } from 'react';
import { SaveSetup, GetAllProxyGroups, GetConfigInfo, SaveConfig } from '../../wailsjs/go/main/App';

interface SetupModalProps {
    defaultUrl: string;
    onSuccess: () => void;
    canCancel?: boolean;
    onCancel?: () => void;
}

const SetupModal: React.FC<SetupModalProps> = ({ defaultUrl, onSuccess, canCancel, onCancel }) => {
    const [step, setStep] = useState(1);
    
    // Step 1 State
    const [apiUrl, setApiUrl] = useState(defaultUrl || 'http://127.0.0.1:9090');
    const [apiSecret, setApiSecret] = useState('');
    const [isSubmitting1, setIsSubmitting1] = useState(false);
    const [errorMsg1, setErrorMsg1] = useState('');

    // Step 2 State
    const [groups, setGroups] = useState<any[]>([]);
    const [selectedGroups, setSelectedGroups] = useState<Set<string>>(new Set());
    const [checkInterval, setCheckInterval] = useState(60);
    const [toleranceMs, setToleranceMs] = useState(20);
    const [isSubmitting2, setIsSubmitting2] = useState(false);
    const [errorMsg2, setErrorMsg2] = useState('');
    const [currentConfig, setCurrentConfig] = useState<any>(null);

    const handleStep1Submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setErrorMsg1('');
        setIsSubmitting1(true);

        try {
            await SaveSetup(apiUrl, apiSecret);
            // 連線成功，載入設定與群組清單
            const config = await GetConfigInfo();
            setCurrentConfig(config);
            setCheckInterval(config.check_interval || 60);
            setToleranceMs(config.tolerance_ms || 20);
            setSelectedGroups(new Set(config.target_groups || []));
            
            const allGroups = await GetAllProxyGroups();
            setGroups(allGroups || []);
            setStep(2);
        } catch (err: any) {
            setErrorMsg1(err.message || String(err) || '發生未知錯誤');
        } finally {
            setIsSubmitting1(false);
        }
    };

    const handleStep2Submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setErrorMsg2('');
        setIsSubmitting2(true);

        try {
            const newConfig = { ...currentConfig };
            newConfig.target_groups = Array.from(selectedGroups);
            newConfig.check_interval = checkInterval;
            newConfig.tolerance_ms = toleranceMs;
            
            await SaveConfig(newConfig);
            onSuccess();
        } catch (err: any) {
            setErrorMsg2(err.message || String(err) || '發生未知錯誤');
        } finally {
            setIsSubmitting2(false);
        }
    };

    const toggleGroup = (groupName: string) => {
        const newSet = new Set(selectedGroups);
        if (newSet.has(groupName)) {
            newSet.delete(groupName);
        } else {
            newSet.add(groupName);
        }
        setSelectedGroups(newSet);
    };

    return (
        <div className="setup-modal-overlay">
            <div className="setup-modal-box" style={{ maxWidth: step === 2 ? '500px' : '400px', transition: 'max-width 0.3s ease' }}>
                <div className="setup-header" style={{ marginBottom: '16px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                        <h2 style={{ margin: 0 }}>{step === 1 ? '連線至 Clash API' : '基礎測速設定'}</h2>
                        <span style={{ fontSize: '13px', color: 'var(--hig-system-blue)', fontWeight: 600 }}>Step {step} of 2</span>
                    </div>
                    <p style={{ margin: 0, color: 'var(--hig-text-secondary)', fontSize: '14px' }}>
                        {step === 1 ? '請輸入您的 Clash 外部控制 API 資訊' : '請選擇要監控的代理群組與測速參數'}
                    </p>
                </div>
                
                {step === 1 && (
                    <form onSubmit={handleStep1Submit}>
                        <div className="apple-list-group" style={{ marginBottom: '24px' }}>
                            <div className="apple-list-item">
                                <label className="apple-list-label">API 網址</label>
                                <div className="apple-list-control">
                                    <input
                                        type="text"
                                        className="apple-input"
                                        value={apiUrl}
                                        onChange={e => setApiUrl(e.target.value)}
                                        placeholder="http://127.0.0.1:9090"
                                        required
                                    />
                                </div>
                            </div>
                            <div className="apple-list-item">
                                <label className="apple-list-label">API 密鑰</label>
                                <div className="apple-list-control">
                                    <input
                                        type="password"
                                        className="apple-input"
                                        value={apiSecret}
                                        onChange={e => setApiSecret(e.target.value)}
                                        placeholder="無密鑰請留空"
                                    />
                                </div>
                            </div>
                        </div>
                        
                        {errorMsg1 && <div className="setup-error" style={{ marginBottom: '24px' }}>{errorMsg1}</div>}
                        
                        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center' }}>
                            {canCancel && (
                                <button type="button" onClick={onCancel} disabled={isSubmitting1} className="btn secondary" style={{ flex: 1 }}>
                                    取消
                                </button>
                            )}
                            <button type="submit" disabled={isSubmitting1} className="btn" style={{ flex: canCancel ? 1 : 'none', width: canCancel ? 'auto' : '100%' }}>
                                {isSubmitting1 ? '正在連線...' : '連線並繼續'}
                            </button>
                        </div>
                    </form>
                )}

                {step === 2 && (
                    <form onSubmit={handleStep2Submit}>
                        <div className="hig-footnote" style={{ marginBottom: '8px', marginLeft: '16px', color: 'var(--hig-text-secondary)', fontSize: '13px' }}>測速目標群組</div>
                        <div className="apple-list-group" style={{ marginBottom: '24px', maxHeight: '200px', overflowY: 'auto' }}>
                            {groups.length === 0 ? (
                                <div className="apple-list-item" style={{ justifyContent: 'center', color: 'var(--hig-text-secondary)' }}>
                                    沒有找到可用的代理群組
                                </div>
                            ) : (
                                groups.map(g => (
                                    <div key={g.name} className="apple-list-item" style={{ cursor: 'pointer' }} onClick={() => toggleGroup(g.name)}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1 }}>
                                            <input 
                                                type="checkbox" 
                                                checked={selectedGroups.has(g.name)} 
                                                onChange={() => {}} // Handle change via div onClick
                                                style={{ cursor: 'pointer' }}
                                            />
                                            <span style={{ fontWeight: 500, fontSize: '15px' }}>{g.name}</span>
                                        </div>
                                        <span style={{ fontSize: '13px', color: 'var(--hig-text-secondary)' }}>{g.all_count} 個節點</span>
                                    </div>
                                ))
                            )}
                        </div>

                        <div className="hig-footnote" style={{ marginBottom: '8px', marginLeft: '16px', color: 'var(--hig-text-secondary)', fontSize: '13px' }}>基礎參數</div>
                        <div className="apple-list-group" style={{ marginBottom: '24px' }}>
                            <div className="apple-list-item">
                                <label className="apple-list-label">測速週期 (秒)</label>
                                <div className="apple-list-control">
                                    <input
                                        type="number"
                                        className="apple-input"
                                        value={checkInterval}
                                        onChange={e => setCheckInterval(Number(e.target.value))}
                                        min={5}
                                    />
                                </div>
                            </div>
                            <div className="apple-list-item">
                                <label className="apple-list-label">容忍誤差 (毫秒)</label>
                                <div className="apple-list-control">
                                    <input
                                        type="number"
                                        className="apple-input"
                                        value={toleranceMs}
                                        onChange={e => setToleranceMs(Number(e.target.value))}
                                        min={0}
                                    />
                                </div>
                            </div>
                        </div>
                        
                        {errorMsg2 && <div className="setup-error" style={{ marginBottom: '24px' }}>{errorMsg2}</div>}
                        
                        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center' }}>
                            <button type="button" onClick={() => setStep(1)} disabled={isSubmitting2} className="btn secondary" style={{ flex: 1 }}>
                                上一步
                            </button>
                            <button type="submit" disabled={isSubmitting2} className="btn" style={{ flex: 1 }}>
                                {isSubmitting2 ? '儲存中...' : '完成並啟動'}
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
};

export default SetupModal;
