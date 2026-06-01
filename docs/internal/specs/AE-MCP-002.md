# AE-MCP-002

- Severity: High
- Category: MCP Safety
- Description: Every remote MCP server origin must be known — present in the repo's trusted-remote allowlist. Unknown remote origins are flagged (OWASP MCP Top 10 beta, MCP09 Shadow MCP Servers).
- Detection logic: scans tracked JSON MCP config files (`.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`). A server is remote when it has a `url` or a `type` of `http`/`sse`. The URL host is compared (case-insensitively) against `charter.yaml → mcp.trustedRemotes` (a list of hostnames). A non-local remote whose host is absent from the allowlist is flagged. When no `charter.yaml` allowlist is present, every non-local remote is flagged as unverifiable with a distinct summary so the remediation (add an allowlist) is clear. Localhost origins (`localhost`, `127.0.0.1`, `::1`, `0.0.0.0`, `*.localhost`) are exempt. A bare env-reference URL (`${API_URL}`) yields no parseable host and is skipped.
- Pass example: `.mcp.json` server `"url": "https://mcp.asana.com/mcp"` with `charter.yaml` listing `mcp.trustedRemotes: [mcp.asana.com]` — host allowlisted, passes.
- Fail example: `.mcp.json` server `"url": "https://unknown.example.net/mcp"` with no matching allowlist entry — flagged High; evidence names the server and host.
- Evidence expectations: a structured location (config file path + 1-based line of the server entry) and an evidence string naming the config file, server name, and resolved host. The summary distinguishes "not in allowlist" from "no allowlist configured".
- Edge cases: localhost and the `127.0.0.0/8` loopback range (plus `::1`, `0.0.0.0`, `*.localhost`) never fire; SSE (`type: sse`, deprecated in favor of HTTP) is treated as remote; a scheme-less or dynamic `${VAR}` URL has no parseable host and is skipped; allowlist matching is host-only (no scheme or path).
- Remediation: add the reviewed host to `charter.yaml → mcp.trustedRemotes`, or replace the server with a trusted origin, then commit the change.
- Catalog (Slice 13, ADR-0021): the effective allowlist is `union(charter.yaml → mcp.trustedRemotes, catalog trustedHosts)`. The catalog ships a baseline of vendor-operated remote MCP hosts (GitHub, Sentry, Linear, Atlassian, Notion, Stripe, Vercel, Asana, and the full Cloudflare managed `*.mcp.cloudflare.com` set), so those origins pass without per-repo config (a pure false-positive reduction). Unknown remotes still flag; `charter.yaml` remains the per-repo override. Matching stays host-only and case-insensitive.
- Scoring impact: each finding is `High` (−10); no hard cap.
- Related ADRs: ADR-0021, ADR-0011, ADR-0006, ADR-0009
- Related evals: None yet
