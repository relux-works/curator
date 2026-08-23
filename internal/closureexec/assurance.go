package closureexec

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
)

// AssuranceMode is the closed execution assurance selection.
type AssuranceMode string

const (
	// AssurancePortable selects manager-owned portable controls.
	AssurancePortable AssuranceMode = "portable"
	// AssuranceVerified selects the provider-backed verified contract.
	AssuranceVerified AssuranceMode = "verified"

	// PortablePolicyID is the non-provider assurance policy identity.
	PortablePolicyID = "portable-cli-policy-v1"
	// PortableExecutionPolicyID preserves the historical portable execution identity.
	PortableExecutionPolicyID = "manager-worker-v1"
	// VerifiedPolicyID is the explicit provider-backed assurance identity.
	VerifiedPolicyID = "verified-provider-policy-v1"
	// VerifiedExecutionPolicyID is disjoint from portable execution.
	VerifiedExecutionPolicyID = "verified-provider-execution-v1"
	// VerifiedProviderContractID identifies the platform-neutral provider contract.
	VerifiedProviderContractID = "host-execution-provider-v1"
)

// AssuranceConfig is operator-owned configuration. Provider selection is
// forbidden in portable mode and mandatory in verified mode.
type AssuranceConfig struct {
	Mode                  AssuranceMode
	ProviderID            string
	ProviderVersion       string
	ProviderBinarySHA256  closuregraph.ID
	ProviderTrustEvidence string
}

// DefaultAssuranceConfig returns the non-provider portable policy.
func DefaultAssuranceConfig() AssuranceConfig { return AssuranceConfig{Mode: AssurancePortable} }

func (config AssuranceConfig) validate() error {
	if config.Mode == "" {
		config.Mode = AssurancePortable
	}
	switch config.Mode {
	case AssurancePortable:
		if config.ProviderID != "" || config.ProviderVersion != "" || config.ProviderBinarySHA256 != "" || config.ProviderTrustEvidence != "" {
			return failure("assurance_evidence_mismatch", "portable mode cannot select a provider")
		}
	case AssuranceVerified:
		if config.ProviderID == "" {
			return failure("verified_provider_missing", "verified mode requires an explicit provider")
		}
		if config.ProviderVersion == "" || !config.ProviderBinarySHA256.Valid() || config.ProviderTrustEvidence == "" {
			return failure("verified_provider_identity_invalid", "verified provider policy is incomplete")
		}
	default:
		return failure("execution_mode_unknown", "unknown execution mode %q", config.Mode)
	}
	return nil
}

func (config AssuranceConfig) normalized() AssuranceConfig {
	if config.Mode == "" {
		config.Mode = AssurancePortable
	}
	return config
}

// CapabilityEvidence records only a guarantee actually established for this
// operation. Portable evidence deliberately has no lossless-observation entry.
type CapabilityEvidence struct {
	CapabilityID string `json:"capability_id"`
	Status       string `json:"status"`
}

var portableCapabilities = []CapabilityEvidence{
	{CapabilityID: "immutable-intake-recheck-v1", Status: "established"},
	{CapabilityID: "immediate-toolchain-recheck-v1", Status: "established"},
	{CapabilityID: "declared-output-verification-v1", Status: "established"},
}

var verifiedCapabilities = []string{
	"total-network-denial-v1",
	"read-only-source-and-toolchain-v1",
	"exact-executable-allowlisting-v1",
	"private-build-root-only-writes-v1",
	"hard-aggregate-descendant-resource-bounds-v1",
	"fail-closed-capability-preflight-v1",
}

// ProviderIdentity is platform-neutral trusted provider identity evidence.
type ProviderIdentity struct {
	Contract      string          `json:"provider_contract"`
	ProviderID    string          `json:"provider_id"`
	Version       string          `json:"provider_version"`
	BinarySHA256  closuregraph.ID `json:"provider_binary_sha256"`
	TrustEvidence string          `json:"trust_evidence"`
}

func (identity ProviderIdentity) validate(config AssuranceConfig) error {
	if identity.Contract != VerifiedProviderContractID || identity.ProviderID != config.ProviderID || identity.Version != config.ProviderVersion || identity.BinarySHA256 != config.ProviderBinarySHA256 || identity.TrustEvidence != config.ProviderTrustEvidence {
		return failure("verified_provider_identity_invalid", "provider identity or trust evidence differs from policy")
	}
	return nil
}

// ProviderCapabilityReceipt is fresh health and exact-capability evidence.
type ProviderCapabilityReceipt struct {
	Provider     ProviderIdentity     `json:"provider"`
	Health       string               `json:"health"`
	Capabilities []CapabilityEvidence `json:"capabilities"`
	Nonce        string               `json:"nonce"`
	ObservedAt   time.Time            `json:"observed_at"`
	ExpiresAt    time.Time            `json:"expires_at"`
}

func (receipt ProviderCapabilityReceipt) validate(expected ProviderIdentity, nonce string, now time.Time) error {
	if receipt.Provider != expected || receipt.Health != "healthy" || receipt.Nonce != nonce || !receipt.ObservedAt.Before(receipt.ExpiresAt) || now.Before(receipt.ObservedAt) || !now.Before(receipt.ExpiresAt) {
		return failure("verified_provider_unavailable", "provider health receipt is missing, stale, or mismatched")
	}
	if len(receipt.Capabilities) != len(verifiedCapabilities) {
		return failure("verified_capabilities_unsatisfied", "provider capability set is incomplete")
	}
	for index, capability := range receipt.Capabilities {
		if capability.CapabilityID != verifiedCapabilities[index] || capability.Status != "established" {
			return failure("verified_capabilities_unsatisfied", "provider capability set differs at index %d", index)
		}
	}
	return nil
}

// ID derives the exact typed provider capability receipt identity.
func (receipt ProviderCapabilityReceipt) ID() (closuregraph.ID, error) {
	capabilities := make([]any, len(receipt.Capabilities))
	for index, capability := range receipt.Capabilities {
		capabilities[index] = map[string]any{"capability_id": capability.CapabilityID, "status": capability.Status}
	}
	return closuregraph.DomainID("provider-capability-receipt-v1", map[string]any{
		"capabilities": capabilities, "expires_at": receipt.ExpiresAt.UTC().Format(time.RFC3339Nano), "health": receipt.Health,
		"nonce": receipt.Nonce, "observed_at": receipt.ObservedAt.UTC().Format(time.RFC3339Nano), "provider": providerValue(receipt.Provider),
	})
}

// VerifiedProvider is the future platform implementation seam. It is neutral
// to macOS, Linux, and Windows mechanisms.
type VerifiedProvider interface {
	EnforceObserveProvider
	Identity() ProviderIdentity
	Negotiate(context.Context, string) (ProviderCapabilityReceipt, error)
}

// PortableRunResult exposes exit status and a manager-owned output root. It
// contains no process, read, write, or network observation claims.
type PortableRunResult struct {
	ExitCode   int64
	OutputRoot string
	// EvidenceRoot is the manager-owned root from which exact declared
	// evidence is hashed. It may be the execution root when a build's final
	// artifact lives beside authorized intermediate work copies.
	EvidenceRoot string
	cleanup      func()
}

// Release removes non-retained replay state after a provider has consumed the
// portable runner result. The shared portable executor calls the same cleanup
// internally; verified test providers that wrap the concrete launch seam use
// this method explicitly.
func (result *PortableRunResult) Release() {
	if result != nil && result.cleanup != nil {
		result.cleanup()
		result.cleanup = nil
	}
}

// PortableProcessRunner is the manager-owned CLI process seam.
type PortableProcessRunner interface {
	Run(context.Context, ExecutionRequest) (PortableRunResult, error)
}

// ProcessStartObserver is test and audit instrumentation at the sole
// manager-owned portable process-start seam. The supplied permit has already
// passed assurance preflight, been committed, and survived all immediate
// toolchain and admitted-input rechecks.
type ProcessStartObserver func(DerivationPermit)

// ProcessLaunch is the exact command handed to the operating-system process
// start seam. It is instrumentation, not portable lossless evidence.
type ProcessLaunch struct {
	Executable  string
	CWD         string
	Argv        []string
	Environment []string
}

// ProcessLaunchObserver receives the concrete command immediately before the
// process start. Verified test providers use it to prove the observed launch
// boundary rather than reconstructing process evidence from a permit.
type ProcessLaunchObserver func(ProcessLaunch)

// PreflightAssurance is the production selection boundary used before an
// operation may inspect execution caches or start a child. Portable selection
// returns only its static manager-owned capability record. Verified selection
// resolves and negotiates the configured provider with no fallback.
func PreflightAssurance(ctx context.Context, config AssuranceConfig, provider VerifiedProvider) (AssuranceBinding, error) {
	config = config.normalized()
	if err := config.validate(); err != nil {
		return AssuranceBinding{}, err
	}
	if config.Mode == AssurancePortable {
		if provider != nil {
			return AssuranceBinding{}, failure("assurance_evidence_mismatch", "portable mode cannot silently use a verified provider")
		}
		return portableAssuranceBinding(), nil
	}
	executor, err := NewAssuredExecutor(config, nil, provider, "cli-assurance-preflight-v1")
	if err != nil {
		return AssuranceBinding{}, err
	}
	operation, err := executor.Preflight(ctx)
	if err != nil {
		return AssuranceBinding{}, err
	}
	return operation.binding, nil
}

func freshNonce() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate provider nonce: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func providerValue(identity ProviderIdentity) map[string]any {
	return map[string]any{"provider_binary_sha256": string(identity.BinarySHA256), "provider_contract": identity.Contract, "provider_id": identity.ProviderID, "provider_version": identity.Version, "trust_evidence": identity.TrustEvidence}
}

func capabilitiesValue(capabilities []CapabilityEvidence) []any {
	values := make([]any, len(capabilities))
	for index, capability := range capabilities {
		values[index] = map[string]any{"capability_id": capability.CapabilityID, "status": capability.Status}
	}
	return values
}

// AssuranceBinding is the exact operation assurance established before any
// cache lookup or process dispatch. Verified bindings contain one fresh
// provider capability receipt; portable bindings contain no provider claims.
type AssuranceBinding struct {
	AssuranceMode       AssuranceMode        `json:"assurance_mode"`
	PolicyID            string               `json:"policy_id"`
	ExecutionPolicyID   string               `json:"execution_policy"`
	ProviderContract    *string              `json:"provider_contract"`
	Provider            *ProviderIdentity    `json:"provider"`
	CapabilityReceiptID *closuregraph.ID     `json:"capability_receipt_sha256"`
	ActualCapabilities  []CapabilityEvidence `json:"actual_capabilities"`
}

func (binding AssuranceBinding) validate() error {
	return validateAssuranceBinding(binding.AssuranceMode, binding.PolicyID, binding.ExecutionPolicyID, binding.ProviderContract, binding.Provider, binding.CapabilityReceiptID, binding.ActualCapabilities)
}

// Validate checks the complete closed assurance binding. It is exported for
// additive build-session integrations; the one-shot derivation contract keeps
// using the same validator.
func (binding AssuranceBinding) Validate() error { return binding.validate() }

func (binding AssuranceBinding) canonicalValue() map[string]any {
	return map[string]any{
		"actual_capabilities":       capabilitiesValue(binding.ActualCapabilities),
		"assurance_mode":            string(binding.AssuranceMode),
		"capability_receipt_sha256": optionalID(binding.CapabilityReceiptID),
		"execution_policy":          binding.ExecutionPolicyID,
		"policy_id":                 binding.PolicyID,
		"provider":                  optionalProvider(binding.Provider),
		"provider_contract":         optionalString(binding.ProviderContract),
	}
}

// CanonicalValue returns a detached canonical object for cache and receipt
// envelopes owned by other manager subsystems.
func (binding AssuranceBinding) CanonicalValue() map[string]any {
	value := binding.canonicalValue()
	return value
}

func portableAssuranceBinding() AssuranceBinding {
	return AssuranceBinding{
		AssuranceMode: AssurancePortable, PolicyID: PortablePolicyID,
		ExecutionPolicyID:  PortableExecutionPolicyID,
		ActualCapabilities: append([]CapabilityEvidence(nil), portableCapabilities...),
	}
}

// PortableAssuranceBinding returns the exact non-provider default binding.
func PortableAssuranceBinding() AssuranceBinding { return portableAssuranceBinding() }

// AssuredCacheInput is the only protected-cache address accepted by the
// closure store. It keeps portable and verified entries in disjoint typed
// namespaces and binds exact provider and fresh capability evidence.
type AssuredCacheInput struct {
	Expected closuregraph.ExpectedCacheInput
	Binding  AssuranceBinding
}

// ID derives the protected entry key from the independent closure input and
// the complete assurance binding.
func (input AssuredCacheInput) ID() (closuregraph.ID, error) {
	if err := input.Expected.Validate(); err != nil {
		return "", err
	}
	if err := input.Binding.validate(); err != nil {
		return "", err
	}
	expectedID, err := input.Expected.ID()
	if err != nil {
		return "", err
	}
	value := input.Binding.canonicalValue()
	value["cache_identity"] = "portable-cache-identity-v1"
	if input.Binding.AssuranceMode == AssuranceVerified {
		value["cache_identity"] = "verified-cache-identity-v1"
	}
	value["build_input_sha256"] = string(expectedID)
	return closuregraph.DomainID("curator-assured-cache-input-v1", value)
}

func (input AssuredCacheInput) canonicalValue() (map[string]any, error) {
	id, err := input.ID()
	if err != nil {
		return nil, err
	}
	expectedID, _ := input.Expected.ID()
	return map[string]any{
		"assurance":               input.Binding.canonicalValue(),
		"assured_cache_input_id":  string(id),
		"expected_cache_input_id": string(expectedID),
	}, nil
}

// ValidateFor rejects cross-mode and cross-provider receipt reuse.
func (receipt DerivationReceipt) ValidateFor(config AssuranceConfig, provider *ProviderIdentity) error {
	config = config.normalized()
	if err := config.validate(); err != nil {
		return err
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	if receipt.AssuranceMode != config.Mode {
		return failure("assurance_evidence_mismatch", "receipt mode differs from required policy")
	}
	if config.Mode == AssurancePortable {
		if provider != nil {
			return failure("assurance_evidence_mismatch", "portable receipt cannot carry provider authority")
		}
		return nil
	}
	if provider == nil || receipt.Provider == nil || *provider != *receipt.Provider || provider.validate(config) != nil {
		return failure("verified_execution_receipt_invalid", "receipt provider differs from required provider")
	}
	return nil
}

// ExecutionCheckpointIdentity is recovery identity, never cache authority.
// Mode and exact provider/capability bindings prevent namespace aliasing.
type ExecutionCheckpointIdentity struct {
	AssuranceMode       AssuranceMode
	PolicyID            string
	ExecutionPolicyID   string
	ProviderContract    *string
	Provider            *ProviderIdentity
	CapabilityReceiptID *closuregraph.ID
	ActualCapabilities  []CapabilityEvidence
	OperationID         closuregraph.ID
}

// ID returns a checkpoint-only domain identity.
func (checkpoint ExecutionCheckpointIdentity) ID() (closuregraph.ID, error) {
	if !checkpoint.OperationID.Valid() {
		return "", failure("verified_checkpoint_invalid", "checkpoint operation identity is invalid")
	}
	if err := validateAssuranceBinding(checkpoint.AssuranceMode, checkpoint.PolicyID, checkpoint.ExecutionPolicyID, checkpoint.ProviderContract, checkpoint.Provider, checkpoint.CapabilityReceiptID, checkpoint.ActualCapabilities); err != nil {
		return "", err
	}
	return closuregraph.DomainID("curator-assurance-checkpoint-identity-v1", map[string]any{
		"actual_capabilities": capabilitiesValue(checkpoint.ActualCapabilities), "assurance_mode": string(checkpoint.AssuranceMode),
		"capability_receipt_sha256": optionalID(checkpoint.CapabilityReceiptID), "execution_policy": checkpoint.ExecutionPolicyID,
		"operation_id": string(checkpoint.OperationID), "policy_id": checkpoint.PolicyID, "provider": optionalProvider(checkpoint.Provider),
		"provider_contract": optionalString(checkpoint.ProviderContract), "record_type": "assurance-checkpoint-identity-v1",
	})
}
