# Permissions

## Default Edit Zones

- Allowed by default: repo docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon and mise config
- Off-limits by default: `.env*`, `secrets/`, signing keys, credentials, future production infra without explicit task need

## Mandatory Checks

Latest docs lookup is mandatory before code that touches:

- frameworks or SDKs
- CI actions
- MCP
- API contracts or schemas
- build or tool configs

## Network and Escalation

- Network access is for current official docs, package/tool verification, or explicitly required research
- Elevated permissions require explicit approval when writes escape repo bounds, network access is unavailable but required, or actions are destructive

Escalate immediately when:

- a command would delete tracked files outside the requested scope
- a command would alter git history or revert user work
- a task requires external services, credentials, or write-capable remote integrations
- a required latest-docs or package-verification step cannot complete inside normal constraints

## Destructive Actions

- No `git reset --hard`
- No `rm -rf`
- No checkout-based reverts of user changes
- No forceful cleanup without explicit request

Destructive means any action that deletes user work, rewrites history, reverts tracked content, or mutates files outside the requested change set.

## Generated Code

- Must be clearly labeled in PR notes
- Must have rationale and verification
- Must not become an unowned boundary
- Must stay inside an explicitly owned package or directory boundary

## MCP Policy

- No `.mcp.json` during bootstrap
- Future MCP tools must be pinned, local-first where possible, and least-privilege
- Tool, prompt, and resource layers should remain distinct
- Write-capable MCP operations require explicit approval semantics

## Secrets

- Never place live secrets in docs, prompts, tests, fixtures, or workflows
- Use obviously fake placeholders only
