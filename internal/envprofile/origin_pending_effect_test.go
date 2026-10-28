package envprofile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// J1 is produced by the actual native publication owner and stopped after
// durable preparation. A different selector cannot grant J2 authority over its
// captured surface. A physically disjoint J2 reaches a real commit while J1 is
// pending; a fresh engine then completes J1. This is in-process restart evidence.
func TestProvisioningOriginPendingNativeEffects(t *testing.T) {
	fx, req, _ := recordedOriginFixture(t, "a")
	reached := false
	op, e := beginOperation(fx.home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointPrepared {
			reached = true
			panic("pending-native-stop")
		}
		return nil
	}}))
	if e != nil {
		t.Fatal(e)
	}
	func() {
		defer op.close()
		defer func() {
			if got := recover(); got != "pending-native-stop" {
				t.Fatalf("wrong native stop: %v", got)
			}
		}()
		_, _ = materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, req.NativeHomeOf)
	}()
	if !reached {
		t.Fatal("actual native prepare boundary not reached")
	}
	if e = envauthority.Validate(fx.home); e != nil {
		t.Fatalf("genuine pending native unit: %v", e)
	}
	native := fx.native["codex_cli"]
	before := nativeSnapshot(t, native)
	t.Setenv("CODEX_HOME", t.TempDir())
	manager, e := managerlock.New(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	held, e := manager.AcquireHomeOnly(context.Background(), false)
	if e != nil {
		t.Fatal(e)
	}
	defer held.Close()
	engine, e := transaction.New(fx.home)
	if e != nil {
		t.Fatal(e)
	}
	staged := filepath.Join(t.TempDir(), "j2")
	if e = os.WriteFile(staged, []byte("j2 control"), 0600); e != nil {
		t.Fatal(e)
	}
	nativeTarget := filepath.Join(native, "AGENTS.md")
	digest, e := transaction.DigestTarget(transaction.KindEntry, nativeTarget)
	if e != nil {
		t.Fatal(e)
	}
	_, e = engine.Prepare(held, transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile-record", Identifier: nativeTarget, Kind: transaction.KindEntry, LivePath: nativeTarget, StagedSource: staged, PreimageDigest: digest}}})
	if e == nil || !strings.Contains(e.Error(), envorigin.DiagInvalid) || !strings.Contains(e.Error(), "pending unit") {
		t.Fatalf("intended pending unit overlap refusal: %v", e)
	}
	if nativeSnapshot(t, native) != before {
		t.Fatal("refused J2 changed pending native preimages or stages")
	}
	healthy := filepath.Join(ProfileDir(fx.home, fx.profile), "disjoint-j2-control")
	id := newTransactionID()
	_, e = engine.Prepare(held, transaction.Plan{TransactionID: id, ProjectIdentity: "profiles", Targets: []transaction.Target{{Class: "profile-record", Identifier: healthy, LivePath: healthy, StagedSource: staged, PreimageDigest: transaction.DigestAbsent, PrivateFile: true}}})
	if e != nil {
		t.Fatalf("disjoint J2 prepare: %v", e)
	}
	if e = engine.Commit(held, id); e != nil {
		t.Fatalf("disjoint J2 actual commit: %v", e)
	}
	if raw, e := os.ReadFile(healthy); e != nil || string(raw) != "j2 control" {
		t.Fatal("disjoint owner did not reach effect")
	}
	if nativeSnapshot(t, native) != before {
		t.Fatal("disjoint J2 changed J1 custody")
	}
	if e = engine.Recover(held); e != nil {
		t.Fatalf("fresh native recovery: %v", e)
	}
	if _, e = os.Readlink(nativeTarget); e != nil {
		t.Fatalf("actual J1 native publication did not finish: %v", e)
	}
	pending, e := envauthority.PendingUnits(fx.home)
	if e != nil || len(pending.Units) != 0 {
		t.Fatalf("J1 cleanup: %v %+v", e, pending)
	}
}
