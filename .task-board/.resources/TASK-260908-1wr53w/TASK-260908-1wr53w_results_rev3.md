# TASK-260908-1wr53w producer outcome — revision 3 (accepted-gate adoption)

Producer: developer (Muse Spark xhigh). Base `3ff66a9`, preserved CR2
tree `fbe90d5e60593a3a069721b2ad9e53cd071d8c02`. This revision adopts the
independently accepted owner/form gate (TASK-260909-3d1589 rev3,
review-verdict-rev3 `accepted`) to resolve repeated R3. No
redevelopment: R1 framing and R2 owner/nil runtime behavior are
byte-preserved (see blob check below).

## Candidate delta vs CR2 (exact)

Preserved byte-equal to CR2 (`git hash-object` == pinned blob):
`cmd/curator-run/main.go`, `cmd/curator-run/main_test.go`,
`internal/diagnostics/diagnostics.go`,
`internal/diagnostics/helpers_test.go`, `SPEC.md`, `README.md` (except
the mutant paragraph below).

Changed:
- `internal/diagnostics/gate_conformance_test.go` (NEW) — accepted
  gate conformance overlay committed verbatim under its adopted name;
  only the header comment is new (provenance, no overlay/do-not-commit
  wording). 8 TestGate tests: 3 per-owner foreign rejections, joined
  owned acceptance, 36 owner/form positives, 23 nil cases, coverage
  counts, own-family/nil preservation.
- `cmd/curator-run/gate_framing_test.go` (NEW) — accepted framing
  overlay committed verbatim under its adopted name (header only new).
  2 TestGateFraming tests at `run -> Resolver.Resolve (ExecRunner) ->
  Emit/Line` with the fixed absent binary.
- `internal/diagnostics/diagnostics_test.go` — `TestCodeOfRejectsForeignFamily`
  strangers now derived (SPEC §6 normative set minus per-owner family
  transcribed from owner constants, plus 4 extras) instead of the
  hand list that omitted 5 resolve codes for layer/refusal.
- `.scripts/diagnostics-mutants.sh` — 12 → 21 mutants: added D13–D21
  single-member narrowings adopted from the accepted gate (same diffs,
  same named assertions); D06/D12 honestly relabeled RETAINED broad
  (token-preserving but not single-member); summary gains a
  named_assertion column required for D13–D21.
- `README.md` — mutant paragraph: 21 mutants, 14 narrowing + 7
  retained; tools-table row names the owner/form and framing probes.

## AC coverage — 15 of 18 code rows driven; 8 of 18 through main

Reachability through production call sites (unchanged from rev2; no new
production path in this revision, no installed launcher claimed):

| Rows | Production call site | Named tests |
|---|---|---|
| usage (1) | run -> cli.Parse | TestRunDiagnosticsContract, TestRunUsageErrorsExit2, TestRunDetailInjectionCannotForgeLine/usage, TestGateOwnerFormPositives/usage |
| resolve (6) | run -> Resolver.Resolve | TestRunDiagnosticsContract, TestRunResolveFailuresExit1, TestRunDetailInjectionCannotForgeLine/resolve, TestGateFramingSingleDetailAtRealResolver, TestGateResolveRejectsForeignCodes |
| environment (1) | run -> mapping.Resolve | TestRunDiagnosticsContract, TestRunMapping |
| defaults_config_invalid (1) | axconfig.Load -> CodeOf in test | TestCodeOfProductionTypes/axconfigFailure, TestGateOwnerFormPositives/axconfig |
| mcp (2) | Compose -> CheckLaunchBoundary; Launch.Run late check | TestCodeOfProductionTypes/codexBoundary, TestLateChecksBothModes, TestGateLayerRejectsForeignCodes |
| system prompt (2) | Select / ProbeFiles -> CodeOf in test | TestCodeOfProductionTypes, TestGateRefusalRejectsForeignCodes |
| exec (1) | execution.Launch.Run | TestLateChecksBothModes |
| ax (1) | execution.Launch.Run | TestAxFailureNoFallback (fake-process JSON+NUL+non-UTF8, no fallback) |
| defaults_unresolvable, plan_refused, plan_provider_limited (3) | no producer in this build | stated bounds (TASK-260909-2vy977, TASK-260908-2so46q), not behavioral proof |

Child exit propagation (PR11) untouched: exit 23 / signal 143 preserved,
never routed through ExitForCode; ax Structured Error bytes verbatim.

## R3 evidence (this revision)

Producer harness `.scripts/diagnostics-mutants.sh` on the working tree:
exit 0, 21/21 KILLED, every mutant applied. 14 narrowing
(D01/D03/D09/D10/D11/D13–D21, the last nine with named single-member
assertions in-log); 7 retained broad/drop-one/format (D02/D04/D05/D06/
D07/D08/D12, labeled as such). Log dir `.temp/diagnostics-mutants/`
(`D*.log` + `summary.tsv`).
- D14/D15 (prior R3 survivors: layer/refusal admit
  resolve_invocation_failed) now fail BOTH the adopted
  TestGateLayer/TestGateRefusalRejectsForeignCodes AND the derived
  TestCodeOfRejectsForeignFamily — the derivation fix is proven
  independently of the overlay.
- D17 is the exact rev2 reviewer survivor (joined UsageError); D16 the
  rev1 joined-positive survivor; D21 the exact-Detail framing exemption
  at the real entry (companions asserted in the independent
  TestGateFramingCompanionsStayFramed).

Accepted gate replay (pinned CR2 tree, unchanged objects):
- Manual block from TASK-260909-3d1589_rev3_adoption.md: exit 0 —
  baseline 8+2 gate tests, all 9 gate narrowings exit 1 with named
  assertions. Log `.temp/TASK-260909-3d1589/gate-manual.log`.
- `--self-check`: exit 0, 5/5 negative probes trip. Log
  `.temp/TASK-260909-3d1589/gate-selfcheck.log`.
- New-tree gate coverage = committed TestGate runs on the working tree
  (green, see below) + D13–D21 harness kills on the working tree.

## Validation commands and real exits (Darwin arm64, Go 1.25.5)

- `go test ./internal/diagnostics/ ./cmd/curator-run/ -count=1`: exit 0
  (both ok; includes 10 TestGate* tests).
- `go test ./internal/diagnostics/ -run 'TestGate|TestCodeOfRejectsForeignFamily|TestClosedSet' -count=1 -v`: exit 0, all PASS.
- `go test ./cmd/curator-run/ -run TestGateFraming -count=1 -v`: exit 0.
- `.scripts/diagnostics-mutants.sh .temp/diagnostics-mutants`: exit 0.
- Gate manual block / `--self-check`: exit 0 / exit 0.
- `go vet ./internal/diagnostics/ ./cmd/curator-run/`: exit 0.
  `gofmt -l` over both dirs: empty. `git diff --check`: exit 0.
- Full `make check` deliberately left to the runtime handoff per the
  rework brief (no duplicate broad suite); rev2 `make check` evidence
  remains attached to the prior CR.

## Bounds (unchanged, carried explicitly)

- CodeOf is API-only: no non-test production caller; main classifies
  via cli.IsUsage / fragment.IsResolve and call-site codes.
- Remaining main wiring stays with TASK-260908-1o7i8y
  (production-main-obligations.md), defaults/lineup with
  TASK-260909-2vy977, plan/provider-limits with TASK-260908-2so46q.
- `not_implemented` refusal remains outside the closed registry until
  integration. No Pi scope/dependency change, no real ax/model calls,
  no installs/tags/hosted CI, no control-root writes. Candidate left
  UNCOMMITTED for the handoff snapshot.
