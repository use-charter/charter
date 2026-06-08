import { spawnSync } from 'node:child_process';
import { exitWithStatus, readStdoutText, resolveRepoRoot, runTextCommand } from './lib/process.ts';

process.chdir(resolveRepoRoot());

const cgo = runTextCommand('go', ['env', 'CGO_ENABLED']);

if (cgo.status !== 0) {
  exitWithStatus(cgo.status);
}

const enabledRace = readStdoutText(cgo.stdout).trim() === '1';
const args = enabledRace
  ? ['test', '-tags=perf', '-run', 'TestDoctorPerformance', '-count=1', '-race', './internal/perf']
  : ['test', '-tags=perf', '-run', 'TestDoctorPerformance', '-count=1', './internal/perf'];

const result = spawnSync('go', args, { stdio: 'inherit' });

exitWithStatus(result.status);
