# rev4 hosted gate 37307472674: remaining failures (orchestrator extract)
Down from 46+ (rev3) to 2 on ubuntu/macOS (family C: env-marker schema-1 parsing) and 11 Windows-only leaves in internal/envprofile migration.

===== ubuntu-latest: 2 failed (2 leaves)
-- cmd/curator TestEnvResolveKeepsSchema1BytesForMetadataOnly
       env_credential_marker_test.go:190: schema-1 repair = 1
       env_credential_marker_test.go:190: schema-1 repair = 1
-- cmd/curator TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
===== macos-latest: 2 failed (2 leaves)
-- cmd/curator TestEnvResolveKeepsSchema1BytesForMetadataOnly
       env_credential_marker_test.go:190: schema-1 repair = 1
       env_credential_marker_test.go:190: schema-1 repair = 1
-- cmd/curator TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
===== windows-latest: 12 failed (11 leaves)
-- cmd/curator TestEnvResolveKeepsSchema1BytesForMetadataOnly
       env_credential_marker_test.go:190: schema-1 repair = 1
       env_credential_marker_test.go:190: schema-1 repair = 1
-- cmd/curator TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
       env_credential_marker_test.go:253: pre-rule schema-1 repair = 1
-- internal/envprofile TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode
       migrate_test.go:2033: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMo2788271051\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMo2788271051\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchem
       migrate_test.go:2033: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMo2788271051\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMo2788271051\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestMigrateSchem
-- internal/envprofile TestRC14IdentityMigrationPreservesFallbackCopies
       rc14_cutover_test.go:374: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCopies2871097615\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCopies2871097615\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCo
       rc14_cutover_test.go:374: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCopies2871097615\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCopies2871097615\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationPreservesFallbackCo
-- internal/envprofile TestRC14IdentityMigrationRecoversInterruptedCommit
       rc14_cutover_test.go:263: unexpected interruption: <nil>
       rc14_cutover_test.go:263: unexpected interruption: <nil>
-- internal/envprofile TestRC14IdentityMigrationRollsBackEveryEntry
       rc14_cutover_test.go:167: retry of rolled-back plan: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationRollsBackEveryEntry3286443137\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationRollsBackEveryEntry3286443137\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigration
       rc14_cutover_test.go:167: retry of rolled-back plan: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationRollsBackEveryEntry3286443137\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigrationRollsBackEveryEntry3286443137\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14IdentityMigration
-- internal/envprofile TestRC14MigrationRehashesLegacyIdentities
       rc14_cutover_test.go:48: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002 DAC
       rc14_cutover_test.go:48: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14MigrationRehashesLegacyIdentities1027712106\002 DAC
-- internal/envprofile TestRC14ResolveRepairMigratesSiblingHomes
       rc14_cutover_test.go:123: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002 DACL 
       rc14_cutover_test.go:123: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14ResolveRepairMigratesSiblingHomes771345173\002 DACL 
-- internal/envprofile TestRC14Schema1IdentityMigrationWithoutCredentialOperations/migrate
       rc14_cutover_test.go:330: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsmigr841986372\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsmigr841986372\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1I
       rc14_cutover_test.go:330: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsmigr841986372\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsmigr841986372\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1I
-- internal/envprofile TestRC14Schema1IdentityMigrationWithoutCredentialOperations/resolve
       rc14_cutover_test.go:315: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsreso3059529932\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsreso3059529932\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema
       rc14_cutover_test.go:315: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsreso3059529932\002 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema1IdentityMigrationWithoutCredentialOperationsreso3059529932\002: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14Schema
-- internal/envprofile TestRC14UseMigratesLegacyProfileBeforeNativePublication
       rc14_cutover_test.go:404: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfileBeforeNativePublication2788201817\003 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfileBeforeNativePublication2788201817\003: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfi
       rc14_cutover_test.go:404: permissions: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfileBeforeNativePublication2788201817\003 boundary check failed at C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfileBeforeNativePublication2788201817\003: C:\Users\RUNNER~1\AppData\Local\Temp\TestRC14UseMigratesLegacyProfi

## Orchestrator hints (not verified)
- Windows: every remaining Windows failure is a `permissions: … boundary check failed` on the test temp root (internal/pathboundary). Compare with how passing envprofile tests prepare a Windows home: look at `pinOperatorHome(t, dir)` and the managedFixture setup. The new rc14 cutover and migrate tests probably skip that protection on Windows. `TestRC14IdentityMigrationRecoversInterruptedCommit` (`unexpected interruption: <nil>`) is likely downstream of the same failure.
- Unix and Windows: `schema-1 repair = 1` in the env_credential_marker tests is family C. A schema-1 env marker must still repair byte-for-byte.
