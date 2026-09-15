# TASK-260910-3eu4cy: publish-source-installs-and-locks-atomically

## Description
Publish source installs and locks atomically. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/transaction, install, adapters, runtimestore and repair/refresh integration.

## Acceptance Criteria
Publish lock, marker, runtime and adapters as one recoverable transaction; faults roll back consistently. Recheck physical output boundaries at write time. Status, repair and refresh enforce exact currentness and frozen inputs; protect unmanaged files and prove retarget/copy mutation failure paths.
