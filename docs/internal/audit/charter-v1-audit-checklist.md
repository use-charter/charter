# Charter v1 — Rule Audit Checklist

18 rules across context, secrets, MCP, agent-config, environment, CI, testing, autonomy, and governance. Pass all three signals (toolchain + lockfile + hook config) for AE-ENV-001.

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
| Suppressed findings | excluded from score, listed separately |

> Phase 1.5 (not implemented): unscanned/unknown repo state cap ≤ 79.

> ⚠ AE-SUPPRESS-003 is informational — failing it does **not** deduct points or affect the score.

## Reference Metadata

- **Scan engine:** Charter rule engine — high-confidence token detection (secrets) + agent-config rules
- **Agent formats covered:** AGENTS.md · CLAUDE.md · .cursor/rules · .windsurfrules · .github/copilot-instructions.md · opencode.md · codex.md · DESIGN.md · SKILL.md
- **MCP config locations:** .mcp.json · .mcp.yml · .cursor/mcp.json · .claude/settings.json · claude_desktop_config.json · cline_mcp_settings.json · *.pkl
- **Toolchain files (AE-ENV-001):** mise.toml (recommended, polyglot) · go.mod · .nvmrc · bunfig.toml · pyproject.toml · rust-toolchain.toml · .swift-version · gradle-wrapper.properties · .ruby-version · .tool-versions (asdf) · devcontainer.json · flake.nix
- **Hook managers (AE-ENV-001):** hk (hk.pkl, preferred) · husky (.husky/) · lefthook (lefthook.yml) · pre-commit (.pre-commit-config.yaml) · simple-git-hooks · lint-staged · overcommit (.overcommit.yml) · cargo-husky
- **Suppression syntax:** `# charter:ignore AE-RULE-NNN reason="…"`
- **Suppression sources:** `.charter-suppress.yml` (external) and inline `charter:ignore` comment directives (in-source, detected on a finding's own line). Fields: `reason`, `expires`, `approver`.
- **Suppression TTL:** an entry with no `expires` is a default-TTL suppression (honored); `charter suppress` writes an absolute date defaulting to 90 days out. Only the literal `expires: permanent` is a permanent waiver and requires an `approver`; without one it is **not honored** (the finding stays active) and is flagged by AE-SUPPRESS-002 (full org-level permanent enforcement is reserved for Phase 2 Cloud).

---

## AE-CTX-001 — AGENTS.md Missing or Empty
**Severity:** 🔴 BLOCKER  

**Check:** Does the repo root contain an agent context file? Charter recognises all nine formats — check for any of: AGENTS.md , CLAUDE.md , .cursor/rules , .windsurfrules , .github/copilot-instructions.md , opencode.md , codex.md , DESIGN.md (Google Labs standard), SKILL.md (agentskills.io). Is the file non-empty and does it contain meaningful content — at minimum: project summary, tech stack, directories safe for agent edits, off-limits paths, and a verification command (e.g., charter doctor )? Also check: is the file within a reasonable token budget (the v1 limit is a fixed **≤ 600 tokens**, estimated with a deterministic word/char heuristic; files over budget risk being partially ignored by agents)? A per-rule `charter.yaml → rules.AE-CTX-001.token_budget` override is a planned Phase 1.5 item, not yet implemented.  

**Evidence:** File path and format detected. Note if empty or stub-only. Quote the first 2–3 substantive lines. Note the approximate token count if it appears large.  

**False Positive Risk:** FP Risk: Very Low. A file must exist and contain meaningful content. An empty placeholder, a single-line README copy, or a file with only TODO comments still fails. A repo that uses DESIGN.md or SKILL.md instead of AGENTS.md passes if the file meets the content requirements.  

**Fix:** Run charter init to scaffold AGENTS.md from a template matched to your detected tech stack. Minimum content: project summary, tech stack, directories safe for agent edits, off-limits paths, and the command agents should run to verify their changes ( charter doctor ). If you also use charter serve , add a note in AGENTS.md so agents know Charter is available as an MCP tool.  

---

## AE-CTX-002 — Agent Context Stale
**Severity:** 🟡 MEDIUM  

**Check:** If an agent context file exists, is its content consistent with the actual repo state? Compare: (1) stated tech stack vs. language toolchain files — mise.toml (polyglot tool manager, recommended), go.mod (Go), package.json / bunfig.toml (JS/TS), pyproject.toml / .python-version (Python), rust-toolchain.toml / Cargo.toml (Rust), .swift-version (Swift), gradle-wrapper.properties (Kotlin/JVM), .ruby-version (Ruby), .tool-versions (asdf); (2) stated off-limits paths vs. current directory structure; (3) stated test/verify command — a recognized verification command such as `charter doctor` or `moon run :check` — vs. current CI config; (4) stated hook tooling vs. the repo's committed hook config — check all supported managers: hk.pkl (hk, preferred for Go/JS), .husky/ (husky), lefthook.yml / lefthook.toml (lefthook), .pre-commit-config.yaml (pre-commit), simple-git-hooks key in package.json, .lintstagedrc (lint-staged), .overcommit.yml (overcommit), .cargo-husky/hooks/ (cargo-husky) — if a hook manager is active, AGENTS.md should reference it so agents know pre-commit checks will run on their commits; (5) stated MCP tools vs. .mcp.json — if charter serve is configured, AGENTS.md should mention it.  

**Evidence:** Last-modified date of the agent context file. List any specific factual mismatches found (e.g., 'AGENTS.md says Node 18, mise.toml pins Bun 1.3.14' or 'AGENTS.md does not mention charter serve but .mcp.json configures it').  

**False Positive Risk:** FP Risk: Medium. A stale date alone is not enough — focus on factual contradictions between the doc and the repo. Stable repos with slow-moving stacks may have old-dated but accurate docs.  

**Fix:** Update the agent context file to match current stack. Add a Last reviewed: YYYY-MM-DD line at the top. Add a CI lint step that warns when the context file hasn't changed in 90+ days while the repo has active commits ( charter doctor emits AE-CTX-002 at Medium severity when staleness is detected).  

---

## AE-SEC-001 — Raw Secret in Tracked File
**Severity:** 🔴 BLOCKER  

**Check:** Scan agent-visible context files within the repo inventory for raw credential patterns: API keys (sk-*, ghp_*, AKIA*, xoxb-*), private keys (BEGIN RSA/EC/PRIVATE KEY). Charter scans only **tracked** files (via `git ls-files --cached`) — untracked local files and `.gitignore`d files are intentionally not scanned. Focus on: AGENTS.md, CLAUDE.md, .cursor/rules, .windsurfrules, .github/copilot-instructions.md, opencode.md, codex.md, DESIGN.md, SKILL.md. Important: environment-variable references (`${VAR}`, `$VAR`) and the placeholder `your-api-key-here` are neutralized and never flagged.

**Evidence:** File path and redacted match (e.g., 'AGENTS.md: sk-…'), redacted to the first 4 chars + ellipsis; never the raw value.

**False Positive Risk:** FP Risk: Low (implementation uses high-confidence prefix set). High-confidence (mark Fail): exact recognized prefix + correct format length (sk-…T3BlbkFJ, ghp_…, AKIA[A-Z0-9]{16}, xoxb-…, or PEM headers). Environment-variable references and `your-api-key-here` are never flagged — they pass by design.

**Fix:** Revoke the exposed credential immediately. Remove from git history (git-filter-repo or BFG). Switch to environment variable injection: `"key": "${ENV_VAR_NAME}"`. When any AE-SEC-001 or AE-SEC-002 finding is present, Charter score is hard-capped at **49** regardless of other findings. Covers OWASP MCP01 — Token Mismanagement & Secret Exposure.  

---

## AE-SEC-002 — Raw Secret in MCP/Agent Config
**Severity:** 🔴 BLOCKER  

**Check:** Check all MCP config files for literal secret values embedded directly in: env object values, command args arrays, headers objects, or any string value matching a credential pattern. MCP config locations to scan: .mcp.json , .mcp.yml , .cursor/mcp.json , .claude/settings.json , claude_desktop_config.json , cline_mcp_settings.json , *.pkl (MCP Pkl configs). Important: environment-variable references ( `"${MY_API_KEY}"` , `"$MY_KEY"` ) are safe injection patterns and never flagged — these always **pass**. Placeholders like `your-api-key-here` also pass.

**Evidence:** Config file path and redacted value (e.g., '.mcp.json: sk-…'), redacted to the first 4 chars + ellipsis or [REDACTED]; never the raw value.  

**False Positive Risk:** FP Risk: Low (implementation uses high-confidence prefix set). Environment variable references ( `"${MY_API_KEY}"` , `"$MY_KEY"` ) are always safe and never flagged. High-confidence (Fail): literal value matching a recognized credential prefix at correct length (sk-…, ghp_…, AKIA[A-Z0-9]{16}, xoxb-…, or PEM headers). Placeholders and env-refs never flag.

**Fix:** Replace literal secrets with environment variable references: `"env": { "API_KEY": "${MY_API_KEY}" }`. Use a secrets manager to inject values at runtime. When any AE-SEC-001 or AE-SEC-002 finding is present, Charter score is hard-capped at **49** regardless of other findings. Covers OWASP MCP01 — Token Mismanagement & Secret Exposure.  

---

## AE-MCP-001 — MCP Server Unpinned
**Severity:** 🟠 HIGH  

**Check:** For every MCP server in a scanned JSON MCP config (`.mcp.json`, `mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`), check the runner package spec for exact version pinning. Package specs are resolved for direct runners (`npx`, `bunx`, `uvx`) and the `dlx` subcommand of `pnpm`/`yarn`. Flag HIGH when the version is absent, set to `latest` or a dist-tag, a semver range (`^`, `~`, `>=`, `>`, `<`, `*`, `x`), a floating git ref (`github:`, `git+`, or a `#branch`), or a dynamic `${VAR}` value. Exact semver (`1.2.3`) and digests (`sha256:…`, 40-hex) are pinned. Local path args (`./`, `/`, `../`) and `exec`/`run` forms launch local binaries and are not treated as packages.  

**Evidence:** Config file path with a 1-based line, the server name, and the offending package spec (e.g., `.mcp.json:4: server gum uses gumroad-mcp@latest`).  

**Catalog (Slice 13, ADR-0021):** the rule is catalog-aware with a one-finding-per-server precedence ladder — **deprecated > unpinned > advisory > behind-stable > clean**. A package the catalog marks **archived/deprecated** (e.g. `@modelcontextprotocol/server-github`) → HIGH, "migrate to `<successor>`" (even when unpinned). A pinned version in a catalog **advisory** (CVE/GHSA) `affected` set → HIGH, names `id`/`fixedIn`. A pin **behind** the catalog `stable_version` with no advisory → **informational** (re-surfaces, does **not** deduct). Comparison is **exact-match only**; a version absent from `known_versions` is silent (a stale catalog under-reports, never misreports).  

**False Positive Risk:** FP Risk: Low. A server pinned to an exact semver or a digest is safe; semver ranges and floating tags are genuine supply-chain risks. The behind-stable nudge is informational (non-deducting), so catalog staleness cannot create a scoring FP. Mark N/A if the repo has no scanned MCP config files. Non-runner commands (`node`, `python3`, absolute binaries) carry no pin assertion.  

**Fix:** Pin the MCP server package to an exact version or digest instead of `@latest`, a semver range, or a floating git ref (or migrate an archived package to its successor), then commit the change. Covers OWASP MCP04 — Software Supply Chain Attacks & Dependency Tampering.  

---

## AE-MCP-002 — MCP Server Untrusted Remote Origin
**Severity:** 🟠 HIGH  

**Check:** For remote MCP servers (a `url` or `type` of `http`/`sse`) in a scanned JSON MCP config, compare the URL host against the effective allowlist — `union(charter.yaml → mcp.trustedRemotes, catalog trustedHosts)`. The catalog (Slice 13, ADR-0021) ships a baseline of 60+ major vendor-operated remote hosts (GitHub, Vercel, Supabase, Neon, Figma, Stripe, PayPal, Linear, Notion, Atlassian, Sentry, Semgrep, Hugging Face, Context7, Exa, Google Cloud endpoints, the full Cloudflare managed `*.mcp.cloudflare.com` set, and more — reputable vendor servers only), so those pass without per-repo config. Flag HIGH when the host is absent from both. Local/internal origins are exempt: loopback (`127.0.0.0/8`, `::1`), `0.0.0.0`, RFC1918 private ranges, link-local, and the reserved `.localhost`/`.local`/`.internal` TLDs (a LAN/internal server is not a public shadow origin); scheme-less and `${VAR}` URLs have no parseable host and are skipped.  

**Evidence:** Config file path with a 1-based line, the server name, and the resolved host (e.g., `.mcp.json:6: server shadow -> unknown.example.net`).  

**False Positive Risk:** FP Risk: Low–Medium. The catalog baseline clears the common vendor-host case; a remote from a trusted vendor not yet in the catalog or `charter.yaml` is a FP — add the reviewed host to `charter.yaml`. Mark N/A if the repo has no scanned MCP config files. Local/loopback servers never fire.  

**Fix:** Add the reviewed host to `charter.yaml → mcp.trustedRemotes`, or replace the server with a trusted origin, then commit the change. Covers OWASP MCP09 — Shadow MCP Servers.  

---

## AE-MCP-003 — Remote MCP Server Lacks Auth Metadata
**Severity:** 🟠 HIGH  

**Check:** For non-local remote MCP servers (`type` of `http`/`sse`) in a scanned JSON MCP config, check for the presence of an auth header — `Authorization`, `X-Api-Key`, `Api-Key`, or `X-Auth-Token` (case-insensitive; an env-reference value such as `Bearer ${TOKEN}` counts as declared). Flag HIGH when no auth header is present. Detection is presence-based rather than OAuth-field-specific, aligned with MCP spec revision `2025-11-25` and resilient to the `2026-07-28` OAuth changes. Local/internal origins are exempt (loopback, RFC1918 private, link-local, `.localhost`/`.local`/`.internal`), and **catalog-known OAuth vendor hosts are exempt** (they authenticate via the OAuth flow, not a config header).  

**Evidence:** Config file path with a 1-based line, the server name, and the host (e.g., `.mcp.json:3: server open (self-hosted.example) has no auth header`).  

**False Positive Risk:** FP Risk: Low. A genuinely public, read-only remote server may legitimately need no auth — mark N/A with a note. OAuth 2.1 vendor servers (Sentry, Atlassian, Context7, …) are exempt via the catalog (CF-13 resolved). Charter validates only that an auth declaration exists, not the credential itself. Mark N/A if the repo has no scanned MCP config files.  

**Fix:** Declare an auth header (e.g., `Authorization` referencing an environment variable) for the remote MCP server, or switch to a local/trusted integration mode, then commit the change. Covers OWASP MCP07 — Insufficient Authentication & Authorization.  

---

## AE-CC-001 — Dangerous Command in Agent Config
**Severity:** 🔴 BLOCKER  

**Check:** Scan tracked JSON hook config files (`.claude/settings.json`, `.claude/settings.local.json`, `.cursor/hooks.json`) for dangerous shell commands. The parsed `hooks` structure is walked to collect every `command` string and `args` entry (Claude nested + Cursor flat shapes), and each is matched against a high-confidence destructive/privilege-escalation set: `rm -rf`, `git reset --hard`, `git clean -fd`, `dd`, `mkfs`, `truncate`, `sudo`, `chmod 777`, `chown -R`. Single-word commands (`dd`, `truncate`, `mkfs`) match at word boundaries so `git add` and `untruncated` do not false-positive.  

**Evidence:** Config file path with a 1-based line and the offending command (e.g., `.claude/settings.json:7: hook command uses rm -rf`).  

**False Positive Risk:** FP Risk: Low. Operator-chaining and command-substitution injection (`&&`, `;`, `$(…)`, backticks) are intentionally NOT flagged in v1 — they are false-positive-prone (e.g., `cd app && npm test`) and deferred to a context-aware refinement. Mark N/A if the repo has no scanned JSON hook config files. Pkl/YAML/shell hook managers (`hk.pkl`, `.pre-commit-config.yaml`, `lefthook.yml`, `.husky/`) are out of scope for v1.  

**Fix:** Replace the destructive or privilege-escalating command with an explicit, scoped command; prefer array-form (`args`) execution to avoid shell expansion, then commit the change. Covers OWASP MCP05 — Command Injection & Execution.  

---

## AE-CC-002 — Overly Broad Agent Edit Scope
**Severity:** 🟠 HIGH  

**Check:** Read the agent context files (the `agentcontext` set: AGENTS.md, CLAUDE.md, .cursor/rules, .windsurfrules, .github/copilot-instructions.md, opencode.md, codex.md, DESIGN.md, SKILL.md) plus `PERMISSIONS.md`. Pass when the context declares a concrete off-limits / protected-path boundary — a recognized sensitive-path token (`.env`, `secrets`, `.github/workflows`, `terraform`, `infra`, `db/migrations`, `credentials`) or a reference to `PERMISSIONS.md`. Flag HIGH when no concrete off-limits-path declaration is found in any context file. This is stricter than AE-CTX-001, which accepts a generic edit-boundary mention.  

**Evidence:** The context file evaluated (file-level location) plus the list of context files checked.  

**False Positive Risk:** FP Risk: Medium. A small single-purpose repo with no sensitive paths may legitimately have broad scope. The check is presence-based on concrete sensitive-path tokens, not a semantic policy evaluation. When no agent context file exists at all, AE-CTX-001 already fires (Blocker) and AE-CC-002 does not duplicate the absence.  

**Fix:** Add an explicit 'Off-limits for agents' section to the agent context (or `PERMISSIONS.md`) listing at minimum `.github/workflows/`, `terraform/` or `infra/`, `db/migrations/`, `.env*`, and `secrets/`, then commit the change. Covers OWASP MCP02 — Privilege Escalation via Scope Creep.  

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

**False Positive Risk:** FP Risk: Very Low. The missing CI check is expected for most repos before Charter adoption — this is the entry-point finding, not a warning sign. Mark N/A only if the repo has no agent-related config at all and Charter isn't relevant. actionlint/zizmor absence is a real finding, not a false positive — both are fast, free, and widely adopted in 2026 GitHub Actions workflows. SARIF upload absence is a genuine gap: without it, score history and PR annotations are lost. Trusted SLSA reusable workflows ( slsa-framework/slsa-github-generator/.github/workflows/*.yml@vX.Y.Z ) are exempt from the SHA-pin requirement: slsa-verifier resolves the trusted builder identity from the semantic version tag, so SHA-pinning is unsupported. Coverage is recognized for non- moon repos via direct/action forms, not just Charter's moon run :* tasks: repo quality from go test / go build , npm/pnpm/yarn test , cargo test / cargo build , pytest / python -m pytest , bun test , make test / make check ; workflow lint from actionlint + zizmor (direct, or via rhysd/actionlint / zizmorcore/zizmor ); security from govulncheck / osv-scanner / gitleaks / trivy / grype or github/codeql-action . The first-party use-charter/charter-action@<tag> is exempt from the SHA-pin requirement (tag-pinning is the conventional consumer form).  

**Fix:** Create `.github/workflows/charter.yaml` using the Charter GitHub Action: `uses: use-charter/charter-action@v1` with `threshold: 80`. The action uploads SARIF by default (`upload: true`), so a separate `upload-sarif` step is only needed if you disable the built-in upload path and handle SARIF yourself. Add `actionlint` and `zizmor` steps to the same workflow or a dedicated `lint-workflows.yaml`. Pin all third-party actions to full commit SHAs. Optionally add `charter serve` as an MCP entry in `.mcp.json` so AI agents in CI can call `charter_doctor` and `charter_fix` natively. Run `charter fix AE-CI-002` for a scaffolded workflow file.  

---

## AE-SUPPRESS-001 — Suppression Missing Required Reason
**Severity:** 🟡 MEDIUM  

**Check:** Scan all applied suppressions — `.charter-suppress.yml` entries and inline `# charter:ignore AE-RULE-NNN` comments — for entries that lack a `reason` field. Every suppression must include a human-readable reason explaining why the finding is being suppressed. A suppression without a reason is itself a finding — Charter emits AE-SUPPRESS-001 MEDIUM and the suppression is still honored, but the missing reason is flagged.  

**Evidence:** List every suppression that lacks a reason: source (file or inline directive `path:line`) and the rule being suppressed. If all suppressions have reasons, note that explicitly.  

**False Positive Risk:** FP Risk: Very Low. The syntax is unambiguous: # charter:ignore AE-SEC-001 reason=\"test fixture\" is valid; # charter:ignore AE-SEC-001 (no reason) always fails. There are no edge cases — either the reason field is present or it isn't.  

**Fix:** Add a reason string to every bare suppression: # charter:ignore AE-RULE-NNN reason=\"describe why this is safe to suppress here\" . Reasons should be meaningful — 'false positive' alone is not acceptable. State the actual context (e.g., 'test fixture — fake credential with zero real-world access, rotated in CI').  

---

## AE-SUPPRESS-002 — Permanent Suppression Without Approver
**Severity:** 🟠 HIGH  

**Check:** Scan all applied suppressions — inline `# charter:ignore AE-RULE-NNN …` comments and any `.charter-suppress.yml` entries — for an explicit `expires: permanent` waiver. For each, check whether a non-empty `approver` field is present. A permanent waiver without an `approver` is **not honored** — the underlying finding stays active and is scored — and Charter emits AE-SUPPRESS-002 HIGH until an approver is added or a finite `expires` is set. A bare entry with no `expires` is a default-TTL suppression (honored) and is not a permanent waiver.  

**Evidence:** List each permanent waiver: source (file or inline directive `path:line`), rule suppressed, and whether `approver` is present. If absent, note that the finding stays active.  

**False Positive Risk:** FP Risk: Low. Permanent waivers are intentionally rare and require explicit opt-in (`expires: permanent`). A permanent waiver without `approver` is a genuine governance gap — the repo owner intended a permanent waiver but it is not honored. Only mark N/A if the repo has no permanent waivers.  

**Fix:** For each permanent waiver missing `approver`: either add `approver="github-handle"` after a security/founder review, or convert to a finite `expires` date. Org-level enforcement of permanent waivers is reserved for Phase 2 Cloud. Syntax: `# charter:ignore AE-SEC-001 reason="test fixture" expires=permanent approver="tashfiqul-islam"` .  

---

## AE-SUPPRESS-003 — High Suppression Rate
**Severity:** 🟡 MEDIUM  

**Check:** Count the total number of active suppression entries for this repo (all # charter:ignore comments + any suppression file entries). Compare to the total number of findings Charter emitted on the most recent scan. If suppressed findings exceed 30% of total findings, flag this. This finding does not reduce the Charter Score — it is an informational signal that the suppression log deserves a review. A high suppression rate suggests either systematic false positives (fix the rule) or systematic risk acceptance (needs governance).  

**Evidence:** Total finding count vs. suppressed count from the most recent charter doctor output. Suppression rate = suppressed ÷ (findings + suppressed) × 100. Note the top 2–3 rules being suppressed and whether the suppressions have reasons.  

**False Positive Risk:** FP Risk: Medium. A brand-new repo migrating to Charter may legitimately have many suppressions during the initial calibration period. Mark FP if the repo is clearly in a transition state (Charter installed within the last 30 days, most suppressions dated within the same window). Otherwise, a persistent high suppression rate after 30 days is a genuine governance signal.  

**Fix:** Review the suppression log in bulk. For suppressions covering the same rule across many files: consider whether the rule needs a repo-level exception in charter.yaml → rules.ignore (instead of per-line suppression) or whether the rule's FP rate for this codebase pattern warrants a false-positive report. For suppressions that have grown stale: check expiry dates and remove expired ones. This finding does not affect the Charter Score.  

---

## AE-TEST-001 — Automated Tests Missing
**Severity:** 🟠 HIGH  

**Check:** For each **active** code language — one with **both** a project manifest (`go.mod`/`package.json`/`Cargo.toml`/`Gemfile`/`pyproject.toml`/etc.) **and** non-test source outside tooling directories (`scripts/`, `tools/`, `testdata/`, `examples/`, `vendor/`, `node_modules/`, `dist/`, `build/`, `.github/`, `gen/`) **and not referenced by a column-0 `//go:embed` directive in a tracked `.go` file** (embedded web assets bundled into a compiled-language binary are not an independent language surface) — confirm ≥1 recognized test artifact: a file under a `test/`/`tests/`/`spec/`/`__tests__/` segment (any language), or a name convention (Go `*_test.go`; JS/TS `*.{test,spec}.*`; Python `test_*.py`/`*_test.py`/`conftest.py`; Rust inline `#[test]`/`#[cfg(test)]`; Java/Kotlin `*Test.*`/`*Spec.kt`; Ruby `*_spec.rb`/`*_test.rb`; C# `*Tests.cs`; PHP `*Test.php`). Flag HIGH for any active language with no tests.  

**Evidence:** One finding listing each active language lacking tests (e.g., `no test files detected for active language: Go`).  

**False Positive Risk:** FP Risk: Low — validated at **0% FP across 10 real public repos**. The manifest+source dual gate rejects tooling-only manifests and stray secondary-language files; the embedded-asset gate excludes `go:embed`'d web assets (e.g. a single `report.js` in a Go CLI); directory-based test detection catches AVA/tap/RSpec layouts. Mark N/A if no recognized code language is active. Rust inline unit tests count.  

**Fix:** Add tests for the active language(s) so an agent can run them and self-verify before finishing a task. Covers the agent-verifiability gap (ADR-0023).  

---

## AE-AUTO-001 — Verification Command Not Discoverable
**Severity:** 🟡 MEDIUM  

**Check:** When a code language is active, confirm the test command is discoverable via either a task runner declaring a test/check entrypoint (`Makefile`, `justfile`, `Taskfile.yml`, `moon.yml`, `mise.toml [tasks]`, `package.json` `scripts.test`) **or** the active language's conventional zero-config toolchain (Go `go test`, Rust `cargo test`, Python with a configured `pytest`). Flag MEDIUM when neither applies.  

**Evidence:** A finding noting no runner test target and no conventional toolchain for the active language(s).  

**False Positive Risk:** FP Risk: Low. A single-language Go/Rust repo is never penalized for lacking a `Makefile` — its toolchain is the contract. Mark N/A if no language is active.  

**Fix:** Expose a test command via a task runner so an agent can discover and run it. Covers agent operability (ADR-0023).  

---

## AE-CTX-006 — Agent Instructions Over-Emphasized
**Severity:** ⚪ INFORMATIONAL  

**Check:** In the agent context file, measure emphatic-directive density (`IMPORTANT`/`NEVER`/`MUST`/`CRITICAL`/`ALWAYS`/`EXTREMELY`/`ABSOLUTELY`/`FORBIDDEN`/`PROHIBITED`) per 1,000 words. Flag when ≥ 15 per 1K. This finding does not reduce the Charter Score — it is an instruction-quality nudge.  

**Evidence:** Emphatic-directive count, word count, density, and threshold.  

**False Positive Risk:** FP Risk: Low. Informational — never deducts. Grounded in instruction-following research: stacked imperatives create a fragile, competitive instruction topology that degrades adherence.  

**Fix:** Prefer concise, declarative guidance over stacked emphatic directives; state constraints plainly. This finding does not affect the Charter Score.  

---
