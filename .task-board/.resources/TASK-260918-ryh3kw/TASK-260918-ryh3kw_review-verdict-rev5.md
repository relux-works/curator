# TASK-260918-ryh3kw — review verdict rev5: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate: base eca2bf27, tree b82d529f, 17 paths; worktree == candidate (git diff b82d529f empty). Shell zsh, pipefail, real exit codes.
Conformance root: fresh clone of curator-spec @ 23435129 (rc.13), CURATOR_CONFORMANCE_ROOT=<clone>/conformance/v1.

## Rev4 finding F1 (not combined with trunk; ledger named 1wc76r) — FIXED
- `git merge-tree --merge-base 0ffe2e1d rev4 eca2bf27` → 04876fca with conflicts only in conformance-case-counts.tsv and conformance-gaps.tsv; rev5 resolves them to the trunk side (gaps.tsv == trunk: 1wc76r rows gone, 1cenbr rows kept as trunk has them) with count 37 consistent with the profile/restore split. Remaining deltas vs auto-merge = requested work only (root-artifacts/platform-cases notes, conformance test split, stateread symlink-ancestor hardening + rows). No CHANGELOG/LOGBOOK, no stray files.

## Restore vectors (§8.4.1 backup record)
Driven by trunk's cmd/curator TestEnvUnmanageBackupRecordVectors (env_unmanage_test.go:46, landed with 1wc76r) through `env unmanage --restore-backups` → envprofile.newestBackupFiles; envprofile test now asserts exactly those 2 restore cases are split off (37+2) instead of ledgering them.
- go test ./cmd/curator -run EnvUnmanageBackupRecordVectors -v → both PASS, rc=0
- Mutant M5 (ReadDir failure → treated as absent, unmanage.go:190) → unreadable-restore-stops FAILS, rc=1 (killed)
- Mutant M6 (absent → error, unmanage.go:184) → absent-restore-nothing FAILS, rc=1 (killed)
- Bound (trunk code, not this CR): M4 Lstat failure on backup root → absent SURVIVES (rc=0); vector injects only the ReadDir failure. Residual for the owner of unmanage.go, non-blocking.

## stateread changes vs rev4
missingPathWith now Lstat's ancestors; a symlink ancestor counts as a directory only if it Stat-resolves to a directory; broken/looping/file links → unreadable. Strengthening (rev4 Stat walked past a broken link to an existing root → absent). Walk bounded by path depth, stops at first existing ancestor.
- go test ./internal/stateread → ok rc=0
- M1 ancestor walk removed → TestReadsDistinguishAbsentFromBlockedParent FAIL rc=1 (killed)
- M2 any symlink ancestor = traversable → same test FAIL rc=1 (killed; new broken-link row)
- M3 ENOTDIR folded into absence → read-file/read-dir/stat/lstat/read-regular-file subtests FAIL rc=1 (killed)
- M7 new os.IsNotExist site added in envprofile/managed.go → TestManagerOwnedAbsenceReadsAreGuarded FAIL rc=1 (state_read_guard_test.go:215) (killed)

## Focused runs
- go test ./internal/envprofile -run 'ReadFailure|Guarded|Restore|Unmanage' → ok rc=0 (47 PASS lines, incl. 37 read-failure cases)
- Hosted gate on rev5: green per orchestrator/validation log (not rerun here).

Verdict: ACCEPT revision 5.
