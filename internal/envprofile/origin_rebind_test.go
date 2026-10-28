package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

// Shared-lock publication must rebind every already materialized adapter,
// including inactive P homes. Capture bytes survive L0→L1→L2. One home refusal
// leaves the lock and every managed preimage at L0; no partial activation.
// This exercises the actual record publisher, not native/global continuation.
func TestProvisioningOriginManagedGroupRebind(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "conflict"}[conflict], func(t *testing.T) {
			fx, req, raw := recordedOriginFixture(t, "a")
			pi := req
			pi.EnvID = envregistry.Pi
			pi.codexSeedRevisionForTest = ""
			pi.Repair = true
			pi.origin = nil
			piRoot := filepath.Join(t.TempDir(), "agent")
			if e := os.Mkdir(piRoot, 0700); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(filepath.Join(piRoot, "auth.json"), nil, 0600); e != nil {
				t.Fatal(e)
			}
			t.Setenv("PI_CODING_AGENT_DIR", piRoot)
			if result, e := Resolve(pi); e != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("Pi initial actual owner: %v", e)
			}
			target := ManagedHomeDir(fx.home, fx.profile, envregistry.Pi)
			piRaw, e := os.ReadFile(filepath.Join(target, envorigin.Name))
			if e != nil {
				t.Fatal(e)
			}
			if e := SetCurrent(fx.home, "independent-fixture-selection"); e != nil {
				t.Fatal(e)
			}
			if conflict {
				if e := os.Remove(filepath.Join(target, "auth.json")); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(filepath.Join(t.TempDir(), "auth.json"), filepath.Join(target, "auth.json")); e != nil {
					t.Fatal(e)
				}
			}
			before := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
			lock, oldHash, e := readLock(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			for generation := 0; generation < 2; generation++ {
				lock.Members[0].Weight++
				next, e := lock.Canonical()
				if e != nil {
					t.Fatal(e)
				}
				op, e := beginOperation(fx.home)
				if e != nil {
					t.Fatal(e)
				}
				e = op.publish(map[string][]byte{lockPath(fx.home, fx.profile): next})
				_ = op.close()
				if conflict {
					if e == nil || !strings.Contains(e.Error(), envregistry.DiagCredentialConflict) {
						t.Fatalf("intended group refusal: %v", e)
					}
					_, hash, e := readLock(fx.home, fx.profile)
					if e != nil || hash != oldHash || nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != before {
						t.Fatal("group refusal published partial lock/home state")
					}
					return
				}
				if e != nil {
					t.Fatalf("actual group publication: %v", e)
				}
				for _, item := range []struct {
					id      string
					capture []byte
				}{{envregistry.CodexCLI, raw}, {envregistry.Pi, piRaw}} {
					home := ManagedHomeDir(fx.home, fx.profile, item.id)
					marker, e := envmarker.Read(home)
					if e != nil {
						t.Fatal(e)
					}
					hash := contextlock.HashBytes(next)
					if marker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
						t.Fatal("inactive adapter marker was not rebound")
					}
					_, after, e := envorigin.Read(fx.home, envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: home}, item.id)
					if e != nil {
						t.Fatal(e)
					}
					a, _ := envorigin.CaptureFacts(item.capture)
					b, _ := envorigin.CaptureFacts(after)
					if !bytes.Equal(a, b) {
						t.Fatal("group rebind recaptured origin")
					}
				}
			}
		})
	}
}

// A returned error restores the whole L0 unit. An abrupt rename-window stop
// leaves a verifiable exact pending unit; a fresh engine finishes L1 without
// rediscovering native authority. This models restart in-process, not power loss.
func TestProvisioningOriginManagedGroupRecovery(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointAfterBackup, transaction.PointAfterInstall} {
		for _, class := range []string{"profile-record", "origin", "marker"} {
			for _, restart := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/restart=%v", point, class, restart), func(t *testing.T) {
					fx, _, capture := recordedOriginFixture(t, "a")
					lock, oldHash, e := readLock(fx.home, fx.profile)
					if e != nil {
						t.Fatal(e)
					}
					lock.Members[0].Weight++
					next, e := lock.Canonical()
					if e != nil {
						t.Fatal(e)
					}
					before := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
					fault := errors.New("group-publication-fault")
					reached := false
					op, e := beginOperation(fx.home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
						if !reached && event.Point == point && event.Class == class {
							reached = true
							if restart {
								panic("group-stop")
							}
							return fault
						}
						return nil
					}}))
					if e != nil {
						t.Fatal(e)
					}
					func() {
						defer op.close()
						if restart {
							defer func() {
								if v := recover(); v != "group-stop" {
									t.Fatalf("wrong restart signal: %v", v)
								}
							}()
						}
						e = op.publish(map[string][]byte{lockPath(fx.home, fx.profile): next})
					}()
					if !reached {
						t.Fatal("named real publication fault was not reached")
					}
					if !restart {
						if !errors.Is(e, fault) {
							t.Fatalf("wrong returned refusal: %v", e)
						}
						_, got, e := readLock(fx.home, fx.profile)
						if e != nil || got != oldHash || nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != before {
							t.Fatalf("partial L0 rollback: %v", e)
						}
					} else {
						if e := envauthority.Validate(fx.home); e != nil {
							t.Fatalf("genuine pending unit stranded: %v", e)
						}
						t.Setenv("CODEX_HOME", t.TempDir())
						op, e := beginOperation(fx.home)
						if e != nil {
							t.Fatalf("fresh recovery: %v", e)
						}
						op.close()
						_, hash, e := readLock(fx.home, fx.profile)
						if e != nil || hash != contextlock.HashBytes(next) {
							t.Fatalf("recovery did not complete L1: %v", e)
						}
						home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
						_, raw, e := envorigin.Read(fx.home, envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: home}, envregistry.CodexCLI)
						if e != nil {
							t.Fatal(e)
						}
						a, _ := envorigin.CaptureFacts(capture)
						b, _ := envorigin.CaptureFacts(raw)
						if !bytes.Equal(a, b) {
							t.Fatal("restart recaptured authority")
						}
					}
					journals, e := os.ReadDir(filepath.Join(fx.home, "state", "transactions", "v1"))
					if e != nil || len(journals) != 0 {
						t.Fatalf("journal cleanup: %v %v", journals, e)
					}
				})
			}
		}
	}
}
