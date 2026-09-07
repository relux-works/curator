# TASK-260908-ranc5y: a1-fragment-delivery

## Description
SPEC section 4.1: real fragment subprocess with repair, closed parsing and CCJ-1 digest.

## Scope
(define task scope)

## Acceptance Criteria
SPEC4.1 production subprocess always resolves with repair, preserves argv boundaries, forwards stderr and maps documented error codes. Strict closed launch-env-fragment-v1 parsing and CCJ-1 digest match actual A0 fragments; malformed/unknown/duplicate inputs fail closed. Tests exercise real subprocess boundaries, schema negatives, channels and digest canonicalization. Only evidence-backed E6 stderr wording is clarified in SPEC; no other launcher stages or imports are added.
