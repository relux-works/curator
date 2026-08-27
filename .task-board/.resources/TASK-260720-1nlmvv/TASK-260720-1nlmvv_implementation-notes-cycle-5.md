# TASK-260720-1nlmvv — cycle 5: closing the cycle-4 P1

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce plus the composed predecessor tree; nothing staged,
committed, or published). Platform: darwin/arm64, Go 1.25.5, golangci-lint
v2.4.0 (pinned, resolved from `$(go env GOPATH)/bin`).

Cycle 4 was returned **CHANGES REQUESTED** with exactly one blocking finding.
This document is the delta against cycle 4. No other cycle-4 behaviour was
touched, and no command, flag, exit code, or wire schema changed.

---

## The finding, restated

`quarantinePath` renames the live entry aside and then syncs its parent. If that
sync fails it returns `("", err)` — an error, and no record of the rename it had
already performed (`internal/buildcache/publish.go:461-467` in the cycle-4 tree).

Two callers read that as "nothing moved":

1. **`Publish`** assigned `displaced` only from a successful return
   (`publish.go:266-272`), and its compensation returns immediately when
   `displaced == "" && !selected` (`publish.go:188-198`). A post-rename
   quarantine error therefore returned an empty `PublicationResult` and an
   ordinary error with the previous live slot **missing**.
2. **`restoreDisplaced`** called the same helper to withdraw the published
   winner and returned immediately on its error (`publish.go:378-385`). A
   post-rename sync failure left the launcher-visible slot **missing**, and
   `StateChangedError` said state changed without carrying anything `runCommit`
   could recover the winner from.

The cycle-4 `faultQuarantine` and `faultWithdraw` seams fire *before* the
helper, so they never reached these interior exits.

---

## The fix, in two independent layers

Both layers are load-bearing and each was proven so by mutation (below). They
are deliberately not one mechanism: the first keeps the ordinary contract true,
the second keeps it true without trusting the step it wraps.

### 1. The quarantine move answers for itself

`quarantinePath` now treats the rename as the mutation it is. A sync that fails
puts the entry straight back:

```go
syncErr := store.fault(faultQuarantineSync)
if syncErr == nil {
    syncErr = syncDirectory(parent)
}
if syncErr != nil {
    syncErr = fmt.Errorf("sync quarantined cache root: %w", syncErr)
    if back := renameDirectoryNoReplace(quarantinePath, entryPath); back != nil {
        return quarantinePath, errors.Join(syncErr,
            fmt.Errorf("return the quarantined cache entry to the live slot: %w", back))
    }
    return "", syncErr
}
```

This preserves the guarantee every caller already reads — **an error with no
reported path means the slot is untouched** — and it reports the path in exactly
one case: the move could neither be made durable nor undone. There the path is
the only thing that keeps the entry recoverable, which is the verdict's option 2
narrowed to the one exit that needs it. `Quarantine`'s doc comment now states
this, because it is a public return-value contract.

`faultQuarantineSync` is the deterministic fault point the verdict asked for:
it sits between the rename and the sync, which is the only window where a
quarantine has mutated the cache and has not yet returned.

### 2. A withdrawal reports what it did, by observation

`withdrawEntry` wraps the fault seam and the helper together and guarantees:
*if it returns an error, either the entry is still live and the path is empty,
or the path names exactly where the entry went.*

It establishes that by looking at the cache root rather than by trusting the
step it follows — the manager home lock is exclusive, so a quarantine name that
appears across the call is this call's own doing:

```go
func withdrawnTo(entryPath, parent string, before map[string]bool) string {
    if _, err := os.Lstat(entryPath); err == nil || !errors.Is(err, os.ErrNotExist) {
        return ""
    }
    for name := range quarantineNames(parent, filepath.Base(entryPath)) {
        if !before[name] { return filepath.Join(parent, name) }
    }
    return ""
}
```

This is what makes the compensation robust to *any* step added between the
rename and the return, present or future — including the two regression fixtures
the tester wrote, which move the entry through the production helper and then
report failure at the boundary without naming it.

Both callers now own the result unconditionally:

| caller | on a withdrawal error |
|---|---|
| `Publish` | records `displaced` **before** acting on the error, so the existing compensation unwinds it |
| `restoreDisplaced` | calls `returnWithdrawn` immediately, exactly as it already did for a failed restoration |

The second one closes the asymmetry cycle 4 left: a failed *restoration* put the
withdrawn entry back, a failed *withdrawal* did not, though both leave the same
empty slot a live launcher resolves to nothing through.

---

## Tests

### The two cycle-4 regressions, now green on the production path

- `TestAFailedPublicationRestoresAPredecessorMovedBeforeQuarantineError`
- `TestAFailedReversalReturnsTheWinnerMovedBeforeWithdrawError`

Both were confirmed **red first** against the cycle-4 tree
(`live verdict = miss` in each), then green after the change. No test was
weakened or rewritten to accommodate the fix.

### New, covering the real helper rather than a stand-in

| test | boundary |
|---|---|
| `TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack` | `Quarantine`'s own interior sync failure: reports no move *and* the entry is live |
| `TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller` | the same interior fault under a publication and under a reversal, asserting the two correct-and-different answers |
| `TestAQuarantineThatCannotPutTheEntryBackHandsItsCallerTheRecord` | the last exit: sync failed, return rename failed too, and the reported path is what the caller uses to restore |

The reversal case asserts `StateChanged` is **true** while the publication case
asserts it is **false** — deliberately different, because a publication that
recorded no predecessor has nothing to unwind and leaves the entry the helper
put back, whereas a reversal always leaves the entry this run published live.
Each case asserts the live verdict, a byte-identical `entryFingerprint`, the
exact live artifact bytes, and zero surviving private staging directories.

### Mutation checks (both layers independently proven)

| mutation | result |
|---|---|
| drop the move-back inside `quarantinePath` | `TestAQuarantineThatCannotBeMadeDurablePutsTheEntryBack` and `...HandsItsCallerTheRecord` **FAIL** |
| make `withdrawnTo` never report a move | both cycle-4 regressions **FAIL** |

`TestADurabilityFaultInsideAQuarantineIsCompensatedByItsCaller` survives the
first mutation, which is the layering working as intended: layer 2 catches what
layer 1 stopped doing.

---

## Gates

Each run as a standalone process; the exit code below is the real one.

| gate | exit | note |
|---|---|---|
| `go build ./...` | 0 | |
| `go vet ./internal/buildcache ./internal/install ./cmd/curator` | 0 | |
| `gofmt -l internal cmd` | 0 | empty output |
| `git diff --check` | 0 | |
| `golangci-lint run` (v2.4.0, whole repo) | 0 | `0 issues.` |
| `go test ./internal/buildcache -count=1` | 0 | 1.5s, whole package |
| `go test ./internal/install -count=1` | 0 | 199.5s, whole package |
| `go test ./cmd/curator -run '^TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails$' -count=1` | 0 | 93.4s, real CLI, install **and** upgrade |
| `go test ./cmd/curator -count=1` | see below | whole package |
| remaining packages | see below | |

---

## Scope notes

- `internal/buildcache` and `internal/install/commit.go` remain outside this
  task's primary ownership. The change is the minimum the cycle-4 verdict's
  rework instruction requires.
- `curator global status` remains an explicit contract-level exclusion
  (`TASK-260729-2kaopg`). Compiled install idempotence remains
  `TASK-260729-3jku56`.
- README: one sentence added to the build-cache paragraph, stating that the
  individual quarantine move holds to the same put-it-back rule as the
  publication around it.

---

## Findings worth carrying forward

1. **A two-step mutation must report the first step even when the second one
   fails.** "Returned an error" and "changed nothing" are different claims;
   returning `("", err)` after a successful rename asserts the second while only
   having established the first.
2. **A compensation that trusts its own helper is only as good as the helper's
   error contract.** Reconciling by observation under an exclusive lock costs one
   `ReadDir` on the error path and survives any future step inserted before the
   return.
3. **Symmetry is a correctness property here.** Cycle 4 put back a failed
   restoration but not a failed withdrawal, though both end in the same empty
   slot. Whenever two renames undo each other, both need the same rule.
4. Carried forward and still true: a cache entry cannot be atomically swapped
   with a journaled install; a returned `Journal.Commit` error does not by itself
   mean the transaction rolled back; a same-key cache replacement is the only
   cache change that stays a valid hit.
</content>
</invoke>
