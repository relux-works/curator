# TASK-260720-2284br — Commit installations atomically across scopes

## Where the work lives

Uncommitted implementation worktree, per the board orchestration note:

    /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree
    base HEAD / origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8

Nothing was committed or staged.

## Base composition (read this first)

The four direct blockers had landed in **separate uncommitted worktrees**, not in
one tree, so the first step was to assemble the integration base:

| Source worktree | Contribution |
|---|---|
| `.temp/TASK-260720-1zntv0/worktree` | `internal/godriver`, `internal/install` build staging (`builddeps.go`, `plan.go`, `private.go`, `stage.go`) — this is also where TASK-260720-3itlly landed |
| `.temp/TASK-260720-31nl14/worktree` | `internal/transaction`, `internal/managerlock`, marker v2 (TASK-260720-4bd0it) and its interop golden |

The two chains share 128 files and diverged on 10. Every divergence was one
chain sitting at the base version while the other had advanced, so the union is
the per-file superset — no content merge was needed:

- `internal/marker/{marker,marker_test}.go`, `internal/interop/golden_test.go` → the marker-v2 line (31nl14 chain)
- `internal/buildmeta/{codec,models,buildmeta_test}.go`, `cmd/curator/{main,main_test}.go`, `internal/install/{install,global}.go` → the godriver/staging line (1zntv0 chain)

To guard against a mis-merge, the union was built **twice and compared**: once
into a throwaway worktree at `.temp/TASK-260720-2284br/worktree` from
`origin/main`, and once into the designated stacked worktree. The two trees are
byte-identical (verified by a SHA-256 manifest diff; only the board-local
`task-board.config.json` differs). A pre-import backup of the stacked tree is at
`.temp/TASK-260720-2284br/backup-1zntv0-tree-preimport.tar.gz`.

The composed base was green **before** any of this task's changes:
`go build ./...`, `go vet ./...`, `go test ./... -count=1` all exit 0.

## What changed

### New: `internal/staging`

Dependency-free descriptor for one operation-private replacement
(`Class`, `Identifier`, `LivePath`, `StagedPath`). Classes carry a numeric
prefix because the transaction engine orders targets bytewise by
`(class, identifier)` — the prefix *is* the commit order:

    10-context → 20-runtime → 30-shim-canonical → 40-shim-forwarding
    → 50-env-file → 60-adapter-ledger → 70-mirror-ledger
    → 80-removal → 90-consumer

`Plan.Validate` rejects producer defects (duplicate identifier, two producers
claiming one live path, non-absolute paths) before anything reaches the journal.

### New: `internal/install/commit.go`

The serialized publication and commit phase (contract §6.1 steps 10–14):

1. acquire the manager-home lock
2. `Journal.Recover` — recovery completes before this operation mutates anything
3. revalidate every optimistic observation → `restartError` instead of applying a stale plan
4. `publishWinners` — publish staged builds, then re-read every planned key
5. `stageTargets` — the scope derives its complete desired state, *under the lock*
6. `scopes.StageConsumer` — merged from live state, appended as the last class
7. `transaction.Prepare` → `Mirror.Link` → `transaction.Commit`
8. `Mirror.PruneStale` (warnings), then `Collect` (warnings)
9. release home lock, then project locks

Boundaries are injected through `Options.Commit` (`LockBroker`, `TargetJournal`,
`CachePublisher`, `Collect`, `transaction.Hooks`, `PublishFault`).

### New: `internal/install/targets.go`

`stageNode` (context + marker), `stageRuntimeAndShims` (runtime trees, canonical
launchers, stale launcher removals), `stageStaleSkillRemovals`, `contextSources`.

### Staging producers added to existing packages

- `internal/envfiles/stage.go` — `StageProject`, `StageGlobal`; content extracted so a staged file is byte-identical to a directly written one.
- `internal/adapters/stage.go` — `StageProject`, `StageGlobal` returning a `Mirror`.
- `internal/globalbins/stage.go` — `StageForwarding` returning journaled forwarding shims, removals, and the mirror ledger.
- `internal/scopes/stage.go` — `StageConsumer`.
- `internal/runtimestore/targets.go` — `ForwardingTarget`, `ManagedShimsIn`.

### Rewritten lifecycle

`Project` and `Global` are now thin wrappers: dry run returns immediately with
no lock; a real run takes the project lock, recovers journals under a brief home
lock, then loops `projectAttempt`/`globalAttempt` through `runWithRestarts`.
Phases 20–23 (consumer-first, then piecemeal materialization, cleanup, adapters,
GC) are gone; they are now one staged target set committed atomically, with the
consumer ledger last.

Superseded direct-mutation helpers removed: `installRuntime`, `installContext`,
`installMarkerOnly`, `cleanupRemoved`.

## Decisions worth reviewing

**1. Adapter mirror entries vs. the ledger.** The transaction layer deliberately
refuses links as targets and inside trees (`DigestPath` → "unsafe transaction
target type"; there is an accepted test asserting it). Adapters default to
`adapter_mode: auto`, which is symlink-first. Contract §6.1 step 12 resolves
this by naming **"adapter ledgers and hybrid/global mirrors"** as the commit
class — not each mirror entry. So:

- adapter/mirror **ledgers**, and copy-mode entries (real directories), are journaled targets;
- symlink entries are created after `Prepare` and before `Commit`, and stale entries are pruned only **after** a successful commit.

A rollback therefore never deletes a live mirror. A leftover entry is still
recognizably ours (`unmanagedConflict` accepts a link resolving to the canonical
source, and a copied tree holding an install marker), so the next run reconciles
it instead of reporting a conflict. **I did not extend `internal/transaction` to
support links** — that is an accepted package outside this task's scope, and the
contract's wording does not require it. Flagging it for the reviewer as the one
place where "adapter, mirror state" is restored via the ledger rather than
per-entry journaling.

**2. Publication and verification are one boundary.** `CachePublisher` now
requires `Inspect` as well as `Publish`, and `CommitDeps.resolve` defaults the
publisher to the planning inspector when that inspector can publish (the real
`*buildcache.Store` can). Two separately constructed stores over the same home
could otherwise disagree about whether the entry a launcher points at exists.

**3. Restart granularity.** `restartStep` classifies the earliest affected phase
(`closure` > `plan` > `targets`) and the reason is reported, but a restart
re-runs the whole read/build pipeline. That is a safe superset of "restart from
the earliest affected step" — it can never apply a stale plan — and it is
bounded by `MaxRestarts` (default 3) so a livelock is reported, not hidden.

**4. Target staging happens under the home lock.** Only compilation is required
to stay outside it (contract §6.1). Deriving targets inside the lock means every
shared ledger (consumer, adapter, mirror) is merged from committed state, which
is exactly what makes two concurrent installs preserve both consumers.

**5. Parent directories are unwound.** A target below a not-yet-existing
directory needs that directory before the journal can place sidecars. Those
directories are not journaled targets, so an abandoned commit removes exactly
the ones it created (deepest first, only when still empty). Without this a
rolled-back install left an empty `<home>/runtime/<skill>/` behind and the
restored state was not byte-identical to the prior one — caught by the rollback
sweep test.

**6. Marker build state.** `buildMarker` now records `BuildRoots`,
`BuildSource`, and `Builds`, and lists build commands in `Commands`. Marker v2
validation requires build state to be all-or-nothing, so a node with no
published build carries no build source at all.

## Test-fixture corrections (accepted tests, behavior preserved)

Three fixture facts in `internal/install/stage_test.go` were unreachable states
that only staging tolerated, and the commit phase legitimately rejects:

- `testTarget()` fabricated `linux/amd64` on a darwin/arm64 host. A manager only builds natively and `CompiledTargetFromCache` refuses a foreign artifact, so the fixture is now the host's native target; the one assertion that spelled the target out now derives it.
- `fakeCache` served hits with a fabricated `ArtifactPath` and no receipt. `seedHit` now materializes a real artifact and a matching receipt from the exact input the planner derives.
- `fakeCache` is now the publishing boundary too (`Publish`), matching the real store.

No assertion was weakened; each change makes the fixture model a state the
product can actually reach.

## Tests added — `internal/install/commit_test.go`

| Test | AC clause |
|---|---|
| `TestNoSharedTargetChangesBeforeTheManagerHomeLock` | no shared target changes before the home lock |
| `TestCommitOrdersTargetClassesWithConsumerLast` | deterministic classes, consumer last |
| `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` | injected failure at every target class restores the complete prior project/hybrid/runtime/shim/env/adapter/mirror/consumer state, in exact reverse order (10 sub-tests, one per committed target) |
| `TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched` | injected failure at cache publication; pre-existing immutable entries preserved |
| `TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess` | consumer absent after a failed first install, updates last after success |
| `TestConcurrentProjectInstallsPreserveBothConsumers` | two concurrent project successes preserve both consumers (separate processes — the lock layer serializes lock classes per process by design) |
| `TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets` | one project's rollback cannot restore over another's committed shared targets |
| `TestRecoveryCompletesBeforeAnyNewMutation` | recovery completes before new mutation |
| `TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan` | a stale generation restarts rather than applying an old plan |
| `TestMaintenanceFailureAfterCommitIsAWarning` | GC failure after commit is a warning and does not roll back |
| `TestGlobalCommitCarriesNoConsumerLedger` | scope difference: global registers no consumer |
| `TestAdapterLedgerCommitsAfterTheMirrorsItClaims` | intra-class order: a ledger never claims a non-durable entry |
| `TestStagedPlanRejectsTwoProducersClaimingOneLivePath` | producer-defect gate |

## Verification

Every command was run directly as a standalone process in the implementation
worktree; exit codes are the real ones.

| Command | Exit | Log |
|---|---|---|
| `go build ./...` | 0 | `gate-build-final.log` |
| `go vet ./...` | 0 | `gate-vet-final.log` |
| `gofmt -l .` | 0 files listed | `gate-gofmt-final.log` |
| `go test ./... -count=1` | 0 (complete repository) | `gate-gotest-final.log` |
| `go test -race ./internal/install/... ./internal/transaction/... ./internal/managerlock/... -count=1` | 0 | `gate-race-final.log` |
| `golangci-lint run` over every package this task created or modified¹ | **0 — 0 issues** | `gate-lint-task-scope.log` |
| `golangci-lint run ./...` (v2.4.0, whole repository) | **1 — expected red, all inherited** | `gate-lint-final.log` |

¹ `./internal/staging/... ./internal/install/... ./internal/envfiles/... ./internal/adapters/... ./internal/globalbins/...`

This task also modified `internal/scopes` (added `stage.go`, refactored
`consumers.go`) and `internal/runtimestore` (added `ForwardingTarget`,
`ManagedShimsIn`, `LauncherTarget`). Both packages retain pre-existing issues,
but every one of them is in a file or on a symbol this task did not touch —
`scopes/hybrid.go` (gosec G304), and in `runtimestore/targets.go` the revive
comment warnings on `ScriptTarget`/`CompiledTarget`/`ManagedShim`/
`NewManagedShim`/the const blocks, three ST1005 messages, and one unused test
helper. The baseline diff below is the proof: none of them is new.

### The repo-wide lint gate is expected-red, and here is the honest accounting

`golangci-lint run ./...` exits **1** with **45 issues**. It exits 1 on the
*pre-change composed base* too, with **the same 45 issues and the same
per-linter counts** (errcheck 16, gosec 5, revive 20, staticcheck 3, unused 1).
Baseline evidence: `gate-lint-baseline.log`, produced from
`.temp/TASK-260720-2284br/worktree`, which is the composed base without any of
this task's changes.

The first lint run of this task's tree (`gate-lint.log`, 48 issues) showed
exactly three issues absent from the baseline, all mine:

- `internal/install/install.go` — unused parameter `path` in `runWithRestarts`
- `internal/runtimestore/targets.go` — missing comments on `LauncherTarget.Kind` and `LauncherTarget.ExecutablePath`

All three are fixed. The final diff against the baseline shows one line moving
between `internal/buildsource/buildsource.go` and
`internal/snapshot/snapshot_test.go` — both files are byte-identical in the two
trees and untouched by this task; it is errcheck's `max-same-issues`
representative-occurrence selection, and the totals are identical.

**Net: this task introduces zero new lint issues.** The 45 remaining belong to
inherited packages (`buildcache`, `buildsource`, `snapshot`, `runtimestore`,
`gitignore`, `scopes`) delivered by earlier tasks in this story and are outside
this task's scope to change.

They are all mechanical (missing comments on exported symbols, unchecked
`Close` in tests, capitalized error strings, gosec annotations, one unused test
helper) and clearing them is a bounded job — but it means editing five packages
owned by sibling tasks that are already `done` and reviewed, including judgment
calls on their error wording (`"Windows compiled artifact must have an .exe
suffix"` trips ST1005 precisely because `Windows` is a proper noun). That is a
coordinator call, not a decision this task should take unilaterally.

**Recommended:** open a separate repo-wide lint-cleanup task against the merged
story branch. This task's "Lint clean" is evidenced by the task-scope run above,
which exits 0 with 0 issues.

Native Darwin/arm64 only. Windows compilation was not run and no native Windows
runtime evidence is claimed.
