# Phase 0 Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining repo-executable Phase 0 gaps so Charter's foundation becomes drift-free, reproducible, and strong enough to start Phase 1 from a `9.2/10` baseline that can reasonably reach `9.4/10` after the first disciplined slice.

**Architecture:** Execute closure through five hard gates. Each gate produces a small evidence artifact, updates score movement across the 10-category rubric, and blocks the next gate until contradictions, drift, and behavior ambiguity are eliminated. Work proceeds from doctrine to executable behavior to mirror reconciliation to proof standards to final admission.

**Tech Stack:** Go 1.26 toolchain, Moonrepo, mise, hk, GitHub Actions, Markdown/HTML doc surfaces, Bun helper scripts only if committed as tracked repo tooling.

---

### Task 1: Create the Phase 0 Closure Workspace

**Files:**
- Create: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`
- Create: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/specs/2026-05-28-phase-0-closure-design.md`
- Modify: `docs/superpowers/plans/2026-05-28-phase-0-closure.md`

- [ ] **Step 1: Create the closure scorecard file**

Write `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md` with this exact starter content:

```md
# Phase 0 Closure Scorecard

| Category | Current | Gate 0 | Gate 1 | Gate 2 | Gate 3 | Gate 4 | Target |
|---|---:|---:|---:|---:|---:|---:|---:|
| Core philosophy clarity | 9.5 | - | - | - | - | - | 9.7 |
| Foundation architecture quality | 8.5 | - | - | - | - | - | 9.2 |
| Reproducibility and tooling discipline | 9.0 | - | - | - | - | - | 9.5 |
| CI/CD and repo hygiene | 8.5 | - | - | - | - | - | 9.1 |
| Security posture at foundation stage | 9.0 | - | - | - | - | - | 9.4 |
| Documentation system quality | 8.0 | - | - | - | - | - | 9.1 |
| Specs/contracts readiness | 8.5 | - | - | - | - | - | 8.9 |
| DevX foundation | 8.5 | - | - | - | - | - | 9.3 |
| UX/UI/Product experience foundation | 7.5 | - | - | - | - | - | 8.4 |
| Maintainability / future scale | 9.0 | - | - | - | - | - | 9.5 |
```

- [ ] **Step 2: Create the closure ledger file**

Write `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md` with this exact starter content:

```md
# Phase 0 Closure Ledger

## Gate Status

- Gate 0: Not started
- Gate 1: Not started
- Gate 2: Not started
- Gate 3: Not started
- Gate 4: Not started

## Hard Rules

- No gate may be marked complete without repo evidence.
- No score increase may be claimed without a linked evidence entry.
- No contradiction may be deferred silently.
- Any drift found must be either fixed in-gate or explicitly block the gate.

## Evidence Index

- Pending
```

- [ ] **Step 3: Run a quick existence check**

Run: `Test-Path -LiteralPath "docs\superpowers\checklists"`

Expected: `False` before file creation, `True` after creation.

- [ ] **Step 4: Commit the workspace scaffold**

Run:

```powershell
git add docs/superpowers/specs/2026-05-28-phase-0-closure-design.md docs/superpowers/plans/2026-05-28-phase-0-closure.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md
git commit -m "docs: scaffold phase 0 closure gate artifacts"
```

Expected: commit succeeds with only the four closure-artifact paths staged if no other intended changes are included.

### Task 2: Gate 0 - Truth Model Freeze

**Files:**
- Modify: `docs/architecture/charter-architecture-2026.md`
- Modify: `docs/audit/charter-v1-audit-checklist.md`
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `SECURITY.md`
- Modify: `.github/copilot-instructions.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`

- [ ] **Step 1: Write the contradiction matrix**

Add this section to `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`:

```md
## Gate 0 Contradiction Matrix

| Concern | Files | Current conflict | Canonical owner | Resolution status |
|---|---|---|---|---|
| Bootstrap MCP policy | `SECURITY.md`, `.github/copilot-instructions.md`, `docs/architecture/charter-architecture-2026.md` | Bootstrap docs forbid `.mcp.json`; architecture examples create it | `docs/architecture/charter-architecture-2026.md` after reconciliation | Pending |
| Documentation authority ladder | root docs, architecture docs, audit docs, HTML mirrors | authority is implied, not explicit everywhere | `docs/architecture/README.md` plus canonical product markdown | Pending |
```

- [ ] **Step 2: Reconcile bootstrap MCP policy**

Choose one repo-wide policy and update all six truth surfaces so they say the same thing. Keep the canonical wording in `docs/architecture/charter-architecture-2026.md`, then align the others to that wording.

Required result:

```md
Bootstrap policy:

- a single repo-wide rule is stated in canonical wording
- `docs/architecture/charter-architecture-2026.md` owns that wording
- security summary surfaces paraphrase it without redefining it
- examples do not contradict it
```

- [ ] **Step 3: Add an explicit authority ladder**

Add a short authority block to the most appropriate repo truth surface, then cross-reference it from the others.

Required content:

```md
Documentation authority ladder:
1. `docs/architecture/charter-architecture-2026.md` for product behavior
2. `docs/audit/charter-v1-audit-checklist.md` for manual rule-audit companion detail
3. ADRs in `decisions/` for irreversible constraints
4. root workflow and companion docs for execution guidance only
5. HTML artifacts as presentation mirrors only
```

- [ ] **Step 4: Verify no contradiction remains in core doctrine**

Run:

```powershell
rg -n "mcp.json|authority|canonical|mirror-only|bootstrap" README.md AGENTS.md SECURITY.md .github/copilot-instructions.md docs/architecture docs/audit
```

Expected: all bootstrap and authority references align with the chosen policy and ladder.

- [ ] **Step 5: Update Gate 0 scores and ledger**

Update the scorecard for Gate 0 with only evidence-backed gains.

Expected target entries:

```md
Core philosophy clarity -> 9.7
Security posture at foundation stage -> 9.2
Documentation system quality -> 8.6
Maintainability / future scale -> 9.2
```

- [ ] **Step 6: Commit Gate 0**

Run:

```powershell
git add README.md AGENTS.md SECURITY.md .github/copilot-instructions.md docs/architecture/charter-architecture-2026.md docs/audit/charter-v1-audit-checklist.md docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "docs: freeze phase 0 truth model"
```

### Task 3: Gate 1 - Executable Baseline Coherence

**Files:**
- Modify: `moon.yml`
- Modify: `cmd/moon.yml`
- Modify: `docs/moon.yml`
- Modify: `web/moon.yml`
- Modify: `action/moon.yml`
- Modify: `cloud/moon.yml`
- Modify: `hk.pkl`
- Modify: `scripts/install-hooks.sh`
- Create or track: any helper script referenced by tracked config
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`

- [ ] **Step 1: Create the task-path matrix**

Add this table to the closure ledger:

```md
## Gate 1 Task Path Matrix

| Surface | Setup path | Verify path | Notes |
|---|---|---|---|
| Local root | `mise install` | `moon run :check` | Pending |
| Hooks | `./scripts/install-hooks.sh` | hk-managed `moon run` tasks | Pending |
| CI | `jdx/mise-action` | workflow `moon run` tasks | Pending |
| Project task surfaces | project `moon.yml` files | must reflect root behavior intent | Pending |
```

- [ ] **Step 2: Decide tracked helper-script policy**

If helper scripts are the intended baseline, commit them and reference them consistently. If not, remove tracked references to them from `moon.yml`.

Required rule:

```md
Tracked task config must never depend on untracked local files.
```

- [ ] **Step 3: Unify task behavior across root and project surfaces**

Ensure `cmd/moon.yml` and any other project `moon.yml` surfaces do not accidentally express materially different build/test semantics than root `moon.yml`.

Required outcome:

```md
One behavior model:
- local setup
- hook execution
- CI verification
- per-project task entrypoints
```

- [ ] **Step 4: Verify tracked baseline is self-contained**

Run:

```powershell
git status --short
rg -n "scripts/.*\.mjs|moon run|go test|go build|gitleaks|govulncheck|osv-scanner" moon.yml cmd/moon.yml docs/moon.yml web/moon.yml action/moon.yml cloud/moon.yml hk.pkl .github/workflows scripts
```

Expected: no tracked config points at an untracked dependency and no accidental task split remains.

- [ ] **Step 5: Update Gate 1 scores and ledger**

Expected target entries:

```md
Reproducibility and tooling discipline -> 9.5
CI/CD and repo hygiene -> 8.9
DevX foundation -> 9.0
Foundation architecture quality -> 8.9
```

- [ ] **Step 6: Commit Gate 1**

Run:

```powershell
git add moon.yml cmd/moon.yml docs/moon.yml web/moon.yml action/moon.yml cloud/moon.yml hk.pkl scripts/install-hooks.sh README.md AGENTS.md docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "build: unify phase 0 baseline behavior"
```

### Task 4: Gate 2 - Drift-Free Documentation System

**Files:**
- Modify: `docs/architecture/charter-architecture-2026.md`
- Modify: `docs/architecture/charter-architecture-2026.html`
- Modify: `docs/audit/charter-v1-audit-checklist.md`
- Modify: `docs/audit/charter-v1-audit-checklist.html`
- Modify: any root companion doc that still redefines canonical behavior
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`

- [ ] **Step 1: Add a drift ledger section**

Add this block to the closure ledger:

```md
## Gate 2 Drift Ledger

| File | Section | Drift type | Resolution |
|---|---|---|---|
| `docs/architecture/charter-architecture-2026.html` | version/example sections | stale mirror content | Pending |
| `docs/audit/charter-v1-audit-checklist.html` | mirror-only rule text | stale mirror content | Pending |
```

- [ ] **Step 2: Reconcile canonical markdown before mirrors**

If any markdown source still needs clarification, fix markdown first. Only then update HTML mirrors.

Required rule:

```md
If markdown and HTML disagree, markdown wins first and HTML follows.
```

- [ ] **Step 3: Reconcile both HTML mirrors**

Update HTML so it does not introduce stale versions, stale examples, or behavior not present in canonical markdown.

- [ ] **Step 4: Verify drift closure**

Run:

```powershell
rg -n "zizmor|scorecard|charter serve|AE-ENV-001|MCP|canonical|mirror" docs/architecture docs/audit
```

Expected: mirror surfaces no longer contradict markdown and mirror-only role remains explicit.

- [ ] **Step 5: Update Gate 2 scores and ledger**

Expected target entries:

```md
Documentation system quality -> 9.1
Foundation architecture quality -> 9.1
Maintainability / future scale -> 9.4
```

- [ ] **Step 6: Commit Gate 2**

Run:

```powershell
git add docs/architecture/charter-architecture-2026.md docs/architecture/charter-architecture-2026.html docs/audit/charter-v1-audit-checklist.md docs/audit/charter-v1-audit-checklist.html docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "docs: eliminate phase 0 mirror drift"
```

### Task 5: Gate 3 - Proof-of-Foundation

**Files:**
- Modify: `TESTING.md`
- Modify: `testdata/README.md`
- Modify: `evals/README.md`
- Modify: highest-priority rule specs in `specs/`
- Create: `docs/superpowers/checklists/2026-05-28-first-slice-proof-model.md`
- Create: `docs/superpowers/checklists/2026-05-28-cli-quality-principles.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`

- [ ] **Step 1: Write the first-slice proof model**

Create `docs/superpowers/checklists/2026-05-28-first-slice-proof-model.md` with these sections:

```md
# First Slice Proof Model

## Required for the first Phase 1 slice

- failing test or failing fixture-driven assertion first
- minimal implementation second
- explicit verification command third
- docs/spec alignment update in the same slice
- no silent expansion of scope
```

- [ ] **Step 2: Write CLI quality principles**

Create `docs/superpowers/checklists/2026-05-28-cli-quality-principles.md` with these sections:

```md
# CLI Quality Principles

- quiet mode must be intentionally terse
- findings must be specific and actionable
- machine-readable output must not depend on terminal formatting
- diff-first behavior must be preserved for fix flows
- agent-facing output must remain deterministic
```

- [ ] **Step 3: Strengthen proof-related docs**

Update `TESTING.md`, `testdata/README.md`, and `evals/README.md` so they all point to the same first-slice proof expectations.

- [ ] **Step 4: Deepen the highest-priority rule specs**

Expand the specs for `AE-CTX-001`, `AE-CTX-002`, `AE-CTX-004`, `AE-ENV-001`, and `AE-CI-002` with concrete edge cases, evidence expectations, and remediation shape.

- [ ] **Step 5: Verify proof readiness**

Run:

```powershell
rg -n "proof model|quiet mode|fixture|definition of done|edge case|evidence" TESTING.md testdata/README.md evals/README.md specs docs/superpowers/checklists
```

Expected: the first implementation slice can inherit one explicit proof and output standard.

- [ ] **Step 6: Update Gate 3 scores and ledger**

Expected target entries:

```md
Specs/contracts readiness -> 8.9
UX/UI/Product experience foundation -> 8.4
DevX foundation -> 9.3
Maintainability / future scale -> 9.5
```

- [ ] **Step 7: Commit Gate 3**

Run:

```powershell
git add TESTING.md testdata/README.md evals/README.md specs docs/superpowers/checklists/2026-05-28-first-slice-proof-model.md docs/superpowers/checklists/2026-05-28-cli-quality-principles.md docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "docs: freeze phase 0 proof standards"
```

### Task 6: Gate 4 - Closure Review and Phase 1 Admission

**Files:**
- Create: `docs/superpowers/checklists/2026-05-28-phase-1-admission.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md`
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`
- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Write the Phase 1 admission memo**

Create `docs/superpowers/checklists/2026-05-28-phase-1-admission.md` with this exact starter structure:

```md
# Phase 1 Admission

## Gate Results

- Gate 0: Pass or Fail
- Gate 1: Pass or Fail
- Gate 2: Pass or Fail
- Gate 3: Pass or Fail

## Remaining Known Unknowns

- None in repo-executable closure scope, or a short bullet list if anything remains

## First Phase 1 Slice Start Point

- Repository resolver
- File inventory scanner
- Finding model and score engine
- First simple rules: `AE-CTX-001`, `AE-CTX-002`, `AE-CTX-004`, `AE-ENV-001`, `AE-CI-002`
```

- [ ] **Step 2: Perform final score review**

Update the scorecard with final Gate 4 values and confirm the repo closure target state is documented clearly.

- [ ] **Step 3: Verify closure evidence is complete**

Run:

```powershell
rg -n "Gate 0|Gate 1|Gate 2|Gate 3|Gate 4|pass|fail|Target" docs/superpowers/checklists
```

Expected: each gate has an evidence trail and a pass/fail outcome.

- [ ] **Step 4: Publish Phase 1 starting point in repo guidance**

If appropriate, update `README.md` and `AGENTS.md` to reflect that Phase 0 repo-executable closure is complete and name the first Phase 1 slice start point.

- [ ] **Step 5: Commit Gate 4**

Run:

```powershell
git add README.md AGENTS.md docs/superpowers/checklists/2026-05-28-phase-1-admission.md docs/superpowers/checklists/2026-05-28-phase-0-closure-ledger.md docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "docs: admit phase 1 after phase 0 closure"
```

### Task 7: Final Verification Pass

**Files:**
- Verify only

- [ ] **Step 1: Run the closure grep audit**

Run:

```powershell
rg -n "canonical|mirror-only|bootstrap|mcp.json|quiet mode|proof model|phase 1" README.md AGENTS.md SECURITY.md .github/copilot-instructions.md docs specs TESTING.md evals testdata
```

Expected: no unresolved contradiction surfaces in the repo-executable closure scope.

- [ ] **Step 2: Run the repo quality gate**

Run: `moon run :check`

Expected: pass from the final intended baseline.

- [ ] **Step 3: Run final git review**

Run:

```powershell
git status --short
git diff --stat
git log --oneline -10
```

Expected: only intended closure work remains or everything is committed cleanly.

### Task 8: Post-Closure Score Review

**Files:**
- Modify: `docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md`

- [ ] **Step 1: Defend each category score with evidence**

For each category, add a one-line justification below the score table using this format:

```md
- Core philosophy clarity: 9.7 because bootstrap policy and documentation authority are singular across canonical trust surfaces.
```

- [ ] **Step 2: Add a final overall score line**

Add:

```md
Overall repo-executable Phase 0 closure score: 9.2/10
Expected score after first disciplined Phase 1 slice: 9.4/10
```

- [ ] **Step 3: Commit the score defense**

Run:

```powershell
git add docs/superpowers/checklists/2026-05-28-phase-0-closure-scorecard.md
git commit -m "docs: record phase 0 closure score defense"
```
