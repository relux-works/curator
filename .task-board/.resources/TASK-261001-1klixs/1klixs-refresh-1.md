# THE ONLY CURRENT INSTRUCTION — TASK-261001-1klixs base refresh → revision 2 (orchestrator brief, binding)

Revision 1 (the second-operator guide) was ACCEPTED, but landing refused with integration_base_moved: trunk (now bd126a9a) changed README.md (20ao7p added a docs link), and rev1 also changes README.md. The acceptance is released. This is a base refresh only.

1. Run `task-board worktree refresh-candidate TASK-261001-1klixs`.
   - On a conflict, follow `--replay-resolutions`; never hand-commit the replay worktree.
   - README.md keeps BOTH the trunk's new links and the guide link.
2. Prove the refreshed candidate differs from rev1 only by that README combination. Report per-file diff stats in the results.
3. Check that every relative link in the guide and the README resolves. Record the real exit code.
4. Add a "Revision 2 (refresh)" section to the results. Then run `task-board handoff TASK-261001-1klixs --role developer` and END YOUR TURN.

No content change. Never spell any employer name.
