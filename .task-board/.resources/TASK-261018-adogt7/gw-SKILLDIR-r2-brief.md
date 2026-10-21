# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: check the CHANGELOG combination after converge (developer)

Your Change Request revision 1 has not been reviewed. The reviewer spawn was refused with `revision_base_superseded`: your base 402fdbc9 predates the v0.15.0-rc.7 release commit db65e395, which rewrote `CHANGELOG.md`. The orchestrator converged the Story onto current main without a textual conflict, and revision 1 is now `stale` because nobody has looked at the combination.

## Do this

1. Inspect the converged `CHANGELOG.md`. Your entry must sit in the unreleased section above the v0.15.0-rc.7 heading, never inside an already released section; nothing duplicated, trunk's text unchanged. If the file has no unreleased section, remove your CHANGELOG edit entirely and put the entry text into the gaps file under "Changelog entry" instead; the release commit carries it. Change nothing else; the implementation needs no change in this round.
2. Rerun the validation revision 1 used on the converged tree, including the full local suite and the full hosted gate on the new exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
3. If the board requires it, run `task-board worktree refresh-candidate TASK-261018-adogt7`.
4. Add a short section "Refresh onto current main" to `TASK-261018-adogt7_gaps.md` with the commands and exit codes.

## Hand-off

Run `task-board handoff TASK-261018-adogt7 --role developer`, then END YOUR TURN.
