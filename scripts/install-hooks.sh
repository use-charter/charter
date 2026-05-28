#!/bin/sh
set -eu

HOOKS_DIR="$(git rev-parse --git-path hooks)"
MISE_BIN="${MISE_BIN:-$(command -v mise 2>/dev/null || printf '%s' "$HOME/.local/bin/mise") }"
MISE_BIN="${MISE_BIN% }"

if [ ! -x "$MISE_BIN" ]; then
  echo "missing mise binary at $MISE_BIN" >&2
  exit 1
fi

if [ ! -f "$HOOKS_DIR/pre-commit" ] || [ ! -f "$HOOKS_DIR/commit-msg" ] || [ ! -f "$HOOKS_DIR/pre-push" ]; then
  "$MISE_BIN" exec -- hk install --mise
fi

# hk uses config-based hooks on Git 2.54+. Older Git falls back to .git/hooks
# shims that call bare `mise`, which fails when ~/.local/bin is not in PATH.
for hook in "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/commit-msg" "$HOOKS_DIR/pre-push"; do
  if [ -f "$hook" ] && grep -q 'exec mise x -- hk run' "$hook"; then
    tmp="$(mktemp)"
    sed 's|exec mise x -- hk run|exec "$HOME/.local/bin/mise" x -- hk run|' "$hook" > "$tmp"
    mv "$tmp" "$hook"
    chmod +x "$hook"
  fi
done
