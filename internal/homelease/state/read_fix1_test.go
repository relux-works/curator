package state

import (
	"path/filepath"
	"strings"
	"testing"
)

// ledgerWithTicks builds one complete, schema-valid bound ledger whose
// stored process carries the given linux ticks string, plus the artifact
// reader DecodeLedger needs for its immutable binding.
func ledgerWithTicks(t *testing.T, ticks string) (string, []byte, func(string) ([]byte, error)) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "manager")
	admitted := filepath.Join(home, "environments", "sample", "claude_code")
	prepared := Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "manager-1", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "request-1"}
	initial := Record{
		Kind: "home-lease/1", LeaseID: "home-" + strings.Repeat("a", 32),
		ManagerRootID: "root-" + digest([]byte(home)), HomeID: "home-" + digest([]byte(admitted)),
		Environment: "claude_code", Profile: "sample", LockSHA256: strings.Repeat("b", 64),
		Revision: 1, Phase: "prepared", AcquiredAt: "2026-10-27T00:00:00Z",
		Holder: prepared, Observation: Observation{State: "not_observed"},
	}
	payload, err := Canonical(struct {
		Environment string `json:"environment"`
		Profile     string `json:"profile"`
		LockSHA256  string `json:"lock_sha256"`
		RequestID   string `json:"request_id"`
		Holder      Holder `json:"holder"`
	}{initial.Environment, initial.Profile, initial.LockSHA256, prepared.RequestID, prepared})
	if err != nil {
		t.Fatal(err)
	}
	session, generation := "fixture-session", int64(7)
	bound := prepared
	bound.Kind, bound.Session, bound.Generation = "bound", &session, &generation
	bound.Process = &Process{PID: 4242, UID: 1001, BootID: "6f1c0e2a-8b44-4d1e-9a51-2f0c3d7e9b10",
		StartTime: &StartTime{Kind: StartLinux, Ticks: ticks, Hz: 100}}
	record := initial
	record.Revision, record.Phase, record.Holder = 2, "active", bound
	observedAt := "2026-10-27T00:00:01Z"
	record.Observation = Observation{State: "not_observed", ObservedAt: &observedAt}
	bindRevision := int64(1)
	ledger := Ledger{Kind: "home-lease-state/1",
		Leases: map[string]Entry{initial.LeaseID: {
			Record: record, Initial: initial, PayloadSHA256: digest(payload),
			Admission:    Admission{Home: admitted, Protected: []string{admitted}, BuildKeys: []string{}, RuntimeEntries: []string{}},
			BindRevision: &bindRevision,
		}},
		Requests: map[string]string{prepared.RequestID: initial.LeaseID}}
	raw, err := Canonical(ledger)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := Canonical(initial)
	if err != nil {
		t.Fatal(err)
	}
	return home, raw, func(path string) ([]byte, error) {
		if path != BindingPath(home, initial.LeaseID) {
			t.Fatalf("unexpected artifact read: %s", path)
		}
		return binding, nil
	}
}

// Claim: ledger decode refuses a stored bound holder whose ticks exceed the uint64 semantic ceiling, while the ledger at the ceiling beside it decodes.
func TestDecodeLedgerAppliesSemanticCeilings(t *testing.T) {
	home, raw, readArtifact := ledgerWithTicks(t, "18446744073709551615") // 2^64-1
	ledger, err := DecodeLedger(home, raw, readArtifact)
	if err != nil || len(ledger.Leases) != 1 {
		t.Fatalf("ceiling control: %v", err)
	}
	home, raw, readArtifact = ledgerWithTicks(t, "18446744073709551616") // 2^64
	if _, err := DecodeLedger(home, raw, readArtifact); err == nil {
		t.Fatal("stored holder above the semantic ceiling decoded")
	} else if !strings.Contains(err.Error(), "environment_lease_unreadable/invalid_record") {
		t.Fatalf("above-ceiling ledger refusal: %v", err)
	}
}
