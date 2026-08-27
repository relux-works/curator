# TASK-260720-1ljev5 — Collect compiled artifacts safely

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree`
(base `origin/main` 17804ce; nothing staged, committed, or published).
Platform: darwin/arm64, Go 1.25.5. Native Windows validation on host `win`
(Windows 10/11, `DESKTOP-3PBO632\admin`).

## Provenance of the composed base

The four `done` blockers landed as uncommitted working trees, not commits. The
accepted stacked tree of the last one, `TASK-260720-2284br`
(`TASK-260720-2284br_review-verdict-cycle-5.md`, "Accepted. Route to `done`"),
lives in `.temp/TASK-260720-1zntv0/worktree` and already contains the accepted
state of `TASK-260720-3pwg2w` (protected build cache), `TASK-260720-31nl14`
(durable journal), and `TASK-260720-4bd0it` (marker v2). That tree was copied
verbatim into this task's worktree (predecessors left read-only) and verified
green before any edit: `go build ./...` exit 0.

Task-only delta, verified by directory diff against the composed base:

| File | Change |
|---|---|
| `internal/buildcache/collect.go` | new — the locked sweep |
| `internal/buildcache/protection_unix.go` | + `openProtectedDir`, `openProtectedChildFile` |
| `internal/buildcache/protection_windows.go` | + `openProtectedDir`, `openProtectedChildFile` |
| `internal/buildcache/protection_unsupported.go` | + matching stubs |
| `internal/scopes/gc.go` | shared mark phase + `Collect` |
| `internal/scopes/consumers.go` | `LoadConsumers` delegates to one parser |
| `internal/install/commit.go` | maintenance receives the held lock and journal keys |
| `cmd/curator/main.go` | `gc` acquires the home lock and recovers first |
| `README.md` | documents the grace period and the sweep rules |

Plus eight new test files (2022 lines of new code total, 315 changed lines in
the modified non-test files).

## Design

### One mark phase, two sweeps

`scopes.Collect` walks the live project, global, and hybrid scopes exactly once
and produces four things: live consumers, referenced runtime entries, referenced
logical build keys, and a list of uncertainties. `CollectRuntime` keeps its
signature and its exact prior behaviour by using the same traversal, so the
existing runtime GC regressions are unchanged.

Runtime entries are marked from **every supported marker schema**. Build keys
come only from **marker v2 entries that carry them**, plus every in-flight
journal. This is the split `TASK-260720-17llva_review-verdict-cycle-2.md`
required: a still-current schema-1 installation must not lose its runtime entry,
and it contributes no compiled-artifact reference because it has none. Covered
by `TestCollectMarksRuntimeFromEverySupportedMarkerSchema`.

### Uncertainty stops the build sweep, not the runtime sweep

A consumer registry, skill directory, or marker that **exists but cannot be
read** means the reference set is incomplete. The build cache is then not swept
at all and the uncertainty is reported. Runtime collection keeps its established
fail-safe behaviour in that case, unchanged.

`scopes.readConsumers` was added so maintenance can distinguish "no consumers"
from "unknown consumers"; `LoadConsumers` now delegates to it and keeps its
absent-or-unreadable-reads-as-empty contract for every existing caller.

### The sweep is a proof obligation, not a heuristic

`buildcache.Sweep` removes an entry only when **all** of these hold:

1. the caller holds the manager-home mutation lock (`AssertHeld`, checked before
   anything else, and re-checked by the store's existing helpers);
2. the platform can prove protected state at all;
3. the cache root revalidates as protected state through `openProtectedDir`,
   which resolves every component with `O_NOFOLLOW` / `FILE_FLAG_OPEN_REPARSE_POINT`
   and validates type, owner, and mode/DACL — the handle stays open for the whole
   sweep, so every removal resolves against the boundary that was just proven;
4. the directory name is exactly 64 lowercase hex characters;
5. no marker and no journal reference the key it encodes;
6. the entry itself revalidates as protected, structurally exact, and
   self-consistent;
7. its publication time is in the past and older than the grace period.

Anything else is retained with an actionable warning — including corrupt
receipts, untrusted roots, symlink and reparse escapes, ownership and DACL
failures, foreign root members, and manager-private `.stage-`/`.quarantine-`
directories.

**Provenance never comes from the receipt.** `inspectUnexpected` reads the
receipt only to learn which logical input the entry *claims*; that input is then
independently re-derived into a cache key which must equal the directory name,
and the existing full `inspectEntry` re-checks structure, canonical receipt,
artifact hash, and size. A receipt that is merely internally consistent proves
nothing, which is exactly what §5.6 of the accepted contract requires.

### Removal cannot escape the cache root

`retireEntry` refuses any name that is not a direct child of the validated root,
then renames the entry to a private `.sweep-<name>-<random>` sibling. That
rename *is* the removal: afterwards no reader can reach the entry by its key,
and both paths are direct children of the same proven root. The `RemoveAll`
behind it is resumable cleanup — a leftover is finished by the next sweep and
reported meanwhile. `TestRetireEntryCannotEscapeTheCacheRoot` and
`TestRetireEntryRemovesLinksWithoutFollowingThem` pin both properties.

### Grace period

The accepted contract and the manager profile both say "the manager's documented
grace period" without fixing a number, and the protocol config schema is closed
(`internal/config` rejects unknown keys, and `internal/interop` pins the wire
shape), so this is **not** a config knob. `buildcache.DefaultGrace` is a
documented Curator constant of **24 hours**, overridable per call for tests and
future callers, and documented in `README.md`.

Rationale: under the home lock, with journals marked, the only unreferenced
entry that could still be needed is one published by a process that crashed
between publication and journal preparation. Twenty-four hours is far beyond
that window, and the cost of sweeping too eagerly is one rebuild, never
correctness.

A publication time in the future is treated as clock skew: the entry is retained
and the skew is reported.

### Serialization

Maintenance never acquires a lock. It requires the caller's witness.

- `internal/install`: `collectAfterCommit` now runs inside `runCommit` with the
  same held lock, and reads `Journal.ReferencedBuildKeys(lock)` first.
  `TargetJournal` gained that method; `CommitDeps.Collect` now takes a
  `scopes.MaintenanceRequest`. Sweep warnings surface as install warnings, and a
  failure still cannot revert a durable installation.
- `cmd/curator gc`: acquires `managerlock.AcquireHomeOnly`, **recovers every
  incomplete transaction**, reads their build references, collects, and releases.
  It prints removed runtime entries, removed build entries, and warnings.

## Evidence

Reported exit codes are the real status of each command, run standalone.

### macOS (darwin/arm64), primary

| Gate | Exit |
|---|---|
| `gofmt -l .` | 0 (no output) |
| `git diff --check` | 0 |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1` (40 packages) | 0, twice — before and after the final test-file edits |
| `go test -race ./internal/scopes ./internal/buildcache ./cmd/curator` | 0 |
| `go test -race ./internal/install -run 'TestPostCommitMaintenance\|TestMaintenanceFailureAfterCommitIsAWarning\|TestConcurrent'` | 0 (62s) |
| `golangci-lint run ./...` (pinned v2.4.0) | 0, **0 issues** |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |

### Expected-red gate, honestly reported

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

This is **pre-existing and sibling-owned** (`internal/runtimestore`,
`TASK-260720-29hi1h`). The identical command on the untouched accepted base
(`.temp/TASK-260720-1zntv0/worktree`) also exits 1 with the identical message.
Excluding that one package, `GOOS=windows go vet` over this tree exits **0**.

### Native Windows

Windows has no Go toolchain, so the test binaries were cross-compiled on macOS
and executed natively. The first elevated run was uninformative: an elevated
session makes Windows assign `BUILTIN\Administrators` as owner of every object
the process creates, so `validateWindowsOwner` correctly rejects every fixture —
the same limitation recorded in
`TASK-260720-1zntv0_portable-worker-results.md`. The runs below were therefore
executed **non-elevated** through a temporary `schtasks /rl LIMITED` task, which
was deleted afterwards. `newSweepStore` skips with the exact reason on a host
that cannot host protected state, so the elevated case is self-diagnosing rather
than silently red.

| Suite | Exit | Result |
|---|---|---|
| `internal/buildcache` | 1 | **every new sweep test passes**; the three failures are pre-existing |
| `internal/scopes` | 0 | mark, prune, fail-safe, lock discipline all pass |
| `cmd/curator -run TestGC` | 0 | all three serialization tests pass |

The three `internal/buildcache` failures are `TestAtomicPublicationIdenticalRace`,
`TestAtomicPublicationConflictingRace`, and
`TestWindowsProtectedStateMatrix/artifact_has_inherit-only_owner_allow` — all
owned by `TASK-260720-3pwg2w`. The **accepted base binary, run on the same host
in the same session**, fails the same tests; five repeated runs of
`-test.run TestAtomicPublication` exit 1 on base and on this tree alike. No new
Windows failure is introduced by this task.

Two further Windows results are harness artifacts, not product defects:

- `internal/managerlock` fails five subprocess tests with
  `exec: "managerlock.test.exe": cannot run executable found relative to current directory` —
  a standalone cross-compiled test binary cannot re-exec itself. Package
  untouched by this task.
- `cmd/curator/TestCLIEndToEndInstallStatusAndTamperCheck` fails because `git`
  is not installed on the host (`where git` finds nothing). Pre-existing test,
  untouched by this task.

The three `internal/scopes` integration tests skip on Windows: building a
manager-protected home from outside `internal/buildcache` needs the unexported
`protectWindowsPath`, so the fixture cannot be created there. Those exact code
paths do run natively on Windows through the `internal/buildcache` sweep tests.

## Acceptance criteria

| Criterion | Where |
|---|---|
| Referenced build entries survive | `TestSweepRetainsReferencedAndYoungEntries`, `TestCollectSweepsOnlyUnreferencedProtectedEntries` |
| Keys named by incomplete journals survive | `TestCollectRetainsAJournalOwnedEntry`, `TestPostCommitMaintenanceMarksInFlightJournals` |
| Younger than grace survives, older removed atomically | `TestSweepUsesTheDocumentedDefaultGrace`, `TestCollectSweepsOnlyUnreferencedProtectedEntries` (asserts no partial tree remains) |
| Invalid markers never cause unsafe deletion, produce warnings | `TestCollectSkipsTheBuildSweepOnUnprovableReferences`, `TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead` |
| Corrupt receipts | `TestSweepRetainsUnprovableEntries` (5 cases) |
| Untrusted roots, ownership/DACL failures | `TestSweepRetainsUntrustedUnixState` (5), `TestSweepRetainsUntrustedWindowsState` (4) |
| Reparse/symlink escapes | `TestSweepDoesNotFollowSymlinkedEntries`, `TestSweepDoesNotFollowASymlinkedCacheRoot`, `TestSweepRemovalDoesNotFollowLinksInsideAnEntry`, and the Windows reparse pair |
| Install/rollback/recovery/gc serialize on the home lock | `TestGCWaitsForTheHomeLock`, `TestGCRunsSerializedAcrossConcurrentInvocations`, `TestPostCommitMaintenanceRunsUnderTheHeldHomeLock`, `TestSweepRequiresTheHomeLock`, `TestCollectRequiresTheHomeLock` |
| No lost consumer update | `TestGCRunsSerializedAcrossConcurrentInvocations`, `TestCollectPrunesConsumersInsideTheSamePass` |
| Deleting one entry cannot escape the cache root | `TestRetireEntryCannotEscapeTheCacheRoot`, `TestRetireEntryRemovesLinksWithoutFollowingThem` |
| Runtime GC regressions green | `TestGcKeepsReferencedRuntime`, `TestGcPrunesDeadConsumers` unchanged and passing |

## Findings worth carrying forward

1. **Pre-existing hazard, deliberately not changed.** A `consumers.json` that
   exists but does not parse makes `LoadConsumers` return nil, and
   `CollectRuntime` then rewrites the registry as empty — a corrupt registry
   wipes every consumer. That is established runtime GC behaviour and the task
   requires runtime compatibility, so it was left alone; the new build sweep
   treats it as an uncertainty and refuses to sweep. Worth its own task.
2. **Windows validation needs a non-elevated session.** Running the suite from
   an elevated SSH shell makes every protected-state test fail for a reason that
   has nothing to do with the code. `schtasks /rl LIMITED` is a clean, local,
   self-cleaning way to get real coverage, and is now the recommended recipe.
3. **Entry age comes from the entry directory's mtime.** Cache entries are
   immutable, so that is publication time. A future-dated entry is retained and
   reported rather than trusted.
