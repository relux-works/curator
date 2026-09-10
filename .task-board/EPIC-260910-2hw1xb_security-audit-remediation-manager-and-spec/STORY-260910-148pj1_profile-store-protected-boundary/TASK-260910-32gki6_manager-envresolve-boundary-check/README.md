# TASK-260910-32gki6: manager-envresolve-boundary-check

## Description
curator: validate store entry ownership/permissions and verify recorded surface hashes for system-prompt and root-context files during env resolve; report drift instead of trusting link targets alone.

## Scope
(define task scope)

## Acceptance Criteria
env resolve tests prove a swapped store entry is detected
