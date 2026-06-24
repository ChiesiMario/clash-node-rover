import sys
import base64
from PIL import Image
import io

img_path = r"C:\Users\Noah\.gemini\antigravity-ide\brain\8c6cd870-2cdb-4a91-9b6f-b04f91788f83\md3_rover_logo_1782239473052.png"
html_path = "templates/index.html"
ico_path = "icon.ico"

# Read image
img = Image.open(img_path)

# 1. Save as icon.ico (resized if needed)
icon_sizes = [(16,16), (32, 32), (48, 48), (64,64)]
img.save(ico_path, format='ICO', sizes=icon_sizes)
print("Saved icon.ico")

# 2. Resize for web favicon (e.g. 64x64) and base64
img_small = img.resize((64, 64), Image.Resampling.LANCZOS)
buffer = io.BytesIO()
img_small.save(buffer, format="PNG")
img_b64 = base64.b64encode(buffer.getvalue()).decode("utf-8")

# 3. Replace in templates/index.html
with open(html_path, "r", encoding="utf-8") as f:
    html = f.read()

import re
# Find <link rel="icon" type="image/png" href="data:image/png;base64,...">
new_html = re.sub(
    r'<link rel="icon" type="image/png" href="data:image/png;base64,[A-Za-z0-9+/=]+">',
    f'<link rel="icon" type="image/png" href="data:image/png;base64,{img_b64}">',
    html
)

with open(html_path, "w", encoding="utf-8") as f:
    f.write(new_html)

print("Updated index.html")
