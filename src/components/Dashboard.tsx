import { useEffect, useState } from "react";
import { listen } from "@tauri-apps/api/event";

import { invoke } from "@tauri-apps/api/core";
import { CheckCircle2, XCircle, Clock, Play, Loader2, Pause, PauseCircle } from "lucide-react";
import { useTranslation } from "react-i18next";
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
  onNavigate: (tab: string) => void;
}

export function Dashboard({ status, onNavigate }: DashboardProps) {
  const { t } = useTranslation();
  const [apiUrl, setApiUrl] = useState<string>("");
  const [targetGroups, setTargetGroups] = useState<string[] | null>(null);

  useEffect(() => {
    const fetchConfig = () => {
      invoke<any>("get_config").then((cfg) => {
        if (cfg && cfg.api_url) {
          setApiUrl(cfg.api_url);
        }
        if (cfg && Array.isArray(cfg.target_groups)) {
          setTargetGroups(cfg.target_groups);
        }
      }).catch(console.error);
    };

    fetchConfig();

    const unlisten = listen("config_updated", () => {
      fetchConfig();
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);

  if (!status) {
    return (
      <div className="p-8 max-w-4xl mx-auto space-y-8 flex items-center justify-center min-h-[50vh]">
        <div className="flex flex-col items-center gap-4 text-muted-foreground">
          <Loader2 className="w-8 h-8 animate-spin" />
          <p>{t('dashboard.connecting', 'Connecting to background engine...')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-4xl mx-auto space-y-8">
      {/* Derived states */}
      {(() => {
        const isTargetEmpty = targetGroups !== null && targetGroups.length === 0;
        return (
          <>
            <div className="flex items-start justify-between">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">{t('dashboard.system_status', 'System Status')}</h1>
          <div className="flex flex-col gap-2">
            <p className="text-muted-foreground">{t('dashboard.subtitle', 'Monitor your proxy nodes in real-time.')}</p>
            {status.api_connected && apiUrl && (
              <div className="flex items-center">
                <span className="px-2 py-0.5 rounded-md text-xs font-mono bg-muted/50 text-muted-foreground border border-border flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_4px_rgba(16,185,129,0.5)] animate-pulse"></span>
                  {t('dashboard.connected_to', 'Connected to')} {apiUrl.replace(/^https?:\/\//, '')}
                </span>
              </div>
            )}
          </div>
        </div>
        <button
          onClick={() => invoke("toggle_pause")}
          disabled={isTargetEmpty}
          className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-colors ${
            status.is_paused 
              ? "bg-amber-500/10 text-amber-600 hover:bg-amber-500/20 dark:text-amber-400" 
              : isTargetEmpty
              ? "bg-muted/50 text-muted-foreground opacity-50 cursor-not-allowed"
              : "bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
        >
          {status.is_paused ? (
            <>
              <Play className="w-4 h-4 fill-current" />
              {t('dashboard.resume_engine', 'Resume Engine')}
            </>
          ) : (
            <>
              <Pause className="w-4 h-4 fill-current" />
              {t('dashboard.pause_engine', 'Pause Engine')}
            </>
          )}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Status Card 1: Connection */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">{t('dashboard.api_connection', 'API Connection')}</h3>
            {status.api_connected ? (
              <CheckCircle2 className="w-5 h-5 text-emerald-500" />
            ) : (
              <XCircle className="w-5 h-5 text-rose-500" />
            )}
          </div>
          <div className="text-2xl font-semibold">
            {status.api_connected ? t('dashboard.connected', 'Connected') : t('dashboard.disconnected', 'Disconnected')}
          </div>
        </div>

        {/* Status Card 2: Engine State */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">{t('dashboard.engine_state', 'Engine State')}</h3>

            <ActivityIcon isTesting={status.is_testing} isPaused={status.is_paused} isEmpty={isTargetEmpty} />
          </div>
          <div className="text-2xl font-semibold">
            {isTargetEmpty ? <span className="text-muted-foreground">{t('dashboard.state.no_groups', 'No Groups Set')}</span> : status.is_paused ? t('dashboard.state.paused', 'Paused') : status.is_testing ? t('dashboard.state.testing', 'Testing Nodes...') : t('dashboard.state.standby', 'Standby')}
          </div>
        </div>

        {/* Status Card 3: Next Check */}
        <div className="p-6 rounded-xl border border-border bg-muted/30 space-y-4 transition-colors hover:bg-muted/50">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-sm text-muted-foreground">{t('dashboard.next_check', 'Next Check')}</h3>
            <Clock className="w-5 h-5 text-blue-500" />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-2xl font-semibold tabular-nums">
              {isTargetEmpty ? "--" : status.next_check_in > 0 ? `${status.next_check_in}s` : "--"}
            </div>
            {!status.is_paused && (
              <button
                onClick={() => invoke("force_test")}
                disabled={status.is_testing || !status.api_connected || isTargetEmpty}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {status.is_testing ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    {t('dashboard.testing', 'Testing...')}
                  </>
                ) : (
                  <>
                    <Play className="w-3.5 h-3.5" />
                    {t('dashboard.test_now', 'Test Now')}
                  </>
                )}
              </button>
            )}
          </div>
        </div>
      </div>

      <NodeRanking isTesting={status.is_testing} targetGroups={targetGroups} onNavigate={onNavigate} />
          </>
        );
      })()}
    </div>
  );
}

function ActivityIcon({ isTesting, isPaused, isEmpty }: { isTesting: boolean, isPaused: boolean, isEmpty?: boolean }) {
  if (isEmpty) {
    return <div className="h-5 w-5 rounded-full bg-muted border border-dashed border-muted-foreground/30"></div>;
  }
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
