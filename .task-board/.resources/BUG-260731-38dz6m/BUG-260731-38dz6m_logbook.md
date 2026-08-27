# BUG-260731-38dz6m logbook

## 2026-08-01 — Windows unreadable-directory fixture

The first Windows helper used a protected DACL denying
`FILE_LIST_DIRECTORY`. Native run 30643616475 showed that the hosted runner
still listed the directory. Go 1.25's Windows `Openat` path uses
`FILE_OPEN_FOR_BACKUP_INTENT`; Windows grants backup read access when the token
has `SeBackupPrivilege`, so a DACL is not a reliable unreadability boundary for
an administrative runner.

Decision: hold a `GENERIC_READ` directory handle with share mode zero. A later
directory read-open then fails through Windows share-access enforcement rather
than ACL evaluation. The test asserts the fixture by requiring `os.ReadDir` to
fail before exercising either fingerprint implementation.

Evidence: signed commit `a134fdc`; native Windows attempt 2, run 30668611796;
artifact 8809078378.

## 2026-08-01 — unrelated managerlock flake

Attempt 1 on signed head `a134fdc` passed all six cases owned by this bug and
failed only
`internal/managerlock.TestSubprocessBuildKeyDeduplicationAcrossProjects`.
That PR 13-owned case passed on the prior PR 15 run, and this head changes no
managerlock file. The failed-job rerun on the identical SHA passed the case and
the whole Windows lane. No managerlock change was absorbed into this scope.
