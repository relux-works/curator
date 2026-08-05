package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/probe"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// TestMain doubles as the in-domain agent dispatcher.
//
// The probe binary is its own agent: the exact-executable allowlist has to have
// exactly one entry, so the harness re-invokes the file it is running from. The
// test binary is the file os.Executable() resolves to here, so it has to answer
// the same "__inside" invocation or every end-to-end path would fail for a
// reason that has nothing to do with the host.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "__inside" {
		os.Exit(inside.Main(os.Args[2:]))
	}
	os.Exit(m.Run())
}

// captureOutput runs fn with stdout and stderr redirected to pipes and returns
// what each received. The command writes the record to the file descriptor, so
// swapping the package-level variables is the only way to read it back.
func captureOutput(t *testing.T, fn func() int) (code int, stdout, stderr string) {
	t.Helper()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	originalOut, originalErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW

	outDone := make(chan string, 1)
	errDone := make(chan string, 1)
	go func() { outDone <- readAll(outR) }()
	go func() { errDone <- readAll(errR) }()

	code = fn()

	os.Stdout, os.Stderr = originalOut, originalErr
	_ = outW.Close()
	_ = errW.Close()
	stdout, stderr = <-outDone, <-errDone
	_ = outR.Close()
	_ = errR.Close()
	return code, stdout, stderr
}

func readAll(f *os.File) string {
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			return b.String()
		}
	}
}

func requireDarwin(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("an end-to-end run creates probe domains and spawns descendants")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("the probe domains are macOS seatbelt domains")
	}
}

// --------------------------------------------------------------- flag surface

func TestListClassesPrintsTheExhaustiveInventory(t *testing.T) {
	code, stdout, _ := captureOutput(t, func() int { return run([]string{"--list-classes"}) })

	if code != probe.ExitEstablished {
		t.Errorf("exit %d, want %d", code, probe.ExitEstablished)
	}
	printed := strings.Fields(stdout)
	if len(printed) != len(spec.Classes()) {
		t.Fatalf("printed %d classes, want %d: %q", len(printed), len(spec.Classes()), stdout)
	}
	for i, class := range spec.Classes() {
		if printed[i] != class {
			t.Errorf("class %d is %q, want %q", i, printed[i], class)
		}
	}
}

// An unparseable invocation is a harness fault, not a statement about the host.
// Returning the rejection code here would tell a caller the host failed a probe
// that never ran.
func TestUnusableInvocationsExitAsHarnessErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown flag", []string{"--no-such-flag"}},
		{"not a class", []string{"--force-unavailable=not-a-capability-class"}},
		{"class list with a typo", []string{"--force-unavailable=network-syscall-denial,typo"}},
		{"bad expectation", []string{"--expect=maybe", "--quiet"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "bad expectation" {
				requireDarwin(t)
			}
			code, _, stderr := captureOutput(t, func() int { return run(tc.args) })
			if code != probe.ExitHarnessError {
				t.Errorf("exit %d, want %d (stderr %q)", code, probe.ExitHarnessError, stderr)
			}
		})
	}
}

func TestSplitClasses(t *testing.T) {
	cases := []struct {
		raw  string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{" a , b ,, ", []string{"a", "b"}},
	}
	for _, tc := range cases {
		got := splitClasses(tc.raw)
		if len(got) != len(tc.want) {
			t.Errorf("splitClasses(%q) = %v, want %v", tc.raw, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitClasses(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		}
	}
}

func TestDerefOr(t *testing.T) {
	value := "capability-probe"
	if got := derefOr(&value); got != value {
		t.Errorf("derefOr = %q, want %q", got, value)
	}
	if got := derefOr(nil); got != "null" {
		t.Errorf("derefOr(nil) = %q, want \"null\"", got)
	}
}

// ------------------------------------------------------------- end-to-end

// The default invocation writes the closed record to stdout and the operator
// summary to stderr. Mixing them would make the machine-readable result
// unparseable by the caller that reads stdout.
func TestDefaultRunSeparatesTheRecordFromTheSummary(t *testing.T) {
	requireDarwin(t)
	dir := t.TempDir()

	code, stdout, stderr := captureOutput(t, func() int {
		return run([]string{"--work-dir=" + dir})
	})

	if code != probe.ExitEstablished && code != probe.ExitRejected {
		t.Fatalf("exit %d, want an outcome code (stderr %q)", code, stderr)
	}
	record, diag, reason := evidence.Decode([]byte(stdout))
	if diag != "" {
		t.Fatalf("stdout is not a valid closed record: %s: %s (%q)", diag, reason, stdout)
	}
	if record.QualificationStatus != spec.QualificationUnqualified {
		t.Errorf("qualification %q, want %q", record.QualificationStatus, spec.QualificationUnqualified)
	}
	if (code == probe.ExitEstablished) != (record.Outcome == spec.OutcomeEstablished) {
		t.Errorf("exit %d disagrees with outcome %q", code, record.Outcome)
	}
	// The summary must never read as an enforcement claim, whatever it found.
	if !strings.Contains(stderr, "not an enforcement claim") {
		t.Errorf("the summary omits the non-claim: %q", stderr)
	}
	if !strings.Contains(stderr, spec.PlatformMacOS) {
		t.Errorf("the summary does not name the platform: %q", stderr)
	}
}

func TestQuietWritesArtifactsAndNothingElse(t *testing.T) {
	requireDarwin(t)
	dir := t.TempDir()
	evidencePath := filepath.Join(dir, "evidence.json")
	reportPath := filepath.Join(dir, "report.json")

	code, stdout, stderr := captureOutput(t, func() int {
		return run([]string{
			"--work-dir=" + filepath.Join(dir, "work"),
			"--evidence=" + evidencePath,
			"--report=" + reportPath,
			"--quiet",
		})
	})

	if code != probe.ExitEstablished && code != probe.ExitRejected {
		t.Fatalf("exit %d (stderr %q)", code, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("--quiet still wrote stdout %q stderr %q", stdout, stderr)
	}

	written, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if _, diag, reason := evidence.Decode(written); diag != "" {
		t.Fatalf("the written evidence is not a valid closed record: %s: %s", diag, reason)
	}
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report map[string]any
	if err := json.Unmarshal(reportData, &report); err != nil {
		t.Fatalf("the report is not JSON: %v", err)
	}
	if report["report_version"] != probe.ReportVersion {
		t.Errorf("report_version %v, want %q", report["report_version"], probe.ReportVersion)
	}
}

// --expect turns the exit status into the answer to an assertion, which is what
// makes the harness usable from a script that must fail when the host changes.
func TestExpectAssertsTheOutcome(t *testing.T) {
	requireDarwin(t)

	// Forcing a class unavailable makes the outcome rejected by construction,
	// so the assertion has a known answer on any host.
	forced := "--force-unavailable=" + spec.ClassNetworkDenial

	code, _, stderr := captureOutput(t, func() int {
		return run([]string{"--work-dir=" + t.TempDir(), forced, "--expect=rejected", "--quiet"})
	})
	if code != probe.ExitEstablished {
		t.Errorf("a satisfied assertion exited %d, want %d (stderr %q)", code, probe.ExitEstablished, stderr)
	}

	code, _, stderr = captureOutput(t, func() int {
		return run([]string{"--work-dir=" + t.TempDir(), forced, "--expect=established", "--quiet"})
	})
	if code != probe.ExitHarnessError {
		t.Errorf("a violated assertion exited %d, want %d", code, probe.ExitHarnessError)
	}
	if !strings.Contains(stderr, "expected outcome") {
		t.Errorf("the violated assertion is not explained: %q", stderr)
	}
}

// The fail-closed sweep is the executable form of the preflight evidence: every
// class forced unavailable in turn must reject before domain entry.
func TestFailClosedSweepReportsEveryClass(t *testing.T) {
	requireDarwin(t)
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")

	code, _, stderr := captureOutput(t, func() int {
		return run([]string{
			"--work-dir=" + filepath.Join(dir, "work"),
			"--report=" + reportPath,
			"--fail-closed-sweep",
			"--quiet",
		})
	})
	if code == probe.ExitHarnessError {
		t.Fatalf("the sweep reported a harness error: %q", stderr)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report probe.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("the report is not JSON: %v", err)
	}
	if len(report.FailClosed) != len(spec.Classes()) {
		t.Fatalf("the sweep covered %d classes, want %d", len(report.FailClosed), len(spec.Classes()))
	}
	for _, entry := range report.FailClosed {
		if !entry.Pass {
			t.Errorf("forcing %s unavailable did not fail closed: %+v", entry.ForcedClass, entry)
		}
		if entry.ExitCode != probe.ExitRejected {
			t.Errorf("forcing %s exited %d, want %d", entry.ForcedClass, entry.ExitCode, probe.ExitRejected)
		}
	}
}

// A work directory that cannot be created is a harness fault. Reporting it as a
// host rejection would blame the platform for a broken invocation.
func TestUnusableWorkDirectoryIsAHarnessError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	code, _, _ := captureOutput(t, func() int {
		return run([]string{"--work-dir=" + filepath.Join(file, "under-a-file")})
	})
	if code != probe.ExitHarnessError {
		t.Errorf("exit %d, want %d", code, probe.ExitHarnessError)
	}
}

// ------------------------------------------------------------------ Summary

func TestSummaryNamesEveryClassAndGuarantee(t *testing.T) {
	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox,
		spec.QualificationUnqualified, nil)
	result := probe.Result{
		Evidence: record,
		ExitCode: probe.ExitRejected,
		Report: probe.Report{
			Backend: spec.BackendMacOSSandbox,
			Host:    probe.HostInfo{ProductName: "macOS", ProductVersion: "26.5", Arch: "arm64"},
			Classes: []probe.ClassResult{{
				Class:   spec.ClassNetworkDenial,
				Verdict: probe.VerdictUnavailable,
				Reasons: []string{"a reason the operator needs to see"},
			}},
			FailClosed: []probe.FailClosed{{
				ForcedClass:    spec.ClassNetworkDenial,
				Outcome:        spec.OutcomeRejected,
				RejectedBefore: spec.PhaseCapabilityProbe,
				Diagnostic:     spec.DiagCapabilityUnavailable,
				ExitCode:       probe.ExitRejected,
				Pass:           true,
			}},
		},
	}

	text := Summary(result)
	for _, class := range spec.Classes() {
		if !strings.Contains(text, class) {
			t.Errorf("the summary omits class %q", class)
		}
	}
	for _, guarantee := range spec.Guarantees() {
		if !strings.Contains(text, guarantee) {
			t.Errorf("the summary omits guarantee %q", guarantee)
		}
	}
	for _, want := range []string{
		"macOS", "26.5", "arm64",
		spec.QualificationUnqualified,
		spec.OutcomeRejected,
		spec.DiagCapabilityUnavailable,
		"a reason the operator needs to see",
		"fail-closed sweep",
		"not an enforcement claim",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the summary omits %q", want)
		}
	}
}

func TestSummaryOfAnEstablishedRunCarriesNoDiagnostic(t *testing.T) {
	observations := map[string]evidence.Observation{}
	for _, class := range spec.Classes() {
		observations[class] = evidence.Observation{
			Availability: spec.AvailabilityAvailable,
			Applied:      true,
		}
	}
	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox,
		spec.QualificationUnqualified, observations)
	if record.Outcome != spec.OutcomeEstablished {
		t.Fatalf("the fixture did not build an established record: %q", record.Outcome)
	}

	text := Summary(probe.Result{Evidence: record, ExitCode: probe.ExitEstablished})
	if strings.Contains(text, "rejected before") {
		t.Errorf("an established summary names a rejection phase: %q", text)
	}
	if !strings.Contains(text, "not an enforcement claim") {
		t.Error("even an established summary must state that this is not an enforcement claim")
	}
	if strings.Contains(text, "not established") {
		t.Errorf("an established summary says a guarantee is not established: %q", text)
	}
}
