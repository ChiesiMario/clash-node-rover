import re

with open("frontend/src/App.tsx", "r", encoding="utf-8") as f:
    app_tsx = f.read()

# Update title font size
app_tsx = re.sub(r'className="app-title"(.*?)font-size: 22px;(.*?)', r'className="app-title md3-title-large"', app_tsx)
app_tsx = re.sub(r'className="app-title"', r'className="app-title md3-title-large"', app_tsx)

# Update log title
app_tsx = re.sub(r'fontWeight:500, marginBottom:\'16px\'', r'marginBottom:\'16px\'', app_tsx)
app_tsx = app_tsx.replace('<div style={{marginBottom:\'16px\'}}>即時系統日誌</div>', '<div className="md3-title-medium" style={{marginBottom:\'16px\'}}>即時系統日誌</div>')
app_tsx = app_tsx.replace('<div style={{fontWeight:500, marginBottom:\'16px\'}}>即時系統日誌</div>', '<div className="md3-title-medium" style={{marginBottom:\'16px\'}}>即時系統日誌</div>')

with open("frontend/src/App.tsx", "w", encoding="utf-8") as f:
    f.write(app_tsx)

print("App.tsx updated.")
