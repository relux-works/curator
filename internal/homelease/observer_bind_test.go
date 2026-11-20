//go:build linux || darwin

package homelease

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
)

// nativeTuple builds the well-formed bound tuple for this platform's
// variant, so the same refusal rows run on Linux and macOS.
func nativeTuple(pid int64) *state.Process {
	process := &state.Process{PID: pid, UID: 1001, BootID: "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10"}
	if nativeStartTimeKind() == state.StartLinux {
		process.StartTime = &state.StartTime{Kind: state.StartLinux, Ticks: "123456789", Hz: 100}
	} else {
		process.StartTime = &state.StartTime{Kind: state.StartDarwin, Sec: "1731715200", Usec: 123456}
	}
	return process
}

func cloneProcess(p *state.Process) *state.Process {
	out := *p
	if p.StartTime != nil {
		start := *p.StartTime
		out.StartTime = &start
	}
	return &out
}

func shiftStart(p *state.Process) {
	if p.StartTime.Kind == state.StartLinux {
		p.StartTime.Ticks = "123456790"
	} else {
		p.StartTime.Usec = 123457
	}
}

// v2Document returns one observation-v2 reply for the fixture holder, with
// the given session process (or none when process is nil).
func v2Document(t *testing.T, process *state.Process) map[string]any {
	t.Helper()
	doc := observationFixture(t)
	doc["capabilities"] = map[string]any{"process_observation": true}
	session := doc["session"].(map[string]any)
	session["state"] = "busy"
	if process != nil {
		session["process"] = process
	}
	return doc
}

// bindSetup acquires one prepared lease and returns the requested bound
// holder plus a production observer whose transport serves doc and whose
// kernel source is the given fake.
func bindSetup(t *testing.T, doc map[string]any, code int, kernel func(int64) (state.Process, error)) (fixture, *Record, Holder) {
	t.Helper()
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	session := "SES-fixture"
	h.Session = &session
	h.Process = nativeTuple(4242)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	observer, _ := observerExecutable(t, string(raw), code)
	observer.kernel = kernel
	f.m.observer = observer
	return f, r, h
}

func matchingKernel(requested *state.Process) func(int64) (state.Process, error) {
	return func(pid int64) (state.Process, error) {
		if pid != requested.PID {
			return state.Process{}, errors.New("wrong pid")
		}
		return *cloneProcess(requested), nil
	}
}

// Claim: bind succeeds only when the requested holder, the authoritative
// observation and Curator's own kernel reading are byte-equal; every
// refusal row preserves the prepared lease beside this positive control.
func TestProductionBindObserverRows(t *testing.T) {
	requested := nativeTuple(4242)
	// Positive control first: all three sources agree.
	f, r, h := bindSetup(t, v2Document(t, cloneProcess(requested)), 0, matchingKernel(requested))
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed || out.Lease.Revision != 2 {
		t.Fatalf("matching tuple control: %+v %s %v", out, reason, err)
	}
	// An identical retry re-verifies and still reports the bound outcome.
	out, reason, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || out.Changed || out.Lease.Revision != 2 {
		t.Fatalf("verified retry control: %+v %s %v", out, reason, err)
	}

	type row struct {
		name       string
		observed   *state.Process // process member of the v2 reply; nil omits it
		kernel     func(int64) (state.Process, error)
		code       int
		wantCode   string
		wantReason string
	}
	kernelWith := func(mutate func(*state.Process)) func(int64) (state.Process, error) {
		return func(int64) (state.Process, error) {
			process := cloneProcess(requested)
			mutate(process)
			return *process, nil
		}
	}
	observedWith := func(mutate func(*state.Process)) *state.Process {
		process := cloneProcess(requested)
		mutate(process)
		return process
	}
	rows := []row{
		{name: "pid-reuse-within-boot", observed: cloneProcess(requested),
			kernel: kernelWith(shiftStart), wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "pid-reuse-across-boots", observed: cloneProcess(requested),
			kernel:   kernelWith(func(p *state.Process) { p.BootID = "7f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10" }),
			wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "wrong-real-uid", observed: cloneProcess(requested),
			kernel:   kernelWith(func(p *state.Process) { p.UID = 1002 }),
			wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "manager-record-agree-kernel-disagrees", observed: cloneProcess(requested),
			kernel: kernelWith(shiftStart), wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "holder-kernel-agree-observation-disagrees", observed: observedWith(shiftStart),
			kernel: matchingKernel(requested), wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "start-time-member-mismatch", observed: cloneProcess(requested),
			kernel: kernelWith(func(p *state.Process) {
				if p.StartTime.Kind == state.StartLinux {
					p.StartTime.Hz = 250
				} else {
					p.StartTime.Usec = 999999
				}
			}), wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "incarnation-changed-in-pass", observed: cloneProcess(requested),
			kernel:   func(int64) (state.Process, error) { return state.Process{}, errKernelChanged },
			wantCode: "environment_lease_holder_mismatch", wantReason: "process_incarnation_mismatch"},
		{name: "observation-unavailable", observed: cloneProcess(requested), code: 1,
			kernel: matchingKernel(requested), wantCode: "environment_lease_authority_refused", wantReason: "authority_unavailable"},
		{name: "missing-process", observed: nil,
			kernel: matchingKernel(requested), wantCode: "environment_lease_authority_refused", wantReason: "authority_unavailable"},
		{name: "vanished-pid", observed: cloneProcess(requested),
			kernel:   func(int64) (state.Process, error) { return state.Process{}, os.ErrNotExist },
			wantCode: "environment_lease_authority_refused", wantReason: "authority_unavailable"},
		{name: "unreadable-kernel-field", observed: cloneProcess(requested),
			kernel:   func(int64) (state.Process, error) { return state.Process{}, os.ErrPermission },
			wantCode: "environment_lease_authority_refused", wantReason: "authority_unavailable"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			f, r, h := bindSetup(t, v2Document(t, row.observed), row.code, row.kernel)
			before := snapshot(t, f.m.home)
			_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			requireRefusal(t, err, row.wantCode, row.wantReason)
			unchanged(t, before, f.m.home)
			// No retry or replay cures the refusal while its cause remains.
			_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
			requireRefusal(t, err, row.wantCode, row.wantReason)
			unchanged(t, before, f.m.home)
		})
	}
}

// Claim: when the observation and kernel agree with each other but not with
// the requested holder, bind refuses; the corrected holder then binds.
func TestProductionBindHolderDisagrees(t *testing.T) {
	agreed := nativeTuple(4242)
	f, r, h := bindSetup(t, v2Document(t, cloneProcess(agreed)), 0, matchingKernel(agreed))
	shiftStart(h.Process)
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_holder_mismatch", "process_incarnation_mismatch")
	unchanged(t, before, f.m.home)
	h.Process = cloneProcess(agreed)
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("corrected holder control: %+v %s %v", out, reason, err)
	}
}

// Claim: a well-formed opposite-platform variant is unsupported evidence on
// this platform and never binds, while the native variant beside it does.
func TestProductionBindUnsupportedVariant(t *testing.T) {
	requested := nativeTuple(4242)
	f, r, h := bindSetup(t, v2Document(t, cloneProcess(requested)), 0, matchingKernel(requested))
	foreign := cloneProcess(requested)
	if nativeStartTimeKind() == state.StartLinux {
		foreign.StartTime = &state.StartTime{Kind: state.StartDarwin, Sec: "1731715200", Usec: 123456}
	} else {
		foreign.StartTime = &state.StartTime{Kind: state.StartLinux, Ticks: "123456789", Hz: 100}
	}
	h.Process = foreign
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
	h.Process = cloneProcess(requested)
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("native variant control: %+v %s %v", out, reason, err)
	}
}

// Claim: an identical committed bind retry whose fresh kernel reading no
// longer matches is refused, and the already-bound record is left intact.
func TestProductionBindRetryReverifies(t *testing.T) {
	requested := nativeTuple(4242)
	f, r, h := bindSetup(t, v2Document(t, cloneProcess(requested)), 0, matchingKernel(requested))
	out, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || !out.Changed {
		t.Fatalf("initial bind control: %+v %v", out, err)
	}
	observer := f.m.observer.(sessionManagerObserver)
	observer.kernel = func(int64) (state.Process, error) {
		process := cloneProcess(requested)
		shiftStart(process)
		return *process, nil
	}
	f.m.observer = observer
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_holder_mismatch", "process_incarnation_mismatch")
	unchanged(t, before, f.m.home)
}

// Claim: a legacy {pid, started_at} process stays parseable on a prepared
// holder but can never bind; the four-field tuple beside it binds.
func TestLegacyProcessCannotBind(t *testing.T) {
	requested := nativeTuple(4242)
	f, r, h := bindSetup(t, v2Document(t, cloneProcess(requested)), 0, matchingKernel(requested))
	h.Process = state.LegacyProcess(4242, "2026-10-27T00:00:01Z")
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "usage", "invalid_holder")
	unchanged(t, before, f.m.home)
	h.Process = cloneProcess(requested)
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("tuple control: %+v %s %v", out, reason, err)
	}
	// The same legacy form remains a valid prepared acquisition payload.
	f2 := newFixture(t)
	f2.req.Holder.Process = state.LegacyProcess(4242, "2026-10-27T00:00:01Z")
	if got := acquire(t, f2); got.Holder.Process == nil || !got.Holder.Process.IsLegacy() {
		t.Fatalf("legacy prepared control did not retain its form: %+v", got.Holder.Process)
	}
}
