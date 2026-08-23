package closuregraph

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func testDigest(character byte) ID { return ID("sha256:" + string(makeBytes(character, 64))) }
func makeBytes(character byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = character
	}
	return result
}

func fixtureProduct(key string) Node {
	return Node{Kind: NodeCommandProduct, LogicalKey: "product:" + key, Payload: CommandProductPayload{Profile: "fixture-source-v1", SkillKey: "fixture", CommandKey: key, EntryPointContract: "native_command", DeclarationDigest: testDigest('a')}}
}
func fixtureAction(key string, reads, writes []string) Node {
	argv := []string{"$TOOL(executor)"}
	for _, slot := range reads {
		argv = append(argv, "$READ("+slot+")")
	}
	for _, slot := range writes {
		argv = append(argv, "$WRITE("+slot+")")
	}
	return Node{Kind: NodeAction, LogicalKey: "action:" + key, Payload: ActionPayload{Profile: "fixture-source-v1", ActionSubtype: "compiler", ExecutionDomain: ExecutionTarget, ArgvTemplate: argv, ToolSlotNames: []string{"executor"}, ReadSlotNames: reads, WriteSlotNames: writes, EnvironmentPolicyID: "env-v1", ProcessPolicyID: "process-v1", Network: "none"}}
}
func fixtureTarget(key, language string) Node {
	return Node{Kind: NodeTargetUnit, LogicalKey: "target:" + key, Payload: TargetUnitPayload{Profile: "fixture-source-v1", TargetName: key, TargetKind: "library", DeclarationDigest: testDigest('b'), Languages: []string{language}, ExecutionDomain: ExecutionTarget, ConditionExpressions: []Condition{}, ExpectedOutputClass: "native.object"}}
}
func fixtureGenerated(key string) Node {
	return Node{Kind: NodeGeneratedArtifact, LogicalKey: "generated:" + key, Payload: GeneratedArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "gen/" + key, Slot: key, ExpectedClass: "source.generated_text", Grammar: "fixture-generated-v1", Role: "intermediate", DeclarationDigest: testDigest('c')}}
}
func fixtureOrigin(field string) EvidenceOrigin {
	return EvidenceOrigin{Field: field, ManifestDigest: testDigest('d')}
}

func fixtureBundle(t *testing.T, nodes []Node, edges []Edge, product Node) GraphBundle {
	t.Helper()
	bundle, err := fixtureBundleResult(t, nodes, edges, product)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func fixtureBundleResult(t *testing.T, nodes []Node, edges []Edge, product Node) (GraphBundle, error) {
	t.Helper()
	productID := mustNodeID(t, product)
	manifests := []ID{}
	for _, node := range nodes {
		switch payload := node.Payload.(type) {
		case PackageInstancePayload:
			if payload.ArtifactManifestID != "" {
				manifests = append(manifests, payload.ArtifactManifestID)
			}
		case SourceSetPayload:
			manifests = append(manifests, payload.ArtifactManifestID)
		}
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i] < manifests[j] })
	manifests = uniqueIDs(manifests)
	capture, err := NewCaptureGraph("fixture-source-v1", []string{"fixture-policy-v1"}, []ID{productID}, nodes, edges, manifests)
	if err != nil {
		return GraphBundle{}, err
	}
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:fixture", Payload: TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "fixture-sdk-v1", TargetTriple: "x86_64-unknown-linux-gnu"}}
	platformID := mustNodeID(t, platform)
	selection, err := NewSelectionContext([]ID{productID}, map[PlatformRole]ID{PlatformTarget: platformID}, []string{}, false, map[string]string{}, map[string]string{}, []string{})
	if err != nil {
		return GraphBundle{}, err
	}
	bindingNodes := []Node{platform}
	bindingEdges := []Edge{}
	authority := BindingAuthority{}
	hasAction := false
	for _, node := range nodes {
		if node.Kind == NodeAction {
			hasAction = true
			break
		}
	}
	var toolchain Node
	var toolchainID ID
	if hasAction {
		toolchain = Node{Kind: NodeToolchainComponent, LogicalKey: "toolchain:fixture-executor", Payload: ToolchainComponentPayload{ComponentRole: "fixture_executor", ContentFingerprint: testDigest('e'), ExecutableRelativePath: "bin/fixture-executor", PlatformABI: "linux-x86_64", PolicySelector: "fixture-toolchain-v1", VersionOutput: "fixture-executor 1.0"}}
		toolchainID = mustNodeID(t, toolchain)
		bindingNodes = append(bindingNodes, toolchain)
		selector, selectorErr := NewToolchainSelector(toolchain)
		if selectorErr != nil {
			return GraphBundle{}, selectorErr
		}
		selectorID, selectorErr := selector.ID()
		if selectorErr != nil {
			return GraphBundle{}, selectorErr
		}
		authority = BindingAuthority{Toolchains: []ToolchainBindingEvidence{{NodeID: toolchainID, FirstBound: ToolchainBoundAtC4, EvidenceID: selectorID}}, C4Selectors: []ToolchainSelector{selector}}
	}
	for _, node := range nodes {
		id := mustNodeID(t, node)
		for _, role := range node.Payload.declaredPlatformRoles() {
			bindingEdges = append(bindingEdges, Edge{Kind: EdgeTargets, EdgeKey: "edge:" + node.LogicalKey + ":targets:" + string(role), FromNodeID: id, ToNodeID: platformID, Payload: TargetsPayload{BindingRole: role, Origin: EvidenceOrigin{Field: "selection.platform_roles." + string(role)}}})
		}
		if node.Kind == NodeAction {
			bindingEdges = append(bindingEdges, Edge{Kind: EdgeUsesTool, EdgeKey: "edge:" + node.LogicalKey + ":uses-fixture-executor", FromNodeID: id, ToNodeID: toolchainID, Payload: UsesToolPayload{ToolSlot: "executor", ExecutableRelativePath: "bin/fixture-executor"}})
		}
		if node.Kind == NodeInteropBoundary && hasAction {
			bindingEdges = append(bindingEdges, Edge{Kind: EdgeRequires, EdgeKey: "edge:" + node.LogicalKey + ":requires-fixture-toolchain", FromNodeID: id, ToNodeID: toolchainID, Payload: RequiresPayload{Scope: ScopeToolchain, Origin: EvidenceOrigin{Field: "selection.interop_toolchains." + node.LogicalKey}}})
		}
	}
	if hasAction {
		bindingEdges = append(bindingEdges, Edge{Kind: EdgeTargets, EdgeKey: "edge:toolchain:fixture-executor:targets:target", FromNodeID: toolchainID, ToNodeID: platformID, Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}}})
	}
	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		return GraphBundle{}, err
	}
	return ProjectActive(capture, selection, binding, NewRecordTables(nodes, edges, bindingNodes, bindingEdges), authority, nil)
}

func TestRuntimeCycleIsRetainedAsNonOrderingSCC(t *testing.T) {
	product := fixtureProduct("runtime-cycle")
	packageA := Node{Kind: NodePackageInstance, LogicalKey: "package:a", Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "node", Manager: "npm", Origin: "registry://a/1.0.0", LockInstanceKey: "a@1.0.0", Name: "a", Version: "1.0.0", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput}}
	packageB := packageA
	packageB.LogicalKey = "package:b"
	payloadB := packageB.Payload.(PackageInstancePayload)
	payloadB.Origin, payloadB.LockInstanceKey, payloadB.Name, payloadB.ArtifactManifestID = "registry://b/1.0.0", "b@1.0.0", "b", testDigest('2')
	packageB.Payload = payloadB
	productID, aID, bID := mustNodeID(t, product), mustNodeID(t, packageA), mustNodeID(t, packageB)
	edges := []Edge{{Kind: EdgeRequires, EdgeKey: "edge:product-requires-a", FromNodeID: productID, ToNodeID: aID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies.a")}}, {Kind: EdgeRequires, EdgeKey: "edge:a-requires-b", FromNodeID: aID, ToNodeID: bID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("a.dependencies.b")}}, {Kind: EdgeRequires, EdgeKey: "edge:b-requires-a", FromNodeID: bID, ToNodeID: aID, Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("b.dependencies.a")}}}
	bundle := fixtureBundle(t, []Node{product, packageA, packageB}, edges, product)
	if len(bundle.Active.NonOrderingSCCs) != 1 {
		t.Fatalf("non-ordering SCCs = %#v", bundle.Active.NonOrderingSCCs)
	}
	wantNodes := sortedIDs([]ID{aID, bID})
	if !reflect.DeepEqual(bundle.Active.NonOrderingSCCs[0].NodeIDs, wantNodes) {
		t.Fatalf("SCC nodes = %v, want %v", bundle.Active.NonOrderingSCCs[0].NodeIDs, wantNodes)
	}
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.OrderingEdges) != 0 || len(plan.Waves) != 0 {
		t.Fatalf("runtime SCC invented build order: %#v", plan)
	}
}

func TestOneLanguageNeutralModelRepresentsEveryAcceptedEcosystem(t *testing.T) {
	tests := []struct {
		name, profile, ecosystem, manager string
	}{
		{name: "Go", profile: "go-v1", ecosystem: "go", manager: "go-modules"},
		{name: "Rust", profile: "rust-source-v1", ecosystem: "rust", manager: "cargo"},
		{name: "Node TypeScript", profile: "node-source-v1", ecosystem: "node", manager: "npm"},
		{name: "Python reference", profile: "python-reference-v1", ecosystem: "python", manager: "reference-lock"},
		{name: "SwiftPM C family", profile: "swiftpm-source-v1", ecosystem: "swift", manager: "swiftpm"},
	}
	for index, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			product := fixtureProduct("ecosystem-" + testCase.ecosystem)
			manifestID := testDigest(byte('1' + index))
			instance := Node{Kind: NodePackageInstance, LogicalKey: "package:" + testCase.ecosystem, Payload: PackageInstancePayload{Profile: testCase.profile, Ecosystem: testCase.ecosystem, Manager: testCase.manager, NormalizedSourceID: testCase.ecosystem + ".fixture", Origin: "registry://" + testCase.ecosystem + "/fixture/1.0.0", LockInstanceKey: testCase.ecosystem + "@1.0.0", Name: "fixture", Version: "1.0.0", ArtifactManifestID: manifestID, TrustRole: TrustDependencyInput}}
			edge := Edge{Kind: EdgeRequires, EdgeKey: "edge:product-requires-" + testCase.ecosystem, FromNodeID: mustNodeID(t, product), ToNodeID: mustNodeID(t, instance), Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies." + testCase.ecosystem)}}
			bundle := fixtureBundle(t, []Node{product, instance}, []Edge{edge}, product)
			if err := bundle.Validate(); err != nil {
				t.Fatal(err)
			}
			if len(bundle.Active.NodeActivations) != 2 || bundle.Active.NodeActivations[0].State != ActivationSelected || bundle.Active.NodeActivations[1].State != ActivationSelected {
				t.Fatalf("active projection = %#v", bundle.Active.NodeActivations)
			}
		})
	}
}

func TestMixedLanguageInteropProducesProviderBeforeConsumerWave(t *testing.T) {
	bundle, providerActionID, consumerActionID := interopFixture(t, InteropCABI)
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.OrderingEdges) != 1 || plan.OrderingEdges[0].Reason != OrderInterop || plan.OrderingEdges[0].FromActionID != providerActionID || plan.OrderingEdges[0].ToActionID != consumerActionID {
		t.Fatalf("interop ordering = %#v", plan.OrderingEdges)
	}
	want := [][]ID{{providerActionID}, {consumerActionID}}
	if !reflect.DeepEqual(plan.Waves, want) {
		t.Fatalf("waves = %v, want %v", plan.Waves, want)
	}
}

func TestRuntimeSubprocessBoundaryDoesNotInventBuildOrder(t *testing.T) {
	bundle, providerActionID, consumerActionID := interopFixture(t, InteropSubprocessProtocol)
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	wave := sortedIDs([]ID{providerActionID, consumerActionID})
	if len(plan.OrderingEdges) != 0 || !reflect.DeepEqual(plan.Waves, [][]ID{wave}) {
		t.Fatalf("subprocess plan = %#v", plan)
	}
}

func TestSelectedLocalArtifactsRequireExactlyOneProducer(t *testing.T) {
	t.Run("missing expected output producer", func(t *testing.T) {
		product := fixtureProduct("missing-output-producer")
		output := Node{Kind: NodeOutputArtifact, LogicalKey: "output:missing-producer", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/missing", ExpectedClass: "native.executable", OutputRole: "published_command"}}
		productID, outputID := mustNodeID(t, product), mustNodeID(t, output)
		edges := []Edge{{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-missing", FromNodeID: productID, ToNodeID: outputID, Payload: PublishesPayload{Destination: "bin/missing", EntryPoint: "missing-output-producer"}}}
		_, err := fixtureBundleResult(t, []Node{product, output}, edges, product)
		if err == nil || !strings.Contains(err.Error(), string(CodeGeneratedInputUndeclared)) || !strings.Contains(err.Error(), "exactly one producer, got 0") {
			t.Fatalf("error = %v, want missing-producer rejection", err)
		}
	})

	t.Run("duplicate generated artifact producers", func(t *testing.T) {
		product := fixtureProduct("duplicate-generated-producer")
		left := fixtureAction("duplicate-producer-left", []string{}, []string{"generated"})
		right := fixtureAction("duplicate-producer-right", []string{}, []string{"generated"})
		generated := fixtureGenerated("duplicate-producer.go")
		productID := mustNodeID(t, product)
		leftID, rightID, generatedID := mustNodeID(t, left), mustNodeID(t, right), mustNodeID(t, generated)
		edges := []Edge{
			{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-left-producer", FromNodeID: productID, ToNodeID: leftID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.left")}},
			{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-right-producer", FromNodeID: productID, ToNodeID: rightID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.right")}},
			{Kind: EdgeProduces, EdgeKey: "edge:left-produces-generated", FromNodeID: leftID, ToNodeID: generatedID, Payload: ProducesPayload{Path: "gen/duplicate-producer.go", WriteSlot: "generated", WriteClass: "source.generated_text"}},
			{Kind: EdgeProduces, EdgeKey: "edge:right-produces-generated", FromNodeID: rightID, ToNodeID: generatedID, Payload: ProducesPayload{Path: "gen/duplicate-producer.go", WriteSlot: "generated", WriteClass: "source.generated_text"}},
		}
		_, err := fixtureBundleResult(t, []Node{product, left, right, generated}, edges, product)
		if err == nil || !strings.Contains(err.Error(), string(CodeGraphReferenceInvalid)) || !strings.Contains(err.Error(), "exactly one producer, got 2") {
			t.Fatalf("error = %v, want duplicate-producer rejection", err)
		}
		nodes := []Node{product, left, right, generated}
		reverseNodes(nodes)
		reverseEdges(edges)
		_, permutedErr := fixtureBundleResult(t, nodes, edges, product)
		if permutedErr == nil || err.Error() != permutedErr.Error() {
			t.Fatalf("duplicate-producer rejection changed under permutation:\nleft: %v\nright: %v", err, permutedErr)
		}
	})
}

func TestTargetLevelGeneratedReadsOrderProviderBeforeEveryOwnedConsumer(t *testing.T) {
	product := fixtureProduct("target-read-order")
	provider := fixtureAction("generate-target-input", []string{}, []string{"generated"})
	generated := fixtureGenerated("target-input")
	target := fixtureTarget("consumer", "swift")
	consumer := fixtureAction("compile-consumer", []string{}, []string{})
	productID := mustNodeID(t, product)
	providerID, generatedID := mustNodeID(t, provider), mustNodeID(t, generated)
	targetID, consumerID := mustNodeID(t, target), mustNodeID(t, consumer)
	declareProvider := Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-provider", FromNodeID: productID, ToNodeID: providerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.generate")}}
	declareTarget := Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-target", FromNodeID: productID, ToNodeID: targetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}}
	declareConsumer := Edge{Kind: EdgeDeclares, EdgeKey: "edge:target-declares-consumer", FromNodeID: targetID, ToNodeID: consumerID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}}
	produces := Edge{Kind: EdgeProduces, EdgeKey: "edge:provider-produces-generated", FromNodeID: providerID, ToNodeID: generatedID, Payload: ProducesPayload{Path: "gen/target-input", WriteSlot: "generated", WriteClass: "source.generated_text"}}
	reads := Edge{Kind: EdgeReads, EdgeKey: "edge:target-reads-generated", FromNodeID: targetID, ToNodeID: generatedID, Payload: ReadsPayload{Path: "gen/target-input", ReadSlot: "sources", ReadClass: "source.generated_text"}}
	nodes := []Node{product, provider, generated, target, consumer}
	edges := []Edge{declareProvider, declareTarget, declareConsumer, produces, reads}

	derive := func(nodes []Node, edges []Edge) BuildPlan {
		bundle := fixtureBundle(t, nodes, edges, product)
		plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
		if err != nil {
			t.Fatal(err)
		}
		return plan
	}
	plan := derive(nodes, edges)
	if len(plan.OrderingEdges) != 1 {
		t.Fatalf("ordering edges = %#v, want one generated-input arc", plan.OrderingEdges)
	}
	ordering := plan.OrderingEdges[0]
	if ordering.Reason != OrderGeneratedInput || ordering.FromActionID != providerID || ordering.ToActionID != consumerID {
		t.Fatalf("target-level generated read ordering = %#v", ordering)
	}
	wantEvidence := sortedIDs([]ID{mustEdgeID(t, declareConsumer), mustEdgeID(t, produces), mustEdgeID(t, reads)})
	if !reflect.DeepEqual(ordering.SourceEdgeIDs, wantEvidence) {
		t.Fatalf("ordering evidence = %v, want %v", ordering.SourceEdgeIDs, wantEvidence)
	}
	if !reflect.DeepEqual(plan.Waves, [][]ID{{providerID}, {consumerID}}) {
		t.Fatalf("waves = %v, want provider then consumer", plan.Waves)
	}

	reversedNodes := append([]Node{}, nodes...)
	reversedEdges := append([]Edge{}, edges...)
	reverseNodes(reversedNodes)
	reverseEdges(reversedEdges)
	permuted := derive(reversedNodes, reversedEdges)
	leftID, err := plan.ID()
	if err != nil {
		t.Fatal(err)
	}
	rightID, err := permuted.ID()
	if err != nil {
		t.Fatal(err)
	}
	if leftID != rightID || !reflect.DeepEqual(plan, permuted) {
		t.Fatalf("target-level read plan changed under permutation: %s/%s", leftID, rightID)
	}
}

func TestTargetLevelGeneratedReadWithoutConsumerActionRejectsDeterministically(t *testing.T) {
	product := fixtureProduct("target-read-without-consumer")
	provider := fixtureAction("generate-unconsumed-target-input", []string{}, []string{"generated"})
	generated := fixtureGenerated("unconsumed-target-input")
	target := fixtureTarget("consumer-without-action", "swift")
	productID, providerID := mustNodeID(t, product), mustNodeID(t, provider)
	generatedID, targetID := mustNodeID(t, generated), mustNodeID(t, target)
	nodes := []Node{product, provider, generated, target}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-unconsumed-provider", FromNodeID: productID, ToNodeID: providerID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.generate")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-consumer-without-action", FromNodeID: productID, ToNodeID: targetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeProduces, EdgeKey: "edge:provider-produces-unconsumed-generated", FromNodeID: providerID, ToNodeID: generatedID, Payload: ProducesPayload{Path: "gen/unconsumed-target-input", WriteSlot: "generated", WriteClass: "source.generated_text"}},
		{Kind: EdgeReads, EdgeKey: "edge:target-without-action-reads-generated", FromNodeID: targetID, ToNodeID: generatedID, Payload: ReadsPayload{Path: "gen/unconsumed-target-input", ReadSlot: "sources", ReadClass: "source.generated_text"}},
	}
	deriveError := func(nodes []Node, edges []Edge) error {
		bundle := fixtureBundle(t, nodes, edges, product)
		_, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
		return err
	}
	leftErr := deriveError(nodes, edges)
	if leftErr == nil || !strings.Contains(leftErr.Error(), "has no selected consumer action") {
		t.Fatalf("error = %v, want missing target consumer rejection", leftErr)
	}
	reverseNodes(nodes)
	reverseEdges(edges)
	rightErr := deriveError(nodes, edges)
	if rightErr == nil || leftErr.Error() != rightErr.Error() {
		t.Fatalf("target-read rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
	}
}

func TestSelectedTargetDependenciesOrderActualProviderActions(t *testing.T) {
	for _, scope := range []RequirementScope{ScopeDevelopment, ScopeOptional, ScopeWorkspace} {
		t.Run(string(scope), func(t *testing.T) {
			product := fixtureProduct("target-order-" + string(scope))
			providerTarget := fixtureTarget("provider-"+string(scope), "c")
			consumerTarget := fixtureTarget("consumer-"+string(scope), "swift")
			providerAction := fixtureAction("provider-"+string(scope), []string{}, []string{})
			consumerAction := fixtureAction("consumer-"+string(scope), []string{}, []string{})
			productID := mustNodeID(t, product)
			providerTargetID, consumerTargetID := mustNodeID(t, providerTarget), mustNodeID(t, consumerTarget)
			providerActionID, consumerActionID := mustNodeID(t, providerAction), mustNodeID(t, consumerAction)
			edges := []Edge{
				{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-provider-" + string(scope), FromNodeID: productID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider")}},
				{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-consumer-" + string(scope), FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
				{Kind: EdgeDeclares, EdgeKey: "edge:provider-declares-action-" + string(scope), FromNodeID: providerTargetID, ToNodeID: providerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider.action")}},
				{Kind: EdgeDeclares, EdgeKey: "edge:consumer-declares-action-" + string(scope), FromNodeID: consumerTargetID, ToNodeID: consumerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
				{Kind: EdgeRequires, EdgeKey: "edge:consumer-requires-provider-" + string(scope), FromNodeID: consumerTargetID, ToNodeID: providerTargetID, Payload: RequiresPayload{Scope: scope, Origin: fixtureOrigin("targets.consumer.dependencies.provider")}},
			}
			bundle := fixtureBundle(t, []Node{product, providerTarget, consumerTarget, providerAction, consumerAction}, edges, product)
			plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.OrderingEdges) != 1 || plan.OrderingEdges[0].Reason != OrderTargetRequirement || plan.OrderingEdges[0].FromActionID != providerActionID || plan.OrderingEdges[0].ToActionID != consumerActionID {
				t.Fatalf("scope %s ordering = %#v", scope, plan.OrderingEdges)
			}
			if !reflect.DeepEqual(plan.Waves, [][]ID{{providerActionID}, {consumerActionID}}) {
				t.Fatalf("scope %s waves = %v", scope, plan.Waves)
			}
		})
	}
}

func TestSelectedOptionalPackageDependencyOrdersMaterializationProvider(t *testing.T) {
	product := fixtureProduct("optional-package-order")
	consumerTarget := fixtureTarget("package-consumer", "typescript")
	providerTarget := fixtureTarget("package-provider", "typescript")
	consumerAction := fixtureAction("package-consumer", []string{}, []string{})
	providerAction := fixtureAction("package-provider", []string{}, []string{})
	providerPackage := Node{Kind: NodePackageInstance, LogicalKey: "package:provider", Payload: PackageInstancePayload{Profile: "fixture-source-v1", Ecosystem: "node", Manager: "npm", Origin: "registry://provider/1.0.0", LockInstanceKey: "provider@1.0.0", Name: "provider", Version: "1.0.0", ArtifactManifestID: testDigest('1'), TrustRole: TrustDependencyInput}}
	productID, packageID := mustNodeID(t, product), mustNodeID(t, providerPackage)
	consumerTargetID, providerTargetID := mustNodeID(t, consumerTarget), mustNodeID(t, providerTarget)
	consumerActionID, providerActionID := mustNodeID(t, consumerAction), mustNodeID(t, providerAction)
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-consumer-target", FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:consumer-target-declares-action", FromNodeID: consumerTargetID, ToNodeID: consumerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
		{Kind: EdgeRequires, EdgeKey: "edge:product-requires-optional-package", FromNodeID: productID, ToNodeID: packageID, Payload: RequiresPayload{Scope: ScopeOptional, Origin: fixtureOrigin("optionalDependencies.provider")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:package-declares-provider-target", FromNodeID: packageID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("provider.targets.library")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:provider-target-declares-action", FromNodeID: providerTargetID, ToNodeID: providerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("provider.targets.library.action")}},
	}
	bundle := fixtureBundle(t, []Node{product, consumerTarget, providerPackage, providerTarget, consumerAction, providerAction}, edges, product)
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.OrderingEdges) != 1 || plan.OrderingEdges[0].FromActionID != providerActionID || plan.OrderingEdges[0].ToActionID != consumerActionID || plan.OrderingEdges[0].Reason != OrderTargetRequirement {
		t.Fatalf("optional package ordering = %#v", plan.OrderingEdges)
	}
}

func interopFixture(t *testing.T, mode InteropMode) (GraphBundle, ID, ID) {
	t.Helper()
	if mode == InteropSubprocessProtocol {
		nodes, edges, product, providerActionID, consumerActionID := subprocessInteropRecords(t)
		return fixtureBundle(t, nodes, edges, product), providerActionID, consumerActionID
	}
	product := fixtureProduct("mixed-cli")
	providerTarget := fixtureTarget("c-provider", "c")
	consumerTarget := fixtureTarget("swift-consumer", "swift")
	providerAction := fixtureAction("compile-c", []string{}, []string{})
	consumerAction := fixtureAction("compile-swift", []string{}, []string{})
	boundaryPayload := InteropBoundaryPayload{Profile: "fixture-source-v1", Mode: mode, ProviderLanguageClasses: []string{"c"}, ConsumerLanguageClasses: []string{"swift"}, ContractDigest: testDigest('e'), ABI: "c-abi-v1", InterfaceContract: "fixture-header-v1", CallingConvention: "c", LinkLoadSemantics: "static-link"}
	boundary := Node{Kind: NodeInteropBoundary, LogicalKey: "interop:fixture", Payload: boundaryPayload}
	productID, providerTargetID, consumerTargetID, providerActionID, consumerActionID, boundaryID := mustNodeID(t, product), mustNodeID(t, providerTarget), mustNodeID(t, consumerTarget), mustNodeID(t, providerAction), mustNodeID(t, consumerAction), mustNodeID(t, boundary)
	edges := []Edge{{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-provider", FromNodeID: productID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider")}}, {Kind: EdgeDeclares, EdgeKey: "edge:product-declares-consumer", FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}}, {Kind: EdgeDeclares, EdgeKey: "edge:provider-declares-action", FromNodeID: providerTargetID, ToNodeID: providerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider.action")}}, {Kind: EdgeDeclares, EdgeKey: "edge:consumer-declares-action", FromNodeID: consumerTargetID, ToNodeID: consumerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}}, {Kind: EdgeProvidesInterop, EdgeKey: "edge:provider-provides-boundary", FromNodeID: providerTargetID, ToNodeID: boundaryID, Payload: ProvidesInteropPayload{Origin: fixtureOrigin("targets.provider.headers"), EvidenceIDs: []ID{testDigest('f')}, ExportRole: "headers", LinkMode: "static"}}, {Kind: EdgeConsumesInterop, EdgeKey: "edge:consumer-consumes-boundary", FromNodeID: consumerActionID, ToNodeID: boundaryID, Payload: ConsumesInteropPayload{Origin: fixtureOrigin("targets.consumer.import"), Use: "compile", ABIExpectation: "c-abi-v1"}}}
	return fixtureBundle(t, []Node{product, providerTarget, consumerTarget, providerAction, consumerAction, boundary}, edges, product), providerActionID, consumerActionID
}

func subprocessInteropRecords(t *testing.T) ([]Node, []Edge, Node, ID, ID) {
	t.Helper()
	product := fixtureProduct("subprocess-cli")
	providerTarget := fixtureTarget("subprocess-provider", "go")
	consumerTarget := fixtureTarget("subprocess-consumer", "swift")
	providerAction := fixtureAction("build-subprocess-provider", []string{}, []string{"provider"})
	consumerAction := fixtureAction("build-subprocess-consumer", []string{}, []string{"consumer"})
	providerOutput := Node{Kind: NodeOutputArtifact, LogicalKey: "output:subprocess-provider", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/subprocess-provider", ExpectedClass: "native.executable", OutputRole: "published_command"}}
	consumerOutput := Node{Kind: NodeOutputArtifact, LogicalKey: "output:subprocess-consumer", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/subprocess-consumer", ExpectedClass: "native.executable", OutputRole: "published_command"}}
	boundary := roleInteropFixture("subprocess", InteropSubprocessProtocol)
	productID := mustNodeID(t, product)
	providerTargetID, consumerTargetID := mustNodeID(t, providerTarget), mustNodeID(t, consumerTarget)
	providerActionID, consumerActionID := mustNodeID(t, providerAction), mustNodeID(t, consumerAction)
	providerOutputID, consumerOutputID, boundaryID := mustNodeID(t, providerOutput), mustNodeID(t, consumerOutput), mustNodeID(t, boundary)
	nodes := []Node{product, providerTarget, consumerTarget, providerAction, consumerAction, providerOutput, consumerOutput, boundary}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-subprocess-provider", FromNodeID: productID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-subprocess-consumer", FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:subprocess-provider-declares-action", FromNodeID: providerTargetID, ToNodeID: providerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider.action")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:subprocess-consumer-declares-action", FromNodeID: consumerTargetID, ToNodeID: consumerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
		{Kind: EdgeProduces, EdgeKey: "edge:subprocess-provider-produces-output", FromNodeID: providerActionID, ToNodeID: providerOutputID, Payload: ProducesPayload{Path: "bin/subprocess-provider", WriteSlot: "provider", WriteClass: "native.executable"}},
		{Kind: EdgeProduces, EdgeKey: "edge:subprocess-consumer-produces-output", FromNodeID: consumerActionID, ToNodeID: consumerOutputID, Payload: ProducesPayload{Path: "bin/subprocess-consumer", WriteSlot: "consumer", WriteClass: "native.executable"}},
		{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-subprocess-provider", FromNodeID: productID, ToNodeID: providerOutputID, Payload: PublishesPayload{Destination: "bin/subprocess-provider", EntryPoint: "subprocess-provider"}},
		{Kind: EdgePublishes, EdgeKey: "edge:product-publishes-subprocess-consumer", FromNodeID: productID, ToNodeID: consumerOutputID, Payload: PublishesPayload{Destination: "bin/subprocess-consumer", EntryPoint: "subprocess-consumer"}},
		{Kind: EdgeProvidesInterop, EdgeKey: "edge:subprocess-provider-provides-boundary", FromNodeID: providerOutputID, ToNodeID: boundaryID, Payload: ProvidesInteropPayload{Origin: fixtureOrigin("outputs.provider.protocol"), EvidenceIDs: []ID{testDigest('f')}, ExportRole: "command", LinkMode: "runtime"}},
		{Kind: EdgeConsumesInterop, EdgeKey: "edge:subprocess-consumer-consumes-boundary", FromNodeID: consumerOutputID, ToNodeID: boundaryID, Payload: ConsumesInteropPayload{Origin: fixtureOrigin("outputs.consumer.protocol"), Use: "invoke", ABIExpectation: "protocol-v1"}},
		{Kind: EdgeInvokes, EdgeKey: "edge:subprocess-consumer-invokes-provider", FromNodeID: consumerOutputID, ToNodeID: providerOutputID, Payload: InvokesPayload{ProtocolSchema: "fixture-json-v1", ExecutableResolution: "bundle-relative-v1", ArgumentsContract: "argv-v1", EnvironmentContract: "env-v1", WorkingDirectory: "work"}},
	}
	return nodes, edges, product, providerActionID, consumerActionID
}

func TestBuildCycleRejectionIsDeterministicAcrossPermutations(t *testing.T) {
	nodes, edges, product := cyclicActionFixture(t)
	original := fixtureBundle(t, nodes, edges, product)
	reversedNodes := append([]Node{}, nodes...)
	reversedEdges := append([]Edge{}, edges...)
	reverseNodes(reversedNodes)
	reverseEdges(reversedEdges)
	permuted := fixtureBundle(t, reversedNodes, reversedEdges, product)
	getCycle := func(bundle GraphBundle) *BuildCycleError {
		_, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1", LastCheckpointID: testDigest('9')})
		var cycle *BuildCycleError
		if !errors.As(err, &cycle) {
			t.Fatalf("cycle error = %T %v", err, err)
		}
		return cycle
	}
	left, right := getCycle(original), getCycle(permuted)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("cycle changed under permutation\nleft %#v\nright %#v", left, right)
	}
	if len(left.ActionNodeIDs) != 2 || len(left.OrderingEdgeIDs) != 2 || !left.CycleDigest.Valid() {
		t.Fatalf("incomplete cycle evidence %#v", left)
	}
}

func TestBuildCycleAffectedScopeExcludesSharedPlatformHubs(t *testing.T) {
	nodes, edges, product := cyclicActionFixture(t)
	cycleTarget := fixtureTarget("cycle-owner", "go")
	unrelatedTarget := fixtureTarget("unrelated", "rust")
	unrelatedAction := fixtureAction("unrelated", []string{}, []string{})
	cycleTargetID, unrelatedTargetID, unrelatedActionID := mustNodeID(t, cycleTarget), mustNodeID(t, unrelatedTarget), mustNodeID(t, unrelatedAction)
	productID := mustNodeID(t, product)
	actionIDs := []ID{mustNodeID(t, nodes[1]), mustNodeID(t, nodes[2])}
	ownedEdges := make([]Edge, 0, len(edges)+4)
	for _, edge := range edges {
		if edge.Kind != EdgeDeclares {
			ownedEdges = append(ownedEdges, edge)
		}
	}
	ownedEdges = append(ownedEdges,
		Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-cycle-target", FromNodeID: productID, ToNodeID: cycleTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.cycle")}},
		Edge{Kind: EdgeDeclares, EdgeKey: "edge:cycle-target-declares-a", FromNodeID: cycleTargetID, ToNodeID: actionIDs[0], Payload: DeclaresPayload{Origin: fixtureOrigin("targets.cycle.actions.a")}},
		Edge{Kind: EdgeDeclares, EdgeKey: "edge:cycle-target-declares-b", FromNodeID: cycleTargetID, ToNodeID: actionIDs[1], Payload: DeclaresPayload{Origin: fixtureOrigin("targets.cycle.actions.b")}},
		Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-unrelated", FromNodeID: productID, ToNodeID: unrelatedTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.unrelated")}},
		Edge{Kind: EdgeDeclares, EdgeKey: "edge:unrelated-declares-action", FromNodeID: unrelatedTargetID, ToNodeID: unrelatedActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.unrelated.action")}},
	)
	nodes = append(nodes, cycleTarget, unrelatedTarget, unrelatedAction)
	bundle := fixtureBundle(t, nodes, ownedEdges, product)
	_, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1", LastCheckpointID: testDigest('9')})
	var cycle *BuildCycleError
	if !errors.As(err, &cycle) {
		t.Fatalf("cycle error = %T %v", err, err)
	}
	if !reflect.DeepEqual(cycle.AffectedProductIDs, []ID{productID}) {
		t.Fatalf("affected products = %v, want %s", cycle.AffectedProductIDs, productID)
	}
	if !reflect.DeepEqual(cycle.AffectedTargetIDs, []ID{cycleTargetID}) {
		t.Fatalf("affected targets = %v, want only %s (not platform-connected %s)", cycle.AffectedTargetIDs, cycleTargetID, unrelatedTargetID)
	}
}

func TestGraphAndPlanIDsArePermutationStable(t *testing.T) {
	bundle, _, _ := interopFixture(t, InteropCABI)
	nodes := append([]Node{}, bundle.Records.CaptureNodes...)
	edges := append([]Edge{}, bundle.Records.CaptureEdges...)
	reverseNodes(nodes)
	reverseEdges(edges)
	product := findNodeByKind(t, nodes, NodeCommandProduct)
	permuted := fixtureBundle(t, nodes, edges, product)
	leftCapture, _ := bundle.Capture.ID()
	rightCapture, _ := permuted.Capture.ID()
	leftActive, _ := bundle.Active.ID()
	rightActive, _ := permuted.Active.ID()
	if leftCapture != rightCapture || leftActive != rightActive {
		t.Fatalf("graph IDs changed: capture %s/%s active %s/%s", leftCapture, rightCapture, leftActive, rightActive)
	}
	leftPlan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	rightPlan, err := DeriveBuildPlan(permuted, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	leftID, _ := leftPlan.ID()
	rightID, _ := rightPlan.ID()
	if leftID != rightID {
		t.Fatalf("plan IDs changed: %s/%s", leftID, rightID)
	}
}

func cyclicActionFixture(t *testing.T) ([]Node, []Edge, Node) {
	t.Helper()
	product := fixtureProduct("cycle-cli")
	actionA := fixtureAction("generate-a", []string{"input"}, []string{"output"})
	actionB := fixtureAction("generate-b", []string{"input"}, []string{"output"})
	generatedA, generatedB := fixtureGenerated("a.c"), fixtureGenerated("b.c")
	productID, actionAID, actionBID, generatedAID, generatedBID := mustNodeID(t, product), mustNodeID(t, actionA), mustNodeID(t, actionB), mustNodeID(t, generatedA), mustNodeID(t, generatedB)
	edges := []Edge{{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-a", FromNodeID: productID, ToNodeID: actionAID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.a")}}, {Kind: EdgeDeclares, EdgeKey: "edge:product-declares-b", FromNodeID: productID, ToNodeID: actionBID, Payload: DeclaresPayload{Origin: fixtureOrigin("actions.b")}}, {Kind: EdgeProduces, EdgeKey: "edge:a-produces-a", FromNodeID: actionAID, ToNodeID: generatedAID, Payload: ProducesPayload{Path: "gen/a.c", WriteSlot: "output"}}, {Kind: EdgeReads, EdgeKey: "edge:a-reads-b", FromNodeID: actionAID, ToNodeID: generatedBID, Payload: ReadsPayload{Path: "gen/b.c", ReadSlot: "input"}}, {Kind: EdgeProduces, EdgeKey: "edge:b-produces-b", FromNodeID: actionBID, ToNodeID: generatedBID, Payload: ProducesPayload{Path: "gen/b.c", WriteSlot: "output"}}, {Kind: EdgeReads, EdgeKey: "edge:b-reads-a", FromNodeID: actionBID, ToNodeID: generatedAID, Payload: ReadsPayload{Path: "gen/a.c", ReadSlot: "input"}}}
	return []Node{product, actionA, actionB, generatedA, generatedB}, edges, product
}

func uniqueIDs(values []ID) []ID {
	result := []ID{}
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
func reverseNodes(values []Node) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseEdges(values []Edge) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func findNodeByKind(t *testing.T, nodes []Node, kind NodeKind) Node {
	t.Helper()
	for _, node := range nodes {
		if node.Kind == kind {
			return node
		}
	}
	t.Fatalf("missing node kind %s", kind)
	return Node{}
}
