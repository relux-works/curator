# TASK-260720-1nlmvv — cycle 3: answering the cycle-2 review verdict

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce plus the composed accepted predecessor tree; nothing
staged, committed, or published). Platform: darwin/arm64, Go 1.25.5,
golangci-lint v2.4.0 (pinned).

Cycle 2 was returned with **CHANGES REQUESTED** and four blocking findings. This
document is the delta against cycle 2, finding by finding. Every cycle-2 change
that was not named by the verdict is untouched.

---

## 1. P1 — a failed run now puts the shared build cache back

### The defect, restated

`runCommit` published build winners before `stageTargets`, journal planning,
preparation, and commit. `buildcache.Publish` quarantines a corrupt or untrusted
live entry and selects the replacement immediately. The cache is not a
transaction target, so any later fault restored the installed scope and left the
predecessor quarantined with the replacement live — a persistent change made by
an operation that committed nothing. It is not only a purity violation: it
silently rewrites what `status` says about the *previous* installation, from
`corrupt-build-cache` (which install repairs) to receipt or artifact drift
against an entry nobody asked for.

### Why not a journaled cache target

That was the verdict's first option and it does not fit. A launcher can only
point at a protected entry that is already live, so the entry has to exist
before the targets that reference it can be staged at all. Making the cache
entry a journaled target would move the swap after preparation and require the
transaction engine to own protected 0700 directories, their sidecars, and their
interaction with the sweep — a redesign of the transaction, staging, and GC
boundary, none of which this task owns. The verdict's second option —
*reversibly staged through every post-publication failure* — is what landed.

### What landed

`buildcache.PublicationResult` gains `Quarantined`, the path an unusable
predecessor was moved aside to, and the store gains `Revert`:

```go
func (store *Store) Revert(key buildmeta.CacheKey, published PublicationResult, lock HomeLock) error
```

It withdraws the entry this run published — by the same quarantine rename, so no
byte is deleted and the ordinary sweep collects the leftovers — and renames the
predecessor back. Both steps are renames inside the protected cache root under
the same caller-held home lock the publication ran under. It fails closed:
a reused winner changed nothing and is a no-op, a missing lock is refused, and a
quarantine path outside the cache root is refused *before* anything moves, so a
refusal is never half-applied.

`runCommit` registers the compensation the moment `publishWinners` returns —
including when it returns an error, because a publication that failed on the
*second* command still displaced the first. `publishWinners` therefore returns
the publications made so far on every path.

### When the reversal fires, and when it must not

Two independent facts have to hold (`reversiblePublication`):

1. every target this run journaled is back at the exact preimage it recorded
   under the home lock, so the installation really did return to its prior state;
2. no journal still references the build keys this run published, so no durable
   transaction is left whose completion recovery — not this process — owns.

Either one unmet keeps the published entries. The asymmetry is the whole point:
an entry no installation references is valid and the sweep collects it, while
restoring an unusable predecessor over the entry a recovered commit is about to
point at would turn a reported failure into a broken installation. That is also
what closes the narrow window the naive "revert on any commit error" rule would
have opened — a journal-write failure at the `Prepared → Committing` transition
leaves every target at its preimage while recovery will still commit forward.
Check 2 sees that journal and refuses.

A run that could not put the cache back says so. `commitOutcome.retainedBuilds`
becomes `Result.BuildCacheRetained`, and `repairNotices` uses it, so the operator
never reads "the live build cache is unchanged" when it is not.

### Proof

- `TestRevertRestoresExactlyWhatAPublicationDisplaced` (real protected
  directories): a cold publication is withdrawn to a miss; a replacement goes
  back to the corrupt predecessor with the exact bytes it had.
- `TestRevertFailsClosed`: no lock, reused winner, and a quarantine path planted
  outside the cache root — the last one asserts the live entry is still a hit,
  so the refusal did not withdraw anything on its way to refusing.
- `TestAFailedCommitRestoresTheBuildCacheItReplaced`: preparation fault and
  commit fault × {corrupt predecessor, untrusted predecessor, cold miss}. Each
  asserts the run really did rebuild and publish first, that the reversal ran,
  that the exact prior verdict is back, and that the whole installed-state
  snapshot is unchanged.
- `TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced`: the same through
  the **real** durable transaction with a target fault at `PointAfterBackup`.
- `TestAnInFlightTransactionKeepsThePublishedCacheEntry`: the one direction the
  reversal must refuse, including the operator warning and
  `Result.BuildCacheRetained`.
- `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` — the E2E the verdict
  asked for, for **install and upgrade**, through the real CLI with no
  fault-injection seam: the repair really happens, then the context store is
  made unwritable so the durable commit fails. It asserts exactly one withdrawn
  entry exists (which is only true if publication happened and was reversed),
  that the live artifact bytes are the corrupt ones again, that exactly one live
  entry remains, that the marker is untouched, that `status` reports
  `corrupt-build-cache` again, and that the ordinary path still repairs it
  afterwards. The permission trick is verified to actually block writes and the
  case skips rather than passing vacuously if the process can write through it.

---

## 2. P1 — every build-boundary commit failure is redacted

`failBuild` covered planning and staging; `Project` and `Global` still sent
`commitErr` through raw `failf`, so publication text — which can carry
operation-private artifact and cache locations, receipt bytes, and driver
output — reached `Result.Errors` verbatim.

Redacting *every* commit error would be wrong: journal and target failures name
the operator's own project paths, which are the actionable part of the message.
So the boundary is marked instead. `buildPhaseError` wraps every failure that
originated at the protected build cache — publication, receipt encoding, the
publish fault, the reversal, and the two target-staging refusals derived from a
`buildcache.Result` (which carries the private absolute artifact path through
`runtimestore.CompiledTargetFromCache`). `Result.failCommit` is the single sink
for both scopes:

```go
func (r *Result) failCommit(commitErr error) (Result, *restartError) {
    // restart -> retry; buildPhaseError -> failBuild + boundary code; else failf
}
```

`TestBuildPublicationFailuresAreRedactedInTheResult` is the canary: an absolute
`/Users/...` cache path in a publication failure comes out as `<path>` with no
`/Users/` anywhere, and a 400× repeated detail carrying an ANSI escape and an
embedded newline comes out at most `maxDiagnosticRunes` long with no control
characters at all.

---

## 3. P2 — toolchain-refusal rows carry the identities they already had

`toolchainInventory` left `driver` and `build_source` empty, so the human line
printed `driver=` and JSON carried an empty `build_source` object — even though
build sources are validated *before* the toolchain is consulted and the driver
is closed by the schema v6 parser.

Every identity that existed at the moment of the refusal is now carried: the
declared closed driver, the build root, the package directory, and the validated
build-source digest. Only what genuinely depends on the toolchain — native
target, logical key, cache verdict — stays empty, so a row still never publishes
an identity nothing derived.

Pinned at the plan boundary
(`TestToolchainFailurePlansAnInventoryOfEveryActiveCommand`, which also asserts
the rendered line reports the source and that no cache lookup is exposed) and
end to end through the CLI
(`TestStatusReportsAnUnusableToolchainPerCompiledCommand`), where the human line
and the JSON row are compared field for field — `driver`, `root`, `dir`,
`source`, and the planner outcome present; `target=`, `key=`, `artifact=`
absent.

---

## 4. P1 — the cache is re-read before the final status verdict

`status` fingerprinted install markers before and after, but the cache was
inspected once, inside the lock-free dry-run plan. A same-marker removal,
corruption, replacement, or sweep between that inspection and the output could
publish a stale `current` — or a stale *drift* verdict, which is just as wrong
and much less obvious.

`PlannedBuild.Expectation()` exposes the exact read-only lookup each row's
verdict was taken with (empty for a command whose identity was never derived).
After classification, `recheckBuildCache` re-takes precisely those lookups
through a read-only `buildcache.Store` and compares the evidence the
classification actually used: the reusable outcome, and for a hit the receipt
identity and artifact metadata. Any difference — or a cache boundary that cannot
be opened at all — makes the skill `build-state-changed`, which fails `--check`.
The row now says *which* half moved, marker or cache.

`TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` moves the cache
for real between the plan and the classification, four ways: corrupted, removed,
protection dropped, and **replaced by a different but equally valid entry for
the same logical key** — the case that produces a stale drift verdict rather
than a stale current one, and the only one a re-read can catch. Each case
asserts the install marker is byte-identical, so nothing but the cache recheck
can have produced the verdict.

---

## Scope notes

- `internal/buildcache` and `internal/install/commit.go` are outside this task's
  primary ownership. The additions there are the minimum the verdict's own
  rework instruction requires (`Revert`, `Quarantined`, the reversal decision)
  and add no new command, flag, or wire schema.
- `curator global status` remains an explicit contract-level exclusion, tracked
  as `TASK-260729-2kaopg`. Compiled install idempotence remains
  `TASK-260729-3jku56`.

---

## Evidence

Reported exit codes are the real status of each command, run standalone. Raw
logs are in the attached `TASK-260720-1nlmvv_gate-evidence-cycle-3.tar.gz`.

### macOS (darwin/arm64), primary

| Gate | Exit |
|---|---|
| `gofmt -l .` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1 -timeout 90m` | see gate log |
| `golangci-lint run ./...` (pinned v2.4.0) | 0, **0 issues** |
| `go test -race -timeout 60m ./internal/install -count=1` | see gate log |
| `go test -race -timeout 60m ./cmd/curator -count=1` | see gate log |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |

### Expected-red gate, reported as failing

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

Pre-existing and sibling-owned (`internal/runtimestore`, `TASK-260720-29hi1h`).
Measured again this cycle, not assumed: the identical command on the untouched
accepted base (`.temp/TASK-260720-1ljev5/worktree`) also exits **1** with the
identical message, and excluding that one package this tree's Windows vet exits
**0**.

### Not run

- Native Linux execution. The change is platform-neutral planning, presentation,
  and rename-based cache compensation; `GOOS=linux` build and vet are clean.
- Native Windows execution this cycle. The cycle-2 native Windows evidence still
  applies to the redaction and classification suites, which this cycle did not
  change; the new tests are either CLI-level (unrunnable on that host, which has
  neither `git` nor a Go toolchain — pre-existing, tracked for
  `TASK-260720-1pvfj5`) or exercise POSIX-mode-based fault injection that skips
  itself where it cannot block a write.

---

## Findings worth carrying forward

1. **A cache entry cannot be atomically swapped with a journaled install.** The
   artifact must be live before the targets that reference it are staged, and it
   is not a transaction target, so the only correct shapes are compensation
   (this cycle) or teaching the transaction engine to own protected directories
   and their sidecars. Anything that wants true atomicity here is a
   transaction-layer change, not a status or install change.
2. **A returned `Journal.Commit` error does not by itself mean the transaction
   rolled back.** The engine rolls back in-process on a target fault, but a
   journal-write failure at `Prepared → Committing` leaves a journal recovery
   will commit *forward*. Any compensating action taken on a commit error has to
   consult the live journal, not just the error.
3. **A same-key cache replacement is the only cache change that stays a valid
   hit,** so it is invisible to every check except re-reading the receipt
   identity. It is also the only one that produces a stale *drift* verdict
   rather than a stale current one. Worth remembering for any future
   currentness surface.
4. Carried from cycle 2 and still true: `internal/install` and
   `internal/install/atomicity` exceed the default `go test` timeout under load;
   the suite is disk-hungry enough that `ENOSPC` surfaces as an ordinary
   installation failure; the Windows validation host still lacks `git` and Go.
