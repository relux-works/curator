// Package closureexec owns immutable intake, authorized pre-C5 derivation,
// protected offline execution evidence, and multi-output publication. It is
// manager-neutral: ecosystem adapters supply declarations, never authority.
package closureexec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	// SchemaIntakeReceipt identifies immutable intake admission receipts.
	SchemaIntakeReceipt = "closure-intake-admission-receipt-v1"
	// SchemaDerivationPermit identifies committed pre-C5 execution permits.
	SchemaDerivationPermit = "closure-derivation-permit-v1"
	// SchemaDerivationReceipt identifies observed pre-C5 execution receipts.
	SchemaDerivationReceipt = "closure-derivation-receipt-v1"
)

// DerivationKind is the closed class of authorized pre-C5 derivations.
type DerivationKind string

const (
	// DerivationManifest authorizes deterministic manifest generation.
	DerivationManifest DerivationKind = "manifest"
	// DerivationVendor authorizes deterministic vendor-tree generation.
	DerivationVendor DerivationKind = "vendor"
	// DerivationMirror authorizes deterministic mirror generation.
	DerivationMirror DerivationKind = "mirror"
	// DerivationMetadata authorizes deterministic metadata generation.
	DerivationMetadata DerivationKind = "metadata"
)

// ResourceLimits is the complete portable resource contract enforced by a
// lossless provider. Zero never means unlimited.
type ResourceLimits struct {
	OutputBytes    int64 `json:"output_bytes"`
	ReadBytes      int64 `json:"read_bytes"`
	WriteBytes     int64 `json:"write_bytes"`
	WallTimeMillis int64 `json:"wall_time_millis"`
	ProcessCount   int64 `json:"process_count"`
}

// Validate checks the closed positive safe-integer resource contract.
func (limits ResourceLimits) Validate() error {
	for name, value := range map[string]int64{"output_bytes": limits.OutputBytes, "process_count": limits.ProcessCount, "read_bytes": limits.ReadBytes, "wall_time_millis": limits.WallTimeMillis, "write_bytes": limits.WriteBytes} {
		if value <= 0 || value > protocoljson.MaxSafeInteger {
			return failure("closure_derivation_unauthorized", "resource limit %s is invalid", name)
		}
	}
	return nil
}

func (limits ResourceLimits) canonicalValue() map[string]any {
	return map[string]any{"output_bytes": limits.OutputBytes, "process_count": limits.ProcessCount, "read_bytes": limits.ReadBytes, "wall_time_millis": limits.WallTimeMillis, "write_bytes": limits.WriteBytes}
}

// ID derives the exact typed resource-limit identity.
func (limits ResourceLimits) ID() (closuregraph.ID, error) {
	if err := limits.Validate(); err != nil {
		return "", err
	}
	return closuregraph.DomainID("curator-derivation-resource-limits-v1", limits.canonicalValue())
}

// EvidenceRequirement binds an output path to its schema and immutable
// manifest declaration before execution.
type EvidenceRequirement struct {
	Path               string          `json:"path"`
	SchemaID           string          `json:"schema_id"`
	ArtifactManifestID closuregraph.ID `json:"artifact_manifest_id"`
}

func (requirement EvidenceRequirement) validate() error {
	if err := portablePath(requirement.Path); err != nil {
		return err
	}
	if requirement.SchemaID == "" || strings.ContainsAny(requirement.SchemaID, "\x00\r\n") || !requirement.ArtifactManifestID.Valid() {
		return fmt.Errorf("invalid evidence schema or manifest identity")
	}
	return nil
}

func (requirement EvidenceRequirement) canonicalValue() map[string]any {
	return map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}
}

// LocalOutputDeclaration binds a retained derivation tree that is not itself
// portable evidence. Its bytes require a policy-specific authorization before
// they can become a new admitted input.
type LocalOutputDeclaration struct {
	Path     string `json:"path"`
	SchemaID string `json:"schema_id"`
}

func (output LocalOutputDeclaration) validate() error {
	if err := portablePath(output.Path); err != nil {
		return err
	}
	if output.SchemaID == "" || strings.ContainsAny(output.SchemaID, "\x00\r\n") {
		return fmt.Errorf("invalid local-output schema")
	}
	return nil
}

func (output LocalOutputDeclaration) canonicalValue() map[string]any {
	return map[string]any{"path": output.Path, "schema_id": output.SchemaID}
}

// DerivationOutput is authoritative observed evidence for one derived output.
type DerivationOutput struct {
	Path               string          `json:"path"`
	SchemaID           string          `json:"schema_id"`
	ArtifactManifestID closuregraph.ID `json:"artifact_manifest_id"`
	SHA256             closuregraph.ID `json:"sha256"`
	Size               int64           `json:"size"`
}

func (output DerivationOutput) validate() error {
	if err := (EvidenceRequirement{Path: output.Path, SchemaID: output.SchemaID, ArtifactManifestID: output.ArtifactManifestID}).validate(); err != nil {
		return err
	}
	if !output.SHA256.Valid() || output.Size < 0 || output.Size > protocoljson.MaxSafeInteger {
		return fmt.Errorf("invalid evidence digest or size")
	}
	return nil
}

func (output DerivationOutput) canonicalValue() map[string]any {
	return map[string]any{"artifact_manifest_id": string(output.ArtifactManifestID), "path": output.Path, "schema_id": output.SchemaID, "sha256": string(output.SHA256), "size": output.Size}
}

// DerivationDiagnostic is one deterministic provider or reconciliation
// diagnostic retained in the canonical derivation journal.
type DerivationDiagnostic struct {
	Code    string            `json:"code"`
	Subject string            `json:"subject"`
	Fields  map[string]string `json:"fields"`
}

func (diagnostic DerivationDiagnostic) validate() error {
	if diagnostic.Code == "" || diagnostic.Subject == "" || strings.ContainsAny(diagnostic.Code+diagnostic.Subject, "\x00\r\n") || diagnostic.Fields == nil {
		return fmt.Errorf("invalid derivation diagnostic")
	}
	for key, value := range diagnostic.Fields {
		if key == "" || strings.ContainsAny(key+value, "\x00\r\n") {
			return fmt.Errorf("invalid derivation diagnostic field")
		}
	}
	return nil
}

func (diagnostic DerivationDiagnostic) canonicalValue() map[string]any {
	return map[string]any{"code": diagnostic.Code, "fields": stringMap(diagnostic.Fields), "subject": diagnostic.Subject}
}

// DiagnosticError carries the stable fail-closed code for an execution gate.
type DiagnosticError struct{ Code, Detail string }

func (e *DiagnosticError) Error() string { return e.Code + ": " + e.Detail }

func failure(code, format string, args ...any) error {
	return &DiagnosticError{Code: code, Detail: fmt.Sprintf(format, args...)}
}

// IntakeAdmissionReceipt binds admitted bytes to an immutable protected handle.
type IntakeAdmissionReceipt struct {
	SchemaID, PreviousCausalHead, OriginID, ProtectedHandleID string
	ContentSHA256                                             closuregraph.ID
	Size                                                      int64
	ArtifactPolicyID, SourceProfileID, DetectorRegistryID     string
	LimitVectorID                                             string
	ArtifactManifestID                                        closuregraph.ID
	Decision                                                  string
}

func (r IntakeAdmissionReceipt) canonicalValue() map[string]any {
	return map[string]any{"artifact_manifest_id": string(r.ArtifactManifestID), "artifact_policy_id": r.ArtifactPolicyID, "content_sha256": string(r.ContentSHA256), "decision": r.Decision, "detector_registry_id": r.DetectorRegistryID, "limit_vector_id": r.LimitVectorID, "origin_id": r.OriginID, "previous_causal_head": r.PreviousCausalHead, "protected_handle_id": r.ProtectedHandleID, "schema_id": r.SchemaID, "size": r.Size, "source_profile_id": r.SourceProfileID}
}

// Validate checks the closed intake receipt contract.
func (r IntakeAdmissionReceipt) Validate() error {
	if r.SchemaID != SchemaIntakeReceipt || r.Decision != "ADMIT_INPUT" || !r.ContentSHA256.Valid() || !r.ArtifactManifestID.Valid() || r.Size < 0 {
		return failure("closure_derivation_unauthorized", "invalid intake admission receipt")
	}
	for name, value := range map[string]string{"previous causal head": r.PreviousCausalHead, "origin": r.OriginID, "protected handle": r.ProtectedHandleID, "artifact policy": r.ArtifactPolicyID, "source profile": r.SourceProfileID, "detector registry": r.DetectorRegistryID, "limit vector": r.LimitVectorID} {
		if value == "" || strings.ContainsAny(value, "\x00\r\n") {
			return failure("closure_derivation_unauthorized", "%s is invalid", name)
		}
	}
	return nil
}

// CanonicalBytes returns the exact CCJ receipt bytes.
func (r IntakeAdmissionReceipt) CanonicalBytes() ([]byte, error) {
	return canonical(r.Validate, r.canonicalValue())
}

// ID derives the domain-separated receipt identity.
func (r IntakeAdmissionReceipt) ID() (closuregraph.ID, error) {
	b, err := r.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return closuregraph.IDFromCanonical(closuregraph.LabelIntakeAdmissionReceipt, b)
}

// DerivationPermit is committed before any pre-C5 process may start.
type DerivationPermit struct {
	SchemaID                string                   `json:"schema_id"`
	AssuranceMode           AssuranceMode            `json:"assurance_mode"`
	PolicyID                string                   `json:"policy_id"`
	ExecutionPolicyID       string                   `json:"execution_policy"`
	ProviderContract        *string                  `json:"provider_contract"`
	Provider                *ProviderIdentity        `json:"provider"`
	CapabilityReceiptID     *closuregraph.ID         `json:"capability_receipt_sha256"`
	ActualCapabilities      []CapabilityEvidence     `json:"actual_capabilities"`
	PreviousCausalHead      string                   `json:"previous_causal_head"`
	InvocationKey           string                   `json:"invocation_key"`
	InvocationSubtype       DerivationKind           `json:"invocation_subtype"`
	AdmittedInputReceiptIDs []closuregraph.ID        `json:"admitted_input_receipt_ids"`
	InputMounts             []InputMount             `json:"input_mounts"`
	WorkCopies              []WorkCopy               `json:"work_copies"`
	C0CheckpointID          closuregraph.ID          `json:"c0_checkpoint_id"`
	ToolchainNodeID         closuregraph.ID          `json:"toolchain_node_id"`
	ToolchainFingerprint    closuregraph.ID          `json:"toolchain_fingerprint"`
	ExecutableSHA256        closuregraph.ID          `json:"executable_sha256"`
	Executable              string                   `json:"executable"`
	CWD                     string                   `json:"cwd"`
	Argv                    []string                 `json:"argv"`
	Environment             map[string]string        `json:"environment"`
	HostID                  closuregraph.ID          `json:"host_id"`
	TargetID                closuregraph.ID          `json:"target_id"`
	AllowedProcesses        []string                 `json:"allowed_processes"`
	ReadRoots               []string                 `json:"read_roots"`
	WriteRoots              []string                 `json:"write_roots"`
	StdoutEvidencePath      string                   `json:"stdout_evidence_path"`
	ExpectedEvidence        []EvidenceRequirement    `json:"expected_evidence"`
	LocalOutputs            []LocalOutputDeclaration `json:"local_outputs,omitempty"`
	Network                 string                   `json:"network"`
	RecheckRule             string                   `json:"recheck_rule"`
	ResourceLimits          ResourceLimits           `json:"resource_limits"`
	ResourceLimitID         closuregraph.ID          `json:"resource_limit_id"`
	EvidenceSchemaID        closuregraph.ID          `json:"evidence_schema_id"`
}

// Validate checks the closed, offline permit boundary.
func (p DerivationPermit) Validate() error {
	if p.SchemaID != SchemaDerivationPermit || p.PreviousCausalHead == "" || p.InvocationKey == "" || !validDerivationKind(p.InvocationSubtype) || !p.C0CheckpointID.Valid() || !p.ToolchainNodeID.Valid() || !p.ToolchainFingerprint.Valid() || !p.ExecutableSHA256.Valid() {
		return failure("closure_derivation_unauthorized", "permit authority is incomplete")
	}
	if err := validateAssuranceBinding(p.AssuranceMode, p.PolicyID, p.ExecutionPolicyID, p.ProviderContract, p.Provider, p.CapabilityReceiptID, p.ActualCapabilities); err != nil {
		return err
	}
	if len(p.AdmittedInputReceiptIDs) == 0 || !sortedUniqueIDs(p.AdmittedInputReceiptIDs) {
		return failure("closure_derivation_unauthorized", "admitted input receipts must be nonempty, sorted, and unique")
	}
	if p.Network != "none" || p.RecheckRule != "immediate-exact-v1" {
		return failure("closure_derivation_unauthorized", "permit boundary is not fail-closed")
	}
	limitID, err := p.ResourceLimits.ID()
	if err != nil || limitID != p.ResourceLimitID {
		return failure("closure_derivation_unauthorized", "resource-limit identity differs")
	}
	evidenceID, err := evidenceSchemaID(p.ExpectedEvidence)
	if err != nil || evidenceID != p.EvidenceSchemaID {
		return failure("closure_derivation_unauthorized", "evidence-schema identity differs")
	}
	if err := portablePath(p.Executable); err != nil {
		return failure("closure_derivation_unauthorized", "executable: %v", err)
	}
	if err := portablePath(p.CWD); err != nil {
		return failure("closure_derivation_unauthorized", "cwd: %v", err)
	}
	for _, values := range [][]string{p.AllowedProcesses, p.ReadRoots, p.WriteRoots} {
		if !sort.StringsAreSorted(values) || hasDuplicates(values) {
			return failure("closure_derivation_unauthorized", "permit sets must be sorted and unique")
		}
	}
	if len(p.InputMounts) != len(p.AdmittedInputReceiptIDs) {
		return failure("closure_derivation_unauthorized", "every admitted input must have one replay mount")
	}
	mountPaths := make([]string, len(p.InputMounts))
	for i, mount := range p.InputMounts {
		if err := mount.validate(); err != nil || mount.ReceiptID != p.AdmittedInputReceiptIDs[i] || (i > 0 && p.InputMounts[i-1].ReceiptID >= mount.ReceiptID) || !containsString(p.ReadRoots, mount.Path) {
			return failure("closure_derivation_unauthorized", "input mounts must exactly match admitted receipts")
		}
		mountPaths[i] = mount.Path
		for _, writable := range p.WriteRoots {
			if logicalPathsOverlap(mount.Path, writable) {
				return failure("closure_derivation_unauthorized", "immutable replay mount overlaps a writable root")
			}
		}
	}
	if sortedMounts := sortedCopy(mountPaths); hasDuplicates(sortedMounts) {
		return failure("closure_derivation_unauthorized", "input mount paths must be unique")
	}
	// ReadRoots also names manager-selected external-toolchain roots. Those
	// bytes are bound by ToolchainNodeID/Fingerprint rather than intake
	// receipts, so every intake mount must be present but need not exhaust the
	// observed read policy.
	for _, root := range p.ReadRoots {
		if err := portablePath(root); err != nil {
			return failure("closure_derivation_unauthorized", "read root is invalid")
		}
		for _, writable := range p.WriteRoots {
			if logicalPathsOverlap(root, writable) {
				return failure("closure_derivation_unauthorized", "read root overlaps a writable root")
			}
		}
	}
	expectedPaths := make([]string, len(p.ExpectedEvidence))
	for i, requirement := range p.ExpectedEvidence {
		expectedPaths[i] = requirement.Path
	}
	previousLocalOutput := ""
	for _, output := range p.LocalOutputs {
		if err := output.validate(); err != nil || (previousLocalOutput != "" && previousLocalOutput >= output.Path) {
			return failure("closure_derivation_unauthorized", "local outputs must be valid, sorted, and unique")
		}
		for _, mount := range p.InputMounts {
			if logicalPathsOverlap(mount.Path, output.Path) {
				return failure("closure_derivation_unauthorized", "local output overlaps an immutable replay mount")
			}
		}
		expectedPaths = append(expectedPaths, output.Path)
		previousLocalOutput = output.Path
	}
	for index, work := range p.WorkCopies {
		if err := work.validate(); err != nil || !containsID(p.AdmittedInputReceiptIDs, work.ReceiptID) || (index > 0 && p.WorkCopies[index-1].ReceiptID >= work.ReceiptID) {
			return failure("closure_derivation_unauthorized", "work copies must name unique admitted inputs in receipt order")
		}
		for _, mount := range p.InputMounts {
			if logicalPathsOverlap(mount.Path, work.Path) {
				return failure("closure_derivation_unauthorized", "work copy overlaps an immutable replay mount")
			}
		}
		expectedPaths = append(expectedPaths, work.Path)
	}
	if p.StdoutEvidencePath != "" {
		if err := portablePath(p.StdoutEvidencePath); err != nil || !containsString(expectedPaths, p.StdoutEvidencePath) {
			return failure("closure_derivation_unauthorized", "stdout evidence path is not a typed output")
		}
	}
	sort.Strings(expectedPaths)
	if !equalStrings(expectedPaths, p.WriteRoots) {
		return failure("closure_derivation_unauthorized", "write roots must be exactly the typed evidence outputs")
	}
	if !p.HostID.Valid() || !p.TargetID.Valid() {
		return failure("closure_derivation_unauthorized", "host or target identity is invalid")
	}
	return validateEnvironment(p.Environment)
}

func (p DerivationPermit) canonicalValue() map[string]any {
	evidence := make([]any, len(p.ExpectedEvidence))
	for i, requirement := range p.ExpectedEvidence {
		evidence[i] = requirement.canonicalValue()
	}
	mounts := make([]any, len(p.InputMounts))
	for i, mount := range p.InputMounts {
		mounts[i] = mount.canonicalValue()
	}
	works := make([]any, len(p.WorkCopies))
	for i, work := range p.WorkCopies {
		works[i] = work.canonicalValue()
	}
	value := map[string]any{"actual_capabilities": capabilitiesValue(p.ActualCapabilities), "admitted_input_receipt_ids": ids(p.AdmittedInputReceiptIDs), "allowed_processes": stringsAny(p.AllowedProcesses), "argv": stringsAny(p.Argv), "assurance_mode": string(p.AssuranceMode), "c0_checkpoint_id": string(p.C0CheckpointID), "capability_receipt_sha256": optionalID(p.CapabilityReceiptID), "cwd": p.CWD, "environment": stringMap(p.Environment), "evidence_schema_id": string(p.EvidenceSchemaID), "executable": p.Executable, "executable_sha256": string(p.ExecutableSHA256), "execution_policy": p.ExecutionPolicyID, "expected_evidence": evidence, "host_id": string(p.HostID), "input_mounts": mounts, "invocation_key": p.InvocationKey, "invocation_subtype": string(p.InvocationSubtype), "network": p.Network, "policy_id": p.PolicyID, "previous_causal_head": p.PreviousCausalHead, "provider": optionalProvider(p.Provider), "provider_contract": optionalString(p.ProviderContract), "read_roots": stringsAny(p.ReadRoots), "recheck_rule": p.RecheckRule, "resource_limit_id": string(p.ResourceLimitID), "resource_limits": p.ResourceLimits.canonicalValue(), "schema_id": p.SchemaID, "stdout_evidence_path": p.StdoutEvidencePath, "target_id": string(p.TargetID), "toolchain_fingerprint": string(p.ToolchainFingerprint), "toolchain_node_id": string(p.ToolchainNodeID), "work_copies": works, "write_roots": stringsAny(p.WriteRoots)}
	if len(p.LocalOutputs) > 0 {
		outputs := make([]any, len(p.LocalOutputs))
		for index := range p.LocalOutputs {
			outputs[index] = p.LocalOutputs[index].canonicalValue()
		}
		value["local_outputs"] = outputs
	}
	return value
}

// ToolchainIdentity is the immediate tree and executable-byte recheck result.
type ToolchainIdentity struct {
	Fingerprint, ExecutableSHA256 closuregraph.ID
}

// CanonicalBytes returns the exact CCJ permit bytes.
func (p DerivationPermit) CanonicalBytes() ([]byte, error) {
	return canonical(p.Validate, p.canonicalValue())
}

// DecodeDerivationPermit accepts only exact canonical permit bytes and rejects
// unknown fields, identity drift, and noncanonical ordering.
func DecodeDerivationPermit(payload []byte) (DerivationPermit, error) {
	var permit DerivationPermit
	if err := protocoljson.UnmarshalCanonical(payload, &permit); err != nil {
		return DerivationPermit{}, err
	}
	if err := permit.Validate(); err != nil {
		return DerivationPermit{}, err
	}
	roundTrip, err := permit.CanonicalBytes()
	if err != nil || string(roundTrip) != string(payload) {
		return DerivationPermit{}, failure("closure_derivation_unauthorized", "permit canonical round-trip differs")
	}
	return permit, nil
}

// ID derives the domain-separated permit identity.
func (p DerivationPermit) ID() (closuregraph.ID, error) {
	b, e := p.CanonicalBytes()
	if e != nil {
		return "", e
	}
	return closuregraph.IDFromCanonical(closuregraph.LabelDerivationPermit, b)
}

// Audit is produced by the OS-enforced boundary, not by an adapter.
type Audit struct {
	Executable  string             `json:"executable"`
	CWD         string             `json:"cwd"`
	Argv        []string           `json:"argv"`
	Environment map[string]string  `json:"environment"`
	Processes   []string           `json:"processes"`
	Reads       []string           `json:"reads"`
	Writes      []string           `json:"writes"`
	Evidence    []string           `json:"evidence"`
	Network     string             `json:"network"`
	ExitCode    int64              `json:"exit_code"`
	Outputs     []DerivationOutput `json:"outputs"`
}

// DerivationReceipt records the exact permitted and observed invocation.
type DerivationReceipt struct {
	SchemaID            string                 `json:"schema_id"`
	AssuranceMode       AssuranceMode          `json:"assurance_mode"`
	PolicyID            string                 `json:"policy_id"`
	ExecutionPolicyID   string                 `json:"execution_policy"`
	ProviderContract    *string                `json:"provider_contract"`
	Provider            *ProviderIdentity      `json:"provider"`
	CapabilityReceiptID *closuregraph.ID       `json:"capability_receipt_sha256"`
	ActualCapabilities  []CapabilityEvidence   `json:"actual_capabilities"`
	PermitID            closuregraph.ID        `json:"permit_id"`
	BeforeFingerprint   closuregraph.ID        `json:"before_toolchain_fingerprint"`
	AfterFingerprint    closuregraph.ID        `json:"after_toolchain_fingerprint"`
	Audit               Audit                  `json:"audit"`
	InvocationSubtype   DerivationKind         `json:"invocation_subtype"`
	ResourceLimits      ResourceLimits         `json:"resource_limits"`
	ResourceLimitID     closuregraph.ID        `json:"resource_limit_id"`
	ExpectedEvidence    []EvidenceRequirement  `json:"expected_evidence"`
	EvidenceSchemaID    closuregraph.ID        `json:"evidence_schema_id"`
	Outputs             []DerivationOutput     `json:"outputs"`
	Diagnostics         []DerivationDiagnostic `json:"diagnostics"`
	NextCausalHead      closuregraph.ID        `json:"next_causal_head"`
	Decision            string                 `json:"decision"`
}

// Validate checks a successful observed derivation receipt.
func (r DerivationReceipt) Validate() error {
	if r.SchemaID != SchemaDerivationReceipt || !r.PermitID.Valid() || !r.BeforeFingerprint.Valid() || !r.AfterFingerprint.Valid() || !validDerivationKind(r.InvocationSubtype) || !r.ResourceLimitID.Valid() || !r.EvidenceSchemaID.Valid() || r.Decision != "success" || r.Audit.ExitCode != 0 {
		return failure("closure_derivation_drift", "invalid successful derivation receipt")
	}
	if err := validateAssuranceBinding(r.AssuranceMode, r.PolicyID, r.ExecutionPolicyID, r.ProviderContract, r.Provider, r.CapabilityReceiptID, r.ActualCapabilities); err != nil {
		return err
	}
	if (r.AssuranceMode == AssuranceVerified && r.Audit.Network != "none") || (r.AssuranceMode == AssurancePortable && r.Audit.Network != "not-observed") {
		return failure("assurance_claim_inflation", "network evidence exceeds selected assurance mode")
	}
	if !equalOutputs(r.Outputs, r.Audit.Outputs) {
		return failure("closure_derivation_drift", "receipt outputs differ from authoritative audit")
	}
	limitID, err := r.ResourceLimits.ID()
	if err != nil || limitID != r.ResourceLimitID {
		return failure("closure_derivation_drift", "receipt resource-limit identity differs")
	}
	evidenceID, err := evidenceSchemaID(r.ExpectedEvidence)
	if err != nil || evidenceID != r.EvidenceSchemaID || len(r.ExpectedEvidence) != len(r.Outputs) {
		return failure("closure_derivation_drift", "receipt evidence-schema identity differs")
	}
	for i, output := range r.Outputs {
		expected := r.ExpectedEvidence[i]
		if err := output.validate(); err != nil || (i > 0 && r.Outputs[i-1].Path >= output.Path) || output.Path != expected.Path || output.SchemaID != expected.SchemaID || output.ArtifactManifestID != expected.ArtifactManifestID {
			return failure("closure_derivation_drift", "receipt outputs are invalid or noncanonical")
		}
	}
	if r.Diagnostics == nil {
		return failure("closure_derivation_drift", "receipt diagnostics must be explicit")
	}
	previousDiagnostic := ""
	for _, diagnostic := range r.Diagnostics {
		if err := diagnostic.validate(); err != nil {
			return failure("closure_derivation_drift", "receipt diagnostic is invalid")
		}
		value, err := protocoljson.MarshalCanonical(diagnostic.canonicalValue())
		if err != nil || (previousDiagnostic != "" && previousDiagnostic >= string(value)) {
			return failure("closure_derivation_drift", "receipt diagnostics are not sorted and unique")
		}
		previousDiagnostic = string(value)
	}
	next, err := r.deriveNextCausalHead()
	if err != nil || next != r.NextCausalHead {
		return failure("closure_derivation_drift", "next causal head differs")
	}
	return nil
}
func (r DerivationReceipt) canonicalValue() map[string]any {
	outputs := make([]any, len(r.Outputs))
	for i, output := range r.Outputs {
		outputs[i] = output.canonicalValue()
	}
	expected := make([]any, len(r.ExpectedEvidence))
	for i, requirement := range r.ExpectedEvidence {
		expected[i] = requirement.canonicalValue()
	}
	diagnostics := make([]any, len(r.Diagnostics))
	for i, diagnostic := range r.Diagnostics {
		diagnostics[i] = diagnostic.canonicalValue()
	}
	return map[string]any{"actual_capabilities": capabilitiesValue(r.ActualCapabilities), "after_toolchain_fingerprint": string(r.AfterFingerprint), "assurance_mode": string(r.AssuranceMode), "audit": auditValue(r.Audit), "before_toolchain_fingerprint": string(r.BeforeFingerprint), "capability_receipt_sha256": optionalID(r.CapabilityReceiptID), "decision": r.Decision, "diagnostics": diagnostics, "evidence_schema_id": string(r.EvidenceSchemaID), "execution_policy": r.ExecutionPolicyID, "expected_evidence": expected, "invocation_subtype": string(r.InvocationSubtype), "next_causal_head": string(r.NextCausalHead), "outputs": outputs, "permit_id": string(r.PermitID), "policy_id": r.PolicyID, "provider": optionalProvider(r.Provider), "provider_contract": optionalString(r.ProviderContract), "resource_limit_id": string(r.ResourceLimitID), "resource_limits": r.ResourceLimits.canonicalValue(), "schema_id": r.SchemaID}
}

// CanonicalBytes returns the exact CCJ derivation receipt bytes.
func (r DerivationReceipt) CanonicalBytes() ([]byte, error) {
	return canonical(r.Validate, r.canonicalValue())
}

// DecodeDerivationReceipt accepts only exact canonical receipt bytes and
// re-derives its output and next-causal-head identities.
func DecodeDerivationReceipt(payload []byte) (DerivationReceipt, error) {
	var receipt DerivationReceipt
	if err := protocoljson.UnmarshalCanonical(payload, &receipt); err != nil {
		return DerivationReceipt{}, err
	}
	if err := receipt.Validate(); err != nil {
		return DerivationReceipt{}, err
	}
	roundTrip, err := receipt.CanonicalBytes()
	if err != nil || string(roundTrip) != string(payload) {
		return DerivationReceipt{}, failure("closure_derivation_drift", "receipt canonical round-trip differs")
	}
	return receipt, nil
}

// ID derives the domain-separated derivation receipt identity.
func (r DerivationReceipt) ID() (closuregraph.ID, error) {
	b, e := r.CanonicalBytes()
	if e != nil {
		return "", e
	}
	return closuregraph.IDFromCanonical(closuregraph.LabelDerivationReceipt, b)
}

func auditValue(a Audit) map[string]any {
	outputs := make([]any, len(a.Outputs))
	for i, output := range a.Outputs {
		outputs[i] = output.canonicalValue()
	}
	return map[string]any{"argv": stringsAny(a.Argv), "cwd": a.CWD, "environment": stringMap(a.Environment), "evidence": stringsAny(a.Evidence), "executable": a.Executable, "exit_code": a.ExitCode, "network": a.Network, "outputs": outputs, "processes": stringsAny(a.Processes), "reads": stringsAny(a.Reads), "writes": stringsAny(a.Writes)}
}

func validDerivationKind(kind DerivationKind) bool {
	return kind == DerivationManifest || kind == DerivationVendor || kind == DerivationMirror || kind == DerivationMetadata
}
func evidenceSchemaID(requirements []EvidenceRequirement) (closuregraph.ID, error) {
	values := make([]any, len(requirements))
	previousPath := ""
	for i, requirement := range requirements {
		if err := requirement.validate(); err != nil {
			return "", fmt.Errorf("evidence requirements must be valid, sorted, and unique")
		}
		if i > 0 && previousPath >= requirement.Path {
			return "", fmt.Errorf("evidence requirements must be valid, sorted, and unique")
		}
		values[i] = requirement.canonicalValue()
		previousPath = requirement.Path
	}
	if len(values) == 0 {
		return "", fmt.Errorf("evidence requirements must not be empty")
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": values})
}
func equalOutputs(a, b []DerivationOutput) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (r DerivationReceipt) deriveNextCausalHead() (closuregraph.ID, error) {
	outputs := make([]any, len(r.Outputs))
	for i, output := range r.Outputs {
		outputs[i] = output.canonicalValue()
	}
	expected := make([]any, len(r.ExpectedEvidence))
	for i, requirement := range r.ExpectedEvidence {
		expected[i] = requirement.canonicalValue()
	}
	diagnostics := make([]any, len(r.Diagnostics))
	for i, diagnostic := range r.Diagnostics {
		diagnostics[i] = diagnostic.canonicalValue()
	}
	return closuregraph.DomainID("curator-derivation-next-causal-head-v1", map[string]any{"actual_capabilities": capabilitiesValue(r.ActualCapabilities), "after_toolchain_fingerprint": string(r.AfterFingerprint), "assurance_mode": string(r.AssuranceMode), "audit": auditValue(r.Audit), "before_toolchain_fingerprint": string(r.BeforeFingerprint), "capability_receipt_sha256": optionalID(r.CapabilityReceiptID), "decision": r.Decision, "diagnostics": diagnostics, "evidence_schema_id": string(r.EvidenceSchemaID), "execution_policy": r.ExecutionPolicyID, "expected_evidence": expected, "invocation_subtype": string(r.InvocationSubtype), "outputs": outputs, "permit_id": string(r.PermitID), "policy_id": r.PolicyID, "provider": optionalProvider(r.Provider), "provider_contract": optionalString(r.ProviderContract), "resource_limit_id": string(r.ResourceLimitID), "resource_limits": r.ResourceLimits.canonicalValue(), "schema_id": r.SchemaID})
}

func validateAssuranceBinding(mode AssuranceMode, policyID, executionPolicyID string, providerContract *string, provider *ProviderIdentity, capabilityReceiptID *closuregraph.ID, capabilities []CapabilityEvidence) error {
	switch mode {
	case AssurancePortable:
		if policyID != PortablePolicyID || executionPolicyID != PortableExecutionPolicyID || providerContract != nil || provider != nil || capabilityReceiptID != nil || !reflect.DeepEqual(capabilities, portableCapabilities) {
			return failure("assurance_evidence_mismatch", "portable assurance binding differs")
		}
	case AssuranceVerified:
		if policyID != VerifiedPolicyID || executionPolicyID != VerifiedExecutionPolicyID || providerContract == nil || *providerContract != VerifiedProviderContractID || provider == nil || capabilityReceiptID == nil || !capabilityReceiptID.Valid() || len(capabilities) != len(verifiedCapabilities) {
			return failure("assurance_evidence_mismatch", "verified assurance binding differs")
		}
		for index, capability := range capabilities {
			if capability.CapabilityID != verifiedCapabilities[index] || capability.Status != "established" {
				return failure("assurance_claim_inflation", "verified capability record differs")
			}
		}
	default:
		return failure("execution_mode_unknown", "unknown assurance mode")
	}
	return nil
}

func optionalID(value *closuregraph.ID) any {
	if value == nil {
		return nil
	}
	return string(*value)
}
func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
func optionalProvider(value *ProviderIdentity) any {
	if value == nil {
		return nil
	}
	return providerValue(*value)
}
func canonical(validate func() error, value map[string]any) ([]byte, error) {
	if err := validate(); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(value)
}
func digestBytes(b []byte) closuregraph.ID {
	sum := sha256.Sum256(b)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}
func portablePath(v string) error {
	if v == "" || path.IsAbs(v) || path.Clean(v) != v || v == "." || strings.HasPrefix(v, "../") || strings.ContainsAny(v, "\\\x00\r\n") {
		return fmt.Errorf("not a canonical relative path")
	}
	return nil
}
func sortedUniqueIDs(v []closuregraph.ID) bool {
	for i, x := range v {
		if !x.Valid() || (i > 0 && v[i-1] >= x) {
			return false
		}
	}
	return true
}
func hasDuplicates(v []string) bool {
	for i := 1; i < len(v); i++ {
		if v[i-1] == v[i] {
			return true
		}
	}
	return false
}
func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func containsID(values []closuregraph.ID, wanted closuregraph.ID) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func logicalPathsOverlap(left, right string) bool {
	return left == right || strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}
func validateEnvironment(v map[string]string) error {
	if v == nil {
		return failure("closure_derivation_unauthorized", "environment must be explicit")
	}
	for k, x := range v {
		if k == "" || strings.ContainsAny(k, "=\x00\r\n") || strings.ContainsAny(x, "\x00\r\n") {
			return failure("closure_derivation_unauthorized", "environment is invalid")
		}
	}
	return nil
}
func ids(v []closuregraph.ID) []any {
	r := make([]any, len(v))
	for i, x := range v {
		r[i] = string(x)
	}
	return r
}
func stringsAny(v []string) []any {
	r := make([]any, len(v))
	for i, x := range v {
		r[i] = x
	}
	return r
}
func stringMap(v map[string]string) map[string]any {
	r := make(map[string]any, len(v))
	for k, x := range v {
		r[k] = x
	}
	return r
}
func sortedCopy(v []string) []string { r := append([]string(nil), v...); sort.Strings(r); return r }
