package closure

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/manifest"
)

// harness builds skill repositories under a skills root and a machine home.
type harness struct {
	t          *testing.T
	skillsRoot string
	home       string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	return &harness{t: t, skillsRoot: t.TempDir(), home: t.TempDir()}
}

func (h *harness) git(dir string, args ...string) string {
	h.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		h.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// skill creates a repository named name under the skills root with the given
// csk-skill.json requirements (name -> mode) and tags it v1.
func (h *harness) skill(name string, commands []string, requirements map[string]map[string]any) string {
	h.t.Helper()
	dir := filepath.Join(h.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		h.t.Fatal(err)
	}
	h.git(dir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0o644); err != nil {
		h.t.Fatal(err)
	}
	spec := map[string]any{"schema_version": 4, "capabilities": map[string]any{}}
	commandsObj := map[string]any{}
	if len(commands) > 0 {
		scripts := filepath.Join(dir, "scripts")
		if err := os.MkdirAll(scripts, 0o755); err != nil {
			h.t.Fatal(err)
		}
		for _, command := range commands {
			if err := os.WriteFile(filepath.Join(scripts, command), []byte("#!/bin/sh\n"), 0o755); err != nil {
				h.t.Fatal(err)
			}
			commandsObj[command] = map[string]any{"type": "script", "unix_path": "scripts/" + command}
		}
		spec["runtime_roots"] = []string{"scripts"}
	}
	spec["commands"] = commandsObj
	if requirements != nil {
		spec["dependencies"] = map[string]any{"skills": requirements}
	}
	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "csk-skill.json"), payload, 0o644); err != nil {
		h.t.Fatal(err)
	}
	h.git(dir, "add", ".")
	h.git(dir, "commit", "-qm", "init")
	h.git(dir, "tag", "v1")
	return dir
}

func requirement(name, mode string, commands ...string) map[string]any {
	entry := map[string]any{
		"git": "./" + name, // local source; identity-free
		"ref": map[string]any{"kind": "tag", "value": "v1"},
	}
	if mode != "" {
		entry["mode"] = mode
	}
	if len(commands) > 0 {
		entry["commands"] = commands
	}
	return entry
}

func (h *harness) build(decls []manifest.Decl, subs map[string]devsub.Substitution) ([]*Node, error) {
	h.t.Helper()
	return Build(
		Options{SkillsRoot: h.skillsRoot, Home: h.home},
		&manifest.Manifest{Skills: decls},
		subs,
	)
}

func decl(name string) manifest.Decl {
	return manifest.Decl{Name: name, Source: name, Ref: manifest.Ref{Kind: "tag", Value: "v1"}}
}

func names(nodes []*Node) []string {
	var out []string
	for _, node := range nodes {
		out = append(out, node.Name)
	}
	return out
}

func TestDiamondResolvesOnceProvidersFirst(t *testing.T) {
	h := newHarness(t)
	h.skill("leaf", []string{"leaf-tool"}, nil)
	h.skill("left", nil, map[string]map[string]any{"leaf": requirement("leaf", "full")})
	h.skill("right", nil, map[string]map[string]any{"leaf": requirement("leaf", "full")})

	nodes, err := h.build([]manifest.Decl{decl("left"), decl("right")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	order := names(nodes)
	if len(order) != 3 {
		t.Fatalf("closure: %v", order)
	}
	if order[0] != "leaf" {
		t.Fatalf("provider must precede consumers: %v", order)
	}
	// deterministic order
	again, err := h.build([]manifest.Decl{decl("left"), decl("right")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(names(again)) != fmt.Sprint(order) {
		t.Fatalf("order not deterministic: %v vs %v", names(again), order)
	}
}

func TestVersionConflictNamesBothChains(t *testing.T) {
	h := newHarness(t)
	leafDir := h.skill("leaf", nil, nil)
	// second tag on a new commit
	if err := os.WriteFile(filepath.Join(leafDir, "SKILL.md"), []byte("# leaf v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.git(leafDir, "commit", "-qam", "two")
	h.git(leafDir, "tag", "v2")

	left := map[string]map[string]any{"leaf": requirement("leaf", "full")}
	rightReq := requirement("leaf", "full")
	rightReq["ref"] = map[string]any{"kind": "tag", "value": "v2"}
	right := map[string]map[string]any{"leaf": rightReq}
	h.skill("left", nil, left)
	h.skill("right", nil, right)

	_, err := h.build([]manifest.Decl{decl("left"), decl("right")}, nil)
	if err == nil {
		t.Fatal("expected version conflict")
	}
	message := err.Error()
	if !strings.Contains(message, "version conflict for leaf") ||
		!strings.Contains(message, "left -> leaf") || !strings.Contains(message, "right -> leaf") {
		t.Fatalf("conflict message lacks chains: %s", message)
	}
}

func TestSameCommitDifferentRefsUnify(t *testing.T) {
	h := newHarness(t)
	leafDir := h.skill("leaf", nil, nil)
	h.git(leafDir, "tag", "v1-alias") // same commit, second tag

	aliasReq := requirement("leaf", "full")
	aliasReq["ref"] = map[string]any{"kind": "tag", "value": "v1-alias"}
	h.skill("left", nil, map[string]map[string]any{"leaf": requirement("leaf", "full")})
	h.skill("right", nil, map[string]map[string]any{"leaf": aliasReq})

	nodes, err := h.build([]manifest.Decl{decl("left"), decl("right")}, nil)
	if err != nil {
		t.Fatalf("same-commit refs must unify: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("closure: %v", names(nodes))
	}
}

func TestCycleFails(t *testing.T) {
	h := newHarness(t)
	// a requires b; b requires a. Repos must exist before specs reference
	// them, so create b first without requirements, then rewrite.
	h.skill("b", nil, map[string]map[string]any{})
	aDir := h.skill("a", nil, map[string]map[string]any{"b": requirement("b", "full")})
	_ = aDir
	bDir := filepath.Join(h.skillsRoot, "b")
	spec := map[string]any{
		"schema_version": 4, "capabilities": map[string]any{}, "commands": map[string]any{},
		"dependencies": map[string]any{"skills": map[string]any{"a": requirement("a", "full")}},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	if err := os.WriteFile(filepath.Join(bDir, "csk-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	h.git(bDir, "commit", "-qam", "cycle")
	h.git(bDir, "tag", "-f", "v1")

	_, err := h.build([]manifest.Decl{decl("a")}, nil)
	if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("err = %v, want cycle", err)
	}
}

func TestActivationModesAndNarrowing(t *testing.T) {
	h := newHarness(t)
	h.skill("provider", []string{"alpha", "beta"}, nil)
	h.skill("consumer", nil, map[string]map[string]any{
		"provider": requirement("provider", "runtime", "alpha"),
	})

	nodes, err := h.build([]manifest.Decl{decl("consumer")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var provider *Node
	for _, node := range nodes {
		if node.Name == "provider" {
			provider = node
		}
	}
	if provider.ContextActive() {
		t.Fatal("runtime edge must not activate context")
	}
	active := provider.ActiveCommands()
	if !active["alpha"] || active["beta"] {
		t.Fatalf("narrowing failed: %v", active)
	}

	// direct project declaration is a full edge: everything activates
	nodes, err = h.build([]manifest.Decl{decl("consumer"), decl("provider")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if node.Name == "provider" {
			if !node.ContextActive() || !node.ActiveCommands()["beta"] {
				t.Fatalf("full edge must activate all: context=%v commands=%v", node.ContextActive(), node.ActiveCommands())
			}
		}
	}
}

func TestNarrowingUnknownCommandFails(t *testing.T) {
	h := newHarness(t)
	h.skill("provider", []string{"alpha"}, nil)
	h.skill("consumer", nil, map[string]map[string]any{
		"provider": requirement("provider", "runtime", "missing"),
	})
	_, err := h.build([]manifest.Decl{decl("consumer")}, nil)
	if err == nil || !strings.Contains(err.Error(), "does not export a script command") {
		t.Fatalf("err = %v", err)
	}
}

func TestCommandCollisions(t *testing.T) {
	h := newHarness(t)
	h.skill("one", []string{"tool"}, nil)
	h.skill("two", []string{"tool"}, nil)
	nodes, err := h.build([]manifest.Decl{decl("one"), decl("two")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := DetectActiveCommandCollisions(nodes); err == nil || !strings.Contains(err.Error(), `command collision for "tool"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestAllowlistGatesNetworkClone(t *testing.T) {
	h := newHarness(t)
	// declared skill absent from skills root with a remote-looking URL
	declWithGit := manifest.Decl{
		Name: "absent", Source: "absent",
		Ref: manifest.Ref{Kind: "tag", Value: "v1"},
		Git: "git@forbidden.example.com:skills/absent.git",
	}
	_, err := Build(
		Options{SkillsRoot: h.skillsRoot, Home: h.home, AllowedSources: []string{"git.example.com/skills"}},
		&manifest.Manifest{Skills: []manifest.Decl{declWithGit}},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "source not allowed for absent") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "forbidden.example.com/skills/absent") {
		t.Fatalf("identity missing from error: %v", err)
	}
}

func TestDevSubstitutionReplacesAndSkipsUnification(t *testing.T) {
	h := newHarness(t)
	h.skill("leaf", nil, nil)
	h.skill("consumer", nil, map[string]map[string]any{"leaf": requirement("leaf", "full")})

	// local checkout with an extra commit
	sub := t.TempDir()
	h.git(sub, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(sub, "SKILL.md"), []byte("# local leaf"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.git(sub, "add", ".")
	h.git(sub, "commit", "-qm", "local")

	nodes, err := h.build(
		[]manifest.Decl{decl("consumer")},
		map[string]devsub.Substitution{"leaf": {SkillName: "leaf", Path: sub}},
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if node.Name == "leaf" {
			if node.Substituted == "" || !strings.Contains(node.Substituted, "path ") {
				t.Fatalf("substitution not recorded: %+v", node)
			}
			if node.Repo != sub {
				t.Fatalf("repo = %s, want the substituted checkout", node.Repo)
			}
		}
	}
}

func TestMissingRepoWithoutGitFails(t *testing.T) {
	h := newHarness(t)
	_, err := h.build([]manifest.Decl{decl("ghost")}, nil)
	if err == nil || !strings.Contains(err.Error(), "skill repository not found for ghost") {
		t.Fatalf("err = %v", err)
	}
}

func TestFetchExistingIsScopedToReachedClosureAndDeduplicated(t *testing.T) {
	h := newHarness(t)
	h.skill("provider", []string{"provider-tool"}, nil)
	h.skill("consumer", nil, map[string]map[string]any{
		"provider": requirement("provider", "runtime", "provider-tool"),
	})
	h.skill("unrelated", nil, nil)

	fetched := map[string]bool{}
	var calls []string
	opts := Options{
		SkillsRoot:    h.skillsRoot,
		Home:          h.home,
		FetchExisting: true,
		FetchedRepos:  fetched,
		FetchRepo: func(repo string) error {
			calls = append(calls, filepath.Base(repo))
			return nil
		},
	}
	manifestValue := &manifest.Manifest{Skills: []manifest.Decl{decl("consumer")}}
	if _, err := Build(opts, manifestValue, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(calls, ","); got != "consumer,provider" {
		t.Fatalf("fetched repositories = %s, want consumer,provider", got)
	}
	if _, err := Build(opts, manifestValue, nil); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("shared closure repositories fetched again: %v", calls)
	}
}

func TestScratchResolutionLeavesPersistentRootsAbsent(t *testing.T) {
	remoteHarness := newHarness(t)
	remote := remoteHarness.skill("remote", nil, nil)
	root := t.TempDir()
	skillsRoot := filepath.Join(root, "missing-skills")
	home := filepath.Join(root, "missing-home")
	scratch := t.TempDir()

	nodes, err := Build(Options{
		SkillsRoot:  skillsRoot,
		Home:        home,
		ScratchRoot: scratch,
	}, &manifest.Manifest{Skills: []manifest.Decl{{
		Name: "remote", Source: "remote", Git: remote,
		Ref: manifest.Ref{Kind: "tag", Value: "v1"},
	}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || !strings.HasPrefix(nodes[0].Repo, scratch) || !strings.HasPrefix(nodes[0].Snapshot, scratch) {
		t.Fatalf("scratch resolution escaped workspace: %+v", nodes)
	}
	if _, err := os.Stat(skillsRoot); !os.IsNotExist(err) {
		t.Fatalf("dry-run created skills root: %v", err)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("dry-run created manager home: %v", err)
	}
}
