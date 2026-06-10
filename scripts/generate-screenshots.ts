/**
 * Charter documentation screenshot generator — 2026 design
 *
 * Features:
 * - ray.so-style gradient background with Charter blue color orbs
 * - JetBrains Mono via Google Fonts
 * - Layered box-shadows with color-matched ambient glow
 * - ANSI → HTML color conversion (preserves Charter's terminal colors)
 * - macOS Sonoma traffic light chrome with micro-glows
 * - Charter design tokens throughout
 *
 * Usage: bun scripts/generate-screenshots.ts
 */

import { spawnSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { chromium } from "playwright";

// biome-ignore lint/complexity/useLiteralKeys: TypeScript noPropertyAccessFromIndexSignature requires bracket notation
const HOME = process.env["HOME"] ?? "/tmp";
// biome-ignore lint/complexity/useLiteralKeys: TypeScript noPropertyAccessFromIndexSignature requires bracket notation
const PATH_ENV = `/opt/homebrew/bin:/usr/bin:/bin:${process.env["PATH"] ?? ""}`;

const repoRoot = resolve(import.meta.dirname, "..");
const screenshotsDir = join(repoRoot, "docs", "product", "images", "screenshots");
const charter = join(repoRoot, "dist", "charter");
const tmpDir = "/tmp/charter-ss";

mkdirSync(screenshotsDir, { recursive: true });
mkdirSync(tmpDir, { recursive: true });

const env: NodeJS.ProcessEnv = {
  ...process.env,
  PATH: PATH_ENV,
  TERM: "xterm-256color",
  COLORTERM: "truecolor",
  FORCE_COLOR: "1",
};

// ─── ANSI → HTML converter ───────────────────────────────────────────────────
// Maps ANSI SGR codes to CSS styles using Charter's design tokens
const ANSI_COLORS: Record<number, string> = {
  // Standard (dark mode optimised hex — not raw ANSI)
  30: "#6b7280", // black → dim gray
  31: "#ff6b6b", // red
  32: "#57d964", // green
  33: "#f7c948", // yellow
  34: "#58a6ff", // blue (Charter info-dark)
  35: "#d45fff", // magenta
  36: "#39d5d5", // cyan
  37: "#e6edf3", // white
  // Bright variants
  90: "#8b949e",
  91: "#ff5f57",
  92: "#3fb950",
  93: "#fbbf24",
  94: "#60a5fa",
  95: "#e879f9",
  96: "#22d3ee",
  97: "#f0f6fc",
};

function ansiToHtml(raw: string): string {
  let result = "";
  let openSpan = false;

  // Escape HTML entities first (before processing ANSI)
  const safe = (s: string) =>
    s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

  const ESC = "\x1b";
  const parts = raw.split(ESC);

  for (let i = 0; i < parts.length; i++) {
    const part: string = parts[i] ?? "";

    if (i === 0) {
      result += safe(part);
      continue;
    }

    // Must start with CSI introducer `[`
    if (!part.startsWith("[")) {
      result += safe(part);
      continue;
    }

    const semiEnd = part.search(/[A-Za-z]/);
    if (semiEnd < 0) {
      result += safe(part);
      continue;
    }

    const code: string = part[semiEnd] ?? "";
    const params = part.slice(1, semiEnd);
    const rest = part.slice(semiEnd + 1);

    if (code === "m") {
      const nums = params.split(";").map(Number);
      const first = nums[0] ?? 0;
      if (first === 0 || params === "") {
        if (openSpan) {
          result += "</span>";
          openSpan = false;
        }
      } else {
        let style = "";
        for (let n = 0; n < nums.length; n++) {
          const v: number = nums[n] ?? 0;
          if (v === 1) style += "font-weight:600;color:#f0f6fc;";
          else if (v === 2) style += "opacity:0.55;";
          else if (v === 3) style += "font-style:italic;";
          else if (ANSI_COLORS[v]) style += `color:${ANSI_COLORS[v]};`;
          else if (v === 38 && (nums[n + 1] ?? 0) === 2) {
            // Truecolor: 38;2;R;G;B
            const r = nums[n + 2] ?? 0;
            const g = nums[n + 3] ?? 0;
            const b = nums[n + 4] ?? 0;
            style += `color:rgb(${r},${g},${b});`;
            n += 4;
          }
        }
        if (style) {
          if (openSpan) result += "</span>";
          result += `<span style="${style}">`;
          openSpan = true;
        }
      }
    }
    result += safe(rest);
  }

  if (openSpan) result += "</span>";
  return result;
}

// ─── Terminal HTML template ───────────────────────────────────────────────────
function terminalHtml(cwd: string, command: string, rawOutput: string, displayOverride?: string): string {
  const display = displayOverride ?? cwd.replace(HOME, "~");
  const body = ansiToHtml(rawOutput);

  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:ital,wght@0,400;0,500;0,600;1,400&display=swap" rel="stylesheet">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  /* Charter brand: two radial orbs — blue at top-left, indigo at bottom-right */
  background:
    radial-gradient(ellipse 75% 65% at 12% 18%, rgba(37, 99, 235, 0.28) 0%, transparent 58%),
    radial-gradient(ellipse 55% 55% at 90% 85%, rgba(99, 60, 255, 0.22) 0%, transparent 52%),
    radial-gradient(ellipse 40% 40% at 55% 50%, rgba(37, 99, 235, 0.06) 0%, transparent 60%),
    #08090f;
  padding: 56px 64px;
  display: inline-block;
  min-width: 860px;
}

/* ── Ambient outer glow wrapper ── */
.glow {
  border-radius: 14px;
  /* Color-matched ambient halo — Charter blue at ~25% */
  filter: drop-shadow(0 40px 80px rgba(37, 99, 235, 0.28));
}

/* ── Terminal window ── */
.win {
  width: 760px;
  border-radius: 12px;
  overflow: hidden;
  background: #0d1117;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04),     /* inner rim */
    0 2px 4px  rgba(0, 0, 0, 0.6),
    0 8px 16px rgba(0, 0, 0, 0.55),
    0 20px 40px rgba(0, 0, 0, 0.45),
    0 40px 80px rgba(0, 0, 0, 0.35);
}

/* ── Titlebar ── */
.bar {
  height: 40px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  gap: 0;
  background: #161b22;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.lights { display: flex; gap: 8px; align-items: center; }

.d { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }
.dr { background: #ff5f57; box-shadow: 0 0 5px rgba(255,95,87,0.55); }
.dy { background: #febc2e; box-shadow: 0 0 5px rgba(254,188,46,0.45); }
.dg { background: #28c840; box-shadow: 0 0 5px rgba(40,200,64,0.45); }

.title {
  flex: 1;
  text-align: center;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  font-size: 11.5px;
  color: #8b949e;
  letter-spacing: 0.05em;
  /* shift left to visually center between traffic lights and right edge */
  margin-right: 52px;
}

/* ── Terminal body ── */
.body {
  padding: 18px 22px 22px;
  font-family: 'JetBrains Mono', 'Cascadia Code', 'Geist Mono', ui-monospace, monospace;
  font-size: 13px;
  line-height: 1.65;
  font-weight: 400;
  font-feature-settings: 'liga' 1, 'calt' 1;
  letter-spacing: 0.01em;
  color: #e6edf3;
  background: #0d1117;
}

/* ── Prompt line ── */
.prompt {
  margin-bottom: 8px;
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.cwd   { color: #58a6ff; font-size: 12px; }
.arrow { color: #4d9de0; margin: 0 2px; }
.cmd   { color: #f0f6fc; font-weight: 500; }

/* ── Output ── */
.out {
  white-space: pre-wrap;
  word-break: break-word;
  tab-size: 2;
}

/* ── Cursor ── */
.cursor {
  display: inline-block;
  width: 7px;
  height: 14px;
  background: #58a6ff;
  opacity: 0.8;
  vertical-align: text-bottom;
  margin-left: 1px;
}
</style>
</head>
<body>
<div class="glow">
  <div class="win">
    <div class="bar">
      <div class="lights">
        <div class="d dr"></div>
        <div class="d dy"></div>
        <div class="d dg"></div>
      </div>
      <div class="title">${display} — zsh</div>
    </div>
    <div class="body">
      <div class="prompt">
        <span class="cwd">${display.replace(/</g, "&lt;").replace(/>/g, "&gt;")}</span>
        <span class="arrow">&#10095;</span>
        <span class="cmd">${command.replace(/</g, "&lt;").replace(/>/g, "&gt;")}</span>
      </div>
      <div class="out">${body}</div>
    </div>
  </div>
</div>
</body>
</html>`;
}

// ─── Run charter ──────────────────────────────────────────────────────────────
function runCharter(...args: string[]): string {
  const r = spawnSync(charter, args, { encoding: "utf8", env, timeout: 30000 });
  return ((r.stdout ?? "") + (r.stderr ?? "")).replace(/\r/g, "").trimEnd();
}

// ─── Playwright helpers ───────────────────────────────────────────────────────
let browser: Awaited<ReturnType<typeof chromium.launch>> | null = null;

async function getBrowser() {
  if (!browser) browser = await chromium.launch();
  return browser;
}

async function shotWindow(htmlFile: string, outFile: string) {
  const b = await getBrowser();
  const p = await b.newPage();
  await p.setViewportSize({ width: 1000, height: 1400 });
  // High DPI for crispy screenshots
  await p.emulateMedia({ colorScheme: "dark" });
  await p.goto(`file://${htmlFile}`);
  // Wait for Google Fonts to load
  await p.waitForTimeout(1500);
  const el = await p.$(".glow");
  if (el) {
    await el.screenshot({ path: outFile, type: "png", scale: "device" });
  } else {
    await p.screenshot({ path: outFile, type: "png", fullPage: false });
  }
  await p.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

async function shotUrl(url: string, outFile: string, w = 1440, h = 900) {
  const b = await getBrowser();
  const p = await b.newPage();
  await p.setViewportSize({ width: w, height: h });
  await p.emulateMedia({ colorScheme: "dark" });
  await p.goto(url);
  await p.waitForTimeout(2000);
  await p.screenshot({ path: outFile, type: "png", fullPage: false });
  await p.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

// ─── Screenshot jobs ──────────────────────────────────────────────────────────
const fnApiPath = `${HOME}/FN-Projects/fn-api-v3`;

// 1. doctor PASS — charter itself (100/100), shown as generic "my-platform"
console.log("\n📸 1/5  charter doctor — passing (100/100)");
const passOut = runCharter("doctor", "--path", repoRoot);
const passHtml = join(tmpDir, "doctor-pass.html");
writeFileSync(passHtml, terminalHtml(repoRoot, "charter doctor", passOut, "~/projects/my-platform"));
await shotWindow(passHtml, join(screenshotsDir, "doctor-overview.png"));
await shotWindow(passHtml, join(screenshotsDir, "doctor-tty.png"));
await shotWindow(passHtml, join(screenshotsDir, "quickstart-scan.png"));

// 2. doctor FAIL — real findings, displayed as generic "~/work/backend-api"
if (existsSync(fnApiPath)) {
  console.log("\n📸 2/5  charter doctor — failing (59/100 with real findings)");
  const failOut = runCharter("doctor", "--path", fnApiPath);
  const failHtml = join(tmpDir, "doctor-fail.html");
  writeFileSync(failHtml, terminalHtml(fnApiPath, "charter doctor", failOut, "~/work/backend-api"));
  await shotWindow(failHtml, join(screenshotsDir, "adopt-first-scan.png"));
}

// 3. charter fix --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 3/5  charter fix --dry-run");
  const fixOut = runCharter("fix", "--path", fnApiPath, "--dry-run");
  const fixHtml = join(tmpDir, "fix-dry-run.html");
  writeFileSync(fixHtml, terminalHtml(fnApiPath, "charter fix --dry-run", fixOut, "~/work/backend-api"));
  await shotWindow(fixHtml, join(screenshotsDir, "fix-dry-run.png"));
}

// 4. charter init --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 4/5  charter init --dry-run");
  const initOut = runCharter("init", "--path", fnApiPath, "--dry-run");
  const initHtml = join(tmpDir, "init-output.html");
  writeFileSync(initHtml, terminalHtml(fnApiPath, "charter init --dry-run", initOut, "~/work/backend-api"));
  await shotWindow(initHtml, join(screenshotsDir, "init-output.png"));
}

// 5. HTML report — charter self (100/100)
console.log("\n📸 5/5  charter report (charter self)");
const reportPath = join(tmpDir, "charter-report.html");
runCharter("report", "--path", repoRoot, "--out", reportPath);
if (existsSync(reportPath)) {
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-html.png"), 1440, 900);
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-overview.png"), 1440, 900);
}

if (browser != null) await (browser as Awaited<ReturnType<typeof chromium.launch>>).close();

// Summary
console.log("\n✅ Done:");
for (const f of readdirSync(screenshotsDir)) {
  if (f.endsWith(".png")) {
    const kb = Math.round(statSync(join(screenshotsDir, f)).size / 1024);
    console.log(`  ${f.padEnd(40)} ${kb}kb`);
  }
}
