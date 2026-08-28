package closuregraph

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReviewerReworkRejectsUndeclaredPlatformRoleForEveryBoundNodeKind(t *testing.T) {
	for _, kind := range []NodeKind{
		NodeCommandProduct,
		NodeTargetUnit,
		NodeAction,
		NodeToolchainComponent,
		NodeOutputArtifact,
		NodeInteropBoundary,
	} {
		t.Run(string(kind), func(t *testing.T) {
			base := reviewerInputWithDistinctHost(t)
			source := reviewerNodeByKind(t, base.records, kind)
			hostID := base.selection.PlatformRoles[PlatformHost]
			extra := Edge{
				Kind:       EdgeTargets,
				EdgeKey:    "edge:reviewer-undeclared-host:" + string(kind),
				FromNodeID: mustNodeID(t, source),
				ToNodeID:   hostID,
				Payload: TargetsPayload{
					BindingRole: PlatformHost,
					Origin:      EvidenceOrigin{Field: "selection.platform_roles.host.extra." + string(kind)},
				},
			}

			leftEdges := append(append([]Edge{}, base.records.BindingEdges...), extra)
			left := rebuildBinding(t, base, base.records.BindingNodes, leftEdges)
			_, leftErr := projectInputs(t, left)
			leftIssues := reviewerValidationIssues(t, leftErr)
			leftIssue := reviewerIssueAtPath(t, leftIssues, "targets.binding_role")
			if leftIssue.Code != CodeGraphReferenceInvalid ||
				!strings.Contains(leftIssue.Message, "is not declared") ||
				!strings.Contains(leftIssue.Message, source.LogicalKey) {
				t.Fatalf("undeclared %s role issue = %#v", kind, leftIssue)
			}

			rightNodes := append([]Node{}, base.records.BindingNodes...)
			rightEdges := append([]Edge{}, leftEdges...)
			reverseNodes(rightNodes)
			reverseEdges(rightEdges)
			right := rebuildBinding(t, base, rightNodes, rightEdges)
			_, rightErr := projectInputs(t, right)
			rightIssues := reviewerValidationIssues(t, rightErr)
			if !reflect.DeepEqual(leftIssues, rightIssues) {
				t.Fatalf("undeclared %s role diagnostics changed under permutation:\nleft:  %#v\nright: %#v", kind, leftIssues, rightIssues)
			}
		})
	}
}

func TestReviewerReworkRejectsUndeclaredTargetRoleForHostBoundNodeKinds(t *testing.T) {
	for _, source := range []Node{
		hostTargetFixture("reviewer-extra-target"),
		hostActionFixture("reviewer-extra-target"),
		roleToolchainFixture("reviewer-extra-target", ExecutionHost),
		roleInteropFixture("reviewer-extra-target", InteropHostExtension),
	} {
		t.Run(string(source.Kind), func(t *testing.T) {
			target := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:reviewer-extra-target", Payload: TargetPlatformPayload{
				OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc",
				MinimumRuntime: "glibc-2.31", SDKID: "linux-sysroot-v1", TargetTriple: "x86_64-unknown-linux-gnu",
			}}
			host := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:reviewer-extra-target-host", Payload: TargetPlatformPayload{
				OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem",
				MinimumRuntime: "macos-15.0", SDKID: "macosx-sdk-v1", TargetTriple: "arm64-apple-macosx15.0",
			}}
			sourceID, targetID, hostID := mustNodeID(t, source), mustNodeID(t, target), mustNodeID(t, host)
			correct := Edge{
				Kind: EdgeTargets, EdgeKey: "edge:reviewer-host-binding:" + string(source.Kind),
				FromNodeID: sourceID, ToNodeID: hostID,
				Payload: TargetsPayload{BindingRole: PlatformHost, Origin: EvidenceOrigin{Field: "selection.platform_roles.host"}},
			}
			extra := Edge{
				Kind: EdgeTargets, EdgeKey: "edge:reviewer-undeclared-target:" + string(source.Kind),
				FromNodeID: sourceID, ToNodeID: targetID,
				Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target.extra"}},
			}
			resolved := resolvedTables{
				captureNodes: map[ID]Node{},
				bindingNodes: map[ID]Node{targetID: target, hostID: host},
				allNodes:     map[ID]Node{sourceID: source, targetID: target, hostID: host},
			}
			if source.Kind == NodeToolchainComponent {
				resolved.bindingNodes[sourceID] = source
			} else {
				resolved.captureNodes[sourceID] = source
			}
			validate := func(edges []Edge) []Issue {
				selected := make(map[ID]Edge, len(edges))
				for _, edge := range edges {
					selected[mustEdgeID(t, edge)] = edge
				}
				collector := &issueCollector{}
				validatePlatformBindings(collector, map[ID]ActivationState{sourceID: ActivationSelected}, selected, resolved)
				return reviewerValidationIssues(t, collector.err())
			}

			left := validate([]Edge{correct, extra})
			leftIssue := reviewerIssueAtPath(t, left, "targets.binding_role")
			if leftIssue.Code != CodeGraphReferenceInvalid ||
				!strings.Contains(leftIssue.Message, "is not declared") ||
				!strings.Contains(leftIssue.Message, source.LogicalKey) {
				t.Fatalf("undeclared target role issue = %#v", leftIssue)
			}
			right := validate([]Edge{extra, correct})
			if !reflect.DeepEqual(left, right) {
				t.Fatalf("undeclared target role diagnostics changed under permutation:\nleft:  %#v\nright: %#v", left, right)
			}
		})
	}
}

func TestReviewerReworkPreservesHostToTargetPlatformFallback(t *testing.T) {
	product := fixtureProduct("host-fallback")
	action := hostActionFixture("host-fallback")
	edge := Edge{
		Kind:       EdgeDeclares,
		EdgeKey:    "edge:host-fallback-product-declares-action",
		FromNodeID: mustNodeID(t, product),
		ToNodeID:   mustNodeID(t, action),
		Payload:    DeclaresPayload{Origin: fixtureOrigin("actions.host-fallback")},
	}
	bundle := fixtureBundle(t, []Node{product, action}, []Edge{edge}, product)

	input := reviewerGraphInput(bundle)
	actionID := mustNodeID(t, action)
	edges := append([]Edge{}, input.records.BindingEdges...)
	found := false
	for index := range edges {
		if edges[index].Kind != EdgeTargets || edges[index].FromNodeID != actionID {
			continue
		}
		payload := edges[index].Payload.(TargetsPayload)
		payload.BindingRole = PlatformHost
		payload.Origin = EvidenceOrigin{Field: "selection.platform_roles.host.fallback"}
		edges[index].EdgeKey = "edge:action:host-fallback:targets:host-fallback"
		edges[index].Payload = payload
		found = true
		break
	}
	if !found {
		t.Fatal("host action has no platform binding")
	}
	input = rebuildBinding(t, input, input.records.BindingNodes, edges)
	if _, err := projectInputs(t, input); err != nil {
		t.Fatalf("host role did not fall back to the sole target platform: %v", err)
	}
}

func TestReviewerReworkRejectsCrossOriginSemanticDuplicatesForEveryOriginPayload(t *testing.T) {
	cases := []struct {
		name  string
		kind  EdgeKind
		table string
	}{
		{name: "declares", kind: EdgeDeclares, table: "capture"},
		{name: "resolves_to", kind: EdgeResolvesTo, table: "capture"},
		{name: "requires", kind: EdgeRequires, table: "binding"},
		{name: "targets", kind: EdgeTargets, table: "binding"},
		{name: "provides_interop", kind: EdgeProvidesInterop, table: "capture"},
		{name: "consumes_interop", kind: EdgeConsumesInterop, table: "capture"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := reviewerPlatformSemanticBundle(t)
			original := reviewerEdgeByKind(t, bundle.Records, testCase.table, testCase.kind)
			originalOrigin, ok := reviewerEvidenceOrigin(original.Payload)
			if !ok {
				t.Fatalf("%s payload does not expose evidence origin", testCase.kind)
			}
			duplicate := original
			duplicate.EdgeKey = original.EdgeKey + ":cross-origin-duplicate"
			duplicateOrigin := EvidenceOrigin{
				Field:          "reviewer.cross_origin." + string(testCase.kind),
				ManifestDigest: testDigest('f'),
			}
			duplicate.Payload = reviewerPayloadWithOrigin(t, duplicate.Payload, duplicateOrigin)

			captureEdges := append([]Edge{}, bundle.Records.CaptureEdges...)
			bindingEdges := append([]Edge{}, bundle.Records.BindingEdges...)
			if testCase.table == "capture" {
				captureEdges = append(captureEdges, duplicate)
			} else {
				bindingEdges = append(bindingEdges, duplicate)
			}
			_, leftErr := reviewerReproject(t, bundle,
				bundle.Records.CaptureNodes, captureEdges,
				bundle.Records.BindingNodes, bindingEdges,
			)
			leftIssue := reviewerIssueAtPath(t, reviewerValidationIssues(t, leftErr), "semantic edge")

			originalID := mustEdgeID(t, original)
			duplicateID := mustEdgeID(t, duplicate)
			for _, evidence := range []string{
				original.EdgeKey,
				duplicate.EdgeKey,
				string(originalID),
				string(duplicateID),
				originalOrigin.Field,
				duplicateOrigin.Field,
			} {
				if !strings.Contains(leftIssue.Message, evidence) {
					t.Fatalf("%s duplicate issue omits %q: %#v", testCase.kind, evidence, leftIssue)
				}
			}
			if strings.Contains(leftIssue.Key, originalOrigin.Field) || strings.Contains(leftIssue.Key, duplicateOrigin.Field) {
				t.Fatalf("%s semantic key contains provenance-only origin: %s", testCase.kind, leftIssue.Key)
			}

			rightCaptureNodes := append([]Node{}, bundle.Records.CaptureNodes...)
			rightCaptureEdges := append([]Edge{}, captureEdges...)
			rightBindingNodes := append([]Node{}, bundle.Records.BindingNodes...)
			rightBindingEdges := append([]Edge{}, bindingEdges...)
			reverseNodes(rightCaptureNodes)
			reverseEdges(rightCaptureEdges)
			reverseNodes(rightBindingNodes)
			reverseEdges(rightBindingEdges)
			_, rightErr := reviewerReproject(t, bundle,
				rightCaptureNodes, rightCaptureEdges,
				rightBindingNodes, rightBindingEdges,
			)
			rightIssue := reviewerIssueAtPath(t, reviewerValidationIssues(t, rightErr), "semantic edge")
			if !reflect.DeepEqual(leftIssue, rightIssue) {
				t.Fatalf("%s semantic duplicate diagnostic changed under permutation:\nleft:  %#v\nright: %#v", testCase.kind, leftIssue, rightIssue)
			}
		})
	}
}

func TestReviewerReworkSemanticKeysRetainRelationshipFields(t *testing.T) {
	bundle := reviewerPlatformSemanticBundle(t)
	cases := []struct {
		name   string
		table  string
		kind   EdgeKind
		mutate func(EdgePayload) EdgePayload
	}{
		{name: "resolves_to checksum", table: "capture", kind: EdgeResolvesTo, mutate: func(value EdgePayload) EdgePayload {
			payload := value.(ResolvesToPayload)
			payload.Checksum += "-different"
			return payload
		}},
		{name: "requires dependency kind", table: "binding", kind: EdgeRequires, mutate: func(value EdgePayload) EdgePayload {
			payload := value.(RequiresPayload)
			payload.DependencyKind = "different-toolchain-role"
			return payload
		}},
		{name: "targets binding role", table: "binding", kind: EdgeTargets, mutate: func(value EdgePayload) EdgePayload {
			payload := value.(TargetsPayload)
			payload.BindingRole = PlatformHost
			return payload
		}},
		{name: "provides interop export role", table: "capture", kind: EdgeProvidesInterop, mutate: func(value EdgePayload) EdgePayload {
			payload := value.(ProvidesInteropPayload)
			payload.ExportRole += "-different"
			return payload
		}},
		{name: "consumes interop use", table: "capture", kind: EdgeConsumesInterop, mutate: func(value EdgePayload) EdgePayload {
			payload := value.(ConsumesInteropPayload)
			payload.Use += "-different"
			return payload
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			edge := reviewerEdgeByKind(t, bundle.Records, testCase.table, testCase.kind)
			left, err := semanticEdgeKey(edge)
			if err != nil {
				t.Fatal(err)
			}
			edge.Payload = testCase.mutate(edge.Payload)
			right, err := semanticEdgeKey(edge)
			if err != nil {
				t.Fatal(err)
			}
			if left == right {
				t.Fatalf("%s was erased from semantic edge identity", testCase.name)
			}
		})
	}
}

func reviewerPlatformSemanticBundle(t *testing.T) GraphBundle {
	t.Helper()
	nodes, edges, product := interopValidationRecords(t, InteropDynamicLoad, "c", "c", []ID{testDigest('f')}, "c-abi-v1", false)
	packageNode := Node{Kind: NodePackageInstance, LogicalKey: "package:reviewer-semantic", Payload: PackageInstancePayload{
		Profile: "fixture-source-v1", Ecosystem: "fixture", Manager: "fixture-manager",
		Origin: "registry://fixture/reviewer-semantic/1.0.0", LockInstanceKey: "reviewer-semantic@1.0.0",
		Name: "reviewer-semantic", Version: "1.0.0", ArtifactManifestID: testDigest('8'), TrustRole: TrustDependencyInput,
	}}
	source := Node{Kind: NodeSourceSet, LogicalKey: "source:reviewer-semantic", Payload: SourceSetPayload{
		Profile: "fixture-source-v1", Origin: "transform://fixture/reviewer-semantic/1.0.0",
		ArtifactManifestID: testDigest('9'), Projection: []string{}, Grammar: "fixture-source-v1", TrustRole: TrustDependencyInput,
	}}
	edges = append(edges,
		Edge{Kind: EdgeRequires, EdgeKey: "edge:reviewer-product-requires-package", FromNodeID: mustNodeID(t, product), ToNodeID: mustNodeID(t, packageNode), Payload: RequiresPayload{Scope: ScopeRuntime, Origin: fixtureOrigin("dependencies.reviewer-semantic")}},
		Edge{Kind: EdgeResolvesTo, EdgeKey: "edge:reviewer-package-resolves-source", FromNodeID: mustNodeID(t, packageNode), ToNodeID: mustNodeID(t, source), Payload: ResolvesToPayload{LockField: "packages.reviewer-semantic", Origin: fixtureOrigin("packages.reviewer-semantic.resolved"), Checksum: "sha512-reviewer-semantic", ArtifactManifestID: testDigest('9')}},
	)
	nodes = append(nodes, packageNode, source)
	return fixtureBundle(t, nodes, edges, product)
}

func reviewerInputWithDistinctHost(t *testing.T) graphInputs {
	t.Helper()
	bundle := reviewerPlatformSemanticBundle(t)
	input := reviewerGraphInput(bundle)
	host := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:reviewer-host", Payload: TargetPlatformPayload{
		OS: "darwin", Architecture: "arm64", ABI: "darwin", Libc: "libSystem",
		MinimumRuntime: "macos-15.0", SDKID: "macosx-sdk-v1", TargetTriple: "arm64-apple-macosx15.0",
	}}
	hostID := mustNodeID(t, host)
	selection, err := NewSelectionContext(
		input.selection.ProductNodeIDs,
		map[PlatformRole]ID{
			PlatformTarget: input.selection.PlatformRoles[PlatformTarget],
			PlatformHost:   hostID,
		},
		input.selection.Features,
		input.selection.DefaultFeatures,
		input.selection.Markers,
		input.selection.PeerContext,
		input.selection.EvaluatorIDs,
	)
	if err != nil {
		t.Fatal(err)
	}
	input.selection = selection
	nodes := append(append([]Node{}, input.records.BindingNodes...), host)
	return rebuildBinding(t, input, nodes, input.records.BindingEdges)
}

func reviewerGraphInput(bundle GraphBundle) graphInputs {
	return graphInputs{
		capture:   bundle.Capture,
		selection: bundle.Selection,
		binding:   bundle.Binding,
		records:   bundle.Records,
		authority: bundle.Authority,
	}
}

func reviewerNodeByKind(t *testing.T, records RecordTables, kind NodeKind) Node {
	t.Helper()
	tables := [][]Node{records.CaptureNodes, records.BindingNodes}
	for _, nodes := range tables {
		for _, node := range nodes {
			if node.Kind == kind {
				return node
			}
		}
	}
	t.Fatalf("reviewer fixture has no %s node", kind)
	return Node{}
}

func reviewerEdgeByKind(t *testing.T, records RecordTables, table string, kind EdgeKind) Edge {
	t.Helper()
	edges := records.CaptureEdges
	if table == "binding" {
		edges = records.BindingEdges
	}
	for _, edge := range edges {
		if edge.Kind == kind {
			return edge
		}
	}
	t.Fatalf("reviewer fixture has no %s %s edge", table, kind)
	return Edge{}
}

func reviewerPayloadWithOrigin(t *testing.T, value EdgePayload, origin EvidenceOrigin) EdgePayload {
	t.Helper()
	switch payload := value.(type) {
	case DeclaresPayload:
		payload.Origin = origin
		return payload
	case ResolvesToPayload:
		payload.Origin = origin
		return payload
	case RequiresPayload:
		payload.Origin = origin
		return payload
	case TargetsPayload:
		payload.Origin = origin
		return payload
	case ProvidesInteropPayload:
		payload.Origin = origin
		return payload
	case ConsumesInteropPayload:
		payload.Origin = origin
		return payload
	default:
		t.Fatalf("payload %T has no evidence origin", value)
		return nil
	}
}

func reviewerEvidenceOrigin(value EdgePayload) (EvidenceOrigin, bool) {
	switch payload := value.(type) {
	case DeclaresPayload:
		return payload.Origin, true
	case ResolvesToPayload:
		return payload.Origin, true
	case RequiresPayload:
		return payload.Origin, true
	case TargetsPayload:
		return payload.Origin, true
	case ProvidesInteropPayload:
		return payload.Origin, true
	case ConsumesInteropPayload:
		return payload.Origin, true
	default:
		return EvidenceOrigin{}, false
	}
}

func reviewerReproject(t *testing.T, bundle GraphBundle, captureNodes []Node, captureEdges []Edge, bindingNodes []Node, bindingEdges []Edge) (GraphBundle, error) {
	t.Helper()
	capture, err := NewCaptureGraph(
		bundle.Capture.ProfileID,
		bundle.Capture.PolicyIDs,
		bundle.Capture.RootNodeIDs,
		captureNodes,
		captureEdges,
		bundle.Capture.ArtifactManifestIDs,
	)
	if err != nil {
		return GraphBundle{}, err
	}
	captureID, err := capture.ID()
	if err != nil {
		return GraphBundle{}, err
	}
	selectionID, err := bundle.Selection.ID()
	if err != nil {
		return GraphBundle{}, err
	}
	binding, err := NewSelectionBinding(captureID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		return GraphBundle{}, err
	}
	return ProjectActive(
		capture,
		bundle.Selection,
		binding,
		NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges),
		bundle.Authority,
		nil,
	)
}

func reviewerValidationIssues(t *testing.T, err error) []Issue {
	t.Helper()
	if err == nil {
		t.Fatal("invalid reviewer fixture was accepted")
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error %T = %v, want ValidationError", err, err)
	}
	return validation.Issues
}

func reviewerIssueAtPath(t *testing.T, issues []Issue, path string) Issue {
	t.Helper()
	for _, issue := range issues {
		if issue.Path == path {
			return issue
		}
	}
	t.Fatalf("issues %#v do not contain path %q", issues, path)
	return Issue{}
}
