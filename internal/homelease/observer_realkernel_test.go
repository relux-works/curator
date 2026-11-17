//go:build linux || darwin

package homelease

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/homelease/state"
)

// shiftRealStart moves a real kernel tuple to a different incarnation of
// the same pid, simulating pid reuse within one boot.
func shiftRealStart(t *testing.T, p *state.Process) {
	t.Helper()
	if p.StartTime.Kind == state.StartLinux {
		ticks, err := strconv.ParseUint(p.StartTime.Ticks, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		p.StartTime.Ticks = strconv.FormatUint(ticks+1, 10)
		return
	}
	if p.StartTime.Usec < 999999 {
		p.StartTime.Usec++
		return
	}
	sec, err := strconv.ParseUint(p.StartTime.Sec, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	p.StartTime.Sec = strconv.FormatUint(sec+1, 10)
	p.StartTime.Usec = 0
}

// realKernelBindSetup acquires one prepared lease and wires the production
// observer (SessionManagerObserver, hence the real readKernelProcess) to a
// fixture transport serving a v2 document with the given process tuple.
func realKernelBindSetup(t *testing.T, serving *state.Process) (fixture, *Record, Holder) {
	t.Helper()
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	session := "SES-fixture"
	h.Session = &session
	raw, err := json.Marshal(v2Document(t, serving))
	if err != nil {
		t.Fatal(err)
	}
	transport, _ := observerExecutable(t, string(raw), 0)
	prod, ok := SessionManagerObserver().(sessionManagerObserver)
	if !ok || prod.kernel == nil {
		t.Fatal("production observer does not carry the real kernel reader")
	}
	prod.executable = transport.executable
	f.m.observer = prod
	return f, r, h
}

// Claim: bind through the real kernel observer succeeds for the test's own live child and refuses shifted-start, other-boot, wrong-uid and vanished incarnations of it without writes.
func TestRealKernelBindEndToEnd(t *testing.T) {
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Fatalf("owned child fixture requires a sleep binary: %v", err)
	}
	cmd := exec.Command(sleep, "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	reaped := false
	defer func() {
		if !reaped {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	pid := int64(cmd.Process.Pid)
	real, err := readKernelProcess(pid)
	if err != nil {
		t.Fatalf("production reader on owned child: %v", err)
	}
	if real.PID != pid || real.UID != int64(os.Getuid()) {
		t.Fatalf("production tuple does not name the owned child: %+v", real)
	}

	// Positive control: requested, observed and kernel tuples all name the
	// live child, read by the production kernel reader.
	f, r, h := realKernelBindSetup(t, cloneProcess(&real))
	h.Process = cloneProcess(&real)
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed || out.Lease.Revision != 2 {
		t.Fatalf("real kernel bind control: %+v %s %v", out, reason, err)
	}

	// Negatives: holder and manager observation agree with each other, so
	// only Curator's own kernel reading can refuse them.
	rows := []struct {
		name   string
		mutate func(*state.Process)
	}{
		{"pid-reuse-shifted-start", func(p *state.Process) { shiftRealStart(t, p) }},
		{"other-boot", func(p *state.Process) {
			boot := []byte(p.BootID)
			if boot[0] == '0' {
				boot[0] = '1'
			} else {
				boot[0] = '0'
			}
			p.BootID = string(boot)
		}},
		{"wrong-real-uid", func(p *state.Process) { p.UID++ }},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			fake := cloneProcess(&real)
			row.mutate(fake)
			f, r, h := realKernelBindSetup(t, cloneProcess(fake))
			h.Process = fake
			before := snapshot(t, f.m.home)
			_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			requireRefusal(t, err, "environment_lease_holder_mismatch", "process_incarnation_mismatch")
			unchanged(t, before, f.m.home)
		})
	}

	// Zombie: a killed but unreaped child has exited, so its retained
	// tuple is unavailable evidence and bind refuses without writes.
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := readKernelProcess(pid); err != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("zombie child still produced a kernel tuple")
		}
		time.Sleep(10 * time.Millisecond)
	}
	f, r, h = realKernelBindSetup(t, cloneProcess(&real))
	h.Process = cloneProcess(&real)
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)

	// Vanished: after the child is reaped, its former tuple stays
	// unavailable evidence, never a mismatch cure or a bind.
	_ = cmd.Wait()
	reaped = true
	f, r, h = realKernelBindSetup(t, cloneProcess(&real))
	h.Process = cloneProcess(&real)
	before = snapshot(t, f.m.home)
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
}
