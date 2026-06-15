# cmd/

Binary entrypoints. Today there is one: [`charter/`](./charter/) — the Charter CLI.

The command layer is deliberately thin. Each command parses flags, resolves
terminal capabilities, calls into [`internal/`](../internal/), and renders the
result. No scanning, scoring, or fix logic lives here.

## Framework

Cobra commands, wrapped by [Fang](https://github.com/charmbracelet/fang) for the
styled help and version surface. Color and TTY handling are centralized in
`color.go` (`--color auto|always|never`, `--no-color`, `NO_COLOR`), which hands
every command a detected capability set + palette so the `render` layer stays
presentation-agnostic.

## Commands

| Command | What it does | Key flags |
|---------|--------------|-----------|
| `charter doctor` | Scan + score the repo 0–100 with per-category breakdown. The CI gate. | `--path`, `--threshold`, `--quiet`, `--format text\|json\|markdown\|sarif`, `--out`, `--rule`, `-i/--interactive` (TUI) |
| `charter init` | Scaffold missing agent-context files. Never overwrites. | `--profile`, `--agents`, `--dry-run` |
| `charter fix` | Preview and apply unified diffs for fixable findings; backup-then-write, never overwrites. | `--rule`, `--dry-run` |
| `charter explain <RULE>` | Print a rule's name, category, summary, and docs URL. | `--format text\|json` |
| `charter report` | Write a shareable report (HTML/Markdown/JSON). Not a gate — always exits 0 on success. | `--format`, `--out`, `--open` |
| `charter suppress <RULE>` | Record a governed waiver in `.charter-suppress.yml`. | `--reason` (required), `--expires`, `--approver` (required for permanent), `--dry-run` |
| `charter version` | Version + build provenance (commit, date, Go, platform). | `--format`, `--short` |

## Exit codes

The contract is identical in your shell and in CI:

- `0` — pass (score ≥ threshold, or a non-gating command succeeded)
- `1` — below threshold / failure
- `2` — usage error (bad arguments or flags)

Gating applies to `doctor` (text/`--quiet`) — machine-readable formats still emit
their output and report the score. `report` never gates.

## Tests

Each command has unit tests (`*_test.go`) plus integration tests
(`*_integration_test.go`) that exercise the wired command end-to-end. Run them
through `moon run :test` (`go test -race ./...`).
