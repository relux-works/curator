package closuregraph

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

type goldenRecord struct {
	Name    string
	Label   string
	Payload []byte
	ID      ID
	Decoded any
}

func TestAcceptedGoldenCorpusBytesArePinned(t *testing.T) {
	payload, err := os.ReadFile("testdata/canonical-goldens.txt")
	if err != nil {
		t.Fatal(err)
	}
	const want = "fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb"
	if got := fmt.Sprintf("%x", sha256.Sum256(payload)); got != want {
		t.Fatalf("accepted golden corpus SHA-256 = %s, want %s", got, want)
	}
}

func loadGoldenRecords(t *testing.T) map[string]goldenRecord {
	t.Helper()
	payload, err := os.ReadFile("testdata/canonical-goldens.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	records := map[string]goldenRecord{}
	for index := 0; index < len(lines); index++ {
		if !strings.HasPrefix(lines[index], "name=") {
			continue
		}
		if index+3 >= len(lines) {
			t.Fatalf("truncated record at line %d", index+1)
		}
		record := goldenRecord{Name: strings.TrimPrefix(lines[index], "name="), Label: strings.TrimPrefix(lines[index+1], "label="), Payload: []byte(lines[index+2]), ID: ID(lines[index+3])}
		if !strings.HasPrefix(lines[index+1], "label=") || !record.ID.Valid() {
			t.Fatalf("malformed record %q", record.Name)
		}
		if _, exists := records[record.Name]; exists {
			t.Fatalf("duplicate record %q", record.Name)
		}
		if err := protocoljson.RequireCanonical(record.Payload); err != nil {
			t.Fatalf("%s payload: %v", record.Name, err)
		}
		actualID, err := IDFromCanonical(record.Label, record.Payload)
		if err != nil {
			t.Fatalf("%s ID: %v", record.Name, err)
		}
		if actualID != record.ID {
			t.Fatalf("%s ID = %s, want %s", record.Name, actualID, record.ID)
		}
		record.Decoded = decodeGoldenOwnedRecord(t, record)
		records[record.Name] = record
		index += 3
	}
	if len(records) != 53 {
		t.Fatalf("loaded %d records, want 53", len(records))
	}
	return records
}

func decodeGoldenOwnedRecord(t *testing.T, record goldenRecord) any {
	t.Helper()
	var decoded any
	var canonical []byte
	var id ID
	var err error
	switch record.Label {
	case LabelNode:
		value, decodeErr := DecodeNode(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelEdge:
		value, decodeErr := DecodeEdge(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelCaptureGraph:
		value, decodeErr := DecodeCaptureGraph(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelSelectionContext:
		value, decodeErr := DecodeSelectionContext(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelSelectionBinding:
		value, decodeErr := DecodeSelectionBinding(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelActiveGraph:
		value, decodeErr := DecodeActiveGraph(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelBuildPlan:
		value, decodeErr := DecodeBuildPlan(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelCheckpoint:
		value, decodeErr := DecodeCheckpoint(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelSourceClosure:
		value, decodeErr := DecodeSourceClosure(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelExpectedCacheInput:
		value, decodeErr := DecodeExpectedCacheInput(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelProducedArtifactObservation:
		value, decodeErr := DecodeProducedArtifactObservation(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelExecutionReceipt:
		value, decodeErr := DecodeExecutionReceipt(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	case LabelPublicationReceipt:
		value, decodeErr := DecodePublicationReceipt(record.Payload)
		decoded, err = value, decodeErr
		if err == nil {
			canonical, err = value.CanonicalBytes()
		}
		if err == nil {
			id, err = value.ID()
		}
	default:
		// The artifact manifest belongs to the artifact-admission task and the
		// fixture anchor is deliberately not a production checkpoint schema.
		// Their exact labels and CCJ bytes are still independently hashed above.
		return nil
	}
	if err != nil {
		t.Fatalf("decode %s (%s): %v", record.Name, record.Label, err)
	}
	if !bytes.Equal(canonical, record.Payload) {
		t.Fatalf("%s re-encoded as\n%s\nwant\n%s", record.Name, canonical, record.Payload)
	}
	if id != record.ID {
		t.Fatalf("%s typed ID = %s, want %s", record.Name, id, record.ID)
	}
	return decoded
}

func TestAcceptedCGP05AndCGP10RecordsRoundTripAndHash(t *testing.T) {
	records := loadGoldenRecords(t)
	required := []string{
		"cgp05.capture", "cgp05.platform.darwin", "cgp05.platform.linux", "cgp05.selection.darwin", "cgp05.selection.linux", "cgp05.targets.darwin", "cgp05.targets.linux", "cgp05.binding.darwin", "cgp05.binding.linux", "cgp05.active.darwin", "cgp05.active.linux", "cgp05.plan.darwin", "cgp05.plan.linux", "cgp05.c4.darwin", "cgp05.c4.linux", "cgp05.c5.darwin", "cgp05.c5.linux",
		"cgp10.artifact-manifest", "cgp10.product", "cgp10.source", "cgp10.action", "cgp10.output", "cgp10.declares", "cgp10.reads", "cgp10.produces", "cgp10.publishes", "cgp10.capture", "cgp10.platform", "cgp10.toolchain", "cgp10.selection", "cgp10.uses-tool", "cgp10.targets.product", "cgp10.targets.action", "cgp10.targets.toolchain", "cgp10.targets.output", "cgp10.binding", "cgp10.active", "cgp10.plan", "cgp10.c4", "cgp10.c5", "cgp10.closure", "cgp10.expected-cache-input", "cgp10.observation.one", "cgp10.observation.two", "cgp10.execution.one", "cgp10.execution.two", "cgp10.publication.one", "cgp10.publication.two",
	}
	for _, name := range required {
		if _, present := records[name]; !present {
			t.Errorf("missing required golden %s", name)
		}
	}
}

func TestCGP05TargetBranchesReuseCaptureAndProjectExactly(t *testing.T) {
	records := loadGoldenRecords(t)
	capture := mustGolden[CaptureGraph](t, records, "cgp05.capture")
	captureNodes := []Node{mustGolden[Node](t, records, "cgp05.root"), mustGolden[Node](t, records, "cgp05.extra")}
	captureEdges := []Edge{mustGolden[Edge](t, records, "cgp05.requires")}
	for _, branch := range []string{"darwin", "linux"} {
		branch := branch
		t.Run(branch, func(t *testing.T) {
			platform := mustGolden[Node](t, records, "cgp05.platform."+branch)
			target := mustGolden[Edge](t, records, "cgp05.targets."+branch)
			selection := mustGolden[SelectionContext](t, records, "cgp05.selection."+branch)
			binding := mustGolden[SelectionBinding](t, records, "cgp05.binding."+branch)
			wantActive := mustGolden[ActiveGraph](t, records, "cgp05.active."+branch)
			tables := NewRecordTables(captureNodes, captureEdges, []Node{platform}, []Edge{target})
			evaluator := ConditionEvaluatorFunc{EvaluatorID: "fixture-target-v1", EvaluateFunc: func(_ Condition, input EvaluationInput) (bool, error) {
				platformID := input.Selection.PlatformRoles[PlatformTarget]
				for _, node := range input.Records.BindingNodes {
					id, _ := node.ID()
					if id == platformID {
						return node.Payload.(TargetPlatformPayload).OS == "linux", nil
					}
				}
				return false, fmt.Errorf("target platform missing")
			}}
			bundle, err := ProjectActive(capture, selection, binding, tables, BindingAuthority{}, []ConditionEvaluator{evaluator})
			if err != nil {
				t.Fatal(err)
			}
			if got, _ := bundle.Active.CanonicalBytes(); !bytes.Equal(got, records["cgp05.active."+branch].Payload) {
				t.Fatalf("active bytes\n got %s\nwant %s", got, records["cgp05.active."+branch].Payload)
			}
			if !reflect.DeepEqual(bundle.Active, wantActive) {
				t.Fatalf("active graph mismatch\n got %#v\nwant %#v", bundle.Active, wantActive)
			}
			plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1"})
			if err != nil {
				t.Fatal(err)
			}
			if got, _ := plan.CanonicalBytes(); !bytes.Equal(got, records["cgp05.plan."+branch].Payload) {
				t.Fatalf("plan bytes\n got %s\nwant %s", got, records["cgp05.plan."+branch].Payload)
			}
		})
	}
	captureID, _ := capture.ID()
	if captureID != records["cgp05.capture"].ID {
		t.Fatal("capture identity changed")
	}
	for _, kind := range []string{"platform", "selection", "binding", "active", "plan", "c4", "c5"} {
		if records["cgp05."+kind+".darwin"].ID == records["cgp05."+kind+".linux"].ID {
			t.Errorf("%s branch IDs unexpectedly equal", kind)
		}
	}
}

func TestCGP10GraphPlanAndObservationBranches(t *testing.T) {
	records := loadGoldenRecords(t)
	captureNodes := []Node{mustGolden[Node](t, records, "cgp10.product"), mustGolden[Node](t, records, "cgp10.source"), mustGolden[Node](t, records, "cgp10.action"), mustGolden[Node](t, records, "cgp10.output")}
	captureEdges := []Edge{mustGolden[Edge](t, records, "cgp10.declares"), mustGolden[Edge](t, records, "cgp10.reads"), mustGolden[Edge](t, records, "cgp10.produces"), mustGolden[Edge](t, records, "cgp10.publishes")}
	bindingNodes := []Node{mustGolden[Node](t, records, "cgp10.platform"), mustGolden[Node](t, records, "cgp10.toolchain")}
	bindingEdges := []Edge{mustGolden[Edge](t, records, "cgp10.uses-tool"), mustGolden[Edge](t, records, "cgp10.targets.product"), mustGolden[Edge](t, records, "cgp10.targets.action"), mustGolden[Edge](t, records, "cgp10.targets.toolchain"), mustGolden[Edge](t, records, "cgp10.targets.output")}
	bundle, err := ProjectActive(mustGolden[CaptureGraph](t, records, "cgp10.capture"), mustGolden[SelectionContext](t, records, "cgp10.selection"), mustGolden[SelectionBinding](t, records, "cgp10.binding"), NewRecordTables(captureNodes, captureEdges, bindingNodes, bindingEdges), c4AuthorityForNode(t, bindingNodes[1]), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := bundle.Active.CanonicalBytes(); !bytes.Equal(got, records["cgp10.active"].Payload) {
		t.Fatalf("active bytes\n got %s\nwant %s", got, records["cgp10.active"].Payload)
	}
	plan, err := DeriveBuildPlan(bundle, PlanOptions{ExecutionPolicyID: "fixture-execution-v1", LastCheckpointID: records["cgp10.c4"].ID})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := plan.CanonicalBytes(); !bytes.Equal(got, records["cgp10.plan"].Payload) {
		t.Fatalf("plan bytes\n got %s\nwant %s", got, records["cgp10.plan"].Payload)
	}
	for _, branch := range []string{"one", "two"} {
		observation := mustGolden[ProducedArtifactObservation](t, records, "cgp10.observation."+branch)
		if err := observation.ValidateAgainst(bundle.Records); err != nil {
			t.Fatalf("%s observation: %v", branch, err)
		}
		observationID, _ := observation.ID()
		execution := mustGolden[ExecutionReceipt](t, records, "cgp10.execution."+branch)
		if execution.ClosureID != records["cgp10.closure"].ID || !reflect.DeepEqual(execution.ProducedObservationIDs, []ID{observationID}) {
			t.Fatalf("%s execution references = %#v", branch, execution)
		}
		executionID, _ := execution.ID()
		publication := mustGolden[PublicationReceipt](t, records, "cgp10.publication."+branch)
		if publication.ExecutionReceiptID != executionID || publication.ExpectedCacheInputID != records["cgp10.expected-cache-input"].ID || !reflect.DeepEqual(publication.PublishedObservationIDs, []ID{observationID}) {
			t.Fatalf("%s publication references = %#v", branch, publication)
		}
	}
	c4 := mustGolden[Checkpoint](t, records, "cgp10.c4")
	c5 := mustGolden[Checkpoint](t, records, "cgp10.c5")
	if c4.Payload.(C4ClosePayload).ActiveGraphID != records["cgp10.active"].ID || c5.PreviousCheckpointID == nil || *c5.PreviousCheckpointID != records["cgp10.c4"].ID || c5.Payload.(C5PlanPayload).BuildPlanID != records["cgp10.plan"].ID {
		t.Fatalf("C4/C5 chain references do not resolve")
	}
	stableNames := []string{"action", "output", "declares", "reads", "produces", "publishes", "capture", "platform", "toolchain", "selection", "uses-tool", "targets.product", "targets.action", "targets.toolchain", "targets.output", "binding", "active", "plan", "c4", "c5", "closure", "expected-cache-input"}
	for _, name := range stableNames {
		if records["cgp10."+name].ID == "" {
			t.Errorf("stable record %s is absent", name)
		}
	}
	for _, kind := range []string{"observation", "execution", "publication"} {
		if records["cgp10."+kind+".one"].ID == records["cgp10."+kind+".two"].ID {
			t.Errorf("%s branches unexpectedly equal", kind)
		}
	}
}

func mustGolden[T any](t *testing.T, records map[string]goldenRecord, name string) T {
	t.Helper()
	record, ok := records[name]
	if !ok {
		t.Fatalf("missing golden %s", name)
	}
	value, ok := record.Decoded.(T)
	if !ok {
		var zero T
		t.Fatalf("golden %s decoded as %T, want %T", name, record.Decoded, zero)
	}
	return value
}

func TestGoldenCorpusNamesAreCanonical(t *testing.T) {
	records := loadGoldenRecords(t)
	names := make([]string, 0, len(records))
	for name := range records {
		names = append(names, name)
	}
	sort.Strings(names)
	if names[0] != "cgp05.active.darwin" || names[len(names)-1] != "cgp10.uses-tool" {
		t.Fatalf("unexpected accepted corpus bounds %q .. %q", names[0], names[len(names)-1])
	}
}
