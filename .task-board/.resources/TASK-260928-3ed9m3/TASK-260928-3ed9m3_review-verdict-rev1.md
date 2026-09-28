# TASK-260928-3ed9m3 review verdict rev1 — ACCEPTED

Reviewer: claude-opus-5-5 low. Candidate: base 97e85642, tree 83a0c43a (worktree bytes = candidate; 7 paths, no stray files).
Spec: curator-spec main 4ad8042 environments.md §9.4 new paragraph ("MUST install the immutable store entries ... and publish the extended profile lock before attempting any in-place surface materialization ... MUST keep the published extended lock ...").

## Code order (internal/envprofile/global.go globalSkillOperation)
resolve -> audit/store -> ensureResolvedEntries -> op.publish(lock) -> preflightManagedProfile -> preflightInPlaceScope -> materializeScopeWithNativeHome -> syncManagedProfile. On conflict returns without touching lock. No --takeover on global ops (cmd/curator/main.go); carrier set unchanged. New reads via stateread.

## Independent reruns (zsh, set -o pipefail)
- go test ./cmd/curator -run '^TestGlobal(Add|Install)' -count=1 — exit 0 (61s)
- go test ./internal/envprofile -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestUse|TestSync|TestRepair)' -count=1 — exit 0
- go vet ./cmd/curator ./internal/envprofile — exit 0

## Own mutants (restored by cp; cmp OK)
- M1 republish oldLock on in-place preflight conflict: TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers FAIL, exit 1 — KILLED.
- M3 move lock publish after materialization: TestGlobalAddPublishesProfileLockBeforeNativeMaterialization + AddConflict FAIL, exit 1 — KILLED.

## Residuals (non-blocking)
- R1: both mutants SURVIVE the global install rows (the default-profile migration already puts new-skill in the lock, so install rows do not discriminate order/rollback). Global install ordering is shared code with add (same globalSkillOperation), killed there; install-specific discrimination is a bound.
- R2: spec "restore in-place surfaces already written by that invocation" is satisfied by preflight-before-write under the manager lock; a mid-materialize failure after preflight (race / IO) is not restored — bound.
- R3: scope expansion: native in-place `.agent-context/skills` surface recorded for all profile use/sync, `profile sync` now also re-materializes managed homes, `global add` refuses --source/--branch (spec §9.4 tag/revision only). Consistent with §9.4; flag for release notes.
- R4: rows must move to the vector driver (environments-global-lock-publication.json) when SPEC_PIN moves past 4ad8042.
- Hosted gate: green per orchestrator note; not rerun here.
