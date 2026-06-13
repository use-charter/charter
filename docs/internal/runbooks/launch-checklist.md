# v1.0 Launch Checklist

The one-time gate from "code-complete" to a public `v1.0.0` tag. Every launch
requirement lives here with its status and owner. This is **Slice 20** (assemble
the gate); **Slice 21** closes the open items; **Slice 22** is the end-to-end RC
dry-run and sign-off. The recurring release procedure stays in
[`release.md`](release.md).

**Legend** — `[x]` done · `[ ]` open. **Owner:** `code` (in-repo, an engineer
closes it) · `admin` (GitHub/org dashboard or `gh`) · `external` (a third party
or live infra) · `decision` (a founder call).

Definition of done for the gate: every box below is `[x]`, the RC dry-run
(Slice 22) is green, and ADR-0010 is `CLEARED`.

---

## 1. Release pipeline & distribution

- [x] Signed release pipeline built — GoReleaser v2 + cosign keyless + SLSA L3 provenance + SPDX SBOM (ADR-0016/0018). `code`
- [x] `moon run :release-check` / `:release-snapshot` green offline. `code`
- [x] `v0.9.0-rc` dry-run (2026-06-13) verified the pipeline: 6 OS/arch binaries, `checksums.txt`, cosign bundle (`checksums.txt.sigstore.json`), 6 SBOMs, GitHub prerelease, Homebrew tap correctly skipped on the prerelease. **SLSA L3 halts on a private repo by design** ("keep the repository name out of the public transparency log") — auto-resolves once the repo is public for `v1.0.0`; do **not** set `private-repository: true`. Notes are header-only (`release.header`/`footer` + `changelog.disable: true`). Release **tag is GPG-signed** (`tag.gpgsign=true`, key on GitHub) → GitHub shows **Verified**; cut `v1.0.0` from a clone with `tag.gpgsign` on. `external`
- [x] `use-charter/homebrew-tap` **public** + README pushed; `HOMEBREW_TAP_TOKEN` secret set on charter. Pre-staged — the cask auto-publishes on `v1.0.0` (`skip_upload: auto` skipped it on `-rc`). `admin`
- [ ] After `v1.0.0`: confirm `Casks/charter.rb` landed + `brew install use-charter/tap/charter` works (macOS). Verify the tap token isn't expired. `external`
- [ ] `go install go.use-charter.dev/charter/cmd/charter@latest` resolves once the vanity worker is live (CF-4 — see §4). `external`

## 2. CLI product surface

- [x] All 7 commands implemented: `init`, `doctor` (+ `-i` TUI), `explain`, `report` (html/md/json), `fix`, `suppress`, `version`. `code`
- [ ] Smoke-test the matrix on the **released RC binary** (not a local build): each command on a real repo, `--format json|sarif` shapes valid, exit codes correct, `NO_COLOR`/piped output byte-stable. `external` (Slice 22)
- [x] Charter dogfoods to a passing score (`charter doctor --threshold 80` on this repo, enforced in CI). `code`

## 3. GitHub Action

- [ ] Seed `use-charter/charter-action` from `action/`; tag `v1` (+ moving major). `admin`
- [ ] Switch this repo's CI to dogfood `use-charter/charter-action@v1` (replaces the local action path). `code` (after the action repo exists)
- [ ] Verify end-to-end on a sample repo: downloads the signed RC binary, cosign + checksum verify, SARIF uploads to the Security tab, below-threshold gating fails the run. `external` (needs §1 RC binary)

## 4. Web & docs live

- [x] Mintlify docs built + live on the `*.mintlify.dev` subdomain (`MINTLIFY_ORIGIN`). `external`
- [ ] Docs served at `use-charter.dev/docs`,`/cli`,`/rules`,`/changelog` + `/rules/AE-*` `helpUri`s resolve — needs the router + apex flip (CF-9, hard launch dependency). `external`
- [x] Landing site built + live on Cloudflare Pages (`charter-landing.pages.dev`); signup form → Resend → Email Routing verified end-to-end. Apex flip still pending (row below). `external`
- [ ] `go.use-charter.dev` vanity worker deployed (CF-4). `external`
- [ ] Set repo secrets `CLOUDFLARE_API_TOKEN` + `CLOUDFLARE_ACCOUNT_ID`, then repo variable `DEPLOY_WORKERS=true` to enable `deploy-workers.yml`. `admin`
- [x] Verify the Resend sending domain — apex `use-charter.dev` verified (2026-06-12, ready to send). `external`
- [x] Pages vars set — `RESEND_API_KEY` (dashboard secret) + `WAITLIST_TO` (wrangler `[vars]`). `admin`
- [ ] Worker `LANDING_ORIGIN` set to the `*.pages.dev` host; apex has a proxied DNS record. `external`

Full deploy runbook: [`docs/product/DEPLOY.md`](../../product/DEPLOY.md). Topology rationale: ADR-0026.

## 5. Repo & org hardening (public)

- [x] CodeQL workflow (`.github/workflows/codeql.yml`) — built; gated behind `ENABLE_CODEQL=true`. `code`
- [ ] Set `ENABLE_CODEQL=true` once the repo is public (code-scanning upload needs Advanced Security, free on public repos). `admin`
- [x] OSSF Scorecard workflow (`scorecard.yml`). `code`
- [x] Supply-chain gates: govulncheck + osv-scanner + gitleaks + zizmor (pedantic) + actionlint, all green on `main`. `code`
- [ ] Branch protection on `main`: require `Report CI status`, `Report workflow security status`, `Vulnerability Scan`, `CodeQL`, Scorecard; require PRs; no force-push. `admin` (Appendix A)
- [ ] Enable **private vulnerability reporting** — public repos only; enable at go-public. `admin` (Appendix A)
- [x] **Discussions** enabled (Q&A category live; issue-template link resolves). `admin`
- [x] Org: 2FA required + `use-charter.dev` domain **verified** (TXT) (CF-10). `admin`

## 6. Legal & project meta

- [x] `LICENSE` (Apache-2.0). `code`
- [x] `NOTICE`. `code`
- [x] `CHANGELOG.md` + versioning policy (`CONTRIBUTING.md#versioning-policy`). `code`
- [x] `CODE_OF_CONDUCT.md` (Contributor Covenant 2.1). `code`
- [x] `SECURITY.md`, `CONTRIBUTING.md`, issue templates + `config.yml`. `code`
- [x] Cloudflare Email Routing live — `updates@`, `security@`, `conduct@` → maintainer inbox (verified). `external`
- [x] **ADR-0010 trademark resolved → PROCEED as "Charter"** (informed-risk acceptance; verified Stackbilt same-niche collision recorded, with a documented rename trigger). `decision`

## 7. Launch monitoring & assets

- [ ] Alerts for launch signals (architecture §Signals): **Signal 1** organic CI adoption, **Signal 3** unprompted mentions (Google Alert + GitHub search for "charter doctor" / "use-charter"), **Signal 4** community self-help. `admin`
- [x] Demo asset source committed (`docs/internal/demo/charter-demo.tape`, VHS). `code`
- [ ] Render the demo GIF and embed in README / landing. `external` (needs `vhs` + a built binary)

## 8. MCP catalog (CF-12)

- [ ] Refresh the MCP catalog at the release gate: broaden the FP re-validation run, bring advisories/versions current (real CVE IDs only, ADR-0021), seed behind-stable version data beyond `filesystem`. `external` (founder curation)

## 9. Known deferrals (non-blocking)

- macOS binaries are cosign-signed but **not Apple-notarized** (no Apple Developer cert); the Homebrew cask strips the quarantine attribute interim (CF-11). Revisit post-launch.
- Phase 1.5 / v1.1 product backlog: see `carry-forward.md`.

---

## Appendix A — admin steps (execute by hand)

**Branch protection** — Settings → Branches → Add rule for `main`:
require status checks (`Report CI status`, `Report workflow security status`,
`Vulnerability Scan`, `CodeQL`), require a PR before merging, block force-push.
CLI equivalent: `gh api -X PUT repos/use-charter/charter/branches/main/protection`
with the required-checks payload.

**Private vulnerability reporting** — Settings → Code security and analysis →
Private vulnerability reporting → Enable.

**Enable CodeQL** — after the repo is public (Advanced Security is free on
public repos): Settings → Secrets and variables → Actions → Variables → add
`ENABLE_CODEQL=true`. The CodeQL workflow then runs and uploads to the Security
tab; add `CodeQL` to the required status checks once it is green.

**Discussions** — Settings → General → Features → Discussions → Enable; create a
"Q&A" category (the issue-template `config.yml` already links to it).

**Org 2FA** — Org → Settings → Authentication security → Require two-factor
authentication for everyone in the organization.

**Domain verification (CF-10)** — Org → Settings → Verified and approved
domains → add `use-charter.dev` → publish the DNS TXT record Cloudflare shows.

**Cloudflare deploy enablement** — repo Settings → Secrets and variables →
Actions: add secrets `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`; add
variable `DEPLOY_WORKERS=true`. Then merge to `main` (or run `deploy-workers.yml`
manually) to deploy the edge workers.
