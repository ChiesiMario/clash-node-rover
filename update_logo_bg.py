from PIL import Image, ImageDraw
import os

source_image = r"C:\Users\Noah\.gemini\antigravity-ide\brain\5c682920-2755-40ef-a795-f50e8cacbd6d\hig_rover_logo_1782366723034.png"
favicon_path = r"frontend\public\favicon.ico"
ico_path = "icon.ico"
assets_logo = r"frontend\src\assets\logo.png"

def process_logo():
    print("Opening image...")
    img = Image.open(source_image).convert("RGBA")
    w, h = img.size
    
    # We want to find the blue box.
    # The blue box has high saturation.
    min_x, min_y, max_x, max_y = w, h, 0, 0
    
    for y in range(h):
        for x in range(w):
            r, g, b, a = img.getpixel((x, y))
            # Check saturation: max(r,g,b) - min(r,g,b)
            # A blue pixel will have a large difference between b and r.
            # Also the radar is white (255,255,255), so we also need to include that if it's inside.
            # But just finding the blue edges is enough to get the bounding box.
            if b - r > 50 or b - g > 50:
                min_x = min(min_x, x)
                min_y = min(min_y, y)
                max_x = max(max_x, x)
                max_y = max(max_y, y)
                
    # Add a little padding to make sure we don't clip the edges of the squircle
    padding = 10
    min_x = max(0, min_x - padding)
    min_y = max(0, min_y - padding)
    max_x = min(w, max_x + padding)
    max_y = min(h, max_y + padding)
    
    print(f"Cropping to {min_x}, {min_y} - {max_x}, {max_y}")
    cropped = img.crop((min_x, min_y, max_x, max_y))
    
    cw, ch = cropped.size
    size = min(cw, ch)
    cropped = cropped.crop(((cw - size) // 2, (ch - size) // 2, (cw + size) // 2, (ch + size) // 2))
    
    print("Applying rounded corners...")
    mask = Image.new("L", (size, size), 0)
    draw = ImageDraw.Draw(mask)
    # The generated squircle already HAS rounded corners, but we want to make the background transparent.
    # We apply a rounded rectangle mask.
    r = int(size * 0.225)
    draw.rounded_rectangle((0, 0, size, size), radius=r, fill=255)
    
    # Now put the alpha mask
    cropped.putalpha(mask)
    
    # Save the processed image to assets
    cropped.save(assets_logo)
    
    # Save as ICOs
    print("Saving ICOs...")
    cropped.save(favicon_path, format="ICO", sizes=[(256, 256), (128, 128), (64, 64), (32, 32), (16, 16)])
    cropped.save(ico_path, format="ICO", sizes=[(256, 256), (128, 128), (64, 64), (32, 32), (16, 16)])
    
    # Cache bust index.html favicon
    with open(r"frontend\index.html", "r", encoding="utf-8") as f:
        html = f.read()
    html = html.replace('/favicon.ico', '/favicon.ico?v=3')
    with open(r"frontend\index.html", "w", encoding="utf-8") as f:
        f.write(html)
        
    print("Done!")

process_logo()
