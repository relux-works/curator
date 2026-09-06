#!/usr/bin/env bash
# A module installed with an explicit version must not depend on main-module
# replace or exclude directives. Go rejects such modules before compilation.
# For tag releases, also prove the peeled tag commit belongs to the freshly
# fetched public main history, so an unmerged branch cannot publish artifacts.

set -u

module_file="${1:-go.mod}"
candidate_ref="${2:-}"
main_ref="${3:-}"

if [ ! -f "$module_file" ]; then
	echo "release-source-gate: module file not found: $module_file" >&2
	exit 1
fi

if ! awk '
	$1 == "replace" || $1 == "exclude" {
		printf "%s:%d: release-incompatible %s directive\n", FILENAME, NR, $1 > "/dev/stderr"
		bad = 1
	}
	END { exit bad }
' "$module_file"; then
	exit 1
fi

if { [ -n "$candidate_ref" ] && [ -z "$main_ref" ]; } ||
   { [ -z "$candidate_ref" ] && [ -n "$main_ref" ]; }; then
	echo 'release-source-gate: candidate and main refs must be supplied together' >&2
	exit 1
fi

if [ -n "$candidate_ref" ]; then
	candidate_commit=$(git rev-parse --verify "$candidate_ref^{commit}") || {
		echo "release-source-gate: cannot peel candidate ref: $candidate_ref" >&2
		exit 1
	}
	main_commit=$(git rev-parse --verify "$main_ref^{commit}") || {
		echo "release-source-gate: cannot resolve main ref: $main_ref" >&2
		exit 1
	}
	if ! git merge-base --is-ancestor "$candidate_commit" "$main_commit"; then
		echo "release-source-gate: candidate $candidate_commit is not contained in $main_ref ($main_commit)" >&2
		exit 1
	fi
fi

echo "release-source-gate: $module_file is compatible with versioned go install"
if [ -n "$candidate_ref" ]; then
	echo "release-source-gate: candidate $candidate_commit is contained in $main_ref ($main_commit)"
fi
