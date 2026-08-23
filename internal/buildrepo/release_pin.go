package buildrepo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	// SpecReleaseTag is the published curator-spec release tag represented by
	// this immutable pin. CI checks out SpecReleaseCommit rather than this tag.
	SpecReleaseTag = "v1.0.0-rc.8"
	// SpecReleaseCommit is the immutable Git commit targeted by SpecReleaseTag.
	SpecReleaseCommit = "f8c405aa3ad0a39d260c2ed93684e55c5a346359"
	// SpecReleaseTagObject is the signed annotated tag object's identity.
	SpecReleaseTagObject = "ad247840292487d5d88ac44331798b6b4182a79f"
	// ReleaseMetadataPath is relative to the curator-spec repository root.
	ReleaseMetadataPath = "release/1.0.0-rc.8.json"
	// ReleaseMetadataSHA256 pins the published rc.8 release metadata bytes.
	ReleaseMetadataSHA256 = "293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede"
	// LegacyReleaseMetadataSHA256 preserves the rc.7 historical mapping named
	// by the rc.8 metadata. It must not be silently regenerated as rc.8.
	LegacyReleaseMetadataSHA256 = "e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d"
)

var immutableSpecCommitRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

type releasePinIdentity struct {
	ProtocolVersion           string
	SpecCommit                string
	ConformanceManifestSHA256 string
	ReleaseMetadataSHA256     string
}

var publishedReleasePin = releasePinIdentity{
	ProtocolVersion:           ProtocolVersion,
	SpecCommit:                SpecReleaseCommit,
	ConformanceManifestSHA256: ConformanceManifestSHA256,
	ReleaseMetadataSHA256:     ReleaseMetadataSHA256,
}

type releaseMetadata struct {
	Assurance struct {
		DefaultMode                  string            `json:"default_mode"`
		PortableExecutionPolicy      string            `json:"portable_execution_policy"`
		PortablePolicy               string            `json:"portable_policy"`
		SilentDowngradePermitted     bool              `json:"silent_downgrade_permitted"`
		SkillVendoredProviderAllowed bool              `json:"skill_vendored_provider_allowed"`
		VerifiedExecutionPolicy      string            `json:"verified_execution_policy"`
		VerifiedImplementations      []json.RawMessage `json:"verified_implementations"`
		VerifiedPlatformClaims       []json.RawMessage `json:"verified_platform_claims"`
		VerifiedPolicy               string            `json:"verified_policy"`
		VerifiedProviderContract     string            `json:"verified_provider_contract"`
	} `json:"assurance"`
	CandidateProtocolPin struct {
		ManifestSHA256 string `json:"manifest_sha256"`
		SuiteRoot      string `json:"suite_root"`
	} `json:"candidate_protocol_pin"`
	ClaimV4 struct {
		ClaimProtocolVersion string            `json:"claim_protocol_version"`
		ClaimsEmitted        []json.RawMessage `json:"claims_emitted"`
		Schema               string            `json:"schema"`
	} `json:"claim_v4"`
	CreatedAt             string `json:"created_at"`
	DownstreamConsumption struct {
		CommittedReleasePinAdvanced bool   `json:"committed_release_pin_advanced"`
		Environment                 string `json:"environment"`
		RequiredManifestSHA256      string `json:"required_manifest_sha256"`
	} `json:"downstream_consumption"`
	HistoricalRelease struct {
		Immutable       bool   `json:"immutable"`
		MetadataPath    string `json:"metadata_path"`
		MetadataSHA256  string `json:"metadata_sha256"`
		ProtocolVersion string `json:"protocol_version"`
		SourceCommit    string `json:"source_commit"`
	} `json:"historical_release"`
	LegacyRelease        string `json:"legacy_release"`
	ProtocolVersion      string `json:"protocol_version"`
	SourceBaselineCommit string `json:"source_baseline_commit"`
}

// VerifyReleasePin proves that specRoot contains exactly the published rc.8
// manifest and release metadata at the immutable commit Curator selected.
func VerifyReleasePin(specRoot, revision string) error {
	return verifyReleasePin(specRoot, revision, publishedReleasePin)
}

func verifyReleasePin(specRoot, revision string, pin releasePinIdentity) error {
	if !immutableSpecCommitRE.MatchString(revision) {
		return fmt.Errorf("curator-spec revision %q is mutable or is not a full lowercase SHA-1 commit", revision)
	}
	if revision != pin.SpecCommit {
		return fmt.Errorf("curator-spec revision mismatch: got %s, want %s", revision, pin.SpecCommit)
	}

	manifestPath := filepath.Join(specRoot, "conformance", "v1", "manifest.json")
	manifestBytes, err := readPinnedFile(manifestPath, pin.ConformanceManifestSHA256, "conformance manifest")
	if err != nil {
		return err
	}
	var manifest struct {
		ProtocolVersion string `json:"protocol_version"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("decode conformance manifest: %w", err)
	}
	if manifest.ProtocolVersion != pin.ProtocolVersion {
		return fmt.Errorf("conformance manifest protocol_version = %q, want %q", manifest.ProtocolVersion, pin.ProtocolVersion)
	}

	metadataPath := filepath.Join(specRoot, filepath.FromSlash(ReleaseMetadataPath))
	metadataBytes, err := readPinnedFile(metadataPath, pin.ReleaseMetadataSHA256, "release metadata")
	if err != nil {
		return err
	}
	metadata, err := decodeReleaseMetadata(metadataBytes)
	if err != nil {
		return err
	}
	return validateReleaseMetadata(metadata, pin)
}

func readPinnedFile(path, expectedDigest, label string) ([]byte, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- fixed pin-relative path below the explicitly selected spec root
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	digest := sha256.Sum256(payload)
	got := hex.EncodeToString(digest[:])
	if got != expectedDigest {
		return nil, fmt.Errorf("%s SHA-256 mismatch: got %s, want %s", label, got, expectedDigest)
	}
	return payload, nil
}

func decodeReleaseMetadata(payload []byte) (releaseMetadata, error) {
	var metadata releaseMetadata
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return releaseMetadata{}, fmt.Errorf("decode release metadata: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return releaseMetadata{}, fmt.Errorf("decode release metadata: %w", err)
	}
	return metadata, nil
}

func validateReleaseMetadata(metadata releaseMetadata, pin releasePinIdentity) error {
	wantManifest := "sha256:" + pin.ConformanceManifestSHA256
	if metadata.ProtocolVersion != pin.ProtocolVersion || metadata.ClaimV4.ClaimProtocolVersion != pin.ProtocolVersion {
		return fmt.Errorf("release metadata protocol version mismatch")
	}
	if metadata.CandidateProtocolPin.SuiteRoot != "conformance/v1" || metadata.CandidateProtocolPin.ManifestSHA256 != wantManifest || metadata.DownstreamConsumption.RequiredManifestSHA256 != wantManifest {
		return fmt.Errorf("release metadata conformance-suite identity mismatch")
	}
	if len(metadata.Assurance.VerifiedImplementations) != 0 || len(metadata.Assurance.VerifiedPlatformClaims) != 0 || len(metadata.ClaimV4.ClaimsEmitted) != 0 {
		return fmt.Errorf("release metadata inflates an unpublished implementation, platform, or conformance claim")
	}
	if metadata.Assurance.DefaultMode != "portable" || metadata.Assurance.PortableExecutionPolicy != "manager-worker-v1" || metadata.Assurance.PortablePolicy != "portable-cli-policy-v1" || metadata.Assurance.SilentDowngradePermitted || metadata.Assurance.SkillVendoredProviderAllowed || metadata.Assurance.VerifiedExecutionPolicy != "verified-provider-execution-v1" || metadata.Assurance.VerifiedPolicy != "verified-provider-policy-v1" || metadata.Assurance.VerifiedProviderContract != "host-execution-provider-v1" {
		return fmt.Errorf("release metadata assurance policy mismatch")
	}
	if metadata.ClaimV4.Schema != "schemas/v1/conformance-claim-v4.schema.json" || metadata.DownstreamConsumption.Environment != "CURATOR_CONFORMANCE_ROOT" || metadata.DownstreamConsumption.CommittedReleasePinAdvanced {
		return fmt.Errorf("release metadata downstream or claim contract mismatch")
	}
	if !metadata.HistoricalRelease.Immutable || metadata.HistoricalRelease.MetadataPath != "release/1.0.0-rc.7.json" || metadata.HistoricalRelease.MetadataSHA256 != "sha256:"+LegacyReleaseMetadataSHA256 || metadata.HistoricalRelease.ProtocolVersion != "1.0.0-rc.7" || metadata.HistoricalRelease.SourceCommit != "99f70947d6f2447366d6c996127b73eca37a9159" || metadata.LegacyRelease != "1.0.0-rc.7" || metadata.SourceBaselineCommit != "99f70947d6f2447366d6c996127b73eca37a9159" {
		return fmt.Errorf("release metadata legacy compatibility mapping mismatch")
	}
	return nil
}
