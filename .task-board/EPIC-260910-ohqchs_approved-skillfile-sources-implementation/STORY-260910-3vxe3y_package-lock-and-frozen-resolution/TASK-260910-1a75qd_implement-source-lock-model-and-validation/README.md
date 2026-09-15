# TASK-260910-1a75qd: implement-source-lock-model-and-validation

## Description
Implement source lock model and validation. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/managerlock or dedicated package lock module, protocoljson.

## Acceptance Criteria
Persist Skillfile.lock.json with exact package identities and frozen selection; separate machine path binding from portable content identity; validate stale/malformed lock and package membership; distinguish all three source identity arms. Never confuse environment context lock with package lock.
