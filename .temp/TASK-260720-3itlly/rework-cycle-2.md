# TASK-260720-3itlly — rework cycle 2

Addresses the two high findings in `TASK-260720-3itlly_review-verdict-cycle-1.md`.

Work continues in the accepted uncommitted worktree
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
at base HEAD `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. Nothing committed or staged.

## R1 — toolchain trust was finalized only after live mutation

**Cause.** `BuildSession` exposed only `Target`, `Toolchain`, and `Close`, so the
only re-fingerprint available to the install phases was the one inside
`godriver.Session.Close`. `Project`/`Global` deferred `releasePlan` right after
`planBuilds`, so that re-fingerprint ran after consumer recording, context and
runtime installation, cleanup, shims, env files, adapters, and GC. Drift was
reported, but only once live state had already changed.

**Fix.** Trust finalization is now split from private-root cleanup.

- `BuildSession` gains `VerifyToolchain(ctx) error`
  (`internal/install/builddeps.go`). The real `goSession` already satisfies it
  through the embedded `*godriver.Session`.
- `BuildPlan.Verify(ctx)` (`internal/install/plan.go`) re-fingerprints the
  trusted toolchain and rechecks every frozen source. It is a no-op for a plan
  with no build commands, and it joins both failures so a run that broke in two
  ways reports both.
- `stageBuilds` calls `plan.Verify` as its last step
  (`internal/install/stage.go`). A `Staged` result therefore exists only when
  the frozen identities still hold, and that happens before `Options.OnStaged`
  and before phase 20 (`scopes.RecordConsumer`), the first persistent write.
- `releasePlan` still reports a `Close` error. `godriver.Session.Close`
  re-verifies once more; that is now a backstop behind the pre-handoff gate
  rather than the only check. The comment says so.

## R2 — frozen source identity was not finalized before cache reuse or handoff

**Cause.** `planBuilds` validated each snapshot once, then `planOne` called
`CacheInspector.Inspect` bare. A reuse verdict could be taken for identity A and
the snapshot could change during or after the lookup; `BuildPlan.Close` closed
tokens without `Recheck`, and misses were only checked inside each
`godriver.Build`, so no recheck covered the whole plan before handoff.

**Fix.**

- `planOne` brackets the cache decision with `buildsource.Token.Use`, which
  rechecks the exact directory instance and every file byte immediately before
  and immediately after the lookup. A raced lookup fails inside the read-only
  planning phase, so no later command is planned and no miss is compiled.
- `BuildPlan.recheckSources` performs one deterministic final recheck of every
  planned source, lexically by node name, as part of `Verify`. That covers the
  case where a source changes after its own bracket passed — including a plan
  where every command was a cache hit and nothing was compiled at all.

## Phase order (unchanged shape, one new step)

Gates 1–16 → read-only build planning 17 (now with bracketed cache lookups) →
dry-run return 18 → operation-private staging **and trust finalization** 19 →
`OnStaged` handoff → first persistent mutation 20 (`RecordConsumer`).

## Tests added (`internal/install/stage_test.go`)

Fake changes: `fakeSession` gains `VerifyToolchain` with a call counter and an
injectable error; `fakeToolchain` gains `verifyErr`; `fakeCache` gains an
`observe` hook that runs inside the lookup. New helpers: `env.snapshotDir`,
`env.mutateSnapshot`, `env.seedLiveCache`, `env.installBaseline`, and
`liveState`/`captureLiveState`/`requireUnchanged`, which fingerprint the
installed project tree, runtime store, live build cache, consumer ledger, and
global scope byte-for-byte.

- `TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState` —
  drift after the build fails the install, `OnStaged` never runs, staging is
  deleted, and all five live surfaces are byte-for-byte unchanged.
- `TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope` — the same
  for the global scope.
- `TestCacheHitOnlyPlanStillFinalizesToolchainTrust` — trust is finalized even
  when nothing was compiled; every enumerated persistent path stays absent.
- `TestCacheInspectionIsBracketedByTheFrozenSource` — the source changes inside
  the lookup; planning stops at that command (one inspection, zero planned
  builds, zero builder calls), and the error names `build-skill.alpha`.
- `TestSourceMutationDuringStagingBlocksHandoffOfACacheHit` — `alpha` is reused
  from cache, `beta` is compiled, and the shared snapshot changes during `beta`.
  The final all-source recheck blocks the handoff and preserves live state.
- `TestStagedOutputsStayPrivateAndAreReleased` now asserts inside `OnStaged`
  that the session was verified exactly once and not yet closed, pinning the
  order verify → handoff → release.

## Regression (mutation) checks

Each fix was temporarily reverted to confirm the new tests fail without it, then
restored. Verbatim output is in `TASK-260720-3itlly_verification-cycle-2.log`.

- `plan.Verify` removed → the four trust tests fail with `install status = "ok"`.
- the `source.Use` bracket removed → `TestCacheInspectionIsBracketedByTheFrozenSource`
  fails: the mutation is caught late, by the final recheck, and the error is
  `build-skill: build-source snapshot mutated` instead of naming the command.

## Verification (each command run directly, real exit codes)

| Command | Exit |
| --- | --- |
| `gofmt -l internal cmd` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test -count=1 ./internal/install/` | 0 |
| `go test -count=1 ./...` | 0 (36 packages ok) |
| `golangci-lint@v2.1.6 run ./internal/install/...` | 0, `0 issues.` |
| `golangci-lint@v2.1.6 run ./...` | **1** |
| `git diff --check` | 0 |

The repo-wide lint failure is expected-red and pre-existing: 45 issues in
runtimestore (20), buildcache (10), snapshot (9), buildsource (4), scopes (1),
gitignore (1). Zero are in `internal/install`, and none of those files are
touched by this task — they are owned by the sibling build-driver tasks in the
same worktree.

## Scope boundaries still honored

No live target mutation adapter was added; publication of staged outputs and
cache hits remains with TASK-260720-2284br behind `Options.OnStaged`. No journal
call was added. Build commands still receive no shim.
