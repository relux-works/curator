# TASK-260928-28epfn: kill-nofollow-parent-walk-mutants

## Description
Add production-entry rows so the managedPath parent-walk guard and the symlink refusal each have a row that fails when only that guard is removed (M1 Lstat->Stat, M2 refusal disabled). Cite environments §8.3.1.

## Scope
(define task scope)

## Acceptance Criteria
M1 and M2 killed individually with real exit codes; no production behaviour change unless a real gap is found
