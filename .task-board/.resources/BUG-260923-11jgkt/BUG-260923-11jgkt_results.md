# BUG-260923-11jgkt results

## Hosted root cause evidence

Downloaded run 35855672743's `test-evidence-windows-latest` artifact and inspected `test/go-test-served.json`. The workflow run was created at 2026-09-23 11:38:05 UTC; the failing test event was recorded at 2026-09-23 11:56:59.1577912 UTC on the `windows-latest` runner, package `github.com/relux-works/curator/internal/snapshot`:

```text
snapshot_test.go:61: worker 11 = "", snapshot destination conflicts with immutable commit: authenticate destination: open C:\Users\RUNNER~1\AppData\Local\Temp\TestConcurrentGetAcceptsOneImmutablePublisher224870753\002\cache\internal\skill-a\1112d93d63e9641cdfc52d7f2f46896fde5afe19\snapshot: The process cannot access the file because it is being used by another process.; want "C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\TestConcurrentGetAcceptsOneImmutablePublisher224870753\\002\\cache\\internal\\skill-a\\1112d93d63e9641cdfc52d7f2f46896fde5afe19\\snapshot"
```

The artifact captures the Windows sharing-violation message, but not the numeric errno. This maps to Win32 `ERROR_SHARING_VIOLATION` (32). This is a product defect: `Get` can reach `authenticateDestination` after a concurrent publisher installs the same immutable directory; `transaction.DigestPath` then fails opening that destination, and the old code mapped every digest read error to `ErrDestinationConflict`.

I searched five earlier hosted Windows artifacts for this same test. Runs 35855550263, 35852658095, 35758851581, 35739653315, and 35688466491 passed; the target run failed. The sampled frequency is 1 failure in 6 runs, confirming intermittency rather than a deterministic test-harness race.

## Fix and refusal rows

`digestDestinationWithSharingRetry` retries only a destination digest read whose error chain contains Windows `ERROR_SHARING_VIOLATION`: three retries, 10 ms between attempts (four attempts maximum). It does not retry access denied or other errors. A persistent sharing violation still returns `ErrDestinationConflict`; a successfully read destination with different bytes still returns `ErrDestinationConflict`.

- Transient row: Windows-only `TestConcurrentGetRetriesInjectedWindowsSharingViolation` injects code 32 through concurrent calls to the public `Get` entry point. It runs five iterations of 16 simultaneous readers; every result must return the same published path and content.
- Persistent row: `TestGetFailsClosedAfterPersistentWindowsSharingViolation` asserts four total digest attempts and a conflict.
- Genuine/access-denied rows: `TestGetDoesNotRetryWindowsAccessDenied` asserts one attempt and a conflict; existing `TestGetCachesByCommit` tampers with a cached tree and confirms a real `Get` still rejects it.

`CHANGELOG.md` records the bounded retry and fail-closed cases. The Windows-only rows compile locally, but cannot execute on this macOS host. Per campaign rules, the post-fix hosted landing suite is run once by `task-board handoff`; that Windows result is pending. The injected Windows test performs five repeated concurrent scenarios within that single hosted run.

## Validation run by this producer

All listed commands were run as standalone processes.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/snapshot` | 0 | Passed on macOS |
| `go test -count=50 -run '^TestConcurrentGetAcceptsOneImmutablePublisher$' ./internal/snapshot` | 0 | Passed 50 local concurrent runs |
| `go test -race ./internal/snapshot` | 0 | Passed |
| `env GOOS=windows GOARCH=amd64 go test -c -o /tmp/BUG-260923-11jgkt-snapshot.test.exe ./internal/snapshot` | 0 | Windows test binary compiled; not executed locally |
| `go vet ./internal/snapshot` | 0 | Passed |
| `go build ./internal/snapshot` | 0 | Passed |
| `env GOOS=windows GOARCH=amd64 golangci-lint run --allow-serial-runners ./internal/snapshot` | 0 | 0 issues, including Windows build-tagged files |
| `golangci-lint run` | 3 | Refused because another lint process held the lock |
| `golangci-lint run --allow-serial-runners` | 0 | 0 issues; emitted a warning about a deleted external temp-clone file while filtering results |
| `gofmt -l internal/snapshot/snapshot.go internal/snapshot/destination_sharing_violation_windows.go internal/snapshot/destination_sharing_violation_other.go internal/snapshot/snapshot_windows_test.go` | 0 | No files listed |
| `git diff --check` | 0 | Passed |

## Handoff note

The required hosted post-fix Windows gate has not run in this producer turn. The campaign instructions require the single landing suite to run at handoff, so no extra full hosted gate was started. The hosted lane must execute the five-iteration Windows regression before its result can be reported green.

## Revision 2 (refresh)

Revision 1 was accepted; this revision contains no product change. The required `task-board worktree refresh-candidate BUG-260923-11jgkt` exited 0 and advanced the Story worktree to trunk `fad881368b632f43a18b01681a8fc7def110cbae`. The old working tree still carried the 48da2690 base contents, so I rebuilt the uncommitted candidate from refreshed `HEAD` plus the exact accepted rev1 patch. `git apply --check /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/BUG-260923-11jgkt/BUG-260923-11jgkt_change-request_rev1.patch` and `git apply /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/BUG-260923-11jgkt/BUG-260923-11jgkt_change-request_rev1.patch` both exited 0. No commit was made.

Per-file candidate delta vs refreshed `HEAD` matches the accepted five-path patch: `CHANGELOG.md` (+5 lines); `internal/snapshot/destination_sharing_violation_other.go` (+7); `internal/snapshot/destination_sharing_violation_windows.go` (+13); `internal/snapshot/snapshot.go` (+22/-1); and `internal/snapshot/snapshot_windows_test.go` (+152). `git status --short` showed the two tracked modifications and three patch-added Go files. Incoming trunk changes are retained: `.github/ci/platform-cases.tsv`, `.github/ci/platform-exclusions.tsv`, `internal/gitops/gitops.go`, `internal/gitops/dirfold_test.go`, trunk changelog entries, and the incoming `.task-board` checkout files match refreshed `HEAD` with no candidate delta. The changelog contains both trunk entries and the Windows snapshot fix.

### Hosted evidence rechecked

The failed run 35855672743 (`windows-latest`) recorded the test at `2026-09-23T11:56:59.1577912Z`; its exact failure line is:

```text
snapshot_test.go:61: worker 11 = "", snapshot destination conflicts with immutable commit: authenticate destination: open C:\Users\RUNNER~1\AppData\Local\Temp\TestConcurrentGetAcceptsOneImmutablePublisher224870753\002\cache\internal\skill-a\1112d93d63e9641cdfc52d7f2f46896fde5afe19\snapshot: The process cannot access the file because it is being used by another process.; want "C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\TestConcurrentGetAcceptsOneImmutablePublisher224870753\\002\\cache\\internal\\skill-a\\1112d93d63e9641cdfc52d7f2f46896fde5afe19\\snapshot"
```

The hosted message is Windows `ERROR_SHARING_VIOLATION` (32), not an access-denied error. The five sampled prior Windows gate artifacts (35855550263, 35852658095, 35758851581, 35739653315, 35688466491) all contain a `pass` event for `TestConcurrentGetAcceptsOneImmutablePublisher`; with run 35855672743, the sample is 1 failure in 6 runs.

Post-fix hosted gate 35874011891 succeeded. Its `windows-latest` `go-test-served.json` has `pass` events for `TestConcurrentGetAcceptsOneImmutablePublisher`, `TestConcurrentGetRetriesInjectedWindowsSharingViolation`, `TestGetFailsClosedAfterPersistentWindowsSharingViolation`, and `TestGetDoesNotRetryWindowsAccessDenied`. The injected transient row itself runs five concurrent scenarios. The Windows runtime rows therefore have hosted execution evidence; the current macOS producer host can compile, but cannot execute, the Windows-only tests. The new hosted gate launched by this handoff is the next repeated post-fix run; it is not claimed green here.

### Refreshed-tree validation

All commands below were run directly as standalone processes; exit codes are the actual process results.

| Command | Exit | Result |
| --- | ---: | --- |
| `git apply --check /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/BUG-260923-11jgkt/BUG-260923-11jgkt_change-request_rev1.patch` | 0 | Patch applies cleanly to refreshed `HEAD` |
| `git apply /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board/.resources/BUG-260923-11jgkt/BUG-260923-11jgkt_change-request_rev1.patch` | 0 | Restored accepted candidate |
| `go test ./internal/snapshot` | 0 | Passed |
| `go test -count=50 -run '^TestConcurrentGetAcceptsOneImmutablePublisher$' ./internal/snapshot` | 0 | 50 local concurrency repetitions passed |
| `go test -race ./internal/snapshot` | 0 | Passed |
| `env GOOS=windows GOARCH=amd64 go test -c -o ./.11jgkt-evidence/snapshot-refresh.test.exe ./internal/snapshot` | 0 | Windows test binary compiled; not run on this macOS host |
| `go vet ./internal/snapshot` | 0 | Passed |
| `go build ./internal/snapshot` | 0 | Passed |
| `golangci-lint run --allow-serial-runners ./internal/snapshot` | 0 | 0 issues |
| `env GOOS=windows GOARCH=amd64 golangci-lint run --allow-serial-runners ./internal/snapshot` | 0 | 0 issues, including Windows-tagged files |
| `gofmt -l internal/snapshot/snapshot.go internal/snapshot/destination_sharing_violation_windows.go internal/snapshot/destination_sharing_violation_other.go internal/snapshot/snapshot_windows_test.go` | 0 | No files listed |
| `git diff --check` | 0 | Passed |
| `sh .github/ci/ledger-consistency.sh` | 2 | Usage refusal: this script requires an evidence-directory argument |
| `sh .github/ci/ledger-consistency.sh ./.11jgkt-evidence/rev2-ledger` | 0 | 245 rows consistent across linux, darwin, windows |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed |

No full hosted suite was run manually; campaign rules reserve the one landing gate for handoff. Existing post-fix hosted evidence above is from run 35874011891, and the handoff starts the next hosted run.

## Revision 3 (review response)

### Root cause and hosted evidence

I downloaded `test-evidence-windows-latest` for run 35855672743 and inspected `test/go-test-served.json`. The run was created at 2026-09-23T11:38:05Z; `TestConcurrentGetAcceptsOneImmutablePublisher` failed on `windows-latest` at 2026-09-23T11:56:59.172621Z:

```text
snapshot_test.go:61: worker 11 = "", snapshot destination conflicts with immutable commit: authenticate destination: open C:\Users\RUNNER~1\AppData\Local\Temp\TestConcurrentGetAcceptsOneImmutablePublisher224870753\002\cache\internal\skill-a\1112d93d63e9641cdfc52d7f2f46896fde5afe19\snapshot: The process cannot access the file because it is being used by another process.; want "C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\TestConcurrentGetAcceptsOneImmutablePublisher224870753\\002\\cache\\internal\\skill-a\\1112d93d63e9641cdfc52d7f2f46896fde5afe19\\snapshot"
```

This is Windows `ERROR_SHARING_VIOLATION` (32). In `internal/snapshot/snapshot.go`, `authenticateDestination` digested the existing destination and converted any read error to `ErrDestinationConflict`; a concurrent publisher can therefore cause a legitimate `Get` to be refused. The failure is a product defect, not just a test race. The five earlier hosted Windows runs listed in the prior results artifact (35855550263, 35852658095, 35758851581, 35739653315, 35688466491) passed this test; the sampled frequency remains 1 failure in 6 runs.

The production fix remains bounded: retry only errors recognized as Windows `ERROR_SHARING_VIOLATION`, three retries at 10 ms (four digest attempts maximum). Access denied, other errors, persistent sharing violations, and a readable destination whose digest differs still fail closed. Non-Windows builds classify no errors as sharing violations.

### Named refusal row and mutant

Renamed and strengthened the cached-destination test as `TestGetRejectsDifferentDestinationForSameCommit` in `internal/snapshot/snapshot_test.go`. It changes the committed file bytes in an existing cache, calls the public `Get` entry point, requires `ErrDestinationConflict`, and checks that the conflicting bytes remain unchanged. This is the genuine-different-destination refusal row for AC 3; the row runs on Windows as part of the hosted Go suite.

Narrowing mutant in disposable copy `/tmp/BUG-260923-11jgkt-mutant-r3`: changed the digest mismatch guard to `if destinationDigest != expectedDigest && false`. Command `go test ./internal/snapshot -run '^TestGetRejectsDifferentDestinationForSameCommit$'` exited 1 as expected; the row failed with `Get() error for different cached destination = <nil>, want ErrDestinationConflict`. The named row kills the weakened immutability guard. An initial disposable mutant written as `if false` also exited 1, but only because the compiler found unused digest variables; it was discarded and is not the mutation result.

### Hosted gate history

- Run 35874011891 succeeded. I independently downloaded its Windows test artifact and confirmed pass events for `TestConcurrentGetAcceptsOneImmutablePublisher`, `TestConcurrentGetRetriesInjectedWindowsSharingViolation`, `TestGetFailsClosedAfterPersistentWindowsSharingViolation`, and `TestGetDoesNotRetryWindowsAccessDenied`. The injected transient test performs five concurrent iterations; persistent sharing violation asserts the four-attempt bound; access denied asserts one attempt.
- Refreshed-candidate run 35901867517 had a Windows job failure, but all `internal/snapshot` tests and all three Windows-specific sharing rows passed in its artifact. The gate failed in the unrelated `internal/registry` test `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound` at 2026-09-23T18:49:41.8848092Z: `registry_test.go:734: skew 0s offset 1s: refused=false want true ([])`. Windows `go test` and the platform-case gate returned 1 and 0 respectively.
- The attached rev1 verdict artifact says ACCEPTED; the refresh instructions also say rev1 was accepted. The review-note wording that calls rev1 a rejection is inconsistent with that board evidence. This revision adds the explicit named refusal row and mutation proof requested by the review instructions.
- The revision 3 handoff runs the configured hosted landing suite once. Its current result is recorded by the task-board runtime validation log attached during handoff; this artifact makes no claim about that run before handoff returns.

### Validation after the named-row change

Every command below was run as a standalone process on the refreshed candidate:

| Command | Exit | Result |
| --- | ---: | --- |
| `gofmt -w internal/snapshot/snapshot_test.go` | 0 | Formatted |
| `go test ./internal/snapshot` | 0 | Passed |
| `go test -run '^TestGetRejectsDifferentDestinationForSameCommit$' ./internal/snapshot` | 0 | Named refusal row passed |
| `go test -count=50 -run '^TestConcurrentGetAcceptsOneImmutablePublisher$' ./internal/snapshot` | 0 | 50 local repetitions passed |
| `env GOOS=windows GOARCH=amd64 go test -c -o /tmp/BUG-260923-11jgkt-snapshot.test.exe ./internal/snapshot` | 0 | Windows test binary compiled; Windows execution is provided by hosted evidence |
| `go vet ./internal/snapshot` | 0 | Passed |
| `go build ./internal/snapshot` | 0 | Passed |
| `golangci-lint run --allow-serial-runners ./internal/snapshot` | 0 | 0 issues |
| `git diff --check` | 0 | Passed |
| Mutant `go test ./internal/snapshot -run '^TestGetRejectsDifferentDestinationForSameCommit$'` in disposable copy | 1 | Expected failure: weakened digest refusal returned nil |


## Revision 4 (carry-forward republish)

Trunk moved to `1511b345`; the orchestrator converged the Story worktree. No
content edits were made in this revision — verification only, per
`11jgkt-carry-4.md`.

### Per-path verification against `BUG-260923-11jgkt_change-request_rev1.patch`

Applied `BUG-260923-11jgkt_change-request_rev3.patch` (rev1 plus the
review-requested `snapshot_test.go` strengthening; all other paths identical
to rev1) to a clean `git archive HEAD` copy in `/tmp/scratch` and
byte-compared each file with the worktree (`cmp`):

| Path | Rev1 status | Result |
| --- | --- | --- |
| `CHANGELOG.md` (intersecting: trunk touched) | merged | IDENTICAL; trunk side retained (E4 entry and siblings), 5-line fix entry added, no markers, nothing dropped/duplicated |
| `internal/snapshot/snapshot.go` | untouched by trunk | IDENTICAL to rev1 |
| `internal/snapshot/destination_sharing_violation_other.go` (new) | untouched by trunk | IDENTICAL to rev1 |
| `internal/snapshot/destination_sharing_violation_windows.go` (new) | untouched by trunk | IDENTICAL to rev1 |
| `internal/snapshot/snapshot_windows_test.go` (new) | untouched by trunk | IDENTICAL to rev1 |
| `internal/snapshot/snapshot_test.go` (extra vs rev1) | carried strengthening | IDENTICAL to rev3: `TestGetRejectsDifferentDestinationForSameCommit` named refusal row answering the review brief; no other change |

`grep` for conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`) across
`CHANGELOG.md` and `internal/snapshot/` found none.

### Focused bounded run (standalone processes, real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/snapshot/...` | 0 | ok 4.330s |
| `go vet ./internal/snapshot/...` | 0 | clean |
| `gofmt -l internal/snapshot/` | 0 | no output (clean) |
| `go build ./...` | 0 | clean |
| `GOOS=windows go vet ./internal/snapshot/...` | 0 | Windows-only files compile |

All DoD items remain checked on the strength of revisions 1–3 evidence plus
this revision's carry-forward verification.

## Revision 5 (carry-forward republish)

Trunk moved to `948ae7c9`; the orchestrator converged the Story worktree.
No content edits in this revision except the orchestrator-mandated CHANGELOG
revert — verification only, per `11jgkt-carry-5.md`.

### Per-path verification against `BUG-260923-11jgkt_change-request_rev4.patch`

Applied the rev4 patch's `internal/snapshot/*` hunks to a clean
`git archive HEAD` copy in `/tmp/rev5-base` (`git apply --check` and
`git apply` both exited 0) and byte-compared each file with the worktree
(`cmp`):

| Path | Trunk touched? | Result |
| --- | --- | --- |
| `internal/snapshot/snapshot.go` | no | IDENTICAL to rev4 |
| `internal/snapshot/snapshot_test.go` | no | IDENTICAL to rev4 |
| `internal/snapshot/destination_sharing_violation_other.go` (new) | no | IDENTICAL to rev4 |
| `internal/snapshot/destination_sharing_violation_windows.go` (new) | no | IDENTICAL to rev4 |
| `internal/snapshot/snapshot_windows_test.go` (new) | no | IDENTICAL to rev4 |
| `CHANGELOG.md` | yes (intersecting) | both sides present after converge: trunk entries (E4 and siblings at new position) retained, 5-line fix hunk added, no conflict markers — then reverted per CHANGELOG policy below |

`grep` for conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`) across
`CHANGELOG.md` and `internal/snapshot/` found none. No stray root
`TASK-*`/`BUG-*`.md files and no `test/` or `ledger/` paths exist in the
worktree (`git status --short` shows only the two tracked snapshot
modifications plus the three accepted new files).

### CHANGELOG policy (orchestrator, 2026-09-24)

This task's `CHANGELOG.md` hunk was reverted entirely: `git checkout --
CHANGELOG.md`, then `cmp CHANGELOG.md` against `git show HEAD:CHANGELOG.md`
confirmed the file equals trunk byte for byte. The entry text is preserved
verbatim below for the release-prep leaf.

## CHANGELOG entry (for release prep)

```text
- Concurrent Windows snapshot readers now retry destination authentication
  three times at 10 ms intervals when the immutable destination open returns
  `ERROR_SHARING_VIOLATION` (32) during publication. Access denied, other read
  errors, persistent sharing violations, and a destination with different
  contents still fail closed as an immutable-commit conflict.
```

### Focused bounded run (standalone processes, real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/snapshot/...` | 0 | ok 10.094s (uncached; an earlier cached `ok` was discarded and rerun with `-count=1`) |
| `go vet ./internal/snapshot/...` | 0 | clean |
| `gofmt -l internal/snapshot/` | 0 | no output (clean) |
| `git diff --check` | 0 | clean |

Windows-only rows still execute only on hosted Windows runners; prior hosted
evidence (runs 35874011891 and 35901867517, recorded in revisions 2–3) stands.
All 12 board DoD items were already checked on the strength of revisions 1–4
evidence; no unchecked item remained for this revision.

## Revision 5 — rerun verification (this developer run)

Trunk still `948ae7c9`. No content edits in this run — verification only, per
`11jgkt-carry-5.md`. This rerun independently reproduced the Revision 5
verification above with fresh commands.

### Per-path verification against `BUG-260923-11jgkt_change-request_rev4.patch`

`git apply --check` of the full rev4 patch against a clean `git archive HEAD`
copy exited 0 (no conflicts: trunk did not touch the snapshot paths).
After `git apply` in that copy, `cmp` against the worktree reported:

| Path | Result |
| --- | --- |
| `internal/snapshot/snapshot.go` | IDENTICAL to rev4 |
| `internal/snapshot/snapshot_test.go` | IDENTICAL to rev4 |
| `internal/snapshot/destination_sharing_violation_other.go` (new) | IDENTICAL to rev4 |
| `internal/snapshot/destination_sharing_violation_windows.go` (new) | IDENTICAL to rev4 |
| `internal/snapshot/snapshot_windows_test.go` (new) | IDENTICAL to rev4 |
| `CHANGELOG.md` | equals trunk `HEAD` byte for byte (`git diff --quiet -- CHANGELOG.md` clean); entry text preserved in the "CHANGELOG entry (for release prep)" section above per orchestrator policy |

`grep` for conflict markers across `CHANGELOG.md` and `internal/snapshot/`
found none. No stray root `TASK-*`/`BUG-*`.md, `test/`, or `ledger/` paths.
`git status --short` shows only the two tracked snapshot modifications plus
the three accepted new files. Nothing else changed.

### Focused bounded run (standalone processes, real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/snapshot/...` | 0 | ok 2.823s |
| `go vet ./internal/snapshot/...` | 0 | clean |
| `GOOS=windows GOARCH=amd64 go vet ./internal/snapshot/...` | 0 | Windows-only files compile |
| `gofmt -l internal/snapshot/` | 0 | no output (clean) |
| `git diff --check` | 0 | clean |
| `go build ./internal/snapshot/...` | 0 | clean |

Windows-only rows execute only on hosted Windows runners; prior hosted
evidence (runs 35874011891 and 35901867517, recorded in revisions 2–3) stands.
All 12 board DoD items were already checked; no unchecked item remained for
this rerun.

## Revision 5 — hosted anomaly note (run 35967572915, same candidate)

The story-level remote gate pushed by the earlier Revision 5 run
(`gate/STORY-260923-laeycm/260924-070322`, run 35967572915, 2026-09-24
~07:03–07:39 UTC) finished `failure` on one job: Race (ubuntu-latest),
`go test exit=1`. Attribution from `race-evidence-ubuntu-latest`
`race/go-test-served.json`:

- `internal/snapshot`: package-level `pass` under `-race` on ubuntu — this
  task's scope is green on hosted Race-ubuntu.
- The failing package is `internal/install`: `FAIL ... 1800.288s`, a go-test
  30-minute timeout (goroutine dump shows
  `TestScriptAuditLabelsAtGlobalInstallEntry` blocked in `os.CreateTemp` via
  `transaction.saveJournal`). Unrelated to snapshot publication; outside this
  task's scope, recorded here as an anomaly for the Story lane, not reworked
  in this revision.
- `Test (windows-latest)` job succeeded in the same run. From
  `test-evidence-windows-latest` `test/go-test-served.json`, all five
  snapshot rows report `pass` on hosted Windows:
  `TestConcurrentGetAcceptsOneImmutablePublisher`,
  `TestConcurrentGetRetriesInjectedWindowsSharingViolation`,
  `TestGetFailsClosedAfterPersistentWindowsSharingViolation`,
  `TestGetDoesNotRetryWindowsAccessDenied`,
  `TestGetRejectsDifferentDestinationForSameCommit`.
- Local `go test -race -count=1 ./internal/snapshot/...` in this rerun:
  exit 0, ok 26.051s.

## Revision 7 (carry-forward republish) — this developer run

Trunk now `a48f584c`. Content not in question (rev6 ACCEPTED); verification
only per `11jgkt-carry-7.md`, plus the CHANGELOG policy. No product-code edits
in this run.

### Per-path verification against `BUG-260923-11jgkt_change-request_rev6.patch`

rev6 lists 5 paths (no CHANGELOG hunk). Trunk did not touch any of them:
`git log --oneline -8 -- internal/snapshot/` shows the newest touch is
`6645b9b1` (predates rev6), so there are no intersecting paths and no merge
conflict to resolve.

| Path | Trunk touched? | Result |
| --- | --- | --- |
| `internal/snapshot/snapshot.go` | no | worktree `git diff HEAD` hunk-for-hunk identical to rev6 (index `a1395144..e1aa0b26`) |
| `internal/snapshot/snapshot_test.go` | no | worktree `git diff HEAD` hunk-for-hunk identical to rev6 (index `f8ff1428..0df0397a`) |
| `internal/snapshot/destination_sharing_violation_other.go` (new) | no | `cmp` IDENTICAL to rev6 patch content (extracted via `git apply` in a clean `/tmp` copy) |
| `internal/snapshot/destination_sharing_violation_windows.go` (new) | no | `cmp` IDENTICAL to rev6 patch content |
| `internal/snapshot/snapshot_windows_test.go` (new) | no | `cmp` IDENTICAL to rev6 patch content |

`grep` for conflict markers (`<<<<<<<`) across `CHANGELOG.md` and
`internal/snapshot/` printed no matches. `git status` shows only the two
tracked snapshot modifications plus the three accepted new files. No stray
root `TASK-*`/`BUG-*`.md files; no `test/` or `ledger/` paths.

### CHANGELOG policy (orchestrator, 2026-09-24)

Nothing to revert: this task has no `CHANGELOG.md` hunk in the worktree.
`git diff --quiet -- CHANGELOG.md` exits 0 and `cmp CHANGELOG.md` against
`git show HEAD:CHANGELOG.md` confirms the file equals trunk byte for byte.
The verbatim entry text remains in the "CHANGELOG entry (for release prep)"
section above for the release-prep leaf.

### Focused bounded run (standalone process, real exit code)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/snapshot/... -count=3` | 0 | ok 12.940s |

Windows-only rows still execute only on hosted Windows runners; prior hosted
evidence (rev6 validation log: Test windows-latest success, exit 0) stands.
All 12 board DoD items were already checked on the strength of revisions 1–6
evidence; no unchecked item remained for this revision.

## Revision 7 (carry-forward republish) — rerun verification (this developer run)

Trunk `a48f584c`. Content not in question (rev6 ACCEPTED); verification only per `11jgkt-carry-7.md`, plus the CHANGELOG policy. No product-code edits in this run. This rerun independently reproduces the Revision 7 verification above with fresh commands in this developer turn.

### Per-path verification against `BUG-260923-11jgkt_change-request_rev6.patch`

rev6 lists 5 paths (no CHANGELOG hunk). `git log --oneline -8 -- internal/snapshot/` newest touch is `6645b9b1` (predates rev6): trunk touched none of them, so no intersecting paths and no conflict to resolve. Applied the rev6 patch to a clean `git archive HEAD` copy in `/tmp/rev7-base`: `git apply --check` exit 0, `git apply` exit 0. `cmp` of each file against the worktree:

| Path | Trunk touched? | Result |
| --- | --- | --- |
| `internal/snapshot/snapshot.go` | no | `cmp` exit 0 — IDENTICAL to rev6 |
| `internal/snapshot/snapshot_test.go` | no | `cmp` exit 0 — IDENTICAL to rev6 |
| `internal/snapshot/destination_sharing_violation_other.go` (new) | no | `cmp` exit 0 — IDENTICAL to rev6 |
| `internal/snapshot/destination_sharing_violation_windows.go` (new) | no | `cmp` exit 0 — IDENTICAL to rev6 |
| `internal/snapshot/snapshot_windows_test.go` (new) | no | `cmp` exit 0 — IDENTICAL to rev6 |
| `CHANGELOG.md` | n/a (no rev6 hunk) | equals trunk: `git diff --quiet -- CHANGELOG.md` exit 0, `cmp CHANGELOG.md` against `git show HEAD:CHANGELOG.md` content exit 0 |

`grep -rn "<<<<<<<" CHANGELOG.md internal/snapshot/` exit 1 (no conflict markers). `git status --short` shows only the two tracked snapshot modifications plus the three accepted new files. No stray root `TASK-*`/`BUG-*`.md files; no `test/` or `ledger/` paths. Nothing else changed.

### CHANGELOG policy (orchestrator, 2026-09-24)

Nothing to revert: this task has no `CHANGELOG.md` hunk in the worktree or in rev6. The file equals trunk byte for byte (see exits above). The verbatim entry text remains in the "CHANGELOG entry (for release prep)" section above for the release-prep leaf:

```text
- Concurrent Windows snapshot readers now retry destination authentication
  three times at 10 ms intervals when the immutable destination open returns
  `ERROR_SHARING_VIOLATION` (32) during publication. Access denied, other read
  errors, persistent sharing violations, and a destination with different
  contents still fail closed as an immutable-commit conflict.
```

### Focused bounded run (standalone process, real exit code)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/snapshot/... -count=3` | 0 | ok 17.075s |

Windows-only rows execute only on hosted Windows runners; prior hosted evidence (rev6 validation log: Test windows-latest success, exit 0; runs 35874011891 and 35901867517 in revisions 2–3) stands. All 12 board DoD items were already checked on the strength of revisions 1–6 evidence; no unchecked item remained for this rerun.
