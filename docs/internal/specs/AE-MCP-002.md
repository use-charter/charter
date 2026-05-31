# AE-MCP-002

- Severity: High
- Category: MCP Safety
- Description: Every remote MCP server origin must be known — present in the repo's trusted-remote allowlist. Unknown remote origins are flagged (OWASP MCP Top 10 beta, MCP09 Shadow MCP Servers).
- Detection logic: scans tracked JSON MCP config files (`.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`). A server is remote when it has a `url` or a `type` of `http`/`sse`. The URL host is compared (case-insensitively) against `charter.yaml → mcp.trustedRemotes` (a list of hostnames). A non-local remote whose host is absent from the allowlist is flagged. When no `charter.yaml` allowlist is present, every non-local remote is flagged as unverifiable with a distinct summary so the remediation (add an allowlist) is clear. Localhost origins (`localhost`, `127.0.0.1`, `::1`, `0.0.0.0`, `*.localhost`) are exempt. A bare env-reference URL (`${API_URL}`) yields no parseable host and is skipped.
- Pass example: `.mcp.json` server `"url": "https://mcp.asana.com/mcp"` with `charter.yaml` listing `mcp.trustedRemotes: [mcp.asana.com]` — host allowlisted, passes.
- Fail example: `.mcp.json` server `"url": "https://unknown.example.net/mcp"` with no matching allowlist entry — flagged High; evidence names the server and host.
- Evidence expectations: a structured location (config file path + 1-based line of the server entry) and an evidence string naming the config file, server name, and resolved host. The summary distinguishes "not in allowlist" from "no allowlist configured".
- Edge cases: localhost and loopback remotes never fire; SSE (`type: sse`, deprecated in favor of HTTP) is treated as remote; a dynamic `${VAR}` URL is skipped; allowlist matching is host-only (no scheme or path).
- Remediation: add the reviewed host to `charter.yaml → mcp.trustedRemotes`, or replace the server with a trusted origin, then commit the change. (The M1.6 MCP catalog will later supersede the local allowlist as the primary trust source.)
- Scoring impact: each finding is `High` (−10); no hard cap.
- Related ADRs: ADR-0011, ADR-0006, ADR-0009
- Related evals: None yet
