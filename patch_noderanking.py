import re

with open("frontend/src/components/NodeRanking.tsx", "r", encoding="utf-8") as f:
    code = f.read()

# Typography
code = code.replace('<div className="card-title">', '<div className="card-title md3-title-large">')
code = code.replace('<div style={{fontSize:\'13px\',', '<div className="md3-body-medium" style={{')
code = code.replace('<div style={{fontWeight:\'bold\', marginBottom:\'8px\'}}>', '<div className="md3-title-medium" style={{marginBottom:\'8px\'}}>')

# Small text sizing
code = code.replace('fontSize:\'12px\'', 'fontSize:\'13px\'')

with open("frontend/src/components/NodeRanking.tsx", "w", encoding="utf-8") as f:
    f.write(code)

print("NodeRanking.tsx updated.")
