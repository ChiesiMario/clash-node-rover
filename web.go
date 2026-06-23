package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed templates/index.html
var indexHTML []byte

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsClients      = make(map[*websocket.Conn]bool)
	wsClientsMutex sync.Mutex
)

func BroadcastRefresh() {
	wsClientsMutex.Lock()
	defer wsClientsMutex.Unlock()
	for client := range wsClients {
		err := client.WriteJSON(map[string]string{"type": "refresh"})
		if err != nil {
			client.Close()
			delete(wsClients, client)
		}
	}
}

func BroadcastSingleLog(entry WebLogEntry) {
	wsClientsMutex.Lock()
	defer wsClientsMutex.Unlock()
	msg := map[string]interface{}{
		"type":  "log",
		"entry": entry,
	}
	for client := range wsClients {
		if err := client.WriteJSON(msg); err != nil {
			client.Close()
			delete(wsClients, client)
		}
	}
}

func StartWebServer(db *DB, rover *Rover, port int) {
	http.HandleFunc("/", handleIndex)

	http.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		type GroupStatus struct {
			Name     string   `json:"name"`
			Now      string   `json:"now"`
			Provider string   `json:"provider"`
			All      int      `json:"all_count"`
			AllNodes []string `json:"all_nodes"`
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
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		scores, err := db.GetScores(rover.GetConfig().HistoryDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		highestInGroups := make(map[string][]string)
		for _, groupName := range rover.GetConfig().TargetGroups {
			g, err := rover.GetAPI().GetProxyGroup(groupName)
			if err == nil {
				highestScore := -999999
				highestNode := ""
				for _, name := range g.All {
					if sc, ok := scores[name]; ok && sc.Score > highestScore {
						highestScore = sc.Score
						highestNode = name
					}
				}
				if highestNode != "" {
					highestInGroups[highestNode] = append(highestInGroups[highestNode], groupName)
				}
			}
		}

		type StatNode struct {
			NodeScore
			Provider          string   `json:"provider"`
			HighestInGroups   []string `json:"highest_in_groups"`
			LastInterviewTime int64    `json:"last_interview_time"`
			CooldownMinutes   int      `json:"cooldown_minutes"`
			BackoffRemaining  int      `json:"backoff_remaining"`
		}
		list := make([]StatNode, 0)
		for _, sc := range scores {
			t := rover.GetLastInterviewTime(sc.Name)
			var lastInt int64
			if !t.IsZero() {
				lastInt = t.Unix()
			}
			list = append(list, StatNode{
				NodeScore:         sc,
				Provider:          GetNodeProvider(sc.Name),
				HighestInGroups:   highestInGroups[sc.Name],
				LastInterviewTime: lastInt,
				CooldownMinutes:   rover.GetConfig().ExplorationCooldown,
				BackoffRemaining:  rover.GetBackoffRemaining(sc.Name),
			})
		}

		// Sort by score descending
		for i := 0; i < len(list); i++ {
			for j := i + 1; j < len(list); j++ {
				if list[i].Score < list[j].Score {
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"is_running": rover.IsRunning,
			"is_paused":  rover.GetIsPaused(),
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
		log.Fatalf("Web 伺服器啟動失敗: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}
