# Phase 0 Closure Design

## Goal

Close the remaining repo-executable Phase 0 gaps so Charter's foundation becomes a drift-free, evidence-backed baseline for Phase 1 implementation.

Target outcomes:

- Overall foundation score moves from `8.7/10` to about `9.2/10` after closure.
- Phase 1 can start without re-deciding doctrine, task behavior, or proof standards.
- The repo reaches these post-closure category targets:

| Category | Current | Target After Closure |
|---|---:|---:|
| Core philosophy clarity | 9.5 | 9.7 |
| Foundation architecture quality | 8.5 | 9.2 |
| Reproducibility and tooling discipline | 9.0 | 9.5 |
| CI/CD and repo hygiene | 8.5 | 9.1 |
| Security posture at foundation stage | 9.0 | 9.4 |
| Documentation system quality | 8.0 | 9.1 |
| Specs/contracts readiness | 8.5 | 8.9 |
| DevX foundation | 8.5 | 9.3 |
| UX/UI/Product experience foundation | 7.5 | 8.4 |
| Maintainability / future scale | 9.0 | 9.5 |

## Scope

This design covers repo-executable closure only.

Included:

- Doctrine and source-of-truth cleanup
- Task graph, hook, CI, and reproducibility cohesion
- Drift removal across markdown and HTML mirrors
- Foundation proof expectations for early Phase 1 work
- Final closure review and Phase 1 admission criteria

Excluded:

- Founder/external gates such as trademark or domain acquisition
- Phase 1 product implementation itself

## Non-Negotiable Standards

- Zero tolerance on drift across canonical repo truth surfaces
- Repo evidence first
- Latest-doc alignment for any tooling, API, workflow, or security recommendation
- 2026-aligned security and workflow posture only when repo or verified docs support it
- Deterministic behavior over convenience
- No silent mutation or ambiguous ownership of behavior definitions
- One reproducible setup and verification path for local, hook, and CI use

## Recommended Closure Model

Use hard gates, not a flat checklist.

Rationale:

- prevents fake progress
- keeps accountability objective
- makes score movement evidence-based instead of aspirational
- matches the repo's philosophy of deliberate, reviewable progression

## Gate Structure

### Gate 0: Truth Model Freeze

Objective:

- Establish one explicit authority ladder for repo behavior and remove contradictions across core doctrine files.

Primary scope:

- `docs/architecture/charter-architecture-2026.md`
- `docs/audit/charter-v1-audit-checklist.md`
- `README.md`
- `AGENTS.md`
- `SECURITY.md`
- `.github/copilot-instructions.md`

Must resolve:

- bootstrap MCP policy contradiction
- canonical-vs-companion-vs-mirror hierarchy
- any rule or workflow behavior defined inconsistently across trust surfaces

Exit criteria:

- one documentation authority ladder exists and is explicit
- bootstrap security stance has one interpretation only
- no known contradiction remains in repo-core doctrine
- HTML artifacts are clearly mirror-only

Expected category movement:

- Core philosophy clarity: `9.5 -> 9.7`
- Security posture: `9.0 -> 9.2`
- Documentation system quality: `8.0 -> 8.6`
- Maintainability / future scale: `9.0 -> 9.2`

### Gate 1: Executable Baseline Coherence

Objective:

- Make setup, tasks, hooks, and CI behave as one coherent system.

Primary scope:

- `mise.toml`
- `mise.lock`
- `go.mod`
- `moon.yml`
- `.moon/workspace.yml`
- `cmd/moon.yml`
- `docs/moon.yml`
- `web/moon.yml`
- `action/moon.yml`
- `cloud/moon.yml`
- `hk.pkl`
- `scripts/install-hooks.sh`
- any tracked helper script required by tracked config

Must resolve:

- root task model vs project task model divergence
- tracked config depending on untracked files
- local, hook, and CI path ambiguity
- reproducibility gaps in the intended baseline

Exit criteria:

- root and project tasks reflect one behavior model
- tracked baseline is self-contained
- hooks and CI enforce the same intended contract
- local-only files are clearly outside the repo contract

Expected category movement:

- Reproducibility and tooling discipline: `9.0 -> 9.5`
- CI/CD and repo hygiene: `8.5 -> 8.9`
- DevX foundation: `8.5 -> 9.0`
- Foundation architecture quality: `8.5 -> 8.9`

### Gate 2: Drift-Free Documentation System

Objective:

- Make documentation rich and reliable at the same time.

Primary scope:

- `docs/architecture/*.md`
- `docs/architecture/*.html`
- `docs/audit/*.md`
- `docs/audit/*.html`
- companion root docs that restate behavior

Must resolve:

- markdown/HTML drift
- stale version references
- stale examples
- secondary docs redefining canonical behavior

Exit criteria:

- HTML no longer contradicts markdown
- secondary docs no longer redefine canonical behavior
- doc classes have explicit maintenance boundaries
- future drift can be audited repeatably

Expected category movement:

- Documentation system quality: `8.6 -> 9.1`
- Foundation architecture quality: `8.9 -> 9.1`
- Maintainability / future scale: `9.2 -> 9.4`

### Gate 3: Proof-of-Foundation

Objective:

- Freeze how the foundation proves itself before major implementation starts.

Primary scope:

- `TESTING.md`
- `testdata/README.md`
- `evals/README.md`
- `specs/*.md`
- first-slice proof model docs
- CLI quality and output principles

Must define:

- first executable proof standard for Phase 1 slices
- fixture and test-harness expectations
- CLI output quality rules
- definition of done for early implementation slices

Exit criteria:

- the first Phase 1 slice has a predefined proof model
- specs are actionable enough to constrain implementation
- CLI quality principles are documented enough to guide implementation
- testing expectations are operational, not just aspirational

Expected category movement:

- Specs/contracts readiness: `8.5 -> 8.9`
- UX/UI/Product experience foundation: `7.5 -> 8.4`
- DevX foundation: `9.0 -> 9.3`
- Maintainability / future scale: `9.4 -> 9.5`

### Gate 4: Closure Review and Phase 1 Admission

Objective:

- Decide whether Phase 0 is truly closed for repo-executable scope.

Primary scope:

- all prior gates
- final category score sheet
- known-unknowns ledger
- Phase 1 admission memo

Exit criteria:

- all prior gates passed
- no unresolved contradiction remains in repo-executable closure scope
- category targets are met or any delta is explicitly defended with evidence
- exact starting point for the first Phase 1 slice is documented

## Accountability Rules

Each gate must produce:

- objective pass/fail criteria
- a small evidence artifact or checklist result
- updated score movement against the category rubric
- explicit stop/go decision for the next gate

The closure effort fails if any of these remain unresolved at Gate 4:

- doctrine contradiction
- tracked-vs-local baseline ambiguity
- root-vs-project task behavior split
- markdown-vs-HTML behavior drift
- undefined proof expectations for the first implementation slice

## Risks to Watch

- Over-documenting instead of resolving contradictions
- Using generic "2026 best practice" language without verified evidence
- Inflating category scores without repo-backed proof
- Carrying local-only behavior into the tracked baseline by accident

## Success Definition

Phase 0 closure succeeds when:

- repo doctrine is singular
- automation is coherent
- docs are trustworthy
- proof expectations are frozen
- Phase 1 can begin without re-deciding foundational principles
