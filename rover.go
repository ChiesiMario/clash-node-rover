package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
)

type Rover struct {
	cfg               *Config
	cfgMutex          sync.RWMutex
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

func (r *Rover) GetConfig() *Config {
	r.cfgMutex.RLock()
	defer r.cfgMutex.RUnlock()
	return r.cfg
}

func (r *Rover) watchConfig() {
	var lastModTime time.Time
	info, err := os.Stat(ConfigFile)
	if err == nil {
		lastModTime = info.ModTime()
	}

	for {
		time.Sleep(2 * time.Second)
		info, err := os.Stat(ConfigFile)
		if err != nil {
			continue
		}
		if info.ModTime().After(lastModTime) {
			lastModTime = info.ModTime()
			newCfg, err := loadConfig()
			if err == nil {
				r.cfgMutex.Lock()
				r.cfg = newCfg
				r.cfgMutex.Unlock()
				logInfo("🔄 偵測到設定檔變更，已自動重新載入")
			} else {
				logError("載入新設定檔失敗: %v", err)
			}
		}
	}
}

func (r *Rover) pickRandomURL() string {
	if len(r.GetConfig().TestURLs) == 0 {
		return "http://www.gstatic.com/generate_204"
	}
	return r.GetConfig().TestURLs[rand.Intn(len(r.GetConfig().TestURLs))]
}

func (r *Rover) Start() {
	logHeader("Clash Node Rover 啟動")

	// 啟動設定檔監控
	go r.watchConfig()

	// 啟動時先執行一次資料庫瘦身
	if err := r.db.Cleanup(r.GetConfig().CleanupDays); err != nil {
		logError("資料庫自動瘦身失敗: %v", err)
	} else {
		logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.GetConfig().CleanupDays))
	}

	r.runCheckCycle() // 啟動時先跑一次

	if r.GetConfig().EnableFailover {
		go r.runFailoverWatchdog()
	}

	ticker := time.NewTicker(r.GetConfig().CheckInterval)
	defer ticker.Stop()

	// 每天執行一次資料庫瘦身
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			r.runCheckCycle()
		case <-cleanupTicker.C:
			if err := r.db.Cleanup(r.GetConfig().CleanupDays); err != nil {
				logError("資料庫自動瘦身失敗: %v", err)
			} else {
				logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.GetConfig().CleanupDays))
			}
		case <-r.ManualTrigger:
			logInfo("收到手動測速信號，立即執行！")
			ticker.Stop()
			r.runCheckCycle()
			ticker.Reset(r.GetConfig().CheckInterval)
		case <-r.Quit:
			logWarning("背景測速引擎已停止。")
			return
		}
	}
}

func (r *Rover) Stop() {
	r.Quit <- struct{}{}
}

func (r *Rover) ForceCheck() {
	select {
	case r.ManualTrigger <- struct{}{}:
	default:
	}
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
	defer func() { 
		r.IsRunning = false 
		logReportEnd()
		BroadcastRefresh()
	}()

	groupNodesMap := make(map[string][]string)
	groupNowMap := make(map[string]string)
	uniqueNodes := make(map[string]bool)

	// 更新全域 Provider 映射
	providers, err := r.api.GetProxyProviders()
	if err == nil {
		for pName, p := range providers {
			for _, proxy := range p.Proxies {
				SetNodeProvider(proxy.Name, pName)
			}
		}
	} else {
		logWarning("無法取得 Provider 資訊: %v", err)
	}

	for _, groupName := range r.GetConfig().TargetGroups {
		group, err := r.api.GetProxyGroup(groupName)
		if err != nil {
			continue
		}
		if len(group.All) == 0 {
			continue
		}
		groupNodesMap[groupName] = group.All
		groupNowMap[groupName] = group.Now
		for _, name := range group.All {
			uniqueNodes[name] = true
		}
	}

	if len(uniqueNodes) == 0 {
		return
	}

	now := time.Now()
	var nodesToTest []string
	totalBackoff := 0

	// 指數退避檢查
	for name := range uniqueNodes {
		fails := r.failedConsec[name]
		if fails > 0 {
			backoffMins := int(math.Pow(2, float64(fails-1)))
			if backoffMins > r.GetConfig().MaxBackoffMinutes {
				backoffMins = r.GetConfig().MaxBackoffMinutes
			}

			lastCheck := r.lastCheckTime[name]
			if now.Sub(lastCheck) < time.Duration(backoffMins)*time.Minute {
				totalBackoff++
				continue
			}
		}
		nodesToTest = append(nodesToTest, name)
	}

	statResultsMap := make(map[string]nodeStat)

	if len(nodesToTest) > 0 {
		stats := make(chan nodeStat, len(nodesToTest))
		jobs := make(chan string, len(nodesToTest))

		var wg sync.WaitGroup
		workerCount := r.GetConfig().MaxConcurrent
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
					delay, err := r.api.TestProxyDelay(name, testUrl, r.GetConfig().TestTimeout)
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

		for s := range stats {
			statResultsMap[s.Name] = s
			success := (s.Err == nil && s.Delay > 0)
			if success {
				r.failedConsec[s.Name] = 0
			} else {
				r.failedConsec[s.Name]++
			}
			r.db.InsertLog(s.Name, s.Delay, success)
		}
	}

	scores, _ := r.db.GetScores(r.GetConfig().HistoryDays)

	// ---------------------------
	// 報告資料收集
	// ---------------------------
	groupReports := make(map[string][]string)
	var systemReports []string
	systemReports = append(systemReports, fmt.Sprintf("退避：共 %s 個節點連線失敗，目前處於退避期", formatVal(totalBackoff)))
	systemReports = append(systemReports, fmt.Sprintf("測速：本次共 Ping 測試 %s 個節點", formatVal(len(nodesToTest))))

	// 3. 為每個群組獨立計算最佳節點
	groupTargetNodes := make(map[string]string)
	pendingSwitches := make(map[string]string)

	for _, groupName := range r.GetConfig().TargetGroups {
		nodes, ok := groupNodesMap[groupName]
		if !ok || len(nodes) == 0 {
			groupReports[groupName] = append(groupReports[groupName], colorWarning.Sprint("無法取得群組節點或群組為空"))
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
			groupReports[groupName] = append(groupReports[groupName], colorWarning.Sprint("決策：所有參與測試的節點皆失敗，保留目前節點"))
			groupTargetNodes[groupName] = groupNowMap[groupName]
			continue
		}

		targetNode := fastestNode
		reason := colorInfo.Sprint("當前速度最快")

		if highestScoreNode != "" {
			diff := highestScoreCurrentDelay - fastestDelay
			if diff <= r.GetConfig().DelayTolerance {
				targetNode = highestScoreNode
				reason = colorSuccess.Sprintf("最高質量分，且與最快差距僅 %dms", diff)
			} else {
				reason = colorWarning.Sprintf("最高分節點比最快慢 %dms，強制選最快", diff)
			}
		}

		groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("決策：選擇 %s (%s)", formatNode(targetNode), reason))
		groupTargetNodes[groupName] = targetNode
		
		currentNow := groupNowMap[groupName]
		if targetNode != currentNow && currentNow != "" {
			groupReports[groupName] = append(groupReports[groupName], colorInfo.Sprintf("狀態：預計從 %s 切換至 %s", formatNode(currentNow), formatNode(targetNode)))
			systemReports = append(systemReports, colorSuccess.Sprintf("切換：群組 [%s] 成功切換至 %s", groupName, formatNode(targetNode)))
			pendingSwitches[groupName] = targetNode
			
			if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnBetterNode {
				beeep.Notify("Clash Node Rover", fmt.Sprintf("群組 [%s] 更換較佳節點為 %s", groupName, targetNode), "")
			}
		} else {
			groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("狀態：目前的節點 %s 依然是最佳選擇", formatNode(targetNode)))
		}
	}

	// 4. 反壟斷公平探索機制 (Monopoly Breaker) & 頻寬測速
	targetIntervalDuration := time.Duration(r.GetConfig().BandwidthTestInterval) * time.Minute
	explorationDuration := time.Duration(r.GetConfig().ExplorationCooldown) * time.Minute

	scores, _ = r.db.GetScores(r.GetConfig().HistoryDays)

	alreadyTestedInCycle := make(map[string]bool)

	for _, groupName := range r.GetConfig().TargetGroups {
		nodes, ok := groupNodesMap[groupName]
		if !ok || len(nodes) == 0 {
			continue
		}

		var bwTestCandidate string
		highestBaseScore := -999999

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

		targetNode := groupTargetNodes[groupName]
		if bwTestCandidate == "" && targetNode != "" {
			if time.Since(r.lastBandwidthTest[targetNode]) >= targetIntervalDuration {
				bwTestCandidate = targetNode
			}
		}

		if bwTestCandidate == "" {
			groupReports[groupName] = append(groupReports[groupName], colorMuted.Sprint("測速：無 (所有優質節點皆在冷卻期內)"))
			continue
		}

		if alreadyTestedInCycle[bwTestCandidate] {
			groupReports[groupName] = append(groupReports[groupName], colorMuted.Sprint("測速：本週期已測速過，跳過"))
			continue
		}

		// 準備測速
		var borrowGroup string
		var originalTarget string

		if r.GetConfig().DedicatedTestGroup != "" {
			borrowGroup = r.GetConfig().DedicatedTestGroup
			if g, err := r.api.GetProxyGroup(borrowGroup); err == nil {
				originalTarget = g.Now
			}
		} else {
			borrowGroup = groupName
			originalTarget = groupNowMap[groupName]
		}

		isExploration := (bwTestCandidate != originalTarget)

		if isExploration {
			err := r.api.SelectProxy(borrowGroup, bwTestCandidate)
			if err != nil {
				groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("測速：無法切換至探索節點 %s", formatNode(bwTestCandidate)))
				continue
			}
		}

		r.lastBandwidthTest[bwTestCandidate] = time.Now()
		alreadyTestedInCycle[bwTestCandidate] = true
		
		speedKBps, totalBytes, err := r.api.TestBandwidth(r.GetConfig().BandwidthTestURL, r.GetConfig().ClashProxyURL, 15*time.Second)
		
		if err != nil {
			groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("測速：%s 頻寬測試失敗", formatNode(bwTestCandidate)))
			r.db.InsertLog(bwTestCandidate, 9999, false)
		} else {
			mbps := (speedKBps / 1024.0)
			tag := ""
			if isExploration {
				tag = colorInfo.Sprint(" (反壟斷探索)")
			}
			groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("測速：%s 下載達 %s MB/s%s", formatNode(bwTestCandidate), formatVal(fmt.Sprintf("%.2f", mbps)), tag))
			r.db.InsertBandwidthLog(bwTestCandidate, speedKBps, totalBytes)
		}

		// 切回原本的冠軍節點
		if isExploration && originalTarget != "" {
			r.api.SelectProxy(borrowGroup, originalTarget)
		}
	}

	// 5. 統一執行所有主要群組的最終節點切換
	successSwitches := 0
	failSwitches := 0
	if len(pendingSwitches) > 0 {
		for groupName, targetNode := range pendingSwitches {
			if err := r.api.SelectProxy(groupName, targetNode); err != nil {
				failSwitches++
			} else {
				successSwitches++
			}
		}
		systemReports = append(systemReports, fmt.Sprintf("切換：成功切換 %d 個群組, 失敗 %d 個", successSwitches, failSwitches))
	} else {
		systemReports = append(systemReports, "切換：無需切換，維持現狀")
	}

	// ---------------------------
	// 列印樹狀結構報告
	// ---------------------------
	logReportStart()
	for _, groupName := range r.GetConfig().TargetGroups {
		logGroupTitle(groupName)
		lines := groupReports[groupName]
		if len(lines) == 0 {
			logTreeItem(true, "無狀態更新")
		} else {
			for i, line := range lines {
				logTreeItem(i == len(lines)-1, line)
			}
		}
	}
	
	logGroupTitle("🔧 系統狀態")
	for i, line := range systemReports {
		logTreeItem(i == len(systemReports)-1, line)
	}
}

func (r *Rover) runFailoverWatchdog() {
	logInfo("啟動秒級急救機制 Watchdog (偵測間隔: %s 秒)", formatVal(r.GetConfig().FailoverInterval))
	ticker := time.NewTicker(time.Duration(r.GetConfig().FailoverInterval) * time.Second)
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
			for _, groupName := range r.GetConfig().TargetGroups {
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
					
					if consecutiveFails[groupName] >= r.GetConfig().FailoverMaxFails {
						logFailover("[%s] 節點 %s 已癱瘓！觸發秒級急救！", groupName, activeNode)
						
						// 寫入懲罰
						r.db.InsertLog(activeNode, 9999, false)
						
						// 找備胎
						scores, _ := r.db.GetScores(r.GetConfig().HistoryDays)
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
							logFailover("群組 [%s] 已觸發急救機制！預計切換至: %s", groupName, formatNode(bestAlt))
							
							if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnFailover {
								beeep.Notify("🚨 Rover 急救成功", fmt.Sprintf("已為您攔截斷線，群組 [%s] 切換至 %s", groupName, bestAlt), "")
							}
							
							BroadcastRefresh()
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

