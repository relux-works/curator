# THE ONLY CURRENT INSTRUCTION — BUG-261002-ot3ea1 republish → revision 4 (orchestrator, binding; REPLACES the earlier version of this note)

Correction: the orchestrator's `worktree converge` FAILED, because the CHANGELOG.md delta conflicts with trunk 64345d71. Your workspace is still on base c085b4d2.

Resolution:
1. Restore CHANGELOG.md to the BASE bytes: `git checkout -- CHANGELOG.md` is fine for this single file. The rc.3 release-notes step will add the entry centrally, so this CR then no longer touches CHANGELOG.md, and trunk's change does not intersect it.
2. Nothing else changes. Re-run the buildcache tests with -work and record real exit codes.
3. Run `task-board handoff BUG-261002-ot3ea1 --role developer` and END YOUR TURN. Never edit LOGBOOK.md.
