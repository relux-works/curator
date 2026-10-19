# THE ONLY CURRENT INSTRUCTION — TASK-261018-70enqt: refresh the candidate onto current main (developer)

Your Change Request revision 2 has not been reviewed. The reviewer spawn was refused with `revision_base_superseded`: protected trunk advanced to 9c1605a7 (the OWNPRES landing), which also changes `CHANGELOG.md`, a path your candidate changes. The orchestrator's `worktree converge` stopped on a textual conflict in `CHANGELOG.md`; nothing was moved.

## Do this

1. Run `task-board worktree refresh-candidate TASK-261018-70enqt` and resolve the `CHANGELOG.md` conflict: keep trunk's entries and yours, in the file's order, nothing duplicated or dropped. Change nothing else; the qualification work itself needs no change in this round.
2. Rerun the validation revision 2 used on the refreshed tree, including the full local suite and the full hosted gate on the new exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
3. Add a short section "Refresh onto 9c1605a7" to `TASK-261018-70enqt_gaps.md` with the commands and exit codes.

## Hand-off

Run `task-board handoff TASK-261018-70enqt --role developer`, then END YOUR TURN.
