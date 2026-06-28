import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import pngToIco from 'png-to-ico';

const srcLogo = "C:\\Users\\Noah\\Desktop\\Logo.png";
const iconsDir = "C:\\Users\\Noah\\Documents\\GitHub\\clash-node-rover\\src-tauri\\icons";

async function generateSmooth() {
  console.log("Generating smooth, anti-aliased icons...");

  // We include a variety of sizes to ensure Windows has exactly what it needs
  // at any DPI scaling (100%, 125%, 150%, 200%, etc.)
  const sizes = [16, 24, 32, 48, 64, 128, 256];
  const pngPaths = [];

  for (const size of sizes) {
    const destPath = path.join(iconsDir, `temp_smooth_${size}.png`);
    
    // High-quality downscaling using Lanczos3 (default in sharp for downsizing)
    // with anti-aliasing enabled. We add a very subtle sharpen to keep edges crisp but smooth.
    await sharp(srcLogo)
      .resize(size, size, {
        kernel: sharp.kernel.lanczos3,
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      // A gentle unsharp mask to prevent it from being *too* blurry while retaining smoothness
      .sharpen({ sigma: 0.5, m1: 0.5, m2: 1.0 })
      .toFile(destPath);
      
    pngPaths.push(destPath);
    
    // Replace Tauri's standard files
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
  // Create ico file containing ALL the sizes so Windows can pick the perfect one
  const buf = await pngToIco(pngPaths);
  fs.writeFileSync(path.join(iconsDir, 'icon.ico'), buf);

  // Clean up temp files
  for (const p of pngPaths) {
    fs.unlinkSync(p);
  }

  console.log("Done! Smooth, modern icons created.");
}

generateSmooth().catch(console.error);
