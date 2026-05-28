import { spawnSync } from 'node:child_process';

export const runTextCommand = (
  command: string,
  args: string[],
  options: Parameters<typeof spawnSync>[2] = {},
) =>
  spawnSync(command, args, {
    stdio: ['ignore', 'pipe', 'inherit'],
    encoding: 'utf8',
    ...options,
  });

export const exitWithStatus = (status: number | null): never => {
  process.exit(status ?? 1);
};

export const readStdoutText = (stdout: string | NodeJS.ArrayBufferView): string => {
  if (typeof stdout === 'string') {
    return stdout;
  }

  return Buffer.from(stdout.buffer, stdout.byteOffset, stdout.byteLength).toString('utf8');
};

export const resolveRepoRoot = (): string => {
  const result = runTextCommand('git', ['rev-parse', '--show-toplevel']);
  if (result.status !== 0) {
    exitWithStatus(result.status);
  }

  return readStdoutText(result.stdout).trim();
};
