# BUG-261021-2qv4zj results r3

Branch: `fix/audit-markdown-scope`  
Head: `d20c4363cf2dea88ce7418fa594664a466c116c8`  
R3 parent: `9b17768c96d53e9ef7babb45e95d3b688728e815`  
Signature: `G` (verified before publication). Board status remains owned by the orchestrator.

F2 is fixed. Profile source audits now carry the resolved package kind. Context audit and context materialization read declared modules through `contextpkg.ReadModule`; skill Markdown still uses the shared `whitelist.WalkContext` from r2. The selected package root and the non-Markdown scripts/skill-manifest scan are unchanged. Context module read failures produce a critical `audit.context.scan-incomplete` refusal. Ruleset **8** invalidates r2 allowances; cache equality also binds `subject_kind` so identical trees audited under different publication contracts cannot lend an allowance.

## Publication contract sweep

Production callers were traced through `internal/install/install.go`, `internal/install/global.go`, `internal/audit/sourceaudit.go`, and `internal/envprofile/envprofile.go`. The profile resolver/lock accepts the three kinds in `internal/contextlock/contextlock.go`. The Gate struct also permits empty and unknown kind strings.

| Subject kind | Agent publication | Audit coverage |
| --- | --- | --- |
| `skill` | Selected skill package, with the core section 3 prompt-context projection | Shared `whitelist.WalkContext`, including canonical filesystem root resolution, case-sensitive component exclusions, validated runtime/build roots and commandless scripts. Historical non-Markdown scripts and skill manifest remain scanned. |
| empty (legacy skill callers) | Same skill contract in project/global installation and draft source audit | Defaults to the same skill selection. Empty versus explicit kind conservatively misses the cache. |
| `context` | Declared root modules become native `CLAUDE.md`/`AGENTS.md`, referenced module surfaces or monolithic context; admitted system modules become system-prompt surfaces | Every declared module is read through the same `contextpkg.ReadModule` used by `loadMaterial` before `contextmaterialize` assembly. Coverage includes all classes/selectors, every extension and skill-excluded names. It is a superset of each adapter's selected output. Context manifests do not publish undeclared Markdown. |
| `mcp` | Server declarations/configuration, with no package Markdown context surface | Preserve the base all-Markdown detector, plus the historical non-Markdown scan. Do not apply skill exclusions. |
| unknown; context with an invalid/unavailable manifest | No resolver-supported publication contract can be established | Conservative base all-Markdown fallback, plus the historical non-Markdown scan. The profile validation path separately rejects invalid manifests. |

Context and MCP retain implicit-none capabilities (environments section 9.1). The orchestrator explicitly confirmed in the task topic that the brief's declared-host installation control applies to skills; context controls are clean and reserved-host modules. A direct context Gate with supplied capabilities remains a unit-level positive control. No new context capability schema was introduced.

## Changed files

- `docs/profile-skill-directories.md`
- `internal/audit/audit.go`
- `internal/audit/markdown_alias_test.go`
- `internal/audit/markdown_kind_test.go`
- `internal/audit/network_contacts_mutants_test.go`
- `internal/contextpkg/contextpkg.go`
- `internal/envprofile/capability_cache_test.go`
- `internal/envprofile/envprofile.go`
- `internal/envprofile/markdown_context_test.go`
- `internal/envprofile/switch.go`

## Claims and evidence

| Test or check | Claim proved by the exercised cases |
| --- | --- |
| `TestProfileInstallMarkdownContextCLI` | A built exact-source CLI with private HOME/USERPROFILE/native homes and strict/high/null configuration rejects the exact schema-1 `review-context` + `context/a.md` URL witness, exits 1 with `audit.capability.network-undeclared` and canonical module path, writes neither source.json nor lock.json, and puts no attack URL into native CLAUDE.md or AGENTS.md. Clean and reserved-host controls install, publish source/lock and place their bytes in both native documents. Protected fixture directories/files use 0700/0600. |
| `TestMarkdownScopeContextModules` | Both gates and cache paths reject undeclared contacts in ordinary, README, non-.md and selected system modules; clean and explicitly supplied host capabilities allow. These unit capabilities do not imply a context manifest declaration surface. |
| `TestMarkdownScopeContextMissingModule` | A missing declared module refuses with the scan-incomplete finding; restoring clean module bytes admits it. |
| `TestProfileInstallMarkdownMCP`, `TestMarkdownScopeOtherKinds` | The actual profile MCP member path retains base all-Markdown refusal outside skill roots. Clean/reserved MCP controls install. Direct MCP, unknown-kind and invalid-context fallback cases retain undeclared rejection and clean/declared allowance. |
| `TestMarkdownScopeCacheBindsKind` | A cached skill allowance for an identical raw tree/capability set cannot admit context, MCP or an unknown kind. New refusals are subsequently reused as cache hits. |
| `TestMarkdownScopeInvalidatesRulesetSevenAllowance` | Production Gate with an overlay restoring r2's detector selection and cache shape writes and hits a real empty ruleset-7 allowance without subject_kind. Current read-only/writable paths reject the identical context under both hash frames, cache the refusal, and retain the old record byte-for-byte. The old record is not handwritten. |
| All prior Markdown scope/cache/alias/exclusion tests | Nine root aliases, exclusion case parity, repository research/board/vendor noise, commandless/exported scripts, declared-host skill admission and rulesets 4/5/6 invalidation remain covered. |
| `TestSkillDirectory*`, pinned conformance | The CIP-0021 selected-directory witness and declared-host control remain green, with no sibling/root authority borrowing. |
| `TestNetworkContactNarrowingMutants` | Dropping the declared-module loop fails the real F2 CLI refusal assertion. The same overlay is explicitly carried into the child CLI build. Weakening kind equality fails the cache-boundary test; downgrading a module read failure fails its refusal assertion; reusing ruleset 7 overwrites the retained old record and fails the preservation assertion. All r2 mutation controls still fail at their intended assertions. Only exit 1 plus the named test/assertion is accepted as a kill. |
| `TestCapabilityCacheNarrowingMutants` | Existing effective-capability controls still fail as intended. Their profile cache fixture now supplies the real member kind so the capability mutation is not hidden by the independent new kind check. |
| Build / vet / default suite | Compile and static checks pass on the exact signed head. Every default package and every listed top-level Test/Example/Fuzz in the three large packages is exercised exactly once across the bounded shards. |

## Exact-head validation

Validation used a disposable ordinary `.git` clone at the handover head, with inherited GIT_DIR/GIT_WORK_TREE unset for every test command. Go 1.25.5 darwin/arm64; Rust 1.91.0; Node v26.12.0; Git 2.54.0 (Apple Git-157). Private HOME/USERPROFILE, GOCACHE, GOPATH and GOMODCACHE; TMPDIR and GOTMPDIR set to `$SHORT_TMP` (`/tmp/cw-dev3`). Ambient provider, manager, task-board, Git and XDG overrides were stripped. The test PATH contains the pinned Go toolchain, explicit Node helpers, the native Rust toolchain and system tools. No installed provider CLI is exposed. Module/cache scratch is removed after handover.

Default suite: **85 packages**, 80 passed and 5 had no test files; test/subtest results **{'pass': 11713, 'skip': 199}**. Conditional platform/conformance skips are reported as skips, not passing evidence. Pinned conformance supplement: **{'pass': 1147}**, with no skips; spec commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, root `$SPEC/conformance/v1`. All 45 envprofile conformance entry points were included in the supplement.

The CLI shards run sequentially because TestMain holds a shared toolchain guard. The install guard test runs after all CLI shards. No individual Go command exceeds the lane's ten-minute shell limit; each has `-timeout=8m` and a 550-second watchdog.

| Large package | Inventory | Observed top-level runs | Exactly once |
| --- | ---: | ---: | --- |
| `cmd/curator` | 434 | 434 | True |
| `internal/envprofile` | 401 | 401 | True |
| `internal/install` | 371 | 371 | True |

## Commands and exit codes

Each row below records an actual command in `$VALIDATION` on the exact head. `go` means the pinned `$CURATOR_GO` with the private environment above; every test invocation has the explicit prefix `env -u GIT_DIR -u GIT_WORK_TREE`. Conformance rows additionally set `CURATOR_CONFORMANCE_ROOT=$SPEC/conformance/v1`. Full `-run` selectors are preserved in the command ledger below.

| Invocation | Exit | Seconds |
| --- | ---: | ---: |
| `full-cli-01` | 0 | 57.82 |
| `full-cli-02` | 0 | 74.61 |
| `full-cli-03` | 0 | 150.81 |
| `full-cli-04` | 0 | 91.09 |
| `full-cli-05` | 0 | 13.46 |
| `full-cli-06` | 0 | 81.15 |
| `full-cli-07` | 0 | 69.45 |
| `full-cli-08` | 0 | 74.91 |
| `full-cli-09` | 0 | 168.7 |
| `full-cli-10` | 0 | 19.85 |
| `full-cli-11` | 0 | 140.73 |
| `full-cli-last` | 0 | 104.44 |
| `full-env-01` | 0 | 127.83 |
| `full-env-02` | 0 | 37.94 |
| `full-env-03` | 0 | 31.64 |
| `full-env-04` | 0 | 158.84 |
| `full-env-05` | 0 | 78.28 |
| `full-env-06` | 0 | 90.85 |
| `full-env-07` | 0 | 105.56 |
| `full-env-08` | 0 | 49.12 |
| `full-env-09` | 0 | 23.82 |
| `full-env-10` | 0 | 107.63 |
| `full-env-11` | 0 | 38.55 |
| `full-env-12` | 0 | 10.62 |
| `full-env-last` | 0 | 61.36 |
| `full-install-01` | 0 | 44.26 |
| `full-install-02` | 0 | 106.48 |
| `full-install-03` | 0 | 136.4 |
| `full-install-04` | 0 | 88.38 |
| `full-install-05` | 0 | 189.97 |
| `full-install-06` | 0 | 110.17 |
| `full-install-07` | 0 | 7.54 |
| `full-other-01` | 0 | 136.62 |
| `full-other-02` | 0 | 148.29 |
| `full-other-03` | 0 | 75.93 |
| `full-other-04` | 0 | 113.83 |
| `full-other-05` | 0 | 45.82 |
| `full-other-06` | 0 | 42.02 |
| `full-install-last` | 0 | 0.92 |
| `conformance-env-01` | 0 | 52.32 |
| `conformance-env-02` | 0 | 22.85 |
| `conformance-env-03` | 0 | 11.25 |
| `build` | 0 | 6.52 |
| `vet` | 0 | 6.34 |
| `conformance-core` | 0 | 44.57 |

```json
{
  "full-cli-01": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 57.82
  },
  "full-cli-02": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 74.61
  },
  "full-cli-03": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 150.81
  },
  "full-cli-04": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 91.09
  },
  "full-cli-05": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestEnvConfigShowPermissionsWithoutProfiles|TestEnvConfigShowDefaults|TestEnvConfigSetUnsetRoundTrip|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestEnvConfigProviderDirectoriesRoundTrip|TestEnvConfigProviderDirectoriesLockedRefuses|TestEnvConfigLockedKnobRefuses|TestProfileComposeAddListRemove|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestProfileComposeRefusals|TestProfileUseLockedRequireRefuses|TestProfileInstallUseLockedRequireRefuses|TestProfileInstallFirstActivationLockedRequireRefuses|TestProfileUseClearOperandIsUsage|TestProfileScopedUseUnaffectedByLockedRequire|TestProfileImportUseLockedRequireRefuses|TestProfileUpdateResyncPassesLockedRequire|TestEnvStatusReportsLockedRequire|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsNotInheritedNativeCodexServers|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 13.46
  },
  "full-cli-06": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree|TestGCPreservesRuntimeOnUncertainMarks|TestGCRetainsLiveProcessBuild|TestGCPrunesDeadConsumersUnderTheHomeLock|TestGCFailsClosedForUntrustedCurrentPathSource|TestGCWaitsForTheHomeLock|TestGCRunsSerializedAcrossConcurrentInvocations|TestSignerRowsSurviveHostileGlobalGitConfig|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestGlobalStatusRejectsPositionalArguments|TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord|TestStatusUnreadableApprovalStateSurfaced|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApprovalsUnreadableStateNamesReadFailure|TestHookApproveRecordsOperatorApproval|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestHookApproveReRecordsAfterChange|TestHookApproveIsIdempotentWhenRecordMatches|TestHookApproveResolvesSymlinkedOperand|TestHookApprovalsListsReadOnly)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 81.15
  },
  "full-cli-07": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestHookApprovalsToleratesMalformedRecord|TestHookApprovalsOnAbsentStateIsEmpty|TestHookRevokeRemovesRecord|TestHookRevokeAbsentLeavesStateByteIdentical|TestHookUsageErrors|TestHookDiagnosticsAgreeWithShell|TestStatusReportsShellHookTrustPosture|TestStatusJSONCarriesShellHookTrust|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestHookApproveEndToEndWithGeneratedHook|TestProfileImportRow|TestProfileImportLossyRow|TestProfileImportNameTakenRow|TestProfileComposeAddPathRow|TestProfileComposeAddRefusesAFormOnAPathSource|TestPathOverlayFromMachineConfigJoinsTheClosure|TestOverlayFromMachineConfigIsRefusedByKind|TestProfileUseTakeoverRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestResolveTakeoverRequiresRepair|TestSyncTakeoverFlagParses|TestInstallGitignoreEntryPoint|TestLegacyPackageStatusBindingAndRepair|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestRunUsesInjectedConfigSourceAndWriters|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestProductionExternalDepsBindTrustedGitAndAudit|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestRunVersionExitsZero|TestRunNoArgsPrintsUsage|TestRunUnknownCommand|TestShellInitPrintsHooks|TestShellInitInstallCachesHookWithoutConfig|TestSkillCheckOnTempDir)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 69.45
  },
  "full-cli-08": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestInstallFlagsAcceptTrailingOptions|TestInstallAuditFlagAcceptsOptionalMode|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestStatusDriftDetectsContentTampering|TestBootstrapAndProjectCommands|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestBootstrapIfMissingCreatesAbsentConfig|TestBootstrapIfMissingRejectsForce|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestCLIEndToEndInstallStatusAndTamperCheck|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestProductionBinaryDispatchesNetNSProbe|TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSAddListRemove|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestConfigSubcommandDispatch|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextCopyFallback|TestMuseRootContextVectors|TestMuseRootContextDrift|TestMuseRootContextRefusals)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 74.91
  },
  "full-cli-09": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestMuseCredentialPolicyStatus|TestMuseRootContextStatus|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestMuseProfileSeedSecretAndCredentialRefusals|TestMuseFragmentV3PublishedCases|TestMuseCLIAuthParentBoundary|TestMuseFragmentV4PublishedCases|TestMuseCLIStorePolicy|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverEveryRegisteredCommand|TestNativeSkillsTakeoverArgumentDiscardRegression|TestNativeSkillsTakeoverActivationBoundary|TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsCLIOperations|TestNativeSkillsTakeoverFlagBoundary|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestCLIAuditOpaqueNULFollowsHashVersion|TestAuditAllowWritesWriterVersionPinCarrier|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestStatusCheckAdmitsV2NULBearingInstall|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallActivationMatrix|TestProfileInstallStopThenRetryConformanceVector|TestProfileInstallReinstallUsePreservesProjectScope|TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 168.7
  },
  "full-cli-10": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestGitFileURLHasNoBackslash|TestProfileInstallListUse|TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileUseUnknownIsAFailure|TestProfileRemoveInUseIsAFailure|TestProfileInstallRefConflictIsAFailure|TestProfileSyncCoversTheMachineScope|TestProfileComposeListsDeclaredOverlays|TestProfileUseUnknownEnvironmentIsAFailure|TestProfileUseTargetIsStageB|TestProfileUseTargetBoundToAdapter|TestProfileInstallWarnsOnSystemModule|TestProfileScopedUseAndClear|TestProfileUpdatePinnedTagIsUnchanged|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallFileOperandIsRefused|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallUseActivatesThroughSwitch|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording|TestSecurityPostureVectorsThroughCLI|TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningIsNotEmittedByEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestSecurityPostureCuratorStatusRows|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestSecurityPostureProfileInstallRefusesMCPDeclarations|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestSecurityPostureRevisionBDefaultRefusesEmptySources|TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist|TestSecurityPostureRevisionBDefaultEnvStatus|TestCLIHelperProcess)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 19.85
  },
  "full-cli-11": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestCompiledProjectRepairsCorruptCompiledState|TestCompiledProjectRestoresCacheWhenCommitFails|TestStatusReportsATransitivelyResolvedCompiledCommand|TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands|TestGCRetainsAndReportsReferencedCompiledState|TestCompiledInstallFollowsTheNativeControlInventoryExactly|TestDryRunNeverClaimsACompletedCompilerCheck|TestStatusExplainsAnUnusableGoToolchain|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestCuratorStatusProviderPostureAndCheck|TestStatusAcceptsAnUnchangedLegacyMarkerSchema|TestHostGoToolchainIsSelectableOnAnInventoryPlatform|TestInstallPrintsTheRemedyAVersionManagerSelectionEarns|TestUmbrellaProviderResolutionVectors|TestActiveRevisionIsWarningRelease|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd|TestUmbrellaHostilePathPlantRefusedUnderB|TestUmbrellaTrustedRootSymlinkEscape|TestUmbrellaTrustRootSelectionBothRevisions|TestUmbrellaRefusedDirectoriesBothRevisions|TestUmbrellaSymlinkIntoManagedRefused|TestUmbrellaUnreadableRootNeverAbsence|TestUmbrellaMissingNamesRoots|TestUmbrellaWarningText|TestProviderPostureRows|TestProviderSubcommandParsing|TestExecutableCandidates|TestProviderInputsForHostScannedBinWarnsWithoutLedger|TestProviderInputsForHostPublishedBinRefuses|TestProviderPostureReportsUnreadablePublishedShimLedger|TestProviderInputsForHostExplicitBinRefuses|TestUmbrellaTrustRootCaseVariantIdentity)$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 140.73
  },
  "full-cli-last": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^TestCompiledProjectStatusAndUntrustedRecovery$",
      "github.com/relux-works/curator/cmd/curator"
    ],
    "exit": 0,
    "seconds": 104.44
  },
  "full-env-01": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 127.83
  },
  "full-env-02": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 37.94
  },
  "full-env-03": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 31.64
  },
  "full-env-04": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 158.84
  },
  "full-env-05": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestImportAbsentMarkerDetectsNative|TestImportCorruptMarkerIsLoss|TestImportUnreadableMarkerIsLoss|TestDotfileManagerVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestDotfileManagerTableMatchesSpecFixture|TestDotfileStatePathResolution|TestForeignManagerHintLstatDiscipline|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestManagedFallbackPermissions|TestManagedSeedPreservesNativePermissions|TestManagedHomeRepairTightensDirectories|TestManagedAuthoredFilesArePrivate|TestManagedSeedRefusalsPreserved|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveProvisionRepair|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestResolveStaleUnprovisioned|TestResolveUnknowns|TestResolveDefaultProfile|TestResolveDriftRepair|TestResolvePassthroughLiveness|TestClaudeRootAlwaysCopied|TestResolveClaudeProjectEntry|TestCodexKeyringAmbient|TestCodexUnreadableConfigFails|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormUnsupported|TestResolveLockContention|TestResolveFormats|TestProfileInstallMarkdownContextCLI)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 78.28
  },
  "full-env-06": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestProfileInstallMarkdownMCP|TestProfileInstallMarkdownScope|TestProfileInstallMarkdownCaseAliasesCLI|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureCredentialStoreAdmission|TestMaterializeCapturePendingDoesNotRecover|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureWaitReselectsAndTimeout|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureDoesNotProvision|TestMaterializeCaptureObjectBounds|TestMaterializeCaptureSerializesProfileWriters|TestMaterializeCaptureConcurrentRepair|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureAuthorizingDrift|TestMaterializeCaptureUnreadableJournal|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigratePiWrongTargetToAgentRoot|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateNoSecretCopies|TestMigrateConflictsRefuse|TestMigratePiRootInspectionFailure|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigrateOldRootBytesWarnAndProceed|TestMigratePendingNeedsNothing|TestMigrateUnknownOperands|TestMigrateInspectIsReadOnly|TestMigrateApplyRequiresPlan|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateMarkerDriftRefuses|TestMigrateFailedRelinkKeepsOldLink)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 90.85
  },
  "full-env-07": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestMigrateSyscallFailureRollsBack|TestMigrateInterruptedApplyRecovers|TestMigratePublishFailureRevertsLinks|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrateBrokenJournalRefuses|TestMigrateRecoveryRefusesUnexpectedTarget|TestMigrateRecoveryCleansOwnedTemp|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestMusePolicyExplicitMigration|TestMusePublishedLinkStates|TestMuseForkRefusedRegression|TestMuseAuthParentBoundaryAtResolve|TestMuseDirectoriesAndIsolation|TestMuseAuthRepairAdditionalStates|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMuseStatusOptionalProvisioning|TestMusePublishedRegistryLayout|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsNarrowingMutants|TestNativeSkillsUseAndSync|TestNativeSkillsTakeover|TestNativeSkillsLinkedParent|TestNativeSkillsLedger|TestNativeSkillsUnreadableMarker|TestNativeSkillsRepair|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsInstallAndUpdate|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestGitFileURLForKeepsVolumesOutOfTheHost|TestServeWritesTheGitSafeRemote)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 105.56
  },
  "full-env-08": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember|TestResolveRefusesV1CommitPinnedContextNUL|TestResolveCleanV1CommitContextStaysCurrent|TestResolveAdmitsV2CommitPinnedContextNUL|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL|TestStrictAuditMemberRefusesV1NULBeforeHashing|TestValidateNamedStorePinsRefusesV1NULBeforeHashing|TestSkillsOfRefusesV1NULBeforeHashing|TestGuardedReadersV1NULRefusalComputesNoV1Identity|TestInstallAdmitsV1CollidingSkillTreesUnderV2|TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2|TestPrecedencePrimitivesDriveEmission|TestPrecedenceTieKeepsTopologicalOrder|TestPathOverlayJoinsClosure|TestOverlayExplicitWeightOverridesDefault|TestGitOverlayJoinsClosure|TestPathOverlayFormIsSourceInvalid|TestOverlayInvalidSourceKindIsSourceInvalid|TestSCPOverlayResolvesAsGit|TestOverlayRepeatedNameIsCompositionInvalid|TestOverlayDuplicateNameIsCompositionInvalid|TestForbiddenOverlaysResolveAlone|TestRemoveRefusesOverlayOfAnotherProfile|TestWeightRulesApplyInOrder|TestWeightRulesAllFourDisagree|TestWeightDuplicateRefused|TestWeightUnknownRefused|TestNonRootWeightsMapRefused|TestWeightConflictErrorAndWarning|TestPathInstallCapturesDirtyUntrackedInsideGit|TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 49.12
  },
  "full-env-09": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestPathRootAndImportedMCPDeclarationsRefused|TestPathOverlayMCPDeclarationRefused|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestResolveRejectsGroupWritableOnlyPathOverlay|TestMissingPathOperandIsPathMissing|TestMissingOverlayPathIsPathMissing|TestNonDirectoryPathIsSourceInvalid|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsUntrusted|TestFoldedPathsAreSourceInvalid|TestRootGitExcluded|TestUnreadablePathIsUnreadable|TestPathSnapshotImmutableAcrossUpdateSyncUse|TestPathUpdateAllSucceedsWithImportedProfile|TestUnreadablePathRootLeadsUnreadable|TestUpdatePathWithoutStatePinIsSourceInvalid|TestUpdatePathMissingSnapshotIsSourceInvalid|TestUpdatePathSnapshotNameMismatchIsSourceInvalid|TestPathReinstallBareAndNonCurrentMovesPin|TestUpdateBlocksSecretOverlayMember|TestReinstallBlocksSecretOverlayMember|TestUpdateBlocksRevokedOverlayMember|TestUpdateBlocksV1RevokedOverlayMemberUnderV1Writers|TestReinstallBlocksRevokedOverlayMember|TestPermissionsForProfileProperties|TestRepairPreserveLockVectors|TestPreserveLockRefusesBeforeStoreRebuild|TestPreserveLockRaceRefusesBeforeJournalRecovery|TestPreserveLockMigrationRefusalPrecedesBusyLock|TestPreserveLockRevalidationHoldsRepairLock|TestPlainRepairMigrationKeepsLockPrivate|TestSystemDeltaVectors)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 23.82
  },
  "full-env-10": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestRC14MigrationRehashesLegacyIdentities|TestRC14ResolveRepairMigratesSiblingHomes|TestRC14IdentityMigrationRollsBackEveryEntry|TestRC14IdentityMigrationRefusesSiblingPlanDrift|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestRC14IdentityMigrationRecoversInterruptedCommit|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14IdentityMigrationPreservesFallbackCopies|TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestEnvironmentsReadFailureVectors|TestPassthroughReadFailuresAreUnknownAndNeverRepaired|TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestUnreadableLockRefusesUpdateWithoutRewriting|TestAbsentLockRemainsProfileUnknown|TestUnreadableCodexSeedDoesNotSelectAbsentDefault|TestReviewerRecoveryMarkerDrift|TestReviewerRecoveryPreservesRegularTemp|TestSkillDirectoryNarrowingMutants|TestSkillDirectoryLifecycleVectors|TestSkillDirectoryMarkdownDeclaredControl|TestSkillDirectoryMetadata|TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse|TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer|TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 107.63
  },
  "full-env-11": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestUpdateConfirmationPostureVectorAtStatus|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusShadowAcknowledgment|TestStatusProfileMembersAndUnregistered|TestSeedUnreadableStopsBeforeFirstWrite|TestReservedSkillName|TestProfilePathCollision|TestOpencodeFormFallback|TestRepairFailedMissingStore|TestOpencodeXDGSeeds|TestClaudeSeedMergePreservesToolState|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestStoreBoundaryStatusVectorsDriveStatus|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestInstallEmitsSurfacingBeforePublication|TestUpdateEmitsSurfacingBeforePublication|TestReinstallEmitsSurfacingBeforePublication|TestSurfacingEmittedDespitePublicationFailure|TestSurfacingSinkWriteFailureIsNonFatal|TestInstallSurfacesMCPDeclarations|TestInstallSurfacingSilentWithoutMCP|TestUpdateSurfacesCandidateMCPSet|TestAllowlistWarningOperations|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 38.55
  },
  "full-env-12": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestResolvePassthroughKnobs|TestStatusS4Posture|TestSurfacingUnreadableManifest|TestUseWithoutTakeoverRefusesUnmanaged|TestUseTakeoverBacksUpAndNotifies|TestForeignSymlinkStopsSwitch|TestTakeoverWarnsDotfileHeuristic|TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn|TestReadOnlyCommandsNeverOnboard|TestRepairTakeoverProvisionsManagedHome|TestUnmanageRestorePlanRetainsMetadata|TestUnmanageRestoreRefusesLinkedParent|TestUnmanageRestoreSymlinkReplacesDestinationEntry|TestEnvironmentWriteNofollowVectors|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 10.62
  },
  "full-env-last": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^TestProfileNetworkContactVectors$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 61.36
  },
  "full-install-01": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 44.26
  },
  "full-install-02": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 106.48
  },
  "full-install-03": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 136.4
  },
  "full-install-04": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestDraftInstallGitTamperRefused|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftTransportEnabled|TestDraftTransportPlanConversion|TestDraftTransportAuth|TestDraftPlanNeedsSSH|TestDraftPlanNamesProvider|TestDraftSSHWrapperBase|TestDraftManagerExecutable|TestAcquireDraftNetworkSelection|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestDefaultAcquireFetchesTheDeclaredURL|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAcquireDraftNetworkRevision2Refusals|TestRevision1ReaderRejectsV2Shapes|TestUserConfigIgnoredByResolvedLane|TestDraftTransportPlanV2Conversion|TestDraftPlanNeedsSSHRevision2|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestDeclarationGenerationIsContentAddressed|TestDocumentGenerationIsStableWhenUnreadable|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGitFixtureNeverRetriesNonZeroExit|TestGitFixtureNeverRetriesWrappedExitError|TestGitFixtureRetryPredicate|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestGitFixtureOuterEntryRunsRealGit|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution|TestGlobalInstallRecognizesAdoptedForwardingShim|TestGlobalReconcilesOwnedLegacyForwardingLink|TestEndToEndInstall|TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestSecondInstallIsUpToDate|TestTamperTriggersReinstall|TestRemovedSkillCleanedUp|TestDryRunTouchesNothing)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 88.38
  },
  "full-install-05": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestGitignoreGateSkips|TestMissingSystemCommandFails|TestMovedTagWarningAndStrict|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestRuntimeOnlyProviderStillRequiresSkillMd|TestHybridSkillActivatesWithoutTouchingProjectStore|TestHybridShadowedByProjectDeclaration|TestGlobalInstall|TestGlobalInstallRejectsSchema2Skillfile|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestGlobalStrictTagsDetectMovedTag|TestMcpRequirementGatesInstall|TestAuditGateBlocksUndeclaredNetwork|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestLegacyLaneRecordsDependencyDirectory|TestLegacyLockPackageArms|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestLegacyLaneRefusesGlobDependencyDirectory|TestLegacyLaneRejectsUnboundPreviousGeneration|TestLegacyLaneDependencyTagMovement|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestReviewChangedDeclaredTagIsNotMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestLegacyLaneChangedTagAfterOldTagDeleted|TestLegacyLaneRuntimeMatchesPackageMarker|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPostCommitMaintenanceWarningsReachTheResult|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestASchemaSixCommandDeclaresNoModules|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV1ReinstallRefusesNULAppearingAfterInstall|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestV2InstallDistinguishesNULCollisionTwins|TestDraftLaneStillBlocksNULBearingMember|TestInstallDraftOwnershipCases|TestInstallPreservesUnownedStaleStoreEntries|TestInstallRefusesUnownedSelectedCanonicalDestination|TestInstallRefusesUnprovenStaleOwnership|TestInstallOwnedCanonicalPositiveControls|TestGlobalCanonicalOwnershipPreservation|TestHybridCanonicalOwnershipPreservation|TestInstallDraftMarkerWriterCases|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestV1InstallRejectsV2RegistryEvidence)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 189.97
  },
  "full-install-06": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestV2RegistryAttestationUsesVersionedInstallPath|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistryBootstrapVectorsDriveInstall|TestStrictRegistryPolicyFailsUnknown|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStableDeclarationInputsCommitWithoutRestarting|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStageProjectTargetsRefusesMissingSnapshot|TestStageGlobalTargetsPopulatesBoundaries|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestRunCommitWithRealStagingJournalsOneTransaction|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged|TestDryRunReportsCacheHitWithoutBuilding|TestCacheHitPerformsNoSourceAwareGoCommand|TestStagingRunsProviderFirstAndCommandLexical|TestStagedOutputsStayPrivateAndAreReleased|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestCacheHitOnlyPlanStillFinalizesToolchainTrust|TestCacheInspectionIsBracketedByTheFrozenSource|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestCorruptCacheEntryIsRebuiltAndNeverReused|TestUnsupportedCacheProtectionFailsClosed|TestIncompletePlansNeverClaimACompleteDiagnostic|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 110.17
  },
  "full-install-07": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestScriptOnlyInstallPerformsNoToolchainOrCacheWork|TestInjectedClockAndGenerationReaderDriveInstallMarkers|TestPlannedBuildAccessorsAreImmutable|TestReleaseTakesNoToolchainVerdictAfterLiveMutation|TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation|TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall|TestDefaultToolchainProbeRemovesItsProbeRootOnFailure|TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess|TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure|TestProjectWriteBlobsSpawnFailureIsWrapped)$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 7.54
  },
  "full-other-01": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/cmd/curator-spec-pin",
      "github.com/relux-works/curator/internal/adapters",
      "github.com/relux-works/curator/internal/artifactpolicy",
      "github.com/relux-works/curator/internal/artifactpolicy/conformance",
      "github.com/relux-works/curator/internal/audit",
      "github.com/relux-works/curator/internal/buildcache",
      "github.com/relux-works/curator/internal/buildmeta",
      "github.com/relux-works/curator/internal/buildrepo",
      "github.com/relux-works/curator/internal/buildsource",
      "github.com/relux-works/curator/internal/capabilities",
      "github.com/relux-works/curator/internal/closure",
      "github.com/relux-works/curator/internal/closureexec",
      "github.com/relux-works/curator/internal/closuregraph",
      "github.com/relux-works/curator/internal/config",
      "github.com/relux-works/curator/internal/conformancecoverage"
    ],
    "exit": 0,
    "seconds": 136.62
  },
  "full-other-02": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/internal/contextaudit",
      "github.com/relux-works/curator/internal/contextlock",
      "github.com/relux-works/curator/internal/contextmaterialize",
      "github.com/relux-works/curator/internal/contextpkg",
      "github.com/relux-works/curator/internal/contextresolve",
      "github.com/relux-works/curator/internal/contextstore",
      "github.com/relux-works/curator/internal/crossconformance",
      "github.com/relux-works/curator/internal/devsub",
      "github.com/relux-works/curator/internal/envfiles",
      "github.com/relux-works/curator/internal/envfragment",
      "github.com/relux-works/curator/internal/envmarker",
      "github.com/relux-works/curator/internal/envregistry",
      "github.com/relux-works/curator/internal/envsnapshot",
      "github.com/relux-works/curator/internal/gitcred",
      "github.com/relux-works/curator/internal/gitignore"
    ],
    "exit": 0,
    "seconds": 148.29
  },
  "full-other-03": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/internal/gitops",
      "github.com/relux-works/curator/internal/globalbins",
      "github.com/relux-works/curator/internal/godriver",
      "github.com/relux-works/curator/internal/hashing",
      "github.com/relux-works/curator/internal/hookapproval",
      "github.com/relux-works/curator/internal/identifiers",
      "github.com/relux-works/curator/internal/identity",
      "github.com/relux-works/curator/internal/install/atomicity",
      "github.com/relux-works/curator/internal/interop",
      "github.com/relux-works/curator/internal/interop/environments",
      "github.com/relux-works/curator/internal/locale",
      "github.com/relux-works/curator/internal/managerlock",
      "github.com/relux-works/curator/internal/manifest",
      "github.com/relux-works/curator/internal/marker",
      "github.com/relux-works/curator/internal/mcp"
    ],
    "exit": 0,
    "seconds": 75.93
  },
  "full-other-04": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/internal/moduleroots",
      "github.com/relux-works/curator/internal/nodesource",
      "github.com/relux-works/curator/internal/npmsource",
      "github.com/relux-works/curator/internal/opaquescan",
      "github.com/relux-works/curator/internal/pathboundary",
      "github.com/relux-works/curator/internal/pkgversion",
      "github.com/relux-works/curator/internal/pnpmsource",
      "github.com/relux-works/curator/internal/privatedir",
      "github.com/relux-works/curator/internal/protocoljson",
      "github.com/relux-works/curator/internal/registry",
      "github.com/relux-works/curator/internal/runtimestore",
      "github.com/relux-works/curator/internal/rustsource",
      "github.com/relux-works/curator/internal/scopes",
      "github.com/relux-works/curator/internal/scriptpolicy",
      "github.com/relux-works/curator/internal/scriptworker"
    ],
    "exit": 0,
    "seconds": 113.83
  },
  "full-other-05": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/internal/shell",
      "github.com/relux-works/curator/internal/skillcheck",
      "github.com/relux-works/curator/internal/skillspec",
      "github.com/relux-works/curator/internal/snapshot",
      "github.com/relux-works/curator/internal/sourcelock",
      "github.com/relux-works/curator/internal/staging",
      "github.com/relux-works/curator/internal/stateread",
      "github.com/relux-works/curator/internal/swiftpmbuild",
      "github.com/relux-works/curator/internal/swiftpminterop",
      "github.com/relux-works/curator/internal/swiftpmsource",
      "github.com/relux-works/curator/internal/testcli",
      "github.com/relux-works/curator/internal/testgitenv",
      "github.com/relux-works/curator/internal/testtoolchain",
      "github.com/relux-works/curator/internal/transaction",
      "github.com/relux-works/curator/internal/ui"
    ],
    "exit": 0,
    "seconds": 45.82
  },
  "full-other-06": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "github.com/relux-works/curator/internal/verr",
      "github.com/relux-works/curator/internal/version",
      "github.com/relux-works/curator/internal/whitelist",
      "github.com/relux-works/curator/internal/yarnclassicsource",
      "github.com/relux-works/curator/internal/yarnmodernsource",
      "github.com/relux-works/curator/tools/ciworkflow",
      "github.com/relux-works/curator/tools/goreleaserconfig"
    ],
    "exit": 0,
    "seconds": 42.02
  },
  "full-install-last": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^TestAuthoritativeDryRunCasesMutateNothingPersistent$",
      "github.com/relux-works/curator/internal/install"
    ],
    "exit": 0,
    "seconds": 0.92
  },
  "conformance-env-01": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestClaudeSeedMergePreservesToolState|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCodexSeedStripsInlineMCPTable|TestDotfileManagerVectors|TestEnvironmentWriteNofollowVectors|TestEnvironmentsCodexSeedVectors|TestEnvironmentsEnvPassthroughVectors|TestEnvironmentsReadFailureVectors|TestMuseAuthParentBoundaryAtResolve|TestMuseAuthRepairAdditionalStates|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMusePublishedLinkStates|TestMusePublishedRegistryLayout|TestMuseStatusOptionalProvisioning|TestOpencodeFormFallback|TestOpencodeXDGSeeds)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 52.32
  },
  "conformance-env-02": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathKindMCPVectorsDriveInstallEntry|TestProfilePathCollision|TestRepairFailedMissingStore|TestReservedSkillName|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestSeedUnreadableStopsBeforeFirstWrite|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusProfileMembersAndUnregistered|TestStatusShadowAcknowledgment|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 22.85
  },
  "conformance-env-03": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "-run",
      "^(TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryStatusVectorsDriveStatus|TestSystemDeltaVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateConfirmationPostureVectorAtStatus)$",
      "github.com/relux-works/curator/internal/envprofile"
    ],
    "exit": 0,
    "seconds": 11.25
  },
  "build": {
    "args": [
      "build",
      "./..."
    ],
    "exit": 0,
    "seconds": 6.52
  },
  "vet": {
    "args": [
      "vet",
      "./..."
    ],
    "exit": 0,
    "seconds": 6.34
  },
  "conformance-core": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-json",
      "./internal/audit",
      "./internal/skillcheck",
      "./internal/whitelist",
      "./internal/contextpkg",
      "./internal/contextlock",
      "./internal/contextmaterialize"
    ],
    "exit": 0,
    "seconds": 44.57
  }
}
```

Development command ledger (owning checkout before the final commit):

```json
{
  "focused-audit-rerun": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-v",
      "./internal/audit",
      "./internal/whitelist",
      "./internal/contextpkg",
      "./internal/contextmaterialize",
      "./internal/skillcheck"
    ],
    "exit": 0,
    "seconds": 42.27
  },
  "focused-audit": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-v",
      "./internal/audit",
      "./internal/whitelist",
      "./internal/contextpkg",
      "./internal/contextmaterialize",
      "./internal/skillcheck"
    ],
    "exit": 1,
    "seconds": 41.36
  },
  "focused-kinds": {
    "args": [
      "test",
      "-count=1",
      "-timeout=5m",
      "-v",
      "-run",
      "TestMarkdownScope(Context|OtherKinds|CacheBindsKind|InvalidatesRulesetSeven)",
      "./internal/audit"
    ],
    "exit": 0,
    "seconds": 7.89
  },
  "focused-last": {
    "args": [
      "test",
      "-count=1",
      "-timeout=3m",
      "-v",
      "-run",
      "TestMarkdownScope(CacheBindsKind|InvalidatesRulesetSevenAllowance)|TestStrictAuditMemberCapabilityCacheIsolation",
      "./internal/audit",
      "./internal/envprofile"
    ],
    "exit": 0,
    "seconds": 2.08
  },
  "focused-mcp": {
    "args": [
      "test",
      "-count=1",
      "-timeout=2m",
      "-v",
      "-run",
      "^TestProfileInstallMarkdownMCP$",
      "./internal/envprofile"
    ],
    "exit": 0,
    "seconds": 3.65
  },
  "focused-profile": {
    "args": [
      "test",
      "-count=1",
      "-timeout=8m",
      "-v",
      "-run",
      "Test(ProfileInstallMarkdown|SkillDirectory|ReviewPrior|Manager.*Read)",
      "./internal/envprofile"
    ],
    "exit": 0,
    "seconds": 80.41
  }
}
```

Operational commands completed with exit 0 before attachment: `a2a inbox`; board assignment/status query; `task-board resource get` for all six preconditions and the r2 review outcome; `a2a post` ack; `git fetch origin`; `git diff --check`; `git commit -S`; `git log -1 --format=%G?` (G); disposable `git clone --no-hardlinks --no-checkout` and detached checkout of the exact head; pinned separate-git-dir spec clone and detached checkout; `git push -u origin fix/audit-markdown-scope`; fresh `git ls-remote --heads origin fix/audit-markdown-scope` equality check. Git network/signing commands use the configured agent Git environment. No main push, force push or PR operation was performed.

## Development controls and limits

Initial focused audit execution exited 1 because independent kind equality masked two older capability mutants, and the newly added CLI mutant did not propagate its overlay into the child CLI build. The fixtures now bind the actual kind and the CLI mutation carries GOFLAGS=-overlay to the child build. The ruleset-7 reuse mutant correctly failed old-record preservation rather than stale allowance reuse; its expected assertion was corrected to that precise failure. The focused audit rerun exited 0 and all controls passed again at the exact head in the conformance/default runs. The product F2 regression and its clean controls passed independently throughout. Build, vet and the full final suite have no unresolved failures.

Validation here is macOS arm64. Windows/Linux-specific cases remain conditional; no hosted CI status, external review, PR or landing is claimed. Publication ends at the signed feature branch, as required by the worker contract.

Open decisions: none. The context capability interpretation was resolved by the orchestrator in the task topic. Changelog line: “Fix profile audit coverage for declared context modules while preserving skill prompt exclusions; invalidate stale context allowances with ruleset 8.”

Task-local validation clone, spec clone, private caches, fixtures, helpers and copied binaries are deleted after handover. This results file retains the command ledger and evidence summary; no temporary log paths are promised to remain available.
