package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/envmarker"
)

type nativeGlobalFixture struct {
	home, config, repo string
}

func newNativeGlobalFixture(t *testing.T) nativeGlobalFixture {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(home, ".pi"))
	fixture := nativeGlobalFixture{home: home, config: filepath.Join(home, ".curator", "config.json"), repo: t.TempDir()}
	t.Setenv("CURATOR_CONFIG", fixture.config)
	writeCLISkill(t, fixture.repo, "shared-skill")
	runGit(t, fixture.repo, "init", "-q", "-b", "main")
	runGit(t, fixture.repo, "add", ".")
	runGit(t, fixture.repo, "commit", "-qm", "first skill")
	runGit(t, fixture.repo, "tag", "v1.0.0")
	runGit(t, fixture.repo, "tag", "same-commit")
	writeFile(t, filepath.Join(fixture.repo, "SKILL.md"), "# shared-skill\nsecond pin\n")
	runGit(t, fixture.repo, "add", ".")
	runGit(t, fixture.repo, "commit", "-qm", "second skill")
	runGit(t, fixture.repo, "tag", "v2.0.0")
	gitconfig := filepath.Join(t.TempDir(), "gitconfig")
	writeFile(t, gitconfig, "[url \""+gitFileURL(fixture.repo)+"\"]\n\tinsteadOf = https://example.com/shared-skill\n")
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	fixture.requireAbsent(t)
	fixture.command(t, exitOK, "bootstrap", "--if-missing", "--non-interactive", "--skills-root", filepath.Join(home, "sources"))
	fixture.requireAbsent(t)
	fixture.command(t, exitOK, "global", "init")
	fixture.requireAbsent(t)
	return fixture
}

func (fixture nativeGlobalFixture) command(t *testing.T, want int, args ...string) (string, string) {
	t.Helper()
	code, stdout, stderr := runFile(t, fixture.config, args...)
	if code != want {
		t.Fatalf("curator %v = %d, want %d; stdout=%s stderr=%s", args, code, want, stdout, stderr)
	}
	if strings.Contains(stdout+stderr, "taking over unmanaged") {
		t.Fatalf("global flow performed unsolicited takeover: %s%s", stdout, stderr)
	}
	return stdout, stderr
}

func (fixture nativeGlobalFixture) requireAbsent(t *testing.T) {
	t.Helper()
	for _, agent := range []string{".claude", ".codex"} {
		if _, err := os.Lstat(filepath.Join(fixture.home, agent, "skills", "shared-skill")); !os.IsNotExist(err) {
			t.Fatalf("native skill appeared before global install in %s: %v", agent, err)
		}
	}
}

func (fixture nativeGlobalFixture) declare(t *testing.T, tag string) {
	t.Helper()
	writeFile(t, filepath.Join(fixture.home, ".curator", "global", "Skillfile.json"),
		`{"schema_version":1,"agents":["codex_cli","claude_code"],"skills":[{"name":"shared-skill","git":"https://example.com/shared-skill","tag":"`+tag+`"}]}`)
}

func (fixture nativeGlobalFixture) requireProfileOwner(t *testing.T) map[string]string {
	t.Helper()
	links := map[string]string{}
	for _, agent := range []string{".claude", ".codex"} {
		native := filepath.Join(fixture.home, agent)
		target := filepath.Join(native, "skills", "shared-skill")
		link, err := os.Readlink(target)
		if err != nil || !strings.Contains(filepath.ToSlash(link), "/contexts/skill-context/") {
			t.Fatalf("profile native owner lost: %q, %v", link, err)
		}
		links[agent] = link
		marker, err := envmarker.Read(native)
		if err != nil || marker == nil {
			t.Fatalf("profile native marker missing: %v", err)
		}
		paths := marker.Surfaces[envmarker.SurfaceSkills].Paths
		if len(paths) != 1 || paths[0] != "skills/shared-skill" {
			t.Fatalf("profile native ownership = %v", paths)
		}
		ledger, err := os.ReadFile(filepath.Join(native, "skills", adapters.LedgerName))
		if err == nil {
			var data struct {
				Entries []string `json:"entries"`
			}
			if err := json.Unmarshal(ledger, &data); err != nil {
				t.Fatal(err)
			}
			for _, name := range data.Entries {
				if name == "shared-skill" {
					t.Fatal("profile skill also claimed by legacy ledger")
				}
			}
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return links
}

// A clean production bootstrap, global init and no-flag global install first
// publish each native link exactly once under profile ownership, without
// takeover; retry and profile sync retain that owner and exact link target.
func TestGlobalNativeOwnerCleanBootstrapInstall(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	fixture.declare(t, "v1.0.0")
	fixture.requireAbsent(t) // The regression first creates the entry inside install.
	fixture.command(t, exitOK, "global", "install")
	links := fixture.requireProfileOwner(t)
	for _, agent := range []string{".claude", ".codex"} {
		if _, err := os.Lstat(filepath.Join(fixture.home, agent, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("first clean install created a backup: %v", err)
		}
	}
	for _, args := range [][]string{{"global", "install"}, {"profile", "sync"}} {
		fixture.command(t, exitOK, args...)
		for agent, link := range fixture.requireProfileOwner(t) {
			if link != links[agent] {
				t.Fatalf("profile link changed after %v", args)
			}
		}
	}
}

// Exact declarations with different commits conflict per environments §§1.4,
// 9.4, even for one skill name; no precedence or second native owner is invented.
func TestGlobalNativeOwnerDifferentPinRefuses(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	profile := t.TempDir()
	writeContextPackage(t, profile, "selected-profile", "1.0.0", "profile context\n")
	writeFile(t, filepath.Join(profile, "agent-context.json"), `{"schema_version":1,"name":"selected-profile","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"shared-skill":{"git":"https://example.com/shared-skill","tag":"v1.0.0"}}}}`)
	fixture.command(t, exitOK, "profile", "install", profile, "--use")
	before := fixture.requireProfileOwner(t)
	fixture.declare(t, "v2.0.0")
	snapshots := map[string]map[string]helpTreeEntry{}
	for _, path := range []string{filepath.Join(fixture.home, ".claude"), filepath.Join(fixture.home, ".codex"), filepath.Join(fixture.home, ".curator", "profiles", "selected-profile")} {
		snapshots[path] = snapshotHelpTree(t, path)
	}
	_, stderr := fixture.command(t, exitFail, "global", "install")
	if !strings.Contains(stderr, "context_range_conflict") || !strings.Contains(stderr, "shared-skill") || !strings.Contains(stderr, "machine") || !strings.Contains(stderr, "selected-profile") {
		t.Fatalf("different-pin rejection lost its identity diagnostic: %s", stderr)
	}
	for path, before := range snapshots {
		if after := snapshotHelpTree(t, path); !reflect.DeepEqual(after, before) {
			t.Fatalf("different-pin rejection changed mutable state: %s", path)
		}
	}
	for agent, link := range fixture.requireProfileOwner(t) {
		if link != before[agent] {
			t.Fatal("different-pin rejection changed native link")
		}
	}
	// Matching the profile's pin is the nearby positive control.
	fixture.declare(t, "same-commit")
	fixture.command(t, exitOK, "global", "install")
	fixture.requireProfileOwner(t)
}

// A Skillfile written after a profile already exists extends that profile's
// lock and native surfaces, rather than silently installing a legacy-only copy.
// A later direct pin change replaces the old direct constraint, while actual
// profile requirements remain constraints (the conflicting-pin test above).
func TestGlobalNativeOwnerNewDeclarationAfterProfile(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	profile := t.TempDir()
	writeContextPackage(t, profile, "empty-profile", "1.0.0", "profile context\n")
	fixture.command(t, exitOK, "profile", "install", profile, "--use")
	fixture.requireAbsent(t)
	fixture.declare(t, "v1.0.0")
	fixture.command(t, exitOK, "global", "install")
	fixture.requireProfileOwner(t)
	fixture.declare(t, "v2.0.0")
	fixture.command(t, exitOK, "global", "install")
	fixture.requireProfileOwner(t)
	for _, agent := range []string{".claude", ".codex"} {
		body, err := os.ReadFile(filepath.Join(fixture.home, agent, "skills", "shared-skill", "SKILL.md"))
		if err != nil || !strings.Contains(string(body), "second pin") {
			t.Fatalf("updated direct pin did not reach native surface: %s %v", body, err)
		}
	}
}

// A genuinely foreign native directory survives the production global install
// refusal unchanged, without a legacy ledger, backup, or partial native link.
func TestGlobalNativeOwnerForeignRefuses(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	fixture.declare(t, "v1.0.0")
	foreign := filepath.Join(fixture.home, ".claude", "skills", "shared-skill", "SKILL.md")
	writeFile(t, foreign, "foreign preimage\n")
	_, stderr := fixture.command(t, exitFail, "global", "install")
	if !strings.Contains(stderr, "environment_surface_unmanaged_conflict") {
		t.Fatalf("foreign-entry rejection = %s", stderr)
	}
	if data, err := os.ReadFile(foreign); err != nil || string(data) != "foreign preimage\n" {
		t.Fatalf("foreign entry changed: %q, %v", data, err)
	}
	for _, path := range []string{filepath.Join(fixture.home, ".codex", "skills", "shared-skill"), filepath.Join(filepath.Dir(foreign), "..", adapters.LedgerName), filepath.Join(fixture.home, ".claude", ".agent-environment-backup")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("foreign refusal wrote %s: %v", path, err)
		}
	}
}
