#!/usr/bin/env bash
# No broad suppression.
#
# Security findings may be suppressed narrowly and by name; they may not be
# suppressed wholesale. A suppression that names no rule silences every rule,
# including ones written after the comment -- which is how a finding in new
# security code disappears without anyone deciding that it should.
#
# Lives in a script rather than inline in the workflow so `gate-selftest.sh`
# can prove it rejects what it claims to reject.
#
# Usage:
#   no-broad-suppression.sh [source-dir...]      (default: cmd internal)
#
# Environment:
#   CI_GOLANGCI_CONFIG  golangci config path     (default: .golangci.yml)

set -u

CONFIG="${CI_GOLANGCI_CONFIG:-.golangci.yml}"
[ "$#" -gt 0 ] || set -- cmd internal

rc=0

# A `//nolint` with no linter named, or an explicit `//nolint:all`.
if grep -rInE '//[[:space:]]*nolint([[:space:]]|$|:all)' --include='*.go' "$@"; then
	echo 'lint: bare or all-linter //nolint is not allowed; name the linter and the reason.'
	rc=1
fi
# A `//#nosec` with no gosec rule named.
if grep -rInE '//[[:space:]]*#nosec([[:space:]]*$)' --include='*.go' "$@"; then
	echo 'lint: bare //#nosec is not allowed; name the gosec rule and the reason.'
	rc=1
fi

if [ -f "$CONFIG" ]; then
	# The config may exclude linters for test files only. A rule that exempts a
	# production path is a broad suppression by another name.
	if grep -nE '^[[:space:]]*-[[:space:]]*path:' "$CONFIG" \
	     | grep -vE 'path:[[:space:]]*.*_test\\?\.go'; then
		echo "lint: $CONFIG exclusion rules may only target _test.go paths."
		rc=1
	fi
	if grep -nE 'disable-all|default:[[:space:]]*none' "$CONFIG"; then
		echo "lint: wholesale linter disabling is not allowed in $CONFIG."
		rc=1
	fi

	# The gosec exclusion set is narrow and each entry carries a recorded
	# reason. Pinning it here means new security code cannot quietly acquire a
	# new blanket exemption: adding a rule to the config without adding it to
	# this list, with a reason, fails the gate.
	allowed="${CI_GOSEC_ALLOWED:-G306 G301 G122 G703}"
	excludes="$(awk '
		/^[[:space:]]*excludes:[[:space:]]*$/ { inlist = 1; next }
		inlist && /^[[:space:]]*-[[:space:]]*G[0-9]+[[:space:]]*$/ {
			line = $0; sub(/^[^G]*/, "", line); sub(/[[:space:]]*$/, "", line); print line; next
		}
		inlist && /^[[:space:]]*[A-Za-z_-]+:/ { inlist = 0 }
	' "$CONFIG")"
	for rule in $excludes; do
		case " $allowed " in
			*" $rule "*) ;;
			*)
				echo "lint: $CONFIG excludes gosec rule $rule, which is not in the recorded allowlist ($allowed)."
				echo 'lint: a new blanket gosec exemption needs a recorded reason, not a quiet addition.'
				rc=1
				;;
		esac
	done
fi

if [ "$rc" -eq 0 ]; then echo 'no-broad-suppression: ok'; fi
exit "$rc"
