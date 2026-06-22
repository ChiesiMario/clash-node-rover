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
	cfg               *Config
	api               *APIClient
	db                *DB
	consecutiveFailed int
	failedConsec      map[string]int
	lastCheckTime     map[string]time.Time
	lastBandwidthTest map[string]time.Time
	
	// 進階功能控制
	ManualTrigger     chan struct{}
	Quit              chan struct{}
	IsRunning         bool
}

func NewRover(cfg *Config, api *APIClient, db *DB) *Rover {
	return &Rover{
		cfg:               cfg,
		api:               api,
		db:                db,
		failedConsec:      make(map[string]int),
		lastCheckTime:     make(map[string]time.Time),
		lastBandwidthTest: make(map[string]time.Time),
		ManualTrigger:     make(chan struct{}, 1),
		Quit:              make(chan struct{}, 1),
		IsRunning:         false,
	}
}

func (r *Rover) pickRandomURL() string {
	if len(r.cfg.TestURLs) == 0 {
		return "http://www.gstatic.com/generate_204"
	}
	return r.cfg.TestURLs[rand.Intn(len(r.cfg.TestURLs))]
}

func (r *Rover) Start() {
	log.Println("Clash Node Rover 啟動...")
	r.runCheckCycle() // 啟動時先跑一次

	ticker := time.NewTicker(r.cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.runCheckCycle()
		case <-r.ManualTrigger:
			log.Println("收到手動測速信號，立即執行！")
			ticker.Stop()
			r.runCheckCycle()
			ticker.Reset(r.cfg.CheckInterval)
		case <-r.Quit:
			log.Println("背景測速引擎已停止。")
			return
		}
	}
}

func (r *Rover) Stop() {
	r.Quit <- struct{}{}
}

func (r *Rover) runCheckCycle() {
	if r.IsRunning {
		return
	}
	r.IsRunning = true
	defer func() { r.IsRunning = false }()

	log.Println("----------------------------------------")
	log.Println("開始新一輪節點測試...")

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

	// 4. 反壟斷公平探索機制 (Monopoly Breaker)
	// 尋找「成功率高、BaseScore 高，且不在測速冷卻期內」的潛力節點進行頻寬測試
	var bwTestCandidate string
	highestBaseScore := -999999
	targetIntervalDuration := time.Duration(r.cfg.BandwidthTestInterval) * time.Minute
	explorationDuration := time.Duration(r.cfg.ExplorationCooldown) * time.Minute

	// 確保至少取得最新的分數資訊來判斷
	scores, _ = r.db.GetScores(r.cfg.HistoryDays)

	for name, sc := range scores {
		// 只面試「優等生」：成功率至少 80%，避免切換後連線中斷
		if sc.SuccessRate < 0.8 {
			continue
		}
		// 必須已過探索冷卻期
		if time.Since(r.lastBandwidthTest[name]) >= explorationDuration {
			if sc.BaseScore > highestBaseScore {
				highestBaseScore = sc.BaseScore
				bwTestCandidate = name
			}
		}
	}

	// 如果所有好節點都在冷卻期，但 targetNode 尚未測速過 (例如剛開機)，就測 targetNode
	if bwTestCandidate == "" && targetNode != "" && time.Since(r.lastBandwidthTest[targetNode]) >= targetIntervalDuration {
		bwTestCandidate = targetNode
	}

		if bwTestCandidate != "" {
		if bwTestCandidate != targetNode && targetNode != "" {
			log.Printf("💡 觸發反壟斷探索機制：切暫時換至潛力節點 [%s] 進行測速 (BaseScore: %d)", bwTestCandidate, highestBaseScore)
			err := r.api.SelectProxy(r.cfg.TargetGroup, bwTestCandidate)
			if err != nil {
				log.Printf("無法切換至候選節點: %v", err)
				bwTestCandidate = "" // 取消本次面試
			}
		} else {
			log.Printf("選定目標節點: [%s]，準備進行真實頻寬測試...", bwTestCandidate)
		}

		if bwTestCandidate != "" {
			r.lastBandwidthTest[bwTestCandidate] = time.Now()
			
			// 透過 Clash proxy 進行下載測試
			speedKBps, totalBytes, err := r.api.TestBandwidth(r.cfg.BandwidthTestURL, r.cfg.ClashProxyURL, 15*time.Second)
			
			if err != nil {
				log.Printf("頻寬測試失敗: %v", err)
				// 真實的連線異常或 HTTP 錯誤，才視為失敗並寫入 Ping 紀錄
				r.db.InsertLog(bwTestCandidate, 9999, false)
			} else {
				log.Printf("頻寬測試完成: %.2f KB/s", speedKBps)
				r.db.InsertBandwidthLog(bwTestCandidate, speedKBps, totalBytes)
			}

			// 如果有切換過代理，記得切回冠軍節點
			if bwTestCandidate != targetNode && targetNode != "" {
				log.Printf("探索測速完成，將代理切回最佳節點 [%s]", targetNode)
				r.api.SelectProxy(r.cfg.TargetGroup, targetNode)
			}
		}
	} else if targetNode != "" {
		log.Printf("目前所有優質節點皆在頻寬測速冷卻期 (%d 分鐘) 內，跳過下載測試以節省流量。", r.cfg.BandwidthTestInterval)
	}
}
