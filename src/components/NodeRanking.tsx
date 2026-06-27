import { useEffect, useState } from "react";
import { listen } from "@tauri-apps/api/event";
import { invoke } from "@tauri-apps/api/core";
import { Zap, WifiOff, Star, Lock, Unlock, Check } from "lucide-react";

interface NodeResult {
  name: string;
  delay: number | null; // This is the Score now
  mean?: number | null;
  jitter?: number | null;
  is_active: boolean;
}

interface GroupResult {
  group_name: string;
  nodes: NodeResult[];
  is_locked: boolean;
}

export function NodeRanking() {
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

  const handleManualSwitch = async (groupName: string) => {
    const node = selectedNodes[groupName];
    if (!node) return;
    try {
      await invoke("manual_switch", { group: groupName, node });
      setGroups((prev) =>
        prev.map((g) => (g.group_name === groupName ? { ...g, is_locked: true } : g))
      );
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
    <div className="space-y-8 animate-in fade-in duration-500">
      
      {/* Groups Section */}
      <div className="space-y-3">
        <h2 className="text-xl font-semibold tracking-tight">Monitored Groups</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {groups.map((group) => {
            const activeNode = group.nodes.find((n) => n.is_active);
            const currentValue = selectedNodes[group.group_name] || (activeNode ? activeNode.name : "");

            return (
              <div key={group.group_name} className="flex flex-col justify-between gap-3 bg-muted/30 p-4 rounded-xl border border-border transition-colors hover:bg-muted/50 overflow-hidden">
                <div className="flex flex-col gap-1.5">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="font-semibold truncate">{group.group_name}</h3>
                    <button
                      onClick={() => handleToggleLock(group.group_name, group.is_locked)}
                      className={`p-1.5 rounded-md transition-colors shrink-0 ${
                        group.is_locked
                          ? "bg-amber-500/10 text-amber-500 hover:bg-amber-500/20"
                          : "bg-background/50 hover:bg-background text-muted-foreground border"
                      }`}
                      title={group.is_locked ? "解鎖群組 (允許自動切換)" : "鎖定群組 (停止自動切換)"}
                    >
                      {group.is_locked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                    </button>
                  </div>
                  <div className="text-sm text-muted-foreground flex items-center gap-1.5">
                    <Zap className="w-3.5 h-3.5 shrink-0 text-emerald-500" />
                    <span className="text-foreground font-medium truncate">{activeNode ? activeNode.name : "None"}</span>
                  </div>
                </div>

                <div className="flex items-center gap-2 w-full pt-3 mt-1 border-t border-border/50">
                  <select
                    className="bg-background border border-border rounded-md px-2 py-1.5 text-sm focus:outline-none focus:border-primary flex-1 min-w-0 truncate"
                    value={currentValue}
                    onChange={(e) =>
                      setSelectedNodes((prev) => ({ ...prev, [group.group_name]: e.target.value }))
                    }
                  >
                    {group.nodes.map((node) => (
                      <option key={node.name} value={node.name}>
                        {node.name}
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={() => handleManualSwitch(group.group_name)}
                    disabled={!currentValue}
                    className="flex items-center justify-center shrink-0 w-8 h-8 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
                    title="確認切換"
                  >
                    <Check className="w-4 h-4" />
                  </button>
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
                              ? "bg-rose-500/50"
                              : node.delay < 150
                              ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]"
                              : node.delay < 300
                              ? "bg-amber-500"
                              : "bg-rose-500"
                          }`}
                        />
                      </div>
                    </td>
                    <td className={`px-4 py-3 font-medium ${node.activeInGroups.length > 0 ? "text-primary" : "text-foreground"}`}>
                      {node.name}
                    </td>
                    <td className="px-4 py-3 font-mono">
                      {node.delay === null ? (
                        <div className="flex items-center gap-1.5 text-muted-foreground">
                          <WifiOff className="w-3.5 h-3.5" />
                          <span>Timeout</span>
                        </div>
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
                        <div className="flex items-center gap-1.5 flex-wrap">
                          <Star className="w-3.5 h-3.5 text-amber-500 fill-amber-500" />
                          {node.activeInGroups.map(g => (
                            <span key={g} className="text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-primary/10 text-primary border border-primary/20">
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
