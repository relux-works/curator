#!/usr/bin/env bash
# Platform case gate: enforces `.github/ci/platform-cases.tsv` against a
# `go test -json` stream.
#
#   * a skip is only tolerated when a ledger row lists this GOOS in
#     `skip_allowed_on`; every other skip fails the gate, so a case can never
#     disappear quietly;
#   * every ledger row listing this GOOS in `must_run_on` must be observed
#     passing. A rename, a deletion, a `-run` filter that matches nothing, a
#     missing interpreter or an unset conformance root all surface here by
#     name instead of shrinking the run.
#
# Usage:
#   platform-case-gate.sh <go-test-json> <evidence-dir>
#
# Environment:
#   CI_PLATFORM_CASES  ledger path       (default: .github/ci/platform-cases.tsv)
#   CI_GATE_GOOS       GOOS to enforce   (default: `go env GOOS`)
#   CI_GATE_MODULE     module path       (default: the `module` line of go.mod)
#
# CI_GATE_GOOS and CI_GATE_MODULE exist so `gate-selftest.sh` can drive this
# gate against synthetic streams for platforms the current host is not.

set -u

LEDGER="${CI_PLATFORM_CASES:-.github/ci/platform-cases.tsv}"

if [ "$#" -ne 2 ]; then
	echo 'platform-case-gate: usage: platform-case-gate.sh <go-test-json> <evidence-dir>' >&2
	exit 2
fi
JSON="$1"
EVIDENCE="$2"

[ -f "$JSON" ]   || { echo "platform-case-gate: no such stream: $JSON" >&2; exit 2; }
[ -f "$LEDGER" ] || { echo "platform-case-gate: ledger not found: $LEDGER" >&2; exit 2; }
mkdir -p "$EVIDENCE" || exit 2

GOOS="${CI_GATE_GOOS:-$(go env GOOS)}"
[ -n "$GOOS" ] || { echo 'platform-case-gate: cannot determine GOOS' >&2; exit 2; }
MODULE="${CI_GATE_MODULE:-$(awk '/^module[ \t]/{print $2; exit}' go.mod 2>/dev/null)}"
[ -n "$MODULE" ] || { echo 'platform-case-gate: cannot determine the module path' >&2; exit 2; }

OBSERVED="$EVIDENCE/observed-cases.tsv"
REPORT="$EVIDENCE/platform-cases.txt"

# ---------------------------------------------------------------------------
# Extract (action, package, test) triples from the test2json stream.
#
# test2json emits keys in a fixed order -- Time, Action, Package, Test, Output --
# so the FIRST occurrence of a key is always the real field, even when an
# `output` event carries JSON-looking test output inside its Output string.
# ---------------------------------------------------------------------------
awk -v module="$MODULE" '
function jstr(line, key,   i, n, out, c, esc) {
	i = index(line, "\"" key "\":\"")
	if (i == 0) return ""
	i += length(key) + 4
	out = ""; esc = 0
	for (n = i; n <= length(line); n++) {
		c = substr(line, n, 1)
		if (esc) {
			if      (c == "n") out = out "\n"
			else if (c == "t") out = out "\t"
			else if (c == "r") out = out "\r"
			else               out = out c
			esc = 0
			continue
		}
		if (c == "\\") { esc = 1; continue }
		if (c == "\"") break
		out = out c
	}
	return out
}
{
	act = jstr($0, "Action")
	if (act != "skip" && act != "pass" && act != "fail") next
	tst = jstr($0, "Test")
	if (tst == "") next            # package-level event, not a test
	pkg = jstr($0, "Package")
	if (pkg == module) pkg = "."
	else if (substr(pkg, 1, length(module) + 1) == module "/") pkg = substr(pkg, length(module) + 2)
	print act "\t" pkg "\t" tst
}
' "$JSON" | sort -u >"$OBSERVED" || { echo 'platform-case-gate: stream extraction failed' >&2; exit 2; }

# ---------------------------------------------------------------------------
# Enforce the ledger.
# ---------------------------------------------------------------------------
awk -v goos="$GOOS" -v ledger="$LEDGER" -v observed="$OBSERVED" '
function listed(set, want,   n, i, a) {
	if (set == "-" || set == "") return 0
	n = split(set, a, ",")
	for (i = 1; i <= n; i++) if (a[i] == want) return 1
	return 0
}
BEGIN {
	fail = 0

	while ((getline line < ledger) > 0) {
		if (line ~ /^[ \t]*#/ || line ~ /^[ \t]*$/) continue
		split(line, f, "\t")
		key = f[1] "\t" f[2]
		mustrun[key]   = f[3]
		skipok[key]    = f[4]
		behaviour[key] = f[5]
		order[++rows]  = key
	}
	close(ledger)

	while ((getline line < observed) > 0) {
		split(line, o, "\t")
		key = o[2] "\t" o[3]
		seen[o[1] "\t" key] = 1
		# a subtest result is also evidence about its parent case
		parent = o[3]
		if (sub("/.*$", "", parent)) seen[o[1] "\t" o[2] "\t" parent] = 1
		if (o[1] == "skip") { skipped[key] = 1; skiporder[++skips] = key }
	}
	close(observed)

	print "platform-case gate: GOOS=" goos
	print "platform-case gate: ledger=" ledger
	print ""

	# Rule 1 -- no skip this GOOS does not explicitly tolerate.
	for (i = 1; i <= skips; i++) {
		key = skiporder[i]
		if ((key in skipok) && listed(skipok[key], goos)) continue
		split(key, k, "\t")
		printf "FAIL  unexpected skip on %s: %s :: %s\n", goos, k[1], k[2]
		print  "      the ledger does not tolerate this skip here; either the case"
		print  "      regressed or it needs an explicit row with a stated reason."
		fail = 1
	}

	# Rule 2 -- every case this GOOS requires, observed passing.
	for (i = 1; i <= rows; i++) {
		key = order[i]
		if (!listed(mustrun[key], goos)) continue
		split(key, k, "\t")
		if (seen["pass\t" key]) { printf "ok    %s :: %s\n", k[1], k[2]; continue }
		if (seen["fail\t" key]) {
			printf "FAIL  required case failed: %s :: %s\n", k[1], k[2]
		} else if (key in skipped) {
			printf "FAIL  required case skipped on %s: %s :: %s\n", goos, k[1], k[2]
			printf "      required behaviour: %s\n", behaviour[key]
		} else {
			printf "FAIL  required case never ran on %s: %s :: %s\n", goos, k[1], k[2]
			printf "      required behaviour: %s\n", behaviour[key]
			print  "      a rename, a deleted test or a -run filter matching nothing all look like this."
		}
		fail = 1
	}

	print ""
	print (fail ? "platform-case gate: FAILED" : "platform-case gate: ok")
	exit fail
}
' </dev/null >"$REPORT" 2>&1
GATE_RC=$?
cat "$REPORT"
exit "$GATE_RC"
