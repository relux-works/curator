package probe

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestReadMarkerPIDAcceptsOnlyAMarker(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return path
	}

	if got := readMarkerPID(write("good", "pid=4242 pgid=1 sid=2 ppid=3\n")); got != 4242 {
		t.Errorf("readMarkerPID(marker) = %d, want 4242", got)
	}
	for name, content := range map[string]string{
		"empty":        "",
		"not-a-marker": "hello\n",
		"truncated":    "pid=7\n",
	} {
		if got := readMarkerPID(write(name, content)); got != -1 {
			t.Errorf("readMarkerPID(%s) = %d, want -1", name, got)
		}
	}
	if got := readMarkerPID(filepath.Join(dir, "absent")); got != -1 {
		t.Errorf("readMarkerPID(absent) = %d, want -1", got)
	}
}

func TestWaitForMarkerPIDGivesUp(t *testing.T) {
	start := time.Now()
	if got := waitForMarkerPID(filepath.Join(t.TempDir(), "never"), 50*time.Millisecond); got != -1 {
		t.Errorf("waitForMarkerPID = %d, want -1", got)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("waiting for an absent marker took %s", elapsed)
	}
}

// Two probes in one run must not read each other's markers: a stale pid from an
// earlier probe would be reported as a survivor of a later one.
func TestFreshMarkerBaseIsolatesProbes(t *testing.T) {
	env := &Environment{MarkerDir: t.TempDir()}

	first, err := env.freshMarkerBase("alpha")
	if err != nil {
		t.Fatalf("freshMarkerBase: %v", err)
	}
	if err := os.WriteFile(first+".detached", []byte("pid=1 pgid=1 sid=1 ppid=1\n"), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	second, err := env.freshMarkerBase("beta")
	if err != nil {
		t.Fatalf("freshMarkerBase: %v", err)
	}
	if first == second {
		t.Fatal("two probes were given the same marker base")
	}
	if readMarkerPID(second+".detached") != -1 {
		t.Error("a fresh marker base already carried a marker")
	}

	// Re-using a name clears what the previous run left behind.
	again, err := env.freshMarkerBase("alpha")
	if err != nil {
		t.Fatalf("freshMarkerBase: %v", err)
	}
	if readMarkerPID(again+".detached") != -1 {
		t.Error("re-using a marker base kept the previous run's marker")
	}
}

func TestAlivePIDsReportsOnlyLivingProcesses(t *testing.T) {
	// This process is alive; pid 0 and a negative pid are not process handles
	// this probe should ever report as survivors.
	got := alivePIDs(os.Getpid(), 0, -1)
	if len(got) != 1 || got[0] != os.Getpid() {
		t.Errorf("alivePIDs = %v, want just this process (%d)", got, os.Getpid())
	}
}

// The end-to-end deadline measurement. It asserts what the probe owes its host
// and its reader — that the domain really started, that the deadline really
// fired, and that nothing is left running — but not which of the descendants
// survived, because that is the platform result this prototype exists to
// observe rather than to require.
func TestProbeWallClockMeasuresADeadlineAndCleansUp(t *testing.T) {
	agent := requireAgent(t)

	env, err := NewEnvironment(t.TempDir(), agent)
	if err != nil {
		t.Fatalf("NewEnvironment: %v", err)
	}
	defer env.Close()

	run := runContext{ctx: context.Background(), env: env}
	out := run.probeWallClock()

	if out.err != nil {
		t.Fatalf("the wall-clock probe failed: %v", out.err)
	}
	if !out.started {
		t.Fatalf("the domain or its descendants never started: %s", out.startFailure)
	}
	if !out.deadlineFired {
		t.Errorf("the deadline did not fire after %s against a declared %s", out.elapsed, out.declared)
	}
	if out.elapsed < out.declared {
		t.Errorf("the domain root lived %s, less than the declared %s, so the deadline is not what ended it",
			out.elapsed, out.declared)
	}
	if out.rootAliveAfterDeadline {
		t.Errorf("the domain root (pid %d) outlived its own deadline", out.rootPID)
	}

	// The requirement this probe exists to check: whatever the platform did,
	// the probe leaves no descendant of the domain running on the host.
	if len(out.survivorsAfterCleanup) != 0 {
		for _, pid := range out.survivorsAfterCleanup {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
		t.Errorf("the wall-clock probe left %v running after its own cleanup", out.survivorsAfterCleanup)
	}

	// The control has to have run, otherwise the deadline result above is not
	// attributable to the deadline.
	if !out.controlRan || out.controlErr != nil {
		t.Errorf("the wall-clock control did not run: ran=%v err=%v", out.controlRan, out.controlErr)
	}
	if out.controlElapsed >= out.declared {
		t.Errorf("the control took %s, which is not inside the declared %s", out.controlElapsed, out.declared)
	}
	if out.controlExitCode != 0 {
		t.Errorf("the control domain exited with status %d, so it did not finish its work in time", out.controlExitCode)
	}

	// The reduction has to be able to say something about what was measured.
	checks := wallClockChecks(out)
	if len(checks) < 2 {
		t.Fatalf("a completed wall-clock measurement reduced to %d checks", len(checks))
	}
	for _, check := range checks {
		if check.Observed == "probe-failed" {
			t.Errorf("a completed measurement still produced a failed check: %+v", check)
		}
	}
	if !requireCheck(t, checks, "wall-clock:harness-leaves-no-descendant-behind").Pass {
		t.Error("the hygiene check does not agree with the measured cleanup")
	}
}
