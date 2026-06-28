import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Save } from "lucide-react";
import { useTranslation } from "react-i18next";

export function Settings() {
  const { t, i18n } = useTranslation();
  const [config, setConfig] = useState<any>(null);
  const [saving, setSaving] = useState(false);
  const [availableGroups, setAvailableGroups] = useState<string[]>([]);
  const [autostartEnabled, setAutostartEnabled] = useState(false);
  const [autostartLoading, setAutostartLoading] = useState(true);

  useEffect(() => {
    invoke("get_config").then((cfg: any) => {
      setConfig(cfg);
      if (cfg && cfg.api_url) {
        invoke<string[]>("get_clash_selectors")
          .then((groups) => setAvailableGroups(groups))
          .catch((e) => console.error("Failed to fetch groups on mount:", e));
      }
    });

    import('@tauri-apps/plugin-autostart').then(({ isEnabled }) => {
      isEnabled().then(setAutostartEnabled).finally(() => setAutostartLoading(false));
    }).catch((e) => {
      console.error("Failed to load autostart plugin:", e);
      setAutostartLoading(false);
    });
  }, []);

  const handleToggleAutostart = async (checked: boolean) => {
    setAutostartLoading(true);
    try {
      const { enable, disable } = await import('@tauri-apps/plugin-autostart');
      if (checked) {
        await enable();
      } else {
        await disable();
      }
      setAutostartEnabled(checked);
    } catch (e) {
      console.error("Autostart error:", e);
      alert("Failed to toggle auto-start: " + e);
    }
    setAutostartLoading(false);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await invoke("save_config", { newConfig: config });
    } catch (e) {
      console.error(e);
    }
    setSaving(false);
  };

  if (!config) return <div className="p-8 text-muted-foreground">{t('settings.loading', 'Loading settings...')}</div>;

  return (
    <div className="p-8 max-w-2xl mx-auto space-y-8 animate-in fade-in duration-300">
      <div className="space-y-2 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">{t('settings.configuration', 'Configuration')}</h1>
          <p className="text-muted-foreground">{t('settings.subtitle', 'Manage your Clash API and Node Rover settings.')}</p>
        </div>
        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-md hover:opacity-90 transition-opacity disabled:opacity-50 font-medium text-sm shadow-sm"
        >
          <Save className="w-4 h-4" />
          {saving ? t('settings.saving', 'Saving...') : t('settings.save_changes', 'Save Changes')}
        </button>
      </div>

      <div className="space-y-6">
        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <h2 className="text-lg font-medium">{t('settings.api_connection', 'API Connection')}</h2>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">{t('settings.clash_api_url', 'Clash API URL')}</label>
              <input
                type="text"
                value={config.api_url}
                onChange={(e) => setConfig({ ...config, api_url: e.target.value })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="http://127.0.0.1:9090"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">{t('settings.api_secret', 'API Secret (Optional)')}</label>
              <input
                type="password"
                value={config.api_secret}
                onChange={(e) => setConfig({ ...config, api_secret: e.target.value })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="Enter secret"
              />
            </div>
          </div>
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">{t('settings.monitored_groups', 'Monitored Groups')}</h2>
            <button
              onClick={async () => {
                try {
                  const groups = await invoke<string[]>("get_clash_selectors");
                  setAvailableGroups(groups);
                } catch (e: any) {
                  alert("Failed to fetch groups: " + e);
                }
              }}
              className="text-xs bg-secondary text-secondary-foreground px-3 py-1.5 rounded-md hover:bg-secondary/80 font-medium transition-colors"
            >
              {t('settings.fetch_groups', 'Fetch Groups')}
            </button>
          </div>
          
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t('settings.monitored_groups_desc', 'Select which Clash proxy groups the watchdog should monitor and speed-test.')}</p>
            
            {availableGroups.length > 0 ? (
              <div className="grid grid-cols-2 gap-3 mt-4">
                {availableGroups.map(group => (
                  <label key={group} className="flex items-center space-x-3 p-3 border rounded-lg hover:bg-muted/50 cursor-pointer transition-colors">
                    <input
                      type="checkbox"
                      checked={config.target_groups.includes(group)}
                      disabled={config.enable_browser_test && config.dedicated_test_group === group}
                      onChange={(e) => {
                        const newGroups = e.target.checked
                          ? [...config.target_groups, group]
                          : config.target_groups.filter((g: string) => g !== group);
                        setConfig({ ...config, target_groups: newGroups });
                      }}
                      className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary disabled:opacity-50"
                    />
                    <span className={`text-sm font-medium ${config.enable_browser_test && config.dedicated_test_group === group ? "text-muted-foreground" : ""}`}>{group}</span>
                  </label>
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                <input
                  type="text"
                  value={config.target_groups.join(", ")}
                  onChange={(e) => setConfig({ ...config, target_groups: e.target.value.split(",").map((s: string) => s.trim()).filter(Boolean) })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="e.g. PROXIES, FALLBACK"
                />
                <p className="text-xs text-muted-foreground">{t('settings.comma_separated_desc', 'Comma separated list of groups. Click "Fetch Groups" to select from Clash API.')}</p>
              </div>
            )}
          </div>
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <h2 className="text-lg font-medium">{t('settings.advanced_algo', 'Advanced Speed Test Algorithm')}</h2>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">{t('settings.switch_tolerance', 'Switch Tolerance (Score)')}</label>
              <input
                type="number"
                value={config.tolerance}
                onChange={(e) => setConfig({ ...config, tolerance: parseInt(e.target.value) || 0 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="10"
              />
              <p className="text-xs text-muted-foreground">{t('settings.tolerance_desc', "Don't switch if the current node is within this many score points of the best node.")}</p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">{t('settings.ping_timeout', 'Ping Timeout (ms)')}</label>
              <input
                type="number"
                value={config.test_timeout}
                onChange={(e) => setConfig({ ...config, test_timeout: parseInt(e.target.value) || 0 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="2000"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">{t('settings.max_concurrency', 'Max Concurrency')}</label>
              <input
                type="number"
                value={config.max_concurrent}
                onChange={(e) => setConfig({ ...config, max_concurrent: parseInt(e.target.value) || 10 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="10"
              />
              <p className="text-xs text-muted-foreground">{t('settings.max_concurrency_desc', 'Maximum number of simultaneous ping requests to send.')}</p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">{t('settings.ping_count', 'Ping Count')}</label>
                <div className="relative">
                  <input
                    type="number"
                    min="1"
                    max="10"
                    value={config.ping_count || 1}
                    onChange={(e) => setConfig({ ...config, ping_count: parseInt(e.target.value) || 1 })}
                    className="w-full bg-background border border-border rounded-md px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
                  />
                  <div className="absolute inset-y-0 right-3 flex items-center pointer-events-none text-muted-foreground text-sm">
                    {t('settings.times', 'times')}
                  </div>
                </div>
                <p className="text-xs text-muted-foreground mt-1">{t('settings.ping_count_desc', 'Number of pings per node for better average')}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">{t('settings.max_backoff_times', 'Max Backoff Times')}</label>
                <div className="relative">
                  <input
                    type="number"
                    min="0"
                    max="10"
                    value={config.max_backoff_times ?? 5}
                    onChange={(e) => setConfig({ ...config, max_backoff_times: parseInt(e.target.value) || 0 })}
                    className="w-full bg-background border border-border rounded-md px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
                  />
                  <div className="absolute inset-y-0 right-3 flex items-center pointer-events-none text-muted-foreground text-sm">
                    {t('settings.rounds', 'rounds')}
                  </div>
                </div>
                <p className="text-xs text-muted-foreground mt-1">{t('settings.max_backoff_desc', 'Max rounds to skip failing nodes')}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">{t('settings.http_proxy_test', 'HTTP Proxy Testing (Pre-switch Verification)')}</h2>
            <label className="flex items-center gap-2 cursor-pointer">
              <span className="text-sm font-medium text-muted-foreground">{t('settings.enable_testing', 'Enable Testing')}</span>
              <input
                type="checkbox"
                checked={config.enable_browser_test}
                onChange={(e) => setConfig({ ...config, enable_browser_test: e.target.checked })}
                className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer"
              />
            </label>
          </div>
          
          {config.enable_browser_test && (
            <div className="space-y-4 pt-2 border-t border-border/50">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium text-muted-foreground">{t('settings.dedicated_test_group', 'Dedicated Test Group (Required)')}</label>
                  {availableGroups.length === 0 && (
                    <button
                      onClick={async () => {
                        try {
                          const groups = await invoke<string[]>("get_clash_selectors");
                          setAvailableGroups(groups);
                        } catch (e: any) {
                          alert("Failed to fetch groups: " + e);
                        }
                      }}
                      className="text-[10px] bg-secondary text-secondary-foreground px-2 py-1 rounded-md hover:bg-secondary/80 font-medium transition-colors"
                    >
                      {t('settings.fetch_groups', 'Fetch Groups')}
                    </button>
                  )}
                </div>
                {availableGroups.length > 0 ? (
                  <select
                    value={config.dedicated_test_group}
                    onChange={(e) => setConfig({ ...config, dedicated_test_group: e.target.value })}
                    className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="" disabled>Select a group...</option>
                    {availableGroups.map((group) => (
                      <option 
                        key={group} 
                        value={group} 
                        disabled={config.target_groups.includes(group)}
                      >
                        {group} {config.target_groups.includes(group) ? t('settings.used_in_monitored', '(Used in Monitored)') : ""}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    type="text"
                    value={config.dedicated_test_group}
                    onChange={(e) => setConfig({ ...config, dedicated_test_group: e.target.value })}
                    className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="e.g., SpeedTest"
                  />
                )}
                <p className="text-xs text-muted-foreground">{t('settings.test_group_desc', 'The system will switch this group to the target node and send requests through the proxy below.')}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">{t('settings.http_proxy_addr', 'HTTP Proxy Server Address (Required)')}</label>
                <input
                  type="text"
                  value={config.clash_proxy_url}
                  onChange={(e) => setConfig({ ...config, clash_proxy_url: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="e.g., 127.0.0.1:7890"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">{t('settings.target_urls', 'Target Test URLs (One per line)')}</label>
                <textarea
                  value={config.browser_test_urls.join('\n')}
                  onChange={(e) => setConfig({ ...config, browser_test_urls: e.target.value.split('\n').filter(s => s.trim() !== '') })}
                  className="w-full h-24 bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                  placeholder="https://www.google.com&#10;https://www.youtube.com"
                />
                <p className="text-xs text-muted-foreground">{t('settings.target_urls_desc', 'Before switching, the system will send GET requests to all these URLs simultaneously. All requests must succeed to proceed.')}</p>
              </div>
            </div>
          )}
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-medium">{t('settings.system_integration', 'System Integration')}</h2>
              <p className="text-sm text-muted-foreground mt-1">{t('settings.sys_int_desc', 'Configure how Node Rover integrates with your operating system.')}</p>
            </div>
          </div>
          <div className="space-y-4 pt-2 border-t border-border/50">
            <div className="flex items-center justify-between">
              <div>
                <label className="text-sm font-medium text-foreground">{t('settings.autostart', 'Auto-Start on Boot (Silent Boot)')}</label>
                <p className="text-xs text-muted-foreground mt-0.5">{t('settings.autostart_desc', 'Start Node Rover automatically when you log in. It will start silently in the system tray.')}</p>
              </div>
              <label className="flex items-center gap-2 cursor-pointer relative">
                {autostartLoading && <span className="absolute -left-6 text-muted-foreground animate-spin">◷</span>}
                <input
                  type="checkbox"
                  checked={autostartEnabled}
                  disabled={autostartLoading}
                  onChange={(e) => handleToggleAutostart(e.target.checked)}
                  className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer disabled:opacity-50"
                />
              </label>
            </div>

            <div className="flex items-center justify-between pt-4 border-t border-border/50">
              <div>
                <label className="text-sm font-medium text-foreground">{t('settings.language', 'Language')}</label>
                <p className="text-xs text-muted-foreground mt-0.5">{t('settings.language_desc', 'Select the application language.')}</p>
              </div>
              <select
                value={config.language || "auto"}
                onChange={(e) => {
                  const newLang = e.target.value;
                  setConfig({ ...config, language: newLang });
                  if (newLang !== "auto") {
                    i18n.changeLanguage(newLang);
                  } else {
                    // Assuming we let browser-detector fallback, but simple reload or just not calling change is fine.
                    // Better to just call it with detected language if we can, but setting auto will let the backend handle tray, and here we can refresh or let it be.
                  }
                }}
                className="bg-background border border-border rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
              >
                <option value="auto">{t('settings.lang_auto', 'Auto-Detect')}</option>
                <option value="en">English</option>
                <option value="zh-TW">繁體中文</option>
                <option value="zh-CN">简体中文</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
