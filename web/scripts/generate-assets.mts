// Rasterize the SVG sources in public/ into the PNG assets that social
// crawlers and the web manifest require (SVG is not valid for OpenGraph or
// manifest icons). Re-run after changing og.svg or favicon.svg:
//   bun scripts/generate-assets.mts
import { writeFile } from 'node:fs/promises';
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

// favicon.ico — browsers and bots auto-request /favicon.ico regardless of the
// <link rel=icon> SVG, so ship a real one to avoid a guaranteed 404. ICO can
// embed a PNG (supported by every modern browser), so wrap a 48×48 render in a
// minimal single-image ICO container rather than pulling in an encoder dependency.
const icoPng = await sharp(resolve(pub, 'favicon.svg'), { density: (48 * 72) / 32 })
  .resize(48, 48)
  .png()
  .toBuffer();
const icoHeader = Buffer.alloc(6);
icoHeader.writeUInt16LE(1, 2); // image type: icon
icoHeader.writeUInt16LE(1, 4); // image count
const icoEntry = Buffer.alloc(16);
icoEntry.writeUInt8(48, 0); // width
icoEntry.writeUInt8(48, 1); // height
icoEntry.writeUInt16LE(1, 4); // color planes
icoEntry.writeUInt16LE(32, 6); // bits per pixel
icoEntry.writeUInt32LE(icoPng.length, 8); // image byte size
icoEntry.writeUInt32LE(icoHeader.length + icoEntry.length, 12); // offset to PNG data
await writeFile(resolve(pub, 'favicon.ico'), Buffer.concat([icoHeader, icoEntry, icoPng]));

console.log('Generated og.png, icon-192.png, icon-512.png, icon-maskable-512.png, favicon.ico');
