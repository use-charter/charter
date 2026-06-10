/**
 * Charter documentation screenshot generator — 2026 design
 *
 * Matches Charter's actual styled TTY output format exactly (from internal/render/text/render.go).
 * Uses --format json for data, then renders HTML that mirrors the real Charter output structure:
 *   [C] charter header → divider → findings (│ bar + icons) → scorecard → score hero
 *
 * Usage: bun scripts/generate-screenshots.ts
 */

import { spawnSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { chromium } from "playwright";
import sharp from "sharp";

// Destructuring avoids noPropertyAccessFromIndexSignature without bracket notation
const { HOME: rawHome = "/tmp", PATH: rawPath = "" } = process.env;
const HOME: string = rawHome;
const PATH_ENV = `/opt/homebrew/bin:/usr/bin:/bin:${rawPath}`;

const repoRoot = resolve(import.meta.dirname, "..");
const screenshotsDir = join(repoRoot, "docs", "product", "images", "screenshots");
const charter = join(repoRoot, "dist", "charter");
const tmpDir = "/tmp/charter-ss";

mkdirSync(screenshotsDir, { recursive: true });
mkdirSync(tmpDir, { recursive: true });

const env: NodeJS.ProcessEnv = { ...process.env, PATH: PATH_ENV, NO_COLOR: "1", TERM: "dumb" };

function runCharter(...args: string[]): string {
  const r = spawnSync(charter, args, { encoding: "utf8", env, timeout: 30000 });
  return ((r.stdout ?? "") + (r.stderr ?? "")).trim();
}

function runCharterJson(...args: string[]): unknown {
  try { return JSON.parse(runCharter(...args, "--format", "json")); } catch { return null; }
}

// ─── Design tokens (matching DESIGN-TOKENS.md) ────────────────────────────
const T = {
  textPrimary:   "#f9fafb",
  textSecondary: "#d1d5db",
  textTertiary:  "#9ca3af",
  textInfo:      "#60a5fa",   // Charter blue accent (text-info-dark)
  textSuccess:   "#4ade80",
  textDanger:    "#f87171",
  textWarning:   "#fbbf24",
  borderInfo:    "#1e40af",
  bgPrimary:     "#0d1117",
  bgSecondary:   "#161b22",
};

// satisfies Record<string,string> lets TypeScript infer the specific key union
// so dot notation is valid (no index signature) while still checking values.
const SEV = {
  BLOCKER:       T.textDanger,
  HIGH:          T.textWarning,
  MEDIUM:        "#fcd34d",
  LOW:           T.textInfo,
  INFORMATIONAL: T.textTertiary,
} satisfies Record<string, string>;

const SEV_BORDER = {
  BLOCKER:       "#ef4444",
  HIGH:          "#f59e0b",
  MEDIUM:        "#eab308",
  LOW:           "#3b82f6",
  INFORMATIONAL: "#374151",
} satisfies Record<string, string>;

// Dynamic runtime lookups (severity value comes from JSON at runtime)
function sevFg(sev: string): string {
  return (SEV as Record<string, string>)[sev] ?? T.textPrimary;
}
function sevBar(sev: string): string {
  return (SEV_BORDER as Record<string, string>)[sev] ?? T.textTertiary;
}

// ─── Severity icon (matches render.go unicodeGlyphs) ──────────────────────
function sevIcon(sev: string): string {
  if (sev === "BLOCKER") return "✗";
  if (sev === "HIGH" || sev === "MEDIUM") return "⚠";
  return "•";
}

// ─── Score bar (matches render.go scoreBar, width=24) ─────────────────────
function scoreBar(score: number, color: string): string {
  const filled = Math.max(0, Math.min(24, Math.round(score * 24 / 100)));
  const empty = 24 - filled;
  return `<span style="color:${color}">${"█".repeat(filled)}</span><span style="color:#21262d">${"░".repeat(empty)}</span>`;
}

function c(color: string, text: string, bold = false): string {
  const fw = bold ? "font-weight:600;" : "";
  return `<span style="${fw}color:${color}">${text}</span>`;
}

// ─── Doctor result renderer (mirrors renderStyled exactly) ────────────────
interface Finding {
  rule_id: string; severity: string; category: string; summary: string;
  locations?: Array<{ path: string; line: number }>;
  evidence?: string[]; remediation?: string;
}
interface CategorySummary { category: string; findings: number; deduction: number; worst_severity: string; }
interface DoctorResult {
  threshold: number; passed: boolean;
  findings: Finding[]; suppressed: unknown[];
  summary: { blocker: number; high: number; medium: number; low: number };
  score: { base: number; final: number };
  categories: CategorySummary[];
}

const DIVIDER = c(T.textTertiary, "─".repeat(48));

function renderDoctor(d: DoctorResult, displayName: string): string {
  const lines: string[] = [];

  // Header: [C] charter  ·  v1.0.0  ·  ~/path
  lines.push(
    `${c(T.textInfo, "[C]", true)} ${c(T.textInfo, "charter", true)}` +
    c(T.textTertiary, `  v1.0.0  ·  ${displayName}`)
  );
  lines.push(DIVIDER);
  lines.push("");

  // Findings
  if (d.findings.length > 0) {
    lines.push(c(T.textSecondary, "Findings"));
    for (const f of d.findings) {
      const fgColor  = sevFg(f.severity);
      const bgColor  = sevBar(f.severity);
      const bar      = `${c(bgColor, "│")} `;
      const icon     = c(fgColor, sevIcon(f.severity));
      const badge    = c(fgColor, f.severity, true);
      const ruleId   = c(T.textInfo, f.rule_id, true);
      const cat      = f.category ? `  ${c(T.textTertiary, f.category)}` : "";
      lines.push(`${bar}${icon} ${badge}  ${ruleId}${cat}`);
      lines.push(`${bar}${c(T.textPrimary, f.summary)}`);
      for (const loc of f.locations ?? []) {
        const locText = loc.line > 0 ? `${loc.path}:${loc.line}` : loc.path;
        lines.push(`${bar}${c(T.textTertiary, "loc ")}${c(T.textTertiary, locText)}`);
      }
      for (const ev of (f.evidence ?? []).slice(0, 3)) {
        lines.push(`${bar}${c(T.textTertiary, "• ")}${c(T.textSecondary, ev)}`);
      }
      if (f.remediation) {
        lines.push(`${bar}${c(T.textTertiary, "fix ")}${c(T.textSecondary, f.remediation)}`);
      }
      lines.push("");
    }
  }

  // Summary: "Checked 18 rules · N findings · 1 BLOCKER"
  const dot = c(T.textTertiary, " · ");
  const total = d.findings.length;
  const worstSev   = d.findings[0]?.severity ?? "INFORMATIONAL";
  const countColor = total > 0 ? sevFg(worstSev) : T.textSuccess;
  const countText  = total > 0 ? `${total} finding${total > 1 ? "s" : ""}` : "0 findings ✓";
  let summary = `${c(T.textSecondary, "Checked 18 rules")}${dot}${c(countColor, countText)}`;
  if (d.summary.blocker > 0) summary += `${dot}${c(SEV.BLOCKER, `${d.summary.blocker} BLOCKER`, true)}`;
  if (d.summary.high > 0)    summary += `${dot}${c(SEV.HIGH, `${d.summary.high} HIGH`)}`;
  if (d.summary.medium > 0)  summary += `${dot}${c(SEV.MEDIUM, `${d.summary.medium} MEDIUM`)}`;
  if (d.summary.low > 0)     summary += `${dot}${c(SEV.LOW, `${d.summary.low} LOW`)}`;
  lines.push(summary);

  // Scorecard: "readiness by category"
  if (d.categories.length > 0) {
    lines.push("");
    lines.push(c(T.textSecondary, "readiness by category"));
    for (const cat of d.categories) {
      const sc   = sevFg(cat.worst_severity);
      const name = c(T.textSecondary, `  ${cat.category.padEnd(12)}`);
      const ded  = c(sc, `−${cat.deduction.toString().padEnd(3)}`);
      const det  = c(T.textTertiary, `${cat.findings} finding(s), worst ${cat.worst_severity}`);
      lines.push(`${name} ${ded} ${det}`);
    }
  }

  // Score hero
  lines.push("");
  lines.push(DIVIDER);
  const scoreTok  = d.passed ? T.textSuccess : T.textDanger;
  const verdict   = d.passed ? "PASS ✓" : "FAIL ✗";
  const scoreNum  = `${c(scoreTok, String(d.score.final), true)}${c(T.textTertiary, "/100")}`;
  const bar24     = scoreBar(d.score.final, scoreTok);
  const badge2    = c(scoreTok, verdict, true);
  lines.push(`${c(T.textSecondary, "Score ")}${scoreNum}  ${bar24}  ${badge2}`);
  if (d.score.final < d.score.base) {
    lines.push(`      ${c(T.textDanger, `cap   score capped at ${d.score.final}`)}`);
  }
  lines.push(`      ${c(T.textTertiary, `threshold ${d.threshold}`)}`);

  return lines.join("\n");
}

// ─── Syntax highlighter for init/fix plain text output ────────────────────
function highlight(raw: string): string {
  const safe = raw.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  return safe
    .replace(/(\bcharter\b)/g, `<span style="color:${T.textInfo};font-weight:600">$1</span>`)
    .replace(/\b(AE-[A-Z]+-\d+)\b/g, `<span style="color:${T.textInfo};font-weight:600">$1</span>`)
    .replace(/\b(BLOCKER)\b/g, `<span style="color:${T.textDanger};font-weight:600">$1</span>`)
    .replace(/\b(HIGH)\b/g, `<span style="color:${T.textWarning};font-weight:600">$1</span>`)
    .replace(/\b(MEDIUM)\b/g, `<span style="color:#fcd34d">$1</span>`)
    .replace(/\b(LOW)\b/g, `<span style="color:${T.textInfo}">$1</span>`)
    .replace(/^(\+[^+].*)$/gm, `<span style="color:${T.textSuccess}">$1</span>`)
    .replace(/^(-[^-].*)$/gm, `<span style="color:${T.textDanger}">$1</span>`)
    .replace(/^(@@.+@@.*)$/gm, `<span style="color:${T.textInfo}">$1</span>`)
    .replace(/(✓)/g, `<span style="color:${T.textSuccess}">$1</span>`)
    .replace(/(✗|✘)/g, `<span style="color:${T.textDanger}">$1</span>`)
    .replace(/(─{3,}|-{3,})/g, `<span style="color:#21262d">$1</span>`)
    .replace(/(›)/g, `<span style="color:${T.textInfo}">$1</span>`);
}

// ─── HTML wrapper ──────────────────────────────────────────────────────────
// NOTE: .wrap has overflow:visible + large padding so border-radius renders
// fully without Playwright clipping the shadow/corners.
function html(cwd: string, cmd: string, body: string): string {
  return `<!DOCTYPE html><html><head><meta charset="utf-8">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:ital,wght@0,400;0,500;0,600;1,400&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
/* Transparent page — Playwright omitBackground:true makes corners transparent */
html,body{background:transparent}
/* .wrap = transparent padding container so box-shadow has room to render */
.wrap{display:inline-block;padding:48px 52px;background:transparent}
.win{
  width:760px;border-radius:12px;overflow:hidden;
  border:1px solid rgba(255,255,255,0.08);
  box-shadow:
    0 0 0 1px rgba(255,255,255,0.04),
    0 2px 4px rgba(0,0,0,0.6),
    0 8px 20px rgba(0,0,0,0.55),
    0 24px 52px rgba(0,0,0,0.45),
    0 52px 100px rgba(0,0,0,0.35),
    0 0 80px rgba(37,99,235,0.18)
}
.bar{
  height:40px;padding:0 16px;display:flex;align-items:center;
  background:#161b22;border-bottom:1px solid rgba(255,255,255,0.07);
  box-shadow:inset 0 1px 0 rgba(255,255,255,0.05)
}
.lights{display:flex;gap:8px}
.d{width:12px;height:12px;border-radius:50%;flex-shrink:0}
.dr{background:#ff5f57;box-shadow:0 0 6px rgba(255,95,87,0.6)}
.dy{background:#febc2e;box-shadow:0 0 6px rgba(254,188,46,0.5)}
.dg{background:#28c840;box-shadow:0 0 6px rgba(40,200,64,0.5)}
.ttl{flex:1;text-align:center;margin-right:52px;font-family:'JetBrains Mono',ui-monospace,monospace;font-size:11.5px;color:#8b949e;letter-spacing:.05em}
.body{
  padding:18px 22px 24px;background:#0d1117;
  font-family:'JetBrains Mono','Cascadia Code',ui-monospace,monospace;
  font-size:13px;line-height:1.65;letter-spacing:.01em;
  font-feature-settings:'liga' 1,'calt' 1;color:#e6edf3
}
.prompt{margin-bottom:10px}
.pcwd{color:#60a5fa;font-size:12px}
.parr{color:#4d9de0;margin:0 4px}
.pcmd{color:#f0f6fc;font-weight:500}
.out{white-space:pre-wrap;word-break:break-word;tab-size:2}
</style></head><body>
<div class="wrap"><div class="win">
  <div class="bar">
    <div class="lights"><div class="d dr"></div><div class="d dy"></div><div class="d dg"></div></div>
    <div class="ttl">${cwd} — zsh</div>
  </div>
  <div class="body">
    <div class="prompt">
      <span class="pcwd">${cwd.replace(/</g,"&lt;")}</span>
      <span class="parr">&#10095;</span>
      <span class="pcmd">${cmd.replace(/</g,"&lt;")}</span>
    </div>
    <div class="out">${body}</div>
  </div>
</div></div>
</body></html>`;
}

// ─── Playwright ────────────────────────────────────────────────────────────
let browser: Awaited<ReturnType<typeof chromium.launch>> | null = null;
async function getBrowser() {
  if (!browser) browser = await chromium.launch();
  return browser;
}

// Playwright PNG buffer → sharp lossless WebP.
// Lossless WebP: bit-perfect quality like PNG, typically 25–35% smaller.
async function toWebp(buf: Buffer, outFile: string): Promise<void> {
  await sharp(buf).webp({ lossless: true }).toFile(outFile);
}

async function shot(htmlFile: string, outFile: string) {
  const b = await getBrowser();
  // deviceScaleFactor:2 — retina 2× pixel density
  const ctx = await b.newContext({ deviceScaleFactor: 2 });
  const p = await ctx.newPage();
  await p.setViewportSize({ width: 1200, height: 2000 });
  await p.emulateMedia({ colorScheme: "dark" });
  await p.goto(`file://${htmlFile}`);
  await p.waitForTimeout(1800); // Google Fonts load
  const el = await p.$(".wrap");
  // omitBackground:true → transparent page → border-radius corners are
  // transparent pixels in the WebP, not clipped to an opaque square.
  const buf = el
    ? await el.screenshot({ type: "png", omitBackground: true })
    : await p.screenshot({ type: "png", omitBackground: true, fullPage: false });
  await toWebp(buf, outFile);
  await ctx.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

async function shotUrl(url: string, outFile: string) {
  const b = await getBrowser();
  const ctx = await b.newContext({ deviceScaleFactor: 2 });
  const p = await ctx.newPage();
  await p.setViewportSize({ width: 1440, height: 900 });
  await p.emulateMedia({ colorScheme: "dark" });
  await p.goto(url);
  await p.waitForTimeout(2000);
  const buf = await p.screenshot({ type: "png", fullPage: false });
  await toWebp(buf, outFile);
  await ctx.close();
  console.log(`   ✓ ${outFile.split("/").pop()}`);
}

// ─── Jobs ──────────────────────────────────────────────────────────────────
const fnApiPath = `${HOME}/FN-Projects/fn-api-v3`;

// 1. doctor PASS
console.log("\n📸 1/5  doctor — 100/100 passing");
const passData = runCharterJson("doctor", "--path", repoRoot) as DoctorResult | null;
if (passData) {
  const body = renderDoctor(passData, "~/projects/my-platform");
  const f = join(tmpDir, "doctor-pass.html");
  writeFileSync(f, html("~/projects/my-platform", "charter doctor", body));
  await shot(f, join(screenshotsDir, "doctor-overview.webp"));
  await shot(f, join(screenshotsDir, "doctor-tty.webp"));
  await shot(f, join(screenshotsDir, "quickstart-scan.webp"));
}

// 2. doctor FAIL
if (existsSync(fnApiPath)) {
  console.log("\n📸 2/5  doctor — 59/100 failing");
  const failData = runCharterJson("doctor", "--path", fnApiPath) as DoctorResult | null;
  if (failData) {
    const body = renderDoctor(failData, "~/work/backend-api");
    const f = join(tmpDir, "doctor-fail.html");
    writeFileSync(f, html("~/work/backend-api", "charter doctor", body));
    await shot(f, join(screenshotsDir, "adopt-first-scan.webp"));
  }
}

// 3. fix --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 3/5  fix --dry-run");
  const raw = runCharter("fix", "--path", fnApiPath, "--dry-run");
  const f = join(tmpDir, "fix.html");
  writeFileSync(f, html("~/work/backend-api", "charter fix --dry-run", highlight(raw)));
  await shot(f, join(screenshotsDir, "fix-dry-run.webp"));
}

// 4. init --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 4/5  init --dry-run");
  const raw = runCharter("init", "--path", fnApiPath, "--dry-run");
  const f = join(tmpDir, "init.html");
  writeFileSync(f, html("~/work/backend-api", "charter init --dry-run", highlight(raw)));
  await shot(f, join(screenshotsDir, "init-output.webp"));
}

// 5. explain AE-CTX-001
console.log("\n📸 5/7  explain AE-CTX-001");
{
  const raw = runCharter("explain", "AE-CTX-001");
  const f = join(tmpDir, "explain.html");
  writeFileSync(f, html("~/projects/my-platform", "charter explain AE-CTX-001", highlight(raw)));
  await shot(f, join(screenshotsDir, "explain-output.webp"));
}

// 6. suppress --dry-run
if (existsSync(fnApiPath)) {
  console.log("\n📸 6/7  suppress --dry-run");
  const raw = runCharter(
    "suppress", "AE-CI-002",
    "--reason", "CI integration scheduled for next sprint",
    "--expires", "90d",
    "--path", fnApiPath,
    "--dry-run",
  );
  const f = join(tmpDir, "suppress.html");
  writeFileSync(f, html("~/work/backend-api", 'charter suppress AE-CI-002 --reason "..." --expires 90d --dry-run', highlight(raw)));
  await shot(f, join(screenshotsDir, "suppress-output.webp"));
}

// 7. version
console.log("\n📸 7/7  version");
{
  const raw = runCharter("version");
  const f = join(tmpDir, "version.html");
  writeFileSync(f, html("~/projects/my-platform", "charter version", highlight(raw)));
  await shot(f, join(screenshotsDir, "version-output.webp"));
}

// 8. HTML report
console.log("\n📸 8/8  HTML report");
const reportPath = join(tmpDir, "charter-report.html");
runCharter("report", "--path", repoRoot, "--out", reportPath);
if (existsSync(reportPath)) {
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-html.webp"));
  await shotUrl(`file://${reportPath}`, join(screenshotsDir, "report-overview.webp"));
}

if (browser != null) await (browser as Awaited<ReturnType<typeof chromium.launch>>).close();

console.log("\n✅ Done:");
for (const f of readdirSync(screenshotsDir)) {
  if (f.endsWith(".webp")) {
    const kb = Math.round(statSync(join(screenshotsDir, f)).size / 1024);
    console.log(`  ${f.padEnd(40)} ${kb}kb`);
  }
}
