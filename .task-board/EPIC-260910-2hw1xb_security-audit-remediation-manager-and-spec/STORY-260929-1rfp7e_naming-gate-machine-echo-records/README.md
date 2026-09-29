# STORY-260929-1rfp7e: naming-gate-machine-echo-records

## Description
Naming gate feedback loop: its failure report echoes the offending line content; the runner stores failure-log excerpts in board records (progress.md, append-only .activity events.ndjson), so every naming-gate failure reintroduces the token into main (10821e67 red on BUG-260928-uyak0e records). Fix: report path:line only; allowlist existing immutable machine-echo occurrences by path + line sha256.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
