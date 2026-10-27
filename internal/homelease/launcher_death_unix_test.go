//go:build !windows

package homelease

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
)

// Corpus 8: killing only our recorded launcher PID leaves its real sleeping
// descendant alive and the real store protected. Killing that descendant plus
// synthetic exact reconciliation permits writes. This exercises retention,
// not production OS-incarnation observation or arbitrary process-tree fencing.
func TestSemanticDeadLauncherRetainsLivingDescendant(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "/bin/sleep 120 & echo \"$!\"; wait")
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	line, err := bufio.NewReader(pipe).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || pid <= 1 {
		t.Fatalf("fixture descendant pid: %q", line)
	}
	descendant, err := os.FindProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = descendant.Kill() })
	if err := descendant.Signal(syscall.Signal(0)); err != nil {
		t.Fatal("descendant never started", err)
	}
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	h.Process = &state.Process{PID: int64(pid), StartedAt: "2026-10-27T00:00:01Z"}
	f.o.proof.Holder = h
	if _, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	if err := descendant.Signal(syscall.Signal(0)); err != nil {
		t.Fatal("launcher death also stopped descendant", err)
	}
	before := snapshot(t, f.m.home)
	requireRefusal(t, guard(t, f, nil), "environment_home_leased", "blocking_holders")
	unchanged(t, before, f.m.home)
	f.o.err = errObservationUnavailable
	_, _, err = f.m.Release(context.Background(), r.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
	unchanged(t, before, f.m.home)
	if err := descendant.Kill(); err != nil {
		t.Fatal(err)
	}
	f.o.err = nil
	f.o.proof = Evidence{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
	if _, _, err := f.m.Release(context.Background(), r.LeaseID); err != nil {
		t.Fatal(err)
	}
	if err := guard(t, f, nil); err != nil {
		t.Fatal("reconciled control", err)
	}
}
