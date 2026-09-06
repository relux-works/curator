#!/usr/bin/env bash

# Validate the complete workflow path protected by release-source-gate.sh.
# Optional arguments allow gate-selftest.sh to exercise narrowed workflow
# mutants without changing the production workflows.

set -u

CI_WORKFLOW="${1:-.github/workflows/ci.yml}"
RELEASE_WORKFLOW="${2:-.github/workflows/release.yml}"

fail() {
	printf 'release workflow gate: %s\n' "$1" >&2
	exit 1
}

[ -r "$CI_WORKFLOW" ] || fail "cannot read CI workflow: $CI_WORKFLOW"
[ -r "$RELEASE_WORKFLOW" ] || fail "cannot read release workflow: $RELEASE_WORKFLOW"

exact_run_lines() {
	awk -v command="$2" '
		{
			line = $0
			sub(/^[ \t]*run:[ \t]*/, "", line)
			sub(/[ \t]*$/, "", line)
			if (line == command) print NR
		}
	' "$1"
}

ci_gate_lines="$(exact_run_lines "$CI_WORKFLOW" 'bash .github/ci/release-source-gate.sh')"
[ "$(printf '%s\n' "$ci_gate_lines" | awk 'NF { count++ } END { print count + 0 }')" -eq 1 ] ||
	fail "normal CI must invoke the module release-source gate exactly once"

fetch_lines="$(grep -nF 'git fetch --no-tags origin +refs/heads/main:refs/remotes/origin/main' "$RELEASE_WORKFLOW" | cut -d: -f1)"
release_gate_lines="$(exact_run_lines "$RELEASE_WORKFLOW" 'bash .github/ci/release-source-gate.sh go.mod "$GITHUB_SHA" refs/remotes/origin/main')"
publisher_lines="$(grep -nF 'uses: goreleaser/goreleaser-action@' "$RELEASE_WORKFLOW" | cut -d: -f1)"

[ "$(printf '%s\n' "$fetch_lines" | awk 'NF { count++ } END { print count + 0 }')" -eq 1 ] ||
	fail "release workflow must fetch public main exactly once"
[ "$(printf '%s\n' "$release_gate_lines" | awk 'NF { count++ } END { print count + 0 }')" -eq 1 ] ||
	fail "release workflow must invoke the ancestry gate exactly once"
[ -n "$publisher_lines" ] || fail "release workflow has no GoReleaser publication action"

fetch_line="$fetch_lines"
release_gate_line="$release_gate_lines"
[ "$fetch_line" -lt "$release_gate_line" ] ||
	fail "fresh public-main fetch must precede the release ancestry gate"
for publisher_line in $publisher_lines; do
	[ "$release_gate_line" -lt "$publisher_line" ] ||
		fail "release ancestry gate must precede every GoReleaser publication action"
done

printf 'release workflow gate: protected CI and publication paths verified\n'
