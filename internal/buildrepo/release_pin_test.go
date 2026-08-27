package buildrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleasedRC8Pin(t *testing.T) {
	if ProtocolVersion != "1.0.0-rc.8" || SpecReleaseTag != "v1.0.0-rc.8" || SpecReleaseCommit != "f8c405aa3ad0a39d260c2ed93684e55c5a346359" || SpecReleaseTagObject != "ad247840292487d5d88ac44331798b6b4182a79f" || ConformanceManifestSHA256 != "d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1" || ReleaseMetadataSHA256 != "293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede" {
		t.Fatal("rc.8 release identity changed")
	}
}

func TestVerifyReleasePinRejectsMutableMismatchedUnknownAndClaimInflatingPins(t *testing.T) {
	root, pin := writeSyntheticReleasePin(t)
	if err := verifyReleasePin(root, pin.SpecCommit, pin); err != nil {
		t.Fatalf("synthetic valid release pin rejected: %v", err)
	}

	for _, revision := range []string{"main", "v1.0.0-rc.8", "HEAD", "f8c405a", strings.ToUpper(pin.SpecCommit)} {
		if err := verifyReleasePin(root, revision, pin); err == nil || !strings.Contains(err.Error(), "mutable") {
			t.Errorf("mutable revision %q: error = %v", revision, err)
		}
	}
	if err := verifyReleasePin(root, strings.Repeat("b", 40), pin); err == nil || !strings.Contains(err.Error(), "revision mismatch") {
		t.Fatalf("mismatched immutable revision: error = %v", err)
	}

	writeTestFile(t, filepath.Join(root, "conformance", "v1", "manifest.json"), []byte("{\"protocol_version\":\"1.0.0-rc.8\",\"drift\":true}\n"))
	if err := verifyReleasePin(root, pin.SpecCommit, pin); err == nil || !strings.Contains(err.Error(), "conformance manifest SHA-256 mismatch") {
		t.Fatalf("mismatched conformance suite: error = %v", err)
	}

	root, pin = writeSyntheticReleasePin(t)
	mutateSyntheticMetadata(t, root, &pin, func(metadata map[string]any) {
		metadata["protocol_version"] = "1.0.0-rc.unknown"
	})
	if err := verifyReleasePin(root, pin.SpecCommit, pin); err == nil || !strings.Contains(err.Error(), "protocol version mismatch") {
		t.Fatalf("unknown release version: error = %v", err)
	}

	root, pin = writeSyntheticReleasePin(t)
	mutateSyntheticMetadata(t, root, &pin, func(metadata map[string]any) {
		metadata["assurance"].(map[string]any)["verified_platform_claims"] = []any{map[string]any{"platform": "unpublished"}}
	})
	if err := verifyReleasePin(root, pin.SpecCommit, pin); err == nil || !strings.Contains(err.Error(), "inflates") {
		t.Fatalf("claim-inflating release: error = %v", err)
	}
}

func TestVerifyPublishedReleasePinWhenConformanceRootIsSupplied(t *testing.T) {
	conformanceRoot := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if conformanceRoot == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	// The supplied root and this module's immutable pin are promoted by
	// separate owners: CI's SPEC_PIN moves when a newer release qualifies,
	// and this pin moves only with the promotion task named in the workflow.
	// Verifying an unrelated release against this pin would report a drift
	// that is not one, so the release the root actually publishes is
	// identified first and a different one is declared rather than asserted.
	manifest, err := os.ReadFile(filepath.Join(filepath.Clean(conformanceRoot), "manifest.json"))
	if err != nil {
		t.Fatalf("read the supplied conformance manifest: %v", err)
	}
	if digestTestBytes(manifest) != ConformanceManifestSHA256 {
		t.Skipf("the supplied conformance root publishes a release other than the pinned %s", SpecReleaseTag)
	}
	specRoot := filepath.Dir(filepath.Dir(filepath.Clean(conformanceRoot)))
	if err := VerifyReleasePin(specRoot, SpecReleaseCommit); err != nil {
		t.Fatal(err)
	}
}

func writeSyntheticReleasePin(t *testing.T) (string, releasePinIdentity) {
	t.Helper()
	root := t.TempDir()
	manifest := []byte("{\"protocol_version\":\"1.0.0-rc.8\"}\n")
	writeTestFile(t, filepath.Join(root, "conformance", "v1", "manifest.json"), manifest)
	manifestDigest := digestTestBytes(manifest)
	metadata := validSyntheticMetadata("sha256:" + manifestDigest)
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, filepath.FromSlash(ReleaseMetadataPath)), metadataBytes)
	return root, releasePinIdentity{
		ProtocolVersion:           "1.0.0-rc.8",
		SpecCommit:                strings.Repeat("a", 40),
		ConformanceManifestSHA256: manifestDigest,
		ReleaseMetadataSHA256:     digestTestBytes(metadataBytes),
	}
}

func validSyntheticMetadata(manifestDigest string) map[string]any {
	return map[string]any{
		"assurance":              map[string]any{"default_mode": "portable", "portable_execution_policy": "manager-worker-v1", "portable_policy": "portable-cli-policy-v1", "silent_downgrade_permitted": false, "skill_vendored_provider_allowed": false, "verified_execution_policy": "verified-provider-execution-v1", "verified_implementations": []any{}, "verified_platform_claims": []any{}, "verified_policy": "verified-provider-policy-v1", "verified_provider_contract": "host-execution-provider-v1"},
		"candidate_protocol_pin": map[string]any{"manifest_sha256": manifestDigest, "suite_root": "conformance/v1"},
		"claim_v4":               map[string]any{"claim_protocol_version": "1.0.0-rc.8", "claims_emitted": []any{}, "schema": "schemas/v1/conformance-claim-v4.schema.json"},
		"created_at":             "2026-08-19T00:00:00Z",
		"downstream_consumption": map[string]any{"committed_release_pin_advanced": false, "environment": "CURATOR_CONFORMANCE_ROOT", "required_manifest_sha256": manifestDigest},
		"historical_release":     map[string]any{"immutable": true, "metadata_path": "release/1.0.0-rc.7.json", "metadata_sha256": "sha256:" + LegacyReleaseMetadataSHA256, "protocol_version": "1.0.0-rc.7", "source_commit": "99f70947d6f2447366d6c996127b73eca37a9159"},
		"legacy_release":         "1.0.0-rc.7", "protocol_version": "1.0.0-rc.8", "source_baseline_commit": "99f70947d6f2447366d6c996127b73eca37a9159",
	}
}

func mutateSyntheticMetadata(t *testing.T, root string, pin *releasePinIdentity, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(ReleaseMetadataPath))
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(payload, &metadata); err != nil {
		t.Fatal(err)
	}
	mutate(metadata)
	payload, err = json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, path, payload)
	pin.ReleaseMetadataSHA256 = digestTestBytes(payload)
}

func writeTestFile(t *testing.T, path string, payload []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func digestTestBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
