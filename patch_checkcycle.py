import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

# We need to find the runCheckCycle function and replace it up to the start of runFailoverWatchdog
start_idx = code.find('func (r *Rover) runCheckCycle(isManual bool) {')
end_idx = code.find('func (r *Rover) runFailoverWatchdog(ctx context.Context) {')

if start_idx != -1 and end_idx != -1:
    new_checkcycle = """func (r *Rover) runCheckCycle(isManual bool) {
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
		groupNodesMap[groupName] = group.All
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
					
					if len(delays) == 0 {
						if lastErr == nil {
							lastErr = fmt.Errorf("all 5 pings failed")
						}
						stats <- nodeStat{Name: name, Err: lastErr}
					} else {
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
		for _, s := range collectedStats {
			statResultsMap[s.Name] = s
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
		logSuccess("Ping 測試完成！有效節點: %s, 失敗/超時: %s", formatVal(successCount), formatVal(failCount))

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

		browserTestCache := make(map[string]bool)
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

				if success, exists := browserTestCache[candidate]; exists {
					if success {
						targetNode = candidate
						targetReason = fmt.Sprintf("共用跨群組快取：綜合分數 (%d ms) 且網頁測試成功", stat.Score)
						groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 網頁：共用快取 %s 成功", formatNode(candidate)))
						break
					} else {
						groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 網頁：共用快取 %s 失敗，順延", formatNode(candidate)))
						continue
					}
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
				for _, targetURL := range r.GetConfig().BrowserTestURLs {
					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)
					ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)

					startTime := time.Now()
					err := chromedp.Run(ctx,
						chromedp.Navigate(targetURL),
						chromedp.WaitReady("body", chromedp.ByQuery),
					)
					loadTimeMs := int(time.Since(startTime).Milliseconds())

					cancelTimeout()
					cancelCtx()

					if err != nil {
						logWarning("網頁開啟失敗: %s (%v)", formatNode(candidate), err)
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						break 
					} else {
						logSuccess("網頁成功開啟: %s (%d ms)", formatNode(candidate), loadTimeMs)
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}
				}

				if candidate != originalTarget && originalTarget != "" {
					r.api.SelectProxy(borrowGroup, originalTarget)
				}
				if borrowGroup == groupName {
					r.activeBorrowing.Store(false)
				}

				browserTestCache[candidate] = allSuccess

				if allSuccess {
					targetNode = candidate
					targetReason = fmt.Sprintf("綜合分數 (%d ms) 且網頁測試成功 (%d ms)", stat.Score, totalLoadTime)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 網頁：測試 %s 成功", formatNode(candidate)))
					break
				} else {
					groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 網頁：測試 %s 失敗，順延", formatNode(candidate)))
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
"""

    code = code[:start_idx] + new_checkcycle + '\n\n' + code[end_idx:]

with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
