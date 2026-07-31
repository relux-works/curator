#!/usr/bin/env bash
# Which packages this platform does not execute, and why.
#
# One implementation, used by both `suite-plan.sh` (which decides what to run)
# and `ledger-consistency.sh` (which decides what the ledger must declare).
# Two implementations of this rule would eventually disagree, and the
# disagreement would look like a coverage gap in whichever gate was checked
# second.
#
# The supplied conformance root's own
# `vectors/conformance-claim-v3-qualification.json` is authoritative when it
# exists. `default_excluded_on` in `.github/ci/platform-exclusions.tsv` applies
# only to a root that predates that vector -- the committed released pin is
# such a root.
#
# Usage:
#   excluded-packages.sh <goos> [conformance-root]
#
# Prints one TAB separated line per excluded package:
#   package  assertion_case  source_of_truth
#
# Prints nothing and exits 0 when this platform excludes nothing.
#
# Environment:
#   CI_PLATFORM_EXCLUSIONS  default: .github/ci/platform-exclusions.tsv

set -u

EXCLUSIONS="${CI_PLATFORM_EXCLUSIONS:-.github/ci/platform-exclusions.tsv}"
QUALIFICATION_VECTOR='vectors/conformance-claim-v3-qualification.json'

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
	echo 'excluded-packages: usage: excluded-packages.sh <goos> [conformance-root]' >&2
	exit 2
fi
GOOS_WANT="$1"
ROOT="${2:-}"
[ -f "$EXCLUSIONS" ] || { echo "excluded-packages: exclusion table not found: $EXCLUSIONS" >&2; exit 2; }

QV=''
[ -n "$ROOT" ] && [ -f "$ROOT/$QUALIFICATION_VECTOR" ] && QV="$ROOT/$QUALIFICATION_VECTOR"

vector_excludes_this_goos=0
if [ -n "$QV" ]; then
	# The vector spells macOS `macos`; the other two names match their GOOS.
	if awk '
		/"name"[ \t]*:/   { name = $0; sub(/^.*"name"[ \t]*:[ \t]*"/, "", name); sub(/".*$/, "", name) }
		/"status"[ \t]*:/ {
			status = $0; sub(/^.*"status"[ \t]*:[ \t]*"/, "", status); sub(/".*$/, "", status)
			if (status == "excluded" && name != "") {
				goos = (name == "macos") ? "darwin" : name
				print goos
			}
			name = ""
		}
	' "$QV" | grep -qx "$GOOS_WANT"; then
		vector_excludes_this_goos=1
	fi
fi

while IFS="$(printf '\t')" read -r pkg assertcase defaults _note; do
	case "$pkg" in ''|\#*) continue ;; esac
	[ -n "$assertcase" ] || continue

	if [ -n "$QV" ]; then
		[ "$vector_excludes_this_goos" -eq 1 ] || continue
		source_of_truth="the root's own $QUALIFICATION_VECTOR"
	else
		case ",${defaults:-}," in
			*",$GOOS_WANT,"*) ;;
			*) continue ;;
		esac
		source_of_truth="default_excluded_on (this root publishes no $QUALIFICATION_VECTOR)"
	fi

	printf '%s\t%s\t%s\n' "$pkg" "$assertcase" "$source_of_truth"
done <"$EXCLUSIONS"

exit 0
