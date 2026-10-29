package envorigin

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Proves explicit attestation presence/idempotence without reading process
// selectors or an account home. No CLI, credential-entry or transaction claim.
func TestProvisioningOriginAdoptPresence(t *testing.T) {
	r, b, manager := fixtureRecord(t)
	t.Setenv("CODEX_HOME", filepath.Join(manager, "environments", "b", "codex_cli"))
	t.Setenv("HOME", filepath.Join(t.TempDir(), "poison"))
	req := AdoptRequest{Adapter: r.Adapter, Binding: b, OperatorHome: r.OperatorHome, NativeRoot: r.NativeRoot, SelectorState: "absent"}
	raw, changed, err := Attest(manager, req, nil)
	if err != nil || !changed {
		t.Fatalf("attest: %v %v", changed, err)
	}
	got, err := Parse(raw, b, r.Adapter, manager)
	if err != nil || got.Capture.Method != "adoption" || got.Selector.State != "absent" {
		t.Fatalf("attested %+v: %v", got, err)
	}
	again, changed, err := Attest(manager, req, raw)
	if err != nil || changed || !bytes.Equal(raw, again) {
		t.Fatalf("non-idempotent attestation: %v %v", changed, err)
	}
	empty := ""
	value := filepath.Join(manager, "environments", "original", "codex_cli")
	for _, tc := range []struct {
		state string
		value *string
		valid bool
	}{{"absent", nil, true}, {"absent", &empty, false}, {"present-empty", nil, false}, {"present-empty", &empty, true}, {"present-empty", &value, false}, {"present-value", nil, false}, {"present-value", &empty, false}, {"present-value", &value, true}} {
		req.SelectorState = tc.state
		req.SelectorValue = tc.value
		_, _, err := Attest(manager, req, nil)
		if (err == nil) != tc.valid {
			t.Fatalf("presence %s/%v: %v", tc.state, tc.value, err)
		}
	}
	req.SelectorState = "absent"
	req.SelectorValue = nil
	req.NativeRoot = filepath.Join(t.TempDir(), "different")
	if _, _, err := Attest(manager, req, raw); err == nil {
		t.Fatal("attestation overwrote recorded A")
	}
	if _, err := os.Lstat(r.NativeRoot); !os.IsNotExist(err) {
		t.Fatal("attestation created native root")
	}
}
