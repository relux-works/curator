package artifactpolicy

import (
	"path/filepath"
	"strconv"
)

const (
	toolchainFingerprintAlgorithm = "curator-toolchain-tree-v1"
	findingsDigestAlgorithm       = "artifact-finding-set-v1"
)

func validateToolchainAuthorization(
	authorization ToolchainAuthorization,
	pathValue string,
	captured blob,
	expectedSeal *authorizationSeal,
) (DiagnosticCode, string, []Fact) {
	if authorization == nil {
		return CodeToolchainUntrusted, "toolchain_checkpoint_missing", nil
	}
	record := authorization.artifactPolicyToolchainAuthorization()
	factsValue := toolchainRoleFacts(record, expectedSeal)
	if code, reason := validateToolchainRecord(record, expectedSeal); code != "" {
		return code, reason, factsValue
	}
	if record.payloadPath != pathValue || record.executableRelativePath != pathValue {
		return CodeToolchainUntrusted, "toolchain_checkpoint_path_replay", factsValue
	}
	if record.payloadSHA256 != captured.sha256 || record.payloadSize != captured.size {
		return CodeToolchainIdentityChanged, "toolchain_checkpoint_payload_replay", factsValue
	}
	return "", "", factsValue
}

func validateToolchainRecord(record toolchainAuthorizationRecord, expectedSeal *authorizationSeal) (DiagnosticCode, string) {
	if expectedSeal == nil || record.seal != expectedSeal {
		return CodeToolchainUntrusted, "toolchain_checkpoint_not_manager_issued"
	}
	if record.policySelector == "" || !filepath.IsAbs(record.resolvedRoot) ||
		record.executableRelativePath == "" || record.environmentSearchResolution == "" ||
		record.version == "" || record.platform == "" ||
		record.fingerprintAlgorithm != toolchainFingerprintAlgorithm ||
		!record.outsideDependencyClosure || !record.containedLinksValidated || !record.ordinaryNodesValidated ||
		!sha256Identity.MatchString(record.checkpointFingerprintSHA256) ||
		!sha256Identity.MatchString(record.timeOfUseFingerprintSHA256) ||
		!sha256Identity.MatchString(record.payloadSHA256) || record.payloadSize < 0 {
		return CodeToolchainUntrusted, "toolchain_role_evidence_incomplete"
	}
	validated, err := ValidateVirtualPath(filepath.ToSlash(record.executableRelativePath))
	if err != nil || validated.Canonical != filepath.ToSlash(record.executableRelativePath) {
		return CodeToolchainUntrusted, "toolchain_executable_path_invalid"
	}
	payloadPath, err := ValidateVirtualPath(record.payloadPath)
	if err != nil || payloadPath.Canonical != record.payloadPath {
		return CodeToolchainUntrusted, "toolchain_payload_path_invalid"
	}
	resolvedExecutable := filepath.Clean(filepath.Join(record.resolvedRoot, filepath.FromSlash(validated.Canonical)))
	relative, err := filepath.Rel(filepath.Clean(record.resolvedRoot), resolvedExecutable)
	if err != nil || relative == ".." || filepath.IsAbs(relative) ||
		len(relative) > 3 && relative[:3] == ".."+string(filepath.Separator) {
		return CodeToolchainUntrusted, "toolchain_executable_path_escape"
	}
	if filepath.Clean(record.environmentSearchResolution) != resolvedExecutable {
		return CodeToolchainUntrusted, "toolchain_environment_resolution_mismatch"
	}
	if record.checkpointFingerprintSHA256 != record.timeOfUseFingerprintSHA256 {
		return CodeToolchainIdentityChanged, "toolchain_fingerprint_changed_before_use"
	}
	return "", ""
}

func toolchainRoleFacts(record toolchainAuthorizationRecord, expectedSeal *authorizationSeal) []Fact {
	return facts(map[string]any{
		"authorization":                 authorizationEvidence(record.seal, expectedSeal),
		"checkpoint_fingerprint":        record.checkpointFingerprintSHA256,
		"contained_links_validated":     record.containedLinksValidated,
		"environment_search_resolution": record.environmentSearchResolution,
		"executable_relative_path":      record.executableRelativePath,
		"fingerprint_algorithm":         record.fingerprintAlgorithm,
		"ordinary_nodes_validated":      record.ordinaryNodesValidated,
		"outside_dependency_closure":    record.outsideDependencyClosure,
		"payload_path":                  record.payloadPath,
		"payload_sha256":                record.payloadSHA256,
		"payload_size":                  record.payloadSize,
		"platform":                      record.platform,
		"policy_selector":               record.policySelector,
		"resolved_root":                 record.resolvedRoot,
		"time_of_use_fingerprint":       record.timeOfUseFingerprintSHA256,
		"version":                       record.version,
	})
}

func validateLocalOutputAuthorization(
	authorization LocalOutputAuthorization,
	pathValue string,
	captured blob,
	expectedSeal *authorizationSeal,
) (DiagnosticCode, string, []Fact, ArtifactExpectation) {
	if authorization == nil {
		return CodeLocalOutputUnreceipted, "local_output_receipt_missing", nil, ArtifactExpectation{}
	}
	record := authorization.artifactPolicyLocalOutputAuthorization()
	factsValue := localOutputRoleFacts(record, expectedSeal)
	if code, reason := validateLocalOutputRecord(record, expectedSeal); code != "" {
		return code, reason, factsValue, record.expectation
	}
	if record.payloadPath != pathValue || record.expectation.Path != pathValue {
		return CodeLocalOutputDrift, "output_path_replay", factsValue, record.expectation
	}
	if record.payloadSHA256 != captured.sha256 || record.payloadSize != captured.size {
		return CodeLocalOutputDrift, "output_payload_replay", factsValue, record.expectation
	}
	return "", "", factsValue, record.expectation
}

func validateLocalOutputRecord(record localOutputAuthorizationRecord, expectedSeal *authorizationSeal) (DiagnosticCode, string) {
	if expectedSeal == nil || record.seal != expectedSeal {
		return CodeLocalOutputUnreceipted, "local_output_receipt_not_manager_issued"
	}
	if record.declaredActionID == "" || record.stagingRootIdentity == "" ||
		record.protectedStoreIdentity == "" || record.payloadPath == "" ||
		!sha256Identity.MatchString(record.sourceClosureDigest) ||
		!sha256Identity.MatchString(record.artifactManifestDigest) ||
		!sha256Identity.MatchString(record.buildPlanDigest) ||
		!sha256Identity.MatchString(record.executionReceiptSHA256) ||
		!sha256Identity.MatchString(record.protectedReceiptSHA256) ||
		!sha256Identity.MatchString(record.payloadSHA256) || record.payloadSize < 0 ||
		!record.stagingStartedEmpty || !record.observedProduction || !record.writeSetMatched ||
		!record.preexistingInputExcluded || !record.hardlinkSourceExcluded ||
		!record.expectationIndependentlyDerived ||
		!record.protectedPublicationValidated {
		return CodeLocalOutputUnreceipted, "local_output_causal_receipt_incomplete"
	}
	if !record.completeInputMatched {
		return CodeLocalOutputDrift, "local_output_complete_input_drift"
	}
	if _, ok := artifactClasses[record.expectation.Class]; !ok ||
		record.expectation.Path != record.payloadPath ||
		record.expectation.SHA256 != record.payloadSHA256 ||
		record.expectation.Size != record.payloadSize {
		return CodeLocalOutputDrift, "local_output_expectation_binding_invalid"
	}
	if _, err := ValidateVirtualPath(record.payloadPath); err != nil {
		return CodeLocalOutputDrift, "local_output_path_invalid"
	}
	return "", ""
}

func localOutputRoleFacts(record localOutputAuthorizationRecord, expectedSeal *authorizationSeal) []Fact {
	return facts(map[string]any{
		"artifact_manifest_digest":          record.artifactManifestDigest,
		"authorization":                     authorizationEvidence(record.seal, expectedSeal),
		"build_plan_digest":                 record.buildPlanDigest,
		"complete_input_matched":            record.completeInputMatched,
		"declared_action_id":                record.declaredActionID,
		"execution_receipt":                 record.executionReceiptSHA256,
		"expectation_class":                 record.expectation.Class,
		"expectation_digest":                record.expectation.SHA256,
		"expectation_independently_derived": record.expectationIndependentlyDerived,
		"expectation_path":                  record.expectation.Path,
		"expectation_size":                  record.expectation.Size,
		"hardlink_source_excluded":          record.hardlinkSourceExcluded,
		"observed_production":               record.observedProduction,
		"payload_path":                      record.payloadPath,
		"payload_sha256":                    record.payloadSHA256,
		"payload_size":                      record.payloadSize,
		"preexisting_input_excluded":        record.preexistingInputExcluded,
		"protected_publication_validated":   record.protectedPublicationValidated,
		"protected_receipt":                 record.protectedReceiptSHA256,
		"protected_store_identity":          record.protectedStoreIdentity,
		"source_closure_digest":             record.sourceClosureDigest,
		"staging_root_identity":             record.stagingRootIdentity,
		"staging_started_empty":             record.stagingStartedEmpty,
		"write_set_matched":                 record.writeSetMatched,
	})
}

func authorizationEvidence(seal, expected *authorizationSeal) string {
	if expected != nil && seal == expected {
		return "manager-issued-v1"
	}
	return "unsealed-or-foreign"
}

func dependencyRoleFacts(origin OriginEvidence) []Fact {
	return facts(map[string]any{
		"origin_checksum_sha256": origin.ChecksumSHA256,
		"origin_immutable_id":    origin.ImmutableID,
		"origin_locator":         origin.Locator,
		"origin_lock_record":     origin.LockRecord,
		"origin_verified":        origin.Verified,
	})
}

func verifiedBinaryRoleFacts() []Fact {
	return []Fact{
		{Key: "capability", Value: "verified-binary-v1"},
		{Key: "capability_available", Value: strconv.FormatBool(false)},
	}
}
