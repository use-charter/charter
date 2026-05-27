import { mkdirSync } from 'node:fs';
import { spawnSync } from 'node:child_process';

const root = spawnSync('git', ['rev-parse', '--show-toplevel'], {
  stdio: ['ignore', 'pipe', 'inherit'],
  encoding: 'utf8',
});

if (root.status !== 0) {
  process.exit(root.status ?? 1);
}

process.chdir(root.stdout.trim());

mkdirSync('dist', { recursive: true });

const result = spawnSync('go', ['build', '-o', 'dist/charter', './cmd/charter'], {
  stdio: 'inherit',
});

process.exit(result.status ?? 1);
