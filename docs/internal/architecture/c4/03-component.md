# C4 Component

Early logical components:
- command entrypoints (`cmd/charter/`)
- config loading (Cobra)
- repo resolution and file inventory (`internal/repository`)
- agent context registry (`internal/agentcontext` — shared by context and secret rules)
- rule evaluation (`internal/rules/` — context, environment, ci, secrets)
- findings model with Location support (`internal/findings` — path, 1-based line)
- scoring with hard caps (`internal/scoring`)
- output rendering (`internal/render/` — text, JSON)
- secret detection and redaction (`internal/secrets`)
- safe-fix planning (M1.4 roadmap)
