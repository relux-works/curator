# TASK-260922-cww1ov — Revision 8 refresh logbook

## Finding: refresh candidate lifecycle refused

On 2026-09-24, the producer combined the `fad88136..1511b345` incoming delta into the uncommitted Story candidate, excluding `.task-board/**`. The `git apply --3way` attempt against all paths returned 1 at the task-edited `CHANGELOG.md`; applying the independent paths separately returned 0, and both task and trunk changelog blocks remain in the final file. All incoming independent files are present and unstaged. No incoming `internal/envprofile` reader changed.

`task-board worktree refresh-candidate TASK-260922-cww1ov` refused with `candidate refresh requires a rework revision; TASK-260922-cww1ov is ready`. Worktree status reports `CR-TASK-260922-cww1ov` revision 7 as ready. A producer-side attempt to retire it with `withdraw_cr` returned `change_request_withdraw_unauthorized`: only the orchestrator may retire the unintegrated CR that governs its own routing; otherwise it must be answered through review. The producer did not bypass this ownership boundary or hand-edit replay state.

The orchestrator needs to route revision 7 through authorized withdrawal or review and return the element to a refresh-eligible state. The producer must then retry the sanctioned refresh and run the post-refresh checks.

## Verification anomaly

`go test -count=1 -timeout 9m ./internal/envprofile/...` exited 1 after 540.821 seconds. The timeout was in pre-existing `TestSCPOverlayResolvesAsGit`, subtest `git@example.com:personal`, blocked waiting for `git clone` over SSH. Focused 0017 API tests and CLI tests passed; `go vet`, build, formatting, the incoming GoReleaser gate tests, `gate-selftest.sh` (198/198), and `git diff --check` also passed. These checks used the combined but unrefreshed candidate. Hosted lanes were not rerun after refresh because the board lifecycle refused the refresh.

## Board route

After attaching the updated results and this logbook, `set_status(TASK-260922-cww1ov, status=blocked)` exited 0. The board moved the task to `blocked` and demoted the Story to `integrating`. No refreshed Change Request was published, and developer handoff was not run while the board-owned refresh route remains unavailable.
