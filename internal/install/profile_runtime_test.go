package install

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/skillruntime"
)

// The fixture supplies an explicitly trusted adapter contract. It does not
// represent a real supervisor or qualify a session/launcher capability.
type profileRuntimeAdmission struct{ domain skillruntime.Domain }

func (a *profileRuntimeAdmission) Domain() skillruntime.Domain { return a.domain }
func (a *profileRuntimeAdmission) Revalidate() error           { return nil }
func (a *profileRuntimeAdmission) Close() error                { return nil }
func (a *profileRuntimeAdmission) Acquire(_ context.Context, request skillruntime.Request) (skillruntime.Admission, error) {
	if request.Home != a.domain.Home {
		return nil, fmt.Errorf("profile_skill_companion_required: wrong home")
	}
	return a, nil
}

type profileRuntimeChecks struct{ inventoryErr, baseErr error }

func (c profileRuntimeChecks) Inventory(skillruntime.Binding, []skillruntime.Export) error {
	return c.inventoryErr
}
func (c profileRuntimeChecks) BasePATH(skillruntime.Domain, skillruntime.Binding, []skillruntime.Export) error {
	return c.baseErr
}
func profileDomain(t *testing.T, e *env) *profileRuntimeAdmission {
	t.Helper()
	user := t.TempDir()
	bin := filepath.Join(user, "private-bin")
	for _, path := range []string{skillruntime.DomainDir(e.home), filepath.Join(e.home, "global", "bin"), bin} {
		if err := privatedir.MakeAll(path); err != nil {
			t.Fatal(err)
		}
	}
	return &profileRuntimeAdmission{domain: skillruntime.Domain{Home: e.home, UserHome: user, UserBin: bin, FixedPATH: bin + string(os.PathListSeparator) + os.Getenv("PATH")}}
}

func signedRuntimeProfile(t *testing.T, e *env, profile, skill, command string, build, broken bool) (string, envprofile.Policy) {
	t.Helper()
	if build {
		e.buildSkill(skill, command)
	} else {
		e.skill(skill)
		repo := filepath.Join(e.skillsRoot, skill)
		raw, err := os.ReadFile(filepath.Join(repo, "csk-skill.json"))
		if err != nil {
			t.Fatal(err)
		}
		var spec map[string]any
		if err := json.Unmarshal(raw, &spec); err != nil {
			t.Fatal(err)
		}
		commands := spec["commands"].(map[string]any)
		old := commands[skill+"-tool"]
		delete(commands, skill+"-tool")
		commands[command] = old
		payload, err := json.Marshal(spec)
		if err != nil {
			t.Fatal(err)
		}
		e.write(repo, "agent-skill.json", string(payload))
		if err := os.Remove(filepath.Join(repo, "csk-skill.json")); err != nil {
			t.Fatal(err)
		}
	}
	repo := filepath.Join(e.skillsRoot, skill)
	if broken {
		e.write(repo, "src/cmd/"+command+"/main.go", "package main\nfunc main() { broken syntax }\n")
	}
	key := filepath.Join(t.TempDir(), "signer")
	if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key).CombinedOutput(); err != nil {
		t.Fatalf("fixture signer: %v %s", err, out)
	}
	e.git(repo, "add", ".")
	e.git(repo, "-c", "gpg.format=ssh", "-c", "gpg.ssh.program=ssh-keygen", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
	pin, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	pub, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	gitConfig := filepath.Join(t.TempDir(), "gitconfig")
	e.write(filepath.Dir(gitConfig), filepath.Base(gitConfig), fmt.Sprintf("[url %q]\n insteadOf = https://example.com/%s\n", draftReplayFileURL(repo), skill))
	// Keep earlier identities when a test prepares two profiles.
	if old := os.Getenv("GIT_CONFIG_GLOBAL"); old != "" {
		if raw, err := os.ReadFile(old); err == nil {
			file, err := os.OpenFile(gitConfig, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			_, err = file.Write(raw)
			_ = file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := t.TempDir()
	e.write(root, "agent-context.json", fmt.Sprintf(`{"schema_version":1,"name":%q,"version":"1.0.0","requires":{"skills":{%q:{"git":%q,"revision":%q}}}}`, profile, skill, "https://example.com/"+skill, pin.Commit))
	e.write(root, "context/main.md", "# Profile\n")
	_, state, err := contextstore.EnsureState(e.home, "context", profile, root)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := contextlock.Write(filepath.Join(e.home, "profiles", profile, "lock.json"), &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: profile, Members: []contextlock.Member{{Kind: "context", Name: profile, Version: "1.0.0", StateHash: state, RequiredBy: []string{}}}})
	if err != nil {
		t.Fatal(err)
	}
	return hash, envprofile.Policy{RequireSourceSigners: true, SourceSigners: map[string][]contextresolve.Signer{"example.com/" + skill: {{Type: "ssh", Key: strings.TrimSpace(string(pub))}}}}
}

// Actual signed source admission, ordinary script/build staging and both shim
// hops reach a private final process. Inactive profiles publish no shims and
// report B1. Stale pins, unmanaged roots, companion refusals and failed native
// compilation leave lock/current/root/declaration preimages unchanged.
// POSIX executes the final child; Windows proves staging only. This is an
// install witness, not final-agent PATH or real supervisor proof.
func TestProfileRuntimePublication(t *testing.T) {
	modes := []string{"current-script", "inactive-script", "stale-lock", "unmanaged", "reserved", "inventory-refusal", "base-path-refusal", "missing-supervisor"}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		modes = append(modes, "current-build", "compiler-failure")
	}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			e := newEnv(t)
			command := "slice-tool"
			if mode == "reserved" {
				command = "CuRaToR-secret.EXE.CmD"
			}
			build := mode == "current-build" || mode == "compiler-failure"
			hash, policy := signedRuntimeProfile(t, e, "profile", "slice-skill", command, build, mode == "compiler-failure")
			current := "profile"
			if mode == "inactive-script" {
				current = "other"
			}
			e.write(e.home, "profiles/current", current+"\n")
			e.write(e.home, "global/Skillfile.json", "{\"schema_version\":1,\"skills\":[]}\n")
			admission := profileDomain(t, e)
			var supervisor skillruntime.Supervisor = admission
			checks := profileRuntimeChecks{}
			switch mode {
			case "stale-lock":
				hash = "stale"
			case "unmanaged":
				e.write(e.home, "global/bin/manual", "unmanaged bytes")
			case "inventory-refusal":
				checks.inventoryErr = fmt.Errorf("environment_reserved_command_name: protected inventory")
			case "base-path-refusal":
				checks.baseErr = fmt.Errorf("profile_skill_command_conflict: admitted base PATH")
			case "missing-supervisor":
				supervisor = nil
			}
			path := filepath.Join(e.home, "profiles/profile/lock.json")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			result, err := PrepareProfileSkillRuntime(e.cfg, "profile", hash, policy, Options{}, supervisor, checks)
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			positive := mode == "current-script" || mode == "inactive-script" || mode == "current-build"
			if positive {
				if err != nil {
					t.Fatal(err)
				}
				if result.Lock.SchemaVersion != 5 || bytes.Equal(before, after) {
					t.Fatal("runtime did not publish v5")
				}
				// Public adoption remains unavailable even with prepared shims and
				// an admitted internal publisher; the refusal preserves both records.
				if _, err := envprofile.UseWithPolicy(e.home, "profile", "", "", false, policy); err == nil || !strings.Contains(err.Error(), skillruntime.CompanionRequired) {
					t.Fatalf("public adoption gate: %v", err)
				}
				if mode == "inactive-script" {
					if result.Diagnostic != skillruntime.Unavailable {
						t.Fatalf("B1: %q", result.Diagnostic)
					}
					for _, root := range []string{filepath.Join(e.home, "global/bin"), admission.domain.UserBin} {
						entries, err := os.ReadDir(root)
						if err != nil || len(entries) != 0 {
							t.Fatal("inactive profile got shims")
						}
					}
				} else {
					canonical, err := runtimestore.NewManagedShim(runtimestore.GlobalCanonicalShim, filepath.Join(e.home, "global/bin"), command, runtimestore.Platform())
					if err != nil {
						t.Fatal(err)
					}
					forward, err := runtimestore.NewManagedShim(runtimestore.SafeForwardingShim, admission.domain.UserBin, command, runtimestore.Platform())
					if err != nil {
						t.Fatal(err)
					}
					if _, err := os.Stat(canonical.Path()); err != nil {
						t.Fatal(err)
					}
					if runtime.GOOS != "windows" {
						child := exec.Command(forward.Path())
						child.Env = append(os.Environ(), "PATH="+admission.domain.FixedPATH)
						out, err := child.CombinedOutput()
						if err != nil {
							t.Fatalf("real private shim child: %v %s", err, out)
						}
						if !build && !strings.Contains(string(out), "slice-skill") {
							t.Fatalf("wrong reached script: %q", out)
						}
					}
					// Repeat from the exact v5 pin, with no new publication authority.
					if _, err := PrepareProfileSkillRuntime(e.cfg, "profile", result.LockHash, policy, Options{}, supervisor, checks); err != nil {
						t.Fatalf("pinned repeat: %v", err)
					}
				}
			} else {
				want := map[string]string{"stale-lock": "profile_skill_plan_binding_invalid", "unmanaged": "environment_surface_unmanaged_conflict", "reserved": "environment_reserved_command_name", "inventory-refusal": "environment_reserved_command_name", "base-path-refusal": "profile_skill_command_conflict", "missing-supervisor": "profile_skill_routing_unproven", "compiler-failure": "profile_skill_compile_failed"}[mode]
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended refusal %q: %v", want, err)
				}
				if !bytes.Equal(before, after) {
					t.Fatal("refusal published lock")
				}
				if entries, err := os.ReadDir(admission.domain.UserBin); err != nil || len(entries) != 0 {
					t.Fatal("refusal published forwarders")
				}
				if mode == "unmanaged" {
					raw, err := os.ReadFile(filepath.Join(e.home, "global/bin/manual"))
					if err != nil || string(raw) != "unmanaged bytes" {
						t.Fatal("unmanaged preimage changed")
					}
				}
			}
			now, err := os.ReadFile(filepath.Join(e.home, "profiles/current"))
			if err != nil || string(now) != current+"\n" {
				t.Fatal("preparation moved current")
			}
			direct, err := os.ReadFile(filepath.Join(e.home, "global/Skillfile.json"))
			if err != nil || string(direct) != "{\"schema_version\":1,\"skills\":[]}\n" {
				t.Fatal("direct declarations changed")
			}
			if err := envprofile.RefuseSkillPlanAdoption(); err == nil {
				t.Fatal("adoption gate opened")
			}
			// Legacy global install/forwarder refresh cannot recover or mutate this
			// routing domain without the admitted coordinator.
			if global := Global(e.cfg, admission.domain.UserHome, Options{}); global.Status != "failed" {
				t.Fatalf("legacy global writer admitted: %+v", global)
			}
			messages := globalbins.Refresh(e.home, map[string]bool{}, runtimestore.Platform(), nil, admission.domain.UserHome)
			if len(messages) != 1 || !strings.Contains(messages[0], skillruntime.RoutingUnproven) {
				t.Fatalf("legacy refresh admitted: %v", messages)
			}
		})
	}
}

// A prior owner inventory is consumed through exact preimages. After an
// externally established current-profile change, reconciliation removes both
// stale shim hops. The pointer+shim atomic join itself is covered in envprofile;
// public v5 Use remains gated pending PR3's adapter/supervisor integration.
func TestProfileRuntimeRemovesPriorOwnerShims(t *testing.T) {
	e := newEnv(t)
	hash, policy := signedRuntimeProfile(t, e, "first", "first-skill", "first-tool", false, false)
	e.write(e.home, "profiles/current", "first\n")
	admission := profileDomain(t, e)
	if _, err := PrepareProfileSkillRuntime(e.cfg, "first", hash, policy, Options{}, admission, profileRuntimeChecks{}); err != nil {
		t.Fatal(err)
	}
	hash, policy = signedRuntimeProfile(t, e, "second", "second-skill", "second-tool", false, false)
	// Isolated fixture simulates the integration owner's successful adapter phase.
	e.write(e.home, "profiles/current", "second\n")
	if _, err := PrepareProfileSkillRuntime(e.cfg, "second", hash, policy, Options{}, admission, profileRuntimeChecks{}); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{filepath.Join(e.home, "global/bin"), admission.domain.UserBin} {
		old, err := runtimestore.NewManagedShim(runtimestore.GlobalCanonicalShim, root, "first-tool", runtimestore.Platform())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(old.Path()); !os.IsNotExist(err) {
			t.Fatal("prior-profile shim survived")
		}
		next, err := runtimestore.NewManagedShim(runtimestore.GlobalCanonicalShim, root, "second-tool", runtimestore.Platform())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(next.Path()); err != nil {
			t.Fatal("new current shim absent")
		}
	}
}

// Unrelated direct routing is retained by protected byte/generation records,
// without acquiring its declaration or rewriting either shim hop. This seeds
// trusted retained-owner state; adoption of historical global installs is not
// supplied by this test or by PR2.
func TestProfileRuntimePreservesRetainedDirectExports(t *testing.T) {
	e := newEnv(t)
	hash, policy := signedRuntimeProfile(t, e, "first", "retained-skill", "retained-tool", false, false)
	e.write(e.home, "profiles/current", "first\n")
	admission := profileDomain(t, e)
	if _, err := PrepareProfileSkillRuntime(e.cfg, "first", hash, policy, Options{}, admission, profileRuntimeChecks{}); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(skillruntime.DomainDir(e.home), "routing.json")
	raw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		Schema  int                      `json:"schema"`
		Binding skillruntime.Binding     `json:"binding"`
		Files   []skillruntime.OwnedFile `json:"files"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	retained := map[string][]byte{}
	for i := range state.Files {
		file := &state.Files[i]
		if file.Export.Name == "retained-tool" {
			file.Export.Profile = ""
			retained[file.Path], err = os.ReadFile(file.Path)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(retained) != 2 {
		t.Fatalf("retained shim hops=%d", len(retained))
	}
	raw, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	hash, policy = signedRuntimeProfile(t, e, "second", "selected-skill", "selected-tool", false, false)
	direct := `{"schema_version":1,"skills":[{"name":"retained-skill","git":"https://unrelated.invalid/retained","revision":"1111111111111111111111111111111111111111"}]}`
	e.write(e.home, "global/Skillfile.json", direct)
	e.write(e.home, "profiles/current", "second\n")
	result, err := PrepareProfileSkillRuntime(e.cfg, "second", hash, policy, Options{}, admission, profileRuntimeChecks{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Drift) != 1 || result.Drift[0].Name != "retained-skill" {
		t.Fatalf("unrelated drift=%+v", result.Drift)
	}
	for path, before := range retained {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("retained export rewritten: %s %v", path, err)
		}
	}
	after, err := os.ReadFile(filepath.Join(e.home, "global/Skillfile.json"))
	if err != nil || string(after) != direct {
		t.Fatal("unrelated declaration rewritten")
	}
}

// A child package test starts with fixture "native" roots outside its HOME.
// TestMain must replace each override, so a regression would fail this probe
// before any product test could address those roots.
func TestProfileRuntimePackageHomeIsolation(t *testing.T) {
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
	child := exec.Command(image, "-test.run=^TestProfileRuntimePackageHomeIsolation$")
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
