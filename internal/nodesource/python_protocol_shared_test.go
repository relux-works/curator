package nodesource

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

const (
	pythonTargetOutcomeLabel = "python-protocol-target-outcome-v1"
	pythonReuseNegativeLabel = "python-protocol-cross-target-reuse-v1"
)

func derivePythonP10SharedOutcomes(t *testing.T, testCase pythonProtocolCase) ([]pythonCanonicalRecord, pythonCanonicalRecord, pythonProtocolExpected) {
	t.Helper()
	fixture := testCase.Input
	if testCase.ID != "P10" || len(fixture.Packages) != 1 || len(fixture.Targets) != 2 || fixture.ReuseAttempt == nil {
		t.Fatal("P10 must declare one selection-neutral package, two target bindings, and a reuse attempt")
	}
	artifactID := id("python-p10-artifact")
	declarationID := id("python-p10-declaration")
	product := closuregraph.Node{Kind: closuregraph.NodeCommandProduct, LogicalKey: "python:portable", Payload: closuregraph.CommandProductPayload{
		Profile: "python-source-v1", SkillKey: "python-protocol", CommandKey: "portable", EntryPointContract: "python-module", DeclarationDigest: declarationID,
	}}
	packageNode := closuregraph.Node{Kind: closuregraph.NodePackageInstance, LogicalKey: "python:portable@1", Payload: closuregraph.PackageInstancePayload{
		Profile: "python-source-v1", Ecosystem: "python", Origin: "pylock://portable/1", LockInstanceKey: "portable@1", Name: "portable", Version: "1", ArtifactManifestID: artifactID, TrustRole: closuregraph.TrustDependencyInput,
	}}
	productID := mustSharedNodeID(t, product)
	packageID := mustSharedNodeID(t, packageNode)
	requires := closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: "python:portable:requires", FromNodeID: productID, ToNodeID: packageID, Payload: closuregraph.RequiresPayload{
		Scope: closuregraph.ScopeRuntime, Origin: closuregraph.EvidenceOrigin{Field: "packages[portable]"}, DependencyKind: "runtime",
	}}
	mustSharedEdgeID(t, requires)
	capture, err := closuregraph.NewCaptureGraph("python-source-v1", []string{"curator-artifact-policy-v1"}, []closuregraph.ID{productID}, []closuregraph.Node{product, packageNode}, []closuregraph.Edge{requires}, []closuregraph.ID{artifactID})
	if err != nil {
		t.Fatal(err)
	}
	captureRecord := mustSharedRecord(t, closuregraph.LabelCaptureGraph, capture)
	captureID := closuregraph.ID(captureRecord.ID)
	graphRecords := []map[string]any{mustWireRecord(t, closuregraph.LabelNode, product), mustWireRecord(t, closuregraph.LabelNode, packageNode), mustWireRecord(t, closuregraph.LabelEdge, requires)}
	sort.Slice(graphRecords, func(i, j int) bool { return graphRecords[i]["id"].(string) < graphRecords[j]["id"].(string) })

	targetOutcomes := make([]pythonCanonicalRecord, 0, 2)
	bindingIDs := make([]closuregraph.ID, 0, 2)
	activeIDs := make([]closuregraph.ID, 0, 2)
	bindingsByKey := map[string]closuregraph.ID{}
	for _, target := range fixture.Targets {
		targetKey := target.Interpreter + "/" + target.Platform + "/" + target.ABI
		architecture, libc, triple := "x86_64", "glibc", "x86_64-unknown-linux-gnu"
		if target.Platform == "darwin" {
			architecture, libc, triple = "arm64", "libSystem", "arm64-apple-darwin"
		}
		platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "python-target:" + targetKey, Payload: closuregraph.TargetPlatformPayload{
			OS: target.Platform, Architecture: architecture, ABI: target.ABI, Libc: libc, MinimumRuntime: "profile-v1", SDKID: "python-runtime-v1", TargetTriple: triple, Runtime: target.Interpreter,
			LanguageModes: map[string]string{"python": target.Interpreter}, Tuning: map[string]string{},
		}}
		platformID := mustSharedNodeID(t, platform)
		selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{productID}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, true,
			map[string]string{"python_abi": target.ABI, "python_interpreter": target.Interpreter, "sys_platform": target.Platform}, map[string]string{}, []string{"python-marker-v1"})
		if err != nil {
			t.Fatal(err)
		}
		selectionRecord := mustSharedRecord(t, closuregraph.LabelSelectionContext, selection)
		selectionID := closuregraph.ID(selectionRecord.ID)
		targetEdge := closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: "python:portable:target", FromNodeID: productID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}}
		mustSharedEdgeID(t, targetEdge)
		binding, err := closuregraph.NewSelectionBinding(captureID, selectionID, []closuregraph.Node{platform}, []closuregraph.Edge{targetEdge})
		if err != nil {
			t.Fatal(err)
		}
		bindingRecord := mustSharedRecord(t, closuregraph.LabelSelectionBinding, binding)
		bindingID := closuregraph.ID(bindingRecord.ID)
		active := closuregraph.ActiveGraph{SchemaID: closuregraph.SchemaActiveGraph, CapturedGraphID: captureID, SelectionContextID: selectionID, SelectionBindingID: bindingID,
			NodeActivations: []closuregraph.NodeActivation{{NodeID: productID, State: closuregraph.ActivationSelected}, {NodeID: packageID, State: closuregraph.ActivationSelected}}, EdgeActivations: []closuregraph.EdgeActivation{}, NonOrderingSCCs: []closuregraph.NonOrderingSCC{}}
		sort.Slice(active.NodeActivations, func(i, j int) bool { return active.NodeActivations[i].NodeID < active.NodeActivations[j].NodeID })
		activeRecord := mustSharedRecord(t, closuregraph.LabelActiveGraph, active)
		activeID := closuregraph.ID(activeRecord.ID)
		targetGraphRecords := append([]map[string]any{}, graphRecords...)
		targetGraphRecords = append(targetGraphRecords, mustWireRecord(t, closuregraph.LabelNode, platform), mustWireRecord(t, closuregraph.LabelEdge, targetEdge))
		sort.Slice(targetGraphRecords, func(i, j int) bool {
			return targetGraphRecords[i]["id"].(string) < targetGraphRecords[j]["id"].(string)
		})
		payload := map[string]any{
			"active_graph": recordEnvelope(activeRecord), "case_id": "P10", "capture_graph": recordEnvelope(captureRecord), "decision": "admit", "diagnostic": nil,
			"graph_records": recordsAsAny(targetGraphRecords), "schema_id": "python-protocol-target-outcome-v1", "selection_binding": recordEnvelope(bindingRecord),
			"selection_context": recordEnvelope(selectionRecord), "target_key": targetKey,
		}
		targetOutcomes = append(targetOutcomes, mustCanonicalProtocolRecord(t, pythonTargetOutcomeLabel, payload))
		bindingsByKey[targetKey] = bindingID
		bindingIDs = append(bindingIDs, bindingID)
		activeIDs = append(activeIDs, activeID)
	}
	from := bindingsByKey[fixture.ReuseAttempt.FromTarget]
	to := bindingsByKey[fixture.ReuseAttempt.ToTarget]
	if !from.Valid() || !to.Valid() || from == to {
		t.Fatal("P10 reuse attempt must reference the two exact distinct bindings")
	}
	diagnostic := map[string]any{"code": string(closuregraph.CodeTargetIdentityChanged), "fields": map[string]any{"from_binding_id": string(from), "to_binding_id": string(to)}, "subject": "python-target-binding"}
	reuse := mustCanonicalProtocolRecord(t, pythonReuseNegativeLabel, map[string]any{
		"case_id": "P10", "diagnostic": diagnostic, "from_selection_binding_id": string(from), "schema_id": "python-protocol-cross-target-reuse-v1", "to_selection_binding_id": string(to),
	})
	sort.Slice(bindingIDs, func(i, j int) bool { return bindingIDs[i] < bindingIDs[j] })
	sort.Slice(activeIDs, func(i, j int) bool { return activeIDs[i] < activeIDs[j] })
	summary := pythonProtocolExpected{Decision: "admit", CaptureGraphID: captureID, BindingIDs: bindingIDs, ActiveGraphIDs: activeIDs, ReuseDiagnosticCode: string(closuregraph.CodeTargetIdentityChanged)}
	return targetOutcomes, reuse, summary
}

func TestPythonP10SharedWireRejectsMissingAndUnknownFields(t *testing.T) {
	corpusBytes, err := os.ReadFile("testdata/python_protocol_shared_records.json")
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := decodePythonProtocolCorpus(corpusBytes)
	if err != nil {
		t.Fatal(err)
	}
	var p10 pythonProtocolCase
	for _, testCase := range corpus.Cases {
		if testCase.ID == "P10" {
			p10 = testCase
		}
	}
	outcomes, reuse, _ := derivePythonP10SharedOutcomes(t, p10)
	var outcome map[string]any
	if err := json.Unmarshal([]byte(outcomes[0].CCJ), &outcome); err != nil {
		t.Fatal(err)
	}
	decoders := map[string]func([]byte) error{
		"capture_graph":     func(data []byte) error { _, err := closuregraph.DecodeCaptureGraph(data); return err },
		"selection_context": func(data []byte) error { _, err := closuregraph.DecodeSelectionContext(data); return err },
		"selection_binding": func(data []byte) error { _, err := closuregraph.DecodeSelectionBinding(data); return err },
		"active_graph":      func(data []byte) error { _, err := closuregraph.DecodeActiveGraph(data); return err },
	}
	for name, decode := range decoders {
		envelope := outcome[name].(map[string]any)
		original := envelope["payload"].(map[string]any)
		for _, mutation := range []string{"missing", "unknown"} {
			t.Run(name+"_"+mutation, func(t *testing.T) {
				payload := cloneProtocolObject(t, original)
				if mutation == "missing" {
					delete(payload, "schema_id")
				} else {
					payload["unknown_field"] = true
				}
				data, err := json.Marshal(payload)
				if err != nil {
					t.Fatal(err)
				}
				if err := decode(data); err == nil {
					t.Fatalf("%s accepted %s field set", name, mutation)
				}
			})
		}
	}
	var reusePayload map[string]any
	if err := json.Unmarshal([]byte(reuse.CCJ), &reusePayload); err != nil {
		t.Fatal(err)
	}
	diagnostic := reusePayload["diagnostic"].(map[string]any)
	if err := validatePythonProtocolDiagnostic(diagnostic); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing", "unknown"} {
		mutated := cloneProtocolObject(t, diagnostic)
		if mutation == "missing" {
			delete(mutated, "subject")
		} else {
			mutated["unknown_field"] = true
		}
		if err := validatePythonProtocolDiagnostic(mutated); err == nil {
			t.Fatalf("diagnostic accepted %s field set", mutation)
		}
	}
}

func cloneProtocolObject(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func validatePythonProtocolDiagnostic(value map[string]any) error {
	if err := requireJSONKeys(value, "diagnostic", "code", "fields", "subject"); err != nil {
		return err
	}
	fields, ok := value["fields"].(map[string]any)
	if !ok {
		return fmt.Errorf("diagnostic fields must be an object")
	}
	if err := requireJSONKeys(fields, "diagnostic.fields", "from_binding_id", "to_binding_id"); err != nil {
		return err
	}
	if value["code"] != string(closuregraph.CodeTargetIdentityChanged) || value["subject"] != "python-target-binding" {
		return fmt.Errorf("diagnostic semantic mismatch")
	}
	for _, key := range []string{"from_binding_id", "to_binding_id"} {
		text, ok := fields[key].(string)
		if !ok || !closuregraph.ID(text).Valid() {
			return fmt.Errorf("diagnostic %s is invalid", key)
		}
	}
	return nil
}

type sharedCanonicalRecord interface {
	CanonicalBytes() ([]byte, error)
	ID() (closuregraph.ID, error)
}

func mustSharedRecord(t *testing.T, label string, value sharedCanonicalRecord) pythonCanonicalRecord {
	t.Helper()
	ccj, err := value.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	recordID, err := value.ID()
	if err != nil {
		t.Fatal(err)
	}
	switch label {
	case closuregraph.LabelCaptureGraph:
		_, err = closuregraph.DecodeCaptureGraph(ccj)
	case closuregraph.LabelSelectionContext:
		_, err = closuregraph.DecodeSelectionContext(ccj)
	case closuregraph.LabelSelectionBinding:
		_, err = closuregraph.DecodeSelectionBinding(ccj)
	case closuregraph.LabelActiveGraph:
		_, err = closuregraph.DecodeActiveGraph(ccj)
	default:
		t.Fatalf("unsupported shared record label %s", label)
	}
	if err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	return pythonCanonicalRecord{CCJ: string(ccj), ID: string(recordID), Label: label}
}

func mustSharedNodeID(t *testing.T, node closuregraph.Node) closuregraph.ID {
	t.Helper()
	ccj, err := node.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := closuregraph.DecodeNode(ccj); err != nil {
		t.Fatal(err)
	}
	recordID, err := node.ID()
	if err != nil {
		t.Fatal(err)
	}
	return recordID
}

func mustSharedEdgeID(t *testing.T, edge closuregraph.Edge) closuregraph.ID {
	t.Helper()
	ccj, err := edge.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := closuregraph.DecodeEdge(ccj); err != nil {
		t.Fatal(err)
	}
	recordID, err := edge.ID()
	if err != nil {
		t.Fatal(err)
	}
	return recordID
}

func mustWireRecord(t *testing.T, label string, value interface {
	CanonicalBytes() ([]byte, error)
	ID() (closuregraph.ID, error)
}) map[string]any {
	t.Helper()
	record := mustSharedLeafRecord(t, label, value)
	var payload map[string]any
	if err := json.Unmarshal([]byte(record.CCJ), &payload); err != nil {
		t.Fatal(err)
	}
	return map[string]any{"id": record.ID, "label": record.Label, "payload": payload}
}

func mustSharedLeafRecord(t *testing.T, label string, value interface {
	CanonicalBytes() ([]byte, error)
	ID() (closuregraph.ID, error)
}) pythonCanonicalRecord {
	t.Helper()
	ccj, err := value.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	recordID, err := value.ID()
	if err != nil {
		t.Fatal(err)
	}
	return pythonCanonicalRecord{CCJ: string(ccj), ID: string(recordID), Label: label}
}

func recordEnvelope(record pythonCanonicalRecord) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(record.CCJ), &payload); err != nil {
		panic(err)
	}
	return map[string]any{"id": record.ID, "label": record.Label, "payload": payload}
}

func mustCanonicalProtocolRecord(t *testing.T, label string, payload map[string]any) pythonCanonicalRecord {
	t.Helper()
	ccj, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	recordID, err := closuregraph.IDFromCanonical(label, ccj)
	if err != nil {
		t.Fatal(fmt.Errorf("canonical %s: %w", label, err))
	}
	return pythonCanonicalRecord{CCJ: string(ccj), ID: string(recordID), Label: label}
}
