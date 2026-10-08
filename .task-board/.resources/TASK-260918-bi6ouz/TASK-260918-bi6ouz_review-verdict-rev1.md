# TASK-260918-bi6ouz — review verdict, CR rev1: CHANGES REQUESTED

Reviewer: claude-opus-5-5 (low). Disposable clone /tmp/rv at 48da2690 + CR patch; `git write-tree` = 86e89d6017b60b4afa374ff2fac48b5dbd376b1b (tree identical).

## Verified OK
- Table constant `dotfileStateTable` (managed.go) matches 802caee §9.5 cell by cell, rows in order, explicit none cells.
- Fixture testdata/dotfile-manager-table-802caee.md sha256 6575912115cf5b87facc0a6981e2d17bbcc1716f6ca7db3e0da5ddbf98853670 (matches cited hash); all 7 lines occur verbatim in `git show 802caee:protocol/environments.md`.
- Resolution rule (unset/empty/non-absolute XDG → default) matches §9.5 "Resolution".
- Skip row uses existing `root-content` class in platform-cases.tsv.
- zsh, pipefail: `CURATOR_CONFORMANCE_ROOT=<curator-spec main eadb1c0>/conformance/v1 go test ./internal/envprofile -run 'Dotfile|ForeignManager|Takeover' -count=1 -v` → PASS rc=0; TestDotfileManagerVectors ran (6 subtest PASS lines, no skip).

## Findings
F1 (surviving mutant, production entry unpinned) — internal/envprofile/managed.go `foreignManagerHint`: replacing `os.Lstat` with `os.Stat` in the production call `foreignManagerHintAt(home, runtime.GOOS, os.Getenv, os.Lstat)` SURVIVES every test, including TestDotfileManagerVectors with the real spec root and TestTakeoverWarnsDotfileHeuristic. The lstat discipline is only proven at helper level with an injected lstat; the AC's "executed through the real inventory path" does not reach it for the symlink row. Fix: a test through the production entry (repair/use takeover path) with a symlink-to-directory at a resolved table path (e.g. ~/.local/share/chezmoi → real dir) asserting no `environment_foreign_manager_suspected`; confirm it kills the Stat mutant.

F2 (spec deviation) — managed.go `foreignManagerHintAt`: a non-ENOENT lstat failure on one row aborts the whole scan and returns no hint; TestForeignManagerHintLstatDiscipline "failed inspection stays unknown" pins that abort. §9.5 (802caee, environments.md:2500-2516) says a failed inspection is "not present" (never absent) and "The manager scans the table in row order and reports the first row whose location is present". With chezmoi path EACCES and ~/.local/share/yadm a real directory, spec → notice names yadm; candidate → no notice. The vector set has no case distinguishing this (home-manager-linux-unreadable-quiet has all later rows absent). Fix: treat a failed row as not-present and continue the scan (never recording it as absent); update the unit test to require the later present row, or, if the producer reads the spec differently, raise it as a spec question rather than pinning the abort.

## Not re-verified
Windows-lane execution of the Windows cells: no hosted gate evidence for this revision reviewed yet (gate runs at handoff); unverified.
