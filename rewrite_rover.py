import os

with open("rover.go", "r", encoding="utf-8") as f:
    content = f.read()

replacement = """				if urlTestCache[candidate] == nil {
					urlTestCache[candidate] = make(map[string]bool)
				}

				// 嘗試從資料庫載入持久化快取 (避免反覆切換節點進行耗時網頁測試)
				for _, targetURL := range testURLs {
					if _, exists := urlTestCache[candidate][targetURL]; !exists {
						lastSuccess, err := r.db.GetLastBrowserSuccessTime(candidate, targetURL)
						if err == nil && !lastSuccess.IsZero() {
							if time.Since(lastSuccess) < r.GetConfig().BrowserCacheDuration {
								urlTestCache[candidate][targetURL] = true
							}
						}
					}
				}"""

content = content.replace("""				if urlTestCache[candidate] == nil {
					urlTestCache[candidate] = make(map[string]bool)
				}""", replacement)

content = content.replace('targetReason = fmt.Sprintf("共用跨群組快取：綜合分數 (%d ms) 且各項服務驗證皆成功", stat.Score)', 'targetReason = fmt.Sprintf("快取命中：綜合分數 (%d ms) 且各項服務驗證皆在有效期內", stat.Score)')
content = content.replace('groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：共用快取 %s 支援所有必要服務", formatNode(candidate)))', 'groupReports[groupName] = append(groupReports[groupName], colorSuccess.Sprintf("🌐 驗證通過：快取命中 %s 支援所有必要服務", formatNode(candidate)))')

with open("rover.go", "w", encoding="utf-8") as f:
    f.write(content)
