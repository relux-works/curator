package envprofile

import (
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/gitops"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
)

// Install must carry the exact dependency declaration across packageOf and
// Identity into clone and missing-pin fetch. Only that endpoint is served;
// transport-independent cache/lock identity and independently specified bytes
// must survive both transport orderings. Dispatch does not prove SSH crypto.
func TestInstallDeclaredDependencyTransport(t *testing.T) {
	endpoints := []string{"https://fixture.invalid/Team/dep", "ssh://git@fixture.invalid/Team/dep.git", "git@fixture.invalid:Team/dep.git"}
	for _, endpoint := range endpoints {
		for _, stale := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stale-%t", strings.Split(endpoint, ":")[0], stale), func(t *testing.T) {
				pinHomes(t)
				pinOperatorHome(t, t.TempDir())
				ids := newGitIdentities(t)
				repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`, "context/a.md": "old\n"}, "v1.0.0")
				ids.serve(repo, endpoint)
				ids.configureDispatch(func(f *declaredFixture) { f.EnforceHeadless = true })
				home := t.TempDir()
				m := newGitManager(home)
				source, err := m.Identity("context", "dep", endpoint)
				if err != nil {
					t.Fatal(err)
				}
				if stale {
					previous := endpoints[0]
					if endpoint == previous {
						previous = endpoints[1]
					}
					ids.serve(repo, previous)
					old := newGitManager(home)
					_, _ = old.Identity("context", "dep", previous)
					if _, err := old.ensureRepo(source); err != nil {
						t.Fatal(err)
					}
				}
				const expected = "independent exact bytes\r\ntrailing  \n"
				writeGitFile(t, repo, "context/a.md", expected)
				gitRun(t, repo, "add", "--all")
				gitRun(t, repo, "commit", "-m", "exact pin")
				pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
				root := t.TempDir()
				writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"transport-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, endpoint, pin), nil)
				_ = os.Remove(ids.dispatch + ".traffic")
				info, _, _, err := Install(home, InstallOptions{Operand: root})
				if err != nil {
					t.Fatal(err)
				}
				member, ok := info.Lock.Find(contextlock.KindContext, "dep")
				if !ok || member.Source != "fixture.invalid/Team/dep" || member.Commit != pin {
					t.Fatalf("wrong lock identity/pin: %+v", member)
				}
				bytes, err := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
				if err != nil || string(bytes) != expected {
					t.Fatalf("wrong exact bytes %q: %v", bytes, err)
				}
				traffic, err := os.ReadFile(ids.dispatch + ".traffic")
				if err != nil {
					t.Fatal(err)
				}
				lines := strings.Split(strings.TrimSpace(string(traffic)), "\n")
				if len(lines) != 1 || !strings.Contains(lines[0], "\t"+endpoint+"\t") {
					t.Fatalf("declaration dropped, swapped or retried: %s", traffic)
				}
				entries, err := os.ReadDir(m.reposDir())
				if err != nil || len(entries) != 1 {
					t.Fatalf("cache split: %v %v", entries, err)
				}
			})
		}
	}
}

// Public Install refuses auth/transport/missing-ref failures without fallback,
// leaking stderr, or publishing profile/lock/current. The positive siblings
// above use the same dispatcher and exact object verification.
func TestInstallDeclaredEndpointRefusal(t *testing.T) {
	for _, failure := range []struct{ name, text, class string }{
		{"auth", "fatal: Authentication failed SECRET-SENTINEL", "authentication_failed"},
		{"unreachable", "connection refused SECRET-SENTINEL", "transport_failed"},
		{"missing-ref", "fatal: not our ref SECRET-SENTINEL", "ref_unavailable"},
	} {
		for _, stale := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stale-%t", failure.name, stale), func(t *testing.T) {
				pinHomes(t)
				pinOperatorHome(t, t.TempDir())
				ids := newGitIdentities(t)
				repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0"}`}, "v1.0.0")
				endpoint := "ssh://git@fixture.invalid/Team/dep.git"
				ids.serve(repo, endpoint)
				home := t.TempDir()
				m := newGitManager(home)
				source, err := m.Identity("context", "dep", endpoint)
				if err != nil {
					t.Fatal(err)
				}
				if stale {
					if _, err := m.ensureRepo(source); err != nil {
						t.Fatal(err)
					}
				}
				gitRun(t, repo, "commit", "--allow-empty", "-m", "missing pin")
				pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
				root := t.TempDir()
				writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"refused-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, endpoint, pin), nil)
				ids.configureDispatch(func(f *declaredFixture) {
					f.EnforceHeadless = true
					f.Refusals = map[string]string{endpoint: failure.text}
				})
				_ = os.Remove(ids.dispatch + ".traffic")
				_, _, _, err = Install(home, InstallOptions{Operand: root})
				var refusal *gitops.AcquisitionError
				if !errors.As(err, &refusal) || refusal.Class != failure.class || refusal.Identity != source || refusal.Endpoint != "ssh://fixture.invalid/Team/dep.git" {
					t.Fatalf("wrong structured refusal: %v", err)
				}
				if strings.Contains(err.Error(), "SECRET-SENTINEL") {
					t.Fatal("credential-bearing stderr leaked")
				}
				traffic, readErr := os.ReadFile(ids.dispatch + ".traffic")
				if readErr != nil {
					t.Fatal(readErr)
				}
				lines := strings.Split(strings.TrimSpace(string(traffic)), "\n")
				if len(lines) != 1 || !strings.Contains(lines[0], "\t"+endpoint+"\t") {
					t.Fatalf("alternate/repeated acquisition: %s", traffic)
				}
				if _, err := os.Stat(ProfileDir(home, "refused-root")); !os.IsNotExist(err) {
					t.Fatal("refused profile published")
				}
				if current, err := Current(home); err != nil || current != "" {
					t.Fatalf("refusal published current: %q %v", current, err)
				}
			})
		}
	}
}

// Install applies malformed-source and allowlist denials before network even
// when the legitimate canonical cache is warm. Valid siblings are above.
func TestInstallDeclaredSourceAdmission(t *testing.T) {
	for _, mode := range []string{"denied", "port", "ext", "fd", "file"} {
		t.Run(mode, func(t *testing.T) {
			pinHomes(t)
			pinOperatorHome(t, t.TempDir())
			ids := newGitIdentities(t)
			repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0"}`}, "v1.0.0")
			endpoint := "ssh://git@fixture.invalid/Team/dep.git"
			ids.serve(repo, endpoint)
			home := t.TempDir()
			m := newGitManager(home)
			source, err := m.Identity("context", "dep", endpoint)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := m.ensureRepo(source); err != nil {
				t.Fatal(err)
			}
			pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
			policy := Policy{}
			switch mode {
			case "denied":
				policy.AllowedSources = []string{"fixture.invalid/Other"}
			case "port":
				endpoint = "ssh://git@fixture.invalid:2222/Team/dep"
			case "ext":
				endpoint = "ext::forbidden"
			case "fd":
				endpoint = "fd::99"
			case "file":
				endpoint = gitFileURL(repo)
			}
			root := t.TempDir()
			writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"denied-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, endpoint, pin), nil)
			before := hashTreeForTest(t, m.repoDir(source))
			_ = os.Remove(ids.dispatch + ".traffic")
			_, _, _, err = Install(home, InstallOptions{Operand: root, Policy: policy})
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
				t.Fatalf("source denial missing: %v", err)
			}
			if traffic, err := os.ReadFile(ids.dispatch + ".traffic"); !os.IsNotExist(err) {
				t.Fatalf("denied source attempted acquisition: %s %v", traffic, err)
			}
			if hashTreeForTest(t, m.repoDir(source)) != before {
				t.Fatal("denial changed warm cache")
			}
			if _, err := os.Stat(ProfileDir(home, "denied-root")); !os.IsNotExist(err) {
				t.Fatal("denied profile published")
			}
		})
	}
}

// Real Git SSH dispatch reaches the declared host on Clone and Fetch, with
// inherited core.sshCommand/insteadOf ignored. The helper observes headless
// options; it does not implement encryption or host-key authentication.
func TestSharedGitDeclaredSSH(t *testing.T) {
	f := newPinTransportFixture(t)
	sshConfig := filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	if err := os.MkdirAll(filepath.Dir(sshConfig), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sshConfig, []byte("Host example.com\n Hostname forbidden.invalid\n ProxyCommand forbidden-command\n"), 0600); err != nil {
		t.Fatal(err)
	}
	installPinSSH(t, f)
	endpoint := "ssh://git@example.com/review-transport"
	f.ids.configureDispatch(func(c *declaredFixture) { c.Direct = append(c.Direct, endpoint) })
	gitRun(t, f.cache, "config", "--file", f.ids.path, "core.sshCommand", "forbidden-command")
	gitRun(t, f.cache, "config", "--file", f.ids.path, "url.https://forbidden.invalid/.insteadOf", endpoint)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := gitops.Clone(endpoint, dest); err != nil {
		t.Fatal(err)
	}
	if err := gitops.Fetch(dest); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(f.marker)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"git@example.com", "BatchMode=yes", "StrictHostKeyChecking=yes", "-F"} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("missing actual SSH option/endpoint %q: %s", want, raw)
		}
	}
	if got := fixtureGitOutput(t, dest, "rev-parse", "HEAD"); got != f.pin {
		t.Fatalf("wrong exact SSH object: %s", got)
	}
	f.configureHelper(t, true)
	installPinSSH(t, f)
	_ = os.Remove(f.ids.dispatch + ".traffic")
	err = gitops.Fetch(dest)
	var refusal *gitops.AcquisitionError
	if !errors.As(err, &refusal) || refusal.Class != "transport_failed" || refusal.Endpoint != "ssh://example.com/review-transport" {
		t.Fatalf("wrong failed SSH refusal: %v", err)
	}
	traffic, err := os.ReadFile(f.ids.dispatch + ".traffic")
	if err != nil || strings.Count(string(traffic), "\n") != 1 || strings.Contains(string(traffic), "https://") {
		t.Fatalf("SSH failure retried/fell back: %s %v", traffic, err)
	}

}

// Install of a Git root refreshes a warm cache through gitManager.fetch;
// that public error must retain the same typed sanitized refusal as a missing
// dependency pin, without changing cached state or publishing a profile.
func TestInstallDeclaredRootFetchRefusal(t *testing.T) {
	pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"root","version":"1.0.0"}`}, "v1.0.0")
	endpoint := "ssh://git@fixture.invalid/Team/root.git"
	ids.serve(repo, endpoint)
	home := t.TempDir()
	m := newGitManager(home)
	source, err := m.Identity("context", "root", endpoint)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := m.ensureRepo(source)
	if err != nil {
		t.Fatal(err)
	}
	before := hashTreeForTest(t, cache)
	ids.configureDispatch(func(f *declaredFixture) {
		f.EnforceHeadless = true
		f.Refusals = map[string]string{endpoint: "fatal: Authentication failed SECRET-SENTINEL"}
	})
	_, _, _, err = Install(home, InstallOptions{Operand: endpoint})
	var refusal *gitops.AcquisitionError
	if !errors.As(err, &refusal) || refusal.Class != "authentication_failed" || refusal.Identity != source || strings.Contains(err.Error(), "SECRET-SENTINEL") {
		t.Fatalf("root fetch lost typed sanitized refusal: %v", err)
	}
	if hashTreeForTest(t, cache) != before {
		t.Fatal("failed refresh changed cache")
	}
	if _, err := os.Stat(ProfileDir(home, "root")); !os.IsNotExist(err) {
		t.Fatal("failed root published")
	}
}

// Reinstall carries the newly declared root endpoint through updateLocked;
// a cache first acquired over HTTPS cannot restore that endpoint after SSH
// was explicitly requested. Lock identity stays transport-independent.
func TestInstallReinstallDeclaredRootEndpoint(t *testing.T) {
	pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"root","version":"1.0.0"}`}, "v1.0.0")
	https, ssh := "https://fixture.invalid/Team/root", "ssh://git@fixture.invalid/Team/root"
	ids.serve(repo, https)
	ids.serve(repo, ssh)
	home := t.TempDir()
	first, _, _, err := Install(home, InstallOptions{Operand: https})
	if err != nil {
		t.Fatal(err)
	}
	ids.serve(filepath.Join(t.TempDir(), "unreachable"), https)
	_ = os.Remove(ids.dispatch + ".traffic")
	next, _, updated, err := Install(home, InstallOptions{Operand: ssh})
	if err != nil {
		t.Fatal(err)
	}
	if !updated || first.LockHash != next.LockHash {
		t.Fatal("reinstall changed portable lock identity")
	}
	traffic, err := os.ReadFile(ids.dispatch + ".traffic")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(traffic)), "\n") {
		if !strings.Contains(line, "\t"+ssh+"\t") {
			t.Fatalf("reinstall lost current endpoint: %s", traffic)
		}
	}
}

// Standalone Update has no new URL declaration: its canonical portable lock
// must not cause generation of HTTPS in place of the existing SSH origin.
func TestUpdateRetainsDeclaredCachedEndpoint(t *testing.T) {
	pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"root","version":"1.0.0"}`}, "v1.0.0")
	endpoint := "ssh://git@fixture.invalid/Team/root"
	ids.serve(repo, endpoint)
	home := t.TempDir()
	first, _, _, err := Install(home, InstallOptions{Operand: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	next, _, err := UpdateWithPolicy(home, first.Name, Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if first.LockHash != next.LockHash {
		t.Fatal("unchanged cached source changed lock identity")
	}
}

// Update must read a dependency's current manifest declaration before refresh;
// the canonical old lock and an HTTPS-created cache cannot replace its SSH URL.
func TestUpdateDeclaredDependencyEndpoint(t *testing.T) {
	pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	ids := newGitIdentities(t)
	manifest := func(version string) string {
		return fmt.Sprintf(`{"schema_version":1,"name":"dep","version":%q,"context":{"modules":[{"path":"a.md"}]}}`, version)
	}
	repo := gitRepo(t, map[string]string{"agent-context.json": manifest("1.0.0"), "context/a.md": "old"}, "v1.0.0")
	https, ssh := "https://fixture.invalid/Team/dep", "ssh://git@fixture.invalid/Team/dep"
	ids.serve(repo, https)
	ids.serve(repo, ssh)
	home := t.TempDir()
	manager := newGitManager(home)
	source, err := manager.Identity("context", "dep", https)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ensureRepo(source); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"range":"*"}}}}`, ssh), nil)
	if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, repo, "agent-context.json", manifest("1.1.0"))
	writeGitFile(t, repo, "context/a.md", "new exact bytes\r\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "new version")
	gitRun(t, repo, "tag", "v1.1.0")
	pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	ids.serve(filepath.Join(t.TempDir(), "absent"), https)
	_ = os.Remove(ids.dispatch + ".traffic")
	info, _, err := UpdateWithPolicy(home, "root", Policy{})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := info.Lock.Find(contextlock.KindContext, "dep")
	if !ok || member.Commit != pin || member.Source != source {
		t.Fatalf("wrong updated identity/pin: %+v", member)
	}
	raw, err := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
	if err != nil || string(raw) != "new exact bytes\r\n" {
		t.Fatalf("wrong updated bytes: %q %v", raw, err)
	}
	traffic, err := os.ReadFile(ids.dispatch + ".traffic")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(traffic)), "\n")
	if len(lines) != 1 || !strings.Contains(lines[0], "\t"+ssh+"\t") {
		t.Fatalf("update lost declaration or repeated refresh: %s", traffic)
	}
}

// Reinstall reads the current local declaration and refreshes an existing
// dependency even for an exact revision. The resolver visits Candidates before
// selecting the fixed revision, so the same on-demand refresh applies. It must
// acquire that pin through SSH rather than the old HTTPS cache origin.
func TestReinstallDeclaredPinnedDependencyEndpoint(t *testing.T) {
	pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`, "context/a.md": "old"}, "v1.0.0")
	https, ssh := "https://fixture.invalid/Team/dep", "ssh://git@fixture.invalid/Team/dep"
	ids.serve(repo, https)
	ids.serve(repo, ssh)
	home, root := t.TempDir(), t.TempDir()
	declaration := func(endpoint, pin string) {
		writeManifestPackage(t, root, fmt.Sprintf(`{"schema_version":1,"name":"root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, endpoint, pin), nil)
	}
	declaration(https, fixtureGitOutput(t, repo, "rev-parse", "HEAD"))
	if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, repo, "context/a.md", "new pinned bytes\r\n")
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", "new exact revision")
	pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	declaration(ssh, pin)
	ids.serve(filepath.Join(t.TempDir(), "absent"), https)
	_ = os.Remove(ids.dispatch + ".traffic")
	info, _, _, err := Install(home, InstallOptions{Operand: root})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := info.Lock.Find(contextlock.KindContext, "dep")
	if !ok || member.Commit != pin || member.Source != "fixture.invalid/Team/dep" {
		t.Fatalf("wrong pinned member: %+v", member)
	}
	raw, err := os.ReadFile(filepath.Join(contextstore.EntryDir(home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
	if err != nil || string(raw) != "new pinned bytes\r\n" {
		t.Fatalf("wrong pinned bytes: %q %v", raw, err)
	}
	traffic, err := os.ReadFile(ids.dispatch + ".traffic")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(traffic)), "\n")
	if len(lines) != 1 || !strings.Contains(lines[0], "\t"+ssh+"\t") {
		t.Fatalf("wrong pinned refresh: %s", traffic)
	}
}
