import { useEffect, useState, Fragment } from "react";
import { listen } from "@tauri-apps/api/event";
import { invoke } from "@tauri-apps/api/core";
import { Zap, WifiOff, Loader2, LineChart as LineChartIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { NodeHistoryChart } from "./NodeHistoryChart";
import { CustomNodeSelect, getColorClass, getJitterColorClass, NodeResult } from "./NodeSelect";

function TimeAgo({ timestamp, t }: { timestamp: number; t: any }) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    // Update every second to show real-time seconds
    const interval = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(interval);
  }, []);

  const diffSeconds = Math.floor((now - timestamp) / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  
  if (diffMinutes < 1) {
    return <span className="text-[10px] text-muted-foreground/70 text-center block">{t('ranking.seconds_ago', '{{count}} 秒前', { count: Math.max(0, diffSeconds) })}</span>;
  }
  
  return <span className="text-[10px] text-muted-foreground/70 text-center block">{t('ranking.minutes_ago', '{{count}} 分鐘前', { count: diffMinutes })}</span>;
}

interface GroupResult {
  group_name: string;
  nodes: NodeResult[];
  is_locked: boolean;
  selected_regions?: string[];
}

const AVAILABLE_REGIONS = ["US", "JP", "HK", "SG", "TW", "KR", "UK"];

interface NodeRankingProps {
  isTesting?: boolean;
  targetGroups?: string[] | null;
  onNavigate?: (tab: string) => void;
}



export function NodeRanking({ isTesting, targetGroups, onNavigate }: NodeRankingProps = {}) {
  const { t } = useTranslation();
  const [groups, setGroups] = useState<GroupResult[]>([]);
  const [selectedNodes, setSelectedNodes] = useState<Record<string, string>>({});
  const [expandedNode, setExpandedNode] = useState<string | null>(null);
  const [lastSwitchTimes, setLastSwitchTimes] = useState<Record<string, number>>({});
  const [currentActiveNodes, setCurrentActiveNodes] = useState<Record<string, string>>({});

  useEffect(() => {
    // Detect active node changes
    setGroups((prevGroups) => {
      // We do this logic outside of setGroups usually, but since groups is state, we can use an effect that depends on groups.
      return prevGroups;
    });
  }, []);

  useEffect(() => {
    let changed = false;
    const newActiveNodes = { ...currentActiveNodes };
    const newSwitchTimes = { ...lastSwitchTimes };
    const now = Date.now();

    groups.forEach(g => {
      const active = g.nodes.find(n => n.is_active);
      if (active) {
        const prev = newActiveNodes[g.group_name];
        if (prev !== active.name) {
          newActiveNodes[g.group_name] = active.name;
          newSwitchTimes[g.group_name] = now;
          changed = true;
        }
      }
    });

    if (changed) {
      setCurrentActiveNodes(newActiveNodes);
      setLastSwitchTimes(newSwitchTimes);
    }
  }, [groups]);

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

  const handleToggleRegion = async (groupName: string, region: string) => {
    try {
      await invoke("toggle_group_region", { group: groupName, region });
      setGroups((prev) =>
        prev.map((g) => {
          if (g.group_name === groupName) {
            const currentRegions = g.selected_regions || [];
            const newRegions = currentRegions.includes(region)
              ? currentRegions.filter((r) => r !== region)
              : [...currentRegions, region];
            return { ...g, selected_regions: newRegions };
          }
          return g;
        })
      );
    } catch (error) {
      console.error("Failed to toggle region:", error);
    }
  };

  const handleManualSwitch = async (groupName: string, nodeName: string) => {
    if (!nodeName) return;
    try {
      await invoke("manual_switch", { group: groupName, node: nodeName });
      setGroups((prev) =>
        prev.map((g) => {
          if (g.group_name === groupName) {
            return {
              ...g,
              is_locked: true,
              nodes: g.nodes.map(n => ({ ...n, is_active: n.name === nodeName }))
            };
          }
          return g;
        })
      );
      setSelectedNodes((prev) => ({ ...prev, [groupName]: nodeName }));
    } catch (error) {
      console.error("Failed to switch node:", error);
    }
  };

  const isTargetEmpty = targetGroups !== null && targetGroups?.length === 0;

  if (isTargetEmpty) {
    return (
      <div className="space-y-8 animate-in fade-in duration-300">
        <div className="space-y-3">
          <h2 className="text-xl font-semibold tracking-tight">{t('ranking.monitored_groups', 'Monitored Groups')}</h2>
          <div className="p-8 rounded-xl border border-dashed border-border bg-card/30 flex flex-col items-center justify-center gap-4">
            <div className="text-muted-foreground text-center">
              <p className="font-medium text-foreground mb-1">{t('ranking.no_groups_configured', 'No groups configured')}</p>
              <p className="text-sm">{t('ranking.no_groups_desc', "You haven't added any Clash proxy groups to monitor yet.")}</p>
            </div>
            <button 
              onClick={() => onNavigate?.("settings")}
              className="bg-primary text-primary-foreground px-4 py-2 rounded-md font-medium text-sm transition-opacity hover:opacity-90 shadow-sm"
            >
              {t('ranking.go_to_settings', 'Go to Settings')}
            </button>
          </div>
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold tracking-tight">{t('ranking.node_ranking', 'Node Ranking')}</h2>
            <span className="text-sm text-muted-foreground">{t('ranking.0_nodes', '0 nodes')}</span>
          </div>
          <div className="p-8 rounded-xl border border-dashed border-border bg-card/30 flex flex-col items-center justify-center gap-4">
            <div className="text-muted-foreground text-center text-sm">
              <p>{t('ranking.add_groups_desc', 'Add some groups in Settings to discover and rank your nodes here.')}</p>
            </div>
            <button 
              onClick={() => onNavigate?.("settings")}
              className="bg-primary text-primary-foreground px-4 py-2 rounded-md font-medium text-sm transition-opacity hover:opacity-90 shadow-sm"
            >
              {t('ranking.go_to_settings', 'Go to Settings')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Removed the early return for groups.length === 0

  // Deduplicate and aggregate nodes
  const allNodesMap = new Map<string, {
    name: string;
    delay: number | null;
    mean?: number | null;
    jitter?: number;
    provider?: string;
    backoff_rounds?: number | null;
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
          backoff_rounds: node.backoff_rounds,
          activeInGroups: []
        });
      }
      
      if (node.is_active) {
        allNodesMap.get(node.name)!.activeInGroups.push(group.group_name);
      }
    });
  });

  const displayGroups = [...groups];
  if (targetGroups) {
    targetGroups.forEach(tg => {
      if (!displayGroups.find(g => g.group_name === tg)) {
        displayGroups.push({
          group_name: tg,
          nodes: [],
          is_locked: false,
        });
      }
    });
  }

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
        <h2 className="text-xl font-semibold tracking-tight">{t('ranking.monitored_groups', 'Monitored Groups')}</h2>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(200px,1fr))] gap-4 items-start">
          {displayGroups.map((group) => {
            const activeNode = group.nodes.find((n) => n.is_active);
            const currentValue = selectedNodes[group.group_name] || (activeNode ? activeNode.name : "");
            const hasNodes = group.nodes.length > 0;

            return (
              <div key={group.group_name} className="flex flex-col justify-between gap-3 bg-muted/30 p-4 rounded-xl border border-border transition-colors hover:bg-muted/50">
                <div className="flex flex-col gap-3">
                  <div className="flex items-center justify-between gap-2 border-b border-border/50 pb-2">
                    <h3 className="font-semibold truncate">{group.group_name}</h3>
                  </div>
                  <div className="flex flex-col gap-1.5 flex-1 min-w-0">
                    {/* Row 1: Delay and TimeAgo */}
                    <div className="flex items-center justify-between w-full">
                      <div className="shrink-0 flex items-center gap-1">
                        {!hasNodes || activeNode?.delay === null || activeNode?.delay === undefined ? (
                          <Zap className="w-4 h-4 text-muted-foreground/50" />
                        ) : (
                          <>
                            <Zap className={`w-3.5 h-3.5 ${getColorClass(activeNode.delay, "text")}`} />
                            <span className={`font-mono text-[13px] font-bold ${getColorClass(activeNode.delay, "text")}`}>{activeNode.delay}</span>
                          </>
                        )}
                      </div>
                      {lastSwitchTimes[group.group_name] && (
                        <div className="shrink-0">
                          <TimeAgo timestamp={lastSwitchTimes[group.group_name]} t={t} />
                        </div>
                      )}
                    </div>

                    {/* Row 2: Node Name */}
                    <div className="flex items-center w-full min-w-0">
                      <span className={`font-medium truncate ${!hasNodes ? "text-muted-foreground animate-pulse" : "text-foreground"}`}>
                        {!hasNodes ? t('ranking.syncing_nodes', 'Syncing nodes...') : activeNode ? activeNode.name : t('ranking.waiting', 'Waiting...')}
                      </span>
                    </div>

                    {/* Row 3: Provider */}
                    {activeNode?.provider && (
                      <div className="flex items-center mt-0.5">
                        <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md border border-border/50 truncate max-w-full">
                          {activeNode.provider}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex flex-col gap-3 pt-1">
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
                      {t('ranking.auto_switch', 'Auto Switch')}
                    </button>
                    <button
                      onClick={() => !group.is_locked && handleToggleLock(group.group_name, false)}
                      className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-all ${
                        group.is_locked 
                          ? "bg-background shadow-sm text-amber-600 dark:text-amber-500" 
                          : "text-muted-foreground hover:text-foreground hover:bg-background/50"
                      }`}
                    >
                      {t('ranking.manual_switch', 'Manual Switch')}
                    </button>
                  </div>

                  {/* Manual Selection Dropdown */}
                  {group.is_locked ? (
                    <div className="flex items-center gap-2 w-full animate-in slide-in-from-top-2 duration-200 fade-in">
                      {hasNodes ? (
                        <CustomNodeSelect
                          nodes={group.nodes}
                          value={currentValue}
                          onChange={(val) => handleManualSwitch(group.group_name, val)}
                        />
                      ) : (
                        <div className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm text-muted-foreground shadow-sm">
                          {t('ranking.loading_nodes', 'Loading nodes...')}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-2 pt-2 animate-in slide-in-from-top-2 duration-200 fade-in">
                      <span className="text-[11px] font-medium text-muted-foreground w-full -mb-0.5">{t('ranking.region_filter', 'Region Filter:')}</span>
                      {AVAILABLE_REGIONS.map((region) => {
                        const isSelected = (group.selected_regions || []).includes(region);
                        return (
                          <button
                            key={region}
                            onClick={() => handleToggleRegion(group.group_name, region)}
                            className={`px-2.5 py-1 rounded-full text-[11px] font-medium transition-colors border ${
                              isSelected
                                ? "bg-primary text-primary-foreground border-primary shadow-sm"
                                : "bg-transparent text-muted-foreground border-border hover:border-primary/50 hover:text-foreground"
                            }`}
                          >
                            {region}
                          </button>
                        );
                      })}
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
          <h2 className="text-xl font-semibold tracking-tight">{t('ranking.node_ranking', 'Node Ranking')}</h2>
          <span className="text-sm text-muted-foreground">{t('ranking.n_nodes', '{{count}} nodes', { count: unifiedNodes.length })}</span>
        </div>
        
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 font-medium w-24 whitespace-nowrap text-center">{t('ranking.status', 'Status')}</th>
                  <th className="px-4 py-3 font-medium">{t('ranking.node_name', 'Node Name')}</th>
                  <th className="px-4 py-3 font-medium w-32">{t('ranking.score', 'Score')}</th>
                  <th className="px-4 py-3 font-medium w-32">{t('ranking.mean_jitter', 'Mean/Jitter')}</th>
                  <th className="px-4 py-3 font-medium w-48">{t('ranking.active_in_groups', 'Active In Groups')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {unifiedNodes.map((node) => (
                  <Fragment key={node.name}>
                    <tr 
                      className={`transition-colors cursor-pointer hover:bg-muted/30 ${node.activeInGroups.length > 0 ? "bg-primary/5" : ""} ${expandedNode === node.name ? "bg-muted/50" : ""}`}
                      onClick={() => setExpandedNode(expandedNode === node.name ? null : node.name)}
                    >
                    <td className="px-4 py-3 text-center">
                      <div className="flex justify-center items-center h-full">
                        {isTesting ? (
                          <Loader2 className="w-4 h-4 animate-spin text-muted-foreground/70" />
                        ) : (
                          <div
                            className={`w-2.5 h-2.5 rounded-full ${getColorClass(node.delay, "bg")}`}
                          />
                        )}
                      </div>
                    </td>
                    <td className={`px-4 py-3 min-w-[200px] max-w-[400px] font-medium ${node.activeInGroups.length > 0 ? "text-primary" : "text-foreground"}`}>
                      <div className="flex flex-col gap-1.5">
                        <div className="flex items-center gap-1.5">
                          <span className="truncate">{node.name}</span>
                          <LineChartIcon className="w-3.5 h-3.5 text-muted-foreground/50 opacity-0 group-hover:opacity-100 transition-opacity" />
                        </div>
                        {(node.provider || (node.backoff_rounds && node.backoff_rounds > 0)) && (
                          <div className="flex items-center gap-1.5">
                            {node.provider && (
                              <span className="text-[10px] font-medium bg-muted text-muted-foreground px-1.5 py-0.5 rounded-md border border-border/50 whitespace-nowrap">
                                {node.provider}
                              </span>
                            )}
                            {node.backoff_rounds && node.backoff_rounds > 0 && (
                              <span className="text-[10px] font-medium bg-rose-500/10 text-rose-500 px-1.5 py-0.5 rounded-md border border-rose-500/20 whitespace-nowrap">
                                Backoff: {node.backoff_rounds}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 font-mono">
                      {node.delay === null ? (
                        <div className={`flex items-center gap-1.5 ${isTesting && (!node.backoff_rounds || node.backoff_rounds === 0) ? "text-muted-foreground" : "text-rose-500 dark:text-rose-400 font-medium"}`}>
                          {isTesting && (!node.backoff_rounds || node.backoff_rounds === 0) ? (
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                          ) : (
                            <WifiOff className="w-3.5 h-3.5" />
                          )}
                          <span>
                            {isTesting && (!node.backoff_rounds || node.backoff_rounds === 0)
                              ? t('ranking.testing', 'Testing...')
                              : t('ranking.timeout', 'Timeout')}
                          </span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1.5">
                          <Zap
                            className={`w-3.5 h-3.5 ${getColorClass(node.delay, "text")}`}
                          />
                          <span
                            className={`font-semibold ${getColorClass(node.delay, "text")}`}
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
                          <span className={getJitterColorClass(node.jitter)}>±{node.jitter}ms</span> <span className="opacity-50">jit</span>
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
                    {expandedNode === node.name && (
                      <tr>
                        <td colSpan={5} className="p-0 border-b border-border/50">
                          <div className="animate-in slide-in-from-top-2 duration-200">
                            <NodeHistoryChart nodeName={node.name} />
                          </div>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

    </div>
  );
}
