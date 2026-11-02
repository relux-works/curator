package homelease

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/homelease/state"
)

func command(t *testing.T, f fixture, args ...string) (Result, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := RunCLI(context.Background(), args, func() (*Manager, error) { return f.m, nil }, &stdout, &stderr)
	raw := stdout.Bytes()
	if bytes.Count(raw, []byte("\n")) != 1 || !bytes.HasSuffix(raw, []byte("\n")) {
		t.Fatalf("stdout is not one JSON line: %q", raw)
	}
	result, err := DecodeResult(bytes.TrimSpace(raw))
	if err != nil || result.Exit != code {
		t.Fatalf("closed result exit mismatch: %s %v", raw, err)
	}
	if code != 0 && !strings.HasPrefix(stderr.String(), "curator: ") {
		t.Fatalf("missing diagnostic: %q", stderr.String())
	}
	return result, stderr.String()
}

// HL05/12/18: recognized usage errors keep the closed envelope, exact code and
// exit, never echo malformed holder/flag data, and leave all owned state intact.
// A valid list and acquire are paired controls using the same real manager.
func TestCLIClosedRefusalsAndControls(t *testing.T) {
	const sentinel = "PRIVATE_INPUT_SENTINEL"
	for _, tc := range []struct {
		name, reason string
		args         []string
	}{
		{"format", "invalid_format", []string{"list", "--format", sentinel}},
		{"help", "invalid_flags", []string{"acquire", "--help"}},
		{"unknown-flag", "invalid_flags", []string{"list", "--" + sentinel}},
		{"duplicate", "invalid_flags", []string{"list", "--profile", "sample", "--profile", sentinel}},
		{"holder", "invalid_holder", []string{"acquire", "--env", "claude", "--profile", "sample", "--lock-sha256", strings.Repeat("a", 64), "--request-id", "request", "--holder", `{"` + sentinel + `":true}`}},
		{"operator", "operator_confirmation_required", []string{"break", "--lease", "missing", "--if-revision", "1", "--reason", "deliberate abandon"}},
		{"reason", "reason_required", []string{"break", "--lease", "missing", "--if-revision", "1", "--operator"}},
		{"revision", "revision_required", []string{"break", "--lease", "missing", "--operator", "--reason", "deliberate abandon"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			acquire(t, f)
			before := snapshot(t, f.m.home)
			r, diagnostic := command(t, f, tc.args...)
			if r.Exit != 2 || r.OK || r.Code == nil || *r.Code != "usage" || r.Reason != tc.reason || strings.Contains(diagnostic, sentinel) {
				t.Fatalf("usage: %+v %q", r, diagnostic)
			}
			unchanged(t, before, f.m.home)
			r, _ = command(t, f, "list", "--format", "json")
			if !r.OK || r.Reason != "listed" {
				t.Fatalf("control: %+v", r)
			}
			unchanged(t, before, f.m.home)
		})
	}
	f := newFixture(t)
	holder, _ := state.Canonical(f.req.Holder)
	r, _ := command(t, f, "acquire", "--env", "claude", "--profile", f.req.Profile, "--lock-sha256", f.req.LockSHA256, "--request-id", f.req.RequestID, "--holder", string(holder))
	if !r.OK || r.Reason != "acquired" {
		t.Fatalf("acquire control: %+v", r)
	}
	raw, _ := json.Marshal(r.Result)
	var out Outcome
	if err := json.Unmarshal(raw, &out); err != nil || out.Lease == nil || !out.Changed {
		t.Fatalf("outcome: %s %v", raw, err)
	}
	before := snapshot(t, f.m.home)
	r, _ = command(t, f, "release", "--lease", "absent")
	if !r.OK || r.Reason != "unknown_lease" {
		t.Fatalf("absent: %+v", r)
	}
	unchanged(t, before, f.m.home)
}

// HL07/09/12/20: transport status is derived from typed store outcomes, with
// no result fragment on refusal. The subsequent exact CAS control succeeds.
func TestCLITypedStoreOutcomes(t *testing.T) {
	f := newFixture(t)
	r := acquire(t, f)
	before := snapshot(t, f.m.home)
	for _, tc := range []struct {
		args         []string
		code, reason string
	}{
		{[]string{"release", "--lease", r.LeaseID}, "environment_lease_release_unconfirmed", "session_state_unknown"},
		{[]string{"break", "--lease", r.LeaseID, "--if-revision", "2", "--operator", "--reason", "protection deliberately abandoned"}, "environment_lease_revision_conflict", "revision_changed"},
		{[]string{"break", "--lease", "unknown", "--if-revision", "1", "--operator", "--reason", "protection deliberately abandoned"}, "environment_lease_unknown", "unknown_lease"},
	} {
		got, _ := command(t, f, tc.args...)
		if got.Exit != 1 || got.OK || got.Code == nil || *got.Code != tc.code || got.Reason != tc.reason || got.Result != nil {
			t.Fatalf("typed refusal: %+v", got)
		}
		unchanged(t, before, f.m.home)
	}
	got, _ := command(t, f, "break", "--lease", r.LeaseID, "--if-revision", "1", "--operator", "--reason", "protection deliberately abandoned")
	if !got.OK || got.Exit != 0 || got.Reason != "broken" || got.Code != nil {
		t.Fatalf("control: %+v", got)
	}
}
