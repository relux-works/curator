# TASK-260910-1o9x1f: load-machine-owned-repository-endpoint-policy

## Description
Load machine owned repository endpoint policy. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/config, identity, gitcred.

## Acceptance Criteria
Implement source-policy schema, exact canonical identity, one/two distinct endpoints, provider refs, order and pin. URL without policy attempts declared URL once; logical identity without entry fails; invalid/unreadable policy fails before network. Package inputs cannot introduce providers or commands.
