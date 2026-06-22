package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIUrl                 string        `json:"api_url"`
	APISecret              string        `json:"api_secret"`
	CheckInterval          time.Duration `json:"check_interval"`
	TargetGroup            string        `json:"target_group"`
	TestURLs               []string      `json:"test_urls"`
	TestTimeout            time.Duration `json:"test_timeout"`
	DelayTolerance         int           `json:"delay_tolerance"` // milliseconds
	HistoryDays            int           `json:"history_days"`    // days
	MaxConcurrent          int           `json:"max_concurrent"`
	WebPort                int           `json:"web_port"`
	ClashProxyURL          string        `json:"clash_proxy_url"`
	BandwidthTestURL       string        `json:"bandwidth_test_url"`
	BandwidthThresholdKbps float64       `json:"bandwidth_threshold_kbps"`
	BandwidthTestInterval  int           `json:"bandwidth_test_interval"` // minutes
	MaxBackoffMinutes      int           `json:"max_backoff_minutes"`
}

const ConfigFile = "rover_config.json"

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return promptForConfig()
		}
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
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
	if cfg.BandwidthThresholdKbps <= 0 {
		cfg.BandwidthThresholdKbps = 500
	}
	if cfg.BandwidthTestInterval <= 0 {
		cfg.BandwidthTestInterval = 60
	}
	if cfg.MaxBackoffMinutes <= 0 {
		cfg.MaxBackoffMinutes = 30
	}

	// 自動將補齊預設值後的完整設定寫回檔案，方便用戶直接編輯
	updatedData, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		os.WriteFile(ConfigFile, updatedData, 0644)
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
		TargetGroup:            "🤖 Node Rover",
		TestURLs:               []string{"http://www.gstatic.com/generate_204", "http://cp.cloudflare.com/generate_204", "http://www.apple.com/library/test/success.html"},
		TestTimeout:            5 * time.Second,
		DelayTolerance:         100,
		HistoryDays:            7,
		MaxConcurrent:          10,
		WebPort:                9091,
		ClashProxyURL:          "http://127.0.0.1:7890",
		BandwidthTestURL:       "http://speedtest.tele2.net/1MB.zip",
		BandwidthThresholdKbps: 500.0,
		MaxBackoffMinutes:      30,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(ConfigFile, data, 0644); err != nil {
		return nil, err
	}
	fmt.Println("設定已儲存至", ConfigFile)

	return cfg, nil
}
