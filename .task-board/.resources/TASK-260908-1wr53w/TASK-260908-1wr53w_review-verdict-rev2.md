# TASK-260908-1wr53w independent review — revision 2

Verdict: **changes_requested**.
repeat-of: TASK-260908-1wr53w_review-verdict-rev1.md / R3 (narrowing evidence and incomplete gate coverage).
CR: CR-TASK-260908-1wr53w-2 revision 2.
Candidate tree: fbe90d5e60593a3a069721b2ad9e53cd071d8c02.
Base/HEAD: 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3.

All seven candidate paths inspected and verified byte-for-byte against the tree before and after checks. Producer main.go mutant baseline equals this candidate; diagnostics.go baseline differs only in RemainingObligations owner strings (diff attached), so producer mutant evidence predates that ownership-only correction. All mutated logic is identical; independent overlays ran against the exact candidate. HEAD..main = 0, local refs only. Darwin arm64, Go 1.25.5. No candidate edits, commits, installs, hosted CI, real ax/model calls, runtime-home changes or control-root source writes. Additional tests and mutants used read-only Go overlays under .temp/ only.

## Required rework: R3 persists (medium)

1. `.scripts/diagnostics-mutants.sh` D12 replaces `strings.ReplaceAll(detail, "\n", "\n  ")` with `strings.ReplaceAll(detail, "\n", "\n")`. This disables framing for ALL multiline details, not one rejected member. It is a useful token-preserving behavioral probe, but cannot satisfy R1's required narrowing mutant. D06 replaces prefix matching with substring matching and admits arbitrary prefixes; it is likewise broad, not single-member. Producer claims seven narrowing probes; only D01/D03/D09/D10/D11 are single-code/member widenings. All 12 producer mutants do fail their named tests (suite exit 1); 5 of 12 qualify as single-member narrowing. Preserve the useful broad probes, correct their labels, and add a true single-member framing probe exercised through TestRunDetailInjectionCannotForgeLine at run -> Resolver.Resolve -> ExecRunner -> diagnostics.Emit/Line. Token preservation does not make deletion-equivalent behavior narrowing.

2. `internal/diagnostics/diagnostics_test.go:TestCodeOfRejectsForeignFamily` still hand-selects its strangers list, omitting five resolve codes for LayerError and Refusal. The normative closed-set test is now source-derived, but this enforcement matrix is not. Independently weakened each owner gate separately in diagnostics.CodeOf:
   - LayerError accepts its family OR fragment.CodeInvocationFailed.
   - Refusal accepts its family OR fragment.CodeInvocationFailed.
   Both exact diffs are attached. Each admits exactly one previously rejected code and leaves the gate in place. The ENTIRE candidate diagnostics + main behavioral suite exits **0** for each mutant: **0 of 2 independent narrowing mutants killed**. An independent TestReviewAllNormativeForeignCodes, deriving codes through the candidate's SPEC reader, passes against the unmodified candidate (exit 0) and fails against each mutant (exit 1), explicitly reporting the foreign resolve_invocation_failed admission. This verifies that the green mutant results are actual coverage holes, not unapplied overlays. Add source-derived foreign-code vectors for every code-bearing owner, preserve direct/wrapped/joined and own-family coverage, and ship named narrowing probes for each owner gate. D10 only attacks ResolveError; D05 drops an accepted LayerError member and does not establish rejection coverage.

This is repeat-of R3, not evidence that the original R2 implementation bug remains: the actual candidate correctly rejects these foreign values. It is a mandatory negative-evidence acceptance failure. Under the repeated-finding skill contract, the parent should route a focused conformance/mutant gate task for this class before another rework/review cycle, rather than another hand-picked-case patch.

## Confirmed fixes and preserved behavior

- R1 implementation passes: old TestReviewOneLineAtRealResolver now passes at actual run -> real Resolver -> ExecRunner with a missing binary and injected multiline profile. Candidate TestRunDetailInjectionCannotForgeLine passes. Line documents CRLF normalization, CR normalization and two-space continuation framing. No forwarded bytes enter that renderer at current main call sites. The narrowing-proof portion remains open above.
- R2 implementation passes: wrapped unknown/wrong-family values, all valid owner members and direct/wrapped typed nil for five owners pass candidate tests; old reviewer unknown-code and typed-nil probes also pass. CodeOf remains API-only, with no non-test production caller. Its guarded traversal recognizes both single and joined wrap chains.
- SPEC completeness is now independently sourced from the normative section 6 Codes column, and matches all 18 codes. Usage 2, operational/unknown 1, terminal refusal behavior, actual filesystem absence versus malformed/unreadable/directory states pass focused tests.
- README ownership now matches TASK-260909-2vy977 defaults/Lineup, TASK-260908-2so46q plan/provider limits, and TASK-260908-1o7i8y full main. The latter's existing production-main-obligations.md explicitly carries policy-before-parse, late boundaries, warning/diagnostic and byte/status preservation obligations.
- execution.Launch.Run is unchanged from accepted PR11. Independently rerun TestAxFailureNoFallback proves exact fake-ax Structured Error JSON + NUL + non-UTF8 bytes and no direct fallback. TestDirectExitAndSignal proves child exit 23 and signal exit 143 remain unchanged, not remapped through ExitForCode. TestLateChecksBothModes passes the refusal matrix. These are fake-process API checks, not installed ax integration.

## AC reachability and bounds

**15 of 18 diagnostic-code AC rows driven at stage production APIs; 8 of 18 driven through main; 3 of 18 explicitly pending.** This measures reachability, not satisfaction of the failed mutant gates. Candidate tests are preserved in the managed CR tree, not yet committed; signed publication remains producer/parent owned.

| Rows | Production call site | Named tests |
|---|---|---|
| usage (1) | run -> cli.Parse | TestRunDiagnosticsContract, TestRunUsageErrorsExit2, TestRunDetailInjectionCannotForgeLine/usage |
| resolve (6) | run -> fragment.Resolver.Resolve | TestRunDiagnosticsContract, TestRunResolveFailuresExit1, TestRunDetailInjectionCannotForgeLine/resolve, TestReviewOneLineAtRealResolver |
| environment (1) | run -> mapping.Resolve | TestRunDiagnosticsContract |
| defaults_config_invalid (1) | axconfig.Load -> CodeOf in test | TestCodeOfProductionTypes / axconfigFailure |
| mcp (2) | Compose -> CheckLaunchBoundary; execution.Launch.Run | TestCodeOfProductionTypes / codexBoundary; TestLateChecksBothModes |
| system prompt (2) | Select / ProbeFiles -> CodeOf in test | TestCodeOfProductionTypes / syspromptSelection / syspromptProbe |
| exec (1) | execution.Launch.Run | TestLateChecksBothModes |
| ax (1) | execution.Launch.Run | TestAxFailureNoFallback |
| defaults_unresolvable, plan_refused, plan_provider_limited (3) | no producer in this tree | stated future owners, not behavioral evidence |

No full-main launch or installed working launcher is established. The pre-existing not_implemented refusal remains outside the closed registry until integration. Provider-limits verdict byte transport has no current production path here; its preservation is an explicit future obligation, not inferred from tests or code absence. New defaults/plan/main work is not waived. No Pi/MCP support or dependency changes are introduced.

## Validation and evidence

Attached TASK-260908-1wr53w_review-evidence-rev2.txt includes independent logs, overlay source/maps/diffs, original producer mutant logs with their diffs and named failures, hashes and exact-tree verification.

- Independently ran `go test ./internal/diagnostics ./cmd/curator-run -count=1 -v`: exit 0.
- Independently replayed prior review overlay with `-run TestReview -count=1 -v`: exit 0.
- Independently ran `go test ./internal/execution -run 'Test(AxFailureNoFallback|DirectExitAndSignal|LateChecksBothModes)$' -count=1 -v`: exit 0.
- Independently ran full diagnostics/main behavioral suites with layer-mutant.json and refusal-mutant.json separately, `-count=1 -v`: both exit 0 (SURVIVED; acceptance failure).
- Independent normative probe with normative.json: exit 0. With layer-attacked.json and refusal-attacked.json: each exit 1, expected named rejection assertion failure.
- `git diff --check`: exit 0. Reused exact rev2 runtime TASK-260908-1wr53w_change-request_rev2-validation.log: `make check` exit 0 (build, formatting, vet, tests, race). Did not redundantly replay make check or execute the candidate-writing producer harness.

No external blocker. Do not accept rev2; route to-dev with the repeated evidence-gate finding. R1/R2 functionality is improved, but this does not waive R3 or permit closure.

## Lifecycle receipt

Both review artifacts attached before routing. Live checklist items 2 (acceptance) and 8 (narrowing evidence) unchecked to reflect the rejection. `set_status(..., status=to-dev)` succeeded (exit 0). The requested `task-board handoff TASK-260908-1wr53w --role reviewer` was attempted and returned exit 1: `role "reviewer" has no end_status and cannot use handoff`. No acceptance or commit acknowledgement was supplied; the supported rejection route remains to-dev. This is a role configuration limitation of the extra handoff command, not a blocker to the recorded changes-requested verdict.
