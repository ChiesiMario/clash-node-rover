use crate::{AppState, clash::ClashApi, config::Config};
use std::time::Duration;
use tokio::time::sleep;

#[derive(serde::Serialize)]
pub struct ProbeResult {
    pub domain: String,
    pub rule: String,
    pub rule_payload: String,
    pub proxy_chain: Vec<String>,
}

#[tauri::command]
pub async fn probe_rule(state: tauri::State<'_, AppState>, domain: String) -> Result<ProbeResult, String> {
    let config = state.config.lock().unwrap().clone();
    
    if config.api_url.is_empty() {
        return Err("API URL is empty".to_string());
    }

    let clash = ClashApi::new(&config.api_url, &config.api_secret);
    
    // We send a request through the proxy to trigger the rule evaluation.
    // The request doesn't even need to succeed, it just needs to hit Clash.
    let proxy_url_opt = if config.probe_use_proxy {
        if config.probe_proxy_url.starts_with("http") {
            Some(config.probe_proxy_url.clone())
        } else if !config.probe_proxy_url.is_empty() {
            Some(format!("http://{}", config.probe_proxy_url))
        } else {
            None
        }
    } else {
        None
    };

    let target_url = format!("http://{}/?probe=rover", domain);
    
    let client = if let Some(p_url) = proxy_url_opt {
        match reqwest::Proxy::all(&p_url) {
            Ok(p) => reqwest::Client::builder()
                .proxy(p)
                .timeout(Duration::from_millis(1500))
                .build()
                .unwrap_or_default(),
            Err(_) => reqwest::Client::new(),
        }
    } else {
        reqwest::Client::builder()
            .timeout(Duration::from_millis(1500))
            .build()
            .unwrap_or_default()
    };

    // Send the request in the background
    let _ = tauri::async_runtime::spawn(async move {
        let _ = client.get(&target_url).send().await;
    });

    // Poll for the connection for up to 2 seconds (40 iterations * 50ms)
    for _ in 0..40 {
        sleep(Duration::from_millis(50)).await;
        
        if let Ok(connections) = clash.get_connections().await {
            for conn in connections.connections {
                if conn.metadata.host == domain {
                    // Found it!
                    return Ok(ProbeResult {
                        domain,
                        rule: conn.rule,
                        rule_payload: conn.rule_payload,
                        proxy_chain: conn.chains,
                    });
                }
            }
        }
    }

    Err("No matching connection found in time. The rule might be REJECT or the proxy is not used.".to_string())
}
