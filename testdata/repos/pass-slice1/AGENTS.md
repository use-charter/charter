# pass-slice1

Project summary: pass-slice1 is a fixture repository for Charter doctor tests.

- Tech stack: uses Go and Bun with repo automation routed through Moon.
- Edit boundaries: off-limits paths include `.github/workflows/`, `.env*`, and `secrets/`.
- Verify with `moon run :check` before claiming the fixture passes.
- Hooks use `hk.pkl` and product truth lives in `docs/internal/architecture/charter-architecture-2026.md`.
