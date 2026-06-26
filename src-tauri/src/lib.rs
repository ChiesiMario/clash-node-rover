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
fn get_logs(state: tauri::State<db::Db>) -> Vec<db::LogEntry> {
    state.get_logs(100)
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
            get_logs
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
