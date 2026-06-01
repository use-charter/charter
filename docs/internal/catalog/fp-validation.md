# MCP Catalog — False-Positive Validation (T1.6.3)

> **⚑ FOUNDER / LAUNCH GATE.** This validation **blocks the public tag (Slice 17)**.
> Until it is completed with a recorded ≤ 10% false-positive rate, the catalog
> ships as an engine + seed only (carry-forward CF-12).

## Why

The catalog makes `AE-MCP-001`/`AE-MCP-002` re-fire on previously-passing repos.
If it produces noise on legitimate servers, early adopters mute Charter — the
exact "rules too noisy" failure the architecture's validation analysis warns
about. This gate proves the catalog is trustworthy before it reaches users.

The design already bounds the risk (ADR-0021): behind-stable findings are
informational (non-deducting), and comparison is exact-match (a stale catalog
under-reports). This gate validates the **HIGH** signals — deprecated packages
and CVE advisories — plus the AE-MCP-002 host baseline.

## Method

1. Select **5+ real public repos** with committed MCP configs (`.mcp.json`,
   `.cursor/mcp.json`, etc.), spanning vendors and the archived-package set.
2. Build Charter at the candidate catalog version and scan each:
   `charter doctor --path <repo> --format json`.
3. For **every** `AE-MCP-001` and `AE-MCP-002` finding, classify as
   **true positive** or **false positive** and record the reasoning below.
4. Compute `FP rate = false positives / total MCP findings`. Target **≤ 10%**.
5. Fix the catalog (or rule) for any FP, re-run, and re-record until green.

## Results

- Catalog version under test: _<fill: catalog `version`>_
- Date / reviewer: _<fill>_
- Charter build (commit): _<fill>_

| Repo | Finding (rule + server) | Severity | Classification | Notes / action |
|------|-------------------------|----------|----------------|----------------|
| _<repo>_ | _<AE-MCP-00x: server>_ | _<HIGH/info>_ | _<TP/FP>_ | _<why; catalog fix if FP>_ |

- Total MCP findings: _<n>_  ·  False positives: _<n>_  ·  **FP rate: _<%>_**
- Gate: **PASS / FAIL** at ≤ 10%.

## Sign-off

- [ ] FP rate ≤ 10% recorded above.
- [ ] Every FP has a catalog/rule fix landed and re-verified.
- [ ] Version data refreshed to a verified ~20–30 server set (CF-12).
- [ ] Founder sign-off to unblock the public tag (Slice 17).
