package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// The class probes cannot be unit-tested against a fake: what they measure is
// whether this kernel refuses an operation, and a stub that answers "denied"
// would make the harness agree with itself. They are therefore driven end to
// end against the real binary, exactly as an operator drives them, and the
// assertions are about the shape and the internal consistency of the result
// rather than about which verdict this host happens to produce.
//
// A verdict assertion would be wrong here for a second reason: it would turn a
// host that gained or lost a capability into a red test instead of a changed
// observation, which is the opposite of what an evidence harness is for.

var (
	probeBinary   string
	probeBuildErr error
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "probe-harness-")
	if err != nil {
		panic("probe tests: temp dir: " + err.Error())
	}

	binary := filepath.Join(dir, "hardened-probe")
	build := exec.Command("go", "build", "-o", binary, "./cmd/hardened-probe")
	build.Dir = filepath.Join("..", "..")
	var buildOutput strings.Builder
	build.Stdout = &buildOutput
	build.Stderr = &buildOutput
	if err := build.Run(); err != nil {
		probeBuildErr = fmt.Errorf("%w: %s", err, buildOutput.String())
	} else {
		probeBinary = binary
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	if sharedWorkDir != "" {
		_ = os.RemoveAll(sharedWorkDir)
	}
	os.Exit(code)
}

// The probe binary is the subject of every end-to-end test below. If it does
// not exist, none of them measured anything, so its absence is a failure in its
// own right rather than a condition the other tests quietly skip on.
func TestProbeBinaryBuilds(t *testing.T) {
	if probeBuildErr != nil {
		t.Fatalf("the probe binary did not build, so no end-to-end test can measure anything: %v", probeBuildErr)
	}
	if probeBinary == "" {
		t.Fatal("the probe binary was not built and no build error was recorded")
	}
	info, err := os.Stat(probeBinary)
	if err != nil {
		t.Fatalf("stat the built probe binary: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("the built probe binary %s is not executable", probeBinary)
	}
}

// requireAgent returns the built agent, or ends the test.
//
// The two skips below are environment statements: a short run is asking not to
// create probe domains, and a non-darwin host has no seatbelt to measure. A
// missing binary is neither — it means the code under test did not compile, and
// skipping on it would let the whole end-to-end suite report green while
// exercising nothing on the very host it exists to measure.
func requireAgent(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("an end-to-end run creates probe domains and spawns descendants")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("the probe domains are macOS seatbelt domains")
	}
	if probeBinary == "" {
		t.Fatalf("the probe binary is not available, so this end-to-end test measured nothing: %v", probeBuildErr)
	}
	return probeBinary
}

// runHarness performs one full run against the real host.
//
// A run with no injected verdict is measured once and shared. It creates probe
// domains, burns CPU against a real limit and waits out a real deadline, and
// every test that takes no forced class asserts a different property of the
// same measurement — repeating it per test would multiply the suite's wall time
// without observing anything new. The Result is only read here, so sharing it
// changes nothing about what is asserted.
//
// A run with a forced class is never shared: the injection is the subject of
// those tests, and one of them must not be able to see another's injection.
func runHarness(t *testing.T, forced ...string) Result {
	t.Helper()
	if len(forced) == 0 {
		return sharedRun(t)
	}
	result, err := Run(context.Background(), Options{
		WorkDir:          t.TempDir(),
		SelfPath:         requireAgent(t),
		ForceUnavailable: forced,
		Timeout:          4 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return result
}

var (
	sharedRunOnce sync.Once
	sharedResult  Result
	sharedRunErr  error
	sharedWorkDir string
)

func sharedRun(t *testing.T) Result {
	t.Helper()
	// The environment gates come first and per test: a short or non-darwin run
	// must skip before anything is measured.
	agent := requireAgent(t)
	sharedRunOnce.Do(func() {
		sharedWorkDir, sharedRunErr = os.MkdirTemp("", "probe-shared-run-")
		if sharedRunErr != nil {
			return
		}
		sharedResult, sharedRunErr = Run(context.Background(), Options{
			WorkDir:  sharedWorkDir,
			SelfPath: agent,
			Timeout:  4 * time.Minute,
		})
	})
	if sharedRunErr != nil {
		t.Fatalf("the shared measurement run failed: %v", sharedRunErr)
	}
	return sharedResult
}

func TestRunProducesAValidClosedRecord(t *testing.T) {
	result := runHarness(t)

	if diag, reason := result.Evidence.Validate(); diag != "" {
		t.Fatalf("the emitted record fails its own validation: %s: %s", diag, reason)
	}
	// The record has to survive the closed-shape decode too, not only the
	// in-memory check: an extra or missing field would make it unusable to
	// anything that reads the artifact rather than the struct.
	decoded, diag, reason := evidence.Decode(result.Evidence.JSON())
	if diag != "" {
		t.Fatalf("the emitted record does not decode as the closed shape: %s: %s", diag, reason)
	}
	if decoded.Outcome != result.Evidence.Outcome {
		t.Errorf("decoded outcome %q, want %q", decoded.Outcome, result.Evidence.Outcome)
	}

	if result.Evidence.Platform != spec.PlatformMacOS ||
		result.Evidence.EnforcementBackend != spec.BackendMacOSSandbox {
		t.Errorf("record names platform %q backend %q",
			result.Evidence.Platform, result.Evidence.EnforcementBackend)
	}
	// A prototype must never emit a qualified claim, whatever it measured.
	if result.Evidence.QualificationStatus != spec.QualificationUnqualified {
		t.Errorf("qualification %q, want %q",
			result.Evidence.QualificationStatus, spec.QualificationUnqualified)
	}
}

// The exit code is the harness's contract with a caller that reads nothing
// else, so it has to follow the record rather than be set independently.
func TestRunExitCodeFollowsTheOutcome(t *testing.T) {
	result := runHarness(t)

	switch result.Evidence.Outcome {
	case spec.OutcomeEstablished:
		if result.ExitCode != ExitEstablished {
			t.Errorf("established outcome with exit %d, want %d", result.ExitCode, ExitEstablished)
		}
	case spec.OutcomeRejected:
		if result.ExitCode != ExitRejected {
			t.Errorf("rejected outcome with exit %d, want %d", result.ExitCode, ExitRejected)
		}
		if result.Evidence.RejectedBefore == nil || result.Evidence.Diagnostic == nil {
			t.Fatal("a rejection carries no phase or diagnostic")
		}
		if !spec.IsPhase(*result.Evidence.RejectedBefore) {
			t.Errorf("rejected_before %q is not a phase", *result.Evidence.RejectedBefore)
		}
		if !spec.IsDiagnostic(*result.Evidence.Diagnostic) {
			t.Errorf("diagnostic %q is not a stable hardened diagnostic", *result.Evidence.Diagnostic)
		}
	default:
		t.Fatalf("unknown outcome %q", result.Evidence.Outcome)
	}
}

// Every class of the exhaustive inventory must produce a real observation in
// every run. A class that quietly went unprobed would leave the record claiming
// something nobody measured.
func TestRunObservesEveryClassOfTheInventory(t *testing.T) {
	result := runHarness(t)

	byClass := map[string]ClassResult{}
	for _, class := range result.Report.Classes {
		if _, duplicate := byClass[class.Class]; duplicate {
			t.Errorf("class %q was probed twice", class.Class)
		}
		byClass[class.Class] = class
	}
	for _, class := range spec.Classes() {
		got, ok := byClass[class]
		if !ok {
			t.Errorf("class %q produced no result", class)
			continue
		}
		if got.Verdict != VerdictAvailable && got.Verdict != VerdictUnavailable {
			t.Errorf("class %q is %q on a host where a probe domain could be built", class, got.Verdict)
		}
		if len(got.Checks) == 0 {
			t.Errorf("class %q has a verdict but recorded no check", class)
		}
		if got.Mechanism == "" {
			t.Errorf("class %q names no mechanism", class)
		}
		if got.Verdict != VerdictAvailable && len(got.Reasons) == 0 {
			t.Errorf("class %q is %q with no reason recorded", class, got.Verdict)
		}
	}
}

// The acceptance requirement: every capability class carries both a positive
// capability test and an adversarial escape or negative control. A class proved
// only by positives cannot tell enforcement from a broken measurement.
func TestEveryClassCarriesAPositiveAndAControl(t *testing.T) {
	result := runHarness(t)

	for _, class := range result.Report.Classes {
		kinds := map[string]int{}
		for _, check := range class.Checks {
			kinds[check.Kind]++
		}
		if kinds[kindPositive] == 0 {
			t.Errorf("class %q has no positive capability test", class.Class)
		}
		if kinds[kindNegative]+kinds[kindAdversarial] == 0 {
			t.Errorf("class %q has neither a negative control nor an adversarial escape", class.Class)
		}
		for _, check := range class.Checks {
			switch check.Kind {
			case kindPositive, kindNegative, kindAdversarial:
			default:
				t.Errorf("class %q check %q has unknown kind %q", class.Class, check.Name, check.Kind)
			}
			if check.Expectation == "" {
				t.Errorf("class %q check %q states no expectation", class.Class, check.Name)
			}
			if check.Observed == "" {
				t.Errorf("class %q check %q records no observation", class.Class, check.Name)
			}
		}
	}
}

// A class may only be reported applied when it was actually installed in a
// probe domain and observed to hold. Anything else would put a control into the
// record that this run never established.
func TestAppliedImpliesObserved(t *testing.T) {
	result := runHarness(t)

	byClass := map[string]ClassResult{}
	for _, class := range result.Report.Classes {
		byClass[class.Class] = class
	}
	for _, capability := range result.Evidence.Capabilities {
		class := byClass[capability.Name]
		if capability.Status == spec.StatusApplied && class.Verdict != VerdictAvailable {
			t.Errorf("class %q is applied in the record but %q in the report", capability.Name, class.Verdict)
		}
		if capability.Availability == spec.AvailabilityAvailable && capability.Status != spec.StatusApplied {
			t.Errorf("class %q is available but %q", capability.Name, capability.Status)
		}
		if capability.ProbedAt != spec.ProbedAt {
			t.Errorf("class %q was probed at %q, want %q", capability.Name, capability.ProbedAt, spec.ProbedAt)
		}
	}
}

// The fail-closed boundary: forcing any single class unavailable must reject
// before domain entry with the mapped diagnostic, whatever the host measured.
func TestForcingAClassUnavailableRejectsBeforeDomainEntry(t *testing.T) {
	requireAgent(t)

	sweep, err := FailClosedSweep(context.Background(), Options{
		WorkDir:  t.TempDir(),
		SelfPath: probeBinary,
		Timeout:  8 * time.Minute,
	})
	if err != nil {
		t.Fatalf("FailClosedSweep: %v", err)
	}
	if len(sweep) != len(spec.Classes()) {
		t.Fatalf("the sweep covered %d classes, want %d", len(sweep), len(spec.Classes()))
	}
	for _, entry := range sweep {
		if !entry.Pass {
			t.Errorf("forcing %s unavailable did not fail closed: outcome=%s before=%s diagnostic=%s exit=%d",
				entry.ForcedClass, entry.Outcome, entry.RejectedBefore, entry.Diagnostic, entry.ExitCode)
		}
		if entry.RejectedBefore != spec.PhaseCapabilityProbe {
			t.Errorf("forcing %s rejected before %q, want %q",
				entry.ForcedClass, entry.RejectedBefore, spec.PhaseCapabilityProbe)
		}
		if entry.Diagnostic != spec.DiagCapabilityUnavailable {
			t.Errorf("forcing %s produced diagnostic %q, want %q",
				entry.ForcedClass, entry.Diagnostic, spec.DiagCapabilityUnavailable)
		}
	}
}

// An injected verdict must be visible as an injection. A forced value that read
// like a measurement would be the one way this harness could lie.
func TestForcedUnavailableIsMarkedInTheReport(t *testing.T) {
	result := runHarness(t, spec.ClassNetworkDenial)

	var forced ClassResult
	for _, class := range result.Report.Classes {
		if class.Class == spec.ClassNetworkDenial {
			forced = class
		}
	}
	if forced.Verdict != VerdictUnavailable {
		t.Errorf("the forced class is %q, want %q", forced.Verdict, VerdictUnavailable)
	}
	if forced.Applied {
		t.Error("the forced class is reported applied")
	}
	if !strings.Contains(strings.Join(forced.Reasons, " "), "injected, not measured") {
		t.Errorf("the injection is not named in the reasons: %v", forced.Reasons)
	}
	if result.Evidence.Outcome != spec.OutcomeRejected {
		t.Errorf("outcome %q with a class forced unavailable, want %q",
			result.Evidence.Outcome, spec.OutcomeRejected)
	}
}

// The detailed report is the artifact an operator reads when the record says
// only "rejected". It has to be complete and parseable on its own.
func TestReportIsSelfContained(t *testing.T) {
	result := runHarness(t)

	data := result.Report.JSON()
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("the report is not JSON: %v", err)
	}
	for _, field := range []string{
		"report_version", "spec_revision", "platform", "enforcement_backend",
		"host", "classes", "mechanisms", "evidence_record", "exit_code",
	} {
		if _, ok := decoded[field]; !ok {
			t.Errorf("the report has no %q", field)
		}
	}
	if result.Report.Host.ProductVersion == "" {
		t.Error("the report does not say which OS version produced it")
	}
	if result.Report.DurationMS <= 0 {
		t.Error("the report claims a run of no duration")
	}
	if len(result.Report.Mechanisms) == 0 {
		t.Error("the report lists no platform mechanism")
	}
}

// The harness writes both artifacts where the operator asked for them, and the
// evidence file is byte-identical to what the run reported.
func TestWriteArtifactsRoundTrips(t *testing.T) {
	result := runHarness(t)

	dir := t.TempDir()
	evidencePath := filepath.Join(dir, "evidence.json")
	reportPath := filepath.Join(dir, "report.json")
	if err := WriteArtifacts(result, evidencePath, reportPath); err != nil {
		t.Fatalf("WriteArtifacts: %v", err)
	}

	written, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if string(written) != string(result.Evidence.JSON()) {
		t.Error("the written evidence differs from the reported record")
	}
	if _, diag, reason := evidence.Decode(written); diag != "" {
		t.Fatalf("the written evidence is not a valid closed record: %s: %s", diag, reason)
	}
	if info, err := os.Stat(reportPath); err != nil || info.Size() == 0 {
		t.Errorf("the report was not written: %v", err)
	}
}

// The aggregate-bounds class has to carry an executed measurement for every
// bound the guarantee names, not only the two the earlier revision measured. A
// class that concluded about CPU, memory or time without having run anything
// against them would be stating more than it observed.
func TestAggregateBoundsMeasuresEveryDeclaredBound(t *testing.T) {
	result := runHarness(t)

	var bounds ClassResult
	for _, class := range result.Report.Classes {
		if class.Class == spec.ClassAggregateBounds {
			bounds = class
		}
	}
	if bounds.Class == "" {
		t.Fatal("the run produced no aggregate-resource-bounds result")
	}

	names := map[string]Check{}
	for _, check := range bounds.Checks {
		names[check.Name] = check
	}

	// One executed positive per bound the guarantee names.
	for _, required := range []string{
		"descriptors:per-process-bound-binds",
		"disk-bytes:write-past-declared-byte-budget",
		inside.BoundCPU + ":bound-can-be-declared",
		inside.BoundAddressSpace + ":bound-can-be-declared",
		inside.BoundDataSegment + ":bound-can-be-declared",
		inside.BoundProcessCount + ":bound-can-be-declared",
		"wall-clock:deadline-terminates-the-domain-root",
		"supervisor-side-accounting-is-unescapable",
	} {
		check, ok := names[required]
		if !ok {
			t.Errorf("the aggregate-bounds class recorded no %q", required)
			continue
		}
		if check.Observed == "" {
			t.Errorf("%q recorded no observation", required)
		}
		if check.Observed == "probe-failed" {
			t.Errorf("%q could not be measured on this host: %s", required, check.Detail)
		}
	}

	// And a matched control for every bound: a sweep of refusals with nothing
	// to compare against cannot tell enforcement from a broken stress.
	for _, kind := range inside.BoundKinds() {
		control, ok := names[kind+":unbounded-control-exceeds-the-declared-budget"]
		if !ok {
			t.Errorf("%s carries no unbounded control", kind)
			continue
		}
		if !control.Pass {
			t.Errorf("%s: the unbounded control did not pass the declared budget (%s), so no "+
				"refusal measured against it means anything: %s", kind, control.Observed, control.Detail)
		}
	}
}

// The supervisor verdict must agree with the membership and termination this
// same run measured. It used to be a constant, which meant a host that gained
// an unescapable domain would still have been reported as escapable.
func TestSupervisorVerdictAgreesWithTheMeasuredDomain(t *testing.T) {
	result := runHarness(t)

	var verdict Check
	for _, class := range result.Report.Classes {
		for _, check := range class.Checks {
			if check.Name == "supervisor-side-accounting-is-unescapable" {
				verdict = check
			}
		}
	}
	if verdict.Name == "" {
		t.Fatal("the run recorded no supervisor-accounting verdict")
	}
	if verdict.Observed == "probe-failed" {
		t.Fatalf("the supervisor verdict could not be derived: %s", verdict.Detail)
	}

	membership := classVerdict(result, spec.ClassDomainMembership)
	termination := classVerdict(result, spec.ClassDomainTermination)
	want := membership == VerdictAvailable && termination == VerdictAvailable

	if verdict.Pass != want {
		t.Errorf("the supervisor verdict is %v while membership is %q and termination is %q; "+
			"supervisor-side accounting can only be unescapable when both hold",
			verdict.Pass, membership, termination)
	}
}

func classVerdict(result Result, class string) Verdict {
	for _, got := range result.Report.Classes {
		if got.Class == class {
			return got.Verdict
		}
	}
	return VerdictInconclusive
}

// A mechanism the run did not exercise must say so. Otherwise a status written
// by hand reads exactly like one the probes established.
func TestMechanismsSayWhetherTheyWereMeasured(t *testing.T) {
	result := runHarness(t)

	measured := 0
	for _, mechanism := range result.Report.Mechanisms {
		if mechanism.Observation == "" {
			t.Errorf("mechanism %q carries no observation either way", mechanism.Name)
		}
		if mechanism.Exercised {
			measured++
		}
	}
	if measured == 0 {
		t.Error("the run exercised no mechanism at all")
	}
	// The boundary of the evidence has to be visible, not implied.
	if len(UnexercisedMechanisms(result.Report.Mechanisms)) == 0 {
		t.Error("every mechanism claims to have been exercised, which the inventory is wider than")
	}
}

// The harness must leave nothing running on the host it measured, whatever the
// platform failed to tear down. This is checked through the report rather than
// by scanning the process table, so it names the exact descendants the probe
// created.
func TestRunLeavesNoDescendantBehind(t *testing.T) {
	result := runHarness(t)

	for _, class := range result.Report.Classes {
		for _, check := range class.Checks {
			if check.Name != "wall-clock:harness-leaves-no-descendant-behind" {
				continue
			}
			if !check.Pass {
				t.Errorf("the run left descendants behind: %s (%s)", check.Observed, check.Detail)
			}
			return
		}
	}
	t.Error("the run recorded no statement about its own descendants")
}

// A cancelled context must not turn into a clean "the host cannot do this".
// The run either reports a rejection or an error, never an established one.
func TestRunNeverEstablishesUnderACancelledContext(t *testing.T) {
	requireAgent(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := Run(ctx, Options{
		WorkDir:  t.TempDir(),
		SelfPath: probeBinary,
		Timeout:  time.Minute,
	})
	if err == nil && result.Evidence.Outcome == spec.OutcomeEstablished {
		t.Error("a cancelled run reported every guarantee established")
	}
	if result.ExitCode == ExitEstablished && err == nil {
		t.Error("a cancelled run exited as if everything had been established")
	}
}
