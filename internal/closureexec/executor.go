package closureexec

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
)

// EnforceObserveProvider is the pluggable authoritative process boundary.
// Implementations must enforce the permit and return the complete attempted
// and successful process/read/write/network/output event set. An enforcement-
// only sandbox, sampled tracer, or child-supplied manifest is not lossless.
type EnforceObserveProvider interface {
	EnforceAndObserve(context.Context, ExecutionRequest) (Audit, error)
	LosslessObservation() bool
}

// ReplayInput exposes only permit-named protected inputs to a provider.
type ReplayInput struct {
	ReceiptID closuregraph.ID
	MountPath string
	input     AdmittedInput
}

// VerifyAtUse rechecks the admitted immutable identity immediately before the
// provider maps this input read-only.
func (input ReplayInput) VerifyAtUse() error { return input.input.recheck(input.ReceiptID) }

// ProtectedPath returns the protected source after a fresh identity recheck.
// Providers must map it read-only at MountPath and must not copy it into a
// writable ambient workspace.
func (input ReplayInput) ProtectedPath() (string, error) {
	if err := input.VerifyAtUse(); err != nil {
		return "", err
	}
	if input.input.Tree != nil {
		return input.input.Tree.path, nil
	}
	return input.input.Handle.path, nil
}

// IsTree reports whether the replay input is a source snapshot tree.
func (input ReplayInput) IsTree() bool { return input.input.Tree != nil }

// ExecutionRequest is the process-start request passed to a lossless provider.
type ExecutionRequest struct {
	Permit DerivationPermit
	Inputs []ReplayInput
}

// Executor commits permits and serializes pre-C5 derivations by causal head.
type Executor struct {
	mu                 sync.Mutex
	boundary           EnforceObserveProvider
	runner             PortableProcessRunner
	provider           VerifiedProvider
	assurance          AssuranceConfig
	committed          map[closuregraph.ID]DerivationPermit
	issuedPermits      map[closuregraph.ID]DerivationPermit
	capabilityReceipts map[closuregraph.ID]ProviderCapabilityReceipt
	head               string
	receipts           map[closuregraph.ID]DerivationReceipt
}

// AssuredOperation is a preflighted assurance scope. In verified mode its
// nonce-bound capability receipt must dominate both cache lookup and process
// dispatch; callers cannot manufacture this value.
type AssuredOperation struct {
	executor *Executor
	binding  AssuranceBinding
	receipt  *ProviderCapabilityReceipt
}

// CurrentCausalHead returns the predecessor for the next manager-issued
// derivation permit.
func (e *Executor) CurrentCausalHead() string {
	if e == nil {
		return ""
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.head
}

// NewExecutor creates an explicitly verified executor. It is retained as the
// provider-facing compatibility entry point; new manager code should call
// NewAssuredExecutor so the portable default is explicit.
func NewExecutor(boundary EnforceObserveProvider, initialCausalHead string) (*Executor, error) {
	provider, ok := boundary.(VerifiedProvider)
	if !ok || provider == nil {
		return nil, failure("verified_provider_unavailable", "provider does not implement the verified contract")
	}
	identity := provider.Identity()
	return NewAssuredExecutor(AssuranceConfig{Mode: AssuranceVerified, ProviderID: identity.ProviderID, ProviderVersion: identity.Version, ProviderBinarySHA256: identity.BinarySHA256, ProviderTrustEvidence: identity.TrustEvidence}, nil, provider, initialCausalHead)
}

// NewAssuredExecutor selects exactly one non-aliasing assurance strategy.
func NewAssuredExecutor(config AssuranceConfig, runner PortableProcessRunner, provider VerifiedProvider, initialCausalHead string) (*Executor, error) {
	config = config.normalized()
	if err := config.validate(); err != nil {
		return nil, err
	}
	if initialCausalHead == "" {
		return nil, fmt.Errorf("initial causal head is empty")
	}
	executor := &Executor{runner: runner, provider: provider, assurance: config, committed: map[closuregraph.ID]DerivationPermit{}, issuedPermits: map[closuregraph.ID]DerivationPermit{}, capabilityReceipts: map[closuregraph.ID]ProviderCapabilityReceipt{}, receipts: map[closuregraph.ID]DerivationReceipt{}, head: initialCausalHead}
	if config.Mode == AssurancePortable {
		if runner == nil {
			return nil, failure("portable_runner_missing", "portable mode requires the manager process runner")
		}
		if provider != nil {
			return nil, failure("assurance_evidence_mismatch", "portable mode cannot silently use a verified provider")
		}
		return executor, nil
	}
	if provider == nil {
		return nil, failure("verified_provider_missing", "verified mode requires a configured provider")
	}
	if !provider.LosslessObservation() {
		return nil, failure("verified_provider_unavailable", "provider observation is not lossless")
	}
	if err := provider.Identity().validate(config); err != nil {
		return nil, err
	}
	executor.boundary = provider
	return executor, nil
}

// Preflight establishes the operation assurance before any cache lookup or
// process start. Verified mode negotiates fresh provider health and exact
// capabilities here and never falls back to portable mode.
func (e *Executor) Preflight(ctx context.Context) (*AssuredOperation, error) {
	if e == nil {
		return nil, failure("verified_provider_unavailable", "executor is absent")
	}
	if e.assurance.Mode == AssurancePortable {
		return &AssuredOperation{executor: e, binding: portableAssuranceBinding()}, nil
	}
	if e.provider == nil {
		return nil, failure("verified_provider_missing", "verified mode requires a configured provider")
	}
	identity := e.provider.Identity()
	if err := identity.validate(e.assurance); err != nil {
		return nil, err
	}
	nonce, err := freshNonce()
	if err != nil {
		return nil, err
	}
	receipt, err := e.provider.Negotiate(ctx, nonce)
	if err != nil {
		return nil, failure("verified_provider_unavailable", "provider negotiation failed: %v", err)
	}
	if err = receipt.validate(identity, nonce, timeNow()); err != nil {
		return nil, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return nil, err
	}
	contract := VerifiedProviderContractID
	binding := AssuranceBinding{
		AssuranceMode: AssuranceVerified, PolicyID: VerifiedPolicyID,
		ExecutionPolicyID: VerifiedExecutionPolicyID, ProviderContract: &contract,
		Provider: &identity, CapabilityReceiptID: &receiptID,
		ActualCapabilities: append([]CapabilityEvidence(nil), receipt.Capabilities...),
	}
	return &AssuredOperation{executor: e, binding: binding, receipt: &receipt}, nil
}

// Binding returns a detached copy of the exact assurance established by this
// preflighted operation.
func (operation *AssuredOperation) Binding() AssuranceBinding {
	if operation == nil {
		return AssuranceBinding{}
	}
	binding := operation.binding
	binding.ActualCapabilities = append([]CapabilityEvidence(nil), binding.ActualCapabilities...)
	return binding
}

// CurrentCausalHead returns the executor-owned predecessor for the next permit.
// Callers cannot advance it without accepting an issued receipt.
func (operation *AssuredOperation) CurrentCausalHead() string {
	if operation == nil || operation.executor == nil {
		return ""
	}
	operation.executor.mu.Lock()
	defer operation.executor.mu.Unlock()
	return operation.executor.head
}

// Revalidate checks the retained verified provider immediately before a cache
// or process boundary. Portable operations have no provider state to refresh.
func (operation *AssuredOperation) Revalidate(ctx context.Context) error {
	if operation == nil || operation.executor == nil {
		return failure("assurance_evidence_mismatch", "assurance operation is absent")
	}
	if operation.binding.AssuranceMode == AssurancePortable {
		return nil
	}
	if operation.receipt == nil || operation.executor.provider == nil || operation.binding.Provider == nil {
		return failure("verified_provider_unavailable", "verified provider operation is incomplete")
	}
	if operation.executor.provider.Identity() != *operation.binding.Provider {
		return failure("verified_provider_identity_invalid", "provider identity changed after preflight")
	}
	refreshed, err := operation.executor.provider.Negotiate(ctx, operation.receipt.Nonce)
	if err != nil {
		return failure("verified_provider_unavailable", "provider health recheck failed: %v", err)
	}
	if err := refreshed.validate(*operation.binding.Provider, operation.receipt.Nonce, timeNow()); err != nil {
		return err
	}
	if !reflect.DeepEqual(refreshed.Capabilities, operation.binding.ActualCapabilities) {
		return failure("verified_capabilities_unsatisfied", "provider capability evidence differs from the preflighted operation")
	}
	return nil
}

// CacheInput binds one independently derived closure input to this preflight.
// A stale verified preflight cannot be used to address the protected cache.
func (operation *AssuredOperation) CacheInput(expected closuregraph.ExpectedCacheInput) (AssuredCacheInput, error) {
	if operation == nil || operation.executor == nil {
		return AssuredCacheInput{}, failure("assurance_evidence_mismatch", "assurance operation is absent")
	}
	if operation.receipt != nil {
		if err := operation.receipt.validate(*operation.binding.Provider, operation.receipt.Nonce, timeNow()); err != nil {
			return AssuredCacheInput{}, err
		}
	}
	input := AssuredCacheInput{Expected: expected, Binding: operation.binding}
	if _, err := input.ID(); err != nil {
		return AssuredCacheInput{}, err
	}
	return input, nil
}

// Commit records authority before the process-start seam.
func (e *Executor) Commit(permit DerivationPermit) (closuregraph.ID, error) {
	operation, err := e.Preflight(context.Background())
	if err != nil {
		return "", err
	}
	return operation.Commit(permit)
}

// Commit binds a permit to the exact preflight used for cache selection.
func (operation *AssuredOperation) Commit(permit DerivationPermit) (closuregraph.ID, error) {
	if operation == nil || operation.executor == nil {
		return "", failure("assurance_evidence_mismatch", "assurance operation is absent")
	}
	e := operation.executor
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.bindPermit(&permit, operation.receipt); err != nil {
		return "", err
	}
	if err := permit.Validate(); err != nil {
		return "", err
	}
	if permit.PreviousCausalHead != e.head {
		return "", failure("closure_derivation_unauthorized", "permit predecessor is not the current causal head")
	}
	id, err := permit.ID()
	if err != nil {
		return "", err
	}
	e.committed[id] = permit
	e.issuedPermits[id] = permit
	if operation.receipt != nil {
		e.capabilityReceipts[id] = *operation.receipt
	}
	return id, nil
}

// Execute consumes a permit committed by this exact preflighted operation.
func (operation *AssuredOperation) Execute(ctx context.Context, permitID closuregraph.ID, recheckToolchain func(context.Context) (ToolchainIdentity, error), inputs map[closuregraph.ID]AdmittedInput) (DerivationReceipt, error) {
	if operation == nil || operation.executor == nil {
		return DerivationReceipt{}, failure("assurance_evidence_mismatch", "assurance operation is absent")
	}
	return operation.executor.Execute(ctx, permitID, recheckToolchain, inputs)
}

// VerifyIssuedDerivationReceipt proves the receipt came from this operation's
// executor rather than from adapter-supplied evidence.
func (operation *AssuredOperation) VerifyIssuedDerivationReceipt(receipt DerivationReceipt) error {
	if operation == nil || operation.executor == nil {
		return failure("assurance_evidence_mismatch", "assurance operation is absent")
	}
	return operation.executor.VerifyIssuedDerivationReceipt(receipt)
}

// Execute rechecks the toolchain and every intake before crossing the process-start seam.
func (e *Executor) Execute(ctx context.Context, permitID closuregraph.ID, recheckToolchain func(context.Context) (ToolchainIdentity, error), inputs map[closuregraph.ID]AdmittedInput) (DerivationReceipt, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	permit, ok := e.committed[permitID]
	if !ok {
		return DerivationReceipt{}, failure("closure_derivation_unauthorized", "no committed permit")
	}
	// A committed permit is single-use, including failures. Checking and
	// consuming it while holding the causal-head lock is the atomic start seam.
	delete(e.committed, permitID)
	capabilityReceipt, hasCapabilityReceipt := e.capabilityReceipts[permitID]
	delete(e.capabilityReceipts, permitID)
	if permit.PreviousCausalHead != e.head {
		return DerivationReceipt{}, failure("closure_derivation_unauthorized", "stale permit predecessor at process-start seam")
	}
	if recheckToolchain == nil {
		return DerivationReceipt{}, failure("closure_derivation_unauthorized", "toolchain recheck is absent")
	}
	if e.assurance.Mode == AssuranceVerified {
		if !hasCapabilityReceipt || e.provider.Identity() != *permit.Provider {
			return DerivationReceipt{}, failure("verified_provider_identity_invalid", "provider identity changed before execution")
		}
		refreshed, providerErr := e.provider.Negotiate(ctx, capabilityReceipt.Nonce)
		if providerErr != nil {
			return DerivationReceipt{}, failure("verified_provider_unavailable", "provider health recheck failed: %v", providerErr)
		}
		if providerErr = refreshed.validate(*permit.Provider, capabilityReceipt.Nonce, timeNow()); providerErr != nil {
			return DerivationReceipt{}, providerErr
		}
		originalID, _ := capabilityReceipt.ID()
		refreshedID, _ := refreshed.ID()
		if originalID != refreshedID || permit.CapabilityReceiptID == nil || *permit.CapabilityReceiptID != originalID {
			return DerivationReceipt{}, failure("verified_capabilities_unsatisfied", "provider capability evidence drifted before execution")
		}
	}
	before, err := recheckToolchain(ctx)
	if err != nil {
		return DerivationReceipt{}, err
	}
	if before.Fingerprint != permit.ToolchainFingerprint || before.ExecutableSHA256 != permit.ExecutableSHA256 {
		return DerivationReceipt{}, failure("artifact_toolchain_identity_changed", "toolchain changed before process start")
	}
	replays := make([]ReplayInput, len(permit.InputMounts))
	for i, id := range permit.AdmittedInputReceiptIDs {
		input, ok := inputs[id]
		if !ok {
			return DerivationReceipt{}, failure("closure_derivation_unauthorized", "permit names an unadmitted input")
		}
		if err = input.recheck(id); err != nil {
			return DerivationReceipt{}, err
		}
		replays[i] = ReplayInput{ReceiptID: id, MountPath: permit.InputMounts[i].Path, input: input}
	}
	var audit Audit
	if e.assurance.Mode == AssurancePortable {
		result, runErr := e.runner.Run(ctx, ExecutionRequest{Permit: permit, Inputs: replays})
		if runErr != nil {
			return DerivationReceipt{}, runErr
		}
		if result.cleanup != nil {
			defer result.cleanup()
		}
		evidenceRoot := result.EvidenceRoot
		if evidenceRoot == "" {
			evidenceRoot = result.OutputRoot
		}
		outputs, outputErr := observePortableOutputs(evidenceRoot, permit.ExpectedEvidence, permit.ResourceLimits.OutputBytes, evidenceRoot == result.OutputRoot)
		if outputErr != nil {
			return DerivationReceipt{}, outputErr
		}
		evidence := make([]string, len(outputs))
		for index := range outputs {
			evidence[index] = outputs[index].Path
		}
		audit = Audit{Executable: permit.Executable, CWD: permit.CWD, Argv: append([]string{}, permit.Argv...), Environment: cloneStringMap(permit.Environment), Processes: []string{}, Reads: []string{}, Writes: []string{}, Evidence: evidence, Network: "not-observed", ExitCode: result.ExitCode, Outputs: outputs}
	} else {
		audit, err = e.boundary.EnforceAndObserve(ctx, ExecutionRequest{Permit: permit, Inputs: replays})
		if err != nil {
			return DerivationReceipt{}, err
		}
	}
	for _, replay := range replays {
		if err := replay.VerifyAtUse(); err != nil {
			return DerivationReceipt{}, err
		}
	}
	after, checkErr := recheckToolchain(ctx)
	if checkErr != nil {
		return DerivationReceipt{}, checkErr
	}
	if after != before {
		return DerivationReceipt{}, failure("artifact_toolchain_identity_changed", "toolchain changed during invocation")
	}
	if e.assurance.Mode == AssuranceVerified && audit.Network != "none" {
		return DerivationReceipt{}, failure("closure_network_attempted", "verified provider observed a network attempt")
	}
	if diff := auditDifference(permit, audit, e.assurance.Mode); diff != "" {
		return DerivationReceipt{}, failure("closure_derivation_drift", "observed %s differs from permit", diff)
	}
	receipt := DerivationReceipt{SchemaID: SchemaDerivationReceipt, AssuranceMode: permit.AssuranceMode, PolicyID: permit.PolicyID, ExecutionPolicyID: permit.ExecutionPolicyID, ProviderContract: permit.ProviderContract, Provider: permit.Provider, CapabilityReceiptID: permit.CapabilityReceiptID, ActualCapabilities: append([]CapabilityEvidence(nil), permit.ActualCapabilities...), PermitID: permitID, BeforeFingerprint: before.Fingerprint, AfterFingerprint: after.Fingerprint, Audit: audit, InvocationSubtype: permit.InvocationSubtype, ResourceLimits: permit.ResourceLimits, ResourceLimitID: permit.ResourceLimitID, ExpectedEvidence: append([]EvidenceRequirement(nil), permit.ExpectedEvidence...), EvidenceSchemaID: permit.EvidenceSchemaID, Outputs: append([]DerivationOutput(nil), audit.Outputs...), Diagnostics: []DerivationDiagnostic{}, Decision: "success"}
	receipt.NextCausalHead, err = receipt.deriveNextCausalHead()
	if err != nil {
		return DerivationReceipt{}, err
	}
	rid, err := receipt.ID()
	if err != nil {
		return DerivationReceipt{}, err
	}
	e.receipts[rid] = receipt
	e.head = string(receipt.NextCausalHead)
	return receipt, nil
}

// VerifyIssuedDerivationReceipt proves that a canonical receipt was issued by
// this executor's committed causal chain rather than reconstructed by a caller.
func (e *Executor) VerifyIssuedDerivationReceipt(receipt DerivationReceipt) error {
	if e == nil {
		return failure("closure_derivation_unauthorized", "derivation executor is absent")
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	id, err := receipt.ID()
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	issued, ok := e.receipts[id]
	if !ok || !reflect.DeepEqual(issued, receipt) {
		return failure("closure_derivation_unauthorized", "derivation receipt was not issued by this executor")
	}
	return nil
}

// IssuedDerivationPermit returns the exact assurance-bound permit committed by
// this executor. Commit accepts an adapter declaration by value and binds the
// manager-selected assurance fields before issuing its ID; callers that must
// prove a complete permit/receipt chain use this accessor rather than
// reconstructing those manager-owned fields.
func (e *Executor) IssuedDerivationPermit(id closuregraph.ID) (DerivationPermit, error) {
	if e == nil || !id.Valid() {
		return DerivationPermit{}, failure("closure_derivation_unauthorized", "derivation permit identity is invalid")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	permit, ok := e.issuedPermits[id]
	if !ok {
		return DerivationPermit{}, failure("closure_derivation_unauthorized", "derivation permit was not issued by this executor")
	}
	return permit, nil
}

// VerifyIssuedDerivationChain proves both the complete committed permit and
// its resulting receipt were issued by this executor. This is the authority
// seam used by narrowly scoped local-output authorization issuers.
func (e *Executor) VerifyIssuedDerivationChain(permit DerivationPermit, receipt DerivationReceipt) error {
	if e == nil {
		return failure("closure_derivation_unauthorized", "derivation executor is absent")
	}
	permitID, err := permit.ID()
	if err != nil || permitID != receipt.PermitID {
		return failure("closure_derivation_unauthorized", "derivation receipt does not name the exact permit")
	}
	if err = receipt.Validate(); err != nil {
		return err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	issuedPermit, permitOK := e.issuedPermits[permitID]
	issuedReceipt, receiptOK := e.receipts[receiptID]
	if !permitOK || !receiptOK || !reflect.DeepEqual(issuedPermit, permit) || !reflect.DeepEqual(issuedReceipt, receipt) {
		return failure("closure_derivation_unauthorized", "derivation permit or receipt was not issued by this executor")
	}
	return nil
}

func auditDifference(p DerivationPermit, a Audit, mode AssuranceMode) string {
	if a.Executable != p.Executable {
		return "executable"
	}
	if a.CWD != p.CWD {
		return "cwd"
	}
	if !reflect.DeepEqual(a.Argv, p.Argv) {
		return "argv"
	}
	if !reflect.DeepEqual(a.Environment, p.Environment) {
		return "environment"
	}
	if mode == AssuranceVerified && !reflect.DeepEqual(a.Processes, p.AllowedProcesses) {
		return "process set"
	}
	if mode == AssuranceVerified && !reflect.DeepEqual(a.Reads, p.ReadRoots) {
		return "read set"
	}
	if mode == AssuranceVerified && !reflect.DeepEqual(a.Writes, p.WriteRoots) {
		return "write set"
	}
	expectedPaths := make([]string, len(p.ExpectedEvidence))
	for i, expected := range p.ExpectedEvidence {
		expectedPaths[i] = expected.Path
	}
	if !reflect.DeepEqual(a.Evidence, expectedPaths) {
		return "evidence outputs"
	}
	if len(a.Outputs) != len(p.ExpectedEvidence) {
		return "evidence records"
	}
	for i, output := range a.Outputs {
		expected := p.ExpectedEvidence[i]
		if output.Path != expected.Path || output.SchemaID != expected.SchemaID || output.ArtifactManifestID != expected.ArtifactManifestID || output.validate() != nil {
			return "evidence records"
		}
	}
	if (mode == AssuranceVerified && a.Network != "none") || (mode == AssurancePortable && a.Network != "not-observed") {
		return "network"
	}
	if a.ExitCode != 0 {
		return "exit status"
	}
	return ""
}

func (e *Executor) bindPermit(permit *DerivationPermit, capabilityReceipt *ProviderCapabilityReceipt) error {
	if permit.AssuranceMode != "" || permit.PolicyID != "" || permit.ExecutionPolicyID != "" || permit.ProviderContract != nil || permit.Provider != nil || permit.CapabilityReceiptID != nil || permit.ActualCapabilities != nil {
		return failure("assurance_evidence_mismatch", "caller supplied pre-bound assurance evidence")
	}
	permit.WorkCopies = append([]WorkCopy{}, permit.WorkCopies...)
	if e.assurance.Mode == AssurancePortable {
		permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID = AssurancePortable, PortablePolicyID, PortableExecutionPolicyID
		permit.ProviderContract, permit.Provider, permit.CapabilityReceiptID = nil, nil, nil
		permit.ActualCapabilities = append([]CapabilityEvidence(nil), portableCapabilities...)
		return nil
	}
	if capabilityReceipt == nil {
		return failure("verified_capabilities_unsatisfied", "provider capability receipt is missing")
	}
	identity := capabilityReceipt.Provider
	receiptID, err := capabilityReceipt.ID()
	if err != nil {
		return err
	}
	contract := VerifiedProviderContractID
	permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID = AssuranceVerified, VerifiedPolicyID, VerifiedExecutionPolicyID
	permit.ProviderContract, permit.Provider, permit.CapabilityReceiptID = &contract, &identity, &receiptID
	permit.ActualCapabilities = append([]CapabilityEvidence(nil), capabilityReceipt.Capabilities...)
	return nil
}

var timeNow = time.Now

func observePortableOutputs(root string, expected []EvidenceRequirement, outputLimit int64, rejectUndeclared bool) ([]DerivationOutput, error) {
	absolute, err := filepath.Abs(root)
	if err != nil || root == "" {
		return nil, failure("closure_derivation_drift", "portable output root is invalid")
	}
	rootInfo, err := os.Lstat(absolute)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, failure("closure_derivation_drift", "portable output root is not a real directory")
	}
	seen := map[string]bool{}
	if rejectUndeclared {
		err = filepath.WalkDir(absolute, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if current == absolute {
				return nil
			}
			rel, relErr := filepath.Rel(absolute, current)
			if relErr != nil {
				return relErr
			}
			logical := filepath.ToSlash(rel)
			if entry.Type()&fs.ModeSymlink != 0 {
				return failure("closure_write_undeclared", "portable output contains a link")
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return failure("closure_write_undeclared", "portable output contains a special node")
			}
			seen[logical] = true
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		for _, requirement := range expected {
			seen[requirement.Path] = true
		}
	}
	outputs := make([]DerivationOutput, len(expected))
	var totalSize int64
	for index, requirement := range expected {
		if !seen[requirement.Path] {
			return nil, failure("closure_derivation_drift", "declared output %s is missing", requirement.Path)
		}
		outputPath := filepath.Join(absolute, filepath.FromSlash(requirement.Path))
		before, statErr := os.Lstat(outputPath)
		if statErr != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 {
			return nil, failure("closure_derivation_drift", "declared output %s is not a regular file", requirement.Path)
		}
		payload, readErr := os.ReadFile(outputPath) // #nosec G304 -- requirement is a canonical relative permit path below a link-free manager output root.
		if readErr != nil {
			return nil, readErr
		}
		after, statErr := os.Lstat(outputPath)
		if statErr != nil || !os.SameFile(before, after) || before.Size() != after.Size() || int64(len(payload)) != after.Size() {
			return nil, failure("closure_derivation_drift", "declared output %s changed during verification", requirement.Path)
		}
		totalSize += int64(len(payload))
		if totalSize > outputLimit {
			return nil, failure("portable_output_limit", "declared output set exceeds its byte bound")
		}
		outputs[index] = DerivationOutput{Path: requirement.Path, SchemaID: requirement.SchemaID, ArtifactManifestID: requirement.ArtifactManifestID, SHA256: digestBytes(payload), Size: int64(len(payload))}
		delete(seen, requirement.Path)
	}
	if len(seen) != 0 {
		return nil, failure("closure_write_undeclared", "portable output root contains undeclared files")
	}
	return outputs, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

// CanonicalAudit sorts set-like observed fields before receipt construction.
func CanonicalAudit(a Audit) Audit {
	a.Processes = sortedCopy(a.Processes)
	a.Reads = sortedCopy(a.Reads)
	a.Writes = sortedCopy(a.Writes)
	a.Evidence = sortedCopy(a.Evidence)
	sort.Slice(a.Outputs, func(i, j int) bool { return a.Outputs[i].Path < a.Outputs[j].Path })
	sort.Strings(a.Processes)
	return a
}
