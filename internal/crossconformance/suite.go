package crossconformance

import (
	"fmt"
	"sort"
	"strings"
)

// PathID names one delivered adapter path. The six paths below are exactly the
// ecosystems this delivery cycle implements; Go remains the untouched baseline
// and every other ecosystem is deferred.
type PathID string

const (
	// PathRust is the rust-source-v1 Cargo path.
	PathRust PathID = "rust"
	// PathNPM is the npm lock and private-cache profile.
	PathNPM PathID = "npm"
	// PathPNPM is the pnpm lock and private-store profile.
	PathPNPM PathID = "pnpm"
	// PathYarnClassic is the Yarn 1.x lock and offline-mirror profile.
	PathYarnClassic PathID = "yarn-classic"
	// PathYarnModern is the modern Yarn lock, cache, and linker profile.
	PathYarnModern PathID = "yarn-modern"
	// PathSwiftPM is the swiftpm-source-v1 Swift and C-family path.
	PathSwiftPM PathID = "swiftpm"
)

// DeliveredPaths is the closed set every normative obligation runs against.
func DeliveredPaths() []PathID {
	return []PathID{PathRust, PathNPM, PathPNPM, PathYarnClassic, PathYarnModern, PathSwiftPM}
}

// Obligation is one normative semantic requirement of the accepted shared
// contract. Every delivered path proves every obligation; the record family a
// path uses to prove one may differ, the requirement may not.
type Obligation string

const (
	// ObligationSelectionNeutralCapture requires that no exact target
	// platform, toolchain component, targets edge, or uses_tool edge occurs
	// inside the selection-neutral capture.
	ObligationSelectionNeutralCapture Obligation = "capture.selection_neutral"
	// ObligationCaptureStableAcrossTargets requires one exact capture identity
	// for two different requested targets over the same source inputs.
	ObligationCaptureStableAcrossTargets Obligation = "capture.stable_across_targets"
	// ObligationBindingOwnsTargetAuthority requires the exact target platform
	// and every concrete tool identity to enter only through the selection
	// overlay, bound by an explicit targets edge where the path emits edges.
	ObligationBindingOwnsTargetAuthority Obligation = "binding.target_authority"
	// ObligationSelectionDivergesPerTarget requires the selection, binding,
	// active graph, and plan identities to differ between two targets.
	ObligationSelectionDivergesPerTarget Obligation = "binding.diverges_per_target"
	// ObligationDeterministicProjection requires repeated projection of the
	// same inputs to produce byte-identical identities.
	ObligationDeterministicProjection Obligation = "records.deterministic"
	// ObligationCausalEvidenceChain requires every emitted checkpoint to name
	// its exact predecessor, requires C5 to add no graph record, and requires
	// every pre-C5 evidence derivation to answer with a causal receipt. A path
	// that derives evidence by running a manager and reports neither fails.
	ObligationCausalEvidenceChain Obligation = "evidence.causal_chain"
	// ObligationSharedArtifactAdmission requires the shared artifact corpus to
	// return one class, decision, and primary diagnostic through every path.
	ObligationSharedArtifactAdmission Obligation = "artifact.shared_admission"
)

// Obligations returns the closed normative suite in stable order.
func Obligations() []Obligation {
	return []Obligation{
		ObligationSelectionNeutralCapture,
		ObligationCaptureStableAcrossTargets,
		ObligationBindingOwnsTargetAuthority,
		ObligationSelectionDivergesPerTarget,
		ObligationDeterministicProjection,
		ObligationCausalEvidenceChain,
		ObligationSharedArtifactAdmission,
	}
}

// CheckpointLink is one emitted checkpoint in the order the path emitted it.
type CheckpointLink struct {
	Name       string
	Identity   string
	Previous   string
	PayloadKey []string
}

// TargetProjection is what one adapter path reports after closing one exact
// target selection over one fixed set of source inputs. Paths whose selected
// overlay is not a closuregraph selection binding report their equivalent
// selection-bound identity in BindingIdentity and leave the record-kind
// censuses empty; the obligations below say so explicitly instead of skipping.
type TargetProjection struct {
	Path        PathID
	TargetLabel string

	CaptureIdentity   string
	SelectionIdentity string
	BindingIdentity   string
	ActiveIdentity    string
	PlanIdentity      string

	CaptureNodeKinds map[string]int
	CaptureEdgeKinds map[string]int
	BindingNodeKinds map[string]int
	BindingEdgeKinds map[string]int

	// TargetPlatformIdentity is the exact bound target. For a path with no
	// platform node record it is the exact target triple, which must still
	// never occur in the capture identity's inputs.
	TargetPlatformIdentity string
	// ExplicitTargetEdges counts targets edges that reach the bound target
	// platform. A path that emits binding edges must report at least one.
	ExplicitTargetEdges int
	// EmitsBindingRecords tells the suite whether the record-kind censuses and
	// ExplicitTargetEdges are meaningful for this path.
	EmitsBindingRecords bool
	// ToolIdentities are the exact toolchain fingerprints bound for this
	// target. They must never be inputs to the capture identity.
	ToolIdentities []string

	Checkpoints []CheckpointLink
	// DerivationReceipts are the causal receipts the path issued for evidence
	// it derived by running a manager before C5.
	DerivationReceipts []string
}

// CheckSelectionNeutralCapture proves ObligationSelectionNeutralCapture.
func CheckSelectionNeutralCapture(projection TargetProjection) error {
	if projection.CaptureIdentity == "" {
		return fmt.Errorf("%s/%s reported no capture identity", projection.Path, projection.TargetLabel)
	}
	if !ValidIdentity(projection.CaptureIdentity) {
		return fmt.Errorf("%s capture identity %q is not domain separated", projection.Path, projection.CaptureIdentity)
	}
	for kind := range projection.CaptureNodeKinds {
		if selectionOnlyNodeKinds[kind] {
			return fmt.Errorf("%s capture retains selection-specific node kind %q", projection.Path, kind)
		}
	}
	for kind := range projection.CaptureEdgeKinds {
		if selectionOnlyEdgeKinds[kind] {
			return fmt.Errorf("%s capture retains selection-specific edge kind %q", projection.Path, kind)
		}
	}
	return nil
}

// CheckBindingOwnsTargetAuthority proves ObligationBindingOwnsTargetAuthority.
func CheckBindingOwnsTargetAuthority(projection TargetProjection) error {
	if projection.TargetPlatformIdentity == "" {
		return fmt.Errorf("%s/%s bound no exact target", projection.Path, projection.TargetLabel)
	}
	if projection.BindingIdentity == "" {
		return fmt.Errorf("%s/%s reported no selection-bound identity", projection.Path, projection.TargetLabel)
	}
	if projection.BindingIdentity == projection.CaptureIdentity {
		return fmt.Errorf("%s/%s binds the target into its own capture identity", projection.Path, projection.TargetLabel)
	}
	if len(projection.ToolIdentities) == 0 {
		return fmt.Errorf("%s/%s bound no exact tool identity", projection.Path, projection.TargetLabel)
	}
	if !projection.EmitsBindingRecords {
		return nil
	}
	if projection.BindingNodeKinds["target_platform"] != 1 {
		return fmt.Errorf("%s/%s binds %d target platform nodes, want exactly one",
			projection.Path, projection.TargetLabel, projection.BindingNodeKinds["target_platform"])
	}
	for kind := range projection.BindingNodeKinds {
		if !bindingNodeKinds[kind] {
			return fmt.Errorf("%s/%s binds forbidden node kind %q", projection.Path, projection.TargetLabel, kind)
		}
	}
	for kind := range projection.BindingEdgeKinds {
		if !bindingEdgeKinds[kind] {
			return fmt.Errorf("%s/%s binds forbidden edge kind %q", projection.Path, projection.TargetLabel, kind)
		}
	}
	if projection.ExplicitTargetEdges < 1 {
		return fmt.Errorf("%s/%s binds its target platform without an explicit targets edge", projection.Path, projection.TargetLabel)
	}
	return nil
}

// CheckTargetDivergence proves ObligationCaptureStableAcrossTargets and
// ObligationSelectionDivergesPerTarget from two projections of one input set.
func CheckTargetDivergence(first, second TargetProjection) error {
	if first.Path != second.Path {
		return fmt.Errorf("cannot compare %s with %s", first.Path, second.Path)
	}
	if first.TargetLabel == second.TargetLabel {
		return fmt.Errorf("%s compared one target with itself", first.Path)
	}
	if first.CaptureIdentity != second.CaptureIdentity {
		return fmt.Errorf("%s changed its selection-neutral capture between %s and %s: %s != %s",
			first.Path, first.TargetLabel, second.TargetLabel, first.CaptureIdentity, second.CaptureIdentity)
	}
	if first.TargetPlatformIdentity == second.TargetPlatformIdentity {
		return fmt.Errorf("%s bound one target for %s and %s", first.Path, first.TargetLabel, second.TargetLabel)
	}
	if first.BindingIdentity == "" || second.BindingIdentity == "" {
		return fmt.Errorf("%s reported no selection-bound identity", first.Path)
	}
	if first.BindingIdentity == second.BindingIdentity {
		return fmt.Errorf("%s reused one selection-bound identity for %s and %s", first.Path, first.TargetLabel, second.TargetLabel)
	}
	// A path that also emits an active projection or a build plan must move
	// both with the target. Absent records are reported, not assumed equal.
	for _, pair := range []struct {
		field, left, right string
	}{
		{"active", first.ActiveIdentity, second.ActiveIdentity},
		{"plan", first.PlanIdentity, second.PlanIdentity},
	} {
		if pair.left == "" && pair.right == "" {
			continue
		}
		if pair.left == "" || pair.right == "" {
			return fmt.Errorf("%s emitted a %s identity for one target only", first.Path, pair.field)
		}
		if pair.left == pair.right {
			return fmt.Errorf("%s reused one %s identity for %s and %s", first.Path, pair.field, first.TargetLabel, second.TargetLabel)
		}
	}
	return nil
}

// CheckDeterministicProjection proves ObligationDeterministicProjection: two
// independent runs over the same inputs must agree on every identity.
func CheckDeterministicProjection(runs []TargetProjection) error {
	if len(runs) < 2 {
		return fmt.Errorf("determinism needs at least two runs, got %d", len(runs))
	}
	first := runs[0]
	for _, run := range runs[1:] {
		for _, pair := range []struct {
			field, left, right string
		}{
			{"capture", first.CaptureIdentity, run.CaptureIdentity},
			{"selection", first.SelectionIdentity, run.SelectionIdentity},
			{"binding", first.BindingIdentity, run.BindingIdentity},
			{"active", first.ActiveIdentity, run.ActiveIdentity},
			{"plan", first.PlanIdentity, run.PlanIdentity},
		} {
			if pair.left != pair.right {
				return fmt.Errorf("%s %s identity is not deterministic: %s != %s", first.Path, pair.field, pair.left, pair.right)
			}
		}
	}
	return nil
}

// CheckCausalEvidenceChain proves ObligationCausalEvidenceChain.
func CheckCausalEvidenceChain(projection TargetProjection) error {
	if len(projection.Checkpoints)+len(projection.DerivationReceipts) == 0 {
		return fmt.Errorf("%s/%s reported neither a committed checkpoint nor a causal receipt", projection.Path, projection.TargetLabel)
	}
	seen := map[string]bool{}
	for _, receipt := range projection.DerivationReceipts {
		if !ValidIdentity(receipt) {
			return fmt.Errorf("%s issued a malformed causal receipt %q", projection.Path, receipt)
		}
		if seen[receipt] {
			return fmt.Errorf("%s issued causal receipt %s twice", projection.Path, receipt)
		}
		seen[receipt] = true
	}
	var previous CheckpointLink
	for index, checkpoint := range projection.Checkpoints {
		if !ValidIdentity(checkpoint.Identity) {
			return fmt.Errorf("%s checkpoint %s has malformed identity %q", projection.Path, checkpoint.Name, checkpoint.Identity)
		}
		if seen[checkpoint.Identity] {
			return fmt.Errorf("%s emitted checkpoint identity %s twice", projection.Path, checkpoint.Identity)
		}
		seen[checkpoint.Identity] = true
		if index == 0 {
			previous = checkpoint
			continue
		}
		if checkpoint.Previous != previous.Identity {
			return fmt.Errorf("%s checkpoint %s chains to %s, want its emitted predecessor %s (%s)",
				projection.Path, checkpoint.Name, checkpoint.Previous, previous.Identity, previous.Name)
		}
		if checkpoint.Name == "C5.plan" {
			for _, key := range checkpoint.PayloadKey {
				if key != "build_plan_id" {
					return fmt.Errorf("%s C5 carries %q; C5 adds no graph record", projection.Path, key)
				}
			}
		}
		previous = checkpoint
	}
	return nil
}

// CheckCaptureExcludesToolIdentities proves that no exact tool fingerprint the
// binding introduced is spelled inside the capture identity's own record
// census. It complements CheckSelectionNeutralCapture for paths that publish
// no typed node kinds.
func CheckCaptureExcludesToolIdentities(projection TargetProjection, captureRecordText string) error {
	for _, fingerprint := range projection.ToolIdentities {
		if fingerprint != "" && strings.Contains(captureRecordText, fingerprint) {
			return fmt.Errorf("%s capture text contains bound tool identity %s", projection.Path, fingerprint)
		}
	}
	if projection.TargetPlatformIdentity != "" && strings.Contains(captureRecordText, projection.TargetPlatformIdentity) {
		return fmt.Errorf("%s capture text contains the bound target %s", projection.Path, projection.TargetPlatformIdentity)
	}
	return nil
}

// RejectionVector is one published fail-closed requirement. Every delivered
// path runs every vector whose family it can express, and a vector always
// requires the same three things: a stable diagnostic, no affected process,
// and no publication.
type RejectionVector struct {
	ID          string
	Family      string
	Requirement string
	// Codes is the closed set of stable diagnostics that may answer this
	// vector. Human text is never a machine interface.
	Codes []string
	// OwnedBy names the package whose accepted suite proves this vector when
	// its failing condition cannot be constructed through the delivered
	// adapters' exported seams. Those conditions need a live verified
	// execution provider or a sealed in-package authority, so an integration
	// package can only reach them by forging evidence, which would prove
	// nothing. An empty value means the cross-adapter suite proves it.
	OwnedBy []string
}

// CrossDrivable reports whether this suite must prove the vector itself.
func (vector RejectionVector) CrossDrivable() bool { return len(vector.OwnedBy) == 0 }

// Rejection families.
const (
	// FamilyGraph covers graph schema, reference, binding, and cycle failures.
	FamilyGraph = "graph"
	// FamilyArtifact covers recursive byte admission failures.
	FamilyArtifact = "artifact"
	// FamilyIdentity covers target, toolchain, and runtime drift.
	FamilyIdentity = "identity"
	// FamilyExecution covers network, process, input, and write violations.
	FamilyExecution = "execution"
	// FamilyOutput covers output and receipt failures.
	FamilyOutput = "output"
)

// RejectionVectors returns the published cross-adapter rejection matrix.
func RejectionVectors() []RejectionVector {
	return []RejectionVector{
		{ID: "binding-duplicate-record", Family: FamilyGraph, Requirement: "A binding that names one record twice rejects.", Codes: []string{"closure_graph_schema_unsupported", "closure_graph_reference_invalid", "closure_graph_incomplete"}},
		{ID: "binding-dangling-reference", Family: FamilyGraph, Requirement: "A binding edge whose record is absent from capture and binding rejects.", Codes: []string{"closure_graph_incomplete", "closure_graph_reference_invalid"}},
		{ID: "binding-wrong-kind", Family: FamilyGraph, Requirement: "A binding that carries a non-platform, non-toolchain node rejects.", Codes: []string{"closure_graph_reference_invalid", "closure_graph_schema_unsupported"}},
		{ID: "binding-replaces-capture", Family: FamilyGraph, Requirement: "A binding that replaces a captured record rejects.", Codes: []string{"closure_graph_reference_invalid"}},
		{ID: "binding-missing-target", Family: FamilyGraph, Requirement: "A selection whose target platform is not bound rejects.", Codes: []string{"closure_graph_incomplete", "closure_graph_reference_invalid"}},
		{ID: "build-cycle", Family: FamilyGraph, Requirement: "A cycle in the execution projection rejects.", Codes: []string{"closure_build_cycle"}},
		{ID: "compiled-dependency-bytes", Family: FamilyArtifact, Requirement: "Compiled dependency bytes reject before any manager process starts.", Codes: []string{"artifact_compiled_dependency_forbidden"}},
		{ID: "opaque-dependency-bytes", Family: FamilyArtifact, Requirement: "Bytes with no complete allowed interpretation reject.", Codes: []string{"artifact_opaque_dependency_forbidden", "artifact_type_ambiguous"}},
		{ID: "verified-binary-unavailable", Family: FamilyArtifact, Requirement: "A binary-admission request rejects while the capability is absent.", Codes: []string{"artifact_binary_admission_unavailable"}},
		{ID: "integrity-mismatch", Family: FamilyArtifact, Requirement: "Package bytes that differ from the frozen lock integrity reject before materialization.", Codes: []string{"closure_integrity_mismatch", "closure_metadata_mismatch"}},
		{ID: "offline-input-missing", Family: FamilyExecution, Requirement: "A captured input that is unavailable at replay rejects before execution.", Codes: []string{"closure_offline_input_missing", "swiftpm_dependency_mirror_missing", "swiftpm_dependency_origin_unsupported"}},
		{ID: "target-identity-drift", Family: FamilyIdentity, Requirement: "A target that differs from the closed binding rejects before execution.", Codes: []string{"closure_target_identity_changed", "rust_target_unsupported", "swiftpm_target_platform_unsupported"}},
		{ID: "toolchain-identity-drift", Family: FamilyIdentity, Requirement: "A tool whose identity changed between checkpoint and use rejects.", Codes: []string{"artifact_toolchain_identity_changed", "closure_runtime_identity_changed", "closure_derivation_drift", "swiftpm_manifest_replay_drift"}},
		{ID: "undeclared-process", Family: FamilyExecution, Requirement: "Evidence that no committed permit authorized rejects.", Codes: []string{"closure_process_undeclared", "closure_derivation_unauthorized", "closure_derivation_drift"}},
		{ID: "undeclared-input", Family: FamilyExecution, Requirement: "An input outside the admitted closure rejects before it can be read.", Codes: []string{"closure_input_undeclared", "closure_build_dependency_unlocked", "artifact_generated_input_undeclared", "swiftpm_local_dependency_outside_closure", "swiftpm_source_inventory_drift", "swiftpm_header_input_undeclared", "rust_undeclared_input"}},
		{ID: "unreceipted-output", Family: FamilyOutput, Requirement: "Output bytes with no causal receipt reject.", Codes: []string{"artifact_local_output_unreceipted"}},
		{
			ID: "network-attempted", Family: FamilyExecution,
			Requirement: "Any post-capture network attempt rejects and issues no receipt.",
			Codes:       []string{"closure_network_attempted"},
			OwnedBy:     []string{"internal/closureexec", "internal/npmsource", "internal/rustsource", "internal/swiftpmbuild"},
		},
		{
			ID: "undeclared-write", Family: FamilyExecution,
			Requirement: "A write outside the committed write set rejects and issues no receipt.",
			Codes:       []string{"closure_write_undeclared"},
			OwnedBy:     []string{"internal/closureexec", "internal/rustsource", "internal/swiftpmbuild"},
		},
		{
			ID: "output-drift", Family: FamilyOutput,
			Requirement: "A published artifact that differs from its exact expectation rejects.",
			Codes:       []string{"artifact_local_output_drift", "closure_generated_output_drift"},
			OwnedBy:     []string{"internal/artifactpolicy", "internal/swiftpmbuild", "internal/nodesource"},
		},
	}
}

// RejectionOutcome is what one path observed for one vector.
type RejectionOutcome struct {
	Vector string
	Path   PathID
	// Err is the failure the adapter returned. A nil error is always a
	// conformance failure for a rejection vector.
	Err error
	// Code is the stable diagnostic the adapter reported.
	Code string
	// ProcessStarts counts affected manager or build processes that launched.
	ProcessStarts int
	// Published reports whether any protected output or cache entry landed.
	Published bool
}

// CheckRejection proves one rejection vector for one path.
func CheckRejection(vector RejectionVector, outcome RejectionOutcome) error {
	if outcome.Err == nil {
		return fmt.Errorf("%s/%s admitted the %s vector", outcome.Path, vector.ID, vector.Family)
	}
	if !containsString(vector.Codes, outcome.Code) {
		return fmt.Errorf("%s/%s reported %q, want one of %s", outcome.Path, vector.ID, outcome.Code, strings.Join(vector.Codes, ", "))
	}
	if outcome.ProcessStarts != 0 {
		return fmt.Errorf("%s/%s started %d affected processes before failing closed", outcome.Path, vector.ID, outcome.ProcessStarts)
	}
	if outcome.Published {
		return fmt.Errorf("%s/%s published an output after failing closed", outcome.Path, vector.ID)
	}
	return nil
}

// Coverage records which path proved which obligation and vector. A harness
// builds one and the suite refuses an incomplete matrix, so a path that stops
// running an obligation is a failure rather than a silent gap.
type Coverage struct {
	obligations map[Obligation]map[PathID]bool
	rejections  map[string]map[PathID]bool
}

// NewCoverage creates an empty coverage matrix.
func NewCoverage() *Coverage {
	return &Coverage{obligations: map[Obligation]map[PathID]bool{}, rejections: map[string]map[PathID]bool{}}
}

// RecordObligation marks one obligation proved for one path.
func (coverage *Coverage) RecordObligation(obligation Obligation, path PathID) {
	if coverage.obligations[obligation] == nil {
		coverage.obligations[obligation] = map[PathID]bool{}
	}
	coverage.obligations[obligation][path] = true
}

// RecordRejection marks one rejection vector proved for one path.
func (coverage *Coverage) RecordRejection(vector string, path PathID) {
	if coverage.rejections[vector] == nil {
		coverage.rejections[vector] = map[PathID]bool{}
	}
	coverage.rejections[vector][path] = true
}

// MissingObligations lists every obligation and path pair that was never
// proved, in stable order.
func (coverage *Coverage) MissingObligations() []string {
	missing := []string{}
	for _, obligation := range Obligations() {
		for _, path := range DeliveredPaths() {
			if !coverage.obligations[obligation][path] {
				missing = append(missing, string(obligation)+"/"+string(path))
			}
		}
	}
	sort.Strings(missing)
	return missing
}

// UncoveredRejections lists every cross-drivable rejection vector that no path
// proved, in stable order. A vector delegated to an owning package is not
// counted here; DelegatedRejections reports those separately.
func (coverage *Coverage) UncoveredRejections() []string {
	missing := []string{}
	for _, vector := range RejectionVectors() {
		if vector.CrossDrivable() && len(coverage.rejections[vector.ID]) == 0 {
			missing = append(missing, vector.ID)
		}
	}
	sort.Strings(missing)
	return missing
}

// DelegatedRejections lists every published vector this suite does not drive,
// with the packages whose accepted suites own it.
func (coverage *Coverage) DelegatedRejections() []string {
	delegated := []string{}
	for _, vector := range RejectionVectors() {
		if !vector.CrossDrivable() {
			delegated = append(delegated, vector.ID+"="+strings.Join(vector.OwnedBy, "+"))
		}
	}
	sort.Strings(delegated)
	return delegated
}

// RejectionPaths returns how many paths proved each vector, in stable order.
func (coverage *Coverage) RejectionPaths() []string {
	lines := make([]string, 0, len(coverage.rejections))
	for _, vector := range RejectionVectors() {
		paths := make([]string, 0, len(coverage.rejections[vector.ID]))
		for path := range coverage.rejections[vector.ID] {
			paths = append(paths, string(path))
		}
		sort.Strings(paths)
		lines = append(lines, vector.ID+"="+strings.Join(paths, "+"))
	}
	return lines
}
