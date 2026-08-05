package probe

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// The reduction from observations to a verdict is where a probe could quietly
// turn "I could not tell" into "the host is fine". Every one of these tests is
// about that direction being closed.

// ------------------------------------------------------------------ reduce

func TestReduceRequiresEveryCheckToPass(t *testing.T) {
	cases := []struct {
		name        string
		checks      []Check
		wantVerdict Verdict
		wantReasons int
	}{
		{
			name:        "no checks at all",
			checks:      nil,
			wantVerdict: VerdictAvailable,
		},
		{
			name: "every check passes",
			checks: []Check{
				{Name: "a", Kind: kindPositive, Observed: inside.OutcomeDenied, Pass: true},
				{Name: "b", Kind: kindNegative, Observed: inside.OutcomeAllowed, Pass: true},
			},
			wantVerdict: VerdictAvailable,
		},
		{
			name: "one observed failure",
			checks: []Check{
				{Name: "a", Kind: kindPositive, Observed: inside.OutcomeDenied, Pass: true},
				{Name: "escape", Kind: kindAdversarial, Observed: inside.OutcomeAllowed, Pass: false},
			},
			wantVerdict: VerdictUnavailable,
			wantReasons: 1,
		},
		{
			name: "a check that could not be evaluated",
			checks: []Check{
				{Name: "a", Kind: kindPositive, Observed: inside.OutcomeInconclusive, Pass: false},
			},
			wantVerdict: VerdictInconclusive,
			wantReasons: 1,
		},
		{
			name: "a check that never ran",
			checks: []Check{
				{Name: "a", Kind: kindPositive, Observed: "probe-failed", Pass: false},
			},
			wantVerdict: VerdictInconclusive,
			wantReasons: 1,
		},
		{
			// A run that both failed and could not be evaluated is not "mostly
			// fine": the unevaluated part dominates, because nobody knows what it
			// would have said.
			name: "a real failure alongside an unevaluated check",
			checks: []Check{
				{Name: "escape", Kind: kindAdversarial, Observed: inside.OutcomeAllowed, Pass: false},
				{Name: "b", Kind: kindPositive, Observed: "probe-failed", Pass: false},
			},
			wantVerdict: VerdictInconclusive,
			wantReasons: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict, reasons := reduce(tc.checks)
			if verdict != tc.wantVerdict {
				t.Errorf("verdict %q, want %q (reasons %v)", verdict, tc.wantVerdict, reasons)
			}
			if len(reasons) != tc.wantReasons {
				t.Errorf("%d reasons, want %d: %v", len(reasons), tc.wantReasons, reasons)
			}
		})
	}
}

func TestReduceReasonsNameTheCheck(t *testing.T) {
	_, reasons := reduce([]Check{
		{Name: "write-tmp", Kind: kindAdversarial, Observed: inside.OutcomeAllowed, Pass: false},
	})
	if len(reasons) != 1 {
		t.Fatalf("reasons = %v", reasons)
	}
	for _, want := range []string{"write-tmp", kindAdversarial, inside.OutcomeAllowed} {
		if !strings.Contains(reasons[0], want) {
			t.Errorf("reason %q does not mention %q", reasons[0], want)
		}
	}
}

// ------------------------------------------------------------- observation

func TestObservationScoresAgainstTheExpectedOutcome(t *testing.T) {
	cases := []struct {
		name     string
		attempt  inside.Attempt
		found    bool
		want     string
		wantPass bool
		wantObs  string
	}{
		{
			name:    "denied where a denial was required",
			attempt: inside.Attempt{Outcome: inside.OutcomeDenied, Errno: "EPERM"},
			found:   true, want: inside.OutcomeDenied, wantPass: true, wantObs: inside.OutcomeDenied,
		},
		{
			name:    "allowed where a denial was required",
			attempt: inside.Attempt{Outcome: inside.OutcomeAllowed},
			found:   true, want: inside.OutcomeDenied, wantPass: false, wantObs: inside.OutcomeAllowed,
		},
		{
			// The one that matters: an operation that failed because the target
			// was missing has not been refused by anything.
			name:    "inconclusive is not a denial",
			attempt: inside.Attempt{Outcome: inside.OutcomeInconclusive, Errno: "ENOENT"},
			found:   true, want: inside.OutcomeDenied, wantPass: false, wantObs: inside.OutcomeInconclusive,
		},
		{
			name:    "allowed where success was required",
			attempt: inside.Attempt{Outcome: inside.OutcomeAllowed},
			found:   true, want: inside.OutcomeAllowed, wantPass: true, wantObs: inside.OutcomeAllowed,
		},
		{
			name:    "the agent never reported it",
			attempt: inside.Attempt{},
			found:   false, want: inside.OutcomeDenied, wantPass: false, wantObs: "probe-failed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := observation("probe", kindPositive, "expectation", tc.attempt, tc.found, tc.want)
			if check.Pass != tc.wantPass {
				t.Errorf("pass = %v, want %v", check.Pass, tc.wantPass)
			}
			if check.Observed != tc.wantObs {
				t.Errorf("observed = %q, want %q", check.Observed, tc.wantObs)
			}
			if check.Name != "probe" || check.Kind != kindPositive || check.Expectation != "expectation" {
				t.Errorf("check lost its identity: %+v", check)
			}
		})
	}
}

// The errno is what tells a reader whether a denial was EPERM from the profile
// or something else, so it must survive into the detail either way.
func TestObservationKeepsTheErrno(t *testing.T) {
	withDetail := observation("probe", kindPositive, "e",
		inside.Attempt{Outcome: inside.OutcomeDenied, Errno: "EPERM", Detail: "open /etc/passwd"}, true, inside.OutcomeDenied)
	if !strings.HasPrefix(withDetail.Detail, "EPERM: ") || !strings.Contains(withDetail.Detail, "/etc/passwd") {
		t.Errorf("detail %q, want the errno prefixed to the detail", withDetail.Detail)
	}

	bare := observation("probe", kindPositive, "e",
		inside.Attempt{Outcome: inside.OutcomeDenied, Errno: "EACCES"}, true, inside.OutcomeDenied)
	if bare.Detail != "EACCES" {
		t.Errorf("detail %q, want just the errno", bare.Detail)
	}

	none := observation("probe", kindPositive, "e",
		inside.Attempt{Outcome: inside.OutcomeAllowed}, true, inside.OutcomeAllowed)
	if none.Detail != "" {
		t.Errorf("detail %q, want empty", none.Detail)
	}
}

func TestFailedCheckNeverPasses(t *testing.T) {
	check := failedCheck("probe", kindPositive, "expectation", "the domain would not start")
	if check.Pass {
		t.Error("a check that could not run reported a pass")
	}
	if check.Observed != "probe-failed" {
		t.Errorf("observed %q, want probe-failed", check.Observed)
	}
	if verdict, _ := reduce([]Check{check}); verdict != VerdictInconclusive {
		t.Errorf("a failed check reduced to %q, want inconclusive", verdict)
	}
}

// --------------------------------------------------------------------- timed

func TestTimedMarksAppliedOnlyWhenAvailable(t *testing.T) {
	cases := []struct {
		name        string
		checks      []Check
		applied     bool
		wantVerdict Verdict
		wantApplied bool
	}{
		{
			name:        "available and installed",
			checks:      []Check{{Name: "a", Pass: true}},
			applied:     true,
			wantVerdict: VerdictAvailable, wantApplied: true,
		},
		{
			// The control was installed but the class did not hold. Reporting it
			// applied would put "available" in the evidence record.
			name:        "installed but the class failed",
			checks:      []Check{{Name: "a", Observed: inside.OutcomeAllowed}},
			applied:     true,
			wantVerdict: VerdictUnavailable, wantApplied: false,
		},
		{
			name:        "never installed",
			checks:      []Check{{Name: "a", Pass: true}},
			applied:     false,
			wantVerdict: VerdictAvailable, wantApplied: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := timed("a-class", "a-mechanism", func() ([]Check, bool) {
				return tc.checks, tc.applied
			})
			if result.Class != "a-class" || result.Mechanism != "a-mechanism" {
				t.Errorf("result lost its identity: %+v", result)
			}
			if result.Verdict != tc.wantVerdict {
				t.Errorf("verdict %q, want %q", result.Verdict, tc.wantVerdict)
			}
			if result.Applied != tc.wantApplied {
				t.Errorf("applied = %v, want %v", result.Applied, tc.wantApplied)
			}
			if result.DurationMS < 0 {
				t.Errorf("duration %d ms", result.DurationMS)
			}
		})
	}
}

// -------------------------------------------------------------- Observations

func TestObservationsCollapseInconclusiveToUnavailable(t *testing.T) {
	results := []ClassResult{
		{Class: spec.ClassNetworkDenial, Verdict: VerdictAvailable, Applied: true},
		{Class: spec.ClassExecAllowlist, Verdict: VerdictUnavailable},
		{Class: spec.ClassViewRestriction, Verdict: VerdictInconclusive},
	}
	got := Observations(results)

	if o := got[spec.ClassNetworkDenial]; o.Availability != spec.AvailabilityAvailable || !o.Applied {
		t.Errorf("available class reduced to %+v", o)
	}
	if o := got[spec.ClassExecAllowlist]; o.Availability != spec.AvailabilityUnavailable || o.Applied {
		t.Errorf("unavailable class reduced to %+v", o)
	}
	// "The probe did not work" must never become "the host provides it".
	if o := got[spec.ClassViewRestriction]; o.Availability != spec.AvailabilityUnavailable || o.Applied {
		t.Errorf("inconclusive class reduced to %+v, want unavailable", o)
	}
	if len(got) != len(results) {
		t.Errorf("Observations returned %d entries for %d results", len(got), len(results))
	}
}

// A class result that never made it into the list must not appear in the
// observations at all, so evidence.Build reports it unprobed.
func TestObservationsOmitsClassesThatWereNotProbed(t *testing.T) {
	got := Observations([]ClassResult{{Class: spec.ClassNetworkDenial, Verdict: VerdictAvailable, Applied: true}})
	if _, ok := got[spec.ClassExecAllowlist]; ok {
		t.Error("a class that was never probed appears in the observations")
	}

	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, got)
	for _, capability := range record.Capabilities {
		if capability.Name == spec.ClassExecAllowlist && capability.Availability != spec.AvailabilityUnprobed {
			t.Errorf("unprobed class recorded as %q", capability.Availability)
		}
	}
}

// ------------------------------------------------------- forced unavailability

func TestApplyForcedUnavailableIsANoOpWithoutForcing(t *testing.T) {
	results := []ClassResult{{Class: spec.ClassNetworkDenial, Verdict: VerdictAvailable, Applied: true}}
	got := applyForcedUnavailable(results, nil)
	if len(got) != 1 || got[0].Verdict != VerdictAvailable || !got[0].Applied {
		t.Errorf("forcing nothing changed the results: %+v", got)
	}
}

// A forced value is a control, not a measurement. It has to be visible as such
// in the artifact, or a sweep run could be mistaken for a host observation.
func TestApplyForcedUnavailableMarksTheInjection(t *testing.T) {
	results := []ClassResult{
		{Class: spec.ClassNetworkDenial, Verdict: VerdictAvailable, Applied: true, Checks: []Check{{Name: "real", Pass: true}}},
		{Class: spec.ClassExecAllowlist, Verdict: VerdictAvailable, Applied: true},
	}
	got := applyForcedUnavailable(results, []string{spec.ClassNetworkDenial, "not-a-class"})

	forced := got[0]
	if forced.Verdict != VerdictUnavailable || forced.Applied {
		t.Errorf("forced class = %q/applied=%v, want unavailable and not applied", forced.Verdict, forced.Applied)
	}
	var sawInjection bool
	for _, check := range forced.Checks {
		if check.Name == "forced-unavailable-injection" {
			sawInjection = true
			if check.Pass {
				t.Error("the injected check reports a pass")
			}
			if !strings.Contains(check.Detail, "not a host observation") {
				t.Errorf("the injected check does not disclaim itself: %q", check.Detail)
			}
		}
	}
	if !sawInjection {
		t.Error("no injected check was recorded in the forced class")
	}
	var sawReason bool
	for _, reason := range forced.Reasons {
		if strings.Contains(reason, "injected, not measured") {
			sawReason = true
		}
	}
	if !sawReason {
		t.Errorf("reasons %v do not disclose the injection", forced.Reasons)
	}

	// A class nobody forced is untouched, and an unknown name forces nothing.
	if got[1].Verdict != VerdictAvailable || !got[1].Applied {
		t.Errorf("an unforced class was changed: %+v", got[1])
	}
}

func TestClassIsUnapplied(t *testing.T) {
	obs := map[string]evidence.Observation{}
	for _, class := range spec.Classes() {
		obs[class] = evidence.Observation{Availability: spec.AvailabilityAvailable, Applied: true}
	}
	obs[spec.ClassExecAllowlist] = evidence.Observation{Availability: spec.AvailabilityUnavailable}
	record := evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)

	if !classIsUnapplied(record, spec.ClassExecAllowlist) {
		t.Error("the unavailable class was not reported unapplied")
	}
	if classIsUnapplied(record, spec.ClassNetworkDenial) {
		t.Error("an applied class was reported unapplied")
	}
	// A class the record does not carry cannot be confirmed unapplied; saying
	// otherwise would let a missing entry satisfy the fail-closed check.
	if classIsUnapplied(record, "not-a-class") {
		t.Error("an absent class was reported unapplied")
	}
}

// ------------------------------------------------------------- small helpers

func TestObservedHelpers(t *testing.T) {
	if boolObserved(true) != "absent" || boolObserved(false) != "present" {
		t.Error("boolObserved does not describe escape artifacts")
	}
	if boolObservedYes(true) != "yes" || boolObservedYes(false) != "no" {
		t.Error("boolObservedYes is wrong")
	}
	if survivorObserved(true) != "survivor" || survivorObserved(false) != "no-survivor" {
		t.Error("survivorObserved is wrong")
	}
	if aggregateObserved(true) != "aggregate-bound-held" || aggregateObserved(false) != "per-process-only" {
		t.Error("aggregateObserved is wrong")
	}
	if describeStat(nil) != "the escape artifact exists" {
		t.Error("describeStat does not report an existing artifact")
	}
	if got := describeStat(errors.New("no such file")); got != "no such file" {
		t.Errorf("describeStat = %q", got)
	}
	if got := observedOutcome(inside.Attempt{Outcome: inside.OutcomeDenied}, true); got != inside.OutcomeDenied {
		t.Errorf("observedOutcome = %q", got)
	}
	if got := observedOutcome(inside.Attempt{}, false); got != "probe-failed" {
		t.Errorf("observedOutcome for a missing attempt = %q, want probe-failed", got)
	}
}

func TestMembershipAndPolicyObserved(t *testing.T) {
	held := domainOutcome{detachedSID: 42, domainSID: 42}
	if membershipObserved(held) != "membership-held" {
		t.Error("a descendant in the same session is not reported as a member")
	}
	renounced := domainOutcome{detachedSID: 99, domainSID: 42}
	if membershipObserved(renounced) != "membership-renounced" {
		t.Error("a descendant in another session is not reported as escaped")
	}
	// A session of 0 or -1 means the measurement failed; it must not read as a
	// match just because both sides are equally unknown.
	unknown := domainOutcome{detachedSID: -1, domainSID: -1}
	if membershipObserved(unknown) != "membership-renounced" {
		t.Error("an unmeasured session was reported as held membership")
	}

	if policyObserved(domainOutcome{detachedSurvived: false}) != "no-survivor-to-test" {
		t.Error("with no survivor there is nothing to test")
	}
	if policyObserved(domainOutcome{detachedSurvived: true, sandboxInherited: true}) != "policy-inherited" {
		t.Error("an inherited policy is not reported")
	}
	if policyObserved(domainOutcome{detachedSurvived: true}) != "policy-escaped" {
		t.Error("an escaped policy is not reported")
	}
}

// Without descendants there is nothing to kill, so "no survivor" says nothing.
// The classes that depend on them must be inconclusive rather than pass.
func TestSpawnFailure(t *testing.T) {
	cases := []struct {
		name string
		out  domainOutcome
		want string
	}{
		{"both started", domainOutcome{detachedStarted: true, attachedStarted: true}, ""},
		{"neither started", domainOutcome{}, "neither descendant started"},
		{"only the control started", domainOutcome{attachedStarted: true}, "detached descendant did not start"},
		{"only the subject started", domainOutcome{detachedStarted: true}, "negative control is missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.out.spawnFailure()
			if tc.want == "" {
				if got != "" {
					t.Errorf("spawnFailure = %q, want no failure", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("spawnFailure = %q, want it to mention %q", got, tc.want)
			}
		})
	}
}

func TestAtoiOr(t *testing.T) {
	cases := map[string]int{"42": 42, "0": 0, "-1": -1, "": -7, "junk": -7, "1.5": -7}
	for raw, want := range cases {
		if got := atoiOr(raw, -7); got != want {
			t.Errorf("atoiOr(%q) = %d, want %d", raw, got, want)
		}
	}
}

func TestReadMarkerSID(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good")
	if err := os.WriteFile(good, []byte("pid=101 pgid=102 sid=103 ppid=104\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := readMarkerSID(good); got != 103 {
		t.Errorf("readMarkerSID = %d, want 103", got)
	}

	// A marker that is missing or malformed must read as unknown, not as zero:
	// zero would compare equal to another unknown and look like membership.
	for name, content := range map[string]string{
		"malformed": "nothing useful here\n",
		"partial":   "pid=1 pgid=2\n",
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if got := readMarkerSID(path); got != -1 {
			t.Errorf("readMarkerSID(%s) = %d, want -1", name, got)
		}
	}
	if got := readMarkerSID(filepath.Join(dir, "missing")); got != -1 {
		t.Errorf("readMarkerSID(missing) = %d, want -1", got)
	}
}

func TestAlive(t *testing.T) {
	if !alive(os.Getpid()) {
		t.Error("this process was reported dead")
	}
	for _, pid := range []int{0, -1, -12345} {
		if alive(pid) {
			t.Errorf("alive(%d) = true, want false", pid)
		}
	}
}

func TestCanReachOutside(t *testing.T) {
	base := filepath.Join(t.TempDir(), "descendant")
	if canReachOutside(base) {
		t.Error("an escape was reported with no escape marker present")
	}
	if err := os.WriteFile(base+".detached.escaped", []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !canReachOutside(base) {
		t.Error("an escape marker was not detected")
	}
}

// --------------------------------------------------------------- mechanisms

// The mechanism inventory is the reusable part of this prototype: another
// implementation reads it to learn what not to try. A mechanism pointing at a
// class that does not exist would send that reader looking for nothing.
func TestSupportedMechanismsAreWellFormed(t *testing.T) {
	mechanisms := supportedMechanisms()
	if len(mechanisms) == 0 {
		t.Fatal("the mechanism inventory is empty")
	}
	statuses := map[string]bool{
		StatusSupported: true, StatusDeprecated: true, StatusPrivate: true,
		StatusUnavailable: true, StatusConditional: true,
	}
	seen := map[string]bool{}
	for _, mechanism := range mechanisms {
		if mechanism.Name == "" {
			t.Error("a mechanism has no name")
		}
		if seen[mechanism.Name] {
			t.Errorf("mechanism %q is listed twice", mechanism.Name)
		}
		seen[mechanism.Name] = true
		if !statuses[mechanism.Status] {
			t.Errorf("mechanism %q has unknown status %q", mechanism.Name, mechanism.Status)
		}
		if mechanism.Note == "" {
			t.Errorf("mechanism %q carries no explanation", mechanism.Name)
		}
		for _, class := range mechanism.Classes {
			if !spec.IsClass(class) {
				t.Errorf("mechanism %q names unknown class %q", mechanism.Name, class)
			}
		}
	}
}

// The three classes macOS cannot provide are the reason the platform stays
// unqualified. Each must have at least one mechanism explaining why, or the
// outcome is a bare assertion.
func TestBlockingClassesHaveAMechanismExplainingThem(t *testing.T) {
	blocking := []string{spec.ClassDomainMembership, spec.ClassDomainTermination, spec.ClassAggregateBounds}
	for _, class := range blocking {
		var explained bool
		for _, mechanism := range supportedMechanisms() {
			for _, named := range mechanism.Classes {
				if named == class {
					explained = true
				}
			}
		}
		if !explained {
			t.Errorf("no mechanism in the inventory speaks to class %q", class)
		}
	}
}

// ---------------------------------------------------------------- artifacts

func TestWriteArtifacts(t *testing.T) {
	dir := t.TempDir()
	result := Result{
		Report:   Report{ReportVersion: ReportVersion, ExitCode: ExitRejected},
		Evidence: evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil),
		ExitCode: ExitRejected,
	}
	evidencePath := filepath.Join(dir, "evidence.json")
	reportPath := filepath.Join(dir, "report.json")

	if err := WriteArtifacts(result, evidencePath, reportPath); err != nil {
		t.Fatalf("WriteArtifacts: %v", err)
	}
	for _, path := range []string{evidencePath, reportPath} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !json.Valid(data) {
			t.Errorf("%s is not valid JSON:\n%s", path, data)
		}
	}
	// The evidence file must be the closed record and nothing else: a reader
	// validating it against the specification would reject extra fields.
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if _, diag, reason := evidence.Decode(data); diag != "" {
		t.Errorf("the written evidence does not validate: %s: %s", diag, reason)
	}
}

func TestWriteArtifactsSkipsEmptyPaths(t *testing.T) {
	dir := t.TempDir()
	result := Result{Evidence: evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)}

	if err := WriteArtifacts(result, "", ""); err != nil {
		t.Fatalf("WriteArtifacts with no paths: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("WriteArtifacts wrote %d files for empty paths", len(entries))
	}
}

func TestWriteArtifactsReportsWriteFailures(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-dir")
	result := Result{Evidence: evidence.Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)}

	err := WriteArtifacts(result, filepath.Join(missing, "evidence.json"), "")
	if err == nil || !strings.Contains(err.Error(), "write evidence") {
		t.Errorf("WriteArtifacts error = %v, want a write-evidence failure", err)
	}
	err = WriteArtifacts(result, "", filepath.Join(missing, "report.json"))
	if err == nil || !strings.Contains(err.Error(), "write report") {
		t.Errorf("WriteArtifacts error = %v, want a write-report failure", err)
	}
}

func TestReportJSONIsValidAndNewlineTerminated(t *testing.T) {
	report := Report{
		ReportVersion: ReportVersion,
		SpecRevision:  spec.SpecRevision,
		Platform:      spec.PlatformMacOS,
		Classes:       []ClassResult{{Class: spec.ClassNetworkDenial, Verdict: VerdictAvailable}},
		Mechanisms:    supportedMechanisms(),
	}
	data := report.JSON()
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("report JSON does not end with a newline")
	}
	if !json.Valid(data[:len(data)-1]) {
		t.Errorf("report JSON is not valid JSON:\n%s", data)
	}
	// The prototype report must never be mistaken for a curator-spec artifact.
	if !strings.Contains(string(data), ReportVersion) {
		t.Error("the report does not identify its own prototype schema")
	}
}

// --------------------------------------------------------------- host facts

func TestDescribeHost(t *testing.T) {
	host := DescribeHost()
	if host.Arch == "" || host.GoVersion == "" {
		t.Errorf("host facts are incomplete: %+v", host)
	}
	if host.UID != os.Getuid() {
		t.Errorf("host uid %d, want %d", host.UID, os.Getuid())
	}
	// A capability observation only means something together with the host that
	// produced it, so an unreadable fact must say so rather than be empty.
	for name, value := range map[string]string{
		"product version": host.ProductVersion,
		"kernel version":  host.KernelVersion,
		"SIP status":      host.SIPStatus,
	} {
		if value == "" {
			t.Errorf("%s is empty; an unreadable fact must be recorded as unavailable", name)
		}
	}
}

func TestTrimmedOutputReportsAMissingTool(t *testing.T) {
	got := trimmedOutput(filepath.Join(t.TempDir(), "not-a-tool"))
	if !strings.HasPrefix(got, "unavailable: ") {
		t.Errorf("trimmedOutput = %q, want an unavailable marker", got)
	}
}
