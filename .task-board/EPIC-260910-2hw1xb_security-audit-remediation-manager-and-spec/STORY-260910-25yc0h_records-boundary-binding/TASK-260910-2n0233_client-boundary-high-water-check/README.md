# TASK-260910-2n0233: client-boundary-high-water-check

## Description
curator: parse the boundary from /v1/records responses and reject or warn when it is below the persisted registry high-water; optional strict mode replays /v1/log for revoked decisions.

## Scope
(define task scope)

## Acceptance Criteria
Client test proves a stale-page response is not accepted as current
