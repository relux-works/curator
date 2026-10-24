package envprofile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/privatedir"
)

// Production entry point: Install -> installLocked -> contextresolve.Resolve.
// All repositories, native homes and manager state are synthetic.
func TestInstallMissingDependencyPin(t *testing.T) {
	for _, mode := range []string{"fresh", "stale", "cached-offline", "missing-offline", "fetch-failure", "still-absent", "denied", "malformed", "invalid-content", "probe-failure", "non-commit"} {
		t.Run(mode, func(t *testing.T) {
			native := pinHomes(t)
			pinOperatorHome(t, t.TempDir())
			ids := newGitIdentities(t)
			home := gitFixtureHome(t)
			existing := t.TempDir()
			writePackage(t, existing, "existing", "1.0.0", "existing bytes\n")
			if _, _, _, err := Install(home, InstallOptions{Operand: existing}); err != nil {
				t.Fatal(err)
			}
			beforeProfile := hashTreeForTest(t, ProfileDir(home, "existing"))
			beforeDefault := hashTreeForTest(t, ProfileDir(home, DefaultProfile))
			if err := privatedir.MakeAll(EnvRoot(home)); err != nil {
				t.Fatal(err)
			}
			beforeManaged := hashTreeForTest(t, EnvRoot(home))
			beforeNative := map[string]string{}
			for id, dir := range native {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				beforeNative[id] = hashTreeForTest(t, dir)
			}
			beforeCurrent, err := Current(home)
			if err != nil {
				t.Fatal(err)
			}
			repo := gitRepo(t, map[string]string{
				"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`,
				"context/a.md":       "old bytes\n",
			}, "v1.0.0")
			operand := ids.serve(repo, "https://example.com/missing-pin-dep")
			manager := newGitManager(home)
			source, err := manager.Identity("context", "dep", operand)
			if err != nil {
				t.Fatal(err)
			}
			if mode != "fresh" {
				if _, err := manager.ensureRepo(source); err != nil {
					t.Fatal(err)
				}
			}
			const expected = "independently specified new pinned bytes\n"
			writeGitFile(t, repo, "context/a.md", expected)
			if mode == "invalid-content" {
				writeGitFile(t, repo, "agent-context.json", "{invalid\n")
			}
			gitRun(t, repo, "add", "--all")
			gitRun(t, repo, "commit", "-m", "required pin")
			pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
			if mode == "cached-offline" {
				if err := manager.fetch(source); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "missing-offline" || mode == "cached-offline" {
				gitRun(t, manager.repoDir(source), "remote", "set-url", "origin", gitFileURL(filepath.Join(t.TempDir(), "absent")))
			}
			if mode == "fetch-failure" {
				pin = strings.Repeat("a", 40)
			}
			if mode == "malformed" {
				pin = "not-a-pin"
			}
			if mode == "non-commit" {
				pin = fixtureGitOutput(t, manager.repoDir(source), "rev-parse", "HEAD:context/a.md")
			}
			root := t.TempDir()
			writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"new-profile","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, operand, pin), nil)
			trace := filepath.Join(t.TempDir(), "git-trace")
			if mode == "still-absent" || mode == "probe-failure" {
				installMissingPinGit(t, mode, trace)
			}
			// An unrelated repository, including all refs/objects, is outside
			// the selected cache acquisition surface.
			unrelated := gitRepo(t, map[string]string{"sentinel": "untouched"}, "v1.0.0")
			unrelatedURL := ids.serve(unrelated, "https://example.com/unrelated")
			unrelatedSource, err := manager.Identity("context", "unrelated", unrelatedURL)
			if err != nil {
				t.Fatal(err)
			}
			unrelated, err = manager.ensureRepo(unrelatedSource)
			if err != nil {
				t.Fatal(err)
			}
			unrelatedBefore := hashTreeForTest(t, unrelated)
			refsBefore := ""
			fetchHead := ""
			if mode != "fresh" {
				refsBefore = fixtureGitOutput(t, manager.repoDir(source), "show-ref")
				fetchHead = filepath.Join(manager.repoDir(source), ".git", "FETCH_HEAD")
				if err := os.WriteFile(fetchHead, []byte("pre-existing fetch metadata\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			policy := Policy{}
			if mode == "denied" {
				policy.AllowedSources = []string{"example.com/other"}
			}
			t.Setenv("GIT_TRACE", trace)
			info, _, _, installErr := Install(home, InstallOptions{Operand: root, Policy: policy})
			success := mode == "fresh" || mode == "stale" || mode == "cached-offline"
			if success {
				if installErr != nil {
					t.Fatal(installErr)
				}
				member, ok := info.Lock.Find(contextlock.KindContext, "dep")
				if !ok || member.Commit != pin {
					t.Fatalf("wrong exact pin: %+v", member)
				}
				payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
				if err != nil || string(payload) != expected {
					t.Fatalf("pinned bytes = %q, error %v", payload, err)
				}
			} else {
				if installErr == nil {
					t.Fatal("invalid acquisition admitted")
				}
				if !strings.Contains(installErr.Error(), DiagSourceInvalid) {
					t.Fatalf("wrong refusal: %v", installErr)
				}
				if _, err := os.Stat(ProfileDir(home, "new-profile")); !os.IsNotExist(err) {
					t.Fatalf("target profile published: %v", err)
				}
				if _, err := os.Stat(filepath.Join(EnvRoot(home), "new-profile")); !os.IsNotExist(err) {
					t.Fatalf("target managed home published: %v", err)
				}
			}
			if got := hashTreeForTest(t, ProfileDir(home, "existing")); got != beforeProfile {
				t.Fatal("existing profile changed")
			}
			if hashTreeForTest(t, ProfileDir(home, DefaultProfile)) != beforeDefault {
				t.Fatal("default profile changed")
			}
			if hashTreeForTest(t, EnvRoot(home)) != beforeManaged {
				t.Fatal("managed homes changed")
			}
			for id, dir := range native {
				if hashTreeForTest(t, dir) != beforeNative[id] {
					t.Fatalf("native home %s changed", id)
				}
			}
			if hashTreeForTest(t, unrelated) != unrelatedBefore {
				t.Fatal("unrelated repository changed")
			}
			if mode != "fresh" && fixtureGitOutput(t, manager.repoDir(source), "show-ref") != refsBefore {
				t.Fatal("selected cache refs changed")
			}
			if fetchHead != "" {
				raw, err := os.ReadFile(fetchHead)
				if err != nil || string(raw) != "pre-existing fetch metadata\n" {
					t.Fatalf("FETCH_HEAD changed: %q %v", raw, err)
				}
			}
			afterCurrent, err := Current(home)
			if err != nil || afterCurrent != beforeCurrent {
				t.Fatalf("current changed: %q %v", afterCurrent, err)
			}
			raw, err := os.ReadFile(trace)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			fetches := strings.Count(string(raw), "built-in: git fetch ")
			clones := strings.Count(string(raw), "built-in: git clone ")
			if mode == "fresh" {
				if clones != 1 {
					t.Fatalf("fresh clone count %d, want 1", clones)
				}
			} else if clones != 0 {
				t.Fatalf("unexpected clone count %d", clones)
			}
			if mode == "cached-offline" || mode == "denied" || mode == "malformed" || mode == "fresh" || mode == "probe-failure" || mode == "non-commit" {
				if fetches != 0 {
					t.Fatalf("unexpected acquisition (%d): %s", fetches, raw)
				}
			} else if fetches != 1 {
				t.Fatalf("fetch count = %d, want 1: %s", fetches, raw)
			}
		})
	}
}

// A portable subprocess fault plant exercises the production Install path:
// successful transport with no object, or a database read failure. Other Git
// commands run the real executable, with the original isolated fixture config.
func installMissingPinGit(t *testing.T, mode, trace string) {
	t.Helper()
	gitBinary, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o700); err != nil {
		t.Fatal(err)
	}
	config, err := json.Marshal(pinGitConfig{Binary: gitBinary, Mode: mode, Trace: trace})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pin-git.json"), config, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

type pinGitConfig struct{ Binary, Mode, Trace string }

func runMissingPinGit() int {
	self, err := os.Executable()
	if err != nil {
		return 98
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(self), "pin-git.json"))
	if err != nil {
		return 98
	}
	var config pinGitConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 98
	}
	args := os.Args[1:]
	mode := config.Mode
	// Isolated Git deliberately discards ambient tracing. The executable
	// observer's sidecar survives that environment boundary.
	if os.Getenv("GIT_TRACE") == "" && mode == "observe" {
		f, err := os.OpenFile(config.Trace, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return 98
		}
		command := args
		for len(command) >= 2 && (command[0] == "-C" || command[0] == "-c") {
			command = command[2:]
		}
		_, err = fmt.Fprintln(f, "observer: built-in: git "+strings.Join(command, " "))
		if closeErr := f.Close(); err != nil || closeErr != nil {
			return 98
		}
	}
	if len(args) >= 3 && args[2] == "fetch" && mode == "still-absent" {
		prior, _ := os.ReadFile(config.Trace)
		f, err := os.OpenFile(config.Trace, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return 98
		}
		_, err = fmt.Fprintln(f, "fault plant: built-in: git fetch successful without object")
		if closeErr := f.Close(); err != nil || closeErr != nil {
			return 98
		}
		// Bound the known-violation control even if the retry guard mutates.
		if strings.Contains(string(prior), "fault plant: built-in: git fetch") {
			return 96
		}
		return 0
	}
	if len(args) >= 3 && args[2] == "cat-file" && mode == "probe-failure" {
		return 97
	}
	cmd := exec.Command(config.Binary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		return 99
	}
	return 0
}

func TestMissingPinManagerAttemptBoundAndCacheOnlyDefault(t *testing.T) {
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0"}`}, "v1.0.0")
	url := ids.serve(repo, "https://example.com/bounded-pin")
	manager := newGitManager(t.TempDir())
	source, err := manager.Identity("context", "dep", url)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ensureRepo(source); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, repo, "new", "new object")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "new")
	pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)
	// A plain manager remains cache-only, even with a live advertised pin.
	if _, err := manager.Manifest("context", "dep", source, "", pin); err == nil {
		t.Fatal("plain manager fetched missing pin")
	}
	raw, _ := os.ReadFile(trace)
	if strings.Contains(string(raw), "built-in: git fetch ") {
		t.Fatal("plain manager acquired pin")
	}
	manager.missingPins = make(map[string]bool)
	installMissingPinGit(t, "still-absent", trace)
	for range 3 {
		if _, err := manager.Manifest("context", "dep", source, "", pin); err == nil || !strings.Contains(err.Error(), "remains unavailable") {
			t.Fatalf("missing pin: %v", err)
		}
	}
	raw, err = os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(raw), "built-in: git fetch "); n != 1 {
		t.Fatalf("duplicate acquisition count %d, want 1", n)
	}
}

func TestMissingPinResolveNeverFetches(t *testing.T) {
	for _, repair := range []bool{false, true} {
		t.Run(fmt.Sprintf("repair-%t", repair), func(t *testing.T) {
			fx := writeManagedFixture(t, "pinned")
			repo := newGitManager(fx.home).repoDir("github.com/example/mcp-figma")
			remote := filepath.Join(t.TempDir(), "advertised.git")
			gitRun(t, repo, "clone", "--bare", repo, remote)
			gitRun(t, repo, "remote", "add", "origin", gitFileURL(remote))
			objects := filepath.Join(repo, ".git", "objects")
			if err := os.RemoveAll(objects); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(objects, 0o755); err != nil {
				t.Fatal(err)
			}
			before := hashTreeForTest(t, fx.home)
			trace := filepath.Join(t.TempDir(), "trace")
			t.Setenv("GIT_TRACE", trace)
			installMissingPinGit(t, "observe", trace)
			req := fx.request("codex_cli")
			req.Repair = repair
			result, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), "pinned commit object unavailable") {
				t.Fatalf("expected local-pin refusal, got %v", err)
			}
			if result != nil && len(result.Document) != 0 {
				t.Fatal("missing pin emitted fragment")
			}
			raw, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "built-in: git fetch ") || strings.Contains(string(raw), "built-in: git clone ") {
				t.Fatalf("resolve acquired objects: %s", raw)
			}
			if hashTreeForTest(t, fx.home) != before {
				t.Fatal("resolve changed managed state")
			}
		})
	}
}

func TestMissingPinManifestRejectsMalformedBeforeAcquisition(t *testing.T) {
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"sentinel": "valid source"}, "v1.0.0")
	url := ids.serve(repo, "https://example.com/malformed-pin")
	manager := newGitManager(t.TempDir())
	source, err := manager.Identity("context", "dep", url)
	if err != nil {
		t.Fatal(err)
	}
	manager.missingPins = make(map[string]bool)
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)
	if _, err := manager.Manifest("context", "dep", source, "", "not-a-pin"); err == nil || !strings.Contains(err.Error(), "invalid exact commit") {
		t.Fatalf("malformed pin: %v", err)
	}
	raw, err := os.ReadFile(trace)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(raw) != 0 {
		t.Fatalf("malformed pin invoked Git: %s", raw)
	}
}

func TestMissingPinObserverDetectsIsolatedFetch(t *testing.T) {
	repo := gitRepo(t, map[string]string{"sentinel": "advertised bytes"}, "v1.0.0")
	clone := filepath.Join(t.TempDir(), "clone")
	gitRun(t, repo, "clone", repo, clone)
	pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	trace := filepath.Join(t.TempDir(), "trace")
	installMissingPinGit(t, "observe", trace)
	// Known-violation control: this is precisely the network operation the
	// Resolve refusal rows must observe even across the isolated environment.
	if err := gitops.FetchCommitIsolated(clone, pin); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(raw), "built-in: git fetch "); n != 1 {
		t.Fatalf("observer missed isolated fetch (%d): %s", n, raw)
	}
}
