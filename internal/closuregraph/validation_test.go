package closuregraph

import (
	"errors"
	"strings"
	"testing"
)

type graphInputs struct {
	capture   CaptureGraph
	selection SelectionContext
	binding   SelectionBinding
	records   RecordTables
	authority BindingAuthority
}

func cgp10Inputs(t *testing.T) graphInputs {
	t.Helper()
	goldens := loadGoldenRecords(t)
	captureNodes := []Node{mustGolden[Node](t, goldens, "cgp10.product"), mustGolden[Node](t, goldens, "cgp10.source"), mustGolden[Node](t, goldens, "cgp10.action"), mustGolden[Node](t, goldens, "cgp10.output")}
	captureEdges := []Edge{mustGolden[Edge](t, goldens, "cgp10.declares"), mustGolden[Edge](t, goldens, "cgp10.reads"), mustGolden[Edge](t, goldens, "cgp10.produces"), mustGolden[Edge](t, goldens, "cgp10.publishes")}
	bindingNodes := []Node{mustGolden[Node](t, goldens, "cgp10.platform"), mustGolden[Node](t, goldens, "cgp10.toolchain")}
	bindingEdges := []Edge{mustGolden[Edge](t, goldens, "cgp10.uses-tool"), mustGolden[Edge](t, goldens, "cgp10.targets.product"), mustGolden[Edge](t, goldens, "cgp10.targets.action"), mustGolden[Edge](t, goldens, "cgp10.targets.toolchain"), mustGolden[Edge](t, goldens, "cgp10.targets.output")}
	return graphInputs{capture: mustGolden[CaptureGraph](t, goldens, "cgp10.capture"), selection: mustGolden[SelectionContext](t, goldens, "cgp10.selection"), binding: mustGolden[SelectionBinding](t, goldens, "cgp10.binding"), records: NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges), authority: c4AuthorityForNode(t, bindingNodes[1])}
}

func c4AuthorityForNode(t *testing.T, node Node) BindingAuthority {
	t.Helper()
	selector, err := NewToolchainSelector(node)
	if err != nil {
		t.Fatal(err)
	}
	selectorID, err := selector.ID()
	if err != nil {
		t.Fatal(err)
	}
	nodeID, _ := node.ID()
	return BindingAuthority{
		Toolchains:  []ToolchainBindingEvidence{{NodeID: nodeID, FirstBound: ToolchainBoundAtC4, EvidenceID: selectorID}},
		C4Selectors: []ToolchainSelector{selector},
	}
}

func projectInputs(t *testing.T, input graphInputs) (GraphBundle, error) {
	t.Helper()
	return ProjectActive(input.capture, input.selection, input.binding, input.records, input.authority, nil)
}

func rebuildBinding(t *testing.T, input graphInputs, nodes []Node, edges []Edge) graphInputs {
	t.Helper()
	captureID, _ := input.capture.ID()
	selectionID, _ := input.selection.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, nodes, edges)
	if err != nil {
		t.Fatal(err)
	}
	input.binding = binding
	input.records = NewRecordTables(input.records.CaptureNodes, input.records.CaptureEdges, nodes, edges)
	return input
}

func TestGraphValidationRejectsCanonicalReferenceFailures(t *testing.T) {
	fakeID := ID("sha256:" + strings.Repeat("f", 64))
	tests := []struct {
		name      string
		mutate    func(*testing.T, graphInputs) (GraphBundle, error)
		issueText string
	}{
		{name: "duplicate logical key across tables", issueText: "duplicate logical key", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			duplicate := input.records.BindingNodes[0]
			duplicate.LogicalKey = input.records.CaptureNodes[0].LogicalKey
			nodes := append(append([]Node{}, input.records.BindingNodes...), duplicate)
			input = rebuildBinding(t, input, nodes, input.records.BindingEdges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Active: ActiveGraph{}, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "duplicate node identity across tables", issueText: "duplicate node ID", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			nodes := append(append([]Node{}, input.records.BindingNodes...), input.records.CaptureNodes[0])
			captureID, _ := input.capture.ID()
			selectionID, _ := input.selection.ID()
			input.binding = SelectionBinding{SchemaID: SchemaSelectionBinding, CapturedGraphID: captureID, SelectionContextID: selectionID, BindingNodeIDs: append(sortedIDs(input.binding.BindingNodeIDs), mustNodeID(t, input.records.CaptureNodes[0])), BindingEdgeIDs: input.binding.BindingEdgeIDs}
			input.binding.BindingNodeIDs = sortedIDs(input.binding.BindingNodeIDs)
			input.records = NewRecordTables(input.records.CaptureNodes, input.records.CaptureEdges, nodes, input.records.BindingEdges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "forbidden binding node kind", issueText: "binding node kind", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			bad := input.records.CaptureNodes[1]
			bad.LogicalKey = "package:binding-forbidden"
			nodes := append(append([]Node{}, input.records.BindingNodes...), bad)
			input = rebuildBinding(t, input, nodes, input.records.BindingEdges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "dangling binding endpoint", issueText: "dangling endpoint", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			edges := append([]Edge{}, input.records.BindingEdges...)
			edges[0].ToNodeID = fakeID
			input = rebuildBinding(t, input, input.records.BindingNodes, edges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "wrong kind endpoint", issueText: "wrong-kind endpoints", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			edges := append([]Edge{}, input.records.BindingEdges...)
			toolID := mustNodeID(t, input.records.BindingNodes[1])
			for index := range edges {
				if edges[index].Kind == EdgeTargets {
					edges[index].ToNodeID = toolID
					break
				}
			}
			input = rebuildBinding(t, input, input.records.BindingNodes, edges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "capture replacing binding edge", issueText: "replaces capture semantics", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			extra := Edge{Kind: EdgeRequires, EdgeKey: "edge:binding-capture-replacement", FromNodeID: nodeIDByKind(t, input.records.CaptureNodes, NodeCommandProduct), ToNodeID: nodeIDByKind(t, input.records.CaptureNodes, NodeOutputArtifact), Payload: RequiresPayload{Scope: ScopeBuild, Origin: EvidenceOrigin{Field: "selection.replacement"}}}
			edges := append(append([]Edge{}, input.records.BindingEdges...), extra)
			input = rebuildBinding(t, input, input.records.BindingNodes, edges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "duplicate semantic target edge", issueText: "duplicate semantic edge", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			duplicate := input.records.BindingEdges[1]
			duplicate.EdgeKey += ":duplicate"
			edges := append(append([]Edge{}, input.records.BindingEdges...), duplicate)
			input = rebuildBinding(t, input, input.records.BindingNodes, edges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "duplicate edge identity across tables", issueText: "duplicate edge ID", mutate: func(t *testing.T, input graphInputs) (GraphBundle, error) {
			edges := append(append([]Edge{}, input.records.BindingEdges...), input.records.CaptureEdges[0])
			input = rebuildBinding(t, input, input.records.BindingNodes, edges)
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "missing toolchain authority", issueText: "binding authority", mutate: func(_ *testing.T, input graphInputs) (GraphBundle, error) {
			input.authority = BindingAuthority{}
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "arbitrary C4 authority digest", issueText: "selector identity", mutate: func(_ *testing.T, input graphInputs) (GraphBundle, error) {
			input.authority.Toolchains = append([]ToolchainBindingEvidence{}, input.authority.Toolchains...)
			input.authority.Toolchains[0].EvidenceID = testDigest('f')
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "C4 selector fingerprint drift", issueText: "selector does not match", mutate: func(_ *testing.T, input graphInputs) (GraphBundle, error) {
			input.authority.C4Selectors = append([]ToolchainSelector{}, input.authority.C4Selectors...)
			input.authority.C4Selectors[0].ContentFingerprint = testDigest('f')
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
		{name: "noncanonical record table", issueText: "record table is not canonical", mutate: func(_ *testing.T, input graphInputs) (GraphBundle, error) {
			input.records.CaptureNodes[0], input.records.CaptureNodes[1] = input.records.CaptureNodes[1], input.records.CaptureNodes[0]
			return GraphBundle{Capture: input.capture, Selection: input.selection, Binding: input.binding, Records: input.records, Authority: input.authority}, nil
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := cgp10Inputs(t)
			bundle, preliminary := testCase.mutate(t, input)
			if preliminary != nil {
				requireReferenceIssue(t, preliminary, testCase.issueText)
				return
			}
			requireReferenceIssue(t, bundle.Validate(), testCase.issueText)
		})
	}
}

func TestC0ToolchainAuthorityResolvesExactCheckpointMembership(t *testing.T) {
	input := cgp10Inputs(t)
	selectionID, _ := input.selection.ID()
	platformID := nodeIDByKind(t, input.records.BindingNodes, NodeTargetPlatform)
	toolchainID := nodeIDByKind(t, input.records.BindingNodes, NodeToolchainComponent)
	c0Payload := C0ProfilePayload{AdapterProfileID: "fixture-source-v1", SchemaIDs: []string{"closure-v1"}, ArtifactPolicyID: "artifact-policy-v1", DetectorRegistryID: "detectors-v1", SourceGrammarIDs: []string{"source-v1"}, LimitVectorID: "limits-v1", SelectionContextID: selectionID, PlatformNodeIDs: []ID{platformID}, PlatformRoles: clonePlatformRoles(input.selection.PlatformRoles), ManagerSchemaIDs: []string{"manager-v1"}, ConfigurationPolicyID: "configuration-v1", CapabilityIDs: []string{}, EvidenceToolchainNodeIDs: []ID{toolchainID}}
	c0, err := NewCheckpoint(c0Payload, nil, []Diagnostic{})
	if err != nil {
		t.Fatal(err)
	}
	c0ID, _ := c0.ID()
	input.authority = BindingAuthority{Toolchains: []ToolchainBindingEvidence{{NodeID: toolchainID, FirstBound: ToolchainBoundAtC0, EvidenceID: c0ID}}, C0Checkpoint: &c0}
	if _, err := projectInputs(t, input); err != nil {
		t.Fatal(err)
	}

	c0Payload.EvidenceToolchainNodeIDs = []ID{}
	drifted, err := NewCheckpoint(c0Payload, nil, []Diagnostic{})
	if err != nil {
		t.Fatal(err)
	}
	driftedID, _ := drifted.ID()
	input.authority = BindingAuthority{Toolchains: []ToolchainBindingEvidence{{NodeID: toolchainID, FirstBound: ToolchainBoundAtC0, EvidenceID: driftedID}}, C0Checkpoint: &drifted}
	_, err = projectInputs(t, input)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error = %T %v, want ValidationError", err, err)
	}
	found := false
	for _, issue := range validation.Issues {
		if issue.Code == CodeCheckpointInvalid && strings.Contains(issue.Message, "exact C0 checkpoint and tool table") {
			found = true
		}
	}
	if !found {
		t.Fatalf("issues = %#v, want exact C0 membership failure", validation.Issues)
	}
}

func TestActiveValidationRejectsMissingOrMultiplyBoundSlotsAndTargets(t *testing.T) {
	tests := []struct {
		name, text string
		mutate     func(*testing.T, graphInputs) graphInputs
	}{
		{name: "missing tool slot", text: "tool slot", mutate: func(t *testing.T, input graphInputs) graphInputs {
			edges := filterEdges(input.records.BindingEdges, func(edge Edge) bool { return edge.Kind != EdgeUsesTool })
			return rebuildBinding(t, input, input.records.BindingNodes, edges)
		}},
		{name: "duplicate tool slot", text: "tool slot", mutate: func(t *testing.T, input graphInputs) graphInputs {
			var tool Edge
			for _, edge := range input.records.BindingEdges {
				if edge.Kind == EdgeUsesTool {
					tool = edge
					break
				}
			}
			tool.EdgeKey += ":second"
			payload := tool.Payload.(UsesToolPayload)
			payload.InvocationRole = "secondary"
			tool.Payload = payload
			edges := append(append([]Edge{}, input.records.BindingEdges...), tool)
			return rebuildBinding(t, input, input.records.BindingNodes, edges)
		}},
		{name: "duplicate target slot", text: "duplicate semantic edge", mutate: func(t *testing.T, input graphInputs) graphInputs {
			var target Edge
			for _, edge := range input.records.BindingEdges {
				if edge.Kind == EdgeTargets && edge.FromNodeID == nodeIDByKind(t, input.records.CaptureNodes, NodeOutputArtifact) {
					target = edge
					break
				}
			}
			target.EdgeKey += ":second"
			payload := target.Payload.(TargetsPayload)
			payload.Origin.Field = "selection.outputs.second-target"
			target.Payload = payload
			edges := append(append([]Edge{}, input.records.BindingEdges...), target)
			return rebuildBinding(t, input, input.records.BindingNodes, edges)
		}},
		{name: "missing target binding", text: "platform role", mutate: func(t *testing.T, input graphInputs) graphInputs {
			outputID := nodeIDByKind(t, input.records.CaptureNodes, NodeOutputArtifact)
			edges := filterEdges(input.records.BindingEdges, func(edge Edge) bool { return edge.Kind != EdgeTargets || edge.FromNodeID != outputID })
			return rebuildBinding(t, input, input.records.BindingNodes, edges)
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.mutate(t, cgp10Inputs(t))
			_, err := projectInputs(t, input)
			requireReferenceIssue(t, err, testCase.text)
		})
	}
}

func TestSelectionRejectsRoleMismatchedPlatformEdge(t *testing.T) {
	input := cgp10Inputs(t)
	host := input.records.BindingNodes[0]
	host.LogicalKey = "platform:host-linux-x86_64"
	host.Payload = TargetPlatformPayload{OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc", MinimumRuntime: "glibc-2.31", SDKID: "linux-sysroot-v1", TargetTriple: "x86_64-unknown-linux-gnu"}
	hostID := mustNodeID(t, host)
	selection := input.selection
	selection.PlatformRoles = clonePlatformRoles(selection.PlatformRoles)
	selection.PlatformRoles[PlatformHost] = hostID
	selectionID, err := selection.ID()
	if err != nil {
		t.Fatal(err)
	}
	input.selection = selection
	edges := append([]Edge{}, input.records.BindingEdges...)
	for index := range edges {
		if edges[index].Kind == EdgeTargets {
			payload := edges[index].Payload.(TargetsPayload)
			payload.BindingRole = PlatformHost
			edges[index].Payload = payload
			break
		}
	}
	nodes := append(append([]Node{}, input.records.BindingNodes...), host)
	captureID, _ := input.capture.ID()
	binding, err := NewSelectionBinding(captureID, selectionID, nodes, edges)
	if err != nil {
		t.Fatal(err)
	}
	input.binding = binding
	input.records = NewRecordTables(input.records.CaptureNodes, input.records.CaptureEdges, nodes, edges)
	_, err = projectInputs(t, input)
	requireReferenceIssue(t, err, "expected")
}

func TestNodeCodecRejectsRelationalContaminationAndUnknownKinds(t *testing.T) {
	goldens := loadGoldenRecords(t)
	raw, err := decodeCanonicalObject(goldens["cgp10.output"].Payload, "test node")
	if err != nil {
		t.Fatal(err)
	}
	raw["payload"].(map[string]any)["producer_node_id"] = string(goldens["cgp10.action"].ID)
	contaminated, err := canonicalMapBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeNode(contaminated); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("relational field error = %v", err)
	}
	raw, _ = decodeCanonicalObject(goldens["cgp10.output"].Payload, "test node")
	raw["kind"] = "recursive_magic"
	unknown, _ := canonicalMapBytes(raw)
	if _, err := DecodeNode(unknown); err == nil || !strings.Contains(err.Error(), string(CodeGraphSchemaUnsupported)) {
		t.Fatalf("unknown kind error = %v", err)
	}
}

func filterEdges(edges []Edge, keep func(Edge) bool) []Edge {
	result := []Edge{}
	for _, edge := range edges {
		if keep(edge) {
			result = append(result, edge)
		}
	}
	return result
}
func mustNodeID(t *testing.T, node Node) ID {
	t.Helper()
	id, err := node.ID()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func nodeIDByKind(t *testing.T, nodes []Node, kind NodeKind) ID {
	t.Helper()
	for _, node := range nodes {
		if node.Kind == kind {
			return mustNodeID(t, node)
		}
	}
	t.Fatalf("missing node kind %s", kind)
	return ""
}
func requireReferenceIssue(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatal("invalid graph was accepted")
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error %T = %v, want ValidationError", err, err)
	}
	if validation.Code() != CodeGraphReferenceInvalid && validation.Code() != CodeGraphIncomplete {
		t.Fatalf("code = %s, want graph reference/incomplete", validation.Code())
	}
	for _, issue := range validation.Issues {
		if strings.Contains(issue.Path+" "+issue.Message, contains) {
			return
		}
	}
	t.Fatalf("issues %#v do not contain %q", validation.Issues, contains)
}
