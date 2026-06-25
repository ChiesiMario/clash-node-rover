import React, { useState } from 'react';

interface SetupModalProps {
    defaultUrl: string;
    onSuccess: () => void;
    canCancel?: boolean;
    onCancel?: () => void;
}

const SetupModal: React.FC<SetupModalProps> = ({ defaultUrl, onSuccess, canCancel, onCancel }) => {
    const [apiUrl, setApiUrl] = useState(defaultUrl || 'http://127.0.0.1:9090');
    const [apiSecret, setApiSecret] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [errorMsg, setErrorMsg] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setErrorMsg('');
        setIsSubmitting(true);

        try {
            const res = await fetch('/api/setup', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ api_url: apiUrl, api_secret: apiSecret })
            });

            if (!res.ok) {
                const text = await res.text();
                throw new Error(text || '驗證失敗');
            }

            // 成功
            onSuccess();
        } catch (err: any) {
            setErrorMsg(err.message || '發生未知錯誤');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="setup-modal-overlay">
            <div className="setup-modal-box">
                <div className="setup-header">
                    <h2>Clash API 設定</h2>
                    <p>請輸入您的 Clash 外部控制 API 資訊</p>
                </div>
                <form onSubmit={handleSubmit} className="setup-form">
                    <div className="form-group">
                        <label>Clash API 網址</label>
                        <input
                            type="text"
                            value={apiUrl}
                            onChange={e => setApiUrl(e.target.value)}
                            placeholder="例如: http://127.0.0.1:9090"
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label>API 密鑰 (Secret)</label>
                        <input
                            type="password"
                            value={apiSecret}
                            onChange={e => setApiSecret(e.target.value)}
                            placeholder="無密鑰請留空"
                        />
                    </div>
                    
                    {errorMsg && <div className="setup-error">{errorMsg}</div>}
                    
                    <div style={{ display: 'flex', gap: '12px' }}>
                        {canCancel && (
                            <button type="button" onClick={onCancel} disabled={isSubmitting} className="btn secondary" style={{ flex: 1 }}>
                                取消
                            </button>
                        )}
                        <button type="submit" disabled={isSubmitting} className="btn-primary" style={{ flex: canCancel ? 1 : 'none' }}>
                            {isSubmitting ? '正在驗證連線...' : '確認並驗證'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default SetupModal;
