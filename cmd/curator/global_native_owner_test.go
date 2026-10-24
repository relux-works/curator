package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/testgitenv"
	"github.com/relux-works/curator/internal/transaction"
)

type nativeGlobalFixture struct {
	home, config, repo string
}

func newNativeGlobalFixture(t *testing.T) nativeGlobalFixture {
	t.Helper()
	home := testgitenv.ShortTempDir(t)
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
		snapshots[path] = nativeOwnerSnapshot(t, path)
	}
	_, stderr := fixture.command(t, exitFail, "global", "install")
	if !strings.Contains(stderr, "context_range_conflict") || !strings.Contains(stderr, "shared-skill") || !strings.Contains(stderr, "machine") || !strings.Contains(stderr, "selected-profile") {
		t.Fatalf("different-pin rejection lost its identity diagnostic: %s", stderr)
	}
	for path, before := range snapshots {
		if after := nativeOwnerSnapshot(t, path); !reflect.DeepEqual(after, before) {
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

// Removing a machine declaration retires its native entries and lock member;
// an independently required profile member remains active after the same flow.
func TestGlobalNativeOwnerRemoval(t *testing.T) {
	for _, required := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct-only", true: "profile-required"}[required], func(t *testing.T) {
			fixture := newNativeGlobalFixture(t)
			profile := t.TempDir()
			writeContextPackage(t, profile, "selected-profile", "1.0.0", "profile context\n")
			if required {
				writeFile(t, filepath.Join(profile, "agent-context.json"), `{"schema_version":1,"name":"selected-profile","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"shared-skill":{"git":"https://example.com/shared-skill","tag":"v1.0.0"}}}}`)
			}
			fixture.command(t, exitOK, "profile", "install", profile, "--use")
			fixture.declare(t, "v1.0.0")
			fixture.command(t, exitOK, "global", "install")
			fixture.requireProfileOwner(t)
			fixture.command(t, exitOK, "global", "remove", "shared-skill")
			fixture.command(t, exitOK, "global", "install")
			if required {
				fixture.requireProfileOwner(t)
			} else {
				fixture.requireAbsent(t)
			}
			lock, err := os.ReadFile(filepath.Join(fixture.home, ".curator", "profiles", "selected-profile", "lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(lock), `"name":"shared-skill"`) != required {
				t.Fatalf("removed declaration persisted in profile lock: %s", lock)
			}
		})
	}
}

// A canonical global-store ownership refusal cannot activate a new profile
// member or change existing profile/native surfaces; a clean retry can.
func TestGlobalNativeOwnerCanonicalRefusal(t *testing.T) {
	for _, command := range []string{"install", "add"} {
		t.Run(command, func(t *testing.T) {
			fixture := newNativeGlobalFixture(t)
			profile := t.TempDir()
			writeContextPackage(t, profile, "empty-profile", "1.0.0", "profile context\n")
			fixture.command(t, exitOK, "profile", "install", profile, "--use")
			args := []string{"global", "install"}
			if command == "install" {
				fixture.declare(t, "v1.0.0")
			} else {
				args = []string{"global", "add", "shared-skill", "--git", "https://example.com/shared-skill", "--tag", "v1.0.0"}
			}
			foreign := filepath.Join(fixture.home, ".curator", "global", "skills", "shared-skill", "SKILL.md")
			writeFile(t, foreign, "foreign preimage\n")
			snapshots := map[string]map[string]helpTreeEntry{}
			for _, path := range []string{filepath.Join(fixture.home, ".claude"), filepath.Join(fixture.home, ".codex"), filepath.Join(fixture.home, ".curator", "profiles", "empty-profile"), filepath.Join(fixture.home, ".curator", "global", "Skillfile.json")} {
				snapshots[path] = nativeOwnerSnapshot(t, path)
			}
			_, stderr := fixture.command(t, exitFail, args...)
			if !strings.Contains(stderr, "source_ownership_unproven") {
				t.Fatalf("canonical refusal lost intended diagnostic: %s", stderr)
			}
			for path, before := range snapshots {
				if after := nativeOwnerSnapshot(t, path); !reflect.DeepEqual(before, after) {
					t.Errorf("canonical refusal changed profile/native state: %s: %s", path, nativeOwnerDiff(before, after))
				}
			}
			fixture.requireAbsent(t)
			if data, err := os.ReadFile(foreign); err != nil || string(data) != "foreign preimage\n" {
				t.Fatalf("foreign canonical entry changed: %q %v", data, err)
			}
			if err := os.RemoveAll(filepath.Dir(foreign)); err != nil {
				t.Fatal(err)
			}
			fixture.command(t, exitOK, args...)
			fixture.requireProfileOwner(t)
		})
	}
}

// An actual legacy journal fault after a committed target restores the prior
// profile lock, native/managed surfaces and backup inventory. A fault-free retry
// publishes both phases. Synthetic credentials remain inside the fixture.
func TestGlobalNativeOwnerCommitRollback(t *testing.T) {
	fixture := newGlobalLockCLIFixture(t)
	requireGlobalCommand(t, fixture, exitOK, "profile", "sync")
	fixture.writeLegacyGlobalDeclaration(t)
	snapshots := map[string]map[string]helpTreeEntry{}
	for _, path := range []string{filepath.Join(fixture.home, "profiles"), envprofile.EnvRoot(fixture.home), filepath.Dir(fixture.codex)} {
		snapshots[path] = nativeOwnerSnapshot(t, path)
	}
	fired := false
	opts := install.Options{Commit: install.CommitDeps{Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointTargetCommitted && !fired {
			fired = true
			return fmt.Errorf("planted legacy commit failure")
		}
		return nil
	}}}}
	opts.WithGlobalProfile = func(lock install.HomeLock, next func() install.Result) (install.Result, error) {
		var result install.Result
		_, err := envprofile.GlobalSkillTransaction(fixture.home, envprofile.PolicyFromConfig(fixture.source.cfg), envregistry.MachineConfig{}, nil, nil, lock, func() error {
			result = next()
			if result.Status == "failed" {
				return fmt.Errorf("%s", strings.Join(result.Errors, "; "))
			}
			return nil
		})
		if err != nil {
			result.Status = "failed"
			result.Errors = append(result.Errors, err.Error())
		}
		return result, err
	}
	result := install.Global(fixture.source.cfg, filepath.Dir(fixture.codex), opts)
	if !fired || result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "planted legacy commit failure") {
		t.Fatalf("intended commit fault not exercised: fired=%v %+v", fired, result)
	}
	for path, before := range snapshots {
		if after := nativeOwnerSnapshot(t, path); !reflect.DeepEqual(before, after) {
			t.Errorf("commit refusal changed profile/native state: %s: %s", path, nativeOwnerDiff(before, after))
		}
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, "global", "skills", "new-skill")); !os.IsNotExist(err) {
		t.Fatalf("failed commit left canonical skill: %v", err)
	}
	opts.Commit.Hooks = transaction.Hooks{}
	result = install.Global(fixture.source.cfg, filepath.Dir(fixture.codex), opts)
	if result.Status != "ok" {
		t.Fatalf("fault-free retry refused: %+v", result)
	}
	fixture.requireNativeSkill(t)
}

func nativeOwnerSnapshot(t *testing.T, root string) map[string]helpTreeEntry {
	t.Helper()
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		t.Fatal(err)
	}
	snapshot := snapshotHelpTree(t, root)
	for rel, entry := range snapshot {
		// Directory allocation size is not a content or permission invariant.
		if entry.Mode.IsDir() {
			entry.Size = 0
		}
		if entry.Mode&os.ModeSymlink != 0 {
			target, err := os.Readlink(filepath.Join(root, rel))
			if err != nil {
				t.Fatal(err)
			}
			entry.Hash = sha256.Sum256([]byte(target))
		}
		snapshot[rel] = entry
	}
	return snapshot
}

func nativeOwnerDiff(before, after map[string]helpTreeEntry) string {
	var diffs []string
	for path, entry := range before {
		if got, ok := after[path]; !ok || got != entry {
			diffs = append(diffs, fmt.Sprintf("%s: %+v -> %+v", path, entry, got))
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			diffs = append(diffs, "added "+path)
		}
	}
	return strings.Join(diffs, "; ")
}

// A first-use global add conflict retains a usable pending profile even when
// its Skillfile declaration was never published; explicit takeover can recover
// the native skill from that lock, and retry then commits the declaration.
func TestGlobalNativeOwnerFreshAddConflict(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	foreign := filepath.Join(fixture.home, ".codex", "skills", "shared-skill", "SKILL.md")
	writeFile(t, foreign, "foreign preimage\n")
	args := []string{"global", "add", "shared-skill", "--git", "https://example.com/shared-skill", "--tag", "v1.0.0"}
	_, stderr := fixture.command(t, exitFail, args...)
	if !strings.Contains(stderr, "environment_surface_unmanaged_conflict") {
		t.Fatalf("fresh add conflict diagnostic: %s", stderr)
	}
	lock, hash, err := contextlock.Read(filepath.Join(fixture.home, ".curator", "profiles", "default", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := lock.Find(contextlock.KindSkill, "shared-skill"); !ok {
		t.Fatal("fresh add lost its pending skill lock")
	}
	code, stdout, stderr := runFile(t, fixture.config, "profile", "sync", "--takeover")
	if code != exitOK {
		t.Fatalf("explicit takeover refused: %d %s%s", code, stdout, stderr)
	}
	fixture.requireProfileOwner(t)
	_, after, err := contextlock.Read(filepath.Join(fixture.home, ".curator", "profiles", "default", "lock.json"))
	if err != nil || after != hash {
		t.Fatalf("takeover replaced the pending global lock: %s -> %s (%v)", hash, after, err)
	}
	fixture.command(t, exitOK, args...)
	fixture.requireProfileOwner(t)
}

func configureNativeCompiledSkill(t *testing.T, fixture nativeGlobalFixture, compiled string) {
	t.Helper()
	f, err := os.OpenFile(os.Getenv("GIT_CONFIG_GLOBAL"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("\n[url \"" + filepath.ToSlash(compiled) + "\"]\n\tinsteadOf = https://example.com/build-skill\n")
	if closeErr := f.Close(); err != nil || closeErr != nil {
		t.Fatalf("write fixture mapping: %v %v", err, closeErr)
	}
	writeFile(t, filepath.Join(fixture.home, ".curator", "global", "Skillfile.json"), `{"schema_version":1,"agents":["codex_cli","claude_code"],"skills":[{"name":"build-skill","git":"https://example.com/build-skill","tag":"v1"}]}`)
}

func requireCompiledNativeOwner(t *testing.T, fixture nativeGlobalFixture) {
	t.Helper()
	for _, path := range []string{".curator/global/skills/build-skill", ".codex/skills/build-skill", ".claude/skills/build-skill"} {
		if _, err := os.Stat(filepath.Join(fixture.home, filepath.FromSlash(path))); err != nil {
			t.Fatalf("real build did not publish %s: %v", path, err)
		}
	}
}

// Covered inventory hosts perform a real global compiled install. Uncovered
// hosts return the specified control diagnostic and preserve native/profile
// state without publishing a canonical entry. No host is silently skipped.
func TestGlobalNativeOwnerBuildInventory(t *testing.T) {
	fixture := newNativeGlobalFixture(t)
	profile := t.TempDir()
	writeContextPackage(t, profile, "empty-profile", "1.0.0", "profile context\n")
	fixture.command(t, exitOK, "profile", "install", profile, "--use")
	compiled := t.TempDir()
	writeCompiledSkillRepo(t, compiled)
	configureNativeCompiledSkill(t, fixture, compiled)
	snapshots := map[string]map[string]helpTreeEntry{}
	for _, path := range []string{filepath.Join(fixture.home, ".claude"), filepath.Join(fixture.home, ".codex"), filepath.Join(fixture.home, ".curator", "profiles", "empty-profile")} {
		snapshots[path] = nativeOwnerSnapshot(t, path)
	}
	code, stdout, stderr := runFile(t, fixture.config, "global", "install")
	if godriver.InventoryPlatform(runtime.GOOS) != "" {
		if code != exitOK {
			t.Fatalf("covered global compiled install failed: %d %s%s", code, stdout, stderr)
		}
		requireCompiledNativeOwner(t, fixture)
		return
	}
	if code != exitFail || !strings.Contains(stderr, godriver.CodeControlUnavailable) || !strings.Contains(stderr, "no record for host "+runtime.GOOS) {
		t.Fatalf("uncovered global build lost its specific refusal: %d %s%s", code, stdout, stderr)
	}
	for path, before := range snapshots {
		if after := nativeOwnerSnapshot(t, path); !reflect.DeepEqual(before, after) {
			t.Fatalf("uncovered global build published profile/native state: %s", nativeOwnerDiff(before, after))
		}
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, ".curator", "global", "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("uncovered global build published canonical entry: %v", err)
	}
}

// Directory allocation length is opaque filesystem metadata. Snapshots omit
// it but preserve file length, bytes and entry kind; identical trees compare
// equal, while same-length edits, length changes and file-to-directory changes
// are detected. Existing ownership tests retain mode and link-target checks.
func TestNativeOwnerSnapshotContract(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "entry")
	writeFile(t, path, "before")
	before := nativeOwnerSnapshot(t, root)
	if before["."].Size != 0 {
		t.Fatal("directory allocation leaked into snapshot")
	}
	if before["entry"].Size != 6 {
		t.Fatal("regular file size lost")
	}
	if !reflect.DeepEqual(before, nativeOwnerSnapshot(t, root)) {
		t.Fatal("unchanged snapshot differs")
	}
	writeFile(t, path, "mutant")
	if reflect.DeepEqual(before, nativeOwnerSnapshot(t, root)) {
		t.Fatal("same-length file edit ignored")
	}
	writeFile(t, path, "longer content")
	after := nativeOwnerSnapshot(t, root)
	if after["entry"].Size != 14 || reflect.DeepEqual(before, after) {
		t.Fatal("file length change ignored")
	}
	writeFile(t, path, "before")
	if !reflect.DeepEqual(before, nativeOwnerSnapshot(t, root)) {
		t.Fatal("restored content differs after directory allocation changes")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(before, nativeOwnerSnapshot(t, root)) {
		t.Fatal("file replacement by directory ignored")
	}
}
