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

	query2 := `DELETE FROM bandwidth_logs WHERE timestamp < ?`
	_, err = d.sqlDB.Exec(query2, cutoff)
	if err != nil {
		return err
	}

	query3 := `DELETE FROM browser_logs WHERE timestamp < ?`
	_, err = d.sqlDB.Exec(query3, cutoff)
	return err
}

type NodeScore struct {
	Name               string
	Score              int
	BaseScore          int
	SuccessRate        float64
	AvgDelay           float64
	Jitter             int
	AvgBandwidth       float64
	TotalConsumedBytes int64
	BrowserSuccessRate float64
	AvgBrowserLoadTime float64
}

func (d *DB) GetScores(days int) (map[string]NodeScore, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	query := `
		SELECT 
			p.node_name, 
			COUNT(p.id) as total_tests,
			SUM(CASE WHEN p.success THEN 1 ELSE 0 END) as success_tests,
			AVG(CASE WHEN p.success THEN p.delay ELSE NULL END) as avg_delay,
			MAX(CASE WHEN p.success THEN p.delay ELSE NULL END) - MIN(CASE WHEN p.success THEN p.delay ELSE NULL END) as jitter
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
		var jitter sql.NullInt64

		if err := rows.Scan(&name, &total, &successCount, &avgDelay, &jitter); err != nil {
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

		// 公式： (成功率 * 10000) - (平均延遲) - (Jitter * 2)
		j := 0
		if jitter.Valid {
			j = int(jitter.Int64)
		}
		
		score := int(successRate*10000) - int(delay) - (j * 2)

		scores[name] = NodeScore{
			Name:               name,
			Score:              score,
			BaseScore:          score,
			SuccessRate:        successRate,
			AvgDelay:           delay,
			Jitter:             j,
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

	// 取得瀏覽器測試結果
	brQuery := `
		SELECT node_name, AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END), AVG(CASE WHEN success THEN load_time_ms ELSE NULL END)
		FROM browser_logs 
		WHERE timestamp >= ? 
		GROUP BY node_name
	`
	brRows, err := d.sqlDB.Query(brQuery, cutoff)
	if err == nil {
		defer brRows.Close()
		for brRows.Next() {
			var name string
			var successRate sql.NullFloat64
			var avgLoad sql.NullFloat64
			if err := brRows.Scan(&name, &successRate, &avgLoad); err == nil {
				if sc, exists := scores[name]; exists {
					if successRate.Valid {
						sc.BrowserSuccessRate = successRate.Float64
						// 成功率太低則扣分
						if successRate.Float64 < 0.5 {
							sc.Score -= 1000
						}
					}
					if avgLoad.Valid {
						sc.AvgBrowserLoadTime = avgLoad.Float64
						// 載入時間加分：越快分數越高。基準為 5 秒(5000ms)，每快 100ms 加 10 分
						bonus := int(5000 - avgLoad.Float64) / 10
						if bonus > 0 {
							sc.Score += bonus
						}
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

func (d *DB) Cleanup(days int) error {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
	
	// Delete old ping logs
	if _, err := d.sqlDB.Exec("DELETE FROM ping_logs WHERE timestamp < ?", cutoff); err != nil {
		return err
	}
	
	// Delete old bandwidth logs
	if _, err := d.sqlDB.Exec("DELETE FROM bandwidth_logs WHERE timestamp < ?", cutoff); err != nil {
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
