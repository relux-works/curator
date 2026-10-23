// Package codexexclusion records exact-build discovery observations, not admission authority.
package codexexclusion

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const binarySHA = "74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0"

type fixture struct {
	t                                                 *testing.T
	root, home, codex, work, binary, sandbox, private string
	env                                               []string
}

func put(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) *fixture {
	t.Helper()
	binary := os.Getenv("CODEX_EXCL_BINARY")
	if binary == "" {
		t.Skip("opt-in: CODEX_EXCL_BINARY required; no vendor fixture or process")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("requires qualified darwin/arm64 build")
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != binarySHA {
		t.Fatal("unqualified binary digest; refusing vendor execution")
	}
	base := os.Getenv("CODEX_EXCL_RUN_ROOT")
	if base == "" || !filepath.IsAbs(base) || !strings.Contains(base, "/.temp/") {
		t.Fatal("explicit absolute task-local .temp run root required")
	}
	root := filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("run directory must be fresh")
	}
	f := &fixture{t: t, root: root}
	f.private, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	f.home, f.codex, f.work = filepath.Join(root, "home"), filepath.Join(root, "managed"), filepath.Join(root, "project")
	for _, p := range []string{f.home, f.codex, f.work, filepath.Join(root, "tmp"), filepath.Join(root, "evidence")} {
		if err := os.MkdirAll(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	f.binary = filepath.Join(root, "codex")
	t.Cleanup(func() {
		if err := os.Remove(f.binary); err != nil {
			t.Errorf("remove fixture binary: %v", err)
		}
	})
	put(t, f.binary, string(data))
	if err := os.Chmod(f.binary, 0700); err != nil {
		t.Fatal(err)
	}
	// No inherited auth, socket, proxy, Git, shell-init or session environment.
	f.env = []string{"HOME=" + f.home, "CODEX_HOME=" + f.codex, "XDG_CONFIG_HOME=" + filepath.Join(f.home, "config"), "XDG_CACHE_HOME=" + filepath.Join(f.home, "cache"), "XDG_DATA_HOME=" + filepath.Join(f.home, "data"), "XDG_STATE_HOME=" + filepath.Join(f.home, "state"), "TMPDIR=" + filepath.Join(root, "tmp"), "PATH=/usr/bin:/bin:/usr/sbin:/sbin", "LANG=en_US.UTF-8", "TERM=dumb", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CEILING_DIRECTORIES=" + root}
	f.sandbox = filepath.Join(root, "offline.sb")
	sb := "(version 1)\n(allow default)\n(deny network*)\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n"
	for _, rel := range []string{".codex", ".claude", ".curator", ".config", ".agents", ".local", ".muse", "Library", ".ssh"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(f.private, rel)) + "))\n"
	}
	// Prevent system Codex policy from silently changing a fixture's semantics.
	sb += "(deny file-read* file-write* (subpath \"/etc/codex\") (subpath \"/Library/Application Support/Codex\"))\n"
	if os.Getenv("CODEX_EXCL_MUTANT") == "allow-loopback" {
		sb += "(allow network-outbound (remote ip \"localhost:*\"))\n"
	}
	put(t, f.sandbox, sb)
	f.config("")
	out, code := f.run("identity", "", f.binary, "--version")
	require(t, out, code, "codex-cli 0.162.1")
	f.save("identity.txt", fmt.Sprintf("version=codex-cli 0.162.1\nplatform=darwin/arm64\nsha256=%s\n", binarySHA))
	// Separate metadata: never create .git in a project fixture.
	out, code = f.run("git-init", "", "/usr/bin/git", "init", "-q", "--separate-git-dir", filepath.Join(root, "metadata"), f.work)
	require(t, out, code)
	return f
}

func (f *fixture) config(extra string) {
	put(f.t, filepath.Join(f.codex, "config.toml"), `model = "offline-probe"
model_provider = "offline"
cli_auth_credentials_store = "file"
approval_policy = "never"
sandbox_mode = "read-only"
check_for_update_on_startup = false
`+extra+`
[model_providers.offline]
name = "Offline synthetic provider"
base_url = "http://127.0.0.1:9"
wire_api = "responses"
requires_openai_auth = false
request_max_retries = 0
stream_max_retries = 0
`)
}

func (f *fixture) save(name, data string) {
	data = strings.ReplaceAll(data, f.root, "$RUN")
	data = strings.ReplaceAll(data, f.private, "$PRIVATE_HOME")
	put(f.t, filepath.Join(f.root, "evidence", name), data)
}
func (f *fixture) command(ctx context.Context, binary string, args ...string) *exec.Cmd {
	argv := append([]string{"-f", f.sandbox, "/usr/bin/env", "-i"}, f.env...)
	argv = append(argv, binary)
	argv = append(argv, args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", argv...)
	cmd.Dir = f.work
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	return cmd
}
func (f *fixture) run(name, input, binary string, args ...string) (string, int) {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := f.command(ctx, binary, args...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		} else {
			f.t.Fatal(err)
		}
	}
	f.save(name+".log", fmt.Sprintf("cwd=%s\nargv=%q\npid=%d\nexit=%d\n%s", f.work, append([]string{binary}, args...), cmd.Process.Pid, code, out))
	if ctx.Err() != nil {
		f.t.Fatalf("%s timed out", name)
	}
	return string(out), code
}
func require(t *testing.T, out string, code int, want ...string) {
	t.Helper()
	if code != 0 {
		t.Fatalf("unexpected exit %d: %s", code, out)
	}
	for _, s := range want {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q: %s", s, out)
		}
	}
}
func absent(t *testing.T, out string, markers ...string) {
	t.Helper()
	for _, s := range markers {
		if strings.Contains(out, s) {
			t.Fatalf("excluded marker %q returned", s)
		}
	}
}
func (f *fixture) prompt(name string, args ...string) string {
	out, code := f.run(name, "", f.binary, append(args, "debug", "prompt-input", "EXCLUSION_PROBE")...)
	require(f.t, out, code)
	// A failed or truncated command must never establish absence.
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(out), &items); err != nil || len(items) == 0 {
		f.t.Fatalf("invalid prompt observation: %v", err)
	}
	return out
}
func skill(t *testing.T, path, name, marker string) {
	put(t, filepath.Join(path, name, "SKILL.md"), "---\nname: "+name+"\ndescription: "+marker+"\n---\nSynthetic inert qualification skill.\n")
}

// P1/P2: a zero project document budget suppresses root and nested cwd
// instructions while retaining approved managed-home instructions. Lazy tool
// entry within a running turn is outside this debug-startup observation.
func TestInstructions(t *testing.T) {
	f := setup(t)
	put(t, filepath.Join(f.codex, "AGENTS.md"), "APPROVED_ROOT\n")
	put(t, filepath.Join(f.root, "AGENTS.md"), "OUTSIDE_REPOSITORY_ANCESTOR\n")
	put(t, filepath.Join(f.work, "AGENTS.md"), "PROJECT_ROOT\n")
	put(t, filepath.Join(f.work, "CLAUDE.md"), "PROJECT_CLAUDE\n")
	put(t, filepath.Join(f.work, "nested", "AGENTS.md"), "PROJECT_NESTED\n")
	out := f.prompt("root-baseline")
	require(t, out, 0, "APPROVED_ROOT", "PROJECT_ROOT")
	absent(t, out, "OUTSIDE_REPOSITORY_ANCESTOR", "PROJECT_CLAUDE", "PROJECT_NESTED")
	old := f.work
	f.work = filepath.Join(old, "nested")
	out = f.prompt("nested-baseline")
	require(t, out, 0, "APPROVED_ROOT", "PROJECT_ROOT", "PROJECT_NESTED")
	controls := []string{"-c", "project_doc_max_bytes=0"}
	if os.Getenv("CODEX_EXCL_MUTANT") == "drop-document-cap" {
		controls = nil
	}
	out = f.prompt("nested-excluded", controls...)
	require(t, out, 0, "APPROVED_ROOT")
	absent(t, out, "PROJECT_ROOT", "PROJECT_NESTED", "PROJECT_CLAUDE", "OUTSIDE_REPOSITORY_ANCESTOR")
	f.work = old
	put(t, filepath.Join(f.work, "AGENTS.override.md"), "PROJECT_OVERRIDE\n")
	out = f.prompt("override-baseline")
	require(t, out, 0, "PROJECT_OVERRIDE")
	absent(t, out, "PROJECT_ROOT")
	out = f.prompt("override-excluded", "-c", "project_doc_max_bytes=0")
	require(t, out, 0, "APPROVED_ROOT")
	absent(t, out, "PROJECT_OVERRIDE")
}

// P3: project distrust and the document cap do not exclude repository skills.
// An exact-path disable list excludes the planted files and keeps an approved
// home skill; it is not a wildcard or a complete late-file exclusion control.
func TestSkills(t *testing.T) {
	f := setup(t)
	skill(t, filepath.Join(f.codex, "skills"), "approved", "APPROVED_SKILL")
	skill(t, filepath.Join(f.work, ".agents", "skills"), "repo-agents", "PROJECT_AGENTS_SKILL")
	skill(t, filepath.Join(f.work, ".codex", "skills"), "repo-codex", "PROJECT_CODEX_SKILL")
	f.config(fmt.Sprintf("\n[projects.%q]\ntrust_level = \"untrusted\"\n", f.work))
	out := f.prompt("untrusted-baseline", "-c", "project_doc_max_bytes=0")
	require(t, out, 0, "APPROVED_SKILL", "PROJECT_AGENTS_SKILL")
	disabled := fmt.Sprintf("skills.config=[{path=%q,enabled=false},{path=%q,enabled=false}]", filepath.Join(f.work, ".agents", "skills", "repo-agents", "SKILL.md"), filepath.Join(f.work, ".codex", "skills", "repo-codex", "SKILL.md"))
	controls := []string{"-c", disabled, "-c", "project_doc_max_bytes=0"}
	if os.Getenv("CODEX_EXCL_MUTANT") == "drop-skill-disable" {
		controls = []string{"-c", "project_doc_max_bytes=0"}
	}
	out = f.prompt("known-paths-excluded", controls...)
	require(t, out, 0, "APPROVED_SKILL")
	absent(t, out, "PROJECT_AGENTS_SKILL", "PROJECT_CODEX_SKILL")
	skill(t, filepath.Join(f.work, ".agents", "skills"), "late", "LATE_PROJECT_SKILL")
	out = f.prompt("late-path-bypasses-list", "-c", disabled, "-c", "project_doc_max_bytes=0")
	require(t, out, 0, "APPROVED_SKILL", "LATE_PROJECT_SKILL")
	absent(t, out, "PROJECT_AGENTS_SKILL", "PROJECT_CODEX_SKILL")
}

// The boundary test proves ordinary fixture reads work, Keychains-shaped reads
// are denied with bytes preserved, and a reachable loopback socket is denied.
// It never opens a real native credential file or a real Keychain item.
func TestOfflineBoundary(t *testing.T) {
	f := setup(t)
	public := filepath.Join(f.root, "ordinary")
	secret := filepath.Join(f.root, "Keychains", "synthetic")
	put(t, public, "ordinary-fixture\n")
	put(t, secret, "synthetic-private-fixture\n")
	out, code := f.run("ordinary-readable", "", "/bin/cat", public)
	require(t, out, code, "ordinary-fixture")
	out, code = f.run("keychains-denied", "", "/bin/cat", secret)
	if code != 1 || !strings.Contains(out, "Operation not permitted") {
		t.Fatalf("expected specific file denial, got %d: %s", code, out)
	}
	data, err := os.ReadFile(secret)
	if err != nil || string(data) != "synthetic-private-fixture\n" {
		t.Fatal("fixture bytes changed")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := &httpControl{listener: listener}
	server.start()
	defer server.close()
	// The unsandboxed positive is a single direct loopback request, no proxy.
	c, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(c, "GET / HTTP/1.0\r\n\r\n")
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(c)
	_ = c.Close()
	if err != nil || !bytes.Contains(b, []byte("LOCAL_BOUNDARY")) {
		t.Fatal("loopback positive failed")
	}
	out, code = f.run("loopback-denied", "", "/usr/bin/curl", "--noproxy", "*", "--max-time", "2", "--silent", "--show-error", "http://"+listener.Addr().String())
	if code != 7 {
		t.Fatalf("expected connect denial exit 7, got %d: %s", code, out)
	}
}

type httpControl struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (h *httpControl) start() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for {
			c, e := h.listener.Accept()
			if e != nil {
				return
			}
			_ = c.SetDeadline(time.Now().Add(time.Second))
			_, _ = bufio.NewReader(c).ReadString('\n')
			_, _ = io.WriteString(c, "HTTP/1.0 200 OK\r\nContent-Length: 14\r\n\r\nLOCAL_BOUNDARY")
			_ = c.Close()
		}
	}()
}
func (h *httpControl) close() { _ = h.listener.Close(); h.wg.Wait() }

// app-server is a foreground, fixture-owned child over stdio. No daemon or
// inference request is started, and every process is waited before test exit.
type api struct {
	f      *fixture
	cmd    *exec.Cmd
	cancel context.CancelFunc
	in     io.WriteCloser
	lines  chan []byte
	stderr bytes.Buffer
	id     int
	trace  strings.Builder
}

func (f *fixture) api(name string, args ...string) *api {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	a := &api{f: f, cancel: cancel, lines: make(chan []byte, 100)}
	a.cmd = f.command(ctx, f.binary, append(args, "app-server")...)
	var err error
	a.in, err = a.cmd.StdinPipe()
	if err != nil {
		f.t.Fatal(err)
	}
	stdout, err := a.cmd.StdoutPipe()
	if err != nil {
		f.t.Fatal(err)
	}
	a.cmd.Stderr = &a.stderr
	if err = a.cmd.Start(); err != nil {
		f.t.Fatal(err)
	}
	f.t.Logf("fixture child pid=%d", a.cmd.Process.Pid)
	go func() {
		defer close(a.lines)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 65536), 8<<20)
		for scanner.Scan() {
			a.lines <- append([]byte(nil), scanner.Bytes()...)
		}
	}()
	f.t.Cleanup(func() {
		_ = a.in.Close()
		done := make(chan error, 1)
		go func() { done <- a.cmd.Wait() }()
		var err error
		select {
		case err = <-done:
		case <-time.After(5 * time.Second):
			cancel()
			err = <-done
		}
		cancel()
		f.save(name+".log", a.trace.String()+"\nstderr:\n"+a.stderr.String()+fmt.Sprintf("\nwait=%v\n", err))
		if err != nil {
			f.t.Errorf("app-server did not exit cleanly: %v", err)
		}
	})
	a.call("initialize", map[string]any{"clientInfo": map[string]string{"name": "qualification", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}})
	a.notify("initialized", map[string]any{})
	return a
}
func (a *api) notify(method string, params any) {
	b, _ := json.Marshal(map[string]any{"method": method, "params": params})
	a.trace.WriteString("send " + string(b) + "\n")
	if _, e := fmt.Fprintln(a.in, string(b)); e != nil {
		a.f.t.Fatal(e)
	}
}
func (a *api) call(method string, params any) map[string]any {
	a.f.t.Helper()
	a.id++
	b, _ := json.Marshal(map[string]any{"id": a.id, "method": method, "params": params})
	a.trace.WriteString("send " + string(b) + "\n")
	if _, e := fmt.Fprintln(a.in, string(b)); e != nil {
		a.f.t.Fatal(e)
	}
	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-a.lines:
			if !ok {
				a.f.t.Fatalf("API stopped during %s", method)
			}
			a.trace.WriteString("recv " + string(line) + "\n")
			var msg map[string]any
			if e := json.Unmarshal(line, &msg); e != nil {
				a.f.t.Fatal(e)
			}
			if msg["id"] == float64(a.id) {
				if msg["error"] != nil {
					a.f.t.Fatalf("%s API error: %v", method, msg["error"])
				}
				v, ok := msg["result"].(map[string]any)
				if !ok {
					a.f.t.Fatalf("%s invalid result", method)
				}
				return v
			}
		case <-timer.C:
			a.f.t.Fatalf("API timeout during %s", method)
		}
	}
}
func encoded(v any) string { b, _ := json.Marshal(v); return string(b) }

// P4/P5/P6: trusted config applies project hooks/plugins/MCP/permissions;
// explicit distrust leaves approved home values effective. config/read is a
// loader/configuration observation, not hook execution or MCP tool invocation.
func TestConfigLayers(t *testing.T) {
	for _, trust := range []string{"trusted", "untrusted"} {
		t.Run(trust, func(t *testing.T) {
			f := setup(t)
			effectiveTrust := trust
			if os.Getenv("CODEX_EXCL_MUTANT") == "trust-project" {
				effectiveTrust = "trusted"
			}
			f.config(fmt.Sprintf(`
model_reasoning_effort = "low"
[projects.%q]
trust_level = %q
[mcp_servers.approved]
command = "/usr/bin/true"
[plugins."approved@fixture"]
enabled = true
[[hooks.SessionStart]]
hooks = [{type="command", command="printf APPROVED_HOOK"}]
`, f.work, effectiveTrust))
			// Keep top-level keys before table headers in fixture TOML.
			put(t, filepath.Join(f.work, ".codex", "config.toml"), `model_reasoning_effort = "high"
approval_policy = "on-request"
sandbox_mode = "workspace-write"
[mcp_servers.project]
command = "/usr/bin/false"
[plugins."project@fixture"]
enabled = true
[[hooks.SessionStart]]
hooks = [{type="command", command="printf PROJECT_HOOK"}]
`)
			a := f.api("config-api")
			result := a.call("config/read", map[string]any{"includeLayers": true, "cwd": f.work})
			f.save("config-result.json", encoded(result))
			config, ok := result["config"].(map[string]any)
			if !ok {
				t.Fatal("missing effective config")
			}
			out := encoded(config)
			require(t, out, 0, "approved@fixture", "approved")
			if trust == "trusted" {
				require(t, out, 0, "PROJECT_HOOK", "project@fixture", `"approval_policy":"on-request"`, `"sandbox_mode":"workspace-write"`, `"model_reasoning_effort":"high"`)
			} else {
				absent(t, out, "PROJECT_HOOK", "project@fixture", `"project":`, `"sandbox_mode":"workspace-write"`)
				require(t, out, 0, "APPROVED_HOOK")
				require(t, out, 0, `"approval_policy":"never"`, `"model_reasoning_effort":"low"`)
			}
			hooks := a.call("hooks/list", map[string]any{"cwds": []string{f.work}})
			f.save("hooks-result.json", encoded(hooks))
			require(t, encoded(hooks), 0, "APPROVED_HOOK")
			if trust == "trusted" {
				require(t, encoded(hooks), 0, "PROJECT_HOOK")
			} else {
				absent(t, encoded(hooks), "PROJECT_HOOK")
			}
			started := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
			f.save("permission-thread.json", encoded(started))
			if trust == "trusted" {
				require(t, encoded(started), 0, `"approvalPolicy":"on-request"`, `"type":"workspaceWrite"`)
			} else {
				require(t, encoded(started), 0, `"approvalPolicy":"never"`, `"type":"readOnly"`)
			}
			// Empty project config cannot erase approved managed-home MCP entries.
			put(t, filepath.Join(f.work, ".codex", "config.toml"), "[mcp_servers]\n")
			result = a.call("config/read", map[string]any{"includeLayers": true, "cwd": f.work})
			f.save("empty-mcp-result.json", encoded(result))
			cfg, ok := result["config"].(map[string]any)
			if !ok {
				t.Fatal("missing reloaded config")
			}
			out = encoded(cfg)
			require(t, out, 0, `"approved":`)
			absent(t, out, `"project":`)
		})
	}
}

// P8: config/read re-reads late project settings; distrust still excludes them
// while retaining approved MCP. skills/list observes a late project skill in
// the same server, including with project distrust (no model turn required).
func TestLateDiscovery(t *testing.T) {
	f := setup(t)
	skill(t, filepath.Join(f.codex, "skills"), "approved", "APPROVED_SKILL")
	f.config(fmt.Sprintf("\n[projects.%q]\ntrust_level = \"untrusted\"\n[mcp_servers.approved]\ncommand=\"/usr/bin/true\"\n", f.work))
	a := f.api("late-api")
	result := a.call("skills/list", map[string]any{"cwds": []string{f.work}, "forceReload": true})
	require(t, encoded(result), 0, "APPROVED_SKILL")
	absent(t, encoded(result), "LATE_PROJECT_SKILL")
	skill(t, filepath.Join(f.work, ".agents", "skills"), "late", "LATE_PROJECT_SKILL")
	put(t, filepath.Join(f.work, ".codex", "config.toml"), "[mcp_servers.late_project]\ncommand=\"/usr/bin/false\"\n")
	result = a.call("skills/list", map[string]any{"cwds": []string{f.work}, "forceReload": true})
	require(t, encoded(result), 0, "APPROVED_SKILL", "LATE_PROJECT_SKILL")
	result = a.call("config/read", map[string]any{"cwd": f.work})
	config, ok := result["config"].(map[string]any)
	if !ok {
		t.Fatal("missing effective config")
	}
	require(t, encoded(config), 0, `"approved":`)
	absent(t, encoded(config), `"late_project":`)
}

// P7: actual thread/start and thread/resume retain the managed-home source
// with and without the document cap. The source list omits the project even
// in the uncapped baseline, so it CANNOT establish project-source exclusion.
// No model turn, child agent or previously tainted model history is exercised.
func TestResumeApprovedRoot(t *testing.T) {
	for _, capBytes := range []int{32768, 0} {
		t.Run(fmt.Sprint(capBytes), func(t *testing.T) {
			f := setup(t)
			put(t, filepath.Join(f.codex, "AGENTS.md"), "APPROVED_RESUME_ROOT\n")
			put(t, filepath.Join(f.work, "AGENTS.md"), "PROJECT_RESUME_ROOT\n")
			f.config(fmt.Sprintf("\nproject_doc_max_bytes=%d\n[projects.%q]\ntrust_level=\"untrusted\"\n", capBytes, f.work))
			a := f.api("resume-api")
			started := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": false})
			sources := encoded(started["instructionSources"])
			require(t, sources, 0, filepath.Join(f.codex, "AGENTS.md"))
			// Both baselines omit project files here; do not interpret this as exclusion.
			absent(t, sources, filepath.Join(f.work, "AGENTS.md"))
			thread, ok := started["thread"].(map[string]any)
			if !ok {
				t.Fatal("missing thread")
			}
			id, ok := thread["id"].(string)
			if !ok || id == "" {
				t.Fatal("missing durable id")
			}
			a.call("thread/inject_items", map[string]any{"threadId": id, "items": []any{map[string]any{"type": "message", "role": "user", "content": []any{map[string]string{"type": "input_text", "text": "SYNTHETIC_OFFLINE_HISTORY"}}}}})
			a.call("thread/unsubscribe", map[string]any{"threadId": id})
			put(t, filepath.Join(f.work, "AGENTS.md"), "PROJECT_CHANGED_BEFORE_RESUME\n")
			resumed := a.call("thread/resume", map[string]any{"threadId": id, "cwd": f.work})
			sources = encoded(resumed["instructionSources"])
			require(t, sources, 0, filepath.Join(f.codex, "AGENTS.md"))
			// Both baselines omit project files here; do not interpret this as exclusion.
			absent(t, sources, filepath.Join(f.work, "AGENTS.md"))
			if resumed["approvalPolicy"] != "never" {
				t.Fatal("home approval policy not preserved")
			}
		})
	}
}

// P5: stdio MCP actually initializes and calls an approved tool; a trusted
// project tool also initializes, whereas distrust prevents its process from
// starting. An empty project set retains the approved tool. No model or
// network transport participates in this executable positive/negative pair.
func TestMCPRuntime(t *testing.T) {
	for _, trust := range []string{"trusted", "untrusted"} {
		for _, nonempty := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", trust, nonempty), func(t *testing.T) {
				f := setup(t)
				script, err := os.ReadFile("testdata/mcp.py")
				if err != nil {
					t.Fatal(err)
				}
				fixtureScript := filepath.Join(f.root, "mcp.py")
				put(t, fixtureScript, string(script))
				approvedReceipt := filepath.Join(f.root, "approved-process.json")
				projectReceipt := filepath.Join(f.root, "project-process.json")
				stanza := func(name, receipt string) string {
					return fmt.Sprintf("\n[mcp_servers.%s]\ncommand=\"/usr/bin/python3\"\nargs=[%q,%q,%q]\n", name, fixtureScript, name, receipt)
				}
				f.config(fmt.Sprintf("\n[projects.%q]\ntrust_level=%q\n", f.work, trust) + stanza("approved_tool", approvedReceipt))
				projectConfig := "[mcp_servers]\n"
				if nonempty {
					projectConfig = stanza("project_tool", projectReceipt)
				}
				put(t, filepath.Join(f.work, ".codex", "config.toml"), projectConfig)
				a := f.api("mcp-api")
				start := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
				thread, ok := start["thread"].(map[string]any)
				if !ok {
					t.Fatal("missing thread")
				}
				id, ok := thread["id"].(string)
				if !ok {
					t.Fatal("missing id")
				}
				status := a.call("mcpServerStatus/list", map[string]any{"threadId": id})
				out := encoded(status)
				require(t, out, 0, "approved_tool")
				f.save("mcp-status.json", out)
				called := a.call("mcpServer/tool/call", map[string]any{"threadId": id, "server": "approved_tool", "tool": "approved_tool", "arguments": map[string]any{}})
				require(t, encoded(called), 0, "approved_tool_CALLED")
				if _, err := os.Stat(approvedReceipt); err != nil {
					t.Fatal("approved process did not start")
				}
				if trust == "trusted" && nonempty {
					require(t, out, 0, "project_tool")
					called = a.call("mcpServer/tool/call", map[string]any{"threadId": id, "server": "project_tool", "tool": "project_tool", "arguments": map[string]any{}})
					require(t, encoded(called), 0, "project_tool_CALLED")
				} else {
					absent(t, out, "project_tool")
					if _, err := os.Stat(projectReceipt); !os.IsNotExist(err) {
						t.Fatal("excluded project MCP process started")
					}
				}
			})
		}
	}
}

// The real vendor entry skips without opt-in, rejects wrong bytes specifically
// at the digest gate and rejects an invalid run root before creating a fixture.
// Positive qualified-byte execution is supplied by this test's outer setup.
func TestIntakeBoundary(t *testing.T) {
	f := setup(t)
	childRoot := filepath.Join(f.root, "intake-child")
	oldEnv := append([]string(nil), f.env...)
	for _, tc := range []struct {
		name, binary, root, want string
		exit                     int
	}{
		{"missing-opt-in", "", childRoot, "--- SKIP: TestInstructions", 0},
		{"wrong-digest", "/usr/bin/true", childRoot, "unqualified binary digest", 1},
		{"invalid-run-root", f.binary, "relative-root", "explicit absolute task-local .temp run root required", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f.env = append(append([]string(nil), oldEnv...), "CODEX_EXCL_BINARY="+tc.binary, "CODEX_EXCL_RUN_ROOT="+tc.root)
			out, code := f.run(tc.name, "", os.Args[0], "-test.run=^TestInstructions$", "-test.v")
			if code != tc.exit || !strings.Contains(out, tc.want) {
				t.Fatalf("expected specific intake verdict %q, exit %d; got %d: %s", tc.want, tc.exit, code, out)
			}
			if _, err := os.Stat(filepath.Join(childRoot, "TestInstructions")); !os.IsNotExist(err) {
				t.Fatal("refused entry created fixture")
			}
		})
	}
	f.env = oldEnv
}

// P4: installed plugin skill bytes are loaded through the real skills loader.
// Project plugin enablement overrides a home deny only when trusted; the
// approved home plugin stays available in both cases. This measures installed
// local plugins, not every marketplace, plugin hook event or remote refresh.
func TestPluginSkills(t *testing.T) {
	for _, trust := range []string{"trusted", "untrusted"} {
		t.Run(trust, func(t *testing.T) {
			f := setup(t)
			for _, name := range []string{"approved", "project"} {
				root := filepath.Join(f.codex, "plugins", "cache", "fixture", name, "local")
				put(t, filepath.Join(root, ".codex-plugin", "plugin.json"), fmt.Sprintf(`{"name":%q}`, name))
				skill(t, filepath.Join(root, "skills"), name+"-skill", strings.ToUpper(name)+"_PLUGIN_SKILL")
			}
			f.config(fmt.Sprintf(`
[features]
plugins=true
remote_plugin=false
[projects.%q]
trust_level=%q
[plugins."approved@fixture"]
enabled=true
[plugins."project@fixture"]
enabled=false
`, f.work, trust))
			put(t, filepath.Join(f.work, ".codex", "config.toml"), "[plugins.\"project@fixture\"]\nenabled=true\n")
			a := f.api("plugin-skills-api")
			result := a.call("skills/list", map[string]any{"cwds": []string{f.work}, "forceReload": true})
			f.save("plugin-skills.json", encoded(result))
			require(t, encoded(result), 0, "APPROVED_PLUGIN_SKILL")
			if trust == "trusted" {
				require(t, encoded(result), 0, "PROJECT_PLUGIN_SKILL")
			} else {
				absent(t, encoded(result), "PROJECT_PLUGIN_SKILL")
			}
		})
	}
}
