# Revalidate SNAP-LOCK on current trunk and publish revision 3 (TASK-261016-4zofp0)

Revision 2 was ACCEPTED, but trunk then advanced (snapshot identity verifier; credential_store selector slice 1). The operator ran `task-board worktree converge`: the workspace now sits on current curator main with your revision 2 delta carried over by a clean three-way merge. The board demoted revision 2 to stale (`integration_base_moved`) because trunk also changed `README.md` and `internal/envprofile/managed.go`.

The one semantic interaction to check: trunk added `req.admitCredentialStore(adapter)` inside `Resolve` (after the source/adapter load) and inside `checkIsolatedCredentialStore`. Your revision moved the body of `Resolve` into `resolveVerified(req, verify)`, so after the merge that admission check also runs on the snapshot capture path (`verifyHomeContent`).

Do this and nothing more:
1. Confirm the merged `managed.go` is correct: the admission check runs exactly once per resolve on both `Resolve` and the capture path, and a credential store that is not admitted makes snapshot capture refuse with the same refusal as `Resolve`, before any capture write. If the spec (CIP-0015 / rc.7 leaf) says capture must behave differently, stop and record the contradiction instead of guessing.
2. Add one focused negative test: a not-admitted credential store makes capture refuse with no snapshot written, next to a positive control that captures. Make no other behavior changes.
3. Check that `git diff` against the new Story base contains your revision 2 delta, plus that test, and nothing else.
4. Re-run the focused tests and the full gate (`sh scripts/remote-gate.sh`). Record the exact commands and results.
5. Publish the candidate as a new Change Request revision and hand off for review as the board requires. The rules in rc7-code-common.md (attached) still apply.
