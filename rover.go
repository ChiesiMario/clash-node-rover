package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
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
				r.checkBrowserTestURLsChanged()
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

func (r *Rover) checkBrowserTestURLsChanged() {
	currentURLs := strings.Join(r.GetConfig().BrowserTestURLs, ",")
	lastURLs, _ := r.db.GetMetadata("browser_test_urls")
	
	if lastURLs != "" && lastURLs != currentURLs {
		logWarning("偵測到 browser_test_urls 發生變更，已清空舊有網頁測試紀錄")
		r.db.ClearBrowserLogs()
	}
	r.db.SetMetadata("browser_test_urls", currentURLs)
}

func (r *Rover) Start() {
	logHeader("Clash Node Rover 啟動")

	// 啟動設定檔監控
	go r.watchConfig()

	// 檢查網頁測試目標是否變更
	r.checkBrowserTestURLsChanged()

	// 啟動時先執行一次資料庫瘦身
	if err := r.db.Cleanup(r.GetConfig().CleanupDays); err != nil {
		logError("資料庫自動瘦身失敗: %v", err)
	} else {
		logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.GetConfig().CleanupDays))
	}

	r.runCheckCycle(false) // 啟動時先跑一次

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
			r.runCheckCycle(false)
		case <-cleanupTicker.C:
			if err := r.db.Cleanup(r.GetConfig().CleanupDays); err != nil {
				logError("資料庫自動瘦身失敗: %v", err)
			} else {
				logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.GetConfig().CleanupDays))
			}
		case <-r.ManualTrigger:
			logInfo("收到手動測速信號，立即執行！")
			ticker.Stop()
			r.runCheckCycle(true)
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

func (r *Rover) runCheckCycle(isManual bool) {
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

		var urlSample string
		if len(r.GetConfig().TestURLs) > 0 {
			urlSample = r.GetConfig().TestURLs[0]
		}
		logInfo("開始並發 Ping 測試 %s 個節點 (目標: %s..., 超時: %v)...", formatVal(len(nodesToTest)), urlSample, r.GetConfig().TestTimeout)

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

		successCount := 0
		failCount := 0
		var successfulStats []nodeStat
		for s := range stats {
			statResultsMap[s.Name] = s
			success := (s.Err == nil && s.Delay > 0)
			if success {
				r.failedConsec[s.Name] = 0
				successCount++
				successfulStats = append(successfulStats, s)
			} else {
				r.failedConsec[s.Name]++
				failCount++
				if s.Err != nil {
					logMuted("  - ❌ [失敗] %s: %v", formatNode(s.Name), s.Err)
				} else {
					logMuted("  - ❌ [失敗] %s: 延遲為 0 或未知錯誤", formatNode(s.Name))
				}
			}
			r.db.InsertLog(s.Name, s.Delay, success)
		}
		logSuccess("Ping 測試完成！有效節點: %s, 失敗/超時: %s", formatVal(successCount), formatVal(failCount))

		sort.Slice(successfulStats, func(i, j int) bool {
			return successfulStats[i].Delay < successfulStats[j].Delay
		})
		for i, s := range successfulStats {
			if i < 5 {
				logMuted("  - ✅ [成功] %s: %d ms", formatNode(s.Name), s.Delay)
			}
		}
		if len(successfulStats) > 5 {
			logMuted("  - ... 及其他 %d 個成功節點", len(successfulStats)-5)
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

	// 3. 重新讀取 Clash 真實狀態（防止在 Ping 期間被 Failover 或手動切換）
	for _, groupName := range r.GetConfig().TargetGroups {
		freshGroup, err := r.api.GetProxyGroup(groupName)
		if err == nil && freshGroup.Now != groupNowMap[groupName] {
			logWarning("偵測到群組 [%s] 在 Ping 測試期間發生節點變更：%s → %s",
				groupName, formatNode(groupNowMap[groupName]), formatNode(freshGroup.Now))
			groupNowMap[groupName] = freshGroup.Now
		}
	}

	// 4. 為每個群組獨立計算最佳節點
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

		if highestScoreNode != "" && highestScoreNode != fastestNode {
			diff := highestScoreCurrentDelay - fastestDelay
			if diff <= r.GetConfig().DelayTolerance {
				targetNode = highestScoreNode
				reason = colorSuccess.Sprintf("質量分最高，且與最快差距僅 %dms", diff)
			} else {
				reason = colorWarning.Sprintf("高分節點比最快慢 %dms (超過容忍度)，因此退而求其次選最快", diff)
			}
		}

		groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("⚖️ 決策邏輯：%s", reason))
		groupTargetNodes[groupName] = targetNode
		
		currentNow := groupNowMap[groupName]
		if targetNode != currentNow && currentNow != "" {
			groupReports[groupName] = append(groupReports[groupName], colorInfo.Sprintf("🔌 狀態：預計從 %s 切換至 %s", formatNode(currentNow), formatNode(targetNode)))
			systemReports = append(systemReports, colorSuccess.Sprintf("✅ 成功：群組 [%s] 切換至 %s", groupName, formatNode(targetNode)))
			pendingSwitches[groupName] = targetNode
			
			if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnBetterNode {
				beeep.Notify("Clash Node Rover", fmt.Sprintf("群組 [%s] 更換較佳節點為 %s", groupName, targetNode), "")
			}
		} else {
			groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("🛡️ 狀態：目前的節點 %s 依然是最佳選擇，無需切換", formatNode(currentNow)))
		}
	}

	// 5. 反壟斷公平探索機制 (Monopoly Breaker) & 頻寬測速
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
		targetNode := groupTargetNodes[groupName]

		// 優先檢查冠軍節點 (目標節點) 是否已過常規冷卻期 (例如 5 分鐘)
		if targetNode != "" && (isManual || time.Since(r.lastBandwidthTest[targetNode]) >= targetIntervalDuration) {
			bwTestCandidate = targetNode
		} else {
			// 如果冠軍節點還在冷卻中，則把這次測速機會讓給「潛力股探索」 (冷卻期例如 60 分鐘)
			highestBaseScore := -999999
			for _, name := range nodes {
				sc, ok := scores[name]
				if !ok || sc.SuccessRate < 0.8 {
					continue
				}
				// 排除掉目前的目標節點 (因為上面已經判斷過它還在冷卻中)
				if name == targetNode {
					continue
				}
				if isManual || time.Since(r.lastBandwidthTest[name]) >= explorationDuration {
					if sc.BaseScore > highestBaseScore {
						highestBaseScore = sc.BaseScore
						bwTestCandidate = name
					}
				}
			}
		}

		if bwTestCandidate == "" {
			groupReports[groupName] = append(groupReports[groupName], colorMuted.Sprint("⏳ 測速：無 (所有優質節點皆在冷卻期內)"))
			continue
		}

		if alreadyTestedInCycle[bwTestCandidate] {
			groupReports[groupName] = append(groupReports[groupName], colorMuted.Sprint("⏭️ 測速：本週期已測速過，跳過重複測試"))
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
		
		if isExploration {
			logInfo("開始對潛力節點 %s 進行 15 秒極限頻寬測試 (目標 URL: %s)...", formatNode(bwTestCandidate), r.GetConfig().BandwidthTestURL)
		} else {
			logInfo("開始對目前冠軍節點 %s 進行常規頻寬測試 (目標 URL: %s)...", formatNode(bwTestCandidate), r.GetConfig().BandwidthTestURL)
		}

		speedKBps, totalBytes, err := r.api.TestBandwidth(r.GetConfig().BandwidthTestURL, r.GetConfig().ClashProxyURL, 15*time.Second)
		
		if err != nil {
			logWarning("頻寬測試失敗或超時: %s (錯誤: %v)", formatNode(bwTestCandidate), err)
			groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("💥 測速：頻寬測試失敗或超時 (原因: %v)", err))
			// V3: 頻寬測試失敗插入 3 筆懲罰，防止分數快速恢復
			for i := 0; i < 3; i++ {
				r.db.InsertLog(bwTestCandidate, 9999, false)
			}
		} else {
			mbps := (speedKBps / 1024.0)
			consumedMB := float64(totalBytes) / (1024.0 * 1024.0)
			tag := ""
			if isExploration {
				tag = colorInfo.Sprint(" (反壟斷探索)")
			}
			logSuccess("頻寬測試完成: %s 下載達 %s MB/s (共消耗 %s MB)", formatNode(bwTestCandidate), formatVal(fmt.Sprintf("%.2f", mbps)), formatVal(fmt.Sprintf("%.1f", consumedMB)))
			groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("📈 測速：下載達 %s MB/s%s", formatVal(fmt.Sprintf("%.2f", mbps)), tag))
			r.db.InsertBandwidthLog(bwTestCandidate, speedKBps, totalBytes)
		}

		// ----------------------------------------------------
		// 無頭瀏覽器網頁開啟測試
		// ----------------------------------------------------
		if r.GetConfig().EnableBrowserTest && len(r.GetConfig().BrowserTestURLs) > 0 {
			logInfo("開始使用無頭瀏覽器測試網頁連通性: %s (共 %d 個目標)...", formatNode(bwTestCandidate), len(r.GetConfig().BrowserTestURLs))
			
			for _, targetURL := range r.GetConfig().BrowserTestURLs {
				// 設定 Chromedp 的 Proxy 參數
				opts := append(chromedp.DefaultExecAllocatorOptions[:],
					chromedp.ProxyServer(r.GetConfig().ClashProxyURL),
				)
				
				allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
				ctx, cancelCtx := chromedp.NewContext(allocCtx)
				
				// 設定超時
				ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)
				
				startTime := time.Now()
				err := chromedp.Run(ctx,
					chromedp.Navigate(targetURL),
					chromedp.WaitReady("body", chromedp.ByQuery),
				)
				loadTimeMs := int(time.Since(startTime).Milliseconds())
				
				cancelTimeout()
				cancelCtx()
				cancelAlloc()
				
				if err != nil {
					logWarning("網頁開啟失敗: %s (目標: %s, 錯誤: %v)", formatNode(bwTestCandidate), targetURL, err)
					groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 網頁：開啟 %s 失敗", targetURL))
					r.db.InsertBrowserLog(bwTestCandidate, targetURL, false, loadTimeMs)
				} else {
					logSuccess("網頁成功開啟: %s (目標: %s, 耗時 %d ms)", formatNode(bwTestCandidate), targetURL, loadTimeMs)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 網頁：成功開啟 %s (耗時 %d ms)", targetURL, loadTimeMs))
					r.db.InsertBrowserLog(bwTestCandidate, targetURL, true, loadTimeMs)
				}
			}
		}

		// 切回原本的冠軍節點
		if isExploration && originalTarget != "" {
			r.api.SelectProxy(borrowGroup, originalTarget)
		}
	}

	// 6. 統一執行所有主要群組的最終節點切換
	successSwitches := 0
	failSwitches := 0
	if len(pendingSwitches) > 0 {
		for groupName, targetNode := range pendingSwitches {
			if err := r.api.SelectProxy(groupName, targetNode); err != nil {
				failSwitches++
				logError("群組 [%s] 切換至 %s 失敗: %v", groupName, formatNode(targetNode), err)
			} else {
				// 驗證切換是否真的生效
				verifyGroup, verifyErr := r.api.GetProxyGroup(groupName)
				if verifyErr == nil && verifyGroup.Now != targetNode {
					logError("⚠️ 切換驗證失敗：群組 [%s] 預期 %s 但 Clash 實際為 %s",
						groupName, formatNode(targetNode), formatNode(verifyGroup.Now))
					failSwitches++
				} else {
					successSwitches++
				}
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
						
						// V3: 寫入 5 筆懲罰紀錄，使懲罰效果不會被快速稀釋
						for i := 0; i < 5; i++ {
							r.db.InsertLog(activeNode, 9999, false)
						}
						
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

