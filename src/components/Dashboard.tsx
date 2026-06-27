import { useEffect, useState } from "react";

import { listen } from "@tauri-apps/api/event";
import { invoke } from "@tauri-apps/api/core";
import { CheckCircle2, XCircle, Clock, Play, Loader2, Pause, PauseCircle } from "lucide-react";
import { NodeRanking } from "./NodeRanking";

export interface AppStatus {
  api_connected: boolean;
  is_testing: boolean;
  next_check_in: number;
  is_paused: boolean;
}

export interface NodeResult {
  name: string;
  latency?: number;
  jitter?: number;
  is_active: boolean;
  provider?: string;
}

interface DashboardProps {
  status: AppStatus | null;
}

export function Dashboard({ status }: DashboardProps) {
  const [apiUrl, setApiUrl] = useState<string>("");

  useEffect(() => {
    invoke<any>("get_config").then((cfg) => {
      if (cfg && cfg.api_url) {
        setApiUrl(cfg.api_url);
      }
    }).catch(console.error);
  }, []);

  if (!status) {
    return (
      <div className="p-8 max-w-4xl mx-auto space-y-8 flex items-center justify-center min-h-[50vh]">
        <div className="flex flex-col items-center gap-4 text-muted-foreground">
          <Loader2 className="w-8 h-8 animate-spin" />
          <p>Connecting to background engine...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-4xl mx-auto space-y-8">
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">System Status</h1>
          <p className="text-muted-foreground">Monitor your proxy nodes in real-time.</p>
        </div>
        <button
          onClick={() => invoke("toggle_pause")}
          className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-colors ${
            status.is_paused 
              ? "bg-amber-500/10 text-amber-600 hover:bg-amber-500/20 dark:text-amber-400" 
              : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          {status.is_paused ? (
            <>
              <Play className="w-4 h-4 fill-current" />
              Resume Engine
            </>
          ) : (
            <>
              <Pause className="w-4 h-4 fill-current" />
              Pause Engine
            </>
          )}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Status Card 1: Connection */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 shadow-sm space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">API Connection</h3>
            {status.api_connected ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-500" />
            ) : (
              <XCircle className="w-5 h-5 text-rose-500" />
            )}
          </div>
          <div className="flex flex-col">
            <div className="text-2xl font-semibold">
              {status.api_connected ? "Connected" : "Disconnected"}
            </div>
            {status.api_connected && apiUrl && (
              <span className="text-sm font-medium text-muted-foreground/70 mt-0.5 tracking-tight font-mono">
                {apiUrl}
              </span>
            )}
          </div>
        </div>

        {/* Status Card 2: Engine State */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 shadow-sm space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">Engine State</h3>

            <ActivityIcon isTesting={status.is_testing} isPaused={status.is_paused} />
          </div>
          <div className="text-2xl font-semibold">
            {status.is_paused ? "Paused" : status.is_testing ? "Testing Nodes..." : "Standby"}
          </div>
        </div>

        {/* Status Card 3: Next Check */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 shadow-sm space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">Next Check</h3>
            <Clock className="w-5 h-5 text-blue-500" />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-2xl font-semibold tabular-nums">
              {status.next_check_in > 0 ? `${status.next_check_in}s` : "--"}
            </div>
            {!status.is_paused && (
              <button
                onClick={() => invoke("force_test")}
                disabled={status.is_testing || !status.api_connected}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {status.is_testing ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    Testing...
                  </>
                ) : (
                  <>
                    <Play className="w-3.5 h-3.5" />
                    Test Now
                  </>
                )}
              </button>
            )}
          </div>
        </div>
      </div>

      <NodeRanking isTesting={status.is_testing} />
    </div>
  );
}

function ActivityIcon({ isTesting, isPaused }: { isTesting: boolean, isPaused: boolean }) {
  if (isPaused) {
    return <PauseCircle className="w-5 h-5 text-amber-500" />;
  }
  if (isTesting) {
    return (
      <div className="relative flex h-5 w-5">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
        <span className="relative inline-flex rounded-full h-5 w-5 bg-blue-500"></span>
      </div>
    );
  }
  return <div className="h-5 w-5 rounded-full bg-muted border border-border"></div>;
}
