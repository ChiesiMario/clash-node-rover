import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import pngToIco from 'png-to-ico';

const iconsDir = "C:\\Users\\Noah\\Documents\\GitHub\\clash-node-rover\\src-tauri\\icons";
// We don't even need the source image because we are generating a perfect mathematical 16-bit style SVG 
// based on the new all-black logo design you provided!

// 64x64 grid (16-bit style: finer than 32x32, but still visibly pixelated)
const svg64 = `
<svg width="64" height="64" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg" shape-rendering="crispEdges">
  <!-- Cross (Vertical and Horizontal) -->
  <rect x="28" y="4" width="8" height="56" fill="black" />
  <rect x="4" y="28" width="56" height="8" fill="black" />
  
  <!-- Diagonals (Stepped for 16-bit pixel art feel) -->
  <!-- Top Left to Bottom Right -->
  <rect x="12" y="12" width="8" height="8" fill="black" />
  <rect x="20" y="20" width="8" height="8" fill="black" />
  <rect x="36" y="36" width="8" height="8" fill="black" />
  <rect x="44" y="44" width="8" height="8" fill="black" />
  
  <!-- Top Right to Bottom Left -->
  <rect x="44" y="12" width="8" height="8" fill="black" />
  <rect x="36" y="20" width="8" height="8" fill="black" />
  <rect x="20" y="36" width="8" height="8" fill="black" />
  <rect x="12" y="44" width="8" height="8" fill="black" />

  <!-- Outer Circles (Approximated as 8x8 pixel blocks) -->
  <rect x="28" y="0" width="8" height="8" fill="black" />
  <rect x="28" y="56" width="8" height="8" fill="black" />
  <rect x="0" y="28" width="8" height="8" fill="black" />
  <rect x="56" y="28" width="8" height="8" fill="black" />
  
  <rect x="8" y="8" width="8" height="8" fill="black" />
  <rect x="48" y="8" width="8" height="8" fill="black" />
  <rect x="8" y="48" width="8" height="8" fill="black" />
  <rect x="48" y="48" width="8" height="8" fill="black" />

  <!-- Center Dot (Larger 16x16 block) -->
  <rect x="24" y="24" width="16" height="16" fill="black" />
</svg>
`;

async function build16Bit() {
  console.log("Generating 16-bit Retro Pixel Icons (64x64 base)...");
  
  const path64 = path.join(iconsDir, 'temp_64.png');
  
  // Render the 64x64 SVG to a PNG
  await sharp(Buffer.from(svg64)).png().toFile(path64);

  // Now scale this 64x64 pixel art up and down using NEAREST NEIGHBOR to preserve the hard pixel edges
  const sizes = [16, 32, 48, 64, 128, 256];
  const pngPaths = [];

  for (const size of sizes) {
    const destPath = path.join(iconsDir, `temp_16bit_${size}.png`);
    
    await sharp(path64)
      .resize(size, size, {
        kernel: sharp.kernel.nearest, // <--- Magic sauce for retro pixel art scaling
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .toFile(destPath);
      
    pngPaths.push(destPath);
    
    if (size === 32) fs.copyFileSync(destPath, path.join(iconsDir, '32x32.png'));
    if (size === 128) fs.copyFileSync(destPath, path.join(iconsDir, '128x128.png'));
    if (size === 256) fs.copyFileSync(destPath, path.join(iconsDir, 'icon.png'));
  }

  console.log("Bundling into icon.ico...");
  const buf = await pngToIco(pngPaths);
  fs.writeFileSync(path.join(iconsDir, 'icon.ico'), buf);

  // Clean up temp files
  for (const p of pngPaths) {
    fs.unlinkSync(p);
  }
  fs.unlinkSync(path64);

  console.log("Done! 16-bit SNES style icons applied.");
}

build16Bit().catch(console.error);
