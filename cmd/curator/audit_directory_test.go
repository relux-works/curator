package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/manifest"
)

// auditDirectoryProject lays out a schema-1 project whose schema-9 consumer
// requires a subdirectory-selected schema-9 package, and returns the config
// path that points the CLI at it, the manager home that will hold the
// persisted verdict records, the project root, and the locked provider and
// consumer commits.
func auditDirectoryProject(t *testing.T) (configPath, home, project, providerCommit, consumerCommit string) {
	t.Helper()
	root := t.TempDir()
	home = filepath.Join(root, "home")
	configPath = filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project = filepath.Join(root, "project")
	provider := filepath.Join(skillsRoot, "backend")
	consumer := filepath.Join(skillsRoot, "consumer")
	for _, dir := range []string{provider, consumer, project} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeAuditDirectorySkill(t, filepath.Join(provider, "skills", "backend"), "backend", 9, nil)
	runGit(t, provider, "init", "-q", "-b", "main")
	runGit(t, provider, "add", ".")
	runGit(t, provider, "commit", "-qm", "role package")
	runGit(t, provider, "tag", "v1")
	providerCommit = runGitOutput(t, provider, "rev-parse", "HEAD")

	writeAuditDirectorySkill(t, consumer, "consumer", 9, map[string]any{
		"backend": map[string]any{
			"git":       "https://github.com/example/role-skills.git",
			"ref":       map[string]any{"kind": "tag", "value": "v1"},
			"directory": "skills/backend",
		},
	})
	runGit(t, consumer, "init", "-q", "-b", "main")
	runGit(t, consumer, "add", ".")
	runGit(t, consumer, "commit", "-qm", "init")
	runGit(t, consumer, "tag", "v1")
	consumerCommit = runGitOutput(t, consumer, "rev-parse", "HEAD")

	runGit(t, project, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"consumer","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	return configPath, home, project, providerCommit, consumerCommit
}

// runGitOutput runs git like runGit and returns its trimmed standard output.
func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	command := exec.Command("git", gitArgs...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(output))
}

func writeAuditDirectorySkill(t *testing.T, dir, name string, schema int, requirements map[string]any) {
	t.Helper()
	if requirements == nil {
		requirements = map[string]any{}
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: "+name+"\ndescription: Test "+name+"\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"schema_version": schema,
		"capabilities":   map[string]any{},
		"commands":       map[string]any{},
		"dependencies":   map[string]any{"skills": requirements},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCLIAuditBindsDependencyDirectory drives `curator audit` over a
// schema-1 project with a subdirectory-selected dependency and proves the
// persisted verdicts bind the complete package identity (core §4.4,
// skillfile-sources §4 cache equality): the subfolder package binds its
// network-git repository, locked commit, and directory, and the root
// consumer binds its configured-git root identity. Without the subject
// identity the backend verdict would be stored under the root and a later
// audit of identical content in another repository, commit, or directory
// would reuse it.
func TestCLIAuditBindsDependencyDirectory(t *testing.T) {
	t.Parallel()
	configPath, home, _, providerCommit, consumerCommit := auditDirectoryProject(t)
	code, stdout, stderr := capture(t, configPath, "audit", "app")
	if code != exitOK {
		t.Fatalf("audit = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	matches, err := filepath.Glob(filepath.Join(home, "audit", "*", "verdict-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	type auditPackage struct {
		Kind       string `json:"kind"`
		Repository string `json:"repository"`
		Source     string `json:"source"`
		Commit     struct {
			ObjectFormat string `json:"object_format"`
			Hex          string `json:"hex"`
		} `json:"commit"`
		Directory string `json:"directory"`
	}
	type auditRecord struct {
		Skill     string       `json:"skill"`
		Commit    string       `json:"commit"`
		Directory string       `json:"directory"`
		Package   auditPackage `json:"package"`
	}
	audited := map[string]auditRecord{}
	for _, path := range matches {
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var record auditRecord
		if err := json.Unmarshal(payload, &record); err != nil {
			t.Fatal(err)
		}
		audited[record.Skill] = record
	}
	backend, ok := audited["backend"]
	if !ok {
		t.Fatalf("no audit record for backend (records: %v)", audited)
	}
	if backend.Directory != "skills/backend" || backend.Commit != providerCommit ||
		backend.Package.Kind != "network-git" || backend.Package.Repository != "github.com/example/role-skills" ||
		backend.Package.Commit.Hex != providerCommit || backend.Package.Directory != "skills/backend" {
		t.Fatalf("audit record for backend binds %+v, want the subfolder package identity", backend)
	}
	consumer, ok := audited["consumer"]
	if !ok {
		t.Fatalf("no audit record for consumer (records: %v)", audited)
	}
	if consumer.Directory != "." || consumer.Commit != consumerCommit ||
		consumer.Package.Kind != "configured-git" || consumer.Package.Source != "consumer" ||
		consumer.Package.Commit.Hex != consumerCommit || consumer.Package.Directory != "." {
		t.Fatalf("audit record for consumer binds %+v, want the configured-git root identity", consumer)
	}
}

// TestCLIStatusReportsLegacyPackageMarkerCurrent is the F1 currentness row
// on the status surface: after a real CLI install of a schema-1 project
// with schema-9 skills, `curator status` reports the migrated
// installations up-to-date by comparing the recorded package and lock
// against the staged effective plan.
func TestCLIStatusReportsLegacyPackageMarkerCurrent(t *testing.T) {
	t.Parallel()
	configPath, _, _, _, _ := auditDirectoryProject(t)
	if code, stdout, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	report := decodeStatus(t, stdout)
	if report.Skills["consumer"] != stateUpToDate {
		t.Fatalf("skills = %v, want consumer up-to-date", report.Skills)
	}
	if code, _, stderr := capture(t, configPath, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check = %d, want %d\nstderr:\n%s", code, exitOK, stderr)
	}
}

// TestReviewCLIStatusDetectsChangedTagAtSameCommit is the F6 regression
// through the real CLI: after changing the declared tag from v1 to an
// unchanged v2 at the same commit, the v6 package installation is
// non-current — the staged lock binds the declaration, so commit
// equality alone must not report up-to-date (skillfile-sources §4).
// The narrowing mutant `status-commit-only` compares the recorded
// commit against the live resolution instead of the staged package and
// lock; it fails both assertions below. The currentness test above pins
// that an unchanged installation still reports up-to-date.
func TestReviewCLIStatusDetectsChangedTagAtSameCommit(t *testing.T) {
	t.Parallel()
	configPath, home, project, _, _ := auditDirectoryProject(t)
	if code, stdout, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	runGit(t, filepath.Join(filepath.Dir(home), "skills", "consumer"), "tag", "v2")
	path := filepath.Join(project, "Skillfile.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	doc["skills"].([]any)[0].(map[string]any)["tag"] = "v2"
	payload, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	report := decodeStatus(t, stdout)
	if report.Skills["consumer"] == stateUpToDate {
		t.Errorf("changed declaration is incorrectly current: %s", stdout)
	}
	if code, stdout, stderr := capture(t, configPath, "status", "app", "--check"); code == exitOK {
		t.Errorf("status --check admitted stale ref binding: %s %s", stdout, stderr)
	}
}
