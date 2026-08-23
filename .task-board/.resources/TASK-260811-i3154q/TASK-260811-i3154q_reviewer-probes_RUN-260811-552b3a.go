package closuregraph

import "testing"

func TestReviewerProbeRejectsFallbackRoleAliasOnTargetOnlyProduct(t *testing.T) {
	input := cgp10Inputs(t)
	productID := nodeIDByKind(t, input.records.CaptureNodes, NodeCommandProduct)
	originalBindingID, err := input.binding.ID()
	if err != nil {
		t.Fatal(err)
	}

	edges := append([]Edge{}, input.records.BindingEdges...)
	found := false
	for index := range edges {
		if edges[index].Kind != EdgeTargets || edges[index].FromNodeID != productID {
			continue
		}
		payload := edges[index].Payload.(TargetsPayload)
		payload.BindingRole = PlatformHost
		payload.Origin = EvidenceOrigin{Field: "selection.platform_roles.host.alias"}
		edges[index].EdgeKey = "edge:reviewer-product-targets-host-alias"
		edges[index].Payload = payload
		found = true
		break
	}
	if !found {
		t.Fatal("CGP10 product target binding not found")
	}

	input = rebuildBinding(t, input, input.records.BindingNodes, edges)
	aliasedBindingID, err := input.binding.ID()
	if err != nil {
		t.Fatal(err)
	}
	if originalBindingID == aliasedBindingID {
		t.Fatal("probe did not change the binding identity")
	}
	if _, err := projectInputs(t, input); err == nil {
		t.Fatalf("ProjectActive accepted host as a role alias for a target-only product; equivalent bindings have distinct IDs %s and %s", originalBindingID, aliasedBindingID)
	}
}

func TestReviewerProbeNodePayloadPointersFailClosedWithoutPanic(t *testing.T) {
	input := cgp10Inputs(t)
	nodes := append([]Node{}, input.records.CaptureNodes...)
	for index := range nodes {
		if nodes[index].Kind != NodeAction {
			continue
		}
		payload := nodes[index].Payload.(ActionPayload)
		nodes[index].Payload = &payload
		break
	}
	input.records = NewRecordTables(nodes, input.records.CaptureEdges, input.records.BindingNodes, input.records.BindingEdges)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ProjectActive panicked on a pointer form of a closed action payload: %v", recovered)
		}
	}()
	if _, err := projectInputs(t, input); err == nil {
		t.Fatal("ProjectActive accepted a pointer form of a closed action payload instead of rejecting the noncanonical dynamic representation")
	}
}

func TestReviewerProbeEdgePayloadPointersFailClosedWithoutPanic(t *testing.T) {
	input := cgp10Inputs(t)
	edges := append([]Edge{}, input.records.BindingEdges...)
	for index := range edges {
		if edges[index].Kind != EdgeTargets {
			continue
		}
		payload := edges[index].Payload.(TargetsPayload)
		edges[index].Payload = &payload
		break
	}
	input.records = NewRecordTables(input.records.CaptureNodes, input.records.CaptureEdges, input.records.BindingNodes, edges)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ProjectActive panicked on a pointer form of a closed targets payload: %v", recovered)
		}
	}()
	if _, err := projectInputs(t, input); err == nil {
		t.Fatal("ProjectActive accepted a pointer form of a closed targets payload instead of rejecting the noncanonical dynamic representation")
	}
}

func TestReviewerProbeCheckpointPayloadPointersFailClosedWithoutPanic(t *testing.T) {
	payloads := checkpointPayloadFixtures()
	c0 := payloads[0].(C0ProfilePayload)
	payloads[0] = &c0
	chain := buildCheckpointChain(t, payloads)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("checkpoint validation panicked on a pointer form of a closed C0 payload: %v", recovered)
		}
	}()
	if err := validateCheckpointSequence(chain); err == nil {
		t.Fatal("checkpoint validation accepted a pointer form of a closed C0 payload instead of rejecting the noncanonical dynamic representation")
	}
}
