# BUG-260823-1vx45a: managerlock-windows-tempdir-contention

## Description
Windows CI intermittently fails managerlock subprocess tests: t.TempDir cleanup hits a still-held .lock file. Pre-existing, reproduced twice on 2026-08-23 PR rounds in tests untouched by the changes under review.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Root cause identified; fix or deterministic guard; both tests stable across 5 consecutive windows CI runs.
