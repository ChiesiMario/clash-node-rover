with open('frontend/src/components/NodeRanking.tsx', 'r', encoding='utf-8') as f:
    nr = f.read()

nr = nr.replace("safeId: string, ", "")
nr = nr.replace(" safeId={btoa(encodeURIComponent(node.Name)).replace(/=/g, '')}", "")
nr = nr.replace("<React.Fragment key={node.Name}>", "<Fragment key={node.Name}>")
nr = nr.replace("</React.Fragment>", "</Fragment>")
nr = nr.replace("import { useState", "import { Fragment, useState")

with open('frontend/src/components/NodeRanking.tsx', 'w', encoding='utf-8') as f:
    f.write(nr)
