package inside

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// The stress functions are exercised in this process where that is safe, and as
// subprocesses of the built agent where it is not. A bound that ends the process
// it binds cannot be measured by the process it binds, which is the same reason
// the agent itself runs every bounded attempt in a separate process.

// ------------------------------------------------------------ budget shapes

// Every ceiling has to sit above its declared bound. If it did not, a stress
// that stopped at the ceiling would be indistinguishable from one the bound
// refused, and the negative control could never show the attempt was real.
//
// This is not a hypothetical: the first version of this file declared the CPU
// bound in milliseconds and handed the same number to setrlimit, which takes
// seconds, and every CPU measurement came back unrefused because the bound that
// was installed was a thousand times the one that was declared.
func TestEveryBoundKindHasACeilingAboveItsDeclaredBudget(t *testing.T) {
	for _, kind := range BoundKinds() {
		declared, ceiling := boundBudget(kind)
		if declared <= 0 {
			t.Errorf("%s declares a budget of %d", kind, declared)
		}
		if ceiling <= declared {
			t.Errorf("%s: ceiling %d is not above the declared budget %d, so a stress that "+
				"stopped at the ceiling could not be told from one the bound refused",
				kind, ceiling, declared)
		}
	}
}

func TestEveryBoundKindNamesAResource(t *testing.T) {
	for _, kind := range BoundKinds() {
		if _, ok := boundResource(kind); !ok {
			t.Errorf("%s maps to no resource", kind)
		}
		if name := boundResourceName(kind); name == "" || name == "unknown" {
			t.Errorf("%s has resource name %q", kind, name)
		}
	}
	if _, ok := boundResource("not-a-kind"); ok {
		t.Error("an unknown kind resolved to a resource")
	}
	if got := boundResourceName("not-a-kind"); got != "unknown" {
		t.Errorf("boundResourceName(unknown) = %q, want %q", got, "unknown")
	}
}

// RLIMIT_CPU counts whole seconds; the bound is declared in milliseconds
// because that is the resolution getrusage reports consumption at. The
// conversion is the seam between the two, and getting it wrong installs a bound
// nothing can hit.
func TestRlimitValueConvertsCPUMillisecondsToWholeSeconds(t *testing.T) {
	cases := map[int64]uint64{
		1000: 1,
		1500: 2, // rounded up: a bound must never be looser than declared
		2000: 2,
		1:    1, // never zero, which would refuse instantly and measure nothing
		0:    1,
	}
	for declared, want := range cases {
		if got := rlimitValue(BoundCPU, declared); got != want {
			t.Errorf("rlimitValue(cpu, %d) = %d, want %d", declared, got, want)
		}
	}
	// Every other kind is already in the unit its resource takes.
	for _, kind := range []string{BoundAddressSpace, BoundDataSegment, BoundProcessCount} {
		if got := rlimitValue(kind, 4096); got != 4096 {
			t.Errorf("rlimitValue(%s, 4096) = %d, want 4096", kind, got)
		}
	}
}

// ----------------------------------------------------------- the stresses

func TestStressCPUStopsAtItsCeiling(t *testing.T) {
	if testing.Short() {
		t.Skip("burns CPU")
	}
	const ceiling = 300

	before := cpuMillis()
	run := stressCPU(ceiling)

	if !run.Ran {
		t.Fatalf("the CPU stress produced no measurement: %+v", run)
	}
	if run.Refused {
		t.Errorf("with no bound installed the CPU stress was refused: %+v", run)
	}
	if run.Reached < ceiling {
		t.Errorf("reached %d CPU-milliseconds, want at least the ceiling %d", run.Reached, ceiling)
	}
	// It must be measuring this process's consumption, not wall time.
	if run.Reached < before {
		t.Errorf("reached %d, which is below the %d already consumed before it started",
			run.Reached, before)
	}
}

func TestStressMemoryStopsAtItsCeiling(t *testing.T) {
	if testing.Short() {
		t.Skip("maps and touches memory")
	}
	const ceiling = 64 << 20

	run := stressMemory(ceiling)
	if !run.Ran {
		t.Fatalf("the memory stress produced no measurement: %+v", run)
	}
	if run.Refused {
		t.Errorf("with no bound installed the memory stress was refused: %+v", run)
	}
	if run.Reached < ceiling {
		t.Errorf("mapped %d bytes, want at least the ceiling %d", run.Reached, ceiling)
	}
	// The mappings are released: a stress that leaked them would make every
	// later measurement in the same process meaningless.
	second := stressMemory(ceiling)
	if !second.Ran || second.Refused {
		t.Errorf("a second memory stress did not behave like the first: %+v", second)
	}
}

func TestStressProcessesStopsAtItsCeiling(t *testing.T) {
	if testing.Short() {
		t.Skip("starts processes")
	}
	dir := t.TempDir()
	t.Setenv(EnvSelf, buildProbeBinary(t))
	t.Setenv(EnvMarker, filepath.Join(dir, "m"))

	const ceiling = 3
	run := stressProcesses(ceiling)
	if !run.Ran {
		t.Fatalf("the process stress produced no measurement: %+v", run)
	}
	if run.Refused {
		t.Errorf("with no bound installed the process stress was refused: %+v", run)
	}
	if run.Reached != ceiling {
		t.Errorf("started %d descendants, want the ceiling %d", run.Reached, ceiling)
	}

	// The descendants really ran: each writes a marker naming its own pid.
	// Without this the count would only prove that fork returned.
	for i := 0; i < ceiling; i++ {
		path := filepath.Join(dir, "m.proc-"+strconv.Itoa(i))
		if !waitForFile(path, 3*time.Second) {
			t.Errorf("descendant %d wrote no marker at %s", i, path)
			continue
		}
		data, err := os.ReadFile(path) //nolint:gosec // a path this test just built under its own temp dir
		if err != nil {
			t.Errorf("read marker %s: %v", path, err)
			continue
		}
		if !bytes.Contains(data, []byte("pid=")) {
			t.Errorf("marker %s does not name a pid: %q", path, data)
		}
	}

	// Nothing it started is still running once it returned.
	for i := 0; i < ceiling; i++ {
		pid := markerPID(t, filepath.Join(dir, "m.proc-"+strconv.Itoa(i)))
		if pid > 0 && syscall.Kill(pid, 0) == nil {
			t.Errorf("descendant %d (pid %d) is still running after the stress returned", i, pid)
		}
	}
}

func TestStressProcessesWithoutAnAgent(t *testing.T) {
	unsetEnv(t, EnvSelf)
	run := stressProcesses(2)
	if run.Ran {
		t.Errorf("with no agent to start, the process stress claimed a measurement: %+v", run)
	}
}

// markerPID reads the pid a descendant wrote into its marker, or -1 when the
// marker is missing or is not one.
func markerPID(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // a path this test built under its own temp dir
	if err != nil {
		return -1
	}
	var pid, pgid, sid, ppid int
	if _, err := fmt.Sscanf(string(data), "pid=%d pgid=%d sid=%d ppid=%d", &pid, &pgid, &sid, &ppid); err != nil {
		return -1
	}
	return pid
}

// ------------------------------------------------------------- the floor

// measureSoftLimitFloor changes the limit it is searching over. Its contract is
// that it puts the original back: a floor measurement that left the process
// under a lowered limit would poison the stress that runs after it.
func TestMeasureSoftLimitFloorRestoresTheOriginalLimit(t *testing.T) {
	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		t.Skipf("cannot read RLIMIT_NOFILE: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &original) })

	measureSoftLimitFloor(syscall.RLIMIT_NOFILE, 16, original)

	var after syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &after); err != nil {
		t.Fatalf("getrlimit: %v", err)
	}
	if after.Cur != original.Cur || after.Max != original.Max {
		t.Errorf("the floor search left the limit at {cur %d max %d}, want {cur %d max %d}",
			after.Cur, after.Max, original.Cur, original.Max)
	}
}

// With nothing above the refused value there is no bracket to search, and the
// function must say so rather than return a number it did not measure.
func TestMeasureSoftLimitFloorWithoutABracket(t *testing.T) {
	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		t.Skipf("cannot read RLIMIT_NOFILE: %v", err)
	}
	if got := measureSoftLimitFloor(syscall.RLIMIT_NOFILE, original.Max, original); got != -1 {
		t.Errorf("measureSoftLimitFloor with no bracket = %d, want -1", got)
	}
}

func TestClampToInt64(t *testing.T) {
	const maxInt64 = uint64(1)<<63 - 1
	for in, want := range map[uint64]int64{
		0:            0,
		4096:         4096,
		maxInt64:     int64(maxInt64),
		maxInt64 + 1: int64(maxInt64),
	} {
		if got := clampToInt64(in); got != want {
			t.Errorf("clampToInt64(%d) = %d, want %d", in, got, want)
		}
	}
}

// ------------------------------------------------- the agent as a subprocess

// runBoundAgent invokes one bound operation of the built agent and returns what
// it reported.
func runBoundAgent(t *testing.T, op string, env []string) Report {
	t.Helper()
	cmd := exec.Command(buildProbeBinary(t), "__inside", op)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		t.Fatalf("agent %s produced no report: %v", op, err)
	}
	report, parseErr := ParseReport(string(out))
	if parseErr != nil {
		t.Fatalf("agent %s report: %v (stdout %q)", op, parseErr, out)
	}
	return report
}

// A bounded stress process reports one measurement of the kind it was asked
// for, and the numbers in it are internally consistent. The verdict this host
// produces is deliberately not asserted: a host that gained an aggregate bound
// must change the observation, not turn this test red.
func TestBoundStressMeasuresTheKindItWasAskedFor(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the bounded stresses")
	}
	for _, kind := range BoundKinds() {
		t.Run(kind, func(t *testing.T) {
			declared, ceiling := boundBudget(kind)
			report := runBoundAgent(t, OpBoundStress, []string{
				EnvSelf + "=" + buildProbeBinary(t),
				EnvMarker + "=" + filepath.Join(t.TempDir(), "m"),
				EnvBoundKind + "=" + kind,
				EnvBoundDeclared + "=" + strconv.FormatInt(declared, 10),
				EnvBoundCeiling + "=" + strconv.FormatInt(ceiling, 10),
				EnvBoundInstall + "=1",
				EnvBoundNest + "=0",
			})

			if len(report.Bounds) != 1 {
				t.Fatalf("got %d measurements, want exactly 1", len(report.Bounds))
			}
			m := report.Bounds[0]
			if m.Kind != kind {
				t.Errorf("measured kind %q, want %q", m.Kind, kind)
			}
			if m.Resource != boundResourceName(kind) {
				t.Errorf("named resource %q, want %q", m.Resource, boundResourceName(kind))
			}
			if m.Declared != declared || m.Ceiling != ceiling {
				t.Errorf("reported declared %d ceiling %d, want %d and %d",
					m.Declared, m.Ceiling, declared, ceiling)
			}
			if !m.Installed && m.InstallErrno == "" {
				t.Error("a bound that was not installed does not say why")
			}
			if m.Installed && m.Bounded.Ran && m.Bounded.Reached > m.Ceiling {
				t.Errorf("the bounded run reached %d, past its own ceiling %d", m.Bounded.Reached, m.Ceiling)
			}
			if m.Bounded.Refused && m.Bounded.Refusal == "" {
				t.Error("a refused run does not name the refusal")
			}
			// A bound that could not be installed must have looked for the
			// value the kernel would accept, so the report can distinguish a
			// refused value from a broken call.
			if !m.Installed && m.SoftLimitFloor == 0 {
				t.Error("a refused bound recorded neither a floor nor -1 for it")
			}
		})
	}
}

// The unbounded control has to reach past the declared budget, otherwise no
// refusal measured against that budget means anything.
func TestBoundStressControlPassesTheDeclaredBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the unbounded stresses")
	}
	for _, kind := range BoundKinds() {
		t.Run(kind, func(t *testing.T) {
			declared, ceiling := boundBudget(kind)
			report := runBoundAgent(t, OpBoundStress, []string{
				EnvSelf + "=" + buildProbeBinary(t),
				EnvMarker + "=" + filepath.Join(t.TempDir(), "m"),
				EnvBoundKind + "=" + kind,
				EnvBoundDeclared + "=" + strconv.FormatInt(declared, 10),
				EnvBoundCeiling + "=" + strconv.FormatInt(ceiling, 10),
				EnvBoundInstall + "=0",
				EnvBoundNest + "=0",
			})
			if len(report.Bounds) != 1 {
				t.Fatalf("got %d measurements, want exactly 1", len(report.Bounds))
			}
			run := report.Bounds[0].Bounded
			if !run.Ran {
				t.Fatalf("the unbounded run produced no measurement: %+v", run)
			}
			if run.Refused {
				t.Errorf("the unbounded run was refused with no bound installed: %+v", run)
			}
			if run.Reached <= declared {
				t.Errorf("the unbounded run reached %d, which does not pass the declared budget %d",
					run.Reached, declared)
			}
		})
	}
}

// An unknown kind must be reported as one, not silently measured as something
// else.
func TestBoundStressRejectsAnUnknownKind(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the agent")
	}
	report := runBoundAgent(t, OpBoundStress, []string{
		EnvBoundKind + "=not-a-kind",
		EnvBoundDeclared + "=1",
		EnvBoundCeiling + "=2",
		EnvBoundInstall + "=1",
	})
	if len(report.Bounds) != 1 {
		t.Fatalf("got %d measurements, want exactly 1", len(report.Bounds))
	}
	if report.Bounds[0].Error == "" {
		t.Error("an unknown kind produced no error")
	}
	if report.Bounds[0].Bounded.Ran {
		t.Error("an unknown kind still claimed a measurement")
	}
}

// The escape process answers both halves of the soft-limit question, and the
// second is the matched control for the first.
func TestBoundEscapeReportsBothHalves(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the agent")
	}
	for _, kind := range BoundKinds() {
		t.Run(kind, func(t *testing.T) {
			declared, _ := boundBudget(kind)
			report := runBoundAgent(t, OpBoundEscape, []string{
				EnvBoundKind + "=" + kind,
				EnvBoundDeclared + "=" + strconv.FormatInt(declared, 10),
			})
			preserved := report.Values["soft_raise_hard_preserved"]
			lowered := report.Values["soft_raise_hard_lowered"]
			if preserved == "" || lowered == "" {
				t.Fatalf("the escape process reported %q and %q", preserved, lowered)
			}
			for _, value := range []string{preserved, lowered} {
				switch {
				case value == EscapeRaised, value == EscapeNotAttempted:
				case len(value) > 8 && value[:8] == "refused:":
				case len(value) > 12 && value[:12] == "unavailable:":
				default:
					t.Errorf("unrecognised escape result %q", value)
				}
			}
			// When the bound could be installed at all, the two halves must not
			// be identical: if lowering the hard limit changed nothing, the
			// control shows nothing.
			if preserved == EscapeRaised && lowered == EscapeRaised {
				t.Errorf("%s: the member restored its budget even with the hard limit lowered, "+
					"so the matched control distinguishes nothing", kind)
			}
		})
	}
}

// The matrix covers the whole kind list exactly once, so a bound cannot quietly
// go unmeasured while the report still looks complete.
func TestBoundMatrixCoversEveryKindExactlyOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("runs every bounded stress")
	}
	dir := t.TempDir()
	report := runBoundAgent(t, OpBoundMatrix, []string{
		EnvSelf + "=" + buildProbeBinary(t),
		EnvMarker + "=" + filepath.Join(dir, "m"),
	})

	seen := map[string]int{}
	for _, m := range report.Bounds {
		seen[m.Kind]++
	}
	for _, kind := range BoundKinds() {
		switch seen[kind] {
		case 1:
		case 0:
			t.Errorf("the matrix produced no measurement for %s", kind)
		default:
			t.Errorf("the matrix produced %d measurements for %s", seen[kind], kind)
		}
	}
	if len(report.Bounds) != len(BoundKinds()) {
		t.Errorf("the matrix produced %d measurements for %d kinds",
			len(report.Bounds), len(BoundKinds()))
	}
	// The matrix reports the same order every run, so two artifacts from the
	// same host can be compared line by line.
	for i, kind := range BoundKinds() {
		if report.Bounds[i].Kind != kind {
			t.Errorf("measurement %d is %s, want %s", i, report.Bounds[i].Kind, kind)
		}
	}

	// Every measurement carries its control, which is what makes the matrix
	// evidence rather than a list of refusals.
	for _, m := range report.Bounds {
		if !m.Control.Ran {
			t.Errorf("%s has no unbounded control: %+v", m.Kind, m.Control)
		}
	}
}

// Without an agent to start, the matrix cannot measure anything and must say so
// rather than report an empty set of bounds as a clean result.
func TestBoundMatrixWithoutAnAgent(t *testing.T) {
	report := Report{Values: map[string]string{}}
	unsetEnv(t, EnvSelf)
	attempts := boundMatrix(&report)
	if len(report.Bounds) != 0 {
		t.Errorf("with no agent, the matrix still produced %d measurements", len(report.Bounds))
	}
	if len(attempts) != 1 || attempts[0].Outcome != OutcomeInconclusive {
		t.Errorf("with no agent, the matrix reported %+v", attempts)
	}
}

// The bound processes must not survive the matrix that started them. The whole
// prototype is a statement about descendant control; a harness that leaked its
// own descendants would be arguing against itself.
func TestBoundMatrixLeavesNoDescendantBehind(t *testing.T) {
	if testing.Short() {
		t.Skip("runs every bounded stress")
	}
	dir := t.TempDir()
	runBoundAgent(t, OpBoundMatrix, []string{
		EnvSelf + "=" + buildProbeBinary(t),
		EnvMarker + "=" + filepath.Join(dir, "m"),
	})

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read the marker directory: %v", err)
	}
	for _, entry := range entries {
		pid := markerPID(t, filepath.Join(dir, entry.Name()))
		if pid > 0 && syscall.Kill(pid, 0) == nil {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Errorf("%s named pid %d, which is still running after the matrix returned",
				entry.Name(), pid)
		}
	}
}

// The measurement encodes and decodes without loss, because the probe layer
// reads it out of the agent's stdout rather than out of this struct.
func TestBoundMeasurementRoundTrips(t *testing.T) {
	want := BoundMeasurement{
		Kind: BoundCPU, Resource: "RLIMIT_CPU", Declared: 1000, Ceiling: 1600,
		Installed: true, SoftLimitFloor: -1, InheritedSoft: 1, InheritedHard: 2,
		Bounded:                BoundRun{Ran: true, Reached: 1001, Refused: true, Refusal: "SIGXCPU"},
		Nested:                 BoundRun{Ran: true, Reached: 1002, Refused: true, Refusal: "SIGXCPU"},
		Control:                BoundRun{Ran: true, Reached: 1600},
		SoftRaiseHardPreserved: EscapeRaised,
		SoftRaiseHardLowered:   "refused:EPERM",
	}
	data, err := json.Marshal([]BoundMeasurement{want})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got []BoundMeasurement
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("round trip changed the measurement:\n got %+v\nwant %+v", got, want)
	}
}

// ------------------------------------------------- the orchestration itself

// The orchestration is exercised in this process, not through the built
// binary, because a subprocess measures the host but reports nothing about
// which branches of the harness ran. Only the runs that install no limit are
// safe here: a bounded run installs a real limit on the process that hosts it,
// and lowering a hard limit is irreversible, so those keep running as
// subprocesses of their own.

func TestBoundStressInProcessWithNoBoundInstalled(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the unbounded stress and one descendant")
	}
	declared, ceiling := boundBudget(BoundCPU)
	t.Setenv(EnvSelf, buildProbeBinary(t))
	t.Setenv(EnvMarker, filepath.Join(t.TempDir(), "m"))
	t.Setenv(EnvBoundKind, BoundCPU)
	t.Setenv(EnvBoundDeclared, strconv.FormatInt(declared, 10))
	t.Setenv(EnvBoundCeiling, strconv.FormatInt(ceiling, 10))
	t.Setenv(EnvBoundInstall, "0")
	t.Setenv(EnvBoundNest, "1")

	report := Report{Values: map[string]string{}}
	attempts := boundStress(&report)

	if len(report.Bounds) != 1 {
		t.Fatalf("got %d measurements, want exactly 1", len(report.Bounds))
	}
	m := report.Bounds[0]
	if m.Installed {
		t.Error("a control run reported that it installed a bound")
	}
	if !m.Bounded.Ran || m.Bounded.Refused {
		t.Errorf("the unbounded run did not complete cleanly: %+v", m.Bounded)
	}
	if !m.Nested.Ran {
		t.Errorf("the descendant produced no measurement: %+v", m.Nested)
	}
	// The inherited limits have to be reported, so a reader can tell a bound
	// already in force from one this run installed.
	if m.InheritedSoft == -1 || m.InheritedHard == -1 {
		t.Errorf("the run did not report the limits it inherited: soft=%d hard=%d",
			m.InheritedSoft, m.InheritedHard)
	}

	byName := map[string]Attempt{}
	for _, a := range attempts {
		byName[a.Name] = a
	}
	install, ok := byName["install-bound"]
	if !ok || install.Outcome != OutcomeInconclusive {
		t.Errorf("the control run recorded install-bound as %+v, want an inconclusive non-attempt", install)
	}
	if _, ok := byName["stress:"+BoundCPU]; !ok {
		t.Errorf("no stress attempt was recorded; got %v", attempts)
	}
	if _, ok := byName["nested-descendant-stress:"+BoundCPU]; !ok {
		t.Errorf("no nested-descendant attempt was recorded; got %v", attempts)
	}
}

// An unknown kind must not be silently measured as something else, and the
// in-process path has to say so the same way the subprocess one does.
func TestBoundStressInProcessRejectsAnUnknownKind(t *testing.T) {
	t.Setenv(EnvBoundKind, "not-a-kind")
	t.Setenv(EnvBoundDeclared, "1")
	t.Setenv(EnvBoundCeiling, "2")
	t.Setenv(EnvBoundInstall, "0")

	report := Report{Values: map[string]string{}}
	attempts := boundStress(&report)
	if len(report.Bounds) != 1 || report.Bounds[0].Error == "" {
		t.Fatalf("an unknown kind produced %+v", report.Bounds)
	}
	if len(attempts) != 1 || attempts[0].Outcome != OutcomeInconclusive {
		t.Errorf("an unknown kind reported %+v", attempts)
	}
}

// measureBound is the part that merges three processes into one measurement.
// It installs nothing in this process; every bound it declares is installed by
// a child.
func TestMeasureBoundMergesItsThreeProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("runs three processes per bound")
	}
	t.Setenv(EnvSelf, buildProbeBinary(t))
	t.Setenv(EnvMarker, filepath.Join(t.TempDir(), "m"))

	m := measureBound(buildProbeBinary(t), BoundProcessCount)
	if m.Error != "" {
		t.Fatalf("measureBound reported %q", m.Error)
	}
	if m.Kind != BoundProcessCount {
		t.Errorf("measured kind %q", m.Kind)
	}
	if !m.Control.Ran {
		t.Errorf("the control did not run: %+v", m.Control)
	}
	if m.SoftRaiseHardPreserved == "" || m.SoftRaiseHardLowered == "" {
		t.Errorf("the escape process contributed nothing: %q / %q",
			m.SoftRaiseHardPreserved, m.SoftRaiseHardLowered)
	}
	if measurementOutcome(m) != OutcomeAllowed {
		t.Errorf("a complete measurement was reported as %q", measurementOutcome(m))
	}
	// This process must be unchanged: every limit was installed in a child.
	var after syscall.Rlimit
	if err := syscall.Getrlimit(rlimitNPROC, &after); err != nil {
		t.Fatalf("getrlimit: %v", err)
	}
	if int64(after.Cur) <= m.Declared {
		t.Errorf("measureBound left this process at a soft process limit of %d", after.Cur)
	}
}

// A measurement whose control never ran cannot be reduced, and must be reported
// as inconclusive rather than as a clean set of refusals.
func TestMeasurementOutcomeNeedsAControl(t *testing.T) {
	if got := measurementOutcome(BoundMeasurement{Control: BoundRun{Ran: false}}); got != OutcomeInconclusive {
		t.Errorf("a measurement with no control = %q, want inconclusive", got)
	}
	if got := measurementOutcome(BoundMeasurement{Error: "boom", Control: BoundRun{Ran: true}}); got != OutcomeInconclusive {
		t.Errorf("a measurement carrying an error = %q, want inconclusive", got)
	}
}

func TestBoundMatrixInProcessMeasuresEveryKind(t *testing.T) {
	if testing.Short() {
		t.Skip("runs every bounded stress")
	}
	t.Setenv(EnvSelf, buildProbeBinary(t))
	t.Setenv(EnvMarker, filepath.Join(t.TempDir(), "m"))

	report := Report{Values: map[string]string{}}
	attempts := boundMatrix(&report)

	if len(report.Bounds) != len(BoundKinds()) {
		t.Fatalf("the matrix produced %d measurements for %d kinds",
			len(report.Bounds), len(BoundKinds()))
	}
	if len(attempts) != len(BoundKinds()) {
		t.Errorf("the matrix recorded %d attempts for %d kinds", len(attempts), len(BoundKinds()))
	}
	for i, kind := range BoundKinds() {
		if report.Bounds[i].Kind != kind {
			t.Errorf("measurement %d is %s, want %s", i, report.Bounds[i].Kind, kind)
		}
		if attempts[i].Name != "bound-measured:"+kind {
			t.Errorf("attempt %d is %q", i, attempts[i].Name)
		}
	}
}

// ---------------------------------------------------------- small vocabulary

func TestRunOutcomeDistinguishesTheThreeCases(t *testing.T) {
	for name, tc := range map[string]struct {
		run  BoundRun
		want string
	}{
		"no measurement": {BoundRun{Ran: false}, OutcomeInconclusive},
		"refused":        {BoundRun{Ran: true, Refused: true}, OutcomeDenied},
		"passed":         {BoundRun{Ran: true}, OutcomeAllowed},
	} {
		if got := runOutcome(tc.run); got != tc.want {
			t.Errorf("%s: runOutcome = %q, want %q", name, got, tc.want)
		}
	}
}

func TestStressRejectsAnUnknownKind(t *testing.T) {
	run := stress("not-a-kind", 1)
	if run.Ran {
		t.Errorf("an unknown kind produced a measurement: %+v", run)
	}
}

func TestBoundBudgetOfAnUnknownKindIsZero(t *testing.T) {
	if declared, ceiling := boundBudget("not-a-kind"); declared != 0 || ceiling != 0 {
		t.Errorf("boundBudget(unknown) = (%d, %d), want (0, 0)", declared, ceiling)
	}
}

func TestSignalNameCoversTheSignalsABoundCanDeliver(t *testing.T) {
	for sig, want := range map[syscall.Signal]string{
		syscall.SIGXCPU: "SIGXCPU",
		syscall.SIGKILL: "SIGKILL",
		syscall.SIGSEGV: "SIGSEGV",
		syscall.SIGBUS:  "SIGBUS",
		syscall.SIGABRT: "SIGABRT",
	} {
		if got := signalName(sig); got != want {
			t.Errorf("signalName(%d) = %q, want %q", sig, got, want)
		}
	}
	if got := signalName(syscall.SIGUSR2); got != "signal(31)" && got != "signal(12)" {
		t.Errorf("signalName(SIGUSR2) = %q, want a numbered fallback", got)
	}
}

// A process that exited normally was not killed by its bound, and deathSignal
// must not invent a signal for it.
func TestDeathSignalOnlyReportsARealSignal(t *testing.T) {
	if _, ok := deathSignal(nil); ok {
		t.Error("deathSignal reported a signal for a clean exit")
	}
	if _, ok := deathSignal(errFakeInside); ok {
		t.Error("deathSignal reported a signal for a non-exit error")
	}
	killed := exec.Command(buildProbeBinary(t), "__inside", OpDetachedChild)
	killed.Env = append(os.Environ(), EnvMarker+"="+filepath.Join(t.TempDir(), "m"), EnvHold+"=30")
	pipeStdio(killed)
	if err := killed.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	_ = killed.Process.Kill()
	name, ok := deathSignal(killed.Wait())
	if !ok || name != "SIGKILL" {
		t.Errorf("deathSignal of a killed process = (%q, %v), want SIGKILL", name, ok)
	}
}

func TestEscapeRefusedNamesTheErrno(t *testing.T) {
	if got := escapeRefused(syscall.EPERM); got != "refused:EPERM" {
		t.Errorf("escapeRefused(EPERM) = %q", got)
	}
	if got := escapeRefused(errFakeInside); got != "refused:"+errFakeInside.Error() {
		t.Errorf("escapeRefused(other) = %q", got)
	}
}

func TestErrnoOfExtractsOnlyErrnos(t *testing.T) {
	if got := errnoOf(nil); got != "" {
		t.Errorf("errnoOf(nil) = %q", got)
	}
	if got := errnoOf(syscall.EAGAIN); got != "EAGAIN" {
		t.Errorf("errnoOf(EAGAIN) = %q", got)
	}
	if got := errnoOf(errFakeInside); got != "" {
		t.Errorf("errnoOf(non-errno) = %q", got)
	}
}

// A malformed or negative value must fall back rather than turn a budget into
// something the stress would interpret as unbounded.
func TestEnvInt64FallsBackOnUnusableValues(t *testing.T) {
	const key = "PROBE_TEST_INT"
	for value, want := range map[string]int64{
		"":     7,
		"x":    7,
		"-1":   7,
		"0":    0,
		"1024": 1024,
	} {
		t.Setenv(key, value)
		if value == "" {
			unsetEnv(t, key)
		}
		if got := envInt64(key, 7); got != want {
			t.Errorf("envInt64(%q) = %d, want %d", value, got, want)
		}
	}
}

func TestBoolEnvAndAppendError(t *testing.T) {
	if boolEnv(true) != "1" || boolEnv(false) != "0" {
		t.Error("boolEnv does not render 1 and 0")
	}
	if got := appendError("", "first"); got != "first" {
		t.Errorf("appendError onto empty = %q", got)
	}
	if got := appendError("first", "second"); got != "first; second" {
		t.Errorf("appendError = %q", got)
	}
}

var errFakeInside = fakeInsideError("not a syscall error")

type fakeInsideError string

func (e fakeInsideError) Error() string { return string(e) }

// The escape measurement is exercised in this process against RLIMIT_CORE.
//
// It cannot be run here against any of the four real bounds: its last step
// lowers a hard limit, which POSIX makes irreversible, and a test binary that
// capped its own CPU seconds or process count would take the rest of the suite
// down with it. RLIMIT_CORE has the same soft/hard semantics and bounds only
// the size of a core dump this process is never going to write, so the logic is
// measured for real while the consequence is nothing.
func TestEscapeAttemptsMeasuresBothHalves(t *testing.T) {
	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_CORE, &original); err != nil {
		t.Skipf("cannot read RLIMIT_CORE: %v", err)
	}
	if original.Max == 0 {
		t.Skip("the core-dump hard limit is already zero, so there is no room to lower it")
	}

	report := Report{Values: map[string]string{}}
	attempts := escapeAttempts(syscall.RLIMIT_CORE, "RLIMIT_CORE", 0, &report)

	byName := map[string]Attempt{}
	for _, a := range attempts {
		byName[a.Name] = a
	}
	if a, ok := byName["lower-soft-limit-only"]; !ok || a.Outcome != OutcomeAllowed {
		t.Fatalf("the soft limit could not be lowered: %+v", a)
	}
	if _, ok := byName["raise-soft-limit-when-hard-preserved"]; !ok {
		t.Error("the escape attempt was not recorded")
	}
	if a, ok := byName["lower-hard-limit"]; !ok || a.Outcome != OutcomeAllowed {
		t.Fatalf("the hard limit could not be lowered, so the matched control did not run: %+v", a)
	}
	if _, ok := byName["raise-soft-limit-when-hard-lowered"]; !ok {
		t.Error("the matched control was not recorded")
	}

	// With the hard limit preserved a member can restore its own budget; with
	// the hard limit lowered it cannot. That difference is the whole finding,
	// and if it ever stops holding the escape control proves nothing.
	if report.Values["soft_raise_hard_preserved"] != EscapeRaised {
		t.Errorf("raising the soft limit under a preserved hard limit was %q, want %q",
			report.Values["soft_raise_hard_preserved"], EscapeRaised)
	}
	if got := report.Values["soft_raise_hard_lowered"]; got == EscapeRaised {
		t.Error("the soft limit was restored even with the hard limit lowered, " +
			"so lowering the hard limit distinguishes nothing")
	}
	if report.Values["inherited_soft"] == "" || report.Values["inherited_hard"] == "" {
		t.Error("the escape measurement did not report the limits it inherited")
	}
}

// A kind with no resource behind it must be reported as unmeasurable rather
// than measured against whatever resource number zero happens to be.
func TestBoundEscapeRejectsAnUnknownKind(t *testing.T) {
	t.Setenv(EnvBoundKind, "not-a-kind")
	report := Report{Values: map[string]string{}}
	attempts := boundEscape(&report)

	if len(attempts) != 1 || attempts[0].Outcome != OutcomeInconclusive {
		t.Errorf("an unknown kind produced %+v", attempts)
	}
	for _, key := range []string{"soft_raise_hard_preserved", "soft_raise_hard_lowered"} {
		if report.Values[key] != EscapeNotAttempted {
			t.Errorf("%s = %q, want %q", key, report.Values[key], EscapeNotAttempted)
		}
	}
}
