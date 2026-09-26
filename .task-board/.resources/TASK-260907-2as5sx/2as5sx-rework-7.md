# TASK-260907-2as5sx — rework 7 (THE ONLY CURRENT INSTRUCTION)

Revision 9 CHANGES_REQUESTED, F1 (see `TASK-260907-2as5sx_review-verdict-rev9*`): your Windows fix moved the row from the unreadable-Lstat
branch (internal/install/draftsources.go:228-230) to the present-but-unusable branch (:234-236); reviewer mutant M1 (`continue` as the
first statement of the Lstat `if err != nil {`) now SURVIVES on every platform. Required:
1. Keep a row that drives the Lstat-failure branch and kills M1: the rev8 `blocked/child` shape on POSIX (guarded by runtime.GOOS), plus a
   Windows-real Lstat failure if one exists (reserved device name / invalid name / deny-ACL parent). If no Windows-real Lstat failure
   exists, record a platform-cases/skip-classes ledger row with that exact reason — never a silent rename.
2. Keep the present-but-unusable row as its own subtest.
3. State explicitly whether Windows ERROR_PATH_NOT_FOUND under a regular file counting as "absent" is correct behaviour (production
   semantics) and why.
4. Show M1 killed (real exit codes) on darwin; `GOOS=windows go vet ./internal/install`. Focused tests only — host memory is tight.
No CHANGELOG edit; artifacts only in $TMPDIR. Append "Revision 10 — Lstat-failure row restored", `resource update`,
`task-board handoff TASK-260907-2as5sx --role developer`. A write-boundary `policy warn` block is a warning.
