#!/bin/sh
set -eu

MISE_BIN="${MISE_BIN:-$HOME/.local/bin/mise}"

if [ ! -x "$MISE_BIN" ]; then
  echo "missing mise binary at $MISE_BIN" >&2
  exit 1
fi

if [ ! -f .git/hooks/pre-commit ] || [ ! -f .git/hooks/commit-msg ] || [ ! -f .git/hooks/pre-push ]; then
  "$MISE_BIN" exec -- hk install --mise
fi

# hk uses config-based hooks on Git 2.54+. Older Git falls back to .git/hooks
# shims that call bare `mise`, which fails when ~/.local/bin is not in PATH.
for hook in .git/hooks/pre-commit .git/hooks/commit-msg .git/hooks/pre-push; do
  if [ -f "$hook" ] && grep -q 'exec mise x -- hk run' "$hook"; then
    tmp="$(mktemp)"
    sed 's|exec mise x -- hk run|exec "$HOME/.local/bin/mise" x -- hk run|' "$hook" > "$tmp"
    mv "$tmp" "$hook"
    chmod +x "$hook"
  fi
done
