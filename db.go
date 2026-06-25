package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sqlDB *sql.DB
}

func InitDB() (*DB, error) {
	db, err := sql.Open("sqlite", "rover.db")
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS ping_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_name TEXT,
		timestamp INTEGER,
		delay INTEGER,
		success BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS browser_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_name TEXT,
		timestamp INTEGER,
		url TEXT,
		success BOOLEAN,
		load_time_ms INTEGER
	);
	CREATE TABLE IF NOT EXISTS metadata (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_node_time ON ping_logs(node_name, timestamp);
	`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	// 開啟 WAL 模式與調整同步層級以優化高併發寫入
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")
	db.Exec("PRAGMA busy_timeout=5000;")
	db.SetMaxOpenConns(1) // 強制 Go 單線程操作資料庫，從根源消滅 SQLITE_BUSY

	// 自動升級資料庫結構 (如果舊版沒有 downloaded_bytes 欄位)
	db.Exec("ALTER TABLE bandwidth_logs ADD COLUMN downloaded_bytes INTEGER DEFAULT 0;")

	return &DB{sqlDB: db}, nil
}

func (d *DB) InsertLog(nodeName string, delay int, success bool) error {
	query := `INSERT INTO ping_logs (node_name, timestamp, delay, success) VALUES (?, ?, ?, ?)`
	_, err := d.sqlDB.Exec(query, nodeName, time.Now().Unix(), delay, success)
	return err
}


func (d *DB) InsertBrowserLog(nodeName string, url string, success bool, loadTimeMs int) error {
	query := `INSERT INTO browser_logs (node_name, timestamp, url, success, load_time_ms) VALUES (?, ?, ?, ?, ?)`
	_, err := d.sqlDB.Exec(query, nodeName, time.Now().Unix(), url, success, loadTimeMs)
	return err
}

func (d *DB) ClearBrowserLogs() error {
	_, err := d.sqlDB.Exec(`DELETE FROM browser_logs`)
	return err
}

func (d *DB) GetMetadata(key string) (string, error) {
	var value string
	err := d.sqlDB.QueryRow(`SELECT value FROM metadata WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *DB) SetMetadata(key string, value string) error {
	_, err := d.sqlDB.Exec(`INSERT INTO metadata (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?`, key, value, value)
	return err
}

func (d *DB) CleanOldLogs(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()

	query1 := `DELETE FROM ping_logs WHERE timestamp < ?`
	_, err := d.sqlDB.Exec(query1, cutoff)
	if err != nil {
		return err
	}

	query3 := `DELETE FROM browser_logs WHERE timestamp < ?`
	_, err = d.sqlDB.Exec(query3, cutoff)
	return err
}





type PingLog struct {
	Timestamp int64
	Delay     int
}

func (d *DB) GetNodeHistory(nodeName string, hours int) ([]PingLog, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	query := `
		SELECT timestamp, CASE WHEN success = 1 THEN delay ELSE 0 END as delay
		FROM ping_logs 
		WHERE node_name = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`

	rows, err := d.sqlDB.Query(query, nodeName, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []PingLog
	for rows.Next() {
		var log PingLog
		if err := rows.Scan(&log.Timestamp, &log.Delay); err == nil {
			history = append(history, log)
		}
	}

	return history, nil
}

type BrowserLog struct {
	Timestamp  int64
	LoadTimeMs int
}

func (d *DB) GetBrowserHistory(nodeName string, hours int) ([]BrowserLog, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	query := `
		SELECT timestamp, load_time_ms 
		FROM browser_logs 
		WHERE node_name = ? AND timestamp >= ?
		ORDER BY timestamp ASC
	`

	rows, err := d.sqlDB.Query(query, nodeName, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []BrowserLog
	for rows.Next() {
		var log BrowserLog
		if err := rows.Scan(&log.Timestamp, &log.LoadTimeMs); err == nil {
			history = append(history, log)
		}
	}

	return history, nil
}

func (d *DB) GetLastBrowserSuccessTime(nodeName string, url string) (time.Time, error) {
	var timestamp int64
	query := `
		SELECT timestamp 
		FROM browser_logs 
		WHERE node_name = ? AND url = ? AND success = 1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	err := d.sqlDB.QueryRow(query, nodeName, url).Scan(&timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}

func (d *DB) Cleanup(days int) error {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()

	// Delete old ping logs
	if _, err := d.sqlDB.Exec("DELETE FROM ping_logs WHERE timestamp < ?", cutoff); err != nil {
		return err
	}

	// Delete old browser logs
	if _, err := d.sqlDB.Exec("DELETE FROM browser_logs WHERE timestamp < ?", cutoff); err != nil {
		return err
	}

	// Reclaim space
	if _, err := d.sqlDB.Exec("VACUUM"); err != nil {
		return err
	}

	return nil
}

func (d *DB) Close() error {
	return d.sqlDB.Close()
}
