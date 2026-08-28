package closuregraph

import (
	"fmt"
	"sort"
	"strings"
)

// OrderingReason is the closed reason for a provider-before-consumer action
// arc in D_build.
type OrderingReason string

const (
	// OrderGeneratedInput and the related constants enumerate the reasons for
	// deterministic action-ordering arcs.
	OrderGeneratedInput OrderingReason = "generated_input"
	// OrderLocalTool orders a producing action before a local-tool consumer.
	OrderLocalTool OrderingReason = "local_tool"
	// OrderInterop orders compile or link providers before interop consumers.
	OrderInterop OrderingReason = "interop"
	// OrderTargetRequirement orders a target producer before its dependent target.
	OrderTargetRequirement OrderingReason = "target_requirement"
)

// OrderingEdge is one derived action DAG arc with all source graph evidence.
type OrderingEdge struct {
	FromActionID  ID
	ToActionID    ID
	Reason        OrderingReason
	SourceEdgeIDs []ID
}

// Validate checks a canonical ordering edge.
func (edge OrderingEdge) Validate() error {
	if err := validateID(edge.FromActionID, "ordering edge from_action_id"); err != nil {
		return err
	}
	if err := validateID(edge.ToActionID, "ordering edge to_action_id"); err != nil {
		return err
	}
	switch edge.Reason {
	case OrderGeneratedInput, OrderLocalTool, OrderInterop, OrderTargetRequirement:
	default:
		return fmt.Errorf("unsupported ordering reason %q", edge.Reason)
	}
	if err := validateIDSlice(edge.SourceEdgeIDs, "ordering edge source_edge_ids", true); err != nil {
		return err
	}
	if len(edge.SourceEdgeIDs) == 0 {
		return fmt.Errorf("ordering edge source_edge_ids must not be empty")
	}
	return nil
}

func (edge OrderingEdge) value() map[string]any {
	return map[string]any{"from_action_id": string(edge.FromActionID), "reason": string(edge.Reason), "source_edge_ids": idsToAny(edge.SourceEdgeIDs), "to_action_id": string(edge.ToActionID)}
}

// ID derives a stable internal ordering-edge identity.
func (edge OrderingEdge) ID() (ID, error) {
	if err := edge.Validate(); err != nil {
		return "", err
	}
	return DomainID(LabelOrderingEdge, edge.value())
}

// BuildPlan is the canonical acyclic D_build projection.
type BuildPlan struct {
	SchemaID              string
	ActiveGraphID         ID
	ActionNodeIDs         []ID
	DeclaredOutputNodeIDs []ID
	ExecutionPolicyID     string
	OrderingEdges         []OrderingEdge
	Waves                 [][]ID
}

// Validate checks action/output sets, ordering records, and exact stable Kahn
// waves independently of adapter discovery order.
func (plan BuildPlan) Validate() error {
	if plan.SchemaID != SchemaBuildPlan {
		return fmt.Errorf("%s: unsupported build plan schema %q", CodeGraphSchemaUnsupported, plan.SchemaID)
	}
	if err := validateID(plan.ActiveGraphID, "build plan active_graph_id"); err != nil {
		return err
	}
	if err := validateIDSlice(plan.ActionNodeIDs, "build plan action_node_ids", true); err != nil {
		return err
	}
	if err := validateIDSlice(plan.DeclaredOutputNodeIDs, "build plan declared_output_node_ids", true); err != nil {
		return err
	}
	if err := validatePortableText(plan.ExecutionPolicyID, "build plan execution_policy_id", false); err != nil {
		return err
	}
	if plan.OrderingEdges == nil {
		return fmt.Errorf("build plan ordering_edges must be an explicit array")
	}
	previous := ID("")
	actions := idSet(plan.ActionNodeIDs)
	for index, edge := range plan.OrderingEdges {
		if err := edge.Validate(); err != nil {
			return fmt.Errorf("ordering_edges[%d]: %w", index, err)
		}
		id, _ := edge.ID()
		if index > 0 && previous >= id {
			return fmt.Errorf("build plan ordering_edges must be sorted and unique")
		}
		previous = id
		if !actions[edge.FromActionID] || !actions[edge.ToActionID] {
			return fmt.Errorf("ordering edge endpoint is absent from action_node_ids")
		}
	}
	waves, cyclic := stableWaves(plan.ActionNodeIDs, plan.OrderingEdges)
	if len(cyclic) > 0 {
		return fmt.Errorf("%s: ordering graph contains action cycle %v", CodeBuildCycle, cyclic)
	}
	if !sameWaves(plan.Waves, waves) {
		return fmt.Errorf("build plan waves do not match stable Kahn projection")
	}
	return nil
}

// CanonicalBytes returns exact curator-build-plan-v1 CCJ bytes.
func (plan BuildPlan) CanonicalBytes() ([]byte, error) { return canonicalBytes(plan) }

// ID derives build_plan_id.
func (plan BuildPlan) ID() (ID, error) { return recordID(plan) }

func (plan BuildPlan) domainLabel() string { return LabelBuildPlan }
func (plan BuildPlan) canonicalValue() map[string]any {
	edges := make([]any, len(plan.OrderingEdges))
	for i, edge := range plan.OrderingEdges {
		edges[i] = edge.value()
	}
	waves := make([]any, len(plan.Waves))
	for i, wave := range plan.Waves {
		waves[i] = idsToAny(wave)
	}
	return map[string]any{"action_node_ids": idsToAny(plan.ActionNodeIDs), "active_graph_id": string(plan.ActiveGraphID), "declared_output_node_ids": idsToAny(plan.DeclaredOutputNodeIDs), "execution_policy_id": plan.ExecutionPolicyID, "ordering_edges": edges, "schema_id": plan.SchemaID, "waves": waves}
}

// DecodeBuildPlan accepts exact canonical build-plan bytes.
func DecodeBuildPlan(payload []byte) (BuildPlan, error) {
	raw, err := decodeCanonicalObject(payload, "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	if err := exactFields(raw, "build plan", []string{"action_node_ids", "active_graph_id", "declared_output_node_ids", "execution_policy_id", "ordering_edges", "schema_id", "waves"}, nil); err != nil {
		return BuildPlan{}, err
	}
	plan := BuildPlan{}
	plan.SchemaID, err = requiredString(raw, "schema_id", "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	active, err := requiredString(raw, "active_graph_id", "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	plan.ActiveGraphID = ID(active)
	plan.ActionNodeIDs, err = requiredIDSlice(raw, "action_node_ids", "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	plan.DeclaredOutputNodeIDs, err = requiredIDSlice(raw, "declared_output_node_ids", "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	plan.ExecutionPolicyID, err = requiredString(raw, "execution_policy_id", "build plan")
	if err != nil {
		return BuildPlan{}, err
	}
	edgesRaw, ok := raw["ordering_edges"].([]any)
	if !ok {
		return BuildPlan{}, fmt.Errorf("build plan ordering_edges must be an array")
	}
	plan.OrderingEdges = make([]OrderingEdge, len(edgesRaw))
	for i, item := range edgesRaw {
		object, ok := item.(map[string]any)
		if !ok {
			return BuildPlan{}, fmt.Errorf("ordering_edges[%d] must be an object", i)
		}
		if err := exactFields(object, "ordering edge", []string{"from_action_id", "reason", "source_edge_ids", "to_action_id"}, nil); err != nil {
			return BuildPlan{}, err
		}
		from, err := requiredString(object, "from_action_id", "ordering edge")
		if err != nil {
			return BuildPlan{}, err
		}
		to, err := requiredString(object, "to_action_id", "ordering edge")
		if err != nil {
			return BuildPlan{}, err
		}
		reason, err := requiredString(object, "reason", "ordering edge")
		if err != nil {
			return BuildPlan{}, err
		}
		evidence, err := requiredIDSlice(object, "source_edge_ids", "ordering edge")
		if err != nil {
			return BuildPlan{}, err
		}
		plan.OrderingEdges[i] = OrderingEdge{FromActionID: ID(from), ToActionID: ID(to), Reason: OrderingReason(reason), SourceEdgeIDs: evidence}
	}
	wavesRaw, ok := raw["waves"].([]any)
	if !ok {
		return BuildPlan{}, fmt.Errorf("build plan waves must be an array")
	}
	plan.Waves = make([][]ID, len(wavesRaw))
	for i, item := range wavesRaw {
		array, ok := item.([]any)
		if !ok {
			return BuildPlan{}, fmt.Errorf("waves[%d] must be an array", i)
		}
		plan.Waves[i] = make([]ID, len(array))
		for j, rawID := range array {
			text, ok := rawID.(string)
			if !ok {
				return BuildPlan{}, fmt.Errorf("waves[%d][%d] must be a string", i, j)
			}
			plan.Waves[i][j] = ID(text)
		}
	}
	if err := plan.Validate(); err != nil {
		return BuildPlan{}, err
	}
	if err := requireDecodedRecordRoundTrip(payload, plan); err != nil {
		return BuildPlan{}, err
	}
	return plan, nil
}

// PlanOptions carries evidence needed only for deterministic cycle reporting.
type PlanOptions struct {
	ExecutionPolicyID string
	LastCheckpointID  ID
}

// BuildCycleError is the canonical closure_build_cycle finding. No BuildPlan
// or C5 checkpoint is issued when this error is returned.
type BuildCycleError struct {
	ActionNodeIDs      []ID
	OrderingEdgeIDs    []ID
	CycleDigest        ID
	AffectedProductIDs []ID
	AffectedTargetIDs  []ID
	LastCheckpointID   ID
}

func (e *BuildCycleError) Error() string {
	return fmt.Sprintf("%s: actions=%s ordering_edges=%s cycle_digest=%s", CodeBuildCycle, joinIDs(e.ActionNodeIDs), joinIDs(e.OrderingEdgeIDs), e.CycleDigest)
}

// DeriveBuildPlan projects stable action waves from a validated active graph.
func DeriveBuildPlan(bundle GraphBundle, options PlanOptions) (BuildPlan, error) {
	if err := bundle.Validate(); err != nil {
		return BuildPlan{}, err
	}
	if err := validatePortableText(options.ExecutionPolicyID, "execution policy ID", false); err != nil {
		return BuildPlan{}, err
	}
	if options.LastCheckpointID != "" {
		if err := validateID(options.LastCheckpointID, "last checkpoint ID"); err != nil {
			return BuildPlan{}, err
		}
	}
	resolved, _ := validateStructure(bundle.Capture, bundle.Selection, bundle.Binding, bundle.Records, bundle.Authority)
	states := map[ID]ActivationState{}
	for _, activation := range bundle.Active.NodeActivations {
		states[activation.NodeID] = activation.State
	}
	actions, outputs := []ID{}, []ID{}
	for id, node := range resolved.captureNodes {
		if states[id] != ActivationSelected {
			continue
		}
		if node.Kind == NodeAction {
			actions = append(actions, id)
		}
		if node.Kind == NodeOutputArtifact {
			outputs = append(outputs, id)
		}
	}
	actions = sortedIDs(actions)
	outputs = sortedIDs(outputs)
	activeEdges := selectedEdges(bundle.Active, resolved)
	if _, err := selectedWritePathsFromEdges(activeEdges, idSet(actions)); err != nil {
		return BuildPlan{}, err
	}
	ordering, err := deriveOrderingEdges(actions, activeEdges, resolved.allNodes)
	if err != nil {
		return BuildPlan{}, err
	}
	waves, cyclic := stableWaves(actions, ordering)
	if len(cyclic) > 0 {
		return BuildPlan{}, newBuildCycleError(cyclic, ordering, activeEdges, resolved.allNodes, options.LastCheckpointID)
	}
	activeID, _ := bundle.Active.ID()
	plan := BuildPlan{SchemaID: SchemaBuildPlan, ActiveGraphID: activeID, ActionNodeIDs: actions, DeclaredOutputNodeIDs: outputs, ExecutionPolicyID: options.ExecutionPolicyID, OrderingEdges: ordering, Waves: waves}
	if err := plan.Validate(); err != nil {
		return BuildPlan{}, err
	}
	return plan, nil
}

type actionEvidence struct {
	ActionID ID
	EdgeIDs  []ID
}

func deriveOrderingEdges(actions []ID, edges map[ID]Edge, nodes map[ID]Node) ([]OrderingEdge, error) {
	actionSet := idSet(actions)
	edgeIDs := make([]ID, 0, len(edges))
	for edgeID := range edges {
		edgeIDs = append(edgeIDs, edgeID)
	}
	edgeIDs = sortedIDs(edgeIDs)
	producers := map[ID][]actionEvidence{}
	for _, edgeID := range edgeIDs {
		edge := edges[edgeID]
		if edge.Kind == EdgeProduces && actionSet[edge.FromNodeID] {
			producers[edge.ToNodeID] = append(producers[edge.ToNodeID], actionEvidence{ActionID: edge.FromNodeID, EdgeIDs: []ID{edgeID}})
		}
	}
	actionsByOwner := declaredActionEvidence(edges, nodes, actionSet)
	artifacts := make([]ID, 0, len(producers))
	for artifact := range producers {
		artifacts = append(artifacts, artifact)
	}
	for _, artifact := range sortedIDs(artifacts) {
		values := producers[artifact]
		if len(values) > 1 {
			ids := make([]string, len(values))
			for i, value := range values {
				ids[i] = string(value.ActionID)
			}
			sort.Strings(ids)
			return nil, fmt.Errorf("%s: output %s has multiple producer actions %s", CodeGraphReferenceInvalid, artifact, strings.Join(ids, ","))
		}
	}
	arcEvidence := map[string]map[ID]bool{}
	add := func(from, to ID, reason OrderingReason, evidence ...ID) {
		key := string(from) + "\x00" + string(to) + "\x00" + string(reason)
		if arcEvidence[key] == nil {
			arcEvidence[key] = map[ID]bool{}
		}
		for _, id := range evidence {
			arcEvidence[key][id] = true
		}
	}
	for _, edgeID := range edgeIDs {
		edge := edges[edgeID]
		switch edge.Kind {
		case EdgeReads:
			artifactProducers := producers[edge.ToNodeID]
			if len(artifactProducers) == 0 {
				continue
			}
			directActionRead := actionSet[edge.FromNodeID]
			var consumerActions []actionEvidence
			if directActionRead {
				consumerActions = []actionEvidence{{ActionID: edge.FromNodeID, EdgeIDs: []ID{}}}
			} else {
				consumerActions = actionsByOwner[edge.FromNodeID]
			}
			if len(consumerActions) == 0 {
				return nil, fmt.Errorf("%s: generated read %s has no selected consumer action", CodeGraphReferenceInvalid, edgeID)
			}
			ordered := false
			for _, producer := range artifactProducers {
				for _, consumer := range consumerActions {
					if !directActionRead && producer.ActionID == consumer.ActionID {
						continue
					}
					evidence := append(append([]ID{}, producer.EdgeIDs...), consumer.EdgeIDs...)
					evidence = append(evidence, edgeID)
					add(producer.ActionID, consumer.ActionID, OrderGeneratedInput, evidence...)
					ordered = true
				}
			}
			if !ordered {
				return nil, fmt.Errorf("%s: generated read %s has no distinct provider-before-consumer action pair", CodeGraphReferenceInvalid, edgeID)
			}
		case EdgeRequires:
			if !actionSet[edge.FromNodeID] {
				continue
			}
			for _, producer := range producers[edge.ToNodeID] {
				add(producer.ActionID, edge.FromNodeID, OrderGeneratedInput, append(producer.EdgeIDs, edgeID)...)
			}
		case EdgeUsesTool:
			if !actionSet[edge.FromNodeID] {
				continue
			}
			if nodes[edge.ToNodeID].Kind == NodeOutputArtifact {
				for _, producer := range producers[edge.ToNodeID] {
					add(producer.ActionID, edge.FromNodeID, OrderLocalTool, append(producer.EdgeIDs, edgeID)...)
				}
			}
		}
	}
	for _, edgeID := range edgeIDs {
		edge := edges[edgeID]
		if edge.Kind != EdgeRequires {
			continue
		}
		payload := edge.Payload.(RequiresPayload)
		if payload.Scope == ScopeRuntime || payload.Scope == ScopePeer {
			continue
		}
		consumers := actionsByOwner[edge.FromNodeID]
		providers := actionsByOwner[edge.ToNodeID]
		if len(producers[edge.ToNodeID]) > 0 && !actionSet[edge.FromNodeID] {
			providers = producers[edge.ToNodeID]
		}
		if node, ok := nodes[edge.FromNodeID]; ok && node.Kind == NodeInteropBoundary {
			consumers = interopConsumerActionEvidence(edge.FromNodeID, edgeIDs, edges, producers, actionsByOwner, actionSet)
		}
		for _, provider := range providers {
			for _, consumer := range consumers {
				if provider.ActionID == consumer.ActionID {
					continue
				}
				evidence := append(append([]ID{}, provider.EdgeIDs...), consumer.EdgeIDs...)
				evidence = append(evidence, edgeID)
				add(provider.ActionID, consumer.ActionID, OrderTargetRequirement, evidence...)
			}
		}
	}
	type boundarySide struct {
		actions []actionEvidence
		edgeID  ID
	}
	providersByBoundary := map[ID][]boundarySide{}
	consumersByBoundary := map[ID][]boundarySide{}
	for _, edgeID := range edgeIDs {
		edge := edges[edgeID]
		if edge.Kind != EdgeProvidesInterop && edge.Kind != EdgeConsumesInterop {
			continue
		}
		boundary := nodes[edge.ToNodeID]
		if boundary.Kind != NodeInteropBoundary {
			continue
		}
		mode := boundary.Payload.(InteropBoundaryPayload).Mode
		if mode == InteropDynamicLoad || mode == InteropSubprocessProtocol {
			continue
		}
		sideActions := actionEvidenceForNode(edge.FromNodeID, edgeID, producers, actionsByOwner, actionSet)
		side := boundarySide{actions: sideActions, edgeID: edgeID}
		if edge.Kind == EdgeProvidesInterop {
			providersByBoundary[edge.ToNodeID] = append(providersByBoundary[edge.ToNodeID], side)
		} else {
			consumersByBoundary[edge.ToNodeID] = append(consumersByBoundary[edge.ToNodeID], side)
		}
	}
	for boundary, providerSides := range providersByBoundary {
		for _, providerSide := range providerSides {
			for _, consumerSide := range consumersByBoundary[boundary] {
				for _, provider := range providerSide.actions {
					for _, consumer := range consumerSide.actions {
						evidence := append(append([]ID{}, provider.EdgeIDs...), consumer.EdgeIDs...)
						evidence = append(evidence, providerSide.edgeID, consumerSide.edgeID)
						add(provider.ActionID, consumer.ActionID, OrderInterop, evidence...)
					}
				}
			}
		}
	}
	result := make([]OrderingEdge, 0, len(arcEvidence))
	for key, evidenceSet := range arcEvidence {
		parts := strings.Split(key, "\x00")
		evidence := make([]ID, 0, len(evidenceSet))
		for id := range evidenceSet {
			evidence = append(evidence, id)
		}
		result = append(result, OrderingEdge{FromActionID: ID(parts[0]), ToActionID: ID(parts[1]), Reason: OrderingReason(parts[2]), SourceEdgeIDs: sortedIDs(evidence)})
	}
	sort.Slice(result, func(i, j int) bool { left, _ := result[i].ID(); right, _ := result[j].ID(); return left < right })
	return result, nil
}

type writePathBinding struct {
	outputNodeID ID
	producesID   ID
}

func selectedWritePathsFromEdges(edges map[ID]Edge, actionSet map[ID]bool) ([]string, error) {
	bindings := map[string][]writePathBinding{}
	edgeIDs := make([]ID, 0, len(edges))
	for edgeID := range edges {
		edgeIDs = append(edgeIDs, edgeID)
	}
	for _, edgeID := range sortedIDs(edgeIDs) {
		edge := edges[edgeID]
		if edge.Kind != EdgeProduces || !actionSet[edge.FromNodeID] {
			continue
		}
		path := edge.Payload.(ProducesPayload).Path
		bindings[path] = append(bindings[path], writePathBinding{outputNodeID: edge.ToNodeID, producesID: edgeID})
	}
	paths := sortedMapKeys(bindings)
	for _, path := range paths {
		pathBindings := bindings[path]
		if len(pathBindings) < 2 {
			continue
		}
		outputIDs := make([]ID, len(pathBindings))
		producesIDs := make([]ID, len(pathBindings))
		for index, binding := range pathBindings {
			outputIDs[index] = binding.outputNodeID
			producesIDs[index] = binding.producesID
		}
		collector := &issueCollector{}
		collector.add(CodeGraphReferenceInvalid, path, "active", "produces.path", "selected write path %q has multiple output nodes %s and produces edges %s", path, joinIDs(sortedIDs(outputIDs)), joinIDs(sortedIDs(producesIDs)))
		return nil, collector.err()
	}
	return paths, nil
}

func interopConsumerActionEvidence(boundaryID ID, edgeIDs []ID, edges map[ID]Edge, producers, actionsByOwner map[ID][]actionEvidence, actionSet map[ID]bool) []actionEvidence {
	result := []actionEvidence{}
	for _, edgeID := range edgeIDs {
		edge := edges[edgeID]
		if edge.Kind != EdgeConsumesInterop || edge.ToNodeID != boundaryID {
			continue
		}
		result = append(result, actionEvidenceForNode(edge.FromNodeID, edgeID, producers, actionsByOwner, actionSet)...)
	}
	return result
}

func declaredActionEvidence(edges map[ID]Edge, nodes map[ID]Node, actionSet map[ID]bool) map[ID][]actionEvidence {
	type declaration struct {
		to     ID
		edgeID ID
	}
	adjacency := map[ID][]declaration{}
	for edgeID, edge := range edges {
		if edge.Kind == EdgeDeclares {
			adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], declaration{to: edge.ToNodeID, edgeID: edgeID})
		}
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool {
			if adjacency[id][i].to != adjacency[id][j].to {
				return adjacency[id][i].to < adjacency[id][j].to
			}
			return adjacency[id][i].edgeID < adjacency[id][j].edgeID
		})
	}
	owners := make([]ID, 0, len(nodes))
	for id := range nodes {
		owners = append(owners, id)
	}
	owners = sortedIDs(owners)
	result := map[ID][]actionEvidence{}
	for _, owner := range owners {
		if actionSet[owner] {
			result[owner] = []actionEvidence{{ActionID: owner, EdgeIDs: []ID{}}}
			continue
		}
		evidenceByNode := map[ID]map[ID]bool{owner: {}}
		pending := []ID{owner}
		actionEvidenceIDs := map[ID]map[ID]bool{}
		for len(pending) > 0 {
			current := pending[0]
			pending = pending[1:]
			for _, declaration := range adjacency[current] {
				candidate := cloneIDSet(evidenceByNode[current])
				candidate[declaration.edgeID] = true
				if actionSet[declaration.to] {
					if actionEvidenceIDs[declaration.to] == nil {
						actionEvidenceIDs[declaration.to] = map[ID]bool{}
					}
					mergeIDSets(actionEvidenceIDs[declaration.to], candidate)
					continue
				}
				if evidenceByNode[declaration.to] == nil {
					evidenceByNode[declaration.to] = map[ID]bool{}
				}
				if mergeIDSets(evidenceByNode[declaration.to], candidate) {
					pending = append(pending, declaration.to)
				}
			}
		}
		actionIDs := make([]ID, 0, len(actionEvidenceIDs))
		for actionID := range actionEvidenceIDs {
			actionIDs = append(actionIDs, actionID)
		}
		actionIDs = sortedIDs(actionIDs)
		for _, actionID := range actionIDs {
			edgeIDs := make([]ID, 0, len(actionEvidenceIDs[actionID]))
			for edgeID := range actionEvidenceIDs[actionID] {
				edgeIDs = append(edgeIDs, edgeID)
			}
			result[owner] = append(result[owner], actionEvidence{ActionID: actionID, EdgeIDs: sortedIDs(edgeIDs)})
		}
	}
	return result
}

func cloneIDSet(values map[ID]bool) map[ID]bool {
	result := make(map[ID]bool, len(values)+1)
	for id := range values {
		result[id] = true
	}
	return result
}

func mergeIDSets(destination, source map[ID]bool) bool {
	changed := false
	for id := range source {
		if !destination[id] {
			destination[id] = true
			changed = true
		}
	}
	return changed
}

func actionEvidenceForNode(nodeID, relationshipEdgeID ID, producers map[ID][]actionEvidence, actionsByTarget map[ID][]actionEvidence, actionSet map[ID]bool) []actionEvidence {
	if actionSet[nodeID] {
		return []actionEvidence{{ActionID: nodeID, EdgeIDs: []ID{relationshipEdgeID}}}
	}
	if values := producers[nodeID]; len(values) > 0 {
		return values
	}
	return actionsByTarget[nodeID]
}

func stableWaves(actions []ID, edges []OrderingEdge) ([][]ID, []ID) {
	indegree := map[ID]int{}
	outgoing := map[ID]map[ID]bool{}
	for _, action := range actions {
		indegree[action] = 0
	}
	for _, edge := range edges {
		if outgoing[edge.FromActionID] == nil {
			outgoing[edge.FromActionID] = map[ID]bool{}
		}
		if !outgoing[edge.FromActionID][edge.ToActionID] {
			outgoing[edge.FromActionID][edge.ToActionID] = true
			indegree[edge.ToActionID]++
		}
	}
	remaining := idSet(actions)
	waves := [][]ID{}
	for len(remaining) > 0 {
		wave := []ID{}
		for action := range remaining {
			if indegree[action] == 0 {
				wave = append(wave, action)
			}
		}
		sort.Slice(wave, func(i, j int) bool { return wave[i] < wave[j] })
		if len(wave) == 0 {
			return nil, firstOrderingCycle(remaining, edges)
		}
		waves = append(waves, wave)
		for _, action := range wave {
			delete(remaining, action)
			for next := range outgoing[action] {
				indegree[next]--
			}
		}
	}
	return waves, nil
}

func firstOrderingCycle(remaining map[ID]bool, edges []OrderingEdge) []ID {
	adjacency := map[ID][]ID{}
	selfLoops := map[ID]bool{}
	for _, edge := range edges {
		if !remaining[edge.FromActionID] || !remaining[edge.ToActionID] {
			continue
		}
		adjacency[edge.FromActionID] = append(adjacency[edge.FromActionID], edge.ToActionID)
		if edge.FromActionID == edge.ToActionID {
			selfLoops[edge.FromActionID] = true
		}
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i] < adjacency[id][j] })
	}
	nodes := make([]ID, 0, len(remaining))
	for id := range remaining {
		nodes = append(nodes, id)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
	index := 0
	indexes := map[ID]int{}
	lowlink := map[ID]int{}
	onStack := map[ID]bool{}
	stack := []ID{}
	cycles := [][]ID{}
	var visit func(ID)
	visit = func(node ID) {
		indexes[node], lowlink[node] = index, index
		index++
		stack = append(stack, node)
		onStack[node] = true
		for _, next := range adjacency[node] {
			nextIndex, seen := indexes[next]
			if !seen {
				visit(next)
				if lowlink[next] < lowlink[node] {
					lowlink[node] = lowlink[next]
				}
			} else if onStack[next] && nextIndex < lowlink[node] {
				lowlink[node] = nextIndex
			}
		}
		if lowlink[node] != indexes[node] {
			return
		}
		component := []ID{}
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		component = sortedIDs(component)
		if len(component) > 1 || selfLoops[component[0]] {
			cycles = append(cycles, component)
		}
	}
	for _, node := range nodes {
		if _, seen := indexes[node]; !seen {
			visit(node)
		}
	}
	if len(cycles) == 0 {
		return nodes
	}
	sort.Slice(cycles, func(i, j int) bool { return cycles[i][0] < cycles[j][0] })
	return cycles[0]
}

func newBuildCycleError(cyclic []ID, edges []OrderingEdge, graphEdges map[ID]Edge, nodes map[ID]Node, last ID) *BuildCycleError {
	cyclicSet := idSet(cyclic)
	orderingIDs := []ID{}
	for _, edge := range edges {
		if cyclicSet[edge.FromActionID] && cyclicSet[edge.ToActionID] {
			id, _ := edge.ID()
			orderingIDs = append(orderingIDs, id)
		}
	}
	orderingIDs = sortedIDs(orderingIDs)
	digest, _ := DomainID(LabelBuildCycle, map[string]any{"action_node_ids": idsToAny(cyclic), "ordering_edge_ids": idsToAny(orderingIDs)})
	connected := affectedCycleReachability(cyclic, graphEdges, nodes)
	products, targets := []ID{}, []ID{}
	for id, node := range nodes {
		if !connected[id] {
			continue
		}
		if node.Kind == NodeCommandProduct {
			products = append(products, id)
		}
		if node.Kind == NodeTargetUnit {
			targets = append(targets, id)
		}
	}
	return &BuildCycleError{ActionNodeIDs: sortedIDs(cyclic), OrderingEdgeIDs: orderingIDs, CycleDigest: digest, AffectedProductIDs: sortedIDs(products), AffectedTargetIDs: sortedIDs(targets), LastCheckpointID: last}
}

// affectedCycleReachability follows only causal build ownership and
// consumption. In particular, target-platform and external-toolchain hubs are
// preconditions, not evidence that every node sharing them is affected by a
// cycle.
func affectedCycleReachability(cyclic []ID, graphEdges map[ID]Edge, nodes map[ID]Node) map[ID]bool {
	adjacency := map[ID][]ID{}
	for _, edge := range graphEdges {
		switch payload := edge.Payload.(type) {
		case DeclaresPayload:
			adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
		case RequiresPayload:
			if payload.Scope != ScopeRuntime && payload.Scope != ScopePeer {
				adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
			}
		case ReadsPayload, UsesToolPayload:
			if nodes[edge.ToNodeID].Kind != NodeToolchainComponent {
				adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
			}
		case ProducesPayload:
			adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], edge.ToNodeID)
		case ProvidesInteropPayload:
			adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], edge.ToNodeID)
		case ConsumesInteropPayload:
			boundary, ok := nodes[edge.ToNodeID]
			if ok && boundary.Kind == NodeInteropBoundary {
				mode := boundary.Payload.(InteropBoundaryPayload).Mode
				if mode != InteropDynamicLoad && mode != InteropSubprocessProtocol {
					adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
				}
			}
		case PublishesPayload:
			adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
		}
	}
	for id := range adjacency {
		adjacency[id] = sortedIDs(adjacency[id])
	}
	queue := sortedIDs(cyclic)
	reached := map[ID]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if reached[id] {
			continue
		}
		reached[id] = true
		for _, next := range adjacency[id] {
			if !reached[next] {
				queue = append(queue, next)
			}
		}
	}
	return reached
}

func sameWaves(left, right [][]ID) bool {
	if left == nil || right == nil || len(left) != len(right) {
		return left != nil && right != nil && len(left) == 0 && len(right) == 0
	}
	for i := range left {
		if len(left[i]) != len(right[i]) {
			return false
		}
		for j := range left[i] {
			if left[i][j] != right[i][j] {
				return false
			}
		}
	}
	return true
}
func joinIDs(ids []ID) string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = string(id)
	}
	return strings.Join(values, ",")
}
