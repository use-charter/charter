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

  // Scorecard — colored dot indicator per category
  if (d.categories.length > 0) {
    lines.push("");
    lines.push(c(T.textSecondary, "readiness by category"));
    for (const cat of d.categories) {
      const hasIssues = cat.findings > 0;
      const dotColor  = hasIssues ? sevFg(cat.worst_severity) : "#3fb950";
      const dot       = `<span style="color:${dotColor}">●</span>`;
      const name      = c(T.textTertiary, `  ${cat.category.padEnd(13)}`);
      const status    = hasIssues
        ? `${c(dotColor, `${cat.findings} finding${cat.findings > 1 ? "s" : ""}`, true)}  ${c(T.textTertiary, `worst ${cat.worst_severity}`)}`
        : c("#3fb950", "passed");
      lines.push(`  ${dot}${name} ${status}`);
    }
  }

  // Score hero — score number prominent, bar, verdict badge
  lines.push("");
  lines.push(DIVIDER);
  const scoreTok = d.passed ? T.textSuccess : T.textDanger;
  const verdict  = d.passed ? "PASS ✓" : "FAIL ✗";
  const scoreNum = `<span style="color:${scoreTok};font-weight:700;font-size:1.05em">${d.score.final}</span>${c(T.textTertiary, "/100")}`;
  const bar24    = scoreBar(d.score.final, scoreTok);
  const badge2   = `<span style="color:${scoreTok};font-weight:700">${verdict}</span>`;
  lines.push(`${c(T.textSecondary, "Score ")}${scoreNum}  ${bar24}  ${badge2}`);
  if (d.score.final < d.score.base) {
    lines.push(`      ${c(T.textDanger, `cap   score capped at ${d.score.final}`)}`);
  }
  lines.push(`      ${c(T.textTertiary, `threshold ${d.threshold}`)}`);

  return lines.join("\n");
}

// ─── Diff highlighter (for fix --dry-run) ────────────────────────────────
function highlightDiff(raw: string): string {
  const safe = raw.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  const DIV  = `<span style="color:#21262d">${"─".repeat(52)}</span>`;

  // Parse into rule blocks then re-render with styled headers
  const blocks: string[] = [];
  let current: string[] = [];
  let ruleCount = 0;

  for (const line of safe.split("\n")) {
    const ruleMatch = line.match(/^(AE-[A-Z]+-\d+)\s+(.+)$/);
    if (ruleMatch) {
      if (current.length) blocks.push(current.join("\n"));
      current = [];
      ruleCount++;
      const ruleId   = `<span style="color:${T.textInfo};font-weight:600">${ruleMatch[1]}</span>`;
      const filePath = `<span style="color:${T.textTertiary}">${ruleMatch[2]}</span>`;
      current.push(`${ruleId}  ${filePath}`);
    } else if (line.startsWith("+") && !line.startsWith("+++")) {
      current.push(`<span style="color:${T.textSuccess}">${line}</span>`);
    } else if (line.startsWith("-") && !line.startsWith("---")) {
      current.push(`<span style="color:${T.textDanger}">${line}</span>`);
    } else if (line.startsWith("@@")) {
      current.push(`<span style="color:#444d56">${line}</span>`);
    } else if (line.startsWith("---") || line.startsWith("+++")) {
      current.push(`<span style="color:#30363d">${line}</span>`);
    } else if (line.trim()) {
      current.push(`<span style="color:${T.textTertiary}">${line}</span>`);
    } else {
      current.push("");
    }
  }
  if (current.length) blocks.push(current.join("\n"));

  const footer = [
    DIV,
    `<span style="color:${T.textWarning}">▸</span>  <span style="color:${T.textWarning}">dry run</span>  <span style="color:${T.textTertiary}">·  ${ruleCount} fix${ruleCount !== 1 ? "es" : ""} ready  ·  </span><span style="color:${T.textInfo}">charter fix</span><span style="color:${T.textTertiary}"> to apply</span>`,
  ].join("\n");

  return `${blocks.join(`\n${DIV}\n`)}\n\n${footer}`;
}

// ─── explain renderer ─────────────────────────────────────────────────────
interface ExplainEntry { ID: string; Name: string; Category: string; ShortDescription: string; HelpURI: string; Severity?: string; }

function renderExplain(e: ExplainEntry): string {
  const DIV = c(T.textTertiary, "─".repeat(52));
  return [
    `${c(T.textInfo, e.ID, true)}  ${c(T.textPrimary, e.Name, true)}`,
    DIV,
    "",
    `  ${c(T.textTertiary, "category  ")}${c(T.textSecondary, e.Category)}`,
    "",
    `  ${c(T.textPrimary, e.ShortDescription)}`,
    "",
    DIV,
    `  ${c(T.textTertiary, "docs  ")}${c(T.textInfo, e.HelpURI)}`,
  ].join("\n");
}

// ─── version renderer ────────────────────────────────────────────────────
interface VersionData { version: string; commit: string; date: string; go: string; platform: string; }

function renderVersion(v: VersionData): string {
  const commit = v.commit.slice(0, 8);
  const built  = v.date.slice(0, 10);
  const sep    = c(T.textTertiary, "  ·  ");
  // Single headline: brand + version
  const headline = `${c(T.textInfo, "[C]", true)} ${c(T.textInfo, "charter", true)}  ${c(T.textPrimary, v.version, true)}`;
  // One compact metadata line — all context on a single dim row
  const meta = [
    c(T.textTertiary, `go ${v.go}`),
    c(T.textTertiary, v.platform),
    c(T.textTertiary, `commit ${commit}`),
    c(T.textTertiary, built),
  ].join(sep);
  return [headline, "", `  ${meta}`].join("\n");
}

// ─── init --dry-run renderer ──────────────────────────────────────────────
interface InitAction { path: string; action: string; note: string; }

function renderInit(displayName: string, actions: InitAction[]): string {
  const lines: string[] = [];
  const DIV = c(T.textTertiary, "─".repeat(44));

  lines.push(`${c(T.textInfo, "[C]", true)} ${c(T.textInfo, "charter", true)}${c(T.textTertiary, `  v1.0.0  ·  ${displayName}`)}`);
  lines.push("");
  lines.push(`${c(T.textSecondary, "Creating")}  ${c(T.textTertiary, "(dry run)")}`);

  for (const a of actions) {
    const name   = c("#79c0ff", a.path);
    const dots   = c(T.textTertiary, "·".repeat(Math.max(2, 22 - a.path.length)));
    const status = a.action === "skip"
      ? c(T.textTertiary, "skip")
      : c(T.textSuccess, "would create");
    const note   = c(T.textTertiary, `  ${a.note}`);
    lines.push(`  ${name} ${dots}  ${status}${note}`);
  }

  lines.push(DIV);
  const created = actions.filter(a => a.action !== "skip").length;
  const skipped = actions.filter(a => a.action === "skip").length;
  lines.push(
    `  ${c(T.textSuccess, `${created} file${created !== 1 ? "s" : ""} would be created`)}` +
    `  ${c(T.textTertiary, `·  ${skipped} skipped`)}`
  );
  lines.push(`  ${c(T.textInfo, "›")} ${c(T.textSecondary, "Run without --dry-run to apply")}`);
  return lines.join("\n");
}

// ─── suppress renderer ───────────────────────────────────────────────────
function renderSuppress(rule: string, reason: string, expires: string, dryRun = true): string {
  const DIV = c(T.textTertiary, "─".repeat(52));
  const kv  = (label: string, val: string) =>
    `  ${c(T.textTertiary, label.padEnd(10))}${c(T.textSecondary, val)}`;
  // Headline: brand + command + rule all on one row
  const headline = `${c(T.textInfo, "[C]", true)} ${c(T.textInfo, "charter", true)}  ${c(T.textTertiary, "suppress")}  ${c(T.textInfo, rule, true)}`;
  const outcome  = dryRun
    ? `  ${c(T.textWarning, "▸")} ${c(T.textWarning, "dry run")}${c(T.textTertiary, "  ·  .charter-suppress.yml not written  ·  remove --dry-run to apply")}`
    : `  ${c(T.textSuccess, "✓")} ${c(T.textSuccess, "written")}${c(T.textTertiary, "  .charter-suppress.yml")}`;
  return [
    headline,
    DIV,
    "",
    kv("reason",  reason),
    kv("expires", expires),
    "",
    DIV,
    outcome,
  ].join("\n");
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
  writeFileSync(f, html("~/work/backend-api", "charter fix --dry-run", highlightDiff(raw)));
  await shot(f, join(screenshotsDir, "fix-dry-run.webp"));
}

// 4. init --dry-run — parse real command output into styled renderer
if (existsSync(fnApiPath)) {
  console.log("\n📸 4/5  init --dry-run");
  const initRaw = runCharter("init", "--path", fnApiPath, "--dry-run");
  // parse "would create FILE" / "would skip FILE" lines
  const initActions: InitAction[] = initRaw
    .split("\n")
    .filter(l => l.startsWith("would ") || l.startsWith("create ") || l.startsWith("skip "))
    .map(l => {
      if (l.startsWith("would create ")) return { path: l.slice(13).trim(), action: "create", note: "" };
      if (l.startsWith("would skip "))   return { path: l.slice(11).trim(), action: "skip",   note: "" };
      if (l.startsWith("create "))       return { path: l.slice(7).trim(),  action: "create", note: "" };
      return { path: l.slice(5).trim(),  action: "skip", note: "" };
    });
  const f = join(tmpDir, "init.html");
  writeFileSync(f, html("~/work/backend-api", "charter init --dry-run", renderInit("~/work/backend-api", initActions)));
  await shot(f, join(screenshotsDir, "init-output.webp"));
}

// 5. explain — structured renderer matching explain.go textStyled
console.log("\n📸 5/7  explain AE-CTX-001");
{
  const entry = runCharterJson("explain", "AE-CTX-001") as ExplainEntry | null;
  if (entry) {
    // catalog intentionally has no Severity — don't inject one
    const f = join(tmpDir, "explain.html");
    writeFileSync(f, html("~/projects/my-platform", "charter explain AE-CTX-001", renderExplain(entry)));
    await shot(f, join(screenshotsDir, "explain-output.webp"));
  }
}

// 6. suppress --dry-run — structured renderer
console.log("\n📸 6/7  suppress AE-CI-002 --dry-run");
{
  const f = join(tmpDir, "suppress.html");
  writeFileSync(f, html(
    "~/work/backend-api",
    'charter suppress AE-CI-002 --reason "..." --expires 90d --dry-run',
    renderSuppress("AE-CI-002", "CI integration scheduled for next sprint", "2026-09-07 (90 days)"),
  ));
  await shot(f, join(screenshotsDir, "suppress-output.webp"));
}

// 7. version — structured renderer
console.log("\n📸 7/7  version");
{
  const vdata = runCharterJson("version") as VersionData | null;
  if (vdata) {
    // show a clean fictional version for docs (not a dev snapshot hash)
    const display: VersionData = {
      version:  "1.0.0",
      commit:   "abc1234f",
      date:     "2026-06-10T00:00:00Z",
      go:       vdata.go,
      platform: vdata.platform,
    };
    const f = join(tmpDir, "version.html");
    writeFileSync(f, html("~/projects/my-platform", "charter version", renderVersion(display)));
    await shot(f, join(screenshotsDir, "version-output.webp"));
  }
}

// 8. HTML report
console.log("\n📸 8/8  HTML report");
const reportPath = join(tmpDir, "charter-report.html");
runCharter("report", "--path", repoRoot, "--out", reportPath);
if (existsSync(reportPath)) {
  await shot(reportPath, join(screenshotsDir, "report-full.webp"));
  const b = await getBrowser();
  const ctx = await b.newContext({ deviceScaleFactor: 2 });
  const p = await ctx.newPage();
  await p.setViewportSize({ width: 1440, height: 900 });
  await p.emulateMedia({ colorScheme: "dark" });
  await p.goto(`file://${reportPath}`);
  await p.waitForTimeout(2000);
  const buf = await p.screenshot({ type: "png", fullPage: false });
  await toWebp(buf, join(screenshotsDir, "report-hero.webp"));
  await ctx.close();
  console.log(`   ✓ report-hero.webp`);
}

if (browser != null) await (browser as Awaited<ReturnType<typeof chromium.launch>>).close();

console.log("\n✅ Done:");
for (const f of readdirSync(screenshotsDir)) {
  if (f.endsWith(".webp")) {
    const kb = Math.round(statSync(join(screenshotsDir, f)).size / 1024);
    console.log(`  ${f.padEnd(40)} ${kb}kb`);
  }
}
