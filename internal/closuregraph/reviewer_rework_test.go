package closuregraph

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReviewerReworkTargetRequiresProducedOutputOrdersConsumer(t *testing.T) {
	product := fixtureProduct("requires-output-order")
	target := fixtureTarget("requires-output-consumer", "go")
	provider := fixtureAction("requires-output-provider", []string{}, []string{"artifact"})
	consumer := fixtureAction("requires-output-consumer", []string{}, []string{})
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:requires-output", Payload: OutputArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "lib/generated.a", ExpectedClass: "native.static_library", OutputRole: "build_input",
	}}
	productID := mustNodeID(t, product)
	targetID := mustNodeID(t, target)
	providerID := mustNodeID(t, provider)
	consumerID := mustNodeID(t, consumer)
	outputID := mustNodeID(t, output)
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-provider", FromNodeID: productID, ToNodeID: providerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.provider")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-consumer-target", FromNodeID: productID, ToNodeID: targetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:target-declares-consumer", FromNodeID: targetID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
		{Kind: EdgeProduces, EdgeKey: "edge:provider-produces-output", FromNodeID: providerID, ToNodeID: outputID, Payload: ProducesPayload{Path: "lib/generated.a", WriteSlot: "artifact", WriteClass: "native.static_library"}},
		{Kind: EdgeRequires, EdgeKey: "edge:target-requires-output", FromNodeID: targetID, ToNodeID: outputID, Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("targets.consumer.requires.output")}},
	}
	bundle := fixtureBundle(t, []Node{product, target, provider, consumer, output}, edges, product)
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range plan.OrderingEdges {
		if edge.FromActionID == providerID && edge.ToActionID == consumerID {
			return
		}
	}
	t.Fatalf("selected target requires produced output but plan has no provider-before-consumer edge: %#v", plan.OrderingEdges)
}

func TestReviewerReworkOwnerRequiresProducedOutputOrderingIsCanonical(t *testing.T) {
	for _, ownerKind := range []NodeKind{NodeCommandProduct, NodePackageInstance, NodeTargetUnit} {
		t.Run(string(ownerKind), func(t *testing.T) {
			nodes, edges, product, providerID, consumerID := ownerRequiresOutputFixture(t, ownerKind)
			left := assertOutputRequirementPlan(t, fixtureBundle(t, nodes, edges, product), providerID, consumerID)

			reverseNodes(nodes)
			reverseEdges(edges)
			permuted := fixtureBundle(t, nodes, edges, product)
			right := assertOutputRequirementPlan(t, permuted, providerID, consumerID)
			leftID, leftErr := left.ID()
			rightID, rightErr := right.ID()
			if leftErr != nil || rightErr != nil || leftID != rightID || !reflect.DeepEqual(left.OrderingEdges, right.OrderingEdges) {
				t.Fatalf("%s plan changed under record permutation:\nleft: %#v (%v)\nright: %#v (%v)", ownerKind, left, leftErr, right, rightErr)
			}
		})
	}
}

func ownerRequiresOutputFixture(t *testing.T, ownerKind NodeKind) ([]Node, []Edge, Node, ID, ID) {
	t.Helper()
	product := fixtureProduct("owner-output-" + string(ownerKind))
	provider := fixtureAction("owner-output-provider-"+string(ownerKind), []string{}, []string{"artifact"})
	consumer := fixtureAction("owner-output-consumer-"+string(ownerKind), []string{}, []string{})
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:owner-output-" + string(ownerKind), Payload: OutputArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "lib/" + string(ownerKind) + ".a", ExpectedClass: "native.static_library", OutputRole: "build_input",
	}}
	productID := mustNodeID(t, product)
	providerID := mustNodeID(t, provider)
	consumerID := mustNodeID(t, consumer)
	outputID := mustNodeID(t, output)
	nodes := []Node{product, provider, consumer, output}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-owner-output-provider-" + string(ownerKind), FromNodeID: productID, ToNodeID: providerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.provider")}},
		{Kind: EdgeProduces, EdgeKey: "edge:owner-output-provider-produces-" + string(ownerKind), FromNodeID: providerID, ToNodeID: outputID, Payload: ProducesPayload{Path: "lib/" + string(ownerKind) + ".a", WriteSlot: "artifact", WriteClass: "native.static_library"}},
	}
	ownerID := productID
	switch ownerKind {
	case NodeCommandProduct:
		edges = append(edges, Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-owner-output-consumer", FromNodeID: productID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.consumer")}})
	case NodePackageInstance:
		owner := Node{Kind: NodePackageInstance, LogicalKey: "package:owner-output", Payload: PackageInstancePayload{
			Profile: "fixture-source-v1", Ecosystem: "fixture", Manager: "fixture-manager", Origin: "registry://fixture/owner-output/1.0.0", LockInstanceKey: "owner-output@1.0.0", Name: "owner-output", Version: "1.0.0", ArtifactManifestID: testDigest('4'), TrustRole: TrustDependencyInput,
		}}
		ownerID = mustNodeID(t, owner)
		nodes = append(nodes, owner)
		edges = append(edges,
			Edge{Kind: EdgeRequires, EdgeKey: "edge:product-requires-owner-package", FromNodeID: productID, ToNodeID: ownerID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies.owner")}},
			Edge{Kind: EdgeDeclares, EdgeKey: "edge:owner-package-declares-consumer", FromNodeID: ownerID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.consumer")}},
		)
	case NodeTargetUnit:
		owner := fixtureTarget("owner-output", "go")
		ownerID = mustNodeID(t, owner)
		nodes = append(nodes, owner)
		edges = append(edges,
			Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-owner-target", FromNodeID: productID, ToNodeID: ownerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.owner")}},
			Edge{Kind: EdgeDeclares, EdgeKey: "edge:owner-target-declares-consumer", FromNodeID: ownerID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.owner.action")}},
		)
	default:
		t.Fatalf("unsupported owner kind %s", ownerKind)
	}
	edges = append(edges, Edge{Kind: EdgeRequires, EdgeKey: "edge:owner-requires-produced-output-" + string(ownerKind), FromNodeID: ownerID, ToNodeID: outputID, Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("owner.requires.output")}})
	return nodes, edges, product, providerID, consumerID
}

func assertOutputRequirementPlan(t *testing.T, bundle GraphBundle, providerID, consumerID ID) BuildPlan {
	t.Helper()
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range plan.OrderingEdges {
		if edge.FromActionID == providerID && edge.ToActionID == consumerID && edge.Reason == OrderTargetRequirement && len(edge.SourceEdgeIDs) >= 2 {
			producedOutputID := ID("")
			hasInteropConsumer := false
			for _, graphEdge := range bundle.Records.CaptureEdges {
				if graphEdge.Kind == EdgeProduces && graphEdge.FromNodeID == providerID {
					producedOutputID = graphEdge.ToNodeID
				}
				if graphEdge.Kind == EdgeConsumesInterop && graphEdge.FromNodeID == consumerID {
					hasInteropConsumer = true
				}
			}
			expectedEvidence := []ID{}
			for _, graphEdge := range bundle.Records.CaptureEdges {
				isProvider := graphEdge.Kind == EdgeProduces && graphEdge.FromNodeID == providerID
				isConsumer := graphEdge.Kind == EdgeDeclares && graphEdge.ToNodeID == consumerID
				if hasInteropConsumer {
					isConsumer = graphEdge.Kind == EdgeConsumesInterop && graphEdge.FromNodeID == consumerID
				}
				isRequirement := graphEdge.Kind == EdgeRequires && graphEdge.ToNodeID == producedOutputID
				if isProvider || isConsumer || isRequirement {
					id, idErr := graphEdge.ID()
					if idErr != nil {
						t.Fatal(idErr)
					}
					expectedEvidence = append(expectedEvidence, id)
				}
			}
			for _, evidenceID := range expectedEvidence {
				if !containsID(edge.SourceEdgeIDs, evidenceID) {
					t.Fatalf("ordering edge omits canonical source evidence %s: %#v", evidenceID, edge)
				}
			}
			return plan
		}
	}
	t.Fatalf("plan has no canonical produced-output owner ordering: %#v", plan.OrderingEdges)
	return BuildPlan{}
}

func TestReviewerReworkBoundaryRequiresProducedOutputOrdersConsumer(t *testing.T) {
	product := fixtureProduct("boundary-output-order")
	providerTarget := fixtureTarget("boundary-provider", "c")
	consumerTarget := fixtureTarget("boundary-consumer", "swift")
	provider := fixtureAction("boundary-provider", []string{}, []string{"library"})
	consumer := fixtureAction("boundary-consumer", []string{}, []string{})
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:boundary-library", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "lib/boundary.a", ExpectedClass: "native.static_library", OutputRole: "build_input"}}
	boundary := Node{Kind: NodeInteropBoundary, LogicalKey: "interop:boundary-output-order", Payload: InteropBoundaryPayload{
		Profile: "fixture-source-v1", Mode: InteropCABI, ProviderLanguageClasses: []string{"c"}, ConsumerLanguageClasses: []string{"swift"}, ContractDigest: testDigest('5'), ABI: "c-abi-v1", InterfaceContract: "header-v1", CallingConvention: "c", LinkLoadSemantics: "static-link",
	}}
	productID, providerTargetID, consumerTargetID := mustNodeID(t, product), mustNodeID(t, providerTarget), mustNodeID(t, consumerTarget)
	providerID, consumerID, outputID, boundaryID := mustNodeID(t, provider), mustNodeID(t, consumer), mustNodeID(t, output), mustNodeID(t, boundary)
	nodes := []Node{product, providerTarget, consumerTarget, provider, consumer, output, boundary}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-boundary-provider", FromNodeID: productID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-boundary-consumer", FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:boundary-provider-declares-action", FromNodeID: providerTargetID, ToNodeID: providerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider.action")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:boundary-consumer-declares-action", FromNodeID: consumerTargetID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
		{Kind: EdgeProduces, EdgeKey: "edge:boundary-provider-produces-output", FromNodeID: providerID, ToNodeID: outputID, Payload: ProducesPayload{Path: "lib/boundary.a", WriteSlot: "library", WriteClass: "native.static_library"}},
		{Kind: EdgeProvidesInterop, EdgeKey: "edge:boundary-provider-provides", FromNodeID: providerTargetID, ToNodeID: boundaryID, Payload: ProvidesInteropPayload{Origin: fixtureOrigin("interop.provider"), EvidenceIDs: []ID{testDigest('6')}, ExportRole: "headers", LinkMode: "static"}},
		{Kind: EdgeConsumesInterop, EdgeKey: "edge:boundary-consumer-consumes", FromNodeID: consumerID, ToNodeID: boundaryID, Payload: ConsumesInteropPayload{Origin: fixtureOrigin("interop.consumer"), Use: "compile", ABIExpectation: "c-abi-v1"}},
		{Kind: EdgeRequires, EdgeKey: "edge:boundary-requires-produced-output", FromNodeID: boundaryID, ToNodeID: outputID, Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("interop.requires.output")}},
	}
	left := assertOutputRequirementPlan(t, fixtureBundle(t, nodes, edges, product), providerID, consumerID)
	reverseNodes(nodes)
	reverseEdges(edges)
	right := assertOutputRequirementPlan(t, fixtureBundle(t, nodes, edges, product), providerID, consumerID)
	leftID, _ := left.ID()
	rightID, _ := right.ID()
	if leftID != rightID || !reflect.DeepEqual(left.OrderingEdges, right.OrderingEdges) {
		t.Fatalf("boundary output ordering changed under permutation:\nleft: %#v\nright: %#v", left, right)
	}
}

func TestReviewerReworkOwnerOutputRequirementCycleIsCanonical(t *testing.T) {
	product := fixtureProduct("owner-output-cycle")
	targetA, targetB := fixtureTarget("owner-output-cycle-a", "go"), fixtureTarget("owner-output-cycle-b", "go")
	actionA, actionB := fixtureAction("owner-output-cycle-a", []string{}, []string{"a"}), fixtureAction("owner-output-cycle-b", []string{}, []string{"b"})
	outputA := Node{Kind: NodeOutputArtifact, LogicalKey: "output:owner-output-cycle-a", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "lib/a.a", ExpectedClass: "native.static_library", OutputRole: "build_input"}}
	outputB := Node{Kind: NodeOutputArtifact, LogicalKey: "output:owner-output-cycle-b", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "lib/b.a", ExpectedClass: "native.static_library", OutputRole: "build_input"}}
	productID, targetAID, targetBID := mustNodeID(t, product), mustNodeID(t, targetA), mustNodeID(t, targetB)
	actionAID, actionBID, outputAID, outputBID := mustNodeID(t, actionA), mustNodeID(t, actionB), mustNodeID(t, outputA), mustNodeID(t, outputB)
	nodes := []Node{product, targetA, targetB, actionA, actionB, outputA, outputB}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:cycle-product-declares-a", FromNodeID: productID, ToNodeID: targetAID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.a")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:cycle-product-declares-b", FromNodeID: productID, ToNodeID: targetBID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.b")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:cycle-target-a-declares-action", FromNodeID: targetAID, ToNodeID: actionAID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.a.action")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:cycle-target-b-declares-action", FromNodeID: targetBID, ToNodeID: actionBID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.b.action")}},
		{Kind: EdgeProduces, EdgeKey: "edge:cycle-action-a-produces", FromNodeID: actionAID, ToNodeID: outputAID, Payload: ProducesPayload{Path: "lib/a.a", WriteSlot: "a", WriteClass: "native.static_library"}},
		{Kind: EdgeProduces, EdgeKey: "edge:cycle-action-b-produces", FromNodeID: actionBID, ToNodeID: outputBID, Payload: ProducesPayload{Path: "lib/b.a", WriteSlot: "b", WriteClass: "native.static_library"}},
		{Kind: EdgeRequires, EdgeKey: "edge:cycle-target-a-requires-b", FromNodeID: targetAID, ToNodeID: outputBID, Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("targets.a.requires.b")}},
		{Kind: EdgeRequires, EdgeKey: "edge:cycle-target-b-requires-a", FromNodeID: targetBID, ToNodeID: outputAID, Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("targets.b.requires.a")}},
	}
	cycleFor := func(nodes []Node, edges []Edge) *BuildCycleError {
		_, err := DeriveBuildPlan(fixtureBundle(t, nodes, edges, product), PlanOptions{ExecutionPolicyID: "fixture-execution-v1", LastCheckpointID: testDigest('9')})
		var cycle *BuildCycleError
		if !errors.As(err, &cycle) {
			t.Fatalf("output requirement cycle error = %T %v", err, err)
		}
		return cycle
	}
	left := cycleFor(nodes, edges)
	reverseNodes(nodes)
	reverseEdges(edges)
	right := cycleFor(nodes, edges)
	if !reflect.DeepEqual(left, right) || !reflect.DeepEqual(left.ActionNodeIDs, sortedIDs([]ID{actionAID, actionBID})) {
		t.Fatalf("owner output cycle evidence changed under permutation:\nleft: %#v\nright: %#v", left, right)
	}
}

func TestReviewerReworkInteropRejectsNonToolchainCaptureRequirement(t *testing.T) {
	base, _, _ := interopFixture(t, InteropCABI)
	boundaryID := ID("")
	targetID := ID("")
	for _, node := range base.Records.CaptureNodes {
		id := mustNodeID(t, node)
		switch {
		case node.Kind == NodeInteropBoundary:
			boundaryID = id
		case node.Kind == NodeTargetUnit && strings.Contains(node.LogicalKey, "provider"):
			targetID = id
		}
	}
	if boundaryID == "" || targetID == "" {
		t.Fatal("interop fixture lacks boundary or provider target")
	}
	captureEdges := append([]Edge{}, base.Records.CaptureEdges...)
	captureEdges = append(captureEdges, Edge{
		Kind: EdgeRequires, EdgeKey: "edge:boundary-requires-provider-target", FromNodeID: boundaryID, ToNodeID: targetID,
		Payload: RequiresPayload{Scope: ScopeBuild, Origin: fixtureOrigin("interop.fake_tool_binding")},
	})
	capture, err := NewCaptureGraph(base.Capture.ProfileID, base.Capture.PolicyIDs, base.Capture.RootNodeIDs, base.Records.CaptureNodes, captureEdges, base.Capture.ArtifactManifestIDs)
	if err != nil {
		t.Fatal(err)
	}
	bindingEdges := filterEdges(base.Records.BindingEdges, func(edge Edge) bool {
		return edge.Kind != EdgeRequires
	})
	captureID, _ := capture.ID()
	selectionID, _ := base.Selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, base.Records.BindingNodes, bindingEdges)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ProjectActive(capture, base.Selection, binding, NewRecordTables(base.Records.CaptureNodes, captureEdges, base.Records.BindingNodes, bindingEdges), base.Authority, nil)
	if err == nil {
		t.Fatal("selected c_abi boundary accepted a build-scoped capture requirement to a target in place of an explicit toolchain-scoped binding")
	}
}

func TestReviewerReworkInteropToolchainBindingRejectsWrongKindAndDuplicate(t *testing.T) {
	base, _, _ := interopFixture(t, InteropCABI)
	reproject := func(edges []Edge) error {
		captureID, err := base.Capture.ID()
		if err != nil {
			t.Fatal(err)
		}
		selectionID, err := base.Selection.ID()
		if err != nil {
			t.Fatal(err)
		}
		binding, err := NewSelectionBinding(captureID, selectionID, base.Records.BindingNodes, edges)
		if err != nil {
			return err
		}
		_, err = ProjectActive(base.Capture, base.Selection, binding, NewRecordTables(base.Records.CaptureNodes, base.Records.CaptureEdges, base.Records.BindingNodes, edges), base.Authority, nil)
		return err
	}

	t.Run("wrong endpoint kind", func(t *testing.T) {
		edges := append([]Edge{}, base.Records.BindingEdges...)
		platformID := nodeIDByKind(t, base.Records.BindingNodes, NodeTargetPlatform)
		for index := range edges {
			if edges[index].Kind == EdgeRequires {
				edges[index].ToNodeID = platformID
			}
		}
		err := reproject(edges)
		if err == nil || !strings.Contains(err.Error(), "wrong-kind endpoints") {
			t.Fatalf("wrong-kind interop tool binding error = %v", err)
		}
	})

	t.Run("duplicate binding", func(t *testing.T) {
		edges := append([]Edge{}, base.Records.BindingEdges...)
		for _, edge := range base.Records.BindingEdges {
			if edge.Kind != EdgeRequires {
				continue
			}
			duplicate := edge
			duplicate.EdgeKey += ":duplicate"
			payload := duplicate.Payload.(RequiresPayload)
			payload.Origin = fixtureOrigin("selection.interop_toolchains.duplicate")
			duplicate.Payload = payload
			edges = append(edges, duplicate)
			break
		}
		err := reproject(edges)
		if err == nil || !strings.Contains(err.Error(), string(CodeGraphReferenceInvalid)) || !strings.Contains(err.Error(), "duplicate semantic edge") {
			t.Fatalf("duplicate interop tool binding error = %v", err)
		}
	})
}

func TestReviewerReworkDuplicateExpectedOutputPathsRejectBeforeC5(t *testing.T) {
	product := fixtureProduct("duplicate-output-path")
	actionA := fixtureAction("duplicate-output-a", []string{}, []string{"a"})
	actionB := fixtureAction("duplicate-output-b", []string{}, []string{"b"})
	outputA := Node{Kind: NodeOutputArtifact, LogicalKey: "output:duplicate-a", Payload: OutputArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "bin/shared", ExpectedClass: "native.executable", OutputRole: "published_command", DeclarationDigest: testDigest('1'),
	}}
	outputB := Node{Kind: NodeOutputArtifact, LogicalKey: "output:duplicate-b", Payload: OutputArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "bin/shared", ExpectedClass: "native.executable", OutputRole: "published_command", DeclarationDigest: testDigest('2'),
	}}
	productID := mustNodeID(t, product)
	actionAID := mustNodeID(t, actionA)
	actionBID := mustNodeID(t, actionB)
	outputAID := mustNodeID(t, outputA)
	outputBID := mustNodeID(t, outputB)
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-action-a", FromNodeID: productID, ToNodeID: actionAID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.a")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-action-b", FromNodeID: productID, ToNodeID: actionBID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.b")}},
		{Kind: EdgeProduces, EdgeKey: "edge:action-a-produces-output-a", FromNodeID: actionAID, ToNodeID: outputAID, Payload: ProducesPayload{Path: "bin/shared", WriteSlot: "a", WriteClass: "native.executable"}},
		{Kind: EdgeProduces, EdgeKey: "edge:action-b-produces-output-b", FromNodeID: actionBID, ToNodeID: outputBID, Payload: ProducesPayload{Path: "bin/shared", WriteSlot: "b", WriteClass: "native.executable"}},
		{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-output-a", FromNodeID: productID, ToNodeID: outputAID, Payload: PublishesPayload{Destination: "bin/shared", EntryPoint: "duplicate-a"}},
		{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-output-b", FromNodeID: productID, ToNodeID: outputBID, Payload: PublishesPayload{Destination: "bin/shared", EntryPoint: "duplicate-b"}},
	}
	nodes := []Node{product, actionA, actionB, outputA, outputB}
	bundle := fixtureBundle(t, nodes, edges, product)
	plan, leftErr := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if leftErr == nil {
		t.Fatalf("C5 accepted two expected output nodes for the same write path: %#v", plan.DeclaredOutputNodeIDs)
	}
	if ErrorCode(leftErr) != CodeGraphReferenceInvalid {
		t.Fatalf("duplicate write path code = %q, want %q", ErrorCode(leftErr), CodeGraphReferenceInvalid)
	}
	if plan.SchemaID != "" {
		t.Fatalf("invalid write set returned a partial build plan: %#v", plan)
	}
	reverseNodes(nodes)
	reverseEdges(edges)
	permuted := fixtureBundle(t, nodes, edges, product)
	_, rightErr := DeriveBuildPlan(permuted, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if rightErr == nil || leftErr.Error() != rightErr.Error() {
		t.Fatalf("duplicate write-path rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
	}
}

func TestReviewerReworkGeneratedAndExpectedOutputWritePathConflictRejectsBeforeC5(t *testing.T) {
	product := fixtureProduct("generated-output-path-conflict")
	generate := fixtureAction("generated-path-conflict", []string{}, []string{"generated"})
	link := fixtureAction("output-path-conflict", []string{}, []string{"output"})
	generated := Node{Kind: NodeGeneratedArtifact, LogicalKey: "generated:path-conflict", Payload: GeneratedArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "build/shared", Slot: "generated", ExpectedClass: "source.generated_text", Grammar: "fixture-generated-v1", Role: "intermediate", DeclarationDigest: testDigest('7'),
	}}
	output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:path-conflict", Payload: OutputArtifactPayload{
		Profile: "fixture-source-v1", LogicalPath: "build/shared", ExpectedClass: "native.executable", OutputRole: "published_command", DeclarationDigest: testDigest('8'),
	}}
	productID, generateID, linkID := mustNodeID(t, product), mustNodeID(t, generate), mustNodeID(t, link)
	generatedID, outputID := mustNodeID(t, generated), mustNodeID(t, output)
	nodes := []Node{product, generate, link, generated, output}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-generate-conflict", FromNodeID: productID, ToNodeID: generateID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.generate")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-link-conflict", FromNodeID: productID, ToNodeID: linkID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.link")}},
		{Kind: EdgeProduces, EdgeKey: "edge:generate-produces-conflict", FromNodeID: generateID, ToNodeID: generatedID, Payload: ProducesPayload{Path: "build/shared", WriteSlot: "generated", WriteClass: "source.generated_text"}},
		{Kind: EdgeProduces, EdgeKey: "edge:link-produces-conflict", FromNodeID: linkID, ToNodeID: outputID, Payload: ProducesPayload{Path: "build/shared", WriteSlot: "output", WriteClass: "native.executable"}},
		{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-conflict", FromNodeID: productID, ToNodeID: outputID, Payload: PublishesPayload{Destination: "build/shared", EntryPoint: "path-conflict"}},
	}
	leftBundle := fixtureBundle(t, nodes, edges, product)
	_, leftErr := DeriveBuildPlan(leftBundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if leftErr == nil || !strings.Contains(leftErr.Error(), "selected write path \"build/shared\"") {
		t.Fatalf("generated/output write conflict error = %v", leftErr)
	}
	reverseNodes(nodes)
	reverseEdges(edges)
	_, rightErr := DeriveBuildPlan(fixtureBundle(t, nodes, edges, product), PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if rightErr == nil || leftErr.Error() != rightErr.Error() {
		t.Fatalf("generated/output write conflict changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
	}
}

func TestReviewerReworkResolvesToManifestMustResolveAndMatchSource(t *testing.T) {
	product := fixtureProduct("resolves-manifest")
	packageNode := Node{Kind: NodePackageInstance, LogicalKey: "package:resolves-manifest", Payload: PackageInstancePayload{
		Profile: "fixture-source-v1", Ecosystem: "fixture", Manager: "fixture-manager", Origin: "registry://fixture/package/1.0.0",
		LockInstanceKey: "package@1.0.0", Name: "package", Version: "1.0.0", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput,
	}}
	source := Node{Kind: NodeSourceSet, LogicalKey: "source:resolves-manifest", Payload: SourceSetPayload{
		Profile: "fixture-source-v1", Origin: "registry://fixture/package/1.0.0", ArtifactManifestID: testDigest('2'), Projection: []string{}, Grammar: "fixture-source-v1", TrustRole: TrustDependencyInput,
	}}
	productID := mustNodeID(t, product)
	packageID := mustNodeID(t, packageNode)
	sourceID := mustNodeID(t, source)
	edges := []Edge{
		{Kind: EdgeRequires, EdgeKey: "edge:product-requires-package", FromNodeID: productID, ToNodeID: packageID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies.package")}},
		{Kind: EdgeResolvesTo, EdgeKey: "edge:package-resolves-source", FromNodeID: packageID, ToNodeID: sourceID, Payload: ResolvesToPayload{
			LockField: "packages.package", Origin: fixtureOrigin("packages.package.resolved"), Checksum: "sha512-fixture", ArtifactManifestID: testDigest('3'),
		}},
	}
	if _, err := fixtureBundleResult(t, []Node{product, packageNode, source}, edges, product); err == nil {
		t.Fatal("resolves_to accepted an artifact manifest absent from capture and different from both package/source immutable manifests")
	}
}

func TestReviewerReworkResolvesToAcceptsCapturedPackageOrTransformedSourceManifest(t *testing.T) {
	for _, edgeManifest := range []ID{testDigest('1'), testDigest('2')} {
		t.Run(string(edgeManifest), func(t *testing.T) {
			nodes, edges, product := resolutionManifestFixture(t, edgeManifest)
			if _, err := fixtureBundleResult(t, nodes, edges, product); err != nil {
				t.Fatalf("captured endpoint manifest mapping rejected: %v", err)
			}
		})
	}
}

func TestReviewerReworkResolvesToRejectsCapturedUnrelatedManifestCanonically(t *testing.T) {
	unrelated := testDigest('3')
	nodes, edges, product := resolutionManifestFixture(t, unrelated)
	manifestIDs := []ID{testDigest('1'), testDigest('2'), unrelated}
	leftErr := projectResolutionWithManifests(t, nodes, edges, product, manifestIDs)
	if leftErr == nil || !strings.Contains(leftErr.Error(), "matches neither package manifest") {
		t.Fatalf("captured unrelated resolves_to manifest error = %v", leftErr)
	}
	reverseNodes(nodes)
	reverseEdges(edges)
	rightErr := projectResolutionWithManifests(t, nodes, edges, product, manifestIDs)
	if rightErr == nil || leftErr.Error() != rightErr.Error() {
		t.Fatalf("resolves_to mismatch rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
	}
}

func resolutionManifestFixture(t *testing.T, edgeManifest ID) ([]Node, []Edge, Node) {
	t.Helper()
	product := fixtureProduct("resolution-manifest-mapping")
	packageNode := Node{Kind: NodePackageInstance, LogicalKey: "package:resolution-manifest-mapping", Payload: PackageInstancePayload{
		Profile: "fixture-source-v1", Ecosystem: "fixture", Manager: "fixture-manager", Origin: "registry://fixture/resolution/1.0.0", LockInstanceKey: "resolution@1.0.0", Name: "resolution", Version: "1.0.0", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput,
	}}
	source := Node{Kind: NodeSourceSet, LogicalKey: "source:resolution-manifest-mapping", Payload: SourceSetPayload{
		Profile: "fixture-source-v1", Origin: "transform://fixture/resolution/1.0.0", ArtifactManifestID: testDigest('2'), Projection: []string{}, Grammar: "fixture-source-v1", TrustRole: TrustDependencyInput,
	}}
	productID, packageID, sourceID := mustNodeID(t, product), mustNodeID(t, packageNode), mustNodeID(t, source)
	edges := []Edge{
		{Kind: EdgeRequires, EdgeKey: "edge:product-requires-resolution-package", FromNodeID: productID, ToNodeID: packageID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies.resolution")}},
		{Kind: EdgeResolvesTo, EdgeKey: "edge:resolution-package-resolves-source", FromNodeID: packageID, ToNodeID: sourceID, Payload: ResolvesToPayload{LockField: "packages.resolution", Origin: fixtureOrigin("packages.resolution.resolved"), Checksum: "sha512-resolution", ArtifactManifestID: edgeManifest}},
	}
	return []Node{product, packageNode, source}, edges, product
}

func projectResolutionWithManifests(t *testing.T, nodes []Node, edges []Edge, product Node, manifests []ID) error {
	t.Helper()
	baseNodes, baseEdges, baseProduct := resolutionManifestFixture(t, testDigest('2'))
	base := fixtureBundle(t, baseNodes, baseEdges, baseProduct)
	productID := mustNodeID(t, product)
	capture, err := NewCaptureGraph("fixture-source-v1", []string{"fixture-policy-v1"}, []ID{productID}, nodes, edges, manifests)
	if err != nil {
		return err
	}
	captureID, err := capture.ID()
	if err != nil {
		return err
	}
	selectionID, err := base.Selection.ID()
	if err != nil {
		return err
	}
	binding, err := NewSelectionBinding(captureID, selectionID, base.Records.BindingNodes, base.Records.BindingEdges)
	if err != nil {
		return err
	}
	_, err = ProjectActive(capture, base.Selection, binding, NewRecordTables(nodes, edges, base.Records.BindingNodes, base.Records.BindingEdges), base.Authority, nil)
	return err
}

func TestReviewerReworkIntrinsicValidationFailureIsDeterministic(t *testing.T) {
	node := Node{Kind: NodeCommandProduct, LogicalKey: "product:invalid-fields", Payload: CommandProductPayload{
		Profile: "fixture-source-v1", SkillKey: "bad\n", CommandKey: "bad\r", EntryPointContract: "bad\t", DeclarationDigest: testDigest('a'),
	}}
	messages := map[string]bool{}
	for index := 0; index < 1000; index++ {
		err := node.Validate()
		if err == nil {
			t.Fatal("multiply-invalid intrinsic node was accepted")
		}
		messages[err.Error()] = true
	}
	if len(messages) != 1 {
		t.Fatalf("same invalid canonical record produced %d different primary diagnostics across repeated validation: %#v", len(messages), messages)
	}
}

func TestReviewerReworkIntrinsicValidationIsDeterministicAcrossRecordFamilies(t *testing.T) {
	badID := ID("sha256:not-canonical")
	action := fixtureAction("invalid-fields", []string{}, []string{})
	actionPayload := action.Payload.(ActionPayload)
	actionPayload.EnvironmentPolicyID = "bad\n"
	actionPayload.ProcessPolicyID = "bad\r"
	actionPayload.Network = "bad\t"
	action.Payload = actionPayload
	cases := []struct {
		name     string
		validate func() error
	}{
		{name: "command product", validate: func() error {
			return (CommandProductPayload{Profile: "fixture-source-v1", SkillKey: "bad\n", CommandKey: "bad\r", EntryPointContract: "bad\t", DeclarationDigest: testDigest('a')}).validate()
		}},
		{name: "package instance", validate: func() error {
			return (PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "bad\n", Origin: "bad\r", LockInstanceKey: "bad\t", Name: "bad\v", Version: "bad\f", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput}).validate()
		}},
		{name: "target unit", validate: func() error {
			return (TargetUnitPayload{Profile: "fixture-source-v1", TargetName: "bad\n", TargetKind: "bad\r", ExpectedOutputClass: "bad\t", DeclarationDigest: testDigest('1'), Languages: []string{"go"}, ExecutionDomain: ExecutionTarget, ConditionExpressions: []Condition{}}).validate()
		}},
		{name: "action", validate: action.Validate},
		{name: "generated artifact", validate: func() error {
			return (GeneratedArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "gen/out", Slot: "bad\n", ExpectedClass: "bad\r", Grammar: "bad\t", Role: "bad\v", DeclarationDigest: testDigest('1')}).validate()
		}},
		{name: "interop boundary", validate: func() error {
			return (InteropBoundaryPayload{Profile: "fixture-source-v1", Mode: InteropCABI, ProviderLanguageClasses: []string{"c"}, ConsumerLanguageClasses: []string{"swift"}, ContractDigest: testDigest('1'), ABI: "bad\n", CallingConvention: "bad\r", InterfaceContract: "bad\t", LinkLoadSemantics: "bad\v"}).validate()
		}},
		{name: "toolchain component", validate: func() error {
			return (ToolchainComponentPayload{ComponentRole: "bad\n", PlatformABI: "bad\r", PolicySelector: "bad\t", VersionOutput: "bad\v", ContentFingerprint: testDigest('1'), ExecutableRelativePath: "bin/tool"}).validate()
		}},
		{name: "target platform", validate: func() error {
			return (TargetPlatformPayload{OS: "bad\n", Architecture: "bad\r", ABI: "bad\t", Libc: "bad\v", MinimumRuntime: "bad\f", SDKID: "bad\a", TargetTriple: "bad\b"}).validate()
		}},
		{name: "target platform maps", validate: func() error {
			return (TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu", LanguageModes: map[string]string{"z": "bad\n", "a": "bad\r"}, Tuning: map[string]string{"z": "bad\t", "a": "bad\v"}}).validate()
		}},
		{name: "output artifact", validate: func() error {
			return (OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/out", ExpectedClass: "bad\n", OutputRole: "bad\r"}).validate()
		}},
		{name: "provides interop", validate: func() error {
			return (ProvidesInteropPayload{Origin: fixtureOrigin("interop.provider"), EvidenceIDs: []ID{testDigest('1')}, ExportRole: "bad\n", LinkMode: "bad\r"}).validate()
		}},
		{name: "consumes interop", validate: func() error {
			return (ConsumesInteropPayload{Origin: fixtureOrigin("interop.consumer"), Use: "bad\n", ABIExpectation: "bad\r"}).validate()
		}},
		{name: "invokes", validate: func() error {
			return (InvokesPayload{ProtocolSchema: "bad\n", ExecutableResolution: "bad\r", ArgumentsContract: "bad\t", EnvironmentContract: "bad\v"}).validate()
		}},
		{name: "selection maps", validate: func() error {
			return (SelectionContext{SchemaID: SchemaSelectionContext, ProductNodeIDs: []ID{testDigest('1')}, PlatformRoles: map[PlatformRole]ID{PlatformTarget: testDigest('2')}, Features: []string{}, Markers: map[string]string{"z": "bad\n", "a": "bad\r"}, PeerContext: map[string]string{}, EvaluatorIDs: []string{}}).Validate()
		}},
		{name: "active graph IDs", validate: func() error {
			return (ActiveGraph{SchemaID: SchemaActiveGraph, CapturedGraphID: badID, SelectionContextID: badID, SelectionBindingID: badID, NodeActivations: []NodeActivation{}, EdgeActivations: []EdgeActivation{}, NonOrderingSCCs: []NonOrderingSCC{}}).Validate()
		}},
		{name: "diagnostic fields", validate: func() error {
			return (Diagnostic{Code: CodeGraphReferenceInvalid, Subject: "fixture", Fields: map[string]string{"z": "bad\n", "a": "bad\r"}}).validate()
		}},
		{name: "C0 fields", validate: func() error {
			return (C0ProfilePayload{AdapterProfileID: "bad\n", ArtifactPolicyID: "bad\r", DetectorRegistryID: "bad\t", LimitVectorID: "bad\v", ConfigurationPolicyID: "bad\f", SchemaIDs: []string{}, SourceGrammarIDs: []string{}, ManagerSchemaIDs: []string{}, CapabilityIDs: []string{}, SelectionContextID: testDigest('1'), PlatformNodeIDs: []ID{testDigest('2')}, PlatformRoles: map[PlatformRole]ID{PlatformTarget: testDigest('2')}, EvidenceToolchainNodeIDs: []ID{}}).validate()
		}},
		{name: "C1 ID sets", validate: func() error {
			return (C1ResolvePayload{RootDeclarationIDs: []ID{badID}, WorkspaceDeclarationIDs: []ID{badID}, ConditionEdgeIDs: []ID{badID}, CandidateNodeIDs: []ID{badID}, CandidateEdgeIDs: []ID{badID}, JournalEntryIDs: []ID{badID}, LockCandidateID: testDigest('1'), ParserEvaluatorIDs: []string{}, SelectionContextID: testDigest('2')}).validate()
		}},
		{name: "C2 ID sets", validate: func() error {
			return (C2CapturePayload{IntakeReceiptIDs: []ID{badID}, OriginIDs: []ID{badID}, ProtectedHandleIDs: []ID{badID}, BrokerReceiptIDs: []ID{badID}}).validate()
		}},
		{name: "C3 ID sets", validate: func() error {
			return (C3AdmitPayload{Phase: "main", IntakeReceiptIDs: []ID{badID}, ArtifactManifestIDs: []ID{badID}, DerivationReceiptIDs: []ID{badID}}).validate()
		}},
		{name: "C4 IDs", validate: func() error {
			return (C4ClosePayload{ActiveGraphID: badID, CapturedGraphID: badID, SelectionBindingID: badID, SelectionContextID: badID}).validate()
		}},
		{name: "observation IDs", validate: func() error {
			return (ProducedArtifactObservation{Class: "native.executable", ExpectedOutputNodeID: badID, ProducerActionID: badID, ProducesEdgeID: badID, SHA256: badID, Path: "bin/out"}).Validate()
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assertRepeatedErrorStable(t, test.validate)
		})
	}
}

func TestReviewerReworkDecoderFailuresAreCanonicalAcrossMapPermutations(t *testing.T) {
	unknownA := map[string]any{"kind": string(NodeCommandProduct), "logical_key": "product:decode-invalid", "payload": map[string]any{}, "zzz": true, "aaa": true}
	unknownB := map[string]any{"aaa": true, "zzz": true, "payload": map[string]any{}, "logical_key": "product:decode-invalid", "kind": string(NodeCommandProduct)}
	leftUnknown, err := canonicalMapBytes(unknownA)
	if err != nil {
		t.Fatal(err)
	}
	rightUnknown, err := canonicalMapBytes(unknownB)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(leftUnknown, rightUnknown) {
		t.Fatal("canonical map bytes changed with insertion order")
	}
	leftErr := stableDecodeError(t, func() error { _, err := DecodeNode(leftUnknown); return err })
	rightErr := stableDecodeError(t, func() error { _, err := DecodeNode(rightUnknown); return err })
	if leftErr != rightErr || !strings.Contains(leftErr, `unknown field "aaa"`) {
		t.Fatalf("unknown-field decode error is not canonical:\nleft: %s\nright: %s", leftErr, rightErr)
	}

	selectionValue := func(markers map[string]any) map[string]any {
		return map[string]any{
			"default_features": false,
			"evaluator_ids":    []any{},
			"features":         []any{},
			"markers":          markers,
			"peer_context":     map[string]any{},
			"platform_roles":   map[string]any{"target": string(testDigest('2'))},
			"product_node_ids": []any{string(testDigest('1'))},
			"schema_id":        SchemaSelectionContext,
		}
	}
	leftSelection, err := canonicalMapBytes(selectionValue(map[string]any{"z": false, "a": false}))
	if err != nil {
		t.Fatal(err)
	}
	rightSelection, err := canonicalMapBytes(selectionValue(map[string]any{"a": false, "z": false}))
	if err != nil {
		t.Fatal(err)
	}
	leftErr = stableDecodeError(t, func() error { _, err := DecodeSelectionContext(leftSelection); return err })
	rightErr = stableDecodeError(t, func() error { _, err := DecodeSelectionContext(rightSelection); return err })
	if leftErr != rightErr || !strings.Contains(leftErr, `member "a" must be a string`) {
		t.Fatalf("string-map decode error is not canonical:\nleft: %s\nright: %s", leftErr, rightErr)
	}
}

func TestReviewerReworkUnselectedEvaluatorFailureIsCanonical(t *testing.T) {
	input := cgp10Inputs(t)
	evaluator := func(id string) ConditionEvaluator {
		return ConditionEvaluatorFunc{EvaluatorID: id, EvaluateFunc: func(Condition, EvaluationInput) (bool, error) { return false, nil }}
	}
	validate := func(values []ConditionEvaluator) string {
		return stableDecodeError(t, func() error {
			_, err := ProjectActive(input.capture, input.selection, input.binding, input.records, input.authority, values)
			return err
		})
	}
	left := validate([]ConditionEvaluator{evaluator("z-extra-v1"), evaluator("a-extra-v1")})
	right := validate([]ConditionEvaluator{evaluator("a-extra-v1"), evaluator("z-extra-v1")})
	if left != right || !strings.Contains(left, `evaluator "a-extra-v1" is not selected`) {
		t.Fatalf("unselected evaluator failure changed under permutation:\nleft: %s\nright: %s", left, right)
	}
}

func TestReviewerReworkEvaluatorRegistryFailuresAreCanonical(t *testing.T) {
	input := cgp10Inputs(t)
	evaluator := func(id string) ConditionEvaluator {
		return ConditionEvaluatorFunc{EvaluatorID: id, EvaluateFunc: func(Condition, EvaluationInput) (bool, error) { return false, nil }}
	}
	validate := func(values []ConditionEvaluator) string {
		return stableDecodeError(t, func() error {
			_, err := ProjectActive(input.capture, input.selection, input.binding, input.records, input.authority, values)
			return err
		})
	}

	left := validate([]ConditionEvaluator{evaluator("z\n"), evaluator("a\r")})
	right := validate([]ConditionEvaluator{evaluator("a\r"), evaluator("z\n")})
	if left != right || !strings.Contains(left, "evaluator ID") {
		t.Fatalf("invalid evaluator failure changed under permutation:\nleft: %s\nright: %s", left, right)
	}

	left = validate([]ConditionEvaluator{evaluator("z-duplicate-v1"), evaluator("a-duplicate-v1"), evaluator("z-duplicate-v1"), evaluator("a-duplicate-v1")})
	right = validate([]ConditionEvaluator{evaluator("a-duplicate-v1"), evaluator("z-duplicate-v1"), evaluator("a-duplicate-v1"), evaluator("z-duplicate-v1")})
	if left != right || !strings.Contains(left, `duplicate evaluator "a-duplicate-v1"`) {
		t.Fatalf("duplicate evaluator failure changed under permutation:\nleft: %s\nright: %s", left, right)
	}

	left = validate([]ConditionEvaluator{evaluator("a-duplicate-v1"), nil, evaluator("a-duplicate-v1")})
	right = validate([]ConditionEvaluator{nil, evaluator("a-duplicate-v1"), evaluator("a-duplicate-v1")})
	if left != right || !strings.Contains(left, "nil condition evaluator") {
		t.Fatalf("nil evaluator failure changed under permutation:\nleft: %s\nright: %s", left, right)
	}
}

func assertRepeatedErrorStable(t *testing.T, validate func() error) {
	t.Helper()
	stableDecodeError(t, validate)
}

func stableDecodeError(t *testing.T, validate func() error) string {
	t.Helper()
	message := ""
	for index := 0; index < 1000; index++ {
		err := validate()
		if err == nil {
			t.Fatal("multiply-invalid record was accepted")
		}
		if index == 0 {
			message = err.Error()
			continue
		}
		if err.Error() != message {
			t.Fatalf("validation error changed across repeated evaluation:\nfirst: %s\nnext:  %s", message, err)
		}
	}
	return message
}
