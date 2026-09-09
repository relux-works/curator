# Adopt the diagnostics conformance gate rev2 (CR2 follow-up)

Exact candidate tree: `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`
Base: `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`
Platform: Darwin arm64, Go 1.25.5. No production changes in this task.
Supersedes rev1 (`TASK-260909-xtvqf3_gate-*`); rev1 artifacts stay available.

## What to copy

The five rev2 resources below live adjacent to the runner. Names matter:
the runner refuses to run unless these exact filenames sit beside it.

- `TASK-260909-xtvqf3_rev2_gate-conformance_test.go` -> `internal/diagnostics/gate_conformance_test.go`
- `TASK-260909-xtvqf3_rev2_gate-framing_test.go` -> `cmd/curator-run/gate_framing_test.go`
- `TASK-260909-xtvqf3_rev2_vectors.json` -> reference only (vectors are derived at test runtime)
- `TASK-260909-xtvqf3_rev2_run-gate.sh` -> the runner (run from its own directory)
- `TASK-260909-xtvqf3_rev2_adoption.md` -> this file

Do not copy mutants; they are defined in the runner and vectors file.

## How to run (exact commands)

In a disposable copy of the exact tree (never the original CR workspace).
The runner requires a fresh destination: an existing non-empty directory is
refused. Create it first, or omit the argument for an automatic mktemp dir.

```bash
mkdir -p /tmp/cr2-gate && rm -rf /tmp/cr2-gate
git archive fbe90d5e60593a3a069721b2ad9e53cd071d8c02 | tar -x -C /tmp/cr2-gate
cp TASK-260909-xtvqf3_rev2_gate-conformance_test.go /tmp/cr2-gate/internal/diagnostics/gate_conformance_test.go
cp TASK-260909-xtvqf3_rev2_gate-framing_test.go /tmp/cr2-gate/cmd/curator-run/gate_framing_test.go
cd /tmp/cr2-gate
go test ./internal/diagnostics -run TestGate -count=1 -v     # expect exit 0, 6/6 named PASS
go test ./cmd/curator-run -run TestGateFraming -count=1 -v   # expect exit 0, 2/2 named PASS
```

Or run the full fail-closed gate from the directory holding the five files:

```bash
./TASK-260909-xtvqf3_rev2_run-gate.sh /tmp/cr2-gate-fresh
./TASK-260909-xtvqf3_rev2_run-gate.sh --self-check
```

Expected: baseline diagnostics (6 named TestGate tests) + framing (2 named
TestGateFraming tests) + full-package suites exit 0; resolve / layer /
refusal / joined-positive / framing mutants each exit 1 with the named
failure in the log; `--self-check` proves all 5 negative guards trip
(stale destination, missing overlays, zero selection, baseline failure,
surviving mutant) and exits 0.

## What the gate proves

- `TestGateResolve/Layer/RefusalRejectsForeignCodes`: every normative
  non-family code (SPEC section 6 derived) plus `environment_home_stale`,
  `environment_unknown`, `invented_code`, `""` yields no code in
  direct/wrapped/joined forms; every owned code stays accepted in all forms.
- `TestGateJoinedOwnedAccepted`: joined acceptance for every owned code of
  every mutable-Code owner; kills the rev1 reviewer survivor that rejected
  joined `resolve_invocation_failed`.
- `TestGateCoverageCounts`: pins 18 normative codes, 44 rejection pairs,
  168 cases with extras x forms. Honest scope: only the three mutable-Code
  owners (resolve 6, layer 2, refusal 2) carry single-foreign-code gates.
  `cli.UsageError` (constant code) and `axconfig.Error` (fixed mapping) have
  no expressible foreign admission; `defaults_unresolvable`, `plan_refused`,
  `plan_provider_limited`, `env_unsupported`, `exec_provider_missing`,
  `ax_handoff_failed` are call-site-selected with no CodeOf classification.
- `TestGateOwnFamilyAndNilPreserved`: production-typed values incl. joined,
  typed-nil fails closed in direct/wrapped/joined (incl. all-typed-nil Join).
- `TestGateFramingSingleDetailAtRealResolver`: the exact complete Detail
  for fixed binary `/nonexistent-missing-curator-xtvqf3` and profile
  `normal\ncurator-run: usage: forged` renders one line at
  `run -> Resolver.Resolve (ExecRunner) -> Emit/Line`.
- `TestGateFramingCompanionsStayFramed`: byte-exact framing of the 3 other
  hostile details; runs independently, stays green under the mutant.

## Single-member narrowings (all must fail, all do)

- resolve admits `mcp_layer_missing`
- layer admits `resolve_invocation_failed` (prior R3 survivor)
- refusal admits `resolve_invocation_failed` (prior R3 survivor)
- joined `resolve_invocation_failed` rejected (rev1 reviewer survivor)
- Line exempts exactly the complete Detail above via equality

Broad `IsDiagnosticLine` prefix-contains and whole-framing deletions remain
useful probes but are not narrowing; do not label them as such.

## Next revision checklist

1. Replace hand-selected `strangers` in `TestCodeOfRejectsForeignFamily`
   with the gate's derived foreign sets (or copy the gate file verbatim).
2. Add joined owned positives (`TestGateJoinedOwnedAccepted`) and joined
   typed-nil cases to the candidate's own tests.
3. Keep the framing unit matrix and add the fixed-binary main entry plus
   independent byte-exact companions.
4. Run the gate and `--self-check` before requesting review; attach the log.
