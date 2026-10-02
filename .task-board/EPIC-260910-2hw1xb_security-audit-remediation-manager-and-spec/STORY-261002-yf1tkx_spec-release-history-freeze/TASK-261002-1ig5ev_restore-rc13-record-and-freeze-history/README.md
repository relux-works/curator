# TASK-261002-1ig5ev: restore-rc13-record-and-freeze-history

## Description
Restore release/1.0.0-rc.13.json to the v1.0.0-rc.13 tagged bytes and add a byte-freeze guard for published release records in generator and validate.

## Scope
(define task scope)

## Acceptance Criteria
release/1.0.0-rc.13.json byte-identical to the tag; generator never rewrites published records; byte-freeze guard with a failing negative; regenerate-check and validate green.
