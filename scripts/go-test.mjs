import { spawnSync } from 'node:child_process';

const root = spawnSync('git', ['rev-parse', '--show-toplevel'], {
  stdio: ['ignore', 'pipe', 'inherit'],
  encoding: 'utf8',
});

if (root.status !== 0) {
  process.exit(root.status ?? 1);
}

process.chdir(root.stdout.trim());

const cgo = spawnSync('go', ['env', 'CGO_ENABLED'], {
  stdio: ['ignore', 'pipe', 'inherit'],
  encoding: 'utf8',
});

if (cgo.status !== 0) {
  process.exit(cgo.status ?? 1);
}

const args = cgo.stdout.trim() === '1' ? ['test', '-race', './...'] : ['test', './...'];
const result = spawnSync('go', args, { stdio: 'inherit' });

process.exit(result.status ?? 1);
