# BUG-260923-3mazfw results

## Root cause

Hosted macOS evidence shows that the runner intermittently refuses to start its preinstalled Git with the error “fork/exec /opt/homebrew/bin/git: permission denied”. The reported dry run reaches internal/install.Project, which calls gitignore.Ensure; internal/gitignore.Missing ran git check-ignore once and returned a child-start error as terminal. That converted a transient runner Git spawn refusal into failure for a candidate whose packages did not otherwise touch the affected path.

Run [35855550263](https://github.com/relux-works/curator/actions/runs/35855550263) failed in Test (macos-latest) on 2026-09-23. The job was runner GitHub Actions 1000017779, started at 11:36:56Z. At 11:46:47Z, internal/install.TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution reported:

> git check-ignore failed for ".claude/skills/": fork/exec /opt/homebrew/bin/git: permission denied

The underlying runner filesystem cause is unknown: these artifacts contain no file-mode, quarantine/xattr, or Homebrew update state. The CI workflow does not install or upgrade Homebrew Git in the test lane, so attributing this to a specific update or xattr would be unsupported.

## Frequency evidence

I checked the exact test in ten earlier macOS test-evidence artifacts. It passed in nine and failed in one (run 35855550263). This is a sample of earlier hosted artifacts, not a census of every run.

| Run | Time (UTC) | Exact test |
| --- | --- | --- |
| [35340496757](https://github.com/relux-works/curator/actions/runs/35340496757) | Sep 18 11:48:20 | Passed |
| [35481906193](https://github.com/relux-works/curator/actions/runs/35481906193) | Sep 20 01:47:35 | Passed |
| [35510984798](https://github.com/relux-works/curator/actions/runs/35510984798) | Sep 20 12:47:54 | Passed |
| [35544589324](https://github.com/relux-works/curator/actions/runs/35544589324) | Sep 20 23:36:36 | Passed |
| [35735392681](https://github.com/relux-works/curator/actions/runs/35735392681) | Sep 22 13:57:57 | Passed |
| [35739653315](https://github.com/relux-works/curator/actions/runs/35739653315) | Sep 22 14:34:33 | Passed |
| [35852658095](https://github.com/relux-works/curator/actions/runs/35852658095) | Sep 23 11:22:45 | Passed |
| [35852783703](https://github.com/relux-works/curator/actions/runs/35852783703) | Sep 23 11:24:43 | Passed |
| [35855550263](https://github.com/relux-works/curator/actions/runs/35855550263) | Sep 23 11:46:47 | Failed with the error above |
| [35855672743](https://github.com/relux-works/curator/actions/runs/35855672743) | Sep 23 11:49:23 | Passed |

The same macOS artifacts also show the error “fork/exec /opt/homebrew/bin/git: permission denied” at other install Git call sites: TestDryRunEffectBindingsSeeWhatARealOperationWrites in 35340496757, TestEndToEndInstall in 35481906193, and TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState in 35510984798. Those test rows passed, but confirm that the permission error recurs across unrelated Git operations.

## Fix and bounds

internal/gitignore now uses the named runCheckIgnoreWithEACCESRetry path. It retries once after 100 ms only when Git failed to start with an os.PathError whose operation is fork/exec and errno is EACCES. Git exit statuses, EPERM, ENOENT, and EACCES from another operation are not retried. A second error is returned with its error chain intact; persistent EACCES stays a tool error, so Ensure(..., fix=true) does not write .gitignore.

Injected tests drive the production Missing and Ensure entry points for transient EACCES followed by success, persistent EACCES, Git process exit, EPERM, ENOENT, and EACCES outside fork/exec. The persistent row asserts exactly two attempts, no policy-success error, preserved EACCES, and no .gitignore write. The installer-level persistent-spawn refusal test remains covered. CHANGELOG.md records the fix.

A narrowing mutation was previously run against the operation check: after temporarily removing the fork/exec restriction, TestMissingDoesNotRetryOtherSpawnErrors/EACCES_outside_fork-exec failed with two calls instead of one (expected exit 1). The production predicate was restored and the final tests below passed.

## Verification run in this worktree

Platform: Darwin/amd64.

| Command | Exit | Result |
| --- | ---: | --- |
| go test -count=5 ./internal/gitignore | 0 | Passed, including the new Git-exit no-retry row |
| go test -count=5 -run '^TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution$' ./internal/install | 0 | Passed |
| go test -count=5 -run '^TestProjectRefusesOnGitignoreSpawnFailure$' ./internal/install | 0 | Passed |
| golangci-lint run ./internal/gitignore ./internal/install | 0 | 0 issues |
| go vet ./internal/gitignore ./internal/install | 0 | Passed |
| go build ./internal/gitignore ./internal/install | 0 | Passed |
| gofmt -l internal/gitignore/gitignore.go internal/gitignore/gitignore_test.go | 0 | No output |
| git diff --check | 0 | No output |

The broader go test ./internal/gitignore ./internal/install was previously attempted and exited 1 when the installer package hit its 10-minute timeout in unrelated TestDraftFailureAtEveryTargetClassRestoresPriorState/90-consumer. It was not rerun; the affected install cases above were run five times each.

## Hosted post-change evidence and handoff

Run [35881328824](https://github.com/relux-works/curator/actions/runs/35881328824) includes the production retry and its transient/persistent tests. Its Test (macos-latest) job passed; the affected install test passed at 15:38:43Z and the retry/persistent/no-other-spawn-retry tests passed at 15:29Z. The overall workflow failed in the Windows lane at internal/snapshot.TestConcurrentGetAcceptsOneImmutablePublisher, unrelated to this change.

The configured hosted landing gate is reserved for task-board handoff and has not been rerun from this worktree yet. Findings are recorded in this board note and outcome resource; campaign rules prohibit direct LOGBOOK.md edits.


## Producer verification rerun (2026-09-23)

The affected checks were rerun directly in the Story worktree on Darwin/amd64. Each command exited 0:

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=5 ./internal/gitignore` | 0 | Passed five runs |
| `go test -count=5 -run '^TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution$' ./internal/install` | 0 | Passed five runs |
| `go test -count=5 -run '^TestProjectRefusesOnGitignoreSpawnFailure$' ./internal/install` | 0 | Passed five runs |
| `golangci-lint run ./internal/gitignore ./internal/install` | 0 | 0 issues |
| `go vet ./internal/gitignore ./internal/install` | 0 | Passed |
| `go build ./internal/gitignore ./internal/install` | 0 | Passed |
| `gofmt -l internal/gitignore/gitignore.go internal/gitignore/gitignore_test.go` | 0 | No output |
| `git diff --check` | 0 | No output |

The post-change hosted run [35881328824](https://github.com/relux-works/curator/actions/runs/35881328824) had a successful Test (macos-latest) job on runner 1000017947. Its artifact records the affected installer test passing at 15:38:43Z and the injected transient, persistent, and non-retry rows passing in internal/gitignore at 15:29Z. The workflow overall failed in the unrelated Windows snapshot test, as noted above. The configured handoff gate has not run yet; it is the next and only hosted landing-gate run for this handoff.

## Revision 3 (refresh)

### Refreshed base and candidate proof

The worktree is now aligned with protected trunk `fad881368b632f43a18b01681a8fc7def110cbae`. The successful `task-board worktree refresh-candidate BUG-260923-3mazfw` returned `refresh_already_current` with `TrunkOID`, `ReviewedTrunkOID`, and `BranchOID` all equal to that OID (exit 0). An earlier retry returned `candidate recovery index CAS conflict: staged state is not the recorded checkpoint` (exit 1); after aligning the dirty tree with the incoming trunk and combining the changelog, the refresh retry succeeded.

`git check-attr merge -- CHANGELOG.md` reported `merge: union`. The trunk's four new `.github/ci/platform-cases.tsv` rows remain present alongside its existing rows; no platform ledger or other trunk-only path remains in the candidate diff. `CHANGELOG.md` was combined with `git merge-file --union`, retaining trunk's new entries and the revision 2 fix.

`git status --short`:

```text
 M CHANGELOG.md
 M internal/gitignore/gitignore.go
 M internal/gitignore/gitignore_test.go
```

Per-file diff against refreshed `HEAD`:

| File | Added | Removed |
| --- | ---: | ---: |
| `CHANGELOG.md` | 5 | 0 |
| `internal/gitignore/gitignore.go` | 49 | 5 |
| `internal/gitignore/gitignore_test.go` | 114 | 0 |

These are revision 2's same three candidate paths: the changelog fix and the unchanged implementation/test patch. The refreshed tree contains no other candidate delta. The attached revision 2 verdict resource states `ACCEPTED`; it names `TestMissingDoesNotRetryOtherSpawnErrors` and the narrowing mutant that removes the `fork/exec` predicate.

### Refreshed-tree verification

| Command | Exit | Result |
| --- | ---: | --- |
| `sh .github/ci/ledger-consistency.sh` | 2 | Usage error: this script requires an evidence directory. |
| `sh .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` | 0 | 245 ledger rows checked across Linux, Darwin, Windows. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed. |
| `go test -count=5 ./internal/gitignore` | 0 | Passed five runs. |
| `go test -count=5 -run '^TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution$' ./internal/install` | 0 | Passed five runs. |
| `go test -count=5 -run '^TestProjectRefusesOnGitignoreSpawnFailure$' ./internal/install` | 0 | Passed five runs. |
| `golangci-lint run ./internal/gitignore ./internal/install` | 0 | 0 issues. |
| `go vet ./internal/gitignore ./internal/install` | 0 | Passed. |
| `go build ./internal/gitignore ./internal/install` | 0 | Passed. |
| `gofmt -l internal/gitignore/gitignore.go internal/gitignore/gitignore_test.go` | 0 | No output. |
| `git diff --check` | 0 | No output. |

For the review's narrowing mutant, a disposable copy removed only `pathErr.Op == "fork/exec"`. `go test -run '^TestMissingDoesNotRetryOtherSpawnErrors$' -count=1 ./internal/gitignore` exited 1 as expected: `TestMissingDoesNotRetryOtherSpawnErrors/EACCES_outside_fork-exec` observed 2 runner calls instead of 1. The refreshed-tree unmutated package test passed; the mutant was confined to `/tmp` and is not in the candidate.

Hosted evidence remains the accepted revision 2 evidence: run 35881328824 passed its macOS test job (its overall run failed on an unrelated Windows snapshot test), and run 35897148993 passed the full configured hosted lanes, including macOS and Windows. This revision only refreshes the base and reruns bounded local rows; it introduces no product or test changes. No `LOGBOOK.md` edit was made per the refresh instruction.

## Revision 4 (carry-forward republish)

Trunk moved to `1511b345`, so the orchestrator converged the Story worktree: the accepted revision 2 delta is carried uncommitted here. Per-path verification against `BUG-260923-3mazfw_change-request_rev2.patch`:

| Path | Trunk touched? | Result |
| --- | --- | --- |
| `internal/gitignore/gitignore.go` | No (diff base `2c87f375..bc3804c5` identical to rev2) | Byte-identical to revision 2; no change made |
| `internal/gitignore/gitignore_test.go` | No (diff base `87dd7fd1..fed82acf` identical to rev2) | Byte-identical to revision 2; no change made |
| `CHANGELOG.md` | Yes (intersecting path) | Both sides present: trunk's `Added` E4 entry retained, rev2 `Fixed` check-ignore EACCES entry retained; no duplication, no conflict markers |

`git status --short` shows only these three modified paths; no other candidate delta. Conflict-marker grep over `CHANGELOG.md` and `internal/gitignore/` found none. Content is unchanged from the ACCEPTED revision 2; hosted evidence (runs 35881328824 macOS green, 35897148993 full green) is carried forward from revisions 2–3.

Focused bounded run in this worktree (Darwin, real exit codes):

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/gitignore/...` | 0 | Passed |
| `go vet ./internal/gitignore/` | 0 | Passed |
| `go build ./internal/gitignore/` | 0 | Passed |
| `gofmt -l internal/gitignore/` | 0 | No output |
| `git diff --check` | 0 | No output |

All 12 DoD checklist items were already checked from the accepted revision; this republish changes no code and adds no new behavior. No `LOGBOOK.md` edit per the carry-forward instruction.

## Revision 5 (carry-forward republish)

Trunk moved to `948ae7c9`, so the orchestrator converged the Story worktree: the ACCEPTED revision 4 delta is carried uncommitted here. Per-path verification against `BUG-260923-3mazfw_change-request_rev4.patch` (3 paths):

| Path | Trunk touched? | Result |
| --- | --- | --- |
| `internal/gitignore/gitignore.go` | No (HEAD blob `2c87f37` equals rev4 pre-image `2c87f375`) | Byte-identical to revision 4 post-image (169 lines); verified by applying the rev4 hunks to the HEAD blob and comparing; no change made |
| `internal/gitignore/gitignore_test.go` | No (HEAD blob `87dd7fd` equals rev4 pre-image `87dd7fd1`) | Byte-identical to revision 4 post-image (305 lines); verified the same way; no change made |
| `CHANGELOG.md` | Yes (intersecting path; HEAD blob `171ec1f8` differs from rev4 pre-image `17773d62`) | Converge kept both sides: trunk's `Added` R5 script-worker-v1 entry retained, rev4 `Fixed` check-ignore EACCES entry retained; no duplication, no conflict markers |

`git status --short` showed only these three modified paths at verification time (before the CHANGELOG revert below; now two); no stray root `TASK-*`/`BUG-*` files and no `test/` or `ledger/` strays. Conflict-marker grep over `CHANGELOG.md` and `internal/gitignore/` found none. No other file was changed.

### CHANGELOG policy (orchestrator, 2026-09-24)

Per the carry-forward instruction, this task's `CHANGELOG.md` hunk was reverted entirely: `git checkout HEAD -- CHANGELOG.md`, and `git diff HEAD --stat -- CHANGELOG.md` is now empty (the file equals trunk's). The worktree now carries only the two Go-file modifications. The entry text is preserved verbatim below for the release-prep leaf, which writes all entries.

### Focused bounded run in this worktree (Darwin, real exit codes)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/gitignore/...` | 0 | Passed (`ok ... 8.106s`) |
| `go test -count=1 ./internal/gitignore/...` | 0 | Passed (`ok ... 5.112s`) |
| `go vet ./internal/gitignore/...` | 0 | Passed |
| `go build ./internal/gitignore/...` | 0 | Passed |
| `gofmt -l internal/gitignore/` | 0 | No output |
| `git diff --check` | 0 | No output |

Content is unchanged from ACCEPTED revision 4; hosted evidence is carried forward from revisions 2–4. All 12 DoD checklist items remain checked; this republish changes no code and adds no new behavior. No `LOGBOOK.md` edit per the carry-forward instruction.

## CHANGELOG entry (for release prep)

### Fixed

- `git check-ignore` now retries one child-start `EACCES` after 100 ms, then
  returns persistent permission errors so installation still fails closed.
