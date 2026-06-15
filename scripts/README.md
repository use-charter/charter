# scripts/

Build, test, and documentation automation. These are **Bun + TypeScript**
(plus one shell script) invoked through Moon tasks — they wrap the Go toolchain,
generate docs assets, and gate CI. They are not part of the shipped binary.

Run them via their Moon tasks (e.g. `moon run :build`, `moon run :docs`) rather
than directly, so the toolchain is pinned by mise.

## Build & test

| Script | Purpose |
|--------|---------|
| `go-build.ts` | `go build -o dist/charter ./cmd/charter`. |
| `go-test.ts` | `go test ./...` with `-race` (when CGO is enabled). |
| `go-perf.ts` | The perf gate: `go test -tags=perf -run TestDoctorPerformance` (50k-file budget). |
| `full-test.ts` | Automated command tour — runs every CLI command against generated fixtures and the `:perf`/`:check` gates. Verifies behavior, not styling. |
| `visual-tour.ts` | Manual counterpart: runs commands so a human can eyeball the styled TTY/TUI/HTML output. |

## Documentation generation

| Script | Purpose |
|--------|---------|
| `generate-doc-html.ts` | Renders internal Markdown (architecture, audit) to deterministic self-contained HTML mirrors. `--check` fails on drift. |
| `generate-rule-pages.ts` | Bootstraps `docs/product/rules/AE-*.mdx` from rule specs + the catalog. `--check` validates structure only — **pages are hand-maintained after generation; do not regenerate in place.** |
| `generate-report-fonts.ts` | Builds the base64-embedded `@font-face` CSS for the offline HTML report (no remote fonts at view time). |
| `generate-screenshots.ts` | Renders styled-output screenshots (Playwright → WebP) for the product docs. |
| `validate-product-docs.ts` | Asserts every page in `docs/product/docs.json` resolves to a real `.mdx`. |
| `validate-action.ts` | Asserts the composite GitHub Action is well-formed and every `uses:` is pinned to a full 40-char commit SHA. |
| `scan-doc-todos.ts` | Fails if any unresolved doc or spec TODO markers remain in Markdown. |

## Setup & helpers

| Script | Purpose |
|--------|---------|
| `install-hooks.sh` | Installs git pre-commit / commit-msg / pre-push hooks via `hk` (handles older-Git PATH shims). |
| `check-file-exists.ts` | Exit-0/1 file gate for conditional task logic. |
| `noop.ts` | Intentional no-op for optional/skipped tasks. |
| `lib/` | Shared spawn + fixture utilities (`process.ts`, `full-test.ts`). |

## `github-actions/`

| File | Purpose |
|------|---------|
| `classify-changes.sh` | Maps changed files to CI lanes (go / web / docs / infra / security) so `ci.yml` only runs what's affected. Unknown paths trigger everything (conservative). |
| `launch-signals.ts` | Weekly GitHub-search monitor for external Charter adoption; maintains a tracking issue and posts new signals. |

## Conventions

- **Deterministic output.** Generated files (HTML mirrors, fonts CSS) carry no timestamps; dependency versions are pinned in `bun.lock`.
- **Offline at runtime.** Anything the binary or HTML report ships is self-contained — scripts may fetch at build time (e.g. fonts), never at view time.
- **`--check` everywhere.** Generators double as CI validators so docs can't silently drift from source.
