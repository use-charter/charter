# AE-MCP-001

- Severity: High
- Category: MCP Safety
- Description: Every MCP server entry in a repo's MCP configuration must be pinned to an exact version or digest. Floating references — `@latest`, missing versions, dist-tags, semver ranges, and branch/commit-less git sources — are a supply-chain risk (OWASP MCP Top 10 beta, MCP04 Software Supply Chain Attacks & Dependency Tampering).
- Detection logic: scans tracked JSON MCP config files (`.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`; VS Code `servers` key parsed as an alias of `mcpServers`). For each stdio server launched by a recognized runner (`npx`, `bunx`, `uvx`, `pnpm dlx`, `yarn dlx`), the first non-flag argument is the package spec; the version is the substring after the last `@` that is not the leading scope prefix. Pinned = an exact semver (`1.2.3`, optional leading `v` / prerelease) or a digest (`sha256:<64hex>` or `<40hex>`). Unpinned = empty version, `latest` or any dist-tag, a semver range (`^`, `~`, `>=`, `>`, `<`, `*`, `x`), a floating git ref (`github:`, `git+`, or a `#branch`), or a dynamic value (`pkg@${VAR}`, which cannot be verified statically). Non-runner commands (`node`, `python3`, absolute paths) carry no pin assertion in v1.
- Pass example: `.mcp.json` with `"args": ["-y", "@modelcontextprotocol/server-filesystem@1.0.4"]` — exact version, passes.
- Fail example: `.mcp.json` with `"args": ["-y", "gumroad-mcp@latest"]` — floating tag, flagged High; evidence names `gumroad-mcp@latest`.
- Evidence expectations: a structured location (config file path + 1-based line of the server entry) and an evidence string naming the config file, server name, and the offending package spec. The raw token is a public package name, not a secret, so it is shown verbatim.
- Edge cases: scoped packages (`@scope/name@1.2.3`) are pinned only when the trailing version is exact; `pkg@${VERSION}` is treated as unpinned; docker/other runners beyond the recognized set are out of scope for v1.
- Remediation: pin the MCP server package to an exact version or digest instead of `@latest`, a semver range, or a floating git ref, then commit the change.
- Scoring impact: each finding is `High` (−10); no hard cap (caps are reserved for raw-secret findings).
- Related ADRs: ADR-0011, ADR-0006, ADR-0009
- Related evals: None yet
