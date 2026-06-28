import sharp from 'sharp';
import fs from 'fs';
import path from 'path';
import pngToIco from 'png-to-ico';

const srcLogo = "C:\\Users\\Noah\\Desktop\\Logo.png";
const iconsDir = "C:\\Users\\Noah\\Documents\\GitHub\\clash-node-rover\\src-tauri\\icons";

async function generate() {
  console.log("Generating high-quality icons with sharp (Lanczos3 + Sharpen)...");

  const sizes = [16, 32, 48, 64, 128, 256];
  const pngPaths = [];

  for (const size of sizes) {
    const destPath = path.join(iconsDir, `temp_${size}.png`);
    
    // Use sharp with lanczos3 and mild unsharp mask
    await sharp(srcLogo)
      .resize(size, size, {
        kernel: sharp.kernel.lanczos3,
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .sharpen({ sigma: 1, m1: 1, m2: 2 })
      .toFile(destPath);
      
    pngPaths.push(destPath);
    
    // Also overwrite some specific files that Tauri uses
    if (size === 32) {
      await sharp(destPath).toFile(path.join(iconsDir, '32x32.png'));
    }
    if (size === 128) {
      await sharp(destPath).toFile(path.join(iconsDir, '128x128.png'));
    }
    if (size === 256) {
      await sharp(destPath).toFile(path.join(iconsDir, 'icon.png'));
    }
  }

  console.log("Bundling into icon.ico...");
  // Create ico file
  const buf = await pngToIco(pngPaths);
  fs.writeFileSync(path.join(iconsDir, 'icon.ico'), buf);

  // Clean up temp files
  for (const p of pngPaths) {
    fs.unlinkSync(p);
  }

  console.log("Done! High-quality icons created.");
}

generate().catch(console.error);
