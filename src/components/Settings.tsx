import { useEffect, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Save } from "lucide-react";

export function Settings() {
  const [config, setConfig] = useState<any>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    invoke("get_config").then((cfg) => setConfig(cfg));
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
    <div className="p-8 max-w-2xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
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
          <h2 className="text-lg font-medium">Advanced Speed Test Algorithm</h2>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">Switch Tolerance (ms)</label>
              <input
                type="number"
                value={config.tolerance_ms}
                onChange={(e) => setConfig({ ...config, tolerance_ms: parseInt(e.target.value) || 0 })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="10"
              />
              <p className="text-xs text-muted-foreground">Don't switch if the current node is within this many ms of the fastest node.</p>
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
          </div>
        </div>
      </div>
    </div>
  );
}
