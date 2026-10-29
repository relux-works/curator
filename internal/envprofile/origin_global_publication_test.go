package envprofile

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/transaction"
)

// A real joined global publication must survive a consumer mutating only its
// callback snapshot. Repeated commit checks see the original closed legacy
// inventory, and the actual env-file continuation still publishes. This is an
// in-process alias-isolation control, not an OS-crash or hostile-UID witness.
func TestProvisioningOriginGlobalCallbackSnapshot(t *testing.T) {
	fx, resolver := recordedGlobalOriginFixture(t)
	live := filepath.Join(fx.home, "global", "env.sh")
	if err := os.MkdirAll(filepath.Dir(live), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(live, []byte("old env"), 0644); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(t.TempDir(), "env.sh")
	if err := os.WriteFile(stage, []byte("new env"), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := transaction.DigestTarget(transaction.KindBytes, live)
	if err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	plan := transaction.Plan{Legacy: &envunit.LegacyPublication{UserHome: userHome, Entries: []envunit.LegacyTarget{{Class: "50-env-file", Path: live}}}, Targets: []transaction.Target{{Class: "50-env-file", Identifier: live, Kind: transaction.KindBytes, LivePath: live, StagedSource: stage, PreimageDigest: before}}}
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	}()
	saved := newOperationEngine
	calls := 0
	newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
		return transaction.New(home, append(options, transaction.WithCommitCheck(func(lock transaction.HomeLock, journal transaction.Journal) error {
			if err := projectlease.TransactionCheck(home)(lock, journal); err != nil {
				return err
			}
			calls++
			if journal.Unit == nil || journal.Unit.Legacy == nil || journal.Unit.Legacy.UserHome != userHome || len(journal.Unit.Legacy.Entries) != 1 || journal.Unit.Legacy.Entries[0].Class != "50-env-file" || journal.Unit.Legacy.Entries[0].Path != live {
				t.Fatal("commit callback snapshot changed the closed legacy inventory")
			}
			journal.Unit.Legacy.UserHome = "poisoned callback snapshot"
			journal.Unit.Legacy.Entries[0].Class = "poisoned callback entry"
			return nil
		}))...)
	}
	t.Cleanup(func() { newOperationEngine = saved })
	if _, err := PublishGlobalSkillPlan(fx.home, Policy{}, envregistry.DefaultMachineConfig(), nil, resolver, held, plan); err != nil {
		t.Fatalf("callback snapshot affected actual joined publication: %v", err)
	}
	if calls < 2 {
		t.Fatalf("did not reach repeated actual commit checks: %d", calls)
	}
	if raw, err := os.ReadFile(live); err != nil || string(raw) != "new env" {
		t.Fatalf("actual joined continuation did not publish env file: %v", err)
	}
}

// The real joined executor publishes actual env-file targets with the profile
// group. Faults after the env write and after a home marker must restore L0 and
// every native/managed preimage. A confirmed handled conflict instead keeps L1
// with those same preimages, including the untouched legacy file. The stop is
// an in-process model followed by a fresh engine, not an OS crash witness.
func TestProvisioningOriginJoinedGlobalPublication(t *testing.T) {
	for _, mode := range []string{"healthy", "legacy-returned", "marker-stop", "retained"} {
		t.Run(mode, func(t *testing.T) {
			fx, resolver := recordedGlobalOriginFixture(t)
			if mode == "retained" {
				for _, name := range []string{"AGENTS.md", envmarker.Name} {
					if err := os.Remove(filepath.Join(fx.native["codex_cli"], name)); err != nil && !os.IsNotExist(err) {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "AGENTS.md"), []byte("unmanaged native"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			live := filepath.Join(fx.home, "global", "env.sh")
			if err := os.MkdirAll(filepath.Dir(live), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(live, []byte("old env"), 0644); err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(t.TempDir(), "env.sh")
			if err := os.WriteFile(stage, []byte("new env"), 0644); err != nil {
				t.Fatal(err)
			}
			before, err := transaction.DigestTarget(transaction.KindBytes, live)
			if err != nil {
				t.Fatal(err)
			}
			plan := transaction.Plan{Legacy: &envunit.LegacyPublication{UserHome: t.TempDir(), Entries: []envunit.LegacyTarget{{Class: "50-env-file", Path: live}}}, Targets: []transaction.Target{{Class: "50-env-file", Identifier: live, Kind: transaction.KindBytes, LivePath: live, StagedSource: stage, PreimageDigest: before}}}
			managed := nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile))
			native := map[string]string{}
			for id, root := range fx.native {
				native[id] = nativeSnapshot(t, root)
			}
			manager, err := managerlock.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			held, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := held.Close(); err != nil {
					t.Error(err)
				}
			}()
			saved := newOperationEngine
			hit := false
			if mode == "legacy-returned" || mode == "marker-stop" {
				newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
					return saved(home, append(options, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
						if event.Point == transaction.PointAfterInstall && (mode == "legacy-returned" && event.LivePath == live || mode == "marker-stop" && event.Class == "marker") {
							hit = true
							if data, _ := os.ReadFile(live); string(data) != "new env" {
								t.Fatal("fault did not reach actual legacy continuation")
							}
							if mode == "marker-stop" {
								panic("joined-global-stop")
							}
							return errors.New("joined legacy returned fault")
						}
						return nil
					}}))...)
				}
				t.Cleanup(func() { newOperationEngine = saved })
			}
			func() {
				defer func() {
					if value := recover(); value != nil && value != "joined-global-stop" {
						panic(value)
					}
				}()
				_, err = PublishGlobalSkillPlan(fx.home, Policy{}, envregistry.DefaultMachineConfig(), nil, resolver, held, plan)
			}()
			newOperationEngine = saved
			if mode == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
			} else if mode == "retained" {
				var retained *transaction.RetainedGlobalConflict
				if !errors.As(err, &retained) || retained.Cleanup != nil {
					t.Fatalf("expected retained conflict: %v", err)
				}
			} else if !hit || mode == "legacy-returned" && (err == nil || !strings.Contains(err.Error(), "joined legacy returned fault")) {
				t.Fatalf("intended fault not reached: hit=%t err=%v", hit, err)
			}
			fresh, err := saved(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			if err = fresh.Recover(held); err != nil {
				t.Fatal(err)
			}
			_, hash, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			if (hash != fx.lockHash) != (mode == "healthy" || mode == "retained") {
				t.Fatalf("wrong lock disposition: %s", hash)
			}
			data, err := os.ReadFile(live)
			if err != nil {
				t.Fatal(err)
			}
			want := "old env"
			if mode == "healthy" {
				want = "new env"
			}
			if string(data) != want {
				t.Fatalf("legacy pre/postimage: %q", data)
			}
			if mode != "healthy" {
				if nativeSnapshot(t, filepath.Join(EnvRoot(fx.home), fx.profile)) != managed {
					t.Fatal("ordinary failure/conflict changed managed home")
				}
				for id, root := range fx.native {
					if nativeSnapshot(t, root) != native[id] {
						t.Fatalf("ordinary failure/conflict changed native %s", id)
					}
				}
			}
			if err = envauthority.Validate(fx.home); err != nil {
				t.Fatal(err)
			}
			identity, err := envunit.ObserveManager(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			ledger, _, err := envlifecycle.Read(identity)
			if err != nil {
				t.Fatal(err)
			}
			if (len(ledger.Entries) == 1) != (mode == "retained") {
				t.Fatalf("unexpected lifecycle disposition: %+v", ledger)
			}
		})
	}
}
