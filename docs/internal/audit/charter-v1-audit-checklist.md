# Charter v1 — Rule Audit Checklist

15 rules: 12 core + 3 governance. Pass all three signals (toolchain + lockfile + hook config) for AE-ENV-001.

`charter-architecture-2026` is the canonical source for Charter product behavior, command surface, transports, output formats, and roadmap. This checklist is the v1 audit companion and must stay aligned with the architecture source. It does not redefine canonical behavior, and HTML mirrors remain presentation-only.

Documentation authority ladder:

1. `docs/internal/architecture/charter-architecture-2026.md` for product behavior
2. `docs/internal/audit/charter-v1-audit-checklist.md` for manual rule-audit companion detail
3. ADRs in `docs/internal/decisions/` for irreversible constraints
4. root workflow and companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only

## Scoring Reference

```
score = max(0, 100 − B×20 − H×10 − M×4 − L×1)
final = min(base, applicable_cap)
```

| Severity | Penalty |
|---|---|
| Blocker | −20 per finding |
| High | −10 per finding |
| Medium | −4 per finding |
| Low | −1 per finding |

**Hard caps** (lowest applicable wins):

| Condition | Cap |
|---|---|
| Raw secret in any agent-visible file (AE-SEC-001/002) | score ≤ 49 |
| Any Blocker-severity finding | score ≤ 59 |
| Unscanned or unknown repo state | score ≤ 79 |
| Suppressed findings | excluded from score, listed separately |

> ⚠ AE-SUPPRESS-003 is informational — failing it does **not** deduct points or affect the score.

## Reference Metadata

- **Scan engine:** Gitleaks v8.30.1 (secrets) + Charter rule engine (agent config)
- **Agent formats covered:** AGENTS.md · CLAUDE.md · .cursor/rules · .windsurfrules · .github/copilot-instructions.md · opencode.md · codex.md · DESIGN.md · SKILL.md
- **MCP config locations:** .mcp.json · .mcp.yml · .cursor/mcp.json · .claude/settings.json · claude_desktop_config.json · cline_mcp_settings.json · *.pkl
- **Toolchain files (AE-ENV-001):** mise.toml (recommended, polyglot) · go.mod · .nvmrc · bunfig.toml · pyproject.toml · rust-toolchain.toml · .swift-version · gradle-wrapper.properties · .ruby-version · .tool-versions (asdf) · devcontainer.json · flake.nix
- **Hook managers (AE-ENV-001):** hk (hk.pkl, preferred) · husky (.husky/) · lefthook (lefthook.yml) · pre-commit (.pre-commit-config.yaml) · simple-git-hooks · lint-staged · overcommit (.overcommit.yml) · cargo-husky
- **Suppression syntax:** `# charter:ignore AE-RULE-NNN reason="…"`
- **Suppression TTL:** 90 days default (configurable). Permanent suppression is reserved for Phase 2 Cloud and requires an `ApprovedBy` field; without it, Charter treats the suppression as 90-day.

---

## AE-CTX-001 — AGENTS.md Missing or Empty
**Severity:** 🔴 BLOCKER  

**Check:** Does the repo root contain an agent context file? Charter recognises all nine formats — check for any of: AGENTS.md , CLAUDE.md , .cursor/rules , .windsurfrules , .github/copilot-instructions.md , opencode.md , codex.md , DESIGN.md (Google Labs standard), SKILL.md (agentskills.io). Is the file non-empty and does it contain meaningful content — at minimum: project summary, tech stack, directories safe for agent edits, off-limits paths, and a verification command (e.g., charter doctor )? Also check: is the file within a reasonable token budget (≤ 600 tokens recommended; configurable via charter.yaml → rules.AE-CTX-001.token_budget ; files over budget risk being partially ignored by agents)?  

**Evidence:** File path and format detected. Note if empty or stub-only. Quote the first 2–3 substantive lines. Note the approximate token count if it appears large.  

**False Positive Risk:** FP Risk: Very Low. A file must exist and contain meaningful content. An empty placeholder, a single-line README copy, or a file with only TODO comments still fails. A repo that uses DESIGN.md or SKILL.md instead of AGENTS.md passes if the file meets the content requirements.  

**Fix:** Run charter init to scaffold AGENTS.md from a template matched to your detected tech stack. Minimum content: project summary, tech stack, directories safe for agent edits, off-limits paths, and the command agents should run to verify their changes ( charter doctor ). If you also use charter serve , add a note in AGENTS.md so agents know Charter is available as an MCP tool.  

---

## AE-CTX-002 — Agent Context Stale
**Severity:** 🟡 MEDIUM  

**Check:** If an agent context file exists, is its content consistent with the actual repo state? Compare: (1) stated tech stack vs. language toolchain files — mise.toml (polyglot tool manager, recommended), go.mod (Go), package.json / bunfig.toml (JS/TS), pyproject.toml / .python-version (Python), rust-toolchain.toml / Cargo.toml (Rust), .swift-version (Swift), gradle-wrapper.properties (Kotlin/JVM), .ruby-version (Ruby), .tool-versions (asdf); (2) stated off-limits paths vs. current directory structure; (3) stated test/verify command vs. current CI config; (4) stated hook tooling vs. the repo's committed hook config — check all supported managers: hk.pkl (hk, preferred for Go/JS), .husky/ (husky), lefthook.yml / lefthook.toml (lefthook), .pre-commit-config.yaml (pre-commit), simple-git-hooks key in package.json, .lintstagedrc (lint-staged), .overcommit.yml (overcommit), .cargo-husky/hooks/ (cargo-husky) — if a hook manager is active, AGENTS.md should reference it so agents know pre-commit checks will run on their commits; (5) stated MCP tools vs. .mcp.json — if charter serve is configured, AGENTS.md should mention it.  

**Evidence:** Last-modified date of the agent context file. List any specific factual mismatches found (e.g., 'AGENTS.md says Node 18, mise.toml pins Bun 1.3.14' or 'AGENTS.md does not mention charter serve but .mcp.json configures it').  

**False Positive Risk:** FP Risk: Medium. A stale date alone is not enough — focus on factual contradictions between the doc and the repo. Stable repos with slow-moving stacks may have old-dated but accurate docs.  

**Fix:** Update the agent context file to match current stack. Add a Last reviewed: YYYY-MM-DD line at the top. Add a CI lint step that warns when the context file hasn't changed in 90+ days while the repo has active commits ( charter doctor emits AE-CTX-002 at Medium severity when staleness is detected).  

---

## AE-SEC-001 — Raw Secret in Tracked File
**Severity:** 🔴 BLOCKER  

**Check:** Scan all tracked files for raw credential patterns: API keys (sk-*, ghp_*, AKIA*, xoxb-*), private keys (BEGIN RSA/EC/PRIVATE KEY), bearer tokens, connection strings with passwords, and high-entropy strings in config files. Focus on: .env committed to git, AGENTS.md, shell scripts, CI yamls, JSON config. Scan engine: Charter uses Gitleaks v8.30.1 under the hood for history and working-tree scanning, then applies its own agent-config-specific patterns on top (AGENTS.md, MCP config files, hk.pkl). Gitleaks covers git history; Charter adds the agent-visible file surface. A finding from Gitleaks is always High-confidence; a Charter-layer finding uses Confidence tiers.  

**Evidence:** File path + line number. Never record the actual secret value. Record only the pattern matched (e.g., 'OPENAI_API_KEY=sk-… found at .env:3'). Note whether the file is in .gitignore.  

**False Positive Risk:** FP Risk: Medium — use Confidence tiers. High-confidence (mark Fail): exact recognized prefix + correct format length (sk-…T3BlbkFJ, ghp_…, AKIA[A-Z0-9]{16}, xoxb-…). Medium-confidence (note but verify): high entropy string matching a known prefix but wrong length, or entropy-only match without a recognizable prefix. Output as 'likely credential — verify manually.' Low-confidence (note for review, do not block): suspicious env var assignment without a recognizable pattern. Common FPs: test fixtures with clearly fake keys, documentation code blocks, commented-out placeholders. Only mark Fail for High-confidence matches. Entropy alone without a recognizable prefix is at most Medium-confidence.  

**Fix:** Revoke the exposed credential immediately. Remove from git history (git-filter-repo or BFG). Add the file to .gitignore. Switch to environment variable injection or a secrets manager. Covers OWASP MCP01 — Token Mismanagement & Secret Exposure.  

---

## AE-SEC-002 — Raw Secret in MCP/Agent Config
**Severity:** 🔴 BLOCKER  

**Check:** Check all MCP config files for secrets embedded directly in: env object values, command args arrays, headers objects, or any string value that matches a credential pattern or has high entropy (≥ 4.0 bits/char over 20+ chars). MCP config locations to scan: .mcp.json , .mcp.yml , .cursor/mcp.json , .claude/settings.json , claude_desktop_config.json , cline_mcp_settings.json , *.pkl (MCP Pkl configs). Also check hk.pkl if it passes environment variables or tokens to hook commands.  

**Evidence:** Config file path + key path (e.g., 'servers.filesystem.env.API_KEY at .mcp.json'). Evidence.Snippet must be [REDACTED] — never the raw value.  

**False Positive Risk:** FP Risk: Medium — use Confidence tiers. Environment variable references ( \"${MY_API_KEY}\" , \"$MY_KEY\" ) are safe injection patterns — never flag these. High-confidence (Fail): literal value matching a recognized credential prefix at correct length. Medium-confidence (verify): high-entropy literal string (≥4.0 bits/char, ≥20 chars) in a semantically sensitive key (env, args, headers). Low-confidence (note only): high-entropy string in tool name or server description. Do not block on Low-confidence.  

**Fix:** Remove the literal secret from the MCP config. Instead reference the value via the host environment: \"env\": { \"API_KEY\": \"${MY_API_KEY}\" } . Use a secrets manager to inject values at runtime. If suppressing: # charter:ignore AE-SEC-002 reason=\"test credential — zero real-world access, rotated monthly\" . Suppressions expire in 90 days by default. Covers OWASP MCP01 — Token Mismanagement & Secret Exposure.  

---

## AE-MCP-001 — MCP Server Unpinned or Vulnerable
**Severity:** 🟠 HIGH  

**Check:** For every MCP server configured in any MCP config file, check two things: (1) Version pinning — is the package version pinned to a specific release? Scan all MCP config locations: .mcp.json , .mcp.yml , .cursor/mcp.json , .claude/settings.json , claude_desktop_config.json , cline_mcp_settings.json , *.pkl . Flag: version absent entirely, version set to latest , version set to a semver range (^1.0, ~2.1, >=3), or git source using a branch name rather than a full 40-char commit SHA — these fire as HIGH . (2) CVE advisory check — for servers that ARE pinned, cross-reference the pinned version against Charter's embedded MCP catalog ( catalog/v1/*.yaml ) advisory entries. If the catalog has an advisories[] entry whose affected version range includes the pinned version, this finding escalates to BLOCKER . A properly pinned server running a CVE-affected version is more dangerous than an unpinned server, because the team believes it is safe.  

**Evidence:** Config file path, server name, and the version string — whether unpinned (e.g., \"latest\" ) or pinned but advisory-matched (e.g., \"1.0.2\" with catalog advisory CVE-2025-XXXXX affecting ≤1.0.3). Note severity: HIGH for unpinned/outdated, BLOCKER for advisory match. Cross-reference the catalog's stable_version field to confirm the recommended upgrade target.  

**False Positive Risk:** FP Risk: Low. A server pinned to an exact semver (e.g., @1.2.3) or a full 40-char git SHA is safe for the unpinned check. Semver ranges (^, ~, >=) are genuinely risky for supply chain reasons. For the CVE escalation: a finding only fires if the catalog has an explicit advisory entry for that server and version range — not speculative. Mark N/A if the repo has no MCP config files at all — this rule only applies to repos with active MCP server configuration. Note: charter serve is a local STDIO server with no package version — it is always safe and never flagged by this rule.  

**Fix:** For unpinned servers: pin to exact version: \"@modelcontextprotocol/server-filesystem\": \"1.0.4\" . For git sources, pin to a full commit SHA. Run npm install --save-exact or bun add --exact and commit the lockfile. For CVE-affected versions (BLOCKER): upgrade immediately to the catalog's stable_version . Run charter fix AE-MCP-001 for an auto-generated upgrade diff. Check charter catalog show <id> for advisory details and recommended version. Covers OWASP MCP04 — Supply Chain Attacks & Dependency Tampering.  

---

## AE-MCP-002 — MCP Server Unknown or Untrusted URL
**Severity:** 🟠 HIGH  

**Check:** For remote MCP servers (HTTP/HTTPS/SSE type), check whether the server URL appears in Charter's embedded static MCP catalog ( catalog/v1/*.yaml , embedded in the binary) or in a team allowlist ( charter.yaml → mcp.trustedRemotes ). Flag any remote URL that is not in either source. Also flag servers sourced from git URLs at unknown organizations or accounts. Note: STDIO-type servers (like charter serve ) are local by definition and are never flagged by this rule.  

**Evidence:** Server name and remote URL. Note whether the URL matches a catalog entry or a local allowlist. If the server is from a known vendor not yet in Charter's catalog, note it for catalog contribution.  

**False Positive Risk:** FP Risk: Medium. A remote URL from a well-known vendor (Anthropic, GitHub, Stripe, Cloudflare) that isn't in Charter's embedded catalog yet is a FP risk — file a catalog contribution. Mark as FP with a note if the server is from a verifiably trusted organization. charter serve (STDIO transport, local) is always safe — never flag it. Mark N/A if the repo has no MCP config files at all.  

**Fix:** Add trusted remote URLs to charter.yaml → mcp.trustedRemotes . Or switch to a local/npm-installed equivalent of the server. For unknown origins, audit the server's tool list before adding it. Covers OWASP MCP09 — Shadow MCP Servers.  

---

## AE-MCP-003 — Remote MCP Server Lacks Auth Metadata
**Severity:** 🟠 HIGH  

**Check:** For remote MCP servers that require authentication (servers with type: 'http' or 'sse'), check whether the config includes an authorization block with OAuth 2.1 / PKCE metadata. A server that accepts user credentials without a documented auth flow is a trust gap. Per MCP spec 2025-11-25, remote servers should declare their auth requirements.  

**Evidence:** Server name, type, and URL. Note whether an auth/authorization block exists in the config. Note if the server is listed as 'public/unauthenticated' in its own docs.  

**False Positive Risk:** FP Risk: High. Many legitimate read-only remote servers don't require auth. If the server is genuinely public and read-only (e.g., a public docs search server), mark N/A with a note. Only flag if the server clearly handles sensitive data or user-specific state without an auth declaration. Mark N/A if the repo has no MCP config files at all — this rule only applies to repos with active MCP server configuration.  

**Fix:** Add an authorization block to the server config per MCP spec: specify authorizationUrl, tokenUrl, and scopes. Or switch to an npm-installed local server that doesn't require remote auth. Covers OWASP MCP07 — Insufficient Authentication & Authorization.  

---

## AE-CC-001 — Dangerous Command in Agent Config
**Severity:** 🔴 BLOCKER  

**Check:** Check all agent hook configurations for dangerous shell patterns: shell injection sequences ( ; , && , || , $(...) , backtick), destructive commands ( rm -rf , git reset --hard , truncate , dd ), or privilege escalation ( sudo , chmod 777 , chown -R ). Hook config locations to check: .claude/settings.json (hooks block), .cursor/rules (tool-use config), Codex hook definitions, and hk.pkl — hk is the modern Git hook manager (used by Charter itself); check its pre-commit , pre-push , and commit-msg hook definitions for any open-ended shell patterns that could be weaponized by prompt injection.  

**Evidence:** Config file path and the specific hook type + dangerous command found. Quote the minimal relevant snippet.  

**False Positive Risk:** FP Risk: Low. Hook scripts that use && for command chaining in a controlled, single-purpose command (e.g., cd app && npm test ) are generally safe. Flag patterns that are open-ended or could be exploited by prompt injection — not every use of shell operators.  

**Fix:** Replace open-ended shell patterns with explicit, scoped commands. Use array-form command execution where possible to avoid shell expansion. Review every hook against the principle: 'if an agent were prompt-injected, could this hook be weaponized?' Covers OWASP MCP05 — Command Injection.  

---

## AE-CC-002 — Overly Broad Agent Edit Scope
**Severity:** 🟠 HIGH  

**Check:** Does any agent context file explicitly restrict what directories, file types, or operations the agent may perform? Check all formats: AGENTS.md , CLAUDE.md , .cursor/rules , .windsurfrules , .github/copilot-instructions.md , opencode.md , codex.md . Flag: no off-limits paths defined, edit scope covers the entire repo root without exclusions, infrastructure or secrets paths are not explicitly excluded.  

**Evidence:** Note what scope definition exists (or doesn't). Quote the relevant permission/scope block. Flag specific dangerous paths that are not excluded (e.g., .github/workflows/, terraform/, db/migrations/, secrets/).  

**False Positive Risk:** FP Risk: Medium. A small single-purpose repo with no sensitive paths may legitimately have broad scope. Focus on repos where broad scope creates real risk: infra code, migration scripts, CI pipeline definitions, credential files.  

**Fix:** Add an explicit 'Off-limits for agents' section to AGENTS.md listing at minimum: .github/workflows/ , terraform/ or infra/ , db/migrations/ , .env* , secrets/ . Consider using allowed_paths in CLAUDE.md to hard-restrict what charter doctor reports on. Covers OWASP MCP02 — Permissioning Failures.  

---

## AE-ENV-001 — Reproducibility Missing
**Severity:** 🟡 MEDIUM  

**Check:** Does the repo have a machine-readable specification of the development toolchain that agents (and CI) can use to reproduce a known-good environment? Step 1 — Toolchain declaration. Check for any of the following (one is sufficient): Polyglot tool managers (recommended — one file covers all runtimes): mise.toml / .mise.toml · .tool-versions (asdf-compatible) · devcontainer.json / .devcontainer/** · flake.nix + flake.lock Go: go.mod with a go directive and/or toolchain line (Go 1.21+) · .go-version JavaScript / TypeScript: .nvmrc · .node-version · package.json engines.node or volta field · bunfig.toml (Bun projects) Python: pyproject.toml with requires-python · .python-version · uv.toml / .python-version (uv/pyenv) Rust: rust-toolchain.toml or rust-toolchain (channel + targets + components) Swift: .swift-version · Package.swift swift-tools-version comment Kotlin / JVM: gradle/wrapper/gradle-wrapper.properties ( distributionUrl pins Gradle version) · .java-version · jvmToolchain in build.gradle.kts Ruby: .ruby-version · Gemfile with ruby declaration Step 2 — Lockfile committed. Without a lockfile, dependency resolution is non-deterministic even with a pinned runtime. Check per language: Go: go.sum JS/TS (npm): package-lock.json · (yarn): yarn.lock · (pnpm): pnpm-lock.yaml · (Bun): bun.lock / bun.lockb Python (pip): requirements.txt pinned · (poetry): poetry.lock · (uv): uv.lock Rust: Cargo.lock (always commit for binaries; optional for libraries) Swift (SPM): Package.resolved Kotlin/Gradle: Gradle dependency verification file ( gradle/verification-metadata.xml ) or pinned libs.versions.toml Ruby: Gemfile.lock Step 3 — Committed hook configuration. Consistent pre-commit behavior is part of reproducibility — if an agent's commits bypass hooks, quality gates silently disappear. Any one of the following satisfies this check: hk — hk.pkl committed (preferred for Go/JS stacks using Pkl config) husky — .husky/ directory committed with ≥1 hook file (JS/TS ecosystem standard) lefthook — lefthook.yml or lefthook.toml committed (polyglot-friendly, fast) pre-commit — .pre-commit-config.yaml committed (Python-heavy or polyglot projects) simple-git-hooks — package.json simple-git-hooks key present (lightweight JS alternative to husky) lint-staged — .lintstagedrc or package.json lint-staged key (commonly paired with husky / simple-git-hooks) overcommit — .overcommit.yml committed (Ruby projects) cargo-husky — .cargo-husky/hooks/ committed (Rust projects) Mark this check as pass if any one toolchain declaration + any one lockfile (where applicable) + any one hook config is committed.  

**Evidence:** List which reproducibility signals are present or absent. Note: (a) detected language(s) and which toolchain file covers each (e.g., 'Go → go.mod toolchain go1.26.3; Node → .nvmrc 22.x; Python → pyproject.toml requires-python >=3.12'); (b) whether each lockfile is committed and matches CI; (c) which hook manager is in use and whether its config file is committed. Note any language whose runtime version is completely undeclared.  

**False Positive Risk:** FP Risk: Low. A repo with no build system or no external dependencies may genuinely need no lockfile. Only flag if there are clear dependencies (e.g., go.mod without go.sum , package.json without lockfile, Cargo.toml binary without Cargo.lock ). Hook config absence is only a finding if the repo has evidence of hooks in use ( .git/hooks/ is populated, or hook install instructions exist in README/AGENTS.md) but the config file is missing or gitignored. For polyglot repos, partial coverage is acceptable — flag only languages that have active dependencies and no toolchain declaration.  

**Fix:** Pin runtimes per language and commit the lockfile: Go: ensure go.mod has a go directive + toolchain line; commit go.sum . Optionally add to mise.toml : [tools] go = "1.26.3" JS/TS: add .nvmrc or package.json volta field; commit the lockfile; for Bun add bunfig.toml with [run] bun = "1.x.x" Python: add requires-python = ">=3.12" to pyproject.toml and commit uv.lock / poetry.lock ; add .python-version for pyenv/mise users Rust: add rust-toolchain.toml with channel = "1.82.0" and components ; commit Cargo.lock for binaries Swift: add .swift-version ; commit Package.resolved Kotlin/JVM: pin Gradle in gradle-wrapper.properties + commit libs.versions.toml ; add .java-version for toolchain auto-provisioning Ruby: add .ruby-version and commit Gemfile.lock Polyglot: use mise.toml to pin all runtimes in one file — supports Go, Node, Python, Ruby, Java, Rust, Swift and 100+ others Add a committed hook config: hk (Go/Pkl stacks): commit hk.pkl , run hk install husky (JS/TS): commit .husky/pre-commit , run npm run prepare or npx husky install lefthook (polyglot / Go): commit lefthook.yml , run lefthook install pre-commit (Python / polyglot): commit .pre-commit-config.yaml , run pre-commit install simple-git-hooks (minimal JS): add simple-git-hooks key to package.json , run npx simple-git-hooks overcommit (Ruby): commit .overcommit.yml , run overcommit --install cargo-husky (Rust): commit hook scripts to .cargo-husky/hooks/ , add cargo-husky as dev-dependency Reference the setup commands in AGENTS.md so agents know how to bootstrap the environment. Run charter fix AE-ENV-001 — Charter detects the primary language and creates the appropriate toolchain file if absent.  

---

## AE-CTX-004 — .gitignore Missing Agent Artifact Patterns
**Severity:** 🟡 MEDIUM  

**Check:** Does .gitignore exclude local agent session artifacts that should never be committed? Check for these patterns: .charter/ (Charter local audit sessions), *.charter-session files, .claude/local/ (local Claude Code settings), .cursor/cache/ (Cursor local cache), agent scratch/temp directories. Also check for .hk/ or any hk local state that shouldn't be committed — hk.pkl itself should be committed (it's shared team config), but any generated hook cache files should not be. Also verify that agent artifact directories aren't already tracked in git using git ls-files | grep -E '\\.(charter|claude/local|cursor/cache)' .  

**Evidence:** Quote the relevant .gitignore lines if present (or note their absence). Note any agent artifact files that are already tracked in git.  

**False Positive Risk:** FP Risk: Medium. Some .cursor/ and .claude/ subdirs are intentionally committed team config ( .cursor/rules , .claude/settings.json , hk.pkl ). Only flag if local/personal session data or caches are tracked — not shared team config.  

**Fix:** Add to .gitignore: .charter/ , *.charter-session , .claude/local/ , .cursor/cache/ . Keep committed: .cursor/rules , .claude/settings.json , hk.pkl , .mcp.json . Run git rm --cached <path> on any accidentally tracked agent artifacts.  

---

## AE-CI-002 — Charter Action Missing
**Severity:** 🟢 LOW  

**Check:** Is there a GitHub Actions workflow file that runs charter doctor (or uses the Charter GitHub Action) on pull requests? Check .github/workflows/ for any charter-related steps. Verify all of the following: (1) charter doctor --format sarif output is uploaded to GitHub Code Scanning via github/codeql-action/upload-sarif@v4 — SARIF upload is the expected CI artifact for score history and PR annotations; (2) actionlint v1.7.12 runs on all workflow files for syntax and logic validation; (3) zizmor v1.25.2 runs on workflow files for supply-chain security analysis; (4) no third-party actions are pinned to a mutable tag instead of a full commit SHA; (5) optionally, .mcp.json (or .claude/settings.json ) includes a charter entry pointing to charter serve (STDIO transport) — this lets AI agents in the repo invoke charter_doctor and charter_score directly without a subprocess call.  

**Evidence:** Note the workflow file and job name if found. Note the threshold setting. Note whether SARIF upload step is present and wired to Code Scanning. Note whether actionlint v1.7.12 and zizmor v1.25.2 steps exist. List any unpinned third-party actions (tag-pinned rather than SHA-pinned). Note whether a charter serve MCP entry exists in any MCP config file.  

**False Positive Risk:** FP Risk: Very Low. The missing CI check is expected for most repos before Charter adoption — this is the entry-point finding, not a warning sign. Mark N/A only if the repo has no agent-related config at all and Charter isn't relevant. actionlint/zizmor absence is a real finding, not a false positive — both are fast, free, and widely adopted in 2026 GitHub Actions workflows. SARIF upload absence is a genuine gap: without it, score history and PR annotations are lost.  

**Fix:** Create .github/workflows/charter.yaml using the Charter GitHub Action: uses: use-charter/charter-action@v1 with threshold: 80 and SARIF upload enabled. Minimal workflow: - uses: use-charter/charter-action@v1\n with: { threshold: 80, sarif: true }\n- uses: github/codeql-action/upload-sarif@v4\n with: { sarif_file: charter.sarif } Add actionlint and zizmor steps to the same workflow or a dedicated lint-workflows.yaml . Pin all third-party actions to full commit SHAs. Optionally add charter serve as an MCP entry in .mcp.json so AI agents in CI can call charter_doctor and charter_fix natively. Run charter fix AE-CI-002 for a scaffolded workflow file.  

---

## AE-SUPPRESS-001 — Suppression Missing Required Reason
**Severity:** 🟡 MEDIUM  

**Check:** Scan all suppression comments in the repo for bare # charter:ignore AE-RULE-NNN entries that lack a reason=\"…\" field. Every suppression must include a human-readable reason string explaining why the finding is being suppressed. A suppression without a reason is itself a finding — Charter emits AE-SUPPRESS-001 MEDIUM and the suppression is still honored, but the missing reason is flagged.  

**Evidence:** List every suppression comment that lacks a reason field: file path, line number, and rule being suppressed. If all suppressions have reasons, note that explicitly.  

**False Positive Risk:** FP Risk: Very Low. The syntax is unambiguous: # charter:ignore AE-SEC-001 reason=\"test fixture\" is valid; # charter:ignore AE-SEC-001 (no reason) always fails. There are no edge cases — either the reason field is present or it isn't.  

**Fix:** Add a reason string to every bare suppression: # charter:ignore AE-RULE-NNN reason=\"describe why this is safe to suppress here\" . Reasons should be meaningful — 'false positive' alone is not acceptable. State the actual context (e.g., 'test fixture — fake credential with zero real-world access, rotated in CI').  

---

## AE-SUPPRESS-002 — Permanent Suppression Without Approver
**Severity:** 🟠 HIGH  

**Check:** Scan all suppression comments ( # charter:ignore AE-RULE-NNN reason="…" ) and any .charter-suppress.yml file for entries that set ExpiresAt: permanent (or equivalent). For each permanent suppression found, check whether a non-empty ApprovedBy field is present containing a valid GitHub handle. A permanent suppression without an ApprovedBy value is treated by Charter as a 90-day suppression on the next scan — the finding is not actually suppressed long-term.  

**Evidence:** List each permanent suppression entry: file path or suppression key, rule suppressed, and whether ApprovedBy is present. If absent, note that Charter will re-fire the original finding on the next scan.  

**False Positive Risk:** FP Risk: Low. Permanent suppressions are intentionally rare and require explicit opt-in. A permanent entry without ApprovedBy is a genuine governance gap — the repo owner intended a permanent waiver but it won't be honored. Only mark N/A if the repo has no suppression entries at all.  

**Fix:** For each permanent suppression missing ApprovedBy : either add ApprovedBy: github-handle after a security/founder review, or convert to a 90-day TTL suppression. Permanent suppressions require Phase 2 Cloud to be fully enforced at the org level. In v1, enforce as a manual gate: no permanent suppression ships without a documented approver. Syntax: # charter:ignore AE-SEC-001 reason="test fixture" expires=permanent approved-by=tashfiqul-islam .  

---

## AE-SUPPRESS-003 — High Suppression Rate
**Severity:** 🟡 MEDIUM  

**Check:** Count the total number of active suppression entries for this repo (all # charter:ignore comments + any suppression file entries). Compare to the total number of findings Charter emitted on the most recent scan. If suppressed findings exceed 30% of total findings, flag this. This finding does not reduce the Charter Score — it is an informational signal that the suppression log deserves a review. A high suppression rate suggests either systematic false positives (fix the rule) or systematic risk acceptance (needs governance).  

**Evidence:** Total finding count vs. suppressed count from the most recent charter doctor output. Suppression rate = suppressed ÷ (findings + suppressed) × 100. Note the top 2–3 rules being suppressed and whether the suppressions have reasons.  

**False Positive Risk:** FP Risk: Medium. A brand-new repo migrating to Charter may legitimately have many suppressions during the initial calibration period. Mark FP if the repo is clearly in a transition state (Charter installed within the last 30 days, most suppressions dated within the same window). Otherwise, a persistent high suppression rate after 30 days is a genuine governance signal.  

**Fix:** Review the suppression log in bulk. For suppressions covering the same rule across many files: consider whether the rule needs a repo-level exception in charter.yaml → rules.ignore (instead of per-line suppression) or whether the rule's FP rate for this codebase pattern warrants a false-positive report. For suppressions that have grown stale: check expiry dates and remove expired ones. This finding does not affect the Charter Score.  

---
