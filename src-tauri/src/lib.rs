mod config;
mod db;
mod clash;
mod watchdog;

use tauri::{
    AppHandle, Manager,
    menu::{Menu, MenuItem},
    tray::TrayIconBuilder,
};
use config::Config;
use std::sync::{Arc, Mutex};
use tokio::sync::Notify;

struct AppState {
    pub config: Mutex<Config>,
    pub force_test: Arc<Notify>,
    pub last_results: Mutex<Vec<watchdog::GroupResult>>,
}

#[tauri::command]
fn get_config(state: tauri::State<AppState>) -> Config {
    state.config.lock().unwrap().clone()
}

#[tauri::command]
fn save_config(app: AppHandle, state: tauri::State<AppState>, new_config: Config) -> Result<(), String> {
    *state.config.lock().unwrap() = new_config.clone();
    config::save_config(&app, &new_config)
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
fn toggle_group_lock(app: tauri::AppHandle, state: tauri::State<AppState>, group: String, locked: bool) -> Result<(), String> {
    let mut config = state.config.lock().unwrap().clone();
    if locked {
        if !config.locked_groups.contains(&group) {
            config.locked_groups.push(group);
        }
    } else {
        config.locked_groups.retain(|g| g != &group);
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
    
    if !config.locked_groups.contains(&group) {
        config.locked_groups.push(group.clone());
        *state.config.lock().unwrap() = config.clone();
        let _ = config::save_config(&app, &config);
    }
    
    state.force_test.notify_one();
    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            let cfg = config::load_config(app.handle());
            
            // 建立 db (若出錯直接印出或由上層處理)
            if let Ok(db) = db::Db::new(app.handle()) {
                app.manage(db);
            }

            app.manage(AppState {
                config: Mutex::new(cfg),
                force_test: Arc::new(Notify::new()),
                last_results: Mutex::new(Vec::new()),
            });
            
            // 啟動背景測速守門員
            watchdog::start_watchdog(app.handle().clone());
            
            // 建立系統列選單與圖示
            let show_i = MenuItem::with_id(app, "show", "開啟儀表板", true, None::<&str>)?;
            let quit_i = MenuItem::with_id(app, "quit", "結束程式", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &quit_i])?;

            let _tray = TrayIconBuilder::new()
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
        .invoke_handler(tauri::generate_handler![
            get_config,
            save_config,
            get_clash_selectors,
            force_test,
            get_logs,
            toggle_group_lock,
            manual_switch,
            get_latest_results
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
