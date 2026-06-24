import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

# Fix the missing 'P' issue by rewording
code = code.replace(
    'logSuccess("Ping 測試完成！有效節點: %s, 失敗/超時: %s", formatVal(successCount), formatVal(failCount))',
    'logSuccess(" ICMP 測速環節完成！有效節點: %s, 失敗/超時: %s", formatVal(successCount), formatVal(failCount))'
)

# Improve browser test logging details
find_browser_loop = """				for _, targetURL := range testURLs {
					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)
					ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)

					var innerText string
					startTime := time.Now()
					err := chromedp.Run(ctx,"""

replace_browser_loop = """				for _, targetURL := range testURLs {
					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)
					ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)

					var checkingWhat string
					if strings.Contains(targetURL, "chatgpt.com") {
						checkingWhat = "ChatGPT"
					} else if strings.Contains(targetURL, "gemini.google.com") {
						checkingWhat = "Gemini"
					} else if strings.Contains(targetURL, "generativelanguage.googleapis") {
						checkingWhat = "Antigravity"
					} else {
						checkingWhat = "基礎連線"
					}
					logInfo("  ➤ 檢查 %s 中... (%s)", formatNode(candidate), checkingWhat)

					var innerText string
					startTime := time.Now()
					err := chromedp.Run(ctx,"""
code = code.replace(find_browser_loop, replace_browser_loop)

find_block_check = """						if blocked {
							logWarning("網頁載入成功但服務被封鎖/地區限制: %s", formatNode(candidate))
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
							break
						}

						logSuccess("網頁成功開啟且服務可用: %s (%d ms)", formatNode(candidate), loadTimeMs)
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}"""

replace_block_check = """						if blocked {
							logWarning("  ❌ 服務驗證失敗 (地區限制或封鎖): %s", checkingWhat)
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
							break
						}

						logSuccess("  ✅ 服務驗證通過: %s (%d ms)", checkingWhat, loadTimeMs)
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}"""
code = code.replace(find_block_check, replace_block_check)

find_browser_fail = """					if err != nil {
						logWarning("網頁開啟失敗: %s (%v)", formatNode(candidate), err)
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						break 
					} else {"""

replace_browser_fail = """					if err != nil {
						logWarning("  ❌ 服務連線超時或失敗: %s (%v)", checkingWhat, err)
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
						break 
					} else {"""
code = code.replace(find_browser_fail, replace_browser_fail)

# Also update groupReports so the tree looks better
find_tree_success = """targetReason = fmt.Sprintf("綜合分數 (%d ms) 且網頁測試成功 (%d ms)", stat.Score, totalLoadTime)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 網頁：測試 %s 成功", formatNode(candidate)))"""
replace_tree_success = """targetReason = fmt.Sprintf("綜合分數 (%d ms) 且各項服務驗證皆成功 (%d ms)", stat.Score, totalLoadTime)
					groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：節點 %s 支援所有必要服務", formatNode(candidate)))"""
code = code.replace(find_tree_success, replace_tree_success)

find_tree_fail = """groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 網頁：測試 %s 失敗，順延", formatNode(candidate)))"""
replace_tree_fail = """groupReports[groupName] = append(groupReports[groupName], colorError.Sprintf("🌐 驗證失敗：節點 %s 未通過服務驗證，淘汰並順延", formatNode(candidate)))"""
code = code.replace(find_tree_fail, replace_tree_fail)


with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
