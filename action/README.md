# Charter Action

Scan a repository for AI-agent readiness and surface the findings in GitHub Code Scanning. The action downloads the signed `charter` release binary, verifies it (cosign keyless + sha256) against the release identity, runs `charter doctor --format sarif`, and uploads the resulting SARIF report. The step passes, warns, or fails your build based on the score gate.

## Usage

```yaml
permissions:
  contents: read
  security-events: write   # required for SARIF upload
steps:
  - uses: actions/checkout@<sha>   # pin to a SHA
  - uses: use-charter/charter-action@v1
    with:
      version: v1.0.0
      threshold: "80"
```

## Inputs

| Input | Default | Description |
| --- | --- | --- |
| `version` | `latest` | Charter release to use (e.g. `v1.0.0`), or `latest`. |
| `path` | `.` | Repository path to scan. |
| `threshold` | `""` | Minimum passing score (0-100). Empty defers to `charter.yaml` policy / default 80. |
| `format` | `sarif` | Output format: `text`, `json`, `markdown`, or `sarif`. |
| `sarif-file` | `charter.sarif` | Path to write the SARIF report. |
| `upload` | `true` | Upload SARIF to GitHub Code Scanning (requires `security-events: write`). |
| `verify` | `true` | Verify the binary with cosign + sha256 before running. |
| `fail-below` | `true` | Fail the step when the score is below the threshold. |
| `category` | `""` | Optional Code Scanning category for the SARIF results. |

## Outputs

| Output | Description |
| --- | --- |
| `exit-code` | `charter doctor` exit code (`0` pass, `1` below threshold, `2` error). |
| `sarif-file` | Path to the written SARIF report. |

## Permissions

```yaml
permissions:
  contents: read
  security-events: write   # required for SARIF upload
```

`security-events: write` is only needed when `upload` is `true` (the default). With `upload: false` you can drop it and keep `contents: read`.

## Verification

By default (`verify: true`) the action performs full supply-chain verification before the binary ever runs:

- Installs `cosign` and verifies the release `checksums.txt` against its keyless signing bundle (`checksums.txt.sigstore.json`), pinned to the Charter release workflow identity and the GitHub Actions OIDC issuer.
- Confirms the downloaded archive's sha256 matches the verified `checksums.txt` entry.

Escape hatches:

- `verify: false` — skip cosign + sha256 verification (not recommended; only for air-gapped or pre-release testing).
- `upload: false` — run the scan and apply the gate without uploading SARIF to Code Scanning (drops the `security-events: write` requirement).

## Gate behavior

The step exits with the same semantics as `charter doctor`:

- exit `0` — score at or above threshold (pass).
- exit `1` — score below threshold; the step fails when `fail-below: true` (default), otherwise it is reported but does not fail the build.
- exit `2` — scan error; the step always fails.

## Development

This action is developed here in the Charter monorepo under `action/`, and is published to the standalone [`use-charter/charter-action`](https://github.com/use-charter/charter-action) repository at release (consumers reference `use-charter/charter-action@v1`).
