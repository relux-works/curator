# TASK-260823-3fnobk: land-buildsource-encoded-path-fix

## Description
Land the focused Ubuntu candidate regression fix produced by RUN-260823-4cda48: internal/buildsource must admit case- and normalization-distinct encoded paths and reject only exact encoded duplicates (candidate test duplicate-build-source-path). The reviewed patch with passing unit tests is attached to TASK-260823-1l1p8q as buildsource-encoded-path-fix.patch (sha256 4d62e862...), worktree .temp/TASK-260823-1l1p8q/worktree. Review it on the merits, apply on a fresh worktree from origin/main, run the buildsource tests, land via PR with green required CI. Maintainer pre-authorization 2026-08-22: fully autonomous.

## Scope
(define task scope)

## Acceptance Criteria
Patch (or equivalent) merged to main with green CI; duplicate-build-source-path candidate case passes locally against the rc.9 candidate root.
