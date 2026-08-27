#!/usr/bin/env bash
# Self-test for the compiled-build CI gates.
#
# A gate is only worth its exit code if it rejects what it claims to reject, so
# every rule in `platform-case-gate.sh` and `candidate-suite.sh` is exercised
# here against a case that should fail as well as one that should pass. It runs
# the real scripts -- no reimplementation -- on synthetic inputs, so it needs no
# network, no Go build and no second platform.
#
# Usage: gate-selftest.sh
# Exit:  0 when every case behaves as declared, 1 otherwise.

set -u

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
GATE="$HERE/platform-case-gate.sh"
CANDIDATE="$HERE/candidate-suite.sh"
REAL_LEDGER="$HERE/platform-cases.tsv"

WORK="$(mktemp -d)" || exit 2
trap 'rm -rf "$WORK"' EXIT

MODULE="example.com/m"
PASSED=0
FAILED=0

# expect <want-rc> <name> -- reads the command from the remaining arguments.
expect() {
	want="$1"; name="$2"; shift 2
	"$@" >"$WORK/out.txt" 2>&1
	got=$?
	if [ "$got" -eq "$want" ]; then
		printf 'ok    %-58s (exit %s)\n' "$name" "$got"
		PASSED=$((PASSED + 1))
	else
		printf 'FAIL  %-58s (exit %s, want %s)\n' "$name" "$got" "$want"
		sed 's/^/          | /' "$WORK/out.txt"
		FAILED=$((FAILED + 1))
	fi
}

ev() { # ev <action> <package> <test>
	printf '{"Time":"2026-01-01T00:00:00Z","Action":"%s","Package":"%s/%s","Test":"%s","Elapsed":0}\n' \
		"$1" "$MODULE" "$2" "$3"
}

run_gate() { # run_gate <goos> <ledger> <stream>
	CI_GATE_GOOS="$1" CI_GATE_MODULE="$MODULE" CI_PLATFORM_CASES="$2" \
		bash "$GATE" "$3" "$WORK/evidence"
}

echo '== platform-case gate =='

# A small synthetic ledger with one case of each shape.
LEDGER="$WORK/ledger.tsv"
{
	printf '# comment line is ignored\n'
	printf '\n'
	printf 'internal/win\tTestCmdLauncher\twindows\t-\twindows .cmd launcher\n'
	printf 'internal/win\tTestReparsePoint\tlinux,darwin,windows\t-\tlink or reparse point creation\n'
	printf 'internal/nix\tTestNoFollow\tlinux,darwin\twindows\tno-follow link rejection\n'
	printf 'internal/nix\tTestOwnership\tlinux,darwin,windows\t-\townership and permission bits\n'
	printf 'internal/opt\tTestOptIn\t-\tlinux,darwin,windows\topt-in developer probe\n'
} >"$LEDGER"

# A windows stream that satisfies the ledger exactly.
{
	ev pass internal/win TestCmdLauncher
	ev pass internal/win TestReparsePoint
	ev skip internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
	ev skip internal/opt TestOptIn
} >"$WORK/win-ok.json"
expect 0 'windows: every required case ran, tolerated skips only' run_gate windows "$LEDGER" "$WORK/win-ok.json"

# The same stream on linux: TestCmdLauncher is not required there, and
# TestNoFollow must NOT be skipped.
{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
	ev skip internal/opt TestOptIn
} >"$WORK/linux-ok.json"
expect 0 'linux: every required case ran, tolerated skips only' run_gate linux "$LEDGER" "$WORK/linux-ok.json"

# Required-case regressions.
{
	ev pass internal/win TestCmdLauncher
	ev skip internal/win TestReparsePoint
	ev skip internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
} >"$WORK/win-reparse-skipped.json"
expect 1 'windows: a skipped reparse-point case is rejected' run_gate windows "$LEDGER" "$WORK/win-reparse-skipped.json"

{
	ev skip internal/win TestCmdLauncher
	ev pass internal/win TestReparsePoint
	ev skip internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
} >"$WORK/win-cmd-skipped.json"
expect 1 'windows: a skipped .cmd launcher case is rejected' run_gate windows "$LEDGER" "$WORK/win-cmd-skipped.json"

{
	ev pass internal/win TestReparsePoint
	ev skip internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
} >"$WORK/linux-nofollow-skipped.json"
expect 1 'linux: a skipped no-follow case is rejected' run_gate linux "$LEDGER" "$WORK/linux-nofollow-skipped.json"

{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
} >"$WORK/linux-missing.json"
expect 1 'a required case that never ran is rejected (rename or -run filter)' run_gate linux "$LEDGER" "$WORK/linux-missing.json"

{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
	ev fail internal/nix TestOwnership
} >"$WORK/linux-failed.json"
expect 1 'a required case that failed is rejected' run_gate linux "$LEDGER" "$WORK/linux-failed.json"

# Skips nobody declared.
{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
	ev skip internal/other TestUndeclared
} >"$WORK/linux-undeclared-skip.json"
expect 1 'an undeclared skip is rejected' run_gate linux "$LEDGER" "$WORK/linux-undeclared-skip.json"

{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
	ev skip internal/opt TestOptIn
} >"$WORK/linux-tolerated-skip.json"
expect 0 'a ledger-tolerated skip is accepted' run_gate linux "$LEDGER" "$WORK/linux-tolerated-skip.json"

# A subtest result is evidence about its parent case.
{
	ev pass 'internal/win' 'TestReparsePoint/relative'
	ev pass internal/win TestReparsePoint
	ev pass 'internal/nix' 'TestNoFollow/symlink'
	ev pass internal/nix TestNoFollow
	ev pass internal/nix TestOwnership
} >"$WORK/linux-subtests.json"
expect 0 'subtest results resolve to their parent case' run_gate linux "$LEDGER" "$WORK/linux-subtests.json"

# An `output` event whose text contains a JSON-looking action must not be
# mistaken for a result: test2json puts the real Action before Output.
{
	ev pass internal/win TestReparsePoint
	ev pass internal/nix TestNoFollow
	printf '{"Time":"2026-01-01T00:00:00Z","Action":"output","Package":"%s/internal/nix","Test":"TestOwnership","Output":"logged: {\\"Action\\":\\"skip\\",\\"Test\\":\\"TestOwnership\\"}\\n"}\n' "$MODULE"
	ev pass internal/nix TestOwnership
} >"$WORK/linux-output-noise.json"
expect 0 'test output that looks like a result event is not parsed as one' run_gate linux "$LEDGER" "$WORK/linux-output-noise.json"

# The real ledger must be internally satisfiable on each platform: a stream
# that passes exactly its required rows and skips exactly its tolerated rows is
# green. This keeps the shipped ledger honest as it grows.
for goos in linux darwin windows; do
	stream="$WORK/real-$goos.json"
	: >"$stream"
	while IFS="$(printf '\t')" read -r pkg tst must skipok _rest; do
		case "$pkg" in ''|\#*) continue ;; esac
		case ",$must," in *",$goos,"*) ev pass "$pkg" "$tst" >>"$stream"; continue ;; esac
		case ",$skipok," in *",$goos,"*) ev skip "$pkg" "$tst" >>"$stream" ;; esac
	done <"$REAL_LEDGER"
	CI_GATE_MODULE="$MODULE" expect 0 "shipped ledger is satisfiable on $goos" run_gate "$goos" "$REAL_LEDGER" "$stream"
done

echo
echo '== candidate suite: revision shape =='

PIN=00b1688a9b2457ca397a0bb550acf47cad8ee967
vr() { SPEC_PIN="$PIN" bash "$CANDIDATE" verify-ref "$1"; }

expect 1 'a branch name is rejected as a candidate pin'      vr 'main'
expect 1 'a tag is rejected as a candidate pin'              vr 'v1.0.0-rc.5'
expect 1 'HEAD is rejected as a candidate pin'               vr 'HEAD'
expect 1 'a short hash is rejected as a candidate pin'       vr '00b1688'
expect 1 'uppercase hex is rejected as a candidate pin'      vr '00B1688A9B2457CA397A0BB550ACF47CAD8EE967'
expect 1 'a placeholder is rejected as a candidate pin'      vr 'TBD'
expect 1 'an empty revision is rejected'                     vr ''
expect 1 'the null commit is rejected'                       vr '0000000000000000000000000000000000000000'
expect 1 'a candidate equal to the committed pin is refused' vr "$PIN"
expect 0 'a full 40-hex revision is accepted'                vr '1234567890abcdef1234567890abcdef12345678'

echo
echo '== candidate suite: identity and evidence =='

SUITE="$WORK/suite/conformance/v1"
mkdir -p "$SUITE/vectors"
printf '{"protocol_version":"1.0.0-rc.5","generator":"selftest"}' >"$SUITE/manifest.json"
printf 'vector-a' >"$SUITE/vectors/a.json"

EMPTY="$WORK/empty"
mkdir -p "$EMPTY"

rec() { bash "$CANDIDATE" record "$1" "$WORK/candidate-evidence"; }

expect 1 'a root without manifest.json is rejected'  rec "$EMPTY"
expect 1 'a nonexistent root is rejected'            rec "$WORK/nope"
expect 0 'a well-formed candidate root is recorded'  rec "$SUITE"

MANIFEST_SHA="$(shasum -a 256 "$SUITE/manifest.json" 2>/dev/null || sha256sum "$SUITE/manifest.json")"
MANIFEST_SHA="${MANIFEST_SHA%% *}"

expect 0 'a matching expected manifest digest is accepted' \
	env CANDIDATE_EXPECTED_MANIFEST_SHA256="$MANIFEST_SHA" bash "$CANDIDATE" record "$SUITE" "$WORK/candidate-evidence"
expect 1 'a mismatched expected manifest digest aborts the run' \
	env CANDIDATE_EXPECTED_MANIFEST_SHA256=deadbeef bash "$CANDIDATE" record "$SUITE" "$WORK/candidate-evidence"
expect 1 'a candidate identical to the committed pin is refused' \
	env PIN_MANIFEST_SHA256="$MANIFEST_SHA" bash "$CANDIDATE" record "$SUITE" "$WORK/candidate-evidence"

IDENTITY="$WORK/candidate-evidence/candidate-suite-identity.txt"
check_evidence() {
	[ -f "$IDENTITY" ] || { echo "no identity file at $IDENTITY"; return 1; }
	grep -q 'NOT A RELEASE'                  "$IDENTITY" || { echo 'identity file does not state NOT A RELEASE'; return 1; }
	grep -q '^release_claim           none'  "$IDENTITY" || { echo 'identity file does not disclaim a release'; return 1; }
	grep -q '^conformance_claim       none'  "$IDENTITY" || { echo 'identity file does not disclaim a conformance claim'; return 1; }
	grep -q '^manifest_sha256         sha256:' "$IDENTITY" || { echo 'identity file records no manifest digest'; return 1; }
	grep -q '^tree_sha256             sha256:' "$IDENTITY" || { echo 'identity file records no tree digest'; return 1; }
	grep -q '^file_count              2$'    "$IDENTITY" || { echo 'identity file records the wrong file count'; return 1; }
	grep -q '^protocol_version        1.0.0-rc.5$' "$IDENTITY" || { echo 'identity file records the wrong protocol version'; return 1; }
	return 0
}
expect 0 'evidence records revision, digests and both disclaimers' check_evidence

echo
printf 'gate-selftest: %s passed, %s failed\n' "$PASSED" "$FAILED"
[ "$FAILED" -eq 0 ] || exit 1
exit 0
