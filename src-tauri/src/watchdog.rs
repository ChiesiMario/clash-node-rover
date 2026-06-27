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
    pub delay: Option<u64>, // Now acts as Score
    pub mean: Option<u64>,
    pub jitter: Option<u64>,
    pub is_active: bool,
    pub provider: Option<String>,
}

#[derive(serde::Serialize, Clone)]
pub struct GroupResult {
    pub group_name: String,
    pub nodes: Vec<NodeResult>,
    pub is_locked: bool,
    pub selected_regions: Vec<String>,
}

fn matches_region(node_name: &str, region: &str) -> bool {
    let lower = node_name.to_lowercase();
    let keywords = match region {
        "US" => vec!["us", "united states", "美國", "美国", "洛杉磯", "洛杉矶", "波特蘭", "波特兰", "聖何塞", "圣何塞", "西雅圖", "西雅图", "芝加哥", "紐約", "纽约"],
        "JP" => vec!["jp", "japan", "日本", "東京", "东京", "大阪", "川崎"],
        "HK" => vec!["hk", "hong kong", "香港", "深港"],
        "SG" => vec!["sg", "singapore", "新加坡", "獅城", "狮城"],
        "TW" => vec!["tw", "taiwan", "台灣", "台湾", "台北", "彰化", "新北", "中華電信", "中华电信"],
        "KR" => vec!["kr", "korea", "韓國", "韩国", "首爾", "首尔"],
        "UK" => vec!["uk", "united kingdom", "英國", "英国", "倫敦", "伦敦"],
        _ => vec![],
    };
    keywords.iter().any(|k| lower.contains(k))
}

async fn perform_http_test(node_name: &str, config: &crate::config::Config, clash: &ClashApi, db: &Db) -> bool {
    db.insert_log("DEBUG", &format!("HTTP 測試: 準備測試節點 [{}]", node_name));
    if let Err(e) = clash.select_proxy(&config.dedicated_test_group, node_name).await {
        db.insert_log("ERROR", &format!("HTTP 測試: 無法切換專屬測速群組: {}", e));
        return false;
    }
    
    // 等待 500ms
    sleep(Duration::from_millis(500)).await;
    
    let proxy_url = if config.clash_proxy_url.starts_with("http") { config.clash_proxy_url.clone() } else { format!("http://{}", config.clash_proxy_url) };
    let proxy = match reqwest::Proxy::all(&proxy_url) {
        Ok(p) => p,
        Err(e) => { db.insert_log("ERROR", &format!("HTTP 測試: Proxy 解析失敗: {}", e)); return false; }
    };
    
    let client = match reqwest::Client::builder().proxy(proxy).timeout(Duration::from_secs(5)).build() {
        Ok(c) => c,
        Err(e) => { db.insert_log("ERROR", &format!("HTTP 測試: Client 建立失敗: {}", e)); return false; }
    };
    
    let mut futures = Vec::new();
    for url in &config.browser_test_urls {
        let c = client.clone();
        let u = url.clone();
        futures.push(tauri::async_runtime::spawn(async move {
            match c.get(&u).send().await {
                Ok(resp) => resp.status().is_success(),
                Err(_) => false,
            }
        }));
    }
    
    let mut all_success = true;
    for f in futures {
        if let Ok(success) = f.await {
            if !success { all_success = false; break; }
        } else {
            all_success = false; break;
        }
    }
    
    if all_success {
        db.insert_log("INFO", &format!("HTTP 測試: 節點 [{}] 測試通過", node_name));
    } else {
        db.insert_log("WARN", &format!("HTTP 測試: 節點 [{}] 測試失敗", node_name));
    }
    
    all_success
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
            if let Ok(mut current_status) = app.state::<AppState>().status.lock() {
                *current_status = status.clone();
            }

            if !status.api_connected {
                sleep(Duration::from_secs(3)).await;
                continue;
            }

            // 2. 開始測速演算法
            db.insert_log("INFO", "==================================================");
            db.insert_log("INFO", "[測速任務開始]");
            let current_tolerance = config.tolerance;
            db.insert_log("INFO", &format!("- 容忍度 (Tolerance): {} 分 (新節點需領先大於此分數才會切換)", current_tolerance));
            db.insert_log("INFO", &format!("- 併發數 (Max Concurrency): {}", config.max_concurrent));
            db.insert_log("INFO", &format!("- 測試次數 (Ping Count): {} 次", std::cmp::max(1, config.ping_count)));
            if config.enable_browser_test {
                db.insert_log("INFO", &format!("- HTTP 測試: 啟用 (群組: {}, 代理: {})", config.dedicated_test_group, config.clash_proxy_url));
            } else {
                db.insert_log("INFO", "- HTTP 測試: 未啟用");
            }
            db.insert_log("INFO", "--------------------------------------------------");
            status.is_testing = true;
            let _ = app.emit("status_update", &status);
            if let Ok(mut current_status) = app.state::<AppState>().status.lock() {
                *current_status = status.clone();
            }

            let mut all_group_results: Vec<GroupResult> = Vec::new();

            // 階段一：收集所有目標群組的節點資料與代理集對應
            use std::collections::{HashSet, HashMap};
            struct GroupInfo {
                group_name: String,
                nodes: Vec<String>,
                current_node: String,
            }
            let mut groups_info = Vec::new();
            let mut unique_nodes = HashSet::new();
            
            // 抓取所有的 Providers 來建立 Node -> Provider 映射表
            let mut node_to_provider: HashMap<String, String> = HashMap::new();
            if let Ok(providers) = clash.get_providers().await {
                for (prov_name, prov_detail) in providers {
                    if prov_name != "default" {
                        for proxy in prov_detail.proxies {
                            node_to_provider.insert(proxy.name, prov_name.clone());
                        }
                    }
                }
            }

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

            // 階段一.五：立刻發送保留上一輪數據的初始狀態給前端，避免畫面空白與資料閃爍
            let mut initial_group_results: Vec<GroupResult> = Vec::new();
            let last_results = {
                app.state::<AppState>().last_results.lock().unwrap().clone()
            };
            for info in &groups_info {
                let mut initial_nodes = Vec::new();
                for node in &info.nodes {
                    let mut old_delay = None;
                    let mut old_mean = None;
                    let mut old_jitter = None;
                    for lr in &last_results {
                        if lr.group_name == info.group_name {
                            if let Some(old_node) = lr.nodes.iter().find(|n| n.name == *node) {
                                old_delay = old_node.delay;
                                old_mean = old_node.mean;
                                old_jitter = old_node.jitter;
                            }
                        }
                    }

                    initial_nodes.push(NodeResult {
                        name: node.clone(),
                        delay: old_delay,
                        mean: old_mean,
                        jitter: old_jitter,
                        is_active: node == &info.current_node,
                        provider: node_to_provider.get(node).cloned(),
                    });
                }
                initial_group_results.push(GroupResult {
                    group_name: info.group_name.clone(),
                    nodes: initial_nodes,
                    is_locked: config.locked_groups.contains(&info.group_name),
                    selected_regions: config.group_regions.get(&info.group_name).cloned().unwrap_or_default(),
                });
            }
            let _ = app.emit("node_results", &initial_group_results);
            if let Ok(mut last) = app.state::<AppState>().last_results.lock() {
                *last = initial_group_results.clone();
            }

            // 階段二：對所有不重複的節點進行並發測速（每個節點內部進行多次測速）
            let test_url = config.test_urls.first().unwrap_or(&"http://www.gstatic.com/generate_204".to_string()).clone();
            let timeout = config.test_timeout;
            let max_concurrent = if config.max_concurrent > 0 { config.max_concurrent as usize } else { 10 };
            let ping_count = std::cmp::max(1, config.ping_count);

            let node_stream = futures::stream::iter(unique_nodes.into_iter());
            let clash_client = clash.clone();
            
            // Result<(Score, Mean, Jitter), String>
            let mut ping_results: HashMap<String, Result<(u64, u64, u64), String>> = HashMap::new();
            
            let mut tasks = node_stream.map(|node| {
                let client = clash_client.clone();
                let url = test_url.clone();
                async move {
                    let mut delays = Vec::new();
                    for _ in 0..ping_count {
                        match client.ping_proxy(&node, &url, timeout).await {
                            Ok(d) => delays.push(d),
                            Err(e) => {
                                // Option A: 只要有 1 次超時，就判定該節點「不穩定/Timeout」
                                return (node, Err(e));
                            }
                        }
                    }

                    if delays.is_empty() {
                        return (node, Err("Timeout".to_string()));
                    }

                    // 計算 Mean 和 Jitter (標準差)
                    let sum: u64 = delays.iter().sum();
                    let mean = sum / (delays.len() as u64);
                    
                    let mut variance_sum = 0.0;
                    for &d in &delays {
                        let diff = (d as f64) - (mean as f64);
                        variance_sum += diff * diff;
                    }
                    let variance = variance_sum / (delays.len() as f64);
                    let jitter = variance.sqrt().round() as u64;
                    let score = mean + jitter;

                    (node, Ok((score, mean, jitter)))
                }
            }).buffer_unordered(max_concurrent);

            while let Some((node, res)) = tasks.next().await {
                ping_results.insert(node, res);
            }

            let mut http_test_cache: HashMap<String, bool> = HashMap::new();

            // 階段三：分配測速結果至各群組進行決策
            for info in groups_info {
                let group = &info.group_name;
                let current_node = &info.current_node;
                
                let selected_regions = config.group_regions.get(group).cloned().unwrap_or_default();
                let has_region_filter = !selected_regions.is_empty();

                let mut fastest_node = None;
                let mut min_delay = u64::MAX;
                let mut current_node_delay = u64::MAX;

                let mut final_results: Vec<NodeResult> = Vec::new();

                for node in &info.nodes {
                    let (delay_opt, mean_opt, jitter_opt) = match ping_results.get(node) {
                        Some(Ok((s, m, j))) => (Some(*s), Some(*m), Some(*j)),
                        _ => (None, None, None),
                    };

                    let is_region_matched = if has_region_filter {
                        selected_regions.iter().any(|r| matches_region(node, r))
                    } else {
                        true
                    };

                    if let Some(delay) = delay_opt {
                        if node == current_node {
                            current_node_delay = delay;
                            if has_region_filter && !is_region_matched {
                                current_node_delay = u64::MAX; // 若當前節點不符合篩選地區，強制視為失效以觸發切換
                            }
                        }
                        if is_region_matched && delay < min_delay {
                            min_delay = delay;
                            fastest_node = Some(node.clone());
                        }
                    }

                    final_results.push(NodeResult {
                        name: node.clone(),
                        delay: delay_opt,
                        mean: mean_opt,
                        jitter: jitter_opt,
                        is_active: node == current_node,
                        provider: node_to_provider.get(node).cloned(),
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

                let get_display_name = |n: &str| -> String {
                    node_to_provider.get(n).map(|p| format!("[{}] {}", p, n)).unwrap_or_else(|| n.to_string())
                };

                let mut debug_msgs = Vec::new();
                if has_region_filter {
                    debug_msgs.push(format!("  (目前套用地區篩選: [{}])", selected_regions.join(", ")));
                }

                let top_n = std::cmp::min(10, final_results.len());
                for res in final_results.iter().take(top_n) {
                    let display_name = get_display_name(&res.name);
                    let mut filter_tag = "".to_string();
                    
                    if has_region_filter {
                        let is_match = selected_regions.iter().any(|r| matches_region(&res.name, r));
                        if !is_match {
                            filter_tag = " (略過: 區域不符)".to_string();
                        }
                    }

                    if let (Some(s), Some(m), Some(j)) = (res.delay, res.mean, res.jitter) {
                        debug_msgs.push(format!("  - {}: Score {} (Mean: {}ms, Jitter: {}ms){}", display_name, s, m, j, filter_tag));
                    } else {
                        debug_msgs.push(format!("  - {}: Timeout{}", display_name, filter_tag));
                    }
                }
                
                if final_results.len() > 10 {
                    debug_msgs.push(format!("  - ... 以及其他 {} 個節點 (省略顯示)", final_results.len() - 10));
                }
                
                db.insert_log("DEBUG", &format!("[群組: {}] 測速結果 (Top {}):", group, std::cmp::min(10, final_results.len())));
                for msg in debug_msgs {
                    db.insert_log("DEBUG", &msg);
                }

                let current_delay_str = if current_node_delay == u64::MAX { "Timeout".to_string() } else { format!("{} 分", current_node_delay) };
                let current_node_str = get_display_name(current_node);

                db.insert_log("DEBUG", &format!("[群組: {}] 決策分析開始...", group));
                db.insert_log("DEBUG", &format!("- 當前使用節點: {} (分數: {})", current_node_str, current_delay_str));
                
                // Jitter 容忍度邏輯判斷
                let is_locked = config.locked_groups.contains(group);
                
                if is_locked {
                    db.insert_log("INFO", &format!("[群組: {}] 狀態為「已鎖定 (手動切換)」，略過自動切換邏輯。", group));
                } else {
                    let mut switched = false;
                    let mut switched_to_node = None;
                    for res in &final_results {
                        if let Some(delay) = res.delay {
                            // 檢查是否符合地區篩選
                            if has_region_filter && !selected_regions.iter().any(|r| matches_region(&res.name, r)) {
                                db.insert_log("DEBUG", &format!("- 候選節點 {} (分數: {}): 區域不符 (套用地區: {:?})，略過。", get_display_name(&res.name), delay, selected_regions));
                                continue;
                            }

                            let diff = if current_node_delay == u64::MAX {
                                u64::MAX
                            } else {
                                current_node_delay.saturating_sub(delay)
                            };

                            if diff > config.tolerance {
                                let node_name = &res.name;
                                db.insert_log("DEBUG", &format!("- 候選節點 {} (分數: {}): 比當前節點快 {} 分，大於容忍度 ({} 分)，進入切換驗證階段...", get_display_name(node_name), delay, diff, config.tolerance));

                                let mut is_good = true;

                                // 如果啟用 HTTP 測試防護
                                if config.enable_browser_test && !config.dedicated_test_group.is_empty() && !config.clash_proxy_url.is_empty() && !config.browser_test_urls.is_empty() {
                                    is_good = if let Some(&cached_res) = http_test_cache.get(node_name) {
                                        if cached_res {
                                            db.insert_log("DEBUG", "  -> 沿用跨群組快取: HTTP 測試成功。");
                                        } else {
                                            db.insert_log("DEBUG", "  -> 沿用跨群組快取: HTTP 測試失敗。");
                                        }
                                        cached_res
                                    } else {
                                        db.insert_log("DEBUG", "  -> 開始發送真實 HTTP 請求進行連通性驗證...");
                                        let result = perform_http_test(node_name, &config, &clash, &db).await;
                                        http_test_cache.insert(node_name.clone(), result);
                                        result
                                    };
                                }

                                if is_good {
                                    // 執行切換
                                    if let Ok(_) = clash.select_proxy(group, node_name).await {
                                        let advantage_str = if diff == u64::MAX { "當前節點超時或區域不符".to_string() } else { format!("領先 {} 分", diff) };
                                        db.insert_log("INFO", &format!("[群組: {}] 成功切換至: {} ({})", group, get_display_name(node_name), advantage_str));
                                        
                                        switched_to_node = Some(node_name.to_string());
                                        switched = true;
                                        break; // 成功切換，結束尋找
                                    } else {
                                        db.insert_log("ERROR", &format!("[群組: {}] 嘗試切換至 {} 失敗 (API 錯誤)。", group, get_display_name(node_name)));
                                        break;
                                    }
                                } else {
                                    db.insert_log("DEBUG", &format!("  -> 節點 {} 未通過 HTTP 測試，放棄並繼續尋找下一順位。", get_display_name(node_name)));
                                    continue; // 繼續看下一個節點
                                }
                            } else {
                                // 第一個符合地區且能連上的節點，如果連它都沒有超過 tolerance，後面的更慢就不用看了
                                db.insert_log("INFO", &format!("[群組: {}] 候選節點 {} (分數: {}): 與當前節點差距 {} 分，未超過容忍度 ({} 分)。保持當前節點不切換。", group, get_display_name(&res.name), delay, diff, config.tolerance));
                                switched = true; // 視為已處理
                                break;
                            }
                        } else {
                            // delay 是 None 代表 Timeout，後面都是 Timeout 了
                            db.insert_log("DEBUG", "- 剩餘候選節點皆為 Timeout，結束尋找。");
                            break;
                        }
                    }

                    if let Some(target_node_name) = switched_to_node {
                        for r in &mut final_results {
                            r.is_active = r.name == target_node_name;
                        }
                    }

                    if !switched {
                        // 如果跑到這裡且 current_node_delay 是 MAX，代表全部都死了或 HTTP 測試全滅
                        if current_node_delay == u64::MAX {
                            db.insert_log("WARN", &format!("[群組: {}] 警告：所有候補節點皆無法使用或未通過 HTTP 測試！", group));
                        }
                    }
                }

                all_group_results.push(GroupResult {
                    group_name: group.clone(),
                    nodes: final_results,
                    is_locked,
                    selected_regions,
                });
            }

            let _ = app.emit("node_results", &all_group_results);
            if let Ok(mut last) = app.state::<AppState>().last_results.lock() {
                *last = all_group_results.clone();
            }

            status.is_testing = false;
            let _ = app.emit("status_update", &status);
            if let Ok(mut current_status) = app.state::<AppState>().status.lock() {
                *current_status = status.clone();
            }

            // 3. 進入休眠倒數
            let mut remaining = config.check_interval;
            while remaining > 0 {
                status.next_check_in = remaining;
                let _ = app.emit("status_update", &status);
                if let Ok(mut current_status) = app.state::<AppState>().status.lock() {
                    *current_status = status.clone();
                }
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
