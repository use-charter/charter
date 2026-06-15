# Go-Public Runbook

Step-by-step companion to [`launch-checklist.md`](launch-checklist.md) for the
admin/external items. Unambiguous: exact paths, fields, and Charter-specific
values. Sourced from current docs (links per section, verified 2026-06-12).

**Fixed facts for this project**
- Repo: `use-charter/charter` (org `use-charter`)
- Cloudflare account: id `<cloudflare-account-id>` (log in as the project account owner)
- Zone: `use-charter.dev` — id `<cloudflare-zone-id>`
- Pages project name: `charter-landing` (matches `web/wrangler.toml`)

**Suggested order:** C (Resend) → B (Cloudflare) → A (GitHub). A's CodeQL step
needs the repo public; the rest of A can be done anytime.

---

## Batch C — Resend sending domain

Docs: https://resend.com/docs/dashboard/domains/introduction · Cloudflare DNS
guide: https://resend.com/docs/knowledge-base/cloudflare

**Status: DONE.** The **apex** `use-charter.dev` is verified in Resend
(provider Cloudflare, region us-east-1, "ready to send"). The footer form sends
from `updates@use-charter.dev` (`web/functions/api/waitlist.ts`), which the
verified apex covers — **no code change**. Steps 1–4 below are kept for
reference / re-verification only.

1. **Add the domain.** https://resend.com/domains → **Add Domain** → enter
   `use-charter.dev` → choose region → **Add**.
2. **Copy the generated records.** Resend shows an **MX** (bounce feedback), an
   **SPF** `TXT`, a **DKIM** `TXT`, and a **DMARC** record.
3. **Add them to Cloudflare DNS.** Dashboard → `use-charter.dev` → **DNS** →
   **Records** → **Add record** for each. Set **Proxy status = DNS only** (grey
   cloud) on every mail record.
   - **Cloudflare gotcha:** Cloudflare auto-appends the zone to the record name.
     If Resend says the name is `resend._domainkey.use-charter.dev`, enter only
     `resend._domainkey` in Cloudflare (do not type the full domain twice).
4. **Verify.** Back in Resend → the domain → **Verify DNS Records**. Status goes
   `pending` → `verified`.

> **Tracking metrics (optional, skip):** the domain page's "Enable tracking
> metrics" is the *only* place a subdomain appears — a custom tracking subdomain
> for click/open tracking. Charter's signup form needs none of it. Leave it off.

5. **Make sure `WAITLIST_TO` is a real inbox.** Signup notices are sent *to*
   `WAITLIST_TO` (you set this in Batch B). If you use an `@use-charter.dev`
   address, add a Cloudflare **Email Routing** rule (zone → **Email** → Routing)
   forwarding it to your real mailbox, or set `WAITLIST_TO` to a mailbox you
   already read.

---

## Batch B — Cloudflare live deploy

The landing site is a **Pages** project (Git integration). The two workers in
`infra/` deploy from CI once enabled. The apex `use-charter.dev` is served by
the `charter-router` worker (it proxies to the Pages `*.pages.dev` origin), so
**do not** attach `use-charter.dev` as a Pages custom domain.

### B1 — Create the Pages project

Docs: https://developers.cloudflare.com/pages/get-started/git-integration/

1. Dashboard → **Workers & Pages** → **Create application** → **Pages** →
   **Connect to Git**.
2. **Install & Authorize** the Cloudflare GitHub app for the `use-charter` org →
   select repository `use-charter/charter` → **Begin setup**.
3. Set build settings exactly:
   - **Project name:** `charter-landing`
   - **Production branch:** `main`
   - **Framework preset:** None
   - **Build command:** `bun run build`
   - **Build output directory:** `dist`
   - **Root directory (advanced):** `web`
4. **Save and Deploy.** First build runs; note the assigned hostname
   `charter-landing.pages.dev` — you need it in B4.
5. **Scope build watch paths.** Settings → Builds → **Build watch paths** →
   Include paths: `web/*` (the site lives under `web/`). Without this the
   project's `path_includes` defaults to `*` and Cloudflare rebuilds on every
   push — including README/docs-only commits. Equivalent API call:

   ```
   PATCH /accounts/{account_id}/pages/projects/charter-landing
   { "source": { "type": "github", "config": { "path_includes": ["web/*"] } } }
   ```

### B2 — Pages environment variables + secret

In the **Pages project** → **Settings** → **Variables and secrets** (Production):
- Add **secret** `RESEND_API_KEY` = your Resend key (`re_…`). Use the **Encrypt**
  option so it is stored as a secret, not plaintext.
- Add **variable** `WAITLIST_TO` = the address signups are emailed to.
- **Re-deploy** (Deployments → … → Retry/redeploy) so the Function picks them up.

### B3 — Repo secrets + enable the worker deploy

Create a scoped Cloudflare API token: **My Profile** (top-right) → **API Tokens**
→ **Create Token** → **Create Custom Token**. Permissions:
- **Account** → **Workers Scripts** → **Edit**
- **Zone** → **Workers Routes** → **Edit** (zone: `use-charter.dev`)
- **Zone** → **DNS** → **Edit** (zone: `use-charter.dev`) — lets the `go-vanity`
  `custom_domain` route auto-create its DNS.

Then add to the repo (`gh` from the repo root, or Settings → Secrets and
variables → Actions):

```bash
gh secret   set CLOUDFLARE_API_TOKEN   --body '<the token>'
gh secret   set CLOUDFLARE_ACCOUNT_ID  --body '<cloudflare-account-id>'
gh variable set DEPLOY_WORKERS         --body 'true'   # un-gates deploy-workers.yml
```

### B4 — Deploy the workers + wire the router

Trigger `deploy-workers.yml` (Actions tab → **Deploy Workers** → **Run
workflow**, or push any `infra/**` change). It deploys `charter-router` and
`charter-go-vanity`. Then set the router's vars: Dashboard → **Workers & Pages**
→ `charter-router` → **Settings** → **Variables and secrets**:
- `MINTLIFY_ORIGIN` = `tashfiq.mintlify.app` (default; set explicitly)
- `LANDING_ORIGIN` = `charter-landing.pages.dev` (from B1)

`charter-go-vanity` needs no vars — its `custom_domain` route auto-provisions
`go.use-charter.dev` + TLS.

### B5 — Apex DNS for the router route

The route `use-charter.dev/*` only binds if a **proxied** record exists at the
apex. Dashboard → `use-charter.dev` → **DNS** → **Add record**:
- Type **AAAA**, Name `@`, IPv6 `100::` (discard prefix), **Proxy status =
  Proxied** (orange cloud).

(Any proxied apex record works; `100::`/`192.0.2.1` are standard placeholders
because the worker, not an origin server, answers.)

### B6 — Verify

```bash
curl -sI https://use-charter.dev/                         # 200 text/html (landing)
curl -sI https://use-charter.dev/docs                      # 30x → Mintlify
curl -s  "https://go.use-charter.dev/charter?go-get=1" | grep go-import
go install go.use-charter.dev/charter/cmd/charter@latest   # resolves
```

Full deploy reference: [`docs/product/DEPLOY.md`](../../product/DEPLOY.md). Rationale: ADR-0026.

---

## Batch A — GitHub repo + org hardening

### A1 — Make the repo public (gates the rest)

Settings → **General** → bottom **Danger Zone** → **Change repository
visibility** → **Make public** → confirm. (Advanced Security / CodeQL upload and
free Scorecard need this.)

### A2 — Enable CodeQL

After the repo is public, set the variable that un-gates `codeql.yml`:

```bash
gh variable set ENABLE_CODEQL --body 'true'
```

Push a Go change (or Actions → **CodeQL** → **Run workflow**) and confirm it
uploads to **Security → Code scanning**.

### A3 — Branch protection ruleset

Docs: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository

Settings → **Rules** → **Rulesets** → **New ruleset** → **New branch ruleset**:
1. **Ruleset name:** `main protection`. **Enforcement status:** **Active**.
2. **Target branches** → **Add target** → **Include default branch**.
3. Enable, under **Branch protections / rules**:
   - **Require a pull request before merging** (≥ 1 approval is optional for a
     solo maintainer; keep the rule on).
   - **Require status checks to pass** → **Add checks** and select (type to
     filter; the list populates from recent runs): `Report CI status`,
     `Report workflow security status`, `Run security gate`, and — once A2 is
     green — `Analyze Go` (CodeQL). Tick **Require branches to be up to date**.
   - **Block force pushes**.
4. **Create**.

### A4 — Private vulnerability reporting

Docs: https://docs.github.com/en/code-security/security-advisories/working-with-repository-security-advisories/configuring-private-vulnerability-reporting-for-a-repository

Settings → **Advanced Security** → **Private vulnerability reporting** →
**Enable**. CLI: `gh api -X PUT repos/use-charter/charter/private-vulnerability-reporting`.

### A5 — Discussions

Settings → **General** → **Features** → tick **Discussions** → **Set up
discussions**; create a **Q&A** category (the issue-template `config.yml` links
to it). CLI: `gh api -X PATCH repos/use-charter/charter -F has_discussions=true`.

### A6 — Require org 2FA

Docs: https://docs.github.com/en/organizations/keeping-your-organization-secure/managing-two-factor-authentication-for-your-organization/requiring-two-factor-authentication-in-your-organization

Prerequisite: you (owner) already have 2FA on. Org → **Settings** →
**Authentication security** → tick **Require two-factor authentication for
everyone in your organization** → **Save** → **Confirm**. (Members/collaborators
without 2FA lose access until they enable it.)

### A7 — Verify the org domain (CF-10)

Docs: https://docs.github.com/en/organizations/managing-organization-settings/verifying-or-approving-a-domain-for-your-organization

Org → **Settings** → **Verified and approved domains** → **Add a domain** →
`use-charter.dev` → GitHub shows a **TXT** record (name like
`_github-challenge-use-charter-org.use-charter.dev`, with a code value). Add it in
Cloudflare DNS (**DNS only**), wait for propagation, then **Continue verifying**
→ **Verify**. Confirm with:

```bash
dig _github-challenge-use-charter-org.use-charter.dev +short TXT
```

---

## Done when

Every `[ ]` in [`launch-checklist.md`](launch-checklist.md) §4–§7 that maps to
these batches is checked, the verification commands above pass, and the Security
tab shows CodeQL + Scorecard results. Then proceed to the `v0.9.0-rc` cut
(Slice 22).
