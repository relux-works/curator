package closure

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

// writeDraftSkill creates one local package with SKILL.md frontmatter and
// a minimal agent-skill.json. commands non-empty exports a script command
// with runtime_roots scripts; runtimeExtra and buildExtra add files under
// those roots so refresh tests can mutate them without touching context.
func writeDraftSkill(t *testing.T, dir string, name string, withCommand bool, runtimeExtra, buildExtra map[string]string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := "---\nname: " + name + "\ndescription: Test " + name + "\n---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(frontmatter), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := map[string]any{"schema_version": 4, "capabilities": map[string]any{}, "commands": map[string]any{}, "dependencies": map[string]any{"skills": map[string]any{}}}
	if withCommand {
		if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "scripts", "tool.sh"), []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		spec["commands"] = map[string]any{"tool": map[string]any{"type": "script", "unix_path": "scripts/tool.sh"}}
		spec["runtime_roots"] = []string{"scripts"}
	}
	if len(runtimeExtra) > 0 {
		if _, ok := spec["runtime_roots"]; !ok {
			spec["runtime_roots"] = []string{"runtime"}
		} else {
			roots := spec["runtime_roots"].([]string)
			found := false
			for _, r := range roots {
				if r == "runtime" {
					found = true
				}
			}
			if !found {
				spec["runtime_roots"] = append(roots, "runtime")
			}
		}
		for rel, content := range runtimeExtra {
			full := filepath.Join(dir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(buildExtra) > 0 {
		spec["schema_version"] = 6
		spec["build_roots"] = []string{"build"}
		commands := spec["commands"].(map[string]any)
		commands["builder"] = map[string]any{"type": "build", "driver": "go-v1", "source_dir": "build"}
		if err := os.MkdirAll(filepath.Join(dir, "build"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "build", "go.mod"), []byte("module example.com/"+name+"\n\ngo 1.21\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for rel, content := range buildExtra {
			full := filepath.Join(dir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func parseDraftManifest(t *testing.T, projectRoot, payload string) (*manifest.Manifest, []byte) {
	t.Helper()
	path := filepath.Join(projectRoot, manifest.Name)
	m, err := manifest.ParseBytesWithOptions([]byte(payload), path, manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	return m, []byte(payload)
}

// gitFrozenOptsForTest maps every named root member to one repository and
// its accepted manifest declaration, mirroring what production frozen
// consumption derives from the manifest and machine bindings.
func gitFrozenOptsForTest(repo string, members ...string) FrozenOptions {
	opts := FrozenOptions{GitRepos: map[string]string{}, Refs: map[string]FrozenRef{}}
	for _, name := range members {
		opts.GitRepos[name] = repo
		opts.Refs[name] = FrozenRef{Kind: "tag", Ref: "v1", Git: "https://example.org/kit.git", Source: name}
	}
	return opts
}

func TestResolveDraftLocalDeterministic(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "zulu"), "zulu", false, nil, nil)
	writeDraftSkill(t, filepath.Join(project, "skills", "alpha"), "alpha", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	first, err := ResolveDraft(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ResolveDraft(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if first.Lock.LockSHA256 != second.Lock.LockSHA256 {
		t.Fatalf("non-deterministic lock: %s vs %s", first.Lock.LockSHA256, second.Lock.LockSHA256)
	}
	if len(first.Lock.Members) != 2 || first.Lock.Members[0].Name != "alpha" || first.Lock.Members[1].Name != "zulu" {
		t.Fatalf("members not sorted: %+v", first.Lock.Members)
	}
	for _, member := range first.Lock.Members {
		if member.Selection == nil || *member.Selection != 0 {
			t.Fatalf("%s selection = %+v, want 0 (shared collection index)", member.Name, member.Selection)
		}
		if member.Package.Kind != sourcelock.KindLocalSnapshot {
			t.Fatalf("%s kind = %s", member.Name, member.Package.Kind)
		}
		if !strings.HasPrefix(member.Package.Snapshot, "sha256:") || !strings.HasPrefix(member.ContentSHA256, "sha256:") {
			t.Fatalf("%s digests malformed", member.Name)
		}
		tree, ok := first.Frozen[member.Name]
		if !ok || strings.HasPrefix(tree, project) && !strings.Contains(tree, "local-snapshots") {
			// Frozen trees must be store paths, never the live project dir.
			if ok && tree == filepath.Join(project, "skills", member.Name) {
				t.Fatalf("%s frozen points at live bytes: %s", member.Name, tree)
			}
		}
		if _, err := os.Stat(tree); err != nil {
			t.Fatalf("%s frozen missing: %v", member.Name, err)
		}
	}
	if err := first.Lock.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := first.Bindings.CheckFresh(first.Lock); err != nil {
		t.Fatal(err)
	}
}

func TestOpenDraftFrozenPinsMembership(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
	if err != nil {
		t.Fatal(err)
	}
	// Live collection gains a member; frozen launch must still use only
	// the locked review package (frozen-membership).
	writeDraftSkill(t, filepath.Join(project, "skills", "docs"), "docs", false, nil, nil)
	frozen, err := OpenDraftFrozen(home, plan.Lock, FrozenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(frozen) != 1 || frozen["review"] == "" {
		t.Fatalf("frozen = %+v, want only review", frozen)
	}
	if _, ok := frozen["docs"]; ok {
		t.Fatalf("frozen adopted live docs member")
	}
	nodes, _, err := LoadDraftFrozenNodes(home, plan.Lock, FrozenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Name != "review" {
		t.Fatalf("frozen nodes = %+v", nodes)
	}
}

func TestOpenDraftFrozenMissingSnapshotFails(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	// Note: include names a folder under directory "skills".
	m, raw := parseDraftManifest(t, project, `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`)
	_ = payload
	plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
	if err != nil {
		t.Fatal(err)
	}
	// Remove the immutable store tree: frozen consumption must fail
	// unavailable and never recreate from live bytes.
	for _, tree := range plan.Frozen {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(tree))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := OpenDraftFrozen(home, plan.Lock, FrozenOptions{}); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("err = %v, want source_snapshot_unavailable", err)
	}
	if _, _, err := LoadDraftFrozenNodes(home, plan.Lock, FrozenOptions{}); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("nodes err = %v, want source_snapshot_unavailable", err)
	}
}

func TestRefreshCatchesRuntimeOnlyAndBuildOnly(t *testing.T) {
	for _, mode := range []string{"runtime", "build"} {
		t.Run(mode, func(t *testing.T) {
			project := t.TempDir()
			home := t.TempDir()
			var runtimeExtra, buildExtra map[string]string
			if mode == "runtime" {
				runtimeExtra = map[string]string{"runtime/helper.sh": "v1"}
			} else {
				buildExtra = map[string]string{"build/main.go": "package main"}
			}
			writeDraftSkill(t, filepath.Join(project, "pkgs", "review"), "review", true, runtimeExtra, buildExtra)
			rawPayload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"pkgs/review"}]}`
			m, raw := parseDraftManifest(t, project, rawPayload)
			before, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
			if err != nil {
				t.Fatal(err)
			}
			beforeMember, _ := before.Lock.Find("review")
			// Mutate only the runtime/build input; SKILL.md and projected
			// context stay byte-identical.
			if mode == "runtime" {
				if err := os.WriteFile(filepath.Join(project, "pkgs", "review", "runtime", "helper.sh"), []byte("v2"), 0o644); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(filepath.Join(project, "pkgs", "review", "build", "main.go"), []byte("package main // v2"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			after, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
			if err != nil {
				t.Fatal(err)
			}
			afterMember, _ := after.Lock.Find("review")
			if beforeMember.Package.Snapshot == afterMember.Package.Snapshot {
				t.Fatalf("%s-only edit preserved package identity; refresh would miss it", mode)
			}
			if beforeMember.ContentSHA256 != afterMember.ContentSHA256 {
				t.Fatalf("%s-only edit changed context hash %s -> %s; want identical", mode, beforeMember.ContentSHA256, afterMember.ContentSHA256)
			}
			if before.Lock.LockSHA256 == after.Lock.LockSHA256 {
				t.Fatalf("lock unchanged after %s-only edit", mode)
			}
		})
	}
}

func TestResolveDraftRefusals(t *testing.T) {
	for _, mode := range []string{"unknown-alias", "duplicate-name", "case-conflict", "stale-manifest", "branch-transitive"} {
		t.Run(mode, func(t *testing.T) {
			project := t.TempDir()
			home := t.TempDir()
			switch mode {
			case "unknown-alias":
				m, raw := parseDraftManifest(t, project, `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[]}`)
				m.Skills = append(m.Skills, manifest.Decl{Name: "x", Selector: &manifest.Selector{From: "absent", Directory: "."}})
				if _, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}); err == nil || !strings.Contains(err.Error(), "source_alias_unknown") {
					t.Fatalf("err = %v, want source_alias_unknown", err)
				}
			case "duplicate-name":
				writeDraftSkill(t, filepath.Join(project, "a", "review"), "review", false, nil, nil)
				writeDraftSkill(t, filepath.Join(project, "b", "review"), "review", false, nil, nil)
				payload := `{"schema_version":2,"sources":{"a":{"path":"./a"},"b":{"path":"./b"}},"skills":[{"name":"review","from":"a","directory":"review"},{"name":"review","from":"b","directory":"review"}]}`
				// The parser already retains the repeated-selection refusal.
				if _, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, manifest.Name), manifest.ParseOptions{DraftSourcesV1: true}); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
					t.Fatalf("err = %v, want source_name_conflict", err)
				}
			case "case-conflict":
				writeDraftSkill(t, filepath.Join(project, "a", "review"), "review", false, nil, nil)
				// Second package differs only by SKILL.md name case.
				writeDraftSkill(t, filepath.Join(project, "b", "other"), "REVIEW", false, nil, nil)
				payload := `{"schema_version":2,"sources":{"a":{"path":"./a"},"b":{"path":"./b"}},"skills":[{"name":"review","from":"a","directory":"review"},{"name":"REVIEW","from":"b","directory":"other"}]}`
				// Parser accepts both spellings; resolution must refuse
				// filesystem-equivalent destinations.
				m, raw := parseDraftManifest(t, project, payload)
				if _, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
					t.Fatalf("err = %v, want source_name_conflict", err)
				}
			case "stale-manifest":
				writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
				payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
				m, raw := parseDraftManifest(t, project, payload)
				plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
				if err != nil {
					t.Fatal(err)
				}
				changed := `{"schema_version":2,"sources":{"s":{"path":"./skills"}},"skills":[{"from":"s","directory":".","include":["review"]}]}`
				if err := plan.Lock.CheckStale([]byte(changed)); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
					t.Fatalf("err = %v, want source_lock_stale", err)
				}
			case "branch-transitive":
				// The defensive transitive branch gate fires even if a
				// branch requirement reaches closure outside the parser.
				h := newHarness(t)
				h.skill("provider", nil, nil)
				m := selectionProject(t, t.TempDir(), []string{"alpha"}, map[string]map[string]any{"provider": requirement("provider", "context")})
				nodes, err := BuildExpanded(Options{SkillsRoot: h.skillsRoot, Home: h.home}, m, manifest.ExpansionOptions{}, fixtureAcquisition, nil)
				if err != nil {
					t.Fatal(err)
				}
				for _, node := range nodes {
					if node.Name == "alpha" {
						req := node.Spec.Requirements["provider"]
						req.RefKind = "branch"
						req.RefValue = "main"
						node.Spec.Requirements["provider"] = req
					}
				}
				if err := retainBranchRule(nodes); err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
					t.Fatalf("err = %v, want source_selection_invalid", err)
				}
			}
		})
	}
}

func TestRefreshDraftWritesAtomically(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := filepath.Join(home, "bindings.json")
	plan, err := RefreshDraft(cfg, lockPath, bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if stored.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("stored %s != planned %s", stored.LockSHA256, plan.Lock.LockSHA256)
	}
	// A failing refresh preserves the prior files.
	bad := cfg
	bad.Manifest = nil
	if _, err := RefreshDraft(bad, lockPath, bindingsPath); err == nil {
		t.Fatalf("bad refresh succeeded")
	}
	restored, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if restored.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("failed refresh clobbered the lock")
	}
}

func TestRefreshDraftSecondWriteFailurePreservesLock(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := filepath.Join(home, "bindings.json")
	plan, err := RefreshDraft(cfg, lockPath, bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	priorBindings, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	// Change the live bytes so a successful refresh would publish a new
	// generation: without a mutation this test cannot tell rollback apart
	// from a no-op write.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Inject a failure at the second publication step by replacing the
	// bindings file with a directory: the lock rename succeeds, the
	// bindings rename fails.
	if err := os.Remove(bindingsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(bindingsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := RefreshDraft(cfg, lockPath, bindingsPath); err == nil {
		t.Fatalf("refresh with unwritable bindings succeeded")
	}
	restored, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(priorLock) {
		t.Fatalf("failed refresh replaced the lock")
	}
	restoredLock, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if restoredLock.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("lock generation moved %s -> %s", plan.Lock.LockSHA256, restoredLock.LockSHA256)
	}
	if info, err := os.Lstat(bindingsPath); err != nil || !info.IsDir() {
		t.Fatalf("bindings placeholder changed: %v %v", info, err)
	}
	// The previously published snapshot stays usable: installed state is
	// not rolled forward or removed by the failed attempt.
	if _, ok := plan.Lock.Find("review"); !ok {
		t.Fatalf("prior lock lost review")
	}
	if _, err := OpenDraftFrozen(home, plan.Lock, FrozenOptions{}); err != nil {
		t.Fatalf("prior snapshot unusable after failed refresh: %v", err)
	}
	if len(priorBindings) == 0 {
		t.Fatalf("prior bindings unreadable")
	}
}

func TestRefreshDraftSecondWriteFailureWithoutPriorLock(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsDir := filepath.Join(home, "bindings-dir")
	if err := os.Mkdir(bindingsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// No prior lock exists and the bindings write cannot succeed: the
	// half-published lock must be removed again.
	if _, err := RefreshDraft(cfg, lockPath, bindingsDir); err == nil {
		t.Fatalf("refresh with unwritable bindings succeeded")
	}
	if _, err := os.Lstat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("failed first publish left a lock behind: %v", err)
	}
}

// draftGitRepo commits packages under a local Git repository and tags the
// result, so Git acquisition tests resolve and extract offline through the
// production gitops path with GitRoots pointed at the repository.
func draftGitRepo(t *testing.T, packages map[string]string, tag string) string {
	t.Helper()
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	for dir, name := range packages {
		writeDraftSkill(t, filepath.Join(repo, filepath.FromSlash(dir)), name, false, nil, nil)
	}
	run("add", ".")
	run("commit", "-qm", "fixture")
	run("tag", tag)
	return repo
}

func draftGitConfig(t *testing.T, repo, home, payload string) DraftResolveConfig {
	t.Helper()
	project := t.TempDir()
	m, raw := parseDraftManifest(t, project, payload)
	return DraftResolveConfig{
		ProjectRoot:     project,
		Home:            home,
		Manifest:        m,
		ManifestPayload: raw,
		Expansion:       manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	}
}

// draftGitRepoWithCommands commits one script package (with an exported
// command and runtime roots) under a local Git repository and tags it, so
// frozen-consumption tests can tamper runtime bytes the projected context
// hash excludes.
func draftGitRepoWithCommands(t *testing.T, dir, name, tag string) string {
	t.Helper()
	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	writeDraftSkill(t, filepath.Join(repo, filepath.FromSlash(dir)), name, true, nil, nil)
	run("add", ".")
	run("commit", "-qm", "fixture")
	run("tag", tag)
	return repo
}

func TestResolveDraftGitSubtreeFrozen(t *testing.T) {
	for _, mode := range []string{"individual", "collection"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			repo := draftGitRepo(t, map[string]string{"skills/review": "review", "skills/docs": "docs"}, "v1")
			resolved, err := gitops.Resolve(repo, "tag", "v1")
			if err != nil {
				t.Fatal(err)
			}
			var payload string
			if mode == "individual" {
				payload = `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
			} else {
				payload = `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
			}
			cfg := draftGitConfig(t, repo, home, payload)
			plan, err := ResolveDraft(cfg)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{}
			if mode == "individual" {
				want["review"] = "skills/review"
			} else {
				want["review"] = "skills/review"
				want["docs"] = "skills/docs"
			}
			if len(plan.Lock.Members) != len(want) {
				t.Fatalf("members = %+v, want %d", plan.Lock.Members, len(want))
			}
			repoRoot := filepath.Join(home, "cache", "example.org", "kit", resolved.Commit, "snapshot")
			for name, directory := range want {
				member, ok := plan.Lock.Find(name)
				if !ok {
					t.Fatalf("lock misses %s", name)
				}
				if member.Package.Kind != sourcelock.KindNetworkGit || member.Package.Directory != directory || member.Directory != directory {
					t.Fatalf("%s package = %+v directory = %q", name, member.Package, member.Directory)
				}
				tree, ok := plan.Frozen[name]
				if !ok {
					t.Fatalf("no frozen tree for %s", name)
				}
				if tree != filepath.Join(repoRoot, filepath.FromSlash(directory)) {
					t.Fatalf("%s frozen = %s, want subtree under %s", name, tree, repoRoot)
				}
				if tree == repoRoot {
					t.Fatalf("%s frozen serves the repository root instead of %q", name, directory)
				}
			}
			names := make([]string, 0, len(want))
			for name := range want {
				names = append(names, name)
			}
			frozenOpts := gitFrozenOptsForTest(repo, names...)
			frozen, err := OpenDraftFrozen(home, plan.Lock, frozenOpts)
			if err != nil {
				t.Fatal(err)
			}
			nodes, _, err := LoadDraftFrozenNodes(home, plan.Lock, frozenOpts)
			if err != nil {
				t.Fatal(err)
			}
			if len(nodes) != len(want) {
				t.Fatalf("frozen nodes = %d, want %d", len(nodes), len(want))
			}
			for _, node := range nodes {
				if _, ok := want[node.Name]; !ok {
					t.Fatalf("unexpected frozen node %s", node.Name)
				}
				if node.Snapshot != frozen[node.Name] {
					t.Fatalf("%s snapshot %s != frozen %s", node.Name, node.Snapshot, frozen[node.Name])
				}
			}
		})
	}
}

// gitAt runs one git command in dir with deterministic authorship.
func gitAt(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestResolveDraftExpandsOverResolvedCommitNotCheckout proves selection
// names, validation and collection membership come from the declared tag's
// commit tree, never from the working checkout: tag v1 stays valid while
// HEAD removes a package SKILL.md (individual) or a collection member. A
// checkout-based expansion would reject the first attempt and silently
// drop the member in the second.
func TestResolveDraftExpandsOverResolvedCommitNotCheckout(t *testing.T) {
	for _, mode := range []string{"individual", "collection"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			repo := draftGitRepo(t, map[string]string{"skills/review": "review", "skills/docs": "docs"}, "v1")
			resolved, err := gitops.Resolve(repo, "tag", "v1")
			if err != nil {
				t.Fatal(err)
			}
			// HEAD diverges from the tag after tagging: review loses its
			// SKILL.md and docs vanishes entirely.
			if err := os.Remove(filepath.Join(repo, "skills", "review", "SKILL.md")); err != nil {
				t.Fatal(err)
			}
			if err := os.RemoveAll(filepath.Join(repo, "skills", "docs")); err != nil {
				t.Fatal(err)
			}
			gitAt(t, repo, "add", "-A")
			gitAt(t, repo, "commit", "-qm", "head-diverges")
			var payload string
			if mode == "individual" {
				payload = `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
			} else {
				payload = `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`
			}
			plan, err := ResolveDraft(draftGitConfig(t, repo, home, payload))
			if err != nil {
				t.Fatalf("resolve over the tagged commit failed: %v", err)
			}
			want := map[string]string{"review": "skills/review"}
			if mode == "collection" {
				want["docs"] = "skills/docs"
			}
			if len(plan.Lock.Members) != len(want) {
				t.Fatalf("members = %+v, want %d", plan.Lock.Members, len(want))
			}
			for name, directory := range want {
				member, ok := plan.Lock.Find(name)
				if !ok {
					t.Fatalf("lock misses %s: %+v", name, plan.Lock.Members)
				}
				if member.Package.Commit.Hex != resolved.Commit {
					t.Fatalf("%s commit = %s, want the tagged %s", name, member.Package.Commit.Hex, resolved.Commit)
				}
				if member.Package.Directory != directory {
					t.Fatalf("%s directory = %q, want %q", name, member.Package.Directory, directory)
				}
			}
		})
	}
}

// TestResolveDraftResolvesEachGitAliasOnce proves one alias shared by two
// selections resolves exactly once: a counting resolver must observe a
// single call, since the pin phase runs before expansion, not per member.
func TestResolveDraftResolvesEachGitAliasOnce(t *testing.T) {
	home := t.TempDir()
	repo := draftGitRepo(t, map[string]string{"skills/review": "review", "skills/docs": "docs"}, "v1")
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"},{"name":"docs","from":"s","directory":"skills/docs"}]}`
	cfg := draftGitConfig(t, repo, home, payload)
	calls := 0
	cfg.GitResolve = func(repoRoot, kind, value string) (gitops.ResolvedRef, error) {
		calls++
		return gitops.Resolve(repoRoot, kind, value)
	}
	plan, err := ResolveDraft(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Lock.Members) != 2 {
		t.Fatalf("members = %+v, want 2", plan.Lock.Members)
	}
	if calls != 1 {
		t.Fatalf("GitResolve calls = %d, want exactly 1 for one shared alias", calls)
	}
}

func TestOpenDraftFrozenGitTamperRefused(t *testing.T) {
	setup := func(t *testing.T) (string, *sourcelock.Lock, string, FrozenOptions) {
		t.Helper()
		home := t.TempDir()
		repo := draftGitRepo(t, map[string]string{"skills/review": "review"}, "v1")
		payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
		plan, err := ResolveDraft(draftGitConfig(t, repo, home, payload))
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := gitops.Resolve(repo, "tag", "v1")
		if err != nil {
			t.Fatal(err)
		}
		repoRoot := filepath.Join(home, "cache", "example.org", "kit", resolved.Commit, "snapshot")
		return home, plan.Lock, repoRoot, gitFrozenOptsForTest(repo, "review")
	}
	t.Run("tampered-context-file", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		contextFile := filepath.Join(repoRoot, "skills", "review", "references", "info.md")
		if err := os.WriteFile(contextFile, []byte("tampered"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("symlink", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		link := filepath.Join(repoRoot, "skills", "review", "references", "evil.md")
		if err := os.Symlink(filepath.Join(repoRoot, "skills", "review", "SKILL.md"), link); err != nil {
			t.Fatal(err)
		}
		if _, err := OpenDraftFrozen(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("nodes err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("partial-cache", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		if err := os.Remove(filepath.Join(repoRoot, "skills", "review", "references", "info.md")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("missing-subtree", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		if err := os.RemoveAll(filepath.Join(repoRoot, "skills", "review")); err != nil {
			t.Fatal(err)
		}
		if _, err := OpenDraftFrozen(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
}

// TestOpenDraftFrozenGitRuntimeBytesAuthenticated proves frozen Git reads
// authenticate the complete locked snapshot, not just the projected
// context: runtime-only tampering, a removed runtime member, and a flipped
// executable bit all fail with source_snapshot_changed even though the
// locked content hash excludes those bytes.
func TestOpenDraftFrozenGitRuntimeBytesAuthenticated(t *testing.T) {
	setup := func(t *testing.T) (string, *sourcelock.Lock, string, FrozenOptions) {
		t.Helper()
		home := t.TempDir()
		repo := draftGitRepoWithCommands(t, "skills/review", "review", "v1")
		payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
		plan, err := ResolveDraft(draftGitConfig(t, repo, home, payload))
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := gitops.Resolve(repo, "tag", "v1")
		if err != nil {
			t.Fatal(err)
		}
		repoRoot := filepath.Join(home, "cache", "example.org", "kit", resolved.Commit, "snapshot")
		return home, plan.Lock, repoRoot, gitFrozenOptsForTest(repo, "review")
	}
	t.Run("tampered-runtime-script", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		script := filepath.Join(repoRoot, "skills", "review", "scripts", "tool.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho TAMPERED\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("missing-runtime-member", func(t *testing.T) {
		home, lock, repoRoot, frozenOpts := setup(t)
		if err := os.Remove(filepath.Join(repoRoot, "skills", "review", "scripts", "tool.sh")); err != nil {
			t.Fatal(err)
		}
		if _, err := OpenDraftFrozen(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("executable-bit-flip", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows does not expose portable executable permission bits")
		}
		home, lock, repoRoot, frozenOpts := setup(t)
		script := filepath.Join(repoRoot, "skills", "review", "scripts", "tool.sh")
		if err := os.Chmod(script, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, frozenOpts); err == nil || !strings.Contains(err.Error(), "source_snapshot_changed") {
			t.Fatalf("err = %v, want source_snapshot_changed", err)
		}
	})
	t.Run("missing-repository-fails-closed", func(t *testing.T) {
		home, lock, _, _ := setup(t)
		if _, err := OpenDraftFrozen(home, lock, FrozenOptions{}); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
			t.Fatalf("err = %v, want source_snapshot_unavailable", err)
		}
		if _, _, err := LoadDraftFrozenNodes(home, lock, FrozenOptions{}); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
			t.Fatalf("nodes err = %v, want source_snapshot_unavailable", err)
		}
	})
}

// TestDraftFrozenNodesCarryDeclaredIdentity proves frozen nodes carry the
// accepted declared identity into materialization: root Git members keep
// the manifest ref kind/value, the declared endpoint, and the installed
// source name, so the unchanged marker builder receives real declared
// values instead of empty fields.
func TestDraftFrozenNodesCarryDeclaredIdentity(t *testing.T) {
	home := t.TempDir()
	repo := draftGitRepoWithCommands(t, "skills/review", "review", "v1")
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	plan, err := ResolveDraft(draftGitConfig(t, repo, home, payload))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	nodes, _, err := LoadDraftFrozenNodes(home, plan.Lock, gitFrozenOptsForTest(repo, "review"))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(nodes))
	}
	node := nodes[0]
	if node.Resolved.Kind != "tag" || node.Resolved.Ref != "v1" || node.Resolved.Commit != resolved.Commit {
		t.Fatalf("resolved = %+v, want tag v1 %s", node.Resolved, resolved.Commit)
	}
	if node.Decl.Source != "review" || node.Decl.Git != "https://example.org/kit.git" {
		t.Fatalf("decl = %+v, want source review with the declared endpoint", node.Decl)
	}
}

// TestDraftTransitiveGitDependencyFrozen proves a draft closure with a
// transitive skill requirement resolves to a deterministic lock and
// consumes it from pinned bytes: the transitive member authenticates
// against the SkillsRoot checkout the legacy lane resolved it from and
// recovers its declared identity from the requirer's spec.
func TestDraftTransitiveGitDependencyFrozen(t *testing.T) {
	home := t.TempDir()
	skillsRoot := t.TempDir()
	providerRepo := filepath.Join(skillsRoot, "provider")
	if err := os.MkdirAll(providerRepo, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, providerRepo, "provider", false, nil, nil)
	run := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = providerRepo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("add", ".")
	run("commit", "-qm", "provider")
	run("tag", "v1")
	providerCommit, err := gitops.Resolve(providerRepo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}

	project := t.TempDir()
	consumerDir := filepath.Join(project, "pkgs", "review")
	if err := os.MkdirAll(filepath.Join(consumerDir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerDir, "SKILL.md"), []byte("---\nname: review\ndescription: Test review\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	consumerSpec := `{"schema_version":4,"capabilities":{},"commands":{},"dependencies":{"skills":{"provider":{"git":"https://example.org/provider.git","ref":{"kind":"tag","value":"v1"}}}}}`
	if err := os.WriteFile(filepath.Join(consumerDir, "agent-skill.json"), []byte(consumerSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerDir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"pkgs/review"}]}`
	m, raw := parseDraftManifest(t, project, payload)
	plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, SkillsRoot: skillsRoot, Manifest: m, ManifestPayload: raw})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Lock.Members) != 2 {
		t.Fatalf("members = %+v, want review plus transitive provider", plan.Lock.Members)
	}
	providerMember, ok := plan.Lock.Find("provider")
	if !ok {
		t.Fatalf("lock misses transitive provider: %+v", plan.Lock.Members)
	}
	if providerMember.Selection != nil {
		t.Fatalf("transitive provider carries a root selection index: %+v", providerMember)
	}
	nodes, _, err := LoadDraftFrozenNodes(home, plan.Lock, FrozenOptions{GitRepos: map[string]string{"provider": providerRepo}})
	if err != nil {
		t.Fatal(err)
	}
	var provider *Node
	for _, node := range nodes {
		if node.Name == "provider" {
			provider = node
		}
	}
	if provider == nil {
		t.Fatalf("frozen nodes miss provider: %+v", nodes)
	}
	if provider.Resolved.Kind != "tag" || provider.Resolved.Ref != "v1" || provider.Resolved.Commit != providerCommit.Commit {
		t.Fatalf("provider resolved = %+v, want tag v1 %s", provider.Resolved, providerCommit.Commit)
	}
	if provider.Decl.Source != "provider" || provider.Decl.Git != "https://example.org/provider.git" {
		t.Fatalf("provider decl = %+v, want the accepted requirement declaration", provider.Decl)
	}
}
