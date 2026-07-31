#!/usr/bin/env bash
# Compiled-build test gate.
#
# One entry point for every lane. It plans the run from the supplied
# conformance root, executes it, and enforces the platform-case ledger against
# the real `go test -json` stream. Every status is reported and every status is
# fatal: a green ledger never masks a red `go test`, and a green `go test` never
# masks a case that stopped running on this platform.
#
# The run is up to three `go test` invocations, decided by
# `.github/ci/suite-plan.sh` from the root alone:
#
#   served    packages whose conformance artefacts the root publishes, run WITH
#             CURATOR_CONFORMANCE_ROOT exported;
#   deferred  packages the root cannot serve, run with the variable UNSET so
#             their own `root is not set` path is taken -- named in the report,
#             and fatal in a lane that sets CI_REQUIRE_FULL_ROOT=1;
#   excluded  packages the root's qualification vector does not qualify on this
#             platform. These are not skipped quietly: the rejection case named
#             in `.github/ci/platform-exclusions.tsv` is run on this very runner
#             and must pass, proving the exclusion is fail-closed behaviour.
#
# Usage:
#   test-gate.sh <evidence-dir>
#
# Environment:
#   GO                       go launcher             (default: go)
#   GO_TEST_TIMEOUT          per-package timeout     (default: 30m)
#   GO_TEST_FLAGS            extra flags, word split (e.g. "-race")
#   CURATOR_CONFORMANCE_ROOT required; forwarded to the served packages untouched
#   CI_REQUIRE_FULL_ROOT     1 = the root must serve every package
#
# The gate never rewrites, relaxes or retries a failing run.

set -u

GO="${GO:-go}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-30m}"
HERE="$(cd "$(dirname "$0")" && pwd)"

if [ "$#" -ne 1 ]; then
	echo 'test-gate: usage: test-gate.sh <evidence-dir>' >&2
	exit 2
fi
EVIDENCE="$1"

# A gate that runs with the conformance suite unset is a smaller gate wearing
# the same name. Refuse up front rather than reporting a green partial run.
if [ -z "${CURATOR_CONFORMANCE_ROOT:-}" ]; then
	echo 'test-gate: CURATOR_CONFORMANCE_ROOT is required.' >&2
	echo 'test-gate: CI exports it from the committed SPEC_PIN checkout; locally, point' >&2
	echo 'test-gate: it at a materialised <curator-spec>/conformance/v1 directory.' >&2
	exit 2
fi
if [ ! -f "${CURATOR_CONFORMANCE_ROOT}/manifest.json" ]; then
	echo "test-gate: CURATOR_CONFORMANCE_ROOT has no manifest.json: $CURATOR_CONFORMANCE_ROOT" >&2
	exit 2
fi

mkdir -p "$EVIDENCE" || exit 2
ROOT="$CURATOR_CONFORMANCE_ROOT"
MODULE="$(awk '/^module[ \t]/{print $2; exit}' go.mod 2>/dev/null)"
[ -n "$MODULE" ] || { echo 'test-gate: cannot determine the module path' >&2; exit 2; }

echo "test-gate: flags=${GO_TEST_FLAGS:-<none>} timeout=$GO_TEST_TIMEOUT"
echo "test-gate: conformance root=$ROOT"
echo ''

# --- plan ------------------------------------------------------------------
bash "$HERE/suite-plan.sh" "$ROOT" "$EVIDENCE"
PLAN_RC=$?
if [ "$PLAN_RC" -ne 0 ]; then
	echo "test-gate: suite-plan exit=$PLAN_RC -- refusing to run a suite whose shape is already wrong" >&2
	exit "$PLAN_RC"
fi

SERVED="$EVIDENCE/plan-served.txt"
DEFERRED="$EVIDENCE/plan-deferred.txt"
EXCLUDED="$EVIDENCE/plan-excluded.txt"
ASSERT="$EVIDENCE/plan-assert.txt"
JSON="$EVIDENCE/go-test.json"
: >"$JSON"

short_names() { sed -e "s#^$MODULE/##" -e "s#^$MODULE\$#.#" "$1" | tr '\n' ' '; }

run_stage() {
	stage="$1"; shift
	out="$EVIDENCE/go-test-$stage.json"
	# The gate command runs as its own process, unpiped, so $? is its real status.
	# shellcheck disable=SC2086
	"$@" >"$out" 2>&1
	rc=$?
	cat "$out" >>"$JSON"
	echo "test-gate: stage $stage exit=$rc (stream: $out)"
	return $rc
}

TEST_RC=0

# --- served: the root is exported ------------------------------------------
if [ -s "$SERVED" ]; then
	# shellcheck disable=SC2046,SC2086
	run_stage served env CURATOR_CONFORMANCE_ROOT="$ROOT" \
		"$GO" test -json -count=1 -timeout "$GO_TEST_TIMEOUT" ${GO_TEST_FLAGS:-} $(cat "$SERVED")
	rc=$?; [ "$rc" -ne 0 ] && TEST_RC="$rc"
else
	echo 'test-gate: no served packages' >&2
	exit 2
fi

# --- deferred: the root is unset, deliberately and visibly -----------------
if [ -s "$DEFERRED" ]; then
	# shellcheck disable=SC2046,SC2086
	run_stage deferred env -u CURATOR_CONFORMANCE_ROOT \
		"$GO" test -json -count=1 -timeout "$GO_TEST_TIMEOUT" ${GO_TEST_FLAGS:-} $(cat "$DEFERRED")
	rc=$?; [ "$rc" -ne 0 ] && TEST_RC="$rc"
fi

# --- excluded: the exclusion is asserted, not assumed ----------------------
if [ -s "$ASSERT" ]; then
	while IFS="$(printf '\t')" read -r importpath assertcase; do
		[ -n "$importpath" ] || continue
		# shellcheck disable=SC2086
		run_stage "assert-$(basename "$importpath")" env CURATOR_CONFORMANCE_ROOT="$ROOT" \
			"$GO" test -json -count=1 -timeout "$GO_TEST_TIMEOUT" ${GO_TEST_FLAGS:-} \
			-run "^${assertcase}\$" "$importpath"
		rc=$?; [ "$rc" -ne 0 ] && TEST_RC="$rc"
	done <"$ASSERT"
fi

echo ''
echo "test-gate: go test overall exit=$TEST_RC (merged stream: $JSON)"
echo ''

# --- ledger ----------------------------------------------------------------
CI_DEFERRED_PKGS="$(short_names "$DEFERRED")" \
CI_EXCLUDED_PKGS="$(short_names "$EXCLUDED")" \
	bash "$HERE/platform-case-gate.sh" "$JSON" "$EVIDENCE"
GATE_RC=$?

echo ''
echo "test-gate: go test exit=$TEST_RC, platform-case gate exit=$GATE_RC"

if [ "$TEST_RC" -ne 0 ]; then exit "$TEST_RC"; fi
exit "$GATE_RC"
