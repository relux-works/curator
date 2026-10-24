package envprofile

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/privatedir"
)

// The reviewer regression keeps its original name. Both origin substitution
// and operator rewrites must obey the same protocol policy as gitops.Fetch.
// ext uses a portable executable that either fails or serves real Git objects;
// fd must fail at the policy check, not at file-descriptor/helper setup.
func TestReviewMissingPinPreservesProtocolAllowlist(t *testing.T) {
	for _, mode := range []string{"fails", "serves", "serves-rewrite", "fails-rewrite", "fd", "fd-rewrite"} {
		t.Run(mode, func(t *testing.T) {
			f := newPinTransportFixture(t)
			protocol := "ext"
			remote := "ext::" + strings.ReplaceAll(filepath.ToSlash(f.helper), " ", "% ")
			if strings.HasPrefix(mode, "fd") {
				protocol, remote = "fd", "fd::99"
			}
			f.configureHelper(t, strings.HasPrefix(mode, "fails"))
			gitRun(t, f.cache, "config", "protocol."+protocol+".allow", "always")
			// Also prove an inherited permissive protocol environment cannot
			// override the production restriction.
			t.Setenv("GIT_ALLOW_PROTOCOL", gitops.AllowedProtocols+":ext:fd")
			f.selectRemote(t, remote, strings.HasSuffix(mode, "-rewrite"))
			beforeCache := hashTreeForTest(t, f.cache)
			if err := gitops.Fetch(f.cache); err == nil || !strings.Contains(err.Error(), "transport '"+protocol+"' not allowed") {
				t.Fatalf("reference refusal missing: %v", err)
			}
			_, _, _, installErr := Install(f.home, InstallOptions{Operand: f.root})
			if _, err := os.Stat(f.marker); !os.IsNotExist(err) {
				t.Fatalf("missing-pin Install executed forbidden transport; install error: %v", installErr)
			}
			if installErr == nil || !strings.Contains(installErr.Error(), DiagSourceInvalid) ||
				!strings.Contains(installErr.Error(), "transport '"+protocol+"' not allowed") {
				t.Fatalf("wrong protocol refusal: %v", installErr)
			}
			if hashTreeForTest(t, f.cache) != beforeCache {
				t.Fatal("denied acquisition changed selected cache objects, refs or metadata")
			}
			assertPinTransportMissing(t, f.cache, f.pin)
			f.assertPreserved(t, false)
		})
	}
}

// Actual Git HTTPS transport talks to a loopback TLS server with an explicitly
// trusted fixture certificate. SSH dispatches through Git's SSH transport into
// a controlled upload-pack executable. It tests protocol admission and operator
// configuration, not SSH encryption/authentication or external connectivity.
func TestInstallMissingPinAllowedTransports(t *testing.T) {
	for _, transport := range []string{"https", "ssh"} {
		for _, rewrite := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/rewrite-%t", transport, rewrite), func(t *testing.T) {
				f := newPinTransportFixture(t)
				var requests atomic.Int32
				remote := "ssh://fixture.invalid/repository"
				if transport == "https" {
					gitRun(t, f.repo, "update-server-info")
					handler := http.FileServer(http.Dir(filepath.Join(f.repo, ".git")))
					server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						handler.ServeHTTP(w, r)
					}))
					defer server.Close()
					cert := filepath.Join(t.TempDir(), "ca.pem")
					if err := os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0o600); err != nil {
						t.Fatal(err)
					}
					t.Setenv("GIT_SSL_CAINFO", cert)
					t.Setenv("GIT_SSL_NO_VERIFY", "false")
					t.Setenv("NO_PROXY", "*")
					// Git for Windows may select Schannel; explicitly let it
					// consume this fixture CA instead of the machine trust store.
					gitRun(t, f.cache, "config", "http.schannelUseSSLCAInfo", "true")
					remote = server.URL
				} else {
					// Git parses this as a shell command on all native platforms.
					t.Setenv("GIT_SSH_COMMAND", "'"+strings.ReplaceAll(filepath.ToSlash(f.helper), "'", "'\\''")+"'")
					t.Setenv("GIT_SSH_VARIANT", "ssh")
				}
				f.selectRemote(t, remote, rewrite)
				refs := fixtureGitOutput(t, f.cache, "show-ref")
				info, _, _, err := Install(f.home, InstallOptions{Operand: f.root})
				if err != nil {
					t.Fatal(err)
				}
				member, ok := info.Lock.Find(contextlock.KindContext, "dep")
				if !ok || member.Commit != f.pin || member.Source != f.source {
					t.Fatalf("pin/canonical identity changed: %+v", member)
				}
				payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(f.home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
				if err != nil || string(payload) != pinTransportBytes {
					t.Fatalf("wrong pinned bytes: %q %v", payload, err)
				}
				if transport == "https" && requests.Load() == 0 {
					t.Fatal("HTTPS transport was not exercised")
				}
				if transport == "ssh" {
					raw, err := os.ReadFile(f.marker)
					if err != nil || !strings.Contains(string(raw), "git-upload-pack") || !strings.Contains(string(raw), "fixture.invalid") {
						t.Fatalf("SSH transport was not exercised: %q %v", raw, err)
					}
				}
				if fixtureGitOutput(t, f.cache, "show-ref") != refs {
					t.Fatal("selected refs changed")
				}
				f.assertPreserved(t, true)
			})
		}
	}
}

const pinTransportBytes = "exact transport fixture bytes\r\nwith trailing spaces  \n"

type pinTransportFixture struct {
	home, root, repo, cache, source, operand, pin, helper, marker string
	ids                                                           *gitIdentities
	preserved                                                     map[string]string
	current                                                       string
}

func newPinTransportFixture(t *testing.T) *pinTransportFixture {
	t.Helper()
	native := pinHomes(t)
	pinOperatorHome(t, t.TempDir())
	f := &pinTransportFixture{home: gitFixtureHome(t), root: t.TempDir(), preserved: map[string]string{}}
	f.ids = newGitIdentities(t)
	existing := t.TempDir()
	writePackage(t, existing, "existing", "1.0.0", "existing bytes\n")
	if _, _, _, err := Install(f.home, InstallOptions{Operand: existing}); err != nil {
		t.Fatal(err)
	}
	var err error
	f.current, err = Current(f.home)
	if err != nil {
		t.Fatal(err)
	}
	// The manager-owned environment boundary must be private on Windows too;
	// native homes and unrelated caches below retain their original fixtures.
	if err := privatedir.MakeAll(EnvRoot(f.home)); err != nil {
		t.Fatal(err)
	}
	paths := []string{ProfileDir(f.home, "existing"), ProfileDir(f.home, DefaultProfile), EnvRoot(f.home)}
	for _, path := range native {
		paths = append(paths, path)
	}
	f.repo = gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`,
		"context/a.md":       "old cached HEAD bytes\n",
	}, "v1.0.0")
	f.operand = f.ids.serve(f.repo, "https://example.com/review-transport")
	manager := newGitManager(f.home)
	f.source, err = manager.Identity("context", "dep", f.operand)
	if err != nil {
		t.Fatal(err)
	}
	f.cache, err = manager.ensureRepo(f.source)
	if err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, f.repo, "context/a.md", pinTransportBytes)
	gitRun(t, f.repo, "add", "--all")
	gitRun(t, f.repo, "commit", "-m", "required pin")
	f.pin = fixtureGitOutput(t, f.repo, "rev-parse", "HEAD")
	assertPinTransportMissing(t, f.cache, f.pin)
	writeManifestPackage(t, f.root, fmt.Sprintf(`{"schema_version":1,"name":"review-transport-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, f.operand, f.pin), nil)
	unrelated := gitRepo(t, map[string]string{"sentinel": "unrelated"}, "v1.0.0")
	unrelatedURL := f.ids.serve(unrelated, "https://example.com/unrelated-transport")
	unrelatedSource, err := manager.Identity("context", "unrelated", unrelatedURL)
	if err != nil {
		t.Fatal(err)
	}
	unrelatedCache, err := manager.ensureRepo(unrelatedSource)
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, unrelatedCache)
	for _, path := range paths {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
		f.preserved[path] = hashTreeForTest(t, path)
	}
	if err := os.WriteFile(filepath.Join(f.cache, ".git", "FETCH_HEAD"), []byte("keep fetch metadata\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	f.marker = filepath.Join(dir, "transport-ran")
	f.helper = filepath.Join(dir, "pin-transport")
	if runtime.GOOS == "windows" {
		f.helper += ".exe"
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.helper, data, 0o700); err != nil {
		t.Fatal(err)
	}
	f.configureHelper(t, false)
	return f
}

func assertPinTransportMissing(t *testing.T, cache, pin string) {
	t.Helper()
	cmd := exec.Command("git", "-C", cache, "cat-file", "--batch-check")
	cmd.Stdin = strings.NewReader(pin + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != pin+" missing\n" {
		t.Fatalf("fixture pin must be absent: %q %v", out, err)
	}
}

func (f *pinTransportFixture) selectRemote(t *testing.T, remote string, rewrite bool) {
	t.Helper()
	// Remove the initial file rewrite; positives must use HTTPS/SSH themselves.
	if err := os.WriteFile(f.ids.path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if rewrite {
		gitRun(t, f.cache, "config", "--file", f.ids.path, "url."+remote+".insteadOf", f.operand)
		if got := fixtureGitOutput(t, f.cache, "config", "--get", "remote.origin.url"); got != f.operand {
			t.Fatalf("canonical origin changed: %q", got)
		}
	} else {
		gitRun(t, f.cache, "remote", "set-url", "origin", remote)
	}
}

func (f *pinTransportFixture) assertPreserved(t *testing.T, success bool) {
	t.Helper()
	for path, before := range f.preserved {
		if hashTreeForTest(t, path) != before {
			t.Fatalf("existing fixture state changed: %s", filepath.Base(path))
		}
	}
	current, err := Current(f.home)
	if err != nil || current != f.current {
		t.Fatalf("current changed: %q %v", current, err)
	}
	raw, err := os.ReadFile(filepath.Join(f.cache, ".git", "FETCH_HEAD"))
	if err != nil || string(raw) != "keep fetch metadata\n" {
		t.Fatalf("FETCH_HEAD changed: %q %v", raw, err)
	}
	if !success {
		for _, path := range []string{ProfileDir(f.home, "review-transport-root"), filepath.Join(EnvRoot(f.home), "review-transport-root")} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("target profile/source/lock/home published: %v", err)
			}
		}
	}
}

type pinTransportConfig struct {
	Repo, Marker string
	Fail         bool
}

func (f *pinTransportFixture) configureHelper(t *testing.T, fail bool) {
	t.Helper()
	data, err := json.Marshal(pinTransportConfig{Repo: f.repo, Marker: f.marker, Fail: fail})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.helper+".json", data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// Runs before TestMain isolates config. This is an actual Git ext/SSH helper,
// not a replacement git command; admission is decided by the real Git binary.
func runPinTransport() int {
	self, err := os.Executable()
	if err != nil {
		return 98
	}
	raw, err := os.ReadFile(self + ".json")
	if err != nil {
		return 98
	}
	var config pinTransportConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return 98
	}
	if err := os.WriteFile(config.Marker, []byte(strings.Join(os.Args, "\n")), 0o600); err != nil {
		return 98
	}
	if config.Fail {
		return 1
	}
	cmd := exec.Command("git", "upload-pack", config.Repo)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return 1
	}
	return 0
}
