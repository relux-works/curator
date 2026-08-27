# TASK-260823-37vckg: land-multi-project-dryrun-binding

## Description
curator repo: candidate-conformance Ubuntu lane fails internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only — published dry-run scope multi-project has no executable binding. The fix exists on open curator PR 14 (commit d345420) but is not on main. Review PR 14 on its merits, verify it fixes exactly this failure (run the test), and land it on curator main via the repo PR flow with green required CI. Work from a worktree based on origin/main (the local main checkout has diverged; do not use it). Maintainer pre-authorization 2026-08-22: fully autonomous, no human approval gates.

## Scope
(define task scope)

## Acceptance Criteria
PR 14 (or an equivalent extraction of d345420) merged to curator main with green CI; the named test passes on main.
