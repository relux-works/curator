# BUG-260922-6chzf9 results

## Guarded contract

The helper path runs through runHelperWithDeadline in internal/managerlock/managerlock_test.go:621-630, then TestManagerLockHelper invokes Operation.AcquireProjects for try-project at lines 666-677. The helper maps context.DeadlineExceeded to the blocked report at lines 698-704.

Operation.AcquireProjects passes the request to acquireFileLock at internal/managerlock/managerlock.go:158-161. acquireFileLock checks ctx.Err before creating the lock directory or opening the lock file at internal/managerlock/filelock.go:22-34. It also rechecks after a successful OS lock and releases the lock if the deadline has expired at lines 43-49; a contended wait returns the context error at lines 52-63.

The row at internal/managerlock/managerlock_test.go:525-550 now guards the precise pre-expired case: an acquisition request whose context is already expired reports blocked and does not create or acquire the project lock path. A negative timeout is synchronously expired by context.WithTimeout. The row checks the path is absent before and after the helper, so the contract cannot pass merely because the helper reported blocked after opening or briefly acquiring the stable lock file. No production change was needed.

## Changes

- internal/managerlock/managerlock_test.go: pass a negative timeout for the already-expired case, allow that value in the subprocess helper, and assert the project lock path remains absent.
- CHANGELOG.md: add an Unreleased / Fixed entry describing the determinism change.

## Narrowing mutant

Temporarily removed only the pre-attempt ctx.Err check at internal/managerlock/filelock.go:26-28, leaving the post-acquisition deadline check intact. The focused row still received a blocked report, but detected that the lock file had been created:

Command: go test ./internal/managerlock -run '^TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked$' -count=1
Shell: zsh
Exit code: 1 (expected mutant failure; the row failed at managerlock_test.go:549 because the lock path existed).

Restored the production check. git diff --exit-code -- internal/managerlock/filelock.go returned exit code 0.

## Verification

Commands were run directly in zsh:

- go test ./internal/managerlock -run '^TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked$' -count=200 — exit 0; package reported 9.267s.
- go test ./internal/managerlock — exit 0; includes TestSubprocessContentionAndIndependentProjects and TestSubprocessBuildKeyDeduplicationAcrossProjects.
- go vet ./internal/managerlock — exit 0.
- golangci-lint run ./internal/managerlock — exit 0, 0 issues.
- go build ./internal/managerlock — exit 0.
- git diff --check — exit 0.
- gofmt -w internal/managerlock/managerlock_test.go — exit 0.

The local checks ran on the current development host. I did not claim Windows execution evidence; the brief assigns the hosted landing gate to the runtime, to run once at handoff.

## Story base preflight

The managed workspace record selected main at OID 48da2690fe79ddb24eff078c5efb13d4869aa6a9. A fresh git ls-remote --symref origin HEAD refs/heads/main advertised that same OID; git fetch --no-tags origin refs/heads/main returned it in FETCH_HEAD, and FETCH_HEAD, HEAD, and refs/remotes/origin/main matched exactly. No branch refresh or rebase was needed.

## Revision 2 (final-leaf republish)

No file was changed in this run (republish only). The `converge`ed trunk left the
accepted delta carried uncommitted in the Story worktree; identity verified:

- `git status --short` shows exactly `M CHANGELOG.md` and
  `M internal/managerlock/managerlock_test.go`; no other paths, no untracked files.
- `git diff -- internal/managerlock/managerlock_test.go` is byte-identical to the
  `internal/managerlock/managerlock_test.go` section of
  `BUG-260922-6chzf9_change-request_rev1.patch` (`diff` exit 0).
- `git diff -- CHANGELOG.md` carries the same Fixed entry text as rev1
  ("The managerlock subprocess deadline row now uses a context that is already
  expired ..."); only the base index differs (trunk combination
  `2518c29a..56714df6` vs rev1 `4c241afe..76ae04fb`). No conflict markers
  (`grep` for `<<<<<<<|=======|>>>>>>>` exit 1, no matches).

Focused bounded verification of the touched package, each run directly as a
standalone process (no pipes), real exit codes, this run on darwin:

- `go test ./internal/managerlock -run '^TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked$' -count=200` — exit 0 (`ok ... 6.023s`).
- `go test ./internal/managerlock -count=1` (all sibling rows) — exit 0 (`ok ... 2.743s`).
- `go vet ./internal/managerlock` — exit 0, no output.
- `go build ./internal/managerlock` — exit 0, no output.
- `gofmt -l internal/managerlock/` — exit 0, no files listed; `git diff --check` — exit 0, no output.
- `golangci-lint run ./internal/managerlock` — exit 0, 0 issues.

Narrowing-mutant evidence is already attached and unchanged: the rev1 results
section above records the pre-attempt `ctx.Err` removal killed by the rewritten
row (exit 1 at `managerlock_test.go:549`, lock file appeared), and the
independent rev1 review verdict (`BUG-260922-6chzf9_review-verdict-rev1.md`)
confirms its own mutant kill. The mutant was not re-run in this revision per the
no-change republish instruction; the worktree is untouched.

DoD: all 12 board checklist items were already checked before this run; none
unchecked remained. Windows execution remains deferred to the once-at-handoff
hosted gate.

## Revision 3 (refresh republish)

No file was changed in this run (republish only; trunk `1511b345` after
`2kqa77` landed, CHANGELOG-only). The converged trunk left the accepted
delta carried uncommitted in the Story worktree; identity verified:

- `git status --short` shows exactly `M CHANGELOG.md` and
  `M internal/managerlock/managerlock_test.go`; no other paths, no
  untracked files, no stash entries.
- `git diff -- internal/managerlock/managerlock_test.go` is byte-identical
  to the `internal/managerlock/managerlock_test.go` section of
  `BUG-260922-6chzf9_change-request_rev2.patch` (`diff` exit 0).
- `git diff -- CHANGELOG.md` carries the same Fixed entry text as rev2
  ("The managerlock subprocess deadline row now uses a context that is
  already expired ..."); only the base index differs (trunk combination
  `17773d62..c1d4d7cc` vs rev2 `2518c29a..56714df6`). No conflict markers
  (`grep -E '<<<<<<<|=======|>>>>>>>'` exit 1, no matches).
- `git diff --check` exit 0.

Focused bounded verification of the touched package, each run directly as a
standalone process (no pipes), real exit codes, this run on darwin:

- `go test ./internal/managerlock -run '^TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked$' -count=200` — exit 0 (`ok ... 6.854s`).
- `go test ./internal/managerlock -count=1` (all sibling rows) — exit 0 (`ok ... 2.935s`).
- `go vet ./internal/managerlock` — exit 0, no output.
- `go build ./internal/managerlock` — exit 0, no output.
- `gofmt -l internal/managerlock/` — exit 0, no files listed.
- `golangci-lint run ./internal/managerlock` — exit 0, 0 issues.

Narrowing-mutant evidence is already attached and unchanged: the results
section above records the pre-attempt `ctx.Err` removal killed by the
rewritten row (exit 1 at `managerlock_test.go:549`, lock file appeared),
and the rev2 review verdict (`BUG-260922-6chzf9_review-verdict-rev2.md`)
accepts the candidate. The mutant was not re-run in this revision per the
no-change republish instruction; the worktree is untouched.

DoD: all 12 board checklist items were already checked before this run;
none unchecked remained. Windows execution remains deferred to the
once-at-handoff hosted gate.

## Revision 4 (carry-forward republish)

No test/production content changed in this run (republish after converge onto trunk `948ae7c9`; rev3 ACCEPTED on content). The converged Story worktree carried the accepted delta uncommitted; verified per path of `BUG-260922-6chzf9_change-request_rev3.patch`:

- `internal/managerlock/managerlock_test.go` (trunk untouched): `git log --oneline 1511b345..HEAD -- internal/managerlock/managerlock_test.go` empty; worktree diff body is line-identical to the rev3 patch section (37 diff-body lines, BODY-IDENTICAL; both hunks: pre-expired `-time.Nanosecond` row plus lock-path absent assertions, and helper `deadline <= 0` guard relaxation). No conflict markers (`grep -E '<<<<<<<|=======|>>>>>>>'` exit 1). `git diff --check` exit 0.
- `CHANGELOG.md` (trunk touched by `9a59d568`, +105 lines incl. R5/R4/R3 entries): carried hunk sat cleanly under `### Fixed` at the new base (`171ec1f8..2c96d50b`, line 243 vs rev3 `17773d62..c1d4d7cc` line 138) with the same 3-line entry text, trunk entries preserved, nothing dropped/duplicated, no conflict markers. Per orchestrator CHANGELOG POLICY (2026-09-24) the hunk was then reverted entirely: `git restore --source=HEAD -- CHANGELOG.md` exit 0; `git diff --exit-code -- CHANGELOG.md` exit 0 (file equals trunk). Entry text preserved verbatim in the section below for the release-prep leaf.
- Worktree after revert: `git status --porcelain` shows exactly `M internal/managerlock/managerlock_test.go`; no other paths, no untracked files. Stray check: no root `BUG-*.md`/`TASK-*.md` (`ls` exit 1, no matches), no `test/` or `ledger/` dirs (`ls -d` exit 1).

Focused bounded verification of the touched package, each run directly as a standalone process (no pipes), real exit codes, this run on darwin:

- `go test ./internal/managerlock -run '^TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked$' -count=200` — exit 0 (`ok ... 19.886s`).
- `go test ./internal/managerlock -count=1` (all sibling rows) — exit 0 (`ok ... 3.800s`).
- `go vet ./internal/managerlock` — exit 0, no output.
- `go build ./internal/managerlock` — exit 0, no output.
- `gofmt -l internal/managerlock/` — exit 0, no files listed; `git diff --check` — exit 0, no output.
- `golangci-lint run ./internal/managerlock` — exit 0, 0 issues.

Narrowing-mutant evidence is already attached and unchanged: the results section above records the pre-attempt `ctx.Err` removal killed by the rewritten row (exit 1 at `managerlock_test.go:549`, lock file appeared). The mutant was not re-run in this revision per the no-change republish instruction; test/production content is identical to rev3.

DoD: all 12 board checklist items were already checked before this run; none unchecked remained. Windows execution remains deferred to the once-at-handoff hosted gate.

## CHANGELOG entry (for release prep)

Verbatim entry reverted from CHANGELOG.md per orchestrator policy; the release-prep leaf writes all entries:

- The managerlock subprocess deadline row now uses a context that is already
  expired and asserts that no project lock file was created, removing the
  platform-speed race from an uncontended 1 ns acquisition test.
