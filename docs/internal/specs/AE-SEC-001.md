# AE-SEC-001

- Severity: Blocker
- Category: Secrets
- Description: No raw secret patterns in agent-visible context files — AGENTS.md, CLAUDE.md, .cursor/rules, .windsurfrules, .github/copilot-instructions.md, opencode.md, codex.md, DESIGN.md, SKILL.md.
- Detection logic: shared detector `DetectLine` scans tracked files only (via `git ls-files --cached`); env-var references (`${VAR}`, `$VAR`) and the placeholder `your-api-key-here` are neutralized before matching; flags only high-confidence token patterns: OpenAI `sk-` + 20+ chars, GitHub `ghp_` + 30+ chars, AWS `AKIA` + 16 chars, Slack `xoxb-` + 20+ chars, and PEM headers (`BEGIN RSA/EC/PRIVATE KEY`); detected secrets are redacted in output (first 4 chars + `…`).
- Pass example: `AGENTS.md` containing `"key": "${OPENAI_API_KEY}"` or `"placeholder": "your-api-key-here"` — both pass because env-refs and placeholder are neutralized.
- Fail example: `AGENTS.md` containing `"OPENAI_API_KEY": "sk-proj-abc123…T3BlbkFJXxyzABC"` (a real `sk-` token of length ≥ 20) — detected and redacted as `sk-p…`.
- Evidence expectations: a structured location (file path + 1-based line) for the matched secret, plus a redacted excerpt (never the raw value); scans only tracked inventory files; untracked/gitignored local files are intentionally not scanned.
- Edge cases: file must be in repo inventory AND git-tracked to be scanned; uncommitted local scratch files are safe by design; `.gitignore`d files are never scanned.
- Remediation: remove the literal secret, rotate it externally, reference an environment variable instead, and commit the fix.
- Scoring impact: when any AE-SEC-001 or AE-SEC-002 finding is present, final Charter score is hard-capped at **49**.
- Why: Agent context files are read on every task. A raw credential in AGENTS.md is visible to every model, every session, and every log that captures context windows. Rotation after exposure is the only safe recovery.
- Auto-fixable: No
- Related rules: AE-SEC-002, AE-CTX-001
- Related ADRs: ADR-0006, ADR-0007, ADR-0008
- Related evals: None yet
