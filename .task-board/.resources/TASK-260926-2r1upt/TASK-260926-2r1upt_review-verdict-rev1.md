# TASK-260926-2r1upt review verdict — CR rev1: ACCEPTED

Candidate tree d3b94b63 (base bd3c0f43); worktree bytes == candidate. Only internal/install/draftsources_test.go changes (+172). No CHANGELOG/LOGBOOK/production change.

## Independent verification (disposable clone /tmp/rv2r, bash, set -o pipefail, precompiled test binaries)
- New test TestDraftFreshMachineTransitiveReplaySourceGuards (3 subtests) PASS on unmutated tree, exit 0. Drives production entry install.Project (dry-run fresh machine) -> declaredDependencyReplaySources.
- gofmt -l clean; go vet ./internal/install exit 0.

## Per-guard mutants (draftsources.go)
| Mutant | Before (old install draftsources tests exit / crossconformance Playbook exit) | After (new row) |
|---|---|---|
| M1 `if false && directory != target.Package.Directory` | 0 / 0 (survives) | exit 1, directory-mismatch FAIL only |
| M2 `if false && resolution.Identity != target.Package.Repository` | 0 / 0 (survives) | exit 1, repository-identity-mismatch FAIL only |
| M3 success branch: Alias `requirementName` instead of `consumer:requirement` | 0 / 0 (survives) | exit 1, matching-declaration-replays-dependency FAIL only |

Each row kills exactly its own mutant.

## Stated bounds (residuals, not blocking)
- "Every other branch" was realised as the success/insert branch. The skip `continue`s (local-snapshot w/o selection, non-Git, legacy refs, missing GitRepos, source_snapshot_unavailable, non-network target), first-consumer-wins dedupe, invalid requirer manifest error, and policy/ResolveRepositoryEndpoints error propagation have no dedicated rows.
- Full internal/install gate not rerun by reviewer (-run TestDraft exceeded 10 min under host load); hosted gate is arbiter.
