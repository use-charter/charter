# pass-secrets-config

Project summary: pass-secrets-config is a fixture repository for Charter secret-rule tests.

- Tech stack: uses Go and Bun with repo automation routed through Moon.
- Edit boundaries: off-limits paths include `.github/workflows/`, `.env*`, and `secrets/`.
- Verify with `moon run :check` before claiming the fixture passes.
- Hooks use `hk.pkl` and product truth lives in `docs/internal/architecture/charter-architecture-2026.md`.
