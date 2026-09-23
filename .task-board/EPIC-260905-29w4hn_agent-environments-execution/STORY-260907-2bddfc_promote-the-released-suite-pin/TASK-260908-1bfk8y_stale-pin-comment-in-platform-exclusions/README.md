# TASK-260908-1bfk8y: stale-pin-comment-in-platform-exclusions

## Description
F1 from TASK-260906-284db9 review cycle 1: .github/ci/platform-exclusions.tsv:11 claims the committed released pin publishes no qualification vector; both the old and the new pin publish vectors/conformance-claim-v3-qualification.json. Correct or remove the rationale. Non-blocking, not rc-blocking.

## Scope
.github/ci/platform-exclusions.tsv prose (and, only if the default_excluded_on column has no consumer, the column plus its reader and the gate self-test); no exclusion row change

## Acceptance Criteria
1) the rationale no longer claims the committed pin publishes no qualification vector; it states what default_excluded_on is for, backed by a cited consumer analysis (file:line) — or the column and reader are removed together; 2) gate self-tests pass with real exit codes; 3) CHANGELOG only if behaviour changes
