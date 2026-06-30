mod config;
mod db;
mod clash;
mod watchdog;

use tauri::{
    AppHandle, Manager, Emitter,
    menu::{Menu, MenuItem, CheckMenuItem, PredefinedMenuItem},
    tray::TrayIconBuilder,
};
use config::Config;
use std::sync::{Arc, Mutex};
use tokio::sync::Notify;
use tauri_plugin_autostart::{MacosLauncher, ManagerExt};
use std::env;

struct AppState {
    pub config: Mutex<Config>,
    pub force_test: Arc<Notify>,
    pub last_results: Mutex<Vec<watchdog::GroupResult>>,
    pub status: Mutex<watchdog::AppStatus>,
}

#[tauri::command]
fn get_config(state: tauri::State<AppState>) -> Config {
    state.config.lock().unwrap().clone()
}

#[tauri::command]
fn save_config(app: AppHandle, state: tauri::State<AppState>, new_config: Config) -> Result<(), String> {
    *state.config.lock().unwrap() = new_config.clone();
    let res = config::save_config(&app, &new_config);
    if res.is_ok() {
        let _ = app.emit("config_updated", ());
    }
    res
}

#[tauri::command]
async fn verify_clash_api(api_url: String, api_secret: String) -> Result<bool, String> {
    let api = clash::ClashApi::new(&api_url, &api_secret);
    api.verify_connection().await.map(|_| true).map_err(|e| e)
}

#[tauri::command]
async fn get_clash_selectors(state: tauri::State<'_, AppState>) -> Result<Vec<String>, String> {
    let config = state.config.lock().unwrap().clone();
    let api = clash::ClashApi::new(&config.api_url, &config.api_secret);
    api.get_selectors().await
}

#[tauri::command]
fn force_test(state: tauri::State<AppState>) {
    state.force_test.notify_one();
}

#[tauri::command]
fn get_latest_results(state: tauri::State<AppState>) -> Vec<watchdog::GroupResult> {
    state.last_results.lock().unwrap().clone()
}

#[tauri::command]
fn get_logs(state: tauri::State<db::Db>) -> Vec<db::LogEntry> {
    state.get_logs(100)
}

#[tauri::command]
fn get_node_history(state: tauri::State<db::Db>, node_name: String, hours: u32) -> Vec<db::NodeHistoryEntry> {
    state.get_node_history(&node_name, hours)
}

#[tauri::command]
fn get_status(state: tauri::State<AppState>) -> watchdog::AppStatus {
    state.status.lock().unwrap().clone()
}

#[tauri::command]
fn toggle_group_lock(app: tauri::AppHandle, state: tauri::State<AppState>, group: String, locked: bool) -> Result<(), String> {
    let mut config = state.config.lock().unwrap().clone();
    if locked {
        if !config.locked_groups.contains(&group) {
            config.locked_groups.push(group.clone());
        }
    } else {
        config.locked_groups.retain(|g| g != &group);
        config.manual_nodes.remove(&group);
    }
    *state.config.lock().unwrap() = config.clone();
    config::save_config(&app, &config)
}

#[tauri::command]
fn toggle_group_region(app: tauri::AppHandle, state: tauri::State<AppState>, group: String, region: String) -> Result<(), String> {
    let mut config = state.config.lock().unwrap().clone();
    let regions = config.group_regions.entry(group.clone()).or_insert_with(Vec::new);
    if regions.contains(&region) {
        regions.retain(|r| r != &region);
    } else {
        regions.push(region);
    }
    *state.config.lock().unwrap() = config.clone();
    config::save_config(&app, &config)
}

#[tauri::command]
async fn manual_switch(app: tauri::AppHandle, state: tauri::State<'_, AppState>, db: tauri::State<'_, db::Db>, group: String, node: String) -> Result<(), String> {
    let mut config = state.config.lock().unwrap().clone();
    let api = clash::ClashApi::new(&config.api_url, &config.api_secret);
    api.select_proxy(&group, &node).await?;
    db.insert_log("INFO", &format!("手動切換群組 [{}] 至節點: {}", group, node));
    
    config.manual_nodes.insert(group.clone(), node.clone());
    
    if !config.locked_groups.contains(&group) {
        config.locked_groups.push(group.clone());
    }
    
    *state.config.lock().unwrap() = config.clone();
    let _ = config::save_config(&app, &config);
    
    if let Ok(mut last_results) = state.last_results.lock() {
        for g in last_results.iter_mut() {
            if g.group_name == group {
                for n in &mut g.nodes {
                    n.is_active = n.name == node;
                }
            }
        }
    }
    
    Ok(())
}

#[tauri::command]
fn toggle_pause(app: tauri::AppHandle, state: tauri::State<AppState>) {
    let mut status = state.status.lock().unwrap();
    status.is_paused = !status.is_paused;
    let _ = app.emit("status_update", &*status);
    state.force_test.notify_one(); // Wake up the watchdog if it's sleeping so it can pause immediately or resume immediately
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_autostart::init(MacosLauncher::LaunchAgent, Some(vec!["--autostart"])))
        .setup(|app| {
            let args: Vec<String> = env::args().collect();
            let is_autostart = args.contains(&"--autostart".to_string());
            
            let hi_res_icon = tauri::image::Image::from_bytes(include_bytes!("../icons/128x128.png")).unwrap_or_else(|_| app.default_window_icon().unwrap().clone());
            
            if !is_autostart {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.set_icon(hi_res_icon.clone());
                    let _ = window.show();
                    let _ = window.set_focus();
                }
            } else {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.set_icon(hi_res_icon.clone());
                }
            }
            
            let cfg = config::load_config(app.handle());
            
            // 建立 db (若出錯直接印出或由上層處理)
            if let Ok(db) = db::Db::new(app.handle()) {
                app.manage(db);
            }

            app.manage(AppState {
                config: Mutex::new(cfg.clone()),
                force_test: Arc::new(Notify::new()),
                last_results: Mutex::new(Vec::new()),
                status: Mutex::new(watchdog::AppStatus {
                    api_connected: false,
                    is_testing: false,
                    next_check_in: 0,
                    is_paused: false,
                }),
            });
            
            // 啟動背景測速守門員
            watchdog::start_watchdog(app.handle().clone());
            
            // 建立系統列選單與圖示
            let is_autostart_enabled = app.autolaunch().is_enabled().unwrap_or(false);
            
            let show_text = match cfg.language.as_str() {
                "zh-TW" => "顯示儀表板",
                "zh-CN" => "显示仪表板",
                _ => "Show Dashboard",
            };
            let force_test_text = match cfg.language.as_str() {
                "zh-TW" => "強制測速",
                "zh-CN" => "强制测速",
                _ => "Force Test",
            };
            let toggle_pause_text = match cfg.language.as_str() {
                "zh-TW" => "暫停 / 恢復",
                "zh-CN" => "暂停 / 恢复",
                _ => "Pause / Resume",
            };
            let autostart_text = match cfg.language.as_str() {
                "zh-TW" => "開機自動啟動",
                "zh-CN" => "开机自动启动",
                _ => "Auto-start on Boot",
            };
            let quit_text = match cfg.language.as_str() {
                "zh-TW" => "退出",
                "zh-CN" => "退出",
                _ => "Quit",
            };

            let show_i = MenuItem::with_id(app, "show", show_text, true, None::<&str>)?;
            let force_test_i = MenuItem::with_id(app, "force_test", force_test_text, true, None::<&str>)?;
            let toggle_pause_i = MenuItem::with_id(app, "toggle_pause", toggle_pause_text, true, None::<&str>)?;
            let autostart_i = CheckMenuItem::with_id(app, "toggle_autostart", autostart_text, true, is_autostart_enabled, None::<&str>)?;
            let separator = PredefinedMenuItem::separator(app)?;
            let quit_i = MenuItem::with_id(app, "quit", quit_text, true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &force_test_i, &toggle_pause_i, &separator, &autostart_i, &separator, &quit_i])?;

            let tray_icon = tauri::image::Image::from_bytes(include_bytes!("../icons/128x128.png")).unwrap_or_else(|_| app.default_window_icon().unwrap().clone());
            let _tray = TrayIconBuilder::new()
                .icon(tray_icon)
                .menu(&menu)
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "quit" => {
                        std::process::exit(0);
                    }
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "force_test" => {
                        let state = app.state::<AppState>();
                        state.force_test.notify_one();
                    }
                    "toggle_pause" => {
                        let state = app.state::<AppState>();
                        let mut status = state.status.lock().unwrap();
                        status.is_paused = !status.is_paused;
                        let _ = app.emit("status_update", &*status);
                        state.force_test.notify_one();
                    }
                    "toggle_autostart" => {
                        let manager = app.autolaunch();
                        let is_enabled = manager.is_enabled().unwrap_or(false);
                        if is_enabled {
                            let _ = manager.disable();
                        } else {
                            let _ = manager.enable();
                        }
                    }
                    _ => {}
                })
                .on_tray_icon_event(|tray, event| {
                    if let tauri::tray::TrayIconEvent::Click {
                        button: tauri::tray::MouseButton::Left,
                        button_state: tauri::tray::MouseButtonState::Up,
                        ..
                    } = event
                    {
                        if let Some(window) = tray.app_handle().get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                })
                .build(app)?;
            
            Ok(())
        })
        .on_window_event(|window, event| match event {
            tauri::WindowEvent::CloseRequested { api, .. } => {
                let _ = window.hide();
                api.prevent_close();
            }
            _ => {}
        })
        .invoke_handler(tauri::generate_handler![
            get_config,
            save_config,
            verify_clash_api,
            get_clash_selectors,
            force_test,
            get_logs,
            toggle_group_lock,
            toggle_group_region,
            manual_switch,
            get_latest_results,
            get_status,
            toggle_pause,
            get_node_history
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
