import re

with open('templates/index.html', 'r', encoding='utf-8') as f:
    code = f.read()

# Replace Table Header
code = re.sub(
    r'<th>排名</th>.*?<th>上次面試</th>',
    r'''<th>排名</th>
                                <th>節點名稱</th>
                                <th>綜合分數 (越低越好)</th>
                                <th>平均延遲</th>
                                <th>抖動 (Jitter)</th>
                                <th>群組狀態</th>''',
    code,
    flags=re.DOTALL
)

# Replace Javascript logic for table row
code = re.sub(
    r'const scRate = .*?interviewStr = .*?;',
    r'''
                    let providerTag = node.provider ? '<div class="badge primary" style="margin-top:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">corporate_fare</span> ' + node.provider + '</div>' : '';
                    let groupBadges = '';
                    if (node.highest_in_groups && node.highest_in_groups.length > 0) {
                        groupBadges = node.highest_in_groups.map(g => '<div class="badge success" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">workspace_premium</span> ' + escapeHtml(g) + '</div>').join('');
                    }
                    if (node.backoff_remaining && node.backoff_remaining > 0) {
                        groupBadges += '<div class="badge error" style="margin-top:4px; margin-left:4px; font-size: 10px;"><span class="material-symbols-outlined" style="font-size:12px">timer_off</span> 退避中 (' + node.backoff_remaining + ' 輪)</div>';
                    }
''',
    code,
    flags=re.DOTALL
)

# Replace tr.innerHTML
code = re.sub(
    r'tr\.innerHTML = \'.*?\'<td style="color: var\(--md-sys-color-on-surface-variant\);">\' \+ interviewStr \+ \'</td>\';',
    r'''tr.innerHTML = '<td>#' + (index + 1) + '</td>' +
                        '<td style="font-weight:500;">' + escapeHtml(node.Name) + '<br>' + providerTag + '</td>' +
                        '<td>' + scoreHtml + '</td>' +
                        '<td>' + node.AvgDelay + ' ms</td>' +
                        '<td>' + node.Jitter + ' ms</td>' +
                        '<td>' + groupBadges + '</td>';''',
    code,
    flags=re.DOTALL
)

# Replace toggleChart detailsHtml
code = re.sub(
    r'const speedStr = node\.AvgBandwidth .*?let detailsHtml = .*?\'</div>\';',
    r'''let detailsHtml = '<div style="display:flex; gap:16px; padding:16px; margin-bottom:16px; background:var(--md-sys-color-surface-container); border-radius:16px;">' +
                '<div style="flex:1"><div style="font-size:12px; color:var(--md-sys-color-on-surface-variant);">平均延遲</div><div style="font-size:18px;">' + node.AvgDelay + ' ms</div></div>' +
                '<div style="flex:1"><div style="font-size:12px; color:var(--md-sys-color-on-surface-variant);">抖動 (Jitter)</div><div style="font-size:18px;">' + node.Jitter + ' ms</div></div>' +
                '<div style="flex:1"><div style="font-size:12px; color:var(--md-sys-color-on-surface-variant);">綜合分數</div><div style="font-size:18px;">' + node.Score + '</div></div>' +
            '</div>';''',
    code,
    flags=re.DOTALL
)

# Change <h2 class="card-title"><span class="material-symbols-outlined">leaderboard</span> 節點排行榜 (EMA)</h2>
code = re.sub(
    r'<h2 class="card-title"><span class="material-symbols-outlined">leaderboard</span> 節點排行榜 \(EMA\)</h2>',
    r'<h2 class="card-title"><span class="material-symbols-outlined">leaderboard</span> 節點排行榜 (即時)</h2>',
    code
)


with open('templates/index.html', 'w', encoding='utf-8') as f:
    f.write(code)
