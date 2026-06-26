use rusqlite::{Connection, Result};
use std::sync::Mutex;
use tauri::{AppHandle, Manager};

pub struct Db {
    pub conn: Mutex<Connection>,
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
        })
    }

    pub fn insert_log(&self, level: &str, message: &str) {
        if let Ok(conn) = self.conn.lock() {
            let _ = conn.execute(
                "INSERT INTO logs (level, message) VALUES (?1, ?2)",
                rusqlite::params![level, message],
            );
        }
    }
}
