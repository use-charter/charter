# AGENTS.md

Last reviewed: 2026-05-23

## Current State

- Phase: Phase 0 foundation complete; Phase 1 implementation not started
- Product truth: `docs/architecture/charter-architecture-2026.md`
- Module path: `go.charter.dev/charter`
- Current CLI: bootstrap placeholder only

## Commands

- Setup: `mise install` then `./scripts/install-hooks.sh`
- Golden path: `moon run :check`
- Focused: `moon run :test`, `moon run :vet`, `moon run :lint`, `moon run :build`, `moon run :docs`, `moon run :security`, `moon run :eval`

## Hard Constraints

- Model knowledge stale by default.
- Before changing tools, SDKs, CI actions, APIs, MCP, schemas, or frameworks: inspect local manifests and lockfiles, fetch latest official docs, then inspect relevant installed skills or tool docs.
- If latest-docs lookup is unavailable, stop and report reduced confidence.
- Prefer repo evidence over memory.
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

# [charter] recent context, 2026-05-23 1:51am GMT+6

Legend: 🎯session 🔴bugfix 🟣feature 🔄refactor ✅change 🔵discovery ⚖️decision 🚨security_alert 🔐security_note
Format: ID TIME TYPE TITLE
Fetch details: get_observations([IDs]) | Search: mem-search skill

Stats: 50 obs (19,771t read) | 430,174t work | 95% savings

### May 21, 2026
S476 Understand Charter project HTML and MD docs thoroughly, check consistency, then plan pre-Phase-0 repo bootstrap with interactive architectural decisions (May 21 at 6:26 AM)
S477 Charter Pre-Phase-0 AI Readiness Bootstrap — Initial repo reconnaissance before writing all bootstrap files (May 21 at 6:30 AM)
### May 23, 2026
1648 1:21a 🔵 Charter Repo Tool Stack Confirmed from .mise.toml
1649 " 🔵 Charter Moonrepo Task Definitions Confirmed
1650 " 🔵 Charter Phase 0 Milestone Structure Fully Documented in Architecture Doc
1651 " 🔵 Charter Repo Hard Constraints and AI-Readiness Contract Confirmed in AGENTS.md
1652 " 🔵 Charter Security Posture and Standards Targets Confirmed
1654 1:22a 🔵 context7-mcp MCP Tool Unavailable Due to Transport Deserialization Error
1655 " 🔵 Charter Repo Full File Tree Enumerated — Pre-Phase-0 Scaffold Complete
1656 1:23a 🔵 Charter Go Source is Minimal Bootstrap Placeholder — No Domain Logic Exists Yet
1657 " 🔵 golangci-lint v2 Config Uses Standard Linters + gosec + misspell with gofumpt Formatter
1658 " 🔵 hk.pkl Hook Config Confirmed — Pre-commit Runs Lint+Docs, Pre-push Runs Test+Security
1659 " 🔵 .gitignore Includes Charter-Specific Artifact Paths and Tool Cache Exceptions
1660 1:27a 🔵 Context7 MCP Tool Fails with Deserialization Error on resolve-library-id
1661 1:31a ⚖️ Charter Phase 0 Foundation Plan Approved for Implementation
1662 " ⚖️ Charter Go Architecture Rules Locked for Phase 0
1663 " ⚖️ Quality Gate Stack Finalized: :check Umbrella with Explicit :vet Addition
1664 1:32a 🔵 scorecard.yml Workflow Is a Stub Running :docs Instead of OpenSSF Scorecard
1665 " 🔵 Architecture Doc Uses repo: Task Prefix; Repo Implementation Uses Root : Tasks
1666 " 🔵 Charter Repo Companion Docs and CI Templates Already Present
1667 " 🔵 All Five GitHub Workflows Confirmed Pinned with Correct SHA and mise-action Version
1668 1:33a 🔵 Charter Uses Multi-Layer AI Agent Instruction Architecture
1669 " 🔵 Rule Spec Schema Requires 11 Fields Including Traceability to ADRs and Evals
1670 1:35a 🔵 ossf/scorecard-action Latest Release Is v2.4.3 at SHA 4eaacf05
1671 " 🔵 github/codeql-action Latest v3 Release Is v3.36.0, Latest v4 Is v4.36.0
1672 1:36a 🟣 OSV Scanner v2.3.5 Pinned and Wired into :security Gate
1673 " 🟣 Explicit :vet Gate Added to Moon Task Graph
1674 " 🟣 Renovate Configuration Added with Patch Automerge and Pinned Action Digests
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

Access 430k tokens of past work via get_observations([IDs]) or mem-search skill.
</claude-mem-context>
