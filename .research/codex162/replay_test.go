package codex162_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// These opt-in tests exercise the installed vendor CLI with synthetic homes,
// credentials and a loopback Responses endpoint. They do not qualify a live
// service, token refresh, interactive model behavior, or other vendor builds.
type fixture struct {
	t                                         *testing.T
	root, home, codex, work, binary, evidence string
	privateHome                               string
	env                                       []string
	mu                                        sync.Mutex
	requests                                  []string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	binary := os.Getenv("CODEX162_BINARY")
	if binary == "" {
		t.Skip("opt-in: set CODEX162_BINARY to Codex 0.162.1; no vendor process started")
	}
	var err error
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	// Keep all disposable writes within the run worktree when TMPDIR is set.
	f := &fixture{t: t, root: t.TempDir(), binary: binary}
	f.privateHome, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	f.home, f.codex, f.work = filepath.Join(f.root, "home"), filepath.Join(f.root, "codex"), filepath.Join(f.root, "work")
	for _, dir := range []string{f.home, f.codex, f.work, filepath.Join(f.root, "tmp")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if base := os.Getenv("CODEX162_EVIDENCE"); base != "" {
		f.evidence = filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
		if err := os.MkdirAll(f.evidence, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	// No inherited credentials, proxies, session metadata or user configuration.
	f.env = []string{"PATH=/usr/bin:/bin:/usr/sbin:/sbin", "HOME=" + f.home,
		"CODEX_HOME=" + f.codex, "XDG_CONFIG_HOME=" + filepath.Join(f.home, "config"),
		"XDG_CACHE_HOME=" + filepath.Join(f.home, "cache"), "XDG_DATA_HOME=" + filepath.Join(f.home, "data"),
		"XDG_STATE_HOME=" + filepath.Join(f.home, "state"), "TMPDIR=" + filepath.Join(f.root, "tmp"),
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CEILING_DIRECTORIES=" + f.root,
		"TERM=xterm-256color", "LANG=en_US.UTF-8"}
	out, code := f.run("version", "", "--version")
	if code != 0 || strings.TrimSpace(out) != "codex-cli 0.162.1" {
		t.Fatalf("requires exact 0.162.1, exit %d: %s", code, out)
	}
	payload, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	f.save("identity.txt", fmt.Sprintf("version=%splatform=%s/%s\nsha256=%x\n", out, runtime.GOOS, runtime.GOARCH, sha256.Sum256(payload)))
	return f
}

func (f *fixture) write(path, value string) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) save(name, value string) {
	f.t.Helper()
	if f.evidence == "" {
		return
	}
	value = strings.ReplaceAll(value, f.root, "$RUN")
	value = strings.ReplaceAll(value, f.binary, "$CODEX")
	value = strings.ReplaceAll(value, filepath.Clean(os.TempDir()), "$TMP")
	// The TUI can abbreviate a path before the full fixture prefix ends.
	// Redact the shorter operator-home prefix as well, even in that case.
	value = strings.ReplaceAll(value, f.privateHome, "$OPERATOR_HOME")
	if err := os.WriteFile(filepath.Join(f.evidence, name), []byte(value), 0o600); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) command(name, input, binary string, args ...string) (string, int) {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env, cmd.Dir, cmd.Stdin = f.env, f.work, strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			f.t.Fatalf("start %s: %v", name, err)
		}
	}
	f.save(name+".log", fmt.Sprintf("cwd=%s\nargv=%q\nexit=%d\n%s", f.work, append([]string{binary}, args...), code, out))
	f.t.Logf("%s exit=%d", name, code)
	if ctx.Err() != nil {
		f.t.Fatalf("%s timed out: %s", name, out)
	}
	return string(out), code
}

func (f *fixture) run(name, input string, args ...string) (string, int) {
	f.t.Helper()
	return f.command(name, input, f.binary, args...)
}

func requireOutput(t *testing.T, out string, code, wantCode int, contains ...string) {
	t.Helper()
	if code != wantCode {
		t.Fatalf("exit=%d want=%d: %s", code, wantCode, out)
	}
	for _, s := range contains {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q: %s", s, out)
		}
	}
}

func (f *fixture) endpoint() {
	f.t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read", 500)
			return
		}
		f.mu.Lock()
		f.requests = append(f.requests, string(payload))
		f.mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		if _, err := fmt.Fprint(w, "event: response.created\ndata: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_probe\"}}\n\n"); err != nil {
			return
		}
		_, _ = fmt.Fprint(w, "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_probe\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":0,\"total_tokens\":1}}}\n\n")
	}))
	f.t.Cleanup(server.Close)
	f.write(filepath.Join(f.codex, "config.toml"), fmt.Sprintf(`model = "probe-base"
model_provider = "qualification"
cli_auth_credentials_store = "file"
approval_policy = "never"
sandbox_mode = "read-only"
[model_providers.qualification]
name = "Loopback qualification"
base_url = %q
wire_api = "responses"
requires_openai_auth = false
request_max_retries = 0
stream_max_retries = 0
`, server.URL))
}

// CODEX_HOME selects both the root instructions and skills. A second empty home
// is a negative control: neither synthetic marker may leak into that prompt.
func TestCodex162HomeRootSkills(t *testing.T) {
	f := newFixture(t)
	f.write(filepath.Join(f.codex, "AGENTS.md"), "ROOT_162_SENTINEL\n")
	f.write(filepath.Join(f.codex, "skills", "qualification", "SKILL.md"), "---\nname: qualification\ndescription: SKILL_162_SENTINEL\n---\nSynthetic qualification skill.\n")
	// A local Git root prevents unrelated ancestor instructions from entering evidence.
	out, code := f.command("git-init", "", "/usr/bin/git", "init", "-q")
	requireOutput(t, out, code, 0)
	out, code = f.run("root-skills", "", "debug", "prompt-input", "probe")
	requireOutput(t, out, code, 0, "ROOT_162_SENTINEL", "SKILL_162_SENTINEL")
	var messages []json.RawMessage
	if err := json.Unmarshal([]byte(out), &messages); err != nil {
		t.Fatal(err)
	}
	skill := filepath.Join(f.codex, "skills", "qualification")
	skillSource := filepath.Join(f.root, "skill-source")
	if err := os.Rename(skill, skillSource); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(skillSource, skill); err != nil {
		t.Fatal(err)
	}
	out, code = f.run("linked-skill", "", "debug", "prompt-input", "probe")
	requireOutput(t, out, code, 0, "ROOT_162_SENTINEL", "SKILL_162_SENTINEL")
	other := filepath.Join(f.root, "other-codex")
	if err := os.Mkdir(other, 0o700); err != nil {
		t.Fatal(err)
	}
	for i, e := range f.env {
		if strings.HasPrefix(e, "CODEX_HOME=") {
			f.env[i] = "CODEX_HOME=" + other
		}
	}
	out, code = f.run("empty-home-control", "", "debug", "prompt-input", "probe")
	requireOutput(t, out, code, 0)
	if strings.Contains(out, "ROOT_162_SENTINEL") || strings.Contains(out, "SKILL_162_SENTINEL") {
		t.Fatal("home isolation failed")
	}
}

// The vendor's global AGENTS.md loader is measured above 32 KiB, including its
// tail. The 32768-byte registry value is a Curator advisory, not a vendor cap.
func TestCodex162RootSize(t *testing.T) {
	f := newFixture(t)
	out, code := f.command("git-init", "", "/usr/bin/git", "init", "-q")
	requireOutput(t, out, code, 0)
	f.write(filepath.Join(f.codex, "AGENTS.md"), "ROOT_START\n"+strings.Repeat("x", 33000)+"\nROOT_TAIL_162\n")
	out, code = f.run("large-root", "", "debug", "prompt-input", "probe")
	requireOutput(t, out, code, 0, "ROOT_START", "ROOT_TAIL_162")
}

// Production Install -> Resolve retains an oversized managed root and warns,
// while a small root does not warn. Codex then loads that actual managed root.
// This measures two sizes around the advisory, not an unlimited context cap.
func TestCodex162ManagedAdvisory(t *testing.T) {
	f := newFixture(t)
	for key, value := range map[string]string{
		"HOME": f.home, "CODEX_HOME": f.codex, "CLAUDE_CONFIG_DIR": filepath.Join(f.home, "claude"),
		"PI_CODING_AGENT_DIR": filepath.Join(f.home, "pi"), "XDG_CONFIG_HOME": filepath.Join(f.home, "config"),
		"XDG_DATA_HOME": filepath.Join(f.home, "data"), "XDG_STATE_HOME": filepath.Join(f.home, "state"),
		"XDG_CACHE_HOME": filepath.Join(f.home, "cache"),
	} {
		t.Setenv(key, value)
	}
	f.write(filepath.Join(f.codex, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
	manager := filepath.Join(f.root, "manager")
	for _, tc := range []struct {
		name string
		size int
		warn bool
	}{
		{"small", 100, false}, {"large", 33000, true},
	} {
		source := filepath.Join(f.root, tc.name+"-source")
		f.write(filepath.Join(source, "agent-context.json"), fmt.Sprintf(`{"schema_version":1,"name":%q,"version":"1.0.0","context":{"modules":[{"path":"root.md"}]}}`, tc.name))
		f.write(filepath.Join(source, "context", "root.md"), strings.Repeat("x", tc.size)+"\nMANAGED_TAIL_162\n")
		if _, _, _, err := envprofile.Install(manager, envprofile.InstallOptions{Operand: source}); err != nil {
			t.Fatal(err)
		}
		result, err := envprofile.Resolve(envprofile.ResolveRequest{
			Home: manager, Profile: tc.name, EnvID: envregistry.CodexCLI, LaunchDir: f.work, Repair: true,
			Machine: envregistry.DefaultMachineConfig(), Detect: func(envregistry.Adapter) string { return "0.162.1" },
			NativeHomeOf: func(string) (string, error) { return f.codex, nil },
		})
		if err != nil {
			t.Fatal(err)
		}
		warning := strings.Join(result.Warnings, "\n")
		f.save(tc.name+"-resolve.txt", warning)
		if strings.Contains(warning, envregistry.DiagSizeExceeded) != tc.warn {
			t.Fatalf("%s advisory=%q", tc.name, warning)
		}
		managed := envprofile.ManagedHomeDir(manager, tc.name, envregistry.CodexCLI)
		data, err := os.ReadFile(filepath.Join(managed, "AGENTS.md"))
		if err != nil || !strings.Contains(string(data), "MANAGED_TAIL_162") {
			t.Fatalf("managed bytes lost: %v", err)
		}
		for i, e := range f.env {
			if strings.HasPrefix(e, "CODEX_HOME=") {
				f.env[i] = "CODEX_HOME=" + managed
			}
		}
		out, code := f.command(tc.name+"-git-init", "", "/usr/bin/git", "init", "-q")
		requireOutput(t, out, code, 0)
		out, code = f.run(tc.name+"-managed-prompt", "", "debug", "prompt-input", "probe")
		requireOutput(t, out, code, 0, "MANAGED_TAIL_162")
	}
}

// Actual exec accepts a seeded config and named layer, silently ignores an
// absent layer, rejects a repeated -p, and refuses a non-repository without
// --skip-git-repo-check. Loopback requests prove which model was selected.
func TestCodex162ExecSeedProfile(t *testing.T) {
	f := newFixture(t)
	f.endpoint()
	f.write(filepath.Join(f.codex, "curator-mcp.config.toml"), "model = \"probe-layer\"\n[mcp_servers]\n")
	// Inspect from a pre-existing non-Git directory, without writing there.
	// A nested fixture is still inside the outer Story repository; Codex does
	// not use GIT_CEILING_DIRECTORIES for its own repository discovery.
	var err error
	f.work, err = filepath.EvalSymlinks("/tmp")
	if err != nil {
		t.Fatal(err)
	}
	out, code := f.command("confirm-outside-git", "", "/usr/bin/git", "rev-parse", "--show-toplevel")
	requireOutput(t, out, code, 128, "not a git repository")
	out, code = f.run("outside-git-refusal", "", "exec", "--ephemeral", "probe")
	requireOutput(t, out, code, 1, "--skip-git-repo-check")
	for _, tc := range []struct{ name, profile, model string }{
		{"seed", "", "probe-base"}, {"layer", "curator-mcp", "probe-layer"}, {"missing", "missing", "probe-base"},
	} {
		args := []string{"exec", "--ephemeral", "--skip-git-repo-check"}
		if tc.profile != "" {
			args = append(args, "-p", tc.profile)
		}
		args = append(args, "probe")
		out, code = f.run(tc.name, "", args...)
		requireOutput(t, out, code, 0)
		f.mu.Lock()
		requests := append([]string(nil), f.requests...)
		f.mu.Unlock()
		if len(requests) == 0 {
			t.Fatal("no loopback request")
		}
		request := requests[len(requests)-1]
		f.save(tc.name+"-request.json", request)
		var body struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal([]byte(request), &body); err != nil {
			t.Fatal(err)
		}
		if body.Model != tc.model {
			t.Fatalf("model=%s want=%s", body.Model, tc.model)
		}
	}
	out, code = f.run("repeated-profile", "", "exec", "--skip-git-repo-check", "-p", "one", "-p", "two", "probe")
	requireOutput(t, out, code, 2, "cannot be used multiple times")
}

// Linked file credentials are read from CODEX_HOME. A synthetic login writes
// through the link without replacing it, and another empty home is logged out.
// This measures API-key file auth, not OAuth refresh or native credential stores.
func TestCodex162LinkedAuth(t *testing.T) {
	f := newFixture(t)
	f.write(filepath.Join(f.codex, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
	native := filepath.Join(f.root, "native", "auth.json")
	f.write(native, "{\"OPENAI_API_KEY\":\"synthetic-qualification-old\"}\n")
	link := filepath.Join(f.codex, "auth.json")
	if err := os.Symlink(native, link); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(native)
	if err != nil {
		t.Fatal(err)
	}
	out, code := f.run("linked-status", "", "login", "status")
	requireOutput(t, out, code, 0, "Logged in")
	out, code = f.run("linked-login", "synthetic-qualification-new\n", "login", "--with-api-key")
	requireOutput(t, out, code, 0, "Successfully logged in")
	target, err := os.Readlink(link)
	if err != nil || target != native {
		t.Fatalf("auth link replaced: %q %v", target, err)
	}
	after, err := os.Stat(native)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("credential write replaced native inode")
	}
	data, err := os.ReadFile(native)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "synthetic-qualification-new") || strings.Contains(string(data), "synthetic-qualification-old") {
		t.Fatal("native credential not updated")
	}
	f.save("auth-state.txt", "link preserved=true\nnative inode preserved=true\nnative synthetic credential updated=true\n")
	other := filepath.Join(f.root, "other-codex")
	f.write(filepath.Join(other, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
	for i, e := range f.env {
		if strings.HasPrefix(e, "CODEX_HOME=") {
			f.env[i] = "CODEX_HOME=" + other
		}
	}
	out, code = f.run("other-home-status", "", "login", "status")
	requireOutput(t, out, code, 1, "Not logged in")
	for i, e := range f.env {
		if strings.HasPrefix(e, "CODEX_HOME=") {
			f.env[i] = "CODEX_HOME=" + f.codex
		}
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	out, code = f.run("unlinked-status", "", "login", "status")
	requireOutput(t, out, code, 1, "Not logged in")
}

// Curator's real MCPFile output loads through the registry's named -p layer;
// without the layer that server is absent. model_instructions_file replaces
// the base instructions with synthetic bytes in the actual exec request.
func TestCodex162ManagedChannels(t *testing.T) {
	f := newFixture(t)
	f.endpoint()
	out, code := f.command("git-init", "", "/usr/bin/git", "init", "-q")
	requireOutput(t, out, code, 0)
	path, document, written, err := contextmaterialize.MCPFile("codex_cli", []contextmaterialize.MCPServer{{
		Name: "qualification", Transport: "stdio", Command: "/usr/bin/true", Args: []string{},
	}})
	if err != nil || !written {
		t.Fatalf("MCPFile: written=%v err=%v", written, err)
	}
	f.write(filepath.Join(f.codex, path), string(document))
	f.save("curator-mcp.config.toml", string(document))
	out, code = f.run("mcp-layer", "", "-p", "curator-mcp", "mcp", "list", "--json")
	requireOutput(t, out, code, 0, "qualification")
	out, code = f.run("mcp-no-layer", "", "mcp", "list", "--json")
	requireOutput(t, out, code, 0)
	if strings.Contains(out, "qualification") {
		t.Fatal("MCP layer leaked without -p")
	}
	instructions := filepath.Join(f.codex, "instructions.md")
	f.write(instructions, "SYSTEM_REPLACEMENT_162\n")
	out, code = f.run("system-replacement", "", "exec", "--ephemeral", "-c", "model_instructions_file="+fmt.Sprintf("%q", instructions), "probe")
	requireOutput(t, out, code, 0)
	f.mu.Lock()
	requests := append([]string(nil), f.requests...)
	f.mu.Unlock()
	if len(requests) != 1 {
		t.Fatalf("requests=%d want=1", len(requests))
	}
	f.save("system-request.json", requests[0])
	var body struct {
		Input []struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"input"`
	}
	if err := json.Unmarshal([]byte(requests[0]), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Input) == 0 || body.Input[0].Role != "developer" || len(body.Input[0].Content) != 1 || body.Input[0].Content[0].Text != "SYSTEM_REPLACEMENT_162" {
		t.Fatalf("first developer message does not carry the replacement: %+v", body.Input)
	}
}

// Actual workspace-write permits ordinary files while refusing writes to the
// root-level .git of each writable root. A descendant .git control is writable.
// A separate Git directory outside the writable root
// remains writable when explicitly added, including through real git config.
// These are macOS Seatbelt observations, not a Linux/Windows qualification.
func TestCodex162GitSandbox(t *testing.T) {
	f := newFixture(t)
	if runtime.GOOS != "darwin" {
		t.Skip("pilot sandbox is macOS Seatbelt")
	}
	for _, work := range []string{f.work, filepath.Join(f.work, "nested"), filepath.Join(f.root, "additional")} {
		out, code := f.command("git-init-"+filepath.Base(work), "", "/usr/bin/git", "init", "-q", work)
		requireOutput(t, out, code, 0)
	}
	for _, path := range []string{".git", "nested/.git"} {
		f.write(filepath.Join(f.work, path, "sentinel"), "unchanged\n")
	}
	for _, tc := range []struct {
		name, path string
		code       int
	}{
		{"ordinary", "ordinary", 0}, {"root-git", ".git/sentinel", 1}, {"nested-git", "nested/.git/sentinel", 0},
	} {
		out, code := f.run(tc.name, "", "sandbox", "-c", "sandbox_mode=\"workspace-write\"", "--", "/bin/sh", "-c", "printf changed > \"$1\"", "probe", tc.path)
		requireOutput(t, out, code, tc.code)
		if tc.code != 0 {
			requireOutput(t, out, code, tc.code, "Operation not permitted")
			data, err := os.ReadFile(filepath.Join(f.work, tc.path))
			if err != nil || string(data) != "unchanged\n" {
				t.Fatalf("protected state changed: %q %v", data, err)
			}
		}
	}
	additional := filepath.Join(f.root, "additional")
	f.write(filepath.Join(additional, ".git", "sentinel"), "unchanged\n")
	out, code := f.run("additional-root-git", "", "sandbox", "-c", "sandbox_mode=\"workspace-write\"", "-c", fmt.Sprintf("sandbox_workspace_write.writable_roots=[%q]", additional), "--", "/bin/sh", "-c", "printf changed > \"$1\"", "probe", filepath.Join(additional, ".git", "sentinel"))
	requireOutput(t, out, code, 1, "Operation not permitted")
	data, err := os.ReadFile(filepath.Join(additional, ".git", "sentinel"))
	if err != nil || string(data) != "unchanged\n" {
		t.Fatalf("additional .git changed: %q %v", data, err)
	}
	separate := filepath.Join(f.root, "git-metadata")
	// Remove only this synthetic .git directory before initializing the separate layout.
	if err := os.RemoveAll(filepath.Join(f.work, ".git")); err != nil {
		t.Fatal(err)
	}
	out, code = f.command("separate-init", "", "/usr/bin/git", "init", "--separate-git-dir", separate, ".")
	requireOutput(t, out, code, 0)
	f.env = append(f.env, "GIT_DIR="+separate, "GIT_WORK_TREE="+f.work)
	out, code = f.run("separate-write", "", "sandbox", "-c", "sandbox_mode=\"workspace-write\"", "-c", fmt.Sprintf("sandbox_workspace_write.writable_roots=[%q]", separate), "--", "/usr/bin/git", "config", "qualification.sentinel", "accepted")
	requireOutput(t, out, code, 0)
	out, code = f.command("separate-read", "", "/usr/bin/git", "config", "--get", "qualification.sentinel")
	requireOutput(t, out, code, 0, "accepted")
}

// A private background terminal exercises the actual startup consent screen.
// Consent persists a hook hash for the first project; a fresh invocation there
// skips review, while a second project needs its own hook review. This test
// qualifies startup consent and its scope, not execution of trusted hooks.
// tmux is used only to provide a PTY; no user's terminal or server is reused.
func TestCodex162ProjectHookTrust(t *testing.T) {
	f := newFixture(t)
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Fatal("opt-in trust replay requires tmux for a private PTY")
	}
	f.endpoint()
	second := filepath.Join(f.root, "second-work")
	if err := os.Mkdir(second, 0o700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(f.codex, "config.toml")
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	f.write(configPath, string(config)+fmt.Sprintf("\n[projects.%q]\ntrust_level = \"trusted\"\n[projects.%q]\ntrust_level = \"trusted\"\n", f.work, second))
	f.write(filepath.Join(f.codex, "auth.json"), "{\"OPENAI_API_KEY\":\"synthetic-qualification-only\"}\n")
	for _, work := range []string{f.work, second} {
		out, code := f.command("git-init", "", "/usr/bin/git", "init", "-q", work)
		requireOutput(t, out, code, 0)
		f.write(filepath.Join(work, ".codex", "config.toml"), "[[hooks.SessionStart]]\nhooks = [{ type = \"command\", command = \"printf hook-ran > hook-ran\" }]\n")
	}
	// A relative socket avoids macOS sockaddr_un's 104-byte path limit in
	// deeply nested managed Story worktrees; every call uses f.work as cwd.
	socket := "tmux.sock"
	call := func(name string, args ...string) string {
		t.Helper()
		out, code := f.command(name, "", tmux, append([]string{"-S", socket}, args...)...)
		requireOutput(t, out, code, 0)
		return out
	}
	t.Cleanup(func() {
		// This socket belongs solely to this fixture, including its child PIDs.
		f.command("terminal-cleanup", "", tmux, "-S", socket, "kill-server")
	})
	start := func(name, work string) {
		t.Helper()
		call(name, "-f", "/dev/null", "new-session", "-d", "-s", name, "-x", "120", "-y", "50", f.binary, "--no-daemon", "--no-alt-screen", "-C", work)
		call(name+"-pid", "display-message", "-p", "-t", name, "#{pane_pid}")
	}
	wait := func(name, needle string) string {
		t.Helper()
		deadline := time.Now().Add(15 * time.Second)
		var screen string
		for time.Now().Before(deadline) {
			screen = call(name+"-screen", "capture-pane", "-p", "-t", name, "-S", "-")
			if strings.Contains(screen, needle) {
				return screen
			}
			if needle == "Trust all and continue" && strings.Contains(screen, "probe-base default") {
				t.Fatalf("%s entered main UI without hook consent: %s", name, screen)
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatalf("terminal %s never showed %q: %s", name, needle, screen)
		return ""
	}
	start("first", f.work)
	wait("first", "Trust all and continue")
	if _, err := os.Stat(filepath.Join(f.work, "hook-ran")); !os.IsNotExist(err) {
		t.Fatalf("hook ran before consent, or unreadable state: %v", err)
	}
	f.save("before-consent.txt", "hook marker absent=true\n")
	call("consent", "send-keys", "-t", "first", "Down", "Enter")
	deadline := time.Now().Add(15 * time.Second)
	trusted := false
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "trusted_hash") {
			f.save("trusted-config.toml", string(data))
			trusted = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !trusted {
		t.Fatal("consent did not persist trusted_hash")
	}
	start("repeat", f.work)
	screen := wait("repeat", "probe-base default")
	if strings.Contains(screen, "Hooks need review") {
		t.Fatal("unchanged trusted project prompts again")
	}
	f.save("after-consent.txt", "trusted hash persisted=true\nfresh invocation in same project skips hook review=true\ntrusted hook execution=unmeasured\n")
	start("second", second)
	wait("second", "Trust all and continue")
	if _, err := os.Stat(filepath.Join(second, "hook-ran")); !os.IsNotExist(err) {
		t.Fatalf("second hook ran without its consent: %v", err)
	}
	f.save("second-project.txt", "second project prompts again=true\nhook marker absent=true\n")
}
