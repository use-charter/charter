#!/usr/bin/env bash
set -euo pipefail

# Classifies the changed files in a push/PR into per-area CI lanes so each lane
# runs only when its inputs change.
#   go / web / docs / infra → consumed by .github/workflows/ci.yml
#   actions_security        → consumed by .github/workflows/actions-security.yml
#   security                → consumed by .github/workflows/vuln-scan.yml
#
# Lane mapping:
#   docs/product/*                    → docs (the published Mintlify site only)
#   scripts/*                         → go + docs (run by both lanes' tasks)
#   internal/rules/catalog/*, specs   → go + docs (catalog drives rule docs)
#   toolchain/config (mise, moon, …)  → all lanes
#   agent-context (AGENTS.md, CLAUDE.md, …) → go (scanned by the doctor gate)
# Other prose — non-context root markdown, engineering docs under
# docs/internal/**, demo
# assets, LICENSE — triggers NO lane and NO security scan: it gates no build,
# no published doc, and contains no code or dependencies. The security scan
# (gitleaks + govulncheck + osv-scanner) runs whenever code, dependencies,
# toolchain, or infra change. Unknown paths conservatively trigger everything.

if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
  echo "GITHUB_OUTPUT is required" >&2
  exit 2
fi

base_sha="${BASE_SHA:-}"
head_sha="${HEAD_SHA:-}"

emit() {
  printf '%s=%s\n' "$1" "$2" >>"$GITHUB_OUTPUT"
}

# emit_all <lanes> <actions_security> <security>
emit_all() {
  emit go "$1"
  emit web "$1"
  emit docs "$1"
  emit infra "$1"
  emit actions_security "$2"
  emit security "$3"
}

# No usable base (first push / force push): run everything.
if [[ -z "$base_sha" || -z "$head_sha" || "$base_sha" =~ ^0+$ ]]; then
  emit_all true true true
  exit 0
fi

mapfile -t changed_files < <(git diff --name-only "$base_sha" "$head_sha")

# No file delta: run nothing.
if [[ ${#changed_files[@]} -eq 0 ]]; then
  emit_all false false false
  exit 0
fi

go=false
web=false
docs=false
infra=false
actions_security=false
security=false

for file in "${changed_files[@]}"; do
  case "$file" in
    .github/workflows/*|.github/action.yml|.github/instructions/*)
      actions_security=true
      ;;
    # Toolchain / workspace config affects every lane and the security gate.
    moon.yml|.moon/*|mise.toml|mise.lock|package.json|bun.lock|tsconfig.json|biome.json|.gitleaks.toml|osv-scanner.toml|charter.yaml)
      go=true; web=true; docs=true; infra=true; security=true
      ;;
    # Shared TS scripts drive both Go Moon tasks and the doc generators.
    scripts/*)
      go=true; docs=true; security=true
      ;;
    # The rule catalog and rule specs feed both the spec-sync Go test and the
    # generated rule docs.
    internal/rules/catalog/*|docs/internal/specs/AE-*.md)
      go=true; docs=true; security=true
      ;;
    # Only the published product docs trigger the docs validation lane.
    docs/product/*)
      docs=true
      ;;
    web/*)
      web=true; security=true
      ;;
    infra/*)
      infra=true; security=true
      ;;
    cmd/*|internal/*|*.go|go.mod|go.sum)
      go=true; security=true
      ;;
    # Agent-context files are scanned by `charter doctor` (AE-CTX-*/AE-SEC-*),
    # so editing them must re-run the Go lane's self-scan gate — otherwise an
    # over-budget or secret-bearing context file slips past CI.
    AGENTS.md|CLAUDE.md|GEMINI.md|.github/copilot-instructions.md|.cursor/rules/*)
      go=true; security=true
      ;;
    # Pure prose / non-published docs / assets — no lane, no security scan.
    *.md|*.mdx|docs/*|LICENSE|*.txt|.gitignore|.gitattributes|.editorconfig)
      :
      ;;
    *)
      go=true; web=true; docs=true; infra=true; security=true
      ;;
  esac
done

emit go "$go"
emit web "$web"
emit docs "$docs"
emit infra "$infra"
emit actions_security "$actions_security"
emit security "$security"
