package main

import (
	"fmt"
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
	logHeader("Clash Node Rover 啟動")

	// 啟動時先執行一次資料庫瘦身
	if err := r.db.Cleanup(r.cfg.CleanupDays); err != nil {
		logError("資料庫自動瘦身失敗: %v", err)
	} else {
		logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.cfg.CleanupDays))
	}

	r.runCheckCycle() // 啟動時先跑一次

	if r.cfg.EnableFailover {
		go r.runFailoverWatchdog()
	}

	ticker := time.NewTicker(r.cfg.CheckInterval)
	defer ticker.Stop()

	// 每天執行一次資料庫瘦身
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			r.runCheckCycle()
		case <-cleanupTicker.C:
			if err := r.db.Cleanup(r.cfg.CleanupDays); err != nil {
				logError("資料庫自動瘦身失敗: %v", err)
			} else {
				logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.cfg.CleanupDays))
			}
		case <-r.ManualTrigger:
			logInfo("收到手動測速信號，立即執行！")
			ticker.Stop()
			r.runCheckCycle()
			ticker.Reset(r.cfg.CheckInterval)
		case <-r.Quit:
			logWarning("背景測速引擎已停止。")
			return
		}
	}
}

func (r *Rover) Stop() {
	r.Quit <- struct{}{}
}

type nodeStat struct {
	Name  string
	Delay int
	Err   error
}

func (r *Rover) runCheckCycle() {
	if r.IsRunning {
		return
	}
	r.IsRunning = true
	defer func() { r.IsRunning = false }()

	logHeader("開始新一輪節點測試 (多群組並行模式)")

	groupNodesMap := make(map[string][]string)
	groupNowMap := make(map[string]string)
	uniqueNodes := make(map[string]bool)

	for _, groupName := range r.cfg.TargetGroups {
		group, err := r.api.GetProxyGroup(groupName)
		if err != nil {
			logError("取得代理群組 [%s] 時發生錯誤: %v", groupName, err)
			continue
		}
		if len(group.All) == 0 {
			logWarning("代理群組 [%s] 中沒有找到節點。", groupName)
			continue
		}
		groupNodesMap[groupName] = group.All
		groupNowMap[groupName] = group.Now
		for _, name := range group.All {
			uniqueNodes[name] = true
		}
	}

	if len(uniqueNodes) == 0 {
		logWarning("所有目標群組中都沒有有效節點，退出本次檢查。")
		return
	}

	now := time.Now()
	var nodesToTest []string

	// 指數退避檢查
	for name := range uniqueNodes {
		fails := r.failedConsec[name]
		if fails > 0 {
			backoffMins := int(math.Pow(2, float64(fails-1)))
			if backoffMins > r.cfg.MaxBackoffMinutes {
				backoffMins = r.cfg.MaxBackoffMinutes
			}

			lastCheck := r.lastCheckTime[name]
			if now.Sub(lastCheck) < time.Duration(backoffMins)*time.Minute {
				logMuted("節點 %s 處於退避期 (連續失敗 %d 次，退避 %d 分鐘)，跳過測速。", formatNode(name), fails, backoffMins)
				continue
			}
		}
		nodesToTest = append(nodesToTest, name)
	}

	if len(nodesToTest) == 0 {
		logWarning("所有節點都在退避期，本次跳過檢查。")
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

	statResultsMap := make(map[string]nodeStat)
	for s := range stats {
		statResultsMap[s.Name] = s

		success := (s.Err == nil && s.Delay > 0)
		if success {
			r.failedConsec[s.Name] = 0
		} else {
			r.failedConsec[s.Name]++
		}

		if err := r.db.InsertLog(s.Name, s.Delay, success); err != nil {
			logError("寫入日誌失敗 %s: %v", formatNode(s.Name), err)
		}
	}

	scores, err := r.db.GetScores(r.cfg.HistoryDays)
	if err != nil {
		logError("取得歷史分數失敗: %v", err)
	}

	// 3. 為每個群組獨立計算最佳節點
	groupTargetNodes := make(map[string]string)
	pendingSwitches := make(map[string]string)

	for _, groupName := range r.cfg.TargetGroups {
		nodes, ok := groupNodesMap[groupName]
		if !ok || len(nodes) == 0 {
			continue
		}

		fastestNode := ""
		fastestDelay := math.MaxInt32

		highestScoreNode := ""
		highestScore := math.MinInt32
		highestScoreCurrentDelay := 0

		for _, name := range nodes {
			s, tested := statResultsMap[name]
			if !tested || s.Err != nil {
				continue
			}

			if s.Delay > 0 && s.Delay < fastestDelay {
				fastestDelay = s.Delay
				fastestNode = s.Name
			}

			scoreData, okScore := scores[s.Name]
			if okScore && scoreData.Score > highestScore {
				highestScore = scoreData.Score
				highestScoreNode = s.Name
				highestScoreCurrentDelay = s.Delay
			}
		}

		if fastestNode == "" {
			logGroup(groupName, colorWarning.Sprintf("所有參與測試的節點皆失敗，保留目前節點。"))
			groupTargetNodes[groupName] = groupNowMap[groupName]
			continue
		}

		targetNode := fastestNode
		reason := colorInfo.Sprint("當前速度最快")

		if highestScoreNode != "" {
			diff := highestScoreCurrentDelay - fastestDelay
			if diff <= r.cfg.DelayTolerance {
				targetNode = highestScoreNode
				reason = colorSuccess.Sprintf("最高質量分，且與最快節點差距僅 %s 毫秒", formatVal(diff))
			} else {
				reason = colorWarning.Sprintf("最高分節點比最快節點慢 %s 毫秒，強制使用最快節點", formatVal(diff))
			}
		}

		logGroup(groupName, "選擇節點: %s | 理由: %s", formatNode(targetNode), reason)
		groupTargetNodes[groupName] = targetNode

		if targetNode != groupNowMap[groupName] && groupNowMap[groupName] != "" {
			logGroup(groupName, colorWarning.Sprintf("預計從 %s 切換至 %s (將於測速結束後執行)", formatNode(groupNowMap[groupName]), formatNode(targetNode)))
			pendingSwitches[groupName] = targetNode
		} else {
			logGroup(groupName, colorSuccess.Sprintf("目前的節點 %s 依然是最佳選擇。", formatNode(targetNode)))
		}
	}

	// 4. 反壟斷公平探索機制 (Monopoly Breaker) & 頻寬測速
	targetIntervalDuration := time.Duration(r.cfg.BandwidthTestInterval) * time.Minute
	explorationDuration := time.Duration(r.cfg.ExplorationCooldown) * time.Minute

	scores, _ = r.db.GetScores(r.cfg.HistoryDays)

	alreadyTestedInCycle := make(map[string]bool)

	for _, groupName := range r.cfg.TargetGroups {
		nodes, ok := groupNodesMap[groupName]
		if !ok || len(nodes) == 0 {
			continue
		}

		var bwTestCandidate string
		highestBaseScore := -999999

		// 1. 在此群組內尋找探索面試候選人
		for _, name := range nodes {
			sc, ok := scores[name]
			if !ok || sc.SuccessRate < 0.8 {
				continue
			}
			if time.Since(r.lastBandwidthTest[name]) >= explorationDuration {
				if sc.BaseScore > highestBaseScore {
					highestBaseScore = sc.BaseScore
					bwTestCandidate = name
				}
			}
		}

		// 2. 如果沒有面試候選人，看原本的最佳節點是否需要測速
		targetNode := groupTargetNodes[groupName]
		if bwTestCandidate == "" && targetNode != "" {
			if time.Since(r.lastBandwidthTest[targetNode]) >= targetIntervalDuration {
				bwTestCandidate = targetNode
			}
		}

		if bwTestCandidate == "" {
			logGroup(groupName, colorMuted.Sprintf("目前所有優質節點皆在頻寬測速冷卻期內，跳過下載測試以節省流量。"))
			continue
		}

		if alreadyTestedInCycle[bwTestCandidate] {
			logGroup(groupName, colorMuted.Sprintf("節點 %s 在本週期已經測速過，跳過重複測速。", formatNode(bwTestCandidate)))
			continue
		}

		// 準備測速
		var borrowGroup string
		var originalTarget string

		if r.cfg.DedicatedTestGroup != "" {
			borrowGroup = r.cfg.DedicatedTestGroup
			// 取得專屬測速群組目前的節點，以便測速完可以切回來
			if g, err := r.api.GetProxyGroup(borrowGroup); err == nil {
				originalTarget = g.Now
			}
		} else {
			// 如果沒有設定專屬群組，就借用目前正在處理的這個群組
			borrowGroup = groupName
			// 借用完畢後，切回原本「測速前」的節點，以防影響後續的流程
			originalTarget = groupNowMap[groupName]
		}

		isExploration := (bwTestCandidate != originalTarget)

		if isExploration {
			logGroup(groupName, colorInfo.Sprintf("💡 觸發反壟斷探索機制：切換群組 [%s] 至潛力節點 %s 進行測速", borrowGroup, formatNode(bwTestCandidate)))
			err := r.api.SelectProxy(borrowGroup, bwTestCandidate)
			if err != nil {
				logGroup(groupName, colorError.Sprintf("無法切換至候選節點: %v", err))
				continue
			}
		} else {
			logGroup(groupName, colorInfo.Sprintf("準備針對最佳節點 %s 進行真實頻寬測試...", formatNode(bwTestCandidate)))
		}

		r.lastBandwidthTest[bwTestCandidate] = time.Now()
		alreadyTestedInCycle[bwTestCandidate] = true
		
		speedKBps, totalBytes, err := r.api.TestBandwidth(r.cfg.BandwidthTestURL, r.cfg.ClashProxyURL, 15*time.Second)
		
		if err != nil {
			logGroup(groupName, colorError.Sprintf("頻寬測試失敗: %v", err))
			r.db.InsertLog(bwTestCandidate, 9999, false)
		} else {
			logGroup(groupName, colorSuccess.Sprintf("頻寬測試完成: %s KB/s", formatVal(fmt.Sprintf("%.2f", speedKBps))))
			r.db.InsertBandwidthLog(bwTestCandidate, speedKBps, totalBytes)
		}

		// 切回原本的冠軍節點
		if isExploration && originalTarget != "" {
			logGroup(groupName, colorMuted.Sprintf("探索測速完成，將群組 [%s] 切回節點 %s", borrowGroup, formatNode(originalTarget)))
			r.api.SelectProxy(borrowGroup, originalTarget)
		}
	}

	// 5. 統一執行所有主要群組的最終節點切換
	if len(pendingSwitches) > 0 {
		logHeader("執行主要群組節點切換")
		for groupName, targetNode := range pendingSwitches {
			if err := r.api.SelectProxy(groupName, targetNode); err != nil {
				logGroup(groupName, colorError.Sprintf("切換代理節點失敗: %v", err))
			} else {
				logGroup(groupName, colorSuccess.Sprintf("成功切換代理節點至 %s", formatNode(targetNode)))
			}
		}
	}
}

func (r *Rover) runFailoverWatchdog() {
	logInfo("啟動秒級急救機制 Watchdog (偵測間隔: %s 秒)", formatVal(r.cfg.FailoverInterval))
	ticker := time.NewTicker(time.Duration(r.cfg.FailoverInterval) * time.Second)
	defer ticker.Stop()

	consecutiveFails := make(map[string]int)

	for {
		select {
		case <-r.Quit:
			return
		case <-ticker.C:
			if r.IsRunning {
				// 如果主測速引擎正在執行中，不要干擾
				continue
			}

			// 針對每個群組檢查目前的節點
			for _, groupName := range r.cfg.TargetGroups {
				group, err := r.api.GetProxyGroup(groupName)
				if err != nil || group.Now == "" {
					continue
				}

				activeNode := group.Now
				testUrl := r.pickRandomURL()
				// 只等 3 秒的超時，要求快速回應
				_, err = r.api.TestProxyDelay(activeNode, testUrl, 3*time.Second)
				
				if err != nil {
					consecutiveFails[groupName]++
					if consecutiveFails[groupName] == 1 {
						logGroup(groupName, colorWarning.Sprintf("節點 %s 失去回應，進入黃色警戒...", formatNode(activeNode)))
					}
					
					if consecutiveFails[groupName] >= r.cfg.FailoverMaxFails {
						logFailover("[%s] 節點 %s 已癱瘓！觸發秒級急救！", groupName, activeNode)
						
						// 寫入懲罰
						r.db.InsertLog(activeNode, 9999, false)
						
						// 找備胎
						scores, _ := r.db.GetScores(r.cfg.HistoryDays)
						var bestAlt string
						var highestScore = -999999
						
						for _, candidate := range group.All {
							if candidate == activeNode {
								continue
							}
							if sc, ok := scores[candidate]; ok {
								if sc.Score > highestScore {
									highestScore = sc.Score
									bestAlt = candidate
								}
							}
						}
						
						if bestAlt != "" {
							r.api.SelectProxy(groupName, bestAlt)
							logFailover("[%s] 已強制切換至備用節點 %s", groupName, bestAlt)
						} else {
							logFailover("[%s] 找不到其他可用的備用節點！", groupName)
						}
						
						// 重置計數
						consecutiveFails[groupName] = 0
					}
				} else {
					// 成功回應，重置計數
					if consecutiveFails[groupName] > 0 {
						logGroup(groupName, colorSuccess.Sprintf("節點 %s 恢復連線，解除黃色警戒。", formatNode(activeNode)))
						consecutiveFails[groupName] = 0
					}
				}
			}
		}
	}
}

