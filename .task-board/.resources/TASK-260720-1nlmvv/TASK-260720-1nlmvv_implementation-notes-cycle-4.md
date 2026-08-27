# TASK-260720-1nlmvv — cycle 4: closing the cycle-3 P1

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce plus the composed accepted predecessor tree; nothing
staged, committed, or published). Platform: darwin/arm64, Go 1.25.5,
golangci-lint v2.4.0 (pinned).

Cycle 3 was returned with **CHANGES REQUESTED** and exactly one blocking
finding. The three cycle-2 blockers the verdict re-checked as PASS were not
touched. This document is the delta against cycle 3.

---

## The finding, restated

> Cache compensation is still not closed over failures inside `Publish` and
> `Revert`.

Two concrete holes, both real:

1. **`Publish` mutates live state and then returns an empty result with an
   error.** It quarantines an unusable predecessor
   (`internal/buildcache/publish.go:171-178` in the cycle-3 tree), then can fail
   winner validation or cache-root sync after the staged directory is already
   live (lines 182-190), or exhaust the selection loop after the predecessor was
   already moved aside (line 199). Every one of those returns
   `PublicationResult{}`. `publishWinners` returned on the error *before*
   appending the publication, so `runCommit` had no compensation record for that
   mutation at all. A failed install or upgrade could leave the predecessor
   quarantined with a replacement live, or the logical slot empty.

2. **`Revert` is not fail-closed.** It withdrew the published winner and only
   then restored the predecessor and synced. A failure in either later step left
   the slot changed with nothing attempting to put the withdrawn winner back,
   and `runCommit` joined the error without setting `BuildCacheRetained`, so the
   presentation layer could still claim the live build cache was unchanged.

---

## 1. Publication compensates its own mutations

The verdict offered two shapes — make every mutation observable to the caller,
or compensate internally before returning. Internal compensation is the one that
landed, because the alternative asks every caller to correctly own a partial
mutation record it cannot inspect, and there is exactly one place inside the
package that knows what actually moved.

`Publish` now arms its compensation before the selection loop can move anything:

```go
displaced := ""
selected := false
defer func() {
    if err == nil || (displaced == "" && !selected) { return }
    result = PublicationResult{}
    if restoreErr := store.restoreDisplaced(entryPath, base, displaced, lock); restoreErr != nil {
        err = &StateChangedError{Key: key, Err: errors.Join(err, restoreErr)}
    }
}()
```

`displaced` and `selected` are the only two live mutations the loop can make, so
every exit below — the quarantine failure, the conflict, the lost lock, the
validation failure, the sync failure, and the exhausted retry loop — unwinds
exactly what happened. A call that moved neither returns its error untouched and
pays nothing.

The contract this buys the caller is the important part, and it is now stated on
both the store method and the `CachePublisher` interface: **an error from
`Publish` means the protected cache root is exactly what `Publish` found.** The
commit phase therefore records nothing to revert for a failed publication, which
also removes the double-compensation hazard the "return a partial record" shape
would have introduced.

One production-path change came out of making the fault seam honest rather than
decorative: the selection fault stands in for the rename itself instead of
short-circuiting the loop, so a selection that never succeeds reaches the
exhaustion exit with the predecessor already quarantined. That is one of the
paths the verdict named, and it is only reachable if the fault behaves like the
failure it models.

## 2. Reversal is fail-closed, and both halves share one implementation

`Revert` and the publication compensation are now the same code
(`restoreDisplaced`), so a failure in one path cannot behave differently from
the same failure in the other.

It is fail-closed at its own seam. Withdrawal and restoration are two renames.
A fault between them used to leave the slot empty while a perfectly usable entry
sat in quarantine — strictly worse than either end state, because a launcher
that already points into that slot resolves to nothing. So a restoration that
fails now puts the withdrawn entry back before it reports:

```go
if restoreErr != nil {
    return fmt.Errorf("restore the quarantined cache entry: %w",
        errors.Join(restoreErr, returnWithdrawn(withdrawn, entryPath)))
}
```

No byte is deleted on any path. Everything moved aside is quarantined exactly
like any other unusable entry and the ordinary sweep collects it.

## 3. The truth is carried to the operator, once

A failed restoration is marked in the error rather than described in prose, so
no caller has to parse anything:

```go
type StateChangedError struct { Key buildmeta.CacheKey; Err error }
func StateChanged(err error) bool
```

`Publish` returns it only when its own compensation failed — that is the one
case where the "cache root is unchanged" guarantee does not hold. `Revert`
returns it for **every** failure, including the refusals that move nothing: a
refused reversal leaves the entry this run published exactly where it is, so a
caller reading it as an ordinary failure would go on claiming the live cache is
unchanged.

`runCommit` now maps all three retention causes to the one flag the presentation
layer actually branches on:

| cause | `BuildCacheRetained` | warning |
|---|---|---|
| an incomplete transaction still references the entry (cycle 3) | true | reason from `reversiblePublication` |
| a publication could not put the cache back | true | `a failed publication could not put the protected cache back` |
| the reversal did not complete | true | `the reversal did not complete` |

The last two are new. The repair notice no longer names a single cause it cannot
know — it says the live build cache was left changed and points at the warning —
because with three causes the old "because an incomplete transaction may still
reference it" text would have been a guess two thirds of the time.

`Result.Errors` is unchanged in kind: reversal failures still go through
`failedBuildBoundary`, so they reach the operator bounded and redacted.

---

## 4. Deterministic fault seams and what they cover

The seam is one unexported hook on `Store`, consulted at each mutation boundary.
`New` never sets it, so no store constructed outside the package can take the
hooked path.

Several of these boundaries cannot be provoked from outside the package at all —
a post-selection validation failure needs the live entry to break between the
rename and the read, and `syncDirectory` is a no-op on Windows — so without the
seam the compensation they guard is untestable on the production path. That is
the whole reason it exists.

| boundary | fault point | test |
|---|---|---|
| quarantine the predecessor | `quarantine` | `TestAFailedPublicationRestoresTheCacheItDisplaced` |
| winner selection / loop exhaustion | `select` | same |
| winner validation | `validate` | same |
| cache-root sync | `sync` | same |
| withdrawal | `withdraw` | `TestAPublicationThatCannotRestoreReportsAChangedCache`, `TestAFailedReversalIsFailClosedAndReportsAChangedCache` |
| predecessor restore | `restore` | same |
| reversal sync | `restore-sync` | same |

What the cases assert, beyond "it returned an error":

- the live entry is **byte-identical** to what it was (`entryFingerprint` digests
  every member, its mode and its bytes; an absent slot is its own value);
- the live entry count is unchanged, the quarantine count never shrinks (nothing
  was deleted), and **no private staging directory survives** in the shared root;
- the error does **not** claim a changed cache when the compensation worked, and
  does when it did not;
- a failed compensation never leaves the slot empty — the case asserts which
  entry is live, and it is always a usable one.

The table includes an unfaulted control case, so a passing failure case cannot
be a publication that never happened.

## 5. Install and upgrade

`internal/install`:

- `TestAPublicationThatChangedTheCacheIsReportedAsRetained` — both directions of
  the new publication contract: a plain failure is not retained and is not
  reverted by the caller, a `StateChangedError` is retained.
- `TestAReversalThatDidNotCompleteIsReportedAsRetained` — the run really
  rebuilt and published, the reversal failed, the published entry is still live,
  the run says the cache was left changed, and the redacted boundary error
  reaches `Result.Errors`.
- Both assert the whole `snapshotState` map is unchanged.
- The cycle-3 tests are untouched and still green, including
  `TestAnInFlightTransactionKeepsThePublishedCacheEntry` (the live-journal
  safeguard) and `TestAFailedCommitRestoresTheBuildCacheItReplaced`.

`cmd/curator`, real CLI, real store, no fault-injection seam,
`TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` for **install and
upgrade**, extended this cycle with:

- `installedFingerprint` over every class one installation commits — the project
  `.agents` tree (launcher bytes included), the agent adapter mirrors, the
  machine runtime, the hybrid store, and the consumer ledger. It refuses to
  fingerprint a missing project tree or consumer ledger, so it cannot pass
  vacuously by comparing nothing to nothing.
- `stagedEntries` — no private staging survives in the shared cache root.

Already asserted before this cycle and still: exactly one withdrawn entry,
exactly one live entry, the exact predecessor artifact bytes, a byte-identical
install marker, `status` reporting `corrupt-build-cache` again, and the ordinary
path still repairing it afterwards.

---

## Scope notes

- `internal/buildcache` and `internal/install/commit.go` remain outside this
  task's primary ownership. The additions are the minimum the verdict's rework
  instruction requires and add no new command, flag, or wire schema.
- `curator global status` remains an explicit contract-level exclusion
  (`TASK-260729-2kaopg`). Compiled install idempotence remains
  `TASK-260729-3jku56`.
- README updated: publication now compensates itself, a failed publication
  leaves the cache as it found it, and a run that leaves the live cache changed
  says so with the reason in the warning.

---

## Findings worth carrying forward

1. **A compensation is a mutation too, and needs its own fail-closed rule.**
   Withdraw-then-restore has an interior state — slot empty, usable entry in
   quarantine — that is worse than either endpoint. Any two-rename reversal
   needs an explicit answer for what happens if the second one fails.
2. **"Returned an error" and "changed nothing" are different claims, and the
   caller cannot derive one from the other.** Marking the boundary in the error
   type is what let three unrelated retention causes collapse into one honest
   presentation decision.
3. **A fault seam that short-circuits is not the failure it models.** Making the
   selection fault stand in for the rename — rather than returning early —
   is what reached the loop-exhaustion exit the verdict named.
4. Carried from cycle 3 and still true: a cache entry cannot be atomically
   swapped with a journaled install; a returned `Journal.Commit` error does not
   by itself mean the transaction rolled back; a same-key cache replacement is
   the only cache change that stays a valid hit.
