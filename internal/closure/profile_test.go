package closure

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
)

type profileFixture struct {
	old   *contextlock.Lock
	roots []ProfileRequirement
	nodes map[string]*Node
	calls []string
}

func newProfileFixture(t *testing.T) *profileFixture {
	t.Helper()
	commit := strings.Repeat("1", 40)
	f := &profileFixture{old: &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: "profile", Members: []contextlock.Member{{Kind: "context", Name: "profile", Source: "example.com/profile", Commit: commit, Version: "1.0.0", RequiredBy: []string{}}, {Kind: "skill", Name: "unrelated", Source: "example.com/unrelated", Commit: strings.Repeat("2", 40), RequiredBy: []string{}}}}, roots: []ProfileRequirement{{Consumer: "profile", Name: "memory", Git: "https://example.com/memory", Ref: manifest.Ref{Kind: "revision", Value: commit}, Mode: "runtime", Commands: []string{"memory"}}}, nodes: map[string]*Node{}}
	for _, name := range []string{"memory", "helper"} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
		f.nodes[name] = &Node{Name: name, Identity: "example.com/" + name, Directory: ".", Snapshot: root, Resolved: gitops.ResolvedRef{Kind: "revision", Ref: commit, Commit: commit}, Spec: &skillspec.Spec{Commands: map[string]skillspec.Command{"memory": {Name: "memory", Type: "build"}, "doctor": {Name: "doctor", Type: "script"}, "git": {Name: "git", Type: "system"}}, Requirements: map[string]skillspec.Requirement{}}}
	}
	f.nodes["memory"].Spec.Requirements["helper"] = skillspec.Requirement{Name: "helper", Git: "https://example.com/helper", RefKind: "revision", RefValue: commit, Mode: "context"}
	return f
}
func (f *profileFixture) acquire(name, source, directory string, ref manifest.Ref) (*Node, error) {
	f.calls = append(f.calls, name)
	original := f.nodes[name]
	if original == nil {
		return nil, fmt.Errorf("unexpected acquisition: %s", name)
	}
	node := *original
	node.Directory = directory
	node.Identity = identity.Canonical(source)
	node.Resolved = gitops.ResolvedRef{Kind: ref.Kind, Ref: ref.Value, Commit: ref.Value}
	if ref.Kind == "tag" {
		node.Resolved.Commit = original.Resolved.Commit
	}
	return &node, nil
}
func (f *profileFixture) build(direct []manifest.Decl, raw map[string]any) (*ProfilePlan, error) {
	return BuildProfile(f.old, f.roots, direct, raw, f.acquire)
}

// PSP02–04, PSR01: provider-first transitive traversal retains independent
// modes, narrowing and unrelated old pins; unrelated direct roots are not read.
// A missing/system command, empty/unsorted selection, invalid mode and cycle
// each refuse instead of producing a partial plan.
func TestProfileClosureActivationAndRefusals(t *testing.T) {
	for _, mode := range []string{"valid", "missing-command", "system-command", "empty-selection", "unsorted-selection", "context-selection", "bad-mode", "cycle", "wrong-revision"} {
		t.Run(mode, func(t *testing.T) {
			f := newProfileFixture(t)
			switch mode {
			case "missing-command":
				f.roots[0].Commands = []string{"absent"}
			case "system-command":
				f.roots[0].Commands = []string{"git"}
			case "empty-selection":
				f.roots[0].Commands = []string{}
			case "unsorted-selection":
				f.roots[0].Commands = []string{"memory", "doctor"}
			case "context-selection":
				f.roots[0].Mode = "context"
			case "bad-mode":
				f.roots[0].Mode = "arbitrary"
			case "cycle":
				f.nodes["helper"].Spec.Requirements["memory"] = skillspec.Requirement{Name: "memory", Git: "https://example.com/memory", RefKind: "revision", RefValue: strings.Repeat("1", 40), Mode: "full"}
			case "wrong-revision":
				f.roots[0].Ref.Value = "short"
			}
			before := *f.old
			before.Members = append([]contextlock.Member(nil), f.old.Members...)
			result, err := f.build([]manifest.Decl{{Name: "unrelated", Git: "https://invalid.example/no-access", Ref: manifest.Ref{Kind: "revision", Value: strings.Repeat("3", 40)}}}, nil)
			if mode != "valid" {
				want := "profile_skill_plan_invalid"
				if mode == "missing-command" || mode == "system-command" {
					want = "profile_skill_command_invalid"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("intended %s refusal: %v", want, err)
				}
				if !reflect.DeepEqual(before, *f.old) {
					t.Fatal("refusal mutated old lock")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Nodes) != 2 || result.Nodes[0].Name != "helper" || result.Nodes[1].Name != "memory" {
				t.Fatalf("order: %+v", result.Nodes)
			}
			helper, memory := result.Nodes[0], result.Nodes[1]
			if !helper.ContextActive() || len(helper.ActiveCommands()) != 0 || memory.ContextActive() || !reflect.DeepEqual(memory.ActiveCommandNames(), []string{"memory"}) {
				t.Fatal("activation widened or inherited")
			}
			oldMember, _ := f.old.Find("skill", "unrelated")
			retained, _ := result.Lock.Find("skill", "unrelated")
			oldMember.Directory = "."
			if !reflect.DeepEqual(oldMember, retained) {
				t.Fatal("unrelated retained member rewritten")
			}
			for _, name := range f.calls {
				if name == "unrelated" {
					t.Fatal("unrelated root acquired")
				}
			}
			if len(result.Lock.SkillPlan.DirectConstraints) != 0 {
				t.Fatal("unrelated direct provenance leaked into plan")
			}
			if !reflect.DeepEqual(before, *f.old) {
				t.Fatal("planning mutated old lock")
			}
		})
	}
}

// PSR03/PSP04–06: a compatible direct declaration adds full activation and
// exact original-object provenance. Source/commit/directory conflicts refuse;
// a changed provenance cannot pass pinned reuse by recomputing its digest.
func TestProfileClosureOverlapAndReuse(t *testing.T) {
	for _, mode := range []string{"compatible", "alias", "source", "commit", "directory"} {
		t.Run(mode, func(t *testing.T) {
			f := newProfileFixture(t)
			direct := manifest.Decl{Name: "memory", Git: "https://example.com/memory", Ref: f.roots[0].Ref}
			switch mode {
			case "alias":
				direct.Ref = manifest.Ref{Kind: "tag", Value: "v1"}
			case "source":
				direct.Git = "https://example.com/other"
			case "commit":
				direct.Ref.Value = strings.Repeat("2", 40)
			case "directory":
				f.roots[0].Directory = "skills/memory"
			}
			raw := map[string]any{"memory": map[string]any{"name": "memory", "git": direct.Git, "revision": direct.Ref.Value}}
			result, err := f.build([]manifest.Decl{direct}, raw)
			if mode != "compatible" && mode != "alias" {
				if err == nil || !strings.Contains(err.Error(), "profile_skill_pin_conflict") {
					t.Fatalf("overlap refusal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			node := result.Nodes[1]
			if !node.ContextActive() || !reflect.DeepEqual(node.ActiveCommandNames(), []string{"doctor", "memory"}) {
				t.Fatal("full edge failed to dominate narrowing")
			}
			if err := VerifyProfilePlan(result.Lock, result.Lock); err != nil {
				t.Fatal(err)
			}
			changed := *result.Lock
			plan := *changed.SkillPlan
			plan.DirectConstraints = append([]contextlock.DirectConstraint(nil), plan.DirectConstraints...)
			plan.DirectConstraints[0].DeclarationSHA256 = "sha256:" + strings.Repeat("0", 64)
			changed.SkillPlan = &plan
			changed.SkillPlanSHA256, _ = plan.Hash()
			if err := VerifyProfilePlan(result.Lock, &changed); err == nil {
				t.Fatal("changed declaration admitted on reuse")
			}
			legacy := *f.old
			if err := VerifyProfilePlan(&legacy, result.Lock); err == nil || !strings.Contains(err.Error(), "profile_skill_plan_required") {
				t.Fatalf("legacy runtime: %v", err)
			}
		})
	}
}

// PSP04: distinct incoming runtime selections union; a context edge adds
// context without widening that selection; full dominates both. Reusing an
// identical consumer/provider edge or inventing its context refuses.
func TestProfileClosureIncomingUnion(t *testing.T) {
	for _, mode := range []string{"runtime", "context", "full", "duplicate", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			f := newProfileFixture(t)
			member := f.old.Members[0]
			member.Name = "second"
			f.old.Members = append(f.old.Members, member)
			f.old.Sort()
			second := f.roots[0]
			second.Consumer = "second"
			second.Mode = mode
			second.Commands = nil
			if mode == "runtime" {
				second.Commands = []string{"doctor"}
			}
			if mode == "duplicate" {
				second = f.roots[0]
			}
			if mode == "unknown" {
				second.Consumer = "unknown"
				second.Mode = "context"
			}
			f.roots = append(f.roots, second)
			got, err := f.build(nil, nil)
			if mode == "duplicate" || mode == "unknown" {
				if err == nil || !strings.Contains(err.Error(), "profile_skill_plan_invalid") {
					t.Fatalf("intended graph refusal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			node := got.Nodes[1]
			wantCommands := []string{"doctor", "memory"}
			if mode == "context" {
				wantCommands = []string{"memory"}
			}
			if node.ContextActive() != (mode != "runtime") || !reflect.DeepEqual(node.ActiveCommandNames(), wantCommands) {
				t.Fatalf("wrong union: context=%v commands=%v", node.ContextActive(), node.ActiveCommandNames())
			}
		})
	}
}
