package probe

import (
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// The reduction from measurements to checks is the one part of this package
// that can be tested against constructed input, because it is arithmetic over
// numbers the host produced rather than a question put to the kernel. These
// tests therefore assert exactly what the reduction concludes from a given set
// of measurements, including the conclusions the host this prototype runs on
// does not currently produce — a host that gained a capability must change the
// verdict, and that can only be shown by feeding the reduction the numbers such
// a host would report.

func checksByName(checks []Check) map[string]Check {
	out := map[string]Check{}
	for _, check := range checks {
		out[check.Name] = check
	}
	return out
}

func requireCheck(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	check, ok := checksByName(checks)[name]
	if !ok {
		var names []string
		for _, c := range checks {
			names = append(names, c.Name)
		}
		t.Fatalf("no check named %q; got %v", name, names)
	}
	return check
}

// ------------------------------------------------------- one bound reduced

func TestMeasurementChecksWhenTheBoundCannotBeInstalled(t *testing.T) {
	m := inside.BoundMeasurement{
		Kind:           inside.BoundAddressSpace,
		Resource:       "RLIMIT_AS",
		Declared:       256 << 20,
		Ceiling:        512 << 20,
		Installed:      false,
		InstallErrno:   "EINVAL",
		SoftLimitFloor: 400 << 30,
		Control:        inside.BoundRun{Ran: true, Reached: 512 << 20},
	}

	checks := measurementChecks(m)
	if len(checks) != 3 {
		t.Errorf("a bound that cannot be installed produced %d checks, want 3", len(checks))
	}

	declarable := requireCheck(t, checks, inside.BoundAddressSpace+":bound-can-be-declared")
	if declarable.Pass {
		t.Error("a bound the kernel refused is reported as declarable")
	}
	if !strings.Contains(declarable.Detail, "EINVAL") {
		t.Errorf("the refusal detail does not name the errno: %q", declarable.Detail)
	}

	// The matched control: the call itself works, so the refusal is about the
	// value. Without it, EINVAL could equally mean a broken instrument.
	floor := requireCheck(t, checks, inside.BoundAddressSpace+":some-bound-is-accepted-by-the-kernel")
	if !floor.Pass {
		t.Error("a measured floor above the declared budget did not satisfy the control")
	}
	if floor.Kind != kindNegative {
		t.Errorf("the floor control has kind %q, want %q", floor.Kind, kindNegative)
	}

	control := requireCheck(t, checks, inside.BoundAddressSpace+":unbounded-control-exceeds-the-declared-budget")
	if !control.Pass {
		t.Error("an unbounded control that passed the declared budget did not satisfy the control check")
	}

	// A host that accepted no value at all must fail the control rather than
	// pass it silently.
	m.SoftLimitFloor = -1
	if requireCheck(t, measurementChecks(m), inside.BoundAddressSpace+":some-bound-is-accepted-by-the-kernel").Pass {
		t.Error("a kernel that accepted no value still satisfied the floor control")
	}
}

func TestMeasurementChecksWhenTheBoundBindsPerProcess(t *testing.T) {
	// The shape this host produces for CPU: the bound installs, binds the
	// process that set it, and hands its descendant a second full budget.
	m := inside.BoundMeasurement{
		Kind:                   inside.BoundCPU,
		Resource:               "RLIMIT_CPU",
		Declared:               1000,
		Ceiling:                1600,
		Installed:              true,
		Bounded:                inside.BoundRun{Ran: true, Reached: 1001, Refused: true, Refusal: "SIGXCPU"},
		Nested:                 inside.BoundRun{Ran: true, Reached: 1002, Refused: true, Refusal: "SIGXCPU"},
		Control:                inside.BoundRun{Ran: true, Reached: 1600},
		SoftRaiseHardPreserved: inside.EscapeRaised,
		SoftRaiseHardLowered:   "refused:EPERM",
	}

	checks := measurementChecks(m)
	for name, want := range map[string]bool{
		inside.BoundCPU + ":bound-can-be-declared":                         true,
		inside.BoundCPU + ":bound-binds-the-process-that-set-it":           true,
		inside.BoundCPU + ":unbounded-control-exceeds-the-declared-budget": true,
		inside.BoundCPU + ":declared-budget-is-available-to-the-domain":    true,
		// The finding: two processes under one declared bound reached twice it.
		inside.BoundCPU + ":descendant-shares-the-declared-budget": false,
		// The second escape: the member put its own budget back.
		inside.BoundCPU + ":member-cannot-raise-its-own-bound":    false,
		inside.BoundCPU + ":lowered-hard-limit-refuses-the-raise": true,
	} {
		if got := requireCheck(t, checks, name).Pass; got != want {
			t.Errorf("%s pass = %v, want %v", name, got, want)
		}
	}

	aggregate := requireCheck(t, checks, inside.BoundCPU+":descendant-shares-the-declared-budget")
	if aggregate.Observed != "per-process-only" {
		t.Errorf("aggregate observed %q, want per-process-only", aggregate.Observed)
	}
	if !strings.Contains(aggregate.Detail, "2003") {
		t.Errorf("the aggregate detail does not state the total that was reached: %q", aggregate.Detail)
	}
}

// A host that did account the bound over the tree must produce a passing
// aggregate check. This is what makes the check a measurement rather than a
// constant that always reports the macOS answer.
func TestMeasurementChecksCreditAnAggregateBound(t *testing.T) {
	m := inside.BoundMeasurement{
		Kind:      inside.BoundCPU,
		Resource:  "RLIMIT_CPU",
		Declared:  1000,
		Ceiling:   1600,
		Installed: true,
		Bounded:   inside.BoundRun{Ran: true, Reached: 700, Refused: true, Refusal: "SIGXCPU"},
		// The descendant only got what was left of the one shared budget.
		Nested:                 inside.BoundRun{Ran: true, Reached: 300, Refused: true, Refusal: "SIGXCPU"},
		Control:                inside.BoundRun{Ran: true, Reached: 1600},
		SoftRaiseHardPreserved: "refused:EPERM",
		SoftRaiseHardLowered:   "refused:EPERM",
	}

	checks := measurementChecks(m)
	aggregate := requireCheck(t, checks, inside.BoundCPU+":descendant-shares-the-declared-budget")
	if !aggregate.Pass {
		t.Errorf("a domain total inside the declared bound was not credited: %+v", aggregate)
	}
	if aggregate.Observed != "aggregate-bound-held" {
		t.Errorf("aggregate observed %q, want aggregate-bound-held", aggregate.Observed)
	}
	raise := requireCheck(t, checks, inside.BoundCPU+":member-cannot-raise-its-own-bound")
	if !raise.Pass {
		t.Error("a refused soft-limit raise was not credited")
	}
}

// A descendant that never produced a measurement must not be scored as though
// the bound had held. "Nothing was observed" is not "the bound was accounted".
func TestMeasurementChecksDoNotCreditAnUnmeasuredDescendant(t *testing.T) {
	m := inside.BoundMeasurement{
		Kind:      inside.BoundProcessCount,
		Resource:  "RLIMIT_NPROC",
		Declared:  4,
		Ceiling:   6,
		Installed: true,
		Bounded:   inside.BoundRun{Ran: true, Reached: 0, Refused: true, Refusal: "EAGAIN"},
		Nested:    inside.BoundRun{Ran: false, Reached: -1, Refused: true, Refusal: "EAGAIN", Detail: "fork refused"},
		Control:   inside.BoundRun{Ran: true, Reached: 6},

		SoftRaiseHardPreserved: inside.EscapeRaised,
		SoftRaiseHardLowered:   "refused:EPERM",
	}

	checks := measurementChecks(m)
	aggregate := requireCheck(t, checks, inside.BoundProcessCount+":descendant-shares-the-declared-budget")
	if aggregate.Pass {
		t.Error("an unmeasured descendant was credited as sharing the budget")
	}
	if aggregate.Observed != "not-evaluable" {
		t.Errorf("aggregate observed %q, want not-evaluable", aggregate.Observed)
	}

	// The scope finding: the bound refused before the domain had used any of
	// the budget declared for it, which means the budget is not the domain's.
	scoped := requireCheck(t, checks, inside.BoundProcessCount+":declared-budget-is-available-to-the-domain")
	if scoped.Pass {
		t.Error("a bound that refused the domain's first process was scored as domain-scoped")
	}
	if scoped.Observed != "budget-consumed-outside-the-domain" {
		t.Errorf("scope observed %q", scoped.Observed)
	}
}

// A raise that was never attempted is not evidence that the hard limit closed
// the route, so the matched control must not pass on it.
func TestLoweredHardLimitControlRequiresAnAttempt(t *testing.T) {
	m := inside.BoundMeasurement{
		Kind:                   inside.BoundCPU,
		Resource:               "RLIMIT_CPU",
		Declared:               1000,
		Ceiling:                1600,
		Installed:              true,
		Bounded:                inside.BoundRun{Ran: true, Reached: 1001, Refused: true},
		Control:                inside.BoundRun{Ran: true, Reached: 1600},
		SoftRaiseHardPreserved: inside.EscapeNotAttempted,
		SoftRaiseHardLowered:   inside.EscapeNotAttempted,
	}
	control := requireCheck(t, measurementChecks(m), inside.BoundCPU+":lowered-hard-limit-refuses-the-raise")
	if control.Pass {
		t.Error("an unattempted raise satisfied the lowered-hard-limit control")
	}
}

// Every kind the agent measures must be reduced into at least one positive test
// and at least one control, whatever the numbers say. A kind that produced only
// positives could not tell enforcement from a broken measurement.
func TestEveryBoundKindReducesToAPositiveAndAControl(t *testing.T) {
	for _, kind := range inside.BoundKinds() {
		for _, installed := range []bool{true, false} {
			m := inside.BoundMeasurement{
				Kind: kind, Resource: "R", Declared: 100, Ceiling: 200,
				Installed:              installed,
				SoftLimitFloor:         1 << 40,
				Bounded:                inside.BoundRun{Ran: true, Reached: 100, Refused: true},
				Nested:                 inside.BoundRun{Ran: true, Reached: 100, Refused: true},
				Control:                inside.BoundRun{Ran: true, Reached: 200},
				SoftRaiseHardPreserved: "refused:EPERM",
				SoftRaiseHardLowered:   "refused:EPERM",
			}
			kinds := map[string]int{}
			for _, check := range measurementChecks(m) {
				kinds[check.Kind]++
				if check.Expectation == "" || check.Observed == "" {
					t.Errorf("%s (installed=%v): check %q states no expectation or observation",
						kind, installed, check.Name)
				}
			}
			if kinds[kindPositive] == 0 {
				t.Errorf("%s (installed=%v) reduced to no positive test", kind, installed)
			}
			if kinds[kindNegative]+kinds[kindAdversarial] == 0 {
				t.Errorf("%s (installed=%v) reduced to no control", kind, installed)
			}
		}
	}
}

// ------------------------------------------------- supervisor-side accounting

// The verdict the reviewer required to be derived. Each case below is a host
// this prototype does not run on, and the reduction has to answer differently
// for each of them; a hard-coded verdict would answer the same every time.
func TestSupervisorAccountingIsDerivedFromThisRun(t *testing.T) {
	unescapableHost := domainOutcome{
		domainSID: 100, detachedSID: 100,
		detachedStarted: true, attachedStarted: true,
		teardownPGID: 42,
	}

	cases := []struct {
		name     string
		outcome  domainOutcome
		pass     bool
		observed string
	}{
		{
			name:     "membership held, teardown issued, nothing survived",
			outcome:  unescapableHost,
			pass:     true,
			observed: "unescapable",
		},
		{
			name: "a descendant renounced membership",
			outcome: func() domainOutcome {
				out := unescapableHost
				out.detachedSID = 999
				return out
			}(),
			observed: "escapable",
		},
		{
			name: "a descendant survived the teardown",
			outcome: func() domainOutcome {
				out := unescapableHost
				out.detachedSurvived = true
				return out
			}(),
			observed: "escapable",
		},
		{
			name: "the supervisor had no teardown handle",
			outcome: func() domainOutcome {
				out := unescapableHost
				out.teardownPGID = 0
				return out
			}(),
			observed: "escapable",
		},
		{
			name: "the teardown signal failed",
			outcome: func() domainOutcome {
				out := unescapableHost
				out.teardownErr = errFake
				return out
			}(),
			observed: "escapable",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := supervisorAccountingCheck(tc.outcome)
			if check.Pass != tc.pass {
				t.Errorf("pass = %v, want %v (detail %q)", check.Pass, tc.pass, check.Detail)
			}
			if check.Observed != tc.observed {
				t.Errorf("observed = %q, want %q", check.Observed, tc.observed)
			}
			// The verdict has to carry the numbers it was derived from,
			// otherwise a reader cannot tell a derivation from a declaration.
			if !strings.Contains(check.Detail, "derived from this run") {
				t.Errorf("the detail does not say what it was derived from: %q", check.Detail)
			}
		})
	}
}

// A domain probe that did not run says nothing about supervisor accounting, and
// must not be reduced to a verdict either way.
func TestSupervisorAccountingReportsAnUnrunProbe(t *testing.T) {
	cases := map[string]domainOutcome{
		"the domain probe errored":     {err: errFake},
		"no descendant started":        {domainSID: 1},
		"the control never started":    {domainSID: 1, detachedStarted: true},
		"the escape probe never began": {domainSID: 1, attachedStarted: true},
	}
	for name, outcome := range cases {
		t.Run(name, func(t *testing.T) {
			check := supervisorAccountingCheck(outcome)
			if check.Pass {
				t.Error("an unrun domain probe produced a passing supervisor verdict")
			}
			if check.Observed != "probe-failed" {
				t.Errorf("observed = %q, want probe-failed", check.Observed)
			}
		})
	}
}

// ---------------------------------------------------------------- wall clock

func TestWallClockChecksReduceAMeasuredDeadline(t *testing.T) {
	// The shape this host produces: the deadline ends the supervised process,
	// neither descendant goes with it, the group signal reaches the one that
	// stayed, and the detached one outlives everything the supervisor can do.
	w := wallClockOutcome{
		declared: 3 * time.Second, started: true,
		rootPID: 10, rootPGID: 10, attachedPID: 11, detachedPID: 12,
		elapsed: 3 * time.Second, deadlineFired: true,
		attachedAliveAfterDeadline: true, detachedAliveAfterDeadline: true,
		detachedAliveAfterGroupKill: true,
		survivorsBeforeCleanup:      []int{12},
		controlRan:                  true, controlElapsed: 40 * time.Millisecond,
	}

	checks := wallClockChecks(w)
	for name, want := range map[string]bool{
		"wall-clock:deadline-terminates-the-domain-root":            true,
		"wall-clock:control-exits-before-the-deadline":              true,
		"wall-clock:deadline-reaches-the-attached-descendant":       false,
		"wall-clock:deadline-reaches-the-detached-descendant":       false,
		"wall-clock:group-teardown-reaches-the-attached-descendant": true,
		"wall-clock:deadline-cancellation-leaves-no-descendant":     false,
		"wall-clock:harness-leaves-no-descendant-behind":            true,
	} {
		if got := requireCheck(t, checks, name).Pass; got != want {
			t.Errorf("%s pass = %v, want %v", name, got, want)
		}
	}
}

// A host on which the deadline really did end the whole tree must produce a
// passing cancellation check, so the check tracks the host rather than macOS.
func TestWallClockChecksCreditADeadlineThatReachesTheTree(t *testing.T) {
	w := wallClockOutcome{
		declared: 3 * time.Second, started: true,
		rootPID: 10, rootPGID: 10, attachedPID: 11, detachedPID: 12,
		elapsed: 3 * time.Second, deadlineFired: true,
		controlRan: true, controlElapsed: 40 * time.Millisecond,
	}
	checks := wallClockChecks(w)
	for _, name := range []string{
		"wall-clock:deadline-reaches-the-attached-descendant",
		"wall-clock:deadline-reaches-the-detached-descendant",
		"wall-clock:deadline-cancellation-leaves-no-descendant",
	} {
		if !requireCheck(t, checks, name).Pass {
			t.Errorf("%s did not pass on a host where the deadline ended the tree", name)
		}
	}
}

// The harness's own hygiene is reported separately from the platform result. A
// survivor the harness failed to clean up must fail its own check without
// changing what the platform was measured to do.
func TestWallClockChecksReportHarnessSurvivorsSeparately(t *testing.T) {
	w := wallClockOutcome{
		declared: 3 * time.Second, started: true,
		rootPID: 10, rootPGID: 10, attachedPID: 11, detachedPID: 12,
		deadlineFired: true, elapsed: 3 * time.Second,
		survivorsBeforeCleanup: []int{12},
		survivorsAfterCleanup:  []int{12},
		controlRan:             true, controlElapsed: 40 * time.Millisecond,
	}
	if requireCheck(t, wallClockChecks(w), "wall-clock:harness-leaves-no-descendant-behind").Pass {
		t.Error("a descendant the harness failed to clean up did not fail the hygiene check")
	}
}

// A wall-clock probe whose domain never started proves nothing about the
// deadline, so it must report a failed probe rather than a clean result.
func TestWallClockChecksRejectAnUnstartedDomain(t *testing.T) {
	for name, w := range map[string]wallClockOutcome{
		"the probe errored":        {err: errFake},
		"the domain never started": {startFailure: "domain root pid -1"},
	} {
		t.Run(name, func(t *testing.T) {
			checks := wallClockChecks(w)
			if len(checks) != 1 {
				t.Fatalf("got %d checks, want the single failed-probe check", len(checks))
			}
			if checks[0].Pass || checks[0].Observed != "probe-failed" {
				t.Errorf("an unstarted wall-clock probe reduced to %+v", checks[0])
			}
		})
	}
}

// A control that hit the deadline itself cannot show that the deadline is what
// ended the measured run.
func TestWallClockControlMustFinishInTime(t *testing.T) {
	w := wallClockOutcome{
		declared: 3 * time.Second, started: true,
		rootPID: 10, rootPGID: 10, attachedPID: 11, detachedPID: 12,
		deadlineFired: true, elapsed: 3 * time.Second,
		controlRan: true, controlElapsed: 3 * time.Second,
	}
	control := requireCheck(t, wallClockChecks(w), "wall-clock:control-exits-before-the-deadline")
	if control.Pass {
		t.Error("a control that ran into the deadline satisfied the control check")
	}
	if control.Observed != "control-hit-the-deadline" {
		t.Errorf("observed = %q", control.Observed)
	}
}

// ---------------------------------------------------------------- mechanisms

// The gap this annotation closes: an inventory entry that was never probed must
// not read like a measurement.
func TestAnnotateMechanismsSeparatesMeasuredFromDeclared(t *testing.T) {
	results := []ClassResult{
		{
			Class: spec.ClassAggregateBounds,
			Checks: []Check{
				{Name: "cpu-milliseconds:bound-can-be-declared", Observed: "bound-installed", Pass: true},
				{Name: "cpu-milliseconds:bound-binds-the-process-that-set-it", Observed: "refused-at-the-bound", Pass: true},
				{Name: "cpu-milliseconds:descendant-shares-the-declared-budget", Observed: "per-process-only"},
				{Name: "cpu-milliseconds:member-cannot-raise-its-own-bound", Observed: "member-restored-its-budget"},
			},
		},
	}

	byName := map[string]Mechanism{}
	for _, mechanism := range annotateMechanisms(supportedMechanisms(), results) {
		byName[mechanism.Name] = mechanism
	}

	cpu, ok := byName["setrlimit RLIMIT_CPU"]
	if !ok {
		t.Fatal("the inventory lost the RLIMIT_CPU entry")
	}
	if !cpu.Exercised {
		t.Error("a mechanism with recorded observations is not marked exercised")
	}
	if !strings.Contains(cpu.Observation, "per-process-only") {
		t.Errorf("the observation does not carry what was measured: %q", cpu.Observation)
	}

	// RLIMIT_AS is named in the evidence map but this run recorded nothing for
	// it, so it must say the observation is missing rather than inherit CPU's.
	as := byName["setrlimit RLIMIT_AS (RLIMIT_RSS on Darwin)"]
	if !strings.Contains(as.Observation, "not-recorded") {
		t.Errorf("a mechanism with no recorded check does not say so: %q", as.Observation)
	}

	// A mechanism that was only considered must say it was not measured.
	guest := byName["Virtualization.framework guest"]
	if guest.Exercised {
		t.Error("a mechanism nothing probed is marked exercised")
	}
	if !strings.Contains(guest.Observation, "not exercised") {
		t.Errorf("an unexercised mechanism does not say so: %q", guest.Observation)
	}
}

// Every mechanism the evidence map names must exist in the inventory. A typo
// there would silently downgrade a measured mechanism to a declared one.
func TestMechanismEvidenceNamesRealMechanisms(t *testing.T) {
	known := map[string]bool{}
	for _, mechanism := range supportedMechanisms() {
		if known[mechanism.Name] {
			t.Errorf("the inventory lists %q twice", mechanism.Name)
		}
		known[mechanism.Name] = true
	}
	for name := range mechanismEvidence() {
		if !known[name] {
			t.Errorf("the evidence map names %q, which is not in the inventory", name)
		}
	}
}

func TestUnexercisedMechanismsListsTheBoundaryOfTheEvidence(t *testing.T) {
	mechanisms := []Mechanism{
		{Name: "measured", Exercised: true},
		{Name: "considered-b"},
		{Name: "considered-a"},
	}
	got := UnexercisedMechanisms(mechanisms)
	if len(got) != 2 || got[0] != "considered-a" || got[1] != "considered-b" {
		t.Errorf("UnexercisedMechanisms = %v, want the two unexercised names in order", got)
	}
}

// ------------------------------------------------------------------- shared

// errFake stands in for a failure whose text carries nothing the reduction
// reads. Using a real syscall error here would suggest the reduction inspects
// it, which it must not: any failure to measure is a failure to measure.
var errFake = fakeError("the probe did not run")

type fakeError string

func (e fakeError) Error() string { return string(e) }
