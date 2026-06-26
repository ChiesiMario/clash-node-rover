package main

import (
	"embed"
	"io"
	"log"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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

	// 建立 Wails App
	app := NewApp(rover, db)

	// 啟動 Wails
	err = wails.Run(&options.App{
		Title:  "Clash Node Rover",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
