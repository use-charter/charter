// Generates the read-only HTML mirrors of the canonical internal Markdown docs
// (architecture, audit checklist). The Markdown is the single source of truth;
// these HTML files are presentation-only renders.
//
// Output is deterministic — no timestamps — and `marked` is pinned in
// package.json/bun.lock, so the same Markdown always renders to the same HTML.
// `moon run docs-html` (this script with `--check`, wired into `:docs`) fails
// the build if a mirror drifts from its source; regenerate with
// `bun scripts/generate-doc-html.ts`.
import { readFileSync, writeFileSync } from 'node:fs';
import { marked } from 'marked';
import { resolveRepoRoot } from './lib/process.ts';

interface DocPair {
  md: string;
  html: string;
  title: string;
}

const DOCS: DocPair[] = [
  {
    md: 'docs/internal/architecture/charter-architecture-2026.md',
    html: 'docs/internal/architecture/charter-architecture-2026.html',
    title: 'Charter Architecture 2026',
  },
  {
    md: 'docs/internal/audit/charter-v1-audit-checklist.md',
    html: 'docs/internal/audit/charter-v1-audit-checklist.html',
    title: 'Charter v1 Audit Checklist',
  },
];

marked.setOptions({ gfm: true, breaks: false });

const CSS = `:root { color-scheme: light dark; }
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  line-height: 1.6; color: #1f2328; background: #ffffff;
  margin: 0; padding: 2.5rem 1.25rem;
}
main.md { max-width: 56rem; margin: 0 auto; }
h1, h2, h3, h4 { line-height: 1.25; margin-top: 1.8em; margin-bottom: 0.6em; font-weight: 600; }
h1 { font-size: 2rem; border-bottom: 1px solid #d0d7de; padding-bottom: 0.3em; }
h2 { font-size: 1.5rem; border-bottom: 1px solid #d0d7de; padding-bottom: 0.3em; }
h3 { font-size: 1.2rem; } h4 { font-size: 1rem; }
a { color: #0969da; text-decoration: none; } a:hover { text-decoration: underline; }
p, ul, ol, table, pre, blockquote { margin: 0 0 1rem; }
code { font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace; font-size: 0.88em;
  background: rgba(129,139,152,0.12); padding: 0.15em 0.35em; border-radius: 4px; }
pre { background: #f6f8fa; padding: 1rem; border-radius: 6px; overflow: auto; }
pre code { background: none; padding: 0; font-size: 0.85em; }
table { border-collapse: collapse; width: 100%; display: block; overflow: auto; }
th, td { border: 1px solid #d0d7de; padding: 0.45rem 0.7rem; text-align: left; vertical-align: top; }
th { background: #f6f8fa; font-weight: 600; }
blockquote { border-left: 4px solid #d0d7de; color: #57606a; padding: 0 1rem; }
hr { border: 0; border-top: 1px solid #d0d7de; margin: 2rem 0; }
@media (prefers-color-scheme: dark) {
  body { color: #e6edf3; background: #0d1117; }
  h1, h2, hr { border-color: #30363d; }
  a { color: #4493f8; }
  code { background: rgba(110,118,129,0.4); }
  pre { background: #161b22; }
  th, td { border-color: #30363d; } th { background: #161b22; }
  blockquote { border-color: #30363d; color: #8b949e; }
}`;

const escapeHtml = (s: string): string =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

const render = (doc: DocPair, mdText: string): string => {
  const body = marked.parse(mdText) as string;
  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${escapeHtml(doc.title)}</title>
<!-- GENERATED from ${doc.md} by scripts/generate-doc-html.ts — do not edit by hand. -->
<style>
${CSS}
</style>
</head>
<body>
<main class="md">
${body}</main>
</body>
</html>
`;
};

const checkOnly = process.argv.slice(2).includes('--check');
process.chdir(resolveRepoRoot());

const drifted: string[] = [];
for (const doc of DOCS) {
  const html = render(doc, readFileSync(doc.md, 'utf8'));
  if (!checkOnly) {
    writeFileSync(doc.html, html);
    console.log(`generate-doc-html: wrote ${doc.html}`);
    continue;
  }
  let current = '';
  try {
    current = readFileSync(doc.html, 'utf8');
  } catch {
    current = '';
  }
  if (current !== html) {
    drifted.push(doc.html);
  }
}

if (checkOnly) {
  if (drifted.length > 0) {
    console.error(
      'generate-doc-html: HTML mirror(s) out of date — regenerate with ' +
        '`bun scripts/generate-doc-html.ts`:\n  ' +
        drifted.join('\n  '),
    );
    process.exit(1);
  }
  console.log('generate-doc-html: PASS — HTML mirrors match their Markdown source.');
}
