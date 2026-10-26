package projectcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/managerlock"
)

func claudeFixture(t *testing.T, empty bool, withMCP bool) (fixture, ComposeRequest, CompositionPreview, string) {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	base := t.TempDir()
	home := filepath.Join(base, "manager")
	checkout := filepath.Join(base, "checkout")
	initRepo(t, checkout)
	source := filepath.Join(base, "profile")
	contextDir(t, source, "base")
	if empty {
		write(t, filepath.Join(source, "agent-context.json"), `{"schema_version":2,"name":"base","version":"1.0.0","context":{"modules":[]}}`)
	}
	_, pin, err := contextstore.EnsureState(home, "context", "base", source)
	if err != nil {
		t.Fatal(err)
	}
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: int(hashing.WriteVersion()), Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.TrimPrefix(pin, "sha256:"), RequiredBy: []string{}}}}
	if withMCP {
		mcp := filepath.Join(base, "mcp")
		write(t, filepath.Join(mcp, "agent-mcp.json"), `{"schema_version":1,"name":"approved-server","version":"1.0.0","server":{"transport":"stdio","command":"fixture-server","args":["one","two"],"env_names":["FIXTURE_CONFIG"]}}`)
		initRepo(t, mcp)
		gitTest(t, mcp, "add", ".")
		gitTest(t, mcp, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "MCP fixture")
		commit := gitTest(t, mcp, "rev-parse", "HEAD")
		cache := filepath.Join(home, "profile-repos/github.com_example_approved-server")
		if err := os.MkdirAll(filepath.Dir(cache), 0700); err != nil {
			t.Fatal(err)
		}
		gitTest(t, mcp, "clone", mcp, cache)
		if _, err = contextstore.EnsureGit(home, "mcp", "approved-server", cache, commit); err != nil {
			t.Fatal(err)
		}
		lock.Members = append(lock.Members, contextlock.Member{Kind: "mcp", Name: "approved-server", Version: "1.0.0", Source: "github.com/example/approved-server", Commit: commit, RequiredBy: []string{"machine"}})
	}
	b, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, "profiles/base/lock.json"), string(b))
	s := Store{Home: home}
	reg, err := s.Register("fixture", checkout, "base", "claude_code")
	if err != nil {
		t.Fatal(err)
	}
	canonicalHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	s.Home = canonicalHome
	f := fixture{s, checkout, reg}
	write(t, filepath.Join(checkout, ".claude/settings.json"), `{"permissions":{"allow":["Bash"],"deny":["Read"]},"model":"foreign-model"}`)
	write(t, filepath.Join(checkout, ".claude/settings.local.json"), `{"permissions":{"allow":["Write"]}}`)
	write(t, filepath.Join(checkout, "CLAUDE.md"), "Unapproved native document.\n")
	write(t, filepath.Join(checkout, "unselected.md"), "Unapproved project file.\n")
	contextDir(t, filepath.Join(checkout, "approved"), "guide")
	write(t, filepath.Join(checkout, "approved/context/entry.md"), "Project document.\n")
	write(t, filepath.Join(checkout, "approved/not-selected.md"), "Unselected inventory text.\n")
	skill := compositionSkillSource(t, "manual", "")
	write(t, filepath.Join(checkout, "manual/references/manual.md"), "Whole skill reference.\n")
	copyCompositionSkill(t, skill, filepath.Join(checkout, "manual"))
	d := Declaration{1, []Request{{ID: "guide", Type: "context", Environment: "claude_code", Required: true, Source: "approved", Dependencies: map[string]string{}}, {ID: "manual", Type: "skill", Environment: "claude_code", Required: true, Source: "manual", Dependencies: map[string]string{}}}}
	approveCompositionDeclaration(t, f, d)
	req := compositionRequest(f)
	req.Selections = []Selection{{"guide", false}, {"manual", false}}
	req.Adapter.ProviderVersion = ClaudeProjectVersion
	req.Claude = &ClaudeControls{Version: 1}
	p, err := s.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return f, req, p, principal(t, f)
}

// This publisher fixture exercises the adapter's exact interface and bytes in
// a disposable home. Lease transactions/liveness remain I-LEASE's tests.
type claudePublisherFixture struct {
	t        *testing.T
	home     string
	expected Composition
	called   bool
	failure  error
}

func (p *claudePublisherFixture) PublishPrepared(ctx context.Context, id string, files map[string][]byte, wait bool) error {
	p.t.Helper()
	p.called = true
	if id != "reserved" || wait {
		p.t.Fatal("publication did not use reserved lease interface")
	}
	if p.failure != nil {
		return p.failure
	}
	if len(files) != len(p.expected.Generated) {
		p.t.Fatal("publication receipt count differs")
	}
	for path, b := range files {
		s, ok := p.expected.Generated[path]
		if !ok || s.SHA256 != digest(b) || s.Size != int64(len(b)) {
			p.t.Fatal("publication receipt differs")
		}
		write(p.t, filepath.Join(p.home, filepath.FromSlash(path)), string(b))
		if err := os.Chmod(filepath.Join(p.home, filepath.FromSlash(path)), os.FileMode(s.Mode)); err != nil {
			p.t.Fatal(err)
		}
	}
	return ctx.Err()
}
func publishClaudeFixture(t *testing.T, f fixture, req ComposeRequest, p CompositionPreview, owner string) (Authority, *claudePublisherFixture) {
	t.Helper()
	a := Authority{Store: f.store, CWD: f.root}
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(f.store.Home, "environments/project-views", f.registration.CheckoutID, "base/claude_code")
	pub := &claudePublisherFixture{t: t, home: home, expected: c}
	if err = a.PublishClaude(context.Background(), c.Ref, c.SHA256, owner, "reserved", pub); err != nil {
		t.Fatal(err)
	}
	return a, pub
}

// Claim: real admission/composition feeds byte-exact profile-first context,
// whole selected skills, explicit MCP and default-deny settings into a temporary
// home via the lease seam. Foreign native settings/files and unselected snapshot
// text remain absent. An empty-module profile retains its renderer's header.
func TestClaudeCompositionMaterialization(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(map[bool]string{false: "profile-first", true: "empty-profile"}[empty], func(t *testing.T) {
			f, req, p, owner := claudeFixture(t, empty, true)
			op, err := f.store.begin(false)
			if err != nil {
				t.Fatal(err)
			}
			candidate, err := op.admitted(f.registration.CheckoutID, req.Selections)
			if err != nil {
				t.Fatal(err)
			}
			c := candidate.Items[0]
			raw, err := op.lockBytes(c.Item.Payload.LockRef, c.Item.Payload.LockSHA256)
			if err != nil {
				t.Fatal(err)
			}
			op.close()
			projectLock, err := contextlock.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			projectDoc, _, err := contextmaterialize.Monolithic(projectLock, "sha256:"+c.Item.Payload.LockSHA256, contextmaterialize.Precedence{Winner: "higher-weight", Placement: "winner-last"}, "claude_code", map[string]contextmaterialize.Package{"guide": {HasContext: true, Modules: []contextmaterialize.Module{{Module: contextpkg.Module{Path: "entry.md", Class: "root"}, Bytes: []byte("Project document.\n")}}}})
			if err != nil {
				t.Fatal(err)
			}
			var expected []byte
			{
				b, err := os.ReadFile(filepath.Join(f.store.Home, "profiles/base/lock.json"))
				if err != nil {
					t.Fatal(err)
				}
				lock, err := contextlock.Parse(b)
				if err != nil {
					t.Fatal(err)
				}
				modules := []contextmaterialize.Module{}
				if !empty {
					modules = append(modules, contextmaterialize.Module{Module: contextpkg.Module{Path: "entry.md", Class: "root"}, Bytes: []byte("Package context.\n")})
				}
				base, _, err := contextmaterialize.Monolithic(lock, "sha256:"+f.registration.Profile.LockSHA256, contextmaterialize.Precedence{Winner: "higher-weight", Placement: "winner-last"}, "claude_code", map[string]contextmaterialize.Package{"base": {HasContext: true, Modules: modules}})
				if err != nil {
					t.Fatal(err)
				}
				expected = append(append(append([]byte{}, base...), '\n'), projectDoc...)
			}
			a, pub := publishClaudeFixture(t, f, req, p, owner)
			actual, err := os.ReadFile(filepath.Join(pub.home, "CLAUDE.md"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expected) {
				t.Fatal("Claude profile-first bytes differ")
			}
			settings, err := os.ReadFile(filepath.Join(pub.home, "settings.json"))
			if err != nil {
				t.Fatal(err)
			}
			if string(settings) != "{\"permissions\":{\"allow\":[],\"defaultMode\":\"dontAsk\",\"deny\":[\"mcp__*\"]}}\n" {
				t.Fatal("repository permission imported")
			}
			if string(p.Files["skills/manual/SKILL.md"]) != compositionSkillDocument("manual", "# manual\n") || string(p.Files["skills/manual/references/manual.md"]) != "Whole skill reference.\n" {
				t.Fatal("whole skill missing")
			}
			if string(p.Files[contextmaterialize.MCPClaudePath]) != "{\"mcpServers\":{\"approved-server\":{\"args\":[\"one\",\"two\"],\"command\":\"fixture-server\",\"type\":\"stdio\"}}}\n" {
				t.Fatal("MCP union encoding differs")
			}
			marker := map[string]any{}
			if err = json.Unmarshal(p.Files[claudeOwnershipPath], &marker); err != nil {
				t.Fatal(err)
			}
			if marker["owner"] != "curator" || marker["mcp_sha256"] != digest(p.Files[contextmaterialize.MCPClaudePath]) {
				t.Fatal("ownership marker differs")
			}
			for path, b := range p.Files {
				if strings.Contains(string(b), "Unapproved") || strings.Contains(string(b), "Unselected inventory") || strings.HasPrefix(path, ".claude/") {
					t.Fatal("unapproved project data reached home")
				}
			}
			if err = verifyClaudeReceipt(pub.home, p.Receipt); err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
				lock := compositionLock(t, f.store.Home)
				defer func() { _ = lock.Close() }()
				contribution, err := a.ClaudeLaunch(context.Background(), lock.Held(), p.Ref, p.SHA256, owner, ClaudeFresh)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(contribution.Fragment.MCP.EnvNames, []string{"FIXTURE_CONFIG"}) {
					t.Fatal("MCP name-only union lost")
				}
				if !reflect.DeepEqual(contribution.Args, []string{"--setting-sources", "user", "--mcp-config", filepath.Join(pub.home, contextmaterialize.MCPClaudePath), "--strict-mcp-config", "--permission-mode", "dontAsk", "--tools", ""}) {
					t.Fatal("Claude source controls missing")
				}
				if contribution.Environment()["CLAUDE_CODE_DISABLE_AUTO_MEMORY"] != "1" || contribution.Fragment.Env["CLAUDE_CONFIG_DIR"] != pub.home {
					t.Fatal("managed home or memory control missing")
				}
				required(t, contribution.Certification.RequireStrict(), "project_context_unsupported")
			}
		})
	}
}

// Claim: explicit empty MCP remains a real strict-file channel; manager tool
// grants/denies affect settings and protected policy identity, while invalid
// versions, unsorted/duplicate names and native permission expressions refuse.
func TestClaudeManagerControls(t *testing.T) {
	f, req, p, owner := claudeFixture(t, false, false)
	if string(p.Files[contextmaterialize.MCPClaudePath]) != "{\"mcpServers\":{}}\n" {
		t.Fatal("empty MCP channel missing")
	}
	req.Claude = &ClaudeControls{Version: 1, AllowTools: []string{"Read"}, DenyTools: []string{"Write"}}
	granted, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if string(granted.Files["settings.json"]) != "{\"permissions\":{\"allow\":[\"Read\"],\"defaultMode\":\"dontAsk\",\"deny\":[\"Write\",\"mcp__*\"]}}\n" {
		t.Fatal("manager policy not programmed")
	}
	if p.Record.PolicySHA256 == granted.Record.PolicySHA256 || p.Record.MaterializationSHA256 == granted.Record.MaterializationSHA256 {
		t.Fatal("manager grants not bound")
	}
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		a, pub := publishClaudeFixture(t, f, req, granted, owner)
		lock := compositionLock(t, f.store.Home)
		contribution, e := a.ClaudeLaunch(context.Background(), lock.Held(), granted.Ref, granted.SHA256, owner, ClaudeFresh)
		_ = lock.Close()
		if e != nil {
			t.Fatal(e)
		}
		if contribution.Args[len(contribution.Args)-1] != "Read" || contribution.ContextSHA256 != granted.SHA256 || contribution.HomeRef != granted.Record.HomeRef || contribution.Fragment.Env["CLAUDE_CONFIG_DIR"] != pub.home {
			t.Fatal("manager tool grant or launch binding lost")
		}
	}
	for _, controls := range []ClaudeControls{{Version: 2}, {Version: 1, AllowTools: []string{"Write", "Read"}}, {Version: 1, AllowTools: []string{"Read", "Read"}}, {Version: 1, AllowTools: []string{"Bash(*)"}}, {Version: 1, AllowTools: []string{"default"}}, {Version: 1, AllowTools: []string{"mcp__server__tool"}}} {
		files := map[string][]byte{}
		err := claudeOutputs(files, controls)
		required(t, err, "project_context_invalid")
		if len(files) != 0 {
			t.Fatal("invalid policy partially rendered")
		}
	}
	names := []string{}
	for i := 0; i < 257; i++ {
		names = append(names, fmt.Sprintf("Tool%03d", i))
	}
	if err := (ClaudeControls{Version: 1, AllowTools: names[:256]}).validate(); err != nil {
		t.Fatal(err)
	}
	required(t, (ClaudeControls{Version: 1, AllowTools: names}).validate(), "project_context_invalid")
	files := map[string][]byte{contextmaterialize.MCPClaudePath: []byte("{\"mcpServers\":{\"remote\":{\"type\":\"http\",\"url\":\"https://example.invalid/mcp\"}}}\n")}
	required(t, claudeOutputs(files, ClaudeControls{Version: 1}), "project_context_unsupported")
}

// Claim: receipt verification accepts original bytes but refuses changed bytes,
// modes, missing files and symlink substitution without rebaselining the receipt.
// Publication failures and revoked admission produce no home output.
func TestClaudeReceiptAndPublicationRefusals(t *testing.T) {
	for _, kind := range []string{"bytes", "mode", "missing", "symlink", "root-mode"} {
		t.Run(kind, func(t *testing.T) {
			f, req, p, owner := claudeFixture(t, false, false)
			a, pub := publishClaudeFixture(t, f, req, p, owner)
			path := filepath.Join(pub.home, "CLAUDE.md")
			switch kind {
			case "root-mode":
				if err := os.Chmod(pub.home, 0755); err != nil {
					t.Fatal(err)
				}
			case "bytes":
				original, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				original[0] ^= 1
				write(t, path, string(original))
			case "mode":
				if err := os.Chmod(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("settings.json", path); err != nil {
					t.Fatal(err)
				}
			}
			required(t, verifyClaudeReceipt(pub.home, p.Receipt), "project_context_drift")
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
				lock := compositionLock(t, f.store.Home)
				defer func() { _ = lock.Close() }()
				launch, err := a.ClaudeLaunch(context.Background(), lock.Held(), p.Ref, p.SHA256, owner, ClaudeFresh)
				required(t, err, "project_context_drift")
				if launch.Args != nil {
					t.Fatal("drift emitted launch contribution")
				}
			}
		})
	}
	f, req, p, owner := claudeFixture(t, false, false)
	a := Authority{Store: f.store, CWD: f.root}
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	pub := &claudePublisherFixture{t: t, failure: errors.New("project_view_busy: session retains home")}
	required(t, a.PublishClaude(context.Background(), p.Ref, p.SHA256, owner, "reserved", pub), "project_view_busy")
	if !pub.called {
		t.Fatal("lease publisher not called")
	}
	pub.called = false
	if err = f.store.Revoke(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
	required(t, a.PublishClaude(context.Background(), c.Ref, c.SHA256, owner, "reserved", pub), "project_item_unapproved")
	if pub.called {
		t.Fatal("revoked admission reached publisher")
	}
}

// Claim: resume, fork and unknown entry kinds refuse with the typed diagnostic
// before authority/filesystem lookup. Exact-version reports retain measured
// startup positives and name every known reload gap; no tuple certifies strict.
func TestClaudeCertificationAndResume(t *testing.T) {
	for _, entry := range []ClaudeEntry{ClaudeResume, ClaudeFork, "unknown"} {
		out, err := (Authority{}).ClaudeLaunch(context.Background(), managerlock.Held{}, "", "", "", entry)
		var diagnostic *ClaudeDiagnostic
		if !errors.As(err, &diagnostic) || diagnostic.Code != "project_context_unsupported" || out.Args != nil {
			t.Fatalf("resume refusal lost: %v", err)
		}
	}
	for _, tuple := range [][3]string{{ClaudeProjectVersion, "macos", "arm64"}, {"2.1.306", "macos", "arm64"}, {ClaudeProjectVersion, "linux", "arm64"}, {ClaudeProjectVersion, "macos", "amd64"}} {
		report := ClaudeCertificationReport(tuple[0], tuple[1], tuple[2])
		if tuple == [3]string{ClaudeProjectVersion, "macos", "arm64"} {
			if len(report.Certified) != 7 || len(report.Refused) != 7 {
				t.Fatal("exact tuple evidence differs")
			}
		} else if len(report.Certified) != 0 || len(report.Refused) != 14 {
			t.Fatal("unsupported tuple borrowed certification")
		}
		required(t, report.RequireStrict(), "project_context_unsupported")
	}
}

// Claim: the native opt-in refuses another release, platform or environment;
// a generic composition without manager controls cannot use Claude publication.
// Refusals leave the publisher untouched.
func TestClaudeNativeOptIn(t *testing.T) {
	f, req, p, owner := claudeFixture(t, false, false)
	for _, change := range []string{"version", "platform", "environment"} {
		changed := req
		switch change {
		case "version":
			changed.Adapter.ProviderVersion = "2.1.306"
		case "platform":
			changed.Adapter.Platform = "linux"
		case "environment":
			changed.Adapter.Environment = "codex_cli"
		}
		_, err := f.store.PreviewComposition(context.Background(), changed)
		required(t, err, "project_context_unsupported")
	}
	a := Authority{Store: f.store, CWD: f.root}
	pub := &claudePublisherFixture{t: t, failure: errors.New("unexpected publication")}
	required(t, a.PublishClaude(context.Background(), p.Ref, p.SHA256, owner, "reserved", nil), "project_lease_mismatch")
	req.Claude = nil
	generic, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.PublishComposition(context.Background(), req, generic.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	required(t, a.PublishClaude(context.Background(), c.Ref, c.SHA256, owner, "reserved", pub), "project_context_unsupported")
	if pub.called {
		t.Fatal("generic composition reached Claude publisher")
	}
}

// Claim: empty documents and empty middle documents follow I-COMP's byte rule
// through native output construction on every platform; the adapter adds no root
// bytes. Receipt verification exercises the qualified macOS/Linux binding only.
func TestClaudeEmptyDocuments(t *testing.T) {
	for _, tc := range []struct {
		profile  string
		projects []string
		want     string
	}{{"", []string{"", ""}, ""}, {"Profile\n", []string{"", "Project\n", ""}, "Profile\n\nProject\n"}, {"", []string{"", "Project\n"}, "Project\n"}} {
		documents := [][]byte{}
		for _, d := range tc.projects {
			documents = append(documents, []byte(d))
		}
		files := map[string][]byte{"CLAUDE.md": JoinRootDocuments([]byte(tc.profile), documents...)}
		if err := claudeOutputs(files, ClaudeControls{Version: 1}); err != nil {
			t.Fatal(err)
		}
		if string(files["CLAUDE.md"]) != tc.want {
			t.Fatal("empty-document bytes differ")
		}
		t.Run("receipt", func(t *testing.T) {
			if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("CIP filesystem binding supports macOS/Linux")
			}
			receipt, err := compositionReceipt(files, nil, strings.Repeat("1", 64), strings.Repeat("2", 64), "worktree:fixture")
			if err != nil {
				t.Fatal(err)
			}
			home := t.TempDir()
			if err := os.Chmod(home, 0700); err != nil {
				t.Fatal(err)
			}
			pub := &claudePublisherFixture{t: t, home: home, expected: Composition{Generated: map[string]Surface{}}}
			for _, f := range receipt.Files {
				pub.expected.Generated[f.Path] = Surface{SHA256: f.SHA256, Size: f.Size, Mode: 0600}
			}
			if err = pub.PublishPrepared(context.Background(), "reserved", files, false); err != nil {
				t.Fatal(err)
			}
			if err = verifyClaudeReceipt(home, receipt); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Claim: native delivery retains profile system bytes followed by admitted
// project system bytes in the separate append channel, never in root context.
func TestClaudeAdditiveSystemChannel(t *testing.T) {
	f := compositionFixture(t, "claude_code", nil, true)
	source := filepath.Join(f.root, "context")
	write(t, filepath.Join(source, "agent-context.json"), `{"schema_version":2,"name":"project","version":"1.0.0","context":{"modules":[{"path":"root.md"},{"path":"system.md","class":"system"}]}}`)
	write(t, filepath.Join(source, "context/root.md"), "Project root.\n")
	write(t, filepath.Join(source, "context/system.md"), "Project system.\n")
	d := declaration("guide", "context", false)
	d.Items[0].Environment = "claude_code"
	approveCompositionDeclaration(t, f, d)
	req := compositionRequest(f)
	req.Adapter.ProviderVersion = ClaudeProjectVersion
	req.Claude = &ClaudeControls{Version: 1}
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if string(p.Files[".agent-context/system-prompt.md"]) != "Base system.\n\nProject system.\n" || strings.Contains(string(p.Files["CLAUDE.md"]), "Project system.") {
		t.Fatal("additive system context lost")
	}
	// The admission store canonicalizes its home; compare that physical spelling.
	f.store.Home, err = filepath.EvalSymlinks(f.store.Home)
	if err != nil {
		t.Fatal(err)
	}
	owner := principal(t, f)
	a, pub := publishClaudeFixture(t, f, req, p, owner)
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		lock := compositionLock(t, f.store.Home)
		defer func() { _ = lock.Close() }()
		c, err := a.ClaudeLaunch(context.Background(), lock.Held(), p.Ref, p.SHA256, owner, ClaudeFresh)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"--append-system-prompt-file", filepath.Join(pub.home, ".agent-context/system-prompt.md")}
		if !reflect.DeepEqual(c.Args[len(c.Args)-2:], want) || c.Fragment.SystemPrompt == nil || c.Fragment.SystemPrompt.Channels[0].Semantics != "append" {
			t.Fatal("system append channel lost")
		}
	}
}
