package closuregraph

import (
	"fmt"
	"sort"
)

// EvaluationInput gives an ecosystem evaluator the immutable selection and
// graph-local records needed to evaluate its own versioned expression syntax.
type EvaluationInput struct {
	Selection SelectionContext
	Records   RecordTables
}

// ConditionEvaluator evaluates one versioned condition language. Adapters own
// ecosystem syntax; the shared projector owns completeness and canonical state.
type ConditionEvaluator interface {
	ID() string
	Evaluate(Condition, EvaluationInput) (bool, error)
}

// ConditionEvaluatorFunc is a convenient sealed-registry adapter for tests and
// ecosystem implementations whose evaluator is naturally a function.
type ConditionEvaluatorFunc struct {
	EvaluatorID  string
	EvaluateFunc func(Condition, EvaluationInput) (bool, error)
}

// ID returns the evaluator identity committed by SelectionContext.
func (e ConditionEvaluatorFunc) ID() string { return e.EvaluatorID }

// Evaluate delegates to EvaluateFunc.
func (e ConditionEvaluatorFunc) Evaluate(condition Condition, input EvaluationInput) (bool, error) {
	if e.EvaluateFunc == nil {
		return false, fmt.Errorf("condition evaluator %q has no implementation", e.EvaluatorID)
	}
	return e.EvaluateFunc(condition, input)
}

// ProjectActive validates capture/context/binding structure, evaluates every
// conditional edge exactly once, computes reachability and permitted SCCs,
// then validates the resulting complete C4 graph bundle.
func ProjectActive(capture CaptureGraph, selection SelectionContext, binding SelectionBinding, records RecordTables, authority BindingAuthority, evaluators []ConditionEvaluator) (GraphBundle, error) {
	resolved, err := validateStructure(capture, selection, binding, records, authority)
	if err != nil {
		return GraphBundle{}, err
	}
	type evaluatorRegistration struct {
		id        string
		evaluator ConditionEvaluator
	}
	registrations := make([]evaluatorRegistration, len(evaluators))
	for index, evaluator := range evaluators {
		registration := evaluatorRegistration{evaluator: evaluator}
		if evaluator != nil {
			registration.id = evaluator.ID()
		}
		registrations[index] = registration
	}
	sort.Slice(registrations, func(i, j int) bool {
		leftNil := registrations[i].evaluator == nil
		rightNil := registrations[j].evaluator == nil
		if leftNil != rightNil {
			return leftNil
		}
		return registrations[i].id < registrations[j].id
	})
	registry := map[string]ConditionEvaluator{}
	for _, registration := range registrations {
		if registration.evaluator == nil {
			return GraphBundle{}, fmt.Errorf("%s: nil condition evaluator", CodeGraphReferenceInvalid)
		}
		id := registration.id
		if err := validatePortableText(id, "evaluator ID", false); err != nil {
			return GraphBundle{}, err
		}
		if _, exists := registry[id]; exists {
			return GraphBundle{}, fmt.Errorf("%s: duplicate evaluator %q", CodeGraphReferenceInvalid, id)
		}
		registry[id] = registration.evaluator
	}
	allowedEvaluators := make(map[string]bool, len(selection.EvaluatorIDs))
	for _, id := range selection.EvaluatorIDs {
		allowedEvaluators[id] = true
	}
	registryIDs := make([]string, 0, len(registry))
	for id := range registry {
		registryIDs = append(registryIDs, id)
	}
	sort.Strings(registryIDs)
	for _, id := range registryIDs {
		if !allowedEvaluators[id] {
			return GraphBundle{}, fmt.Errorf("%s: evaluator %q is not selected", CodeGraphReferenceInvalid, id)
		}
	}

	input := EvaluationInput{Selection: selection, Records: records}
	evaluations := map[ID]bool{}
	conditionalIDs := make([]ID, 0)
	for id, edge := range resolved.captureEdges {
		if edge.Payload.condition() != nil {
			conditionalIDs = append(conditionalIDs, id)
		}
	}
	sort.Slice(conditionalIDs, func(i, j int) bool { return conditionalIDs[i] < conditionalIDs[j] })
	for _, id := range conditionalIDs {
		condition := resolved.captureEdges[id].Payload.condition()
		evaluator, present := registry[condition.EvaluatorID]
		if !present {
			return GraphBundle{}, fmt.Errorf("%s: conditional edge %s requires missing evaluator %q", CodeGraphIncomplete, id, condition.EvaluatorID)
		}
		result, err := evaluator.Evaluate(*condition, input)
		if err != nil {
			return GraphBundle{}, fmt.Errorf("evaluate edge %s with %s: %w", id, condition.EvaluatorID, err)
		}
		evaluations[id] = result
	}

	usable := map[ID]Edge{}
	for id, edge := range resolved.captureEdges {
		if condition := edge.Payload.condition(); condition == nil || evaluations[id] {
			usable[id] = edge
		}
	}
	for id, edge := range resolved.bindingEdges {
		usable[id] = edge
	}
	reachable := selectionReachability(selection.ProductNodeIDs, usable)

	nodeActivations := make([]NodeActivation, 0, len(resolved.captureNodes))
	for id := range resolved.captureNodes {
		state := ActivationPruned
		if reachable[id] {
			state = ActivationSelected
		}
		nodeActivations = append(nodeActivations, NodeActivation{NodeID: id, State: state})
	}
	sortNodeActivations(nodeActivations)
	edgeActivations := make([]EdgeActivation, 0, len(conditionalIDs))
	for _, id := range conditionalIDs {
		edge := resolved.captureEdges[id]
		evaluation := evaluations[id]
		state := ActivationPruned
		reason := ReasonConditionFalse
		if evaluation {
			if reachable[edge.FromNodeID] && reachable[edge.ToNodeID] {
				state = ActivationSelected
				reason = ReasonConditionTrue
			} else {
				reason = ReasonUnreachable
			}
		}
		edgeActivations = append(edgeActivations, EdgeActivation{EdgeID: id, Evaluation: evaluation, Reason: reason, State: state})
	}

	captureID, _ := capture.ID()
	selectionID, _ := selection.ID()
	bindingID, _ := binding.ID()
	active := ActiveGraph{SchemaID: SchemaActiveGraph, CapturedGraphID: captureID, SelectionContextID: selectionID, SelectionBindingID: bindingID, NodeActivations: nodeActivations, EdgeActivations: edgeActivations, NonOrderingSCCs: []NonOrderingSCC{}}
	selected := selectedEdges(active, resolved)
	active.NonOrderingSCCs = deriveNonOrderingSCCs(selected, resolved.allNodes)
	bundle := GraphBundle{Capture: capture, Selection: selection, Binding: binding, Active: active, Records: records, Authority: authority}
	if err := bundle.Validate(); err != nil {
		return GraphBundle{}, err
	}
	return bundle, nil
}

// selectionReachability follows the semantic direction of closure edges from
// each requested product to its requirements. Producer relationships are the
// only reverse traversals: reaching a produced artifact or an interop boundary
// also selects the action or target that provides it. Treating every edge as
// undirected would incorrectly select an unrelated product merely because it
// shares a dependency or target platform with a requested product.
func selectionReachability(roots []ID, edges map[ID]Edge) map[ID]bool {
	adjacency := map[ID][]ID{}
	for _, edge := range edges {
		adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], edge.ToNodeID)
		if edge.Kind == EdgeProduces || edge.Kind == EdgeProvidesInterop {
			adjacency[edge.ToNodeID] = append(adjacency[edge.ToNodeID], edge.FromNodeID)
		}
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i] < adjacency[id][j] })
	}
	queue := sortedIDs(roots)
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

func deriveNonOrderingSCCs(edges map[ID]Edge, nodes map[ID]Node) []NonOrderingSCC {
	nonOrdering := map[ID]Edge{}
	for id, edge := range edges {
		if edgeIsNonOrdering(edge, nodes) {
			nonOrdering[id] = edge
		}
	}
	components := stronglyConnectedComponents(nonOrdering)
	result := make([]NonOrderingSCC, 0)
	for _, component := range components {
		members := idSet(component)
		internal := make([]ID, 0)
		selfLoop := false
		for id, edge := range nonOrdering {
			if members[edge.FromNodeID] && members[edge.ToNodeID] {
				internal = append(internal, id)
				if edge.FromNodeID == edge.ToNodeID {
					selfLoop = true
				}
			}
		}
		if len(component) < 2 && !selfLoop {
			continue
		}
		result = append(result, NonOrderingSCC{NodeIDs: sortedIDs(component), EdgeIDs: sortedIDs(internal)})
	}
	sortSCCs(result)
	return result
}

func edgeIsNonOrdering(edge Edge, nodes map[ID]Node) bool {
	switch payload := edge.Payload.(type) {
	case RequiresPayload:
		return payload.Scope == ScopeRuntime || payload.Scope == ScopePeer
	case UsesToolPayload:
		return nodes[edge.ToNodeID].Kind == NodeToolchainComponent
	case ProvidesInteropPayload, ConsumesInteropPayload:
		boundary, ok := nodes[edge.ToNodeID]
		if !ok || boundary.Kind != NodeInteropBoundary {
			return false
		}
		mode := boundary.Payload.(InteropBoundaryPayload).Mode
		return mode == InteropDynamicLoad || mode == InteropSubprocessProtocol
	case ProducesPayload:
		return false
	default:
		return edge.Kind == EdgeDeclares || edge.Kind == EdgeResolvesTo || edge.Kind == EdgeReads || edge.Kind == EdgeTargets || edge.Kind == EdgeInvokes || edge.Kind == EdgePublishes
	}
}

func stronglyConnectedComponents(edges map[ID]Edge) [][]ID {
	nodeSet := map[ID]bool{}
	adjacency := map[ID][]ID{}
	for _, edge := range edges {
		nodeSet[edge.FromNodeID] = true
		nodeSet[edge.ToNodeID] = true
		adjacency[edge.FromNodeID] = append(adjacency[edge.FromNodeID], edge.ToNodeID)
	}
	nodes := make([]ID, 0, len(nodeSet))
	for id := range nodeSet {
		nodes = append(nodes, id)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return adjacency[id][i] < adjacency[id][j] })
	}
	index := 0
	indexes := map[ID]int{}
	lowlink := map[ID]int{}
	onStack := map[ID]bool{}
	stack := []ID{}
	components := [][]ID{}
	var visit func(ID)
	visit = func(node ID) {
		indexes[node] = index
		lowlink[node] = index
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
		sort.Slice(component, func(i, j int) bool { return component[i] < component[j] })
		components = append(components, component)
	}
	for _, node := range nodes {
		if _, seen := indexes[node]; !seen {
			visit(node)
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i][0] < components[j][0] })
	return components
}
