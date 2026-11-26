//go:build linux || darwin

package homelease

import (
	"context"
	"encoding/json"
	"testing"
)

// Claim: even a byte-equal v2 observation and kernel tuple add no release
// authority in the first use; release keeps HL09's unconfirmed refusal and
// the bound lease keeps blocking, beside the synthetic-evidence control.
func TestProductionReleaseHasNoFirstUseAuthority(t *testing.T) {
	requested := nativeTuple(4242)
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	session := "SES-fixture"
	h.Session = &session
	h.Process = cloneProcess(requested)
	f.o.proof.Holder = h
	bound, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || bound.Lease == nil {
		t.Fatalf("bind control: %+v %v", bound, err)
	}
	raw, err := json.Marshal(v2Document(t, cloneProcess(requested)))
	if err != nil {
		t.Fatal(err)
	}
	observer, _ := observerExecutable(t, string(raw), 0)
	observer.kernel = matchingKernel(requested)
	f.m.observer = observer
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Release(context.Background(), r.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
	unchanged(t, before, f.m.home)
	// A termination member is equally only a future release-rule input.
	down := v2Document(t, nil)
	downSession := down["session"].(map[string]any)
	downSession["state"] = "down"
	downSession["termination"] = map[string]any{"scope": "dedicated-uid", "witness": "zero-process-proof"}
	raw, err = json.Marshal(down)
	if err != nil {
		t.Fatal(err)
	}
	observer, _ = observerExecutable(t, string(raw), 0)
	observer.kernel = matchingKernel(requested)
	f.m.observer = observer
	_, _, err = f.m.Release(context.Background(), r.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
	unchanged(t, before, f.m.home)
	// Explicitly synthetic authoritative evidence is the positive control;
	// it does not claim the production integration exists.
	f.m.observer = f.o
	f.o.proof = Evidence{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
	out, reason, err := f.m.Release(context.Background(), r.LeaseID)
	if err != nil || reason != "released" || !out.Changed {
		t.Fatalf("synthetic evidence control: %+v %s %v", out, reason, err)
	}
}
