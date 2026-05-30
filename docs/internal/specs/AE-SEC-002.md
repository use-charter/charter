# AE-SEC-002

- Severity: Blocker
- Category: Secrets
- Description: No raw secret patterns in MCP or adjacent config files — .mcp.json, .mcp.yml, .cursor/mcp.json, .claude/settings.json, claude_desktop_config.json, cline_mcp_settings.json, and any `*.pkl` file with path containing `mcp` or `config`.
- Detection logic: shared detector `DetectLine` scans config files for high-confidence token patterns; env-var references (`"${OPENAI_API_KEY}"`, `"$OPENAI_API_KEY"`) are neutralized and **pass**; literal secret values matching OpenAI/GitHub/AWS/Slack prefixes or PEM headers are flagged; redacted in output (first 4 chars + `…`).
- Pass example: `.mcp.json` with `"env": {"OPENAI_API_KEY": "${OPENAI_API_KEY}"}` — env-ref passes; `"api_key": "your-api-key-here"` — placeholder passes.
- Fail example: `.mcp.json` containing `"api_key": "sk-proj-abc123…T3BlbkFJXxyzABC"` — detected and redacted as `sk-p…`.
- Evidence expectations: config file path and redacted secret value (never the raw value); env-refs and placeholders explicitly pass.
- Edge cases: environment variable references (both `${VAR}` and `$VAR` syntax) are intentionally safe and never flagged; only literal string values match detection.
- Remediation: replace literal secrets with environment variable references (`"key": "${ENV_VAR_NAME}"`); move secrets to a runtime secrets manager.
- Scoring impact: when any AE-SEC-001 or AE-SEC-002 finding is present, final Charter score is hard-capped at **49** (per ADR-0008, independent of other caps).
- Related ADRs: ADR-0006, ADR-0007, ADR-0008
- Related evals: None yet
