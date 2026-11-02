package homelease

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func observationFixture(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("observation/testdata/observation-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Document map[string]any }
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	r := rows[0].Document
	m := r["manager"].(map[string]any)
	m["id"] = "manager-1"
	m["registered_at"] = "2026-10-27T00:00:00Z"
	s := r["session"].(map[string]any)
	s["manager"] = "manager-1"
	s["session"] = "SES-fixture"
	s["state"] = "down"
	s["generation"] = float64(7)
	return r
}

// A production bind refusal preserves the prepared lease. Its paired exact
// synthetic observer admits bind, while production prepared release requires
// an unavailable no-start interface and cannot invoke a sessionless command.
func TestProductionObserverBindAndPreparedRelease(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	h := boundHolder(r.Holder)
	session := "SES-fixture"
	h.Session = &session
	doc := observationFixture(t)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	o, calls := observerExecutable(t, string(raw), 0)
	f.m.observer = o
	before := snapshot(t, f.m.home)
	_, _, err = f.m.Release(context.Background(), r.LeaseID)
	requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
	unchanged(t, before, f.m.home)
	if _, err := os.Stat(calls); !os.IsNotExist(err) {
		t.Fatal("prepared release invoked session command")
	}
	_, _, err = f.m.Bind(context.Background(), r.LeaseID, 1, h)
	requireRefusal(t, err, "environment_lease_authority_refused", "authority_unavailable")
	unchanged(t, before, f.m.home)
	if _, err := os.Stat(calls); err != nil {
		t.Fatal("bind did not reach production observer transport")
	}
	f.m.observer = f.o
	f.o.proof.Holder = h
	out, reason, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
	if err != nil || reason != "bound" || !out.Changed {
		t.Fatalf("qualified bind control: %+v %s %v", out, reason, err)
	}
}

func observerExecutable(t *testing.T, output string, code int) (sessionManagerObserver, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable transport fixture")
	}
	root := t.TempDir()
	path := filepath.Join(root, "curator-run")
	calls := filepath.Join(root, "calls")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	text := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + quote(calls) + "\nprintf '%s' " + quote(output) + "\nexit " + string(rune('0'+code)) + "\n"
	if err := os.WriteFile(path, []byte(text), 0700); err != nil {
		t.Fatal(err)
	}
	return sessionManagerObserver{executable: path}, calls
}

// The real command transport is bounded and uses the verified argv. Even a
// schema-valid exact-generation "down" response cannot manufacture the missing
// process, no-start, descendant or ownership proofs. Invalid/unavailable output
// also preserves the real store; independent lifecycle evidence is the positive
// release control. No live manager or provider is called by this fixture.
func TestProductionObserverRetainsUnprovenLease(t *testing.T) {
	for _, variant := range []string{"down", "unknown", "unreadable", "wrong-manager", "wrong-registration", "wrong-generation", "invented-proof", "malformed", "oversize", "exit", "missing", "prepared"} {
		t.Run(variant, func(t *testing.T) {
			f := newFixture(t)
			r := acquire(t, f)
			h := boundHolder(r.Holder)
			session := "SES-fixture"
			h.Session = &session
			f.o.proof.Holder = h
			out, _, err := f.m.Bind(context.Background(), r.LeaseID, 1, h)
			if err != nil {
				t.Fatal(err)
			}
			doc := observationFixture(t)
			m := doc["manager"].(map[string]any)
			s := doc["session"].(map[string]any)
			switch variant {
			case "unknown":
				s["state"] = "unknown"
			case "unreadable":
				s["state"] = "unknown"
				s["readable"] = false
				for _, k := range []string{"operation_id", "runtime", "native_id", "generation", "started_at", "last_activity_at"} {
					s[k] = nil
				}
			case "wrong-manager":
				m["id"] = "different"
			case "wrong-registration":
				m["registered_at"] = "2026-10-27T00:00:02Z"
			case "wrong-generation":
				s["generation"] = float64(9)
			case "invented-proof":
				s["process"] = h.Process
			}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if variant == "malformed" {
				raw = []byte("private sentinel not JSON")
			}
			if variant == "oversize" {
				raw = []byte(strings.Repeat("x", 300<<10))
			}
			code := 0
			if variant == "exit" {
				code = 1
			}
			o, calls := observerExecutable(t, string(raw), code)
			if variant == "missing" {
				o.executable = filepath.Join(t.TempDir(), "absent")
			}
			input := h
			if variant == "prepared" {
				input = r.Holder
			}
			proof, err := o.Observe(context.Background(), input)
			want := errObservationUnavailable
			if variant == "down" || variant == "unknown" {
				want = errObservationIncomplete
			}
			if !errors.Is(err, want) || !reflect.DeepEqual(proof, Evidence{}) {
				t.Fatalf("unsupported authority escaped: %+v %v", proof, err)
			}
			if variant != "missing" && variant != "prepared" {
				args, err := os.ReadFile(calls)
				if err != nil {
					t.Fatal(err)
				}
				if string(args) != "session-manager\nobserve\nmanager-1\n--session\nSES-fixture\n--json\n" {
					t.Fatalf("wrong command: %q", args)
				}
			} else if _, err := os.Stat(calls); !os.IsNotExist(err) {
				t.Fatal("unavailable/prepared observation invoked command")
			}
			f.m.observer = o
			before := snapshot(t, f.m.home)
			_, _, err = f.m.Release(context.Background(), r.LeaseID)
			requireRefusal(t, err, "environment_lease_release_unconfirmed", "session_state_unknown")
			unchanged(t, before, f.m.home)
			f.m.observer = f.o
			f.o.proof = Evidence{Holder: h, Terminal: true, DescendantsTerminated: true, OwnershipReconciled: true}
			got, reason, err := f.m.Release(context.Background(), r.LeaseID)
			if err != nil || reason != "released" || !got.Changed || got.Lease.Revision != out.Lease.Revision+1 {
				t.Fatalf("independent evidence control: %+v %s %v", got, reason, err)
			}
		})
	}
}
