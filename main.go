package main

import (
	"log"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("讀取設定檔失敗: %v", err)
	}

	api := NewAPIClient(cfg.APIUrl, cfg.APISecret)
	rover := NewRover(cfg, api)

	rover.Run()
}
