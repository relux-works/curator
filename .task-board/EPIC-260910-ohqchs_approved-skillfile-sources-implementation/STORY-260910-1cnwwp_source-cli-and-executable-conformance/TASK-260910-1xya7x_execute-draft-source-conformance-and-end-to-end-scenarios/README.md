# TASK-260910-1xya7x: execute-draft-source-conformance-and-end-to-end-scenarios

## Description
Execute draft source conformance and end to end scenarios. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/crossconformance and integration fixtures; consumer schema pinning and platform test lanes.

## Acceptance Criteria
Run all 102 draft schema cases, 3 snapshot vectors and 73 semantic cases against real production entry points, plus v1 regressions. Test local skill+script+compiled CLI+dependencies and broker-backed transport fixtures, transactions and OS path semantics. Record actual platform coverage, unsupported lanes and exact tested revision; do not claim release qualification from schema-only tests or alter frozen spec releases.
