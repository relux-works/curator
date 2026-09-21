# TASK-260919-2cmg0y — review verdict, revision 1: **ACCEPT**

Reviewer run RUN-260921-4d5324 (claude-opus-5), 2026-09-21 05:40–06:05Z, macOS host
load 7–8. Change Request `CR-TASK-260919-2cmg0y-1`, base `d4fe8347`, candidate tree
`a8c089b963f394858c184161625017508edac1a2`. Reviewed against 2cmg0y-brief.md (R1–R4),
2cmg0y-review-rev1-note.md, campaign-producer-rules.md. Nothing in the Story worktree was
modified; every command ran in disposable clones under `/tmp/2cmg0y-review/` (evidence
bundle `TASK-260919-2cmg0y_review-rev1-evidence.tar.gz`).

## 1. Provenance (all reproduced, not taken from results.md)

| check | result |
|---|---|
| Story worktree tree (temp-index `write-tree` over tracked+untracked) | `a8c089b9…` = CR candidate OID |
| Leaf delta `bc29bdc0 → a8c089b9` (5 paths) `git patch-id --stable` | `79e9a5727b9c…` = captured `TASK-260919-2cmg0y_rev1-candidate.patch` |
| Per-file patch-ids leaf vs captured (CHANGELOG.md, engine.go, journal.go, namespace.go, namespace_cache_test.go) | all five identical (CHANGELOG hunk applied at the same anchor after the union merge) |
| CR patch resource sha256 `da81805e…`, patch-id `773293e7…` | = `git diff d4fe8347 a8c089b9` |
| Gate run 35563543978 (`success`, all 11 jobs) head `6295d245` | tree `a8c089b9…`, parent `bc29bdc0` (Story tip on trunk d4fe8347) |
| Baseline run 35555033097 head `2e82f98c` | tree `48ad45d6…` = pre-refresh Story tip `4f213e77` (BUG-2d9gfv rev2) |
| Sweep test file `internal/install/draftatomic_test.go` | blob `56f520d2…` identical in baseline and candidate trees; baseline→candidate differs only by this leaf + trunk's gitops/pnpmsource/git-fixture test refresh (no engine change) |

## 2. R1 — per-transaction cache, invalidation seam (verified in code and by rows)

- Cache: `Engine.nsCache map[txID]map[inputPath]canonicalCacheEntry` (engine.go:36–52), value =
  resolved path or the walk's error, keyed by the exact input path; all callers hold `engine.mu`.
- Gate: `engine.validateJournal` (journal.go:402–421) passes cached resolvers only when
  `namespaceRecheckPresent` (engine.go:547) sees an in-memory guard or a durable `BoundaryProof`;
  otherwise the direct walks (`resolveTargetPath`/`resolveCanonical`, through the same test seam).
  Both production stagers attach boundaries (`install.go:879`, `global.go:462`), so every install
  transaction is guarded; `envprofile` plans carry no check → legacy → unchanged per-save walks.
- **Invalidation seam: `Engine.checkBoundary`, engine.go:519–520** — `delete(engine.nsCache, txID)`
  is the first statement, before the guard / `RecheckDurableOne` call; `checkBoundary` is called from
  `commitTarget` immediately before the backup rename (engine.go:414) and before the install rename
  (engine.go:472). The recheck itself (`staging.RecheckOne`/`RecheckDurableOne`) never reads the
  cache. Terminal release: `removeJournalDurably` journal.go:953. Rollback never rechecks by design
  (engine.go:517–518) — see bound B-1.
- `resolveTargetPath` and `cachedCanonicalTargetPath` are textually identical to
  `canonicalNamespaceTargetPath` modulo the resolver call (awk-extracted bodies diffed, 18 lines);
  `canonicalNamespacePath` itself is untouched, so same input → same resolution / same error;
  `TestCachedTargetPathMirrorsFreeFunction` (entry true/false × existing/missing) passed on
  windows-latest, macos-latest, ubuntu-latest and both race lanes.

### Rows (rerun by me, candidate clone, precompiled binary)

`go build ./...` 0 · `go vet ./internal/transaction/` 0 · `gofmt -l` clean · `golangci-lint run
./internal/transaction/...` 0 issues · `internal/transaction` full `-count=1`: **72 top-level PASS
(168 subtests), 0 FAIL, 0 SKIP, 86 s** — including `TestCommitRunsBoundaryCheckBeforeEachWrite`,
`TestCommitBoundaryRefusalAt{Backup,Install}RollsBackPublishedTargets`,
`TestCommitRollbackNeverConsultsTheBoundaryGuard`, `TestRecovery*` (6), `TestSaveJournalRejects
NamespaceAliasIntroducedBetweenSaves` (hard link + parent symlink), `TestRecoverRejectsDecoded
TargetNamespacesAliasedWhileStopped`, `TestBuildJournalRejectsOverlappingAndAliasedTargetNamespaces`,
`TestFaultAtEveryTargetBoundaryRollsBackInReverse`, and the five new rows. Hosted: the five new rows
`pass` (not skip) on all three Test lanes and both Race lanes.

### Reviewer probes (throwaway `zz_review_probe_test.go`, candidate clone only; all PASS)

| probe | what it proves |
|---|---|
| P1 `SwappedAliasRefusedByPostRecheckSave` | dir target A + file target B under another parent; B's parent swapped for a link into A's tree between two saves of one epoch (containment-only alias, identity cannot see it). Same-epoch save: `err=nil`, 0 new walks (stale key); after `checkBoundary` with a permissive guard the next save refuses `target namespaces are not independent … overlaps`; legacy control refuses at the very next save |
| P2 `CrossEngineNoSharing` | engine A prepares (15 walks/15 distinct); a second engine on the same home walks all 15 itself on load; restart-style `Commit` fails closed (`boundary proof has no recorded targets`), journal removed, cache entry released |
| P3 `RollbackEpochWalkCounts` | 2 targets, fault at `PointBeforeInstall` of target 1: prepare 15 walks, 60 after the 3 rechecks, **0 new walks during rollback**; prior bytes restored, journal gone |
| P4 `GuardedRecoverRefusesAliasWhileStopped` | guarded variant of the existing legacy row: parent aliased while stopped → restarted engine's `Recover` refuses (14 fresh walks), live bytes untouched |
| P5 `GuardedHardLinkAliasCaughtSameEpoch` | hard-link alias between two saves of one epoch is refused by the very next save with 0 new walks — the identity half (`Stat`/`SameFile`) is per pass and not cached |

### Mutants (candidate clone; committed rows = the 5 new rows; boundary set = 14 boundary/recovery/alias rows)

| id | mutant | committed rows | probes | boundary set |
|---|---|---|---|---|
| M1 | drop `delete(engine.nsCache, …)` in `checkBoundary` | KILLED (epoch row, swapped row) | KILLED (P1) | green |
| M2 | `cachedCanonicalPath` ignores txID | KILLED (epoch, swapped, cross-tx rows) | KILLED (P1) | green |
| M3 | `namespaceRecheckPresent` always true (cache legacy) | KILLED (legacy row) | KILLED (P1 legacy control) | green |
| M4 | package-global cache shared across engines (invalidation on the global) | KILLED (cross-tx row's `engine.nsCache` presence assertion) | KILLED (P2, behavioural) | green |
| M5 | invalidate only after a *passing* guard | KILLED (swapped row: post-refusal save did not re-walk) | pass (P1 uses a permissive guard, expected) | green |
| M6 | no release in `removeJournalDurably` | KILLED (epoch row terminal assertion) | KILLED (P2) | green |
| M7 | invalidate only the rechecked target's own paths | **SURVIVES** all committed rows | KILLED (P1) | green |

M7 is a narrowing survivor: the committed invalidation rows are single-target-per-recheck shapes
(the swapped row has one target; the epoch row asserts `maxPerPath == 2`, which per-target
invalidation also satisfies), so they cannot distinguish whole-transaction invalidation (what the
candidate implements, and what R1 requires) from per-target invalidation. Recorded as residual R-1.

## 3. R2 — fsync: no change, bound confirmed

The leaf patch contains no `Sync`/`fsync`/durability line (only the test's `sync.Mutex` import and
comments); `durability_unix.go`, `durability_windows.go`, `files.go`, `staging.go`, `replace_*.go`,
`rename_noreplace_*.go` are byte-identical to HEAD. `TestPrepareSyncsStagedParentBeforePrepared
Journal`, `TestPrepareStagedParentSyncFailureDoesNotPublishPreparedJournal`, `TestSubprocessCrash
RecoveryAfterStagingChunkSync` green. The producer's inventory (no same-object double sync; Windows
directory syncs are best-effort no-ops) matches the code I read.

## 4. R3 — no proof weakened

Only 5 paths changed; no test file edited or deleted; the sweep file is the same blob as in the
baseline gate; `.github/ci/skip-classes.tsv` / `platform-cases.tsv` untouched by the leaf (the
diff between baseline and candidate `skips-observed.tsv` on windows-latest contains only trunk's
3vfwch/30ycv0 git-fixture rows). The new symlink row's `t.Skipf("symlink unavailable")` is the
existing `host-capability` class and did not fire on any lane.

## 5. R4 — measurement (extracted from the gate artifacts myself)

Sweep `TestDraftFailureAtEveryTargetClassRestoresPriorState`, `test-evidence-<os>/test/go-test.json`:

| lane | baseline 35555033097 | candidate 35563543978 (CR tree) | raw | brief proxy (286 common install tests) | engine-free proxy r | conservative normalized (raw × r) |
|---|---|---|---|---|---|---|
| windows-latest | 272.63 s | **103.54 s** | **2.63×** | 2925.77 → 1776.89 s (0.607) | 0.982 (2090 tests / 67 unchanged pkgs); 1.545 (1451 tests / 61 engine-free pkgs); median per-test 0.99 | **≥ 2.58×** |
| macos-latest | 23.15 s | 12.89 s | 1.80× | 550.62 → 518.89 s (0.942) | 1.006 / 1.208 | 1.81–2.17× |
| ubuntu-latest | 19.85 s | 10.06 s | 1.97× | 173.11 → 103.15 s (0.596) | 0.833 / 0.995 | 1.64–1.96× |

Producer's pre-refresh run 35560962044 (tree `ee6e3916`, same leaf bytes): Windows 94.43 s, raw
2.89×, brief proxy 0.551 — my extraction reproduces results.md exactly. The brief's proxy (sum of
`internal/install` top-level tests) is contaminated by the change itself (every install saves
journals), so it cannot normalize this comparison; dividing by it (producer's "norm 5.24×/4.34×")
overstates. Using tests in packages that do not reach the engine, the candidate runner was equal or
slower (r ≥ 0.98), so the Windows raw 2.63× is a floor: **AC met on the Windows hosted gate on the
exact candidate tree (2.63× raw, ≥ 2.58× runner-normalized)**. Windows per-class: 21.01/23.61/21.82/
26.29/42.59/46.23/32.18 s → 8.55/9.63/10.62/11.64/12.34/13.00/15.78 s. Windows `internal/install`
package 2180 → 1290 s; `internal/install/atomicity` 1012 → 434 s; `internal/transaction` 199 → 111 s.

Local (this macOS host, load ≈7.6, precompiled binaries, one run each): base tree `bc29bdc0`
65.24 s → candidate 54.98 s (1.19×); per class 5.98/6.40/6.76/6.97/8.36/7.68/9.12 s →
4.81/5.51/5.86/6.34/7.00/6.29/6.88 s. Consistent with the producer's 1.37× (macOS symlink walks are
cheap; the Windows lane is the one that mattered).

## 6. CHANGELOG / architecture

Entry under `## Unreleased → ### Changed`, behaviour-preserving wording, names the seam. The change
stays inside `internal/transaction`, keeps the free validators for tests, injects resolvers rather
than adding global state, and releases with the guard — fits the engine's existing guard lifecycle.

## 7. Residuals and bounds (none blocking)

- **R-1 (evidence gap, M7 survivor):** add a two-target row of the P1 shape (rechecked target ≠
  swapped target; containment-only alias) so per-target invalidation is pinned as wrong. Product code
  is correct today.
- **R-2 (doc drift):** namespace.go:30–34 (`resolvedNamespacePath`: "every saveJournal call
  resolves … from scratch, and a symlink or alias that changes between two saves is seen by the
  second one") and the comment on `TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves`
  now describe legacy journals only; guarded journals share keys within a recheck epoch.
- **B-1 (epoch semantics, as ruled):** within one epoch — the whole staging phase of `Prepare`,
  between consecutive rechecks in commit, and the whole rollback/recovery (rollback never rechecks) —
  a key-only alias (parent link swapped into another target's tree or the journal root) is not seen
  by saves until the next recheck (P1: same-epoch save passes, 0 walks); the identity half is still
  per pass, so file-level symlink and hard-link aliases are refused by the very next save (P5).
  Trunk saw every alias at the next save. The per-write boundary recheck still refuses the swapped
  target's own write and rollback restores prior state.
- **B-2 (error caching, as ruled "resolution error class"):** a non-ENOENT walk error is held for
  the epoch, so every later save of that epoch — including the in-process rollback's saves — fails
  closed and restoration is deferred to recovery (fresh engine, fresh walk). Reasoned from
  engine.go:359–372/700–719, not probed; same terminal outcome as a persistent error on trunk.
- **Simplification (optional):** `resolveTargetPath`/`cachedCanonicalTargetPath` triplicate
  `canonicalNamespaceTargetPath`; a resolver parameter on the free function would remove two copies.

## 8. Verdict

ACCEPT — `accept_cr(TASK-260919-2cmg0y, revision=1, evidence=TASK-260919-2cmg0y_review-verdict-rev1.md)`.
Checklist items 12–15 checked by this run. Story STORY-260919-37szes is complete on integration.
