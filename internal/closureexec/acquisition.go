package closureexec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	// SchemaSourceAcquisitionPermit identifies manager-issued, network-capable
	// source acquisition authority. It is intentionally disjoint from the
	// network-none derivation permit.
	SchemaSourceAcquisitionPermit = "source-acquisition-permit-v1"
	// SchemaSourceAcquisitionReceipt identifies issued acquisition evidence.
	SchemaSourceAcquisitionReceipt = "source-acquisition-receipt-v1"
)

// SourceAcquisitionPermit commits the complete C2 broker boundary before a
// source-control process starts. It authorizes one exact origin and immutable
// requested revision, and writes only below one task-private quarantine root.
type SourceAcquisitionPermit struct {
	SchemaID             string                `json:"schema_id"`
	AssuranceMode        AssuranceMode         `json:"assurance_mode"`
	PolicyID             string                `json:"policy_id"`
	ExecutionPolicyID    string                `json:"execution_policy"`
	ProviderContract     *string               `json:"provider_contract"`
	Provider             *ProviderIdentity     `json:"provider"`
	CapabilityReceiptID  *closuregraph.ID      `json:"capability_receipt_sha256"`
	ActualCapabilities   []CapabilityEvidence  `json:"actual_capabilities"`
	PreviousCausalHead   string                `json:"previous_causal_head"`
	SourceProfileID      string                `json:"source_profile_id"`
	CanonicalOrigin      string                `json:"canonical_origin"`
	RequestedRevision    string                `json:"requested_revision"`
	C0CheckpointID       closuregraph.ID       `json:"c0_checkpoint_id"`
	ToolchainNodeID      closuregraph.ID       `json:"toolchain_node_id"`
	ToolchainFingerprint closuregraph.ID       `json:"toolchain_fingerprint"`
	ExecutableSHA256     closuregraph.ID       `json:"executable_sha256"`
	Executable           string                `json:"executable"`
	Argv                 []string              `json:"argv"`
	CWD                  string                `json:"cwd"`
	Environment          map[string]string     `json:"environment"`
	HostID               closuregraph.ID       `json:"host_id"`
	TargetID             closuregraph.ID       `json:"target_id"`
	AllowedProcesses     []string              `json:"allowed_processes"`
	ReadRoots            []string              `json:"read_roots"`
	QuarantineWriteRoots []string              `json:"quarantine_write_roots"`
	NetworkPolicy        string                `json:"network_policy"`
	ExpectedEvidence     []EvidenceRequirement `json:"expected_evidence"`
	StdoutEvidencePath   string                `json:"stdout_evidence_path"`
	ResourceLimits       ResourceLimits        `json:"resource_limits"`
	ResourceLimitID      closuregraph.ID       `json:"resource_limit_id"`
	EvidenceSchemaID     closuregraph.ID       `json:"evidence_schema_id"`
	RecheckRule          string                `json:"recheck_rule"`
}

// Validate checks the closed acquisition contract without pretending that
// manager-only portable execution can observe host network behavior.
func (permit SourceAcquisitionPermit) Validate() error {
	if permit.SchemaID != SchemaSourceAcquisitionPermit || permit.PreviousCausalHead == "" ||
		permit.SourceProfileID == "" || permit.CanonicalOrigin == "" || permit.RequestedRevision == "" ||
		!permit.C0CheckpointID.Valid() || !permit.ToolchainNodeID.Valid() ||
		!permit.ToolchainFingerprint.Valid() || !permit.ExecutableSHA256.Valid() ||
		!permit.HostID.Valid() || !permit.TargetID.Valid() || permit.RecheckRule != "immediate-exact-v1" {
		return failure("closure_derivation_unauthorized", "source acquisition authority is incomplete")
	}
	if strings.ContainsAny(permit.SourceProfileID+permit.CanonicalOrigin+permit.RequestedRevision, "\x00\r\n") {
		return failure("closure_derivation_unauthorized", "source acquisition identity contains control bytes")
	}
	if permit.NetworkPolicy != "exact-origin-only" {
		return failure("closure_derivation_unauthorized", "source acquisition network policy is not exact-origin-only")
	}
	if err := validateAssuranceBinding(permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID, permit.ProviderContract, permit.Provider, permit.CapabilityReceiptID, permit.ActualCapabilities); err != nil {
		return err
	}
	if err := portablePath(permit.Executable); err != nil {
		return failure("closure_derivation_unauthorized", "source acquisition executable: %v", err)
	}
	if err := portablePath(permit.CWD); err != nil {
		return failure("closure_derivation_unauthorized", "source acquisition cwd: %v", err)
	}
	for _, values := range [][]string{permit.AllowedProcesses, permit.ReadRoots, permit.QuarantineWriteRoots} {
		if len(values) == 0 || !sort.StringsAreSorted(values) || hasDuplicates(values) {
			return failure("closure_derivation_unauthorized", "source acquisition policy sets must be nonempty, sorted, and unique")
		}
		for _, value := range values {
			if err := portablePath(value); err != nil {
				return failure("closure_derivation_unauthorized", "source acquisition policy path is invalid")
			}
		}
	}
	if !containsString(permit.AllowedProcesses, permit.Executable) {
		return failure("closure_derivation_unauthorized", "source acquisition executable is outside its process family")
	}
	if err := validateEnvironment(permit.Environment); err != nil {
		return err
	}
	limitID, err := permit.ResourceLimits.ID()
	if err != nil || limitID != permit.ResourceLimitID {
		return failure("closure_derivation_unauthorized", "source acquisition resource-limit identity differs")
	}
	evidenceID, err := evidenceSchemaID(permit.ExpectedEvidence)
	if err != nil || evidenceID != permit.EvidenceSchemaID {
		return failure("closure_derivation_unauthorized", "source acquisition evidence schema differs")
	}
	if permit.StdoutEvidencePath != "" {
		if err = portablePath(permit.StdoutEvidencePath); err != nil || len(permit.ExpectedEvidence) != 1 || permit.ExpectedEvidence[0].Path != permit.StdoutEvidencePath {
			return failure("closure_derivation_unauthorized", "source acquisition stdout evidence is not exactly declared")
		}
	}
	return nil
}

func (permit SourceAcquisitionPermit) canonicalValue() map[string]any {
	evidence := make([]any, len(permit.ExpectedEvidence))
	for index := range permit.ExpectedEvidence {
		evidence[index] = permit.ExpectedEvidence[index].canonicalValue()
	}
	return map[string]any{
		"actual_capabilities": capabilitiesValue(permit.ActualCapabilities), "allowed_processes": stringsAny(permit.AllowedProcesses),
		"argv": stringsAny(permit.Argv), "assurance_mode": string(permit.AssuranceMode), "c0_checkpoint_id": string(permit.C0CheckpointID),
		"canonical_origin": permit.CanonicalOrigin, "capability_receipt_sha256": optionalID(permit.CapabilityReceiptID), "cwd": permit.CWD,
		"environment": stringMap(permit.Environment), "evidence_schema_id": string(permit.EvidenceSchemaID), "executable": permit.Executable,
		"executable_sha256": string(permit.ExecutableSHA256), "execution_policy": permit.ExecutionPolicyID, "expected_evidence": evidence,
		"host_id": string(permit.HostID), "network_policy": permit.NetworkPolicy, "policy_id": permit.PolicyID,
		"previous_causal_head": permit.PreviousCausalHead, "provider": optionalProvider(permit.Provider), "provider_contract": optionalString(permit.ProviderContract),
		"quarantine_write_roots": stringsAny(permit.QuarantineWriteRoots), "read_roots": stringsAny(permit.ReadRoots), "recheck_rule": permit.RecheckRule,
		"requested_revision": permit.RequestedRevision, "resource_limit_id": string(permit.ResourceLimitID), "resource_limits": permit.ResourceLimits.canonicalValue(),
		"schema_id": permit.SchemaID, "source_profile_id": permit.SourceProfileID, "stdout_evidence_path": permit.StdoutEvidencePath, "target_id": string(permit.TargetID),
		"toolchain_fingerprint": string(permit.ToolchainFingerprint), "toolchain_node_id": string(permit.ToolchainNodeID),
	}
}

// CanonicalBytes returns exact CCJ bytes.
func (permit SourceAcquisitionPermit) CanonicalBytes() ([]byte, error) {
	return canonical(permit.Validate, permit.canonicalValue())
}

// ID derives the domain-separated permit identity.
func (permit SourceAcquisitionPermit) ID() (closuregraph.ID, error) {
	payload, err := permit.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return closuregraph.IDFromCanonical("source-acquisition-permit-v1", payload)
}

// DecodeSourceAcquisitionPermit rejects unknown fields and noncanonical bytes.
func DecodeSourceAcquisitionPermit(payload []byte) (SourceAcquisitionPermit, error) {
	var permit SourceAcquisitionPermit
	if err := protocoljson.UnmarshalCanonical(payload, &permit); err != nil {
		return SourceAcquisitionPermit{}, err
	}
	roundTrip, err := permit.CanonicalBytes()
	if err != nil || !bytes.Equal(roundTrip, payload) {
		return SourceAcquisitionPermit{}, failure("closure_derivation_unauthorized", "source acquisition permit canonical round-trip differs")
	}
	return permit, nil
}

// SourceAcquisitionObservation is manager/provider evidence for one command.
type SourceAcquisitionObservation struct {
	Executable       string             `json:"executable"`
	Argv             []string           `json:"argv"`
	CWD              string             `json:"cwd"`
	Environment      map[string]string  `json:"environment"`
	Processes        []string           `json:"processes"`
	Reads            []string           `json:"reads"`
	Writes           []string           `json:"writes"`
	Network          string             `json:"network"`
	ExitCode         int                `json:"exit_code"`
	Evidence         []DerivationOutput `json:"evidence"`
	ResolvedRevision string             `json:"resolved_revision"`
	GitTree          string             `json:"git_tree"`
	ObjectIDs        []string           `json:"object_ids"`
}

func (observation SourceAcquisitionObservation) canonicalValue() map[string]any {
	evidence := make([]any, len(observation.Evidence))
	for index := range observation.Evidence {
		evidence[index] = observation.Evidence[index].canonicalValue()
	}
	return map[string]any{"argv": stringsAny(observation.Argv), "cwd": observation.CWD, "environment": stringMap(observation.Environment), "evidence": evidence, "executable": observation.Executable, "exit_code": observation.ExitCode, "git_tree": observation.GitTree, "network": observation.Network, "object_ids": stringsAny(observation.ObjectIDs), "processes": stringsAny(observation.Processes), "reads": stringsAny(observation.Reads), "resolved_revision": observation.ResolvedRevision, "writes": stringsAny(observation.Writes)}
}

// SourceAcquisitionReceipt is the sole issued acquisition evidence authority.
type SourceAcquisitionReceipt struct {
	SchemaID            string                       `json:"schema_id"`
	AssuranceMode       AssuranceMode                `json:"assurance_mode"`
	PolicyID            string                       `json:"policy_id"`
	ExecutionPolicyID   string                       `json:"execution_policy"`
	ProviderContract    *string                      `json:"provider_contract"`
	Provider            *ProviderIdentity            `json:"provider"`
	CapabilityReceiptID *closuregraph.ID             `json:"capability_receipt_sha256"`
	ActualCapabilities  []CapabilityEvidence         `json:"actual_capabilities"`
	PermitID            closuregraph.ID              `json:"permit_id"`
	BeforeFingerprint   closuregraph.ID              `json:"before_fingerprint"`
	AfterFingerprint    closuregraph.ID              `json:"after_fingerprint"`
	CanonicalOrigin     string                       `json:"canonical_origin"`
	RequestedRevision   string                       `json:"requested_revision"`
	Observation         SourceAcquisitionObservation `json:"observation"`
	ResourceLimits      ResourceLimits               `json:"resource_limits"`
	ResourceLimitID     closuregraph.ID              `json:"resource_limit_id"`
	ExpectedEvidence    []EvidenceRequirement        `json:"expected_evidence"`
	EvidenceSchemaID    closuregraph.ID              `json:"evidence_schema_id"`
	Diagnostics         []DerivationDiagnostic       `json:"diagnostics"`
	Decision            string                       `json:"decision"`
	NextCausalHead      closuregraph.ID              `json:"next_causal_head"`
}

// Validate checks receipt identity and honest assurance claims.
func (receipt SourceAcquisitionReceipt) Validate() error {
	if receipt.SchemaID != SchemaSourceAcquisitionReceipt || receipt.Decision != "success" || !receipt.PermitID.Valid() || !receipt.BeforeFingerprint.Valid() || receipt.BeforeFingerprint != receipt.AfterFingerprint || !receipt.NextCausalHead.Valid() {
		return failure("closure_derivation_drift", "source acquisition receipt is incomplete")
	}
	if err := validateAssuranceBinding(receipt.AssuranceMode, receipt.PolicyID, receipt.ExecutionPolicyID, receipt.ProviderContract, receipt.Provider, receipt.CapabilityReceiptID, receipt.ActualCapabilities); err != nil {
		return err
	}
	if receipt.AssuranceMode == AssurancePortable && receipt.Observation.Network != "not-observed" {
		return failure("assurance_evidence_mismatch", "portable acquisition claimed unobserved network enforcement")
	}
	if receipt.AssuranceMode == AssuranceVerified && receipt.Observation.Network != "exact-origin-only" {
		return failure("closure_network_attempted", "verified acquisition did not prove exact-origin network")
	}
	if receipt.CanonicalOrigin == "" || receipt.RequestedRevision == "" || receipt.Observation.ExitCode != 0 {
		return failure("closure_derivation_drift", "source acquisition result is unsuccessful")
	}
	limitID, err := receipt.ResourceLimits.ID()
	if err != nil || limitID != receipt.ResourceLimitID {
		return failure("closure_derivation_drift", "source acquisition resource limits drifted")
	}
	evidenceID, err := evidenceSchemaID(receipt.ExpectedEvidence)
	if err != nil || evidenceID != receipt.EvidenceSchemaID || len(receipt.Observation.Evidence) != len(receipt.ExpectedEvidence) {
		return failure("closure_derivation_drift", "source acquisition evidence drifted")
	}
	for index := range receipt.ExpectedEvidence {
		expected, actual := receipt.ExpectedEvidence[index], receipt.Observation.Evidence[index]
		if expected.Path != actual.Path || expected.SchemaID != actual.SchemaID || expected.ArtifactManifestID != actual.ArtifactManifestID || actual.validate() != nil {
			return failure("closure_derivation_drift", "source acquisition evidence record differs")
		}
	}
	return nil
}

func (receipt SourceAcquisitionReceipt) canonicalValue() map[string]any {
	evidence := make([]any, len(receipt.ExpectedEvidence))
	for index := range receipt.ExpectedEvidence {
		evidence[index] = receipt.ExpectedEvidence[index].canonicalValue()
	}
	diagnostics := make([]any, len(receipt.Diagnostics))
	for index := range receipt.Diagnostics {
		diagnostics[index] = receipt.Diagnostics[index].canonicalValue()
	}
	return map[string]any{"actual_capabilities": capabilitiesValue(receipt.ActualCapabilities), "after_fingerprint": string(receipt.AfterFingerprint), "assurance_mode": string(receipt.AssuranceMode), "before_fingerprint": string(receipt.BeforeFingerprint), "canonical_origin": receipt.CanonicalOrigin, "capability_receipt_sha256": optionalID(receipt.CapabilityReceiptID), "decision": receipt.Decision, "diagnostics": diagnostics, "evidence_schema_id": string(receipt.EvidenceSchemaID), "execution_policy": receipt.ExecutionPolicyID, "expected_evidence": evidence, "next_causal_head": string(receipt.NextCausalHead), "observation": receipt.Observation.canonicalValue(), "permit_id": string(receipt.PermitID), "policy_id": receipt.PolicyID, "provider": optionalProvider(receipt.Provider), "provider_contract": optionalString(receipt.ProviderContract), "requested_revision": receipt.RequestedRevision, "resource_limit_id": string(receipt.ResourceLimitID), "resource_limits": receipt.ResourceLimits.canonicalValue(), "schema_id": receipt.SchemaID}
}

// CanonicalBytes returns exact CCJ bytes.
func (receipt SourceAcquisitionReceipt) CanonicalBytes() ([]byte, error) {
	return canonical(receipt.Validate, receipt.canonicalValue())
}

// ID derives the domain-separated receipt identity.
func (receipt SourceAcquisitionReceipt) ID() (closuregraph.ID, error) {
	payload, err := receipt.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return closuregraph.IDFromCanonical("source-acquisition-receipt-v1", payload)
}

// DecodeSourceAcquisitionReceipt rejects unknown fields and noncanonical bytes.
func DecodeSourceAcquisitionReceipt(payload []byte) (SourceAcquisitionReceipt, error) {
	var receipt SourceAcquisitionReceipt
	if err := protocoljson.UnmarshalCanonical(payload, &receipt); err != nil {
		return SourceAcquisitionReceipt{}, err
	}
	roundTrip, err := receipt.CanonicalBytes()
	if err != nil || !bytes.Equal(roundTrip, payload) {
		return SourceAcquisitionReceipt{}, failure("closure_derivation_drift", "source acquisition receipt canonical round-trip differs")
	}
	return receipt, nil
}

// AcquisitionRunResult contains only manager-established portable evidence.
type AcquisitionRunResult struct {
	ExitCode         int
	Evidence         []DerivationOutput
	ResolvedRevision string
	GitTree          string
	ObjectIDs        []string
}

// SourceAcquisitionRunner owns the portable process seam.
type SourceAcquisitionRunner interface {
	RunSourceAcquisition(context.Context, SourceAcquisitionPermit) (AcquisitionRunResult, error)
}

// VerifiedSourceAcquisitionProvider is the lossless provider contract for
// explicit verified acquisition. There is no portable fallback.
type VerifiedSourceAcquisitionProvider interface {
	VerifiedProvider
	EnforceAndObserveSourceAcquisition(context.Context, SourceAcquisitionPermit) (SourceAcquisitionObservation, error)
}

// SourceAcquisitionExecutor serializes committed acquisition authority and
// owns its causal receipt journal.
type SourceAcquisitionExecutor struct {
	mu        sync.Mutex
	assurance AssuranceConfig
	runner    SourceAcquisitionRunner
	provider  VerifiedSourceAcquisitionProvider
	head      string
	committed map[closuregraph.ID]SourceAcquisitionPermit
	receipts  map[closuregraph.ID]SourceAcquisitionReceipt
}

// NewSourceAcquisitionExecutor creates an acquisition plane disjoint from the
// offline derivation executor.
func NewSourceAcquisitionExecutor(config AssuranceConfig, runner SourceAcquisitionRunner, provider VerifiedSourceAcquisitionProvider, initialCausalHead string) (*SourceAcquisitionExecutor, error) {
	config = config.normalized()
	if err := config.validate(); err != nil || initialCausalHead == "" {
		if err != nil {
			return nil, err
		}
		return nil, failure("closure_derivation_unauthorized", "source acquisition causal head is empty")
	}
	if config.Mode == AssurancePortable && (runner == nil || provider != nil) {
		return nil, failure("assurance_evidence_mismatch", "portable acquisition requires only a manager runner")
	}
	if config.Mode == AssuranceVerified && (provider == nil || !provider.LosslessObservation()) {
		return nil, failure("verified_provider_unavailable", "verified source acquisition provider is unavailable")
	}
	return &SourceAcquisitionExecutor{assurance: config, runner: runner, provider: provider, head: initialCausalHead, committed: map[closuregraph.ID]SourceAcquisitionPermit{}, receipts: map[closuregraph.ID]SourceAcquisitionReceipt{}}, nil
}

// CurrentCausalHead returns the next acquisition predecessor.
func (executor *SourceAcquisitionExecutor) CurrentCausalHead() string {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	return executor.head
}

// Commit binds assurance and records authority before execution.
func (executor *SourceAcquisitionExecutor) Commit(ctx context.Context, permit SourceAcquisitionPermit) (closuregraph.ID, error) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if permit.AssuranceMode != "" || permit.PolicyID != "" || permit.ExecutionPolicyID != "" || permit.Provider != nil || permit.ProviderContract != nil || permit.CapabilityReceiptID != nil || permit.ActualCapabilities != nil {
		return "", failure("assurance_evidence_mismatch", "caller supplied acquisition assurance evidence")
	}
	if executor.assurance.Mode == AssurancePortable {
		permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID = AssurancePortable, PortablePolicyID, PortableExecutionPolicyID
		permit.ActualCapabilities = append([]CapabilityEvidence(nil), portableCapabilities...)
	} else {
		identity := executor.provider.Identity()
		nonce, err := freshNonce()
		if err != nil {
			return "", err
		}
		capability, err := executor.provider.Negotiate(ctx, nonce)
		if err != nil {
			return "", failure("verified_provider_unavailable", "provider negotiation failed: %v", err)
		}
		if err = capability.validate(identity, nonce, timeNow()); err != nil {
			return "", err
		}
		capabilityID, _ := capability.ID()
		contract := VerifiedProviderContractID
		permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID = AssuranceVerified, VerifiedPolicyID, VerifiedExecutionPolicyID
		permit.ProviderContract, permit.Provider, permit.CapabilityReceiptID = &contract, &identity, &capabilityID
		permit.ActualCapabilities = append([]CapabilityEvidence(nil), capability.Capabilities...)
	}
	if permit.PreviousCausalHead != executor.head {
		return "", failure("closure_derivation_unauthorized", "source acquisition predecessor is stale")
	}
	if err := permit.Validate(); err != nil {
		return "", err
	}
	id, err := permit.ID()
	if err != nil {
		return "", err
	}
	executor.committed[id] = permit
	return id, nil
}

// Execute consumes one permit and issues its exact receipt. Rechecks happen
// before and after the sole runner/provider process seam while holding the
// causal lock, so stale competitors start zero processes.
func (executor *SourceAcquisitionExecutor) Execute(ctx context.Context, permitID closuregraph.ID, recheck func(context.Context) (ToolchainIdentity, error)) (SourceAcquisitionReceipt, error) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	permit, ok := executor.committed[permitID]
	if !ok {
		return SourceAcquisitionReceipt{}, failure("closure_derivation_unauthorized", "source acquisition permit was not committed")
	}
	delete(executor.committed, permitID)
	if permit.PreviousCausalHead != executor.head || recheck == nil {
		return SourceAcquisitionReceipt{}, failure("closure_derivation_unauthorized", "source acquisition authority changed before start")
	}
	before, err := recheck(ctx)
	if err != nil {
		return SourceAcquisitionReceipt{}, err
	}
	if before.Fingerprint != permit.ToolchainFingerprint || before.ExecutableSHA256 != permit.ExecutableSHA256 {
		return SourceAcquisitionReceipt{}, failure("artifact_toolchain_identity_changed", "source acquisition tool changed before start")
	}
	var observation SourceAcquisitionObservation
	if permit.AssuranceMode == AssurancePortable {
		result, runErr := executor.runner.RunSourceAcquisition(ctx, permit)
		if runErr != nil {
			return SourceAcquisitionReceipt{}, runErr
		}
		observation = SourceAcquisitionObservation{Executable: permit.Executable, Argv: append([]string(nil), permit.Argv...), CWD: permit.CWD, Environment: cloneStringMap(permit.Environment), Processes: []string{}, Reads: []string{}, Writes: []string{}, Network: "not-observed", ExitCode: result.ExitCode, Evidence: append([]DerivationOutput(nil), result.Evidence...), ResolvedRevision: result.ResolvedRevision, GitTree: result.GitTree, ObjectIDs: append([]string(nil), result.ObjectIDs...)}
	} else {
		observation, err = executor.provider.EnforceAndObserveSourceAcquisition(ctx, permit)
		if err != nil {
			return SourceAcquisitionReceipt{}, err
		}
	}
	after, err := recheck(ctx)
	if err != nil {
		return SourceAcquisitionReceipt{}, err
	}
	if before != after {
		return SourceAcquisitionReceipt{}, failure("artifact_toolchain_identity_changed", "source acquisition tool changed during execution")
	}
	if observation.Executable != permit.Executable || !reflect.DeepEqual(observation.Argv, permit.Argv) || observation.CWD != permit.CWD || !reflect.DeepEqual(observation.Environment, permit.Environment) || observation.ExitCode != 0 {
		return SourceAcquisitionReceipt{}, failure("closure_derivation_drift", "source acquisition observation differs from permit")
	}
	if permit.AssuranceMode == AssuranceVerified && (!reflect.DeepEqual(observation.Processes, permit.AllowedProcesses) || !reflect.DeepEqual(observation.Reads, permit.ReadRoots) || !reflect.DeepEqual(observation.Writes, permit.QuarantineWriteRoots) || observation.Network != "exact-origin-only") {
		return SourceAcquisitionReceipt{}, failure("closure_derivation_drift", "verified source acquisition boundary differs")
	}
	receipt := SourceAcquisitionReceipt{SchemaID: SchemaSourceAcquisitionReceipt, AssuranceMode: permit.AssuranceMode, PolicyID: permit.PolicyID, ExecutionPolicyID: permit.ExecutionPolicyID, ProviderContract: permit.ProviderContract, Provider: permit.Provider, CapabilityReceiptID: permit.CapabilityReceiptID, ActualCapabilities: append([]CapabilityEvidence(nil), permit.ActualCapabilities...), PermitID: permitID, BeforeFingerprint: before.Fingerprint, AfterFingerprint: after.Fingerprint, CanonicalOrigin: permit.CanonicalOrigin, RequestedRevision: permit.RequestedRevision, Observation: observation, ResourceLimits: permit.ResourceLimits, ResourceLimitID: permit.ResourceLimitID, ExpectedEvidence: append([]EvidenceRequirement(nil), permit.ExpectedEvidence...), EvidenceSchemaID: permit.EvidenceSchemaID, Diagnostics: []DerivationDiagnostic{}, Decision: "success"}
	headPayload := map[string]any{"permit_id": string(permitID), "observation": observation.canonicalValue()}
	receipt.NextCausalHead, err = closuregraph.DomainID("source-acquisition-causal-head-v1", headPayload)
	if err != nil {
		return SourceAcquisitionReceipt{}, err
	}
	if err = receipt.Validate(); err != nil {
		return SourceAcquisitionReceipt{}, err
	}
	receiptID, _ := receipt.ID()
	executor.receipts[receiptID] = receipt
	executor.head = string(receipt.NextCausalHead)
	return receipt, nil
}

// VerifyIssuedReceipt rejects caller-reconstructed acquisition evidence.
func (executor *SourceAcquisitionExecutor) VerifyIssuedReceipt(receipt SourceAcquisitionReceipt) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	id, err := receipt.ID()
	if err != nil {
		return err
	}
	executor.mu.Lock()
	defer executor.mu.Unlock()
	issued, ok := executor.receipts[id]
	if !ok || !reflect.DeepEqual(issued, receipt) {
		return failure("closure_derivation_unauthorized", "source acquisition receipt was not issued")
	}
	return nil
}

// ManagerSourceAcquisitionRunner is the portable manager-worker process seam.
// It makes no lossless network/process/read/write observation claim.
type ManagerSourceAcquisitionRunner struct {
	ExecutionRoot         string
	ExecutableRoot        string
	ProcessStartObserver  func(SourceAcquisitionPermit)
	ProcessLaunchObserver ProcessLaunchObserver
}

// NewManagerSourceAcquisitionRunner binds one symlink-aware execution root.
func NewManagerSourceAcquisitionRunner(executionRoot string) (*ManagerSourceAcquisitionRunner, error) {
	return NewManagerSourceAcquisitionRunnerWithExecutableRoot(executionRoot, executionRoot)
}

// NewManagerSourceAcquisitionRunnerWithExecutableRoot separates the trusted
// read-only tool root from task-private quarantine and evidence writes.
func NewManagerSourceAcquisitionRunnerWithExecutableRoot(executionRoot, executableRoot string) (*ManagerSourceAcquisitionRunner, error) {
	root, err := filepath.Abs(executionRoot)
	if err != nil || executionRoot == "" {
		return nil, fmt.Errorf("source acquisition execution root is invalid")
	}
	executable, err := filepath.Abs(executableRoot)
	if err != nil || executableRoot == "" {
		return nil, fmt.Errorf("source acquisition executable root is invalid")
	}
	return &ManagerSourceAcquisitionRunner{ExecutionRoot: root, ExecutableRoot: executable}, nil
}

// RunSourceAcquisition executes exactly one committed command and hashes its
// declared regular-file evidence from quarantine.
func (runner *ManagerSourceAcquisitionRunner) RunSourceAcquisition(ctx context.Context, permit SourceAcquisitionPermit) (AcquisitionRunResult, error) {
	if runner == nil {
		return AcquisitionRunResult{}, failure("portable_runner_missing", "source acquisition runner is absent")
	}
	executable, err := safeExecutionPath(runner.ExecutableRoot, permit.Executable)
	if err != nil {
		return AcquisitionRunResult{}, err
	}
	cwd, err := safeExecutionPath(runner.ExecutionRoot, permit.CWD)
	if err != nil {
		return AcquisitionRunResult{}, err
	}
	if err = verifyPortableExecutable(executable, permit.ExecutableSHA256); err != nil {
		return AcquisitionRunResult{}, err
	}
	for _, root := range permit.QuarantineWriteRoots {
		if _, err = safeExecutionPath(runner.ExecutionRoot, root); err != nil {
			return AcquisitionRunResult{}, err
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(permit.ResourceLimits.WallTimeMillis)*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(runCtx, executable, permit.Argv...) // #nosec G204 -- exact manager-issued acquisition permit.
	configurePortableProcess(command)
	command.Dir = cwd
	keys := make([]string, 0, len(permit.Environment))
	for key := range permit.Environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+permit.Environment[key])
	}
	budget := &portableOutputBudget{remaining: permit.ResourceLimits.OutputBytes, cancel: cancel}
	stdout, stderr := &portableBoundedBuffer{budget: budget}, &portableBoundedBuffer{budget: budget}
	command.Stdout, command.Stderr = stdout, stderr
	if runner.ProcessStartObserver != nil {
		runner.ProcessStartObserver(permit)
	}
	if runner.ProcessLaunchObserver != nil {
		runner.ProcessLaunchObserver(ProcessLaunch{Executable: command.Path, CWD: command.Dir, Argv: append([]string(nil), command.Args[1:]...), Environment: append([]string(nil), command.Env...)})
	}
	if err = command.Run(); err != nil {
		return AcquisitionRunResult{}, fmt.Errorf("source acquisition command failed: %w: %s", err, strings.TrimSpace(stderr.buffer.String()))
	}
	if budget.exceeded() {
		return AcquisitionRunResult{}, failure("portable_output_limit", "source acquisition output exceeded its byte bound")
	}
	if permit.StdoutEvidencePath != "" {
		stdoutPath, pathErr := safeExecutionPath(runner.ExecutionRoot, permit.StdoutEvidencePath)
		if pathErr != nil {
			return AcquisitionRunResult{}, pathErr
		}
		if pathErr = os.MkdirAll(filepath.Dir(stdoutPath), 0o700); pathErr != nil {
			return AcquisitionRunResult{}, pathErr
		}
		if pathErr = os.WriteFile(stdoutPath, stdout.buffer.Bytes(), 0o600); pathErr != nil {
			return AcquisitionRunResult{}, pathErr
		}
	}
	outputs := make([]DerivationOutput, len(permit.ExpectedEvidence))
	for index, expected := range permit.ExpectedEvidence {
		pathValue, pathErr := safeExecutionPath(runner.ExecutionRoot, expected.Path)
		if pathErr != nil {
			return AcquisitionRunResult{}, pathErr
		}
		payload, pathErr := os.ReadFile(pathValue) // #nosec G304 -- safeExecutionPath resolved the exact permit-declared evidence below the manager root.
		if pathErr != nil {
			return AcquisitionRunResult{}, pathErr
		}
		outputs[index] = DerivationOutput{Path: expected.Path, SchemaID: expected.SchemaID, ArtifactManifestID: expected.ArtifactManifestID, SHA256: digestBytes(payload), Size: int64(len(payload))}
	}
	objectIDs := []string{}
	if value := permit.Environment["CURATOR_ACQUISITION_OBJECT_IDS"]; value != "" {
		objectIDs = strings.Split(value, ",")
		sort.Strings(objectIDs)
	}
	resolvedRevision, gitTree := permit.Environment["CURATOR_ACQUISITION_REVISION"], permit.Environment["CURATOR_ACQUISITION_GIT_TREE"]
	switch permit.Environment["CURATOR_ACQUISITION_PARSE"] {
	case "git-tree":
		gitTree = strings.ToLower(strings.TrimSpace(stdout.buffer.String()))
	case "git-revision":
		resolvedRevision = strings.ToLower(strings.TrimSpace(stdout.buffer.String()))
	}
	return AcquisitionRunResult{ExitCode: 0, Evidence: outputs, ResolvedRevision: resolvedRevision, GitTree: gitTree, ObjectIDs: objectIDs}, nil
}
