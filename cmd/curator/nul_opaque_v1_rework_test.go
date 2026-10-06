package main

// Rework-1 regressions for finding 1 at the CLI production entry: `status
// --check` and reinstall over a v1 installation after NUL bytes appear —
// including a v1-collision-preserving edit — refuse with the opaque
// finding instead of reporting up-to-date. The v2 lane admits the same
// bytes with a v2 identity. These tests run on the hosted gate (R194),
// like every cmd/curator suite.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// addCLISkillFilesAndRetag appends files to a CLI skill repository fixture
// and moves its v1 tag onto the new commit.
func addCLISkillFilesAndRetag(t *testing.T, skill string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		writeFile(t, filepath.Join(skill, filepath.FromSlash(name)), content)
	}
	runGit(t, skill, "add", ".")
	runGit(t, skill, "commit", "-qm", "add opaque snapshot files")
	runGit(t, skill, "tag", "-f", "v1")
}

func cliSkillsRoot(home string) string {
	return filepath.Join(filepath.Dir(home), "skills")
}

// A v1 installation that gains NUL bytes afterwards refuses at every
// status surface: `status --check` exits nonzero with the opaque finding,
// reinstall fails the same way, and the drift classifiers report
// non-current without hashing. The collision-preserving edit proves the
// pre-hash guard specifically, because the v1 digests match by
// construction and only the scan-before-hash can refuse.
func TestStatusCheckRefusesV1NULAppearingAfterInstall(t *testing.T) {
	pinV1MarkerWriters(t)
	project, home := legacyProject(t)
	configPath := filepath.Join(home, "config.json")
	skill := filepath.Join(cliSkillsRoot(home), "skill-a")
	addCLISkillFilesAndRetag(t, skill, map[string]string{
		"assets/a.bin": "x",
		"assets/b.md":  "y",
	})
	if code, stdout, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("clean v1 install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	skillsDir := filepath.Join(project, ".agents", "skills")
	recorded := marker.Read(installed)
	if recorded == nil || recorded.ContentHashVersion() != hashing.VersionV1 {
		t.Fatalf("installed marker = %+v, want a recorded v1 identity", recorded)
	}

	code, stdout, stderr := capture(t, configPath, "status", "app", "--json", "--check")
	if code != exitOK {
		t.Fatalf("clean v1 status --check = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if report := decodeStatus(t, stdout); report.Skills["skill-a"] != stateUpToDate {
		t.Fatalf("clean v1 skills = %v, want skill-a up-to-date", report.Skills)
	}
	cfg := &config.Config{SkillsRoot: cliSkillsRoot(home)}
	if drift := scopeStatusDrift(cfg, project, skillsDir); drift["skill-a"] != stateUpToDate {
		t.Fatalf("clean v1 drift = %v, want skill-a up-to-date", drift)
	}
	if state := classifyDraftMember(skillsDir, sourcelock.Member{Name: "skill-a"}, "", nil); state != stateNeedsInstall {
		t.Fatalf("clean legacy member draft classification = %q, want needs-install", state)
	}

	before, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(installed, "assets", "a.bin"), "x\x00assets/b.md\x00y")
	if err := os.Remove(filepath.Join(installed, "assets", "b.md")); err != nil {
		t.Fatal(err)
	}
	after, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("fixture is not a v1 collision: %s != %s", before, after)
	}

	code, _, stderr = capture(t, configPath, "status", "app", "--check")
	if code != exitFail {
		t.Fatalf("status --check after v1-collision edit = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	}
	if !strings.Contains(stderr, "audit.opaque.nul-byte") {
		t.Fatalf("status --check stderr lacks the opaque finding:\n%s", stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitFail ||
		!strings.Contains(stderr, "audit.opaque.nul-byte") {
		t.Fatalf("reinstall after v1-collision edit = %d, want a failing opaque refusal\nstderr:\n%s", code, stderr)
	}
	if drift := scopeStatusDrift(cfg, project, skillsDir); drift["skill-a"] != stateContentDrift {
		t.Fatalf("drift after v1-collision edit = %v, want skill-a content-drift", drift)
	}
	if state := classifyDraftMember(skillsDir, sourcelock.Member{Name: "skill-a"}, "", nil); state != stateUnresolvable {
		t.Fatalf("draft classification after v1-collision edit = %q, want unresolvable", state)
	}
}

// The same NUL bytes under v2 install, report current, and re-check
// clean: NUL is ordinary v2 data, and the recorded identity is v2.
func TestStatusCheckAdmitsV2NULBearingInstall(t *testing.T) {
	pinAuditHashWriters(t, true)
	project, home := legacyProject(t)
	configPath := filepath.Join(home, "config.json")
	skill := filepath.Join(cliSkillsRoot(home), "skill-a")
	addCLISkillFilesAndRetag(t, skill, map[string]string{
		"assets/a.bin": "x\x00assets/b.md\x00y",
	})
	if code, stdout, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("v2 NUL install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	recorded := marker.Read(installed)
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 || recorded.HashVersion != hashing.VersionV2 {
		t.Fatalf("installed marker = %+v, want a recorded v2 identity", recorded)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app", "--json", "--check")
	if code != exitOK {
		t.Fatalf("v2 NUL status --check = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if report := decodeStatus(t, stdout); report.Skills["skill-a"] != stateUpToDate {
		t.Fatalf("v2 NUL skills = %v, want skill-a up-to-date", report.Skills)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("v2 NUL reinstall = %d, want %d\nstderr:\n%s", code, exitOK, stderr)
	}
}
