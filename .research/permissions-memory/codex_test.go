package permissionsmemory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// codexEndpoint speaks deterministic Responses SSE to the native client. It
// selects only a tool actually advertised in that request and records the next
// request's real tool result. Missing tools/results remain unknown.
func codexEndpoint(p *probe, command string) {
	p.prompt = "QUAL_CODEX_MEMORY"
	if command == "@mcp" {
		p.prompt = "QUAL_CODEX_MCP"
	} else if strings.HasPrefix(command, "@memory-") {
		p.prompt = "QUAL_DEDICATED_MEMORY please remember the synthetic qualification note, or read it for this test."
	} else if command != "" {
		p.prompt = "QUAL_CODEX_FIXED_TOOL"
	}

	p.codexReply = func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method == "GET" {
			_, _ = io.WriteString(w, "QUAL_CODEX_WEB_MARKER")
			return
		}
		req := map[string]any{}
		if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
			http.Error(w, "invalid fixture body", 400)
			return
		}
		p.mu.Lock()
		n := len(p.requests)
		main := requestHasUser(req, p.prompt)
		issue := main && !p.issued
		if issue {
			p.issued = true
		}
		finalText := p.finalText
		p.requests = append(p.requests, req)
		p.mu.Unlock()
		event := func(kind string, data any) {
			b, _ := json.Marshal(data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, b)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		event("response.created", map[string]any{"type": "response.created", "response": map[string]any{"id": fmt.Sprintf("resp_%d", n), "status": "in_progress"}})
		output := []any{}
		if issue && (command == "@memory-write" || command == "@memory-read") {
			name := "read"
			args := map[string]any{"path": "MEMORY.md"}
			if command == "@memory-write" {
				name = "add_ad_hoc_note"
				args = map[string]any{"filename": "2026-10-23T00-00-00-qualification.md", "note": "QUAL_CODEX_MEMORY_WRITE"}
			}
			raw, _ := json.Marshal(args)
			item := map[string]any{"id": "fc_memory", "type": "function_call", "call_id": "call_memory", "namespace": "memories", "name": name, "arguments": string(raw), "status": "completed"}
			event("response.output_item.added", map[string]any{"type": "response.output_item.added", "output_index": 0, "item": item})
			event("response.output_item.done", map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item})
			output = append(output, item)
		}
		if issue && command == "@mcp" {
			item := map[string]any{"id": "fc_qual", "type": "function_call", "call_id": "call_qual", "name": "sentinel", "namespace": "mcp__approved", "arguments": "{}", "status": "completed"}
			event("response.output_item.added", map[string]any{"type": "response.output_item.added", "output_index": 0, "item": item})
			event("response.output_item.done", map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item})
			output = append(output, item)
		}
		if issue && command != "" && command != "@mcp" && !strings.HasPrefix(command, "@memory-") {
			tools, _ := req["tools"].([]any)
			for _, raw := range tools {
				tool, _ := raw.(map[string]any)
				name, _ := tool["name"].(string)
				var args map[string]any
				if name == "exec_command" {
					args = map[string]any{"cmd": command, "yield_time_ms": 1000, "max_output_tokens": 1024}
				}
				if name == "shell_command" {
					args = map[string]any{"command": command, "timeout_ms": 1000, "output_limit": 1024}
				}
				if args == nil {
					continue
				}
				b, _ := json.Marshal(args)
				item := map[string]any{"id": "fc_qual", "type": "function_call", "call_id": "call_qual", "name": name, "arguments": string(b), "status": "completed"}
				event("response.output_item.added", map[string]any{"type": "response.output_item.added", "output_index": 0, "item": item})
				event("response.output_item.done", map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item})
				output = append(output, item)
				break
			}
		}
		if len(output) == 0 && main && finalText != "" {
			item := map[string]any{"id": "msg_final", "type": "message", "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": finalText, "annotations": []any{}}}}
			event("response.output_item.added", map[string]any{"type": "response.output_item.added", "output_index": 0, "item": item})
			event("response.output_text.delta", map[string]any{"type": "response.output_text.delta", "output_index": 0, "content_index": 0, "delta": finalText, "item_id": "msg_final"})
			event("response.output_item.done", map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item})
			output = append(output, item)
		}
		event("response.completed", map[string]any{"type": "response.completed", "response": map[string]any{"id": fmt.Sprintf("resp_%d", n), "status": "completed", "output": output, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1, "total_tokens": 2}}})
	}
}
func codexConfig(p *probe, sandbox string, memory bool) {
	if os.Getenv("QUAL_MUTANT") == "memory-on" {
		memory = true
	}
	write(p.t, filepath.Join(p.root, "codex", "AGENTS.md"), "QUAL_CODEX_PROFILE\n")
	off := "false"
	if memory {
		off = "true"
	}
	config := fmt.Sprintf(`model="offline-probe"
model_provider="qualification"
cli_auth_credentials_store="file"
approval_policy="never"
sandbox_mode=%q
check_for_update_on_startup=false
[features]
memories=%s
memory_tool=%s
[memories]
generate_memories=%s
use_memories=%s
dedicated_tools=%s
[projects.%q]
trust_level="trusted"
[model_providers.qualification]
name="Synthetic local Responses"
base_url=%q
wire_api="responses"
requires_openai_auth=false
request_max_retries=0
stream_max_retries=0
`, sandbox, off, off, off, off, off, p.work, p.server.URL)
	write(p.t, filepath.Join(p.root, "codex", "config.toml"), config)
	p.managerArgs = []string{"-c", "approval_policy=\"never\"", "-c", fmt.Sprintf("sandbox_mode=%q", sandbox)}
	for _, key := range []string{"features.memories", "features.memory_tool", "memories.generate_memories", "memories.use_memories", "memories.dedicated_tools"} {
		p.managerArgs = append(p.managerArgs, "-c", key+"="+off)
	}
	write(p.t, filepath.Join(p.work, "AGENTS.md"), "QUAL_PROJECT_DISCOVERY\n")
	write(p.t, filepath.Join(p.work, ".codex/config.toml"), "sandbox_mode='danger-full-access'\napproval_policy='never'\n[features]\nmemories=true\nmemory_tool=true\n[memories]\ngenerate_memories=true\nuse_memories=true\ndedicated_tools=true\n")

}

// Claim: compare the same native shell write under manager workspace-write
// and read-only despite a repository config trying to grant writes. A failed
// workspace-write control remains unknown. Tool results and effects are retained.
// This tests shell filesystem enforcement, not every tool or permission atom.
func TestCodexShellDenial(t *testing.T) {
	if os.Getenv("QUAL_CODEX_BINARY") == "" {
		t.Skip("explicit exact-digest vendor opt-in unset; no qualification")
	}
	var positive, forbidden, denial bool
	var last *probe
	for _, sandbox := range []string{"workspace-write", "read-only"} {
		t.Run(sandbox, func(t *testing.T) {
			p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
			last = p
			target := filepath.Join(p.work, "effect.txt")
			codexEndpoint(p, "printf QUAL_CODEX_WRITE > "+quote(target))
			codexConfig(p, sandbox, false)
			write(t, filepath.Join(p.work, ".codex", "config.toml"), "sandbox_mode='danger-full-access'\napproval_policy='never'\n")
			out, code := p.nativeRun("turn", codexArgs(p, "exec", "--json", "--skip-git-repo-check", "QUAL_CODEX_FIXED_TOOL")...)
			p.save("requests.json", mustJSON(t, p.requests))
			if code != 0 {
				t.Fatalf("incomplete native turn exit=%d: %s", code, p.sanitize(out))
			}
			b, _ := os.ReadFile(target)
			actual := string(b) == "QUAL_CODEX_WRITE"
			raw := functionResults(mainRequests(p))
			completed := len(mainRequests(p)) > 1 && raw != "[]"
			if sandbox == "workspace-write" {
				positive = actual && completed
			} else {
				forbidden = actual
				denial = completed && filesystemDenial(raw)
			}
		})
	}
	if last != nil {
		last.save("verdict.json", mustJSON(t, observation{"shell-write-read-only", verdict(positive, denial, forbidden), positive, denial, forbidden, "exec startup only; no arbitrary tool, approval escalation, resume or child claim"}))
	}
}

// Claim: seeded memory prompt reads are compared with all five manager memory
// flags off. A missing positive canary is unknown, even if every native turn
// finishes cleanly. This is not an automatic-memory-write or lifecycle probe.
func TestCodexMemoryRead(t *testing.T) {
	if os.Getenv("QUAL_CODEX_BINARY") == "" {
		t.Skip("explicit exact-digest vendor opt-in unset; no qualification")
	}
	var positive, forbidden bool
	var last *probe
	for _, on := range []bool{true, false} {
		t.Run(fmt.Sprintf("enabled-%t", on), func(t *testing.T) {
			p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
			last = p
			codexEndpoint(p, "")
			codexConfig(p, "read-only", on)
			write(t, filepath.Join(p.root, "codex", "memories", "memory_summary.md"), "QUAL_CODEX_MEMORY_CANARY\n")
			write(t, filepath.Join(p.root, "codex", "memories", "MEMORY.md"), "QUAL_CODEX_MEMORY_CANARY\n")
			out, code := p.nativeRun("turn", codexArgs(p, "exec", "--json", "--skip-git-repo-check", "QUAL_CODEX_MEMORY")...)
			p.save("requests.json", mustJSON(t, p.requests))
			if code != 0 || len(mainRequests(p)) == 0 {
				t.Fatalf("incomplete memory turn: %d %s", code, p.sanitize(out))
			}
			contains(t, mustJSON(t, mainRequests(p)[0]), "QUAL_CODEX_PROFILE", true)
			contains(t, mustJSON(t, mainRequests(p)[0]), "QUAL_PROJECT_DISCOVERY", true)
			actual := strings.Contains(mustJSON(t, mainRequests(p)[0]), "QUAL_CODEX_MEMORY_CANARY")
			if on {
				positive = actual
			} else {
				forbidden = actual
			}
		})
	}
	if !positive {
		t.Fatal("memory positive control unavailable")
	}
	if forbidden {
		t.Fatal("memory canary escaped off gate")
	}
	if last != nil {
		last.save("verdict.json", mustJSON(t, observation{"memory-prompt-read", verdict(positive, !forbidden, forbidden), positive, !forbidden, forbidden, "seeded summary probe only; no positive means unknown; auto-writes/resume/child unqualified"}))
	}
}

// Claim: an actual stdio MCP call is paired with disabled_tools for that same
// tool under approval=never/read-only. Transport startup alone never proves
// permission behavior. Qualification is limited to the explicit tool deny;
// approval/sandbox alone are not interpreted as MCP default deny.
func TestCodexMCPDenial(t *testing.T) { codexMCPPair(t, false) }

// Claim: an MCP default prompt ceiling rejects an actual read-only annotated
// call with approval=never, beside an automatic-mode successful call. This is
// one MCP policy atom, not a manager-wide default-deny tool contract.
func TestCodexMCPDefaultPrompt(t *testing.T) { codexMCPPair(t, true) }
func codexMCPPair(t *testing.T, defaultPrompt bool) {
	if os.Getenv("QUAL_CODEX_BINARY") == "" {
		t.Skip("explicit exact-digest vendor opt-in unset; no qualification")
	}
	var positive, forbidden, denial bool
	var last *probe
	for _, disabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("disabled-%t", disabled), func(t *testing.T) {
			p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
			last = p
			codexEndpoint(p, "@mcp")
			codexConfig(p, "read-only", false)
			script := filepath.Join(p.root, "mcp.py")
			write(t, script, read(t, "testdata/mcp.py"))
			marker := filepath.Join(p.root, "mcp.witness")
			file := filepath.Join(p.root, "codex", "config.toml")
			extra := fmt.Sprintf("\n[mcp_servers.approved]\ncommand=\"/usr/bin/python3\"\nargs=[%q,%q]\n", script, marker)
			if defaultPrompt {
				mode := "auto"
				if disabled && os.Getenv("QUAL_MUTANT") != "mcp-prompt-auto" {
					mode = "prompt"
				}
				extra += fmt.Sprintf("default_tools_approval_mode=%q\n", mode)
			}
			if disabled && !defaultPrompt && os.Getenv("QUAL_MUTANT") != "mcp-enabled" {
				extra += "disabled_tools=[\"sentinel\"]\n"
			}
			write(t, file, read(t, file)+extra)
			out, code := p.nativeRun("turn", codexArgs(p, "exec", "--json", "--skip-git-repo-check", "QUAL_CODEX_MCP")...)
			p.save("requests.json", mustJSON(t, p.requests))
			if code != 0 {
				t.Fatalf("incomplete MCP turn: %d %s", code, p.sanitize(out))
			}
			_, err := os.Stat(marker + ".called")
			actual := err == nil
			raw := functionResults(mainRequests(p))
			if disabled {
				forbidden = actual
				denial = len(mainRequests(p)) > 1 && (strings.Contains(raw, "unknown tool") || strings.Contains(raw, "Unknown tool") || strings.Contains(raw, "not found") || strings.Contains(raw, "unsupported call: mcp__approvedsentinel") || strings.Contains(raw, "MCP tool call requires approval, but approval policy is never"))
			} else {
				positive = actual && strings.Contains(raw, "MCP_TOOL_SENTINEL")
			}
		})
	}
	if !positive {
		t.Fatal("MCP positive control unavailable")
	}
	if forbidden {
		t.Fatal("disabled MCP tool executed")
	}
	if !denial {
		t.Fatal("specific MCP denial missing")
	}
	if last != nil {
		limit := "explicit disabled_tools only; approval=never/read-only positive allows MCP; no default-deny certificate"
		atom := "MCP-explicit-disabled-tool"
		if defaultPrompt {
			atom = "MCP-default-prompt"
			limit = "default_tools_approval_mode prompt rejects one annotated MCP tool under approval never; no all-tool or lifecycle certificate"
		}
		last.save("verdict.json", mustJSON(t, observation{atom, verdict(positive, denial, forbidden), positive, denial, forbidden, limit}))
	}
}

func filesystemDenial(output string) bool {
	return (os.Getenv("QUAL_MUTANT") == "nested-as-policy" || !strings.Contains(output, "sandbox_apply")) && (strings.Contains(output, "Operation not permitted") || strings.Contains(output, "Read-only file system") || strings.Contains(output, "Permission denied"))
}

// Claim: a shell's native filesystem rejection is distinct from an identical
// permission string emitted by failed nested sandbox startup.
func TestFilesystemDenialAttribution(t *testing.T) {
	for _, row := range []struct {
		output string
		want   bool
	}{{"Operation not permitted", true}, {"Read-only file system", true}, {"sandbox-exec: sandbox_apply: Operation not permitted", false}, {"unknown tool", false}, {"", false}} {
		if filesystemDenial(row.output) != row.want {
			t.Fatal("nested startup misattributed to manager denial")
		}
	}
}

func codexArgs(p *probe, args ...string) []string {
	return append(append([]string{}, p.managerArgs...), args...)
}

// Claim: dedicated native memory read/write tools have real effects when on;
// all five off ceilings must withhold that same call despite trusted repository
// flags trying to enable it. Manual writes do not qualify automatic generation.
func TestCodexDedicatedMemory(t *testing.T) {
	if os.Getenv("QUAL_CODEX_BINARY") == "" {
		t.Skip("explicit exact-digest vendor opt-in unset; no qualification")
	}
	for _, atom := range []string{"read", "write"} {
		t.Run(atom, func(t *testing.T) {
			var positive, forbidden, denial bool
			var last *probe
			for _, on := range []bool{true, false} {
				t.Run(fmt.Sprintf("on-%t", on), func(t *testing.T) {
					p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
					last = p
					codexConfig(p, "read-only", on)
					codexEndpoint(p, "@memory-"+atom)
					write(t, filepath.Join(p.root, "codex/memories/MEMORY.md"), "QUAL_CODEX_MEMORY_CANARY")
					_, code := p.nativeRun("turn", codexArgs(p, "exec", "--json", "--skip-git-repo-check", "QUAL_DEDICATED_MEMORY please remember the synthetic qualification note, or read it for this test.")...)
					p.save("requests.json", mustJSON(t, p.requests))
					if code != 0 || len(mainRequests(p)) < 2 {
						t.Fatal("incomplete dedicated memory call")
					}
					raw := functionResults(mainRequests(p))
					actual := strings.Contains(raw, "QUAL_CODEX_MEMORY_CANARY")
					if atom == "write" {
						actual = false
						_ = filepath.WalkDir(filepath.Join(p.root, "codex/memories"), func(path string, d os.DirEntry, err error) error {
							if err == nil && !d.IsDir() && strings.Contains(path, "ad_hoc") {
								b, _ := os.ReadFile(path)
								actual = actual || strings.Contains(string(b), "QUAL_CODEX_MEMORY_WRITE")
							}
							return nil
						})
					}
					if on {
						positive = actual
					} else {
						forbidden = actual
						denial = strings.Contains(raw, "unknown tool") || strings.Contains(raw, "unsupported call") || strings.Contains(raw, "Unknown tool")
					}
				})
			}
			if !positive {
				t.Fatal("dedicated memory positive unavailable")
			}
			if forbidden {
				t.Fatal("memory tool escaped off gate")
			}
			if !denial {
				t.Fatal("dedicated memory specific denial missing")
			}
			if last != nil {
				last.save("verdict.json", mustJSON(t, observation{"dedicated-memory-" + atom, verdict(positive, denial, forbidden), positive, denial, forbidden, "manual memory tool only; trusted project; no automatic-write or child qualification"}))
			}
		})
	}
}

// Claim: a fresh memory-off turn is clean, but a saved assistant response that
// actually followed an enabled seeded-memory read may survive resume with off
// ceilings. This checks retained history separately from fresh memory discovery;
// no child agent is started and no automatic-memory-write claim is made.
func TestCodexMemoryResume(t *testing.T) {
	p := setupBinary(t, "QUAL_CODEX_BINARY", map[string]bool{codexSHA: true})
	codexConfig(p, "read-only", true)
	codexEndpoint(p, "")
	p.prompt = "QUAL_RESUME_INITIAL"
	p.finalText = "QUAL_CODEX_MEMORY_CANARY"
	write(t, filepath.Join(p.root, "codex/memories/memory_summary.md"), "QUAL_CODEX_MEMORY_CANARY")
	_, code := p.nativeRun("initial", codexArgs(p, "exec", "--json", "--skip-git-repo-check", "QUAL_RESUME_INITIAL")...)
	if code != 0 || len(mainRequests(p)) == 0 {
		t.Fatal("resume initial baseline incomplete")
	}
	contains(t, mustJSON(t, mainRequests(p)[0]), "QUAL_CODEX_MEMORY_CANARY", true)
	p.save("initial-requests.json", mustJSON(t, p.requests))
	p.mu.Lock()
	p.requests = nil
	p.mu.Unlock()
	p.prompt = "QUAL_RESUME_OFF"
	p.issued = false
	p.finalText = "QUAL_RESUMED"
	codexConfig(p, "read-only", false)
	_, code = p.nativeRun("resume", codexArgs(p, "exec", "resume", "--last", "--json", "--skip-git-repo-check", "QUAL_RESUME_OFF")...)
	if code != 0 || len(mainRequests(p)) == 0 {
		t.Fatal("native resume did not complete")
	}
	p.save("requests.json", mustJSON(t, p.requests))
	tainted := strings.Contains(mustJSON(t, mainRequests(p)[0]), "QUAL_CODEX_MEMORY_CANARY")
	p.save("verdict.json", mustJSON(t, observation{"tainted-history-resume", verdict(true, !tainted, tainted), true, !tainted, tainted, "actual enabled-memory assistant history resumed with all five off flags; fresh off read is tested separately; history is not new memory discovery"}))
}

func requestHasUser(req map[string]any, prompt string) bool {
	if os.Getenv("QUAL_MUTANT") == "request-attribution" {
		b, _ := json.Marshal(req)
		return strings.Contains(string(b), prompt)
	}
	for _, raw := range toSlice(req["input"]) {
		item, _ := raw.(map[string]any)
		if item["role"] != "user" {
			continue
		}
		for _, c := range toSlice(item["content"]) {
			content, _ := c.(map[string]any)
			if content["text"] == prompt {
				return true
			}
		}
	}
	return false
}
func mainRequests(p *probe) []map[string]any {
	var rows []map[string]any
	// NativeRun has completed before callers inspect these immutable bodies.
	for _, r := range p.requests {
		if requestHasUser(r, p.prompt) {
			rows = append(rows, r)
		}
	}
	return rows
}

// Claim: a background memory job and a tool-schema mention of the marker are
// not main-turn requests; only the exact synthetic user prompt selects one.
func TestRequestAttribution(t *testing.T) {
	for _, row := range []struct {
		req  map[string]any
		want bool
	}{
		{map[string]any{"input": []any{map[string]any{"role": "user", "content": []any{map[string]any{"text": "QUAL_TARGET"}}}}}, true},
		{map[string]any{"input": []any{map[string]any{"role": "user", "content": []any{map[string]any{"text": "memory consolidation quotes QUAL_TARGET"}}}}}, false},
		{map[string]any{"tools": []any{map[string]any{"description": "QUAL_TARGET"}}}, false},
	} {
		if requestHasUser(row.req, "QUAL_TARGET") != row.want {
			t.Fatal("background request misattributed to main turn")
		}
	}
}
