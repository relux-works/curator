# TASK-261014-2f0qnv: repair-preserve-lock-impl

## Description
Implement env resolve --repair --preserve-lock per curator-spec 6648e18: repair that never changes the lock hash, refusing with environment_repair_needs_lock_migration before the repair lock and any write

## Scope
(define task scope)

## Acceptance Criteria
curator implements env resolve --repair --preserve-lock per curator-spec 6648e18 with tests covering every vector case
