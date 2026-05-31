# AE-CC-001

- Severity: Blocker
- Category: Agent Config
- Description: Agent hook configurations must not contain dangerous shell commands that a prompt-injected agent could weaponize (OWASP MCP Top 10 beta, MCP05 Command Injection & Execution).
- Detection logic: scans tracked JSON hook config files (`.claude/settings.json`, `.claude/settings.local.json`, `.cursor/hooks.json`). Each file's parsed `hooks` structure is walked to collect every `command` string and `args` entry across all events and handler shapes (Claude Code's nested `{type:"command", command, args}` and Cursor's flatter `{command}`). Each collected command is matched against a high-confidence dangerous-command set: destructive operations (`rm -rf`, `git reset --hard`, `git clean -fd`, `dd `, `mkfs`, `truncate`) and privilege escalation (`sudo `, `chmod 777`, `chown -R `). Operator-chaining and command-substitution injection (`&&`, `||`, `;`, `$(…)`, backticks) is intentionally not flagged in v1 — it is false-positive-prone and deferred to a later, context-aware refinement (matching the high-confidence posture of `AE-SEC-001`).
- Pass example: `.claude/settings.json` with a `PreToolUse` hook running `"\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/format.sh"` — a scoped script path, passes.
- Fail example: `.claude/settings.json` with a hook `command` of `"rm -rf $CLAUDE_PROJECT_DIR/build"` or `"sudo chmod 777 ./bin"` — flagged Blocker; evidence quotes the offending command.
- Evidence expectations: a structured location (config file path + 1-based line of the matched command) and an evidence string naming the config file and the offending command (e.g., `.claude/settings.json:7: hook command uses rm -rf`).
- Edge cases: only the three JSON hook files are scanned in v1; Pkl (`hk.pkl`), YAML (`.pre-commit-config.yaml`, `lefthook.yml`), and shell-dir (`.husky/`) hook managers are out of scope for v1; a controlled `&&` chain (`cd app && npm test`) is not flagged (operator chaining is deferred); a malformed JSON hook file fails the scan fast with a wrapped error.
- Remediation: replace the dangerous command with an explicit, scoped, non-destructive command; prefer array-form (`args`) execution to avoid shell expansion; review each hook against "if an agent were prompt-injected, could this hook be weaponized?", then commit the change.
- Scoring impact: `Blocker` — engages the hard blocker score cap (final score ≤ 59 whenever any blocker finding is present); no separate per-finding cap.
- Related ADRs: ADR-0012, ADR-0006, ADR-0009
- Related evals: None yet
