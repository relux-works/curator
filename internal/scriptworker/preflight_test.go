package scriptworker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/scriptpolicy"
)

// errInjectedProbeFailureForTest fails one native-control probe through the
// production probe path.
var errInjectedProbeFailureForTest = errors.New("injected probe failure")

// TestPreflightRefusesUnavailableControlAtInvocation drives vector case
// `mandatory-control-unavailable-at-invocation` through the production
// entry: with a mandatory control unavailable the launch refuses with
// `script_execution_control_unavailable` and starts no worker. The forge
// binary stands in for the manager so a started worker would record its
// frames; no marker may appear.
func TestPreflightRefusesUnavailableControlAtInvocation(t *testing.T) {
	// Sequential by construction: the test forces the process-global probe
	// fault. Never add t.Parallel here.
	defer OverrideScriptProbeFaultForTest(func(control string) error {
		if control == ScriptControlDescendantDomainTermination {
			return errInjectedProbeFailureForTest
		}
		return nil
	})()
	forge := mustPhysical(t, forgeWorkerBinary(t))
	fixture := newLaunchFixture(t, forge)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := Launch(ctx, fixture.request)
	if DiagnosticCode(err) != scriptpolicy.ControlUnavailable {
		t.Fatalf("Launch error = %v, want %s", err, scriptpolicy.ControlUnavailable)
	}
	if !strings.Contains(err.Error(), ScriptControlDescendantDomainTermination) {
		t.Fatalf("refusal does not name the unavailable control: %v", err)
	}
	for _, marker := range []string{".forge-permit", ".forge-shutdown", ".forge-request"} {
		if _, statErr := os.Stat(fixture.entry + marker); statErr == nil {
			t.Fatalf("the worker started despite the preflight refusal (marker %s exists)", marker)
		}
	}
}

// fixtureCgroupFS builds a delegated-cgroup-v2 fixture hierarchy: the own
// membership `Delegated` offers exactly the named controllers. Tests point
// the parent's cgroup seam at it to drive the Linux aggregate-limit file
// protocol deterministically on any unix host; the kernel-delegation
// reality of any particular host is proven separately by
// TestLinuxHostProbeAndEvidenceAreConsistent on ubuntu-latest.
func fixtureCgroupFS(t *testing.T, controllers ...string) scriptCgroupFS {
	t.Helper()
	root := mustPhysical(t, t.TempDir())
	own := filepath.Join(root, "Delegated")
	if err := os.MkdirAll(own, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(own, "cgroup.controllers"), []byte(strings.Join(controllers, " ")+"\n"), 0o644)
	writeTestFile(t, filepath.Join(own, "cgroup.subtree_control"), []byte("\n"), 0o644)
	writeTestFile(t, filepath.Join(own, "cgroup.procs"), []byte(""), 0o644)
	self := filepath.Join(root, "proc-self-cgroup")
	writeTestFile(t, self, []byte("0::/Delegated\n"), 0o644)
	return scriptCgroupFS{Root: root, SelfPath: self}
}

func evidenceStatus(t *testing.T, record ScriptEvidence, control string) string {
	t.Helper()
	for _, entry := range record.Controls {
		if entry.Name == control {
			return entry.Status
		}
	}
	t.Fatalf("the record carries no entry for %q", control)
	return ""
}

// TestLinuxPidsMaxProbeAvailableApplies drives vector case
// `linux-pids-max-probe-available-evidence-applied-invocation-succeeds`
// through the production session: with a delegated fixture hierarchy the
// Linux probe finds the aggregate controls present, the session applies
// them, the evidence reports them applied, and the invocation succeeds.
func TestLinuxPidsMaxProbeAvailableApplies(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the delegated-cgroup fixture is a unix-only script-inventory fixture in script-worker-v1-native-control-inventory-v1")
	}
	// Sequential by construction: the test forces the process-global
	// inventory platform and cgroup roots. Never add t.Parallel here.
	defer OverrideInventoryPlatformForTest(ScriptPlatformLinux)()
	defer OverrideCgroupFSForTest(fixtureCgroupFS(t, "pids", "memory"))()
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	if evidenceStatus(t, result.Evidence, ScriptControlActiveProcessCountLimit) != ScriptStatusApplied {
		t.Fatalf("pids evidence status = %q, want applied",
			evidenceStatus(t, result.Evidence, ScriptControlActiveProcessCountLimit))
	}
	if evidenceStatus(t, result.Evidence, ScriptControlAggregateMemoryLimit) != ScriptStatusApplied {
		t.Fatalf("memory evidence status = %q, want applied",
			evidenceStatus(t, result.Evidence, ScriptControlAggregateMemoryLimit))
	}
	assertValidScriptEvidence(t, result.Evidence, ScriptPlatformLinux)
}

// TestLinuxPidsMaxProbeUnavailableSucceeds drives vector case
// `linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds`
// through the production session: with a fixture hierarchy that offers no
// pids controller the Linux probe reports the control absent, the evidence
// reports it unavailable, and the invocation still succeeds.
func TestLinuxPidsMaxProbeUnavailableSucceeds(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the delegated-cgroup fixture is a unix-only script-inventory fixture in script-worker-v1-native-control-inventory-v1")
	}
	defer OverrideInventoryPlatformForTest(ScriptPlatformLinux)()
	defer OverrideCgroupFSForTest(fixtureCgroupFS(t, "memory"))()
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	if evidenceStatus(t, result.Evidence, ScriptControlActiveProcessCountLimit) != ScriptStatusUnavailable {
		t.Fatalf("pids evidence status = %q, want unavailable",
			evidenceStatus(t, result.Evidence, ScriptControlActiveProcessCountLimit))
	}
	if evidenceStatus(t, result.Evidence, ScriptControlAggregateMemoryLimit) != ScriptStatusApplied {
		t.Fatalf("memory evidence status = %q, want applied",
			evidenceStatus(t, result.Evidence, ScriptControlAggregateMemoryLimit))
	}
	assertValidScriptEvidence(t, result.Evidence, ScriptPlatformLinux)
}

// TestLinuxHostConditionalUnavailableSucceeds drives vector case
// `valid-linux-host-conditional-unavailable` on the true Linux host: the
// real probe determines every host-conditional control, absent controls
// are reported unavailable, and the invocation succeeds.
func TestLinuxHostConditionalUnavailableSucceeds(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("host-conditional Linux probing is a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	assertValidScriptEvidence(t, result.Evidence, ScriptPlatformLinux)
	// The case's precondition is an absent host-conditional control; the
	// row pins the probe↔evidence consistency for every host-conditional
	// control instead of assuming absence, so it stays green on hosts
	// that delegate and proves the unavailable branch wherever the probe
	// finds one.
	_, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	for _, probe := range probes {
		if probe.Availability != ScriptAvailabilityHostConditional {
			continue
		}
		want := ScriptStatusUnavailable
		if probe.Present {
			want = ScriptStatusApplied
		}
		if got := evidenceStatus(t, result.Evidence, probe.Name); got != want {
			t.Fatalf("control %q evidence status = %q, want %q for this host's probe", probe.Name, got, want)
		}
	}
}

// TestLinuxHostProbeAndEvidenceAreConsistent proves the true Linux probe
// is deterministic and the session honors it: two consecutive real probes
// agree, and a fresh invocation's record validates against them.
func TestLinuxHostProbeAndEvidenceAreConsistent(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("host-conditional Linux probing is a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	firstPlatform, first, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	secondPlatform, second, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	if firstPlatform != secondPlatform || len(first) != len(second) {
		t.Fatal("two consecutive host probes disagree on platform or shape")
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("two consecutive host probes disagree on %+v vs %+v", first[index], second[index])
		}
	}
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	if err := validateScriptEvidence(result.Evidence, firstPlatform, first); err != nil {
		t.Fatalf("the invocation record does not validate against the host probe: %v", err)
	}
}

// TestFixedUnavailableControlDoesNotReject drives vector case
// `fixed-unavailable-control-does-not-reject` through the production
// session: a fixed-unavailable control is reported unavailable and the
// invocation succeeds. Every platform contributes its own
// fixed-unavailable control; Linux, which defines none, drives the macOS
// case through the platform override.
func TestFixedUnavailableControlDoesNotReject(t *testing.T) {
	if ScriptInventoryPlatform(runtime.GOOS) == ScriptPlatformLinux {
		// Sequential by construction when forcing. Never add t.Parallel
		// here.
		defer OverrideInventoryPlatformForTest(ScriptPlatformMacOS)()
	}
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	platform := ScriptInventoryPlatform(runtime.GOOS)
	if testInventoryPlatform != "" {
		platform = testInventoryPlatform
	}
	assertValidScriptEvidence(t, result.Evidence, platform)
	fixed := 0
	for _, entry := range result.Evidence.Controls {
		if entry.Availability != ScriptAvailabilityUnavailable {
			continue
		}
		fixed++
		if entry.Status != ScriptStatusUnavailable {
			t.Fatalf("fixed-unavailable control %q reports status %q", entry.Name, entry.Status)
		}
	}
	if fixed == 0 {
		t.Fatalf("no fixed-unavailable control was driven on platform %q", platform)
	}
}

// TestWindowsJobLimitsAppliedAndConfirmed proves the Windows inventory at
// the production entry: the job-backed controls probe present, install on
// the private job, confirm by query in the worker, and report applied,
// while the fixed-unavailable file-size control reports unavailable and
// the invocation succeeds.
func TestWindowsJobLimitsAppliedAndConfirmed(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Job Object limits are a Windows-only control in script-worker-v1-native-control-inventory-v1")
	}
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	assertValidScriptEvidence(t, result.Evidence, ScriptPlatformWindows)
	for _, control := range []string{
		ScriptControlDescendantDomainTermination,
		ScriptControlActiveProcessCountLimit,
		ScriptControlAggregateMemoryLimit,
		ScriptControlInheritedHandleRestriction,
	} {
		if got := evidenceStatus(t, result.Evidence, control); got != ScriptStatusApplied {
			t.Fatalf("control %q evidence status = %q, want applied", control, got)
		}
	}
	if got := evidenceStatus(t, result.Evidence, ScriptControlPerFileSizeLimit); got != ScriptStatusUnavailable {
		t.Fatalf("per-file-size-limit evidence status = %q, want unavailable", got)
	}
}

// TestMacOSNativeControlsAppliedAtInvocation proves the real macOS worker
// inventory on its native lane: process-group teardown, RLIMIT_FSIZE, and
// inherited-handle restriction apply; the remaining native mechanisms are
// reported unavailable and do not claim enforcement.
func TestMacOSNativeControlsAppliedAtInvocation(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform-control is exercised on the macOS runner")
	}
	fixture := newLaunchFixture(t, os.Args[0])
	result, _ := runDerivedSession(t, fixture)
	assertValidScriptEvidence(t, result.Evidence, ScriptPlatformMacOS)
	for _, control := range []string{
		ScriptControlDescendantDomainTermination,
		ScriptControlPerFileSizeLimit,
		ScriptControlInheritedHandleRestriction,
	} {
		if got := evidenceStatus(t, result.Evidence, control); got != ScriptStatusApplied {
			t.Fatalf("macOS control %q status = %q, want applied", control, got)
		}
	}
	for _, control := range []string{
		ScriptControlActiveProcessCountLimit,
		ScriptControlAggregateMemoryLimit,
		ScriptControlDescendantExecDenial,
		ScriptControlFilesystemWriteConfinement,
		ScriptControlNetworkIsolationDomain,
	} {
		if got := evidenceStatus(t, result.Evidence, control); got != ScriptStatusUnavailable {
			t.Fatalf("macOS control %q status = %q, want unavailable", control, got)
		}
	}
}

// TestLinuxLandlockConfinementMatchesProbe proves the Landlock controls
// behaviorally at the production entry: with a `repo` filesystem the
// worker grants exactly the derived project root, so the interpreter can
// write inside it and — when the probe found the control present —
// cannot write outside it (EACCES) nor spawn an ungranted program
// (EACCES). Creating the denied file outside the grants is refused on
// the ungranted creation right, consistent with the handled mutation
// set. When the probe found the control absent the same attempts
// succeed and the evidence reports unavailable, so the row proves both
// branches on any Linux host; the require-present row above it pins
// which branch the hosted runner takes.
func TestLinuxLandlockConfinementMatchesProbe(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Fatalf("the probe program is unavailable: %v", err)
	}
	_, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	present := map[string]bool{}
	for _, probe := range probes {
		present[probe.Name] = probe.Present
	}
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.ProjectRoot = fixture.root
	fixture.declareCapabilities(t, `{"filesystem": "repo", "env_read": ["STUB_WRITE_TARGETS", "STUB_EXEC_TRY"]}`)
	allowedDir := filepath.Join(fixture.root, "writable")
	if err := os.MkdirAll(allowedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	allowed := filepath.Join(allowedDir, "allowed.txt")
	deniedDir := mustPhysical(t, t.TempDir())
	denied := filepath.Join(deniedDir, "denied.txt")
	fixture.request.HostEnvironment = []string{
		"STUB_WRITE_TARGETS=" + allowed + string(os.PathListSeparator) + denied,
		"STUB_EXEC_TRY=" + truePath,
	}
	result, report := runDerivedSession(t, fixture)
	for _, control := range []string{ScriptControlDescendantExecDenial, ScriptControlFilesystemWriteConfinement} {
		want := ScriptStatusUnavailable
		if present[control] {
			want = ScriptStatusApplied
		}
		if got := evidenceStatus(t, result.Evidence, control); got != want {
			t.Fatalf("control %q evidence status = %q, want %q for this host's probe", control, got, want)
		}
	}
	if report.Write[allowed] != "" {
		t.Fatalf("a write inside the derived path set failed: %s", report.Write[allowed])
	}
	if present[ScriptControlFilesystemWriteConfinement] {
		// Landlock denials surface as EACCES ("permission denied");
		// any other failure would name a different mechanism.
		if !strings.Contains(report.Write[denied], "permission denied") {
			t.Fatalf("a write outside the derived path set was not refused with EACCES under write confinement: %q", report.Write[denied])
		}
	} else if report.Write[denied] != "" {
		t.Fatalf("a write outside the derived path set failed without confinement: %s", report.Write[denied])
	}
	if present[ScriptControlDescendantExecDenial] {
		if !strings.Contains(report.ExecTry, "permission denied") {
			t.Fatalf("a descendant spawn outside the farm was not refused with EACCES under exec denial: %q", report.ExecTry)
		}
	} else if report.ExecTry != "" {
		t.Fatalf("a descendant spawn failed without exec denial: %s", report.ExecTry)
	}
}

// TestLinuxWriteConfinementTruncateMatchesProbe proves the Landlock
// truncation right behaviorally at the production entry: with an
// individual-file grant the interpreter truncates and overwrites the
// granted file (truncate(2), open with O_TRUNC, and overwrite all need
// the file rights the grant carries), while the same truncation
// operations on an existing file outside every grant are refused with
// EACCES and the content is preserved — when the running kernel's ABI
// handles truncation (ABI 3+); on older ABIs the right is unhandled
// and the operations succeed. The evidence reports the control applied
// in the same invocation either way. The row requires the probe to
// find the control present: without enforcement it would prove
// nothing.
func TestLinuxWriteConfinementTruncateMatchesProbe(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	_, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	present := false
	for _, probe := range probes {
		if probe.Name == ScriptControlFilesystemWriteConfinement {
			present = probe.Present
		}
	}
	if !present {
		t.Fatal("the host probe did not find filesystem-write-confinement present, so the row proves nothing")
	}
	abi, err := landlockABI()
	if err != nil || abi < 1 {
		t.Fatalf("the running kernel reports no Landlock ABI: %v", err)
	}
	truncateHandled := abi >= 3
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.ProjectRoot = fixture.root
	fixture.declareCapabilities(t, `{"filesystem": ["granted.txt"], "env_read": ["STUB_TRUNCATE_TARGETS", "STUB_WRITE_TARGETS"]}`)
	granted := filepath.Join(fixture.root, "granted.txt")
	writeTestFile(t, granted, []byte("granted\n"), 0o644)
	outside := filepath.Join(mustPhysical(t, t.TempDir()), "outside.txt")
	writeTestFile(t, outside, []byte("outside\n"), 0o644)
	fixture.request.HostEnvironment = []string{
		"STUB_TRUNCATE_TARGETS=" + granted + string(os.PathListSeparator) + outside,
		"STUB_WRITE_TARGETS=" + granted,
	}
	result, report := runDerivedSession(t, fixture)
	if got := evidenceStatus(t, result.Evidence, ScriptControlFilesystemWriteConfinement); got != ScriptStatusApplied {
		t.Fatalf("write-confinement evidence status = %q, want applied", got)
	}
	// The granted individual file truncates and overwrites on every
	// ABI: the file-typed grant carries the file rights, and an
	// unhandled right is allowed.
	for _, check := range []struct{ outcome, name string }{
		{report.Truncate[granted], "truncate"},
		{report.OTrunc[granted], "O_TRUNC open"},
		{report.Write[granted], "overwrite"},
	} {
		if check.outcome != "" {
			t.Fatalf("%s of the granted file failed: %s", check.name, check.outcome)
		}
	}
	for _, check := range []struct {
		values map[string]string
		name   string
	}{
		{report.Truncate, "truncate"}, {report.OTrunc, "O_TRUNC open"}, {report.Write, "overwrite"},
	} {
		if _, ok := check.values[granted]; !ok {
			t.Fatalf("the stub never attempted to %s the granted file", check.name)
		}
	}
	if _, ok := report.Truncate[outside]; !ok {
		t.Fatal("the stub never attempted to truncate the outside file")
	}
	if _, ok := report.OTrunc[outside]; !ok {
		t.Fatal("the stub never attempted an O_TRUNC open of the outside file")
	}
	if payload, err := os.ReadFile(outside); truncateHandled {
		// Landlock denials surface as EACCES ("permission denied");
		// any other failure would name a different mechanism.
		if !strings.Contains(report.Truncate[outside], "permission denied") {
			t.Fatalf("truncate outside the grants was not refused with EACCES: %q", report.Truncate[outside])
		}
		if !strings.Contains(report.OTrunc[outside], "permission denied") {
			t.Fatalf("O_TRUNC open outside the grants was not refused with EACCES: %q", report.OTrunc[outside])
		}
		if err != nil || string(payload) != "outside\n" {
			t.Fatalf("outside file = %q, want the preserved content", payload)
		}
	} else {
		if report.Truncate[outside] != "" || report.OTrunc[outside] != "" {
			t.Fatalf("truncate outside the grants failed without a handled right: %q / %q",
				report.Truncate[outside], report.OTrunc[outside])
		}
		if err != nil || len(payload) != 0 {
			t.Fatalf("outside file = %q, want the truncated content", payload)
		}
	}
	if payload, err := os.ReadFile(granted); err != nil || string(payload) != "stub-write\n" {
		t.Fatalf("granted file = %q, want the overwrite payload", payload)
	}
}

// TestLinuxWriteConfinementDirectoryMutationMatchesProbe proves the
// Landlock directory mutation rights behaviorally at the production
// entry: unlink, rmdir, mkdir, FIFO creation, and same-directory
// rename outside every grant are refused with EACCES and preserve
// filesystem state on every ABI that provides them (ABI 1+ — the
// creation/removal rights exist since ABI 1), while the same
// operations inside a granted directory succeed on every ABI. The
// evidence reports the control applied in the same invocation. The row
// requires the probe to find the control present: without enforcement
// it would prove nothing.
func TestLinuxWriteConfinementDirectoryMutationMatchesProbe(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	_, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	present := false
	for _, probe := range probes {
		if probe.Name == ScriptControlFilesystemWriteConfinement {
			present = probe.Present
		}
	}
	if !present {
		t.Fatal("the host probe did not find filesystem-write-confinement present, so the row proves nothing")
	}
	abi, err := landlockABI()
	if err != nil || abi < 1 {
		t.Fatalf("the running kernel reports no Landlock ABI: %v", err)
	}
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.ProjectRoot = fixture.root
	fixture.declareCapabilities(t, `{"filesystem": "repo", "env_read": ["STUB_UNLINK_TARGETS", "STUB_RMDIR_TARGETS", "STUB_MKDIR_TARGETS", "STUB_MKFIFO_TARGETS", "STUB_RENAME_SRC", "STUB_RENAME_DST"]}`)
	inside := filepath.Join(fixture.root, "mutable")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := mustPhysical(t, t.TempDir())
	side := func(root string) (victim, empty, newDir, fifo, renameSrc, renameDst string) {
		victim = filepath.Join(root, "victim.txt")
		writeTestFile(t, victim, []byte("victim\n"), 0o644)
		empty = filepath.Join(root, "emptydir")
		if err := os.MkdirAll(empty, 0o755); err != nil {
			t.Fatal(err)
		}
		newDir = filepath.Join(root, "newdir")
		fifo = filepath.Join(root, "newfifo")
		renameSrc = filepath.Join(root, "rename-src.txt")
		writeTestFile(t, renameSrc, []byte("renamed\n"), 0o644)
		renameDst = filepath.Join(root, "rename-dst.txt")
		return victim, empty, newDir, fifo, renameSrc, renameDst
	}
	insideVictim, insideEmpty, insideNew, insideFifo, insideSrc, insideDst := side(inside)
	outsideVictim, outsideEmpty, outsideNew, outsideFifo, outsideSrc, outsideDst := side(outside)
	join := func(first, second string) string { return first + string(os.PathListSeparator) + second }
	fixture.request.HostEnvironment = []string{
		"STUB_UNLINK_TARGETS=" + join(insideVictim, outsideVictim),
		"STUB_RMDIR_TARGETS=" + join(insideEmpty, outsideEmpty),
		"STUB_MKDIR_TARGETS=" + join(insideNew, outsideNew),
		"STUB_MKFIFO_TARGETS=" + join(insideFifo, outsideFifo),
		"STUB_RENAME_SRC=" + join(insideSrc, outsideSrc),
		"STUB_RENAME_DST=" + join(insideDst, outsideDst),
	}
	result, report := runDerivedSession(t, fixture)
	if got := evidenceStatus(t, result.Evidence, ScriptControlFilesystemWriteConfinement); got != ScriptStatusApplied {
		t.Fatalf("write-confinement evidence status = %q, want applied", got)
	}
	// The same operations inside the granted directory succeed on
	// every ABI, and the filesystem shows their effects.
	for _, check := range []struct{ outcome, name string }{
		{report.Unlink[insideVictim], "unlink"},
		{report.Rmdir[insideEmpty], "rmdir"},
		{report.Mkdir[insideNew], "mkdir"},
		{report.Mkfifo[insideFifo], "mkfifo"},
		{report.Rename[insideSrc], "rename"},
	} {
		if check.outcome != "" {
			t.Fatalf("%s inside the grants failed: %s", check.name, check.outcome)
		}
	}
	if _, ok := report.Unlink[insideVictim]; !ok {
		t.Fatal("the stub never attempted the inside unlink")
	}
	if _, ok := report.Rmdir[insideEmpty]; !ok {
		t.Fatal("the stub never attempted the inside rmdir")
	}
	if _, ok := report.Mkdir[insideNew]; !ok {
		t.Fatal("the stub never attempted the inside mkdir")
	}
	if _, ok := report.Mkfifo[insideFifo]; !ok {
		t.Fatal("the stub never attempted the inside mkfifo")
	}
	if _, ok := report.Rename[insideSrc]; !ok {
		t.Fatal("the stub never attempted the inside rename")
	}
	if _, err := os.Stat(insideVictim); !os.IsNotExist(err) {
		t.Fatalf("the inside victim still exists: %v", err)
	}
	if _, err := os.Stat(insideEmpty); !os.IsNotExist(err) {
		t.Fatalf("the inside empty directory still exists: %v", err)
	}
	if info, err := os.Stat(insideNew); err != nil || !info.IsDir() {
		t.Fatalf("the inside directory was not created: %v", err)
	}
	if info, err := os.Stat(insideFifo); err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatalf("the inside FIFO was not created: %v", err)
	}
	if payload, err := os.ReadFile(insideDst); err != nil || string(payload) != "renamed\n" {
		t.Fatalf("the inside rename destination = %q, want the moved content", payload)
	}
	if _, err := os.Stat(insideSrc); !os.IsNotExist(err) {
		t.Fatalf("the inside rename source still exists: %v", err)
	}
	// Outside every grant the handled rights refuse with EACCES and
	// preserve state on every supported ABI: the creation/removal
	// rights exist since ABI 1 and the confinement handles all of
	// them, so no ABI allows these operations outside the grants.
	outcomes := []struct{ outcome, name string }{
		{report.Unlink[outsideVictim], "unlink"},
		{report.Rmdir[outsideEmpty], "rmdir"},
		{report.Mkdir[outsideNew], "mkdir"},
		{report.Mkfifo[outsideFifo], "mkfifo"},
		{report.Rename[outsideSrc], "rename"},
	}
	if _, ok := report.Unlink[outsideVictim]; !ok {
		t.Fatal("the stub never attempted the outside unlink")
	}
	if _, ok := report.Rmdir[outsideEmpty]; !ok {
		t.Fatal("the stub never attempted the outside rmdir")
	}
	if _, ok := report.Mkdir[outsideNew]; !ok {
		t.Fatal("the stub never attempted the outside mkdir")
	}
	if _, ok := report.Mkfifo[outsideFifo]; !ok {
		t.Fatal("the stub never attempted the outside mkfifo")
	}
	if _, ok := report.Rename[outsideSrc]; !ok {
		t.Fatal("the stub never attempted the outside rename")
	}
	for _, check := range outcomes {
		if !strings.Contains(check.outcome, "permission denied") {
			t.Fatalf("%s outside the grants was not refused with EACCES: %q", check.name, check.outcome)
		}
	}
	if payload, err := os.ReadFile(outsideVictim); err != nil || string(payload) != "victim\n" {
		t.Fatalf("the outside victim = %q, want the preserved content", payload)
	}
	if info, err := os.Stat(outsideEmpty); err != nil || !info.IsDir() {
		t.Fatalf("the outside empty directory is gone: %v", err)
	}
	for _, path := range []string{outsideNew, outsideFifo, outsideDst} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("outside path %q exists, want it never created: %v", path, err)
		}
	}
	if payload, err := os.ReadFile(outsideSrc); err != nil || string(payload) != "renamed\n" {
		t.Fatalf("the outside rename source = %q, want the preserved content", payload)
	}
}

// TestLinuxMissingDerivedPathRefusesWriteConfinement proves a derived
// path that cannot be ruled is an apply failure of its own control: the
// worker refuses with worker-protocol-invalid naming
// filesystem-write-confinement and the missing member — never by
// reporting an installable control unavailable (a probe contradiction),
// and never by voiding the execute denial (whose grants are valid, so
// the refusal must not name it). The interpreter never starts.
func TestLinuxMissingDerivedPathRefusesWriteConfinement(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	_, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	for _, probe := range probes {
		if probe.Name == ScriptControlFilesystemWriteConfinement && !probe.Present {
			t.Fatal("the host probe did not find filesystem-write-confinement present, so the row proves nothing")
		}
	}
	fixture := newLaunchFixture(t, os.Args[0])
	project := mustPhysical(t, t.TempDir())
	fixture.request.ProjectRoot = project
	fixture.declareCapabilities(t, `{"filesystem": ["no-such-path"], "env_read": ["STUB_MARKER"]}`)
	marker := filepath.Join(mustPhysical(t, t.TempDir()), "ran")
	fixture.request.HostEnvironment = []string{"STUB_MARKER=" + marker}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err = runSession(ctx, fixture.request, nil)
	if DiagnosticCode(err) != CodeWorkerProtocolInvalid {
		t.Fatalf("missing-derived-path error = %v, want %s", err, CodeWorkerProtocolInvalid)
	}
	if !strings.Contains(err.Error(), "cannot install inventory control") {
		t.Fatalf("missing-derived-path error = %v, want the install refusal", err)
	}
	if !strings.Contains(err.Error(), ScriptControlFilesystemWriteConfinement) {
		t.Fatalf("missing-derived-path error = %v, want it to name %q", err, ScriptControlFilesystemWriteConfinement)
	}
	if strings.Contains(err.Error(), ScriptControlDescendantExecDenial) {
		t.Fatalf("missing-derived-path error = %v, must not void the execute denial", err)
	}
	if !strings.Contains(err.Error(), "no-such-path") {
		t.Fatalf("missing-derived-path error = %v, want it to name the missing member", err)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Fatal("the interpreter ran for an unconstructible control")
	}
}

// TestHostProbeReportsClosedInventory bounds the per-invocation probe on
// the host inventory: it succeeds, reports exactly the eight closed
// controls with the table's availabilities and the probe timing, marks no
// fixed-unavailable control present, and substitutes no cached result.
func TestHostProbeReportsClosedInventory(t *testing.T) {
	platform, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	if platform != ScriptInventoryPlatform(runtime.GOOS) {
		t.Fatalf("platform = %q, want the host inventory", platform)
	}
	if len(probes) != len(scriptInventoryOrder) {
		t.Fatalf("the probe reports %d controls, want %d", len(probes), len(scriptInventoryOrder))
	}
	for index, probe := range probes {
		if probe.Name != scriptInventoryOrder[index] {
			t.Fatalf("probe order diverges at %d: %q", index, probe.Name)
		}
		want := scriptInventoryPlatforms[platform][probe.Name]
		if probe.Availability != want.Availability {
			t.Fatalf("control %q availability = %q, want %q", probe.Name, probe.Availability, want.Availability)
		}
		if probe.ProbedAt != ScriptProbeTiming {
			t.Fatalf("control %q probed_at = %q, want %q", probe.Name, probe.ProbedAt, ScriptProbeTiming)
		}
		if probe.Availability == ScriptAvailabilityUnavailable && probe.Present {
			t.Fatalf("fixed-unavailable control %q probes present", probe.Name)
		}
		if probe.Availability == ScriptAvailabilityAvailable && !probe.Present {
			t.Fatalf("available control %q probes absent without refusing", probe.Name)
		}
	}
}
