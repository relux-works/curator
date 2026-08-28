package closureexec

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	// PortableBuildSessionReceipt identifies evidence from the authenticated
	// Curator manager/worker build session. It is intentionally distinct from
	// the generic one-shot derivation receipt.
	PortableBuildSessionReceipt = "portable-build-session-receipt-v1"
	// VerifiedBuildSessionReceipt identifies the additive provider-backed build
	// session envelope. The provider's own execution receipt remains separately
	// typed and is referenced by digest.
	VerifiedBuildSessionReceipt = "verified-build-session-receipt-v1"
)

// AssuredBuildCacheInput is the typed production build-cache address. Both
// modes use a namespace distinct from the historical assurance-blind build
// input, and the complete established binding participates in the identity.
type AssuredBuildCacheInput struct {
	BuildInput buildmeta.Input
	Binding    AssuranceBinding
}

// ID returns the exact cache address for the selected assurance strategy.
func (input AssuredBuildCacheInput) ID() (closuregraph.ID, error) {
	if err := input.BuildInput.Validate(); err != nil {
		return "", fmt.Errorf("build input: %w", err)
	}
	if err := input.Binding.Validate(); err != nil {
		return "", err
	}
	logical, err := input.BuildInput.CacheKey()
	if err != nil {
		return "", err
	}
	cacheIdentity := "portable-build-cache-identity-v1"
	if input.Binding.AssuranceMode == AssuranceVerified {
		cacheIdentity = "verified-build-cache-identity-v1"
	}
	return closuregraph.DomainID("curator-assured-build-cache-input-v1", map[string]any{
		"assurance":          input.Binding.CanonicalValue(),
		"build_input_sha256": string(logical),
		"cache_identity":     cacheIdentity,
	})
}

// RuntimeControlEvidence is an honest manager/worker observation. It records
// applied control status without re-labelling it as lossless host observation.
type RuntimeControlEvidence struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Status       string `json:"status"`
}

// BuildSessionReceipt binds one real compiler dispatch to the selected
// authority, logical build input, exact toolchain, and locally verified output.
type BuildSessionReceipt struct {
	ReceiptType              string                   `json:"receipt_type"`
	Binding                  AssuranceBinding         `json:"assurance"`
	BuildInputSHA256         closuregraph.ID          `json:"build_input_sha256"`
	ToolchainSHA256          closuregraph.ID          `json:"toolchain_sha256"`
	ArtifactSHA256           closuregraph.ID          `json:"artifact_sha256"`
	RuntimeControls          []RuntimeControlEvidence `json:"runtime_controls"`
	ProviderExecutionReceipt *closuregraph.ID         `json:"provider_execution_receipt_sha256"`
}

// Validate checks the receipt without trusting the process that returned it.
func (receipt BuildSessionReceipt) Validate() error {
	if err := receipt.Binding.Validate(); err != nil {
		return err
	}
	if !receipt.BuildInputSHA256.Valid() || !receipt.ToolchainSHA256.Valid() || !receipt.ArtifactSHA256.Valid() {
		return failure("verified_execution_receipt_invalid", "build-session receipt identities are malformed")
	}
	switch receipt.Binding.AssuranceMode {
	case AssurancePortable:
		if receipt.ReceiptType != PortableBuildSessionReceipt || receipt.ProviderExecutionReceipt != nil {
			return failure("assurance_evidence_mismatch", "portable build receipt carries verified authority")
		}
	case AssuranceVerified:
		if receipt.ReceiptType != VerifiedBuildSessionReceipt || receipt.ProviderExecutionReceipt == nil || !receipt.ProviderExecutionReceipt.Valid() || len(receipt.RuntimeControls) != 0 {
			return failure("verified_execution_receipt_invalid", "verified build receipt lacks exact provider execution evidence")
		}
	default:
		return failure("execution_mode_unknown", "unknown execution mode %q", receipt.Binding.AssuranceMode)
	}
	seen := map[string]bool{}
	for _, control := range receipt.RuntimeControls {
		if control.Name == "" || control.Availability == "" || control.Status == "" || seen[control.Name] {
			return failure("assurance_evidence_mismatch", "runtime control evidence is malformed or duplicated")
		}
		seen[control.Name] = true
	}
	return nil
}

// ValidateFor proves this receipt belongs to the exact lookup/publication
// authority and locally derived input, toolchain, and artifact.
func (receipt BuildSessionReceipt) ValidateFor(binding AssuranceBinding, input buildmeta.Input, artifact buildmeta.Artifact) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	if !reflect.DeepEqual(receipt.Binding, binding) {
		return failure("assurance_evidence_mismatch", "build-session receipt assurance differs from the operation")
	}
	logical, err := input.CacheKey()
	if err != nil {
		return err
	}
	toolchainID, err := toolchainIdentity(input.Toolchain)
	if err != nil {
		return err
	}
	if receipt.BuildInputSHA256 != closuregraph.ID(logical) || receipt.ToolchainSHA256 != toolchainID || receipt.ArtifactSHA256 != closuregraph.ID(artifact.SHA256) {
		return failure("verified_execution_receipt_invalid", "build-session receipt input, toolchain, or artifact differs")
	}
	return nil
}

// ID returns the typed receipt digest used by protected cache publication.
func (receipt BuildSessionReceipt) ID() (closuregraph.ID, error) {
	if err := receipt.Validate(); err != nil {
		return "", err
	}
	return closuregraph.DomainID("curator-build-session-receipt-v1", receipt.canonicalValue())
}

// CanonicalBytes returns the strict protected-cache sidecar bytes.
func (receipt BuildSessionReceipt) CanonicalBytes() ([]byte, error) {
	if err := receipt.Validate(); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(receipt.canonicalValue())
}

// DecodeBuildSessionReceipt accepts only the exact canonical closed receipt.
func DecodeBuildSessionReceipt(payload []byte) (BuildSessionReceipt, error) {
	var receipt BuildSessionReceipt
	if err := protocoljson.UnmarshalCanonical(payload, &receipt); err != nil {
		return BuildSessionReceipt{}, err
	}
	if err := receipt.Validate(); err != nil {
		return BuildSessionReceipt{}, err
	}
	want, err := receipt.CanonicalBytes()
	if err != nil || !reflect.DeepEqual(want, payload) {
		return BuildSessionReceipt{}, failure("verified_execution_receipt_invalid", "build-session receipt is not exact canonical evidence")
	}
	return receipt, nil
}

func (receipt BuildSessionReceipt) canonicalValue() map[string]any {
	controls := append([]RuntimeControlEvidence(nil), receipt.RuntimeControls...)
	sort.Slice(controls, func(i, j int) bool { return controls[i].Name < controls[j].Name })
	controlValues := make([]any, len(controls))
	for index, control := range controls {
		controlValues[index] = map[string]any{"availability": control.Availability, "name": control.Name, "status": control.Status}
	}
	return map[string]any{
		"artifact_sha256":                   string(receipt.ArtifactSHA256),
		"assurance":                         receipt.Binding.CanonicalValue(),
		"build_input_sha256":                string(receipt.BuildInputSHA256),
		"provider_execution_receipt_sha256": optionalID(receipt.ProviderExecutionReceipt),
		"receipt_type":                      receipt.ReceiptType,
		"runtime_controls":                  controlValues,
		"toolchain_sha256":                  string(receipt.ToolchainSHA256),
	}
}

func toolchainIdentity(toolchain buildmeta.Toolchain) (closuregraph.ID, error) {
	return closuregraph.DomainID("curator-build-toolchain-input-v1", map[string]any{
		"algorithm": toolchain.Algorithm, "content_sha256": toolchain.ContentSHA256,
		"go_relpath": toolchain.GoRelpath, "go_version": toolchain.GoVersion,
	})
}

// NewPortableBuildSessionReceipt constructs honest evidence from the controls
// the manager/worker session itself reported after dispatch.
func NewPortableBuildSessionReceipt(input buildmeta.Input, artifact buildmeta.Artifact, controls []RuntimeControlEvidence) (BuildSessionReceipt, error) {
	logical, err := input.CacheKey()
	if err != nil {
		return BuildSessionReceipt{}, err
	}
	toolchainID, err := toolchainIdentity(input.Toolchain)
	if err != nil {
		return BuildSessionReceipt{}, err
	}
	receipt := BuildSessionReceipt{
		ReceiptType: PortableBuildSessionReceipt, Binding: PortableAssuranceBinding(),
		BuildInputSHA256: closuregraph.ID(logical), ToolchainSHA256: toolchainID,
		ArtifactSHA256: closuregraph.ID(artifact.SHA256), RuntimeControls: append([]RuntimeControlEvidence(nil), controls...),
	}
	return receipt, receipt.ValidateFor(receipt.Binding, input, artifact)
}

// NewVerifiedBuildSessionReceipt constructs the manager-side envelope around a
// provider execution receipt after local input, toolchain, and artifact
// verification. The provider receipt stays separately typed and hash-linked.
func NewVerifiedBuildSessionReceipt(binding AssuranceBinding, input buildmeta.Input, artifact buildmeta.Artifact, providerExecutionReceipt closuregraph.ID) (BuildSessionReceipt, error) {
	logical, err := input.CacheKey()
	if err != nil {
		return BuildSessionReceipt{}, err
	}
	toolchainID, err := toolchainIdentity(input.Toolchain)
	if err != nil {
		return BuildSessionReceipt{}, err
	}
	receipt := BuildSessionReceipt{
		ReceiptType: VerifiedBuildSessionReceipt, Binding: binding,
		BuildInputSHA256: closuregraph.ID(logical), ToolchainSHA256: toolchainID,
		ArtifactSHA256: closuregraph.ID(artifact.SHA256), ProviderExecutionReceipt: &providerExecutionReceipt,
	}
	return receipt, receipt.ValidateFor(binding, input, artifact)
}
