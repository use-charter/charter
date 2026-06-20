// Renders the use-charter org cover with headless Chromium (Playwright) so the
// brand web fonts load through @font-face exactly as they do on the site: Ruda
// for display (its 800 weight for the wordmark) and Atkinson Hyperlegible Mono
// for the capability footer. Fonts are embedded as base64 so the render is
// self-contained. Outputs the source under docs/internal/designs/brand and the
// served /charter-cover asset. Re-run:  bun scripts/generate-cover.mts
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';
import sharp from 'sharp';

const web = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const b64 = (p: string) => readFileSync(resolve(web, p)).toString('base64');
const ruda = b64('public/fonts/ruda-400-800.woff2');
const mono = b64('public/fonts/atkinson-hyperlegible-mono-400-700.woff2');

const W = 1280;
const H = 480;

const html = `<!doctype html><html><head><meta charset="utf-8"><style>
  @font-face { font-family:'Ruda'; src:url(data:font/woff2;base64,${ruda}) format('woff2'); font-weight:400 800; font-display:block; }
  @font-face { font-family:'Atkinson Mono'; src:url(data:font/woff2;base64,${mono}) format('woff2'); font-display:block; }
  * { margin:0; padding:0; box-sizing:border-box; }
  body {
    width:${W}px; height:${H}px; overflow:hidden;
    background:radial-gradient(60% 55% at 50% 34%, rgba(90,162,247,0.20), rgba(90,162,247,0.04) 60%, transparent 100%), #0b0e14;
    color:#e9edf3; font-family:'Ruda', sans-serif;
    display:flex; flex-direction:column; align-items:center; justify-content:center; gap:26px;
    -webkit-font-smoothing:antialiased;
  }
  .mark { font-weight:800; font-size:104px; letter-spacing:-3px; line-height:1; }
  .mark .b { color:#5aa2f7; }
  .tag { font-weight:400; font-size:24px; letter-spacing:5px; color:#9aa3b2; }
  .mission { font-weight:400; font-size:25px; color:#9aa3b2; }
  .rule { width:480px; height:1px; background:#1c2330; margin:6px 0; }
  .foot { font-family:'Atkinson Mono', monospace; font-size:18px; letter-spacing:1.5px; color:#5c6470; display:flex; gap:14px; align-items:center; }
  .foot .dot { color:#3fb950; }
  .foot b { color:#9aa3b2; font-weight:400; }
</style></head><body>
  <div class="mark"><span class="b">[C]</span> charter</div>
  <div class="tag">AI-AGENT READINESS, SCORED</div>
  <div class="mission">Open-source tools that make any repository safe for AI coding agents.</div>
  <div class="rule"></div>
  <div class="foot"><span class="dot">●</span> <b>CLI</b> · <b>GitHub Action</b> · offline · deterministic · Apache-2.0</div>
</body></html>`;

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: W, height: H }, deviceScaleFactor: 2 });
await page.setContent(html, { waitUntil: 'load' });
await page.evaluate('document.fonts.ready');
const png = await page.screenshot({ type: 'png', clip: { x: 0, y: 0, width: W, height: H } });
await browser.close();

writeFileSync(resolve(web, '../docs/internal/designs/brand/charter-cover.png'), png);
writeFileSync(resolve(web, 'public/charter-cover.png'), png);
await sharp(png).webp({ quality: 92 }).toFile(resolve(web, '../docs/internal/designs/brand/charter-cover.webp'));
console.log(`org cover ${W}x${H} @2x (Chromium) → docs/internal/designs/brand + web/public`);
