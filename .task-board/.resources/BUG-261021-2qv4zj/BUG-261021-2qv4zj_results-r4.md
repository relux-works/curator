# BUG-261021-2qv4zj — round 4 results

Branch: `fix/audit-markdown-scope`.
Head: `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2`.
Round base: `d20c4363cf2dea88ce7418fa594664a466c116c8`.
Original base: `2c89fd46443d461e25bd377fb94a0c6754282838`.

The round-4 brief chooses publication of the core section 3.1 skill projection.
Native and managed homes now share a content-addressed projected store. Link
and fallback-copy trees contain the same selected files and bytes. Planning
and status are read-only; materialization publishes complete staged projections
under the existing mutation lock. Publication revision `core-3.1-v1` versions
both projection paths and skills surface identities. Old raw-root links and
markers are stale even when their file bytes happen to match. Repair moves
Curator-recorded links and tree copies to the projection without takeover.
Missing projected stores are stale and rebuilt by repair. Corrupt/relinked
projected stores refuse reuse.

Changelog: Publish profile skill surfaces through the core section 3.1 context
projection and repair owned legacy links/copies without takeover.

Audit ruleset 8 is unchanged: audited inputs, verdicts and cache identity are
unchanged in this round. Publication identity is confined to projection store
paths and surface hashes. Earlier F1/F2 changes remain intact.

## Publisher inventory

The initial table was posted in the task topic before production edits. A
follow-up added the coordinated identity-migration publisher found during the
implementation.

| Publisher | Previous source | Current source |
| --- | --- | --- |
| Native profile skills | Selected raw package root | Shared core 3.1 projection |
| Managed profile skills | Raw root via `skillsOf` | Same projected target |
| Native tree-copy fallback | Link source tree | Projected source tree |
| Managed tree-copy fallback | Link source tree | Same projected source tree |
| Native status/repair | Raw target and raw content hash | Projected target, versioned surface identity, projected-store verification |
| Managed status/repair/markers | Raw `skillsOf` plan | Projected plan and surface identity |
| Coordinated legacy identity migration | Rebuilt raw-root links/copies | Publish projection before activating links or complete copied trees |
| Profile-global operations | Profile publishers | Same shared profile publishers |
| Legacy global/project context and adapter discovery | `install.stageNode` uses `CopyContext`; adapters expose installed canonical context | Unchanged shared core projection |
| Script/compiled commands | `install/targets.go`, `runtimestore`, build caches | Unchanged; no command publisher consumes profile skill links |

Runtime evidence: `install/targets.go` prepares script runtimes directly from
snapshot/spec inputs using `runtimestore.PrepareScriptRuntime`, and constructs
compiled targets from build-cache entries. `runtimestore` independently names
runtime entries under `runtime/<skill>/<commit>`. Neither depends on the
`skillOf.root` link target. Raw immutable package snapshots remain available
for pin verification and runtime/build input selection; context projection
never modifies them. Build roots remain absent from installed context and
command runtime trees; generated outputs remain in build caches.

## Files and commits

- `internal/envprofile/skill_projection.go`: pure shared-walker projection plan,
  content-addressed publication, store verification, recorded-copy selection,
  recoverable complete-tree fallback replacement.
- `internal/envprofile/managed.go`: projected skills/identity, publication during
  writes, read-only verification, migration of recorded fallback copies.
- `internal/envprofile/switch.go`: native projected publication and identity.
- `internal/envprofile/native_skills.go`: projected native currentness/repair.
- `internal/envprofile/identity_migration.go`: populate immutable projections
  before coordinated activation, including existing fallback copies.
- `internal/envprofile/skill_directory_test.go`: selected-directory lifecycle
  now expects its canonical projected target; the vendored corpus is unchanged.
- `internal/envprofile/skill_projection_test.go`: publication, selection,
  currentness, migration, missing/corrupt stores and rollback/retry evidence.
- `internal/envprofile/skill_projection_mutants_test.go`: thirteen narrowing
  controls through source overlays and named behavioral assertions.
- `internal/envprofile/global_migration_capabilities_test.go`: the admitted skill control includes the core-required `SKILL.md`; its capability assertions and context controls are unchanged.
- `cmd/curator/muse_credential_mutants_test.go`: update the existing positive
  copy-fallback source-overlay anchor after link-repair refactoring; retain its
  production copied-root assertions.
- `internal/envprofile/nul_opaque_test.go` and `nul_opaque_v2_test.go`: add
  identical required skill entries to both framing fixtures and the clean
  control. The actual v1 collision and v1/v2 installation assertions remain.
- `internal/envprofile/network_contacts_test.go`: retain the required skill
  entry in all 40 unchanged corpus cases; keep admission, refusal evidence
  and publication-state assertions.
- `docs/profile-skill-directories.md`: publication/identity/repair/runtime notes.

Signed commits in this round:

- `f381e5deb08a5a1d36162b898ac6770edb8b5ddb`: publish core context projection.
- `997d5db4ebb65355eb358757af628f492dc1129c`: migrate old recorded skill copies
  through complete-tree replacement.
- `931a98fd894a17594f22eb495617a3f426573aab`: preserve legacy framing
  refusals in projection planning/staging/verification; correct the admitted
  skill fixture to satisfy core section 3.

- `736589d0c9f968f5a14393650bdb47d7a6c96da0`: retain the production
  copied-root positive control after the link-repair refactoring.

- `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2`: keep framing and contact
  admission fixtures valid for projected skill publication.

## Test claims

| Test | Claim and limits |
| --- | --- |
| Supplied native projection probe | Real built CLI admits repository-only inputs without exposing `.research/notes.md`, `references/README.md`, or `assets/tests/a.md` through native Claude/Codex skill trees; declared `references/guide.md` remains readable. |
| Supplied managed projection probe | Same cases after production managed `Resolve(Repair:true)` provisioning for both adapters. |
| `TestSkillProjectionCoreSelection` | All 19 fixed exclusions are absent at nested depths; visible neighboring files survive; planning writes nothing; v1/v2 digests match canonical `CopyContext`; shared publication and both copy helpers preserve the expected tree. |
| `TestProfileSkillProjectionPublication` | A repository-root skill with board, research and vendored build inputs installs. Native/managed Claude/Codex surfaces expose exactly the selected bytes; command-less scripts survive, exported scripts/runtime/build roots do not; both fallback helpers copy identical trees, and real recorded fallback copies verify as current. |
| `TestProfileSkillProjectionRepair` | Old marker-only, raw-root link and raw-root copied homes are stale; repair needs no takeover, preserves the lock, removes excluded old files and converges for both adapters. |
| `TestSkillProjectionStoreReuse` | Intact store reuse succeeds; changed bytes, extra links, linked destinations and a source changed after planning refuse with the intended diagnostic; failures do not change the external control target or publish an incomplete projection. |
| `TestProfileSkillProjectionMissingStore` | Missing projected stores make native/managed state stale without recreation by read-only verification; repair recreates the projected store. |
| `TestSkillProjectionFallbackRollback` | A fault after actual copied-tree target commit restores the complete old tree; retry publishes exactly the projected tree and removes old excluded files. The test requires the fault to fire. |
| `TestSkillProjectionLegacyFraming` | Instrumented planning, staging and store verification reject NUL before any v1 computation (`CountV1Hashes == 0`); v2 publishes and verifies identical bytes. |
| `TestSkillProjectionNarrowingMutants` | Restoring either raw publisher, exported scripts, omitted runtime/build exclusions, old surface identity, missing native/managed store checks, old tree-copy migration, changed-byte reuse, linked-tree reuse, changed-source publication, or unguarded legacy planning/tree hashing must fail its named behavioral assertion. Only exit 1 plus that assertion counts. |
| Existing Markdown/kind/cache/alias and skill-directory suites | Preserve F1/F2, core-walker alias/exclusion behavior, context-module publication reader, declared/undeclared controls, cache generations/kind binding, selected-package isolation and existing mutation controls. |
| `TestMuseRootContextCopyFallback` | An overlay forces only the root-context link failure; production fallback must record and verify the copied root on repair and current resolve. |
| Framing collision installation pair | Both fixtures retain identical valid skill entries; their raw v1 digests still collide. The v1 refusals name the intended data files, the clean control installs, and both data fixtures install under v2. |
| `TestProfileNetworkContactVectors` | All 40 unchanged corpus cases exercise migration and installation; permitted cases retain the pinned member and refused cases name the intended file/contact and publish no source/lock. The fixture now includes the required context entry. |
| Full default and pinned conformance | Broad platform/CLI/runtime/build/transaction/profile behavior; final exact inventory and totals follow below. |

The fallback helpers and recorded copied homes execute directly on Darwin arm64;
this does not claim an actual Windows run or an operating-system symlink failure.
No installed operator state is used.

## Validation chronology and limits

Both supplied reviewer probes were run before edits: exit 1, their three
excluded-path cases failed on both adapters while the declared-visible control
passed. After the initial implementation both probes passed.

Additional migration inspection prompted a raw copied-home regression: before
the copy repair it failed specifically when the link publisher attempted to
rename over a non-empty copied directory. That regression now passes. Complete
copied-tree replacement uses the existing recoverable entry transaction, with
an exercised commit-failure rollback and retry.

A preliminary full run at `f381e5de` completed eleven groups without failures;
remaining groups were intentionally not started after discovery of the copied
home migration case. That partial run is not counted as final full-suite evidence.
A second preliminary run at `997d5db4` was superseded to retain legacy hash-frame
refusals in the new projection helpers. It also found one positive capability
fixture missing the `SKILL.md` required by core section 3; that fixture now
includes the required context entry. Its original capability assertions and
negative context cases remain unchanged and pass. Neither partial run counts as
final full-suite evidence. A third partial run at `931a98fd` found that the existing
`TestMuseRootContextCopyFallback` source-overlay anchor no longer matched the
refactored link loop. Its actual production fallback path and copied-root
assertions are unchanged; the replacement anchor is unique and the focused
control passes. The partial run is superseded. The entire default suite is
rerun from a normal disposable `.git` checkout at `a2244ec1`.
A diagnostic pass at `736589d0` completed all 39 groups: two profile groups
found admitted framing and contact fixtures without the core-required skill
entry; every other group passed. Those fixtures now include a valid entry.
The framing maps receive identical bytes, and their explicit v1 collision
assertion still passes. The contact corpus and its expected outcomes are
unchanged. Focused corrected checks pass; the entire suite is repeated at
the final head. No earlier partial or diagnostic pass counts as the final suite. A local validation-fetch command initially used a relative source path that
was invalid from the clone (exit 128), followed by an unavailable-head checkout
(exit 128). The explicit local source path and detached-head equality check
succeeded before final build, vet, inventory and test commands were started.
Intermediate fixture/expectation and test-runner selection corrections
receive no final passing credit; the final inventory check requires every named
top-level test exactly once.

## Final validation

All final checks ran at `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2` in a clean,
disposable normal `.git` checkout. Build and vet passed. Both supplied probes
passed separately against this same head. All nine introduced commit signatures
verify as `G`; the feature branch was pushed normally and its remote head was
read back as the exact validated commit. No pull request or default-branch write
was made.

| Check | Result | Exit |
| --- | --- | ---: |
| `go build ./...` | Pass | 0 |
| `go vet ./...` | Pass | 0 |
| Supplied native probe | All exclusion cases and visible control pass | 0 |
| Supplied managed probe | All exclusion cases and visible control pass | 0 |
| Full default suite | 39 groups, 85 packages, all 3594 listed top-level test/example names run exactly once; 11759 test/subtest passes, 199 skips, zero failures | 0 for every group |
| Pinned conformance | 1424 test/subtest passes, zero skips/failures; all 45 selected environment top-level tests run exactly once | 0 for every group |
| Diff whitespace and signed-object verification | Clean; nine signatures `G` | 0 |

Pinned conformance spec: `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, root
`$SPEC/conformance/v1`, matching the repository workflow pin. Core packages cover
audit, skill checking, whitelist walking, context packages/locks/materialization;
three environment groups exercise the selected production integration vectors.
The shared protocol and environment interop packages also pass all 277 checks.

Platform: Darwin arm64; Go 1.25.5, Rust 1.91.0, Node v26.12.0, Git 2.54.0.
All Go tests unset `GIT_DIR` and `GIT_WORK_TREE`. Validation used private HOME,
GOCACHE, GOPATH/module cache and application configuration; `TMPDIR` and
`GOTMPDIR` were `/tmp/cw-dev3`. Tests did not read installed operator state.
The mutation suite requires the intended named assertion and exit 1 for every
control; unrelated setup failures do not count. The final default inventory
includes all thirteen new narrowing controls as well as prior controls.

The full default suite intentionally leaves conformance-gated cases skipped when
`CURATOR_CONFORMANCE_ROOT` is absent. The separate pinned run above exercises
the selected required vectors without skips. Remaining default skips retain
their existing platform/tool/contract conditions; they are not counted as passes.

Observed default skipped test/subtest names:

- `cmd/curator/TestEnvUnmanageBackupRecordVectors`
- `cmd/curator/TestSignerRowsSurviveHostileGlobalGitConfig`
- `cmd/curator/TestAuthoritativeLifecycleUpgradeCasesAreExecutable`
- `cmd/curator/TestAuthoritativeLifecycleBootstrapCasesAreExecutable`
- `cmd/curator/TestMuseFragmentV3PublishedCases`
- `cmd/curator/TestProfileUpdateAllDeltaVectorsAtCLI`
- `cmd/curator/TestProfileUpdateAllStopsAtFirstUnconfirmedDelta`
- `cmd/curator/TestProfileInstallReinstallDeltaVectorsAtCLI`
- `cmd/curator/TestSourceSignerVectorsAtProfileInstall/unsigned-refused`
- `cmd/curator/TestSourceSignerVectorsAtProfileInstall/wrong-signer-refused`
- `cmd/curator/TestSourceSignerVectorsAtProfileInstall/ssh-tag-signature-accepted`
- `cmd/curator/TestSourceSignerVectorsAtProfileInstall/empty-allowlist-signed-refused`
- `cmd/curator/TestSecurityPostureVectorsThroughCLI`
- `cmd/curator/TestUmbrellaProviderResolutionVectors`
- `internal/envprofile/TestEnvironmentsCodexSeedVectors`
- `internal/envprofile/TestEnvironmentsEnvPassthroughVectors`
- `internal/envprofile/TestDotfileManagerVectors`
- `internal/envprofile/TestMusePublishedLinkStates`
- `internal/envprofile/TestMusePublishedRegistryLayout`
- `internal/envprofile/TestPathKindMCPVectorsDriveInstallEntry`
- `internal/envprofile/TestPathKindBoundaryVectorsDriveProductionEntries`
- `internal/envprofile/TestPathKindDryRunVectorsDriveReadOnlyResolve`
- `internal/envprofile/TestFoldedPathsAreSourceInvalid`
- `internal/envprofile/TestSystemDeltaVectors`
- `internal/envprofile/TestEnvironmentsReadFailureVectors`
- `internal/envprofile/TestUpdateConfirmationPostureVectorAtStatus`
- `internal/envprofile/TestSignerPostureVectorsAtStatus`
- `internal/envprofile/TestStoreBoundaryResolveVectorsDriveResolve`
- `internal/envprofile/TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan`
- `internal/envprofile/TestStoreBoundaryRepairVectorsDriveResolveRepair`
- `internal/envprofile/TestStoreBoundaryStatusVectorsDriveStatus`
- `internal/envprofile/TestEnvironmentWriteNofollowVectors`
- `internal/install/TestInstallHelperProcess`
- `internal/install/TestAuthoritativeCacheOutcomesDriveInstallation`
- `internal/install/TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted`
- `internal/install/TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess`
- `internal/install/TestDryRunEffectBindingsSeeWhatARealOperationWrites`
- `internal/install/TestAuthoritativeMixedBuildCasesUseProjectInstallEntry`
- `internal/install/TestAuthoritativePathShimCasesUseProjectInstallEntry`
- `internal/install/TestAuthoritativeSigningCasesUseProjectInstallEntry`
- `internal/install/TestAuthoritativeTransactionCasesUseProjectInstallEntry`
- `internal/install/TestRegistryBootstrapVectorsDriveInstall`
- `internal/install/TestRegistryMirrorComparisonVectorsDriveInstall`
- `internal/buildmeta/TestPortableExecutionPolicyIsTheOnlyAdmittedBuildInput`
- `internal/buildmeta/TestCandidateBuildMetadataArtifacts`
- `internal/buildmeta/TestCandidateBuildReceiptSchemaCase`
- `internal/buildcache/TestProtectedCacheHitVector`
- `internal/buildcache/TestCompilerFreeDryRunMissVector`
- `internal/buildcache/TestCacheRejectionClustersMapToStableCuratorOutcomes`
- `internal/conformancecoverage/TestContentHashV2VectorsWhenPublished`
- `internal/conformancecoverage/TestMuseCandidateHistoricalHashV2FamiliesAreKnownGapsWhenPublished`
- `internal/conformancecoverage/TestMuseCandidateHistoricalHashV2VectorsAreKnownGapsWhenPublished`
- `internal/buildrepo/TestExternalRepositoryAcquisitionConformance`
- `internal/buildsource/TestBuildSourceConformanceVectors`
- `internal/config/TestManagerConfigV2SchemaCases`
- `internal/config/TestManagerConfigV3SchemaCases`
- `internal/config/TestSourceSignerMergeVectorsAtLoad`
- `internal/config/TestSystemConfigV2SchemaCases`
- `internal/config/TestSystemConfigV2IsolationDirectionsFromPinnedCases`
- `internal/config/TestManagerConfigV2Vectors`
- `internal/config/TestOverlayConformancePassesWithPermissionsAndSourceSigners`
- `internal/config/TestSecurityPostureVectors`
- `internal/config/TestSystemModuleSchemaSubset`
- `internal/buildcache/TestCandidateCacheOutcomeVocabulary`
- `internal/closure/TestClosureOrderingConformanceVector`
- `internal/closure/TestBuildCommandOrderingConformanceVector`
- `internal/buildrepo/TestReleasedSkillBuildSchemaCases`
- `internal/buildrepo/TestReleasedSourceIdentityVectors`
- `internal/buildrepo/TestExternalReceiptV2CacheKeyVector`
- `internal/buildrepo/TestVerifyPublishedReleasePinWhenConformanceRootIsSupplied`
- `internal/buildrepo/TestResolvedTransportWindowsRefusesBeforeAnyProcess`
- `internal/contextmaterialize/TestSystemModuleAdmissionVectors`
- `internal/contextlock/TestContextLockV2SchemaCases`
- `internal/contextresolve/TestSourceSignerVectorsAtResolve`
- `internal/contextlock/TestParseRejectsHashVersionOnFrozenContextLock`
- `internal/envfragment/TestFragmentAuthoritativeSchemaCases`
- `internal/envmarker/TestParseAuthoritativeEnvMarkerSchemaCases`
- `internal/envmarker/TestParseAuthoritativeEnvMarkerV2SchemaCases`
- `internal/envmarker/TestParseAuthoritativeEnvMarkerV3SchemaCases`
- `internal/devsub/TestSkillfileDevV2PublishedCasesUseCoverageRatchet`
- `internal/hashing/TestCrossCheckTree`
- `internal/hookapproval/TestWindowsDriveIdentity`
- `internal/marker/TestReadRejectsHashVersionOnFrozenMarkerV3`
- `internal/interop/environments/TestConformanceContextDetectors`
- `internal/interop/environments/TestConformanceEnvironmentsHeader`
- `internal/interop/environments/TestConformanceEnvironmentsMonolithic`
- `internal/interop/environments/TestConformanceContextResolution`
- `internal/interop/environments/TestConformanceContextVersions`
- `internal/interop/environments/TestConformanceEveryPathTheVectorsNameIsDeclared`
- `internal/interop/TestGoldenMarkerObject`
- `internal/interop/TestGoldenSnapshotHash`
- `internal/interop/TestGoldenContextCopy`
- `internal/interop/TestGoldenRegistryObjects`
- `internal/interop/TestCanonicalJSONVectors`
- `internal/interop/TestSourceIdentityVectors`
- `internal/interop/TestIdentifierVectors`
- `internal/interop/TestLocaleSelectorVectors`
- `internal/interop/TestManagerConfigVectors`
- `internal/interop/TestGoldenFederationSemantics`
- `internal/interop/environments/TestConformanceSnapshotAcquisition`
- `internal/marker/TestReadAuthoritativeValidMarkers`
- `internal/marker/TestReadAuthoritativeMarkerExternalCrossFields/install-marker-v3`
- `internal/marker/TestReadAuthoritativeMarkerExternalCrossFields/install-marker-v4`
- `internal/marker/TestAuthoritativeCompiledMarkerRoundTripsThroughV2Writer`
- `internal/marker/TestAuthoritativeCompiledMarkerRoundTripsThroughWriter`
- `internal/godriver/TestBuildFailsClosedWhenTheGoChildCannotStart`
- `internal/godriver/TestPortableExecutionPolicyMatchesTheAcceptedVector`
- `internal/godriver/TestMandatoryControlsMatchTheAcceptedVector`
- `internal/godriver/TestNativeControlInventoryMatchesTheAcceptedVector`
- `internal/godriver/TestCapabilityEvidenceRecordMatchesTheAcceptedVector`
- `internal/godriver/TestCapabilityEvidenceCasesMatchTheAcceptedVector`
- `internal/godriver/TestDeferredGuaranteesAndFailureBoundaryMatchTheAcceptedVector`
- `internal/godriver/TestIdentityProtocolAndPackageInfluenceCodesMatchTheAcceptedVector`
- `internal/godriver/TestCacheIdentityMatchesTheAcceptedVector`
- `internal/godriver/TestCandidateGoV1SourceAwareContract`
- `internal/marker/TestReadAuthoritativeMarkerV2SchemaCases`
- `internal/marker/TestReadAuthoritativeMarkerV3SchemaCases`
- `internal/marker/TestReadAuthoritativeMarkerV4SchemaCases`
- `internal/marker/TestReadAuthoritativeMarkerV5SchemaCases`
- `internal/gitops/TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive`
- `internal/godriver/TestRealGoV1VendoredBuildIsBoundedAndNotLaunched`
- `internal/godriver/TestFixedEnvironmentAndFiveDirectArgvFormsVector`
- `internal/godriver/TestToolchainIdentityVectors`
- `internal/godriver/TestValidPackageGraphVectors`
- `internal/godriver/TestDriverRejectionClustersMapToStableCuratorErrors`
- `internal/godriver/TestCandidateRC4ToolchainArtifacts`
- `internal/godriver/TestFingerprintRejectsInvalidUnicodePath`
- `internal/godriver/TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched`
- `internal/godriver/TestModuleRootVectorsDriveTheWholeBuild`
- `internal/godriver/TestRealTrustedGoProbe`
- `internal/moduleroots/TestModuleRootVectors`
- `internal/scopes/TestAuthoritativeGarbageCollectionRootsAreRetained`
- `internal/registry/TestPageBoundaryConformanceVectors`
- `internal/runtimestore/TestCandidateManagerLauncherContract`
- `internal/scriptpolicy/TestScriptHostExecutionPolicySectionsAreAllClassified`
- `internal/scriptpolicy/TestScriptHostExecutionPolicyProductionConsumersCoverAllCases`
- `internal/scriptpolicy/TestScriptExecutionPolicyIdentityMatchesTheSuite`
- `internal/scriptpolicy/TestScriptExecutionOptInCases`
- `internal/scriptpolicy/TestEnforcedShapesProceedPastAdmission`
- `internal/scriptpolicy/TestMandatoryControlsMatchTheSuite`
- `internal/scriptpolicy/TestPreflightRefusalCases`
- `internal/scriptpolicy/TestScriptAuditLabelCases`
- `internal/runtimestore/TestAuthoritativeLauncherCasesForwardArgvPathRolesAndExitStatus`
- `internal/pnpmsource/TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall`
- `internal/pnpmsource/TestRealPinnedPNPMLockSupersetSnapshotDependencies`
- `internal/npmsource/TestVerifiedProviderObservesRealNodeLaunchedNPMBoundary`
- `internal/scriptworker/TestLandlockProbeCancellationKillsProcessGroup`
- `internal/scriptworker/TestLandlockConstructionConfirmsTypedRules`
- `internal/scriptworker/TestLinuxLandlockProbeFindsPresentControls`
- `internal/scriptworker/TestLinuxLandlockProbeChildAnswersAtTheBoundary`
- `internal/npmsource/TestN01RealNPMCIUsesOnlyDerivedPrivateCache`
- `internal/scriptworker/TestLinuxHostConditionalUnavailableSucceeds`
- `internal/scriptworker/TestLinuxHostProbeAndEvidenceAreConsistent`
- `internal/scriptworker/TestWindowsJobLimitsAppliedAndConfirmed`
- `internal/scriptworker/TestLinuxLandlockConfinementMatchesProbe`
- `internal/scriptworker/TestLinuxWriteConfinementTruncateMatchesProbe`
- `internal/scriptworker/TestLinuxWriteConfinementDirectoryMutationMatchesProbe`
- `internal/scriptworker/TestLinuxMissingDerivedPathRefusesWriteConfinement`
- `internal/registry/TestAuditRecordV2SchemaCases`
- `internal/registry/TestRegistryV2CarrierSchemaCases`
- `internal/registry/TestParseRecordRejectsHashVersionOnFrozenAuditRecord`
- `internal/registry/TestParseRecordRejectsV2RecordInsideFrozenRegistryLogEntry`
- `internal/registry/TestParseLogEntryRejectsV2RecordInsideFrozenRegistryLogEntry`
- `internal/pnpmsource/TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization`
- `internal/pnpmsource/TestPNPMTemporaryTreeRemovable`
- `internal/pnpmsource/TestPNPMTemporaryTreeChild`
- `internal/scriptworker/TestWindowsRealInterpretersRunDeclaredExec`
- `internal/shell/TestShellHookTrustVectors`
- `internal/skillcheck/TestBuildRootContentInContextVector`
- `internal/shell/TestShellHookRefusesHostileCheckout/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/two-fields/powershell`
- `internal/skillspec/TestManifestAndFilesystemRejectionVectors`
- `internal/skillspec/TestSchemaSixMixedScriptAndBuildCommandsVector`
- `internal/skillspec/TestPortablePathConformanceVectors`
- `internal/skillspec/TestReleasedSchemaCases`
- `internal/shell/TestShellHookRejectsMalformedRecords/five-fields/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/forged-approver/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/empty-approver/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/uppercase-approver/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/uppercase-digest/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/short-digest/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/nonhex-digest/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/bad-timestamp/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/empty-timestamp/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/month-13/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/february-30/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/second-60/powershell`
- `internal/shell/TestShellHookRejectsMalformedRecords/lowercase-timestamp/powershell`
- `internal/shell/TestShellHookTrustResolvesSymlinkedProject/powershell`
- `internal/shell/TestShellHookTrustNativeRecordAuthorizesMSYSSpelling`
- `internal/shell/TestPowerShellHookRunsOnEveryPrompt`
- `internal/whitelist/TestBuildRootExcludedFromAgentContextVector`
- `internal/whitelist/TestBuildRootContextExclusionConformanceFixture`
- `internal/yarnmodernsource/TestModernRuntimeCyclesThroughRealPinnedYarn`
- `internal/yarnclassicsource/TestRealPinnedYarnClassicReadsPrivateMirrorWithEmptyOrdinaryCache`
- `internal/yarnmodernsource/TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor`
- `internal/yarnmodernsource/TestN10RealPinnedYarnWorkspacePeerPnPInvokeThroughVerifiedExecutor`
- `internal/yarnmodernsource/TestN10RealPinnedYarnRemotePeerContextsThroughVerifiedExecutor`
- `internal/install/TestAuthoritativeDryRunCasesMutateNothingPersistent`


## Open decisions

None. Publication projection and unchanged audit ruleset follow the round-4
orchestrator decision. Corrupt projected stores are diagnosed and refused;
missing stores are rebuilt from pinned package inputs.

## Command ledger

`$WORKTREE` is the owning checkout, `$VALIDATION` the disposable exact-head
checkout, `$SCRATCH` the task scratch directory, `$SPEC` the pinned spec
checkout, `$CURATOR_GO` the pinned Go executable, and `$GOFMT` its formatter.
The private environment described above applies to every entry. Final supplied
probe commands additionally set `GOFLAGS=-overlay=$SCRATCH/validation-<mode>.json`;
development probes use `$SCRATCH/<mode>.json` against the owning checkout.
pinned conformance commands set `CURATOR_CONFORMANCE_ROOT=$SPEC/conformance/v1`.
All final suite commands are split by anchored test-name expressions to keep
each invocation below the shell time limit. The deferred install guard ran after
the CLI-building queues completed. The inventory reconciliation itself exited 0.

```sh
python3 "$SCRATCH/check_full.py"         # exit 0
python3 "$SCRATCH/check_conformance.py"  # exit 0
```

The entries outside the final build/vet, `full-*`, `conformance-*`, and
`probe-*-exact-head` checks are development evidence from the evolving worktree.
The `*-before` probe/regression refusals are expected red witnesses. Other
nonzero development entries are superseded fixture/expectation/control-selector
iterations described above, not final passing evidence. Preliminary full runs
are listed separately and receive no final-suite credit.

### Final and development commands

`cli-fallback-control-final` — exit `0`; `$WORKTREE`; 162.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./cmd/curator -run '^TestMuseRootContextCopyFallback$'
```

`cli-fallback-control` — exit `1`; `$WORKTREE`; 192.74s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./cmd/curator -run '^TestMuseRootContextCopyFallback$'
```

`conformance-core` — exit `0`; `$VALIDATION`; 34.21s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json ./internal/audit ./internal/skillcheck ./internal/whitelist ./internal/contextpkg ./internal/contextlock ./internal/contextmaterialize
```

`conformance-env-1` — exit `0`; `$VALIDATION`; 114.07s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestClaudeSeedMergePreservesToolState|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCodexSeedStripsInlineMCPTable|TestDotfileManagerVectors|TestEnvironmentWriteNofollowVectors|TestEnvironmentsCodexSeedVectors|TestEnvironmentsEnvPassthroughVectors|TestEnvironmentsReadFailureVectors|TestMuseAuthParentBoundaryAtResolve|TestMuseAuthRepairAdditionalStates|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMusePublishedLinkStates|TestMusePublishedRegistryLayout|TestMuseStatusOptionalProvisioning|TestOpencodeFormFallback|TestOpencodeXDGSeeds)$' ./internal/envprofile
```

`conformance-env-2` — exit `0`; `$VALIDATION`; 48.67s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathKindMCPVectorsDriveInstallEntry|TestProfilePathCollision|TestRepairFailedMissingStore|TestReservedSkillName|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestSeedUnreadableStopsBeforeFirstWrite|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusProfileMembersAndUnregistered|TestStatusShadowAcknowledgment|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair)$' ./internal/envprofile
```

`conformance-env-3` — exit `0`; `$VALIDATION`; 27.04s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryStatusVectorsDriveStatus|TestSystemDeltaVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateConfirmationPostureVectorAtStatus)$' ./internal/envprofile
```

`conformance-interop` — exit `0`; `$VALIDATION`; 0.82s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json ./internal/interop ./internal/interop/environments
```

`copy-migration-regression-after` — exit `0`; `$WORKTREE`; 11.84s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestProfileSkillProjectionRepair$/./^raw-copy$'
```

`copy-migration-regression-before` — exit `1`; `$WORKTREE`; 5.49s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestProfileSkillProjectionRepair$/./^raw-copy$'
```

`copy-repair-focused-complete` — exit `0`; `$WORKTREE`; 44.42s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestSkillProjection.*|TestProfileSkillProjection.*|TestRC14IdentityMigrationPreservesFallbackCopies)$'
```

`copy-repair-focused-final` — exit `0`; `$WORKTREE`; 3.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestSkillProjection|TestProfileSkillProjection|TestRC14IdentityMigrationPreservesFallbackCopies)$'
```

`focused-audit` — exit `0`; `$WORKTREE`; 46.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/audit ./internal/whitelist ./internal/contextpkg ./internal/contextmaterialize
```

`focused-existing-final` — exit `0`; `$WORKTREE`; 127.22s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestProfileInstallMarkdown|TestSkillDirectory|TestNativeSkills|TestRC14|TestManagedHome)'
```

`focused-existing` — exit `1`; `$WORKTREE`; 121.71s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestProfileInstallMarkdown|TestSkillDirectory|TestNativeSkills|TestRC14|TestManagedHome)'
```

`focused-mutants-2` — exit `1`; `$WORKTREE`; 35.27s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^(TestSkillProjectionNarrowingMutants|TestSkillProjectionStoreReuse|TestProfileSkillProjectionRepair|TestProfileSkillProjectionPublication)$'
```

`focused-mutants` — exit `1`; `$WORKTREE`; 29.79s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^(TestSkillProjectionNarrowingMutants|TestProfileSkillProjectionRepair|TestProfileSkillProjectionPublication)$'
```

`focused-projection-2` — exit `0`; `$WORKTREE`; 32.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^(TestSkillProjectionCoreSelection|TestProfileSkillProjectionPublication|TestProfileSkillProjectionRepair|TestRC14IdentityMigrationPreservesFallbackCopies|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestSkillDirectoryLifecycleVectors)$'
```

`focused-projection-final` — exit `0`; `$WORKTREE`; 37.32s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestSkillProjection|TestProfileSkillProjection)'
```

`focused-projection` — exit `1`; `$WORKTREE`; 6.05s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^(TestSkillProjectionCoreSelection|TestProfileSkillProjectionPublication|TestProfileSkillProjectionRepair)$'
```

`focused-store` — exit `0`; `$WORKTREE`; 3.31s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^(TestSkillProjectionStoreReuse|TestProfileSkillProjectionMissingStore)$'
```

`format-2` — exit `0`; `$WORKTREE`; 0.09s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-3` — exit `0`; `$WORKTREE`; 0.15s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-4` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-5` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-cli-control-final` — exit `0`; `$WORKTREE`; 0.11s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./cmd/curator
```

`format-cli-control` — exit `0`; `$WORKTREE`; 0.2s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./cmd/curator
```

`format-copy-last` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-copy-repair` — exit `0`; `$WORKTREE`; 0.21s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-final` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-fixture` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-framing-final` — exit `0`; `$WORKTREE`; 0.1s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-framing-fixtures` — exit `0`; `$WORKTREE`; 0.1s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-framing` — exit `0`; `$WORKTREE`; 0.2s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-last` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format-network-fixture` — exit `0`; `$WORKTREE`; 0.09s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`format` — exit `0`; `$WORKTREE`; 0.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" fmt ./internal/envprofile
```

`framing-valid-skill-fixtures` — exit `0`; `$WORKTREE`; 9.8s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestInstall(BlocksBothV1CollidingSkillTrees|AdmitsV1CollidingSkillTreesUnderV2)$'
```

`full-build` — exit `0`; `$VALIDATION`; 0.94s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```

`full-curator-01` — exit `0`; `$VALIDATION`; 56.19s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-02` — exit `0`; `$VALIDATION`; 71.52s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-03` — exit `0`; `$VALIDATION`; 140.85s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-04` — exit `0`; `$VALIDATION`; 82.46s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-05` — exit `0`; `$VALIDATION`; 12.48s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestEnvConfigShowPermissionsWithoutProfiles|TestEnvConfigShowDefaults|TestEnvConfigSetUnsetRoundTrip|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestEnvConfigProviderDirectoriesRoundTrip|TestEnvConfigProviderDirectoriesLockedRefuses|TestEnvConfigLockedKnobRefuses|TestProfileComposeAddListRemove|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestProfileComposeRefusals|TestProfileUseLockedRequireRefuses|TestProfileInstallUseLockedRequireRefuses|TestProfileInstallFirstActivationLockedRequireRefuses|TestProfileUseClearOperandIsUsage|TestProfileScopedUseUnaffectedByLockedRequire|TestProfileImportUseLockedRequireRefuses|TestProfileUpdateResyncPassesLockedRequire|TestEnvStatusReportsLockedRequire|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsNotInheritedNativeCodexServers|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-06` — exit `0`; `$VALIDATION`; 75.92s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree|TestGCPreservesRuntimeOnUncertainMarks|TestGCRetainsLiveProcessBuild|TestGCPrunesDeadConsumersUnderTheHomeLock|TestGCFailsClosedForUntrustedCurrentPathSource|TestGCWaitsForTheHomeLock|TestGCRunsSerializedAcrossConcurrentInvocations|TestSignerRowsSurviveHostileGlobalGitConfig|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestGlobalStatusRejectsPositionalArguments|TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord|TestStatusUnreadableApprovalStateSurfaced|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApprovalsUnreadableStateNamesReadFailure|TestHookApproveRecordsOperatorApproval|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestHookApproveReRecordsAfterChange|TestHookApproveIsIdempotentWhenRecordMatches|TestHookApproveResolvesSymlinkedOperand|TestHookApprovalsListsReadOnly)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-07` — exit `0`; `$VALIDATION`; 65.74s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestHookApprovalsToleratesMalformedRecord|TestHookApprovalsOnAbsentStateIsEmpty|TestHookRevokeRemovesRecord|TestHookRevokeAbsentLeavesStateByteIdentical|TestHookUsageErrors|TestHookDiagnosticsAgreeWithShell|TestStatusReportsShellHookTrustPosture|TestStatusJSONCarriesShellHookTrust|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestHookApproveEndToEndWithGeneratedHook|TestProfileImportRow|TestProfileImportLossyRow|TestProfileImportNameTakenRow|TestProfileComposeAddPathRow|TestProfileComposeAddRefusesAFormOnAPathSource|TestPathOverlayFromMachineConfigJoinsTheClosure|TestOverlayFromMachineConfigIsRefusedByKind|TestProfileUseTakeoverRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestResolveTakeoverRequiresRepair|TestSyncTakeoverFlagParses|TestInstallGitignoreEntryPoint|TestLegacyPackageStatusBindingAndRepair|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestRunUsesInjectedConfigSourceAndWriters|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestProductionExternalDepsBindTrustedGitAndAudit|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestRunVersionExitsZero|TestRunNoArgsPrintsUsage|TestRunUnknownCommand|TestShellInitPrintsHooks|TestShellInitInstallCachesHookWithoutConfig|TestSkillCheckOnTempDir)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-08` — exit `0`; `$VALIDATION`; 61.6s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallFlagsAcceptTrailingOptions|TestInstallAuditFlagAcceptsOptionalMode|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestStatusDriftDetectsContentTampering|TestBootstrapAndProjectCommands|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestBootstrapIfMissingCreatesAbsentConfig|TestBootstrapIfMissingRejectsForce|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestCLIEndToEndInstallStatusAndTamperCheck|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestProductionBinaryDispatchesNetNSProbe|TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSAddListRemove|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestConfigSubcommandDispatch|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextCopyFallback|TestMuseRootContextVectors|TestMuseRootContextDrift|TestMuseRootContextRefusals)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-09` — exit `0`; `$VALIDATION`; 150.28s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMuseCredentialPolicyStatus|TestMuseRootContextStatus|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestMuseProfileSeedSecretAndCredentialRefusals|TestMuseFragmentV3PublishedCases|TestMuseCLIAuthParentBoundary|TestMuseFragmentV4PublishedCases|TestMuseCLIStorePolicy|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverEveryRegisteredCommand|TestNativeSkillsTakeoverArgumentDiscardRegression|TestNativeSkillsTakeoverActivationBoundary|TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsCLIOperations|TestNativeSkillsTakeoverFlagBoundary|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestCLIAuditOpaqueNULFollowsHashVersion|TestAuditAllowWritesWriterVersionPinCarrier|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestStatusCheckAdmitsV2NULBearingInstall|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallActivationMatrix|TestProfileInstallStopThenRetryConformanceVector|TestProfileInstallReinstallUsePreservesProjectScope|TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-10` — exit `0`; `$VALIDATION`; 17.14s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGitFileURLHasNoBackslash|TestProfileInstallListUse|TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileUseUnknownIsAFailure|TestProfileRemoveInUseIsAFailure|TestProfileInstallRefConflictIsAFailure|TestProfileSyncCoversTheMachineScope|TestProfileComposeListsDeclaredOverlays|TestProfileUseUnknownEnvironmentIsAFailure|TestProfileUseTargetIsStageB|TestProfileUseTargetBoundToAdapter|TestProfileInstallWarnsOnSystemModule|TestProfileScopedUseAndClear|TestProfileUpdatePinnedTagIsUnchanged|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallFileOperandIsRefused|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallUseActivatesThroughSwitch|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording|TestSecurityPostureVectorsThroughCLI|TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningIsNotEmittedByEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestSecurityPostureCuratorStatusRows|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestSecurityPostureProfileInstallRefusesMCPDeclarations|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestSecurityPostureRevisionBDefaultRefusesEmptySources|TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist|TestSecurityPostureRevisionBDefaultEnvStatus|TestCLIHelperProcess)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-11` — exit `0`; `$VALIDATION`; 116.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCompiledProjectRepairsCorruptCompiledState|TestCompiledProjectRestoresCacheWhenCommitFails|TestStatusReportsATransitivelyResolvedCompiledCommand|TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands|TestGCRetainsAndReportsReferencedCompiledState|TestCompiledInstallFollowsTheNativeControlInventoryExactly|TestDryRunNeverClaimsACompletedCompilerCheck|TestStatusExplainsAnUnusableGoToolchain|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestCuratorStatusProviderPostureAndCheck|TestStatusAcceptsAnUnchangedLegacyMarkerSchema|TestHostGoToolchainIsSelectableOnAnInventoryPlatform|TestInstallPrintsTheRemedyAVersionManagerSelectionEarns|TestUmbrellaProviderResolutionVectors|TestActiveRevisionIsWarningRelease|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd|TestUmbrellaHostilePathPlantRefusedUnderB|TestUmbrellaTrustedRootSymlinkEscape|TestUmbrellaTrustRootSelectionBothRevisions|TestUmbrellaRefusedDirectoriesBothRevisions|TestUmbrellaSymlinkIntoManagedRefused|TestUmbrellaUnreadableRootNeverAbsence|TestUmbrellaMissingNamesRoots|TestUmbrellaWarningText|TestProviderPostureRows|TestProviderSubcommandParsing|TestExecutableCandidates|TestProviderInputsForHostScannedBinWarnsWithoutLedger|TestProviderInputsForHostPublishedBinRefuses|TestProviderPostureReportsUnreadablePublishedShimLedger|TestProviderInputsForHostExplicitBinRefuses|TestUmbrellaTrustRootCaseVariantIdentity)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-last` — exit `0`; `$VALIDATION`; 101.62s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestCompiledProjectStatusAndUntrustedRecovery$' github.com/relux-works/curator/cmd/curator
```

`full-envprofile-01` — exit `0`; `$VALIDATION`; 119.56s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-02` — exit `0`; `$VALIDATION`; 38.89s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-03` — exit `0`; `$VALIDATION`; 36.72s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-04` — exit `0`; `$VALIDATION`; 143.03s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-05` — exit `0`; `$VALIDATION`; 68.1s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestImportAbsentMarkerDetectsNative|TestImportCorruptMarkerIsLoss|TestImportUnreadableMarkerIsLoss|TestDotfileManagerVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestDotfileManagerTableMatchesSpecFixture|TestDotfileStatePathResolution|TestForeignManagerHintLstatDiscipline|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestManagedFallbackPermissions|TestManagedSeedPreservesNativePermissions|TestManagedHomeRepairTightensDirectories|TestManagedAuthoredFilesArePrivate|TestManagedSeedRefusalsPreserved|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveProvisionRepair|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestResolveStaleUnprovisioned|TestResolveUnknowns|TestResolveDefaultProfile|TestResolveDriftRepair|TestResolvePassthroughLiveness|TestClaudeRootAlwaysCopied|TestResolveClaudeProjectEntry|TestCodexKeyringAmbient|TestCodexUnreadableConfigFails|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormUnsupported|TestResolveLockContention|TestResolveFormats|TestProfileInstallMarkdownContextCLI)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-06` — exit `0`; `$VALIDATION`; 76.89s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileInstallMarkdownMCP|TestProfileInstallMarkdownScope|TestProfileInstallMarkdownCaseAliasesCLI|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureCredentialStoreAdmission|TestMaterializeCapturePendingDoesNotRecover|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureWaitReselectsAndTimeout|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureDoesNotProvision|TestMaterializeCaptureObjectBounds|TestMaterializeCaptureSerializesProfileWriters|TestMaterializeCaptureConcurrentRepair|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureAuthorizingDrift|TestMaterializeCaptureUnreadableJournal|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigratePiWrongTargetToAgentRoot|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateNoSecretCopies|TestMigrateConflictsRefuse|TestMigratePiRootInspectionFailure|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigrateOldRootBytesWarnAndProceed|TestMigratePendingNeedsNothing|TestMigrateUnknownOperands|TestMigrateInspectIsReadOnly|TestMigrateApplyRequiresPlan|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateMarkerDriftRefuses|TestMigrateFailedRelinkKeepsOldLink)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-07` — exit `0`; `$VALIDATION`; 102.73s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMigrateSyscallFailureRollsBack|TestMigrateInterruptedApplyRecovers|TestMigratePublishFailureRevertsLinks|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrateBrokenJournalRefuses|TestMigrateRecoveryRefusesUnexpectedTarget|TestMigrateRecoveryCleansOwnedTemp|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestMusePolicyExplicitMigration|TestMusePublishedLinkStates|TestMuseForkRefusedRegression|TestMuseAuthParentBoundaryAtResolve|TestMuseDirectoriesAndIsolation|TestMuseAuthRepairAdditionalStates|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMuseStatusOptionalProvisioning|TestMusePublishedRegistryLayout|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsNarrowingMutants|TestNativeSkillsUseAndSync|TestNativeSkillsTakeover|TestNativeSkillsLinkedParent|TestNativeSkillsLedger|TestNativeSkillsUnreadableMarker|TestNativeSkillsRepair|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsInstallAndUpdate|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestGitFileURLForKeepsVolumesOutOfTheHost|TestServeWritesTheGitSafeRemote)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-08` — exit `0`; `$VALIDATION`; 47.6s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember|TestResolveRefusesV1CommitPinnedContextNUL|TestResolveCleanV1CommitContextStaysCurrent|TestResolveAdmitsV2CommitPinnedContextNUL|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL|TestStrictAuditMemberRefusesV1NULBeforeHashing|TestValidateNamedStorePinsRefusesV1NULBeforeHashing|TestSkillsOfRefusesV1NULBeforeHashing|TestGuardedReadersV1NULRefusalComputesNoV1Identity|TestInstallAdmitsV1CollidingSkillTreesUnderV2|TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2|TestPrecedencePrimitivesDriveEmission|TestPrecedenceTieKeepsTopologicalOrder|TestPathOverlayJoinsClosure|TestOverlayExplicitWeightOverridesDefault|TestGitOverlayJoinsClosure|TestPathOverlayFormIsSourceInvalid|TestOverlayInvalidSourceKindIsSourceInvalid|TestSCPOverlayResolvesAsGit|TestOverlayRepeatedNameIsCompositionInvalid|TestOverlayDuplicateNameIsCompositionInvalid|TestForbiddenOverlaysResolveAlone|TestRemoveRefusesOverlayOfAnotherProfile|TestWeightRulesApplyInOrder|TestWeightRulesAllFourDisagree|TestWeightDuplicateRefused|TestWeightUnknownRefused|TestNonRootWeightsMapRefused|TestWeightConflictErrorAndWarning|TestPathInstallCapturesDirtyUntrackedInsideGit|TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-09` — exit `0`; `$VALIDATION`; 23.84s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPathRootAndImportedMCPDeclarationsRefused|TestPathOverlayMCPDeclarationRefused|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestResolveRejectsGroupWritableOnlyPathOverlay|TestMissingPathOperandIsPathMissing|TestMissingOverlayPathIsPathMissing|TestNonDirectoryPathIsSourceInvalid|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsUntrusted|TestFoldedPathsAreSourceInvalid|TestRootGitExcluded|TestUnreadablePathIsUnreadable|TestPathSnapshotImmutableAcrossUpdateSyncUse|TestPathUpdateAllSucceedsWithImportedProfile|TestUnreadablePathRootLeadsUnreadable|TestUpdatePathWithoutStatePinIsSourceInvalid|TestUpdatePathMissingSnapshotIsSourceInvalid|TestUpdatePathSnapshotNameMismatchIsSourceInvalid|TestPathReinstallBareAndNonCurrentMovesPin|TestUpdateBlocksSecretOverlayMember|TestReinstallBlocksSecretOverlayMember|TestUpdateBlocksRevokedOverlayMember|TestUpdateBlocksV1RevokedOverlayMemberUnderV1Writers|TestReinstallBlocksRevokedOverlayMember|TestPermissionsForProfileProperties|TestRepairPreserveLockVectors|TestPreserveLockRefusesBeforeStoreRebuild|TestPreserveLockRaceRefusesBeforeJournalRecovery|TestPreserveLockMigrationRefusalPrecedesBusyLock|TestPreserveLockRevalidationHoldsRepairLock|TestPlainRepairMigrationKeepsLockPrivate|TestSystemDeltaVectors)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-10` — exit `0`; `$VALIDATION`; 147.58s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestRC14MigrationRehashesLegacyIdentities|TestRC14ResolveRepairMigratesSiblingHomes|TestRC14IdentityMigrationRollsBackEveryEntry|TestRC14IdentityMigrationRefusesSiblingPlanDrift|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestRC14IdentityMigrationRecoversInterruptedCommit|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14IdentityMigrationPreservesFallbackCopies|TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestEnvironmentsReadFailureVectors|TestPassthroughReadFailuresAreUnknownAndNeverRepaired|TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestUnreadableLockRefusesUpdateWithoutRewriting|TestAbsentLockRemainsProfileUnknown|TestUnreadableCodexSeedDoesNotSelectAbsentDefault|TestReviewerRecoveryMarkerDrift|TestReviewerRecoveryPreservesRegularTemp|TestSkillDirectoryNarrowingMutants|TestSkillDirectoryLifecycleVectors|TestSkillDirectoryMarkdownDeclaredControl|TestSkillDirectoryMetadata|TestSkillProjectionNarrowingMutants|TestSkillProjectionCoreSelection|TestProfileSkillProjectionPublication|TestProfileSkillProjectionRepair|TestSkillProjectionStoreReuse|TestProfileSkillProjectionMissingStore|TestSkillProjectionFallbackRollback|TestSkillProjectionLegacyFraming|TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse|TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-11` — exit `0`; `$VALIDATION`; 18.14s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestUpdateConfirmationPostureVectorAtStatus|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusShadowAcknowledgment|TestStatusProfileMembersAndUnregistered|TestSeedUnreadableStopsBeforeFirstWrite|TestReservedSkillName|TestProfilePathCollision|TestOpencodeFormFallback|TestRepairFailedMissingStore|TestOpencodeXDGSeeds|TestClaudeSeedMergePreservesToolState|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestStoreBoundaryStatusVectorsDriveStatus|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestInstallEmitsSurfacingBeforePublication|TestUpdateEmitsSurfacingBeforePublication)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-12` — exit `0`; `$VALIDATION`; 15.63s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestReinstallEmitsSurfacingBeforePublication|TestSurfacingEmittedDespitePublicationFailure|TestSurfacingSinkWriteFailureIsNonFatal|TestInstallSurfacesMCPDeclarations|TestInstallSurfacingSilentWithoutMCP|TestUpdateSurfacesCandidateMCPSet|TestAllowlistWarningOperations|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent|TestResolvePassthroughKnobs|TestStatusS4Posture|TestSurfacingUnreadableManifest|TestUseWithoutTakeoverRefusesUnmanaged|TestUseTakeoverBacksUpAndNotifies|TestForeignSymlinkStopsSwitch|TestTakeoverWarnsDotfileHeuristic|TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn|TestReadOnlyCommandsNeverOnboard|TestRepairTakeoverProvisionsManagedHome|TestUnmanageRestorePlanRetainsMetadata|TestUnmanageRestoreRefusesLinkedParent|TestUnmanageRestoreSymlinkReplacesDestinationEntry|TestEnvironmentWriteNofollowVectors|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-last` — exit `0`; `$VALIDATION`; 41.71s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestProfileNetworkContactVectors$' github.com/relux-works/curator/internal/envprofile
```

`full-install-01` — exit `0`; `$VALIDATION`; 44.5s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' github.com/relux-works/curator/internal/install
```

`full-install-02` — exit `0`; `$VALIDATION`; 92.93s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' github.com/relux-works/curator/internal/install
```

`full-install-03` — exit `0`; `$VALIDATION`; 133.18s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' github.com/relux-works/curator/internal/install
```

`full-install-04` — exit `0`; `$VALIDATION`; 83.87s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestDraftInstallGitTamperRefused|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftTransportEnabled|TestDraftTransportPlanConversion|TestDraftTransportAuth|TestDraftPlanNeedsSSH|TestDraftPlanNamesProvider|TestDraftSSHWrapperBase|TestDraftManagerExecutable|TestAcquireDraftNetworkSelection|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestDefaultAcquireFetchesTheDeclaredURL|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAcquireDraftNetworkRevision2Refusals|TestRevision1ReaderRejectsV2Shapes|TestUserConfigIgnoredByResolvedLane|TestDraftTransportPlanV2Conversion|TestDraftPlanNeedsSSHRevision2|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestDeclarationGenerationIsContentAddressed|TestDocumentGenerationIsStableWhenUnreadable|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGitFixtureNeverRetriesNonZeroExit|TestGitFixtureNeverRetriesWrappedExitError|TestGitFixtureRetryPredicate|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestGitFixtureOuterEntryRunsRealGit|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution|TestGlobalInstallRecognizesAdoptedForwardingShim|TestGlobalReconcilesOwnedLegacyForwardingLink|TestEndToEndInstall|TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestSecondInstallIsUpToDate|TestTamperTriggersReinstall|TestRemovedSkillCleanedUp|TestDryRunTouchesNothing)$' github.com/relux-works/curator/internal/install
```

`full-install-05` — exit `0`; `$VALIDATION`; 178.71s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGitignoreGateSkips|TestMissingSystemCommandFails|TestMovedTagWarningAndStrict|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestRuntimeOnlyProviderStillRequiresSkillMd|TestHybridSkillActivatesWithoutTouchingProjectStore|TestHybridShadowedByProjectDeclaration|TestGlobalInstall|TestGlobalInstallRejectsSchema2Skillfile|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestGlobalStrictTagsDetectMovedTag|TestMcpRequirementGatesInstall|TestAuditGateBlocksUndeclaredNetwork|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestLegacyLaneRecordsDependencyDirectory|TestLegacyLockPackageArms|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestLegacyLaneRefusesGlobDependencyDirectory|TestLegacyLaneRejectsUnboundPreviousGeneration|TestLegacyLaneDependencyTagMovement|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestReviewChangedDeclaredTagIsNotMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestLegacyLaneChangedTagAfterOldTagDeleted|TestLegacyLaneRuntimeMatchesPackageMarker|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPostCommitMaintenanceWarningsReachTheResult|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestASchemaSixCommandDeclaresNoModules|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV1ReinstallRefusesNULAppearingAfterInstall|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestV2InstallDistinguishesNULCollisionTwins|TestDraftLaneStillBlocksNULBearingMember|TestInstallDraftOwnershipCases|TestInstallPreservesUnownedStaleStoreEntries|TestInstallRefusesUnownedSelectedCanonicalDestination|TestInstallRefusesUnprovenStaleOwnership|TestInstallOwnedCanonicalPositiveControls|TestGlobalCanonicalOwnershipPreservation|TestHybridCanonicalOwnershipPreservation|TestInstallDraftMarkerWriterCases|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestV1InstallRejectsV2RegistryEvidence)$' github.com/relux-works/curator/internal/install
```

`full-install-06` — exit `0`; `$VALIDATION`; 99.65s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestV2RegistryAttestationUsesVersionedInstallPath|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistryBootstrapVectorsDriveInstall|TestStrictRegistryPolicyFailsUnknown|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStableDeclarationInputsCommitWithoutRestarting|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStageProjectTargetsRefusesMissingSnapshot|TestStageGlobalTargetsPopulatesBoundaries|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestRunCommitWithRealStagingJournalsOneTransaction|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged|TestDryRunReportsCacheHitWithoutBuilding|TestCacheHitPerformsNoSourceAwareGoCommand|TestStagingRunsProviderFirstAndCommandLexical|TestStagedOutputsStayPrivateAndAreReleased|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestCacheHitOnlyPlanStillFinalizesToolchainTrust|TestCacheInspectionIsBracketedByTheFrozenSource|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestCorruptCacheEntryIsRebuiltAndNeverReused|TestUnsupportedCacheProtectionFailsClosed|TestIncompletePlansNeverClaimACompleteDiagnostic|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$' github.com/relux-works/curator/internal/install
```

`full-install-07` — exit `0`; `$VALIDATION`; 7.87s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScriptOnlyInstallPerformsNoToolchainOrCacheWork|TestInjectedClockAndGenerationReaderDriveInstallMarkers|TestPlannedBuildAccessorsAreImmutable|TestReleaseTakesNoToolchainVerdictAfterLiveMutation|TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation|TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall|TestDefaultToolchainProbeRemovesItsProbeRootOnFailure|TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess|TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure|TestProjectWriteBlobsSpawnFailureIsWrapped)$' github.com/relux-works/curator/internal/install
```

`full-install-last` — exit `0`; `$VALIDATION`; 1.32s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$' github.com/relux-works/curator/internal/install
```

`full-other-01` — exit `0`; `$VALIDATION`; 140.55s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` — exit `0`; `$VALIDATION`; 129.31s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-other-03` — exit `0`; `$VALIDATION`; 90.92s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/gitops github.com/relux-works/curator/internal/globalbins github.com/relux-works/curator/internal/godriver github.com/relux-works/curator/internal/hashing github.com/relux-works/curator/internal/hookapproval github.com/relux-works/curator/internal/identifiers github.com/relux-works/curator/internal/identity github.com/relux-works/curator/internal/install/atomicity github.com/relux-works/curator/internal/interop github.com/relux-works/curator/internal/interop/environments github.com/relux-works/curator/internal/locale github.com/relux-works/curator/internal/managerlock github.com/relux-works/curator/internal/manifest github.com/relux-works/curator/internal/marker github.com/relux-works/curator/internal/mcp
```

`full-other-04` — exit `0`; `$VALIDATION`; 105.13s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/moduleroots github.com/relux-works/curator/internal/nodesource github.com/relux-works/curator/internal/npmsource github.com/relux-works/curator/internal/opaquescan github.com/relux-works/curator/internal/pathboundary github.com/relux-works/curator/internal/pkgversion github.com/relux-works/curator/internal/pnpmsource github.com/relux-works/curator/internal/privatedir github.com/relux-works/curator/internal/protocoljson github.com/relux-works/curator/internal/registry github.com/relux-works/curator/internal/runtimestore github.com/relux-works/curator/internal/rustsource github.com/relux-works/curator/internal/scopes github.com/relux-works/curator/internal/scriptpolicy github.com/relux-works/curator/internal/scriptworker
```

`full-other-05` — exit `0`; `$VALIDATION`; 38.7s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/shell github.com/relux-works/curator/internal/skillcheck github.com/relux-works/curator/internal/skillspec github.com/relux-works/curator/internal/snapshot github.com/relux-works/curator/internal/sourcelock github.com/relux-works/curator/internal/staging github.com/relux-works/curator/internal/stateread github.com/relux-works/curator/internal/swiftpmbuild github.com/relux-works/curator/internal/swiftpminterop github.com/relux-works/curator/internal/swiftpmsource github.com/relux-works/curator/internal/testcli github.com/relux-works/curator/internal/testgitenv github.com/relux-works/curator/internal/testtoolchain github.com/relux-works/curator/internal/transaction github.com/relux-works/curator/internal/ui
```

`full-other-06` — exit `0`; `$VALIDATION`; 44.57s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/verr github.com/relux-works/curator/internal/version github.com/relux-works/curator/internal/whitelist github.com/relux-works/curator/internal/yarnclassicsource github.com/relux-works/curator/internal/yarnmodernsource github.com/relux-works/curator/tools/ciworkflow github.com/relux-works/curator/tools/goreleaserconfig
```

`full-vet` — exit `0`; `$VALIDATION`; 0.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./...
```

`git-version` — exit `0`; `$WORKTREE`; 0.01s.

```sh
git --version
```

`go-version` — exit `0`; `$WORKTREE`; 0.01s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" version
```

`inventory` — exit `0`; `$VALIDATION`; 4.83s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -list . -timeout=8m ./...
```

`kind-capability-positive-fixture` — exit `0`; `$WORKTREE`; 4.8s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestInstalledMemberCapabilitiesStayKindScoped$'
```

`network-valid-skill-fixture` — exit `0`; `$WORKTREE`; 38.68s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^TestProfileNetworkContactVectors$'
```

`node-version` — exit `0`; `$WORKTREE`; 0.55s.

```sh
node --version
```

`probe-managed-after` — exit `0`; `$WORKTREE`; 6.01s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-managed-before` — exit `1`; `$WORKTREE`; 6.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-managed-exact-head` — exit `0`; `$VALIDATION`; 8.56s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-native-after` — exit `0`; `$WORKTREE`; 5.46s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

`probe-native-before` — exit `1`; `$WORKTREE`; 15.09s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

`probe-native-exact-head` — exit `0`; `$VALIDATION`; 7.56s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

`projection-final-controls` — exit `0`; `$WORKTREE`; 102.98s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestSkillProjection.*|TestProfileSkillProjection.*)$'
```

`projection-framing` — exit `0`; `$WORKTREE`; 9.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestSkillProjectionLegacyFraming$'
```

`projection-with-fallback-verification` — exit `0`; `$WORKTREE`; 35.5s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -json -timeout=8m ./internal/envprofile -run '^(TestSkillProjection|TestProfileSkillProjection)'
```

`rust-version` — exit `0`; `$WORKTREE`; 0.24s.

```sh
rustc --version
```

### preliminary-736589d0

`conformance-core` — exit `0`; `$VALIDATION`; 36.44s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json ./internal/audit ./internal/skillcheck ./internal/whitelist ./internal/contextpkg ./internal/contextlock ./internal/contextmaterialize
```

`conformance-env-1` — exit `0`; `$VALIDATION`; 227.45s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestClaudeSeedMergePreservesToolState|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCodexSeedStripsInlineMCPTable|TestDotfileManagerVectors|TestEnvironmentWriteNofollowVectors|TestEnvironmentsCodexSeedVectors|TestEnvironmentsEnvPassthroughVectors|TestEnvironmentsReadFailureVectors|TestMuseAuthParentBoundaryAtResolve|TestMuseAuthRepairAdditionalStates|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMusePublishedLinkStates|TestMusePublishedRegistryLayout|TestMuseStatusOptionalProvisioning|TestOpencodeFormFallback|TestOpencodeXDGSeeds)$' ./internal/envprofile
```

`conformance-env-2` — exit `0`; `$VALIDATION`; 60.65s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathKindMCPVectorsDriveInstallEntry|TestProfilePathCollision|TestRepairFailedMissingStore|TestReservedSkillName|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestSeedUnreadableStopsBeforeFirstWrite|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusProfileMembersAndUnregistered|TestStatusShadowAcknowledgment|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair)$' ./internal/envprofile
```

`conformance-env-3` — exit `0`; `$VALIDATION`; 24.22s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryStatusVectorsDriveStatus|TestSystemDeltaVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateConfirmationPostureVectorAtStatus)$' ./internal/envprofile
```

`full-build` — exit `0`; `$VALIDATION`; 1.06s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```

`full-curator-01` — exit `0`; `$VALIDATION`; 145.83s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-02` — exit `0`; `$VALIDATION`; 79.92s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-03` — exit `0`; `$VALIDATION`; 159.69s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-04` — exit `0`; `$VALIDATION`; 89.84s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-05` — exit `0`; `$VALIDATION`; 19.17s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestEnvConfigShowPermissionsWithoutProfiles|TestEnvConfigShowDefaults|TestEnvConfigSetUnsetRoundTrip|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestEnvConfigProviderDirectoriesRoundTrip|TestEnvConfigProviderDirectoriesLockedRefuses|TestEnvConfigLockedKnobRefuses|TestProfileComposeAddListRemove|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestProfileComposeRefusals|TestProfileUseLockedRequireRefuses|TestProfileInstallUseLockedRequireRefuses|TestProfileInstallFirstActivationLockedRequireRefuses|TestProfileUseClearOperandIsUsage|TestProfileScopedUseUnaffectedByLockedRequire|TestProfileImportUseLockedRequireRefuses|TestProfileUpdateResyncPassesLockedRequire|TestEnvStatusReportsLockedRequire|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsNotInheritedNativeCodexServers|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-06` — exit `0`; `$VALIDATION`; 79.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree|TestGCPreservesRuntimeOnUncertainMarks|TestGCRetainsLiveProcessBuild|TestGCPrunesDeadConsumersUnderTheHomeLock|TestGCFailsClosedForUntrustedCurrentPathSource|TestGCWaitsForTheHomeLock|TestGCRunsSerializedAcrossConcurrentInvocations|TestSignerRowsSurviveHostileGlobalGitConfig|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestGlobalStatusRejectsPositionalArguments|TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord|TestStatusUnreadableApprovalStateSurfaced|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApprovalsUnreadableStateNamesReadFailure|TestHookApproveRecordsOperatorApproval|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestHookApproveReRecordsAfterChange|TestHookApproveIsIdempotentWhenRecordMatches|TestHookApproveResolvesSymlinkedOperand|TestHookApprovalsListsReadOnly)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-07` — exit `0`; `$VALIDATION`; 80.97s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestHookApprovalsToleratesMalformedRecord|TestHookApprovalsOnAbsentStateIsEmpty|TestHookRevokeRemovesRecord|TestHookRevokeAbsentLeavesStateByteIdentical|TestHookUsageErrors|TestHookDiagnosticsAgreeWithShell|TestStatusReportsShellHookTrustPosture|TestStatusJSONCarriesShellHookTrust|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestHookApproveEndToEndWithGeneratedHook|TestProfileImportRow|TestProfileImportLossyRow|TestProfileImportNameTakenRow|TestProfileComposeAddPathRow|TestProfileComposeAddRefusesAFormOnAPathSource|TestPathOverlayFromMachineConfigJoinsTheClosure|TestOverlayFromMachineConfigIsRefusedByKind|TestProfileUseTakeoverRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestResolveTakeoverRequiresRepair|TestSyncTakeoverFlagParses|TestInstallGitignoreEntryPoint|TestLegacyPackageStatusBindingAndRepair|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestRunUsesInjectedConfigSourceAndWriters|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestProductionExternalDepsBindTrustedGitAndAudit|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestRunVersionExitsZero|TestRunNoArgsPrintsUsage|TestRunUnknownCommand|TestShellInitPrintsHooks|TestShellInitInstallCachesHookWithoutConfig|TestSkillCheckOnTempDir)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-08` — exit `0`; `$VALIDATION`; 71.8s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallFlagsAcceptTrailingOptions|TestInstallAuditFlagAcceptsOptionalMode|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestStatusDriftDetectsContentTampering|TestBootstrapAndProjectCommands|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestBootstrapIfMissingCreatesAbsentConfig|TestBootstrapIfMissingRejectsForce|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestCLIEndToEndInstallStatusAndTamperCheck|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestProductionBinaryDispatchesNetNSProbe|TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSAddListRemove|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestConfigSubcommandDispatch|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextCopyFallback|TestMuseRootContextVectors|TestMuseRootContextDrift|TestMuseRootContextRefusals)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-09` — exit `0`; `$VALIDATION`; 165.17s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMuseCredentialPolicyStatus|TestMuseRootContextStatus|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestMuseProfileSeedSecretAndCredentialRefusals|TestMuseFragmentV3PublishedCases|TestMuseCLIAuthParentBoundary|TestMuseFragmentV4PublishedCases|TestMuseCLIStorePolicy|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverEveryRegisteredCommand|TestNativeSkillsTakeoverArgumentDiscardRegression|TestNativeSkillsTakeoverActivationBoundary|TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsCLIOperations|TestNativeSkillsTakeoverFlagBoundary|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestCLIAuditOpaqueNULFollowsHashVersion|TestAuditAllowWritesWriterVersionPinCarrier|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestStatusCheckAdmitsV2NULBearingInstall|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallActivationMatrix|TestProfileInstallStopThenRetryConformanceVector|TestProfileInstallReinstallUsePreservesProjectScope|TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-10` — exit `0`; `$VALIDATION`; 19.19s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGitFileURLHasNoBackslash|TestProfileInstallListUse|TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileUseUnknownIsAFailure|TestProfileRemoveInUseIsAFailure|TestProfileInstallRefConflictIsAFailure|TestProfileSyncCoversTheMachineScope|TestProfileComposeListsDeclaredOverlays|TestProfileUseUnknownEnvironmentIsAFailure|TestProfileUseTargetIsStageB|TestProfileUseTargetBoundToAdapter|TestProfileInstallWarnsOnSystemModule|TestProfileScopedUseAndClear|TestProfileUpdatePinnedTagIsUnchanged|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallFileOperandIsRefused|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallUseActivatesThroughSwitch|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording|TestSecurityPostureVectorsThroughCLI|TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningIsNotEmittedByEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestSecurityPostureCuratorStatusRows|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestSecurityPostureProfileInstallRefusesMCPDeclarations|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestSecurityPostureRevisionBDefaultRefusesEmptySources|TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist|TestSecurityPostureRevisionBDefaultEnvStatus|TestCLIHelperProcess)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-11` — exit `0`; `$VALIDATION`; 125.36s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCompiledProjectRepairsCorruptCompiledState|TestCompiledProjectRestoresCacheWhenCommitFails|TestStatusReportsATransitivelyResolvedCompiledCommand|TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands|TestGCRetainsAndReportsReferencedCompiledState|TestCompiledInstallFollowsTheNativeControlInventoryExactly|TestDryRunNeverClaimsACompletedCompilerCheck|TestStatusExplainsAnUnusableGoToolchain|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestCuratorStatusProviderPostureAndCheck|TestStatusAcceptsAnUnchangedLegacyMarkerSchema|TestHostGoToolchainIsSelectableOnAnInventoryPlatform|TestInstallPrintsTheRemedyAVersionManagerSelectionEarns|TestUmbrellaProviderResolutionVectors|TestActiveRevisionIsWarningRelease|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd|TestUmbrellaHostilePathPlantRefusedUnderB|TestUmbrellaTrustedRootSymlinkEscape|TestUmbrellaTrustRootSelectionBothRevisions|TestUmbrellaRefusedDirectoriesBothRevisions|TestUmbrellaSymlinkIntoManagedRefused|TestUmbrellaUnreadableRootNeverAbsence|TestUmbrellaMissingNamesRoots|TestUmbrellaWarningText|TestProviderPostureRows|TestProviderSubcommandParsing|TestExecutableCandidates|TestProviderInputsForHostScannedBinWarnsWithoutLedger|TestProviderInputsForHostPublishedBinRefuses|TestProviderPostureReportsUnreadablePublishedShimLedger|TestProviderInputsForHostExplicitBinRefuses|TestUmbrellaTrustRootCaseVariantIdentity)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-last` — exit `0`; `$VALIDATION`; 98.86s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestCompiledProjectStatusAndUntrustedRecovery$' github.com/relux-works/curator/cmd/curator
```

`full-envprofile-01` — exit `0`; `$VALIDATION`; 214.16s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-02` — exit `0`; `$VALIDATION`; 44.12s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-03` — exit `0`; `$VALIDATION`; 38.43s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-04` — exit `0`; `$VALIDATION`; 114.3s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-05` — exit `0`; `$VALIDATION`; 112.36s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestImportAbsentMarkerDetectsNative|TestImportCorruptMarkerIsLoss|TestImportUnreadableMarkerIsLoss|TestDotfileManagerVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestDotfileManagerTableMatchesSpecFixture|TestDotfileStatePathResolution|TestForeignManagerHintLstatDiscipline|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestManagedFallbackPermissions|TestManagedSeedPreservesNativePermissions|TestManagedHomeRepairTightensDirectories|TestManagedAuthoredFilesArePrivate|TestManagedSeedRefusalsPreserved|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveProvisionRepair|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestResolveStaleUnprovisioned|TestResolveUnknowns|TestResolveDefaultProfile|TestResolveDriftRepair|TestResolvePassthroughLiveness|TestClaudeRootAlwaysCopied|TestResolveClaudeProjectEntry|TestCodexKeyringAmbient|TestCodexUnreadableConfigFails|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormUnsupported|TestResolveLockContention|TestResolveFormats|TestProfileInstallMarkdownContextCLI)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-06` — exit `0`; `$VALIDATION`; 92.08s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileInstallMarkdownMCP|TestProfileInstallMarkdownScope|TestProfileInstallMarkdownCaseAliasesCLI|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureCredentialStoreAdmission|TestMaterializeCapturePendingDoesNotRecover|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureWaitReselectsAndTimeout|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureDoesNotProvision|TestMaterializeCaptureObjectBounds|TestMaterializeCaptureSerializesProfileWriters|TestMaterializeCaptureConcurrentRepair|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureAuthorizingDrift|TestMaterializeCaptureUnreadableJournal|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigratePiWrongTargetToAgentRoot|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateNoSecretCopies|TestMigrateConflictsRefuse|TestMigratePiRootInspectionFailure|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigrateOldRootBytesWarnAndProceed|TestMigratePendingNeedsNothing|TestMigrateUnknownOperands|TestMigrateInspectIsReadOnly|TestMigrateApplyRequiresPlan|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateMarkerDriftRefuses|TestMigrateFailedRelinkKeepsOldLink)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-07` — exit `0`; `$VALIDATION`; 110.9s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMigrateSyscallFailureRollsBack|TestMigrateInterruptedApplyRecovers|TestMigratePublishFailureRevertsLinks|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrateBrokenJournalRefuses|TestMigrateRecoveryRefusesUnexpectedTarget|TestMigrateRecoveryCleansOwnedTemp|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestMusePolicyExplicitMigration|TestMusePublishedLinkStates|TestMuseForkRefusedRegression|TestMuseAuthParentBoundaryAtResolve|TestMuseDirectoriesAndIsolation|TestMuseAuthRepairAdditionalStates|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMuseStatusOptionalProvisioning|TestMusePublishedRegistryLayout|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsNarrowingMutants|TestNativeSkillsUseAndSync|TestNativeSkillsTakeover|TestNativeSkillsLinkedParent|TestNativeSkillsLedger|TestNativeSkillsUnreadableMarker|TestNativeSkillsRepair|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsInstallAndUpdate|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestGitFileURLForKeepsVolumesOutOfTheHost|TestServeWritesTheGitSafeRemote)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-08` — exit `1`; `$VALIDATION`; 50.33s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember|TestResolveRefusesV1CommitPinnedContextNUL|TestResolveCleanV1CommitContextStaysCurrent|TestResolveAdmitsV2CommitPinnedContextNUL|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL|TestStrictAuditMemberRefusesV1NULBeforeHashing|TestValidateNamedStorePinsRefusesV1NULBeforeHashing|TestSkillsOfRefusesV1NULBeforeHashing|TestGuardedReadersV1NULRefusalComputesNoV1Identity|TestInstallAdmitsV1CollidingSkillTreesUnderV2|TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2|TestPrecedencePrimitivesDriveEmission|TestPrecedenceTieKeepsTopologicalOrder|TestPathOverlayJoinsClosure|TestOverlayExplicitWeightOverridesDefault|TestGitOverlayJoinsClosure|TestPathOverlayFormIsSourceInvalid|TestOverlayInvalidSourceKindIsSourceInvalid|TestSCPOverlayResolvesAsGit|TestOverlayRepeatedNameIsCompositionInvalid|TestOverlayDuplicateNameIsCompositionInvalid|TestForbiddenOverlaysResolveAlone|TestRemoveRefusesOverlayOfAnotherProfile|TestWeightRulesApplyInOrder|TestWeightRulesAllFourDisagree|TestWeightDuplicateRefused|TestWeightUnknownRefused|TestNonRootWeightsMapRefused|TestWeightConflictErrorAndWarning|TestPathInstallCapturesDirtyUntrackedInsideGit|TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-09` — exit `0`; `$VALIDATION`; 22.33s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPathRootAndImportedMCPDeclarationsRefused|TestPathOverlayMCPDeclarationRefused|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestResolveRejectsGroupWritableOnlyPathOverlay|TestMissingPathOperandIsPathMissing|TestMissingOverlayPathIsPathMissing|TestNonDirectoryPathIsSourceInvalid|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsUntrusted|TestFoldedPathsAreSourceInvalid|TestRootGitExcluded|TestUnreadablePathIsUnreadable|TestPathSnapshotImmutableAcrossUpdateSyncUse|TestPathUpdateAllSucceedsWithImportedProfile|TestUnreadablePathRootLeadsUnreadable|TestUpdatePathWithoutStatePinIsSourceInvalid|TestUpdatePathMissingSnapshotIsSourceInvalid|TestUpdatePathSnapshotNameMismatchIsSourceInvalid|TestPathReinstallBareAndNonCurrentMovesPin|TestUpdateBlocksSecretOverlayMember|TestReinstallBlocksSecretOverlayMember|TestUpdateBlocksRevokedOverlayMember|TestUpdateBlocksV1RevokedOverlayMemberUnderV1Writers|TestReinstallBlocksRevokedOverlayMember|TestPermissionsForProfileProperties|TestRepairPreserveLockVectors|TestPreserveLockRefusesBeforeStoreRebuild|TestPreserveLockRaceRefusesBeforeJournalRecovery|TestPreserveLockMigrationRefusalPrecedesBusyLock|TestPreserveLockRevalidationHoldsRepairLock|TestPlainRepairMigrationKeepsLockPrivate|TestSystemDeltaVectors)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-10` — exit `0`; `$VALIDATION`; 150.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestRC14MigrationRehashesLegacyIdentities|TestRC14ResolveRepairMigratesSiblingHomes|TestRC14IdentityMigrationRollsBackEveryEntry|TestRC14IdentityMigrationRefusesSiblingPlanDrift|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestRC14IdentityMigrationRecoversInterruptedCommit|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14IdentityMigrationPreservesFallbackCopies|TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestEnvironmentsReadFailureVectors|TestPassthroughReadFailuresAreUnknownAndNeverRepaired|TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestUnreadableLockRefusesUpdateWithoutRewriting|TestAbsentLockRemainsProfileUnknown|TestUnreadableCodexSeedDoesNotSelectAbsentDefault|TestReviewerRecoveryMarkerDrift|TestReviewerRecoveryPreservesRegularTemp|TestSkillDirectoryNarrowingMutants|TestSkillDirectoryLifecycleVectors|TestSkillDirectoryMarkdownDeclaredControl|TestSkillDirectoryMetadata|TestSkillProjectionNarrowingMutants|TestSkillProjectionCoreSelection|TestProfileSkillProjectionPublication|TestProfileSkillProjectionRepair|TestSkillProjectionStoreReuse|TestProfileSkillProjectionMissingStore|TestSkillProjectionFallbackRollback|TestSkillProjectionLegacyFraming|TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse|TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-11` — exit `0`; `$VALIDATION`; 14.83s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestUpdateConfirmationPostureVectorAtStatus|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusShadowAcknowledgment|TestStatusProfileMembersAndUnregistered|TestSeedUnreadableStopsBeforeFirstWrite|TestReservedSkillName|TestProfilePathCollision|TestOpencodeFormFallback|TestRepairFailedMissingStore|TestOpencodeXDGSeeds|TestClaudeSeedMergePreservesToolState|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestStoreBoundaryStatusVectorsDriveStatus|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestInstallEmitsSurfacingBeforePublication|TestUpdateEmitsSurfacingBeforePublication)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-12` — exit `0`; `$VALIDATION`; 13.88s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestReinstallEmitsSurfacingBeforePublication|TestSurfacingEmittedDespitePublicationFailure|TestSurfacingSinkWriteFailureIsNonFatal|TestInstallSurfacesMCPDeclarations|TestInstallSurfacingSilentWithoutMCP|TestUpdateSurfacesCandidateMCPSet|TestAllowlistWarningOperations|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent|TestResolvePassthroughKnobs|TestStatusS4Posture|TestSurfacingUnreadableManifest|TestUseWithoutTakeoverRefusesUnmanaged|TestUseTakeoverBacksUpAndNotifies|TestForeignSymlinkStopsSwitch|TestTakeoverWarnsDotfileHeuristic|TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn|TestReadOnlyCommandsNeverOnboard|TestRepairTakeoverProvisionsManagedHome|TestUnmanageRestorePlanRetainsMetadata|TestUnmanageRestoreRefusesLinkedParent|TestUnmanageRestoreSymlinkReplacesDestinationEntry|TestEnvironmentWriteNofollowVectors|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-last` — exit `1`; `$VALIDATION`; 38.98s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestProfileNetworkContactVectors$' github.com/relux-works/curator/internal/envprofile
```

`full-install-01` — exit `0`; `$VALIDATION`; 96.65s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' github.com/relux-works/curator/internal/install
```

`full-install-02` — exit `0`; `$VALIDATION`; 145.75s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' github.com/relux-works/curator/internal/install
```

`full-install-03` — exit `0`; `$VALIDATION`; 155.79s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' github.com/relux-works/curator/internal/install
```

`full-install-04` — exit `0`; `$VALIDATION`; 92.44s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestDraftInstallGitTamperRefused|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftTransportEnabled|TestDraftTransportPlanConversion|TestDraftTransportAuth|TestDraftPlanNeedsSSH|TestDraftPlanNamesProvider|TestDraftSSHWrapperBase|TestDraftManagerExecutable|TestAcquireDraftNetworkSelection|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestDefaultAcquireFetchesTheDeclaredURL|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAcquireDraftNetworkRevision2Refusals|TestRevision1ReaderRejectsV2Shapes|TestUserConfigIgnoredByResolvedLane|TestDraftTransportPlanV2Conversion|TestDraftPlanNeedsSSHRevision2|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestDeclarationGenerationIsContentAddressed|TestDocumentGenerationIsStableWhenUnreadable|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGitFixtureNeverRetriesNonZeroExit|TestGitFixtureNeverRetriesWrappedExitError|TestGitFixtureRetryPredicate|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestGitFixtureOuterEntryRunsRealGit|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution|TestGlobalInstallRecognizesAdoptedForwardingShim|TestGlobalReconcilesOwnedLegacyForwardingLink|TestEndToEndInstall|TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestSecondInstallIsUpToDate|TestTamperTriggersReinstall|TestRemovedSkillCleanedUp|TestDryRunTouchesNothing)$' github.com/relux-works/curator/internal/install
```

`full-install-05` — exit `0`; `$VALIDATION`; 214.22s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGitignoreGateSkips|TestMissingSystemCommandFails|TestMovedTagWarningAndStrict|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestRuntimeOnlyProviderStillRequiresSkillMd|TestHybridSkillActivatesWithoutTouchingProjectStore|TestHybridShadowedByProjectDeclaration|TestGlobalInstall|TestGlobalInstallRejectsSchema2Skillfile|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestGlobalStrictTagsDetectMovedTag|TestMcpRequirementGatesInstall|TestAuditGateBlocksUndeclaredNetwork|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestLegacyLaneRecordsDependencyDirectory|TestLegacyLockPackageArms|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestLegacyLaneRefusesGlobDependencyDirectory|TestLegacyLaneRejectsUnboundPreviousGeneration|TestLegacyLaneDependencyTagMovement|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestReviewChangedDeclaredTagIsNotMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestLegacyLaneChangedTagAfterOldTagDeleted|TestLegacyLaneRuntimeMatchesPackageMarker|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPostCommitMaintenanceWarningsReachTheResult|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestASchemaSixCommandDeclaresNoModules|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV1ReinstallRefusesNULAppearingAfterInstall|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestV2InstallDistinguishesNULCollisionTwins|TestDraftLaneStillBlocksNULBearingMember|TestInstallDraftOwnershipCases|TestInstallPreservesUnownedStaleStoreEntries|TestInstallRefusesUnownedSelectedCanonicalDestination|TestInstallRefusesUnprovenStaleOwnership|TestInstallOwnedCanonicalPositiveControls|TestGlobalCanonicalOwnershipPreservation|TestHybridCanonicalOwnershipPreservation|TestInstallDraftMarkerWriterCases|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestV1InstallRejectsV2RegistryEvidence)$' github.com/relux-works/curator/internal/install
```

`full-install-06` — exit `0`; `$VALIDATION`; 113.02s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestV2RegistryAttestationUsesVersionedInstallPath|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistryBootstrapVectorsDriveInstall|TestStrictRegistryPolicyFailsUnknown|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStableDeclarationInputsCommitWithoutRestarting|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStageProjectTargetsRefusesMissingSnapshot|TestStageGlobalTargetsPopulatesBoundaries|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestRunCommitWithRealStagingJournalsOneTransaction|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged|TestDryRunReportsCacheHitWithoutBuilding|TestCacheHitPerformsNoSourceAwareGoCommand|TestStagingRunsProviderFirstAndCommandLexical|TestStagedOutputsStayPrivateAndAreReleased|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestCacheHitOnlyPlanStillFinalizesToolchainTrust|TestCacheInspectionIsBracketedByTheFrozenSource|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestCorruptCacheEntryIsRebuiltAndNeverReused|TestUnsupportedCacheProtectionFailsClosed|TestIncompletePlansNeverClaimACompleteDiagnostic|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$' github.com/relux-works/curator/internal/install
```

`full-install-07` — exit `0`; `$VALIDATION`; 7.0s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScriptOnlyInstallPerformsNoToolchainOrCacheWork|TestInjectedClockAndGenerationReaderDriveInstallMarkers|TestPlannedBuildAccessorsAreImmutable|TestReleaseTakesNoToolchainVerdictAfterLiveMutation|TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation|TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall|TestDefaultToolchainProbeRemovesItsProbeRootOnFailure|TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess|TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure|TestProjectWriteBlobsSpawnFailureIsWrapped)$' github.com/relux-works/curator/internal/install
```

`full-install-last` — exit `0`; `$VALIDATION`; 1.19s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$' github.com/relux-works/curator/internal/install
```

`full-other-01` — exit `0`; `$VALIDATION`; 268.92s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` — exit `0`; `$VALIDATION`; 140.54s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-other-03` — exit `0`; `$VALIDATION`; 77.28s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/gitops github.com/relux-works/curator/internal/globalbins github.com/relux-works/curator/internal/godriver github.com/relux-works/curator/internal/hashing github.com/relux-works/curator/internal/hookapproval github.com/relux-works/curator/internal/identifiers github.com/relux-works/curator/internal/identity github.com/relux-works/curator/internal/install/atomicity github.com/relux-works/curator/internal/interop github.com/relux-works/curator/internal/interop/environments github.com/relux-works/curator/internal/locale github.com/relux-works/curator/internal/managerlock github.com/relux-works/curator/internal/manifest github.com/relux-works/curator/internal/marker github.com/relux-works/curator/internal/mcp
```

`full-other-04` — exit `0`; `$VALIDATION`; 111.9s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/moduleroots github.com/relux-works/curator/internal/nodesource github.com/relux-works/curator/internal/npmsource github.com/relux-works/curator/internal/opaquescan github.com/relux-works/curator/internal/pathboundary github.com/relux-works/curator/internal/pkgversion github.com/relux-works/curator/internal/pnpmsource github.com/relux-works/curator/internal/privatedir github.com/relux-works/curator/internal/protocoljson github.com/relux-works/curator/internal/registry github.com/relux-works/curator/internal/runtimestore github.com/relux-works/curator/internal/rustsource github.com/relux-works/curator/internal/scopes github.com/relux-works/curator/internal/scriptpolicy github.com/relux-works/curator/internal/scriptworker
```

`full-other-05` — exit `0`; `$VALIDATION`; 52.83s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/shell github.com/relux-works/curator/internal/skillcheck github.com/relux-works/curator/internal/skillspec github.com/relux-works/curator/internal/snapshot github.com/relux-works/curator/internal/sourcelock github.com/relux-works/curator/internal/staging github.com/relux-works/curator/internal/stateread github.com/relux-works/curator/internal/swiftpmbuild github.com/relux-works/curator/internal/swiftpminterop github.com/relux-works/curator/internal/swiftpmsource github.com/relux-works/curator/internal/testcli github.com/relux-works/curator/internal/testgitenv github.com/relux-works/curator/internal/testtoolchain github.com/relux-works/curator/internal/transaction github.com/relux-works/curator/internal/ui
```

`full-other-06` — exit `0`; `$VALIDATION`; 45.28s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/verr github.com/relux-works/curator/internal/version github.com/relux-works/curator/internal/whitelist github.com/relux-works/curator/internal/yarnclassicsource github.com/relux-works/curator/internal/yarnmodernsource github.com/relux-works/curator/tools/ciworkflow github.com/relux-works/curator/tools/goreleaserconfig
```

`full-vet` — exit `0`; `$VALIDATION`; 1.15s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./...
```

`inventory` — exit `0`; `$VALIDATION`; 4.96s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -list . -timeout=8m ./...
```

`probe-managed-exact-head` — exit `0`; `$VALIDATION`; 8.97s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-native-exact-head` — exit `0`; `$VALIDATION`; 7.45s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

### preliminary-931a98fd

`conformance-core` — exit `0`; `$VALIDATION`; 82.14s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json ./internal/audit ./internal/skillcheck ./internal/whitelist ./internal/contextpkg ./internal/contextlock ./internal/contextmaterialize
```

`conformance-env-1` — exit `0`; `$VALIDATION`; 190.58s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestClaudeSeedMergePreservesToolState|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCodexSeedStripsInlineMCPTable|TestDotfileManagerVectors|TestEnvironmentWriteNofollowVectors|TestEnvironmentsCodexSeedVectors|TestEnvironmentsEnvPassthroughVectors|TestEnvironmentsReadFailureVectors|TestMuseAuthParentBoundaryAtResolve|TestMuseAuthRepairAdditionalStates|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMusePublishedLinkStates|TestMusePublishedRegistryLayout|TestMuseStatusOptionalProvisioning|TestOpencodeFormFallback|TestOpencodeXDGSeeds)$' ./internal/envprofile
```

`conformance-env-2` — exit `0`; `$VALIDATION`; 72.69s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathKindMCPVectorsDriveInstallEntry|TestProfilePathCollision|TestRepairFailedMissingStore|TestReservedSkillName|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestSeedUnreadableStopsBeforeFirstWrite|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusProfileMembersAndUnregistered|TestStatusShadowAcknowledgment|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair)$' ./internal/envprofile
```

`conformance-env-3` — exit `0`; `$VALIDATION`; 42.42s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryStatusVectorsDriveStatus|TestSystemDeltaVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateConfirmationPostureVectorAtStatus)$' ./internal/envprofile
```

`full-build` — exit `0`; `$VALIDATION`; 0.98s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```

`full-curator-01` — exit `0`; `$VALIDATION`; 72.26s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-02` — exit `0`; `$VALIDATION`; 101.76s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-03` — exit `0`; `$VALIDATION`; 310.35s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-04` — exit `0`; `$VALIDATION`; 99.28s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-05` — exit `0`; `$VALIDATION`; 46.03s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestEnvConfigShowPermissionsWithoutProfiles|TestEnvConfigShowDefaults|TestEnvConfigSetUnsetRoundTrip|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestEnvConfigProviderDirectoriesRoundTrip|TestEnvConfigProviderDirectoriesLockedRefuses|TestEnvConfigLockedKnobRefuses|TestProfileComposeAddListRemove|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestProfileComposeRefusals|TestProfileUseLockedRequireRefuses|TestProfileInstallUseLockedRequireRefuses|TestProfileInstallFirstActivationLockedRequireRefuses|TestProfileUseClearOperandIsUsage|TestProfileScopedUseUnaffectedByLockedRequire|TestProfileImportUseLockedRequireRefuses|TestProfileUpdateResyncPassesLockedRequire|TestEnvStatusReportsLockedRequire|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsNotInheritedNativeCodexServers|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-06` — exit `0`; `$VALIDATION`; 110.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree|TestGCPreservesRuntimeOnUncertainMarks|TestGCRetainsLiveProcessBuild|TestGCPrunesDeadConsumersUnderTheHomeLock|TestGCFailsClosedForUntrustedCurrentPathSource|TestGCWaitsForTheHomeLock|TestGCRunsSerializedAcrossConcurrentInvocations|TestSignerRowsSurviveHostileGlobalGitConfig|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestGlobalStatusRejectsPositionalArguments|TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord|TestStatusUnreadableApprovalStateSurfaced|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApprovalsUnreadableStateNamesReadFailure|TestHookApproveRecordsOperatorApproval|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestHookApproveReRecordsAfterChange|TestHookApproveIsIdempotentWhenRecordMatches|TestHookApproveResolvesSymlinkedOperand|TestHookApprovalsListsReadOnly)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-07` — exit `0`; `$VALIDATION`; 98.9s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestHookApprovalsToleratesMalformedRecord|TestHookApprovalsOnAbsentStateIsEmpty|TestHookRevokeRemovesRecord|TestHookRevokeAbsentLeavesStateByteIdentical|TestHookUsageErrors|TestHookDiagnosticsAgreeWithShell|TestStatusReportsShellHookTrustPosture|TestStatusJSONCarriesShellHookTrust|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestHookApproveEndToEndWithGeneratedHook|TestProfileImportRow|TestProfileImportLossyRow|TestProfileImportNameTakenRow|TestProfileComposeAddPathRow|TestProfileComposeAddRefusesAFormOnAPathSource|TestPathOverlayFromMachineConfigJoinsTheClosure|TestOverlayFromMachineConfigIsRefusedByKind|TestProfileUseTakeoverRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestResolveTakeoverRequiresRepair|TestSyncTakeoverFlagParses|TestInstallGitignoreEntryPoint|TestLegacyPackageStatusBindingAndRepair|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestRunUsesInjectedConfigSourceAndWriters|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestProductionExternalDepsBindTrustedGitAndAudit|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestRunVersionExitsZero|TestRunNoArgsPrintsUsage|TestRunUnknownCommand|TestShellInitPrintsHooks|TestShellInitInstallCachesHookWithoutConfig|TestSkillCheckOnTempDir)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-08` — exit `1`; `$VALIDATION`; 89.02s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallFlagsAcceptTrailingOptions|TestInstallAuditFlagAcceptsOptionalMode|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestStatusDriftDetectsContentTampering|TestBootstrapAndProjectCommands|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestBootstrapIfMissingCreatesAbsentConfig|TestBootstrapIfMissingRejectsForce|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestCLIEndToEndInstallStatusAndTamperCheck|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestProductionBinaryDispatchesNetNSProbe|TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSAddListRemove|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestConfigSubcommandDispatch|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextCopyFallback|TestMuseRootContextVectors|TestMuseRootContextDrift|TestMuseRootContextRefusals)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-09` — exit `0`; `$VALIDATION`; 238.74s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMuseCredentialPolicyStatus|TestMuseRootContextStatus|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestMuseProfileSeedSecretAndCredentialRefusals|TestMuseFragmentV3PublishedCases|TestMuseCLIAuthParentBoundary|TestMuseFragmentV4PublishedCases|TestMuseCLIStorePolicy|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverEveryRegisteredCommand|TestNativeSkillsTakeoverArgumentDiscardRegression|TestNativeSkillsTakeoverActivationBoundary|TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsCLIOperations|TestNativeSkillsTakeoverFlagBoundary|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestCLIAuditOpaqueNULFollowsHashVersion|TestAuditAllowWritesWriterVersionPinCarrier|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestStatusCheckAdmitsV2NULBearingInstall|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallActivationMatrix|TestProfileInstallStopThenRetryConformanceVector|TestProfileInstallReinstallUsePreservesProjectScope|TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication)$' github.com/relux-works/curator/cmd/curator
```

`full-envprofile-01` — exit `0`; `$VALIDATION`; 164.19s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-02` — exit `0`; `$VALIDATION`; 66.89s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-03` — exit `0`; `$VALIDATION`; 64.52s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-04` — exit `0`; `$VALIDATION`; 175.54s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-05` — exit `0`; `$VALIDATION`; 136.54s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestImportAbsentMarkerDetectsNative|TestImportCorruptMarkerIsLoss|TestImportUnreadableMarkerIsLoss|TestDotfileManagerVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestDotfileManagerTableMatchesSpecFixture|TestDotfileStatePathResolution|TestForeignManagerHintLstatDiscipline|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestManagedFallbackPermissions|TestManagedSeedPreservesNativePermissions|TestManagedHomeRepairTightensDirectories|TestManagedAuthoredFilesArePrivate|TestManagedSeedRefusalsPreserved|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveProvisionRepair|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestResolveStaleUnprovisioned|TestResolveUnknowns|TestResolveDefaultProfile|TestResolveDriftRepair|TestResolvePassthroughLiveness|TestClaudeRootAlwaysCopied|TestResolveClaudeProjectEntry|TestCodexKeyringAmbient|TestCodexUnreadableConfigFails|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormUnsupported|TestResolveLockContention|TestResolveFormats|TestProfileInstallMarkdownContextCLI)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-06` — exit `0`; `$VALIDATION`; 144.21s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProfileInstallMarkdownMCP|TestProfileInstallMarkdownScope|TestProfileInstallMarkdownCaseAliasesCLI|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureCredentialStoreAdmission|TestMaterializeCapturePendingDoesNotRecover|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureWaitReselectsAndTimeout|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureDoesNotProvision|TestMaterializeCaptureObjectBounds|TestMaterializeCaptureSerializesProfileWriters|TestMaterializeCaptureConcurrentRepair|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureAuthorizingDrift|TestMaterializeCaptureUnreadableJournal|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigratePiWrongTargetToAgentRoot|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateNoSecretCopies|TestMigrateConflictsRefuse|TestMigratePiRootInspectionFailure|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigrateOldRootBytesWarnAndProceed|TestMigratePendingNeedsNothing|TestMigrateUnknownOperands|TestMigrateInspectIsReadOnly|TestMigrateApplyRequiresPlan|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateMarkerDriftRefuses|TestMigrateFailedRelinkKeepsOldLink)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-07` — exit `0`; `$VALIDATION`; 142.57s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestMigrateSyscallFailureRollsBack|TestMigrateInterruptedApplyRecovers|TestMigratePublishFailureRevertsLinks|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrateBrokenJournalRefuses|TestMigrateRecoveryRefusesUnexpectedTarget|TestMigrateRecoveryCleansOwnedTemp|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestMusePolicyExplicitMigration|TestMusePublishedLinkStates|TestMuseForkRefusedRegression|TestMuseAuthParentBoundaryAtResolve|TestMuseDirectoriesAndIsolation|TestMuseAuthRepairAdditionalStates|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMuseStatusOptionalProvisioning|TestMusePublishedRegistryLayout|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsNarrowingMutants|TestNativeSkillsUseAndSync|TestNativeSkillsTakeover|TestNativeSkillsLinkedParent|TestNativeSkillsLedger|TestNativeSkillsUnreadableMarker|TestNativeSkillsRepair|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsInstallAndUpdate|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestGitFileURLForKeepsVolumesOutOfTheHost|TestServeWritesTheGitSafeRemote)$' github.com/relux-works/curator/internal/envprofile
```

`full-install-01` — exit `0`; `$VALIDATION`; 49.42s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' github.com/relux-works/curator/internal/install
```

`full-install-02` — exit `0`; `$VALIDATION`; 139.63s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' github.com/relux-works/curator/internal/install
```

`full-install-03` — exit `0`; `$VALIDATION`; 177.77s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' github.com/relux-works/curator/internal/install
```

`full-install-04` — exit `0`; `$VALIDATION`; 74.18s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestDraftInstallGitTamperRefused|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftTransportEnabled|TestDraftTransportPlanConversion|TestDraftTransportAuth|TestDraftPlanNeedsSSH|TestDraftPlanNamesProvider|TestDraftSSHWrapperBase|TestDraftManagerExecutable|TestAcquireDraftNetworkSelection|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestDefaultAcquireFetchesTheDeclaredURL|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAcquireDraftNetworkRevision2Refusals|TestRevision1ReaderRejectsV2Shapes|TestUserConfigIgnoredByResolvedLane|TestDraftTransportPlanV2Conversion|TestDraftPlanNeedsSSHRevision2|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestDeclarationGenerationIsContentAddressed|TestDocumentGenerationIsStableWhenUnreadable|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGitFixtureNeverRetriesNonZeroExit|TestGitFixtureNeverRetriesWrappedExitError|TestGitFixtureRetryPredicate|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestGitFixtureOuterEntryRunsRealGit|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution|TestGlobalInstallRecognizesAdoptedForwardingShim|TestGlobalReconcilesOwnedLegacyForwardingLink|TestEndToEndInstall|TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestSecondInstallIsUpToDate|TestTamperTriggersReinstall|TestRemovedSkillCleanedUp|TestDryRunTouchesNothing)$' github.com/relux-works/curator/internal/install
```

`full-install-05` — exit `0`; `$VALIDATION`; 350.31s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestGitignoreGateSkips|TestMissingSystemCommandFails|TestMovedTagWarningAndStrict|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestRuntimeOnlyProviderStillRequiresSkillMd|TestHybridSkillActivatesWithoutTouchingProjectStore|TestHybridShadowedByProjectDeclaration|TestGlobalInstall|TestGlobalInstallRejectsSchema2Skillfile|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestGlobalStrictTagsDetectMovedTag|TestMcpRequirementGatesInstall|TestAuditGateBlocksUndeclaredNetwork|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestLegacyLaneRecordsDependencyDirectory|TestLegacyLockPackageArms|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestLegacyLaneRefusesGlobDependencyDirectory|TestLegacyLaneRejectsUnboundPreviousGeneration|TestLegacyLaneDependencyTagMovement|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestReviewChangedDeclaredTagIsNotMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestLegacyLaneChangedTagAfterOldTagDeleted|TestLegacyLaneRuntimeMatchesPackageMarker|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPostCommitMaintenanceWarningsReachTheResult|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestASchemaSixCommandDeclaresNoModules|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV1ReinstallRefusesNULAppearingAfterInstall|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestV2InstallDistinguishesNULCollisionTwins|TestDraftLaneStillBlocksNULBearingMember|TestInstallDraftOwnershipCases|TestInstallPreservesUnownedStaleStoreEntries|TestInstallRefusesUnownedSelectedCanonicalDestination|TestInstallRefusesUnprovenStaleOwnership|TestInstallOwnedCanonicalPositiveControls|TestGlobalCanonicalOwnershipPreservation|TestHybridCanonicalOwnershipPreservation|TestInstallDraftMarkerWriterCases|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestV1InstallRejectsV2RegistryEvidence)$' github.com/relux-works/curator/internal/install
```

`full-install-06` — exit `0`; `$VALIDATION`; 147.7s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestV2RegistryAttestationUsesVersionedInstallPath|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistryBootstrapVectorsDriveInstall|TestStrictRegistryPolicyFailsUnknown|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStableDeclarationInputsCommitWithoutRestarting|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStageProjectTargetsRefusesMissingSnapshot|TestStageGlobalTargetsPopulatesBoundaries|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestRunCommitWithRealStagingJournalsOneTransaction|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged|TestDryRunReportsCacheHitWithoutBuilding|TestCacheHitPerformsNoSourceAwareGoCommand|TestStagingRunsProviderFirstAndCommandLexical|TestStagedOutputsStayPrivateAndAreReleased|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestCacheHitOnlyPlanStillFinalizesToolchainTrust|TestCacheInspectionIsBracketedByTheFrozenSource|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestCorruptCacheEntryIsRebuiltAndNeverReused|TestUnsupportedCacheProtectionFailsClosed|TestIncompletePlansNeverClaimACompleteDiagnostic|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$' github.com/relux-works/curator/internal/install
```

`full-other-01` — exit `0`; `$VALIDATION`; 294.11s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` — exit `0`; `$VALIDATION`; 188.29s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-other-03` — exit `0`; `$VALIDATION`; 95.55s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/gitops github.com/relux-works/curator/internal/globalbins github.com/relux-works/curator/internal/godriver github.com/relux-works/curator/internal/hashing github.com/relux-works/curator/internal/hookapproval github.com/relux-works/curator/internal/identifiers github.com/relux-works/curator/internal/identity github.com/relux-works/curator/internal/install/atomicity github.com/relux-works/curator/internal/interop github.com/relux-works/curator/internal/interop/environments github.com/relux-works/curator/internal/locale github.com/relux-works/curator/internal/managerlock github.com/relux-works/curator/internal/manifest github.com/relux-works/curator/internal/marker github.com/relux-works/curator/internal/mcp
```

`full-other-04` — exit `0`; `$VALIDATION`; 127.69s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/moduleroots github.com/relux-works/curator/internal/nodesource github.com/relux-works/curator/internal/npmsource github.com/relux-works/curator/internal/opaquescan github.com/relux-works/curator/internal/pathboundary github.com/relux-works/curator/internal/pkgversion github.com/relux-works/curator/internal/pnpmsource github.com/relux-works/curator/internal/privatedir github.com/relux-works/curator/internal/protocoljson github.com/relux-works/curator/internal/registry github.com/relux-works/curator/internal/runtimestore github.com/relux-works/curator/internal/rustsource github.com/relux-works/curator/internal/scopes github.com/relux-works/curator/internal/scriptpolicy github.com/relux-works/curator/internal/scriptworker
```

`full-other-05` — exit `0`; `$VALIDATION`; 74.2s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/shell github.com/relux-works/curator/internal/skillcheck github.com/relux-works/curator/internal/skillspec github.com/relux-works/curator/internal/snapshot github.com/relux-works/curator/internal/sourcelock github.com/relux-works/curator/internal/staging github.com/relux-works/curator/internal/stateread github.com/relux-works/curator/internal/swiftpmbuild github.com/relux-works/curator/internal/swiftpminterop github.com/relux-works/curator/internal/swiftpmsource github.com/relux-works/curator/internal/testcli github.com/relux-works/curator/internal/testgitenv github.com/relux-works/curator/internal/testtoolchain github.com/relux-works/curator/internal/transaction github.com/relux-works/curator/internal/ui
```

`full-other-06` — exit `0`; `$VALIDATION`; 42.91s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/verr github.com/relux-works/curator/internal/version github.com/relux-works/curator/internal/whitelist github.com/relux-works/curator/internal/yarnclassicsource github.com/relux-works/curator/internal/yarnmodernsource github.com/relux-works/curator/tools/ciworkflow github.com/relux-works/curator/tools/goreleaserconfig
```

`full-vet` — exit `0`; `$VALIDATION`; 0.99s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./...
```

`inventory` — exit `0`; `$VALIDATION`; 8.83s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -list . -timeout=8m ./...
```

`probe-managed-exact-head` — exit `0`; `$VALIDATION`; 16.0s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-native-exact-head` — exit `0`; `$VALIDATION`; 13.92s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

### preliminary-997d5db4

`full-build` — exit `0`; `$VALIDATION`; 1.29s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```

`full-curator-01` — exit `0`; `$VALIDATION`; 81.07s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-02` — exit `0`; `$VALIDATION`; 95.91s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-03` — exit `0`; `$VALIDATION`; 173.19s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-04` — exit `0`; `$VALIDATION`; 87.32s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$' github.com/relux-works/curator/cmd/curator
```

`full-envprofile-01` — exit `0`; `$VALIDATION`; 173.37s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-02` — exit `0`; `$VALIDATION`; 59.45s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-03` — exit `0`; `$VALIDATION`; 45.48s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-04` — exit `1`; `$VALIDATION`; 161.41s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$' github.com/relux-works/curator/internal/envprofile
```

`full-install-01` — exit `0`; `$VALIDATION`; 60.4s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' github.com/relux-works/curator/internal/install
```

`full-install-02` — exit `0`; `$VALIDATION`; 135.12s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' github.com/relux-works/curator/internal/install
```

`full-install-03` — exit `0`; `$VALIDATION`; 188.13s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' github.com/relux-works/curator/internal/install
```

`full-other-01` — exit `0`; `$VALIDATION`; 247.86s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` — exit `0`; `$VALIDATION`; 169.91s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-vet` — exit `0`; `$VALIDATION`; 1.25s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./...
```

`inventory` — exit `0`; `$VALIDATION`; 8.49s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -list . -timeout=8m ./...
```

`probe-managed-exact-head` — exit `0`; `$VALIDATION`; 21.51s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-native-exact-head` — exit `0`; `$VALIDATION`; 23.42s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

### preliminary-f381e5de

`full-build` — exit `0`; `$VALIDATION`; 1.2s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```

`full-curator-01` — exit `0`; `$VALIDATION`; 62.49s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-02` — exit `0`; `$VALIDATION`; 76.35s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' github.com/relux-works/curator/cmd/curator
```

`full-curator-03` — exit `0`; `$VALIDATION`; 162.53s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' github.com/relux-works/curator/cmd/curator
```

`full-envprofile-01` — exit `0`; `$VALIDATION`; 127.28s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-02` — exit `0`; `$VALIDATION`; 42.51s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' github.com/relux-works/curator/internal/envprofile
```

`full-envprofile-03` — exit `0`; `$VALIDATION`; 40.55s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' github.com/relux-works/curator/internal/envprofile
```

`full-install-01` — exit `0`; `$VALIDATION`; 45.22s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' github.com/relux-works/curator/internal/install
```

`full-install-02` — exit `0`; `$VALIDATION`; 104.47s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' github.com/relux-works/curator/internal/install
```

`full-install-03` — exit `0`; `$VALIDATION`; 139.04s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' github.com/relux-works/curator/internal/install
```

`full-other-01` — exit `0`; `$VALIDATION`; 145.95s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` — exit `0`; `$VALIDATION`; 148.06s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-vet` — exit `0`; `$VALIDATION`; 1.96s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./...
```

`inventory` — exit `0`; `$VALIDATION`; 11.78s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -list . -timeout=8m ./...
```

`probe-managed-exact-head` — exit `0`; `$VALIDATION`; 14.63s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerManagedSkillProjection$'
```

`probe-native-exact-head` — exit `0`; `$VALIDATION`; 11.54s.

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -count=1 -v -timeout=8m ./internal/envprofile -run '^TestReviewerNativeSkillProjection$'
```

Delivery checks (owning checkout; every command exit 0):

```sh
git status --short
git fetch origin
git merge-base --is-ancestor origin/fix/audit-markdown-scope HEAD
git diff 2c89fd46443d461e25bd377fb94a0c6754282838..HEAD --check
git log 2c89fd46443d461e25bd377fb94a0c6754282838..HEAD --format='%H %G?'
git push -u origin fix/audit-markdown-scope
git ls-remote origin refs/heads/fix/audit-markdown-scope
```

The status output was empty, every signature was `G`, and the remote feature
head matched `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2`.
