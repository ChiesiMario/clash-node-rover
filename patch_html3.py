import re

with open('templates/index.html', 'r', encoding='utf-8') as f:
    code = f.read()

# Remove filterModal
if "id=\"filterModal\"" in code:
    code = re.sub(r'<!-- Filter Modal -->.*?</div>\s*</div>\s*</div>', '', code, flags=re.DOTALL)

# Update group card generation to include inline filters
find_card = """                    let hasFilter = g.filter && (g.filter.keyword_regex || g.filter.check_chatgpt || g.filter.check_gemini);
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

replace_card = """                    let lockBtnClass = g.locked ? 'btn icon-btn' : 'btn secondary icon-btn';
                    let lockIcon = g.locked ? 'lock' : 'lock_open';
                    let lockTitle = g.locked ? '點擊解鎖 (恢復自動切換)' : '點擊鎖定 (停止自動切換)';
                    
                    const safeGroupName = escapeHtml(g.name);
                    const rx = (g.filter && g.filter.keyword_regex) ? g.filter.keyword_regex : "";
                    
                    const isUS = rx.includes("US|");
                    const isHK = rx.includes("HK|");
                    const isTW = rx.includes("TW|");
                    const isJP = rx.includes("JP|");
                    const isSG = rx.includes("SG|");
                    const isUK = rx.includes("UK|");

                    const isChatGPT = (g.filter && g.filter.check_chatgpt) ? "checked" : "";
                    const isGemini = (g.filter && g.filter.check_gemini) ? "checked" : "";

                    html += '<div class="group-card">' +
                            '<div class="group-header" style="display:flex; justify-content:space-between; align-items:center;">' + 
                                '<span>' + g.name + '</span>' +
                                '<button class="' + lockBtnClass + '" onclick="toggleGroupLock(\\'' + safeGroupName + '\\', ' + !g.locked + ')" title="' + lockTitle + '" style="width:32px; height:32px;"><span class="material-symbols-outlined" style="font-size:16px">' + lockIcon + '</span></button>' +
                            '</div>' +
                            '<div class="group-now">' + (g.now || '未選擇') + '</div>' +
                            (g.provider ? '<div class="badge primary" style="align-self: flex-start;"><span class="material-symbols-outlined" style="font-size:14px">corporate_fare</span> ' + g.provider + '</div>' : '') +
                            '<div style="color: var(--md-sys-color-on-surface-variant); font-size:14px; margin-top:8px; margin-bottom:8px;">運行中 &bull; 共 ' + g.all_count + ' 個節點</div>' +
                            '<div style="display: flex; gap: 8px;">' +
                                '<select id="select-' + g.name + '" style="flex:1; background: var(--md-sys-color-surface-container-high); color: var(--md-sys-color-on-surface); border: 1px solid var(--md-sys-color-outline); border-radius: 8px; padding: 8px;">' + options + '</select>' +
                                '<button onclick="manualSwitch(\\'' + safeGroupName + '\\')" class="btn" style="padding: 8px 16px;">切換</button>' +
                            '</div>' +
                            
                            // Inline Filter Section
                            '<div style="margin-top: 16px; padding-top: 12px; border-top: 1px solid var(--md-sys-color-outline-variant);">' +
                                '<div style="font-size: 13px; font-weight: 500; margin-bottom: 8px; color: var(--md-sys-color-primary);">節點地區篩選</div>' +
                                '<div style="display: flex; gap: 12px; flex-wrap: wrap; font-size: 13px;">' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="US" class="filter-region-' + safeGroupName + '" ' + (isUS?'checked':'') + '> 🇺🇸 美國</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="HK" class="filter-region-' + safeGroupName + '" ' + (isHK?'checked':'') + '> 🇭🇰 香港</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="TW" class="filter-region-' + safeGroupName + '" ' + (isTW?'checked':'') + '> 🇹🇼 台灣</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="JP" class="filter-region-' + safeGroupName + '" ' + (isJP?'checked':'') + '> 🇯🇵 日本</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="SG" class="filter-region-' + safeGroupName + '" ' + (isSG?'checked':'') + '> 🇸🇬 新加坡</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" value="UK" class="filter-region-' + safeGroupName + '" ' + (isUK?'checked':'') + '> 🇬🇧 英國</label>' +
                                '</div>' +
                                '<div style="font-size: 13px; font-weight: 500; margin-top: 12px; margin-bottom: 8px; color: var(--md-sys-color-primary);">必備服務驗證</div>' +
                                '<div style="display: flex; gap: 12px; font-size: 13px;">' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" id="filter-chatgpt-' + safeGroupName + '" ' + isChatGPT + '> 🤖 ChatGPT</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" id="filter-gemini-' + safeGroupName + '" ' + isGemini + '> ✨ Gemini</label>' +
                                '</div>' +
                            '</div>' +
                        '</div>';"""

code = code.replace(find_card, replace_card)

# Update Javascript functions
find_js = """        window.openFilterModal = async function(groupName) {"""
replace_js = """        const REGION_PRESETS = {
            'US': 'US|United States|us|美國|美国',
            'HK': 'HK|Hong Kong|香港',
            'TW': 'TW|Taiwan|台灣|台湾|臺',
            'JP': 'JP|Japan|日本',
            'SG': 'SG|Singapore|新加坡|狮城',
            'UK': 'UK|United Kingdom|英國|英国'
        };

        window.inlineSaveFilter = async function(groupName) {
            const regionCheckboxes = document.querySelectorAll('.filter-region-' + CSS.escape(groupName));
            let selectedRegexes = [];
            regionCheckboxes.forEach(cb => {
                if(cb.checked && REGION_PRESETS[cb.value]) {
                    selectedRegexes.push(REGION_PRESETS[cb.value]);
                }
            });
            const finalRegex = selectedRegexes.join('|');
            
            const chatgptChecked = document.getElementById('filter-chatgpt-' + groupName).checked;
            const geminiChecked = document.getElementById('filter-gemini-' + groupName).checked;

            const data = {
                keyword_regex: finalRegex,
                check_chatgpt: chatgptChecked,
                check_gemini: geminiChecked
            };
            await fetch('/api/groups/filter?group=' + encodeURIComponent(groupName), {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
        }
        
        // Remove old modal functions
        window.openFilterModal = async function(groupName) {"""

code = code.replace(find_js, replace_js)

# Delete the rest of openFilterModal, closeFilterModal, saveFilterModal if we want, or just leave them unused.
# Let's remove them properly.
js_functions_to_remove = """        window.openFilterModal = async function(groupName) {
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
        }"""
code = code.replace(js_functions_to_remove, "")

with open('templates/index.html', 'w', encoding='utf-8') as f:
    f.write(code)
