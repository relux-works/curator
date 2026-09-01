package probe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
)

// The wall-clock bound is the one aggregate bound no rlimit expresses: a build
// must end when its declared wall time is up, and every descendant must end
// with it. A supervisor implements it with a deadline, so that is what this
// probe measures — a real deadline, on a domain that is still working when it
// arrives, with the descendant tree inspected afterwards.
//
// The probe measures three separate things, because on macOS they have three
// different answers:
//
//   - whether the deadline ends the process the supervisor started;
//   - whether ending that process ends the descendants it created;
//   - whether anything at all survives the strongest teardown a plain
//     supervisor can issue.

const (
	// wallClockDeadline is the declared wall-clock bound. It is long enough for
	// the descendants to start and register, and short enough that the probe
	// stays fast.
	wallClockDeadline = 3 * time.Second
	// descendantHold is how long the descendants stay alive. It is far past the
	// deadline so that a descendant found alive afterwards is a survivor rather
	// than a process that had simply not exited yet.
	descendantHold = 30
	// settle is how long the probe waits before asking whether a signalled
	// process is gone. A SIGKILL is delivered synchronously, but reaping is not.
	settle = 400 * time.Millisecond
)

// wallClockOutcome is what one deadline measurement observed.
type wallClockOutcome struct {
	declared time.Duration

	rootPID     int
	rootPGID    int
	detachedPID int
	attachedPID int

	// started reports that the domain root and both descendants really existed,
	// without which "nothing survived the deadline" says nothing at all.
	started      bool
	startFailure string

	// elapsed is how long the domain root actually lived, and deadlineFired
	// reports that the supervisor's deadline is what ended it rather than the
	// root exiting on its own.
	elapsed       time.Duration
	deadlineFired bool

	// State immediately after the deadline cancelled the supervised process.
	rootAliveAfterDeadline     bool
	attachedAliveAfterDeadline bool
	detachedAliveAfterDeadline bool

	// State after the supervisor additionally signalled the process group,
	// which is the strongest domain-wide teardown macOS offers a plain
	// supervisor.
	groupKillErr                error
	attachedAliveAfterGroupKill bool
	detachedAliveAfterGroupKill bool

	// survivorsBeforeCleanup are the domain members still alive once the
	// supervisor has done everything a hardened implementation could do.
	// survivorsAfterCleanup are the ones still alive after this harness hunted
	// them down by pid, which a production implementation cannot rely on but
	// which this probe owes the host it ran on.
	survivorsBeforeCleanup []int
	survivorsAfterCleanup  []int

	// The control: the same domain with work that finishes inside the deadline.
	// Without it, a root that died at the deadline could not be distinguished
	// from a root that was going to exit anyway.
	controlRan      bool
	controlElapsed  time.Duration
	controlExitCode int
	controlErr      error

	err error
}

// probeWallClock runs the deadline measurement.
func (r runContext) probeWallClock() wallClockOutcome {
	out := wallClockOutcome{declared: wallClockDeadline, rootPID: -1, rootPGID: -1, detachedPID: -1, attachedPID: -1}

	markerBase, err := r.env.freshMarkerBase("wallclock")
	if err != nil {
		out.err = err
		return out
	}

	profile := r.baseProfile()
	profile.WritablePaths = append(profile.WritablePaths, r.env.MarkerDir)
	argv := []string{r.env.SelfPath, "__inside", inside.OpDescendant}
	env := r.env.AgentEnv(
		inside.EnvMarker+"="+markerBase,
		inside.EnvHold+"="+fmt.Sprint(descendantHold),
		inside.EnvRootHold+"="+fmt.Sprint(descendantHold),
	)

	deadlineCtx, cancel := context.WithTimeout(r.ctx, out.declared)
	defer cancel()

	started := time.Now()
	handle, err := r.env.startInOwnProcessGroup(deadlineCtx, "wallclock", profile, argv, env)
	if err != nil {
		out.err = err
		return out
	}
	out.rootPID = handle.pid
	out.rootPGID = handle.pgid

	// The descendants announce themselves by writing their markers. The domain
	// root writes no report here: it is holding when the deadline arrives, so
	// its stdout never carries one.
	out.detachedPID = waitForMarkerPID(markerBase+".detached", out.declared)
	out.attachedPID = waitForMarkerPID(markerBase+".attached", out.declared)
	out.started = out.rootPID > 0 && out.detachedPID > 0 && out.attachedPID > 0
	if !out.started {
		out.startFailure = fmt.Sprintf("domain root pid %d, detached descendant pid %d, attached descendant pid %d",
			out.rootPID, out.detachedPID, out.attachedPID)
	}

	waitErr := handle.wait()
	out.elapsed = time.Since(started)
	out.deadlineFired = errors.Is(deadlineCtx.Err(), context.DeadlineExceeded)
	_ = waitErr // the exit status of a process killed by its deadline carries nothing the probe needs

	time.Sleep(settle)
	out.rootAliveAfterDeadline = alive(out.rootPID)
	out.attachedAliveAfterDeadline = alive(out.attachedPID)
	out.detachedAliveAfterDeadline = alive(out.detachedPID)

	// Everything a hardened macOS supervisor could still do: signal the process
	// group it created.
	if out.rootPGID > 0 {
		out.groupKillErr = syscall.Kill(-out.rootPGID, syscall.SIGKILL)
	} else {
		out.groupKillErr = fmt.Errorf("the supervisor obtained no process-group handle for the domain")
	}
	time.Sleep(settle)
	out.attachedAliveAfterGroupKill = alive(out.attachedPID)
	out.detachedAliveAfterGroupKill = alive(out.detachedPID)

	out.survivorsBeforeCleanup = alivePIDs(out.rootPID, out.attachedPID, out.detachedPID)

	// Whatever the platform did or did not do, this probe does not leave
	// processes on the host it measured. The cleanup is by pid, which is exactly
	// what a production implementation must not have to rely on, so it is
	// recorded separately from the platform result rather than folded into it.
	for _, pid := range out.survivorsBeforeCleanup {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	time.Sleep(settle)
	out.survivorsAfterCleanup = alivePIDs(out.rootPID, out.attachedPID, out.detachedPID)

	r.runWallClockControl(&out)
	return out
}

// runWallClockControl runs the same domain with work that fits inside the
// deadline. It is the negative control for the whole measurement: it shows the
// deadline is what ended the run above, and that a domain which finishes in time
// is left alone.
func (r runContext) runWallClockControl(out *wallClockOutcome) {
	markerBase, err := r.env.freshMarkerBase("wallclock-control")
	if err != nil {
		out.controlErr = err
		return
	}

	profile := r.baseProfile()
	profile.WritablePaths = append(profile.WritablePaths, r.env.MarkerDir)
	argv := []string{r.env.SelfPath, "__inside", inside.OpDescendant}
	env := r.env.AgentEnv(
		inside.EnvMarker+"="+markerBase,
		// Descendants that exit almost at once, and a root that does not hold.
		inside.EnvHold+"=1",
	)

	ctx, cancel := context.WithTimeout(r.ctx, out.declared)
	defer cancel()

	started := time.Now()
	handle, err := r.env.startInOwnProcessGroup(ctx, "wallclock-control", profile, argv, env)
	if err != nil {
		out.controlErr = err
		return
	}
	waitErr := handle.wait()
	out.controlElapsed = time.Since(started)
	out.controlExitCode = handle.exitCode()
	out.controlRan = true
	if waitErr != nil && out.controlExitCode < 0 {
		out.controlErr = waitErr
	}

	// The control's own descendants are short-lived, but the probe does not
	// depend on that: it signals the group it created either way.
	if handle.pgid > 0 {
		_ = syscall.Kill(-handle.pgid, syscall.SIGKILL)
	}
	if pid := readMarkerPID(markerBase + ".detached"); pid > 0 {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}

// alivePIDs returns the subset of pids that still exist.
func alivePIDs(pids ...int) []int {
	var out []int
	for _, pid := range pids {
		if alive(pid) {
			out = append(out, pid)
		}
	}
	return out
}

// waitForMarkerPID waits for a descendant to announce itself and returns the pid
// it wrote, or -1 when it never did.
func waitForMarkerPID(path string, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pid := readMarkerPID(path); pid > 0 {
			return pid
		}
		time.Sleep(20 * time.Millisecond)
	}
	return -1
}

func readMarkerPID(path string) int {
	data, err := os.ReadFile(path) //nolint:gosec // path is the harness's own marker file under its private work directory
	if err != nil {
		return -1
	}
	var pid, pgid, sid, ppid int
	if _, err := fmt.Sscanf(string(data), "pid=%d pgid=%d sid=%d ppid=%d", &pid, &pgid, &sid, &ppid); err != nil {
		return -1
	}
	return pid
}

// freshMarkerBase gives one probe its own marker prefix under the marker
// directory, so two probes in the same run cannot read each other's markers.
func (e *Environment) freshMarkerBase(name string) (string, error) {
	dir := filepath.Join(e.MarkerDir, name)
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "descendant"), nil
}
