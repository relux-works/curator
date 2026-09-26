# TASK-260925-h4syhu Results

> Historical note: the earlier sections record accepted Revision 3. Revision 4 below is current and supersedes the earlier ERROR_PATH_NOT_FOUND analysis and Windows fixture description.

## Scope

Changed only the task-owned test file `internal/install/draftsources_test.go` and platform rows in `.github/ci/platform-cases.tsv`. The production implementation and all other Story paths remain byte-identical to TASK-260907-2as5sx rev9. No CHANGELOG file was edited.

The test calls `lockedNetworkRepository` and gives each checkout outcome its own subtest:

- `absent_checkout_permits_fallback`: a missing first checkout falls through to the later usable checkout.
- `lstat_failure_stops_fallback`: on POSIX, a regular `blocked` file makes `blocked/child` fail Lstat with ENOTDIR. On Windows, the real input is `invalid?name`, which returns ERROR_INVALID_NAME. Both cases assert the typed `manager_state_unreadable` error, its underlying Lstat `os.PathError`, and no fallback.
- `present_but_unusable_stops_fallback`: a regular file at the checkout path is present, but not a Git repository; the test asserts `stateread.UnusableError` and no fallback.

The platform ledger requires all three subtests on Linux, Darwin, and Windows. No skip row was needed.

Windows ERROR_PATH_NOT_FOUND beneath a known regular-file parent is **not semantically correct absence**: the parent exists but blocks traversal. Go's Windows `syscall.Errno.Is(os.ErrNotExist)` maps ERROR_PATH_NOT_FOUND to not-exist, and `stateread.Lstat` consequently returns KindAbsent for that error; `lockedNetworkRepository` can then fall back. This remains a documented bound of the shared `stateread.Lstat` seam in this leaf. The Windows test uses ERROR_INVALID_NAME to exercise a distinct, genuine Lstat read failure. The Windows test itself was not executed locally; the Windows vet and cross-platform ledger checks are listed below.

## M1 narrowing proof (Darwin)

Ran the focused test before changing its rows: exit 0. Inserted M1 (`continue` as the first statement of the Lstat error branch) and reran the rev9 test: exit 0, so the old coverage survived the mutant. After adding the dedicated Lstat-failure subtest, the same mutant made the test fail: exit 1. The failure was `lstat_failure_stops_fallback` receiving nil error because execution continued to the later checkout. The temporary mutant was removed.

## Validation

All commands ran directly in zsh, as standalone processes.

- `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` — exit 0 on Darwin; all three subtests pass.
- M1 survivor before the new row — exit 0.
- M1 after the new row — exit 1 as expected; the failure identifies the Lstat-failure subtest.
- `GOOS=windows go vet ./internal/install` — exit 0.
- `golangci-lint run ./internal/install` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/task-h4syhu-ledger-final.TkpMMc` — exit 0; 390 rows checked across Linux, Darwin, and Windows. The three new rows are present and compile for every listed GOOS.
- `gofmt -l internal/install/draftsources_test.go` — exit 0, no output.
- `git diff --check -- internal/install/draftsources_test.go .github/ci/platform-cases.tsv` — exit 0.
- Compared Git blob hashes against rev9 for all 72 paths outside this leaf's two owned paths — exit 0, all byte-identical. `git apply --reverse --check` excluding those same two paths also exited 0.
- `git diff --quiet HEAD -- CHANGELOG.md` — exit 0, no CHANGELOG change.

The actual hosted test lanes, including Windows runtime execution, were not run in this producer worktree. They remain for the parent-owned hosted PR checks; `GOOS=windows go vet` is static validation, not a Windows test run.

## CHANGELOG entry (for release prep)

- A same-source reinstall of a git root honours `--use` and `--takeover`
  exactly as a first install does. The reinstall re-resolves as an update
  and then runs the install row's activation — activating when the machine
  has no current or the operator passes `--use`, with `--takeover`
  covering the unmanaged files the install would write — instead of
  delegating to the update and reporting success for no work. Previously
  the §9.5 stop-and-retry `profile install <git-url> --use --takeover`
  accepted both flags, did nothing, and exited 0 saying `updated profile`;
  a retry that still meets unmanaged state without `--takeover` now fails
  loudly with the stop diagnostic (Protocol environments §9.1, §9.5).
  Path-root reinstall behaviour is unchanged.

## Revision 1b — refresh onto ab34556e

### Trunk refresh and candidate identity

The candidate was refreshed onto trunk `ab34556ebf17ab95532a3f795aa677b2234ecd8a`. The first prescribed `git diff a48f584c ab34556e ... | git apply --3way` attempt returned exit 1 because the checkpoint index was already at `d9cb8465`, not the patch base; Git rolled the patch application back. Retrying with a temporary index also returned exit 1 because the candidate worktree and base index differed. I combined the 61 trunk-changed paths with content-level three-way merges instead: 59 merged cleanly, and the two conflicts were import blocks in `internal/envprofile/envprofile.go` and `internal/install/draftsources_test.go`; both imports were unioned, then the now-unused `snapshot` import was removed. `task-board worktree refresh-candidate TASK-260925-h4syhu` exited 0 and advanced the Story checkpoint to `6f94e1a93dbcb7860907bd099e9e9a321c77d0f4`, whose parent is `ab34556e`.

`CHANGELOG.md` is byte-identical to trunk `ab34556e` (`git diff --quiet ab34556e -- CHANGELOG.md`, exit 0). The checked-out `.task-board` snapshot is clean against the refreshed checkpoint; authoritative board writes continued through `task-board` only.

For the rev9 identity check, I reconstructed the rev9 tree from `a48f584c` plus `TASK-260907-2as5sx_change-request_rev9.patch`. Of the 74 rev9 paths, after excluding the 61 paths touched by the new trunk overlay, the two leaf-owned paths, the explicitly required AST guard update, and `CHANGELOG.md`, all 60 remaining paths matched byte-for-byte and mode-for-mode (0 mismatches). The earlier pre-refresh comparison recorded in the prior results also found all 72 non-leaf-owned paths identical to rev9.

### Read-failure rows and Windows bound

`TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` still drives `lockedNetworkRepository` and keeps each outcome in a separate subtest:

- `absent_checkout_permits_fallback`: the missing checkout reaches the later valid checkout.
- `lstat_failure_stops_fallback`: POSIX uses a regular `blocked` file and `blocked/child` (`ENOTDIR`); Windows uses the invalid component `invalid?name` (`ERROR_INVALID_NAME`). Both assert typed `manager_state_unreadable`, the underlying `lstat` path error, and no fallback.
- `present_but_unusable_stops_fallback`: a regular file at the checkout path is present but not a Git repository; the error is a `stateread.UnusableError` and no fallback occurs.

All three rows are listed for Linux, Darwin, and Windows in `.github/ci/platform-cases.tsv`. The Windows Lstat failure row is a real invalid-name input; Windows runtime execution was not available locally.

`ERROR_PATH_NOT_FOUND` under a known regular-file parent is not semantically absence: path traversal is blocked by an existing file. Go's Windows `os.IsNotExist` mapping treats that code as not-exist, so `stateread.Lstat` currently classifies this specific shape as absent. This remains an explicit bound of the shared seam in this leaf; the Windows row uses `ERROR_INVALID_NAME` to exercise a real Lstat read failure instead.

### M1 narrowing proof (Darwin)

I inserted M1 (`continue` as the first statement of the `stateread.Lstat` error branch) and ran the pre-row shapes and the new row separately:

- Absent plus present-unusable rows under M1 — exit 0 (the old coverage survives).
- `lstat_failure_stops_fallback` under M1 — exit 1; it received nil instead of a typed unreadable error after falling through to the later checkout (the new row kills M1).
- Restored `internal/install/draftsources.go` from its saved pre-mutant bytes; `cmp` — exit 0.

After restoration, the three-subtest test passed on Darwin.

### AST scan of refreshed trunk

The first post-refresh guard run returned exit 1 and identified two new absence-sensitive readers plus one stale allowlist key: `internal/gitops/gitops.go:FetchCommitFromURLIsolated`, `internal/manifest/manifest.go:Load`, and the old `LoadWithOptions` key. The two sites preserve distinct semantics: a missing private locked-replay checkout is initialized while other Lstat errors stop; a missing project manifest means the project is uninitialized while other read errors propagate. I updated the manifest allowlist key, added the specific private locked-replay rationale, and made the guard log its covered/total percentage.

Final guard result: `TestManagerOwnedAbsenceReadsAreGuarded` and `TestManagerReadScannerFindsAliasedNotExistCollapse` pass with 178 seam-guarded + 119 allowlisted = 297/297 relevant readers (100.0%), across 403 production files.

### Validation on the refreshed tree

All commands below ran directly as standalone processes in zsh. Exit codes are the actual process results.

- `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` — exit 0.
- Focused locked-replay/install tests (`LockedNetworkRepository`, missing/stale lock, pinned bytes, local snapshot replay, Git subtree/tamper and portable replay URL) — exit 0, 9.403s.
- Fresh-machine replay tests (`TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade`, declared Git without bindings, locked commit replay, hash/path mismatch and unavailable endpoint cases) — exit 0, 18.516s.
- The requested combined command `go test ./internal/install -run 'LockedNetworkRepository|Draft' -count=1` reached Go's 10m0s package timeout and returned exit 1 while running `TestDraftLocalEnforcedCommandRefused` (34s into that test at timeout). I reran the `Draft` selection in bounded groups; every group passed:
  - Failure-class, audit, build and evidence tests — exit 0, 322.886s.
  - Git/install, selector/tag, repair/retarget and runtime-key tests — exit 0, 126.978s.
  - Local, marker and transport tests — exit 0, 89.657s.
  - Plan, SSH and manager executable tests — exit 0, 0.478s.
  - Fresh/replay cases were run separately above and passed.
- `go test ./internal/envprofile -run 'TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse' -count=1 -v` — exit 0; 297/297 coverage shown above.
- Native `go vet` over the 20 changed Go packages, in four sequential package groups — exit 0 for each group.
- `GOOS=windows go vet ./internal/install` — exit 0. This is static validation; it does not claim Windows runtime execution.
- `golangci-lint run --concurrency 2` over the 20 changed Go packages — exit 0, 0 issues.
- `go build` over the 20 changed Go packages — exit 0.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-refresh-20260926` — exit 0; 390 rows checked across Linux, Darwin, and Windows.
- `gofmt -l` on the three manually merged Go files — exit 0, no output. `git diff --check` — exit 0.

The hosted runtime lanes, including the Windows execution of the invalid-name fixture, remain for the parent-owned hosted gate after handoff; no hosted result is claimed here.

## Revision 1c — Windows PathError operation assertion

The first refreshed hosted gate (run 36191751614) passed the Linux and macOS test lanes and failed the Windows test lane only in `lstat_failure_stops_fallback`. I inspected the uploaded Windows `go-test-served.json` artifact under `/tmp` (not in the Story worktree). The invalid-name fixture reached the intended typed `manager_state_unreadable` error with the original Windows path error, but the test incorrectly required `os.PathError.Op == "lstat"`. Windows reports the underlying `CreateFile` operation used by `os.Lstat`; this was an assertion mismatch, not an absent-path classification or skipped row.

Updated only the owned test assertion: POSIX expects `lstat`, Windows expects `CreateFile`, and both require the exact attempted checkout path. The Windows row remains a real `invalid?name` input producing `ERROR_INVALID_NAME`; it is not ledgered as a skip. The prior run's hosted output established that the input yields a typed unreadable error and stopped fallback. A fresh hosted gate will run through this handoff.

### Revision 1c verification (Darwin unless stated)

- M1 inserted as the first statement of the production Lstat-error branch; absent and present-unusable subtests — exit 0 (mutant survives those rows).
- Same M1 mutant with only `lstat_failure_stops_fallback` — exit 1 as expected; it received nil after falling through to the later valid checkout (mutant killed).
- Removed M1; all three checkout subtests with `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` — exit 0.
- Refreshed AST guard command over `internal/envprofile` — exit 0; 297/297 relevant readers covered (178 seam-guarded + 119 reasoned allowlist entries; 100.0%, 403 production files scanned).
- `go vet ./internal/install` — exit 0.
- `GOOS=windows go vet ./internal/install` — exit 0 (static validation; not a Windows runtime rerun).
- `go build ./internal/install` — exit 0.
- `golangci-lint run ./internal/install` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-handoff` — exit 0; 390 rows checked across Linux, Darwin, and Windows, including all three checkout subtests on every OS.
- `gofmt -l internal/install/draftsources_test.go` — exit 0, no output; `git diff --check -- internal/install/draftsources_test.go .github/ci/platform-cases.tsv` — exit 0.
- `git diff --quiet ab34556e -- CHANGELOG.md` — exit 0; no CHANGELOG delta against the refreshed trunk base. The release entry text remains in the section above.

The broad `go test ./internal/install -run 'LockedNetworkRepository|Draft' -count=1` command was previously run after the refresh and reached the 10-minute package timeout (exit 1). Its Draft coverage was rerun in bounded test groups with exit 0 as recorded above; the dedicated checkout test was rerun on the current revision with exit 0. This revision changes only the platform-specific PathError assertion in the owned test file. Hosted Windows runtime verification is pending the fresh handoff gate.

## Revision 3 — refresh onto 9f0da708

### Trunk refresh and preserved Story delta

The candidate was refreshed onto trunk 9f0da708350c15e78b7c900022629aa8e42e5e12.
task-board worktree refresh-candidate TASK-260925-h4syhu exited 0 and advanced
the candidate branch to b2084346537c01d9e0f1f7e43891f25504e88635.

The prescribed overlay command was run directly under set -o pipefail:

- git diff ab34556e 9f0da708 -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way — exit 1.
  Git reported index mismatches on .github/ci/platform-cases.tsv,
  internal/install/draftsources.go, and internal/scopes/gc_conformance_test.go.
  A content-level three-way merge retained both sides: the other 66 changed paths
  were byte-identical to trunk; the platform ledger contains both sets of rows;
  draftsources.go retains the full 11burj replay implementation plus this leaf's
  checkout classification; and the GC conformance file retains the trunk corpus
  binding plus the mustLoadConsumers assertions.

In trunk's new draftFrozenInput code, the direct os.Lstat / os.IsNotExist
collapse at internal/install/draftsources.go:draftFrozenInput was migrated to
stateread.Lstat: only KindAbsent continues; every other read error propagates
as a typed read failure. lockedNetworkRepository also remains on the same
stateread.Lstat seam. No allowlist entry was added.

The refreshed worktree comparison against 9f0da708 lists only Story candidate
paths. git diff --quiet 9f0da708 -- CHANGELOG.md exited 0, and the index had
no staged changes. The checked-out .task-board snapshot is clean.

Revision 2's accepted review verdict records the byte-for-byte comparison against
TASK-260907-2as5sx revision 9 for every non-leaf Story path. For this refresh,
59 other Story paths were compared byte-for-byte and mode-for-mode with the
revision 2 safety snapshot (0 mismatches); 10 untracked files absent from that
snapshot were untouched because they are outside the trunk overlay. Three
overlapping paths were handled explicitly as listed above. This refresh did not
change any other Story path.

### M1 narrowing proof (Darwin)

- Current three-row test before mutation:
  go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1
  — exit 0.
- Inserted M1 (continue as the first statement of the stateread.Lstat error
  branch). The absent and present-but-unusable control rows survive it:
  go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/(absent_checkout_permits_fallback|present_but_unusable_stops_fallback)$' -count=1
  — exit 0.
- The dedicated read-failure row kills M1:
  go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/lstat_failure_stops_fallback$' -count=1
  — exit 1 as expected. It received nil after the mutant continued to the later
  checkout instead of returning manager_state_unreadable.
- Restored the production file. The full three-row command above passed again
  — exit 0.

### Rows and platform bound

The three separate subtests remain registered for Linux, Darwin, and Windows.
The Windows input is a real invalid-name Lstat failure (invalid?name,
ERROR_INVALID_NAME); the assertion expects the Windows CreateFile operation
and exact attempted path. POSIX uses a regular-file parent and child path
(ENOTDIR). The present-but-unusable row checks stateread.UnusableError and
no fallback.

ERROR_PATH_NOT_FOUND beneath a known regular-file parent is not semantically
correct absence: traversal is blocked by an existing file. The Go Windows
os.IsNotExist mapping currently lets stateread.Lstat classify that shape as
absent; this remains an explicit bound of the shared seam. The Windows failure
row uses ERROR_INVALID_NAME, so it proves a real non-absence Lstat failure
without claiming the seam resolves that separate mapping bound.

### Revision 3 validation

All commands below ran directly as standalone processes on Darwin unless noted.
Exit codes are the real process results.

- go test ./internal/install -run 'TestLockedNetworkRepository|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftInstallUsesPinnedBytes|TestDraftInstallGitPinnedSubtree|TestDraftInstallGitTamperRefused|TestDraftReplayFileURLUsesPortableGitConfigSyntax' -count=1 — exit 0.
- go test ./internal/install -run 'TestDraftFreshMachineReplays|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock' -count=1 — exit 0.
- go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1 — exit 0 (83.882s); covers transitive dependency replay through the production CLI.
- go test ./internal/envprofile -run 'TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse' -count=1 -v — exit 0; 297/297 relevant readers covered (100.0%): 178 through the seam, 119 allowlisted with reasons; 406 production files scanned.
- go vet ./internal/install — exit 0.
- GOOS=windows go vet ./internal/install — exit 0 (static validation only).
- go build ./internal/install — exit 0.
- golangci-lint run ./internal/install — exit 0, 0 issues.
- bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-refresh-9f0da708-20260926 — exit 0; 391 rows checked across Linux, Darwin, and Windows.
- gofmt -l internal/install/draftsources.go internal/install/draftsources_test.go internal/scopes/gc_conformance_test.go — exit 0, no output.
- git diff --check — exit 0.
- git diff --quiet 9f0da708 -- CHANGELOG.md — exit 0; no CHANGELOG delta.
- git diff --cached --quiet — exit 0; nothing staged.

The local worktree does not claim Windows runtime execution. The configured
hosted validation is run once by the developer handoff; its generated
revision 3 validation log is the evidence for the Linux, macOS, and Windows
runtime lanes.

The existing CHANGELOG entry text above is retained in this results resource.
No CHANGELOG file edit is included.


## Revision 4 — refresh onto trunk incl. cww1ov

### Refreshed candidate and conflict resolutions

Refreshed the accepted Revision 3 candidate onto trunk `3bdcfe072ce6de591b4028261757c5db6ea7f217`. The task-board checkpoint now points at `827a30a2b4a53149c2e8cc5cbf5389d0cb822119`, whose parent is `3bdcfe07`; the working candidate tree including this revision's uncommitted changes is `d7d2b7f803e2bd0d8674220f2712bdb39daffda2`.

The prescribed overlay first returned exit 1 because the existing worktree/index did not match the patch base. Applying through a temporary index exposed five real overlaps, all resolved in the refreshed candidate:

- `cmd/curator/profile.go`: retained trunk's generic error reporting and profile-use output while preserving the candidate's update wording.
- `internal/envregistry/envregistry.go`: retained trunk's 0017 credential modes and Pi target, then restored the candidate's `DiagBackupRecordUnreadable` diagnostic.
- `internal/install/draftsources.go`: kept one `stateread` seam and trunk's unknown-metadata guard; both frozen-input and locked-network readers use it.
- `internal/runtimestore/enforced.go`: kept one shared seam and the unknown-state refusal.
- `internal/runtimestore/enforced_test.go`: kept trunk's portable regular-file-parent case in place of the old mode-bit fixture.

The reader audit also exposed new manager-owned absence collapses from the trunk overlay. Migrated the 15 affected readers in `internal/envprofile/managed.go` and `internal/envprofile/migrate.go` to `stateread` (credential config, managed links, migration inventory, roots, temporary links, and recovery records). Removed five stale audit allowlist entries whose readers now use the seam. The final audit is **354/354 (100.0%)**, with **240 seam-guarded**, **114 reviewed allowlisted**, across **408 production files**.

The platform ledger contained two stale `internal/stateread` rows for test names removed by the trunk seam consolidation. Removed those obsolete rows; the surviving `TestReadsDistinguishAbsentFromBlockedParent` covers file, directory, stat, and lstat reads. The final ledger consistency run checks **449 rows across Linux, Darwin, and Windows**.

### Locked repository rows and Windows semantics

The production `lockedNetworkRepository` Lstat error branch returns typed `manager_state_unreadable` and stops fallback. The production seam's `missingPath` checks ancestors after an `os.IsNotExist` result: for Windows `ERROR_PATH_NOT_FOUND` (3) on `blocked/child`, `blocked` is an existing regular file, so it is **not semantic absence** and must remain unreadable. Counting it absent would permit an unsafe fallback.

The three production-entry subtests are separate and registered on all three OS lanes:

- `absent_checkout_permits_fallback` — absent checkout reaches the later valid checkout.
- `lstat_failure_stops_fallback` — regular-file parent plus child path; POSIX `ENOTDIR`, Windows `ERROR_PATH_NOT_FOUND`, typed unreadable, no fallback.
- `present_but_unusable_stops_fallback` — present regular file returns `stateread.UnusableError`, no fallback.

The Windows case is a real failure input, not a skip row. Local Windows runtime was unavailable; `GOOS=windows go vet` and the ledger check are static evidence. The handoff's hosted Windows lane is the runtime check.

### Rev9 path identity

Reconstructed the accepted 2as5sx rev9 tree from base `a48f584c` and its rev9 patch, then compared the refreshed candidate against it after excluding the 160-path trunk overlay and the six explicit leaf paths (`.github/ci/platform-cases.tsv`, the two `internal/install/draftsources` files, `internal/envprofile/state_read_guard_test.go`, `managed.go`, and `migrate.go`). Result: **49 unaffected rev9 paths compared, 0 byte/mode mismatches; 68 candidate-delta paths, 0 unexplained paths**. `git diff --quiet origin/main -- CHANGELOG.md` exited 0: there is no CHANGELOG delta against the refreshed trunk, and no CHANGELOG file was edited for this leaf.

### Revision 4 verification

All listed commands ran directly as standalone processes. Exit codes are the actual process results.

- `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` before M1 — exit 0.
- With M1 (`continue` first in the Lstat-error branch), absent and present-unusable controls using `-run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/(absent_checkout_permits_fallback|present_but_unusable_stops_fallback)$'` — exit 0; both controls survive the mutant.
- With M1, `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/lstat_failure_stops_fallback$' -count=1` — exit 1 as expected; it got nil after continuing to the later checkout. Restored the clean production file and verified it with `cmp` — exit 0.
- Restored three-row test, `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1 -v` — exit 0; all three subtests pass.
- `go test ./internal/install -run 'LockedNetworkRepository|StateRead|Guard' -count=1` — exit 0.
- `go test ./internal/envprofile -run 'LockedNetworkRepository|StateRead|Guard' -count=1` — exit 0.
- `go test ./internal/stateread -count=1` — exit 0.
- `go test ./internal/envprofile -run 'TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse' -count=1 -v` — exit 0; guard reports 354/354 (100.0%).
- `go test ./internal/envprofile -run 'Test(Migrate|CredentialLink|StaleCredential|CodexCredentialStore|ResolvePassthrough)' -count=1` — exit 0.
- `go vet ./internal/install ./internal/envprofile` — exit 0.
- `GOOS=windows go vet ./internal/install ./internal/envprofile` — exit 0 (static validation; not a Windows runtime test).
- `go build ./internal/install ./internal/envprofile` — exit 0.
- `golangci-lint run ./internal/install ./internal/envprofile` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-final` first exited 1 because the two obsolete stateread rows named deleted tests; after their removal, rerun exited 0 with 449 rows across Linux, Darwin, Windows.
- `gofmt -l internal/envprofile/managed.go internal/envprofile/migrate.go internal/envprofile/state_read_guard_test.go internal/install/draftsources.go internal/install/draftsources_test.go` — exit 0, no output.
- `git diff --check` — exit 0; `git diff --cached --quiet` — exit 0; nothing staged.
- `git diff --quiet origin/main -- CHANGELOG.md` — exit 0.
- Rev9 path-identity audit described above — exit 0.

The full package commands were also attempted directly, but each reached Go's default 10-minute test timeout (exit 1): `go test ./internal/envprofile -count=1` timed out while running `TestStatusShellHookTrustMissingUnreadableAndStateFailures` (6 seconds into a child-process probe), and `go test ./internal/install -count=1` timed out while `TestDraftRepairRestoresDriftedContent` was active in Darwin `renamexNp`. These are timeout results, not passes; the focused task, migration, and guard suites above passed. The configured hosted handoff gate has not run yet and is expected to provide the Linux, macOS, and Windows runtime results.

The existing `## CHANGELOG entry (for release prep)` text earlier in this resource is retained. No CHANGELOG edit is included.

## Revision 6 — unreadable checkout code after cww1ov merge (2026-09-26)

### Classification decision

`lockedNetworkRepository` now returns the typed error from `stateread.Lstat` directly when inspecting a locked local checkout fails. Callers see `manager_state_unreadable`, and `errors.As` still reaches `*stateread.Error`. The other `source_snapshot_unavailable` wrapping branches remain unchanged.

This follows curator-spec `protocol/environments.md` §8.4.1: a failed stat and a path component that is not a directory are present-but-unreadable state, never absence, and cannot trigger an absence-shaped fallback. `protocol/skillfile-sources.md` §3 reserves `source_snapshot_unavailable` for a declared source that cannot be reached or read. Here the failure is inspecting the manager's local checkout path, so the outer diagnostic is `manager_state_unreadable`.

Counting Windows `ERROR_PATH_NOT_FOUND` beneath an existing regular-file parent as absent is incorrect. The parent exists and blocks traversal; `stateread.Lstat`'s ancestor check keeps the result unreadable. The Windows row uses this real fixture, asserts errno 3 and `os.IsNotExist` shape, and is required on Windows in `.github/ci/platform-cases.tsv` with no skip. Go reports `os.PathError.Op` as `lstat` on Darwin and `Lstat` (capital L) on Windows; the regression test pins each platform-specific API label.

The named regression is `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/lstat_failure_stops_fallback`. It now pins the outer diagnostic in addition to the typed error, original path failure, and no-fallback behavior. The absent, Lstat-failure, and present-but-unusable cases remain separate subtests.

### M1 narrowing proof on Darwin

M1 (`continue` as the first statement in the Lstat-error branch) was inserted temporarily into the refreshed source:

- Absent and present-but-unusable controls with M1: exit 0; those rows survive the narrowing mutant.
- `lstat_failure_stops_fallback` with M1: exit 1, expected; it received nil after continuing to the later checkout, so the row kills M1.
- Restored `internal/install/draftsources.go` and verified byte identity with the saved source using `cmp`: exit 0.
- All three rows after restoration: exit 0.

### Trunk refresh and Story identity

Fetched `origin/main` and refreshed the candidate onto `3803a75e` (2elcdc), then fetched again when trunk advanced to `e8620502` (20o9dk). The second refresh is recorded as `refresh_advanced`, exit 0, with checkpoint `fefd55d9` parented by `e8620502`. Combined all 296 non-board paths from the latest trunk update; the two overlapping files, `internal/install/draftsources.go` and `internal/registry/registry_test.go`, merged cleanly and keep both sides.

Reconstructed `TASK-260907-2as5sx` revision 9 from base `a48f584c` in a temporary audit copy. Across 74 rev9 paths, 50 non-leaf paths remain byte-identical. Twenty-one differences are paths updated by trunk since the rev9 base. `internal/envprofile/state_read_guard_test.go` is the previously documented explicit AST-guard update; its guard run reports 354/354 covered. There are zero unexplained byte mismatches and zero candidate paths outside the rev9, upstream-overlay, and leaf scopes.

`CHANGELOG.md` matches the recorded checkpoint (no revision-6 hunk; `git diff --quiet HEAD -- CHANGELOG.md` exit 0). The release-prep entry text remains in this results resource. No `LOGBOOK.md` edit was made; the task-scoped resource records the findings.

### Validation on this host (`GOOS=darwin`, each command ran directly)

- `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1 -v` — exit 0; all three subtests pass.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v` — exit 0; 354/354 production readers covered (240 via seam, 114 allowlisted).
- With M1, absent/present-unusable controls — exit 0; Lstat-failure row — exit 1 as expected; restored source `cmp` — exit 0; restored three-row test — exit 0.
- `GOOS=windows go vet ./internal/install` — exit 0 (static validation, not a Windows runtime test).
- `go build ./internal/install ./internal/envprofile ./internal/envregistry` — exit 0.
- `golangci-lint run ./internal/install ./internal/envprofile ./internal/envregistry` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-final` — exit 0; 449 rows checked across Linux, Darwin, and Windows, including all three locked-checkout rows.
- `gofmt -l internal/install/draftsources.go internal/install/draftsources_test.go internal/registry/registry_test.go` — exit 0, no files listed.
- `git diff --check` — exit 0; `git diff --cached --quiet` — exit 0.
- `git diff --quiet HEAD -- CHANGELOG.md` — exit 0.

The hosted Linux, macOS, and Windows gate is run by the configured developer handoff exactly once; its validation log will be attached by the handoff. Windows runtime behavior is pending that hosted lane.

### Final platform-label and M1 revalidation

Go's Windows `os.stat_windows.go` calls `stat("Lstat", ...)`; the Windows regression assertion now expects `PathError.Op == "Lstat"`, while Darwin expects `"lstat"`. This assertion pins the runtime API value for each platform.

After that test correction, the focused three-row test ran on Darwin and exited 0. M1 was reinserted as `continue` at the start of the production Lstat-error branch: the absent and present-unusable controls exited 0; the Lstat-failure row exited 1 as expected because it received nil and incorrectly fell through to the later checkout. The mutation was removed, `cmp` against the saved production file exited 0, and all three rows passed again with exit 0.

Final direct validations after the test correction: `env GOOS=windows go vet ./internal/install` exit 0; `go build ./internal/install ./internal/envprofile ./internal/envregistry` exit 0; `golangci-lint run ./internal/install ./internal/envprofile ./internal/envregistry` exit 0 with 0 issues; the 354/354 absence guard test exit 0; ledger consistency exit 0 with 449 rows; `gofmt -l` listed no files; `git diff --check` exit 0. These checks were run as separate processes. Windows runtime remains pending the hosted handoff lane.

## Revision 7 — platform-correct assertion

### Trunk refresh and CHANGELOG

Fetched `origin/main` at `316438cc3f830801c44adba75750a6b28a156933`, including the 2gt5f6 revision-2 landing. `task-board worktree refresh-candidate TASK-260925-h4syhu` exited 0 with `refresh_advanced`; refreshed candidate `243e914bd301234e3b19fe42c568f0723cfdf0fe` is parented by `316438cc`. The refresh retained the Story candidate on top of the new trunk.

Restored `CHANGELOG.md` from `origin/main`; `git diff --quiet origin/main -- CHANGELOG.md` exited 0. The existing `## CHANGELOG entry (for release prep)` section above still contains the entry text. No CHANGELOG delta is included against trunk.

### Windows assertion and classification

Updated `lstat_failure_stops_fallback` to expect the actual Windows `os.PathError.Op` value, `GetFileAttributesEx`; POSIX remains `lstat`. The assertion still unwraps the typed `*stateread.Error`, checks its unreadable kind and exact checkout path, checks the underlying `*os.PathError` operation and path, and requires no returned checkout. On Windows it also requires ERROR_PATH_NOT_FOUND (3) and its `os.IsNotExist` shape. That errno shape is not semantic absence here: the regular-file parent exists and blocks traversal, so the expected manager state remains unreadable and fallback must stop. The three real production-entry rows remain separate and listed for Linux, Darwin, and Windows; no skip row is used.

### Darwin tests and M1

All commands ran directly as standalone processes:

- Clean three-row test, `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1 -v` — exit 0; absent, Lstat-failure, and present-unusable subtests passed.
- With M1 inserted as the first statement in the Lstat error branch, absent and present-unusable controls — exit 0; both survived the mutant.
- With M1, `lstat_failure_stops_fallback` — exit 1 as expected; it received nil after continuing to the later checkout, so M1 is killed.
- Restored `internal/install/draftsources.go`; `cmp` against the saved pre-mutant bytes — exit 0.
- Three-row test after restoration — exit 0; all subtests passed.

### Validation

- `GOOS=windows go vet ./internal/install` — exit 0 (static validation; hosted Windows runtime is provided by the handoff gate).
- `go vet ./internal/install` — exit 0.
- `go build ./internal/install` — exit 0.
- `golangci-lint run ./internal/install` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-r7` — exit 0; 449 rows checked across Linux, Darwin, and Windows, including the three checkout subtests.
- `gofmt -l internal/install/draftsources_test.go` — exit 0, no output.
- `git diff --check -- internal/install/draftsources_test.go CHANGELOG.md` — exit 0.

The configured hosted gate, including Windows runtime execution, has not run in this producer worktree; developer handoff will run it once and attach its result.

### Latest trunk overlay merge and post-merge reruns

The 2gt5f6 trunk update changed 38 paths from the previous base; 18 non-board paths also intersected the preserved Story worktree delta. Merged the four CI ledgers and eleven other overlapping files with a three-way merge; restored the two new trunk credential test files. The sole textual conflict was the import block in `internal/envprofile/managed.go`; resolved it by retaining both `stateread` and `transaction`. No conflict markers remain. The candidate keeps the new trunk tests and the Story changes together.

This supersedes Revision 6's Windows `PathError.Op == "Lstat"` expectation. The Windows lane reported `GetFileAttributesEx`; Revision 7 now asserts that actual operation while retaining the original errno and path checks.

After this merge, all commands ran directly as standalone processes:

- `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1 -v` — exit 0; all three rows passed.
- `go test ./internal/envmarker -count=1` — exit 0.
- `go test ./internal/envprofile -run '^Test(ResolvePublishesCredentialMarkerThroughLockedJournal|ResolveRecoversCredentialMarkerAfterRenameCrash|MigrateSchema1UnlinkPublishesCompleteSchema2Marker|MigrateFailedRelinkKeepsOldLink|CredentialLinkRegularFileRefuses|StaleCredentialLinkRefusals|CodexCredentialStoreTOMLSpellings|CredentialLinkDirectory|CredentialLinkUnrecordedSymlinkRefuses|CredentialLinkTargetInspectionFailure|ManagerOwnedAbsenceReadsAreGuarded)$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^TestEnv(ResolveCredentialRecord|Migrate)' -count=1` — exit 0.
- `GOOS=windows go vet ./internal/install` — exit 0.
- `go vet ./internal/install ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0.
- `go build ./internal/install ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/install ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0, 0 issues.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260925-h4syhu-ledger-r7-merged` — exit 0; 450 rows checked across Linux, Darwin, and Windows.
- `gofmt -l` on the merged Go files and `internal/install/draftsources_test.go` — exit 0, no output.
- `git diff --check` — exit 0.
- `git diff --quiet origin/main -- CHANGELOG.md` — exit 0; `CHANGELOG.md` still matches trunk.
