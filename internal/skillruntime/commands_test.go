package skillruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

type testAdmission struct {
	domain  Domain
	refusal error
	closed  bool
}

func (a *testAdmission) Domain() Domain { return a.domain }
func (a *testAdmission) Revalidate() error {
	if a.closed {
		return fmt.Errorf("admission closed")
	}
	return a.refusal
}
func (a *testAdmission) Close() error { a.closed = true; return nil }

type testSupervisor struct{ admission *testAdmission }

func (s testSupervisor) Acquire(context.Context, Request) (Admission, error) { return s.admission, nil }

type testChecks struct {
	inventoryErr, baseErr error
	seen                  []Export
}

func (c *testChecks) Inventory(_ Binding, exports []Export) error {
	c.seen = append([]Export(nil), exports...)
	return c.inventoryErr
}
func (c *testChecks) BasePATH(Domain, Binding, []Export) error { return c.baseErr }

func runtimeFixture(t *testing.T) (*Session, *testAdmission, *testChecks, Binding) {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	user := filepath.Join(root, "operator")
	for _, path := range []string{DomainDir(home), filepath.Join(home, "global", "bin"), filepath.Join(user, "bin")} {
		if err := privatedir.MakeAll(path); err != nil {
			t.Fatal(err)
		}
	}
	plan := &contextlock.SkillPlan{SchemaVersion: 1, Nodes: []contextlock.SkillPin{}, Edges: []contextlock.SkillEdge{}, DirectConstraints: []contextlock.DirectConstraint{}}
	planHash, err := plan.Hash()
	if err != nil {
		t.Fatal(err)
	}
	hash, err := contextlock.Write(filepath.Join(home, "profiles", "profile", "lock.json"), &contextlock.Lock{SchemaVersion: 5, HashVersion: 2, Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "profile", SkillPlan: plan, SkillPlanSHA256: planHash, Members: []contextlock.Member{{Kind: "context", Name: "profile", Version: "1.0.0", StateHash: strings.Repeat("a", 64), RequiredBy: []string{}}}})
	if err != nil {
		t.Fatal(err)
	}
	a := &testAdmission{domain: Domain{Home: home, UserHome: user, UserBin: filepath.Join(user, "bin"), FixedPATH: filepath.Join(user, "bin")}}
	checks := &testChecks{}
	session, err := Open(context.Background(), Request{Home: home, Profile: "profile", Current: "profile"}, testSupervisor{a}, checks)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, a, checks, Binding{Profile: "profile", LockHash: hash, PlanHash: planHash}
}

// Portable static keys reject case/suffix aliases before OS-specific shim
// staging. Neighbour names remain valid, so rejecting all exports cannot pass.
func TestCommandNamesAndOwners(t *testing.T) {
	for _, name := range []string{"tool", "git.helper", "curatorial", "other.ps1.exe"} {
		if err := CheckNames([]Export{{Name: name, Skill: "a", Profile: "p"}}); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"", "../tool", "tool/name", "tööl"} {
		if err := CheckNames([]Export{{Name: name, Skill: "a", Profile: "p"}}); err == nil || !strings.Contains(err.Error(), "profile_skill_command_invalid") {
			t.Fatalf("invalid export %q: %v", name, err)
		}
	}
	for _, name := range []string{"curator", "CURATOR-run", "CuRaToR-x.EXE.CmD", "curator.PS1"} {
		if err := CheckNames([]Export{{Name: name, Skill: "a", Profile: "p"}}); err == nil || !strings.Contains(err.Error(), "environment_reserved_command_name") {
			t.Fatalf("reservation %q: %v", name, err)
		}
	}
	for _, owners := range [][2]Export{
		{{Name: "Tool", Skill: "a", Profile: "p"}, {Name: "tool.exe.cmd", Skill: "b", Profile: "p"}},
		{{Name: "tool", Skill: "a", Profile: "p"}, {Name: "TOOL", Skill: "a", Profile: "other"}},
	} {
		if err := CheckNames(owners[:]); err == nil || !strings.Contains(err.Error(), "profile_skill_command_conflict") {
			t.Fatalf("collision: %v", err)
		}
	}
	if NameKey("Git.EXE.CmD") != "git" || NameKey("git.helper") != "git.helper" {
		t.Fatal("wrong portable normalization")
	}
}

// Missing/changed native authority and companion checks refuse read-only.
// These are adapter contract fixtures, not supervisor/runtime qualification.
func TestAdmissionAndCompanionRefusals(t *testing.T) {
	for _, mode := range []string{"valid", "missing-supervisor", "missing-checks", "lost-admission", "changed-domain", "no-fixed-path", "missing-boundary", "inventory", "base-path"} {
		t.Run(mode, func(t *testing.T) {
			s, a, checks, binding := runtimeFixture(t)
			var supervisor Supervisor = testSupervisor{a}
			var companion CommandChecks = checks
			switch mode {
			case "missing-supervisor":
				supervisor = nil
			case "missing-checks":
				companion = nil
			case "lost-admission":
				a.refusal = fmt.Errorf("profile_skill_companion_required: unknown consumer")
			case "changed-domain":
				a.domain.Home = filepath.Join(t.TempDir(), "foreign")
			case "no-fixed-path":
				a.domain.FixedPATH = "/usr/bin"
			case "missing-boundary":
				if err := os.Remove(DomainDir(s.domain.Home)); err != nil {
					t.Fatal(err)
				}
			case "inventory":
				checks.inventoryErr = fmt.Errorf("environment_reserved_command_name: inventory witness")
			case "base-path":
				checks.baseErr = fmt.Errorf("profile_skill_command_conflict: admitted base tool")
			}
			opened, err := Open(context.Background(), s.request, supervisor, companion)
			if err == nil {
				_, err = opened.Preflight(binding, []Export{{Name: "unique-tool", Skill: "a", Profile: "profile"}})
			}
			if mode == "valid" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				want := RoutingUnproven
				switch mode {
				case "missing-checks":
					want = CompanionRequired
				case "inventory":
					want = "environment_reserved_command_name"
				case "base-path":
					want = "profile_skill_command_conflict"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended admission refusal %q: %v", want, err)
				}
			}
			entries, readErr := os.ReadDir(filepath.Join(s.domain.Home, "global", "bin"))
			if readErr != nil || len(entries) != 0 {
				t.Fatal("admission refusal changed root")
			}
		})
	}
}

func seedOwned(t *testing.T, s *Session, binding Binding, export Export) OwnedFile {
	t.Helper()
	path := filepath.Join(s.domain.Home, "global", "bin", export.Name)
	if err := os.WriteFile(path, []byte("owned preimage\n"), 0700); err != nil {
		t.Fatal(err)
	}
	digest, err := transaction.DigestTarget(transaction.KindBytes, path)
	if err != nil {
		t.Fatal(err)
	}
	owned := OwnedFile{Path: path, Digest: digest, Export: export, Binding: binding}
	payload, err := json.Marshal(routingState{Schema: 1, Binding: binding, Files: []OwnedFile{owned}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath(s.domain.Home), payload, 0600); err != nil {
		t.Fatal(err)
	}
	return owned
}

// A protected byte preimage is required even for stale/retained names. Unknown
// files (including directories, broken links and metadata impostors) cannot be
// silently removed by the ordinary stager's filename-based enumeration.
func TestClosedRootInventory(t *testing.T) {
	for _, mode := range []string{"empty", "owned", "unmanaged", "directory", "broken-link", "metadata", "changed-owned", "retained-collision", "forged-record-key", "forged-plan-binding", "forged-owner-binding"} {
		t.Run(mode, func(t *testing.T) {
			s, _, _, binding := runtimeFixture(t)
			selected := []Export{{Name: "new-tool", Skill: "selected", Profile: "profile"}}
			root := filepath.Join(s.domain.Home, "global", "bin")
			switch mode {
			case "owned", "changed-owned":
				owned := seedOwned(t, s, binding, Export{Name: "old-tool", Skill: "old", Profile: "profile"})
				if mode == "changed-owned" {
					if err := os.WriteFile(owned.Path, []byte("foreign"), 0700); err != nil {
						t.Fatal(err)
					}
				}
			case "retained-collision":
				seedOwned(t, s, binding, Export{Name: "NEW-tool.exe", Skill: "direct"})
			case "unmanaged":
				if err := os.WriteFile(filepath.Join(root, "manual"), []byte("unmanaged"), 0700); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(filepath.Join(root, "manual"), 0700); err != nil {
					t.Fatal(err)
				}
			case "broken-link":
				if err := os.Symlink(filepath.Join(t.TempDir(), "absent"), filepath.Join(root, "manual")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			case "metadata":
				if err := os.WriteFile(filepath.Join(s.domain.UserBin, ".curator-managed.json"), []byte(`{"schema_version":1,"entries":[]}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "forged-plan-binding", "forged-owner-binding":
				seedOwned(t, s, binding, Export{Name: "old-tool", Skill: "old", Profile: "profile"})
				payload, err := os.ReadFile(statePath(s.domain.Home))
				if err != nil {
					t.Fatal(err)
				}
				var state routingState
				if err := json.Unmarshal(payload, &state); err != nil {
					t.Fatal(err)
				}
				if mode == "forged-plan-binding" {
					state.Binding.PlanHash = "sha256:" + strings.Repeat("c", 64)
				} else {
					state.Files[0].Binding.PlanHash = "sha256:" + strings.Repeat("c", 64)
				}
				payload, err = json.Marshal(state)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(statePath(s.domain.Home), payload, 0600); err != nil {
					t.Fatal(err)
				}
			case "forged-record-key":
				seedOwned(t, s, binding, Export{Name: "old-tool", Skill: "old", Profile: "profile"})
				raw, err := os.ReadFile(statePath(s.domain.Home))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(statePath(s.domain.Home), []byte(strings.Replace(string(raw), `"schema"`, `"SCHEMA"`, 1)), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := s.Preflight(binding, selected)
			if mode == "empty" || mode == "owned" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				want := "environment_surface_unmanaged_conflict"
				if mode == "retained-collision" {
					want = "profile_skill_command_conflict"
				}
				if strings.HasPrefix(mode, "forged-") {
					want = "profile_skill_plan_binding_invalid"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended %s refusal: %v", mode, err)
				}
			}
		})
	}
}

// The transition carries retained bytes opaquely and rejects an omitted active
// export or stale removal. It cannot publish a second live command transaction.
func TestBindRequiresCompleteTransition(t *testing.T) {
	for _, mode := range []string{"complete", "missing-selected", "missing-stale", "retained", "retained-replacement"} {
		t.Run(mode, func(t *testing.T) {
			s, _, _, binding := runtimeFixture(t)
			retained := mode == "retained" || mode == "retained-replacement"
			oldExport := Export{Name: "old-tool", Skill: "old", Profile: "profile"}
			if retained {
				oldExport.Profile = ""
			}
			old := seedOwned(t, s, binding, oldExport)
			selected := []Export{{Name: "new-tool", Skill: "new", Profile: "profile"}}
			inventory, err := s.Preflight(binding, selected)
			if err != nil {
				t.Fatal(err)
			}
			stage := t.TempDir()
			path := filepath.Join(stage, "new")
			if err := os.WriteFile(path, []byte("new executable"), 0700); err != nil {
				t.Fatal(err)
			}
			var plan staging.Plan
			if mode != "missing-selected" {
				plan.Replace(staging.ClassCanonicalShim, "new-tool", filepath.Join(s.domain.Home, "global", "bin", "new-tool"), path)
			}
			if mode != "missing-stale" {
				plan.Remove("old-tool", old.Path)
			}
			if mode == "retained-replacement" {
				plan.Targets = nil
				plan.Replace(staging.ClassCanonicalShim, "old-tool", old.Path, path)
			}
			publication, err := s.Bind(stage, binding, selected, inventory, plan, nil)
			if mode == "complete" || mode == "retained" {
				if err != nil {
					t.Fatal(err)
				}
				if err := publication.Verify(); err != nil {
					t.Fatal(err)
				}
				if retained {
					for _, target := range publication.Plan.Targets {
						if target.LivePath == old.Path {
							t.Fatal("retained path journaled")
						}
					}
				}
			} else {
				want := "profile_skill_plan_binding_invalid"
				if mode == "retained-replacement" {
					want = "profile_skill_command_conflict"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended transition refusal: %v", err)
				}
			}
			raw, err := os.ReadFile(old.Path)
			if err != nil || string(raw) != "owned preimage\n" {
				t.Fatal("planning changed live preimage")
			}
		})
	}
}
