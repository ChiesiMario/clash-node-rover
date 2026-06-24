from PIL import Image

def main():
    img_path = r'C:\Users\Noah\.gemini\antigravity-ide\brain\8c6cd870-2cdb-4a91-9b6f-b04f91788f83\media__1782236004036.png'
    img = Image.open(img_path).convert('RGBA')
    bg_color = img.getpixel((0,0))
    
    # We want to find the bounding box of the leftmost object.
    # Scan columns from left to right.
    in_logo = False
    logo_left = -1
    logo_right = -1
    for x in range(img.width):
        col_has_pixels = False
        for y in range(img.height):
            p = img.getpixel((x,y))
            if sum(abs(p[i] - bg_color[i]) for i in range(3)) > 30:
                col_has_pixels = True
                break
        if col_has_pixels:
            if not in_logo:
                in_logo = True
                logo_left = x
        else:
            if in_logo:
                # Reached the gap between logo and text
                logo_right = x
                break
                
    # Now find top and bottom
    logo_top = -1
    logo_bottom = -1
    for y in range(img.height):
        row_has_pixels = False
        for x in range(logo_left, logo_right):
            p = img.getpixel((x,y))
            if sum(abs(p[i] - bg_color[i]) for i in range(3)) > 30:
                row_has_pixels = True
                break
        if row_has_pixels:
            if logo_top == -1:
                logo_top = y
            logo_bottom = y
            
    print(f"Logo bounds: {logo_left}, {logo_top}, {logo_right}, {logo_bottom}")
    
    # Crop the logo
    # Make it square
    w = logo_right - logo_left
    h = logo_bottom - logo_top
    size = max(w, h)
    
    # create a square image with transparent background
    square = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    
    # Copy pixels and make background transparent
    for x in range(w):
        for y in range(h):
            p = img.getpixel((logo_left + x, logo_top + y))
            # simple transparency: if it's close to bg, make it transparent
            dist = sum(abs(p[i] - bg_color[i]) for i in range(3))
            if dist <= 30:
                # Anti-aliasing edges might have some bg color mixed in. 
                # For simplicity, we just set true bg to alpha 0
                if dist < 10:
                    square.putpixel((x + (size-w)//2, y + (size-h)//2), (p[0], p[1], p[2], 0))
                else:
                    # Partial transparency
                    alpha = int(min(255, (dist - 10) * 10))
                    square.putpixel((x + (size-w)//2, y + (size-h)//2), (p[0], p[1], p[2], alpha))
            else:
                square.putpixel((x + (size-w)//2, y + (size-h)//2), (p[0], p[1], p[2], 255))
                
    # Save as PNG
    square.save('templates/logo.png')
    # Save as ICO (requires sizes)
    square.save('icon.ico', format='ICO', sizes=[(size, size)])
    print("Saved templates/logo.png and icon.ico")

if __name__ == '__main__':
    main()
