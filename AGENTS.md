# AGENTS.md

Last reviewed: 2026-05-23

## Current State

- Phase: Phase 0 repo-executable closure complete; first Phase 1 slice ready
- Product truth: `docs/architecture/charter-architecture-2026.md`
- Module path: `go.charter.dev/charter`
- Current CLI: bootstrap placeholder only

## First Phase 1 Slice

- Repository resolver
- File inventory scanner
- Finding model and score engine
- First simple rules: `AE-CTX-001`, `AE-CTX-002`, `AE-CTX-004`, `AE-ENV-001`, `AE-CI-002`

## Documentation Authority

1. `docs/architecture/charter-architecture-2026.md` for product behavior
2. `docs/audit/charter-v1-audit-checklist.md` for manual audit companion detail
3. ADRs in `decisions/` for irreversible constraints
4. root companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Golden path: `moon run :check`
- Focused: `moon run :test`, `moon run :vet`, `moon run :lint`, `moon run :build`, `moon run :docs`, `moon run :security`, `moon run :eval`

## Hard Constraints

- Model knowledge stale by default.
- Before changing tools, SDKs, CI actions, APIs, MCP, schemas, or frameworks: inspect local manifests and lockfiles, fetch latest official docs, then inspect relevant installed skills or tool docs.
- If latest-docs lookup is unavailable, stop and report reduced confidence.
- Prefer repo evidence over memory.
- Bootstrap keeps tracked MCP config absent until a pinned, reviewed integration exists.
- No LLM calls in Charter core.
- No silent mutation. Diff-first fixes only.
- No secrets in docs, prompts, configs, tests, or logs.
- No speculative refactors outside task scope.

## Edit Scope

- Default edit zones: repo docs, Go source, tests, specs, ADRs, RFCs, workflows, Moon config, mise config
- Off-limits by default: `.env*`, `secrets/`, signing keys, credentials, production infra, generated local state
- Treat `docs/architecture/charter-architecture-2026.md` as canonical for product behavior

## Architecture

- Single root Go module. No `go.work`. No extra modules.
- Command entrypoint in `cmd/charter/`.
- Non-public code in `internal/`.
- Public Go API deferred until a stable external integration surface exists.
- Contract-first for APIs and schemas.
- ADR before irreversible architecture changes. RFC before cross-cutting changes.
- `charter fix` must always diff before apply; never silent mutation.

## Repo Flow

- Hooks managed by `hk` via `hk.pkl`
- Install hooks with `./scripts/install-hooks.sh`
- Pre-commit runs `moon run :lint` and `moon run :docs`
- Commit-msg enforces Conventional Commits
- Pre-push runs `moon run :test` and `moon run :security`

## Context Loading

- `CONTEXT_MAP.md`: load map
- `ARCHITECTURE.md`: module layout, slices, error contracts
- `SECURITY.md`: secrets, MCP, supply-chain posture
- `CONTRIBUTING.md`: workflow, commits, PRs, ADR/RFC expectations
- `TESTING.md`: fixtures, evals, verification commands
- `PERMISSIONS.md`: off-limits paths, escalation, destructive-action policy


<claude-mem-context>
# Memory Context

# [charter] recent context, 2026-05-23 2:14am GMT+6

Legend: 🎯session 🔴bugfix 🟣feature 🔄refactor ✅change 🔵discovery ⚖️decision 🚨security_alert 🔐security_note
Format: ID TIME TYPE TITLE
Fetch details: get_observations([IDs]) | Search: mem-search skill

Stats: 50 obs (20,004t read) | 323,722t work | 94% savings

### May 21, 2026
S476 Understand Charter project HTML and MD docs thoroughly, check consistency, then plan pre-Phase-0 repo bootstrap with interactive architectural decisions (May 21 at 6:26 AM)
S477 Charter Pre-Phase-0 AI Readiness Bootstrap — Initial repo reconnaissance before writing all bootstrap files (May 21 at 6:30 AM)
### May 23, 2026
1675 1:37a 🟣 scorecard.yml Workflow Replaced with Real OpenSSF Scorecard Implementation
1676 " 🔵 govulncheck Has Known Incompatibility with Go 1.26.3 (golang/go#77670)
1677 1:38a ✅ Architecture Doc Reconciled: repo:* Task Naming Replaced with Root :* Tasks
1678 " ✅ HTML Architecture Doc Synced to Match Markdown Source Task Name Changes
1679 " 🔵 mise Not on PATH in Sandboxed Shell; Full Path Required
1680 " 🟣 mise.lock Regenerated with osv-scanner v2.3.5 Across All 11 Target Platforms
1681 1:39a 🔵 osv-scanner v2.3.5 Depends on modelcontextprotocol/go-sdk v1.4.1
1682 " 🔵 Repo Has No Commits Yet — All Files Are Untracked
1683 " 🔵 SECURITY.md Threat Posture Section Explicitly Addresses Prompt Injection and MCP Risks
1684 " 🟣 SECURITY.md Reporting Section Added for Vulnerability Disclosure Process
1685 " 🔵 Residual repo:* References Remain in HTML Architecture Doc Error/Edge/Notes Sections
1686 1:40a 🔵 HTML Architecture Doc Has Mixed Update State: Happy Path Updated, Code Block and Side-Sections Still Use repo:doctor
1687 " ✅ HTML Architecture Doc Final Pass: All Remaining repo:doctor References Replaced
1688 1:41a 🟣 mise.lock Successfully Written with 138 Platform Entries Including osv-scanner
1689 " 🔵 Stale Reference Grep Finds cli: and /cli/ Only in AGENTS.md Historical Memory Entries
1690 1:42a 🔵 Quality Gate Results: :vet, :docs, :actionlint, :zizmor All Pass; zizmor Reports 11 Suppressed Findings
1691 " 🔵 moon run :security Passes Clean: No Leaks, No Vulns, No OSV Issues
1692 " 🔵 SECURITY.md Runbook Link Uses Absolute Local Path Instead of Relative Path
1693 1:43a 🔵 decisions/ Has ADRs 0001–0008 and 0010 but ADR-0009 Is Missing
1694 " 🟣 moon run :check Passes Clean — Phase 0 Umbrella Gate Verified
1695 " 🔵 README.md and AGENTS.md Phase Labels Still Say "pre-Phase-0 bootstrap" — Need Update to Reflect Phase 0 Completion
1696 " ✅ Phase Labels Updated to "Phase 0 Foundation Complete" and ADR-0010 Clarified
1697 " 🟣 Phase 0 Final Gate: moon run :check Passes Clean After All Closure Updates
1698 1:45a 🔵 Final moon.yml Phase 0 Task Graph Confirmed: 10 Tasks with File Group Inputs
1699 1:52a 🔵 Charter Project Security Tooling Documentation Verified Present
1700 1:53a 🔵 renovate.json Implementation Confirmed Matching Architecture Spec
1701 " 🔴 Gitleaks moon.yml Security Task Fixed: --exit-code 0 Suppression Removed
1702 " 🟣 .gitleaks.toml Created to Extend Default Ruleset
1703 " ✅ Architecture Doc and Supporting Docs Aligned on Admin-Side vs Source-Controlled Controls
1704 1:54a 🔵 Architecture Doc Patch Failed — Earlier write_file Calls Were No-ops
1705 " 🔵 Architecture Doc Sections T0.3.1–T0.3.3 Still Contain Stale Content
1706 " 🔴 Architecture Doc T0.3.1–T0.3.3 Sections Successfully Updated on Retry
1707 1:55a 🔵 HTML Export and Audit Checklist Contain Stale Security Content Requiring Separate Updates
1708 1:57a ✅ HTML Export and Audit Checklist Fully Synced with Markdown Architecture Doc
1709 " 🔵 Post-Patch Verification Confirms All Stale Security Terms Removed Across Full Repo
1710 " 🔴 moon.yml Security Task: Gitleaks Explicitly Loads .gitleaks.toml Config + Cache Inputs Expanded
1711 1:58a ✅ HTML Workflow Code Examples Updated to Match Actual Committed Workflows
1712 " 🔴 Final Docs Audit Clean; Single Remaining setup-go@v5 is Phase 1 Code Example Only
1713 " 🔵 Final setup-go@v5 at HTML Line 5679 Confirmed Present but Patch Match Failed
1714 1:59a 🟣 moon run :check Passes All 10 Quality Gates After Documentation and Config Fixes
1715 " ✅ Charter Repository Documentation and Security Configuration Alignment Complete
1716 " 🔵 Charter Repository Has No Git Commits — All Files Are Untracked (Pre-Initial-Commit)
1717 " 🔵 Initial Commit Failed: GPG Signing Requires TTY Not Available in Subprocess Context
1718 2:00a 🔵 git index.lock Error Is Transient — Lock File Does Not Exist
1719 " 🟣 Charter Phase 0 Foundation Committed to git — Initial Commit e8b8921
1720 2:01a 🔵 moon run --force Flag Discovered; Post-Commit Security Scan Passes Against Real Git History
1721 2:06a 🔵 Charter Project Phase 0 Completion Check – Only AGENTS.md Modified
1722 2:07a 🟣 Charter Phase 0 Foundation Committed – 99 Files, 10,939 Lines
1723 " 🔵 Charter Phase 0 – All Moon Check Tasks Pass Clean
1724 " 🔵 AGENTS.md Has Pending Unstaged Changes After Phase 0 Commit

Access 324k tokens of past work via get_observations([IDs]) or mem-search skill.
</claude-mem-context>
