# Changelog

All notable changes to Charter are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). See
[CONTRIBUTING.md](CONTRIBUTING.md#versioning-policy) for what the version number
covers.

This is the canonical engineering changelog. The customer-facing release notes
live in the product docs (`docs/product/changelog.mdx`).

## [Unreleased]

First public release in preparation (v1.0.0). Highlights of the pre-1.0 work:

### Added
- Offline-first `charter` CLI: `init`, `doctor` (with an interactive `-i` TUI),
  `explain`, `report` (HTML/Markdown/JSON), `fix`, `suppress`, `version`.
- 18 agent-readiness rules across context, secrets, MCP safety, agent config,
  environment/CI, testing/autonomy, and suppression governance.
- SARIF output with rule `helpUri`s, GitHub Action (`action/`), and a signed
  release pipeline (GoReleaser + cosign + SLSA provenance + SPDX SBOM).
- Public docs (Mintlify) and landing site (Astro on Cloudflare Pages), plus a
  `go.use-charter.dev` vanity-import worker.

[Unreleased]: https://github.com/use-charter/charter/commits/main
