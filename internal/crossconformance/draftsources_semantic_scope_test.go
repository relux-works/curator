package crossconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

func init() {
	registerDraftSemantic("global-schema2-without-profile-lock-refused", driveGlobalSchema2WithoutProfileLock)
	registerDraftSemantic("project-schema2-without-profile-lock-accepted", driveProjectSchema2WithoutProfileLock)
}

func driveGlobalSchema2WithoutProfileLock(t *testing.T, c draftSemanticCase) {
	root := t.TempDir()
	configPath, _, home := setupCLIProject(t, root)
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "global", "init"); code != 0 {
		t.Fatalf("global init = %d\n%s\n%s", code, stdout, stderr)
	}
	globalSkillfile := filepath.Join(install.GlobalRoot(home), "Skillfile.json")
	initialized, err := manifest.Load(install.GlobalRoot(home))
	if err != nil || initialized == nil || initialized.SchemaVersion != 1 {
		t.Fatalf("global init manifest = %+v, err %v; want schema 1", initialized, err)
	}
	requested := []byte(`{"schema_version":2,"sources":{},"skills":[]}`)
	if err := os.WriteFile(globalSkillfile, requested, 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCurator(t, home, configPath, nil, "global", "install", "--dry-run")
	if code == 0 || !strings.Contains(stderr, "global scope requires schema 1") {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s; want schema-1 upgrade error", code, stdout, stderr)
	}
	after, err := os.ReadFile(globalSkillfile)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(requested) {
		t.Fatal("refused global schema-2 load rewrote the Skillfile")
	}
	got := "upgrade-error;machine-global-scope-remains-schema-1"
	if got != c.Expected {
		t.Fatalf("observed outcome %q differs from corpus expectation %q", got, c.Expected)
	}
}

func driveProjectSchema2WithoutProfileLock(t *testing.T, c draftSemanticCase) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	writeDraftSkill(t, filepath.Join(project, "src", "skills", "review"), "review")
	payload := []byte(`{"schema_version":2,"sources":{"s":{"path":"src"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`)
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app")
	if code != 0 {
		t.Fatalf("project resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil || lock == nil {
		t.Fatalf("project lock = %+v, err %v", lock, err)
	}
	if _, ok := lock.Find("review"); !ok {
		t.Fatal("schema-2 project lock misses review")
	}
	if c.Expected != "accept-schema-2" {
		t.Fatalf("corpus expectation = %q, want accept-schema-2", c.Expected)
	}
}
