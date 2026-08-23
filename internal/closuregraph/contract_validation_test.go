package closuregraph

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCaptureInputNodesCannotClaimExternalTrust(t *testing.T) {
	records := loadGoldenRecords(t)
	tests := []struct {
		name string
		node Node
	}{
		{name: "package", node: mustGolden[Node](t, records, "cgp05.extra")},
		{name: "source", node: mustGolden[Node](t, records, "cgp10.source")},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			node := testCase.node
			switch payload := node.Payload.(type) {
			case PackageInstancePayload:
				payload.TrustRole = TrustExternalToolchain
				node.Payload = payload
			case SourceSetPayload:
				payload.TrustRole = TrustLocalBuildOutput
				node.Payload = payload
			}
			if err := node.Validate(); err == nil || !strings.Contains(err.Error(), string(TrustDependencyInput)) {
				t.Fatalf("invalid capture trust role error = %v", err)
			}
		})
	}
}

func TestActiveGraphRejectsInconsistentConditionalResults(t *testing.T) {
	base := mustGolden[ActiveGraph](t, loadGoldenRecords(t), "cgp05.active.linux")
	tests := []struct {
		name       string
		evaluation bool
		state      ActivationState
		reason     ActivationReason
	}{
		{name: "false selected", evaluation: false, state: ActivationSelected, reason: ReasonConditionTrue},
		{name: "true pruned false", evaluation: true, state: ActivationPruned, reason: ReasonConditionFalse},
		{name: "true selected unreachable", evaluation: true, state: ActivationSelected, reason: ReasonUnreachable},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			graph := base
			graph.EdgeActivations = append([]EdgeActivation{}, base.EdgeActivations...)
			graph.EdgeActivations[0].Evaluation = testCase.evaluation
			graph.EdgeActivations[0].State = testCase.state
			graph.EdgeActivations[0].Reason = testCase.reason
			if err := graph.Validate(); err == nil {
				t.Fatal("inconsistent conditional result was accepted")
			}
		})
	}
}

func TestGraphBundleRejectsActivationReachabilityDrift(t *testing.T) {
	input := cgp10Inputs(t)
	bundle, err := projectInputs(t, input)
	if err != nil {
		t.Fatal(err)
	}
	actionID := nodeIDByKind(t, input.records.CaptureNodes, NodeAction)
	for index := range bundle.Active.NodeActivations {
		if bundle.Active.NodeActivations[index].NodeID == actionID {
			bundle.Active.NodeActivations[index].State = ActivationPruned
		}
	}
	requireReferenceIssue(t, bundle.Validate(), "selected reachability")
}

func TestActiveProjectionDoesNotSelectProductsThroughSharedDependencies(t *testing.T) {
	selectedProduct := fixtureProduct("selected")
	prunedProduct := fixtureProduct("pruned")
	sharedPackage := Node{Kind: NodePackageInstance, LogicalKey: "package:shared", Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "node", Manager: "npm", Origin: "registry://shared/1.0.0", LockInstanceKey: "shared@1.0.0", Name: "shared", Version: "1.0.0", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput}}
	selectedID := mustNodeID(t, selectedProduct)
	prunedID := mustNodeID(t, prunedProduct)
	sharedID := mustNodeID(t, sharedPackage)
	captureEdges := []Edge{
		{Kind: EdgeRequires, EdgeKey: "edge:selected-requires-shared", FromNodeID: selectedID, ToNodeID: sharedID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("selected.dependencies.shared")}},
		{Kind: EdgeRequires, EdgeKey: "edge:pruned-requires-shared", FromNodeID: prunedID, ToNodeID: sharedID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("pruned.dependencies.shared")}},
	}
	captureNodes := []Node{selectedProduct, prunedProduct, sharedPackage}
	capture, err := NewCaptureGraph("fixture-source-v1", []string{"fixture-policy-v1"}, []ID{selectedID, prunedID}, captureNodes, captureEdges, []ID{testDigest('1')})
	if err != nil {
		t.Fatal(err)
	}
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:fixture", Payload: TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "fixture-sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu"}}
	platformID := mustNodeID(t, platform)
	selection, err := NewSelectionContext([]ID{selectedID}, map[PlatformRole]ID{PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	targetSelected := Edge{Kind: EdgeTargets, EdgeKey: "edge:selected-targets-platform", FromNodeID: selectedID, ToNodeID: platformID, Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}}}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, []Node{platform}, []Edge{targetSelected})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := ProjectActive(capture, selection, binding, NewRecordTables(captureNodes, captureEdges, []Node{platform}, []Edge{targetSelected}), BindingAuthority{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	states := map[ID]ActivationState{}
	for _, activation := range bundle.Active.NodeActivations {
		states[activation.NodeID] = activation.State
	}
	if states[selectedID] != ActivationSelected || states[sharedID] != ActivationSelected || states[prunedID] != ActivationPruned {
		t.Fatalf("active states = %#v", states)
	}

	targetPruned := targetSelected
	targetPruned.EdgeKey = "edge:pruned-targets-platform"
	targetPruned.FromNodeID = prunedID
	bindingEdges := []Edge{targetSelected, targetPruned}
	binding, err = NewSelectionBinding(captureID, selectionID, []Node{platform}, bindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ProjectActive(capture, selection, binding, NewRecordTables(captureNodes, captureEdges, []Node{platform}, bindingEdges), BindingAuthority{}, nil)
	requireReferenceIssue(t, err, "pruned endpoint")
}

func TestConditionalEdgeRequiresEvaluatorSelectedByContext(t *testing.T) {
	capture, selection, _, tables := cgp05ContractInputs(t)
	selection.EvaluatorIDs = []string{}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, tables.BindingNodes, tables.BindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ProjectActive(capture, selection, binding, tables, BindingAuthority{}, nil)
	requireReferenceIssue(t, err, "absent from selection")
}

func TestConditionEvaluatorRegistryFailsClosed(t *testing.T) {
	capture, selection, binding, tables := cgp05ContractInputs(t)
	valid := ConditionEvaluatorFunc{EvaluatorID: "fixture-target-v1", EvaluateFunc: func(Condition, EvaluationInput) (bool, error) { return false, nil }}
	tests := []struct {
		name       string
		evaluators []ConditionEvaluator
		contains   string
	}{
		{name: "nil evaluator", evaluators: []ConditionEvaluator{nil}, contains: "nil condition evaluator"},
		{name: "unselected evaluator", evaluators: []ConditionEvaluator{ConditionEvaluatorFunc{EvaluatorID: "other-v1", EvaluateFunc: valid.EvaluateFunc}}, contains: "not selected"},
		{name: "duplicate evaluator", evaluators: []ConditionEvaluator{valid, valid}, contains: "duplicate evaluator"},
		{name: "missing implementation", evaluators: []ConditionEvaluator{ConditionEvaluatorFunc{EvaluatorID: "fixture-target-v1"}}, contains: "no implementation"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ProjectActive(capture, selection, binding, tables, BindingAuthority{}, testCase.evaluators)
			if err == nil || !strings.Contains(err.Error(), testCase.contains) {
				t.Fatalf("evaluator registry error = %v", err)
			}
		})
	}
}

func TestC0RequiresExactTargetPlatformMembership(t *testing.T) {
	valid := checkpointPayloadFixtures()[0].(C0ProfilePayload)
	tests := []struct {
		name   string
		mutate func(C0ProfilePayload) C0ProfilePayload
	}{
		{name: "no platform nodes", mutate: func(payload C0ProfilePayload) C0ProfilePayload {
			payload.PlatformNodeIDs = []ID{}
			return payload
		}},
		{name: "no target role", mutate: func(payload C0ProfilePayload) C0ProfilePayload {
			payload.PlatformRoles = map[PlatformRole]ID{}
			return payload
		}},
		{name: "role outside table", mutate: func(payload C0ProfilePayload) C0ProfilePayload {
			payload.PlatformRoles = map[PlatformRole]ID{PlatformTarget: testDigest('f')}
			return payload
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := NewCheckpoint(testCase.mutate(valid), nil, []Diagnostic{}); err == nil {
				t.Fatal("invalid C0 platform membership was accepted")
			}
		})
	}
}

func TestCheckpointChainRejectsSelectionContextDrift(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{name: "C1", index: 1},
		{name: "C4", index: 4},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			payloads := checkpointPayloadFixtures()
			if testCase.index == 1 {
				payload := payloads[1].(C1ResolvePayload)
				payload.SelectionContextID = testDigest('f')
				payloads[1] = payload
			} else {
				payload := payloads[4].(C4ClosePayload)
				payload.SelectionContextID = testDigest('f')
				payloads[4] = payload
			}
			chain := buildCheckpointChain(t, payloads)
			if err := validateCheckpointSequence(chain); err == nil || !strings.Contains(err.Error(), testCase.name+" selection_context_id") {
				t.Fatalf("selection drift error = %v", err)
			}
		})
	}
}

func TestCheckpointDiagnosticsAndShapeFailClosed(t *testing.T) {
	if _, err := NewCheckpoint(nil, nil, nil); err == nil {
		t.Fatal("nil checkpoint payload was accepted")
	}
	diagnostics := []Diagnostic{
		{Code: CodeGraphReferenceInvalid, Subject: "z", Fields: map[string]string{"kind": "edge"}},
		{Code: CodeGraphIncomplete, Subject: "a", Fields: map[string]string{}},
	}
	c0, err := NewCheckpoint(checkpointPayloadFixtures()[0], nil, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if c0.Diagnostics[0].Code != CodeGraphIncomplete {
		t.Fatalf("diagnostics not canonical: %#v", c0.Diagnostics)
	}
	wire, err := c0.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCheckpoint(wire)
	if err != nil || !reflect.DeepEqual(decoded, c0) {
		t.Fatalf("diagnostic checkpoint round trip = %#v, %v", decoded, err)
	}

	previousID := testDigest('1')
	badC0 := c0
	badC0.PreviousCheckpointID = &previousID
	if err := badC0.Validate(); err == nil {
		t.Fatal("C0 predecessor was accepted")
	}
	badC1 := Checkpoint{SchemaID: SchemaCheckpoint, Name: CheckpointC1, Payload: checkpointPayloadFixtures()[1], Decision: DecisionAdmit, Diagnostics: []Diagnostic{}}
	if err := badC1.Validate(); err == nil {
		t.Fatal("missing C1 predecessor was accepted")
	}
	badDecision := c0
	badDecision.Decision = DecisionPublished
	if err := badDecision.Validate(); err == nil {
		t.Fatal("wrong checkpoint decision was accepted")
	}
	badDiagnostics := c0
	badDiagnostics.Diagnostics = nil
	if err := badDiagnostics.Validate(); err == nil {
		t.Fatal("nil diagnostics were accepted")
	}
	duplicateDiagnostics := c0
	duplicateDiagnostics.Diagnostics = []Diagnostic{{Code: CodeGraphIncomplete, Subject: "same", Fields: map[string]string{}}, {Code: CodeGraphIncomplete, Subject: "same", Fields: map[string]string{"field": "other"}}}
	if err := duplicateDiagnostics.Validate(); err == nil {
		t.Fatal("duplicate diagnostics were accepted")
	}
	invalidDiagnostic := c0
	invalidDiagnostic.Diagnostics = []Diagnostic{{Code: CodeGraphIncomplete, Subject: "subject", Fields: nil}}
	if err := invalidDiagnostic.Validate(); err == nil {
		t.Fatal("nil diagnostic fields were accepted")
	}
}

func TestSourceClosureReceiptsAndErrorCodes(t *testing.T) {
	chain := buildCheckpointChain(t, checkpointPayloadFixtures())
	closure, err := NewSourceClosure(chain[5])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewSourceClosure(chain[4]); err == nil {
		t.Fatal("non-C5 source closure was accepted")
	}
	closureID, _ := closure.ID()
	if err := (ExpectedCacheInput{SchemaID: SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: []ID{}}).Validate(); err == nil {
		t.Fatal("empty expected output set was accepted")
	}

	records := loadGoldenRecords(t)
	observation := mustGolden[ProducedArtifactObservation](t, records, "cgp10.observation.one")
	input := cgp10Inputs(t)
	if err := observation.ValidateAgainst(input.records); err != nil {
		t.Fatal(err)
	}
	drifted := observation
	drifted.Path = "bin/other"
	if err := drifted.ValidateAgainst(input.records); err == nil || !strings.Contains(err.Error(), "artifact_local_output_drift") {
		t.Fatalf("output drift error = %v", err)
	}
	negative := observation
	negative.Size = -1
	if err := negative.Validate(); err == nil {
		t.Fatal("negative observation size was accepted")
	}

	execution := mustGolden[ExecutionReceipt](t, records, "cgp10.execution.one")
	execution.Network = "allowed"
	if err := execution.Validate(); err == nil {
		t.Fatal("networked execution receipt was accepted")
	}
	publication := mustGolden[PublicationReceipt](t, records, "cgp10.publication.one")
	publication.ProtectedResult = "copied"
	if err := publication.Validate(); err == nil {
		t.Fatal("unprotected publication receipt was accepted")
	}

	validation := &ValidationError{Issues: []Issue{{Code: CodeGraphIncomplete}}}
	if text := validation.Error(); !strings.Contains(text, string(CodeGraphIncomplete)) {
		t.Fatalf("validation error text = %q", text)
	}
	if got := ErrorCode(validation); got != CodeGraphIncomplete {
		t.Fatalf("validation error code = %q", got)
	}
	cycle := &BuildCycleError{CycleDigest: testDigest('c')}
	if text := cycle.Error(); !strings.Contains(text, string(CodeBuildCycle)) {
		t.Fatalf("cycle error text = %q", text)
	}
	if got := ErrorCode(cycle); got != CodeBuildCycle {
		t.Fatalf("cycle error code = %q", got)
	}
	if got := ErrorCode(errors.New("plain")); got != "" {
		t.Fatalf("plain error code = %q", got)
	}
}

func TestStrictDecodersRejectNoncanonicalAndUnknownRecords(t *testing.T) {
	if _, err := DecodeEdge([]byte("{ \"kind\": \"declares\" }")); err == nil {
		t.Fatal("noncanonical JSON was accepted")
	}
	records := loadGoldenRecords(t)
	raw, err := decodeCanonicalObject(records["cgp10.plan"].Payload, "test plan")
	if err != nil {
		t.Fatal(err)
	}
	raw["recursive_reference"] = string(records["cgp10.plan"].ID)
	wire, _ := canonicalMapBytes(raw)
	if _, err := DecodeBuildPlan(wire); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown build-plan field error = %v", err)
	}
	raw, _ = decodeCanonicalObject(records["cgp10.c4"].Payload, "test checkpoint")
	raw["checkpoint_name"] = "C8.magic"
	wire, _ = canonicalMapBytes(raw)
	if _, err := DecodeCheckpoint(wire); err == nil || !strings.Contains(err.Error(), string(CodeGraphSchemaUnsupported)) {
		t.Fatalf("unknown checkpoint error = %v", err)
	}
}

func TestCanonicalEnvelopeValidatorsFailClosed(t *testing.T) {
	records := loadGoldenRecords(t)
	capture := mustGolden[CaptureGraph](t, records, "cgp10.capture")
	selection := mustGolden[SelectionContext](t, records, "cgp10.selection")
	binding := mustGolden[SelectionBinding](t, records, "cgp10.binding")
	active := mustGolden[ActiveGraph](t, records, "cgp10.active")
	plan := mustGolden[BuildPlan](t, records, "cgp10.plan")
	tests := []struct {
		name     string
		validate func() error
	}{
		{name: "capture schema", validate: func() error { value := capture; value.SchemaID = "closure-capture-graph-v2"; return value.Validate() }},
		{name: "capture roots", validate: func() error { value := capture; value.RootNodeIDs = []ID{}; return value.Validate() }},
		{name: "capture node array", validate: func() error { value := capture; value.NodeIDs = nil; return value.Validate() }},
		{name: "capture policy order", validate: func() error { value := capture; value.PolicyIDs = []string{"z", "a"}; return value.Validate() }},
		{name: "selection products", validate: func() error { value := selection; value.ProductNodeIDs = []ID{}; return value.Validate() }},
		{name: "selection roles", validate: func() error { value := selection; value.PlatformRoles = nil; return value.Validate() }},
		{name: "selection target", validate: func() error { value := selection; value.PlatformRoles = map[PlatformRole]ID{}; return value.Validate() }},
		{name: "binding nodes", validate: func() error { value := binding; value.BindingNodeIDs = []ID{}; return value.Validate() }},
		{name: "binding edge array", validate: func() error { value := binding; value.BindingEdgeIDs = nil; return value.Validate() }},
		{name: "active node array", validate: func() error { value := active; value.NodeActivations = nil; return value.Validate() }},
		{name: "active edge array", validate: func() error { value := active; value.EdgeActivations = nil; return value.Validate() }},
		{name: "active SCC array", validate: func() error { value := active; value.NonOrderingSCCs = nil; return value.Validate() }},
		{name: "plan action array", validate: func() error { value := plan; value.ActionNodeIDs = nil; return value.Validate() }},
		{name: "plan ordering array", validate: func() error { value := plan; value.OrderingEdges = nil; return value.Validate() }},
		{name: "node payload", validate: func() error { return (Node{Kind: NodeAction, LogicalKey: "action:missing"}).Validate() }},
		{name: "node kind", validate: func() error {
			return (Node{Kind: "magic", LogicalKey: "node:magic", Payload: fixtureProduct("magic").Payload}).Validate()
		}},
		{name: "edge payload", validate: func() error {
			return (Edge{Kind: EdgeReads, EdgeKey: "edge:missing", FromNodeID: testDigest('1'), ToNodeID: testDigest('2')}).Validate()
		}},
		{name: "edge kind", validate: func() error {
			return (Edge{Kind: "magic", EdgeKey: "edge:magic", FromNodeID: testDigest('1'), ToNodeID: testDigest('2'), Payload: DeclaresPayload{Origin: fixtureOrigin("magic")}}).Validate()
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.validate(); err == nil {
				t.Fatal("invalid canonical envelope was accepted")
			}
		})
	}
}

func buildCheckpointChain(t *testing.T, payloads []CheckpointPayload) []Checkpoint {
	t.Helper()
	chain := make([]Checkpoint, 0, len(payloads))
	var previous *Checkpoint
	for _, payload := range payloads {
		checkpoint, err := NewCheckpoint(payload, previous, []Diagnostic{})
		if err != nil {
			t.Fatal(err)
		}
		chain = append(chain, checkpoint)
		previous = &chain[len(chain)-1]
	}
	return chain
}

func cgp05ContractInputs(t *testing.T) (CaptureGraph, SelectionContext, SelectionBinding, RecordTables) {
	t.Helper()
	records := loadGoldenRecords(t)
	capture := mustGolden[CaptureGraph](t, records, "cgp05.capture")
	selection := mustGolden[SelectionContext](t, records, "cgp05.selection.darwin")
	binding := mustGolden[SelectionBinding](t, records, "cgp05.binding.darwin")
	tables := NewRecordTables(
		[]Node{mustGolden[Node](t, records, "cgp05.root"), mustGolden[Node](t, records, "cgp05.extra")},
		[]Edge{mustGolden[Edge](t, records, "cgp05.requires")},
		[]Node{mustGolden[Node](t, records, "cgp05.platform.darwin")},
		[]Edge{mustGolden[Edge](t, records, "cgp05.targets.darwin")},
	)
	return capture, selection, binding, tables
}
