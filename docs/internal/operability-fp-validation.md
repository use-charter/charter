# Agent-Operability Rules — False-Positive Validation (Slice 14)

Validation of `AE-TEST-001` and `AE-AUTO-001` against real public repositories,
mirroring the Slice 13 catalog FP gate. Target: ≤ 10% false positives.

## Run (2026-06-02)

Method: `git clone --depth 1` of 10 diverse, well-known public repos spanning
Go, JS/TS, Python, Rust, and packaging-polyglot layouts; scan each with
`charter doctor --format json`; classify every `AE-TEST-001`/`AE-AUTO-001`
finding TP/FP.

| # | Repo | Stack | Initial | After fix | Note |
|---|------|-------|---------|-----------|------|
| 1 | spf13/cobra | Go | clean | clean | TN — has `*_test.go`, `go test` conventional |
| 2 | stretchr/testify | Go | clean | clean | TN |
| 3 | charmbracelet/bubbletea | Go | clean | clean | TN |
| 4 | colinhacks/zod | TS | clean | clean | TN |
| 5 | sindresorhus/ky | TS | **FP** (AE-TEST-001) | clean | Tests live in `test/*.ts` (AVA), not `*.test.ts` — detection missed them |
| 6 | pallets/click | Python | clean | clean | TN |
| 7 | psf/requests | Python | clean | clean | TN |
| 8 | BurntSushi/ripgrep | Rust | **FP** (AE-TEST-001 "Ruby") | clean | A single `pkg/brew/ripgrep-bin.rb` Homebrew formula (no `Gemfile`) wrongly activated Ruby |
| 9 | serde-rs/serde | Rust | clean | clean | TN — inline `#[cfg(test)]` |
| 10 | junegunn/fzf | Go | clean | clean | TN |

- Initial: 2 FP / 2 findings = **100% of findings were FP** (both the same root class: over-eager language activation).
- After the two fixes below: **0 FP / 0 findings across all 10 repos.**

## Fixes landed this run

1. **Manifest gate** — a language is "active" only when its project manifest is
   present (`go.mod`, `package.json`, `Cargo.toml`, `Gemfile`, `pyproject.toml`/…,
   `*.csproj`, `composer.json`) **in addition to** having non-test source outside
   tooling dirs. A stray `pkg/brew/*.rb` Homebrew formula no longer activates Ruby
   in a Rust repo.
2. **Directory-based test detection** — a source file under a `test/`/`tests/`/
   `spec/`/`__tests__/` segment counts as a test for any language, alongside the
   per-file name conventions. Catches AVA/tap/node:test/RSpec layouts (ky's
   `test/*.ts`).

## Sign-off

- [x] FP rate ≤ 10% (0% after fixes) across 10 real repos.
- [x] Both FPs fixed and re-verified clean.
- [x] True positives preserved (unit tests + a synthetic Go-without-tests repo still fire AE-TEST-001/AE-AUTO-001).
- [ ] Broaden the run to more polyglot/monorepo layouts over time.
