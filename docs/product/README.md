# Product Docs

Mintlify documentation site for Charter. Deployed at `https://use-charter.dev/docs` and `https://use-charter.dev/rules`.

## Structure

```
docs/product/
  docs.json               — Mintlify site config: navigation, icons, navbar, SEO
  introduction.mdx        — landing page
  installation.mdx        — install paths (brew, binary, go install, source)
  quickstart.mdx          — first scan walkthrough
  design-philosophy.mdx   — the ten commitments
  changelog.mdx           — release history
  concepts/               — how Charter works (readiness model, scoring, fix engine, MCP, suppression)
  how-to/                 — task-oriented guides
  config/                 — charter.yaml and policy profiles reference
  ci/                     — GitHub Action guide
  cli/                    — command reference (doctor, init, fix, report, explain, suppress, version)
  rules/                  — AE-* rule reference (one page per rule)
  images/                 — brand assets (logo-light.svg, logo-dark.svg, favicon.svg)
  DEPLOY.md               — Cloudflare Worker routing + Mintlify custom domain setup
```

## Local preview

```bash
npx mintlify dev
```

Run from `docs/product/`. Requires Node.

## Rule pages

Rule pages (`rules/AE-*.mdx`) are **bootstrapped** by `scripts/generate-rule-pages.ts` and then **hand-maintained**. Do not run the generator against existing rule pages without intent — it overwrites.

To add a new rule page (when a new AE-* rule is added to the engine):

```bash
# 1. Add the rule spec to the rule spec in docs/internal/specs/
# 2. Add the rule to internal/rules/catalog/catalog.go
# 3. Generate the initial page:
bun scripts/generate-rule-pages.ts
# 4. Edit docs/product/rules/AE-XXX-NNN.mdx — replace generated prose with
#    customer-facing content (Why, What triggers it, Examples, How to fix)
# 5. Add the rule ID to the correct group in docs.json
```

The generator check (`bun scripts/generate-rule-pages.ts --check`) validates that all catalog IDs have a rule page with the correct title and CLI section. It does **not** enforce content equality.

## Navigation

Navigation is defined in `docs.json` under `navigation.tabs`. Three tabs:

- **Docs** — uses `anchors` (icon-forward zones): Getting Started, How Charter Works, Guides, Configuration, CI & GitHub Action, Design Philosophy, Changelog
- **CLI Reference** — uses `groups` with icons: Overview, Commands
- **Rules** — uses `groups` with icons: one group per category

Icons use the Lucide library (`"icons": { "library": "lucide" }`). See `tasks/plan.md` for the full icon map.

## Adding a page

1. Create the `.mdx` file in the appropriate directory
2. Add the page path to `docs.json` under the correct anchor or group
3. Run `moon run :docs-product` — validates all nav entries resolve to real files

## URL contracts

These paths are permanent and must not change — SARIF `helpUri`s point at them:

```
/rules/AE-CTX-001   /rules/AE-CTX-002   /rules/AE-CTX-004   /rules/AE-CTX-006
/rules/AE-SEC-001   /rules/AE-SEC-002
/rules/AE-MCP-001   /rules/AE-MCP-002   /rules/AE-MCP-003
/rules/AE-CC-001    /rules/AE-CC-002
/rules/AE-ENV-001   /rules/AE-CI-002
/rules/AE-TEST-001  /rules/AE-AUTO-001
/rules/AE-SUPPRESS-001  /rules/AE-SUPPRESS-002  /rules/AE-SUPPRESS-003
```

If a page path must change, add a redirect in `docs.json` under `"redirects"`.

## Deployment

See `DEPLOY.md` for the full Cloudflare Worker + Mintlify custom domain setup.

Validation gate: `moon run :docs` (runs `docs-html`, `docs-product-rules`, `docs-product`, `docs` in sequence).
