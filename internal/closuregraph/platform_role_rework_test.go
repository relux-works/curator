package closuregraph

import (
	"errors"
	"strings"
	"testing"
)

func TestPlatformBindingRejectsHostAliasForTargetOnlyDeclaration(t *testing.T) {
	input := cgp10Inputs(t)
	productID := nodeIDByKind(t, input.records.CaptureNodes, NodeCommandProduct)
	originalBindingID, err := input.binding.ID()
	if err != nil {
		t.Fatal(err)
	}

	edges := append([]Edge{}, input.records.BindingEdges...)
	for index := range edges {
		if edges[index].Kind != EdgeTargets || edges[index].FromNodeID != productID {
			continue
		}
		payload := edges[index].Payload.(TargetsPayload)
		payload.BindingRole = PlatformHost
		payload.Origin = EvidenceOrigin{Field: "selection.platform_roles.host.alias"}
		edges[index].EdgeKey = "edge:product-targets-host-alias"
		edges[index].Payload = payload
		break
	}
	input = rebuildBinding(t, input, input.records.BindingNodes, edges)
	aliasedBindingID, err := input.binding.ID()
	if err != nil {
		t.Fatal(err)
	}
	if originalBindingID == aliasedBindingID {
		t.Fatal("host alias did not create the distinct identity exercised by this regression")
	}

	_, err = projectInputs(t, input)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("host alias error = %T %v, want ValidationError", err, err)
	}
	foundUndeclaredHost, foundMissingTarget := false, false
	for _, issue := range validation.Issues {
		if issue.Code != CodeGraphReferenceInvalid {
			continue
		}
		if issue.Path == "targets.binding_role" &&
			strings.Contains(issue.Message, "raw platform role \"host\"") &&
			strings.Contains(issue.Message, "declared raw roles are [target]") {
			foundUndeclaredHost = true
		}
		if issue.Path == "targets" &&
			strings.Contains(issue.Message, "raw platform role \"target\" must be bound exactly once, got 0") {
			foundMissingTarget = true
		}
	}
	if !foundUndeclaredHost || !foundMissingTarget {
		t.Fatalf("host alias issues = %#v, want undeclared raw host plus missing exact target slot", validation.Issues)
	}
}

func TestPlatformBindingsCountRawHostAndTargetSlotsSeparately(t *testing.T) {
	product := fixtureProduct("raw-role-slots")
	payload := product.Payload.(CommandProductPayload)
	payload.PlatformRoleNames = []PlatformRole{PlatformHost, PlatformTarget}
	product.Payload = payload
	platform := Node{Kind: NodeTargetPlatform, LogicalKey: "platform:raw-role-slots", Payload: TargetPlatformPayload{
		OS: "linux", Architecture: "x86_64", ABI: "gnu", Libc: "glibc",
		MinimumRuntime: "glibc-2.31", SDKID: "linux-sysroot-v1", TargetTriple: "x86_64-unknown-linux-gnu",
	}}
	productID, platformID := mustNodeID(t, product), mustNodeID(t, platform)
	targetEdge := Edge{
		Kind: EdgeTargets, EdgeKey: "edge:raw-role-slots:target", FromNodeID: productID, ToNodeID: platformID,
		Payload: TargetsPayload{BindingRole: PlatformTarget, Origin: EvidenceOrigin{Field: "selection.platform_roles.target"}},
	}
	hostEdge := Edge{
		Kind: EdgeTargets, EdgeKey: "edge:raw-role-slots:host", FromNodeID: productID, ToNodeID: platformID,
		Payload: TargetsPayload{BindingRole: PlatformHost, Origin: EvidenceOrigin{Field: "selection.platform_roles.host.fallback"}},
	}
	resolved := resolvedTables{
		captureNodes: map[ID]Node{productID: product},
		bindingNodes: map[ID]Node{platformID: platform},
		allNodes:     map[ID]Node{productID: product, platformID: platform},
	}
	selected := map[ID]Edge{mustEdgeID(t, targetEdge): targetEdge, mustEdgeID(t, hostEdge): hostEdge}
	collector := &issueCollector{}
	validatePlatformBindings(collector, map[ID]ActivationState{productID: ActivationSelected}, selected, resolved)
	if err := collector.err(); err != nil {
		t.Fatalf("distinct raw host/target slots sharing the fallback platform were rejected: %v", err)
	}
}
