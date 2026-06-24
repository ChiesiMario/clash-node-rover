import re

# Update rover.go
with open('rover.go', 'r', encoding='utf-8') as f:
    rover_code = f.read()

find_rover_backoff = """func (r *NodeRover) GetBackoffRemaining(node string) int {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.backoffRemaining[node]
}"""

replace_rover_backoff = """func (r *NodeRover) GetBackoffRemaining(node string) int {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.backoffRemaining[node]
}

func (r *NodeRover) GetBrowserBackoffRemaining(name string) map[string]int {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	
	if r.browserBackoffRemaining[name] == nil {
		return nil
	}
	
	res := make(map[string]int)
	for k, v := range r.browserBackoffRemaining[name] {
		res[k] = v
	}
	return res
}"""
rover_code = rover_code.replace(find_rover_backoff, replace_rover_backoff)

with open('rover.go', 'w', encoding='utf-8') as f:
    f.write(rover_code)

# Update web.go
with open('web.go', 'r', encoding='utf-8') as f:
    web_code = f.read()

find_stat_node = """		type StatNode struct {
			Name              string   `json:"Name"`
			AvgDelay          int      `json:"AvgDelay"`
			Jitter            int      `json:"Jitter"`
			Score             int      `json:"Score"`
			Provider          string   `json:"provider"`
			HighestInGroups   []string `json:"highest_in_groups"`
			BackoffRemaining  int      `json:"backoff_remaining"`
		}
		
		list := make([]StatNode, 0)
		for _, sc := range statMap {
			if sc.Err != nil {
				continue
			}
			list = append(list, StatNode{
				Name:              sc.Name,
				AvgDelay:          sc.AvgDelay,
				Jitter:            sc.Jitter,
				Score:             sc.Score,
				Provider:          GetNodeProvider(sc.Name),
				HighestInGroups:   highestInGroups[sc.Name],
				BackoffRemaining:  rover.GetBackoffRemaining(sc.Name),
			})
		}"""

replace_stat_node = """		type StatNode struct {
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
		}"""
web_code = web_code.replace(find_stat_node, replace_stat_node)

with open('web.go', 'w', encoding='utf-8') as f:
    f.write(web_code)

# Update index.html
with open('templates/index.html', 'r', encoding='utf-8') as f:
    html_code = f.read()

find_render = """                    let scoreHtml = '<span class="score-box">' + node.Score + '</span>';
                    if (!isNewRow && window.previousScores[node.Name] !== undefined) {
                        const oldScore = window.previousScores[node.Name];
                        if (node.Score > oldScore) scoreHtml = '<span class="score-box flash-green">' + node.Score + '</span>';
                        else if (node.Score < oldScore) scoreHtml = '<span class="score-box flash-red">' + node.Score + '</span>';
                    }
                    window.previousScores[node.Name] = node.Score;
                    
                    
                    let providerTag = node.provider ? '<div class="badge primary" style="margin-top:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">corporate_fare</span> ' + node.provider + '</div>' : '';
                    let groupBadges = '';
                    if (node.highest_in_groups && node.highest_in_groups.length > 0) {
                        groupBadges = node.highest_in_groups.map(g => '<div class="badge success" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">workspace_premium</span> ' + escapeHtml(g) + '</div>').join('');
                    }
                    if (node.backoff_remaining && node.backoff_remaining > 0) {
                        groupBadges += '<div class="badge error" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">timer_off</span> 退避中 (' + node.backoff_remaining + ' 輪)</div>';
                    }

                    if (node.last_interview_time > 0) {
                        const remainMin = node.cooldown_minutes - (Math.floor(Date.now() / 60000) - Math.floor(node.last_interview_time / 60));
                        interviewStr = remainMin <= 0 ? timeAgo(node.last_interview_time) + ' <span style="color:var(--md-sys-color-success)">(OK)</span>' : timeAgo(node.last_interview_time) + ' <span style="color:var(--md-sys-color-warning)">(' + remainMin + 'm)</span>';
                    }

                    tr.innerHTML = '<td>#' + (index + 1) + '</td>' +
                        '<td style="font-weight:500;">' + escapeHtml(node.Name) + '<br>' + providerTag + '</td>' +
                        '<td>' + scoreHtml + '</td>' +
                        '<td>' + node.AvgDelay + ' ms</td>' +
                        '<td>' + node.Jitter + ' ms</td>' +
                        '<td>' + groupBadges + '</td>';"""

replace_render = """                    let scoreHtml = '';
                    if (node.is_dead) {
                        scoreHtml = '<span class="score-box" style="background:var(--md-sys-color-error-container); color:var(--md-sys-color-on-error-container);">失敗</span>';
                    } else {
                        scoreHtml = '<span class="score-box">' + node.Score + '</span>';
                        if (!isNewRow && window.previousScores[node.Name] !== undefined) {
                            const oldScore = window.previousScores[node.Name];
                            if (node.Score > oldScore) scoreHtml = '<span class="score-box flash-green">' + node.Score + '</span>';
                            else if (node.Score < oldScore) scoreHtml = '<span class="score-box flash-red">' + node.Score + '</span>';
                        }
                    }
                    window.previousScores[node.Name] = node.Score;
                    
                    let providerTag = node.provider ? '<div class="badge primary" style="margin-top:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">corporate_fare</span> ' + node.provider + '</div>' : '';
                    let groupBadges = '';
                    if (node.highest_in_groups && node.highest_in_groups.length > 0) {
                        groupBadges = node.highest_in_groups.map(g => '<div class="badge success" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">workspace_premium</span> ' + escapeHtml(g) + '</div>').join('');
                    }
                    if (node.backoff_remaining && node.backoff_remaining > 0) {
                        groupBadges += '<div class="badge error" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">timer_off</span> Ping退避 (' + node.backoff_remaining + ' 輪)</div>';
                    }
                    if (node.browser_backoff_remaining) {
                        for (const [url, rem] of Object.entries(node.browser_backoff_remaining)) {
                            if (rem > 0) {
                                let sName = url;
                                if (url.includes("chatgpt")) sName = "ChatGPT";
                                else if (url.includes("gemini")) sName = "Gemini";
                                else if (url.includes("generative")) sName = "Antigravity";
                                else sName = "網頁";
                                groupBadges += '<div class="badge warning" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">web_asset_off</span> ' + sName + '退避 (' + rem + ' 輪)</div>';
                            }
                        }
                    }

                    if (node.last_interview_time > 0) {
                        const remainMin = node.cooldown_minutes - (Math.floor(Date.now() / 60000) - Math.floor(node.last_interview_time / 60));
                        interviewStr = remainMin <= 0 ? timeAgo(node.last_interview_time) + ' <span style="color:var(--md-sys-color-success)">(OK)</span>' : timeAgo(node.last_interview_time) + ' <span style="color:var(--md-sys-color-warning)">(' + remainMin + 'm)</span>';
                    }
                    
                    let delayHtml = node.is_dead ? '<span style="color:var(--md-sys-color-outline)">N/A</span>' : (node.AvgDelay + ' ms');
                    let jitterHtml = node.is_dead ? '<span style="color:var(--md-sys-color-outline)">N/A</span>' : (node.Jitter + ' ms');

                    tr.innerHTML = '<td>#' + (index + 1) + '</td>' +
                        '<td style="font-weight:500; color:' + (node.is_dead ? 'var(--md-sys-color-outline)' : 'inherit') + ';">' + escapeHtml(node.Name) + '<br>' + providerTag + '</td>' +
                        '<td>' + scoreHtml + '</td>' +
                        '<td>' + delayHtml + '</td>' +
                        '<td>' + jitterHtml + '</td>' +
                        '<td>' + groupBadges + '</td>';"""
html_code = html_code.replace(find_render, replace_render)

with open('templates/index.html', 'w', encoding='utf-8') as f:
    f.write(html_code)
