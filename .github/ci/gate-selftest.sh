#!/usr/bin/env bash
# Gate self-test.
#
# Every gate in this directory is only worth its exit code if it rejects what
# it claims to reject. This drives each one against synthetic inputs and
# asserts a REAL exit code for every case -- including the negative cases,
# which are the ones a gate silently loses when it is refactored.
#
# It needs no conformance root, no network and no Go build. The handful of
# cases that need a `go` launcher are named and reported when one is absent,
# never quietly dropped.
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

WORKFLOW='.github/workflows/ci.yml'
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
	assert_contains 'evidence records the committed pin'     "committed_released_pin  $PIN"     "$EV"
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
echo '=== ci.yml: every windows lane that runs test-gate.sh buys the slower runner its own budget ==='

# A per-package timeout is a hosted-runner budget, and the Windows runner needs
# a bigger one than the unix runners for the same passing work. That budget
# lives in the workflow, so a refactor can drop it without any gate noticing --
# the run just dies at the old deadline and the ledger reports the truncated
# cases as cases that never ran. These cases read the wiring back.
#
# Emits one "<job> <windows-lane?> <has-timeout?>" record per test-gate.sh step.
gate_lane_records() {
	awk '
		/^  [a-z][a-z0-9-]*:[ \t\r]*$/ { job = $1; sub(/:[ \t\r]*$/, "", job) }
		/^[ \t]*os:[ \t]*\[/           { if ($0 ~ /windows-latest/) win[job] = 1 }
		/^      - name:/               { timeout = 0 }
		/^[ \t]*GO_TEST_TIMEOUT:/      { timeout = 1 }
		/test-gate\.sh/ && /^[ \t]*run:/ {
			printf "%s %d %d\n", job, (win[job] ? 1 : 0), timeout
		}
	' "$WORKFLOW"
}

lane_records="$(gate_lane_records)"
if [ -z "$lane_records" ]; then
	bad 'the workflow still runs test-gate.sh' 'no test-gate.sh step found in .github/workflows/ci.yml'
else
	unbudgeted=''
	while read -r job is_win has_timeout; do
		[ -n "$job" ] || continue
		[ "$is_win" = '1' ] || continue
		[ "$has_timeout" = '1' ] || unbudgeted="$unbudgeted $job"
	done <<-RECORDS
	$lane_records
	RECORDS
	if [ -z "$unbudgeted" ]; then
		ok 'every windows test-gate.sh lane declares GO_TEST_TIMEOUT'
	else
		bad 'every windows test-gate.sh lane declares GO_TEST_TIMEOUT' \
			"lanes running on windows-latest with no GO_TEST_TIMEOUT:$unbudgeted"
	fi
fi

# minutes <duration>  -- "60m" and "1h" both answer 60; anything else answers -1.
minutes() {
	case "$1" in
	*m) printf '%s\n' "${1%m}" ;;
	*h) printf '%s\n' "$(( ${1%h} * 60 ))" ;;
	*)  printf '%s\n' '-1' ;;
	esac
}

GATE_DEFAULT_TIMEOUT="$(awk -F'-' '/^GO_TEST_TIMEOUT=/{print $2; exit}' "$HERE/test-gate.sh" | tr -d '}"')"
win_timeout="$(awk -F"'" '/GO_TEST_TIMEOUT:.*Windows/{print $4; exit}' "$WORKFLOW")"
other_timeout="$(awk -F"'" '/GO_TEST_TIMEOUT:.*Windows/{print $6; exit}' "$WORKFLOW")"
win_minutes="$(minutes "$win_timeout")"
other_minutes="$(minutes "$other_timeout")"
default_minutes="$(minutes "$GATE_DEFAULT_TIMEOUT")"

if [ "$win_minutes" -gt 0 ] && [ "$other_minutes" -gt 0 ] && [ "$win_minutes" -gt "$other_minutes" ]; then
	ok 'the windows per-package budget is larger than the unix one'
else
	bad 'the windows per-package budget is larger than the unix one' \
		"windows='$win_timeout' other='$other_timeout'"
fi

if [ "$default_minutes" -gt 0 ] && [ "$other_minutes" -eq "$default_minutes" ]; then
	ok 'the unix per-package budget still matches the gate default'
else
	bad 'the unix per-package budget still matches the gate default' \
		"workflow='$other_timeout' test-gate.sh default='$GATE_DEFAULT_TIMEOUT'"
fi

echo ''
printf 'gate-selftest: %d passed, %d failed' "$PASS" "$FAIL"
[ "$SKIPPED" -gt 0 ] && printf ', %d skipped (reported above, not hidden)' "$SKIPPED"
printf '\n'
[ "$FAIL" -eq 0 ] || exit 1
exit 0
