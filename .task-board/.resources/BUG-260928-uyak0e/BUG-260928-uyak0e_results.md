# BUG-260928-uyak0e results — rev2 (Windows gate fix, run 36482327099)

Role body present (full context profile).

## Change since rev1 (tree a970450d; `git diff --stat a970450d`: pathboundary.go +38/-8, pathboundary_test.go +58)
- internal/pathboundary `validateWithOwnerAndEntryInfo`: one `vanished(err)` rule is now applied to every per-entry probe of the walk:
  the WalkDir error (directory open/readdir for recursion, not the root), Lstat/entry info, `isLink` (Windows handle/reparse probe), and `checkNode`
  (owner lookup + mutation-permission/DACL probe, via Failure.Cause). `errors.Is(err, fs.ErrNotExist)` → skip (SkipDir for directories);
  any other error → failure (fail closed). The root itself never counts as vanished. A symlink observed by lstat is still refused even if Readlink then fails.
- Cause on Windows: `entry.Info()` from ReadDir is cached, so the first probe to see the removal was the handle-based isLink → link_safety failure.

## New rows
- TestValidateSkipsEntryVanishedDuringOwnerProbe — owner probe returns a not-exist PathError → passes (portable repro of the Windows shape).
- TestValidateFailsClosedForOtherOwnerProbeErrors — any other owner probe error → ownership failure keeping the cause.
- Existing rows kept: removed-between-readdir-and-lstat → pass; replaced by symlink → link_safety refused; other lstat error → regular_types failure.

## Evidence (zsh, pipefail, real exit codes, darwin)
- `go test -count=1 ./internal/pathboundary` → exit 0
- `GOOS=windows go vet ./internal/pathboundary` → exit 0; `go vet ./internal/pathboundary` → exit 0; gofmt clean
- Mutant (vanished treats every error as skip): `go test -run FailsClosed ./internal/pathboundary` → exit 1, both FailsClosed rows fail (killed); source restored (diff -q identical)
- `go test -count=20 -run 'TestPathInstallCapturesDirtyUntrackedInsideGit|TestManagerOwnedAbsenceReadsAreGuarded' ./internal/envprofile` → exit 0 (270.6 s)
- Windows: not run locally (unverified until the hosted gate).

## CHANGELOG entry (for release prep)
- Fixed: path-source boundary validation no longer fails when an entry (e.g. `.git/objects/maintenance.lock`) vanishes during the walk; other probe errors still fail closed and replacement symlinks are still refused.
