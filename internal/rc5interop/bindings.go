// Package rc5interop binds every shared rc.5 external-repository case to the
// Curator black-box or native regression that exercises the behavior.
package rc5interop

// Binding names the native Go package and focused test used by qualification.
// The shared corpus owns inputs and expected outcomes; these names are only the
// Curator adapter layer and contain no copied normative values.
type Binding struct {
	Package string
	Test    string
}

var Bindings = map[string]Binding{
	"sha1-tag-match-https":                {"./internal/buildrepo", "TestNetworkAndLocalSHA1SHA256RawObjectParity"},
	"sha1-tag-match-ssh":                  {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"sha1-tag-moved":                      {"./internal/buildrepo", "TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit"},
	"sha1-tag-missing":                    {"./internal/buildrepo", "TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit"},
	"sha256-untagged":                     {"./internal/buildrepo", "TestNetworkAndLocalSHA1SHA256RawObjectParity"},
	"untagged-object-missing":             {"./internal/buildrepo", "TestAdmissionFailuresPrecedeAuditCacheAndCompiler"},
	"canonical-https-ssh-scp":             {"./internal/buildrepo", "TestCanonicalRepositorySourceVectors"},
	"operator-local-identity":             {"./internal/buildrepo", "TestLocalIdentityDoesNotExposeHostPath"},
	"clean-git-session":                   {"./internal/buildrepo", "TestAdmissionFailuresPrecedeAuditCacheAndCompiler"},
	"exact-fetch-closed-shape":            {"./internal/buildrepo", "TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit"},
	"ssh-wrapper-closed-shape":            {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"raw-object-reader-closed-shape":      {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"monorepo-root-target":                {"./internal/buildrepo", "TestExternalPipelineCompilerSeesOnlySelectedBuildRoot"},
	"monorepo-nested-target":              {"./internal/buildrepo", "TestExternalPipelineCompilerSeesOnlySelectedBuildRoot"},
	"local-substitution":                  {"./internal/devsub", "TestSchema2BuildRepositorySubstitutions"},
	"network-substitution-revision":       {"./internal/devsub", "TestSchema2BuildRepositorySubstitutions"},
	"network-substitution-tag":            {"./internal/devsub", "TestSchema2BuildRepositorySubstitutions"},
	"network-substitution-branch":         {"./internal/devsub", "TestSchema2BuildRepositorySubstitutions"},
	"raw-object-malformed":                {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"lfs-pointer":                         {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"submodule-gitlink":                   {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"symbolic-link":                       {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"special-file-mode":                   {"./internal/buildrepo", "TestRawObjectAndLFSPinnedConformanceFixtures"},
	"alternate-object-store":              {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"replace-ref":                         {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"graft":                               {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"promisor-pack":                       {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"partial-clone":                       {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"gitfile":                             {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"linked-worktree":                     {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"bare-repository":                     {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"reftable":                            {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"object-link":                         {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"filter-config-inert":                 {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"credential-helper-inert":             {"./internal/buildrepo", "TestLocalConfigAndAdministrationAdversarialBoundaries"},
	"pack-v2-sha1":                        {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"pack-v3-sha1":                        {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"pack-v2-sha256":                      {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"pack-index-checksum-mismatch":        {"./internal/buildrepo", "TestPackIndexConformanceAndExactSSHWrapper"},
	"audit-order-cache-hit":               {"./internal/buildrepo", "TestExternalPipelineCacheHitRepeatsAdmissionValidationAndAudit"},
	"audit-order-cache-miss":              {"./internal/buildrepo", "TestExternalPipelineOrderingAcrossOperations"},
	"cache-corrupt-receipt":               {"./internal/install", "TestCorruptCacheEntryIsRebuiltAndNeverReused"},
	"cache-corrupt-artifact":              {"./internal/install", "TestCorruptCacheEntryIsRebuiltAndNeverReused"},
	"protected-offline-reuse":             {"./internal/buildrepo", "TestExternalPipelineOfflineProtectedSnapshotAndTagFailure"},
	"offline-syntax-only":                 {"./internal/buildrepo", "TestExternalPipelineOfflineProtectedSnapshotAndTagFailure"},
	"offline-install-without-snapshot":    {"./internal/buildrepo", "TestExternalPipelineOfflineProtectedSnapshotAndTagFailure"},
	"mixed-receipt-v1-v2-marker-v3":       {"./internal/marker", "TestMarkerV3StructurallyRepresentsLocalExternalAndMixed"},
	"external-receipt-v2-exact-bytes":     {"./internal/buildrepo", "TestExternalReceiptV2CacheKeyVector"},
	"status-current":                      {"./cmd/curator", "TestCompiledProjectStatusRepairRollbackRecovery"},
	"status-corrupt":                      {"./cmd/curator", "TestCompiledProjectStatusRepairRollbackRecovery"},
	"repair-reacquires":                   {"./cmd/curator", "TestCompiledProjectStatusRepairRollbackRecovery"},
	"gc-retains-marker-and-journal-roots": {"./internal/buildrepo", "TestExternalGCUsesMarkerAndJournalKeysAsOnlyRoots"},
	"shim-path-structural":                {"./internal/runtimestore", "TestCompiledShimsStageWithoutLaunchThenPostInstallForwardExactly"},
	"path-collision":                      {"./internal/runtimestore", "TestStagingRejectsLiveOverlapAndUnsafePathEntries"},
	"package-argv-forbidden":              {"./internal/skillspec", "TestSchema7RepositoryCommandAndDeclarationStayClosed"},
	"shim-collision-rollback":             {"./internal/install", "TestSecondBuildFailurePreservesPriorInstallationAndLiveCache"},
	"consumer-last-rollback":              {"./internal/install", "TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess"},
	"package-signing-request":             {"./internal/buildrepo", "TestExternalPipelineOrderingAcrossOperations"},
	"platform-requires-signing":           {"./internal/buildrepo", "TestExternalPipelineOrderingAcrossOperations"},
	"truthful-platform-claims":            {"./internal/godriver", "TestNativeControlInventoryIsExhaustiveAndClosed"},
}
