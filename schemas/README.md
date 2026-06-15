# schemas/

Machine-readable contracts for Charter's inputs and outputs. These are
JSON Schema (draft 2020-12) and are part of the public product contract — they
change only with a corresponding ADR or RFC, and stay stable within a major
version.

| Schema | `$id` | Describes |
|--------|-------|-----------|
| `charter-config.schema.json` | `https://use-charter.dev/schemas/charter-config.schema.json` | The `charter.yaml` config file — `mcp` (trusted remotes) and `policy` (profile, threshold). |
| `doctor-result.schema.json` | `https://use-charter.dev/schemas/doctor-result.schema.json` | The `charter doctor --format json` result — `repo_path`, `threshold`, `passed`, `findings[]`, `suppressed[]`, `summary`, `score`, `categories[]`. |

## Using them

- **Editor validation** — point your YAML/JSON tooling at the `$id` (or this file) to get completion and validation for `charter.yaml`.
- **Consuming output** — the doctor-result schema is the contract for scripts and integrations parsing JSON output; validate against it rather than scraping fields.

## Rules

- A schema change is a contract change: link the ADR or RFC that authorizes it.
- Keep schemas in sync with the Go types they describe (`internal/config`, `internal/doctor`, `internal/render/json`) — drift is a bug.
- The `$id` URLs are permanent; don't rename or move a published schema without a redirect.
