import { useEffect, useState } from "react";
import { listen } from "@tauri-apps/api/event";
import { invoke } from "@tauri-apps/api/core";
import { Zap, WifiOff, Star, Lock, Unlock, Check, Loader2 } from "lucide-react";

interface NodeResult {
  name: string;
  delay: number | null; // This is the Score now
  mean?: number | null;
  jitter?: number;
  is_active: boolean;
  provider?: string;
}

interface GroupResult {
  group_name: string;
  nodes: NodeResult[];
  is_locked: boolean;
}

interface NodeRankingProps {
  isTesting?: boolean;
}

export function NodeRanking({ isTesting }: NodeRankingProps = {}) {
  const [groups, setGroups] = useState<GroupResult[]>([]);
  const [selectedNodes, setSelectedNodes] = useState<Record<string, string>>({});

  useEffect(() => {
    // Fetch initial state
    invoke<GroupResult[]>("get_latest_results").then((initialGroups) => {
      if (initialGroups.length > 0) {
        setGroups(initialGroups);
      }
    }).catch(console.error);

    const unlisten = listen<GroupResult[]>("node_results", (event) => {
      setGroups(event.payload);
    });

    return () => {
      unlisten.then((f) => f());
    };
  }, []);

  const handleToggleLock = async (groupName: string, isLocked: boolean) => {
    try {
      await invoke("toggle_group_lock", { group: groupName, locked: !isLocked });
      setGroups((prev) =>
        prev.map((g) => (g.group_name === groupName ? { ...g, is_locked: !isLocked } : g))
      );
    } catch (error) {
      console.error("Failed to toggle lock:", error);
    }
  };

  const handleManualSwitch = async (groupName: string, nodeName: string) => {
    if (!nodeName) return;
    try {
      await invoke("manual_switch", { group: groupName, node: nodeName });
      setGroups((prev) =>
        prev.map((g) => (g.group_name === groupName ? { ...g, is_locked: true } : g))
      );
      setSelectedNodes((prev) => ({ ...prev, [groupName]: nodeName }));
    } catch (error) {
      console.error("Failed to switch node:", error);
    }
  };

  if (groups.length === 0) {
    return (
      <div className="p-8 rounded-xl border border-border bg-card/50 text-center text-muted-foreground">
        Waiting for next speed test cycle...
      </div>
    );
  }

  // Deduplicate and aggregate nodes
  const allNodesMap = new Map<string, {
    name: string;
    delay: number | null;
    mean?: number | null;
    jitter?: number;
    provider?: string;
    activeInGroups: string[];
  }>();

  groups.forEach(group => {
    group.nodes.forEach(node => {
      if (!allNodesMap.has(node.name)) {
        allNodesMap.set(node.name, {
          name: node.name,
          delay: node.delay,
          mean: node.mean,
          jitter: node.jitter,
          provider: node.provider,
          activeInGroups: []
        });
      }
      
      if (node.is_active) {
        allNodesMap.get(node.name)!.activeInGroups.push(group.group_name);
      }
    });
  });

  const unifiedNodes = Array.from(allNodesMap.values()).sort((a, b) => {
    if (a.delay === null && b.delay === null) return 0;
    if (a.delay === null) return 1;
    if (b.delay === null) return -1;
    return a.delay - b.delay;
  });

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      
      {/* Groups Section */}
      <div className="space-y-3">
        <h2 className="text-xl font-semibold tracking-tight">Monitored Groups</h2>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-4">
          {groups.map((group) => {
            const activeNode = group.nodes.find((n) => n.is_active);
            const currentValue = selectedNodes[group.group_name] || (activeNode ? activeNode.name : "");

            return (
              <div key={group.group_name} className="flex flex-col justify-between gap-3 bg-muted/30 p-4 rounded-xl border border-border transition-colors hover:bg-muted/50 overflow-hidden">
                <div className="flex flex-col gap-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="font-semibold truncate">{group.group_name}</h3>
                  </div>
                  <div className="flex flex-col gap-1 flex-1 min-w-0">
                    <div className="text-sm text-muted-foreground flex items-center gap-1.5">
                      <Zap className="w-3.5 h-3.5 shrink-0 text-emerald-500" />
                      <div className="font-medium text-sm truncate text-foreground flex-1">
                        {activeNode ? activeNode.name : "None"}
                      </div>
                    </div>
                    {activeNode?.provider && (
                      <div className="flex pl-5">
                        <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md truncate max-w-full border border-border/50">
                          {activeNode.provider}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex flex-col gap-3 pt-3 mt-1 border-t border-border/50">
                  {/* Mode Toggle Buttons */}
                  <div className="flex bg-background/50 p-1 rounded-lg border border-border">
                    <button
                      onClick={() => group.is_locked && handleToggleLock(group.group_name, true)}
                      className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-all ${
                        !group.is_locked 
                          ? "bg-background shadow-sm text-foreground" 
                          : "text-muted-foreground hover:text-foreground hover:bg-background/50"
                      }`}
                    >
                      自動切換
                    </button>
                    <button
                      onClick={() => !group.is_locked && handleToggleLock(group.group_name, false)}
                      className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-all ${
                        group.is_locked 
                          ? "bg-background shadow-sm text-amber-600 dark:text-amber-500" 
                          : "text-muted-foreground hover:text-foreground hover:bg-background/50"
                      }`}
                    >
                      手動切換
                    </button>
                  </div>

                  {/* Manual Selection Dropdown */}
                  {group.is_locked && (
                    <div className="flex items-center gap-2 w-full animate-in slide-in-from-top-2 duration-200 fade-in">
                      <select
                        className="bg-background border border-border rounded-md px-2 py-1.5 text-sm focus:outline-none focus:border-amber-500/50 flex-1 min-w-0 truncate transition-colors cursor-pointer hover:border-border/80"
                        value={currentValue}
                        onChange={(e) => handleManualSwitch(group.group_name, e.target.value)}
                      >
                        {group.nodes.map((node) => (
                          <option key={node.name} value={node.name}>
                            {node.provider ? `[${node.provider}] ` : ""}{node.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Unified Nodes Ranking Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold tracking-tight">Node Ranking</h2>
          <span className="text-sm text-muted-foreground">{unifiedNodes.length} nodes</span>
        </div>
        
        <div className="rounded-xl border border-border bg-card overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 font-medium w-12 text-center">Status</th>
                  <th className="px-4 py-3 font-medium">Node Name</th>
                  <th className="px-4 py-3 font-medium w-32">Score</th>
                  <th className="px-4 py-3 font-medium w-32">Mean/Jitter</th>
                  <th className="px-4 py-3 font-medium w-48">Active In Groups</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {unifiedNodes.map((node) => (
                  <tr 
                    key={node.name} 
                    className={`transition-colors hover:bg-muted/30 ${node.activeInGroups.length > 0 ? "bg-primary/5" : ""}`}
                  >
                    <td className="px-4 py-3 text-center">
                      <div className="flex justify-center">
                        <div
                          className={`w-2.5 h-2.5 rounded-full ${
                            node.delay === null
                              ? isTesting 
                                ? "bg-muted-foreground/50 animate-pulse" 
                                : "bg-rose-500/50"
                              : node.delay < 150
                              ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]"
                              : node.delay < 300
                              ? "bg-amber-500"
                              : "bg-rose-500"
                          }`}
                        />
                      </div>
                    </td>
                    <td className={`px-4 py-3 min-w-[200px] max-w-[400px] font-medium ${node.activeInGroups.length > 0 ? "text-primary" : "text-foreground"}`}>
                      <div className="flex flex-col gap-1.5">
                        <span className="truncate">{node.name}</span>
                        {node.provider && (
                          <div className="flex">
                            <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md truncate max-w-full border border-border/50 whitespace-nowrap">
                              {node.provider}
                            </span>
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 font-mono">
                      {node.delay === null ? (
                        isTesting ? (
                          <div className="flex items-center gap-1.5 text-muted-foreground">
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                            <span>Testing...</span>
                          </div>
                        ) : (
                          <div className="flex items-center gap-1.5 text-muted-foreground">
                            <WifiOff className="w-3.5 h-3.5" />
                            <span>Timeout</span>
                          </div>
                        )
                      ) : (
                        <div className="flex items-center gap-1.5">
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
                            title="Score = Mean + Jitter"
                          >
                            {node.delay}
                          </span>
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {node.delay !== null && node.mean !== undefined && node.jitter !== undefined && (
                        <div>
                          <span className="text-foreground/70">{node.mean}ms</span> <span className="opacity-50">avg</span>
                          <br />
                          <span className="text-amber-500/80">±{node.jitter}ms</span> <span className="opacity-50">jit</span>
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {node.activeInGroups.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {node.activeInGroups.map(g => (
                            <span key={g} className="text-[11px] font-medium px-2 py-1 rounded-md bg-primary/10 text-primary">
                              {g}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-muted-foreground/50 text-xs">-</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

    </div>
  );
}
