import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import pngToIco from 'png-to-ico';

const iconsDir = "C:\\Users\\Noah\\Documents\\GitHub\\clash-node-rover\\src-tauri\\icons";
const srcLogo = "C:\\Users\\Noah\\Desktop\\Gemini_Generated_Image_ltudpoltudpoltud.png";

// Generate a perfectly crisp 16x16 and 32x32 SVG
const svg16 = `
<svg width="16" height="16" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg" shape-rendering="crispEdges">
  <!-- Cross -->
  <rect x="7" y="1" width="2" height="14" fill="black" />
  <rect x="1" y="7" width="14" height="2" fill="black" />
  
  <!-- Diagonals -->
  <rect x="3" y="3" width="2" height="2" fill="black" />
  <rect x="5" y="5" width="2" height="2" fill="black" />
  
  <rect x="11" y="3" width="2" height="2" fill="black" />
  <rect x="9" y="5" width="2" height="2" fill="black" />
  
  <rect x="3" y="11" width="2" height="2" fill="black" />
  <rect x="5" y="9" width="2" height="2" fill="black" />
  
  <!-- Bottom right (Green Spoke) -->
  <rect x="9" y="9" width="2" height="2" fill="black" />
  <rect x="11" y="11" width="2" height="2" fill="black" />

  <!-- Center Dot -->
  <rect x="6" y="6" width="4" height="4" fill="black" />
</svg>
`;

const svg32 = `
<svg width="32" height="32" viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg" shape-rendering="crispEdges">
  <!-- Cross -->
  <rect x="14" y="2" width="4" height="28" fill="black" />
  <rect x="2" y="14" width="28" height="4" fill="black" />
  
  <!-- Diagonals -->
  <rect x="6" y="6" width="4" height="4" fill="black" />
  <rect x="10" y="10" width="4" height="4" fill="black" />
  
  <rect x="22" y="6" width="4" height="4" fill="black" />
  <rect x="18" y="10" width="4" height="4" fill="black" />
  
  <rect x="6" y="22" width="4" height="4" fill="black" />
  <rect x="10" y="18" width="4" height="4" fill="black" />
  
  <!-- Bottom right (Green Spoke) -->
  <rect x="18" y="18" width="4" height="4" fill="black" />
  <rect x="22" y="22" width="4" height="4" fill="black" />

  <!-- Center Dot -->
  <rect x="12" y="12" width="8" height="8" fill="black" />
</svg>
`;

async function build() {
  console.log("Generating Pixel-Perfect Icons...");
  
  const path16 = path.join(iconsDir, 'temp_16.png');
  const path32 = path.join(iconsDir, '32x32.png'); // Replace 32x32
  
  // Render SVGs
  await sharp(Buffer.from(svg16)).png().toFile(path16);
  await sharp(Buffer.from(svg32)).png().toFile(path32);

  // Generate larger sizes using standard scaling from original for crispness at high res
  const path64 = path.join(iconsDir, '64x64.png');
  const path128 = path.join(iconsDir, '128x128.png');
  const path256 = path.join(iconsDir, 'icon.png');
  
  await sharp(srcLogo).resize(64).toFile(path64);
  await sharp(srcLogo).resize(128).toFile(path128);
  await sharp(srcLogo).resize(256).toFile(path256);

  console.log("Packing into icon.ico...");
  const buf = await pngToIco([path16, path32, path64, path128, path256]);
  fs.writeFileSync(path.join(iconsDir, 'icon.ico'), buf);
  
  fs.unlinkSync(path16); // Cleanup
  
  console.log("Done! Pixel perfect magic applied.");
}

build().catch(console.error);
