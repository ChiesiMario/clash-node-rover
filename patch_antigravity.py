import re

def patch_rover():
    with open('rover.go', 'r', encoding='utf-8') as f:
        code = f.read()

    # Update GroupFilter struct
    code = code.replace(
"""	CheckChatGPT bool   `json:"check_chatgpt"`
	CheckGemini  bool   `json:"check_gemini"`
}""",
"""	CheckChatGPT     bool   `json:"check_chatgpt"`
	CheckGemini      bool   `json:"check_gemini"`
	CheckAntigravity bool   `json:"check_antigravity"`
}""")

    # Update URL append
    code = code.replace(
"""				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}""",
"""				if filter.CheckGemini {
					testURLs = append(testURLs, "https://gemini.google.com/app")
				}
				if filter.CheckAntigravity {
					testURLs = append(testURLs, "https://generativelanguage.googleapis.com")
				}""")

    # Update Blocked logic
    find_block = """						} else if strings.Contains(targetURL, "gemini.google.com") {
							if strings.Contains(lowerText, "isn't supported in your country") || strings.Contains(lowerText, "未在該地區推出") || strings.Contains(lowerText, "not available") {
								blocked = true
							}
						}"""
    
    replace_block = """						} else if strings.Contains(targetURL, "gemini.google.com") {
							if strings.Contains(lowerText, "isn't supported in your country") || strings.Contains(lowerText, "未在該地區推出") || strings.Contains(lowerText, "not available") {
								blocked = true
							}
						} else if strings.Contains(targetURL, "generativelanguage.googleapis") {
							if strings.Contains(lowerText, "user location is not supported") {
								blocked = true
							}
						}"""
    code = code.replace(find_block, replace_block)
    
    with open('rover.go', 'w', encoding='utf-8') as f:
        f.write(code)


def patch_web():
    with open('web.go', 'r', encoding='utf-8') as f:
        code = f.read()

    code = code.replace(
"""			val = `{"keyword_regex": "", "check_chatgpt": false, "check_gemini": false}`""",
"""			val = `{"keyword_regex": "", "check_chatgpt": false, "check_gemini": false, "check_antigravity": false}`""")

    code = code.replace(
"""				CheckChatGPT bool   `json:"check_chatgpt"`
				CheckGemini  bool   `json:"check_gemini"`
			}""",
"""				CheckChatGPT     bool   `json:"check_chatgpt"`
				CheckGemini      bool   `json:"check_gemini"`
				CheckAntigravity bool   `json:"check_antigravity"`
			}""")

    with open('web.go', 'w', encoding='utf-8') as f:
        f.write(code)


def patch_html():
    with open('templates/index.html', 'r', encoding='utf-8') as f:
        code = f.read()

    # Update badge checks
    find_badge = """                        if (g.filter.check_gemini) filterBadges += '<div class="badge success" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">auto_awesome</span> Gemini</div>';
                        filterBadges += '</div>';"""
    replace_badge = """                        if (g.filter.check_gemini) filterBadges += '<div class="badge success" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">auto_awesome</span> Gemini</div>';
                        if (g.filter.check_antigravity) filterBadges += '<div class="badge success" style="font-size:10px;"><span class="material-symbols-outlined" style="font-size:12px">rocket_launch</span> Antigravity</div>';
                        filterBadges += '</div>';"""
    code = code.replace(find_badge, replace_badge)

    find_has = """                    let hasFilter = g.filter && (g.filter.keyword_regex || g.filter.check_chatgpt || g.filter.check_gemini);"""
    replace_has = """                    let hasFilter = g.filter && (g.filter.keyword_regex || g.filter.check_chatgpt || g.filter.check_gemini || g.filter.check_antigravity);"""
    code = code.replace(find_has, replace_has)

    # Add boolean mapping
    find_bools = """                    const isGemini = (g.filter && g.filter.check_gemini) ? "checked" : "";"""
    replace_bools = """                    const isGemini = (g.filter && g.filter.check_gemini) ? "checked" : "";
                    const isAntigravity = (g.filter && g.filter.check_antigravity) ? "checked" : "";"""
    code = code.replace(find_bools, replace_bools)

    # Add Checkbox
    find_cb = """                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" id="filter-gemini-' + safeGroupName + '" ' + isGemini + '> ✨ Gemini</label>' +
                                '</div>' +"""
    replace_cb = """                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" id="filter-gemini-' + safeGroupName + '" ' + isGemini + '> ✨ Gemini</label>' +
                                    '<label style="cursor:pointer; display:flex; align-items:center; gap:4px;"><input type="checkbox" onchange="inlineSaveFilter(\\'' + safeGroupName + '\\')" id="filter-antigravity-' + safeGroupName + '" ' + isAntigravity + '> 🚀 Antigravity</label>' +
                                '</div>' +"""
    code = code.replace(find_cb, replace_cb)

    # Add to inlineSaveFilter
    find_js = """            const geminiChecked = document.getElementById('filter-gemini-' + groupName).checked;

            const data = {
                keyword_regex: finalRegex,
                check_chatgpt: chatgptChecked,
                check_gemini: geminiChecked
            };"""
    replace_js = """            const geminiChecked = document.getElementById('filter-gemini-' + groupName).checked;
            const antigravityChecked = document.getElementById('filter-antigravity-' + groupName).checked;

            const data = {
                keyword_regex: finalRegex,
                check_chatgpt: chatgptChecked,
                check_gemini: geminiChecked,
                check_antigravity: antigravityChecked
            };"""
    code = code.replace(find_js, replace_js)

    with open('templates/index.html', 'w', encoding='utf-8') as f:
        f.write(code)

patch_rover()
patch_web()
patch_html()
