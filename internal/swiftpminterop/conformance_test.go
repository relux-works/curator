package swiftpminterop

import (
	"context"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

type mutableRecords struct {
	captureNodes []closuregraph.Node
	captureEdges []closuregraph.Edge
	bindingNodes []closuregraph.Node
	bindingEdges []closuregraph.Edge
}

func (records *mutableRecords) dropCaptureEdge(edgeKey string) {
	kept := records.captureEdges[:0]
	for _, edge := range records.captureEdges {
		if edge.EdgeKey != edgeKey {
			kept = append(kept, edge)
		}
	}
	records.captureEdges = append([]closuregraph.Edge{}, kept...)
}

func (records *mutableRecords) dropBindingEdge(edgeKey string) {
	kept := records.bindingEdges[:0]
	for _, edge := range records.bindingEdges {
		if edge.EdgeKey != edgeKey {
			kept = append(kept, edge)
		}
	}
	records.bindingEdges = append([]closuregraph.Edge{}, kept...)
}

// replaceCaptureNode rewrites one capture node and repoints every edge that
// named its previous identity so the mutation stays otherwise well formed.
func (records *mutableRecords) replaceCaptureNode(t *testing.T, logicalKey string, mutate func(closuregraph.Node) closuregraph.Node) {
	t.Helper()
	for index, node := range records.captureNodes {
		if node.LogicalKey != logicalKey {
			continue
		}
		previous, err := node.ID()
		if err != nil {
			t.Fatal(err)
		}
		replacement := mutate(node)
		next, err := replacement.ID()
		if err != nil {
			t.Fatal(err)
		}
		records.captureNodes[index] = replacement
		for edgeIndex := range records.captureEdges {
			if records.captureEdges[edgeIndex].FromNodeID == previous {
				records.captureEdges[edgeIndex].FromNodeID = next
			}
			if records.captureEdges[edgeIndex].ToNodeID == previous {
				records.captureEdges[edgeIndex].ToNodeID = next
			}
		}
		for edgeIndex := range records.bindingEdges {
			if records.bindingEdges[edgeIndex].FromNodeID == previous {
				records.bindingEdges[edgeIndex].FromNodeID = next
			}
			if records.bindingEdges[edgeIndex].ToNodeID == previous {
				records.bindingEdges[edgeIndex].ToNodeID = next
			}
		}
		return
	}
	t.Fatalf("capture node %s is absent", logicalKey)
}

func (records *mutableRecords) captureEdge(t *testing.T, edgeKey string) closuregraph.Edge {
	t.Helper()
	for _, edge := range records.captureEdges {
		if edge.EdgeKey == edgeKey {
			return edge
		}
	}
	t.Fatalf("capture edge %s is absent", edgeKey)
	return closuregraph.Edge{}
}

func (records *mutableRecords) replaceCaptureEdge(t *testing.T, edgeKey string, replacement closuregraph.Edge) {
	t.Helper()
	records.dropCaptureEdge(edgeKey)
	records.captureEdges = append(records.captureEdges, replacement)
}

// reproject rebuilds the exact capture, binding, and active records from a
// successful closure after an adversarial mutation and returns the shared
// projector's verdict.
func reproject(t *testing.T, fixture *fixture, result *Result, mutate func(*mutableRecords)) error {
	t.Helper()
	records := &mutableRecords{
		captureNodes: append([]closuregraph.Node(nil), result.Records.CaptureNodes...),
		captureEdges: append([]closuregraph.Edge(nil), result.Records.CaptureEdges...),
		bindingNodes: append([]closuregraph.Node(nil), result.Records.BindingNodes...),
		bindingEdges: append([]closuregraph.Edge(nil), result.Records.BindingEdges...),
	}
	mutate(records)
	graph, err := closuregraph.NewCaptureGraph(ProfileID, result.Graph.PolicyIDs, result.Graph.RootNodeIDs, records.captureNodes, records.captureEdges, result.Graph.ArtifactManifestIDs)
	if err != nil {
		return err
	}
	graphID, err := graph.ID()
	if err != nil {
		return err
	}
	selectionID, err := result.Selection.ID()
	if err != nil {
		return err
	}
	binding, err := closuregraph.NewSelectionBinding(graphID, selectionID, records.bindingNodes, records.bindingEdges)
	if err != nil {
		return err
	}
	tables := closuregraph.NewRecordTables(records.captureNodes, records.captureEdges, records.bindingNodes, records.bindingEdges)
	_, err = closuregraph.ProjectActive(graph, result.Selection, binding, tables, result.Authority, []closuregraph.ConditionEvaluator{fixtureEvaluator(fixture)})
	return err
}

func fixtureEvaluator(fixture *fixture) closuregraph.ConditionEvaluator {
	markers := fixture.source.Destination.Markers
	return closuregraph.ConditionEvaluatorFunc{EvaluatorID: swiftpmsource.ConditionEvaluatorID, EvaluateFunc: func(condition closuregraph.Condition, _ closuregraph.EvaluationInput) (bool, error) {
		parts := strings.Split(condition.Expression, "=")
		if len(parts) != 2 {
			return false, fail(CodeGraphReferenceInvalid, "unsupported condition")
		}
		return strings.EqualFold(markers[parts[0]], parts[1]), nil
	}}
}

// CGP05: one exact selection-neutral interop capture serves both accepted
// destinations. Platform nodes and concrete targets/uses_tool edges appear
// only in the binding overlays, which differ.
func TestCGP05InteropCaptureIsSelectionNeutralAcrossDestinations(t *testing.T) {
	darwin := newFixture(t).mustClose()
	linuxFixture := newFixture(t)
	linuxFixture.useLinuxDestination()
	linux := linuxFixture.mustClose()

	if darwin.GraphDigest != linux.GraphDigest {
		t.Fatalf("destination changed the selection-neutral interop capture: %s != %s", darwin.GraphDigest, linux.GraphDigest)
	}
	darwinBinding, _ := darwin.Binding.ID()
	linuxBinding, _ := linux.Binding.ID()
	if darwinBinding == linuxBinding {
		t.Fatal("destination did not change the exact selection binding")
	}
	darwinActive, _ := darwin.Active.ID()
	linuxActive, _ := linux.Active.ID()
	if darwinActive == linuxActive {
		t.Fatal("destination did not change the active projection")
	}
	if darwin.EvidenceDigest == linux.EvidenceDigest {
		t.Fatal("destination did not change the exact interop evidence record")
	}
	for _, result := range []*Result{darwin, linux} {
		for _, node := range result.Records.CaptureNodes {
			if node.Kind == closuregraph.NodeTargetPlatform || node.Kind == closuregraph.NodeToolchainComponent {
				t.Fatalf("selection-specific node leaked into interop capture: %s %s", node.Kind, node.LogicalKey)
			}
		}
		for _, edge := range result.Records.CaptureEdges {
			if edge.Kind == closuregraph.EdgeTargets || edge.Kind == closuregraph.EdgeUsesTool {
				t.Fatalf("selection-specific edge leaked into interop capture: %s %s", edge.Kind, edge.EdgeKey)
			}
		}
		if !hasBindingEdge(result, closuregraph.EdgeUsesTool) {
			t.Fatal("selection binding declares no uses_tool edge")
		}
		if !hasBindingEdge(result, closuregraph.EdgeTargets) {
			t.Fatal("selection binding declares no targets edge")
		}
	}
}

// CGP05: a conditional target edge must not move a capture record. The same
// admitted closure closed against the accepted Darwin and Linux destinations
// publishes one byte-identical interop capture even though the edge is selected
// on one destination and pruned on the other; only the binding, the active
// projection, and the exact evidence record differ.
func TestCGP05ConditionalEdgeKeepsInteropCaptureSelectionNeutral(t *testing.T) {
	build := func(linux bool) *Result {
		fixture := newFixture(t)
		fixture.target("App").Dependencies = []swiftpmsource.TargetDependency{{Name: "CLib", Condition: swiftpmCondition("platform=macos")}}
		if linux {
			fixture.useLinuxDestination()
		}
		return fixture.mustClose()
	}
	darwin, linux := build(false), build(true)
	if darwin.GraphDigest != linux.GraphDigest {
		t.Fatalf("a conditional edge changed the selection-neutral interop capture: %s != %s", darwin.GraphDigest, linux.GraphDigest)
	}
	if len(darwin.Targets) != len(linux.Targets) || len(darwin.Boundaries) != len(linux.Boundaries) {
		t.Fatalf("capture target/boundary counts diverged: %d/%d vs %d/%d", len(darwin.Targets), len(darwin.Boundaries), len(linux.Targets), len(linux.Boundaries))
	}
	if !darwin.Boundaries[0].Selected || linux.Boundaries[0].Selected {
		t.Fatalf("selection verdicts = darwin %v, linux %v", darwin.Boundaries[0].Selected, linux.Boundaries[0].Selected)
	}
	if state := activationForID(t, darwin, darwin.Boundaries[0].NodeID); state != closuregraph.ActivationSelected {
		t.Fatalf("darwin boundary activation = %q", state)
	}
	if state := activationForID(t, linux, linux.Boundaries[0].NodeID); state != closuregraph.ActivationPruned {
		t.Fatalf("linux boundary activation = %q", state)
	}
	darwinActive, _ := darwin.Active.ID()
	linuxActive, _ := linux.Active.ID()
	if darwinActive == linuxActive {
		t.Fatal("destination did not change the active projection")
	}
	if darwin.EvidenceDigest == linux.EvidenceDigest {
		t.Fatal("destination did not change the exact interop evidence record")
	}
}

// CGP05: the exact binding must name the platform, the Swift and Clang
// toolchains, the SDK, and every selected system or interop component.
func TestCGP05BindingNamesEveryExactSelectedIdentity(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	roles := map[string]bool{}
	platforms := 0
	for _, node := range result.Records.BindingNodes {
		switch payload := node.Payload.(type) {
		case closuregraph.ToolchainComponentPayload:
			roles[payload.ComponentRole] = true
		case closuregraph.TargetPlatformPayload:
			platforms++
			if payload.TargetTriple != "arm64-apple-macosx14.0" {
				t.Fatalf("bound platform = %#v", payload)
			}
		}
	}
	for _, role := range []string{"swift", "clang", "clang++", "macos-sdk"} {
		if !roles[role] {
			t.Fatalf("selection binding omits the exact %s identity: %v", role, roles)
		}
	}
	if platforms != 1 {
		t.Fatalf("selection binding names %d platforms", platforms)
	}
	compileTool := false
	for _, edge := range result.Records.BindingEdges {
		if payload, ok := edge.Payload.(closuregraph.UsesToolPayload); ok && payload.ToolSlot == "compiler" && payload.InvocationRole == "clang-compile" {
			compileTool = true
		}
	}
	if !compileTool {
		t.Fatal("selection binding omits the exact Clang uses_tool edge")
	}
}

// CGN03: an ABI-incompatible consumer side is rejected by the shared interop
// contract before any compiler starts.
func TestCGN03IncompatibleInteropContractIsRejected(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	key := "swiftpm.interop.consumes.root:CLib->root:App"
	err := reproject(t, fixture, result, func(records *mutableRecords) {
		edge := records.captureEdge(t, key)
		payload := edge.Payload.(closuregraph.ConsumesInteropPayload)
		payload.ABIExpectation = "cxx-abi-v9"
		edge.Payload = payload
		records.replaceCaptureEdge(t, key, edge)
	})
	requireCode(t, err, CodeInteropUndeclared)
}

// CGN03: a boundary that loses its consumer side has no declared interop edge.
func TestCGN03MissingConsumerSideIsRejected(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	err := reproject(t, fixture, result, func(records *mutableRecords) {
		records.dropCaptureEdge("swiftpm.interop.consumes.root:CLib->root:App")
	})
	requireCode(t, err, CodeInteropUndeclared)
}

// CGN03: a boundary that loses its exact toolchain binding is rejected.
func TestCGN03MissingBoundaryToolchainBindingIsRejected(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	err := reproject(t, fixture, result, func(records *mutableRecords) {
		records.dropBindingEdge("swiftpm.interop.boundary-toolchain.root:CLib->root:App")
	})
	requireCode(t, err, CodeInteropUndeclared)
}

// CGN09: an omitted or duplicated targets edge is invalid binding structure,
// and time-of-use toolchain drift is rejected before any admitted byte is read.
func TestCGN09TargetBindingStructureAndDriftAreRejected(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	t.Run("omitted targets edge", func(t *testing.T) {
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			records.dropBindingEdge("swiftpm.interop.compile-target.root.CLib")
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("duplicated targets edge", func(t *testing.T) {
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			for _, edge := range records.bindingEdges {
				if edge.EdgeKey == "swiftpm.interop.compile-target.root.CLib" {
					duplicate := edge
					duplicate.EdgeKey = edge.EdgeKey + ".duplicate"
					records.bindingEdges = append(records.bindingEdges, duplicate)
					break
				}
			}
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("toolchain drift", func(t *testing.T) {
		drift := newFixture(t)
		drift.materializeHook = func() {
			drift.interop.Recheck = func(_ context.Context, _ swiftpmsource.ToolIdentity) (closureexec.ToolchainIdentity, error) {
				return closureexec.ToolchainIdentity{Fingerprint: id('a'), ExecutableSHA256: id('b')}, nil
			}
		}
		_, err := drift.close()
		requireCode(t, err, CodeToolchainChanged)
	})
}

// CGN15: duplicate keys, dangling endpoints, wrong-kind endpoints, capture
// replacement, and multiply bound action slots all reject canonically.
func TestCGN15GraphReferenceViolationsRejectCanonically(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	t.Run("duplicate logical key", func(t *testing.T) {
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			for _, node := range records.captureNodes {
				if node.LogicalKey == "swiftpm.interop.compile.root.CLib" {
					duplicate := node
					payload := duplicate.Payload.(closuregraph.ActionPayload)
					payload.ActionSubtype = "clang-compile-duplicate"
					duplicate.Payload = payload
					records.captureNodes = append(records.captureNodes, duplicate)
					break
				}
			}
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("dangling endpoint", func(t *testing.T) {
		key := "swiftpm.interop.read-headers.root.CLib"
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			edge := records.captureEdge(t, key)
			edge.ToNodeID = id('c')
			records.replaceCaptureEdge(t, key, edge)
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("wrong-kind endpoint", func(t *testing.T) {
		key := "swiftpm.interop.produce-object.root.CLib.0000"
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			edge := records.captureEdge(t, key)
			edge.ToNodeID = captureNodeID(t, records, "swiftpm.interop.headers.root.CLib")
			records.replaceCaptureEdge(t, key, edge)
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("action slot bound twice", func(t *testing.T) {
		err := reproject(t, fixture, result, func(records *mutableRecords) {
			duplicate := records.captureEdge(t, "swiftpm.interop.read-headers.root.CLib")
			payload := duplicate.Payload.(closuregraph.ReadsPayload)
			payload.ReadSlot = "sources"
			duplicate.Payload = payload
			duplicate.EdgeKey = "swiftpm.interop.read-headers.root.CLib.duplicate-slot"
			records.captureEdges = append(records.captureEdges, duplicate)
		})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("capture replacement", func(t *testing.T) {
		selectionID, _ := result.Selection.ID()
		binding, err := closuregraph.NewSelectionBinding(id('d'), selectionID, result.Records.BindingNodes, result.Records.BindingEdges)
		if err != nil {
			t.Fatal(err)
		}
		_, err = closuregraph.ProjectActive(result.Graph, result.Selection, binding, result.Records, result.Authority, []closuregraph.ConditionEvaluator{fixtureEvaluator(fixture)})
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
}

// The C4 interop checkpoint chains from the accepted source-closure C4 and
// binds the four exact graph identities.
func TestInteropCheckpointChainsFromAcceptedSourceClosure(t *testing.T) {
	fixture := newFixture(t)
	capture := func() *swiftpmsource.Capture {
		fixture.materialize()
		value, err := fixture.capture()
		if err != nil {
			t.Fatal(err)
		}
		return value
	}()
	result, err := Close(t.Context(), fixture.interop, capture)
	if err != nil {
		t.Fatal(err)
	}
	previous, err := capture.C4.ID()
	if err != nil {
		t.Fatal(err)
	}
	if result.C4.PreviousCheckpointID == nil || *result.C4.PreviousCheckpointID != previous {
		t.Fatalf("interop checkpoint predecessor = %v, want %s", result.C4.PreviousCheckpointID, previous)
	}
	payload, ok := result.C4.Payload.(closuregraph.C4ClosePayload)
	activeID, _ := result.Active.ID()
	bindingID, _ := result.Binding.ID()
	selectionID, _ := result.Selection.ID()
	if !ok || payload.ActiveGraphID != activeID || payload.CapturedGraphID != result.GraphDigest || payload.SelectionBindingID != bindingID || payload.SelectionContextID != selectionID {
		t.Fatalf("interop checkpoint payload = %#v", result.C4.Payload)
	}
}

// activationForKey returns the active projection's verdict for one capture
// node named by its logical key.
func activationForKey(t *testing.T, result *Result, logicalKey string) closuregraph.ActivationState {
	t.Helper()
	for _, node := range result.Records.CaptureNodes {
		if node.LogicalKey != logicalKey {
			continue
		}
		id, err := node.ID()
		if err != nil {
			t.Fatal(err)
		}
		return activationForID(t, result, id)
	}
	t.Fatalf("capture node %s is absent", logicalKey)
	return ""
}

// activationForID returns the active projection's verdict for one capture node.
func activationForID(t *testing.T, result *Result, nodeID closuregraph.ID) closuregraph.ActivationState {
	t.Helper()
	for _, activation := range result.Active.NodeActivations {
		if activation.NodeID == nodeID {
			return activation.State
		}
	}
	t.Fatalf("active projection has no activation for %s", nodeID)
	return ""
}

// edgeActivation returns the recorded conditional-edge verdict for one capture
// edge key.
func edgeActivation(t *testing.T, result *Result, edgeKey string) closuregraph.EdgeActivation {
	t.Helper()
	for _, edge := range result.Records.CaptureEdges {
		if edge.EdgeKey != edgeKey {
			continue
		}
		id, err := edge.ID()
		if err != nil {
			t.Fatal(err)
		}
		for _, activation := range result.Active.EdgeActivations {
			if activation.EdgeID == id {
				return activation
			}
		}
		t.Fatalf("capture edge %s has no recorded activation", edgeKey)
	}
	t.Fatalf("capture edge %s is absent", edgeKey)
	return closuregraph.EdgeActivation{}
}

func hasBindingEdge(result *Result, kind closuregraph.EdgeKind) bool {
	for _, edge := range result.Records.BindingEdges {
		if edge.Kind == kind {
			return true
		}
	}
	return false
}

func captureNodeID(t *testing.T, records *mutableRecords, logicalKey string) closuregraph.ID {
	t.Helper()
	for _, node := range records.captureNodes {
		if node.LogicalKey != logicalKey {
			continue
		}
		id, err := node.ID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	t.Fatalf("capture node %s is absent", logicalKey)
	return ""
}

// newMultiPackageFixture builds a root package with an in-root
// `.package(path:)` dependency, a C target in each package, and a cross-package
// target edge. Every earlier fixture is single-package, so the cross-package
// hop the include worklist performs — `confineInclude` returns a unit named by
// the *resolved* package and the worklist opens it under that package's root
// while the consumer's search path stays fixed — was read-verified but never
// executed.
func newMultiPackageFixture(t *testing.T) *fixture {
	t.Helper()
	fixture := newFixture(t)
	delete(fixture.files, "Sources/CLib/lib.c")
	delete(fixture.files, "Sources/CLib/include/CLib.h")
	fixture.addFiles(map[string]string{
		"Sources/App/main.swift":                               "import RootC\nprint(root_value())\n",
		"Sources/RootC/lib.c":                                  "#include \"RootC.h\"\n#include \"VendorLib.h\"\nint root_value(void) { return vendor_value(); }\n",
		"Sources/RootC/include/RootC.h":                        "#ifndef ROOTC_H\n#define ROOTC_H\nint root_value(void);\n#endif\n",
		"Vendor/VendorLib/Package.swift":                       "// swift-tools-version: 6.1\nimport PackageDescription\n",
		"Vendor/VendorLib/Sources/CLib/lib.c":                  "#include \"VendorLib.h\"\nint vendor_value(void) { return 7; }\n",
		"Vendor/VendorLib/Sources/CLib/include/VendorLib.h":    "#ifndef VENDORLIB_H\n#define VENDORLIB_H\n#include \"VendorDetail.h\"\nint vendor_value(void);\n#endif\n",
		"Vendor/VendorLib/Sources/CLib/include/VendorDetail.h": "#ifndef VENDORDETAIL_H\n#define VENDORDETAIL_H\ntypedef int vendor_int;\n#endif\n",
	})
	fixture.manifest = swiftpmsource.Manifest{
		PackageName: "root", ToolsVersion: "6.1",
		Dependencies: []swiftpmsource.ManifestDependency{{Identity: "vendorlib", Kind: swiftpmsource.SourcePath, LocalPath: "Vendor/VendorLib", Requirement: "path"}},
		Products:     []swiftpmsource.Product{{Name: "cli", Type: "executable", Targets: []string{"App"}}},
		Targets: []swiftpmsource.Target{
			{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "RootC"}}},
			{Name: "RootC", Type: "regular", Path: "Sources/RootC", Sources: []string{"Sources/RootC/lib.c"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "CLib", Package: "vendorlib", Product: "VendorProduct"}}},
		},
	}
	fixture.manifests = map[string]swiftpmsource.Manifest{"vendorlib": {
		PackageName: "vendorlib", ToolsVersion: "6.1",
		Products: []swiftpmsource.Product{{Name: "VendorProduct", Type: "library", Targets: []string{"CLib"}}},
		Targets:  []swiftpmsource.Target{{Name: "CLib", Type: "regular", Path: "Sources/CLib", Sources: []string{"Sources/CLib/lib.c"}}},
	}}
	return fixture
}

// The include worklist crosses a package boundary: a header reached from
// another package is opened under *that* package's root, its own directives are
// scanned, and the reference records which package the path is relative to.
func TestCrossPackageIncludeClosureIsScanned(t *testing.T) {
	result := newMultiPackageFixture(t).mustClose()
	root, vendor := mustTarget(t, result, "root:RootC"), mustTarget(t, result, "vendorlib:CLib")
	if root.Package != "root" || vendor.Package != "vendorlib" {
		t.Fatalf("target packages = %q/%q", root.Package, vendor.Package)
	}
	// The transitive reference lives in the vendor package's header, so it is
	// recorded only if the worklist opened that header under the vendor root.
	// Its Source is relative to that root, which is why SourcePackage has to
	// name the vendor package rather than the consuming target's package.
	found := false
	for _, reference := range root.Includes {
		if reference.Spelling != "VendorDetail.h" {
			continue
		}
		found = true
		if reference.Package != "root" || reference.SourcePackage != "vendorlib" || reference.Source != "Sources/CLib/include/VendorLib.h" {
			t.Fatalf("cross-package reference = %#v", reference)
		}
	}
	if !found {
		t.Fatalf("cross-package include closure = %#v", root.Includes)
	}
	for _, reference := range root.Includes {
		if reference.Spelling == "VendorLib.h" && (reference.SourcePackage != "root" || reference.Source != "Sources/RootC/lib.c") {
			t.Fatalf("consumer include record = %#v", reference)
		}
	}
	if len(result.Boundaries) != 2 {
		t.Fatalf("boundaries = %#v", result.Boundaries)
	}
	for _, boundary := range result.Boundaries {
		if boundary.Mode != closuregraph.InteropCABI {
			t.Fatalf("boundary = %#v", boundary)
		}
	}
}

// A directive inside a cross-package header is really executed by the scanner:
// the same graph fails closed when the vendor package's own header escapes.
func TestCrossPackageIncludeEscapeFailsClosed(t *testing.T) {
	fixture := newMultiPackageFixture(t)
	fixture.files["Vendor/VendorLib/Sources/CLib/include/VendorDetail.h"] = "#ifndef VENDORDETAIL_H\n#define VENDORDETAIL_H\n#include </etc/passwd>\n#endif\n"
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)
}

// The include search path is fixed per translation unit, not per scanned file:
// a vendor header that only resolves against the *consumer's* search roots
// fails closed, because the vendor package's own target compiles that same
// header with only its own roots. Nothing in the worklist may widen a
// dependency's search path to its consumer's.
func TestCrossPackageIncludeKeepsThePerTargetSearchPath(t *testing.T) {
	fixture := newMultiPackageFixture(t)
	fixture.files["Vendor/VendorLib/Sources/CLib/include/VendorDetail.h"] = "#ifndef VENDORDETAIL_H\n#define VENDORDETAIL_H\n#include \"RootC.h\"\n#endif\n"
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)
}
