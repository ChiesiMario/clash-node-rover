import re

# 1. App.tsx
with open("frontend/src/App.tsx", "r", encoding="utf-8") as f:
    app_tsx = f.read()

# Change top app bar layout
app_tsx = app_tsx.replace(
    '<div className="container">\n            <div className="top-app-bar" style={{marginBottom: \'24px\', borderRadius: \'16px\'}}>',
    '<>\n            <div className="top-app-bar" style={{marginBottom: \'24px\', boxShadow: \'var(--md-sys-elevation-1)\'}}>'
)
app_tsx = app_tsx.replace('                </button>\n            </div>', '                </button>\n            </div>\n\n            <div className="container">')
app_tsx = app_tsx.replace('        </div>\n    );\n}', '        </div>\n        </>\n    );\n}')

with open("frontend/src/App.tsx", "w", encoding="utf-8") as f:
    f.write(app_tsx)


# 2. GroupCard.tsx
with open("frontend/src/components/GroupCard.tsx", "r", encoding="utf-8") as f:
    gc_tsx = f.read()

# Fix select and button
gc_tsx = gc_tsx.replace(
    'style={{flex:1}}',
    'style={{flex:1, height: \'40px\'}}'
)
gc_tsx = gc_tsx.replace(
    'className="btn" style={{padding: \'8px 16px\'}}',
    'className="btn" style={{padding: \'0 20px\', whiteSpace: \'nowrap\', height: \'40px\'}}'
)

# Fix gaps for checkboxes
gc_tsx = gc_tsx.replace(
    'gap: \'12px\'',
    'gap: \'16px\', rowGap: \'12px\''
)

with open("frontend/src/components/GroupCard.tsx", "w", encoding="utf-8") as f:
    f.write(gc_tsx)


# 3. index.css
with open("frontend/src/index.css", "r", encoding="utf-8") as f:
    css = f.read()

css = css.replace(
    '.group-card {\n    background-color: var(--md-sys-color-surface-container-highest);\n    border-radius: 16px;\n    padding: 16px;',
    '.group-card {\n    background-color: var(--md-sys-color-surface-container-highest);\n    border-radius: 16px;\n    padding: 24px;'
)

with open("frontend/src/index.css", "w", encoding="utf-8") as f:
    f.write(css)

print("Patch applied.")
