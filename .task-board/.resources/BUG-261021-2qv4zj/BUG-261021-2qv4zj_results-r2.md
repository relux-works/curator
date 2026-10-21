# BUG-261021-2qv4zj: rework r2

Branch: `fix/audit-markdown-scope`.
Head: `9b17768c96d53e9ef7babb45e95d3b688728e815`.
Rework base: `2c01da83de4007c5d6f6d4edf3a887dc85707795`.
Original task base: `2c89fd46443d461e25bd377fb94a0c6754282838`.

Changelog: Fix Markdown audit bypasses through filesystem root case aliases by sharing context enumeration and invalidating ruleset-6 cached allowances.

## F1 correction

Implemented brief option (a). `whitelist.WalkContext` owns the canonical root lookups, traversal and existing `ContextPathIncluded` predicate. Both `CopyContext` and the Markdown detector consume this walk. Canonical root lookups can resolve differently cased physical entries on case-insensitive volumes; both consumers receive the same source file and destination-relative spelling. On case-sensitive volumes an unrelated root spelling does not become eligible. Descendant names, fixed exclusions and validated runtime/build exclusions retain the protocol's case-sensitive comparison. There is no second case-folded matcher and no global case-folding of protocol paths.

`CopyContext` retains its mandatory SKILL.md check and copying behavior. The audit scans the Markdown callbacks and separately retains the historical non-Markdown scripts/manifest walk. The finding classifier and body are unchanged. Ruleset 7 gives the corrected detector a fresh cache namespace.

The selected-package subject, capability authority and pinned CIP-0021 corpus remain unchanged. The prior references/notes.md lifecycle witness and declared-host control remain enforced.

## Files changed in this rework

| File | Change |
| --- | --- |
| `internal/whitelist/whitelist.go` | Extract the existing copying traversal into shared `WalkContext`; copying delegates to it. |
| `internal/audit/audit.go` | Audit Markdown through that traversal; preserve the non-Markdown surface; bump ruleset 6 to 7. |
| `internal/audit/markdown_alias_test.go` | Actual copy/gate alias witnesses, case-sensitive host controls, all fixed exclusion patterns with case variants, Markdown leaf variants, and real ruleset-6 Gate cache upgrade tests for both hash framings. |
| `internal/audit/markdown_cache_test.go` | Keep the historical pre-Markdown cache seed faithful after traversal extraction. |
| `internal/audit/network_contacts_mutants_test.go` | Reanchor prior mutants; restore the old physical-entry matcher and reuse ruleset 6 as additional bounded controls. |
| `internal/envprofile/markdown_scope_test.go` | Build and invoke the actual CLI with explicit strict/high/null audit; cover nine alias refusals and canonical/noise/declared controls. |
| `internal/envprofile/skill_directory_mutants_test.go` | Reanchor the ignored-contact mutant to the shared Markdown callback. |
| `docs/profile-skill-directories.md` | Explain filesystem-aware shared enumeration and ruleset-7 cache invalidation. |

## Tests and claims

| Test | Specific claim |
| --- | --- |
| `TestMarkdownScopeCaseAliases` | On the real case-insensitive volume, SKILL.MD, References, AGENTS, ASSETS, .SKILL_TRIGGERS, Templates, Examples, Data and Scripts all supply actual copied URL bytes, and both public gates refuse them. Persisted refusals cache; declared-host versions admit. On the case-sensitive image, directory aliases are absent from copied context and audit findings; missing canonical SKILL.md makes copying fail. |
| `TestMarkdownScopeExclusionCaseParity` | Each of the 19 fixed exclusion patterns, applied at depth as an ancestor, excludes its canonical example. A specified neighboring case variant remains copied and is network-blocked. README has both readme.md and ReadMe.MD variants. Expectations are explicit, not computed by the production matcher. |
| `TestMarkdownScopeExclusionCaseLeaves` | As actual Markdown leaves, README.md is excluded while readme.md and ReadMe.MD are copied and blocked. |
| `TestMarkdownScopeInvalidatesRulesetSixAllowance` | A subprocess restores the real ruleset-6 physical-entry enumeration, then uses production Gate and its cache writer to persist and reuse an actual allowance for References/guide.md. Current read-only and writing gates refuse those identical bytes in hash v1 and v2, read-only evaluation writes nothing, current refusals cache, and old records remain byte-identical. No hand-authored allowance substitutes for the historical gate. |
| `TestProfileInstallMarkdownCaseAliasesCLI` | A binary built from the checked-out source rejects all nine aliases with exit 1 and the network finding under strict/high/null audit. Neither source nor lock is published, and the native Claude skill context is absent. Canonical SKILL/references/assets contacts also refuse. Repository noise admits; declared canonical and alias references admit and publish readable URL bytes through the native Claude skill path. Alias-specific CLI cases skip on case-sensitive hosts, where gate/copy controls cover the different required behavior. |
| Existing Markdown scope, profile install, cache and network tests | Research/board/build-vendor noise still admits; canonical context contacts refuse; declared hosts admit; scripts exports and non-Markdown scripts/manifest behavior remain covered. Prior ruleset-4 allowances and ruleset-5 refusals still miss. |
| `TestNetworkContactNarrowingMutants` | The restored ruleset-6 matcher fails the alias regression with the specific missing-refusal assertion; admitting every Markdown file fails the research regression; dropping references, altering scripts visibility and reusing rulesets 3/5/6 each fail their named behavioral assertion. Child exit 1 plus the expected assertion is mandatory; setup failure does not count. |
| `TestSkillDirectoryNarrowingMutants` | The adapted contact mutant still defeats the selected-package lifecycle witness and is detected, alongside all existing directory/capability controls. |

## Validation environment

Go 1.25.5, darwin/arm64; Rust 1.91.0; Node 26.12.0; Git 2.54.0. All Go test commands unset GIT_DIR and GIT_WORK_TREE. A private HOME, module/cache directories, isolated Git config and a bounded PATH keep tests off operator configuration and provider executables. Both TMPDIR and GOTMPDIR use the required short temporary path for the default suite. No XDG override is inherited. Path variables below are redactions of task-local paths, not omitted dependencies.

The default suite runs in a disposable, clean checkout with a normal `.git` directory at the exact head above. `$VALIDATION` denotes that checkout, `$SCRATCH` the task scratch, `$SPEC` the spec checkout and `$CASE_SENSITIVE_TMP` the APFS image mount. `$SHORT_TMP` denotes the prescribed short temporary directory. GOCACHE and GOMODCACHE are private and are deleted after handover.

### Filesystem runs

The default volume was proven case-insensitive by creating a lowercase name and successfully resolving its uppercase alias; all nine copy/gate and CLI witnesses execute there without alias-related skips.

A separate 256 MiB sparse image formatted as APFS (Case-sensitive) was mounted inside task scratch. The focused Markdown scope tests and CLI canonical/noise/declared controls passed there. All nine gate/copy host controls ran; the cache seed and ten alias-only CLI subcases skipped because the historical alias allowance cannot arise there. The image was detached successfully after the run.

Image setup commands: `diskutil listFilesystems` exit 0; `hdiutil create -size 128m -type SPARSE -fs APFSX ...` exit 1 (Invalid argument); bounded retry `hdiutil create -size 256m -type SPARSE -fs 'Case-sensitive APFS' ...` exit 0; `hdiutil attach -nobrowse -mountpoint "$CASE_SENSITIVE_TMP" ...` exit 0; `hdiutil detach <recorded-device>` exit 0. No mounted image remains.

### Corrected test-development failures

The first focused audit run exited 1 only because the cache-version mutant's nested seed read the on-disk ruleset-7 constant while its outer compiled overlay used 6, so no ruleset-6 record was present. The seed now replaces the source version through a unique version-pattern anchor; the required subsequent failure is the reused allowance's missing network refusal.

The first focused CLI run exited 1 only for a mistaken fixture assertion that .research must disappear from the native package link. Existing native linking exposes the selected package tree; the scope contract concerns agent-facing context eligibility. That extra assertion was removed. The required refusal checks remain: exact network error, no profile source/lock and absent native skill context. All nine alias refusal subcases had already passed in the first run and pass again.

## Open decisions

None. Independent review remains owned by the orchestrator. No board status change, PR or default-branch push was performed.

## Validation receipts

The full default suite verified all 85 packages: 80 passed and 5 had no test files. It produced 11,684 passing tests/subtests and 199 conditional skips, with no failed test or package events. All 434 CLI, 399 envprofile and 371 install top-level test/example/fuzz entries were observed exactly once against the independently enumerated inventory. Package/test selections were disjoint. No alias regression was skipped on the default volume.

The three large packages were split into bounded selections; the remaining 82 packages were covered by six groups. The CLI's GOROOT-lock lifetime was respected: install's GOROOT fault-injection test ran separately after all CLI groups finished. Each test invocation had an eight-minute timeout and an outer watchdog shorter than ten minutes. The validation checkout was clean at the handed-over head after the suite.

Supplemental checks used CI's exact spec pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. The complete audit/skillcheck/whitelist packages and 45 opt-in envprofile entries selected from conformance-consuming files passed with zero skips. CIP-0021's vendored lifecycle/identity/schema tests and narrowing controls also passed in the full suite. The vendored corpus was not modified.

Focused development and case-sensitive runs used the source later committed at the handed-over head. Build, vet, the full suite and pinned supplements ran in the exact-head validation checkout.

| Run | Exit | Seconds |
| --- | ---: | ---: |
| `build` | 0 | 3.44 |
| `vet` | 0 | 2.66 |
| `focused-audit` | 1 | 21.88 |
| `focused-audit-r2` | 0 | 28.03 |
| `focused-cli` | 1 | 15.18 |
| `focused-cli-r2` | 0 | 60.60 |
| `sensitive-focused` | 0 | 7.95 |
| `pinned-skillcheck` | 0 | 29.78 |
| `pinned-env-1` | 0 | 154.61 |
| `pinned-env-2` | 0 | 37.33 |
| `full-cli-01` | 0 | 34.38 |
| `full-cli-02` | 0 | 51.48 |
| `full-cli-03` | 0 | 145.34 |
| `full-cli-04` | 0 | 74.96 |
| `full-cli-05` | 0 | 11.31 |
| `full-cli-06` | 0 | 70.23 |
| `full-cli-07` | 0 | 55.76 |
| `full-cli-08` | 0 | 84.65 |
| `full-cli-09` | 0 | 208.52 |
| `full-cli-10` | 0 | 26.82 |
| `full-cli-11` | 0 | 144.01 |
| `full-cli-12` | 0 | 126.56 |
| `full-env-01` | 0 | 73.64 |
| `full-env-02` | 0 | 28.16 |
| `full-env-03` | 0 | 25.76 |
| `full-env-04` | 0 | 92.73 |
| `full-env-05` | 0 | 105.92 |
| `full-env-06` | 0 | 59.40 |
| `full-env-07` | 0 | 90.99 |
| `full-env-08` | 0 | 59.84 |
| `full-env-09` | 0 | 47.94 |
| `full-env-10` | 0 | 142.23 |
| `full-env-11` | 0 | 36.22 |
| `full-env-12` | 0 | 7.38 |
| `full-env-13` | 0 | 92.84 |
| `full-install-01` | 0 | 40.15 |
| `full-install-02` | 0 | 89.64 |
| `full-install-03` | 0 | 106.14 |
| `full-install-04` | 0 | 46.86 |
| `full-install-05` | 0 | 180.96 |
| `full-install-06` | 0 | 83.49 |
| `full-install-07` | 0 | 7.66 |
| `full-install-08` | 0 | 0.89 |
| `full-other-01` | 0 | 94.31 |
| `full-other-02` | 0 | 126.17 |
| `full-other-03` | 0 | 76.75 |
| `full-other-04` | 0 | 103.08 |
| `full-other-05` | 0 | 31.92 |
| `full-other-06` | 0 | 43.75 |

The two development exits 1 above are explained in the corrected-test section; subsequent focused checks and the complete final-head suite passed. Expected-red mutant subprocess exits are not outer-suite failures.

### Commands

Commands below preserve their actual arguments with task-local path values redacted. `build`, `vet`, `full-*` and `pinned-*` run from `$VALIDATION`. Other checks run from the source checkout. All test commands inherit the isolated environment described above; the case-sensitive run overrides only test TMPDIR, leaving compiler GOTMPDIR short.

`build` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go build ./...
```

`vet` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go vet ./...
```

`focused-audit` (exit 1):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -v -run 'TestMarkdown|TestReviewPriorMarkdownAllowance|TestNetworkContactNarrowingMutants' ./internal/audit
```

`focused-audit-r2` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -v ./internal/audit ./internal/whitelist ./internal/skillcheck
```

`focused-cli` (exit 1):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -v -run '^TestProfileInstallMarkdown' ./internal/envprofile
```

`focused-cli-r2` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -v -run '^TestProfileInstallMarkdown|^TestSkillDirectory' ./internal/envprofile
```

`sensitive-focused` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE 'TMPDIR=$CASE_SENSITIVE_TMP' go test -count=1 -timeout=8m -v -run '^TestMarkdownScope|^TestProfileInstallMarkdown' ./internal/audit ./internal/envprofile ./internal/whitelist
```

`pinned-skillcheck` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE 'CURATOR_CONFORMANCE_ROOT=$SPEC/conformance/v1' go test -count=1 -timeout=8m -v ./internal/skillcheck ./internal/whitelist ./internal/audit
```

`pinned-env-1` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE 'CURATOR_CONFORMANCE_ROOT=$SPEC/conformance/v1' go test -count=1 -timeout=8m -v -run '^(TestClaudeSeedMergePreservesToolState|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestCodexSeedStripsInlineMCPTable|TestDotfileManagerVectors|TestEnvironmentWriteNofollowVectors|TestEnvironmentsCodexSeedVectors|TestEnvironmentsEnvPassthroughVectors|TestEnvironmentsReadFailureVectors|TestMuseAuthParentBoundaryAtResolve|TestMuseAuthRepairAdditionalStates|TestMuseDirectoriesAndIsolation|TestMuseForkRefusedRegression|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMusePublishedLinkStates|TestMusePublishedRegistryLayout|TestMuseStatusOptionalProvisioning|TestOpencodeFormFallback|TestOpencodeXDGSeeds|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathKindMCPVectorsDriveInstallEntry)$' ./internal/envprofile
```

`pinned-env-2` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE 'CURATOR_CONFORMANCE_ROOT=$SPEC/conformance/v1' go test -count=1 -timeout=8m -v -run '^(TestProfilePathCollision|TestRepairFailedMissingStore|TestReservedSkillName|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestSeedUnreadableStopsBeforeFirstWrite|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusProfileMembersAndUnregistered|TestStatusShadowAcknowledgment|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestStatusSignerPostureReadFailureIsNonCurrent|TestStatusSignerPostureUnknownWithoutCachedSource|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryStatusVectorsDriveStatus|TestSystemDeltaVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestUpdateConfirmationPostureVectorAtStatus)$' ./internal/envprofile
```

`full-cli-01` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestAuditAllowPinsOnlySupportedContentIdentity|TestAuditAllowRefusalPrecedesConfigLoad|TestCLIAuditBindsDependencyDirectory|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestReviewCLIStatusDetectsChangedTagAtSameCommit|TestAuditAllowRecordsPinCreationTime|TestCLIAuditEmitsScriptLabels|TestCLIAuditMixedSkillRecordsBothPolicies|TestCLIAuditJSONCarriesScriptLabels|TestCLIAuditJSONRecordsExplicitAbsence|TestClassifyBuildCommandMapsEveryCacheOutcomeToADistinctCode|TestUnusableToolchainRowsCarryTheDriverBoundaryCode|TestClassifyBuildCommandDetectsRecordedIdentityDrift|TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves|TestInputCausesAreDistinctAndDocumented|TestClassifySkillBuildsAcceptsOnlyAnExactlyCurrentInstallation|TestClassifySkillBuildsWithoutCompiledStateStaysSilent|TestClassifySkillBuildsDetectsContextExposure|TestClassifySkillBuildsDetectsCommandDriftInBothDirections|TestClassifySkillBuildsRefusesAMarkerThatCannotDescribeABuild|TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestCurrentnessCodesAreDistinctAndOnlyExactStatesPass|TestEveryCurrentnessCodeIsDocumented|TestCheckFailsForEveryNonCurrentCode|TestBuildReportsNeverPublishAnAbsolutePath|TestBuildReportDetailIsBounded|TestGoToolchainGuidanceNamesTheAcceptedSelectionAndTestedFamilies|TestRepairNoticesDistinguishRepairFromAPreservedInstallation|TestMarkerDigestsDetectAConcurrentInstallMarkerChange|TestStatusReportMarksCompiledStateThatMovedDuringTheCheck|TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill|TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema|TestStatusReportFindsASchema8InstallationCurrent|TestCommandHelpNarrowingMutants|TestRegisteredSubcommandHelpNoWrites|TestCommandHelpInitPositiveControl|TestCommandHelpRunPassthrough|TestCommandHelpUnknownTopic)$' ./cmd/curator
```

`full-cli-02` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestWithDraftRemediationTable|TestDraftAttemptTransportTable|TestDraftAttemptClauseShape|TestEndpointExhaustionCarriesRemediation|TestDraftProjectResolveHelp|TestDraftInstallStatusHelpIncludesSchema2Workflow|TestDraftDocumentedLocalShapesThroughCLI|TestDraftRemediationThroughCLI|TestDraftMachinePolicySetupThroughCLI|TestDraftDocumentedGitRevisionThroughCLI|TestDraftFetchFailureSanitizedThroughCLI|TestDraftAvailabilityFallbackThroughCLI|TestDraftStripCloneFramingTable|TestDraftLaunchConsumesFrozenRuntimeThroughCLI|TestDraftRepairRestoresDriftedContentThroughCLI|TestDraftDocsPinExamples|TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveSchema1KeepsReadOnlyMeaning|TestProjectResolveIgnoresRemovedSwitch|TestProjectResolveGitSelectionThroughCLI|TestProjectResolveGitRuntimeTamperRefusedThroughCLI|TestProjectResolveGitMissingMemberRefusedThroughCLI|TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI|TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI|TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveLegacyMixedThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveCanonicalRepositorySourceThroughCLI|TestProjectResolveGitCollectionPinnedToTagThroughCLI|TestProjectRefreshGitBranchMembershipChangeThroughCLI|TestProjectResolveDraftAllowlistLegacyThroughCLI|TestProjectResolveDraftAllowlistTransitiveThroughCLI|TestProjectResolveGitRealInstallThroughCLI|TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI|TestProjectRefreshFetchFailurePreservesStateThroughCLI|TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI)$' ./cmd/curator
```

`full-cli-03` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestProjectResolveOriginLessSkipsFetchThroughCLI|TestProjectRefreshAliasFetchDedupedThroughCLI|TestDraftStatusReportsUpToDateThroughCLI|TestDraftStatusNeedsInstallAfterRefreshWithoutInstall|TestDraftStatusLockOnlyChangeNeedsInstall|TestDraftStatusLegacyMarkerNeedsInstall|TestDraftStatusStaleManifestRefuses|TestDraftStatusStaleEmptyLockRefuses|TestDraftStatusStaleSwappedSelectionRefuses|TestDraftStatusContentDrift|TestDraftStatusNotInstalled|TestDraftStatusLiveMutationWithoutRefreshStaysCurrent|TestDraftStatusInvalidMarker|TestDraftStatusAppliesKeepsSchema1ProjectsUntouched|TestDraftStatusDriftFailsClosedWithoutFrozenInputs|TestDraftStatusDriftFailsClosedOnStaleLock|TestClassifyDraftMemberComparesExactCurrentness|TestClassifyDraftMemberPresenceRows|TestProductionExternalDepsAssignsTransportProvenanceSink|TestProductionExternalDepsFalseDrivesMirrorFetchToSink|TestDraftTransportLegacyGolden|TestDraftTransportResolvedMatrix|TestDraftTransportProviderAdmission|TestDraftTransportIdentityInvalidSurfacesThroughCLI|TestReviewProbeLegacyLaneIdentityInvalid|TestProductionBinaryDispatchesSSHWrapper|TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome|TestCredentialStoreCLIConfigRoundTrip|TestCredentialStoreCLIConfigRefusesMalformed|TestCredentialStoreCLIPromotion|TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery|TestReviewDottedCredentialProfile|TestReviewDottedCredentialProfileExisting|TestCredentialStoreCLIDottedProfileRouting|TestCredentialStoreCLIInvalidProfileRouting|TestEnvMigratePlanApplyPi)$' ./cmd/curator
```

`full-cli-04` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigrateUsage|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite|TestEnvResolveRepairFailedOnUninspectable|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases|TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestEnvResolveBareStaleIsAFailure|TestEnvResolveRepairEmitsFragment|TestEnvResolveRejectsSwappedStoreEntry|TestEnvResolveUnknowns|TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration|TestEnvStatusMatrix|TestEnvStatusRegistryBoundaryPostureAndCheck|TestEnvStatusReportsDroppedSystemModuleThroughCLI|TestUmbrellaMissingProvider|TestUmbrellaDispatchesToProvider|TestUmbrellaPropagatesExitCode|TestUmbrellaInvalidNameIsUsage|TestImplementedCommandWinsOverProvider|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList|TestEnvResolvePreserveLockRequiresRepair|TestEnvResolvePreserveLockRepairBindsUnchangedLock|TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates|TestEnvUnmanageTakeoverPreservesFileMode|TestEnvUnmanageTakeoverRestoresSymlink|TestEnvUnmanageSymlinkRestoreRefusesLinkedParent|TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent|TestEnvUnmanageBackupRecordVectors|TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord|TestEnvUnmanageRestoresNewestBackupGeneration|TestEnvUnmanageRestoresRecordedTakeoverBackup|TestEnvUnmanageLeavesBackupsWithoutRestoreFlag|TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases)$' ./cmd/curator
```

`full-cli-05` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestEnvConfigAliasOutputPrintsCanonicalKnob|TestEnvConfigAliasLockRefusalPrintsCanonicalKnob|TestRunDispatchNormalizesAliasOperand|TestNormalizeRunEnvOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions|TestEnvConfigShowPermissionsAtProductionEntry|TestEnvConfigShowPermissionsRejectsUnreadableConfig|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent|TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor|TestEnvConfigShowPermissionsWithoutProfiles|TestEnvConfigShowDefaults|TestEnvConfigSetUnsetRoundTrip|TestEnvConfigSetListKnob|TestEnvConfigRefusals|TestEnvConfigProviderDirectoriesRoundTrip|TestEnvConfigProviderDirectoriesLockedRefuses|TestEnvConfigLockedKnobRefuses|TestProfileComposeAddListRemove|TestProfileComposeAddWarnsWhenOverlaysForbidden|TestProfileComposeRefusals|TestProfileUseLockedRequireRefuses|TestProfileInstallUseLockedRequireRefuses|TestProfileInstallFirstActivationLockedRequireRefuses|TestProfileUseClearOperandIsUsage|TestProfileScopedUseUnaffectedByLockedRequire|TestProfileImportUseLockedRequireRefuses|TestProfileUpdateResyncPassesLockedRequire|TestEnvStatusReportsLockedRequire|TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsNotInheritedNativeCodexServers|TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable|TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$' ./cmd/curator
```

`full-cli-06` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree|TestGCPreservesRuntimeOnUncertainMarks|TestGCRetainsLiveProcessBuild|TestGCPrunesDeadConsumersUnderTheHomeLock|TestGCFailsClosedForUntrustedCurrentPathSource|TestGCWaitsForTheHomeLock|TestGCRunsSerializedAcrossConcurrentInvocations|TestSignerRowsSurviveHostileGlobalGitConfig|TestGlobalAdoptCLIAdoptsAndIsIdempotent|TestGlobalAdoptCLIRefusesByteMismatch|TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites|TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers|TestGlobalStatusRejectsSchema2SkillfileThroughCLI|TestReviewGlobalStatusDetectsChangedTagAtSameCommit|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestGlobalStatusReportsATransitivelyResolvedCompiledCommand|TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved|TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven|TestGlobalStatusRejectsPositionalArguments|TestStatusRecordedButMissingStaysInInventory|TestStatusUnreadableCandidatesKeepRecord|TestStatusUnreadableApprovalStateSurfaced|TestEnvStatusMissingAndUnreadableKeepRecord|TestEnvStatusUnreadableApprovalStateSurfaced|TestHookApprovalsUnreadableStateNamesReadFailure|TestHookApproveRecordsOperatorApproval|TestHookApproveFailsDistinctlyOnAbsentAndUnreadable|TestHookApproveReRecordsAfterChange|TestHookApproveIsIdempotentWhenRecordMatches|TestHookApproveResolvesSymlinkedOperand|TestHookApprovalsListsReadOnly)$' ./cmd/curator
```

`full-cli-07` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestHookApprovalsToleratesMalformedRecord|TestHookApprovalsOnAbsentStateIsEmpty|TestHookRevokeRemovesRecord|TestHookRevokeAbsentLeavesStateByteIdentical|TestHookUsageErrors|TestHookDiagnosticsAgreeWithShell|TestStatusReportsShellHookTrustPosture|TestStatusJSONCarriesShellHookTrust|TestStatusReportsRecordedPathsBeyondTheCurrentProject|TestEnvStatusReportsShellHookTrustPosture|TestHookApproveEndToEndWithGeneratedHook|TestProfileImportRow|TestProfileImportLossyRow|TestProfileImportNameTakenRow|TestProfileComposeAddPathRow|TestProfileComposeAddRefusesAFormOnAPathSource|TestPathOverlayFromMachineConfigJoinsTheClosure|TestOverlayFromMachineConfigIsRefusedByKind|TestProfileUseTakeoverRow|TestProfileInstallReinstallHonoursUseAndTakeover|TestResolveTakeoverRequiresRepair|TestSyncTakeoverFlagParses|TestInstallGitignoreEntryPoint|TestLegacyPackageStatusBindingAndRepair|TestAuthoritativeLifecycleBootstrapCasesAreExecutable|TestAuthoritativeLifecycleUpgradeCasesAreExecutable|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestRunUsesInjectedConfigSourceAndWriters|TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestProductionExternalDepsBindTrustedGitAndAudit|TestProductionExternalAuditReturnsAdvisoryWarnings|TestUsageEnumeratesDocumentedCommands|TestRunVersionExitsZero|TestRunNoArgsPrintsUsage|TestRunUnknownCommand|TestShellInitPrintsHooks|TestShellInitInstallCachesHookWithoutConfig|TestSkillCheckOnTempDir)$' ./cmd/curator
```

`full-cli-08` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestInstallFlagsAcceptTrailingOptions|TestInstallAuditFlagAcceptsOptionalMode|TestInstallFlagsCarryTheRunWideBuildSSHSelection|TestSelectProjectTargetsUsesAliasesAndStableAllOrder|TestStatusDriftDetectsContentTampering|TestBootstrapAndProjectCommands|TestBootstrapIfMissingKeepsExistingConfigWithoutParsing|TestBootstrapIfMissingCreatesAbsentConfig|TestBootstrapIfMissingRejectsForce|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot|TestCLIEndToEndInstallStatusAndTamperCheck|TestHiddenWorkerModeIsNotAUserVisibleCommand|TestProductionBinaryDispatchesNetNSProbe|TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestConfigBuildSSHAddListRemove|TestConfigBuildSSHListWithoutScopesPrintsNothing|TestConfigBuildSSHAddRejectsInvalidInvocations|TestConfigBuildSSHAgentFlagKeepsScopePositional|TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSAddListRemove|TestConfigBuildHTTPSListWithoutScopesPrintsNothing|TestConfigBuildHTTPSAddRejectsInvalidInvocations|TestConfigBuildHTTPSRemoveRejectsMissingAndMalformedTargets|TestConfigBuildHTTPSLoginRejectsInvalidInvocations|TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt|TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential|TestConfigHelpDocumentsPrecedenceAndSubcommands|TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure|TestConfigSubcommandDispatch|TestTheCredentialPromptIsWiredOnlyWhereAnOperatorCanAnswerIt|TestThePromptPersistsThroughTheOrdinaryConfigWriter|TestSSHThisRunOnlyPromptNeverReachesTheSavedConfig|TestHTTPSThisRunOnlyPromptNeverReachesConfigOrCredentialStore|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextCopyFallback|TestMuseRootContextVectors|TestMuseRootContextDrift|TestMuseRootContextRefusals)$' ./cmd/curator
```

`full-cli-09` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestMuseCredentialPolicyStatus|TestMuseRootContextStatus|TestMuseRootContextNarrowingMutants|TestMuseCLIFragmentAndFakeLaunch|TestMuseProfileSeedSecretAndCredentialRefusals|TestMuseFragmentV3PublishedCases|TestMuseCLIAuthParentBoundary|TestMuseFragmentV4PublishedCases|TestMuseCLIStorePolicy|TestNativeBlackboxExternalBuildLifecycle|TestNativeSkillsTakeoverEveryRegisteredCommand|TestNativeSkillsTakeoverArgumentDiscardRegression|TestNativeSkillsTakeoverActivationBoundary|TestReviewNativeTakeoverHelpBoundary|TestNativeSkillsTakeoverHelpCrossProduct|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsCLIOperations|TestNativeSkillsTakeoverFlagBoundary|TestProductionExternalAuditOpaqueNULFollowsHashVersion|TestCLIAuditOpaqueNULFollowsHashVersion|TestAuditAllowWritesWriterVersionPinCarrier|TestStatusCheckRefusesV1NULAppearingAfterInstall|TestStatusCheckAdmitsV2NULBearingInstall|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileGitReinstallHonoursUseAndTakeover|TestProfileInstallActivationMatrix|TestProfileInstallStopThenRetryConformanceVector|TestProfileInstallReinstallUsePreservesProjectScope|TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication)$' ./cmd/curator
```

`full-cli-10` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestGitFileURLHasNoBackslash|TestProfileInstallListUse|TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileUseUnknownIsAFailure|TestProfileRemoveInUseIsAFailure|TestProfileInstallRefConflictIsAFailure|TestProfileSyncCoversTheMachineScope|TestProfileComposeListsDeclaredOverlays|TestProfileUseUnknownEnvironmentIsAFailure|TestProfileUseTargetIsStageB|TestProfileUseTargetBoundToAdapter|TestProfileInstallWarnsOnSystemModule|TestProfileScopedUseAndClear|TestProfileUpdatePinnedTagIsUnchanged|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallFileOperandIsRefused|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallUseActivatesThroughSwitch|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording|TestSecurityPostureVectorsThroughCLI|TestHardenedRegistryOutageRefusesWithAdvisoryRegistryPolicy|TestPermissiveRegistryOutageWarnsWithArtifactIdentity|TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningIsNotEmittedByEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestSecurityPostureCuratorStatusRows|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestSecurityPostureProfileInstallRefusesMCPDeclarations|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestSecurityPostureRevisionBDefaultRefusesEmptySources|TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist|TestSecurityPostureRevisionBDefaultEnvStatus|TestCLIHelperProcess)$' ./cmd/curator
```

`full-cli-11` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestCompiledProjectRepairsCorruptCompiledState|TestCompiledProjectRestoresCacheWhenCommitFails|TestStatusReportsATransitivelyResolvedCompiledCommand|TestStatusReportsAnUnusableToolchainPerCompiledCommand|TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands|TestGCRetainsAndReportsReferencedCompiledState|TestCompiledInstallFollowsTheNativeControlInventoryExactly|TestDryRunNeverClaimsACompletedCompilerCheck|TestStatusExplainsAnUnusableGoToolchain|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestCuratorStatusProviderPostureAndCheck|TestStatusAcceptsAnUnchangedLegacyMarkerSchema|TestHostGoToolchainIsSelectableOnAnInventoryPlatform|TestInstallPrintsTheRemedyAVersionManagerSelectionEarns|TestUmbrellaProviderResolutionVectors|TestActiveRevisionIsWarningRelease|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd|TestUmbrellaHostilePathPlantRefusedUnderB|TestUmbrellaTrustedRootSymlinkEscape|TestUmbrellaTrustRootSelectionBothRevisions|TestUmbrellaRefusedDirectoriesBothRevisions|TestUmbrellaSymlinkIntoManagedRefused|TestUmbrellaUnreadableRootNeverAbsence|TestUmbrellaMissingNamesRoots|TestUmbrellaWarningText|TestProviderPostureRows|TestProviderSubcommandParsing|TestExecutableCandidates|TestProviderInputsForHostScannedBinWarnsWithoutLedger|TestProviderInputsForHostPublishedBinRefuses|TestProviderPostureReportsUnreadablePublishedShimLedger|TestProviderInputsForHostExplicitBinRefuses|TestUmbrellaTrustRootCaseVariantIdentity)$' ./cmd/curator
```

`full-cli-12` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestCompiledProjectStatusAndUntrustedRecovery)$' ./cmd/curator
```

`full-env-01` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestInstallErrorRefusesTransitiveSystemModule|TestInstallErrorWaiverAdmits|TestUpdateErrorLeavesLockUnchanged|TestDropRepairWarnsAndFragmentFollowsAdmitted|TestDropAllDroppedOmitsFragmentSection|TestStatusReportsPolicyAndDropped|TestStatusErrorReportsPolicyWithoutDropped|TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops|TestPolicyFromConfigCarriesAdmission|TestPathInstallAdmitsOrdinaryDirectory|TestLegacyPathInstallIgnoresStoreOverlap|TestPathOverlayAdmitsOrdinaryDirectory|TestStrictAuditMemberCapabilityCacheIsolation|TestClaudeSeedNarrowingMutants|TestClaudeRepairPreservesNumberText|TestClaudeRepairSatisfiedIsNoOp|TestClaudeAbsentSeedGolden|TestClaudeRepairRefusesUnsafeState|TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression|TestCodexSeedStripsInlineMCPTable|TestCodexSeedRejectsInvalidTOMLBeforePublishingHome|TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair|TestEnvironmentsCodexSeedVectors|TestSharedToIsolatedRemovesStaleLink|TestCredentialLinkRegularFileRefuses|TestStaleCredentialLinkRefusals|TestDanglingPiLinkReportedDetached|TestCodexIsolatedAdmission|TestCodexUnknownStoreSharedRefuses|TestPiProvisionTargetsAgentRoot|TestCredentialLinkDirectory|TestCredentialLinkUnrecordedSymlinkRefuses|TestReviewerDanglingExpectedTarget|TestReviewerLiteralCodexSelector|TestDanglingExpectedTargetRepairSucceeds)$' ./internal/envprofile
```

`full-env-02` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestCredentialLinkTargetInspectionFailure|TestCodexCredentialStoreTOMLSpellings|TestCodexMalformedStoreSharedRefuses|TestPiProvisionTargetsAgentRootV2WriterMode|TestCredentialStorePolicyActivation|TestCredentialStorePolicyNativeInspection|TestCredentialLinkInspectionRepairFails|TestMigrateApplyLockContention|TestCredentialProjectionSingleObservation|TestCredentialProjectionIsolationRefusals|TestCredentialProjectionStatusObservation|TestCredentialProjectionIdentitySubrequest|TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestResolveRecoversCredentialMarkerAfterRenameCrashV2WriterMode|TestUpdateSystemDeltaRefusesBeforePublication|TestEnvironmentsEnvPassthroughVectors|TestMigrationHonoursSystemLockedAllowlist|TestMigrationHonoursSystemLockedRevocation|TestMigrationSucceedsWhenSystemPolicyPermits|TestMigrationHonoursConfigPathOverride|TestMachinePolicyLoadFailureFailsMigration|TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestSourceAllowlistRejectsAdjacentPathSegment|TestRevokedSkillMemberIsRefused|TestRevokedMCPMemberIsRefused|TestStrictAuditCanaryFailureBlocksInstall|TestCanonicalGitRejectsFileRemote|TestGateSourceRejectsFileRemote|TestFileOperandIsRefused|TestFileRequirementSourceIsRefused|TestInstallUseSwitchesAndAgrees|TestInstallUsePartialLeavesCurrent)$' ./internal/envprofile
```

`full-env-03` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestOperandShadowedByDirectoryResolvesAsGit|TestAbsentPathOperandWithRequirementIsRefConflict|TestInstallOperandKindIsSyntactic|TestInstallNeverDemotesANetworkIdentityToAPath|TestInstallRefusesANonSourceOperand|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce|TestScopedRecordFilenameIsPortable|TestCanonicalIdentityUnifiesSpellings|TestCanonicalIdentityRejectsMalformedSource|TestSourceAllowlistRefusesBeforeClone|TestRevokedSourceIsRefused|TestMCPAllowlistRefusesOutsidePackage|TestStrictAuditCanaryPasses|TestMigratedSkillSourceIsCanonical|TestPolicyFromConfigCarriesEnvGates|TestPolicyFromConfigCarriesStateWaiverHashVersion|TestMachineClearReachesSeamGate|TestInstallPathAndList|TestInstallRejectsSecretMember|TestReinstallSameSourceIsAnUpdate|TestNameTaken|TestUseMaterializesLinkedHomes|TestUseReplacesWithBackup|TestUsePartialRecordsNothing|TestRemoveRefusesCurrent|TestRemovePurgeCleansHomes|TestUpdatePathLeavesLockInPlace|TestInstallGitResolvesNetworkIdentity|TestInstallSurfacesSystemModuleWarning|TestInstallSurfacesUnresolvedMCPCommand|TestUpdateDefaultIsBlocked)$' ./internal/envprofile
```

`full-env-04` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestScopedUseAndClear|TestListFailsOnCorruptRecord|TestInstallDirectoryRejectsSecretUnderSubdirectory|TestInstallTransitiveDirectoryRejectsSecret|TestEnsureDefaultCreatesLocalProfile|TestDefaultProfileMaterializesOnFreshHome|TestEnsureDefaultMigratesGlobalSkills|TestMutationLockContention|TestIncompleteJournalRecoversOnEntry|TestOperationsLeaveNoJournalResidue|TestGlobalSkillCapabilitiesNarrowingMutants|TestMigratedGlobalSkillNetworkCapabilities|TestMigratedGlobalSkillIgnoresPreManifestVerdictCache|TestMigratedGlobalSkillDeclarationKeepsOtherGates|TestMigratedGlobalSkillManifestErrorsBlock|TestInstalledMemberCapabilitiesStayKindScoped|TestMigratedGlobalSkillBinaryWriteVersion|TestMigratedGlobalSkillNetworkUndeclaredFollowUp|TestMigratedGlobalSkillPromptScopeValidation|TestMigratedGlobalSkillNullCapabilityDeclarations|TestImportLosslessInstallsThroughPathPipeline|TestImportNormalizesOrdersAndPins|TestNormalizeImportText|TestImportLossyStopsWithoutConsent|TestImportLossyProceedsWithConsent|TestImportUnreadableRootIsLoss|TestImportInvalidUTF8IsLoss|TestImportPartialActivationKeepsLock|TestImportNameTakenBeforeAnyWrite|TestImportSkillForeignPinnedByRevision|TestImportSkillMarkerRecovery|TestImportRecordsImportedFromNative|TestImportSkipsManagedRootContext|TestImportDivergentSkillsAreLoss|TestImportLedgerFailureIsLoss)$' ./internal/envprofile
```

`full-env-05` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestImportAbsentMarkerDetectsNative|TestImportCorruptMarkerIsLoss|TestImportUnreadableMarkerIsLoss|TestDotfileManagerVectors|TestUnreadableDotfileStateKeepsTakeoverQuiet|TestDotfileManagerTableMatchesSpecFixture|TestDotfileStatePathResolution|TestForeignManagerHintLstatDiscipline|TestDotfileManagerStateAloneDoesNotWarn|TestManagedModesNarrowingMutants|TestManagedFallbackPermissions|TestManagedSeedPreservesNativePermissions|TestManagedHomeRepairTightensDirectories|TestManagedAuthoredFilesArePrivate|TestManagedSeedRefusalsPreserved|TestResolveVerifiesGitStoreEntryAgainstPinnedTree|TestResolveRejectsGitPinMissingFromLocalObjectDatabase|TestResolveProvisionRepair|TestResolveReportsRecordedContextSurfaceHashDrift|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestResolveStaleUnprovisioned|TestResolveUnknowns|TestResolveDefaultProfile|TestResolveDriftRepair|TestResolvePassthroughLiveness|TestClaudeRootAlwaysCopied|TestResolveClaudeProjectEntry|TestCodexKeyringAmbient|TestCodexUnreadableConfigFails|TestRepairNeverRefreshesSeeds|TestResolveIsolatedOpencode|TestResolveFormUnsupported|TestResolveLockContention|TestResolveFormats|TestProfileInstallMarkdownScope)$' ./internal/envprofile
```

`full-env-06` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestProfileInstallMarkdownCaseAliasesCLI|TestMaterializeCaptureCurrentAndRelease|TestMaterializeCaptureCredentialStoreAdmission|TestMaterializeCapturePendingDoesNotRecover|TestMaterializeCaptureBadRecoveryEvidence|TestMaterializeCaptureWaitReselectsAndTimeout|TestMaterializeCaptureSameFragmentDrift|TestMaterializeCaptureAndStatusNeverProbeVendor|TestMaterializeCaptureDoesNotProvision|TestMaterializeCaptureObjectBounds|TestMaterializeCaptureSerializesProfileWriters|TestMaterializeCaptureConcurrentRepair|TestMaterializeCaptureStoreObjectIdentity|TestMaterializeCaptureAuthorizingDrift|TestMaterializeCaptureUnreadableJournal|TestMaterializeCaptureExternalTornGeneration|TestMaterializeCaptureCredentialPolicyReleaseObservation|TestMaterializeCaptureCredentialPolicyReleaseCache|TestMigratePiWrongTargetToAgentRoot|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateNoSecretCopies|TestMigrateConflictsRefuse|TestMigratePiRootInspectionFailure|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigrateOldRootBytesWarnAndProceed|TestMigratePendingNeedsNothing|TestMigrateUnknownOperands|TestMigrateInspectIsReadOnly|TestMigrateApplyRequiresPlan|TestMigratePrintFailureRefusesBeforeMutation|TestMigrateMarkerDriftRefuses|TestMigrateFailedRelinkKeepsOldLink|TestMigrateSyscallFailureRollsBack|TestMigrateInterruptedApplyRecovers)$' ./internal/envprofile
```

`full-env-07` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestMigratePublishFailureRevertsLinks|TestMigrateSealedJournalDeletesWithoutTouching|TestMigrateBrokenJournalRefuses|TestMigrateRecoveryRefusesUnexpectedTarget|TestMigrateRecoveryCleansOwnedTemp|TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode|TestMuseCredentialPolicyMatrix|TestMuseMissingNativeFilePolicy|TestMusePolicySwitchPreservesCredentials|TestMusePolicyExplicitMigration|TestMusePublishedLinkStates|TestMuseForkRefusedRegression|TestMuseAuthParentBoundaryAtResolve|TestMuseDirectoriesAndIsolation|TestMuseAuthRepairAdditionalStates|TestMuseGlobalReconciliationUsesExplicitProvisioning|TestMuseStatusOptionalProvisioning|TestMusePublishedRegistryLayout|TestResolveRefusesMissingLockNamedStoreEntryAtBoundary|TestNativeSkillsNarrowingMutants|TestNativeSkillsUseAndSync|TestNativeSkillsTakeover|TestNativeSkillsLinkedParent|TestNativeSkillsLedger|TestNativeSkillsUnreadableMarker|TestNativeSkillsRepair|TestNativeSkillsConvergenceAndSwitch|TestNativeSkillsInstallAndUpdate|TestNativeSkillsBackupCollision|TestNativeSkillsUnrecordedLedger|TestNativeSkillsEmptyProfileLeavesUnnamedParent|TestGitFileURLForKeepsVolumesOutOfTheHost|TestServeWritesTheGitSafeRemote|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot)$' ./internal/envprofile
```

`full-env-08` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestUpdateBlocksDeepNULFileInNewContextMember|TestResolveRefusesV1CommitPinnedContextNUL|TestResolveCleanV1CommitContextStaysCurrent|TestResolveAdmitsV2CommitPinnedContextNUL|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL|TestStrictAuditMemberRefusesV1NULBeforeHashing|TestValidateNamedStorePinsRefusesV1NULBeforeHashing|TestSkillsOfRefusesV1NULBeforeHashing|TestGuardedReadersV1NULRefusalComputesNoV1Identity|TestInstallAdmitsV1CollidingSkillTreesUnderV2|TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2|TestPrecedencePrimitivesDriveEmission|TestPrecedenceTieKeepsTopologicalOrder|TestPathOverlayJoinsClosure|TestOverlayExplicitWeightOverridesDefault|TestGitOverlayJoinsClosure|TestPathOverlayFormIsSourceInvalid|TestOverlayInvalidSourceKindIsSourceInvalid|TestSCPOverlayResolvesAsGit|TestOverlayRepeatedNameIsCompositionInvalid|TestOverlayDuplicateNameIsCompositionInvalid|TestForbiddenOverlaysResolveAlone|TestRemoveRefusesOverlayOfAnotherProfile|TestWeightRulesApplyInOrder|TestWeightRulesAllFourDisagree|TestWeightDuplicateRefused|TestWeightUnknownRefused|TestNonRootWeightsMapRefused|TestWeightConflictErrorAndWarning|TestPathInstallCapturesDirtyUntrackedInsideGit|TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve|TestPathRootAndImportedMCPDeclarationsRefused|TestPathOverlayMCPDeclarationRefused)$' ./internal/envprofile
```

`full-env-09` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestResolveRejectsGroupWritableOnlyPathOverlay|TestMissingPathOperandIsPathMissing|TestMissingOverlayPathIsPathMissing|TestNonDirectoryPathIsSourceInvalid|TestNestedGitIsSourceInvalid|TestSymlinkInPathIsUntrusted|TestFoldedPathsAreSourceInvalid|TestRootGitExcluded|TestUnreadablePathIsUnreadable|TestPathSnapshotImmutableAcrossUpdateSyncUse|TestPathUpdateAllSucceedsWithImportedProfile|TestUnreadablePathRootLeadsUnreadable|TestUpdatePathWithoutStatePinIsSourceInvalid|TestUpdatePathMissingSnapshotIsSourceInvalid|TestUpdatePathSnapshotNameMismatchIsSourceInvalid|TestPathReinstallBareAndNonCurrentMovesPin|TestUpdateBlocksSecretOverlayMember|TestReinstallBlocksSecretOverlayMember|TestUpdateBlocksRevokedOverlayMember|TestUpdateBlocksV1RevokedOverlayMemberUnderV1Writers|TestReinstallBlocksRevokedOverlayMember|TestPermissionsForProfileProperties|TestRepairPreserveLockVectors|TestPreserveLockRefusesBeforeStoreRebuild|TestPreserveLockRaceRefusesBeforeJournalRecovery|TestPreserveLockMigrationRefusalPrecedesBusyLock|TestPreserveLockRevalidationHoldsRepairLock|TestPlainRepairMigrationKeepsLockPrivate|TestSystemDeltaVectors|TestRC14MigrationRehashesLegacyIdentities|TestRC14ResolveRepairMigratesSiblingHomes)$' ./internal/envprofile
```

`full-env-10` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestRC14IdentityMigrationRollsBackEveryEntry|TestRC14IdentityMigrationRefusesSiblingPlanDrift|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestRC14IdentityMigrationRecoversInterruptedCommit|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14IdentityMigrationPreservesFallbackCopies|TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestEnvironmentsReadFailureVectors|TestPassthroughReadFailuresAreUnknownAndNeverRepaired|TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestUnreadableLockRefusesUpdateWithoutRewriting|TestAbsentLockRemainsProfileUnknown|TestUnreadableCodexSeedDoesNotSelectAbsentDefault|TestReviewerRecoveryMarkerDrift|TestReviewerRecoveryPreservesRegularTemp|TestSkillDirectoryNarrowingMutants|TestSkillDirectoryLifecycleVectors|TestSkillDirectoryMarkdownDeclaredControl|TestSkillDirectoryMetadata|TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse|TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer|TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer)$' ./internal/envprofile
```

`full-env-11` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestUpdateConfirmationPostureVectorAtStatus|TestSignerPostureVectorsAtStatus|TestStatusOrphanRetention|TestStatusShadowAcknowledgment|TestStatusProfileMembersAndUnregistered|TestSeedUnreadableStopsBeforeFirstWrite|TestReservedSkillName|TestProfilePathCollision|TestOpencodeFormFallback|TestRepairFailedMissingStore|TestOpencodeXDGSeeds|TestClaudeSeedMergePreservesToolState|TestStatusShellHookTrustFromLaunchProject|TestStatusShellHookTrustMissingUnreadableAndStateFailures|TestResolveRepairKeepsEnclosingStoreRootFailureClass|TestStoreBoundaryResolveVectorsDriveResolve|TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan|TestStoreBoundaryRepairVectorsDriveResolveRepair|TestStoreBoundaryStatusVectorsDriveStatus|TestStoreBoundaryCheckedRootsExcludeCuratorHome|TestInstallEmitsSurfacingBeforePublication|TestUpdateEmitsSurfacingBeforePublication|TestReinstallEmitsSurfacingBeforePublication|TestSurfacingEmittedDespitePublicationFailure|TestSurfacingSinkWriteFailureIsNonFatal|TestInstallSurfacesMCPDeclarations|TestInstallSurfacingSilentWithoutMCP|TestUpdateSurfacesCandidateMCPSet|TestAllowlistWarningOperations|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent|TestResolvePassthroughKnobs|TestStatusS4Posture)$' ./internal/envprofile
```

`full-env-12` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestSurfacingUnreadableManifest|TestUseWithoutTakeoverRefusesUnmanaged|TestUseTakeoverBacksUpAndNotifies|TestForeignSymlinkStopsSwitch|TestTakeoverWarnsDotfileHeuristic|TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn|TestReadOnlyCommandsNeverOnboard|TestRepairTakeoverProvisionsManagedHome|TestUnmanageRestorePlanRetainsMetadata|TestUnmanageRestoreRefusesLinkedParent|TestUnmanageRestoreSymlinkReplacesDestinationEntry|TestEnvironmentWriteNofollowVectors|TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$' ./internal/envprofile
```

`full-env-13` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestProfileNetworkContactVectors)$' ./internal/envprofile
```

`full-install-01` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestProjectManifestABAAroundTheReadRestartsClosure|TestGlobalManifestABAAroundTheReadRestartsClosure|TestSubstitutionsABAAroundTheReadRestartsClosure|TestHybridActivationABAAroundTheReadRestartsClosure|TestByteIdenticalRewriteDuringTheReadDoesNotRestart|TestVerifiedBuildAuthorityCarriesOneBindingThroughCacheAndDispatch|TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint|TestVerifiedBuildAuthorityFailsBeforeCacheOrDispatchOnProviderDrift|TestVerifiedBuildAuthorityRechecksProviderBeforeCacheAndDispatch|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry|TestDeclaredOnlyWithHostsGlobalInstallsWithOnlyDeclaredOnly|TestDeclaredOnlyWithHostsInstallsWithOnlyDeclaredOnly|TestBuildHTTPSPrecedenceLongestScopeAndAnonymousFallback|TestBuildHTTPSDiscoveryOnlyListsAndNeverSelectsAHostCredential|TestBuildHTTPSHostPinMakesOtherHostsResolveWithoutTheOverride|TestCaptureBuildHTTPSSelectionFreezesAndRedactsSecrets|TestBuildHTTPSResolutionSkipsOtherEffectiveTransports|TestBuildHTTPSSelectedSourcesFailClosedWithExactRemedies|TestBuildHTTPSCredentialReadersMaterializeConfiguredSources|TestBuildHTTPSResolutionIsCarriedByTheExternalInstallPlan|TestPromptedBuildHTTPSAbortStopsTheProductionPlanBeforeAnyFetch|TestResolvedHTTPSCredentialIsBoundToOnlyItsEffectiveFetch|TestHTTPSPromptDefaultSelectsExistingCredentialAndNarrowestSavedScope|TestHTTPSPromptThisRunOnlyNeverReachesPersistenceOnEitherCredentialSurface|TestHTTPSPromptPersistsEnteredTokenOnlyAfterScopeChoice|TestHTTPSPromptAbortNeverPersistsOrSelects|TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes|TestConfiguredScopesAreSelectedPerRepository|TestRepositoriesThatNeedNoCredentialSkipSelection|TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport|TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode|TestDefaultBuildSSHScopeIsTheRepositoryNamespace|TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment|TestConfiguredPathsExpandTheOperatorHome|TestKnownHostsFallBackFromScopeToRunWideToOperator|TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections|TestConfiguredAgentResolvesToTheLiveSocket|TestRunWideAgentAutoRequiresALiveSocket|TestEveryConfiguredShapeReachesTheWrapperPolicy|TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly|TestDiscoveryDegradesWhenTheAgentCannotBeAsked|TestDiscoveryWithoutAnAgentOrAHomeFindsNothing|TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder|TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise|TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent|TestAddCommandsAreBuiltFromTheDiscoveredCandidates|TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates|TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath|TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource|TestProvenanceNamesEverySourceOnePerRepository|TestAResolverThatCoversNothingStillFailsClosed|TestAnAbortedResolverFailsTheRunWithItsOwnError|TestPromptedScopesDoNotMutateTheLoadedConfiguration|TestTheCredentialPrecheckRunsBeforeAnyFetch|TestADryRunReportsThePerRepositoryCredentialSource|TestTheCredentialReportIsLabelledByTheScopeThatProducedIt|TestPromptDefaultSelectionPinsTheAgentToTheFirstDiscoveredKey|TestPromptThisRunOnlyReturnsSSHCredentialWithoutPersisting|TestPromptNumberedSelectionAndAnExplicitScope)$' ./internal/install
```

`full-install-02` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestPromptManualIdentityPathCoversAKeyDiscoveryDidNotList|TestPromptRejectsAMalformedManualPathAndAsksAgain|TestPromptRejectsAnUnofferedChoiceAndAMalformedScope|TestPromptAbortFailsClosedWithoutPersistingAnything|TestAbortCarriesTheProtocolCodeSoTheRunStillFailsClosed|TestOneScopeChoiceCoversEverySiblingRepository|TestPromptWithoutCandidatesStillOffersTheManualPath|TestPromptListsTheUnlistedIdentityCount|TestAPersistFailureStopsTheRunRatherThanContinuing|TestAuthoritativeCacheOutcomesDriveInstallation|TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted|TestCacheWrongTargetRebuildCompletesWithinDeadline|TestRunCommitRecheckRefusesDestinationInsideAdmitted|TestRunCommitRecheckRefusesSwappedParent|TestRunCommitRecheckPassesUnchangedBoundaries|TestRunCommitWithoutBoundariesSkipsRecheck|TestCommitOrdersTargetClassesWithConsumerLast|TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess|TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched|TestAFailedCommitRestoresTheBuildCacheItReplaced|TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced|TestAnInFlightTransactionKeepsThePublishedCacheEntry|TestAPublicationThatChangedTheCacheIsReportedAsRetained|TestAReversalThatDidNotCompleteIsReportedAsRetained|TestBuildPublicationFailuresAreRedactedInTheResult|TestConcurrentProjectInstallsPreserveBothConsumers|TestInstallHelperProcess|TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets|TestRecoveryCompletesBeforeAnyNewMutation|TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan|TestMaintenanceFailureAfterCommitIsAWarning|TestNoSharedTargetChangesBeforeTheManagerHomeLock|TestGlobalCommitCarriesNoConsumerLedger|TestStagedPlanRejectsTwoProducersClaimingOneLivePath|TestAdapterLedgerCommitsAfterTheMirrorsItClaims|TestRedactDiagnosticReplacesEveryAbsoluteLocation|TestRedactDiagnosticCannotDriveATerminal|TestRedactDiagnosticIsBounded|TestRedactDiagnosticKeepsInvalidUTF8Out|TestPlanLinesRedactAnUntrustedReason|TestBuildFailuresAreRedactedInTheResult|TestGoToolchainRemedyReachesTheOperatorIntact|TestDraftFailureAtEveryTargetClassRestoresPriorState|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditAdvisoryLocalPasses|TestDraftAuditBindingLifecycle|TestDraftAuditHostileContentBlocked|TestDraftAuditPinDoesNotWaiveStrictFindings|TestDraftAuditRevocationBlocks|TestDraftAuditBrokenReportRefuses|TestDraftAuditFirstIssuanceVsBrokenRecord|TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditRenewalNeverHidesRefusal|TestDraftAuditRenewalMarkersNeverRenewable|TestDraftAuditMalformedShapeRefuses|TestDraftAuditClosedPackageShapeRefuses|TestDraftAuditEnforcedScriptAdmitted|TestDraftAuditDuplicateKeysRefuse|TestDraftAuditUnreadableBindingRefuses|TestExternalAdoptionsListEveryPublishedEntryOnce)$' ./internal/install
```

`full-install-03` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestAdoptExternalPublicationsWarnsWithoutCreatingAnything|TestDraftBuildsPublishReceipt3OnBothArms|TestDraftBuildKeyFollowsThePackageIdentity|TestDraftBuildRefusesASessionThatBindsTheLegacyDigest|TestDraftBuildToolchainFailureBehaviorIsPreserved|TestLegacyBuildsKeepReceipt1|TestDraftBuildFinalRootParentsAreStoreCreated|TestReviewExternalExecutionBindsExactReceipt3Input|TestDraftBuildRefusesCompilerViewDigest|TestDraftEvidenceExactMatch|TestDraftGitInstallWritesMarkerV5|TestDraftGitMarkerLockBindingRoundTrip|TestDraftGitRuntimeMaterializesUnderSourceV1Key|TestDraftGitInstallCarriesRegistryAttestation|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker|TestDraftLocalRejectsConfusedAttestation|TestDraftMarkerArmGuards|TestDraftSelectorSubstitutionForbidden|TestDraftGitBuildsPublishReceipt3|TestDraftDeclaredTagBumpIsNotAMovedTag|TestDraftSameTagMoveInstallsWithoutWarning|TestLegacyDeclaredTagBumpIsNotAMovedTag|TestDraftDryRunExposesResolvedAttestations|TestDraftRepairRevalidatesLockedSource|TestDraftRepairRestoresDriftedContent|TestDraftRefusedRepairPreservesLock|TestDraftInstallFailsStaleBindingsGeneration|TestDraftRefreshThenInstallReplacesLockAndMarker|TestDraftRetargetFailsAtWriteTimeAndRollsBack|TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack|TestDraftUnmanagedAdapterDestinationRefuses|TestDraftLocalRuntimeMaterializesFromFrozenSnapshot|TestDraftLocalRefreshReplacesFrozenRuntime|TestDraftLocalTamperedFrozenRuntimeRefuses|TestDraftLocalMissingSnapshotReplays|TestDraftLocalEnforcedCommandRefused|TestDraftLocalMissingSystemCommandRefused|TestDraftLocalSkillDependencyRefused|TestDraftLocalInvalidCapabilitiesRefuseAtResolve|TestDraftMarkerPackageDigestMatchesLockDigest|TestDraftRuntimeKeysCoverAllArms|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftInstallMissingLockFails|TestDraftInstallStaleLockFails|TestDraftInstallReplaysMissingLocalSnapshot|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache|TestDraftInstallUsesPinnedBytes|TestSchema2InstallRequiresLockAndLegacyStillWorks|TestDraftFreshMachineTransitiveReplaySourceGuards|TestDraftReplayFileURLUsesPortableGitConfigSyntax|TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade|TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings|TestDraftReplaySeedSurvivesIsolatedFetch|TestDraftFreshGitSourceWithoutEndpointIsUnavailableAndPreservesLock|TestDraftFreshMachineMovedTagReplaysLockedCommitWithoutRefResolution|TestDraftFreshGitReplayChecksContentHashBeforePublishingSnapshot|TestDraftFreshPathMismatchAndUnreachableSourcePreserveLock|TestDraftInstallDoesNotIgnoreOrPrivatizeSkillfileLock|TestDraftInstallGitPinnedSubtree)$' ./internal/install
```

`full-install-04` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestDraftInstallGitTamperRefused|TestAcquireDraftNetworkRevision2ProvenanceSink|TestDraftTransportEnabled|TestDraftTransportPlanConversion|TestDraftTransportAuth|TestDraftPlanNeedsSSH|TestDraftPlanNamesProvider|TestDraftSSHWrapperBase|TestDraftManagerExecutable|TestAcquireDraftNetworkSelection|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess|TestDefaultAcquireFetchesTheDeclaredURL|TestAcquireDraftNetworkRevision2Positives|TestAcquireDraftNetworkRevision2ExhaustionSanitized|TestAcquireDraftNetworkRevision2Refusals|TestRevision1ReaderRejectsV2Shapes|TestUserConfigIgnoredByResolvedLane|TestDraftTransportPlanV2Conversion|TestDraftPlanNeedsSSHRevision2|TestDryRunEffectBindingsSeeWhatARealOperationWrites|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal|TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestAuthoritativePathShimCasesUseProjectInstallEntry|TestAuthoritativeSigningCasesUseProjectInstallEntry|TestAuthoritativeTransactionCasesUseProjectInstallEntry|TestRevisionSubstitutionDerivesEffectiveObjectFormat|TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestReadDocumentBindsGenerationToBytesRewrittenInPlace|TestReadDocumentBindsGenerationToBytesReplacedByRename|TestDeclarationGenerationIsContentAddressed|TestDocumentGenerationIsStableWhenUnreadable|TestObservationsRecheckDocumentsAgainstTheParsedBytes|TestGitFixtureRetriesOnceOnSpawnEACCES|TestGitFixtureRetriesOnceOnSpawnEAGAIN|TestGitFixtureNeverRetriesNonZeroExit|TestGitFixtureNeverRetriesWrappedExitError|TestGitFixtureRetryPredicate|TestGitFixturePersistentSpawnFailureCarriesDiagnostic|TestGitFixtureNonRetryableSpawnErrorFailsFastWithDiagnostic|TestGitFixtureRealSpawnEACCESRetriesOnce|TestGitFixtureUnresolvableGitFailsFastWithDiagnostic|TestGitFixtureOuterEntryRunsRealGit|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES|TestProjectRefusesOnGitignoreSpawnFailure|TestProjectNonGitWithDevSubstitution|TestGlobalInstallRecognizesAdoptedForwardingShim|TestGlobalReconcilesOwnedLegacyForwardingLink|TestEndToEndInstall|TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals|TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution|TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook|TestRuntimeLauncherCapturesDeclaredSystemDependency|TestSecondInstallIsUpToDate|TestTamperTriggersReinstall|TestRemovedSkillCleanedUp|TestDryRunTouchesNothing)$' ./internal/install
```

`full-install-05` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestGitignoreGateSkips|TestMissingSystemCommandFails|TestMovedTagWarningAndStrict|TestRuntimeOnlyProviderGetsMarkerNoAdapter|TestRuntimeOnlyProviderStillRequiresSkillMd|TestHybridSkillActivatesWithoutTouchingProjectStore|TestHybridShadowedByProjectDeclaration|TestGlobalInstall|TestGlobalInstallRejectsSchema2Skillfile|TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged|TestGlobalInstallUsesManifestLocaleAndAuditGate|TestGlobalStrictTagsDetectMovedTag|TestMcpRequirementGatesInstall|TestAuditGateBlocksUndeclaredNetwork|TestBrokenTransitiveManifestNamesTheBrokenSkill|TestAuditAdmitsVendoredThirdPartyText|TestInstallScriptSecurityRows|TestLegacyLaneBuildsPublishReceipt3|TestLegacyLaneBuildMarkerRebindsTamperedReceipt|TestLegacyLaneRecordsDependencyDirectory|TestLegacyLockPackageArms|TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode|TestLegacyLaneRefusesGlobDependencyDirectory|TestLegacyLaneRejectsUnboundPreviousGeneration|TestLegacyLaneDependencyTagMovement|TestLegacyLaneMovedTagRefusesStrictSecondInstall|TestReviewChangedDeclaredTagIsNotMovedTag|TestLegacyLaneSameTagMoveWithAlias|TestLegacyLaneChangedTagAfterOldTagDeleted|TestLegacyLaneRuntimeMatchesPackageMarker|TestGlobalLaneRuntimeMatchesPackageMarker|TestPostCommitMaintenanceRunsUnderTheHeldHomeLock|TestPostCommitMaintenanceWarningsReachTheResult|TestPostCommitMaintenanceMarksInFlightJournals|TestDeclaredModuleRootsReachTheBuilder|TestASchemaSixCommandDeclaresNoModules|TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled|TestV1ReinstallRefusesNULAppearingAfterInstall|TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal|TestV2InstallDistinguishesNULCollisionTwins|TestDraftLaneStillBlocksNULBearingMember|TestInstallDraftOwnershipCases|TestInstallPreservesUnownedStaleStoreEntries|TestInstallRefusesUnownedSelectedCanonicalDestination|TestInstallRefusesUnprovenStaleOwnership|TestInstallOwnedCanonicalPositiveControls|TestGlobalCanonicalOwnershipPreservation|TestHybridCanonicalOwnershipPreservation|TestInstallDraftMarkerWriterCases|TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestDryRunRemovesItsOperationPrivateRootOnFailure|TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure|TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot|TestGlobalMcpFailureBlocksToolchainCacheAndBuild|TestGlobalRegistryFailureBlocksToolchainCacheAndBuild|TestGlobalMarkersCarryMcpAndAttestationEvidence|TestRegistryRevocationDeniesInstall|TestRegistryAttestationLandsInMarker|TestV1InstallRejectsV2RegistryEvidence)$' ./internal/install
```

`full-install-06` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestV2RegistryAttestationUsesVersionedInstallPath|TestRegistryFirstUseReportsAndPersistsTOFUPosture|TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot|TestRegistryInvalidFirstUseCheckpointFailsClosed|TestRegistryRebootstrapRegressionKeepsExistingHighWater|TestRegistryBootstrapVectorsDriveInstall|TestStrictRegistryPolicyFailsUnknown|TestRegistryMirrorComparisonVectorsDriveInstall|TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch|TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses|TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries|TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall|TestStableDeclarationInputsCommitWithoutRestarting|TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState|TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure|TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestStageProjectTargetsPopulatesBoundaries|TestStageProjectTargetsRefusesMirrorOverSnapshot|TestStageProjectTargetsRefusesProjectInsideSnapshot|TestStageProjectTargetsRefusesMissingSnapshot|TestStageGlobalTargetsPopulatesBoundaries|TestStageGlobalTargetsRefusesMirrorOverSnapshot|TestRunCommitWithRealStagingJournalsOneTransaction|TestRunCommitWithRealStagingRefusesOverlapBeforeJournaling|TestPerWriteGuardRefusesSwappedParentAndRollsBack|TestProjectInstallRefusesAdapterDestinationOverSnapshot|TestGlobalInstallRefusesAdapterDestinationOverSnapshot|TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState|TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState|TestGlobalStagingFailureLeavesGlobalScopeUnchanged|TestDryRunReportsCacheHitWithoutBuilding|TestCacheHitPerformsNoSourceAwareGoCommand|TestStagingRunsProviderFirstAndCommandLexical|TestStagedOutputsStayPrivateAndAreReleased|TestSecondBuildFailurePreservesPriorInstallationAndLiveCache|TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState|TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope|TestCacheHitOnlyPlanStillFinalizesToolchainTrust|TestCacheInspectionIsBracketedByTheFrozenSource|TestSourceMutationDuringStagingBlocksHandoffOfACacheHit|TestToolchainFailureBlocksEveryPersistentMutation|TestCorruptCacheEntryIsRebuiltAndNeverReused|TestUnsupportedCacheProtectionFailsClosed|TestIncompletePlansNeverClaimACompleteDiagnostic|TestToolchainFailurePlansAnInventoryOfEveryActiveCommand|TestUntrustedCacheEntryIsRebuiltAndNeverReused)$' ./internal/install
```

`full-install-07` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestScriptOnlyInstallPerformsNoToolchainOrCacheWork|TestInjectedClockAndGenerationReaderDriveInstallMarkers|TestPlannedBuildAccessorsAreImmutable|TestReleaseTakesNoToolchainVerdictAfterLiveMutation|TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation|TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall|TestDefaultToolchainProbeRemovesItsProbeRootOnFailure|TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess|TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure|TestProjectWriteBlobsSpawnFailureIsWrapped)$' ./internal/install
```

`full-install-08` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json -run '^(TestAuthoritativeDryRunCasesMutateNothingPersistent)$' ./internal/install
```

`full-other-01` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/cmd/curator-spec-pin github.com/relux-works/curator/internal/adapters github.com/relux-works/curator/internal/artifactpolicy github.com/relux-works/curator/internal/artifactpolicy/conformance github.com/relux-works/curator/internal/audit github.com/relux-works/curator/internal/buildcache github.com/relux-works/curator/internal/buildmeta github.com/relux-works/curator/internal/buildrepo github.com/relux-works/curator/internal/buildsource github.com/relux-works/curator/internal/capabilities github.com/relux-works/curator/internal/closure github.com/relux-works/curator/internal/closureexec github.com/relux-works/curator/internal/closuregraph github.com/relux-works/curator/internal/config github.com/relux-works/curator/internal/conformancecoverage
```

`full-other-02` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/contextaudit github.com/relux-works/curator/internal/contextlock github.com/relux-works/curator/internal/contextmaterialize github.com/relux-works/curator/internal/contextpkg github.com/relux-works/curator/internal/contextresolve github.com/relux-works/curator/internal/contextstore github.com/relux-works/curator/internal/crossconformance github.com/relux-works/curator/internal/devsub github.com/relux-works/curator/internal/envfiles github.com/relux-works/curator/internal/envfragment github.com/relux-works/curator/internal/envmarker github.com/relux-works/curator/internal/envregistry github.com/relux-works/curator/internal/envsnapshot github.com/relux-works/curator/internal/gitcred github.com/relux-works/curator/internal/gitignore
```

`full-other-03` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/gitops github.com/relux-works/curator/internal/globalbins github.com/relux-works/curator/internal/godriver github.com/relux-works/curator/internal/hashing github.com/relux-works/curator/internal/hookapproval github.com/relux-works/curator/internal/identifiers github.com/relux-works/curator/internal/identity github.com/relux-works/curator/internal/install/atomicity github.com/relux-works/curator/internal/interop github.com/relux-works/curator/internal/interop/environments github.com/relux-works/curator/internal/locale github.com/relux-works/curator/internal/managerlock github.com/relux-works/curator/internal/manifest github.com/relux-works/curator/internal/marker github.com/relux-works/curator/internal/mcp
```

`full-other-04` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/moduleroots github.com/relux-works/curator/internal/nodesource github.com/relux-works/curator/internal/npmsource github.com/relux-works/curator/internal/opaquescan github.com/relux-works/curator/internal/pathboundary github.com/relux-works/curator/internal/pkgversion github.com/relux-works/curator/internal/pnpmsource github.com/relux-works/curator/internal/privatedir github.com/relux-works/curator/internal/protocoljson github.com/relux-works/curator/internal/registry github.com/relux-works/curator/internal/runtimestore github.com/relux-works/curator/internal/rustsource github.com/relux-works/curator/internal/scopes github.com/relux-works/curator/internal/scriptpolicy github.com/relux-works/curator/internal/scriptworker
```

`full-other-05` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/shell github.com/relux-works/curator/internal/skillcheck github.com/relux-works/curator/internal/skillspec github.com/relux-works/curator/internal/snapshot github.com/relux-works/curator/internal/sourcelock github.com/relux-works/curator/internal/staging github.com/relux-works/curator/internal/stateread github.com/relux-works/curator/internal/swiftpmbuild github.com/relux-works/curator/internal/swiftpminterop github.com/relux-works/curator/internal/swiftpmsource github.com/relux-works/curator/internal/testcli github.com/relux-works/curator/internal/testgitenv github.com/relux-works/curator/internal/testtoolchain github.com/relux-works/curator/internal/transaction github.com/relux-works/curator/internal/ui
```

`full-other-06` (exit 0):

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test -count=1 -timeout=8m -json github.com/relux-works/curator/internal/verr github.com/relux-works/curator/internal/version github.com/relux-works/curator/internal/whitelist github.com/relux-works/curator/internal/yarnclassicsource github.com/relux-works/curator/internal/yarnmodernsource github.com/relux-works/curator/tools/ciworkflow github.com/relux-works/curator/tools/goreleaserconfig
```

### Narrowing-control receipts

The restored physical-entry matcher produced child exit 1 for all nine alias witnesses. Examples: `visible Markdown verdict: [], want network refusal for SKILL.md` and the corresponding `references/guide.md` assertion.

The namespace-reuse mutant first generated a real ruleset-6 allowance successfully (seed child exit 0), then produced child exit 1 with `visible Markdown verdict: [], want network refusal for references/guide.md` in both v1 and v2. Thus the kill was the security regression, not a missing cache file or unrelated setup error.

The every-Markdown mutant produced child exit 1 with the named `.research/notes.md`, `.task-board/.resources/x.md`, nested build/runtime and vendor README findings. The outer mutant harness exited 0 only after confirming these intended behavioral failures.

## Delivery verification

`git diff --check` exit 0. The owning checkout and the normal-.git validation checkout were clean. `git log` verified all three introduced commits with `G`:

- `35f8a18d50795636e342baf87abc491bf7ce1aca`
- `2c01da83de4007c5d6f6d4edf3a887dc85707795`
- `9b17768c96d53e9ef7babb45e95d3b688728e815`

The new commit used explicit `git commit -S` (exit 0). `git push -u origin fix/audit-markdown-scope` exited 0; `git ls-remote --heads origin fix/audit-markdown-scope` exited 0 and returned the exact handed-over head. No force push was used.

Board assignment and all four preconditions were read successfully; the review outcome was read in full. Ack and fetch exited 0. The board status was not changed. Task-local validation clone, module/build caches, image and scratch are scheduled for deletion immediately after handover; this report is retained as the outcome evidence.
