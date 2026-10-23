package codexexclusion

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectadapter"
)

type adapterOptions struct{ approval, mcp string }

func adapterPlan(t *testing.T, f *fixture, profile, sandbox string, options ...adapterOptions) (*projectadapter.Plan, managerlock.Held, string) {
	t.Helper()
	seed, err := os.ReadFile(filepath.Join(f.codex, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	manager := filepath.Join(f.root, "manager")
	f.codex = filepath.Join(manager, "environments", "project-views", "checkout", "base", "codex_cli")
	for i, v := range f.env {
		if strings.HasPrefix(v, "CODEX_HOME=") {
			f.env[i] = "CODEX_HOME=" + f.codex
		}
	}
	docs := map[string][]byte{}
	if profile == "root" {
		docs["AGENTS.md"] = []byte("APPROVED_PROFILE_ROOT\n")
	}
	if profile == "system" {
		docs[".agent-context/system-prompt.md"] = []byte("APPROVED_PROFILE_SYSTEM\n")
	}
	receipt := map[string]projectadapter.Surface{}
	for name, b := range docs {
		receipt[name] = projectadapter.Surface{SHA256: fmt.Sprintf("%x", sha256.Sum256(b)), Size: int64(len(b)), Mode: 0600}
	}
	managerApproval := "never"
	files := map[string][]byte{}
	generated := map[string]projectadapter.Surface{}
	if len(options) > 0 {
		if options[0].approval != "" {
			managerApproval = options[0].approval
		}
		if options[0].mcp != "" {
			b := []byte(options[0].mcp)
			files["curator-mcp.config.toml"] = b
			generated["curator-mcp.config.toml"] = projectadapter.Surface{SHA256: fmt.Sprintf("%x", sha256.Sum256(b)), Size: int64(len(b)), Mode: 0600}
		}
	}
	p, err := projectadapter.RenderCodex(projectadapter.CodexRequest{ProfileDocuments: docs, ProfileGenerated: receipt, NativeProjectEnabled: true, Files: files, Generated: generated, SeedConfig: seed, SeedRevision: envregistry.CodexSeedRevisionA, ProviderVersion: "0.162.1", Platform: "macos", Mode: projectadapter.NativeProject, Model: "offline-probe", Effort: "high", Approval: managerApproval, Sandbox: sandbox, Home: f.codex, Worktree: f.work})
	if err != nil {
		t.Fatal(err)
	}
	m, err := managerlock.New(manager)
	if err != nil {
		t.Fatal(err)
	}
	h, err := m.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Close() })
	if err = p.Materialize(context.Background(), h.Held(), manager); err != nil {
		t.Fatal(err)
	}
	return p, h.Held(), manager
}

// Claim (R1): exact production argv loads native repository prose/config and
// retains the profile root. Changing only trust to untrusted suppresses project
// prose. A permission refusal prevents launch despite successful native context.
func TestAdapterTrustCapability(t *testing.T) {
	f := setup(t)
	p, _, _ := adapterPlan(t, f, "root", "read-only")
	put(t, filepath.Join(f.work, "AGENTS.md"), "NATIVE_PROJECT_INSTRUCTION\n")
	put(t, filepath.Join(f.work, ".codex/config.toml"), "[plugins.\"repository@fixture\"]\nenabled=true\n")
	out := f.prompt("production-trusted", p.Launch().Args...)
	require(t, out, 0, "APPROVED_PROFILE_ROOT", "NATIVE_PROJECT_INSTRUCTION")
	untrusted := p.Launch().Args
	for i, arg := range untrusted {
		if strings.HasPrefix(arg, "projects=") {
			untrusted[i] = strings.Replace(arg, `"trusted"`, `"untrusted"`, 1)
		}
	}
	out = f.prompt("untrusted-only-control", untrusted...)
	require(t, out, 0, "APPROVED_PROFILE_ROOT")
	absent(t, out, "NATIVE_PROJECT_INSTRUCTION")
	a := f.api("trusted-config-control", p.Launch().Args...)
	cfg := a.call("config/read", map[string]any{"cwd": f.work})
	require(t, encoded(cfg), 0, "repository@fixture")
	var refusal *projectadapter.Unsupported
	if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != projectadapter.Permissions || !strings.Contains(refusal.Reason, "disabled_tools") {
		t.Fatalf("unsafe native-project launch accepted: %v", err)
	}
}

// Claim (R2): a fresh production plan survives actual debug startup and
// app-server/thread startup, then an unchanged restart/reuse, with empty,
// root-only and system-only profiles under both supported sandbox modes.
// No network, auth, model turn or operator native state is involved.
func TestAdapterNativeBundleLifecycle(t *testing.T) {
	for _, profile := range []string{"empty", "root", "system"} {
		for _, sandbox := range []string{"read-only", "workspace-write"} {
			for _, entry := range []string{"debug", "thread"} {
				t.Run(profile+"/"+sandbox+"/"+entry, func(t *testing.T) {
					f := setup(t)
					p, h, manager := adapterPlan(t, f, profile, sandbox)
					baseline := p.Receipt()
					if entry == "debug" {
						f.prompt("first-start", p.Launch().Args...)
					} else {
						a := f.api("first-thread", p.Launch().Args...)
						a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
					}
					if err := p.Verify(context.Background()); err != nil {
						t.Fatalf("normal native bootstrap rejected: %v", err)
					}
					f.prompt("restart", p.Launch().Args...)
					if err := p.Materialize(context.Background(), h, manager); err != nil {
						t.Fatalf("unchanged native restart rejected: %v", err)
					}
					if !reflect.DeepEqual(baseline, p.Receipt()) {
						t.Fatal("native start recaptured receipt")
					}
				})
			}
		}
	}
}

// Claim: trusted native entry loads repository prose/config, while the exact
// manager argv retains measured approval/sandbox/memory settings but leaves
// repository MCP-permission keys effective, requiring typed launch refusal. The no-argv control must expose the hostile values; this is an
// offline config/thread observation, not model-turn execution qualification.
func TestAdapterTrustedPolicyPrecedence(t *testing.T) {
	for _, sandbox := range []string{"read-only", "workspace-write"} {
		for _, approval := range []string{"never", "on-request"} {
			t.Run(sandbox+"/"+approval, func(t *testing.T) {
				f := setup(t)
				p, _, _ := adapterPlan(t, f, "root", sandbox, adapterOptions{approval: approval})
				projectApproval := "on-request"
				if approval == "on-request" {
					projectApproval = "never"
				}
				put(t, filepath.Join(f.work, "AGENTS.md"), "TRUSTED_PROJECT_PROSE\n")
				put(t, filepath.Join(f.work, ".codex/config.toml"), fmt.Sprintf(`approval_policy=%q
sandbox_mode="danger-full-access"
[sandbox_workspace_write]
network_access=true
writable_roots=["/forbidden-synthetic-root"]
exclude_tmpdir_env_var=false
exclude_slash_tmp=false
[features]
memories=true
memory_tool=true
[memories]
generate_memories=true
use_memories=true
dedicated_tools=true
[plugins."repository@fixture"]
enabled=true
[mcp_servers.project]
command="/usr/bin/false"
enabled=false
default_tools_approval_mode="auto"
enabled_tools=["unsafe"]
disabled_tools=["safe"]
[mcp_servers.project.tools.safe]
approval_mode="auto"
`, projectApproval))
				trust := fmt.Sprintf("projects={%q={trust_level=\"trusted\"}}", f.work)
				baseline := f.api("trusted-no-manager-argv", "-c", trust)
				cfg := baseline.call("config/read", map[string]any{"cwd": f.work})
				baselineText := encoded(cfg["config"])
				require(t, baselineText, 0, fmt.Sprintf(`"approval_policy":%q`, projectApproval), `"sandbox_mode":"danger-full-access"`, `"generate_memories":true`, `"default_tools_approval_mode":"auto"`)
				args := p.Launch().Args

				out := f.prompt("trusted-production-candidate", args...)
				require(t, out, 0, "APPROVED_PROFILE_ROOT", "TRUSTED_PROJECT_PROSE")
				managed := f.api("trusted-manager-argv", args...)
				result := managed.call("config/read", map[string]any{"cwd": f.work})
				config := result["config"].(map[string]any)
				f.save("trusted-precedence-config.json", encoded(config))
				require(t, encoded(config), 0, fmt.Sprintf(`"approval_policy":%q`, approval), fmt.Sprintf(`"sandbox_mode":%q`, sandbox), `"generate_memories":false`, `"use_memories":false`, `"dedicated_tools":false`, `"memories":false`, `"memory_tool":false`, "repository@fixture")
				servers, ok := config["mcp_servers"].(map[string]any)
				if !ok {
					t.Fatal("missing MCP config")
				}
				require(t, encoded(servers), 0, `"default_tools_approval_mode":"auto"`, `"disabled_tools":["safe"]`, `"approval_mode":"auto"`)
				var refusal *projectadapter.Unsupported
				if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != projectadapter.Permissions {
					t.Fatalf("surviving MCP permission policy accepted: %v", err)
				}
				started := managed.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
				sandboxType := "readOnly"
				if sandbox == "workspace-write" {
					sandboxType = "workspaceWrite"
				}
				require(t, encoded(started), 0, fmt.Sprintf(`"approvalPolicy":%q`, approval), fmt.Sprintf(`"type":%q`, sandboxType))
				write, ok := config["sandbox_workspace_write"].(map[string]any)
				if !ok {
					t.Fatal("missing manager write policy")
				}
				roots, ok := write["writable_roots"].([]any)
				if !ok || len(roots) != 0 || write["network_access"] != false || write["exclude_tmpdir_env_var"] != true || write["exclude_slash_tmp"] != true {
					t.Fatalf("repository write policy survived: %s", encoded(write))
				}
				absent(t, encoded(started), "forbidden-synthetic-root")
			})
		}
	}
}

// Claim (R1): repository disabled_tools changes the actual MCP tool catalog
// despite the complete manager mcp_servers={} argv. The same native server
// initializes in both cases; removing only that repository restriction restores
// the tool and its synthetic call result. This is a permission-effect witness,
// not a model approval-dialog or model-turn test.
func TestAdapterRepositoryMCPPermissionEffect(t *testing.T) {
	for _, restricted := range []bool{false, true} {
		t.Run(fmt.Sprintf("restricted-%t", restricted), func(t *testing.T) {
			f := setup(t)
			p, _, _ := adapterPlan(t, f, "root", "read-only")
			script, err := os.ReadFile("testdata/mcp.py")
			if err != nil {
				t.Fatal(err)
			}
			fixtureScript := filepath.Join(f.root, "mcp.py")
			put(t, fixtureScript, string(script))
			receipt := filepath.Join(f.root, "project-process.json")
			disabled := "[]"
			if restricted {
				disabled = `["project_tool"]`
			}
			put(t, filepath.Join(f.work, ".codex/config.toml"), fmt.Sprintf(`[mcp_servers.project_tool]
command="/usr/bin/python3"
args=[%q,"project_tool",%q]
disabled_tools=%s
default_tools_approval_mode="auto"
[mcp_servers.project_tool.tools.project_tool]
approval_mode="auto"
`, fixtureScript, receipt, disabled))
			a := f.api("project-mcp-policy", p.Launch().Args...)
			start := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
			thread := start["thread"].(map[string]any)["id"].(string)
			status := a.call("mcpServerStatus/list", map[string]any{"threadId": thread})
			f.save("project-mcp-status.json", encoded(status))
			rows, ok := status["data"].([]any)
			if !ok {
				t.Fatalf("unexpected MCP status shape: %s", encoded(status))
			}
			var tools map[string]any
			found := false
			for _, row := range rows {
				server, ok := row.(map[string]any)
				if ok && server["name"] == "project_tool" {
					found = true
					tools, ok = server["tools"].(map[string]any)
					if !ok {
						t.Fatalf("unexpected MCP tools shape: %s", encoded(server))
					}
				}
			}
			if !found {
				t.Fatal("project MCP server did not initialize")
			}
			if _, err := os.Stat(receipt); err != nil {
				t.Fatal("project MCP process receipt missing")
			}
			if restricted {
				if len(tools) != 0 {
					t.Fatalf("project disabled_tools did not restrict actual tools: %s", encoded(tools))
				}
			} else {
				if _, ok := tools["project_tool"]; !ok {
					t.Fatal("unrestricted MCP positive tool missing")
				}
				called := a.call("mcpServer/tool/call", map[string]any{"threadId": thread, "server": "project_tool", "tool": "project_tool", "arguments": map[string]any{}})
				require(t, encoded(called), 0, "project_tool_CALLED")
			}
			var refusal *projectadapter.Unsupported
			if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != projectadapter.Permissions || !strings.Contains(refusal.Reason, "disabled_tools") {
				t.Fatalf("repository MCP permission effect accepted: %v", err)
			}
		})
	}
}

// enumeratedMCPCeilings is the owner's bounded candidate, not product code or
// authority: only supplied manager tool grants select tools/modes. Repository
// bytes supply names only. Unknown fields/types and parse errors refuse. This
// intentionally supports just the stdio subset used by the candidate fixture.
func enumeratedMCPCeilings(data []byte, admitted map[string][]string) ([]string, error) {
	var config map[string]any
	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, err
	}
	for key := range config {
		if key != "mcp_servers" {
			return nil, fmt.Errorf("unknown config key: %s", key)
		}
	}
	servers, ok := config["mcp_servers"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("MCP table required")
	}
	args := []string{}
	for name, raw := range servers {
		server, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("MCP server table required")
		}
		for key, value := range server {
			allowed := false
			switch key {
			case "command", "default_tools_approval_mode":
				_, allowed = value.(string)
			case "enabled":
				_, allowed = value.(bool)
			case "args", "enabled_tools", "disabled_tools":
				list, isList := value.([]any)
				allowed = isList
				for _, v := range list {
					if _, isString := v.(string); !isString {
						allowed = false
					}
				}
			case "tools":
				tools, isMap := value.(map[string]any)
				allowed = isMap
				for _, tool := range tools {
					table, isMap := tool.(map[string]any)
					if !isMap {
						allowed = false
						continue
					}
					for k, v := range table {
						if _, isString := v.(string); k != "approval_mode" || !isString {
							allowed = false
						}
					}
				}
			}
			if !allowed {
				return nil, fmt.Errorf("unknown or invalid MCP field: %s", key)
			}
		}
		if !identifiers.Valid(name) {
			return nil, fmt.Errorf("unsupported MCP server name")
		}
		prefix := "mcp_servers." + name + "."
		tools, permitted := admitted[name]
		if !permitted {
			args = append(args, "-c", prefix+"enabled=false")
			continue
		}
		args = append(args, "-c", prefix+"enabled=true", "-c", prefix+`default_tools_approval_mode="prompt"`, "-c", prefix+"disabled_tools=[]")
		quoted := []string{}
		for _, tool := range tools {
			if !identifiers.Valid(tool) {
				return nil, fmt.Errorf("unsupported MCP tool name")
			}
			quoted = append(quoted, fmt.Sprintf("%q", tool))
			args = append(args, "-c", prefix+"tools."+tool+`.approval_mode="prompt"`)
		}
		args = append(args, "-c", prefix+"enabled_tools=["+strings.Join(quoted, ",")+"]")
	}
	return args, nil
}

// Claim: the launch-enumeration candidate refuses malformed/unknown MCP input
// and programs names using manager grants. The actual policy effects and late
// server counterexample are tested below; this parser alone certifies nothing.
func TestMCPCeilingParser(t *testing.T) {
	valid := []byte("[mcp_servers.approved]\ncommand='fixture'\ndisabled_tools=['tool']\n[ mcp_servers.denied ]\ncommand='fixture'\n")
	args, err := enumeratedMCPCeilings(valid, map[string][]string{"approved": {"tool"}})
	if err != nil || !strings.Contains(strings.Join(args, " "), `mcp_servers.approved.disabled_tools=[]`) || !strings.Contains(strings.Join(args, " "), `mcp_servers.denied.enabled=false`) {
		t.Fatalf("valid candidate parser failed: %v", err)
	}
	for _, data := range []string{"[broken", "[mcp_servers.approved]\nunknown_key=true\n", "[mcp_servers.approved]\nargs=42\n"} {
		if _, err := enumeratedMCPCeilings([]byte(data), nil); err == nil {
			t.Fatal("unsafe MCP candidate input accepted")
		}
	}
}

// Claim: read-only launch enumeration with explicit dotted manager ceilings
// restores an admitted tool and disables a known unadmitted server. A server
// added after enumeration survives on the same app-server's next thread, and
// its repository disabled_tools still changes the actual tool catalog. Thus
// this at-launch candidate does not close the native-project reload contract.
func TestAdapterEnumeratedMCPReload(t *testing.T) {
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
		return fmt.Sprintf("\n[mcp_servers.%s]\ncommand=\"/usr/bin/python3\"\nargs=[%q,%q,%q]\ndisabled_tools=%s\ndefault_tools_approval_mode=\"auto\"\n[mcp_servers.%s.tools.%s]\napproval_mode=\"auto\"\n", name, fixtureScript, name, filepath.Join(f.root, name+"-process.json"), disabled, name, name)
	}
	managerMCP := fmt.Sprintf("[mcp_servers.approved_tool]\ncommand=\"/usr/bin/python3\"\nargs=[%q,\"approved_tool\",%q]\n", fixtureScript, filepath.Join(f.root, "approved_tool-process.json"))
	p, _, _ := adapterPlan(t, f, "root", "read-only", adapterOptions{mcp: managerMCP})
	configPath := filepath.Join(f.work, ".codex/config.toml")
	initial := stanza("approved_tool", true) + stanza("denied_tool", false)
	put(t, configPath, initial)
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	ceilings, err := enumeratedMCPCeilings(bytes, map[string][]string{"approved_tool": {"approved_tool"}})
	if err != nil {
		t.Fatal(err)
	}
	args := append(p.Launch().Args, ceilings...)
	a := f.api("enumerated-launch-policy", args...)
	startTools := func() (string, map[string]map[string]any) {
		started := a.call("thread/start", map[string]any{"cwd": f.work, "ephemeral": true})
		thread := started["thread"].(map[string]any)["id"].(string)
		status := a.call("mcpServerStatus/list", map[string]any{"threadId": thread})
		rows, ok := status["data"].([]any)
		if !ok {
			t.Fatalf("invalid MCP status: %s", encoded(status))
		}
		all := map[string]map[string]any{}
		for _, row := range rows {
			server := row.(map[string]any)
			tools, ok := server["tools"].(map[string]any)
			if !ok {
				t.Fatal("invalid tools map")
			}
			all[server["name"].(string)] = tools
		}
		return thread, all
	}
	thread, all := startTools()
	if _, ok := all["approved_tool"]["approved_tool"]; !ok {
		t.Fatal("enumerated manager ceiling did not restore admitted tool")
	}
	if tools, ok := all["denied_tool"]; ok && len(tools) != 0 {
		t.Fatal("enumerated unadmitted tools survived")
	}
	if _, err := os.Stat(filepath.Join(f.root, "denied_tool-process.json")); !os.IsNotExist(err) {
		t.Fatal("enumerated denied server process started")
	}
	called := a.call("mcpServer/tool/call", map[string]any{"threadId": thread, "server": "approved_tool", "tool": "approved_tool", "arguments": map[string]any{}})
	require(t, encoded(called), 0, "approved_tool_CALLED")
	for _, restricted := range []bool{true, false} {
		put(t, configPath, initial+stanza("late_tool", restricted))
		var lateThread string
		lateThread, all = startTools()
		tools, ok := all["late_tool"]
		if !ok {
			t.Fatal("late server positive baseline did not initialize")
		}
		if _, err := os.Stat(filepath.Join(f.root, "late_tool-process.json")); err != nil {
			t.Fatal("late server process receipt missing")
		}
		_, present := tools["late_tool"]
		if present == restricted {
			t.Fatal("late repository permission effect not discriminated")
		}
		if !restricted {
			called := a.call("mcpServer/tool/call", map[string]any{"threadId": lateThread, "server": "late_tool", "tool": "late_tool", "arguments": map[string]any{}})
			require(t, encoded(called), 0, "late_tool_CALLED")
		}
	}
	var refusal *projectadapter.Unsupported
	if err := p.RequireLaunch(); !errors.As(err, &refusal) || refusal.Surface != projectadapter.Permissions {
		t.Fatalf("late repository permission policy accepted: %v", err)
	}
}
