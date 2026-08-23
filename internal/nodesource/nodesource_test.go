package nodesource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

func id(seed string) closuregraph.ID {
	value, _ := closuregraph.DomainID("nodesource-test-v1", map[string]any{"seed": seed})
	return value
}

func TestIndependentPythonProtocolGoldens(t *testing.T) {
	corpusBytes, err := os.ReadFile("testdata/python_protocol_shared_records.json")
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := decodePythonProtocolCorpus(corpusBytes)
	if err != nil {
		t.Fatal(err)
	}
	if corpus.SchemaID != "node-python-protocol-golden-v3" || len(corpus.Cases) != 13 {
		t.Fatalf("invalid protocol corpus schema/count: %s/%d", corpus.SchemaID, len(corpus.Cases))
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 unavailable for independent protocol oracle")
	}
	command := exec.Command(python, "testdata/python_protocol_golden.py")
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	count := 0
	for scanner.Scan() {
		var record pythonOracleRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		if count >= len(corpus.Cases) {
			t.Fatal("Python oracle emitted extra outcomes")
		}
		fixture := corpus.Cases[count]
		if record.CaseID != fixture.ID {
			t.Fatalf("Python oracle case order mismatch: %s != %s", record.CaseID, fixture.ID)
		}
		if fixture.ID == "P10" {
			goOutcomes, goReuse, summary := derivePythonP10SharedOutcomes(t, fixture)
			if !reflect.DeepEqual(summary, fixture.Expected) {
				t.Fatalf("P10 shared summary mismatch: got=%+v expected=%+v", summary, fixture.Expected)
			}
			if len(record.TargetOutcomes) != 2 || record.Outcome != nil || record.ReuseNegative == nil {
				t.Fatalf("P10 must emit two target outcomes and one separate reuse negative: %+v", record)
			}
			assertPythonCanonicalRecords(t, record.TargetOutcomes, goOutcomes)
			assertPythonCanonicalRecords(t, []pythonCanonicalRecord{*record.ReuseNegative}, []pythonCanonicalRecord{goReuse})
			if record.TargetOutcomes[0].ID == record.TargetOutcomes[1].ID {
				t.Fatal("P10 target-scoped outcome identities unexpectedly match")
			}
			assertP10TargetScopedRecords(t, record, fixture.Expected)
			count++
			continue
		}
		if record.Outcome == nil || len(record.TargetOutcomes) != 0 || record.ReuseNegative != nil {
			t.Fatalf("%s emitted an invalid oracle envelope", fixture.ID)
		}
		got, err := closuregraph.IDFromCanonical(record.Outcome.Label, []byte(record.Outcome.CCJ))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != record.Outcome.ID {
			t.Fatalf("python protocol mismatch: %s != %s", got, record.Outcome.ID)
		}
		goPayload, goID, summary := derivePythonProtocolOutcome(t, fixture)
		if !reflect.DeepEqual(summary, fixture.Expected) || goID != closuregraph.ID(record.Outcome.ID) {
			t.Fatalf("%s independent outcome mismatch: go=%s python=%s summary=%+v expected=%+v payload=%v", fixture.ID, goID, record.Outcome.ID, summary, fixture.Expected, goPayload)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 13 {
		t.Fatalf("got %d independent Python vectors", count)
	}
}

func TestBuildCaptureDeduplicatesSharedWorkspaceArtifactManifest(t *testing.T) {
	manifestID := id("shared-workspace-manifest")
	snapshotID := id("shared-workspace-snapshot")
	capture, err := BuildCapture(CaptureInput{
		Manager:  ManagerNPM,
		RootKeys: []string{"root"},
		Packages: []PackageInstance{
			{Key: "root", Name: "root", Version: "1.0.0", Origin: "workspace:.", Checksum: string(snapshotID), ArtifactManifestID: manifestID, SnapshotDigest: snapshotID},
			{Key: "workspace", Name: "workspace", Version: "1.0.0", Origin: "workspace:packages/workspace", Checksum: string(snapshotID), WorkspacePath: "packages/workspace", ArtifactManifestID: manifestID, SnapshotDigest: snapshotID},
		},
		PolicyIDs: []string{"policy-v1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := capture.Graph.ArtifactManifestIDs; len(got) != 1 || got[0] != manifestID {
		t.Fatalf("shared workspace manifest IDs were not deduplicated: %v", got)
	}
}

func assertP10TargetScopedRecords(t *testing.T, record pythonOracleRecord, expected pythonProtocolExpected) {
	t.Helper()
	selectionIDs := map[string]bool{}
	bindingIDs := map[string]bool{}
	activeIDs := map[string]bool{}
	for _, outcome := range record.TargetOutcomes {
		var payload map[string]any
		if err := json.Unmarshal([]byte(outcome.CCJ), &payload); err != nil {
			t.Fatal(err)
		}
		capture := payload["capture_graph"].(map[string]any)
		selection := payload["selection_context"].(map[string]any)
		binding := payload["selection_binding"].(map[string]any)
		active := payload["active_graph"].(map[string]any)
		if capture["id"] != string(expected.CaptureGraphID) || payload["diagnostic"] != nil || payload["decision"] != "admit" {
			t.Fatalf("P10 target outcome is not an admitted branch over the shared capture: %v", payload)
		}
		selectionIDs[selection["id"].(string)] = true
		bindingIDs[binding["id"].(string)] = true
		activeIDs[active["id"].(string)] = true
	}
	if len(selectionIDs) != 2 || len(bindingIDs) != 2 || len(activeIDs) != 2 {
		t.Fatalf("P10 target branches are not independently selected and bound: selection=%v binding=%v active=%v", selectionIDs, bindingIDs, activeIDs)
	}
	for _, expectedID := range expected.BindingIDs {
		if !bindingIDs[string(expectedID)] {
			t.Fatalf("P10 missing expected binding %s", expectedID)
		}
	}
	for _, expectedID := range expected.ActiveGraphIDs {
		if !activeIDs[string(expectedID)] {
			t.Fatalf("P10 missing expected active graph %s", expectedID)
		}
	}
	var reuse map[string]any
	if err := json.Unmarshal([]byte(record.ReuseNegative.CCJ), &reuse); err != nil {
		t.Fatal(err)
	}
	from := reuse["from_selection_binding_id"].(string)
	to := reuse["to_selection_binding_id"].(string)
	if from == to || !bindingIDs[from] || !bindingIDs[to] {
		t.Fatalf("P10 reuse negative does not reference the exact target bindings: %v", reuse)
	}
	if err := validatePythonProtocolDiagnostic(reuse["diagnostic"].(map[string]any)); err != nil {
		t.Fatal(err)
	}
}

type pythonCanonicalRecord struct {
	CCJ   string `json:"ccj"`
	ID    string `json:"id"`
	Label string `json:"label"`
}

type pythonOracleRecord struct {
	CaseID         string                  `json:"case_id"`
	Outcome        *pythonCanonicalRecord  `json:"outcome"`
	TargetOutcomes []pythonCanonicalRecord `json:"target_outcomes"`
	ReuseNegative  *pythonCanonicalRecord  `json:"reuse_negative"`
}

func assertPythonCanonicalRecords(t *testing.T, python, goRecords []pythonCanonicalRecord) {
	t.Helper()
	if !reflect.DeepEqual(python, goRecords) {
		t.Fatalf("independent canonical records differ: python=%+v go=%+v", python, goRecords)
	}
	for _, record := range python {
		got, err := closuregraph.IDFromCanonical(record.Label, []byte(record.CCJ))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != record.ID {
			t.Fatalf("canonical record mismatch: %s != %s", got, record.ID)
		}
	}
}

type pythonProtocolCorpus struct {
	SchemaID string                 `json:"schema_id"`
	Records  []pythonProtocolRecord `json:"records"`
	Cases    []pythonProtocolCase   `json:"cases"`
}

type pythonProtocolRecord struct {
	Name, Label string
	ID          closuregraph.ID
	Payload     map[string]any
}

type pythonProtocolCase struct {
	ID       string                 `json:"id"`
	Input    pythonProtocolFixture  `json:"input"`
	Expected pythonProtocolExpected `json:"expected"`
}

type pythonProtocolExpected struct {
	Decision            string            `json:"decision"`
	DiagnosticCode      string            `json:"diagnostic_code"`
	CaptureGraphID      closuregraph.ID   `json:"capture_graph_id"`
	BindingIDs          []closuregraph.ID `json:"binding_ids"`
	ActiveGraphIDs      []closuregraph.ID `json:"active_graph_ids"`
	ReuseDiagnosticCode string            `json:"reuse_diagnostic_code"`
}

type pythonProtocolPackage struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Dependencies []string `json:"dependencies"`
}

type pythonProtocolTarget struct {
	Interpreter string `json:"interpreter"`
	Platform    string `json:"platform"`
	ABI         string `json:"abi"`
}

type pythonProtocolReuseAttempt struct {
	FromTarget string `json:"from_target"`
	ToTarget   string `json:"to_target"`
}

type pythonProtocolFixture struct {
	SchemaID string                  `json:"schema_id"`
	Packages []pythonProtocolPackage `json:"packages"`
	Lock     struct {
		FormatSupported bool `json:"format_supported"`
		HashesComplete  bool `json:"hashes_complete"`
		GraphComplete   bool `json:"graph_complete"`
		LocalPathEscape bool `json:"local_path_escape"`
	} `json:"lock"`
	Artifact struct {
		Class       string `json:"class"`
		RecordValid bool   `json:"record_valid"`
	} `json:"artifact"`
	Build struct {
		DependenciesLocked bool `json:"dependencies_locked"`
		MetadataMatches    bool `json:"metadata_matches"`
		NetworkAttempted   bool `json:"network_attempted"`
		NativeBuild        bool `json:"native_build"`
	} `json:"build"`
	Targets      []pythonProtocolTarget      `json:"targets"`
	ReuseAttempt *pythonProtocolReuseAttempt `json:"reuse_attempt"`
}

func decodePythonProtocolCorpus(data []byte) (pythonProtocolCorpus, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return pythonProtocolCorpus{}, err
	}
	if err := requireJSONKeys(raw, "corpus", "schema_id", "records", "cases"); err != nil {
		return pythonProtocolCorpus{}, err
	}
	cases, ok := raw["cases"].([]any)
	if !ok {
		return pythonProtocolCorpus{}, fmt.Errorf("cases must be an array")
	}
	for index, item := range cases {
		entry, ok := item.(map[string]any)
		if !ok {
			return pythonProtocolCorpus{}, fmt.Errorf("cases[%d] must be an object", index)
		}
		if err := requireJSONKeys(entry, fmt.Sprintf("cases[%d]", index), "id", "input", "expected"); err != nil {
			return pythonProtocolCorpus{}, err
		}
		input, ok := entry["input"].(map[string]any)
		if !ok {
			return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].input must be an object", index)
		}
		if err := requireJSONKeys(input, fmt.Sprintf("cases[%d].input", index), "schema_id", "packages", "lock", "artifact", "build", "targets", "reuse_attempt"); err != nil {
			return pythonProtocolCorpus{}, err
		}
		for _, nested := range []struct {
			name string
			keys []string
		}{{"lock", []string{"format_supported", "hashes_complete", "graph_complete", "local_path_escape"}}, {"artifact", []string{"class", "record_valid"}}, {"build", []string{"dependencies_locked", "metadata_matches", "network_attempted", "native_build"}}} {
			object, ok := input[nested.name].(map[string]any)
			if !ok {
				return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].input.%s must be an object", index, nested.name)
			}
			if err := requireJSONKeys(object, fmt.Sprintf("cases[%d].input.%s", index, nested.name), nested.keys...); err != nil {
				return pythonProtocolCorpus{}, err
			}
		}
		for _, collection := range []struct {
			name string
			keys []string
		}{{"packages", []string{"name", "version", "dependencies"}}, {"targets", []string{"interpreter", "platform", "abi"}}} {
			items, ok := input[collection.name].([]any)
			if !ok {
				return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].input.%s must be an array", index, collection.name)
			}
			for itemIndex, value := range items {
				object, ok := value.(map[string]any)
				if !ok {
					return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].input.%s[%d] must be an object", index, collection.name, itemIndex)
				}
				if err := requireJSONKeys(object, fmt.Sprintf("cases[%d].input.%s[%d]", index, collection.name, itemIndex), collection.keys...); err != nil {
					return pythonProtocolCorpus{}, err
				}
			}
		}
		if reuse := input["reuse_attempt"]; reuse != nil {
			object, ok := reuse.(map[string]any)
			if !ok {
				return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].input.reuse_attempt must be null or an object", index)
			}
			if err := requireJSONKeys(object, fmt.Sprintf("cases[%d].input.reuse_attempt", index), "from_target", "to_target"); err != nil {
				return pythonProtocolCorpus{}, err
			}
		}
		expected, ok := entry["expected"].(map[string]any)
		if !ok {
			return pythonProtocolCorpus{}, fmt.Errorf("cases[%d].expected must be an object", index)
		}
		if err := requireJSONKeys(expected, fmt.Sprintf("cases[%d].expected", index), "decision", "diagnostic_code", "capture_graph_id", "binding_ids", "active_graph_ids", "reuse_diagnostic_code"); err != nil {
			return pythonProtocolCorpus{}, err
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var corpus pythonProtocolCorpus
	if err := decoder.Decode(&corpus); err != nil {
		return pythonProtocolCorpus{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return pythonProtocolCorpus{}, fmt.Errorf("corpus has trailing JSON")
	}
	return corpus, nil
}

func requireJSONKeys(object map[string]any, path string, keys ...string) error {
	if len(object) != len(keys) {
		return fmt.Errorf("%s has missing or unknown fields", path)
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%s is missing %s", path, key)
		}
	}
	return nil
}

func derivePythonProtocolOutcome(t *testing.T, testCase pythonProtocolCase) (map[string]any, closuregraph.ID, pythonProtocolExpected) {
	t.Helper()
	fixture := testCase.Input
	if fixture.SchemaID != "python-protocol-fixture-v2" || len(fixture.Packages) == 0 || len(fixture.Targets) == 0 {
		t.Fatalf("%s invalid fixture schema", testCase.ID)
	}
	graphIDs := make([]closuregraph.ID, 0, len(fixture.Packages))
	packagesByName := map[string]closuregraph.ID{}
	for _, pkg := range fixture.Packages {
		if pkg.Name == "" || pkg.Version == "" {
			t.Fatalf("%s invalid package", testCase.ID)
		}
		packageID, err := closuregraph.DomainID("python-protocol-package-node-v1", map[string]any{"kind": "package_instance", "name": pkg.Name, "version": pkg.Version})
		if err != nil {
			t.Fatal(err)
		}
		graphIDs = append(graphIDs, packageID)
		if packagesByName[pkg.Name].Valid() {
			t.Fatalf("%s duplicate package %s", testCase.ID, pkg.Name)
		}
		packagesByName[pkg.Name] = packageID
	}
	sort.Slice(graphIDs, func(i, j int) bool { return graphIDs[i] < graphIDs[j] })
	edgeIDs := []closuregraph.ID{}
	for _, pkg := range fixture.Packages {
		for _, dependency := range pkg.Dependencies {
			to := packagesByName[dependency]
			if !to.Valid() {
				t.Fatalf("%s dependency %s is unresolved", testCase.ID, dependency)
			}
			edgeID, err := closuregraph.DomainID("python-protocol-requires-edge-v1", map[string]any{"from_node_id": string(packagesByName[pkg.Name]), "kind": "requires", "to_node_id": string(to)})
			if err != nil {
				t.Fatal(err)
			}
			edgeIDs = append(edgeIDs, edgeID)
		}
	}
	sort.Slice(edgeIDs, func(i, j int) bool { return edgeIDs[i] < edgeIDs[j] })
	capturePayload := map[string]any{"schema_id": "python-protocol-capture-graph-v1", "node_ids": idStrings(graphIDs), "edge_ids": idStrings(edgeIDs)}
	captureID, err := closuregraph.DomainID("python-protocol-capture-graph-v1", capturePayload)
	if err != nil {
		t.Fatal(err)
	}

	bindingIDs, activeIDs := []closuregraph.ID{}, []closuregraph.ID{}
	bindingRecords, activeRecords := []map[string]any{}, []map[string]any{}
	targetBindings := map[string]closuregraph.ID{}
	for _, target := range fixture.Targets {
		if target.Interpreter == "" || target.Platform == "" || target.ABI == "" {
			t.Fatalf("%s has incomplete target", testCase.ID)
		}
		targetKey := target.Interpreter + "/" + target.Platform + "/" + target.ABI
		targetID, err := closuregraph.DomainID("python-protocol-target-node-v1", map[string]any{"kind": "target_platform", "interpreter": target.Interpreter, "platform": target.Platform, "abi": target.ABI})
		if err != nil {
			t.Fatal(err)
		}
		bindingPayload := map[string]any{"schema_id": "python-protocol-selection-binding-v1", "captured_graph_id": string(captureID), "target_node_id": string(targetID)}
		bindingID, _ := closuregraph.DomainID("python-protocol-selection-binding-v1", bindingPayload)
		activePayload := map[string]any{"schema_id": "python-protocol-active-graph-v1", "captured_graph_id": string(captureID), "selection_binding_id": string(bindingID), "node_ids": idStrings(graphIDs), "edge_ids": idStrings(edgeIDs)}
		activeID, _ := closuregraph.DomainID("python-protocol-active-graph-v1", activePayload)
		if targetBindings[targetKey].Valid() {
			t.Fatalf("%s has duplicate target %s", testCase.ID, targetKey)
		}
		targetBindings[targetKey] = bindingID
		bindingIDs = append(bindingIDs, bindingID)
		activeIDs = append(activeIDs, activeID)
		bindingRecords = append(bindingRecords, protocolRecord("python-protocol-selection-binding-v1", bindingID, bindingPayload))
		activeRecords = append(activeRecords, protocolRecord("python-protocol-active-graph-v1", activeID, activePayload))
	}
	sort.Slice(bindingIDs, func(i, j int) bool { return bindingIDs[i] < bindingIDs[j] })
	sort.Slice(activeIDs, func(i, j int) bool { return activeIDs[i] < activeIDs[j] })
	sort.Slice(bindingRecords, func(i, j int) bool { return bindingRecords[i]["id"].(string) < bindingRecords[j]["id"].(string) })
	sort.Slice(activeRecords, func(i, j int) bool { return activeRecords[i]["id"].(string) < activeRecords[j]["id"].(string) })

	var diagnostic any
	processStarted := false
	switch {
	case !fixture.Lock.FormatSupported:
		diagnostic = map[string]any{"code": "closure_lock_format_unsupported", "phase": "C1.resolve"}
	case fixture.Lock.LocalPathEscape:
		diagnostic = map[string]any{"code": "closure_local_path_escape", "phase": "C1.resolve"}
	case !fixture.Lock.HashesComplete:
		diagnostic = map[string]any{"code": "closure_integrity_missing", "phase": "C1.resolve"}
	case !fixture.Lock.GraphComplete:
		diagnostic = map[string]any{"code": "closure_graph_incomplete", "phase": "C1.resolve"}
	case fixture.Artifact.Class == "native.shared-library" || fixture.Artifact.Class == "python.bytecode":
		diagnostic = map[string]any{"code": "artifact_compiled_dependency_forbidden", "phase": "C3.admit"}
	case !fixture.Artifact.RecordValid:
		diagnostic = map[string]any{"code": "closure_metadata_mismatch", "phase": "C3.admit"}
	case !fixture.Build.DependenciesLocked:
		diagnostic = map[string]any{"code": "closure_build_dependency_unlocked", "phase": "C5.plan"}
	case fixture.Build.NativeBuild:
		diagnostic = map[string]any{"code": "closure_native_build_unsupported", "phase": "C5.plan"}
	default:
		processStarted = true
		if fixture.Build.NetworkAttempted {
			diagnostic = map[string]any{"code": "closure_network_attempted", "phase": "C6.offline"}
		} else if !fixture.Build.MetadataMatches {
			diagnostic = map[string]any{"code": "closure_metadata_mismatch", "phase": "C7.publish"}
		}
	}
	admitted := diagnostic == nil
	if !admitted {
		bindingIDs, activeIDs = []closuregraph.ID{}, []closuregraph.ID{}
		bindingRecords, activeRecords = []map[string]any{}, []map[string]any{}
	}
	diagnosticRecord := any(nil)
	if value, ok := diagnostic.(map[string]any); ok {
		payload := map[string]any{"schema_id": "python-protocol-diagnostic-v1", "code": value["code"], "phase": value["phase"]}
		recordID, _ := closuregraph.DomainID("python-protocol-diagnostic-v1", payload)
		diagnosticRecord = protocolRecord("python-protocol-diagnostic-v1", recordID, payload)
	}
	reuseDiagnostic := any(nil)
	reuseCode := ""
	if fixture.ReuseAttempt != nil {
		from := targetBindings[fixture.ReuseAttempt.FromTarget]
		to := targetBindings[fixture.ReuseAttempt.ToTarget]
		if !from.Valid() || !to.Valid() || from == to {
			t.Fatalf("%s reuse attempt does not name two distinct bindings", testCase.ID)
		}
		reuseCode = "closure_target_identity_changed"
		payload := map[string]any{"schema_id": "python-protocol-diagnostic-v1", "code": reuseCode, "phase": "C4.close", "from_binding_id": string(from), "to_binding_id": string(to)}
		recordID, _ := closuregraph.DomainID("python-protocol-diagnostic-v1", payload)
		reuseDiagnostic = protocolRecord("python-protocol-diagnostic-v1", recordID, payload)
	}
	payload := map[string]any{"schema_id": "python-protocol-outcome-v2", "case_id": testCase.ID, "decision": map[bool]string{true: "admit", false: "reject"}[admitted], "diagnostic": diagnosticRecord, "capture_graph": protocolRecord("python-protocol-capture-graph-v1", captureID, capturePayload), "bindings": recordsAsAny(bindingRecords), "active_graphs": recordsAsAny(activeRecords), "reuse_diagnostic": reuseDiagnostic, "process_started": processStarted, "published": admitted}
	outcomeID, err := closuregraph.DomainID("python-protocol-outcome-v1", payload)
	if err != nil {
		t.Fatal(err)
	}
	diagnosticCode := ""
	if value, ok := diagnostic.(map[string]any); ok {
		diagnosticCode = value["code"].(string)
	}
	summary := pythonProtocolExpected{Decision: map[bool]string{true: "admit", false: "reject"}[admitted], DiagnosticCode: diagnosticCode, CaptureGraphID: captureID, BindingIDs: bindingIDs, ActiveGraphIDs: activeIDs, ReuseDiagnosticCode: reuseCode}
	return payload, outcomeID, summary
}

func TestPythonProtocolCorpusRejectsMissingAndUnknownNestedFields(t *testing.T) {
	data, err := os.ReadFile("testdata/python_protocol_shared_records.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, nested := range []string{"lock", "artifact", "build"} {
		t.Run(nested+" missing", func(t *testing.T) {
			mutated := mutateProtocolNestedField(t, data, nested, false)
			if _, err := decodePythonProtocolCorpus(mutated); err == nil {
				t.Fatalf("missing %s field was accepted", nested)
			}
		})
		t.Run(nested+" unknown", func(t *testing.T) {
			mutated := mutateProtocolNestedField(t, data, nested, true)
			if _, err := decodePythonProtocolCorpus(mutated); err == nil {
				t.Fatalf("unknown %s field was accepted", nested)
			}
		})
	}
}

func mutateProtocolNestedField(t *testing.T, data []byte, nested string, addUnknown bool) []byte {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	first := raw["cases"].([]any)[0].(map[string]any)["input"].(map[string]any)[nested].(map[string]any)
	if addUnknown {
		first["unknown_field"] = true
	} else {
		for key := range first {
			delete(first, key)
			break
		}
	}
	mutated, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

func idStrings(ids []closuregraph.ID) []string {
	result := make([]string, len(ids))
	for index, value := range ids {
		result[index] = string(value)
	}
	return result
}

func protocolRecord(label string, id closuregraph.ID, payload map[string]any) map[string]any {
	return map[string]any{"label": label, "id": string(id), "payload": payload}
}

func recordsAsAny(records []map[string]any) []any {
	result := make([]any, len(records))
	for index := range records {
		result[index] = records[index]
	}
	return result
}

func TestNodeCGP05AcceptedExactRecordDigests(t *testing.T) {
	want := map[string]string{
		"cgp05.capture":          "sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2",
		"cgp05.platform.darwin":  "sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b",
		"cgp05.platform.linux":   "sha256:17527a5f8337dc55fb9390ac4671179d1dc14ec6433e5d2b6324314cd4fe0367",
		"cgp05.selection.darwin": "sha256:eb95dae28edbea38b77d2a0f7d702d0fa214e7126f714ef33ff857ff40b435e7",
		"cgp05.selection.linux":  "sha256:44523f093fb255a4b9015ac640faa9f16cc873817248754222c049c23f8849e2",
		"cgp05.targets.darwin":   "sha256:b3e04cda6e0419ff5d8281ab20dc1ab05986cbd5541a3b50f805c2a469e9578d",
		"cgp05.targets.linux":    "sha256:6cb771139f53d164abcef33d460c427da94c74009989e9560960cc82ba88c430",
		"cgp05.binding.darwin":   "sha256:5e2b1414ffeef8a4d8100c18e06a13ca8853515244025291db658096dcc2770f",
		"cgp05.binding.linux":    "sha256:ae2c74d58e22e117c217c462c3eccc510d137b84dff4961a3aa28aba5a1ceb26",
		"cgp05.active.darwin":    "sha256:c8a8a70de7cece61ad01eabebd6171cd9af62ae11b783761cd4562cb4fb9e3e8",
		"cgp05.active.linux":     "sha256:ad0b68d8939eb9e1ab74f8fec4b164e418096dad244ece858e02e7d4b34beabc",
		"cgp05.plan.darwin":      "sha256:e71959451bf642d7860caaac2c14d0ea8c18bcc27d5d743255f5afc94de5a139",
		"cgp05.plan.linux":       "sha256:3f568ae879be512fdf808f11ce9ed534337ff597c49be265dab972fa406a3b75",
		"cgp05.c4.darwin":        "sha256:af7645b4942ff42586144fdf69455166a13ad7dcab5e2ea7d08b083b0b8dd2cf",
		"cgp05.c4.linux":         "sha256:f35a82ebc0b023aa466c629060545f3b4ea64ea2bb7d1eb0775a56337404370a",
		"cgp05.c5.darwin":        "sha256:74b020032b7466dafdf5ed33e35a57008884fe4086f53990713ba4d21d506b14",
		"cgp05.c5.linux":         "sha256:10ee2b87c903e1b74c95568fa343c03e1ac638a93b9919a1b90dacd062931e97",
	}
	contents, err := os.ReadFile("../../.research/260811_cross-language-closure-graph-and-checkpoints.md")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(contents), "\n")
	seen := map[string]bool{}
	for index, line := range lines {
		if !strings.HasPrefix(line, "name=cgp05.") || index+3 >= len(lines) {
			continue
		}
		name := strings.TrimPrefix(line, "name=")
		expected, required := want[name]
		if !required {
			continue
		}
		label := strings.TrimPrefix(lines[index+1], "label=")
		actual, err := closuregraph.IDFromCanonical(label, []byte(lines[index+2]))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(actual) != expected || lines[index+3] != expected {
			t.Fatalf("%s digest = %s record = %s, want %s", name, actual, lines[index+3], expected)
		}
		seen[name] = true
	}
	if len(seen) != len(want) {
		t.Fatalf("validated %d/%d exact CGP05 records", len(seen), len(want))
	}
}

func packageSet() []PackageInstance {
	return []PackageInstance{
		{Key: "app@workspace", Name: "app", Version: "1.0.0", Origin: "workspace:workspace", Checksum: "sha256-root", WorkspacePath: "workspace", ArtifactManifestID: id("am-app"), SnapshotDigest: id("tree-app"), Dependencies: []Dependency{{PackageKey: "peer@1[p=react@18]", Scope: closuregraph.ScopePeer, DeclarationField: "peerDependencies.react"}}},
		{Key: "peer@1[p=react@18]", Name: "peer", Version: "1.0.0", Origin: "https://registry.invalid/peer.tgz", Checksum: "sha512-peer", PeerKey: "react@18", ArtifactManifestID: id("am-peer"), SnapshotDigest: id("tree-peer")},
	}
}

func TestManagerProfilesEmitSameSelectionNeutralCapture(t *testing.T) {
	profiles := []ManagerProfile{ManagerNPM, ManagerPNPM, ManagerYarnClassic, ManagerYarnModern}
	var want closuregraph.ID
	for _, profile := range profiles {
		capture, err := BuildCapture(CaptureInput{Manager: profile, RootKeys: []string{"app@workspace"}, Packages: packageSet(), PolicyIDs: []string{"artifact-policy-v1"}})
		if err != nil {
			t.Fatalf("%s: %v", profile, err)
		}
		got, _ := capture.Graph.ID()
		if want == "" {
			want = got
		} else if got != want {
			t.Fatalf("manager changed common capture: %s != %s", got, want)
		}
		for _, node := range capture.Nodes {
			if node.Kind == closuregraph.NodeTargetPlatform || node.Kind == closuregraph.NodeToolchainComponent {
				t.Fatalf("selection record leaked into capture")
			}
		}
	}
}

func TestN04N05LifecycleNativeAndExtensionRejection(t *testing.T) {
	cases := []struct {
		name, code string
		mutate     func(*PackageInstance)
	}{
		{"lifecycle", CodeHookUndeclared, func(p *PackageInstance) { p.LifecycleScripts = []string{"postinstall"} }},
		{"implicit-gyp", CodeNativeBuildUnsupported, func(p *PackageInstance) { p.BindingGYP = true }},
		{"plugin", CodeManagerPluginUndeclared, func(p *PackageInstance) { p.ManagerExtensions = []string{".pnpmfile.cjs"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packages := packageSet()
			tc.mutate(&packages[1])
			_, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: packages, PolicyIDs: []string{"p"}})
			if ErrorCode(err) != tc.code {
				t.Fatalf("got %v, want %s", err, tc.code)
			}
		})
	}
}

func TestTargetAndRuntimeLiveOnlyInDistinctBindings(t *testing.T) {
	capture, err := BuildCapture(CaptureInput{Manager: ManagerNPM, RootKeys: []string{"app@workspace"}, Packages: packageSet(), PolicyIDs: []string{"p"}})
	if err != nil {
		t.Fatal(err)
	}
	root := capture.ProductNodeIDs["app@workspace"]
	makeBinding := func(os string) (closuregraph.ID, closuregraph.ID, closuregraph.ID) {
		platform := closuregraph.TargetPlatformPayload{OS: os, Architecture: "arm64", ABI: "node", Libc: "none", MinimumRuntime: "22", SDKID: "none", TargetTriple: "arm64-" + os, Runtime: "node", LanguageModes: map[string]string{"module": "esm"}, Tuning: map[string]string{}}
		platformNode := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "node.platform.target", Payload: platform}
		platformID, _ := platformNode.ID()
		selection, selectionErr := closuregraph.NewSelectionContext([]closuregraph.ID{root}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID}, []string{}, false, map[string]string{"os": os}, map[string]string{"react": "18"}, []string{})
		if selectionErr != nil {
			t.Fatal(selectionErr)
		}
		tool := func(role, exe string) ToolIdentity {
			return ToolIdentity{Role: role, PolicySelector: role + "-v1", ExecutableRelativePath: exe, VersionOutput: role + " 1", PlatformABI: os + "-arm64", Fingerprint: id(role + os), ExecutableSHA256: id(role + os + "-exe")}
		}
		exact := RuntimeBinding{Platform: platform, Node: tool("node-runtime", "bin/node"), Manager: tool("package-manager", "bin/npm"), TargetNodeIDs: []closuregraph.ID{root}}
		c0, profileErr := NewC0Checkpoint(capture, selection, exact)
		if profileErr != nil {
			t.Fatal(profileErr)
		}
		exact.C0Checkpoint = &c0
		binding, bindingNodes, bindingEdges, authority, bindErr := Bind(capture, selection, exact)
		if bindErr != nil {
			t.Fatal(bindErr)
		}
		bundle, projectErr := closuregraph.ProjectActive(capture.Graph, selection, binding, closuregraph.NewRecordTables(capture.Nodes, capture.Edges, bindingNodes, bindingEdges), authority, nil)
		if projectErr != nil {
			t.Fatal(projectErr)
		}
		bindingID, _ := binding.ID()
		captureID, _ := capture.Graph.ID()
		activeID, _ := bundle.Active.ID()
		return bindingID, captureID, activeID
	}
	darwinBinding, darwinCapture, darwinActive := makeBinding("darwin")
	linuxBinding, linuxCapture, linuxActive := makeBinding("linux")
	if darwinCapture != linuxCapture {
		t.Fatal("target changed capture identity")
	}
	if darwinBinding == linuxBinding {
		t.Fatal("target/runtime did not change binding identity")
	}
	if darwinActive == linuxActive {
		t.Fatal("target/runtime did not change active identity")
	}
}

func TestDeclaredTypeScriptLineageAndExactOutputs(t *testing.T) {
	spec := GeneratedAction{Name: "compile", Argv: []string{"--project"}, WorkingDirectory: "workspace", Compiler: ToolIdentity{Role: "typescript-compiler", PolicySelector: "typescript-v1", ExecutableRelativePath: "bin/tsc", VersionOutput: "5.9", PlatformABI: "node", Fingerprint: id("tsc"), ExecutableSHA256: id("tsc-exe"), ExecutionDomain: closuregraph.ExecutionHost}, Inputs: []GeneratedInput{{NodeID: id("source"), Path: "workspace/src", Class: "source.typescript", Role: "source"}, {NodeID: id("tsconfig"), Path: "workspace/tsconfig.json", Class: "source.config", Role: "config"}, {NodeID: id("plugin"), Path: "workspace/plugin.js", Class: "source.javascript", Role: "plugin"}}, TargetNodeID: id("target"), EnvironmentPolicyID: "node-env-v1", ProcessPolicyID: "node-process-v1", Outputs: []GeneratedOutput{{Path: "dist/cli.js", Class: "source.generated_text", Grammar: "javascript-v1"}}}
	action, nodes, edges, err := BuildGeneratedAction(spec)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != closuregraph.NodeAction || len(nodes) != 1 || len(edges) != 6 {
		t.Fatalf("unexpected lineage sizes: nodes=%d edges=%d", len(nodes), len(edges))
	}
	if err := ValidateObservedOutputs(spec.Outputs, map[string]closuregraph.ID{"dist/cli.js": id("bytes")}); err != nil {
		t.Fatal(err)
	}
	err = ValidateObservedOutputs(spec.Outputs, map[string]closuregraph.ID{"dist/cli.js": id("bytes"), "dist/extra.js": id("extra")})
	if ErrorCode(err) != CodeGeneratedOutputDrift {
		t.Fatalf("got %v", err)
	}
	_, _, _, err = BuildGeneratedAction(GeneratedAction{Name: "implicit"})
	if ErrorCode(err) != CodeBuildDependencyUnlocked || !strings.Contains(err.Error(), "lineage") {
		t.Fatalf("got %v", err)
	}
}
