import re

with open('templates/index.html', 'r', encoding='utf-8') as f:
    code = f.read()

# Add Filter Modal HTML
filter_modal_html = """
    <!-- Filter Modal -->
    <div id="filterModal" class="modal">
        <div class="modal-content">
            <h2 id="filterGroupTitle" style="margin-top:0">過濾器設定</h2>
            <input type="hidden" id="filterGroupName">
            
            <div style="margin-bottom: 16px;">
                <label style="display:block; margin-bottom: 8px; font-weight: 500;">節點名稱匹配 (支援正則或以 | 分隔)</label>
                <input type="text" id="filterKeyword" placeholder="例如: US|United States|美國" class="input-field" style="width: 100%; box-sizing: border-box; padding: 12px; border-radius: 8px; border: 1px solid var(--md-sys-color-outline); background: var(--md-sys-color-surface-container-high); color: var(--md-sys-color-on-surface);">
                <div style="font-size: 12px; color: var(--md-sys-color-on-surface-variant); margin-top: 4px;">留空代表允許所有節點。支援正則表達式，例如 <code>美國|香港</code>。</div>
            </div>

            <div style="margin-bottom: 16px;">
                <label style="display:flex; align-items:center; gap:8px; cursor:pointer;">
                    <input type="checkbox" id="filterChatGPT" style="width:18px; height:18px;">
                    <span style="font-weight: 500;">確保 ChatGPT 服務可用</span>
                </label>
                <div style="font-size: 12px; color: var(--md-sys-color-on-surface-variant); margin-left: 26px;">自動開啟網頁並檢查是否遇到 Access Denied / 地區限制。</div>
            </div>

            <div style="margin-bottom: 24px;">
                <label style="display:flex; align-items:center; gap:8px; cursor:pointer;">
                    <input type="checkbox" id="filterGemini" style="width:18px; height:18px;">
                    <span style="font-weight: 500;">確保 Gemini 服務可用</span>
                </label>
                <div style="font-size: 12px; color: var(--md-sys-color-on-surface-variant); margin-left: 26px;">自動開啟網頁並檢查是否遇到「未在該地區推出」等限制。</div>
            </div>

            <div style="display:flex; justify-content: flex-end; gap: 12px;">
                <button class="btn secondary" onclick="closeFilterModal()">取消</button>
                <button class="btn primary" onclick="saveFilterModal()">儲存設定</button>
            </div>
        </div>
    </div>
"""
if "id=\"filterModal\"" not in code:
    code = code.replace("</body>", filter_modal_html + "\n</body>")

# Replace group card generation
find_card = """                    let options = g.all_nodes.map(n => '<option value="' + n + '" ' + (n === g.now ? 'selected' : '') + '>' + n + '</option>').join('');
                    let lockBtnClass = g.locked ? 'btn icon-btn' : 'btn secondary icon-btn';
                    let lockIcon = g.locked ? 'lock' : 'lock_open';
                    let lockTitle = g.locked ? '點擊解鎖 (恢復自動切換)' : '點擊鎖定 (停止自動切換)';
                    
                    html += '<div class="group-card">' +
                            '<div class="group-header" style="display:flex; justify-content:space-between; align-items:center;">' + 
                                '<span>' + g.name + '</span>' +
                                '<button class="' + lockBtnClass + '" onclick="toggleGroupLock(\\'' + g.name + '\\', ' + !g.locked + ')" title="' + lockTitle + '" style="width:32px; height:32px;"><span class="material-symbols-outlined" style="font-size:16px">' + lockIcon + '</span></button>' +
                            '</div>' +
                            '<div class="group-now">' + (g.now || '未選擇') + '</div>' +
                            (g.provider ? '<div class="badge primary" style="align-self: flex-start;"><span class="material-symbols-outlined" style="font-size:14px">corporate_fare</span> ' + g.provider + '</div>' : '') +
                            '<div style="color: var(--md-sys-color-on-surface-variant); font-size:14px; margin-bottom:8px;">運行中 &bull; 共 ' + g.all_count + ' 個節點</div>' +
                            '<div style="display: flex; gap: 8px;">' +
                                '<select id="select-' + g.name + '" style="flex:1; background: var(--md-sys-color-surface-container-high); color: var(--md-sys-color-on-surface); border: 1px solid var(--md-sys-color-outline); border-radius: 8px; padding: 8px;">' + options + '</select>' +
                                '<button onclick="manualSwitch(\\'' + g.name + '\\')" class="btn" style="padding: 8px 16px;">切換</button>' +
                            '</div>' +
                        '</div>';"""

replace_card = """                    let options = g.all_nodes.map(n => '<option value="' + n + '" ' + (n === g.now ? 'selected' : '') + '>' + n + '</option>').join('');
                    let lockBtnClass = g.locked ? 'btn icon-btn' : 'btn secondary icon-btn';
                    let lockIcon = g.locked ? 'lock' : 'lock_open';
                    let lockTitle = g.locked ? '點擊解鎖 (恢復自動切換)' : '點擊鎖定 (停止自動切換)';
                    
                    let hasFilter = g.filter && (g.filter.keyword_regex || g.filter.check_chatgpt || g.filter.check_gemini);
                    let filterBtnClass = hasFilter ? 'btn primary icon-btn' : 'btn secondary icon-btn';
                    let filterBadges = '';
                    if (hasFilter) {
                        filterBadges += '<div style="display:flex; gap:4px; flex-wrap:wrap; margin-top:8px;">';
                        if (g.filter.keyword_regex) filterBadges += '<div class="badge primary" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">filter_alt</span> ' + escapeHtml(g.filter.keyword_regex) + '</div>';
                        if (g.filter.check_chatgpt) filterBadges += '<div class="badge success" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">smart_toy</span> ChatGPT</div>';
                        if (g.filter.check_gemini) filterBadges += '<div class="badge success" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">auto_awesome</span> Gemini</div>';
                        filterBadges += '</div>';
                    }

                    html += '<div class="group-card">' +
                            '<div class="group-header" style="display:flex; justify-content:space-between; align-items:center;">' + 
                                '<span>' + g.name + '</span>' +
                                '<div style="display:flex;">' +
                                    '<button class="' + filterBtnClass + '" onclick="openFilterModal(\\'' + escapeHtml(g.name) + '\\')" title="設定過濾規則" style="width:32px; height:32px; margin-right:8px;"><span class="material-symbols-outlined" style="font-size:16px">tune</span></button>' +
                                    '<button class="' + lockBtnClass + '" onclick="toggleGroupLock(\\'' + escapeHtml(g.name) + '\\', ' + !g.locked + ')" title="' + lockTitle + '" style="width:32px; height:32px;"><span class="material-symbols-outlined" style="font-size:16px">' + lockIcon + '</span></button>' +
                                '</div>' +
                            '</div>' +
                            '<div class="group-now">' + (g.now || '未選擇') + '</div>' +
                            (g.provider ? '<div class="badge primary" style="align-self: flex-start;"><span class="material-symbols-outlined" style="font-size:14px">corporate_fare</span> ' + g.provider + '</div>' : '') +
                            filterBadges +
                            '<div style="color: var(--md-sys-color-on-surface-variant); font-size:14px; margin-top:8px; margin-bottom:8px;">運行中 &bull; 共 ' + g.all_count + ' 個節點</div>' +
                            '<div style="display: flex; gap: 8px;">' +
                                '<select id="select-' + g.name + '" style="flex:1; background: var(--md-sys-color-surface-container-high); color: var(--md-sys-color-on-surface); border: 1px solid var(--md-sys-color-outline); border-radius: 8px; padding: 8px;">' + options + '</select>' +
                                '<button onclick="manualSwitch(\\'' + g.name + '\\')" class="btn" style="padding: 8px 16px;">切換</button>' +
                            '</div>' +
                        '</div>';"""

code = code.replace(find_card, replace_card)

# Add Javascript functions
js_functions = """
        window.openFilterModal = async function(groupName) {
            document.getElementById('filterGroupTitle').innerText = '過濾器: ' + groupName;
            document.getElementById('filterGroupName').value = groupName;
            try {
                const res = await fetch('/api/groups/filter?group=' + encodeURIComponent(groupName));
                const data = await res.json();
                document.getElementById('filterKeyword').value = data.keyword_regex || '';
                document.getElementById('filterChatGPT').checked = data.check_chatgpt || false;
                document.getElementById('filterGemini').checked = data.check_gemini || false;
            } catch(e) {}
            document.getElementById('filterModal').classList.add('active');
        }

        window.closeFilterModal = function() {
            document.getElementById('filterModal').classList.remove('active');
        }

        window.saveFilterModal = async function() {
            const groupName = document.getElementById('filterGroupName').value;
            const data = {
                keyword_regex: document.getElementById('filterKeyword').value,
                check_chatgpt: document.getElementById('filterChatGPT').checked,
                check_gemini: document.getElementById('filterGemini').checked
            };
            await fetch('/api/groups/filter?group=' + encodeURIComponent(groupName), {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            closeFilterModal();
            fetchGroups();
        }
"""
if "window.openFilterModal" not in code:
    code = code.replace("function closeNodeModal() {", js_functions + "\n        function closeNodeModal() {")

with open('templates/index.html', 'w', encoding='utf-8') as f:
    f.write(code)
