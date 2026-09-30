#!/usr/bin/env bash
# Gate self-test.
#
# Every gate in this directory is only worth its exit code if it rejects what
# it claims to reject. This drives each one against synthetic inputs and
# asserts a REAL exit code for every case -- including the negative cases,
# which are the ones a gate silently loses when it is refactored.
#
# It needs no conformance root or network. One timeout regression builds and
# runs a tiny synthetic Go package; cases that need a `go` launcher are named
# and reported when one is absent, never quietly dropped.
#
# Usage:
#   gate-selftest.sh

set -u

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOTDIR="$(cd "$HERE/../.." && pwd)"
cd "$ROOTDIR" || exit 2

WORK="$(mktemp -d "${TMPDIR:-/tmp}/gate-selftest.XXXXXX")" || exit 2
trap 'rm -rf "$WORK"' EXIT

PASS=0; FAIL=0; SKIPPED=0

# Read the committed pin out of the workflow rather than duplicating it, so
# these cases also assert that the pin CI actually uses is itself a full,
# immutable, lowercase 40-hex revision -- the same shape a candidate must have.
PIN="$(awk '/^[ \t]*SPEC_PIN:[ \t]*/{print $2; exit}' .github/workflows/ci.yml)"

# The module path, so cases can name import paths the way `go list` prints them.
MODULE_PATH="$(awk '/^module[ \t]/{print $2; exit}' go.mod)"

ok()   { PASS=$((PASS + 1)); printf 'ok    %s\n' "$1"; }
bad()  { FAIL=$((FAIL + 1)); printf 'FAIL  %s\n      %s\n' "$1" "$2"; }
skip() { SKIPPED=$((SKIPPED + 1)); printf 'skip  %s\n      %s\n' "$1" "$2"; }

# assert <name> <want-exit> <command...>
assert() {
	name="$1"; want="$2"; shift 2
	out="$WORK/out.txt"
	"$@" >"$out" 2>&1
	got=$?
	if [ "$got" -eq "$want" ]; then ok "$name"
	else bad "$name" "exit $got, want $want; output: $(tr '\n' ' ' <"$out" | cut -c1-200)"
	fi
}

# assert_contains <name> <needle> <file>
assert_contains() {
	if grep -qF "$2" "$3"; then ok "$1"
	else bad "$1" "expected to find: $2"
	fi
}

echo '=== candidate-suite.sh verify-ref: only a full immutable revision is a candidate ==='
CS="$HERE/candidate-suite.sh"
assert 'candidate inputs reject revision plus root' 1 bash "$CS" verify-inputs '1234567890abcdef1234567890abcdef12345678' '/candidate/root'
assert 'candidate inputs accept revision only'     0 bash "$CS" verify-inputs '1234567890abcdef1234567890abcdef12345678' ''
assert 'candidate inputs accept root only'         0 bash "$CS" verify-inputs '' '/candidate/root'

echo ''
echo '=== release-source-gate.sh: versioned Go installs stay publishable ==='
RSG="$HERE/release-source-gate.sh"
printf 'module example.test/clean\n\ngo 1.25.5\n\nrequire example.test/helper v1.0.0\n' >"$WORK/go-clean.mod"
printf 'module example.test/replaced\n\ngo 1.25.5\n\nreplace example.test/helper => ./helper\n' >"$WORK/go-replace.mod"
printf 'module example.test/excluded\n\ngo 1.25.5\n\nexclude example.test/helper v1.0.0\n' >"$WORK/go-exclude.mod"
assert 'a release module without local interpretation directives passes' 0 bash "$RSG" "$WORK/go-clean.mod"
assert 'a replace directive that breaks versioned go install is rejected' 1 bash "$RSG" "$WORK/go-replace.mod"
assert 'an exclude directive that changes versioned resolution is rejected' 1 bash "$RSG" "$WORK/go-exclude.mod"
assert 'a missing module file is a read failure, not a clean module' 1 bash "$RSG" "$WORK/go-missing.mod"
assert 'the repository module passes the release-source gate' 0 bash "$RSG" go.mod

ANCESTRY_REPO="$WORK/release-ancestry"
git init -q -b main "$ANCESTRY_REPO"
printf 'main\n' >"$ANCESTRY_REPO/content.txt"
git -C "$ANCESTRY_REPO" add content.txt
git -C "$ANCESTRY_REPO" -c user.name='Gate Fixture' -c user.email='gate@example.invalid' commit -q -m main
MAIN_COMMIT="$(git -C "$ANCESTRY_REPO" rev-parse HEAD)"
git -C "$ANCESTRY_REPO" switch -q -c topic
printf 'topic\n' >"$ANCESTRY_REPO/content.txt"
git -C "$ANCESTRY_REPO" add content.txt
git -C "$ANCESTRY_REPO" -c user.name='Gate Fixture' -c user.email='gate@example.invalid' commit -q -m topic
TOPIC_COMMIT="$(git -C "$ANCESTRY_REPO" rev-parse HEAD)"
assert 'a candidate commit contained in main is accepted' 0 bash -c 'cd "$1" && "$2" "$3" "$4" main' _ "$ANCESTRY_REPO" "$RSG" "$WORK/go-clean.mod" "$MAIN_COMMIT"
assert 'an unmerged branch candidate is rejected' 1 bash -c 'cd "$1" && "$2" "$3" "$4" main' _ "$ANCESTRY_REPO" "$RSG" "$WORK/go-clean.mod" "$TOPIC_COMMIT"
assert 'missing ancestry evidence is rejected, not treated as absent' 1 bash -c 'cd "$1" && "$2" "$3" "$4" missing-main' _ "$ANCESTRY_REPO" "$RSG" "$WORK/go-clean.mod" "$MAIN_COMMIT"
assert 'a half-specified ancestry check is rejected' 1 bash "$RSG" "$WORK/go-clean.mod" "$MAIN_COMMIT"

WORKFLOW='.github/workflows/ci.yml'
RELEASE_WORKFLOW='.github/workflows/release.yml'
RWG="$HERE/release-workflow-gate.sh"
assert 'production workflows invoke the source gates before every publisher' 0 bash "$RWG" "$WORKFLOW" "$RELEASE_WORKFLOW"

cp "$WORKFLOW" "$WORK/ci-without-release-source-gate.yml"
sed -i.bak '/run: bash \.github\/ci\/release-source-gate\.sh$/d' "$WORK/ci-without-release-source-gate.yml"
assert 'workflow checker rejects a missing normal-CI source gate' 1 bash "$RWG" "$WORK/ci-without-release-source-gate.yml" "$RELEASE_WORKFLOW"

cp "$RELEASE_WORKFLOW" "$WORK/release-without-source-gate.yml"
sed -i.bak '/run: bash \.github\/ci\/release-source-gate\.sh go\.mod/d' "$WORK/release-without-source-gate.yml"
assert 'workflow checker rejects a missing release ancestry gate' 1 bash "$RWG" "$WORKFLOW" "$WORK/release-without-source-gate.yml"

awk '
	!inserted && /^[[:space:]]*steps:/ {
		print
		print "      - uses: goreleaser/goreleaser-action@v6"
		inserted = 1
		next
	}
	{ print }
' "$RELEASE_WORKFLOW" >"$WORK/release-with-early-publisher.yml"
assert 'workflow checker rejects an additional publisher before the gate' 1 bash "$RWG" "$WORKFLOW" "$WORK/release-with-early-publisher.yml"

validation_line="$(grep -nF 'Reject ambiguous candidate inputs' "$WORKFLOW" | cut -d: -f1)"
checkout_line="$(grep -nF 'Check out the candidate suite' "$WORKFLOW" | cut -d: -f1)"
record_line="$(grep -nF 'Resolve and record the candidate suite identity' "$WORKFLOW" | cut -d: -f1)"
if [ -n "$validation_line" ] &&
   [ -n "$checkout_line" ] &&
   [ -n "$record_line" ] &&
   [ "$validation_line" -lt "$checkout_line" ] &&
   [ "$validation_line" -lt "$record_line" ]; then
	ok 'workflow rejects ambiguous inputs before candidate checkout or recording'
else
	bad 'workflow rejects ambiguous inputs before candidate checkout or recording' \
		"validation=$validation_line checkout=$checkout_line record=$record_line"
fi

if [ -z "$PIN" ]; then
	bad 'the workflow declares a SPEC_PIN' 'no SPEC_PIN: line in .github/workflows/ci.yml'
	PIN='0000000000000000000000000000000000000001'
else
	# The committed pin must satisfy the same immutability shape a candidate
	# does: a branch or a mutable tag committed as the default pin is exactly
	# what `verify-ref` exists to make impossible.
	assert 'the committed SPEC_PIN is itself a full immutable revision' 0 env -u SPEC_PIN bash "$CS" verify-ref "$PIN"
fi
export SPEC_PIN="$PIN"
assert 'verify-ref rejects a branch'                1 bash "$CS" verify-ref 'main'
assert 'verify-ref rejects a tag'                   1 bash "$CS" verify-ref 'v1.0.0-rc.5'
assert 'verify-ref rejects HEAD'                    1 bash "$CS" verify-ref 'HEAD'
assert 'verify-ref rejects a short hash'            1 bash "$CS" verify-ref '00b1688'
assert 'verify-ref rejects uppercase hex'           1 bash "$CS" verify-ref '00B1688A9B2457CA397A0BB550ACF47CAD8EE967'
assert 'verify-ref rejects a placeholder'           1 bash "$CS" verify-ref '<candidate-revision>'
assert 'verify-ref rejects the empty string'        1 bash "$CS" verify-ref ''
assert 'verify-ref rejects the null commit'         1 bash "$CS" verify-ref '0000000000000000000000000000000000000000'
assert 'verify-ref rejects a revision equal to the pin' 1 bash "$CS" verify-ref "$PIN"
assert 'verify-ref accepts a full 40-hex revision'  0 bash "$CS" verify-ref '1234567890abcdef1234567890abcdef12345678'

echo ''
echo '=== candidate-suite.sh record: identity, anti-confusion, evidence wording ==='
FAKEROOT="$WORK/root"
mkdir -p "$FAKEROOT/vectors"
printf '{\n  "protocol_version": "1.0.0-rc.5"\n}\n' >"$FAKEROOT/manifest.json"
printf 'vector-a\n' >"$FAKEROOT/vectors/a.json"
MANIFEST_SHA="$(shasum -a 256 <"$FAKEROOT/manifest.json" 2>/dev/null | awk '{print $1}')"
[ -n "$MANIFEST_SHA" ] || MANIFEST_SHA="$(sha256sum <"$FAKEROOT/manifest.json" | awk '{print $1}')"

assert 'record rejects a nonexistent root'          1 bash "$CS" record "$WORK/absent" "$WORK/ev1"
mkdir -p "$WORK/noman"
assert 'record rejects a root with no manifest.json' 1 bash "$CS" record "$WORK/noman" "$WORK/ev2"
assert 'record rejects a manifest digest mismatch'  1 env CANDIDATE_EXPECTED_MANIFEST_SHA256=deadbeef bash "$CS" record "$FAKEROOT" "$WORK/ev3"
assert 'record rejects a candidate identical to the pin' 1 env PIN_MANIFEST_SHA256="$MANIFEST_SHA" bash "$CS" record "$FAKEROOT" "$WORK/ev4"
assert 'record accepts a distinct candidate'        0 env CANDIDATE_EXPECTED_MANIFEST_SHA256="$MANIFEST_SHA" bash "$CS" record "$FAKEROOT" "$WORK/ev5"

# Git for Windows shasum prefixes the entire output line with `\` when a
# filename needs escaping. This shim emits that form only when given a path,
# proving candidate-suite hashes through stdin and never consumes the escaped
# filename form.
WINDOWS_SHASUM_BIN="$WORK/windows-shasum-bin"
mkdir -p "$WINDOWS_SHASUM_BIN"
cat >"$WINDOWS_SHASUM_BIN/shasum" <<'EOF'
#!/usr/bin/env bash
if [ "$#" -gt 2 ]; then
	printf '\\%s  %s\n' "$SIMULATED_SHA256" "${!#}"
else
	printf '%s  -\n' "$SIMULATED_SHA256"
fi
EOF
chmod +x "$WINDOWS_SHASUM_BIN/shasum"
assert 'record avoids Git for Windows shasum filename escaping' 0 env PATH="$WINDOWS_SHASUM_BIN:$PATH" SIMULATED_SHA256="$MANIFEST_SHA" CANDIDATE_EXPECTED_MANIFEST_SHA256="$MANIFEST_SHA" bash "$CS" record "$FAKEROOT" "$WORK/ev6"

# A malformed hash tool must fail closed rather than persisting a textual
# backslash-prefixed digest as candidate identity.
PREFIXED_SHASUM_BIN="$WORK/prefixed-shasum-bin"
mkdir -p "$PREFIXED_SHASUM_BIN"
cat >"$PREFIXED_SHASUM_BIN/shasum" <<'EOF'
#!/usr/bin/env bash
printf '\\%s  -\n' "$SIMULATED_SHA256"
EOF
chmod +x "$PREFIXED_SHASUM_BIN/shasum"
assert 'record rejects a backslash-prefixed digest' 1 env PATH="$PREFIXED_SHASUM_BIN:$PATH" SIMULATED_SHA256="$MANIFEST_SHA" bash "$CS" record "$FAKEROOT" "$WORK/ev7"
assert_contains 'prefixed digest rejection names non-canonical output' 'hash tool returned a non-canonical sha256 digest' "$WORK/out.txt"

EV="$WORK/ev5/candidate-suite-identity.txt"
if [ -f "$EV" ]; then
	assert_contains 'evidence states it is NOT A RELEASE'    'NOT A RELEASE'         "$EV"
	assert_contains 'evidence claims no release'             'release_claim           none'    "$EV"
	assert_contains 'evidence claims no conformance'         'conformance_claim       none'    "$EV"
	assert_contains 'evidence is stamped candidate-only'     'evidence_class          candidate-only' "$EV"
	assert_contains 'evidence records the manifest digest'   "manifest_sha256         sha256:$MANIFEST_SHA" "$EV"
	assert_contains 'evidence records a tree digest'         'tree_sha256             sha256:'  "$EV"
	assert_contains 'evidence records the file count'        'file_count              2'        "$EV"
	assert_contains 'evidence records the protocol version'  '1.0.0-rc.5'            "$EV"
	assert_contains 'evidence records the committed pin'     "committed_conformance_pin  $PIN"     "$EV"
else
	bad 'record wrote its evidence file' "missing: $EV"
fi
unset SPEC_PIN

echo ''
echo '=== platform-case-gate.sh: the ledger is enforced, not decorated ==='
GATE="$HERE/platform-case-gate.sh"
LEDGER="$WORK/ledger.tsv"
CLASSES="$HERE/skip-classes.tsv"
{
	printf '# package\ttest\tmust_run_on\tskip_allowed_on\tclass\tbehaviour\n'
	printf 'internal/buildcache\tTestWindowsProtectedStateMatrix\twindows\t-\t-\tWindows DACL matrix\n'
	printf 'internal/buildcache\tTestWindowsProtectedStateMatrix/*\t-\twindows\thost-capability\ta symlink subtest may skip\n'
	printf 'internal/shell\tTestPowerShellHookRunsOnEveryPrompt\twindows\tlinux,darwin\tplatform-control\tPowerShell prompt hook\n'
	printf 'internal/godriver\tTestProbeRejectsAnUncoveredPlatformBeforeTheWorker\tlinux,darwin,windows\t-\t-\tuncovered platform rejected\n'
	# A row that TOLERATES a deferred-only skip. It requires the case nowhere,
	# so it constrains nothing else in this file; it exists to prove that
	# tolerating a skip does not repeal the policy of its class.
	printf 'internal/ledgerlisted\tTestRootUnsetIsToleratedHere\t-\tlinux,darwin,windows\troot-unset\ta ledger case whose skip is legitimate only while suite-plan defers it\n'
} >"$LEDGER"

# minimal test2json event builders
ev() { printf '{"Time":"2026-07-29T00:00:00Z","Action":"%s","Package":"github.com/relux-works/curator/%s","Test":"%s"}\n' "$1" "$2" "$3"; }
evout() { printf '{"Time":"2026-07-29T00:00:00Z","Action":"output","Package":"github.com/relux-works/curator/%s","Test":"%s","Output":"    x_test.go:1: %s\\n"}\n' "$1" "$2" "$3"; }

run_gate() {
	stream="$1"; shift
	CI_PLATFORM_CASES="$LEDGER" CI_SKIP_CLASSES="$CLASSES" CI_GATE_MODULE='github.com/relux-works/curator' \
		"$@" bash "$GATE" "$stream" "$WORK/gate-ev"
}

S="$WORK/s.json"
{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a complete windows run passes'              0 run_gate "$S" env CI_GATE_GOOS=windows

{ evout internal/buildcache TestWindowsProtectedStateMatrix 'creating Windows symlink requires host support: nope'
  ev skip internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a required DACL case that skips fails'      1 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a required case that never runs fails'      1 run_gate "$S" env CI_GATE_GOOS=windows

{ ev fail internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a required case that fails, fails the gate' 1 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  evout internal/buildcache 'TestWindowsProtectedStateMatrix/symlink' 'creating Windows symlink requires host support: nope'
  ev skip internal/buildcache 'TestWindowsProtectedStateMatrix/symlink'
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a declared subtest skip is tolerated'       0 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  evout internal/buildcache 'TestWindowsProtectedStateMatrix/symlink' 'this conformance root publishes no symlink cases'
  ev skip internal/buildcache 'TestWindowsProtectedStateMatrix/symlink'
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'a declared subtest skip for the WRONG class fails' 1 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/registry TestSomethingElse 'because I felt like it'
  ev skip internal/registry TestSomethingElse; } >"$S"
assert 'a skip with an unrecognised reason fails'   1 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/registry TestSomethingElse 'symlink unavailable: operation not permitted'
  ev skip internal/registry TestSomethingElse; } >"$S"
assert 'a host-capability skip is recorded and allowed' 0 run_gate "$S" env CI_GATE_GOOS=windows

{ ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/marker TestAuthoritativeThing 'CURATOR_CONFORMANCE_ROOT is not set'
  ev skip internal/marker TestAuthoritativeThing
  evout internal/shell TestPowerShellHookRunsOnEveryPrompt 'PowerShell prompt integration is exercised on Windows'
  ev skip internal/shell TestPowerShellHookRunsOnEveryPrompt; } >"$S"
assert 'a root-unset skip in a SERVED package fails' 1 run_gate "$S" env CI_GATE_GOOS=linux
assert 'a root-unset skip in a DEFERRED package passes' 0 run_gate "$S" env CI_GATE_GOOS=linux CI_DEFERRED_PKGS=internal/marker

# The two cases above go through the class-policy branch because no ledger row
# lists them. A row that TOLERATES the skip must not be able to buy it back: the
# `deferred-only` policy decides whether an unset root is legitimate, and a lane
# that serves the package has already answered no. Without this pair the gate
# tolerates a root-unset skip in a served package for every ledger-listed case
# -- which is every case a candidate lane actually demands.
{ ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/ledgerlisted TestRootUnsetIsToleratedHere 'CURATOR_CONFORMANCE_ROOT is not set'
  ev skip internal/ledgerlisted TestRootUnsetIsToleratedHere; } >"$S"
assert 'a LEDGER-TOLERATED root-unset skip in a SERVED package fails' 1 run_gate "$S" env CI_GATE_GOOS=linux
assert 'the same ledger-tolerated skip passes once the package is deferred' 0 \
	run_gate "$S" env CI_GATE_GOOS=linux CI_DEFERRED_PKGS=internal/ledgerlisted

# ...and the toleration itself still works for a class whose policy is `allow`,
# so the fix above narrows exactly one class and not the ledger as a whole.
{ ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/shell TestPowerShellHookRunsOnEveryPrompt 'PowerShell prompt integration is exercised on Windows'
  ev skip internal/shell TestPowerShellHookRunsOnEveryPrompt
  evout internal/ledgerlisted TestRootUnsetIsToleratedHere 'CURATOR_CONFORMANCE_ROOT is not set'
  ev skip internal/ledgerlisted TestRootUnsetIsToleratedHere; } >"$S"
assert 'an allow-policy tolerated skip survives beside a deferred package' 0 \
	run_gate "$S" env CI_GATE_GOOS=linux CI_DEFERRED_PKGS=internal/ledgerlisted

{ ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  evout internal/shell TestPowerShellHookRunsOnEveryPrompt 'PowerShell prompt integration is exercised on Windows'
  ev skip internal/shell TestPowerShellHookRunsOnEveryPrompt; } >"$S"
assert 'a tolerated platform-control skip passes'   0 run_gate "$S" env CI_GATE_GOOS=linux

{ ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt; } >"$S"
assert 'an excluded package is not demanded'        0 run_gate "$S" env CI_GATE_GOOS=windows CI_EXCLUDED_PKGS=internal/buildcache

# test output that looks like a result event must not be read as one
{ ev pass internal/buildcache TestWindowsProtectedStateMatrix
  ev pass internal/shell TestPowerShellHookRunsOnEveryPrompt
  printf '{"Time":"2026-07-29T00:00:00Z","Action":"output","Package":"github.com/relux-works/curator/internal/godriver","Test":"TestProbeRejectsAnUncoveredPlatformBeforeTheWorker","Output":"    got {\\"Action\\":\\"skip\\",\\"Test\\":\\"TestWindowsProtectedStateMatrix\\"}\\n"}\n'
  ev pass internal/godriver TestProbeRejectsAnUncoveredPlatformBeforeTheWorker; } >"$S"
assert 'output text mimicking a result event is not counted' 0 run_gate "$S" env CI_GATE_GOOS=windows

echo ''
echo '=== platform-case-gate.sh: the SHIPPED ledger is satisfiable on each runner ==='
SHIPPED="${CI_PLATFORM_CASES:-.github/ci/platform-cases.tsv}"
for goos in linux darwin windows; do
	stream="$WORK/shipped-$goos.json"
	excluded=''
	[ "$goos" = linux ] && excluded='internal/godriver'
	awk -F'\t' -v goos="$goos" -v excl="$excluded" '
		function listed(set, want,   n, i, a) {
			if (set == "-" || set == "") return 0
			n = split(set, a, ",")
			for (i = 1; i <= n; i++) if (a[i] == want) return 1
			return 0
		}
		/^[ \t]*#/ || /^[ \t]*$/ { next }
		{
			if ($2 ~ /\/\*$/) next
			if (!listed($3, goos)) next
			if ($1 == excl && $2 != "TestProbeRejectsAnUncoveredPlatformBeforeTheWorker") next
			printf "{\"Action\":\"pass\",\"Package\":\"github.com/relux-works/curator/%s\",\"Test\":\"%s\"}\n", $1, $2
		}
	' "$SHIPPED" >"$stream"
	if [ -s "$stream" ]; then
		CI_SKIP_CLASSES="$CLASSES" CI_GATE_MODULE='github.com/relux-works/curator' \
		assert "the shipped ledger is satisfiable on $goos" 0 \
			env CI_GATE_GOOS="$goos" CI_EXCLUDED_PKGS="$excluded" CI_SKIP_CLASSES="$CLASSES" \
			    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$stream" "$WORK/shipped-ev-$goos"
	else
		bad "the shipped ledger is satisfiable on $goos" 'the ledger required nothing at all on this platform'
	fi
done

echo ''
echo '=== platform-cases.tsv: the packages this pin promoted tolerate no skip ==='
#
# The committed SPEC_PIN publishes every artefact `root-artifacts.tsv` declares
# for these four packages. The previous pin published none of them, so
# `suite-plan.sh` deferred all four and their ledger rows tolerated the
# `root-unset` skip that deferral produced. Against this pin the plan defers
# nothing, so that tolerance can never fire again -- and a tolerance that cannot
# fire is a false ledger row: it states a survivable skip the gate would in fact
# refuse, and it becomes live again, unnoticed, if the pin is ever moved back.
#
# The list is the DELTA this pin move made and not the set of every served
# package. The `root-unset` rows for internal/skillspec, internal/marker,
# internal/moduleroots and internal/scriptpolicy were already unreachable under
# the previous pin, which published their families too; they lost no deferral
# here and are left alone. That bound is stated rather than left to be inferred
# from the absence of a check.
PROMOTED_PACKAGES='internal/config internal/envfragment internal/envmarker internal/interop/environments'
for pkg in $PROMOTED_PACKAGES; do
	rows="$(awk -F'\t' -v p="$pkg" '$1 == p { n++ } END { print n + 0 }' "$SHIPPED")"
	if [ "$rows" -eq 0 ]; then
		bad "$pkg has ledger rows to check" \
			'no row names this package, so the toleration check below would pass vacuously'
		continue
	fi
	tolerating="$(awk -F'\t' -v p="$pkg" '$1 == p && ($4 != "-" || $5 != "-") { printf "%s[skip=%s class=%s] ", $2, $4, $5 }' "$SHIPPED")"
	if [ -n "$tolerating" ]; then
		bad "$pkg tolerates no skip in the shipped ledger ($rows rows)" \
			"the committed pin serves this package, so these rows tolerate a skip that can never legitimately happen: $tolerating"
	else
		ok "$pkg tolerates no skip in the shipped ledger ($rows rows)"
	fi
done

# ...and the removal is behavioural, not cosmetic.
#
# Take the same OTHERWISE-SATISFYING stream the loop above built for each
# runner -- every case the shipped ledger requires there, observed passing --
# and change exactly one event: `TestConformanceContextVersions` skips with the
# unset-root reason, in a lane that DEFERRED its package. That is the one
# combination the old `linux,darwin,windows` / `root-unset` columns made
# survivable, and it is the narrowest weakening of the stream there is.
#
# The exit code alone would not prove this. A stream carrying only the skip
# fails anyway, because every other required case is then missing -- so the
# assertion has to run against a stream that would otherwise pass, and the
# verdict recorded for the skip itself is checked by name.
PROMOTED_CASE=TestConformanceContextVersions
for goos in linux darwin windows; do
	stream="$WORK/promoted-skip-$goos.json"
	excluded=''
	[ "$goos" = linux ] && excluded='internal/godriver'
	awk -F'\t' -v goos="$goos" -v excl="$excluded" -v skipcase="$PROMOTED_CASE" '
		function listed(set, want,   n, i, a) {
			if (set == "-" || set == "") return 0
			n = split(set, a, ",")
			for (i = 1; i <= n; i++) if (a[i] == want) return 1
			return 0
		}
		/^[ \t]*#/ || /^[ \t]*$/ { next }
		{
			if ($2 ~ /\/\*$/) next
			if (!listed($3, goos)) next
			if ($1 == excl && $2 != "TestProbeRejectsAnUncoveredPlatformBeforeTheWorker") next
			if ($2 == skipcase) {
				printf "{\"Action\":\"output\",\"Package\":\"github.com/relux-works/curator/%s\",\"Test\":\"%s\",\"Output\":\"    suite_test.go:1: CURATOR_CONFORMANCE_ROOT is not set\\n\"}\n", $1, $2
				printf "{\"Action\":\"skip\",\"Package\":\"github.com/relux-works/curator/%s\",\"Test\":\"%s\"}\n", $1, $2
				seen = 1
				next
			}
			printf "{\"Action\":\"pass\",\"Package\":\"github.com/relux-works/curator/%s\",\"Test\":\"%s\"}\n", $1, $2
		}
		END { if (!seen) exit 3 }
	' "$SHIPPED" >"$stream"
	if [ $? -ne 0 ]; then
		bad "the shipped ledger still requires $PROMOTED_CASE on $goos" \
			'the case this narrowing is built around is no longer a required row'
		continue
	fi
	evdir="$WORK/promoted-ev-$goos"
	assert "a root-unset skip of $PROMOTED_CASE fails an otherwise-passing $goos stream, even deferred" 1 \
		env CI_GATE_GOOS="$goos" CI_DEFERRED_PKGS=internal/interop/environments \
		    CI_EXCLUDED_PKGS="$excluded" \
		    CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' \
		    bash "$GATE" "$stream" "$evdir"
	if [ -f "$evdir/skips-observed.tsv" ]; then
		assert_contains "the $goos verdict is the ledger refusing it, not a class policy" \
			"$PROMOTED_CASE	root-unset	FATAL-not-tolerated" "$evdir/skips-observed.tsv"
	else
		bad "the $goos run recorded its skip verdict" "missing $evdir/skips-observed.tsv"
	fi
done

echo ''
echo '=== suite-plan.sh: the plan comes from the root, not from the lane ==='
PLAN="$HERE/suite-plan.sh"
assert 'suite-plan rejects a nonexistent root'      2 bash "$PLAN" "$WORK/absent" "$WORK/planev"

if command -v go >/dev/null 2>&1 && go list ./... >/dev/null 2>&1; then
	SERVING="$WORK/serving"; PARTIAL="$WORK/partial"
	for r in "$SERVING" "$PARTIAL"; do
		mkdir -p "$r/vectors" "$r/schema-cases" "$r/expected/build-driver" "$r/fixtures"
		printf '{"protocol_version":"1.0.0-rc.5"}\n' >"$r/manifest.json"
	done
	# a root that serves every declared artefact
	while IFS="$(printf '\t')" read -r _pkg artefacts _note; do
		case "$_pkg" in ''|\#*) continue ;; esac
		oldifs="$IFS"; IFS=','
		for a in $artefacts; do mkdir -p "$SERVING/$(dirname "$a")"; : >"$SERVING/$a"; done
		IFS="$oldifs"
	done <"$HERE/root-artifacts.tsv"
	# ...and one that serves none of them
	printf '{"platforms":[{"name":"linux","status":"excluded","until_task":"T"}]}\n' >"$SERVING/vectors/conformance-claim-v3-qualification.json"

	assert 'a fully serving root defers nothing, even under CI_REQUIRE_FULL_ROOT' 0 \
		env CI_GATE_GOOS=darwin CI_REQUIRE_FULL_ROOT=1 bash "$PLAN" "$SERVING" "$WORK/plan-serving"
	assert 'a partial root defers, and that is fatal under CI_REQUIRE_FULL_ROOT' 1 \
		env CI_GATE_GOOS=darwin CI_REQUIRE_FULL_ROOT=1 bash "$PLAN" "$PARTIAL" "$WORK/plan-partial-strict"
	assert 'a partial root defers and is tolerated without it' 0 \
		env CI_GATE_GOOS=darwin bash "$PLAN" "$PARTIAL" "$WORK/plan-partial"

	if [ -f "$WORK/plan-partial/plan-deferred.txt" ]; then
		assert_contains 'the partial root defers internal/marker' 'internal/marker' "$WORK/plan-partial/plan-deferred.txt"
	else
		bad 'the partial plan wrote a deferred list' 'missing plan-deferred.txt'
	fi

	# --- one missing environments family is fatal in the candidate lane -----
	#
	# `internal/interop/environments` exists so that the environments vector
	# families can be DECLARED without deferring the pre-environments cases in
	# `internal/interop`, which the default lane runs against the committed pin.
	# The declaration is only worth something if EACH declared artefact is
	# load-bearing on its own: a root that publishes five of the six and drops
	# one must fail the candidate lane by name, not skip the family quietly.
	#
	# These cases delete exactly one artefact at a time from the otherwise fully
	# serving root -- the narrowest possible weakening of the root -- and require
	# the plan to name the package AND the missing family every time.
	ENVPKG='internal/interop/environments'
	ENV_ARTEFACTS="$(awk -F'\t' -v p="$ENVPKG" '$1 == p { gsub(/,/, " ", $2); print $2; exit }' "$HERE/root-artifacts.tsv")"

	# The row must declare at least every family the package's own cases require.
	# Reading the required set out of the row instead would make this loop
	# self-referential: shrink the row and the loop shrinks with it, and a
	# dropped family would test nothing while still reporting `ok`.
	ENV_REQUIRED="$(sed -n 's/.*requireFamily(t, root, "\([^"]*\)").*/\1/p' \
		"$ROOTDIR/$ENVPKG"/*_test.go 2>/dev/null | sort -u)"
	# The cross-referenced paths -- the trees and the expected file the VECTORS
	# name, which no requireFamily call mentions -- are deliberately NOT listed
	# here. A third hand-kept copy is a third place to forget, and one was
	# already forgotten that way. They are derived from the vectors themselves by
	# TestConformanceEveryPathTheVectorsNameIsDeclared, and the holing loop below
	# proves each of them is load-bearing for the plan whatever the row says.
	if [ -z "$(printf '%s' "$ENV_REQUIRED" | tr -d ' \n')" ]; then
		bad "$ENVPKG requires at least one environments family" 'no requireFamily call found in the package'
	fi
	for required in $ENV_REQUIRED; do
		case " $ENV_ARTEFACTS " in
		*" $required "*) ok "root-artifacts.tsv declares $required for $ENVPKG" ;;
		*) bad "root-artifacts.tsv declares $required for $ENVPKG" \
			"the package reads it, so a root dropping it must DEFER the package; undeclared, the candidate lane stays green" ;;
		esac
	done

	if [ -z "$ENV_ARTEFACTS" ]; then
		bad "$ENVPKG declares its environments artefacts" "no row for $ENVPKG in root-artifacts.tsv"
	else
		for artefact in $ENV_ARTEFACTS; do
			HOLED="$WORK/holed"
			rm -rf "$HOLED"
			cp -R "$SERVING" "$HOLED"
			rm -rf "${HOLED:?}/${artefact:?}"
			evdir="$WORK/plan-holed"
			rm -rf "$evdir"
			assert "a candidate root missing $artefact fails the lane" 1 \
				env CI_GATE_GOOS=darwin CI_REQUIRE_FULL_ROOT=1 bash "$PLAN" "$HOLED" "$evdir"
			if [ -f "$evdir/suite-plan.txt" ]; then
				assert_contains "the failure names $ENVPKG"  "FAIL  $ENVPKG was deferred" "$evdir/suite-plan.txt"
				assert_contains "the failure names $artefact" "missing: $artefact"          "$evdir/suite-plan.txt"
				assert_contains "internal/interop itself is still served for $artefact" \
					"$MODULE_PATH/internal/interop" "$evdir/plan-served.txt"
			else
				bad "the holed plan wrote its report for $artefact" "missing $evdir/suite-plan.txt"
			fi
		done
	fi

	# The pre-environments consumer must never be dragged into the deferral: it
	# is the package whose cases the default lane runs against the pinned root.
	if grep -qx "$MODULE_PATH/internal/interop" "$WORK/plan-partial/plan-deferred.txt"; then
		bad 'internal/interop is never deferred by a partial root' \
			'the pre-environments conformance cases would stop running on the default lane'
	else
		ok 'internal/interop is never deferred by a partial root'
	fi

	assert 'a root excluding linux excludes godriver there' 0 \
		env CI_GATE_GOOS=linux bash "$PLAN" "$SERVING" "$WORK/plan-linux"
	assert_contains 'godriver is excluded on linux' 'internal/godriver' "$WORK/plan-linux/plan-excluded.txt"
	assert_contains 'the exclusion is asserted by a case' 'TestProbeRejectsAnUncoveredPlatformBeforeTheWorker' "$WORK/plan-linux/plan-assert.txt"

	assert 'the same root does NOT exclude godriver on darwin' 0 \
		env CI_GATE_GOOS=darwin bash "$PLAN" "$SERVING" "$WORK/plan-darwin"
	if [ -s "$WORK/plan-darwin/plan-excluded.txt" ]; then
		bad 'darwin excludes nothing under a linux-only exclusion' "excluded: $(tr '\n' ' ' <"$WORK/plan-darwin/plan-excluded.txt")"
	else
		ok 'darwin excludes nothing under a linux-only exclusion'
	fi

	# a root that predates the qualification vector falls back to the recorded default
	rm -f "$PARTIAL/vectors/conformance-claim-v3-qualification.json"
	assert 'a pre-vector root still excludes godriver on linux' 0 \
		env CI_GATE_GOOS=linux bash "$PLAN" "$PARTIAL" "$WORK/plan-prevector"
	assert_contains 'the fallback exclusion is recorded as such' 'default_excluded_on' "$WORK/plan-prevector/suite-plan.txt"

	echo ''
	echo '=== ledger-consistency.sh: a ledger claim is checked against the real builds ==='
	BADLEDGER="$WORK/bad-ledger.tsv"
	printf 'internal/godriver\tTestThisCaseDoesNotExist\tlinux,darwin,windows\t-\t-\tinvented\n' >"$BADLEDGER"
	assert 'a ledger row naming a nonexistent case fails' 1 \
		env CI_PLATFORM_CASES="$BADLEDGER" bash "$HERE/ledger-consistency.sh" "$WORK/lc-bad"
	printf 'internal/nosuchpackage\tTestX\tlinux\t-\t-\tinvented\n' >"$BADLEDGER"
	assert 'a ledger row naming a nonexistent package fails' 1 \
		env CI_PLATFORM_CASES="$BADLEDGER" bash "$HERE/ledger-consistency.sh" "$WORK/lc-badpkg"
	assert 'the shipped ledger is consistent with every build' 0 \
		env CI_EXCLUDED_PKGS=internal/godriver CI_EXCLUDED_GOOS=linux bash "$HERE/ledger-consistency.sh" "$WORK/lc-shipped"
else
	skip 'suite-plan and ledger-consistency cases' 'no usable `go list` on this runner; these gates need the module graph'
fi

echo ''
echo '=== no-broad-suppression.sh: narrow and named, or not at all ==='
NBS="$HERE/no-broad-suppression.sh"
SRC="$WORK/src"; mkdir -p "$SRC/pkg"
printf 'package pkg\n\nfunc A() {}\n' >"$SRC/pkg/a.go"
CFG_OK="$WORK/ok.yml"
printf 'version: "2"\nlinters:\n  exclusions:\n    rules:\n      - path: _test\\.go\n        linters:\n          - gosec\n' >"$CFG_OK"
assert 'a clean tree passes'                        0 env CI_GOLANGCI_CONFIG="$CFG_OK" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"

printf 'package pkg\n\n//nolint\nfunc B() {}\n' >"$SRC/pkg/b.go"
assert 'a bare //nolint is rejected'                1 env CI_GOLANGCI_CONFIG="$CFG_OK" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"
printf 'package pkg\n\n//nolint:all\nfunc B() {}\n' >"$SRC/pkg/b.go"
assert 'a //nolint:all is rejected'                 1 env CI_GOLANGCI_CONFIG="$CFG_OK" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"
printf 'package pkg\n\n//nolint:gosec // G304: the path is protocol-validated\nfunc B() {}\n' >"$SRC/pkg/b.go"
assert 'a named //nolint with a reason is accepted' 0 env CI_GOLANGCI_CONFIG="$CFG_OK" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"
printf 'package pkg\n\n//#nosec\nfunc B() {}\n' >"$SRC/pkg/b.go"
assert 'a bare //#nosec is rejected'                1 env CI_GOLANGCI_CONFIG="$CFG_OK" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"
rm -f "$SRC/pkg/b.go"

CFG_PROD="$WORK/prod.yml"
printf 'version: "2"\nlinters:\n  exclusions:\n    rules:\n      - path: internal/install\n        linters:\n          - gosec\n' >"$CFG_PROD"
assert 'a production-path lint exclusion is rejected' 1 env CI_GOLANGCI_CONFIG="$CFG_PROD" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"

CFG_OFF="$WORK/off.yml"
printf 'version: "2"\nlinters:\n  default: none\n' >"$CFG_OFF"
assert 'wholesale linter disabling is rejected'     1 env CI_GOLANGCI_CONFIG="$CFG_OFF" CI_GOSEC_ALLOWED='' bash "$NBS" "$SRC"

CFG_SEC="$WORK/sec.yml"
printf 'version: "2"\nlinters:\n  settings:\n    gosec:\n      excludes:\n        - G306\n        - G999\n' >"$CFG_SEC"
assert 'an unrecorded gosec exclusion is rejected'  1 env CI_GOLANGCI_CONFIG="$CFG_SEC" CI_GOSEC_ALLOWED='G306' bash "$NBS" "$SRC"
assert 'a recorded gosec exclusion is accepted'     0 env CI_GOLANGCI_CONFIG="$CFG_SEC" CI_GOSEC_ALLOWED='G306 G999' bash "$NBS" "$SRC"
assert 'the shipped .golangci.yml passes the gate'  0 bash "$NBS"

echo ''
echo '=== toolchain-identity.sh: the toolchain is read back, never assumed ==='
TI="$HERE/toolchain-identity.sh"
STUBROOT="$WORK/goroot"
mkdir -p "$STUBROOT/bin" "$WORK/bin"
make_stub() {
	version="$1"; toolchain="$2"; goenv="$3"
	cat >"$WORK/bin/go" <<STUB
#!/bin/sh
case "\$1" in
  version) echo "go version $version \$(uname -s)/\$(uname -m)" ;;
  env)
    case "\$2" in
      GOROOT) echo '$STUBROOT' ;;
      GOTOOLCHAIN) echo '$toolchain' ;;
      GOENV) echo '$goenv' ;;
      *) echo '' ;;
    esac ;;
  *) exit 0 ;;
esac
STUB
	chmod +x "$WORK/bin/go"
	cp "$WORK/bin/go" "$STUBROOT/bin/go"
	printf '#!/bin/sh\nexit 0\n' >"$WORK/bin/gofmt"; chmod +x "$WORK/bin/gofmt"
	cp "$WORK/bin/gofmt" "$STUBROOT/bin/gofmt"
}
WANT_GO="go$(awk '/^go[ \t]/{print $2; exit}' go.mod)"
make_stub "$WANT_GO" local off
assert 'a healthy toolchain passes'                 0 env PATH="$STUBROOT/bin:$PATH" bash "$TI"
make_stub 'go1.25.1' local off
assert 'a toolchain that is not go.mod'"'"'s version is rejected' 1 env PATH="$STUBROOT/bin:$PATH" bash "$TI"
make_stub "$WANT_GO" auto off
assert 'GOTOOLCHAIN=auto is rejected'               1 env PATH="$STUBROOT/bin:$PATH" bash "$TI"
make_stub "$WANT_GO" local ''
assert 'go 1.25'"'"'s empty GOENV spelling passes'    0 env PATH="$STUBROOT/bin:$PATH" bash "$TI"
make_stub "$WANT_GO" local "$WORK/user.env"
assert 'a per-user go env file is rejected'         1 env PATH="$STUBROOT/bin:$PATH" bash "$TI"

echo ''
echo '=== ci.yml: every test-gate lane pins its per-package timeout ==='

# Every workflow lane must retain the selected deadline. The matrix expression
# covers its Windows value explicitly; fixed-host and race lanes are pinned
# separately. Adding a lane or changing a timeout requires updating these pins.
gate_lane_records() {
	awk '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ {
			job = $1
			sub(/:[ \t\r]*$/, "", job)
		}
		/^      - name:/ {
			step = $0
			sub(/^      - name:[ \t]*/, "", step)
			timeout = ""
		}
		/^[ \t]*GO_TEST_TIMEOUT:/ {
			timeout = $0
			sub(/^[ \t]*GO_TEST_TIMEOUT:[ \t]*/, "", timeout)
			sub(/[ \t\r]*$/, "", timeout)
		}
		/test-gate\.sh/ && /^[ \t]*run:/ {
			printf "%s\t%s\t%s\n", job, step, timeout
		}
	' "$WORKFLOW"
}

lane_records="$(gate_lane_records)"
lane_count="$(printf '%s\n' "$lane_records" | awk 'NF { count++ } END { print count + 0 }')"
if [ "$lane_count" -eq 4 ]; then
	ok 'the workflow has four explicitly budgeted test-gate lanes'
else
	bad 'the workflow has four explicitly budgeted test-gate lanes' \
		"found $lane_count lanes; records: $(printf '%s ' "$lane_records")"
fi

WINDOWS_MATRIX_TIMEOUT="\${{ runner.os == 'Windows' && '120m' || '60m' }}"
WINDOWS_MATRIX_DRIFT="\${{ runner.os == 'Windows' && '90m' || '60m' }}"

observed_lane_timeout() {
	records="$1"; job="$2"
	printf '%s\n' "$records" | awk -F '\t' -v want="$job" \
		'$1 == want { count++; value = $3 } END { if (count == 1) print value }'
}

lane_timeout_matches() {
	[ "$(observed_lane_timeout "$1" "$2")" = "$3" ]
}

workflow_timeouts_match() {
	records="$1"
	count="$(printf '%s\n' "$records" | awk 'NF { count++ } END { print count + 0 }')"
	[ "$count" -eq 4 ] || return 1
	lane_timeout_matches "$records" test "$WINDOWS_MATRIX_TIMEOUT" &&
		lane_timeout_matches "$records" test-self-hosted '60m' &&
		lane_timeout_matches "$records" race '60m' &&
		lane_timeout_matches "$records" candidate-conformance "$WINDOWS_MATRIX_TIMEOUT"
}

check_lane_timeout() {
	job="$1"; expected="$2"
	actual="$(observed_lane_timeout "$lane_records" "$job")"
	if [ "$actual" = "$expected" ]; then
		ok "$job timeout is pinned to $expected"
	else
		bad "$job timeout is pinned to $expected" "observed: ${actual:-missing or duplicate lane}"
	fi
}

if workflow_timeouts_match "$lane_records"; then
	ok 'the self-test enforces all four workflow lane timeout pins'
else
	bad 'the self-test enforces all four workflow lane timeout pins' 'one or more lane expressions drifted'
fi
check_lane_timeout test "$WINDOWS_MATRIX_TIMEOUT"
check_lane_timeout test-self-hosted '60m'
check_lane_timeout race '60m'
# The candidate matrix pin includes its Windows value as well as Ubuntu/macOS.
check_lane_timeout candidate-conformance "$WINDOWS_MATRIX_TIMEOUT"

mutate_timeout_lane() {
	target_job="$1"; output="$2"; replacement="$3"
	awk -v target="$target_job" -v replacement="$replacement" '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ {
			job = $1
			sub(/:[ \t\r]*$/, "", job)
		}
		job == target && /^[ \t]*GO_TEST_TIMEOUT:/ {
			line = $0
			sub(/GO_TEST_TIMEOUT:.*/, "", line)
			print line "GO_TEST_TIMEOUT: " replacement
			changed++
			next
		}
		{ print }
		END { if (changed != 1) exit 2 }
	' "$WORKFLOW" >"$output"
}

check_timeout_drift_rejected() {
	name="$1"; job="$2"; expected="$3"; replacement="$4"
	mutant="$WORK/timeout-mutant-$job.yml"
	if ! mutate_timeout_lane "$job" "$mutant" "$replacement"; then
		bad "$name" "could not produce a single-lane $job mutant"
		return
	fi
	original_workflow="$WORKFLOW"
	WORKFLOW="$mutant"
	mutant_records="$(gate_lane_records)"
	WORKFLOW="$original_workflow"
	if ! workflow_timeouts_match "$mutant_records" && \
		! lane_timeout_matches "$mutant_records" "$job" "$expected"; then
		ok "$name"
	else
		bad "$name" "the changed $job timeout remained admitted"
	fi
}

check_timeout_drift_rejected 'the self-test rejects Test timeout drift' test "$WINDOWS_MATRIX_TIMEOUT" '30m'
check_timeout_drift_rejected 'the self-test rejects self-hosted Test timeout drift' test-self-hosted '60m' '30m'
check_timeout_drift_rejected 'the self-test rejects Race timeout drift' race '60m' '30m'
check_timeout_drift_rejected 'the self-test rejects candidate timeout drift' candidate-conformance "$WINDOWS_MATRIX_TIMEOUT" '30m'
check_timeout_drift_rejected 'the self-test rejects Windows candidate budget drift' \
	candidate-conformance "$WINDOWS_MATRIX_TIMEOUT" "$WINDOWS_MATRIX_DRIFT"

# The Test (rose-air) lane is pinned to the rose-air runner by label: without
# it any organisation self-hosted Apple Silicon runner can take the job.
self_hosted_runs_on() {
	awk '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ {
			job = $1
			sub(/:[ \t\r]*$/, "", job)
		}
		job == "test-self-hosted" && /^    runs-on:/ {
			line = $0
			sub(/^    runs-on:[ \t]*/, "", line)
			sub(/[ \t\r]*$/, "", line)
			count++; value = line
		}
		END { if (count == 1) print value }
	' "$1"
}

runs_on_pins_rose_air() {
	printf '%s\n' "$(self_hosted_runs_on "$1")" |
		awk '/^\[/ && /\]$/ { gsub(/[][ \t]/, ""); n = split($0, l, ","); for (i = 1; i <= n; i++) if (l[i] == "rose-air") found = 1 } END { exit !found }'
}

if runs_on_pins_rose_air "$WORKFLOW"; then
	ok 'test-self-hosted runs-on is pinned to the rose-air label'
else
	bad 'test-self-hosted runs-on is pinned to the rose-air label' \
		"observed runs-on: $(self_hosted_runs_on "$WORKFLOW")"
fi
sed 's/^    runs-on: \[self-hosted, macOS, ARM64, rose-air\]/    runs-on: [self-hosted, macOS, ARM64]/' \
	"$WORKFLOW" >"$WORK/runs-on-unpinned.yml"
if cmp -s "$WORKFLOW" "$WORK/runs-on-unpinned.yml"; then
	bad 'the self-test rejects an unpinned self-hosted runs-on' 'could not produce the unpinned mutant'
elif runs_on_pins_rose_air "$WORK/runs-on-unpinned.yml"; then
	bad 'the self-test rejects an unpinned self-hosted runs-on' 'the unpinned runs-on remained admitted'
else
	ok 'the self-test rejects an unpinned self-hosted runs-on'
fi

GATE_DEFAULT_TIMEOUT="$(awk -F ':-' '/^GO_TEST_TIMEOUT=/{value=$2; sub(/}.*/, "", value); print value; exit}' "$HERE/test-gate.sh")"
if [ "$GATE_DEFAULT_TIMEOUT" = '60m' ]; then
	ok 'test-gate.sh defaults to the 60m per-package budget'
else
	bad 'test-gate.sh defaults to the 60m per-package budget' "observed: ${GATE_DEFAULT_TIMEOUT:-missing}"
fi
MAKEFILE_DEFAULT_TIMEOUT="$(awk '/^GO_TEST_TIMEOUT[ \t]*\?=/{print $3; exit}' Makefile)"
if [ "$MAKEFILE_DEFAULT_TIMEOUT" = '60m' ]; then
	ok 'Makefile CI gate targets default to the 60m per-package budget'
else
	bad 'Makefile CI gate targets default to the 60m per-package budget' "observed: ${MAKEFILE_DEFAULT_TIMEOUT:-missing}"
fi

echo ''
echo '=== test-gate.sh: a synthetic hanging package still fails at its deadline ==='
if command -v go >/dev/null 2>&1; then
	SYNTHETIC_MODULE="$MODULE_PATH"
	SYNTHETIC_PACKAGE="$MODULE_PATH/internal/gatefixture"
	SYNTHETIC_FIXTURE="$WORK/gate-timeout-fixture"
	SYNTHETIC_ROOT="$WORK/gate-timeout-root"
	SYNTHETIC_GO="$WORK/gate-timeout-go"
	SYNTHETIC_CAPTURE="$WORK/gate-timeout-captured-budget.txt"
	SYNTHETIC_EVIDENCE="$WORK/gate-timeout-evidence"
	# The workflow-level self-test jobs do not install the repository toolchain,
	# and GOTOOLCHAIN=local intentionally forbids downloading it. Give this
	# standalone fixture the running launcher version so a lagging hosted image
	# still reaches the real test timeout instead of refusing its go.mod.
	REAL_GO="$(command -v go)"
	MODULE_GO_VERSION="$("$REAL_GO" env GOVERSION)"
	MODULE_GO_VERSION="${MODULE_GO_VERSION#go}"
	mkdir -p "$SYNTHETIC_FIXTURE" "$SYNTHETIC_ROOT"
	printf 'module %s\n\ngo %s\n' "$SYNTHETIC_PACKAGE" "$MODULE_GO_VERSION" >"$SYNTHETIC_FIXTURE/go.mod"
	cat >"$SYNTHETIC_FIXTURE/hang_test.go" <<'GO'
package gatefixture

import "testing"

func TestSyntheticHang(t *testing.T) {
	select {}
}
GO
	printf '{"synthetic":"timeout-test"}\n' >"$SYNTHETIC_ROOT/manifest.json"
	printf 'internal/gatefixture\tTestSyntheticHang\tlinux\t-\t-\tgate timeout negative case\n' >"$WORK/gate-timeout-ledger.tsv"
	cat >"$SYNTHETIC_GO" <<'GO'
#!/usr/bin/env bash
set -u
case "${1:-}" in
list)
	printf '%s\n' "${SYNTHETIC_PACKAGE:?}"
	;;
test)
	timeout=''
	while [ "$#" -gt 0 ]; do
		if [ "$1" = '-timeout' ]; then
			[ "$#" -ge 2 ] || exit 2
			timeout="$2"
			shift 2
			continue
		fi
		shift
	done
	[ -n "$timeout" ] || { echo 'synthetic go: missing -timeout' >&2; exit 2; }
	printf '%s\n' "$timeout" >"${SYNTHETIC_CAPTURE:?}"
	printf 'GIT_CONFIG_GLOBAL=%s\nGIT_CONFIG_NOSYSTEM=%s\n' "${GIT_CONFIG_GLOBAL-<unset>}" \
		"${GIT_CONFIG_NOSYSTEM-<unset>}" >"${SYNTHETIC_CAPTURE:?}.gitenv"
	cd "${SYNTHETIC_FIXTURE:?}" || exit 2
	exec "${REAL_GO:?}" test -json -count=1 -timeout "$timeout" .
	;;
*)
	exec "${REAL_GO:?}" "$@"
	;;
esac
GO
	chmod +x "$SYNTHETIC_GO"

	# The ambient git config is hostile and reached the way a runner's copied
	# ~/.gitconfig is: through the default global lookup, with neither variable
	# set. Only the gate's own export can put them into go test's environment.
	HOSTILE_XDG="$WORK/hostile-xdg"
	HOSTILE_GITCONFIG="$HOSTILE_XDG/git/config"
	mkdir -p "$HOSTILE_XDG/git"
	printf '[commit]\n\tgpgsign = true\n[tag]\n\tgpgsign = true\n' >"$HOSTILE_GITCONFIG"
	HANG_STARTED="$SECONDS"
	env -u GIT_CONFIG_GLOBAL -u GIT_CONFIG_NOSYSTEM \
		XDG_CONFIG_HOME="$HOSTILE_XDG" \
		GO="$SYNTHETIC_GO" REAL_GO="$REAL_GO" \
		SYNTHETIC_MODULE="$SYNTHETIC_MODULE" SYNTHETIC_PACKAGE="$SYNTHETIC_PACKAGE" \
		SYNTHETIC_FIXTURE="$SYNTHETIC_FIXTURE" \
		SYNTHETIC_CAPTURE="$SYNTHETIC_CAPTURE" GO_TEST_TIMEOUT=2s \
		CURATOR_CONFORMANCE_ROOT="$SYNTHETIC_ROOT" CI_GATE_GOOS=linux \
		CI_GATE_MODULE="$MODULE_PATH" CI_PLATFORM_CASES="$WORK/gate-timeout-ledger.tsv" \
		CI_SKIP_CLASSES="$HERE/skip-classes.tsv" \
		bash "$HERE/test-gate.sh" "$SYNTHETIC_EVIDENCE" >"$WORK/gate-timeout.out" 2>&1
	HANG_RC=$?
	HANG_ELAPSED=$((SECONDS - HANG_STARTED))
	if [ "$HANG_RC" -eq 1 ]; then
		PASS=$((PASS + 1)); printf 'ok    synthetic package hang fails test-gate (exit=%s)\n' "$HANG_RC"
	else
		bad 'synthetic package hang fails test-gate (exit=1)' "real exit=$HANG_RC"
	fi
	if [ -f "$SYNTHETIC_CAPTURE" ] && [ "$(cat "$SYNTHETIC_CAPTURE")" = '2s' ]; then
		ok 'test-gate forwards the configured 2s package deadline to go test'
	else
		bad 'test-gate forwards the configured 2s package deadline to go test' 'captured timeout missing or changed'
	fi
	GITENV_CAPTURE="$SYNTHETIC_CAPTURE.gitenv"
	captured_global="$(sed -n 's/^GIT_CONFIG_GLOBAL=//p' "$GITENV_CAPTURE" 2>/dev/null)"
	if [ "$(sed -n 's/^GIT_CONFIG_NOSYSTEM=//p' "$GITENV_CAPTURE" 2>/dev/null)" = '1' ]; then
		ok 'test-gate exports GIT_CONFIG_NOSYSTEM=1 to go test'
	else
		bad 'test-gate exports GIT_CONFIG_NOSYSTEM=1 to go test' \
			"observed: $(cat "$GITENV_CAPTURE" 2>/dev/null || echo 'no capture')"
	fi
	if [ -n "$captured_global" ] && [ "$captured_global" != "$HOSTILE_GITCONFIG" ] && \
		[ -f "$captured_global" ] && [ ! -s "$captured_global" ]; then
		ok 'test-gate exports an empty gate-owned GIT_CONFIG_GLOBAL over a hostile ambient config'
	else
		bad 'test-gate exports an empty gate-owned GIT_CONFIG_GLOBAL over a hostile ambient config' \
			"observed: ${captured_global:-missing}"
	fi
	if grep -qF "test-gate: GIT_CONFIG_GLOBAL=$captured_global GIT_CONFIG_NOSYSTEM=1" "$WORK/gate-timeout.out"; then
		ok 'test-gate logs the isolated git config it exports'
	else
		bad 'test-gate logs the isolated git config it exports' 'log line missing'
	fi
	if [ -f "$SYNTHETIC_EVIDENCE/go-test.json" ] && \
		grep -qF 'panic: test timed out after 2s' "$SYNTHETIC_EVIDENCE/go-test.json"; then
		ok 'the synthetic Go test fails because its package deadline expires'
	else
		observed="$(awk 'NR <= 8 { print substr($0, 1, 180) }' \
			"$SYNTHETIC_EVIDENCE/go-test.json" 2>/dev/null)"
		gate_output="$(awk 'NR <= 8 { print substr($0, 1, 180) }' "$WORK/gate-timeout.out" 2>/dev/null)"
		bad 'the synthetic Go test fails because its package deadline expires' \
			"timed-out test output not found; Go stream: ${observed:-missing}; test-gate: ${gate_output:-missing}"
	fi
	ok "synthetic gate failure elapsed ${HANG_ELAPSED}s including compilation"
else
	skip 'synthetic package timeout negative case' 'no usable Go launcher on this runner'
fi

echo ''
echo '=== pnpm-pin-guard.sh: the workflow copy of the pin cannot drift ==='
PPG="$HERE/pnpm-pin-guard.sh"
WF_PIN="$(awk '/^[ \t]*PNPM_PIN:[ \t]*/{gsub(/"/, "", $2); print $2; exit}' "$WORKFLOW")"
if [ -z "$WF_PIN" ]; then
	bad 'the workflow declares a PNPM_PIN' 'no PNPM_PIN: line in .github/workflows/ci.yml'
	WF_PIN='0.0.0-drift-probe'
fi
assert 'the workflow PNPM_PIN agrees with SupportedPNPMVersion' 0 env PNPM_PIN="$WF_PIN" bash "$PPG"
assert 'a drifted workflow pin is rejected' 1 env PNPM_PIN='9.9.9-drift-probe' bash "$PPG"
assert 'an unset pin fails closed' 1 env -u PNPM_PIN bash "$PPG"
assert 'an unreadable Go source fails closed' 1 env PNPM_PIN="$WF_PIN" CI_PNPM_GO_SOURCE="$WORK/absent.go" bash "$PPG"
printf 'package pnpmsource\n\nconst SomethingElse = "1.0.0"\n' >"$WORK/no-const.go"
assert 'a Go source with no declaration fails closed' 1 env PNPM_PIN="$WF_PIN" CI_PNPM_GO_SOURCE="$WORK/no-const.go" bash "$PPG"

echo ''
echo '=== ci.yml: every suite lane verifies and installs the pinned pnpm first ==='

# A lane that runs test-gate.sh runs the TestRealPinnedPNPM* cases, so it
# must verify the workflow pin against the Go constant, then install
# exactly that release with npm into a lane-local prefix, then prepend
# that prefix to PATH -- guard, then install, then PATH, in that order,
# before any test runs. A lane that drops the guard can drift; a lane that
# drops the install silently reverts to skipping the real-pnpm cases
# (hosted: pnpm absent) or to resolving an ambient shim (self-hosted:
# macbook-iv run 35072267145); a lane that installs without prepending
# leaves the ambient shim first. These cases read the wiring back, so a
# refactor cannot drop a step without this self-test noticing.
#
# Emits one "<job> <guard?> <install?> <path?> <ordered?>" record per
# test-gate.sh lane.
pnpm_lane_records() {
	awk '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ { job = $1; sub(/:[ \t\r]*$/, "", job); next }
		/^[ \t]*run:.*test-gate\.sh/      { gate[job] = 1; next }
		/^[ \t]*run:.*pnpm-pin-guard\.sh/ { guard[job] = 1; guardline[job] = NR; next }
		/npm install -g --prefix .*pnpm-prefix/ { install[job] = 1; installline[job] = NR; next }
		/pnpm-prefix.*GITHUB_PATH/        { path[job] = 1; pathline[job] = NR; next }
		END {
			for (j in gate) {
				ordered = (guard[j] && install[j] && path[j] && \
				           guardline[j] < installline[j] && installline[j] < pathline[j]) ? 1 : 0
				printf "%s %d %d %d %d\n", j, (guard[j] ? 1 : 0), (install[j] ? 1 : 0), \
					(path[j] ? 1 : 0), ordered
			}
		}
	' "$WORKFLOW"
}

lane_records="$(pnpm_lane_records)"
if [ -z "$lane_records" ]; then
	bad 'the workflow still runs test-gate.sh' 'no test-gate.sh step found in .github/workflows/ci.yml'
else
	unprovisioned=''
	while read -r job has_guard has_install has_path ordered; do
		[ -n "$job" ] || continue
		if [ "$has_guard" != '1' ] || [ "$has_install" != '1' ] || \
		   [ "$has_path" != '1' ] || [ "$ordered" != '1' ]; then
			unprovisioned="$unprovisioned $job(guard=$has_guard,install=$has_install,path=$has_path,ordered=$ordered)"
		fi
	done <<-RECORDS
	$lane_records
	RECORDS
	if [ -z "$unprovisioned" ]; then
		ok 'every test-gate.sh lane verifies the pin, then installs it via npm first on PATH'
	else
		bad 'every test-gate.sh lane verifies the pin, then installs it via npm first on PATH' \
			"lanes missing the guard, the install, or the order:$unprovisioned"
	fi
fi

# The pin lives in exactly one place. A lane that hardcodes pnpm@<release>
# instead of reading ${{ env.PNPM_PIN }} reinstalls the drift the guard
# exists to prevent.
if grep -nE 'pnpm@[0-9]' "$WORKFLOW" >"$WORK/out.txt"; then
	bad 'no lane hardcodes a pnpm release' "$(tr '\n' ' ' <"$WORK/out.txt" | cut -c1-200)"
else
	ok 'no lane hardcodes a pnpm release; every install reads PNPM_PIN'
fi

echo ''
echo '=== platform-cases.tsv: the real-pnpm cases are required on windows with no deferral ==='
#
# BUG-260916-2f3xbf declared pnpm's Windows junction spelling of the
# writable-store registry and the materialized links, so the two install cases
# that gate run 35098955988 deferred are required on windows-latest exactly as
# on unix. These cases read the shipped files back, so a resurrected skip, a
# narrowed column, or a reworded row fails here instead of silently changing
# what windows-latest proves.
PNPM_GO='internal/pnpmsource/conformance_test.go'
PNPM_CASES='TestRealPinnedPNPMLockSupersetSnapshotDependencies TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization'

if grep -q 'BUG-260916-2f3xbf\|skipOnWindowsForStoreRegistryGap' "$PNPM_GO"; then
	bad 'the Go suite carries no Windows pnpm deferral' "deferral text survives in $PNPM_GO"
else
	ok 'the Go suite carries no Windows pnpm deferral'
fi

if grep -v '^[ \t]*#' "$CLASSES" | grep -q 'BUG-260916-2f3xbf'; then
	bad 'the class table carries no bug-naming deferral row' "deferral row survives in $CLASSES"
else
	ok 'the class table carries no bug-naming deferral row'
fi

for pcase in $PNPM_CASES; do
	row="$(awk -F'\t' -v c="$pcase" '$1 == "internal/pnpmsource" && $2 == c {print $3"|"$4"|"$5}' "$SHIPPED")"
	if [ "$row" = 'linux,darwin,windows|-|-' ]; then
		ok "the ledger requires $pcase on every runner and tolerates no skip"
	else
		bad "the ledger requires $pcase on every runner and tolerates no skip" "row: ${row:-<missing>}"
	fi
done

# Behavioural: the shipped streams the satisfiability loop above built are the
# otherwise-passing runs; narrow exactly the two install cases inside them.
# The gate counts a pass as satisfied, so the pass events are removed first
# and the verdicts below rest on the skip alone.
if [ -f "$WORK/shipped-windows.json" ] && [ -f "$WORK/shipped-linux.json" ]; then
	win_stream="$WORK/pnpm-windows-skip.json"
	grep -v -e TestRealPinnedPNPMLockSupersetSnapshotDependencies -e TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization \
		"$WORK/shipped-windows.json" >"$win_stream"
	for pcase in $PNPM_CASES; do
		evout internal/pnpmsource "$pcase" 'pinned pnpm executable unavailable' >>"$win_stream"
		ev skip internal/pnpmsource "$pcase" >>"$win_stream"
	done
	assert 'a pnpm-absent skip of an install case fails on windows' 1 \
		env CI_GATE_GOOS=windows CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$win_stream" "$WORK/pnpm-ev-win"
	if [ -f "$WORK/pnpm-ev-win/skips-observed.tsv" ]; then
		for pcase in $PNPM_CASES; do
			assert_contains "windows records $pcase as ledger-refused, not tolerated" \
				"$(printf '%s\thost-capability\tFATAL-not-tolerated' "$pcase")" "$WORK/pnpm-ev-win/skips-observed.tsv"
		done
	else
		bad 'the windows run recorded its skip verdicts' 'missing $WORK/pnpm-ev-win/skips-observed.tsv'
	fi

	lin_stream="$WORK/pnpm-linux-skip.json"
	grep -v -e TestRealPinnedPNPMLockSupersetSnapshotDependencies -e TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization \
		"$WORK/shipped-linux.json" >"$lin_stream"
	for pcase in $PNPM_CASES; do
		evout internal/pnpmsource "$pcase" 'pinned pnpm executable unavailable' >>"$lin_stream"
		ev skip internal/pnpmsource "$pcase" >>"$lin_stream"
	done
	assert 'the same skip fails on linux too' 1 \
		env CI_GATE_GOOS=linux CI_EXCLUDED_PKGS=internal/godriver \
		    CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$lin_stream" "$WORK/pnpm-ev-lin"
else
	bad 'the pnpm behavioural cases have shipped streams to narrow' 'missing $WORK/shipped-windows.json or $WORK/shipped-linux.json'
fi

echo ''
echo '=== rust-pin-guard.sh: the committed toolchain file cannot drift ==='
RPG="$HERE/rust-pin-guard.sh"
assert 'the committed rust-toolchain.toml agrees with SupportedRustToolchainVersion' 0 bash "$RPG"
printf '[toolchain]\nchannel = "1.92.0"\nprofile = "minimal"\n' >"$WORK/rust-bumped.toml"
assert 'a bumped toolchain file is rejected' 1 env CI_RUST_TOOLCHAIN_FILE="$WORK/rust-bumped.toml" bash "$RPG"
printf 'package rustsource\n\nconst SupportedRustToolchainVersion = "1.92.0"\n' >"$WORK/rust-const-bumped.go"
assert 'a changed supported value is rejected' 1 env CI_RUST_GO_SOURCE="$WORK/rust-const-bumped.go" bash "$RPG"
assert 'a missing toolchain file fails closed' 1 env CI_RUST_TOOLCHAIN_FILE="$WORK/absent.toml" bash "$RPG"
printf '[toolchain]\nprofile = "minimal"\n' >"$WORK/rust-no-channel.toml"
assert 'a toolchain file with no channel fails closed' 1 env CI_RUST_TOOLCHAIN_FILE="$WORK/rust-no-channel.toml" bash "$RPG"
printf '[toolchain]\nchannel = "stable"\nprofile = "minimal"\n' >"$WORK/rust-stable.toml"
assert 'a floating channel is rejected' 1 env CI_RUST_TOOLCHAIN_FILE="$WORK/rust-stable.toml" bash "$RPG"
printf '[toolchain]\nchannel = "1.91.0"\nprofile = "default"\n' >"$WORK/rust-default-profile.toml"
assert 'a non-minimal profile is rejected' 1 env CI_RUST_TOOLCHAIN_FILE="$WORK/rust-default-profile.toml" bash "$RPG"
assert 'an unreadable Go source fails closed' 1 env CI_RUST_GO_SOURCE="$WORK/absent.go" bash "$RPG"
printf 'package rustsource\n\nconst SomethingElse = "1.0.0"\n' >"$WORK/rust-no-const.go"
assert 'a Go source with no declaration fails closed' 1 env CI_RUST_GO_SOURCE="$WORK/rust-no-const.go" bash "$RPG"

echo ''
echo '=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ==='
IRS="$HERE/install-rust-toolchain.sh"
BASH_ABS="$(command -v bash)"

# A fake rustup records its install invocation; fake shims satisfy the
# post-install PATH check. Nothing here touches the network.
FAKEBIN="$WORK/fake-rust-bin"
mkdir -p "$FAKEBIN"
cat >"$FAKEBIN/rustup" <<'EOF'
#!/usr/bin/env bash
echo "rustup $*" >>"$FAKE_RUSTUP_LOG"
if [ "$1" = "show" ]; then
	echo "1.92.0-test (overridden by fake rust-toolchain.toml)"
elif [ "$1" = "which" ]; then
	if [ -n "${FAKE_PINNED_RUSTC:-}" ]; then
		echo "$FAKE_PINNED_RUSTC"
	else
		echo "rustup which fallback is unavailable in this fixture" >&2
		exit 1
	fi
fi
EOF
chmod +x "$FAKEBIN/rustup"
printf '#!/usr/bin/env bash\necho "rustc fake: $0"\n' >"$FAKEBIN/rustc"
printf '#!/usr/bin/env bash\necho "cargo fake: $0"\n' >"$FAKEBIN/cargo"
chmod +x "$FAKEBIN/rustc" "$FAKEBIN/cargo"

printf '[toolchain]\nchannel = "1.92.0"\nprofile = "minimal"\n' >"$WORK/rust-install-channel.toml"
: >"$WORK/rustup-log.txt"; : >"$WORK/github-path.txt"
assert 'the installer installs exactly the filed channel' 0 \
	env PATH="$FAKEBIN:$PATH" FAKE_RUSTUP_LOG="$WORK/rustup-log.txt" GITHUB_PATH="$WORK/github-path.txt" \
	    CARGO_HOME="$WORK/fake-cargo" CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" bash "$IRS"
assert_contains 'the install invocation names the filed channel' 'toolchain install 1.92.0 --profile minimal' "$WORK/rustup-log.txt"
assert_contains 'the shim directory is prepended for the rest of the lane' "$WORK/fake-cargo/bin" "$WORK/github-path.txt"

# A PATH with no rustup anywhere on it: the hosted images ship rustup, so
# the negative case must hide the real one rather than assume its absence.
# Entries are filtered, never replaced wholesale: a bare single-directory
# PATH breaks tool startup on Git Bash for Windows (awk died with exit 127
# loading its DLLs), failing the row for the wrong reason. Every entry stays
# except ones shipping a rustup executable, so the installer still parses
# the channel file and then fails exactly on the missing rustup.
NORUSTPATH=""
OLDIFS="$IFS"; IFS=':'
# shellcheck disable=SC2086
for _entry in $PATH; do
	[ -n "$_entry" ] || _entry='.'
	if [ -x "$_entry/rustup" ] || [ -x "$_entry/rustup.exe" ]; then
		continue
	fi
	if [ -z "$NORUSTPATH" ]; then
		NORUSTPATH="$_entry"
	else
		NORUSTPATH="$NORUSTPATH:$_entry"
	fi
done
IFS="$OLDIFS"

# For the Homebrew-keg fixture, also remove any preinstalled rustc/cargo
# directories so only its explicit keg proxies can satisfy the PATH checks.
NO_RUST_TOOLS_PATH=""
OLDIFS="$IFS"; IFS=':'
# shellcheck disable=SC2086
for _entry in $PATH; do
	[ -n "$_entry" ] || _entry='.'
	if [ -x "$_entry/rustup" ] || [ -x "$_entry/rustup.exe" ] || \
	   [ -x "$_entry/rustc" ] || [ -x "$_entry/rustc.exe" ] || \
	   [ -x "$_entry/cargo" ] || [ -x "$_entry/cargo.exe" ]; then
		continue
	fi
	if [ -z "$NO_RUST_TOOLS_PATH" ]; then
		NO_RUST_TOOLS_PATH="$_entry"
	else
		NO_RUST_TOOLS_PATH="$NO_RUST_TOOLS_PATH:$_entry"
	fi
done
IFS="$OLDIFS"

# rustup present ONLY under CARGO_HOME/bin: the self-hosted runner service
# starts from launchd with a minimal PATH and the lane shell reads no
# profiles, so a per-user rustup (~/.cargo/bin, --no-modify-path) is
# invisible until the installer prepends it. RUSTUP_HOME points elsewhere
# on purpose: resolution must not depend on it.
CARGOBIN="$WORK/fake-cargo-only/bin"
mkdir -p "$CARGOBIN" "$WORK/empty-rustup-home"
cp "$FAKEBIN/rustup" "$FAKEBIN/rustc" "$FAKEBIN/cargo" "$CARGOBIN/"
: >"$WORK/rustup-log-cargo.txt"; : >"$WORK/github-path-cargo.txt"
assert 'the installer finds rustup under CARGO_HOME/bin without PATH help' 0 \
	env PATH="$NORUSTPATH" FAKE_RUSTUP_LOG="$WORK/rustup-log-cargo.txt" GITHUB_PATH="$WORK/github-path-cargo.txt" \
	    CARGO_HOME="$WORK/fake-cargo-only" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
assert_contains 'the CARGO_HOME-only install names the filed channel' 'toolchain install 1.92.0 --profile minimal' "$WORK/rustup-log-cargo.txt"
assert_contains 'the CARGO_HOME bin dir is recorded for the rest of the lane' "$WORK/fake-cargo-only/bin" "$WORK/github-path-cargo.txt"
assert_contains 'the CARGO_HOME-only success still names its rustup' 'rust-pin: using rustup at' "$WORK/out.txt"
if [ "$(cat "$WORK/github-path-cargo.txt")" = "$WORK/fake-cargo-only/bin" ]; then
	ok 'the CARGO_HOME-only GITHUB_PATH write is exactly the shim dir'
else
	bad 'the CARGO_HOME-only GITHUB_PATH write is exactly the shim dir' \
		"got: $(tr '\n' ' ' <"$WORK/github-path-cargo.txt" | cut -c1-200)"
fi
if grep -q 'runner diagnostics:' "$WORK/out.txt"; then
	bad 'the CARGO_HOME-only success prints no diagnostics block' "$(tr '\n' ' ' <"$WORK/out.txt" | cut -c1-200)"
else
	ok 'the CARGO_HOME-only success prints no diagnostics block'
fi

# Row (b), TASK-260929-2wgjam: rustup in CARGO_HOME/bin with a go elsewhere
# on PATH -- unchanged behaviour, and go still resolves to the original.
GOONLYBIN="$WORK/fake-go-only/bin"
mkdir -p "$GOONLYBIN"
printf '#!/usr/bin/env bash\necho "go fake: $0"\n' >"$GOONLYBIN/go"
chmod +x "$GOONLYBIN/go"
: >"$WORK/github-path-cargo-b.txt"
assert 'row (b): rustup in CARGO_HOME/bin installs with go on PATH' 0 \
	env PATH="$GOONLYBIN:$NORUSTPATH" FAKE_RUSTUP_LOG="$WORK/rustup-log-cargo-b.txt" GITHUB_PATH="$WORK/github-path-cargo-b.txt" \
	    CARGO_HOME="$WORK/fake-cargo-only" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
if [ "$(cat "$WORK/github-path-cargo-b.txt")" = "$WORK/fake-cargo-only/bin" ]; then
	ok 'row (b): the GITHUB_PATH write is exactly CARGO_HOME/bin'
else
	bad 'row (b): the GITHUB_PATH write is exactly CARGO_HOME/bin' "got: $(tr '\n' ' ' <"$WORK/github-path-cargo-b.txt" | cut -c1-200)"
fi

# Row (c): a directory the script would add (CARGO_HOME/bin) holds a
# different go -> the script refuses, naming the shadowing directory, and
# records nothing on GITHUB_PATH.
SHADOWCARGO="$WORK/fake-cargo-shadow"
mkdir -p "$SHADOWCARGO/bin"
cp "$FAKEBIN/rustup" "$FAKEBIN/rustc" "$FAKEBIN/cargo" "$SHADOWCARGO/bin/"
printf '#!/usr/bin/env bash\necho "shadowing go: $0"\n' >"$SHADOWCARGO/bin/go"
chmod +x "$SHADOWCARGO/bin/go"
: >"$WORK/github-path-shadow.txt"
assert 'row (c): a Rust directory holding a different go is refused' 1 \
	env PATH="$GOONLYBIN:$NORUSTPATH" FAKE_RUSTUP_LOG="$WORK/rustup-log-shadow.txt" GITHUB_PATH="$WORK/github-path-shadow.txt" \
	    CARGO_HOME="$SHADOWCARGO" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
assert_contains 'row (c): the refusal names the shadowing directory' "rust-pin: refusing to add $SHADOWCARGO/bin to PATH: it would shadow go" "$WORK/out.txt"
if [ -s "$WORK/github-path-shadow.txt" ]; then
	bad 'row (c): the refusal records nothing on GITHUB_PATH' "got: $(tr '\n' ' ' <"$WORK/github-path-shadow.txt" | cut -c1-200)"
else
	ok 'row (c): the refusal records nothing on GITHUB_PATH'
fi

# Simulate the Homebrew rustup formula: `rustup` is linked into the prefix's
# bin directory, but its rustc/cargo proxies exist only beside the real
# executable in the versioned keg. Git Bash on Windows does not guarantee
# real symlink support, and Homebrew's layout is POSIX-only.
case "$(uname -s)" in
	MINGW*|MSYS*|CYGWIN*)
		skip 'Homebrew rustup keg fixture and mutant' 'Homebrew layout is POSIX-only; Windows Git Bash does not guarantee real symlinks'
		;;
	*)
BREWBIN="$WORK/fake-brew/bin"
BREWKEGBIN="$WORK/fake-brew/Cellar/rustup/1.28.2/bin"
BREWCARGO="$WORK/fake-cargo-brew/bin"
mkdir -p "$BREWBIN" "$BREWKEGBIN" "$BREWCARGO"
cp "$FAKEBIN/rustup" "$BREWKEGBIN/rustup"
cp "$FAKEBIN/rustc" "$FAKEBIN/cargo" "$BREWKEGBIN/"
ln -s ../Cellar/rustup/1.28.2/bin/rustup "$BREWBIN/rustup"
BREWKEG_REALBIN="$(cd -P "$BREWKEGBIN" && pwd)"
: >"$WORK/rustup-log-brew.txt"; : >"$WORK/github-path-brew.txt"
assert 'the Homebrew fixture links rustup from bin into its versioned keg' 0 \
	bash -c 'test -L "$1/rustup" && test "$(readlink "$1/rustup")" = ../Cellar/rustup/1.28.2/bin/rustup' _ "$BREWBIN"
assert 'the Homebrew fixture keeps compiler proxies only in the keg' 0 \
	bash -c 'test -x "$1/rustc" && test -x "$1/cargo" && test ! -e "$2/rustc" && test ! -e "$2/cargo" && test ! -e "$3/rustc" && test ! -e "$3/cargo"' _ "$BREWKEGBIN" "$BREWBIN" "$BREWCARGO"
assert 'the installer finds rustc and cargo from a Homebrew rustup keg' 0 \
	env PATH="$NO_RUST_TOOLS_PATH" FAKE_RUSTUP_LOG="$WORK/rustup-log-brew.txt" GITHUB_PATH="$WORK/github-path-brew.txt" \
	    CARGO_HOME="$WORK/fake-cargo-brew" HOMEBREW_PREFIX="$WORK/fake-brew" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
assert_contains 'the Homebrew-prefix install names the filed channel' 'toolchain install 1.92.0 --profile minimal' "$WORK/rustup-log-brew.txt"
assert_contains 'the rustc invocation resolves to the keg proxy' "rustc fake: $BREWKEG_REALBIN/rustc" "$WORK/out.txt"
assert_contains 'the cargo invocation resolves to the keg proxy' "cargo fake: $BREWKEG_REALBIN/cargo" "$WORK/out.txt"
assert_contains 'the CARGO_HOME bin dir (proxies) is still recorded alongside it' "$WORK/fake-cargo-brew/bin" "$WORK/github-path-brew.txt"
assert_contains 'the Homebrew keg proxy dir is recorded for the rest of the lane' "$BREWKEG_REALBIN" "$WORK/github-path-brew.txt"
_brew_want="$(printf '%s\n%s' "$WORK/fake-cargo-brew/bin" "$BREWKEG_REALBIN")"
if [ "$(cat "$WORK/github-path-brew.txt")" = "$_brew_want" ]; then
	ok 'the Homebrew GITHUB_PATH writes are CARGO_HOME then keg proxies, never the prefix bin'
else
	bad 'the Homebrew GITHUB_PATH writes are CARGO_HOME then keg proxies, never the prefix bin' \
		"got: $(tr '\n' ' ' <"$WORK/github-path-brew.txt" | cut -c1-200)"
fi
if grep -q 'runner diagnostics:' "$WORK/out.txt"; then
	bad 'the Homebrew-keg success prints no diagnostics block' "$(tr '\n' ' ' <"$WORK/out.txt" | cut -c1-200)"
else
	ok 'the Homebrew-keg success prints no diagnostics block'
fi

# Narrowing mutant: remove the single production step that adds the resolved
# Homebrew proxy directory to PATH/GITHUB_PATH. The same fixture must reject
# it at the compiler check; its rustup-which fallback is unavailable by
# design, so this specifically proves the keg-proxy row drives that step.
_keg_step_count="$(grep -cF 'prepend_lane_path "$rustup_proxy_dir"' "$IRS")"
_keg_mutant="$WORK/install-rust-without-keg-proxies.sh"
awk 'index($0, "prepend_lane_path \"$rustup_proxy_dir\"") { next } { print }' "$IRS" >"$_keg_mutant"
_keg_mutant_count="$(grep -cF 'prepend_lane_path "$rustup_proxy_dir"' "$_keg_mutant" || true)"
if [ "$_keg_step_count" -eq 1 ] && [ "$_keg_mutant_count" -eq 0 ]; then
	ok 'the Homebrew keg mutant removes exactly the keg PATH step'
else
	bad 'the Homebrew keg mutant removes exactly the keg PATH step' \
		"production occurrences: $_keg_step_count (want 1), mutant: $_keg_mutant_count (want 0)"
fi
if "$BASH_ABS" -n "$_keg_mutant"; then
	ok 'the Homebrew keg mutant is syntactically valid'
else
	bad 'the Homebrew keg mutant is syntactically valid' 'bash -n failed; the kill below would prove nothing'
fi
assert 'the Homebrew keg self-test rejects a missing keg PATH step' 1 \
	env PATH="$NO_RUST_TOOLS_PATH" FAKE_RUSTUP_LOG="$WORK/rustup-log-brew-mutant.txt" GITHUB_PATH="$WORK/github-path-brew-mutant.txt" \
	    CARGO_HOME="$WORK/fake-cargo-brew" HOMEBREW_PREFIX="$WORK/fake-brew" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$_keg_mutant"
assert_contains 'the Homebrew keg mutant fails at the pinned rustc check' \
	'rust-pin: rustc is not on PATH after installing Rust 1.92.0' "$WORK/out.txt"

# Row (a), TASK-260929-2wgjam: the shared-prefix shape of the second
# rose-air runner. rustup is a symlink in a prefix bin that also holds a
# different `go`; the lane's original `go` must still resolve afterwards and
# the prefix bin must never reach GITHUB_PATH.
SHAREDPFX="$WORK/fake-shared"
SHAREDKEG="$SHAREDPFX/Cellar/rustup/1.28.2/bin"
ORIGGOBIN="$WORK/fake-setup-go/bin"
mkdir -p "$SHAREDPFX/bin" "$SHAREDKEG" "$ORIGGOBIN" "$WORK/fake-cargo-shared/bin"
cp "$FAKEBIN/rustup" "$FAKEBIN/rustc" "$FAKEBIN/cargo" "$SHAREDKEG/"
ln -s ../Cellar/rustup/1.28.2/bin/rustup "$SHAREDPFX/bin/rustup"
printf '#!/usr/bin/env bash\necho "go fake: $0"\n' >"$ORIGGOBIN/go"
printf '#!/usr/bin/env bash\necho "shadowing go: $0"\n' >"$SHAREDPFX/bin/go"
chmod +x "$ORIGGOBIN/go" "$SHAREDPFX/bin/go"
SHARED_REALKEG="$(cd -P "$SHAREDKEG" && pwd)"
# The lane resolves rustup through the shared prefix (as a launchd PATH
# with the runner's .path would), and the setup-go bin leads PATH.
cat >"$WORK/rust-shared-probe.sh" <<'EOF'
#!/usr/bin/env bash
set -e
bash "$IRS_UNDER_TEST"
# Replay the lane's later steps: GITHUB_PATH entries lead PATH.
while IFS= read -r _d; do PATH="$_d:$PATH"; done <"$GITHUB_PATH"
echo "later-step go: $(command -v go)"
EOF
: >"$WORK/github-path-shared.txt"
assert 'row (a): a rustup symlinked from a shared prefix with go installs' 0 \
	env PATH="$ORIGGOBIN:$SHAREDPFX/bin:$NO_RUST_TOOLS_PATH" IRS_UNDER_TEST="$IRS" \
	    FAKE_RUSTUP_LOG="$WORK/rustup-log-shared.txt" GITHUB_PATH="$WORK/github-path-shared.txt" \
	    CARGO_HOME="$WORK/fake-cargo-shared" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$WORK/rust-shared-probe.sh"
assert_contains 'row (a): later steps still resolve the original go' "later-step go: $ORIGGOBIN/go" "$WORK/out.txt"
if grep -qxF "$SHAREDPFX/bin" "$WORK/github-path-shared.txt"; then
	bad 'row (a): GITHUB_PATH never records the shared prefix bin' "got: $(tr '\n' ' ' <"$WORK/github-path-shared.txt" | cut -c1-200)"
else
	ok 'row (a): GITHUB_PATH never records the shared prefix bin'
fi
assert_contains 'row (a): the keg proxy dir is recorded instead' "$SHARED_REALKEG" "$WORK/github-path-shared.txt"

# Mutant: restore the old dirname(rustup) prepend. Row (a) must reject it
# (real exit 1 from the probe: the prepend puts the shadowing go first).
_old_mutant="$WORK/install-rust-old-prepend.sh"
awk '{ print } /^echo "rust-pin: using rustup at \$rustup_bin"$/ && !done { print "printf '"'"'%s\\n'"'"' \"$rustup_dir\" >>\"$GITHUB_PATH\"; export PATH=\"$rustup_dir:$PATH\""; done=1 }' "$IRS" >"$_old_mutant"
if [ "$(grep -c 'export PATH="$rustup_dir:$PATH"' "$_old_mutant")" -eq 1 ] && "$BASH_ABS" -n "$_old_mutant"; then
	ok 'the old-prepend mutant restores exactly the dirname(rustup) prepend'
else
	bad 'the old-prepend mutant restores exactly the dirname(rustup) prepend' 'mutant not built; the kill below would prove nothing'
fi
: >"$WORK/github-path-shared-mutant.txt"
assert 'row (a) kills the old dirname(rustup) prepend mutant' 1 \
	env PATH="$ORIGGOBIN:$SHAREDPFX/bin:$NO_RUST_TOOLS_PATH" IRS_UNDER_TEST="$_old_mutant" \
	    FAKE_RUSTUP_LOG="$WORK/rustup-log-shared-mutant.txt" GITHUB_PATH="$WORK/github-path-shared-mutant.txt" \
	    CARGO_HOME="$WORK/fake-cargo-shared" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$WORK/rust-shared-probe.sh"
assert_contains 'the old-prepend mutant is caught by the resolution invariant' 'the Rust lane changed go resolution' "$WORK/out.txt"
;;
esac

# If the resolved rustup directory does not contain proxies, rustup's exact
# pinned `which` result supplies the toolchain bin directory instead.
FALLBACKBREW="$WORK/fake-brew-fallback"
FALLBACKKEGBIN="$FALLBACKBREW/Cellar/rustup/1.28.2/bin"
FALLBACKBIN="$WORK/fake-pinned-toolchain/bin"
mkdir -p "$FALLBACKBREW/bin" "$FALLBACKKEGBIN" "$FALLBACKBIN"
cp "$FAKEBIN/rustup" "$FALLBACKKEGBIN/rustup"
cp "$FAKEBIN/rustc" "$FAKEBIN/cargo" "$FALLBACKBIN/"
ln -s ../Cellar/rustup/1.28.2/bin/rustup "$FALLBACKBREW/bin/rustup"
: >"$WORK/rustup-log-fallback.txt"; : >"$WORK/github-path-fallback.txt"
assert 'the pinned rustup-which fallback finds rustc and cargo' 0 \
	env PATH="$NO_RUST_TOOLS_PATH" FAKE_PINNED_RUSTC="$FALLBACKBIN/rustc" \
	    FAKE_RUSTUP_LOG="$WORK/rustup-log-fallback.txt" GITHUB_PATH="$WORK/github-path-fallback.txt" \
	    CARGO_HOME="$WORK/fake-cargo-brew" HOMEBREW_PREFIX="$FALLBACKBREW" RUSTUP_HOME="$WORK/empty-rustup-home" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
assert_contains 'the fallback asks rustup for the filed channel compiler' \
	'which --toolchain 1.92.0 rustc' "$WORK/rustup-log-fallback.txt"
assert_contains 'the fallback compiler resolves from its pinned toolchain bin' \
	"rustc fake: $FALLBACKBIN/rustc" "$WORK/out.txt"
assert_contains 'the fallback cargo resolves from its pinned toolchain bin' \
	"cargo fake: $FALLBACKBIN/cargo" "$WORK/out.txt"
# Where ln -s makes a real symlink the middle entry is the keg; where Git
# Bash copies instead, rustup is a plain file in the prefix bin and the
# middle entry is the lane-private link directory. Never the prefix bin.
if [ -L "$FALLBACKBREW/bin/rustup" ]; then
	_fallback_mid="$(cd -P "$FALLBACKKEGBIN" && pwd)"
else
	_fallback_mid="$(sed -n 2p "$WORK/github-path-fallback.txt")"
	case "$_fallback_mid" in */rust-lane-bin.*) ;; *) _fallback_mid="<lane-private rust-lane-bin dir>" ;; esac
fi
_fallback_want="$(printf '%s\n%s\n%s' "$WORK/fake-cargo-brew/bin" "$_fallback_mid" "$FALLBACKBIN")"
if [ "$(cat "$WORK/github-path-fallback.txt")" = "$_fallback_want" ]; then
	ok 'the pinned fallback GITHUB_PATH writes CARGO_HOME, keg, then toolchain bin'
else
	bad 'the pinned fallback GITHUB_PATH writes CARGO_HOME, keg, then toolchain bin' \
		"got: $(tr '\n' ' ' <"$WORK/github-path-fallback.txt" | cut -c1-200)"
fi

# Absent from PATH, CARGO_HOME/bin and the Homebrew prefix: CARGO_HOME and
# HOMEBREW_PREFIX point at empty prefixes so neither a real ~/.cargo/bin nor
# a real Homebrew rustup on this host can leak into the row. The fixed
# /opt/homebrew/bin and /usr/local/bin probes are hidden the same way as
# PATH entries: the row is skipped (named) on a host that ships rustup
# there, since the script cannot be told to ignore them.
mkdir -p "$WORK/empty-cargo/bin" "$WORK/empty-brew/bin"
# A present-but-not-executable rustup under the Homebrew prefix: resolution
# ([ -x ]) skips it, so the row still fails, and the diagnostics must name
# the middle state exactly.
: >"$WORK/empty-brew/bin/rustup"
chmod -x "$WORK/empty-brew/bin/rustup" 2>/dev/null || true
if [ -x /opt/homebrew/bin/rustup ] || [ -x /usr/local/bin/rustup ]; then
	skip 'a runner without rustup fails' 'this host ships rustup under /opt/homebrew/bin or /usr/local/bin, which the script probes unconditionally; the row runs on the hosted lanes'
	skip 'rustup-absent diagnostics rows' 'same host rustup; the block, the defaulted-CARGO_HOME fixture and the narrowing mutant run on the hosted lanes'
else
	assert 'a runner without rustup fails' 1 \
		env PATH="$NORUSTPATH" GITHUB_PATH="$WORK/github-path.txt" CARGO_HOME="$WORK/empty-cargo" HOMEBREW_PREFIX="$WORK/empty-brew" \
		    RUNNER_NAME='fake-rose-air' RUNNER_OS='fake-macOS' \
		    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
	assert_contains 'the failure names the runner-setup note' 'docs/self-hosted-runner-setup.md' "$WORK/out.txt"
	assert_contains 'the failure prints the runner diagnostics block' 'rust-pin: rustup not found; runner diagnostics:' "$WORK/out.txt"
	assert_contains 'the failure names RUNNER_NAME' 'rust-pin:   RUNNER_NAME=fake-rose-air' "$WORK/out.txt"
	assert_contains 'the failure names RUNNER_OS' 'rust-pin:   RUNNER_OS=fake-macOS' "$WORK/out.txt"
	assert_contains 'the failure names hostname' 'rust-pin:   hostname=' "$WORK/out.txt"
	assert_contains 'the failure names whoami' 'rust-pin:   whoami=' "$WORK/out.txt"
	assert_contains 'the failure names HOME' 'rust-pin:   HOME=' "$WORK/out.txt"
	assert_contains 'the failure names a set CARGO_HOME' "rust-pin:   CARGO_HOME=$WORK/empty-cargo (set)" "$WORK/out.txt"
	assert_contains 'the failure names HOMEBREW_PREFIX' "rust-pin:   HOMEBREW_PREFIX=$WORK/empty-brew" "$WORK/out.txt"
	assert_contains 'the failure names the searched PATH' 'rust-pin:   PATH=' "$WORK/out.txt"
	assert_contains 'the failure lists the CARGO_HOME candidate as absent' "rust-pin:   candidate $WORK/empty-cargo/bin/rustup: absent" "$WORK/out.txt"
	assert_contains 'the failure lists the Homebrew-prefix candidate as not executable' "rust-pin:   candidate $WORK/empty-brew/bin/rustup: exists-not-executable" "$WORK/out.txt"
	assert_contains 'the failure lists the Apple-silicon Homebrew candidate' 'rust-pin:   candidate /opt/homebrew/bin/rustup:' "$WORK/out.txt"
	assert_contains 'the failure lists the Intel Homebrew candidate' 'rust-pin:   candidate /usr/local/bin/rustup:' "$WORK/out.txt"
	assert_contains 'the failure lists the empty CARGO_HOME bin dir' "rust-pin:   listing $WORK/empty-cargo/bin: no names containing rust or cargo" "$WORK/out.txt"
	assert_contains 'the failure lists the Apple-silicon Homebrew bin dir' 'rust-pin:   listing /opt/homebrew/bin' "$WORK/out.txt"
	assert_contains 'the failure lists the Intel Homebrew bin dir' 'rust-pin:   listing /usr/local/bin' "$WORK/out.txt"
	assert_contains 'the failure reports command -v rustup' 'rust-pin:   command -v rustup:' "$WORK/out.txt"
	assert_contains 'the failure reports type -a rustup' 'rust-pin:   type -a rustup:' "$WORK/out.txt"
	_diag_last="$(tail -n 1 "$WORK/out.txt")"
	if [ "$_diag_last" = 'rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane' ]; then
		ok 'the remedy sentence is still the last line of the failure'
	else
		bad 'the remedy sentence is still the last line of the failure' "last line: $(printf '%s' "$_diag_last" | cut -c1-200)"
	fi

	# CARGO_HOME unset and HOMEBREW_PREFIX unset: HOME points at an empty
	# fixture home so the defaulted $HOME/.cargo/bin resolves inside the
	# fixture, and the conditional Homebrew-prefix candidate drops out of
	# both the resolution and the enumeration -- three candidate lines, not
	# four. A diagnostics enumeration that hardcoded all four would fail
	# the count below.
	mkdir -p "$WORK/empty-home"
	assert 'a runner without rustup and without CARGO_HOME fails' 1 \
		env -u CARGO_HOME -u HOMEBREW_PREFIX -u RUNNER_NAME -u RUNNER_OS \
		    PATH="$NORUSTPATH" HOME="$WORK/empty-home" GITHUB_PATH="$WORK/github-path.txt" \
		    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
	assert_contains 'the failure names a defaulted CARGO_HOME' "rust-pin:   CARGO_HOME=$WORK/empty-home/.cargo (defaulted)" "$WORK/out.txt"
	assert_contains 'the failure names an unset RUNNER_NAME' 'rust-pin:   RUNNER_NAME=unset' "$WORK/out.txt"
	assert_contains 'the failure names an unset HOMEBREW_PREFIX' 'rust-pin:   HOMEBREW_PREFIX=unset' "$WORK/out.txt"
	assert_contains 'the failure lists the defaulted candidate as absent' "rust-pin:   candidate $WORK/empty-home/.cargo/bin/rustup: absent" "$WORK/out.txt"
	_diag_count="$(grep -c 'rust-pin:   candidate ' "$WORK/out.txt")"
	if [ "$_diag_count" -eq 3 ]; then
		ok 'the failure lists exactly the three probed candidates when HOMEBREW_PREFIX is unset'
	else
		bad 'the failure lists exactly the three probed candidates when HOMEBREW_PREFIX is unset' "candidate lines: $_diag_count, want 3"
	fi

	# Bounded listing: 25 rust-matching names under CARGO_HOME/bin must
	# surface at most 20 entry lines. The row still fails -- none of the
	# 25 is named rustup -- with the remedy sentence last.
	mkdir -p "$WORK/many-cargo/bin"
	_seed=1
	while [ "$_seed" -le 25 ]; do
		: >"$WORK/many-cargo/bin/rust-tool-$(printf '%02d' "$_seed")"
		_seed=$((_seed + 1))
	done
	assert 'a runner with many rust-named files still fails' 1 \
		env PATH="$NORUSTPATH" GITHUB_PATH="$WORK/github-path.txt" CARGO_HOME="$WORK/many-cargo" HOMEBREW_PREFIX="$WORK/empty-brew" \
		    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$IRS"
	assert_contains 'the failure lists the seeded CARGO_HOME bin dir' "rust-pin:   listing $WORK/many-cargo/bin:" "$WORK/out.txt"
	_many_count="$(grep -c 'rust-tool-' "$WORK/out.txt")"
	if [ "$_many_count" -eq 20 ]; then
		ok 'the bounded listing shows exactly 20 of the 25 rust-named entries'
	else
		bad 'the bounded listing shows exactly 20 of the 25 rust-named entries' "entry lines: $_many_count, want 20"
	fi
	_many_last="$(tail -n 1 "$WORK/out.txt")"
	if [ "$_many_last" = 'rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane' ]; then
		ok 'the remedy sentence is still the last line after the bounded listing'
	else
		bad 'the remedy sentence is still the last line after the bounded listing' "last line: $(printf '%s' "$_many_last" | cut -c1-200)"
	fi

	# Narrowing mutant: drop the /opt/homebrew/bin candidate from the
	# DIAGNOSTIC enumeration only (production call site:
	# print_rustup_diagnostics in .github/ci/install-rust-toolchain.sh,
	# reached from the `Install pinned Rust toolchain via rustup` lane
	# step in .github/workflows/ci.yml). A middle line of the loop is
	# dropped so the mutant stays syntactically valid -- dropping the
	# last word would take `; do` with it and fail for the wrong reason.
	# The mutant must still exit 1 -- resolution is untouched -- and the
	# named Apple-silicon-candidate row above must fail against it,
	# proving the row kills the narrowing rather than passing vacuously.
	_diag_prod_count="$(grep -c '/opt/homebrew/bin/rustup' "$IRS")"
	_mutant="$WORK/install-rust-mutant.sh"
	awk 'BEGIN{in_fn=0} /^print_rustup_diagnostics\(\)/{in_fn=1} in_fn && /^}/{in_fn=0} in_fn && /\/opt\/homebrew\/bin\/rustup/{next} {print}' "$IRS" >"$_mutant"
	_diag_mutant_count="$(grep -c '/opt/homebrew/bin/rustup' "$_mutant")"
	if [ "$_diag_prod_count" -eq 2 ] && [ "$_diag_mutant_count" -eq 1 ]; then
		ok 'the narrowing mutant drops exactly the diagnostic Apple-silicon candidate'
	else
		bad 'the narrowing mutant drops exactly the diagnostic Apple-silicon candidate' \
			"production occurrences: $_diag_prod_count (want 2), mutant: $_diag_mutant_count (want 1)"
	fi
	if "$BASH_ABS" -n "$_mutant"; then
		ok 'the narrowing mutant is syntactically valid'
	else
		bad 'the narrowing mutant is syntactically valid' 'bash -n failed; the kill below would prove nothing'
	fi
	env PATH="$NORUSTPATH" GITHUB_PATH="$WORK/github-path.txt" CARGO_HOME="$WORK/empty-cargo" HOMEBREW_PREFIX="$WORK/empty-brew" \
	    CI_RUST_TOOLCHAIN_FILE="$WORK/rust-install-channel.toml" "$BASH_ABS" "$_mutant" >"$WORK/mutant-out.txt" 2>&1
	_mutant_got=$?
	if [ "$_mutant_got" -eq 1 ]; then
		ok 'the narrowed mutant still exits 1'
	else
		bad 'the narrowed mutant still exits 1' "exit $_mutant_got, want 1"
	fi
	if grep -qF 'rust-pin:   candidate /opt/homebrew/bin/rustup:' "$WORK/mutant-out.txt"; then
		bad 'the named Apple-silicon-candidate row kills the narrowing mutant' \
			'the /opt/homebrew/bin candidate line still appears in the mutant output, so the row would pass vacuously'
	else
		ok 'the named Apple-silicon-candidate row kills the narrowing mutant'
	fi
	assert_contains 'the narrowed mutant keeps its other candidate lines' 'rust-pin:   candidate /usr/local/bin/rustup:' "$WORK/mutant-out.txt"
fi

if grep -nE '[0-9]+\.[0-9]+\.[0-9]+' "$IRS" >"$WORK/out.txt"; then
	bad 'the installer names no release; it reads the channel from the file' "$(tr '\n' ' ' <"$WORK/out.txt" | cut -c1-200)"
else
	ok 'the installer names no release; it reads the channel from the file'
fi
if grep -qF 'CI_RUST_TOOLCHAIN_FILE:-rust-toolchain.toml' "$IRS"; then
	ok 'the installer reads its channel from rust-toolchain.toml'
else
	bad 'the installer reads its channel from rust-toolchain.toml' 'install-rust-toolchain.sh does not reference rust-toolchain.toml'
fi

echo ''
echo '=== ci.yml: every suite lane verifies and installs the pinned Rust toolchain first ==='

# A lane that runs test-gate.sh runs the rustsource production cases, so it
# must verify the toolchain file against the Go constant, then install
# exactly that channel with rustup and prepend the shims to PATH -- guard,
# then install, in that order, before any test runs. A lane that drops the
# guard can drift; a lane that drops the install fails the production cases
# on rustc-not-found (rose-air run 35121791685). These cases read the wiring
# back, so a refactor cannot drop a step without this self-test noticing.
#
# Emits one "<job> <guard?> <install?> <ordered?>" record per test-gate.sh
# lane.
rust_lane_records() {
	awk '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ { job = $1; sub(/:[ \t\r]*$/, "", job); next }
		/^[ \t]*run:.*test-gate\.sh/      { gate[job] = 1; gateline[job] = NR; next }
		/^[ \t]*run:.*rust-pin-guard\.sh/ { guard[job] = 1; guardline[job] = NR; next }
		/^[ \t]*run:.*install-rust-toolchain\.sh/ { install[job] = 1; installline[job] = NR; next }
		END {
			for (j in gate) {
				ordered = (guard[j] && install[j] && \
				           guardline[j] < installline[j] && installline[j] < gateline[j]) ? 1 : 0
				printf "%s %d %d %d\n", j, (guard[j] ? 1 : 0), (install[j] ? 1 : 0), ordered
			}
		}
	' "$WORKFLOW"
}

rust_records="$(rust_lane_records)"
if [ -z "$rust_records" ]; then
	bad 'the workflow still runs test-gate.sh' 'no test-gate.sh step found in .github/workflows/ci.yml'
else
	rust_unprovisioned=''
	while read -r job has_guard has_install ordered; do
		[ -n "$job" ] || continue
		if [ "$has_guard" != '1' ] || [ "$has_install" != '1' ] || [ "$ordered" != '1' ]; then
			rust_unprovisioned="$rust_unprovisioned $job(guard=$has_guard,install=$has_install,ordered=$ordered)"
		fi
	done <<-RECORDS
	$rust_records
	RECORDS
	if [ -z "$rust_unprovisioned" ]; then
		ok 'every test-gate.sh lane verifies the rust pin, then installs it via rustup before go test'
	else
		bad 'every test-gate.sh lane verifies the rust pin, then installs it via rustup before go test' \
			"lanes missing the guard, the install, or the order:$rust_unprovisioned"
	fi
fi

# The install lives in exactly one place. A lane that spells out its own
# `rustup toolchain install` reinstalls the drift the guard exists to
# prevent.
if grep -nF 'rustup toolchain install' "$WORKFLOW" >"$WORK/out.txt"; then
	bad 'no lane spells its own rustup install; every lane uses install-rust-toolchain.sh' "$(tr '\n' ' ' <"$WORK/out.txt" | cut -c1-200)"
else
	ok 'no lane spells its own rustup install; every lane uses install-rust-toolchain.sh'
fi

echo ''
echo '=== platform-cases.tsv: the production rust cases run on darwin, tolerate host absence elsewhere ==='
RUST_CASES='TestProductionManagerCapturesRegistryFromRawPaths TestProductionManagerCapturesGitWithoutCallerProjection'
for rcase in $RUST_CASES; do
	row="$(awk -F'\t' -v c="$rcase" '$1 == "internal/rustsource" && $2 == c { print }' "$SHIPPED")"
	if [ -z "$row" ]; then
		bad "$rcase has a ledger row" 'no internal/rustsource row names this case'
		continue
	fi
	must="$(printf '%s' "$row" | awk -F'\t' '{print $3}')"
	skip="$(printf '%s' "$row" | awk -F'\t' '{print $4}')"
	cls="$(printf '%s' "$row" | awk -F'\t' '{print $5}')"
	if [ "$must" = "darwin" ] && [ "$skip" = "linux,windows" ] && [ "$cls" = "host-capability" ]; then
		ok "$rcase is required on darwin and tolerates host-capability absence on linux,windows"
	else
		bad "$rcase is required on darwin and tolerates host-capability absence on linux,windows" \
			"must_run_on=$must skip_allowed_on=$skip class=$cls"
	fi
done

if [ -f "$WORK/shipped-darwin.json" ] && [ -f "$WORK/shipped-linux.json" ]; then
	# A skip on darwin is fatal even though the same reason is tolerated
	# on linux: the lane installed the toolchain, so absence there means
	# the install regressed, not a platform gap.
	rust_dar_stream="$WORK/rust-darwin-skip.json"
	cp "$WORK/shipped-darwin.json" "$rust_dar_stream"
	for rcase in $RUST_CASES; do
		evout internal/rustsource "$rcase" 'pinned Cargo toolchain root or executable unavailable for native target aarch64-apple-darwin' >>"$rust_dar_stream"
		ev skip internal/rustsource "$rcase" >>"$rust_dar_stream"
	done
	assert 'a toolchain-absent skip of a production rust case fails on darwin' 1 \
		env CI_GATE_GOOS=darwin CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$rust_dar_stream" "$WORK/rust-ev-darwin"

	rust_lin_stream="$WORK/rust-linux-skip.json"
	cp "$WORK/shipped-linux.json" "$rust_lin_stream"
	for rcase in $RUST_CASES; do
		evout internal/rustsource "$rcase" 'no operator-approved Cargo descriptor for native target x86_64-unknown-linux-gnu' >>"$rust_lin_stream"
		ev skip internal/rustsource "$rcase" >>"$rust_lin_stream"
	done
	assert 'the same absence skip of a production rust case passes on linux' 0 \
		env CI_GATE_GOOS=linux CI_EXCLUDED_PKGS=internal/godriver \
		    CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$rust_lin_stream" "$WORK/rust-ev-linux"

	# A different reason on linux is fatal too: the ledger pins the class,
	# so an unrelated skip cannot masquerade as the toolchain absence.
	rust_cls_stream="$WORK/rust-linux-wrongclass.json"
	cp "$WORK/shipped-linux.json" "$rust_cls_stream"
	for rcase in $RUST_CASES; do
		evout internal/rustsource "$rcase" 'CURATOR_CONFORMANCE_ROOT is not set' >>"$rust_cls_stream"
		ev skip internal/rustsource "$rcase" >>"$rust_cls_stream"
	done
	assert 'a non-absence skip of a production rust case fails on linux' 1 \
		env CI_GATE_GOOS=linux CI_EXCLUDED_PKGS=internal/godriver \
		    CI_PLATFORM_CASES="$SHIPPED" CI_SKIP_CLASSES="$CLASSES" \
		    CI_GATE_MODULE='github.com/relux-works/curator' bash "$GATE" "$rust_cls_stream" "$WORK/rust-ev-linux-cls"
else
	bad 'the rust behavioural cases have shipped streams to narrow' 'missing $WORK/shipped-darwin.json or $WORK/shipped-linux.json'
fi

echo ''
echo '=== goreleaser rc channel values: the lint lane runs the Go gate ==='
# rev4: the rev1-rev3 awk gate (.github/ci/goreleaser-config-gate.sh) is
# deleted -- a text walk cannot tell a block scalar body or a
# first-position nested map from a real entry key (rev3 R1), and it
# admitted duplicate keys yaml.v3 rejects (rev3 R2). tools/goreleaserconfig
# (gopkg.in/yaml.v3) enforces the values; it runs in the lint job on every
# push and in every go list ./... lane. This job has no setup-go step, so
# this section pins what it can without Go: the awk gate stays deleted,
# and ci.yml structurally carries exactly one live lint step running the
# gate. The value negatives live as executed rows in
# tools/goreleaserconfig/gate_test.go.
if [ -e "$HERE/goreleaser-config-gate.sh" ]; then
	bad 'the awk config gate stays deleted (values are enforced by tools/goreleaserconfig)' '.github/ci/goreleaser-config-gate.sh exists'
else
	ok 'the awk config gate stays deleted (values are enforced by tools/goreleaserconfig)'
fi

# Structural wiring pin (rev3 R3): the old substring grep counted a
# commented-out run line and an if:false step as live invocations. This pin
# walks the workflow structure -- jobs/lint/steps, one live step whose
# parsed run value is the gate command and which carries no if key -- with
# stdlib-only python3, present on all three runners. Quoting is YAML
# syntax, so a quoted run value still counts; duplicate run keys in one
# step disqualify it.
GRW_PIN="$WORK/grw-pin.py"
cat >"$GRW_PIN" <<'PYEOF'
import sys

EXPECTED = "go test -count=1 ./tools/goreleaserconfig/"

KEY_CHARS = set("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-")


def strip_comment(line):
    out = []
    quote = None
    prev_ws = True
    i = 0
    while i < len(line):
        c = line[i]
        if quote is None:
            if c == "#" and prev_ws:
                break
            if c == '"' or c == "'":
                quote = c
            out.append(c)
            prev_ws = c == " " or c == "\t"
        else:
            out.append(c)
            if quote == '"' and c == "\\" and i + 1 < len(line):
                i += 1
                out.append(line[i])
            elif c == quote:
                if quote == "'" and i + 1 < len(line) and line[i + 1] == "'":
                    i += 1
                    out.append(line[i])
                else:
                    quote = None
            prev_ws = False
        i += 1
    return "".join(out)


def unquote(value):
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
        inner = value[1:-1]
        if value[0] == "'":
            inner = inner.replace("''", "'")
        return inner
    return value


def split_key(text):
    cut = text.find(":")
    if cut <= 0:
        return None, None
    key = text[:cut]
    if not key or any(c not in KEY_CHARS for c in key):
        return None, None
    rest = text[cut + 1:]
    if rest != "" and rest[0] != " " and rest[0] != "\t":
        return None, None
    return key, rest.strip()


def main(path):
    try:
        with open(path, encoding="utf-8") as handle:
            raw_lines = handle.read().splitlines()
    except OSError as exc:
        print("wiring-pin: cannot read %s: %s" % (path, exc))
        return 1
    in_jobs = False
    job = None
    in_steps = False
    steps_indent = 0
    step_open = False
    run = None
    run_keys = 0
    has_if = False
    live = 0

    def flush():
        nonlocal live, step_open, run, run_keys, has_if
        if step_open and run_keys == 1 and not has_if and run == EXPECTED:
            live += 1
        step_open = False
        run = None
        run_keys = 0
        has_if = False

    def step_key(key, value):
        nonlocal run, run_keys, has_if
        if key == "run":
            run_keys += 1
            run = unquote(value)
        elif key == "if":
            has_if = True

    for raw in raw_lines:
        line = strip_comment(raw)
        if not line.strip():
            continue
        indent = len(line) - len(line.lstrip(" "))
        body = line.strip()
        if indent == 0:
            key, _ = split_key(body)
            in_jobs = key == "jobs"
            job = None
            in_steps = False
            flush()
            continue
        if not in_jobs:
            continue
        if body == "-" or (body.startswith("-") and body[1] in (" ", "\t")):
            if in_steps and indent == steps_indent + 2:
                flush()
                step_open = True
                rest = body[1:].strip()
                if rest:
                    key, value = split_key(rest)
                    if key is not None:
                        step_key(key, value)
            continue
        key, value = split_key(body)
        if key is None:
            continue
        if indent == 2 and not in_steps:
            job = key
            in_steps = False
            flush()
            continue
        if job != "lint":
            continue
        if indent == 4 and not in_steps and key == "steps":
            in_steps = True
            steps_indent = 4
            flush()
            continue
        if in_steps and indent <= steps_indent:
            in_steps = False
            flush()
            if indent == 2:
                job = key
            continue
        if in_steps and step_open and indent == steps_indent + 4:
            step_key(key, value)
    flush()

    if live == 1:
        print("wiring-pin: %s: exactly one live lint step runs the gate" % path)
        return 0
    if live == 0:
        print("wiring-pin: %s: no live lint step runs %s, want exactly one" % (path, EXPECTED))
        return 1
    print("wiring-pin: %s: %d live lint steps run %s, want exactly one" % (path, live, EXPECTED))
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else ".github/workflows/ci.yml"))
PYEOF
GRW='.github/workflows/ci.yml'
if command -v python3 >/dev/null 2>&1; then
	assert 'the lint job carries exactly one live gate step' 0 python3 "$GRW_PIN" "$GRW"
	assert_contains 'the pin reports the live step' 'exactly one live lint step runs the gate' "$WORK/out.txt"
	# R3 negatives, generated from the committed workflow at self-test
	# time so they track the file: each must fail the pin.
	sed 's|^\( *\)run: go test -count=1 \./tools/goreleaserconfig/$|\1# run: go test -count=1 ./tools/goreleaserconfig/|' "$GRW" >"$WORK/grw-commented.yml"
	assert 'a commented-out run line is not a live invocation' 1 python3 "$GRW_PIN" "$WORK/grw-commented.yml"
	assert_contains 'the pin names the miss' 'no live lint step runs' "$WORK/out.txt"
	awk '{ print } index($0, "run: go test -count=1 ./tools/goreleaserconfig/") { print "        if: false" }' "$GRW" >"$WORK/grw-if-false.yml"
	assert 'an if:false gate step is not live' 1 python3 "$GRW_PIN" "$WORK/grw-if-false.yml"
	grep -v 'run: go test -count=1 ./tools/goreleaserconfig/' "$GRW" >"$WORK/grw-deleted.yml"
	assert 'a deleted gate step fails the pin' 1 python3 "$GRW_PIN" "$WORK/grw-deleted.yml"
	sed 's|./tools/goreleaserconfig/|./tools/othergate/|' "$GRW" >"$WORK/grw-otherpath.yml"
	assert 'a changed gate path fails the pin' 1 python3 "$GRW_PIN" "$WORK/grw-otherpath.yml"
	sed 's|run: go test -count=1 ./tools/goreleaserconfig/|run: "go test -count=1 ./tools/goreleaserconfig/"|' "$GRW" >"$WORK/grw-quoted.yml"
	assert 'a quoted run value is the same invocation' 0 python3 "$GRW_PIN" "$WORK/grw-quoted.yml"
	sed "s|run: go test -count=1 ./tools/goreleaserconfig/|run: 'go test -count=1 ./tools/goreleaserconfig/'|" "$GRW" >"$WORK/grw-squoted.yml"
	assert 'a single-quoted run value is the same invocation' 0 python3 "$GRW_PIN" "$WORK/grw-squoted.yml"
	sed 's|run: go test -count=1 ./tools/goreleaserconfig/$|run: go test -count=1 ./tools/goreleaserconfig/ # the rc channel guard|' "$GRW" >"$WORK/grw-trailing.yml"
	assert 'a trailing comment on the run line still counts' 0 python3 "$GRW_PIN" "$WORK/grw-trailing.yml"
	awk '{ print } index($0, "run: go test -count=1 ./tools/goreleaserconfig/") { print "        run: go test -count=1 ./tools/goreleaserconfig/" }' "$GRW" >"$WORK/grw-dup-run.yml"
	assert 'duplicate run keys disqualify the step' 1 python3 "$GRW_PIN" "$WORK/grw-dup-run.yml"
	assert 'an unreadable workflow fails the pin closed' 1 python3 "$GRW_PIN" "$WORK/grw-absent.yml"
else
	skip 'the lint wiring pin needs python3' 'python3 not on PATH (the lint lane still asserts its own wiring via TestCommittedWiring)'
fi

echo '=== naming-gate.sh: real mentions fail, binary-patch base85 noise does not ==='
NG="$HERE/naming-gate.sh"
if command -v python3 >/dev/null 2>&1; then
	# Names assembled at runtime so this file never spells either one.
	ng_full="$(printf '%s%s' wild berries)"
	ng_short="$(printf '%s%s' w b)"
	ng_tree() { rm -rf "$WORK/ng"; mkdir -p "$WORK/ng/docs"; printf 'clean text\n' >"$WORK/ng/docs/readme.md"; }
	ng_binary_patch() { # $1 = one line placed inside the binary block
		printf 'diff --git a/x.bin b/x.bin\nindex 1..2 100644\nGIT binary patch\nliteral 12\n%s\nzcmZ?wbhEHbWMp7uXkcIfVnzmrf\n\nliteral 0\nHcmV?d00001\n\ndiff --git a/y.txt b/y.txt\n+clean\n' "$1" >"$WORK/ng/change.patch"
	}
	ng_tree; ng_binary_patch "zq*>K|@dcWsQ;${ng_short};M7nsyfzNC>Fby"
	assert 'naming (a): short word on a base85 line in a binary block passes' 0 bash "$NG" "$WORK/ng"
	ng_tree; printf 'diff --git a/n.md b/n.md\n+ask %s about it\n' "$ng_short" >"$WORK/ng/change.patch"
	assert 'naming (b): short word on a text line of a .patch fails' 1 bash "$NG" "$WORK/ng"
	assert_contains 'naming (b) fails for the short-name reason' "the employer's short name must not appear" "$WORK/out.txt"
	ng_tree; ng_binary_patch "zq*>K|@dc ${ng_short} M7nsyf"
	assert 'naming (c): non-base85 line inside a binary block fails' 1 bash "$NG" "$WORK/ng"
	assert_contains 'naming (c) fails for the short-name reason' "the employer's short name must not appear" "$WORK/out.txt"
	ng_tree; printf 'Written at %s.\n' "$(printf '%s' "$ng_full" | tr a-z A-Z)" >"$WORK/ng/docs/history.md"
	assert 'naming (d): full name in markdown fails' 1 bash "$NG" "$WORK/ng"
	assert_contains 'naming (d) fails for the full-name reason' 'the employer name must not appear' "$WORK/out.txt"
	ng_tree
	assert 'naming (e): a clean tree passes' 0 bash "$NG" "$WORK/ng"
	ng_sig='Employer name gate'
	ng_tree; printf -- '- CI: %s\t2026-09-28T10:11:12.3456789Z ./docs/x.md:3:ask %s\n' "$ng_sig" "$ng_short" >"$WORK/ng/progress.md"
	assert 'naming (f): a historical CI echo with the step signature passes' 0 bash "$NG" "$WORK/ng"
	ng_tree; printf '{"log":"%s\\t2026-09-28T10:11:12.3456789Z ./docs/x.md:3:ask %s"}\n' "$ng_sig" "$ng_short" >"$WORK/ng/events.ndjson"
	assert 'naming (g): a JSON-escaped CI echo with the step signature passes' 0 bash "$NG" "$WORK/ng"
	ng_tree; printf '%s: ask %s about ./docs\n' "$ng_sig" "$ng_short" >"$WORK/ng/docs/notes.md"
	assert 'naming (h): the step name without timestamp and ./ is still scanned' 1 bash "$NG" "$WORK/ng"
	ng_tree; printf '  log: \342\200\246-29T01:51:52.8528587Z ./docs/x.patch:9:zq*>K|@dc %s M7\n' "$ng_short" >"$WORK/ng/progress.md"
	assert 'naming (j): a prefix-truncated CI echo (ellipsis + timestamp tail + ./) passes' 0 bash "$NG" "$WORK/ng"
	ng_tree; printf 'as said \342\200\246 ask %s about ./docs\n' "$ng_short" >"$WORK/ng/docs/notes.md"
	assert 'naming (k): prose with an ellipsis but no timestamp tail is still scanned' 1 bash "$NG" "$WORK/ng"
	ng_tree; printf 'ask %s\nand %s\n' "$ng_short" "$ng_full" >"$WORK/ng/docs/notes.md"
	assert 'naming (i): a planted mention fails' 1 bash "$NG" "$WORK/ng"
	if grep -qiF "$ng_short" "$WORK/out.txt" || grep -qiF "$ng_full" "$WORK/out.txt"; then
		bad 'naming (i): the failure report carries no line content' "planted word echoed in gate output"
	else
		ok 'naming (i): the failure report carries no line content'
	fi
	assert_contains 'naming (i) reports path:line' './docs/notes.md:1' "$WORK/out.txt"
else
	skip 'the naming gate rows need python3' 'python3 not on PATH'
fi

echo ''
printf 'gate-selftest: %d passed, %d failed' "$PASS" "$FAIL"
[ "$SKIPPED" -gt 0 ] && printf ', %d skipped (reported above, not hidden)' "$SKIPPED"
printf '\n'
[ "$FAIL" -eq 0 ] || exit 1
exit 0
