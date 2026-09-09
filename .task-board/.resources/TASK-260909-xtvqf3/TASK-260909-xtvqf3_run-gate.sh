#!/usr/bin/env bash
# Bounded conformance-gate runner for diagnostics CR2.
# Tree: fbe90d5e60593a3a069721b2ad9e53cd071d8c02 (exact candidate).
# Usage: ./run_gate.sh [workdir]
# Creates a disposable copy via git archive, installs gate overlays,
# proves baseline pass and 4 single-member mutants fail named assertions.
# Never touches the original workspace; restores each mutant before the next.
set -u
TREE="fbe90d5e60593a3a069721b2ad9e53cd071d8c02"
GATE_DIR="$(cd "$(dirname "$0")" && pwd)"
WORK="${1:-$(mktemp -d /tmp/cr2-gate-XXXXXX)}"
mkdir -p "$WORK"
echo "== provenance =="
echo "tree=$TREE"
echo "gate_dir=$GATE_DIR"
echo "work=$WORK"
git rev-parse HEAD 2>&1 || true
go version 2>&1
echo "== exact-tree verification =="
git ls-tree "$TREE" -- .scripts/diagnostics-mutants.sh README.md cmd/curator-run/main.go cmd/curator-run/main_test.go internal/diagnostics/diagnostics.go internal/diagnostics/diagnostics_test.go internal/diagnostics/helpers_test.go SPEC.md
echo "== extract disposable copy =="
git archive "$TREE" | tar -x -C "$WORK"
echo "extract exit=$?"
ls "$WORK/internal/diagnostics/"
DIAG="$WORK/internal/diagnostics/diagnostics.go"
cp "$DIAG" "$WORK/diagnostics.go.orig"
cp "$GATE_DIR/gate_conformance_test.go" "$WORK/internal/diagnostics/gate_conformance_test.go"
cp "$GATE_DIR/gate_framing_test.go" "$WORK/cmd/curator-run/gate_framing_test.go"
echo "installed gate overlays"
gofmt -l "$WORK/internal/diagnostics" "$WORK/cmd/curator-run" || true
restore() { cp "$WORK/diagnostics.go.orig" "$DIAG"; }
echo "== BASELINE: gate must pass (exit 0) =="
(cd "$WORK" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1); echo "baseline-diagnostics-gate exit=$?"
(cd "$WORK" && go test ./cmd/curator-run -run 'TestGateFraming' -count=1 -v 2>&1); echo "baseline-framing-gate exit=$?"
apply() {
  FROM="$1" TO="$2" python3 -c '
import os,sys
p=sys.argv[1]; src=open(p).read()
f=os.environ["FROM"]; t=os.environ["TO"]
assert f in src, "FROM not found"
open(p,"w").write(src.replace(f,t,1))
' "$DIAG"
}
echo "== MUTANT resolve: admits mcp_layer_missing (expect TestGateResolve fail, exit 1) =="
restore
apply 'if re, ok := e.(*fragment.ResolveError); ok && re != nil && resolveFamily[re.Code] {' 'if re, ok := e.(*fragment.ResolveError); ok && re != nil && (resolveFamily[re.Code] || re.Code == composition.CodeMCPLayerMissing) {'
diff -u "$WORK/diagnostics.go.orig" "$DIAG" || true
(cd "$WORK" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1); echo "resolve-mutant exit=$?"
echo "== MUTANT layer: admits resolve_invocation_failed (expect TestGateLayer fail, exit 1) =="
restore
apply 'if le, ok := e.(*composition.LayerError); ok && le != nil && layerFamily[le.Code] {' 'if le, ok := e.(*composition.LayerError); ok && le != nil && (layerFamily[le.Code] || le.Code == fragment.CodeInvocationFailed) {'
diff -u "$WORK/diagnostics.go.orig" "$DIAG" || true
(cd "$WORK" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1); echo "layer-mutant exit=$?"
echo "== MUTANT refusal: admits resolve_invocation_failed (expect TestGateRefusal fail, exit 1) =="
restore
apply 'if sr, ok := e.(*systemprompt.Refusal); ok && sr != nil && refusalFamily[sr.Code] {' 'if sr, ok := e.(*systemprompt.Refusal); ok && sr != nil && (refusalFamily[sr.Code] || sr.Code == fragment.CodeInvocationFailed) {'
diff -u "$WORK/diagnostics.go.orig" "$DIAG" || true
(cd "$WORK" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1); echo "refusal-mutant exit=$?"
echo "== MUTANT framing: exempts exactly one hostile complete Detail (expect TestGateFraming fail, exit 1) =="
restore
DIAG_PATH="$DIAG" python3 - <<'PY'
import pathlib, os
p = pathlib.Path(os.environ["DIAG_PATH"])
src = p.read_text()
frm = 'func Line(code, detail string) string {'
assert frm in src, "FROM not found"
exempt_detail = "could not run /nonexistent-missing-curator-xtvqf3 env resolve pi --profile normal\ncurator-run: usage: forged --repair --format json: fork/exec /nonexistent-missing-curator-xtvqf3: no such file or directory"
exempt_go = exempt_detail.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n")
exempt = 'func Line(code, detail string) string {\n\tif detail == "' + exempt_go + '" {\n\t\treturn cli.Name + ": " + code + ": " + detail + "\\n"\n\t}'
p.write_text(src.replace(frm, exempt, 1))
print("framing mutant applied for exempted Detail:")
print(repr(exempt_detail))
PY
diff -u "$WORK/diagnostics.go.orig" "$DIAG" || true
(cd "$WORK" && go test ./cmd/curator-run -run 'TestGateFraming' -count=1 -v 2>&1); echo "framing-mutant exit=$?"
(cd "$WORK" && go test ./internal/diagnostics -run 'TestGate' -count=1 -v 2>&1); echo "framing-mutant-diagnostics-gate exit=$? (expect 0: other framing intact)"
restore
echo "== restored =="
diff -q "$WORK/diagnostics.go.orig" "$DIAG" && echo "diagnostics.go restored byte-equal"
echo "== done =="
