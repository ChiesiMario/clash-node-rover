#![allow(dead_code)]
use reqwest::{Client, header};
use serde::Deserialize;
use std::time::Duration;
use urlencoding::encode;

#[derive(Clone)]
pub struct ClashApi {
    client: Client,
    base_url: String,
    secret: String,
}

#[derive(Deserialize, Debug)]
pub struct ProxiesResponse {
    pub proxies: std::collections::HashMap<String, ProxyDetail>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ProxyDetail {
    pub name: String,
    #[serde(rename = "type")]
    pub proxy_type: String,
    pub all: Option<Vec<String>>,
    pub now: Option<String>,
}

#[derive(Deserialize, Debug)]
pub struct ProvidersResponse {
    pub providers: std::collections::HashMap<String, ProviderDetail>,
}

#[derive(Deserialize, Debug)]
pub struct ProviderDetail {
    pub name: String,
    pub proxies: Vec<ProxyDetail>,
    #[serde(rename = "vehicleType")]
    pub vehicle_type: String,
}

#[derive(Deserialize, Debug)]
pub struct DelayResponse {
    pub delay: u64,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ConnectionMetadata {
    pub host: String,
    #[serde(rename = "destinationIP")]
    pub destination_ip: String,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ConnectionInfo {
    pub id: String,
    pub metadata: ConnectionMetadata,
    pub rule: String,
    #[serde(rename = "rulePayload")]
    pub rule_payload: String,
    pub chains: Vec<String>,
}

#[derive(Deserialize, Debug, Clone)]
pub struct ConnectionsResponse {
    pub connections: Vec<ConnectionInfo>,
}

impl ClashApi {
    pub fn new(base_url: &str, secret: &str) -> Self {
        let mut headers = header::HeaderMap::new();
        if !secret.is_empty() {
            if let Ok(h) = header::HeaderValue::from_str(&format!("Bearer {}", secret)) {
                headers.insert(header::AUTHORIZATION, h);
            }
        }
        let client = Client::builder()
            .default_headers(headers)
            .timeout(Duration::from_secs(10))
            .build()
            .unwrap_or_default();

        Self {
            client,
            base_url: base_url.trim_end_matches('/').to_string(),
            secret: secret.to_string(),
        }
    }

    pub async fn verify_connection(&self) -> Result<(), String> {
        let url = format!("{}/version", self.base_url);
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        if resp.status().is_success() {
            Ok(())
        } else {
            Err(format!("API returned status: {}", resp.status()))
        }
    }

    pub async fn get_selectors(&self) -> Result<Vec<String>, String> {
        let url = format!("{}/proxies", self.base_url);
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        let data: ProxiesResponse = resp.json().await.map_err(|e| e.to_string())?;
        let mut selectors = Vec::new();
        for (name, detail) in data.proxies {
            if detail.proxy_type == "Selector" {
                selectors.push(name);
            }
        }
        Ok(selectors)
    }

    pub async fn get_proxy_group(&self, name: &str) -> Result<ProxyDetail, String> {
        let url = format!("{}/proxies/{}", self.base_url, encode(name));
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        if !resp.status().is_success() {
            return Err(format!("Group {} not found", name));
        }
        let detail: ProxyDetail = resp.json().await.map_err(|e| e.to_string())?;
        Ok(detail)
    }

    pub async fn get_providers(&self) -> Result<std::collections::HashMap<String, ProviderDetail>, String> {
        let url = format!("{}/providers/proxies", self.base_url);
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        if !resp.status().is_success() {
            return Err(format!("Providers API returned status: {}", resp.status()));
        }
        let data: ProvidersResponse = resp.json().await.map_err(|e| e.to_string())?;
        Ok(data.providers)
    }

    pub async fn select_proxy(&self, group_name: &str, proxy_name: &str) -> Result<(), String> {
        let url = format!("{}/proxies/{}", self.base_url, encode(group_name));
        let body = serde_json::json!({ "name": proxy_name });
        let resp = self.client.put(&url).json(&body).send().await.map_err(|e| e.to_string())?;
        if resp.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to select proxy: {}", resp.status()))
        }
    }

    pub async fn ping_proxy(&self, proxy_name: &str, test_url: &str, timeout_ms: u64) -> Result<u64, String> {
        let url = format!(
            "{}/proxies/{}/delay?url={}&timeout={}",
            self.base_url,
            encode(proxy_name),
            encode(test_url),
            timeout_ms
        );
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        if resp.status().is_success() {
            let data: DelayResponse = resp.json().await.map_err(|e| e.to_string())?;
            Ok(data.delay)
        } else {
            Err(format!("Ping failed with status: {}", resp.status()))
        }
    }

    pub async fn get_connections(&self) -> Result<ConnectionsResponse, String> {
        let url = format!("{}/connections", self.base_url);
        let resp = self.client.get(&url).send().await.map_err(|e| e.to_string())?;
        if resp.status().is_success() {
            let data: ConnectionsResponse = resp.json().await.map_err(|e| e.to_string())?;
            Ok(data)
        } else {
            Err(format!("Failed to get connections: {}", resp.status()))
        }
    }
}
