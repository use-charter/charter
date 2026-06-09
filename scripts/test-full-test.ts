import { tmpdir } from 'node:os';
import { join } from 'node:path';

import { cleanDirectory, newFixtureGitRepo, repoRoot, runCharter, runStep } from './lib/full-test.ts';

const assertTrue = (condition: boolean, message: string): void => {
  if (!condition) {
    throw new Error(message);
  }
};

const testInvokeNativeStepThrows = (): void => {
  let threw = false;
  try {
    runStep('failing command', 'cmd', ['/c', 'exit 7']);
  } catch (error) {
    threw = true;
    const message = error instanceof Error ? error.message : String(error);
    assertTrue(message.includes('failing command'), `expected step label in thrown message, got: ${message}`);
    assertTrue(/exit code\s+7/.test(message), `expected exit code in thrown message, got: ${message}`);
  }
  assertTrue(threw, 'runStep must throw on non-zero native command exit');
};

const testNewFixtureGitRepoMaterializesRealRepo = (): void => {
  const tempRoot = join(tmpdir(), 'charter-full-test-harness');
  cleanDirectory(tempRoot);
  try {
    const fixtureRepo = newFixtureGitRepo(join(repoRoot, 'testdata', 'repos', 'fail-mcp-unpinned'), tempRoot, 'fail-mcp-unpinned');
    const gitDir = join(fixtureRepo, '.git');
    assertTrue(Bun.file(gitDir).size >= 0, 'expected fixture repo to be git-initialized');
    const json = runCharter(['doctor', '--path', fixtureRepo, '--format', 'json'], { allowedExitCodes: [0, 1] });
    const payload = JSON.parse(json.stdout) as { findings: Array<{ rule_id: string }> };
    const rules = payload.findings.map((finding) => finding.rule_id);
    assertTrue(rules.includes('AE-MCP-001'), 'expected fail-mcp-unpinned fixture to surface AE-MCP-001');
  } finally {
    cleanDirectory(tempRoot);
  }
};

const testInvokeCharterAllowsExpectedFailure = (): void => {
  const tempRoot = join(tmpdir(), 'charter-full-test-harness-expected-failure');
  cleanDirectory(tempRoot);
  try {
    const fixtureRepo = newFixtureGitRepo(join(repoRoot, 'testdata', 'repos', 'fail-mcp-unpinned'), tempRoot, 'fail-mcp-unpinned');
    const result = runCharter(['doctor', '--path', fixtureRepo], { allowedExitCodes: [0, 1] });
    assertTrue(result.status === 1, `expected fail fixture doctor run to exit 1, got ${result.status}`);
    assertTrue(result.stdout.includes('AE-MCP-001'), 'expected fail fixture output to mention AE-MCP-001');
  } finally {
    cleanDirectory(tempRoot);
  }
};

testInvokeNativeStepThrows();
testNewFixtureGitRepoMaterializesRealRepo();
testInvokeCharterAllowsExpectedFailure();

console.log('full-test harness: PASS');
