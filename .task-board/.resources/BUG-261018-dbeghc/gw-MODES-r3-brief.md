# THE ONLY CURRENT INSTRUCTION — BUG-261018-dbeghc: check the README combination after converge (developer)

Your Change Request revision 2 has not been reviewed. The reviewer spawn was refused with `revision_base_superseded`: trunk advanced to 513e0ee8 (the OWNPRES landing), which also changes `README.md`. The orchestrator converged the Story onto it without a textual conflict, and revision 2 is now `stale` because nobody has looked at the combination.

## Do this

1. Inspect the converged `README.md`. Keep trunk's text and yours, in a sensible order, nothing duplicated. Change nothing else unless the combination is wrong; the implementation needs no change in this round.
2. Rerun the validation revision 2 used on the converged tree, including the full local suite and the full hosted gate on the new exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
3. If the board requires it, run `task-board worktree refresh-candidate BUG-261018-dbeghc`.
4. Add a short section "Refresh onto 513e0ee8" to the gaps file with the commands and exit codes.

## Hand-off

Run `task-board handoff BUG-261018-dbeghc --role developer`, then END YOUR TURN.
