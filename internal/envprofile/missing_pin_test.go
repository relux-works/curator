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

// The real same-name Install path refreshes a mutable root, while Update keeps
// its accepted immutable snapshot. Only synthetic repositories/homes are used.
// Each rejection checks its specific cause and preserves the accepted target,
// unrelated cache, current selection and native surfaces; acquired objects are
// allowed to remain outside that accepted closure after a later audit failure.
func TestReinstallMissingDependencyPin(t *testing.T) {
	for _, mode := range []string{"warm", "cached", "ssh-origin", "ssh-fetch-failure", "ssh-forbidden-rewrite", "prior-audit-failure", "audit-after-fetch", "still-absent", "probe-failure", "corrupt", "non-commit", "malformed", "forbidden-source", "denied"} {
		t.Run(mode, func(t *testing.T) {
			f := newPinTransportFixture(t)
			declaration, err := os.ReadFile(filepath.Join(f.root, "agent-context.json"))
			if err != nil {
				t.Fatal(err)
			}
			writeManifestPackage(t, f.root, `{"schema_version":1,"name":"review-transport-root","version":"1.0.0"}`, nil)
			original, _, _, err := Install(f.home, InstallOptions{Operand: f.root})
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := original.Lock.Find(contextlock.KindContext, "dep"); ok {
				t.Fatal("fixture dependency is already in the accepted closure")
			}
			beforeTarget := hashTreeForTest(t, ProfileDir(f.home, original.Name))
			manager := newGitManager(f.home)
			if mode == "prior-audit-failure" {
				// A prior refused install leaves this warm cache outside the
				// accepted closure. Its cached pin needs no acquisition.
				oldPin := fixtureGitOutput(t, f.cache, "rev-parse", "HEAD")
				prior := strings.Replace(string(declaration), f.pin, oldPin, 1)
				writeManifestPackage(t, f.root, prior, nil)
				previous := canaryPasses
				canaryPasses = func() bool { return false }
				_, _, _, priorErr := Install(f.home, InstallOptions{Operand: f.root})
				canaryPasses = previous
				if priorErr == nil || !strings.Contains(priorErr.Error(), "audit canary failed") {
					t.Fatalf("prior audit refusal: %v", priorErr)
				}
				if hashTreeForTest(t, ProfileDir(f.home, original.Name)) != beforeTarget {
					t.Fatal("prior audit refusal changed accepted target")
				}
			}
			if mode == "cached" {
				if err := manager.fetch(f.source); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(f.cache, ".git", "FETCH_HEAD"), []byte("keep fetch metadata\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				assertPinTransportMissing(t, f.cache, f.pin)
			}
			wantOrigin := fixtureGitOutput(t, f.cache, "config", "--get", "remote.origin.url")
			if mode == "ssh-origin" || mode == "ssh-fetch-failure" || mode == "ssh-forbidden-rewrite" {
				// Same canonical source, different transport spelling: the
				// cache keeps its SSH origin despite the HTTPS declaration.
				wantOrigin = "git@example.com:review-transport"
				f.selectRemote(t, wantOrigin, false)
				if mode != "ssh-forbidden-rewrite" {
					f.configureHelper(t, mode == "ssh-fetch-failure")
					t.Setenv("GIT_SSH_COMMAND", "'"+strings.ReplaceAll(filepath.ToSlash(f.helper), "'", "'\\''")+"'")
					t.Setenv("GIT_SSH_VARIANT", "ssh")
				} else {
					gitRun(t, f.cache, "config", "--file", f.ids.path, "url.fd::99.insteadOf", wantOrigin)
					gitRun(t, f.cache, "config", "protocol.fd.allow", "always")
					t.Setenv("GIT_ALLOW_PROTOCOL", gitops.AllowedProtocols+":fd")
				}
			}
			pin := f.pin
			if mode == "malformed" {
				pin = "not-a-pin"
			} else if mode == "non-commit" {
				pin = fixtureGitOutput(t, f.cache, "rev-parse", "HEAD:context/a.md")
			}
			if mode == "corrupt" {
				object := filepath.Join(f.cache, ".git", "objects", pin[:2], pin[2:])
				if err := os.MkdirAll(filepath.Dir(object), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(object, []byte("corrupt fixture object\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			freshDeclaration := strings.Replace(string(declaration), f.pin, pin, 1)
			if mode == "forbidden-source" {
				freshDeclaration = strings.Replace(freshDeclaration, f.operand, "file:///forbidden-source", 1)
			}
			writeManifestPackage(t, f.root, freshDeclaration, nil)
			trace := filepath.Join(t.TempDir(), "reinstall-git-trace")
			t.Setenv("GIT_TRACE", trace)
			updated, moved, err := UpdateWithPolicy(f.home, original.Name, Policy{})
			if err != nil || moved || updated.LockHash != original.LockHash {
				t.Fatalf("update adopted mutable declaration: moved=%v error=%v", moved, err)
			}
			updateTrace, err := os.ReadFile(trace)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if strings.Contains(string(updateTrace), "built-in: git fetch ") || strings.Contains(string(updateTrace), "built-in: git clone ") {
				t.Fatal("immutable update acquired newly declared source")
			}
			if err := os.WriteFile(trace, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			if mode == "still-absent" || mode == "probe-failure" {
				installMissingPinGit(t, mode, trace)
			}
			policy := Policy{}
			if mode == "denied" {
				policy.AllowedSources = []string{"example.com/other"}
			}
			if mode == "audit-after-fetch" {
				previous := canaryPasses
				canaryPasses = func() bool { return false }
				t.Cleanup(func() { canaryPasses = previous })
			}
			refs := fixtureGitOutput(t, f.cache, "show-ref")
			info, activated, isUpdated, installErr := Install(f.home, InstallOptions{Operand: f.root, Policy: policy})
			raw, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			wantFetches := 1
			if mode == "cached" || mode == "malformed" || mode == "non-commit" || mode == "probe-failure" || mode == "corrupt" || mode == "forbidden-source" || mode == "denied" {
				wantFetches = 0
			}
			// Count first: a bounded retry plant must fail for duplication,
			// rather than for the observer's emergency second-attempt exit.
			if n := strings.Count(string(raw), "built-in: git fetch "); n != wantFetches {
				t.Fatalf("exact acquisition count %d, want %d; install error: %v", n, wantFetches, installErr)
			}
			if strings.Contains(string(raw), "built-in: git clone ") {
				t.Fatal("warm reinstall cloned an unrelated or replacement source")
			}
			if wantFetches == 1 && mode != "still-absent" {
				want := "built-in: git fetch --no-tags --no-write-fetch-head --no-auto-maintenance --recurse-submodules=no --refmap= origin " + pin
				if !strings.Contains(string(raw), want) {
					t.Fatal("acquisition did not use the exact no-ref/no-submodule command")
				}
			}
			success := mode == "warm" || mode == "cached" || mode == "ssh-origin" || mode == "prior-audit-failure"
			if success {
				if installErr != nil || activated || !isUpdated {
					t.Fatalf("same-name reinstall: activated=%v updated=%v error=%v", activated, isUpdated, installErr)
				}
				member, ok := info.Lock.Find(contextlock.KindContext, "dep")
				if !ok || member.Commit != pin || member.Source != f.source || info.LockHash == original.LockHash {
					t.Fatalf("wrong acquired closure member: %+v", member)
				}
				payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(f.home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
				if err != nil || string(payload) != pinTransportBytes {
					t.Fatalf("pinned bytes = %q, error %v", payload, err)
				}
			} else {
				reason := map[string]string{
					"ssh-fetch-failure": "fetch exact pin", "forbidden-source": "file:// git sources carry no network identity and are not accepted", "corrupt": "inspect pinned object",
					"ssh-forbidden-rewrite": "transport 'fd' not allowed", "audit-after-fetch": "audit canary failed",
					"still-absent": "remains unavailable after acquisition", "probe-failure": "inspect pinned object",
					"non-commit": "not an available commit", "malformed": "revision is not a full lowercase commit", "denied": "outside the machine's allowed sources",
				}[mode]
				if installErr == nil || !strings.Contains(installErr.Error(), reason) {
					t.Fatalf("expected %q refusal, got %v", reason, installErr)
				}
				if hashTreeForTest(t, ProfileDir(f.home, original.Name)) != beforeTarget {
					t.Fatal("refused reinstall changed accepted target/source/lock")
				}
				if mode == "audit-after-fetch" && fixtureGitOutput(t, f.cache, "cat-file", "-t", f.pin) != "commit" {
					t.Fatal("audit refusal was not reached after successful acquisition")
				}
			}
			if fixtureGitOutput(t, f.cache, "show-ref") != refs || fixtureGitOutput(t, f.cache, "config", "--get", "remote.origin.url") != wantOrigin {
				t.Fatal("acquisition changed selected refs or rewrote origin")
			}
			if mode == "ssh-origin" {
				marker, err := os.ReadFile(f.marker)
				if err != nil || !strings.Contains(string(marker), "git-upload-pack") || !strings.Contains(string(marker), "example.com") {
					t.Fatalf("SSH transport not exercised: %q %v", marker, err)
				}
			}
			if mode == "ssh-forbidden-rewrite" || mode == "still-absent" || mode == "probe-failure" {
				// The database-failure plant deliberately prevents a readback;
				// the other refusals must still leave the exact pin absent.
				if mode != "probe-failure" {
					assertPinTransportMissing(t, f.cache, f.pin)
				}
			}
			f.assertPreserved(t, true)
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
	manager := newGitManager(gitFixtureHome(t))
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
	manager := newGitManager(gitFixtureHome(t))
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
