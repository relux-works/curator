package swiftpminterop

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// publish republishes the accepted source closure as an interop capture graph
// plus one exact selection binding. Capture stays selection-neutral: every
// concrete platform, toolchain, SDK, system, `targets`, and `uses_tool` fact
// lives only in the binding overlay.
func (state *closeState) publish(boundaries []Boundary, reads ReadSetEvidence) (*Result, error) {
	nodes := append([]closuregraph.Node(nil), state.capture.Records.CaptureNodes...)
	edges := append([]closuregraph.Edge(nil), state.capture.Records.CaptureEdges...)
	bindingNodes := append([]closuregraph.Node(nil), state.capture.Records.BindingNodes...)
	bindingEdges := append([]closuregraph.Edge(nil), state.capture.Records.BindingEdges...)

	platformID, err := state.platformNodeID()
	if err != nil {
		return nil, err
	}
	swiftToolID, ok := state.idsByKey["swiftpm.tool.swift"]
	if !ok {
		if swiftToolID, ok = bindingNodeID(bindingNodes, "swiftpm.tool.swift"); !ok {
			return nil, fail(CodeGraphIncomplete, "accepted binding omits the exact Swift toolchain node")
		}
	}
	for _, binding := range state.components {
		bindingNodes = append(bindingNodes, binding.node)
		bindingEdges = append(bindingEdges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.interop.component-target." + binding.component.Role, FromNodeID: binding.nodeID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
		bindingEdges = append(bindingEdges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: "swiftpm.interop.component-requires." + binding.component.Role, FromNodeID: state.capture.ProductNodeID, ToNodeID: binding.nodeID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeToolchain, Origin: closuregraph.EvidenceOrigin{Field: "selection.interop_components." + binding.component.Role}, DependencyKind: "external-toolchain"}})
	}

	for index := range state.targets {
		interop := &state.targets[index]
		if interop.Kind == KindSystem {
			interop.ToolNodeID = state.systemByKey[interop.Package+":"+interop.Target].nodeID
			continue
		}
		targetNodes, targetEdges, targetBindingEdges, err := state.targetRecords(interop, swiftToolID, platformID)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, targetNodes...)
		edges = append(edges, targetEdges...)
		bindingEdges = append(bindingEdges, targetBindingEdges...)
	}

	for boundaryIndex := range boundaries {
		boundary := &boundaries[boundaryIndex]
		boundaryNodes, boundaryEdges, boundaryBindingEdges, err := state.boundaryRecords(boundary, platformID, boundaryIndex)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, boundaryNodes...)
		edges = append(edges, boundaryEdges...)
		bindingEdges = append(bindingEdges, boundaryBindingEdges...)
	}

	graph, err := closuregraph.NewCaptureGraph(ProfileID, state.capture.Graph.PolicyIDs, state.capture.Graph.RootNodeIDs, nodes, edges, state.capture.Graph.ArtifactManifestIDs)
	if err != nil {
		return nil, err
	}
	graphID, err := graph.ID()
	if err != nil {
		return nil, err
	}
	selectionID, err := state.capture.Selection.ID()
	if err != nil {
		return nil, err
	}
	binding, err := closuregraph.NewSelectionBinding(graphID, selectionID, bindingNodes, bindingEdges)
	if err != nil {
		return nil, err
	}
	authority, err := state.bindingAuthority()
	if err != nil {
		return nil, err
	}
	records := closuregraph.NewRecordTables(nodes, edges, bindingNodes, bindingEdges)
	bundle, err := closuregraph.ProjectActive(graph, state.capture.Selection, binding, records, authority, []closuregraph.ConditionEvaluator{state.evaluator()})
	if err != nil {
		return nil, err
	}
	activeID, err := bundle.Active.ID()
	if err != nil {
		return nil, err
	}
	bindingID, err := binding.ID()
	if err != nil {
		return nil, err
	}
	checkpoint, err := closuregraph.NewCheckpoint(closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: graphID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, &state.capture.C4, nil)
	if err != nil {
		return nil, err
	}
	evidenceDigest, err := state.evidenceDigest(boundaries, reads, graphID, bindingID)
	if err != nil {
		return nil, err
	}
	moduleMaps := []ModuleMapEvidence{}
	for _, interop := range state.targets {
		if interop.ModuleMap != nil {
			moduleMaps = append(moduleMaps, *interop.ModuleMap)
		}
	}
	return &Result{Targets: state.targets, Boundaries: boundaries, ModuleMaps: moduleMaps, Reads: reads, Graph: graph, Selection: state.capture.Selection, Binding: binding, Active: bundle.Active, Records: records, Authority: authority, C4: checkpoint, GraphDigest: graphID, EvidenceDigest: evidenceDigest}, nil
}

func (state *closeState) targetRecords(interop *TargetInterop, swiftToolID, platformID closuregraph.ID) ([]closuregraph.Node, []closuregraph.Edge, []closuregraph.Edge, error) {
	key := interop.Package + "." + interop.Target
	if _, ok := state.idsByKey["swiftpm.source."+interop.Package]; !ok {
		return nil, nil, nil, failFields(CodeGraphIncomplete, map[string]string{"package": interop.Package}, "accepted capture omits the package source set")
	}
	sourceNode, sourceSetID, err := state.targetSourceSet(interop)
	if err != nil {
		return nil, nil, nil, err
	}
	interop.SourceSetNodeID = sourceSetID
	nodes := []closuregraph.Node{sourceNode}
	readSlots := []string{"sources"}
	if len(interop.Headers) != 0 {
		readSlots = append(readSlots, "headers")
	}
	sort.Strings(readSlots)
	writeSlots := make([]string, len(interop.Sources))
	for index := range interop.Sources {
		writeSlots[index] = ObjectWriteSlot(index)
	}
	argv := []string{"$TOOL(compiler)"}
	for _, slot := range readSlots {
		argv = append(argv, "$READ("+slot+")")
	}
	for _, slot := range writeSlots {
		argv = append(argv, "$WRITE("+slot+")")
	}
	subtype := "swift-compile"
	toolID, toolRole := swiftToolID, state.capture.SelectionToolchain().Swift.Role
	if interop.Kind == KindClang {
		subtype = "clang-compile"
		role := state.config.Clang.Role
		if containsLanguage(interop.Languages, LanguageCXX) || containsLanguage(interop.Languages, LanguageObjCXX) {
			if state.config.ClangCXX.Role == "" {
				return nil, nil, nil, failFields(CodeToolchainUntrusted, map[string]string{"target": interop.Package + ":" + interop.Target}, "C++ family target has no selected C++ driver identity")
			}
			role = state.config.ClangCXX.Role
		}
		binding, present := state.byRole[role]
		if !present {
			return nil, nil, nil, failFields(CodeToolchainUntrusted, map[string]string{"role": role}, "C-family target names no selected driver component")
		}
		toolID, toolRole = binding.nodeID, role
	}
	interop.ToolNodeID = toolID
	action := closuregraph.Node{Kind: closuregraph.NodeAction, LogicalKey: "swiftpm.interop.compile." + key, Payload: closuregraph.ActionPayload{
		Profile: ProfileID, ActionSubtype: subtype, ExecutionDomain: closuregraph.ExecutionTarget, ArgvTemplate: argv,
		ToolSlotNames: []string{"compiler"}, ReadSlotNames: readSlots, WriteSlotNames: writeSlots,
		EnvironmentPolicyID: "swiftpm-interop-environment-v1", ProcessPolicyID: "swiftpm-interop-process-v1", Network: "none",
		PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}}
	if err = action.Validate(); err != nil {
		return nil, nil, nil, err
	}
	actionID, err := action.ID()
	if err != nil {
		return nil, nil, nil, err
	}
	interop.ActionNodeID = actionID
	nodes = append(nodes, action)
	sourceRoot := targetSourceRoot(interop)
	declaration, err := closuregraph.DomainID("swiftpm-interop-action-v1", map[string]any{"kind": string(interop.Kind), "languages": anyStrings(languageStrings(interop.Languages)), "package": interop.Package, "sources": anyStrings(interop.Sources), "target": interop.Target})
	if err != nil {
		return nil, nil, nil, err
	}
	objectNodes := make([]closuregraph.Node, len(interop.Sources))
	objectIDs := make([]closuregraph.ID, len(interop.Sources))
	for index, source := range interop.Sources {
		object := closuregraph.Node{Kind: closuregraph.NodeOutputArtifact, LogicalKey: fmt.Sprintf("swiftpm.interop.object.%s.%04d", key, index), Payload: closuregraph.OutputArtifactPayload{
			Profile: ProfileID, LogicalPath: ObjectLogicalPath(interop.Package, interop.Target, source), ExpectedClass: "native.object",
			OutputRole: "intermediate", CompatibilityPredicate: string(interop.Kind), DeclarationDigest: declaration,
			PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
		}}
		if err = object.Validate(); err != nil {
			return nil, nil, nil, err
		}
		objectNodes[index] = object
		if objectIDs[index], err = object.ID(); err != nil {
			return nil, nil, nil, err
		}
	}
	interop.ObjectNodeIDs = append([]closuregraph.ID(nil), objectIDs...)
	nodes = append(nodes, objectNodes...)
	edges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeDeclares, EdgeKey: "swiftpm.interop.declares-compile." + key, FromNodeID: interop.NodeID, ToNodeID: actionID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + interop.Target + ".compile"}}},
		{Kind: closuregraph.EdgeDeclares, EdgeKey: "swiftpm.interop.declares-sources." + key, FromNodeID: interop.NodeID, ToNodeID: sourceSetID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + interop.Target + ".sources"}}},
		{Kind: closuregraph.EdgeReads, EdgeKey: "swiftpm.interop.read-sources." + key, FromNodeID: actionID, ToNodeID: sourceSetID, Payload: closuregraph.ReadsPayload{Path: sourceRoot, ReadSlot: "sources", ReadClass: "source." + string(interop.Kind), Projection: append([]string(nil), interop.Sources...)}},
	}
	for index, source := range interop.Sources {
		edges = append(edges, closuregraph.Edge{
			Kind: closuregraph.EdgeProduces, EdgeKey: fmt.Sprintf("swiftpm.interop.produce-object.%s.%04d", key, index),
			FromNodeID: actionID, ToNodeID: objectIDs[index],
			Payload: closuregraph.ProducesPayload{Path: ObjectLogicalPath(interop.Package, interop.Target, source), WriteSlot: ObjectWriteSlot(index), WriteClass: "native.object"},
		})
	}
	if len(interop.Headers) != 0 {
		headerNode, headerID, headerErr := state.headerSourceSet(interop)
		if headerErr != nil {
			return nil, nil, nil, headerErr
		}
		interop.HeaderSetNodeID = headerID
		nodes = append(nodes, headerNode)
		edges = append(edges,
			closuregraph.Edge{Kind: closuregraph.EdgeDeclares, EdgeKey: "swiftpm.interop.declares-headers." + key, FromNodeID: interop.NodeID, ToNodeID: headerID, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + interop.Target + ".publicHeadersPath"}}},
			closuregraph.Edge{Kind: closuregraph.EdgeReads, EdgeKey: "swiftpm.interop.read-headers." + key, FromNodeID: actionID, ToNodeID: headerID, Payload: closuregraph.ReadsPayload{Path: interop.PublicHeaderRoot, ReadSlot: "headers", ReadClass: "source.header", Projection: headerProjection(interop)}})
	}
	if !interop.Selected {
		// The declaration is captured; only the exact selected subset gains a
		// concrete platform and toolchain overlay.
		return nodes, edges, nil, nil
	}
	bindingEdges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeUsesTool, EdgeKey: "swiftpm.interop.uses-compiler." + key, FromNodeID: actionID, ToNodeID: toolID, Payload: closuregraph.UsesToolPayload{ExecutableRelativePath: state.executablePath(toolRole), ToolSlot: "compiler", InvocationRole: subtype}},
		{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.interop.compile-target." + key, FromNodeID: actionID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
	}
	for index := range interop.Sources {
		bindingEdges = append(bindingEdges, closuregraph.Edge{
			Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("swiftpm.interop.object-target.%s.%04d", key, index),
			FromNodeID: objectIDs[index], ToNodeID: platformID,
			Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}},
		})
	}
	return nodes, edges, bindingEdges, nil
}

func (state *closeState) targetSourceSet(interop *TargetInterop) (closuregraph.Node, closuregraph.ID, error) {
	pkg := state.packages[interop.Package]
	values := make([]any, len(interop.Sources))
	for index, source := range interop.Sources {
		values[index] = source
	}
	treeDigest, err := closuregraph.DomainID("swiftpm-interop-source-set-v1", map[string]any{"package": interop.Package, "sources": values, "target": interop.Target, "tree_digest": string(pkg.SnapshotDigest)})
	if err != nil {
		return closuregraph.Node{}, "", err
	}
	node := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "swiftpm.interop.sources." + interop.Package + "." + interop.Target, Payload: closuregraph.SourceSetPayload{
		Profile: ProfileID, Origin: pkg.Origin, ArtifactManifestID: pkg.ArtifactManifestID, Projection: append([]string(nil), interop.Sources...),
		Grammar: "swiftpm-target-sources-v1", TrustRole: closuregraph.TrustDependencyInput, SourceClass: "source." + string(interop.Kind), TreeDigest: treeDigest,
	}}
	if err = node.Validate(); err != nil {
		return closuregraph.Node{}, "", err
	}
	id, err := node.ID()
	return node, id, err
}

func (state *closeState) headerSourceSet(interop *TargetInterop) (closuregraph.Node, closuregraph.ID, error) {
	pkg := state.packages[interop.Package]
	digestValues := make([]any, len(interop.Headers))
	for index, header := range interop.Headers {
		digestValues[index] = map[string]any{"path": header.Relative, "sha256": string(header.SHA256)}
	}
	treeDigest, err := closuregraph.DomainID("swiftpm-interop-header-set-v1", map[string]any{"headers": digestValues, "package": interop.Package, "target": interop.Target})
	if err != nil {
		return closuregraph.Node{}, "", err
	}
	node := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "swiftpm.interop.headers." + interop.Package + "." + interop.Target, Payload: closuregraph.SourceSetPayload{
		Profile: ProfileID, Origin: pkg.Origin, ArtifactManifestID: pkg.ArtifactManifestID, Projection: headerProjection(interop),
		Grammar: ModuleMapGrammarID, TrustRole: closuregraph.TrustDependencyInput, SourceClass: "source.header", TreeDigest: treeDigest,
	}}
	if err = node.Validate(); err != nil {
		return closuregraph.Node{}, "", err
	}
	id, err := node.ID()
	return node, id, err
}

func (state *closeState) boundaryRecords(boundary *Boundary, platformID closuregraph.ID, index int) ([]closuregraph.Node, []closuregraph.Edge, []closuregraph.Edge, error) {
	providerIndex, providerOK := state.byTargetKey[boundary.Provider]
	consumerIndex, consumerOK := state.byTargetKey[boundary.Consumer]
	if !providerOK || !consumerOK {
		return nil, nil, nil, failFields(CodeGraphReferenceInvalid, map[string]string{"provider": boundary.Provider, "consumer": boundary.Consumer}, "interop boundary names a target outside the selected closure")
	}
	provider, consumer := &state.targets[providerIndex], &state.targets[consumerIndex]
	contract, err := closuregraph.DomainID("swiftpm-interop-contract-v1", map[string]any{"abi": boundary.ABI, "consumer": boundary.Consumer, "interface_contract": boundary.InterfaceContract, "mode": string(boundary.Mode), "provider": boundary.Provider, "runtime": boundary.Runtime})
	if err != nil {
		return nil, nil, nil, err
	}
	providerClasses := languageStrings(boundary.ProviderLanguages)
	node := closuregraph.Node{Kind: closuregraph.NodeInteropBoundary, LogicalKey: fmt.Sprintf("swiftpm.interop.boundary.%04d.%s", index, boundary.Provider+"->"+boundary.Consumer), Payload: closuregraph.InteropBoundaryPayload{
		Profile: ProfileID, Mode: boundary.Mode, ProviderLanguageClasses: providerClasses, ConsumerLanguageClasses: languageStrings(boundary.ConsumerLanguages),
		ContractDigest: contract, ABI: boundary.ABI, Runtime: boundary.Runtime, InterfaceContract: boundary.InterfaceContract,
		CallingConvention: boundary.CallingConvention, LinkLoadSemantics: boundary.LinkLoadSemantics,
		PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}}
	if err = node.Validate(); err != nil {
		return nil, nil, nil, failFields(CodeInteropUndeclared, map[string]string{"boundary": boundary.Provider + "->" + boundary.Consumer}, "interop boundary declaration is incomplete: %v", err)
	}
	boundaryID, err := node.ID()
	if err != nil {
		return nil, nil, nil, err
	}
	boundary.NodeID = boundaryID
	evidence := []closuregraph.ID{}
	if provider.ModuleMap != nil {
		evidence = append(evidence, provider.ModuleMap.SHA256)
	}
	for _, header := range provider.Headers {
		evidence = append(evidence, header.SHA256)
	}
	evidence = sortedUniqueIDs(evidence)
	if len(evidence) == 0 {
		return nil, nil, nil, failFields(CodeInteropUndeclared, map[string]string{"provider": boundary.Provider}, "interop provider has no immutable interface evidence")
	}
	provides := closuregraph.Edge{Kind: closuregraph.EdgeProvidesInterop, EdgeKey: "swiftpm.interop.provides." + boundary.Provider + "->" + boundary.Consumer, ToNodeID: boundaryID, Payload: closuregraph.ProvidesInteropPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + provider.Target + ".publicHeadersPath"}, EvidenceIDs: evidence, ExportRole: "headers", LinkMode: "static"}}
	consumes := closuregraph.Edge{Kind: closuregraph.EdgeConsumesInterop, EdgeKey: "swiftpm.interop.consumes." + boundary.Provider + "->" + boundary.Consumer, FromNodeID: consumer.ActionNodeID, ToNodeID: boundaryID, Payload: closuregraph.ConsumesInteropPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + consumer.Target + ".dependencies"}, Use: "compile", ABIExpectation: boundary.ABI, Condition: boundary.Condition}}
	toolchain, present := state.byRole[boundary.ToolchainRole]
	if !present {
		return nil, nil, nil, failFields(CodeToolchainUntrusted, map[string]string{"role": boundary.ToolchainRole}, "interop boundary names no selected toolchain component")
	}
	if !boundary.Selected {
		// A pruned boundary is a captured declaration only; binding a concrete
		// platform or toolchain to it would claim a selection that never happens.
		if provider.Kind == KindSystem {
			return []closuregraph.Node{node}, []closuregraph.Edge{consumes}, nil, nil
		}
		provides.FromNodeID = provider.NodeID
		return []closuregraph.Node{node}, []closuregraph.Edge{provides, consumes}, nil, nil
	}
	bindingEdges := []closuregraph.Edge{
		{Kind: closuregraph.EdgeTargets, EdgeKey: "swiftpm.interop.boundary-target." + boundary.Provider + "->" + boundary.Consumer, FromNodeID: boundaryID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}},
		{Kind: closuregraph.EdgeRequires, EdgeKey: "swiftpm.interop.boundary-toolchain." + boundary.Provider + "->" + boundary.Consumer, FromNodeID: boundaryID, ToNodeID: toolchain.nodeID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeToolchain, Origin: closuregraph.EvidenceOrigin{Field: "selection.interop_toolchains." + boundary.ToolchainRole}, DependencyKind: "external-toolchain"}},
	}
	if provider.Kind == KindSystem {
		provides.FromNodeID = provider.ToolNodeID
		return []closuregraph.Node{node}, []closuregraph.Edge{consumes}, append(bindingEdges, provides), nil
	}
	provides.FromNodeID = provider.NodeID
	return []closuregraph.Node{node}, []closuregraph.Edge{provides, consumes}, bindingEdges, nil
}

func (state *closeState) bindingAuthority() (closuregraph.BindingAuthority, error) {
	authority := closuregraph.BindingAuthority{C0Checkpoint: state.capture.Authority.C0Checkpoint, Toolchains: append([]closuregraph.ToolchainBindingEvidence(nil), state.capture.Authority.Toolchains...), C4Selectors: append([]closuregraph.ToolchainSelector(nil), state.capture.Authority.C4Selectors...)}
	for _, binding := range state.components {
		selector, err := closuregraph.NewToolchainSelector(binding.node)
		if err != nil {
			return closuregraph.BindingAuthority{}, err
		}
		selectorID, err := selector.ID()
		if err != nil {
			return closuregraph.BindingAuthority{}, err
		}
		authority.C4Selectors = append(authority.C4Selectors, selector)
		authority.Toolchains = append(authority.Toolchains, closuregraph.ToolchainBindingEvidence{NodeID: binding.nodeID, FirstBound: closuregraph.ToolchainBoundAtC4, EvidenceID: selectorID})
	}
	sort.Slice(authority.Toolchains, func(i, j int) bool { return authority.Toolchains[i].NodeID < authority.Toolchains[j].NodeID })
	sort.Slice(authority.C4Selectors, func(i, j int) bool { return authority.C4Selectors[i].NodeID < authority.C4Selectors[j].NodeID })
	return authority, nil
}

func (state *closeState) platformNodeID() (closuregraph.ID, error) {
	id, ok := state.capture.Selection.PlatformRoles[closuregraph.PlatformTarget]
	if !ok || !id.Valid() {
		return "", fail(CodeGraphIncomplete, "accepted selection binds no exact target platform")
	}
	return id, nil
}

func (state *closeState) executablePath(role string) string {
	if binding, ok := state.byRole[role]; ok {
		return binding.component.ExecutableRelativePath
	}
	return state.capture.SelectionToolchain().Swift.ExecutableRelativePath
}

func (state *closeState) evaluator() closuregraph.ConditionEvaluator {
	markers := state.markers
	return closuregraph.ConditionEvaluatorFunc{EvaluatorID: swiftpmsource.ConditionEvaluatorID, EvaluateFunc: func(condition closuregraph.Condition, _ closuregraph.EvaluationInput) (bool, error) {
		if condition.EvaluatorID != swiftpmsource.ConditionEvaluatorID {
			return false, fail(CodeGraphReferenceInvalid, "wrong SwiftPM condition evaluator")
		}
		parts := strings.Split(condition.Expression, "=")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition %q", condition.Expression)
		}
		switch parts[0] {
		case "platform", "configuration", "architecture":
			return strings.EqualFold(markers[parts[0]], parts[1]), nil
		case "trait":
			return markers["trait:"+parts[1]] == "true", nil
		default:
			return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition key %q", parts[0])
		}
	}}
}

func (state *closeState) evidenceDigest(boundaries []Boundary, reads ReadSetEvidence, graphID, bindingID closuregraph.ID) (closuregraph.ID, error) {
	targets := make([]any, len(state.targets))
	for index, interop := range state.targets {
		moduleMap := ""
		moduleMapDigest := ""
		if interop.ModuleMap != nil {
			moduleMap, moduleMapDigest = interop.ModuleMap.Relative, string(interop.ModuleMap.SHA256)
		}
		headers := make([]any, len(interop.Headers))
		for headerIndex, header := range interop.Headers {
			headers[headerIndex] = map[string]any{"path": header.Relative, "sha256": string(header.SHA256)}
		}
		targets[index] = map[string]any{"cxx_interop": interop.CxxInteropMode, "headers": headers, "kind": string(interop.Kind), "languages": anyStrings(languageStrings(interop.Languages)), "module_map": moduleMap, "module_map_sha256": moduleMapDigest, "package": interop.Package, "public_header_root": interop.PublicHeaderRoot, "selected": interop.Selected, "sources": anyStrings(interop.Sources), "target": interop.Target}
	}
	declared := make([]any, len(boundaries))
	for index, boundary := range boundaries {
		var condition any
		if boundary.Condition != nil {
			condition = map[string]any{"evaluator_id": boundary.Condition.EvaluatorID, "expression": boundary.Condition.Expression}
		}
		declared[index] = map[string]any{"abi": boundary.ABI, "condition": condition, "consumer": boundary.Consumer, "mode": string(boundary.Mode), "node_id": string(boundary.NodeID), "provider": boundary.Provider, "runtime": boundary.Runtime, "selected": boundary.Selected, "toolchain_role": boundary.ToolchainRole}
	}
	components := make([]any, len(state.components))
	for index, binding := range state.components {
		components[index] = map[string]any{"fingerprint": string(binding.component.Fingerprint), "node_id": string(binding.nodeID), "role": binding.component.Role}
	}
	receipts := make([]any, len(reads.ReceiptIDs))
	for index, receipt := range reads.ReceiptIDs {
		receipts[index] = string(receipt)
	}
	return closuregraph.DomainID(InteropSchemaID, map[string]any{
		"boundaries": declared, "components": components, "capture_graph_id": string(graphID),
		"profile": state.config.Profile.ID, "read_mode": reads.Mode, "read_receipt_ids": receipts,
		"selection_binding_id": string(bindingID), "targets": targets,
	})
}

func bindingNodeID(nodes []closuregraph.Node, logicalKey string) (closuregraph.ID, bool) {
	for _, node := range nodes {
		if node.LogicalKey != logicalKey {
			continue
		}
		id, err := node.ID()
		if err != nil {
			return "", false
		}
		return id, true
	}
	return "", false
}

func headerProjection(interop *TargetInterop) []string {
	values := make([]string, len(interop.Headers))
	for index, header := range interop.Headers {
		values[index] = header.Relative
	}
	sort.Strings(values)
	return values
}

// targetSourceRoot derives the exact declared read root of one capture target
// from its admitted source projection.
func targetSourceRoot(interop *TargetInterop) string {
	if len(interop.Sources) == 0 {
		return "Sources"
	}
	common := path.Dir(interop.Sources[0])
	for _, source := range interop.Sources[1:] {
		for common != "." && !strings.HasPrefix(source+"/", common+"/") {
			common = path.Dir(common)
		}
	}
	if common == "." || common == "" {
		return interop.Sources[0]
	}
	return common
}

func anyStrings(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func sortedUniqueIDs(values []closuregraph.ID) []closuregraph.ID {
	sorted := append([]closuregraph.ID(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	result := sorted[:0]
	for _, value := range sorted {
		if value.Valid() && (len(result) == 0 || result[len(result)-1] != value) {
			result = append(result, value)
		}
	}
	return append([]closuregraph.ID{}, result...)
}

// ObjectWriteSlot names the compile action's write slot for the source at the
// given position in the target's ordered source list. SwiftPM's native build
// system emits one object per source file, so the declared write set is one
// slot per source rather than one anonymous slot per target. The downstream
// build stage binds each slot to the exact produced object it observes; this
// stage never compiles and therefore declares no produced bytes itself.
func ObjectWriteSlot(index int) string {
	return fmt.Sprintf("objects.%04d", index)
}

// ObjectLogicalPath is the exact declared logical path of the object the
// compile action produces for one source file. It keeps the package-relative
// source path so two sources sharing a base name never collide.
func ObjectLogicalPath(pkg, target, source string) string {
	return path.Join(".curator", "objects", pkg, target, source+".o")
}
