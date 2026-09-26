# TASK-260907-2as5sx — revision 10 routing evidence

## Decision

No revision-10 repository changes were made in this developer run. The attached loop-detector response for rev9 says the leaf exceeded its revision budget (9 revisions against threshold 3), requires one leaf per catalog surface with its surface-table rows, coverage-map obligation, and revision budget, and explicitly says not to produce another revision of this leaf. That scope instruction conflicts with `2as5sx-rework-7.md`, which asks for a revision-10 code/test change and handoff.

## Evidence reviewed

- `TASK-260907-2as5sx_review-verdict-rev9.md` says the rev9 `blocked` input moved off the unreadable-Lstat branch and now tests the present-but-unusable branch at `internal/install/draftsources.go:234-236`; the narrowing mutant M1 (`continue` at the start of the Lstat error branch) survives on every platform.
- `2as5sx-rework-7.md` asks to restore the POSIX `blocked/child` Lstat-failure case, preserve present-but-unusable as a separate subtest, account for Windows cases, run M1 on darwin, and run `GOOS=windows go vet ./internal/install`.
- `TASK-260907-2as5sx_loop-response.md` is the controlling scope instruction for the next step: split by catalog surface before further rework; do not revise this leaf again.
- The Story currently has four child tasks, including this cross-surface task. The required catalog surface table and coverage-map rows are not included in this developer instruction set, so creating substitute rows here would be an unverified decomposition.

## Options considered

1. Implement rework 7 on this leaf. This would address the valid rev9 regression finding, but it directly violates the loop response's instruction not to produce another revision of this leaf.
2. Split and re-scope through the Story orchestrator. This follows the project-management bounded-leaf rule and allows each new leaf to carry the authoritative surface-table rows, coverage-map obligation, and its own revision budget. The Lstat regression should be routed to the leaf that owns `internal/install/draftsources.go` and its platform-case ledger.

Recommendation: the Story orchestrator creates the per-surface leaves and routes the outstanding rev9 Lstat finding to its owning leaf. Then a developer can implement that bounded follow-up and attach its own mutant and platform evidence.

## Validation performed in this run

- Required `set_status(..., status=development)`: exit 0.
- No repository file was changed by this run.
- No tests, lint, build, or mutant gates were run because the loop response prohibits another revision of this leaf pending decomposition.

## Board status effect

`task-board m 'set_status(TASK-260907-2as5sx, status=blocked)'` exited 0. The board set this task to `blocked` and reported a side effect demoting parent `STORY-260906-1a2i5a` from `development` to `integrating`.
