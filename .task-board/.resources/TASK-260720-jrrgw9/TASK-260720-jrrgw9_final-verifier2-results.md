# TASK-260720-jrrgw9 — independent verifier 2 results

Date: 2026-07-29  
Role: tester  
Verdict: development handback

## First actionable failure

The exact required macOS repository gate failed:

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/verifier2/full-gotmp \
go test -count=1 ./...
```

- Real exit: **1**
- Elapsed from log timestamps: **601 seconds**
- Failing package: `github.com/relux-works/curator/cmd/curator`
- Package result: `FAIL .../cmd/curator 600.591s`
- Failure: `panic: test timed out after 10m0s`
- Test active at the deadline: `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`
- All other packages printed `ok`, including `internal/install` at 411.094s and
  `internal/install/atomicity` at 543.719s.

No `-timeout` flag was supplied. The one-file timing rework preserved its focused behavior but did
not make the exact uncached repository gate fit the unchanged package deadline on this run.

Per the verifier ordering, `go test -race -count=1 ./...`, Windows cross-compilation, remote copy,
digest comparison, and native Windows launcher execution were **not run** and are **not claimed**.
The `ssh win` host was untouched.

## Provenance and immutable inputs

- Candidate: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
- Candidate HEAD: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Accepted integrated comparison tree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`
- Immutable conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Accepted matrix SHA-256:
  `3e4e2ee020841a9f45ce11c788f7617b8dd7ec2a64dfcace9fc968c8dbe7e9f2`
- `vectors/build-drivers.json` SHA-256:
  `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
- `vectors/manager-lifecycle.json` SHA-256:
  `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`
- `vectors/external-repository-lifecycle.json` SHA-256:
  `175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072`
- Go: `go version go1.25.5 darwin/arm64`
- Host: macOS 26.5, build 25F71

`rsync -nrc --delete` against the accepted integrated tree identified exactly **23 test files**:
20 candidate-only additions and three modified tests (`cmd/curator/status_test.go`,
`internal/buildcache/conformance_test.go`, and `internal/closure/conformance_test.go`). There was no
product, schema, golden, registry, release-pin, or configuration delta. Baseline and post-runtime
delta listings were byte-identical, and baseline and post-runtime SHA-256 lists for all 23 files
were byte-identical. The verifier did not edit candidate source or tests.

## Focused barrier

The accepted 12-package authoritative consumer barrier ran with `-count=1`, the immutable root, the
task-owned `focused-gotmp`, and the exact accepted matrix filter:

```text
go test -count=1 \
  ./internal/runtimestore ./internal/install ./internal/scopes ./cmd/curator \
  ./internal/buildcache ./internal/buildsource ./internal/buildmeta ./internal/godriver \
  ./internal/skillcheck ./internal/skillspec ./internal/whitelist ./internal/interop \
  -run '^(TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus|TestAuthoritativeDryRunCasesMutateNothingPersistent|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestAuthoritativeGarbageCollectionRootsAreRetained|TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable|TestDriverRejectionClustersMapToStableCuratorErrors|TestFixedEnvironmentAndFiveDirectArgvFormsVector|TestToolchainIdentityVectors|TestValidPackageGraphVectors|TestPortableExecutionPolicyMatchesTheAcceptedVector|TestCacheIdentityMatchesTheAcceptedVector|TestCandidateGoV1SourceAwareContract|TestProtectedCacheHitVector|TestCompilerFreeDryRunMissVector|TestCacheRejectionClustersMapToStableCuratorOutcomes|TestBuildSourceIdentityVectors|TestPortableExecutionPolicyIsTheOnlyAdmittedBuildInput|TestBuildRootContentInContextVector|TestManifestAndFilesystemRejectionVectors|TestSchemaSixMixedScriptAndBuildCommandsVector|TestBuildRootExcludedFromAgentContextVector|TestManagerLifecycleVectors|TestGoldenMarkerObject|TestGoldenFederationSemantics)$'
```

- Real exit: **0**
- Elapsed from log timestamps: **35 seconds**
- All 12 packages printed `ok`.

The patched status/lifecycle/repair barrier then ran:

```text
go test -count=1 ./cmd/curator \
  -run '^(TestStatusReportsCompiledCurrentnessAndFailsCheck|TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable|TestInstallAndUpgradeRepairCorruptCompiledState|TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails|TestDryRunNeverClaimsACompletedCompilerCheck)$'
```

- Real exit: **0**
- Elapsed from log timestamps: **268 seconds**
- Package result: `ok .../cmd/curator 267.207s`

The focused evidence confirms the timing patch did not remove the accepted authoritative,
status/currentness, lifecycle, corrupt/untrusted repair, rollback, or dry-run assertions.

## Gate ledger

| Gate | Real exit | Evidence |
| --- | ---: | --- |
| Candidate HEAD | 0 | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| Exact task delta | 0 | 20 added tests, 3 modified tests |
| Candidate 23-file SHA-256 inventory | 0 | captured before runtime |
| Authoritative-root SHA-256 inventory | 0 | 448 files captured before runtime |
| Go version | 0 | Go 1.25.5 darwin/arm64 |
| Initial disk | 0 | 22,910,500 KB available |
| Initial process barrier, first attempt | **1** | transient external Go PID 81279 |
| Initial process barrier, retry after PID terminal | 0 | no matching Go/test process |
| Focused authoritative barrier | 0 | all 12 packages green |
| Post-authoritative process barrier | 0 | no matching Go/test process |
| Focused status/lifecycle/repair barrier | 0 | `cmd/curator` 267.207s |
| Post-focused process barrier | 0 | no matching Go/test process |
| Required `go test -count=1 ./...` | **1** | `cmd/curator` default 10-minute timeout |
| Immediate post-full process barrier | **1** | unrelated PID 1167 running in TASK-260729-3jmqgl |
| Post-full process barrier after external PID terminal | 0 | no matching Go/test process |
| Post-runtime exact-delta comparison | 0 | byte-identical to baseline |
| Post-runtime candidate digest comparison | 0 | byte-identical to baseline |
| Task-owned GOTMPDIR cleanup | 0 | both trees were 0 KB and removed |
| GOTMPDIR absence verification | 0 | both exact paths absent |
| Final process barrier | 0 | no matching Go/test process |
| Final disk | 0 | 22,762,196 KB available |

The two red process-barrier attempts are recorded as failures, not passes. Both involved external
processes and were retried only after those exact PIDs were terminal. No overlapping verifier Go
gate was started.

## Logs and cleanup

- `focused-authoritative.log` SHA-256:
  `dd5f645e9e559e41f8bed721cd15e87395c098af6b7d021e5c9f0c9a8bff2627`
- `focused-status-lifecycle-repair.log` SHA-256:
  `5b758a1e42e0864f4e7bafaefc7ce7f4d45f68e0b555f50734f8777fc06b1966`
- `go-test-all.log` SHA-256:
  `f6f0f57aaf2fd6551de16e02c18f5a11ad898cd6e5b45a6e9bb7064736c2d799`

Available disk remained above the 8 GiB stop threshold throughout:

| Point | Available KB |
| --- | ---: |
| Initial | 22,910,500 |
| After focused barriers | 22,905,216 |
| After failed full gate | 22,757,048 |
| After cleanup | 22,762,196 |

Both task-owned GOTMPDIR trees measured 0 KB after all descendants terminated. Only the exact
`verifier2/focused-gotmp` and `verifier2/full-gotmp` paths were deleted, and their absence was
verified. Logs and provenance files remain under `verifier2`.

## Development handback

The task remains in development. The first required action is to make the exact uncached
`CURATOR_CONFORMANCE_ROOT=<immutable-root> go test -count=1 ./...` fit the unchanged default
10-minute package deadline. After rework, an independent verifier must rerun the focused barrier,
the exact full suite, the exact full race suite only after the required process/disk checks, and
then the native Windows launcher matrix.
