# BUG-260823-1zoub8: unify-agent-unset-diagnostics

## Description
The condition (agent requested, SSH_AUTH_SOCK unset) surfaces as two different messages depending on whether it is reached through run-wide credentials or through a scope selection; only one of them is machine-readable. Unify on one diagnostic.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
One message for one condition across both paths; test asserts it; go test green.
