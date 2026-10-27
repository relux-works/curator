package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/stateread"
)

type preparationFixture struct {
	home, hash, repo, pin, rootEntry string
	global                           []byte
	policy                           Policy
}

func newPreparationFixture(t *testing.T, variants ...string) *preparationFixture {
	t.Helper()
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	key, pub := statusSignerKey(t)
	files := map[string]string{"SKILL.md": "---\nname: memory\ndescription: Memory fixture\n---\n# Memory\n", "agent-skill.json": `{"schema_version":4,"capabilities":{},"commands":{}}`}
	if len(variants) > 0 && variants[0] == "audit-failure" {
		files["SKILL.md"] += "Read https://api.example.net/guide\n"
	}
	repo := gitRepo(t, files, "v1")
	gitRun(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
	pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	source := ids.serve(repo, "https://example.com/memory")
	root := t.TempDir()
	writePackage(t, root, "profile", "1.0.0", "fixture\n")
	raw := fmt.Sprintf(`{"schema_version":1,"name":"profile","version":"1.0.0","requires":{"skills":{"memory":{"git":%q,"revision":%q}}}}`, source, pin)
	writeGitFile(t, root, "agent-context.json", raw)
	entry, state, err := contextstore.EnsureState(home, "context", "profile", root)
	if err != nil {
		t.Fatal(err)
	}
	lock := &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: "profile", Members: []contextlock.Member{{Kind: "context", Name: "profile", Version: "1.0.0", StateHash: state, RequiredBy: []string{}}, {Kind: "skill", Name: "unrelated", Source: "example.com/unrelated", Commit: strings.Repeat("2", 40), RequiredBy: []string{}}}}
	hash, err := contextlock.Write(lockPath(home, "profile"), lock)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := marshalSource(Source{Kind: KindPath, Path: root})
	if err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, home, "profiles/profile/source.json", string(payload))
	global := []byte("{\n  \"schema_version\": 1, \"skills\": [{\"name\":\"unrelated\",\"git\":\"https://unreachable.example/keep\",\"revision\":\"" + strings.Repeat("3", 40) + "\"}]\n}\n")
	writeGitFile(t, home, "global/Skillfile.json", string(global))
	writeGitFile(t, home, "global/skills/unrelated/kept", "drifting bytes\n")
	writeGitFile(t, home, "profiles/current", "other\n")
	return &preparationFixture{home: home, hash: hash, repo: repo, pin: pin, rootEntry: entry, global: global, policy: Policy{RequireSourceSigners: true, SourceSigners: map[string][]contextresolve.Signer{"example.com/memory": {{Type: "ssh", Key: pub}}}}}
}

// PSP01/02/05, PSR01/03/05: real signed source acquisition and strict profile
// audit precede a preparation callback and the record journal. Success writes
// a complete v5 lock and retains unrelated state; each paired refusal leaves
// the old lock/current/unrelated artifact byte-identical. The callback failure
// isolates the manager publication boundary; install tests run the compiler.
func TestSkillPlanPreparationPublication(t *testing.T) {
	for _, mode := range []string{"success", "stale-lock", "compile-failure", "declaration-race", "source-race", "context-race", "state-tamper", "wrong-signer", "wrong-cache-origin", "audit-failure", "overlap-conflict", "selected-store-tamper"} {
		t.Run(mode, func(t *testing.T) {
			f := newPreparationFixture(t, mode)
			expected := f.hash
			diagnostic := ""
			calls := 0
			switch mode {
			case "stale-lock":
				expected = "sha256:" + strings.Repeat("0", 64)
				diagnostic = "profile_skill_plan_binding_invalid"
			case "compile-failure":
				diagnostic = "compiler fixture failed"
			case "declaration-race", "source-race", "context-race", "state-tamper", "selected-store-tamper":
				diagnostic = "profile_skill_plan_invalid"
			case "wrong-cache-origin":
				diagnostic = "profile_source_invalid"
			case "audit-failure":
				diagnostic = "audit.capability.network-undeclared"
			case "wrong-signer":
				_, pub := statusSignerKey(t)
				f.policy.SourceSigners["example.com/memory"] = []contextresolve.Signer{{Type: "ssh", Key: pub}}
				diagnostic = "context_source_signer_rejected"
			case "overlap-conflict":
				f.global = []byte(fmt.Sprintf(`{"schema_version":1,"skills":[{"name":"memory","git":"https://example.com/memory","revision":%q}]}`, strings.Repeat("2", 40)))
				writeGitFile(t, f.home, "global/Skillfile.json", string(f.global))
				diagnostic = "profile_skill_pin_conflict"
			}
			if mode == "wrong-cache-origin" {
				cache := newSkillPlanGitManager(f.home).repoDir("example.com/memory")
				if err := gitops.Clone("https://example.com/memory", cache); err != nil {
					t.Fatal(err)
				}
				gitRun(t, cache, "remote", "set-url", "origin", "https://example.com/other")
			}
			if mode == "state-tamper" {
				writeGitFile(t, f.rootEntry, "context/a.md", "tampered state\n")
			}
			if mode == "selected-store-tamper" {
				entry, err := contextstore.EnsureGit(f.home, "skill", "memory", f.repo, f.pin)
				if err != nil {
					t.Fatal(err)
				}
				writeGitFile(t, entry, "references/tamper.md", "tamper\n")
			}
			old, _ := os.ReadFile(lockPath(f.home, "profile"))
			current, _ := os.ReadFile(CurrentFile(f.home))
			unrelatedPath := filepath.Join(f.home, "global/skills/unrelated/kept")
			unrelated, _ := os.ReadFile(unrelatedPath)
			result, err := PrepareSkillPlan(f.home, "profile", expected, f.policy, func(nodes []*closure.Node) error {
				calls++
				if len(nodes) != 1 || nodes[0].Name != "memory" {
					t.Fatal("wrong selected closure")
				}
				during, _ := os.ReadFile(lockPath(f.home, "profile"))
				if !bytes.Equal(old, during) {
					t.Fatal("lock published before preparation")
				}
				if mode == "compile-failure" {
					return errors.New("compiler fixture failed")
				}
				if mode == "declaration-race" {
					writeGitFile(t, f.home, "global/Skillfile.json", `{"schema_version":1,"skills":[]}`)
				}
				if mode == "context-race" {
					writeGitFile(t, f.rootEntry, "context/a.md", "changed during preparation")
				}
				if mode == "source-race" {
					writeGitFile(t, nodes[0].Snapshot, "references/changed.md", "changed during preparation")
				}
				return nil
			})
			now, _ := os.ReadFile(lockPath(f.home, "profile"))
			if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if calls != 1 || bytes.Equal(old, now) || result.Lock.SchemaVersion != 5 {
					t.Fatal("prepared lock was not published")
				}
				if err := result.Lock.VerifySkillBinding(result.LockHash); err != nil {
					t.Fatal(err)
				}
				retained, _ := result.Lock.Find("skill", "unrelated")
				if retained.Commit != strings.Repeat("2", 40) || retained.Directory != "." || len(result.Lock.SkillPlan.Nodes) != 1 {
					t.Fatal("selective member retention failed")
				}
				if len(result.Drift) != 1 || result.Drift[0].Name != "unrelated" || result.Drift[0].State != "unknown" || result.Drift[0].Declared.Value != strings.Repeat("3", 40) || result.Drift[0].Observed == nil || *result.Drift[0].Observed != "unreadable" {
					t.Fatalf("drift: %+v", result.Drift)
				}
				previous, _ := os.ReadFile(prevLockPath(f.home, "profile"))
				if !bytes.Equal(previous, old) {
					t.Fatal("rollback lock missing")
				}
				loaded, _, readErr := readLock(f.home, "profile")
				if readErr != nil || loaded.SchemaVersion != 5 {
					t.Fatalf("inspect v5: %v", readErr)
				}
				if _, useErr := UseWithPolicy(f.home, "profile", "", "", false, Policy{}); useErr == nil || !strings.Contains(useErr.Error(), "profile_skill_companion_required") {
					t.Fatalf("adoption refusal: %v", useErr)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), diagnostic) {
					t.Fatalf("expected %s, got %v", diagnostic, err)
				}
				wantCalls := 1
				switch mode {
				case "stale-lock", "state-tamper", "wrong-signer", "wrong-cache-origin", "audit-failure", "overlap-conflict":
					wantCalls = 0
				}
				if calls != wantCalls {
					t.Fatalf("preparation reached %d times, want %d", calls, wantCalls)
				}
				if !bytes.Equal(old, now) {
					t.Fatal("refusal published lock")
				}
			}
			afterCurrent, _ := os.ReadFile(CurrentFile(f.home))
			afterUnrelated, _ := os.ReadFile(unrelatedPath)
			if !bytes.Equal(current, afterCurrent) || !bytes.Equal(unrelated, afterUnrelated) {
				t.Fatal("routing/unrelated preimage changed")
			}
			afterGlobal, _ := os.ReadFile(filepath.Join(f.home, "global/Skillfile.json"))
			if mode != "declaration-race" && !bytes.Equal(afterGlobal, f.global) {
				t.Fatal("direct declaration bytes changed")
			}
			if _, err := os.Stat(filepath.Join(f.home, "global/bin")); !os.IsNotExist(err) {
				t.Fatal("command publication occurred")
			}
		})
	}
}

// PSR02: read-only unrelated observation distinguishes absent drift and unknown and
// never acquires the declaration's unreachable source or changes its bytes.
func TestSkillPlanUnrelatedDriftObservation(t *testing.T) {
	home := gitFixtureHome(t)
	direct := []manifest.Decl{{Name: "absent"}, {Name: "unreadable"}}
	writeGitFile(t, home, "global/skills/unreadable", "not a directory")
	got := skillPlanDrift(home, nil, direct)
	if len(got) != 2 || got[0].Name != "absent" || got[0].State != "drift" || got[0].Observed != nil || got[1].State != "unknown" || got[1].Observed == nil || *got[1].Observed != "unreadable" {
		t.Fatalf("drift: %+v", got)
	}
	if err := RefuseSkillPlanAdoption(); err == nil || !strings.Contains(err.Error(), "profile_skill_companion_required") {
		t.Fatalf("companion refusal: %v", err)
	}
}

// PSR02: a valid independent installation is observed as in-sync; tampering
// reports drift without repairing bytes. Missing/corrupt marker is unknown.
func TestSkillPlanUnrelatedCurrentness(t *testing.T) {
	home := gitFixtureHome(t)
	entry := filepath.Join(home, "global/skills/other")
	writeGitFile(t, entry, "SKILL.md", "context\n")
	hash, err := hashing.ContentSHA256WithVersion(entry, nil, hashing.WriteVersion())
	if err != nil {
		t.Fatal(err)
	}
	m := &marker.Marker{Name: "other", Source: "other", Git: "https://example.com/other", RefKind: "tag", Ref: "v1", Commit: strings.Repeat("1", 40), ContentSHA256: hash, InstalledAt: "2026-10-27T00:00:00Z"}
	if err := marker.Write(entry, m); err != nil {
		t.Fatal(err)
	}
	direct := []manifest.Decl{{Name: "other", Git: m.Git, Ref: manifest.Ref{Kind: "tag", Value: "v1"}}}
	if got := skillPlanDrift(home, nil, direct); len(got) != 1 || got[0].State != "in-sync" || got[0].Declared != direct[0].Ref || got[0].Observed == nil || *got[0].Observed != m.Commit {
		t.Fatalf("in-sync: %+v", got)
	}
	original := direct[0].Ref
	direct[0].Ref = manifest.Ref{Kind: "revision", Value: strings.Repeat("2", 40)}
	if got := skillPlanDrift(home, nil, direct); got[0].State != "drift" || got[0].Observed == nil || *got[0].Observed != m.Commit {
		t.Fatalf("different declared pin: %+v", got)
	}
	direct[0].Ref = original
	writeGitFile(t, entry, "SKILL.md", "changed\n")
	if got := skillPlanDrift(home, nil, direct); got[0].State != "drift" {
		t.Fatalf("drift: %+v", got)
	}
	payload, _ := os.ReadFile(filepath.Join(entry, "SKILL.md"))
	if string(payload) != "changed\n" {
		t.Fatal("drift repaired bytes")
	}
	writeGitFile(t, entry, marker.Name, "{")
	if got := skillPlanDrift(home, nil, direct); got[0].State != "unknown" {
		t.Fatalf("unknown: %+v", got)
	}
}

// A prepared v5 lock remains inspectable, but legacy publication paths cannot
// downgrade or activate it. Existing legacy success controls run in this suite.
func TestSkillPlanLegacyWritersRefusePreparedLock(t *testing.T) {
	f := newPreparationFixture(t)
	if _, err := PrepareSkillPlan(f.home, "profile", f.hash, f.policy, func([]*closure.Node) error { return nil }); err != nil {
		t.Fatal(err)
	}
	old, _ := os.ReadFile(lockPath(f.home, "profile"))
	tests := map[string]func() error{
		"update":           func() error { _, _, err := UpdateWithPolicy(f.home, "profile", f.policy); return err },
		"native-preflight": func() error { return preflightInPlaceScope(f.home, "profile", "", f.policy, nil) },
		"managed-preflight": func() error {
			return preflightManagedProfile(f.home, "profile", f.policy, envregistry.MachineConfig{}, nil)
		},
		"materialize": func() error {
			_, err := materializeScopeWithNativeHome(nil, f.home, "profile", "", f.policy, nil)
			return err
		},
		"skills": func() error {
			lock, _, err := readLock(f.home, "profile")
			if err != nil {
				return err
			}
			_, err = skillsOf(f.home, newGitManager(f.home), lock)
			return err
		},
		"resolve": func() error {
			_, err := Resolve(ResolveRequest{Home: f.home, Profile: "profile", EnvID: "codex_cli", Policy: f.policy})
			return err
		},
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil || !strings.Contains(err.Error(), "profile_skill_companion_required") {
				t.Fatalf("companion gate: %v", err)
			}
			after, _ := os.ReadFile(lockPath(f.home, "profile"))
			if !bytes.Equal(old, after) {
				t.Fatal("legacy writer changed v5 lock")
			}
		})
	}
}

// PSR02: a missing leaf under a real directory is proven absence; the same
// lookup behind a broken parent symlink stays unknown and never repairs it.
func TestSkillPlanDriftBrokenAncestor(t *testing.T) {
	for _, broken := range []bool{false, true} {
		name := "real-parent"
		if broken {
			name = "broken-parent-link"
		}
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			parent := filepath.Join(home, "global", "skills")
			if err := os.MkdirAll(filepath.Dir(parent), 0700); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(home, "missing")
			if broken {
				if err := os.Symlink(target, parent); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			metadata, readErr := stateread.Lstat(filepath.Join(parent, "other"))
			got := skillPlanDrift(home, nil, []manifest.Decl{{Name: "other"}})
			if len(got) != 1 {
				t.Fatalf("drift rows: %+v", got)
			}
			if broken {
				if metadata.Kind != stateread.KindUnreadable || readErr == nil || !strings.Contains(readErr.Error(), stateread.DiagUnreadable) {
					t.Fatalf("probe did not reach unreadable parent: %+v %v", metadata, readErr)
				}
				if got[0].State != "unknown" || got[0].Observed == nil || *got[0].Observed != "unreadable" {
					t.Fatalf("unreadable parent reported as proven absence: %+v", got)
				}
				if after, err := os.Readlink(parent); err != nil || after != target {
					t.Fatalf("parent repaired: %q %v", after, err)
				}
			} else if metadata.Kind != stateread.KindAbsent || readErr != nil || got[0].State != "drift" || got[0].Observed != nil {
				t.Fatalf("valid absence control: %+v %+v %v", got, metadata, readErr)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("missing target changed: %v", err)
			}
		})
	}
}
