# TASK-260720-3itlly — rework cycle 3

Addresses the single high finding in `TASK-260720-3itlly_review-verdict-cycle-2.md`.

Work continues in the accepted uncommitted worktree
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
at base HEAD `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. Nothing committed or staged.

## R1 — the deferred release still reported toolchain drift after live mutation

**Cause.** Cycle 2 split trust finalization (`BuildPlan.Verify`) out of teardown and
documented `BuildSession.Close` as a pure private-root release, but the real
adapter never implemented that contract. `goSession.Close` delegated to
`godriver.Session.Close`, which re-runs `VerifyToolchain` before removing its
operation root, and `releasePlan` turned any close error into `result.failf`.
Because `releasePlan` is deferred immediately after planning, that path runs
after `OnStaged`, `RecordConsumer`, context/runtime installation, cleanup,
shims/global bins, env files, adapters, and runtime GC — so a toolchain that
drifted after the pre-handoff check produced a `failed` install once live state
had already changed.

The mutation run in `TASK-260720-3itlly_verification-cycle-3.log` reproduces the
old behaviour verbatim: `Status:failed`, `Errors:[toolchain tree changed during
operation]`, in a result whose own message list already reads
`build-skill tag v1 70e7c35 context=yes commands=[alpha] via=<project> installed`.

**Fix — separate the cleanup from the verdict at every layer.**

1. `godriver.Session` (`internal/godriver/session.go`). The private-root removal
   moved into `removePrivateState`. `Close` keeps its exact previous behaviour
   (verify, then remove, joining both failures) and is still what `probe` uses.
   New exported `Release` runs `removePrivateState` only — no fingerprinting, no
   drift verdict — and shares `closeOnce` with `Close`, so teardown happens once
   whichever entry point a caller uses.
2. `BuildSession` (`internal/install/builddeps.go`). The interface method
   `Close() error` became `Release() error`, documented as the counterpart of
   `VerifyToolchain`: teardown that verifies nothing. `goSession.Release` now
   calls the driver's cleanup-only release and removes its own private base.
3. `BuildPlan.Close` became `BuildPlan.Release` (`internal/install/plan.go`) with
   the same contract: it releases the session and the frozen source tokens and
   rechecks nothing.
4. `releasePlan` no longer touches `result.Status`. A private root that outlived
   its operation is reported as a scope-prefixed message, because by the time the
   deferred release runs the installation has already failed or committed, and
   nothing discovered there can protect live state.

The trust boundary is unchanged and remains where cycle 2 put it: `stageBuilds`
ends with `plan.Verify`, so the toolchain is re-fingerprinted through the last
build child and every frozen source is rechecked before `Options.OnStaged` and
before phase 20 `scopes.RecordConsumer`, the first persistent write. There is now
no code path that can report a toolchain trust failure after that point.

## Deviation from the reviewer's suggested test shape

Required rework item 2 asked for project and global tests "where pre-handoff
verification succeeds but the release path returns a toolchain-drift error,
asserting no consumer, install, runtime, shim/adapter, or live-cache state
changed". After this fix the release path cannot return a toolchain-drift error
at all, so that state is unreachable. The equivalent guarantee is asserted the
other way round: the fake session fails **every** verification after the first,
and both new tests require the installation to finish `ok`, to carry no drift in
`result.Errors`, and to have verified exactly once. Any re-check reintroduced
into teardown flips those tests to `failed` — which is precisely what the
mutation run shows.

## Tests

`internal/install/stage_test.go`

- `fakeSession.Close` → `Release` (`released`/`releaseErr` counters), plus
  `lateVerifyErr`, returned by every verification after the first.
- `TestReleaseTakesNoToolchainVerdictAfterLiveMutation` — project scope: the
  install completes `ok`, no drift reaches `result.Errors`, `verified == 1`,
  `released == 1`, the skill really is installed, and staging is gone.
- `TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation` — the same for the
  global scope.
- `TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall` — the only
  failure teardown can still produce (an unremovable private root) is reported in
  `result.Messages`, leaves `result.Errors` empty, and does not retract the
  committed installation. Replaces `TestSessionReleaseFailureIsReported`.
- `TestStagedOutputsStayPrivateAndAreReleased` additionally asserts
  `verified == 1` after the whole run, pinning the pre-handoff check as the only
  toolchain verdict taken.

`internal/godriver/session_test.go`

- `TestReleaseIsCleanupOnly` — after the toolchain tree is mutated,
  `VerifyToolchain` still reports `toolchain_mutated` (the drift is real and
  detectable), `Release` returns nil, the operation root is gone, and a later
  `Close` returns the memoized release result rather than resurrecting the
  verdict.

## Regression (mutation) checks

Both halves of the fix were reverted to prove the new tests fail without them,
then restored. Verbatim output is appended to
`TASK-260720-3itlly_verification-cycle-3.log`.

- M3 — `godriver.Session.Release` delegating to `Close` again:
  `TestReleaseIsCleanupOnly` fails with
  `Release error = go-v1 toolchain_mutated: toolchain tree changed during operation`.
- M1+M2 — `releasePlan` failing the result again, plus a fake session that
  re-verifies during `Release`: all three new install tests fail, each showing a
  `failed` status on a result that already reported the skill as installed.

`grep -rn MUTATION internal/` is empty in the delivered tree.

## Cross-package touch

`internal/godriver` is owned by TASK-260720-6i3cya / TASK-260720-1zntv0 (both
`done`, both uncommitted in this shared worktree). The change there is additive:
one extracted private helper, one new exported method, and one test. `Close`
behaves exactly as before — including for `godriver.probe`, which depends on
close-time verification — and `TestToolchainMutationFailsRecheck` still passes
unchanged. `go test ./internal/godriver/` and
`golangci-lint run ./internal/godriver/...` are both green.

## Verification (each command run directly, real exit codes)

| Command | Exit |
| --- | --- |
| `gofmt -l internal cmd` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test -count=1 ./internal/install/` | 0 |
| `go test -count=1 ./internal/godriver/` | 0 |
| `go test -count=1 ./...` | 0 (36 packages ok) |
| `golangci-lint@v2.1.6 run ./internal/install/...` | 0, `0 issues.` |
| `golangci-lint@v2.1.6 run ./internal/godriver/...` | 0, `0 issues.` |
| `golangci-lint@v2.1.6 run ./...` | **1** |
| `git diff --check` | 0 |

The repo-wide lint failure is expected-red and pre-existing: the same 45 issues
as cycle 2, in runtimestore (20), buildcache (10), snapshot (9), buildsource (4),
scopes (1), and gitignore (1). Zero are in `internal/install` or
`internal/godriver`, and none of those files are touched by this task.

## Scope boundaries still honored

No live target mutation adapter was added; publication of staged outputs and
cache hits remains with TASK-260720-2284br behind `Options.OnStaged`. No journal
call was added. Build commands still receive no shim.
