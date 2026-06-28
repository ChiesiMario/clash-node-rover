import { useEffect, useState, useRef } from "react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { useTranslation } from "react-i18next";

interface LogEntry {
  id: number;
  timestamp: string;
  level: string;
  message: string;
}

export function Console() {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [showDebugLogs, setShowDebugLogs] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    invoke<LogEntry[]>("get_logs").then((initialLogs) => {
      setLogs(initialLogs);
    });

    const unlisten = listen<LogEntry>("new_log", (event) => {
      setLogs((prev) => [...prev, event.payload]);
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  const getLevelColor = (level: string) => {
    switch (level) {
      case "INFO": return "text-emerald-400";
      case "WARN": return "text-amber-400";
      case "ERROR": return "text-rose-400";
      case "DEBUG": return "text-purple-400";
      default: return "text-gray-400";
    }
  };

  const formatTime = (ts: string) => {
    return ts.split(" ")[1] || ts;
  };

  return (
    <div className="h-full flex flex-col p-8 max-w-5xl mx-auto space-y-6 animate-in fade-in duration-300">
      <div className="flex justify-between items-end">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">{t('console.title', 'System Logs')}</h1>
          <p className="text-muted-foreground">{t('console.subtitle', 'Detailed execution traces from the background engine.')}</p>
        </div>
        <button
          onClick={() => setShowDebugLogs(!showDebugLogs)}
          className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${showDebugLogs ? 'bg-purple-500/20 text-purple-300 border border-purple-500/30' : 'bg-gray-800 text-gray-400 border border-gray-700 hover:bg-gray-700'}`}
        >
          {showDebugLogs ? t('console.hide_debug', 'Hide Debug Logs') : t('console.show_debug', 'Show Debug Logs')}
        </button>
      </div>

      <div ref={scrollRef} className="flex-1 bg-[#0c0c0c] border border-border rounded-xl shadow-inner p-4 font-mono text-sm text-gray-300 overflow-auto flex flex-col gap-2">
        {logs.filter(log => showDebugLogs || log.level !== "DEBUG").map((log) => (
          <div key={log.id} className="flex gap-4">
            <span className="text-blue-400 whitespace-nowrap">{formatTime(log.timestamp)}</span>
            <span className={`whitespace-nowrap ${getLevelColor(log.level)}`}>[{log.level}]</span>
            <span className="whitespace-pre-wrap break-words">{log.message}</span>
          </div>
        ))}
        {logs.filter(log => showDebugLogs || log.level !== "DEBUG").length === 0 && (
          <div className="opacity-50 flex gap-4">
            <span>{t('console.no_logs', 'No logs available yet.')}</span>
          </div>
        )}
      </div>
    </div>
  );
}
