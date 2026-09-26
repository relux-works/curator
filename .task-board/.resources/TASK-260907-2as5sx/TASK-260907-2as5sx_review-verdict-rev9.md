# Review verdict — TASK-260907-2as5sx rev9: CHANGES REQUESTED

## Verified OK
- (1) marker.go refresh: diff a48f584c..cf41dd62 on internal/marker/marker.go is purely additive (ReadState/InvalidError over the 1kpw4w base); `go test ./internal/marker ./internal/stateread` ok.
- (3) The CHANGELOG/profile.go/profile_git_reinstall_test.go delta vs base is exactly the landed trunk commit d9cb8465 (187z6x). `git diff HEAD cf41dd62` on those paths is empty, so the candidate adds no CHANGELOG edit. No stray files: the untracked files are all state-read tests or the stateread package.

## F1 (blocking): the Windows fix removes the only kill for the failed-read branch of lockedNetworkRepository
- In rev8 the row used `makeLock("blocked/child")`, where `blocked` is a regular file. On POSIX, Lstat gets ENOTDIR, which reaches the `stateread.Lstat` error branch at internal/install/draftsources.go:228-230. On Windows it gets ERROR_PATH_NOT_FOUND, which is treated as absent (the reported failure).
- In rev9 (internal/install/draftsources_test.go:56-66) the row uses `makeLock("blocked")`. That exercises the *present-but-not-a-repo* branch (EnsureRepo → `stateread.UnusableError`, draftsources.go:234-236). It no longer exercises the unreadable Lstat branch on any platform.
- Mutant M1 inserts `continue` as the first statement of `if err != nil {` at draftsources.go:228, so an Lstat failure falls back like absence. It is the exact §8.4 collapse at this site. It **survives**: `go test ./internal/install -run 'TestLockedNetworkRepository|StateRead|ReadState|Generation'` passes, and so does `go test ./internal/envprofile -run 'StateRead|Guard'`. The AST guard does not see it.
- So "unreadable" is not made real on Windows. The row was renamed onto a different branch, and the site's narrowing-mutant proof (AC) is lost on every platform.
- Required: keep a row that drives the Lstat failure branch and kills M1. For example, keep the rev8 `blocked/child` shape on POSIX with a `runtime.GOOS != "windows"` condition, and add a Windows-real unreadable input (a checkout path Lstat cannot stat: an invalid name/reserved device path, or a deny-ACL parent). Keep the present-but-unusable row as a separate subtest. If no Windows-real Lstat failure exists, record it as a platform-cases/skip-classes ledger row with the reason, not as a silent rename. Also say whether Windows ERROR_PATH_NOT_FOUND-under-a-file counting as absent is an accepted bound of stateread.Lstat (a Windows §8.4 collapse in the seam itself). Either type it or state it as a bound.

## (4) Hosted gate
There is no hosted-gate evidence for rev9 on the board or in the CR. It must be green on all lanes, Windows included, for the resubmission.

Focused runs only (host memory constraint); I did not run the full suite.
