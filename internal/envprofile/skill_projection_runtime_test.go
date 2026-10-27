package envprofile

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillruntime"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

type runtimeAdmissionFixture struct {
	domain skillruntime.Domain
	lost   bool
}

func (a *runtimeAdmissionFixture) Domain() skillruntime.Domain { return a.domain }
func (a *runtimeAdmissionFixture) Revalidate() error {
	if a.lost {
		return fmt.Errorf("profile_skill_companion_required: admission lost")
	}
	return nil
}
func (a *runtimeAdmissionFixture) Close() error { return nil }
func (a *runtimeAdmissionFixture) Acquire(context.Context, skillruntime.Request) (skillruntime.Admission, error) {
	return a, nil
}

type runtimeChecksFixture struct{}

func (runtimeChecksFixture) Inventory(skillruntime.Binding, []skillruntime.Export) error { return nil }
func (runtimeChecksFixture) BasePATH(skillruntime.Domain, skillruntime.Binding, []skillruntime.Export) error {
	return nil
}
func runtimeSession(t *testing.T, home, profile, current string) (*skillruntime.Session, *runtimeAdmissionFixture) {
	t.Helper()
	user := t.TempDir()
	bin := filepath.Join(user, "bin")
	for _, path := range []string{skillruntime.DomainDir(home), filepath.Join(home, "global", "bin"), bin} {
		if err := privatedir.MakeAll(path); err != nil {
			t.Fatal(err)
		}
	}
	a := &runtimeAdmissionFixture{domain: skillruntime.Domain{Home: home, UserHome: user, UserBin: bin, FixedPATH: bin}}
	session, err := skillruntime.Open(context.Background(), skillruntime.Request{Home: home, Profile: profile, Current: current}, a, runtimeChecksFixture{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, a
}

// The real profile planner joins staged outputs and v5 lock records through
// one operation. A stale lock/current or failed compiler cannot reach staging
// or change any profile/routing preimage. Empty closure here tests the join;
// real script/compiled command staging is exercised by install package tests.
func TestSkillRuntimePlanJoinAndRefusals(t *testing.T) {
	for _, mode := range []string{"valid", "stale-lock", "stale-current", "compile-failure", "staging-failure", "late-declarations", "late-current", "late-source"} {
		t.Run(mode, func(t *testing.T) {
			f := newPreparationFixture(t)
			session, _ := runtimeSession(t, f.home, "profile", "other")
			expected := f.hash
			if mode == "stale-lock" {
				expected = "stale"
			}
			if mode == "stale-current" {
				writeGitFile(t, f.home, "profiles/current", "changed\n")
			}
			before, err := os.ReadFile(lockPath(f.home, "profile"))
			if err != nil {
				t.Fatal(err)
			}
			compiled, staged := false, false
			result, err := PrepareSkillRuntime(f.home, "profile", expected, f.policy, session, func([]*closure.Node) error {
				compiled = true
				if mode == "compile-failure" {
					return fmt.Errorf("profile_skill_compile_failed: deliberate fixture compiler refusal")
				}
				return nil
			}, func(_ *contextlock.Lock, nodes []*closure.Node, w transaction.HomeLock) (skillruntime.Publication, error) {
				staged = true
				if err := w.AssertHeld(); err != nil {
					t.Fatal(err)
				}
				if mode == "staging-failure" {
					return skillruntime.Publication{}, fmt.Errorf("environment_surface_unmanaged_conflict: deliberate fixture refusal")
				}
				if mode == "late-declarations" {
					writeGitFile(t, f.home, "global/Skillfile.json", "{\"schema_version\":1,\"skills\":[]}\n")
				}
				if mode == "late-current" {
					writeGitFile(t, f.home, "profiles/current", "late\n")
				}
				if mode == "late-source" {
					writeGitFile(t, nodes[0].Snapshot, "SKILL.md", "changed after receipt\n")
				}
				stage := t.TempDir()
				path := filepath.Join(stage, "runtime-record")
				if err := os.WriteFile(path, []byte("prepared runtime"), 0600); err != nil {
					t.Fatal(err)
				}
				var plan staging.Plan
				plan.Replace(staging.ClassRuntime, "fixture", filepath.Join(f.home, "runtime", "prepared"), path)
				return skillruntime.Publication{Plan: plan, Verify: session.Revalidate, Diagnostic: skillruntime.Unavailable}, nil
			})
			after, readErr := os.ReadFile(lockPath(f.home, "profile"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if mode == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if !compiled || !staged || bytes.Equal(before, after) || result.Lock.SchemaVersion != 5 || result.Diagnostic != skillruntime.Unavailable {
					t.Fatal("lock/runtime join not reached")
				}
				runtime, err := os.ReadFile(filepath.Join(f.home, "runtime", "prepared"))
				if err != nil || string(runtime) != "prepared runtime" {
					t.Fatalf("runtime %q: %v", runtime, err)
				}
			} else {
				want := map[string]string{"stale-lock": "profile_skill_plan_binding_invalid", "stale-current": "profile_skill_plan_binding_invalid", "compile-failure": "profile_skill_compile_failed", "staging-failure": "environment_surface_unmanaged_conflict", "late-declarations": "profile_skill_plan_binding_invalid", "late-current": "profile_skill_plan_binding_invalid", "late-source": "profile_skill_plan_invalid"}[mode]
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended refusal: %v", err)
				}
				if !bytes.Equal(before, after) {
					t.Fatal("refusal published lock")
				}
				if (mode == "stale-lock" || mode == "stale-current") && (compiled || staged) {
					t.Fatal("stale binding reached compiler")
				}
				if mode == "compile-failure" && staged {
					t.Fatal("failed compile reached stager")
				}
				if _, err := os.Stat(filepath.Join(f.home, "runtime", "prepared")); !os.IsNotExist(err) {
					t.Fatal("refusal published runtime")
				}
			}
			if err := RefuseSkillPlanAdoption(); err == nil || !strings.Contains(err.Error(), skillruntime.CompanionRequired) {
				t.Fatal("adoption gate opened")
			}
			global, err := os.ReadFile(filepath.Join(f.home, "global", "Skillfile.json"))
			expectedGlobal := f.global
			if mode == "late-declarations" {
				expectedGlobal = []byte("{\"schema_version\":1,\"skills\":[]}\n")
			}
			if err != nil || !bytes.Equal(global, expectedGlobal) {
				t.Fatal("unrelated declarations changed")
			}
		})
	}
}

// A process interruption (panic, not an ordinary returned error) leaves real
// durable intent for a new engine. Preparing recovers old state; prepared and
// later points converge to the full new set. Legacy recovery is fenced first.
// These are transaction integration witnesses, not adapter/session qualification.
func TestSkillRuntimeJournalInterruptionAndRecovery(t *testing.T) {
	points := []struct {
		point    transaction.Point
		rollback bool
	}{{"", false}, {transaction.PointDuringStagingCopy, false}, {transaction.PointPrepared, false}, {transaction.PointAfterBackup, false}, {transaction.PointAfterInstall, false}, {transaction.PointBeforeCleanup, false}, {transaction.PointAfterCleanupRename, false}, {transaction.PointAfterRestore, true}, {transaction.PointAfterCleanupRename, true}}
	for _, test := range points {
		point := test.point
		name := string(point)
		if test.rollback {
			name = "rollback-" + name
		}
		if name == "" {
			name = "complete"
		}
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			session, _ := runtimeSession(t, home, "new", "old")
			current := CurrentFile(home)
			lock := lockPath(home, "new")
			runtime := filepath.Join(home, "runtime", "fixture")
			shim := filepath.Join(home, "global", "bin", "tool")
			before := map[string]string{current: "old\n", lock: "old lock", runtime: "old runtime", shim: "old shim"}
			for path, payload := range before {
				writeGitFile(t, filepath.Dir(path), filepath.Base(path), payload)
			}
			engineFactory := newOperationEngine
			defer func() { newOperationEngine = engineFactory }()
			transactions := map[string]bool{}
			planted := false
			newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
				return engineFactory(home, append(options, transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
					transactions[event.TransactionID] = true
					if point != "" && event.Point == point && !planted {
						planted = true
						panic("simulated process interruption")
					}
				}, Fault: func(event transaction.Event) error {
					if test.rollback && event.Point == transaction.PointTargetCommitted {
						return fmt.Errorf("deliberate rollback trigger")
					}
					return nil
				}}))...)
			}
			op, err := beginOperationAdmitted(home, session.Revalidate)
			if err != nil {
				t.Fatal(err)
			}
			stage := t.TempDir()
			var outputs staging.Plan
			for _, target := range []struct{ class, path, payload string }{{staging.ClassRuntime, runtime, "new runtime"}, {staging.ClassCanonicalShim, shim, "new shim"}} {
				path := filepath.Join(stage, filepath.Base(filepath.Dir(target.path))+"-"+filepath.Base(target.path))
				if err := os.WriteFile(path, []byte(target.payload), 0700); err != nil {
					t.Fatal(err)
				}
				outputs.Replace(target.class, target.path, target.path, path)
			}
			interrupted := false
			func() {
				defer func() {
					if recover() != nil {
						interrupted = true
					}
				}()
				err = op.publishCombined(map[string][]byte{current: []byte("new\n"), lock: []byte("new lock")}, nil, nil, skillruntime.Publication{Plan: outputs})
			}()
			_ = op.close()
			newOperationEngine = engineFactory
			if point != "" && (!planted || !interrupted) {
				t.Fatalf("plant not reached: %v", err)
			}
			if point == "" && err != nil {
				t.Fatal(err)
			}
			if len(transactions) != 1 {
				t.Fatalf("runtime and records used %d journals", len(transactions))
			}
			observed := map[string][]byte{}
			for path := range before {
				observed[path], _ = os.ReadFile(path)
			}
			if denied, err := beginOperation(home); err == nil {
				_ = denied.close()
				t.Fatal("legacy recovery admitted")
			}
			for path, payload := range observed {
				now, _ := os.ReadFile(path)
				if !bytes.Equal(now, payload) {
					t.Fatal("legacy refusal wrote a live path")
				}
			}
			recovered, err := beginOperationAdmitted(home, session.Revalidate)
			if err != nil {
				t.Fatal(err)
			}
			_ = recovered.close()
			want := map[string]string{current: "new\n", lock: "new lock", runtime: "new runtime", shim: "new shim"}
			if test.rollback || point == transaction.PointDuringStagingCopy {
				want = before
			}
			for path, payload := range want {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != payload {
					t.Fatalf("%s: got %q, want %q: %v", path, got, payload, err)
				}
			}
			fresh, err := beginOperationAdmitted(home, session.Revalidate)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.recovered {
				t.Fatal("recovery not idempotent")
			}
			_ = fresh.close()
		})
	}
}

// Admission loss is a no-write condition even during rollback/cleanup. A new
// admitted owner can recover the retained journal without compiling again.
func TestSkillRuntimeLostAdmissionFencesRollback(t *testing.T) {
	home := t.TempDir()
	session, admission := runtimeSession(t, home, "profile", "profile")
	live := filepath.Join(home, "global", "bin", "tool")
	writeGitFile(t, filepath.Dir(live), filepath.Base(live), "old")
	original := newOperationEngine
	defer func() { newOperationEngine = original }()
	newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
		return original(home, append(options, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
			if event.Point == transaction.PointAfterInstall {
				admission.lost = true
				return fmt.Errorf("deliberate post-swap failure")
			}
			return nil
		}}))...)
	}
	op, err := beginOperationAdmitted(home, session.Revalidate)
	if err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(t.TempDir(), "new")
	if err := os.WriteFile(stage, []byte("new"), 0700); err != nil {
		t.Fatal(err)
	}
	var plan staging.Plan
	plan.Replace(staging.ClassCanonicalShim, "tool", live, stage)
	err = op.publishCombined(nil, nil, nil, skillruntime.Publication{Plan: plan})
	_ = op.close()
	if err == nil || !strings.Contains(err.Error(), "admission lost") {
		t.Fatalf("lost-admission refusal: %v", err)
	}
	got, err := os.ReadFile(live)
	if err != nil || string(got) != "new" {
		t.Fatalf("unadmitted rollback wrote: %q %v", got, err)
	}
	newOperationEngine = original
	if recovered, err := beginOperationAdmitted(home, session.Revalidate); err == nil {
		_ = recovered.close()
		t.Fatal("lost-admission recovery admitted")
	}
	admission.lost = false
	recovered, err := beginOperationAdmitted(home, session.Revalidate)
	if err != nil {
		t.Fatal(err)
	}
	_ = recovered.close()
	got, err = os.ReadFile(live)
	if err != nil || string(got) != "new" {
		t.Fatalf("inconsistent recovered preimage: %q %v", got, err)
	}
}

// A child package test starts with fixture "native" roots outside its HOME.
// TestMain must replace each override, so a regression would fail this probe
// before any product test could address those roots.
func TestSkillRuntimePackageHomeIsolation(t *testing.T) {
	if os.Getenv("CURATOR_TEST_HOME_PROBE") == "1" {
		home := os.Getenv("HOME")
		if !strings.HasPrefix(filepath.Base(home), "curator-package-home-") {
			t.Fatal("HOME not pinned by TestMain")
		}
		for _, key := range []string{"USERPROFILE", "CODEX_HOME", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_STATE_HOME", "PI_CODING_AGENT_DIR", "CURATOR_CONFIG", "CURATOR_SYSTEM_CONFIG", "CURATOR_GLOBAL_USER_BIN"} {
			path := os.Getenv(key)
			rel, err := filepath.Rel(home, path)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				t.Fatalf("%s escapes package HOME", key)
			}
		}
		return
	}
	poisoned := t.TempDir()
	for _, name := range []string{".claude", ".codex", ".config", ".curator"} {
		if err := os.Mkdir(filepath.Join(poisoned, name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(poisoned, name, "sentinel"), []byte("untouched"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	image, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(image, "-test.run=^TestSkillRuntimePackageHomeIsolation$")
	keys := map[string]bool{"HOME": true, "USERPROFILE": true, "CODEX_HOME": true, "CLAUDE_CONFIG_DIR": true, "XDG_CONFIG_HOME": true, "XDG_STATE_HOME": true, "PI_CODING_AGENT_DIR": true, "CURATOR_CONFIG": true, "CURATOR_SYSTEM_CONFIG": true, "CURATOR_GLOBAL_USER_BIN": true, "CURATOR_TEST_HOME_PROBE": true}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !keys[key] {
			child.Env = append(child.Env, entry)
		}
	}
	for key := range keys {
		value := poisoned
		if key == "CURATOR_TEST_HOME_PROBE" {
			value = "1"
		}
		child.Env = append(child.Env, key+"="+value)
	}
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("HOME isolation probe: %v %s", err, out)
	}
	for _, name := range []string{".claude", ".codex", ".config", ".curator"} {
		raw, err := os.ReadFile(filepath.Join(poisoned, name, "sentinel"))
		if err != nil || string(raw) != "untouched" {
			t.Fatal("inherited root modified")
		}
	}
}

// Malformed durable intent is never guessed or discarded. A valid recovered
// journal is the nearby positive control in TestSkillRuntimeJournalInterruptionAndRecovery.
func TestSkillRuntimeCorruptRecoveryKeepsDomainClosed(t *testing.T) {
	home := t.TempDir()
	session, _ := runtimeSession(t, home, "profile", "profile")
	writeGitFile(t, home, "profiles/current", "profile\n")
	writeGitFile(t, home, "global/bin/tool", "old shim")
	journals := filepath.Join(home, "state", "transactions", "v1")
	if err := privatedir.MakeAll(journals); err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(journals, "interrupted.json")
	if err := os.WriteFile(journal, []byte(`{"unrecognized":"intent"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if op, err := beginOperationAdmitted(home, session.Revalidate); err == nil {
		_ = op.close()
		t.Fatal("corrupt recovery admitted")
	} else if !strings.Contains(err.Error(), "profile_skill_recovery_failed") {
		t.Fatalf("wrong recovery refusal: %v", err)
	}
	for path, want := range map[string]string{CurrentFile(home): "profile\n", filepath.Join(home, "global/bin/tool"): "old shim", journal: `{"unrecognized":"intent"}`} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("corrupt recovery changed %s: %q %v", path, got, err)
		}
	}
}

// A registered isolated domain fences legacy direct-current/global paths as
// well as the ordinary journal owner, before their callbacks or writes run.
func TestSkillRuntimeLegacyDirectWritersRefuse(t *testing.T) {
	home := t.TempDir()
	session, _ := runtimeSession(t, home, "profile", "profile")
	writeGitFile(t, home, "profiles/current", "profile\n")
	if err := SetCurrent(home, "foreign"); err == nil || !strings.Contains(err.Error(), skillruntime.RoutingUnproven) {
		t.Fatalf("direct current admitted: %v", err)
	}
	op, err := beginOperationAdmitted(home, session.Revalidate)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = op.close() }()
	called := false
	_, err = GlobalSkillTransaction(home, Policy{}, envregistry.MachineConfig{}, nil, nil, op.lock, func() error { called = true; return nil })
	if err == nil || !strings.Contains(err.Error(), skillruntime.RoutingUnproven) || called {
		t.Fatalf("legacy global continuation admitted: %v", err)
	}
	got, err := os.ReadFile(CurrentFile(home))
	if err != nil || string(got) != "profile\n" {
		t.Fatal("legacy writer changed current")
	}
}

// A private domain without legacy project-lease state is admitted by the journal
// controls above. Existing or unreadable lease lifecycle state cannot be adopted
// by this separate supervisor and is refused before recovery or publication.
func TestSkillRuntimeRefusesLegacyProjectLeaseState(t *testing.T) {
	home := t.TempDir()
	session, _ := runtimeSession(t, home, "profile", "profile")
	writeGitFile(t, home, "profiles/current", "profile\n")
	lease := filepath.Join(home, "state", "project-leases", "v1", "ledger.json")
	writeGitFile(t, filepath.Dir(lease), filepath.Base(lease), "unadopted lease bytes")
	if op, err := beginOperationAdmitted(home, session.Revalidate); err == nil {
		_ = op.close()
		t.Fatal("legacy lifecycle admitted")
	} else if !strings.Contains(err.Error(), skillruntime.RoutingUnproven) {
		t.Fatalf("wrong isolation refusal: %v", err)
	}
	for path, want := range map[string]string{CurrentFile(home): "profile\n", lease: "unadopted lease bytes"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("isolation refusal changed %s: %q %v", path, got, err)
		}
	}
}
