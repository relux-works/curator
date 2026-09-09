#!/usr/bin/env bash
# Owner/form conformance-gate runner (TASK-260909-3d1589).
# Tree: fbe90d5e60593a3a069721b2ad9e53cd071d8c02 (exact candidate).
#
# Usage:
#   ./TASK-260909-3d1589_run-gate.sh [workdir]
#   ./TASK-260909-3d1589_run-gate.sh --self-check
#
# Gate flow: verifies exact source identity, extracts a disposable copy via
# git archive, installs the gate overlays published under these exact names:
#   TASK-260909-3d1589_gate-conformance_test.go -> internal/diagnostics/gate_conformance_test.go
#   TASK-260909-3d1589_gate-framing_test.go      -> cmd/curator-run/gate_framing_test.go
# then proves baseline pass and that 9 single-member narrowings (3 owner
# foreign-code admissions + 1 joined-positive rejection + 4 fixed-owner
# positive-form rejections covering BOTH fixed owners in joined and wrapped
# forms + 1 exact framing exemption) each fail their named assertion. Never
# touches the original workspace; every mutant is restored byte-equal before
# the next.
#
# Fail-closed: every setup/extraction/copy/mutation/test/restoration step is
# checked explicitly. A compile error, a missing test ("no tests to run"),
# or a zero-selection run is a RUNNER failure, never a killed mutant.
# Expected negative Go exits are captured (no `set -e` anywhere near them).
set -u

TREE="fbe90d5e60593a3a069721b2ad9e53cd071d8c02"
SPEC_SHA="5a7ccf0ba95708cb573a586977eb0ba4bad1e46233a540dc99784c1d4d922d48"
GATE_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_SRC="$GATE_DIR/TASK-260909-3d1589_gate-conformance_test.go"
FRAM_SRC="$GATE_DIR/TASK-260909-3d1589_gate-framing_test.go"

PINNED_FILES=".scripts/diagnostics-mutants.sh README.md cmd/curator-run/main.go cmd/curator-run/main_test.go internal/diagnostics/diagnostics.go internal/diagnostics/diagnostics_test.go internal/diagnostics/helpers_test.go SPEC.md"

DIAG_BASELINE_TESTS="TestGateResolveRejectsForeignCodes TestGateLayerRejectsForeignCodes TestGateRefusalRejectsForeignCodes TestGateJoinedOwnedAccepted TestGateCoverageCounts TestGateOwnFamilyAndNilPreserved TestGateOwnerFormPositives TestGateOwnerFormNil"
FRAM_BASELINE_TESTS="TestGateFramingSingleDetailAtRealResolver TestGateFramingCompanionsStayFramed"

fail() { echo "GATE-FAIL: $1" >&2; exit 1; }
info() { echo "$1"; }

need_cmd() { command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"; }

# check_go_output <label> <output> <required-names...>
# Fails closed when the Go run selected nothing, did not build, or failed.
check_go_output() {
  local label="$1" output="$2"; shift 2
  if printf '%s\n' "$output" | grep -q "no tests to run"; then
    fail "$label: zero tests selected (missing overlay or wrong -run filter)"
  fi
  if printf '%s\n' "$output" | grep -Eq "build failed|cannot |undefined:|does not import|no Go files"; then
    fail "$label: compile/setup error, not a behavioral result"
  fi
  local name
  for name in "$@"; do
    printf '%s\n' "$output" | grep -q "^--- PASS: $name" || fail "$label: required test $name did not PASS"
  done
  if printf '%s\n' "$output" | grep -Eq "^(FAIL|--- FAIL)"; then
    fail "$label: unexpected FAIL lines in baseline output"
  fi
}

# expect_kill <label> <output> <go-exit> <named-assertion...>
# Requires exit 1 AND the intended named behavioral assertion. Any compile
# error, missing test, or exit 0 (survival) is a gate failure.
expect_kill() {
  local label="$1" output="$2" code="$3"; shift 3
  if printf '%s\n' "$output" | grep -q "no tests to run"; then
    fail "$label: zero tests selected; mutant verdict is void"
  fi
  if printf '%s\n' "$output" | grep -Eq "build failed|cannot |undefined:|does not import|no Go files"; then
    fail "$label: compile/setup error; a broken build is not a killed mutant"
  fi
  [ "$code" = "1" ] || fail "$label: SURVIVED (exit $code, want 1)"
  local needle
  for needle in "$@"; do
    printf '%s\n' "$output" | grep -F -q "$needle" || fail "$label: exit 1 but named assertion missing: $needle"
  done
}

verify_identity() {
  need_cmd git; need_cmd go; need_cmd python3
  [ "$(git cat-file -t "$TREE" 2>/dev/null)" = "tree" ] || fail "tree $TREE not present in git object database"
  local f want got
  for f in $PINNED_FILES; do
    case "$f" in
      .scripts/diagnostics-mutants.sh) want="5ff00da1d66e5def3bfc43e12b7b01708f0cefc3" ;;
      README.md) want="2fee2deebdccfb374fc7fa81952f3ecd666b6158" ;;
      cmd/curator-run/main.go) want="f2a23748d8eb3f3f8bca389d390373e6cf4fdc10" ;;
      cmd/curator-run/main_test.go) want="849e6aaee17c4040e88a7ab2d7b8c01e4ad2ca70" ;;
      internal/diagnostics/diagnostics.go) want="4baf919afcb56981a1cdc7f208eb5dafe340aedb" ;;
      internal/diagnostics/diagnostics_test.go) want="dcfbc7e1f8369cff46a6b54aa84b0c1b205996cb" ;;
      internal/diagnostics/helpers_test.go) want="a46ad729092296ac9be3dcc4619ac825a61c0b30" ;;
      SPEC.md) want="997f00651692999577a3e775d28c72747f83e4d2" ;;
    esac
    got="$(git rev-parse "$TREE:$f" 2>/dev/null)" || fail "cannot resolve $TREE:$f"
    [ "$got" = "$want" ] || fail "blob drift for $f: $got != $want"
  done
  got="$(git show "$TREE:SPEC.md" 2>/dev/null | python3 -c 'import hashlib,sys; print(hashlib.sha256(sys.stdin.buffer.read()).hexdigest())')"
  [ "$got" = "$SPEC_SHA" ] || fail "SPEC.md sha256 drift: $got"
  info "identity: tree $TREE pinned (8 blobs + SPEC sha256 match)"
}

extract_tree() {
  # $1 = workdir (must already exist and be empty)
  local work="$1" top
  # git archive without paths archives only the cwd subdirectory (empty and
  # exit 0 from an untracked dir). Always archive from the toplevel.
  top="$(git rev-parse --show-toplevel 2>/dev/null)" || fail "not inside a git worktree; cannot reach tree $TREE"
  git -C "$top" archive "$TREE" 2>"$work/.archive.err" | tar -x -C "$work" 2>"$work/.tar.err"
  local arc_status="${PIPESTATUS[0]}" tar_status="${PIPESTATUS[1]}"
  [ "$arc_status" = "0" ] || { cat "$work/.archive.err" >&2; fail "git archive of $TREE failed"; }
  [ "$tar_status" = "0" ] || { cat "$work/.tar.err" >&2; fail "tar extraction into $work failed"; }
  [ -f "$work/internal/diagnostics/diagnostics.go" ] || fail "extracted copy missing internal/diagnostics/diagnostics.go"
  info "extract: disposable copy ready at $work"
}

install_overlays() {
  # $1 = workdir
  local work="$1"
  [ -f "$CONF_SRC" ] || fail "missing overlay source: $CONF_SRC"
  [ -f "$FRAM_SRC" ] || fail "missing overlay source: $FRAM_SRC"
  cp "$work/internal/diagnostics/diagnostics.go" "$work/diagnostics.go.orig" || fail "cannot snapshot diagnostics.go"
  cp "$CONF_SRC" "$work/internal/diagnostics/gate_conformance_test.go" || fail "cannot install conformance overlay"
  cp "$FRAM_SRC" "$work/cmd/curator-run/gate_framing_test.go" || fail "cannot install framing overlay"
  [ -f "$work/internal/diagnostics/gate_conformance_test.go" ] || fail "conformance overlay absent after copy"
  [ -f "$work/cmd/curator-run/gate_framing_test.go" ] || fail "framing overlay absent after copy"
  info "install: gate overlays staged under published names"
}

restore_check() {
  # $1 = workdir
  local work="$1"
  cp "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || fail "mutant restoration copy failed"
  cmp -s "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || fail "diagnostics.go not byte-equal after restore"
}

apply_replace() {
  # $1 = workdir, FROM/TO via env
  local work="$1"
  FROM="$FROM" TO="$TO" python3 -c '
import os,sys
p=sys.argv[1]; src=open(p).read()
f=os.environ["FROM"]; t=os.environ["TO"]
assert src.count(f)==1, "FROM anchor not unique/found"
open(p,"w").write(src.replace(f,t,1))
' "$work/internal/diagnostics/diagnostics.go" || fail "mutant application failed ($MNAME)"
}

run_baselines() {
  # $1 = workdir; fails closed unless every named gate test passes.
  local work="$1" out code
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  [ "$code" = "0" ] || fail "baseline diagnostics gate exit=$code, want 0"
  # shellcheck disable=SC2086
  check_go_output "baseline diagnostics gate" "$out" $DIAG_BASELINE_TESTS
  info "baseline-diagnostics-gate exit=0 (8/8 named TestGate tests PASS)"
  out="$(cd "$work" && go test ./cmd/curator-run -run 'TestGateFraming' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  [ "$code" = "0" ] || fail "baseline framing gate exit=$code, want 0"
  # shellcheck disable=SC2086
  check_go_output "baseline framing gate" "$out" $FRAM_BASELINE_TESTS
  info "baseline-framing-gate exit=0 (2/2 named TestGateFraming tests PASS)"
  out="$(cd "$work" && go test ./internal/diagnostics ./cmd/curator-run -count=1 2>&1)"; code="$?"
  printf '%s\n' "$out" | tail -n 5
  [ "$code" = "0" ] || fail "baseline full-package suites exit=$code, want 0"
  info "baseline full-package suites exit=0 (overlays break no existing test)"
}

run_owner_mutants() {
  # $1 = workdir
  local work="$1" out code
  info "== MUTANT resolve: admits mcp_layer_missing (expect TestGateResolveRejectsForeignCodes fail, exit 1) =="
  FROM='if re, ok := e.(*fragment.ResolveError); ok && re != nil && resolveFamily[re.Code] {' \
  TO='if re, ok := e.(*fragment.ResolveError); ok && re != nil && (resolveFamily[re.Code] || re.Code == composition.CodeMCPLayerMissing) {' \
  MNAME="resolve" apply_replace "$work"
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "resolve-mutant" "$out" "$code" 'resolve accepted foreign "mcp_layer_missing"'
  info "resolve-mutant exit=1 via TestGateResolveRejectsForeignCodes (other owners still PASS)"
  restore_check "$work"

  info "== MUTANT layer: admits resolve_invocation_failed (expect TestGateLayerRejectsForeignCodes fail, exit 1) =="
  FROM='if le, ok := e.(*composition.LayerError); ok && le != nil && layerFamily[le.Code] {' \
  TO='if le, ok := e.(*composition.LayerError); ok && le != nil && (layerFamily[le.Code] || le.Code == fragment.CodeInvocationFailed) {' \
  MNAME="layer" apply_replace "$work"
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "layer-mutant" "$out" "$code" 'layer accepted foreign "resolve_invocation_failed"'
  info "layer-mutant exit=1 via TestGateLayerRejectsForeignCodes (prior R3 survivor killed)"
  restore_check "$work"

  info "== MUTANT refusal: admits resolve_invocation_failed (expect TestGateRefusalRejectsForeignCodes fail, exit 1) =="
  FROM='if sr, ok := e.(*systemprompt.Refusal); ok && sr != nil && refusalFamily[sr.Code] {' \
  TO='if sr, ok := e.(*systemprompt.Refusal); ok && sr != nil && (refusalFamily[sr.Code] || sr.Code == fragment.CodeInvocationFailed) {' \
  MNAME="refusal" apply_replace "$work"
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "refusal-mutant" "$out" "$code" 'refusal accepted foreign "resolve_invocation_failed"'
  info "refusal-mutant exit=1 via TestGateRefusalRejectsForeignCodes (prior R3 survivor killed)"
  restore_check "$work"

  info "== MUTANT joined-positive: joined resolve_invocation_failed rejected (expect TestGateJoinedOwnedAccepted fail, exit 1) =="
  FROM='func CodeOf(err error) (code string, ok bool) {' \
  TO='func CodeOf(err error) (code string, ok bool) {
	if joined, yes := err.(interface{ Unwrap() []error }); yes {
		for _, child := range joined.Unwrap() {
			if re, yes := child.(*fragment.ResolveError); yes && re != nil && re.Code == fragment.CodeInvocationFailed { return "", false }
		}
	}' \
  MNAME="joined-positive" apply_replace "$work"
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGateJoinedOwnedAccepted$' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "joined-positive-mutant" "$out" "$code" 'joined owned resolve "resolve_invocation_failed" lost'
  info "joined-positive-mutant exit=1 via TestGateJoinedOwnedAccepted (rev1 reviewer survivor killed)"
  restore_check "$work"
}

run_fixed_mutant() {
  # $1 = workdir, $2 = label, $3 = needle, FROM/TO via env.
  # Focused run of the registry positive test: only the narrowed fixed-owner
  # cell may fail.
  local work="$1" label="$2" needle="$3" out code
  info "== MUTANT $label (expect TestGateOwnerFormPositives fail, exit 1) =="
  MNAME="$label" apply_replace "$work"
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGateOwnerFormPositives$' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "$label-mutant" "$out" "$code" "$needle"
  info "$label-mutant exit=1 via TestGateOwnerFormPositives ($needle)"
  restore_check "$work"
}

run_fixed_mutants() {
  # $1 = workdir. One narrowing per fixed owner per non-direct form: the
  # usage-joined mutant is the exact rev2 reviewer survivor; the other three
  # prove the fix is not limited to that remembered example. Wrapped mutants
  # guard with the existing nilValue helper before calling Unwrap: several
  # owner Unwrap methods dereference the receiver and panic on typed-nil
  # values, which the production chain() walk skips first for the same reason.
  local work="$1"
  FROM='func CodeOf(err error) (code string, ok bool) {' \
  TO='func CodeOf(err error) (code string, ok bool) {
	if joined, yes := err.(interface{ Unwrap() []error }); yes {
		for _, child := range joined.Unwrap() {
			if ue, yes := child.(*cli.UsageError); yes && ue != nil { return "", false }
		}
	}' \
  run_fixed_mutant "$work" "usage-joined" 'usage joined lost: "" false'

  FROM='func CodeOf(err error) (code string, ok bool) {' \
  TO='func CodeOf(err error) (code string, ok bool) {
	if w, yes := err.(interface{ Unwrap() error }); yes && !nilValue(err) {
		if ue, yes := w.Unwrap().(*cli.UsageError); yes && ue != nil { return "", false }
	}' \
  run_fixed_mutant "$work" "usage-wrapped" 'usage wrapped lost: "" false'

  FROM='func CodeOf(err error) (code string, ok bool) {' \
  TO='func CodeOf(err error) (code string, ok bool) {
	if joined, yes := err.(interface{ Unwrap() []error }); yes {
		for _, child := range joined.Unwrap() {
			if ae, yes := child.(*axconfig.Error); yes && ae != nil { return "", false }
		}
	}' \
  run_fixed_mutant "$work" "axconfig-joined" 'axconfig joined lost: "" false'

  FROM='func CodeOf(err error) (code string, ok bool) {' \
  TO='func CodeOf(err error) (code string, ok bool) {
	if w, yes := err.(interface{ Unwrap() error }); yes && !nilValue(err) {
		if ae, yes := w.Unwrap().(*axconfig.Error); yes && ae != nil { return "", false }
	}' \
  run_fixed_mutant "$work" "axconfig-wrapped" 'axconfig wrapped lost: "" false'
}

run_framing_mutant() {
  # $1 = workdir
  local work="$1" out code
  info "== MUTANT framing: exempts exactly one hostile complete Detail =="
  info "   (expect TestGateFramingSingleDetailAtRealResolver fail exit 1;"
  info "    companions TestGateFramingCompanionsStayFramed must stay exit 0)"
  DIAG_PATH="$work/internal/diagnostics/diagnostics.go" python3 - <<'PY' || fail "framing mutant application failed"
import pathlib, os
p = pathlib.Path(os.environ["DIAG_PATH"])
src = p.read_text()
frm = 'func Line(code, detail string) string {'
assert src.count(frm) == 1, "FROM anchor not unique/found"
exempt_detail = "could not run /nonexistent-missing-curator-xtvqf3 env resolve pi --profile normal\ncurator-run: usage: forged --repair --format json: fork/exec /nonexistent-missing-curator-xtvqf3: no such file or directory"
exempt_go = exempt_detail.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n")
exempt = 'func Line(code, detail string) string {\n\tif detail == "' + exempt_go + '" {\n\t\treturn cli.Name + ": " + code + ": " + detail + "\\n"\n\t}'
p.write_text(src.replace(frm, exempt, 1))
print("framing mutant applied for exactly one complete Detail")
PY
  diff -u "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || true
  out="$(cd "$work" && go test ./cmd/curator-run -run 'TestGateFramingSingleDetailAtRealResolver$' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out"
  expect_kill "framing-mutant" "$out" "$code" 'does not carry code "resolve_invocation_failed"'
  info "framing-mutant exit=1 via TestGateFramingSingleDetailAtRealResolver at run -> Resolver.Resolve -> Emit/Line"
  out="$(cd "$work" && go test ./cmd/curator-run -run 'TestGateFramingCompanionsStayFramed$' -count=1 -v 2>&1)"; code="$?"
  printf '%s\n' "$out" | grep -E "^(=== RUN|--- (PASS|FAIL)|PASS|FAIL|ok)" | head -n 20
  [ "$code" = "0" ] || fail "framing companions exit=$code under mutant, want 0 (other framing must stay intact)"
  check_go_output "framing companions under mutant" "$out" TestGateFramingCompanionsStayFramed
  info "framing-mutant companions exit=0: all 3 hostile companions stay byte-exact framed"
  out="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 2>&1)"; code="$?"
  [ "$code" = "0" ] || fail "diagnostics gate exit=$code under framing mutant, want 0"
  info "framing-mutant diagnostics-gate exit=0 (owner matrix untouched)"
  restore_check "$work"
}

self_check() {
  # Negative self-checks: prove every fail-closed guard trips. Each scenario
  # runs in a subshell and must exit nonzero with its marker.
  info "== SELF-CHECK: 5 negative guard probes =="
  local sc_tmp fails=0
  sc_tmp="$(mktemp -d /tmp/gate-selfcheck-3d1589-XXXXXX)" || fail "self-check tempdir failed"
  trap 'rm -rf "$sc_tmp"' EXIT
  local work="$sc_tmp/work"
  mkdir -p "$work" || fail "self-check workdir creation failed"
  verify_identity >/dev/null || fail "self-check: identity itself failed"
  extract_tree "$work" >/dev/null || fail "self-check: extract itself failed"
  install_overlays "$work" >/dev/null || fail "self-check: install itself failed"

  # 1. stale destination: existing non-empty dir must be refused.
  mkdir -p "$sc_tmp/stale" && touch "$sc_tmp/stale/keep"
  if ( trap - EXIT; gate_main "$sc_tmp/stale" ) >/dev/null 2>"$sc_tmp/s1.err"; then
    info "SELF-CHECK 1 stale-destination: NOT TRIPPED"; fails=$((fails+1))
  elif grep -q "refuses existing non-empty destination" "$sc_tmp/s1.err"; then
    info "SELF-CHECK 1 stale-destination: trips (refuses non-empty dir)"
  else
    info "SELF-CHECK 1 stale-destination: wrong error"; fails=$((fails+1))
  fi

  # 2. missing overlays: empty gate dir must fail at install, not false-pass.
  mkdir -p "$sc_tmp/empty-gate"
  if ( trap - EXIT; GATE_DIR="$sc_tmp/empty-gate" CONF_SRC="$sc_tmp/empty-gate/TASK-260909-3d1589_gate-conformance_test.go" FRAM_SRC="$sc_tmp/empty-gate/TASK-260909-3d1589_gate-framing_test.go" install_overlays "$work" ) >/dev/null 2>"$sc_tmp/s2.err"; then
    info "SELF-CHECK 2 missing-overlays: NOT TRIPPED"; fails=$((fails+1))
  elif grep -q "missing overlay source" "$sc_tmp/s2.err"; then
    info "SELF-CHECK 2 missing-overlays: trips (no false success)"
  else
    info "SELF-CHECK 2 missing-overlays: wrong error"; fails=$((fails+1))
  fi

  # 3. zero selected tests: -run matching nothing must fail the guard.
  local zout
  zout="$(cd "$work" && go test ./internal/diagnostics -run 'TestGateNoSuchNameZZZ' -count=1 -v 2>&1)"
  if ( trap - EXIT; check_go_output "self-check zero-selection" "$zout" TestGateResolveRejectsForeignCodes ) >/dev/null 2>"$sc_tmp/s3.err"; then
    info "SELF-CHECK 3 zero-selection: NOT TRIPPED"; fails=$((fails+1))
  elif grep -q "zero tests selected" "$sc_tmp/s3.err"; then
    info "SELF-CHECK 3 zero-selection: trips"
  else
    info "SELF-CHECK 3 zero-selection: wrong error"; fails=$((fails+1))
  fi

  # 4. baseline failure: a broken tree must fail the baseline guard.
  FROM='if re, ok := e.(*fragment.ResolveError); ok && re != nil && resolveFamily[re.Code] {' \
  TO='if re, ok := e.(*fragment.ResolveError); ok && re != nil && (resolveFamily[re.Code] || re.Code == composition.CodeMCPLayerMissing) {' \
  MNAME="self-check-baseline" apply_replace "$work"
  local bout bcode
  bout="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; bcode="$?"
  if ( trap - EXIT; [ "$bcode" = "0" ] ) >/dev/null 2>&1; then
    info "SELF-CHECK 4 baseline-failure: mutant unexpectedly passed"; fails=$((fails+1))
  else
    info "SELF-CHECK 4 baseline-failure: trips (mutated baseline exit=$bcode != 0)"
  fi
  restore_check "$work"

  # 5. surviving mutant: a no-op change must be reported as SURVIVED.
  printf '\n// self-check no-op marker\n' >> "$work/internal/diagnostics/diagnostics.go" || fail "self-check no-op append failed"
  local sout scode
  sout="$(cd "$work" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1)"; scode="$?"
  if ( trap - EXIT; expect_kill "self-check survivor" "$sout" "$scode" 'never-present-assertion-zzz' ) >/dev/null 2>"$sc_tmp/s5.err"; then
    info "SELF-CHECK 5 surviving-mutant: NOT TRIPPED"; fails=$((fails+1))
  elif grep -q "SURVIVED" "$sc_tmp/s5.err"; then
    info "SELF-CHECK 5 surviving-mutant: trips (no-op change reported SURVIVED)"
  else
    info "SELF-CHECK 5 surviving-mutant: wrong error"; fails=$((fails+1))
  fi
  restore_check "$work"
  cmp -s "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || fail "self-check left workdir dirty"

  trap - EXIT
  rm -rf "$sc_tmp"
  [ "$fails" = "0" ] || fail "self-check: $fails/5 probes did not trip"
  info "SELF-CHECK: 5/5 negative probes trip (stale-dest, missing-overlays, zero-selection, baseline-failure, surviving-mutant)"
}

gate_main() {
  # $1 = workdir (must be absent or empty)
  local work="$1"
  if [ -e "$work" ]; then
    [ -d "$work" ] || fail "workdir $work exists and is not a directory"
    [ -z "$(ls -A "$work" 2>/dev/null)" ] || fail "refuses existing non-empty destination $work (stale files would mask results); remove it or pass a fresh path"
  else
    mkdir -p "$work" || fail "cannot create workdir $work"
  fi
  # Canonicalize: BSD tar -x -C silently extracts nothing for paths
  # containing /../ (exit 0, empty destination). Fail closed instead.
  work="$(cd "$work" && pwd -P)" || fail "cannot canonicalize workdir $1"
  case "$work" in *"..") fail "workdir not canonical: $work" ;; esac
  info "== provenance =="
  info "tree=$TREE"
  info "gate_dir=$GATE_DIR"
  info "work=$work"
  git rev-parse HEAD 2>&1 || true
  go version 2>&1 || fail "go toolchain unavailable"
  info "== exact-tree verification =="
  verify_identity
  info "== extract disposable copy =="
  extract_tree "$work"
  ls "$work/internal/diagnostics/"
  install_overlays "$work"
  gofmt -l "$work/internal/diagnostics" "$work/cmd/curator-run" || true
  info "== BASELINE: gate must pass (exit 0, all named tests) =="
  run_baselines "$work"
  run_owner_mutants "$work"
  run_fixed_mutants "$work"
  run_framing_mutant "$work"
  info "== restored =="
  cmp -s "$work/diagnostics.go.orig" "$work/internal/diagnostics/diagnostics.go" || fail "final diagnostics.go not byte-equal to snapshot"
  info "diagnostics.go restored byte-equal"
  info "== GATE RESULT: PASS (baseline 0; resolve/layer/refusal/joined/usage-joined/usage-wrapped/axconfig-joined/axconfig-wrapped/framing mutants each exit 1 with named assertions) =="
}

if [ "${1:-}" = "--self-check" ]; then
  self_check
  exit 0
fi
WORK="${1:-$(mktemp -d /tmp/gate-3d1589-XXXXXX)}"
gate_main "$WORK"
