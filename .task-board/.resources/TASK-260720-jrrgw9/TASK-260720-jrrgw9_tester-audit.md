# TASK-260720-jrrgw9 tester audit

Date: 2026-07-29  
Role: tester  
Verdict: return to development; the narrow candidate gates pass, but the task acceptance gates are not established.

## Scope and provenance

- Candidate worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
- Authoritative conformance root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- The precondition named `build-drivers.json` at the root. That exact path is absent. The published file is `vectors/build-drivers.json`.
- SHA-256 of the published file: `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`, matching the mandatory precondition.
- Candidate HEAD: `17804ce` (`v0.12.5`, rc.3 pin); no staging, commit, publication, or release pin mutation was performed.

The preserved four-file test-only addition is:

| File | SHA-256 |
| --- | --- |
| `internal/godriver/builddriver_rejection_conformance_test.go` | `b5f125b8851e426e82200387f7275f845363a72a53a3c442828b4c76e270c8c7` |
| `internal/skillspec/builddriver_conformance_test.go` | `4ec82f2f29d6d45f10085212738621cf86798c16961bc2c725bd1c35e21e9e98` |
| `internal/skillspec/builddriver_conformance_unix_test.go` | `7b80d69a124c59fdf726d5a4a88c204dad2b68aa7c8bdcb1cd4a77d27526a668` |
| `internal/skillspec/builddriver_conformance_other_test.go` | `c23ce292af74140a2e49271d60f2a8c593b097fd543a0b0ad89235554aac2d4d` |

All four remain untracked test files. No product file was edited by this tester.

## Static vector mapping

The authoritative build-driver publication contains 8 positive cases, 77 rejection cases, 12 toolchain cases, 10 build-source cases, and 5 fixed argv forms.

Rejection boundary counts:

| Boundary | Published | Candidate consumer |
| --- | ---: | --- |
| manifest | 8 | `TestManifestAndFilesystemRejectionVectors` |
| filesystem | 14 | `TestManifestAndFilesystemRejectionVectors` |
| module | 2 | `TestDriverRejectionClustersMapToStableCuratorErrors` |
| dependency-graph | 14 | `TestDriverRejectionClustersMapToStableCuratorErrors` |
| compiler-directive | 3 | `TestDriverRejectionClustersMapToStableCuratorErrors` |
| toolchain | 5 | `TestDriverRejectionClustersMapToStableCuratorErrors` |
| process | 11 | `TestDriverRejectionClustersMapToStableCuratorErrors` |
| cache | 16 | existing `internal/buildcache` consumers; execution forbidden in this audit |
| context | 2 | existing `internal/skillcheck` / `internal/whitelist` consumers; execution forbidden |
| execution-policy | 2 | existing `internal/buildmeta` / godriver consumers; only the godriver portion was permitted |

The four additions consume authoritative JSON at runtime and do not embed a private copy of published case names as expected outputs. The manifest/filesystem test asserts stable `verr.Error` paths. The godriver test checks rejection, no reuse/artifact execution, stable Curator diagnostic codes, and prevents package compiler execution at the relevant seam. Its explicitly documented Curator/protocol code equivalences remain visible rather than being reported as exact portable-code matches.

The positive package graph, toolchain identity, fixed environment, argv, and execution-policy consumers were found in `internal/godriver` and passed the focused gate. Schema-6 mixed commands passed in `internal/skillspec`. Cache, context, marker, build-source, and lifecycle consumers exist statically in other packages, but the mandatory allowlist did not authorize their execution.

The authoritative `manager-lifecycle.json` contains:

- 2 launcher cases
- 3 bootstrap cases
- 3 upgrade cases
- 2 dry-run cases

`internal/interop.TestManagerLifecycleVectors` consumes this candidate but only checks publication shape and qualitative values. `internal/runtimestore.TestCandidateManagerLauncherContract` consumes launcher cases and inspects generated shim content. Real argument/exit-status process launches are in `internal/runtimestore` tests; bootstrap/upgrade/dry-run execution is in `cmd/curator` and `internal/install`; project concurrency, rollback, and recovery are in `internal/install`; GC execution is in `internal/scopes` and `cmd/curator`. Those packages and lifecycle/race aggregates were expressly forbidden.

Static inspection found relevant scenario tests, including:

- `TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly`
- `TestCacheHitPerformsNoSourceAwareGoCommand`
- `TestSecondBuildFailurePreservesPriorInstallationAndLiveCache`
- `TestCorruptCacheEntryIsRebuiltAndNeverReused`
- `TestUntrustedCacheEntryIsRebuiltAndNeverReused`
- `TestConcurrentProjectInstallsPreserveBothConsumers`
- `TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets`
- `TestRecoveryCompletesBeforeAnyNewMutation`
- project/global install and dry-run tests
- GC serialization and retention tests

Static presence is not pass evidence, especially for the task-required race scenarios.

## Exact executable gates

Every gate below ran as a standalone process without `tee`, cache clearing, timeout changes, or a pipe.

| Command | Exit | Elapsed/result |
| --- | ---: | --- |
| `gofmt -l internal/godriver/builddriver_rejection_conformance_test.go internal/skillspec/builddriver_conformance_test.go internal/skillspec/builddriver_conformance_unix_test.go internal/skillspec/builddriver_conformance_other_test.go` | 0 | no output; all four formatted |
| `git diff --check` | 0 | no whitespace errors |
| `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./internal/godriver -run '^(13 exact build-driver conformance/rejection tests)$'` | 0 | package reports `11.806s` |
| `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./internal/skillspec -run '^(TestManifestAndFilesystemRejectionVectors\|TestSchemaSixMixedScriptAndBuildCommandsVector)$'` | 0 | package reports `0.393s` |
| `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./internal/interop -run '^TestManagerLifecycleVectors$'` | 0 | package reports `0.385s` |
| `go vet ./internal/godriver ./internal/skillspec` | 0 | no diagnostics |

Preflight diagnostics:

- `shasum` against the literal root-level path from the precondition exited 1 because that path does not exist. Re-running against `vectors/build-drivers.json` exited 0 and matched the required digest.
- PID 13051 was absent (`ps -p` exit 1).
- No `go`, compiler, linker, `go test`, or `cmd/curator` process was present before validation (`pgrep` exit 1); none remained afterward (`pgrep` exit 1).

One static `jq` inventory attempt exited 5 because of expression precedence. The corrected read exited 0 and reported `8 77 12 10 5`. This was not a validation gate.

## Disk and process barrier

| Point | Available KB | Candidate KB | Conformance root KB |
| --- | ---: | ---: | ---: |
| before | 5,083,364 | 3,608 | 2,068 |
| after | 4,490,284 | 3,608 | 2,068 |

Available filesystem space decreased by 593,080 KB during the audit. The candidate and conformance-root sizes were unchanged. No cleanup was attempted because cache clearing and destructive cleanup were forbidden. The filesystem reports 100% capacity by rounded display and only about 4.28 GiB available after validation.

## Immutable fixture observations

- The four additions do not alter any schema case, conformance expected file, script fixture, registry object, or golden fixture.
- `internal/interop/golden_test.go` is byte-identical to the accepted integrated source worktree used for comparison.
- `internal/skillspec/parse.go` is also byte-identical to that source worktree.
- Registry and fixture SHA-256 values were inventoried read-only. The conformance root was not mutated.
- A broad read-only directory comparison showed other earlier conformance test additions in this task worktree beyond the four preserved files. This does not change the four-file gate result, but means the worktree as a whole is not literally a four-file delta relative to the accepted integrated worktree.

## Acceptance gaps and proposed ownership

Do not route this item to review from the narrow green gates alone.

1. The task explicitly requires `go test ./...` and `go test -race ./...` against the authoritative suite. Both were forbidden and were not run.
2. The required project/global/hybrid/dependency-closure/cache repair/rollback/recovery/GC/real-launch scenarios were not executable under this allowlist.
3. The task explicitly requires concurrent project success and rollback under `go test -race`; those tests were found statically but no race command was authorized.
4. Coverage near 80% was not measured; no coverage command was authorized.
5. Lint evidence is limited to green `go vet` for `internal/godriver` and `internal/skillspec`; no repository-wide lint was authorized.
6. This Darwin audit cannot execute Windows launch behavior. The Windows shim contract was inspected statically only.

Recommended development/test ownership:

- `internal/runtimestore`: bind the real launcher execution tests directly to both authoritative launcher cases, or add a candidate-consuming executable table that proves arguments, inherited PATH roles, and exact exit status.
- `cmd/curator` + `internal/install`: candidate-consuming bootstrap, selected/all/global upgrade, and project/global dry-run tests; keep the existing no-mutation and closure assertions.
- `internal/install`: run the existing second-build failure, corrupt/untrusted rebuild, cache-hit, recovery, concurrent-project success, and rollback scenarios, with the two concurrency scenarios under race.
- `internal/scopes` + `cmd/curator`: execute GC retention/pruning/serialization coverage.
- An independent verifier with an amended allowlist and adequate disk should run the required full and full-race gates. The present tester must not broaden this audit unilaterally.

This is ordinary validation/rework, not an external platform or product blocker.
