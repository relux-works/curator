package envprofile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/skillintegrity"
)

func TestProfileFileSelectionInstallUpdateAndProjectBoundedSource(t *testing.T) {
	pinHomes(t)
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	selection := map[string]any{
		"include": []string{"data/public.json"},
		"exclude": []string{".auxiliary-cache/**", ".excluded-cache/**"},
	}
	skillRepo := gitRepo(t, map[string]string{
		"SKILL.md":                        "---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n",
		"references/guide.md":             "Guide.\n",
		"data/public.json":                "{}\n",
		"agent-skill.json":                `{"schema_version":8,"capabilities":{},"build_roots":["tools/runner"],"commands":{"runner":{"type":"build","driver":"go-v1","source_dir":"tools/runner/cmd/runner","modules":["modules/support"]}}}`,
		"tools/runner/go.mod":             "module example.invalid/runner\n\ngo 1.24\n",
		"tools/runner/cmd/runner/main.go": "package main\nfunc main() {}\n",
		"tools/runner/vendor/modules.txt": "# example.invalid/support v0.1.0\n",
		"tools/runner/vendor/example.invalid/support/support.go": "package support\n",
		"modules/support/go.mod":                                 "module example.invalid/support\n\ngo 1.24\n",
		"modules/support/support.go":                             "package support\n",
	}, "v1.0.0")
	largePath := filepath.Join(skillRepo, ".auxiliary-cache", "over-limit.bin")
	objects := filepath.Join(skillRepo, ".excluded-cache", "objects")
	if err := os.MkdirAll(objects, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < skillintegrity.MaxFiles+1; i++ {
		name := filepath.Join(objects, fmt.Sprintf("%05d.bin", i))
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatalf("create excluded object %d: %v", i, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(largePath), 0o755); err != nil {
		t.Fatal(err)
	}
	large, err := os.Create(largePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(skillintegrity.MaxBytes + 1); err != nil {
		_ = large.Close()
		t.Fatal(err)
	}
	if err := large.Close(); err != nil {
		t.Fatal(err)
	}
	gitRun(t, skillRepo, "add", ".auxiliary-cache", ".excluded-cache")
	gitRun(t, skillRepo, "commit", "-m", "add excluded capture overflow")
	skillCommit := fixtureGitOutput(t, skillRepo, "rev-parse", "HEAD")
	skillURL := ids.serve(skillRepo, "https://example.org/fixture-skill")

	rootManifest := profileFileSelectionManifest(t, "sample-profile", "1.0.0", skillURL, skillCommit, selection)
	profileRepo := gitRepo(t, map[string]string{"agent-context.json": rootManifest}, "v1.0.0")
	profileURL := ids.serve(profileRepo, "https://example.org/sample-profile")
	info, _, _, err := Install(home, InstallOptions{Operand: profileURL, Range: "*"})
	if err != nil {
		t.Fatalf("filtered profile install failed: %v", err)
	}
	if info.Lock.SchemaVersion != contextlock.SkillIntegritySchemaVersion {
		t.Fatalf("lock schema=%d, want lock 6", info.Lock.SchemaVersion)
	}
	assertProfileFileSelectionEvidence(t, info.Lock)

	updatedManifest := profileFileSelectionManifest(t, "sample-profile", "1.0.1", skillURL, skillCommit, selection)
	fullName := filepath.Join(profileRepo, "agent-context.json")
	if err := os.WriteFile(fullName, []byte(updatedManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, profileRepo, "add", "agent-context.json")
	gitRun(t, profileRepo, "commit", "-m", "update profile version")
	gitRun(t, profileRepo, "tag", "v1.0.1")
	updated, _, err := UpdateWithPolicy(home, info.Name, Policy{})
	if err != nil {
		t.Fatalf("filtered profile update failed: %v", err)
	}
	assertProfileFileSelectionEvidence(t, updated.Lock)
	if _, err := UseWithPolicy(home, updated.Name, "codex_cli", "", false, Policy{}); err != nil {
		t.Fatalf("updated profile use failed: %v", err)
	}
	assertProfileFileSelectionProjection(t, home, updated.Lock)

	plainManifest := profileFileSelectionManifest(t, "plain-profile", "1.0.0", skillURL, skillCommit, nil)
	plainRepo := gitRepo(t, map[string]string{"agent-context.json": plainManifest}, "v1.0.0")
	plainURL := ids.serve(plainRepo, "https://example.org/plain-profile")
	if _, _, _, err := Install(home, InstallOptions{Operand: plainURL, Range: "*"}); !skillintegrity.IsCode(err, "skill_integrity_manifest_invalid") || !strings.Contains(err.Error(), "capture budget exceeded") {
		t.Fatalf("unfiltered install did not hit the unchanged capture limit: %v", err)
	}
	if _, err := os.Stat(lockPath(home, "plain-profile")); !os.IsNotExist(err) {
		t.Fatalf("unfiltered refusal published a profile lock: %v", err)
	}
}

func profileFileSelectionManifest(t *testing.T, name, version, skillURL, skillCommit string, selection map[string]any) string {
	t.Helper()
	requirement := map[string]any{"git": skillURL, "revision": skillCommit}
	if selection != nil {
		requirement["files"] = selection
	}
	manifest := map[string]any{
		"schema_version": 5,
		"name":           name,
		"version":        version,
		"requires": map[string]any{
			"skills": map[string]any{"fixture-skill": requirement},
		},
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload) + "\n"
}

func assertProfileFileSelectionEvidence(t *testing.T, lock *contextlock.Lock) {
	t.Helper()
	want := []string{"SKILL.md", "data/public.json", "references/guide.md"}
	if lock.SchemaVersion != contextlock.SkillIntegritySchemaVersion || len(lock.SkillIntegrity) != 1 {
		t.Fatalf("unexpected lock evidence: schema=%d entries=%d", lock.SchemaVersion, len(lock.SkillIntegrity))
	}
	entry := lock.SkillIntegrity[0]
	if !slices.Equal(entry.Manifest.ContextFiles, want) {
		t.Fatalf("locked context files = %v, want %v", entry.Manifest.ContextFiles, want)
	}
	for _, path := range []string{"agent-skill.json", "tools/runner/go.mod", "tools/runner/cmd/runner/main.go", "tools/runner/vendor/modules.txt", "modules/support/go.mod", "modules/support/support.go"} {
		if !hasIntegrityFile(entry.Manifest.Files, path) {
			t.Errorf("declared build input %q missing from lock evidence", path)
		}
	}
	for _, path := range []string{".auxiliary-cache/over-limit.bin", ".excluded-cache/objects/00000.bin"} {
		if hasIntegrityFile(entry.Manifest.Files, path) {
			t.Errorf("excluded file %q entered lock evidence", path)
		}
	}
}

func assertProfileFileSelectionProjection(t *testing.T, home string, lock *contextlock.Lock) {
	t.Helper()
	member, ok := lock.Find(contextlock.KindSkill, "fixture-skill")
	if !ok {
		t.Fatal("selected skill missing from lock")
	}
	root := packageRoot(newGitManager(home).entryPath(home, resolvedOf(member)), member.Directory)
	projection, err := projectSkillWithFiles(home, root, lock.ContentHashVersion(), lock.SkillIntegrity[0].Manifest.FileSelection)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := filepath.WalkDir(projection.skill.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(projection.skill.root, path)
		if err == nil {
			got = append(got, filepath.ToSlash(rel))
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	if !slices.Equal(got, projection.files) || !slices.Equal(got, lock.SkillIntegrity[0].Manifest.ContextFiles) {
		t.Fatalf("installed projection = %v, planned=%v, locked=%v", got, projection.files, lock.SkillIntegrity[0].Manifest.ContextFiles)
	}
}

func hasIntegrityFile(files []skillintegrity.File, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}
