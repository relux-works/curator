# TASK-260908-2kqa77: goreleaser-config-value-gate

## Description
Follow-up recommended by TASK-260908-1jv1h3 cycle-3 verdict section 5: nothing in-repo reads .goreleaser.yml and goreleaser check discriminates 0 of 6 wrong-value mutants. Add a gate asserting the three parsed auto values including case (a grep for the token auto is defeated by the Auto mutant, proven). Use the cycle-3 row C (pre-change rc publishes) as the negative test. Non-blocking, not rc-blocking.

## Scope
(define task scope)

## Acceptance Criteria
(define acceptance criteria)
