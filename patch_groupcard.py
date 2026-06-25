import re

with open("frontend/src/components/GroupCard.tsx", "r", encoding="utf-8") as f:
    code = f.read()

# Update select to md3-select
code = code.replace('<select id={`select-${group.name}`} defaultValue={group.now} style={{flex:1, background: \'var(--md-sys-color-surface-container-high)\', color: \'var(--md-sys-color-on-surface)\', border: \'1px solid var(--md-sys-color-outline)\', borderRadius: \'8px\', padding: \'8px\'}}>',
                    '<select className="md3-select" id={`select-${group.name}`} defaultValue={group.now} style={{flex:1}}>')

# Update typography
code = code.replace('className="group-header" style={{display:\'flex\'', 'className="group-header md3-title-medium" style={{display:\'flex\'')
code = code.replace('<div className="group-now">', '<div className="group-now md3-title-large" style={{color:\'var(--md-sys-color-primary)\'}}>')
code = code.replace('<div style={{color: \'var(--md-sys-color-on-surface-variant)\', fontSize:\'14px\',', '<div className="md3-body-medium" style={{color: \'var(--md-sys-color-on-surface-variant)\',')

# Update labels for checkbox
code = code.replace('<label key={r} style={{cursor:\'pointer\', display:\'flex\', alignItems:\'center\', gap:\'4px\'}}>', '<label key={r} className="md3-checkbox-label">')
code = code.replace('<label style={{cursor:\'pointer\', display:\'flex\', alignItems:\'center\', gap:\'4px\'}}>', '<label className="md3-checkbox-label">')

with open("frontend/src/components/GroupCard.tsx", "w", encoding="utf-8") as f:
    f.write(code)

print("GroupCard.tsx updated.")
