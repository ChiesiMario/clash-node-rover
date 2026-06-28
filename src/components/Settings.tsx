import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Save } from "lucide-react";

export function Settings() {
  const [config, setConfig] = useState<any>(null);
  const [saving, setSaving] = useState(false);
  const [availableGroups, setAvailableGroups] = useState<string[]>([]);

  useEffect(() => {
    invoke("get_config").then((cfg: any) => {
      setConfig(cfg);
      if (cfg && cfg.api_url) {
        invoke<string[]>("get_clash_selectors")
          .then((groups) => setAvailableGroups(groups))
          .catch((e) => console.error("Failed to fetch groups on mount:", e));
      }
    });
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await invoke("save_config", { newConfig: config });
    } catch (e) {
      console.error(e);
    }
    setSaving(false);
  };

  if (!config) return <div className="p-8 text-muted-foreground">Loading settings...</div>;

  return (
    <div className="p-8 max-w-2xl mx-auto space-y-8 animate-in fade-in duration-300">
      <div className="space-y-2 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">Configuration</h1>
          <p className="text-muted-foreground">Manage your Clash API and Node Rover settings.</p>
        </div>
        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-md hover:opacity-90 transition-opacity disabled:opacity-50 font-medium text-sm shadow-sm"
        >
          <Save className="w-4 h-4" />
          {saving ? "Saving..." : "Save Changes"}
        </button>
      </div>

      <div className="space-y-6">
        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <h2 className="text-lg font-medium">API Connection</h2>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Clash API URL</label>
              <input
                type="text"
                value={config.api_url}
                onChange={(e) => setConfig({ ...config, api_url: e.target.value })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="http://127.0.0.1:9090"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">API Secret (Optional)</label>
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
            <h2 className="text-lg font-medium">Monitored Groups</h2>
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
              Fetch Groups
            </button>
          </div>
          
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">Select which Clash proxy groups the watchdog should monitor and speed-test.</p>
            
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
                <p className="text-xs text-muted-foreground">Comma separated list of groups. Click "Fetch Groups" to select from Clash API.</p>
              </div>
            )}
          </div>
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <h2 className="text-lg font-medium">Advanced Speed Test Algorithm</h2>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Switch Tolerance (Score)</label>
              <input
                type="number"
                value={config.tolerance}
                onChange={(e) => setConfig({ ...config, tolerance: parseInt(e.target.value) || 0 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="10"
              />
              <p className="text-xs text-muted-foreground">Don't switch if the current node is within this many score points of the best node.</p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Ping Timeout (ms)</label>
              <input
                type="number"
                value={config.test_timeout}
                onChange={(e) => setConfig({ ...config, test_timeout: parseInt(e.target.value) || 0 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="2000"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Max Concurrency</label>
              <input
                type="number"
                value={config.max_concurrent}
                onChange={(e) => setConfig({ ...config, max_concurrent: parseInt(e.target.value) || 10 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="10"
              />
              <p className="text-xs text-muted-foreground">Maximum number of simultaneous ping requests to send.</p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Ping Count</label>
              <input
                type="number"
                value={config.ping_count}
                onChange={(e) => setConfig({ ...config, ping_count: parseInt(e.target.value) || 3 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="3"
                min="1"
              />
              <p className="text-xs text-muted-foreground">Number of times to test each node. Higher values improve jitter calculation but take longer.</p>
            </div>
          </div>
        </div>

        <div className="space-y-4 p-6 rounded-xl border border-border bg-card">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">HTTP Proxy Testing (Pre-switch Verification)</h2>
            <label className="flex items-center gap-2 cursor-pointer">
              <span className="text-sm font-medium text-muted-foreground">Enable Testing</span>
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
                  <label className="text-sm font-medium text-muted-foreground">Dedicated Test Group (Required)</label>
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
                      Fetch Groups
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
                        {group} {config.target_groups.includes(group) ? "(Used in Monitored)" : ""}
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
                <p className="text-xs text-muted-foreground">The system will switch this group to the target node and send requests through the proxy below.</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">HTTP Proxy Server Address (Required)</label>
                <input
                  type="text"
                  value={config.clash_proxy_url}
                  onChange={(e) => setConfig({ ...config, clash_proxy_url: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  placeholder="e.g., 127.0.0.1:7890"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">Target Test URLs (One per line)</label>
                <textarea
                  value={config.browser_test_urls.join('\n')}
                  onChange={(e) => setConfig({ ...config, browser_test_urls: e.target.value.split('\n').filter(s => s.trim() !== '') })}
                  className="w-full h-24 bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none"
                  placeholder="https://www.google.com&#10;https://www.youtube.com"
                />
                <p className="text-xs text-muted-foreground">Before switching, the system will send GET requests to all these URLs simultaneously. All requests must succeed to proceed.</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
