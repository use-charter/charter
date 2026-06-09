import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { basename, join, resolve } from 'node:path';

type CatalogEntry = {
  id: string;
  name: string;
  category: string;
  shortDescription: string;
};

type RuleSpec = {
  id: string;
  fields: Map<string, string>;
};

const repoRoot = resolve(import.meta.dirname, '..');
const specsDir = join(repoRoot, 'docs', 'internal', 'specs');
const outputDir = join(repoRoot, 'docs', 'product', 'rules');
const catalogPath = join(repoRoot, 'internal', 'rules', 'catalog', 'catalog.go');

const checkMode = process.argv.includes('--check');

function main(): void {
  ensureOutputDir();

  const catalog = loadCatalogEntries();
  const specs = loadSpecs();

  const errors: string[] = [];
  const written: string[] = [];

  for (const spec of specs) {
    const outPath = join(outputDir, `${spec.id}.mdx`);
    const entry = catalog.get(spec.id);

    if (!entry) {
      errors.push(`Catalog entry missing for ${spec.id}`);
      continue;
    }

    const content = renderRulePage(spec, entry);
    if (checkMode) {
      if (!existsSync(outPath)) {
        errors.push(`Missing generated page: ${relativeToRepo(outPath)}`);
        continue;
      }
      const existing = readFileSync(outPath, 'utf8');
      if (normalizeNewlines(existing) !== normalizeNewlines(content)) {
        errors.push(`Drifted generated page: ${relativeToRepo(outPath)}`);
      }
      continue;
    }

    writeFileSync(outPath, content);
    written.push(relativeToRepo(outPath));
  }

  for (const id of catalog.keys()) {
    const outPath = join(outputDir, `${id}.mdx`);
    if (!existsSync(outPath) && checkMode) {
      errors.push(`Catalog helpUri target missing: ${relativeToRepo(outPath)}`);
    }
    if (!specs.some((spec) => spec.id === id)) {
      errors.push(`Spec missing for catalog entry ${id}`);
    }
  }

  if (errors.length > 0) {
    for (const error of errors) {
      console.error(error);
    }
    process.exit(1);
  }

  if (checkMode) {
    console.log(`Rule pages OK: ${specs.length} generated pages match internal specs and catalog coverage.`);
    return;
  }

  console.log(`Generated ${written.length} rule pages.`);
}

function ensureOutputDir(): void {
  if (!existsSync(outputDir)) {
    mkdirSync(outputDir, { recursive: true });
  }
}

function loadCatalogEntries(): Map<string, CatalogEntry> {
  const src = readFileSync(catalogPath, 'utf8');
  const entries = new Map<string, CatalogEntry>();
  const entryPattern = /"(AE-[A-Z]+-\d+)":\s*\{ID:\s*"[^"]+",\s*Name:\s*"([^"]+)",\s*Category:\s*"([^"]+)",\s*ShortDescription:\s*"([^"]+)"/g;

  for (const match of src.matchAll(entryPattern)) {
    const id = match[1];
    const name = match[2];
    const category = match[3];
    const shortDescription = match[4];
    if (!id || !name || !category || !shortDescription) {
      continue;
    }
    entries.set(id, { id, name, category, shortDescription });
  }

  return entries;
}

function loadSpecs(): RuleSpec[] {
  return readdirSync(specsDir)
    .filter((file) => /^AE-[A-Z]+-\d+\.md$/.test(file))
    .sort()
    .map((file) => parseSpec(join(specsDir, file)));
}

function parseSpec(filePath: string): RuleSpec {
  const src = normalizeNewlines(readFileSync(filePath, 'utf8')).trim();
  const lines = src.split('\n');
  const id = basename(filePath, '.md');
  const fields = new Map<string, string>();
  let currentKey = '';

  for (const line of lines.slice(1)) {
    const bullet = line.match(/^- ([^:]+):\s*(.*)$/);
    if (bullet) {
      const rawKey = bullet[1];
      const rawValue = bullet[2];
      if (!rawKey || rawValue === undefined) {
        continue;
      }
      currentKey = canonicalKey(rawKey);
      fields.set(currentKey, rawValue.trim());
      continue;
    }

    if (!currentKey) {
      continue;
    }

    if (line.startsWith('  ') || line.startsWith('\t')) {
      const existing = fields.get(currentKey) ?? '';
      fields.set(currentKey, `${existing}\n${line}`.trimEnd());
      continue;
    }

    if (line.trim() === '') {
      const existing = fields.get(currentKey) ?? '';
      fields.set(currentKey, `${existing}\n`.trimEnd());
    }
  }

  return { id, fields };
}

function canonicalKey(raw: string): string {
  return raw.toLowerCase().replace(/[^a-z0-9]+/g, ' ').trim();
}

function renderRulePage(spec: RuleSpec, entry: CatalogEntry): string {
  const description = firstSentence(field(spec, 'description') || entry.shortDescription);
  const severity = field(spec, 'severity');
  const category = field(spec, 'category') || entry.category;
  const sections: string[] = [];

  sections.push('---');
  sections.push(`title: "${escapeQuotes(spec.id)}"`);
  sections.push(`description: "${escapeQuotes(description)}"`);
  sections.push('---');
  sections.push('');
  sections.push(`> Generated from the internal rule spec. Edit \`docs/internal/specs/${spec.id}.md\`, then re-run \`bun scripts/generate-rule-pages.ts\`.`);
  sections.push('');
  sections.push(`**Rule name:** ${entry.name}`);
  sections.push('');
  sections.push(`**Severity:** ${severity || 'Unspecified'}  `);
  sections.push(`**Category:** ${category}`);
  sections.push('');
  sections.push(description);
  sections.push('');

  addSection(sections, 'Detection Logic', field(spec, 'detection logic'));
  addSection(sections, 'Pass Example', field(spec, 'pass example'));
  addSection(sections, 'Fail Example', field(spec, 'fail example'));
  addSection(sections, 'Evidence Expectations', field(spec, 'evidence expectations'));
  addSection(sections, 'Edge Cases', field(spec, 'edge cases'));
  addSection(sections, 'Remediation', field(spec, 'remediation'));
  addSection(sections, 'Scoring Impact', field(spec, 'scoring impact'));
  addSection(sections, 'Catalog Notes', field(spec, 'catalog slice 13 adr 0021'));

  const relatedAdrs = field(spec, 'related adrs');
  if (relatedAdrs) {
    sections.push('## Related ADRs');
    sections.push('');
    sections.push(listify(relatedAdrs));
    sections.push('');
  }

  const relatedEvals = field(spec, 'related evals');
  if (relatedEvals) {
    sections.push('## Related Evals');
    sections.push('');
    sections.push(listify(relatedEvals));
    sections.push('');
  }

  sections.push('## CLI');
  sections.push('');
  sections.push('```bash');
  sections.push(`charter explain ${spec.id}`);
  sections.push('```');
  sections.push('');

  return `${sections.join('\n').trimEnd()}\n`;
}

function addSection(target: string[], heading: string, body: string): void {
  if (!body) {
    return;
  }
  target.push(`## ${heading}`);
  target.push('');
  target.push(formatBody(body));
  target.push('');
}

function field(spec: RuleSpec, key: string): string {
  return spec.fields.get(canonicalKey(key)) ?? '';
}

function formatBody(body: string): string {
  return normalizeNewlines(body)
    .split('\n')
    .map((line) => line.trimEnd())
    .join('\n')
    .trim();
}

function listify(body: string): string {
  const trimmed = formatBody(body);
  if (trimmed === '' || trimmed === 'None yet') {
    return trimmed;
  }
  if (trimmed.startsWith('- ')) {
    return trimmed;
  }
  return trimmed
    .split(/,\s*/)
    .filter(Boolean)
    .map((item) => `- ${item}`)
    .join('\n');
}

function firstSentence(value: string): string {
  const trimmed = formatBody(value).replace(/\s+/g, ' ');
  return trimmed;
}

function escapeQuotes(value: string): string {
  return value.replace(/"/g, '&quot;');
}

function normalizeNewlines(value: string): string {
  return value.replace(/\r\n/g, '\n');
}

function relativeToRepo(filePath: string): string {
  return filePath.replace(`${repoRoot}\\`, '').replace(`${repoRoot}/`, '').split('\\').join('/');
}

main();
