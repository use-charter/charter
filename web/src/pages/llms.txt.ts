import type { APIRoute } from 'astro';

// llms.txt (llmstxt.org): a concise, link-rich brief that AI answer engines and
// crawlers read to understand the product and find authoritative sources. Reuses
// `site` from astro.config so every link stays absolute and in sync. Content is
// ASCII-only so it renders identically in every client and encoding.
const getLlmsTxt = (base: URL) => {
  const url = (path: string) => new URL(path, base).href;
  return `# Charter

> Charter is an offline-first CLI that scores any repository for AI-agent readiness: a deterministic 0-100 score across 18 rules and 9 categories, in under two seconds. No network, no LLM calls, no telemetry; nothing leaves the machine.

Charter makes repositories safe and legible for AI coding agents (Claude Code, Cursor, Copilot, Windsurf, Codex, Gemini, Grok). It checks for missing context files, secrets exposed to agents, unpinned MCP servers, agent configuration, and discoverable verify commands. Every finding carries a rule ID and a fix. Apache-2.0 licensed, SLSA Level 3 signed, with SARIF 2.1.0 output for GitHub Code Scanning.

The nine rule categories are Context, Secrets, MCP Safety, Agent Config, Environment, CI, Testing, Autonomy, and Governance. A repository passes when its score meets the configured threshold (default 80); CI can gate every pull request on that threshold.

## Commands
- [charter doctor](${url('/docs/cli/doctor')}): Scan the repo and print a 0-100 readiness score with a per-category breakdown.
- [charter init](${url('/docs/cli/init')}): Scaffold the context files an agent needs (AGENTS.md, charter.yaml).
- [charter fix](${url('/docs/cli/fix')}): Apply diff-first repairs for the safe fixers; secrets are never auto-touched.
- [charter explain](${url('/docs/cli/explain')}): Explain any rule or finding by its rule ID.
- [charter report](${url('/docs/cli/report')}): Produce a shareable report in SARIF, JSON, or HTML.
- [charter suppress](${url('/docs/cli/suppress')}): Log an accepted risk with a reason, owner, and expiry.
- [charter version](${url('/docs/cli/version')}): Print the installed version and build provenance.

## Documentation
- [Quickstart](${url('/docs/quickstart')}): Install Charter and run the first scan.
- [CLI overview](${url('/docs/cli/overview')}): Every command, flag, and exit code.
- [Rules overview](${url('/docs/rules/overview')}): The 18 rules across 9 categories.
- [GitHub Action](${url('/docs/ci/github-action')}): Gate every pull request on a readiness threshold.

## Source
- [GitHub repository](https://github.com/use-charter/charter): Source, issues, and releases.
- [License (Apache-2.0)](https://github.com/use-charter/charter/blob/main/LICENSE): Permissive open-source license.

## Optional
- [Changelog](${url('/docs/changelog')}): Release notes and version history.
`;
};

export const GET: APIRoute = ({ site }) => {
  const base = site ?? new URL('https://use-charter.dev');
  return new Response(getLlmsTxt(base), {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
};
