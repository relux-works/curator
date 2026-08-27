package inside

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

// These tests exercise the descriptor-budget probe in this process rather than
// in a subprocess, because that is the only way to observe the code the harness
// actually runs. They lower RLIMIT_NOFILE and then open descriptors until the
// kernel refuses, so each one restores the original limit before returning and
// none of them may run in parallel with anything that opens files.

// withRestoredNofile snapshots RLIMIT_NOFILE and puts it back afterwards, so a
// probe that leaves the limit lowered cannot break every later test.
func withRestoredNofile(t *testing.T) {
	t.Helper()
	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		t.Skipf("cannot read RLIMIT_NOFILE: %v", err)
	}
	if original.Cur < 128 {
		t.Skipf("the soft descriptor limit is already %d; lowering it further would starve the test binary", original.Cur)
	}
	t.Cleanup(func() {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
			t.Errorf("could not restore RLIMIT_NOFILE to %d: %v", original.Cur, err)
		}
	})
}

func TestCountOpenableStopsAtTheSoftLimit(t *testing.T) {
	withRestoredNofile(t)

	var original syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &original); err != nil {
		t.Fatalf("getrlimit: %v", err)
	}
	const soft = 96
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: soft, Max: original.Max}); err != nil {
		t.Skipf("cannot lower RLIMIT_NOFILE on this host: %v", err)
	}

	count := countOpenable()
	if count <= 0 {
		t.Fatalf("countOpenable = %d; it opened nothing under a soft limit of %d", count, soft)
	}
	if count >= soft {
		t.Errorf("countOpenable = %d under a soft limit of %d; the limit was not reached", count, soft)
	}

	// Every descriptor it opened must be closed again: a probe that leaked
	// would poison the run it is measuring.
	second := countOpenable()
	if second < count-8 {
		t.Errorf("a second count got %d after the first got %d; the first leaked descriptors", second, count)
	}
}

func TestCountForHelloCountsWhenAsked(t *testing.T) {
	withRestoredNofile(t)
	t.Setenv("PROBE_COUNT_DESCRIPTORS", "96")

	count, ok := countForHello()
	if !ok {
		t.Fatal("countForHello did not count when asked")
	}
	if count <= 0 || count >= 96 {
		t.Errorf("countForHello = %d, want a positive count below the requested 96", count)
	}
}

func TestChildDescriptorCount(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns the built agent")
	}
	withRestoredNofile(t)

	count := childDescriptorCount(buildProbeBinary(t), 96)
	if count <= 0 {
		t.Errorf("childDescriptorCount = %d, want a positive count", count)
	}
	if count >= 96 {
		t.Errorf("childDescriptorCount = %d, want a count below the requested soft limit", count)
	}
}

// A child that cannot be started, or that answers with something other than a
// report, must produce "unknown" rather than a number the class probe would
// then reason about.
func TestChildDescriptorCountRejectsUnusableChildren(t *testing.T) {
	cases := map[string]string{
		"missing program":  filepath.Join(t.TempDir(), "not-a-program"),
		"not a probe":      "/bin/echo",
		"exits nonzero":    "/usr/bin/false",
		"empty agent path": "",
	}
	for name, program := range cases {
		t.Run(name, func(t *testing.T) {
			if got := childDescriptorCount(program, 96); got != -1 {
				t.Errorf("childDescriptorCount(%q) = %d, want -1", program, got)
			}
		})
	}
}

// The finding this probe exists to record: a descriptor budget on macOS is per
// process, so a parent and its child each get the full allowance and the two
// add up past it. Nothing here is contained, so the result is a property of the
// host and not of any profile.
func TestResourceBoundAttemptsFindsPerProcessBudgets(t *testing.T) {
	if testing.Short() {
		t.Skip("opens descriptors and writes megabytes")
	}
	withRestoredNofile(t)

	buildRoot := t.TempDir()
	const soft = 96
	const budget = 1 << 20

	t.Setenv(EnvNoFileCap, strconv.Itoa(soft))
	t.Setenv(EnvWriteBytes, strconv.Itoa(budget))
	t.Setenv(EnvBuildRoot, buildRoot)
	t.Setenv(EnvSelf, buildProbeBinary(t))

	report := Report{Values: map[string]string{}}
	attempts := resourceBoundAttempts(&report)

	byName := map[string]Attempt{}
	for _, a := range attempts {
		byName[a.Name] = a
	}
	if a, ok := byName["setrlimit-nofile"]; !ok || a.Outcome != OutcomeAllowed {
		t.Fatalf("setrlimit-nofile = %+v; the soft limit was never established", a)
	}

	self, err := strconv.Atoi(report.Values["self_descriptors"])
	if err != nil {
		t.Fatalf("self_descriptors %q: %v", report.Values["self_descriptors"], err)
	}
	child, err := strconv.Atoi(report.Values["child_descriptors"])
	if err != nil {
		t.Fatalf("child_descriptors %q: %v", report.Values["child_descriptors"], err)
	}
	if self <= 0 || child <= 0 {
		t.Fatalf("descriptor counts self=%d child=%d; the aggregate claim cannot be evaluated", self, child)
	}
	if self+child <= soft {
		t.Errorf("parent %d + child %d = %d under a soft limit of %d: this host appears to "+
			"account descriptors in aggregate, so the class verdict must be re-measured",
			self, child, self+child, soft)
	}
	if a, ok := byName["descriptor-budget-is-per-process"]; !ok || a.Outcome != OutcomeAllowed {
		t.Errorf("descriptor-budget-is-per-process = %+v, want an allowed observation", a)
	}

	written, err := strconv.ParseInt(report.Values["bytes_written"], 10, 64)
	if err != nil {
		t.Fatalf("bytes_written %q: %v", report.Values["bytes_written"], err)
	}
	if report.Values["byte_budget"] != strconv.Itoa(budget) {
		t.Errorf("byte_budget %q, want %d", report.Values["byte_budget"], budget)
	}
	if written <= budget {
		t.Errorf("wrote %d bytes against a declared budget of %d; a host that enforced it "+
			"would have refused earlier", written, budget)
	}
	if a, ok := byName["write-past-declared-byte-budget"]; ok && a.Outcome == OutcomeDenied {
		t.Errorf("the byte budget was enforced with no mechanism applied: %+v", a)
	}
	// The probe must leave nothing behind in the build root it wrote through.
	entries, err := os.ReadDir(buildRoot)
	if err != nil {
		t.Fatalf("read build root: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == "byte-budget.bin" {
			t.Errorf("the probe left its %d-byte scratch file behind", written)
		}
	}

	// The probe restores the soft limit it lowered; otherwise every later probe
	// in the same agent would run under a limit it never asked for.
	var after syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &after); err != nil {
		t.Fatalf("getrlimit: %v", err)
	}
	if after.Cur == soft {
		t.Errorf("resourceBoundAttempts left the soft limit at %d", after.Cur)
	}
}

// Without an agent to spawn there is no second budget to compare against, so
// the probe must report that it could not evaluate aggregation.
func TestResourceBoundAttemptsWithoutAnAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("opens descriptors")
	}
	withRestoredNofile(t)

	t.Setenv(EnvNoFileCap, "96")
	unsetEnv(t, EnvSelf)
	unsetEnv(t, EnvBuildRoot)

	report := Report{Values: map[string]string{}}
	byName := map[string]Attempt{}
	for _, a := range resourceBoundAttempts(&report) {
		byName[a.Name] = a
	}

	a, ok := byName["descriptor-budget-is-per-process"]
	if !ok {
		t.Fatal("no descriptor-budget-is-per-process attempt")
	}
	if a.Outcome != OutcomeInconclusive {
		t.Errorf("with no agent to spawn, outcome = %q, want inconclusive", a.Outcome)
	}
	if report.Values["child_descriptors"] != "-1" {
		t.Errorf("child_descriptors = %q, want -1", report.Values["child_descriptors"])
	}
}
