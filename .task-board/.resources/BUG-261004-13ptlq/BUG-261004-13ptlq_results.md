# BUG-261004-13ptlq — unmanage cannot restore its own symlink backup

Built on N3 commit 505e1526 already in HEAD. No branch switches, rebases, or developer commits on the Story branch; the hosted gate creates detached snapshot commits.
Restore plan now carries literal link target text separately from regular-file bytes and permission bits. Saved links use stateread.Readlink and atomicManagedLink, with no copy fallback. Parent checks remain in preflight and managed writer.

## Evidence so far
- Initial CLI fixture attempt: exit 1 due to incorrect assumption that default takeover writes a managed link. Corrected fixture before regression proof.
- RED: GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanageTakeoverRestoresSymlink$' -count=1 -timeout=4m — exit 1; all 4/4 cases reached CLI env unmanage and failed with environment_backup_record_unreadable, backup entry is not a regular file.
- GREEN: GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanage' -count=1 -timeout=4m — exit 0 (before extending parent routes).
- GREEN: GOFLAGS=-work go test ./internal/envprofile -run '^TestUnmanage' -count=1 -timeout=4m — exit 0.
- Mutant: production restore replaced atomicManagedLink with os.WriteFile(full, literalTarget, 0600). GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(UnmanageRestoreSymlinkReplacesDestinationEntry|EnvUnmanageTakeoverRestoresSymlink)$' -count=1 -timeout=4m — exit 1. Internal regression detected changed external destination bytes; CLI round trips rejected regular-file restoration in 4/4 shapes. Original code restored from saved copy and compared equal.
- git diff --check — exit 0; sh -n scripts/remote-gate.sh — exit 0.

## Decisions
- Unix CLI coverage: absolute, relative, dangling and directory link targets. External bytes, inode identity, permission mode and modification time must remain unchanged. Backup link text is checked before and after restore.
- Parent refusal coverage is extended to direct and nested linked parents and requires no native surface or marker writes.
- Gate snapshot commit message was corrected to omit the absolute workspace path, per AGENTS.md public-repository privacy rule.
- Current task brief forbids LOGBOOK.md and CHANGELOG.md edits. Findings are recorded in task notes and this outcome instead; no edits to either file.

Full local checks, narrowed-parent mutant and hosted validation are pending. No previous attached validation is accepted as proof of this changed candidate.

## Frozen candidate local validation
- Parent diagnostic mutant (reject symlinks only at index 0): GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanageSymlinkRestoreRefusesLinkedParent$' -count=1 -timeout=4m — exit 1; nested route refused with wrong diagnostic.
- Admission mutant (both preflight guards permit linked parents below index 0): the same command — exit 1; nested route changed the managed surface before the downstream atomic writer refused. This demonstrates why preflight must check the full parent route before any write. Original implementation restored from saved copy and compared equal.
- GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(Unmanage|EnvUnmanage)' -count=1 -timeout=4m — exit 0, after all mutations were restored and parent coverage extended. Internal package 2.055s; CLI package 50.680s.
- golangci-lint run ./internal/envprofile/... ./cmd/curator/... — exit 0; 0 issues.
- go build -o .temp/BUG-261004-13ptlq/curator ./cmd/curator — exit 0.
- git diff --check — exit 0.
- sh -n scripts/remote-gate.sh — exit 0.
- test -z "$(gofmt -l internal/envprofile/unmanage.go internal/envprofile/unmanage_symlink_unix_test.go cmd/curator/env_unmanage_symlink_unix_test.go)" — exit 0.
- git merge-base --is-ancestor 505e1526 HEAD — exit 0. N3 ancestry proven without rebase, branch switching or commits.
- Retained package TestMain host-GOROOT lock and used GOFLAGS=-work for every targeted test run.

Coverage bound: 4/4 Unix CLI round-trip link shapes; 2/2 linked destination parent routes; 1/1 destination-link replacement regression. These checks cover literal link text, entry type, external bytes and observed inode/mode/mtime preservation, not access-time or syscall tracing. CLI conformance vector test has no local CURATOR_CONFORMANCE_ROOT and skips locally; the hosted full gate supplies its pinned conformance root. Windows ACL recovery is unchanged; these newly added filesystem checks are Unix-only. No old validation evidence was reused.

Hosted full CI gate launched against this frozen candidate; outcome pending.

## Hosted validation in progress
- Command: sh scripts/remote-gate.sh. Its terminal exit code is not yet available.
- Detached snapshot: 4842951c2e3f2f78e4dc4236bb6b4612eaaa018c.
- Evidence: https://github.com/relux-works/curator/actions/runs/37221096954
- A standalone Python comparison of snapshot file blobs against worktree bytes exited 0: 4/4 changed files match. The snapshot commit message contains only the Story ID and base OID; no absolute workspace path.
- Naming, interop conformance, lint and all 3 platform gate self-tests have passed; full test and race job verdicts remain pending. Optional rose-air and candidate-suite jobs are skipped by the workflow.
- Production route under regression: run -> cli.cmdEnvUnmanage -> envprofile.Unmanage -> planUnmanageHomes/readBackupTree -> preflightUnmanagePlan -> applyUnmanagePlan -> atomicManagedLink. CLI setup also drives real profile install and profile use --takeover handlers.

## Additional parent-route finding and candidate revision
The first hosted snapshot passed sh scripts/remote-gate.sh with real exit code 0: all 11/11 required jobs passed, 2/2 optional jobs skipped. Snapshot 4842951c2e3f2f78e4dc4236bb6b4612eaaa018c is now historical evidence, not validation of the revised candidate.

Further review found that a saved directory link could become the parent of a later recorded removal during apply. Existing-parent preflight does not see a link that the same plan has yet to restore. A schema-valid marker with an absent recorded child and a saved parent-directory link reproduces deletion of the external child through that restored link.

RED, before the guard fix:
- GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent$' -count=1 -timeout=4m — exit 1 (first direct-child case). External child was deleted.
- The same command after extending to direct and nested recorded children — exit 1, 2/2 cases deleted the external child.

Fix stays within the parent-route requirement: preflight checks every planned path ancestor against restored links before any native write. Recorded removals also use existing removeManagedEntry to recheck their parent route at the removal boundary. There is no fallback that follows or copies a link target. The candidate must be validated again because source and tests changed; no prior hosted result is reused for it.

## Revised candidate local validation and review preparation
- GREEN: GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(Unmanage|EnvUnmanage)' -count=1 -timeout=4m — exit 0 after planned-parent fix (internal 0.590s, CLI 44.374s).
- Narrowed planned-parent mutant: check only path.Dir(rel), omitting higher ancestors. GOFLAGS=-work go test ./cmd/curator -run '^TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent$' -count=1 -timeout=4m — exit 1; nested child caused native surface mutation before the guarded removal refused. Direct child still refused. Restored source from saved revised copy.
- Repeated follow-link mutant on revised source: replace only production applyUnmanagePlan atomicManagedLink call with os.WriteFile(destination, literalTarget, 0600). GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(UnmanageRestoreSymlinkReplacesDestinationEntry|EnvUnmanageTakeoverRestoresSymlink)$' -count=1 -timeout=4m — exit 1; destination target bytes changed in internal test, 4/4 CLI shapes restored incorrect regular-file type.
- cmp .temp/BUG-261004-13ptlq/unmanage.go.revised internal/envprofile/unmanage.go — exit 0 after restoring the production implementation.
- GREEN after all revised-source mutants restored: GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(Unmanage|EnvUnmanage)' -count=1 -timeout=4m — exit 0 (internal 0.505s, CLI 45.824s).
- golangci-lint run ./internal/envprofile/... ./cmd/curator/... — exit 0; 0 issues.
- go build -o .temp/BUG-261004-13ptlq/curator ./cmd/curator — exit 0.
- git diff --check — exit 0.
- sh -n scripts/remote-gate.sh — exit 0.
- test -z "$(gofmt -l internal/envprofile/unmanage.go internal/envprofile/unmanage_symlink_unix_test.go cmd/curator/env_unmanage_symlink_unix_test.go)" — exit 0.

Current regression coverage: 4/4 full CLI round-trip link shapes; 2/2 existing linked-parent routes; 2/2 planned linked-parent routes (direct child and deeper descendant); 1/1 replacement of a destination link without modifying its external target. Existing N3 mode restoration cases also reran in the targeted suite. New planned-route guards reject unsafe plans before any surface/marker write, while a saved directory-target link with no descendant plan remains restorable.

Candidate is frozen with four changed files. A second sh scripts/remote-gate.sh is running against it because source and tests changed after the first hosted snapshot. Handoff will follow an updated outcome containing this candidate's hosted result. No LOGBOOK.md/CHANGELOG.md edits, branch operations, or Story-branch commits.

## Revised hosted snapshot
- Detached snapshot: 4dad4dee869f1180380bb4387f51d83f7a9dae00.
- Hosted run: https://github.com/relux-works/curator/actions/runs/37225068846
- Standalone Python blob comparison — exit 0, 4/4 revised worktree files match this snapshot.
- sh scripts/remote-gate.sh terminal exit code remains pending. Do not treat the earlier snapshot's exit 0 as validation of this one.

## Review handoff packet — revised candidate
The revised standalone hosted command sh scripts/remote-gate.sh exited 0. All 11/11 required jobs passed: naming, lint, interop conformance, 3 platform gate self-tests, tests on Linux/macOS/Windows, and race suites on Linux/macOS. The 2 optional workflow jobs (candidate suite and rose-air) were skipped, not counted as passing.

Hosted evidence: https://github.com/relux-works/curator/actions/runs/37225068846
Detached tested snapshot: 4dad4dee869f1180380bb4387f51d83f7a9dae00; 4/4 changed file blobs matched the frozen worktree. The earlier successful snapshot 4842951c2e3f2f78e4dc4236bb6b4612eaaa018c remains historical only. Current local test/lint/build exit codes are all 0; current follow-link and narrowed-planned-parent mutants exited 1 as expected and were removed before green validation.

Acceptance evidence:
1. Saved link entry type and literal target survive CLI takeover and restoration, using the existing atomic managed link writer. A destination link is replaced without changing its external target bytes or observed inode/mode/mtime. Regular-file mode preservation from N3 reran green.
2. Existing parent links are refused before native writes in both direct/nested cases. Planned restored parent links are also refused before any native surface/marker mutation, and removals recheck the parent route using removeManagedEntry. Both direct/deeper stale child cases are green. Narrowing the planned check to immediate parents failed the deeper case.
3. Full production run() CLI round trips cover 4/4 absolute, relative, dangling and directory-target foreign links: install, profile use --takeover, env unmanage --restore-backups. Original link text and retained backup text match; external file bytes, identity, mode and mtime remain unchanged. The restore-following mutant changed external destination bytes and failed regression checks.
4. N3 505e1526 is an ancestor of the unchanged Story HEAD (git merge-base --is-ancestor exited 0). No branch switching, rebasing, or Story-branch developer commits.

Scope: internal/envprofile/unmanage.go, 2 task regression test files, and a one-line gate snapshot message correction required to avoid publishing absolute workspace paths. No LOGBOOK.md/CHANGELOG.md edits; findings are in task notes and this outcome per the current brief. New filesystem regressions are Unix-only; Windows compilation and the existing Windows suite passed, but Windows ACL recovery is unchanged. Access-time and syscall tracing are not attested by the external-file checks.

Handoff invokes the configured board-managed validation against its captured candidate and owns the durable Change Request validation receipt. This standalone hosted run is not substituted for that receipt. Source remains uncommitted in the Story worktree, ready for the developer role handoff to review.
