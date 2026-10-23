package codexexclusion

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/projectadapter"
	"github.com/relux-works/curator/internal/stateread"
)

// This is a bounded research candidate, never product admission authority. The
// caller names all inputs for the fixture. Only the existing stdio parser and
// manager tool grants provide dotted ceilings. It claims neither complete
// Codex discovery nor atomicity with a native process reopening these paths.
type configFence struct {
	root     string
	inputs   []string
	grants   map[string][]string
	baseline map[string]configDigest
	args     []string
}

type configDigest struct {
	Kind               stateread.Kind
	FileSHA, MCPSetSHA string
}

type projectConfigChanged struct{ reason string }

func (e *projectConfigChanged) Error() string {
	return "project_config_changed: " + e.reason + "; relaunch required"
}

func snapshotConfig(root string, inputs []string, grants map[string][]string) (map[string]configDigest, []string, error) {
	snapshot := map[string]configDigest{}
	args := []string{}
	for _, name := range inputs {
		// Reject quiescent links at every fixture-relative component, including
		// a linked .codex directory. This check does not freeze native reads.
		parts := strings.Split(filepath.ToSlash(name), "/")
		for i := range parts {
			path := filepath.Join(root, filepath.FromSlash(strings.Join(parts[:i+1], "/")))
			metadata, err := stateread.Lstat(path)
			if err != nil {
				return nil, nil, err
			}
			if metadata.Kind == stateread.KindPresent && metadata.Info.Mode()&os.ModeSymlink != 0 {
				return nil, nil, fmt.Errorf("config input is linked")
			}
		}
		file, err := stateread.ReadRegularFile(filepath.Join(root, name))
		if err != nil {
			return nil, nil, err
		}
		if file.Kind == stateread.KindAbsent {
			snapshot[name] = configDigest{Kind: file.Kind}
			continue
		}
		ceilings, err := enumeratedMCPCeilings(file.Bytes, grants)
		if err != nil {
			return nil, nil, err
		}
		var parsed map[string]any
		if _, err := toml.Decode(string(file.Bytes), &parsed); err != nil {
			return nil, nil, err
		}
		// encoded sorts map keys, so this independently inventories every
		// server and key/value, rather than comparing only config filenames.
		snapshot[name] = configDigest{Kind: file.Kind, FileSHA: fmt.Sprintf("%x", sha256.Sum256(file.Bytes)), MCPSetSHA: fmt.Sprintf("%x", sha256.Sum256([]byte(encoded(parsed["mcp_servers"]))))}
		args = append(args, ceilings...)
	}
	return snapshot, args, nil
}

func newConfigFence(root string, inputs []string, grants map[string][]string) (*configFence, error) {
	baseline, args, err := snapshotConfig(root, inputs, grants)
	if err != nil {
		return nil, err
	}
	return &configFence{root: root, inputs: append([]string(nil), inputs...), grants: grants, baseline: baseline, args: args}, nil
}

func (f *configFence) check() error {
	current, _, err := snapshotConfig(f.root, f.inputs, f.grants)
	if err != nil {
		return &projectConfigChanged{reason: "config input cannot be validated"}
	}
	if !reflect.DeepEqual(current, f.baseline) {
		return &projectConfigChanged{reason: "config bytes or MCP server set changed"}
	}
	return nil
}

// configPaths exists for the narrowing plant that ignores content digests.
func configPaths(snapshot map[string]configDigest) []string {
	paths := []string{}
	for path := range snapshot {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (f *configFence) admit(dispatch func()) error {
	if err := f.check(); err != nil {
		return err
	}
	dispatch()
	return nil
}

// Claim: before thread/turn dispatch, this candidate admits unchanged inputs
// and rejects each listed quiescent mutation with project_config_changed,
// without dispatch, repair or baseline recapture. This tests a preflight check,
// not atomicity with the independent native config reader tested below.
func TestConfigFenceAdmission(t *testing.T) {
	initial := "[mcp_servers.approved]\ncommand='fixture'\ndisabled_tools=['tool']\n[mcp_servers.removed]\ncommand='fixture'\n"
	for _, admission := range []string{"thread", "turn"} {
		for _, change := range []string{"late-server", "disabled-tools", "removed-server", "removed-file", "linked-file", "linked-directory", "added-config", "other-config", "unknown-key", "unparseable"} {
			t.Run(admission+"/"+change, func(t *testing.T) {
				root := t.TempDir()
				name := filepath.Join(root, ".codex/config.toml")
				other := filepath.Join(root, "nested/.codex/config.toml")
				put(t, name, initial)
				if change == "other-config" {
					put(t, other, initial)
				}
				fence, err := newConfigFence(root, []string{".codex/config.toml", "nested/.codex/config.toml"}, map[string][]string{"approved": {"tool"}})
				if err != nil {
					t.Fatal(err)
				}
				baseline := encoded(fence.baseline)
				calls := 0
				if err := fence.admit(func() { calls++ }); err != nil || calls != 1 {
					t.Fatalf("unchanged config admission failed: %v", err)
				}
				switch change {
				case "late-server":
					put(t, name, initial+"[mcp_servers.late]\ncommand='fixture'\n")
				case "disabled-tools":
					put(t, name, strings.Replace(initial, "disabled_tools=['tool']", "disabled_tools=[]", 1))
				case "removed-server":
					put(t, name, strings.Split(initial, "[mcp_servers.removed]")[0])
				case "removed-file":
					if err := os.Remove(name); err != nil {
						t.Fatal(err)
					}
				case "linked-file":
					if err := os.Rename(name, name+".regular"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(name+".regular", name); err != nil {
						t.Fatal(err)
					}
				case "linked-directory":
					dir := filepath.Dir(name)
					if err := os.Rename(dir, dir+".regular"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(dir+".regular", dir); err != nil {
						t.Fatal(err)
					}
				case "added-config", "other-config":
					put(t, other, initial+"[mcp_servers.late]\ncommand='fixture'\n")
				case "unknown-key":
					put(t, name, initial+"unknown_key=true\n")
				case "unparseable":
					put(t, name, "[broken")
				}
				before, beforeErr := os.ReadFile(name)
				for i := 0; i < 2; i++ {
					err := fence.admit(func() { calls++ })
					var changed *projectConfigChanged
					if !errors.As(err, &changed) || !strings.Contains(err.Error(), "project_config_changed") || !strings.Contains(err.Error(), "relaunch required") || calls != 1 {
						t.Fatalf("changed config reached %s dispatch: %v calls=%d", admission, err, calls)
					}
					if encoded(fence.baseline) != baseline {
						t.Fatal("changed config baseline recaptured")
					}
				}
				after, afterErr := os.ReadFile(name)
				if string(before) != string(after) || (beforeErr == nil) != (afterErr == nil) {
					t.Fatal("rejected config repaired")
				}
				if strings.HasPrefix(change, "linked-") {
					link := name
					if change == "linked-directory" {
						link = filepath.Dir(name)
					}
					info, err := os.Lstat(link)
					if err != nil || info.Mode()&os.ModeSymlink == 0 {
						t.Fatal("rejected link repaired")
					}
				}
			})
		}
	}
}

// Claim: malformed/unknown initial input cannot create a fence; unchanged
// valid input produces content and MCP digests and explicit manager ceilings.
func TestConfigFenceIntake(t *testing.T) {
	root := t.TempDir()
	name := filepath.Join(root, ".codex/config.toml")
	put(t, name, "[mcp_servers.approved]\ncommand='fixture'\n")
	fence, err := newConfigFence(root, []string{".codex/config.toml"}, map[string][]string{"approved": {"tool"}})
	if err != nil || fence.baseline[".codex/config.toml"].FileSHA == "" || fence.baseline[".codex/config.toml"].MCPSetSHA == "" || !strings.Contains(strings.Join(fence.args, " "), "mcp_servers.approved.disabled_tools=[]") {
		t.Fatalf("valid fence intake failed: %v", err)
	}
	for _, data := range []string{"[broken", "unknown=true", "[mcp_servers.approved]\nunknown=true\n"} {
		put(t, name, data)
		if _, err := newConfigFence(root, []string{".codex/config.toml"}, nil); err == nil {
			t.Fatal("unsafe initial fence input accepted")
		}
	}
}

// Claim: a successful digest check does not bind the later native read. A
// deterministic post-check edit loads an unenumerated MCP server on the same
// app-server. Restoring original bytes before the next check still leaves its
// runtime tools effective: disabled_tools removes the catalog entry, and the
// unrestricted positive makes a real synthetic call. All native thread starts
// pass through the candidate fence; no check is omitted. This is not model-turn
// or approval-dialog qualification, and the product launch must remain refused.
func TestAdapterConfigFenceReadGap(t *testing.T) {
	f := setup(t)
	script, err := os.ReadFile("testdata/mcp.py")
	if err != nil {
		t.Fatal(err)
	}
	fixtureScript := filepath.Join(f.root, "mcp.py")
	put(t, fixtureScript, string(script))
	stanza := func(name string, restricted bool) string {
		disabled := "[]"
		if restricted {
			disabled = fmt.Sprintf("[%q]", name)
		}
		return fmt.Sprintf("\n[mcp_servers.%s]\ncommand=\"/usr/bin/python3\"\nargs=[%q,%q,%q]\ndisabled_tools=%s\n", name, fixtureScript, name, filepath.Join(f.root, name+"-process.json"), disabled)
	}
	managerMCP := fmt.Sprintf("[mcp_servers.approved_tool]\ncommand=\"/usr/bin/python3\"\nargs=[%q,\"approved_tool\",%q]\n", fixtureScript, filepath.Join(f.root, "approved_tool-process.json"))
	p, _, _ := adapterPlan(t, f, "root", "read-only", adapterOptions{mcp: managerMCP})
	initial := stanza("approved_tool", true) + stanza("denied_tool", false)
	name := filepath.Join(f.work, ".codex/config.toml")
	put(t, name, initial)
	grants := map[string][]string{"approved_tool": {"approved_tool"}}
	fence, err := newConfigFence(f.work, []string{".codex/config.toml"}, grants)
	if err != nil {
		t.Fatal(err)
	}
	args := append(p.Launch().Args, fence.args...)
	a := f.api("config-fence-read-gap", args...)
	start := func() string {
		started := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
		return started["thread"].(map[string]any)["id"].(string)
	}
	tools := func(thread, server string) map[string]any {
		status := a.call("mcpServerStatus/list", map[string]any{"threadId": thread})
		for _, row := range status["data"].([]any) {
			s := row.(map[string]any)
			if s["name"] == server {
				return s["tools"].(map[string]any)
			}
		}
		return nil
	}
	var thread string
	if err := fence.admit(func() { thread = start() }); err != nil {
		t.Fatal(err)
	}
	if _, ok := tools(thread, "approved_tool")["approved_tool"]; !ok {
		t.Fatal("unchanged fenced tool missing")
	}
	called := a.call("mcpServer/tool/call", map[string]any{"threadId": thread, "server": "approved_tool", "tool": "approved_tool", "arguments": map[string]any{}})
	require(t, encoded(called), 0, "approved_tool_CALLED")
	if _, err := os.Stat(filepath.Join(f.root, "denied_tool-process.json")); !os.IsNotExist(err) {
		t.Fatal("known denied server started")
	}

	// Each quiescent mutation refuses before any native request, beside a
	// restored-input control that passes the fence and calls the admitted tool.
	changed := initial + stanza("late_tool", false)
	for _, change := range []string{"late-server", "disabled-tools", "removed-server", "linked-file"} {
		switch change {
		case "late-server":
			put(t, name, changed)
		case "disabled-tools":
			put(t, name, strings.Replace(initial, `disabled_tools=["approved_tool"]`, "disabled_tools=[]", 1))
		case "removed-server":
			put(t, name, stanza("approved_tool", true))
		case "linked-file":
			put(t, name+".regular", initial)
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(name+".regular", name); err != nil {
				t.Fatal(err)
			}
		}
		requestID := a.id
		var changedErr *projectConfigChanged
		if err := fence.admit(func() { thread = start() }); !errors.As(err, &changedErr) || a.id != requestID {
			t.Fatalf("quiescent %s dispatched", change)
		}
		if change == "linked-file" {
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
		}
		put(t, name, initial)
		if err := fence.admit(func() {
			called := a.call("mcpServer/tool/call", map[string]any{"threadId": thread, "server": "approved_tool", "tool": "approved_tool", "arguments": map[string]any{}})
			require(t, encoded(called), 0, "approved_tool_CALLED")
		}); err != nil {
			t.Fatalf("restored %s control refused: %v", change, err)
		}
	}
	// Relaunch re-enumerates the changed input, disabling the newly known
	// unadmitted server while retaining the admitted tool's actual call.
	put(t, name, changed)
	relaunch, err := newConfigFence(f.work, []string{".codex/config.toml"}, grants)
	if err != nil {
		t.Fatal(err)
	}
	b := f.api("config-fence-relaunch", append(p.Launch().Args, relaunch.args...)...)
	if err := relaunch.admit(func() {
		started := b.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
		id := started["thread"].(map[string]any)["id"].(string)
		status := b.call("mcpServerStatus/list", map[string]any{"threadId": id})
		for _, row := range status["data"].([]any) {
			s := row.(map[string]any)
			if s["name"] == "late_tool" && len(s["tools"].(map[string]any)) != 0 {
				t.Fatal("relaunch exposed unadmitted late tools")
			}
		}
		called := b.call("mcpServer/tool/call", map[string]any{"threadId": id, "server": "approved_tool", "tool": "approved_tool", "arguments": map[string]any{}})
		require(t, encoded(called), 0, "approved_tool_CALLED")
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(f.root, "late_tool-process.json")); !os.IsNotExist(err) {
		t.Fatal("relaunch started unadmitted late server")
	}
	put(t, name, initial)

	for _, restricted := range []bool{true, false} {
		checks := 0
		var late map[string]any
		err := fence.admit(func() {
			checks++ // This callback is reached only after the actual check succeeds.
			put(t, name, initial+stanza("late_tool", restricted))
			thread = start()
			// thread/start can acknowledge before asynchronous MCP startup has
			// consumed config. Observe the loaded catalog before restoring it.
			late = tools(thread, "late_tool")
			if late == nil {
				t.Fatal("post-check native reader never loaded changed config")
			}
			put(t, name, initial)
		})
		if err != nil || checks != 1 {
			t.Fatalf("post-check witness never admitted: %v", err)
		}
		if err := fence.check(); err != nil {
			t.Fatalf("restored bytes did not pass next fence: %v", err)
		}
		_, present := late["late_tool"]
		if present == restricted {
			t.Fatal("post-check native permission effect missing")
		}
		if _, err := os.Stat(filepath.Join(f.root, "late_tool-process.json")); err != nil {
			t.Fatal("late process receipt missing")
		}
		if !restricted {
			// Even a fresh check immediately before the actual call passes after
			// the config is restored, while the loaded runtime server survives.
			if err := fence.admit(func() {
				called := a.call("mcpServer/tool/call", map[string]any{"threadId": thread, "server": "late_tool", "tool": "late_tool", "arguments": map[string]any{}})
				require(t, encoded(called), 0, "late_tool_CALLED")
			}); err != nil {
				t.Fatal(err)
			}
		}
		f.save(fmt.Sprintf("config-fence-restricted-%t.json", restricted), encoded(map[string]any{"preflightPassed": true, "restoredPreflightPassed": true, "lateToolPresent": present, "syntheticCall": !restricted, "baseline": fence.baseline}))
	}
	var refusal *projectadapter.Unsupported
	if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != projectadapter.Permissions {
		t.Fatal("unsafe fenced launch accepted")
	}
}
