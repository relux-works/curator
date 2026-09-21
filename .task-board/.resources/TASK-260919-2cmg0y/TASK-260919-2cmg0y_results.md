# TASK-260919-2cmg0y results — transaction journal symlink-walk performance

Developer handoff evidence. Tree `ee6e3916beee03c47ef058d8c8f3c94108498a48`
(5 paths: CHANGELOG, engine.go, journal.go, namespace.go,
namespace_cache_test.go); hosted gate run 35560962044 `success` on that
exact tree (head `f1ed3754`, branch deleted after the run).

## Hot path

`saveJournal -> engine.validateJournal -> validateIndependentTargetNamespaces
-> canonicalNamespacePath -> filepath.EvalSymlinks`
(`internal/transaction/journal.go`, `namespace.go:201`). Every journal save
validates all target paths twice (targets-only pass inside `validateJournal`,
then targets+reserved journal-root pass in `engine.validateJournal`), each
pass walking `EvalSymlinks` per live/staged/backup/rollback path plus cleanup
tombs. Staging does many saves per file chunk with no recheck between them;
commit/rollback do 2+ saves per target. On Windows each walk is milliseconds,
so a late-class rollback costs minutes and the sweep
`TestDraftFailureAtEveryTargetClassRestoresPriorState` cost 272.63 s on the
baseline gate (F-W1 class of cost).

## Cache design + invalidation seam

- `Engine.nsCache map[transactionID]map[inputPath]canonicalCacheEntry`
  (`internal/transaction/engine.go`): keyed by the exact input path, value =
  resolved path or the resolution error from the first walk in the epoch.
- `validateIndependentTargetNamespacesWithResolver` /
  `validateJournalWithResolver` take injected resolvers;
  `engine.validateJournal` passes the cached resolvers when the transaction
  carries a recheck, else direct walks. `Engine.canonicalResolver` is the
  test seam (counting wrapper).
- Invalidation seam: `Engine.checkBoundary`
  (`internal/transaction/engine.go:512`, called from `commitTarget` before
  the backup rename `:394`-area and before the install rename `:451`-area)
  deletes the transaction's cache entry before running the recheck. The
  recheck itself (`staging.RecheckOne` / `RecheckDurableOne`) never uses the
  cache, so it always re-walks the live filesystem; the next save
  re-resolves everything. A symlink swapped between two saves in one epoch
  is therefore still caught at the next recheck.
- No cache across transactions (per-txID map) or engines (per-Engine map,
  released with the guard in `removeJournalDurably`); no cache when the
  recheck is absent (legacy `guard==nil && proof==nil` bypasses to direct
  walks through the same seam).

## Measurement

Baseline = hosted gate run 35555033097 (BUG-2d9gfv rev2, green)
`test-evidence-*` `go-test.json`; candidate = hosted gate run 35560962044
(this tree). Runner-speed proxy = sum of top-level `internal/install` tests
present in both runs (F-W1 method, 286 tests), relative to baseline. The
table also gives the proxy excluding the sweep itself (clean of the
improvement); both clear 2x on every lane.

Local (this macOS host, `go test ./internal/install/ -count=1 -run
TestDraftFailureAtEveryTargetClassRestoresPriorState`, exit 0 both):

|              | total  | 10-ctx | 20-rt | 30-shim | 50-env | 60-ledger | 80-rem | 90-cons |
|--------------|--------|--------|-------|---------|--------|-----------|--------|---------|
| before       | 72.17s | 6.94s  | 7.41s | 7.74s   | 8.10s  | 8.22s     | 7.56s  | 9.63s   |
| after        | 52.68s | 4.44s  | 4.91s | 5.07s   | 5.49s  | 6.68s     | 6.85s  | 7.01s   |

Local ratio 72.17/52.68 = 1.37x.

Hosted sweep `TestDraftFailureAtEveryTargetClassRestoresPriorState`:

| lane    | baseline | candidate | raw   | proxy base->cand | norm (brief) | norm (excl sweep) |
|---------|----------|-----------|-------|------------------|--------------|-------------------|
| windows | 272.63s  | 94.43s    | 2.89x | 2925.77->1613.07 | 5.24x        | 5.04x             |
| macos   | 23.15s   | 8.99s     | 2.58x | 550.62->463.71   | 3.06x        | 2.99x             |
| ubuntu  | 19.85s   | 12.02s    | 1.65x | 173.11->119.13   | 2.40x        | 2.36x             |

Candidate Windows per-class: 10-context 7.24s, 20-runtime 8.23s,
30-shim-canonical 10.17s, 50-env-file 11.79s, 60-adapter-ledger 12.58s,
80-removal 11.41s, 90-consumer 12.22s (baseline: 21.01/23.61/21.82/26.29/
42.59/46.23/32.18s). AC (>=2x on Windows hosted) met raw and normalized.

Note: the proxy itself falls because the cache speeds every install in the
suite, not only the sweep; the excl-sweep proxy is the cleaner runner
comparison. Either way the sweep improves far beyond any runner difference.

## fsync decision (R2): no change, bound recorded

Inventory found no admissible removal: `saveJournal` temp Sync (file) +
rename dir sync are different objects; commit backup/install rename dir sync
+ `syncTree` file sync are different objects; staging per-chunk file Sync +
dir syncs back the prefix guarantee; `syncStagedParent` is pinned by
`TestPrepareSyncsStagedParentBeforePreparedJournal` (removing it breaks a
durability row, inadmissible per R2a); `removeTreeDurably` per-entry dir
syncs repeat for different removes, and on Windows dir syncs are
best-effort no-ops (`durability_windows.go`: `FlushFileBuffers` on dirs
returns `ERROR_INVALID_HANDLE`/`ACCESS_DENIED` and is ignored;
`MOVEFILE_WRITE_THROUGH` is the primitive), so batching them cannot matter
on the lane that matters. Remaining cost after the cache is per-save
identity stats, digests, renames, and file syncs — not redundant syncs.
Durability rows unchanged and green (full transaction package below).

## Tests (all exit 0 unless noted)

- `go build ./...`, `go vet ./...`, `gofmt -l cmd internal` (clean),
  `golangci-lint run ./internal/transaction/...` (0 issues).
- New `internal/transaction/namespace_cache_test.go` (5 rows, all pass;
  symlink row skips only when `os.Symlink` fails, the sibling-accepted
  guard): epoch sharing (Prepare with 3 targets walks 22 distinct paths
  exactly once; 3 further saves walk nothing; recheck+save re-walks to 2
  per path; cache released at terminal removal), swapped-symlink refusal
  (same-epoch save walks 0 new paths; `checkBoundary` refuses
  `source_output_overlap`; post-recheck save re-walks; commit refuses and
  preserves the link with no journal left), cross-transaction isolation,
  legacy bypass, cached/free mirror.
- `go test ./internal/transaction/ -count=1` → ok 72.8s (every
  namespace/boundary negative row green, no test touched).
- `go test ./internal/staging/ -count=1` → ok; `./internal/install/atomicity/`
  → ok 266s; sweep itself unchanged and green locally (52.68s) and on all
  three hosted lanes.
- No test deleted/skipped, no new skip class (ledger vocabulary unchanged).

## Mutants (all killed, tree restored to `ee6e3916` after each)

| id | shape | killer | result |
|----|-------|--------|--------|
| M1 | drop `delete(engine.nsCache, ...)` in `checkBoundary` | epoch row (walks after recheck+save = 22, want >22) + swapped row (post-recheck save did not re-walk) | KILLED (exit 1, both rows fail) |
| M2 | `cachedCanonicalPath` ignores txID (`= "shared"`) | cross-tx row (27 walks, want >=30) | KILLED (exit 1) |
| M3 | `namespaceRecheckPresent` always true (cache legacy) | legacy-bypass row (legacy load walked nothing) | KILLED (exit 1) |

## Ratio line

No corpus change: no crossconformance rows added, removed, or reclassified
by this leaf (transaction-only change plus its unit rows).

## Re-apply on d4fe8347 (rev1b, no other content changes)

Candidate patch `TASK-260919-2cmg0y_rev1-candidate.patch`
(sha256 c37b18c4791353d0fb4e7abc92c3b5df94e20545326c0d7fb6214def85b7aa58)
re-applied onto Story tip descending from trunk `d4fe8347`
(`git merge-base --is-ancestor d4fe8347 HEAD` → ok, workspace was clean).
`git apply --index` applied all 5 paths cleanly (no `--3way` needed;
CHANGELOG entry landed once under `## Unreleased` → `### Changed`);
`git diff --cached --binary HEAD | git patch-id --stable` =
`79e9a5727b9c6ed982113c224f1223157bc95db2`, equal to the patch's own
`git patch-id --stable`. Unstaged afterwards (`git reset`); tree holds
the bytes uncommitted (4 modified + 1 new file).

Fast checks on the re-applied tree (this macOS host), all exit 0:

| command | result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./internal/transaction/` | exit 0 |
| `go test ./internal/transaction/ -count=1` | ok 122.911s, exit 0 |
| `go test ./internal/install/ -count=1 -run TestDraftFailureAtEveryTargetClassRestoresPriorState` | PASS 93.29s (pkg 93.924s), exit 0 |

Sweep per-class on re-apply: 10-context 8.73s, 20-runtime 10.62s,
30-shim-canonical 10.82s, 50-env-file 11.22s, 60-adapter-ledger 9.32s,
80-removal 9.78s, 90-consumer 12.75s. Slower than the rev1 local 52.68s
on a different host state; still green with every assertion unchanged.
Hosted-gate measurement from rev1 (Windows 272.63s → 94.43s, norm 5.24x)
stands; the configured gate reruns the suite on this tree.
