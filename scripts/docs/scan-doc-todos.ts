import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';

const roots = process.argv.slice(2);
const matcher = /TODO\((docs|spec)\)/;
const seen = new Set<string>();
let findingCount = 0;

const scan = (target: string) => {
  if (!existsSync(target) || seen.has(target)) return;

  seen.add(target);

  const stat = statSync(target);
  if (stat.isDirectory()) {
    for (const entry of readdirSync(target)) {
      if (entry === '.git') continue;
      scan(path.join(target, entry));
    }
    return;
  }

  if (!target.endsWith('.md')) return;

  const lines = readFileSync(target, 'utf8').split(/\r?\n/);
  lines.forEach((line, index) => {
    if (matcher.test(line)) {
      findingCount += 1;
      console.log(`${target}:${index + 1}:${line}`);
    }
  });
};

for (const root of roots) scan(root);

if (findingCount > 0) {
  console.error(`scan-doc-todos: found ${findingCount} unresolved TODO(docs|spec) marker(s)`);
  process.exit(1);
}
