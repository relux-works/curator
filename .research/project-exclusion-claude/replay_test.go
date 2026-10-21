// Package claudeexclusion measures one native build; it grants no adapter admission.
package claudeexclusion

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

const binarySHA = "ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f"
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
}

func setup(t *testing.T) *probe {
	t.Helper()
	binary := os.Getenv("CLAUDE_EXCL_BINARY")
	if binary == "" {
		t.Skip("opt-in: CLAUDE_EXCL_BINARY unset; no vendor process or fixture")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatal("pinned probe requires darwin/arm64")
	}
	b, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != binarySHA {
		t.Fatal("unqualified binary digest")
	}
	base := os.Getenv("CLAUDE_EXCL_RUN_ROOT")
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
	if n < len(p.actions) {
		content = p.actions[n]
		stop = "tool_use"
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
	args := []string{"-p", "EXCLUSION_PROBE", "--output-format", "stream-json", "--verbose", "--debug-file", filepath.Join(p.root, name+".debug.log"), "--max-turns", "5", "--permission-mode", "dontAsk"}
	if p.input != "" {
		args = append([]string{"-p", "--input-format", "stream-json"}, args[2:]...)
	}
	if exclude && os.Getenv("CLAUDE_EXCL_OMIT") != "setting-sources" {
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

// P1 proves initial root, ancestor, local, rule and AGENTS fallback injection
// with a valid native source, then exclusion while managed-home root still loads.
func TestP1Instructions(t *testing.T) {
	for _, kind := range []string{"root", "ancestor", "dot-claude", "local", "rule", "agents-fallback"} {
		for _, exclude := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/exclude-%t", kind, exclude), func(t *testing.T) {
				p := setup(t)
				path := filepath.Join(p.work, "CLAUDE.md")
				switch kind {
				case "ancestor":
					path = filepath.Join(filepath.Dir(p.work), "CLAUDE.md")
				case "dot-claude":
					path = filepath.Join(p.work, ".claude", "CLAUDE.md")
				case "local":
					path = filepath.Join(p.work, "CLAUDE.local.md")
				case "rule":
					path = filepath.Join(p.work, ".claude", "rules", "probe.md")
				case "agents-fallback":
					path = filepath.Join(p.work, "AGENTS.md")
				}
				write(t, path, foreign+"\n")
				r := p.run("turn", exclude)
				p.assertRoot(r)
				contains(t, payload(t, r), foreign, !exclude)
			})
		}
	}
}

// P2 proves lazy nested CLAUDE.md/rules injection on a real Read tool result,
// and its exclusion without hiding readable work data or approved instructions.
func TestP2Nested(t *testing.T) {
	for _, kind := range []string{"claude", "rule", "agents"} {
		for _, exclude := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/exclude-%t", kind, exclude), func(t *testing.T) {
				p := setup(t)
				file := filepath.Join(p.work, "sub", "ordinary.txt")
				write(t, file, "EXCL_ORDINARY_DATA\n")
				path := filepath.Join(p.work, "sub", "CLAUDE.md")
				if kind == "rule" {
					path = filepath.Join(p.work, ".claude", "rules", "nested.md")
				}
				if kind == "agents" {
					path = filepath.Join(p.work, "sub", "AGENTS.md")
				}
				body := foreign + "\n"
				if kind == "rule" {
					body = "---\npaths:\n  - \"sub/**\"\n---\n" + body
				}
				write(t, path, body)
				p.actions = []map[string]any{tool("Read", map[string]any{"file_path": file})}
				r := p.run("turn", exclude, "--allowedTools", "Read")
				p.assertRoot(r)
				contains(t, payload(t, r), "EXCL_ORDINARY_DATA", true)
				contains(t, payload(t, r), foreign, !exclude)
			})
		}
	}
}
func (p *probe) skill(base, name, description string) {
	write(p.t, filepath.Join(base, "skills", name, "SKILL.md"), "---\nname: "+name+"\ndescription: "+description+"\n---\nEXCL_SKILL_BODY_"+name+"\n")
}

// P3 proves project skill catalog exclusion and approved home skill discovery;
// it does not infer invocation behavior from the model's response text.
func TestP3Skills(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			p.skill(filepath.Join(p.work, ".claude"), "foreign-probe", "EXCL_FOREIGN_SKILL")
			p.skill(p.home, "approved-probe", "EXCL_APPROVED_SKILL")
			p.actions = []map[string]any{tool("Skill", map[string]any{"skill": "approved-probe"}), tool("Skill", map[string]any{"skill": "foreign-probe"})}
			r := p.run("turn", exclude, "--allowedTools", "Skill")
			p.assertRoot(r)
			contains(t, payload(t, r), "EXCL_SKILL_BODY_approved-probe", true)
			contains(t, payload(t, r), "EXCL_SKILL_BODY_foreign-probe", !exclude)
			contains(t, payload(t, r), "Unknown skill: foreign-probe", exclude)
			contains(t, mustJSON(t, r.Init["skills"]), "approved-probe", true)
			contains(t, mustJSON(t, r.Init["skills"]), "foreign-probe", !exclude)
			contains(t, payload(t, r), "EXCL_APPROVED_SKILL", true)
			contains(t, payload(t, r), "EXCL_FOREIGN_SKILL", !exclude)
		})
	}
}
func (p *probe) hook(name string) map[string]any {
	path := filepath.Join(p.root, name+".witness")
	return map[string]any{"SessionStart": []any{map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "printf '" + name + "\\n' > " + quote(path)}}}}}
}

// P4 hooks proves startup execution and forbidden-effect absence using files,
// with an approved home hook executed in both the baseline and exclusion case.
func TestP4Hooks(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			p.settings("user", map[string]any{"hooks": p.hook("approved-hook")})
			p.settings("project", map[string]any{"hooks": p.hook("foreign-hook")})
			r := p.run("turn", exclude, "--include-hook-events")
			p.assertRoot(r)
			contains(t, read(t, filepath.Join(p.root, "approved-hook.witness")), "approved-hook", true)
			_, err := os.Stat(filepath.Join(p.root, "foreign-hook.witness"))
			if exclude {
				if !os.IsNotExist(err) {
					t.Fatalf("forbidden project hook effect: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// P4 plugins probes project-scoped installed-plugin activation, using synthetic
// local metadata and a home-approved plugin. It never syncs a remote marketplace.
func TestP4Plugins(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			plugins := map[string]any{}
			known := map[string]any{}
			for _, scope := range []string{"project", "user"} {
				name := "foreign-plugin"
				marketplace := filepath.Join(p.work, ".claude", "fixture-marketplace")
				if scope == "user" {
					name = "approved-plugin"
					marketplace = filepath.Join(p.home, "fixture-marketplace")
				}
				marketName := scope + "-fixture"
				base := filepath.Join(marketplace, "plugin")
				write(t, filepath.Join(base, ".claude-plugin", "plugin.json"), mustJSON(t, map[string]any{"name": name, "version": "1.0.0"}))
				p.skill(base, name, "EXCL_"+name)
				write(t, filepath.Join(base, "hooks", "hooks.json"), mustJSON(t, map[string]any{"hooks": p.hook(name)}))
				write(t, filepath.Join(marketplace, ".claude-plugin", "marketplace.json"), mustJSON(t, map[string]any{"name": marketName, "owner": map[string]string{"name": "Synthetic fixture"}, "plugins": []any{map[string]string{"name": name, "source": "./plugin"}}}))
				known[marketName] = map[string]any{"source": map[string]string{"source": "directory", "path": marketplace}, "installLocation": marketplace, "lastUpdated": "2026-10-01T00:00:00Z", "autoUpdate": false}
				entry := map[string]any{"scope": scope, "installPath": base, "version": "1.0.0", "installedAt": "2026-10-01T00:00:00Z", "lastUpdated": "2026-10-01T00:00:00Z"}
				if scope == "project" {
					entry["projectPath"] = p.work
				}
				plugins[name+"@"+marketName] = []any{entry}
			}
			write(t, filepath.Join(p.home, "plugins", "known_marketplaces.json"), mustJSON(t, known))
			write(t, filepath.Join(p.home, "plugins", "installed_plugins.json"), mustJSON(t, map[string]any{"version": 2, "plugins": plugins}))
			p.settings("user", map[string]any{"enabledPlugins": map[string]bool{"approved-plugin@user-fixture": true}})
			p.settings("project", map[string]any{"enabledPlugins": map[string]bool{"foreign-plugin@project-fixture": true}})
			r := p.run("turn", exclude)
			p.assertRoot(r)
			for _, tc := range []struct {
				name string
				want bool
			}{{"approved-plugin", true}, {"foreign-plugin", !exclude}} {
				contains(t, mustJSON(t, r.Init["plugins"]), tc.name, tc.want)
				contains(t, mustJSON(t, r.Init["skills"]), tc.name+":"+tc.name, tc.want)
				_, err := os.Stat(filepath.Join(p.root, tc.name+".witness"))
				if tc.want {
					if err != nil {
						t.Fatalf("missing plugin hook %s: %v", tc.name, err)
					}
				} else if !os.IsNotExist(err) {
					t.Fatalf("forbidden plugin hook %s: %v", tc.name, err)
				}
			}
		})
	}
}

func (p *probe) mcp(name string) map[string]any {
	script := filepath.Join(p.root, "mcp.py")
	write(p.t, script, read(p.t, "testdata/mcp.py"))
	return map[string]any{"type": "stdio", "command": "/usr/bin/python3", "args": []string{script, filepath.Join(p.root, name+".witness")}}
}

// P5 proves actual stdio startup for a nonempty project MCP source, strict
// empty/nonempty exclusion, and approved MCP process admission through a home file.
func TestP5MCP(t *testing.T) {
	for _, projectEmpty := range []bool{false, true} {
		for _, exclude := range []bool{false, true} {
			for _, strict := range []bool{false, true} {
				for _, admitted := range []bool{false, true} {
					t.Run(fmt.Sprintf("project-empty-%t/exclude-%t/strict-%t/approved-%t", projectEmpty, exclude, strict, admitted), func(t *testing.T) {
						p := setup(t)
						project := map[string]any{}
						if !projectEmpty {
							project["foreign-probe"] = p.mcp("foreign-mcp")
						}
						write(t, filepath.Join(p.work, ".mcp.json"), mustJSON(t, map[string]any{"mcpServers": project}))
						p.settings("user", map[string]any{"enableAllProjectMcpServers": true})
						accepted := map[string]any{}
						if admitted {
							accepted["approved-probe"] = p.mcp("approved-mcp")
						}
						path := filepath.Join(p.home, "approved-mcp.json")
						write(t, path, mustJSON(t, map[string]any{"mcpServers": accepted}))
						args := []string{"--mcp-config", path}
						if strict && os.Getenv("CLAUDE_EXCL_OMIT") != "strict-mcp" {
							args = append(args, "--strict-mcp-config")
						}
						r := p.run("turn", exclude, args...)
						p.assertRoot(r)
						wantForeign := !projectEmpty && !strict && !exclude
						for _, tc := range []struct {
							name string
							want bool
						}{{"foreign-mcp", wantForeign}, {"approved-mcp", admitted}} {
							_, err := os.Stat(filepath.Join(p.root, tc.name+".witness"))
							if tc.want {
								if err != nil {
									t.Fatalf("missing %s startup: %v", tc.name, err)
								}
							} else if !os.IsNotExist(err) {
								t.Fatalf("forbidden %s startup: %v", tc.name, err)
							}
						}
						contains(t, mustJSON(t, r.Init["mcp_servers"]), "foreign-probe", wantForeign)
						contains(t, mustJSON(t, r.Init["mcp_servers"]), "approved-probe", admitted)
					})
				}
			}
		}
	}
}

// P6 proves settings precedence via the requested model and permission behavior
// via real Read results: a home deny survives while a project deny is excluded.
func TestP6SettingsPermissions(t *testing.T) {
	for _, source := range []string{"project", "local"} {
		for _, exclude := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/exclude-%t", source, exclude), func(t *testing.T) {
				p := setup(t)
				write(t, filepath.Join(p.work, "project-denied.txt"), "EXCL_PROJECT_READABLE\n")
				write(t, filepath.Join(p.work, "home-denied.txt"), "EXCL_HOME_MUST_NOT_LEAK\n")
				p.settings("user", map[string]any{"model": "claude-sonnet-4-6", "permissions": map[string]any{"deny": []string{"Read(./home-denied.txt)"}}})
				p.settings(source, map[string]any{"model": "claude-haiku-4-5", "permissions": map[string]any{"deny": []string{"Read(./project-denied.txt)"}}})
				p.actions = []map[string]any{tool("Read", map[string]any{"file_path": filepath.Join(p.work, "project-denied.txt")}), tool("Read", map[string]any{"file_path": filepath.Join(p.work, "home-denied.txt")})}
				r := p.run("turn", exclude, "--allowedTools", "Read")
				p.assertRoot(r)
				s := payload(t, r)
				contains(t, s, "EXCL_PROJECT_READABLE", exclude)
				contains(t, s, "EXCL_HOME_MUST_NOT_LEAK", false)
				model := "claude-haiku-4-5"
				if exclude {
					model = "claude-sonnet-4-6"
				}
				if r.Init["model"] != model {
					t.Fatalf("settings model=%v want %s", r.Init["model"], model)
				}
				contains(t, mustJSON(t, r.Result["permission_denials"]), "home-denied.txt", true)
				contains(t, mustJSON(t, r.Result["permission_denials"]), "project-denied.txt", !exclude)
			})
		}
	}
}

// P7 resume proves a session begun with excluded sources still excludes them
// on explicit --resume, and reloads the approved home context.
func TestP7Resume(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			write(t, filepath.Join(p.work, "CLAUDE.md"), foreign+"\n")
			p.skill(filepath.Join(p.work, ".claude"), "foreign-probe", "EXCL_FOREIGN_SKILL")
			p.skill(p.home, "approved-probe", "EXCL_APPROVED_SKILL")
			first := p.run("first", exclude)
			p.assertRoot(first)
			contains(t, payload(t, first), foreign, !exclude)
			id, ok := first.Init["session_id"].(string)
			if !ok || id == "" {
				t.Fatal("missing resume identity")
			}
			resumed := p.run("resume", exclude, "--resume", id)
			p.assertRoot(resumed)
			contains(t, payload(t, resumed), foreign, !exclude)
			contains(t, mustJSON(t, resumed.Init["skills"]), "approved-probe", true)
			contains(t, mustJSON(t, resumed.Init["skills"]), "foreign-probe", !exclude)
		})
	}
}

// P7 child drives the native Task tool with a fixed loopback tool response.
// All parent/child requests are inspected; there is no model judgment or live inference.
func TestP7Child(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			write(t, filepath.Join(p.work, "CLAUDE.md"), foreign+"\n")
			p.skill(p.home, "approved-probe", "EXCL_APPROVED_SKILL")
			p.skill(filepath.Join(p.work, ".claude"), "foreign-probe", "EXCL_FOREIGN_SKILL")
			for _, tc := range []struct{ base, name, body string }{{p.home, "approved-agent", "EXCL_APPROVED_AGENT"}, {filepath.Join(p.work, ".claude"), "foreign-agent", "EXCL_FOREIGN_AGENT"}} {
				write(t, filepath.Join(tc.base, "agents", tc.name+".md"), "---\nname: "+tc.name+"\ndescription: Synthetic context probe\n---\n"+tc.body+"\n")
			}
			p.actions = []map[string]any{tool("Task", map[string]any{"description": "Synthetic context probe", "prompt": "Return the fixed offline response.", "subagent_type": "approved-agent"})}
			r := p.run("turn", exclude, "--allowedTools", "Task")
			p.assertRoot(r)
			contains(t, mustJSON(t, r.Init["agents"]), "approved-agent", true)
			contains(t, mustJSON(t, r.Init["agents"]), "foreign-agent", !exclude)
			if len(r.Requests) < 3 {
				t.Fatalf("child request did not execute; requests=%d", len(r.Requests))
			}
			child := mustJSON(t, r.Requests[1])
			contains(t, child, approved, true)
			contains(t, child, "EXCL_APPROVED_AGENT", true)
			contains(t, child, foreign, !exclude)
			contains(t, mustJSON(t, r.Result["subagent_stats"]), `"completed": 1`, true)
		})
	}
}

// P8 creates a new nested instruction after the first request, then performs
// Read in the same session. Only this late instruction path is qualified here.
func TestP8LateInstructions(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			file := filepath.Join(p.work, "late", "ordinary.txt")
			write(t, file, "EXCL_LATE_WORK_DATA\n")
			p.before = func(n int) {
				if n == 0 {
					path := filepath.Join(p.work, "late", "CLAUDE.md")
					if err := os.WriteFile(path, []byte(foreign+"\n"), 0600); err != nil {
						panic(err)
					}
				}
			}
			p.actions = []map[string]any{tool("Read", map[string]any{"file_path": file})}
			r := p.run("turn", exclude, "--allowedTools", "Read")
			p.assertRoot(r)
			contains(t, mustJSON(t, r.Requests[0]), foreign, false)
			contains(t, payload(t, r), "EXCL_LATE_WORK_DATA", true)
			contains(t, payload(t, r), foreign, !exclude)
		})
	}
}

// Identity proves the exact release reports 2.1.305 only after the digest gate.
func TestIdentity(t *testing.T) {
	p := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-f", p.policy, "/usr/bin/env", "-i"}
	args = append(args, p.env...)
	args = append(args, p.binary, "--version")
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", args...)
	cmd.Dir = p.work
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	p.save("identity.json", mustJSON(t, map[string]any{"sha256": binarySHA, "platform": runtime.GOOS + "/" + runtime.GOARCH, "version": string(out), "success": err == nil, "pid": cmd.Process.Pid}))
	if err != nil || strings.TrimSpace(string(out)) != "2.1.305 (Claude Code)" {
		t.Fatalf("identity: %v %s", err, out)
	}
}

// The intake control proves missing opt-in skips and a wrong digest is rejected
// before a fixture exists. The fabricated file is never accepted as a vendor.
func TestIntakeBoundary(t *testing.T) {
	p := setup(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrong := filepath.Join(p.root, "wrong-binary")
	write(t, wrong, "synthetic non-vendor bytes\n")
	for _, tc := range []struct{ name, binary, want string }{{"missing", "", "opt-in: CLAUDE_EXCL_BINARY unset"}, {"wrong-digest", wrong, "unqualified binary digest"}} {
		t.Run(tc.name, func(t *testing.T) {
			childRoot := filepath.Join(p.root, tc.name)
			env := []string{}
			for _, v := range os.Environ() {
				if !strings.HasPrefix(v, "CLAUDE_EXCL_BINARY=") && !strings.HasPrefix(v, "CLAUDE_EXCL_RUN_ROOT=") {
					env = append(env, v)
				}
			}
			env = append(env, "CLAUDE_EXCL_BINARY="+tc.binary, "CLAUDE_EXCL_RUN_ROOT="+childRoot)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, self, "-test.run=^TestIdentity$", "-test.v")
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			if (err == nil) != (tc.binary == "") {
				t.Fatalf("unexpected child verdict: %v %s", err, out)
			}
			contains(t, string(out), tc.want, true)
			if _, err := os.Stat(filepath.Join(childRoot, "TestIdentity")); !os.IsNotExist(err) {
				t.Fatalf("intake created vendor fixture: %v", err)
			}
			p.save("intake-"+tc.name+".log", string(out))
		})
	}
}

// Synthetic boundary controls pair reachable provider and ordinary-file reads
// with denied unrelated loopback, synthetic Keychains bytes and outside writes.
// Mutants alter only these curl/cat/sh controls; no vendor runs under a mutant.
func TestOfflineBoundary(t *testing.T) {
	p := setup(t)
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
	if mutant := os.Getenv("CLAUDE_EXCL_BOUNDARY_MUTANT"); mutant != "" {
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

// Late skills are tested across two real SDK turns and explicit /reload-skills;
// approved and project copies are created after the first provider request.
func TestP8LateSkills(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			for _, base := range []string{p.home, filepath.Join(p.work, ".claude")} {
				if err := os.MkdirAll(filepath.Join(base, "skills"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			p.input = ""
			for _, content := range []string{"EXCLUSION_PROBE_FIRST", "/reload-skills", "EXCLUSION_PROBE_SECOND"} {
				b, err := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}})
				if err != nil {
					t.Fatal(err)
				}
				p.input += string(b) + "\n"
			}
			p.before = func(n int) {
				if n == 0 {
					p.skill(p.home, "approved-late", "EXCL_APPROVED_LATE_SKILL")
					p.skill(filepath.Join(p.work, ".claude"), "foreign-late", "EXCL_FOREIGN_LATE_SKILL")
					time.Sleep(600 * time.Millisecond)
				}
			}
			r := p.run("turn", exclude)
			p.assertRoot(r)
			if len(r.Requests) < 2 {
				t.Fatal("second SDK turn did not execute")
			}
			contains(t, mustJSON(t, r.Requests[0]), "EXCL_APPROVED_LATE_SKILL", false)
			contains(t, mustJSON(t, r.Requests[0]), "EXCL_FOREIGN_LATE_SKILL", false)
			later := mustJSON(t, r.Requests[1:])
			contains(t, later, "EXCL_APPROVED_LATE_SKILL", true)
			contains(t, later, "EXCL_FOREIGN_LATE_SKILL", !exclude)
		})
	}
}

// This bounded late-hook observation preserves the failed activation witness:
// a project hook added mid-turn is not executed by the next Read after 800ms.
// A startup-approved hook executes; this does not qualify later hook reloads.
func TestP8LateHooks(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude-%t", exclude), func(t *testing.T) {
			p := setup(t)
			p.settings("project", map[string]any{})
			h := p.hook("approved-late-hook")
			p.settings("user", map[string]any{"hooks": map[string]any{"PreToolUse": h["SessionStart"]}})
			file := filepath.Join(p.work, "ordinary.txt")
			write(t, file, "EXCL_LATE_HOOK_DATA\n")
			p.before = func(n int) {
				if n == 0 {
					for _, tc := range []struct{ source, name string }{{"project", "foreign-late-hook"}} {
						h := p.hook(tc.name)
						p.settings(tc.source, map[string]any{"hooks": map[string]any{"PreToolUse": h["SessionStart"]}})
					}
					time.Sleep(800 * time.Millisecond)
				}
			}
			p.actions = []map[string]any{tool("Read", map[string]any{"file_path": file})}
			r := p.run("turn", exclude, "--allowedTools", "Read")
			p.assertRoot(r)
			contains(t, payload(t, r), "EXCL_LATE_HOOK_DATA", true)
			for _, tc := range []struct {
				name string
				want bool
			}{{"approved-late-hook", true}, {"foreign-late-hook", false}} {
				_, err := os.Stat(filepath.Join(p.root, tc.name+".witness"))
				if tc.want {
					if err != nil {
						t.Fatalf("late hook %s did not execute: %v", tc.name, err)
					}
				} else if !os.IsNotExist(err) {
					t.Fatalf("forbidden late hook effect %s: %v", tc.name, err)
				}
			}
		})
	}
}

// Exclusion does not scrub a transcript previously started with project context.
// This witness covers resume and fork-resume; safe-origin resume is tested above.
func TestP7PreviouslyContaminatedResume(t *testing.T) {
	for _, fork := range []bool{false, true} {
		t.Run(fmt.Sprintf("fork-%t", fork), func(t *testing.T) {
			p := setup(t)
			write(t, filepath.Join(p.work, "CLAUDE.md"), foreign+"\n")
			first := p.run("unexcluded-origin", false)
			p.assertRoot(first)
			contains(t, payload(t, first), foreign, true)
			id, ok := first.Init["session_id"].(string)
			if !ok || id == "" {
				t.Fatal("missing resume identity")
			}
			args := []string{"--resume", id}
			if fork {
				args = append(args, "--fork-session")
			}
			resumed := p.run("excluded-resume", true, args...)
			p.assertRoot(resumed)
			contains(t, payload(t, resumed), foreign, true)
		})
	}
}

// The publication checker accepts a synthetic neutral record, rejects ordinary
// and flattened private locators with its intended diagnostic, and writes no
// partial artifact on rejection. An empty run cannot become absence evidence.
func TestCollectorBoundary(t *testing.T) {
	p := setup(t)
	for _, tc := range []struct {
		name, text string
		valid      bool
	}{{"neutral", "$RUN/home/.claude/CLAUDE.md EXCL_APPROVED_FIXTURE", true}, {"private", "/Users/synthetic-native/CLAUDE.md EXCL_APPROVED_FIXTURE", false}, {"flattened", "-Users-synthetic-native-checkout EXCL_APPROVED_FIXTURE", false}, {"empty", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(p.root, tc.name)
			if err := os.MkdirAll(root, 0700); err != nil {
				t.Fatal(err)
			}
			if tc.name != "empty" {
				data := map[string]any{"argv": []string{"-p"}, "exit": 0, "init": map[string]any{}, "result": map[string]any{}, "debug": "", "requests": []any{map[string]any{"model": "synthetic", "messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": tc.text}}}}}}}
				write(t, filepath.Join(root, "case", "evidence", "turn.json"), mustJSON(t, data))
			}
			target := filepath.Join(root, "public", "out.json")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "/usr/bin/python3", "-B", "collect.py", root, target)
			cmd.Env = []string{"PATH=/usr/bin:/bin"}
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			if tc.valid {
				if err != nil {
					t.Fatalf("valid projection failed: %v %s", err, out)
				}
				contains(t, read(t, target), "EXCL_APPROVED_FIXTURE", true)
			} else {
				if err == nil {
					t.Fatal("checker accepted private or empty evidence")
				}
				cause := "private locator remains; refusing publication"
				if tc.name == "empty" {
					cause = "no completed vendor records; refusing empty evidence"
				}
				contains(t, string(out), cause, true)
				if _, err := os.Stat(target); !os.IsNotExist(err) {
					t.Fatalf("partial publication: %v", err)
				}
			}
			p.save("collector-"+tc.name+".log", string(out))
		})
	}
}
