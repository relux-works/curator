#!/usr/bin/env bash
# pnpm pin guard: the workflow's PNPM_PIN must name exactly the release
# internal/pnpmsource admits, or the lane stops before installing anything.
#
# The Go constant internal/pnpmsource.SupportedPNPMVersion is the single
# source of truth; the workflow-level `env: PNPM_PIN` is a copy of it, kept
# as a copy (rather than extracted at install time) so the installed release
# is visible in the workflow and in every job log. This script is what keeps
# the copy honest: a Go-side bump that forgets the workflow fails the lane
# here, instead of installing the old pnpm and skipping (or failing) the
# real-pnpm integration tests downstream.
#
# A read failure is never a match: an unset pin, an unreadable source file,
# or a source file with no declaration all fail closed.
#
# Usage:
#   PNPM_PIN=10.33.0 bash .github/ci/pnpm-pin-guard.sh
#
# Self-test override:
#   CI_PNPM_GO_SOURCE  Go source to read the constant from
#                      (default: internal/pnpmsource/errors.go)

set -u

fail() { echo "pnpm-pin: $*" >&2; exit 1; }

[ -n "${PNPM_PIN:-}" ] || fail 'PNPM_PIN is not set; the workflow must export the pinned pnpm release'

src="${CI_PNPM_GO_SOURCE:-internal/pnpmsource/errors.go}"
[ -r "$src" ] || fail "cannot read $src"

line="$(grep -E 'SupportedPNPMVersion[[:space:]]*=[[:space:]]*"' "$src" | head -n 1)"
[ -n "$line" ] || fail "no SupportedPNPMVersion declaration in $src"
go_pin="$(printf '%s' "$line" | sed -n 's/.*SupportedPNPMVersion[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p')"
[ -n "$go_pin" ] || fail "cannot parse SupportedPNPMVersion from $src"

[ "$PNPM_PIN" = "$go_pin" ] || fail "workflow PNPM_PIN=$PNPM_PIN disagrees with internal/pnpmsource SupportedPNPMVersion=$go_pin"

printf 'pnpm-pin: workflow and Go agree on pnpm %s\n' "$PNPM_PIN"
