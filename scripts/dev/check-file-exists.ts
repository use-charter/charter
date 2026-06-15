import { existsSync } from 'node:fs';

const target = process.argv[2];

if (!target) {
  console.error('missing file path argument');
  process.exit(2);
}

process.exit(existsSync(target) ? 0 : 1);
