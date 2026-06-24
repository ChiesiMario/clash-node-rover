package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

var (
	globalRover *Rover
	globalDB    *DB
	logFile     io.WriteCloser
)

type LogRotator struct {
	filename string
	maxSize  int64
	file     *os.File
	size     int64
	mu       sync.Mutex
}

func NewLogRotator(filename string, maxSize int64) (*LogRotator, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	size := int64(0)
	if err == nil {
		size = info.Size()
	}
	return &LogRotator{filename: filename, maxSize: maxSize, file: f, size: size}, nil
}

func (l *LogRotator) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.size+int64(len(p)) > l.maxSize {
		l.file.Close()
		os.Rename(l.filename, l.filename+".old")
		f, err := os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		l.file = f
		l.size = 0
	}

	n, err = l.file.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *LogRotator) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func main() {
	// Initialize logger (Max 10MB per file)
	rotator, err := NewLogRotator("rover.log", 10*1024*1024)
	if err == nil {
		logFile = rotator
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}
	log.SetFlags(0) // 關閉預設的日期時間前綴
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
