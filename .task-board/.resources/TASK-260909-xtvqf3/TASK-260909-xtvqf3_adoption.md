# Adopt the diagnostics conformance gate (CR2 follow-up)

Exact candidate tree: `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`
Base: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`
Platform: Darwin arm64, Go 1.25.5. No production changes in this task.

## What to copy

From the attached gate resources (all `TASK-260909-xtvqf3_*`):

- `TASK-260909-xtvqf3_gate-conformance_test.go` -> `internal/diagnostics/gate_conformance_test.go`
- `TASK-260909-xtvqf3_gate-framing_test.go` -> `cmd/curator-run/gate_framing_test.go`
- `TASK-260909-xtvqf3_vectors.json` -> reference only (vectors are derived at test runtime)
- `TASK-260909-xtvqf3_run-gate.sh` -> `.temp/` runner (or run inline)

Do not copy mutants; they are defined in the runner and vectors file.

## How to run

In a disposable copy of the exact tree (never the original CR workspace):

```bash
git archive fbe90d5e60593a3a069721b2ad9e53cd071d8c02 | tar -x -C /tmp/cr2-gate
cp TASK-260909-xtvqf3_gate-conformance_test.go /tmp/cr2-gate/internal/diagnostics/gate_conformance_test.go
cp TASK-260909-xtvqf3_gate-framing_test.go /tmp/cr2-gate/cmd/curator-run/gate_framing_test.go
cd /tmp/cr2-gate
go test ./internal/diagnostics -run TestGate -count=1 -v     # expect exit 0
go test ./cmd/curator-run -run TestGateFraming -count=1 -v   # expect exit 0
./run-gate.sh /tmp/cr2-gate                                  # baseline 0, 4 mutants 1
```

Or use the attached runner directly against git objects:

```bash
./TASK-260909-xtvqf3_run-gate.sh /tmp/cr2-gate
```

Expected: baseline diagnostics+framing exit 0; resolve/layer/refusal/framing
mutants each exit 1 with the named failure in the evidence log.

## What the gate proves

- `TestGateResolve/Layer/RefusalRejectsForeignCodes`: every normative
  non-family code (SPEC section 6 derived) plus `environment_home_stale`,
  `environment_unknown`, `invented_code`, `""` yields no code in
  direct/wrapped/joined forms; own family stays accepted.
- `TestGateOwnFamilyAndNilPreserved`: production-typed values, typed-nil
  fails closed, unclaimed errors invent nothing.
- `TestGateFramingSingleDetailAtRealResolver`: the exact complete Detail
  for fixed binary `/nonexistent-missing-curator-xtvqf3` and profile
  `normal\ncurator-run: usage: forged` renders one line at
  `run -> Resolver.Resolve (ExecRunner) -> Emit/Line`; companions stay framed.

## Single-member mutants (all must fail, all do)

- resolve admits `mcp_layer_missing`
- layer admits `resolve_invocation_failed` (reviewer survivor)
- refusal admits `resolve_invocation_failed` (reviewer survivor)
- Line exempts exactly the complete Detail above via equality

Broad `IsDiagnosticLine` prefix-contains and whole-framing deletions remain
useful probes but are not narrowing; do not label them as such.

## Next revision checklist

1. Replace hand-selected `strangers` in `TestCodeOfRejectsForeignFamily`
   with the gate's derived foreign sets (or copy the gate file verbatim).
2. Keep the framing unit matrix and add the fixed-binary main entry above.
3. Run the gate before requesting review; attach the runner log.
