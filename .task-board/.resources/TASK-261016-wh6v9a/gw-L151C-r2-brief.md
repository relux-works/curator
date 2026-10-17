# Revalidate L151C on current trunk and publish revision 2 (TASK-261016-wh6v9a)

Revision 1 was ACCEPTED. Then SNAP-LOCK landed (67e3bcda) and changed `internal/envprofile/credential_policy.go`, `internal/envprofile/managed.go`, `README.md` and `scripts/remote-gate.sh`, which your revision also changes. The operator ran `task-board worktree converge`. The workspace now sits on current curator main, and your delta was carried over by a clean three-way merge.

Only `scripts/remote-gate.sh` was set aside. Trunk already carries the same signed-snapshot fix (`git commit-tree -S` plus a `%G?` check), so do not re-add your copy.

The semantic interaction to check: SNAP-LOCK made capture and status verify content from file evidence only. The observed release is threaded through the shared credential-policy path, so capture and status never reach `req.detected` (the vendor tool observation), and the policy cache never serves a result computed under a different release observation. Your Muse file policy, shared/isolated admission and v4 fragment composition now run on that same path.

Do this and nothing more:
1. Confirm that your new Muse policy code does not observe the vendor tool on the capture or status path, directly or through the cache. If it does, route it through the observed release as SNAP-LOCK does, without a second policy derivation path.
2. Run SNAP-LOCK's vendor-probe test (`TestMaterializeCaptureAndStatusNeverProbeVendor`) and its narrowing plants together with yours on the combined tree. Add one Muse capture case to the no-probe test if the existing cases do not already cover the Muse file policy.
3. Check that `git diff` against the new Story base contains your revision 1 delta (minus `remote-gate.sh`), plus anything step 1 or 2 required, and nothing else.
4. Re-run the focused tests, all your narrowing plants and the full gate (`sh scripts/remote-gate.sh`). Record the exact commands and exits.
5. Refresh the candidate as revision 2 and hand off for review as the board requires. The rules in rc7-code-common.md (attached) still apply.

## Revision 3 after review (2026-10-17)

Review B of revision 2 found no defect in the Muse/SNAP-LOCK implementation (all eight swept product surfaces held, 28 narrowing controls passed). It requested changes only because the hosted Ubuntu Test job failed with Go probe timeouts (`TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only`, `TestReviewProbeLegacyLaneIdentityInvalid`).

- Cause: those runs fell inside the self-hosted host overload from the Landlock probe recursion bug, which your old-base test run itself re-triggered. The fix is on main (6e85a91, BUG-261017-ygfdal); your Story workspace has been converged onto it with your changes carried (LOGBOOK.md merged cleanly). The second hosted attempt was cancelled by the orchestrator to stop the recursion.
- Do: refresh the candidate as revision 3 on the converged workspace, move your LOGBOOK.md entries to the top of the log (newest first, as the file header says), rerun the full remote gate, and hand off. Do not change product code unless a failure reproduces on the fresh run; if one does, investigate it as a real defect.
