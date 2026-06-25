import os
import re

with open("frontend/src/components/NodeRanking.tsx", "r", encoding="utf-8") as f:
    content = f.read()

# Replace header
content = content.replace('<th>分發狀態</th>', '<th>分發狀態</th>\\n                        <th style={{textAlign: \\\'center\\\'}}>服務可用性</th>')

# Function to get service status
service_fn = """
    const getServiceStatus = (node: any, url: string) => {
        if (!node.browser_backoff_remaining) return "unknown";
        if (node.browser_backoff_remaining[url] === undefined) return "unknown";
        if (node.browser_backoff_remaining[url] > 0) return "fail";
        return "ok";
    };

    const renderServiceBadge = (node: any, name: string, url: string) => {
        const status = getServiceStatus(node, url);
        if (status === "unknown") {
            return <span key={name} className="hig-badge" style={{backgroundColor: 'var(--hig-bg-tertiary)', color: 'var(--hig-text-secondary)'}} title="未測試或未知">{name}: ?</span>;
        } else if (status === "ok") {
            return <span key={name} className="hig-badge green" title="驗證通過">{name}: OK</span>;
        } else {
            return <span key={name} className="hig-badge red" title="驗證失敗">{name}: ERR</span>;
        }
    };
"""

content = content.replace('export default function NodeRanking({ stats }: any) {\\n    const { fetchNodeHistory } = useApi();', 'export default function NodeRanking({ stats }: any) {\\n    const { fetchNodeHistory } = useApi();\\n' + service_fn)

# Replace table body row
td_content = """
                                    <td>
                                        <div style={{display:'flex', gap:'4px', flexWrap:'wrap', justifyContent: 'center'}}>
                                            {renderServiceBadge(s, "GPT", "https://chatgpt.com")}
                                            {renderServiceBadge(s, "Gem", "https://gemini.google.com/app")}
                                            {renderServiceBadge(s, "Anti", "https://generativelanguage.googleapis.com/v1beta/models")}
                                        </div>
                                    </td>
"""
content = content.replace('                                    </td>\\n                                </tr>', '                                    </td>\\n' + td_content + '                                </tr>')
content = content.replace('colSpan={6}', 'colSpan={7}')

with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f:
    f.write(content)
