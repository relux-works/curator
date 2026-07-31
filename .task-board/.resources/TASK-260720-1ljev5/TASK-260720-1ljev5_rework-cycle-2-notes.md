# TASK-260720-1ljev5 — rework cycle 2

Closes both P1 findings in `TASK-260720-1ljev5_review-verdict-cycle-2.md`.

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree`
(unchanged from the earlier cycles; nothing staged, committed, or published).
Platform: darwin/arm64, Go 1.25.5, `golangci-lint` 2.4.0.
Native Windows validation on host `win` (Windows 10 19045, `desktop-3pbo632\admin`),
non-elevated through a temporary `schtasks /rl LIMITED` task.

Predecessor worktrees stayed read-only. The accepted base
(`.temp/TASK-260720-1zntv0/worktree`) was only *read* and cross-compiled into
this task's temp directory for like-for-like Windows comparisons.

## Finding 1 (P1) — a repeated registry member bypassed the strict gate

`parseConsumers` decoded `consumers.json` into a struct with
`DisallowUnknownFields`. Go's struct decoding accepts a **repeated** known
member and keeps the last one, so

```json
{"schema_version": 1, "consumers": ["/live/checkout"], "consumers": []}
```

read as an empty *trusted* registry. Maintenance would rewrite it empty, the
next pass would no longer visit the live checkout, and the build artifact that
checkout still runs would be collected — the exact pass-one-forgets /
pass-two-sweeps failure cycle 1 set out to close, through a different door.

**The registry is now read as a token stream.** `parseConsumers` walks the
document with `json.Decoder.Token`, so ambiguity is visible instead of being
resolved silently:

- a repeated `schema_version` or `consumers` member is refused, and so is a
  repeat of any other member name;
- an unsupported member, a non-string member name, a non-array or null
  `consumers`, a non-string, empty, or relative checkout path, a nested list,
  and trailing content after the object are all refused;
- `schema_version` must be an exact JSON integer equal to 1 — `1.0`, `1e0`,
  `"1"`, and `null` are refused rather than rounded or coerced.

Repetition *inside* the consumers array stays accepted, deliberately.
Deduplicating a repeated checkout cannot drop a checkout, so it carries none of
the ambiguity that makes a repeated member unsafe; refusing it would only make
maintenance give up for a shape that loses nothing.

Because both writers already route through `readConsumers`, the refusal reaches
them for free: `RecordConsumer` and `StageConsumer` fail closed on an ambiguous
registry instead of merging into "whatever came last" and unregistering every
other checkout on the machine. `LoadConsumers` keeps its lenient read-only
contract and reports no consumers.

### Regressions

- `TestARepeatedRegistryMemberNeverEmptiesTheRegistry` — two consecutive passes
  over a registry whose `consumers` member is stated twice. Neither pass sweeps,
  both report the ambiguity, and the file bytes are byte-identical afterwards.
  It then repairs the registry and proves the **live reference survived**: the
  next pass marks exactly the build key the checkout's marker v2 names.
- `TestCollectStaysFailSafeAcrossConsecutivePasses/registry_repeating_a_known_member`
  — the same case inside the shared two-pass fail-safe runner.
- `TestWritersRefuseARepeatedRegistryMember` — `RecordConsumer` and
  `StageConsumer` both fail, the bytes are unchanged, and `LoadConsumers` reads
  no consumers.
- `TestParseConsumersAcceptsOnlyTheCanonicalRegistry` — extended with repeated
  members (both orders), a repeated version, fractional, exponent, string, and
  null versions, an object `consumers`, and a nested consumer list.
- `TestStructDecodingWouldTrustARepeatedRegistryMember` — an **in-suite negative
  control**: it runs the previous decoding strategy, shows it accepting the same
  document as an empty trusted registry, and then shows `parseConsumers`
  refusing it.

**Negative control** (`gates3/negative-control-registry.log`, exit **1**):
restoring the struct decoder makes 5 test functions fail, including all three
new ones.

## Finding 2 (P1) — decisive classification reopened the mutable pathname

`inspectUnexpected` proved the candidate, asserted its parent was the proven
cache root, and read its receipt through that handle — and then called
`store.inspectEntry(entryPath, …)`, which started a **fresh traversal from the
manager-home pathname**. On Unix an open descriptor does not pin later pathname
resolution, so a cache-root exchange in that window let the exact-content,
receipt, and artifact-hash checks validate a *replacement* entry while
`retireEntry` removed the original through the mutator still bound to the proven
root: validate one directory, delete another.

**The whole decision now runs on the proof.** `inspectProvenEntry` classifies
from the candidate's own descriptor:

- `openProtectedEntryFrom` opens the receipt, the `bin` directory, and the
  artifact **relative to the proven entry handle** — on Unix with `openat` and
  `O_NOFOLLOW` at every step, validating type, owner, and mode as before;
- the exact-members listing reads the caller's proven directory handle
  (`openedEntry.borrowedEntryDir`), so no pathname can redirect it;
- receipt decode, canonical receipt validation, receipt hash, artifact open,
  hash, and size all read those handles;
- the publication time is `Stat` on the proven handle, as before.

Nothing in the sweep path reopens `store.home`, the cache-root path, or the
candidate pathname after the boundary is proven. `Store.Inspect` — the ordinary
cold lookup, which owns no prior proof — keeps the pathname entry point;
`classifyEntry` is the single decision procedure both share, so the two paths
cannot drift apart.

On Windows the exchange is refused one layer lower: every component from the
manager home down to the candidate is held open **without `FILE_SHARE_DELETE`**,
so none of them can be renamed or deleted while the sweep runs. The Windows
`openProtectedEntryFrom` additionally asserts, before opening any child, that
the entry pathname still names the very object the handle holds
(`os.SameFile` against a fresh `Lstat`), and the decisive member listing comes
from the held handle either way.

### Regressions

- `TestSweepClassificationSurvivesACacheRootExchangedMidPass` (Unix and Windows)
  — a new `beforeClassifyForTests` seam exchanges the cache-root pathname
  **during classification**, after the boundary and receipt are proven. The
  original candidate is structurally unexpected (and backdated *after* the
  write, so it cannot be retained merely for being young); the replacement is a
  fully valid entry published under the same logical key in a second protected
  home. Nothing is removed, the original and the replacement are both intact,
  and the candidate is retained with an `unexpected contents` warning.
- The same test ends with an **in-test negative control**: it classifies the
  cache-root pathname directly and requires the replacement to come back a
  `Hit` — i.e. the verdict the old reopen would have borrowed for the entry it
  was about to delete.
- The Windows variant accepts either outcome of the swap and asserts the same
  guarantee in both. On the native run the OS refused the exchange
  ("The process cannot access the file because it is being used by another
  process") and the test still passed.

**Negative control** (`gates3/negative-control-classification.log`, exit **1**):
restoring `store.inspectEntry(entryPath, …)` makes the new test fail with
*"an entry was retired on someone else's proof"* — the reviewer's defect,
reproduced deterministically.

## Task-only delta of this cycle

| File | Change |
|---|---|
| `internal/scopes/consumers.go` | duplicate-aware token parsing of the registry |
| `internal/buildcache/cache.go` | `inspectProvenEntry`, shared `classifyEntry`, borrowed entry handle, `artifactChildName` |
| `internal/buildcache/collect.go` | classification bound to the proven handle; `beforeClassifyForTests` seam |
| `internal/buildcache/protection_unix.go` | `openProtectedEntryFrom` via `openat`/`O_NOFOLLOW` |
| `internal/buildcache/protection_windows.go` | `openProtectedEntryFrom` with entry-identity revalidation |
| `internal/buildcache/protection_unsupported.go` | matching stub |
| `README.md` | ambiguous-registry and handle-bound-classification rules |
| tests | `gc_conservative_test.go`, `collect_test.go`, `collect_unix_test.go`, `collect_windows_test.go` extended |

## Gate evidence

Every command was run standalone; the exit code reported is the real one.
Logs in `TASK-260720-1ljev5_gate-evidence-cycle-3.tar.gz`.

### macOS (darwin/arm64), primary

| Gate | Log | Exit |
|---|---|---|
| `gofmt -l .` | `gofmt.log` | 0 (empty) |
| `git diff --check` | `diffcheck.log` | 0 |
| `go build ./...` | `build.log` | 0 |
| `go vet ./...` | `vet.log` | 0 |
| `go test ./... -count=1` (40 packages) | `test-all.log` | 0 |
| `go test -race ./internal/scopes ./internal/buildcache ./cmd/curator` | `race-scoped.log` | 0 |
| `go test -race ./internal/install -run 'TestPostCommitMaintenance\|TestMaintenanceFailureAfterCommitIsAWarning\|TestConcurrent'` | `race-install.log` | 0 |
| `golangci-lint run ./...` | `lint.log` | 0, **0 issues** |
| `GOOS=windows GOARCH=amd64 go build ./...` | `crossbuild-windows.log` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | `crossbuild-linux.log` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | `crossvet-linux.log` | 0 |

### Expected-red gate, honestly reported

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1** (`crossvet-windows.log`):

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

Pre-existing and sibling-owned (`internal/runtimestore`, `TASK-260720-29hi1h`).
The same command against the untouched accepted base exits 1 with the identical
message (`crossvet-windows-base.log`). Excluding that one package, the same vet
over this tree exits **0** (`crossvet-windows-scoped.log`).

### Negative controls (expected red, by construction)

| Control | Log | Exit |
|---|---|---|
| struct decoding restored → registry tests | `negative-control-registry.log` | 1 |
| pathname classification restored → swap test | `negative-control-classification.log` | 1 |

### Native Windows

| Suite | Log | Exit |
|---|---|---|
| `internal/scopes` | `win/win-scopes.log` | 0 |
| `cmd/curator -test.run TestGC` | `win/win-curator-gc.log` | 0 |
| `internal/buildcache` | `win/win-buildcache.log` | 1 (inherited failures only) |
| `TestAtomicPublication` ×5, this tree | `win/win-mine-race.log` | 1 |
| `TestAtomicPublication` ×5, accepted base | `win/win-base-race.log` | 1 |

Like-for-like, five runs of each binary in the same non-elevated session:

| Test | This tree | Accepted base |
|---|---|---|
| `TestAtomicPublicationConflictingRace` | 3 pass / 2 fail | 1 pass / 4 fail |
| `TestAtomicPublicationIdenticalRace` | 0 pass / 5 fail | 0 pass / 5 fail |

Both binaries fail with the identical message — `prepare protected cache root:
untrusted cache provenance: … DACL is not protected from inheritance` — so the
flakiness is the sibling-owned `ensureProtectedBase` race, not this tree; this
tree is no worse than the base on either test.

Every sweep test passes natively, including both new classification-swap cases
and the four registry regressions. Inherited failures, kept separate from this
task's gate:

- `TestAtomicPublicationIdenticalRace` — the DACL-inheritance race inside
  sibling-owned `ensureProtectedBase` (`TASK-260720-3pwg2w`), flaky on the
  accepted base at the same rate.
- `TestWindowsProtectedStateMatrix/artifact_has_inherit-only_owner_allow` —
  fails on the accepted base as well.

Honestly reported skips, none of them task-owned:
`TestWindowsProtectedStateMatrix/artifact_reparse_point` needs a **file**
reparse point (`SeCreateSymbolicLinkPrivilege` / Developer Mode);
`TestCandidateCacheOutcomeVocabulary` needs `CURATOR_CONFORMANCE_ROOT`; the two
`unreadable directory` scopes subtests have no POSIX meaning on Windows and
self-diagnose.

## Findings worth carrying forward

1. **`DisallowUnknownFields` is not a strictness gate.** It says nothing about
   repeated members, and Go's decoder keeps the last one. Any JSON document
   whose *absence of content* is a safety-relevant reading has to be parsed with
   duplicate-aware tokens or compared against canonical bytes.
2. **Proving a boundary and then re-deriving it are different things.** The
   cycle-1 fix bound the *mutation* to the proven object; the decision that
   authorises the mutation has to be bound to the same object, or the two halves
   can be made to talk about different directories.
3. **A "retained" outcome is not automatically evidence.** The first draft of
   the classification-swap test wrote into the entry after backdating it, so the
   entry was retained for being *young*; the negative control is what exposed
   that the test was proving the wrong thing.
