// Generates internal/render/html/assets/fonts.css — the base64-embedded @font-face
// blocks the offline HTML report uses for its brand typefaces (DESIGN-TOKENS.md,
// ADR-0025). Each face is a Latin-subset woff2 fetched once from the Google Fonts
// CSS2 API (which serves woff2 + the `latin` unicode-range to a modern browser
// User-Agent), then inlined as `url(data:font/woff2;base64,…)`. The report ships
// the *committed* fonts.css — NOT this script's network fetch — so the report
// stays 100% self-contained and offline at render/view time.
//
// Regenerate (requires network) with:
//   bun scripts/generate-report-fonts.ts
//
// Families/weights mirror DESIGN-TOKENS.md "Offline / self-contained constraint".
// All families are SIL OFL 1.1 / Apache-2.0 → embeddable; notices are vendored in
// internal/render/html/assets/fonts/. This step is intentionally NOT wired into the
// offline build/lint (it needs the network); the base64 artifact is what is verified.
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { resolveRepoRoot } from '../lib/process.ts';

// A modern desktop Chrome UA so the CSS2 API serves woff2 (the most compact format)
// and the `latin` subset blocks we want.
const USER_AGENT =
  'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 ' +
  '(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';

// Only the `latin` subset is embedded (ASCII + Latin-1 + common typographic punctuation,
// which covers every glyph the report renders). Other subsets (latin-ext, cyrillic,
// vietnamese, …) are dropped to keep the payload bounded.
const SUBSET = 'latin';

interface Family {
  // The CSS2 `family=` query value (weights pinned to exactly what the report uses).
  query: string;
  // Human label for the generated section comment.
  label: string;
  // CSS var(s) the family backs, for the section comment.
  role: string;
}

const FAMILIES: Family[] = [
  { query: 'Ruda:wght@400;500;700;800', label: 'Ruda', role: '--font-site / --font-audit (body, headings, badges, wordmark)' },
  // 400 (body code/CLI/evidence) + 700 (hero score numerals, metric values). Atkinson is a
  // variable font, so the single woff2 must declare a 400–700 range; 500 falls within it.
  { query: 'Atkinson+Hyperlegible+Mono:wght@400;700', label: 'Atkinson Hyperlegible Mono', role: '--font-code (CLI, code, paths — primary mono; 700 for hero score)' },
  { query: 'IBM+Plex+Mono:wght@400;500', label: 'IBM Plex Mono', role: '--font-meta (rule IDs, metadata)' },
  { query: 'Share Tech', label: 'Share Tech', role: '--font-accent (system labels, status accents)' },
];

interface Face {
  family: string;
  style: string;
  weight: string;
  unicodeRange: string;
  woff2Url: string;
}

// A FaceGroup collapses faces that resolve to the *same* woff2 file. Google serves a
// single variable woff2 for every requested weight of a variable family (Ruda, Atkinson
// Hyperlegible Mono), so without grouping we would embed the identical file N times.
// Grouping emits one @font-face with a `font-weight: <min> <max>` range the browser
// instances from the variable font. Static families (IBM Plex Mono, Share Tech) have a
// distinct file per weight, so each lands in its own single-weight group.
interface FaceGroup {
  family: string;
  style: string;
  minWeight: number;
  maxWeight: number;
  unicodeRange: string;
  woff2Url: string;
}

const fetchCss = async (query: string): Promise<string> => {
  const url = `https://fonts.googleapis.com/css2?family=${query}&display=swap`;
  const res = await fetch(url, { headers: { 'User-Agent': USER_AGENT } });
  if (!res.ok) {
    throw new Error(`CSS2 fetch failed for ${query}: ${res.status} ${res.statusText}`);
  }
  return res.text();
};

// Parse the CSS2 response into faces, keeping only the wanted subset. Each block is
// preceded by a `/* <subset> */` comment that names the unicode-range it serves.
const parseFaces = (css: string): Face[] => {
  const blockRe = /\/\*\s*([\w-]+)\s*\*\/\s*(@font-face\s*\{[^}]*\})/g;
  const faces: Face[] = [];
  for (const m of css.matchAll(blockRe)) {
    const subset = m[1];
    const block = m[2];
    if (subset !== SUBSET || !block) continue;
    const family = /font-family:\s*'([^']+)'/.exec(block)?.[1];
    const style = /font-style:\s*([\w-]+)/.exec(block)?.[1];
    const weight = /font-weight:\s*([\d ]+)/.exec(block)?.[1]?.trim();
    const woff2Url = /src:\s*url\(([^)]+)\)\s*format\('woff2'\)/.exec(block)?.[1];
    const unicodeRange = /unicode-range:\s*([^;]+);/.exec(block)?.[1]?.trim();
    if (!family || !style || !weight || !woff2Url || !unicodeRange) {
      throw new Error(`unparseable @font-face block: ${block}`);
    }
    faces.push({ family, style, weight, unicodeRange, woff2Url });
  }
  return faces;
};

// Collapse faces backed by the same woff2 URL into one group spanning their weights.
const groupFaces = (faces: Face[]): FaceGroup[] => {
  const byUrl = new Map<string, FaceGroup>();
  for (const f of faces) {
    const w = Number(f.weight);
    const existing = byUrl.get(f.woff2Url);
    if (existing) {
      existing.minWeight = Math.min(existing.minWeight, w);
      existing.maxWeight = Math.max(existing.maxWeight, w);
    } else {
      byUrl.set(f.woff2Url, {
        family: f.family,
        style: f.style,
        minWeight: w,
        maxWeight: w,
        unicodeRange: f.unicodeRange,
        woff2Url: f.woff2Url,
      });
    }
  }
  // Deterministic order: by min weight ascending.
  return [...byUrl.values()].sort((a, b) => a.minWeight - b.minWeight);
};

const fetchWoff2Base64 = async (url: string): Promise<string> => {
  const res = await fetch(url, { headers: { 'User-Agent': USER_AGENT } });
  if (!res.ok) {
    throw new Error(`woff2 fetch failed for ${url}: ${res.status} ${res.statusText}`);
  }
  const buf = Buffer.from(await res.arrayBuffer());
  return buf.toString('base64');
};

const renderGroup = (g: FaceGroup, b64: string): string => {
  const weight = g.minWeight === g.maxWeight ? `${g.minWeight}` : `${g.minWeight} ${g.maxWeight}`;
  return (
    `@font-face{` +
    `font-family:'${g.family}';` +
    `font-style:${g.style};` +
    `font-weight:${weight};` +
    `font-display:swap;` +
    `src:url(data:font/woff2;base64,${b64}) format('woff2');` +
    `unicode-range:${g.unicodeRange};` +
    `}`
  );
};

const main = async (): Promise<void> => {
  process.chdir(resolveRepoRoot());

  const sections: string[] = [];
  let totalWoff2 = 0;

  for (const fam of FAMILIES) {
    const css = await fetchCss(fam.query);
    const faces = parseFaces(css);
    if (faces.length === 0) {
      throw new Error(`no '${SUBSET}' subset faces found for ${fam.label}`);
    }
    const groups = groupFaces(faces);

    const rendered: string[] = [];
    for (const g of groups) {
      const b64 = await fetchWoff2Base64(g.woff2Url);
      totalWoff2 += Math.ceil((b64.length * 3) / 4);
      rendered.push(renderGroup(g, b64));
    }
    const weights = faces.map((f) => f.weight).join('/');
    const fileCount = groups.length;
    sections.push(`/* ${fam.label} — ${SUBSET} subset, weights ${weights} — ${fam.role} */\n${rendered.join('\n')}`);
    console.log(`generate-report-fonts: ${fam.label} → ${faces.length} weight(s) in ${fileCount} file(s) [${weights}]`);
  }

  // NOTE: this comment is inlined verbatim into the report's <style>. The
  // self-containment test (render_test.go) rejects the substrings "url(http",
  // "url('http", "url(\"http" — so describe the no-remote-fetch guarantee WITHOUT
  // ever writing a literal remote url(...) token here.
  const header =
    `/* Charter HTML report — embedded brand fonts (GENERATED, do not edit by hand).\n` +
    `   Regenerate with: bun scripts/generate-report-fonts.ts\n` +
    `   Each face is a Latin-subset woff2 from the Google Fonts CSS2 API, inlined as a\n` +
    `   base64 data: URI so the report is 100% self-contained + offline (ADR-0025): no\n` +
    `   CDN, no remote font URLs, no external fetch at render or view time. Families are\n` +
    `   SIL OFL 1.1 / Apache-2.0 — license notices vendored in assets/fonts/. */\n`;

  const out = `${header}\n${sections.join('\n\n')}\n`;
  const dest = join('internal', 'render', 'html', 'assets', 'fonts.css');
  mkdirSync(join('internal', 'render', 'html', 'assets'), { recursive: true });
  writeFileSync(dest, out);

  const kb = (n: number): string => `${(n / 1024).toFixed(1)} KiB`;
  console.log(`generate-report-fonts: wrote ${dest}`);
  console.log(`generate-report-fonts: woff2 payload ~${kb(totalWoff2)}; fonts.css ${kb(Buffer.byteLength(out))}`);
};

main().catch((err) => {
  console.error(`generate-report-fonts: ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
