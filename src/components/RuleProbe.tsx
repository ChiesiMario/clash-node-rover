import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { invoke } from '@tauri-apps/api/core';
import { Search, ChevronRight } from 'lucide-react';

interface ProbeResult {
  domain: string;
  rule: string;
  rule_payload: string;
  proxy_chain: string[];
}

export function RuleProbe() {
  const { t } = useTranslation();
  const [domain, setDomain] = useState('');
  const [isProbing, setIsProbing] = useState(false);
  const [result, setResult] = useState<ProbeResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [config, setConfig] = useState<any>(null);

  useEffect(() => {
    invoke<any>('get_config').then(setConfig).catch(console.error);
  }, []);

  const updateConfig = (key: string, value: any) => {
    if (!config) return;
    const newConfig = { ...config, [key]: value };
    setConfig(newConfig);
    invoke('save_config', { newConfig }).catch(console.error);
  };

  const handleProbe = async () => {
    if (!domain.trim()) {
      setError(t('probe.error_empty', 'Please enter a domain'));
      return;
    }

    setIsProbing(true);
    setError(null);
    setResult(null);

    // Clean up domain (remove http://, path, etc.)
    let cleanDomain = domain.trim();
    try {
      if (cleanDomain.startsWith('http')) {
        const url = new URL(cleanDomain);
        cleanDomain = url.hostname;
      }
    } catch (e) {
      // Ignore if not a valid URL
    }

    try {
      const res: ProbeResult = await invoke('probe_rule', { domain: cleanDomain });
      setResult(res);
    } catch (e: any) {
      setError(e.toString());
    } finally {
      setIsProbing(false);
    }
  };

  return (
    <div className="p-8 max-w-2xl mx-auto space-y-8 animate-in fade-in duration-300">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">{t('probe.title', 'Routing Rule Probe')}</h1>
        <p className="text-muted-foreground">{t('probe.subtitle', 'Enter a domain to probe its triggered routing rule and final proxy chain.')}</p>
      </div>

      <div className="space-y-10">
        
        {/* Proxy Settings Card */}
        {config && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium">{t('probe.use_proxy', 'Use HTTP Proxy')}</h2>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.probe_use_proxy}
                  onChange={(e) => updateConfig('probe_use_proxy', e.target.checked)}
                  className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer"
                />
              </label>
            </div>
            
            {config.probe_use_proxy ? (
              <div className="space-y-2 pt-2 border-t border-border/50">
                <input
                  type="text"
                  value={config.probe_proxy_url}
                  onChange={(e) => updateConfig('probe_proxy_url', e.target.value)}
                  placeholder={t('probe.proxy_url_placeholder', 'e.g. http://127.0.0.1:7890')}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>
            ) : (
              <div className="pt-2 border-t border-border/50">
                <p className="text-sm text-muted-foreground">{t('probe.tun_mode_hint', 'Default is no proxy. Ensure Tun Mode is enabled in Clash to intercept requests.')}</p>
              </div>
            )}
          </div>
        )}

        {/* Input Area Card */}
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <input
              type="text"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleProbe()}
              placeholder={t('probe.domain_placeholder', 'e.g. youtube.com')}
              className="flex-1 bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
            <button 
              onClick={handleProbe} 
              disabled={isProbing}
              className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-md hover:opacity-90 transition-opacity disabled:opacity-50 font-medium text-sm shadow-sm whitespace-nowrap"
            >
              {isProbing ? <Search className="w-4 h-4 animate-pulse" /> : <Search className="w-4 h-4" />}
              {isProbing ? t('probe.probing', 'Probing...') : t('probe.probe_btn', 'Probe')}
            </button>
          </div>
          {error && (
            <p className="text-sm text-destructive mt-2">{error}</p>
          )}
        </div>

        {/* Result Area */}
        {result && (
          <div className="space-y-4">
            <h2 className="text-lg font-medium">Probe Result</h2>
            <div className="space-y-4 pt-4 border-t border-border/50">
              <div className="grid grid-cols-[100px_1fr] md:grid-cols-[120px_1fr] gap-4">
                <div className="text-sm font-medium text-muted-foreground flex items-center">{t('probe.domain', 'Domain')}</div>
                <div className="text-sm font-medium break-all">{result.domain}</div>
              </div>

              <div className="grid grid-cols-[100px_1fr] md:grid-cols-[120px_1fr] gap-4">
                <div className="text-sm font-medium text-muted-foreground flex items-center">{t('probe.rule', 'Rule')}</div>
                <div className="text-sm flex flex-wrap items-center gap-2">
                  <span className="font-medium">{result.rule_payload || result.rule}</span>
                  {result.rule_payload && <span className="text-muted-foreground">[{result.rule}]</span>}
                </div>
              </div>

              <div className="grid grid-cols-[100px_1fr] md:grid-cols-[120px_1fr] gap-4">
                <div className="text-sm font-medium text-muted-foreground flex items-center">{t('probe.proxy_server', 'Proxy Server')}</div>
                <div className="text-sm">
                  {result.proxy_chain && result.proxy_chain.length > 0 ? (
                    <div className="flex flex-wrap items-center gap-1.5 text-foreground/90">
                      {result.proxy_chain.slice().reverse().map((node, i, arr) => (
                        <div key={i} className="flex items-center">
                          <span className="px-2 py-0.5 rounded-md bg-muted border border-border/50">
                            {node}
                          </span>
                          {i < arr.length - 1 && (
                            <ChevronRight className="w-4 h-4 mx-0.5 text-muted-foreground/50" />
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <span className="text-muted-foreground italic">DIRECT</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
