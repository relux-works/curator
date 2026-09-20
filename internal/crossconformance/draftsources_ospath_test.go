package crossconformance

// OS path-semantics integration for draft sources.
//
// The semantic suites prove selection and boundary rules against the
// snapshot validator in-process. These tests prove the same behaviors
// survive the real filesystem and the compiled CLI: hostile names
// (symlink members escaping the package) refuse with a stable code,
// ordinary-but-awkward names (spaces, unicode) resolve and install,
// and the manager binary still builds for the platforms this host
// cannot execute.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testcli"
)

func TestDraftSourcesCLIProjectPathWithSpaceAndUnicode(t *testing.T) {
	root := filepath.Join(t.TempDir(), "my project", "ünicode")
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review")
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app"); code != 0 {
		t.Fatalf("resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := lock.Find("review"); !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "install", "app", "--dry-run"); code != 0 {
		t.Fatalf("install --dry-run = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

func TestDraftSourcesCLISymlinkMemberRefused(t *testing.T) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	skill := filepath.Join(project, "skills", "review")
	writeDraftSkill(t, skill, "review")
	outside := filepath.Join(root, "outside", "secret.txt")
	if err := os.MkdirAll(filepath.Dir(outside), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(skill, "evil-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	code, _, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app")
	if code == 0 {
		t.Fatal("resolve succeeded with a symlink member escaping the package")
	}
	if !strings.Contains(stderr, "source_member_invalid") && !strings.Contains(stderr, "source_selection_invalid") {
		t.Fatalf("stderr misses the link refusal:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

func TestDraftSourcesCrossCompile(t *testing.T) {
	// No testing.Short carve-out: the skip vocabulary in
	// .github/ci/skip-classes.tsv names no short-run class, and an
	// unmatched skip reason is fatal in the platform-case gate.
	for _, target := range []struct {
		goos, suffix string
	}{
		{"windows", ".exe"},
		{"linux", ""},
	} {
		t.Run(target.goos, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "curator-"+target.goos+target.suffix)
			testcli.GoBuild(t, testcli.ModuleRoot(t), out, []string{"GOOS=" + target.goos, "GOARCH=amd64", "CGO_ENABLED=0"})
			info, err := os.Stat(out)
			if err != nil || info.Size() == 0 {
				t.Fatalf("GOOS=%s binary missing: %v", target.goos, err)
			}
		})
	}
}
