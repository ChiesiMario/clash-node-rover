package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
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

	api := NewAPIClient(cfg.APIUrl, cfg.APISecret)
	rover := NewRover(cfg, api, db)

	// 啟動背景測速引擎
	go rover.Start()

	// 啟動 Web 儀表板
	go StartWebServer(db, rover, cfg.WebPort)

	// 優雅關機機制 (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n⚠️ 接收到關閉信號，正在安全關閉 Clash Node Rover...")
	
	// 停止核心引擎
	rover.Stop()
	
	// 安全關閉資料庫連線
	if err := db.Close(); err != nil {
		log.Printf("❌ 資料庫關閉時發生錯誤: %v\n", err)
	} else {
		log.Println("✅ 資料庫已安全關閉。")
	}
	
	log.Println("👋 程式已安全退出，感謝使用！")
}
