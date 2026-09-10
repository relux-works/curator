# TASK-260910-35279p: service-boundary-memoization

## Description
curator-skill-registry: memoize SnapshotBoundary per log_size under the store lock; invalidate on head advance; keep the contiguity proof but stop recomputing the Merkle tree per request.

## Scope
(define task scope)

## Acceptance Criteria
Benchmarked test shows O(1) repeated reads; concurrency tests stay green
