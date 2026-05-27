# Docs Topology Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure repo documentation so the root stays a small contract surface, internal engineering documentation lives under `docs/internal/`, and future customer-facing documentation has a clean home under `docs/product/` before Phase 1 code growth compounds structure debt.

**Architecture:** Keep root entry docs where contributors and agents expect them, move all repo-internal reference/archive/process materials under `docs/internal/`, and reserve `docs/product/` for future Mintlify/customer docs. Migrate by moving directories first, then fixing links, then updating authority docs and verification paths so no drift survives the rename.

**Tech Stack:** Git, Markdown, HTML mirror docs, Moonrepo docs checks, ripgrep-based link verification, future Mintlify-facing product docs.

---

### Task 1: Freeze the Target Topology

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `ARCHITECTURE.md`
- Modify: `CONTEXT_MAP.md`
- Modify: `docs/architecture/README.md`
- Create: `docs/internal/README.md`
- Create: `docs/product/README.md`

- [ ] **Step 1: Declare the target topology in repo guidance**

Add one short topology block to `README.md` and `AGENTS.md` stating:

```md
Documentation topology:

- root contract docs stay at repo root
- internal engineering docs live under `docs/internal/`
- future customer-facing docs live under `docs/product/`
```

- [ ] **Step 2: Create internal docs landing page**

Create `docs/internal/README.md` with these sections:

```md
# Internal Docs

Use this tree for repo-internal engineering documentation.

Contains:

- architecture references
- audit companions
- ADRs
- RFCs
- rule specs
- playbooks
- runbooks
- superpowers planning/checklists

This tree is not customer-facing documentation.
```

- [ ] **Step 3: Create product docs landing page**

Create `docs/product/README.md` with these sections:

```md
# Product Docs

Reserved for future customer-facing documentation and Mintlify content.

Expected future content:

- getting started
- install
- command reference
- rule reference
- CI and GitHub Action guides
- MCP guidance
- FAQ
```

- [ ] **Step 4: Update architecture authority docs**

Update `docs/architecture/README.md`, `ARCHITECTURE.md`, and `CONTEXT_MAP.md` so they describe the new split clearly:

```md
- root contract docs = entry guidance
- docs/internal = repo engineering knowledge
- docs/product = future customer documentation
```

- [ ] **Step 5: Verify topology language is consistent**

Run:

```powershell
rg -n "docs/internal|docs/product|root contract docs|customer-facing|internal docs" README.md AGENTS.md ARCHITECTURE.md CONTEXT_MAP.md docs
```

Expected: one consistent description of the new structure across trust surfaces.

- [ ] **Step 6: Commit topology freeze**

Run:

```powershell
git add README.md AGENTS.md ARCHITECTURE.md CONTEXT_MAP.md docs/architecture/README.md docs/internal/README.md docs/product/README.md
git commit -m "docs: define internal and product docs topology"
```

### Task 2: Move Internal Reference Trees Under `docs/internal/`

**Files:**
- Move: `decisions/` -> `docs/internal/decisions/`
- Move: `rfcs/` -> `docs/internal/rfcs/`
- Move: `specs/` -> `docs/internal/specs/`
- Move: `playbooks/` -> `docs/internal/playbooks/`
- Move: `runbooks/` -> `docs/internal/runbooks/`
- Move: `docs/architecture/` -> `docs/internal/architecture/`
- Move: `docs/audit/` -> `docs/internal/audit/`
- Move: `docs/superpowers/` -> `docs/internal/superpowers/`

- [ ] **Step 1: Create target directory skeleton**

Run:

```powershell
New-Item -ItemType Directory -Path "docs/internal" -Force
New-Item -ItemType Directory -Path "docs/product" -Force
```

Expected: both parent folders exist before any moves.

- [ ] **Step 2: Move root-level internal trees**

Run:

```powershell
git mv decisions docs/internal/decisions
git mv rfcs docs/internal/rfcs
git mv specs docs/internal/specs
git mv playbooks docs/internal/playbooks
git mv runbooks docs/internal/runbooks
```

Expected: all five directories move with git history preserved.

- [ ] **Step 3: Move current docs subtrees**

Run:

```powershell
git mv docs/architecture docs/internal/architecture
git mv docs/audit docs/internal/audit
git mv docs/superpowers docs/internal/superpowers
```

Expected: internal docs are now consolidated under `docs/internal/`.

- [ ] **Step 4: Keep root contract docs in place**

Do not move these files:

```text
README.md
AGENTS.md
ARCHITECTURE.md
SECURITY.md
CONTRIBUTING.md
TESTING.md
PERMISSIONS.md
CONTEXT_MAP.md
```

Expected: repo entry guidance remains discoverable from root.

- [ ] **Step 5: Verify target directory layout**

Run:

```powershell
Get-ChildItem -LiteralPath "docs" -Force
Get-ChildItem -LiteralPath "docs/internal" -Force
```

Expected: `docs/internal/` contains all migrated internal reference trees and `docs/product/` exists as the future public-doc surface.

- [ ] **Step 6: Commit the directory moves**

Run:

```powershell
git add docs/internal docs/product
git commit -m "refactor: move internal docs under docs/internal"
```

### Task 3: Repair All Paths and Cross-References

**Files:**
- Modify: root contract docs
- Modify: all moved markdown and HTML files with old paths
- Modify: Moon docs-check config if needed
- Modify: workflow or script references if they point to old paths

- [ ] **Step 1: Find all old-path references**

Run:

```powershell
rg -n "decisions/|rfcs/|specs/|playbooks/|runbooks/|docs/architecture|docs/audit|docs/superpowers" .
```

Expected: full list of references that still point at pre-migration paths.

- [ ] **Step 2: Update root contract docs first**

Repair links in:

```text
README.md
AGENTS.md
ARCHITECTURE.md
SECURITY.md
CONTRIBUTING.md
TESTING.md
PERMISSIONS.md
CONTEXT_MAP.md
CLAUDE.md
```

Required mapping examples:

```text
decisions/            -> docs/internal/decisions/
rfcs/                 -> docs/internal/rfcs/
specs/                -> docs/internal/specs/
playbooks/            -> docs/internal/playbooks/
runbooks/             -> docs/internal/runbooks/
docs/architecture/    -> docs/internal/architecture/
docs/audit/           -> docs/internal/audit/
docs/superpowers/     -> docs/internal/superpowers/
```

- [ ] **Step 3: Update moved docs internal references**

Repair relative links inside the moved trees so they still resolve after relocation.

Pay special attention to:

```text
docs/internal/architecture/**/*.md
docs/internal/audit/**/*.md
docs/internal/superpowers/**/*.md
docs/internal/decisions/**/*.md
docs/internal/rfcs/**/*.md
docs/internal/specs/**/*.md
```

- [ ] **Step 4: Update HTML mirror links only where they are internal-path references**

Do not rewrite content unnecessarily. Only fix file-path references or anchor links broken by the move.

- [ ] **Step 5: Verify no stale path references remain**

Run:

```powershell
rg -n "(^|[^a-z])decisions/|(^|[^a-z])rfcs/|(^|[^a-z])specs/|(^|[^a-z])playbooks/|(^|[^a-z])runbooks/|docs/architecture|docs/audit|docs/superpowers" .
```

Expected: only intentionally preserved historical references remain, if any, and each is explainable.

- [ ] **Step 6: Commit path repair**

Run:

```powershell
git add .
git commit -m "docs: repair internal docs references after move"
```

### Task 4: Align Checks, Scripts, and Contributor Workflow

**Files:**
- Modify: `moon.yml`
- Modify: `docs/moon.yml` or its replacement if retained
- Modify: scripts that scan docs paths
- Modify: `.github/workflows/*.yml` only if path assumptions exist

- [ ] **Step 1: Update docs-scan task inputs**

Inspect and repair any task inputs or arguments that assume the old internal docs paths.

Minimum surfaces to check:

```text
moon.yml
scripts/scan-doc-todos.mjs
.github/workflows/*.yml
```

- [ ] **Step 2: Decide what to do with `docs/moon.yml`**

Choose one:

```text
- keep a docs project rooted at docs/
- move project-local Moon file to docs/internal/moon.yml if internal docs are the only tasked docs tree today
```

Recommended:

```text
keep docs tasking at docs/ root for now, since docs/product/ will eventually live beside docs/internal/
```

- [ ] **Step 3: Verify docs checks still work**

Run:

```powershell
mise x -- moon run :docs
mise x -- moon run :lint
```

Expected: docs and lint flows pass after path migration.

- [ ] **Step 4: Verify contributor entry flow still makes sense**

Run:

```powershell
rg -n "docs/internal|docs/product|mise.toml|moon run :check" README.md AGENTS.md CONTRIBUTING.md CONTEXT_MAP.md
```

Expected: first-time contributors can still bootstrap and find the right docs tree in one pass.

- [ ] **Step 5: Commit workflow alignment**

Run:

```powershell
git add moon.yml scripts .github/workflows README.md AGENTS.md CONTRIBUTING.md CONTEXT_MAP.md
git commit -m "build: align tooling with docs topology move"
```

### Task 5: Final Structure Audit and Merge Readiness

**Files:**
- Verify only

- [ ] **Step 1: Run full repo quality gate**

Run:

```powershell
mise x -- moon run :check
```

Expected: full quality gate passes after the topology migration.

- [ ] **Step 2: Run final path audit**

Run:

```powershell
Get-ChildItem -LiteralPath "." -Force
Get-ChildItem -LiteralPath "docs" -Recurse -Depth 2
```

Expected root shape:

```text
root contract docs only at repo root
docs/internal/ for repo engineering docs
docs/product/ reserved for customer docs
```

- [ ] **Step 3: Run final git review**

Run:

```powershell
git status --short
git diff --stat
git log --oneline -10
```

Expected: only intended topology changes remain or the branch is clean.

- [ ] **Step 4: Record timing recommendation**

Write this note into the PR description or merge notes:

```md
Timing recommendation: execute this topology migration before substantive Phase 1 code growth, because internal-doc sprawl compounds quickly once packages, tests, and customer docs expand in parallel.
```

- [ ] **Step 5: Merge readiness decision**

Use this rule:

```text
Merge now if path audits and :check are green.
Delay only if HTML path repair or Moon project-path assumptions remain unresolved.
```
