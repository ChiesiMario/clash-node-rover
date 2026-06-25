package main

import (
	"regexp"

	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	backoffRemaining  map[string]int

	browserFailedConsec      map[string]map[string]int
	browserBackoffRemaining  map[string]map[string]int
	lastCheckTime     map[string]time.Time
	lastBandwidthTest map[string]time.Time
	lastInterviewTime    map[string]time.Time
	lastNetworkCheckTime time.Time
	isNetworkUp          bool
	stateMutex           sync.RWMutex

	// 進階功能控制
	ManualTrigger chan struct{}
	Quit            chan struct{}
	IsRunning       atomic.Bool
	activeBorrowing atomic.Bool
	IsPaused        bool
	pauseMutex      sync.RWMutex
	lockedGroups  map[string]bool
	lockedMutex   sync.RWMutex

	statResultsMap map[string]nodeStat
	statsMutex     sync.RWMutex

	startTime         time.Time
}

func NewRover(cfg *Config, api *APIClient, db *DB) *Rover {
	r := &Rover{
		startTime:         time.Now(),
		cfg:               cfg,
		api:               api,
		db:                db,
		failedConsec:      make(map[string]int),
		backoffRemaining:  make(map[string]int),
		browserFailedConsec:     make(map[string]map[string]int),
		browserBackoffRemaining: make(map[string]map[string]int),
		lastCheckTime:     make(map[string]time.Time),
		lastBandwidthTest: make(map[string]time.Time),
		lastInterviewTime: make(map[string]time.Time),
		ManualTrigger:     make(chan struct{}, 1),
		Quit:              make(chan struct{}, 1),
		IsPaused:          false,
		lockedGroups:      make(map[string]bool),
		isNetworkUp:       true, // 預設當作有網路
		statResultsMap:    make(map[string]nodeStat),
	}
	r.loadState()
	return r
}

func (r *Rover) TogglePause() bool {
	r.pauseMutex.Lock()
	defer r.pauseMutex.Unlock()
	r.IsPaused = !r.IsPaused
	if r.IsPaused {
		logWarning("⏸️ 系統已手動暫停，停止自動測速與切換")
	} else {
		logSuccess("▶️ 系統已手動恢復，繼續自動測速與切換")
	}
	return r.IsPaused
}

func (r *Rover) GetIsPaused() bool {
	r.pauseMutex.RLock()
	defer r.pauseMutex.RUnlock()
	return r.IsPaused
}

func (r *Rover) IsGroupLocked(groupName string) bool {
	r.lockedMutex.RLock()
	defer r.lockedMutex.RUnlock()
	return r.lockedGroups[groupName]
}

func (r *Rover) SetGroupLocked(groupName string, locked bool) {
	r.lockedMutex.Lock()
	r.lockedGroups[groupName] = locked
	r.lockedMutex.Unlock()
	r.saveState()
	
	if locked {
		logWarning("🔒 群組 [%s] 已被鎖定，將暫停自動切換與急救機制", groupName)
	} else {
		logSuccess("🔓 群組 [%s] 已解鎖，恢復自動切換", groupName)
	}
}

func (r *Rover) checkGlobalNetwork() bool {
	r.stateMutex.RLock()
	lastCheck := r.lastNetworkCheckTime
	isUp := r.isNetworkUp
	r.stateMutex.RUnlock()

	if time.Since(lastCheck) < 5*time.Second {
		return isUp
	}

	endpoints := []string{"1.1.1.1:53", "8.8.8.8:53", "223.5.5.5:53", "114.114.114.114:53"}
	success := false

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, ep := range endpoints {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", address, 2*time.Second)
			if err == nil {
				conn.Close()
				mu.Lock()
				success = true
				mu.Unlock()
			}
		}(ep)
	}

	wg.Wait()

	r.stateMutex.Lock()
	r.lastNetworkCheckTime = time.Now()
	r.isNetworkUp = success
	r.stateMutex.Unlock()

	return success
}

func (r *Rover) loadState() {
	if bwStr, _ := r.db.GetMetadata("last_bandwidth_test"); bwStr != "" {
		json.Unmarshal([]byte(bwStr), &r.lastBandwidthTest)
	}
	if intStr, _ := r.db.GetMetadata("last_interview_time"); intStr != "" {
		json.Unmarshal([]byte(intStr), &r.lastInterviewTime)
	}
	if lockStr, _ := r.db.GetMetadata("locked_groups"); lockStr != "" {
		json.Unmarshal([]byte(lockStr), &r.lockedGroups)
	}
}

func (r *Rover) saveState() {
	r.stateMutex.RLock()
	bwJson, errBw := json.Marshal(r.lastBandwidthTest)
	intJson, errInt := json.Marshal(r.lastInterviewTime)
	r.stateMutex.RUnlock()

	if errBw == nil {
		r.db.SetMetadata("last_bandwidth_test", string(bwJson))
	}
	if errInt == nil {
		r.db.SetMetadata("last_interview_time", string(intJson))
	}
	r.lockedMutex.RLock()
	if lockJson, err := json.Marshal(r.lockedGroups); err == nil {
		r.db.SetMetadata("locked_groups", string(lockJson))
	}
	r.lockedMutex.RUnlock()
}

func (r *Rover) GetConfig() *Config {
	r.cfgMutex.RLock()
	defer r.cfgMutex.RUnlock()
	return r.cfg
}

func (r *Rover) GetAPI() *APIClient {
	return r.api
}

func (r *Rover) GetLastInterviewTime(node string) time.Time {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.lastInterviewTime[node]
}

func (r *Rover) GetBackoffRemaining(node string) int {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.backoffRemaining[node]
}

func (r *Rover) GetBrowserBackoffRemaining(name string) map[string]int {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	
	original, ok := r.browserBackoffRemaining[name]
	if !ok || original == nil {
		return nil
	}
	
	res := make(map[string]int, len(original))
	for k, v := range original {
		res[k] = v
	}
	return res
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

func (r *Rover) checkClashStatus() bool {

	for _, groupName := range r.GetConfig().TargetGroups {
		_, err := r.api.GetProxyGroup(groupName)
		if err != nil {
			logWarning("群組 [%s] 不存在或查詢失敗: %v", groupName, err)
			return false
		}
	}

	if r.GetConfig().DedicatedTestGroup != "" {
		_, err := r.api.GetProxyGroup(r.GetConfig().DedicatedTestGroup)
		if err != nil {
			logWarning("獨立測速群組 [%s] 不存在或查詢失敗: %v", r.GetConfig().DedicatedTestGroup, err)
			return false
		}
	}

	return true
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

	// 每天執行一次資料庫瘦身
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		// 外層迴圈: 檢查先決條件 (Recovery Loop)
		for !r.checkClashStatus() {
			logWarning("偵測到 Clash API 異常或群組設定錯誤，進入 60 秒等待退避模式...")
			select {
			case <-r.Quit:
				logWarning("背景測速引擎已停止。")
				return
			case <-time.After(60 * time.Second):
				// retry
			}
		}

		logSuccess("Clash API 已連線且群組設定無誤，開始正常執行。")

		// 內層迴圈: 正常任務執行 (Active Loop)
		watchdogCtx, watchdogCancel := context.WithCancel(context.Background())

		apiBroken := make(chan struct{}, 1)

		// 啟動背景健康監控，每 3 秒檢查一次，確保能在中途失聯時立刻發現
		go func() {
			healthTicker := time.NewTicker(3 * time.Second)
			defer healthTicker.Stop()
			for {
				select {
				case <-watchdogCtx.Done():
					return
				case <-healthTicker.C:
					// 進行靜默檢查（若在此期間失聯則發送信號）
					if !r.checkClashStatus() {
						select {
						case apiBroken <- struct{}{}:
						default:
						}
						return
					}
				}
			}
		}()

		// 進入正常模式時先跑一次
		r.runCheckCycle(false)

		ticker := time.NewTicker(r.GetConfig().CheckInterval)
		broken := false

		for !broken {
			select {
			case <-apiBroken:
				broken = true
				break
			case <-r.Quit:
				watchdogCancel()
				ticker.Stop()
				logWarning("背景測速引擎已停止。")
				return
			case <-ticker.C:
				if !r.checkClashStatus() {
					broken = true
					break
				}
				r.runCheckCycle(false)
			case <-cleanupTicker.C:
				if err := r.db.Cleanup(r.GetConfig().CleanupDays); err != nil {
					logError("資料庫自動瘦身失敗: %v", err)
				} else {
					logSuccess("資料庫自動瘦身完成 (保留 %s 天)", formatVal(r.GetConfig().CleanupDays))
				}
			case <-r.ManualTrigger:
				logInfo("收到手動測速信號，立即執行！")
				if !r.checkClashStatus() {
					broken = true
					break
				}
				ticker.Stop()
				r.runCheckCycle(true)
				ticker.Reset(r.GetConfig().CheckInterval)
			}
		}

		ticker.Stop()
		watchdogCancel()
		logWarning("中斷正常執行週期，準備重新進行 60 秒退避檢查...")
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
	Name     string
	AvgDelay int
	Jitter   int
	Score    int
	Err      error
}

func (r *Rover) GetStatResults() map[string]nodeStat {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()
	res := make(map[string]nodeStat)
	for k, v := range r.statResultsMap {
		res[k] = v
	}
	return res
}

func (r *Rover) runCheckCycle(isManual bool) {
	if r.GetIsPaused() {
		if isManual {
			logWarning("系統目前處於暫停狀態，無法執行測速。")
		}
		return
	}

	if !r.IsRunning.CompareAndSwap(false, true) {
		return
	}
	defer func() {
		r.saveState()
		r.IsRunning.Store(false)
		r.activeBorrowing.Store(false)
		logReportEnd()
		BroadcastRefresh()
	}()

	groupNodesMap := make(map[string][]string)
	groupNowMap := make(map[string]string)
	uniqueNodes := make(map[string]bool)

	providers, err := r.api.GetProxyProviders()
	if err == nil {
		for pName, p := range providers {
			for _, proxy := range p.Proxies {
				SetNodeProvider(proxy.Name, pName)
			}
		}
	} else {
		logWarning("無法取得 Provider 資訊: %v，本次測速作廢。", err)
		return
	}

	for _, groupName := range r.GetConfig().TargetGroups {
		group, err := r.api.GetProxyGroup(groupName)
		if err != nil {
			continue
		}
		if len(group.All) == 0 {
			continue
		}
		
		filter := r.getGroupFilter(groupName)
		var filteredNodes []string
		var rx *regexp.Regexp
		if filter.KeywordRegex != "" {
			rx, _ = regexp.Compile("(?i)" + filter.KeywordRegex)
		}
		for _, n := range group.All {
			if rx != nil && !rx.MatchString(n) {
				continue
			}
			filteredNodes = append(filteredNodes, n)
		}
		
		if len(filteredNodes) == 0 && len(group.All) > 0 {
			logWarning("群組 [%s] 的節點被過濾規則全部排除了，退回使用全部節點。", groupName)
			filteredNodes = group.All
		}

		groupNodesMap[groupName] = filteredNodes
		groupNowMap[groupName] = group.Now
		for _, name := range group.All {
			uniqueNodes[name] = true
		}
	}

	if len(uniqueNodes) == 0 {
		return
	}

	var nodesToTest []string
	var backedOffNodes []string
	totalBackoff := 0

	r.stateMutex.Lock()
	for name := range uniqueNodes {
		if r.backoffRemaining[name] > 0 {
			backedOffNodes = append(backedOffNodes, name)
			totalBackoff++
			continue
		}
		nodesToTest = append(nodesToTest, name)
	}
	r.stateMutex.Unlock()

	if len(nodesToTest) == 0 || float64(totalBackoff) >= float64(len(uniqueNodes))*0.8 {
		if len(uniqueNodes) > 0 {
			logWarning("偵測到大規模節點癱瘓 (退避比例過高)，強制解除所有節點退避狀態以尋找可用節點！")
			r.stateMutex.Lock()
			for name := range uniqueNodes {
				r.backoffRemaining[name] = 0
				r.failedConsec[name] = 0
			}
			r.stateMutex.Unlock()

			nodesToTest = nil
			backedOffNodes = nil
			totalBackoff = 0
			for name := range uniqueNodes {
				nodesToTest = append(nodesToTest, name)
			}
		}
	}

	defer func() {
		r.stateMutex.Lock()
		for _, name := range backedOffNodes {
			if r.backoffRemaining[name] > 0 {
				r.backoffRemaining[name]--
			}
		}
		r.stateMutex.Unlock()
	}()

	if len(nodesToTest) > 0 {
		stats := make(chan nodeStat, len(nodesToTest))
		jobs := make(chan string, len(nodesToTest))

		var wg sync.WaitGroup
		workerCount := r.GetConfig().MaxConcurrent
		if workerCount > len(nodesToTest) {
			workerCount = len(nodesToTest)
		}

		logInfo("開始並發 Ping 測試 %s 個節點 (每個測試 5 次)...", formatVal(len(nodesToTest)))

		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for name := range jobs {
					var delays []int
					var lastErr error
					for i := 0; i < 5; i++ {
						testUrl := r.pickRandomURL()
						delay, err := r.api.TestProxyDelay(name, testUrl, r.GetConfig().TestTimeout)
						if err == nil && delay > 0 {
							delays = append(delays, delay)
						} else {
							lastErr = err
							break // Short-circuit: if one ping fails, stop testing this node
						}
					}
					
					var avgDelay, jitter float64
					if len(delays) > 0 {
						sum := 0
						for _, d := range delays {
							sum += d
						}
						avgDelay = float64(sum) / float64(len(delays))
						
						if len(delays) > 1 {
							var varSum float64
							for _, d := range delays {
								diff := float64(d) - avgDelay
								varSum += diff * diff
							}
							jitter = math.Sqrt(varSum / float64(len(delays)))
						}
					}
					
					if len(delays) < 5 {
						if lastErr == nil {
							lastErr = fmt.Errorf("incomplete pings (%d/5 successful)", len(delays))
						}
						logWarning("節點測速 [%s] 失敗: %v", name, lastErr)
						stats <- nodeStat{Name: name, Err: lastErr}
					} else {
						logInfo("節點測速 [%s] 成功: 延遲 %dms, 抖動 %dms", name, int(avgDelay), int(jitter))
						stats <- nodeStat{
							Name:     name,
							AvgDelay: int(avgDelay),
							Jitter:   int(jitter),
							Score:    int(avgDelay) + int(jitter),
						}
					}
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

		var collectedStats []nodeStat
		for s := range stats {
			collectedStats = append(collectedStats, s)
		}

		if !r.checkClashStatus() {
			logError("🚨 Ping 測試期間偵測到 Clash API 失聯，本次測速結果作廢！")
			return
		}

		if !r.checkGlobalNetwork() {
			logError("🚨 實體網路斷線！為避免無差別誤殺節點，本次測速結果與懲罰全數作廢！")
			return
		}

		successCount := 0
		failCount := 0
		var successfulStats []nodeStat
		
		r.statsMutex.Lock()
		for _, s := range collectedStats {
			r.statResultsMap[s.Name] = s
			success := (s.Err == nil)
			
			r.stateMutex.Lock()
			if success {
				r.failedConsec[s.Name] = 0
				r.backoffRemaining[s.Name] = 0
			} else {
				r.failedConsec[s.Name]++
				fails := r.failedConsec[s.Name]
				skipCycles := int(math.Pow(2, float64(fails-1)))
				if skipCycles > r.GetConfig().MaxBackoffCycles {
					skipCycles = r.GetConfig().MaxBackoffCycles
				}
				r.backoffRemaining[s.Name] = skipCycles
			}
			r.stateMutex.Unlock()

			if success {
				successCount++
				successfulStats = append(successfulStats, s)
				r.db.InsertLog(s.Name, s.AvgDelay, true)
			} else {
				failCount++
				if s.Err != nil {
					logMuted("  - ❌ [失敗] %s: %v", formatNode(s.Name), s.Err)
				}
				r.db.InsertLog(s.Name, 9999, false)
			}
		}
		r.statsMutex.Unlock()
		logSuccess(" ICMP 測速環節完成！有效節點: %s, 失敗/超時: %s", formatVal(successCount), formatVal(failCount))

		sort.Slice(successfulStats, func(i, j int) bool {
			return successfulStats[i].Score < successfulStats[j].Score
		})
		for i, s := range successfulStats {
			if i < 5 {
				logMuted("  - ✅ [排名 %d] %s: 延遲 %d ms, 抖動 %d ms, 綜合 %d", i+1, formatNode(s.Name), s.AvgDelay, s.Jitter, s.Score)
			}
		}
		if len(successfulStats) > 5 {
			logMuted("  - ... 及其他 %d 個成功節點", len(successfulStats)-5)
		}

		// Short-Circuit Selection Phase
		groupReports := make(map[string][]string)
		var systemReports []string
		systemReports = append(systemReports, fmt.Sprintf("退避：共 %s 個節點連線失敗，目前處於退避期", formatVal(totalBackoff)))
		systemReports = append(systemReports, fmt.Sprintf("測速：本次共 Ping 測試 %s 個節點", formatVal(len(nodesToTest))))

		var browserAllocCtx context.Context
		var browserAllocCancel context.CancelFunc

		if r.GetConfig().EnableBrowserTest && len(r.GetConfig().BrowserTestURLs) > 0 {
			opts := append(chromedp.DefaultExecAllocatorOptions[:],
				chromedp.ProxyServer(r.GetConfig().ClashProxyURL),
				chromedp.Flag("disable-cache", true),
				chromedp.Flag("incognito", true),
			)
			browserAllocCtx, browserAllocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
			defer browserAllocCancel()
		}

		urlTestCache := make(map[string]map[string]bool)
		finalSwitches := make(map[string]string)

		for _, groupName := range r.GetConfig().TargetGroups {
			nodes, ok := groupNodesMap[groupName]
			if !ok || len(nodes) == 0 {
				continue
			}

			currentNow := groupNowMap[groupName]
			logGroupTitle(groupName)

			var groupCandidates []nodeStat
			for _, stat := range successfulStats {
				for _, n := range nodes {
					if n == stat.Name {
						groupCandidates = append(groupCandidates, stat)
						break
					}
				}
			}

			if len(groupCandidates) == 0 {
				groupReports[groupName] = append(groupReports[groupName], colorError.Sprint("🚫 此群組所有節點皆 Ping 失敗，維持原狀"))
				continue
			}

			var currentStat nodeStat
			hasCurrent := false
			for _, stat := range groupCandidates {
				if stat.Name == currentNow {
					currentStat = stat
					hasCurrent = true
					break
				}
			}

			bestNode := groupCandidates[0]
			
			if hasCurrent && bestNode.Name != currentNow {
				if currentStat.Score - bestNode.Score <= r.GetConfig().ToleranceMs {
					groupReports[groupName] = append(groupReports[groupName], colorInfo.Sprintf("⚖️ 目前節點 %s (%d ms) 與最佳節點 %s (%d ms) 差距在容忍度內，優先使用目前節點", formatNode(currentNow), currentStat.Score, formatNode(bestNode.Name), bestNode.Score))
					var newCandidates []nodeStat
					newCandidates = append(newCandidates, currentStat)
					for _, stat := range groupCandidates {
						if stat.Name != currentNow {
							newCandidates = append(newCandidates, stat)
						}
					}
					groupCandidates = newCandidates
				}
			}

			var targetNode string
			var targetReason string

			for _, stat := range groupCandidates {
				candidate := stat.Name
				
				if !r.GetConfig().EnableBrowserTest {
					targetNode = candidate
					targetReason = fmt.Sprintf("未開啟網頁測試，直接選用綜合分數最佳 (%d ms)", stat.Score)
					break
				}



				filter := r.getGroupFilter(groupName)
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)
				if filter.CheckChatGPT {
					testURLs = append(testURLs, "https://chatgpt.com")
				}
				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}
				if filter.CheckAntigravity {
					testURLs = append(testURLs, "https://generativelanguage.googleapis.com/v1beta/models")
				}

				if urlTestCache[candidate] == nil {
					urlTestCache[candidate] = make(map[string]bool)
				}

				// 嘗試從資料庫載入持久化快取 (避免反覆切換節點進行耗時網頁測試)
				for _, targetURL := range testURLs {
					if _, exists := urlTestCache[candidate][targetURL]; !exists {
						lastSuccess, err := r.db.GetLastBrowserSuccessTime(candidate, targetURL)
						if err == nil && !lastSuccess.IsZero() {
							if time.Since(lastSuccess) < r.GetConfig().BrowserCacheDuration {
								urlTestCache[candidate][targetURL] = true
								r.stateMutex.Lock()
								if r.browserBackoffRemaining[candidate] == nil {
									r.browserBackoffRemaining[candidate] = make(map[string]int)
								}
								r.browserBackoffRemaining[candidate][targetURL] = 0
								r.stateMutex.Unlock()
							}
						}
					}
				}

				allCachedAndSuccess := true
				anyCachedFailed := false
				var failedURL string

				r.stateMutex.Lock()
				for _, targetURL := range testURLs {
					if urlMap, ok := r.browserBackoffRemaining[candidate]; ok {
						if rem := urlMap[targetURL]; rem > 0 {
							anyCachedFailed = true
							failedURL = targetURL
							break
						}
					}
				}
				r.stateMutex.Unlock()

				if !anyCachedFailed {
					for _, targetURL := range testURLs {
						if success, exists := urlTestCache[candidate][targetURL]; exists {
							if !success {
								anyCachedFailed = true
								failedURL = targetURL
								break
							}
						} else {
							allCachedAndSuccess = false
						}
					}
				} else {
					allCachedAndSuccess = false
				}

				if anyCachedFailed {
					groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 驗證失敗：部分服務在快取或退避中已失敗 (%s)，淘汰並順延: %s", failedURL, formatNode(candidate)))
					continue
				}

				if allCachedAndSuccess {
					targetNode = candidate
					targetReason = fmt.Sprintf("快取命中：綜合分數 (%d ms) 且各項服務驗證皆在有效期內", stat.Score)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：快取命中 %s 支援所有必要服務", formatNode(candidate)))
					break
				}

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

				if borrowGroup == groupName {
					r.activeBorrowing.Store(true)
				}
				
				if candidate != originalTarget {
					err := r.api.SelectProxy(borrowGroup, candidate)
					if err != nil {
						if !r.checkClashStatus() {
							logError("🚨 準備切換至測試節點時偵測到 Clash API 失聯！")
							return
						}
						groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("測速：無法切換至測試節點 %s", formatNode(candidate)))
						if borrowGroup == groupName {
							r.activeBorrowing.Store(false)
						}
						continue
					}
				}

				logInfo("使用無頭瀏覽器測試: %s...", formatNode(candidate))
				allSuccess := true
				var totalLoadTime int
				for _, targetURL := range testURLs {

					var checkingWhat string
					if strings.Contains(targetURL, "chatgpt.com") {
						checkingWhat = "ChatGPT"
					} else if strings.Contains(targetURL, "gemini.google.com") {
						checkingWhat = "Gemini"
					} else if strings.Contains(targetURL, "generativelanguage.googleapis") {
						checkingWhat = "Antigravity"
					} else {
						checkingWhat = "基礎連線 (" + targetURL + ")"
					}

					if success, exists := urlTestCache[candidate][targetURL]; exists {
						if success {
							logInfo("  ➤ 檢查 %s 中... (%s) [略過: 使用成功快取]", formatNode(candidate), checkingWhat)
							continue
						}
					}

					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)
					ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)

					logInfo("  ➤ 檢查 %s 中... (%s)", formatNode(candidate), checkingWhat)

					var innerText string
					startTime := time.Now()
					err := chromedp.Run(ctx,
						chromedp.Navigate(targetURL),
						chromedp.WaitReady("body", chromedp.ByQuery),
						chromedp.Evaluate(`document.body.innerText`, &innerText),
					)
					loadTimeMs := int(time.Since(startTime).Milliseconds())

					cancelTimeout()
					cancelCtx()

					if err != nil {
						logWarning("  ❌ 服務連線超時或失敗: %s (%v)", checkingWhat, err)
						urlTestCache[candidate][targetURL] = false
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						r.stateMutex.Lock()
						if r.browserFailedConsec[candidate] == nil {
							r.browserFailedConsec[candidate] = make(map[string]int)
						}
						if r.browserBackoffRemaining[candidate] == nil {
							r.browserBackoffRemaining[candidate] = make(map[string]int)
						}
						r.browserFailedConsec[candidate][targetURL]++
						fails := r.browserFailedConsec[candidate][targetURL]
						skipCycles := int(math.Pow(2, float64(fails-1)))
						if skipCycles > r.GetConfig().MaxBackoffCycles {
							skipCycles = r.GetConfig().MaxBackoffCycles
						}
						r.browserBackoffRemaining[candidate][targetURL] = skipCycles
						r.stateMutex.Unlock()
						break 
					} else {
						// 檢查是否有被地區阻擋
						lowerText := strings.ToLower(innerText)
						blocked := false
						if strings.Contains(targetURL, "chatgpt.com") {
							if strings.Contains(lowerText, "access denied") || strings.Contains(lowerText, "not available in your country") {
								blocked = true
							}
						} else if strings.Contains(targetURL, "gemini.google.com") {
							if strings.Contains(lowerText, "isn't supported in your country") || strings.Contains(lowerText, "未在該地區推出") || strings.Contains(lowerText, "not available") {
								blocked = true
							}
						} else if strings.Contains(targetURL, "generativelanguage.googleapis") {
							if strings.Contains(lowerText, "user location is not supported") {
								blocked = true
							}
						}

						if blocked {
							logWarning("  ❌ 服務驗證失敗 (地區限制或封鎖): %s", checkingWhat)
							urlTestCache[candidate][targetURL] = false
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
						r.stateMutex.Lock()
						if r.browserFailedConsec[candidate] == nil {
							r.browserFailedConsec[candidate] = make(map[string]int)
						}
						if r.browserBackoffRemaining[candidate] == nil {
							r.browserBackoffRemaining[candidate] = make(map[string]int)
						}
						r.browserFailedConsec[candidate][targetURL]++
						fails := r.browserFailedConsec[candidate][targetURL]
						skipCycles := int(math.Pow(2, float64(fails-1)))
						if skipCycles > r.GetConfig().MaxBackoffCycles {
							skipCycles = r.GetConfig().MaxBackoffCycles
						}
						r.browserBackoffRemaining[candidate][targetURL] = skipCycles
						r.stateMutex.Unlock()
							break
						}

						logSuccess("  ✅ 服務驗證通過: %s (%d ms)", checkingWhat, loadTimeMs)
						urlTestCache[candidate][targetURL] = true
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
						r.stateMutex.Lock()
						if r.browserFailedConsec[candidate] != nil {
							r.browserFailedConsec[candidate][targetURL] = 0
						}
						if r.browserBackoffRemaining[candidate] != nil {
							r.browserBackoffRemaining[candidate][targetURL] = 0
						}
						r.stateMutex.Unlock()
					}
				}

				if candidate != originalTarget && originalTarget != "" {
					r.api.SelectProxy(borrowGroup, originalTarget)
				}
				if borrowGroup == groupName {
					r.activeBorrowing.Store(false)
				}

				
				if allSuccess {
					targetNode = candidate
					targetReason = fmt.Sprintf("綜合分數 (%d ms) 且各項服務驗證皆成功 (%d ms)", stat.Score, totalLoadTime)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：節點 %s 支援所有必要服務", formatNode(candidate)))
					break
				} else {
					groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 驗證失敗：節點 %s 未通過服務驗證，淘汰並順延", formatNode(candidate)))
				}
			}

			if targetNode == "" {
				groupReports[groupName] = append(groupReports[groupName], colorError.Sprint("🚫 所有候選節點網頁測試皆失敗，維持原狀"))
				continue
			}

			if targetNode != currentNow && currentNow != "" {
				if r.IsGroupLocked(groupName) {
					groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("🔒 最終決策：發現更佳節點 %s (%s)，但群組已鎖定，維持 %s", formatNode(targetNode), targetReason, formatNode(currentNow)))
				} else {
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("✅ 最終決策：切換至 %s (%s)", formatNode(targetNode), targetReason))
					systemReports = append(systemReports, colorSuccess.Sprintf("✅ 成功：群組 [%s] 切換至 %s", groupName, formatNode(targetNode)))
					finalSwitches[groupName] = targetNode
					
					if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnBetterNode {
						beeep.Notify("Clash Node Rover", fmt.Sprintf("群組 [%s] 更換較佳節點為 %s", groupName, targetNode), "")
					}
				}
			} else {
				groupReports[groupName] = append(groupReports[groupName], fmt.Sprintf("🛡️ 最終決策：維持現有節點 %s (%s)", formatNode(currentNow), targetReason))
			}
		}

		successSwitches := 0
		failSwitches := 0
		if len(finalSwitches) > 0 {
			for groupName, targetNode := range finalSwitches {
				if err := r.api.SelectProxy(groupName, targetNode); err != nil {
					failSwitches++
					logError("群組 [%s] 切換至 %s 失敗: %v", groupName, formatNode(targetNode), err)
				} else {
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

		logReportStart()
		for _, groupName := range r.GetConfig().TargetGroups {
			logGroupTitle(groupName)
			lines := groupReports[groupName]
			if len(lines) == 0 {
				logTreeItem(true, "無狀態更新")
			} else {
				for i, line := range lines {
					logTreeItem(i == len(lines)-1, "%s", line)
				}
			}
		}

		logGroupTitle("🔧 系統狀態")
		for i, line := range systemReports {
			logTreeItem(i == len(systemReports)-1, "%s", line)
		}
	}
}


type GroupFilter struct {
	KeywordRegex string `json:"keyword_regex"`
	CheckChatGPT     bool   `json:"check_chatgpt"`
	CheckGemini      bool   `json:"check_gemini"`
	CheckAntigravity bool   `json:"check_antigravity"`
}

func (r *Rover) getGroupFilter(groupName string) GroupFilter {
	var f GroupFilter
	val, err := r.db.GetMetadata("group_filter_" + groupName)
	if err == nil && val != "" {
		json.Unmarshal([]byte(val), &f)
	}
	return f
}
