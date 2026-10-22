package projectcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
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
	roots := []string{}
	if command != "" {
		roots = []string{"scripts"}
	}
	b, _ := json.Marshal(map[string]any{"schema_version": 3, "runtime_roots": roots, "commands": commands, "capabilities": map[string]any{"network": "none", "filesystem": "none", "exec": "none", "secrets": "none", "env_read": "none", "prompt_scope": "none"}})
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
			assertCompositionSchemas(t, req, p)
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
	resolved, err := resolveComposition(t, a, c.Ref, c.SHA256)
	if err != nil || identity(c) != identity(resolved) {
		t.Fatalf("resolve: %v", err)
	}
	if err = authorizeComposition(t, a, c, owner); err != nil {
		t.Fatal(err)
	}
	c.Generated["AGENTS.md"] = Surface{strings.Repeat("0", 64), 1, 0600}
	required(t, authorizeComposition(t, a, c, owner), "project_context_drift")
	other := a
	other.CWD = t.TempDir()
	_, err = resolveComposition(t, other, resolved.Ref, resolved.SHA256)
	required(t, err, "project_registration_mismatch")
	file := filepath.Join(f.store.Home, "project-context/compositions", p.SHA256+".json")
	before, _ := os.ReadFile(file)
	var envelope compositionEnvelope
	_ = json.Unmarshal(before, &envelope)
	envelope.Preview.Receipt.Files[0].SHA256 = strings.Repeat("0", 64)
	b, _ := canonical(envelope)
	write(t, file, string(b))
	_, err = resolveComposition(t, a, resolved.Ref, resolved.SHA256)
	required(t, err, "project_context_drift")
	write(t, file, string(before))
	if err = f.store.Revoke(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
	_, err = resolveComposition(t, a, resolved.Ref, resolved.SHA256)
	required(t, err, "project_item_unapproved")
	if !errors.Is(err, ErrAuthorityRevoked) {
		t.Fatalf("known revoked approval lost revocation identity: %v", err)
	}
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

// Claim: generated receipts bind exact bytes and logical modes, and refuse
// portable aliases, file/ancestor collisions and oversized inventories. Failed
// receipt construction does not modify its caller's generated byte set.
func TestCompositionReceiptBounds(t *testing.T) {
	hash := strings.Repeat("1", 64)
	files := map[string][]byte{"AGENTS.md": []byte("Root\n"), ".agent-context/commands/tool": []byte("print('fixture')\n")}
	receipt, err := compositionReceipt(files, map[string]bool{".agent-context/commands/tool": true}, hash, hash, "worktree:fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range receipt.Files {
		want := "0600"
		if strings.Contains(f.Path, "commands") {
			want = "0700"
		}
		if f.Mode != want || f.SHA256 != digest(files[f.Path]) {
			t.Fatal("receipt bytes/modes differ")
		}
	}
	for _, bad := range []map[string][]byte{{"AGENTS.md": nil, "agents.md": nil}, {"skills": nil, "skills/review/SKILL.md": nil}, {"../escape": nil}} {
		before := identity(bad)
		_, err := compositionReceipt(bad, nil, hash, hash, "worktree:fixture")
		required(t, err, "project_context_invalid")
		if identity(bad) != before {
			t.Fatal("refused receipt changed source files")
		}
	}
	many := map[string][]byte{}
	for i := 0; i <= MaxFiles; i++ {
		many[fmt.Sprintf("file-%05d", i)] = nil
	}
	_, err = compositionReceipt(many, nil, hash, hash, "worktree:fixture")
	required(t, err, "project_context_invalid")
}

// Claim: MCP transport, environment selectors and pinned requirements are
// rechecked; a present but incompatible provider cannot satisfy a dependency.
func TestCompositionMCPRequirements(t *testing.T) {
	reqs := []compositionRequirement{{kind: "mcp", name: "docs", transport: "stdio", environment: "codex_cli", revision: strings.Repeat("1", 40)}}
	servers := map[string]contextmaterialize.MCPServer{"docs": {Name: "docs", Transport: "stdio", Command: "fixture"}}
	members := map[string]contextlock.Member{"docs": {Commit: strings.Repeat("1", 40)}}
	if err := checkCompositionRequirements(reqs, nil, servers, members); err != nil {
		t.Fatal(err)
	}
	server := servers["docs"]
	server.Transport = "http"
	servers["docs"] = server
	required(t, checkCompositionRequirements(reqs, nil, servers, members), "required MCP transport differs")
	server.Transport = "stdio"
	server.Environments = []string{"claude_code"}
	servers["docs"] = server
	required(t, checkCompositionRequirements(reqs, nil, servers, members), "required MCP environment differs")
	server.Environments = nil
	servers["docs"] = server
	members["docs"] = contextlock.Member{Commit: strings.Repeat("2", 40)}
	required(t, checkCompositionRequirements(reqs, nil, servers, members), "required provider pin differs")
}

func copyCompositionSkill(t *testing.T, source, target string) {
	t.Helper()
	for _, p := range []string{"agent-skill.json", "SKILL.md", "scripts/tool.py", "README.md"} {
		b, err := os.ReadFile(filepath.Join(source, p))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(target, p), string(b))
	}
}

// Claim: the real admitted-generation path includes profile exports when
// checking project commands. One profile exporter passes; an admitted second
// skill exporting that same name refuses before any composition publication.
func TestCompositionFullClosureCollision(t *testing.T) {
	source := compositionSkillSource(t, "profile", "tool")
	f := compositionFixture(t, "codex_cli", map[string]string{"profile": source})
	project := compositionSkillSource(t, "project", "tool")
	copyCompositionSkill(t, project, filepath.Join(f.root, "project"))
	d := declaration("project", "project", false)
	d.Items[0].Type = "skill"
	approveCompositionDeclaration(t, f, d)
	req := compositionRequest(f)
	req.Selections = []Selection{{"project", false}}
	_, err := f.store.PreviewComposition(context.Background(), req)
	required(t, err, "command collision")
	if _, err = os.Stat(filepath.Join(f.store.Home, "project-context/compositions")); !os.IsNotExist(err) {
		t.Fatal("collision published composition state")
	}
}

// Claim: an unchanged required provider passes through real profile closure
// reads; replacing its whole admitted skill with one lacking the required
// command refuses even though that replacement's own item approval is current.
func TestCompositionReplacementRechecksConsumer(t *testing.T) {
	provider := compositionSkillSource(t, "provider", "needed")
	consumer := compositionSkillSource(t, "consumer", "")
	b, err := os.ReadFile(filepath.Join(consumer, "agent-skill.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	_ = json.Unmarshal(b, &manifest)
	manifest["dependencies"] = map[string]any{"commands": map[string]any{"needed": map[string]any{"type": "skill", "skill": "provider", "command": "needed"}}}
	b, _ = json.Marshal(manifest)
	write(t, filepath.Join(consumer, "agent-skill.json"), string(b))
	f := compositionFixture(t, "codex_cli", map[string]string{"consumer": consumer, "provider": provider})
	copyCompositionSkill(t, provider, filepath.Join(f.root, "provider"))
	d := declaration("provider", "provider", false)
	d.Items[0].Type = "skill"
	approveCompositionDeclaration(t, f, d)
	req := compositionRequest(f)
	req.Selections = []Selection{{"provider", false}}
	if _, err = f.store.PreviewComposition(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	replacement := compositionSkillSource(t, "replacement", "")
	b, err = os.ReadFile(filepath.Join(replacement, "agent-skill.json"))
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(f.root, "provider/agent-skill.json"), string(b))
	write(t, filepath.Join(f.root, "provider/SKILL.md"), "# Replacement\n")
	approveCompositionDeclaration(t, f, d)
	_, err = f.store.PreviewComposition(context.Background(), req)
	required(t, err, "required provider command missing")
}

func compositionLock(t *testing.T, home string) *managerlock.HomeLock {
	t.Helper()
	m, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := m.AcquireExistingHomeOnly(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return lock
}
func resolveComposition(t *testing.T, a Authority, ref, hash string) (Composition, error) {
	t.Helper()
	lock := compositionLock(t, a.Store.Home)
	defer lock.Close()
	return a.Resolve(context.Background(), lock.Held(), ref, hash)
}
func authorizeComposition(t *testing.T, a Authority, c Composition, owner string) error {
	t.Helper()
	lock := compositionLock(t, a.Store.Home)
	defer lock.Close()
	return a.Authorize(context.Background(), lock.Held(), c, owner)
}

// Claim: resolver callbacks consume a valid caller-held witness without
// reacquisition, and refuse zero, stale and other-home witnesses. Borrowed
// reads neither close the owner's lock nor turn invalid authority into state.
func TestCompositionHeldAuthority(t *testing.T) {
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	req := compositionRequest(f)
	a := Authority{Store: f.store, CWD: f.root}
	owner := principal(t, f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.Resolve(context.Background(), managerlock.Held{}, c.Ref, c.SHA256)
	if !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("zero witness accepted: %v", err)
	}
	if err = a.Authorize(context.Background(), managerlock.Held{}, c, owner); !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("zero authorization witness accepted: %v", err)
	}
	lock := compositionLock(t, f.store.Home)
	held := lock.Held()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err = a.Resolve(ctx, held, c.Ref, c.SHA256); err != nil {
		t.Fatal(err)
	}
	if err = a.Authorize(ctx, held, c, owner); err != nil {
		t.Fatal(err)
	}
	if err = held.Validate(f.store.Home); err != nil {
		t.Fatal("borrowed resolver released caller lock")
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = a.Resolve(context.Background(), held, c.Ref, c.SHA256)
	if !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("stale witness accepted: %v", err)
	}
	other := compositionFixture(t, "codex_cli", nil)
	otherLock := compositionLock(t, other.store.Home)
	defer otherLock.Close()
	_, err = a.Resolve(context.Background(), otherLock.Held(), c.Ref, c.SHA256)
	if !errors.Is(err, managerlock.ErrNotHeld) {
		t.Fatalf("other-home witness accepted: %v", err)
	}
}

// Claim: the adapter contract keeps an exact release value; whitespace,
// comparison operators, empty and oversized versions refuse beside valid
// stable and prerelease controls, including non-ASCII whitespace neighbors.
func TestCompositionAdapterRelease(t *testing.T) {
	good := AdapterContract{"project-adapter/1", "codex_cli", "1.2.3", "macos", []string{"context", "memory-off", "native-source-exclusion", "permissions"}}
	for _, version := range []string{"1.2.3", "0.162.1-preview.2+build"} {
		good.ProviderVersion = version
		if err := good.validate("codex_cli"); err != nil {
			t.Fatal(err)
		}
	}
	for _, version := range []string{"", "1 2", "1\t2", "1\v2", "1\f2", "1\u00a02", "1\ufeff2", "1\u20282", "^1.2.3", ">=1", strings.Repeat("a", 257)} {
		bad := good
		bad.ProviderVersion = version
		required(t, bad.validate("codex_cli"), "invalid adapter release contract")
	}
}
