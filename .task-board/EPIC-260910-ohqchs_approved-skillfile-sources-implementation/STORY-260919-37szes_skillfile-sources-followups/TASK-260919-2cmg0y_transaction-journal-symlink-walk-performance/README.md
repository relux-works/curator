# TASK-260919-2cmg0y: transaction-journal-symlink-walk-performance

## Description
Follow-up from TASK-260910-3eu4cy review (F-W1): the transaction engine walks filepath.EvalSymlinks per target per journal save (saveJournal -> validateIndependentTargetNamespaces -> canonicalNamespacePath), making a late-class rollback cost 5-6 minutes on Windows and the failure-at-every-target-class sweep ~25 minutes there. Cache canonical namespace paths per transaction (invalidate on the boundary recheck), revisit the fsync policy, keep every proof; measure the Windows sweep before/after on the hosted gate. Not a landing blocker; the CI budget was raised by TASK-260919-3ux95w meanwhile.

## Scope
internal/transaction journal save path and namespace canonicalization; tests measuring the sweep

## Acceptance Criteria
Windows internal/install sweep TestDraftFailureAtEveryTargetClassRestoresPriorState at least 2x faster on the hosted gate with every rollback proof unchanged; canonical path cache invalidated by the per-write boundary recheck; no new skip.
