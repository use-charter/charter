#!/usr/bin/env bash
set -euo pipefail

# Classifies the changed files in a push/PR into per-area CI lanes so each lane
# (go, web, docs, infra) runs only when its inputs change. `actions_security`
# is consumed by .github/workflows/actions-security.yml.
#
# Shared inputs fan out to multiple lanes:
#   scripts/*                         → go + docs (run by both lanes' tasks)
#   internal/rules/catalog/*, specs   → go + docs (catalog drives rule docs)
#   toolchain/config (mise, moon, …)  → all lanes
# Unknown paths conservatively trigger all lanes.

if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
  echo "GITHUB_OUTPUT is required" >&2
  exit 2
fi

base_sha="${BASE_SHA:-}"
head_sha="${HEAD_SHA:-}"

emit() {
  printf '%s=%s\n' "$1" "$2" >>"$GITHUB_OUTPUT"
}

emit_all() {
  emit go "$1"
  emit web "$1"
  emit docs "$1"
  emit infra "$1"
  emit actions_security "$2"
}

# No usable base (first push / force push): run everything.
if [[ -z "$base_sha" || -z "$head_sha" || "$base_sha" =~ ^0+$ ]]; then
  emit_all true true
  exit 0
fi

mapfile -t changed_files < <(git diff --name-only "$base_sha" "$head_sha")

# No file delta: run nothing.
if [[ ${#changed_files[@]} -eq 0 ]]; then
  emit_all false false
  exit 0
fi

go=false
web=false
docs=false
infra=false
actions_security=false

for file in "${changed_files[@]}"; do
  case "$file" in
    .github/workflows/*|.github/action.yml|.github/instructions/*)
      actions_security=true
      ;;
    # Toolchain / workspace config affects every lane.
    moon.yml|.moon/*|mise.toml|mise.lock|package.json|bun.lock|tsconfig.json|biome.json|.gitleaks.toml|osv-scanner.toml|charter.yaml)
      go=true; web=true; docs=true; infra=true
      ;;
    # Shared TS scripts drive both Go Moon tasks and the doc generators.
    scripts/*)
      go=true; docs=true
      ;;
    # The rule catalog and rule specs feed both the spec-sync Go test and the
    # generated rule docs.
    internal/rules/catalog/*|docs/internal/specs/AE-*.md)
      go=true; docs=true
      ;;
    web/*)
      web=true
      ;;
    infra/*)
      infra=true
      ;;
    cmd/*|internal/*|*.go|go.mod|go.sum)
      go=true
      ;;
    docs/*|*.md|*.mdx)
      docs=true
      ;;
    *)
      go=true; web=true; docs=true; infra=true
      ;;
  esac
done

emit go "$go"
emit web "$web"
emit docs "$docs"
emit infra "$infra"
emit actions_security "$actions_security"
