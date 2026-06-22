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
	APIUrl         string        `json:"api_url"`
	APISecret      string        `json:"api_secret"`
	CheckInterval  time.Duration `json:"check_interval"`
	TargetGroup    string        `json:"target_group"`
	TestURL        string        `json:"test_url"`
	TestTimeout    time.Duration `json:"test_timeout"`
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
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
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
		APIUrl:        apiUrl,
		APISecret:     apiSecret,
		CheckInterval: time.Duration(interval) * time.Second,
		TargetGroup:   "🤖 Node Rover",
		TestURL:       "http://www.gstatic.com/generate_204",
		TestTimeout:   5 * time.Second,
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
