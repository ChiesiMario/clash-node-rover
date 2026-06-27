import { useEffect, useState } from "react";

import { listen } from "@tauri-apps/api/event";
import { invoke } from "@tauri-apps/api/core";
import { CheckCircle2, XCircle, Clock, Play, Loader2 } from "lucide-react";
import { NodeRanking } from "./NodeRanking";

interface AppStatus {
  api_connected: boolean;
  is_testing: boolean;
  next_check_in: number;
}

export function Dashboard() {
  const [status, setStatus] = useState<AppStatus>({
    api_connected: false,
    is_testing: false,
    next_check_in: 0,
  });

  useEffect(() => {
    invoke<AppStatus>("get_status").then((initialStatus) => {
      setStatus(initialStatus);
    }).catch(console.error);

    const unlisten = listen<AppStatus>("status_update", (event) => {
      setStatus(event.payload);
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);

  return (
    <div className="p-8 max-w-4xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">System Status</h1>
        <p className="text-muted-foreground">Monitor your proxy nodes in real-time.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Status Card 1: Connection */}
        <div className="p-6 rounded-xl border border-border bg-card shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">API Connection</h3>
            {status.api_connected ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-500" />
            ) : (
              <XCircle className="w-5 h-5 text-rose-500" />
            )}
          </div>
          <div className="text-2xl font-semibold">
            {status.api_connected ? "Connected" : "Disconnected"}
          </div>
        </div>

        {/* Status Card 2: Engine State */}
        <div className="p-6 rounded-xl border border-border bg-card shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">Engine State</h3>
            <ActivityIcon isTesting={status.is_testing} />
          </div>
          <div className="text-2xl font-semibold">
            {status.is_testing ? "Testing Nodes..." : "Standby"}
          </div>
        </div>

        {/* Status Card 3: Next Check */}
        <div className="p-6 rounded-xl border border-border bg-card shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">Next Check</h3>
            <Clock className="w-5 h-5 text-blue-500" />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-2xl font-semibold tabular-nums">
              {status.next_check_in > 0 ? `${status.next_check_in}s` : "--"}
            </div>
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
          </div>
        </div>
      </div>

      <NodeRanking />
    </div>
  );
}

function ActivityIcon({ isTesting }: { isTesting: boolean }) {
  if (isTesting) {
    return (
      <div className="relative flex h-5 w-5">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
        <span className="relative inline-flex rounded-full h-5 w-5 bg-amber-500"></span>
      </div>
    );
  }
  return <div className="h-5 w-5 rounded-full bg-muted border border-border"></div>;
}
