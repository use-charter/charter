#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
  echo "GITHUB_OUTPUT is required" >&2
  exit 2
fi

base_sha="${BASE_SHA:-}"
head_sha="${HEAD_SHA:-}"

emit() {
  printf '%s=%s\n' "$1" "$2" >>"$GITHUB_OUTPUT"
}

if [[ -z "$base_sha" || -z "$head_sha" || "$base_sha" =~ ^0+$ ]]; then
  emit full_check true
  emit docs_only false
  emit actions_security true
  exit 0
fi

mapfile -t changed_files < <(git diff --name-only "$base_sha" "$head_sha")

if [[ ${#changed_files[@]} -eq 0 ]]; then
  emit full_check false
  emit docs_only false
  emit actions_security false
  exit 0
fi

docs_changed=false
non_docs_changed=false
workflow_changed=false

for file in "${changed_files[@]}"; do
  case "$file" in
    .github/workflows/*|.github/action.yml|.github/instructions/workflows.instructions.md)
      workflow_changed=true
      non_docs_changed=true
      ;;
    *.md|docs/*)
      docs_changed=true
      ;;
    *)
      non_docs_changed=true
      ;;
  esac
done

if [[ "$non_docs_changed" == true ]]; then
  emit full_check true
  emit docs_only false
else
  emit full_check false
  emit docs_only "$docs_changed"
fi

emit actions_security "$workflow_changed"
