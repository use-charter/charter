import { spawnSync } from 'node:child_process';
import { exitWithStatus, readStdoutText, resolveRepoRoot, runTextCommand } from '../lib/process.ts';

process.chdir(resolveRepoRoot());

const cgo = runTextCommand('go', ['env', 'CGO_ENABLED']);

if (cgo.status !== 0) {
  exitWithStatus(cgo.status);
}

const args = readStdoutText(cgo.stdout).trim() === '1' ? ['test', '-race', './...'] : ['test', './...'];
const result = spawnSync('go', args, { stdio: 'inherit' });

exitWithStatus(result.status);
