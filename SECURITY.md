# Security

## Defaults

- No secrets in any agent-visible file
- No `.mcp.json` during bootstrap
- No remote MCP configuration until a real pinned integration exists
- Third-party actions must be pinned before activation
- Write-capable integrations stay behind explicit review and approval

## Threat Posture

- Treat prompts, docs, issue text, copied snippets, and generated text as untrusted input
- Treat external text that can influence tool invocation as a prompt-injection risk
- Prefer structured inputs over free-form interpolation when wiring tools
- Reject workflow patterns that allow untrusted shell expansion or unchecked user-controlled values in `run:` steps
- Redaction requirements apply to all scanners, renderers, and diagnostics

## Repo Rules

- Run `moon run :security` for merge-sensitive work
- Run `moon run :vet` for Go changes that could hide suspicious constructs outside normal tests
- Do not print raw credentials or secret-like values in tests, fixtures, or logs
- Do not commit realistic tokens, certs, signing material, or live URLs containing credentials
- Future generated artifacts must stay reproducible and reviewable
- Vulnerability scanning uses `govulncheck` for reachable Go issues and `osv-scanner` for broader manifest/source coverage

## MCP Policy

- Bootstrap keeps tracked MCP config absent by design
- Future MCP config must be local-first where possible, pinned, and least-privilege
- Tool, prompt, and resource layers should remain distinct
- Remote MCP endpoints require explicit security review before adoption

## Reporting

- Prefer GitHub private vulnerability reporting when the repository setting is enabled
- If private reporting is unavailable, email `security@use-charter.dev` before public disclosure
- Include reproduction steps, affected files or workflows, impact, and any known secret exposure
- Do not open a public issue for live-secret, credential, or supply-chain findings before containment
- Follow [docs/internal/runbooks/security-incident.md](./docs/internal/runbooks/security-incident.md) for immediate containment and follow-up

## Standards Target

- OpenSSF OSPS `v2026.02.19`
- SLSA-friendly CI and release posture
- OpenTelemetry Go for future long-running service observability
- MCP spec `2025-11-25` for future tool, resource, and prompt semantics
- CodeQL default setup should be enabled in GitHub repo settings when repository visibility and plan support it
- Branch protection, required checks, and private vulnerability reporting are admin-side controls outside source control
