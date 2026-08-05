package crossmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func writeSuiteCorpus(t *testing.T) string {
	t.Helper()
	cases := make([]Case, 0, ExpectedCaseCount)
	for index := 0; index < ExpectedCaseCount; index++ {
		expected := json.RawMessage(`{"outcome":"accepted","mutation":false,"unexpected_children":false,"network_started":false}`)
		cases = append(cases, Case{ID: fmt.Sprintf("case-%02d", index), Category: "test", Source: "vectors/source.json", Expected: expected})
	}
	manifest := map[string]any{
		"schema_version": 1, "corpus_version": CorpusRC5, "protocol_version": ProtocolRC5,
		"implementation_neutral": true, "manager_adapter": nil, "physical_paths": "implementation-specific",
		"architecture_v6_coverage": map[string]any{}, "architecture_v6_threat_matrix": []any{},
		"lifecycle_boundaries": []any{}, "lifecycle_matrix": map[string]any{}, "cases": cases,
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return writeCorpus(t, ProtocolRC5, map[string][]byte{"case-manifest.json": payload, "vectors/source.json": []byte(`{"source":true}`)})
}

func TestSuiteRunsIdenticalCasesAsBlackBoxProcesses(t *testing.T) {
	corpus, err := OpenCorpus(writeSuiteCorpus(t), RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	suite := Suite{Corpus: corpus, SpecRevision: strings.Repeat("d", 40), WorkRoot: filepath.Join(t.TempDir(), "work"), ArtifactRoot: filepath.Join(t.TempDir(), "artifacts"), Probe: staticProbe{}}
	var reports []Report
	for _, name := range []string{"curator", "csk"} {
		adapter := passingAdapter(name)
		report, err := suite.RunAdapter(context.Background(), adapter)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(report.Cases) != ExpectedCaseCount {
			t.Fatalf("%s cases = %d", name, len(report.Cases))
		}
		for _, result := range report.Cases {
			if result.Status != CaseObserved || result.ExitCode != 0 || len(result.Violations) != 0 {
				t.Fatalf("%s %s = %#v", name, result.ID, result)
			}
			if _, err := os.Stat(filepath.Join(result.ArtifactDir, "observation.json")); err != nil {
				t.Fatalf("%s artifact: %v", result.ID, err)
			}
		}
		if _, err := ReportJSON(report); err != nil {
			t.Fatalf("%s report: %v", name, err)
		}
		reports = append(reports, report)
	}
	comparisons, err := CompareReports(reports[0], reports[1])
	if err != nil {
		t.Fatal(err)
	}
	if len(comparisons) != ExpectedCaseCount || !comparisons[0].Equal || comparisons[0].LeftSHA256 != comparisons[0].RightSHA256 {
		t.Fatalf("cross-manager comparison = %#v", comparisons[0])
	}
	reports[1].Cases[0].Observed = json.RawMessage(`{"outcome":"different"}`)
	comparisons, err = CompareReports(reports[0], reports[1])
	if err != nil || comparisons[0].Equal || !strings.Contains(strings.Join(comparisons[0].Mismatches, ","), "observed") {
		t.Fatalf("cross-manager mismatch = %#v, %v", comparisons[0], err)
	}
}

func passingAdapter(name string) Adapter {
	return Adapter{
		Evidence: testManagerEvidence(name),
		Prepare: func(_ context.Context, item Case, workspace Workspace) (Invocation, error) {
			return Invocation{
				Executable: os.Args[0], Args: []string{"-test.run=TestSuiteProcessHelper", "--"}, Dir: workspace.Repo,
				Env: []string{"GO_WANT_SUITE_PROCESS_HELPER=1", "CASE_ID=" + item.ID},
			}, nil
		},
		Normalize: func(_ Case, result RunResult) (map[string]any, error) {
			var observed map[string]any
			err := json.Unmarshal(result.Stdout, &observed)
			return observed, err
		},
	}
}

func TestSuiteDetectsUnexpectedProcessNetworkAndFailedMutation(t *testing.T) {
	corpus, err := OpenCorpus(writeSuiteCorpus(t), RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	adapter := passingAdapter("curator")
	adapter.Prepare = func(_ context.Context, item Case, workspace Workspace) (Invocation, error) {
		return Invocation{
			Executable: os.Args[0], Args: []string{"-test.run=TestSuiteProcessHelper", "--"}, Dir: workspace.Repo,
			Env: []string{"GO_WANT_SUITE_PROCESS_HELPER=1", "CASE_ID=" + item.ID, "WRITE_PATH=" + filepath.Join(workspace.Repo, "unexpected")},
		}, nil
	}
	suite := Suite{
		Corpus: corpus, SpecRevision: strings.Repeat("d", 40), WorkRoot: filepath.Join(t.TempDir(), "work"), ArtifactRoot: filepath.Join(t.TempDir(), "artifacts"),
		Probe: staticProbe{events: BoundaryEvents{UnexpectedProcesses: []string{"shell"}, NetworkConnections: []string{"203.0.113.1:443"}}},
	}
	report, err := suite.RunAdapter(context.Background(), adapter)
	if err != nil {
		t.Fatal(err)
	}
	first := report.Cases[0]
	if first.Status != CaseMismatch || !strings.Contains(first.Detail, "mutation=false") || !strings.Contains(first.Detail, "unexpected child process") || !strings.Contains(first.Detail, "unexpected network connection") {
		t.Fatalf("failed mutation was not detected: %#v", first)
	}
	if len(first.Boundary.UnexpectedProcesses) != 1 || len(first.Boundary.NetworkConnections) != 1 {
		t.Fatalf("boundary evidence was not preserved: %#v", first.Boundary)
	}
}

type staticProbe struct{ events BoundaryEvents }

func (staticProbe) Begin(context.Context, Invocation) (any, error) { return nil, nil }
func (probe staticProbe) End(context.Context, any) (BoundaryEvents, error) {
	return probe.events, nil
}

func TestSuiteProcessHelper(_ *testing.T) {
	if os.Getenv("GO_WANT_SUITE_PROCESS_HELPER") != "1" {
		return
	}
	if target := os.Getenv("WRITE_PATH"); target != "" {
		if err := os.WriteFile(target, []byte("mutation"), 0o600); err != nil {
			os.Exit(92)
		}
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"outcome":"accepted","mutation":false,"unexpected_children":false,"network_started":false,"case":%q}`+"\n", os.Getenv("CASE_ID")); err != nil {
		os.Exit(93)
	}
	os.Exit(0)
}

func TestManagerEvidenceAndBinaryDigestAreExact(t *testing.T) {
	digest, err := BinarySHA256(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
		t.Fatalf("digest = %q", digest)
	}
	evidence := testManagerEvidence("curator")
	evidence.OS, evidence.Architecture = runtime.GOOS, runtime.GOARCH
	if err := validateManagerEvidence(evidence); err != nil {
		t.Fatal(err)
	}
}

func TestNativeBoundaryProbeIsAvailableOnRequiredHosts(t *testing.T) {
	probe, err := NativeBoundaryProbe()
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		if err != nil || probe == nil {
			t.Fatalf("native probe: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatal("unsupported host must fail closed")
	}
}

func TestCommandBoundaryProbePollsAndReturnsDeterministicDeltas(t *testing.T) {
	processState := filepath.Join(t.TempDir(), "process-state")
	networkState := filepath.Join(t.TempDir(), "network-state")
	probe := CommandBoundaryProbe{
		Processes: statefulSampleCommand(processState),
		Network:   statefulSampleCommand(networkState),
		Ignore:    []string{"ignored-row"}, Interval: time.Millisecond,
	}
	token, err := probe.Begin(context.Background(), Invocation{})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(25 * time.Millisecond)
	events, err := probe.End(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if len(events.UnexpectedProcesses) != 1 || events.UnexpectedProcesses[0] != "new-row" {
		t.Fatalf("process delta = %#v", events.UnexpectedProcesses)
	}
	if len(events.NetworkConnections) != 1 || events.NetworkConnections[0] != "new-row" {
		t.Fatalf("network delta = %#v", events.NetworkConnections)
	}
	if _, err := probe.End(context.Background(), struct{}{}); err == nil {
		t.Fatal("invalid probe token must fail")
	}
	if _, err := sampleLines(context.Background(), SampleCommand{}); err == nil {
		t.Fatal("empty sample command must fail")
	}
}

func statefulSampleCommand(statePath string) SampleCommand {
	if runtime.GOOS == "windows" {
		return SampleCommand{Executable: "powershell.exe", Args: []string{"-NoProfile", "-NonInteractive", "-Command", `$p=$args[0]; $n=0; if (Test-Path $p) {$n=[int](Get-Content $p)}; $n++; Set-Content -NoNewline $p $n; 'base-row'; if ($n -gt 1) {'new-row'; 'ignored-row'}`, statePath}}
	}
	return SampleCommand{Executable: "sh", Args: []string{"-c", `n=$(cat "$1" 2>/dev/null || echo 0); n=$((n+1)); printf %s "$n" > "$1"; echo base-row; if [ "$n" -gt 1 ]; then echo new-row; echo ignored-row; fi`, "sh", statePath}}
}

func TestSuiteRejectsIncompleteExecutionContracts(t *testing.T) {
	root := writeSuiteCorpus(t)
	corpus, err := OpenCorpus(root, RC5Boundary)
	if err != nil {
		t.Fatal(err)
	}
	valid := passingAdapter("curator")
	for _, testCase := range []struct {
		name  string
		suite Suite
		adapt Adapter
	}{
		{"nil corpus", Suite{Probe: staticProbe{}}, valid},
		{"incomplete adapter", Suite{Corpus: corpus, Probe: staticProbe{}}, Adapter{}},
		{"missing probe", Suite{Corpus: corpus}, valid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := testCase.suite.RunAdapter(context.Background(), testCase.adapt); err == nil {
				t.Fatal("invalid execution contract must fail")
			}
		})
	}
	badEvidence := valid
	badEvidence.Evidence.Revision = ""
	if _, err := (Suite{Corpus: corpus, SpecRevision: "spec", Probe: staticProbe{}}).RunAdapter(context.Background(), badEvidence); err == nil {
		t.Fatal("incomplete evidence must fail")
	}
}

func TestComparisonHelpersCoverNestedAndEscapedWrites(t *testing.T) {
	violations := compareExpected(
		map[string]any{"nested": map[string]any{"value": "want"}, "missing": true},
		map[string]any{"nested": map[string]any{"value": "got"}}, "",
	)
	if strings.Join(violations, ",") != "missing missing,nested.value differs" {
		t.Fatalf("violations = %#v", violations)
	}
	root := t.TempDir()
	inside := filepath.Join(root, "inside")
	outside := filepath.Join(t.TempDir(), "outside")
	if got := outsideWrites([]string{inside, outside}, []string{root}); len(got) != 1 || got[0] != outside {
		t.Fatalf("outside writes = %#v", got)
	}
	if err := validateSegment("bad/name"); err == nil {
		t.Fatal("unsafe segment must fail")
	}
}

func TestCompareReportsRejectsAmbiguousInputs(t *testing.T) {
	baseCase := CaseResult{ID: "a", Status: CaseObserved, Expected: json.RawMessage(`{"x":1}`), Observed: json.RawMessage(`{"x":1}`)}
	left := Report{Manager: testManagerEvidence("curator"), Corpus: CorpusEvidence{Boundary: CorpusBoundaryV1, ProtocolVersion: ProtocolRC5, Revision: "spec", ManifestSHA256: "sha256:" + strings.Repeat("a", 64)}, Cases: []CaseResult{baseCase}}
	right := left
	if _, err := CompareReports(left, right); err == nil {
		t.Fatal("same-manager comparison must fail")
	}
	right.Manager = testManagerEvidence("csk")
	right.Corpus.ManifestSHA256 = "sha256:" + strings.Repeat("b", 64)
	if _, err := CompareReports(left, right); err == nil {
		t.Fatal("different corpus comparison must fail")
	}
	right.Corpus = left.Corpus
	right.Cases = nil
	if _, err := CompareReports(left, right); err == nil {
		t.Fatal("different case count must fail")
	}
	right.Cases = []CaseResult{baseCase, baseCase}
	if _, err := CompareReports(left, right); err == nil {
		t.Fatal("duplicate case must fail")
	}
	right.Cases = []CaseResult{{ID: "b", Status: CaseObserved}}
	if _, err := CompareReports(left, right); err == nil {
		t.Fatal("missing case must fail")
	}
	if jsonEqual(json.RawMessage(`not-json`), json.RawMessage(`not-json`)) != true {
		t.Fatal("invalid but byte-equal JSON must compare deterministically")
	}
}
