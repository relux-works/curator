# TASK-260720-3itlly — Stage build plans before installation mutation

> **Cycle 2 update.** Reviewer findings R1 and R2 are fixed; see
> `TASK-260720-3itlly_rework-cycle-2.md` for the full rework and
> `TASK-260720-3itlly_verification-cycle-2.log` for the current gate evidence.
> Trust finalization (toolchain re-fingerprint plus an all-source recheck) now
> runs at the end of phase 19, before `Options.OnStaged` and before any
> persistent write, and cache lookups are bracketed by the frozen source. The
> sections below are updated where they described the cycle-1 behavior.

## Where the work lives

Uncommitted implementation worktree, per the board precondition:

    /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree
    base HEAD 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8

Nothing was committed or staged.

## Files

New, owned by this task:

- `internal/install/builddeps.go` — narrow injected boundaries and their real
  implementations: `Toolchain` (probe/establish), `BuildSession`,
  `CacheInspector`, `Builder`, `Clock`, `GenerationReader`, plus `BuildDeps`
  and `BuildDeps.resolve`.
- `internal/install/plan.go` — immutable plan types (`BuildOutcome`,
  `PlannedBuild`, `BuildPlan`) and the read-only `planBuilds` phase.
- `internal/install/stage.go` — `StagedBuild`, `Staged`, and the private
  `stageBuilds` pass.
- `internal/install/stage_test.go` — the tests listed below.

Modified:

- `internal/install/install.go` — `Options.{Context,Build,OnStaged}`,
  `Result.{Builds,Staged}`, named return, phase renumbering, plan/stage
  insertion, clock threading into marker writes, generation-reader threading
  into the moved-tag gate.
- `internal/install/global.go` — the same insertion for the global scope.

## Phase order after the refactor

`Project` (`Global` mirrors it without hybrid/MCP/registry):

1–14. manifest, agents, gitignore, dev substitutions, hybrid scope, locale,
      closure, skill check, active-command collisions, system commands, legacy
      skill dependencies, MCP verification, migration warnings, audit gate,
      registry attestation — unchanged.
15. `BuildDeps.resolve` — narrow boundaries; forbidden private-state roots are
    the project root (global root for `Global`), `<home>/runtime`, and the
    skills root.
16. Moved-tag gate, now reading installed generations only through
    `GenerationReader`.
17. **Build planning (read-only).** Validates each build node's frozen source
    (`buildsource.Validate`), resolves the trusted toolchain — `Probe` on a dry
    run, `Establish` otherwise — derives the `buildmeta.Input` and cache key per
    command, and inspects the protected cache **inside `buildsource.Token.Use`**,
    so a reuse verdict is rechecked immediately before and after the lookup. No
    `go list`, no `go build`, no persistent write.
18. Dry-run return.
19. **Private staging and trust finalization.** Builds every miss through
    `Builder.Stage` into the operation-private session root and verifies the
    receipt key against the plan. The pass then calls `BuildPlan.Verify`, which
    re-fingerprints the trusted toolchain through the last build child and
    rechecks every frozen source in one deterministic pass. Only then does the
    caller hand the result to `Options.OnStaged`. Nothing is published.
20–23. Consumer record, materialization, cleanup/shims/adapters, runtime GC —
      unchanged, and all strictly after 19.

The plan is released through a deferred `releasePlan`, which drops the frozen
snapshots and the private session root and **reports** the release error rather
than swallowing it. The trusted session re-verifies the toolchain once more as
it closes; since cycle 2 that is a backstop behind the phase-19 gate, not the
only check.

## Ordering and determinism

`plannedCommands` walks closure nodes in provider-first order and, inside each
node, `Node.ActiveCommandNames()` (bytewise lexical), filtering to `type:
"build"` commands that the node's edges actually activate. Staging replays the
plan in that same order.

## Outcome vocabulary

`BuildOutcome` values are exactly `buildcache.Result.DryRunOutcome()`:
`cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`,
`corrupt`, `unsupported`. `corrupt` and `unsupported` fail closed during
planning, before any persistent mutation; `would-rebuild-untrusted-cache` never
exposes a reusable artifact path.

Dry-run line format (one per planned command, scope-prefixed):

    <scope>: <skill>.<command> build source=<algorithm>:<sha256> root=<build_root> \
      dir=<source_dir> target=<goos>/<goarch>[+TUNING=value] key=<cache key> \
      outcome=<outcome>[ reason="..."]

## Scope boundaries honored

- No live target mutation adapter was written; publication of staged artifacts
  and cached hits stays with TASK-260720-2284br. `Options.OnStaged` is the seam
  that phase will take over.
- No journal work: `internal/transaction` from TASK-260720-31nl14 is not present
  in this worktree, and this task adds no journal call.
- Build commands still receive no shim, because installing one is publication.
  A schema v6 build skill therefore installs its context and stages its
  artifacts, but exports no runnable command until 2284br lands.

## Tests added (`internal/install/stage_test.go`)

Fakes: `fakeToolchain`/`fakeSession`, `fakeCache`, `fakeBuilder`, `fixedClock`,
`countingGeneration`. Fixtures: `env.buildSkill` /
`env.buildSkillWithRequirement` create schema v6 skills with `build_roots` and
`type: "build"` commands.

- `TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState` — exactly one
  probe, no session, no builder call, exact source/target/key/outcome in the
  reported line, every enumerated persistent path absent, no `.lock` left in the
  home or the project.
- `TestDryRunReportsCacheHitWithoutBuilding`
- `TestCacheHitPerformsNoSourceAwareGoCommand` — only the miss reaches the
  builder.
- `TestStagingRunsProviderFirstAndCommandLexical` — order is
  `p-alpha, p-beta, c-alpha, c-beta`.
- `TestStagedOutputsStayPrivateAndAreReleased` — staged paths readable during
  `OnStaged`, outside both installation scopes, gone after the run; receipt key
  matches the planned key and validates; no live cache entry appears.
- `TestSecondBuildFailurePreservesPriorInstallationAndLiveCache` — build one
  succeeds, build two fails; staging is deleted and the project tree, runtime
  store, live build cache, and consumer record are byte-for-byte identical.
- `TestToolchainFailureBlocksEveryPersistentMutation` — cache is never inspected
  without a trusted toolchain.
- `TestCorruptCacheEntryFailsBeforeAnyPersistentMutation`
- `TestUnsupportedCacheProtectionFailsClosed`
- `TestUntrustedCacheEntryIsRebuiltAndNeverReused`
- `TestScriptOnlyInstallPerformsNoToolchainOrCacheWork` — the legacy path
  touches neither the toolchain nor the cache nor the builder.
- `TestInjectedClockAndGenerationReaderDriveInstallMarkers`
- `TestPlannedBuildAccessorsAreImmutable`
- `TestSessionReleaseFailureIsReported`
- `TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState` and
  `TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope` (cycle 2)
- `TestCacheHitOnlyPlanStillFinalizesToolchainTrust` (cycle 2)
- `TestCacheInspectionIsBracketedByTheFrozenSource` (cycle 2)
- `TestSourceMutationDuringStagingBlocksHandoffOfACacheHit` (cycle 2)
- `TestDefaultToolchainProbeRemovesItsProbeRoot` and
  `TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure` — the real
  `goToolchain` removes its operation-private root even when resolution fails.
- `TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState` and
  `TestGlobalStagingFailureLeavesGlobalScopeUnchanged` — the same guarantees for
  the global scope.

## Verification (commands run directly, real exit codes)

| Command | Exit |
| --- | --- |
| `gofmt -l .` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test -count=1 ./internal/install/` | 0 |
| `go test -count=1 ./...` | 0 (36 packages ok) |
| `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/install/...` | 0, `0 issues.` |

Repo-wide `golangci-lint run` exits 1 with 45 pre-existing issues, none in
`internal/install`: runtimestore 20, snapshot 10, buildcache 10, buildsource 3,
scopes 1, gitignore 1. Those files are untouched by this task.

Note on one combined run: an early `go test ./...` printed
`testing: can't write .../testlog.txt: file too large` for
`internal/godriver` after that package had already printed `PASS`. Re-running
`go test -count=1 ./internal/godriver/` alone exits 0, and the final full
`go test -count=1 ./...` exits 0. It is a temp-filesystem/testlog artifact of
that package's very large output, not a test failure, and `internal/godriver`
is untouched here.

## Not covered by automated tests

`goSession.Close` removing a real `godriver.Session` operation root is exercised
only through godriver's own suite plus the failure-path tests above; constructing
a real session in an install-package test would require fingerprinting a full
GOROOT on every run.
