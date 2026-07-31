#!/usr/bin/env bash
# Platform case gate.
#
# Enforces two tiers against a real `go test -json` stream, so a green
# `go test` can never mean a smaller suite than the one that was asked for.
#
#   TIER 1 -- `.github/ci/platform-cases.tsv`
#     The compiled-build behaviour the acceptance criteria require this runner
#     to execute. Every row listing this GOOS in `must_run_on` must be observed
#     PASSING. A rename, a deletion, a `-run` filter matching nothing, a missing
#     interpreter or a package that stopped being built all fail here BY NAME
#     rather than shrinking the run. A skip of a listed case -- or of any of its
#     subtests -- is fatal unless the row tolerates it on this GOOS, and then
#     only if the reason the test actually printed carries the class the row
#     declared.
#
#   TIER 2 -- `.github/ci/skip-classes.tsv`
#     Every OTHER skip anywhere in the run must have a reason matching a
#     declared class, and that class must be allowed in this lane. An
#     unrecognised reason is fatal: a newly-introduced skip is caught the first
#     time it runs. `root-unset` is allowed only for a package `suite-plan.sh`
#     deferred, so a lane demanding a fully-serving root cannot pass while a
#     package quietly ran without its vectors.
#
# Every skip, tolerated or not, is written to `skips-observed.tsv` with its
# class and verbatim reason. "No case silently skipped" is therefore a property
# of the attached evidence, not of a promise.
#
# Usage:
#   platform-case-gate.sh <go-test-json> <evidence-dir>
#
# Environment:
#   CI_PLATFORM_CASES  ledger path         (default: .github/ci/platform-cases.tsv)
#   CI_SKIP_CLASSES    class table         (default: .github/ci/skip-classes.tsv)
#   CI_DEFERRED_PKGS   space separated packages running without the root
#   CI_EXCLUDED_PKGS   space separated packages this platform does not execute
#   CI_GATE_GOOS       GOOS to enforce     (default: `go env GOOS`)
#   CI_GATE_MODULE     module path         (default: the `module` line of go.mod)
#
# The last two overrides exist so `gate-selftest.sh` can drive this gate against
# synthetic streams for platforms the current host is not.

set -u

LEDGER="${CI_PLATFORM_CASES:-.github/ci/platform-cases.tsv}"
CLASSES="${CI_SKIP_CLASSES:-.github/ci/skip-classes.tsv}"

if [ "$#" -ne 2 ]; then
	echo 'platform-case-gate: usage: platform-case-gate.sh <go-test-json> <evidence-dir>' >&2
	exit 2
fi
JSON="$1"
EVIDENCE="$2"

[ -f "$JSON" ]    || { echo "platform-case-gate: no such stream: $JSON" >&2; exit 2; }
[ -f "$LEDGER" ]  || { echo "platform-case-gate: ledger not found: $LEDGER" >&2; exit 2; }
[ -f "$CLASSES" ] || { echo "platform-case-gate: skip-class table not found: $CLASSES" >&2; exit 2; }
mkdir -p "$EVIDENCE" || exit 2

GOOS="${CI_GATE_GOOS:-$(go env GOOS)}"
[ -n "$GOOS" ] || { echo 'platform-case-gate: cannot determine GOOS' >&2; exit 2; }
MODULE="${CI_GATE_MODULE:-$(awk '/^module[ \t]/{print $2; exit}' go.mod 2>/dev/null)}"
[ -n "$MODULE" ] || { echo 'platform-case-gate: cannot determine the module path' >&2; exit 2; }

OBSERVED="$EVIDENCE/observed-cases.tsv"
SKIPS="$EVIDENCE/skips-observed.tsv"
REPORT="$EVIDENCE/platform-cases.txt"

# ---------------------------------------------------------------------------
# Extract (action, package, test, reason) from the test2json stream.
#
# test2json emits keys in a fixed order -- Time, Action, Package, Test, Output --
# so the FIRST occurrence of a key is always the real field, even when an
# `output` event carries JSON-looking test output inside its Output string.
#
# The skip reason is the last `file.go:NN: ...` line the case printed before its
# result event, which is exactly where testing writes a Skip message.
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
	if (act != "output" && act != "skip" && act != "pass" && act != "fail") next
	tst = jstr($0, "Test")
	if (tst == "") next                 # package-level event, not a case
	pkg = jstr($0, "Package")
	if (pkg == module) pkg = "."
	else if (substr(pkg, 1, length(module) + 1) == module "/") pkg = substr(pkg, length(module) + 2)
	key = pkg "\t" tst

	if (act == "output") {
		out = jstr($0, "Output")
		# the message testing prints for Skip/Log: "    file.go:12: reason"
		if (match(out, /[^ \t]+\.go:[0-9]+: /)) {
			reason = substr(out, RSTART + RLENGTH)
			sub(/[\r\n]+$/, "", reason)
			gsub(/\t/, " ", reason)
			if (reason != "") last[key] = reason
		}
		next
	}
	print act "\t" key "\t" (act == "skip" ? last[key] : "")
}
' "$JSON" | sort -u >"$OBSERVED" || { echo 'platform-case-gate: stream extraction failed' >&2; exit 2; }

# ---------------------------------------------------------------------------
# Enforce both tiers.
# ---------------------------------------------------------------------------
awk -v goos="$GOOS" -v ledger="$LEDGER" -v classes="$CLASSES" -v observed="$OBSERVED" \
    -v skipsout="$SKIPS" -v deferred="${CI_DEFERRED_PKGS:-}" -v excluded="${CI_EXCLUDED_PKGS:-}" '
function listed(set, want,   n, i, a) {
	if (set == "-" || set == "") return 0
	n = split(set, a, ",")
	for (i = 1; i <= n; i++) if (a[i] == want) return 1
	return 0
}
function classify(reason,   i) {
	for (i = 1; i <= nclass; i++)
		if (reason ~ cregex[i]) return i
	return 0
}
BEGIN {
	FS = "\t"; fail = 0

	n = split(deferred, d, " "); for (i = 1; i <= n; i++) isdeferred[d[i]] = 1
	n = split(excluded, x, " "); for (i = 1; i <= n; i++) isexcluded[x[i]] = 1

	while ((getline line < classes) > 0) {
		if (line ~ /^[ \t]*#/ || line ~ /^[ \t]*$/) continue
		split(line, f, "\t")
		nclass++
		cclass[nclass]  = f[1]
		cregex[nclass]  = f[2]
		cpolicy[nclass] = f[3]
	}
	close(classes)
	if (nclass == 0) { print "FAIL  the skip-class table is empty"; exit 1 }

	while ((getline line < ledger) > 0) {
		if (line ~ /^[ \t]*#/ || line ~ /^[ \t]*$/) continue
		split(line, f, "\t")
		key = f[1] "\t" f[2]
		mustrun[key]   = f[3]
		skipok[key]    = f[4]
		wantclass[key] = f[5]
		behaviour[key] = f[6]
		order[++rows]  = key
		if (f[2] ~ /\/\*$/) {
			g = f[2]; sub(/\/\*$/, "", g)
			wildkey = f[1] "\t" g
			wildskip[wildkey]  = f[4]
			wildclass[wildkey] = f[5]
		}
	}
	close(ledger)

	while ((getline line < observed) > 0) {
		split(line, o, "\t")
		act = o[1]; key = o[2] "\t" o[3]; reason = o[4]
		seen[act "\t" key] = 1
		parent = o[3]
		if (sub("/.*$", "", parent)) seen[act "\t" o[2] "\t" parent] = 1
		if (act == "skip") {
			skiporder[++skips] = key
			skipreason[key] = reason
			skippkg[key] = o[2]
			skiptest[key] = o[3]
		}
	}
	close(observed)

	print "platform-case gate: GOOS=" goos
	print "platform-case gate: ledger=" ledger
	if (deferred != "") print "platform-case gate: deferred packages (root unset): " deferred
	if (excluded != "") print "platform-case gate: excluded packages (not run here): " excluded
	print ""

	# ---- every skip is recorded, then judged --------------------------------
	printf "package\ttest\tclass\tverdict\treason\n" >skipsout
	for (i = 1; i <= skips; i++) {
		key = skiporder[i]
		reason = skipreason[key]
		ci = classify(reason)
		cls = (ci ? cclass[ci] : "UNCLASSIFIED")

		# Tier 1: a ledger case, or a subtest of one, has its own rule.
		parent = skiptest[key]; hassub = sub("/.*$", "", parent)
		wkey = skippkg[key] "\t" parent
		tol = ""; want = ""
		if (key in skipok)        { tol = skipok[key];   want = wantclass[key] }
		else if (hassub && (wkey in wildskip)) { tol = wildskip[wkey]; want = wildclass[wkey] }
		else if (!hassub && (wkey in skipok)) { tol = skipok[wkey];   want = wantclass[wkey] }

		verdict = ""
		if (tol != "") {
			if (!listed(tol, goos)) {
				printf "FAIL  ledger case skipped where the ledger does not tolerate it (%s): %s :: %s\n", goos, skippkg[key], skiptest[key]
				printf "      reason: %s\n", reason
				printf "      required behaviour: %s\n", behaviour[skippkg[key] "\t" parent]
				fail = 1; verdict = "FATAL-not-tolerated"
			} else if (want != "-" && want != "" && want != cls) {
				printf "FAIL  ledger case skipped for the wrong reason on %s: %s :: %s\n", goos, skippkg[key], skiptest[key]
				printf "      the ledger tolerates a %s skip here; this one classified as %s\n", want, cls
				printf "      reason: %s\n", reason
				fail = 1; verdict = "FATAL-wrong-class"
			} else {
				verdict = "tolerated-by-ledger"
			}
		} else if (ci == 0) {
			printf "FAIL  skip with an unrecognised reason on %s: %s :: %s\n", goos, skippkg[key], skiptest[key]
			printf "      reason: %s\n", reason
			print  "      add it to .github/ci/skip-classes.tsv with a class, or fix the case."
			fail = 1; verdict = "FATAL-unclassified"
		} else if (cpolicy[ci] == "deferred-only" && !(skippkg[key] in isdeferred)) {
			printf "FAIL  %s skip in a package this lane serves: %s :: %s\n", cls, skippkg[key], skiptest[key]
			printf "      reason: %s\n", reason
			print  "      this lane supplies a root that must serve this package; it ran without one."
			fail = 1; verdict = "FATAL-served-package"
		} else {
			verdict = "allowed-" cls
		}
		printf "%s\t%s\t%s\t%s\t%s\n", skippkg[key], skiptest[key], cls, verdict, reason >skipsout
	}
	close(skipsout)

	# ---- Tier 1: every required case observed passing -----------------------
	for (i = 1; i <= rows; i++) {
		key = order[i]
		split(key, k, "\t")
		if (k[2] ~ /\/\*$/) continue
		if (!listed(mustrun[key], goos)) continue
		if (k[1] in isexcluded) {
			printf "excl  %s :: %s (package not executed on %s)\n", k[1], k[2], goos
			continue
		}
		if (seen["pass\t" key]) { printf "ok    %s :: %s\n", k[1], k[2]; continue }
		# A row may both require a case and tolerate its skip here -- "run it
		# whenever this runner and this root can". The skip pass above has
		# already checked the reason carried the class the row declared, so a
		# tolerated skip is a recorded outcome, not a silent one.
		if (seen["skip\t" key] && listed(skipok[key], goos)) {
			printf "tol   %s :: %s (tolerated skip: %s)\n", k[1], k[2], wantclass[key]
			continue
		}
		if (seen["fail\t" key]) {
			printf "FAIL  required case failed: %s :: %s\n", k[1], k[2]
		} else if (seen["skip\t" key]) {
			printf "FAIL  required case skipped on %s: %s :: %s\n", goos, k[1], k[2]
			printf "      required behaviour: %s\n", behaviour[key]
		} else {
			printf "FAIL  required case never ran on %s: %s :: %s\n", goos, k[1], k[2]
			printf "      required behaviour: %s\n", behaviour[key]
			print  "      a rename, a deleted test or a -run filter matching nothing all look like this."
		}
		fail = 1
	}

	printf "\nplatform-case gate: %d skips recorded in %s\n", skips, skipsout
	print (fail ? "platform-case gate: FAILED" : "platform-case gate: ok")
	exit fail
}
' </dev/null >"$REPORT" 2>&1
GATE_RC=$?
cat "$REPORT"
exit "$GATE_RC"
