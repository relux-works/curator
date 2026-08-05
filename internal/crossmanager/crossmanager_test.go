package crossmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func writeCorpus(t *testing.T, protocol string, files map[string][]byte) string {
	t.Helper()
	root := t.TempDir()
	type entry struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	manifestFiles := make([]entry, 0, len(files))
	for _, name := range names {
		payload := files[name]
		digest := sha256.Sum256(payload)
		manifestFiles = append(manifestFiles, entry{name, "sha256:" + hex.EncodeToString(digest[:]), int64(len(payload))})
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := struct {
		SchemaVersion   int     `json:"schema_version"`
		CorpusVersion   string  `json:"corpus_version"`
		ProtocolVersion string  `json:"protocol_version"`
		GeneratedAt     string  `json:"generated_at"`
		Generator       string  `json:"generator"`
		Files           []entry `json:"files"`
	}{1, CorpusRC5, protocol, "2000-01-01T00:00:00Z", "test", manifestFiles}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCorpusBoundaryAuthenticatesReplaceableInput(t *testing.T) {
	root := writeCorpus(t, ProtocolRC5, map[string][]byte{
		"fixtures/external-repository/a.json": []byte("a"),
		"vectors/external-repository.json":    []byte("vector"),
	})
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	wantEntries := []string{"fixtures/external-repository/a.json", "vectors/external-repository.json"}
	if got := corpus.Entries(""); !reflect.DeepEqual(got, wantEntries) {
		t.Fatalf("entries = %v", got)
	}
	payload, _, err := corpus.Read("vectors/external-repository.json")
	if err != nil || string(payload) != "vector" {
		t.Fatalf("read authenticated vector = %q, %v", payload, err)
	}
	if _, _, err := corpus.Read("not-listed.json"); err == nil {
		t.Fatal("unlisted corpus file must be rejected")
	}
	if err := os.WriteFile(filepath.Join(root, "vectors", "external-repository.json"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := corpus.Read("vectors/external-repository.json"); err == nil || !strings.Contains(err.Error(), "does not match manifest") {
		t.Fatalf("tampered corpus read error = %v", err)
	}
}

func TestCorpusBoundaryRejectsProtocolDrift(t *testing.T) {
	root := writeCorpus(t, "1.0.0-rc.6", map[string][]byte{"vector.json": []byte("x")})
	if _, err := OpenCorpus(root, RC5Boundary); err == nil || !strings.Contains(err.Error(), "does not match boundary") {
		t.Fatalf("protocol drift error = %v", err)
	}
}

func TestCorpusBoundaryRejectsMalformedContract(t *testing.T) {
	root := writeCorpus(t, ProtocolRC5, map[string][]byte{"vector.json": []byte("x")})
	for _, boundary := range []Boundary{
		{Version: "future/v2", ProtocolVersion: ProtocolRC5},
		{Version: CorpusBoundaryV1},
	} {
		if _, err := OpenCorpus(root, boundary); err == nil {
			t.Fatalf("boundary %#v must be rejected", boundary)
		}
	}
	if err := validateCorpusPath("../escape"); err == nil {
		t.Fatal("escaping corpus path must be rejected")
	}
	if _, err := parseDigest("sha256:not-a-digest"); err == nil {
		t.Fatal("malformed corpus digest must be rejected")
	}
}

func TestMaterializeCorpusFilesIsDeterministicAndProvenanced(t *testing.T) {
	root := writeCorpus(t, ProtocolRC5, map[string][]byte{"fixtures/a": []byte("alpha"), "fixtures/b": []byte("beta")})
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "fixture")
	got, err := MaterializeCorpusFiles(corpus, destination, []CorpusFile{
		{Source: "fixtures/b", Target: "z/b", Mode: 0o600},
		{Source: "fixtures/a", Target: "a", Mode: 0o644},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Target != "a" || got[1].Target != "z/b" {
		t.Fatalf("materialization order = %#v", got)
	}
	payload, err := os.ReadFile(filepath.Join(destination, "z", "b"))
	if err != nil || string(payload) != "beta" {
		t.Fatalf("materialized payload = %q, %v", payload, err)
	}
	if got[1].Source != "fixtures/b" || !strings.HasPrefix(got[1].SHA256, "sha256:") {
		t.Fatalf("provenance = %#v", got[1])
	}
	if _, err := MaterializeCorpusFiles(corpus, t.TempDir(), []CorpusFile{{Source: "fixtures/a", Target: "../escape", Mode: 0o644}}); err == nil {
		t.Fatal("escaping fixture target must be rejected")
	}
}

func TestMaterializeCorpusFilesRejectsSymlinkParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated privileges")
	}
	root := writeCorpus(t, ProtocolRC5, map[string][]byte{"fixtures/a": []byte("alpha")})
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(destination, "linked")); err != nil {
		t.Fatal(err)
	}
	_, err = MaterializeCorpusFiles(corpus, destination, []CorpusFile{{Source: "fixtures/a", Target: "linked/a", Mode: 0o644}})
	if err == nil || !strings.Contains(err.Error(), "not a real directory") {
		t.Fatalf("symlink parent error = %v", err)
	}
}

func TestProcessRunnerUsesJSONBlackBoxContract(t *testing.T) {
	runner := ProcessRunner{Executable: os.Args[0], Args: []string{"-test.run=TestProcessRunnerHelper", "--"}, Env: []string{"GO_WANT_CONFORMANCE_HELPER=1"}}
	request := RunRequest{Boundary: CorpusBoundaryV1, CorpusRoot: "/candidate", FixtureRoot: "/fixture", Case: "raw-object"}
	result, err := runner.Run(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 7 || string(result.Stdout) != "raw-object\n" || !strings.HasPrefix(string(result.Stderr), "helper observation\n") {
		t.Fatalf("runner result = %#v", result)
	}
}

func TestProcessRunnerRejectsInvalidConfiguration(t *testing.T) {
	request := RunRequest{Boundary: CorpusBoundaryV1}
	if _, err := (ProcessRunner{}).Run(context.Background(), request); err == nil || !strings.Contains(err.Error(), "executable") {
		t.Fatalf("missing executable error = %v", err)
	}
	request.Boundary = "future/v2"
	if _, err := (ProcessRunner{Executable: os.Args[0]}).Run(context.Background(), request); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("future boundary error = %v", err)
	}
}

func TestProcessRunnerHelper(_ *testing.T) {
	if os.Getenv("GO_WANT_CONFORMANCE_HELPER") != "1" {
		return
	}
	payload, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(90)
	}
	var request RunRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		os.Exit(91)
	}
	if _, err := fmt.Fprintln(os.Stdout, request.Case); err != nil {
		os.Exit(92)
	}
	if _, err := fmt.Fprintln(os.Stderr, "helper observation"); err != nil {
		os.Exit(93)
	}
	os.Exit(7)
}

func TestReportJSONIsDeterministicAndSchemaIsMachineReadable(t *testing.T) {
	report := Report{
		Schema:  ReportSchemaV1,
		Corpus:  CorpusEvidence{Boundary: CorpusBoundaryV1, ProtocolVersion: ProtocolRC5, Revision: strings.Repeat("d", 40), ManifestSHA256: "sha256:" + strings.Repeat("a", 64)},
		Adapter: "fixture-adapter/v1",
		Manager: testManagerEvidence("fixture-adapter"),
		Cases:   []CaseResult{{ID: "z", Status: CaseNotRun, ExitCode: -1}, {ID: "a", Status: CaseObserved, ExitCode: 0}},
	}
	payload, err := ReportJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(payload), `"id": "a"`) > strings.Index(string(payload), `"id": "z"`) {
		t.Fatalf("report cases are not sorted: %s", payload)
	}
	var schema map[string]any
	if err := json.Unmarshal(ReportSchema(), &schema); err != nil {
		t.Fatalf("report schema JSON: %v", err)
	}
	if schema["$id"] != ReportSchemaV1 {
		t.Fatalf("report schema id = %v", schema["$id"])
	}
}

func TestReportJSONRejectsAmbiguousResults(t *testing.T) {
	valid := Report{
		Schema:  ReportSchemaV1,
		Corpus:  CorpusEvidence{Boundary: CorpusBoundaryV1, ProtocolVersion: ProtocolRC5, Revision: strings.Repeat("d", 40), ManifestSHA256: "sha256:" + strings.Repeat("a", 64)},
		Adapter: "fixture-adapter/v1",
		Manager: testManagerEvidence("fixture-adapter"),
		Cases:   []CaseResult{{ID: "same", Status: CaseObserved}},
	}
	duplicate := valid
	duplicate.Cases = append(duplicate.Cases, duplicate.Cases[0])
	if _, err := ReportJSON(duplicate); err == nil || !strings.Contains(err.Error(), "duplicate case") {
		t.Fatalf("duplicate report case error = %v", err)
	}
	invalidStatus := valid
	invalidStatus.Cases[0].Status = "passed"
	if _, err := ReportJSON(invalidStatus); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("qualification-like status error = %v", err)
	}
}

func testManagerEvidence(name string) ManagerEvidence {
	return ManagerEvidence{
		Name: name, Version: "v1", Revision: "0123456789abcdef0123456789abcdef01234567",
		BinarySHA256: "sha256:" + strings.Repeat("b", 64), Toolchain: "test-toolchain",
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
	}
}

func TestMaterializeCorpusFilesRejectsDuplicateAndExistingTargets(t *testing.T) {
	root := writeCorpus(t, ProtocolRC5, map[string][]byte{"fixtures/a": []byte("alpha")})
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := []CorpusFile{
		{Source: "fixtures/a", Target: "same", Mode: 0o644},
		{Source: "fixtures/a", Target: "same", Mode: 0o644},
	}
	if _, err := MaterializeCorpusFiles(corpus, t.TempDir(), duplicate); err == nil || !strings.Contains(err.Error(), "duplicate fixture target") {
		t.Fatalf("duplicate fixture error = %v", err)
	}
	destination := t.TempDir()
	if err := os.WriteFile(filepath.Join(destination, "existing"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := MaterializeCorpusFiles(corpus, destination, []CorpusFile{{Source: "fixtures/a", Target: "existing", Mode: 0o644}}); err == nil {
		t.Fatal("existing fixture target must be rejected")
	}
	if _, err := MaterializeCorpusFiles(nil, destination, nil); err == nil || !strings.Contains(err.Error(), "corpus is required") {
		t.Fatalf("nil corpus error = %v", err)
	}
	if _, err := MaterializeCorpusFiles(corpus, t.TempDir(), []CorpusFile{{Source: "fixtures/a", Target: "bad-mode", Mode: 0}}); err == nil {
		t.Fatal("zero fixture mode must be rejected")
	}
}

func TestReportJSONRejectsInvalidIdentity(t *testing.T) {
	report := Report{}
	if _, err := ReportJSON(report); err == nil || !strings.Contains(err.Error(), "report schema") {
		t.Fatalf("missing schema error = %v", err)
	}
	report.Schema = ReportSchemaV1
	report.Corpus = CorpusEvidence{Boundary: CorpusBoundaryV1, ProtocolVersion: ProtocolRC5, Revision: strings.Repeat("d", 40), ManifestSHA256: "bad"}
	if _, err := ReportJSON(report); err == nil || !strings.Contains(err.Error(), "manifest digest") {
		t.Fatalf("bad manifest digest error = %v", err)
	}
	report.Corpus.ManifestSHA256 = "sha256:" + strings.Repeat("a", 64)
	if _, err := ReportJSON(report); err == nil || !strings.Contains(err.Error(), "adapter is required") {
		t.Fatalf("missing adapter error = %v", err)
	}
}

func TestRC5CandidateBoundary(t *testing.T) {
	root := os.Getenv("CURATOR_EXTERNAL_INTEROP_ROOT")
	if root == "" {
		t.Skip("CURATOR_EXTERNAL_INTEROP_ROOT is not set")
	}
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(corpus.Entries("bundles/")); got != 5 {
		t.Fatalf("external repository bundle count = %d, want 5", got)
	}
	if _, _, err := corpus.Read("case-manifest.json"); err != nil {
		t.Fatal(err)
	}
	cases, err := corpus.Cases()
	if err != nil {
		t.Fatal(err)
	}
	ids := SortedCaseIDs(cases)
	if len(ids) != ExpectedCaseCount || ids[0] == "" || ids[len(ids)-1] == "" {
		t.Fatalf("sorted case ids = %#v", ids)
	}
}
