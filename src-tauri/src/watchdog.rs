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

            // 1. 檢查 API 連線
            if clash.verify_connection().await.is_ok() {
                status.api_connected = true;
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

            for group in &config.target_groups {
                if let Ok(detail) = clash.get_proxy_group(group).await {
                    if let Some(nodes) = detail.all {
                        let current_node = detail.now.clone().unwrap_or_default();
                        
                        // 建立並發測速任務
                        let mut tasks = FuturesUnordered::new();
                        for node in nodes.clone() {
                            let clash_client = clash.clone();
                            let test_url = config.test_urls.first().unwrap_or(&"http://www.gstatic.com/generate_204".to_string()).clone();
                            let timeout = config.test_timeout;
                            tasks.push(async move {
                                let res = clash_client.ping_proxy(&node, &test_url, timeout).await;
                                (node, res)
                            });
                        }

                        let mut results = Vec::new();
                        while let Some(res) = tasks.next().await {
                            results.push(res);
                        }

                        // 分析測速結果
                        let mut fastest_node = None;
                        let mut min_delay = u64::MAX;
                        let mut current_node_delay = u64::MAX;

                        for (node, result) in &results {
                            match result {
                                Ok(delay) => {
                                    if *delay < min_delay {
                                        min_delay = *delay;
                                        fastest_node = Some(node.clone());
                                    }
                                    if node == &current_node {
                                        current_node_delay = *delay;
                                    }
                                }
                                Err(_) => {} // Timeout or error
                            }
                        }

                        // Emit Node Results
                        #[derive(serde::Serialize, Clone)]
                        struct NodeResult {
                            name: String,
                            delay: Option<u64>,
                            is_active: bool,
                        }
                        
                        let mut final_results: Vec<NodeResult> = results.iter().map(|(n, r)| NodeResult {
                            name: n.clone(),
                            delay: r.clone().ok(),
                            is_active: n == &current_node,
                        }).collect();
                        
                        // Sort by delay, put timeouts at the bottom
                        final_results.sort_by(|a, b| {
                            match (a.delay, b.delay) {
                                (Some(d1), Some(d2)) => d1.cmp(&d2),
                                (Some(_), None) => std::cmp::Ordering::Less,
                                (None, Some(_)) => std::cmp::Ordering::Greater,
                                (None, None) => std::cmp::Ordering::Equal,
                            }
                        });

                        let _ = app.emit("node_results", &final_results);

                        let db = app.state::<Db>();
                        
                        // Jitter 容忍度邏輯判斷
                        if let Some(fastest) = fastest_node {
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
                                    let _ = app.emit("node_results", &final_results);
                                    
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
                    }
                }
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
