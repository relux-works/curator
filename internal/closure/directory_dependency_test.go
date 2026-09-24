package closure

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/sourcelock"
)

const directoryDependencyGit = "https://github.com/example/role-skills.git"

func writeDirectoryDependencySkill(t *testing.T, dir, name string, schema int, requirements map[string]map[string]any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	frontmatter := "---\nname: " + name + "\ndescription: Test " + name + "\n---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(frontmatter), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := map[string]any{
		"schema_version": schema,
		"capabilities":   map[string]any{},
		"commands":       map[string]any{},
		"dependencies":   map[string]any{"skills": requirements},
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func directoryRequirement(commit, directory string, includeDirectory bool) map[string]any {
	entry := map[string]any{
		"git": directoryDependencyGit,
		"ref": map[string]any{"kind": "revision", "value": commit},
	}
	if includeDirectory {
		entry["directory"] = directory
	}
	return entry
}

func resolveDirectoryDependencyFixture(t *testing.T, directory string, includeDirectory bool, repoDirs []string, repoFiles map[string]string, symlinks map[string]string) (*DraftPlan, error) {
	t.Helper()
	project := t.TempDir()
	home := t.TempDir()
	skillsRoot := t.TempDir()
	repo := filepath.Join(skillsRoot, "developer")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, "init", "-q", "-b", "main")
	for _, rel := range repoDirs {
		dir := filepath.Join(repo, filepath.FromSlash(rel))
		if _, isSymlink := symlinks[rel]; isSymlink {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		containsTrackedFile := false
		for file := range repoFiles {
			if strings.HasPrefix(file, strings.TrimSuffix(rel, "/")+"/") {
				containsTrackedFile = true
				break
			}
		}
		if !containsTrackedFile {
			if err := os.WriteFile(filepath.Join(dir, ".fixture"), []byte("directory fixture\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	for rel, content := range repoFiles {
		path := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for rel, target := range symlinks {
		path := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	}
	gitAt(t, repo, "add", ".")
	gitAt(t, repo, "commit", "-qm", "provider")
	resolved, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	appDir := filepath.Join(project, "skills", "app")
	writeDirectoryDependencySkill(t, appDir, "app", 9, map[string]map[string]any{
		"developer": directoryRequirement(resolved.Commit, directory, includeDirectory),
	})
	manifestPayload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"app","from":"s","directory":"skills/app"}]}`
	m, raw := parseDraftManifest(t, project, manifestPayload)
	plan, err := ResolveDraft(DraftResolveConfig{
		ProjectRoot: project, Home: home, SkillsRoot: skillsRoot,
		Manifest: m, ManifestPayload: raw,
	})
	return plan, err
}

func TestDraftManifestDependencyDirectoryResolutionVectors(t *testing.T) {
	vectorPath := filepath.Join("..", "skillspec", "testdata", "draft-sources-v1", "manifest-dependency-directories.json")
	payload, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Cases []struct {
			Name       string `json:"name"`
			Dependency struct {
				Directory string `json:"directory"`
			} `json:"dependency"`
			Snapshot struct {
				Directories      []string          `json:"directories"`
				Files            []string          `json:"files"`
				FrontmatterNames map[string]string `json:"frontmatter_names"`
				Symlinks         map[string]string `json:"symlinks"`
			} `json:"snapshot"`
			Expected struct {
				Status        string `json:"status"`
				Reason        string `json:"reason"`
				LockDirectory string `json:"lock_directory"`
				Identity      struct {
					Directory string `json:"directory"`
				} `json:"identity"`
			} `json:"expected"`
		} `json:"resolution_cases"`
	}
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range vector.Cases {
		if strings.HasPrefix(testCase.Name, "diamond-") || strings.HasPrefix(testCase.Name, "same-name-") {
			continue // Covered by production closure graph tests below.
		}
		t.Run(testCase.Name, func(t *testing.T) {
			files := map[string]string{}
			for _, rel := range testCase.Snapshot.Files {
				name := testCase.Snapshot.FrontmatterNames[rel]
				files[rel] = "---\nname: " + name + "\ndescription: vector fixture\n---\n# " + name + "\n"
			}
			plan, err := resolveDirectoryDependencyFixture(t, testCase.Dependency.Directory, testCase.Dependency.Directory != "", testCase.Snapshot.Directories, files, testCase.Snapshot.Symlinks)
			if testCase.Expected.Status == "rejected" {
				if err == nil {
					t.Fatal("ResolveDraft accepted a rejected directory vector")
				}
				for _, diagnostic := range map[string][]string{
					"directory-missing":            {"source_member_missing"},
					"skill-md-missing":             {"no SKILL.md"},
					"directory-escapes-repository": {"source_selection_invalid", "escapes source"},
				}[testCase.Expected.Reason] {
					if !strings.Contains(err.Error(), diagnostic) {
						t.Fatalf("error %q does not include diagnostic %q", err, diagnostic)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			member, ok := plan.Lock.Find("developer")
			if !ok {
				t.Fatalf("lock misses selected dependency: %+v", plan.Lock.Members)
			}
			wantDirectory := testCase.Expected.LockDirectory
			if wantDirectory == "" {
				wantDirectory = testCase.Expected.Identity.Directory
			}
			if member.Directory != wantDirectory || member.Package.Directory != wantDirectory {
				t.Fatalf("lock member directory=%q package directory=%q, want %q", member.Directory, member.Package.Directory, wantDirectory)
			}
		})
	}
}

func TestDraftManifestDependencyDirectoryDiamondSharesRepositoryAndPackageNode(t *testing.T) {
	project, home, skillsRoot := t.TempDir(), t.TempDir(), t.TempDir()
	repo := filepath.Join(skillsRoot, "backend")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, "init", "-q", "-b", "main")
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "backend"), "backend", 9, map[string]map[string]any{})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "frontend"), "frontend", 9, map[string]map[string]any{})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "shared"), "shared", 4, map[string]map[string]any{})
	gitAt(t, repo, "add", ".")
	gitAt(t, repo, "commit", "-qm", "base role packages")
	resolved, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	baseCommit := resolved.Commit
	// The role packages select shared from the base repository commit. Their
	// selected bytes live at the next commit, while the shared ref remains an
	// immutable non-cyclic pin.
	requireShared := map[string]map[string]any{"shared": directoryRequirement(baseCommit, "skills/shared", true)}
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "backend"), "backend", 9, requireShared)
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "frontend"), "frontend", 9, requireShared)
	gitAt(t, repo, "add", ".")
	gitAt(t, repo, "commit", "-qm", "role dependencies")
	resolved, err = gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	roleCommit := resolved.Commit

	writeDirectoryDependencySkill(t, filepath.Join(project, "skills", "app"), "app", 9, map[string]map[string]any{
		"backend":  directoryRequirement(roleCommit, "skills/backend", true),
		"frontend": directoryRequirement(roleCommit, "skills/frontend", true),
	})
	manifestPayload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"app","from":"s","directory":"skills/app"}]}`
	m, raw := parseDraftManifest(t, project, manifestPayload)
	plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, SkillsRoot: skillsRoot, Manifest: m, ManifestPayload: raw})
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []string{"shared", "backend", "frontend", "app"}
	if len(plan.Nodes) != len(wantOrder) {
		t.Fatalf("nodes = %d, want %v: %+v", len(plan.Nodes), wantOrder, plan.Nodes)
	}
	for i, want := range wantOrder {
		if plan.Nodes[i].Name != want {
			got := make([]string, len(plan.Nodes))
			for index, node := range plan.Nodes {
				got[index] = node.Name
			}
			t.Fatalf("node order = %v, want %v", got, wantOrder)
		}
	}
	for _, name := range []string{"backend", "frontend", "shared"} {
		member, ok := plan.Lock.Find(name)
		if !ok || member.Package.Repository != "github.com/example/role-skills" || member.Package.Directory == "." || member.Directory != member.Package.Directory {
			t.Fatalf("%s lock package does not bind repo directory: %+v", name, member)
		}
	}
	shared, ok := plan.Lock.Find("shared")
	if !ok || shared.Package.Directory != "skills/shared" {
		t.Fatalf("shared node was not unified at its selected folder: %+v", shared)
	}
	frozenNodes, _, err := LoadDraftFrozenNodes(home, plan.Lock, FrozenOptions{
		GitRepos: map[string]string{"backend": repo, "frontend": repo, "shared": repo},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(frozenNodes) != 4 {
		t.Fatalf("frozen nodes = %d, want one shared node for the diamond", len(frozenNodes))
	}
}

func TestDraftManifestDependencyDirectorySameNameDifferentFolderConflicts(t *testing.T) {
	project, home, skillsRoot := t.TempDir(), t.TempDir(), t.TempDir()
	repo := filepath.Join(skillsRoot, "backend")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, "init", "-q", "-b", "main")
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "frontend"), "frontend", 9, map[string]map[string]any{})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "backend"), "backend", 9, map[string]map[string]any{})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "shared-a"), "shared", 4, map[string]map[string]any{})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "shared-b"), "shared", 4, map[string]map[string]any{})
	gitAt(t, repo, "add", ".")
	gitAt(t, repo, "commit", "-qm", "base same-name packages")
	resolved, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	baseCommit := resolved.Commit
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "frontend"), "frontend", 9, map[string]map[string]any{
		"shared": directoryRequirement(baseCommit, "skills/shared-a", true),
	})
	writeDirectoryDependencySkill(t, filepath.Join(repo, "skills", "backend"), "backend", 9, map[string]map[string]any{
		"shared": directoryRequirement(baseCommit, "skills/shared-b", true),
	})
	gitAt(t, repo, "add", ".")
	gitAt(t, repo, "commit", "-qm", "conflicting folder requirements")
	resolved, err = gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	roleCommit := resolved.Commit
	rootDir := filepath.Join(project, "skills", "app")
	writeDirectoryDependencySkill(t, rootDir, "app", 9, map[string]map[string]any{
		"backend":  directoryRequirement(roleCommit, "skills/backend", true),
		"frontend": directoryRequirement(roleCommit, "skills/frontend", true),
	})
	manifestPayload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"app","from":"s","directory":"skills/app"}]}`
	m, raw := parseDraftManifest(t, project, manifestPayload)
	_, err = ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, SkillsRoot: skillsRoot, Manifest: m, ManifestPayload: raw})
	if err == nil || !strings.Contains(err.Error(), "source_name_conflict") || !strings.Contains(err.Error(), "selected directory") {
		t.Fatalf("same skill name at different package folders error = %v, want source_name_conflict", err)
	}
}

func TestDraftManifestDependencyDirectoryRequiresMatchingSkillName(t *testing.T) {
	_, err := resolveDirectoryDependencyFixture(t, "skills/developer", true, []string{"skills", "skills/developer"}, map[string]string{
		"skills/developer/SKILL.md": "---\nname: another-skill\ndescription: Wrong package\n---\n# another-skill\n",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "differs from dependency name") {
		t.Fatalf("mismatched SKILL.md name error = %v, want selected package name refusal", err)
	}
}

func TestDraftManifestDependencyDirectoryOmittedLockShapeRemainsRoot(t *testing.T) {
	plan, err := resolveDirectoryDependencyFixture(t, "", false, nil, map[string]string{"SKILL.md": "---\nname: developer\ndescription: root\n---\n# developer\n"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := plan.Lock.Find("developer")
	if !ok || member.Directory != "." || member.Package.Directory != "." || member.Package.Kind != sourcelock.KindNetworkGit {
		t.Fatalf("omitted directory did not preserve root package identity: %+v", member)
	}
}

func TestLegacyDependencyWithoutDirectoryLockBytes(t *testing.T) {
	project, home, skillsRoot := t.TempDir(), t.TempDir(), t.TempDir()
	provider := filepath.Join(skillsRoot, "provider")
	if err := os.MkdirAll(provider, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, provider, "provider", false, nil, nil)
	gitAtFixedDate(t, provider, "init", "-q", "-b", "main")
	gitAtFixedDate(t, provider, "add", ".")
	gitAtFixedDate(t, provider, "commit", "-qm", "provider")
	commit, err := gitops.Resolve(provider, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	writeDirectoryDependencySkill(t, filepath.Join(project, "skills", "app"), "app", 8, map[string]map[string]any{
		"provider": directoryRequirement(commit.Commit, "", false),
	})
	manifestPayload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"app","from":"s","directory":"skills/app"}]}`
	m, raw := parseDraftManifest(t, project, manifestPayload)
	plan, err := ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, SkillsRoot: skillsRoot, Manifest: m, ManifestPayload: raw})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := plan.Lock.Find("provider")
	if !ok || member.Directory != "." || member.Package.Directory != "." {
		t.Fatalf("legacy member = %+v, want repository-root package", member)
	}
	// This deterministic schema-8/no-directory lock hash is the output from
	// the unmodified story base and pins byte-identical legacy serialization.
	const baselineLockSHA256 = "sha256:685572494b1612df0e2f8be58102a9bb01037d404f513c0e756da5103fd5d66c"
	if plan.Lock.LockSHA256 != baselineLockSHA256 {
		t.Fatalf("legacy lock hash = %s, want unchanged baseline %s", plan.Lock.LockSHA256, baselineLockSHA256)
	}
}

func gitAtFixedDate(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
