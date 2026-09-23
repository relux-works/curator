# TASK-260908-2kqa77: goreleaser-config-value-gate

## Description
Follow-up recommended by TASK-260908-1jv1h3 cycle-3 verdict section 5: nothing in-repo reads .goreleaser.yml and goreleaser check discriminates 0 of 6 wrong-value mutants. Add a gate asserting the three parsed auto values including case (a grep for the token auto is defeated by the Auto mutant, proven). Use the cycle-3 row C (pre-change rc publishes) as the negative test. Non-blocking, not rc-blocking.

## Scope
.github/ci gate (or a tools/ Go test) reading .goreleaser.yml, its gate-selftest pin, CHANGELOG; no .goreleaser.yml semantic change, no release.yml change

## Acceptance Criteria
1) a parsing gate asserts every brews[*].skip_upload and release.prerelease equal the exact string auto (case-sensitive) and fails naming field, entry and observed value; 2) it runs in the existing CI lanes on every push and is pinned by gate-selftest.sh; 3) executed negative rows: field absent (row C), Auto (row E), sometimes, true, ato each fail, committed file passes; 4) CHANGELOG entry
