# TASK-260908-1wr53w independent review — revision 1

Verdict: **changes_requested**. repeat-of: none.
Candidate tree: 1839f385bcb3d966620f0fad5ea5aad32b072953.
Base: 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3. All 7 changed files verified byte-for-byte against candidate blobs. HEAD..main = 0 (local refs only; no remote freshness claim). Darwin arm64, Go 1.25.5. No source edits, commits, installs, runtime-home changes, real ax/model calls or control-root source writes.

## Findings requiring rework

### R1 — high: one-code-line contract fails at an existing production entry
`internal/diagnostics/diagnostics.go:196-204` concatenates detail without framing its continuation lines. `run -> fragment.Resolver.Resolve -> ExecRunner` with a missing Curator binary and profile value `normal\ncurator-run: usage: forged` returns operational exit 1 but prints TWO recognized launcher diagnostic lines. The injected line is inside launcher-owned detail, not forwarded subprocess bytes. Independent `TestReviewOneLineAtRealResolver` fails with diagnostic_lines=2. This is within the current diagnostics division, not pending pipeline work. Frame/escape launcher-owned detail consistently, document the byte rule, and add a production-entry regression and narrowing mutant. Do not transform forwarded child stderr, provider verdict evidence, or ax Structured Error bytes as a side effect. The upstream detail construction predates this candidate, but the new shared renderer claims and must establish SPEC6's invariant.

### R2 — medium: new closed-code classifier accepts non-family codes and panics on typed nil
`internal/diagnostics/diagnostics.go:159-187`: a wrapped `fragment.ResolveError{Code:"invented_code"}` is accepted, as is `{Code:"usage"}` (which maps an operational typed error to exit 2). Empty-only checks similarly trust LayerError and Refusal codes. A typed-nil ResolveError panics at line 166. `TestReviewUnknownTypedCode` and `TestReviewTypedNil` fail. Enforce the owner's closed family, handle typed-nil values safely, and test wrapped invalid/foreign-family errors for each owner plus narrowing mutants. CodeOf has no non-test caller today: this is an exported diagnostic-contract API defect, not evidence of a currently reached full-main classifier failure. Keep that integration bound honest.

### R3 — medium: mutant evidence overclaims narrowing coverage
Producer logs show 8 of 8 mutants killed, but D02 replaces the entire operational exit arm with zero; D04 replaces the whole unknown-error fallback; D07 replaces every typed resolve failure return with zero. These are whole-clause replacements, not the required weakening admitting exactly one rejected member. D08 is formatting mutation. Retain these useful tests but add true single-member narrowing probes for the unknown-code terminal gate, no-invention classifier and main operational exit gate. Include R1/R2 regressions. D03 is a valid single-code membership widening; D05 specifically drops mcp_layer_missing; D06 exercises token-preserving anchoring behavior. Do not label all eight narrowing. `TestClosedSetMatchesSpec` also hand-enumerates its expected set rather than deriving it from the normative SPEC; make completeness independently source-derived as required by negative-evidence.md.

## Coverage and division of work

**15 of 18 diagnostic-code AC rows driven at stage production APIs; 8 of 18 driven through main run; 3 of 18 explicitly pending.** This is reachability coverage, not acceptance of all invariants (R1/R2 fail). New tests are candidate files, not yet committed; managed snapshot delivery must preserve them.

| Rows | Production caller | Named test/evidence |
|---|---|---|
| usage (1) | run -> cli.Parse | TestRunDiagnosticsContract; TestRunUsageErrorsExit2 |
| resolve (6) | run -> Resolver.Resolve | TestRunDiagnosticsContract; TestRunResolveFailuresExit1; reviewer TestReviewOneLineAtRealResolver FAIL |
| environment (1) | run -> mapping.Resolve | TestRunDiagnosticsContract |
| defaults_config_invalid (1) | axconfig.Load -> CodeOf in test only | TestCodeOfProductionTypes / axconfigFailure |
| mcp (2) | Compose -> CheckLaunchBoundary -> CodeOf in test; Launch.Run late check | TestCodeOfProductionTypes / codexBoundary; TestLateChecksBothModes |
| system prompt (2) | Select / ProbeFiles -> CodeOf in test only | TestCodeOfProductionTypes / syspromptSelection / syspromptProbe |
| exec (1) | execution.Launch.Run | TestLateChecksBothModes |
| ax (1) | execution.Launch.Run | TestAxFailureNoFallback |
| defaults_unresolvable, plan_refused, plan_provider_limited (3) | no producer in this build | declared bounds, not a production behavioral proof |

Missing-vs-directory/readable MCP and absent-vs-broken config / absent-vs-directory system prompt are exercised by real filesystem APIs. No installed working launcher is established. The pre-existing main not_implemented path remains outside the closed registry; it must disappear with finished pipeline wiring, and is not a completed SPEC6 launch path.

Exact remaining owners: TASK-260908-1o7i8y must bind axconfig.Load before cli.Parse and the composed boundary + PrepareLaunch before execution; TASK-260909-2vy977 owns actual defaults/Lineup and member origin diagnostics; TASK-260908-2so46q owns plan admission and provider-limits enforcement with verbatim evidence. Producer README incorrectly groups defaults solely under integration and lists 2vy977 alongside plan admission: correct these to the existing task descriptions. Query confirmed integration's current description only says install/real launches and has no attached resources; named call-site obligations must be carried into its tracked inputs, not assumed from its title. No new Pi/MCP support or dependency change is authorized here.

Actual ax Structured Error transport is implemented in unchanged execution.Launch.Run: on nonzero handoff it writes one ax_handoff_failed line then axStderr.Bytes unchanged. Independent tests verify JSON plus NUL/non-UTF8 bytes and no fallback. This is fake-process API evidence, not real installed ax plumbing. Direct child exits 23 and signal 143 remain preserved; never route a launched child's status through ExitForCode.

## Validation and provenance

- Independently reran `go test ./internal/diagnostics ./cmd/curator-run -count=1 -v`: exit 0.
- Independently reran `go test ./internal/execution -run 'Test(AxFailureNoFallback|DirectExitAndSignal|LateChecksBothModes)$' -count=1 -v`: exit 0.
- Independent overlay tests (no candidate modifications): `go test -overlay .temp/TASK-260908-1wr53w-review/overlay.json ./internal/diagnostics ./cmd/curator-run -run TestReview -count=1 -v`: exit 1; all three named probes expose the failures above.
- Accepted attached runtime `TASK-260908-1wr53w_change-request_rev1-validation.log`: make check exit 0, build/fmt/vet/test/race all passed. Did not redundantly replay the whole suite.
- Inspected all eight existing mutant logs and their actual diffs/named failures, each suite exit 1; did not mutate candidate source or rerun the source-writing harness. Original logs copied into the attached evidence bundle before handoff.
- Producer outcome `TASK-260908-1wr53w_results.md` inspected; its 8/8 killed count is true but its 8/8 narrowing characterization is false.

No external blocker. Return to developer for R1–R3; do not accept CR revision 1 or close the task.
