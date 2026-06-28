use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::fs;
use tauri::{AppHandle, Manager};

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct Config {
    pub api_url: String,
    pub api_secret: String,
    pub check_interval: u64,
    pub target_groups: Vec<String>,
    pub dedicated_test_group: String,
    pub test_urls: Vec<String>,
    pub test_timeout: u64,
    pub tolerance: u64,
    pub cleanup_days: u64,
    pub max_concurrent: u64,
    pub clash_proxy_url: String,
    pub max_backoff_cycles: u64,
    pub enable_browser_test: bool,
    pub browser_test_urls: Vec<String>,
    pub locked_groups: Vec<String>,
    pub ping_count: u32,
    pub group_regions: std::collections::HashMap<String, Vec<String>>,
    pub has_completed_setup: bool,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            api_url: "http://127.0.0.1:9090".to_string(),
            api_secret: "".to_string(),
            check_interval: 60,
            target_groups: vec![],
            dedicated_test_group: "".into(),
            test_urls: vec![
                "http://www.gstatic.com/generate_204".into(),
                "http://cp.cloudflare.com/generate_204".into(),
            ],
            test_timeout: 2000,
            tolerance: 3,
            cleanup_days: 7,
            max_concurrent: 10,
            clash_proxy_url: "http://127.0.0.1:7890".into(),
            max_backoff_cycles: 5,
            enable_browser_test: true,
            browser_test_urls: vec![
                "https://www.google.com".into(),
                "https://www.youtube.com".into(),
            ],
            locked_groups: vec![],
            ping_count: 3,
            group_regions: std::collections::HashMap::new(),
            has_completed_setup: false,
        }
    }
}

pub fn get_config_path(app: &AppHandle) -> PathBuf {
    app.path().app_config_dir().unwrap().join("config.json")
}

pub fn load_config(app: &AppHandle) -> Config {
    let path = get_config_path(app);
    if let Ok(data) = fs::read_to_string(&path) {
        serde_json::from_str(&data).unwrap_or_default()
    } else {
        Config::default()
    }
}

pub fn save_config(app: &AppHandle, config: &Config) -> Result<(), String> {
    let path = get_config_path(app);
    if let Some(dir) = path.parent() {
        let _ = fs::create_dir_all(dir);
    }
    let data = serde_json::to_string_pretty(config).map_err(|e| e.to_string())?;
    fs::write(&path, data).map_err(|e| e.to_string())?;
    Ok(())
}
