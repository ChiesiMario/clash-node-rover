use rusqlite::{Connection, Result};
use std::sync::Mutex;
use tauri::{AppHandle, Manager, Emitter};

#[derive(serde::Serialize, Clone)]
pub struct LogEntry {
    pub id: i64,
    pub timestamp: String,
    pub level: String,
    pub message: String,
}

pub struct Db {
    pub conn: Mutex<Connection>,
    pub app_handle: AppHandle,
}

impl Db {
    pub fn new(app: &AppHandle) -> Result<Self> {
        let path = app.path().app_data_dir().unwrap().join("rover.db");
        if let Some(dir) = path.parent() {
            let _ = std::fs::create_dir_all(dir);
        }
        let conn = Connection::open(&path)?;

        // 初始化資料表
        conn.execute(
            "CREATE TABLE IF NOT EXISTS logs (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                level TEXT NOT NULL,
                message TEXT NOT NULL
            )",
            [],
        )?;

        conn.execute(
            "CREATE TABLE IF NOT EXISTS node_history (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                group_name TEXT NOT NULL,
                node_name TEXT NOT NULL,
                ping_ms INTEGER NOT NULL,
                status TEXT NOT NULL
            )",
            [],
        )?;

        Ok(Self {
            conn: Mutex::new(conn),
            app_handle: app.clone(),
        })
    }

    pub fn insert_log(&self, level: &str, message: &str) {
        if let Ok(conn) = self.conn.lock() {
            let _ = conn.execute(
                "INSERT INTO logs (level, message) VALUES (?1, ?2)",
                rusqlite::params![level, message],
            );

            if let Ok(mut stmt) = conn.prepare("SELECT id, datetime(timestamp, 'localtime'), level, message FROM logs ORDER BY id DESC LIMIT 1") {
                if let Ok(mut rows) = stmt.query([]) {
                    if let Some(row) = rows.next().unwrap_or(None) {
                        let log = LogEntry {
                            id: row.get(0).unwrap_or(0),
                            timestamp: row.get(1).unwrap_or_default(),
                            level: row.get(2).unwrap_or_default(),
                            message: row.get(3).unwrap_or_default(),
                        };
                        let _ = self.app_handle.emit("new_log", &log);
                    }
                }
            }
        }
    }

    pub fn get_logs(&self, limit: u32) -> Vec<LogEntry> {
        let mut logs = Vec::new();
        if let Ok(conn) = self.conn.lock() {
            if let Ok(mut stmt) = conn.prepare("SELECT id, datetime(timestamp, 'localtime'), level, message FROM (SELECT * FROM logs ORDER BY id DESC LIMIT ?1) ORDER BY id ASC") {
                if let Ok(mut rows) = stmt.query([limit]) {
                    while let Some(row) = rows.next().unwrap_or(None) {
                        logs.push(LogEntry {
                            id: row.get(0).unwrap_or(0),
                            timestamp: row.get(1).unwrap_or_default(),
                            level: row.get(2).unwrap_or_default(),
                            message: row.get(3).unwrap_or_default(),
                        });
                    }
                }
            }
        }
        logs
    }
}
