# CR1 review verdict: changes_requested

Task: TASK-260908-3ued5d. CR-TASK-260908-3ued5d-1 revision 1.
Base: 84747c326eee9863ddfd7e86ac65be1056718fbc.
Candidate tree: 0ee8138ca36a0cfb5037287f1a999d3fa03676a8.
Verdict branch: changes_requested. Route: to-dev.
repeat-of: none

## F1 — One terminal interrupt is delivered twice (high)

Location: internal/execution/execution.go:175-189, production Launch.Run -> run -> exec.Cmd.Start and cmd.Process.Signal.
Negative shape: capability claim that does not reproduce; the existing parent-PID-only signal test bypasses terminal group delivery.

The child shares the launcher's foreground process group. A real terminal VINTR sends SIGINT to both processes. The launcher then relays its SIGINT to the child again. A single Ctrl-C becomes two application interrupts, which can change cancel/quit behavior. This is an observed execution defect, not a hypothetical terminal limitation or a reason to block on a human decision.

Isolated PTY experiment on Darwin arm64, synthetic counting helper only:

| Route | Trials | SIGINT count for one Ctrl-C | Clean exits |
| --- | ---: | ---: | ---: |
| Helper directly (control) | 5 | 1 in 5/5 | 5/5 |
| API driver, untracked | 5 | 2 in 5/5 | 5/5 |
| API driver, tracked fake ax | 5 | 2 in 5/5 | 5/5 |

pty_probe.py uses pty.fork to create a separate controlling terminal/session. After READY, it writes exactly one byte 0x03 to the PTY master. Helper prints PID/process group and counts signal callbacks; it exits normally after 1.5 seconds. The driver is built unchanged from candidate testdata/driver. The helper's PGID equals the launcher PID in both API routes. Inputs/environments are synthetic, no installed ax/model used. The driver's attached empty stdin does not affect terminal output descriptors or foreground process-group signal delivery.

Required rework: establish coherent signal/terminal ownership so one terminal action reaches the child once while preserving the intended direct terminal and process behavior. Add a named regression test through Launch.Run with a real isolated PTY and a counting helper; assert single delivery, and retain relevant parent-only signal/exit coverage. Assess stop/resume and foreground terminal reads for the chosen mechanism; moving a child to a background group without terminal ownership is not sufficient. Do not paper over this with a stated interactive-behavior bound or merely delete the test/forwarding without checking the replacement behavior. No platform/product decision is currently needed.

## Scope review and AC coverage

All 9 changed paths inspected and byte-compared with the exact CR tree. SPEC 4.6/4.7 and local corrected Decision0013 D3.2/D6.4 inspected. This is a present repository delta containing real os/exec implementation. No code changes were made to the candidate. Narrowing mutations ran only in an archived scratch copy.

Producer reports 12 of 15 AC rows driven. Reviewer confirms named driving tests for those 12 rows, but row 10 is partial: TestDirectExitAndSignal and TestForwardedSignalBothModes cover exit and parent-only TERM, while PTY SIGINT fails. Therefore 11 of 15 rows are fully supported, 1 is partially supported with F1, and 3 are explicit retained delivery/integration bounds. Rows 1-9 and 11-12 retain the producer table's named tests and production call sites in TASK-260908-3ued5d_results.md. All test sources are in the candidate snapshot, pending signed producer delivery.

Config Load is machine-first, including false; ignored operator paths are not traversed. Strict JSON/typed schema and filesystem error tests exercise Load. Prepare/Run tests use actual composed values and subprocess helpers for full argv, fullEnv, WorkDir, four stdin forms and exact closed D3.2 JSON, four extensions, UTC naming/profile/workspace argv. No Binary/full inherited env is serialized. Both late provider/MCP checks and required callback are exercised. Ax nonzero/not-startable errors have verbatim bytes and no fallback. No additional finding established in these areas.

Caller search confirms no installed main execution/config wiring yet; that is explicitly retained by TASK-260908-1o7i8y. The real upstream section 5 callback binding remains a separate bound. HEAD..main is 1 commit; no tests here establish the future converged tree. NativePi admission, real ax policy/session/terminal backend and Linux runtime behavior are not claimed. Source-token mutant condition is inapplicable: no production source-text gate. Job-control parity was not established by this review; F1 alone requires rework, and the chosen fix must validate its affected terminal behavior.

## Validation and evidence

- Independently ran go test ./internal/axconfig ./internal/execution -count=1 -v: exit 0, including all 48 real process matrix cases and late checks.
- Built unchanged process-level driver: exit 0.
- Ran PTY probe: exit 0; this means experiment completed, NOT acceptance. Its observations reproduce F1 in 10/10 launcher trials against 5/5 single-delivery controls.
- Independently ran narrowing cases C8 (permission read failure fallback), E2 (tracked callback refusal bypass), E6 (exit16 fallback) in exact candidate archive: harness exit 0, all three named behavioral failures observed with test exit 1. No source changes in review worktree.
- Reused exact-CR runtime make check evidence, which ends [exit 0]; did not rerun the broad suite/race suite or all 16 mutants. Producer's corrected C4 and tightened C1/C3/E3 evidence supersedes the explicitly disclosed earlier variants; no stale green inference made.
- git diff --check: exit 0. All 9 review paths match the CR tree.
- Read spawned goal immediately before verdict: run is not goal-bound; no directives recorded.

Attached TASK-260908-3ued5d_review-evidence-rev1.tar.gz contains PTY script/transcripts, focused and narrowing logs, exact-tree comparison, runtime validation and readiness logs. No LOGBOOK/control-root/private-record writes, commits, installs, daemon operations, CI, releases or real model/ax calls. Board outcomes are the authorized persistence surface.
