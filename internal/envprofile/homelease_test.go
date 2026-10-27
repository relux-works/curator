package envprofile

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease"
	"github.com/relux-works/curator/internal/homelease/state"
)

func isolateHomeLeaseWriters(t *testing.T) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if runtime.GOOS != "windows" {
		if err := os.Symlink(git, filepath.Join(bin, "git")); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", bin)
	} else {
		// Keep Git's executable and DLL installation intact, while excluding
		// ambient provider launchers from the Windows process search path.
		t.Setenv("PATH", filepath.Dir(git))
		for _, provider := range []string{"claude", "codex", "muse", "opencode", "pi"} {
			if _, err := exec.LookPath(provider); err == nil {
				t.Skip("Git directory also exposes a provider; isolated writer fixture unavailable")
			}
		}
	}
	pinOperatorHome(t, t.TempDir())
	pinHomes(t)
}

func leasedManagedFixture(t *testing.T) (*managedFixture, *homelease.Manager, *homelease.Record) {
	t.Helper()
	isolateHomeLeaseWriters(t)
	fx, _ := captureFixture(t)
	t.Setenv("CLAUDE_CONFIG_DIR", fx.native["claude_code"])
	t.Setenv("CODEX_HOME", fx.native["codex_cli"])
	xdg := t.TempDir()
	fx.native["opencode-native"] = filepath.Join(xdg, "opencode")
	if err := os.MkdirAll(fx.native["opencode-native"], 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("PI_CODING_AGENT_DIR", fx.native["pi"])
	m, err := HomeLeaseManager(fx.request(envregistry.CodexCLI), nil)
	if err != nil {
		t.Fatal(err)
	}
	req := homelease.Request{Environment: "codex", Profile: fx.profile, LockSHA256: strings.TrimPrefix(fx.lockHash, "sha256:"), RequestID: "writer-request", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "fixture-manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "writer-request"}}
	out, reason, err := m.Acquire(context.Background(), req)
	if err != nil || reason != "acquired" || !out.Changed || out.Lease.Environment != envregistry.CodexCLI {
		t.Fatalf("real materializer admission: %+v %s %v", out, reason, err)
	}
	return fx, m, out.Lease
}

func requireHomeLeaseRefusal(t *testing.T, err error, id string) {
	t.Helper()
	var r *state.Refusal
	if !errors.As(err, &r) || r.Code != "environment_home_leased" || r.Reason != "blocking_holders" || r.Details.Count != 1 || len(r.Details.Holders) != 1 || r.Details.Holders[0].LeaseID != id {
		t.Fatalf("expected one named blocking holder %s, got %v", id, err)
	}
}

func breakFixtureLease(t *testing.T, m *homelease.Manager, r *homelease.Record) {
	t.Helper()
	out, reason, err := m.Break(context.Background(), r.LeaseID, r.Revision, true, "protection deliberately abandoned for isolated fixture")
	if err != nil || reason != "broken" || !out.Changed {
		t.Fatalf("operator control: %+v %s %v", out, reason, err)
	}
}

// HL03/08/14 and W01/04/05/07/12: these real entry points refuse a prepared
// reservation before changing fixture bytes, links, modes, pointers or audit.
// Breaking that exact reservation permits the same operation's real writes.
func TestHomeLeaseManagedWriters(t *testing.T) {
	writers := map[string]func(*managedFixture) error{
		"W01-repair": func(f *managedFixture) error {
			r := f.request(envregistry.CodexCLI)
			r.Repair = true
			_, err := Resolve(r)
			return err
		},
		"W01-preserve-lock": func(f *managedFixture) error {
			r := f.request(envregistry.CodexCLI)
			r.Repair, r.PreserveLock = true, true
			_, err := Resolve(r)
			return err
		},
		"W04-update": func(f *managedFixture) error {
			_, _, err := UpdateWithOptions(f.home, f.profile, UpdateOptions{})
			return err
		},
		"W05-use": func(f *managedFixture) error {
			_, err := UseWithPolicy(f.home, f.profile, envregistry.CodexCLI, "", false, Policy{})
			return err
		},
		"W05-clear-scoped": func(f *managedFixture) error {
			_, err := UseWithPolicy(f.home, "", envregistry.CodexCLI, "", true, Policy{})
			return err
		},
		"W05-current":        func(f *managedFixture) error { return SetCurrent(f.home, f.profile) },
		"W05-scoped-current": func(f *managedFixture) error { return SetScoped(f.home, "env:codex_cli", f.profile, false) },
		"W07-managed-sync": func(f *managedFixture) error {
			_, err := SyncManagedHomesWithPolicy(f.home, Policy{}, envregistry.DefaultMachineConfig())
			return err
		},
		"W07-command-sync": func(f *managedFixture) error {
			_, _, a, b := SyncHomesWithPolicy(f.home, Policy{}, envregistry.DefaultMachineConfig())
			return errors.Join(a, b)
		},
	}
	for name, write := range writers {
		t.Run(name, func(t *testing.T) {
			fx, m, lease := leasedManagedFixture(t)
			if strings.HasPrefix(name, "W01-") {
				path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "AGENTS.md")
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("drift"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := captureFixtureState(t, fx)
			requireHomeLeaseRefusal(t, write(fx), lease.LeaseID)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("refusal changed the preimage")
			}
			breakFixtureLease(t, m, lease)
			if err := write(fx); err != nil {
				t.Fatalf("unleased positive control: %v", err)
			}
		})
	}
}

// HL03/17: healthy leased homes still resolve and capture read-only under a
// borrowed real lock; corrupt lease authority refuses without a fragment.
func TestHomeLeaseReadOnlyMaterialization(t *testing.T) {
	fx, req := captureFixture(t)
	m, err := HomeLeaseManager(fx.request(envregistry.CodexCLI), nil)
	if err != nil {
		t.Fatal(err)
	}
	r := homelease.Request{Environment: envregistry.CodexCLI, Profile: fx.profile, LockSHA256: strings.TrimPrefix(fx.lockHash, "sha256:"), RequestID: "read-only", Holder: homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "manager", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "read-only"}}
	if _, _, err := m.Acquire(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	if got, err := Resolve(fx.request(envregistry.CodexCLI)); err != nil || len(got.Document) == 0 {
		t.Fatalf("resolve: %v", err)
	}
	if _, err := CaptureMaterialize(req); err != nil {
		t.Fatalf("borrowed capture: %v", err)
	}
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("read-only control changed state")
	}
	if err := os.WriteFile(state.LedgerPath(fx.home), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	before, _ = captureFixtureState(t, fx)
	got, err := Resolve(fx.request(envregistry.CodexCLI))
	var refusal *state.Refusal
	if got != nil || !errors.As(err, &refusal) || refusal.Code != "environment_lease_unreadable" {
		t.Fatalf("corrupt authority produced fragment: %v %v", got, err)
	}
	after, _ = captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("corrupt read changed state")
	}
}

// W09: a shared projection cannot be rebuilt under a consumer's reservation;
// an already verified projection is a genuine no-write control.
func TestHomeLeaseSharedProjection(t *testing.T) {
	fx, m, lease := leasedManagedFixture(t)
	_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	skills, err := skillsOf(fx.home, newGitManager(fx.home), lock)
	if err != nil || len(skills) == 0 {
		t.Fatalf("skills: %v", err)
	}
	op, err := beginOperationGuarded(fx.home, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	before := captureTree(t, fx.home)
	if err := publishSkillProjection(op, fx.home, skills[0]); err != nil {
		t.Fatalf("verified no-write control: %v", err)
	}
	if !reflect.DeepEqual(before, captureTree(t, fx.home)) {
		t.Fatal("verified projection wrote state")
	}
	path := filepath.Join(skills[0].root, "SKILL.md")
	if err := os.WriteFile(path, []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	before = captureTree(t, fx.home)
	requireHomeLeaseRefusal(t, publishSkillProjection(op, fx.home, skills[0]), lease.LeaseID)
	if !reflect.DeepEqual(before, captureTree(t, fx.home)) {
		t.Fatal("projection refusal changed preimage")
	}
	_ = op.close()
	breakFixtureLease(t, m, lease)
	op, err = beginOperation(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = op.close() }()
	if err := publishSkillProjection(op, fx.home, skills[0]); err != nil {
		t.Fatal(err)
	}
	if err := verifySkillProjection(fx.home, skills[0]); err != nil {
		t.Fatalf("released projection not restored: %v", err)
	}
}
