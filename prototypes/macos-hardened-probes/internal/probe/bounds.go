package probe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// ------------------------------------------------- aggregate-resource-bounds
//
// Section 5.4 requires every declared bound to be enforced over the whole
// descendant tree, not over one process. This probe measures that for six
// quantities: descriptors, bytes written below the build root, CPU time,
// address space, data segment, and process count, plus the one bound no
// resource limit expresses at all — wall-clock time.
//
// Nothing here is declared from a platform table. Every statement below is
// reduced from a process that installed a real bound and then tried to exceed
// it, from a matched control that made the same attempt with no bound, and from
// a descendant that inherited the bound and tried again.

// The bounds the probe declares for the domain. They are chosen by the harness,
// never by anything under test, and are stated once so the profile the agent is
// given and the expectation the report is scored against cannot drift apart.
const (
	declaredNoFileCap   = 64
	declaredWriteBudget = 4 << 20
)

func (r runContext) probeAggregateBounds(domain domainOutcome, wall wallClockOutcome) ClassResult {
	return timed(spec.ClassAggregateBounds,
		"POSIX RLIMIT_* (per process) and a supervisor deadline; macOS exposes no job object, cgroup or per-directory byte cap",
		func() ([]Check, bool) {
			var checks []Check
			checks = append(checks, r.descriptorAndDiskChecks()...)
			checks = append(checks, r.boundMatrixChecks()...)
			checks = append(checks, wallClockChecks(wall)...)
			checks = append(checks, supervisorAccountingCheck(domain))
			return checks, true
		})
}

// ------------------------------------------ descriptors and bytes on disk

// descriptorAndDiskChecks are the two bounds the earlier revision of this probe
// measured. They are retained unchanged in substance: a descriptor budget and a
// byte budget below the private build root.
func (r runContext) descriptorAndDiskChecks() []Check {
	if err := r.env.ResetBuildRoot(); err != nil {
		return []Check{failedCheck("descriptor-budget-is-per-process", kindPositive,
			"every bound is accounted over the whole domain in aggregate", err.Error())}
	}
	report, _, err := r.runAgent("resource-bounds", r.baseProfile(), inside.OpResourceBounds,
		[]string{
			inside.EnvNoFileCap + "=" + strconv.Itoa(declaredNoFileCap),
			inside.EnvWriteBytes + "=" + strconv.Itoa(declaredWriteBudget),
		}, nil)
	if err != nil {
		return []Check{failedCheck("descriptor-budget-is-per-process", kindPositive,
			"every bound is accounted over the whole domain in aggregate", err.Error())}
	}

	var checks []Check

	self := atoiOr(report.Values["self_descriptors"], -1)
	child := atoiOr(report.Values["child_descriptors"], -1)

	// Negative control: the per-process bound has to bind at all. If the domain
	// root could open far more than the declared descriptors, the instrument is
	// broken and the aggregate finding below would be an artifact of a limit
	// that never applied.
	setrlimit, setrlimitFound := report.Find("setrlimit-nofile")
	perProcessHeld := setrlimitFound &&
		setrlimit.Outcome == inside.OutcomeAllowed &&
		self > 0 && self <= declaredNoFileCap
	checks = append(checks, Check{
		Name:        "descriptors:per-process-bound-binds",
		Kind:        kindNegative,
		Expectation: "the declared descriptor bound binds the process that set it, so the aggregate result below is a real measurement",
		Observed:    boundObserved(perProcessHeld),
		Pass:        perProcessHeld,
		Detail: fmt.Sprintf("declared soft limit %d; the domain root opened %d descriptors before the kernel refused",
			declaredNoFileCap, self),
	})

	descriptorCheck := Check{
		Name:        "descriptors:descendant-shares-the-declared-budget",
		Kind:        kindAdversarial,
		Expectation: "a descriptor bound is accounted over the domain, so a parent and a child share one budget",
		Detail: fmt.Sprintf("soft limit %d; the domain root opened %d and its child opened %d, aggregate %d",
			declaredNoFileCap, self, child, self+child),
	}
	switch {
	case self <= 0 || child <= 0:
		descriptorCheck.Observed = "probe-failed"
		descriptorCheck.Detail = "the descriptor counts could not be collected; nothing was observed about aggregation"
	case self+child <= declaredNoFileCap:
		descriptorCheck.Observed = "aggregate-bound-held"
		descriptorCheck.Pass = true
	default:
		descriptorCheck.Observed = "per-process-only"
	}
	checks = append(checks, descriptorCheck)

	written := int64(atoiOr(report.Values["bytes_written"], 0))
	budget := int64(atoiOr(report.Values["byte_budget"], 0))
	byteBoundHeld := budget > 0 && written <= budget
	checks = append(checks, Check{
		Name:        "disk-bytes:write-past-declared-byte-budget",
		Kind:        kindAdversarial,
		Expectation: "writing past the declared aggregate byte budget below the build root is refused",
		Observed:    aggregateObserved(byteBoundHeld),
		Pass:        byteBoundHeld,
		Detail: fmt.Sprintf("declared budget %d bytes; the domain wrote %d bytes with no refusal",
			budget, written),
	})
	return checks
}

// ------------------------------------- CPU, memory and process-count bounds

// boundMatrixChecks runs the bound matrix inside a probe domain and reduces the
// four measurements it produced.
func (r runContext) boundMatrixChecks() []Check {
	if err := r.env.ResetBuildRoot(); err != nil {
		return []Check{failedCheck("bound-matrix", kindPositive,
			"CPU, memory and process-count bounds are accounted over the whole domain", err.Error())}
	}
	markerBase, err := r.env.freshMarkerBase("bound-matrix")
	if err != nil {
		return []Check{failedCheck("bound-matrix", kindPositive,
			"CPU, memory and process-count bounds are accounted over the whole domain", err.Error())}
	}

	profile := r.baseProfile()
	profile.WritablePaths = append(profile.WritablePaths, r.env.MarkerDir)
	report, _, err := r.runAgent("bound-matrix", profile, inside.OpBoundMatrix,
		[]string{inside.EnvMarker + "=" + markerBase}, nil)
	if err != nil {
		return []Check{failedCheck("bound-matrix", kindPositive,
			"CPU, memory and process-count bounds are accounted over the whole domain", err.Error())}
	}

	byKind := map[string]inside.BoundMeasurement{}
	for _, m := range report.Bounds {
		byKind[m.Kind] = m
	}

	var checks []Check
	for _, kind := range inside.BoundKinds() {
		m, ok := byKind[kind]
		if !ok {
			checks = append(checks, failedCheck(kind+":bound-can-be-declared", kindPositive,
				"the declared bound can be installed on a domain member",
				"the in-domain agent reported no measurement for this bound"))
			continue
		}
		checks = append(checks, measurementChecks(m)...)
	}
	return checks
}

// measurementChecks reduces one bound measurement.
//
// The shape depends on what was actually observable. A bound the kernel refuses
// to install cannot be asked whether it binds, so those checks are not emitted
// as failures of an experiment that never ran; what is emitted instead is the
// matched control showing the refusal came from the value and not from a broken
// call.
func measurementChecks(m inside.BoundMeasurement) []Check {
	prefix := m.Kind + ":"

	declarable := Check{
		Name:        prefix + "bound-can-be-declared",
		Kind:        kindPositive,
		Expectation: "a bound at the declared build budget can be installed on a domain member",
		Observed:    installObserved(m.Installed),
		Pass:        m.Installed,
		Detail: fmt.Sprintf("%s declared at %d %s", m.Resource, m.Declared,
			unitOf(m.Kind)),
	}
	if !m.Installed {
		declarable.Detail = fmt.Sprintf("%s refused the declared value %d %s with %s",
			m.Resource, m.Declared, unitOf(m.Kind), orUnknown(m.InstallErrno))
	}

	control := Check{
		Name:        prefix + "unbounded-control-exceeds-the-declared-budget",
		Kind:        kindNegative,
		Expectation: "with no bound installed the same attempt passes the declared budget, so a refusal above is attributable to the bound",
		Observed:    reachObserved(m.Control, m.Declared),
		Pass:        m.Control.Ran && !m.Control.Refused && m.Control.Reached > m.Declared,
		Detail: fmt.Sprintf("the unbounded control reached %d %s against a declared budget of %d: %s",
			m.Control.Reached, unitOf(m.Kind), m.Declared, m.Control.Detail),
	}

	if !m.Installed {
		// The bound could not be installed. The only further question that can
		// be answered is whether the call itself works, which is what the floor
		// measurement shows: some value is accepted, so the refusal above is
		// about the value and not about a broken instrument.
		floor := Check{
			Name:        prefix + "some-bound-is-accepted-by-the-kernel",
			Kind:        kindNegative,
			Expectation: "the kernel accepts some value for this resource, so the refusal above is attributable to the declared value rather than to a broken call",
			Observed:    floorObserved(m.SoftLimitFloor),
			Pass:        m.SoftLimitFloor > 0,
			Detail: fmt.Sprintf("the lowest soft limit this kernel accepted for %s was %d %s, which is %s the declared build budget of %d",
				m.Resource, m.SoftLimitFloor, unitOf(m.Kind),
				timesLarger(m.SoftLimitFloor, m.Declared), m.Declared),
		}
		return []Check{declarable, floor, control}
	}

	binds := Check{
		Name:        prefix + "bound-binds-the-process-that-set-it",
		Kind:        kindPositive,
		Expectation: "the process that installed the bound is refused when it tries to pass it",
		Observed:    refusalObserved(m.Bounded),
		Pass:        m.Bounded.Ran && m.Bounded.Refused && m.Bounded.Reached < m.Ceiling,
		Detail: fmt.Sprintf("reached %d %s against a declared budget of %d before %s: %s",
			m.Bounded.Reached, unitOf(m.Kind), m.Declared, orUnknown(m.Bounded.Refusal), m.Bounded.Detail),
	}

	// A bound that refuses before the domain has used the budget it was given
	// is not bounding the domain: it is bounding something the domain shares
	// with processes that are not members. RLIMIT_NPROC is the case that makes
	// this visible, because it is accounted per user.
	scoped := Check{
		Name:        prefix + "declared-budget-is-available-to-the-domain",
		Kind:        kindAdversarial,
		Expectation: "the domain can actually use the budget that was declared for it, so the bound is scoped to the domain rather than shared with processes outside it",
		Observed:    scopeObserved(m.Bounded, m.Declared),
		Pass:        m.Bounded.Ran && m.Bounded.Reached >= m.Declared,
		Detail: fmt.Sprintf("the domain reached %d %s of its declared %d before the bound refused",
			m.Bounded.Reached, unitOf(m.Kind), m.Declared),
	}

	aggregate := Check{
		Name:        prefix + "descendant-shares-the-declared-budget",
		Kind:        kindAdversarial,
		Expectation: "a descendant that inherits the bound draws on the same budget, so the domain total cannot pass the declared bound",
	}
	switch {
	case !m.Nested.Ran:
		aggregate.Observed = "not-evaluable"
		aggregate.Detail = fmt.Sprintf(
			"the descendant produced no measurement (%s), so whether it would have drawn on the same budget was not established",
			orUnknown(m.Nested.Detail))
	case m.Bounded.Reached+m.Nested.Reached <= m.Declared:
		aggregate.Observed = "aggregate-bound-held"
		aggregate.Pass = true
		aggregate.Detail = fmt.Sprintf("the domain root reached %d and its descendant %d %s, aggregate %d against a declared %d",
			m.Bounded.Reached, m.Nested.Reached, unitOf(m.Kind),
			m.Bounded.Reached+m.Nested.Reached, m.Declared)
	default:
		aggregate.Observed = "per-process-only"
		aggregate.Detail = fmt.Sprintf("the domain root reached %d and its descendant a fresh %d %s, aggregate %d against a declared %d",
			m.Bounded.Reached, m.Nested.Reached, unitOf(m.Kind),
			m.Bounded.Reached+m.Nested.Reached, m.Declared)
	}

	// The second escape: a soft limit is a floor the bounded process consented
	// to, and POSIX lets it raise its own soft limit back to the hard limit.
	raise := Check{
		Name:        prefix + "member-cannot-raise-its-own-bound",
		Kind:        kindAdversarial,
		Expectation: "a domain member cannot restore the budget the supervisor took away from it",
		Observed:    raiseObserved(m.SoftRaiseHardPreserved),
		Pass:        m.SoftRaiseHardPreserved != "" && m.SoftRaiseHardPreserved != inside.EscapeRaised,
		Detail: fmt.Sprintf("with the hard limit left as inherited, raising the soft limit back to it was %q",
			orUnknown(m.SoftRaiseHardPreserved)),
	}
	raiseControl := Check{
		Name:        prefix + "lowered-hard-limit-refuses-the-raise",
		Kind:        kindNegative,
		Expectation: "with the hard limit lowered too, the same raise is refused, so the result above is attributable to the hard limit",
		Observed:    raiseObserved(m.SoftRaiseHardLowered),
		Pass: m.SoftRaiseHardLowered != "" && m.SoftRaiseHardLowered != inside.EscapeRaised &&
			m.SoftRaiseHardLowered != inside.EscapeNotAttempted,
		Detail: fmt.Sprintf("with the hard limit lowered to the declared value, raising the soft limit was %q",
			orUnknown(m.SoftRaiseHardLowered)),
	}

	return []Check{declarable, binds, control, scoped, aggregate, raise, raiseControl}
}

// ---------------------------------------------------------- wall-clock time

func wallClockChecks(w wallClockOutcome) []Check {
	if w.err != nil {
		return []Check{failedCheck("wall-clock:deadline-terminates-the-domain-root", kindPositive,
			"a declared wall-clock bound ends the domain and everything in it", w.err.Error())}
	}
	if !w.started {
		return []Check{failedCheck("wall-clock:deadline-terminates-the-domain-root", kindPositive,
			"a declared wall-clock bound ends the domain and everything in it",
			"the domain or its descendants never started, so survival could not be observed: "+w.startFailure)}
	}

	checks := []Check{
		{
			Name:        "wall-clock:deadline-terminates-the-domain-root",
			Kind:        kindPositive,
			Expectation: "the declared wall-clock bound ends the process the supervisor started, while it still had work to do",
			Observed:    survivorObserved(w.rootAliveAfterDeadline),
			Pass:        w.deadlineFired && !w.rootAliveAfterDeadline,
			Detail: fmt.Sprintf("declared %s; the domain root ran %s and the deadline %s",
				w.declared, w.elapsed.Round(time.Millisecond), firedObserved(w.deadlineFired)),
		},
		{
			Name:        "wall-clock:control-exits-before-the-deadline",
			Kind:        kindNegative,
			Expectation: "a domain whose work fits inside the deadline exits on its own, so the termination above is attributable to the deadline",
			Observed:    controlObserved(w),
			Pass:        w.controlRan && w.controlErr == nil && w.controlExitCode == 0 && w.controlElapsed < w.declared,
			Detail: fmt.Sprintf("the control domain exited with status %d after %s, inside the declared %s",
				w.controlExitCode, w.controlElapsed.Round(time.Millisecond), w.declared),
		},
		{
			Name:        "wall-clock:deadline-reaches-the-attached-descendant",
			Kind:        kindAdversarial,
			Expectation: "cancelling the supervised process at the deadline also ends the descendants it created",
			Observed:    survivorObserved(w.attachedAliveAfterDeadline),
			Pass:        !w.attachedAliveAfterDeadline,
			Detail: fmt.Sprintf("attached descendant pid %d after the deadline cancelled the domain root; "+
				"cancelling a supervised process signals that process alone, not the tree below it", w.attachedPID),
		},
		{
			Name:        "wall-clock:deadline-reaches-the-detached-descendant",
			Kind:        kindAdversarial,
			Expectation: "a descendant that left the supervisor's process group is still ended by the deadline",
			Observed:    survivorObserved(w.detachedAliveAfterDeadline),
			Pass:        !w.detachedAliveAfterDeadline,
			Detail:      fmt.Sprintf("detached descendant pid %d after the deadline cancelled the domain root", w.detachedPID),
		},
		{
			Name:        "wall-clock:group-teardown-reaches-the-attached-descendant",
			Kind:        kindNegative,
			Expectation: "the strongest teardown a macOS supervisor can issue does end a descendant that stayed in the group, proving the teardown lands",
			Observed:    survivorObserved(w.attachedAliveAfterGroupKill),
			Pass:        w.groupKillErr == nil && !w.attachedAliveAfterGroupKill,
			Detail:      groupKillDetail(w),
		},
		{
			Name:        "wall-clock:deadline-cancellation-leaves-no-descendant",
			Kind:        kindPositive,
			Expectation: "once the declared wall-clock bound has been enforced by every means the supervisor has, no domain member is left running",
			Observed:    remainingObserved(w.survivorsBeforeCleanup),
			Pass:        len(w.survivorsBeforeCleanup) == 0,
			Detail: fmt.Sprintf("still running after the deadline and the process-group teardown: %v",
				w.survivorsBeforeCleanup),
		},
		{
			Name:        "wall-clock:harness-leaves-no-descendant-behind",
			Kind:        kindNegative,
			Expectation: "the probe itself leaves nothing running on the host, whatever the platform failed to do",
			Observed:    remainingObserved(w.survivorsAfterCleanup),
			Pass:        len(w.survivorsAfterCleanup) == 0,
			Detail: "the harness signals every descendant it recorded by pid; a production implementation cannot " +
				"rely on this, because it only works for descendants the supervisor already knew about",
		},
	}
	return checks
}

func groupKillDetail(w wallClockOutcome) string {
	if w.groupKillErr != nil {
		return fmt.Sprintf("process group %d: %v", w.rootPGID, w.groupKillErr)
	}
	return fmt.Sprintf("SIGKILL delivered to process group %d; attached descendant pid %d, detached descendant pid %d",
		w.rootPGID, w.attachedPID, w.detachedPID)
}

// ----------------------------------------------- supervisor-side accounting

// supervisorAccountingCheck is the conclusion the specification allows a
// supervisor to reach only when membership and atomic termination hold, and it
// is derived here from the membership and termination this very run measured.
//
// It is deliberately not a constant. The harness's stated property is that a
// host which gained a capability changes the observation, and a hard-coded
// verdict would break that for the one class where it matters most.
func supervisorAccountingCheck(domain domainOutcome) Check {
	check := Check{
		Name:        "supervisor-side-accounting-is-unescapable",
		Kind:        kindPositive,
		Expectation: "supervisor-side accounting is equivalent to host accounting only when no domain member can evade or survive it",
	}
	if domain.err != nil {
		check.Observed = "probe-failed"
		check.Detail = "the domain probe did not run, so membership and termination were not measured: " + domain.err.Error()
		return check
	}
	if why := domain.spawnFailure(); why != "" {
		check.Observed = "probe-failed"
		check.Detail = "the domain probe produced no descendants, so membership and termination were not measured: " + why
		return check
	}

	membershipHeld := domain.domainSID > 0 && domain.detachedSID == domain.domainSID
	teardownIssued := domain.teardownPGID > 0 && domain.teardownErr == nil
	noSurvivor := !domain.detachedSurvived && !domain.attachedSurvived

	check.Pass = membershipHeld && teardownIssued && noSurvivor
	check.Observed = escapableObserved(check.Pass)
	check.Detail = fmt.Sprintf(
		"derived from this run: domain session %d, detached descendant session %d (membership %s); "+
			"teardown %s on process group %d; survivors after teardown: detached=%v attached=%v",
		domain.domainSID, domain.detachedSID, heldObserved(membershipHeld),
		teardownObserved(domain), domain.teardownPGID,
		domain.detachedSurvived, domain.attachedSurvived)
	return check
}

// ------------------------------------------------------------- vocabulary

func aggregateObserved(held bool) string {
	if held {
		return "aggregate-bound-held"
	}
	return "per-process-only"
}

// boundObserved is the negative control's vocabulary: it says whether the bound
// bound anything at all, which is a different statement from whether it was
// accounted in aggregate.
func boundObserved(held bool) string {
	if held {
		return "bound-binds"
	}
	return "bound-did-not-bind"
}

func installObserved(installed bool) string {
	if installed {
		return "bound-installed"
	}
	return "install-refused"
}

func refusalObserved(run inside.BoundRun) string {
	switch {
	case !run.Ran:
		return "no-measurement"
	case run.Refused:
		return "refused-at-the-bound"
	default:
		return "passed-the-bound"
	}
}

func reachObserved(run inside.BoundRun, declared int64) string {
	switch {
	case !run.Ran:
		return "no-measurement"
	case run.Refused:
		return "refused-without-a-bound"
	case run.Reached > declared:
		return "exceeded-the-declared-budget"
	default:
		return "stopped-below-the-declared-budget"
	}
}

func scopeObserved(run inside.BoundRun, declared int64) string {
	switch {
	case !run.Ran:
		return "no-measurement"
	case run.Reached >= declared:
		return "domain-received-its-budget"
	default:
		return "budget-consumed-outside-the-domain"
	}
}

func floorObserved(floor int64) string {
	if floor > 0 {
		return "a-larger-bound-is-accepted"
	}
	return "no-bound-is-accepted"
}

func raiseObserved(result string) string {
	switch result {
	case inside.EscapeRaised:
		return "member-restored-its-budget"
	case inside.EscapeNotAttempted, "":
		return "not-attempted"
	default:
		return "refused"
	}
}

func escapableObserved(unescapable bool) string {
	if unescapable {
		return "unescapable"
	}
	return "escapable"
}

func heldObserved(held bool) string {
	if held {
		return "held"
	}
	return "renounced"
}

func firedObserved(fired bool) string {
	if fired {
		return "fired"
	}
	return "did not fire"
}

func controlObserved(w wallClockOutcome) string {
	switch {
	case !w.controlRan || w.controlErr != nil:
		return "control-did-not-run"
	case w.controlElapsed >= w.declared:
		return "control-hit-the-deadline"
	case w.controlExitCode != 0:
		return "control-exited-nonzero"
	default:
		return "control-exited-in-time"
	}
}

func remainingObserved(pids []int) string {
	if len(pids) == 0 {
		return "no-survivor"
	}
	return fmt.Sprintf("%d survivors", len(pids))
}

func unitOf(kind string) string {
	switch kind {
	case inside.BoundCPU:
		return "CPU-milliseconds"
	case inside.BoundAddressSpace, inside.BoundDataSegment:
		return "bytes"
	case inside.BoundProcessCount:
		return "processes"
	default:
		return "units"
	}
}

// timesLarger renders how far a measured floor sits above a declared budget, so
// a reader does not have to divide two large numbers to see the gap.
func timesLarger(floor, declared int64) string {
	if declared <= 0 || floor <= 0 {
		return "not comparable to"
	}
	return fmt.Sprintf("%d times", floor/declared)
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
