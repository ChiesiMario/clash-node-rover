package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

//go:embed all:frontend/dist
var frontendDist embed.FS

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsClients      = make(map[*websocket.Conn]bool)
	wsClientsMutex sync.Mutex
	broadcastChan  = make(chan interface{}, 1000)
)

func init() {
	go func() {
		for msg := range broadcastChan {
			wsClientsMutex.Lock()
			clients := make([]*websocket.Conn, 0, len(wsClients))
			for client := range wsClients {
				clients = append(clients, client)
			}
			wsClientsMutex.Unlock()

			for _, client := range clients {
				client.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
				if err := client.WriteJSON(msg); err != nil {
					client.Close()
					wsClientsMutex.Lock()
					delete(wsClients, client)
					wsClientsMutex.Unlock()
				}
			}
		}
	}()
}

func BroadcastRefresh() {
	select {
	case broadcastChan <- map[string]string{"type": "refresh"}:
	default:
	}
}

func BroadcastSingleLog(entry WebLogEntry) {
	select {
	case broadcastChan <- map[string]interface{}{
		"type":  "log",
		"entry": entry,
	}:
	default:
	}
}

func StartWebServer(db *DB, rover *Rover, port int) {
		distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		log.Fatalf("無法載入前端資源: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(distFS)))

	http.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		type GroupStatus struct {
			Name     string      `json:"name"`
			Now      string      `json:"now"`
			Provider string      `json:"provider"`
			All      int         `json:"all_count"`
			AllNodes []string    `json:"all_nodes"`
			Locked   bool        `json:"locked"`
			Filter   GroupFilter `json:"filter"`
		}
		var statuses []GroupStatus
		for _, gName := range rover.GetConfig().TargetGroups {
			g, err := rover.api.GetProxyGroup(gName)
			if err == nil {
				statuses = append(statuses, GroupStatus{
					Name:     gName,
					Now:      g.Now,
					Provider: GetNodeProvider(g.Now),
					All:      len(g.All),
					AllNodes: g.All,
					Locked:   rover.IsGroupLocked(gName),
					Filter:   rover.getGroupFilter(gName),
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})

	
	http.HandleFunc("/api/groups/filter", func(w http.ResponseWriter, req *http.Request) {
		groupName := req.URL.Query().Get("group")
		if groupName == "" {
			http.Error(w, "Missing group", http.StatusBadRequest)
			return
		}

		if req.Method == "GET" {
			val, _ := db.GetMetadata("group_filter_" + groupName)
			if val == "" {
				val = `{"keyword_regex": ""}`
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(val))
			return
		}

		if req.Method == "POST" {
			var filter struct {
				KeywordRegex string `json:"keyword_regex"`
			}
			if err := json.NewDecoder(req.Body).Decode(&filter); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			
			b, _ := json.Marshal(filter)
			db.SetMetadata("group_filter_"+groupName, string(b))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/groups/lock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Group  string `json:"group"`
			Locked bool   `json:"locked"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		rover.SetGroupLocked(req.Group, req.Locked)
		BroadcastRefresh()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/setup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"is_configured": rover.GetConfig().APIUrl != "",
				"api_url":       rover.GetConfig().APIUrl,
			})
			return
		}

		if r.Method == "POST" {
			var req struct {
				APIUrl    string `json:"api_url"`
				APISecret string `json:"api_secret"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			if req.APIUrl == "" {
				http.Error(w, "API URL cannot be empty", http.StatusBadRequest)
				return
			}

			// Verify connection
			testAPI := NewAPIClient(req.APIUrl, req.APISecret)
			if err := testAPI.VerifyConnection(); err != nil {
				http.Error(w, fmt.Sprintf("連線驗證失敗: %v", err), http.StatusBadRequest)
				return
			}

			// Update config
			rover.cfgMutex.Lock()
			rover.cfg.APIUrl = req.APIUrl
			rover.cfg.APISecret = req.APISecret
			rover.cfgMutex.Unlock()
			
			writeYAMLConfig(rover.GetConfig())
			
			// Update API client
			rover.api = testAPI

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	type ConfigDTO struct {
		APIUrl               string             `json:"api_url"`
		APISecret            string             `json:"api_secret"`
		CheckInterval        int                `json:"check_interval"`
		TargetGroups         []string           `json:"target_groups"`
		DedicatedTestGroup   string             `json:"dedicated_test_group"`
		TestURLs             []string           `json:"test_urls"`
		TestTimeout          int                `json:"test_timeout"`
		ToleranceMs          int                `json:"tolerance_ms"`
		CleanupDays          int                `json:"cleanup_days"`
		MaxConcurrent        int                `json:"max_concurrent"`
		WebPort              int                `json:"web_port"`
		ClashProxyURL        string             `json:"clash_proxy_url"`
		MaxBackoffCycles     int                `json:"max_backoff_cycles"`
		EnableBrowserTest    bool               `json:"enable_browser_test"`
		BrowserTestURLs      []string           `json:"browser_test_urls"`
	}

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		cfg := rover.GetConfig()
		
		if r.Method == "GET" {
			dto := ConfigDTO{
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
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(dto)
			return
		}

		if r.Method == "POST" {
			var dto ConfigDTO
			if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}

			// Apply to a new Config object to write
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
				http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// 立刻熱更新記憶體中的 API 憑證，避免等候 2 秒的檔案掃描
			rover.cfgMutex.Lock()
			rover.cfg = newCfg
			rover.api = NewAPIClient(newCfg.APIUrl, newCfg.APISecret)
			rover.cfgMutex.Unlock()
			rover.checkBrowserTestURLsChanged()
			rover.ForceCheck()


			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/selectors", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			api := rover.GetAPI()
			if api == nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]string{})
				return
			}
			selectors, err := api.GetSelectors()
			if err != nil {
				http.Error(w, fmt.Sprintf("取得 Selector 失敗: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(selectors)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/test-connection", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				APIUrl    string `json:"api_url"`
				APISecret string `json:"api_secret"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			
			testAPI := NewAPIClient(req.APIUrl, req.APISecret)
			if err := testAPI.VerifyConnection(); err != nil {
				http.Error(w, fmt.Sprintf("連線驗證失敗: %v", err), http.StatusBadRequest)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		statMap := rover.GetStatResults()

		highestInGroups := make(map[string][]string)
		for _, groupName := range rover.GetConfig().TargetGroups {
			g, err := rover.GetAPI().GetProxyGroup(groupName)
			if err == nil {
				if g.Now != "" {
					highestInGroups[g.Now] = append(highestInGroups[g.Now], groupName)
				}
			}
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
		
		list := make([]StatNode, 0)
		for _, sc := range statMap {
			isDead := false
			if sc.Err != nil {
				isDead = true
			}
			
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
				BackoffRemaining:        rover.GetBackoffRemaining(sc.Name),
				BrowserBackoffRemaining: rover.GetBrowserBackoffRemaining(sc.Name),
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.URL.Query().Get("node")
		pingHistory, err := db.GetNodeHistory(nodeName, 24)
		browserHistory, _ := db.GetBrowserHistory(nodeName, 24)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ping":    pingHistory,
			"browser": browserHistory,
		})
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		dbSizeMB := 0.0
		if stat, err := os.Stat("rover.db"); err == nil {
			dbSizeMB = float64(stat.Size()) / 1024 / 1024
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_running":    rover.IsRunning.Load(),
			"is_paused":     rover.GetIsPaused(),
			"is_configured": rover.GetConfig().APIUrl != "",
			"api_connected": rover.ApiConnected.Load(),
			"mem_alloc_mb":  float64(m.Alloc) / 1024 / 1024,
			"mem_sys_mb":    float64(m.Sys) / 1024 / 1024,
			"db_size_mb":    dbSizeMB,
			"log_count":     GetLogHistoryCount(),
		})
	})

	http.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		isPaused := rover.TogglePause()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"is_paused": isPaused})
	})

	http.HandleFunc("/api/switch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Group string `json:"group"`
			Node  string `json:"node"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		err := rover.api.SelectProxy(req.Group, req.Node)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logInfo("⚡ 收到 Web UI 手動切換指令：將群組 [%s] 切換至 %s", req.Group, req.Node)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/trigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		select {
		case rover.ManualTrigger <- struct{}{}:
		default:
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wsClientsMutex.Lock()
		wsClients[conn] = true
		wsClientsMutex.Unlock()

		// Send log history to new client
		logHistoryMutex.Lock()
		historyCopy := make([]WebLogEntry, len(logHistory))
		copy(historyCopy, logHistory)
		logHistoryMutex.Unlock()

		conn.WriteJSON(map[string]interface{}{
			"type":    "log_history",
			"history": historyCopy,
		})

		// 讓連線保持開啟直到斷線
		go func() {
			defer func() {
				wsClientsMutex.Lock()
				delete(wsClients, conn)
				wsClientsMutex.Unlock()
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🌐 Web 儀表板已啟動，請訪問: http://127.0.0.1%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Web 伺服器啟動失敗: %v", err)
		
		// 若是在 Windows 且遇到 Port 衝突等錯誤，顯示錯誤對話框提示使用者
		if runtime.GOOS == "windows" {
			errMsg := fmt.Sprintf("Node Rover 啟動失敗！\n\n原因: 可能是 %d Port 已被佔用，或者背景已經有一個 Node Rover 正在執行。\n\n詳細錯誤: %v", port, err)
			exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('%s', 'Node Rover 錯誤', 'OK', 'Error')`, errMsg)).Start()
		}
		
		// 呼叫 systray.Quit() 確保能夠優雅退出，避免在 Windows 留下幽靈 Tray Icon
		systray.Quit()
	}
}


