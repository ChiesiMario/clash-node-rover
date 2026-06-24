import re

with open('rover.go', 'r', encoding='utf-8') as f:
    code = f.read()

# Update testURL populating logic
find_testurl_init = """				allSuccess := true
				var totalLoadTime int
				testURLs := append([]string(nil), r.GetConfig().BrowserTestURLs...)
				if filter.CheckChatGPT {
					testURLs = append(testURLs, "https://chatgpt.com")
				}
				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}
				if filter.CheckAntigravity {
					testURLs = append(testURLs, "https://generativelanguage.googleapis.com")
				}"""

replace_testurl_init = """				allSuccess := true
				var totalLoadTime int
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
code = code.replace(find_testurl_init, replace_testurl_init)

# Update checkingWhat logic
find_checking_what = """					} else {
						checkingWhat = "基礎連線"
					}"""

replace_checking_what = """					} else {
						checkingWhat = "基礎連線 (" + targetURL + ")"
					}"""
code = code.replace(find_checking_what, replace_checking_what)


with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(code)
