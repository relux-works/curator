# THE ONLY CURRENT INSTRUCTION — BUG-261002-ot3ea1 revision 5: drop the CHANGELOG change (orchestrator, binding)

Revision 4 still changes CHANGELOG.md, so it still intersects trunk and cannot be reviewed or landed. The fix is mechanical:

1. In the Story worktree run `git checkout c085b4d2 -- CHANGELOG.md`. CHANGELOG.md must be byte-identical to base c085b4d2; check with `git diff --quiet c085b4d2 -- CHANGELOG.md` and record the exit code.
2. Do NOT add trunk's line or yours. The rc.3 release-notes step adds every entry centrally.
3. Change nothing else. No re-test is needed beyond `git diff --stat c085b4d2`; the product code is unchanged from rev4, whose hosted gate was green.
4. Run `task-board handoff BUG-261002-ot3ea1 --role developer` and END YOUR TURN. Never edit LOGBOOK.md.
