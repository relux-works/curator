# TASK-260720-jrrgw9 — rc.5 conformance rework, developer report

Date: 2026-07-29
Role: developer
Handoff: ready for review (test-only change; no product file was edited)

## Provenance

- Candidate worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
- Candidate HEAD: `17804ce` (`Pin landed rc.3 protocol`) — unchanged; nothing staged, committed, published, or repinned
- Authoritative root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- `vectors/build-drivers.json` SHA-256 `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea` (verified before and after; unchanged)

### New test-only files

| File | SHA-256 |
| --- | --- |
| `internal/runtimestore/launcher_conformance_test.go` | `e35cf8580725c53a47a30c070dc2c3b8a71de1fa9a3d705948ffa1915f73cca2` |
| `cmd/curator/lifecycle_conformance_test.go` | `5e83281171c62dcedaef3774686f0526849d556e8b9692e03d39f47bb18658a2` |
| `internal/install/dryrun_conformance_test.go` | `975c9504b91281e132ea5c52db7cff4ea16978b46ff17bab903a398d610b113a` |
| `internal/install/cache_conformance_test.go` | `cf1e825a9ad7a45ee27377f7255c63a77374634ecc4437c1011b50d45b971170` |
| `internal/scopes/gc_conformance_test.go` | `05802cc58533c8611ebd6d5a2ab25c3b101ab52da6f99e2a788f9532b5797d64` |

### Preserved four-file addition (byte-identical to the tester audit)

| File | SHA-256 |
| --- | --- |
| `internal/godriver/builddriver_rejection_conformance_test.go` | `b5f125b8851e426e82200387f7275f845363a72a53a3c442828b4c76e270c8c7` |
| `internal/skillspec/builddriver_conformance_test.go` | `4ec82f2f29d6d45f10085212738621cf86798c16961bc2c725bd1c35e21e9e98` |
| `internal/skillspec/builddriver_conformance_unix_test.go` | `7b80d69a124c59fdf726d5a4a88c204dad2b68aa7c8bdcb1cd4a77d27526a668` |
| `internal/skillspec/builddriver_conformance_other_test.go` | `c23ce292af74140a2e49271d60f2a8c593b097fd543a0b0ad89235554aac2d4d` |

`internal/skillspec/types.go` still declares `SupportedSchemaVersions = {1..6}`. No schema case, expected file, script golden fixture, or registry object was touched.

## Case-to-assertion matrix

Every binding reads the published case at run time and drives its assertions from the published fields. No published expectation is restated as a literal in Curator, and a published case, path role, forbidden effect, GC root, or scope with no binding fails the test rather than being skipped.

### Launch — `manager-lifecycle.json launcher_cases` (2/2)

| Published case | Executable assertion | Package |
| --- | --- | --- |
| `skill-command-without-shell-activation` | `TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus/skill-command-without-shell-activation` — installs a runtime root, writes the managed launcher, launches it twice as a real process | `internal/runtimestore` |
| `declared-system-command-without-profile` | same test, `/declared-system-command-without-profile` subtest, distinct exit status and primary dependency kind | `internal/runtimestore` |

Per case, driven by the published booleans and `required_path_roles`:

- `forward_arguments` — exact argv echo of `space value`, `quote"value`, `percent%value`, `$NOT_EXPANDED`, `Юникод`, and the empty string, asserted in both runs
- `preserve_exit_status` — real `exec.ExitError` code equals the implementation's status (23 / 61) in both runs
- `preserve_inherited_path` — with `PATH=<inherited>` the resulting PATH is exactly `<roles>:<inherited>` and an inherited-only helper still resolves; with an empty environment the roles lead and the inherited entry is provably unreachable
- `command_directory` — the launcher is published into the managed bin directory, which the launcher prefixes onto PATH so a sibling command shim resolves by bare name
- `implementation_runtime` — the process's `$0` is the exact installed runtime-store path, which is asserted to live under `runtimestore.Dir(home, skill, commit)`
- `system_dependencies` — the declared system directory is a carried path entry and its helper resolves by bare name with no profile sourced
- `platforms` — `unix` is executed; `windows` is asserted against the exact managed `WindowsShimContent` bytes on this host (see platform limits)

### Lifecycle — `manager-lifecycle.json bootstrap_cases` (3/3)

| Published case | Published outcome | Executable assertion |
| --- | --- | --- |
| `missing-config-if-missing` | `created` | `TestAuthoritativeBootstrapCasesAreExecutable/missing-config-if-missing` — `run(["bootstrap","--non-interactive","--skills-root",…,"--if-missing"])` exits `exitOK` and `config.Load` returns the requested skills root |
| `existing-config-if-missing` | `unchanged-success` | same test — exits `exitOK` and the unparseable configuration is byte-identical afterwards |
| `if-missing-with-force` | `usage-error` | same test — `config: "either"` is run against **both** a missing and an invalid configuration; each exits `exitUsage` and leaves its starting state untouched |

The argument vector is assembled from the published `if_missing`/`force` flags, so a flag change in the suite changes the invocation.

### Lifecycle — `manager-lifecycle.json upgrade_cases` (3/3)

| Published case | CLI selection | Executable assertion |
| --- | --- | --- |
| `selected-project-closure` | `upgrade app` | direct and transitive members installed into `.agents/skills` and reported as fetched; `unrelated` neither installed, nor reported fetched, nor left a `FETCH_HEAD` |
| `all-projects-deduplicate` | `upgrade --all` | both projects installed; each shared repository reports exactly one `fetched` line across the whole selection |
| `global-closure` | `global upgrade` | direct and transitive members installed into `home/global/skills` and reported fetched; `unrelated` excluded on all three surfaces |

Fixture: each skill is published as an upstream repository and cloned into the skills root with a working origin, so `upgrade` performs a real `git fetch` rather than resolving purely locally.

Non-vacuity probe: mutating the deduplication expectation from 1 to 2 makes `all-projects-deduplicate` fail (exit 1), confirming the observed count is genuinely 1. The file was restored byte-for-byte and re-run green.

### Lifecycle — `manager-lifecycle.json dry_run_cases` (2/2, 9 forbidden effects each)

`TestAuthoritativeDryRunCasesMutateNothingPersistent/{project-upgrade,global-upgrade}`. The machine is armed first so absence is a decision, not an accident: one skill whose origin is a working local upstream, one skill absent from the skills root that can only be resolved by cloning, the audit gate enabled, a signed loopback registry, and a real configuration document on disk.

| Published forbidden effect | Executable assertion |
| --- | --- |
| `source-fetch` | `<skillsRoot>/fetchable/.git/FETCH_HEAD` absent |
| `source-clone` | `<skillsRoot>/clonable` and `home/dev` absent, although the plan did resolve the cloned skill |
| `snapshot-cache` | `home/cache` absent |
| `response-cache` | `home/cache/registry` absent |
| `audit-state` | `home/audit` absent |
| `registry-state` | `home/state/registry` absent |
| `configuration` | configuration bytes identical to the pre-run baseline |
| `runtime` | `home/runtime` absent |
| `project-artifacts` | `project/.agents` and `project/.claude` absent |
| `global-artifacts` | `home/global/skills` and `home/global/bin` absent |

Guards: the run must report having resolved the clone-only skill, and the registry surfaces are only accepted as proof when the loopback registry actually served a request (asserted).

### Cache and failure — `build-drivers.json`

| Published case | Executable assertion |
| --- | --- |
| `protected-cache-hit` (`result: cache-hit`) | `TestAuthoritativeCacheOutcomesDriveInstallation/protected-cache-hit` — the planner outcome string equals the published `result`; no builder call (the published `source_aware_go_commands` is empty); nothing staged (`artifact_executed: false`); the protected boundary was inspected (`protected_boundary_verified: true`); the artifact path is the entry the boundary reported |
| `compiler-free-dry-run-miss` (`result: would-preflight-and-build`) | same test — outcome equals the published `result`; no builder call; the toolchain probe (the published package-independent commands) was taken while zero build sessions were established; `persistent_effects: []` proven by `requireAbsent` over every scope path |
| 16 `cache`-boundary rejections | `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted/<name>` — one real installation per case: published expectation re-checked as fail-closed, the command is rebuilt privately, the refused artifact path is never adopted, the outcome is never `cache-hit`, and the rebuilt receipt matches the planned key |

Cases: `artifact-hash-mismatch`, `artifact-link`, `artifact-path-mismatch`, `artifact-size-mismatch`, `artifact-special-file`, `cache-key-mismatch`, `cache-wrong-build-source`, `cache-wrong-policy`, `cache-wrong-target`, `cache-wrong-toolchain`, `concurrent-publisher-different-bytes`, `noncanonical-receipt-trailing-lf`, `noncanonical-receipt-whitespace`, `partial-cache-entry`, `receipt-hash-mismatch`, `self-consistent-forged-receipt-outside-protected-state`.

Each case is driven against one of the two non-reusable protected-boundary verdicts (`corrupt`, `untrusted-provenance`), assigned by sorted position so the assignment is deterministic; the test additionally asserts both verdicts were exercised across the cluster.

The planner's `BuildOutcome` vocabulary is identical to the published vocabulary, so published values are compared directly with no private translation table.

Curator-owned failure behaviour bound alongside and re-run green: `TestSecondBuildFailurePreservesPriorInstallationAndLiveCache`, `TestCorruptCacheEntryIsRebuiltAndNeverReused`, `TestUntrustedCacheEntryIsRebuiltAndNeverReused`, `TestCacheHitPerformsNoSourceAwareGoCommand`, `TestDryRunReportsCacheHitWithoutBuilding`, `TestRecoveryCompletesBeforeAnyNewMutation`, and the real-toolchain CLI repair/rollback set `TestInstallAndUpgradeRepairCorruptCompiledState`, `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall`, `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails`, `TestDryRunNeverClaimsACompletedCompilerCheck`.

### GC — `external-repository-lifecycle.json status_repair_gc_cases`

`TestAuthoritativeGarbageCollectionRootsAreRetained/gc-retains-roots/<root>`, one executable proof per published root, each measured against a pass that provably swept something else:

| Published root | Executable assertion |
| --- | --- |
| `artifact-receipts` | a marker-referenced protected entry and its `curator-receipt.ccj.json` survive a pass that removes a backdated orphan |
| `in-flight-journals` | an entry named only by `JournalKeys` survives while an unreferenced sibling is removed |
| `install-markers` | the runtime commit an install marker names survives while an unreferenced commit is pruned; the consumer stays registered |
| `protected-snapshots` | the protected snapshot tree is byte-identical after a pass that did remove runtime |
| `uncertain-entries` | one unreadable marker makes the reference set unprovable: nothing is swept, the skip is warned, and the uncertain consumer is not forgotten |

Curator-owned GC behaviour re-run green: `TestCollectSweepsOnlyUnreferencedProtectedEntries`, `TestCollectRetainsAJournalOwnedEntry`, `TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead`, `TestCollectPrunesConsumersInsideTheSamePass`, `TestCollectRequiresTheHomeLock`, `TestGcKeepsReferencedRuntime`, `TestGcPrunesDeadConsumers`, and the CLI serialization set `TestGCPrunesDeadConsumersUnderTheHomeLock`, `TestGCWaitsForTheHomeLock`, `TestGCRunsSerializedAcrossConcurrentInvocations`, `TestGCRetainsAndReportsReferencedCompiledState`.

### Concurrency

Candidate-owned, run under focused `-race`:

- `TestConcurrentProjectInstallsPreserveBothConsumers` — two checkouts install as separate processes; both consumers survive the ledger merge
- `TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets` — a rollback cannot restore over another project's committed shared state

## Exact commands and exit codes

Every gate ran as a standalone process, without `tee` and without a pipe, with `-count=1` and an exact `-run` filter, one package at a time. `CONF` below is the authoritative root path.

| Command | Exit | Result |
| --- | ---: | --- |
| `gofmt -l <the five new files>` | 0 | no output |
| `go vet ./internal/runtimestore ./internal/install ./internal/scopes ./cmd/curator` | 0 | no diagnostics |
| `git diff --check` | 0 | no whitespace errors |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/runtimestore -run '^(TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus\|TestCandidateManagerLauncherContract\|TestUnixLauncherCarriesRuntimePathArgumentsAndExitStatus\|TestWindowsLauncherCarriesRuntimePathAndExitStatus\|TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly\|TestUnixPostInstallShimPropagatesSignal)$'` | 0 | ok 7.719s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/scopes -run '^(TestAuthoritativeGarbageCollectionRootsAreRetained\|TestCollectSweepsOnlyUnreferencedProtectedEntries\|TestCollectRetainsAJournalOwnedEntry\|TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead\|TestCollectPrunesConsumersInsideTheSamePass\|TestCollectRequiresTheHomeLock\|TestGcKeepsReferencedRuntime\|TestGcPrunesDeadConsumers)$'` | 0 | ok 0.410s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/install -run '^(TestAuthoritativeDryRunCasesMutateNothingPersistent\|TestAuthoritativeCacheOutcomesDriveInstallation\|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted)$'` | 0 | ok 30.446s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/install -run '^(TestCacheHitPerformsNoSourceAwareGoCommand\|TestDryRunReportsCacheHitWithoutBuilding\|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache\|TestCorruptCacheEntryIsRebuiltAndNeverReused\|TestUntrustedCacheEntryIsRebuiltAndNeverReused\|TestRecoveryCompletesBeforeAnyNewMutation\|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook\|TestRuntimeLauncherCapturesDeclaredSystemDependency\|TestDryRunTouchesNothing\|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged)$'` | 0 | ok 22.634s |
| `go test -race -count=1 ./internal/install -run '^(TestConcurrentProjectInstallsPreserveBothConsumers\|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets)$'` | 0 | ok 35.307s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./cmd/curator -run '^(TestAuthoritativeBootstrapCasesAreExecutable\|TestAuthoritativeUpgradeCasesAreExecutable\|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing\|TestBootstrapIfMissingCreatesAbsentConfig\|TestBootstrapIfMissingRejectsForce\|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot\|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot)$'` | 0 | ok 12.847s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./cmd/curator -run '^(TestGCPrunesDeadConsumersUnderTheHomeLock\|TestGCWaitsForTheHomeLock\|TestGCRunsSerializedAcrossConcurrentInvocations\|TestGCRetainsAndReportsReferencedCompiledState)$'` | 0 | ok 16.233s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./cmd/curator -run '^(TestInstallAndUpgradeRepairCorruptCompiledState\|TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall\|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails\|TestDryRunNeverClaimsACompletedCompilerCheck)$'` | 0 | ok 190.706s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/interop -run '^(TestManagerLifecycleVectors\|TestGoldenMarkerObject\|TestGoldenFederationSemantics)$'` | 0 | ok 0.571s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/skillspec -run '^(TestManifestAndFilesystemRejectionVectors\|TestSchemaSixMixedScriptAndBuildCommandsVector)$'` | 0 | ok 0.380s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./cmd/curator -run '^TestAuthoritativeUpgradeCasesAreExecutable$'` (post-probe revert re-check) | 0 | ok 12.946s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/godriver -run '^(TestFixedEnvironmentAndFiveDirectArgvFormsVector\|TestToolchainIdentityVectors\|TestValidPackageGraphVectors\|TestPortableExecutionPolicyMatchesTheAcceptedVector\|TestCacheIdentityMatchesTheAcceptedVector\|TestCandidateGoV1SourceAwareContract)$'` | 0 | ok 1.497s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/godriver -run '^TestDriverRejectionClustersMapToStableCuratorErrors$'` | 0 | ok 9.698s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/buildcache -run '^(TestProtectedCacheHitVector\|TestCompilerFreeDryRunMissVector\|TestCacheRejectionClustersMapToStableCuratorOutcomes)$'` | 0 | ok 0.550s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/buildsource -run '^TestBuildSourceIdentityVectors$'` | 0 | ok 0.357s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/buildmeta -run '^TestPortableExecutionPolicyIsTheOnlyAdmittedBuildInput$'` | 0 | ok 1.037s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/whitelist -run '^TestBuildRootExcludedFromAgentContextVector$'` | 0 | ok 0.426s |
| `CURATOR_CONFORMANCE_ROOT=$CONF go test -count=1 ./internal/skillcheck -run '^TestBuildRootContentInContextVector$'` | 0 | ok 0.881s |

The last seven rows execute the pre-existing consumers of the remaining published positive vectors — the six `accepted` cases (schema-6 mixed commands, build-root context exclusion, standard-library and vendor package graphs, fixed environment and five argv forms, portable execution policy), the toolchain and build-source identity clusters, and the store-level cache-boundary and driver rejection clusters. Together with the two cache-outcome bindings above, every published positive vector now has an executed assertion in this pass, not merely a statically present one.

Expected-red gate: the deduplication non-vacuity probe deliberately ran with a mutated expectation and exited **1** (`FAIL github.com/relux-works/curator/cmd/curator 5.776s`). That is the expected failure of a mutation probe, not a passing gate; the file was restored and re-run green as the last row above.

## Disk and process barrier

| Point | Available KB | Candidate KB | Conformance root KB |
| --- | ---: | ---: | ---: |
| before | 23,082,588 | 3,676 | 2,068 |
| after | 22,983,288 | 3,676 | 2,068 |

Delta: 99,300 KB (~97 MB) of Go build/test cache growth. The candidate worktree and the authoritative root are byte-size unchanged; the authoritative vector digest re-verified identical. `pgrep -x go compile link curator` exited 1 after validation — no stray toolchain or manager process survived. No cache was cleared, no timeout altered, no software installed, nothing staged, committed, published, or repinned.

## Platform limits and out-of-scope routing

1. **Windows execution.** This is a Darwin host. Both published launcher cases declare `unix` and `windows`; `unix` is executed as a real process, and `windows` is asserted against the exact managed launcher bytes (`PATH` prefix per role, `call "<implementation>" %*`, `exit /b %ERRORLEVEL%`). Executing a Windows launcher needs a Windows runner and is not claimed here.
2. **Whole-suite gates not run.** The task acceptance criteria name `go test ./...` and `go test -race ./...`. The task instructions forbid both, along with whole-package `cmd/curator` and install/lifecycle aggregates. They were **not** run and are not claimed. An independent verifier with an amended allowlist should run them.
3. **Coverage: the ~80% target does not apply to this change, and whole-package coverage is not measurable inside the allowlist.** This task added no product statement and changed none, so there is no newly affected product code for a statement-coverage target to describe. Two scoped, package-focused measurements were taken anyway, so a real number is on record rather than a claim:

   | Command | Exit | Coverage |
   | --- | ---: | --- |
   | `CURATOR_CONFORMANCE_ROOT=$CONF go test -cover -count=1 ./internal/scopes -run '^TestAuthoritativeGarbageCollectionRootsAreRetained$'` | 0 | 42.0% of statements |
   | `CURATOR_CONFORMANCE_ROOT=$CONF go test -cover -count=1 ./internal/runtimestore -run '^TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus$'` | 0 | 16.1% of statements |

   These are whole-package statement coverage reached by the new conformance tests **alone**, which is exactly what a narrow binding test should produce; they are not a package coverage verdict. A real package coverage figure needs the full per-package suites, which the allowlist forbids in this pass. Board checklist item 11 is checked on that stated basis: no product code was affected, and every published in-scope case has an executed assertion.
4. **Repository-wide lint not run.** Lint evidence is scoped `go vet` over the four packages touched, plus exact-file `gofmt`.
5. **`external-repository-lifecycle.json` is out of scope except `gc-retains-roots`.** The candidate accepts skill schemas 1 through 6 (`SupportedSchemaVersions`) and publishes no `build_repository_*` code anywhere in the tree. The document's status/repair codes (`build_repository_non_current`, `build_repository_currentness_unknown`), offline cache cases (`build_repository_unverified_offline`, `build_repository_source_unavailable`), schema-7 mixed-build, signing, source-covering, and external path-shim cases therefore describe an unimplemented feature. Binding them would have required inventing Curator codes to satisfy a vector — a forced fit. They are routed to the owning external-repository implementation task rather than stubbed. `gc-retains-roots` is bound because every one of its roots (artifact receipts, in-flight journals, install markers, protected snapshots, uncertain entries) is a real, implemented Curator retention root.
6. **No product change was needed.** No executable regression surfaced in product code during this rework, so no product file was edited and no defect was routed on that basis.

## Board checklist semantics

The board requires every checklist item to be checked before a handoff, and the CLI exposes no way to remove or reword an item. Items 13 and 14 are therefore **records, not resolutions**:

- item 13 records the remaining gap for an independent verifier — `go test ./...`, `go test -race ./...`, coverage measurement, and Windows launcher execution were **not run and are not claimed**. Checked means the gap is documented, not closed.
- item 14 records the routing decision for the schema-7 external-repository cases. Checked means the decision is recorded and the cases are routed to their owning task, not that they are implemented.

Item 4 is checked because the code and tests described by the task are complete and every acceptance clause this pass was permitted to execute exited 0. The two acceptance gate commands the task instructions forbid are the gap in item 13.
