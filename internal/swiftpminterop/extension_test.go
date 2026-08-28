package swiftpminterop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// P01: dormant plugin and macro declarations that the selected product never
// reaches are captured as source and never compiled or invoked.
func TestP01DormantExtensionsAreCapturedAndNeverReached(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Plugins/Dormant/plugin.swift":    "// dormant build-tool plugin\n",
		"Sources/DormantMacro/impl.swift": "// dormant macro implementation\n",
	})
	fixture.addTarget(swiftpmsource.Target{Name: "Dormant", Type: "plugin", Path: "Plugins/Dormant", Sources: []string{"Plugins/Dormant/plugin.swift"}})
	fixture.addTarget(swiftpmsource.Target{Name: "DormantMacro", Type: "macro", Path: "Sources/DormantMacro", Sources: []string{"Sources/DormantMacro/impl.swift"}})
	result := fixture.mustClose()
	for _, target := range result.Targets {
		if target.Target == "Dormant" || target.Target == "DormantMacro" {
			t.Fatalf("dormant extension entered the interop closure: %#v", target)
		}
	}
	for _, node := range result.Records.CaptureNodes {
		if node.LogicalKey == "swiftpm.interop.compile.root.Dormant" || node.LogicalKey == "swiftpm.interop.compile.root.DormantMacro" {
			t.Fatalf("dormant extension gained a compile action: %s", node.LogicalKey)
		}
	}
	if !captureContainsKey(result, "swiftpm.target.root.Dormant") {
		t.Fatal("dormant plugin source declaration was not retained in capture")
	}
}

// P02/P03: a selected graph that reaches a build-tool or command plugin is
// rejected before the plugin, its tool dependencies, or any command runs.
func TestP02P03ReachedPluginIsRejectedBeforeExecution(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{"Plugins/Active/plugin.swift": "// active build-tool plugin\n"})
	fixture.addTarget(swiftpmsource.Target{Name: "Active", Type: "plugin", Path: "Plugins/Active", Sources: []string{"Plugins/Active/plugin.swift"}})
	fixture.target("App").Dependencies = append(fixture.target("App").Dependencies, swiftpmsource.TargetDependency{Name: "Active"})
	_, err := fixture.close()
	requireCode(t, err, CodePluginUnsupported)

	_, _, directErr := classifyTarget("root", "Active", swiftpmsource.Target{Name: "Active", Type: "plugin"})
	requireCode(t, directErr, CodePluginUnsupported)
}

// P04: a selected graph that reaches a macro implementation is rejected before
// macro compilation or any prebuilt retrieval.
func TestP04ReachedMacroIsRejectedBeforeExecution(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{"Sources/ActiveMacro/impl.swift": "// active macro\n"})
	fixture.addTarget(swiftpmsource.Target{Name: "ActiveMacro", Type: "macro", Path: "Sources/ActiveMacro", Sources: []string{"Sources/ActiveMacro/impl.swift"}})
	fixture.target("App").Dependencies = append(fixture.target("App").Dependencies, swiftpmsource.TargetDependency{Name: "ActiveMacro"})
	_, err := fixture.close()
	requireCode(t, err, CodeMacroUnsupported)

	_, _, directErr := classifyTarget("root", "ActiveMacro", swiftpmsource.Target{Name: "ActiveMacro", Type: "macro"})
	requireCode(t, directErr, CodeMacroUnsupported)
}

// P05: the manifest permit that feeds this stage must always disable SwiftPM's
// experimental prebuilt path. The fixture evaluator fails the permit
// otherwise, so the closure cannot reach interop validation.
func TestP05ManifestPermitAlwaysDisablesExperimentalPrebuilts(t *testing.T) {
	fixture := newFixture(t)
	if _, err := fixture.close(); err != nil {
		t.Fatalf("accepted permit was rejected: %v", err)
	}
}

// P06/P08: every binary target is unavailable, dormant or referenced, and the
// binary rejection dominates the plugin handling of a plugin tool.
func TestP06P08BinaryTargetsAreUnavailable(t *testing.T) {
	t.Run("dormant", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addTarget(swiftpmsource.Target{Name: "Prebuilt", Type: "binary"})
		_, err := fixture.close()
		requireCode(t, err, CodeBinaryUnavailable)
	})
	t.Run("referenced", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addTarget(swiftpmsource.Target{Name: "Prebuilt", Type: "binary"})
		fixture.target("App").Dependencies = append(fixture.target("App").Dependencies, swiftpmsource.TargetDependency{Name: "Prebuilt"})
		_, err := fixture.close()
		requireCode(t, err, CodeBinaryUnavailable)
	})
	t.Run("plugin tool", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{"Plugins/Active/plugin.swift": "// plugin using a binary tool\n"})
		fixture.addTarget(swiftpmsource.Target{Name: "Prebuilt", Type: "binary"})
		fixture.addTarget(swiftpmsource.Target{Name: "Active", Type: "plugin", Path: "Plugins/Active", Sources: []string{"Plugins/Active/plugin.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "Prebuilt"}}})
		fixture.target("App").Dependencies = append(fixture.target("App").Dependencies, swiftpmsource.TargetDependency{Name: "Active"})
		_, err := fixture.close()
		requireCode(t, err, CodeBinaryUnavailable)
	})
	_, _, directErr := classifyTarget("root", "Prebuilt", swiftpmsource.Target{Name: "Prebuilt", Type: "binary"})
	requireCode(t, directErr, CodeBinaryUnavailable)
}

// P07: a compiled payload renamed to look like a header is rejected by the
// shared recursive detector before this stage inspects anything.
func TestP07RenamedCompiledPayloadIsRejectedBeforeInterop(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() {
		full := filepath.Join(fixture.root, "Sources", "CLib", "include", "vendor.h")
		if err := os.WriteFile(full, []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	_, err := fixture.close()
	if artifactpolicy.ErrorCode(err) != artifactpolicy.CodeCompiledDependency {
		t.Fatalf("error = %v", err)
	}
}

// P09: a generated header that a compile action reads but no declared action
// produces has no accepted lineage.
func TestP09GeneratedInputWithoutLineageIsRejected(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	generated := closuregraph.Node{Kind: closuregraph.NodeGeneratedArtifact, LogicalKey: "swiftpm.interop.generated.root.CLib.header", Payload: closuregraph.GeneratedArtifactPayload{
		Profile: ProfileID, LogicalPath: ".curator/generated/CLib/Generated.h", Slot: "generated", ExpectedClass: "source.header",
		Grammar: "c-family-header-v1", Role: "intermediate", DeclarationDigest: id('e'),
	}}
	generatedID, nodeErr := generated.ID()
	if nodeErr != nil {
		t.Fatal(nodeErr)
	}
	err := reproject(t, fixture, result, func(records *mutableRecords) {
		records.captureNodes = append(records.captureNodes, generated)
		records.replaceCaptureNode(t, "swiftpm.interop.compile.root.CLib", func(node closuregraph.Node) closuregraph.Node {
			payload := node.Payload.(closuregraph.ActionPayload)
			payload.ReadSlotNames = append([]string{"generated"}, payload.ReadSlotNames...)
			payload.ArgvTemplate = append(payload.ArgvTemplate, "$READ(generated)")
			node.Payload = payload
			return node
		})
		records.captureEdges = append(records.captureEdges, closuregraph.Edge{
			Kind: closuregraph.EdgeReads, EdgeKey: "swiftpm.interop.read-generated.root.CLib",
			FromNodeID: captureNodeID(t, records, "swiftpm.interop.compile.root.CLib"), ToNodeID: generatedID,
			Payload: closuregraph.ReadsPayload{Path: ".curator/generated/CLib/Generated.h", ReadSlot: "generated", ReadClass: "source.header"},
		})
	})
	requireCode(t, err, CodeGeneratedInputUndeclared)
}

func captureContainsKey(result *Result, logicalKey string) bool {
	for _, node := range result.Records.CaptureNodes {
		if node.LogicalKey == logicalKey {
			return true
		}
	}
	return false
}
