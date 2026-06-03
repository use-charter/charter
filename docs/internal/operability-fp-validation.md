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

## Run (2026-06-03) — self-dogfood

Method: `charter doctor --path .` on Charter's own tree after Slice 16 embedded a
single browser asset (`internal/render/html/assets/report.js`) into the binary via
`//go:embed`. Charter carries a `package.json` for its Bun build tooling.

| #  | Repo           | Stack            | Initial               | After fix | Note |
|----|----------------|------------------|-----------------------|-----------|------|
| 11 | charter (self) | Go + embedded JS | **FP** (AE-TEST-001)  | clean     | The lone `//go:embed`'d `report.js` + a build-tooling `package.json` wrongly activated JavaScript/TypeScript as an untested surface; Charter dogfooded to 90 instead of 100 |

- Initial: 1 FP / 1 finding (`AE-TEST-001` HIGH) — same over-eager-activation root class as the Slice 14 FPs, now via an embedded resource.
- After the fix below: **0 FP / 0 findings; Charter scores 100.**

### Fix landed this run

3. **Embedded-asset gate** — a file referenced by a `//go:embed` directive in a
   tracked `.go` file is a bundled resource of the host program, not an
   independent language source surface, so it is excluded from the language
   source/test counts. Patterns resolve relative to the directive's `.go` file
   directory and support exact files, directory subtrees (recursive), and
   `path.Match` globs (with a leading `all:` prefix stripped). `//go:embed` is
   Go-only, so the `.go` scan runs only when a non-Go manifest is present, which
   also keeps the 50k-file perf budget intact. This prevents the general FP
   class: any Go CLI that `go:embed`s a web asset while carrying a `package.json`
   (or similar manifest) for build tooling.

## Sign-off

- [x] FP rate ≤ 10% (0% after fixes) across 10 real repos.
- [x] Both FPs fixed and re-verified clean.
- [x] True positives preserved (unit tests + a synthetic Go-without-tests repo still fire AE-TEST-001/AE-AUTO-001).
- [x] Self-dogfood FP (embedded `report.js`) fixed via the embedded-asset gate; Charter scores 100 with no AE-TEST-001 (2026-06-03).
- [ ] Broaden the run to more polyglot/monorepo layouts over time.
