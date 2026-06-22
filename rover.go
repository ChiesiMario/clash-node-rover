package main

import (
	"log"
	"math"
	"sync"
	"time"
)

type Rover struct {
	cfg *Config
	api *APIClient
}

func NewRover(cfg *Config, api *APIClient) *Rover {
	return &Rover{
		cfg: cfg,
		api: api,
	}
}

func (r *Rover) Run() {
	log.Printf("開始監控 Node Rover，目標群組 '%s'，每 %v 檢查一次", r.cfg.TargetGroup, r.cfg.CheckInterval)
	ticker := time.NewTicker(r.cfg.CheckInterval)
	defer ticker.Stop()

	// initial run
	r.checkAndSwitch()

	for range ticker.C {
		r.checkAndSwitch()
	}
}

func (r *Rover) checkAndSwitch() {
	log.Println("--- 檢查節點 ---")
	group, err := r.api.GetProxyGroup(r.cfg.TargetGroup)
	if err != nil {
		log.Printf("取得代理群組時發生錯誤: %v", err)
		return
	}

	if len(group.All) == 0 {
		log.Println("目標群組中沒有找到節點。")
		return
	}

	type nodeStat struct {
		Name  string
		Delay int
		Err   error
	}

	var wg sync.WaitGroup
	stats := make([]nodeStat, len(group.All))

	for i, nodeName := range group.All {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			delay, err := r.api.TestProxyDelay(name, r.cfg.TestURL, r.cfg.TestTimeout)
			stats[idx] = nodeStat{Name: name, Delay: delay, Err: err}
		}(i, nodeName)
	}

	wg.Wait()

	bestNode := ""
	bestDelay := math.MaxInt32

	for _, s := range stats {
		if s.Err != nil {
			log.Printf("節點 [%s] 測試失敗: %v", s.Name, s.Err)
		} else {
			log.Printf("節點 [%s] 延遲: %d 毫秒", s.Name, s.Delay)
			if s.Delay > 0 && s.Delay < bestDelay {
				bestDelay = s.Delay
				bestNode = s.Name
			}
		}
	}

	if bestNode == "" {
		log.Println("所有節點測試失敗，保留目前節點。")
		return
	}

	if bestNode != group.Now {
		log.Printf("從 [%s] 切換至更好的節點 [%s] (延遲: %d 毫秒)", group.Now, bestNode, bestDelay)
		if err := r.api.SelectProxy(r.cfg.TargetGroup, bestNode); err != nil {
			log.Printf("切換代理節點失敗: %v", err)
		} else {
			log.Println("成功切換代理節點。")
		}
	} else {
		log.Printf("目前的節點 [%s] 依然是最好的，無需切換。", group.Now)
	}
}
