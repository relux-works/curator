package envprofile

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// TestResolvePublishesCredentialMarkerThroughLockedJournal drives Resolve's
// production provisioning path. The marker transaction's before-install event
// observes a sibling staged file while a second manager-home lock acquisition
// is refused.
func TestResolvePublishesCredentialMarkerThroughLockedJournal(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("codex_cli")
	req.Repair = true
	markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
	var reachedMarkerInstall, stagedBesideMarker, secondLockRefused bool
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{
		Observe: func(event transaction.Event) {
			if event.Point != transaction.PointBeforeInstall || event.LivePath != markerPath {
				return
			}
			reachedMarkerInstall = true
			if _, err := os.Lstat(markerPath); !os.IsNotExist(err) {
				t.Errorf("the preimage remains absent before the atomic marker rename: %v", err)
			}
			entries, err := os.ReadDir(filepath.Dir(markerPath))
			if err != nil {
				t.Errorf("read marker directory: %v", err)
			} else {
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".curator-txn-") && strings.HasSuffix(entry.Name(), ".desired") {
						stagedBesideMarker = true
					}
				}
			}
			manager, err := managerlock.New(fx.home)
			if err != nil {
				t.Errorf("construct concurrent lock: %v", err)
				return
			}
			ctx, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
			defer cancel()
			second, err := manager.AcquireHomeOnly(ctx, false)
			if err != nil {
				secondLockRefused = true
				return
			}
			_ = second.Close()
		},
	})}
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if !reachedMarkerInstall || !stagedBesideMarker || !secondLockRefused {
		t.Fatalf("marker publication evidence: install=%t sibling_stage=%t second_lock_refused=%t", reachedMarkerInstall, stagedBesideMarker, secondLockRefused)
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("provision writes one schema-2 record: %+v", marker)
	}
	entry := (*marker.Passthrough)[0]
	if entry.Path != "auth.json" || entry.Isolation != "shared" || entry.Strategy != "keyring-preferred" ||
		entry.SourceRole != "native" || entry.Backend != "file" || entry.BackendVersion != "0.153.2" || entry.Provenance != "provisioned" {
		t.Fatalf("provisioned credential record: %+v", entry)
	}
}

// TestResolveRecoversCredentialMarkerAfterRenameCrash leaves the transaction
// journal and same-directory staging file at the exact before-rename boundary,
// then drives Resolve again to recover and verify the complete v2 marker.
func TestResolveRecoversCredentialMarkerAfterRenameCrash(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("codex_cli")
	req.Repair = true
	markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
	var interrupted bool
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{
		Fault: func(event transaction.Event) error {
			if event.Point == transaction.PointBeforeInstall && event.LivePath == markerPath {
				interrupted = true
				panic("simulated process stop before marker rename")
			}
			return nil
		},
	})}
	func() {
		defer func() { _ = recover() }()
		_, _ = Resolve(req)
	}()
	if !interrupted {
		t.Fatal("the injected interruption reached the marker's before-install boundary")
	}
	if _, err := os.Lstat(markerPath); !os.IsNotExist(err) {
		t.Fatalf("the marker was not installed before the simulated crash: %v", err)
	}
	staged, err := filepath.Glob(filepath.Join(filepath.Dir(markerPath), ".curator-txn-*.desired"))
	if err != nil || len(staged) != 1 {
		t.Fatalf("one same-directory marker stage survives the crash: %v (%v)", staged, err)
	}
	journals, err := os.ReadDir(filepath.Join(fx.home, "state", "transactions", "v1"))
	if err != nil || len(journals) != 1 {
		t.Fatalf("the durable recovery journal survives the crash: %v (%v)", journals, err)
	}
	req.transactionOptions = nil
	if _, err := Resolve(req); err != nil {
		t.Fatalf("the next resolve recovers and verifies the marker: %v", err)
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 || (*marker.Passthrough)[0].Provenance != "provisioned" {
		t.Fatalf("recovery installs the complete intended marker: %+v", marker)
	}
	journals, err = os.ReadDir(filepath.Join(fx.home, "state", "transactions", "v1"))
	if err != nil || len(journals) != 0 {
		t.Fatalf("recovery removes the committed journal: %v (%v)", journals, err)
	}
	staged, err = filepath.Glob(filepath.Join(filepath.Dir(markerPath), ".curator-txn-*.desired"))
	if err != nil || len(staged) != 0 {
		t.Fatalf("recovery removes marker staging files: %v (%v)", staged, err)
	}
}
