package main

import (
	"database/sql"
	"log"
	"math"
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
	AdjustedRate       float64
	AvgDelay           float64
	Jitter             int
	SampleCount        int
	AvgBandwidth       float64
	TotalConsumedBytes int64
	BrowserSuccessRate float64
	AvgBrowserLoadTime float64
	BrowserTested      bool
	LastPingTime       int64
	LastBandwidthTime  int64
	LastBrowserTime    int64
}

func (d *DB) GetScores(days int) (map[string]NodeScore, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	nowUnix := float64(time.Now().Unix())

	// Ping 的半衰期保持 2 小時 (因為測試頻率高，對短期網路波動極度敏感)
	lambdaPing := math.Ln2 / (2.0 * 3600.0)

	// 頻寬與瀏覽器面試的半衰期改為 24 小時 (因為面試頻率極低，需要較長期的記憶)
	lambdaInterview := math.Ln2 / (24.0 * 3600.0)

	// ========================================
	// 第一層：撈取所有 ping_logs 原始行，在 Go 端計算
	// ========================================
	// Ping 日誌的半衰期僅有 2 小時，超過 24 小時的權重極低，因此最多只撈取過去 24 小時的資料以節省 CPU 與記憶體
	pingCutoff := cutoff
	if float64(nowUnix)-float64(pingCutoff) > 24*3600 {
		pingCutoff = int64(nowUnix) - 24*3600
	}

	pingQuery := `
		SELECT node_name, timestamp, delay, success
		FROM ping_logs
		WHERE timestamp >= ?
		ORDER BY node_name
	`
	pingRows, err := d.sqlDB.Query(pingQuery, pingCutoff)
	if err != nil {
		return nil, err
	}
	defer pingRows.Close()

	// 每個節點的原始資料收集
	type nodeRawData struct {
		successWeightSum float64
		totalWeightSum   float64
		delayWeightedSum float64
		delayWeightSum   float64
		delays           []float64 // 成功的延遲值，用於計算標準差
		delayWeights     []float64 // 對應的權重
		totalCount       int
		successCount     int
		lastPingTime     int64
	}

	nodeData := make(map[string]*nodeRawData)

	for pingRows.Next() {
		var name string
		var timestamp int64
		var delay int
		var success bool

		if err := pingRows.Scan(&name, &timestamp, &delay, &success); err != nil {
			log.Printf("掃描 ping_logs 發生錯誤: %v", err)
			continue
		}

		nd, exists := nodeData[name]
		if !exists {
			nd = &nodeRawData{}
			nodeData[name] = nd
		}

		if timestamp > nd.lastPingTime {
			nd.lastPingTime = timestamp
		}

		// 計算時間衰減權重
		ageSeconds := nowUnix - float64(timestamp)
		weight := math.Exp(-lambdaPing * ageSeconds)

		nd.totalCount++
		nd.totalWeightSum += weight

		if success {
			nd.successCount++
			nd.successWeightSum += weight
			nd.delayWeightedSum += float64(delay) * weight
			nd.delayWeightSum += weight
			nd.delays = append(nd.delays, float64(delay))
			nd.delayWeights = append(nd.delayWeights, weight)
		}
	}

	scores := make(map[string]NodeScore)

	for name, nd := range nodeData {
		if nd.totalCount == 0 {
			continue
		}

		// 原始成功率（顯示用）
		rawSuccessRate := float64(nd.successCount) / float64(nd.totalCount)

		// Bayesian 修正成功率（計算用）：虛擬加入 2 成功 + 1 失敗
		adjustedRate := float64(nd.successCount+2) / float64(nd.totalCount+3)

		// 加權成功率（用於評分）
		weightedSuccessRate := adjustedRate
		if nd.totalWeightSum > 0 {
			weightedSuccessRate = nd.successWeightSum / nd.totalWeightSum
			// 將 Bayesian 修正也套用在加權版本上
			if nd.totalCount < 10 {
				// 樣本量小時，混合原始 Bayesian 修正與加權結果
				blendFactor := float64(nd.totalCount) / 10.0
				weightedSuccessRate = weightedSuccessRate*blendFactor + adjustedRate*(1-blendFactor)
			}
		}

		// 加權平均延遲
		avgDelay := 9999.0
		if nd.delayWeightSum > 0 {
			avgDelay = nd.delayWeightedSum / nd.delayWeightSum
		}

		// Jitter：加權標準差
		jitter := 0
		if len(nd.delays) > 1 && nd.delayWeightSum > 0 {
			wMean := nd.delayWeightedSum / nd.delayWeightSum
			var varianceSum float64
			var wSum float64
			for i, d := range nd.delays {
				w := nd.delayWeights[i]
				diff := d - wMean
				varianceSum += w * diff * diff
				wSum += w
			}
			if wSum > 0 {
				jitter = int(math.Sqrt(varianceSum / wSum))
			}
		}

		// V4 評分公式：(加權成功率 × 3000) − (加權平均延遲) − (Jitter標準差 × 2)
		score := int(weightedSuccessRate*3000) - int(avgDelay) - (jitter * 2)

		scores[name] = NodeScore{
			Name:               name,
			Score:              score + 1000 + 5500, // 預先加上樂觀初始值 (Bandwidth: 1000, Browser: 5500)
			BaseScore:          score,
			SuccessRate:        rawSuccessRate,
			AdjustedRate:       adjustedRate,
			AvgDelay:           avgDelay,
			Jitter:             jitter,
			SampleCount:        nd.totalCount,
			AvgBandwidth:       0.0,
			TotalConsumedBytes: 0,
			LastPingTime:       nd.lastPingTime,
		}
	}

	// ========================================
	// 第二層：頻寬加分（對數遞減 + 硬性上限）
	// ========================================
	bwQuery := `
		SELECT node_name, speed_kbps, downloaded_bytes, timestamp 
		FROM bandwidth_logs 
		WHERE timestamp >= ?
	`
	type bwRawData struct {
		weightedSpeedSum float64
		weightSum        float64
		sumBytes         int64
		maxTime          int64
	}
	bwData := make(map[string]*bwRawData)

	bwRows, err := d.sqlDB.Query(bwQuery, cutoff)
	if err == nil {
		defer bwRows.Close()
		for bwRows.Next() {
			var name string
			var speed float64
			var bytes int64
			var timestamp int64
			if err := bwRows.Scan(&name, &speed, &bytes, &timestamp); err == nil {
				bd, exists := bwData[name]
				if !exists {
					bd = &bwRawData{}
					bwData[name] = bd
				}

				ageSeconds := nowUnix - float64(timestamp)
				weight := math.Exp(-lambdaInterview * ageSeconds)

				bd.weightedSpeedSum += speed * weight
				bd.weightSum += weight
				bd.sumBytes += bytes
				if timestamp > bd.maxTime {
					bd.maxTime = timestamp
				}
			}
		}

		for name, bd := range bwData {
			if sc, exists := scores[name]; exists && bd.weightSum > 0 {
				sc.Score -= 1000 // 扣除樂觀初始值

				// 貝式平滑先驗 (Bayesian Prior)：虛擬加入一筆優良的歷史成績
				// 確保長久未測速 (weightSum 衰減趨近於 0) 的節點分數會緩慢回升，重新獲得被選為面試節點的機會
				priorWeight := 0.5
				priorSpeed := 10240.0 // 10 MB/s 樂觀速度

				sc.AvgBandwidth = (bd.weightedSpeedSum + priorWeight*priorSpeed) / (bd.weightSum + priorWeight)
				
				// V4: 對數遞減，前 2 MB/s 加分最快，之後邊際遞減
				// 上限 2000 分
				bwBonus := int(math.Log2(1+sc.AvgBandwidth/1024.0) * 1000)
				if bwBonus > 2000 {
					bwBonus = 2000
				}
				sc.Score += bwBonus
				sc.TotalConsumedBytes = bd.sumBytes
				sc.LastBandwidthTime = bd.maxTime
				scores[name] = sc
			}
		}
	}

	// ========================================
	// 第三層：瀏覽器測試修正（連續懲罰函數）
	// ========================================
	brQuery := `
		SELECT node_name, success, load_time_ms, timestamp
		FROM browser_logs 
		WHERE timestamp >= ?
	`
	type brRawData struct {
		weightedSuccessSum  float64
		weightSum           float64
		weightedLoadTimeSum float64
		loadTimeWeightSum   float64
		maxTime             int64
	}
	brData := make(map[string]*brRawData)

	brRows, err := d.sqlDB.Query(brQuery, cutoff)
	if err == nil {
		defer brRows.Close()
		for brRows.Next() {
			var name string
			var success bool
			var loadTime int
			var timestamp int64
			if err := brRows.Scan(&name, &success, &loadTime, &timestamp); err == nil {
				br, exists := brData[name]
				if !exists {
					br = &brRawData{}
					brData[name] = br
				}

				ageSeconds := nowUnix - float64(timestamp)
				weight := math.Exp(-lambdaInterview * ageSeconds)

				br.weightSum += weight
				if success {
					br.weightedSuccessSum += weight
					br.weightedLoadTimeSum += float64(loadTime) * weight
					br.loadTimeWeightSum += weight
				}
				if timestamp > br.maxTime {
					br.maxTime = timestamp
				}
			}
		}

		for name, br := range brData {
			if sc, exists := scores[name]; exists && br.weightSum > 0 {
				sc.BrowserTested = true
				sc.Score -= 5500 // 扣除樂觀初始值

				// 貝式平滑先驗 (Bayesian Prior)：虛擬加入一筆成功的歷史成績
				priorWeight := 0.5
				priorSuccess := 1.0
				priorLoadTime := 1000.0 // 樂觀載入時間 1000ms

				sc.BrowserSuccessRate = (br.weightedSuccessSum + priorWeight*priorSuccess) / (br.weightSum + priorWeight)
				successBonus := int(sc.BrowserSuccessRate * 4000)
				sc.Score += successBonus

				totalLoadTimeWeight := br.loadTimeWeightSum + priorWeight
				sc.AvgBrowserLoadTime = (br.weightedLoadTimeSum + priorWeight*priorLoadTime) / totalLoadTimeWeight
				bonus := int(5000-sc.AvgBrowserLoadTime) * 6 / 10
				if bonus > 3000 {
					bonus = 3000
				}
				if bonus > 0 {
					sc.Score += bonus
				}
				
				sc.LastBrowserTime = br.maxTime
				scores[name] = sc
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
