package closuregraph

import (
	"fmt"
	"sort"
)

// CheckpointName is the closed C0-C7 checkpoint discriminator. Cargo's C3a
// and C3b records are explicit intermediate admission gates.
type CheckpointName string

const (
	// CheckpointC0 through CheckpointC7 identify the causal checkpoint chain;
	// CheckpointC3A and CheckpointC3B are Cargo's explicit intermediate gates.
	CheckpointC0 CheckpointName = "C0.profile"
	// CheckpointC1 identifies the frozen resolution gate.
	CheckpointC1 CheckpointName = "C1.resolve"
	// CheckpointC2 identifies the immutable capture gate.
	CheckpointC2 CheckpointName = "C2.capture"
	// CheckpointC3A identifies Cargo's origin-admission gate.
	CheckpointC3A CheckpointName = "C3a.origin-admission"
	// CheckpointC3B identifies Cargo's derived-vendor admission gate.
	CheckpointC3B CheckpointName = "C3b.derived-admission"
	// CheckpointC3 identifies the recursive artifact-admission gate.
	CheckpointC3 CheckpointName = "C3.admit"
	// CheckpointC4 identifies the closed active-graph gate.
	CheckpointC4 CheckpointName = "C4.close"
	// CheckpointC5 identifies the deterministic build-plan gate.
	CheckpointC5 CheckpointName = "C5.plan"
	// CheckpointC6 identifies the protected offline-execution gate.
	CheckpointC6 CheckpointName = "C6.offline"
	// CheckpointC7 identifies the protected publication gate.
	CheckpointC7 CheckpointName = "C7.publish"
)

// CheckpointDecision is the successful decision appropriate to a checkpoint.
// Failed checkpoints are not issued and cannot seed a downstream identity.
type CheckpointDecision string

const (
	// DecisionAdmit and the related constants enumerate successful checkpoint
	// outcomes.
	DecisionAdmit CheckpointDecision = "admit"
	// DecisionSuccess marks a successful protected operation.
	DecisionSuccess CheckpointDecision = "success"
	// DecisionPublished marks a successful protected publication.
	DecisionPublished CheckpointDecision = "published"
)

// Diagnostic is one deterministic checkpoint diagnostic. Fields are a closed
// string map chosen by the diagnostic's versioned contract.
type Diagnostic struct {
	Code    DiagnosticCode
	Subject string
	Fields  map[string]string
}

func (diagnostic Diagnostic) validate() error {
	if err := validatePortableText(string(diagnostic.Code), "diagnostic code", false); err != nil {
		return err
	}
	if err := validatePortableText(diagnostic.Subject, "diagnostic subject", false); err != nil {
		return err
	}
	if diagnostic.Fields == nil {
		return fmt.Errorf("diagnostic fields must be an explicit object")
	}
	return validatePortableStringMap(diagnostic.Fields, "diagnostic field", true)
}
func (diagnostic Diagnostic) value() map[string]any {
	return map[string]any{"code": string(diagnostic.Code), "fields": stringMapToAny(diagnostic.Fields), "subject": diagnostic.Subject}
}

// CheckpointPayload is a sealed schema for one C0-C7 payload.
type CheckpointPayload interface {
	checkpointPayload()
	checkpointName() CheckpointName
	validate() error
	value() map[string]any
}

// C0ProfilePayload binds policy, selection, platform, capabilities, and every
// external tool that may execute before C5.
type C0ProfilePayload struct {
	AdapterProfileID         string
	SchemaIDs                []string
	ArtifactPolicyID         string
	DetectorRegistryID       string
	SourceGrammarIDs         []string
	LimitVectorID            string
	SelectionContextID       ID
	PlatformNodeIDs          []ID
	PlatformRoles            map[PlatformRole]ID
	ManagerSchemaIDs         []string
	ConfigurationPolicyID    string
	CapabilityIDs            []string
	EvidenceToolchainNodeIDs []ID
}

// C1ResolvePayload binds frozen declarations, lock candidate, unevaluated
// conditions, evaluator identities, candidate records, and causal journal.
type C1ResolvePayload struct {
	RootDeclarationIDs      []ID
	WorkspaceDeclarationIDs []ID
	LockCandidateID         ID
	ConditionEdgeIDs        []ID
	ParserEvaluatorIDs      []string
	CandidateNodeIDs        []ID
	CandidateEdgeIDs        []ID
	SelectionContextID      ID
	JournalEntryIDs         []ID
}

// C2CapturePayload binds every immutable intake and protected capture handle.
type C2CapturePayload struct {
	IntakeReceiptIDs   []ID
	OriginIDs          []ID
	ProtectedHandleIDs []ID
	BrokerReceiptIDs   []ID
}

// C3AdmitPayload binds recursive artifact manifests and any permitted derived
// admission journal. Phase identifies main, origin, or derived admission.
type C3AdmitPayload struct {
	Phase                string
	IntakeReceiptIDs     []ID
	ArtifactManifestIDs  []ID
	DerivationReceiptIDs []ID
}

// C4ClosePayload binds the four exact graph identities.
type C4ClosePayload struct {
	ActiveGraphID      ID
	CapturedGraphID    ID
	SelectionBindingID ID
	SelectionContextID ID
}

// C5PlanPayload binds the immutable deterministic plan.
type C5PlanPayload struct{ BuildPlanID ID }

// C6OfflinePayload binds the separate execution receipt and observations.
type C6OfflinePayload struct{ ExecutionReceiptID ID }

// C7PublishPayload binds the protected publication receipt.
type C7PublishPayload struct{ PublicationReceiptID ID }

func (C0ProfilePayload) checkpointPayload() {}
func (C1ResolvePayload) checkpointPayload() {}
func (C2CapturePayload) checkpointPayload() {}
func (C3AdmitPayload) checkpointPayload()   {}
func (C4ClosePayload) checkpointPayload()   {}
func (C5PlanPayload) checkpointPayload()    {}
func (C6OfflinePayload) checkpointPayload() {}
func (C7PublishPayload) checkpointPayload() {}

func (C0ProfilePayload) checkpointName() CheckpointName { return CheckpointC0 }
func (C1ResolvePayload) checkpointName() CheckpointName { return CheckpointC1 }
func (C2CapturePayload) checkpointName() CheckpointName { return CheckpointC2 }
func (payload C3AdmitPayload) checkpointName() CheckpointName {
	switch payload.Phase {
	case "origin":
		return CheckpointC3A
	case "derived":
		return CheckpointC3B
	default:
		return CheckpointC3
	}
}
func (C4ClosePayload) checkpointName() CheckpointName   { return CheckpointC4 }
func (C5PlanPayload) checkpointName() CheckpointName    { return CheckpointC5 }
func (C6OfflinePayload) checkpointName() CheckpointName { return CheckpointC6 }
func (C7PublishPayload) checkpointName() CheckpointName { return CheckpointC7 }

// Checkpoint is one exact predecessor-linked C0-C7 record.
type Checkpoint struct {
	SchemaID             string
	Name                 CheckpointName
	PreviousCheckpointID *ID
	Payload              CheckpointPayload
	Decision             CheckpointDecision
	Diagnostics          []Diagnostic
}

// NewCheckpoint links a successful payload to its exact predecessor.
func NewCheckpoint(payload CheckpointPayload, previous *Checkpoint, diagnostics []Diagnostic) (Checkpoint, error) {
	if payload == nil {
		return Checkpoint{}, fmt.Errorf("checkpoint payload is required")
	}
	name, canonical := checkpointPayloadValueName(payload)
	if !canonical {
		return Checkpoint{}, fmt.Errorf("%s: checkpoint payload must use its canonical value representation, got %T", CodeCheckpointInvalid, payload)
	}
	var previousID *ID
	if previous != nil {
		id, err := previous.ID()
		if err != nil {
			return Checkpoint{}, fmt.Errorf("previous checkpoint: %w", err)
		}
		previousID = &id
	}
	checkpoint := Checkpoint{SchemaID: SchemaCheckpoint, Name: name, PreviousCheckpointID: previousID, Payload: payload, Decision: decisionForCheckpoint(name), Diagnostics: append([]Diagnostic{}, diagnostics...)}
	sortDiagnostics(checkpoint.Diagnostics)
	if err := checkpoint.Validate(); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}

// Validate checks checkpoint shape, payload kind, successful decision, and
// predecessor presence. ValidateCheckpointChain proves exact sequencing.
func (checkpoint Checkpoint) Validate() error {
	if checkpoint.SchemaID != SchemaCheckpoint {
		return fmt.Errorf("%s: unsupported checkpoint schema %q", CodeGraphSchemaUnsupported, checkpoint.SchemaID)
	}
	if !validCheckpointName(checkpoint.Name) {
		return fmt.Errorf("%s: unsupported checkpoint name %q", CodeGraphSchemaUnsupported, checkpoint.Name)
	}
	payloadName, canonical := checkpointPayloadValueName(checkpoint.Payload)
	if !canonical {
		return fmt.Errorf("%s: checkpoint %s payload must use its canonical value representation, got %T", CodeCheckpointInvalid, checkpoint.Name, checkpoint.Payload)
	}
	if checkpoint.Name == CheckpointC0 {
		if checkpoint.PreviousCheckpointID != nil {
			return fmt.Errorf("%s: C0 previous_checkpoint_id must be null", CodeCheckpointInvalid)
		}
	} else {
		if checkpoint.PreviousCheckpointID == nil {
			return fmt.Errorf("%s: %s requires previous_checkpoint_id", CodeCheckpointInvalid, checkpoint.Name)
		}
		if err := validateID(*checkpoint.PreviousCheckpointID, "previous_checkpoint_id"); err != nil {
			return err
		}
	}
	if payloadName != checkpoint.Name {
		return fmt.Errorf("%s: checkpoint payload kind does not match %s", CodeCheckpointInvalid, checkpoint.Name)
	}
	if err := checkpoint.Payload.validate(); err != nil {
		return err
	}
	if checkpoint.Decision != decisionForCheckpoint(checkpoint.Name) {
		return fmt.Errorf("%s: checkpoint %s decision must be %q", CodeCheckpointInvalid, checkpoint.Name, decisionForCheckpoint(checkpoint.Name))
	}
	if checkpoint.Diagnostics == nil {
		return fmt.Errorf("checkpoint diagnostics must be an explicit array")
	}
	for index, diagnostic := range checkpoint.Diagnostics {
		if err := diagnostic.validate(); err != nil {
			return fmt.Errorf("diagnostics[%d]: %w", index, err)
		}
		if index > 0 && diagnosticSortKey(checkpoint.Diagnostics[index-1]) >= diagnosticSortKey(diagnostic) {
			return fmt.Errorf("checkpoint diagnostics must be sorted and unique")
		}
	}
	return nil
}

// checkpointPayloadValueName recognizes only payload values emitted by the
// canonical checkpoint codec. Pointer forms otherwise enter the interface's
// method set through value receivers and typed-nil pointers can panic on the
// first method call.
func checkpointPayloadValueName(payload CheckpointPayload) (CheckpointName, bool) {
	switch value := payload.(type) {
	case C0ProfilePayload:
		return CheckpointC0, true
	case C1ResolvePayload:
		return CheckpointC1, true
	case C2CapturePayload:
		return CheckpointC2, true
	case C3AdmitPayload:
		return value.checkpointName(), true
	case C4ClosePayload:
		return CheckpointC4, true
	case C5PlanPayload:
		return CheckpointC5, true
	case C6OfflinePayload:
		return CheckpointC6, true
	case C7PublishPayload:
		return CheckpointC7, true
	default:
		return "", false
	}
}

// CanonicalBytes returns exact curator-checkpoint-v1 CCJ bytes.
func (checkpoint Checkpoint) CanonicalBytes() ([]byte, error) { return canonicalBytes(checkpoint) }

// ID derives checkpoint_id(Cn).
func (checkpoint Checkpoint) ID() (ID, error) { return recordID(checkpoint) }

func (checkpoint Checkpoint) domainLabel() string { return LabelCheckpoint }
func (checkpoint Checkpoint) canonicalValue() map[string]any {
	diagnostics := make([]any, len(checkpoint.Diagnostics))
	for i, diagnostic := range checkpoint.Diagnostics {
		diagnostics[i] = diagnostic.value()
	}
	var previous any
	if checkpoint.PreviousCheckpointID != nil {
		previous = string(*checkpoint.PreviousCheckpointID)
	}
	return map[string]any{"checkpoint_name": string(checkpoint.Name), "decision": string(checkpoint.Decision), "diagnostics": diagnostics, "payload": checkpoint.Payload.value(), "previous_checkpoint_id": previous, "schema_id": checkpoint.SchemaID}
}

// DecodeCheckpoint accepts exact canonical checkpoint bytes.
func DecodeCheckpoint(payload []byte) (Checkpoint, error) {
	raw, err := decodeCanonicalObject(payload, "checkpoint")
	if err != nil {
		return Checkpoint{}, err
	}
	if err := exactFields(raw, "checkpoint", []string{"checkpoint_name", "decision", "diagnostics", "payload", "previous_checkpoint_id", "schema_id"}, nil); err != nil {
		return Checkpoint{}, err
	}
	checkpoint := Checkpoint{}
	checkpoint.SchemaID, err = requiredString(raw, "schema_id", "checkpoint")
	if err != nil {
		return Checkpoint{}, err
	}
	name, err := requiredString(raw, "checkpoint_name", "checkpoint")
	if err != nil {
		return Checkpoint{}, err
	}
	checkpoint.Name = CheckpointName(name)
	decision, err := requiredString(raw, "decision", "checkpoint")
	if err != nil {
		return Checkpoint{}, err
	}
	checkpoint.Decision = CheckpointDecision(decision)
	if rawPrevious := raw["previous_checkpoint_id"]; rawPrevious != nil {
		text, ok := rawPrevious.(string)
		if !ok {
			return Checkpoint{}, fmt.Errorf("checkpoint previous_checkpoint_id must be null or string")
		}
		id := ID(text)
		checkpoint.PreviousCheckpointID = &id
	}
	payloadRaw, err := requiredObject(raw, "payload", "checkpoint")
	if err != nil {
		return Checkpoint{}, err
	}
	checkpoint.Payload, err = decodeCheckpointPayload(checkpoint.Name, payloadRaw)
	if err != nil {
		return Checkpoint{}, err
	}
	diagnosticsRaw, ok := raw["diagnostics"].([]any)
	if !ok {
		return Checkpoint{}, fmt.Errorf("checkpoint diagnostics must be an array")
	}
	checkpoint.Diagnostics = make([]Diagnostic, len(diagnosticsRaw))
	for i, item := range diagnosticsRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return Checkpoint{}, fmt.Errorf("diagnostics[%d] must be an object", i)
		}
		if err := exactFields(object, "diagnostic", []string{"code", "fields", "subject"}, nil); err != nil {
			return Checkpoint{}, err
		}
		code, err := requiredString(object, "code", "diagnostic")
		if err != nil {
			return Checkpoint{}, err
		}
		subject, err := requiredString(object, "subject", "diagnostic")
		if err != nil {
			return Checkpoint{}, err
		}
		fields, err := requiredStringMap(object, "fields", "diagnostic")
		if err != nil {
			return Checkpoint{}, err
		}
		checkpoint.Diagnostics[i] = Diagnostic{Code: DiagnosticCode(code), Subject: subject, Fields: fields}
	}
	if err := checkpoint.Validate(); err != nil {
		return Checkpoint{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, checkpoint); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}

// validateCheckpointSequence checks the exact predecessor-linked envelope
// sequence. ValidateCheckpointChain additionally requires and reconciles the
// records named by every payload.
func validateCheckpointSequence(checkpoints []Checkpoint) error {
	if len(checkpoints) != 8 && len(checkpoints) != 10 {
		return fmt.Errorf("%s: checkpoint chain must contain 8 records, or 10 with Cargo C3a/C3b", CodeCheckpointInvalid)
	}
	want := []CheckpointName{CheckpointC0, CheckpointC1, CheckpointC2}
	if len(checkpoints) == 10 {
		want = append(want, CheckpointC3A, CheckpointC3B)
	}
	want = append(want, CheckpointC3, CheckpointC4, CheckpointC5, CheckpointC6, CheckpointC7)
	for index, checkpoint := range checkpoints {
		if err := checkpoint.Validate(); err != nil {
			return fmt.Errorf("checkpoint[%d]: %w", index, err)
		}
		if checkpoint.Name != want[index] {
			return fmt.Errorf("%s: checkpoint[%d] is %s, want %s", CodeCheckpointInvalid, index, checkpoint.Name, want[index])
		}
		if index == 0 {
			continue
		}
		previousID, err := checkpoints[index-1].ID()
		if err != nil {
			return err
		}
		if checkpoint.PreviousCheckpointID == nil || *checkpoint.PreviousCheckpointID != previousID {
			return fmt.Errorf("%s: checkpoint[%d] predecessor mismatch", CodeCheckpointInvalid, index)
		}
	}
	c0 := checkpoints[0].Payload.(C0ProfilePayload)
	c1 := checkpoints[1].Payload.(C1ResolvePayload)
	if c1.SelectionContextID != c0.SelectionContextID {
		return fmt.Errorf("%s: C1 selection_context_id does not match C0", CodeCheckpointInvalid)
	}
	c4Index := 4
	if len(checkpoints) == 10 {
		c4Index = 6
	}
	c4 := checkpoints[c4Index].Payload.(C4ClosePayload)
	if c4.SelectionContextID != c0.SelectionContextID {
		return fmt.Errorf("%s: C4 selection_context_id does not match C0", CodeCheckpointInvalid)
	}
	return nil
}

func validCheckpointName(name CheckpointName) bool {
	switch name {
	case CheckpointC0, CheckpointC1, CheckpointC2, CheckpointC3A, CheckpointC3B, CheckpointC3, CheckpointC4, CheckpointC5, CheckpointC6, CheckpointC7:
		return true
	default:
		return false
	}
}
func decisionForCheckpoint(name CheckpointName) CheckpointDecision {
	if name == CheckpointC6 {
		return DecisionSuccess
	}
	if name == CheckpointC7 {
		return DecisionPublished
	}
	return DecisionAdmit
}
func diagnosticSortKey(value Diagnostic) string { return string(value.Code) + "\x00" + value.Subject }
func sortDiagnostics(values []Diagnostic) {
	sort.Slice(values, func(i, j int) bool { return diagnosticSortKey(values[i]) < diagnosticSortKey(values[j]) })
}

func (payload C0ProfilePayload) validate() error {
	if err := validatePortableTextFields(map[string]string{"adapter_profile_id": payload.AdapterProfileID, "artifact_policy_id": payload.ArtifactPolicyID, "detector_registry_id": payload.DetectorRegistryID, "limit_vector_id": payload.LimitVectorID, "configuration_policy_id": payload.ConfigurationPolicyID}, false, false); err != nil {
		return err
	}
	if err := validateStringSliceFields(map[string][]string{"schema_ids": payload.SchemaIDs, "source_grammar_ids": payload.SourceGrammarIDs, "manager_schema_ids": payload.ManagerSchemaIDs, "capability_ids": payload.CapabilityIDs}, true); err != nil {
		return err
	}
	if err := validateID(payload.SelectionContextID, "selection_context_id"); err != nil {
		return err
	}
	if err := validateIDSlice(payload.PlatformNodeIDs, "platform_node_ids", true); err != nil {
		return err
	}
	if len(payload.PlatformNodeIDs) == 0 {
		return fmt.Errorf("platform_node_ids must not be empty")
	}
	if payload.PlatformRoles == nil {
		return fmt.Errorf("platform_roles must be an explicit object")
	}
	if _, present := payload.PlatformRoles[PlatformTarget]; !present {
		return fmt.Errorf("platform_roles requires target")
	}
	platforms := idSet(payload.PlatformNodeIDs)
	for _, role := range sortedPlatformRoles(payload.PlatformRoles) {
		id := payload.PlatformRoles[role]
		if role != PlatformTarget && role != PlatformHost {
			return fmt.Errorf("unsupported platform role %q", role)
		}
		if err := validateID(id, "platform role"); err != nil {
			return err
		}
		if !platforms[id] {
			return fmt.Errorf("platform role %q references a node absent from platform_node_ids", role)
		}
	}
	return validateIDSlice(payload.EvidenceToolchainNodeIDs, "evidence_toolchain_node_ids", true)
}
func (payload C0ProfilePayload) value() map[string]any {
	roles := map[string]any{}
	for role, id := range payload.PlatformRoles {
		roles[string(role)] = string(id)
	}
	return map[string]any{"adapter_profile_id": payload.AdapterProfileID, "artifact_policy_id": payload.ArtifactPolicyID, "capability_ids": stringsToAny(payload.CapabilityIDs), "configuration_policy_id": payload.ConfigurationPolicyID, "detector_registry_id": payload.DetectorRegistryID, "evidence_toolchain_node_ids": idsToAny(payload.EvidenceToolchainNodeIDs), "limit_vector_id": payload.LimitVectorID, "manager_schema_ids": stringsToAny(payload.ManagerSchemaIDs), "platform_node_ids": idsToAny(payload.PlatformNodeIDs), "platform_roles": roles, "schema_ids": stringsToAny(payload.SchemaIDs), "selection_context_id": string(payload.SelectionContextID), "source_grammar_ids": stringsToAny(payload.SourceGrammarIDs)}
}

func (payload C1ResolvePayload) validate() error {
	if err := validateIDSliceFields(map[string][]ID{"root_declaration_ids": payload.RootDeclarationIDs, "workspace_declaration_ids": payload.WorkspaceDeclarationIDs, "condition_edge_ids": payload.ConditionEdgeIDs, "candidate_node_ids": payload.CandidateNodeIDs, "candidate_edge_ids": payload.CandidateEdgeIDs, "journal_entry_ids": payload.JournalEntryIDs}, true); err != nil {
		return err
	}
	if err := validateID(payload.LockCandidateID, "lock_candidate_id"); err != nil {
		return err
	}
	if err := validateStringSlice(payload.ParserEvaluatorIDs, "parser_evaluator_ids", true); err != nil {
		return err
	}
	return validateID(payload.SelectionContextID, "selection_context_id")
}
func (payload C1ResolvePayload) value() map[string]any {
	return map[string]any{"candidate_edge_ids": idsToAny(payload.CandidateEdgeIDs), "candidate_node_ids": idsToAny(payload.CandidateNodeIDs), "condition_edge_ids": idsToAny(payload.ConditionEdgeIDs), "journal_entry_ids": idsToAny(payload.JournalEntryIDs), "lock_candidate_id": string(payload.LockCandidateID), "parser_evaluator_ids": stringsToAny(payload.ParserEvaluatorIDs), "root_declaration_ids": idsToAny(payload.RootDeclarationIDs), "selection_context_id": string(payload.SelectionContextID), "workspace_declaration_ids": idsToAny(payload.WorkspaceDeclarationIDs)}
}

func (payload C2CapturePayload) validate() error {
	return validateIDSliceFields(map[string][]ID{"intake_receipt_ids": payload.IntakeReceiptIDs, "origin_ids": payload.OriginIDs, "protected_handle_ids": payload.ProtectedHandleIDs, "broker_receipt_ids": payload.BrokerReceiptIDs}, true)
}
func (payload C2CapturePayload) value() map[string]any {
	return map[string]any{"broker_receipt_ids": idsToAny(payload.BrokerReceiptIDs), "intake_receipt_ids": idsToAny(payload.IntakeReceiptIDs), "origin_ids": idsToAny(payload.OriginIDs), "protected_handle_ids": idsToAny(payload.ProtectedHandleIDs)}
}

func (payload C3AdmitPayload) validate() error {
	if payload.Phase != "main" && payload.Phase != "origin" && payload.Phase != "derived" {
		return fmt.Errorf("unsupported C3 phase %q", payload.Phase)
	}
	return validateIDSliceFields(map[string][]ID{"intake_receipt_ids": payload.IntakeReceiptIDs, "artifact_manifest_ids": payload.ArtifactManifestIDs, "derivation_receipt_ids": payload.DerivationReceiptIDs}, true)
}
func (payload C3AdmitPayload) value() map[string]any {
	return map[string]any{"artifact_manifest_ids": idsToAny(payload.ArtifactManifestIDs), "derivation_receipt_ids": idsToAny(payload.DerivationReceiptIDs), "intake_receipt_ids": idsToAny(payload.IntakeReceiptIDs), "phase": payload.Phase}
}

func (payload C4ClosePayload) validate() error {
	return validateIDFields(map[string]ID{"active_graph_id": payload.ActiveGraphID, "captured_graph_id": payload.CapturedGraphID, "selection_binding_id": payload.SelectionBindingID, "selection_context_id": payload.SelectionContextID})
}
func (payload C4ClosePayload) value() map[string]any {
	return map[string]any{"active_graph_id": string(payload.ActiveGraphID), "captured_graph_id": string(payload.CapturedGraphID), "selection_binding_id": string(payload.SelectionBindingID), "selection_context_id": string(payload.SelectionContextID)}
}
func (payload C5PlanPayload) validate() error {
	return validateID(payload.BuildPlanID, "build_plan_id")
}
func (payload C5PlanPayload) value() map[string]any {
	return map[string]any{"build_plan_id": string(payload.BuildPlanID)}
}
func (payload C6OfflinePayload) validate() error {
	return validateID(payload.ExecutionReceiptID, "execution_receipt_id")
}
func (payload C6OfflinePayload) value() map[string]any {
	return map[string]any{"execution_receipt_id": string(payload.ExecutionReceiptID)}
}
func (payload C7PublishPayload) validate() error {
	return validateID(payload.PublicationReceiptID, "publication_receipt_id")
}
func (payload C7PublishPayload) value() map[string]any {
	return map[string]any{"publication_receipt_id": string(payload.PublicationReceiptID)}
}

func decodeCheckpointPayload(name CheckpointName, raw map[string]any) (CheckpointPayload, error) {
	switch name {
	case CheckpointC0:
		fields := []string{"adapter_profile_id", "artifact_policy_id", "capability_ids", "configuration_policy_id", "detector_registry_id", "evidence_toolchain_node_ids", "limit_vector_id", "manager_schema_ids", "platform_node_ids", "platform_roles", "schema_ids", "selection_context_id", "source_grammar_ids"}
		if err := exactFields(raw, "C0 payload", fields, nil); err != nil {
			return nil, err
		}
		stringFields, err := decodeStringFields(raw, "C0 payload", []string{"adapter_profile_id", "artifact_policy_id", "detector_registry_id", "limit_vector_id", "configuration_policy_id", "selection_context_id"}, nil)
		if err != nil {
			return nil, err
		}
		p := C0ProfilePayload{
			AdapterProfileID:      stringFields["adapter_profile_id"],
			ArtifactPolicyID:      stringFields["artifact_policy_id"],
			DetectorRegistryID:    stringFields["detector_registry_id"],
			LimitVectorID:         stringFields["limit_vector_id"],
			ConfigurationPolicyID: stringFields["configuration_policy_id"],
			SelectionContextID:    ID(stringFields["selection_context_id"]),
		}
		p.SchemaIDs, err = requiredStringSlice(raw, "schema_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.SourceGrammarIDs, err = requiredStringSlice(raw, "source_grammar_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.ManagerSchemaIDs, err = requiredStringSlice(raw, "manager_schema_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.CapabilityIDs, err = requiredStringSlice(raw, "capability_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.PlatformNodeIDs, err = requiredIDSlice(raw, "platform_node_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.EvidenceToolchainNodeIDs, err = requiredIDSlice(raw, "evidence_toolchain_node_ids", "C0 payload")
		if err != nil {
			return nil, err
		}
		rolesRaw, err := requiredObject(raw, "platform_roles", "C0 payload")
		if err != nil {
			return nil, err
		}
		p.PlatformRoles = map[PlatformRole]ID{}
		for _, role := range sortedMapKeys(rolesRaw) {
			value := rolesRaw[role]
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("C0 platform role %s must be a string", role)
			}
			p.PlatformRoles[PlatformRole(role)] = ID(text)
		}
		return p, nil
	case CheckpointC1:
		fields := []string{"candidate_edge_ids", "candidate_node_ids", "condition_edge_ids", "journal_entry_ids", "lock_candidate_id", "parser_evaluator_ids", "root_declaration_ids", "selection_context_id", "workspace_declaration_ids"}
		if err := exactFields(raw, "C1 payload", fields, nil); err != nil {
			return nil, err
		}
		p := C1ResolvePayload{}
		var err error
		p.CandidateEdgeIDs, err = requiredIDSlice(raw, "candidate_edge_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.CandidateNodeIDs, err = requiredIDSlice(raw, "candidate_node_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.ConditionEdgeIDs, err = requiredIDSlice(raw, "condition_edge_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.JournalEntryIDs, err = requiredIDSlice(raw, "journal_entry_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		lock, err := requiredString(raw, "lock_candidate_id", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.LockCandidateID = ID(lock)
		p.ParserEvaluatorIDs, err = requiredStringSlice(raw, "parser_evaluator_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.RootDeclarationIDs, err = requiredIDSlice(raw, "root_declaration_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		selection, err := requiredString(raw, "selection_context_id", "C1 payload")
		if err != nil {
			return nil, err
		}
		p.SelectionContextID = ID(selection)
		p.WorkspaceDeclarationIDs, err = requiredIDSlice(raw, "workspace_declaration_ids", "C1 payload")
		if err != nil {
			return nil, err
		}
		return p, nil
	case CheckpointC2:
		fields := []string{"broker_receipt_ids", "intake_receipt_ids", "origin_ids", "protected_handle_ids"}
		if err := exactFields(raw, "C2 payload", fields, nil); err != nil {
			return nil, err
		}
		p := C2CapturePayload{}
		var err error
		p.BrokerReceiptIDs, err = requiredIDSlice(raw, "broker_receipt_ids", "C2 payload")
		if err != nil {
			return nil, err
		}
		p.IntakeReceiptIDs, err = requiredIDSlice(raw, "intake_receipt_ids", "C2 payload")
		if err != nil {
			return nil, err
		}
		p.OriginIDs, err = requiredIDSlice(raw, "origin_ids", "C2 payload")
		if err != nil {
			return nil, err
		}
		p.ProtectedHandleIDs, err = requiredIDSlice(raw, "protected_handle_ids", "C2 payload")
		if err != nil {
			return nil, err
		}
		return p, nil
	case CheckpointC3A, CheckpointC3B, CheckpointC3:
		fields := []string{"artifact_manifest_ids", "derivation_receipt_ids", "intake_receipt_ids", "phase"}
		if err := exactFields(raw, "C3 payload", fields, nil); err != nil {
			return nil, err
		}
		phase, err := requiredString(raw, "phase", "C3 payload")
		if err != nil {
			return nil, err
		}
		p := C3AdmitPayload{Phase: phase}
		p.ArtifactManifestIDs, err = requiredIDSlice(raw, "artifact_manifest_ids", "C3 payload")
		if err != nil {
			return nil, err
		}
		p.DerivationReceiptIDs, err = requiredIDSlice(raw, "derivation_receipt_ids", "C3 payload")
		if err != nil {
			return nil, err
		}
		p.IntakeReceiptIDs, err = requiredIDSlice(raw, "intake_receipt_ids", "C3 payload")
		if err != nil {
			return nil, err
		}
		return p, nil
	case CheckpointC4:
		if err := exactFields(raw, "C4 payload", []string{"active_graph_id", "captured_graph_id", "selection_binding_id", "selection_context_id"}, nil); err != nil {
			return nil, err
		}
		stringFields, err := decodeStringFields(raw, "C4 payload", []string{"active_graph_id", "captured_graph_id", "selection_binding_id", "selection_context_id"}, nil)
		if err != nil {
			return nil, err
		}
		return C4ClosePayload{ActiveGraphID: ID(stringFields["active_graph_id"]), CapturedGraphID: ID(stringFields["captured_graph_id"]), SelectionBindingID: ID(stringFields["selection_binding_id"]), SelectionContextID: ID(stringFields["selection_context_id"])}, nil
	case CheckpointC5:
		if err := exactFields(raw, "C5 payload", []string{"build_plan_id"}, nil); err != nil {
			return nil, err
		}
		id, err := requiredString(raw, "build_plan_id", "C5 payload")
		if err != nil {
			return nil, err
		}
		return C5PlanPayload{BuildPlanID: ID(id)}, nil
	case CheckpointC6:
		if err := exactFields(raw, "C6 payload", []string{"execution_receipt_id"}, nil); err != nil {
			return nil, err
		}
		id, err := requiredString(raw, "execution_receipt_id", "C6 payload")
		if err != nil {
			return nil, err
		}
		return C6OfflinePayload{ExecutionReceiptID: ID(id)}, nil
	case CheckpointC7:
		if err := exactFields(raw, "C7 payload", []string{"publication_receipt_id"}, nil); err != nil {
			return nil, err
		}
		id, err := requiredString(raw, "publication_receipt_id", "C7 payload")
		if err != nil {
			return nil, err
		}
		return C7PublishPayload{PublicationReceiptID: ID(id)}, nil
	default:
		return nil, fmt.Errorf("%s: unsupported checkpoint name %q", CodeGraphSchemaUnsupported, name)
	}
}
