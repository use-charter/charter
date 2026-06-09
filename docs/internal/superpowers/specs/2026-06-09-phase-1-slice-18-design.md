# Phase 1 Slice 18 Design — External Docs on Mintlify

## Goal

Launch a public documentation system for Charter on Mintlify that is accurate, adoption-friendly, and launch-ready. It must (a) onboard a new user quickly, (b) provide complete CLI, rule, config, and CI reference, (c) resolve every existing rule `helpUri` under `https://use-charter.dev/rules/AE-*`, (d) fit the repo's docs topology and launch roadmap, and (e) stay maintainable with minimal drift from canonical internal specs. Implements roadmap Slice 18; depends on **Slices 9–17** (the shipped command surface + the full rule set + the hardened codebase). Does not build the launch website (Slice 19).

## Audience

- **Primary**: developers evaluating Charter for a repo, developers adopting Charter, developers investigating a specific finding or rule.
- **Secondary**: platform/security engineers integrating Charter in CI, maintainers validating public docs against product truth.
- **Not**: internal contributors reading engineering docs, agent-only consumers of machine-readable output, future hosted/org dashboard users.

## Scope

### In scope
- Mintlify project in `docs/product/` — `docs.json`, MDX pages, navigation, redirects, SEO/search baseline.
- Tutorial: `quickstart` — install/run Charter, scan a repo, interpret pass/fail.
- How-to guides: run in GitHub Actions, adopt in existing repo, suppress a finding, investigate MCP findings, use `charter fix` safely.
- CLI reference for all 7 commands: `doctor`, `init`, `fix`, `report`, `explain`, `suppress`, `version`.
- Rule reference for all 18 launch rules — one page per rule at `/rules/AE-*` URLs.
- `charter.yaml` config reference: `policy.profile`, `policy.threshold`, `mcp.trustedRemotes`.
- GitHub Action guide: inputs/outputs, permissions, common variants.
- Explanation/concept pages: agent-readiness model, scoring and caps, MCP safety, suppression governance.
- Rule-page generator (Bun TS script) that produces `rules/AE-*.mdx` from `docs/internal/specs/AE-*.md` + `internal/rules/catalog/catalog.go`.
- Deployment design: Mintlify origin on `docs.use-charter.dev`; root-domain routing for `/docs/*` and `/rules/*` so `helpUri`s resolve.
- Docs validation tasks wired into `moon run :docs`.

### Out of scope
- Launch website on `use-charter.dev` (Slice 19).
- Hosted dashboard/cloud docs, public API/OpenAPI reference (no public API exists yet).
- Translations/languages, multi-version docs, blog/changelog/news.
- Changing rule behavior, command behavior, `helpUri` scheme.
- Analytics — deployment-time config choice, not content structure.

## Grounding (verified 2026-06-09)

- **Repo constraints**: `docs/product/` is already reserved for customer-facing docs (`README.md:35-39`, `docs/product/README.md`). The roadmap explicitly requires Mintlify, quickstart, CLI reference, CI guide, full rule reference, config reference, suppression governance, and live `/rules/AE-*` pages (`roadmap md:112-118`). Rule `helpUri`s already point to `https://use-charter.dev/rules/<RULE>` (`internal/rules/catalog/catalog.go:19-39`). Internal rule specs in `docs/internal/specs/AE-*.md` are the canonical semantics source.
- **Mintlify capabilities (Context7-verified)**: central site config via `docs.json`; navigation supports tabs, groups, anchors, dropdowns, versions, languages; redirects are native in `docs.json`; `/docs` subpath hosting via proxy/rewrites; SEO metadata and search prompt configurable; common analytics integrations available.
- **2026 docs standards**: Diátaxis framework (quadrants: Tutorial, How-to, Reference, Explanation) is the dominant modern docs architecture. Developer-tool docs consistently optimize for adoption-first onboarding in 2026. Self-contained offline reports already ship (Slice 16); the docs remain a hosted surface, which is appropriate for a discoverability/browsing surface that the HTML report supplements.

## Content architecture

### Diátaxis quadrants

| Quadrant | Pages | Purpose |
|----------|-------|---------|
| Tutorial | quickstart | Fastest successful outcome. First scan in 5 minutes. |
| How-to | run-in-github-actions, adopt-in-existing-repo, suppress-a-finding, investigate-mcp-findings, use-charter-fix-safely | Task-completion recipes. |
| Reference | CLI reference (8 pages), rule reference (18 pages), charter-yaml config, github-action, scoring-and-caps | Exact technical surface. |
| Explanation | agent-readiness-model, mcp-safety-model, suppression-governance | Why the product exists and works as it does. |

This split is better than a flat "guides/reference" layout because it serves both adoption and trust.

### Top-level navigation (Mintlify tabs)

- **Docs** — all non-rule content
- **Rules** — rule overview + 18 per-rule pages

No `API` tab in v1 — no public API surface yet.

### Docs tab groups

- Getting Started
- How-to Guides
- CLI Reference
- Configuration
- CI & GitHub Action
- Concepts

### Rules tab groups

Group by category, preserving flat `/rules/AE-*` URLs:

- Context (4 rules)
- Secrets (2 rules)
- MCP Safety (3 rules)
- Agent Config (2 rules)
- Environment & CI (2 rules)
- Testing & Autonomy (2 rules)
- Governance (3 rules)

## Proposed folder structure

```
docs/product/
├── docs.json
├── introduction.mdx
├── quickstart.mdx
│
├── how-to/
│   ├── run-in-github-actions.mdx
│   ├── adopt-in-existing-repo.mdx
│   ├── suppress-a-finding.mdx
│   ├── investigate-mcp-findings.mdx
│   └── use-charter-fix-safely.mdx
│
├── cli/
│   ├── overview.mdx
│   ├── doctor.mdx
│   ├── init.mdx
│   ├── fix.mdx
│   ├── report.mdx
│   ├── explain.mdx
│   ├── suppress.mdx
│   └── version.mdx
│
├── config/
│   └── charter-yaml.mdx
│
├── ci/
│   └── github-action.mdx
│
├── concepts/
│   ├── agent-readiness-model.mdx
│   ├── scoring-and-caps.mdx
│   ├── mcp-safety-model.mdx
│   └── suppression-governance.mdx
│
├── rules/
│   ├── overview.mdx
│   ├── AE-CTX-001.mdx through AE-SUPPRESS-003.mdx (18 pages)
│
├── images/
└── snippets/
```

## Key design decisions

### Decision: Generated rule pages

- **Choice**: generate `rules/AE-*.mdx` from `docs/internal/specs/AE-*.md` + `internal/rules/catalog/catalog.go`.
- **Why**: internal specs are the behavioral authority; hand-authored copies create a high-maintenance duplication problem. Generated pages keep `helpUri` targets complete and drift-resistant.
- **Generator responsibilities**: enumerate all rule specs, parse structured sections, emit stable Mintlify MDX (frontmatter with title/description, severity, category, detection logic, pass/fail examples, edge cases, remediation, scoring impact, related ADRs). Verify every `catalog.HelpURI` target has a matching page. Fail CI/docs validation when coverage is incomplete.
- **Tradeoff**: requires maintaining one generator script. Risk: low — far lower than drift from 18 hand-authored pages.
- **Status**: Confirmed (repo-verified: internal specs exist; catalog exists).

### Decision: Rule page overview

- `rules/overview.mdx` is hand-authored, not generated. Explains the 18-rule launch set, category grouping, score formula, hard caps, and how rule pages relate to `charter explain`.

### Decision: URL architecture

- **Choice**: Mintlify origin on `docs.use-charter.dev`. Root-domain routing added for `https://use-charter.dev/docs/*` and `https://use-charter.dev/rules/*`.
- **Why**: preserves flexibility for Slice 19 to own the root marketing site. Satisfies the hardcoded `helpUri` contract (`https://use-charter.dev/rules/AE-*`). Aligns with Mintlify-supported `/docs` subpath routing.
- **Tradeoff**: requires proxy/rewrite coordination with Slice 19 deployment. Risk: moderate if Slice 18 ships without recording the routing design.
- **Status**: Inferred from Mintlify subpath docs + existing `helpUri` contract. Locked pending deploy implementation.

### Decision: No versioning

- v1 only. Single version. Versioned docs added when a breaking v2 ships.

### Decision: No analytics in content scope

- Analytics decision deferred to deploy-time configuration. `docs.json` does not include an analytics integration by default. If analytics are used, prefer privacy-light options (Fathom or GA4 with limited data retention). Keep this choice explicit in deploy documentation, not hardcoded in the docs project.

## Migration from old state

Mintlify project is new in `docs/product/`. No migration of existing content — the old `web/docs/` foundation (ADR-0022, stashed) is abandoned and should not be resurrected. The directory is clean; `README.md` already documents the expected content.

## Architecture / ownership

- `docs/product/` — the Mintlify project root. New directory, new files.
- `scripts/generate-rule-pages.ts` — Bun TS rule-page generator. New file.
- `moon.yml` — docs tasks updated: `:docs` gains a `mintlify` sub-task for link/coverage validation.
- **No Go changes.** No changes to the CLI binary, `internal/rules/catalog`, `internal/render`, or any cmd package.
- **Avoid**: hand-maintaining 18 rule pages; shipping OpenAPI/API sections prematurely; blocking on the launch website; coupling docs structure to the later marketing site design.

## Testing & verification

- **Local preview**: `npx mintlify dev` from `docs/product/` resolves all navigation and pages.
- **Rule coverage**: a script asserts every `catalog.HelpURI` path corresponds to a real page file.
- **Link integrity**: no broken internal links in any MDX file.
- **Quickstart accuracy**: the quickstart flow is runnable against the current CLI and produces the claimed result.
- **`moon run :docs`**: docs validation passes (link check + rule coverage + generator staleness guard).
- **Honesty audit**: no page claims capabilities that have not shipped or misrepresents launch state.

## Success criteria

- `docs/product/` contains a complete Mintlify project with `docs.json`, navigation, and all page MDX.
- Quickstart, CLI reference, config reference, GitHub Action guide, MCP guidance, suppression guidance, and concept pages exist.
- All 18 rule pages exist at `<root>/rules/AE-*` paths and match the canonical internal specs.
- Rule-page generator validates every catalog `helpUri` resolves to a real page.
- Deployment routing design is documented so Slice 19/20 can complete the URL chain.
- `moon run :docs` remains green.

## References

- `docs/internal/superpowers/plans/2026-06-01-v1-launch-roadmap.md` (Slice 18 row)
- `docs/internal/architecture/charter-architecture-2026.md` (§1.8 command gallery, rule catalog)
- `internal/rules/catalog/catalog.go` (rule metadata + `helpUri` scheme)
- `docs/internal/specs/AE-*.md` (canonical rule specs)
- `docs/product/README.md` (reserved customer-docs home)
- `README.md` (§ "Documentation topology" + "Repo Map")
- Mintlify docs: `docs.json` config, navigation structure, redirects, `/docs` subpath hosting
