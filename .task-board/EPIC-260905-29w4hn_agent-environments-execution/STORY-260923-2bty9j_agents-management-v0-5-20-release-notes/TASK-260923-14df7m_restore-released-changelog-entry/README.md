# TASK-260923-14df7m: restore-released-changelog-entry

## Description
CHANGELOG-only: restore the heading to '## Unreleased' (repo convention at tags) and the released F-M1b (v0.5.18) bullet byte-for-byte to its 4e229cc text; add one new bullet for F-M1c (Decision 0018 choice 4 known-conflict refusal under yolo, ErrNativePolicyConflict / *NativePolicyConflictError{Selector, Placement}, permission-grammar-v2 for Claude and Codex, Pi stays v1, release v0.5.20).

## Scope
(define task scope)

## Acceptance Criteria
git diff 4e229cc -- CHANGELOG.md shows only the new bullet; no other path changes; go test ./... exit 0
