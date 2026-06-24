import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

find_cache_check = """				if success, exists := browserTestCache[candidate]; exists {"""
replace_cache_check = """				filter := r.getGroupFilter(groupName)
				cacheKey := fmt.Sprintf("%s|%v|%v|%v", candidate, filter.CheckChatGPT, filter.CheckGemini, filter.CheckAntigravity)
				
				if success, exists := browserTestCache[cacheKey]; exists {"""
code = code.replace(find_cache_check, replace_cache_check)

find_filter_init = """				allSuccess := true
				var totalLoadTime int
				filter := r.getGroupFilter(groupName)
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)"""
replace_filter_init = """				allSuccess := true
				var totalLoadTime int
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)"""
code = code.replace(find_filter_init, replace_filter_init)

find_cache_save = """				browserTestCache[candidate] = allSuccess"""
replace_cache_save = """				browserTestCache[cacheKey] = allSuccess"""
code = code.replace(find_cache_save, replace_cache_save)

with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
