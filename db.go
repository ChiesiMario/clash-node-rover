package main

import (
	"database/sql"
	"log"
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
	CREATE TABLE IF NOT EXISTS bandwidth_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_name TEXT,
		timestamp INTEGER,
		speed_kbps REAL,
		downloaded_bytes INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_node_time ON ping_logs(node_name, timestamp);
	`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	// 自動升級資料庫結構 (如果舊版沒有 downloaded_bytes 欄位)
	db.Exec("ALTER TABLE bandwidth_logs ADD COLUMN downloaded_bytes INTEGER DEFAULT 0;")

	return &DB{sqlDB: db}, nil
}

func (d *DB) InsertLog(nodeName string, delay int, success bool) error {
	query := `INSERT INTO ping_logs (node_name, timestamp, delay, success) VALUES (?, ?, ?, ?)`
	_, err := d.sqlDB.Exec(query, nodeName, time.Now().Unix(), delay, success)
	return err
}

func (d *DB) InsertBandwidthLog(nodeName string, speedKbps float64, downloadedBytes int64) error {
	query := `INSERT INTO bandwidth_logs (node_name, timestamp, speed_kbps, downloaded_bytes) VALUES (?, ?, ?, ?)`
	_, err := d.sqlDB.Exec(query, nodeName, time.Now().Unix(), speedKbps, downloadedBytes)
	return err
}

func (d *DB) CleanOldLogs(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	
	query1 := `DELETE FROM ping_logs WHERE timestamp < ?`
	_, err := d.sqlDB.Exec(query1, cutoff)
	if err != nil {
		return err
	}

	query2 := `DELETE FROM bandwidth_logs WHERE timestamp < ?`
	_, err = d.sqlDB.Exec(query2, cutoff)
	return err
}

type NodeScore struct {
	Name               string
	Score              int
	SuccessRate        float64
	AvgDelay           float64
	AvgBandwidth       float64
	TotalConsumedBytes int64
}

func (d *DB) GetScores(days int) (map[string]NodeScore, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	query := `
		SELECT 
			p.node_name, 
			COUNT(p.id) as total_tests,
			SUM(CASE WHEN p.success THEN 1 ELSE 0 END) as success_tests,
			AVG(CASE WHEN p.success THEN p.delay ELSE NULL END) as avg_delay
		FROM ping_logs p
		WHERE p.timestamp >= ?
		GROUP BY p.node_name
	`

	rows, err := d.sqlDB.Query(query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make(map[string]NodeScore)

	for rows.Next() {
		var name string
		var total, successCount int
		var avgDelay sql.NullFloat64

		if err := rows.Scan(&name, &total, &successCount, &avgDelay); err != nil {
			log.Printf("掃描資料庫發生錯誤: %v", err)
			continue
		}

		if total == 0 {
			continue
		}

		successRate := float64(successCount) / float64(total)
		delay := 9999.0
		if avgDelay.Valid {
			delay = avgDelay.Float64
		}

		// 公式： (成功率 * 10000) - (平均延遲)
		score := int(successRate*10000) - int(delay)

		scores[name] = NodeScore{
			Name:               name,
			Score:              score,
			SuccessRate:        successRate,
			AvgDelay:           delay,
			AvgBandwidth:       0.0,
			TotalConsumedBytes: 0,
		}
	}

	// 取得平均頻寬與總消耗流量
	bwQuery := `
		SELECT node_name, AVG(speed_kbps), SUM(downloaded_bytes) 
		FROM bandwidth_logs 
		WHERE timestamp >= ? 
		GROUP BY node_name
	`
	bwRows, err := d.sqlDB.Query(bwQuery, cutoff)
	if err == nil {
		defer bwRows.Close()
		for bwRows.Next() {
			var name string
			var avgBw sql.NullFloat64
			var sumBytes sql.NullInt64
			if err := bwRows.Scan(&name, &avgBw, &sumBytes); err == nil {
				if sc, exists := scores[name]; exists {
					if avgBw.Valid {
						sc.AvgBandwidth = avgBw.Float64
						// Score Algorithm V2: 將下載速度加入質量分數計算
						// 每 1 KB/s 增加 0.5 分 (等同於每 1 MB/s 增加約 500 分)
						sc.Score += int(sc.AvgBandwidth / 2)
					}
					if sumBytes.Valid {
						sc.TotalConsumedBytes = sumBytes.Int64
					}
					scores[name] = sc
				}
			}
		}
	}

	return scores, nil
}

type PingLog struct {
	Timestamp int64
	Delay     int
}

func (d *DB) GetNodeHistory(nodeName string, hours int) ([]PingLog, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
	
	query := `
		SELECT timestamp, delay 
		FROM ping_logs 
		WHERE node_name = ? AND timestamp >= ? AND success = 1
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

func (d *DB) Close() error {
	return d.sqlDB.Close()
}
