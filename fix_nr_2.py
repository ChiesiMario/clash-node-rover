with open('frontend/src/components/NodeRanking.tsx', 'r', encoding='utf-8') as f:
    nr = f.read()

nr = nr.replace("{ nodeName, safeId, avgDelay, jitter, score }", "{ nodeName, avgDelay, jitter, score }")

with open('frontend/src/components/NodeRanking.tsx', 'w', encoding='utf-8') as f:
    f.write(nr)
