package closuregraph

import (
	"fmt"
	"sort"
)

// CaptureGraph is the selection-neutral conservative lock/resolution
// superset. Concrete platforms, external toolchains, and target/tool edges are
// forbidden from its record table.
type CaptureGraph struct {
	SchemaID            string
	ProfileID           string
	PolicyIDs           []string
	RootNodeIDs         []ID
	NodeIDs             []ID
	EdgeIDs             []ID
	ArtifactManifestIDs []ID
}

// NewCaptureGraph builds a canonically ordered capture envelope from records.
func NewCaptureGraph(profileID string, policyIDs []string, rootNodeIDs []ID, nodes []Node, edges []Edge, artifactManifestIDs []ID) (CaptureGraph, error) {
	nodeIDs := make([]ID, len(nodes))
	for i, node := range nodes {
		id, err := node.ID()
		if err != nil {
			return CaptureGraph{}, fmt.Errorf("capture node %d: %w", i, err)
		}
		nodeIDs[i] = id
	}
	edgeIDs := make([]ID, len(edges))
	for i, edge := range edges {
		id, err := edge.ID()
		if err != nil {
			return CaptureGraph{}, fmt.Errorf("capture edge %d: %w", i, err)
		}
		edgeIDs[i] = id
	}
	graph := CaptureGraph{SchemaID: SchemaCaptureGraph, ProfileID: profileID, PolicyIDs: sortedStrings(policyIDs), RootNodeIDs: sortedIDs(rootNodeIDs), NodeIDs: sortedIDs(nodeIDs), EdgeIDs: sortedIDs(edgeIDs), ArtifactManifestIDs: sortedIDs(artifactManifestIDs)}
	return graph, graph.Validate()
}

// Validate checks the capture envelope independently of its record table.
func (graph CaptureGraph) Validate() error {
	if graph.SchemaID != SchemaCaptureGraph {
		return fmt.Errorf("%s: unsupported capture schema %q", CodeGraphSchemaUnsupported, graph.SchemaID)
	}
	if err := validatePortableText(graph.ProfileID, "capture profile_id", false); err != nil {
		return err
	}
	if err := validateStringSlice(graph.PolicyIDs, "capture policy_ids", true); err != nil {
		return err
	}
	if err := validateIDSlice(graph.RootNodeIDs, "capture root_node_ids", true); err != nil {
		return err
	}
	if len(graph.RootNodeIDs) == 0 {
		return fmt.Errorf("capture root_node_ids must not be empty")
	}
	if err := validateIDSlice(graph.NodeIDs, "capture node_ids", true); err != nil {
		return err
	}
	if err := validateIDSlice(graph.EdgeIDs, "capture edge_ids", true); err != nil {
		return err
	}
	return validateIDSlice(graph.ArtifactManifestIDs, "capture artifact_manifest_ids", true)
}

// CanonicalBytes returns exact curator-capture-graph-v1 CCJ bytes.
func (graph CaptureGraph) CanonicalBytes() ([]byte, error) { return canonicalBytes(graph) }

// ID derives captured_graph_id.
func (graph CaptureGraph) ID() (ID, error) { return recordID(graph) }

func (graph CaptureGraph) domainLabel() string { return LabelCaptureGraph }
func (graph CaptureGraph) canonicalValue() map[string]any {
	return map[string]any{"artifact_manifest_ids": idsToAny(graph.ArtifactManifestIDs), "edge_ids": idsToAny(graph.EdgeIDs), "node_ids": idsToAny(graph.NodeIDs), "policy_ids": stringsToAny(graph.PolicyIDs), "profile_id": graph.ProfileID, "root_node_ids": idsToAny(graph.RootNodeIDs), "schema_id": graph.SchemaID}
}

// DecodeCaptureGraph accepts exact canonical graph bytes.
func DecodeCaptureGraph(payload []byte) (CaptureGraph, error) {
	raw, err := decodeCanonicalObject(payload, "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	if err := exactFields(raw, "capture graph", []string{"artifact_manifest_ids", "edge_ids", "node_ids", "policy_ids", "profile_id", "root_node_ids", "schema_id"}, nil); err != nil {
		return CaptureGraph{}, err
	}
	graph := CaptureGraph{}
	graph.SchemaID, err = requiredString(raw, "schema_id", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.ProfileID, err = requiredString(raw, "profile_id", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.PolicyIDs, err = requiredStringSlice(raw, "policy_ids", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.RootNodeIDs, err = requiredIDSlice(raw, "root_node_ids", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.NodeIDs, err = requiredIDSlice(raw, "node_ids", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.EdgeIDs, err = requiredIDSlice(raw, "edge_ids", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.ArtifactManifestIDs, err = requiredIDSlice(raw, "artifact_manifest_ids", "capture graph")
	if err != nil {
		return CaptureGraph{}, err
	}
	if err := graph.Validate(); err != nil {
		return CaptureGraph{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, graph); err != nil {
		return CaptureGraph{}, err
	}
	return graph, nil
}

// SelectionContext contains only requested selection values and intrinsic
// target-platform IDs. Platform records themselves live in the binding table.
type SelectionContext struct {
	SchemaID        string
	ProductNodeIDs  []ID
	PlatformRoles   map[PlatformRole]ID
	Features        []string
	DefaultFeatures bool
	Markers         map[string]string
	PeerContext     map[string]string
	EvaluatorIDs    []string
}

// NewSelectionContext canonicalizes set-like selection members.
func NewSelectionContext(productNodeIDs []ID, platformRoles map[PlatformRole]ID, features []string, defaultFeatures bool, markers, peerContext map[string]string, evaluatorIDs []string) (SelectionContext, error) {
	selection := SelectionContext{SchemaID: SchemaSelectionContext, ProductNodeIDs: sortedIDs(productNodeIDs), PlatformRoles: clonePlatformRoles(platformRoles), Features: sortedStrings(features), DefaultFeatures: defaultFeatures, Markers: cloneStringMap(markers), PeerContext: cloneStringMap(peerContext), EvaluatorIDs: sortedStrings(evaluatorIDs)}
	return selection, selection.Validate()
}

// Validate checks the selection context schema and canonical order.
func (selection SelectionContext) Validate() error {
	if selection.SchemaID != SchemaSelectionContext {
		return fmt.Errorf("%s: unsupported selection schema %q", CodeGraphSchemaUnsupported, selection.SchemaID)
	}
	if err := validateIDSlice(selection.ProductNodeIDs, "selection product_node_ids", true); err != nil {
		return err
	}
	if len(selection.ProductNodeIDs) == 0 {
		return fmt.Errorf("selection product_node_ids must not be empty")
	}
	if selection.PlatformRoles == nil {
		return fmt.Errorf("selection platform_roles must be an explicit object")
	}
	target, present := selection.PlatformRoles[PlatformTarget]
	if !present {
		return fmt.Errorf("selection platform_roles requires target")
	}
	if err := validateID(target, "selection platform_roles.target"); err != nil {
		return err
	}
	for _, role := range sortedPlatformRoles(selection.PlatformRoles) {
		id := selection.PlatformRoles[role]
		if role != PlatformTarget && role != PlatformHost {
			return fmt.Errorf("unsupported platform role %q", role)
		}
		if err := validateID(id, "selection platform role"); err != nil {
			return err
		}
	}
	if err := validateStringSlice(selection.Features, "selection features", true); err != nil {
		return err
	}
	if selection.Markers == nil {
		return fmt.Errorf("selection markers must be an explicit object")
	}
	if selection.PeerContext == nil {
		return fmt.Errorf("selection peer_context must be an explicit object")
	}
	if err := validatePortableStringMap(selection.Markers, "markers", true); err != nil {
		return err
	}
	if err := validatePortableStringMap(selection.PeerContext, "peer_context", true); err != nil {
		return err
	}
	return validateStringSlice(selection.EvaluatorIDs, "selection evaluator_ids", true)
}

// CanonicalBytes returns exact curator-selection-context-v1 CCJ bytes.
func (selection SelectionContext) CanonicalBytes() ([]byte, error) { return canonicalBytes(selection) }

// ID derives selection_context_id.
func (selection SelectionContext) ID() (ID, error) { return recordID(selection) }

func (selection SelectionContext) domainLabel() string { return LabelSelectionContext }
func (selection SelectionContext) canonicalValue() map[string]any {
	roles := make(map[string]any, len(selection.PlatformRoles))
	for role, id := range selection.PlatformRoles {
		roles[string(role)] = string(id)
	}
	return map[string]any{"default_features": selection.DefaultFeatures, "evaluator_ids": stringsToAny(selection.EvaluatorIDs), "features": stringsToAny(selection.Features), "markers": stringMapToAny(selection.Markers), "peer_context": stringMapToAny(selection.PeerContext), "platform_roles": roles, "product_node_ids": idsToAny(selection.ProductNodeIDs), "schema_id": selection.SchemaID}
}

// DecodeSelectionContext accepts exact canonical selection bytes.
func DecodeSelectionContext(payload []byte) (SelectionContext, error) {
	raw, err := decodeCanonicalObject(payload, "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	if err := exactFields(raw, "selection context", []string{"default_features", "evaluator_ids", "features", "markers", "peer_context", "platform_roles", "product_node_ids", "schema_id"}, nil); err != nil {
		return SelectionContext{}, err
	}
	selection := SelectionContext{}
	selection.SchemaID, err = requiredString(raw, "schema_id", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.DefaultFeatures, err = requiredBool(raw, "default_features", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.EvaluatorIDs, err = requiredStringSlice(raw, "evaluator_ids", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.Features, err = requiredStringSlice(raw, "features", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.Markers, err = requiredStringMap(raw, "markers", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.PeerContext, err = requiredStringMap(raw, "peer_context", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	rolesRaw, err := requiredObject(raw, "platform_roles", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	selection.PlatformRoles = make(map[PlatformRole]ID, len(rolesRaw))
	for _, key := range sortedMapKeys(rolesRaw) {
		rawID := rolesRaw[key]
		text, ok := rawID.(string)
		if !ok {
			return SelectionContext{}, fmt.Errorf("platform_roles.%s must be a string", key)
		}
		selection.PlatformRoles[PlatformRole(key)] = ID(text)
	}
	selection.ProductNodeIDs, err = requiredIDSlice(raw, "product_node_ids", "selection context")
	if err != nil {
		return SelectionContext{}, err
	}
	if err := selection.Validate(); err != nil {
		return SelectionContext{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, selection); err != nil {
		return SelectionContext{}, err
	}
	return selection, nil
}

// SelectionBinding is the only selection-specific record overlay.
type SelectionBinding struct {
	SchemaID           string
	CapturedGraphID    ID
	SelectionContextID ID
	BindingNodeIDs     []ID
	BindingEdgeIDs     []ID
}

// NewSelectionBinding creates a canonical binding envelope.
func NewSelectionBinding(capturedGraphID, selectionContextID ID, nodes []Node, edges []Edge) (SelectionBinding, error) {
	nodeIDs := make([]ID, len(nodes))
	for i, node := range nodes {
		id, err := node.ID()
		if err != nil {
			return SelectionBinding{}, err
		}
		nodeIDs[i] = id
	}
	edgeIDs := make([]ID, len(edges))
	for i, edge := range edges {
		id, err := edge.ID()
		if err != nil {
			return SelectionBinding{}, err
		}
		edgeIDs[i] = id
	}
	binding := SelectionBinding{SchemaID: SchemaSelectionBinding, CapturedGraphID: capturedGraphID, SelectionContextID: selectionContextID, BindingNodeIDs: sortedIDs(nodeIDs), BindingEdgeIDs: sortedIDs(edgeIDs)}
	return binding, binding.Validate()
}

// Validate checks the binding envelope independently of record tables.
func (binding SelectionBinding) Validate() error {
	if binding.SchemaID != SchemaSelectionBinding {
		return fmt.Errorf("%s: unsupported binding schema %q", CodeGraphSchemaUnsupported, binding.SchemaID)
	}
	if err := validateID(binding.CapturedGraphID, "binding captured_graph_id"); err != nil {
		return err
	}
	if err := validateID(binding.SelectionContextID, "binding selection_context_id"); err != nil {
		return err
	}
	if err := validateIDSlice(binding.BindingNodeIDs, "binding binding_node_ids", true); err != nil {
		return err
	}
	if len(binding.BindingNodeIDs) == 0 {
		return fmt.Errorf("binding binding_node_ids must not be empty")
	}
	return validateIDSlice(binding.BindingEdgeIDs, "binding binding_edge_ids", true)
}

// CanonicalBytes returns exact curator-selection-binding-v1 CCJ bytes.
func (binding SelectionBinding) CanonicalBytes() ([]byte, error) { return canonicalBytes(binding) }

// ID derives selection_binding_id.
func (binding SelectionBinding) ID() (ID, error) { return recordID(binding) }

func (binding SelectionBinding) domainLabel() string { return LabelSelectionBinding }
func (binding SelectionBinding) canonicalValue() map[string]any {
	return map[string]any{"binding_edge_ids": idsToAny(binding.BindingEdgeIDs), "binding_node_ids": idsToAny(binding.BindingNodeIDs), "captured_graph_id": string(binding.CapturedGraphID), "schema_id": binding.SchemaID, "selection_context_id": string(binding.SelectionContextID)}
}

// DecodeSelectionBinding accepts exact canonical binding bytes.
func DecodeSelectionBinding(payload []byte) (SelectionBinding, error) {
	raw, err := decodeCanonicalObject(payload, "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	if err := exactFields(raw, "selection binding", []string{"binding_edge_ids", "binding_node_ids", "captured_graph_id", "schema_id", "selection_context_id"}, nil); err != nil {
		return SelectionBinding{}, err
	}
	binding := SelectionBinding{}
	binding.SchemaID, err = requiredString(raw, "schema_id", "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	capture, err := requiredString(raw, "captured_graph_id", "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	binding.CapturedGraphID = ID(capture)
	selection, err := requiredString(raw, "selection_context_id", "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	binding.SelectionContextID = ID(selection)
	binding.BindingNodeIDs, err = requiredIDSlice(raw, "binding_node_ids", "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	binding.BindingEdgeIDs, err = requiredIDSlice(raw, "binding_edge_ids", "selection binding")
	if err != nil {
		return SelectionBinding{}, err
	}
	if err := binding.Validate(); err != nil {
		return SelectionBinding{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, binding); err != nil {
		return SelectionBinding{}, err
	}
	return binding, nil
}

// ActivationState is the exact selected/pruned state of a capture record.
type ActivationState string

const (
	// ActivationSelected and ActivationPruned are the closed selection states.
	ActivationSelected ActivationState = "selected"
	// ActivationPruned marks a captured record excluded by exact selection.
	ActivationPruned ActivationState = "pruned"
)

// ActivationReason is the closed explanation for a conditional edge result.
type ActivationReason string

const (
	// ReasonConditionTrue and the related constants explain active-projection
	// outcomes independently of traversal order.
	ReasonConditionTrue ActivationReason = "condition_true"
	// ReasonConditionFalse marks an unevaluated declaration that evaluated false.
	ReasonConditionFalse ActivationReason = "condition_false"
	// ReasonUnreachable marks a declaration outside the selected reachable graph.
	ReasonUnreachable ActivationReason = "unreachable"
)

// NodeActivation records selection state for one capture node.
type NodeActivation struct {
	NodeID ID
	State  ActivationState
}

// EdgeActivation records the evaluated state of one conditional capture edge.
type EdgeActivation struct {
	EdgeID     ID
	Evaluation bool
	Reason     ActivationReason
	State      ActivationState
}

// NonOrderingSCC retains a permitted runtime/peer/non-ordering cycle.
type NonOrderingSCC struct {
	NodeIDs []ID
	EdgeIDs []ID
}

func (scc NonOrderingSCC) validate() error {
	if err := validateIDSlice(scc.NodeIDs, "non-ordering SCC node_ids", true); err != nil {
		return err
	}
	if len(scc.NodeIDs) == 0 {
		return fmt.Errorf("non-ordering SCC node_ids must not be empty")
	}
	return validateIDSlice(scc.EdgeIDs, "non-ordering SCC edge_ids", true)
}
func (scc NonOrderingSCC) value() map[string]any {
	return map[string]any{"edge_ids": idsToAny(scc.EdgeIDs), "node_ids": idsToAny(scc.NodeIDs)}
}

// ID derives the canonical identity of this non-ordering SCC record.
func (scc NonOrderingSCC) ID() (ID, error) {
	if err := scc.validate(); err != nil {
		return "", err
	}
	return DomainID(LabelNonOrderingSCC, scc.value())
}

// ActiveGraph is the exact selected capture projection plus binding overlay.
// Binding records are referenced once and never copied into activations.
type ActiveGraph struct {
	SchemaID           string
	CapturedGraphID    ID
	SelectionContextID ID
	SelectionBindingID ID
	NodeActivations    []NodeActivation
	EdgeActivations    []EdgeActivation
	NonOrderingSCCs    []NonOrderingSCC
}

// Validate checks active-record shape and canonical ordering. Cross-table
// completeness is enforced by GraphBundle.Validate.
func (graph ActiveGraph) Validate() error {
	if graph.SchemaID != SchemaActiveGraph {
		return fmt.Errorf("%s: unsupported active graph schema %q", CodeGraphSchemaUnsupported, graph.SchemaID)
	}
	if err := validateIDFields(map[string]ID{"active captured_graph_id": graph.CapturedGraphID, "active selection_context_id": graph.SelectionContextID, "active selection_binding_id": graph.SelectionBindingID}); err != nil {
		return err
	}
	if graph.NodeActivations == nil {
		return fmt.Errorf("active node_activations must be an explicit array")
	}
	for index, activation := range graph.NodeActivations {
		if err := validateID(activation.NodeID, fmt.Sprintf("node_activations[%d].node_id", index)); err != nil {
			return err
		}
		if activation.State != ActivationSelected && activation.State != ActivationPruned {
			return fmt.Errorf("unsupported node activation state %q", activation.State)
		}
		if index > 0 && graph.NodeActivations[index-1].NodeID >= activation.NodeID {
			return fmt.Errorf("active node_activations must be sorted and unique")
		}
	}
	if graph.EdgeActivations == nil {
		return fmt.Errorf("active edge_activations must be an explicit array")
	}
	for index, activation := range graph.EdgeActivations {
		if err := validateID(activation.EdgeID, fmt.Sprintf("edge_activations[%d].edge_id", index)); err != nil {
			return err
		}
		if activation.State != ActivationSelected && activation.State != ActivationPruned {
			return fmt.Errorf("unsupported edge activation state %q", activation.State)
		}
		if activation.Reason != ReasonConditionTrue && activation.Reason != ReasonConditionFalse && activation.Reason != ReasonUnreachable {
			return fmt.Errorf("unsupported edge activation reason %q", activation.Reason)
		}
		switch {
		case !activation.Evaluation && (activation.State != ActivationPruned || activation.Reason != ReasonConditionFalse):
			return fmt.Errorf("false edge evaluation must be pruned with reason %q", ReasonConditionFalse)
		case activation.Evaluation && activation.State == ActivationSelected && activation.Reason != ReasonConditionTrue:
			return fmt.Errorf("selected true edge evaluation must have reason %q", ReasonConditionTrue)
		case activation.Evaluation && activation.State == ActivationPruned && activation.Reason != ReasonUnreachable:
			return fmt.Errorf("pruned true edge evaluation must have reason %q", ReasonUnreachable)
		}
		if index > 0 && graph.EdgeActivations[index-1].EdgeID >= activation.EdgeID {
			return fmt.Errorf("active edge_activations must be sorted and unique")
		}
	}
	if graph.NonOrderingSCCs == nil {
		return fmt.Errorf("active non_ordering_sccs must be an explicit array")
	}
	previous := ID("")
	for index, scc := range graph.NonOrderingSCCs {
		if err := scc.validate(); err != nil {
			return fmt.Errorf("non_ordering_sccs[%d]: %w", index, err)
		}
		id, _ := scc.ID()
		if index > 0 && previous >= id {
			return fmt.Errorf("active non_ordering_sccs must be sorted and unique")
		}
		previous = id
	}
	return nil
}

// CanonicalBytes returns exact curator-active-graph-v1 CCJ bytes.
func (graph ActiveGraph) CanonicalBytes() ([]byte, error) { return canonicalBytes(graph) }

// ID derives active_graph_id.
func (graph ActiveGraph) ID() (ID, error) { return recordID(graph) }

func (graph ActiveGraph) domainLabel() string { return LabelActiveGraph }
func (graph ActiveGraph) canonicalValue() map[string]any {
	nodes := make([]any, len(graph.NodeActivations))
	for i, activation := range graph.NodeActivations {
		nodes[i] = map[string]any{"node_id": string(activation.NodeID), "state": string(activation.State)}
	}
	edges := make([]any, len(graph.EdgeActivations))
	for i, activation := range graph.EdgeActivations {
		edges[i] = map[string]any{"edge_id": string(activation.EdgeID), "evaluation": activation.Evaluation, "reason": string(activation.Reason), "state": string(activation.State)}
	}
	sccs := make([]any, len(graph.NonOrderingSCCs))
	for i, scc := range graph.NonOrderingSCCs {
		sccs[i] = scc.value()
	}
	return map[string]any{"captured_graph_id": string(graph.CapturedGraphID), "edge_activations": edges, "node_activations": nodes, "non_ordering_sccs": sccs, "schema_id": graph.SchemaID, "selection_binding_id": string(graph.SelectionBindingID), "selection_context_id": string(graph.SelectionContextID)}
}

// DecodeActiveGraph accepts exact canonical active-graph bytes.
func DecodeActiveGraph(payload []byte) (ActiveGraph, error) {
	raw, err := decodeCanonicalObject(payload, "active graph")
	if err != nil {
		return ActiveGraph{}, err
	}
	if err := exactFields(raw, "active graph", []string{"captured_graph_id", "edge_activations", "node_activations", "non_ordering_sccs", "schema_id", "selection_binding_id", "selection_context_id"}, nil); err != nil {
		return ActiveGraph{}, err
	}
	graph := ActiveGraph{}
	graph.SchemaID, err = requiredString(raw, "schema_id", "active graph")
	if err != nil {
		return ActiveGraph{}, err
	}
	capture, err := requiredString(raw, "captured_graph_id", "active graph")
	if err != nil {
		return ActiveGraph{}, err
	}
	graph.CapturedGraphID = ID(capture)
	selection, err := requiredString(raw, "selection_context_id", "active graph")
	if err != nil {
		return ActiveGraph{}, err
	}
	graph.SelectionContextID = ID(selection)
	binding, err := requiredString(raw, "selection_binding_id", "active graph")
	if err != nil {
		return ActiveGraph{}, err
	}
	graph.SelectionBindingID = ID(binding)
	nodesRaw, ok := raw["node_activations"].([]any)
	if !ok {
		return ActiveGraph{}, fmt.Errorf("active graph node_activations must be an array")
	}
	graph.NodeActivations = make([]NodeActivation, len(nodesRaw))
	for i, item := range nodesRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return ActiveGraph{}, fmt.Errorf("node_activations[%d] must be an object", i)
		}
		if err := exactFields(object, "node activation", []string{"node_id", "state"}, nil); err != nil {
			return ActiveGraph{}, err
		}
		id, err := requiredString(object, "node_id", "node activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		state, err := requiredString(object, "state", "node activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		graph.NodeActivations[i] = NodeActivation{NodeID: ID(id), State: ActivationState(state)}
	}
	edgesRaw, ok := raw["edge_activations"].([]any)
	if !ok {
		return ActiveGraph{}, fmt.Errorf("active graph edge_activations must be an array")
	}
	graph.EdgeActivations = make([]EdgeActivation, len(edgesRaw))
	for i, item := range edgesRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return ActiveGraph{}, fmt.Errorf("edge_activations[%d] must be an object", i)
		}
		if err := exactFields(object, "edge activation", []string{"edge_id", "evaluation", "reason", "state"}, nil); err != nil {
			return ActiveGraph{}, err
		}
		id, err := requiredString(object, "edge_id", "edge activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		evaluation, err := requiredBool(object, "evaluation", "edge activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		reason, err := requiredString(object, "reason", "edge activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		state, err := requiredString(object, "state", "edge activation")
		if err != nil {
			return ActiveGraph{}, err
		}
		graph.EdgeActivations[i] = EdgeActivation{EdgeID: ID(id), Evaluation: evaluation, Reason: ActivationReason(reason), State: ActivationState(state)}
	}
	sccRaw, ok := raw["non_ordering_sccs"].([]any)
	if !ok {
		return ActiveGraph{}, fmt.Errorf("active graph non_ordering_sccs must be an array")
	}
	graph.NonOrderingSCCs = make([]NonOrderingSCC, len(sccRaw))
	for i, item := range sccRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return ActiveGraph{}, fmt.Errorf("non_ordering_sccs[%d] must be an object", i)
		}
		if err := exactFields(object, "non-ordering SCC", []string{"edge_ids", "node_ids"}, nil); err != nil {
			return ActiveGraph{}, err
		}
		edges, err := requiredIDSlice(object, "edge_ids", "non-ordering SCC")
		if err != nil {
			return ActiveGraph{}, err
		}
		nodes, err := requiredIDSlice(object, "node_ids", "non-ordering SCC")
		if err != nil {
			return ActiveGraph{}, err
		}
		graph.NonOrderingSCCs[i] = NonOrderingSCC{NodeIDs: nodes, EdgeIDs: edges}
	}
	if err := graph.Validate(); err != nil {
		return ActiveGraph{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, graph); err != nil {
		return ActiveGraph{}, err
	}
	return graph, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
func clonePlatformRoles(values map[PlatformRole]ID) map[PlatformRole]ID {
	if values == nil {
		return nil
	}
	result := make(map[PlatformRole]ID, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func sortNodeActivations(values []NodeActivation) {
	sort.Slice(values, func(i, j int) bool { return values[i].NodeID < values[j].NodeID })
}
func sortSCCs(values []NonOrderingSCC) {
	sort.Slice(values, func(i, j int) bool { left, _ := values[i].ID(); right, _ := values[j].ID(); return left < right })
}
