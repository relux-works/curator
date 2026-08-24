package crossconformance_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
)

// The published corpus is an accepted research artifact. This gate proves the
// bytes this package validates are the accepted bytes, and that the copy the
// graph package tests against is the same file, so the cross-adapter proof and
// the owning package can never drift apart silently.
func TestEmbeddedCorpusIsTheAcceptedBytes(t *testing.T) {
	embedded := crossconformance.AcceptedCorpusBytes()
	sum := sha256.Sum256(embedded)
	if got := hex.EncodeToString(sum[:]); got != crossconformance.AcceptedCorpusSHA256 {
		t.Fatalf("embedded corpus SHA-256 = %s, want %s", got, crossconformance.AcceptedCorpusSHA256)
	}
	owner, err := os.ReadFile(filepath.Join("..", "closuregraph", "testdata", "canonical-goldens.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(owner) != string(embedded) {
		t.Fatal("the graph package copy and the cross-adapter copy of the accepted corpus differ")
	}
}

// Every one of the 53 published records is decoded, canonicalized, and hashed
// by this package's own scanner and emitter. The production encoder is not
// consulted, so a change in either implementation surfaces here.
func TestAllPublishedRecordsIndependentlyDeriveTheirHashes(t *testing.T) {
	corpus := mustCorpus(t)
	if len(corpus.Records) != crossconformance.AcceptedRecordCount {
		t.Fatalf("corpus holds %d records, want %d", len(corpus.Records), crossconformance.AcceptedRecordCount)
	}
	for _, record := range corpus.Records {
		if record.Derived != record.Published {
			t.Errorf("%s derived %s, published %s", record.Name, record.Derived, record.Published)
		}
		canonical, err := crossconformance.Canonical(record.Value)
		if err != nil {
			t.Fatalf("%s: %v", record.Name, err)
		}
		if string(canonical) != string(record.Payload) {
			t.Errorf("%s re-emits\n got %s\nwant %s", record.Name, canonical, record.Payload)
		}
	}
}

// The independent oracle and the production identity function must agree on
// every published record. They are separate implementations of CCJ-1 plus
// domain separation; disagreement is a contract break in one of them.
func TestIndependentOracleAgreesWithProductionIdentity(t *testing.T) {
	corpus := mustCorpus(t)
	for _, record := range corpus.Records {
		production, err := closuregraph.IDFromCanonical(record.Label, record.Payload)
		if err != nil {
			t.Fatalf("%s: %v", record.Name, err)
		}
		if string(production) != record.Derived {
			t.Errorf("%s: production %s, independent %s", record.Name, production, record.Derived)
		}
	}
}

// Reference resolution, CGP05 binding-only target authority, and the CGP10
// stable/branch split are all proved from the records themselves.
func TestAcceptedCorpusSatisfiesEveryPublishedStructuralClaim(t *testing.T) {
	report, err := crossconformance.Validate(mustCorpus(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.LabeledRecords != crossconformance.AcceptedRecordCount {
		t.Fatalf("labeled records = %d", report.LabeledRecords)
	}
	if !report.CGP05CaptureReused || report.CGP05TargetBranches != 2 || report.ExplicitTargetEdges != 2 {
		t.Fatalf("CGP05 claims not proved: %+v", report)
	}
	if report.CGP10ObservationSets != 2 || !report.CGP10AllRefsResolve {
		t.Fatalf("CGP10 claims not proved: %+v", report)
	}
	if report.ResolvedReferences < 100 {
		t.Fatalf("only %d typed references resolved; the corpus is barely cross-linked", report.ResolvedReferences)
	}
	want := []string{
		"canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2",
		"canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true",
	}
	got := report.GoldenSummary()
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("summary line %d =\n %s\nwant\n %s", index, got[index], want[index])
		}
	}
}

// A corpus that hides a target-platform node inside capture, replaces a
// capture record from a binding, dangles a reference, or collapses the CGP10
// branch pair must be rejected. Otherwise the validator above proves nothing.
func TestValidatorRejectsEveryTamperedCorpus(t *testing.T) {
	for _, testCase := range []struct {
		name, from, to, want string
	}{
		{
			name: "capture-absorbs-the-target-platform-node",
			from: `"node_ids":["sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50","sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3"]`,
			to:   `"node_ids":["sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50","sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"]`,
			want: "selection-specific node kind",
		},
		{
			name: "binding-replaces-a-capture-node",
			from: `"binding_node_ids":["sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"]`,
			to:   `"binding_node_ids":["sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"]`,
			want: "binds forbidden node kind",
		},
		{
			name: "reference-dangles",
			from: `"c5_checkpoint_id":"sha256:8730ad8567f6874837242937fc00f76f8fddbd77d7e6cb3096625f4abcf1a3a6"`,
			to:   `"c5_checkpoint_id":"sha256:` + strings.Repeat("9", 64) + `"`,
			want: "unresolved fixture record",
		},
		{
			name: "c5-carries-a-graph-record",
			from: `"payload":{"build_plan_id":"sha256:2ecd72056ccf4a3dce3ef8d7b9288f2a895f4eadf6841f35700c7c302fe02ed2"}`,
			to:   `"payload":{"active_graph_id":"sha256:5491c83f7169f4cdd3416382f89205c9e960a13dbaa05ceffaeb428beca25ef0","build_plan_id":"sha256:2ecd72056ccf4a3dce3ef8d7b9288f2a895f4eadf6841f35700c7c302fe02ed2"}`,
			want: "C5 may add no graph record",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload := string(crossconformance.AcceptedCorpusBytes())
			if !strings.Contains(payload, testCase.from) {
				t.Fatalf("tamper anchor %q is absent from the accepted corpus", testCase.from)
			}
			tampered := rehash(t, strings.Replace(payload, testCase.from, testCase.to, 1))
			corpus, err := crossconformance.ParseCorpus([]byte(tampered))
			if err != nil {
				t.Fatalf("tampered corpus did not parse: %v", err)
			}
			_, err = crossconformance.Validate(corpus)
			if err == nil {
				t.Fatalf("tampered corpus %q validated", testCase.name)
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("rejection reason = %v, want it to mention %q", err, testCase.want)
			}
		})
	}
}

// A published hash that no longer matches its own bytes must fail before any
// structural claim is examined.
func TestParseRejectsAPublishedHashThatDoesNotMatchItsBytes(t *testing.T) {
	payload := strings.Replace(string(crossconformance.AcceptedCorpusBytes()), `"size":3}`, `"size":4}`, 1)
	if _, err := crossconformance.ParseCorpus([]byte(payload)); err == nil {
		t.Fatal("a record whose payload no longer hashes to its published identity was accepted")
	}
}

// rehash rewrites every published identity so that a tampered corpus is
// internally hash-consistent. Without this the structural claims would never
// be reached, and the tamper cases above would prove only that hashing works.
func rehash(t *testing.T, payload string) string {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(payload, "\n"), "\n")
	// Two passes: the first fixes each record's own identity, the second
	// rewrites every reference to an identity the first pass changed.
	replacements := map[string]string{}
	for index := 0; index+3 < len(lines); index++ {
		if !strings.HasPrefix(lines[index], "name=") {
			continue
		}
		label := strings.TrimPrefix(lines[index+1], "label=")
		derived, err := crossconformance.DomainID(label, []byte(lines[index+2]))
		if err != nil {
			t.Fatalf("%s: %v", lines[index], err)
		}
		if derived != lines[index+3] {
			replacements[lines[index+3]] = derived
			lines[index+3] = derived
		}
	}
	for round := 0; round < len(lines) && len(replacements) > 0; round++ {
		changed := false
		joined := strings.Join(lines, "\n")
		for from, to := range replacements {
			if strings.Contains(joined, from) {
				joined = strings.ReplaceAll(joined, from, to)
				changed = true
			}
		}
		lines = strings.Split(joined, "\n")
		replacements = map[string]string{}
		for index := 0; index+3 < len(lines); index++ {
			if !strings.HasPrefix(lines[index], "name=") {
				continue
			}
			label := strings.TrimPrefix(lines[index+1], "label=")
			derived, err := crossconformance.DomainID(label, []byte(lines[index+2]))
			if err != nil {
				t.Fatalf("%s: %v", lines[index], err)
			}
			if derived != lines[index+3] {
				replacements[lines[index+3]] = derived
				lines[index+3] = derived
			}
		}
		if !changed && len(replacements) == 0 {
			break
		}
	}
	if len(replacements) != 0 {
		t.Fatalf("tampered corpus did not reach a hash fixed point: %v", replacements)
	}
	return strings.Join(lines, "\n") + "\n"
}

func mustCorpus(t *testing.T) crossconformance.Corpus {
	t.Helper()
	corpus, err := crossconformance.AcceptedCorpus()
	if err != nil {
		t.Fatal(err)
	}
	return corpus
}
