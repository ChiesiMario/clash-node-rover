import { useEffect, useState } from "react";
import { listen } from "@tauri-apps/api/event";
import { Zap, WifiOff, Star } from "lucide-react";

interface NodeResult {
  name: string;
  delay: number | null;
  is_active: boolean;
}

export function NodeRanking() {
  const [nodes, setNodes] = useState<NodeResult[]>([]);

  useEffect(() => {
    const unlisten = listen<NodeResult[]>("node_results", (event) => {
      setNodes(event.payload);
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);

  if (nodes.length === 0) {
    return (
      <div className="p-8 rounded-xl border border-border bg-card/50 text-center text-muted-foreground">
        Waiting for next speed test cycle...
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold tracking-tight">Live Node Rankings</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {nodes.map((node) => (
          <div
            key={node.name}
            className={`p-4 rounded-lg border text-sm flex items-center justify-between transition-colors
              ${
                node.is_active
                  ? "border-primary bg-primary/5 shadow-sm"
                  : "border-border bg-card hover:bg-muted/30"
              }`}
          >
            <div className="flex items-center gap-3 truncate pr-4">
              <div
                className={`w-2.5 h-2.5 rounded-full shrink-0 ${
                  node.delay === null
                    ? "bg-rose-500/50"
                    : node.delay < 150
                    ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]"
                    : node.delay < 300
                    ? "bg-amber-500"
                    : "bg-rose-500"
                }`}
              />
              <span className={`truncate font-medium ${node.is_active ? "text-primary" : "text-foreground"}`}>
                {node.name}
              </span>
              {node.is_active && (
                <Star className="w-3.5 h-3.5 shrink-0 text-amber-500 fill-amber-500" />
              )}
            </div>

            <div className="shrink-0 flex items-center gap-1.5 font-mono">
              {node.delay === null ? (
                <>
                  <WifiOff className="w-3.5 h-3.5 text-muted-foreground" />
                  <span className="text-muted-foreground">Timeout</span>
                </>
              ) : (
                <>
                  <Zap
                    className={`w-3.5 h-3.5 ${
                      node.delay < 150 ? "text-emerald-500" : "text-muted-foreground"
                    }`}
                  />
                  <span
                    className={
                      node.delay < 150
                        ? "text-emerald-600 dark:text-emerald-400 font-semibold"
                        : "text-muted-foreground"
                    }
                  >
                    {node.delay}ms
                  </span>
                </>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
