# TASK-260908-1wr53w independent exact-CR3 review

Verdict: accepted. repeat-of: none surviving. Prior repeated R3 is resolved by adoption of the independently accepted gate; R1/R2 remain fixed.

CR-TASK-260908-1wr53w-3, revision 3 (not the producer outcome filename rev4).
Base/HEAD: 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3.
Candidate tree: 489e695df7ada7598347233c6553b791dacebb60.
All nine changed paths inspected and verified byte-for-byte before and after review. Main and diagnostics runtime blobs equal CR2 and the producer mutant baselines. Both adopted gate bodies equal accepted TASK-260909-3d1589 rev2 resource bodies from package declaration onward; only headers differ. The accepted gate rev3 verdict was read and its unchanged-artifact evidence reused. HEAD..main=0 is local-ref evidence only. Darwin arm64, Go 1.25.5.

## Findings and negative evidence

No blocking finding remains within the original explicit diagnostics/API division.

- R1: actual run -> Resolver.Resolve -> ExecRunner -> Emit/Line rejects forged launcher-owned multiline detail. CRLF/CR normalize to LF; continuation LF is followed by two spaces. Exact-Detail D21 exempts one complete string and is caught by TestGateFramingSingleDetailAtRealResolver; TestGateFramingCompanionsStayFramed remains green. No forwarded bytes pass through Line at current main call sites. Existing execution transport remains unchanged.
- R2: CodeOf recognizes five concrete owners safely through direct/wrapped/joined forms. Three mutable owners reject the full source-derived normative complement plus empty/unknown/Curator-internal extras. Fixed usage/axconfig owners have no arbitrary Code field; their correct mapping survives all three forms. Typed nils do not panic. The adopted loops exercise 36 positive cells, 23 nil cases, 44 normative foreign owner/code pairs and 168 rejection cases including extras/forms. Coverage arithmetic assertions are not independent execution counters.
- R3: TestCodeOfRejectsForeignFamily now derives strangers from SPEC section 6, not a selected list. D14/D15 independently replayed against actual CR3 each fail both the per-owner gate and this original derived test on resolve_invocation_failed. D13 attacks the resolve owner's foreign admission. D16 attacks joined resolve acceptance; D17-D20 attack usage/axconfig wrapped/joined cells. D21 is an equality-based framing exemption, not whole-arm replacement.
- Inspected all 21 producer diffs and named failures; each producer suite exit is 1. Independently replayed D13-D21 with Go overlays against actual CR3: 9 of 9 killed, each exit 1 with its named failure; overlay orchestration exit 0. Candidate files were never mutated. The harness's 14 single-member probes consist of D01/D03/D09/D10/D11/D13-D15/D21 admission/exit/framing weakenings and D16-D20 positive owner/form losses. The latter prove preservation, not foreign rejection. Seven retained broad/drop/format probes D02/D04/D05/D06/D07/D08/D12 are not counted as refusal-narrowing evidence. D06/D12 preserve tokens and execute behavior, but remain broad.
- Unknown classification produces no invented code; unknown ExitForCode remains operational 1. Usage is 2. Real filesystem tests distinguish missing layer from directory/unreadable layer, absent configuration from malformed configuration, and absent prompt file from non-regular prompt file. No terminal diagnostic silently retries or degrades in the reviewed paths.
- TestAxFailureNoFallback independently passes for nonzero/missing/nonexec/bad-format fake ax, preserving JSON plus NUL and non-UTF8 stderr exactly. TestDirectExitAndSignal independently preserves child 23 and signal 143. Child status must not be mapped through launcher ExitForCode. TestLateChecksBothModes independently passes.

## AC reachability and explicit bounds

15 of 18 diagnostic-code AC rows driven at stage production APIs; 8 of 18 through main; 3 of 18 have no producer here and remain explicitly pending. These denominators concern diagnostic-code rows, not every textual subcondition of SPEC. Tests are in the managed candidate tree awaiting signed producer delivery, not yet committed.

| Rows | Production call site | Named tests |
|---|---|---|
| usage (1) | run -> cli.Parse | TestRunDiagnosticsContract, TestRunUsageErrorsExit2, TestRunDetailInjectionCannotForgeLine |
| resolve (6) | run -> fragment.Resolver.Resolve | TestRunDiagnosticsContract, TestRunResolveFailuresExit1, TestGateFramingSingleDetailAtRealResolver |
| environment (1) | run -> mapping.Resolve | TestRunDiagnosticsContract, TestRunMapping |
| defaults_config_invalid (1) | axconfig.Load; CodeOf called by test | TestCodeOfProductionTypes / axconfigFailure |
| mcp (2) | composition.Compose -> Value.CheckLaunchBoundary; execution.Launch.Run | TestCodeOfProductionTypes / codexBoundary, TestLateChecksBothModes |
| system prompt (2) | systemprompt.Select / ProbeFiles; CodeOf called by test | TestCodeOfProductionTypes / syspromptSelection / syspromptProbe |
| exec (1) | execution.Launch.Run | TestLateChecksBothModes |
| ax (1) | execution.Launch.Run | TestAxFailureNoFallback |
| defaults_unresolvable, plan_refused, plan_provider_limited (3) | no producer in this tree | stated bounds, not behavioral evidence |

CodeOf has no non-test production caller (verified by caller search). Main uses owner predicates and explicit call-site codes. Full main remains TASK-260908-1o7i8y, defaults/Lineup TASK-260909-2vy977, plan/provider limits TASK-260908-2so46q. Read the existing production-main-obligations.md attached to 1o7i8y and checked owner descriptions/README: policy-before-parse, late execution boundaries, warnings, all final diagnostic call sites, and byte/status preservation remain explicit obligations. Provider-limits evidence transport has no current production path here and is not proven by this review. The interim not_implemented refusal remains a stated incomplete-main boundary. No installed working launcher, complete SPEC pipeline, new Pi/MCP support or real ax integration is claimed or waived.

## Validation and provenance

Independently executed:
- go test ./internal/diagnostics ./cmd/curator-run -count=1 -v: exit 0.
- go test ./internal/execution -run 'Test(AxFailureNoFallback|DirectExitAndSignal|LateChecksBothModes)$' -count=1 -v: exit 0.
- Nine full diagnostics/main behavioral suites with separate D13-D21 overlays, -count=1 -v: exit 1 each, required named failures. D21 companions pass. Replay script exit 0.
- git diff --check: exit 0; exact nine-path blob verification passes before and after.

Reused actual CR3 runtime TASK-260908-1wr53w_change-request_rev3-validation.log: make check exit 0 (build, fmt-check, vet, all package tests and race). Did not replay make check or run the source-writing mutant shell harness. Reviewed producer results_rev3/results_rev4 and prior diagnostics verdicts/rev2 evidence. Accepted gate artifact provenance and unchanged runner evidence are inherited; acceptance here binds to the actual adopted CR3 tree.

Evidence bundle TASK-260908-1wr53w_review-evidence-rev3.txt includes hashes, independent commands/logs, overlay sources/maps, replay script, producer diffs/named failures and exact runtime validation. All work stayed read-only for candidate source; scratch only under .temp/TASK-260908-1wr53w-review3 and board resource mutations through task-board. No commits, installs, hosted CI, real models/ax, runtime-home or private/control-root source writes. No external blocker.

Live task/reviewer checklist inspected complete. Acceptance must use accept_cr revision=3 and route to integrating, never done; the bound producer owns checkpoint/integration and signed delivery.

## Lifecycle receipt

accept_cr(TASK-260908-1wr53w, revision=3, evidence=TASK-260908-1wr53w_review-verdict-rev3.md) exited 0: CR state accepted, task integrating, reviewer RUN-260909-88927a, bound producer developer/implementer. Additional requested task-board handoff TASK-260908-1wr53w --role reviewer exited 1: role "reviewer" has no end_status and cannot use handoff. This configured-role limitation does not undo acceptance; the supported accept_cr transaction already completed the reviewer verdict route. No done transition or commit_ack supplied. Next action belongs to a new bound producer integration run.
