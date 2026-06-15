import { readFileSync } from 'node:fs';
import { resolveRepoRoot } from '../lib/process.ts';

const ACTION_PATH = 'action/action.yml';

const REQUIRED_INPUTS = [
  'version',
  'path',
  'threshold',
  'format',
  'sarif-file',
  'upload',
  'verify',
  'fail-below',
  'category',
] as const;

process.chdir(resolveRepoRoot());

const fail = (message: string): never => {
  console.error(`validate-action: ${message}`);
  process.exit(1);
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const requireRecord = (value: unknown, label: string): Record<string, unknown> => {
  if (isRecord(value)) {
    return value;
  }
  return fail(`${label} must be a YAML mapping.`);
};

const raw = readFileSync(ACTION_PATH, 'utf8');

if (raw.includes('gh release view')) {
  fail(
    'action/action.yml must not depend on the GitHub CLI for release resolution; ' +
      'use an authenticated REST call so the default path works on self-hosted runners too.',
  );
}

// Critical security assertion: zizmor/actionlint never see action.yml, so this
// script is the only authority that every `uses:` is pinned to a full commit SHA.
// Works purely on lines so it stays valid regardless of the YAML parser.
const usesLine = /^\s*(?:-\s*)?uses:\s*(.+?)\s*$/;
const shaPinned = /@[0-9a-f]{40}$/;

raw.split(/\r?\n/).forEach((line, index) => {
  const match = usesLine.exec(line);
  if (!match) return;

  const value = (match[1] ?? '')
    .split('#')[0]
    ?.trim()
    .replace(/^['"]|['"]$/g, '') ?? '';

  if (!shaPinned.test(value)) {
    fail(
      `unpinned action reference at action/action.yml:${index + 1}: "${line.trim()}" — ` +
        'pin to a full 40-char commit SHA (owner/repo@<sha>) instead of a tag/branch.',
    );
  }
});

// Structural assertions via Bun's built-in YAML parser (no extra dependency).
const doc = requireRecord(Bun.YAML.parse(raw), 'action/action.yml');

const runs = requireRecord(doc['runs'], 'runs:');
if (runs['using'] !== 'composite') {
  fail(`runs.using must be "composite" (found ${JSON.stringify(runs['using'])}).`);
}

for (const key of ['name', 'description'] as const) {
  const value = doc[key];
  if (typeof value !== 'string' || value.trim() === '') {
    fail(`missing or empty top-level \`${key}:\`.`);
  }
}

requireRecord(doc['branding'], 'branding:');

const inputs = requireRecord(doc['inputs'], 'inputs:');
for (const name of REQUIRED_INPUTS) {
  if (!(name in inputs)) {
    fail(`missing documented input \`${name}\`.`);
  }
  const spec = requireRecord(inputs[name], `input \`${name}\``);
  if (!('default' in spec)) {
    fail(`input \`${name}\` must declare a \`default:\`.`);
  }
}

console.log(
  `validate-action: PASS — composite action, ${REQUIRED_INPUTS.length} documented inputs, all uses: pinned to full SHAs.`,
);
