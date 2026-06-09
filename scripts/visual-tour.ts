import { join } from 'node:path';

import { repoRoot, runStep, writeSection } from './lib/full-test.ts';

console.log('== Charter visual terminal tour ==');
console.log('This script is for human visual inspection of the styled CLI/TUI surfaces.');
console.log('Commands below run directly in your terminal so you can inspect the real styled output.');
console.log('The interactive TUI and browser-open step are manual so they get your full terminal/session.');

writeSection('Prep');
console.log('Recommended first:');
console.log('  mise install');
console.log('  bun install --frozen-lockfile');
console.log('  mise exec -- moon run :check');

writeSection('Styled doctor output');
console.log('Running: go run ./cmd/charter doctor --path .');
runStep('charter doctor --path .', 'go', ['run', './cmd/charter', 'doctor', '--path', '.']);

writeSection('Explain output');
console.log('Running: go run ./cmd/charter explain AE-MCP-001');
runStep('charter explain AE-MCP-001', 'go', ['run', './cmd/charter', 'explain', 'AE-MCP-001']);

writeSection('JSON/markdown fallback spot check');
console.log('Running: go run ./cmd/charter doctor --path . --format json');
runStep('charter doctor json', 'go', ['run', './cmd/charter', 'doctor', '--path', '.', '--format', 'json']);
console.log('Running: go run ./cmd/charter doctor --path . --format markdown');
runStep('charter doctor markdown', 'go', ['run', './cmd/charter', 'doctor', '--path', '.', '--format', 'markdown']);

writeSection('Interactive TUI');
console.log('Run this manually in your terminal to inspect the full interactive UI:');
console.log('  go run ./cmd/charter doctor --path . -i');
console.log('Review: header, scorecard, finding navigation, filters, search, rescan, help, quit.');

writeSection('HTML report');
console.log('Run this manually to inspect the HTML report in your browser:');
console.log('  go run ./cmd/charter report --path . --format html --open');

writeSection('Visual tour complete');
console.log('Styled CLI/TUI/HTML review steps completed.');
console.log(`Default report path: ${join(repoRoot, 'charter-report.html')}`);
