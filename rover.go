package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Rover struct {
	cfg           *Config
	api           *APIClient
	db            *DB
	failedConsec  map[string]int
	lastCheckTime map[string]time.Time
}

func NewRover(cfg *Config, api *APIClient, db *DB) *Rover {
	return &Rover{
		cfg:           cfg,
		api:           api,
		db:            db,
		failedConsec:  make(map[string]int),
		lastCheckTime: make(map[string]time.Time),
	}
}

func (r *Rover) pickRandomURL() string {
	if len(r.cfg.TestURLs) == 0 {
		return "http://www.gstatic.com/generate_204"
	}
	return r.cfg.TestURLs[rand.Intn(len(r.cfg.TestURLs))]
}

func (r *Rover) Run() {
	log.Printf("開始監控 Node Rover，目標群組 '%s'，每 %v 檢查一次", r.cfg.TargetGroup, r.cfg.CheckInterval)
	ticker := time.NewTicker(r.cfg.CheckInterval)
	defer ticker.Stop()

	r.checkAndSwitch()

	for range ticker.C {
		r.checkAndSwitch()
	}
}

func (r *Rover) checkAndSwitch() {
	log.Println("--- 檢查節點 ---")

	if err := r.db.CleanOldLogs(r.cfg.HistoryDays); err != nil {
		log.Printf("清理舊日誌失敗: %v", err)
	}

	group, err := r.api.GetProxyGroup(r.cfg.TargetGroup)
	if err != nil {
		log.Printf("取得代理群組時發生錯誤: %v", err)
		return
	}

	if len(group.All) == 0 {
		log.Println("目標群組中沒有找到節點。")
		return
	}

	type nodeStat struct {
		Name  string
		Delay int
		Err   error
	}

	now := time.Now()
	var nodesToTest []string

	// 指數退避檢查
	for _, name := range group.All {
		fails := r.failedConsec[name]
		if fails > 0 {
			backoffMins := int(math.Pow(2, float64(fails-1)))
			if backoffMins > r.cfg.MaxBackoffMinutes {
				backoffMins = r.cfg.MaxBackoffMinutes
			}

			lastCheck := r.lastCheckTime[name]
			if now.Sub(lastCheck) < time.Duration(backoffMins)*time.Minute {
				log.Printf("節點 [%s] 處於退避期 (連續失敗 %d 次，退避 %d 分鐘)，跳過測速。", name, fails, backoffMins)
				continue
			}
		}
		nodesToTest = append(nodesToTest, name)
	}

	if len(nodesToTest) == 0 {
		log.Println("所有節點都在退避期，本次跳過檢查。")
		return
	}

	stats := make(chan nodeStat, len(nodesToTest))
	jobs := make(chan string, len(nodesToTest))

	var wg sync.WaitGroup
	workerCount := r.cfg.MaxConcurrent
	if workerCount > len(nodesToTest) {
		workerCount = len(nodesToTest)
	}

	// 啟動 Worker Pool
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range jobs {
				testUrl := r.pickRandomURL()
				delay, err := r.api.TestProxyDelay(name, testUrl, r.cfg.TestTimeout)
				stats <- nodeStat{Name: name, Delay: delay, Err: err}
			}
		}()
	}

	for _, name := range nodesToTest {
		r.lastCheckTime[name] = time.Now()
		jobs <- name
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(stats)
	}()

	var statResults []nodeStat
	// 1. 邊測邊寫入資料庫，讓網頁能即時看到進度
	for s := range stats {
		statResults = append(statResults, s)

		success := (s.Err == nil && s.Delay > 0)
		if success {
			r.failedConsec[s.Name] = 0
		} else {
			r.failedConsec[s.Name]++
		}

		if err := r.db.InsertLog(s.Name, s.Delay, success); err != nil {
			log.Printf("寫入日誌失敗 [%s]: %v", s.Name, err)
		}
	}

	// 2. 重新取得包含最新紀錄的歷史分數
	scores, err := r.db.GetScores(r.cfg.HistoryDays)
	if err != nil {
		log.Printf("取得歷史分數失敗: %v", err)
	}

	fastestNode := ""
	fastestDelay := math.MaxInt32

	highestScoreNode := ""
	highestScore := math.MinInt32
	highestScoreCurrentDelay := 0

	// 3. 評估與印出日誌
	for _, s := range statResults {
		if s.Err != nil {
			log.Printf("節點 [%s] 測試失敗: %v", s.Name, s.Err)
		} else {
			scoreData, ok := scores[s.Name]
			scoreStr := "無資料"
			if ok {
				scoreStr = fmt.Sprintf("%d (成功率: %.1f%%)", scoreData.Score, scoreData.SuccessRate*100)
			}

			log.Printf("節點 [%s] 延遲: %d 毫秒 | 質量分數: %s", s.Name, s.Delay, scoreStr)

			if s.Delay > 0 && s.Delay < fastestDelay {
				fastestDelay = s.Delay
				fastestNode = s.Name
			}

			if ok && scoreData.Score > highestScore {
				highestScore = scoreData.Score
				highestScoreNode = s.Name
				highestScoreCurrentDelay = s.Delay
			}
		}
	}

	if fastestNode == "" {
		log.Println("所有參與測試的節點皆失敗，保留目前節點。")
		return
	}

	targetNode := fastestNode
	reason := "當前速度最快"

	if highestScoreNode != "" {
		diff := highestScoreCurrentDelay - fastestDelay
		if diff <= r.cfg.DelayTolerance {
			targetNode = highestScoreNode
			reason = fmt.Sprintf("最高質量分，且與最快節點差距僅 %d 毫秒", diff)
		} else {
			reason = fmt.Sprintf("最高質量分節點比最快節點慢太多 (差距 %d 毫秒)，強制使用最快節點", diff)
		}
	}

	log.Printf("初步選擇節點: [%s] | 理由: %s", targetNode, reason)

	if targetNode != group.Now {
		log.Printf("從 [%s] 切換至更好的節點 [%s]", group.Now, targetNode)
		if err := r.api.SelectProxy(r.cfg.TargetGroup, targetNode); err != nil {
			log.Printf("切換代理節點失敗: %v", err)
		} else {
			log.Println("成功切換代理節點。")
			group.Now = targetNode
		}
	} else {
		log.Printf("目前的節點 [%s] 依然是最佳選擇。", group.Now)
	}

	// 頻寬測試 (針對當下使用的節點)
	log.Printf("對當下節點 [%s] 進行頻寬測試...", group.Now)
	speedKBps, totalBytes, err := r.api.TestBandwidth(r.cfg.BandwidthTestURL, r.cfg.ClashProxyURL, 15*time.Second)
	if err != nil {
		log.Printf("頻寬測試失敗: %v", err)
		// 懲罰機制：寫入一筆極大延遲的失敗紀錄
		r.db.InsertLog(targetNode, 9999, false)
	} else {
		log.Printf("頻寬測試完成: %.2f KB/s", speedKBps)
		r.db.InsertBandwidthLog(targetNode, speedKBps, totalBytes)
		
		if speedKBps < r.cfg.BandwidthThresholdKbps {
			log.Printf("警告: 節點 [%s] 頻寬低於閾值 (%.0f KB/s)，寫入劣跡懲罰", targetNode, r.cfg.BandwidthThresholdKbps)
			r.db.InsertLog(targetNode, 9999, false)
		} else {
			// 速度達標，給予獎勵紀錄
			r.db.InsertLog(targetNode, int(1000/speedKBps), true) // 用速度換算一個低延遲作為獎勵
		}
	}
}
