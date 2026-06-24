import base64

with open('logo_base64.txt', 'r') as f:
    b64 = f.read().strip()

with open('templates/index.html', 'r', encoding='utf-8') as f:
    html = f.read()

# Replace favicon
new_favicon = '<link rel="icon" type="image/png" href="data:image/png;base64,' + b64 + '">'
if '<link rel="icon"' in html:
    pass # Needs manual replacement
else:
    html = html.replace('</title>', '</title>\n    ' + new_favicon)

# Replace logo text with image
logo_img = '<div style="display: flex; align-items: center; gap: 12px;"><img src="data:image/png;base64,' + b64 + '" width="40" height="40" style="border-radius: 50%;"> <div class="logo" style="margin-bottom: 0;">NODE ROVER</div></div>'
html = html.replace('<div class="logo">NODE ROVER</div>', logo_img)

with open('templates/index.html', 'w', encoding='utf-8') as f:
    f.write(html)
print('Updated templates/index.html')
