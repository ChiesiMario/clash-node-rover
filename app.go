package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx   context.Context
	rover *Rover
	db    *DB
}

var globalApp *App

func NewApp(rover *Rover, db *DB) *App {
	globalApp = &App{
		rover: rover,
		db:    db,
	}
	return globalApp
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 背景啟動核心
	go a.rover.Start()
}

func (a *App) shutdown(ctx context.Context) {
	logInfo("⚠️ 接收到關閉信號，正在安全關閉 Node Rover...")
	if a.rover != nil {
		a.rover.Stop()
	}
	if a.db != nil {
		a.db.Close()
	}
}

// 供其他模組廣播事件使用
func BroadcastRefresh() {
	if globalApp != nil && globalApp.ctx != nil {
		wailsruntime.EventsEmit(globalApp.ctx, "refresh")
	}
}

func BroadcastSingleLog(entry WebLogEntry) {
	if globalApp != nil && globalApp.ctx != nil {
		wailsruntime.EventsEmit(globalApp.ctx, "log", entry)
	}
}

// ==========================================
// API Bindings for Frontend
// ==========================================

type GroupStatus struct {
	Name     string      `json:"name"`
	Now      string      `json:"now"`
	Provider string      `json:"provider"`
	All      int         `json:"all_count"`
	AllNodes []string    `json:"all_nodes"`
	Locked   bool        `json:"locked"`
	Filter   GroupFilter `json:"filter"`
}

func (a *App) GetGroups() []GroupStatus {
	var statuses []GroupStatus
	if a.rover.api == nil {
		return statuses
	}
	for _, gName := range a.rover.GetConfig().TargetGroups {
		g, err := a.rover.api.GetProxyGroup(gName)
		if err == nil {
			statuses = append(statuses, GroupStatus{
				Name:     gName,
				Now:      g.Now,
				Provider: GetNodeProvider(g.Now),
				All:      len(g.All),
				AllNodes: g.All,
				Locked:   a.rover.IsGroupLocked(gName),
				Filter:   a.rover.getGroupFilter(gName),
			})
		}
	}
	return statuses
}

func (a *App) GetAllProxyGroups() []GroupStatus {
	var statuses []GroupStatus
	if a.rover.api == nil {
		return statuses
	}
	allGroupNames, err := a.rover.api.GetSelectors()
	if err != nil {
		return statuses
	}

	for _, gName := range allGroupNames {
		// 預設過濾掉 GLOBAL, DIRECT, REJECT 等不該作為測速對象的群組
		if gName == "GLOBAL" || gName == "DIRECT" || gName == "REJECT" || gName == "COMPATIBLE" {
			continue
		}
		g, err := a.rover.api.GetProxyGroup(gName)
		if err == nil {
			statuses = append(statuses, GroupStatus{
				Name:     gName,
				Now:      g.Now,
				Provider: GetNodeProvider(g.Now),
				All:      len(g.All),
				AllNodes: g.All,
				Locked:   a.rover.IsGroupLocked(gName),
				Filter:   a.rover.getGroupFilter(gName),
			})
		}
	}
	return statuses
}

func (a *App) SetGroupLocked(group string, locked bool) {
	a.rover.SetGroupLocked(group, locked)
	BroadcastRefresh()
}

func (a *App) GetSetup() map[string]interface{} {
	return map[string]interface{}{
		"is_configured": a.rover.GetConfig().APIUrl != "",
		"api_url":       a.rover.GetConfig().APIUrl,
	}
}

func (a *App) SaveSetup(apiUrl, secret string) error {
	if apiUrl == "" {
		return fmt.Errorf("API URL cannot be empty")
	}
	testAPI := NewAPIClient(apiUrl, secret)
	if err := testAPI.VerifyConnection(); err != nil {
		return fmt.Errorf("連線驗證失敗: %v", err)
	}

	a.rover.cfgMutex.Lock()
	a.rover.cfg.APIUrl = apiUrl
	a.rover.cfg.APISecret = secret
	a.rover.cfgMutex.Unlock()

	writeYAMLConfig(a.rover.GetConfig())

	a.rover.api = testAPI
	a.rover.ApiConnected.Store(true)
	return nil
}

type ConfigDTO struct {
	APIUrl               string   `json:"api_url"`
	APISecret            string   `json:"api_secret"`
	CheckInterval        int      `json:"check_interval"`
	TargetGroups         []string `json:"target_groups"`
	DedicatedTestGroup   string   `json:"dedicated_test_group"`
	TestURLs             []string `json:"test_urls"`
	TestTimeout          int      `json:"test_timeout"`
	ToleranceMs          int      `json:"tolerance_ms"`
	CleanupDays          int      `json:"cleanup_days"`
	MaxConcurrent        int      `json:"max_concurrent"`
	WebPort              int      `json:"web_port"`
	ClashProxyURL        string   `json:"clash_proxy_url"`
	MaxBackoffCycles     int      `json:"max_backoff_cycles"`
	EnableBrowserTest    bool     `json:"enable_browser_test"`
	BrowserTestURLs      []string `json:"browser_test_urls"`
}

func (a *App) GetConfigInfo() ConfigDTO {
	cfg := a.rover.GetConfig()
	return ConfigDTO{
		APIUrl:               cfg.APIUrl,
		APISecret:            cfg.APISecret,
		CheckInterval:        int(cfg.CheckInterval.Seconds()),
		TargetGroups:         cfg.TargetGroups,
		DedicatedTestGroup:   cfg.DedicatedTestGroup,
		TestURLs:             cfg.TestURLs,
		TestTimeout:          int(cfg.TestTimeout.Seconds()),
		ToleranceMs:          cfg.ToleranceMs,
		CleanupDays:          cfg.CleanupDays,
		MaxConcurrent:        cfg.MaxConcurrent,
		WebPort:              cfg.WebPort,
		ClashProxyURL:        cfg.ClashProxyURL,
		MaxBackoffCycles:     cfg.MaxBackoffCycles,
		EnableBrowserTest:    cfg.EnableBrowserTest,
		BrowserTestURLs:      cfg.BrowserTestURLs,
	}
}

func (a *App) SaveConfig(dto ConfigDTO) error {
	newCfg := &Config{
		APIUrl:               dto.APIUrl,
		APISecret:            dto.APISecret,
		CheckInterval:        time.Duration(dto.CheckInterval) * time.Second,
		TargetGroups:         dto.TargetGroups,
		DedicatedTestGroup:   dto.DedicatedTestGroup,
		TestURLs:             dto.TestURLs,
		TestTimeout:          time.Duration(dto.TestTimeout) * time.Second,
		ToleranceMs:          dto.ToleranceMs,
		CleanupDays:          dto.CleanupDays,
		MaxConcurrent:        dto.MaxConcurrent,
		WebPort:              dto.WebPort,
		ClashProxyURL:        dto.ClashProxyURL,
		MaxBackoffCycles:     dto.MaxBackoffCycles,
		EnableBrowserTest:    dto.EnableBrowserTest,
		BrowserTestURLs:      dto.BrowserTestURLs,
	}

	if err := writeYAMLConfig(newCfg); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	a.rover.cfgMutex.Lock()
	a.rover.cfg = newCfg
	a.rover.api = NewAPIClient(newCfg.APIUrl, newCfg.APISecret)
	a.rover.cfgMutex.Unlock()
	a.rover.checkBrowserTestURLsChanged()
	
	// 強制刷新狀態並通知前端
	a.rover.ApiConnected.Store(true)
	BroadcastRefresh()
	a.rover.ForceCheck()

	return nil
}

func (a *App) GetSelectors() ([]string, error) {
	api := a.rover.GetAPI()
	if api == nil {
		return []string{}, nil
	}
	return api.GetSelectors()
}

func (a *App) TestConnection(apiUrl, secret string) error {
	testAPI := NewAPIClient(apiUrl, secret)
	return testAPI.VerifyConnection()
}

type StatNode struct {
	Name                    string         `json:"Name"`
	AvgDelay                int            `json:"AvgDelay"`
	Jitter                  int            `json:"Jitter"`
	Score                   int            `json:"Score"`
	Provider                string         `json:"provider"`
	HighestInGroups         []string       `json:"highest_in_groups"`
	BackoffRemaining        int            `json:"backoff_remaining"`
	BrowserBackoffRemaining map[string]int `json:"browser_backoff_remaining"`
	IsDead                  bool           `json:"is_dead"`
}

func (a *App) GetStats() []StatNode {
	statMap := a.rover.GetStatResults()
	highestInGroups := make(map[string][]string)
	
	if a.rover.GetAPI() != nil {
		for _, groupName := range a.rover.GetConfig().TargetGroups {
			g, err := a.rover.GetAPI().GetProxyGroup(groupName)
			if err == nil && g.Now != "" {
				highestInGroups[g.Now] = append(highestInGroups[g.Now], groupName)
			}
		}
	}

	list := make([]StatNode, 0)
	for _, sc := range statMap {
		isDead := sc.Err != nil
		score := sc.Score
		if isDead {
			score = 99999
		}
		list = append(list, StatNode{
			Name:                    sc.Name,
			AvgDelay:                sc.AvgDelay,
			Jitter:                  sc.Jitter,
			Score:                   score,
			Provider:                GetNodeProvider(sc.Name),
			HighestInGroups:         highestInGroups[sc.Name],
			BackoffRemaining:        a.rover.GetBackoffRemaining(sc.Name),
			BrowserBackoffRemaining: a.rover.GetBrowserBackoffRemaining(sc.Name),
			IsDead:                  isDead,
		})
	}

	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].Score > list[j].Score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	return list
}

func (a *App) GetHistory(nodeName string) map[string]interface{} {
	pingHistory, _ := a.db.GetNodeHistory(nodeName, 24)
	browserHistory, _ := a.db.GetBrowserHistory(nodeName, 24)
	return map[string]interface{}{
		"ping":    pingHistory,
		"browser": browserHistory,
	}
}

func (a *App) GetStatus() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	dbSizeMB := 0.0
	if stat, err := os.Stat("rover.db"); err == nil {
		dbSizeMB = float64(stat.Size()) / 1024 / 1024
	}

	return map[string]interface{}{
		"is_running":    a.rover.IsRunning.Load(),
		"is_paused":     a.rover.GetIsPaused(),
		"is_configured": a.rover.GetConfig().APIUrl != "",
		"api_connected": a.rover.ApiConnected.Load(),
		"is_ready":      len(a.rover.GetConfig().TargetGroups) > 0,
		"mem_alloc_mb":  float64(m.Alloc) / 1024 / 1024,
		"mem_sys_mb":    float64(m.Sys) / 1024 / 1024,
		"db_size_mb":    dbSizeMB,
		"log_count":     GetLogHistoryCount(),
	}
}

func (a *App) TogglePause() bool {
	return a.rover.TogglePause()
}

func (a *App) SwitchNode(group, node string) error {
	err := a.rover.api.SelectProxy(group, node)
	if err == nil {
		logInfo("⚡ 收到手動切換指令：將群組 [%s] 切換至 %s", group, node)
	}
	return err
}

func (a *App) ManualTrigger() {
	select {
	case a.rover.ManualTrigger <- struct{}{}:
	default:
	}
}

func (a *App) GetLogHistory() []WebLogEntry {
	logHistoryMutex.Lock()
	defer logHistoryMutex.Unlock()
	historyCopy := make([]WebLogEntry, len(logHistory))
	copy(historyCopy, logHistory)
	return historyCopy
}
