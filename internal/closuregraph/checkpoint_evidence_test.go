package closuregraph

import (
	"strings"
	"testing"
)

func TestCheckpointChainResolvesEveryC0ThroughC7Record(t *testing.T) {
	chain, evidence := checkpointEvidenceFixture(t)
	if err := ValidateCheckpointChain(chain, evidence); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointChainResolvesCargoC3aAndC3bAggregates(t *testing.T) {
	chain, evidence := checkpointEvidenceFixtureWithCargo(t, true)
	if err := ValidateCheckpointChain(chain, evidence); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointChainRejectsSelfConsistentPlanRecordsAbsentFromActiveGraph(t *testing.T) {
	chain, evidence := checkpointEvidenceFixture(t)
	foreignActionID := testDigest('f')
	plan := evidence.Plan
	plan.ActionNodeIDs = sortedIDs(append(append([]ID{}, plan.ActionNodeIDs...), foreignActionID))
	plan.Waves = [][]ID{append([]ID{}, plan.ActionNodeIDs...)}
	if err := plan.Validate(); err != nil {
		t.Fatalf("mutated plan must remain internally valid: %v", err)
	}
	planID, err := plan.ID()
	if err != nil {
		t.Fatal(err)
	}

	payloads := chainPayloads(chain)
	payloads[5] = C5PlanPayload{BuildPlanID: planID}
	partial := buildCheckpointChain(t, payloads[:6])
	closure, err := NewSourceClosure(partial[5])
	if err != nil {
		t.Fatal(err)
	}
	closureID, err := closure.ID()
	if err != nil {
		t.Fatal(err)
	}
	expected := evidence.ExpectedCacheInput
	expected.ClosureID = closureID
	expectedID, err := expected.ID()
	if err != nil {
		t.Fatal(err)
	}
	execution := evidence.Execution
	execution.ActionOrder = append([]ID{}, plan.ActionNodeIDs...)
	execution.ClosureID = closureID
	executionID, err := execution.ID()
	if err != nil {
		t.Fatal(err)
	}
	publication := evidence.Publication
	publication.ExecutionReceiptID = executionID
	publication.ExpectedCacheInputID = expectedID
	publicationID, err := publication.ID()
	if err != nil {
		t.Fatal(err)
	}
	payloads[6] = C6OfflinePayload{ExecutionReceiptID: executionID}
	payloads[7] = C7PublishPayload{PublicationReceiptID: publicationID}
	evidence.Plan = plan
	evidence.Closure = closure
	evidence.ExpectedCacheInput = expected
	evidence.Execution = execution
	evidence.Publication = publication
	chain = buildCheckpointChain(t, payloads)

	err = ValidateCheckpointChain(chain, evidence)
	if err == nil || !strings.Contains(err.Error(), string(CodeCheckpointInvalid)) || !strings.Contains(err.Error(), "exact action, output, ordering, and wave projection") {
		t.Fatalf("error = %v, want exact-plan projection rejection for foreign action %s", err, foreignActionID)
	}
}

func TestCheckpointChainRejectsCrossRecordTrustDrift(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		mutate func(*testing.T, []Checkpoint, CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence)
	}{
		{
			name: "C0 platform resolves to toolchain",
			want: "platform_node_ids",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				toolID := nodeIDByKind(t, evidence.Graph.Records.BindingNodes, NodeToolchainComponent)
				payloads := chainPayloads(chain)
				c0 := payloads[0].(C0ProfilePayload)
				c0.PlatformNodeIDs = []ID{toolID}
				c0.PlatformRoles = map[PlatformRole]ID{PlatformTarget: toolID}
				payloads[0] = c0
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C0 evidence tool is wrong kind",
			want: "evidence_toolchain_node_ids",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				platformID := nodeIDByKind(t, evidence.Graph.Records.BindingNodes, NodeTargetPlatform)
				payloads := chainPayloads(chain)
				c0 := payloads[0].(C0ProfilePayload)
				c0.EvidenceToolchainNodeIDs = []ID{platformID}
				payloads[0] = c0
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C1 candidate graph is stale",
			want: "C1 candidate graph",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c1 := payloads[1].(C1ResolvePayload)
				c1.CandidateNodeIDs = append([]ID{}, c1.CandidateNodeIDs[1:]...)
				payloads[1] = c1
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C1 root declaration is missing",
			want: "root/workspace declarations",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c1 := payloads[1].(C1ResolvePayload)
				c1.RootDeclarationIDs = []ID{}
				payloads[1] = c1
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C1 lock is unrelated",
			want: "C1 lock or journal",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c1 := payloads[1].(C1ResolvePayload)
				c1.LockCandidateID = testDigest('f')
				payloads[1] = c1
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C2 protected handle is unrelated",
			want: "C2 aggregate",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c2 := payloads[2].(C2CapturePayload)
				c2.ProtectedHandleIDs = []ID{testDigest('f')}
				payloads[2] = c2
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C2 and C3 intake sets differ",
			want: "intake_receipt_ids do not match C2",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c3 := payloads[3].(C3AdmitPayload)
				c3.IntakeReceiptIDs = []ID{testDigest('f')}
				payloads[3] = c3
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C3 admission is unrelated",
			want: "artifact_manifest_ids",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c3 := payloads[3].(C3AdmitPayload)
				c3.ArtifactManifestIDs = []ID{testDigest('f')}
				payloads[3] = c3
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C3 derivation receipt is unrelated",
			want: "final C3 aggregate",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c3 := payloads[3].(C3AdmitPayload)
				c3.DerivationReceiptIDs = []ID{testDigest('f')}
				payloads[3] = c3
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C4 active graph is stale",
			want: "C4 graph references",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				c4 := payloads[4].(C4ClosePayload)
				c4.ActiveGraphID = testDigest('f')
				payloads[4] = c4
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C5 plan is stale",
			want: "C5 plan reference",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				payloads[5] = C5PlanPayload{BuildPlanID: testDigest('f')}
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C5 execution policy is not independently authorized",
			want: "independently supplied execution policy",
			mutate: func(_ *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				evidence.ExecutionPolicyID = "different-execution-policy-v1"
				return chain, evidence
			},
		},
		{
			name: "C6 receipt is unrelated",
			want: "C6 execution reference",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				payloads[6] = C6OfflinePayload{ExecutionReceiptID: testDigest('f')}
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "C7 receipt is unrelated",
			want: "C7 publication reference",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				payloads := chainPayloads(chain)
				payloads[7] = C7PublishPayload{PublicationReceiptID: testDigest('f')}
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "closure names a different C5",
			want: "closure does not derive",
			mutate: func(_ *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				evidence.Closure.C5CheckpointID = testDigest('f')
				return chain, evidence
			},
		},
		{
			name: "observation set drifts from execution",
			want: "execution observation IDs",
			mutate: func(_ *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				observations := append([]ProducedArtifactObservation{}, evidence.Observations...)
				observations[0].SHA256 = testDigest('f')
				evidence.Observations = observations
				return chain, evidence
			},
		},
		{
			name: "execution names a different closure",
			want: "C6 execution reference",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				execution := evidence.Execution
				execution.ClosureID = testDigest('f')
				evidence.Execution = execution
				executionID, _ := execution.ID()
				publication := evidence.Publication
				publication.ExecutionReceiptID = executionID
				evidence.Publication = publication
				publicationID, _ := publication.ID()
				payloads := chainPayloads(chain)
				payloads[6] = C6OfflinePayload{ExecutionReceiptID: executionID}
				payloads[7] = C7PublishPayload{PublicationReceiptID: publicationID}
				return buildCheckpointChain(t, payloads), evidence
			},
		},
		{
			name: "expected cache input names different outputs",
			want: "expected cache input",
			mutate: func(_ *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				evidence.ExpectedCacheInput.ExpectedOutputNodeIDs = []ID{testDigest('f')}
				return chain, evidence
			},
		},
		{
			name: "publication names a different expected input",
			want: "C7 publication reference",
			mutate: func(t *testing.T, chain []Checkpoint, evidence CheckpointChainEvidence) ([]Checkpoint, CheckpointChainEvidence) {
				publication := evidence.Publication
				publication.ExpectedCacheInputID = testDigest('f')
				evidence.Publication = publication
				publicationID, _ := publication.ID()
				payloads := chainPayloads(chain)
				payloads[7] = C7PublishPayload{PublicationReceiptID: publicationID}
				return buildCheckpointChain(t, payloads), evidence
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			chain, evidence := checkpointEvidenceFixture(t)
			chain, evidence = testCase.mutate(t, chain, evidence)
			err := ValidateCheckpointChain(chain, evidence)
			if err == nil || !strings.Contains(err.Error(), string(CodeCheckpointInvalid)) || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("error = %v, want checkpoint failure containing %q", err, testCase.want)
			}
		})
	}
}

func checkpointEvidenceFixture(t *testing.T) ([]Checkpoint, CheckpointChainEvidence) {
	return checkpointEvidenceFixtureWithCargo(t, false)
}

func checkpointEvidenceFixtureWithCargo(t *testing.T, cargo bool) ([]Checkpoint, CheckpointChainEvidence) {
	t.Helper()
	input := cgp10Inputs(t)
	graph, err := projectInputs(t, input)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := DeriveBuildPlan(graph, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	selectionID, _ := graph.Selection.ID()
	platformIDs := []ID{nodeIDByKind(t, graph.Records.BindingNodes, NodeTargetPlatform)}
	conditionIDs := []ID{}
	for _, edge := range graph.Records.CaptureEdges {
		if edge.Payload.condition() != nil {
			conditionIDs = append(conditionIDs, mustEdgeID(t, edge))
		}
	}
	payloads := []CheckpointPayload{
		C0ProfilePayload{AdapterProfileID: "fixture-source-v1", SchemaIDs: []string{"closure-v1"}, ArtifactPolicyID: "artifact-policy-v1", DetectorRegistryID: "detectors-v1", SourceGrammarIDs: []string{"source-v1"}, LimitVectorID: "limits-v1", SelectionContextID: selectionID, PlatformNodeIDs: platformIDs, PlatformRoles: clonePlatformRoles(graph.Selection.PlatformRoles), ManagerSchemaIDs: []string{"manager-v1"}, ConfigurationPolicyID: "configuration-v1", CapabilityIDs: []string{}, EvidenceToolchainNodeIDs: []ID{}},
		C1ResolvePayload{RootDeclarationIDs: append([]ID{}, graph.Capture.RootNodeIDs...), WorkspaceDeclarationIDs: []ID{}, LockCandidateID: testDigest('9'), ConditionEdgeIDs: sortedIDs(conditionIDs), ParserEvaluatorIDs: append([]string{}, graph.Selection.EvaluatorIDs...), CandidateNodeIDs: append([]ID{}, graph.Capture.NodeIDs...), CandidateEdgeIDs: append([]ID{}, graph.Capture.EdgeIDs...), SelectionContextID: selectionID, JournalEntryIDs: []ID{}},
		C2CapturePayload{IntakeReceiptIDs: []ID{testDigest('1')}, OriginIDs: []ID{testDigest('2')}, ProtectedHandleIDs: []ID{testDigest('3')}, BrokerReceiptIDs: []ID{testDigest('4')}},
		C3AdmitPayload{Phase: "main", IntakeReceiptIDs: []ID{testDigest('1')}, ArtifactManifestIDs: append([]ID{}, graph.Capture.ArtifactManifestIDs...), DerivationReceiptIDs: []ID{}},
	}
	if cargo {
		main := payloads[3].(C3AdmitPayload)
		main.DerivationReceiptIDs = []ID{testDigest('5')}
		payloads = append(payloads[:3],
			C3AdmitPayload{Phase: "origin", IntakeReceiptIDs: []ID{testDigest('1')}, ArtifactManifestIDs: append([]ID{}, graph.Capture.ArtifactManifestIDs...), DerivationReceiptIDs: []ID{}},
			C3AdmitPayload{Phase: "derived", IntakeReceiptIDs: []ID{testDigest('1')}, ArtifactManifestIDs: []ID{testDigest('6')}, DerivationReceiptIDs: []ID{testDigest('5')}},
			main,
		)
	}
	activeID, _ := graph.Active.ID()
	captureID, _ := graph.Capture.ID()
	bindingID, _ := graph.Binding.ID()
	payloads = append(payloads, C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: captureID, SelectionBindingID: bindingID, SelectionContextID: selectionID})
	planID, _ := plan.ID()
	payloads = append(payloads, C5PlanPayload{BuildPlanID: planID})
	partial := buildCheckpointChain(t, payloads)
	closure, err := NewSourceClosure(partial[len(partial)-1])
	if err != nil {
		t.Fatal(err)
	}
	closureID, _ := closure.ID()
	observation := mustGolden[ProducedArtifactObservation](t, loadGoldenRecords(t), "cgp10.observation.one")
	observationID, _ := observation.ID()
	writes, err := selectedWritePaths(graph, plan.ActionNodeIDs)
	if err != nil {
		t.Fatal(err)
	}
	execution := ExecutionReceipt{SchemaID: SchemaExecutionReceipt, ActionOrder: append([]ID{}, plan.ActionNodeIDs...), ClosureID: closureID, Decision: "success", Network: "none", ProducedObservationIDs: []ID{observationID}, ToolchainRechecks: "match", WriteSet: writes}
	executionID, _ := execution.ID()
	payloads = append(payloads, C6OfflinePayload{ExecutionReceiptID: executionID})
	expected := ExpectedCacheInput{SchemaID: SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: append([]ID{}, plan.DeclaredOutputNodeIDs...)}
	expectedID, _ := expected.ID()
	publication := PublicationReceipt{SchemaID: SchemaPublicationReceipt, Decision: "published", ExecutionReceiptID: executionID, ExpectedCacheInputID: expectedID, ProtectedResult: "exact_write", PublishedObservationIDs: []ID{observationID}}
	publicationID, _ := publication.ID()
	payloads = append(payloads, C7PublishPayload{PublicationReceiptID: publicationID})
	chain := buildCheckpointChain(t, payloads)
	derivationReceiptIDs := []ID{}
	if cargo {
		derivationReceiptIDs = []ID{testDigest('5')}
	}
	recordEvidence := CheckpointRecordEvidence{RootDeclarationIDs: append([]ID{}, graph.Capture.RootNodeIDs...), WorkspaceDeclarationIDs: []ID{}, LockCandidateID: testDigest('9'), JournalEntryIDs: []ID{}, IntakeReceiptIDs: []ID{testDigest('1')}, OriginIDs: []ID{testDigest('2')}, ProtectedHandleIDs: []ID{testDigest('3')}, BrokerReceiptIDs: []ID{testDigest('4')}, ArtifactManifestIDs: append([]ID{}, graph.Capture.ArtifactManifestIDs...), DerivationReceiptIDs: derivationReceiptIDs}
	return chain, CheckpointChainEvidence{Records: recordEvidence, Graph: graph, ExecutionPolicyID: plan.ExecutionPolicyID, Plan: plan, Closure: closure, ExpectedCacheInput: expected, Observations: []ProducedArtifactObservation{observation}, Execution: execution, Publication: publication}
}

func chainPayloads(chain []Checkpoint) []CheckpointPayload {
	payloads := make([]CheckpointPayload, len(chain))
	for index := range chain {
		payloads[index] = chain[index].Payload
	}
	return payloads
}

func mustEdgeID(t *testing.T, edge Edge) ID {
	t.Helper()
	id, err := edge.ID()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
