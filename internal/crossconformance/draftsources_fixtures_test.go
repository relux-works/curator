package crossconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testcli"
)

// Shared fixtures for the draft-sources conformance suites. Every helper
// builds inputs; every behavior verdict comes from a production entry
// point (install.Project, closure.ResolveDraft/RefreshDraft, the
// snapshot/audit/transport executors, or the compiled CLI binary).

func requireGit(t *testing.T) string {
	t.Helper()
	return testcli.RequireGit(t)
}

func runDraftGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	testcli.Git(t, dir, args...)
}

func draftGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return testcli.GitOutput(t, dir, args...)
}

// writeDraftSkill materializes one minimal admitted skill package.
func writeDraftSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := map[string]any{"schema_version": 4, "capabilities": map[string]any{}, "commands": map[string]any{}, "dependencies": map[string]any{"skills": map[string]any{}}}
	payload, _ := json.Marshal(spec)
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

// writeManifestDependencySkill materializes a minimal package whose skill
// manifest can carry schema-9 directory requirements.
func writeManifestDependencySkill(t *testing.T, dir, name string, schema int, requirements map[string]map[string]any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test "+name+"\n---\n# "+name+"\n"), 0o644); err != nil {
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

// draftProject creates a git-initialized project with the given skill
// packages and Skillfile payload, plus an isolated manager home.
func draftProject(t *testing.T, payload string, skills map[string]string) (project, home string) {
	t.Helper()
	project = t.TempDir()
	home = t.TempDir()
	for dir, name := range skills {
		writeDraftSkill(t, filepath.Join(project, filepath.FromSlash(dir)), name)
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return project, home
}

func draftManifest(t *testing.T, project, payload string) *manifest.Manifest {
	t.Helper()
	m, err := manifest.ParseBytesWithOptions([]byte(payload), filepath.Join(project, "Skillfile.json"), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatalf("parse draft manifest: %v", err)
	}
	return m
}

// resolveDraftPlan resolves and publishes the lock through the
// production closure entry, returning the plan.
func resolveDraftPlan(t *testing.T, project, home, payload string) *closure.DraftPlan {
	t.Helper()
	m := draftManifest(t, project, payload)
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload)})
	if err != nil {
		t.Fatalf("ResolveDraft: %v", err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	return plan
}

// refreshDraftLocked re-resolves through the production refresh entry
// and publishes the lock, returning the plan.
func refreshDraftLocked(t *testing.T, project, home, payload string) *closure.DraftPlan {
	t.Helper()
	return refreshDraftLockedWithRoots(t, project, home, payload, nil)
}

// refreshDraftLockedWithRoots refreshes with acquired Git trees for
// Git-selected aliases.
func refreshDraftLockedWithRoots(t *testing.T, project, home, payload string, gitRoots map[string]string) *closure.DraftPlan {
	t.Helper()
	m := draftManifest(t, project, payload)
	plan, err := closure.RefreshDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload), Expansion: manifest.ExpansionOptions{GitRoots: gitRoots}}, sourcelock.PathIn(project), install.DraftBindingsPath(home, project))
	if err != nil {
		t.Fatalf("RefreshDraft: %v", err)
	}
	return plan
}

// draftInstallConfig is deliberately a bare Go-API literal: zero clock skew
// (a literal zero, not the 300 s loader default) with the attest stub
// minting created_at at serve time is the strict case that exposed
// BUG-260920-2d9gfv — a refusal class must not depend on the fetch's own
// duration. Keep it strict: do not paper over a future-timestamp flake with
// a larger skew or a fixed created_at in the stub.
func draftInstallConfig(home string) *config.Config {
	return &config.Config{Path: filepath.Join(home, "config.json"), SkillsRoot: filepath.Join(home, "skills-root"), DefaultAgents: []string{"claude_code"}, AdapterMode: "auto"}
}

// draftCLIConfig builds a Go-API config over a bootstrapped CLI home so
// Go-API installs land where the compiled CLI reads them.
func draftCLIConfig(t *testing.T, root, configPath string) *config.Config {
	t.Helper()
	return &config.Config{Path: configPath, SkillsRoot: filepath.Join(root, "skills-root"), DefaultAgents: []string{"codex_cli"}, AdapterMode: "auto"}
}

// draftInstall runs the production project installer on the draft lane.
func draftInstall(t *testing.T, project, home string, opts install.Options) install.Result {
	t.Helper()
	opts.DraftSourcesV1 = true
	if opts.Platform == "" {
		opts.Platform = runtimestore.Platform()
	}
	return install.Project(draftInstallConfig(home), project, "test", opts)
}

func draftPlatform() string { return runtimestore.Platform() }

func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// treeDigest hashes a directory tree (sorted relative paths, symlink
// targets, and file bytes) for preserved-state proofs.
func treeDigest(t *testing.T, root string) string {
	t.Helper()
	return treeDigestFiltered(t, root, nil)
}

// treeDigestFiltered hashes like treeDigest, additionally skipping every
// path for which skip reports true.
func treeDigestFiltered(t *testing.T, root string, skip func(rel string) bool) string {
	t.Helper()
	var entries []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if skip != nil && skip(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Home lock files are operation-private serialization state,
		// not installed state; a refused run holding its locks is not
		// a publication.
		if draftDigestSkipsManagerLocks(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			entries = append(entries, "link "+rel+" -> "+target)
			return nil
		}
		if !info.Mode().IsRegular() {
			// Directories and special files carry no installed state;
			// only regular files and links enter the digest.
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(payload)
		entries = append(entries, fmt.Sprintf("file %s %x %d", rel, sum, len(payload)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:])
}

// draftDigestSkipsManagerLocks reports whether a digest-relative path
// names operation-private manager lock state. The root is joined with
// the native separator so the exclusion holds for the relative
// spelling filepath.Walk produces on every OS: a literal "state/locks"
// prefix never matches the backslash-joined walk on Windows, and the
// fresh lock files of a refused run would count as publication.
func draftDigestSkipsManagerLocks(rel string) bool {
	return draftDigestSkipsManagerLocksSep(rel, string(filepath.Separator))
}

// draftDigestSkipsManagerLocksSep is the separator-explicit form of
// the exclusion, so the Windows spelling is testable on every host.
func draftDigestSkipsManagerLocksSep(rel, sep string) bool {
	root := "state" + sep + "locks"
	return rel == root || strings.HasPrefix(rel, root+sep)
}

// TestDraftDigestSkipsManagerLocksOnEverySeparator locks the
// state/locks exclusion for both the unix and the Windows relative
// spelling. Revision 2 compared against the literal "state/locks"
// prefix, which never matches the backslash-joined walk on Windows,
// so every raw refused-install probe counted fresh lock files (gate
// run 35466449728).
func TestDraftDigestSkipsManagerLocksOnEverySeparator(t *testing.T) {
	for _, sep := range []string{"/", `\`} {
		for _, rel := range []string{
			"state" + sep + "locks",
			"state" + sep + "locks" + sep + "v1" + sep + "project-abc.lock",
		} {
			if !draftDigestSkipsManagerLocksSep(rel, sep) {
				t.Errorf("sep %q: draftDigestSkipsManagerLocksSep(%q) = false, want true", sep, rel)
			}
		}
		for _, rel := range []string{
			"state",
			"state" + sep + "transactions" + sep + "v1" + sep + "x.json",
			"stateful",
			"Skillfile.lock",
			"cache" + sep + "entry",
		} {
			if draftDigestSkipsManagerLocksSep(rel, sep) {
				t.Errorf("sep %q: draftDigestSkipsManagerLocksSep(%q) = true, want false", sep, rel)
			}
		}
	}
	if !draftDigestSkipsManagerLocks(filepath.Join("state", "locks", "v1", "project-abc.lock")) {
		t.Error("native lock path is not skipped")
	}
}

// TestDraftTreeDigestIgnoresManagerLockFiles proves fresh lock files
// under a manager home leave the preserved-state digest unchanged,
// which is what the refused-install probes assert around
// install.Project.
func TestDraftTreeDigestIgnoresManagerLockFiles(t *testing.T) {
	home := t.TempDir()
	locks := filepath.Join(home, "state", "locks", "v1")
	if err := os.MkdirAll(locks, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locks, "a.lock"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := treeDigest(t, home)
	if err := os.WriteFile(filepath.Join(locks, "b.lock"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	if after := treeDigest(t, home); after != before {
		t.Fatal("treeDigest counted manager lock files")
	}
}

// curatorBinary builds the production CLI once per test process and
// returns its absolute path. The build itself lives in testcli: this
// package must not spawn child processes of its own (guard_test.go).
func curatorBinary(t *testing.T) string {
	t.Helper()
	return testcli.Binary(t)
}

// runCurator executes the compiled CLI with an isolated home and the
// draft switch enabled, returning exit code, stdout, and stderr.
func runCurator(t *testing.T, home, configPath string, extraEnv []string, args ...string) (int, string, string) {
	t.Helper()
	bin := curatorBinary(t)
	env := append([]string{
		"HOME=" + home,
		"USERPROFILE=" + home,
		"CURATOR_CONFIG=" + configPath,
		"CURATOR_DRAFT_SOURCES_V1=1",
	}, extraEnv...)
	return testcli.Run(t, "", env, "", bin, args...)
}

// setupCLIProject bootstraps a manager home and registers a project
// through the compiled CLI, returning the config path and project dir.
func setupCLIProject(t *testing.T, root string) (configPath, project, home string) {
	t.Helper()
	home = filepath.Join(root, "cli-home")
	configPath = filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills-root")
	project = filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "bootstrap", "--non-interactive", "--skills-root", skillsRoot); code != 0 {
		t.Fatalf("bootstrap = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "add", "app", project, "--agents", "codex_cli"); code != 0 {
		t.Fatalf("project add = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	runDraftGit(t, project, "init", "-q")
	return configPath, project, home
}
