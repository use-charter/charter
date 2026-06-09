import { join } from 'node:path';
import {
  cleanDirectory,
  defaultTempRoot,
  newDemoRepo,
  newFixtureGitRepo,
  repoRoot,
  runCharter,
  runStep,
  writeSection,
} from './lib/full-test.ts';

const tempRoot = process.argv[2] ? join(process.argv[2]) : defaultTempRoot;

console.log('== Charter command tour ==');
console.log('Behavior verification only: this script checks command correctness, exit behavior, fixtures, and gates.');
console.log('It does not verify the final styled TTY/TUI presentation layer; use scripts/visual-tour.ts for that.');

const demo = join(tempRoot, 'demo-repo');
const fixtureRoot = join(tempRoot, 'fixtures');

writeSection('Prep');
cleanDirectory(tempRoot);
cleanDirectory(fixtureRoot);
runStep('mise trust', 'mise', ['trust']);
runStep('mise install', 'mise', ['install']);
runStep('bun install --frozen-lockfile', 'bun', ['install', '--frozen-lockfile']);

writeSection('Version');
runStep('charter version', 'go', ['run', './cmd/charter', 'version']);
runStep('charter version --short', 'go', ['run', './cmd/charter', 'version', '--short']);
runStep('charter version --format json', 'go', ['run', './cmd/charter', 'version', '--format', 'json']);

writeSection('Explain');
runStep('charter explain AE-CTX-001', 'go', ['run', './cmd/charter', 'explain', 'AE-CTX-001']);
runStep('charter explain AE-MCP-001 --format json', 'go', [
  'run',
  './cmd/charter',
  'explain',
  'AE-MCP-001',
  '--format',
  'json',
]);

writeSection('Doctor on repo (text)');
runStep('charter doctor --path .', 'go', ['run', './cmd/charter', 'doctor', '--path', '.']);

writeSection('Doctor on repo (json)');
console.log(runCharter(['doctor', '--path', '.', '--format', 'json']).stdout);

writeSection('Doctor on repo (markdown)');
console.log(runCharter(['doctor', '--path', '.', '--format', 'markdown']).stdout);

writeSection('Doctor on repo (single rule)');
runStep('charter doctor single rule', 'go', ['run', './cmd/charter', 'doctor', '--path', '.', '--rule', 'AE-CTX-001']);

writeSection('Doctor on fixtures');
const passFixture = newFixtureGitRepo(join(repoRoot, 'testdata', 'repos', 'pass-slice1'), fixtureRoot, 'pass-slice1');
const mcpFixture = newFixtureGitRepo(join(repoRoot, 'testdata', 'repos', 'fail-mcp-unpinned'), fixtureRoot, 'fail-mcp-unpinned');
const ccFixture = newFixtureGitRepo(join(repoRoot, 'testdata', 'repos', 'fail-cc-dangerous-hook'), fixtureRoot, 'fail-cc-dangerous-hook');
console.log(runCharter(['doctor', '--path', passFixture, '--format', 'json'], { allowedExitCodes: [0, 1] }).stdout);
console.log(runCharter(['doctor', '--path', mcpFixture], { allowedExitCodes: [0, 1] }).stdout);
console.log(runCharter(['doctor', '--path', ccFixture], { allowedExitCodes: [0, 1] }).stdout);

writeSection('Report on repo');
runStep('charter report html', 'go', ['run', './cmd/charter', 'report', '--path', '.', '--format', 'html']);
runStep('charter report json', 'go', ['run', './cmd/charter', 'report', '--path', '.', '--format', 'json']);
runStep('charter report markdown', 'go', ['run', './cmd/charter', 'report', '--path', '.', '--format', 'markdown']);

writeSection('Build demo repo for init/fix/suppress');
newDemoRepo(demo);

writeSection('Init (dry-run)');
runStep('charter init dry-run', 'go', ['run', './cmd/charter', 'init', '--dry-run', '--path', demo]);

writeSection('Init (real)');
runStep('charter init', 'go', ['run', './cmd/charter', 'init', '--path', demo]);

writeSection('Doctor on demo repo after init');
console.log(runCharter(['doctor', '--path', demo], { allowedExitCodes: [0, 1] }).stdout);

writeSection('Fix (dry-run) on demo repo');
runStep('charter fix dry-run', 'go', ['run', './cmd/charter', 'fix', '--path', demo, '--dry-run']);

writeSection('Suppress (dry-run) on repo');
runStep('charter suppress dry-run', 'go', [
  'run',
  './cmd/charter',
  'suppress',
  'AE-MCP-001',
  '--reason',
  'review-test',
  '--dry-run',
  '--path',
  '.',
]);

writeSection('Perf and full verification');
runStep('moon run :perf', 'mise', ['exec', '--', 'moon', 'run', ':perf']);
runStep('moon run :check', 'mise', ['exec', '--', 'moon', 'run', ':check']);

writeSection('Optional interactive/TUI step');
console.log('Run this manually in your terminal to see the full interactive UI and styled TTY surface:');
console.log('go run ./cmd/charter doctor --path . -i');

writeSection('Optional HTML open step');
console.log('Open the default generated report in the repo root:');
console.log(join(repoRoot, 'charter-report.html'));

writeSection('Tour complete');
console.log('Raw command output was printed directly for inspection.');
console.log(`Temp workspace kept at: ${tempRoot}`);
