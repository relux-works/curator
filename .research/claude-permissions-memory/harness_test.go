// Package claudequalification measures bounded native atoms; it grants no adapter admission.
package claudequalification

import (
	"bufio"
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

var binaryDigests = map[string]string{
	"2.1.305": "ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f",
	"2.1.307": "5b113a7752119470f1da419cb0e608af3361cae25b614216b13022093958fb2a",
}

const approved = "EXCL_APPROVED_ROOT_7352"
const foreign = "EXCL_FOREIGN_ROOT_7352"

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
	actions                                  []map[string]any
	before                                   func(int)
	input                                    string
	actionFor                                func(map[string]any, int) map[string]any
	actionStart                              int
}

func setup(t *testing.T) *probe {
	t.Helper()
	binary := os.Getenv("CLAUDE_QUAL_BINARY")
	if binary == "" {
		t.Skip("opt-in: CLAUDE_QUAL_BINARY unset; no vendor process or fixture")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("pinned probe requires darwin/arm64")
	}
	b, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if expected := binaryDigests[os.Getenv("CLAUDE_QUAL_VERSION")]; expected == "" || hex.EncodeToString(sum[:]) != expected {
		t.Fatal("unqualified binary digest")
	}
	base := os.Getenv("CLAUDE_QUAL_RUN_ROOT")
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
	p.home = filepath.Join(p.root, "home", ".claude")
	p.work = filepath.Join(p.root, "ancestor", "work")
	for _, d := range []string{p.home, p.work, filepath.Join(p.root, "tmp"), filepath.Join(p.root, "config"), filepath.Join(p.root, "data"), filepath.Join(p.root, "state"), filepath.Join(p.root, "cache")} {
		if err := os.MkdirAll(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	p.server = httptest.NewServer(http.HandlerFunc(p.respond))
	t.Cleanup(p.server.Close)
	p.env = []string{"HOME=" + filepath.Dir(p.home), "CLAUDE_CONFIG_DIR=" + p.home, "PATH=/usr/bin:/bin:/usr/sbin:/sbin", "TMPDIR=" + filepath.Join(p.root, "tmp"), "CLAUDE_CODE_TMPDIR=" + filepath.Join(p.root, "tmp"), "XDG_CONFIG_HOME=" + filepath.Join(p.root, "config"), "XDG_DATA_HOME=" + filepath.Join(p.root, "data"), "XDG_STATE_HOME=" + filepath.Join(p.root, "state"), "XDG_CACHE_HOME=" + filepath.Join(p.root, "cache"), "ANTHROPIC_API_KEY=synthetic-not-a-real-key", "ANTHROPIC_BASE_URL=" + p.server.URL, "NO_PROXY=127.0.0.1", "DISABLE_AUTOUPDATER=1", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1", "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1", "DISABLE_TELEMETRY=1", "DISABLE_ERROR_REPORTING=1", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "LANG=en_US.UTF-8", "TERM=dumb"}
	// Only the exact loopback provider port is reachable. All writes stay in the
	// fixture; native directories and security services remain denied throughout.
	port := strings.TrimPrefix(p.server.URL, "http://127.0.0.1:")
	sb := "(version 1)\n(allow default)\n(deny network*)\n(allow network-outbound (remote ip \"localhost:" + port + "\"))\n(deny file-write*)\n(allow file-write* (subpath " + strconv.Quote(p.root) + "))\n(allow file-write* (literal \"/dev/null\"))\n(deny file-read* file-write* (regex #\"/Keychains(/|$)\"))\n(deny mach-lookup (global-name-regex #\"^com\\.apple\\.(security|Security)\"))\n"
	for _, rel := range []string{".claude", ".codex", ".config", ".local", ".curator", ".ssh", "Library"} {
		sb += "(deny file-read* file-write* (subpath " + strconv.Quote(filepath.Join(p.native, rel)) + "))\n"
	}
	for _, dir := range []string{"/Library/Application Support/ClaudeCode", "/Library/Managed Preferences"} {
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
	write(t, filepath.Join(p.home, "CLAUDE.md"), approved+"\n")
	p.settings("user", map[string]any{"model": "claude-sonnet-4-6"})
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

func (p *probe) settings(source string, values map[string]any) {
	p.t.Helper()
	path := filepath.Join(p.home, "settings.json")
	if source == "project" {
		path = filepath.Join(p.work, ".claude", "settings.json")
	}
	if source == "local" {
		path = filepath.Join(p.work, ".claude", "settings.local.json")
	}
	write(p.t, path, mustJSON(p.t, values))
}
func tool(name string, input map[string]any) map[string]any {
	return map[string]any{"type": "tool_use", "id": "tool_probe", "name": name, "input": input}
}

func (p *probe) respond(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
		http.Error(w, "bad fixture request", 400)
		return
	}
	if strings.Contains(r.URL.Path, "count_tokens") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"input_tokens":100}`)
		return
	}
	if !strings.HasSuffix(r.URL.Path, "/messages") {
		http.Error(w, "unexpected fixture endpoint", 404)
		return
	}
	p.mu.Lock()
	n := len(p.requests)
	p.requests = append(p.requests, req)
	p.mu.Unlock()
	if p.before != nil {
		p.before(n)
	}
	content := map[string]any{"type": "text", "text": "OFFLINE_OK"}
	stop := "end_turn"
	if index := n - p.actionStart; index >= 0 && index < len(p.actions) {
		content = p.actions[index]
		stop = "tool_use"
	}
	if p.actionFor != nil {
		content = p.actionFor(req, n)
		stop = "end_turn"
		if content["type"] == "tool_use" {
			stop = "tool_use"
		}
	}
	content = cloneMap(content)
	content["id"] = fmt.Sprintf("tool_probe_%d", n)
	if content["type"] == "text" {
		delete(content, "id")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	event := func(kind string, body any) {
		b, _ := json.Marshal(body)
		_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, b)
	}
	event("message_start", map[string]any{"type": "message_start", "message": map[string]any{"id": fmt.Sprintf("msg_probe_%d", n), "type": "message", "role": "assistant", "model": req["model"], "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": map[string]int{"input_tokens": 100, "output_tokens": 1}}})
	if content["type"] == "tool_use" {
		input, _ := json.Marshal(content["input"])
		content["input"] = map[string]any{}
		event("content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": content})
		event("content_block_delta", map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "input_json_delta", "partial_json": string(input)}})
	} else {
		text := content["text"]
		content["text"] = ""
		event("content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": content})
		event("content_block_delta", map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": text}})
	}
	event("content_block_stop", map[string]any{"type": "content_block_stop", "index": 0})
	event("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": stop, "stop_sequence": nil}, "usage": map[string]int{"output_tokens": 1}})
	event("message_stop", map[string]any{"type": "message_stop"})
}
func cloneMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

type run struct {
	Output, Debug string
	Requests      []map[string]any
	Init, Result  map[string]any
	Exit          int
}

func (p *probe) run(name string, exclude bool, extra ...string) run {
	p.t.Helper()
	args := []string{"-p", "QUALIFICATION_PROBE", "--output-format", "stream-json", "--verbose", "--debug-file", filepath.Join(p.root, name+".debug.log"), "--max-turns", "12", "--permission-mode", "dontAsk"}
	if p.input != "" {
		args = append([]string{"-p", "--input-format", "stream-json"}, args[2:]...)
	}
	if exclude && os.Getenv("CLAUDE_QUAL_OMIT") != "setting-sources" {
		args = append(args, "--setting-sources", "user")
	}
	args = append(args, extra...)
	p.mu.Lock()
	start := len(p.requests)
	p.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	wrapped := []string{"-f", p.policy, "/usr/bin/env", "-i"}
	wrapped = append(wrapped, p.env...)
	wrapped = append(wrapped, p.binary)
	wrapped = append(wrapped, args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", wrapped...)
	cmd.Dir = p.work
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	cmd.Stdin = strings.NewReader(p.input)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		e, ok := err.(*exec.ExitError)
		if !ok {
			p.t.Fatal(err)
		}
		code = e.ExitCode()
	}
	p.mu.Lock()
	requests := append([]map[string]any(nil), p.requests[start:]...)
	p.mu.Unlock()
	result := run{Output: string(out), Requests: requests, Exit: code}
	if b, e := os.ReadFile(filepath.Join(p.root, name+".debug.log")); e == nil {
		result.Debug = string(b)
	}
	scanner := bufio.NewScanner(strings.NewReader(result.Output))
	scanner.Buffer(make([]byte, 4096), 4<<20)
	for scanner.Scan() {
		var msg map[string]any
		if json.Unmarshal(scanner.Bytes(), &msg) == nil {
			if msg["subtype"] == "init" {
				result.Init = msg
			}
			if msg["type"] == "result" {
				result.Result = msg
			}
		}
	}
	if err := scanner.Err(); err != nil {
		p.t.Fatal(err)
	}
	p.save(name+".json", mustJSON(p.t, map[string]any{"argv": args, "exit": code, "pid": cmd.Process.Pid, "init": result.Init, "result": result.Result, "requests": requests, "debug": result.Debug}))
	if ctx.Err() != nil {
		p.t.Fatalf("vendor timeout (not absence evidence); see %s", name)
	}
	if code != 0 || result.Init == nil || result.Result == nil || result.Result["is_error"] != false || len(requests) == 0 {
		p.t.Fatalf("incomplete vendor turn: exit=%d requests=%d: %s", code, len(requests), p.sanitize(string(out)))
	}
	return result
}
func (p *probe) sanitize(s string) string {
	s = strings.ReplaceAll(s, p.root, "$RUN")
	s = strings.ReplaceAll(s, regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(p.root, "-"), "$RUN_ID")
	s = strings.ReplaceAll(s, p.binary, "$CLAUDE")
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
func payload(t *testing.T, r run) string { t.Helper(); return mustJSON(t, r.Requests) }
func contains(t *testing.T, s, marker string, want bool) {
	t.Helper()
	if strings.Contains(s, marker) != want {
		t.Fatalf("observed %s=%t; want %t", marker, !want, want)
	}
}
func (p *probe) assertRoot(r run) { p.t.Helper(); contains(p.t, payload(p.t, r), approved, true) }
