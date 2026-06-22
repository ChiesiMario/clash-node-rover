package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIUrl                 string        `yaml:"api_url"`
	APISecret              string        `yaml:"api_secret"`
	CheckInterval          time.Duration `yaml:"check_interval"`
	TargetGroups           []string      `yaml:"target_groups"`
	DedicatedTestGroup     string        `yaml:"dedicated_test_group"`
	TestURLs               []string      `yaml:"test_urls"`
	TestTimeout            time.Duration `yaml:"test_timeout"`
	DelayTolerance         int           `yaml:"delay_tolerance"` // milliseconds
	HistoryDays            int           `yaml:"history_days"`    // days
	CleanupDays            int           `yaml:"cleanup_days"`    // days
	MaxConcurrent          int           `yaml:"max_concurrent"`
	WebPort                int           `yaml:"web_port"`
	ClashProxyURL          string        `yaml:"clash_proxy_url"`
	BandwidthTestURL       string        `yaml:"bandwidth_test_url"`
	BandwidthTestInterval  int           `yaml:"bandwidth_test_interval"` // minutes
	ExplorationCooldown    int           `yaml:"exploration_cooldown_minutes"` // minutes
	MaxBackoffMinutes      int           `yaml:"max_backoff_minutes"`
}

const ConfigFile = "rover_config.yaml"
const OldConfigFile = "rover_config.json"

const defaultYAMLTemplate = `# Clash Node Rover 設定檔

# Clash 外部控制 API 的網址，通常為 http://127.0.0.1:9090
api_url: "%s"

# Clash API 的密碼 (secret)，如果沒有設定請留空
api_secret: "%s"

# 背景檢查與測速的間隔時間 (例如 60s, 1m, 5m)
check_interval: %ds

# 要被 Rover 控制與切換節點的 Clash 代理群組名稱 (支援多個群組)
target_groups:
  - "🤖 Node Rover"

# 專屬的無感測速群組 (Optional，留空則自動借用上述的 target_groups)。
# 如果設定此群組，Rover 測速時將不再借用你正在上網的群組，達成 100% 無感背景測速。
# (強烈建議：需配合 Clash 設定檔開啟獨立 Port 並將 clash_proxy_url 改為該 Port)
dedicated_test_group: ""

# 用來進行 Ping 測試的目標網址列表，會隨機抽取一個進行測試
test_urls:
  - "http://www.gstatic.com/generate_204"
  - "http://cp.cloudflare.com/generate_204"
  - "http://www.apple.com/library/test/success.html"

# Ping 測試的超時時間
test_timeout: 5s

# 延遲容忍度 (毫秒)。只有當新節點的延遲比目前節點快超過此數值時，才會進行切換
delay_tolerance: 100

# 歷史紀錄保留天數 (用於計算品質分數與網頁折線圖)
history_days: 7

# 資料庫自動瘦身機制 (超過此天數的陳舊日誌將被自動刪除並壓縮資料庫體積)
cleanup_days: 7

# 最大併發測速數量 (數字越大測速越快，但可能短暫佔用系統資源)
max_concurrent: 10

# Web 儀表板的監聽埠
web_port: 9091

# Clash 的 HTTP 代理網址 (用於真實下載測速)
clash_proxy_url: "http://127.0.0.1:7890"

# 真實頻寬測速用的下載檔案網址 (建議使用測速專用檔案)
bandwidth_test_url: "http://speedtest.tele2.net/1MB.zip"

# 同一個節點的真實頻寬測速冷卻時間 (分鐘)。這段時間內不會重複消耗流量測速
bandwidth_test_interval: 60

# 潛力節點面試 (探索) 的獨立冷卻時間 (分鐘)。
# 面試過的節點在這段時間內不會再次被面試，將機會讓給其他潛力節點
exploration_cooldown_minutes: 60

# 發生連線錯誤時的退避冷卻上限 (分鐘)。失敗越多次，冷卻越久，最高不超過此數值
max_backoff_minutes: 30
`

func writeYAMLConfig(cfg *Config) error {
	yamlStr := fmt.Sprintf(defaultYAMLTemplate,
		cfg.APIUrl,
		cfg.APISecret,
		int(cfg.CheckInterval.Seconds()),
	)
	return os.WriteFile(ConfigFile, []byte(yamlStr), 0644)
}

func loadConfig() (*Config, error) {
	// 檢查是否有舊的 JSON 設定檔，如果有，則讀取並升級成 YAML
	if _, err := os.Stat(OldConfigFile); err == nil {
		fmt.Println("發現舊版 rover_config.json，正在為您升級為 YAML 格式...")
		data, err := os.ReadFile(OldConfigFile)
		if err == nil {
			var cfg Config
			// 為了相容 JSON 欄位，我們暫時用一個匿名的 struct 來解析舊格式
			var oldCfg struct {
				APIUrl                 string        `json:"api_url"`
				APISecret              string        `json:"api_secret"`
				CheckInterval          time.Duration `json:"check_interval"`
				TargetGroup            string        `json:"target_group"`
				TargetGroups           []string      `json:"target_groups"`
			}
			json.Unmarshal(data, &oldCfg)
			
			cfg.APIUrl = oldCfg.APIUrl
			if cfg.APIUrl == "" { cfg.APIUrl = "http://127.0.0.1:9090" }
			cfg.APISecret = oldCfg.APISecret
			cfg.CheckInterval = oldCfg.CheckInterval
			if cfg.CheckInterval == 0 { cfg.CheckInterval = 60 * time.Second }
			
			// 自動將單一群組轉換為陣列
			if len(oldCfg.TargetGroups) > 0 {
				cfg.TargetGroups = oldCfg.TargetGroups
			} else if oldCfg.TargetGroup != "" {
				cfg.TargetGroups = []string{oldCfg.TargetGroup}
			} else {
				cfg.TargetGroups = []string{"🤖 Node Rover"}
			}
			
			// 寫入帶有註解的 YAML
			writeYAMLConfig(&cfg)
			// 刪除舊檔案
			os.Rename(OldConfigFile, OldConfigFile+".bak")
			fmt.Println("設定檔已成功升級為 rover_config.yaml！")
		}
	}

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return promptForConfig()
		}
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// 確保就算讀取舊版設定檔，也能有合理的預設值
	if len(cfg.TestURLs) == 0 {
		cfg.TestURLs = []string{"http://www.gstatic.com/generate_204", "http://cp.cloudflare.com/generate_204"}
	}
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = 60 * time.Second
	}
	if cfg.TestTimeout <= 0 {
		cfg.TestTimeout = 5 * time.Second
	}
	if cfg.DelayTolerance <= 0 {
		cfg.DelayTolerance = 100
	}
	if cfg.HistoryDays <= 0 {
		cfg.HistoryDays = 7
	}
	if cfg.CleanupDays <= 0 {
		cfg.CleanupDays = 7
	}
	if len(cfg.TargetGroups) == 0 {
		// 如果 YAML 裡面剛好沒有設定 (可能是舊版 YAML 升級)，則放入預設
		cfg.TargetGroups = []string{"🤖 Node Rover"}
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 10
	}
	if cfg.WebPort <= 0 {
		cfg.WebPort = 9091
	}
	if cfg.ClashProxyURL == "" {
		cfg.ClashProxyURL = "http://127.0.0.1:7890"
	}
	if cfg.BandwidthTestURL == "" {
		cfg.BandwidthTestURL = "http://speedtest.tele2.net/1MB.zip"
	}
	if cfg.BandwidthTestInterval <= 0 {
		cfg.BandwidthTestInterval = 60
	}
	if cfg.ExplorationCooldown <= 0 {
		cfg.ExplorationCooldown = 60
	}
	if cfg.MaxBackoffMinutes <= 0 {
		cfg.MaxBackoffMinutes = 30
	}

	return &cfg, nil
}

func promptForConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("首次執行設定。請設定 Node Rover。")

	fmt.Print("請輸入 Clash API 網址 [http://127.0.0.1:9090]: ")
	apiUrl, _ := reader.ReadString('\n')
	apiUrl = strings.TrimSpace(apiUrl)
	if apiUrl == "" {
		apiUrl = "http://127.0.0.1:9090"
	}

	fmt.Print("請輸入 Clash API 密碼 (如果沒有請留空): ")
	apiSecret, _ := reader.ReadString('\n')
	apiSecret = strings.TrimSpace(apiSecret)

	fmt.Print("請輸入檢查間隔 (秒) [60]: ")
	intervalStr, _ := reader.ReadString('\n')
	intervalStr = strings.TrimSpace(intervalStr)
	interval := 60
	if intervalStr != "" {
		if val, err := strconv.Atoi(intervalStr); err == nil && val > 0 {
			interval = val
		} else {
			fmt.Println("無效的間隔，使用預設值 60 秒。")
		}
	}

	cfg := &Config{
		APIUrl:                 apiUrl,
		APISecret:              apiSecret,
		CheckInterval:          time.Duration(interval) * time.Second,
		TargetGroups:           []string{"🤖 Node Rover"},
		TestURLs:               []string{"http://www.gstatic.com/generate_204", "http://cp.cloudflare.com/generate_204", "http://www.apple.com/library/test/success.html"},
		TestTimeout:            5 * time.Second,
		DelayTolerance:         100,
		HistoryDays:            7,
		CleanupDays:            7,
		MaxConcurrent:          10,
		WebPort:                9091,
		ClashProxyURL:          "http://127.0.0.1:7890",
		BandwidthTestURL:       "http://speedtest.tele2.net/1MB.zip",
		ExplorationCooldown:    60,
		MaxBackoffMinutes:      30,
	}

	if err := writeYAMLConfig(cfg); err != nil {
		return nil, err
	}
	
	fmt.Println("設定已儲存至", ConfigFile)

	return cfg, nil
}
