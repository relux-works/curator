#!/usr/bin/env bash
# Go toolchain identity assertion, run by every Go-consuming CI job immediately
# after actions/setup-go and before any other Go command.
#
# Setting GOTOOLCHAIN/GOENV in the workflow `env:` block states an intent; a
# shim, a wrapper or a later setup-go change can override it. This script reads
# the values back from the toolchain that is actually on PATH, so "which Go
# produced this result" is answerable from the job log alone.
#
# The required version is derived from go.mod rather than duplicated here:
# setup-go resolves the toolchain from `go-version-file: go.mod`, so go.mod is
# the single source of truth and a bump needs one edit, not two. What this
# script asserts is that the resolved toolchain is EXACTLY that version -- not
# merely the same major.minor family.

set -u

fail() { echo "toolchain: $*" >&2; exit 1; }

want="$(awk '/^go[ \t]/{print $2; exit}' go.mod)" || fail 'cannot read go.mod'
[ -n "$want" ] || fail 'go.mod declares no go directive'
required="go$want"

goexe="$(command -v go)"     || fail 'no go on PATH after setup-go'
fmtexe="$(command -v gofmt)" || fail 'no gofmt on PATH after setup-go'

v="$(go version)" || fail 'go version failed'
printf '%s\n' "toolchain: $v"
case "$v" in
	*" $required "*) ;;
	*) fail "expected $required (from go.mod), got: $v" ;;
esac

# Absolute reported root. Under git-bash on windows-latest `go env GOROOT`
# reports C:\hostedtoolcache\... while `command -v` reports
# /c/hostedtoolcache/.../go.exe; MSYS accepts C:/..., so swapping separators
# makes both spellings nameable without weakening the comparison.
r="$(go env GOROOT)" || fail 'go env GOROOT failed'
printf '%s\n' "toolchain: GOROOT=$r"
rp="$(printf '%s' "$r" | tr '\\' '/')"
[ -n "$rp" ] || fail 'go env GOROOT is empty'
case "$rp" in
	[A-Za-z]:/*|/*) ;;
	*) fail "reported GOROOT is not absolute: $r" ;;
esac

# The launcher IS this root's go and the formatter IS this root's gofmt.
# Textual first, then -ef (device+inode), which is what survives /c/... vs
# C:/... and the .exe suffix without accepting a different toolchain.
same() { [ "$1" = "$2" ] || [ "$1" = "$2.exe" ] || [ "$1" -ef "$2" ] || [ "$1" -ef "$2.exe" ]; }
same "$goexe"  "$rp/bin/go"    || fail "launcher $goexe is not $r/bin/go"
same "$fmtexe" "$rp/bin/gofmt" || fail "formatter $fmtexe is not $r/bin/gofmt"

# No implicit toolchain download, and no per-user go env file in the loop.
# Read back, never assume the workflow env: block took effect.
tc="$(go env GOTOOLCHAIN)" || fail 'go env GOTOOLCHAIN failed'
printf '%s\n' "toolchain: GOTOOLCHAIN=$tc"
[ "$tc" = local ] || fail "GOTOOLCHAIN=$tc, not local"

ge="$(go env GOENV)" || fail 'go env GOENV failed'
printf '%s\n' "toolchain: GOENV=$ge"
[ "$ge" = off ] || fail "GOENV=$ge, not off"

echo 'ci-toolchain-identity: ok'
