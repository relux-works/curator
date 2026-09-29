# BUG-260928-uyak0e review verdict rev2 — ACCEPTED
Reviewer: claude-opus-5-5 low. Candidate: base 3f60f7f0, tree 7096aa0a; worktree diff vs base identical (3 files, +205/-1).
## Checks
1. Vanished rule: pathboundary.go vanished() applied to walkErr (dir open for recursion; SkipDir for dirs), entryInfo (Lstat), isLink, and checkNode Failure.Cause (owner/DACL probes). IsAbsent = errors.Is(err, fs.ErrNotExist) (pathboundary.go:251) covers Windows codes. Root itself never skipped (path==absolute). Other errors fail closed.
2. Symlink replacement: TestValidateRejectsEntryReplacedBySymlinkBetweenReadDirAndLstat → CheckLinkSafety. Readlink error after an observed link still fails (not skipped).
3. Fixture: path_capture_test.go sets gc.auto=0, maintenance.auto=false.
4. Local runs (zsh, pipefail, real rc): go test ./internal/pathboundary -count=1 rc=0; GOOS=windows go vet ./internal/pathboundary rc=0; go test ./internal/envprofile -run 'TestPathInstallCapturesDirtyUntrackedInsideGit$|TestManagerOwnedAbsenceReadsAreGuarded' -count=10 rc=0 (103 s).
## Mutants (disposable copy)
- M1 every error skipped (drop !IsAbsent): KILLED — FailsClosedForOtherEntryLstatErrors + FailsClosedForOtherOwnerProbeErrors, rc=1.
- M2 vanished only in Lstat (disable checkNode-cause skip): KILLED — SkipsEntryVanishedDuringOwnerProbe, rc=1.
- Bound: the isLink-probe skip is not killable on unix (isLink uses lstat info, no error); pinned only by the hosted Windows lane (rev1 failure row removed-directory, green on rev2 gate).
## Security
Skipping a vanished entry admits nothing: validation is a check of the source directory, not the store; import.go re-validates the staged+protected final store (import.go:586) and path sources are revalidated at use (pathsource.go:92/102). A file re-appearing after the walk is the same class as "later edits change nothing until reinstall" (environments §4).
No CHANGELOG/LOGBOOK edits.
