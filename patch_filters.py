import re

def patch_rover():
    with open('rover.go', 'r', encoding='utf-8') as f:
        code = f.read()

    # Add GroupFilter and getGroupFilter
    group_filter_code = """
type GroupFilter struct {
	KeywordRegex string `json:"keyword_regex"`
	CheckChatGPT bool   `json:"check_chatgpt"`
	CheckGemini  bool   `json:"check_gemini"`
}

func (r *Rover) getGroupFilter(groupName string) GroupFilter {
	var f GroupFilter
	val, err := r.db.GetMetadata("group_filter_" + groupName)
	if err == nil && val != "" {
		json.Unmarshal([]byte(val), &f)
	}
	return f
}
"""
    if "type GroupFilter struct" not in code:
        code += group_filter_code

    # Add regex import if not there
    if '"regexp"' not in code:
        code = re.sub(r'import \(', 'import (\n\t"regexp"\n', code)

    # Patch runCheckCycle Name filtering
    find_str = """		if len(group.All) == 0 {
			continue
		}
		groupNodesMap[groupName] = group.All
		groupNowMap[groupName] = group.Now"""
    
    replace_str = """		if len(group.All) == 0 {
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
		groupNowMap[groupName] = group.Now"""
    
    if "var rx *regexp.Regexp" not in code:
        code = code.replace(find_str, replace_str)


    # Patch Browser Testing loop
    # We need to inject additional TargetURLs for Gemini and ChatGPT, and check for specific text
    find_browser_loop = """				for _, targetURL := range r.GetConfig().BrowserTestURLs {
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
				}"""

    replace_browser_loop = """				filter := r.getGroupFilter(groupName)
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)
				if filter.CheckChatGPT {
					testURLs = append(testURLs, "https://chatgpt.com")
				}
				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}

				for _, targetURL := range testURLs {
					ctx, cancelCtx := chromedp.NewContext(browserAllocCtx)
					ctx, cancelTimeout := context.WithTimeout(ctx, 15*time.Second)

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
						logWarning("網頁開啟失敗: %s (%v)", formatNode(candidate), err)
						r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
						allSuccess = false
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
						}

						if blocked {
							logWarning("網頁載入成功但服務被封鎖/地區限制: %s", formatNode(candidate))
							r.db.InsertBrowserLog(candidate, targetURL, false, loadTimeMs)
							allSuccess = false
							break
						}

						logSuccess("網頁成功開啟且服務可用: %s (%d ms)", formatNode(candidate), loadTimeMs)
						r.db.InsertBrowserLog(candidate, targetURL, true, loadTimeMs)
						totalLoadTime += loadTimeMs
					}
				}"""
    if "testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)" not in code:
        code = code.replace(find_browser_loop, replace_browser_loop)

    with open('rover.go', 'w', encoding='utf-8') as f:
        f.write(code)

def patch_web():
    with open('web.go', 'r', encoding='utf-8') as f:
        code = f.read()

    # Add API Handlers for GET and POST /api/groups/filter
    api_handlers = """
	http.HandleFunc("/api/groups/filter", func(w http.ResponseWriter, req *http.Request) {
		groupName := req.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing group", http.StatusBadRequest)
			return
		}

		if req.Method == "GET" {
			val, _ := db.GetMetadata("group_filter_" + groupName)
			if val == "" {
				val = `{"keyword_regex": "", "check_chatgpt": false, "check_gemini": false}`
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(val))
			return
		}

		if req.Method == "POST" {
			var filter struct {
				KeywordRegex string `json:"keyword_regex"`
				CheckChatGPT bool   `json:"check_chatgpt"`
				CheckGemini  bool   `json:"check_gemini"`
			}
			if err := json.NewDecoder(req.Body).Decode(&filter); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			
			b, _ := json.Marshal(filter)
			db.SetMetadata("group_filter_"+groupName, string(b))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
"""
    if "/api/groups/filter" not in code:
        code = code.replace('http.HandleFunc("/api/groups/lock", func(w http.ResponseWriter, r *http.Request) {', api_handlers + '\n\thttp.HandleFunc("/api/groups/lock", func(w http.ResponseWriter, r *http.Request) {')

    with open('web.go', 'w', encoding='utf-8') as f:
        f.write(code)


patch_rover()
patch_web()
