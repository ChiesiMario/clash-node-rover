import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

# 1. Replace browserTestCache initialization
code = code.replace("browserTestCache := make(map[string]bool)", "urlTestCache := make(map[string]map[string]bool)")

# 2. Remove the old cacheKey and upfront cache check
old_cache_check = """				filter := r.getGroupFilter(groupName)
				cacheKey := fmt.Sprintf("%s|%v|%v|%v", candidate, filter.CheckChatGPT, filter.CheckGemini, filter.CheckAntigravity)
				
				if success, exists := browserTestCache[cacheKey]; exists {
					if success {
						targetNode = candidate
						targetReason = fmt.Sprintf("共用跨群組快取：綜合分數 (%d ms) 且網頁測試成功", stat.Score)
						groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 網頁：共用快取 %s 成功", formatNode(candidate)))
						break
					} else {
						groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 網頁：共用快取 %s 失敗，順延", formatNode(candidate)))
						continue
					}
				}"""
code = code.replace(old_cache_check, "")

# 3. Before proxy switching, define testURLs and do the new cache logic
old_proxy_switch_start = """				var borrowGroup string"""

new_proxy_switch_start = """				filter := r.getGroupFilter(groupName)
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)
				if filter.CheckChatGPT {
					testURLs = append(testURLs, "https://chatgpt.com")
				}
				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}
				if filter.CheckAntigravity {
					testURLs = append(testURLs, "https://generativelanguage.googleapis.com")
				}

				if urlTestCache[candidate] == nil {
					urlTestCache[candidate] = make(map[string]bool)
				}

				allCachedAndSuccess := true
				anyCachedFailed := false
				for _, targetURL := range testURLs {
					if success, exists := urlTestCache[candidate][targetURL]; exists {
						if !success {
							anyCachedFailed = true
							break
						}
					} else {
						allCachedAndSuccess = false
					}
				}

				if anyCachedFailed {
					groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 驗證失敗：部分服務在快取中已失敗，淘汰並順延: %s", formatNode(candidate)))
					continue
				}

				if allCachedAndSuccess {
					targetNode = candidate
					targetReason = fmt.Sprintf("共用跨群組快取：綜合分數 (%d ms) 且各項服務驗證皆成功", stat.Score)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：共用快取 %s 支援所有必要服務", formatNode(candidate)))
					break
				}

				var borrowGroup string"""
code = code.replace(old_proxy_switch_start, new_proxy_switch_start)

# 4. Remove the redundant testURLs initialization further down
redundant_testurls = """				filter := r.getGroupFilter(groupName)
				var testURLs []string
				if filter.CheckChatGPT {
					testURLs = append(testURLs, "https://chatgpt.com")
				}
				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}
				if filter.CheckAntigravity {
					testURLs = append(testURLs, "https://generativelanguage.googleapis.com")
				}
				if len(testURLs) == 0 {
					testURLs = append(testURLs, r.GetConfig().BrowserTestURLs...)
				}"""
code = code.replace(redundant_testurls, "")

# 5. In the testURLs loop, check cache for the individual URL
find_loop_start = """				for _, targetURL := range testURLs {
					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)"""

replace_loop_start = """				for _, targetURL := range testURLs {
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

					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)"""
code = code.replace(find_loop_start, replace_loop_start)

# 6. Remove redundant checkingWhat inside loop
redundant_checkingwhat = """					var checkingWhat string
					if strings.Contains(targetURL, "chatgpt.com") {
						checkingWhat = "ChatGPT"
					} else if strings.Contains(targetURL, "gemini.google.com") {
						checkingWhat = "Gemini"
					} else if strings.Contains(targetURL, "generativelanguage.googleapis") {
						checkingWhat = "Antigravity"
					} else {
						checkingWhat = "基礎連線 (" + targetURL + ")"
					}"""
code = code.replace(redundant_checkingwhat, "")

# 7. Update cache saving and old `browserTestCache[cacheKey] = allSuccess`
find_block_check = """						if blocked {
							logWarning("  ❌ 服務驗證失敗 (地區限制或封鎖): %s", checkingWhat)
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
							break
						}

						logSuccess("  ✅ 服務驗證通過: %s (%d ms)", checkingWhat, loadTimeMs)
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}"""
replace_block_check = """						if blocked {
							logWarning("  ❌ 服務驗證失敗 (地區限制或封鎖): %s", checkingWhat)
							urlTestCache[candidate][targetURL] = false
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
							break
						}

						logSuccess("  ✅ 服務驗證通過: %s (%d ms)", checkingWhat, loadTimeMs)
						urlTestCache[candidate][targetURL] = true
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}"""
code = code.replace(find_block_check, replace_block_check)

find_fail_save = """					if err != nil {
						logWarning("  ❌ 服務連線超時或失敗: %s (%v)", checkingWhat, err)
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						break 
					} else {"""
replace_fail_save = """					if err != nil {
						logWarning("  ❌ 服務連線超時或失敗: %s (%v)", checkingWhat, err)
						urlTestCache[candidate][targetURL] = false
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						break 
					} else {"""
code = code.replace(find_fail_save, replace_fail_save)

# Remove the old node-level cache saving at the end
code = code.replace("browserTestCache[cacheKey] = allSuccess\n", "")

with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
