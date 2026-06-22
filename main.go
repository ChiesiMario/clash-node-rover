package main

import (
	"log"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("讀取設定檔失敗: %v", err)
	}

	db, err := InitDB()
	if err != nil {
		log.Fatalf("初始化資料庫失敗: %v", err)
	}
	defer db.Close()

	webServer := NewWebServer(cfg, db)
	go webServer.Start()

	api := NewAPIClient(cfg.APIUrl, cfg.APISecret)
	rover := NewRover(cfg, api, db)

	rover.Run()
}
