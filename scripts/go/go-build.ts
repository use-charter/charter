import { mkdirSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { exitWithStatus, resolveRepoRoot } from '../lib/process.ts';

process.chdir(resolveRepoRoot());

mkdirSync('dist', { recursive: true });

const result = spawnSync('go', ['build', '-o', 'dist/charter', './cmd/charter'], {
  stdio: 'inherit',
});

exitWithStatus(result.status);
