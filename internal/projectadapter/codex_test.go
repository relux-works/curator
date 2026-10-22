package projectadapter

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
)

func request(t *testing.T) CodexRequest {
	t.Helper()
	root := t.TempDir()
	files := map[string][]byte{"AGENTS.md": []byte("profile\n\nproject\n"), "skills/tool/SKILL.md": []byte("# Tool\n"), "skills/tool/data.txt": []byte("whole skill\n"), "curator-mcp.config.toml": []byte("[mcp_servers.approved]\ncommand = \"fixture-mcp\"\nargs = [\"run\"]\n")}
	receipt := map[string]Surface{}
	for name, b := range files {
		receipt[name] = surface(b, 0600)
	}
	profile := map[string][]byte{"AGENTS.md": []byte("profile\n")}
	profileReceipt := map[string]Surface{"AGENTS.md": surface(profile["AGENTS.md"], 0600)}
	return CodexRequest{ProfileDocuments: profile, ProfileGenerated: profileReceipt, NativeProjectEnabled: true, Files: files, Generated: receipt, ProviderVersion: "0.162.1", Platform: "macos", Mode: NativeProject, SeedRevision: envregistry.CodexSeedRevisionA, Model: "fixture-model", Effort: "high", Approval: "never", Sandbox: "read-only", Home: filepath.Join(root, "manager", "environments", "project-views", "checkout", "base", "codex_cli"), Worktree: filepath.Join(root, "worktree")}
}
func render(t *testing.T, r CodexRequest) *Plan {
	t.Helper()
	p, err := RenderCodex(r)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func document(t *testing.T, b []byte) map[string]any {
	t.Helper()
	v := map[string]any{}
	if _, err := toml.Decode(string(b), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// Claim: profile-only instructions and admitted skill bytes survive rendering;
// manager policy and MCP
// replace conflicting revision-A/B seed keys with named notes, retaining an
// unrelated seed field. The empty MCP control removes every seeded server too.
func TestCodexCompositionAndSeed(t *testing.T) {
	for _, rev := range []string{envregistry.CodexSeedRevisionA, envregistry.CodexSeedRevisionB} {
		for _, empty := range []bool{false, true} {
			t.Run(rev+"/"+map[bool]string{true: "empty", false: "servers"}[empty], func(t *testing.T) {
				r := request(t)
				r.SeedRevision = rev
				r.SeedConfig = []byte("theme = \"seed-theme\"\napproval_policy=\"on-request\"\nsandbox_mode=\"danger-full-access\"\nprofile=\"unsafe\"\n[features]\nmemories=true\nshell_snapshot=true\n[memories]\ngenerate_memories=true\nuse_memories=true\n[mcp_servers.seed]\ncommand=\"unapproved\"\n[profiles.unsafe]\napproval_policy=\"on-request\"\n")
				if empty {
					delete(r.Files, "curator-mcp.config.toml")
					delete(r.Generated, "curator-mcp.config.toml")
				}
				p := render(t, r)
				files := p.Files()
				if !bytes.Equal(files["AGENTS.md"], r.ProfileDocuments["AGENTS.md"]) {
					t.Fatal("root bytes differ")
				}
				for _, name := range []string{"skills/tool/SKILL.md", "skills/tool/data.txt"} {
					if !bytes.Equal(files[name], r.Files[name]) {
						t.Fatalf("whole skill lost: %s", name)
					}
				}
				d := document(t, files["config.toml"])
				servers := d["mcp_servers"].(map[string]any)
				if _, exists := servers["seed"]; exists {
					t.Fatal("seed MCP survived composition")
				}
				if len(servers) != map[bool]int{true: 0, false: 1}[empty] {
					t.Fatal("composed MCP differs")
				}
				if d["theme"] != "seed-theme" || d["approval_policy"] != "never" || d["sandbox_mode"] != "read-only" || d["profile"] != nil || d["profiles"] != nil {
					t.Fatal("manager policy or unrelated seed field differs")
				}
				if d["features"].(map[string]any)["memories"] != false || d["features"].(map[string]any)["shell_snapshot"] != true || d["memories"].(map[string]any)["use_memories"] != false {
					t.Fatal("memory controls differ")
				}
				notes := p.Notes()
				for _, key := range []string{"mcp_servers.seed", "approval_policy", "sandbox_mode", "profile", "profiles", "features.memories", "memories.use_memories"} {
					found := false
					for _, note := range notes {
						if note.Key == key {
							found = true
						}
					}
					if !found {
						t.Fatalf("seed key silently changed: %s", key)
					}
				}
				for name, b := range files {
					if p.Receipt()[name] != surface(b, 0600) {
						t.Fatalf("receipt differs: %s", name)
					}
				}
			})
		}
	}
}

// Claim: emitted -c arguments are independently parseable TOML and bind the
// same manager-owned policy, memory, model and worktree distrust as config.toml.
// Typed strict refusal is preserved; native-project needs explicit opt-in and
// reports the same non-isolation notice in status and launch.
func TestCodexLaunchAndCertification(t *testing.T) {
	r := request(t)
	r.Model = "fixture \"quoted\" model"
	p := render(t, r)
	launch := p.Launch()
	if launch.Env["CODEX_HOME"] != r.Home || launch.Mode != NativeProject || launch.Notice != NativeProjectNotice {
		t.Fatal("missing managed home")
	}
	overrides := map[string]any{}
	var overrideText strings.Builder
	for i := 0; i < len(launch.Args); i += 2 {
		if launch.Args[i] != "-c" {
			t.Fatal("unexpected launch channel")
		}
		overrideText.WriteString(launch.Args[i+1] + "\n")
	}
	if _, err := toml.Decode(overrideText.String(), &overrides); err != nil {
		t.Fatal(err)
	}
	if overrides["model"] != r.Model || overrides["approval_policy"] != r.Approval || overrides["sandbox_mode"] != r.Sandbox {
		t.Fatal("manager overrides differ")
	}
	if overrides["projects"].(map[string]any)[r.Worktree].(map[string]any)["trust_level"] != "untrusted" || overrides["project_doc_max_bytes"] != int64(32768) {
		t.Fatal("project controls not applied")
	}
	if overrides["memories"].(map[string]any)["generate_memories"] != false || overrides["features"].(map[string]any)["memories"] != false {
		t.Fatal("memory overrides missing")
	}
	report := p.Report()
	if report.Notice != NativeProjectNotice || report.Environment != "codex_cli" || report.Mode != NativeProject || len(report.Certified) != 5 || len(report.Refused) != 1 || report.Refused[0].Surface != NativeSourceExclusion || !strings.Contains(report.Refused[0].Reason, "late-created") {
		t.Fatal("certification overclaims")
	}
	if err := p.RequireLaunch(); err != nil {
		t.Fatal(err)
	}
	var unsupported *Unsupported
	if !errors.As(p.RequireStrict(), &unsupported) || unsupported.Surface != NativeSourceExclusion {
		t.Fatal("strict launch did not refuse")
	}
	// Accessor mutation cannot replace the protected baseline.
	files := p.Files()
	files["AGENTS.md"][0] = 'x'
	receipt := p.Receipt()
	delete(receipt, "AGENTS.md")
	launch.Env["CODEX_HOME"] = "wrong"
	report.Refused[0].Reason = "wrong"
	if !bytes.Equal(p.Files()["AGENTS.md"], r.ProfileDocuments["AGENTS.md"]) || len(p.Receipt()) == 0 || p.Launch().Env["CODEX_HOME"] != r.Home || p.Report().Refused[0].Reason == "wrong" {
		t.Fatal("plan baseline is mutable")
	}
}

// Claim: unsupported modes/releases/policies, unreceipted bytes, native config
// outputs, malformed or executable seed control sources refuse before writing.
// The adjacent ordinary request remains renderable.
func TestCodexRefusals(t *testing.T) {
	cases := []struct {
		name string
		edit func(*CodexRequest)
		want string
	}{
		{"strict", func(r *CodexRequest) { r.Mode = Strict }, "only native-project"},
		{"no-opt-in", func(r *CodexRequest) { r.NativeProjectEnabled = false }, "not enabled by the operator"},
		{"missing-profile-documents", func(r *CodexRequest) { r.ProfileDocuments = nil }, "profile documents"},
		{"version", func(r *CodexRequest) { r.ProviderVersion = "0.162.2" }, "release/platform"},
		{"linux", func(r *CodexRequest) { r.Platform = "linux" }, "release/platform"},
		{"relative-home", func(r *CodexRequest) { r.Home = "relative" }, "manager home/worktree"},
		{"seed-revision", func(r *CodexRequest) { r.SeedRevision = "unknown" }, "seed revision"},
		{"policy", func(r *CodexRequest) { r.Sandbox = "danger-full-access" }, "manager approval"},
		{"model", func(r *CodexRequest) { r.Model = "" }, "manager model"},
		{"effort", func(r *CodexRequest) { r.Effort = "invalid" }, "reasoning effort"},
		{"drift", func(r *CodexRequest) { r.Files["AGENTS.md"] = []byte("unapproved") }, "receipt"},
		{"native-config", func(r *CodexRequest) {
			r.Files[".codex/config.toml"] = []byte("approval_policy='never'")
			r.Generated[".codex/config.toml"] = surface(r.Files[".codex/config.toml"], 0600)
		}, "unsupported native output"},
		{"parent", func(r *CodexRequest) {
			r.Files["skills/tool/../../../escape"] = []byte("bad")
			r.Generated["skills/tool/../../../escape"] = surface([]byte("bad"), 0600)
		}, "unsafe output"},
		{"malformed-seed", func(r *CodexRequest) { r.SeedConfig = []byte("[broken") }, "valid TOML"},
		{"seed-hooks", func(r *CodexRequest) { r.SeedConfig = []byte("[hooks]\n") }, "seed control key hooks"},
		{"seed-memory-kind", func(r *CodexRequest) { r.SeedConfig = []byte("memories=true") }, "memories table"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := request(t)
			tc.edit(&r)
			p, err := RenderCodex(r)
			if p != nil || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %s refusal, got %v", tc.want, err)
			}
			if _, err = os.Lstat(r.Home); !os.IsNotExist(err) {
				t.Fatal("refusal wrote home")
			}
		})
	}
	render(t, request(t))
}

func held(t *testing.T, home string) managerlock.Held {
	t.Helper()
	m, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := m.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lock.Close() })
	return lock.Held()
}

// Claim: fresh publication is complete and receipt-exact; identical reuse
// preserves native mutable state. Byte/mode/missing/extra/linked output drift
// refuses without changing any affected file or replacing the baseline.
func TestCodexMaterializeAndDrift(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("project materialization boundary is qualified on macOS/Linux; rendering contracts remain portable")
	}
	for _, kind := range []string{"bytes", "mode", "missing", "extra", "symlink", "override"} {
		t.Run(kind, func(t *testing.T) {
			r := request(t)
			p := render(t, r)
			managerHome := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(r.Home)))))
			h := held(t, managerHome)
			if err := p.Materialize(context.Background(), h, managerHome); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(r.Home, "sessions.json"), []byte("mutable"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := p.Materialize(context.Background(), h, managerHome); err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(r.Home, "AGENTS.md")
			switch kind {
			case "bytes":
				_ = os.WriteFile(name, bytes.Repeat([]byte("x"), len(r.ProfileDocuments["AGENTS.md"])), 0600)
			case "mode":
				_ = os.Chmod(name, 0644)
			case "missing":
				_ = os.Remove(name)
			case "extra":
				_ = os.WriteFile(filepath.Join(r.Home, "skills/tool/extra.md"), []byte("late"), 0600)
			case "symlink":
				_ = os.Remove(name)
				_ = os.Symlink("sessions.json", name)
			case "override":
				_ = os.WriteFile(filepath.Join(r.Home, "AGENTS.override.md"), []byte("override"), 0600)
			}
			before := driftObservation(t, r.Home)
			baseline := p.Receipt()
			if err := p.Verify(context.Background()); err == nil {
				t.Fatal("drift accepted")
			}
			if err := p.Materialize(context.Background(), h, managerHome); err == nil {
				t.Fatal("drift repaired silently")
			}
			if !reflect.DeepEqual(before, driftObservation(t, r.Home)) {
				t.Fatal("drift verification changed native files")
			}
			if !reflect.DeepEqual(baseline, p.Receipt()) {
				t.Fatal("drift replaced baseline")
			}
			b, err := os.ReadFile(filepath.Join(r.Home, "sessions.json"))
			if err != nil || string(b) != "mutable" {
				t.Fatal("native state changed")
			}
		})
	}
}

// Claim: both supported explicit manager policies render, and a retained
// system document uses the registry's separate model_instructions_file channel.
// Named seed permission profiles and alternate base instructions refuse.
func TestCodexManagerPolicyAndSystem(t *testing.T) {
	for _, sandbox := range []string{"read-only", "workspace-write"} {
		for _, approval := range []string{"never", "on-request"} {
			r := request(t)
			r.Sandbox = sandbox
			r.Approval = approval
			r.ProfileDocuments[".agent-context/system-prompt.md"] = []byte("Profile system.\n")
			r.Files[".agent-context/system-prompt.md"] = []byte("Profile system.\n\nProject system.\n")
			r.Generated[".agent-context/system-prompt.md"] = surface(r.Files[".agent-context/system-prompt.md"], 0600)
			r.ProfileGenerated[".agent-context/system-prompt.md"] = surface(r.ProfileDocuments[".agent-context/system-prompt.md"], 0600)
			r.SeedConfig = []byte("approvals_reviewer='auto_review'\nmodel_instructions_file='/unapproved/source'\n[skills]\nconfig=[]\n")
			p := render(t, r)
			if !bytes.Equal(p.Files()[".agent-context/system-prompt.md"], r.ProfileDocuments[".agent-context/system-prompt.md"]) {
				t.Fatal("project system context installed in native-project")
			}
			d := document(t, p.Files()["config.toml"])
			if d["sandbox_mode"] != sandbox || d["approval_policy"] != approval || d["approvals_reviewer"] != "user" || d["skills"] != nil || d["model_instructions_file"] != filepath.Join(r.Home, ".agent-context/system-prompt.md") {
				t.Fatal("manager system/policy delivery differs")
			}
		}
	}
	for _, key := range []string{"instructions", "permissions", "default_permissions"} {
		r := request(t)
		r.SeedConfig = []byte(key + "='conflict'\n")
		if _, err := RenderCodex(r); err == nil || !strings.Contains(err.Error(), "seed control key "+key) {
			t.Fatalf("alternate seed control accepted: %s: %v", key, err)
		}
	}
}

// Claim: receipt aliases, missing receipt entries, file/ancestor collisions,
// non-MCP composed config keys and foreign MCP settings refuse alongside the
// valid input control. These are in-memory inputs; refusal publishes nothing.
func TestCodexInputBoundaries(t *testing.T) {
	cases := []struct {
		name string
		edit func(*CodexRequest)
		want string
	}{

		{"alias", func(r *CodexRequest) {
			for _, name := range []string{"skills/tool/ALIAS.md", "skills/tool/alias.md"} {
				r.Files[name] = []byte("alias")
				r.Generated[name] = surface([]byte("alias"), 0600)
			}
		}, "output path alias"},
		{"missing", func(r *CodexRequest) { delete(r.Generated, "AGENTS.md") }, "inventory mismatch"},
		{"ancestor", func(r *CodexRequest) {
			r.Files["skills/tool"] = []byte("ancestor")
			r.Generated["skills/tool"] = surface([]byte("ancestor"), 0600)
		}, "ancestor collision"},
		{"mcp-root", func(r *CodexRequest) {
			r.Files["curator-mcp.config.toml"] = []byte("approval_policy='never'\n")
			r.Generated["curator-mcp.config.toml"] = surface(r.Files["curator-mcp.config.toml"], 0600)
		}, "non-MCP"},
		{"mcp-http", func(r *CodexRequest) {
			r.Files["curator-mcp.config.toml"] = []byte("[mcp_servers.approved]\nurl='https://example.invalid/mcp'\n")
			r.Generated["curator-mcp.config.toml"] = surface(r.Files["curator-mcp.config.toml"], 0600)
		}, "HTTP native entry loading is unqualified"},
		{"mcp-setting", func(r *CodexRequest) {
			r.Files["curator-mcp.config.toml"] = []byte("[mcp_servers.approved]\ncommand='fixture'\nenv={SECRET='literal'}\n")
			r.Generated["curator-mcp.config.toml"] = surface(r.Files["curator-mcp.config.toml"], 0600)
		}, "unowned MCP"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := request(t)
			tc.edit(&r)
			if _, err := RenderCodex(r); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %s refusal, got %v", tc.want, err)
			}
		})
	}
	render(t, request(t))
}

// Claim: a zero/wrong-home witness, cancelled context, invalid allocation,
// overlapping worktree or linked managed parent refuses before home publication.
// The success control is the complete materialization test above.
func TestCodexPublicationBoundaries(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("project materialization boundary is qualified on macOS/Linux; rendering contracts remain portable")
	}
	for _, kind := range []string{"zero-lock", "other-lock", "cancelled", "allocation", "overlap", "linked-parent"} {
		t.Run(kind, func(t *testing.T) {
			r := request(t)
			managerHome := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(r.Home)))))
			h := held(t, managerHome)
			ctx := context.Background()
			switch kind {
			case "zero-lock":
				h = managerlock.Held{}
			case "other-lock":
				h = held(t, filepath.Join(t.TempDir(), "other"))
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "allocation":
				r.Home = filepath.Join(managerHome, "environments", "foreign-views", "checkout", "base", "codex_cli")
			case "overlap":
				r.Worktree = filepath.Join(managerHome, "worktree")
			case "linked-parent":
				if err := os.Symlink(t.TempDir(), filepath.Join(managerHome, "environments")); err != nil {
					t.Fatal(err)
				}
			}
			p := render(t, r)
			err := p.Materialize(ctx, h, managerHome)
			if err == nil {
				t.Fatal("publication boundary accepted")
			}
			if kind == "cancelled" && !errors.Is(err, context.Canceled) {
				t.Fatalf("wrong cancellation error: %v", err)
			}
			if _, err = os.Lstat(r.Home); !os.IsNotExist(err) {
				t.Fatal("refusal published native home")
			}
		})
	}
}

func driftObservation(t *testing.T, home string) map[string]any {
	t.Helper()
	out := map[string]any{}
	for _, name := range []string{"AGENTS.md", "AGENTS.override.md", "skills/tool/extra.md", "sessions.json"} {
		target := filepath.Join(home, filepath.FromSlash(name))
		st, err := os.Lstat(target)
		if os.IsNotExist(err) {
			out[name] = "absent"
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		var payload string
		if st.Mode()&os.ModeSymlink != 0 {
			payload, err = os.Readlink(target)
		} else if st.Mode().IsRegular() {
			var b []byte
			b, err = os.ReadFile(target)
			payload = string(b)
		} else {
			t.Fatal("unexpected fixture path kind")
		}
		if err != nil {
			t.Fatal(err)
		}
		out[name] = struct {
			Mode    os.FileMode
			Payload string
		}{st.Mode(), payload}
	}
	return out
}
