package claudequalification

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

const projectMarker = "QUAL_NATIVE_PROJECT_INPUT_9821"
const memoryMarker = "QUAL_AUTOMATIC_MEMORY_INPUT_9821"

func (p *probe) nativeProject() {
	write(p.t, filepath.Join(p.work, "CLAUDE.md"), projectMarker+"\n")
}

func (p *probe) assertContext(r run) {
	p.t.Helper()
	p.assertRoot(r)
	contains(p.t, payload(p.t, r), projectMarker, true)
}

func absent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("forbidden file effect: %v", err)
	}
}

// Identity pins a release and digest independently; one release never inherits
// another release's observations. All vendor commands use the same sandbox.
func TestIdentity(t *testing.T) {
	p := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := append([]string{"-f", p.policy, "/usr/bin/env", "-i"}, p.env...)
	args = append(args, p.binary, "--version")
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", args...)
	cmd.Dir, cmd.Env = p.work, []string{"PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	want := os.Getenv("CLAUDE_QUAL_VERSION") + " (Claude Code)"
	if err != nil || strings.TrimSpace(string(out)) != want {
		t.Fatalf("identity: %v %s", err, p.sanitize(string(out)))
	}
	p.save("identity.json", mustJSON(t, map[string]any{"version": want, "sha256": binaryDigests[os.Getenv("CLAUDE_QUAL_VERSION")], "platform": "darwin/arm64"}))
}

// Each permission atom pairs a real granted operation with a manager deny
// challenged by both project and local allow rules. WebFetch is observational:
// a loopback refusal cannot qualify its positive operation or the deny atom.
func TestPermissions(t *testing.T) {
	for _, form := range []string{"fresh", "resume", "child"} {
		t.Run(form, func(t *testing.T) {
			for _, atom := range []string{"Bash", "Edit", "Write", "WebFetch", "MCP"} {
				for _, deny := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/deny-%t", atom, deny), func(t *testing.T) {
						p := setup(t)
						p.nativeProject()
						file := filepath.Join(p.work, "effect.txt")
						name, rule := atom, atom
						input := map[string]any{}
						args := []string{}
						switch atom {
						case "Bash":
							input = map[string]any{"command": "printf 'QUAL_EFFECT_9821\\n' > " + quote(file), "description": "Synthetic local file witness"}
						case "Edit":
							write(t, file, "QUAL_BEFORE_9821\n")
							// Native Edit requires an actual preceding Read, even when granted.
							p.actions = append(p.actions, tool("Read", map[string]any{"file_path": file}))
							input = map[string]any{"file_path": file, "old_string": "QUAL_BEFORE_9821", "new_string": "QUAL_EFFECT_9821"}
						case "Write":
							input = map[string]any{"file_path": file, "content": "QUAL_EFFECT_9821\n"}
						case "WebFetch":
							input = map[string]any{"url": p.server.URL + "/ordinary", "prompt": "Return the fixture text"}
						case "MCP":
							name, rule = "mcp__approved-probe__sentinel", "mcp__*"
							config := filepath.Join(p.home, "managed-mcp.json")
							write(t, config, mustJSON(t, map[string]any{"mcpServers": map[string]any{"approved-probe": p.mcp("approved")}}))
							args = append(args, "--strict-mcp-config", "--mcp-config", config, "--tools", "")
						}
						permissions := map[string]any{"defaultMode": "dontAsk"}
						if deny && os.Getenv("CLAUDE_QUAL_MUTANT") != "deny-"+atom {
							permissions["deny"] = []string{rule}
						}
						p.settings("user", map[string]any{"model": "claude-sonnet-4-6", "permissions": permissions})
						for _, source := range []string{"project", "local"} {
							p.settings(source, map[string]any{"permissions": map[string]any{"allow": []string{name}}})
						}
						p.actions = append(p.actions, tool(name, input))
						args = append(args, "--allowedTools", "Read,"+name)
						if form == "resume" {
							actions := p.actions
							p.actions = nil
							first := p.run("origin", false, args...)
							p.assertContext(first)
							id, ok := first.Init["session_id"].(string)
							if !ok || id == "" {
								t.Fatal("missing persisted session")
							}
							p.actions = actions
							p.actionStart = len(p.requests)
							args = append(args, "--resume", id)
						}
						if form == "child" {
							p.planChild(p.actions)
							args = append(args, "--tools", "Read,Task,Bash,Edit,Write,WebFetch", "--allowedTools", "Read,Task,"+name)
						}
						r := p.run("turn", false, args...)
						p.assertContext(r)
						if form == "child" {
							p.assertChild(r)
						}
						if atom == "WebFetch" {
							p.save("verdict.json", mustJSON(t, map[string]any{"atom": atom, "manager_deny": deny, "permission_denials": r.Result["permission_denials"], "qualification": "unknown: loopback operation has no successful WebFetch positive"}))
							if deny {
								contains(t, payload(t, r), "No such tool available: "+name, true)
							}
							return
						}
						contains(t, payload(t, r), "No such tool available: "+name, deny)
						switch atom {
						case "Edit":
							want := "QUAL_EFFECT_9821\n"
							if deny {
								want = "QUAL_BEFORE_9821\n"
							}
							if read(t, file) != want {
								t.Fatal("Edit effect disagrees with manager deny")
							}
						case "Bash", "Write":
							if deny {
								absent(t, file)
							} else if read(t, file) != "QUAL_EFFECT_9821\n" {
								t.Fatal("missing granted file effect")
							}
						case "MCP":
							contains(t, payload(t, r), "MCP_TOOL_SENTINEL", !deny)
							contains(t, read(t, filepath.Join(p.root, "approved.witness")), "MCP_STARTED", true)
						}
					})
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

// This counterexample measures whether native project permissions still apply:
// an otherwise granted Read is denied only when the repository adds its deny.
// It deliberately records a refusal of the stronger source-isolation contract.
func TestProjectPermissions(t *testing.T) {
	for _, source := range []string{"none", "project", "local"} {
		t.Run(source, func(t *testing.T) {
			p := setup(t)
			p.nativeProject()
			file := filepath.Join(p.work, "ordinary.txt")
			write(t, file, "QUAL_ORDINARY_READ_9821\n")
			if source != "none" {
				p.settings(source, map[string]any{"permissions": map[string]any{"deny": []string{"Read(./ordinary.txt)"}}})
			}
			p.actions = []map[string]any{tool("Read", map[string]any{"file_path": file})}
			r := p.run("turn", false, "--allowedTools", "Read", "--tools", "Read")
			p.assertContext(r)
			contains(t, payload(t, r), "QUAL_ORDINARY_READ_9821", source == "none")
			contains(t, mustJSON(t, r.Result["permission_denials"]), "Read", source != "none")
		})
	}
}

// Empty and Read-only builtin catalogs reject injected Bash/Edit/Write/WebFetch
// despite repository allow rules. The Read-only sibling must still read data.
func TestBuiltinCeiling(t *testing.T) {
	for _, catalog := range []string{"", "Read"} {
		name := catalog
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			p := setup(t)
			p.nativeProject()
			file := filepath.Join(p.work, "ceiling.txt")
			ordinary := filepath.Join(p.work, "ordinary.txt")
			shell := filepath.Join(p.work, "shell.txt")
			write(t, ordinary, "QUAL_CEILING_READ_9821\n")
			for _, source := range []string{"project", "local"} {
				p.settings(source, map[string]any{"permissions": map[string]any{"allow": []string{"Write", "Bash", "Edit", "Read", "WebFetch"}}})
			}
			p.actions = []map[string]any{
				tool("Read", map[string]any{"file_path": ordinary}),
				tool("Write", map[string]any{"file_path": file, "content": "QUAL_CEILING_EFFECT_9821"}),
				tool("Bash", map[string]any{"command": "printf 'QUAL_SHELL_EFFECT_9821' > " + quote(shell), "description": "Synthetic file witness"}),
				tool("Edit", map[string]any{"file_path": ordinary, "old_string": "QUAL_CEILING_READ_9821", "new_string": "QUAL_CEILING_EDIT_9821"}),
				tool("WebFetch", map[string]any{"url": p.server.URL + "/ordinary", "prompt": "Return the fixture text"}),
			}
			args := []string{"--tools", catalog}
			if os.Getenv("CLAUDE_QUAL_MUTANT") == "ceiling" {
				args = nil
			}
			r := p.run("turn", false, args...)
			p.assertContext(r)
			denied := []string{"Bash", "Edit", "Write", "WebFetch"}
			if catalog == "" {
				denied = append(denied, "Read")
			}
			for _, atom := range denied {
				if !strings.Contains(payload(t, r), "No such tool available: "+atom) {
					t.Fatalf("builtin catalog ceiling bypassed: %s", atom)
				}
			}
			contains(t, toolResults(t, r), "QUAL_CEILING_READ_9821", catalog == "Read")
			absent(t, file)
			absent(t, shell)
			if read(t, ordinary) != "QUAL_CEILING_READ_9821\n" {
				t.Fatal("forbidden Edit through builtin ceiling")
			}
		})
	}
}

func (p *probe) memory(off bool) string {
	for i, v := range p.env {
		if strings.HasPrefix(v, "CLAUDE_CODE_DISABLE_AUTO_MEMORY=") {
			value := "0"
			if off && os.Getenv("CLAUDE_QUAL_MUTANT") != "memory" {
				value = "1"
			}
			p.env[i] = "CLAUDE_CODE_DISABLE_AUTO_MEMORY=" + value
		}
	}
	p.settings("user", map[string]any{"model": "claude-sonnet-4-6", "autoMemoryEnabled": true})
	// This is the native repository-keyed auto-memory location, not CLAUDE.md.
	dir := filepath.Join(p.home, "projects", regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(p.work, "-"), "memory")
	write(p.t, filepath.Join(dir, "MEMORY.md"), memoryMarker+"\n")
	return dir
}

// Memory reads compare enabled and disabled states at fresh/resume/child entry.
// A seeded automatic-memory marker must actually reach model input when on.
// Stable files without an automatic-write positive do not certify write-off.
func TestMemory(t *testing.T) {
	for _, off := range []bool{false, true} {
		for _, form := range []string{"fresh", "resume", "child"} {
			t.Run(fmt.Sprintf("off-%t/%s", off, form), func(t *testing.T) {
				p := setup(t)
				p.nativeProject()
				dir := p.memory(off)
				args := []string{"--tools", "Read,Task", "--allowedTools", "Read,Task"}
				if form == "child" {
					write(t, filepath.Join(p.home, "agents", "probe-agent.md"), "---\nname: probe-agent\ndescription: Synthetic memory probe\n---\nQUAL_CHILD_INPUT_9821\n")
					p.actions = []map[string]any{tool("Task", map[string]any{"description": "Synthetic memory probe", "prompt": "Return the fixed offline response", "subagent_type": "probe-agent"})}
				}
				r := p.run("first", false, args...)
				p.assertContext(r)
				contains(t, payload(t, r), memoryMarker, !off)
				if form == "resume" {
					id, ok := r.Init["session_id"].(string)
					if !ok || id == "" {
						t.Fatal("missing persisted session")
					}
					r = p.run("resume", false, append(args, "--resume", id)...)
					p.assertContext(r)
					contains(t, payload(t, r), memoryMarker, !off)
				}
				if form == "child" {
					if len(r.Requests) < 3 {
						t.Fatal("missing actual child request")
					}
					child := ""
					for _, request := range r.Requests {
						if candidate := mustJSON(t, request); strings.Contains(candidate, "QUAL_CHILD_INPUT_9821") {
							child = candidate
							break
						}
					}
					contains(t, child, "QUAL_CHILD_INPUT_9821", true)
					contains(t, child, memoryMarker, !off)
					contains(t, mustJSON(t, r.Result["subagent_stats"]), `"completed": 1`, true)
				}
				if read(t, filepath.Join(dir, "MEMORY.md")) != memoryMarker+"\n" {
					t.Fatal("seed memory mutated")
				}
				entries, err := os.ReadDir(dir)
				if err != nil || len(entries) != 1 {
					t.Fatal("unexpected memory files")
				}
				p.save("verdict.json", mustJSON(t, map[string]any{"read_marker_expected": !off, "write_qualification": "unknown: deterministic turn did not activate an automatic write positive"}))
			})
		}
	}
}

// Child routing identifies the actual native child input instead of assuming
// request order; foreground and background Task dispatch can interleave.
func (p *probe) planChild(actions []map[string]any) {
	write(p.t, filepath.Join(p.home, "agents", "probe-agent.md"), "---\nname: probe-agent\ndescription: Synthetic permission probe\n---\nQUAL_CHILD_INPUT_9821\n")
	p.actions = nil
	sentTask := false
	next := 0
	p.actionFor = func(request map[string]any, _ int) map[string]any {
		p.mu.Lock()
		defer p.mu.Unlock()
		if strings.Contains(mustJSON(p.t, request["system"]), "QUAL_CHILD_INPUT_9821") {
			if next < len(actions) {
				action := actions[next]
				next++
				return action
			}
		} else if !sentTask {
			sentTask = true
			return tool("Task", map[string]any{"description": "Synthetic permission probe", "prompt": "Return the fixed offline response", "subagent_type": "probe-agent"})
		}
		return map[string]any{"type": "text", "text": "OFFLINE_OK"}
	}
}
func (p *probe) assertChild(r run) {
	p.t.Helper()
	found := false
	for _, request := range r.Requests {
		if strings.Contains(mustJSON(p.t, request["system"]), "QUAL_CHILD_INPUT_9821") {
			found = true
		}
	}
	if !found {
		p.t.Fatal("missing actual child request")
	}
	contains(p.t, mustJSON(p.t, r.Result["subagent_stats"]), `"completed": 1`, true)
}

// A fixed model response can still direct an allowed Write into the auto-memory
// directory with the off flag set. This measures the native Write effect, not a
// spontaneous automatic writer; the latter remains unqualified in this replay.
func TestMemoryDirectedWrite(t *testing.T) {
	for _, form := range []string{"fresh", "resume", "child"} {
		for _, off := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/off-%t", form, off), func(t *testing.T) {
				p := setup(t)
				p.nativeProject()
				dir := p.memory(off)
				args := []string{"--tools", "Write,Task", "--allowedTools", "Write,Task"}
				if form == "resume" {
					first := p.run("origin", false, args...)
					p.assertContext(first)
					id, ok := first.Init["session_id"].(string)
					if !ok || id == "" {
						t.Fatal("missing persisted session")
					}
					args = append(args, "--resume", id)
					p.actionStart = len(p.requests)
				}
				path := filepath.Join(dir, "directed.md")
				p.actions = []map[string]any{tool("Write", map[string]any{"file_path": path, "content": "QUAL_MEMORY_WRITE_EFFECT_9821\n"})}
				if form == "child" {
					p.planChild(p.actions)
				}
				r := p.run("turn", false, args...)
				p.assertContext(r)
				if form == "child" {
					p.assertChild(r)
				}
				if read(t, path) != "QUAL_MEMORY_WRITE_EFFECT_9821\n" {
					t.Fatal("missing directed memory Write positive")
				}
				p.save("verdict.json", mustJSON(t, map[string]any{"off": off, "directed_write": true, "automatic_write": "unknown: directed model tool call is not a spontaneous writer activation"}))
			})
		}
	}
}

// Turning memory off on a resume does not scrub memory text already persisted
// in an enabled-origin conversation. This is a separate unsafe-origin witness.
func TestMemoryResumeHistory(t *testing.T) {
	p := setup(t)
	p.nativeProject()
	p.memory(false)
	first := p.run("origin", false, "--tools", "")
	p.assertContext(first)
	contains(t, payload(t, first), memoryMarker, true)
	id, ok := first.Init["session_id"].(string)
	if !ok || id == "" {
		t.Fatal("missing persisted session")
	}
	for i, v := range p.env {
		if strings.HasPrefix(v, "CLAUDE_CODE_DISABLE_AUTO_MEMORY=") {
			p.env[i] = "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1"
		}
	}
	resumed := p.run("resume", false, "--tools", "", "--resume", id)
	p.assertContext(resumed)
	contains(t, payload(t, resumed), memoryMarker, true)
}

func toolResults(t *testing.T, r run) string {
	t.Helper()
	results := []any{}
	for _, request := range r.Requests {
		messages, _ := request["messages"].([]any)
		for _, value := range messages {
			message, _ := value.(map[string]any)
			blocks, _ := message["content"].([]any)
			for _, value := range blocks {
				block, _ := value.(map[string]any)
				if block["type"] == "tool_result" {
					results = append(results, block)
				}
			}
		}
	}
	return mustJSON(t, results)
}
