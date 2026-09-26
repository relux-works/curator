# TASK-260926-2r1upt bound land preconditions (RUN-260926-27b879)

Accepted Change Request: CR-TASK-260926-2r1upt-1 revision 1 (review verdict rev1: accepted).
Board status observed: integrating (task and STORY-260924-3gd2d6). No status writes made by this run.

## Worktree delta (landing candidate)
- `git status --short`: only `M internal/install/draftsources_test.go`
- `git diff --stat`: 1 file changed, 172 insertions(+), 0 deletions
- No production change, no CHANGELOG.md edit, no LOGBOOK.md edit.
- Scope matches accepted CR: test-only per-guard rows for declaredDependencyReplaySources (directory-mismatch, repository-identity-mismatch, matching-declaration happy path).

## Fresh bounded verification run by this integration run
- `go test ./internal/install/ -run TestDraftFreshMachineTransitiveReplaySourceGuards -count=1 -v` -> exit code 0, PASS (all 3 subtests pass, ~31s).
- `go vet ./internal/install/` -> exit code 0, no output.
- `gofmt -l internal/install/draftsources_test.go` -> no output (clean), exit code 0.

## Mutant/AC standing (accepted from producer+reviewer evidence, not re-attacked here)
- Producer outcome resource TASK-260926-2r1upt_results.md records per-guard mutant survive-before/killed-after with real exit codes.
- Reviewer verdict resource TASK-260926-2r1upt_review-verdict-rev1.md records acceptance of rev1.
- This run did not mutate mutants; it re-ran the committed gate green.

## Landing transaction
- Per the binding integration assignment, this run did NOT execute `task-board worktree integrate` and made no board status writes.
- Worktree left UNCOMMITTED with the accepted candidate delta in place for the runner-operated landing transaction.
