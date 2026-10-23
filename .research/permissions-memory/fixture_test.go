// Package permissionsmemory measures exact native builds; it grants no adapter admission.
package permissionsmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}
func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	return string(b) + "\n"
}

type probe struct {
	t                                        *testing.T
	root, home, work, binary, native, policy string
	env                                      []string
	server                                   *httptest.Server
	mu                                       sync.Mutex
	requests                                 []map[string]any
	codexReply                               http.HandlerFunc
	nativeInput                              string
	webHits                                  int
	managerArgs                              []string
	finalText                                string
	prompt                                   string
	issued                                   bool
}

func setupBinary(t *testing.T, key string, pins map[string]bool) *probe {
	t.Helper()
	binary := os.Getenv(key)
	if binary == "" {
		t.Skip("opt-in: " + key + " unset; no vendor process or fixture")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("pinned probe requires darwin/arm64")
	}
	b, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if !pins[hex.EncodeToString(sum[:])] && os.Getenv("QUAL_MUTANT") != "digest" {
		t.Fatal("unqualified binary digest")
	}
	base := os.Getenv("QUAL_RUN_ROOT")
	if !filepath.IsAbs(base) || !strings.Contains(base, string(filepath.Separator)+".temp"+string(filepath.Separator)) {
		t.Fatal("explicit absolute .temp run root required")
	}
	p := &probe{t: t, binary: binary, root: filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))}
	if _, err := os.Stat(p.root); !os.IsNotExist(err) {
		t.Fatal("run root exists; choose a fresh evidence directory")
	}
	p.native, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	p.home = filepath.Join(p.root, "home", ".native")
	p.work = filepath.Join(p.root, "ancestor", "work")
	for _, d := range []string{p.home, p.work, filepath.Join(p.root, "tmp"), filepath.Join(p.root, "config"), filepath.Join(p.root, "data"), filepath.Join(p.root, "state"), filepath.Join(p.root, "cache")} {
		if err := os.MkdirAll(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	p.server = httptest.NewServer(http.HandlerFunc(p.respond))
	t.Cleanup(p.server.Close)
	p.env = []string{"HOME=" + filepath.Dir(p.home), "PATH=/usr/bin:/bin:/usr/sbin:/sbin", "TMPDIR=" + filepath.Join(p.root, "tmp"), "XDG_CONFIG_HOME=" + filepath.Join(p.root, "config"), "XDG_DATA_HOME=" + filepath.Join(p.root, "data"), "XDG_STATE_HOME=" + filepath.Join(p.root, "state"), "XDG_CACHE_HOME=" + filepath.Join(p.root, "cache"), "NO_PROXY=127.0.0.1", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "LANG=en_US.UTF-8", "TERM=dumb"}
	// Only the exact loopback provider port is reachable. All writes stay in the
	// fixture; native directories and security services remain denied throughout.
	port := strings.TrimPrefix(p.server.URL, "http://127.0.0.1:")
	sb := "(version 1)\n(allow default)\n(deny network*)\n(allow network-outbound (remote ip \"localhost:" + port + "\"))\n(deny file-write*)\n(allow file-write* (subpath " + strconv.Quote(p.root) + "))\n(allow file-write* (literal \"/dev/null\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n"
	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".curator", ".ssh", ".muse", ".agents", "Library"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(p.native, rel)) + "))\n"
	}
	for _, dir := range []string{"/etc/codex", "/etc/muse", "/Library/Application Support/Codex", "/Library/Application Support/Muse", "/Library/Managed Preferences"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(dir) + "))\n"
	}
	// Existing checkout/ancestor instructions must not contaminate the fixture.
	for dir := filepath.Dir(p.root); ; dir = filepath.Dir(dir) {
		for _, rel := range []string{"CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", ".claude"} {
			sb += "(deny file-read* (subpath " + strconv.Quote(filepath.Join(dir, rel)) + "))\n"
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}
	p.policy = filepath.Join(p.root, "probe.sb")
	write(t, p.policy, sb)
	// A separate metadata directory gives the fixture a repository boundary
	// without creating a .git directory or inheriting the caller's GIT_DIR.
	cmd := exec.Command("/usr/bin/git", "init", "-q", "--separate-git-dir="+filepath.Join(p.root, "metadata"), p.work)
	cmd.Env = p.env
	cmd.Dir = p.root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture git init: %v %s", err, out)
	}
	return p
}

func (p *probe) respond(w http.ResponseWriter, r *http.Request) {
	if p.codexReply != nil {
		p.codexReply(w, r)
		return
	}
	_, _ = io.WriteString(w, "OFFLINE_OK")
}
func (p *probe) sanitize(s string) string {
	s = strings.ReplaceAll(s, p.root, "$RUN")
	s = strings.ReplaceAll(s, regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(p.root, "-"), "$RUN_ID")
	s = strings.ReplaceAll(s, p.binary, "$BINARY")
	s = strings.ReplaceAll(s, p.native, "$NATIVE_HOME")
	if hostname, err := os.Hostname(); err == nil {
		s = strings.ReplaceAll(s, hostname, "$HOST")
	}
	// Provider keys are fabricated, but no bearer/API header is an evidence field.
	return s
}
func (p *probe) save(name, body string) {
	p.t.Helper()
	write(p.t, filepath.Join(p.root, "evidence", name), p.sanitize(body))
}
func contains(t *testing.T, s, marker string, want bool) {
	t.Helper()
	if strings.Contains(s, marker) != want {
		t.Fatalf("observed %s=%t; want %t", marker, !want, want)
	}
}

// Claim: fixture API and ordinary files work, while unrelated network, synthetic
// Keychains reads and outside-fixture writes are specifically denied unchanged.
func TestOfflineBoundary(t *testing.T) {
	p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
	var mu sync.Mutex
	hits := 0
	unrelated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		hits++
		mu.Unlock()
		_, _ = io.WriteString(w, "UNRELATED_ENDPOINT")
	}))
	defer unrelated.Close()
	policy := p.policy
	if mutant := os.Getenv("QUAL_BOUNDARY_MUTANT"); mutant != "" {
		sb := read(t, p.policy)
		switch mutant {
		case "network":
			sb = strings.Replace(sb, "(deny network*)\n", "", 1)
		case "keychains":
			sb = strings.Replace(sb, "(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n", "", 1)
		default:
			t.Fatal("unknown boundary mutant")
		}
		policy = filepath.Join(p.root, "synthetic-control.sb")
		write(t, policy, sb)
	}
	command := func(name string, args ...string) (string, error) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		wrapped := append([]string{"-f", policy, name}, args...)
		cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", wrapped...)
		cmd.Env = p.env
		cmd.Dir = p.work
		out, err := cmd.CombinedOutput()
		code := 0
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				code = e.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		p.save(fmt.Sprintf("control-%d-%s.log", len(p.requests), filepath.Base(name)+"-"+strconv.Itoa(code)), fmt.Sprintf("command=%s %v\nexit=%d\n%s", name, args, code, out))
		if ctx.Err() != nil {
			t.Fatal("control timed out")
		}
		return string(out), err
	}
	out, err := command("/usr/bin/curl", "--silent", "--show-error", "--max-time", "2", "-H", "Content-Type: application/json", "--data", `{"model":"fixture","messages":[]}`, p.server.URL+"/v1/messages")
	if err != nil || !strings.Contains(out, "OFFLINE_OK") {
		t.Fatalf("allowed provider failed: %v %s", err, out)
	}
	out, err = command("/usr/bin/curl", "--silent", "--show-error", "--max-time", "2", unrelated.URL)
	if err == nil || !strings.Contains(out, "Failed to connect") {
		t.Fatalf("unrelated loopback not denied: %v %s", err, out)
	}
	mu.Lock()
	gotHits := hits
	mu.Unlock()
	if gotHits != 0 {
		t.Fatalf("unrelated network effects=%d", gotHits)
	}
	ordinary := filepath.Join(p.root, "ordinary.txt")
	blocked := filepath.Join(p.root, "Keychains", "synthetic.txt")
	for _, path := range []string{ordinary, blocked} {
		write(t, path, "SYNTHETIC_FILE_CONTROL\n")
	}
	out, err = command("/bin/cat", ordinary)
	if err != nil || out != "SYNTHETIC_FILE_CONTROL\n" {
		t.Fatalf("ordinary read: %v %s", err, out)
	}
	out, err = command("/bin/cat", blocked)
	if err == nil || !strings.Contains(out, "Operation not permitted") {
		t.Fatalf("synthetic Keychains read not denied: %v %s", err, out)
	}
	if read(t, blocked) != "SYNTHETIC_FILE_CONTROL\n" {
		t.Fatal("denied source changed")
	}
	outside := filepath.Join(filepath.Dir(p.root), "outside-write.txt")
	write(t, outside, "ORIGINAL\n")
	out, err = command("/bin/sh", "-c", "printf changed > \"$1\"", "probe", outside)
	if err == nil || !strings.Contains(out, "Operation not permitted") || read(t, outside) != "ORIGINAL\n" {
		t.Fatalf("outside write not rejected unchanged: %v %s", err, out)
	}
	inside := filepath.Join(p.root, "allowed-write.txt")
	out, err = command("/bin/sh", "-c", "printf changed > \"$1\"", "probe", inside)
	if err != nil || read(t, inside) != "changed" {
		t.Fatalf("inside write: %v %s", err, out)
	}
}
