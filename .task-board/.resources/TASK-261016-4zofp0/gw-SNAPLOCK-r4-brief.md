# Re-apply SNAP-LOCK onto current trunk as revision 4 (TASK-261016-4zofp0)

Revision 3 was ready for review, but trunk then advanced twice: MUSE148 changed `README.md`, and the L15-1 policy projection changed `README.md` and `internal/envprofile/managed.go`. The operator ran `task-board worktree converge`. The workspace now sits on current curator main. All your files were carried over unchanged, except `internal/envprofile/managed.go`, which had a content conflict.

The operator restored that one file to trunk's version. Your revision 3 change to it is attached as `SNAPLOCK-rev3-managed.patch` (a diff against the old base 782823d8).

The conflict is semantic, not textual. Trunk moved the at-or-above and isolation computation into `req.resolveCredentialPolicy(adapter)` (`internal/envprofile/credential_policy.go`). That function calls `req.detected(adapter)`, which observes the vendor tool, and it caches its result per profile and environment. Your revision passes an already observed `release` into `assembleHomeObserved`, and `verifyHomeContent` passes "unknown" so that content verification never runs a vendor tool.

Do this:
1. Re-apply your `managed.go` changes onto trunk's file: `assembleHomeObserved`, `verifyHomeContent`/`verifyHomeObserved`, `resolveVerified`, and anything else in the patch.
2. Thread the observed release through the shared credential-policy path. Capture and status must never call `req.detected`, directly or through `resolveCredentialPolicy`. The policy cache must not return a policy computed under a different release observation to an operation that observed another one. Keep a single policy derivation path: no second resolver, no copy of the at-or-above logic.
3. Your vendor-probe test (`TestMaterializeCaptureAndStatusNeverProbeVendor`) must still pass. Add a narrowing plant that makes the policy path probe the vendor again, and show that a named test fails on it.
4. Re-run the credential-store admission test, the focused suites (both yours and trunk's credential-policy tests), all your narrowing plants, and the full gate (`sh scripts/remote-gate.sh`). Record the exact commands and exits.
5. Refresh the candidate as revision 4 and hand off for review as the board requires. The rules in rc7-code-common.md (attached) still apply. Make no other behavior changes.
