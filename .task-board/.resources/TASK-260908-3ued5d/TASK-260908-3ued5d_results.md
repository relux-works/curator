# TASK-260908-3ued5d — execution API developer handoff

Ready for review. Candidate is uncommitted in the managed Story worktree.
The parent owns independent review and signed delivery; no commits, installs,
real ax/model calls, daemon operations, CI/release changes, main/SPEC edits,
managed-home/global configuration edits or LOGBOOK/private-record writes were made.
Task outcomes carry the findings because the task explicitly forbids LOGBOOK writes.

## Scope and provenance

Read current SPEC §4.6/4.7, composition/CLI/mapping interfaces and corrected
Decision 0013 D3.2 at curator-spec/decisions/0013-execution-ownership-and-launch-plans.md.
Checkpoint: `84747c326eee9863ddfd7e86ac65be1056718fbc`.
Local main: `adf627607eb334e9839288cfffce63e1268ae688`; HEAD..main is 1 commit.
All validation below exercised the checkpoint plus this uncommitted candidate,
not the newer main. Dependency remains v0.5.10 without overrides or additions.

Files: internal/axconfig (loader and tests), internal/execution (real process API,
behavioral tests and disposable helper/driver sources), .scripts/execution-mutants.py,
and README.md. No production fragment/composition changes remain after mutants.

The upstream §5 package is absent at this managed checkpoint. Operator directive
RUN-260908-59650d:nudge:486b05 explicitly allows retaining the required callback
until normal source convergence. Upstream bytes were inspected read-only;
PrepareLaunch calls Select, rechecks Pi flag-file readability, calls ProbeFiles,
and formats warnings. No duplicate implementation was copied.

## Behavior

- axconfig.Load reads machine first, otherwise operator, otherwise false. A false
  machine answer ignores the operator path entirely. Strict existing JSON reader
  rejects duplicates; exactly schema/enabled, correct schema, required boolean.
  Ancestor traversal prevents dangling ancestors being laundered into absence.
  Broken/unreadable/nonregular inputs return defaults_config_invalid, with no writes.
- execution.Prepare snapshots the actual composed Value and closes the D3.2
  document with schema/schema_version and exactly four extensions. Full argv
  suffix and D4 stdin are retained. Binary/fullEnv never enter the document.
  The explicit composition timestamp fixes the UTC default name; explicit name,
  absent/standard/yolo ax profile and workspace retain the specified argv shape.
- Launch.Run executes real os/exec children. Direct receives exact fullEnv,
  including empty, WorkDir, argv and attached/unattached stdin. Tracked always
  runs ax (explicit fake path in validation), forwards its output and preserves
  nonzero stderr bytes after the own ax_handoff_failed line, without fallback.
- Both routes emit name-only composition warnings, require a typed Boundary,
  call it, freshly check provider availability and call CheckLaunchBoundary
  immediately before process start. Missing callbacks and returned errors refuse.
- Waited direct children preserve exit codes; signal exits map to 128+signal.
  Supplied/default stdio file descriptors are shared; incoming TERM is tested
  through a separate process-level API driver in both routes.

## Coverage: 12 of 15 AC rows driven

This denominator decomposes this task's acceptance text and explicit retained
obligations; it is not a whole-SPEC coverage percentage. Named tests are present
in the candidate sources and will enter signed history through parent delivery.

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
| 10 | Exit/stdout/stderr and signal behavior | Launch.Run → run; process-level driver main | TestRealProcessMatrix; TestDirectExitAndSignal; TestForwardedSignalBothModes |
| 11 | Name-only composition warnings both modes | Launch.Run warning loop | TestRealProcessMatrix; TestLookupCollisionWarningsBothModes |
| 12 | Snapshot prepared transport against caller mutation | execution.Prepare → Launch.Run | TestStandardProfileAndSnapshot |
| 13 | Installed main pipeline and config-before-usage diagnostic priority | Pending cmd/curator-run main wiring | BOUND: TASK-260908-1o7i8y owns defaults, admitted plan, prompt application, Compose/Prepare/Run and exit propagation; main still not_implemented |
| 14 | Actual §5 third probe and prompt warnings in full pipeline | Pending binding to systemprompt.PrepareLaunch | BOUND: typed callback exercised, real upstream §5 absent in this checkpoint; normal convergence and finaltree validation remain |
| 15 | Independent review and signed delivery | Parent board/PR lifecycle | BOUND: producer handoff only; canonical reviewer and parent delivery remain mandatory |

Process matrix: 3 environment shapes × 2 routes × 4 stdin forms × 2 naming
choices = 48 real launches. The tests compose actual agentic.Plan values through
composition.Compose and validated Fragment/parsed Invocation/mapping. The test
owner supplies only ChildEnv literals; process creation is never mocked.
Pi values prove transport/composition only, not nativePi admission at v0.5.10.

## Validation ledger

Host: Darwin arm64, Go 1.25.5, non-root. Commands ran as standalone processes
with direct log redirection, never tee pipelines. Exits below are actual exits.

| Command | Exit | Evidence / interpretation |
| --- | ---: | --- |
| go test ./internal/axconfig ./internal/execution -count=1 (first) | 1 | focused-01.log: initial macOS path alias and inherited PWD expectation failures; raw environment diagnostic removed, retained sanitized failure summary |
| same focused command (second) | 0 | focused-02.log; normalized fixture paths, synthetic-only child environments |
| same focused command (third) | 1 | focused-03.log; compile error in newly added signal-test fixture after editing; repaired |
| gofmt -w internal/execution (signal edit attempt) | not captured separately | Syntax diagnostics before the third test run; the enclosing shell returned test exit 1. No green claim or inferred individual exit for this attempt |
| same focused command (fourth) | 0 | focused-04.log; real signal forwarding included |
| python3 .scripts/execution-mutants.py .temp/execution/mutants | 1 | mutants-01.log; 15 killed, initial C4 survivor, documented below |
| EXECUTION_MUTANT_IDS=C4 python3 .scripts/execution-mutants.py .temp/execution/mutants-c4 | 0 | mutants-02.log; corrected last-wins narrowing mutant killed, its go test exit 1 |
| EXECUTION_MUTANT_IDS='C1 C3 E3' python3 .scripts/execution-mutants.py .temp/execution/mutants-tightened | 0 | mutants-03.log; tightened single-class probes killed, each go test exit 1 |
| make check | 0 | make-check-01.log: build, fmt-check, vet, all package tests and race tests; executed once for handoff |
| git diff --check | 0 | diff-check-02.log, after README provenance clarification |

Mutants restore exact original bytes in finally and check the restoration.
The tightened mutant reruns changed only the harness, not production sources;
make check evidence therefore applies to the same production/test candidate.
No new full suite was added merely to repeat that evidence.

The first environment assertion accidentally included ambient environment in
failure diagnostics. That scratch log was replaced with a sanitized failure
record. TestMain now builds the two disposable helpers, then clears the test
process environment and installs synthetic PATH/AX_PARENT values before any
behavioral launch or assertion. Evidence attached here contains no ambient
credential values. macOS physical cwd normalization is a fixture concern;
production keeps the supplied workspace argv unchanged.

## Narrowing-mutant evidence

All 16 current harness mutants are killed by named behavioral tests. Test command
exit 1 is expected-red evidence, not a passing test. Harness exit 0 means the
expected named failures were observed. C6 is an intentional real FIFO read
admission: the named test times out at 2 seconds and go test exits 1.

| Mutant | What it narrows the gate to | Named failing test | Test exit | Survival bound |
| --- | --- | --- | ---: | --- |
| C1 | admit exactly one extra trailing locked member | TestLoadClosedSchema/unknown | 1 | none |
| C2 | admit exactly schema other | TestLoadClosedSchema/bad_schema | 1 | none |
| C3 | admit null enabled while refusing string/number | TestLoadClosedSchema/null | 1 | none |
| C4 | admit duplicate enabled only with last-wins object storage | TestLoadClosedSchema/duplicate | 1 | none |
| C5 | ignore only explicit false machine policy | TestLoadPrecedenceBeforeParse | 1 | none |
| C6 | admit FIFO only; bound test timeout prevents hanging read | TestLoadFilesystemFailures/fifo | 1 | none |
| C7 | treat dangling symlinks only as absence | TestLoadFilesystemFailures/dangling_file | 1 | none |
| C8 | turn permission-denied read only into fallback | TestLoadFilesystemFailures/unreadable_file | 1 | none |
| C9 | non-directory ancestor only becomes absence; other ancestor failures remain | TestLoadFilesystemFailures/file_ancestor | 1 | none |
| E1 | admit absent third boundary in direct mode only | TestLateChecksBothModes/tracked=false/nil_boundary | 1 | none |
| E2 | admit prompt-file refusal in tracked mode only | TestLateChecksBothModes/tracked=true/boundary_refusal | 1 | none |
| E3 | admit only absent binary named provider in tracked mode; nonexecutable still refuses | TestLateChecksBothModes/tracked=true/provider_removed | 1 | none |
| E4 | admit missing MCP only in tracked mode | TestLateChecksBothModes/tracked=true/mcp_removed | 1 | none |
| E5 | admit missing MCP only in direct mode | TestLateChecksBothModes/tracked=false/mcp_removed | 1 | none |
| E6 | fallback only for ax exit 16; other ax failures still refuse | TestAxFailureNoFallback/nonzero | 1 | none |
| E7 | inherit parent only for empty fullEnv | TestEmptyEnvironmentAndLaunchPATH/empty | 1 | none |
| C4-initial (superseded) | Disable duplicate-enabled reader error while retaining duplicate object members | none — survivor | 0 | Load's exact-two-member gate still rejects the three-member object. This run proves no last-wins duplicate admission; the current C4 also models last-wins storage and is killed. |

There are no production source-text gates. The mutant harness edits source but
executes behavioral suites; the conditional preserved-token production gate
requirement is not applicable. Kernel exec refusal also subsumes some direct
binary availability failures; tracked-mode negative tests are essential to show
the launcher refuses before fake ax sees any request.

## Remaining bounds and integration recipe

1. TASK-260908-1o7i8y must read axconfig.Load before cli.Parse, including bad argv;
   select defaults/admit one real plan, select/apply §5, Compose, then Prepare
   with the same composition-time clock. No second plan construction here.
2. After normal convergence to upstream §5, bind a closure that freshly invokes
   systemprompt.PrepareLaunch on the validated fragment and selected semantics,
   propagates its error and prints its warnings. Bind it in BOTH modes, including
   no explicit prompt flag. Do not replace it with a constant-success callback.
3. Return Launch.Run's exit code at the executable boundary. No main integration
   or actual third-probe end-to-end claim is made by the process-level test driver.
4. Real ax's session recording, provider policy/secret validation, destination
   execution and terminal backend are external and untested here by instruction.
   The API is real os/exec behavior, but every ax used in tests is the fake helper.
5. Direct execution is a waited subprocess, not replacement or a new PTY. Shell
   job-control stop/resume and interactive terminal behavior remain unverified;
   TERM forwarding is tested, INT/HUP/QUIT forwarding is implemented but untested.
6. Pathname checks have a residual change-between-probe-and-open window. The loader
   also does not claim atomic snapshot reads under concurrent filesystem mutation.
7. Linux is an intended Go/Unix target but only Darwin was executed in this run.
   The parent must validate its integrated head; this worktree is one commit
   behind local main and does not automatically follow it.

Review the uncommitted candidate and attached command logs; canonical reviewer
is next. Checklist logbook/source-text conditions are not applicable as described
above. No authorized producer delivery remains beyond the board handoff.
