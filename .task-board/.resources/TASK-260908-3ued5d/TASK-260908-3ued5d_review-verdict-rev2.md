# CR2 review verdict: accepted

Task: TASK-260908-3ued5d. CR-TASK-260908-3ued5d-2 revision 2.
Base: 84747c326eee9863ddfd7e86ac65be1056718fbc.
Candidate tree: ff61be4a8bd43fa4ffb179d31aa38e41891d4313.
Verdict: accepted; route via accept_cr to integrating, not done.
repeat-of: none. Prior F1 is resolved; no new blocking findings.

## Review conclusion

Reviewed production process.go, execution/config APIs, behavioral helpers/tests, narrowing harness, README integration, original canonical F1 reproduction/verdict, producer CR1 and F1 outcomes and exact CR2 runtime validation. Every candidate tracked blob matches the worktree. The three upstream section 5 code/test/harness blob hashes match adf627607eb334e9839288cfffce63e1268ae688 exactly (provenance.log). README preserves the upstream section and tool entry. Delta against that upstream commit is 11 execution/config/documentation paths only; main and SPEC unchanged. Upstream contents are carried forward, not new implementation.

F1 repair uses a separate child process group and foreground assignment in exec.Cmd.Start. Terminal-generated INT goes directly to the child group once; launcher-only signals relay to that group. Wait4 observes stops, restores the original terminal group and stops the launcher; SIGCONT restores child foreground ownership and resumes the child. Exit cleanup calls Cmd.Wait after explicit reaping to close transport/join copying, then restores the original foreground group. No debounce or background-only workaround.

Independent TestRealPTYOwnershipBothModes passed five trials per route through actual Launch.Run: one terminal VINTR -> one INT, parent-only INT -> exactly one additional INT, foreground reads, actual VSUSP stop, restored launcher group, continuation, resumed read and successful terminal restoration. Reviewer cleanup_probe.py additionally used the unchanged candidate API driver and isolated PTYs for exit0, exit23, TERM, HUP, QUIT and invalid executable format in both modes: all 12 cases restored terminal ownership and returned exact expected codes. Direct codes were 0/23/143/129/131/1; tracked codes were 0/1/1/1/1/1. This supplemental scratch probe is evidence, not a newly committed regression test.

## Coverage: 12 of 15 AC rows driven

Reviewer independently reran the named focused tests, including the 48-case matrix and all late-check refusals, without skips. The following producer row mapping is verified against test bodies and production callers. Row 10 is now supported within its explicit terminal bounds. Tests are present in the immutable candidate snapshot, intentionally uncommitted in this managed worktree; signed inclusion is owned by subsequent producer delivery.

| Row | Requirement | Production call site | Named behavioral test / bound |
| --- | --- | --- | --- |
| 1 | Closed ax configuration including unknown/duplicate/types | axconfig.Load → parse → fragment.ParseJSON | TestLoadClosedSchema |
| 2 | Machine/operator/absentfalse and caller-supplied CLI fact | axconfig.Load; cli.Parse | TestLoadPrecedenceBeforeParse |
| 3 | Honest filesystem failures, no invalid-read fallback | axconfig.Load → exists/directory/ReadFile | TestLoadFilesystemFailures (non-root Darwin; no permission skips) |
| 4 | Real direct argv/fullEnv/stdin/WorkDir | execution.Prepare → Launch.Run → run → exec.Cmd.Start | TestRealProcessMatrix; TestEmptyEnvironmentAndLaunchPATH |
| 5 | Exact closed D3.2 document, extensions, no inherited serialization | execution.Prepare → Launch.Run, fake ax stdin | TestRealProcessMatrix (independent expected JSON object) |
| 6 | Exact name/profile/workspace argv, composition UTC timestamp | execution.Prepare → Launch.Run | TestRealProcessMatrix; TestStandardProfileAndSnapshot |
| 7 | Ax nonzero/not-startable/error bytes/no fallback | Launch.Run tracked branch → run | TestAxFailureNoFallback (exit16, missing, nonexec, bad executable format) |
| 8 | Actual subprocess and late provider/MCP checks both modes | Launch.Run → executable/Value.CheckLaunchBoundary | TestLateChecksBothModes (removed/nonexec/missing PATH, missing/dangling/directory/unreadable MCP) |
| 9 | Mandatory third-boundary callback, fail closed | Launch.Run → Options.Boundary | TestLateChecksBothModes (nil callback and returned late refusal in both modes) |
| 10 | Exit/stdout/stderr and signal behavior | Launch.Run → run; process-level driver main | TestRealProcessMatrix; TestDirectExitAndSignal; TestForwardedSignalBothModes; TestRealPTYOwnershipBothModes (5 trials per route: terminal INT, parent-only INT, foreground read, stop/resume/read, restoration) |
| 11 | Name-only composition warnings both modes | Launch.Run warning loop | TestRealProcessMatrix; TestLookupCollisionWarningsBothModes |
| 12 | Snapshot prepared transport against caller mutation | execution.Prepare → Launch.Run | TestStandardProfileAndSnapshot |
| 13 | Installed main pipeline and config-before-usage diagnostic priority | Pending cmd/curator-run main wiring | BOUND: TASK-260908-1o7i8y owns defaults, admitted plan, prompt application, Compose/Prepare/Run and exit propagation; main still not_implemented |
| 14 | Actual §5 third probe and prompt warnings in full pipeline | Pending binding to systemprompt.PrepareLaunch | BOUND: typed callback exercised, upstream §5 now carried byte-for-byte; final main binding remains TASK-260908-1o7i8y |
| 15 | Independent review and signed delivery | Parent board/PR lifecycle | BOUND: producer handoff only; canonical reviewer and parent delivery remain mandatory |


Row 15's independent review is discharged by this verdict; signed delivery remains pending. Rows 13/14 are explicitly assigned to TASK-260908-1o7i8y and remain pending, not silently counted. Caller search confirms main still does not invoke these APIs. Required Boundary remains enforced in both modes; carrying systemprompt does not connect it.

## Validation ledger and negative evidence

- Reviewer: go test ./internal/axconfig ./internal/execution -count=1 -v -timeout=60s; exit 0, focused-01.log. All 48 matrix cases, config schema/read-failure/precedence negatives, both late-check routes, exact errors/no fallback, exit/signals and real PTYs passed. No permission skips on this non-root Darwin arm64 host.
- Reviewer: go build -o .temp/TASK-260908-3ued5d-review2/driver ./internal/execution/testdata/driver; exit 0.
- Reviewer: python3 .temp/TASK-260908-3ued5d-review2/cleanup_probe.py; exit 0, cleanup-01.log (12 expected exit/restoration results).
- Reviewer: EXECUTION_MUTANT_IDS='F1 F2 E6' python3 <exact-candidate-scratch>/.scripts/execution-mutants.py <review-evidence>/mutants; harness exit 0. Mutations ran only in an archived scratch copy; candidate source remained read-only. F1 withholds foreground ownership only in tracked mode: TestRealPTYOwnershipBothModes/tracked=true fails, go test exit 1, production Launch.Run -> run -> cmd.Start. F2 omits only parent-only INT relay: TestRealPTYOwnershipBothModes/tracked=false fails (also true), exit 1, Launch.Run -> run signal relay. E6 admits direct fallback only on ax exit16: TestAxFailureNoFallback/nonzero fails, exit 1, Launch.Run tracked error branch. Gates remain present; these are narrowing, not delete-only mutations. Exact scratch source restoration is enforced by harness.
- Accepted from attached earlier evidence, not rerun: unchanged C1-C9 and E1-E5/E7 narrowing probes, including corrected last-wins C4 and tightened C1/C3/E3; prior independent C8/E2 review probes. Their named failures and production sites are recorded in TASK-260908-3ued5d_results.md and review-verdict-rev1.md. F1/F2/E6 were rerun because affected process behavior warranted fresh evidence.
- Accepted from the actual CR2 validation resource: make check, ending [exit 0], including build, fmt-check, vet, all package tests and race tests. Downloaded runtime-validation.log is exact CR2 evidence; producer prefinalization wording that validation was pending is superseded. No redundant full-suite/all-mutants replay.
- Reviewer git diff --check exit 0. All candidate tracked blob bytes checked against the immutable tree; upstream three-blob equality recorded in provenance.log.

## Bounds and merged checklist disposition

All 17 currently merged checklist clauses reviewed. Implementation, architectural fit, fake-ax-only production calls, tests, lint/build and negative evidence are supported above. Candidate remains uncommitted as required; no reviewer source edits/commits, branch movement, installations, restarts, CI, tags/releases, actual ax/model calls, real homes or global configuration changes occurred. Board outcomes are the authorized persistence surface in place of expressly forbidden LOGBOOK/control/private writes. Outcome evidence is attached before acceptance. Changes-requested routing is conditionally inapplicable to this accepted verdict.

There is no production source-text gate; preserved-token static-gate mutant clause is inapplicable. Runtime evidence is Darwin arm64; Linux runtime, complete shell bg/disown semantics, termios restoration after an application changes terminal modes, and parent-only HUP/QUIT delivery are not independently established here. HUP/QUIT child signal exit and terminal group cleanup are measured by the supplemental probe. Default foreground launch/read/stop/resume works in real PTYs. Process replacement and PTY allocation are not claimed. Pathname probe/open races remain stated bounds. Real ax destination/session/backend/secret policy and native Pi admission are not claimed; every helper environment is synthetic. Main/defaults/admitted-plan integration, actual systemprompt callback/warnings and installed exit propagation remain TASK-260908-1o7i8y. Parent routes the required bound producer for exact signed delivery; reviewer does not integrate or mark done.

Evidence bundle: TASK-260908-3ued5d_review-evidence-rev2.tar.gz contains reviewer focused, cleanup, mutant, readiness, provenance and copied exact-tree runtime logs plus supplemental probe source. No ambient environment values are persisted.
