# THE ONLY CURRENT INSTRUCTION — TASK-261017-sxydvh: check the combination after converge (developer)

Revision 1 was ready for review, but the reviewer spawn was refused with `revision_base_superseded`. Trunk advanced to 402fdbc9, and the CLIHELP and SNAPPUB landings there also change README.md. The orchestrator converged the Story onto 402fdbc9 without a conflict. Revision 1 is now `stale`, because nobody has looked at the README combination.

## Do this

1. Inspect the converged README.md. Keep trunk's text and yours, in a sensible order, with nothing duplicated. Change nothing else unless the combination is wrong. The implementation needs no change in this round.
2. Rerun the validation that revision 1 used on the converged tree. That includes the full local suite and the full hosted gate on the new exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
3. If the board requires it, run `task-board worktree refresh-candidate TASK-261017-sxydvh`.
4. Add a short section "Refresh onto 402fdbc9" to the task's outcome, with the commands and exit codes.

## Hand-off

Run `task-board handoff TASK-261017-sxydvh --role developer`, then END YOUR TURN.
