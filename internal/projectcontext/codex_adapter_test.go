package projectcontext

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/projectadapter"
)

// Claim: real registration, itemized approval and I-COMP render profile-first
// strict-preview root bytes plus a whole admitted skill and profile MCP into a fresh Codex
// home with an exact new receipt and a profile-only native instruction file. Repository config/approval files and an
// unapproved native instruction never become manager policy or root context.
// This exercises manager-side delivery, not a certified vendor start.
func TestCodexAdapterCompositionEndToEnd(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("I-COMP registration requires macOS/Linux filesystem identity")
	}
	base := t.TempDir()
	home := filepath.Join(base, "manager")
	checkout := filepath.Join(base, "checkout")
	// Separate metadata keeps .git a file, as in the prepared sparse worktree.
	if err := os.MkdirAll(checkout, 0700); err != nil {
		t.Fatal(err)
	}
	gitTest(t, checkout, "init", "--separate-git-dir="+filepath.Join(base, "metadata"), "-q")
	write(t, filepath.Join(checkout, "README.md"), "Fixture\n")
	gitTest(t, checkout, "add", "README.md")
	gitTest(t, checkout, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Fixture")
	source := filepath.Join(base, "profile-context")
	contextDir(t, source, "base")
	_, pin, err := contextstore.EnsureState(home, "context", "base", source)
	if err != nil {
		t.Fatal(err)
	}
	mcpSource := filepath.Join(base, "mcp")
	write(t, filepath.Join(mcpSource, "agent-mcp.json"), `{"schema_version":1,"name":"approved","version":"1.0.0","server":{"transport":"stdio","command":"fixture-mcp","args":["serve"],"env_names":[]}}`)
	gitTest(t, mcpSource, "init", "--separate-git-dir="+filepath.Join(base, "mcp-metadata"), "-q")
	gitTest(t, mcpSource, "add", "agent-mcp.json")
	gitTest(t, mcpSource, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "MCP fixture")
	mcpCommit := gitTest(t, mcpSource, "rev-parse", "HEAD")
	cache := filepath.Join(home, "profile-repos/github.com_example_approved")
	if err = os.MkdirAll(filepath.Dir(cache), 0700); err != nil {
		t.Fatal(err)
	}
	gitTest(t, mcpSource, "clone", "--separate-git-dir="+filepath.Join(base, "mcp-cache-metadata"), mcpSource, cache)
	if _, err = contextstore.EnsureGit(home, "mcp", "approved", cache, mcpCommit); err != nil {
		t.Fatal(err)
	}

	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: int(hashing.WriteVersion()), Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.TrimPrefix(pin, "sha256:"), RequiredBy: []string{}}, {Kind: "mcp", Name: "approved", Version: "1.0.0", Source: "github.com/example/approved", Commit: mcpCommit, RequiredBy: []string{"base"}}}}
	lockBytes, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, "profiles/base/lock.json"), string(lockBytes))
	store := Store{Home: home}
	registration, err := store.Register("fixture", checkout, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{store, checkout, registration}
	skill := filepath.Join(checkout, "tool")
	write(t, filepath.Join(skill, "SKILL.md"), "# Approved tool\n")
	write(t, filepath.Join(skill, "references/data.txt"), "Whole skill bytes.\n")
	write(t, filepath.Join(skill, "agent-skill.json"), `{"schema_version":3,"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none","prompt_scope":"none"}}`)
	d := declaration("tool", "tool", false)
	d.Items[0].Type = "skill"
	approveCompositionDeclaration(t, f, d)
	write(t, filepath.Join(checkout, "AGENTS.md"), "Selected project context.\n")
	packaged, err := store.PackageInstruction(registration.CheckoutID, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range packaged.Review.Items {
		if _, err = store.Approve(registration.CheckoutID, candidate.Item.ID, candidate.ItemSHA256, packaged.PackageSHA256); err != nil {
			t.Fatal(err)
		}
	}
	// These files are deliberately present, and never passed to the renderer.
	write(t, filepath.Join(checkout, ".codex/config.toml"), "approval_policy='on-request'\nsandbox_mode='danger-full-access'\n[mcp_servers.repo]\ncommand='unapproved'\n")
	write(t, filepath.Join(checkout, ".codex/approvals.json"), `{"approved":true,"permissions":"all"}`)
	write(t, filepath.Join(checkout, "nested/AGENTS.md"), "UNAPPROVED_NATIVE\n")
	req := compositionRequest(f)
	req.Adapter.ProviderVersion = "0.162.1"
	req.Selections = []Selection{{"guide", false}, {"tool", false}}
	preview, err := store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	// Build the two root documents independently of JoinRootDocuments and assert
	// the literal separator bytes. Reversing the I-COMP join must fail here.
	profileDoc, _, err := contextmaterialize.Monolithic(lock, "sha256:"+digest(lockBytes), contextmaterialize.DefaultPrecedence, "codex_cli", map[string]contextmaterialize.Package{"base": {HasContext: true, Modules: []contextmaterialize.Module{{Module: contextpkg.Module{Path: "entry.md"}, Bytes: []byte("Package context.\n")}}}})
	if err != nil {
		t.Fatal(err)
	}
	var guide Candidate
	for _, candidate := range packaged.Review.Items {
		if candidate.Item.ID == "guide" {
			guide = candidate
		}
	}
	op, err := store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	guideLockBytes, err := op.lockBytes(guide.Item.Payload.LockRef, guide.Item.Payload.LockSHA256)
	if err != nil {
		op.close()
		t.Fatal(err)
	}
	guideLock, err := contextlock.Parse(guideLockBytes)
	if err != nil {
		op.close()
		t.Fatal(err)
	}
	pkgs := map[string]contextmaterialize.Package{}
	for _, member := range guide.Members {
		capt, err := op.get(member.SnapshotRef, member.SnapshotSHA256)
		if err != nil {
			op.close()
			t.Fatal(err)
		}
		man, err := manifest(capt)
		if err != nil {
			op.close()
			t.Fatal(err)
		}
		pkgs[member.Name] = materialPackage(man, capt.Bytes)
	}
	op.close()
	projectDoc, _, err := contextmaterialize.Monolithic(guideLock, "sha256:"+guide.Item.Payload.LockSHA256, contextmaterialize.DefaultPrecedence, "codex_cli", pkgs)
	if err != nil {
		t.Fatal(err)
	}
	wantRoot := append(append(bytes.Clone(profileDoc), '\n'), projectDoc...)
	actual := preview.Files["AGENTS.md"]
	if !bytes.Equal(actual, wantRoot) {
		t.Fatal("profile-first root bytes differ")
	}

	generated := map[string]projectadapter.Surface{}
	for _, file := range preview.Receipt.Files {
		generated[file.Path] = projectadapter.Surface{SHA256: file.SHA256, Size: file.Size, Mode: 0600}
	}
	profileGenerated := map[string]projectadapter.Surface{}
	for _, file := range preview.ProfileDocumentReceipt {
		profileGenerated[file.Path] = projectadapter.Surface{SHA256: file.SHA256, Size: file.Size, Mode: 0600}
	}
	target := filepath.Join(home, "environments/project-views", registration.CheckoutID, "base/codex_cli")
	plan, err := projectadapter.RenderCodex(projectadapter.CodexRequest{ProfileDocuments: preview.ProfileDocuments, ProfileGenerated: profileGenerated, NativeProjectEnabled: true, Files: preview.Files, Generated: generated, SeedConfig: []byte("[mcp_servers.seed]\ncommand='unapproved'\n"), SeedRevision: envregistry.CodexSeedRevisionA, ProviderVersion: "0.162.1", Platform: "macos", Mode: projectadapter.NativeProject, Model: "fixture-model", Effort: "high", Approval: "never", Sandbox: "read-only", Home: target, Worktree: checkout})
	if err != nil {
		t.Fatal(err)
	}
	held := compositionLock(t, home)
	if err = plan.Materialize(context.Background(), held.Held(), home); err != nil {
		t.Fatal(err)
	}
	for name, want := range plan.Files() {
		got, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("materialized bytes differ: %s", name)
		}
	}
	got, _ := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if !bytes.Equal(got, profileDoc) || bytes.Contains(got, projectDoc) {
		t.Fatal("native-project managed root is not profile-only")
	}
	got, _ = os.ReadFile(filepath.Join(target, "skills/tool/references/data.txt"))
	if string(got) != "Whole skill bytes.\n" {
		t.Fatal("whole skill missing")
	}
	configBytes, _ := os.ReadFile(filepath.Join(target, "config.toml"))
	config := map[string]any{}
	if _, err = toml.Decode(string(configBytes), &config); err != nil {
		t.Fatal(err)
	}
	if config["approval_policy"] != "never" || config["sandbox_mode"] != "read-only" {
		t.Fatal("repository approval imported")
	}
	servers := config["mcp_servers"].(map[string]any)
	if len(servers) != 1 || servers["approved"] == nil || servers["seed"] != nil || servers["repo"] != nil {
		t.Fatal("MCP union differs")
	}
	if err = plan.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(home, "profiles/base/lock.json"))
	if !bytes.Equal(after, lockBytes) {
		t.Fatal("adapter changed profile lock")
	}
	if err = plan.RequireStrict(); err == nil {
		t.Fatal("uncertified strict launch accepted")
	}
}
