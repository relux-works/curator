package closuregraph

import (
	"strings"
	"testing"
)

func TestReviewerProbeTargetRequiresProducedOutputOrdersConsumer(t *testing.T) {
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

func TestReviewerProbeInteropRejectsNonToolchainCaptureRequirement(t *testing.T) {
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

func TestReviewerProbeDuplicateExpectedOutputPathsRejectBeforeC5(t *testing.T) {
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
	bundle := fixtureBundle(t, []Node{product, actionA, actionB, outputA, outputB}, edges, product)
	if plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"}); err == nil {
		t.Fatalf("C5 accepted two expected output nodes for the same write path: %#v", plan.DeclaredOutputNodeIDs)
	}
}

func TestReviewerProbeResolvesToManifestMustResolveAndMatchSource(t *testing.T) {
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

func TestReviewerProbeIntrinsicValidationFailureIsDeterministic(t *testing.T) {
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
