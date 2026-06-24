import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

# Replace nodeStat
code = re.sub(
    r'type nodeStat struct \{\s*Name\s*string\s*Delay\s*int\s*Err\s*error\s*\}',
    r'type nodeStat struct {\n\tName     string\n\tAvgDelay int\n\tJitter   int\n\tScore    int\n\tErr      error\n}',
    code
)

# Fix Watchdog completely
watchdog_new = """func (r *Rover) runFailoverWatchdog(ctx context.Context) {
	logInfo("啟動秒級急救機制 Watchdog (偵測間隔: %s 秒)", formatVal(r.GetConfig().FailoverInterval))
	ticker := time.NewTicker(time.Duration(r.GetConfig().FailoverInterval) * time.Second)
	defer ticker.Stop()

	consecutiveFails := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			logInfo("停止秒級急救機制 Watchdog (Context Cancelled)")
			return
		case <-ticker.C:
			if r.activeBorrowing.Load() {
				continue
			}

			if r.GetIsPaused() {
				continue
			}

			for _, groupName := range r.GetConfig().TargetGroups {
				group, err := r.api.GetProxyGroup(groupName)
				if err != nil || group.Now == "" {
					continue
				}

				if r.IsGroupLocked(groupName) {
					continue
				}

				testUrl := r.pickRandomURL()
				_, err = r.api.TestProxyDelay(group.Now, testUrl, r.GetConfig().TestTimeout)

				if err != nil {
					consecutiveFails[group.Now]++
					logWarning("Watchdog 偵測到 %s [%s] 無法連線 (%d/%d)", groupName, group.Now, consecutiveFails[group.Now], r.GetConfig().FailoverMaxFails)

					if consecutiveFails[group.Now] >= r.GetConfig().FailoverMaxFails {
						logError("🚑 %s [%s] 確認斷線，啟動緊急急救！", groupName, group.Now)

						if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnFailover {
							beeep.Notify("Clash 節點斷線", fmt.Sprintf("群組 [%s] 的節點 %s 失去連線，正在尋找替代節點...", groupName, group.Now), "")
						}

						candidates := group.All
						batchSize := 10
						found := false

						for i := 0; i < len(candidates); i += batchSize {
							end := i + batchSize
							if end > len(candidates) {
								end = len(candidates)
							}
							batch := candidates[i:end]
							
							var wg sync.WaitGroup
							results := make(chan string, len(batch))
							
							for _, alt := range batch {
								if alt == group.Now || r.GetBackoffRemaining(alt) > 0 {
									continue
								}
								wg.Add(1)
								go func(name string) {
									defer wg.Done()
									u := r.pickRandomURL()
									d, e := r.api.TestProxyDelay(name, u, r.GetConfig().TestTimeout)
									if e == nil && d > 0 {
										results <- name
									}
								}(alt)
							}
							
							go func() {
								wg.Wait()
								close(results)
							}()
							
							bestAlt := ""
							for res := range results {
								bestAlt = res
								break // Just take the first one that succeeds in this batch
							}
							
							if bestAlt != "" {
								logSuccess("🚑 找到替代節點: %s", bestAlt)
								if err := r.api.SelectProxy(groupName, bestAlt); err == nil {
									consecutiveFails[group.Now] = 0 // Reset
									if r.GetConfig().Notifications.Enable && r.GetConfig().Notifications.NotifyOnFailover {
										beeep.Notify("急救成功", fmt.Sprintf("已成功切換至備用節點: %s", bestAlt), "")
									}
									BroadcastRefresh()
									found = true
									break
								}
							}
						}
						
						if !found {
							logError("🚑 無法找到任何可用的替代節點。")
						}
					}
				} else {
					if consecutiveFails[group.Now] > 0 {
						logSuccess("Watchdog 偵測到 %s [%s] 恢復連線", groupName, group.Now)
					}
					consecutiveFails[group.Now] = 0
				}
			}
		}
	}
}"""

code = re.sub(r'func \(r \*Rover\) runFailoverWatchdog\(ctx context\.Context\) \{.*', watchdog_new, code, flags=re.DOTALL)

with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
