# TASK-260917-2vapkz — integration precondition check (RUN-260930-73aac2): NOT CONFIRMED

This run got two instructions that contradict each other:
- Integration assignment: land accepted CR-TASK-260917-2vapkz-3 revision 3 as-is. Keep the status at integrating.
- 2vapkz-rework-3.md: rev3 fails the Python line (csk 4a88aa0e pins the digest of conformance/skillfile-sources-v1/index.json). Split the skillfile-sources half out to TASK-260930-3ny11n before landing.

Worktree state (not modified by this run): HEAD 4ad8042, with rev3 staged (148 files, +10950/-317 vs 4ad8042b).
It still carries the skillfile-sources half. `git diff --stat HEAD -- conformance/skillfile-sources-v1 schemas/skillfile-sources-v1 protocol/skillfile-sources.md`
shows 17 files, +891/-14.

The decision is held:
- If I land rev3 as accepted, main gets a change we already know breaks CI.
- If I apply rework-3 here, the tree no longer matches the accepted, reviewed revision, so the runner would be landing an unreviewed candidate.

So this run did not change the tree and did not call handoff. No landing was triggered.

Orchestrator decision needed. Pick one:
- (a) Route rework-3 as a normal development run, then review it as rev4, then integrate. Recommended.
- (b) Explicitly authorize applying rework-3 inside this integration run.
