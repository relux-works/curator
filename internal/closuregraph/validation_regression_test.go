package closuregraph

import (
	"reflect"
	"strings"
	"testing"
)

func TestPlatformBearingNodesRejectExplicitEmptyOrIncompleteRoles(t *testing.T) {
	cases := []struct {
		name     string
		node     Node
		required PlatformRole
	}{
		{name: "command product", node: fixtureProduct("role-product"), required: PlatformTarget},
		{name: "target unit", node: fixtureTarget("role-target", "go"), required: PlatformTarget},
		{name: "target unit host", node: hostTargetFixture("role-host-target"), required: PlatformHost},
		{name: "action", node: fixtureAction("role-action", []string{}, []string{}), required: PlatformTarget},
		{name: "action host", node: hostActionFixture("role-host-action"), required: PlatformHost},
		{name: "toolchain", node: roleToolchainFixture("role-toolchain", ExecutionTarget), required: PlatformTarget},
		{name: "toolchain host", node: roleToolchainFixture("role-host-toolchain", ExecutionHost), required: PlatformHost},
		{name: "interop boundary", node: roleInteropFixture("role-interop", InteropCABI), required: PlatformTarget},
		{name: "interop host extension", node: roleInteropFixture("role-host-interop", InteropHostExtension), required: PlatformHost},
		{name: "output", node: Node{Kind: NodeOutputArtifact, LogicalKey: "output:role", Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "bin/role", ExpectedClass: "native.executable", OutputRole: "published_command"}}, required: PlatformTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.node.Validate(); err != nil {
				t.Fatalf("default role declaration must be valid: %v", err)
			}
			empty := nodeWithPlatformRoles(testCase.node, []PlatformRole{})
			if err := empty.Validate(); err == nil || !strings.Contains(err.Error(), "must not be explicitly empty") {
				t.Fatalf("explicit-empty roles error = %v", err)
			}
			wrong := PlatformHost
			if testCase.required == PlatformHost {
				wrong = PlatformTarget
			}
			incomplete := nodeWithPlatformRoles(testCase.node, []PlatformRole{wrong})
			if err := incomplete.Validate(); err == nil || !strings.Contains(err.Error(), "must include required platform role") {
				t.Fatalf("incomplete roles error = %v", err)
			}
			complete := nodeWithPlatformRoles(testCase.node, []PlatformRole{testCase.required})
			if err := complete.Validate(); err != nil {
				t.Fatalf("explicit required role rejected: %v", err)
			}
		})
	}
}

func TestEveryInteropModeRequiresItsIntrinsicContractEvidence(t *testing.T) {
	required := map[InteropMode][]string{
		InteropCABI:               {"abi", "calling_convention", "interface_contract", "link_load_semantics"},
		InteropCXX:                {"abi", "calling_convention", "interface_contract", "link_load_semantics", "runtime"},
		InteropObjCRuntime:        {"calling_convention", "interface_contract", "link_load_semantics", "runtime"},
		InteropNativeLink:         {"abi", "interface_contract", "link_load_semantics"},
		InteropDynamicLoad:        {"abi", "interface_contract", "link_load_semantics", "runtime"},
		InteropHostExtension:      {"interface_contract", "link_load_semantics", "runtime"},
		InteropSubprocessProtocol: {"interface_contract", "link_load_semantics", "protocol_schema"},
	}
	for _, mode := range []InteropMode{InteropCABI, InteropCXX, InteropObjCRuntime, InteropNativeLink, InteropDynamicLoad, InteropHostExtension, InteropSubprocessProtocol} {
		t.Run(string(mode), func(t *testing.T) {
			node := roleInteropFixture("mode-"+string(mode), mode)
			if err := node.Validate(); err != nil {
				t.Fatalf("full %s contract rejected: %v", mode, err)
			}
			for _, field := range required[mode] {
				t.Run("missing "+field, func(t *testing.T) {
					payload := node.Payload.(InteropBoundaryPayload)
					clearInteropField(&payload, field)
					mutated := node
					mutated.Payload = payload
					if err := mutated.Validate(); err == nil || !strings.Contains(err.Error(), field) {
						t.Fatalf("missing %s error = %v", field, err)
					}
				})
			}
		})
	}

	for _, side := range []string{"provider", "consumer"} {
		t.Run("empty "+side+" language classes", func(t *testing.T) {
			node := roleInteropFixture("empty-language-"+side, InteropCABI)
			payload := node.Payload.(InteropBoundaryPayload)
			if side == "provider" {
				payload.ProviderLanguageClasses = []string{}
			} else {
				payload.ConsumerLanguageClasses = []string{}
			}
			node.Payload = payload
			if err := node.Validate(); err == nil || !strings.Contains(err.Error(), "non-empty provider and consumer language classes") {
				t.Fatalf("empty %s language error = %v", side, err)
			}
		})
	}
}

func TestSelectedInteropBoundaryRequiresExactCompatibleSides(t *testing.T) {
	for _, missingKind := range []EdgeKind{EdgeProvidesInterop, EdgeConsumesInterop} {
		t.Run("missing "+string(missingKind), func(t *testing.T) {
			nodes, edges, product := interopValidationRecords(t, InteropCABI, "c", "c", []ID{testDigest('f')}, "c-abi-v1", false)
			filtered := make([]Edge, 0, len(edges)-1)
			for _, edge := range edges {
				if edge.Kind != missingKind {
					filtered = append(filtered, edge)
				}
			}
			_, err := fixtureBundleResult(t, nodes, filtered, product)
			want := "exactly one provider side"
			if missingKind == EdgeConsumesInterop {
				want = "exactly one consumer side"
			}
			if err == nil || !strings.Contains(err.Error(), string(CodeInteropUndeclared)) || !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %v, want missing interop-side rejection containing %q", err, want)
			}
		})
	}

	tests := []struct {
		name             string
		providerLanguage string
		evidence         []ID
		abiExpectation   string
		duplicate        bool
		want             string
	}{
		{name: "duplicate provider", providerLanguage: "c", evidence: []ID{testDigest('f')}, abiExpectation: "c-abi-v1", duplicate: true, want: "exactly one provider side, got 2"},
		{name: "provider language mismatch", providerLanguage: "rust", evidence: []ID{testDigest('f')}, abiExpectation: "c-abi-v1", want: "languages are incompatible"},
		{name: "missing provider evidence", providerLanguage: "c", evidence: []ID{}, abiExpectation: "c-abi-v1", want: "at least one immutable interface evidence ID"},
		{name: "consumer ABI mismatch", providerLanguage: "c", evidence: []ID{testDigest('f')}, abiExpectation: "different-abi-v1", want: "does not match boundary ABI"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			nodes, edges, product := interopValidationRecords(t, InteropCABI, testCase.providerLanguage, "c", testCase.evidence, testCase.abiExpectation, testCase.duplicate)
			_, leftErr := fixtureBundleResult(t, nodes, edges, product)
			if leftErr == nil || !strings.Contains(leftErr.Error(), string(CodeInteropUndeclared)) || !strings.Contains(leftErr.Error(), testCase.want) {
				t.Fatalf("error = %v, want interop rejection containing %q", leftErr, testCase.want)
			}
			reverseNodes(nodes)
			reverseEdges(edges)
			_, rightErr := fixtureBundleResult(t, nodes, edges, product)
			if rightErr == nil || leftErr.Error() != rightErr.Error() {
				t.Fatalf("interop rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
			}
		})
	}
}

func TestEveryInteropModeClosesExactSidesAndBuildOrdering(t *testing.T) {
	for _, mode := range []InteropMode{InteropCABI, InteropCXX, InteropObjCRuntime, InteropNativeLink, InteropDynamicLoad, InteropHostExtension, InteropSubprocessProtocol} {
		t.Run(string(mode), func(t *testing.T) {
			var nodes []Node
			var edges []Edge
			var product Node
			if mode == InteropSubprocessProtocol {
				nodes, edges, product, _, _ = subprocessInteropRecords(t)
			} else {
				nodes, edges, product = interopValidationRecords(t, mode, "c", "c", []ID{testDigest('f')}, "c-abi-v1", false)
			}
			bundle := fixtureBundle(t, nodes, edges, product)
			plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
			if err != nil {
				t.Fatal(err)
			}
			wantOrdering := 1
			if mode == InteropDynamicLoad || mode == InteropSubprocessProtocol {
				wantOrdering = 0
			}
			if len(plan.OrderingEdges) != wantOrdering {
				t.Fatalf("%s ordering edges = %#v, want %d", mode, plan.OrderingEdges, wantOrdering)
			}
			if wantOrdering == 1 && plan.OrderingEdges[0].Reason != OrderInterop {
				t.Fatalf("%s ordering reason = %s", mode, plan.OrderingEdges[0].Reason)
			}
		})
	}
}

func TestSubprocessBoundaryRequiresInvocationAndPublicationContract(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		mutate func(*testing.T, []Node, []Edge) ([]Node, []Edge)
	}{
		{name: "missing invocation", want: "exactly one consumer-to-provider invokes edge", mutate: func(_ *testing.T, nodes []Node, edges []Edge) ([]Node, []Edge) {
			return nodes, filterEdges(edges, func(edge Edge) bool { return edge.Kind != EdgeInvokes })
		}},
		{name: "protocol mismatch", want: "does not match boundary protocol", mutate: func(_ *testing.T, nodes []Node, edges []Edge) ([]Node, []Edge) {
			for index := range edges {
				if edges[index].Kind == EdgeInvokes {
					payload := edges[index].Payload.(InvokesPayload)
					payload.ProtocolSchema = "different-json-v1"
					edges[index].Payload = payload
				}
			}
			return nodes, edges
		}},
		{name: "missing working directory", want: "working-directory contract", mutate: func(_ *testing.T, nodes []Node, edges []Edge) ([]Node, []Edge) {
			for index := range edges {
				if edges[index].Kind == EdgeInvokes {
					payload := edges[index].Payload.(InvokesPayload)
					payload.WorkingDirectory = ""
					edges[index].Payload = payload
				}
			}
			return nodes, edges
		}},
		{name: "missing provider publication", want: "provider output requires exactly one publication edge", mutate: func(t *testing.T, nodes []Node, edges []Edge) ([]Node, []Edge) {
			providerID := ID("")
			for _, edge := range edges {
				if edge.Kind == EdgeProvidesInterop {
					providerID = edge.FromNodeID
					break
				}
			}
			if providerID == "" {
				t.Fatal("fixture has no subprocess provider")
			}
			return nodes, filterEdges(edges, func(edge Edge) bool { return edge.Kind != EdgePublishes || edge.ToNodeID != providerID })
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			nodes, edges, product, _, _ := subprocessInteropRecords(t)
			nodes, edges = testCase.mutate(t, nodes, edges)
			_, leftErr := fixtureBundleResult(t, nodes, edges, product)
			if leftErr == nil || !strings.Contains(leftErr.Error(), string(CodeInteropUndeclared)) || !strings.Contains(leftErr.Error(), testCase.want) {
				t.Fatalf("error = %v, want subprocess rejection containing %q", leftErr, testCase.want)
			}
			reverseNodes(nodes)
			reverseEdges(edges)
			_, rightErr := fixtureBundleResult(t, nodes, edges, product)
			if rightErr == nil || leftErr.Error() != rightErr.Error() {
				t.Fatalf("subprocess rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
			}
		})
	}

	t.Run("non-output sides", func(t *testing.T) {
		nodes, edges, product := interopValidationRecords(t, InteropSubprocessProtocol, "go", "go", []ID{testDigest('f')}, "protocol-v1", false)
		_, err := fixtureBundleResult(t, nodes, edges, product)
		if err == nil || !strings.Contains(err.Error(), "distinct produced output_artifact nodes") {
			t.Fatalf("error = %v, want non-output subprocess-side rejection", err)
		}
	})
}

func TestExecutableRuntimeInteropModesRequireProducedOrSelectedProviders(t *testing.T) {
	for _, mode := range []InteropMode{InteropDynamicLoad, InteropHostExtension} {
		t.Run(string(mode), func(t *testing.T) {
			nodes, edges, product := interopValidationRecords(t, mode, "c", "c", []ID{testDigest('f')}, "c-abi-v1", false)
			providerTargetID := ID("")
			for _, node := range nodes {
				if node.Kind == NodeTargetUnit && node.LogicalKey == "target:interop-provider" {
					providerTargetID = mustNodeID(t, node)
					break
				}
			}
			if providerTargetID == "" {
				t.Fatal("fixture has no provider target")
			}
			for index := range edges {
				if edges[index].Kind == EdgeProvidesInterop {
					edges[index].FromNodeID = providerTargetID
				}
			}
			_, leftErr := fixtureBundleResult(t, nodes, edges, product)
			if leftErr == nil || !strings.Contains(leftErr.Error(), string(CodeInteropUndeclared)) || !strings.Contains(leftErr.Error(), "provider must be an exact produced output or selected external toolchain") {
				t.Fatalf("error = %v, want executable-provider rejection", leftErr)
			}
			reverseNodes(nodes)
			reverseEdges(edges)
			_, rightErr := fixtureBundleResult(t, nodes, edges, product)
			if rightErr == nil || leftErr.Error() != rightErr.Error() {
				t.Fatalf("runtime-provider rejection changed under permutation:\nleft: %v\nright: %v", leftErr, rightErr)
			}
		})
	}
}

func TestInteropToolchainBindingIsExplicitAndToolchainScoped(t *testing.T) {
	bundle, _, _ := interopFixture(t, InteropCABI)
	reproject := func(edges []Edge) error {
		captureID, err := bundle.Capture.ID()
		if err != nil {
			t.Fatal(err)
		}
		selectionID, err := bundle.Selection.ID()
		if err != nil {
			t.Fatal(err)
		}
		binding, err := NewSelectionBinding(captureID, selectionID, bundle.Records.BindingNodes, edges)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ProjectActive(bundle.Capture, bundle.Selection, binding, NewRecordTables(bundle.Records.CaptureNodes, bundle.Records.CaptureEdges, bundle.Records.BindingNodes, edges), bundle.Authority, nil)
		return err
	}

	t.Run("missing boundary toolchain", func(t *testing.T) {
		edges := filterEdges(bundle.Records.BindingEdges, func(edge Edge) bool { return edge.Kind != EdgeRequires })
		err := reproject(edges)
		if err == nil || !strings.Contains(err.Error(), string(CodeInteropUndeclared)) || !strings.Contains(err.Error(), "explicit toolchain-scoped binding") {
			t.Fatalf("error = %v, want missing interop toolchain rejection", err)
		}
	})

	t.Run("wrong binding scope", func(t *testing.T) {
		edges := append([]Edge{}, bundle.Records.BindingEdges...)
		for index := range edges {
			if edges[index].Kind == EdgeRequires {
				payload := edges[index].Payload.(RequiresPayload)
				payload.Scope = ScopeBuild
				edges[index].Payload = payload
			}
		}
		err := reproject(edges)
		if err == nil || !strings.Contains(err.Error(), string(CodeGraphReferenceInvalid)) || !strings.Contains(err.Error(), "must have toolchain scope") {
			t.Fatalf("error = %v, want binding-scope rejection", err)
		}
	})
}

func interopValidationRecords(t *testing.T, mode InteropMode, providerLanguage, declaredProviderLanguage string, evidence []ID, abiExpectation string, duplicateProvider bool) ([]Node, []Edge, Node) {
	t.Helper()
	product := fixtureProduct("interop-validation")
	providerTarget := fixtureTarget("interop-provider", providerLanguage)
	consumerTarget := fixtureTarget("interop-consumer", "swift")
	providerWrites := []string{}
	localExecutableProvider := mode == InteropDynamicLoad || mode == InteropHostExtension
	if localExecutableProvider {
		providerWrites = []string{"module"}
	}
	providerAction := fixtureAction("interop-provider", []string{}, providerWrites)
	consumerAction := fixtureAction("interop-consumer", []string{}, []string{})
	boundary := roleInteropFixture("validation-"+string(mode), mode)
	boundaryPayload := boundary.Payload.(InteropBoundaryPayload)
	boundaryPayload.ProviderLanguageClasses = []string{declaredProviderLanguage}
	boundaryPayload.ConsumerLanguageClasses = []string{"swift"}
	boundary.Payload = boundaryPayload
	productID := mustNodeID(t, product)
	providerTargetID, consumerTargetID := mustNodeID(t, providerTarget), mustNodeID(t, consumerTarget)
	providerActionID, consumerActionID, boundaryID := mustNodeID(t, providerAction), mustNodeID(t, consumerAction), mustNodeID(t, boundary)
	nodes := []Node{product, providerTarget, consumerTarget, providerAction, consumerAction, boundary}
	providerSideID := providerTargetID
	var providerOutput Node
	if localExecutableProvider {
		providerOutput = Node{Kind: NodeOutputArtifact, LogicalKey: "output:interop-provider-" + string(mode), Payload: OutputArtifactPayload{Profile: "fixture-source-v1", LogicalPath: "lib/interop-provider", ExpectedClass: "native.module", OutputRole: "runtime_module"}}
		providerSideID = mustNodeID(t, providerOutput)
		nodes = append(nodes, providerOutput)
	}
	edges := []Edge{
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-interop-provider", FromNodeID: productID, ToNodeID: providerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-interop-consumer", FromNodeID: productID, ToNodeID: consumerTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:interop-provider-declares-action", FromNodeID: providerTargetID, ToNodeID: providerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider.action")}},
		{Kind: EdgeDeclares, EdgeKey: "edge:interop-consumer-declares-action", FromNodeID: consumerTargetID, ToNodeID: consumerActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.consumer.action")}},
		{Kind: EdgeProvidesInterop, EdgeKey: "edge:interop-provider-provides", FromNodeID: providerSideID, ToNodeID: boundaryID, Payload: ProvidesInteropPayload{Origin: fixtureOrigin("targets.provider.interface"), EvidenceIDs: append([]ID{}, evidence...), ExportRole: "interface", LinkMode: "selected-mode"}},
		{Kind: EdgeConsumesInterop, EdgeKey: "edge:interop-consumer-consumes", FromNodeID: consumerActionID, ToNodeID: boundaryID, Payload: ConsumesInteropPayload{Origin: fixtureOrigin("targets.consumer.import"), Use: "compile", ABIExpectation: abiExpectation}},
	}
	if localExecutableProvider {
		edges = append(edges, Edge{Kind: EdgeProduces, EdgeKey: "edge:interop-provider-produces-module", FromNodeID: providerActionID, ToNodeID: providerSideID, Payload: ProducesPayload{Path: "lib/interop-provider", WriteSlot: "module", WriteClass: "native.module"}})
	}
	if duplicateProvider {
		secondTarget := fixtureTarget("interop-provider-two", providerLanguage)
		secondAction := fixtureAction("interop-provider-two", []string{}, []string{})
		secondTargetID, secondActionID := mustNodeID(t, secondTarget), mustNodeID(t, secondAction)
		nodes = append(nodes, secondTarget, secondAction)
		edges = append(edges,
			Edge{Kind: EdgeDeclares, EdgeKey: "edge:product-declares-interop-provider-two", FromNodeID: productID, ToNodeID: secondTargetID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider-two")}},
			Edge{Kind: EdgeDeclares, EdgeKey: "edge:interop-provider-two-declares-action", FromNodeID: secondTargetID, ToNodeID: secondActionID, Payload: DeclaresPayload{Origin: fixtureOrigin("targets.provider-two.action")}},
			Edge{Kind: EdgeProvidesInterop, EdgeKey: "edge:interop-provider-two-provides", FromNodeID: secondTargetID, ToNodeID: boundaryID, Payload: ProvidesInteropPayload{Origin: fixtureOrigin("targets.provider-two.interface"), EvidenceIDs: append([]ID{}, evidence...), ExportRole: "headers", LinkMode: "static"}},
		)
	}
	return nodes, edges, product
}

func hostTargetFixture(key string) Node {
	node := fixtureTarget(key, "go")
	payload := node.Payload.(TargetUnitPayload)
	payload.ExecutionDomain = ExecutionHost
	node.Payload = payload
	return node
}

func hostActionFixture(key string) Node {
	node := fixtureAction(key, []string{}, []string{})
	payload := node.Payload.(ActionPayload)
	payload.ExecutionDomain = ExecutionHost
	node.Payload = payload
	return node
}

func roleToolchainFixture(key string, domain ExecutionDomain) Node {
	return Node{Kind: NodeToolchainComponent, LogicalKey: "toolchain:" + key, Payload: ToolchainComponentPayload{ComponentRole: "compiler", ContentFingerprint: testDigest('7'), ExecutableRelativePath: "bin/compiler", PlatformABI: "fixture-platform-v1", PolicySelector: "toolchain-v1", VersionOutput: "compiler 1.0", ExecutionDomain: domain}}
}

func roleInteropFixture(key string, mode InteropMode) Node {
	payload := InteropBoundaryPayload{
		Profile:                 "fixture-source-v1",
		Mode:                    mode,
		ProviderLanguageClasses: []string{"c"},
		ConsumerLanguageClasses: []string{"swift"},
		ContractDigest:          testDigest('6'),
		ABI:                     "c-abi-v1",
		Runtime:                 "fixture-runtime-v1",
		ProtocolSchema:          "fixture-json-v1",
		InterfaceContract:       "fixture-interface-v1",
		CallingConvention:       "c",
		LinkLoadSemantics:       "static-link",
	}
	if mode == InteropHostExtension {
		payload.ProviderLanguageClasses = []string{"host"}
		payload.ConsumerLanguageClasses = []string{"plugin"}
	}
	if mode == InteropSubprocessProtocol {
		payload.ProviderLanguageClasses = []string{"go"}
		payload.ConsumerLanguageClasses = []string{"swift"}
		payload.ABI = ""
		payload.LinkLoadSemantics = "runtime-invoke"
	}
	return Node{Kind: NodeInteropBoundary, LogicalKey: "interop:" + key, Payload: payload}
}

func nodeWithPlatformRoles(node Node, roles []PlatformRole) Node {
	switch payload := node.Payload.(type) {
	case CommandProductPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	case TargetUnitPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	case ActionPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	case ToolchainComponentPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	case InteropBoundaryPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	case OutputArtifactPayload:
		payload.PlatformRoleNames = roles
		node.Payload = payload
	default:
		panic("node kind does not declare platform roles")
	}
	return node
}

func clearInteropField(payload *InteropBoundaryPayload, field string) {
	switch field {
	case "abi":
		payload.ABI = ""
	case "runtime":
		payload.Runtime = ""
	case "protocol_schema":
		payload.ProtocolSchema = ""
	case "interface_contract":
		payload.InterfaceContract = ""
	case "calling_convention":
		payload.CallingConvention = ""
	case "link_load_semantics":
		payload.LinkLoadSemantics = ""
	default:
		panic("unknown interop field " + field)
	}
}

func TestInteropFixtureRemainsCanonicalAfterValidationTightening(t *testing.T) {
	bundle, providerActionID, consumerActionID := interopFixture(t, InteropCABI)
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Waves, [][]ID{{providerActionID}, {consumerActionID}}) {
		t.Fatalf("interop waves = %v", plan.Waves)
	}
}
