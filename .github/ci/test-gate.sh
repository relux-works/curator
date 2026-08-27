#!/usr/bin/env bash
# Compiled-build test gate.
#
# Runs `go test -json` over the supplied packages, then hands the stream to
# `platform-case-gate.sh`, which enforces `.github/ci/platform-cases.tsv`.
#
# Both statuses are reported and both are fatal: a green platform-case report
# never masks a red `go test`, and a green `go test` never masks a case that
# stopped running on this platform.
#
# Usage:
#   test-gate.sh <evidence-dir> [package...]
#
# Environment:
#   GO                       go launcher             (default: go)
#   GO_TEST_TIMEOUT          per-package timeout     (default: 30m)
#   GO_TEST_FLAGS            extra flags, word split (e.g. "-race")
#   CI_PLATFORM_CASES        ledger path             (default: .github/ci/platform-cases.tsv)
#   CURATOR_CONFORMANCE_ROOT required; forwarded to the tests untouched
#
# The gate never rewrites, relaxes or retries a failing run.

set -u

GO="${GO:-go}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-30m}"
HERE="$(cd "$(dirname "$0")" && pwd)"

if [ "$#" -lt 1 ]; then
	echo 'test-gate: usage: test-gate.sh <evidence-dir> [package...]' >&2
	exit 2
fi
EVIDENCE="$1"
shift
[ "$#" -gt 0 ] || set -- ./...

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
JSON="$EVIDENCE/go-test.json"

echo "test-gate: packages=$*"
echo "test-gate: flags=${GO_TEST_FLAGS:-<none>} timeout=$GO_TEST_TIMEOUT"
echo "test-gate: conformance root=$CURATOR_CONFORMANCE_ROOT"

# The gate command runs as its own process, unpiped, so $? is its real status.
# shellcheck disable=SC2086
"$GO" test -json -count=1 -timeout "$GO_TEST_TIMEOUT" ${GO_TEST_FLAGS:-} "$@" >"$JSON" 2>&1
TEST_RC=$?
echo "test-gate: go test exit=$TEST_RC (stream: $JSON)"

bash "$HERE/platform-case-gate.sh" "$JSON" "$EVIDENCE"
GATE_RC=$?

echo "test-gate: go test exit=$TEST_RC, platform-case gate exit=$GATE_RC"

if [ "$TEST_RC" -ne 0 ]; then exit "$TEST_RC"; fi
exit "$GATE_RC"
