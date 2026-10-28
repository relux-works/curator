package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
)

// Actual generic record publication cannot bypass the complete home unit even
// with a valid stable origin. Nearby disjoint record publication must reach its
// write. This covers executor effect admission, not every CLI continuation.
func TestProvisioningOriginRequiresWholeUnit(t *testing.T) {
	for _, leaf := range []string{envorigin.Name, envmarker.Name, "AGENTS.md"} {
		t.Run(leaf, func(t *testing.T) {
			fx, _, _ := recordedOriginFixture(t, "a")
			home := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			before := nativeSnapshot(t, home)
			op, e := beginOperation(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			defer op.close()
			e = op.publish(map[string][]byte{filepath.Join(home, leaf): []byte("forbidden direct publication")})
			if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
				t.Fatalf("generic publisher bypassed whole-unit admission: %v", e)
			}
			if nativeSnapshot(t, home) != before {
				t.Fatal("refused publisher changed home")
			}
			healthy := filepath.Join(ProfileDir(fx.home, fx.profile), "owned-control")
			if e := op.publish(map[string][]byte{healthy: []byte("reached")}); e != nil {
				t.Fatalf("disjoint real publisher refused: %v", e)
			}
			raw, e := os.ReadFile(healthy)
			if e != nil || string(raw) != "reached" {
				t.Fatal("positive owner did not reach effect")
			}
		})
	}
}

// A hand-built partial profile-lock transaction cannot bypass the all-home
// integration owner. It is refused before changing any lock/home preimage.
func TestProvisioningOriginRejectsPartialLockUnit(t *testing.T) {
	fx, _, _ := recordedOriginFixture(t, "a")
	lock, hash, e := readLock(fx.home, fx.profile)
	if e != nil {
		t.Fatal(e)
	}
	lock.Members[0].Weight++
	raw, e := lock.Canonical()
	if e != nil {
		t.Fatal(e)
	}
	file, e := os.CreateTemp("", "partial-lock")
	if e != nil {
		t.Fatal(e)
	}
	defer os.Remove(file.Name())
	if _, e := file.Write(raw); e != nil {
		t.Fatal(e)
	}
	file.Close()
	op, e := beginOperation(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	defer op.close()
	path := lockPath(fx.home, fx.profile)
	digest, e := transaction.DigestPath(path)
	if e != nil {
		t.Fatal(e)
	}
	_, e = op.engine.Prepare(op.lock, transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "partial-record-publisher", Targets: []transaction.Target{{Class: "profile-record", Identifier: path, Kind: transaction.KindBytes, LivePath: path, StagedSource: file.Name(), PreimageDigest: digest}}})
	if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) {
		t.Fatalf("partial lock publisher admitted: %v", e)
	}
	_, after, e := readLock(fx.home, fx.profile)
	if e != nil || after != hash {
		t.Fatal("partial lock refusal changed installed authority")
	}
}

// Exact repeated installed-lock publication is an actual no-op, paired with
// the changed-lock all-home tests. It must not be rejected as an unbound edit.
func TestProvisioningOriginUnchangedLockPublication(t *testing.T) {
	fx, _, _ := recordedOriginFixture(t, "a")
	before := nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli"))
	raw, e := os.ReadFile(lockPath(fx.home, fx.profile))
	if e != nil {
		t.Fatal(e)
	}
	op, e := beginOperation(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	defer op.close()
	if e = op.publish(map[string][]byte{lockPath(fx.home, fx.profile): raw}); e != nil {
		t.Fatal(e)
	}
	if nativeSnapshot(t, ManagedHomeDir(fx.home, fx.profile, "codex_cli")) != before {
		t.Fatal("no-op lock publication changed home")
	}
}
