# TASK-260908-1wr53w producer outcome — rev4 gate-adoption validation (this run)

Producer: developer (Muse Spark xhigh). Base `3ff66a9`, branch
`task-board/story/STORY-260908-18kdnq`. Candidate left UNCOMMITTED for
the handoff snapshot. This run resumes the preserved CR2 tree
`fbe90d5e60593a3a069721b2ad9e53cd071d8c02` with the independently
accepted gate (TASK-260909-3d1589 rev3 `accepted`,
review-verdict-rev3) adopted per `accepted-gate-adoption-rework.md`.
No redevelopment of R1 framing or R2 owner/nil runtime behavior; no Pi
scope/dependency change, no real ax/model calls, no installs, no hosted
CI, no tags, no runtime-home/private/control-root writes.
Platform: Darwin arm64, Go 1.25.5.

## Gate adoption provenance (verified this run)

- `internal/diagnostics/gate_conformance_test.go` vs board
  `TASK-260909-3d1589_rev2_gate-conformance_test.go` (`8495946d…`):
  diff is header-only (16 lines), test bodies byte-identical.
- `cmd/curator-run/gate_framing_test.go` vs board
  `TASK-260909-3d1589_rev2_gate-framing_test.go` (`5031d27f…`):
  diff is header-only (16 lines), test bodies byte-identical.
- `internal/diagnostics/diagnostics_test.go`
  `TestCodeOfRejectsForeignFamily`: strangers derived from normative
  SPEC §6 set minus per-owner family (transcribed from owner
  constants) plus 4 extras — the hand-list hole (5 resolve codes
  missing for layer/refusal) is closed.
- `.scripts/diagnostics-mutants.sh`: 21 mutants, 14 narrowing
  (D01/D03/D09/D10/D11/D13–D21, D13–D21 with required named
  single-member assertion) + 7 retained broad/drop-one/format
  (D02/D04/D05/D06/D07/D08/D12, labeled RETAINED, not narrowing).
- `README.md`: mutant paragraph states 21 mutants (14 narrowing + 7
  retained) and names the owner/form + framing probes.

## AC coverage — 15 of 18 code rows driven; 8 of 18 through main

| Rows | Production call site | Named tests |
|---|---|---|
| usage (1) | run -> cli.Parse | TestRunDiagnosticsContract, TestRunUsageErrorsExit2, TestRunDetailInjectionCannotForgeLine/usage, TestGateOwnerFormPositives/usage |
| resolve (6) | run -> fragment.Resolver.Resolve | TestRunDiagnosticsContract, TestRunResolveFailuresExit1, TestRunDetailInjectionCannotForgeLine/resolve, TestGateFramingSingleDetailAtRealResolver, TestGateResolveRejectsForeignCodes |
| environment (1) | run -> mapping.Resolve | TestRunDiagnosticsContract, TestRunMapping |
| defaults_config_invalid (1) | axconfig.Load -> CodeOf in test | TestCodeOfProductionTypes/axconfigFailure, TestGateOwnerFormPositives/axconfig |
| mcp (2) | Compose -> CheckLaunchBoundary; Launch.Run late check | TestCodeOfProductionTypes/codexBoundary, TestLateChecksBothModes, TestGateLayerRejectsForeignCodes |
| system prompt (2) | Select / ProbeFiles -> CodeOf in test | TestCodeOfProductionTypes, TestGateRefusalRejectsForeignCodes |
| exec (1) | execution.Launch.Run | TestLateChecksBothModes |
| ax (1) | execution.Launch.Run | TestAxFailureNoFallback (fake-process JSON+NUL+non-UTF8, no fallback) |
| defaults_unresolvable, plan_refused, plan_provider_limited (3) | no producer in this build | stated bounds (TASK-260909-2vy977, TASK-260908-2so46q), not behavioral proof |

Child exit propagation (PR11) untouched: exit 23 / signal 143
preserved, never routed through ExitForCode; ax Structured Error bytes
verbatim. SPEC §6 completeness is source-derived
(TestClosedSetMatchesSpec parses the Codes column; 18 codes).

## Validation reran in THIS run (real exits)

- `go test ./internal/diagnostics/ ./cmd/curator-run/ -count=1`: exit 0
  (both ok).
- `go test ./internal/diagnostics/ -run
  'TestGate|TestCodeOfRejectsForeignFamily|TestClosedSet' -count=1 -v`:
  exit 0 — TestClosedSetMatchesSpec, TestCodeOfRejectsForeignFamily,
  TestGateResolve/Layer/RefusalRejectsForeignCodes,
  TestGateJoinedOwnedAccepted, TestGateOwnerFormPositives,
  TestGateOwnerFormNil, TestGateCoverageCounts,
  TestGateOwnFamilyAndNilPreserved, all PASS.
- `go test ./cmd/curator-run/ -run TestGateFraming -count=1 -v`:
  exit 0 — TestGateFramingSingleDetailAtRealResolver PASS,
  TestGateFramingCompanionsStayFramed PASS (3/3 companions).
- `.scripts/diagnostics-mutants.sh .temp/diagnostics-mutants`: exit 0,
  21/21 KILLED, every mutant applied; D13–D21 each
  `expected_test_failed=yes, named_assertion=yes` (log dir
  `.temp/diagnostics-mutants/`, `D*.log` + `summary.tsv`).
- R3-survivor spot checks from those logs (this run):
  D14/D15 each fail BOTH TestGateLayer/TestGateRefusalRejectsForeignCodes
  AND the derived TestCodeOfRejectsForeignFamily (named `layer/refusal
  accepted foreign "resolve_invocation_failed"`, 3 hits each) — the
  derivation fix is proven independently of the overlay.
  D21 fails only TestGateFramingSingleDetailAtRealResolver with
  `does not carry code "resolve_invocation_failed"` while
  TestGateFramingCompanionsStayFramed stays PASS — exact single-detail
  exemption narrowing.
- `go test ./internal/execution/ -run
  'Test(AxFailureNoFallback|DirectExitAndSignal|LateChecksBothModes)$'
  -count=1`: exit 0 (execution preservation, fake-process only).
- `go vet ./internal/diagnostics/ ./cmd/curator-run/`: exit 0.
  `gofmt -l` over both dirs: empty. `git diff --check`: exit 0.
- Post-harness `git status`: only intended candidate files differ
  (M README.md, M cmd/curator-run/main.go, M cmd/curator-run/main_test.go;
  untracked .scripts/diagnostics-mutants.sh,
  cmd/curator-run/gate_framing_test.go, internal/diagnostics/).

## Inherited, not rerun (explicit)

- Full `make check` (build, fmt-check, vet, tests, race) NOT rerun here
  per the adoption brief (no duplicate broad suite); rev2 runtime
  `TASK-260908-1wr53w_change-request_rev2-validation.log` (`make check`
  exit 0) remains the attached broad-suite evidence. No runtime source
  changed since CR2 except test/harness/README-strings, and the focused
  suites + vet + harness above are green on this tree.
- Gate runner manual/automatic blocks and `--self-check` (rev3 gate
  evidence) inherited from accepted TASK-260909-3d1589 rev3; new-tree
  gate coverage here is the committed TestGate runs + D13–D21 kills.
- Execution tests above are fake-process API checks, not installed
  ax integration; provider-limits verdict transport remains a future
  obligation, not inferred.

## Bounds (carried explicitly)

- CodeOf is API-only: no non-test production caller; main classifies via
  cli.IsUsage / fragment.IsResolve and call-site codes.
- Remaining main wiring stays with TASK-260908-1o7i8y
  (production-main-obligations.md), defaults/lineup with
  TASK-260909-2vy977, plan/provider-limits with TASK-260908-2so46q.
- `not_implemented` refusal remains outside the closed registry until
  integration. Next review: independent Astra medium.
