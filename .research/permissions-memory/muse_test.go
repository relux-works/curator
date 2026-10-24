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

// Claim: attempt a credential-free local Responses turn with a fabricated stdin
// key and explicit model. A failed handshake or timeout is unknown, never a
// permission/memory verdict. No real login or remote provider is contacted.
func museEndpoint(p *probe, toolName string, args map[string]any) {
	called := false
	p.nativeInput = "synthetic-local-fixture-not-a-real-key\n"
	p.codexReply = func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/page" {
			p.mu.Lock()
			p.webHits++
			p.mu.Unlock()
			io.WriteString(w, "QUAL_MUSE_PAGE")
			return
		}
		if r.Method == "GET" {
			p.save("catalog-route.json", mustJSON(p.t, map[string]any{"method": r.Method, "path": r.URL.Path}))
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"object":"list","data":[{"id":"fixture-model","object":"model","created":0,"owned_by":"fixture"}]}`)
			return
		}
		defer r.Body.Close()
		var req map[string]any
		if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
			http.Error(w, "bad fixture request", 400)
			return
		}
		p.mu.Lock()
		p.requests = append(p.requests, req)
		n := len(p.requests)
		p.mu.Unlock()
		event := func(kind string, value any) {
			b, _ := json.Marshal(value)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, b)
		}
		response := map[string]any{"id": fmt.Sprintf("resp_%d", n), "object": "response", "created_at": 1780000000, "status": "in_progress", "error": nil, "incomplete_details": nil, "instructions": nil, "max_output_tokens": 1024, "model": "fixture-model", "output": []any{}, "parallel_tool_calls": false, "previous_response_id": nil, "reasoning": map[string]any{"effort": nil, "summary": nil}, "store": false, "temperature": 1, "text": map[string]any{"format": map[string]any{"type": "text"}}, "tool_choice": "auto", "tools": []any{}, "top_p": 1, "truncation": "disabled", "usage": nil, "user": nil, "metadata": map[string]any{}}
		w.Header().Set("Content-Type", "text/event-stream")
		event("response.created", map[string]any{"type": "response.created", "sequence_number": 0, "response": response})
		item := map[string]any{"id": "msg_qual", "type": "message", "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": "QUAL_MUSE_LOCAL_OK", "annotations": []any{}}}}
		p.mu.Lock()
		call := false
		for _, raw := range toSlice(req["tools"]) {
			group, _ := raw.(map[string]any)
			if group["name"] == "muse" {
				call = toolName != "" && !called
			}
		}
		if call {
			called = true
		}
		p.mu.Unlock()
		if call {
			encoded, _ := json.Marshal(args)
			item = map[string]any{"id": "fc_qual", "type": "function_call", "call_id": "call_qual", "name": toolName, "namespace": "muse", "arguments": string(encoded), "status": "completed"}
		}
		event("response.output_item.added", map[string]any{"type": "response.output_item.added", "sequence_number": 1, "output_index": 0, "item": item})
		if !call {
			event("response.output_text.delta", map[string]any{"type": "response.output_text.delta", "sequence_number": 2, "output_index": 0, "content_index": 0, "item_id": "msg_qual", "delta": "QUAL_MUSE_LOCAL_OK"})
		}
		event("response.output_item.done", map[string]any{"type": "response.output_item.done", "sequence_number": 3, "output_index": 0, "item": item})
		response["status"] = "completed"
		response["output"] = []any{item}
		response["usage"] = map[string]any{"input_tokens": 1, "output_tokens": 1, "total_tokens": 2, "input_tokens_details": map[string]any{"cached_tokens": 0}, "output_tokens_details": map[string]any{"reasoning_tokens": 0}}
		event("response.completed", map[string]any{"type": "response.completed", "sequence_number": 4, "response": response})
	}
}
func functionResults(requests []map[string]any) string {
	var results []any
	if os.Getenv("QUAL_MUTANT") == "result-attribution" {
		raw, _ := json.Marshal(requests)
		return string(raw)
	}
	for _, req := range requests {
		for _, raw := range toSlice(req["input"]) {
			item, _ := raw.(map[string]any)
			if item["type"] == "function_call_output" {
				results = append(results, item["output"])
			}
		}
	}
	b, _ := json.Marshal(results)
	return string(b)
}
func toSlice(v any) []any { rows, _ := v.([]any); return rows }
func museArgs(p *probe, extra ...string) []string {
	args := []string{"exec", "--json", "--provider", "meta", "--model", "fixture-model", "--base-url", p.server.URL, "--api-key-stdin", "--approval-mode", "never", "--approval-judge", "off", "--max-model-steps", "3", "--disable-reminders"}
	args = append(args, extra...)
	return append(args, "QUAL_MUSE_LOCAL")
}

// Claim: a fabricated key and exact local Responses fixture complete a real
// model turn. This provides a transport baseline only, not tool denial evidence.
func TestMuseLocalBaseline(t *testing.T) {
	p := setupBinary(t, "QUAL_MUSE_BINARY", map[string]bool{museSHA: true})
	museEndpoint(p, "", nil)
	t.Cleanup(func() { p.mu.Lock(); defer p.mu.Unlock(); p.save("requests.json", mustJSON(t, p.requests)) })
	out, code := p.nativeRun("local-turn", museArgs(p)...)
	if code != 0 || !strings.Contains(out, `"terminal":"completed"`) || !strings.Contains(out, "QUAL_MUSE_LOCAL_OK") {
		t.Fatal("incomplete local model baseline")
	}
	p.save("baseline.json", mustJSON(t, map[string]any{"exit": code, "requests": len(p.requests), "qualification": "local transport only; no permission or memory claim"}))
}

// Claim: pinned native help exposes the configuration entry points without
// reading any real configuration or inferring enforcement from help text.
func TestMuseConfigHelp(t *testing.T) {
	p := setupBinary(t, "QUAL_MUSE_BINARY", map[string]bool{museSHA: true})
	p.nativeRun("config-help", "config", "--help")
	p.nativeRun("init-help", "init", "--help")
	p.nativeRun("config-status", "config", "status")
	p.nativeRun("init-dry", "init", "--dry-run")
	p.nativeRun("foreign-profile", "exec", "--json", "--provider", "echo", "--permission-profile", "qualification-unknown", "QUAL_MUSE_PROFILE")
}

// Claim: native Write/Edit and shell flags are paired with the same exact tool
// and explicit approval grant; only real effects plus specific dispatcher
// refusal qualify a bounded flag. No general default-deny/profile claim follows.
func TestMuseToolDenials(t *testing.T) {
	if os.Getenv("QUAL_MUSE_BINARY") == "" {
		t.Skip("explicit vendor opt-in unset")
	}
	for _, atom := range []string{"write_file", "edit_file", "bash", "legacy-bash", "web_fetch", "add_memory", "read_memory"} {
		t.Run(atom, func(t *testing.T) {
			var positive, forbidden, denial bool
			var last *probe
			for _, disabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("disabled-%t", disabled), func(t *testing.T) {
					p := setupBinary(t, "QUAL_MUSE_BINARY", map[string]bool{museSHA: true})
					last = p
					target := filepath.Join(p.work, "effect.txt")
					args := map[string]any{}
					flag := "--disable-write"
					switch atom {
					case "write_file":
						args = map[string]any{"path": target, "content": "QUAL_MUSE_EFFECT"}
					case "edit_file":
						write(t, target, "QUAL_OLD")
						args = map[string]any{"path": target, "find": "QUAL_OLD", "replace": "QUAL_MUSE_EFFECT"}
					case "bash", "legacy-bash":
						flag = "--disable-shell"
						args = map[string]any{"command": "printf QUAL_MUSE_EFFECT > " + quote(target), "description": "Write synthetic marker", "yield_time_ms": 1000, "timeout_ms": 2000, "login": false}
					case "web_fetch":
						flag = "--disable-web-tools"
						args = map[string]any{"url": p.server.URL + "/page"}
					case "read_memory":
						target = filepath.Join(p.work, ".agents", "memory", "qualification.md")
						write(t, target, "QUAL_MUSE_MEMORY_READ")
						args = map[string]any{"scope": "project", "path": "qualification.md"}
					case "add_memory":
						args = map[string]any{"scope": "project", "path": "qualification.md", "content": "QUAL_MUSE_MEMORY_WRITE"}
					}
					toolName := atom
					if atom == "legacy-bash" {
						toolName = "shell"
						args = map[string]any{"command": "printf QUAL_MUSE_EFFECT > " + quote(target)}
					}
					museEndpoint(p, toolName, args)
					extra := []string{"--disable-approval"}
					if atom == "legacy-bash" {
						extra = append(extra, "--enable-shell-tool", "--disable-sandbox")
					}
					if disabled && os.Getenv("QUAL_MUTANT") != "muse-tool-enabled" {
						extra = append(extra, flag)
					}
					_, code := p.nativeRun("turn", museArgs(p, extra...)...)
					p.save("requests.json", mustJSON(t, p.requests))
					if code != 0 {
						t.Fatalf("incomplete Muse turn: %d", code)
					}
					p.mu.Lock()
					raw := functionResults(p.requests)
					p.mu.Unlock()
					b, _ := os.ReadFile(target)
					actual := string(b) == "QUAL_MUSE_EFFECT"
					changed := false
					if atom == "write_file" || atom == "bash" || atom == "legacy-bash" {
						_, err := os.Stat(target)
						changed = !os.IsNotExist(err)
					}
					if atom == "edit_file" {
						changed = string(b) != "QUAL_OLD"
					}
					if atom == "read_memory" {
						actual = strings.Contains(raw, "QUAL_MUSE_MEMORY_READ")
						changed = string(b) != "QUAL_MUSE_MEMORY_READ"
					}
					if atom == "web_fetch" {
						actual = strings.Contains(raw, "QUAL_MUSE_PAGE")
						p.mu.Lock()
						changed = p.webHits > 0
						p.mu.Unlock()
					}
					if atom == "add_memory" {
						path := filepath.Join(p.work, ".agents", "memory", "qualification.md")
						data, _ := os.ReadFile(path)
						actual = string(data) == "QUAL_MUSE_MEMORY_WRITE\n" || string(data) == "QUAL_MUSE_MEMORY_WRITE"
						_, err := os.Stat(path)
						changed = !os.IsNotExist(err)
					}
					specific := strings.Contains(raw, "tool unavailable: unknown tool `"+toolName+"`")

					if disabled {
						forbidden = actual || changed
						denial = specific
					} else {
						positive = actual
					}
				})
			}
			if last != nil {
				limit := "explicit startup flag with approval disabled in both siblings; manual memory only, no automatic memory-off"
				if atom == "legacy-bash" {
					limit = "legacy shell under --enable-shell-tool --disable-sandbox; outer privacy fence intact; --disable-shell gate only, not managed-shell sandbox"
				}
				last.save("verdict.json", mustJSON(t, observation{"Muse-" + atom, verdict(positive, denial, forbidden), positive, denial, forbidden, limit}))
			}
			if atom == "read_memory" && (!positive || !forbidden) {
				t.Fatal("disable-write memory-read witness unavailable")
			}
			if os.Getenv("QUAL_EXPECT_DENY") == "1" && (atom == "write_file" || atom == "edit_file" || atom == "web_fetch" || atom == "add_memory" || atom == "legacy-bash") {
				if !positive {
					t.Fatal("Muse tool positive unavailable")
				}
				if forbidden {
					t.Fatalf("Muse disabled %s executed", atom)
				}
				if !denial {
					t.Fatal("Muse tool specific denial missing")
				}
			}

		})
	}
}

// Claim: the pinned settings parser either accepts this dedicated-memory
// candidate or specifically rejects it before any provider request. Rejection
// means this spelling is unavailable, not that no memory control can exist.
func TestMuseMemoryConfigCandidate(t *testing.T) {
	p := setupBinary(t, "QUAL_MUSE_BINARY", map[string]bool{museSHA: true})
	museEndpoint(p, "", nil)
	out, code := p.nativeRun("memory-config", museArgs(p, "--config", "memory.enabled=false")...)
	p.save("candidate.json", mustJSON(t, map[string]any{"exit": code, "accepted": code == 0, "providerRequests": len(p.requests), "interpretation": "bounded parser candidate only; no dedicated-memory-off qualification"}))
	if code != 0 && (!strings.Contains(out, "unknown") && !strings.Contains(out, "Unknown")) {
		t.Fatal("unrelated memory candidate rejection")
	}
}

// Claim: an actual memory-tool read seeds a logged session; same-ID headless
// re-entry with write disabled and the source removed either exposes retained
// history or records an unsupported/incomplete entry. It never qualifies full
// resume or automatic-memory suppression from a fresh-session absence.
func TestMuseMemoryReentry(t *testing.T) {
	p := setupBinary(t, "QUAL_MUSE_BINARY", map[string]bool{museSHA: true})
	target := filepath.Join(p.work, ".agents/memory/qualification.md")
	write(t, target, "QUAL_MUSE_MEMORY_READ")
	museEndpoint(p, "read_memory", map[string]any{"scope": "project", "path": "qualification.md"})
	session := "11111111-1111-4111-8111-111111111111"
	_, code := p.nativeRun("initial", museArgs(p, "--disable-approval", "--session-id", session)...)
	if code != 0 || !strings.Contains(functionResults(p.requests), "QUAL_MUSE_MEMORY_READ") {
		t.Fatal("Muse memory reentry positive unavailable")
	}
	p.save("initial-requests.json", mustJSON(t, p.requests))
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.requests = nil
	p.mu.Unlock()
	museEndpoint(p, "", nil)
	_, code = p.nativeRun("reentry", museArgs(p, "--disable-approval", "--disable-write", "--session-id", session)...)
	p.save("requests.json", mustJSON(t, p.requests))
	tainted := strings.Contains(mustJSON(t, p.requests), "QUAL_MUSE_MEMORY_READ")
	result := "unknown"
	if code == 0 && tainted {
		result = "refused"
	}
	p.save("verdict.json", mustJSON(t, observation{"Muse-same-id-memory-reentry", result, true, false, tainted, "source removed after real tool read; same-ID exec only; canonical resume/TUI and child entries unqualified"}))
}
