# TASK-260908-3ued5d F1 rework / CR2 evidence

Ready for independent review, not accepted completion. Candidate remains uncommitted at checkpoint 84747c326eee9863ddfd7e86ac65be1056718fbc. No branch movement, commits, main/SPEC changes, installations, daemon restarts, CI, real ax/models, home/config writes or private-record edits. Board outcomes substitute the expressly forbidden logbook.

## F1 repair

Read canonical review-verdict-rev1 and its PTY experiment: one terminal Ctrl-C produced two SIGINT in 5/5 trials per route. Launch.Run -> run now creates a separate child process group. On a controlling terminal, exec.Cmd.Start gives that group foreground ownership before exec. Parent-only INT/TERM/HUP/QUIT relay to the child group, while terminal-generated signals reach that group directly once. Wait4 observes child stops as well as exit; a stop restores the original terminal group and suspends the launcher. SIGCONT restores the child foreground group before resuming it. Exit restores the original group; normal exit and direct 128+signal remain intact. Cmd.Wait closes process transport and joins I/O copying after explicit reaping. No debounce, replacement claim, background-only child workaround, or new dependency.

TestRealPTYOwnershipBothModes drives the actual Launch.Run via the compiled testdata/driver, using Python POSIX pty.fork and synthetic executable helpers. Five trials per route assert foreground child ownership, one VINTR byte -> one INT, parent-PID-only INT -> exactly one more INT, foreground input (direct stdin / fake ax controlling tty), VSUSP -> stopped launcher, restored foreground on stop, SIGCONT -> resumed foreground input, terminal restoration after success and exit 0. The Python requirement is documented in README. Fixtures use only synthetic environments; cleanup targets the recorded helper process group and launcher PID.

## Coverage: 12 of 15 AC rows driven

This preserves the prior denominator, restores row 10 after F1, and does not claim full SPEC coverage. Test sources are in the uncommitted candidate for the managed snapshot and subsequent parent-signed delivery; the producer must not commit them here.

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

48 real composed subprocess matrix cases remain (3 environments × 2 modes × 4 stdin forms × 2 naming choices). Main/defaults/admitted plan/real third probe binding remain owned by TASK-260908-1o7i8y. Carrying upstream API code does not discharge row 14. Independent reviewer and parent signed delivery remain required.

## Upstream convergence provenance

Four-path source: adf627607eb334e9839288cfffce63e1268ae688. Applied its ordinary Git patch for internal/systemprompt/systemprompt.go, internal/systemprompt/systemprompt_test.go, .scripts/systemprompt-mutants.py; byte equality and SHA-256 values are in upstream-provenance.log. README merges its System-prompt API boundary section, tool row and integration wording with the execution documentation. Upstream code/tests/harness are unchanged and not claimed as new implementation. HEAD..main remains 1 because this managed branch cannot move; the upstream contents are present in the candidate tree.

## Validation actually run in this rework

- signal-01.log: focused build exit 1, unused import after extracting process code; fixed.
- signal-02.log: TestDirectExitAndSignal and TestForwardedSignalBothModes exit 0.
- pty-01.log: exit 1, fixture named pty.py shadowed Python standard library; renamed.
- pty-02.log: enclosing test command exit 137 during failed fixture cleanup; no passing evidence. Replaced terminal-derived group cleanup with explicitly recorded child group and launcher PID; fixed Python tty stream opening.
- pty-03.log: real PTY suite exit 0, five trials per route.
- focused-01.log: go test ./internal/axconfig ./internal/execution -count=1 -v, exit 0.
- mutants-01.log: EXECUTION_MUTANT_IDS='F1 F2 E6' python3 .scripts/execution-mutants.py .temp/F1/mutants, harness exit 0; each mutant test exit 1 at the named assertion, no survivors. Exact source bytes restored after each mutation.
- focused-02.log: final go test ./internal/axconfig ./internal/execution -count=1 -v, exit 0, including parent-only INT and strengthened safe PTY cleanup.
- linux-build-02.log: GOOS=linux GOARCH=amd64 go build ./internal/execution, exit 0. Linux runtime is not claimed.
- git diff --check: exit 0. gofmt -l internal/execution: no offenders.
- task-board handoff TASK-260908-3ued5d --role developer returned exit 0, status to-review, checklist 17/17. No CR2 validation resource exists yet: runtime candidate publication/validation occurs outside this provider turn. The single runtime make check is therefore pending, not claimed green; the parent must inspect its real CR2 validation result before routing acceptance. No manual broad-suite duplicate was run.

## Negative evidence

| Mutant | Narrowing | Named failing test and production call site | Test exit |
| --- | --- | --- | ---: |
| F1 | Withhold foreground ownership only for tracked ax; keep direct ownership and child isolation | TestRealPTYOwnershipBothModes/tracked=true; Launch.Run -> run -> cmd.Start | 1 |
| F2 | Omit only parent-only INT forwarding; keep TERM/HUP/QUIT and terminal delivery | TestRealPTYOwnershipBothModes/tracked=false (also tracked=true); Launch.Run -> run signal relay | 1 |
| E6 | Fall back only on ax exit 16, keep other refusal paths | TestAxFailureNoFallback/nonzero; Launch.Run tracked error branch | 1 |

Accepted from already-attached evidence, not rerun: prior 16 config/execution narrowing mutants (with prior tightened C1/C3/E3 and corrected C4), reviewer independently repeated C8/E2/E6. E6 was rerun here because process exit representation changed. Unchanged gates and 12 behavior rows were rerun by the focused suite; prior broad make check is historical evidence only, not evidence for this new tree. No production source-text gate exists, so preserved-token static-gate clause is inapplicable.

## Bounds

Only Darwin arm64 runtime exercised. Linux compile succeeds but runtime remains unverified. The launcher waits for a child and does not replace its process or allocate a PTY. fg/SIGCONT stop/resume is exercised; complete interactive-shell bg/disown semantics are not claimed. Parent-only TERM and INT are measured; HUP/QUIT relay is implemented but not independently driven. Probe/open pathname races remain. Actual ax session/provider/secret-policy/backend behavior is expressly excluded; every ax here is fake. Real systemprompt binding, installed main behavior and signed delivery await the designated integration/parent lifecycle.
