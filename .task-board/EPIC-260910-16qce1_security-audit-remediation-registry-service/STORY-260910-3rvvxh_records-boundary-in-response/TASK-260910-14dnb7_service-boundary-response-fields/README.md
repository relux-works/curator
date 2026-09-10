# TASK-260910-14dnb7: service-boundary-response-fields

## Description
curator-skill-registry: include the committed boundary (version, log_size, head, merkle_root, created_at) in /v1/records and /v1/log page envelopes per the spec revision from TASK-260910-1b1ens; keep the field closed and covered by conformance tests.

## Scope
(define task scope)

## Acceptance Criteria
Responses carry the boundary; shared conformance vector passes
