#!/usr/bin/env bash
# Run immediately after checkout, before setup-go or any other Go consumer.
# GITHUB_ENV applies to subsequent steps, including action cache hooks.
set -euo pipefail

: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${GITHUB_ENV:?GITHUB_ENV is required}"
state="$(mktemp -d "$RUNNER_TEMP/curator-go.XXXXXXXX")"
mkdir -p "$state/build" "$state/mod" "$state/path" "$state/lint"
{
  printf 'GOCACHE=%s/build\n' "$state"
  printf 'GOMODCACHE=%s/mod\n' "$state"
  printf 'GOPATH=%s/path\n' "$state"
  printf 'GOLANGCI_LINT_CACHE=%s/lint\n' "$state"
} >>"$GITHUB_ENV"
