package homelease

import (
	"context"
	"testing"
)

// observeOnlyObserver is an Observer that deliberately does not implement
// BindVerifier; its Observe evidence alone must never authorize a bind.
type observeOnlyObserver struct{ proof Evidence }

func (o observeOnlyObserver) Observe(context.Context, Holder) (Evidence, error) {
	return o.proof, nil
}

// Claim: bind refuses authority_unavailable when the observer is not a BindVerifier even with exact Observe evidence, while the synthetic BindVerifier beside it binds.
func TestBindRefusesObserverWithoutBindVerifier(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	f.m.observer = observeOnlyObserver{proof: Evidence{Holder: h}}
	before := snapshot(t, f.m.home)
	_, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
	f.m.observer = f.o
	f.o.proof.Holder = h
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed || out.Lease.Revision != 2 {
		t.Fatalf("synthetic verifier control: %+v %s %v", out, reason, err)
	}
}
