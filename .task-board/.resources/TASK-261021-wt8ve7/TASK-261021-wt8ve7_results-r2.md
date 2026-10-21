# TASK-261021-wt8ve7 results r2

## Delivery

Branch: `research/project-source-exclusion-claude-r2`. Head: `0c5c02c8be22878e37f1565318024aa94d261395`.
Base: `2faaa5446c55657d86903b46b647c12d502c6d24`, the current `origin/main` observed by the task fetch.
Accepted old range: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7..9f66334b48b4765f3011d11f0e1170d678cc4b40`.

Both introduced commits are signed and verify as `G`; the pushed feature branch equals the validated head. Source and normal-.git validation checkouts are clean. No PR, main push, force push or board-status mutation.

## Changes and checks

- Signed cherry-pick of the accepted research commit. The sole conflict was README.md. It now retains the complete Codex entry followed by the complete accepted Claude entry. Removing the exact Claude addition reproduces the base README byte for byte.
- All 12 non-README paths in the cherry-pick are byte-identical to the accepted commit.
- Separate signed commit corrects one qualification-record row: the measured Task child is async/background. Foreground children are explicitly unmeasured. The retained native stats show two measured branches with `started_in_background=1`, `spawned=1`, `completed=1`, `failed=0`, and requested mode unset.
- No product code, harness logic, fixtures, frozen source manifest, observations, or prior runtime evidence changed during this revision. No new behavior gate was added; no new mutation test is applicable to this prose correction.
- The previous opt-in runtime replay and its narrowing controls are retained accepted evidence, not newly run r2 checks. This task requested build, vet and the default suite; all were rerun at the new head.

## Files changed relative to base

- `.research/project-exclusion-claude/README.md`
- `.research/project-exclusion-claude/collect.py`
- `.research/project-exclusion-claude/evidence/mutants.log`
- `.research/project-exclusion-claude/evidence/replay.log`
- `.research/project-exclusion-claude/evidence/source-sha256.txt`
- `.research/project-exclusion-claude/evidence/ungated.log`
- `.research/project-exclusion-claude/mutants.sh`
- `.research/project-exclusion-claude/observations.json`
- `.research/project-exclusion-claude/replay_test.go`
- `.research/project-exclusion-claude/testdata/mcp.py`
- `README.md`
- `UNRESOLVED_QUESTIONS.md`
- `docs/qualification/project-source-exclusion-claude.md`

## Validation environment and coverage

Go 1.25.5 via the configured CURATOR_GO. Target: supported local macOS arm64. Every Go test explicitly removes GIT_DIR and GIT_WORK_TREE. Every check runs in a disposable normal-.git validation clone at the exact delivered head, with TMPDIR=/tmp/cw-dev2, workspace-local Go caches and GOTELEMETRY=off.

PATH preserves inherited tools (including node/npx), substituting task-local symlink directories only for directories containing ambient curator-* extensions; those extension executables are omitted. This avoids unrelated installed provider extensions affecting CLI fixture expectations. The rustsource group additionally uses the existing concrete Rust 1.91.0 compiler bin directory. No installation or external inference was requested. Existing tests retain their fixture isolation; no credentials or native user configuration were inspected by the worker.

All 85 default packages are accounted for. All 1387 inventoried top-level Test/Example/Fuzz entries across the five sharded large packages were observed. There are 11606 passing and 255 skipped test/subtest events across the retained checks (including the explicit ungated hidden research checks).

The package list is derived from `go list ./...`. Package groups include every ordinary package. Shard selectors cover every current inventoried test entry, including any entries added since the original accepted head. No failed or omitted shard is counted as passing. The artifactpolicy cleanup test is additionally replayed in isolation because its large disposable-directory cleanup can consume a package time budget under host load.

## Test commands and the claims they prove

Each test command below exercises the existing executable assertions in its named package(s) and selected tests at the delivered head. Their existing per-test claims remain in the source. A passing group establishes those assertions for these inputs and this platform, not a universal property. Negative controls and neighboring positive controls in those suites are preserved. The new record correction is checked independently against retained native stats, and absence of a foreground qualification claim is checked by reviewing its final row.

| Job | Exit | Elapsed | Claim |
|---|---:|---:|---|
| build | 0 | 1.009s | All ordinary packages compile. |
| vet | 0 | 1.973s | Default packages and Claude research harness pass static vet checks. |
| envprofile-0 | 0 | 68.556s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-1 | 0 | 53.767s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-2 | 0 | 14.146s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-3 | 0 | 42.517s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-4 | 0 | 31.135s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-5 | 0 | 25.989s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-6 | 0 | 71.628s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-7 | 0 | 58.769s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-0 | 0 | 31.447s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-1 | 0 | 77.811s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-2 | 0 | 42.219s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-3 | 0 | 20.778s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-4 | 0 | 63.582s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-5 | 0 | 9.284s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-6 | 0 | 29.709s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-7 | 0 | 25.978s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| atomicity | 0 | 50.054s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-0 | 0 | 95.112s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-1a | 0 | 1.103s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-1b | 0 | 4.861s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-1c | 0 | 5.72s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-1d | 0 | 1.31s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-2a | 0 | 2.21s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-2b | 0 | 11.078s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-2c | 0 | 11.175s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-2d | 0 | 97.258s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-3a | 0 | 2.089s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-3b | 0 | 22.744s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-3c | 0 | 21.258s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-3d | 0 | 42.87s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| rest-4 | 0 | 0.585s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-0 | 0 | 5.858s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-1 | 0 | 18.009s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-2 | 0 | 19.012s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-3 | 0 | 31.638s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-4 | 0 | 31.141s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-5 | 0 | 5.267s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-6 | 0 | 34.857s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-7 | 0 | 30.37s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-8 | 0 | 26.821s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-9 | 0 | 17.864s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-10 | 0 | 20.999s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-11 | 0 | 10.72s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-12 | 0 | 90.728s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-13 | 0 | 122.671s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-14 | 0 | 63.038s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-15 | 0 | 27.978s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-16 | 0 | 8.529s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-17 | 0 | 13.455s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-18 | 0 | 23.334s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-19 | 0 | 56.437s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-20 | 0 | 48.515s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-21 | 0 | 14.565s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-22 | 0 | 4.874s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-23 | 0 | 41.003s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-24 | 0 | 85.045s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-25 | 0 | 6.939s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-26 | 0 | 17.582s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-27 | 0 | 12.297s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-28 | 0 | 13.416s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-29 | 0 | 39.878s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-30 | 0 | 24.051s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cli-31 | 0 | 35.547s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-0 | 0 | 18.918s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-1 | 0 | 38.54s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-2 | 0 | 11.259s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-3 | 0 | 2.514s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-4 | 0 | 1.866s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-5 | 0 | 41.366s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-6 | 0 | 43.865s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| cross-7 | 0 | 4.813s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-2b | 0 | 29.317s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-3b | 0 | 43.362s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-4b | 0 | 39.227s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-5b | 0 | 26.433s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-6b | 0 | 9.733s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| envprofile-7b | 0 | 18.276s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-2b | 0 | 36.293s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-3b | 0 | 34.834s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-4b | 0 | 40.426s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-5b | 0 | 24.175s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-6b | 0 | 72.266s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-7b | 0 | 15.39s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-0 | 0 | 11.306s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-1 | 0 | 7.015s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-2 | 0 | 4.123s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-3 | 0 | 4.613s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-4 | 0 | 3.844s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-5 | 0 | 2.522s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-6 | 0 | 3.617s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| godriver-7 | 0 | 16.841s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| install-1b | 0 | 42.363s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| artifactpolicy-cleanup-rerun | 0 | 49.638s | Existing selected regression assertions pass; package and top-level coverage checked against inventory. |
| research-ungated | 0 | 1.177s | Claude and Codex vendor leaves skip without opt-in; no vendor replay is claimed. |

All Go commands below have the shared environment described above. `$CURATOR_GO` denotes the configured absolute tool executable; no personal locator is published.

### build

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" build ./...
```
Exit: 0. Elapsed: 1.009s.

### vet

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" vet ./... ./.research/project-exclusion-claude
```
Exit: 0. Elapsed: 1.973s.

### envprofile-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestDanglingPiLinkReportedDetached|TestReviewerDanglingExpectedTarget|TestDanglingExpectedTargetRepairSucceeds|TestPiProvisionTargetsAgentRootV2WriterMode|TestFileOperandIsRefused|TestRevokedSourceIsRefused|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestDefaultProfileMaterializesOnFreshHome|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportAbsentMarkerDetectsNative|TestManagedSeedRefusalsPreserved|TestResolveProvisionRepair|TestResolveClaudeProjectEntry|TestResolveLockContention|TestMaterializeCaptureConcurrentRepair|TestMigratePlanDriftRefuses|TestMigrateSyscallFailureRollsBack|TestMigrateRecoveryRefusesUnexpectedTarget|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsLedger|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestWeightRulesAllFourDisagree|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestUpdatePathMissingSnapshotIsSourceInvalid|TestPathReinstallBareAndNonCurrentMovesPin|TestPreserveLockRaceRefusesBeforeJournalRecovery|TestSystemDeltaVectors|TestRC14ResolveRepairMigratesSiblingHomes|TestRC14IdentityMigrationRecoversInterruptedCommit|TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestReviewerRecoveryMarkerDrift|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestForeignSymlinkStopsSwitch|TestTakeoverWarnsDotfileHeuristic|TestUnmanageRestoreSymlinkReplacesDestinationEntry|TestEnvironmentWriteNofollowVectors)$'
```
Exit: 0. Elapsed: 68.556s.

### envprofile-1

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestLegacyPathInstallIgnoresStoreOverlap|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestCodexCredentialStoreTOMLSpellings|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestInstallUseSwitchesAndAgrees|TestInstallOperandKindIsSyntactic|TestCanonicalIdentityUnifiesSpellings|TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestMigratedGlobalSkillBinaryWriteVersion|TestImportLossyProceedsWithConsent|TestImportNameTakenBeforeAnyWrite|TestImportRecordsImportedFromNative|TestImportLedgerFailureIsLoss|TestDotfileStatePathResolution|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveFormUnsupported|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigrateUnknownOperands|TestMigrateMarkerDriftRefuses|TestMigrateInterruptedApplyRecovers|TestMuseAuthParentBoundaryAtResolve|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestNativeSkillsTakeover|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestPrecedenceTieKeepsTopologicalOrder|TestSCPOverlayResolvesAsGit|TestWeightRulesApplyInOrder|TestNonDirectoryPathIsSourceInvalid|TestPathUpdateAllSucceedsWithImportedProfile|TestUpdateBlocksV1RevokedOverlayMemberUnderV1Writers|TestRepairPreserveLockVectors|TestRC14MigrationRehashesLegacyIdentities|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestManagerOwnedAbsenceReadsAreGuarded|TestProfileSourceAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestUpdateConfirmationPostureVectorAtStatus|TestStatusOrphanRetention|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestInstallSurfacesMCPDeclarations|TestStatusS4Posture|TestUseWithoutTakeoverRefusesUnmanaged)$'
```
Exit: 0. Elapsed: 53.767s.

### envprofile-2

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestDropRepairWarnsAndFragmentFollowsAdmitted|TestStrictAuditMemberCapabilityCacheIsolation|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestImportCorruptMarkerIsLoss|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveStaleUnprovisioned|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMigratePublishFailureRevertsLinks|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestResolveCleanV1CommitContextStaysCurrent|TestGuardedReadersV1NULRefusalComputesNoV1Identity|TestOverlayRepeatedNameIsCompositionInvalid|TestRemoveRefusesOverlayOfAnotherProfile|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14IdentityMigrationPreservesFallbackCopies|TestEnvironmentsReadFailureVectors|TestUnreadableLockRefusesUpdateWithoutRewriting|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestClaudeSeedMergePreservesToolState|TestRepairTakeoverProvisionsManagedHome|TestUnmanageRestoreRefusesLinkedParent)$'
```
Exit: 0. Elapsed: 14.146s.

### envprofile-3

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestStatusReportsPolicyAndDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPathOverlayAdmitsOrdinaryDirectory|TestMigrationHonoursSystemLockedRevocation|TestUpdateDefaultIsBlocked|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestMigrateConflictsRefuse|TestMigrateInspectIsReadOnly|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestGitFileURLForKeepsVolumesOutOfTheHost|TestSkillsOfRefusesV1NULBeforeHashing|TestPathOverlayJoinsClosure|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPermissionsForProfileProperties|TestReviewerRecoveryPreservesRegularTemp|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusShadowAcknowledgment|TestUpdateEmitsSurfacingBeforePublication|TestSurfacingSinkWriteFailureIsNonFatal|TestSkillDirectoryNarrowingMutants)$'
```
Exit: 0. Elapsed: 42.517s.

### envprofile-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestDropAllDroppedOmitsFragmentSection|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexUnknownStoreSharedRefuses|TestReviewerLiteralCodexSelector|TestMigrateApplyLockContention|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestMigrationSucceedsWhenSystemPolicyPermits|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestInstallPathAndList|TestInstallGitResolvesNetworkIdentity|TestDotfileManagerVectors|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestResolveUnknowns|TestResolveDefaultProfile|TestCodexKeyringAmbient|TestMaterializeCaptureUnreadableJournal|TestServeWritesTheGitSafeRemote|TestWeightUnknownRefused|TestFoldedPathsAreSourceInvalid|TestUnreadablePathRootLeadsUnreadable|TestUpdatePathSnapshotNameMismatchIsSourceInvalid|TestPreserveLockRevalidationHoldsRepairLock|TestStatusProfileMembersAndUnregistered|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestResolvePassthroughKnobs|TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn|TestReadOnlyCommandsNeverOnboard|TestUnmanageRestorePlanRetainsMetadata)$'
```
Exit: 0. Elapsed: 31.135s.

### envprofile-5

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestPolicyFromConfigCarriesAdmission|TestClaudeRepairPreservesNumberText|TestMachinePolicyLoadFailureFailsMigration|TestInstallUseSkipsScopedAdapter|TestScopedRecordFilenameIsPortable|TestUsePartialRecordsNothing|TestImportPartialActivationKeepsLock|TestImportSkipsManagedRootContext|TestDotfileManagerTableMatchesSpecFixture|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormats|TestMigrateNoSecretCopies|TestMigrateBrokenJournalRefuses|TestMusePublishedRegistryLayout|TestNativeSkillsLinkedParent|TestNativeSkillsInstallAndUpdate|TestUpdateBlocksDeepNULFileInNewContextMember|TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2|TestPathOverlayFormIsSourceInvalid|TestRootGitExcluded|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestSurfacingEmittedDespitePublicationFailure|TestAllowlistWarningOperations|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$'
```
Exit: 0. Elapsed: 25.989s.

### envprofile-6

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestClaudeSeedNarrowingMutants|TestClaudeRepairRefusesUnsafeState|TestPiProvisionTargetsAgentRoot|TestEnvironmentsEnvPassthroughVectors|TestGateSourceRejectsFileRemote|TestFileRequirementSourceIsRefused|TestOperandShadowedByDirectoryResolvesAsGit|TestInstallNeverDemotesANetworkIdentityToAPath|TestSyncWritesScopedHomeOnce|TestUseMaterializesLinkedHomes|TestRemovePurgeCleansHomes|TestEnsureDefaultMigratesGlobalSkills|TestImportNormalizesOrdersAndPins|TestManagedSeedPreservesNativePermissions|TestManagedAuthoredFilesArePrivate|TestMaterializeCaptureCredentialStoreAdmission|TestMigrateRollbackRestoresPriorState|TestProfileNetworkContactVectors|TestResolveRefusesV1CommitPinnedContextNUL|TestForbiddenOverlaysResolveAlone|TestUnreadablePathIsUnreadable|TestUpdatePathWithoutStatePinIsSourceInvalid|TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestSignerPostureVectorsAtStatus|TestOpencodeFormFallback|TestStoreBoundaryResolveVectorsDriveResolve)$'
```
Exit: 0. Elapsed: 71.628s.

### envprofile-7

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestPathInstallAdmitsOrdinaryDirectory|TestCodexSeedStripsInlineMCPTable|TestSourceAllowlistRejectsAdjacentPathSegment|TestInstallUsePartialLeavesCurrent|TestInstallRefusesANonSourceOperand|TestReinstallSameSourceIsAnUpdate|TestUseReplacesWithBackup|TestRemoveRefusesCurrent|TestMigratedGlobalSkillNetworkCapabilities|TestClaudeRootAlwaysCopied|TestCodexUnreadableConfigFails|TestMaterializeCaptureAuthorizingDrift|TestMigratePiWrongTargetToAgentRoot|TestMigratePiRootInspectionFailure|TestMigrateFailedRelinkKeepsOldLink|TestMigrateRecoveryCleansOwnedTemp|TestMuseAuthRepairAdditionalStates|TestNativeSkillsNarrowingMutants|TestWeightConflictErrorAndWarning|TestPathRootAndImportedMCPDeclarationsRefused|TestResolveRejectsGroupWritableOnlyPathOverlay|TestMissingOverlayPathIsPathMissing|TestSymlinkInPathIsUntrusted|TestPreserveLockMigrationRefusalPrecedesBusyLock|TestPlainRepairMigrationKeepsLockPrivate|TestReservedSkillName|TestSurfacingUnreadableManifest)$'
```
Exit: 0. Elapsed: 58.769s.

### install-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestSubstitutionsABAAroundTheReadRestartsClosure|TestRepositoriesThatNeedNoCredentialSkipSelection|TestConfiguredPathsExpandTheOperatorHome|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptListsTheUnlistedIdentityCount|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestDraftAuditBindingLifecycle|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftInstallStaleLockFails|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftTransportPlanConversion|TestDraftManagerExecutable|TestGitFixtureNeverRetriesNonZeroExit|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalStrictTagsDetectMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestRegistryBootstrapVectorsDriveInstall|TestRegistryMirrorComparisonVectorsDriveInstall|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache)$'
```
Exit: 0. Elapsed: 31.447s.

### install-1

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestPromptAbortFailsClosedWithoutPersistingAnything|TestBuildPublicationFailuresAreRedactedInTheResult|TestDraftAuditUnreadableBindingRefuses|TestDraftBuildFinalRootParentsAreStoreCreated|TestDraftBuildRefusesCompilerViewDigest|TestDraftSelectorSubstitutionForbidden|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftInstallMissingLockFails|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestInstallRefusesUnownedSelectedCanonicalDestination|TestHybridCanonicalOwnershipPreservation|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestEnforcedScriptCommandInstallsNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestStagedOutputsStayPrivateAndAreReleased)$'
```
Exit: 0. Elapsed: 77.811s.

### install-2

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAResolverThatCoversNothingStillFailsClosed|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestConcurrentProjectInstallsPreserveBothConsumers|TestRecoveryCompletesBeforeAnyNewMutation|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditBrokenReportRefuses|TestDraftAuditMalformedShapeRefuses|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftTransportPlanV2Conversion|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitignoreGateSkips|TestRuntimeOnlyProviderStillRequiresSkillMd|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestEnforcedScriptCommandIsRefusedAtInstall|TestCorruptCacheEntryIsRebuiltAndNeverReused)$'
```
Exit: 0. Elapsed: 42.219s.

### install-3

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestRunCommitWithoutBoundariesSkipsRecheck|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftSameTagMoveInstallsWithoutWarning|TestDraftLocalTamperedFrozenRuntimeRefuses|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftSSHWrapperBase|TestAuthoritativeDryRunCasesMutateNothingPersistent|TestDocumentGenerationIsStableWhenUnreadable|TestProjectInstallRecordsShellHookApprovals|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneRecordsDependencyDirectory|TestDraftLaneStillBlocksNULBearingMember|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestRunCommitWithRealStagingJournalsOneTransaction|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestReleaseTakesNoToolchainVerdictAfterLiveMutation|TestProjectWriteBlobsSpawnFailureIsWrapped)$'
```
Exit: 0. Elapsed: 20.778s.

### install-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestGlobalCommitCarriesNoConsumerLedger|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditRenewalMarkersNeverRenewable|TestLegacyBuildsKeepReceipt1|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestGitFixtureRetryPredicate|TestRemovedSkillCleanedUp|TestHybridShadowedByProjectDeclaration|TestLegacyLaneRejectsUnboundPreviousGeneration|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged)$'
```
Exit: 0. Elapsed: 63.582s.

### install-5

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestConfiguredAgentResolvesToTheLiveSocket|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftRepairRevalidatesLockedSource|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestRevision1ReaderRejectsV2Shapes|TestExternalCommandNameCollisionFailsBeforeMutation|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestEndToEndInstall|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestSecondInstallIsUpToDate|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot)$'
```
Exit: 0. Elapsed: 9.284s.

### install-6

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestScriptAuditLabelsAtInstallEntry|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestPromptNumberedSelectionAndAnExplicitScope|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftInstallUsesPinnedBytes|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftInstallGitTamperRefused|TestDraftTransportAuth|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGlobalInstallRecognizesAdoptedForwardingShim|TestTamperTriggersReinstall|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestCacheHitOnlyPlanStillFinalizesToolchainTrust)$'
```
Exit: 0. Elapsed: 29.709s.

### install-7

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesSwappedParent|TestDraftAuditRenewalNeverHidesRefusal|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftMarkerArmGuards|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftInstallGitPinnedSubtree|TestDraftPlanNeedsSSH|TestUserConfigIgnoredByResolvedLane|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestLegacyLaneChangedTagAfterOldTagDeleted|TestStrictRegistryPolicyFailsUnknown|TestStableDeclarationInputsCommitWithoutRestarting|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestActiveEnforcedScriptCommandsAdmits|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestCacheHitPerformsNoSourceAwareGoCommand|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure)$'
```
Exit: 0. Elapsed: 25.978s.

### atomicity

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install/atomicity -json -count=1 -timeout=8m
```
Exit: 0. Elapsed: 50.054s.

### rest-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve
```
Exit: 0. Elapsed: 95.112s.

### rest-1a

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment
```
Exit: 0. Elapsed: 1.103s.

### rest-1b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```
Exit: 0. Elapsed: 4.861s.

### rest-1c

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/gitops github.com/relux-works/curator/internal/globalbins github.com/relux-works/curator/internal/hashing github.com/relux-works/curator/internal/hookapproval
```
Exit: 0. Elapsed: 5.72s.

### rest-1d

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/identifiers github.com/relux-works/curator/internal/identity github.com/relux-works/curator/internal/interop github.com/relux-works/curator/internal/interop/environments github.com/relux-works/curator/internal/locale
```
Exit: 0. Elapsed: 1.31s.

### rest-2a

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/managerlock github.com/relux-works/curator/internal/manifest github.com/relux-works/curator/internal/marker github.com/relux-works/curator/internal/mcp github.com/relux-works/curator/internal/moduleroots
```
Exit: 0. Elapsed: 2.21s.

### rest-2b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/nodesource github.com/relux-works/curator/internal/npmsource github.com/relux-works/curator/internal/opaquescan github.com/relux-works/curator/internal/pathboundary github.com/relux-works/curator/internal/pkgversion
```
Exit: 0. Elapsed: 11.078s.

### rest-2c

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/pnpmsource github.com/relux-works/curator/internal/privatedir github.com/relux-works/curator/internal/protocoljson github.com/relux-works/curator/internal/registry github.com/relux-works/curator/internal/runtimestore
```
Exit: 0. Elapsed: 11.175s.

### rest-2d

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/rustsource github.com/relux-works/curator/internal/scopes github.com/relux-works/curator/internal/scriptpolicy github.com/relux-works/curator/internal/scriptworker github.com/relux-works/curator/internal/shell
```
Exit: 0. Elapsed: 97.258s.

### rest-3a

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/skillcheck github.com/relux-works/curator/internal/skillspec github.com/relux-works/curator/internal/snapshot github.com/relux-works/curator/internal/sourcelock github.com/relux-works/curator/internal/staging
```
Exit: 0. Elapsed: 2.089s.

### rest-3b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/stateread github.com/relux-works/curator/internal/swiftpmbuild github.com/relux-works/curator/internal/swiftpminterop github.com/relux-works/curator/internal/swiftpmsource github.com/relux-works/curator/internal/testcli
```
Exit: 0. Elapsed: 22.744s.

### rest-3c

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/testgitenv github.com/relux-works/curator/internal/testtoolchain github.com/relux-works/curator/internal/transaction github.com/relux-works/curator/internal/ui github.com/relux-works/curator/internal/verr
```
Exit: 0. Elapsed: 21.258s.

### rest-3d

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/internal/version github.com/relux-works/curator/internal/whitelist github.com/relux-works/curator/internal/yarnclassicsource github.com/relux-works/curator/internal/yarnmodernsource github.com/relux-works/curator/tools/ciworkflow
```
Exit: 0. Elapsed: 42.87s.

### rest-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -p 2 -count=1 -timeout=8m github.com/relux-works/curator/tools/goreleaserconfig
```
Exit: 0. Elapsed: 0.585s.

### cli-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestAuditAllowRefusalPrecedesConfigLoad|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestProductionExternalDepsAssignsTransportProvenanceSink|TestCredentialStoreCLIInvalidProfileRouting|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvConfigShowPermissionsWithoutProfiles|TestProfileComposeAddListRemove|TestHookApproveRecordsOperatorApproval|TestHookApproveResolvesSymlinkedOperand|TestHookRevokeAbsentLeavesStateByteIdentical|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestNativeSkillsTakeoverActivationBoundary|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallUsePartialLeavesCurrent)$'
```
Exit: 0. Elapsed: 5.858s.

### cli-1

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCLIStatusReportsLegacyPackageMarkerCurrent|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestStatusJSONCarriesShellHookTrust|TestRunUsesInjectedConfigSourceAndWriters|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestMuseProfileSeedSecretAndCredentialRefusals|TestProfileInstallRefConflictIsAFailure|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallUseActivatesThroughSwitch)$'
```
Exit: 0. Elapsed: 18.009s.

### cli-2

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestStatusUnreadableApprovalStateSurfaced|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestMuseRootContextCopyFallback|TestProfileInstallStopThenRetryConformanceVector|TestProfileMachineUseSkipsScopedAdapter|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureCuratorStatusRows)$'
```
Exit: 0. Elapsed: 19.012s.

### cli-3

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestInputCausesAreDistinctAndDocumented|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI|TestEnvResolveRepairEmitsFragment|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestNormalizeEnvKnobPositions|TestEnvStatusCheckCurrentScopeOnly|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApproveEndToEndWithGeneratedHook|TestProfileUpdateListsNewDeclaration|TestSecurityPostureProfileInstallRefusesMCPDeclarations)$'
```
Exit: 0. Elapsed: 31.638s.

### cli-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCommandHelpUnknownTopic|TestDraftRemediationThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestRunUnknownCommand|TestProfileInstallReinstallDeltaVectorsAtCLI|TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist)$'
```
Exit: 0. Elapsed: 31.141s.

### cli-5

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestProjectResolveGitAliasSelectionThroughCLI|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestMuseCLIAuthParentBoundary|TestProfileUseUnknownEnvironmentIsAFailure|TestSecurityPostureVectorsThroughCLI|TestSecurityPostureWarningIsNotEmittedByEnforcedShim)$'
```
Exit: 0. Elapsed: 5.267s.

### cli-6

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestCredentialStoreCLIDottedProfileRouting|TestEnvMigrateApplyRequiresPlan|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvStatusMatrix|TestProfileInstallUseLockedRequireRefuses|TestEnvStatusReportsLockedRequire|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestStatusRecordedButMissingStaysInInventory|TestEnvStatusMissingAndUnreadableKeepRecord|TestHookApproveIsIdempotentWhenRecordMatches|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallListsStdioCommands)$'
```
Exit: 0. Elapsed: 34.857s.

### cli-7

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestClassifySkillBuildsDetectsContextExposure|TestStatusReportFindsASchema8InstallationCurrent|TestDraftAttemptTransportTable|TestDraftDocsPinExamples|TestProjectResolveGitSelectionThroughCLI|TestDraftTransportProviderAdmission|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestEnvConfigSetNormalizesEnvKnobs|TestGCWaitsForTheHomeLock|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileScopedUseAndClear|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestUmbrellaMissingNamesRoots|TestUmbrellaWarningText|TestExecutableCandidates|TestUmbrellaTrustRootCaseVariantIdentity)$'
```
Exit: 0. Elapsed: 30.37s.

### cli-8

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestRegisteredSubcommandHelpNoWrites|TestDraftStatusStaleSwappedSelectionRefuses|TestClassifyDraftMemberPresenceRows|TestEnvMigrateConflictRefuses|TestEnvMigratePrintBeforeWrite|TestEnvResolvePreserveLockRequiresRepair|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestGlobalInitSubcommandHelpWithoutConfigRefuses|TestPathOverlayFromMachineConfigJoinsTheClosure|TestRunVersionExitsZero|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestProfileInstallReinstallUsePreservesProjectScope|TestGitFileURLHasNoBackslash)$'
```
Exit: 0. Elapsed: 26.821s.

### cli-9

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestBuildReportsNeverPublishAnAbsolutePath|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusContentDrift|TestEnvResolveRepairNeedsMigration|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestGCFailsClosedForUntrustedCurrentPathSource|TestHookUsageErrors|TestStatusReportsShellHookTrustPosture|TestProfileImportLossyRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestMuseRootContextStatus|TestProfileInstallFileOperandIsRefused|TestInstallPrintsTheRemedyAVersionManagerSelectionEarns)$'
```
Exit: 0. Elapsed: 17.864s.

### cli-10

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCLIAuditMixedSkillRecordsBothPolicies|TestDraftAttemptClauseShape|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestDraftTransportLegacyGolden|TestCredentialStoreCLIPromotion|TestUmbrellaMissingProvider|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestProfileComposeRefusals|TestSyncTakeoverFlagParses|TestBootstrapAndProjectCommands|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestSecurityPostureWarningAndEnvStatusRows|TestStatusExplainsAnUnusableGoToolchain|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands)$'
```
Exit: 0. Elapsed: 20.999s.

### cli-11

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCLIAuditJSONRecordsExplicitAbsence|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestBuildReportDetailIsBounded|TestDraftDocumentedGitRevisionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestSignerRowsSurviveHostileGlobalGitConfig|TestHookApproveReRecordsAfterChange|TestInstallGitignoreEntryPoint|TestProductionBinaryDispatchesNetNSProbe|TestCLIAuditOpaqueNULFollowsHashVersion|TestUmbrellaTrustedRootSymlinkEscape)$'
```
Exit: 0. Elapsed: 10.72s.

### cli-12

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCommandHelpNarrowingMutants|TestCommandHelpInitPositiveControl|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestDraftStatusStaleManifestRefuses|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestNativeSkillsCLIOperations|TestProfileInstallSurfacesBeforePublication|TestProfileInstallListUse|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestCompiledProjectRepairsCorruptCompiledState|TestProviderSubcommandParsing)$'
```
Exit: 0. Elapsed: 90.728s.

### cli-13

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestProjectResolveLegacyMixedThroughCLI|TestEnvUnmanageTakeoverPreservesFileMode|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestHookApprovalsListsReadOnly|TestProfileComposeAddRefusesAFormOnAPathSource|TestShellInitInstallCachesHookWithoutConfig|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestMuseFragmentV4PublishedCases|TestProfileUpdatePinnedTagIsUnchanged|TestCompiledProjectStatusAndUntrustedRecovery|TestCompiledProjectRestoresCacheWhenCommitFails|TestProviderInputsForHostExplicitBinRefuses)$'
```
Exit: 0. Elapsed: 122.671s.

### cli-14

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestEnvMigratePlanApplyPi|TestImplementedCommandWinsOverProvider|TestEnvUnmanageBackupRecordVectors|TestStatusUnreadableCandidatesKeepRecord|TestProfileComposeAddPathRow|TestProfileUseTakeoverRow|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestInstallFlagsAcceptTrailingOptions|TestCLIEndToEndInstallStatusAndTamperCheck|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestMuseRootContextDrift|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverHelpNarrowingMutant|TestProfileUseTargetIsStageB|TestSecurityPostureRevisionBDefaultRefusesEmptySources)$'
```
Exit: 0. Elapsed: 63.038s.

### cli-15

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCommandHelpRunPassthrough|TestDraftStripCloneFramingTable|TestReviewDottedCredentialProfile|TestGCRetainsLiveProcessBuild|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIHelperProcess|TestStatusAcceptsAnUnchangedLegacyMarkerSchema)$'
```
Exit: 0. Elapsed: 27.978s.

### cli-16

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvResolveAcceptsAliases|TestPostureWarningOncePerProcessTree|TestHookApprovalsOnAbsentStateIsEmpty|TestStatusDriftDetectsContentTampering|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestRevisionDoesNotBorrowTagSignature|TestCuratorStatusProviderPostureAndCheck|TestProviderPostureRows|TestProviderInputsForHostScannedBinWarnsWithoutLedger)$'
```
Exit: 0. Elapsed: 8.529s.

### cli-17

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestDraftInstallStatusHelpIncludesSchema2Workflow|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestDraftStatusDriftFailsClosedOnStaleLock|TestEnvConfigLockedKnobRefuses|TestBootstrapDocsExampleRuns|TestGCPrunesDeadConsumersUnderTheHomeLock|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusRejectsPositionalArguments|TestHookDiagnosticsAgreeWithShell|TestResolveTakeoverRequiresRepair|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestProfileInstallWarnsOnSystemModule|TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands|TestUmbrellaSymlinkIntoManagedRefused|TestProviderInputsForHostPublishedBinRefuses)$'
```
Exit: 0. Elapsed: 13.455s.

### cli-18

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestCredentialStoreCLIConfigRefusesMalformed|TestEnvResolveBareStaleIsAFailure|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestProductionExternalDepsBindTrustedGitAndAudit|TestRunNoArgsPrintsUsage|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestNativeSkillsTakeoverArgumentDiscardRegression|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestUmbrellaHostilePathPlantRefusedUnderB)$'
```
Exit: 0. Elapsed: 23.334s.

### cli-19

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestClassifyDraftMemberComparesExactCurrentness|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestEnvMigrateUsage|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvConfigProviderDirectoriesRoundTrip|TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestNativeSkillsTakeoverEveryRegisteredCommand|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestStatusReportsATransitivelyResolvedCompiledCommand|TestCompiledInstallFollowsTheNativeControlInventoryExactly)$'
```
Exit: 0. Elapsed: 56.437s.

### cli-20

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestDraftDocumentedLocalShapesThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProfileImportNameTakenRow|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestProfileInstallActivationMatrix|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestHostGoToolchainIsSelectableOnAnInventoryPlatform)$'
```
Exit: 0. Elapsed: 48.515s.

### cli-21

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestEnvResolveUnknowns|TestEnvStatusReportsNotInheritedNativeCodexServers|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestReviewNativeTakeoverHelpBoundary|TestProfileSyncCoversTheMachineScope|TestProfileComposeListsDeclaredOverlays|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory)$'
```
Exit: 0. Elapsed: 14.565s.

### cli-22

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestProjectResolveLegacyNetworkGitThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestEnvConfigShowDefaults|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestSkillCheckOnTempDir|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileUseTargetBoundToAdapter)$'
```
Exit: 0. Elapsed: 4.874s.

### cli-23

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestProfileScopedUseUnaffectedByLockedRequire|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestShellInitPrintsHooks|TestInstallAuditFlagAcceptsOptionalMode|TestMuseCLIStorePolicy|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileReinstallActivationFailureKeepsUpdatedWording|TestSecurityPostureRevisionBDefaultEnvStatus|TestUmbrellaProviderResolutionVectors|TestUmbrellaRefusedDirectoriesBothRevisions)$'
```
Exit: 0. Elapsed: 41.003s.

### cli-24

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestDraftAvailabilityFallbackThroughCLI|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaPropagatesExitCode|TestProfileImportUseLockedRequireRefuses|TestGCRunsSerializedAcrossConcurrentInvocations|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestProfileImportRow|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextVectors|TestMuseCredentialPolicyStatus|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileInstallWarnsEmptyAllowlist)$'
```
Exit: 0. Elapsed: 85.045s.

### cli-25

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestDraftStatusStaleEmptyLockRefuses|TestUmbrellaDispatchesToProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetUnsetRoundTrip|TestProfileUseLockedRequireRefuses|TestHookApprovalsUnreadableStateNamesReadFailure|TestConfigSubcommandDispatch|TestNativeSkillsTakeoverFlagBoundary|TestStatusCheckAdmitsV2NULBearingInstall|TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy)$'
```
Exit: 0. Elapsed: 6.939s.

### cli-26

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestDraftFetchFailureSanitizedThroughCLI|TestEnvStatusRegistryBoundaryPostureAndCheck|TestUmbrellaInvalidNameIsUsage|TestHookRevokeRemovesRecord|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestMuseFragmentV3PublishedCases|TestProfileUseUnknownIsAFailure)$'
```
Exit: 0. Elapsed: 17.582s.

### cli-27

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCLIAuditBindsDependencyDirectory|TestCLIAuditEmitsScriptLabels|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestEndpointExhaustionCarriesRemediation|TestProjectResolveLocalCreatesLockThroughCLI|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestEnvResolveAliasAdjacentUnknownsRefused|TestEnvConfigProviderDirectoriesLockedRefuses|TestProfileUseClearOperandIsUsage|TestProfileUpdateResyncPassesLockedRequire|TestEnvResolveStaleHintsRepair|TestBootstrapIfMissingRejectsForce|TestProductionBinaryDispatchesLandlockProbe|TestConfigBuildHTTPSAddListRemove|TestAuditAllowWritesWriterVersionPinCarrier|TestDryRunNeverClaimsACompletedCompilerCheck|TestUmbrellaUnreadableRootNeverAbsence)$'
```
Exit: 0. Elapsed: 12.297s.

### cli-28

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftTransportResolvedMatrix|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestProfileInstallFirstActivationLockedRequireRefuses|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitHelpWithoutConfig|TestProfileUpdateSurfacesBeforePublication|TestProfileRemoveInUseIsAFailure)$'
```
Exit: 0. Elapsed: 13.416s.

### cli-29

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestCLIAuditJSONCarriesScriptLabels|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusInvalidMarker|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestProductionBinaryDispatchesSSHWrapper|TestReviewDottedCredentialProfileExisting|TestUsageListsAliases|TestGCPreservesRuntimeOnUncertainMarks|TestHookApprovalsToleratesMalformedRecord|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestBootstrapIfMissingCreatesAbsentConfig|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestNativeSkillsTakeoverRegistryNarrowingMutant)$'
```
Exit: 0. Elapsed: 39.878s.

### cli-30

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestAuditAllowRecordsPinCreationTime|TestWithDraftRemediationTable|TestDraftProjectResolveHelp|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestDraftStatusNotInstalled|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestEnvResolveRepairFailedOnUninspectable|TestProfileUseEnvAliases|TestProfileUseUnknownEnvRefused|TestOverlayFromMachineConfigIsRefusedByKind|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestMuseRootContextRefusals|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestActiveRevisionIsWarningRelease|TestUmbrellaTrustRootSelectionBothRevisions|TestProviderPostureReportsUnreadablePublishedShimLedger)$'
```
Exit: 0. Elapsed: 24.051s.

### cli-31

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/cmd/curator -json -count=1 -timeout=8m -run '^(TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestProjectResolveIgnoresRemovedSwitch|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestCredentialStoreCLIConfigRoundTrip|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveLockedIsolationRequiresMigration|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestLegacyPackageStatusBindingAndRepair|TestNativeSkillsTakeoverHelpCrossProduct|TestSourceSignerVectorsAtProfileInstall|TestGCRetainsAndReportsReferencedCompiledState|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd)$'
```
Exit: 0. Elapsed: 35.547s.

### cross-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftLiteralIgnoresGitSSHCommand|TestDraftSourcesCrossCompile|TestDraftSourcesSnapshotVectors)$'
```
Exit: 0. Elapsed: 18.918s.

### cross-1

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftSourcesCLILocalSkillScriptDependencies|TestDraftLiteralIgnoresProxyEnvironment|TestDraftExternalExecutionPolicyInvalidMarkerReDerives|TestDraftSourcesSemanticCasesBatch2|TestDraftSourcesSemanticCasesBatch4|TestDraftSourcesCLIInstallRestoresPriorState|TestDomainSeparationChangesIdentity|TestAcceptedCorpusSatisfiesEveryPublishedStructuralClaim)$'
```
Exit: 0. Elapsed: 38.54s.

### cross-2

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestSkillfileSourcesPin|TestSkillfileSourcesCorpusCounts|TestDraftSourcesSemanticCasesBatch0|TestDraftSourcesSemanticCoverage|TestIndependentEncoderMatchesProductionEncoderOnSharedValues|TestTypedAccessorsRefuseWrongKinds)$'
```
Exit: 0. Elapsed: 11.259s.

### cross-3

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftDigestSkipsManagerLocksOnEverySeparator|TestDraftTreeDigestIgnoresManagerLockFiles|TestDraftLiteralKeepsAgentAndAskpass)$'
```
Exit: 0. Elapsed: 2.514s.

### cross-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftSourcesCLIProjectPathWithSpaceAndUnicode|TestIndependentOracleAgreesWithProductionIdentity|TestIntegrationProductionSourceImportsNoRepositoryPackage|TestRustUnavailableCoverageRejectsExtraRustGap)$'
```
Exit: 0. Elapsed: 1.866s.

### cross-5

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftSourcesCLIDirectoryManifestDependencyInstallRefreshAndAudit|TestDraftLiteralDropsGitProxyCommand|TestDraftLiteralIgnoresGitExecPath|TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI|TestDraftSourcesReplayRejectsLockedObjectFormatMismatch|TestDraftSourcesSemanticCasesBatch1|TestValidatorRejectsEveryTamperedCorpus|TestProtocolExportMatchesTheCommittedGolden|TestDelegatedVectorsNameRealOwningPackages|TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams|TestDelegatedDiagnosticsAreDeclaredByTheirOwners)$'
```
Exit: 0. Elapsed: 41.366s.

### cross-6

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftSourcesBrokerAskpassDispatch|TestDraftLiteralIgnoresUserSSHConfig|TestDraftSourcesSchemaCases|TestDraftExternalStatusFailsClosedWithoutSource|TestDraftSourcesSemanticCasesBatch3|TestDraftLiteralRefreshIgnoresUserConfig|TestCanonicalFormIsAFixedPointForEveryControlCharacter|TestSurrogateEscapesDecodeAndReemitLiterally|TestCrossAdapterConformance)$'
```
Exit: 0. Elapsed: 43.865s.

### cross-7

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/crossconformance -json -count=1 -timeout=8m -run '^(TestDraftLiteralIgnoresGitSSH|TestDraftSSHEmptyConfigIgnoresHostileFile|TestDraftSourcesCLISymlinkMemberRefused|TestDraftTransportFallbackGate|TestRequireCanonicalRejectsEveryNoncanonicalSpelling|TestEmbeddedCorpusIsTheAcceptedBytes|TestAllPublishedRecordsIndependentlyDeriveTheirHashes|TestParseRejectsAPublishedHashThatDoesNotMatchItsBytes|TestProtocolExportIsCanonicalAndSelfDescribing|TestRustUnavailableCoverageRejectsNonRustGap)$'
```
Exit: 0. Elapsed: 4.813s.

### envprofile-2b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestCanonicalGitRejectsFileRemote|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallSurfacingSilentWithoutMCP|TestMachineUseSkipsScopedAdapter|TestManagedHomeRepairTightensDirectories|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureDoesNotProvision|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillPromptScopeValidation|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMusePublishedLinkStates|TestNameTaken|TestOverlayDuplicateNameIsCompositionInvalid|TestPathKindMCPVectorsDriveInstallEntry|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestRC14IdentityMigrationRollsBackEveryEntry|TestResolveAdmitsV2CommitPinnedContextNUL|TestResolveDriftRepair|TestSeedUnreadableStopsBeforeFirstWrite|TestStaleCredentialLinkRefusals|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestUpdatePathLeavesLockInPlace|TestUpdateSystemDeltaRefusesBeforePublication)$'
```
Exit: 0. Elapsed: 29.317s.

### envprofile-3b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestAbsentLockRemainsProfileUnknown|TestClaudeAbsentSeedGolden|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCredentialLinkTargetInspectionFailure|TestCredentialProjectionIdentitySubrequest|TestGitOverlayJoinsClosure|TestGlobalSkillCapabilitiesNarrowingMutants|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallEmitsSurfacingBeforePublication|TestInstallErrorWaiverAdmits|TestInstallSurfacesUnresolvedMCPCommand|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureWaitReselectsAndTimeout|TestMigrateOldRootBytesWarnAndProceed|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrationHonoursSystemLockedAllowlist|TestPassthroughReadFailuresAreUnknownAndNeverRepaired|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestRC14IdentityMigrationRefusesSiblingPlanDrift|TestSourceAllowlistRefusesBeforeClone|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateErrorLeavesLockUnchanged|TestWeightDuplicateRefused)$'
```
Exit: 0. Elapsed: 43.362s.

### envprofile-4b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestClaudeRepairSatisfiedIsNoOp|TestCredentialProjectionSingleObservation|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportInvalidUTF8IsLoss|TestImportLosslessInstallsThroughPathPipeline|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestInstallRejectsSecretMember|TestInstallSurfacesSystemModuleWarning|TestMachineClearReachesSeamGate|TestMaterializeCaptureObjectBounds|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMuseStatusOptionalProvisioning|TestOpencodeXDGSeeds|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPrecedencePrimitivesDriveEmission|TestPreserveLockRefusesBeforeStoreRebuild|TestProfilePathCollision|TestStatusErrorReportsPolicyWithoutDropped|TestSkillDirectoryLifecycleVectors|TestSkillDirectoryMetadata)$'
```
Exit: 0. Elapsed: 39.227s.

### envprofile-5b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestCredentialLinkInspectionRepairFails|TestCredentialProjectionIsolationRefusals|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestForeignManagerHintLstatDiscipline|TestIncompleteJournalRecoversOnEntry|TestMCPAllowlistRefusesOutsidePackage|TestMachineUseClearsEveryScopeEqualToNewDefault|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureSerializesProfileWriters|TestMigrateApplyRequiresPlan|TestMigratedGlobalSkillNullCapabilityDeclarations|TestMissingPathOperandIsPathMissing|TestMutationLockContention|TestNativeSkillsRepair|TestNestedGitIsSourceInvalid|TestOperationsLeaveNoJournalResidue|TestOverlayInvalidSourceKindIsSourceInvalid|TestPathInstallCapturesDirtyUntrackedInsideGit|TestReinstallEmitsSurfacingBeforePublication|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestRevokedSkillMemberIsRefused|TestStrictAuditMemberRefusesV1NULBeforeHashing|TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestUpdateSurfacesCandidateMCPSet)$'
```
Exit: 0. Elapsed: 26.433s.

### envprofile-6b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestCanonicalIdentityRejectsMalformedSource|TestCredentialLinkDirectory|TestCredentialProjectionStatusObservation|TestEnvironmentsCodexSeedVectors|TestImportDivergentSkillsAreLoss|TestImportSkillMarkerRecovery|TestImportUnreadableMarkerIsLoss|TestImportUnreadableRootIsLoss|TestInstallAdmitsV1CollidingSkillTreesUnderV2|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrationHonoursConfigPathOverride|TestMusePolicyExplicitMigration|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestReinstallBlocksRevokedOverlayMember|TestReinstallBlocksSecretOverlayMember|TestStoreBoundaryStatusVectorsDriveStatus|TestUnreadableCodexSeedDoesNotSelectAbsentDefault|TestValidateNamedStorePinsRefusesV1NULBeforeHashing)$'
```
Exit: 0. Elapsed: 9.733s.

### envprofile-7b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/envprofile -json -count=1 -timeout=8m -run '^(TestAbsentPathOperandWithRequirementIsRefConflict|TestCodexIsolatedAdmission|TestCodexMalformedStoreSharedRefuses|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCredentialLinkUnrecordedSymlinkRefuses|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportSkillForeignPinnedByRevision|TestManagedFallbackPermissions|TestManagerReadScannerFindsAliasedNotExistCollapse|TestMaterializeCapturePendingDoesNotRecover|TestMigratePendingNeedsNothing|TestNativeSkillsUnreadableMarker|TestNativeSkillsUseAndSync|TestNonRootWeightsMapRefused|TestOverlayExplicitWeightOverridesDefault|TestPathOverlayMCPDeclarationRefused|TestPathSnapshotImmutableAcrossUpdateSyncUse|TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestRepairFailedMissingStore|TestResolvePassthroughLiveness|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestScopedUseStillSwitchesOnlyThatHome|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL|TestUpdateBlocksRevokedOverlayMember|TestUpdateBlocksSecretOverlayMember|TestUseTakeoverBacksUpAndNotifies)$'
```
Exit: 0. Elapsed: 18.276s.

### install-2b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestDefaultToolchainProbeRemovesItsProbeRootOnFailure|TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess|TestDraftAuditAdvisoryLocalPasses|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftRepairRestoresDriftedContent|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDryRunTouchesNothing|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestGitFixtureOuterEntryRunsRealGit|TestHybridActivationABAAroundTheReadRestartsClosure|TestHybridSkillActivatesWithoutTouchingProjectStore|TestIncompletePlansNeverClaimACompleteDiagnostic|TestMaintenanceFailureAfterCommitIsAWarning|TestMovedTagWarningAndStrict|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestRegistryAttestationLandsInMarker|TestRegistryRevocationDeniesInstall|TestReviewExternalExecutionBindsExactReceipt3Input|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestScriptOptInCasesAtInstallEntry|TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall|TestV2RegistryAttestationUsesVersionedInstallPath|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch)$'
```
Exit: 0. Elapsed: 36.293s.

### install-3b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestDeclarationGenerationIsContentAddressed|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftPlanNeedsSSHRevision2|TestDraftRuntimeKeysCoverAllArms|TestGlobalEnforcedInstallExcludesForwardingMirror|TestGlobalInstallRejectsSchema2Skillfile|TestHTTPSPromptAbortNeverPersistsOrSelects|TestInjectedClockAndGenerationReaderDriveInstallMarkers|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneRefusesGlobDependencyDirectory|TestOneScopeChoiceCoversEverySiblingRepository|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestReviewChangedDeclaredTagIsNotMovedTag|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestScriptAuditLabelsAtGlobalInstallEntry|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$'
```
Exit: 0. Elapsed: 34.834s.

### install-4b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAuditGateBlocksUndeclaredNetwork|TestDefaultAcquireFetchesTheDeclaredURL|TestDraftAuditDuplicateKeysRefuse|TestDraftEvidenceExactMatch|TestDraftGitMarkerLockBindingRoundTrip|TestDraftLocalEnforcedCommandRefused|TestDraftTransportEnabled|TestDraftUnmanagedAdapterDestinationRefuses|TestEnforcedInstallAndLaunchAtCLIEntry|TestGlobalCanonicalOwnershipPreservation|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation|TestLegacyLockPackageArms|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptWithoutCandidatesStillOffersTheManualPath|TestProvenanceNamesEverySourceOnePerRepository|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates)$'
```
Exit: 0. Elapsed: 40.426s.

### install-5b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAuthoritativeCacheOutcomesDriveInstallation|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestDraftAuditHostileContentBlocked|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftRefreshThenInstallReplacesLockAndMarker|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestInstallDraftMarkerWriterCases|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestLegacyLaneRuntimeMatchesPackageMarker|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestPostCommitMaintenanceWarningsReachTheResult|TestProjectDryRunRecordsNoShellHookApprovals|TestProjectNonGitWithDevSubstitution|TestProjectRefusesOnGitignoreSpawnFailure|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestV2InstallDistinguishesNULCollisionTwins)$'
```
Exit: 0. Elapsed: 24.175s.

### install-6b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestAPersistFailureStopsTheRunRatherThanContinuing|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestCacheInspectionIsBracketedByTheFrozenSource|TestConfiguredScopesAreSelectedPerRepository|TestDraftAuditClosedPackageShapeRefuses|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallReplaysMissingLocalSnapshot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGitFixtureNeverRetriesWrappedExitError|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestInstallDraftOwnershipCases|TestInstallRefusesUnprovenStaleOwnership|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestPlannedBuildAccessorsAreImmutable|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestRedactDiagnosticCannotDriveATerminal|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestStageProjectTargetsRefusesMissingSnapshot|TestStagingRunsProviderFirstAndCommandLexical|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestV1ReinstallRefusesNULAppearingAfterInstall)$'
```
Exit: 0. Elapsed: 72.266s.

### install-7b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestAcquireDraftNetworkSelection|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditRevocationBlocks|TestDraftGitBuildsPublishReceipt3|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalSkillDependencyRefused|TestDraftPlanNamesProvider|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestDryRunReportsCacheHitWithoutBuilding|TestExternalAdoptionsListEveryPublishedEntryOnce|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGlobalManifestABAAroundTheReadRestartsClosure|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestMcpRequirementGatesInstall|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestScriptOnlyInstallPerformsNoToolchainOrCacheWork|TestStageGlobalTargetsPopulatesBoundaries|TestV1InstallRejectsV2RegistryEvidence)$'
```
Exit: 0. Elapsed: 15.39s.

### godriver-0

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestCapabilityEvidenceRecordMatchesTheAcceptedVector|TestBuildEnforcesOutputDeadlineDiskAndNonzeroExit|TestCapabilityEvidenceConsistencyRules|TestUnknownGoFamiliesRemainRefused|TestFingerprintTraversalMatchesLegacyOnRealToolchain|TestToolchainFramingMatchesRC4Vector|TestFingerprintRejectsInvalidUnicodePath|TestPackageGraphValidatesStandardMainModuleAndVendoredTransitiveInputs|TestVerifyArtifactRejectsHardLinks|TestWorkerEnvironmentGuardAcceptsOnlyTheClosedOfflineEnvironment|TestExecutableIdentityResolvesARealSymlinkedProcessLaunch|TestModuleRootsBuildAMultiModuleVendoredSnapshot|TestDeclaredModuleRootsRequireVendorMetadata|TestVendorMetadataMustBeARegularFile|TestALinkStandingInForVendorMetadataNeverReachesTheDriver|TestTheDeclaredModulesSurfaceIsClosed|TestCandidateLinkOutsideDerivedGOROOTFails)$'
```
Exit: 0. Elapsed: 11.306s.

### godriver-1

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestBuildFailsClosedWhenTheGoChildCannotStart|TestIdentityProtocolAndPackageInfluenceCodesMatchTheAcceptedVector|TestBuildStopsBeforeBuildForEveryPreflightRejectionClass|TestBuildRejectsHardLinkedArtifact|TestMandatoryControlSetIsExactAndCarriesNoHardenedClaim|TestCapabilityEvidenceIsNotACacheOrReceiptInput|TestOSExecutorEnforcesOutputAndTimeBounds|TestFingerprintIsStableAcrossRepeatedRuns|TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot|TestFingerprintRejectsDuplicateEncodedClaim|TestUnixLauncherRequiresExecuteBit|TestBuildAcceptsAManagerStartedThroughALauncherLink|TestAuditedVendorAllowancesAreWithheldFromAReplacedModule|TestVendoredGeneratorCarveOutDoesNotHideACgoImportDynamic|TestSelectionPriorityAndNoPathLookup|TestConfigFromEnvironmentAdmitsOnlyTrustedSelectors|TestProbeFailuresCleanPrivateStateAndHaveStableDiagnostics|TestReleaseIsCleanupOnly)$'
```
Exit: 0. Elapsed: 7.015s.

### godriver-2

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes|TestControlSeamFailureNeverDriftsIntoAHardenedClaim|TestNativeControlInventoryMatchesTheAcceptedVector|TestCandidateGoV1SourceAwareContract|TestBuildRejectsDirectivesPGOAndMultipleRootsBeforeBuild|TestPerFileSizeLimitIsReallyApplied|TestFixedEnvironmentForHostSelectsNativeCase|TestValidPackageGraphVectors|TestFingerprintTraversalMatchesLegacyOnRepresentativeTrees|TestFingerprintCancellationStaysFailClosed|TestDigestCopyDiagnosticPrecedence|TestFingerprintDoesNotDescendLinkedDirectories|TestToolchainWalkRejectsDirectoryReplacedByFile|TestDirectiveScanFindsExactTokenAcrossReadBoundary|TestAppliedControlReportMustAgreeWithItsEvidence|TestPackageCommandSurfaceIsClosed|TestAnEmptyDeclarationIsTheSchemaSixSurface|TestModuleRootVectorsDriveTheWholeBuild|TestWorkerRejectsAProbeThatDropsAPlatformAvailableControl)$'
```
Exit: 0. Elapsed: 4.123s.

### godriver-3

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestBuildRejectsArtifactLinksOversizeAndUnexpectedOutputs|TestBuildClassifiesOfflineVendorAndToolchainFailures|TestBuildRejectsEveryPackageInfluenceSurface|TestToolchainIdentityVectors|TestNativeControlInventoryIsExhaustiveAndClosed|TestToolchainWalkAnchorsDirectoryAndLinkMetadataAtTheRoot|TestToolchainWalkTakesDescentFromTheListedEntry|TestGoVersionLFAndCRLFNormalizeToSameIdentity|TestFingerprintRejectsSpecialFile|TestPackageGraphRejectsMalformedMissingAndDuplicateResults|TestFixedBuildArgvGuardAcceptsOnlyTheProtocolVector|TestModuleRootsRejectAContainmentCollisionWithARuntimeRoot|TestModuleRootsRejectAContainmentCollisionWithASiblingBuildRoot|TestRealGoV1ModuleRootsBuildIsBoundedAndNotLaunched|TestWorkerReturnsABoundedResultWhenTheGoChildCannotStart)$'
```
Exit: 0. Elapsed: 4.613s.

### godriver-4

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestDeferredGuaranteesAndFailureBoundaryMatchTheAcceptedVector|TestPortableExecutionPolicyIsABindingCacheInput|TestQualifiedGoFamilies|TestFingerprintTraversalMatchesLegacyOnErrorPrecedence|TestFingerprintDigestPhaseResolvesReplacedAncestors|TestBuildPermitIsBoundToTheSessionAndVector|TestExecutableIdentityRejectsSubstitutionThroughALauncherLink|TestModuleRootsCheckTheBijectionAfterGoListAndBeforeGoBuild|TestEstablishUsesExactPrivateProbeAndFrozenEnvironment|TestToolchainExecutableMismatchCarriesTheOperatorRemedy|TestProbeRejectsAnUncoveredPlatformBeforeTheWorker)$'
```
Exit: 0. Elapsed: 3.844s.

### godriver-5

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestParentClassifiesAChildStartFailureWithoutBuildContinuation|TestPortableExecutionPolicyMatchesTheAcceptedVector|TestCacheIdentityMatchesTheAcceptedVector|TestBuildEmitsExactlyOneClosedCapabilityEvidenceRecord|TestOSExecutorClosesStdinAndCapturesBoundedOutput|TestGodriverHelperProcess|TestFingerprintDetectsMutationBetweenRuns|TestFingerprintRejectsFileRecordResolvingToADirectory|TestFingerprintReportsUnreadableDirectoryIdentically|TestFingerprintRejectsDanglingLink|TestFingerprintImplementationMatchesRC4ToolchainVector|TestResourceLimitsStayInsideManagerBounds|TestReplacementsAreRejectedWithoutADeclaration|TestAReplacedResultOutsideTheAdmittedSetIsRejected|TestSelectionFallsBackOnlyWhenHigherPriorityIsMissing|TestProbeRemovesStateAndNeverMemoizes|TestDiagnosticRenderingIsUnchangedWithoutARemedy)$'
```
Exit: 0. Elapsed: 2.522s.

### godriver-6

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestWorkerDomainTeardownSurvivesARefusedChild|TestCapabilityEvidenceCasesMatchTheAcceptedVector|TestBuildRejectsOversizedPrivateState|TestBuildRequiresThePackageCommandSurface|TestDriverRejectionClustersMapToStableCuratorErrors|TestDeferredHardenedGuaranteesNeverAppearAndNeverReject|TestFingerprintTraversalMatchesLegacyOnFailClosedTrees|TestCandidateRC4ToolchainArtifacts|TestProbeSetGuardMatchesTheInventoryExactly|TestExecutableIdentityResolvesALauncherLink|TestWorkerProvesTheInstalledIdentityWhenLaunchedThroughALink|TestModuleRootsRejectARequestWhoseDeclarationTheDriverCannotProve|TestUnsafeToolchainCandidatesFailBeforeExecution)$'
```
Exit: 0. Elapsed: 3.617s.

### godriver-7

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/godriver -json -count=1 -timeout=8m -run '^(TestControlDomainInstallsBeforeTheWorkerIsReleased|TestParentRejectsAWorkerRecordThatDiffersFromWhatWasInstalled|TestMandatoryControlsMatchTheAcceptedVector|TestBuildRunsExactlyOneFixedListAndBuildInsideTheWorker|TestBuildRejectsWorkspaceAndPackageSelectedToolchainBeforeGo|TestBuildDetectsSourceMutationThroughTheLastChild|TestBuildTerminatesTheCompleteWorkerDomain|TestRealGoV1VendoredBuildIsBoundedAndNotLaunched|TestFixedEnvironmentAndFiveDirectArgvFormsVector|TestFingerprintRejectsEscapingAndAbsoluteLinks|TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows|TestExecutableIdentityRejectsSubstitutionAndMutation|TestSessionFramingRejectsMalformedAndOversizeFrames|TestDeclaredModuleDirectoriesJoinTheScanSurface|TestTheDeclaredScanSurfaceStopsWhereGoWouldNeverCompile|TestMalformedVersionAndEnvironmentInputs|TestToolchainMutationFailsRecheck|TestRealTrustedGoProbe|TestWorkerRejectsEveryIdentityAndRequestMutation|TestWorkerRejectsEveryProtocolDeviationBeforeTheCompiler|TestWorkerRunsTheCompleteHappySession|TestWorkerModeIsNotReachableThroughPackageData)$'
```
Exit: 0. Elapsed: 16.841s.

### install-1b

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/install -json -count=1 -timeout=8m -run '^(TestADryRunReportsThePerRepositoryCredentialSource|TestASchemaSixCommandDeclaresNoModules|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2Refusals|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftGitInstallWritesMarkerV5|TestDraftLocalMissingSnapshotReplays|TestDraftLocalRejectsConfusedAttestation|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalInstall|TestGlobalReconcilesOwnedLegacyForwardingLink|TestInstallOwnedCanonicalPositiveControls|TestInstallPreservesUnownedStaleStoreEntries|TestLegacyLaneDependencyTagMovement|TestMissingSystemCommandFails|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRunWideAgentAutoRequiresALiveSocket|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestUnsupportedCacheProtectionFailsClosed)$'
```
Exit: 0. Elapsed: 42.363s.

### artifactpolicy-cleanup-rerun

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test github.com/relux-works/curator/internal/artifactpolicy -json -count=1 -timeout=9m -run '^TestDirectoryEntryLimitStopsTheLiveWalker$'
```
Exit: 0. Elapsed: 49.638s.

### research-ungated

```sh
env -u GIT_DIR -u GIT_WORK_TREE "$CURATOR_GO" test -json -count=1 -timeout=2m ./.research/project-exclusion-claude ./.research/project-exclusion-codex
```
Exit: 0. Elapsed: 1.177s.

## Other commands and exit codes

| Command/check | Exit | Result |
|---|---:|---|
| a2a read #curator --last 12 --json | 0 | Recovered the newest trusted request after inbox output was truncated by the tool interface. The inbox exit was not visible in the truncated output. |
| task-board q get(TASK-261021-wt8ve7) with status/assignee projection | 0 | Task exists and correct assignment confirmed. |
| task-board resource get (all three preconditions) | 0 each | Brief and both review preconditions read before ack. |
| a2a post ack in the requested topic | 0 | Ack sent. |
| git fetch origin | 0 | New base pinned above. |
| git switch -c research/project-source-exclusion-claude-r2 origin/main | 0 | Fresh branch from fetched base. |
| git cherry-pick -S 9f66334b48b4765f3011d11f0e1170d678cc4b40 | 1 | Expected README-only conflict; no other conflicting path. |
| git -c commit.gpgsign=true cherry-pick --continue | 0 | README retains both entries; signed transplanted commit. |
| git commit -S -m docs: correct measured Claude child execution mode | 0 | Separate one-row documentation correction. |
| git log -2 --format=%G? | 0 | G for each introduced commit. |
| git diff --check BASE..HEAD | 0 | No whitespace errors. |
| Accepted-payload/README byte comparison script | 0 | All 12 non-README files identical; both complete README entries retained. |
| Native-stats evidence traversal | 0 | Both measured child branches report async/background execution and completion. |
| git clone --no-hardlinks --single-branch --branch research/project-source-exclusion-claude-r2 LOCAL_GIT_DIR VALIDATION_CLONE | 0 | Normal .git clone at exact delivered head, initially clean. |
| go list ./... and go test PACKAGE -list . for each of five large packages | 0 each | Complete default package and shard inventories. Test invocations remove inherited Git variables. |
| Coverage/integrity verification | 0 | Every planned job passed; no missing package or inventoried top-level test; both checkouts clean. |
| git push -u origin research/project-source-exclusion-claude-r2 | 0 | Feature branch published without force. |
| git ls-remote origin refs/heads/research/project-source-exclusion-claude-r2 | 0 | Remote head equals delivered head. |
| git range-diff OLD_BASE..ACCEPTED NEW_BASE..HEAD | 0 | Full output below. |

Tool-discovery diagnostics: `a2a inbox --help` returned 2 (help not supported by that subcommand); top-level help returned 0. An initial board query used unsupported `briefResources` and returned 1; the required status/assignee query and all resource reads then succeeded. Interim coverage reads returned 1 while planned jobs were still pending; these were not test failures. No board precondition or assignee verification failed.

## Range diff

```text
1:  9f66334b ! 1:  64e93191 research: measure Claude project source exclusion
    @@ .research/project-exclusion-claude/testdata/mcp.py (new)
     +    print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "result": result}), flush=True)
     
      ## README.md ##
    -@@ README.md: Curator uses an in-repository task board in `.task-board/`. Automated gates run
    - 
    - ### Tools
    +@@ README.md: Seatbelt, stdio MCP, fixture outputs and expected-red controls; the
    + [qualification record](docs/qualification/project-source-exclusion-codex.md)
    + records the incomplete strict-context boundary.
      
     +The [Claude project-source exclusion harness](.research/project-exclusion-claude/README.md)
     +uses Go, Python 3, Git and macOS `sandbox-exec`. Run
-:  -------- > 2:  0c5c02c8 docs: correct measured Claude child execution mode
```

## Open decisions and bounds

No open decision for the requested rebase and correction. This remains research acceptance only: foreground children, arbitrary late reload, and strict-adapter certification are not qualified. Unsafe-origin resume/fork history retention and all other original bounds remain unchanged. Required independent review and landing are owned by the orchestrator.

After the handover, task-local scratch, copied binaries, disposable validation clones and Go cache scratch are deleted under the requested disk-hygiene policy. This outcome, the prior outcomes, and the committed research evidence are the retained records; private raw test logs are not promised as persistent evidence.
