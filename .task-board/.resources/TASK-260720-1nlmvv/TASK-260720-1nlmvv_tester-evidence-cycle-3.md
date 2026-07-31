# TASK-260720-1nlmvv — cycle 3 focused tester evidence

Role: tester  
Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`  
Platform: darwin/arm64, Go 1.25.5

## Verdict

The focused cycle-3 criterion is proven: cache publication is reversible across
post-publication planning, preparation, and rolled-back commit failures, and
the real CLI install and upgrade paths restore the displaced cache entry and
preserve the prior installation when the durable commit fails.

The implementation correctly declines reversal when a durable in-flight
transaction still references the newly published key or when target preimages
are no longer restored. In that state recovery owns forward progress, and
`Result.BuildCacheRetained` plus the operator warning report the retained entry.

## Tester-owned additions

- `internal/install/commit_test.go`
  - Added a journal-planning failure after publication through the existing
    `NewTransactionID` seam.
  - The case proves the corrupt predecessor is restored, the new publication is
    withdrawn, and installed state is byte-identical when no journal exists.
- `internal/buildcache/cache_test.go`
  - Added fail-closed coverage for unsupported cache protection.
  - Added fail-closed coverage for malformed cache keys.
  - Focused statement coverage of `buildcache.(*Store).Revert` is 85.0%.

No product code was edited.

## Production-path inspection

`runCommit` registers compensation immediately after `publishWinners`, including
partial-publication errors. Its deferred compensation therefore brackets target
staging, consumer staging, plan validation, journal planning, journal prepare,
and journal commit. It rechecks every journaled target preimage and all live
journal build-key references while holding the manager-home lock. Reversal runs
newest-first through the real protected cache store.

The real CLI E2E test drives both `install` and `upgrade`: it installs compiled
state, corrupts the live artifact, proves status reports
`corrupt-build-cache`, lets repair publish, forces the durable target commit to
fail with a verified unwritable context store, and then proves the exact corrupt
artifact and marker are restored. A later ordinary invocation repairs
successfully and `status --check` returns zero.

The other three cycle-3 blockers are also pinned through production-reachable
tests:

- build-boundary commit failures enter bounded, control-stripped,
  absolute-path-redacted `Result.Errors`;
- toolchain-refusal rows contain `driver=go-v1` and the validated build-source
  identity in both human and JSON output without inventing target/key/artifact;
- real concurrent cache corruption, removal, equally-valid replacement, and
  protection loss become `build-state-changed`, and the check verdict fails.

## Standalone validation evidence

Every command below ran directly without `tee` or a pipeline.

| Command | Exit | Evidence |
|---|---:|---|
| `go test ./internal/buildcache -run '^TestRevertRestoresExactlyWhatAPublicationDisplaced$' -count=1` | 0 | pass, 0.386s |
| `go test ./internal/buildcache -run '^TestRevertFailsClosed$' -count=1` | 0 | pass before tester additions |
| `go test ./internal/buildcache -run '^TestRevertFailsClosed$' -count=1` | 0 | pass after tester additions, 0.518s |
| `go test ./internal/install -run '^TestAFailedCommitRestoresTheBuildCacheItReplaced$' -count=1 -timeout 10m` | 0 | pass, including new journal-planning case, 1.285s |
| `go test ./internal/install -run '^TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced$' -count=1 -timeout 10m` | 0 | pass, 2.002s |
| `go test ./internal/install -run '^TestAnInFlightTransactionKeepsThePublishedCacheEntry$' -count=1 -timeout 10m` | 0 | pass, 0.436s |
| `go test ./cmd/curator -run '^TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails$' -count=1 -timeout 30m` | 0 | pass, 96.329s |
| `go test ./internal/install -run '^TestBuildPublicationFailuresAreRedactedInTheResult$' -count=1 -timeout 10m` | 0 | pass, 0.774s |
| `go test ./internal/install -run '^TestToolchainFailurePlansAnInventoryOfEveryActiveCommand$' -count=1 -timeout 10m` | 0 | pass, 0.594s |
| `go test ./cmd/curator -run '^TestStatusReportsAnUnusableToolchainPerCompiledCommand$' -count=1 -timeout 15m` | 0 | pass, 14.321s |
| `go test ./cmd/curator -run '^TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck$' -count=1 -timeout 20m` | 0 | pass, 74.142s |
| focused buildcache coverage command | 0 | package 40.6%; affected `Revert` function 85.0% |
| `gofmt -l internal/buildcache/cache_test.go internal/install/commit_test.go` | 0 | no output |
| `go vet ./internal/buildcache ./internal/install ./cmd/curator` | 0 | no output |
| `git diff --check` | 0 | no output |

`command -v golangci-lint && golangci-lint version` exited 1 because this tester
environment has no `golangci-lint` executable on `PATH`. No tester lint pass is
claimed. The unchanged cycle-3 product tree already has producer evidence for
the pinned v2.4.0 full-tree lint gate at exit 0.

Per the focused directive, repository-wide tests and the deferred
`cmd/curator` race gate were not rerun. The authoritative cycle-3 full-suite and
`internal/install` race results remain recorded in the producer evidence.

## Lifecycle gate

The first `task-board handoff TASK-260720-1nlmvv --role tester` attempt exited
1 because tester-role checklist items 18–20 had not yet been checked. The
underlying evidence already existed: the modified tests were green, affected
`Revert` coverage was 85.0%, and this task-scoped outcome was attached. The
tester therefore checked those three required evidence items before retrying
the handoff.
