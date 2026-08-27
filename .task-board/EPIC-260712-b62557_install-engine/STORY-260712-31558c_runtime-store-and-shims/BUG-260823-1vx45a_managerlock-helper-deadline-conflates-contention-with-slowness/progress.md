## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Root cause identified: what still holds the lock file when cleanup runs
- [x] Lifecycle fixed so the file is released before cleanup, not the error suppressed
- [x] Both named tests pass repeatedly, including a repeated local run
- [x] macOS and Linux stay green; no test moved out of t.TempDir and none skipped
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
REVISION 2 BRIEF, orchestrator, 2026-08-27. Revision 1 was rejected. Its premise came from this bug report, and the report was wrong. Discard the revision-1 approach entirely; do not carry runWaitingHelper forward.

GROUND TRUTH, established by measurement rather than inference.

LOGBOOK.md line 279 records what actually failed: internal/managerlock TestSubprocessBuildKeyDeduplicationAcrossProjects with the message independent build key helper = blocked, want acquired. The same entry records that the identical commit failed it in pull_request run 32641695064 and passed it in workflow_dispatch run 32641704975, and that it passes on the native Windows host. There is no t.TempDir cleanup error and no still-held .lock handle anywhere in the evidence; that framing was a guess written into the report.

The base tree was run 40 consecutive times against these two tests on a native Windows host, go1.25.5 windows/amd64, and passed every time in 33.6s total. No handle-lifetime problem exists to fix.

The mechanism was then reproduced directly. acquireFileLock in internal/managerlock/filelock.go retries tryFileLock every 10ms until the caller context expires, and returns ctx.Err() when it does. TestManagerLockHelper gives that path a single 200ms budget covering os.MkdirAll, os.OpenFile and the retry loop, and in try-key mode covers two sequential acquisitions, the project lock and then the build key. Shrinking that one deadline to 1ms on the unmodified base tree, with no contention present, reproduces the exact CI string on the native Windows host across three consecutive runs. Therefore blocked in the helper protocol means the deadline expired, which has two unrelated causes: another process holds the lock, or this host was slow. A loaded runner with a filesystem scanner routinely exceeds 200ms for file creation, and an uncontended lock is then reported as contended.

WHAT TO CHANGE.

Make the deadline asymmetric, because the two expectations have opposite failure modes. Where the parent expects acquired, the helper must use a generous deadline: a free lock is taken on the first tryFileLock, so a long deadline costs nothing in the passing case and cannot be defeated by a slow host. Where the parent expects blocked, a short deadline is the correct instrument and is not flaky in the dangerous direction, because a slow host makes blocking more certain, never less; a held lock cannot be acquired at any deadline. The parent already knows which outcome it expects at every call site, so pass that expectation to the helper rather than letting one constant serve both. Note there is a second timing assumption at managerlock_test.go line 405, a 50ms deadline; judge whether it has the same defect and treat it consistently.

REGRESSION COVERAGE. The mutation above is the regression test. Add a case that runs an expected-acquired helper with a deliberately tiny deadline against an uncontended lock and asserts it reports blocked. That pins the conflation as a known, named property of the short-deadline path instead of an unexplained flake, and it fails if someone later collapses the two deadlines back into one constant.

CONSTRAINTS. Do not suppress errors, do not move state out of t.TempDir, do not mark either test as an expected failure, and do not raise the single shared constant as the whole fix, since that only widens the window it does not close it. Carry forward one correct finding from the rejected revision: in TestManagerLockHelper an acquisition error and a release error are joined and the result classified with errors.Is(err, context.DeadlineExceeded), so a failing Close is reported as blocked and exits successfully. Handle the release error independently and fail the helper on any non-nil close error before classifying the acquisition result.

TREE SELECTION. Cut from origin/main at 9bba77de355c380cf818a14980e1d9a66588e234. The primary checkout main is one commit ahead and 51 behind origin/main and carries unrelated dirty CI files; the revision-1 change request was based on that stale tree. Do not use it.

VALIDATION. macOS and Linux green. The native Windows host is reachable from the orchestrator over ssh alias win as user admin with ~/.ssh/ivanopcode, which is how the measurements above were taken; if that host is not reachable from your environment, say so plainly rather than substituting cross-compilation for execution. The orchestrator owns the commit, the pull request and the Windows CI gate.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260826-175605, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260826-175605)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260826-175605, pid=52830, exit=0)
REVIEW FOCUS for revision 2, orchestrator 2026-08-27. Revision 2 implements the asymmetric-deadline design the brief specified. Judge it on these points.

1. Whether the asymmetry is actually sound, not merely applied. Every expected-acquired call site now runs under 30s and every expected-blocked call site under 200ms. Confirm no call site was missed, including the ones in identity_windows_test.go and TestMissingHomeBelowSymlinkKeepsIdentityAndContends, and that hold-project correctly sits outside the deadline path entirely rather than inheriting a default.

2. Whether the new regression actually guards the property. TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked drives an uncontended lock with a 1ns deadline and requires blocked, and asserts acquiredHelperDeadline exceeds blockedHelperDeadline. Judge whether that pins the conflation against a future change that collapses the two constants back into one, or whether it only documents it.

3. The carried-forward F2 fix. release() errors now fail the helper through t.Fatalf before the acquisition result is classified, so a failing Close can no longer be published as blocked. Confirm no path reaches the classification with an unhandled close error.

4. Neighbouring risk, non-blocking. managerlock_test.go around line 340 asserts elapsed <= 100ms after a lock-order rejection. That is the same family of defect in the dangerous direction: a slow host can exceed it while the code is correct. It is out of scope for this bug and was not implicated in the reported failure. Say whether it deserves its own bug rather than fixing it here.

Orchestrator evidence already gathered, do not re-run it: the unmodified base passed both named tests 40 consecutive times on the native Windows host; shrinking the old single deadline to 1ms on that same base reproduced the exact CI string independent build key helper = blocked, want acquired three times out of three with no contention. Revision 2 is being run on the same host by the orchestrator; treat the Windows gate as the orchestrator PR responsibility and do not block your verdict on it.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260826-bf444b, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260826-bf444b)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260826-bf444b, pid=37888, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260823-1vx45a_spawn-log_-implementer--developer--codex-_RUN-260826-1a42c7.log](file://BUG-260823-1vx45a/BUG-260823-1vx45a_spawn-log_-implementer--developer--codex-_RUN-260826-1a42c7.log) — System spawn log captured by task-board
- [BUG-260823-1vx45a_results.md](file://BUG-260823-1vx45a/BUG-260823-1vx45a_results.md)
- [BUG-260823-1vx45a_change-request_rev1.patch](file://BUG-260823-1vx45a/BUG-260823-1vx45a_change-request_rev1.patch) — Change Request CR-BUG-260823-1vx45a-1 revision 1 candidate patch (repository_delta=present, 2 changed paths)
- [BUG-260823-1vx45a_spawn-log_-reviewer--reviewer--codex-_RUN-260826-99683f.log](file://BUG-260823-1vx45a/BUG-260823-1vx45a_spawn-log_-reviewer--reviewer--codex-_RUN-260826-99683f.log) — System spawn log captured by task-board
- [BUG-260823-1vx45a_review-verdict.md](file://BUG-260823-1vx45a/BUG-260823-1vx45a_review-verdict.md) — Reviewer verdict for Change Request revision 2: accepted with code, regression, repeated test, lint, build, and platform evidence
- [BUG-260823-1vx45a_spawn-log_-implementer--developer--codex-_RUN-260826-175605.log](file://BUG-260823-1vx45a/BUG-260823-1vx45a_spawn-log_-implementer--developer--codex-_RUN-260826-175605.log) — System spawn log captured by task-board
- [BUG-260823-1vx45a_change-request_rev2.patch](file://BUG-260823-1vx45a/BUG-260823-1vx45a_change-request_rev2.patch) — Change Request CR-BUG-260823-1vx45a-2 revision 2 candidate patch (repository_delta=present, 3 changed paths)
- [BUG-260823-1vx45a_spawn-log_-reviewer--reviewer--codex-_RUN-260826-bf444b.log](file://BUG-260823-1vx45a/BUG-260823-1vx45a_spawn-log_-reviewer--reviewer--codex-_RUN-260826-bf444b.log) — System spawn log captured by task-board
- [BUG-260823-1vx45a_review-verdict-rev2.md](file://BUG-260823-1vx45a/BUG-260823-1vx45a_review-verdict-rev2.md) — Reviewer verdict for Change Request revision 2: accepted with code, regression, repeated test, lint, build, and platform evidence

## Created
2026-08-23T01:05:35Z

## Last Update
2026-08-26T22:27:13Z

## Assigned To
[reviewer] reviewer (codex)
