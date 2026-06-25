from PIL import Image
import os
import shutil

source_image = r"C:\Users\Noah\.gemini\antigravity-ide\brain\5c682920-2755-40ef-a795-f50e8cacbd6d\hig_rover_logo_1782366723034.png"
favicon_path = r"frontend\public\favicon.ico"
icon_path = r"icon.ico"
assets_logo = r"frontend\src\assets\logo.png"

print("Opening image...")
img = Image.open(source_image)

# Ensure the image is RGBA
img = img.convert("RGBA")

# Save as ico
print("Saving icons...")
img.save(favicon_path, format="ICO", sizes=[(256, 256), (128, 128), (64, 64), (32, 32), (16, 16)])
img.save(icon_path, format="ICO", sizes=[(256, 256), (128, 128), (64, 64), (32, 32), (16, 16)])

# Copy the original image to frontend assets
print("Copying assets...")
os.makedirs(r"frontend\src\assets", exist_ok=True)
shutil.copy(source_image, assets_logo)

# Update index.html
print("Updating index.html...")
with open(r"frontend\index.html", "r", encoding="utf-8") as f:
    html = f.read()
html = html.replace('<link rel="icon" type="image/svg+xml" href="/favicon.svg" />', '<link rel="icon" type="image/x-icon" href="/favicon.ico" />')
with open(r"frontend\index.html", "w", encoding="utf-8") as f:
    f.write(html)

print("Done!")
