---
applyTo: "cmd/**/*.go,internal/**/*.go"
---

# Go Instructions

- Keep one root module: `go.use-charter.dev/charter`.
- No `go.work` or extra modules.
- Prefer `cmd/` for binaries and `internal/` for non-public code.
- Defer public Go packages until a stable external API is proven.
- Before changing tool, SDK, or API usage: inspect `go.mod`, then fetch latest official docs.
- Avoid catch-all utility packages.
- Keep changes small and verification explicit.
