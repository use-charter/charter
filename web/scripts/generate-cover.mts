// Renders the use-charter org cover banner — a minimal, centered brand hero
// (wordmark, tagline, mission, capability footer) using the brand faces: Ruda
// for display and Atkinson Hyperlegible Mono for the mono lines. Outputs both a
// source copy under docs/internal/designs/brands and the served /charter-cover
// asset. Re-run after editing:  bun scripts/generate-cover.mts
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { Resvg } from '@resvg/resvg-js';
import sharp from 'sharp';

const web = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const fonts = resolve(web, 'src/assets/fonts');
const ruda = readFileSync(resolve(fonts, 'Ruda-ExtraBold.ttf'));
const rudaR = readFileSync(resolve(fonts, 'Ruda-Regular.ttf'));
const mono = readFileSync(resolve(fonts, 'AtkinsonHyperlegibleMono.ttf'));
const MONO = 'Atkinson Hyperlegible Mono';

const W = 1280;
const H = 480;
const cx = W / 2;
const C = {
  bg: '#0b0e14', ink: '#e9edf3', faint: '#9aa3b2', dim: '#5c6470',
  blue: '#5aa2f7', green: '#3fb950', line: '#1c2330',
};

const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${W * 2}" height="${H * 2}" viewBox="0 0 ${W} ${H}">
  <defs>
    <radialGradient id="glow" cx="50%" cy="34%" r="55%">
      <stop offset="0%" stop-color="#5aa2f7" stop-opacity="0.20"/>
      <stop offset="60%" stop-color="#5aa2f7" stop-opacity="0.04"/>
      <stop offset="100%" stop-color="#5aa2f7" stop-opacity="0"/>
    </radialGradient>
  </defs>

  <rect width="${W}" height="${H}" fill="${C.bg}"/>
  <rect width="${W}" height="${H}" fill="url(#glow)"/>

  <!-- wordmark — the C carries a small baseline nudge to sit centered in the brackets -->
  <text x="${cx}" y="224" text-anchor="middle" font-family="Ruda" font-weight="800" font-size="104" letter-spacing="-3">
    <tspan fill="${C.blue}">[</tspan><tspan fill="${C.blue}" dy="7">C</tspan><tspan fill="${C.blue}" dy="-7">]</tspan><tspan fill="${C.ink}"> charter</tspan>
  </text>

  <text x="${cx}" y="288" text-anchor="middle" font-family="${MONO}" font-size="25" letter-spacing="5" fill="${C.faint}">AI-AGENT READINESS, SCORED</text>

  <text x="${cx}" y="356" text-anchor="middle" font-family="Ruda" font-weight="400" font-size="26" fill="${C.faint}">Open-source tools that make any repository safe for AI coding agents.</text>

  <line x1="${cx - 240}" y1="406" x2="${cx + 240}" y2="406" stroke="${C.line}" stroke-width="1.5"/>

  <text x="${cx}" y="444" text-anchor="middle" font-family="${MONO}" font-size="19" letter-spacing="1.5" fill="${C.dim}">
    <tspan fill="${C.green}">●</tspan> <tspan fill="${C.faint}">CLI</tspan>  ·  <tspan fill="${C.faint}">GitHub Action</tspan>  ·  offline  ·  deterministic  ·  Apache-2.0
  </text>
</svg>`;

const png = new Resvg(svg, { font: { fontBuffers: [ruda, rudaR, mono], loadSystemFonts: true } }).render().asPng();
const brands = resolve(web, '../docs/internal/designs/brands');
writeFileSync(resolve(brands, 'charter-cover.png'), png);
await sharp(png).webp({ quality: 92 }).toFile(resolve(brands, 'charter-cover.webp'));
writeFileSync(resolve(web, 'public/charter-cover.png'), png); // served at /charter-cover.png
console.log(`org cover ${W}x${H} @2x → docs/internal/designs/brands + web/public`);
