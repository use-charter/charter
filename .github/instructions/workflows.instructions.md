---
applyTo: ".github/workflows/*.yml"
---

# Workflow Instructions

- Pin third-party actions before activation.
- Keep permissions minimal.
- Avoid shell injection patterns and untrusted interpolation in `run:` steps.
- Route verification through `moon run` golden-path commands.
- Keep workflow names stable for rulesets and required checks.
