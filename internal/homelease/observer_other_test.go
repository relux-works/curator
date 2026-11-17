//go:build !linux && !darwin

package homelease

import (
	"context"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
)

// Claim: without a native observer variant, the production bind refuses as
// unavailable authority for either well-formed tuple, while explicitly
// synthetic store evidence still binds beside it.
func TestProductionBindUnsupportedPlatform(t *testing.T) {
	for _, start := range []*state.StartTime{
		{Kind: state.StartLinux, Ticks: "123456789", Hz: 100},
		{Kind: state.StartDarwin, Sec: "1731715200", Usec: 123456},
	} {
		f := newFixture(t)
		r := acquire(t, f)
		h := boundHolder(r.Holder)
		h.Process = &state.Process{PID: 4242, UID: 1001,
			BootID: "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10", StartTime: start}
		f.m.observer = SessionManagerObserver()
		before := snapshot(t, f.m.home)
		_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
		requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
		unchanged(t, before, f.m.home)
		f.m.observer = f.o
		f.o.proof.Holder = h
		out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
		if err != nil || reason != "bound" || !out.Changed {
			t.Fatalf("synthetic store control: %+v %s %v", out, reason, err)
		}
	}
}
