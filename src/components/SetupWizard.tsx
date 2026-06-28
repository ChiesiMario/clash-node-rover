import { useState, useEffect } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Server, Zap, CheckCircle2, Loader2, ArrowRight, X } from "lucide-react";

interface SetupWizardProps {
  initialConfig: any;
  onComplete: () => void;
}

export function SetupWizard({ initialConfig, onComplete }: SetupWizardProps) {
  const [step, setStep] = useState(1);
  const [config, setConfig] = useState(initialConfig || {
    api_url: "http://127.0.0.1:9090",
    api_secret: "",
    target_groups: [],
  });
  
  const [isTestingApi, setIsTestingApi] = useState(false);
  const [apiError, setApiError] = useState("");
  const [apiSuccess, setApiSuccess] = useState(false);

  const [availableGroups, setAvailableGroups] = useState<string[]>([]);
  const [isLoadingGroups, setIsLoadingGroups] = useState(false);
  
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (step === 3) {
      fetchGroups();
    }
  }, [step]);

  const testConnection = async () => {
    setIsTestingApi(true);
    setApiError("");
    setApiSuccess(false);
    try {
      await invoke("verify_clash_api", { 
        apiUrl: config.api_url, 
        apiSecret: config.api_secret 
      });
      setApiSuccess(true);
    } catch (e: any) {
      setApiError(String(e));
    }
    setIsTestingApi(false);
  };

  const fetchGroups = async () => {
    setIsLoadingGroups(true);
    try {
      const groups = await invoke<string[]>("get_clash_selectors");
      setAvailableGroups(groups);
    } catch (e: any) {
      console.error("Failed to fetch groups", e);
    }
    setIsLoadingGroups(false);
  };

  const handleFinish = async () => {
    setIsSaving(true);
    try {
      const finalConfig = {
        ...config,
        has_completed_setup: true
      };
      await invoke("save_config", { newConfig: finalConfig });
      await invoke("force_test");
      onComplete();
    } catch (e) {
      console.error("Failed to finish setup:", e);
    }
    setIsSaving(false);
  };

  const skipSetup = async () => {
    handleFinish(); // Just saves as completed
  };

  return (
    <div className="min-h-screen bg-background flex justify-center selection:bg-primary selection:text-primary-foreground">
      <div className="w-full max-w-3xl flex flex-col min-h-screen animate-in fade-in duration-500">
        
        {/* Header */}
        <div className="px-6 sm:px-8 py-8 flex justify-between items-center shrink-0">
          <div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight">Setup Wizard</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Step {step} of 4
            </p>
          </div>
          {step < 4 && (
            <button onClick={skipSetup} className="text-sm font-medium text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors">
              Skip <ArrowRight className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Content */}
        <div className="px-6 sm:px-8 pb-12 flex-1 flex flex-col">
          {step === 1 && (
            <div className="space-y-6 text-center animate-in fade-in slide-in-from-right-4 duration-500 flex-1 flex flex-col justify-center">
              <div className="mx-auto w-16 h-16 bg-primary/10 text-primary rounded-2xl flex items-center justify-center mb-6">
                <Zap className="w-8 h-8" />
              </div>
              <h2 className="text-3xl font-bold tracking-tight">Welcome to Clash Node Rover</h2>
              <p className="text-muted-foreground text-lg max-w-md mx-auto">
                An intelligent, high-performance background engine that monitors and automatically switches your Clash proxies to the fastest nodes available.
              </p>
              <div className="pt-8">
                <button 
                  onClick={() => setStep(2)}
                  className="bg-primary text-primary-foreground px-8 py-3 rounded-full font-medium text-lg shadow-lg hover:shadow-xl hover:opacity-90 transition-all hover:scale-105"
                >
                  Get Started
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="flex flex-col h-full animate-in fade-in slide-in-from-right-4 duration-500">
              <div className="space-y-8 flex-1 pb-8">
                <div className="space-y-2">
                <h2 className="text-2xl font-bold tracking-tight">Connect to Clash API</h2>
                <p className="text-muted-foreground">
                  Rover needs to communicate with your Clash or Clash Meta core via its External Controller API.
                </p>
              </div>
              
              <div className="space-y-4 bg-muted/30 p-6 rounded-xl border border-border">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-foreground">API URL</label>
                  <input
                    type="text"
                    value={config.api_url}
                    onChange={(e) => setConfig({ ...config, api_url: e.target.value })}
                    className="w-full bg-background border border-border rounded-md px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="http://127.0.0.1:9090"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-foreground">API Secret (Optional)</label>
                  <input
                    type="password"
                    value={config.api_secret}
                    onChange={(e) => setConfig({ ...config, api_secret: e.target.value })}
                    className="w-full bg-background border border-border rounded-md px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    placeholder="Enter secret if configured in clash"
                  />
                </div>
                
                <div className="pt-2">
                  <button 
                    onClick={testConnection}
                    disabled={isTestingApi}
                    className="flex items-center gap-2 bg-secondary text-secondary-foreground px-4 py-2 rounded-md font-medium text-sm hover:bg-secondary/80 transition-colors"
                  >
                    {isTestingApi ? <Loader2 className="w-4 h-4 animate-spin" /> : <Server className="w-4 h-4" />}
                    Test Connection
                  </button>
                  
                  {apiSuccess && <p className="text-emerald-500 text-sm mt-3 flex items-center gap-1.5"><CheckCircle2 className="w-4 h-4"/> Connection successful!</p>}
                  {apiError && <p className="text-red-500 text-sm mt-3 flex items-center gap-1.5"><X className="w-4 h-4"/> {apiError}</p>}
                </div>
              </div>
              </div>

              <div className="flex justify-end pt-8 mt-auto shrink-0">
                <button 
                  onClick={() => setStep(3)}
                  disabled={!apiSuccess}
                  className="bg-primary text-primary-foreground px-6 py-2.5 rounded-md font-medium flex items-center gap-2 hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
                >
                  Continue <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="flex flex-col h-full animate-in fade-in slide-in-from-right-4 duration-500">
              <div className="space-y-6 flex-1 pb-8">
                <div className="space-y-2">
                <h2 className="text-2xl font-bold tracking-tight">Select Monitored Groups</h2>
                <p className="text-muted-foreground">
                  Choose which proxy groups (Selectors) you want Rover to continuously speed-test and optimize.
                </p>
              </div>

              <div className="bg-muted/30 p-4 rounded-xl border border-border min-h-[200px]">
                {isLoadingGroups ? (
                  <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2 pt-12">
                    <Loader2 className="w-6 h-6 animate-spin" />
                    <p className="text-sm">Fetching groups from Clash...</p>
                  </div>
                ) : availableGroups.length > 0 ? (
                  <div className="grid grid-cols-2 gap-3">
                    {availableGroups.map(group => (
                      <label key={group} className="flex items-center space-x-3 p-3 border rounded-lg hover:bg-muted/50 cursor-pointer transition-colors bg-card">
                        <input
                          type="checkbox"
                          checked={config.target_groups.includes(group)}
                          onChange={(e) => {
                            const newGroups = e.target.checked
                              ? [...config.target_groups, group]
                              : config.target_groups.filter((g: string) => g !== group);
                            setConfig({ ...config, target_groups: newGroups });
                          }}
                          className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <span className="text-sm font-medium">{group}</span>
                      </label>
                    ))}
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center h-full text-muted-foreground pt-12">
                    <p>No groups found or API disconnected.</p>
                  </div>
                )}
              </div>
              </div>

              <div className="flex justify-between pt-8 mt-auto shrink-0">
                <button 
                  onClick={() => setStep(2)}
                  className="text-muted-foreground hover:text-foreground px-4 py-2 font-medium transition-colors"
                >
                  Back
                </button>
                <button 
                  onClick={() => setStep(4)}
                  className="bg-primary text-primary-foreground px-6 py-2.5 rounded-md font-medium flex items-center gap-2 hover:opacity-90 transition-opacity shadow-sm"
                >
                  Continue <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {step === 4 && (
            <div className="space-y-8 text-center animate-in fade-in slide-in-from-right-4 duration-500 flex-1 flex flex-col justify-center">
              <div className="mx-auto w-20 h-20 bg-emerald-500/10 text-emerald-500 rounded-full flex items-center justify-center mb-6">
                <CheckCircle2 className="w-10 h-10" />
              </div>
              <div className="space-y-3">
                <h2 className="text-3xl font-bold tracking-tight">You're All Set!</h2>
                <p className="text-muted-foreground text-lg max-w-md mx-auto">
                  Clash Node Rover is now ready to optimize your network experience. You can always tweak advanced rules, latency tolerance, and HTTP verification in the Settings tab.
                </p>
              </div>
              <div className="pt-8">
                <button 
                  onClick={handleFinish}
                  disabled={isSaving}
                  className="bg-primary text-primary-foreground px-10 py-3.5 rounded-full font-medium text-lg shadow-lg hover:shadow-xl hover:opacity-90 transition-all hover:scale-105 disabled:opacity-50 disabled:scale-100 flex items-center gap-2 mx-auto"
                >
                  {isSaving ? <Loader2 className="w-5 h-5 animate-spin" /> : null}
                  Enter Dashboard
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
