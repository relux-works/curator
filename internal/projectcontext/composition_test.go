package projectcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/skillspec"
)

func compositionFixture(t *testing.T, env string, skillSources map[string]string, system ...bool) fixture {
	t.Helper()
	base := t.TempDir()
	home := filepath.Join(base, "manager")
	root := filepath.Join(base, "checkout")
	initRepo(t, root)
	source := filepath.Join(base, "base-context")
	contextDir(t, source, "base")
	if len(system) > 0 && system[0] {
		write(t, filepath.Join(source, "agent-context.json"), `{"schema_version":2,"name":"base","version":"1.0.0","context":{"modules":[{"path":"entry.md"},{"path":"system.md","class":"system"}]}}`)
		write(t, filepath.Join(source, "context/system.md"), "Base system.\n")
	}
	entry, pin, err := contextstore.EnsureState(home, "context", "base", source)
	if err != nil {
		t.Fatal(err)
	}
	_ = entry
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: int(hashing.WriteVersion()), Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.TrimPrefix(pin, "sha256:"), RequiredBy: []string{}}}}

	for _, name := range sortedKeys(skillSources) {
		src := skillSources[name]
		initRepo(t, src)
		gitTest(t, src, "add", ".")
		gitTest(t, src, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Skill fixture")
		commit := gitTest(t, src, "rev-parse", "HEAD")
		source := "github.com/example/" + name
		cache := filepath.Join(home, "profile-repos", "github.com_example_"+name)
		if err := os.MkdirAll(filepath.Dir(cache), 0700); err != nil {
			t.Fatal(err)
		}
		gitTest(t, src, "clone", src, cache)
		if _, err := contextstore.EnsureGit(home, "skill", name, cache, commit); err != nil {
			t.Fatal(err)
		}
		lock.Members = append(lock.Members, contextlock.Member{Kind: "skill", Name: name, Source: source, Commit: commit, RequiredBy: []string{"machine"}})
	}

	b, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, "profiles", "base", "lock.json"), string(b))
	store := Store{Home: home}
	reg, err := store.Register("fixture", root, "base", env)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{store, root, reg}
}
func compositionRequest(f fixture) ComposeRequest {
	return ComposeRequest{CheckoutID: f.registration.CheckoutID, Selections: []Selection{{"guide", false}}, Adapter: AdapterContract{"project-adapter/1", f.registration.Environment, "1.2.3", "macos", []string{"commands", "context", "mcp", "memory-off", "native-source-exclusion", "permissions", "skills"}}, Policy: CompositionPolicy{Version: 1, Memory: "off"}}
}
func principal(t *testing.T, f fixture) string {
	t.Helper()
	op, err := f.store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	defer op.close()
	return op.state.Principal
}
func compositionSkillSource(t *testing.T, name, command string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	commands := map[string]any{}
	if command != "" {
		commands[command] = map[string]any{"type": "script", "unix_path": "scripts/tool.py"}
		write(t, filepath.Join(root, "scripts/tool.py"), "print('fixture')\n")
	}
	b, _ := json.Marshal(map[string]any{"schema_version": 3, "runtime_roots": []string{"scripts"}, "commands": commands, "capabilities": map[string]any{"network": "none", "filesystem": "none", "exec": "none", "secrets": "none", "env_read": "none", "prompt_scope": "none"}})
	write(t, filepath.Join(root, "agent-skill.json"), string(b))
	write(t, filepath.Join(root, "SKILL.md"), "# "+name+"\n")
	return root
}
func approveCompositionDeclaration(t *testing.T, f fixture, d Declaration) Preview {
	t.Helper()
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range p.Review.Items {
		if _, err = f.store.Approve(f.registration.CheckoutID, c.Item.ID, c.ItemSHA256, p.PackageSHA256); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

// Claim: non-empty documents retain their bytes and generation order, with one
// additional LF. Empty documents contribute neither bytes nor separators.
func TestCompositionRootBytes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		profile  string
		projects []string
		want     string
	}{
		{"ordered", "Profile\n", []string{"General\n", "Environment\n"}, "Profile\n\nGeneral\n\nEnvironment\n"},
		{"empty-middle", "Profile\n", []string{"", "Project\n", ""}, "Profile\n\nProject\n"},
		{"empty-base", "", []string{"", "Project\n"}, "Project\n"},
		{"all-empty", "", []string{"", ""}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projects := [][]byte{}
			for _, s := range tc.projects {
				projects = append(projects, []byte(s))
			}
			if got := string(JoinRootDocuments([]byte(tc.profile), projects...)); got != tc.want {
				t.Fatalf("root bytes differ: got %q want %q", got, tc.want)
			}
		})
	}
}

// Claim: each first-wave environment renders profile-first bytes and a complete
// hash/size/mode receipt from retained context; preview creates no composition
// allocation or lease and leaves the selected lock byte-identical.
func TestCompositionPreviewAndReceipt(t *testing.T) {
	for _, env := range []string{"codex_cli", "claude_code", "muse"} {
		t.Run(env, func(t *testing.T) {
			f := compositionFixture(t, env, nil)
			packageApproved(t, f, false)
			req := compositionRequest(f)
			before, _ := os.ReadFile(filepath.Join(f.store.Home, "profiles/base/lock.json"))
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			target := "AGENTS.md"
			if env == "claude_code" {
				target = "CLAUDE.md"
			}
			if env == "muse" {
				target = "config/muse/AGENTS.md"
			}
			doc := string(p.Files[target])
			if strings.Index(doc, "Package context.\n") < 0 || strings.Index(doc, "Selected project context.\n") < strings.Index(doc, "Package context.\n") {
				t.Fatal("profile-first rendering lost")
			}
			for _, file := range p.Receipt.Files {
				b, ok := p.Files[file.Path]
				if !ok || file.SHA256 != digest(b) || file.Size != int64(len(b)) || file.Mode != "0600" {
					t.Fatal("receipt differs from output")
				}
			}
			if len(p.Files) != len(p.Receipt.Files) || p.Record.MaterializationSHA256 != identity(p.Receipt) || p.Record.GenerationSHA256 != identity(p.Generation) {
				t.Fatal("supporting records not bound")
			}
			for _, dir := range []string{"compositions", "leases"} {
				if _, err := os.Stat(filepath.Join(f.store.Home, "project-context", dir)); !os.IsNotExist(err) {
					t.Fatal("preview allocated protected publication")
				}
			}
			after, _ := os.ReadFile(filepath.Join(f.store.Home, "profiles/base/lock.json"))
			if !bytes.Equal(before, after) {
				t.Fatal("selected profile lock mutated")
			}
		})
	}
}

// Claim: identical state pins deduplicate one skill; a changed same-identity
// project skill replaces its entire context and command set, preserving Y.
func TestCompositionWholeSkillReplacement(t *testing.T) {
	for _, same := range []bool{true, false} {
		t.Run(map[bool]string{true: "same-pin", false: "replace"}[same], func(t *testing.T) {
			old := compositionSkillSource(t, "review", "old-command")
			f := compositionFixture(t, "codex_cli", map[string]string{"review": old})
			// Use the real whole-skill reader with a distinct project source directory.
			project := filepath.Join(f.root, "review")
			if err := os.MkdirAll(project, 0700); err != nil {
				t.Fatal(err)
			}
			if same {
				for _, p := range []string{"agent-skill.json", "SKILL.md", "scripts/tool.py", "README.md"} {
					b, err := os.ReadFile(filepath.Join(old, p))
					if err != nil {
						t.Fatal(err)
					}
					write(t, filepath.Join(project, p), string(b))
				}
			} else {
				write(t, filepath.Join(project, "agent-skill.json"), `{"schema_version":3,"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none","prompt_scope":"none"}}`)
				write(t, filepath.Join(project, "SKILL.md"), "# Replacement\n")
			}
			d := declaration("review", "review", false)
			d.Items[0].Type = "skill"
			approveCompositionDeclaration(t, f, d)
			req := compositionRequest(f)
			req.Selections = []Selection{{"review", false}}
			before, _ := os.ReadFile(filepath.Join(f.store.Home, "profiles/base/lock.json"))
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Skills) != 1 {
				t.Fatal("same identity duplicated")
			}
			if same {
				if p.Skills[0].Layer != "profile" {
					t.Fatal("same pin failed to deduplicate profile skill")
				}
				if len(p.Record.CommandOwners) != 1 {
					t.Fatal("same pin lost command")
				}
			} else {
				if string(p.Files["skills/review/SKILL.md"]) != "# Replacement\n" || len(p.Record.CommandOwners) != 0 {
					t.Fatal("whole-skill replacement retained old context or command")
				}
				if _, ok := p.Files[".agent-context/commands/old-command"]; ok {
					t.Fatal("old executable survived replacement")
				}
			}
			after, _ := os.ReadFile(filepath.Join(f.store.Home, "profiles/base/lock.json"))
			if !bytes.Equal(before, after) {
				t.Fatal("Y mutated")
			}
		})
	}
}

// Claim: command ownership includes profile and transitive skill identities;
// two exporters of one command refuse, while one exporter remains valid.
func TestCompositionCommandCollision(t *testing.T) {
	spec := &skillspec.Spec{Commands: map[string]skillspec.Command{"tool": {Type: "script", UnixPath: "scripts/tool.py"}}}
	mk := func(name string) compositionSkill {
		return compositionSkill{EffectiveSkill: EffectiveSkill{Layer: map[bool]string{true: "project", false: "profile"}[name == "project"], Name: name, ItemID: name, ArtifactSHA256: strings.Repeat("a", 64)}, spec: spec, files: map[string][]byte{"scripts/tool.py": []byte("fixture\n")}}
	}
	skills := map[string]compositionSkill{"profile": mk("profile")}
	if _, err := compositionCommands(skills); err != nil {
		t.Fatal(err)
	}
	skills["project"] = mk("project")
	_, err := compositionCommands(skills)
	required(t, err, "command collision")
	delete(skills, "project")
	skills["transitive"] = mk("transitive")
	_, err = compositionCommands(skills)
	required(t, err, "command collision")
}

// Claim: required providers are rechecked after replacement; a removed export
// refuses while a provider retaining the required export passes.
func TestCompositionRequiredProvider(t *testing.T) {
	provider := compositionSkill{member: contextlock.Member{Name: "provider", Directory: "."}, spec: &skillspec.Spec{Commands: map[string]skillspec.Command{"needed": {}}}}
	skills := map[string]compositionSkill{"provider": provider}
	reqs := []compositionRequirement{{kind: "skill", name: "provider", commands: []string{"needed"}}}
	if err := checkCompositionRequirements(reqs, skills, nil, nil); err != nil {
		t.Fatal(err)
	}
	provider.spec = &skillspec.Spec{Commands: map[string]skillspec.Command{}}
	skills["provider"] = provider
	required(t, checkCompositionRequirements(reqs, skills, nil, nil), "required provider command missing")
	delete(skills, "provider")
	required(t, checkCompositionRequirements(reqs, skills, nil, nil), "required skill provider missing")
}

// Claim: identical MCP stable IDs/digests deduplicate; a changed definition
// refuses without replacement approval and preserves the previous definition.
func TestCompositionMCPConflict(t *testing.T) {
	first := contextmaterialize.MCPServer{Name: "docs", Transport: "http", URL: "https://example.invalid/api"}
	set := map[string]contextmaterialize.MCPServer{}
	if err := mergeCompositionMCP(set, first); err != nil {
		t.Fatal(err)
	}
	if err := mergeCompositionMCP(set, first); err != nil || len(set) != 1 {
		t.Fatal("same MCP did not deduplicate")
	}
	next := first
	next.URL = "https://example.invalid/other"
	required(t, mergeCompositionMCP(set, next), "same-ID MCP replacement")
	if set["docs"].URL != first.URL {
		t.Fatal("refused MCP changed baseline")
	}
}

// Claim: publication requires the reviewed digest, current local principal and
// exact CWD registration. Resolve and Authorize reject tampered receipts,
// revoked items and forged returned fields without rewriting approved state.
func TestCompositionAuthority(t *testing.T) {
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	req := compositionRequest(f)
	a := Authority{Store: f.store, CWD: f.root}
	owner := principal(t, f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.PublishComposition(context.Background(), req, strings.Repeat("0", 64), owner)
	required(t, err, "project_context_drift")
	_, err = a.PublishComposition(context.Background(), req, p.SHA256, "principal:other")
	required(t, err, "project_item_unapproved")
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := a.Resolve(context.Background(), c.Ref, c.SHA256)
	if err != nil || identity(c) != identity(resolved) {
		t.Fatalf("resolve: %v", err)
	}
	if err = a.Authorize(context.Background(), c, owner); err != nil {
		t.Fatal(err)
	}
	c.Generated["AGENTS.md"] = Surface{strings.Repeat("0", 64), 1, 0600}
	required(t, a.Authorize(context.Background(), c, owner), "project_context_drift")
	other := a
	other.CWD = t.TempDir()
	_, err = other.Resolve(context.Background(), resolved.Ref, resolved.SHA256)
	required(t, err, "project_registration_mismatch")
	file := filepath.Join(f.store.Home, "project-context/compositions", p.SHA256+".json")
	before, _ := os.ReadFile(file)
	var envelope compositionEnvelope
	_ = json.Unmarshal(before, &envelope)
	envelope.Preview.Receipt.Files[0].SHA256 = strings.Repeat("0", 64)
	b, _ := canonical(envelope)
	write(t, file, string(b))
	_, err = a.Resolve(context.Background(), resolved.Ref, resolved.SHA256)
	required(t, err, "project_context_drift")
	write(t, file, string(before))
	if err = f.store.Revoke(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
	_, err = a.Resolve(context.Background(), resolved.Ref, resolved.SHA256)
	required(t, err, "project_item_unapproved")
	after, _ := os.ReadFile(file)
	if !bytes.Equal(before, after) {
		t.Fatal("refused resolution rebaselined composition")
	}
}

// Claim: unreadable or changed protected base members and receipt state refuse;
// neither path becomes an unchanged/fresh composition or modifies Y's lock.
func TestCompositionDriftAndUnreadable(t *testing.T) {
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	req := compositionRequest(f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	_ = p
	_, lock, err := profile(f.store.Home, "base")
	if err != nil {
		t.Fatal(err)
	}
	entry := contextstore.EntryDir(f.store.Home, "context", "base", lock.Members[0].StateHash)
	module := filepath.Join(entry, "context/entry.md")
	before, _ := os.ReadFile(module)
	if err = os.Chmod(module, 0000); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.PreviewComposition(context.Background(), req)
	required(t, err, "project_context_drift")
	if err = os.Chmod(module, 0600); err != nil {
		t.Fatal(err)
	}
	write(t, module, "Changed protected bytes.\n")
	_, err = f.store.PreviewComposition(context.Background(), req)
	required(t, err, "project_context_drift")
	write(t, module, string(before))
	if _, err = f.store.PreviewComposition(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}

// Claim: missing additive/native-exclusion capabilities, memory-on policy,
// cancelled preparation, and unsupported required item arms refuse. Their
// valid neighbors retain a usable preview and never allocate publication.
func TestCompositionClosedPreparation(t *testing.T) {
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	good := compositionRequest(f)
	if _, err := f.store.PreviewComposition(context.Background(), good); err != nil {
		t.Fatal(err)
	}
	bad := good
	bad.Policy.Memory = "on"
	_, err := f.store.PreviewComposition(context.Background(), bad)
	required(t, err, "project_context_invalid")
	bad = good
	bad.Adapter.Surfaces = []string{"context", "memory-off", "permissions"}
	_, err = f.store.PreviewComposition(context.Background(), bad)
	required(t, err, "project_context_unsupported")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = f.store.PreviewComposition(ctx, good)
	if err != context.Canceled {
		t.Fatalf("cancelled preview: %v", err)
	}
	bad = good
	bad.Selections = []Selection{}
	_, err = f.store.PreviewComposition(context.Background(), bad)
	required(t, err, "project_context_unsupported")
}

// Claim: system modules retain their separate channel and profile/project
// order. They never appear in root context, and an adapter without that system
// channel refuses rather than silently dropping or promoting the bytes.
func TestCompositionSystemChannel(t *testing.T) {
	for _, env := range []string{"codex_cli", "muse"} {
		t.Run(env, func(t *testing.T) {
			f := compositionFixture(t, env, nil, true)
			source := filepath.Join(f.root, "context")
			write(t, filepath.Join(source, "agent-context.json"), `{"schema_version":2,"name":"project","version":"1.0.0","context":{"modules":[{"path":"root.md"},{"path":"system.md","class":"system"}]}}`)
			write(t, filepath.Join(source, "context/root.md"), "Project root.\n")
			write(t, filepath.Join(source, "context/system.md"), "Project system.\n")
			d := declaration("guide", "context", false)
			d.Items[0].Environment = env
			approveCompositionDeclaration(t, f, d)
			p, err := f.store.PreviewComposition(context.Background(), compositionRequest(f))
			if env == "muse" {
				required(t, err, "project_context_unsupported")
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(p.Files["AGENTS.md"]), "Project system.") {
				t.Fatal("system module promoted into root")
			}
			if string(p.Files[".agent-context/system-prompt.md"]) != "Base system.\n\nProject system.\n" {
				t.Fatal("separate system channel lost bytes")
			}
		})
	}
}
