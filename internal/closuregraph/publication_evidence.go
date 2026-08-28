package closuregraph

import "fmt"

// PublicationEvidence is the immutable C4/C5 authority required before C6
// observations may become visible in protected storage.
type PublicationEvidence struct {
	C4      Checkpoint
	C5      Checkpoint
	Graph   GraphBundle
	Plan    BuildPlan
	Closure SourceClosure
}

// ValidateForPublication resolves the exact graph, plan, closure, action,
// output, produces-edge, path, class, target, and tool evidence for a proposed
// publication. It deliberately stops before C7 because the publication receipt
// is the result produced by the caller after this gate succeeds.
func (evidence PublicationEvidence) ValidateForPublication(expected ExpectedCacheInput, execution ExecutionReceipt, observations []ProducedArtifactObservation) error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%s: %s", CodeCheckpointInvalid, fmt.Sprintf(format, args...))
	}
	if err := evidence.Graph.Validate(); err != nil {
		return fail("C4 graph evidence is invalid: %v", err)
	}
	if err := evidence.Plan.Validate(); err != nil {
		return fail("C5 plan evidence is invalid: %v", err)
	}
	derived, err := DeriveBuildPlan(evidence.Graph, PlanOptions{ExecutionPolicyID: evidence.Plan.ExecutionPolicyID})
	if err != nil {
		return fail("C5 plan cannot be derived from C4: %v", err)
	}
	derivedID, _ := derived.ID()
	planID, _ := evidence.Plan.ID()
	if derivedID != planID {
		return fail("C5 plan differs from the exact C4 projection")
	}
	if err := evidence.C4.Validate(); err != nil || evidence.C4.Name != CheckpointC4 {
		return fail("C4 checkpoint is invalid: %v", err)
	}
	if err := evidence.C5.Validate(); err != nil || evidence.C5.Name != CheckpointC5 {
		return fail("C5 checkpoint is invalid: %v", err)
	}
	captureID, _ := evidence.Graph.Capture.ID()
	selectionID, _ := evidence.Graph.Selection.ID()
	bindingID, _ := evidence.Graph.Binding.ID()
	activeID, _ := evidence.Graph.Active.ID()
	c4 := evidence.C4.Payload.(C4ClosePayload)
	if c4.CapturedGraphID != captureID || c4.SelectionContextID != selectionID || c4.SelectionBindingID != bindingID || c4.ActiveGraphID != activeID {
		return fail("C4 checkpoint does not name the supplied graph records")
	}
	c4ID, _ := evidence.C4.ID()
	if evidence.C5.PreviousCheckpointID == nil || *evidence.C5.PreviousCheckpointID != c4ID || evidence.C5.Payload.(C5PlanPayload).BuildPlanID != planID || evidence.Plan.ActiveGraphID != activeID {
		return fail("C5 checkpoint does not name the supplied C4 and build plan")
	}
	if err := evidence.Closure.Validate(); err != nil {
		return fail("closure evidence is invalid: %v", err)
	}
	c5ID, _ := evidence.C5.ID()
	if evidence.Closure.C5CheckpointID != c5ID {
		return fail("closure does not derive from the supplied C5 checkpoint")
	}
	if err := expected.Validate(); err != nil {
		return fail("expected cache input is invalid: %v", err)
	}
	closureID, _ := evidence.Closure.ID()
	if expected.ClosureID != closureID || !sameIDSlices(expected.ExpectedOutputNodeIDs, evidence.Plan.DeclaredOutputNodeIDs) {
		return fail("expected cache input differs from the closure and C5 output set")
	}
	if err := execution.Validate(); err != nil {
		return fail("execution receipt is invalid: %v", err)
	}
	if execution.ClosureID != closureID {
		return fail("execution receipt names another closure")
	}
	if err := validateExecutionOrder(execution.ActionOrder, evidence.Plan); err != nil {
		return fail("execution action order differs from C5: %v", err)
	}
	expectedWrites, err := selectedWritePaths(evidence.Graph, evidence.Plan.ActionNodeIDs)
	if err != nil || !sameStringSlices(execution.WriteSet, expectedWrites) {
		return fail("execution write set differs from C4/C5: %v", err)
	}
	observationIDs := make([]ID, len(observations))
	outputIDs := make([]ID, len(observations))
	planActions := idSet(evidence.Plan.ActionNodeIDs)
	planOutputs := idSet(evidence.Plan.DeclaredOutputNodeIDs)
	for index, observation := range observations {
		if err := observation.ValidateAgainst(evidence.Graph.Records); err != nil {
			return fail("observation[%d] differs from immutable graph records: %v", index, err)
		}
		if !planActions[observation.ProducerActionID] || !planOutputs[observation.ExpectedOutputNodeID] {
			return fail("observation[%d] action or output is absent from C5", index)
		}
		if !containsString(execution.WriteSet, observation.Path) {
			return fail("observation[%d] path is absent from the exact write set", index)
		}
		observationIDs[index], _ = observation.ID()
		outputIDs[index] = observation.ExpectedOutputNodeID
	}
	if !strictlySortedUniqueIDs(observationIDs) || !sameIDSlices(observationIDs, execution.ProducedObservationIDs) {
		return fail("observation identities differ from the execution receipt")
	}
	if !sameIDSlices(sortedIDs(outputIDs), evidence.Plan.DeclaredOutputNodeIDs) {
		return fail("observations do not cover every C5 output exactly once")
	}
	return nil
}
