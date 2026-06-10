/**
 * Charter documentation screenshot generator
 * Usage: bun scripts/generate-screenshots.ts
 * Prereq: bun add -d playwright && npx playwright install chromium
 */

import { spawnSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { chromium } from "playwright";

const repoRoot = resolve(import.meta.dirname, "..");
const screenshotsDir = join(repoRoot, "docs", "product", "images", "screenshots");
const charter = join(repoRoot, "dist", "charter");
const tmpDir = "/tmp/charter-ss";

mkdirSync(screenshotsDir, { recursive: true });
mkdirSync(tmpDir, { recursive: true });

// biome-ignore lint/complexity/useLiteralKeys: TypeScript noPropertyAccessFromIndexSignature requires bracket notation
const HOME = process.env["HOME"] ?? "/tmp";
// biome-ignore lint/complexity/useLiteralKeys: TypeScript noPropertyAccessFromIndexSignature requires bracket notation
const PATH_ENV = `/opt/homebrew/bin:/usr/bin:/bin:${process.env["PATH"] ?? ""}`;

const env: NodeJS.ProcessEnv = {
  ...process.env,
  PATH: PATH_ENV,
  NO_COLOR: "1",
  TERM: "dumb",
};

// Strip ANSI escape sequences without using control characters in regex
function stripAnsi(s: string): string {
  // Replace ESC sequences by splitting on the ESC character (char code 27)
  const ESC = String.fromCharCode(27);
  const parts = s.split(ESC);
  return parts
    .map((p, i) => {
      if (i === 0) return p;
      // Drop everything up to and including the final letter of the sequence
      const end = p.search(/[A-Za-z]/);
      return end >= 0 ? p.slice(end + 1) : "";
    })
    .join("");
}

function runCharter(...args: string[]): string {
  const r = spawnSync(charter, args, { encoding: "utf8", env, timeout: 30000 });
  return stripAnsi((r.stdout ?? "") + (r.stderr ?? ""))
    .replace(/\r/g, "")
    .trim();
}

function terminalHtml(cwd: string, command: string, output: string): string {
  const safe = output
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");

  const displayCwd = cwd.replace(HOME, "~");

  return `<!DOCTYPE html><html><head><meta charset="utf-8"><style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0d1117;padding:20px;font-family:"SF Mono","Cascadia Code","Fira Code",monospace;font-size:13px}
.win{background:#161b22;border:1px solid #30363d;border-radius:10px;overflow:hidden;box-shadow:0 24px 80px rgba(0,0,0,.7);width:720px}
.bar{background:#21262d;padding:11px 14px;display:flex;align-items:center;gap:7px;border-bottom:1px solid #30363d}
.d{width:12px;height:12px;border-radius:50%}
.dr{background:#ff5f57}.da{background:#ffbd2e}.dg{background:#28c840}
.t{color:#8b949e;font-size:11.5px;margin-left:8px;font-family:-apple-system,sans-serif}
.body{padding:16px 18px}
.pr{margin-bottom:6px}
.sym{color:#58a6ff}
.cmd{color:#e6edf3;font-weight:500}
.out{color:#c9d1d9;white-space:pre-wrap;word-break:break-word;line-height:1.55}
</style></head><body>
<div class="win">
  <div class="bar">
    <div class="d dr"></div><div class="d da"></div><div class="d dg"></div>
    <span class="t">${displayCwd} — zsh</span>
  </div>
  <div class="body">
    <div class="pr"><span class="sym">&#10095; </span><span class="cmd">${command.replace(/</g, "&lt;").replace(/>/g, "&gt;")}</span></div>
    <div class="out">${safe}</div>
  </div>
</div>
</body></html>`;
}

let browser: Awaited<ReturnType<typeof chromium.launch>> | null = null;

async function getBrowser() {
  if (!browser) browser = await chromium.launch();
  return browser;
}

async function shotWindow(htmlFile: string, outFile: string) {
  const b = await getBrowser();
  const p = await b.newPage();
  await p.setViewportSize({ width: 800, height: 1200 });
  await p.goto(`file://${htmlFile}`);
  await p.waitForTimeout(300);
  const el = await p.$(".win");
  if (el) {
    await el.screenshot({ path: outFile, type: "png" });
  } else {
    await p.screenshot({ path: outFile, type: "png" });
  }
  await p.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

async function shotUrl(url: string, outFile: string, w = 1280, h = 800) {
  const b = await getBrowser();
  const p = await b.newPage();
  await p.setViewportSize({ width: w, height: h });
  await p.goto(url);
  await p.waitForTimeout(1500);
  await p.screenshot({ path: outFile, type: "png", fullPage: false });
  await p.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

const fnApiPath = `${HOME}/FN-Projects/fn-api-v3`;

// 1. doctor PASS — charter scores 100 on itself
console.log("\n📸 1/5  charter doctor — passing (charter self, 100/100)");
const passOut = runCharter("doctor", "--path", repoRoot);
const passHtml = join(tmpDir, "doctor-pass.html");
writeFileSync(passHtml, terminalHtml(repoRoot, "charter doctor", passOut));
await shotWindow(passHtml, join(screenshotsDir, "doctor-overview.png"));
await shotWindow(passHtml, join(screenshotsDir, "doctor-tty.png"));
await shotWindow(passHtml, join(screenshotsDir, "quickstart-scan.png"));

// 2. doctor FAIL — fn-api-v3, 59/100 with real findings
if (existsSync(fnApiPath)) {
  console.log("\n📸 2/5  charter doctor — failing (fn-api-v3, 59/100)");
  const failOut = runCharter("doctor", "--path", fnApiPath);
  const failHtml = join(tmpDir, "doctor-fail.html");
  writeFileSync(failHtml, terminalHtml(fnApiPath, "charter doctor", failOut));
  await shotWindow(failHtml, join(screenshotsDir, "adopt-first-scan.png"));
}

// 3. charter fix --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 3/5  charter fix --dry-run (fn-api-v3)");
  const fixOut = runCharter("fix", "--path", fnApiPath, "--dry-run");
  const fixHtml = join(tmpDir, "fix-dry-run.html");
  writeFileSync(fixHtml, terminalHtml(fnApiPath, "charter fix --dry-run", fixOut));
  await shotWindow(fixHtml, join(screenshotsDir, "fix-dry-run.png"));
}

// 4. charter init --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 4/5  charter init --dry-run (fn-api-v3)");
  const initOut = runCharter("init", "--path", fnApiPath, "--dry-run");
  const initHtml = join(tmpDir, "init-output.html");
  writeFileSync(initHtml, terminalHtml(fnApiPath, "charter init --dry-run", initOut));
  await shotWindow(initHtml, join(screenshotsDir, "init-output.png"));
}

// 5. HTML report — charter self (100/100)
console.log("\n📸 5/5  charter report (charter self)");
const reportPath = join(tmpDir, "charter-report.html");
runCharter("report", "--path", repoRoot, "--out", reportPath);
if (existsSync(reportPath)) {
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-html.png"), 1280, 800);
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-overview.png"), 1280, 800);
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
