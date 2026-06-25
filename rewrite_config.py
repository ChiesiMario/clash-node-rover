import os
import re

with open("config.go", "r", encoding="utf-8") as f:
    content = f.read()

# Add to Config struct
content = content.replace('	BrowserTestURLs   []string `yaml:"browser_test_urls"`', 
                          '	BrowserTestURLs      []string      `yaml:"browser_test_urls"`\n	BrowserCacheDuration time.Duration `yaml:"browser_cache_duration"`')

# Add to YAML template
yaml_addition = """# 無頭瀏覽器測試的目標網址清單 (預設為 Google 與 YouTube)
browser_test_urls:
  - "https://www.google.com"
  - "https://www.youtube.com"

# 服務驗證成功的快取時間 (預設 24h)，期間內若成功過則不再重複啟動瀏覽器測試
browser_cache_duration: 24h
"""
content = content.replace("""# 無頭瀏覽器測試的目標網址清單 (預設為 Google 與 YouTube)
browser_test_urls:
  - "https://www.google.com"
  - "https://www.youtube.com"
""", yaml_addition)

# Add to defaults in loadConfig
defaults_addition = """	if len(cfg.BrowserTestURLs) == 0 {
		cfg.BrowserTestURLs = []string{"https://www.google.com", "https://www.youtube.com"}
		// 如果原本沒設定 BrowserTestURLs，代表是舊版升級，預設開啟測試
		cfg.EnableBrowserTest = true
	}
	if cfg.BrowserCacheDuration <= 0 {
		cfg.BrowserCacheDuration = 24 * time.Hour
	}"""
content = content.replace("""	if len(cfg.BrowserTestURLs) == 0 {
		cfg.BrowserTestURLs = []string{"https://www.google.com", "https://www.youtube.com"}
		// 如果原本沒設定 BrowserTestURLs，代表是舊版升級，預設開啟測試
		cfg.EnableBrowserTest = true
	}""", defaults_addition)

# Add to promptForConfig
prompt_addition = """		EnableBrowserTest:   true,
		BrowserTestURLs:     []string{"https://www.google.com", "https://www.youtube.com"},
		BrowserCacheDuration: 24 * time.Hour,"""
content = content.replace("""		EnableBrowserTest:   true,
		BrowserTestURLs:     []string{"https://www.google.com", "https://www.youtube.com"},""", prompt_addition)

with open("config.go", "w", encoding="utf-8") as f:
    f.write(content)
