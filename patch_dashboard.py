import re

with open("frontend/src/components/Dashboard.tsx", "r", encoding="utf-8") as f:
    code = f.read()

code = code.replace('<div style={{fontSize:\'24px\', fontWeight:500, marginBottom:\'8px\'}}>', '<div className="md3-title-large" style={{marginBottom:\'8px\'}}>')
code = code.replace('fontSize:\'14px\', fontWeight:\'normal\'', '')
code = code.replace('<span style={{ padding:\'4px 8px\'', '<span className="md3-label-medium" style={{ padding:\'4px 8px\'')
code = code.replace('<div style={{color:\'var(--md-sys-color-on-surface-variant)\'}}>', '<div className="md3-body-medium" style={{color:\'var(--md-sys-color-on-surface-variant)\'}}>')

with open("frontend/src/components/Dashboard.tsx", "w", encoding="utf-8") as f:
    f.write(code)

print("Dashboard.tsx updated.")
