# THE ONLY CURRENT INSTRUCTION — BUG-261017-47xdyt: check the combination after converge (developer)

The deciding review of CR revision 1 held every acceptance row:
- all six AC rows held;
- all seven narrowing mutants were killed;
- no product defect was found.

It withheld acceptance for one reason only. Trunk advanced to 402fdbc9, and that change also touches README.md (the CLIHELP and SNAPPUB landings both change it). The orchestrator converged the Story onto 402fdbc9 without a conflict. Revision 1 is now `stale`, because nobody has looked at the README combination.

## Do this

1. Inspect the converged README.md. Keep trunk's text and yours, in a sensible order, with nothing duplicated. Change nothing else unless the combination is wrong. The ownership implementation needs no change in this round.
2. Rerun the validation that revision 1 used on the converged tree. That includes the full local suite and the full hosted gate on the new exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
3. If the board requires it, run `task-board worktree refresh-candidate BUG-261017-47xdyt`.
4. Add a short section "Refresh onto 402fdbc9" to the gaps file, with the commands and exit codes.

## Hand-off

Run `task-board handoff BUG-261017-47xdyt --role developer`, then END YOUR TURN.
