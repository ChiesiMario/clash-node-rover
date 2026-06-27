use std::time::Duration;
use tauri::{AppHandle, Manager, Emitter};
use tokio::time::sleep;
use futures::stream::{FuturesUnordered, StreamExt};
use crate::{AppState, db::Db, clash::ClashApi};

#[derive(Clone, serde::Serialize)]
pub struct AppStatus {
    pub api_connected: bool,
    pub is_testing: bool,
    pub next_check_in: u64,
}

#[derive(serde::Serialize, Clone)]
pub struct NodeResult {
    pub name: String,
    pub delay: Option<u64>,
    pub is_active: bool,
}

#[derive(serde::Serialize, Clone)]
pub struct GroupResult {
    pub group_name: String,
    pub nodes: Vec<NodeResult>,
    pub is_locked: bool,
}

pub fn start_watchdog(app: AppHandle) {
    tauri::async_runtime::spawn(async move {
        loop {
            // 讀取最新的 config
            let (config, force_test) = {
                let state = app.state::<AppState>();
                let c = state.config.lock().unwrap().clone();
                let f = state.force_test.clone();
                (c, f)
            };

            if config.api_url.is_empty() {
                sleep(Duration::from_secs(3)).await;
                continue;
            }

            let clash = ClashApi::new(&config.api_url, &config.api_secret);
            let mut status = AppStatus {
                api_connected: false,
                is_testing: false,
                next_check_in: 0,
            };

            let db = app.state::<Db>();
            
            // 1. 檢查 API 連線
            if clash.verify_connection().await.is_ok() {
                status.api_connected = true;
                db.insert_log("DEBUG", "API 連線檢查: 成功");
            } else {
                db.insert_log("DEBUG", "API 連線檢查: 失敗");
            }

            // 發送狀態更新給前端
            let _ = app.emit("status_update", &status);

            if !status.api_connected {
                sleep(Duration::from_secs(3)).await;
                continue;
            }

            // 2. 開始測速演算法
            status.is_testing = true;
            let _ = app.emit("status_update", &status);

            let mut all_group_results: Vec<GroupResult> = Vec::new();

            // 階段一：收集所有目標群組的節點資料
            use std::collections::{HashSet, HashMap};
            struct GroupInfo {
                group_name: String,
                nodes: Vec<String>,
                current_node: String,
            }
            let mut groups_info = Vec::new();
            let mut unique_nodes = HashSet::new();

            for group in &config.target_groups {
                if let Ok(detail) = clash.get_proxy_group(group).await {
                    if let Some(nodes) = detail.all {
                        let current_node = detail.now.unwrap_or_default();
                        for node in &nodes {
                            unique_nodes.insert(node.clone());
                        }
                        groups_info.push(GroupInfo {
                            group_name: group.clone(),
                            nodes,
                            current_node,
                        });
                    }
                }
            }

            if !groups_info.is_empty() {
                db.insert_log("DEBUG", &format!("開始測速來自 {} 個群組的 {} 個節點...", groups_info.len(), unique_nodes.len()));
            }

            // 階段二：對所有不重複的節點進行並發測速
            let test_url = config.test_urls.first().unwrap_or(&"http://www.gstatic.com/generate_204".to_string()).clone();
            let timeout = config.test_timeout;
            let max_concurrent = if config.max_concurrent > 0 { config.max_concurrent as usize } else { 10 };

            let node_stream = futures::stream::iter(unique_nodes.into_iter());
            let clash_client = clash.clone();
            
            let mut ping_results: HashMap<String, Result<u64, ()>> = HashMap::new();
            
            let mut tasks = node_stream.map(|node| {
                let client = clash_client.clone();
                let url = test_url.clone();
                async move {
                    let res = client.ping_proxy(&node, &url, timeout).await;
                    (node, res)
                }
            }).buffer_unordered(max_concurrent);

            while let Some((node, res)) = tasks.next().await {
                ping_results.insert(node, res);
            }

            // 階段三：分配測速結果至各群組進行決策
            for info in groups_info {
                let group = &info.group_name;
                let current_node = &info.current_node;
                
                let mut fastest_node = None;
                let mut min_delay = u64::MAX;
                let mut current_node_delay = u64::MAX;

                let mut final_results: Vec<NodeResult> = Vec::new();

                for node in &info.nodes {
                    let delay_opt = match ping_results.get(node) {
                        Some(Ok(d)) => Some(*d),
                        _ => None,
                    };

                    if let Some(delay) = delay_opt {
                        if delay < min_delay {
                            min_delay = delay;
                            fastest_node = Some(node.clone());
                        }
                        if node == current_node {
                            current_node_delay = delay;
                        }
                    }

                    final_results.push(NodeResult {
                        name: node.clone(),
                        delay: delay_opt,
                        is_active: node == current_node,
                    });
                }
                
                final_results.sort_by(|a, b| {
                    match (a.delay, b.delay) {
                        (Some(d1), Some(d2)) => d1.cmp(&d2),
                        (Some(_), None) => std::cmp::Ordering::Less,
                        (None, Some(_)) => std::cmp::Ordering::Greater,
                        (None, None) => std::cmp::Ordering::Equal,
                    }
                });

                let mut debug_msgs = Vec::new();
                for res in &final_results {
                    if let Some(d) = res.delay {
                        debug_msgs.push(format!("  - {}: {}ms", res.name, d));
                    } else {
                        debug_msgs.push(format!("  - {}: Timeout", res.name));
                    }
                }
                db.insert_log("DEBUG", &format!("群組 [{}] 節點結果:\n{}", group, debug_msgs.join("\n")));
                db.insert_log("DEBUG", &format!("群組 [{}] 決策變數: current={}, current_delay={}, fastest={:?}, min_delay={}", group, current_node, current_node_delay, fastest_node, min_delay));
                
                // Jitter 容忍度邏輯判斷
                let is_locked = config.locked_groups.contains(group);
                
                if is_locked {
                    db.insert_log("INFO", &format!("群組 [{}] 已鎖定，略過自動切換", group));
                } else if let Some(fastest) = fastest_node {
                    let diff = if current_node_delay == u64::MAX {
                        u64::MAX // Current node failed
                    } else {
                        current_node_delay.saturating_sub(min_delay)
                    };

                    if diff > config.tolerance_ms {
                        // 超過容忍度，進行切換
                        if let Ok(_) = clash.select_proxy(group, &fastest).await {
                            db.insert_log("INFO", &format!("群組 [{}] 切換至最快節點: {} (優勢 {}ms)", group, fastest, diff));
                            
                            // Update the active node in our emitted results to reflect the switch
                            for res in &mut final_results {
                                res.is_active = res.name == fastest;
                            }
                        } else {
                            db.insert_log("ERROR", &format!("群組 [{}] 切換至 {} 失敗", group, fastest));
                        }
                    } else {
                        // 在容忍度內，保持當前節點
                        db.insert_log("INFO", &format!("群組 [{}] 保持當前節點: {} (與最快差距 {}ms，容忍度內)", group, current_node, diff));
                    }
                } else {
                    db.insert_log("WARN", &format!("群組 [{}] 所有節點皆無回應 (Timeout: {}ms)", group, config.test_timeout));
                }

                all_group_results.push(GroupResult {
                    group_name: group.clone(),
                    nodes: final_results,
                    is_locked,
                });
            }

            let _ = app.emit("node_results", &all_group_results);
            if let Ok(mut last) = app.state::<AppState>().last_results.lock() {
                *last = all_group_results.clone();
            }

            status.is_testing = false;
            let _ = app.emit("status_update", &status);

            // 3. 進入休眠倒數
            let mut remaining = config.check_interval;
            while remaining > 0 {
                status.next_check_in = remaining;
                let _ = app.emit("status_update", &status);
                tokio::select! {
                    _ = force_test.notified() => {
                        break;
                    }
                    _ = sleep(Duration::from_secs(1)) => {
                        remaining -= 1;
                    }
                }
            }
        }
    });
}
