#!/usr/bin/env bash
# Stages /tmp/acme-api — the demo repository the README GIF scans.
# A realistic small Go service with a few intentional readiness gaps so
# `charter doctor` lands on a PASS (85) with one HIGH finding to explain.
set -euo pipefail

DIR=/tmp/acme-api
rm -rf "$DIR"
mkdir -p "$DIR/.github/workflows" "$DIR/src"
cd "$DIR"
git init -q

cat > AGENTS.md <<'EOF'
# AGENTS.md

acme-api — a Go 1.26 payments service. This file is the contract every AI
coding agent reads before touching the repository.

## Commands

- Install deps: `go mod download`
- Build: `go build ./...`
- Test (verification command): `go test ./...`
- Lint: `gofmt -l . && go vet ./...`

## Architecture

- `src/` holds the HTTP handlers and the payment state machine.
- One Go module, `acme-api`. No generated code is committed.
- Postgres is the only datastore; migrations live in `migrations/`.

## Edit scope

Agents may edit `src/`, tests, and docs. Off-limits paths:

- `.env*`, `secrets/`, signing keys, production infra, generated state.

## Conventions

- Conventional Commits, SemVer tags.
- Every change ships with a test; keep `go test ./...` green.
EOF

cat > go.mod <<'EOF'
module acme-api

go 1.26
EOF

# unpinned MCP server → AE-MCP-001 (HIGH), the finding the demo explains
cat > .mcp.json <<'EOF'
{
  "mcpServers": {
    "git": { "command": "uvx", "args": ["mcp-server-git"] }
  }
}
EOF

cat > src/main_test.go <<'EOF'
package main

import "testing"

func TestPing(t *testing.T) {
	if 1+1 != 2 {
		t.Fatal("math is broken")
	}
}
EOF

cat > .gitignore <<'EOF'
node_modules/
*.charter-session
.charter/
.claude/local/
.cursor/cache/
.env*
.hk/
EOF

cat > .github/workflows/ci.yml <<'EOF'
name: CI
on: [pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: go test ./...
EOF

echo "staged $DIR"
