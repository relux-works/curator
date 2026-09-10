# TASK-260910-1b1ens: spec-records-boundary-envelope

## Description
curator-spec: extend the records/log response schemas with the committed snapshot boundary (version, log_size, head, merkle_root) or equivalent signed inclusion evidence, and require clients to reject pages evaluated below their persisted high-water.

## Scope
(define task scope)

## Acceptance Criteria
Protocol revision merged with vectors for the stale-page rejection
