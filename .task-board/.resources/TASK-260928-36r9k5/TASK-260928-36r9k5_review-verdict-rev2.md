# TASK-260928-36r9k5 review verdict — CR rev2 — ACCEPTED
Reviewer: claude-opus-5-5 (low). Candidate: base 213a53e5, tree 438bacad (worktree write-tree == 438bacad), 34 paths.

## A. 4pv4au re-apply fidelity
Per-path compare of the sorted +/- lines of `git diff 6bd98d49 c7e73fa2` against `git diff 213a53e5 438bacad`. 15/25 paths are identical. Where they differ:
- environments_test.go / environments_conformance_test.go / environments.go: the require_source_signers knob has already landed on trunk (213a53e5 has it in the test literals), so the carrier correctly leaves those paths out. security_posture.go still reads Env.RequireSourceSigners and its absence default (hardened → true), so the posture still takes effect.
- config.go: the same LockableKeys entry was moved to a new position; nothing is lost. envprofile.go: gofmt realignment against trunk's fields.
- main.go / security_posture*_test.go: the 1sapuy additions, plus test fixture adjustments for trunk (require_source_signers:false set explicitly where the test targets the MCP contradiction).
- The warning goes out once, on stderr only, from loadConfig. runEnforcedShim calls config.Load directly and never prints it (F1 still holds).
Nothing was weakened.

## B. 1sapuy (full review)
- Both install.Project and install.Global use resolveRegistryEvidence. SnapshotCheck.Unreachable and Resolution.Unreachable feed registryResolution. ArtifactsWithoutEvidence lists every node with no attestation (name@commit).
- Severity comes from cfg.SecurityPostureHardened(), the part-A API; nothing re-derives it. Hardened means error: Status failed and a return before the narrow boundaries and before staging. Permissive means a warning and the operation continues.
- The notice goes to stderr through printResult (`GATE NOTICE registry_unreachable_during_install`) and names the registries and artifacts. OperationStatus suppresses it, so status stays read-only.
- Mutants, run in a disposable git-archive clone (the worktree was untouched):
  - M1 hardened downgraded to warn (install.go `severity = "error"` removed) → TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy FAILS (exit 1). KILLED.
  - M2 notice removed (Fprintf to io.Discard) → both RegistryOutage tests FAIL. KILLED.
- Stated bound (not blocking): the hardened test checks the exit code and stderr but does not assert "no materialization under the project". The no-write property rests on code order: the return comes before stageProjectTargets and stageGlobalTargets. Persistent registry snapshot and rollback state (MigrateSnapshotStates / CheckSnapshotsWithPolicyDetailed) is still written before the refusal. It is manager-owned anti-rollback state, not an install write, but a follow-up should add a no-write assertion.

## C. Rev2 test and golden updates
- global_adopt, global_lock_publication, profile_delta_confirmation and the two goldens now expect the permissive warning; none ignores stderr wholesale.
- The ledger diff is only +security-posture/vectors 17 plus 3 removed gap rows (the valid-security-posture-* schema cases, which now pass). No trunk-removed row was re-added. No CHANGELOG, LOGBOOK or stray files.

## Runs (zsh, pipefail, clone of 438bacad)
- `go test ./internal/config -count=1` → ok, exit 0
- `go test ./cmd/curator -count=1 -run 'SecurityPosture|RegistryOutage|GlobalAdopt|GlobalAdd|ProfileUpdate|Golden|StatusJSON|Enforced'` → ok (201s), exit 0
- Not rerun by me: internal/install, internal/scriptworker, internal/envprofile. For those I accepted the green hosted gate on rev2 (run cited by the orchestrator).
