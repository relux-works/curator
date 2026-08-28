package closuregraph

import (
	"fmt"
	"sort"
	"strings"
)

// RecordTables are the graph-local authorities named by capture and binding
// envelopes. Constructors canonicalize their record order; validation rejects
// direct noncanonical tables.
type RecordTables struct {
	CaptureNodes []Node
	CaptureEdges []Edge
	BindingNodes []Node
	BindingEdges []Edge
}

// NewRecordTables copies and canonically orders all four record tables.
func NewRecordTables(captureNodes []Node, captureEdges []Edge, bindingNodes []Node, bindingEdges []Edge) RecordTables {
	return RecordTables{CaptureNodes: sortNodes(captureNodes), CaptureEdges: sortEdges(captureEdges), BindingNodes: sortNodes(bindingNodes), BindingEdges: sortEdges(bindingEdges)}
}

// ToolchainBindingStage identifies the only stages that may establish an
// external toolchain's authority.
type ToolchainBindingStage string

const (
	// ToolchainBoundAtC0 and ToolchainBoundAtC4 are the only stages that may
	// establish binding authority for an external toolchain.
	ToolchainBoundAtC0 ToolchainBindingStage = "C0.profile"
	// ToolchainBoundAtC4 identifies a selected build-only toolchain binding.
	ToolchainBoundAtC4 ToolchainBindingStage = "C4.close"
)

// ToolchainBindingEvidence proves that one binding toolchain was either
// prebound for evidence derivation at C0 or selected and fingerprinted for
// build use at C4.
type ToolchainBindingEvidence struct {
	NodeID     ID
	FirstBound ToolchainBindingStage
	EvidenceID ID
}

// ToolchainSelector is the C4 authority record for one selected build-only
// external tool. Its identity repeats the security-relevant node fields so an
// arbitrary digest cannot be presented as selector evidence.
type ToolchainSelector struct {
	NodeID                 ID
	PolicySelector         string
	ContentFingerprint     ID
	ExecutableRelativePath string
}

// NewToolchainSelector creates a selector from an exact binding toolchain.
func NewToolchainSelector(node Node) (ToolchainSelector, error) {
	if err := node.Validate(); err != nil {
		return ToolchainSelector{}, err
	}
	if node.Kind != NodeToolchainComponent {
		return ToolchainSelector{}, fmt.Errorf("%s: C4 selector requires a toolchain_component", CodeGraphReferenceInvalid)
	}
	nodeID, err := node.ID()
	if err != nil {
		return ToolchainSelector{}, err
	}
	payload := node.Payload.(ToolchainComponentPayload)
	selector := ToolchainSelector{
		NodeID:                 nodeID,
		PolicySelector:         payload.PolicySelector,
		ContentFingerprint:     payload.ContentFingerprint,
		ExecutableRelativePath: payload.ExecutableRelativePath,
	}
	return selector, selector.Validate()
}

// Validate checks the closed C4 selector record.
func (selector ToolchainSelector) Validate() error {
	if err := validateID(selector.NodeID, "toolchain selector node_id"); err != nil {
		return err
	}
	if err := validatePortableText(selector.PolicySelector, "toolchain selector policy_selector", false); err != nil {
		return err
	}
	if err := validateID(selector.ContentFingerprint, "toolchain selector content_fingerprint"); err != nil {
		return err
	}
	return validatePortablePath(selector.ExecutableRelativePath, "toolchain selector executable_relative_path")
}

// ID derives the domain-separated selector identity used by C4 authority.
func (selector ToolchainSelector) ID() (ID, error) {
	if err := selector.Validate(); err != nil {
		return "", err
	}
	return DomainID(LabelToolchainSelector, map[string]any{
		"content_fingerprint":      string(selector.ContentFingerprint),
		"executable_relative_path": selector.ExecutableRelativePath,
		"node_id":                  string(selector.NodeID),
		"policy_selector":          selector.PolicySelector,
	})
}

// BindingAuthority is the closed external trust evidence for binding nodes.
type BindingAuthority struct {
	Toolchains   []ToolchainBindingEvidence
	C0Checkpoint *Checkpoint
	C4Selectors  []ToolchainSelector
}

// GraphBundle is the complete C4 graph evidence: selection-neutral capture,
// exact selection and binding, active projection, and local record tables.
type GraphBundle struct {
	Capture   CaptureGraph
	Selection SelectionContext
	Binding   SelectionBinding
	Active    ActiveGraph
	Records   RecordTables
	Authority BindingAuthority
}

type resolvedTables struct {
	captureNodes map[ID]Node
	bindingNodes map[ID]Node
	allNodes     map[ID]Node
	captureEdges map[ID]Edge
	bindingEdges map[ID]Edge
	allEdges     map[ID]Edge
}

// Validate proves the complete capture/context/binding/active contract.
func (bundle GraphBundle) Validate() error {
	resolved, err := validateStructure(bundle.Capture, bundle.Selection, bundle.Binding, bundle.Records, bundle.Authority)
	if err != nil {
		return err
	}
	return validateActive(bundle, resolved)
}

func validateStructure(capture CaptureGraph, selection SelectionContext, binding SelectionBinding, records RecordTables, authority BindingAuthority) (resolvedTables, error) {
	collector := &issueCollector{}
	if err := capture.Validate(); err != nil {
		collector.add(CodeGraphSchemaUnsupported, "capture", "capture", "capture", "%v", err)
	}
	if err := selection.Validate(); err != nil {
		collector.add(CodeGraphSchemaUnsupported, "selection", "selection", "selection", "%v", err)
	}
	if err := binding.Validate(); err != nil {
		collector.add(CodeGraphSchemaUnsupported, "binding", "binding", "binding", "%v", err)
	}

	resolved := resolvedTables{captureNodes: map[ID]Node{}, bindingNodes: map[ID]Node{}, allNodes: map[ID]Node{}, captureEdges: map[ID]Edge{}, bindingEdges: map[ID]Edge{}, allEdges: map[ID]Edge{}}
	logicalKeys := map[string][]string{}
	nodeIDs := map[ID][]string{}

	if !sameNodeOrder(records.CaptureNodes, sortNodes(records.CaptureNodes)) {
		collector.add(CodeGraphReferenceInvalid, "capture.nodes", "capture", "capture nodes", "record table is not canonical (kind, logical_key, node_id) order")
	}
	if !sameNodeOrder(records.BindingNodes, sortNodes(records.BindingNodes)) {
		collector.add(CodeGraphReferenceInvalid, "binding.nodes", "binding", "binding nodes", "record table is not canonical (kind, logical_key, node_id) order")
	}

	consumeNodes := func(table string, nodes []Node, destination map[ID]Node) {
		for index, node := range nodes {
			path := fmt.Sprintf("%s.nodes[%d]", table, index)
			if err := node.Validate(); err != nil {
				code := CodeGraphSchemaUnsupported
				if node.Kind == NodeInteropBoundary {
					code = CodeInteropUndeclared
				}
				collector.add(code, node.LogicalKey, table, path, "%v", err)
				continue
			}
			id, err := node.ID()
			if err != nil {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey, table, path, "derive node ID: %v", err)
				continue
			}
			origin := table + ":" + node.LogicalKey
			logicalKeys[node.LogicalKey] = append(logicalKeys[node.LogicalKey], origin)
			nodeIDs[id] = append(nodeIDs[id], origin)
			if _, exists := destination[id]; !exists {
				destination[id] = node
			}
			if _, exists := resolved.allNodes[id]; !exists {
				resolved.allNodes[id] = node
			}
			if table == "capture" && (node.Kind == NodeTargetPlatform || node.Kind == NodeToolchainComponent) {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey, table, path, "selection-specific node kind %q is forbidden in capture", node.Kind)
			}
			if table == "binding" && node.Kind != NodeTargetPlatform && node.Kind != NodeToolchainComponent {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey, table, path, "binding node kind %q is forbidden", node.Kind)
			}
		}
	}
	consumeNodes("capture", records.CaptureNodes, resolved.captureNodes)
	consumeNodes("binding", records.BindingNodes, resolved.bindingNodes)
	for key, origins := range logicalKeys {
		if len(origins) > 1 {
			sort.Strings(origins)
			collector.add(CodeGraphReferenceInvalid, key, "capture+binding", "logical_key", "duplicate logical key in %s", strings.Join(origins, ", "))
		}
	}
	for id, origins := range nodeIDs {
		if len(origins) > 1 {
			sort.Strings(origins)
			collector.add(CodeGraphReferenceInvalid, string(id), "capture+binding", "node_id", "duplicate node ID in %s", strings.Join(origins, ", "))
		}
	}

	checkReferencedIDs(collector, "capture", "node", capture.NodeIDs, resolved.captureNodes)
	checkReferencedIDs(collector, "binding", "node", binding.BindingNodeIDs, resolved.bindingNodes)
	rootSet := idSet(capture.NodeIDs)
	for _, root := range capture.RootNodeIDs {
		if !rootSet[root] {
			collector.add(CodeGraphReferenceInvalid, string(root), "capture", "root_node_ids", "root does not resolve in capture node table")
		}
	}
	for _, productID := range selection.ProductNodeIDs {
		node, ok := resolved.captureNodes[productID]
		if !ok {
			collector.add(CodeGraphReferenceInvalid, string(productID), "selection", "product_node_ids", "selected product does not resolve in capture")
		} else if node.Kind != NodeCommandProduct {
			collector.add(CodeGraphReferenceInvalid, string(productID), "selection", "product_node_ids", "selected product resolves to %q, want command_product", node.Kind)
		}
	}

	if !sameEdgeOrder(records.CaptureEdges, sortEdges(records.CaptureEdges)) {
		collector.add(CodeGraphReferenceInvalid, "capture.edges", "capture", "capture edges", "record table is not canonical (kind, edge_key, edge_id) order")
	}
	if !sameEdgeOrder(records.BindingEdges, sortEdges(records.BindingEdges)) {
		collector.add(CodeGraphReferenceInvalid, "binding.edges", "binding", "binding edges", "record table is not canonical (kind, edge_key, edge_id) order")
	}
	edgeKeys := map[string][]string{}
	edgeIDs := map[ID][]string{}
	semantics := map[string][]string{}
	consumeEdges := func(table string, edges []Edge, destination map[ID]Edge) {
		for index, edge := range edges {
			path := fmt.Sprintf("%s.edges[%d]", table, index)
			if err := edge.Validate(); err != nil {
				collector.add(CodeGraphSchemaUnsupported, edge.EdgeKey, table, path, "%v", err)
				continue
			}
			id, err := edge.ID()
			if err != nil {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "derive edge ID: %v", err)
				continue
			}
			origin := table + ":" + edge.EdgeKey
			edgeKeys[edge.EdgeKey] = append(edgeKeys[edge.EdgeKey], origin)
			edgeIDs[id] = append(edgeIDs[id], origin)
			semantic, semanticErr := semanticEdgeKey(edge)
			if semanticErr != nil {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "derive semantic edge identity: %v", semanticErr)
			} else {
				evidence, evidenceErr := semanticEdgeEvidence(table, edge, id)
				if evidenceErr != nil {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "derive semantic edge evidence: %v", evidenceErr)
				} else {
					semantics[semantic] = append(semantics[semantic], evidence)
				}
			}
			if _, exists := destination[id]; !exists {
				destination[id] = edge
			}
			if _, exists := resolved.allEdges[id]; !exists {
				resolved.allEdges[id] = edge
			}
			from, fromOK := resolved.allNodes[edge.FromNodeID]
			to, toOK := resolved.allNodes[edge.ToNodeID]
			if !fromOK {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path+".from_node_id", "dangling endpoint %s", edge.FromNodeID)
			}
			if !toOK {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path+".to_node_id", "dangling endpoint %s", edge.ToNodeID)
			}
			if table == "capture" {
				if _, ok := resolved.captureNodes[edge.FromNodeID]; !ok {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "capture edge from endpoint is not capture-owned")
				}
				if _, ok := resolved.captureNodes[edge.ToNodeID]; !ok {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "capture edge to endpoint is not capture-owned")
				}
				if edge.Kind == EdgeTargets {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "concrete targets edge is forbidden in capture")
				}
			}
			if table == "binding" {
				if !allowedBindingEdgeKind(edge.Kind) {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "binding edge kind %q is forbidden", edge.Kind)
				}
				if edge.Kind == EdgeRequires && edge.Payload.(RequiresPayload).Scope != ScopeToolchain {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "binding requires edge must have toolchain scope")
				}
				_, fromBinding := resolved.bindingNodes[edge.FromNodeID]
				_, toBinding := resolved.bindingNodes[edge.ToNodeID]
				if !fromBinding && !toBinding {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "binding edge replaces capture semantics without a binding endpoint")
				}
			}
			if fromOK && toOK && !allowedEndpoints(edge.Kind, from.Kind, to.Kind) {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, table, path, "wrong-kind endpoints %q -> %q for %q", from.Kind, to.Kind, edge.Kind)
			}
		}
	}
	consumeEdges("capture", records.CaptureEdges, resolved.captureEdges)
	consumeEdges("binding", records.BindingEdges, resolved.bindingEdges)
	validateDeclaredEdgeSlots(collector, resolved.allEdges, resolved)
	validateEndpointContracts(collector, resolved.allEdges, resolved)
	for key, origins := range edgeKeys {
		if len(origins) > 1 {
			sort.Strings(origins)
			collector.add(CodeGraphReferenceInvalid, key, "capture+binding", "edge_key", "duplicate edge key in %s", strings.Join(origins, ", "))
		}
	}
	for id, origins := range edgeIDs {
		if len(origins) > 1 {
			sort.Strings(origins)
			collector.add(CodeGraphReferenceInvalid, string(id), "capture+binding", "edge_id", "duplicate edge ID in %s", strings.Join(origins, ", "))
		}
	}
	for semantic, origins := range semantics {
		if len(origins) > 1 {
			sort.Strings(origins)
			collector.add(CodeGraphReferenceInvalid, semantic, "capture+binding", "semantic edge", "duplicate semantic edge in %s", strings.Join(origins, ", "))
		}
	}
	checkReferencedIDs(collector, "capture", "edge", capture.EdgeIDs, resolved.captureEdges)
	checkReferencedIDs(collector, "binding", "edge", binding.BindingEdgeIDs, resolved.bindingEdges)
	selectedEvaluators := make(map[string]bool, len(selection.EvaluatorIDs))
	for _, evaluatorID := range selection.EvaluatorIDs {
		selectedEvaluators[evaluatorID] = true
	}
	for id, edge := range resolved.captureEdges {
		condition := edge.Payload.condition()
		if condition != nil && !selectedEvaluators[condition.EvaluatorID] {
			collector.add(CodeGraphIncomplete, string(id), "capture", "condition.evaluator_id", "conditional edge requires evaluator %q absent from selection", condition.EvaluatorID)
		}
	}

	captureID, captureErr := capture.ID()
	if captureErr == nil && binding.CapturedGraphID != captureID {
		collector.add(CodeGraphReferenceInvalid, string(binding.CapturedGraphID), "binding", "captured_graph_id", "capture replacement: expected %s", captureID)
	}
	selectionID, selectionErr := selection.ID()
	if selectionErr == nil && binding.SelectionContextID != selectionID {
		collector.add(CodeGraphReferenceInvalid, string(binding.SelectionContextID), "binding", "selection_context_id", "expected %s", selectionID)
	}
	validatePlatformRoleRecords(collector, selection, resolved)
	validateToolchainAuthority(collector, records.BindingNodes, authority)
	validateArtifactManifestRefs(collector, capture, records.CaptureNodes, resolved.captureEdges)

	if err := collector.err(); err != nil {
		return resolvedTables{}, err
	}
	return resolved, nil
}

func validateActive(bundle GraphBundle, resolved resolvedTables) error {
	collector := &issueCollector{}
	if err := bundle.Active.Validate(); err != nil {
		collector.add(CodeGraphSchemaUnsupported, "active", "active", "active", "%v", err)
	}
	captureID, _ := bundle.Capture.ID()
	selectionID, _ := bundle.Selection.ID()
	bindingID, _ := bundle.Binding.ID()
	if bundle.Active.CapturedGraphID != captureID {
		collector.add(CodeGraphReferenceInvalid, string(bundle.Active.CapturedGraphID), "active", "captured_graph_id", "expected %s", captureID)
	}
	if bundle.Active.SelectionContextID != selectionID {
		collector.add(CodeGraphReferenceInvalid, string(bundle.Active.SelectionContextID), "active", "selection_context_id", "expected %s", selectionID)
	}
	if bundle.Active.SelectionBindingID != bindingID {
		collector.add(CodeGraphReferenceInvalid, string(bundle.Active.SelectionBindingID), "active", "selection_binding_id", "expected %s", bindingID)
	}
	nodeStates := map[ID]ActivationState{}
	for _, activation := range bundle.Active.NodeActivations {
		if _, ok := resolved.captureNodes[activation.NodeID]; !ok {
			collector.add(CodeGraphReferenceInvalid, string(activation.NodeID), "active", "node_activations", "activation does not reference capture node")
		}
		nodeStates[activation.NodeID] = activation.State
	}
	for id := range resolved.captureNodes {
		if _, ok := nodeStates[id]; !ok {
			collector.add(CodeGraphIncomplete, string(id), "active", "node_activations", "capture node has no activation")
		}
	}
	edgeStates := map[ID]EdgeActivation{}
	for _, activation := range bundle.Active.EdgeActivations {
		edge, ok := resolved.captureEdges[activation.EdgeID]
		if !ok {
			collector.add(CodeGraphReferenceInvalid, string(activation.EdgeID), "active", "edge_activations", "activation does not reference capture edge")
			continue
		}
		if edge.Payload.condition() == nil {
			collector.add(CodeGraphReferenceInvalid, string(activation.EdgeID), "active", "edge_activations", "unconditional edge must not have an activation record")
		}
		edgeStates[activation.EdgeID] = activation
	}
	for id, edge := range resolved.captureEdges {
		if edge.Payload.condition() != nil {
			if _, ok := edgeStates[id]; !ok {
				collector.add(CodeGraphIncomplete, string(id), "active", "edge_activations", "conditional capture edge has no activation")
			}
		}
	}
	usable := make(map[ID]Edge, len(resolved.allEdges))
	for id, edge := range resolved.captureEdges {
		activation, conditional := edgeStates[id]
		if !conditional || activation.Evaluation {
			usable[id] = edge
		}
	}
	for id, edge := range resolved.bindingEdges {
		usable[id] = edge
	}
	reachable := selectionReachability(bundle.Selection.ProductNodeIDs, usable)
	for id, node := range resolved.captureNodes {
		want := ActivationPruned
		if reachable[id] {
			want = ActivationSelected
		}
		if nodeStates[id] != want {
			collector.add(CodeGraphReferenceInvalid, node.LogicalKey, "active", "node_activations", "activation is %q, want %q from selected reachability", nodeStates[id], want)
		}
	}
	for id, node := range resolved.bindingNodes {
		if !reachable[id] {
			collector.add(CodeGraphIncomplete, node.LogicalKey, "active", "binding", "binding node %s is unreachable from selected products", id)
		}
	}
	for id, activation := range edgeStates {
		edge := resolved.captureEdges[id]
		wantState, wantReason := ActivationPruned, ReasonConditionFalse
		if activation.Evaluation {
			wantReason = ReasonUnreachable
			if reachable[edge.FromNodeID] && reachable[edge.ToNodeID] {
				wantState, wantReason = ActivationSelected, ReasonConditionTrue
			}
		}
		if activation.State != wantState || activation.Reason != wantReason {
			collector.add(CodeGraphReferenceInvalid, string(id), "active", "edge_activations", "activation state/reason is %q/%q, want %q/%q", activation.State, activation.Reason, wantState, wantReason)
		}
	}
	selectedEdges := selectedEdges(bundle.Active, resolved)
	for _, edge := range selectedEdges {
		fromSelected := nodeStates[edge.FromNodeID] == ActivationSelected
		if _, binding := resolved.bindingNodes[edge.FromNodeID]; binding {
			fromSelected = reachable[edge.FromNodeID]
		}
		toSelected := nodeStates[edge.ToNodeID] == ActivationSelected
		if _, binding := resolved.bindingNodes[edge.ToNodeID]; binding {
			toSelected = reachable[edge.ToNodeID]
		}
		if !fromSelected || !toSelected {
			collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "selected edge", "selected edge has a pruned endpoint")
		}
	}
	validateActionSlots(collector, nodeStates, selectedEdges, resolved)
	validatePlatformBindings(collector, nodeStates, selectedEdges, resolved)
	validateProducedLineage(collector, nodeStates, selectedEdges, resolved)
	validateInteropBoundaries(collector, nodeStates, selectedEdges, resolved)
	wantSCCs := deriveNonOrderingSCCs(selectedEdges, resolved.allNodes)
	if !sameSCCs(bundle.Active.NonOrderingSCCs, wantSCCs) {
		collector.add(CodeGraphReferenceInvalid, "non_ordering_sccs", "active", "non_ordering_sccs", "records do not match the canonical selected non-ordering projection")
	}
	return collector.err()
}

func validatePlatformRoleRecords(collector *issueCollector, selection SelectionContext, resolved resolvedTables) {
	for role, id := range selection.PlatformRoles {
		node, ok := resolved.bindingNodes[id]
		if !ok {
			collector.add(CodeGraphReferenceInvalid, string(id), "selection", "platform_roles."+string(role), "platform role does not resolve exactly once in binding")
		} else if node.Kind != NodeTargetPlatform {
			collector.add(CodeGraphReferenceInvalid, string(id), "selection", "platform_roles."+string(role), "platform role resolves to %q", node.Kind)
		}
	}
	for _, edge := range resolved.bindingEdges {
		if edge.Kind != EdgeTargets {
			continue
		}
		payload := edge.Payload.(TargetsPayload)
		want, ok := platformIDForRole(selection, payload.BindingRole)
		if !ok {
			collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "binding", "targets.binding_role", "role %q is absent from selection", payload.BindingRole)
		} else if edge.ToNodeID != want {
			collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "binding", "targets.to_node_id", "role %q targets %s, expected %s", payload.BindingRole, edge.ToNodeID, want)
		}
	}
}

func validateToolchainAuthority(collector *issueCollector, bindingNodes []Node, authority BindingAuthority) {
	nodesByID := map[ID]Node{}
	for _, node := range bindingNodes {
		if node.Kind == NodeToolchainComponent {
			id, _ := node.ID()
			nodesByID[id] = node
		}
	}
	selectorByNode := map[ID][]ToolchainSelector{}
	for index, selector := range authority.C4Selectors {
		if err := selector.Validate(); err != nil {
			collector.add(CodeGraphReferenceInvalid, string(selector.NodeID), "authority", fmt.Sprintf("c4_selectors[%d]", index), "%v", err)
			continue
		}
		selectorByNode[selector.NodeID] = append(selectorByNode[selector.NodeID], selector)
		node, ok := nodesByID[selector.NodeID]
		if !ok {
			collector.add(CodeGraphReferenceInvalid, string(selector.NodeID), "authority", fmt.Sprintf("c4_selectors[%d]", index), "selector names no binding toolchain node")
			continue
		}
		payload := node.Payload.(ToolchainComponentPayload)
		if selector.PolicySelector != payload.PolicySelector || selector.ContentFingerprint != payload.ContentFingerprint || selector.ExecutableRelativePath != payload.ExecutableRelativePath {
			collector.add(CodeGraphReferenceInvalid, string(selector.NodeID), "authority", fmt.Sprintf("c4_selectors[%d]", index), "selector does not match the exact toolchain node fingerprint/path/policy")
		}
	}
	var c0ID ID
	var c0Tools map[ID]bool
	if authority.C0Checkpoint != nil {
		if err := authority.C0Checkpoint.Validate(); err != nil {
			collector.add(CodeCheckpointInvalid, "C0.profile", "authority", "c0_checkpoint", "%v", err)
		} else if authority.C0Checkpoint.Name != CheckpointC0 {
			collector.add(CodeCheckpointInvalid, "C0.profile", "authority", "c0_checkpoint", "authority record is %s, want C0.profile", authority.C0Checkpoint.Name)
		} else {
			c0ID, _ = authority.C0Checkpoint.ID()
			c0Tools = idSet(authority.C0Checkpoint.Payload.(C0ProfilePayload).EvidenceToolchainNodeIDs)
		}
	}
	byNode := map[ID][]ToolchainBindingEvidence{}
	for index, evidence := range authority.Toolchains {
		if err := validateID(evidence.NodeID, fmt.Sprintf("toolchain authority[%d].node_id", index)); err != nil {
			collector.add(CodeGraphReferenceInvalid, string(evidence.NodeID), "authority", "toolchains", "%v", err)
		}
		if evidence.FirstBound != ToolchainBoundAtC0 && evidence.FirstBound != ToolchainBoundAtC4 {
			collector.add(CodeGraphReferenceInvalid, string(evidence.NodeID), "authority", "toolchains", "unsupported first binding stage %q", evidence.FirstBound)
		}
		if err := validateID(evidence.EvidenceID, fmt.Sprintf("toolchain authority[%d].evidence_id", index)); err != nil {
			collector.add(CodeGraphReferenceInvalid, string(evidence.NodeID), "authority", "toolchains", "%v", err)
		}
		switch evidence.FirstBound {
		case ToolchainBoundAtC0:
			if authority.C0Checkpoint == nil {
				collector.add(CodeCheckpointInvalid, string(evidence.NodeID), "authority", "toolchains", "C0-bound toolchain has no exact C0 checkpoint authority")
			} else if evidence.EvidenceID != c0ID || !c0Tools[evidence.NodeID] {
				collector.add(CodeCheckpointInvalid, string(evidence.NodeID), "authority", "toolchains", "C0-bound evidence does not resolve to the exact C0 checkpoint and tool table")
			}
		case ToolchainBoundAtC4:
			selectors := selectorByNode[evidence.NodeID]
			if len(selectors) != 1 {
				collector.add(CodeGraphReferenceInvalid, string(evidence.NodeID), "authority", "toolchains", "C4-bound toolchain requires exactly one selector, got %d", len(selectors))
			} else {
				selectorID, _ := selectors[0].ID()
				if evidence.EvidenceID != selectorID {
					collector.add(CodeGraphReferenceInvalid, string(evidence.NodeID), "authority", "toolchains", "C4 evidence_id does not match the exact selector identity")
				}
			}
		}
		byNode[evidence.NodeID] = append(byNode[evidence.NodeID], evidence)
	}
	for _, node := range bindingNodes {
		if node.Kind != NodeToolchainComponent {
			continue
		}
		id, _ := node.ID()
		records := byNode[id]
		if len(records) != 1 {
			collector.add(CodeGraphReferenceInvalid, string(id), "authority", "toolchains", "external toolchain requires exactly one C0/C4 binding authority, got %d", len(records))
		}
	}
	for id := range byNode {
		found := false
		for _, node := range bindingNodes {
			nodeID, _ := node.ID()
			if nodeID == id && node.Kind == NodeToolchainComponent {
				found = true
				break
			}
		}
		if !found {
			collector.add(CodeGraphReferenceInvalid, string(id), "authority", "toolchains", "authority names no binding toolchain node")
		}
	}
	for id, selectors := range selectorByNode {
		if len(selectors) != 1 {
			collector.add(CodeGraphReferenceInvalid, string(id), "authority", "c4_selectors", "toolchain has %d C4 selectors", len(selectors))
		}
		records := byNode[id]
		if len(records) != 1 || records[0].FirstBound != ToolchainBoundAtC4 {
			collector.add(CodeGraphReferenceInvalid, string(id), "authority", "c4_selectors", "selector has no matching C4-bound toolchain authority")
		}
	}
}

func validateArtifactManifestRefs(collector *issueCollector, capture CaptureGraph, nodes []Node, edges map[ID]Edge) {
	manifests := idSet(capture.ArtifactManifestIDs)
	nodesByID := map[ID]Node{}
	for _, node := range nodes {
		nodeID, err := node.ID()
		if err == nil {
			nodesByID[nodeID] = node
		}
		var id ID
		switch payload := node.Payload.(type) {
		case PackageInstancePayload:
			id = payload.ArtifactManifestID
		case SourceSetPayload:
			id = payload.ArtifactManifestID
		}
		if id != "" && !manifests[id] {
			collector.add(CodeGraphIncomplete, node.LogicalKey, "capture", "artifact_manifest_ids", "node references absent artifact manifest %s", id)
		}
	}
	edgeIDs := make([]ID, 0, len(edges))
	for edgeID := range edges {
		edgeIDs = append(edgeIDs, edgeID)
	}
	for _, edgeID := range sortedIDs(edgeIDs) {
		edge := edges[edgeID]
		if edge.Kind != EdgeResolvesTo {
			continue
		}
		payload := edge.Payload.(ResolvesToPayload)
		if !manifests[payload.ArtifactManifestID] {
			collector.add(CodeGraphIncomplete, edge.EdgeKey, "capture", "resolves_to.artifact_manifest_id", "edge references absent artifact manifest %s", payload.ArtifactManifestID)
		}
		packageNode, packageOK := nodesByID[edge.FromNodeID]
		sourceNode, sourceOK := nodesByID[edge.ToNodeID]
		if !packageOK || !sourceOK || packageNode.Kind != NodePackageInstance || sourceNode.Kind != NodeSourceSet {
			continue
		}
		packageManifestID := packageNode.Payload.(PackageInstancePayload).ArtifactManifestID
		sourceManifestID := sourceNode.Payload.(SourceSetPayload).ArtifactManifestID
		if payload.ArtifactManifestID != packageManifestID && payload.ArtifactManifestID != sourceManifestID {
			collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "capture", "resolves_to.artifact_manifest_id", "manifest %s matches neither package manifest %s nor source manifest %s", payload.ArtifactManifestID, packageManifestID, sourceManifestID)
		}
	}
}

func validateActionSlots(collector *issueCollector, nodeStates map[ID]ActivationState, selected map[ID]Edge, resolved resolvedTables) {
	for actionID, node := range resolved.captureNodes {
		if node.Kind != NodeAction || nodeStates[actionID] != ActivationSelected {
			continue
		}
		payload := node.Payload.(ActionPayload)
		readCounts := map[string]int{}
		writeCounts := map[string]int{}
		toolCounts := map[string]int{}
		for _, edge := range selected {
			if edge.FromNodeID != actionID {
				continue
			}
			switch value := edge.Payload.(type) {
			case ReadsPayload:
				readCounts[value.ReadSlot]++
				if !containsString(payload.ReadSlotNames, value.ReadSlot) {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "reads.read_slot", "edge binds undeclared read slot %q", value.ReadSlot)
				}
			case ProducesPayload:
				writeCounts[value.WriteSlot]++
				if !containsString(payload.WriteSlotNames, value.WriteSlot) {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "produces.write_slot", "edge binds undeclared write slot %q", value.WriteSlot)
				}
			case UsesToolPayload:
				toolCounts[value.ToolSlot]++
				if !containsString(payload.ToolSlotNames, value.ToolSlot) {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "uses_tool.tool_slot", "edge binds undeclared tool slot %q", value.ToolSlot)
				}
			}
		}
		for _, slot := range payload.ReadSlotNames {
			if readCounts[slot] != 1 {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey+":read:"+slot, "active", "action read slot", "slot must be bound exactly once, got %d", readCounts[slot])
			}
		}
		for _, slot := range payload.WriteSlotNames {
			if writeCounts[slot] != 1 {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey+":write:"+slot, "active", "action write slot", "slot must be bound exactly once, got %d", writeCounts[slot])
			}
		}
		for _, slot := range payload.ToolSlotNames {
			if toolCounts[slot] != 1 {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey+":tool:"+slot, "active", "action tool slot", "slot must be bound exactly once, got %d", toolCounts[slot])
			}
		}
	}
}

func validateDeclaredEdgeSlots(collector *issueCollector, edges map[ID]Edge, resolved resolvedTables) {
	for _, edge := range edges {
		action, ok := resolved.allNodes[edge.FromNodeID]
		if !ok || action.Kind != NodeAction {
			continue
		}
		payload := action.Payload.(ActionPayload)
		switch value := edge.Payload.(type) {
		case ReadsPayload:
			if !containsString(payload.ReadSlotNames, value.ReadSlot) {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "capture+binding", "reads.read_slot", "edge binds undeclared read slot %q", value.ReadSlot)
			}
		case ProducesPayload:
			if !containsString(payload.WriteSlotNames, value.WriteSlot) {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "capture+binding", "produces.write_slot", "edge binds undeclared write slot %q", value.WriteSlot)
			}
		case UsesToolPayload:
			if !containsString(payload.ToolSlotNames, value.ToolSlot) {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "capture+binding", "uses_tool.tool_slot", "edge binds undeclared tool slot %q", value.ToolSlot)
			}
		}
	}
}

func validateEndpointContracts(collector *issueCollector, edges map[ID]Edge, resolved resolvedTables) {
	for _, edge := range edges {
		endpoint, ok := resolved.allNodes[edge.ToNodeID]
		if !ok {
			continue
		}
		switch payload := edge.Payload.(type) {
		case ReadsPayload:
			validateReadEndpoint(collector, edge, payload, endpoint)
		case UsesToolPayload:
			switch value := endpoint.Payload.(type) {
			case ToolchainComponentPayload:
				if payload.ExecutableRelativePath != value.ExecutableRelativePath {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "uses_tool.executable_relative_path", "path %q does not match toolchain path %q", payload.ExecutableRelativePath, value.ExecutableRelativePath)
				}
			case OutputArtifactPayload:
				if payload.ExecutableRelativePath != value.LogicalPath {
					collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "uses_tool.executable_relative_path", "path %q does not match local-tool output path %q", payload.ExecutableRelativePath, value.LogicalPath)
				}
			}
		case ProducesPayload:
			path, class := declaredArtifactPathClass(endpoint)
			if path != "" && payload.Path != path {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "produces.path", "path %q does not match endpoint path %q", payload.Path, path)
			}
			if payload.WriteClass != "" && class != "" && payload.WriteClass != class {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "produces.write_class", "class %q does not match endpoint class %q", payload.WriteClass, class)
			}
		case PublishesPayload:
			if output, ok := endpoint.Payload.(OutputArtifactPayload); ok && payload.Destination != output.LogicalPath {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "publishes.destination", "destination %q does not match output path %q", payload.Destination, output.LogicalPath)
			}
		}
	}
}

func validateReadEndpoint(collector *issueCollector, edge Edge, payload ReadsPayload, endpoint Node) {
	var path, class string
	switch value := endpoint.Payload.(type) {
	case SourceSetPayload:
		class = value.SourceClass
		covered := len(value.Projection) == 0
		for _, projection := range value.Projection {
			if payload.Path == "." || projection == payload.Path || strings.HasPrefix(projection, payload.Path+"/") {
				covered = true
				break
			}
		}
		if !covered {
			collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "reads.path", "path %q is absent from source projection", payload.Path)
		}
		for _, projection := range payload.Projection {
			if !containsString(value.Projection, projection) {
				collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "reads.projection", "member %q is absent from source projection", projection)
			}
		}
		returnIfClassMismatch(collector, edge, payload.ReadClass, class)
		return
	case GeneratedArtifactPayload:
		path, class = value.LogicalPath, value.ExpectedClass
	case OutputArtifactPayload:
		path, class = value.LogicalPath, value.ExpectedClass
	default:
		return
	}
	if payload.Path != path {
		collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "reads.path", "path %q does not match endpoint path %q", payload.Path, path)
	}
	returnIfClassMismatch(collector, edge, payload.ReadClass, class)
}

func returnIfClassMismatch(collector *issueCollector, edge Edge, actual, expected string) {
	if actual != "" && expected != "" && actual != expected {
		collector.add(CodeGraphReferenceInvalid, edge.EdgeKey, "active", "reads.read_class", "class %q does not match endpoint class %q", actual, expected)
	}
}

func declaredArtifactPathClass(node Node) (string, string) {
	switch payload := node.Payload.(type) {
	case GeneratedArtifactPayload:
		return payload.LogicalPath, payload.ExpectedClass
	case OutputArtifactPayload:
		return payload.LogicalPath, payload.ExpectedClass
	default:
		return "", ""
	}
}

func validatePlatformBindings(collector *issueCollector, nodeStates map[ID]ActivationState, selected map[ID]Edge, resolved resolvedTables) {
	counts := map[string]int{}
	for _, edge := range selected {
		if edge.Kind != EdgeTargets {
			continue
		}
		payload := edge.Payload.(TargetsPayload)
		key := string(edge.FromNodeID) + "\x00" + string(payload.BindingRole)
		counts[key]++
		node, ok := resolved.allNodes[edge.FromNodeID]
		if !ok {
			continue
		}
		declared := declaredPlatformRoleSet(node.Payload.declaredPlatformRoles())
		if !declared[payload.BindingRole] {
			collector.add(
				CodeGraphReferenceInvalid,
				edge.EdgeKey,
				"active",
				"targets.binding_role",
				"raw platform role %q is not declared by %s %q; declared raw roles are [%s]",
				payload.BindingRole,
				node.Kind,
				node.LogicalKey,
				joinPlatformRoleSet(declared),
			)
		}
	}
	for id, node := range resolved.allNodes {
		if _, capture := resolved.captureNodes[id]; capture && nodeStates[id] != ActivationSelected {
			continue
		}
		roles := declaredPlatformRoleSet(node.Payload.declaredPlatformRoles())
		for _, role := range sortedPlatformRoleSet(roles) {
			key := string(id) + "\x00" + string(role)
			if counts[key] != 1 {
				collector.add(CodeGraphReferenceInvalid, node.LogicalKey+":"+string(role), "active", "targets", "raw platform role %q must be bound exactly once, got %d", role, counts[key])
			}
		}
	}
}

func declaredPlatformRoleSet(roles []PlatformRole) map[PlatformRole]bool {
	result := make(map[PlatformRole]bool, len(roles))
	for _, role := range roles {
		result[role] = true
	}
	return result
}

func sortedPlatformRoleSet(roles map[PlatformRole]bool) []PlatformRole {
	result := make([]PlatformRole, 0, len(roles))
	for role := range roles {
		result = append(result, role)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func joinPlatformRoleSet(roles map[PlatformRole]bool) string {
	sorted := sortedPlatformRoleSet(roles)
	values := make([]string, len(sorted))
	for index, role := range sorted {
		values[index] = string(role)
	}
	return strings.Join(values, ", ")
}

func validateProducedLineage(collector *issueCollector, nodeStates map[ID]ActivationState, selected map[ID]Edge, resolved resolvedTables) {
	producers := map[ID][]ID{}
	producerActions := map[ID][]ID{}
	for edgeID, edge := range selected {
		if edge.Kind != EdgeProduces {
			continue
		}
		producers[edge.ToNodeID] = append(producers[edge.ToNodeID], edgeID)
		producerActions[edge.ToNodeID] = append(producerActions[edge.ToNodeID], edge.FromNodeID)
	}
	for nodeID, node := range resolved.captureNodes {
		if nodeStates[nodeID] != ActivationSelected || (node.Kind != NodeGeneratedArtifact && node.Kind != NodeOutputArtifact) {
			continue
		}
		edgeIDs := sortedIDs(producers[nodeID])
		actionIDs := sortedIDs(producerActions[nodeID])
		switch len(edgeIDs) {
		case 0:
			collector.add(CodeGeneratedInputUndeclared, node.LogicalKey, "active", "produces", "selected %s requires exactly one producer, got 0", node.Kind)
		case 1:
			// Exact single-producer lineage is closed.
		default:
			collector.add(CodeGraphReferenceInvalid, node.LogicalKey, "active", "produces", "selected %s requires exactly one producer, got %d actions=%s edges=%s", node.Kind, len(edgeIDs), joinIDs(actionIDs), joinIDs(edgeIDs))
		}
	}
}

type interopSide struct {
	edgeID ID
	edge   Edge
}

func validateInteropBoundaries(collector *issueCollector, nodeStates map[ID]ActivationState, selected map[ID]Edge, resolved resolvedTables) {
	providers := map[ID][]interopSide{}
	consumers := map[ID][]interopSide{}
	platforms := map[ID]map[ID]bool{}
	actionSet := map[ID]bool{}
	producerActions := map[ID][]actionEvidence{}
	invocations := []interopSide{}
	publicationCounts := map[ID]int{}
	boundaryToolchains := map[ID][]ID{}
	for nodeID, node := range resolved.captureNodes {
		if nodeStates[nodeID] == ActivationSelected && node.Kind == NodeAction {
			actionSet[nodeID] = true
		}
	}
	for edgeID, edge := range selected {
		switch edge.Kind {
		case EdgeProvidesInterop:
			providers[edge.ToNodeID] = append(providers[edge.ToNodeID], interopSide{edgeID: edgeID, edge: edge})
		case EdgeConsumesInterop:
			consumers[edge.ToNodeID] = append(consumers[edge.ToNodeID], interopSide{edgeID: edgeID, edge: edge})
		case EdgeTargets:
			if platforms[edge.FromNodeID] == nil {
				platforms[edge.FromNodeID] = map[ID]bool{}
			}
			platforms[edge.FromNodeID][edge.ToNodeID] = true
		case EdgeProduces:
			if actionSet[edge.FromNodeID] {
				producerActions[edge.ToNodeID] = append(producerActions[edge.ToNodeID], actionEvidence{ActionID: edge.FromNodeID, EdgeIDs: []ID{edgeID}})
			}
		case EdgeInvokes:
			invocations = append(invocations, interopSide{edgeID: edgeID, edge: edge})
		case EdgePublishes:
			publicationCounts[edge.ToNodeID]++
		case EdgeRequires:
			_, bindingOwned := resolved.bindingEdges[edgeID]
			from, fromOK := resolved.captureNodes[edge.FromNodeID]
			to, toOK := resolved.bindingNodes[edge.ToNodeID]
			payload := edge.Payload.(RequiresPayload)
			if bindingOwned && fromOK && from.Kind == NodeInteropBoundary && toOK && to.Kind == NodeToolchainComponent && payload.Scope == ScopeToolchain {
				boundaryToolchains[edge.FromNodeID] = append(boundaryToolchains[edge.FromNodeID], edge.ToNodeID)
			}
		}
	}
	actionsByOwner := declaredActionEvidence(selected, resolved.allNodes, actionSet)
	languages := interopLanguages(nodeStates, resolved, actionsByOwner, producerActions)

	boundaryIDs := make([]ID, 0)
	for nodeID, node := range resolved.captureNodes {
		if nodeStates[nodeID] == ActivationSelected && node.Kind == NodeInteropBoundary {
			boundaryIDs = append(boundaryIDs, nodeID)
		}
	}
	boundaryIDs = sortedIDs(boundaryIDs)
	for _, boundaryID := range boundaryIDs {
		boundary := resolved.captureNodes[boundaryID]
		payload := boundary.Payload.(InteropBoundaryPayload)
		providerSides := providers[boundaryID]
		consumerSides := consumers[boundaryID]
		if len(providerSides) != 1 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":provider", "active", "provides_interop", "selected boundary requires exactly one provider side, got %d", len(providerSides))
		}
		if len(consumerSides) != 1 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":consumer", "active", "consumes_interop", "selected boundary requires exactly one consumer side, got %d", len(consumerSides))
		}
		if len(providerSides) != 1 || len(consumerSides) != 1 {
			continue
		}
		provider, consumer := providerSides[0], consumerSides[0]
		providerNode := resolved.allNodes[provider.edge.FromNodeID]
		providerPayload := provider.edge.Payload.(ProvidesInteropPayload)
		consumerPayload := consumer.edge.Payload.(ConsumesInteropPayload)
		if len(providerPayload.EvidenceIDs) == 0 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":evidence", "active", "provides_interop.evidence_ids", "selected provider side requires at least one immutable interface evidence ID")
		}
		if providerNode.Kind != NodeToolchainComponent && !languageClassesOverlap(languages[provider.edge.FromNodeID], payload.ProviderLanguageClasses) {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":provider-language", "active", "provider_language_classes", "provider %s languages are incompatible with declared classes %s", provider.edge.FromNodeID, strings.Join(payload.ProviderLanguageClasses, ","))
		}
		if !languageClassesOverlap(languages[consumer.edge.FromNodeID], payload.ConsumerLanguageClasses) {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":consumer-language", "active", "consumer_language_classes", "consumer %s languages are incompatible with declared classes %s", consumer.edge.FromNodeID, strings.Join(payload.ConsumerLanguageClasses, ","))
		}
		if payload.ABI != "" && consumerPayload.ABIExpectation != payload.ABI {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":abi", "active", "consumes_interop.abi_expectation", "consumer ABI expectation %q does not match boundary ABI %q", consumerPayload.ABIExpectation, payload.ABI)
		}
		if !sharesBoundPlatform(platforms[boundaryID], platforms[provider.edge.FromNodeID]) || !sharesBoundPlatform(platforms[boundaryID], platforms[consumer.edge.FromNodeID]) {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":platform", "active", "targets", "provider, consumer, and boundary must share an exact selected platform binding")
		}
		if payload.Mode != InteropSubprocessProtocol {
			toolchainIDs := sortedIDs(boundaryToolchains[boundaryID])
			if len(toolchainIDs) != 1 {
				collector.add(CodeInteropUndeclared, boundary.LogicalKey+":toolchain", "active", "requires", "selected boundary requires exactly one explicit toolchain-scoped binding, got %d", len(toolchainIDs))
			}
			for _, toolchainID := range toolchainIDs {
				if !sharesBoundPlatform(platforms[boundaryID], platforms[toolchainID]) {
					collector.add(CodeInteropUndeclared, boundary.LogicalKey+":toolchain-platform", "active", "requires", "boundary toolchain %s does not share its exact selected platform", toolchainID)
				}
			}
		}

		providerActionEvidence := actionEvidenceForNode(provider.edge.FromNodeID, provider.edgeID, producerActions, actionsByOwner, actionSet)
		consumerActionEvidence := actionEvidenceForNode(consumer.edge.FromNodeID, consumer.edgeID, producerActions, actionsByOwner, actionSet)
		if payload.Mode == InteropSubprocessProtocol {
			validateSubprocessBoundary(collector, boundary, payload, provider, consumer, invocations, publicationCounts, resolved.allNodes)
			continue
		}
		if payload.Mode == InteropDynamicLoad {
			if providerNode.Kind != NodeOutputArtifact && providerNode.Kind != NodeToolchainComponent {
				collector.add(CodeInteropUndeclared, boundary.LogicalKey+":dynamic-provider", "active", "provides_interop", "dynamic_load provider must be an exact produced output or selected external toolchain module")
			}
			if len(consumerActionEvidence) == 0 {
				collector.add(CodeInteropUndeclared, boundary.LogicalKey+":dynamic-consumer", "active", "consumes_interop", "dynamic_load consumer side has no selected loading action")
			}
			continue
		}
		if payload.Mode == InteropHostExtension && providerNode.Kind != NodeOutputArtifact && providerNode.Kind != NodeToolchainComponent {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":host-extension-provider", "active", "provides_interop", "host_extension provider must be an exact produced output or selected external toolchain executable")
		}
		if providerNode.Kind != NodeToolchainComponent && len(providerActionEvidence) == 0 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":provider-action", "active", "provides_interop", "compile/link provider side has no selected producing action")
		}
		if len(consumerActionEvidence) == 0 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":consumer-action", "active", "consumes_interop", "compile/link consumer side has no selected consuming action")
		}
		if providerNode.Kind != NodeToolchainComponent && !hasDistinctActionPair(providerActionEvidence, consumerActionEvidence) {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":ordering", "active", "interop ordering", "compile/link boundary has no distinct provider-before-consumer action pair")
		}
	}
}

func validateSubprocessBoundary(collector *issueCollector, boundary Node, payload InteropBoundaryPayload, provider, consumer interopSide, invocations []interopSide, publicationCounts map[ID]int, nodes map[ID]Node) {
	providerNode := nodes[provider.edge.FromNodeID]
	consumerNode := nodes[consumer.edge.FromNodeID]
	if providerNode.Kind != NodeOutputArtifact || consumerNode.Kind != NodeOutputArtifact || provider.edge.FromNodeID == consumer.edge.FromNodeID {
		collector.add(CodeInteropUndeclared, boundary.LogicalKey+":subprocess-sides", "active", "subprocess_protocol", "subprocess provider and consumer must be distinct produced output_artifact nodes")
		return
	}
	matchingInvocations := []interopSide{}
	for _, invocation := range invocations {
		if invocation.edge.FromNodeID == consumer.edge.FromNodeID && invocation.edge.ToNodeID == provider.edge.FromNodeID {
			matchingInvocations = append(matchingInvocations, invocation)
		}
	}
	if len(matchingInvocations) != 1 {
		collector.add(CodeInteropUndeclared, boundary.LogicalKey+":invocation", "active", "invokes", "subprocess boundary requires exactly one consumer-to-provider invokes edge, got %d", len(matchingInvocations))
	} else {
		invocation := matchingInvocations[0].edge.Payload.(InvokesPayload)
		if invocation.ProtocolSchema != payload.ProtocolSchema {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":protocol", "active", "invokes.protocol_schema", "invocation protocol %q does not match boundary protocol %q", invocation.ProtocolSchema, payload.ProtocolSchema)
		}
		if invocation.WorkingDirectory == "" {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":working-directory", "active", "invokes.working_directory", "subprocess invocation requires an explicit working-directory contract")
		}
	}
	for _, side := range []struct {
		name   string
		nodeID ID
	}{{name: "provider", nodeID: provider.edge.FromNodeID}, {name: "consumer", nodeID: consumer.edge.FromNodeID}} {
		if publicationCounts[side.nodeID] != 1 {
			collector.add(CodeInteropUndeclared, boundary.LogicalKey+":"+side.name+"-publication", "active", "publishes", "subprocess %s output requires exactly one publication edge, got %d", side.name, publicationCounts[side.nodeID])
		}
	}
}

func interopLanguages(nodeStates map[ID]ActivationState, resolved resolvedTables, actionsByOwner map[ID][]actionEvidence, producers map[ID][]actionEvidence) map[ID]map[string]bool {
	result := map[ID]map[string]bool{}
	add := func(nodeID ID, values []string) {
		if result[nodeID] == nil {
			result[nodeID] = map[string]bool{}
		}
		for _, value := range values {
			result[nodeID][value] = true
		}
	}
	for targetID, node := range resolved.captureNodes {
		if nodeStates[targetID] != ActivationSelected || node.Kind != NodeTargetUnit {
			continue
		}
		values := node.Payload.(TargetUnitPayload).Languages
		add(targetID, values)
		for _, action := range actionsByOwner[targetID] {
			add(action.ActionID, values)
		}
	}
	for artifactID, producerList := range producers {
		for _, producer := range producerList {
			for language := range result[producer.ActionID] {
				add(artifactID, []string{language})
			}
		}
	}
	return result
}

func languageClassesOverlap(actual map[string]bool, declared []string) bool {
	for _, language := range declared {
		if actual[language] {
			return true
		}
	}
	return false
}

func sharesBoundPlatform(left, right map[ID]bool) bool {
	for platformID := range left {
		if right[platformID] {
			return true
		}
	}
	return false
}

func hasDistinctActionPair(providers, consumers []actionEvidence) bool {
	for _, provider := range providers {
		for _, consumer := range consumers {
			if provider.ActionID != consumer.ActionID {
				return true
			}
		}
	}
	return false
}

func selectedEdges(active ActiveGraph, resolved resolvedTables) map[ID]Edge {
	edgeStates := map[ID]ActivationState{}
	for _, activation := range active.EdgeActivations {
		edgeStates[activation.EdgeID] = activation.State
	}
	nodeStates := map[ID]ActivationState{}
	for _, activation := range active.NodeActivations {
		nodeStates[activation.NodeID] = activation.State
	}
	result := make(map[ID]Edge, len(resolved.allEdges))
	for id, edge := range resolved.captureEdges {
		if nodeStates[edge.FromNodeID] != ActivationSelected || nodeStates[edge.ToNodeID] != ActivationSelected {
			continue
		}
		if edge.Payload.condition() == nil || edgeStates[id] == ActivationSelected {
			result[id] = edge
		}
	}
	for id, edge := range resolved.bindingEdges {
		result[id] = edge
	}
	return result
}

func allowedBindingEdgeKind(kind EdgeKind) bool {
	return kind == EdgeTargets || kind == EdgeUsesTool || kind == EdgeRequires || kind == EdgeProvidesInterop
}

func allowedEndpoints(kind EdgeKind, from, to NodeKind) bool {
	in := func(value NodeKind, allowed ...NodeKind) bool {
		for _, candidate := range allowed {
			if value == candidate {
				return true
			}
		}
		return false
	}
	switch kind {
	case EdgeDeclares:
		return in(from, NodePackageInstance, NodeCommandProduct, NodeTargetUnit) && in(to, NodeTargetUnit, NodeAction, NodeInteropBoundary, NodeSourceSet)
	case EdgeResolvesTo:
		return from == NodePackageInstance && to == NodeSourceSet
	case EdgeRequires:
		return in(from, NodeCommandProduct, NodePackageInstance, NodeTargetUnit, NodeAction, NodeInteropBoundary) && in(to, NodePackageInstance, NodeTargetUnit, NodeOutputArtifact, NodeToolchainComponent)
	case EdgeReads:
		return in(from, NodeTargetUnit, NodeAction) && in(to, NodeSourceSet, NodeGeneratedArtifact, NodeOutputArtifact, NodeToolchainComponent)
	case EdgeUsesTool:
		return from == NodeAction && in(to, NodeToolchainComponent, NodeOutputArtifact)
	case EdgeTargets:
		return in(from, NodeCommandProduct, NodeTargetUnit, NodeAction, NodeToolchainComponent, NodeOutputArtifact, NodeInteropBoundary) && to == NodeTargetPlatform
	case EdgeProduces:
		return from == NodeAction && in(to, NodeGeneratedArtifact, NodeOutputArtifact)
	case EdgeProvidesInterop:
		return in(from, NodeTargetUnit, NodeOutputArtifact, NodeToolchainComponent) && to == NodeInteropBoundary
	case EdgeConsumesInterop:
		return in(from, NodeTargetUnit, NodeAction, NodeOutputArtifact) && to == NodeInteropBoundary
	case EdgeInvokes:
		return in(from, NodeCommandProduct, NodeOutputArtifact, NodeAction) && in(to, NodeCommandProduct, NodeOutputArtifact, NodeAction)
	case EdgePublishes:
		return from == NodeCommandProduct && to == NodeOutputArtifact
	default:
		return false
	}
}

func semanticEdgeKey(edge Edge) (string, error) {
	payloadValue := edge.Payload.value()
	if _, originBearing := edgeEvidenceOrigin(edge.Payload); originBearing {
		delete(payloadValue, "origin")
	}
	payload, err := canonicalMapBytes(map[string]any{"from_node_id": string(edge.FromNodeID), "kind": string(edge.Kind), "payload": payloadValue, "to_node_id": string(edge.ToNodeID)})
	return string(payload), err
}

func semanticEdgeEvidence(table string, edge Edge, edgeID ID) (string, error) {
	evidence := map[string]any{
		"edge_id":  string(edgeID),
		"edge_key": edge.EdgeKey,
		"table":    table,
	}
	if origin, ok := edgeEvidenceOrigin(edge.Payload); ok {
		evidence["evidence_origin"] = origin.value()
	}
	encoded, err := canonicalMapBytes(evidence)
	return string(encoded), err
}

func edgeEvidenceOrigin(payload EdgePayload) (EvidenceOrigin, bool) {
	switch value := payload.(type) {
	case DeclaresPayload:
		return value.Origin, true
	case ResolvesToPayload:
		return value.Origin, true
	case RequiresPayload:
		return value.Origin, true
	case TargetsPayload:
		return value.Origin, true
	case ProvidesInteropPayload:
		return value.Origin, true
	case ConsumesInteropPayload:
		return value.Origin, true
	default:
		return EvidenceOrigin{}, false
	}
}

func checkReferencedIDs[T any](collector *issueCollector, table, kind string, referenced []ID, records map[ID]T) {
	want := idSet(referenced)
	for id := range want {
		if _, ok := records[id]; !ok {
			collector.add(CodeGraphIncomplete, string(id), table, kind+"_ids", "referenced %s record is absent", kind)
		}
	}
	for id := range records {
		if !want[id] {
			collector.add(CodeGraphReferenceInvalid, string(id), table, kind+" table", "unreferenced extra %s record", kind)
		}
	}
}

func platformIDForRole(selection SelectionContext, role PlatformRole) (ID, bool) {
	id, ok := selection.PlatformRoles[role]
	if !ok && role == PlatformHost {
		id, ok = selection.PlatformRoles[PlatformTarget]
	}
	return id, ok
}
func containsString(values []string, value string) bool {
	index := sort.SearchStrings(values, value)
	return index < len(values) && values[index] == value
}
func idSet(values []ID) map[ID]bool {
	result := make(map[ID]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func sameNodeOrder(left, right []Node) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		leftID, _ := left[i].ID()
		rightID, _ := right[i].ID()
		if leftID != rightID || left[i].LogicalKey != right[i].LogicalKey || left[i].Kind != right[i].Kind {
			return false
		}
	}
	return true
}
func sameEdgeOrder(left, right []Edge) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		leftID, _ := left[i].ID()
		rightID, _ := right[i].ID()
		if leftID != rightID || left[i].EdgeKey != right[i].EdgeKey || left[i].Kind != right[i].Kind {
			return false
		}
	}
	return true
}
func sameSCCs(left, right []NonOrderingSCC) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		leftID, _ := left[i].ID()
		rightID, _ := right[i].ID()
		if leftID != rightID {
			return false
		}
	}
	return true
}
