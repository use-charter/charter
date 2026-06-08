# Contributing

## Workflow

1. Read `AGENTS.md`.
2. Load only the context files needed for the task.
3. Check local manifests and latest official docs before code.
4. **Grounding & grill gate (required before executing any plan, slice, or multi-step change):** ground the plan against the latest official docs, industry standards, and current best practices; then critically self-assess it — challenge each design decision, self-answer with evidence (explore the codebase or fetch docs rather than guessing), and surface genuine forks; harden the plan with the findings. Only start coding once the plan is verifiably top-tier. Capture confirmed/changed/deferred decisions in the ADR, design spec, and plan.
5. Make the smallest viable change.
6. Run explicit verification commands.
7. Update ADRs, RFCs, specs, docs, and evals when behavior changes.

## Change Contracts

- Commits: Conventional Commits
- Commit message style:
  - subject: `<type>[optional scope]: <specific summary>`
  - add a body when the why is not obvious from the diff
  - for grouped hardening/review work, summarize the batch in the subject and use bullets in the body for the concrete fixes/alignment points
  - keep the subject descriptive, not generic (`fix: harden release and review surfaces`, not `fix: updates`)
- Versioning: SemVer
- License: Apache-2.0
- Contributions: DCO sign-off required on every commit (`Signed-off-by:` trailer)
- Cross-cutting or risky work: RFC first
- Irreversible architecture decisions: ADR first
- New API surfaces: contract-first
- Non-trivial agent workflows or prompts: eval-driven

## Developer Certificate of Origin

This repository uses the Developer Certificate of Origin (DCO) instead of a contributor license agreement.

By contributing, you certify the Developer Certificate of Origin 1.1 at <https://developercertificate.org/>.

Every commit must include a `Signed-off-by:` trailer matching the author identity. Example:

```text
Signed-off-by: Your Name <you@example.com>
```

Use:

```bash
git commit -s
```

CLA is deferred. If contribution volume, governance needs, or enterprise requirements change, the project may introduce a CLA later with an explicit policy update.

Trigger rules:

- Add or update an ADR when changing package boundaries, command ownership, trust model, config loading, release posture, or other hard-to-reverse architecture decisions
- Add or update an RFC when changing cross-cutting behavior, introducing a new subsystem, or accepting migration/compatibility risk
- Add or update specs when command behavior, rule semantics, fix behavior, suppressions, or error contracts change
- Add or update evals when prompt, workflow, or agent-facing behavior changes in a non-trivial way
- Add or update API contracts before implementing new HTTP or machine-readable interfaces

## Pull Requests

- Link task, ADR, or RFC when applicable
- Include verification commands and results
- Call out risks, generated code, and docs/spec changes
- Keep diffs surgical unless the task explicitly needs breadth
- Route verification through the root Moon task family so local and CI checks stay consistent
- Confirm commits in the PR carry valid DCO sign-offs

## Review Culture

- Agents are collaborators, not authorities
- Repo evidence beats memory
- Fresh docs beat assumptions
- Verification beats confidence
