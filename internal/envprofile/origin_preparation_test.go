package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
)

// Initial capture and durable preparation precede every live managed parent.
// Returned faults at parent and marker publication restore original absence;
// the nearby healthy control reaches a fragment and the actual origin record.
// Only raw account facts are injected; this does not qualify an OS account.
func TestProvisioningOriginFirstMutationJournaled(t *testing.T) {
	for _, mode := range []string{"healthy", "parent-fault", "marker-fault"} {
		t.Run(mode, func(t *testing.T) {
			enableV1WritersForTest(t)
			fx := writeManagedFixture(t, "first-mutation")
			native := fx.native["codex_cli"]
			t.Setenv("CODEX_HOME", native)
			if err := os.WriteFile(filepath.Join(native, "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			anchor := t.TempDir()
			req := fx.request("codex_cli")
			req.Repair = true
			captured, prepared, faulted := false, false, false
			req.originAccountLookup = func() (envorigin.Account, error) {
				captured = true
				if _, err := os.Lstat(EnvRoot(fx.home)); !os.IsNotExist(err) {
					t.Fatalf("managed parent appeared before capture: %v", err)
				}
				return envorigin.Account{UID: uint64(os.Geteuid()), RealUID: uint64(os.Getuid()), EffectiveUID: uint64(os.Geteuid()), Home: anchor}, nil
			}
			planted := errors.New("first-mutation-journal-fault")
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointPrepared {
					raw, err := os.ReadFile(filepath.Join(fx.home, "state", "transactions", "v1", e.TransactionID+".json"))
					if err != nil {
						t.Fatal(err)
					}
					var j transaction.Journal
					if err := json.Unmarshal(raw, &j); err != nil {
						t.Fatal(err)
					}
					if j.Unit != nil && len(j.Unit.Homes) > 0 {
						prepared = true
						if !captured {
							t.Fatal("journal preceded native capture")
						}
						if _, err := os.Lstat(EnvRoot(fx.home)); !os.IsNotExist(err) {
							t.Fatalf("preflight mutated managed parent before journal: %v", err)
						}
					}
				}
				if mode == "parent-fault" && e.Point == transaction.PointAfterDirectoryInstall && e.LivePath == EnvRoot(fx.home) || mode == "marker-fault" && e.Point == transaction.PointAfterInstall && e.Class == "marker" {
					faulted = true
					return planted
				}
				return nil
			}})}
			result, err := Resolve(req)
			if !prepared {
				t.Fatalf("actual home preparation not reached: %v", err)
			}
			if mode == "healthy" {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("healthy continuation: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envorigin.Name)); err != nil {
					t.Fatal(err)
				}
			} else {
				if !faulted || !errors.Is(err, planted) || result != nil && len(result.Document) > 0 {
					t.Fatalf("intended fault/no fragment: %t %v", faulted, err)
				}
				if _, err := os.Lstat(EnvRoot(fx.home)); !os.IsNotExist(err) {
					t.Fatalf("failure did not restore parent absence: %v", err)
				}
			}
		})
	}
}

// Existing owned parent protection is itself a durable metadata action. It
// leaves inode identity intact, and faults restore all original modes rather
// than losing an in-memory chmod undo at the publication boundary.
func TestProvisioningOriginParentProtectionJournaled(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "after-home-protection"}[fail], func(t *testing.T) {
			enableV1WritersForTest(t)
			fx := writeManagedFixture(t, "parent-protection")
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if err := os.MkdirAll(target, 0755); err != nil {
				t.Fatal(err)
			}
			parents := []string{EnvRoot(fx.home), filepath.Dir(target), target}
			pins := map[string]os.FileInfo{}
			for _, path := range parents {
				if err := os.Chmod(path, 0755); err != nil {
					t.Fatal(err)
				}
				pins[path], _ = os.Lstat(path)
			}
			t.Setenv("CODEX_HOME", fx.native["codex_cli"])
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			req := fx.request("codex_cli")
			req.Repair = true
			anchor := t.TempDir()
			req.originAccountLookup = func() (envorigin.Account, error) {
				return envorigin.Account{UID: uint64(os.Geteuid()), RealUID: uint64(os.Getuid()), EffectiveUID: uint64(os.Geteuid()), Home: anchor}, nil
			}
			hit := false
			planted := errors.New("parent-protection-fault")
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
				if e.Point == transaction.PointPrepared {
					for _, path := range parents {
						info, err := os.Lstat(path)
						if err != nil || info.Mode().Perm() != 0755 {
							t.Fatalf("parent protected before journal: %v", err)
						}
					}
				}
				if e.Point == transaction.PointAfterDirectoryInstall && e.LivePath == target {
					hit = true
					if fail {
						return planted
					}
				}
				return nil
			}})}
			result, err := Resolve(req)
			if !hit {
				t.Fatalf("actual protection owner not reached: %v", err)
			}
			if fail && (!errors.Is(err, planted) || result != nil && len(result.Document) > 0) || !fail && (err != nil || result == nil || len(result.Document) == 0) {
				t.Fatalf("protection disposition: %v", err)
			}
			for _, path := range parents {
				info, e := os.Lstat(path)
				want := os.FileMode(0700)
				if fail {
					want = 0755
				}
				if e != nil || info.Mode().Perm() != want || !os.SameFile(pins[path], info) {
					t.Fatalf("parent identity/mode not restored or protected: %s %v", path, e)
				}
			}
		})
	}
}
