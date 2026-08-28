package closuregraph

import "testing"

func TestReviewerProbeRejectsUndeclaredExtraHostTarget(t *testing.T) {
	input := cgp10Inputs(t)
	host := input.records.BindingNodes[0]
	host.LogicalKey = "platform:reviewer-host-linux-x86_64"
	host.Payload = TargetPlatformPayload{
		OS:             "linux",
		Architecture:   "x86_64",
		ABI:            "gnu",
		Libc:           "glibc",
		MinimumRuntime: "glibc-2.31",
		SDKID:          "linux-sysroot-v1",
		TargetTriple:   "x86_64-unknown-linux-gnu",
	}
	hostID := mustNodeID(t, host)

	selection := input.selection
	selection.PlatformRoles = clonePlatformRoles(selection.PlatformRoles)
	selection.PlatformRoles[PlatformHost] = hostID
	input.selection = selection

	productID := nodeIDByKind(t, input.records.CaptureNodes, NodeCommandProduct)
	extra := Edge{
		Kind:       EdgeTargets,
		EdgeKey:    "edge:reviewer-undeclared-extra-host-target",
		FromNodeID: productID,
		ToNodeID:   hostID,
		Payload: TargetsPayload{
			BindingRole: PlatformHost,
			Origin:      EvidenceOrigin{Field: "selection.platform_roles.host.extra"},
		},
	}
	nodes := append(append([]Node{}, input.records.BindingNodes...), host)
	edges := append(append([]Edge{}, input.records.BindingEdges...), extra)
	input = rebuildBinding(t, input, nodes, edges)

	if _, err := projectInputs(t, input); err == nil {
		t.Fatal("ProjectActive accepted an extra host targets edge for a product that declares only the target role")
	}
}

func TestReviewerProbeRejectsDuplicateSemanticRequiresAcrossOrigins(t *testing.T) {
	input := cgp10Inputs(t)
	actionID := nodeIDByKind(t, input.records.CaptureNodes, NodeAction)
	toolchainID := nodeIDByKind(t, input.records.BindingNodes, NodeToolchainComponent)

	first := Edge{
		Kind:       EdgeRequires,
		EdgeKey:    "edge:reviewer-duplicate-toolchain-requires-one",
		FromNodeID: actionID,
		ToNodeID:   toolchainID,
		Payload: RequiresPayload{
			Scope:  ScopeToolchain,
			Origin: EvidenceOrigin{Field: "selection.toolchain_requires.one"},
		},
	}
	second := first
	second.EdgeKey = "edge:reviewer-duplicate-toolchain-requires-two"
	second.Payload = RequiresPayload{
		Scope:  ScopeToolchain,
		Origin: EvidenceOrigin{Field: "selection.toolchain_requires.two"},
	}
	edges := append(append([]Edge{}, input.records.BindingEdges...), first, second)
	input = rebuildBinding(t, input, input.records.BindingNodes, edges)

	if _, err := projectInputs(t, input); err == nil {
		t.Fatal("ProjectActive accepted the same requires relation twice solely because its evidence origins differed")
	}
}
