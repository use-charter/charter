# Charter — Product Vision & Build Plan 2026

> A single-binary, offline-first SARIF 2.1.0 scanner that gives every repo an AI-agent readiness score.

> This markdown file is the canonical architecture source. `charter-architecture-2026.html` is a presentation artifact with expanded ICP, distribution, monetization, and visual examples. Product behavior, command surface, rule semantics, transports, output formats, MCP tool surface, and phase timing are normative here first; HTML must not introduce behavior-only requirements that are absent from this markdown source.

> Bootstrap policy: Phase 0 bootstrap keeps shared MCP configuration absent by design. Do not introduce `.mcp.json` or other tracked MCP config during bootstrap. Future MCP configuration begins only after a pinned, reviewed integration exists, and root security guidance may paraphrase this rule without redefining it.

---


## §0 Product Brief


### Score Formula

```
score = max(0, 100 − B×20 − H×10 − M×4 − L×1)
final = min(base, applicable_cap)
```

> Suppressed findings are excluded from the base score; they are listed separately.

| Severity | Penalty | Examples |
|---|---|---|
| Blocker | −20 | Dangerous shell commands, raw secrets |
| High | −10 | Unpinned MCP, untrusted remote servers |
| Medium | −4 | Stale context, missing reproducibility |
| Low | −1 | Missing .gitignore patterns, Charter CI absent |

**Score Zones:**

| Range | Zone |
|---|---|
| 0 – 49 | 🔴 Critical — secret cap forces ≤ 49 |
| 50 – 59 | 🔴 Blocked |
| 60 – 79 | 🟡 Needs work |
| 80 – 100 | 🟢 Ship-ready |

**Hard caps** (lowest applicable cap wins):

| Condition | Cap |
|---|---|
| Raw secret detected in any agent-visible file | score ≤ 49 |
| Any Blocker-severity finding present | score ≤ 59 |
| Unscanned or unknown repo state | score ≤ 79 |
| Suppressed findings | excluded from score, listed separately |


### Design Principles

> These are the non-negotiable architectural choices made before writing a single line of domain logic. They are the principles Charter will be held to at every build decision.

**The Ten Commitments:**

1. Never send data anywhere without explicit opt-in.
2. Never call an LLM — all findings are deterministic.
3. Never delete a user file.
4. Never apply a fix without showing a diff first.
5. Every finding has a rule ID, evidence, and fix suggestion.
6. Every release is signed (cosign) with SLSA Level 3 provenance.
7. The score formula is public and unchanging within a major version.
8. Always stays cross-vendor: works with Claude Code, Codex, Cursor, Windsurf, OpenCode, Copilot, Gemini.
9. Never print raw secret values in any output. Never auto-fix secrets destructively. Keep false-positive rate below a measured threshold — enforced by zero-FP assertions on all clean-repo fixtures.
10. The CLI is free forever. No feature is paywalled in the core scan.

**Non-Goals — Do Not Build These:**

- Charter is **not** an AI code reviewer.
- Not a linter for code style.
- Not a package vulnerability scanner (Snyk/Dependabot do that).
- Not a runtime agent monitor.
- Not a secrets vault.
- Not a git history scanner.

**End Goal:**

> The end goal is for Charter to become the default first step before any developer runs an AI agent in any repo — and eventually, the organization-wide control plane that produces the repo, agent, MCP, and policy audit evidence that compliance teams need.

**Tools Explicitly Not Added (and why):**

| Tool | Reason skipped |
|---|---|
| Nix / devenv | mise already covers toolchain pinning |
| Bazel / Earthly | Moonrepo is sufficient; Bazel adds complexity Charter doesn't need |
| Taskfile | Moonrepo + hk already own this surface |
| MegaLinter | golangci-lint + actionlint + zizmor do the job with lower noise |
| full pre-commit | hk handles hooks; adding pre-commit adds Python dep and config drift |
| Docker-heavy local dev | devcontainer.json covers reproducibility; heavy Docker adds friction |
| Kubernetes | Not relevant until Charter Cloud needs multi-tenant infra |
| OpenSSF Allstar | High ROI at org scale; add when use-charter/org has 3+ repos |

## §0.2 Rule Catalog — v1 Launch Set (15 rules)

| Rule ID | Category | What It Checks | Severity |
|---|---|---|---|
| AE-CTX-001 | Context | AGENTS.md (or CLAUDE.md, .cursor/rules, Copilot instructions) present and non-empty with meaningful content — project summary, tech stack, agent edit scope, and a verification command; file within 600-token budget | **Blocker** |
| AE-CTX-002 | Context | Agent context file content consistent with actual repo state — stated tech stack, off-limits paths, and test command match current codebase | **Medium** |
| AE-CTX-004 | Context | .gitignore excludes local agent session artifacts: `.charter/`, `*.charter-session`, `.claude/local/`, `.cursor/cache/`; no agent artifacts already tracked in git | **Medium** |
| AE-SEC-001 | Secrets | No raw secret patterns in any agent-visible file (high-confidence token detection: OpenAI/GitHub/AWS/Slack token prefixes and PEM private-key headers) | **Blocker** |
| AE-SEC-002 | Secrets | No secret-like values in MCP server config (`.mcp.json`, `mcp.yml`, Pkl config) | **Blocker** |
| AE-MCP-001 | MCP Safety | Every MCP server pinned to an exact, **current, non-deprecated** version per the MCP catalog; no `@latest`, semver ranges (`^`, `~`, `>=`), or branch-based sources. Archived packages + CVE/GHSA advisories → High; behind catalog stable → informational (ADR-0021) | **High** |
| AE-MCP-002 | MCP Safety | Every remote MCP server URL present in a known-server catalog or team allowlist (`charter.yaml → mcp.trustedRemotes`); unknown origins flagged — covers OWASP MCP09 Shadow Servers | **High** |
| AE-MCP-003 | MCP Safety | Remote HTTP/SSE MCP servers declare OAuth 2.1 + PKCE authorization metadata per MCP spec 2025-11-25; servers handling sensitive data without an auth declaration flagged — OWASP MCP07 | **High** |
| AE-CC-001 | Agent Config | No dangerous shell patterns in agent hook configs: shell injection (`$()`, backticks, `&&` with open input), destructive commands (`rm -rf`, `git reset --hard`), or privilege escalation (`sudo`, `chmod 777`) — OWASP MCP05 Command Injection | **Blocker** |
| AE-CC-002 | Agent Config | Agent context explicitly restricts edit scope: off-limits paths declared (`.github/workflows/`, `terraform/`, `db/migrations/`, `.env*`, `secrets/`); no implicit full-repo write access granted — OWASP MCP02 Permissioning Failures | **High** |
| AE-ENV-001 | Environment | Reproducible toolchain declared for all active languages — `mise.toml` (recommended, covers any runtime) or a language-native file: Go `go.mod toolchain` · JS/TS `.nvmrc`/`bunfig.toml`/volta · Python `pyproject.toml requires-python` · Rust `rust-toolchain.toml` · Swift `.swift-version` · Kotlin `gradle-wrapper.properties` · Ruby `.ruby-version` · also accepted: `.tool-versions` (asdf), `devcontainer.json`, `flake.nix`; lockfiles committed; hook config committed — `hk.pkl` (hk, preferred) · `.husky/` (husky) · `lefthook.yml` (lefthook) · `.pre-commit-config.yaml` (pre-commit) · simple-git-hooks/lint-staged (JS) · `.overcommit.yml` (Ruby) · `.cargo-husky/hooks/` (Rust) | **Medium** |
| AE-CI-002 | CI | At least one CI workflow runs `charter doctor`; `actionlint` and `zizmor` pass on all workflow files; no unpinned third-party actions | **Low** |
| AE-SUPPRESS-001 | Governance | Suppression comment missing required `reason="…"` field | **Medium** |
| AE-SUPPRESS-002 | Governance | Permanent suppression present without `approver="…"` field | **High** |
| AE-SUPPRESS-003 | Governance | High suppression rate (informational — does **not** deduct points or affect score) | **Medium** |


## §1.8 Command Gallery

Six core commands ship in v1.0: `charter init`, `charter doctor`, `charter report`, `charter fix`, `charter suppress`, and `charter version`. The examples below show the v1 launch surface first, then a clearly labeled Phase 1.5 / v1.1 preview addendum for `charter serve`, `--for-agent`, `--format toon`, and `--format json-compact`.

---

### charter init — Getting Started

Scaffold the *missing* agent context files from a blank repo in under 2 minutes — create-missing-only, never overwriting or deleting a user file (`--dry-run` previews the file plan). Charter detects your language, CI platform, and agent tools automatically and writes `AGENTS.md`, `charter.yaml`, `.gitignore`, `ARCHITECTURE.md`, `.env.example`, and `.claude/settings.json` (when Claude is detected or requested). Running `charter init` then `charter doctor` immediately after scores ≥ 80 out of the box for Go projects — measured 95 on a blank Go repo (residual AE-ENV-001/AE-CI-002 only).

```
❯ charter init
  ✦ Charter v1.0.0  ·  Analyzing project...
  Detected
    language  Go 1.26
    CI        GitHub Actions
    agents    Claude Code  ·  Cursor
  ──────────────────────────────────────────
  Creating
    AGENTS.md ···········  ✓  universal context
    ARCHITECTURE.md ·····  ✓  repo overview template
    .claude/settings.json  ✓  allowed tools without MCP bootstrap
    .env.example ········  ✓  env refs from codebase
    charter.yaml ········  ✓  profile: standard
  ──────────────────────────────────────────
  5 files created  ·  0 skipped  ·  0 overwritten
  › Next: charter doctor
```

```bash
# Non-interactive (CI / scripted setup)
charter init --yes --profile strict --agents claude,cursor
```

---

### charter doctor — Daily Development

The primary command. Run it before committing, in CI, and in agent sessions.
- Exit 0 = pass (score ≥ threshold)
- Exit 1 = fail
- Exit 2 = scan error

**Passing scan:**
```
❯ charter doctor
  ✦ Charter v1.0.0  ·  my-repo / HEAD  ·  31ms
  Detected  Go  ·  Bun  ·  Python  ·  hook  hk
  ──────────────────────────────────────────────────────────────
  Scanning
    context ············  ████████  ✓  3 rules   all passed
    secrets ············  ████████  ✓  2 rules   all passed
    mcp-safety ·········  ████████  ✓  3 rules   all passed
    agent-config ·······  ████████  ✓  2 rules   all passed
    toolchain ··········  ████████  ✓  1 rule    all passed
    ci ·················  ████████  ✓  1 rule    all passed
    governance ·········  ████████  ✓  3 rules   all passed
  Checked 15 rules across 7 groups  ·  0 findings  ✓
  ──────────────────────────────────────────────────────────────
  Score  94/100  ███████████████████░  PASS
  Gate   threshold 80  ·  all clear
  Exit   0
```

**Pre-commit / quiet mode:**
```
# In hk.pkl / lefthook / .pre-commit-config.yaml
# No output on pass — exits silently in <50ms
❯ charter doctor --quiet --threshold 80
  (exit 0 — silent on pass)

# On failure, a single summary line fires:
❯ charter doctor --quiet --threshold 80
  charter: score 49, threshold 80 — FAIL
  # exit 1 — commit blocked
```

**Single-rule targeted scan:**
```
❯ charter doctor --rule AE-SEC-001
  ✦ Charter v1.0.0  ·  12ms  ·  secrets scan only
  ✗ BLOCKER  AE-SEC-001  AGENTS.md:14
  │
  │  Secret detected in agent-visible context file
  │
  │  Evidence  OPENAI_API_KEY=sk-proj-••••••••••••••••
  │  Fix       Use $OPENAI_API_KEY env var instead
  │  Auto-fix  ✗ no
  ─────────────────────────────────────
```

---

### charter doctor — CI / SARIF

In GitHub Actions, pipe findings directly into Code Scanning via SARIF. The GitHub Action (`use-charter/charter-action@v1`) wraps these flags automatically.

**SARIF output:**
```
❯ charter doctor --format sarif \
      --out charter.sarif \
      --threshold 80
  ✦ Charter v1.0.0  ·  SARIF 2.1.0
  Scanned  15 rules  ·  3 findings
  ├  AE-SEC-001  BLOCKER  AGENTS.md:14
  ├  AE-MCP-001  HIGH     .mcp.json:7
  └  AE-ENV-001  MEDIUM   toolchain (missing)
  ──────────────────────────────────────────
  Written  charter.sarif (4.2 KB)
  Score    49/100  FAIL
  Exit     1
  › Upload: github/codeql-action/upload-sarif@v4
```

**Plain CI log output (no colour):**
```
❯ charter doctor --no-color --threshold 80
AE-SEC-001  BLOCKER  AGENTS.md:14   Secret in agent context file
AE-MCP-001  HIGH     .mcp.json:7    MCP server unpinned
AE-ENV-001  MEDIUM   (repo root)    No toolchain declaration
score=49 gate=FAIL threshold=80 exit=1
```

**SPDX output:**
```
❯ charter report --format spdx --out sbom.spdx
  ✦ Charter v1.0.0  ·  SPDX 2.3
  Written  sbom.spdx  (agent tool inventory)
  Exit     0
```

---

### charter fix — Fixing Findings

`charter fix` always shows a unified diff before writing. It never deletes files, never silently mutates, and never auto-fixes secret rules — those require manual remediation (Commitment #4). Backups land in `.charter/backups/` before every write.

The v1 fixers are `AE-CTX-001` (create `AGENTS.md`), `AE-CTX-004` (create/append `.gitignore` agent-artifact patterns), `AE-CI-002` (create the Charter CI workflow), and `AE-MCP-001` (bump an MCP server package to a catalog-known safe version — an advisory-affected pin to its fixed version, or an unpinned/behind-stable cataloged package to the catalog stable; never an archived package, whose migration is manual). Secret, dangerous-command, and toolchain (`AE-ENV-001`) findings are not auto-fixed.

**Dry run — preview before writing:**
```
❯ charter fix --dry-run
AE-MCP-001  .mcp.json
--- a/.mcp.json
+++ b/.mcp.json
@@ -2,1 +2,1 @@
-      "git": { "command": "uvx", "args": ["mcp-server-git@2025.8.0"] }
+      "git": { "command": "uvx", "args": ["mcp-server-git@2026.1.14"] }
AE-CTX-004  .gitignore
--- a/.gitignore
+++ b/.gitignore
@@ -3,3 +3,6 @@
 .claude/local/
 .cursor/cache/
+
+# Charter / agent session artifacts
+.hk/
+.env*
(dry run — no files written)
```

**Apply (originals backed up, then rescan):**
```
❯ charter fix
AE-MCP-001  .mcp.json   (diff as above)
AE-CTX-004  .gitignore  (diff as above)
wrote .mcp.json
wrote .gitignore
backups: .charter/backups/20260601T212756Z
2 fixed
› Re-run: charter doctor
```

**Single-rule fix:**
```
❯ charter fix --rule AE-MCP-001
AE-MCP-001  .mcp.json
--- a/.mcp.json
+++ b/.mcp.json
@@ -2,1 +2,1 @@
-      "git": { "command": "uvx", "args": ["mcp-server-git@2025.8.0"] }
+      "git": { "command": "uvx", "args": ["mcp-server-git@2026.1.14"] }
wrote .mcp.json
backups: .charter/backups/20260601T212756Z
1 fixed
› Re-run: charter doctor
```

A rule with no registered fixer (e.g. a secret rule) reports its no-op explicitly: `charter fix --rule AE-SEC-001` → `AE-SEC-001 is not auto-fixable; remediate manually.`

---

### charter suppress — Managing False Positives

When a finding is a confirmed false positive or accepted risk, suppress it with a machine-readable reason and expiry. Suppressions are stored in `.charter-suppress.yml`, tracked by the governance rules (AE-SUPPRESS-001 – 003), and re-surface automatically after expiry.

```
❯ charter suppress AE-CC-001 \
      --reason "Claude config lives in infra repo" \
      --expires 90d
  ✦ Charter v1.0.0
  Suppressing  AE-CC-001  Missing agent config files
  Reason       Claude config lives in infra repo
  Expires      2026-08-16 (90 days)
  ──────────────────────────────────────────
  ✓ Written  .charter-suppress.yml
  AE-CC-001 suppressed until 2026-08-16.
  Re-surfaces as a finding on expiry.
```

Governance rules audit suppressions every scan:
```
  AE-SUPPRESS-001  suppression missing --reason
  AE-SUPPRESS-002  permanent suppression requires approver
  AE-SUPPRESS-003  high suppression rate (informational)
```

---

### Advanced & Tooling — JSON, version, and Phase 1.5 agent output

`--format json` and `charter version` are part of the v1.0 launch surface. Phase 1.5 / v1.1 adds `--format toon`, `--format json-compact`, and `--for-agent`. `--for-agent` emits TOON-format structured output optimised for AI agent parsing and includes all findings, the score, auto-fix hints, and a `stack:` block with per-language toolchain state.

```
❯ charter doctor --for-agent
CHARTER_SCAN  v1.0.0
repo:       my-repo
score:      49
gate:       FAIL
threshold:  80
findings:   3
stack:
  go      1.26.3  go.mod · no toolchain directive   go.sum ✓
  bun     1.3.14  package.json volta                 bun.lock ✗
  python  3.12.4  pyproject.toml                     uv.lock ✓
  hook    —       none

FINDING 1  BLOCKER  AE-SEC-001  AGENTS.md:14
message:   Secret detected in agent-visible context file
evidence:  OPENAI_API_KEY=sk-proj-[REDACTED]
guidance:  Remove key; use $OPENAI_API_KEY env var instead
auto_fix:  false

FINDING 2  HIGH  AE-MCP-001  .mcp.json:7
message:   MCP server not pinned to a specific version
guidance:  charter fix --rule AE-MCP-001
auto_fix:  true

FINDING 3  MEDIUM  AE-ENV-001  toolchain (incomplete)
message:   Toolchain declarations found for Go only; Bun and Python runtimes active but unpinned
guidance:  Pin Bun and Python in mise.toml (or .nvmrc / pyproject.toml)
auto_fix:  false
```

**JSON output (pipe to jq):**
```
❯ charter doctor --format json | \
      jq '.findings[] | {rule:.rule_id, sev:.severity, cat:.category}'
{"rule": "AE-SEC-001", "sev": "BLOCKER", "cat": "Secrets"}
{"rule": "AE-MCP-001", "sev": "HIGH",    "cat": "MCP Safety"}
{"rule": "AE-ENV-001", "sev": "MEDIUM",  "cat": "Environment"}
```

**charter version:**
```
❯ charter version
charter   1.0.0
commit    abc1234f
built     2026-05-18T09:42:00Z
go        1.26.3
platform  darwin/arm64

❯ charter version --short
1.0.0

❯ charter version --format json
{"version":"1.0.0","commit":"abc1234f","date":"2026-05-18T09:42:00Z","go":"1.26.3","platform":"darwin/arm64"}
```

---

```bash
charter serve                           # Phase 1.5 / v1.1 MCP server mode (STDIO transport)
charter doctor --format toon            # Phase 1.5 / v1.1 TOON renderer
charter doctor --format json-compact    # Phase 1.5 / v1.1 minified JSON renderer
```

`charter serve` exposes four MCP tools in Phase 1.5 / v1.1:
- `charter_doctor` — full scan, structured findings result
- `charter_score` — score + pass/fail only
- `charter_fix` — safe fix planner / diff-returning fixer for supported rules
- `charter_explain` — full rule explanation, evidence guidance, FP notes, remediation


## §2 Phase 0 — The Foundation

Before writing a single line of domain logic, the monorepo infrastructure and context architecture must be pristine. Foundation is everything. This phase has no external launch criteria and no time estimate — it's done when the acceptance criteria pass and the repo is genuinely excellent to work in.

**Phase Objective:** Build the monorepo and context architecture that makes Charter's own repo agent-ready from day one.

**Key Results:**

| KR | Target |
|---|---|
| KR1 | `moon run cmd:test` passes from a clean `git clone` in under 60 seconds |
| KR2 | An AI agent reading only AGENTS.md can navigate to any task in < 3 hops |
| KR3 | `mise install` from scratch produces an identical dev environment |
| KR4 | All 8 seed ADRs written, each under 400 words |

### Bootstrap — Before Any Task in M0.1

> Two one-time prerequisites before T0.1.1 can run: **mise** must be installed at the machine level, and the git repo must exist with you inside it. Neither is a numbered task — they are the preconditions for the entire workspace.

```bash
# 1 — install mise once per machine
brew install mise          # macOS · or: curl https://mise.run | sh

# 2 — create + clone the repo
gh repo create use-charter/charter --private --clone
cd charter

# 3 — proceed: T0.1.1 (mise + hk) → T0.1.2 (Moonrepo) → T0.1.3 (CI)
```


### Tool Versions (mise.toml)

```toml
[tools]
go            = "1.26.3"
bun           = "1.3.14"
moon          = "2.2.6"
hk            = "1.45.0"
golangci-lint = "2.12.2"
gofumpt       = "0.10.0"
actionlint    = "1.7.12"
zizmor        = "1.25.2"
osv-scanner   = "2.3.8"
gitleaks      = "8.30.1"
```


### M0.1 — Repo Infrastructure

*mise + hk → Moonrepo → CI skeleton — bootstrap in this order; each step depends on the last*


#### T0.1.1 Configure mise + hk `⚡ AI`


**User story:** As a contributor, I want tool versions pinned in mise.toml and git hooks wired by hk v1.45.0 (Pkl config), so that every developer on macOS, Linux, and CI runs identical toolchains without manual setup steps.

**Given:** a fresh git clone with mise installed  
**When:** mise install and hk install run  
**Then:** go version , bun --version , and golangci-lint version match the versions declared in mise.toml  

**Happy Path:**
- Fresh git clone + mise install installs: go 1.26.3 , bun 1.3.14 , golangci-lint 2.12.2 , gofumpt 0.10.0 , moon 2.2.6 , hk 1.45.0 , actionlint 1.7.12 , zizmor 1.25.2 , osv-scanner 2.3.8
- git commit triggers pre-commit hook → moon run :lint ; a lint failure blocks the commit with a non-zero exit
- commit-msg hook validates conventional commit format via hk check-commit-msg ; malformed messages are rejected with a clear error
- pre-push hook runs moon run :test ; a failing test blocks the push
- HK_MISE=1 is set so hk activates the mise environment before each hook; hooks do not rely on system PATH
- mise.lock is committed to version control; .gitignore exception is present
- CI job using mise install --no-telemetry completes in under 30 s on a warm runner cache


#### T0.1.2 Initialize Moonrepo Workspace `⚡ AI`


**User story:** As a repo founder, I want Moonrepo v2.2.6 initialized with per-package moon.yml across cmd/ , action/ , web/ , and docs/ , so that all build, test, and lint tasks run from the repo root with deterministic input/output hashing.

**Given:** mise.toml is committed and mise install has run (T0.1.1 complete)  
**When:** moon init runs and per-package moon.yml files are committed  
**Then:** moon run cmd:test and moon run web:lint both succeed from repo root in under 30 s  

**Happy Path:**
- moon run cmd:test executes go test -race ./... from repo root and exits 0
- moon run :build builds all buildable packages in topological order; outputs land in dist/
- moon run web:lint runs Biome check and exits 0 on a clean working tree
- moon.yml task inputs are declared ( **/*.go , .golangci.yml ) so cache hits are deterministic
- Adding a new workspace package requires editing only one moon.yml + workspace.yml — no root config changes
- Moonrepo version is pinned to 2.2.6 in mise.toml ; moon --version matches


#### T0.1.3 CI/CD Skeleton `⚡ AI`


**User story:** As a maintainer, I want a GitHub Actions CI skeleton that runs test/lint/build on every push and PR and triggers GoReleaser v2 on v* tags, so that Charter has a working quality gate from commit one and a release path ready for Phase 1.

**Given:** a push to main or an open PR  
**When:** GitHub Actions evaluates .github/workflows/ci.yml  
**Then:** the test , lint , and web-build jobs all complete successfully and the PR is unblockable only after all three pass  

**Happy Path:**
- CI pipeline runs on push to main and on pull_request targeting main
- Jobs: test ( go test -race ./... ), lint (golangci-lint v2), web-build ( moon run web:build ); all use mise install --no-telemetry
- Pushing tag v0.0.1-alpha triggers release.yml ; GoReleaser v2 stub runs and exits 0 (binary build only — SLSA/cosign wired in T1.5.2)
- CODEOWNERS is present and routes all cmd/** changes to at least one required reviewer
- .github/pull_request_template.md is present with checklist items (tests, docs, ADR if applicable)
- Branch protection on main : all CI jobs required to pass; direct pushes without a PR are blocked
- Workflow files pass actionlint validation (wired in T0.3.1; skeleton must be actionlint-clean from day one)


#### T0.1.TM Trademark Clearance & Domain Acquisition — Hard Phase Gate `⛔ FOUNDER · BLOCKER`


**User story:** As a founder, I want a signed trademark search result and confirmed domain acquisition committed in docs/internal/decisions/0010-trademark-clearance.md , so that Charter can go public without risk of forced rebranding after user adoption.

**Given:** M0.1 is closing  
**When:** T0.1.TM is reviewed at the Phase 0 gate  
**Then:** docs/internal/decisions/0010-trademark-clearance.md exists with status CLEARED or RENAMED, domain registration confirmed, and Phase 0 → Phase 1 gate checklist updated to reflect this as resolved  


### M0.2 — Context Architecture

*AGENTS.md, progressive context loading, CSI task format, ADRs, token budgets*


#### T0.2.1 Write AGENTS.md `⚑ FOUNDER`

**User story:** As a AI coding agent (Claude Code / Cursor / Copilot / Gemini CLI), I want a single AGENTS.md at repo root that is ≤ 600 tokens, so that I can orient to the codebase, understand hard constraints, and start any task without reading more than 3 files.

**Given:** an AI agent opening the Charter repo for the first time  
**When:** the agent reads only AGENTS.md  
**Then:** it knows: the module path, all build commands with moon run syntax, all hard constraints (no LLM calls, no silent mutation, no secret logging), and the vertical-slice architecture rule  

**Happy Path:**
- AGENTS.md token count ≤ 600 (measured with tiktoken cl100k_base); Charter's own AE-CTX-001 scan passes on the file
- File contains all five required sections: Commands, Hard Constraints, Behavioral Principles, Architecture, Context Loading
- All moon run commands use colored span syntax consistent with the rest of the doc: moon =amber, run =dim, scope=blue, task=green
- Hard constraint "Secrets never logged/printed" is present and links to the MustNotContainSecret test assertion pattern
- Hard constraint "charter fix always diffs before applying — never silent mutation" is present (ADR-0005 reference)
- Behavioral Principles section references Karpathy's "Think Before Coding" and "Surgical Changes" principles
- Charter's own scan ( charter doctor ) scores the repo ≥ 90 after AGENTS.md is committed (AGENTS.md present + within token budget = two rules passing)


#### T0.2.2 Companion Context Files `⚑ FOUNDER`


**User story:** As a AI coding agent, I want domain-specific companion context files ( ARCHITECTURE.md , SECURITY.md , CONTRIBUTING.md , TESTING.md ) loaded on demand, so that I receive scoped, deep context without exceeding token budgets on tasks that don't require it.

**Given:** an agent working on a security-related task  
**When:** the agent reads AGENTS.md and then SECURITY.md  
**Then:** it has complete context for the task without reading any other file  

**Happy Path:**
- ARCHITECTURE.md ≤ 800 tokens; covers: module layout, vertical-slice rule, ADR-0001 (zero LLM calls), plugin interface
- SECURITY.md ≤ 500 tokens; covers: secret handling policy, OWASP MCP Top 10 applicability, threat model summary, MustNotContainSecret assertion
- CONTRIBUTING.md ≤ 600 tokens; covers: task workflow, CSI format, PR requirements, conventional commits
- TESTING.md ≤ 600 tokens; covers: fixture repo pattern, table-driven test style, race detector requirement, coverage thresholds
- Each file references which tasks (by T-ID) it is relevant to — context loading is explicit, not heuristic
- AGENTS.md "Context Loading" section lists all four companion files with their token budgets and trigger conditions
- charter doctor validates all companion files exist and are within budget; missing files emit HIGH findings


#### T0.2.3 Task Template (CSI/FBI Format) `⚑ FOUNDER`

**User story:** As a contributor or AI agent opening a task, I want a standardized CSI/FBI task template in .github/TASK_TEMPLATE.md , so that every task has structured, token-efficient context that both humans and AI agents can parse without ambiguity.

**Given:** a new task is being opened for Charter development  
**When:** the contributor uses the task template  
**Then:** the resulting task document has all required sections and an AI agent can extract all needed context from it in a single read  

**Happy Path:**
- Template includes all required CSI sections: CONTEXT , SITUATION , INTENT , CONSTRAINTS , UNKNOWNS , DEFINITION OF DONE
- Template enforces token budget guidance: each section has a comment indicating max recommended tokens
- Template includes a "Behavioral Principles" reminder block quoting "Think Before Coding" / "Surgical Changes" from AGENTS.md
- Template has a DEFINITION OF DONE checklist: tests written, lint passing, ADR filed if architectural, Charter scan passing
- The UNKNOWNS section is non-optional — contributors must explicitly write "None" if no unknowns exist
- Charter's AE-CTX-001 / AE-CTX-002 rules validate that AGENTS.md references the task template format and stays current


#### T0.2.4 ADR Structure + Seed ADRs `⚡ AI`


**User story:** As a architect, I want a docs/internal/decisions/ ADR directory with seed records for the five foundational decisions, so that architectural constraints are documented, discoverable, and machine-enforceable via Charter's own scanning rules.

**Given:** a contributor asks "why does Charter have zero LLM calls in core?"  
**When:** they read docs/internal/decisions/0007-no-llm-calls-in-core.md  
**Then:** they understand the rationale, the alternatives considered, and the consequences — without asking anyone  

**Happy Path:**
- Eight seed ADRs are committed: ADR-0001 through ADR-0008 as defined in the T0.2.4 subtasks
- Each ADR follows the Nygard format: Title, Status, Context, Decision, Consequences
- ADR status values are one of: Proposed | Accepted | Deprecated | Superseded by ADR-XXXX
- All five seed ADRs have status Accepted at commit time
- Charter's AE-CTX-002 rule checks for the presence of a docs/internal/decisions/ directory; missing directory emits a MEDIUM finding
- Each ADR is ≤ 500 tokens (concise, scannable)
- ADR-0008 contains the exact Charter Score formula: max(0, 100 − B×20 − H×10 − M×4 − L×1) with hard cap rules


#### T0.2.5 Specs Scaffold `⚡ AI`


**User story:** As a developer, I want a docs/internal/specs/ directory scaffolded with rule spec files for all 15 v1 rules (12 core + 3 governance), so that rule behavior is documented and testable before implementation begins in Phase 1.

**Given:** Phase 1 implementation begins for AE-CTX-001  
**When:** the developer opens docs/internal/specs/AE-CTX-001.md  
**Then:** they find: rule ID, description, severity, detection logic pseudocode, passing example, failing example, and remediation guidance  

**Happy Path:**
- All 15 v1 rule spec files exist — 12 core: AE-CTX-001 , AE-CTX-002 , AE-CTX-004 , AE-SEC-001 , AE-SEC-002 , AE-MCP-001 , AE-MCP-002 , AE-MCP-003 , AE-CC-001 , AE-CC-002 , AE-ENV-001 , AE-CI-002 ; plus 3 governance: AE-SUPPRESS-001 , AE-SUPPRESS-002 , AE-SUPPRESS-003
- Each spec contains: Rule ID, Severity, Category, Description (≤ 2 sentences), Detection Logic (pseudocode), Pass Example, Fail Example, Remediation
- Severity assignments match the Charter Score penalty table: BLOCKER=−20, HIGH=−10, MEDIUM=−4, LOW=−1
- AE-MCP-001/002/003 specs reference OWASP MCP Top 10 (2025) item numbers
- AE-SEC-001 spec documents the secret-pattern list (API keys, JWTs, PEM headers, suspicious env var assignments)
- Each spec is linked from its corresponding implementation task (T1.x.x) via a reference comment
- charter doctor 's own specs scaffold check validates all 15 specs exist (12 core + 3 governance)


### M0.3 — CI Security & Quality Gates

*actionlint · zizmor · govulncheck · OSV-Scanner · Gitleaks · Renovate · moon run :check*


#### T0.3.1 Actions Security: actionlint + zizmor `⚡ AI`

**User story:** As a platform engineer, I want actionlint v1.7.12 and zizmor v1.25.2 running as required CI jobs on every PR, so that GitHub Actions workflows are free of syntax errors, dangerous permissions, and supply-chain vulnerabilities aligned with SLSA L3 and OWASP MCP Top 10.

**Given:** a PR that modifies any file under .github/workflows/  
**When:** the actions-security CI job runs  
**Then:** actionlint and zizmor both exit 0 for valid workflows, and any violation blocks the PR with an annotated error  

**Happy Path:**
- actionlint v1.7.12 is pinned in mise.toml and runs via moon run :actionlint ; checks all .github/workflows/*.yml files
- zizmor v1.25.2 is pinned in mise.toml and runs via moon run :zizmor ; checks for: unpinned action versions, pull_request_target misuse, script injection via ${{ github.event.* }} interpolation
- zizmor runs in --pedantic mode; all findings are errors (not warnings) — zero tolerance for supply-chain risk
- All existing workflows in the CI skeleton pass both tools before T0.3.1 is considered done
- All workflow action pins use full SHA (40-char commit hash), not mutable tags — enforced by zizmor's unpinned-uses rule
- Workflows default to minimal top-level permissions and grant additional scopes only at the job level when needed (principle of least privilege — OWASP MCP #5)
- The actions-security job is added to branch protection as a required status check
- SARIF upload for workflow tooling is optional and should only be added when the selected tool emits stable SARIF and the repository has code scanning support enabled


#### T0.3.2 Vulnerability Scanning: govulncheck + OSV-Scanner + Gitleaks `⚡ AI`

**User story:** As a security engineer, I want govulncheck, OSV-Scanner v2.3.8, and Gitleaks v8.30.1 running as required repo and CI gates, so that reachable Go vulnerabilities, manifest issues, and leaked secrets are caught before code merges into main .

**Given:** a PR that updates any go.mod , go.sum , or source file  
**When:** the vuln-scan CI job runs  
**Then:** govulncheck reports zero findings for reachable vulnerabilities, OSV-Scanner reports zero high/critical CVEs, and Gitleaks reports zero secret matches  

**Happy Path:**
- govulncheck runs via moon run :security ; scans all Go packages in the module for reachable vulnerabilities against the Go vuln DB (vuln.go.dev)
- OSV-Scanner v2.3.8 is pinned in mise.toml ; runs as osv-scanner scan source -r . against manifests and lockfiles discovered under the working tree
- Gitleaks v8.30.1 is pinned in mise.toml ; scans git history and the current working tree with blocking exit codes on findings
- .gitleaks.toml is committed and extends the built-in default Gitleaks ruleset so future suppressions or repo-specific rules remain reviewable in source control
- All three tools fail the gate on findings; repo-local Phase 0 does not require SARIF output from these scanners
- False-positive suppression for Gitleaks, if ever added, must stay in source control with an explanation for each allowlist entry


#### T0.3.3 Repo Health: Scorecard + CodeQL + Renovate `⚡ AI`

**User story:** As a maintainer, I want Renovate configuration committed in the repository, with Scorecard and CodeQL documented as admin-side or visibility-dependent controls, so that repo health is measurable from source control without pretending private personal repo features are active when they are not.

**Given:** Charter's GitHub repository with CI configured  
**When:** repo workflows and config are committed, and GitHub-native security settings are enabled where supported  
**Then:** Renovate can consume the committed config once its app is installed, and maintainers have exact instructions for enabling or re-enabling Scorecard, CodeQL default setup, and branch protection when repository visibility or hosting plan makes them available  

**Happy Path:**
- Scorecard stays disabled in the private personal repo baseline because the default Actions token cannot complete the required GraphQL queries there; keep a source-controlled stub so future public/org enablement is explicit and reviewable
- CodeQL default setup is preferred over a custom workflow when the repository is public or GitHub Code Security is enabled; this remains an admin-side setting because it is not source-controlled by default
- Renovate config ( renovate.json ) uses config:recommended base plus GitHub Action digest pinning; patch updates may automerge with passing CI; minor and major updates require manual review
- Renovate groups Go dependency updates by module; groups GitHub Actions updates separately
- Renovate runs on a schedule (daily 02:00 UTC) to avoid PR spam during business hours
- SECURITY.md at repo root documents the vulnerability disclosure process for future Scorecard or public-repo hardening
- Branch protection, required checks, and private vulnerability reporting remain admin-side hardening steps and should be enabled where the hosting plan supports them


#### T0.3.4 moon run :check — Single Quality Gate Command `⚡ AI`

**User story:** As a developer or CI pipeline, I want a single moon run :check command that runs all quality gates (actionlint, zizmor, vet, test, build, docs, govulncheck, OSV-Scanner, Gitleaks) through the root task graph, so that I can verify the repository foundation with one command before pushing.

**Given:** a clean working tree on main  
**When:** moon run :check runs  
**Then:** all sub-tasks pass, the command exits 0, and a summary report prints to stdout listing each tool and its result  

**Happy Path:**
- moon run :check runs the root quality gate dependency set: :lint → :vet → :test → :build → :docs → :security → :eval → :actionlint → :zizmor
- Each sub-task is declared as a dependency in moon.yml so Moon's DAG executor handles ordering and caching
- Final stdout summary uses a tabular format: tool name, status (PASS ✓ / FAIL ✗), duration in ms
- If any sub-task fails, :check exits non-zero; the failing task(s) are highlighted in the summary
- moon run :check passes on a fresh clone of Charter's own repo bootstrap
- Charter's own charter doctor . run is deferred until Phase 1 scanner implementation exists
- Total runtime for all gates on a modern laptop should stay comfortably below 90 seconds


## §3 Phase 1 — v1 Launch

"This PR introduces an MCP server pinned to @latest (AE-MCP-001, High) and has no AGENTS.md (AE-CTX-001, Blocker). Charter Score: 49/100. Fix: pin the MCP server to @1.2.3 and create AGENTS.md with the minimum required fields. charter fix --rule AE-CTX-001 can scaffold the file." A developer sees this PR comment and says "that's real and I can fix it in 5 minutes." That is the entire Phase 1 product goal — a finding that is specific, correct, actionable, and earns trust on first contact.


### M1.1 — Foundation Engine

*Core infrastructure: repo resolution, file scanning, finding model, scoring, renderers, CLI skeleton*


#### T1.1.1 Repository Resolver `⚡ AI`


**User story:** As a developer running charter doctor , I want the CLI to reliably resolve the target repository path from flags, environment variables, and current working directory, so that Charter works correctly in local, CI, GitHub Actions, and pre-commit hook contexts without manual path configuration.

**Given:** a developer runs charter doctor from inside a git repository  
**When:** no --path flag is passed  
**Then:** Charter resolves the repo root by walking up from CWD to the .git directory and uses that as the scan target  


#### T1.1.2 File Inventory Scanner `⚡ AI`


**User story:** As a developer, I want Charter to walk the repository tree and build a normalized file inventory that respects .gitignore , so that rule evaluations operate on a complete and accurate file set without scanning vendored, generated, or ignored files.

**Given:** a repository with a populated .gitignore  
**When:** charter doctor runs the file inventory step  
**Then:** the inventory contains all tracked and untracked-but-non-ignored files; files matching .gitignore patterns are excluded  


#### T1.1.3 Finding Model, Charter Score & Policy Profiles `⚡ AI`


**User story:** As a developer or CISO, I want Charter findings typed by severity with a score computed via max(0, 100 − B×20 − H×10 − M×4 − L×1) and hard caps enforced, so that teams can gate pipelines on objective thresholds and apply policy profiles.

**Given:** a scan that produces 1 BLOCKER, 2 HIGH, 3 MEDIUM, 1 LOW finding  
**When:** Charter computes the score  
**Then:** score = max(0, 100 − 20 − 20 − 12 − 1) = 47; the Blocker cap (≤ 59) applies but 47 < 59 so it has no effect; the raw-secret cap (≤ 49, triggered when AE-SEC-001 or AE-SEC-002 fires) is a separate independent cap — not shown here as no secret finding is in the Given; final score is 47  


#### T1.1.4 Output Renderers (Text, JSON, Markdown) `⚡ AI`


**User story:** As a developer or CI pipeline, I want Charter output in `--format text|json|markdown` modes, so that results are human-readable in the terminal, machine-parseable in CI scripts, and embeddable as GitHub PR comments.

**Given:** a completed scan with findings  
**When:** charter doctor --format json is run  
**Then:** stdout is valid JSON matching the Charter output schema; jq .score returns a number; exit code is non-zero if score is below the policy threshold  


#### T1.1.5 CLI Skeleton & Config Loader `⚑ FOUNDER`


**User story:** As a developer, I want a Cobra-based CLI with charter.yaml config loading using XDG + project-root precedence, so that Charter's behavior is consistently configurable across local dev, CI, and multi-repo environments without per-invocation flag sprawl.

**Given:** a charter.yaml in the repo root with policy: profile: strict  
**When:** charter doctor runs without any flags  
**Then:** Charter uses the strict profile from charter.yaml ; no flags required  


#### T1.1.6 Test Infrastructure & Fixture Repos `⚡ AI`


**User story:** As a contributor, I want a testdata/ directory with fixture repositories covering all 12 core v1 rule scenarios (passing and failing) plus dedicated unit tests for the 3 governance rules, so that every rule can be unit-tested against known-good and known-bad states with deterministic, fast tests.

**Given:** a contributor implements AE-CTX-001 (AGENTS.md token budget)  
**When:** they run moon run cmd:test  
**Then:** TestRuleAECTX001_Pass and TestRuleAECTX001_Fail both pass using fixtures from testdata/repos/  


### M1.2 — Agent Config Intelligence

*Parse and normalize all major AI coding agent configuration formats; detect cross-agent conflicts*


#### T1.2.1 Agent Config Parsers — All Seven Agents `⚡ AI`


**User story:** As a developer, I want Charter to parse configuration files for all seven supported AI agents (Claude Code, Cursor, Copilot, Gemini CLI, Windsurf, Aider, Continue), so that agent-specific readiness rules can be evaluated accurately against each agent's actual config schema.

**Given:** a repo with .claude/settings.json , .cursor/rules/ , and .github/copilot-instructions.md  
**When:** charter doctor runs  
**Then:** Charter parses all three agent configs, evaluates applicable rules, and reports findings per agent with the agent name in the finding's context field  


#### T1.2.2 Agent Workspace Normalizer & Conflict Detector `⚡ AI`


**User story:** As a developer, I want Charter to normalize multi-agent workspaces and detect configuration conflicts (overlapping tool permissions, duplicate MCP server registrations, contradictory ignore patterns), so that teams can safely run multiple AI agents in the same repository.

**Given:** a repo with both Claude Code and Cursor configured, each registering the same MCP server under different names  
**When:** charter doctor runs  
**Then:** Charter emits a HIGH finding for AE-CC-002 identifying the duplicate MCP server endpoint and both agents affected  


### M1.3 — MCP + Secrets Security Layer

*Static MCP config scanner, secrets scanner, env checker, toolchain/CI drift detection — 7 of 10 OWASP MCP Top 10 risks covered in v1*


#### T1.3.1 MCP Static Config Scanner `⚡ AI`


**User story:** As a security engineer, I want Charter to statically analyze MCP server configurations against OWASP MCP Top 10 (2025) and MCP spec 2025-11-25, so that unpinned, untrusted, and unauthenticated MCP tool registrations are flagged before they reach an AI agent's context.

**Given:** a Cursor config with an MCP server pinned to @latest and a second remote server URL not in the trusted catalog  
**When:** charter doctor evaluates AE-MCP-001 and AE-MCP-002  
**Then:** Charter emits a HIGH finding for AE-MCP-001 ("MCP server uses floating version — supply chain risk, OWASP MCP04") and a HIGH finding for AE-MCP-002 ("Remote MCP server URL not in trusted catalog — OWASP MCP09")  


#### T1.3.2 Secrets Scanner `⚡ AI`


**User story:** As a security engineer, I want Charter's AE-SEC-001 rule to detect hardcoded credentials in agent context files, AGENTS.md, .env files, and MCP configs, so that secrets never reach version control or an AI agent's context window.

**Given:** a repo with a hardcoded API key in AGENTS.md  
**When:** charter doctor evaluates AE-SEC-001  
**Then:** Charter emits a BLOCKER finding with the file path and a redacted match (key prefix shown, rest replaced with ***); score is hard-capped at 49

**Roadmap:** expand AE-SEC-001 detection from the shipped high-confidence token set (OpenAI/GitHub/AWS/Slack prefixes and PEM private-key headers) to the full Gitleaks v8.30.1 ruleset (160+ detector patterns).  


#### T1.3.3 Env, Toolchain & CI Checkers `⚡ AI`


**User story:** As a platform engineer, I want Charter to detect whether a repo has a reproducible toolchain declaration for all active languages (Go, JS/TS, Python, Rust, Swift, Kotlin/JVM, Ruby — via language-native files or a universal tool like mise), a committed lockfile, and a committed hook config from any supported hook manager, via AE-ENV-001 and AE-CI-002 , so that every project using Charter — regardless of stack — can achieve reproducible, agent-safe environments.

**Given:** a repo with no language-native toolchain file ( go.mod toolchain directive, rust-toolchain.toml , .nvmrc , bunfig.toml , pyproject.toml requires-python, .swift-version , gradle-wrapper.properties , .ruby-version ) and no universal alternative ( mise.toml , .tool-versions , devcontainer.json , flake.nix ), no lockfile, and no committed hook config  
**When:** charter doctor evaluates AE-ENV-001  
**Then:** Charter emits a MEDIUM finding: "No reproducible toolchain declaration found — no language toolchain file, no lockfile, and no committed hook config detected"  


### M1.4 — Repair Engine (Scoped)

*charter init (config generation), safe charter fix (AGENTS.md, .gitignore, GitHub Action workflow only — complex rewrites deferred)*


#### T1.4.1 Config Template Engine (charter init) `⚡ AI`


**User story:** As a developer onboarding a repository to Charter, I want charter init to scaffold all required agent context files with opinionated, correct defaults, so that a repo can go from zero to Charter-compliant in under 2 minutes without manual file creation.

**Given:** a repo with no agent context files  
**When:** charter init runs interactively  
**Then:** it creates AGENTS.md , ARCHITECTURE.md , SECURITY.md , CONTRIBUTING.md , TESTING.md , and a charter.yaml with the project's detected language and policy profile  
**As built:** non-interactive and deterministic (create-missing-only; `--dry-run` previews the plan, `--yes` is the implicit default) — it writes the missing `AGENTS.md`, `charter.yaml`, `.gitignore`, `ARCHITECTURE.md`, `.env.example`, and (when Claude is detected/requested) `.claude/settings.json`, never overwriting or deleting; `SECURITY.md`/`CONTRIBUTING.md`/`TESTING.md` are not scaffolded. A subsequent `charter doctor` scores ≥ 80 out of the box (measured 95 on a blank Go repo, residual AE-ENV-001/AE-CI-002 only). See ADR-0019.  


#### T1.4.2 Fix Planner & Diff Engine (charter fix) `⚡ AI`


**User story:** As a developer, I want charter fix --dry-run to show a unified diff of all auto-fixable findings before applying, and charter fix to apply only safe, reversible changes, so that no file is ever silently mutated (ADR-0005).

**Given:** a scan with 3 auto-fixable findings (AGENTS.md token overrun, missing .gitignore entry for .env, outdated Go version in mise.toml)  
**When:** charter fix --dry-run runs  
**Then:** a unified diff is printed to stdout for each fixable finding; no files are modified; exit code reflects whether fixes are available  
**As built:** the v1 fixers are `AE-CTX-001` (create AGENTS.md), `AE-CTX-004` (create-or-append .gitignore), and `AE-CI-002` (create `.github/workflows/charter.yaml`) — complex/present-but-weak rewrites are deferred and secret/dangerous rules (AE-SEC-001/002, AE-CC-001) are never fixable. `charter fix --dry-run` prints unified diffs and writes nothing; on apply the engine backs up any existing target to `.charter/backups/<ts>/` before each write and never deletes or overwrites a Create target (measured: `fix` raises a representative non-moon repo from 91 to 96). See ADR-0020.  


### M1.5 — CI Integration & Distribution

*GitHub Action (primary product surface), SARIF output, GoReleaser pipeline, Homebrew tap, signed releases, supply chain provenance*


#### T1.5.1 GitHub Action `⚡ AI`


**User story:** As a platform engineer, I want a uses: use-charter/charter-action@v1 GitHub Action that runs charter doctor and uploads SARIF 2.1.0 results to GitHub Security, so that Charter findings appear natively in GitHub's Security tab and can gate PRs via required checks.

**Given:** a repo with uses: use-charter/charter-action@v1 in a workflow  
**When:** a PR is opened  
**Then:** Charter runs, findings are uploaded to the GitHub Security tab as code scanning alerts, and the workflow exits non-zero if the score is below the configured threshold  
**As built:** the composite action downloads the signed release binary, verifies it (cosign keyless + sha256 against `checksums.txt`), runs `charter doctor --format sarif`, and uploads via `github/codeql-action/upload-sarif@v4`; below-threshold gating fails after the alerts upload. Developed in `action/`, seeded to `use-charter/charter-action@v1` at launch.  


#### T1.5.2 Release Pipeline (GoReleaser + Supply Chain) `⚡ AI`


**User story:** As a maintainer, I want GoReleaser v2 producing SLSA L3 provenance, cosign v3.0.6 keyless signatures, and SPDX 2.3 SBOMs for all release artifacts, so that Charter's own supply chain meets the standards it enforces on others — dogfooding its own rules.

**Given:** a tag v1.0.0 is pushed to GitHub  
**When:** the release workflow completes  
**Then:** GitHub Releases contains: binaries for linux/darwin/windows (amd64+arm64), a cosign-signed SBOM in SPDX 2.3 format, SLSA L3 provenance attestation, and Homebrew formula updated  


#### T1.5.3 Performance Validation `👁 REVIEW`


**User story:** As a developer, I want charter doctor to complete in ≤ 2 seconds on a 50,000-file monorepo and all tests to pass under the race detector, so that Charter does not become a bottleneck in local workflows or CI pipelines.

**Given:** a 50,000-file synthetic fixture monorepo  
**When:** charter doctor --path testdata/repos/large-monorepo runs  
**Then:** the command completes in ≤ 2,000 ms wall-clock time, allocates ≤ 256 MB RSS, and exits with the correct score  
**As built:** validated by `moon run :perf` — a build-tagged test synthesizes a ~50,000-file repo at test time (not committed) and asserts `charter doctor` ≤ 2 s wall-clock / (Linux) ≤ 256 MiB peak RSS, race-clean.  


### M1.6 — MCP Catalog v1 — The Recurring Engagement Loop

*Static MCP Catalog v1 — manually curated list of 20–30 common MCP servers with pinned versions. Founder-maintained CVE advisories Phase 1 (48h SLA). Community-contributed catalog PRs + automated advisory monitoring Phase 2.*


#### T1.6.1 Catalog Schema & Seed Data `⚑ FOUNDER`


**User story:** As a developer running charter doctor , I want AE-MCP-001 and AE-MCP-002 to reference a versioned catalog of known safe MCP servers, so that when a server I depend on receives a CVE or releases a new stable version, Charter re-fires without any action on my part.

**As built (Slice 13, ADR-0021):** severity is split by signal class so catalog *staleness stays safe* (a curated catalog tracking a fast-moving ecosystem will lag). Grounding showed the official servers use **CalVer** (`@modelcontextprotocol/server-filesystem` is at `2026.1.14`, not semver) and that the catalog's most durable value is **deprecation** (≈10 popular `@modelcontextprotocol/server-*` packages were archived to vendor/community successors). So AE-MCP-001 is catalog-aware with a one-finding-per-server ladder — **deprecated > unpinned > advisory > behind-stable > clean**:

**Given:** a repo pinning `@modelcontextprotocol/server-github@<v>` (an archived package)  
**When:** charter doctor runs  
**Then:** AE-MCP-001 fires **HIGH**: "MCP server package @modelcontextprotocol/server-github is archived/deprecated — migrate to github/github-mcp-server" — re-firing on a repo that previously passed (the engagement loop).

A pinned version in a catalog **CVE/GHSA advisory** also fires **HIGH** (names `id`/`fixedIn`). A pin merely **behind** the catalog's `stable_version` (no advisory) is **informational** — it re-surfaces but does **not** deduct (mirrors AE-SUPPRESS-003), matching Dependabot/Renovate convention and protecting Commitment #9. Comparison is **exact-match only** (no cross-scheme ordering); a version absent from the catalog's `known_versions` is silent. Rationale: ADR-0021.  


#### T1.6.2 Catalog Contribution & CVE Update Process `⚑ FOUNDER`


**User story:** As a community contributor, I want a documented process for adding MCP servers to the Charter catalog or reporting CVEs against existing entries, so that the catalog stays current without requiring direct Charter team involvement for every update.


#### T1.6.3 Catalog FP Validation — Pre-Ship Gate `⚑ FOUNDER`


**User story:** As the Charter founder, I want documented evidence that the catalog produces a FP rate ≤ 10% on real-world repos before M1.6 ships, so that early adopters do not mute Charter findings because of noise on legitimate servers.

**Given:** 5+ real public repos with committed MCP configs scanned with the M1.6 catalog build  
**When:** all AE-MCP-001 and AE-MCP-002 findings are reviewed and classified  
**Then:** ≤ 10% of findings are false positives, and `docs/internal/catalog/fp-validation.md` documents every finding and its classification  


## §1.7 Phase 1 Exit Criteria — What "Validated" Means

> Define the answer before you start building. The most common OSS tool failure mode is not a bad tool — it's ending Phase 1 without a clear answer on whether to proceed. Stars and install counts are noise. The signals below are the ones that are hard to manufacture and easy to interpret.

**Charter Phase 1 is complete when at least 3 of these 4 signals are true.** If after Phase 1 milestones are shipped (M1.1–M1.6) fewer than 3 signals are green, stop and diagnose before investing in Phase 2 backend work.

**Signal 1 — Organic CI Adoption**
≥ 5 repos running `charter doctor` in CI that you didn't personally tell to install it.
- Check: search GitHub for `uses: use-charter/charter-action` excluding your own org.
- Each organic install is a real developer who decided Charter was worth adding to their pipeline. This is the hardest signal to fake and the most meaningful one.

**Signal 2 — Stranger Issues**
≥ 3 GitHub issues filed by people you don't know personally, with real repo context.
- A stranger filing "AE-MCP-001 fires incorrectly on our monorepo because…" is proof the tool is in active use and real enough to cause friction worth reporting.
- Issues from acquaintances, colleagues, or people you directly asked don't count.

**Signal 3 — Unprompted Mentions**
≥ 1 blog post, HN comment, or social post mentioning Charter that you didn't write or prompt.
- Set up a Google Alert and a GitHub search notification for "charter doctor" and "use-charter".
- A single genuine third-party mention — even a short one — is a stronger signal than 200 stars from a well-timed HN post.

**Signal 4 — Community Self-Help**
≥ 1 GitHub Discussion or community thread where someone answers another person's Charter question — without you being involved.
- Set up GitHub Discussions from day one. Don't answer every question immediately — give the community 24 hours to respond first.
- The first time a stranger helps another stranger understand a Charter rule is the moment Phase 1 has proven something real.

### If Validation Fails — The Honest Diagnosis

> If after M1.6 ships and 60 days have passed, fewer than 3 signals are green, do not proceed to Phase 2. Instead, diagnose:

| Failure pattern | Likely diagnosis | Action |
|---|---|---|
| Many stars, zero organic CI installs | Interesting to read about, not useful enough to add to CI. Rules may be too noisy or fix instructions too vague. | Talk to 5 developers who starred it but didn't install the Action. Ask why. Fix the top complaint. |
| Some installs, zero stranger issues | People installed it, it passed on their repo, they moved on. One-time setup trap is real. | Accelerate the static MCP catalog so new findings fire on repos that previously passed. |
| Issues exist but only from people you know | You have a support network, not a community. | Run a cold outreach to 10 developers you've never met. If none engage, the ICP hypothesis may be wrong. |

## §4 Phase 2 + 3 — Overview


### §4 Phase 2 — Build What Usage Pulls

**Phase 2 Buyer Proof Artifact:**
> "Across 38 repos, these 7 have risky MCP exposure, these 4 policy exceptions are stale, and your org Charter Score dropped 12 points this sprint." A platform lead sees this dashboard and says "I needed this last quarter." That is the entire Phase 2 product goal.

**Phase Objective:** Give platform teams cross-repo visibility into their AI agent exposure. The first Phase 2 product is "what's happening across all our repos?" — not a full control plane yet.

**Key Results:**
- KR1: Charter Cloud shows multi-repo Charter Score trends, MCP inventory, and top recurring findings across all repos in one view
- KR2: Team baselines and exception workflow: who suppressed what, why, when it expires
- KR3: At least 1 paying Team plan customer (≥ 5 seats) actively using the dashboard

> **Phase 2 scope is pulled from Phase 1 usage — not pushed from this document.** Phase 2 builds only what Phase 1 usage proves people need. The features listed below are candidates, not commitments. Each Phase 2 feature ships only if a Phase 1 validation signal points to it.


### §5 Phase 3 — The Enterprise Control Plane

**Phase 3 Buyer Proof Artifact:**
> "Here is your AI coding-agent governance evidence pack for this quarter — ISO 42001 §8.4 AI system lifecycle records, MCP approval history, policy conformance scores, and exception audit trail." A CISO hands this to an auditor. That is the entire Phase 3 product goal.

**Phase Objective:** Make Charter generate the evidence packs that answer ISO 42001 and SOC 2 auditor questions about AI coding agents. Compliance is an organizational claim — Charter is the tooling that makes it credible.

**Key Results:**
- KR1: Charter Cloud generates a complete ISO 42001 §8.4 AI system lifecycle report on demand
- KR2: Enterprise organizations can enforce a central MCP server allowlist across all repos via policy
- KR3: At least 1 enterprise customer (≥ 25 seats) uses Charter as their primary AI coding governance tool

**Phase 3 adds:** SSO/RBAC, ISO 42001 / SOC 2 evidence exports, SIEM integration, central MCP registry, encrypted audit sync, and on-prem deployment options.

> **Do not build Phase 3 features before at least 3 paying Team plan customers from Phase 2.** Phase 3 requires enterprise procurement relationships that Phase 2 Team plan customers develop naturally. Charter becomes the org-wide AI coding governance layer.
