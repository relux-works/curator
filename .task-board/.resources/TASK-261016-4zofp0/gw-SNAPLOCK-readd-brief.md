# Re-publish SNAP-LOCK as revision 2 after convergence (TASK-261016-4zofp0)

Revision 1 was never reviewed. The operator withdrew it (`withdraw_cr`, state `aborted`) because trunk advanced on `README.md` and the board had no refresh route for a `ready` revision. To converge cleanly, the operator set aside only your README change, then ran `task-board worktree converge`. The workspace now sits on current curator main. All your other uncommitted and untracked files were carried over unchanged.

Do this and nothing more:
1. Re-add your README paragraph exactly once, in the same place relative to the surrounding text (after the "Use `go build ./...`" validation paragraph). The paragraph to restore is in the attached `SNAPLOCK-rev1-README.patch`. Keep every current trunk README line. Make no other README edits.
2. Make no behavior changes. Check that `git diff` against the Story base contains your rev 1 delta and nothing else of yours.
3. Re-run the focused tests and the full gate (`sh scripts/remote-gate.sh`), and record the exact commands and results.
4. Publish the candidate as a new Change Request revision and hand off for review as the board requires. The rules in rc7-code-common.md (attached) still apply. Your earlier refresh-blocker record stays as history; do not edit it.
