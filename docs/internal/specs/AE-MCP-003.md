# AE-MCP-003

- Severity: High
- Category: MCP Safety
- Description: Remote MCP servers must declare authentication metadata. A non-local remote server with no auth declaration is flagged (OWASP MCP Top 10 beta, MCP07 Insufficient Authentication & Authorization), aligned with MCP specification revision `2025-11-25`.
- Detection logic: scans tracked JSON MCP config files (`.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`). For each non-local remote server (`url` or `type` of `http`/`sse`), detection is presence-based: the server passes if it declares any non-empty auth header — `Authorization`, `X-Api-Key`, `Api-Key`, or `X-Auth-Token` (case-insensitive; an env-reference value such as `Bearer ${TOKEN}` counts as declared). A remote with no such header is flagged. Auth detection is intentionally presence-based rather than asserting specific OAuth 2.1/PKCE field names, so it stays resilient to the MCP `2026-07-28` release candidate (which hardens OAuth without changing static config files). Localhost origins are exempt.
- Pass example: `.mcp.json` server `"url": "https://mcp.asana.com/mcp"` with `"headers": { "Authorization": "Bearer ${ASANA_TOKEN}" }` — auth declared, passes.
- Fail example: `.mcp.json` server `"url": "https://mcp.asana.com/mcp"` with no `headers` (or no auth header) — flagged High; evidence names the server and host.
- Evidence expectations: a structured location (config file path + 1-based line of the server entry) and an evidence string naming the config file, server name, and host. Header values are env-references or opaque and are not secrets; no redaction is required.
- Edge cases: localhost/loopback remotes never fire; an env-reference header value satisfies the presence check (Charter does not validate the credential, only its declaration); a bare `${VAR}` URL with no parseable host is skipped.
- Remediation: declare an auth header for the remote MCP server (for example `Authorization` referencing an environment variable), or switch to a local/trusted integration mode, then commit the change.
- Scoring impact: each finding is `High` (−10); no hard cap.
- Related ADRs: ADR-0011, ADR-0006, ADR-0009
- Related evals: None yet
