# TASK-260720-1ljev5 — rework cycle 1

Closes every finding in `TASK-260720-1ljev5_review-verdict-cycle-1.md`.

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree`
(unchanged from cycle 1; nothing staged, committed, or published).
Platform: darwin/arm64, Go 1.25.5, `golangci-lint` 2.4.0.
Native Windows validation on host `win` (Windows 10 19045, `desktop-3pbo632\admin`).

Predecessor worktrees stayed read-only. The accepted base
(`.temp/TASK-260720-1zntv0/worktree`) was only *read* — and cross-compiled once,
into this task's temp directory, to produce the like-for-like Windows baseline
binary demanded by finding 3.

## Finding 1 (P1) — an uncertain project could be forgotten

Cycle 1 warned about an uncertainty and then destroyed the state that produced
it: `Collect` rewrote `consumers.json` before checking `marked.uncertain`, and a
project whose markers could not be read was dropped from the registry. The
second pass then saw a complete-looking reference set and could sweep builds
that were still referenced.

The mark phase is now conservative in the only way that survives a second pass.

**Nothing untrusted is rewritten.** `pruneConsumers` refuses to touch a registry
that could not be parsed, in `Collect` *and* in `CollectRuntime`. The previous
cycle recorded "a corrupt registry wipes every consumer" as a pre-existing
runtime-GC hazard worth its own task; it is the same two-pass hazard through a
different entry point, so it is closed here rather than deferred.

**Nothing uncertain is unregistered.** A consumer is dropped only once its scope
is proven absent or proven valid and empty. Any uncertainty in the scope keeps
the checkout registered, so the next pass visits the same markers again.

**The registry shape is validated, not guessed.** `parseConsumers` accepts only
the exact document Curator writes: a JSON object, `schema_version` present and
equal to 1, `consumers` present as an array of absolute, non-empty paths, no
unknown fields, no trailing content. Everything else is "unknown consumers", not
"no consumers" — the distinction the whole fail-safe rests on. `LoadConsumers`
keeps its lenient read-only contract; the two callers that *write* the registry
(`RecordConsumer`, `StageConsumer`) now fail closed instead of merging into an
empty set and silently unregistering every other checkout.

**Redirected and unreadable scopes are refused, not skipped.** `readScope`
lstats the skills root and every member. A symbolic link or reparse point at
either level is an uncertainty rather than something to follow — the markers
holding the live build keys are still on disk, just not where maintenance
looked. Every non-absence metadata failure is an uncertainty too, including a
marker that cannot be inspected or is not a regular file.

Windows needs the extra care: Go has mapped junctions onto different mode bits
across releases, so `isRedirect` checks `FILE_ATTRIBUTE_REPARSE_POINT` as well
as `ModeSymlink`/`ModeIrregular` (`internal/scopes/redirect_windows.go`, with a
plain mode check in `redirect_other.go`).

**One deliberate non-blocker.** A *non-directory* member (the `.DS_Store` a file
browser leaves behind) cannot hold an install marker and therefore cannot hide a
reference. Treating it as an uncertainty would disable build-cache collection
permanently on any machine someone has browsed in Finder. It is reported as a
warning and does not block the sweep; `TestNonDirectoryScopeMembersDoNotBlockMaintenance`
pins that split.

### Regressions

`TestCollectStaysFailSafeAcrossConsecutivePasses` runs every case twice and
asserts the second pass is exactly as unwilling as the first — corrupt registry,
registry of an unknown schema, invalid-only marker project, unreadable installed
skill directory, skill directory replaced by a file, member replaced by a file.
`TestCollectStaysFailSafeOnRedirectedUnixScopes` and
`TestCollectStaysFailSafeOnRedirectedWindowsScopes` do the same for redirected
project, global, and hybrid scope roots and redirected installed skills.
`TestParseConsumersAcceptsOnlyTheCanonicalRegistry` pins the shape validation and
`TestRecordingAConsumerNeverOverwritesAnUntrustedRegistry` pins the write path.

**Negative control** (`gates/negative-control-twopass.log`): reverting just the
two cycle-1 pruning decisions makes 8 of these subtests fail, so they are real
regression coverage, not restatements of the implementation.

## Finding 2 (P1) — retirement was not bound to the validated root

Cycle 1 proved the boundary with an open handle and then mutated by pathname.
On Unix an open descriptor does not pin later pathname resolution, so exchanging
the cache-root path between validation and retirement redirected the rename and
the deletion into the replacement tree.

`Sweep` now opens a `sweepRoot`: the validated `protectedDir` (type, owner, and
mode or DACL of every component, resolved with `O_NOFOLLOW` /
`FILE_FLAG_OPEN_REPARSE_POINT`) plus an `os.Root` mutator opened on the same
directory and **accepted only when `os.SameFile` says it is the same directory
object**. That comparison is the object-identity revalidation: a path exchanged
before mutation is refused rather than followed, on Unix by inode and on Windows
by volume serial and file index.

Every mutation of the pass then goes through that mutator, never through a
pathname: `root.mutator.Rename` retires the entry, `root.mutator.RemoveAll`
finishes the deletion and the resumable cleanup of an interrupted removal, and
`syncDirHandle` fsyncs the proven handle instead of re-resolving the directory.
`os.Root` resolves every component relative to the held root with no-follow
semantics and refuses any name that escapes it, so this is the standard
library's hardened primitive rather than a hand-rolled one.

The classification path is bound too: `inspectUnexpected` re-resolves the entry
without following a link and then asserts, via `os.SameFile`, that its parent is
the very directory object the sweep proved. A candidate that no longer lives in
the proven root is retained and reported.

`retireEntry` still rejects any name that is not a direct child, and the private
retirement name is now drawn from `crypto/rand` and checked for absence through
the root handle rather than materialised by `os.CreateTemp`.

### Regressions

`TestSweepRemovalSurvivesACacheRootExchangedMidPass` (Unix and Windows) is the
adversarial swap: a `beforeRetireForTests` seam exchanges the cache-root pathname
for a replacement tree after the boundary is proven and before the first
removal. The replacement must be untouched, the proven entry must be the one
that goes, and every candidate that now resolves outside the proven root must be
retained and reported.

Windows answers one component earlier: the proven handles are opened without
`FILE_SHARE_DELETE`, so the root cannot be renamed at all while the sweep holds
it. The Windows test accepts either outcome and asserts the same guarantee in
both. On the native run the swap was refused by the OS and the replacement was
untouched.

`TestRetireEntryRefusesAnUnboundRoot` and
`TestSweepRefusesACacheRootExchangedBeforeMutation` cover the unbound and
mismatched-identity paths directly.

**Negative control** (`gates/negative-control-boundary.log`): restoring
pathname-based `os.Rename`/`os.RemoveAll` while keeping everything else makes the
swap test fail with *"the sweep acted inside the replacement tree"* — the exact
defect the reviewer described.

## Finding 3 (P2) — Windows evidence was red and internally inconsistent

Two causes, both addressed.

**The cases that skipped now run.** The `internal/scopes` end-to-end maintenance
tests skipped because their fixture could not build manager-protected Windows
state; `protectedTestHome` is now platform-split and applies an
inheritance-protected owner-only DACL directly
(`gc_integration_windows_test.go`), so those tests execute natively. The reparse
cases skipped because the test account has no `SeCreateSymbolicLinkPrivilege`;
`linkDirectoryForTest` now falls back to a **junction**, which needs no
privilege, so the reparse coverage actually runs.

**The contradiction is resolved.** Cycle 1 reported
`TestAtomicPublicationConflictingRace` failing on this tree while passing on the
accepted base. Five runs of each binary, in the same non-elevated session on the
same host:

| Test | This tree | Accepted base |
|---|---|---|
| `TestAtomicPublicationConflictingRace` | 3 pass / 2 fail | 4 pass / 1 fail |
| `TestAtomicPublicationIdenticalRace` | 1 pass / 4 fail | 0 pass / 5 fail |

Both tests are nondeterministic on **both** trees, with an identical message:
`prepare protected cache root: untrusted cache provenance: ... DACL is not
protected from inheritance`. That is a race inside the sibling-owned
`ensureProtectedBase` (`TASK-260720-3pwg2w`), where a concurrent publisher
observes a freshly created directory between `Mkdir` and the DACL being applied.
Nothing in this task touches that path. Cycle 1's single-run comparison happened
to catch opposite sides of the same coin flip.

**Native Windows results, cycle 2** (non-elevated via a temporary
`schtasks /rl LIMITED` task, deleted afterwards — an elevated session makes
Windows assign `BUILTIN\Administrators` as owner of everything the process
creates, so every protected-state fixture is rejected for reasons unrelated to
the code):

| Suite | Exit | Result |
|---|---|---|
| `internal/scopes` | **0** | every maintenance test **runs and passes**, including the three end-to-end protected-store tests and the three junction-redirect tests |
| `cmd/curator -test.run TestGC` | **0** | all three lock-serialization tests pass |
| `internal/buildcache` | 1 | **every sweep test passes**, including both reparse tests and the boundary swap; the failures are inherited |

Inherited `internal/buildcache` failures, kept separate from this task's gate:

- `TestAtomicPublicationIdenticalRace`, `TestAtomicPublicationConflictingRace` —
  flaky on the accepted base too, table above.
- `TestWindowsProtectedStateMatrix/artifact_has_inherit-only_owner_allow` — fails
  on the accepted base as well (cycle-1 evidence, unchanged).

Honestly reported remaining skips, none of them task-owned coverage:

- `TestWindowsProtectedStateMatrix/artifact_reparse_point` needs a **file**
  reparse point. A junction is directory-only, so this one genuinely requires
  `SeCreateSymbolicLinkPrivilege` (Developer Mode). It belongs to
  `TASK-260720-3pwg2w`. Enabling Developer Mode is a persistent change to the
  validation host, so it was not made unilaterally; it is the recommended way to
  close that gap.
- The two `unreadable directory` subtests skip on Windows because `chmod 0000`
  has no POSIX meaning there. They self-diagnose with the reason.

## Gate evidence

Every command was run standalone; the exit code reported is the real one.
Logs in `TASK-260720-1ljev5_gate-evidence-cycle-2.tar.gz`.

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

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**
(`crossvet-windows.log`):

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
| pathname retirement restored → swap test | `negative-control-boundary.log` | 1 |
| cycle-1 consumer pruning restored → two-pass tests | `negative-control-twopass.log` | 1 |

### Native Windows

| Suite | Log | Exit |
|---|---|---|
| `internal/scopes` | `win/win-scopes.log` | 0 |
| `cmd/curator -test.run TestGC` | `win/win-curator-gc.log` | 0 |
| `internal/buildcache` | `win/win-buildcache.log` | 1 (inherited failures only) |
| `TestAtomicPublication` ×5, this tree | `win/win-mine-race.log` | 1 |
| `TestAtomicPublication` ×5, accepted base | `win/win-base-race.log` | 1 |

## Task-only delta of this cycle

| File | Change |
|---|---|
| `internal/buildcache/collect.go` | `sweepRoot`, handle-bound retirement, parent-identity check, `crypto/rand` name reservation, test seam |
| `internal/buildcache/protection_{unix,windows,unsupported}.go` | `syncDirHandle`; `protectedDir.parents` typed as `[]*os.File` |
| `internal/scopes/gc.go` | conservative retention, redirect refusal, uncertain/advisory split, `pruneConsumers` |
| `internal/scopes/consumers.go` | strict `parseConsumers`, fail-closed `RecordConsumer` |
| `internal/scopes/stage.go` | fail-closed `StageConsumer` |
| `internal/scopes/redirect_{windows,other}.go` | new — reparse detection |
| `README.md` | conservative-retention and handle-bound-removal rules |
| tests | `gc_conservative{,_unix,_windows}_test.go`, `gc_integration_{other,windows}_test.go` new; `collect{,_unix,_windows}_test.go` and `gc_integration_test.go` extended |

## Findings worth carrying forward

1. **The Windows publication race is real, not a harness artifact.**
   `ensureProtectedBase` creates a directory and applies its protected DACL in
   two steps, so a concurrent publisher can observe the gap. It reproduces on the
   accepted base at roughly the same rate and is owned by `TASK-260720-3pwg2w`.
2. **`os.Root` is the right primitive for bounded mutation.** It gives
   handle-relative, no-follow, escape-refusing rename and recursive removal on
   both platforms; pairing it with an independent `os.SameFile` identity check
   against our own validated handle closes the validate-then-mutate window
   without hand-rolled `*at` plumbing.
3. **Windows validation needs a non-elevated session, and junctions.**
   `schtasks /rl LIMITED` gives real protected-state coverage from an elevated
   SSH shell, and junctions give real reparse coverage without
   `SeCreateSymbolicLinkPrivilege`. Both are now the recommended recipe.
