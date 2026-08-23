package closuregraph

import (
	"fmt"
	"sort"
)

// CheckpointChainEvidence is the complete record authority named by C0-C7.
// The chain validator requires these values rather than accepting syntactically
// valid but stale or unrelated checkpoint references.
type CheckpointChainEvidence struct {
	Records            CheckpointRecordEvidence
	Graph              GraphBundle
	ExecutionPolicyID  string
	Plan               BuildPlan
	Closure            SourceClosure
	ExpectedCacheInput ExpectedCacheInput
	Observations       []ProducedArtifactObservation
	Execution          ExecutionReceipt
	Publication        PublicationReceipt
}

// CheckpointRecordEvidence is the independently supplied C1-C3 aggregate
// record table. The graph package does not own the underlying lock, intake, or
// broker record schemas, but it does require their exact content identities to
// remain unchanged across checkpoint aggregation.
type CheckpointRecordEvidence struct {
	RootDeclarationIDs      []ID
	WorkspaceDeclarationIDs []ID
	LockCandidateID         ID
	JournalEntryIDs         []ID
	IntakeReceiptIDs        []ID
	OriginIDs               []ID
	ProtectedHandleIDs      []ID
	BrokerReceiptIDs        []ID
	ArtifactManifestIDs     []ID
	DerivationReceiptIDs    []ID
}

// Validate checks canonical ordering and identity shape for the independently
// supplied aggregate table.
func (records CheckpointRecordEvidence) Validate() error {
	if err := validateID(records.LockCandidateID, "checkpoint records lock_candidate_id"); err != nil {
		return err
	}
	fields := []struct {
		name   string
		values []ID
	}{
		{name: "root_declaration_ids", values: records.RootDeclarationIDs},
		{name: "workspace_declaration_ids", values: records.WorkspaceDeclarationIDs},
		{name: "journal_entry_ids", values: records.JournalEntryIDs},
		{name: "intake_receipt_ids", values: records.IntakeReceiptIDs},
		{name: "origin_ids", values: records.OriginIDs},
		{name: "protected_handle_ids", values: records.ProtectedHandleIDs},
		{name: "broker_receipt_ids", values: records.BrokerReceiptIDs},
		{name: "artifact_manifest_ids", values: records.ArtifactManifestIDs},
		{name: "derivation_receipt_ids", values: records.DerivationReceiptIDs},
	}
	for _, field := range fields {
		if err := validateIDSlice(field.values, "checkpoint records "+field.name, true); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCheckpointChain requires one exact C0-C7 chain and resolves every
// downstream graph, plan, execution, observation, and publication reference.
// Cargo's C3a/C3b records may appear between C2 and the final C3 aggregate.
func ValidateCheckpointChain(checkpoints []Checkpoint, evidence CheckpointChainEvidence) error {
	if err := validateCheckpointSequence(checkpoints); err != nil {
		return err
	}
	return validateCheckpointEvidence(checkpoints, evidence)
}

func validateCheckpointEvidence(checkpoints []Checkpoint, evidence CheckpointChainEvidence) error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%s: %s", CodeCheckpointInvalid, fmt.Sprintf(format, args...))
	}
	if err := evidence.Records.Validate(); err != nil {
		return fail("C1-C3 record evidence is invalid: %v", err)
	}
	if err := evidence.Graph.Validate(); err != nil {
		return fail("C4 graph evidence is invalid: %v", err)
	}
	if err := validatePortableText(evidence.ExecutionPolicyID, "checkpoint evidence execution_policy_id", false); err != nil {
		return fail("C5 execution policy evidence is invalid: %v", err)
	}
	if err := evidence.Plan.Validate(); err != nil {
		return fail("C5 plan evidence is invalid: %v", err)
	}
	if evidence.Plan.ExecutionPolicyID != evidence.ExecutionPolicyID {
		return fail("C5 plan execution_policy_id does not match the independently supplied execution policy")
	}
	derivedPlan, err := DeriveBuildPlan(evidence.Graph, PlanOptions{ExecutionPolicyID: evidence.ExecutionPolicyID})
	if err != nil {
		return fail("C5 plan cannot be derived from the supplied C4 active graph: %v", err)
	}
	derivedPlanID, _ := derivedPlan.ID()
	suppliedPlanID, _ := evidence.Plan.ID()
	if suppliedPlanID != derivedPlanID {
		return fail("C5 plan does not match the exact action, output, ordering, and wave projection derived from the supplied C4 active graph")
	}
	if err := evidence.Closure.Validate(); err != nil {
		return fail("closure evidence is invalid: %v", err)
	}
	if err := evidence.ExpectedCacheInput.Validate(); err != nil {
		return fail("expected cache input evidence is invalid: %v", err)
	}
	if err := evidence.Execution.Validate(); err != nil {
		return fail("C6 execution evidence is invalid: %v", err)
	}
	if err := evidence.Publication.Validate(); err != nil {
		return fail("C7 publication evidence is invalid: %v", err)
	}

	c0 := checkpoints[0].Payload.(C0ProfilePayload)
	selectionID, _ := evidence.Graph.Selection.ID()
	if c0.SelectionContextID != selectionID {
		return fail("C0 selection_context_id does not resolve to the supplied selection")
	}
	platformIDs := []ID{}
	for _, node := range evidence.Graph.Records.BindingNodes {
		if node.Kind == NodeTargetPlatform {
			id, _ := node.ID()
			platformIDs = append(platformIDs, id)
		}
	}
	if !sameIDSlices(c0.PlatformNodeIDs, sortedIDs(platformIDs)) {
		return fail("C0 platform_node_ids do not resolve exactly to binding target_platform records")
	}
	if !samePlatformRoleMaps(c0.PlatformRoles, evidence.Graph.Selection.PlatformRoles) {
		return fail("C0 platform_roles do not match the supplied selection roles")
	}
	c0ToolIDs := []ID{}
	for _, authority := range evidence.Graph.Authority.Toolchains {
		if authority.FirstBound == ToolchainBoundAtC0 {
			c0ToolIDs = append(c0ToolIDs, authority.NodeID)
		}
	}
	if !sameIDSlices(c0.EvidenceToolchainNodeIDs, sortedIDs(c0ToolIDs)) {
		return fail("C0 evidence_toolchain_node_ids do not match C0 binding authority")
	}
	if len(c0ToolIDs) > 0 {
		if evidence.Graph.Authority.C0Checkpoint == nil {
			return fail("C0 evidence toolchains have no exact C0 checkpoint authority")
		}
		suppliedC0ID, _ := evidence.Graph.Authority.C0Checkpoint.ID()
		chainC0ID, _ := checkpoints[0].ID()
		if suppliedC0ID != chainC0ID {
			return fail("graph C0 authority does not name this checkpoint chain")
		}
	}

	c1 := checkpoints[1].Payload.(C1ResolvePayload)
	if !sameIDSlices(c1.RootDeclarationIDs, evidence.Records.RootDeclarationIDs) || !sameIDSlices(c1.WorkspaceDeclarationIDs, evidence.Records.WorkspaceDeclarationIDs) {
		return fail("C1 root/workspace declarations do not match supplied resolution records")
	}
	if c1.LockCandidateID != evidence.Records.LockCandidateID || !sameIDSlices(c1.JournalEntryIDs, evidence.Records.JournalEntryIDs) {
		return fail("C1 lock or journal references do not match supplied resolution records")
	}
	if !sameIDSlices(c1.CandidateNodeIDs, evidence.Graph.Capture.NodeIDs) || !sameIDSlices(c1.CandidateEdgeIDs, evidence.Graph.Capture.EdgeIDs) {
		return fail("C1 candidate graph does not match the supplied selection-neutral capture")
	}
	conditionIDs := []ID{}
	for _, edge := range evidence.Graph.Records.CaptureEdges {
		if edge.Payload.condition() != nil {
			id, _ := edge.ID()
			conditionIDs = append(conditionIDs, id)
			if !containsString(c1.ParserEvaluatorIDs, edge.Payload.condition().EvaluatorID) {
				return fail("C1 parser_evaluator_ids omit condition evaluator %q", edge.Payload.condition().EvaluatorID)
			}
		}
	}
	if !sameIDSlices(c1.ConditionEdgeIDs, sortedIDs(conditionIDs)) {
		return fail("C1 condition_edge_ids do not match the supplied capture conditions")
	}

	c2 := checkpoints[2].Payload.(C2CapturePayload)
	if !sameIDSlices(c2.IntakeReceiptIDs, evidence.Records.IntakeReceiptIDs) ||
		!sameIDSlices(c2.OriginIDs, evidence.Records.OriginIDs) ||
		!sameIDSlices(c2.ProtectedHandleIDs, evidence.Records.ProtectedHandleIDs) ||
		!sameIDSlices(c2.BrokerReceiptIDs, evidence.Records.BrokerReceiptIDs) {
		return fail("C2 aggregate does not match supplied intake/origin/handle/broker records")
	}
	c4Index := 4
	if len(checkpoints) == 10 {
		c4Index = 6
	}
	finalC3 := checkpoints[c4Index-1].Payload.(C3AdmitPayload)
	for index := 3; index < c4Index; index++ {
		c3 := checkpoints[index].Payload.(C3AdmitPayload)
		if !sameIDSlices(c3.IntakeReceiptIDs, c2.IntakeReceiptIDs) {
			return fail("%s intake_receipt_ids do not match C2", checkpoints[index].Name)
		}
		if index != c4Index-1 {
			for _, receiptID := range c3.DerivationReceiptIDs {
				if !containsID(finalC3.DerivationReceiptIDs, receiptID) {
					return fail("final C3 omits derivation receipt %s from %s", receiptID, checkpoints[index].Name)
				}
			}
		}
	}
	if !sameIDSlices(finalC3.ArtifactManifestIDs, evidence.Graph.Capture.ArtifactManifestIDs) {
		return fail("final C3 artifact_manifest_ids do not match the supplied capture")
	}
	if !sameIDSlices(finalC3.ArtifactManifestIDs, evidence.Records.ArtifactManifestIDs) || !sameIDSlices(finalC3.DerivationReceiptIDs, evidence.Records.DerivationReceiptIDs) {
		return fail("final C3 aggregate does not match supplied admission and derivation records")
	}

	c4 := checkpoints[c4Index].Payload.(C4ClosePayload)
	captureID, _ := evidence.Graph.Capture.ID()
	bindingID, _ := evidence.Graph.Binding.ID()
	activeID, _ := evidence.Graph.Active.ID()
	if c4.CapturedGraphID != captureID || c4.SelectionContextID != selectionID || c4.SelectionBindingID != bindingID || c4.ActiveGraphID != activeID {
		return fail("C4 graph references do not resolve to the supplied capture/selection/binding/active records")
	}

	c5Index := c4Index + 1
	c5 := checkpoints[c5Index].Payload.(C5PlanPayload)
	planID, _ := evidence.Plan.ID()
	if c5.BuildPlanID != planID || evidence.Plan.ActiveGraphID != activeID {
		return fail("C5 plan reference does not resolve to the supplied C4 active graph and plan")
	}
	c5CheckpointID, _ := checkpoints[c5Index].ID()
	if evidence.Closure.C5CheckpointID != c5CheckpointID {
		return fail("closure does not derive from this C5 checkpoint")
	}
	closureID, _ := evidence.Closure.ID()
	if evidence.ExpectedCacheInput.ClosureID != closureID || !sameIDSlices(evidence.ExpectedCacheInput.ExpectedOutputNodeIDs, evidence.Plan.DeclaredOutputNodeIDs) {
		return fail("expected cache input does not derive from this closure and plan outputs")
	}

	observationIDs := make([]ID, len(evidence.Observations))
	observationOutputIDs := make([]ID, len(evidence.Observations))
	observationPaths := make([]string, len(evidence.Observations))
	planActions := idSet(evidence.Plan.ActionNodeIDs)
	planOutputs := idSet(evidence.Plan.DeclaredOutputNodeIDs)
	for index, observation := range evidence.Observations {
		if err := observation.ValidateAgainst(evidence.Graph.Records); err != nil {
			return fail("observation[%d] does not resolve to the supplied graph: %v", index, err)
		}
		if !planActions[observation.ProducerActionID] || !planOutputs[observation.ExpectedOutputNodeID] {
			return fail("observation[%d] names an action or output absent from the supplied plan", index)
		}
		observationIDs[index], _ = observation.ID()
		observationOutputIDs[index] = observation.ExpectedOutputNodeID
		observationPaths[index] = observation.Path
	}
	if !strictlySortedUniqueIDs(observationIDs) {
		return fail("produced observations must be in canonical ID order")
	}
	if !sameIDSlices(sortedIDs(observationOutputIDs), evidence.Plan.DeclaredOutputNodeIDs) {
		return fail("C6 observations do not cover every declared output exactly once")
	}

	c6 := checkpoints[c5Index+1].Payload.(C6OfflinePayload)
	executionID, _ := evidence.Execution.ID()
	if c6.ExecutionReceiptID != executionID || evidence.Execution.ClosureID != closureID {
		return fail("C6 execution reference does not resolve to this closure and receipt")
	}
	if !sameIDSlices(evidence.Execution.ProducedObservationIDs, observationIDs) {
		return fail("C6 execution observation IDs do not match supplied observations")
	}
	if err := validateExecutionOrder(evidence.Execution.ActionOrder, evidence.Plan); err != nil {
		return fail("C6 action order does not execute the supplied plan: %v", err)
	}
	expectedWrites, err := selectedWritePaths(evidence.Graph, evidence.Plan.ActionNodeIDs)
	if err != nil {
		return fail("derive C6 write set: %v", err)
	}
	if !sameStringSlices(evidence.Execution.WriteSet, expectedWrites) {
		return fail("C6 write_set does not match selected produces edges")
	}
	for _, path := range observationPaths {
		if !containsString(evidence.Execution.WriteSet, path) {
			return fail("C6 observation path %q is absent from the execution write_set", path)
		}
	}

	c7 := checkpoints[c5Index+2].Payload.(C7PublishPayload)
	publicationID, _ := evidence.Publication.ID()
	expectedCacheID, _ := evidence.ExpectedCacheInput.ID()
	if c7.PublicationReceiptID != publicationID || evidence.Publication.ExecutionReceiptID != executionID || evidence.Publication.ExpectedCacheInputID != expectedCacheID {
		return fail("C7 publication reference does not resolve to this execution and expected cache input")
	}
	if !sameIDSlices(evidence.Publication.PublishedObservationIDs, observationIDs) {
		return fail("C7 published observations do not match the C6 output set")
	}
	return nil
}

func validateExecutionOrder(order []ID, plan BuildPlan) error {
	if len(order) != len(plan.ActionNodeIDs) {
		return fmt.Errorf("action count is %d, want %d", len(order), len(plan.ActionNodeIDs))
	}
	positions := map[ID]int{}
	for index, actionID := range order {
		if !containsID(plan.ActionNodeIDs, actionID) {
			return fmt.Errorf("action %s is absent from plan", actionID)
		}
		if _, exists := positions[actionID]; exists {
			return fmt.Errorf("action %s appears more than once", actionID)
		}
		positions[actionID] = index
	}
	for _, edge := range plan.OrderingEdges {
		if positions[edge.FromActionID] >= positions[edge.ToActionID] {
			return fmt.Errorf("action %s does not precede %s", edge.FromActionID, edge.ToActionID)
		}
	}
	return nil
}

func selectedWritePaths(graph GraphBundle, actionIDs []ID) ([]string, error) {
	resolved, err := validateStructure(graph.Capture, graph.Selection, graph.Binding, graph.Records, graph.Authority)
	if err != nil {
		return nil, err
	}
	return selectedWritePathsFromEdges(selectedEdges(graph.Active, resolved), idSet(actionIDs))
}

func sameIDSlices(left, right []ID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func samePlatformRoleMaps(left, right map[PlatformRole]ID) bool {
	if len(left) != len(right) {
		return false
	}
	for role, id := range left {
		if right[role] != id {
			return false
		}
	}
	return true
}

func containsID(values []ID, value ID) bool {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= value })
	return index < len(values) && values[index] == value
}

func strictlySortedUniqueIDs(values []ID) bool {
	for index, value := range values {
		if !value.Valid() || (index > 0 && values[index-1] >= value) {
			return false
		}
	}
	return true
}
