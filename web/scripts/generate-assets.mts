// Rasterize the SVG sources in public/ into the PNG assets that social
// crawlers and the web manifest require (SVG is not valid for OpenGraph or
// manifest icons). Re-run after changing og.svg or favicon.svg:
//   bun scripts/generate-assets.mts
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import sharp from 'sharp';

const pub = resolve(dirname(fileURLToPath(import.meta.url)), '../public');

// OpenGraph card — 1200×630 (1.91:1). Source SVG is already at this size.
await sharp(resolve(pub, 'og.svg'))
  .resize(1200, 630)
  .png()
  .toFile(resolve(pub, 'og.png'));

// PWA / manifest icons — rasterize the mark at high density for crisp edges.
for (const size of [192, 512]) {
  await sharp(resolve(pub, 'favicon.svg'), { density: (size * 72) / 32 })
    .resize(size, size)
    .png()
    .toFile(resolve(pub, `icon-${size}.png`));
}

// Maskable icon — the mark centered inside the safe zone on a solid brand
// background so Android adaptive-icon masks never clip it (manifest `purpose:
// maskable`). The mark fills the central ~59%, well within the 80% safe area.
const maskMark = await sharp(resolve(pub, 'favicon.svg'), { density: (512 * 72) / 32 })
  .resize(300, 300)
  .png()
  .toBuffer();
await sharp({ create: { width: 512, height: 512, channels: 4, background: '#0D1117' } })
  .composite([{ input: maskMark, gravity: 'center' }])
  .png()
  .toFile(resolve(pub, 'icon-maskable-512.png'));

console.log('Generated og.png, icon-192.png, icon-512.png, icon-maskable-512.png');
