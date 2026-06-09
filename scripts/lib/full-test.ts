import { cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { spawnSync } from 'node:child_process';
import { resolveRepoRoot } from './process.ts';

export type StepResult = {
  stdout: string;
  stderr: string;
  status: number;
};

export const repoRoot = resolveRepoRoot();

export const defaultTempRoot = join(tmpdir(), 'charter-command-tour');

export const cleanDirectory = (dir: string): void => {
  if (existsSync(dir)) {
    rmSync(dir, { recursive: true, force: true });
  }
  mkdirSync(dir, { recursive: true });
};

export const runStep = (label: string, file: string, args: string[], cwd = repoRoot): void => {
  const result = spawnSync(file, args, { cwd, stdio: 'inherit' });
  const status = result.status ?? 1;
  if (status !== 0) {
    throw new Error(`${label} failed with exit code ${status}`);
  }
};

export const runCapture = (
  label: string,
  file: string,
  args: string[],
  options: { cwd?: string; allowedExitCodes?: number[] } = {},
): StepResult => {
  const result = spawnSync(file, args, {
    cwd: options.cwd ?? repoRoot,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const status = result.status ?? 1;
  const stdout = (result.stdout ?? '').trimEnd();
  const stderr = (result.stderr ?? '').trimEnd();
  const allowed = options.allowedExitCodes ?? [0];
  if (!allowed.includes(status)) {
    const combined = [stdout, stderr].filter(Boolean).join('\n');
    throw new Error(`${label} failed with exit code ${status}${combined ? `\n${combined}` : ''}`);
  }
  return { stdout, stderr, status };
};

export const runCharter = (args: string[], options: { cwd?: string; allowedExitCodes?: number[] } = {}): StepResult =>
  runCapture(`go run ./cmd/charter ${args.join(' ')}`, 'go', ['run', './cmd/charter', ...args], options);

export const newFixtureGitRepo = (fixturePath: string, destinationRoot: string, name: string): string => {
  const target = join(destinationRoot, name);
  cpSync(fixturePath, target, { recursive: true });
  runStep(`git init ${name}`, 'git', ['-C', target, 'init']);
  runStep(`git config user.name ${name}`, 'git', ['-C', target, 'config', 'user.name', 'Charter Demo']);
  runStep(`git config user.email ${name}`, 'git', ['-C', target, 'config', 'user.email', 'charter@example.com']);
  runStep(`git add ${name}`, 'git', ['-C', target, 'add', '.']);
  runStep(`git commit ${name}`, 'git', ['-C', target, '-c', 'commit.gpgsign=false', 'commit', '-m', 'fixture']);
  return target;
};

export const newDemoRepo = (dir: string): void => {
  mkdirSync(dir, { recursive: true });
  runStep('git init demo repo', 'git', ['-C', dir, 'init']);
  runStep('git config demo user.name', 'git', ['-C', dir, 'config', 'user.name', 'Charter Demo']);
  runStep('git config demo user.email', 'git', ['-C', dir, 'config', 'user.email', 'charter@example.com']);

  writeFileSync(
    join(dir, 'go.mod'),
    ['module example.com/demo', '', 'go 1.26', ''].join('\n'),
  );
  writeFileSync(join(dir, 'main.go'), ['package main', '', 'func main() {}', ''].join('\n'));
};

export const writeSection = (title: string): void => {
  console.log(`\n== ${title} ==`);
};
