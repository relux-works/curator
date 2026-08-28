package swiftpmbuild

import (
	"context"
	"sort"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// bindingIndex is the resolved projection of one accepted C4 binding overlay.
type bindingIndex struct {
	nodesByID  map[closuregraph.ID]closuregraph.Node
	idsByRole  map[string]closuregraph.ID
	referenced map[closuregraph.ID]bool
}

// indexBinding resolves the accepted binding records fail-closed. A duplicate
// role, a duplicate node, a wrong-kind node, or a node the selection binding
// never references is rejected before any slot is resolved.
func indexBinding(interop *swiftpminterop.Result) (bindingIndex, error) {
	index := bindingIndex{
		nodesByID:  map[closuregraph.ID]closuregraph.Node{},
		idsByRole:  map[string]closuregraph.ID{},
		referenced: map[closuregraph.ID]bool{},
	}
	for _, id := range interop.Binding.BindingNodeIDs {
		index.referenced[id] = true
	}
	for _, node := range interop.Records.BindingNodes {
		id, err := node.ID()
		if err != nil {
			return bindingIndex{}, failFields(CodeGraphReferenceInvalid, map[string]string{"logical_key": node.LogicalKey}, "accepted binding node is not canonical: %v", err)
		}
		if _, duplicate := index.nodesByID[id]; duplicate {
			return bindingIndex{}, failFields(CodeGraphReferenceInvalid, map[string]string{"node": string(id)}, "accepted binding contains a duplicate node")
		}
		if !index.referenced[id] {
			return bindingIndex{}, failFields(CodeGraphReferenceInvalid, map[string]string{"node": string(id)}, "accepted binding node is not referenced by the selection binding")
		}
		index.nodesByID[id] = node
		if node.Kind != closuregraph.NodeToolchainComponent {
			continue
		}
		role := node.Payload.(closuregraph.ToolchainComponentPayload).ComponentRole
		if _, duplicate := index.idsByRole[role]; duplicate {
			return bindingIndex{}, failFields(CodeGraphReferenceInvalid, map[string]string{"role": role}, "accepted binding binds one component role more than once")
		}
		index.idsByRole[role] = id
	}
	for id := range index.referenced {
		if _, resolved := index.nodesByID[id]; !resolved {
			return bindingIndex{}, failFields(CodeGraphReferenceInvalid, map[string]string{"node": string(id)}, "selection binding references a dangling node")
		}
	}
	return index, nil
}

// resolveSlots binds every required build slot exactly once. The slot map
// names only selection roles; every physical identity is read back from the
// accepted binding node so this stage cannot restate a toolchain fact.
func resolveSlots(config Config, index bindingIndex, requiresCXX bool, linker closuregraph.Node, linkerID closuregraph.ID) (map[ToolSlot]SlotBinding, error) {
	if _, restated := config.Slots[SlotLinker]; restated {
		// The linker is the one component this stage selects itself, so a role
		// for it in the accepted-binding slot map would be silently ignored.
		return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"slot": string(SlotLinker)}, "build binding must not name an accepted-binding role for the self-selected linker slot")
	}
	wanted := append([]ToolSlot(nil), requiredSlots...)
	if requiresCXX {
		wanted = append(wanted, SlotClangCXX)
	}
	sort.Slice(wanted, func(i, j int) bool { return wanted[i] < wanted[j] })
	slots := map[ToolSlot]SlotBinding{}
	seen := map[closuregraph.ID]ToolSlot{}
	for _, slot := range wanted {
		if slot == SlotLinker {
			slots[slot] = SlotBinding{Slot: slot, Role: config.Linker.Role, NodeID: linkerID, Payload: linker.Payload.(closuregraph.ToolchainComponentPayload)}
			seen[linkerID] = slot
			continue
		}
		role, declared := config.Slots[slot]
		if !declared || role == "" {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"slot": string(slot)}, "build binding declares no role for a required tool slot")
		}
		id, resolved := index.idsByRole[role]
		if !resolved {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"role": role, "slot": string(slot)}, "accepted binding has no component for a required tool slot")
		}
		if other, duplicate := seen[id]; duplicate {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"slot": string(slot), "other": string(other)}, "two build tool slots resolve to one binding node")
		}
		node := index.nodesByID[id]
		if node.Kind != closuregraph.NodeToolchainComponent {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"role": role, "kind": string(node.Kind)}, "build tool slot resolves to a wrong-kind binding node")
		}
		seen[id] = slot
		slots[slot] = SlotBinding{Slot: slot, Role: role, NodeID: id, Payload: node.Payload.(closuregraph.ToolchainComponentPayload)}
	}
	for slot := range config.Slots {
		if _, wantedSlot := slots[slot]; !wantedSlot {
			return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"slot": string(slot)}, "build binding declares a slot this selection does not use")
		}
	}
	return slots, nil
}

// recheckSlots proves at the immediate time of use that every bound physical
// component still has the exact identity the accepted binding recorded.
func recheckSlots(ctx context.Context, config Config, binding Binding) error {
	slots := make([]ToolSlot, 0, len(binding.Slots))
	for slot := range binding.Slots {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })
	for _, slot := range slots {
		bound := binding.Slots[slot]
		expected := toolIdentity(bound)
		observed, err := config.Recheck(ctx, expected)
		if err != nil {
			return failFields(CodeToolchainChanged, map[string]string{"slot": string(slot), "role": bound.Role}, "bound build component could not be rechecked: %v", err)
		}
		if observed.Fingerprint != expected.Fingerprint || (expected.ExecutableSHA256.Valid() && observed.ExecutableSHA256 != expected.ExecutableSHA256) {
			return failFields(CodeToolchainChanged, map[string]string{"slot": string(slot), "role": bound.Role}, "bound build component identity drifted before use")
		}
	}
	return nil
}

// toolIdentity reconstructs the recheck subject from the binding node alone.
func toolIdentity(bound SlotBinding) swiftpmsource.ToolIdentity {
	identity := swiftpmsource.ToolIdentity{
		Role: bound.Payload.ComponentRole, ExecutableRelativePath: bound.Payload.ExecutableRelativePath,
		VersionOutput: bound.Payload.VersionOutput, PlatformABI: bound.Payload.PlatformABI,
		PolicySelector: bound.Payload.PolicySelector, Fingerprint: bound.Payload.ContentFingerprint,
	}
	if len(bound.Payload.LinkFingerprintIDs) == 1 {
		identity.ExecutableSHA256 = bound.Payload.LinkFingerprintIDs[0]
	}
	return identity
}

// validateActionBindings proves the exact single-resolution rule: every
// selected action binds one target platform and exactly one component per
// declared tool slot, and every such component is an accepted binding node.
func validateActionBindings(bundle closuregraph.GraphBundle, binding Binding) error {
	nodes := map[closuregraph.ID]closuregraph.Node{}
	for _, node := range append(append([]closuregraph.Node(nil), bundle.Records.CaptureNodes...), bundle.Records.BindingNodes...) {
		id, err := node.ID()
		if err != nil {
			return err
		}
		nodes[id] = node
	}
	selected := map[closuregraph.ID]bool{}
	for _, activation := range bundle.Active.NodeActivations {
		if activation.State == closuregraph.ActivationSelected {
			selected[activation.NodeID] = true
		}
	}
	// The shared projection emits an edge activation for conditional edges
	// only, and graph validation rejects an activation record on an
	// unconditional edge. An unconditional edge is therefore selected exactly
	// when its declaring node is, which is what the loop below asserts.
	pruned := map[closuregraph.ID]bool{}
	for _, activation := range bundle.Active.EdgeActivations {
		if activation.State != closuregraph.ActivationSelected {
			pruned[activation.EdgeID] = true
		}
	}
	tools := map[closuregraph.ID]map[string][]closuregraph.ID{}
	targets := map[closuregraph.ID][]closuregraph.ID{}
	for _, edge := range append(append([]closuregraph.Edge(nil), bundle.Records.CaptureEdges...), bundle.Records.BindingEdges...) {
		edgeID, err := edge.ID()
		if err != nil {
			return err
		}
		if pruned[edgeID] || !selected[edge.FromNodeID] {
			continue
		}
		switch edge.Kind {
		case closuregraph.EdgeUsesTool:
			payload := edge.Payload.(closuregraph.UsesToolPayload)
			component, present := nodes[edge.ToNodeID]
			if !present || component.Kind != closuregraph.NodeToolchainComponent {
				return failFields(CodeGraphReferenceInvalid, map[string]string{"slot": payload.ToolSlot}, "uses_tool edge has a dangling or wrong-kind tool binding")
			}
			if payload.ExecutableRelativePath != component.Payload.(closuregraph.ToolchainComponentPayload).ExecutableRelativePath {
				return failFields(CodeGraphReferenceInvalid, map[string]string{"slot": payload.ToolSlot}, "uses_tool edge names another executable than its bound component")
			}
			if tools[edge.FromNodeID] == nil {
				tools[edge.FromNodeID] = map[string][]closuregraph.ID{}
			}
			tools[edge.FromNodeID][payload.ToolSlot] = append(tools[edge.FromNodeID][payload.ToolSlot], edge.ToNodeID)
		case closuregraph.EdgeTargets:
			if edge.Payload.(closuregraph.TargetsPayload).BindingRole == closuregraph.PlatformTarget {
				targets[edge.FromNodeID] = append(targets[edge.FromNodeID], edge.ToNodeID)
			}
		}
	}
	actionIDs := make([]closuregraph.ID, 0)
	for id, node := range nodes {
		if selected[id] && node.Kind == closuregraph.NodeAction {
			actionIDs = append(actionIDs, id)
		}
	}
	sort.Slice(actionIDs, func(i, j int) bool { return actionIDs[i] < actionIDs[j] })
	for _, actionID := range actionIDs {
		action := nodes[actionID].Payload.(closuregraph.ActionPayload)
		bound := targets[actionID]
		if len(bound) != 1 || bound[0] != binding.PlatformNodeID {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"action": nodes[actionID].LogicalKey}, "selected action must bind exactly one exact target platform, got %d", len(bound))
		}
		if len(tools[actionID]) != len(action.ToolSlotNames) {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"action": nodes[actionID].LogicalKey}, "selected action binds %d tool slots, declared %d", len(tools[actionID]), len(action.ToolSlotNames))
		}
		for _, slot := range action.ToolSlotNames {
			resolved := tools[actionID][slot]
			if len(resolved) != 1 {
				return failFields(CodeGraphReferenceInvalid, map[string]string{"action": nodes[actionID].LogicalKey, "slot": slot}, "declared tool slot resolved %d times", len(resolved))
			}
		}
	}
	return nil
}

// requiresCXXDriver reports whether the exact selection compiles C++ or
// Objective-C++ source and therefore needs a bound C++ driver slot.
func requiresCXXDriver(interop *swiftpminterop.Result) bool {
	for _, target := range interop.Targets {
		if !target.Selected {
			continue
		}
		for _, language := range target.Languages {
			if language == swiftpminterop.LanguageCXX || language == swiftpminterop.LanguageObjCXX {
				return true
			}
		}
	}
	return false
}

// assuranceReads reports whether the accepted interop closure carries observed
// compiler read evidence rather than the portable not-observed verdict.
func assuranceReads(interop *swiftpminterop.Result, mode closureexec.AssuranceMode) error {
	if mode != closureexec.AssuranceVerified {
		return nil
	}
	if interop.Reads.Mode != "observed" || len(interop.Reads.ReceiptIDs) == 0 {
		return fail(CodeHeaderInputUndeclared, "verified build requires an observed compiler read set")
	}
	return nil
}
