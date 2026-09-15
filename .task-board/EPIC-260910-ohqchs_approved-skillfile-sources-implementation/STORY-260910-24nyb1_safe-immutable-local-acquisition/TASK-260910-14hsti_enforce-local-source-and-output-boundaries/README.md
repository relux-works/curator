# TASK-260910-14hsti: enforce-local-source-and-output-boundaries

## Description
Enforce local source and output boundaries. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/snapshot, staging, privatedir and adapter destination planning.

## Acceptance Criteria
Canonicalize physical paths, symlinks and case semantics; distinguish authored agents from managed outputs; validate operator root_inputs; reject unsafe overlap before traversal and recheck at publication; protect unmanaged files. Path source . with a safe selected subdirectory remains valid.
