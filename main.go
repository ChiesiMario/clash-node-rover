package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

var (
	globalRover *Rover
	globalDB    *DB
	logFile     *os.File
)

func main() {
	// Initialize logger
	f, err := os.OpenFile("rover.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		logFile = f
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("讀取設定檔失敗: %v", err)
	}

	db, err := InitDB()
	if err != nil {
		log.Fatalf("初始化資料庫失敗: %v", err)
	}
	globalDB = db

	api := NewAPIClient(cfg.APIUrl, cfg.APISecret)
	rover := NewRover(cfg, api, db)
	globalRover = rover

	// 將原本的 main 邏輯移交給 systray
	systray.Run(func() { onReady(cfg.WebPort) }, onExit)
}

func onReady(webPort int) {
	// 設定系統列圖示與提示
	systray.SetIcon(iconData)
	systray.SetTitle("Node Rover")
	systray.SetTooltip("Clash Node Rover - 網路守護中")

	// 設定右鍵選單
	mOpen := systray.AddMenuItem("🌐 開啟儀表板", "打開 Web 控制台")
	mForce := systray.AddMenuItem("⚡ 強制測速", "立刻進行節點測速")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("❌ 退出程式", "安全關閉 Node Rover")

	// 背景啟動核心與 Web
	go globalRover.Start()
	go StartWebServer(globalDB, globalRover, webPort)

	// 監聽選單點擊
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser(fmt.Sprintf("http://localhost:%d", webPort))
			case <-mForce.ClickedCh:
				if globalRover != nil {
					globalRover.ForceCheck()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Println("\n⚠️ 接收到關閉信號，正在安全關閉 Clash Node Rover...")

	if globalRover != nil {
		globalRover.Stop()
	}

	if globalDB != nil {
		if err := globalDB.Close(); err != nil {
			log.Printf("❌ 資料庫關閉時發生錯誤: %v\n", err)
		} else {
			log.Println("✅ 資料庫已安全關閉。")
		}
	}

	log.Println("👋 程式已安全退出，感謝使用！")

	if logFile != nil {
		logFile.Close()
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("無法開啟瀏覽器: %v", err)
	}
}
